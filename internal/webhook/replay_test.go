package webhook

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockWorkflowExecutor is a mock implementation of WorkflowExecutor
type MockWorkflowExecutor struct {
	mock.Mock
}

func (m *MockWorkflowExecutor) Execute(ctx context.Context, tenantID, workflowID, triggerType string, triggerData []byte) (string, error) {
	args := m.Called(ctx, tenantID, workflowID, triggerType, triggerData)
	return args.String(0), args.Error(1)
}

// MockReplayRepository is a mock implementation of Repository methods needed for replay
type MockReplayRepository struct {
	mock.Mock
}

func (m *MockReplayRepository) GetEventByID(ctx context.Context, tenantID, eventID string) (*WebhookEvent, error) {
	args := m.Called(ctx, tenantID, eventID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*WebhookEvent), args.Error(1)
}

func (m *MockReplayRepository) GetByID(ctx context.Context, id string) (*Webhook, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Webhook), args.Error(1)
}

func (m *MockReplayRepository) CreateEvent(ctx context.Context, event *WebhookEvent) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

func TestReplayService_ReplayEvent_Success(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockReplayRepository)
	mockExecutor := new(MockWorkflowExecutor)

	service := NewReplayService(mockRepo, mockExecutor, nil)

	tenantID := "tenant-123"
	eventID := "event-123"
	webhookID := "webhook-123"
	workflowID := "workflow-123"

	requestBody := json.RawMessage(`{"test": "data"}`)

	event := &WebhookEvent{
		ID:          eventID,
		TenantID:    tenantID,
		WebhookID:   webhookID,
		RequestBody: requestBody,
		ReplayCount: 0,
	}

	webhook := &Webhook{
		ID:         webhookID,
		WorkflowID: workflowID,
		TenantID:   tenantID,
		Enabled:    true,
	}

	mockRepo.On("GetEventByID", ctx, tenantID, eventID).Return(event, nil)
	mockRepo.On("GetByID", ctx, webhookID).Return(webhook, nil)
	mockExecutor.On("Execute", ctx, tenantID, workflowID, "webhook_replay", []byte(requestBody)).Return("exec-123", nil)
	mockRepo.On("CreateEvent", ctx, mock.AnythingOfType("*webhook.WebhookEvent")).Return(nil)

	result := service.ReplayEvent(ctx, tenantID, eventID, nil)

	assert.True(t, result.Success)
	assert.Equal(t, "exec-123", result.ExecutionID)
	assert.Empty(t, result.Error)
	mockRepo.AssertExpectations(t)
	mockExecutor.AssertExpectations(t)
}

func TestReplayService_ReplayEvent_WithModifiedPayload(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockReplayRepository)
	mockExecutor := new(MockWorkflowExecutor)

	service := NewReplayService(mockRepo, mockExecutor, nil)

	tenantID := "tenant-123"
	eventID := "event-123"
	webhookID := "webhook-123"
	workflowID := "workflow-123"

	originalPayload := json.RawMessage(`{"test": "data"}`)
	modifiedPayload := json.RawMessage(`{"test": "modified"}`)

	event := &WebhookEvent{
		ID:          eventID,
		TenantID:    tenantID,
		WebhookID:   webhookID,
		RequestBody: originalPayload,
		ReplayCount: 0,
	}

	webhook := &Webhook{
		ID:         webhookID,
		WorkflowID: workflowID,
		TenantID:   tenantID,
		Enabled:    true,
	}

	mockRepo.On("GetEventByID", ctx, tenantID, eventID).Return(event, nil)
	mockRepo.On("GetByID", ctx, webhookID).Return(webhook, nil)
	mockExecutor.On("Execute", ctx, tenantID, workflowID, "webhook_replay", []byte(modifiedPayload)).Return("exec-123", nil)
	mockRepo.On("CreateEvent", ctx, mock.AnythingOfType("*webhook.WebhookEvent")).Return(nil)

	result := service.ReplayEvent(ctx, tenantID, eventID, modifiedPayload)

	assert.True(t, result.Success)
	assert.Equal(t, "exec-123", result.ExecutionID)
	assert.Empty(t, result.Error)
	mockRepo.AssertExpectations(t)
	mockExecutor.AssertExpectations(t)
}

func TestReplayService_ReplayEvent_EventNotFound(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockReplayRepository)
	mockExecutor := new(MockWorkflowExecutor)

	service := NewReplayService(mockRepo, mockExecutor, nil)

	tenantID := "tenant-123"
	eventID := "event-123"

	mockRepo.On("GetEventByID", ctx, tenantID, eventID).Return(nil, ErrNotFound)

	result := service.ReplayEvent(ctx, tenantID, eventID, nil)

	assert.False(t, result.Success)
	assert.Empty(t, result.ExecutionID)
	assert.Contains(t, result.Error, "event not found")
	mockRepo.AssertExpectations(t)
	mockExecutor.AssertNotCalled(t, "Execute")
}

