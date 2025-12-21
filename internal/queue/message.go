package queue

import (
	"encoding/json"
	"fmt"
	"time"
)

// ExecutionMessage represents a workflow execution message in the queue
type ExecutionMessage struct {
	// Execution identifiers
	ExecutionID     string `json:"execution_id"`
	TenantID        string `json:"tenant_id"`
	WorkflowID      string `json:"workflow_id"`
	WorkflowVersion int    `json:"workflow_version"`

	// Trigger information
	TriggerType string          `json:"trigger_type"`
	TriggerData json.RawMessage `json:"trigger_data,omitempty"`

	// Message metadata
	EnqueuedAt time.Time `json:"enqueued_at"`
	RetryCount int       `json:"retry_count,omitempty"`

	// Tracing and correlation
	CorrelationID string `json:"correlation_id,omitempty"`
}

// NewExecutionMessage creates a new execution message
func NewExecutionMessage(executionID, tenantID, workflowID string, workflowVersion int, triggerType string, triggerData json.RawMessage) *ExecutionMessage {
	return &ExecutionMessage{
		ExecutionID:     executionID,
		TenantID:        tenantID,
		WorkflowID:      workflowID,
		WorkflowVersion: workflowVersion,
		TriggerType:     triggerType,
		TriggerData:     triggerData,
		EnqueuedAt:      time.Now().UTC(),
		RetryCount:      0,
	}
}

// Marshal serializes the execution message to JSON
func (m *ExecutionMessage) Marshal() (string, error) {
	data, err := json.Marshal(m)
	if err != nil {
		return "", fmt.Errorf("failed to marshal execution message: %w", err)
	}
	return string(data), nil
}

// UnmarshalExecutionMessage deserializes an execution message from JSON
func UnmarshalExecutionMessage(data string) (*ExecutionMessage, error) {
	var msg ExecutionMessage
	if err := json.Unmarshal([]byte(data), &msg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal execution message: %w", err)
	}
	return &msg, nil
}

// Validate checks if the execution message is valid
func (m *ExecutionMessage) Validate() error {
	if m.ExecutionID == "" {
		return fmt.Errorf("execution_id is required")
	}
	if m.TenantID == "" {
		return fmt.Errorf("tenant_id is required")
	}
	if m.WorkflowID == "" {
		return fmt.Errorf("workflow_id is required")
	}
	if m.WorkflowVersion <= 0 {
		return fmt.Errorf("workflow_version must be greater than 0")
	}
	if m.TriggerType == "" {
		return fmt.Errorf("trigger_type is required")
	}
	return nil
}

// GetMessageAttributes returns message attributes for SQS
func (m *ExecutionMessage) GetMessageAttributes() map[string]string {
	attrs := map[string]string{
		"tenant_id":    m.TenantID,
		"workflow_id":  m.WorkflowID,
		"trigger_type": m.TriggerType,
	}

	if m.CorrelationID != "" {
		attrs["correlation_id"] = m.CorrelationID
	}

	return attrs
}

// IncrementRetryCount increments the retry count
func (m *ExecutionMessage) IncrementRetryCount() {
	m.RetryCount++
}

// ShouldRetry determines if the message should be retried based on retry count
func (m *ExecutionMessage) ShouldRetry(maxRetries int) bool {
	return m.RetryCount < maxRetries
}
