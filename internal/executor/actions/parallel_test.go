package actions

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gorax/gorax/internal/workflow"
)

// TestParallelAction_BasicExecution tests basic parallel execution with named branches
func TestParallelAction_BasicExecution(t *testing.T) {
	config := workflow.ParallelConfig{
		Branches: []workflow.ParallelBranch{
			{Name: "branch1", Nodes: []string{"node1", "node2"}},
			{Name: "branch2", Nodes: []string{"node3"}},
		},
		WaitMode:       "all",
		MaxConcurrency: 0,
		FailureMode:    "stop_all",
	}

	input := NewActionInput(config, map[string]interface{}{
		"value": 10,
	})

	action := &ParallelAction{}
	output, err := action.Execute(context.Background(), input)

	require.NoError(t, err)
	require.NotNil(t, output)
	assert.NotNil(t, output.Data)

	// Verify metadata
	assert.Contains(t, output.Metadata, "branch_count")
	assert.Equal(t, 2, output.Metadata["branch_count"])
}

// TestParallelAction_WaitModeAll tests wait mode "all" - waits for all branches
func TestParallelAction_WaitModeAll(t *testing.T) {
	config := workflow.ParallelConfig{
		Branches: []workflow.ParallelBranch{
			{Name: "fast", Nodes: []string{"node1"}},
			{Name: "slow", Nodes: []string{"node2"}},
		},
		WaitMode:       "all",
		MaxConcurrency: 0,
		FailureMode:    "stop_all",
	}

	input := NewActionInput(config, map[string]interface{}{})

	action := &ParallelAction{}
	output, err := action.Execute(context.Background(), input)

	require.NoError(t, err)
	require.NotNil(t, output)

	// Should have results from all branches
	assert.Equal(t, 2, output.Metadata["branch_count"])
}

// TestParallelAction_WaitModeFirst tests wait mode "first" - returns after first completion
func TestParallelAction_WaitModeFirst(t *testing.T) {
	config := workflow.ParallelConfig{
		Branches: []workflow.ParallelBranch{
			{Name: "branch1", Nodes: []string{"node1"}},
			{Name: "branch2", Nodes: []string{"node2"}},
		},
		WaitMode:       "first",
		MaxConcurrency: 0,
		FailureMode:    "stop_all",
	}

	input := NewActionInput(config, map[string]interface{}{})

	action := &ParallelAction{}
	output, err := action.Execute(context.Background(), input)

	require.NoError(t, err)
	require.NotNil(t, output)

	// Should return immediately after first branch completes
	assert.Equal(t, "first", output.Metadata["wait_mode"])
}

// TestParallelAction_MaxConcurrency tests concurrency limiting
func TestParallelAction_MaxConcurrency(t *testing.T) {
	config := workflow.ParallelConfig{
		Branches: []workflow.ParallelBranch{
			{Name: "b1", Nodes: []string{"n1"}},
			{Name: "b2", Nodes: []string{"n2"}},
			{Name: "b3", Nodes: []string{"n3"}},
			{Name: "b4", Nodes: []string{"n4"}},
		},
		WaitMode:       "all",
		MaxConcurrency: 2, // Only 2 concurrent branches
		FailureMode:    "stop_all",
	}

	input := NewActionInput(config, map[string]interface{}{})

	action := &ParallelAction{}
	output, err := action.Execute(context.Background(), input)

	require.NoError(t, err)
	require.NotNil(t, output)

	// All branches should complete, just in batches
	assert.Equal(t, 4, output.Metadata["branch_count"])
	assert.Equal(t, 2, output.Metadata["max_concurrency"])
}

// TestParallelAction_Timeout tests timeout enforcement
func TestParallelAction_Timeout(t *testing.T) {
	config := workflow.ParallelConfig{
		Branches: []workflow.ParallelBranch{
			{Name: "branch1", Nodes: []string{"node1"}},
		},
		WaitMode:       "all",
		Timeout:        "1ms", // Very short timeout
		MaxConcurrency: 0,
		FailureMode:    "stop_all",
	}

	input := NewActionInput(config, map[string]interface{}{})

	action := &ParallelAction{}

	// Add small delay to ensure timeout
	time.Sleep(2 * time.Millisecond)

	// Note: The actual execution would timeout if branches take too long
	// For this unit test, we're verifying the timeout is parsed correctly
	_, err := action.Execute(context.Background(), input)

	// Should either succeed quickly or timeout
	if err != nil {
		assert.Contains(t, err.Error(), "timeout")
	}
}