func TestReplayService_ReplayEvent_MaxReplayCountExceeded(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockReplayRepository)
	mockExecutor := new(MockWorkflowExecutor)

	service := NewReplayService(mockRepo, mockExecutor, nil)

	tenantID := "tenant-123"
	eventID := "event-123"
	webhookID := "webhook-123"

	event := &WebhookEvent{
		ID:          eventID,
		TenantID:    tenantID,
		WebhookID:   webhookID,
		RequestBody: json.RawMessage(`{"test": "data"}`),
		ReplayCount: 5, // Already at max
	}

	mockRepo.On("GetEventByID", ctx, tenantID, eventID).Return(event, nil)

	result := service.ReplayEvent(ctx, tenantID, eventID, nil)

	assert.False(t, result.Success)
	assert.Empty(t, result.ExecutionID)
	assert.Contains(t, result.Error, "max replay count")
	mockRepo.AssertExpectations(t)
	mockExecutor.AssertNotCalled(t, "Execute")
}

func TestReplayService_ReplayEvent_WebhookNotFound(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockReplayRepository)
	mockExecutor := new(MockWorkflowExecutor)

	service := NewReplayService(mockRepo, mockExecutor, nil)

	tenantID := "tenant-123"
	eventID := "event-123"
	webhookID := "webhook-123"

	event := &WebhookEvent{
		ID:          eventID,
		TenantID:    tenantID,
		WebhookID:   webhookID,
		RequestBody: json.RawMessage(`{"test": "data"}`),
		ReplayCount: 0,
	}

	mockRepo.On("GetEventByID", ctx, tenantID, eventID).Return(event, nil)
	mockRepo.On("GetByID", ctx, webhookID).Return(nil, ErrNotFound)

	result := service.ReplayEvent(ctx, tenantID, eventID, nil)

	assert.False(t, result.Success)
	assert.Empty(t, result.ExecutionID)
	assert.Contains(t, result.Error, "webhook not found")
	mockRepo.AssertExpectations(t)
	mockExecutor.AssertNotCalled(t, "Execute")
}

func TestReplayService_ReplayEvent_WebhookDisabled(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockReplayRepository)
	mockExecutor := new(MockWorkflowExecutor)

	service := NewReplayService(mockRepo, mockExecutor, nil)

	tenantID := "tenant-123"
	eventID := "event-123"
	webhookID := "webhook-123"
	workflowID := "workflow-123"

	event := &WebhookEvent{
		ID:          eventID,
		TenantID:    tenantID,
		WebhookID:   webhookID,
		RequestBody: json.RawMessage(`{"test": "data"}`),
		ReplayCount: 0,
	}

	webhook := &Webhook{
		ID:         webhookID,
		WorkflowID: workflowID,
		TenantID:   tenantID,
		Enabled:    false, // Disabled
	}

	mockRepo.On("GetEventByID", ctx, tenantID, eventID).Return(event, nil)
	mockRepo.On("GetByID", ctx, webhookID).Return(webhook, nil)

	result := service.ReplayEvent(ctx, tenantID, eventID, nil)

	assert.False(t, result.Success)
	assert.Empty(t, result.ExecutionID)
	assert.Contains(t, result.Error, "webhook is disabled")
	mockRepo.AssertExpectations(t)
	mockExecutor.AssertNotCalled(t, "Execute")
}

func TestReplayService_BatchReplayEvents_Success(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockReplayRepository)
	mockExecutor := new(MockWorkflowExecutor)

	service := NewReplayService(mockRepo, mockExecutor, nil)

	tenantID := "tenant-123"
	webhookID := "webhook-123"
	workflowID := "workflow-123"
	eventIDs := []string{"event-1", "event-2"}

	event1 := &WebhookEvent{
		ID:          "event-1",
		TenantID:    tenantID,
		WebhookID:   webhookID,
		RequestBody: json.RawMessage(`{"test": "data1"}`),
		ReplayCount: 0,
	}

	event2 := &WebhookEvent{
		ID:          "event-2",
		TenantID:    tenantID,
		WebhookID:   webhookID,
		RequestBody: json.RawMessage(`{"test": "data2"}`),
		ReplayCount: 0,
	}

	webhook := &Webhook{
		ID:         webhookID,
		WorkflowID: workflowID,
		TenantID:   tenantID,
		Enabled:    true,
	}

	mockRepo.On("GetEventByID", ctx, tenantID, "event-1").Return(event1, nil)
	mockRepo.On("GetEventByID", ctx, tenantID, "event-2").Return(event2, nil)
	mockRepo.On("GetByID", ctx, webhookID).Return(webhook, nil).Times(2)
	mockExecutor.On("Execute", ctx, tenantID, workflowID, "webhook_replay", []byte(event1.RequestBody)).Return("exec-1", nil)
	mockExecutor.On("Execute", ctx, tenantID, workflowID, "webhook_replay", []byte(event2.RequestBody)).Return("exec-2", nil)
	mockRepo.On("CreateEvent", ctx, mock.AnythingOfType("*webhook.WebhookEvent")).Return(nil).Times(2)

	results := service.BatchReplayEvents(ctx, tenantID, webhookID, eventIDs)

	assert.Len(t, results.Results, 2)
	assert.True(t, results.Results["event-1"].Success)
	assert.Equal(t, "exec-1", results.Results["event-1"].ExecutionID)
	assert.True(t, results.Results["event-2"].Success)
	assert.Equal(t, "exec-2", results.Results["event-2"].ExecutionID)
	mockRepo.AssertExpectations(t)
	mockExecutor.AssertExpectations(t)
}

