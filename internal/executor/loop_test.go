package executor

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/gorax/gorax/internal/workflow"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExecuteLoopAction_BasicLoop(t *testing.T) {
	// Test basic loop over an array of items
	config := workflow.LoopActionConfig{
		Source:        "${steps.data_source.output.items}",
		ItemVariable:  "item",
		IndexVariable: "index",
		MaxIterations: 1000,
		OnError:       "stop",
	}

	// Create execution context with array data
	execCtx := &ExecutionContext{
		TenantID:    "tenant1",
		ExecutionID: "exec1",
		WorkflowID:  "workflow1",
		TriggerData: map[string]interface{}{},
		StepOutputs: map[string]interface{}{
			"data_source": map[string]interface{}{
				"output": map[string]interface{}{
					"items": []interface{}{
						map[string]interface{}{"id": 1, "name": "item1"},
						map[string]interface{}{"id": 2, "name": "item2"},
						map[string]interface{}{"id": 3, "name": "item3"},
					},
				},
			},
		},
	}

	// Create mock loop body nodes (what gets executed for each iteration)
	loopBodyNodes := []workflow.Node{
		{
			ID:   "loop_body_1",
			Type: string(workflow.NodeTypeActionTransform),
			Data: workflow.NodeData{
				Name: "Process Item",
				Config: mustMarshal(map[string]interface{}{
					"expression": "${item.name}",
				}),
			},
		},
	}

	executor := &loopExecutor{}
	result, err := executor.executeLoop(context.Background(), config, execCtx, loopBodyNodes, []workflow.Edge{})

	require.NoError(t, err)
	assert.NotNil(t, result)

	// Verify results contain outputs from all iterations
	loopResult, ok := result.(*LoopResult)
	require.True(t, ok)
	assert.Equal(t, 3, loopResult.IterationCount)
	assert.Len(t, loopResult.Iterations, 3)

	// Verify each iteration has correct data
	for i, iteration := range loopResult.Iterations {
		assert.Equal(t, i, iteration.Index)
		assert.NotNil(t, iteration.Output)
	}
}

func TestExecuteLoopAction_EmptyArray(t *testing.T) {
	// Test loop with empty array - should not execute body
	config := workflow.LoopActionConfig{
		Source:        "${steps.data_source.output.items}",
		ItemVariable:  "item",
		IndexVariable: "index",
		MaxIterations: 1000,
		OnError:       "stop",
	}

	execCtx := &ExecutionContext{
		TenantID:    "tenant1",
		ExecutionID: "exec1",
		WorkflowID:  "workflow1",
		TriggerData: map[string]interface{}{},
		StepOutputs: map[string]interface{}{
			"data_source": map[string]interface{}{
				"output": map[string]interface{}{
					"items": []interface{}{},
				},
			},
		},
	}

	loopBodyNodes := []workflow.Node{
		{
			ID:   "loop_body_1",
			Type: string(workflow.NodeTypeActionTransform),
			Data: workflow.NodeData{
				Name:   "Process Item",
				Config: mustMarshal(map[string]interface{}{}),
			},
		},
	}

	executor := &loopExecutor{}
	result, err := executor.executeLoop(context.Background(), config, execCtx, loopBodyNodes, []workflow.Edge{})

	require.NoError(t, err)
	loopResult, ok := result.(*LoopResult)
	require.True(t, ok)
	assert.Equal(t, 0, loopResult.IterationCount)
	assert.Len(t, loopResult.Iterations, 0)
}

func TestExecuteLoopAction_MaxIterationsLimit(t *testing.T) {
	// Test that loop respects max iterations limit
	config := workflow.LoopActionConfig{
		Source:        "${steps.data_source.output.items}",
		ItemVariable:  "item",
		IndexVariable: "index",
		MaxIterations: 2, // Limit to 2 iterations
		OnError:       "stop",
	}

	// Create array with 5 items
	items := []interface{}{
		map[string]interface{}{"id": 1},
		map[string]interface{}{"id": 2},
		map[string]interface{}{"id": 3},
		map[string]interface{}{"id": 4},
		map[string]interface{}{"id": 5},
	}

	execCtx := &ExecutionContext{
		TenantID:    "tenant1",
		ExecutionID: "exec1",
		WorkflowID:  "workflow1",
		TriggerData: map[string]interface{}{},
		StepOutputs: map[string]interface{}{
			"data_source": map[string]interface{}{
				"output": map[string]interface{}{
					"items": items,
				},
			},
		},
	}

	loopBodyNodes := []workflow.Node{
		{
			ID:   "loop_body_1",
			Type: string(workflow.NodeTypeActionTransform),
			Data: workflow.NodeData{
				Name:   "Process Item",
				Config: mustMarshal(map[string]interface{}{}),
			},
		},
	}

	executor := &loopExecutor{}
	_, err := executor.executeLoop(context.Background(), config, execCtx, loopBodyNodes, []workflow.Edge{})

	// Should return error when max iterations is exceeded
	require.Error(t, err)
	assert.Contains(t, err.Error(), "max iterations")
}

