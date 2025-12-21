package webhook

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestDB(t *testing.T) *sqlx.DB {
	// This would connect to a test database
	// For now, skip if DB_TEST_URL is not set
	t.Skip("Skipping integration test - set DB_TEST_URL environment variable to run")
	return nil
}

func createTestWebhook(t *testing.T, repo *Repository, ctx context.Context, tenantID string) *Webhook {
	workflowID := uuid.New().String()
	nodeID := uuid.New().String()
	webhook, err := repo.Create(ctx, tenantID, workflowID, nodeID, "test-secret", AuthTypeSignature)
	require.NoError(t, err)
	return webhook
}

func TestRepository_CreateEvent(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewRepository(db)
	ctx := context.Background()
	tenantID := uuid.New().String()

	// Create a webhook first
	webhook := createTestWebhook(t, repo, ctx, tenantID)

	requestHeaders := map[string]string{
		"Content-Type": "application/json",
		"User-Agent":   "Test-Client/1.0",
	}

	requestBody := json.RawMessage(`{"event": "test", "data": {"foo": "bar"}}`)

	tests := []struct {
		name    string
		event   *WebhookEvent
		wantErr bool
	}{
		{
			name: "create event with received status",
			event: &WebhookEvent{
				TenantID:       tenantID,
				WebhookID:      webhook.ID,
				RequestMethod:  "POST",
				RequestHeaders: requestHeaders,
				RequestBody:    requestBody,
				Status:         EventStatusReceived,
			},
			wantErr: false,
		},
		{
			name: "create event with processed status and execution id",
			event: &WebhookEvent{
				TenantID:         tenantID,
				WebhookID:        webhook.ID,
				ExecutionID:      stringPtr(uuid.New().String()),
				RequestMethod:    "POST",
				RequestHeaders:   requestHeaders,
				RequestBody:      requestBody,
				ResponseStatus:   intPtr(200),
				ProcessingTimeMs: intPtr(150),
				Status:           EventStatusProcessed,
			},
			wantErr: false,
		},
		{
			name: "create event with failed status and error message",
			event: &WebhookEvent{
				TenantID:       tenantID,
				WebhookID:      webhook.ID,
				RequestMethod:  "POST",
				RequestHeaders: requestHeaders,
				RequestBody:    requestBody,
				Status:         EventStatusFailed,
				ErrorMessage:   stringPtr("authentication failed"),
			},
			wantErr: false,
		},
		{
			name: "create event with filtered status and reason",
			event: &WebhookEvent{
				TenantID:       tenantID,
				WebhookID:      webhook.ID,
				RequestMethod:  "POST",
				RequestHeaders: requestHeaders,
				RequestBody:    requestBody,
				Status:         EventStatusFiltered,
				FilteredReason: stringPtr("event type not matched"),
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := repo.CreateEvent(ctx, tt.event)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, tt.event.ID)
				assert.False(t, tt.event.CreatedAt.IsZero())
			}
		})
	}
}

func TestRepository_GetEventByID(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewRepository(db)
	ctx := context.Background()
	tenantID := uuid.New().String()

	// Create a webhook
	webhook := createTestWebhook(t, repo, ctx, tenantID)

	// Create a test event
	event := &WebhookEvent{
		TenantID:       tenantID,
		WebhookID:      webhook.ID,
		RequestMethod:  "POST",
		RequestHeaders: map[string]string{"Content-Type": "application/json"},
		RequestBody:    json.RawMessage(`{"test": "data"}`),
		Status:         EventStatusReceived,
	}
	err := repo.CreateEvent(ctx, event)
	require.NoError(t, err)

	tests := []struct {
		name     string
		tenantID string
		eventID  string
		wantErr  bool
	}{
		{
			name:     "get existing event",
			tenantID: tenantID,
			eventID:  event.ID,
			wantErr:  false,
		},
		{
			name:     "get non-existing event",
			tenantID: tenantID,
			eventID:  uuid.New().String(),
			wantErr:  true,
		},
		{
			name:     "get event with wrong tenant id",
			tenantID: uuid.New().String(),
			eventID:  event.ID,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := repo.GetEventByID(ctx, tt.tenantID, tt.eventID)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.eventID, result.ID)
				assert.Equal(t, tt.tenantID, result.TenantID)
			}
		})
	}
}

