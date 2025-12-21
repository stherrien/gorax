package executor

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gorax/gorax/internal/workflow"
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

// Tests for findLoopBody function

func TestFindLoopBody_LinearChain(t *testing.T) {
	// Test: Loop -> Node1 -> Node2 -> Node3 -> NodeAfterLoop
	// Expected: Loop body should include Node1, Node2, Node3
	// NodeAfterLoop is connected from something else, not part of loop

	definition := &workflow.WorkflowDefinition{
		Nodes: []workflow.Node{
			{ID: "trigger", Type: string(workflow.NodeTypeTriggerWebhook)},
			{ID: "loop1", Type: string(workflow.NodeTypeControlLoop)},
			{ID: "node1", Type: string(workflow.NodeTypeActionHTTP)},
			{ID: "node2", Type: string(workflow.NodeTypeActionTransform)},
			{ID: "node3", Type: string(workflow.NodeTypeActionHTTP)},
			{ID: "node_after", Type: string(workflow.NodeTypeActionHTTP)},
		},
		Edges: []workflow.Edge{
			{ID: "e1", Source: "trigger", Target: "loop1"},
			{ID: "e2", Source: "loop1", Target: "node1"}, // Loop starts here
			{ID: "e3", Source: "node1", Target: "node2"},
			{ID: "e4", Source: "node2", Target: "node3"},
			{ID: "e5", Source: "loop1", Target: "node_after"}, // Exit loop - goes directly from loop to after
		},
	}

	exec := &Executor{}
	bodyNodes, bodyEdges := exec.findLoopBody("loop1", definition)

	// Should find all nodes in the chain that are reachable from loop
	// but not the node_after (which is outside the loop body)
	bodyNodeIDs := make(map[string]bool)
	for _, node := range bodyNodes {
		bodyNodeIDs[node.ID] = true
	}

	// Should include nodes in the loop body
	assert.True(t, bodyNodeIDs["node1"], "node1 should be in loop body")
	assert.True(t, bodyNodeIDs["node2"], "node2 should be in loop body")
	assert.True(t, bodyNodeIDs["node3"], "node3 should be in loop body")

	// Should NOT include node_after (it's outside loop body)
	assert.False(t, bodyNodeIDs["node_after"], "node_after should not be in loop body")
	assert.False(t, bodyNodeIDs["trigger"], "trigger should not be in loop body")
	assert.False(t, bodyNodeIDs["loop1"], "loop node itself should not be in loop body")

	// Check edges - should only include edges between body nodes
	bodyEdgeMap := make(map[string]bool)
	for _, edge := range bodyEdges {
		bodyEdgeMap[edge.ID] = true
	}

	assert.True(t, bodyEdgeMap["e3"], "edge between node1->node2 should be included")
	assert.True(t, bodyEdgeMap["e4"], "edge between node2->node3 should be included")
	assert.False(t, bodyEdgeMap["e1"], "edge trigger->loop should not be included")
	assert.False(t, bodyEdgeMap["e5"], "edge loop->node_after should not be included")
}

func TestFindLoopBody_BranchingPaths(t *testing.T) {
	// Test: Loop -> Condition -> [TruePath: Node1, Node2] / [FalsePath: Node3]
	// All merge back before exiting loop
	// Expected: All nodes in both branches should be in loop body

	definition := &workflow.WorkflowDefinition{
		Nodes: []workflow.Node{
			{ID: "loop1", Type: string(workflow.NodeTypeControlLoop)},
			{ID: "condition", Type: string(workflow.NodeTypeControlIf)},
			{ID: "node1", Type: string(workflow.NodeTypeActionHTTP)},
			{ID: "node2", Type: string(workflow.NodeTypeActionTransform)},
			{ID: "node3", Type: string(workflow.NodeTypeActionHTTP)},
			{ID: "merge", Type: string(workflow.NodeTypeActionTransform)},
		},
		Edges: []workflow.Edge{
			{ID: "e1", Source: "loop1", Target: "condition"},
			{ID: "e2", Source: "condition", Target: "node1", Label: "true"},
			{ID: "e3", Source: "node1", Target: "node2"},
			{ID: "e4", Source: "condition", Target: "node3", Label: "false"},
			{ID: "e5", Source: "node2", Target: "merge"},
			{ID: "e6", Source: "node3", Target: "merge"},
		},
	}

	exec := &Executor{}
	bodyNodes, bodyEdges := exec.findLoopBody("loop1", definition)

	bodyNodeIDs := make(map[string]bool)
	for _, node := range bodyNodes {
		bodyNodeIDs[node.ID] = true
	}

	// All nodes should be in loop body
	assert.True(t, bodyNodeIDs["condition"], "condition should be in loop body")
	assert.True(t, bodyNodeIDs["node1"], "node1 (true branch) should be in loop body")
	assert.True(t, bodyNodeIDs["node2"], "node2 (true branch) should be in loop body")
	assert.True(t, bodyNodeIDs["node3"], "node3 (false branch) should be in loop body")
	assert.True(t, bodyNodeIDs["merge"], "merge node should be in loop body")

	// Verify we got all expected edges
	assert.Len(t, bodyEdges, 6, "should have all 6 edges in loop body")
}

