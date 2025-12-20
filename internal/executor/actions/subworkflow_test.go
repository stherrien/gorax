package actions

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/gorax/gorax/internal/workflow"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockWorkflowRepository is a mock repository for testing
type MockWorkflowRepository struct {
	workflows  map[string]*workflow.Workflow
	executions map[string]*workflow.Execution
}

func NewMockWorkflowRepository() *MockWorkflowRepository {
	return &MockWorkflowRepository{
		workflows:  make(map[string]*workflow.Workflow),
		executions: make(map[string]*workflow.Execution),
	}
}

func (m *MockWorkflowRepository) GetByID(ctx context.Context, tenantID, workflowID string) (*workflow.Workflow, error) {
	if wf, ok := m.workflows[workflowID]; ok {
		return wf, nil
	}
	return nil, assert.AnError
}

func (m *MockWorkflowRepository) CreateExecution(ctx context.Context, execution *workflow.Execution) error {
	m.executions[execution.ID] = execution
	return nil
}

func (m *MockWorkflowRepository) GetExecutionByID(ctx context.Context, tenantID, executionID string) (*workflow.Execution, error) {
	if exec, ok := m.executions[executionID]; ok {
		return exec, nil
	}
	return nil, assert.AnError
}

// MockWorkflowExecutor is a mock executor for testing
type MockWorkflowExecutor struct {
	executeFunc func(ctx context.Context, execution *workflow.Execution) error
}

func (m *MockWorkflowExecutor) Execute(ctx context.Context, execution *workflow.Execution) error {
	if m.executeFunc != nil {
		return m.executeFunc(ctx, execution)
	}
	// Default: mark as completed
	execution.Status = string(workflow.ExecutionStatusCompleted)
	outputData := json.RawMessage(`{"result": "success"}`)
	execution.OutputData = &outputData
	return nil
}

// TestSubWorkflowAction_Execute_Success tests successful sub-workflow execution
func TestSubWorkflowAction_Execute_Success(t *testing.T) {
	// Setup
	mockRepo := NewMockWorkflowRepository()
	mockExec := &MockWorkflowExecutor{
		executeFunc: func(ctx context.Context, execution *workflow.Execution) error {
			// Simulate successful execution
			execution.Status = string(workflow.ExecutionStatusCompleted)
			outputData := json.RawMessage(`{"result": "success"}`)
			execution.OutputData = &outputData
			mockRepo.executions[execution.ID] = execution
			return nil
		},
	}

	// Create a simple workflow definition
	subWorkflowDef := workflow.WorkflowDefinition{
		Nodes: []workflow.Node{
			{
				ID:   "trigger-1",
				Type: string(workflow.NodeTypeTriggerWebhook),
			},
		},
		Edges: []workflow.Edge{},
	}
	defBytes, _ := json.Marshal(subWorkflowDef)

	mockRepo.workflows["sub-wf-1"] = &workflow.Workflow{
		ID:         "sub-wf-1",
		TenantID:   "tenant-1",
		Name:       "Sub Workflow",
		Definition: defBytes,
		Status:     string(workflow.WorkflowStatusActive),
		Version:    1,
	}

	config := &workflow.SubWorkflowConfig{
		WorkflowID: "sub-wf-1",
		InputMapping: map[string]string{
			"input_field": "${trigger.data}",
		},
		OutputMapping: map[string]string{
			"result": "${result}",
		},
		WaitForResult: true,
		TimeoutMs:     5000,
	}

	ctxData := map[string]interface{}{
		"trigger": map[string]interface{}{
			"data": "test-value",
		},
		"env": map[string]interface{}{
			"tenant_id":    "tenant-1",
			"execution_id": "parent-exec-1",
			"workflow_id":  "parent-wf-1",
		},
	}

	action := NewSubWorkflowAction(mockRepo, mockExec)
	input := NewActionInput(config, ctxData)

	// Execute
	output, err := action.Execute(context.Background(), input)

	// Assert
	require.NoError(t, err)
	assert.NotNil(t, output)
	assert.NotNil(t, output.Data)

	// Verify mapped output
	data := output.Data.(map[string]interface{})
	assert.Equal(t, "success", data["result"])
}

// TestSubWorkflowAction_InputMapping tests input parameter mapping
func TestSubWorkflowAction_InputMapping(t *testing.T) {
	tests := []struct {
		name          string
		inputMapping  map[string]string
		ctxData       map[string]interface{}
		expectedInput map[string]interface{}
	}{
		{
			name: "simple mapping",
			inputMapping: map[string]string{
				"field1": "${trigger.value}",
			},
			ctxData: map[string]interface{}{
				"trigger": map[string]interface{}{
					"value": "test",
				},
			},
			expectedInput: map[string]interface{}{
				"field1": "test",
			},
		},
		{
			name: "multiple mappings",
			inputMapping: map[string]string{
				"field1": "${trigger.value1}",
				"field2": "${steps.step1.output}",
			},
			ctxData: map[string]interface{}{
				"trigger": map[string]interface{}{
					"value1": "test1",
				},
				"steps": map[string]interface{}{
					"step1": map[string]interface{}{
						"output": "test2",
					},
				},
			},
			expectedInput: map[string]interface{}{
				"field1": "test1",
				"field2": "test2",
			},
		},
		{
			name: "nested path mapping",
			inputMapping: map[string]string{
				"userId": "${trigger.user.id}",
			},
			ctxData: map[string]interface{}{
				"trigger": map[string]interface{}{
					"user": map[string]interface{}{
						"id": "user-123",
					},
				},
			},
			expectedInput: map[string]interface{}{
				"userId": "user-123",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := NewMockWorkflowRepository()
			action := NewSubWorkflowAction(mockRepo, nil)

			result := action.mapInputs(tt.inputMapping, tt.ctxData)

			assert.Equal(t, tt.expectedInput, result)
		})
	}
}

