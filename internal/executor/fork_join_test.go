package executor

import (
	"context"
	"testing"
	"time"

	"github.com/gorax/gorax/internal/workflow"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestExecuteForkAction tests fork node execution
func TestExecuteForkAction_BasicFork(t *testing.T) {
	// Test basic fork that splits into multiple branches
	config := workflow.ForkConfig{
		BranchCount: 3,
	}

	execCtx := &ExecutionContext{
		TenantID:    "tenant1",
		ExecutionID: "exec1",
		WorkflowID:  "workflow1",
		TriggerData: map[string]interface{}{
			"value": 100,
		},
		StepOutputs: make(map[string]interface{}),
	}

	forkExec := &forkExecutor{}
	result, err := forkExec.executeFork(context.Background(), config, execCtx)

	require.NoError(t, err)
	assert.NotNil(t, result)

	forkResult, ok := result.(*ForkResult)
	require.True(t, ok)
	assert.Equal(t, 3, forkResult.BranchCount)
	assert.Len(t, forkResult.BranchIDs, 3)
}

func TestExecuteForkAction_InvalidBranchCount(t *testing.T) {
	// Test that fork rejects invalid branch count
	config := workflow.ForkConfig{
		BranchCount: 0, // Invalid
	}

	execCtx := &ExecutionContext{
		TenantID:    "tenant1",
		ExecutionID: "exec1",
		WorkflowID:  "workflow1",
		TriggerData: map[string]interface{}{},
		StepOutputs: make(map[string]interface{}),
	}

	forkExec := &forkExecutor{}
	_, err := forkExec.executeFork(context.Background(), config, execCtx)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "branch_count")
}

func TestExecuteForkAction_NegativeBranchCount(t *testing.T) {
	// Test that fork rejects negative branch count
	config := workflow.ForkConfig{
		BranchCount: -1,
	}

	execCtx := &ExecutionContext{
		TenantID:    "tenant1",
		ExecutionID: "exec1",
		WorkflowID:  "workflow1",
		TriggerData: map[string]interface{}{},
		StepOutputs: make(map[string]interface{}),
	}

	forkExec := &forkExecutor{}
	_, err := forkExec.executeFork(context.Background(), config, execCtx)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "branch_count")
}

func TestExecuteForkAction_SingleBranch(t *testing.T) {
	// Test fork with single branch (edge case)
	config := workflow.ForkConfig{
		BranchCount: 1,
	}

	execCtx := &ExecutionContext{
		TenantID:    "tenant1",
		ExecutionID: "exec1",
		WorkflowID:  "workflow1",
		TriggerData: map[string]interface{}{},
		StepOutputs: make(map[string]interface{}),
	}

	forkExec := &forkExecutor{}
	result, err := forkExec.executeFork(context.Background(), config, execCtx)

	require.NoError(t, err)
	forkResult, ok := result.(*ForkResult)
	require.True(t, ok)
	assert.Equal(t, 1, forkResult.BranchCount)
	assert.Len(t, forkResult.BranchIDs, 1)
}

// TestExecuteJoinAction tests join node execution
func TestExecuteJoinAction_WaitAll(t *testing.T) {
	// Test join that waits for all branches
	config := workflow.JoinConfig{
		JoinStrategy: "wait_all",
	}

	// Simulate 3 completed branches
	branchResults := map[string]interface{}{
		"branch_0": map[string]interface{}{"status": "completed", "value": 10},
		"branch_1": map[string]interface{}{"status": "completed", "value": 20},
		"branch_2": map[string]interface{}{"status": "completed", "value": 30},
	}

	execCtx := &ExecutionContext{
		TenantID:    "tenant1",
		ExecutionID: "exec1",
		WorkflowID:  "workflow1",
		TriggerData: map[string]interface{}{},
		StepOutputs: branchResults,
	}

	joinExec := &joinExecutor{}
	result, err := joinExec.executeJoin(context.Background(), config, execCtx, []string{"branch_0", "branch_1", "branch_2"})

	require.NoError(t, err)
	assert.NotNil(t, result)

	joinResult, ok := result.(*JoinResult)
	require.True(t, ok)
	assert.Equal(t, 3, joinResult.CompletedBranches)
	assert.Len(t, joinResult.BranchOutputs, 3)
}