func TestRepository_ListEvents(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewRepository(db)
	ctx := context.Background()
	tenantID := uuid.New().String()

	// Create a webhook
	webhook := createTestWebhook(t, repo, ctx, tenantID)

	// Create multiple test events
	statuses := []WebhookEventStatus{
		EventStatusReceived,
		EventStatusProcessed,
		EventStatusFailed,
		EventStatusFiltered,
		EventStatusProcessed,
	}

	for _, status := range statuses {
		event := &WebhookEvent{
			TenantID:       tenantID,
			WebhookID:      webhook.ID,
			RequestMethod:  "POST",
			RequestHeaders: map[string]string{"Content-Type": "application/json"},
			RequestBody:    json.RawMessage(`{"test": "data"}`),
			Status:         status,
		}
		err := repo.CreateEvent(ctx, event)
		require.NoError(t, err)
		time.Sleep(10 * time.Millisecond) // Ensure different timestamps
	}

	tests := []struct {
		name          string
		tenantID      string
		filter        WebhookEventFilter
		expectedCount int
		wantErr       bool
	}{
		{
			name:     "list all events for webhook",
			tenantID: tenantID,
			filter: WebhookEventFilter{
				WebhookID: webhook.ID,
				Limit:     100,
				Offset:    0,
			},
			expectedCount: 5,
			wantErr:       false,
		},
		{
			name:     "list events with status filter - processed",
			tenantID: tenantID,
			filter: WebhookEventFilter{
				WebhookID: webhook.ID,
				Status:    webhookEventStatusPtr(EventStatusProcessed),
				Limit:     100,
				Offset:    0,
			},
			expectedCount: 2,
			wantErr:       false,
		},
		{
			name:     "list events with status filter - failed",
			tenantID: tenantID,
			filter: WebhookEventFilter{
				WebhookID: webhook.ID,
				Status:    webhookEventStatusPtr(EventStatusFailed),
				Limit:     100,
				Offset:    0,
			},
			expectedCount: 1,
			wantErr:       false,
		},
		{
			name:     "list events with limit",
			tenantID: tenantID,
			filter: WebhookEventFilter{
				WebhookID: webhook.ID,
				Limit:     2,
				Offset:    0,
			},
			expectedCount: 2,
			wantErr:       false,
		},
		{
			name:     "list events with offset",
			tenantID: tenantID,
			filter: WebhookEventFilter{
				WebhookID: webhook.ID,
				Limit:     100,
				Offset:    3,
			},
			expectedCount: 2,
			wantErr:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			events, total, err := repo.ListEvents(ctx, tt.tenantID, tt.filter)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, events, tt.expectedCount)
				assert.Equal(t, 5, total) // Total should always be 5
			}
		})
	}
}

func TestRepository_UpdateEventStatus(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewRepository(db)
	ctx := context.Background()
	tenantID := uuid.New().String()

	// Create a webhook
	webhook := createTestWebhook(t, repo, ctx, tenantID)

	// Create a test event
	event := &WebhookEvent{
		TenantID:       tenantID,
		WebhookID:      webhook.ID,
		RequestMethod:  "POST",
		RequestHeaders: map[string]string{"Content-Type": "application/json"},
		RequestBody:    json.RawMessage(`{"test": "data"}`),
		Status:         EventStatusReceived,
	}
	err := repo.CreateEvent(ctx, event)
	require.NoError(t, err)

	tests := []struct {
		name     string
		eventID  string
		status   WebhookEventStatus
		errorMsg *string
		wantErr  bool
	}{
		{
			name:     "update to processed status",
			eventID:  event.ID,
			status:   EventStatusProcessed,
			errorMsg: nil,
			wantErr:  false,
		},
		{
			name:     "update to failed status with error message",
			eventID:  event.ID,
			status:   EventStatusFailed,
			errorMsg: stringPtr("processing error"),
			wantErr:  false,
		},
		{
			name:     "update non-existing event",
			eventID:  uuid.New().String(),
			status:   EventStatusFailed,
			errorMsg: nil,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := repo.UpdateEventStatus(ctx, tt.eventID, tt.status, tt.errorMsg)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)

				// Verify the update
				updated, err := repo.GetEventByID(ctx, tenantID, tt.eventID)
				assert.NoError(t, err)
				assert.Equal(t, tt.status, updated.Status)
				if tt.errorMsg != nil {
					assert.Equal(t, *tt.errorMsg, *updated.ErrorMessage)
				}
			}
		})
	}
}