func TestReplayService_BatchReplayEvents_MaxBatchSize(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockReplayRepository)
	mockExecutor := new(MockWorkflowExecutor)

	service := NewReplayService(mockRepo, mockExecutor, nil)

	tenantID := "tenant-123"
	webhookID := "webhook-123"

	// Create more than 10 event IDs
	eventIDs := make([]string, 11)
	for i := 0; i < 11; i++ {
		eventIDs[i] = "event-" + string(rune('0'+i))
	}

	results := service.BatchReplayEvents(ctx, tenantID, webhookID, eventIDs)

	// Should return error for all events due to batch size limit
	assert.Len(t, results.Results, 11)
	for _, result := range results.Results {
		assert.False(t, result.Success)
		assert.Contains(t, result.Error, "batch size exceeds maximum")
	}

	// No repository or executor calls should be made
	mockRepo.AssertNotCalled(t, "GetEventByID")
	mockExecutor.AssertNotCalled(t, "Execute")
}

func TestReplayService_BatchReplayEvents_PartialSuccess(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockReplayRepository)
	mockExecutor := new(MockWorkflowExecutor)

	service := NewReplayService(mockRepo, mockExecutor, nil)

	tenantID := "tenant-123"
	webhookID := "webhook-123"
	workflowID := "workflow-123"
	eventIDs := []string{"event-1", "event-2"}

	event1 := &WebhookEvent{
		ID:          "event-1",
		TenantID:    tenantID,
		WebhookID:   webhookID,
		RequestBody: json.RawMessage(`{"test": "data1"}`),
		ReplayCount: 0,
	}

	webhook := &Webhook{
		ID:         webhookID,
		WorkflowID: workflowID,
		TenantID:   tenantID,
		Enabled:    true,
	}

	// event-1 succeeds, event-2 not found
	mockRepo.On("GetEventByID", ctx, tenantID, "event-1").Return(event1, nil)
	mockRepo.On("GetEventByID", ctx, tenantID, "event-2").Return(nil, ErrNotFound)
	mockRepo.On("GetByID", ctx, webhookID).Return(webhook, nil).Once()
	mockExecutor.On("Execute", ctx, tenantID, workflowID, "webhook_replay", []byte(event1.RequestBody)).Return("exec-1", nil)
	mockRepo.On("CreateEvent", ctx, mock.AnythingOfType("*webhook.WebhookEvent")).Return(nil).Once()

	results := service.BatchReplayEvents(ctx, tenantID, webhookID, eventIDs)

	assert.Len(t, results.Results, 2)
	assert.True(t, results.Results["event-1"].Success)
	assert.False(t, results.Results["event-2"].Success)
	assert.Contains(t, results.Results["event-2"].Error, "event not found")
	mockRepo.AssertExpectations(t)
	mockExecutor.AssertExpectations(t)
}

func TestReplayService_ReplayEvent_ExecutionFailure(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockReplayRepository)
	mockExecutor := new(MockWorkflowExecutor)

	service := NewReplayService(mockRepo, mockExecutor, nil)

	tenantID := "tenant-123"
	eventID := "event-123"
	webhookID := "webhook-123"
	workflowID := "workflow-123"

	event := &WebhookEvent{
		ID:          eventID,
		TenantID:    tenantID,
		WebhookID:   webhookID,
		RequestBody: json.RawMessage(`{"test": "data"}`),
		ReplayCount: 0,
	}

	webhook := &Webhook{
		ID:         webhookID,
		WorkflowID: workflowID,
		TenantID:   tenantID,
		Enabled:    true,
	}

	mockRepo.On("GetEventByID", ctx, tenantID, eventID).Return(event, nil)
	mockRepo.On("GetByID", ctx, webhookID).Return(webhook, nil)
	mockExecutor.On("Execute", ctx, tenantID, workflowID, "webhook_replay", []byte(event.RequestBody)).
		Return("", assert.AnError)

	result := service.ReplayEvent(ctx, tenantID, eventID, nil)

	assert.False(t, result.Success)
	assert.Empty(t, result.ExecutionID)
	assert.Contains(t, result.Error, "execution failed")
	mockRepo.AssertExpectations(t)
	mockExecutor.AssertExpectations(t)
}

