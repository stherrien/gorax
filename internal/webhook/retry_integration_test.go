package webhook

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// HTTPWebhookDeliverer implements WebhookDeliverer using HTTP
type HTTPWebhookDeliverer struct {
	client *http.Client
}

// NewHTTPWebhookDeliverer creates a new HTTP webhook deliverer
func NewHTTPWebhookDeliverer() *HTTPWebhookDeliverer {
	return &HTTPWebhookDeliverer{
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// DeliverWebhook delivers a webhook via HTTP
func (d *HTTPWebhookDeliverer) DeliverWebhook(ctx context.Context, webhook *Webhook, event *WebhookEvent) (*RetryResult, error) {
	// Create request
	req, err := http.NewRequestWithContext(ctx, event.RequestMethod, webhook.Path, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	// Add headers
	for key, value := range event.RequestHeaders {
		req.Header.Set(key, value)
	}

	// Send request
	resp, err := d.client.Do(req)
	if err != nil {
		return &RetryResult{
			Success:    false,
			StatusCode: 0,
			Error:      ErrWebhookConnectionFailed,
		}, ErrWebhookConnectionFailed
	}
	defer resp.Body.Close()

	// Check response
	success := resp.StatusCode >= 200 && resp.StatusCode < 300

	result := &RetryResult{
		Success:    success,
		StatusCode: resp.StatusCode,
	}

	if !success {
		if resp.StatusCode >= 500 {
			result.Error = ErrWebhookServerError
			return result, ErrWebhookServerError
		}
		if resp.StatusCode == 429 {
			result.Error = ErrWebhookRateLimited
			return result, ErrWebhookRateLimited
		}
		if resp.StatusCode == 401 || resp.StatusCode == 403 {
			result.Error = ErrWebhookAuthFailed
			return result, ErrWebhookAuthFailed
		}
		result.Error = ErrWebhookValidationFailed
		return result, ErrWebhookValidationFailed
	}

	return result, nil
}

func TestRetryIntegration_SuccessAfterRetries(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	db := setupTestDB(t)
	defer db.Close()

	const testTenantID = "test-tenant-123"
	const testWorkflowID = "test-workflow-123"

	repo := NewRepository(db)
	ctx := context.Background()

	// Create a test server that fails first 2 times, then succeeds
	var attempts atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		count := attempts.Add(1)
		if count <= 2 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Create webhook with server URL
	webhook, err := repo.Create(ctx, testTenantID, testWorkflowID, "test-node", "secret", AuthTypeSignature)
	require.NoError(t, err)

	// Update webhook path to point to test server
	_, err = db.ExecContext(ctx, "UPDATE webhooks SET path = $1 WHERE id = $2", server.URL, webhook.ID)
	require.NoError(t, err)
	webhook.Path = server.URL

	// Create failed event
	event := &WebhookEvent{
		TenantID:       testTenantID,
		WebhookID:      webhook.ID,
		RequestMethod:  "POST",
		RequestHeaders: map[string]string{"Content-Type": "application/json"},
		RequestBody:    json.RawMessage(`{"test": "data"}`),
		Status:         EventStatusFailed,
	}
	err = repo.CreateEvent(ctx, event)
	require.NoError(t, err)

	// Mark for retry with past time
	pastTime := time.Now().Add(-1 * time.Minute)
	_, err = db.ExecContext(ctx, "UPDATE webhook_events SET next_retry_at = $1, retry_count = 0 WHERE id = $2", pastTime, event.ID)
	require.NoError(t, err)

	// Create deliverer and worker
	deliverer := NewHTTPWebhookDeliverer()
	config := RetryConfig{
		MaxAttempts: 5,
		BaseDelay:   100 * time.Millisecond,
		MaxDelay:    1 * time.Second,
		Multiplier:  2.0,
	}
	worker := NewRetryWorker(repo, deliverer, config)

	// Process retries multiple times
	for i := 0; i < 3; i++ {
		processed, err := worker.ProcessRetries(ctx, 10)
		require.NoError(t, err)
		if processed > 0 {
			// Get updated event
			event, err = repo.GetEventByID(ctx, testTenantID, event.ID)
			require.NoError(t, err)

			if event.Status == EventStatusProcessed {
				break
			}

			// Update next_retry_at to past for next iteration
			_, err = db.ExecContext(ctx, "UPDATE webhook_events SET next_retry_at = $1 WHERE id = $2", pastTime, event.ID)
			require.NoError(t, err)
		}
		time.Sleep(100 * time.Millisecond)
	}

	// Verify final state
	finalEvent, err := repo.GetEventByID(ctx, testTenantID, event.ID)
	require.NoError(t, err)
	assert.Equal(t, EventStatusProcessed, finalEvent.Status)
	assert.Equal(t, int32(3), attempts.Load())
}

func TestRetryIntegration_PermanentFailureAfterMaxRetries(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	db := setupTestDB(t)
	defer db.Close()

	repo := NewRepository(db)
	ctx := context.Background()

	// Create a test server that always fails
	var attempts atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts.Add(1)
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	// Create webhook
	webhook, err := repo.Create(ctx, testTenantID, testWorkflowID, "test-node", "secret", AuthTypeSignature)
	require.NoError(t, err)
	_, err = db.ExecContext(ctx, "UPDATE webhooks SET path = $1 WHERE id = $2", server.URL, webhook.ID)
	require.NoError(t, err)
	webhook.Path = server.URL

	// Create failed event
	event := &WebhookEvent{
		TenantID:       testTenantID,
		WebhookID:      webhook.ID,
		RequestMethod:  "POST",
		RequestHeaders: map[string]string{},
		RequestBody:    json.RawMessage(`{}`),
		Status:         EventStatusFailed,
	}
	err = repo.CreateEvent(ctx, event)
	require.NoError(t, err)

	// Set max retries to 3
	pastTime := time.Now().Add(-1 * time.Minute)
	_, err = db.ExecContext(ctx, "UPDATE webhook_events SET next_retry_at = $1, retry_count = 0, max_retries = 3 WHERE id = $2", pastTime, event.ID)
	require.NoError(t, err)

	// Create worker
	deliverer := NewHTTPWebhookDeliverer()
	config := RetryConfig{
		MaxAttempts: 3,
		BaseDelay:   100 * time.Millisecond,
		MaxDelay:    1 * time.Second,
		Multiplier:  2.0,
	}
	worker := NewRetryWorker(repo, deliverer, config)

	// Process retries until permanently failed
	for i := 0; i < 5; i++ {
		processed, err := worker.ProcessRetries(ctx, 10)
		require.NoError(t, err)
		if processed > 0 {
			// Get updated event
			event, err = repo.GetEventByID(ctx, testTenantID, event.ID)
			require.NoError(t, err)

			if event.PermanentlyFailed {
				break
			}

			// Update next_retry_at to past for next iteration
			_, err = db.ExecContext(ctx, "UPDATE webhook_events SET next_retry_at = $1 WHERE id = $2", pastTime, event.ID)
			require.NoError(t, err)
		}
		time.Sleep(100 * time.Millisecond)
	}

	// Verify final state
	finalEvent, err := repo.GetEventByID(ctx, testTenantID, event.ID)
	require.NoError(t, err)
	assert.True(t, finalEvent.PermanentlyFailed)
	assert.Equal(t, 3, finalEvent.RetryCount)
	assert.Nil(t, finalEvent.NextRetryAt)
}

func TestRetryIntegration_NonRetryableError(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	db := setupTestDB(t)
	defer db.Close()

	repo := NewRepository(db)
	ctx := context.Background()

	// Create a test server that returns 401 (non-retryable)
	var attempts atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts.Add(1)
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer server.Close()

	// Create webhook
	webhook, err := repo.Create(ctx, testTenantID, testWorkflowID, "test-node", "secret", AuthTypeSignature)
	require.NoError(t, err)
	_, err = db.ExecContext(ctx, "UPDATE webhooks SET path = $1 WHERE id = $2", server.URL, webhook.ID)
	require.NoError(t, err)
	webhook.Path = server.URL

	// Create failed event
	event := &WebhookEvent{
		TenantID:       testTenantID,
		WebhookID:      webhook.ID,
		RequestMethod:  "POST",
		RequestHeaders: map[string]string{},
		RequestBody:    json.RawMessage(`{}`),
		Status:         EventStatusFailed,
	}
	err = repo.CreateEvent(ctx, event)
	require.NoError(t, err)

	pastTime := time.Now().Add(-1 * time.Minute)
	_, err = db.ExecContext(ctx, "UPDATE webhook_events SET next_retry_at = $1, retry_count = 0 WHERE id = $2", pastTime, event.ID)
	require.NoError(t, err)

	// Create worker
	deliverer := NewHTTPWebhookDeliverer()
	worker := NewRetryWorker(repo, deliverer, DefaultRetryConfig())

	// Process retry (should fail immediately with non-retryable error)
	processed, err := worker.ProcessRetries(ctx, 10)
	require.NoError(t, err)
	assert.Equal(t, 1, processed)

	// Verify permanently failed after first attempt
	finalEvent, err := repo.GetEventByID(ctx, testTenantID, event.ID)
	require.NoError(t, err)
	assert.True(t, finalEvent.PermanentlyFailed)
	assert.Equal(t, 0, finalEvent.RetryCount) // No additional retries
	assert.Nil(t, finalEvent.NextRetryAt)
	assert.Equal(t, int32(1), attempts.Load())
}

func TestRetryIntegration_Statistics(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	db := setupTestDB(t)
	defer db.Close()

	repo := NewRepository(db)
	ctx := context.Background()

	// Create webhook
	webhook, err := repo.Create(ctx, testTenantID, testWorkflowID, "test-node", "secret", AuthTypeSignature)
	require.NoError(t, err)

	// Create multiple events with different retry states
	// Event 1: 2 retries, still pending
	event1 := &WebhookEvent{
		TenantID:       testTenantID,
		WebhookID:      webhook.ID,
		RequestMethod:  "POST",
		RequestHeaders: map[string]string{},
		RequestBody:    json.RawMessage(`{}`),
		Status:         EventStatusFailed,
	}
	err = repo.CreateEvent(ctx, event1)
	require.NoError(t, err)
	futureTime := time.Now().Add(5 * time.Minute)
	_, err = db.ExecContext(ctx, "UPDATE webhook_events SET retry_count = 2, next_retry_at = $1 WHERE id = $2", futureTime, event1.ID)
	require.NoError(t, err)

	// Event 2: Permanently failed
	event2 := &WebhookEvent{
		TenantID:       testTenantID,
		WebhookID:      webhook.ID,
		RequestMethod:  "POST",
		RequestHeaders: map[string]string{},
		RequestBody:    json.RawMessage(`{}`),
		Status:         EventStatusFailed,
	}
	err = repo.CreateEvent(ctx, event2)
	require.NoError(t, err)
	_, err = db.ExecContext(ctx, "UPDATE webhook_events SET permanently_failed = true, retry_count = 3 WHERE id = $1", event2.ID)
	require.NoError(t, err)

	// Event 3: 1 retry, permanently failed
	event3 := &WebhookEvent{
		TenantID:       testTenantID,
		WebhookID:      webhook.ID,
		RequestMethod:  "POST",
		RequestHeaders: map[string]string{},
		RequestBody:    json.RawMessage(`{}`),
		Status:         EventStatusFailed,
	}
	err = repo.CreateEvent(ctx, event3)
	require.NoError(t, err)
	_, err = db.ExecContext(ctx, "UPDATE webhook_events SET permanently_failed = true, retry_count = 1 WHERE id = $1", event3.ID)
	require.NoError(t, err)

	// Get statistics
	stats, err := repo.GetRetryStatistics(ctx, webhook.ID)
	require.NoError(t, err)

	assert.Equal(t, 3, stats.TotalRetriedEvents)
	assert.Equal(t, 2, stats.PermanentlyFailedEvents)
	assert.Equal(t, 1, stats.PendingRetries)
	assert.Equal(t, 3, stats.MaxRetryCount)
	assert.Equal(t, 2.0, stats.AvgRetryCount)
}
