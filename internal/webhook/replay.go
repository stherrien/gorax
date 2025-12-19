package webhook

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
)

const (
	// MaxReplayCount is the maximum number of times an event can be replayed
	MaxReplayCount = 5
	// MaxBatchReplaySize is the maximum number of events that can be replayed in a single batch
	MaxBatchReplaySize = 10
)

// WorkflowExecutor defines the interface for executing workflows
type WorkflowExecutor interface {
	Execute(ctx context.Context, tenantID, workflowID, triggerType string, triggerData []byte) (string, error)
}

// ReplayRepository defines the repository interface for replay operations
type ReplayRepository interface {
	GetEventByID(ctx context.Context, tenantID, eventID string) (*WebhookEvent, error)
	GetByID(ctx context.Context, id string) (*Webhook, error)
	CreateEvent(ctx context.Context, event *WebhookEvent) error
}

// ReplayService handles webhook event replay logic
type ReplayService struct {
	repo     ReplayRepository
	executor WorkflowExecutor
	logger   *slog.Logger
}

// NewReplayService creates a new replay service
func NewReplayService(repo ReplayRepository, executor WorkflowExecutor, logger *slog.Logger) *ReplayService {
	return &ReplayService{
		repo:     repo,
		executor: executor,
		logger:   logger,
	}
}

// ReplayEvent replays a single webhook event
func (s *ReplayService) ReplayEvent(ctx context.Context, tenantID, eventID string, modifiedPayload json.RawMessage) *ReplayResult {
	result := &ReplayResult{}

	// Get the original event
	event, err := s.repo.GetEventByID(ctx, tenantID, eventID)
	if err != nil {
		result.Success = false
		result.Error = fmt.Sprintf("event not found: %v", err)
		s.logError("failed to get event", err, eventID)
		return result
	}

	// Check replay count
	if event.ReplayCount >= MaxReplayCount {
		result.Success = false
		result.Error = fmt.Sprintf("max replay count (%d) exceeded", MaxReplayCount)
		s.logInfo("replay count exceeded", eventID, event.ReplayCount)
		return result
	}

	// Get the webhook to ensure it still exists and is enabled
	webhook, err := s.repo.GetByID(ctx, event.WebhookID)
	if err != nil {
		result.Success = false
		result.Error = fmt.Sprintf("webhook not found: %v", err)
		s.logError("failed to get webhook", err, event.WebhookID)
		return result
	}

	if !webhook.Enabled {
		result.Success = false
		result.Error = "webhook is disabled"
		s.logInfo("webhook is disabled", webhook.ID, webhook.Enabled)
		return result
	}

	// Determine payload to use
	payload := event.RequestBody
	if modifiedPayload != nil && len(modifiedPayload) > 0 {
		payload = modifiedPayload
	}

	// Execute the workflow
	executionID, err := s.executor.Execute(ctx, tenantID, webhook.WorkflowID, "webhook_replay", payload)
	if err != nil {
		result.Success = false
		result.Error = fmt.Sprintf("execution failed: %v", err)
		s.logError("workflow execution failed", err, webhook.WorkflowID)
		return result
	}

	// Create a new event record for the replay
	replayEvent := &WebhookEvent{
		TenantID:       tenantID,
		WebhookID:      event.WebhookID,
		RequestMethod:  event.RequestMethod,
		RequestHeaders: event.RequestHeaders,
		RequestBody:    payload,
		Status:         EventStatusReceived,
		ReplayCount:    event.ReplayCount + 1,
		SourceEventID:  &event.ID,
	}

	if err := s.repo.CreateEvent(ctx, replayEvent); err != nil {
		s.logError("failed to create replay event record", err, eventID)
		// Don't fail the replay if we can't create the event record
	}

	result.Success = true
	result.ExecutionID = executionID
	s.logSuccess("event replayed successfully", eventID, executionID)

	return result
}

// BatchReplayEvents replays multiple webhook events
func (s *ReplayService) BatchReplayEvents(ctx context.Context, tenantID, webhookID string, eventIDs []string) *BatchReplayResponse {
	response := &BatchReplayResponse{
		Results: make(map[string]*ReplayResult),
	}

	// Check batch size limit
	if len(eventIDs) > MaxBatchReplaySize {
		// Return error for all events
		for _, eventID := range eventIDs {
			response.Results[eventID] = &ReplayResult{
				Success: false,
				Error:   fmt.Sprintf("batch size exceeds maximum of %d events", MaxBatchReplaySize),
			}
		}
		s.logInfo("batch size exceeded", "batch_size", len(eventIDs))
		return response
	}

	// Replay each event
	for _, eventID := range eventIDs {
		result := s.ReplayEvent(ctx, tenantID, eventID, nil)
		response.Results[eventID] = result
	}

	return response
}

// Helper logging methods to reduce complexity

func (s *ReplayService) logError(msg string, err error, args ...interface{}) {
	if s.logger != nil {
		logArgs := []interface{}{"error", err}
		logArgs = append(logArgs, args...)
		s.logger.Error(msg, logArgs...)
	}
}

func (s *ReplayService) logInfo(msg string, args ...interface{}) {
	if s.logger != nil {
		s.logger.Info(msg, args...)
	}
}

func (s *ReplayService) logSuccess(msg string, eventID, executionID string) {
	if s.logger != nil {
		s.logger.Info(msg, "event_id", eventID, "execution_id", executionID)
	}
}
