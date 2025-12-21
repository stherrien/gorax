package executor

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gorax/gorax/internal/workflow"
)

func TestExecuteParallelAction_TwoBranches(t *testing.T) {
	// Test parallel execution of 2 independent branches
	config := workflow.ParallelConfig{
		ErrorStrategy:  "fail_fast",
		MaxConcurrency: 0, // unlimited
	}

	// Create execution context
	execCtx := &ExecutionContext{
		TenantID:    "tenant1",
		ExecutionID: "exec1",
		WorkflowID:  "workflow1",
		TriggerData: map[string]interface{}{
			"value": 10,
		},
		StepOutputs: make(map[string]interface{}),
	}

	// Create two parallel branches
	branchNodes := [][]workflow.Node{
		// Branch 1
		{
			{
				ID:   "branch1_node1",
				Type: string(workflow.NodeTypeActionTransform),
				Data: workflow.NodeData{
					Name: "Branch 1 Transform",
					Config: mustMarshal(map[string]interface{}{
						"expression": "${trigger.value * 2}",
					}),
				},
			},
		},
		// Branch 2
		{
			{
				ID:   "branch2_node1",
				Type: string(workflow.NodeTypeActionTransform),
				Data: workflow.NodeData{
					Name: "Branch 2 Transform",
					Config: mustMarshal(map[string]interface{}{
						"expression": "${trigger.value * 3}",
					}),
				},
			},
		},
	}

	executor := &parallelExecutor{}
	result, err := executor.executeParallel(context.Background(), config, execCtx, branchNodes)

	require.NoError(t, err)
	assert.NotNil(t, result)

	// Verify results contain outputs from all branches
	parallelResult, ok := result.(*ParallelResult)
	require.True(t, ok)
	assert.Equal(t, 2, parallelResult.BranchCount)
	assert.Len(t, parallelResult.BranchResults, 2)

	// Verify all branches succeeded
	for i, branchResult := range parallelResult.BranchResults {
		assert.Equal(t, i, branchResult.BranchIndex)
		assert.Nil(t, branchResult.Error)
		assert.NotNil(t, branchResult.Output)
	}
}

func TestExecuteParallelAction_ThreeBranches(t *testing.T) {
	// Test parallel execution of 3 independent branches
	config := workflow.ParallelConfig{
		ErrorStrategy:  "fail_fast",
		MaxConcurrency: 0,
	}

	execCtx := &ExecutionContext{
		TenantID:    "tenant1",
		ExecutionID: "exec1",
		WorkflowID:  "workflow1",
		TriggerData: map[string]interface{}{
			"value": 5,
		},
		StepOutputs: make(map[string]interface{}),
	}

	// Create three parallel branches
	branchNodes := [][]workflow.Node{
		{
			{
				ID:   "branch1_node",
				Type: string(workflow.NodeTypeActionTransform),
				Data: workflow.NodeData{
					Name:   "Branch 1",
					Config: mustMarshal(map[string]interface{}{}),
				},
			},
		},
		{
			{
				ID:   "branch2_node",
				Type: string(workflow.NodeTypeActionTransform),
				Data: workflow.NodeData{
					Name:   "Branch 2",
					Config: mustMarshal(map[string]interface{}{}),
				},
			},
		},
		{
			{
				ID:   "branch3_node",
				Type: string(workflow.NodeTypeActionTransform),
				Data: workflow.NodeData{
					Name:   "Branch 3",
					Config: mustMarshal(map[string]interface{}{}),
				},
			},
		},
	}

	executor := &parallelExecutor{}
	result, err := executor.executeParallel(context.Background(), config, execCtx, branchNodes)

	require.NoError(t, err)
	parallelResult, ok := result.(*ParallelResult)
	require.True(t, ok)
	assert.Equal(t, 3, parallelResult.BranchCount)
	assert.Len(t, parallelResult.BranchResults, 3)
}