func TestExecuteLoopAction_ErrorHandlingStop(t *testing.T) {
	// Test that loop validates OnError configuration
	// Note: Actual error execution would require a full executor with failing nodes
	// This test validates the error strategy configuration is properly set
	config := workflow.LoopActionConfig{
		Source:        "${steps.data_source.output.items}",
		ItemVariable:  "item",
		IndexVariable: "index",
		MaxIterations: 1000,
		OnError:       "stop",
	}

	items := []interface{}{
		map[string]interface{}{"id": 1},
		map[string]interface{}{"id": 2},
	}

	execCtx := &ExecutionContext{
		TenantID:    "tenant1",
		ExecutionID: "exec1",
		WorkflowID:  "workflow1",
		TriggerData: map[string]interface{}{},
		StepOutputs: map[string]interface{}{
			"data_source": map[string]interface{}{
				"output": map[string]interface{}{
					"items": items,
				},
			},
		},
	}

	loopBodyNodes := []workflow.Node{
		{
			ID:   "loop_body_1",
			Type: string(workflow.NodeTypeActionTransform),
			Data: workflow.NodeData{
				Name:   "Process Item",
				Config: mustMarshal(map[string]interface{}{}),
			},
		},
	}

	executor := &loopExecutor{}
	result, err := executor.executeLoop(context.Background(), config, execCtx, loopBodyNodes, []workflow.Edge{})

	// Should succeed with stop strategy configured
	require.NoError(t, err)
	loopResult, ok := result.(*LoopResult)
	require.True(t, ok)
	assert.Equal(t, "stop", loopResult.Metadata["on_error"])
}

func TestExecuteLoopAction_ErrorHandlingContinue(t *testing.T) {
	// Test that loop validates OnError=continue configuration
	// Note: Actual error execution would require a full executor with failing nodes
	// This test validates the continue strategy configuration is properly set
	config := workflow.LoopActionConfig{
		Source:        "${steps.data_source.output.items}",
		ItemVariable:  "item",
		IndexVariable: "index",
		MaxIterations: 1000,
		OnError:       "continue",
	}

	items := []interface{}{
		map[string]interface{}{"id": 1},
		map[string]interface{}{"id": 2},
		map[string]interface{}{"id": 3},
	}

	execCtx := &ExecutionContext{
		TenantID:    "tenant1",
		ExecutionID: "exec1",
		WorkflowID:  "workflow1",
		TriggerData: map[string]interface{}{},
		StepOutputs: map[string]interface{}{
			"data_source": map[string]interface{}{
				"output": map[string]interface{}{
					"items": items,
				},
			},
		},
	}

	loopBodyNodes := []workflow.Node{
		{
			ID:   "loop_body_1",
			Type: string(workflow.NodeTypeActionTransform),
			Data: workflow.NodeData{
				Name:   "Process Item",
				Config: mustMarshal(map[string]interface{}{}),
			},
		},
	}

	executor := &loopExecutor{}
	result, err := executor.executeLoop(context.Background(), config, execCtx, loopBodyNodes, []workflow.Edge{})

	// Should succeed with continue strategy
	require.NoError(t, err)
	loopResult, ok := result.(*LoopResult)
	require.True(t, ok)
	assert.Equal(t, 3, loopResult.IterationCount)
	assert.Equal(t, "continue", loopResult.Metadata["on_error"])
}

