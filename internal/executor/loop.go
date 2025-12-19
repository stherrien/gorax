package executor

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/gorax/gorax/internal/executor/actions"
	"github.com/gorax/gorax/internal/workflow"
)

const (
	// DefaultMaxIterations is the default maximum number of loop iterations
	DefaultMaxIterations = 1000
	// ErrorStrategyContinue continues loop execution on error
	ErrorStrategyContinue = "continue"
	// ErrorStrategyStop stops loop execution on error
	ErrorStrategyStop = "stop"
)

// LoopResult represents the result of a loop execution
type LoopResult struct {
	IterationCount int                  `json:"iteration_count"`
	Iterations     []IterationResult    `json:"iterations"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
}

// IterationResult represents the result of a single iteration
type IterationResult struct {
	Index  int                    `json:"index"`
	Item   interface{}            `json:"item"`
	Output map[string]interface{} `json:"output,omitempty"`
	Error  *string                `json:"error,omitempty"`
}

// loopExecutor handles loop execution logic
type loopExecutor struct {
	mainExecutor *Executor
}

// executeLoop executes a loop node with for-each semantics
func (le *loopExecutor) executeLoop(
	ctx context.Context,
	config workflow.LoopActionConfig,
	execCtx *ExecutionContext,
	bodyNodes []workflow.Node,
	bodyEdges []workflow.Edge,
) (interface{}, error) {
	// Validate configuration
	if err := le.validateConfig(config); err != nil {
		return nil, fmt.Errorf("invalid loop configuration: %w", err)
	}

	// Resolve source array from execution context
	sourceArray, err := le.resolveSourceArray(config.Source, execCtx)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve source array: %w", err)
	}

	// Apply max iterations limit
	maxIterations := config.MaxIterations
	if maxIterations <= 0 {
		maxIterations = DefaultMaxIterations
	}

	// Check if array exceeds max iterations
	arrayLen := len(sourceArray)
	if arrayLen > maxIterations {
		return nil, fmt.Errorf("array length %d exceeds max iterations limit %d", arrayLen, maxIterations)
	}

	// Determine error handling strategy
	onError := config.OnError
	if onError == "" {
		onError = ErrorStrategyStop
	}

	// Execute loop iterations
	result := &LoopResult{
		IterationCount: arrayLen,
		Iterations:     make([]IterationResult, 0, arrayLen),
		Metadata: map[string]interface{}{
			"item_variable":  config.ItemVariable,
			"index_variable": config.IndexVariable,
			"on_error":       onError,
		},
	}

	for i, item := range sourceArray {
		iterationResult, err := le.executeIteration(
			ctx,
			i,
			item,
			config,
			execCtx,
			bodyNodes,
			bodyEdges,
		)

		if err != nil {
			if onError == ErrorStrategyStop {
				// Stop on first error
				return nil, fmt.Errorf("loop iteration %d failed: %w", i, err)
			}
			// Continue on error - record error but continue
			errMsg := err.Error()
			iterationResult.Error = &errMsg
		}

		result.Iterations = append(result.Iterations, *iterationResult)
	}

	return result, nil
}

// validateConfig validates loop configuration
func (le *loopExecutor) validateConfig(config workflow.LoopActionConfig) error {
	if config.Source == "" {
		return fmt.Errorf("source is required")
	}
	if config.ItemVariable == "" {
		return fmt.Errorf("item_variable is required")
	}
	if config.OnError != "" && config.OnError != ErrorStrategyContinue && config.OnError != ErrorStrategyStop {
		return fmt.Errorf("on_error must be 'continue' or 'stop', got '%s'", config.OnError)
	}
	return nil
}

// resolveSourceArray resolves the source expression to an array
func (le *loopExecutor) resolveSourceArray(source string, execCtx *ExecutionContext) ([]interface{}, error) {
	// Build context for path resolution
	contextData := buildInterpolationContext(execCtx)

	// Remove ${} wrapper if present
	path := source
	if strings.HasPrefix(path, "${") && strings.HasSuffix(path, "}") {
		path = path[2 : len(path)-1]
	}

	// Resolve the path
	value, err := actions.GetValueByPath(contextData, path)
	if err != nil {
		return nil, fmt.Errorf("source path not found: %w", err)
	}

	// Ensure value is an array
	array, ok := value.([]interface{})
	if !ok {
		return nil, fmt.Errorf("source is not an array, got %T", value)
	}

	return array, nil
}

// executeIteration executes a single loop iteration
func (le *loopExecutor) executeIteration(
	ctx context.Context,
	index int,
	item interface{},
	config workflow.LoopActionConfig,
	execCtx *ExecutionContext,
	bodyNodes []workflow.Node,
	bodyEdges []workflow.Edge,
) (*IterationResult, error) {
	// Create iteration-specific execution context
	iterationCtx := le.createIterationContext(index, item, config, execCtx)

	// Execute body nodes for this iteration
	outputs, err := le.executeBodyNodes(ctx, bodyNodes, bodyEdges, iterationCtx)

	result := &IterationResult{
		Index:  index,
		Item:   item,
		Output: outputs,
	}

	if err != nil {
		return result, err
	}

	return result, nil
}

// createIterationContext creates an execution context for a single iteration
func (le *loopExecutor) createIterationContext(
	index int,
	item interface{},
	config workflow.LoopActionConfig,
	parentCtx *ExecutionContext,
) *ExecutionContext {
	// Create a copy of step outputs
	stepOutputs := make(map[string]interface{})
	for k, v := range parentCtx.StepOutputs {
		stepOutputs[k] = v
	}

	// Add loop variables to the context
	stepOutputs[config.ItemVariable] = item
	if config.IndexVariable != "" {
		stepOutputs[config.IndexVariable] = index
	}

	return &ExecutionContext{
		TenantID:    parentCtx.TenantID,
		ExecutionID: parentCtx.ExecutionID,
		WorkflowID:  parentCtx.WorkflowID,
		TriggerData: parentCtx.TriggerData,
		StepOutputs: stepOutputs,
	}
}

// executeBodyNodes executes all nodes in the loop body
func (le *loopExecutor) executeBodyNodes(
	ctx context.Context,
	nodes []workflow.Node,
	edges []workflow.Edge,
	iterationCtx *ExecutionContext,
) (map[string]interface{}, error) {
	// If no nodes in body, return empty output
	if len(nodes) == 0 {
		return make(map[string]interface{}), nil
	}

	// Build execution order using topological sort
	executionOrder, err := topologicalSort(nodes, edges)
	if err != nil {
		return nil, fmt.Errorf("failed to determine execution order for loop body: %w", err)
	}

	// Build node map for quick lookup
	nodeMap := buildNodeMap(nodes)

	// Execute nodes in order
	outputs := make(map[string]interface{})
	for _, nodeID := range executionOrder {
		node, exists := nodeMap[nodeID]
		if !exists {
			continue
		}

		// Execute the node
		var output interface{}
		var execErr error

		// Use the main executor if available, otherwise execute directly
		if le.mainExecutor != nil {
			output, execErr = le.mainExecutor.executeNode(ctx, node, iterationCtx)
		} else {
			// For testing without full executor, skip actual execution
			output = map[string]interface{}{"status": "executed"}
			execErr = nil
		}

		if execErr != nil {
			return outputs, fmt.Errorf("node %s failed: %w", nodeID, execErr)
		}

		// Store output for downstream nodes
		iterationCtx.StepOutputs[nodeID] = output
		outputs[nodeID] = output
	}

	return outputs, nil
}

// executeLoopAction is the main entry point for loop execution
func (e *Executor) executeLoopAction(
	ctx context.Context,
	node workflow.Node,
	execCtx *ExecutionContext,
	definition *workflow.WorkflowDefinition,
) (interface{}, error) {
	// Parse loop configuration
	var config workflow.LoopActionConfig
	if err := parseNodeConfig(node, &config); err != nil {
		return nil, fmt.Errorf("failed to parse loop configuration: %w", err)
	}

	// Find loop body nodes (nodes connected after this loop node)
	bodyNodes, bodyEdges := e.findLoopBody(node.ID, definition)

	// Create loop executor with reference to main executor
	loopExec := &loopExecutor{
		mainExecutor: e,
	}

	// Execute the loop
	result, err := loopExec.executeLoop(ctx, config, execCtx, bodyNodes, bodyEdges)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// findLoopBody finds all nodes that are part of the loop body
// For now, this returns all direct children of the loop node
// TODO: Implement proper loop body detection (nodes between loop start and loop end markers)
func (e *Executor) findLoopBody(loopNodeID string, definition *workflow.WorkflowDefinition) ([]workflow.Node, []workflow.Edge) {
	// Find all nodes directly connected to this loop node
	var bodyNodeIDs []string
	for _, edge := range definition.Edges {
		if edge.Source == loopNodeID {
			bodyNodeIDs = append(bodyNodeIDs, edge.Target)
		}
	}

	// Collect body nodes
	var bodyNodes []workflow.Node
	bodyNodeMap := make(map[string]bool)
	for _, nodeID := range bodyNodeIDs {
		bodyNodeMap[nodeID] = true
	}

	for _, node := range definition.Nodes {
		if bodyNodeMap[node.ID] {
			bodyNodes = append(bodyNodes, node)
		}
	}

	// Collect edges between body nodes
	var bodyEdges []workflow.Edge
	for _, edge := range definition.Edges {
		if bodyNodeMap[edge.Source] && bodyNodeMap[edge.Target] {
			bodyEdges = append(bodyEdges, edge)
		}
	}

	return bodyNodes, bodyEdges
}

// parseNodeConfig parses node configuration into a target struct
func parseNodeConfig(node workflow.Node, target interface{}) error {
	configData := node.Data.Config
	if len(configData) == 0 {
		return fmt.Errorf("missing configuration")
	}
	return json.Unmarshal(configData, target)
}