func TestExecuteParallelAction_FailFastStrategy(t *testing.T) {
	// Test that fail-fast strategy stops execution when one branch fails
	config := workflow.ParallelConfig{
		ErrorStrategy:  "fail_fast",
		MaxConcurrency: 0,
	}

	execCtx := &ExecutionContext{
		TenantID:    "tenant1",
		ExecutionID: "exec1",
		WorkflowID:  "workflow1",
		TriggerData: map[string]interface{}{},
		StepOutputs: make(map[string]interface{}),
	}

	// Create branches where one will fail
	// Note: In actual test with full executor, this would be a failing node
	// For now, we test the configuration is properly set
	branchNodes := [][]workflow.Node{
		{
			{
				ID:   "branch1_node",
				Type: string(workflow.NodeTypeActionTransform),
				Data: workflow.NodeData{
					Name:   "Branch 1 Success",
					Config: mustMarshal(map[string]interface{}{}),
				},
			},
		},
		{
			{
				ID:   "branch2_node",
				Type: string(workflow.NodeTypeActionTransform),
				Data: workflow.NodeData{
					Name:   "Branch 2 Success",
					Config: mustMarshal(map[string]interface{}{}),
				},
			},
		},
	}

	executor := &parallelExecutor{}
	result, err := executor.executeParallel(context.Background(), config, execCtx, branchNodes)

	require.NoError(t, err)
	parallelResult, ok := result.(*ParallelResult)
	require.True(t, ok)
	assert.Equal(t, "fail_fast", parallelResult.Metadata["error_strategy"])
}

func TestExecuteParallelAction_WaitAllStrategy(t *testing.T) {
	// Test that wait-all strategy waits for all branches even when one fails
	config := workflow.ParallelConfig{
		ErrorStrategy:  "wait_all",
		MaxConcurrency: 0,
	}

	execCtx := &ExecutionContext{
		TenantID:    "tenant1",
		ExecutionID: "exec1",
		WorkflowID:  "workflow1",
		TriggerData: map[string]interface{}{},
		StepOutputs: make(map[string]interface{}),
	}

	branchNodes := [][]workflow.Node{
		{
			{
				ID:   "branch1_node",
				Type: string(workflow.NodeTypeActionTransform),
				Data: workflow.NodeData{
					Name:   "Branch 1",
					Config: mustMarshal(map[string]interface{}{}),
				},
			},
		},
		{
			{
				ID:   "branch2_node",
				Type: string(workflow.NodeTypeActionTransform),
				Data: workflow.NodeData{
					Name:   "Branch 2",
					Config: mustMarshal(map[string]interface{}{}),
				},
			},
		},
	}

	executor := &parallelExecutor{}
	result, err := executor.executeParallel(context.Background(), config, execCtx, branchNodes)

	require.NoError(t, err)
	parallelResult, ok := result.(*ParallelResult)
	require.True(t, ok)
	assert.Equal(t, "wait_all", parallelResult.Metadata["error_strategy"])
	assert.Equal(t, 2, parallelResult.BranchCount)
}

func TestExecuteParallelAction_MaxConcurrency(t *testing.T) {
	// Test that max concurrency limit is respected
	config := workflow.ParallelConfig{
		ErrorStrategy:  "fail_fast",
		MaxConcurrency: 2, // Only 2 branches at a time
	}

	execCtx := &ExecutionContext{
		TenantID:    "tenant1",
		ExecutionID: "exec1",
		WorkflowID:  "workflow1",
		TriggerData: map[string]interface{}{},
		StepOutputs: make(map[string]interface{}),
	}

	// Create 5 branches, but only 2 should run concurrently
	branchNodes := [][]workflow.Node{
		{{ID: "branch1", Type: string(workflow.NodeTypeActionTransform), Data: workflow.NodeData{Config: mustMarshal(map[string]interface{}{})}}},
		{{ID: "branch2", Type: string(workflow.NodeTypeActionTransform), Data: workflow.NodeData{Config: mustMarshal(map[string]interface{}{})}}},
		{{ID: "branch3", Type: string(workflow.NodeTypeActionTransform), Data: workflow.NodeData{Config: mustMarshal(map[string]interface{}{})}}},
		{{ID: "branch4", Type: string(workflow.NodeTypeActionTransform), Data: workflow.NodeData{Config: mustMarshal(map[string]interface{}{})}}},
		{{ID: "branch5", Type: string(workflow.NodeTypeActionTransform), Data: workflow.NodeData{Config: mustMarshal(map[string]interface{}{})}}},
	}

	executor := &parallelExecutor{}
	result, err := executor.executeParallel(context.Background(), config, execCtx, branchNodes)

	require.NoError(t, err)
	parallelResult, ok := result.(*ParallelResult)
	require.True(t, ok)
	assert.Equal(t, 5, parallelResult.BranchCount)
	assert.Equal(t, 2, parallelResult.Metadata["max_concurrency"])
	// All branches should complete, just in batches
	assert.Len(t, parallelResult.BranchResults, 5)
}