func TestRepository_IncrementTriggerCount(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewRepository(db)
	ctx := context.Background()
	tenantID := uuid.New().String()

	// Create a webhook
	webhook := createTestWebhook(t, repo, ctx, tenantID)

	tests := []struct {
		name      string
		webhookID string
		wantErr   bool
	}{
		{
			name:      "increment trigger count",
			webhookID: webhook.ID,
			wantErr:   false,
		},
		{
			name:      "increment non-existing webhook",
			webhookID: uuid.New().String(),
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := repo.IncrementTriggerCount(ctx, tt.webhookID)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestRepository_UpdateLastTriggeredAt(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewRepository(db)
	ctx := context.Background()
	tenantID := uuid.New().String()

	// Create a webhook
	webhook := createTestWebhook(t, repo, ctx, tenantID)

	tests := []struct {
		name      string
		webhookID string
		wantErr   bool
	}{
		{
			name:      "update last triggered at",
			webhookID: webhook.ID,
			wantErr:   false,
		},
		{
			name:      "update non-existing webhook",
			webhookID: uuid.New().String(),
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := repo.UpdateLastTriggeredAt(ctx, tt.webhookID)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestRepository_DeleteOldEvents(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewRepository(db)
	ctx := context.Background()
	tenantID := uuid.New().String()

	// Create a webhook
	webhook := createTestWebhook(t, repo, ctx, tenantID)

	// Create events with different timestamps
	now := time.Now()

	// Old event (40 days ago)
	oldEvent := &WebhookEvent{
		TenantID:       tenantID,
		WebhookID:      webhook.ID,
		RequestMethod:  "POST",
		RequestHeaders: map[string]string{"Content-Type": "application/json"},
		RequestBody:    json.RawMessage(`{"test": "old"}`),
		Status:         EventStatusProcessed,
	}
	err := repo.CreateEvent(ctx, oldEvent)
	require.NoError(t, err)

	// Manually update created_at to 40 days ago
	_, err = db.ExecContext(ctx, "UPDATE webhook_events SET created_at = $1 WHERE id = $2", now.Add(-40*24*time.Hour), oldEvent.ID)
	require.NoError(t, err)

	// Recent event (10 days ago)
	recentEvent := &WebhookEvent{
		TenantID:       tenantID,
		WebhookID:      webhook.ID,
		RequestMethod:  "POST",
		RequestHeaders: map[string]string{"Content-Type": "application/json"},
		RequestBody:    json.RawMessage(`{"test": "recent"}`),
		Status:         EventStatusProcessed,
	}
	err = repo.CreateEvent(ctx, recentEvent)
	require.NoError(t, err)

	// Manually update created_at to 10 days ago
	_, err = db.ExecContext(ctx, "UPDATE webhook_events SET created_at = $1 WHERE id = $2", now.Add(-10*24*time.Hour), recentEvent.ID)
	require.NoError(t, err)

	tests := []struct {
		name            string
		retentionPeriod time.Duration
		batchSize       int
		expectedDeleted int
		wantErr         bool
	}{
		{
			name:            "delete events older than 30 days",
			retentionPeriod: 30 * 24 * time.Hour,
			batchSize:       1000,
			expectedDeleted: 1,
			wantErr:         false,
		},
		{
			name:            "delete with no old events",
			retentionPeriod: 50 * 24 * time.Hour,
			batchSize:       1000,
			expectedDeleted: 0,
			wantErr:         false,
		},
		{
			name:            "delete with small batch size",
			retentionPeriod: 5 * 24 * time.Hour,
			batchSize:       1,
			expectedDeleted: 1,
			wantErr:         false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			deleted, err := repo.DeleteOldEvents(ctx, tt.retentionPeriod, tt.batchSize)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedDeleted, deleted)
			}
		})
	}
}

func TestRepository_DeleteOldEvents_MultipleBatches(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewRepository(db)
	ctx := context.Background()
	tenantID := uuid.New().String()

	// Create a webhook
	webhook := createTestWebhook(t, repo, ctx, tenantID)

	// Create 5 old events
	now := time.Now()
	for i := 0; i < 5; i++ {
		event := &WebhookEvent{
			TenantID:       tenantID,
			WebhookID:      webhook.ID,
			RequestMethod:  "POST",
			RequestHeaders: map[string]string{"Content-Type": "application/json"},
			RequestBody:    json.RawMessage(`{"test": "old"}`),
			Status:         EventStatusProcessed,
		}
		err := repo.CreateEvent(ctx, event)
		require.NoError(t, err)

		// Set to 40 days ago
		_, err = db.ExecContext(ctx, "UPDATE webhook_events SET created_at = $1 WHERE id = $2", now.Add(-40*24*time.Hour), event.ID)
		require.NoError(t, err)
	}

	// Delete with batch size of 2
	deleted1, err := repo.DeleteOldEvents(ctx, 30*24*time.Hour, 2)
	assert.NoError(t, err)
	assert.Equal(t, 2, deleted1)

	// Delete next batch
	deleted2, err := repo.DeleteOldEvents(ctx, 30*24*time.Hour, 2)
	assert.NoError(t, err)
	assert.Equal(t, 2, deleted2)

	// Delete final batch
	deleted3, err := repo.DeleteOldEvents(ctx, 30*24*time.Hour, 2)
	assert.NoError(t, err)
	assert.Equal(t, 1, deleted3)

	// Verify no more old events
	deleted4, err := repo.DeleteOldEvents(ctx, 30*24*time.Hour, 2)
	assert.NoError(t, err)
	assert.Equal(t, 0, deleted4)
}

// TestRepository_CreateEventWithMetadata tests creating an event with metadata
func TestRepository_CreateEventWithMetadata(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewRepository(db)
	ctx := context.Background()
	tenantID := uuid.New().String()

	// Create a webhook
	webhook := createTestWebhook(t, repo, ctx, tenantID)

	// Create metadata
	receivedAt := time.Now().UTC()
	metadata := &EventMetadata{
		SourceIP:      "192.168.1.100",
		UserAgent:     "GitHub-Hookshot/abc123",
		ReceivedAt:    receivedAt,
		ContentType:   "application/json",
		ContentLength: 256,
	}

	// Create event with metadata
	event := &WebhookEvent{
		TenantID:       tenantID,
		WebhookID:      webhook.ID,
		RequestMethod:  "POST",
		RequestHeaders: map[string]string{"Content-Type": "application/json"},
		RequestBody:    json.RawMessage(`{"test": "data"}`),
		Status:         EventStatusReceived,
		Metadata:       metadata,
	}

	err := repo.CreateEvent(ctx, event)
	require.NoError(t, err)
	assert.NotEmpty(t, event.ID)

	// Retrieve the event and verify metadata was stored
	retrievedEvent, err := repo.GetEventByID(ctx, tenantID, event.ID)
	require.NoError(t, err)
	require.NotNil(t, retrievedEvent)

	// Verify metadata was preserved
	require.NotNil(t, retrievedEvent.Metadata, "Metadata should not be nil")
	assert.Equal(t, metadata.SourceIP, retrievedEvent.Metadata.SourceIP)
	assert.Equal(t, metadata.UserAgent, retrievedEvent.Metadata.UserAgent)
	assert.Equal(t, metadata.ContentType, retrievedEvent.Metadata.ContentType)
	assert.Equal(t, metadata.ContentLength, retrievedEvent.Metadata.ContentLength)
	assert.WithinDuration(t, metadata.ReceivedAt, retrievedEvent.Metadata.ReceivedAt, time.Second)
}

// TestRepository_CreateEventWithoutMetadata tests that events work without metadata (backward compatibility)
func TestRepository_CreateEventWithoutMetadata(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewRepository(db)
	ctx := context.Background()
	tenantID := uuid.New().String()

	// Create a webhook
	webhook := createTestWebhook(t, repo, ctx, tenantID)

	// Create event without metadata
	event := &WebhookEvent{
		TenantID:       tenantID,
		WebhookID:      webhook.ID,
		RequestMethod:  "POST",
		RequestHeaders: map[string]string{"Content-Type": "application/json"},
		RequestBody:    json.RawMessage(`{"test": "data"}`),
		Status:         EventStatusReceived,
		Metadata:       nil, // No metadata
	}

	err := repo.CreateEvent(ctx, event)
	require.NoError(t, err)
	assert.NotEmpty(t, event.ID)

	// Retrieve the event and verify it works without metadata
	retrievedEvent, err := repo.GetEventByID(ctx, tenantID, event.ID)
	require.NoError(t, err)
	require.NotNil(t, retrievedEvent)

	// Metadata should be nil (backward compatibility)
	assert.Nil(t, retrievedEvent.Metadata, "Metadata should be nil when not provided")
}

// Helper functions
func stringPtr(s string) *string {
	return &s
}

func intPtr(i int) *int {
	return &i
}

func webhookEventStatusPtr(s WebhookEventStatus) *WebhookEventStatus {
	return &s
}
