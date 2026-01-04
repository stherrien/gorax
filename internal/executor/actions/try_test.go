package actions

import (
	"context"
	"errors"
	"testing"

	"github.com/gorax/gorax/internal/workflow"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTryAction_Execute_Success(t *testing.T) {
	// Test successful try block execution without errors
	nodeOutputs := map[string]interface{}{
		"node1": map[string]interface{}{"result": "success"},
		"node2": map[string]interface{}{"result": "done"},
	}

	executeNode := func(ctx context.Context, nodeID string) (interface{}, error) {
		return nodeOutputs[nodeID], nil
	}

	getNode := func(nodeID string) (*workflow.Node, error) {
		return &workflow.Node{ID: nodeID, Type: "action:http"}, nil
	}

	action := NewTryAction(executeNode, getNode)

	config := workflow.TryConfig{
		TryNodes: []string{"node1", "node2"},
	}

	input := NewActionInput(config, map[string]interface{}{})

	output, err := action.Execute(context.Background(), input)
	require.NoError(t, err)
	require.NotNil(t, output)

	data, ok := output.Data.(map[string]interface{})
	require.True(t, ok)
	assert.True(t, data["success"].(bool))
	assert.False(t, data["error_handled"].(bool))
	assert.Contains(t, data, "try_output")
}

func TestTryAction_Execute_WithFinally(t *testing.T) {
	// Test successful try block with finally block
	nodeOutputs := map[string]interface{}{
		"node1":   map[string]interface{}{"result": "success"},
		"finally1": map[string]interface{}{"cleanup": "done"},
	}

	executeNode := func(ctx context.Context, nodeID string) (interface{}, error) {
		return nodeOutputs[nodeID], nil
	}

	getNode := func(nodeID string) (*workflow.Node, error) {
		return &workflow.Node{ID: nodeID, Type: "action:http"}, nil
	}

	action := NewTryAction(executeNode, getNode)

	config := workflow.TryConfig{
		TryNodes:     []string{"node1"},
		FinallyNodes: []string{"finally1"},
	}

	input := NewActionInput(config, map[string]interface{}{})

	output, err := action.Execute(context.Background(), input)
	require.NoError(t, err)
	require.NotNil(t, output)

	data, ok := output.Data.(map[string]interface{})
	require.True(t, ok)
	assert.True(t, data["success"].(bool))
	assert.Contains(t, data, "try_output")
	assert.Contains(t, data, "finally_output")
}

func TestTryAction_Execute_CatchError(t *testing.T) {
	// Test error in try block that gets caught
	testErr := errors.New("test error")
	executeNode := func(ctx context.Context, nodeID string) (interface{}, error) {
		if nodeID == "node1" {
			return nil, testErr
		}
		return map[string]interface{}{"handled": true}, nil
	}

	getNode := func(nodeID string) (*workflow.Node, error) {
		return &workflow.Node{ID: nodeID, Type: "action:http"}, nil
	}

	action := NewTryAction(executeNode, getNode)

	config := workflow.TryConfig{
		TryNodes:     []string{"node1"},
		CatchNodes:   []string{"catch1"},
		ErrorBinding: "err",
	}

	input := NewActionInput(config, map[string]interface{}{})

	output, err := action.Execute(context.Background(), input)
	require.NoError(t, err)
	require.NotNil(t, output)

	data, ok := output.Data.(map[string]interface{})
	require.True(t, ok)
	assert.True(t, data["success"].(bool))
	assert.True(t, data["error_handled"].(bool))
	assert.Contains(t, data, "catch_output")
	assert.Contains(t, data, "try_error")
}

func TestTryAction_Execute_CatchAndFinally(t *testing.T) {
	// Test error with both catch and finally blocks
	testErr := errors.New("test error")
	executeNode := func(ctx context.Context, nodeID string) (interface{}, error) {
		if nodeID == "node1" {
			return nil, testErr
		}
		return map[string]interface{}{"result": nodeID}, nil
	}

	getNode := func(nodeID string) (*workflow.Node, error) {
		return &workflow.Node{ID: nodeID, Type: "action:http"}, nil
	}

	action := NewTryAction(executeNode, getNode)

	config := workflow.TryConfig{
		TryNodes:     []string{"node1"},
		CatchNodes:   []string{"catch1"},
		FinallyNodes: []string{"finally1"},
		ErrorBinding: "error",
	}

	input := NewActionInput(config, map[string]interface{}{})

	output, err := action.Execute(context.Background(), input)
	require.NoError(t, err)
	require.NotNil(t, output)

	data, ok := output.Data.(map[string]interface{})
	require.True(t, ok)
	assert.True(t, data["success"].(bool))
	assert.True(t, data["error_handled"].(bool))
	assert.Contains(t, data, "catch_output")
	assert.Contains(t, data, "finally_output")
	assert.Contains(t, data, "try_error")
}

func TestTryAction_Execute_ErrorPropagation(t *testing.T) {
	// Test error propagation when no catch block
	testErr := errors.New("test error")
	executeNode := func(ctx context.Context, nodeID string) (interface{}, error) {
		if nodeID == "node1" {
			return nil, testErr
		}
		return map[string]interface{}{"result": nodeID}, nil
	}

	getNode := func(nodeID string) (*workflow.Node, error) {
		return &workflow.Node{ID: nodeID, Type: "action:http"}, nil
	}

	action := NewTryAction(executeNode, getNode)

	config := workflow.TryConfig{
		TryNodes:     []string{"node1"},
		FinallyNodes: []string{"finally1"}, // Finally should still run
	}

	input := NewActionInput(config, map[string]interface{}{})

	output, err := action.Execute(context.Background(), input)
	require.Error(t, err)
	assert.Nil(t, output)
	assert.Contains(t, err.Error(), "test error")
}

func TestTryAction_Execute_CatchBlockFailure(t *testing.T) {
	// Test catch block failure
	testErr := errors.New("test error")
	catchErr := errors.New("catch failed")

	executeNode := func(ctx context.Context, nodeID string) (interface{}, error) {
		if nodeID == "node1" {
			return nil, testErr
		}
		if nodeID == "catch1" {
			return nil, catchErr
		}
		return map[string]interface{}{"result": nodeID}, nil
	}

	getNode := func(nodeID string) (*workflow.Node, error) {
		return &workflow.Node{ID: nodeID, Type: "action:http"}, nil
	}

	action := NewTryAction(executeNode, getNode)

	config := workflow.TryConfig{
		TryNodes:   []string{"node1"},
		CatchNodes: []string{"catch1"},
	}

	input := NewActionInput(config, map[string]interface{}{})

	output, err := action.Execute(context.Background(), input)
	require.Error(t, err)
	assert.Nil(t, output)
	assert.Contains(t, err.Error(), "catch block failed")
	assert.Contains(t, err.Error(), "catch failed")
}

func TestTryAction_Execute_EmptyTryNodes(t *testing.T) {
	// Test validation of empty try nodes
	executeNode := func(ctx context.Context, nodeID string) (interface{}, error) {
		return nil, nil
	}

	getNode := func(nodeID string) (*workflow.Node, error) {
		return &workflow.Node{ID: nodeID, Type: "action:http"}, nil
	}

	action := NewTryAction(executeNode, getNode)

	config := workflow.TryConfig{
		TryNodes: []string{},
	}

	input := NewActionInput(config, map[string]interface{}{})

	output, err := action.Execute(context.Background(), input)
	require.Error(t, err)
	assert.Nil(t, output)
	assert.Contains(t, err.Error(), "try block must have at least one node")
}

func TestCatchAction_Execute_CatchAll(t *testing.T) {
	// Test catch action that catches all errors
	executeNode := func(ctx context.Context, nodeID string) (interface{}, error) {
		return map[string]interface{}{"result": "handled"}, nil
	}

	action := NewCatchAction(executeNode)

	errorData := &workflow.ErrorHandlingMetadata{
		ErrorType:      "TestError",
		ErrorMessage:   "test error message",
		Classification: "transient",
		NodeID:         "node1",
		NodeType:       "action:http",
	}

	config := workflow.CatchConfig{
		// No filters - catch all
	}

	input := NewActionInput(config, map[string]interface{}{
		"error": errorData,
	})

	output, err := action.Execute(context.Background(), input)
	require.NoError(t, err)
	require.NotNil(t, output)

	data, ok := output.Data.(map[string]interface{})
	require.True(t, ok)
	assert.True(t, data["error_handled"].(bool))
	assert.Contains(t, data, "caught_error")
}

func TestCatchAction_Execute_ErrorTypeFilter(t *testing.T) {
	// Test catch action with error type filtering
	executeNode := func(ctx context.Context, nodeID string) (interface{}, error) {
		return map[string]interface{}{"result": "handled"}, nil
	}

	action := NewCatchAction(executeNode)

	errorData := &workflow.ErrorHandlingMetadata{
		ErrorType:      "TestError",
		ErrorMessage:   "test error message",
		Classification: "transient",
		NodeID:         "node1",
		NodeType:       "action:http",
	}

	// Test matching error type
	config := workflow.CatchConfig{
		ErrorTypes: []string{"TestError", "AnotherError"},
	}

	input := NewActionInput(config, map[string]interface{}{
		"error": errorData,
	})

	output, err := action.Execute(context.Background(), input)
	require.NoError(t, err)
	require.NotNil(t, output)

	// Test non-matching error type
	config2 := workflow.CatchConfig{
		ErrorTypes: []string{"DifferentError"},
	}

	input2 := NewActionInput(config2, map[string]interface{}{
		"error": errorData,
	})

	output2, err2 := action.Execute(context.Background(), input2)
	require.Error(t, err2)
	assert.Nil(t, output2)
	assert.Contains(t, err2.Error(), "error not caught")
}

func TestCatchAction_Execute_ErrorPatternFilter(t *testing.T) {
	// Test catch action with error pattern filtering
	executeNode := func(ctx context.Context, nodeID string) (interface{}, error) {
		return map[string]interface{}{"result": "handled"}, nil
	}

	action := NewCatchAction(executeNode)

	errorData := &workflow.ErrorHandlingMetadata{
		ErrorType:      "TestError",
		ErrorMessage:   "connection timeout error",
		Classification: "transient",
		NodeID:         "node1",
		NodeType:       "action:http",
	}

	// Test matching pattern
	config := workflow.CatchConfig{
		ErrorPatterns: []string{"timeout", "connection.*error"},
	}

	input := NewActionInput(config, map[string]interface{}{
		"error": errorData,
	})

	output, err := action.Execute(context.Background(), input)
	require.NoError(t, err)
	require.NotNil(t, output)

	// Test non-matching pattern
	config2 := workflow.CatchConfig{
		ErrorPatterns: []string{"database.*error"},
	}

	input2 := NewActionInput(config2, map[string]interface{}{
		"error": errorData,
	})

	output2, err2 := action.Execute(context.Background(), input2)
	require.Error(t, err2)
	assert.Nil(t, output2)
}

func TestCatchAction_Execute_NoErrorInContext(t *testing.T) {
	// Test catch action when no error is in context
	executeNode := func(ctx context.Context, nodeID string) (interface{}, error) {
		return map[string]interface{}{"result": "handled"}, nil
	}

	action := NewCatchAction(executeNode)

	config := workflow.CatchConfig{}

	input := NewActionInput(config, map[string]interface{}{
		// No error in context
	})

	output, err := action.Execute(context.Background(), input)
	require.Error(t, err)
	assert.Nil(t, output)
	assert.Contains(t, err.Error(), "no error data found")
}
