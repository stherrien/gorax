package executor

import (
	"context"
	"encoding/json"
	"log/slog"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gorax/gorax/internal/workflow"
)

// TestLoopIntegration_CompleteWorkflowWithLoop tests a complete workflow
// that fetches data and loops over items
func TestLoopIntegration_CompleteWorkflowWithLoop(t *testing.T) {
	// Create a simple workflow that:
	// 1. Receives trigger data with an array
	// 2. Extracts the array
	// 3. Loops over the array

	// Trigger data with test users
	triggerDataMap := map[string]interface{}{
		"users": []interface{}{
			map[string]interface{}{"id": float64(1), "name": "Alice", "score": float64(85)},
			map[string]interface{}{"id": float64(2), "name": "Bob", "score": float64(92)},
			map[string]interface{}{"id": float64(3), "name": "Charlie", "score": float64(78)},
		},
	}
	triggerDataJSON := mustMarshal(triggerDataMap)

	definition := workflow.WorkflowDefinition{
		Nodes: []workflow.Node{
			// Trigger node
			{
				ID:   "trigger-1",
				Type: string(workflow.NodeTypeTriggerWebhook),
				Data: workflow.NodeData{
					Name: "Manual Trigger",
					Config: mustMarshal(map[string]interface{}{
						"path": "/test",
					}),
				},
			},
			// Data source - extracts users from trigger
			{
				ID:   "data-source",
				Type: string(workflow.NodeTypeActionTransform),
				Data: workflow.NodeData{
					Name: "Extract Users",
					Config: mustMarshal(map[string]interface{}{
						"expression": "trigger.users",
					}),
				},
			},
			// Loop node (with empty body for simplicity)
			{
				ID:   "loop-1",
				Type: string(workflow.NodeTypeControlLoop),
				Data: workflow.NodeData{
					Name: "Process Each User",
					Config: mustMarshal(workflow.LoopActionConfig{
						Source:        "${steps.data-source}",
						ItemVariable:  "user",
						IndexVariable: "idx",
						MaxIterations: 1000,
						OnError:       "continue",
					}),
				},
			},
		},
		Edges: []workflow.Edge{
			{
				ID:     "e1",
				Source: "trigger-1",
				Target: "data-source",
			},
			{
				ID:     "e2",
				Source: "data-source",
				Target: "loop-1",
			},
		},
	}

	// Create mock workflow repository
	mockRepo := &mockWorkflowRepo{
		workflow: &workflow.Workflow{
			ID:         "wf-1",
			TenantID:   "tenant-1",
			Name:       "Loop Integration Test",
			Definition: mustMarshal(definition),
		},
		stepExecutions: make(map[string]*workflow.StepExecution),
	}

	// Create executor
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	executor := &Executor{
		repo:               mockRepo,
		logger:             logger,
		broadcaster:        nil,
		retryStrategy:      NewRetryStrategy(DefaultRetryConfig(), logger),
		circuitBreakers:    NewCircuitBreakerRegistry(DefaultCircuitBreakerConfig(), logger),
		defaultRetryConfig: DefaultNodeRetryConfig(),
	}

	// Create execution with trigger data
	execution := &workflow.Execution{
		ID:          "exec-1",
		TenantID:    "tenant-1",
		WorkflowID:  "wf-1",
		Status:      string(workflow.ExecutionStatusPending),
		TriggerType: "manual",
		TriggerData: &triggerDataJSON,
	}

	// Execute workflow
	err := executor.Execute(context.Background(), execution)
	require.NoError(t, err)

	// Verify execution completed successfully
	assert.Equal(t, string(workflow.ExecutionStatusCompleted), mockRepo.executionStatus)

	// Parse output data
	var outputData map[string]interface{}
	err = json.Unmarshal(mockRepo.executionOutput, &outputData)
	require.NoError(t, err)

	// Verify loop results
	loopOutput, exists := outputData["loop-1"]
	require.True(t, exists, "Loop output should exist")

	loopResult, ok := loopOutput.(map[string]interface{})
	require.True(t, ok, "Loop output should be a map")

	// Verify iteration count
	iterationCount, exists := loopResult["iteration_count"]
	require.True(t, exists)
	assert.Equal(t, float64(3), iterationCount, "Should have processed 3 users")

	// Verify iterations array
	iterations, exists := loopResult["iterations"]
	require.True(t, exists)

	iterationsArray, ok := iterations.([]interface{})
	require.True(t, ok, "Iterations should be an array")
	assert.Len(t, iterationsArray, 3, "Should have 3 iterations")

	// Verify first iteration
	firstIteration := iterationsArray[0].(map[string]interface{})
	assert.Equal(t, float64(0), firstIteration["index"])

	firstItem := firstIteration["item"].(map[string]interface{})
	assert.Equal(t, float64(1), firstItem["id"])
	assert.Equal(t, "Alice", firstItem["name"])

	// Verify second iteration
	secondIteration := iterationsArray[1].(map[string]interface{})
	assert.Equal(t, float64(1), secondIteration["index"])

	secondItem := secondIteration["item"].(map[string]interface{})
	assert.Equal(t, float64(2), secondItem["id"])
	assert.Equal(t, "Bob", secondItem["name"])

	// Verify third iteration
	thirdIteration := iterationsArray[2].(map[string]interface{})
	assert.Equal(t, float64(2), thirdIteration["index"])

	thirdItem := thirdIteration["item"].(map[string]interface{})
	assert.Equal(t, float64(3), thirdItem["id"])
	assert.Equal(t, "Charlie", thirdItem["name"])
}