func TestReplayService_ReplayEvent_CreateEventFailure(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockReplayRepository)
	mockExecutor := new(MockWorkflowExecutor)

	service := NewReplayService(mockRepo, mockExecutor, nil)

	tenantID := "tenant-123"
	eventID := "event-123"
	webhookID := "webhook-123"
	workflowID := "workflow-123"

	requestBody := json.RawMessage(`{"test": "data"}`)

	event := &WebhookEvent{
		ID:          eventID,
		TenantID:    tenantID,
		WebhookID:   webhookID,
		RequestBody: requestBody,
		ReplayCount: 0,
	}

	webhook := &Webhook{
		ID:         webhookID,
		WorkflowID: workflowID,
		TenantID:   tenantID,
		Enabled:    true,
	}

	mockRepo.On("GetEventByID", ctx, tenantID, eventID).Return(event, nil)
	mockRepo.On("GetByID", ctx, webhookID).Return(webhook, nil)
	mockExecutor.On("Execute", ctx, tenantID, workflowID, "webhook_replay", []byte(requestBody)).Return("exec-123", nil)
	mockRepo.On("CreateEvent", ctx, mock.AnythingOfType("*webhook.WebhookEvent")).Return(assert.AnError)

	result := service.ReplayEvent(ctx, tenantID, eventID, nil)

	// Should still succeed even if CreateEvent fails
	assert.True(t, result.Success)
	assert.Equal(t, "exec-123", result.ExecutionID)
	assert.Empty(t, result.Error)
	mockRepo.AssertExpectations(t)
	mockExecutor.AssertExpectations(t)
}

func TestReplayService_ReplayEvent_EmptyModifiedPayload(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockReplayRepository)
	mockExecutor := new(MockWorkflowExecutor)

	service := NewReplayService(mockRepo, mockExecutor, nil)

	tenantID := "tenant-123"
	eventID := "event-123"
	webhookID := "webhook-123"
	workflowID := "workflow-123"

	originalPayload := json.RawMessage(`{"test": "data"}`)

	event := &WebhookEvent{
		ID:          eventID,
		TenantID:    tenantID,
		WebhookID:   webhookID,
		RequestBody: originalPayload,
		ReplayCount: 0,
	}

	webhook := &Webhook{
		ID:         webhookID,
		WorkflowID: workflowID,
		TenantID:   tenantID,
		Enabled:    true,
	}

	mockRepo.On("GetEventByID", ctx, tenantID, eventID).Return(event, nil)
	mockRepo.On("GetByID", ctx, webhookID).Return(webhook, nil)
	// Should use original payload when modified is empty
	mockExecutor.On("Execute", ctx, tenantID, workflowID, "webhook_replay", []byte(originalPayload)).Return("exec-123", nil)
	mockRepo.On("CreateEvent", ctx, mock.AnythingOfType("*webhook.WebhookEvent")).Return(nil)

	// Test with empty JSON array
	result := service.ReplayEvent(ctx, tenantID, eventID, json.RawMessage{})

	assert.True(t, result.Success)
	assert.Equal(t, "exec-123", result.ExecutionID)
	assert.Empty(t, result.Error)
	mockRepo.AssertExpectations(t)
	mockExecutor.AssertExpectations(t)
}

func TestReplayService_ReplayEvent_BoundaryReplayCount(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockReplayRepository)
	mockExecutor := new(MockWorkflowExecutor)

	service := NewReplayService(mockRepo, mockExecutor, nil)

	tenantID := "tenant-123"
	eventID := "event-123"
	webhookID := "webhook-123"
	workflowID := "workflow-123"

	requestBody := json.RawMessage(`{"test": "data"}`)

	// Test at MaxReplayCount - 1 (should succeed)
	event := &WebhookEvent{
		ID:          eventID,
		TenantID:    tenantID,
		WebhookID:   webhookID,
		RequestBody: requestBody,
		ReplayCount: MaxReplayCount - 1,
	}

	webhook := &Webhook{
		ID:         webhookID,
		WorkflowID: workflowID,
		TenantID:   tenantID,
		Enabled:    true,
	}

	mockRepo.On("GetEventByID", ctx, tenantID, eventID).Return(event, nil)
	mockRepo.On("GetByID", ctx, webhookID).Return(webhook, nil)
	mockExecutor.On("Execute", ctx, tenantID, workflowID, "webhook_replay", []byte(requestBody)).Return("exec-123", nil)
	mockRepo.On("CreateEvent", ctx, mock.AnythingOfType("*webhook.WebhookEvent")).Return(nil)

	result := service.ReplayEvent(ctx, tenantID, eventID, nil)

	assert.True(t, result.Success)
	assert.Equal(t, "exec-123", result.ExecutionID)
	assert.Empty(t, result.Error)
	mockRepo.AssertExpectations(t)
	mockExecutor.AssertExpectations(t)
}

func TestReplayService_BatchReplayEvents_EmptyList(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockReplayRepository)
	mockExecutor := new(MockWorkflowExecutor)

	service := NewReplayService(mockRepo, mockExecutor, nil)

	tenantID := "tenant-123"
	webhookID := "webhook-123"
	eventIDs := []string{}

	results := service.BatchReplayEvents(ctx, tenantID, webhookID, eventIDs)

	assert.Len(t, results.Results, 0)
	mockRepo.AssertNotCalled(t, "GetEventByID")
	mockExecutor.AssertNotCalled(t, "Execute")
}

