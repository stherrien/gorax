package websocket

import (
	"encoding/json"
	"time"
)

// EventType represents the type of execution event
type EventType string

const (
	EventTypeExecutionStarted   EventType = "execution.started"
	EventTypeExecutionCompleted EventType = "execution.completed"
	EventTypeExecutionFailed    EventType = "execution.failed"
	EventTypeStepStarted        EventType = "step.started"
	EventTypeStepCompleted      EventType = "step.completed"
	EventTypeStepFailed         EventType = "step.failed"
	EventTypeExecutionProgress  EventType = "execution.progress"
)

// ExecutionEvent represents a WebSocket event for execution updates
type ExecutionEvent struct {
	Type        EventType              `json:"type"`
	ExecutionID string                 `json:"execution_id"`
	WorkflowID  string                 `json:"workflow_id"`
	TenantID    string                 `json:"tenant_id"`
	Status      string                 `json:"status,omitempty"`
	Progress    *ProgressInfo          `json:"progress,omitempty"`
	Step        *StepInfo              `json:"step,omitempty"`
	Error       *string                `json:"error,omitempty"`
	Output      *json.RawMessage       `json:"output,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	Timestamp   time.Time              `json:"timestamp"`
}

// ProgressInfo contains execution progress information
type ProgressInfo struct {
	TotalSteps     int     `json:"total_steps"`
	CompletedSteps int     `json:"completed_steps"`
	Percentage     float64 `json:"percentage"`
}

// StepInfo contains step execution information
type StepInfo struct {
	StepID      string           `json:"step_id"`
	NodeID      string           `json:"node_id"`
	NodeType    string           `json:"node_type"`
	Status      string           `json:"status"`
	OutputData  *json.RawMessage `json:"output_data,omitempty"`
	ErrorMsg    *string          `json:"error,omitempty"`
	DurationMs  *int             `json:"duration_ms,omitempty"`
	StartedAt   *time.Time       `json:"started_at,omitempty"`
	CompletedAt *time.Time       `json:"completed_at,omitempty"`
}

// Broadcaster defines the interface for broadcasting execution events
type Broadcaster interface {
	// BroadcastExecutionStarted broadcasts when execution starts
	BroadcastExecutionStarted(tenantID, workflowID, executionID string, totalSteps int)

	// BroadcastExecutionCompleted broadcasts when execution completes
	BroadcastExecutionCompleted(tenantID, workflowID, executionID string, output json.RawMessage)

	// BroadcastExecutionFailed broadcasts when execution fails
	BroadcastExecutionFailed(tenantID, workflowID, executionID string, errorMsg string)

	// BroadcastStepStarted broadcasts when a step starts
	BroadcastStepStarted(tenantID, workflowID, executionID, nodeID, nodeType string)

	// BroadcastStepCompleted broadcasts when a step completes
	BroadcastStepCompleted(tenantID, workflowID, executionID, nodeID string, output json.RawMessage, durationMs int)

	// BroadcastStepFailed broadcasts when a step fails
	BroadcastStepFailed(tenantID, workflowID, executionID, nodeID string, errorMsg string)

	// BroadcastProgress broadcasts execution progress
	BroadcastProgress(tenantID, workflowID, executionID string, completedSteps, totalSteps int)
}

// HubBroadcaster implements Broadcaster using the WebSocket Hub
type HubBroadcaster struct {
	hub *Hub
}

// NewHubBroadcaster creates a new HubBroadcaster
func NewHubBroadcaster(hub *Hub) *HubBroadcaster {
	return &HubBroadcaster{hub: hub}
}

// Room helpers
func executionRoom(executionID string) string {
	return "execution:" + executionID
}

func workflowRoom(workflowID string) string {
	return "workflow:" + workflowID
}

func tenantRoom(tenantID string) string {
	return "tenant:" + tenantID
}

// BroadcastExecutionStarted broadcasts when execution starts
func (b *HubBroadcaster) BroadcastExecutionStarted(tenantID, workflowID, executionID string, totalSteps int) {
	event := ExecutionEvent{
		Type:        EventTypeExecutionStarted,
		ExecutionID: executionID,
		WorkflowID:  workflowID,
		TenantID:    tenantID,
		Status:      "running",
		Progress: &ProgressInfo{
			TotalSteps:     totalSteps,
			CompletedSteps: 0,
			Percentage:     0,
		},
		Timestamp: time.Now(),
	}

	b.broadcast(executionID, workflowID, tenantID, event)
}

// BroadcastExecutionCompleted broadcasts when execution completes
func (b *HubBroadcaster) BroadcastExecutionCompleted(tenantID, workflowID, executionID string, output json.RawMessage) {
	event := ExecutionEvent{
		Type:        EventTypeExecutionCompleted,
		ExecutionID: executionID,
		WorkflowID:  workflowID,
		TenantID:    tenantID,
		Status:      "completed",
		Output:      &output,
		Timestamp:   time.Now(),
	}

	b.broadcast(executionID, workflowID, tenantID, event)
}

// BroadcastExecutionFailed broadcasts when execution fails
func (b *HubBroadcaster) BroadcastExecutionFailed(tenantID, workflowID, executionID string, errorMsg string) {
	event := ExecutionEvent{
		Type:        EventTypeExecutionFailed,
		ExecutionID: executionID,
		WorkflowID:  workflowID,
		TenantID:    tenantID,
		Status:      "failed",
		Error:       &errorMsg,
		Timestamp:   time.Now(),
	}

	b.broadcast(executionID, workflowID, tenantID, event)
}

// BroadcastStepStarted broadcasts when a step starts
func (b *HubBroadcaster) BroadcastStepStarted(tenantID, workflowID, executionID, nodeID, nodeType string) {
	now := time.Now()
	event := ExecutionEvent{
		Type:        EventTypeStepStarted,
		ExecutionID: executionID,
		WorkflowID:  workflowID,
		TenantID:    tenantID,
		Step: &StepInfo{
			NodeID:    nodeID,
			NodeType:  nodeType,
			Status:    "running",
			StartedAt: &now,
		},
		Timestamp: now,
	}

	b.broadcast(executionID, workflowID, tenantID, event)
}

// BroadcastStepCompleted broadcasts when a step completes
func (b *HubBroadcaster) BroadcastStepCompleted(tenantID, workflowID, executionID, nodeID string, output json.RawMessage, durationMs int) {
	now := time.Now()
	event := ExecutionEvent{
		Type:        EventTypeStepCompleted,
		ExecutionID: executionID,
		WorkflowID:  workflowID,
		TenantID:    tenantID,
		Step: &StepInfo{
			NodeID:      nodeID,
			Status:      "completed",
			OutputData:  &output,
			DurationMs:  &durationMs,
			CompletedAt: &now,
		},
		Timestamp: now,
	}

	b.broadcast(executionID, workflowID, tenantID, event)
}

// BroadcastStepFailed broadcasts when a step fails
func (b *HubBroadcaster) BroadcastStepFailed(tenantID, workflowID, executionID, nodeID string, errorMsg string) {
	now := time.Now()
	event := ExecutionEvent{
		Type:        EventTypeStepFailed,
		ExecutionID: executionID,
		WorkflowID:  workflowID,
		TenantID:    tenantID,
		Step: &StepInfo{
			NodeID:      nodeID,
			Status:      "failed",
			ErrorMsg:    &errorMsg,
			CompletedAt: &now,
		},
		Timestamp: now,
	}

	b.broadcast(executionID, workflowID, tenantID, event)
}

// BroadcastProgress broadcasts execution progress
func (b *HubBroadcaster) BroadcastProgress(tenantID, workflowID, executionID string, completedSteps, totalSteps int) {
	percentage := 0.0
	if totalSteps > 0 {
		percentage = float64(completedSteps) / float64(totalSteps) * 100.0
	}

	event := ExecutionEvent{
		Type:        EventTypeExecutionProgress,
		ExecutionID: executionID,
		WorkflowID:  workflowID,
		TenantID:    tenantID,
		Progress: &ProgressInfo{
			TotalSteps:     totalSteps,
			CompletedSteps: completedSteps,
			Percentage:     percentage,
		},
		Timestamp: time.Now(),
	}

	b.broadcast(executionID, workflowID, tenantID, event)
}

// broadcast sends an event to all relevant rooms
func (b *HubBroadcaster) broadcast(executionID, workflowID, tenantID string, event ExecutionEvent) {
	data, err := json.Marshal(event)
	if err != nil {
		return
	}

	// Broadcast to execution-specific room
	b.hub.BroadcastToRoom(executionRoom(executionID), data)

	// Also broadcast to workflow room (for workflow monitoring)
	b.hub.BroadcastToRoom(workflowRoom(workflowID), data)

	// Also broadcast to tenant room (for dashboard)
	b.hub.BroadcastToRoom(tenantRoom(tenantID), data)
}