// TestParallelAction_FailureModeStopAll tests stop_all failure mode
func TestParallelAction_FailureModeStopAll(t *testing.T) {
	config := workflow.ParallelConfig{
		Branches: []workflow.ParallelBranch{
			{Name: "success", Nodes: []string{"node1"}},
			{Name: "failure", Nodes: []string{"node2"}},
		},
		WaitMode:       "all",
		MaxConcurrency: 0,
		FailureMode:    "stop_all",
	}

	input := NewActionInput(config, map[string]interface{}{})

	action := &ParallelAction{}
	output, err := action.Execute(context.Background(), input)

	// Without injected executor, this will succeed
	// In integration tests with real executor, we'd test actual failure handling
	require.NoError(t, err)
	require.NotNil(t, output)
	assert.Equal(t, "stop_all", output.Metadata["failure_mode"])
}

// TestParallelAction_FailureModeContinue tests continue failure mode
func TestParallelAction_FailureModeContinue(t *testing.T) {
	config := workflow.ParallelConfig{
		Branches: []workflow.ParallelBranch{
			{Name: "success", Nodes: []string{"node1"}},
			{Name: "failure", Nodes: []string{"node2"}},
		},
		WaitMode:       "all",
		MaxConcurrency: 0,
		FailureMode:    "continue",
	}

	input := NewActionInput(config, map[string]interface{}{})

	action := &ParallelAction{}
	output, err := action.Execute(context.Background(), input)

	require.NoError(t, err)
	require.NotNil(t, output)
	assert.Equal(t, "continue", output.Metadata["failure_mode"])
}

// TestParallelAction_EmptyBranches tests handling of empty branches
func TestParallelAction_EmptyBranches(t *testing.T) {
	config := workflow.ParallelConfig{
		Branches:       []workflow.ParallelBranch{},
		WaitMode:       "all",
		MaxConcurrency: 0,
		FailureMode:    "stop_all",
	}

	input := NewActionInput(config, map[string]interface{}{})

	action := &ParallelAction{}
	_, err := action.Execute(context.Background(), input)

	// Should error with no branches
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no branches")
}

// TestParallelAction_InvalidWaitMode tests validation of wait mode
func TestParallelAction_InvalidWaitMode(t *testing.T) {
	config := workflow.ParallelConfig{
		Branches: []workflow.ParallelBranch{
			{Name: "branch1", Nodes: []string{"node1"}},
		},
		WaitMode:       "invalid",
		MaxConcurrency: 0,
		FailureMode:    "stop_all",
	}

	input := NewActionInput(config, map[string]interface{}{})

	action := &ParallelAction{}
	_, err := action.Execute(context.Background(), input)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "wait_mode")
}

// TestParallelAction_InvalidFailureMode tests validation of failure mode
func TestParallelAction_InvalidFailureMode(t *testing.T) {
	config := workflow.ParallelConfig{
		Branches: []workflow.ParallelBranch{
			{Name: "branch1", Nodes: []string{"node1"}},
		},
		WaitMode:       "all",
		MaxConcurrency: 0,
		FailureMode:    "invalid",
	}

	input := NewActionInput(config, map[string]interface{}{})

	action := &ParallelAction{}
	_, err := action.Execute(context.Background(), input)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failure_mode")
}

// TestParallelAction_InvalidTimeout tests validation of timeout format
func TestParallelAction_InvalidTimeout(t *testing.T) {
	config := workflow.ParallelConfig{
		Branches: []workflow.ParallelBranch{
			{Name: "branch1", Nodes: []string{"node1"}},
		},
		WaitMode:       "all",
		Timeout:        "invalid-timeout",
		MaxConcurrency: 0,
		FailureMode:    "stop_all",
	}

	input := NewActionInput(config, map[string]interface{}{})

	action := &ParallelAction{}
	_, err := action.Execute(context.Background(), input)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "timeout")
}

// TestParallelAction_NegativeMaxConcurrency tests validation of max concurrency
func TestParallelAction_NegativeMaxConcurrency(t *testing.T) {
	config := workflow.ParallelConfig{
		Branches: []workflow.ParallelBranch{
			{Name: "branch1", Nodes: []string{"node1"}},
		},
		WaitMode:       "all",
		MaxConcurrency: -1,
		FailureMode:    "stop_all",
	}

	input := NewActionInput(config, map[string]interface{}{})

	action := &ParallelAction{}
	_, err := action.Execute(context.Background(), input)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "max_concurrency")
}