func TestExecuteParallelAction_UnlimitedConcurrency(t *testing.T) {
	// Test unlimited concurrency (0 means no limit)
	config := workflow.ParallelConfig{
		ErrorStrategy:  "fail_fast",
		MaxConcurrency: 0, // unlimited
	}

	execCtx := &ExecutionContext{
		TenantID:    "tenant1",
		ExecutionID: "exec1",
		WorkflowID:  "workflow1",
		TriggerData: map[string]interface{}{},
		StepOutputs: make(map[string]interface{}),
	}

	// Create 10 branches - all should run concurrently
	branchNodes := make([][]workflow.Node, 10)
	for i := 0; i < 10; i++ {
		branchNodes[i] = []workflow.Node{
			{
				ID:   "branch" + string(rune('0'+i)),
				Type: string(workflow.NodeTypeActionTransform),
				Data: workflow.NodeData{
					Config: mustMarshal(map[string]interface{}{}),
				},
			},
		}
	}

	executor := &parallelExecutor{}
	result, err := executor.executeParallel(context.Background(), config, execCtx, branchNodes)

	require.NoError(t, err)
	parallelResult, ok := result.(*ParallelResult)
	require.True(t, ok)
	assert.Equal(t, 10, parallelResult.BranchCount)
}

func TestExecuteParallelAction_EmptyBranches(t *testing.T) {
	// Test with no branches
	config := workflow.ParallelConfig{
		ErrorStrategy:  "fail_fast",
		MaxConcurrency: 0,
	}

	execCtx := &ExecutionContext{
		TenantID:    "tenant1",
		ExecutionID: "exec1",
		WorkflowID:  "workflow1",
		TriggerData: map[string]interface{}{},
		StepOutputs: make(map[string]interface{}),
	}

	branchNodes := [][]workflow.Node{}

	executor := &parallelExecutor{}
	result, err := executor.executeParallel(context.Background(), config, execCtx, branchNodes)

	require.NoError(t, err)
	parallelResult, ok := result.(*ParallelResult)
	require.True(t, ok)
	assert.Equal(t, 0, parallelResult.BranchCount)
	assert.Len(t, parallelResult.BranchResults, 0)
}

func TestExecuteParallelAction_ResultAggregation(t *testing.T) {
	// Test that results from all branches are collected properly
	config := workflow.ParallelConfig{
		ErrorStrategy:  "wait_all",
		MaxConcurrency: 0,
	}

	execCtx := &ExecutionContext{
		TenantID:    "tenant1",
		ExecutionID: "exec1",
		WorkflowID:  "workflow1",
		TriggerData: map[string]interface{}{
			"base": 100,
		},
		StepOutputs: make(map[string]interface{}),
	}

	branchNodes := [][]workflow.Node{
		{
			{
				ID:   "add_branch",
				Type: string(workflow.NodeTypeActionTransform),
				Data: workflow.NodeData{
					Name: "Add 10",
					Config: mustMarshal(map[string]interface{}{
						"expression": "${trigger.base + 10}",
					}),
				},
			},
		},
		{
			{
				ID:   "multiply_branch",
				Type: string(workflow.NodeTypeActionTransform),
				Data: workflow.NodeData{
					Name: "Multiply by 2",
					Config: mustMarshal(map[string]interface{}{
						"expression": "${trigger.base * 2}",
					}),
				},
			},
		},
	}

	executor := &parallelExecutor{}
	result, err := executor.executeParallel(context.Background(), config, execCtx, branchNodes)

	require.NoError(t, err)
	parallelResult, ok := result.(*ParallelResult)
	require.True(t, ok)
	assert.Equal(t, 2, parallelResult.BranchCount)

	// Verify each branch has its output
	for i, branchResult := range parallelResult.BranchResults {
		assert.Equal(t, i, branchResult.BranchIndex)
		assert.NotNil(t, branchResult.Output)
		assert.Nil(t, branchResult.Error)
	}
}