// TestSubWorkflowAction_OutputMapping tests output parameter mapping
func TestSubWorkflowAction_OutputMapping(t *testing.T) {
	tests := []struct {
		name           string
		outputMapping  map[string]string
		subOutput      map[string]interface{}
		expectedOutput map[string]interface{}
	}{
		{
			name: "simple output mapping",
			outputMapping: map[string]string{
				"result": "${output.data}",
			},
			subOutput: map[string]interface{}{
				"output": map[string]interface{}{
					"data": "result-value",
				},
			},
			expectedOutput: map[string]interface{}{
				"result": "result-value",
			},
		},
		{
			name: "multiple output mappings",
			outputMapping: map[string]string{
				"status": "${output.status}",
				"count":  "${output.count}",
			},
			subOutput: map[string]interface{}{
				"output": map[string]interface{}{
					"status": "success",
					"count":  42,
				},
			},
			expectedOutput: map[string]interface{}{
				"status": "success",
				"count":  42,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := NewMockWorkflowRepository()
			action := NewSubWorkflowAction(mockRepo, nil)

			result := action.mapOutputs(tt.outputMapping, tt.subOutput)

			assert.Equal(t, tt.expectedOutput, result)
		})
	}
}

// TestSubWorkflowAction_CircularDependencyDetection tests circular workflow detection
func TestSubWorkflowAction_CircularDependencyDetection(t *testing.T) {
	mockRepo := NewMockWorkflowRepository()

	// Create workflow definitions
	wf1Def := workflow.WorkflowDefinition{
		Nodes: []workflow.Node{
			{
				ID:   "trigger-1",
				Type: string(workflow.NodeTypeTriggerWebhook),
			},
		},
	}
	def1Bytes, _ := json.Marshal(wf1Def)

	mockRepo.workflows["wf-1"] = &workflow.Workflow{
		ID:         "wf-1",
		TenantID:   "tenant-1",
		Name:       "Workflow 1",
		Definition: def1Bytes,
		Status:     string(workflow.WorkflowStatusActive),
		Version:    1,
	}

	config := &workflow.SubWorkflowConfig{
		WorkflowID:    "wf-1",
		WaitForResult: true,
	}

	// Context with workflow chain that already includes wf-1
	ctxData := map[string]interface{}{
		"_execution": map[string]interface{}{
			"workflow_chain": []string{"wf-root", "wf-1"},
		},
	}

	action := NewSubWorkflowAction(mockRepo, nil)
	input := NewActionInput(config, ctxData)

	// Execute
	_, err := action.Execute(context.Background(), input)

	// Assert - should detect circular dependency
	require.Error(t, err)
	assert.Contains(t, err.Error(), "circular")
}

// TestSubWorkflowAction_MaxDepthExceeded tests maximum depth protection
func TestSubWorkflowAction_MaxDepthExceeded(t *testing.T) {
	mockRepo := NewMockWorkflowRepository()

	config := &workflow.SubWorkflowConfig{
		WorkflowID:    "sub-wf-1",
		WaitForResult: true,
	}

	// Context with depth at maximum
	ctxData := map[string]interface{}{
		"_execution": map[string]interface{}{
			"depth": 10, // MaxSubWorkflowDepth
		},
	}

	action := NewSubWorkflowAction(mockRepo, nil)
	input := NewActionInput(config, ctxData)

	// Execute
	_, err := action.Execute(context.Background(), input)

	// Assert - should reject due to max depth
	require.Error(t, err)
	assert.Contains(t, err.Error(), "max depth")
}