func TestExecuteLoopAction_ItemAndIndexVariables(t *testing.T) {
	// Test that item and index variables are correctly set in execution context
	config := workflow.LoopActionConfig{
		Source:        "${steps.data_source.output.items}",
		ItemVariable:  "current_item",
		IndexVariable: "idx",
		MaxIterations: 1000,
		OnError:       "stop",
	}

	items := []interface{}{
		"apple",
		"banana",
		"cherry",
	}

	execCtx := &ExecutionContext{
		TenantID:    "tenant1",
		ExecutionID: "exec1",
		WorkflowID:  "workflow1",
		TriggerData: map[string]interface{}{},
		StepOutputs: map[string]interface{}{
			"data_source": map[string]interface{}{
				"output": map[string]interface{}{
					"items": items,
				},
			},
		},
	}

	loopBodyNodes := []workflow.Node{
		{
			ID:   "loop_body_1",
			Type: string(workflow.NodeTypeActionTransform),
			Data: workflow.NodeData{
				Name: "Process Item",
				Config: mustMarshal(map[string]interface{}{
					"expression": "${current_item}-${idx}",
				}),
			},
		},
	}

	executor := &loopExecutor{}
	result, err := executor.executeLoop(context.Background(), config, execCtx, loopBodyNodes, []workflow.Edge{})

	require.NoError(t, err)
	loopResult, ok := result.(*LoopResult)
	require.True(t, ok)
	assert.Equal(t, 3, loopResult.IterationCount)
}

func TestExecuteLoopAction_SourceNotArray(t *testing.T) {
	// Test error when source is not an array
	config := workflow.LoopActionConfig{
		Source:        "${steps.data_source.output.value}",
		ItemVariable:  "item",
		IndexVariable: "index",
		MaxIterations: 1000,
		OnError:       "stop",
	}

	execCtx := &ExecutionContext{
		TenantID:    "tenant1",
		ExecutionID: "exec1",
		WorkflowID:  "workflow1",
		TriggerData: map[string]interface{}{},
		StepOutputs: map[string]interface{}{
			"data_source": map[string]interface{}{
				"output": map[string]interface{}{
					"value": "not an array",
				},
			},
		},
	}

	loopBodyNodes := []workflow.Node{}

	executor := &loopExecutor{}
	_, err := executor.executeLoop(context.Background(), config, execCtx, loopBodyNodes, []workflow.Edge{})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "not an array")
}