func TestExecuteJoinAction_WaitN(t *testing.T) {
	// Test join that waits for N branches out of M
	config := workflow.JoinConfig{
		JoinStrategy:  "wait_n",
		RequiredCount: 2, // Wait for 2 out of 3
	}

	// Simulate 2 completed branches (3rd still running)
	branchResults := map[string]interface{}{
		"branch_0": map[string]interface{}{"status": "completed", "value": 10},
		"branch_1": map[string]interface{}{"status": "completed", "value": 20},
	}

	execCtx := &ExecutionContext{
		TenantID:    "tenant1",
		ExecutionID: "exec1",
		WorkflowID:  "workflow1",
		TriggerData: map[string]interface{}{},
		StepOutputs: branchResults,
	}

	joinExec := &joinExecutor{}
	result, err := joinExec.executeJoin(context.Background(), config, execCtx, []string{"branch_0", "branch_1", "branch_2"})

	require.NoError(t, err)
	joinResult, ok := result.(*JoinResult)
	require.True(t, ok)
	assert.GreaterOrEqual(t, joinResult.CompletedBranches, 2)
}

func TestExecuteJoinAction_Timeout(t *testing.T) {
	// Test join with timeout
	config := workflow.JoinConfig{
		JoinStrategy: "wait_all",
		TimeoutMs:    100, // 100ms timeout
		OnTimeout:    "fail",
	}

	// Only 1 branch completed out of 3
	branchResults := map[string]interface{}{
		"branch_0": map[string]interface{}{"status": "completed"},
	}

	execCtx := &ExecutionContext{
		TenantID:    "tenant1",
		ExecutionID: "exec1",
		WorkflowID:  "workflow1",
		TriggerData: map[string]interface{}{},
		StepOutputs: branchResults,
	}

	joinExec := &joinExecutor{}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	// Wait for timeout
	time.Sleep(60 * time.Millisecond)

	_, err := joinExec.executeJoin(ctx, config, execCtx, []string{"branch_0", "branch_1", "branch_2"})

	// Should error due to timeout
	require.Error(t, err)
	assert.Contains(t, err.Error(), "timeout")
}

func TestExecuteJoinAction_TimeoutContinue(t *testing.T) {
	// Test join with timeout and continue strategy
	config := workflow.JoinConfig{
		JoinStrategy: "wait_all",
		TimeoutMs:    50, // Very short timeout
		OnTimeout:    "continue", // Continue with partial results
	}

	// Only 1 branch completed (simulating others still running)
	branchResults := map[string]interface{}{
		"branch_0": map[string]interface{}{"status": "completed"},
	}

	execCtx := &ExecutionContext{
		TenantID:    "tenant1",
		ExecutionID: "exec1",
		WorkflowID:  "workflow1",
		TriggerData: map[string]interface{}{},
		StepOutputs: branchResults,
	}

	// Create context with timeout to force timeout scenario
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	// Wait to ensure timeout
	time.Sleep(20 * time.Millisecond)

	joinExec := &joinExecutor{}
	result, err := joinExec.executeJoin(ctx, config, execCtx, []string{"branch_0", "branch_1", "branch_2"})

	// Should succeed with partial results
	require.NoError(t, err)
	joinResult, ok := result.(*JoinResult)
	require.True(t, ok)
	assert.Equal(t, 1, joinResult.CompletedBranches)
	assert.True(t, joinResult.TimedOut)
}

func TestExecuteJoinAction_InvalidStrategy(t *testing.T) {
	// Test that join rejects invalid strategy
	config := workflow.JoinConfig{
		JoinStrategy: "invalid_strategy",
	}

	execCtx := &ExecutionContext{
		TenantID:    "tenant1",
		ExecutionID: "exec1",
		WorkflowID:  "workflow1",
		TriggerData: map[string]interface{}{},
		StepOutputs: make(map[string]interface{}),
	}

	joinExec := &joinExecutor{}
	_, err := joinExec.executeJoin(context.Background(), config, execCtx, []string{"branch_0"})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "join_strategy")
}