// TestLoopIntegration_SimpleLoop tests a simple loop workflow
func TestLoopIntegration_SimpleLoop(t *testing.T) {
	// Simplest possible loop workflow - just loops over an array
	triggerDataMap := map[string]interface{}{
		"items": []interface{}{1, 2, 3},
	}
	triggerDataJSON := mustMarshal(triggerDataMap)

	definition := workflow.WorkflowDefinition{
		Nodes: []workflow.Node{
			{
				ID:   "trigger-1",
				Type: string(workflow.NodeTypeTriggerWebhook),
				Data: workflow.NodeData{
					Name:   "Manual Trigger",
					Config: mustMarshal(map[string]interface{}{}),
				},
			},
			{
				ID:   "data-source",
				Type: string(workflow.NodeTypeActionTransform),
				Data: workflow.NodeData{
					Name: "Extract Items",
					Config: mustMarshal(map[string]interface{}{
						"expression": "trigger.items",
					}),
				},
			},
			{
				ID:   "loop-1",
				Type: string(workflow.NodeTypeControlLoop),
				Data: workflow.NodeData{
					Name: "Process Items",
					Config: mustMarshal(workflow.LoopActionConfig{
						Source:        "${steps.data-source}",
						ItemVariable:  "item",
						IndexVariable: "idx",
						MaxIterations: 100,
						OnError:       "stop",
					}),
				},
			},
		},
		Edges: []workflow.Edge{
			{ID: "e1", Source: "trigger-1", Target: "data-source"},
			{ID: "e2", Source: "data-source", Target: "loop-1"},
		},
	}

	// Create mock repository
	mockRepo := &mockWorkflowRepo{
		workflow: &workflow.Workflow{
			ID:         "wf-2",
			TenantID:   "tenant-1",
			Name:       "Simple Loop Test",
			Definition: mustMarshal(definition),
		},
		stepExecutions: make(map[string]*workflow.StepExecution),
	}

	// Create executor
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	executor := &Executor{
		repo:               mockRepo,
		logger:             logger,
		broadcaster:        nil,
		retryStrategy:      NewRetryStrategy(DefaultRetryConfig(), logger),
		circuitBreakers:    NewCircuitBreakerRegistry(DefaultCircuitBreakerConfig(), logger),
		defaultRetryConfig: DefaultNodeRetryConfig(),
	}

	// Create execution
	execution := &workflow.Execution{
		ID:          "exec-2",
		TenantID:    "tenant-1",
		WorkflowID:  "wf-2",
		Status:      string(workflow.ExecutionStatusPending),
		TriggerType: "manual",
		TriggerData: &triggerDataJSON,
	}

	// Execute workflow
	err := executor.Execute(context.Background(), execution)
	require.NoError(t, err)

	// Verify execution completed
	assert.Equal(t, string(workflow.ExecutionStatusCompleted), mockRepo.executionStatus)

	// Parse output data
	var outputData map[string]interface{}
	err = json.Unmarshal(mockRepo.executionOutput, &outputData)
	require.NoError(t, err)

	// Verify loop ran
	loopOutput, exists := outputData["loop-1"]
	require.True(t, exists)

	loopResult, ok := loopOutput.(map[string]interface{})
	require.True(t, ok)

	// Should have processed 3 items
	assert.Equal(t, float64(3), loopResult["iteration_count"])
}