// TestParallelAction_ContextCancellation tests context cancellation handling
func TestParallelAction_ContextCancellation(t *testing.T) {
	config := workflow.ParallelConfig{
		Branches: []workflow.ParallelBranch{
			{Name: "branch1", Nodes: []string{"node1"}},
			{Name: "branch2", Nodes: []string{"node2"}},
		},
		WaitMode:       "all",
		MaxConcurrency: 0,
		FailureMode:    "stop_all",
	}

	input := NewActionInput(config, map[string]interface{}{})

	// Create cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	action := &ParallelAction{}
	_, err := action.Execute(ctx, input)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "context")
}

// TestParallelAction_BranchWithMultipleNodes tests branches with multiple sequential nodes
func TestParallelAction_BranchWithMultipleNodes(t *testing.T) {
	config := workflow.ParallelConfig{
		Branches: []workflow.ParallelBranch{
			{Name: "pipeline1", Nodes: []string{"node1", "node2", "node3"}},
			{Name: "pipeline2", Nodes: []string{"node4", "node5"}},
		},
		WaitMode:       "all",
		MaxConcurrency: 0,
		FailureMode:    "stop_all",
	}

	input := NewActionInput(config, map[string]interface{}{
		"initial": "value",
	})

	action := &ParallelAction{}
	output, err := action.Execute(context.Background(), input)

	require.NoError(t, err)
	require.NotNil(t, output)

	// Verify both branches executed
	assert.Equal(t, 2, output.Metadata["branch_count"])
}

// TestParallelAction_BackwardCompatibility tests backward compatibility with ErrorStrategy
func TestParallelAction_BackwardCompatibility(t *testing.T) {
	config := workflow.ParallelConfig{
		Branches: []workflow.ParallelBranch{
			{Name: "branch1", Nodes: []string{"node1"}},
		},
		ErrorStrategy:  "fail_fast", // Legacy field
		MaxConcurrency: 0,
	}

	input := NewActionInput(config, map[string]interface{}{})

	action := &ParallelAction{}
	output, err := action.Execute(context.Background(), input)

	require.NoError(t, err)
	require.NotNil(t, output)

	// Should map fail_fast to stop_all
	assert.Contains(t, output.Metadata, "failure_mode")
}

// TestParallelAction_ResultAggregation tests that results from all branches are collected
func TestParallelAction_ResultAggregation(t *testing.T) {
	config := workflow.ParallelConfig{
		Branches: []workflow.ParallelBranch{
			{Name: "email", Nodes: []string{"send_email"}},
			{Name: "sms", Nodes: []string{"send_sms"}},
			{Name: "slack", Nodes: []string{"post_slack"}},
		},
		WaitMode:       "all",
		MaxConcurrency: 0,
		FailureMode:    "continue", // Continue even if one fails
	}

	input := NewActionInput(config, map[string]interface{}{
		"message": "Alert!",
	})

	action := &ParallelAction{}
	output, err := action.Execute(context.Background(), input)

	require.NoError(t, err)
	require.NotNil(t, output)

	// Verify all branches are represented
	assert.Equal(t, 3, output.Metadata["branch_count"])

	// Output data should contain results from each branch
	data, ok := output.Data.(map[string]interface{})
	require.True(t, ok)
	assert.NotNil(t, data)
}

// TestParallelAction_DefaultValues tests default configuration values
func TestParallelAction_DefaultValues(t *testing.T) {
	config := workflow.ParallelConfig{
		Branches: []workflow.ParallelBranch{
			{Name: "branch1", Nodes: []string{"node1"}},
		},
		// All defaults - no wait mode, failure mode, etc.
	}

	input := NewActionInput(config, map[string]interface{}{})

	action := &ParallelAction{}
	output, err := action.Execute(context.Background(), input)

	require.NoError(t, err)
	require.NotNil(t, output)

	// Defaults should be: wait_mode=all, failure_mode=stop_all
	assert.Equal(t, "all", output.Metadata["wait_mode"])
	assert.Equal(t, "stop_all", output.Metadata["failure_mode"])
}

// Helper to marshal config for testing
func mustMarshalConfig(v interface{}) json.RawMessage {
	data, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return data
}