func TestReplayService_BatchReplayEvents_ExactlyAtMaxBatchSize(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockReplayRepository)
	mockExecutor := new(MockWorkflowExecutor)

	service := NewReplayService(mockRepo, mockExecutor, nil)

	tenantID := "tenant-123"
	webhookID := "webhook-123"
	workflowID := "workflow-123"

	// Create exactly MaxBatchReplaySize events
	eventIDs := make([]string, MaxBatchReplaySize)
	for i := 0; i < MaxBatchReplaySize; i++ {
		eventID := "event-" + string(rune('A'+i))
		eventIDs[i] = eventID

		event := &WebhookEvent{
			ID:          eventID,
			TenantID:    tenantID,
			WebhookID:   webhookID,
			RequestBody: json.RawMessage(`{"test": "data"}`),
			ReplayCount: 0,
		}

		mockRepo.On("GetEventByID", ctx, tenantID, eventID).Return(event, nil)
	}

	webhook := &Webhook{
		ID:         webhookID,
		WorkflowID: workflowID,
		TenantID:   tenantID,
		Enabled:    true,
	}

	mockRepo.On("GetByID", ctx, webhookID).Return(webhook, nil).Times(MaxBatchReplaySize)
	mockExecutor.On("Execute", ctx, tenantID, workflowID, "webhook_replay", mock.Anything).
		Return("exec-123", nil).Times(MaxBatchReplaySize)
	mockRepo.On("CreateEvent", ctx, mock.AnythingOfType("*webhook.WebhookEvent")).
		Return(nil).Times(MaxBatchReplaySize)

	results := service.BatchReplayEvents(ctx, tenantID, webhookID, eventIDs)

	assert.Len(t, results.Results, MaxBatchReplaySize)
	for _, result := range results.Results {
		assert.True(t, result.Success)
		assert.Equal(t, "exec-123", result.ExecutionID)
	}
	mockRepo.AssertExpectations(t)
	mockExecutor.AssertExpectations(t)
}

func TestReplayService_BatchReplayEvents_AllEventsFail(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockReplayRepository)
	mockExecutor := new(MockWorkflowExecutor)

	service := NewReplayService(mockRepo, mockExecutor, nil)

	tenantID := "tenant-123"
	webhookID := "webhook-123"
	eventIDs := []string{"event-1", "event-2", "event-3"}

	// All events are at max replay count
	for _, eventID := range eventIDs {
		event := &WebhookEvent{
			ID:          eventID,
			TenantID:    tenantID,
			WebhookID:   webhookID,
			RequestBody: json.RawMessage(`{"test": "data"}`),
			ReplayCount: MaxReplayCount,
		}
		mockRepo.On("GetEventByID", ctx, tenantID, eventID).Return(event, nil)
	}

	results := service.BatchReplayEvents(ctx, tenantID, webhookID, eventIDs)

	assert.Len(t, results.Results, 3)
	for _, result := range results.Results {
		assert.False(t, result.Success)
		assert.Contains(t, result.Error, "max replay count")
	}
	mockRepo.AssertExpectations(t)
	mockExecutor.AssertNotCalled(t, "Execute")
}

