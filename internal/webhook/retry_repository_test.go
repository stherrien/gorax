package webhook

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testTenantID   = "test-tenant-123"
	testWorkflowID = "test-workflow-123"
)

func TestRepository_MarkEventForRetry(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	db := setupTestDB(t)
	defer db.Close()

	repo := NewRepository(db)
	ctx := context.Background()

	// Create test webhook
	webhook, err := repo.Create(ctx, testTenantID, testWorkflowID, "test-node", "secret", AuthTypeSignature)
	require.NoError(t, err)

	// Create webhook event
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

	// Mark for retry
	err = repo.MarkEventForRetry(ctx, event.ID, "connection timeout")
	require.NoError(t, err)

	// Verify retry state
	retrieved, err := repo.GetEventByID(ctx, testTenantID, event.ID)
	require.NoError(t, err)
	assert.Equal(t, 1, retrieved.RetryCount)
	assert.NotNil(t, retrieved.NextRetryAt)
	assert.NotNil(t, retrieved.LastRetryAt)
	assert.Equal(t, "connection timeout", *retrieved.RetryError)
	assert.False(t, retrieved.PermanentlyFailed)
	assert.True(t, retrieved.NextRetryAt.After(time.Now()))
}

func TestRepository_MarkEventForRetry_MaxRetriesExceeded(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	db := setupTestDB(t)
	defer db.Close()

	repo := NewRepository(db)
	ctx := context.Background()

	// Create test webhook
	webhook, err := repo.Create(ctx, testTenantID, testWorkflowID, "test-node", "secret", AuthTypeSignature)
	require.NoError(t, err)

	// Create webhook event with 2 retries already done
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

	// Manually set retry count to max - 1
	_, err = db.ExecContext(ctx, "UPDATE webhook_events SET retry_count = 2, max_retries = 3 WHERE id = $1", event.ID)
	require.NoError(t, err)

	// Mark for retry (this will be the 3rd and final retry)
	err = repo.MarkEventForRetry(ctx, event.ID, "final error")
	require.NoError(t, err)

	// Verify permanently failed
	retrieved, err := repo.GetEventByID(ctx, testTenantID, event.ID)
	require.NoError(t, err)
	assert.Equal(t, 3, retrieved.RetryCount)
	assert.True(t, retrieved.PermanentlyFailed)
	assert.Nil(t, retrieved.NextRetryAt)
	assert.Equal(t, "final error", *retrieved.RetryError)
}

func TestRepository_MarkEventAsNonRetryable(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	db := setupTestDB(t)
	defer db.Close()

	repo := NewRepository(db)
	ctx := context.Background()

	// Create test webhook
	webhook, err := repo.Create(ctx, testTenantID, testWorkflowID, "test-node", "secret", AuthTypeSignature)
	require.NoError(t, err)

	// Create webhook event
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

	// Mark as non-retryable
	err = repo.MarkEventAsNonRetryable(ctx, event.ID, "authentication failed")
	require.NoError(t, err)

	// Verify permanently failed without retry
	retrieved, err := repo.GetEventByID(ctx, testTenantID, event.ID)
	require.NoError(t, err)
	assert.Equal(t, 0, retrieved.RetryCount)
	assert.True(t, retrieved.PermanentlyFailed)
	assert.Nil(t, retrieved.NextRetryAt)
	assert.Equal(t, "authentication failed", *retrieved.RetryError)
}

