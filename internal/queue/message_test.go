package queue

import (
	"encoding/json"
	"testing"
	"time"
)

func TestNewExecutionMessage(t *testing.T) {
	execID := "exec-123"
	tenantID := "tenant-456"
	workflowID := "workflow-789"
	workflowVersion := 1
	triggerType := "webhook"
	triggerData := json.RawMessage(`{"test": "data"}`)

	msg := NewExecutionMessage(execID, tenantID, workflowID, workflowVersion, triggerType, triggerData)

	if msg.ExecutionID != execID {
		t.Errorf("Expected ExecutionID %s, got %s", execID, msg.ExecutionID)
	}
	if msg.TenantID != tenantID {
		t.Errorf("Expected TenantID %s, got %s", tenantID, msg.TenantID)
	}
	if msg.WorkflowID != workflowID {
		t.Errorf("Expected WorkflowID %s, got %s", workflowID, msg.WorkflowID)
	}
	if msg.WorkflowVersion != workflowVersion {
		t.Errorf("Expected WorkflowVersion %d, got %d", workflowVersion, msg.WorkflowVersion)
	}
	if msg.TriggerType != triggerType {
		t.Errorf("Expected TriggerType %s, got %s", triggerType, msg.TriggerType)
	}
	if string(msg.TriggerData) != string(triggerData) {
		t.Errorf("Expected TriggerData %s, got %s", triggerData, msg.TriggerData)
	}
	if msg.RetryCount != 0 {
		t.Errorf("Expected RetryCount 0, got %d", msg.RetryCount)
	}
	if msg.EnqueuedAt.IsZero() {
		t.Error("Expected EnqueuedAt to be set")
	}
}

func TestExecutionMessage_Marshal(t *testing.T) {
	msg := NewExecutionMessage("exec-1", "tenant-1", "workflow-1", 1, "webhook", nil)

	data, err := msg.Marshal()
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	if data == "" {
		t.Error("Expected marshaled data, got empty string")
	}

	// Verify it's valid JSON
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(data), &result); err != nil {
		t.Errorf("Marshaled data is not valid JSON: %v", err)
	}
}

func TestUnmarshalExecutionMessage(t *testing.T) {
	original := NewExecutionMessage("exec-1", "tenant-1", "workflow-1", 1, "webhook", json.RawMessage(`{"test": "data"}`))
	original.RetryCount = 2

	data, err := original.Marshal()
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	unmarshaled, err := UnmarshalExecutionMessage(data)
	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if unmarshaled.ExecutionID != original.ExecutionID {
		t.Errorf("Expected ExecutionID %s, got %s", original.ExecutionID, unmarshaled.ExecutionID)
	}
	if unmarshaled.TenantID != original.TenantID {
		t.Errorf("Expected TenantID %s, got %s", original.TenantID, unmarshaled.TenantID)
	}
	if unmarshaled.WorkflowID != original.WorkflowID {
		t.Errorf("Expected WorkflowID %s, got %s", original.WorkflowID, unmarshaled.WorkflowID)
	}
	if unmarshaled.WorkflowVersion != original.WorkflowVersion {
		t.Errorf("Expected WorkflowVersion %d, got %d", original.WorkflowVersion, unmarshaled.WorkflowVersion)
	}
	if unmarshaled.TriggerType != original.TriggerType {
		t.Errorf("Expected TriggerType %s, got %s", original.TriggerType, unmarshaled.TriggerType)
	}
	if unmarshaled.RetryCount != original.RetryCount {
		t.Errorf("Expected RetryCount %d, got %d", original.RetryCount, unmarshaled.RetryCount)
	}
}