// Table-driven tests for ReplayEvent error scenarios
func TestReplayService_ReplayEvent_TableDriven(t *testing.T) {
	tests := []struct {
		name           string
		setupMocks     func(*MockReplayRepository, *MockWorkflowExecutor)
		tenantID       string
		eventID        string
		modifiedPayload json.RawMessage
		wantSuccess    bool
		wantErrorContains string
		wantExecutionID string
	}{
		{
			name: "success with original payload",
			setupMocks: func(repo *MockReplayRepository, executor *MockWorkflowExecutor) {
				event := &WebhookEvent{
					ID:          "event-1",
					TenantID:    "tenant-1",
					WebhookID:   "webhook-1",
					RequestBody: json.RawMessage(`{"data": "original"}`),
					ReplayCount: 0,
				}
				webhook := &Webhook{
					ID:         "webhook-1",
					WorkflowID: "workflow-1",
					TenantID:   "tenant-1",
					Enabled:    true,
				}
				repo.On("GetEventByID", mock.Anything, "tenant-1", "event-1").Return(event, nil)
				repo.On("GetByID", mock.Anything, "webhook-1").Return(webhook, nil)
				executor.On("Execute", mock.Anything, "tenant-1", "workflow-1", "webhook_replay",
					[]byte(`{"data": "original"}`)).Return("exec-1", nil)
				repo.On("CreateEvent", mock.Anything, mock.AnythingOfType("*webhook.WebhookEvent")).Return(nil)
			},
			tenantID:       "tenant-1",
			eventID:        "event-1",
			wantSuccess:    true,
			wantExecutionID: "exec-1",
		},
		{
			name: "success with modified payload",
			setupMocks: func(repo *MockReplayRepository, executor *MockWorkflowExecutor) {
				event := &WebhookEvent{
					ID:          "event-2",
					TenantID:    "tenant-1",
					WebhookID:   "webhook-1",
					RequestBody: json.RawMessage(`{"data": "original"}`),
					ReplayCount: 0,
				}
				webhook := &Webhook{
					ID:         "webhook-1",
					WorkflowID: "workflow-1",
					TenantID:   "tenant-1",
					Enabled:    true,
				}
				repo.On("GetEventByID", mock.Anything, "tenant-1", "event-2").Return(event, nil)
				repo.On("GetByID", mock.Anything, "webhook-1").Return(webhook, nil)
				executor.On("Execute", mock.Anything, "tenant-1", "workflow-1", "webhook_replay",
					[]byte(`{"data": "modified"}`)).Return("exec-2", nil)
				repo.On("CreateEvent", mock.Anything, mock.AnythingOfType("*webhook.WebhookEvent")).Return(nil)
			},
			tenantID:        "tenant-1",
			eventID:         "event-2",
			modifiedPayload: json.RawMessage(`{"data": "modified"}`),
			wantSuccess:     true,
			wantExecutionID: "exec-2",
		},
		{
			name: "event not found",
			setupMocks: func(repo *MockReplayRepository, executor *MockWorkflowExecutor) {
				repo.On("GetEventByID", mock.Anything, "tenant-1", "event-404").Return(nil, ErrNotFound)
			},
			tenantID:          "tenant-1",
			eventID:           "event-404",
			wantSuccess:       false,
			wantErrorContains: "event not found",
		},
		{
			name: "replay count at max",
			setupMocks: func(repo *MockReplayRepository, executor *MockWorkflowExecutor) {
				event := &WebhookEvent{
					ID:          "event-max",
					TenantID:    "tenant-1",
					WebhookID:   "webhook-1",
					RequestBody: json.RawMessage(`{"data": "test"}`),
					ReplayCount: MaxReplayCount,
				}
				repo.On("GetEventByID", mock.Anything, "tenant-1", "event-max").Return(event, nil)
			},
			tenantID:          "tenant-1",
			eventID:           "event-max",
			wantSuccess:       false,
			wantErrorContains: "max replay count",
		},
		{
			name: "webhook not found",
			setupMocks: func(repo *MockReplayRepository, executor *MockWorkflowExecutor) {
				event := &WebhookEvent{
					ID:          "event-3",
					TenantID:    "tenant-1",
					WebhookID:   "webhook-404",
					RequestBody: json.RawMessage(`{"data": "test"}`),
					ReplayCount: 0,
				}
				repo.On("GetEventByID", mock.Anything, "tenant-1", "event-3").Return(event, nil)
				repo.On("GetByID", mock.Anything, "webhook-404").Return(nil, ErrNotFound)
			},
			tenantID:          "tenant-1",
			eventID:           "event-3",
			wantSuccess:       false,
			wantErrorContains: "webhook not found",
		},
		{
			name: "webhook disabled",
			setupMocks: func(repo *MockReplayRepository, executor *MockWorkflowExecutor) {
				event := &WebhookEvent{
					ID:          "event-4",
					TenantID:    "tenant-1",
					WebhookID:   "webhook-disabled",
					RequestBody: json.RawMessage(`{"data": "test"}`),
					ReplayCount: 0,
				}
				webhook := &Webhook{
					ID:         "webhook-disabled",
					WorkflowID: "workflow-1",
					TenantID:   "tenant-1",
					Enabled:    false,
				}
				repo.On("GetEventByID", mock.Anything, "tenant-1", "event-4").Return(event, nil)
				repo.On("GetByID", mock.Anything, "webhook-disabled").Return(webhook, nil)
			},
			tenantID:          "tenant-1",
			eventID:           "event-4",
			wantSuccess:       false,
			wantErrorContains: "webhook is disabled",
		},
		{
			name: "execution fails",
			setupMocks: func(repo *MockReplayRepository, executor *MockWorkflowExecutor) {
				event := &WebhookEvent{
					ID:          "event-5",
					TenantID:    "tenant-1",
					WebhookID:   "webhook-1",
					RequestBody: json.RawMessage(`{"data": "test"}`),
					ReplayCount: 0,
				}
				webhook := &Webhook{
					ID:         "webhook-1",
					WorkflowID: "workflow-1",
					TenantID:   "tenant-1",
					Enabled:    true,
				}
				repo.On("GetEventByID", mock.Anything, "tenant-1", "event-5").Return(event, nil)
				repo.On("GetByID", mock.Anything, "webhook-1").Return(webhook, nil)
				executor.On("Execute", mock.Anything, "tenant-1", "workflow-1", "webhook_replay",
					mock.Anything).Return("", assert.AnError)
			},
			tenantID:          "tenant-1",
			eventID:           "event-5",
			wantSuccess:       false,
			wantErrorContains: "execution failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			mockRepo := new(MockReplayRepository)
			mockExecutor := new(MockWorkflowExecutor)

			tt.setupMocks(mockRepo, mockExecutor)

			service := NewReplayService(mockRepo, mockExecutor, nil)
			result := service.ReplayEvent(ctx, tt.tenantID, tt.eventID, tt.modifiedPayload)

			assert.Equal(t, tt.wantSuccess, result.Success)
			assert.Equal(t, tt.wantExecutionID, result.ExecutionID)

			if tt.wantErrorContains != "" {
				assert.Contains(t, result.Error, tt.wantErrorContains)
			} else {
				assert.Empty(t, result.Error)
			}

			mockRepo.AssertExpectations(t)
			mockExecutor.AssertExpectations(t)
		})
	}
}

