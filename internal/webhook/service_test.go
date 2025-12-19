package webhook

import (
	"context"
	"encoding/json"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupServiceTest(t *testing.T) (*Service, *Repository, string) {
	db := setupTestDB(t)
	repo := NewRepository(db)
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	service := NewService(repo, logger)
	tenantID := uuid.New().String()

	return service, repo, tenantID
}

func TestService_LogEvent(t *testing.T) {
	service, repo, tenantID := setupServiceTest(t)
	ctx := context.Background()

	// Create a webhook
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
			name: "log received event",
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
			name: "log processed event",
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.LogEvent(ctx, tt.event)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, tt.event.ID)
			}
		})
	}
}

func TestService_GetEvents(t *testing.T) {
	service, repo, tenantID := setupServiceTest(t)
	ctx := context.Background()

	// Create a webhook
	webhook := createTestWebhook(t, repo, ctx, tenantID)

	// Create multiple events
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
		err := service.LogEvent(ctx, event)
		require.NoError(t, err)
		time.Sleep(10 * time.Millisecond) // Ensure different timestamps
	}

	tests := []struct {
		name          string
		webhookID     string
		limit         int
		offset        int
		expectedCount int
		wantErr       bool
	}{
		{
			name:          "get all events",
			webhookID:     webhook.ID,
			limit:         100,
			offset:        0,
			expectedCount: 5,
			wantErr:       false,
		},
		{
			name:          "get events with limit",
			webhookID:     webhook.ID,
			limit:         2,
			offset:        0,
			expectedCount: 2,
			wantErr:       false,
		},
		{
			name:          "get events with offset",
			webhookID:     webhook.ID,
			limit:         100,
			offset:        3,
			expectedCount: 2,
			wantErr:       false,
		},
		{
			name:          "get events for non-existing webhook",
			webhookID:     uuid.New().String(),
			limit:         100,
			offset:        0,
			expectedCount: 0,
			wantErr:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			events, total, err := service.GetEvents(ctx, tenantID, tt.webhookID, tt.limit, tt.offset)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, events, tt.expectedCount)
				if tt.webhookID == webhook.ID {
					assert.Equal(t, 5, total)
				} else {
					assert.Equal(t, 0, total)
				}
			}
		})
	}
}

func TestService_MarkEventProcessed(t *testing.T) {
	service, repo, tenantID := setupServiceTest(t)
	ctx := context.Background()

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
	err := service.LogEvent(ctx, event)
	require.NoError(t, err)

	executionID := uuid.New().String()
	processingTimeMs := 250

	tests := []struct {
		name             string
		eventID          string
		executionID      string
		processingTimeMs int
		wantErr          bool
	}{
		{
			name:             "mark event as processed",
			eventID:          event.ID,
			executionID:      executionID,
			processingTimeMs: processingTimeMs,
			wantErr:          false,
		},
		{
			name:             "mark non-existing event",
			eventID:          uuid.New().String(),
			executionID:      executionID,
			processingTimeMs: processingTimeMs,
			wantErr:          true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.MarkEventProcessed(ctx, tt.eventID, tt.executionID, tt.processingTimeMs)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)

				// Verify the event was updated
				updated, err := repo.GetEventByID(ctx, tenantID, tt.eventID)
				assert.NoError(t, err)
				assert.Equal(t, EventStatusProcessed, updated.Status)
				assert.NotNil(t, updated.ExecutionID)
				assert.Equal(t, tt.executionID, *updated.ExecutionID)
				assert.NotNil(t, updated.ProcessingTimeMs)
				assert.Equal(t, tt.processingTimeMs, *updated.ProcessingTimeMs)
			}
		})
	}
}

func TestService_MarkEventFailed(t *testing.T) {
	service, repo, tenantID := setupServiceTest(t)
	ctx := context.Background()

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
	err := service.LogEvent(ctx, event)
	require.NoError(t, err)

	errorMsg := "authentication failed"

	tests := []struct {
		name     string
		eventID  string
		errorMsg string
		wantErr  bool
	}{
		{
			name:     "mark event as failed",
			eventID:  event.ID,
			errorMsg: errorMsg,
			wantErr:  false,
		},
		{
			name:     "mark non-existing event",
			eventID:  uuid.New().String(),
			errorMsg: errorMsg,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.MarkEventFailed(ctx, tt.eventID, tt.errorMsg)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)

				// Verify the event was updated
				updated, err := repo.GetEventByID(ctx, tenantID, tt.eventID)
				assert.NoError(t, err)
				assert.Equal(t, EventStatusFailed, updated.Status)
				assert.NotNil(t, updated.ErrorMessage)
				assert.Equal(t, tt.errorMsg, *updated.ErrorMessage)
			}
		})
	}
}

func TestService_MarkEventFiltered(t *testing.T) {
	service, repo, tenantID := setupServiceTest(t)
	ctx := context.Background()

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
	err := service.LogEvent(ctx, event)
	require.NoError(t, err)

	reason := "event type not matched"

	tests := []struct {
		name    string
		eventID string
		reason  string
		wantErr bool
	}{
		{
			name:    "mark event as filtered",
			eventID: event.ID,
			reason:  reason,
			wantErr: false,
		},
		{
			name:    "mark non-existing event",
			eventID: uuid.New().String(),
			reason:  reason,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.MarkEventFiltered(ctx, tt.eventID, tt.reason)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)

				// Verify the event was updated
				updated, err := repo.GetEventByID(ctx, tenantID, tt.eventID)
				assert.NoError(t, err)
				assert.Equal(t, EventStatusFiltered, updated.Status)
				assert.NotNil(t, updated.FilteredReason)
				assert.Equal(t, tt.reason, *updated.FilteredReason)
			}
		})
	}
}