func TestExecutionMessage_Validate(t *testing.T) {
	tests := []struct {
		name    string
		msg     *ExecutionMessage
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid message",
			msg: &ExecutionMessage{
				ExecutionID:     "exec-1",
				TenantID:        "tenant-1",
				WorkflowID:      "workflow-1",
				WorkflowVersion: 1,
				TriggerType:     "webhook",
				EnqueuedAt:      time.Now(),
			},
			wantErr: false,
		},
		{
			name: "missing execution_id",
			msg: &ExecutionMessage{
				TenantID:        "tenant-1",
				WorkflowID:      "workflow-1",
				WorkflowVersion: 1,
				TriggerType:     "webhook",
			},
			wantErr: true,
			errMsg:  "execution_id is required",
		},
		{
			name: "missing tenant_id",
			msg: &ExecutionMessage{
				ExecutionID:     "exec-1",
				WorkflowID:      "workflow-1",
				WorkflowVersion: 1,
				TriggerType:     "webhook",
			},
			wantErr: true,
			errMsg:  "tenant_id is required",
		},
		{
			name: "missing workflow_id",
			msg: &ExecutionMessage{
				ExecutionID:     "exec-1",
				TenantID:        "tenant-1",
				WorkflowVersion: 1,
				TriggerType:     "webhook",
			},
			wantErr: true,
			errMsg:  "workflow_id is required",
		},
		{
			name: "invalid workflow_version",
			msg: &ExecutionMessage{
				ExecutionID:     "exec-1",
				TenantID:        "tenant-1",
				WorkflowID:      "workflow-1",
				WorkflowVersion: 0,
				TriggerType:     "webhook",
			},
			wantErr: true,
			errMsg:  "workflow_version must be greater than 0",
		},
		{
			name: "missing trigger_type",
			msg: &ExecutionMessage{
				ExecutionID:     "exec-1",
				TenantID:        "tenant-1",
				WorkflowID:      "workflow-1",
				WorkflowVersion: 1,
			},
			wantErr: true,
			errMsg:  "trigger_type is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.msg.Validate()
			if tt.wantErr {
				if err == nil {
					t.Error("Expected error, got nil")
				} else if err.Error() != tt.errMsg {
					t.Errorf("Expected error message %q, got %q", tt.errMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, got %v", err)
				}
			}
		})
	}
}

func TestExecutionMessage_GetMessageAttributes(t *testing.T) {
	msg := NewExecutionMessage("exec-1", "tenant-1", "workflow-1", 1, "webhook", nil)
	msg.CorrelationID = "corr-123"

	attrs := msg.GetMessageAttributes()

	expectedAttrs := map[string]string{
		"tenant_id":      "tenant-1",
		"workflow_id":    "workflow-1",
		"trigger_type":   "webhook",
		"correlation_id": "corr-123",
	}

	for key, expectedValue := range expectedAttrs {
		if attrs[key] != expectedValue {
			t.Errorf("Expected attribute %s to be %s, got %s", key, expectedValue, attrs[key])
		}
	}
}

func TestExecutionMessage_IncrementRetryCount(t *testing.T) {
	msg := NewExecutionMessage("exec-1", "tenant-1", "workflow-1", 1, "webhook", nil)

	if msg.RetryCount != 0 {
		t.Errorf("Expected initial RetryCount 0, got %d", msg.RetryCount)
	}

	msg.IncrementRetryCount()
	if msg.RetryCount != 1 {
		t.Errorf("Expected RetryCount 1, got %d", msg.RetryCount)
	}

	msg.IncrementRetryCount()
	if msg.RetryCount != 2 {
		t.Errorf("Expected RetryCount 2, got %d", msg.RetryCount)
	}
}

func TestExecutionMessage_ShouldRetry(t *testing.T) {
	msg := NewExecutionMessage("exec-1", "tenant-1", "workflow-1", 1, "webhook", nil)
	maxRetries := 3

	// Initial state
	if !msg.ShouldRetry(maxRetries) {
		t.Error("Expected ShouldRetry to be true")
	}

	// After 1 retry
	msg.IncrementRetryCount()
	if !msg.ShouldRetry(maxRetries) {
		t.Error("Expected ShouldRetry to be true")
	}

	// After 2 retries
	msg.IncrementRetryCount()
	if !msg.ShouldRetry(maxRetries) {
		t.Error("Expected ShouldRetry to be true")
	}

	// After 3 retries (at limit)
	msg.IncrementRetryCount()
	if msg.ShouldRetry(maxRetries) {
		t.Error("Expected ShouldRetry to be false")
	}

	// After 4 retries (exceeded limit)
	msg.IncrementRetryCount()
	if msg.ShouldRetry(maxRetries) {
		t.Error("Expected ShouldRetry to be false")
	}
}
