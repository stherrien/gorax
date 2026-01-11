package webhook

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockWebhookDeliverer is a mock for webhook delivery
type MockWebhookDeliverer struct {
	mock.Mock
}

func (m *MockWebhookDeliverer) DeliverWebhook(ctx context.Context, webhook *Webhook, event *WebhookEvent) (*RetryResult, error) {
	args := m.Called(ctx, webhook, event)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*RetryResult), args.Error(1)
}

func TestRetryWorker_ProcessRetries(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	db := setupTestDB(t)
	defer db.Close()

	const testTenantID = "test-tenant-123"
	const testWorkflowID = "test-workflow-123"

	repo := NewRepository(db)
	ctx := context.Background()

	// Create test webhook
	webhook, err := repo.Create(ctx, testTenantID, testWorkflowID, "test-node", "secret", AuthTypeSignature)
	require.NoError(t, err)

	// Create failed event ready for retry
	event := &WebhookEvent{
		TenantID:       testTenantID,
		WebhookID:      webhook.ID,
		RequestMethod:  "POST",
		RequestHeaders: map[string]string{},
		RequestBody:    json.RawMessage(`{"test": "data"}`),
		Status:         EventStatusFailed,
	}
	err = repo.CreateEvent(ctx, event)
	require.NoError(t, err)

	// Mark for retry with past retry time
	pastTime := time.Now().Add(-5 * time.Minute)
	_, err = db.ExecContext(ctx, "UPDATE webhook_events SET next_retry_at = $1, retry_count = 1 WHERE id = $2", pastTime, event.ID)
	require.NoError(t, err)

	// Create mock deliverer
	mockDeliverer := new(MockWebhookDeliverer)
	mockDeliverer.On("DeliverWebhook", mock.Anything, mock.Anything, mock.Anything).
		Return(&RetryResult{
			Success:    true,
			StatusCode: 200,
		}, nil)

	// Create worker
	worker := NewRetryWorker(repo, mockDeliverer, DefaultRetryConfig())

	// Process retries
	processed, err := worker.ProcessRetries(ctx, 10)
	require.NoError(t, err)
	assert.Equal(t, 1, processed)

	// Verify event was processed
	retrieved, err := repo.GetEventByID(ctx, testTenantID, event.ID)
	require.NoError(t, err)
	assert.Equal(t, EventStatusProcessed, retrieved.Status)
	assert.Nil(t, retrieved.NextRetryAt)

	mockDeliverer.AssertExpectations(t)
}

func TestRetryWorker_ProcessRetries_RetryableFailure(t *testing.T) {
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

	// Create failed event ready for retry
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
	pastTime := time.Now().Add(-5 * time.Minute)
	_, err = db.ExecContext(ctx, "UPDATE webhook_events SET next_retry_at = $1, retry_count = 1 WHERE id = $2", pastTime, event.ID)
	require.NoError(t, err)

	// Create mock deliverer that fails with retryable error
	mockDeliverer := new(MockWebhookDeliverer)
	mockDeliverer.On("DeliverWebhook", mock.Anything, mock.Anything, mock.Anything).
		Return(nil, ErrWebhookTimeout)

	// Create worker
	worker := NewRetryWorker(repo, mockDeliverer, DefaultRetryConfig())

	// Process retries
	processed, err := worker.ProcessRetries(ctx, 10)
	require.NoError(t, err)
	assert.Equal(t, 1, processed)

	// Verify event is still failed but scheduled for next retry
	retrieved, err := repo.GetEventByID(ctx, testTenantID, event.ID)
	require.NoError(t, err)
	assert.Equal(t, EventStatusFailed, retrieved.Status)
	assert.Equal(t, 2, retrieved.RetryCount)
	assert.NotNil(t, retrieved.NextRetryAt)
	assert.True(t, retrieved.NextRetryAt.After(time.Now()))
	assert.False(t, retrieved.PermanentlyFailed)

	mockDeliverer.AssertExpectations(t)
}

func TestRetryWorker_ProcessRetries_NonRetryableFailure(t *testing.T) {
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

	// Create failed event ready for retry
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
	pastTime := time.Now().Add(-5 * time.Minute)
	_, err = db.ExecContext(ctx, "UPDATE webhook_events SET next_retry_at = $1, retry_count = 1 WHERE id = $2", pastTime, event.ID)
	require.NoError(t, err)

	// Create mock deliverer that fails with non-retryable error
	mockDeliverer := new(MockWebhookDeliverer)
	mockDeliverer.On("DeliverWebhook", mock.Anything, mock.Anything, mock.Anything).
		Return(nil, ErrWebhookAuthFailed)

	// Create worker
	worker := NewRetryWorker(repo, mockDeliverer, DefaultRetryConfig())

	// Process retries
	processed, err := worker.ProcessRetries(ctx, 10)
	require.NoError(t, err)
	assert.Equal(t, 1, processed)

	// Verify event is permanently failed
	retrieved, err := repo.GetEventByID(ctx, testTenantID, event.ID)
	require.NoError(t, err)
	assert.Equal(t, EventStatusFailed, retrieved.Status)
	assert.True(t, retrieved.PermanentlyFailed)
	assert.Nil(t, retrieved.NextRetryAt)

	mockDeliverer.AssertExpectations(t)
}