func TestFindLoopBody_NestedLoop(t *testing.T) {
	// Test: OuterLoop -> Node1 -> InnerLoop -> Node2 -> Node3
	// Expected: OuterLoop body should include Node1, InnerLoop (as a node), Node2, Node3
	// InnerLoop body detection is separate

	definition := &workflow.WorkflowDefinition{
		Nodes: []workflow.Node{
			{ID: "outer_loop", Type: string(workflow.NodeTypeControlLoop)},
			{ID: "node1", Type: string(workflow.NodeTypeActionHTTP)},
			{ID: "inner_loop", Type: string(workflow.NodeTypeControlLoop)},
			{ID: "node2", Type: string(workflow.NodeTypeActionTransform)},
			{ID: "node3", Type: string(workflow.NodeTypeActionHTTP)},
		},
		Edges: []workflow.Edge{
			{ID: "e1", Source: "outer_loop", Target: "node1"},
			{ID: "e2", Source: "node1", Target: "inner_loop"},
			{ID: "e3", Source: "inner_loop", Target: "node2"},
			{ID: "e4", Source: "node2", Target: "node3"},
		},
	}

	exec := &Executor{}

	// Test outer loop body detection
	outerBodyNodes, outerBodyEdges := exec.findLoopBody("outer_loop", definition)
	outerBodyNodeIDs := make(map[string]bool)
	for _, node := range outerBodyNodes {
		outerBodyNodeIDs[node.ID] = true
	}

	// Outer loop should include everything reachable from it
	assert.True(t, outerBodyNodeIDs["node1"], "node1 should be in outer loop body")
	assert.True(t, outerBodyNodeIDs["inner_loop"], "inner_loop should be in outer loop body")
	assert.True(t, outerBodyNodeIDs["node2"], "node2 should be in outer loop body")
	assert.True(t, outerBodyNodeIDs["node3"], "node3 should be in outer loop body")
	assert.Len(t, outerBodyEdges, 4, "should have all 4 edges in outer loop body")

	// Test inner loop body detection
	innerBodyNodes, innerBodyEdges := exec.findLoopBody("inner_loop", definition)
	innerBodyNodeIDs := make(map[string]bool)
	for _, node := range innerBodyNodes {
		innerBodyNodeIDs[node.ID] = true
	}

	// Inner loop should only include nodes reachable from it
	assert.False(t, innerBodyNodeIDs["node1"], "node1 should NOT be in inner loop body")
	assert.True(t, innerBodyNodeIDs["node2"], "node2 should be in inner loop body")
	assert.True(t, innerBodyNodeIDs["node3"], "node3 should be in inner loop body")
	assert.Len(t, innerBodyEdges, 2, "should have 2 edges in inner loop body")
}

func TestFindLoopBody_WithExitEdge(t *testing.T) {
	// Test: Loop has an exit edge that bypasses the loop body
	// Loop -> Node1 -> Node2
	//  |                  |
	//  +---> NodeAfter <--+
	// Expected: Loop body should include Node1, Node2 but not NodeAfter

	definition := &workflow.WorkflowDefinition{
		Nodes: []workflow.Node{
			{ID: "loop1", Type: string(workflow.NodeTypeControlLoop)},
			{ID: "node1", Type: string(workflow.NodeTypeActionHTTP)},
			{ID: "node2", Type: string(workflow.NodeTypeActionTransform)},
			{ID: "node_after", Type: string(workflow.NodeTypeActionHTTP)},
		},
		Edges: []workflow.Edge{
			{ID: "e1", Source: "loop1", Target: "node1"},      // Enter loop body
			{ID: "e2", Source: "node1", Target: "node2"},      // Within loop body
			{ID: "e3", Source: "node2", Target: "node_after"}, // Exit from loop body
			{ID: "e4", Source: "loop1", Target: "node_after"}, // Direct exit (bypass loop body)
		},
	}

	exec := &Executor{}
	bodyNodes, bodyEdges := exec.findLoopBody("loop1", definition)

	bodyNodeIDs := make(map[string]bool)
	for _, node := range bodyNodes {
		bodyNodeIDs[node.ID] = true
	}

	// Loop body should include the chain of nodes
	assert.True(t, bodyNodeIDs["node1"], "node1 should be in loop body")
	assert.True(t, bodyNodeIDs["node2"], "node2 should be in loop body")

	// node_after is not in loop body (it's the exit point)
	assert.False(t, bodyNodeIDs["node_after"], "node_after should not be in loop body")

	// Check edges within loop body
	bodyEdgeMap := make(map[string]bool)
	for _, edge := range bodyEdges {
		bodyEdgeMap[edge.ID] = true
	}

	assert.True(t, bodyEdgeMap["e2"], "edge node1->node2 should be in loop body")
	assert.False(t, bodyEdgeMap["e3"], "exit edge node2->node_after should not be in loop body")
	assert.False(t, bodyEdgeMap["e4"], "direct exit edge should not be in loop body")
}