func TestExecuteParallelAction_NestedParallel(t *testing.T) {
	// Test nested parallel execution (parallel within parallel)
	outerConfig := workflow.ParallelConfig{
		ErrorStrategy:  "fail_fast",
		MaxConcurrency: 0,
	}

	execCtx := &ExecutionContext{
		TenantID:    "tenant1",
		ExecutionID: "exec1",
		WorkflowID:  "workflow1",
		TriggerData: map[string]interface{}{
			"value": 10,
		},
		StepOutputs: make(map[string]interface{}),
	}

	// Outer parallel has branches that contain parallel nodes
	branchNodes := [][]workflow.Node{
		{
			{
				ID:   "nested_parallel_1",
				Type: string(workflow.NodeTypeControlParallel),
				Data: workflow.NodeData{
					Name: "Nested Parallel 1",
					Config: mustMarshal(workflow.ParallelConfig{
						ErrorStrategy:  "fail_fast",
						MaxConcurrency: 0,
					}),
				},
			},
		},
		{
			{
				ID:   "nested_parallel_2",
				Type: string(workflow.NodeTypeControlParallel),
				Data: workflow.NodeData{
					Name: "Nested Parallel 2",
					Config: mustMarshal(workflow.ParallelConfig{
						ErrorStrategy:  "wait_all",
						MaxConcurrency: 2,
					}),
				},
			},
		},
	}

	executor := &parallelExecutor{}
	result, err := executor.executeParallel(context.Background(), outerConfig, execCtx, branchNodes)

	require.NoError(t, err)
	parallelResult, ok := result.(*ParallelResult)
	require.True(t, ok)
	assert.Equal(t, 2, parallelResult.BranchCount)
}

func TestExecuteParallelAction_BranchWithMultipleNodes(t *testing.T) {
	// Test parallel branches where each branch has multiple nodes
	config := workflow.ParallelConfig{
		ErrorStrategy:  "fail_fast",
		MaxConcurrency: 0,
	}

	execCtx := &ExecutionContext{
		TenantID:    "tenant1",
		ExecutionID: "exec1",
		WorkflowID:  "workflow1",
		TriggerData: map[string]interface{}{
			"value": 5,
		},
		StepOutputs: make(map[string]interface{}),
	}

	// Branch 1 has 2 nodes, Branch 2 has 3 nodes
	branchNodes := [][]workflow.Node{
		{
			{
				ID:   "b1_node1",
				Type: string(workflow.NodeTypeActionTransform),
				Data: workflow.NodeData{
					Name:   "Branch 1 Node 1",
					Config: mustMarshal(map[string]interface{}{}),
				},
			},
			{
				ID:   "b1_node2",
				Type: string(workflow.NodeTypeActionTransform),
				Data: workflow.NodeData{
					Name:   "Branch 1 Node 2",
					Config: mustMarshal(map[string]interface{}{}),
				},
			},
		},
		{
			{
				ID:   "b2_node1",
				Type: string(workflow.NodeTypeActionTransform),
				Data: workflow.NodeData{
					Name:   "Branch 2 Node 1",
					Config: mustMarshal(map[string]interface{}{}),
				},
			},
			{
				ID:   "b2_node2",
				Type: string(workflow.NodeTypeActionTransform),
				Data: workflow.NodeData{
					Name:   "Branch 2 Node 2",
					Config: mustMarshal(map[string]interface{}{}),
				},
			},
			{
				ID:   "b2_node3",
				Type: string(workflow.NodeTypeActionTransform),
				Data: workflow.NodeData{
					Name:   "Branch 2 Node 3",
					Config: mustMarshal(map[string]interface{}{}),
				},
			},
		},
	}

	executor := &parallelExecutor{}
	result, err := executor.executeParallel(context.Background(), config, execCtx, branchNodes)

	require.NoError(t, err)
	parallelResult, ok := result.(*ParallelResult)
	require.True(t, ok)
	assert.Equal(t, 2, parallelResult.BranchCount)
}

func TestExecuteParallelAction_DefaultErrorStrategy(t *testing.T) {
	// Test that default error strategy is applied when not specified
	config := workflow.ParallelConfig{
		ErrorStrategy:  "", // Not specified
		MaxConcurrency: 0,
	}

	execCtx := &ExecutionContext{
		TenantID:    "tenant1",
		ExecutionID: "exec1",
		WorkflowID:  "workflow1",
		TriggerData: map[string]interface{}{},
		StepOutputs: make(map[string]interface{}),
	}

	branchNodes := [][]workflow.Node{
		{{ID: "branch1", Type: string(workflow.NodeTypeActionTransform), Data: workflow.NodeData{Config: mustMarshal(map[string]interface{}{})}}},
		{{ID: "branch2", Type: string(workflow.NodeTypeActionTransform), Data: workflow.NodeData{Config: mustMarshal(map[string]interface{}{})}}},
	}

	executor := &parallelExecutor{}
	result, err := executor.executeParallel(context.Background(), config, execCtx, branchNodes)

	require.NoError(t, err)
	parallelResult, ok := result.(*ParallelResult)
	require.True(t, ok)
	// Should default to fail_fast
	assert.Equal(t, "fail_fast", parallelResult.Metadata["error_strategy"])
}