func TestRetryWorker_ProcessRetries_MaxRetriesExceeded(t *testing.T) {
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

	// Create failed event ready for retry
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

	// Set to max retries - 1
	pastTime := time.Now().Add(-5 * time.Minute)
	_, err = db.ExecContext(ctx, "UPDATE webhook_events SET next_retry_at = $1, retry_count = 2, max_retries = 3 WHERE id = $2", pastTime, event.ID)
	require.NoError(t, err)

	// Create mock deliverer that fails with retryable error
	mockDeliverer := new(MockWebhookDeliverer)
	mockDeliverer.On("DeliverWebhook", mock.Anything, mock.Anything, mock.Anything).
		Return(nil, ErrWebhookTimeout)

	// Create worker
	worker := NewRetryWorker(repo, mockDeliverer, DefaultRetryConfig())

	// Process retries (this will be the final retry)
	processed, err := worker.ProcessRetries(ctx, 10)
	require.NoError(t, err)
	assert.Equal(t, 1, processed)

	// Verify event is permanently failed
	retrieved, err := repo.GetEventByID(ctx, testTenantID, event.ID)
	require.NoError(t, err)
	assert.Equal(t, 3, retrieved.RetryCount)
	assert.True(t, retrieved.PermanentlyFailed)
	assert.Nil(t, retrieved.NextRetryAt)

	mockDeliverer.AssertExpectations(t)
}

func TestRetryWorker_ProcessRetries_BatchProcessing(t *testing.T) {
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

	// Create multiple failed events ready for retry
	numEvents := 5
	pastTime := time.Now().Add(-5 * time.Minute)
	for i := 0; i < numEvents; i++ {
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

		_, err = db.ExecContext(ctx, "UPDATE webhook_events SET next_retry_at = $1, retry_count = 1 WHERE id = $2", pastTime, event.ID)
		require.NoError(t, err)
	}

	// Create mock deliverer
	mockDeliverer := new(MockWebhookDeliverer)
	mockDeliverer.On("DeliverWebhook", mock.Anything, mock.Anything, mock.Anything).
		Return(&RetryResult{Success: true, StatusCode: 200}, nil).
		Times(numEvents)

	// Create worker
	worker := NewRetryWorker(repo, mockDeliverer, DefaultRetryConfig())

	// Process retries with batch size
	processed, err := worker.ProcessRetries(ctx, 10)
	require.NoError(t, err)
	assert.Equal(t, numEvents, processed)

	mockDeliverer.AssertExpectations(t)
}

func TestRetryWorker_Start_Stop(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	db := setupTestDB(t)
	defer db.Close()

	repo := NewRepository(db)
	mockDeliverer := new(MockWebhookDeliverer)

	config := DefaultRetryConfig()
	worker := NewRetryWorker(repo, mockDeliverer, config)

	// Start worker
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	errChan := make(chan error, 1)
	go func() {
		errChan <- worker.Start(ctx, 100*time.Millisecond)
	}()

	// Let it run for a short time
	time.Sleep(300 * time.Millisecond)

	// Stop worker
	cancel()

	// Wait for shutdown
	select {
	case err := <-errChan:
		assert.NoError(t, err)
	case <-time.After(2 * time.Second):
		t.Fatal("Worker did not stop in time")
	}
}

func TestClassifyWebhookError(t *testing.T) {
	tests := []struct {
		name          string
		err           error
		statusCode    int
		expectedRetry bool
	}{
		{
			name:          "nil error",
			err:           nil,
			statusCode:    200,
			expectedRetry: false,
		},
		{
			name:          "timeout error",
			err:           ErrWebhookTimeout,
			statusCode:    0,
			expectedRetry: true,
		},
		{
			name:          "connection failed",
			err:           ErrWebhookConnectionFailed,
			statusCode:    0,
			expectedRetry: true,
		},
		{
			name:          "server error 500",
			err:           ErrWebhookServerError,
			statusCode:    500,
			expectedRetry: true,
		},
		{
			name:          "rate limited 429",
			err:           ErrWebhookRateLimited,
			statusCode:    429,
			expectedRetry: true,
		},
		{
			name:          "auth failed 401",
			err:           ErrWebhookAuthFailed,
			statusCode:    401,
			expectedRetry: false,
		},
		{
			name:          "validation failed 400",
			err:           ErrWebhookValidationFailed,
			statusCode:    400,
			expectedRetry: false,
		},
		{
			name:          "generic 404",
			err:           errors.New("not found"),
			statusCode:    404,
			expectedRetry: false,
		},
		{
			name:          "generic 503",
			err:           errors.New("service unavailable"),
			statusCode:    503,
			expectedRetry: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			retryable := ClassifyWebhookError(tt.err, tt.statusCode)
			assert.Equal(t, tt.expectedRetry, retryable)
		})
	}
}