// Table-driven tests for BatchReplayEvents scenarios
func TestReplayService_BatchReplayEvents_TableDriven(t *testing.T) {
	tests := []struct {
		name              string
		eventIDs          []string
		setupMocks        func(*MockReplayRepository, *MockWorkflowExecutor)
		wantResultsCount  int
		validateResults   func(*testing.T, *BatchReplayResponse)
	}{
		{
			name:     "empty event list",
			eventIDs: []string{},
			setupMocks: func(repo *MockReplayRepository, executor *MockWorkflowExecutor) {
				// No setup needed
			},
			wantResultsCount: 0,
			validateResults: func(t *testing.T, response *BatchReplayResponse) {
				assert.Empty(t, response.Results)
			},
		},
		{
			name:     "single event success",
			eventIDs: []string{"event-1"},
			setupMocks: func(repo *MockReplayRepository, executor *MockWorkflowExecutor) {
				event := &WebhookEvent{
					ID:          "event-1",
					TenantID:    "tenant-1",
					WebhookID:   "webhook-1",
					RequestBody: json.RawMessage(`{"data": "test"}`),
					ReplayCount: 0,
				}
				webhook := &Webhook{
					ID:         "webhook-1",
					WorkflowID: "workflow-1",
					TenantID:   "tenant-1",
					Enabled:    true,
				}
				repo.On("GetEventByID", mock.Anything, "tenant-1", "event-1").Return(event, nil)
				repo.On("GetByID", mock.Anything, "webhook-1").Return(webhook, nil)
				executor.On("Execute", mock.Anything, "tenant-1", "workflow-1", "webhook_replay",
					mock.Anything).Return("exec-1", nil)
				repo.On("CreateEvent", mock.Anything, mock.AnythingOfType("*webhook.WebhookEvent")).Return(nil)
			},
			wantResultsCount: 1,
			validateResults: func(t *testing.T, response *BatchReplayResponse) {
				assert.True(t, response.Results["event-1"].Success)
				assert.Equal(t, "exec-1", response.Results["event-1"].ExecutionID)
			},
		},
		{
			name:     "exceeds max batch size",
			eventIDs: make([]string, MaxBatchReplaySize+1),
			setupMocks: func(repo *MockReplayRepository, executor *MockWorkflowExecutor) {
				// No setup needed - should fail early
			},
			wantResultsCount: MaxBatchReplaySize + 1,
			validateResults: func(t *testing.T, response *BatchReplayResponse) {
				for _, result := range response.Results {
					assert.False(t, result.Success)
					assert.Contains(t, result.Error, "batch size exceeds maximum")
				}
			},
		},
		{
			name:     "mixed success and failure",
			eventIDs: []string{"event-success", "event-max-replay", "event-not-found"},
			setupMocks: func(repo *MockReplayRepository, executor *MockWorkflowExecutor) {
				successEvent := &WebhookEvent{
					ID:          "event-success",
					TenantID:    "tenant-1",
					WebhookID:   "webhook-1",
					RequestBody: json.RawMessage(`{"data": "test"}`),
					ReplayCount: 0,
				}
				maxReplayEvent := &WebhookEvent{
					ID:          "event-max-replay",
					TenantID:    "tenant-1",
					WebhookID:   "webhook-1",
					RequestBody: json.RawMessage(`{"data": "test"}`),
					ReplayCount: MaxReplayCount,
				}
				webhook := &Webhook{
					ID:         "webhook-1",
					WorkflowID: "workflow-1",
					TenantID:   "tenant-1",
					Enabled:    true,
				}
				repo.On("GetEventByID", mock.Anything, "tenant-1", "event-success").Return(successEvent, nil)
				repo.On("GetEventByID", mock.Anything, "tenant-1", "event-max-replay").Return(maxReplayEvent, nil)
				repo.On("GetEventByID", mock.Anything, "tenant-1", "event-not-found").Return(nil, ErrNotFound)
				repo.On("GetByID", mock.Anything, "webhook-1").Return(webhook, nil).Once()
				executor.On("Execute", mock.Anything, "tenant-1", "workflow-1", "webhook_replay",
					mock.Anything).Return("exec-1", nil).Once()
				repo.On("CreateEvent", mock.Anything, mock.AnythingOfType("*webhook.WebhookEvent")).Return(nil).Once()
			},
			wantResultsCount: 3,
			validateResults: func(t *testing.T, response *BatchReplayResponse) {
				assert.True(t, response.Results["event-success"].Success)
				assert.False(t, response.Results["event-max-replay"].Success)
				assert.Contains(t, response.Results["event-max-replay"].Error, "max replay count")
				assert.False(t, response.Results["event-not-found"].Success)
				assert.Contains(t, response.Results["event-not-found"].Error, "event not found")
			},
		},
		{
			name:     "all events at max replay count",
			eventIDs: []string{"event-1", "event-2"},
			setupMocks: func(repo *MockReplayRepository, executor *MockWorkflowExecutor) {
				event1 := &WebhookEvent{
					ID:          "event-1",
					TenantID:    "tenant-1",
					WebhookID:   "webhook-1",
					RequestBody: json.RawMessage(`{"data": "test"}`),
					ReplayCount: MaxReplayCount,
				}
				event2 := &WebhookEvent{
					ID:          "event-2",
					TenantID:    "tenant-1",
					WebhookID:   "webhook-1",
					RequestBody: json.RawMessage(`{"data": "test"}`),
					ReplayCount: MaxReplayCount,
				}
				repo.On("GetEventByID", mock.Anything, "tenant-1", "event-1").Return(event1, nil)
				repo.On("GetEventByID", mock.Anything, "tenant-1", "event-2").Return(event2, nil)
			},
			wantResultsCount: 2,
			validateResults: func(t *testing.T, response *BatchReplayResponse) {
				for _, result := range response.Results {
					assert.False(t, result.Success)
					assert.Contains(t, result.Error, "max replay count")
				}
			},
		},
		{
			name:     "webhook disabled for all events",
			eventIDs: []string{"event-1", "event-2"},
			setupMocks: func(repo *MockReplayRepository, executor *MockWorkflowExecutor) {
				event1 := &WebhookEvent{
					ID:          "event-1",
					TenantID:    "tenant-1",
					WebhookID:   "webhook-disabled",
					RequestBody: json.RawMessage(`{"data": "test"}`),
					ReplayCount: 0,
				}
				event2 := &WebhookEvent{
					ID:          "event-2",
					TenantID:    "tenant-1",
					WebhookID:   "webhook-disabled",
					RequestBody: json.RawMessage(`{"data": "test"}`),
					ReplayCount: 0,
				}
				webhook := &Webhook{
					ID:         "webhook-disabled",
					WorkflowID: "workflow-1",
					TenantID:   "tenant-1",
					Enabled:    false,
				}
				repo.On("GetEventByID", mock.Anything, "tenant-1", "event-1").Return(event1, nil)
				repo.On("GetEventByID", mock.Anything, "tenant-1", "event-2").Return(event2, nil)
				repo.On("GetByID", mock.Anything, "webhook-disabled").Return(webhook, nil).Times(2)
			},
			wantResultsCount: 2,
			validateResults: func(t *testing.T, response *BatchReplayResponse) {
				for _, result := range response.Results {
					assert.False(t, result.Success)
					assert.Contains(t, result.Error, "webhook is disabled")
				}
			},
		},
		{
			name:     "execution fails for some events",
			eventIDs: []string{"event-success", "event-fail"},
			setupMocks: func(repo *MockReplayRepository, executor *MockWorkflowExecutor) {
				event1 := &WebhookEvent{
					ID:          "event-success",
					TenantID:    "tenant-1",
					WebhookID:   "webhook-1",
					RequestBody: json.RawMessage(`{"data": "test1"}`),
					ReplayCount: 0,
				}
				event2 := &WebhookEvent{
					ID:          "event-fail",
					TenantID:    "tenant-1",
					WebhookID:   "webhook-1",
					RequestBody: json.RawMessage(`{"data": "test2"}`),
					ReplayCount: 0,
				}
				webhook := &Webhook{
					ID:         "webhook-1",
					WorkflowID: "workflow-1",
					TenantID:   "tenant-1",
					Enabled:    true,
				}
				repo.On("GetEventByID", mock.Anything, "tenant-1", "event-success").Return(event1, nil)
				repo.On("GetEventByID", mock.Anything, "tenant-1", "event-fail").Return(event2, nil)
				repo.On("GetByID", mock.Anything, "webhook-1").Return(webhook, nil).Times(2)
				executor.On("Execute", mock.Anything, "tenant-1", "workflow-1", "webhook_replay",
					[]byte(`{"data": "test1"}`)).Return("exec-success", nil).Once()
				executor.On("Execute", mock.Anything, "tenant-1", "workflow-1", "webhook_replay",
					[]byte(`{"data": "test2"}`)).Return("", assert.AnError).Once()
				repo.On("CreateEvent", mock.Anything, mock.AnythingOfType("*webhook.WebhookEvent")).Return(nil).Once()
			},
			wantResultsCount: 2,
			validateResults: func(t *testing.T, response *BatchReplayResponse) {
				assert.True(t, response.Results["event-success"].Success)
				assert.Equal(t, "exec-success", response.Results["event-success"].ExecutionID)
				assert.False(t, response.Results["event-fail"].Success)
				assert.Contains(t, response.Results["event-fail"].Error, "execution failed")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			mockRepo := new(MockReplayRepository)
			mockExecutor := new(MockWorkflowExecutor)

			// Initialize event IDs if needed for batch size test
			if tt.name == "exceeds max batch size" {
				for i := 0; i < MaxBatchReplaySize+1; i++ {
					tt.eventIDs[i] = "event-" + string(rune('A'+i))
				}
			}

			tt.setupMocks(mockRepo, mockExecutor)

			service := NewReplayService(mockRepo, mockExecutor, nil)
			response := service.BatchReplayEvents(ctx, "tenant-1", "webhook-1", tt.eventIDs)

			assert.Len(t, response.Results, tt.wantResultsCount)
			tt.validateResults(t, response)

			mockRepo.AssertExpectations(t)
			mockExecutor.AssertExpectations(t)
		})
	}
}