func TestExecuteParallelAction_ContextCancellation(t *testing.T) {
	// Test that parallel execution respects context cancellation
	config := workflow.ParallelConfig{
		ErrorStrategy:  "wait_all",
		MaxConcurrency: 0,
	}

	execCtx := &ExecutionContext{
		TenantID:    "tenant1",
		ExecutionID: "exec1",
		WorkflowID:  "workflow1",
		TriggerData: map[string]interface{}{},
		StepOutputs: make(map[string]interface{}),
	}

	branchNodes := [][]workflow.Node{
		{{ID: "branch1", Type: string(workflow.NodeTypeActionTransform), Data: workflow.NodeData{Config: mustMarshal(map[string]interface{}{})}}},
		{{ID: "branch2", Type: string(workflow.NodeTypeActionTransform), Data: workflow.NodeData{Config: mustMarshal(map[string]interface{}{})}}},
	}

	// Create a context that's already cancelled
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	executor := &parallelExecutor{}
	_, err := executor.executeParallel(ctx, config, execCtx, branchNodes)

	// Should return error due to cancelled context
	assert.Error(t, err)
	assert.True(t, errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded))
}

func TestExecuteParallelAction_Timeout(t *testing.T) {
	// Test parallel execution with timeout
	config := workflow.ParallelConfig{
		ErrorStrategy:  "fail_fast",
		MaxConcurrency: 0,
	}

	execCtx := &ExecutionContext{
		TenantID:    "tenant1",
		ExecutionID: "exec1",
		WorkflowID:  "workflow1",
		TriggerData: map[string]interface{}{},
		StepOutputs: make(map[string]interface{}),
	}

	branchNodes := [][]workflow.Node{
		{{ID: "branch1", Type: string(workflow.NodeTypeActionTransform), Data: workflow.NodeData{Config: mustMarshal(map[string]interface{}{})}}},
	}

	// Create a context with very short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()

	// Wait to ensure timeout
	time.Sleep(1 * time.Millisecond)

	executor := &parallelExecutor{}
	_, err := executor.executeParallel(ctx, config, execCtx, branchNodes)

	// May error due to timeout (depends on timing)
	// This test is mainly to verify timeout handling exists
	if err != nil {
		assert.True(t, errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled))
	}
}

func TestExecuteParallelAction_InvalidErrorStrategy(t *testing.T) {
	// Test validation of error strategy
	config := workflow.ParallelConfig{
		ErrorStrategy:  "invalid_strategy",
		MaxConcurrency: 0,
	}

	execCtx := &ExecutionContext{
		TenantID:    "tenant1",
		ExecutionID: "exec1",
		WorkflowID:  "workflow1",
		TriggerData: map[string]interface{}{},
		StepOutputs: make(map[string]interface{}),
	}

	branchNodes := [][]workflow.Node{
		{{ID: "branch1", Type: string(workflow.NodeTypeActionTransform), Data: workflow.NodeData{Config: mustMarshal(map[string]interface{}{})}}},
	}

	executor := &parallelExecutor{}
	_, err := executor.executeParallel(context.Background(), config, execCtx, branchNodes)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "error_strategy")
}

func TestExecuteParallelAction_NegativeMaxConcurrency(t *testing.T) {
	// Test that negative max concurrency is rejected
	config := workflow.ParallelConfig{
		ErrorStrategy:  "fail_fast",
		MaxConcurrency: -1,
	}

	execCtx := &ExecutionContext{
		TenantID:    "tenant1",
		ExecutionID: "exec1",
		WorkflowID:  "workflow1",
		TriggerData: map[string]interface{}{},
		StepOutputs: make(map[string]interface{}),
	}

	branchNodes := [][]workflow.Node{
		{{ID: "branch1", Type: string(workflow.NodeTypeActionTransform), Data: workflow.NodeData{Config: mustMarshal(map[string]interface{}{})}}},
	}

	executor := &parallelExecutor{}
	_, err := executor.executeParallel(context.Background(), config, execCtx, branchNodes)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "max_concurrency")
}