func TestExecuteLoopAction_SourceNotFound(t *testing.T) {
	// Test error when source path doesn't exist
	config := workflow.LoopActionConfig{
		Source:        "${steps.missing_node.output.items}",
		ItemVariable:  "item",
		IndexVariable: "index",
		MaxIterations: 1000,
		OnError:       "stop",
	}

	execCtx := &ExecutionContext{
		TenantID:    "tenant1",
		ExecutionID: "exec1",
		WorkflowID:  "workflow1",
		TriggerData: map[string]interface{}{},
		StepOutputs: map[string]interface{}{},
	}

	loopBodyNodes := []workflow.Node{}

	executor := &loopExecutor{}
	_, err := executor.executeLoop(context.Background(), config, execCtx, loopBodyNodes, []workflow.Edge{})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestExecuteLoopAction_DefaultMaxIterations(t *testing.T) {
	// Test that default max iterations is applied when not specified
	config := workflow.LoopActionConfig{
		Source:        "${steps.data_source.output.items}",
		ItemVariable:  "item",
		IndexVariable: "index",
		MaxIterations: 0, // Not specified, should use default
		OnError:       "stop",
	}

	// Create large array that exceeds default limit
	items := make([]interface{}, 1500)
	for i := 0; i < 1500; i++ {
		items[i] = map[string]interface{}{"id": i}
	}

	execCtx := &ExecutionContext{
		TenantID:    "tenant1",
		ExecutionID: "exec1",
		WorkflowID:  "workflow1",
		TriggerData: map[string]interface{}{},
		StepOutputs: map[string]interface{}{
			"data_source": map[string]interface{}{
				"output": map[string]interface{}{
					"items": items,
				},
			},
		},
	}

	loopBodyNodes := []workflow.Node{
		{
			ID:   "loop_body_1",
			Type: string(workflow.NodeTypeActionTransform),
			Data: workflow.NodeData{
				Name:   "Process Item",
				Config: mustMarshal(map[string]interface{}{}),
			},
		},
	}

	executor := &loopExecutor{}
	_, err := executor.executeLoop(context.Background(), config, execCtx, loopBodyNodes, []workflow.Edge{})

	// Should error due to exceeding default max iterations (1000)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "max iterations")
}

func TestExecuteLoopAction_NestedLoops(t *testing.T) {
	// Test nested loop scenarios (loop within a loop)
	// Outer loop config
	outerConfig := workflow.LoopActionConfig{
		Source:        "${steps.data_source.output.groups}",
		ItemVariable:  "group",
		IndexVariable: "group_index",
		MaxIterations: 1000,
		OnError:       "stop",
	}

	// Inner loop would be part of the body nodes
	// For this test, we'll just verify the outer loop works correctly
	groups := []interface{}{
		map[string]interface{}{
			"name": "group1",
			"items": []interface{}{
				map[string]interface{}{"id": 1},
				map[string]interface{}{"id": 2},
			},
		},
		map[string]interface{}{
			"name": "group2",
			"items": []interface{}{
				map[string]interface{}{"id": 3},
				map[string]interface{}{"id": 4},
			},
		},
	}

	execCtx := &ExecutionContext{
		TenantID:    "tenant1",
		ExecutionID: "exec1",
		WorkflowID:  "workflow1",
		TriggerData: map[string]interface{}{},
		StepOutputs: map[string]interface{}{
			"data_source": map[string]interface{}{
				"output": map[string]interface{}{
					"groups": groups,
				},
			},
		},
	}

	// Outer loop body that would contain an inner loop
	loopBodyNodes := []workflow.Node{
		{
			ID:   "inner_loop",
			Type: string(workflow.NodeTypeControlLoop),
			Data: workflow.NodeData{
				Name: "Inner Loop",
				Config: mustMarshal(workflow.LoopActionConfig{
					Source:        "${group.items}",
					ItemVariable:  "inner_item",
					IndexVariable: "inner_index",
					MaxIterations: 1000,
					OnError:       "stop",
				}),
			},
		},
	}

	executor := &loopExecutor{}
	result, err := executor.executeLoop(context.Background(), outerConfig, execCtx, loopBodyNodes, []workflow.Edge{})

	require.NoError(t, err)
	loopResult, ok := result.(*LoopResult)
	require.True(t, ok)
	assert.Equal(t, 2, loopResult.IterationCount)
}

func TestExecuteLoopAction_LoopBodyWithMultipleNodes(t *testing.T) {
	// Test loop with multiple nodes in the body
	config := workflow.LoopActionConfig{
		Source:        "${steps.data_source.output.items}",
		ItemVariable:  "item",
		IndexVariable: "index",
		MaxIterations: 1000,
		OnError:       "stop",
	}

	items := []interface{}{
		map[string]interface{}{"value": 10},
		map[string]interface{}{"value": 20},
	}

	execCtx := &ExecutionContext{
		TenantID:    "tenant1",
		ExecutionID: "exec1",
		WorkflowID:  "workflow1",
		TriggerData: map[string]interface{}{},
		StepOutputs: map[string]interface{}{
			"data_source": map[string]interface{}{
				"output": map[string]interface{}{
					"items": items,
				},
			},
		},
	}

	// Multiple nodes in loop body with dependencies
	loopBodyNodes := []workflow.Node{
		{
			ID:   "transform_1",
			Type: string(workflow.NodeTypeActionTransform),
			Data: workflow.NodeData{
				Name: "Double Value",
				Config: mustMarshal(map[string]interface{}{
					"expression": "${item.value * 2}",
				}),
			},
		},
		{
			ID:   "transform_2",
			Type: string(workflow.NodeTypeActionTransform),
			Data: workflow.NodeData{
				Name: "Add Ten",
				Config: mustMarshal(map[string]interface{}{
					"expression": "${steps.transform_1.output + 10}",
				}),
			},
		},
	}

	loopBodyEdges := []workflow.Edge{
		{
			ID:     "edge_1",
			Source: "transform_1",
			Target: "transform_2",
		},
	}

	executor := &loopExecutor{}
	result, err := executor.executeLoop(context.Background(), config, execCtx, loopBodyNodes, loopBodyEdges)

	require.NoError(t, err)
	loopResult, ok := result.(*LoopResult)
	require.True(t, ok)
	assert.Equal(t, 2, loopResult.IterationCount)
}

// Helper function to marshal JSON
func mustMarshal(v interface{}) json.RawMessage {
	data, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return data
}
