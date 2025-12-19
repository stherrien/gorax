package schedule

import (
	"context"
)

// WorkflowServiceAdapter adapts workflow service for scheduler
type WorkflowServiceAdapter struct {
	executeFunc func(ctx context.Context, tenantID, workflowID, triggerType string, triggerData []byte) (executionID string, err error)
}

// NewWorkflowServiceAdapter creates a new adapter
func NewWorkflowServiceAdapter(executeFunc func(ctx context.Context, tenantID, workflowID, triggerType string, triggerData []byte) (executionID string, err error)) *WorkflowServiceAdapter {
	return &WorkflowServiceAdapter{
		executeFunc: executeFunc,
	}
}

// ExecuteScheduled executes a scheduled workflow
func (w *WorkflowServiceAdapter) ExecuteScheduled(ctx context.Context, tenantID, workflowID, scheduleID string) (executionID string, err error) {
	// Create trigger data with schedule information
	triggerData := []byte(`{"schedule_id":"` + scheduleID + `"}`)
	return w.executeFunc(ctx, tenantID, workflowID, "schedule", triggerData)
}