func TestFindLoopBody_EmptyLoop(t *testing.T) {
	// Test: Loop with no body nodes (edge goes directly to exit)
	// Loop -> NodeAfter
	// Expected: Empty loop body

	definition := &workflow.WorkflowDefinition{
		Nodes: []workflow.Node{
			{ID: "loop1", Type: string(workflow.NodeTypeControlLoop)},
			{ID: "node_after", Type: string(workflow.NodeTypeActionHTTP)},
		},
		Edges: []workflow.Edge{
			{ID: "e1", Source: "loop1", Target: "node_after"},
		},
	}

	exec := &Executor{}
	bodyNodes, bodyEdges := exec.findLoopBody("loop1", definition)

	// Should have no body nodes or edges
	assert.Len(t, bodyNodes, 0, "empty loop should have no body nodes")
	assert.Len(t, bodyEdges, 0, "empty loop should have no body edges")
}

func TestFindLoopBody_MultipleExitPoints(t *testing.T) {
	// Test: Loop with multiple body nodes that branch and both exit to the same node
	// The loop uses an explicit exit edge to mark node_after as outside the body
	//
	// Loop -> Node1 -> Node2 ----+
	//   |         |-> Node3 -----|---> NodeAfter
	//   +---------------------------->
	//
	// Expected: Node1, Node2, Node3 in body; NodeAfter outside (marked by explicit exit)

	definition := &workflow.WorkflowDefinition{
		Nodes: []workflow.Node{
			{ID: "loop1", Type: string(workflow.NodeTypeControlLoop)},
			{ID: "node1", Type: string(workflow.NodeTypeActionHTTP)},
			{ID: "node2", Type: string(workflow.NodeTypeActionTransform)},
			{ID: "node3", Type: string(workflow.NodeTypeActionHTTP)},
			{ID: "node_after", Type: string(workflow.NodeTypeActionTransform)},
		},
		Edges: []workflow.Edge{
			{ID: "e1", Source: "loop1", Target: "node1"},      // Body entrance
			{ID: "e2", Source: "node1", Target: "node2"},      // Branch 1
			{ID: "e3", Source: "node1", Target: "node3"},      // Branch 2
			{ID: "e4", Source: "node2", Target: "node_after"}, // Exit from branch 1
			{ID: "e5", Source: "node3", Target: "node_after"}, // Exit from branch 2
			{ID: "e6", Source: "loop1", Target: "node_after"}, // Explicit loop exit
		},
	}

	exec := &Executor{}
	bodyNodes, bodyEdges := exec.findLoopBody("loop1", definition)

	bodyNodeIDs := make(map[string]bool)
	for _, node := range bodyNodes {
		bodyNodeIDs[node.ID] = true
	}

	// Should include all nodes in the loop body
	assert.True(t, bodyNodeIDs["node1"], "node1 should be in loop body")
	assert.True(t, bodyNodeIDs["node2"], "node2 should be in loop body")
	assert.True(t, bodyNodeIDs["node3"], "node3 should be in loop body")

	// Should not include the exit node (marked by explicit loop exit edge)
	assert.False(t, bodyNodeIDs["node_after"], "node_after should not be in loop body")

	// Check that internal edges are included
	bodyEdgeMap := make(map[string]bool)
	for _, edge := range bodyEdges {
		bodyEdgeMap[edge.ID] = true
	}

	assert.True(t, bodyEdgeMap["e2"], "edge node1->node2 should be in loop body")
	assert.True(t, bodyEdgeMap["e3"], "edge node1->node3 should be in loop body")
	assert.False(t, bodyEdgeMap["e4"], "exit edge node2->node_after should not be in loop body")
	assert.False(t, bodyEdgeMap["e5"], "exit edge node3->node_after should not be in loop body")
	assert.False(t, bodyEdgeMap["e6"], "explicit loop exit edge should not be in loop body")
}