func TestRepository_GetEventsForRetry(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	db := setupTestDB(t)
	defer db.Close()

	repo := NewRepository(db)
	ctx := context.Background()

	// Create test webhook
	webhook, err := repo.Create(ctx, testTenantID, testWorkflowID, "test-node", "secret", AuthTypeSignature)
	require.NoError(t, err)

	// Create multiple events with different retry states
	now := time.Now()

	// Event 1: Ready for retry (next_retry_at in past)
	event1 := &WebhookEvent{
		TenantID:       testTenantID,
		WebhookID:      webhook.ID,
		RequestMethod:  "POST",
		RequestHeaders: map[string]string{},
		RequestBody:    json.RawMessage(`{"event": 1}`),
		Status:         EventStatusFailed,
	}
	err = repo.CreateEvent(ctx, event1)
	require.NoError(t, err)
	pastTime := now.Add(-5 * time.Minute)
	_, err = db.ExecContext(ctx, "UPDATE webhook_events SET next_retry_at = $1, retry_count = 1 WHERE id = $2", pastTime, event1.ID)
	require.NoError(t, err)

	// Event 2: Not ready yet (next_retry_at in future)
	event2 := &WebhookEvent{
		TenantID:       testTenantID,
		WebhookID:      webhook.ID,
		RequestMethod:  "POST",
		RequestHeaders: map[string]string{},
		RequestBody:    json.RawMessage(`{"event": 2}`),
		Status:         EventStatusFailed,
	}
	err = repo.CreateEvent(ctx, event2)
	require.NoError(t, err)
	futureTime := now.Add(5 * time.Minute)
	_, err = db.ExecContext(ctx, "UPDATE webhook_events SET next_retry_at = $1, retry_count = 1 WHERE id = $2", futureTime, event2.ID)
	require.NoError(t, err)

	// Event 3: Permanently failed
	event3 := &WebhookEvent{
		TenantID:       testTenantID,
		WebhookID:      webhook.ID,
		RequestMethod:  "POST",
		RequestHeaders: map[string]string{},
		RequestBody:    json.RawMessage(`{"event": 3}`),
		Status:         EventStatusFailed,
	}
	err = repo.CreateEvent(ctx, event3)
	require.NoError(t, err)
	_, err = db.ExecContext(ctx, "UPDATE webhook_events SET permanently_failed = true, retry_count = 3 WHERE id = $1", event3.ID)
	require.NoError(t, err)

	// Event 4: Successfully processed
	event4 := &WebhookEvent{
		TenantID:       testTenantID,
		WebhookID:      webhook.ID,
		RequestMethod:  "POST",
		RequestHeaders: map[string]string{},
		RequestBody:    json.RawMessage(`{"event": 4}`),
		Status:         EventStatusProcessed,
	}
	err = repo.CreateEvent(ctx, event4)
	require.NoError(t, err)

	// Get events for retry
	events, err := repo.GetEventsForRetry(ctx, 100)
	require.NoError(t, err)

	// Should only return event1
	assert.Len(t, events, 1)
	assert.Equal(t, event1.ID, events[0].ID)
}

func TestRepository_MarkEventRetrySucceeded(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	db := setupTestDB(t)
	defer db.Close()

	repo := NewRepository(db)
	ctx := context.Background()

	// Create test webhook
	webhook, err := repo.Create(ctx, testTenantID, testWorkflowID, "test-node", "secret", AuthTypeSignature)
	require.NoError(t, err)

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

	// Mark for retry
	err = repo.MarkEventForRetry(ctx, event.ID, "initial error")
	require.NoError(t, err)

	// Mark retry as succeeded
	executionID := "test-execution-123"
	err = repo.MarkEventRetrySucceeded(ctx, event.ID, executionID, 250)
	require.NoError(t, err)

	// Verify status updated
	retrieved, err := repo.GetEventByID(ctx, testTenantID, event.ID)
	require.NoError(t, err)
	assert.Equal(t, EventStatusProcessed, retrieved.Status)
	assert.Equal(t, executionID, *retrieved.ExecutionID)
	assert.Equal(t, 250, *retrieved.ProcessingTimeMs)
	assert.Nil(t, retrieved.NextRetryAt)
}

func TestRepository_GetRetryStatistics(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	db := setupTestDB(t)
	defer db.Close()

	repo := NewRepository(db)
	ctx := context.Background()

	// Create test webhook
	webhook, err := repo.Create(ctx, testTenantID, testWorkflowID, "test-node", "secret", AuthTypeSignature)
	require.NoError(t, err)

	// Create events with various retry states
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
	_, err = db.ExecContext(ctx, "UPDATE webhook_events SET retry_count = 2, next_retry_at = NOW() + INTERVAL '5 minutes' WHERE id = $1", event1.ID)
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

	// Get statistics
	stats, err := repo.GetRetryStatistics(ctx, webhook.ID)
	require.NoError(t, err)

	assert.Equal(t, 2, stats.TotalRetriedEvents)
	assert.Equal(t, 1, stats.PermanentlyFailedEvents)
	assert.Equal(t, 1, stats.PendingRetries)
	assert.Equal(t, 3, stats.MaxRetryCount)
}