func TestExecuteJoinAction_WaitNInvalidCount(t *testing.T) {
	// Test wait_n with invalid required count
	config := workflow.JoinConfig{
		JoinStrategy:  "wait_n",
		RequiredCount: 0, // Invalid
	}

	execCtx := &ExecutionContext{
		TenantID:    "tenant1",
		ExecutionID: "exec1",
		WorkflowID:  "workflow1",
		TriggerData: map[string]interface{}{},
		StepOutputs: make(map[string]interface{}),
	}

	joinExec := &joinExecutor{}
	_, err := joinExec.executeJoin(context.Background(), config, execCtx, []string{"branch_0", "branch_1"})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "required_count")
}

func TestExecuteJoinAction_WaitNCountExceedsBranches(t *testing.T) {
	// Test wait_n with required count > total branches
	config := workflow.JoinConfig{
		JoinStrategy:  "wait_n",
		RequiredCount: 5, // More than available branches
	}

	branchResults := map[string]interface{}{
		"branch_0": map[string]interface{}{"status": "completed"},
		"branch_1": map[string]interface{}{"status": "completed"},
	}

	execCtx := &ExecutionContext{
		TenantID:    "tenant1",
		ExecutionID: "exec1",
		WorkflowID:  "workflow1",
		TriggerData: map[string]interface{}{},
		StepOutputs: branchResults,
	}

	joinExec := &joinExecutor{}
	_, err := joinExec.executeJoin(context.Background(), config, execCtx, []string{"branch_0", "branch_1"})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "required_count")
}

func TestExecuteJoinAction_EmptyBranches(t *testing.T) {
	// Test join with no branches
	config := workflow.JoinConfig{
		JoinStrategy: "wait_all",
	}

	execCtx := &ExecutionContext{
		TenantID:    "tenant1",
		ExecutionID: "exec1",
		WorkflowID:  "workflow1",
		TriggerData: map[string]interface{}{},
		StepOutputs: make(map[string]interface{}),
	}

	joinExec := &joinExecutor{}
	result, err := joinExec.executeJoin(context.Background(), config, execCtx, []string{})

	require.NoError(t, err)
	joinResult, ok := result.(*JoinResult)
	require.True(t, ok)
	assert.Equal(t, 0, joinResult.CompletedBranches)
}

func TestExecuteJoinAction_DefaultStrategy(t *testing.T) {
	// Test that default strategy is applied when not specified
	config := workflow.JoinConfig{
		JoinStrategy: "", // Not specified
	}

	branchResults := map[string]interface{}{
		"branch_0": map[string]interface{}{"status": "completed"},
	}

	execCtx := &ExecutionContext{
		TenantID:    "tenant1",
		ExecutionID: "exec1",
		WorkflowID:  "workflow1",
		TriggerData: map[string]interface{}{},
		StepOutputs: branchResults,
	}

	joinExec := &joinExecutor{}
	result, err := joinExec.executeJoin(context.Background(), config, execCtx, []string{"branch_0"})

	require.NoError(t, err)
	joinResult, ok := result.(*JoinResult)
	require.True(t, ok)
	assert.Equal(t, "wait_all", joinResult.Metadata["strategy"])
}

func TestExecuteJoinAction_BranchErrorHandling(t *testing.T) {
	// Test join behavior when branches have errors
	config := workflow.JoinConfig{
		JoinStrategy: "wait_all",
	}

	branchResults := map[string]interface{}{
		"branch_0": map[string]interface{}{"status": "completed", "value": 10},
		"branch_1": map[string]interface{}{"status": "failed", "error": "branch failed"},
		"branch_2": map[string]interface{}{"status": "completed", "value": 30},
	}

	execCtx := &ExecutionContext{
		TenantID:    "tenant1",
		ExecutionID: "exec1",
		WorkflowID:  "workflow1",
		TriggerData: map[string]interface{}{},
		StepOutputs: branchResults,
	}

	joinExec := &joinExecutor{}
	result, err := joinExec.executeJoin(context.Background(), config, execCtx, []string{"branch_0", "branch_1", "branch_2"})

	require.NoError(t, err)
	joinResult, ok := result.(*JoinResult)
	require.True(t, ok)
	// Should include all branches, even failed ones
	assert.Len(t, joinResult.BranchOutputs, 3)
}