// TestLoopIntegration_ErrorHandling tests loop error handling strategies
func TestLoopIntegration_ErrorHandlingContinue(t *testing.T) {
	// Create workflow where some iterations fail but loop continues
	triggerDataMap := map[string]interface{}{
		"items": []interface{}{1, 2, 3},
	}
	triggerDataJSON := mustMarshal(triggerDataMap)

	definition := workflow.WorkflowDefinition{
		Nodes: []workflow.Node{
			{
				ID:   "trigger-1",
				Type: string(workflow.NodeTypeTriggerWebhook),
				Data: workflow.NodeData{
					Name:   "Manual Trigger",
					Config: mustMarshal(map[string]interface{}{}),
				},
			},
			{
				ID:   "data-source",
				Type: string(workflow.NodeTypeActionTransform),
				Data: workflow.NodeData{
					Name: "Extract Items",
					Config: mustMarshal(map[string]interface{}{
						"expression": "trigger.items",
					}),
				},
			},
			{
				ID:   "loop-1",
				Type: string(workflow.NodeTypeControlLoop),
				Data: workflow.NodeData{
					Name: "Process Items",
					Config: mustMarshal(workflow.LoopActionConfig{
						Source:        "${steps.data-source}",
						ItemVariable:  "item",
						IndexVariable: "idx",
						MaxIterations: 1000,
						OnError:       "continue", // Continue on error
					}),
				},
			},
		},
		Edges: []workflow.Edge{
			{ID: "e1", Source: "trigger-1", Target: "data-source"},
			{ID: "e2", Source: "data-source", Target: "loop-1"},
		},
	}

	mockRepo := &mockWorkflowRepo{
		workflow: &workflow.Workflow{
			ID:         "wf-3",
			TenantID:   "tenant-1",
			Name:       "Error Handling Test",
			Definition: mustMarshal(definition),
		},
		stepExecutions: make(map[string]*workflow.StepExecution),
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	executor := &Executor{
		repo:               mockRepo,
		logger:             logger,
		broadcaster:        nil,
		retryStrategy:      NewRetryStrategy(DefaultRetryConfig(), logger),
		circuitBreakers:    NewCircuitBreakerRegistry(DefaultCircuitBreakerConfig(), logger),
		defaultRetryConfig: DefaultNodeRetryConfig(),
	}

	execution := &workflow.Execution{
		ID:          "exec-3",
		TenantID:    "tenant-1",
		WorkflowID:  "wf-3",
		Status:      string(workflow.ExecutionStatusPending),
		TriggerType: "manual",
		TriggerData: &triggerDataJSON,
	}

	// Execute workflow
	err := executor.Execute(context.Background(), execution)
	require.NoError(t, err)

	// Verify execution completed (despite potential errors in iterations)
	assert.Equal(t, string(workflow.ExecutionStatusCompleted), mockRepo.executionStatus)
}

// TestLoopIntegration_MaxIterationsEnforcement tests that max iterations is enforced
func TestLoopIntegration_MaxIterationsEnforcement(t *testing.T) {
	triggerDataMap := map[string]interface{}{
		"items": []interface{}{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
	}
	triggerDataJSON := mustMarshal(triggerDataMap)

	definition := workflow.WorkflowDefinition{
		Nodes: []workflow.Node{
			{
				ID:   "trigger-1",
				Type: string(workflow.NodeTypeTriggerWebhook),
				Data: workflow.NodeData{
					Name:   "Manual Trigger",
					Config: mustMarshal(map[string]interface{}{}),
				},
			},
			{
				ID:   "data-source",
				Type: string(workflow.NodeTypeActionTransform),
				Data: workflow.NodeData{
					Name: "Extract Items",
					Config: mustMarshal(map[string]interface{}{
						"expression": "trigger.items",
					}),
				},
			},
			{
				ID:   "loop-1",
				Type: string(workflow.NodeTypeControlLoop),
				Data: workflow.NodeData{
					Name: "Process Items (Limited)",
					Config: mustMarshal(workflow.LoopActionConfig{
						Source:        "${steps.data-source}",
						ItemVariable:  "item",
						IndexVariable: "idx",
						MaxIterations: 5, // Limit to 5 iterations
						OnError:       "stop",
					}),
				},
			},
		},
		Edges: []workflow.Edge{
			{ID: "e1", Source: "trigger-1", Target: "data-source"},
			{ID: "e2", Source: "data-source", Target: "loop-1"},
		},
	}

	mockRepo := &mockWorkflowRepo{
		workflow: &workflow.Workflow{
			ID:         "wf-4",
			TenantID:   "tenant-1",
			Name:       "Max Iterations Test",
			Definition: mustMarshal(definition),
		},
		stepExecutions: make(map[string]*workflow.StepExecution),
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	executor := &Executor{
		repo:               mockRepo,
		logger:             logger,
		broadcaster:        nil,
		retryStrategy:      NewRetryStrategy(DefaultRetryConfig(), logger),
		circuitBreakers:    NewCircuitBreakerRegistry(DefaultCircuitBreakerConfig(), logger),
		defaultRetryConfig: DefaultNodeRetryConfig(),
	}

	execution := &workflow.Execution{
		ID:          "exec-4",
		TenantID:    "tenant-1",
		WorkflowID:  "wf-4",
		Status:      string(workflow.ExecutionStatusPending),
		TriggerType: "manual",
		TriggerData: &triggerDataJSON,
	}

	// Execute workflow - should fail because array length exceeds max iterations
	err := executor.Execute(context.Background(), execution)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "max iterations")

	// Verify execution failed
	assert.Equal(t, string(workflow.ExecutionStatusFailed), mockRepo.executionStatus)
}

// Mock repository for integration tests
type mockWorkflowRepo struct {
	workflow        *workflow.Workflow
	executionStatus string
	executionOutput json.RawMessage
	stepExecutions  map[string]*workflow.StepExecution
}

func (m *mockWorkflowRepo) GetByID(ctx context.Context, tenantID, id string) (*workflow.Workflow, error) {
	return m.workflow, nil
}

func (m *mockWorkflowRepo) UpdateExecutionStatus(ctx context.Context, id string, status string, outputData json.RawMessage, errorMsg *string) error {
	m.executionStatus = status
	m.executionOutput = outputData
	return nil
}

func (m *mockWorkflowRepo) CreateStepExecution(ctx context.Context, executionID, nodeID, nodeType string, inputData []byte) (*workflow.StepExecution, error) {
	step := &workflow.StepExecution{
		ID:          nodeID + "-step",
		ExecutionID: executionID,
		NodeID:      nodeID,
		NodeType:    nodeType,
		Status:      "running",
	}
	m.stepExecutions[step.ID] = step
	return step, nil
}

func (m *mockWorkflowRepo) UpdateStepExecution(ctx context.Context, id, status string, outputData json.RawMessage, errorMsg *string) error {
	if step, exists := m.stepExecutions[id]; exists {
		step.Status = status
		step.OutputData = &outputData
		step.ErrorMessage = errorMsg
	}
	return nil
}