// TestSubWorkflowAction_AsyncExecution tests async (fire-and-forget) execution
func TestSubWorkflowAction_AsyncExecution(t *testing.T) {
	mockRepo := NewMockWorkflowRepository()

	subWorkflowDef := workflow.WorkflowDefinition{
		Nodes: []workflow.Node{
			{
				ID:   "trigger-1",
				Type: string(workflow.NodeTypeTriggerWebhook),
			},
		},
	}
	defBytes, _ := json.Marshal(subWorkflowDef)

	mockRepo.workflows["sub-wf-1"] = &workflow.Workflow{
		ID:         "sub-wf-1",
		TenantID:   "tenant-1",
		Name:       "Sub Workflow",
		Definition: defBytes,
		Status:     string(workflow.WorkflowStatusActive),
		Version:    1,
	}

	config := &workflow.SubWorkflowConfig{
		WorkflowID:    "sub-wf-1",
		WaitForResult: false, // Async execution
	}

	ctxData := map[string]interface{}{
		"trigger": map[string]interface{}{
			"data": "test",
		},
	}

	action := NewSubWorkflowAction(mockRepo, nil)
	input := NewActionInput(config, ctxData)

	// Execute
	start := time.Now()
	output, err := action.Execute(context.Background(), input)
	duration := time.Since(start)

	// Assert - should return quickly without waiting
	require.NoError(t, err)
	assert.NotNil(t, output)
	assert.Less(t, duration, 1*time.Second, "async execution should not wait")

	// Output should contain execution ID
	data := output.Data.(map[string]interface{})
	assert.Contains(t, data, "execution_id")
	assert.Equal(t, "started", data["status"])
}

// TestSubWorkflowAction_SyncTimeout tests timeout for synchronous execution
func TestSubWorkflowAction_SyncTimeout(t *testing.T) {
	mockRepo := NewMockWorkflowRepository()

	// Create workflow that will timeout
	subWorkflowDef := workflow.WorkflowDefinition{
		Nodes: []workflow.Node{
			{
				ID:   "trigger-1",
				Type: string(workflow.NodeTypeTriggerWebhook),
			},
		},
	}
	defBytes, _ := json.Marshal(subWorkflowDef)

	mockRepo.workflows["sub-wf-1"] = &workflow.Workflow{
		ID:         "sub-wf-1",
		TenantID:   "tenant-1",
		Name:       "Sub Workflow",
		Definition: defBytes,
		Status:     string(workflow.WorkflowStatusActive),
		Version:    1,
	}

	// Mock executor that takes too long
	mockExec := &MockWorkflowExecutor{
		executeFunc: func(ctx context.Context, execution *workflow.Execution) error {
			time.Sleep(200 * time.Millisecond) // Sleep longer than timeout
			return nil
		},
	}

	config := &workflow.SubWorkflowConfig{
		WorkflowID:    "sub-wf-1",
		WaitForResult: true,
		TimeoutMs:     100, // Very short timeout
	}

	ctxData := map[string]interface{}{
		"trigger": map[string]interface{}{
			"data": "test",
		},
		"env": map[string]interface{}{
			"tenant_id":    "tenant-1",
			"execution_id": "parent-exec-1",
			"workflow_id":  "parent-wf-1",
		},
	}

	action := NewSubWorkflowAction(mockRepo, mockExec)
	input := NewActionInput(config, ctxData)

	// Execute
	_, err := action.Execute(context.Background(), input)

	// Assert - should timeout
	require.Error(t, err)
	assert.Contains(t, err.Error(), "timeout")
}

// TestSubWorkflowAction_MissingWorkflow tests error when workflow doesn't exist
func TestSubWorkflowAction_MissingWorkflow(t *testing.T) {
	mockRepo := NewMockWorkflowRepository()

	config := &workflow.SubWorkflowConfig{
		WorkflowID:    "non-existent",
		WaitForResult: true,
	}

	ctxData := map[string]interface{}{}

	action := NewSubWorkflowAction(mockRepo, nil)
	input := NewActionInput(config, ctxData)

	// Execute
	_, err := action.Execute(context.Background(), input)

	// Assert
	require.Error(t, err)
}

// TestSubWorkflowAction_TenantIsolation tests that sub-workflows inherit tenant context
func TestSubWorkflowAction_TenantIsolation(t *testing.T) {
	mockRepo := NewMockWorkflowRepository()

	subWorkflowDef := workflow.WorkflowDefinition{
		Nodes: []workflow.Node{
			{
				ID:   "trigger-1",
				Type: string(workflow.NodeTypeTriggerWebhook),
			},
		},
	}
	defBytes, _ := json.Marshal(subWorkflowDef)

	mockRepo.workflows["sub-wf-1"] = &workflow.Workflow{
		ID:         "sub-wf-1",
		TenantID:   "tenant-1",
		Name:       "Sub Workflow",
		Definition: defBytes,
		Status:     string(workflow.WorkflowStatusActive),
		Version:    1,
	}

	config := &workflow.SubWorkflowConfig{
		WorkflowID:    "sub-wf-1",
		WaitForResult: false,
	}

	ctxData := map[string]interface{}{
		"env": map[string]interface{}{
			"tenant_id": "tenant-1",
		},
	}

	action := NewSubWorkflowAction(mockRepo, nil)
	input := NewActionInput(config, ctxData)

	// Execute
	output, err := action.Execute(context.Background(), input)

	// Assert
	require.NoError(t, err)
	assert.NotNil(t, output)

	// Verify execution was created with correct tenant
	assert.Equal(t, 1, len(mockRepo.executions))
	for _, exec := range mockRepo.executions {
		assert.Equal(t, "tenant-1", exec.TenantID)
	}
}
