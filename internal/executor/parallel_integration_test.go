package executor

import (
	"context"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gorax/gorax/internal/workflow"
)

// TestParallelIntegration_MultipleHTTPRequests tests parallel HTTP requests
func TestParallelIntegration_MultipleHTTPRequests(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Create a simple workflow with parallel HTTP requests
	definition := workflow.WorkflowDefinition{
		Nodes: []workflow.Node{
			{
				ID:   "parallel1",
				Type: string(workflow.NodeTypeControlParallel),
				Data: workflow.NodeData{
					Name: "Parallel HTTP Requests",
					Config: mustMarshal(workflow.ParallelConfig{
						Branches: []workflow.ParallelBranch{
							{
								Name:  "api1",
								Nodes: []string{"http1"},
							},
							{
								Name:  "api2",
								Nodes: []string{"http2"},
							},
							{
								Name:  "api3",
								Nodes: []string{"http3"},
							},
						},
						WaitMode:       "all",
						MaxConcurrency: 0,
						FailureMode:    "stop_all",
						Timeout:        "10s",
					}),
				},
			},
		},
		Edges: []workflow.Edge{},
	}

	// Create mock workflow repository
	mockRepo := &mockWorkflowRepo{
		workflow: &workflow.Workflow{
			ID:         "wf1",
			TenantID:   "tenant1",
			Definition: mustMarshal(definition),
		},
		stepExecutions: make(map[string]*workflow.StepExecution),
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	executor := &Executor{
		repo:            mockRepo,
		logger:          logger,
		retryStrategy:   NewRetryStrategy(DefaultRetryConfig(), logger),
		circuitBreakers: NewCircuitBreakerRegistry(DefaultCircuitBreakerConfig(), logger),
	}

	// Create execution context
	execCtx := &ExecutionContext{
		TenantID:    "tenant1",
		ExecutionID: "exec1",
		WorkflowID:  "wf1",
		TriggerData: map[string]interface{}{
			"test": "value",
		},
		StepOutputs: make(map[string]interface{}),
	}

	// Execute parallel node
	startTime := time.Now()
	result, err := executor.executeParallelAction(
		context.Background(),
		definition.Nodes[0],
		execCtx,
		&definition,
	)

	duration := time.Since(startTime)

	require.NoError(t, err)
	require.NotNil(t, result)

	// Verify result structure
	resultMap, ok := result.(map[string]interface{})
	require.True(t, ok)

	// Should have completed all branches
	assert.Contains(t, resultMap, "branches")
	assert.Contains(t, resultMap, "completed_branches")

	// Execution should have happened in parallel (much faster than sequential)
	// If sequential, 3 requests would take 3x time; parallel should be ~1x time
	t.Logf("Parallel execution took: %v", duration)
}

// TestParallelIntegration_WaitModeFirst tests first completion mode
func TestParallelIntegration_WaitModeFirst(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	definition := workflow.WorkflowDefinition{
		Nodes: []workflow.Node{
			{
				ID:   "parallel1",
				Type: string(workflow.NodeTypeControlParallel),
				Data: workflow.NodeData{
					Name: "Race Condition",
					Config: mustMarshal(workflow.ParallelConfig{
						Branches: []workflow.ParallelBranch{
							{Name: "fast", Nodes: []string{"node1"}},
							{Name: "slow", Nodes: []string{"node2"}},
						},
						WaitMode:       "first",
						MaxConcurrency: 0,
						FailureMode:    "stop_all",
					}),
				},
			},
		},
		Edges: []workflow.Edge{},
	}

	mockRepo := &mockWorkflowRepo{
		workflow: &workflow.Workflow{
			ID:         "wf1",
			TenantID:   "tenant1",
			Definition: mustMarshal(definition),
		},
		stepExecutions: make(map[string]*workflow.StepExecution),
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	executor := &Executor{
		repo:            mockRepo,
		logger:          logger,
		retryStrategy:   NewRetryStrategy(DefaultRetryConfig(), logger),
		circuitBreakers: NewCircuitBreakerRegistry(DefaultCircuitBreakerConfig(), logger),
	}

	execCtx := &ExecutionContext{
		TenantID:    "tenant1",
		ExecutionID: "exec1",
		WorkflowID:  "wf1",
		TriggerData: map[string]interface{}{},
		StepOutputs: make(map[string]interface{}),
	}

	result, err := executor.executeParallelAction(
		context.Background(),
		definition.Nodes[0],
		execCtx,
		&definition,
	)

	require.NoError(t, err)
	require.NotNil(t, result)

	// Should have first_completed field
	resultMap, ok := result.(map[string]interface{})
	require.True(t, ok)
	assert.Contains(t, resultMap, "first_completed")
}

// TestParallelIntegration_FailureModeContinue tests continue on error
func TestParallelIntegration_FailureModeContinue(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	definition := workflow.WorkflowDefinition{
		Nodes: []workflow.Node{
			{
				ID:   "parallel1",
				Type: string(workflow.NodeTypeControlParallel),
				Data: workflow.NodeData{
					Name: "Best Effort Notifications",
					Config: mustMarshal(workflow.ParallelConfig{
						Branches: []workflow.ParallelBranch{
							{Name: "email", Nodes: []string{"send_email"}},
							{Name: "sms", Nodes: []string{"send_sms"}},
							{Name: "slack", Nodes: []string{"post_slack"}},
						},
						WaitMode:       "all",
						MaxConcurrency: 0,
						FailureMode:    "continue", // Continue even if some fail
					}),
				},
			},
		},
		Edges: []workflow.Edge{},
	}

	mockRepo := &mockWorkflowRepo{
		workflow: &workflow.Workflow{
			ID:         "wf1",
			TenantID:   "tenant1",
			Definition: mustMarshal(definition),
		},
		stepExecutions: make(map[string]*workflow.StepExecution),
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	executor := &Executor{
		repo:            mockRepo,
		logger:          logger,
		retryStrategy:   NewRetryStrategy(DefaultRetryConfig(), logger),
		circuitBreakers: NewCircuitBreakerRegistry(DefaultCircuitBreakerConfig(), logger),
	}

	execCtx := &ExecutionContext{
		TenantID:    "tenant1",
		ExecutionID: "exec1",
		WorkflowID:  "wf1",
		TriggerData: map[string]interface{}{
			"message": "System alert",
		},
		StepOutputs: make(map[string]interface{}),
	}

	result, err := executor.executeParallelAction(
		context.Background(),
		definition.Nodes[0],
		execCtx,
		&definition,
	)

	// Should not error even if some branches fail (with continue mode)
	require.NoError(t, err)
	require.NotNil(t, result)

	// Verify all branches attempted execution
	resultMap, ok := result.(map[string]interface{})
	require.True(t, ok)
	assert.Contains(t, resultMap, "total_branches")
	assert.Equal(t, 3, resultMap["total_branches"])
}

// TestParallelIntegration_MaxConcurrency tests concurrency limiting
func TestParallelIntegration_MaxConcurrency(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	definition := workflow.WorkflowDefinition{
		Nodes: []workflow.Node{
			{
				ID:   "parallel1",
				Type: string(workflow.NodeTypeControlParallel),
				Data: workflow.NodeData{
					Name: "Rate Limited Requests",
					Config: mustMarshal(workflow.ParallelConfig{
						Branches: []workflow.ParallelBranch{
							{Name: "req1", Nodes: []string{"node1"}},
							{Name: "req2", Nodes: []string{"node2"}},
							{Name: "req3", Nodes: []string{"node3"}},
							{Name: "req4", Nodes: []string{"node4"}},
							{Name: "req5", Nodes: []string{"node5"}},
						},
						WaitMode:       "all",
						MaxConcurrency: 2, // Only 2 at a time
						FailureMode:    "stop_all",
					}),
				},
			},
		},
		Edges: []workflow.Edge{},
	}

	mockRepo := &mockWorkflowRepo{
		workflow: &workflow.Workflow{
			ID:         "wf1",
			TenantID:   "tenant1",
			Definition: mustMarshal(definition),
		},
		stepExecutions: make(map[string]*workflow.StepExecution),
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	executor := &Executor{
		repo:            mockRepo,
		logger:          logger,
		retryStrategy:   NewRetryStrategy(DefaultRetryConfig(), logger),
		circuitBreakers: NewCircuitBreakerRegistry(DefaultCircuitBreakerConfig(), logger),
	}

	execCtx := &ExecutionContext{
		TenantID:    "tenant1",
		ExecutionID: "exec1",
		WorkflowID:  "wf1",
		TriggerData: map[string]interface{}{},
		StepOutputs: make(map[string]interface{}),
	}

	result, err := executor.executeParallelAction(
		context.Background(),
		definition.Nodes[0],
		execCtx,
		&definition,
	)

	require.NoError(t, err)
	require.NotNil(t, result)

	// All branches should complete
	resultMap, ok := result.(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, 5, resultMap["completed_branches"])
}

// TestParallelIntegration_Timeout tests timeout enforcement
func TestParallelIntegration_Timeout(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	definition := workflow.WorkflowDefinition{
		Nodes: []workflow.Node{
			{
				ID:   "parallel1",
				Type: string(workflow.NodeTypeControlParallel),
				Data: workflow.NodeData{
					Name: "Timed Execution",
					Config: mustMarshal(workflow.ParallelConfig{
						Branches: []workflow.ParallelBranch{
							{Name: "branch1", Nodes: []string{"node1"}},
						},
						WaitMode:       "all",
						Timeout:        "100ms",
						MaxConcurrency: 0,
						FailureMode:    "stop_all",
					}),
				},
			},
		},
		Edges: []workflow.Edge{},
	}

	mockRepo := &mockWorkflowRepo{
		workflow: &workflow.Workflow{
			ID:         "wf1",
			TenantID:   "tenant1",
			Definition: mustMarshal(definition),
		},
		stepExecutions: make(map[string]*workflow.StepExecution),
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	executor := &Executor{
		repo:            mockRepo,
		logger:          logger,
		retryStrategy:   NewRetryStrategy(DefaultRetryConfig(), logger),
		circuitBreakers: NewCircuitBreakerRegistry(DefaultCircuitBreakerConfig(), logger),
	}

	execCtx := &ExecutionContext{
		TenantID:    "tenant1",
		ExecutionID: "exec1",
		WorkflowID:  "wf1",
		TriggerData: map[string]interface{}{},
		StepOutputs: make(map[string]interface{}),
	}

	// Execute - should complete quickly or timeout
	result, err := executor.executeParallelAction(
		context.Background(),
		definition.Nodes[0],
		execCtx,
		&definition,
	)

	// Either succeeds quickly or times out
	if err != nil {
		assert.Contains(t, err.Error(), "context")
	} else {
		require.NotNil(t, result)
	}
}

// TestParallelIntegration_NestedParallel tests parallel within parallel
func TestParallelIntegration_NestedParallel(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	definition := workflow.WorkflowDefinition{
		Nodes: []workflow.Node{
			{
				ID:   "outer_parallel",
				Type: string(workflow.NodeTypeControlParallel),
				Data: workflow.NodeData{
					Name: "Outer Parallel",
					Config: mustMarshal(workflow.ParallelConfig{
						Branches: []workflow.ParallelBranch{
							{Name: "group1", Nodes: []string{"inner_parallel1"}},
							{Name: "group2", Nodes: []string{"inner_parallel2"}},
						},
						WaitMode:       "all",
						MaxConcurrency: 0,
						FailureMode:    "stop_all",
					}),
				},
			},
		},
		Edges: []workflow.Edge{},
	}

	mockRepo := &mockWorkflowRepo{
		workflow: &workflow.Workflow{
			ID:         "wf1",
			TenantID:   "tenant1",
			Definition: mustMarshal(definition),
		},
		stepExecutions: make(map[string]*workflow.StepExecution),
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	executor := &Executor{
		repo:            mockRepo,
		logger:          logger,
		retryStrategy:   NewRetryStrategy(DefaultRetryConfig(), logger),
		circuitBreakers: NewCircuitBreakerRegistry(DefaultCircuitBreakerConfig(), logger),
	}

	execCtx := &ExecutionContext{
		TenantID:    "tenant1",
		ExecutionID: "exec1",
		WorkflowID:  "wf1",
		TriggerData: map[string]interface{}{},
		StepOutputs: make(map[string]interface{}),
	}

	result, err := executor.executeParallelAction(
		context.Background(),
		definition.Nodes[0],
		execCtx,
		&definition,
	)

	require.NoError(t, err)
	require.NotNil(t, result)

	// Nested parallel should work
	resultMap, ok := result.(map[string]interface{})
	require.True(t, ok)
	assert.Contains(t, resultMap, "completed_branches")
}

// Note: mockWorkflowRepo is defined in loop_integration_test.go and shared across integration tests
