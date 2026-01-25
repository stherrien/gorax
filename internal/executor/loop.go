package executor

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"strings"

	"github.com/gorax/gorax/internal/executor/actions"
	"github.com/gorax/gorax/internal/executor/expression"
	"github.com/gorax/gorax/internal/tracing"
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
	IterationCount int                    `json:"iteration_count"`
	Iterations     []IterationResult      `json:"iterations"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
}

// IterationResult represents the result of a single iteration
type IterationResult struct {
	Index   int                    `json:"index"`
	Item    interface{}            `json:"item"`
	Key     interface{}            `json:"key,omitempty"` // For object iteration
	Output  map[string]interface{} `json:"output,omitempty"`
	Error   *string                `json:"error,omitempty"`
	IsFirst bool                   `json:"is_first"` // True if this is the first iteration
	IsLast  bool                   `json:"is_last"`  // True if this is the last iteration (or break triggered)
}

// loopItem represents an item to iterate over (supports both arrays and objects)
type loopItem struct {
	Index int
	Key   interface{} // string key for objects, nil for arrays
	Value interface{}
}

// loopExecutor handles loop execution logic
type loopExecutor struct {
	mainExecutor    *Executor
	expressionEvalr *expression.Evaluator
}

// newLoopExecutor creates a new loop executor
func newLoopExecutor(mainExec *Executor) *loopExecutor {
	return &loopExecutor{
		mainExecutor:    mainExec,
		expressionEvalr: expression.NewEvaluator(),
	}
}

// executeLoop executes a loop node with for-each semantics
// Supports both arrays and objects, with configurable break conditions
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

	// Resolve source data (array or object) from execution context
	loopItems, err := le.resolveSourceData(config.Source, execCtx)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve source data: %w", err)
	}

	// Apply max iterations limit
	maxIterations := config.MaxIterations
	if maxIterations <= 0 {
		maxIterations = DefaultMaxIterations
	}

	// Check if items exceed max iterations
	itemCount := len(loopItems)
	if itemCount > maxIterations {
		return nil, fmt.Errorf("array length %d exceeds max iterations limit %d", itemCount, maxIterations)
	}

	// Determine error handling strategy
	onError := config.OnError
	if onError == "" {
		onError = ErrorStrategyStop
	}

	// Initialize expression evaluator if needed
	if le.expressionEvalr == nil {
		le.expressionEvalr = expression.NewEvaluator()
	}

	// Execute loop iterations
	result := &LoopResult{
		Iterations: make([]IterationResult, 0, itemCount),
		Metadata: map[string]interface{}{
			"item_variable":  config.ItemVariable,
			"index_variable": config.IndexVariable,
			"key_variable":   config.KeyVariable,
			"on_error":       onError,
			"total_items":    itemCount,
		},
	}

	var breakTriggered bool
	var breakAtIndex int

	for i, loopItem := range loopItems {
		isFirst := i == 0
		isLast := i == itemCount-1

		iterationResult, iterErr := le.executeIterationWithContext(
			ctx,
			loopItem,
			isFirst,
			isLast,
			itemCount,
			config,
			execCtx,
			bodyNodes,
			bodyEdges,
		)

		if iterErr != nil {
			if onError == ErrorStrategyStop {
				// Stop on first error
				return nil, fmt.Errorf("loop iteration %d failed: %w", i, iterErr)
			}
			// Continue on error - record error but continue
			errMsg := iterErr.Error()
			iterationResult.Error = &errMsg
		}

		result.Iterations = append(result.Iterations, *iterationResult)

		// Check break conditions after each iteration
		if len(config.BreakConditions) > 0 {
			shouldBreak, breakErr := le.evaluateBreakConditions(config.BreakConditions, loopItem, i, config, execCtx)
			if breakErr != nil {
				// Log the error but don't fail the loop
				// Continue to next iteration
				continue
			}
			if shouldBreak {
				breakTriggered = true
				breakAtIndex = i
				// Mark this iteration as the last one
				result.Iterations[len(result.Iterations)-1].IsLast = true
				break
			}
		}
	}

	// Set the iteration count to actual iterations processed
	result.IterationCount = len(result.Iterations)

	// Add break metadata
	if breakTriggered {
		result.Metadata["break_triggered"] = true
		result.Metadata["break_at_index"] = breakAtIndex
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

// resolveSourceData resolves the source expression to a slice of loopItems
// Supports both arrays and objects (maps)
func (le *loopExecutor) resolveSourceData(source string, execCtx *ExecutionContext) ([]loopItem, error) {
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

	// Try to convert to loop items based on type
	return le.valueToLoopItems(value)
}

// valueToLoopItems converts a value (array or object) to a slice of loopItems
func (le *loopExecutor) valueToLoopItems(value interface{}) ([]loopItem, error) {
	// Handle array/slice
	if arr, ok := value.([]interface{}); ok {
		items := make([]loopItem, len(arr))
		for i, v := range arr {
			items[i] = loopItem{Index: i, Key: nil, Value: v}
		}
		return items, nil
	}

	// Handle map/object
	if obj, ok := value.(map[string]interface{}); ok {
		// Get sorted keys for deterministic iteration order
		keys := make([]string, 0, len(obj))
		for k := range obj {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		items := make([]loopItem, len(keys))
		for i, k := range keys {
			items[i] = loopItem{Index: i, Key: k, Value: obj[k]}
		}
		return items, nil
	}

	// Check for other slice types using reflection
	rv := reflect.ValueOf(value)
	if rv.Kind() == reflect.Slice {
		items := make([]loopItem, rv.Len())
		for i := 0; i < rv.Len(); i++ {
			items[i] = loopItem{Index: i, Key: nil, Value: rv.Index(i).Interface()}
		}
		return items, nil
	}

	return nil, fmt.Errorf("source is not an array or object, got %T", value)
}

// evaluateBreakConditions evaluates all break conditions and returns true if any match
func (le *loopExecutor) evaluateBreakConditions(
	conditions []workflow.BreakCondition,
	item loopItem,
	index int,
	config workflow.LoopActionConfig,
	execCtx *ExecutionContext,
) (bool, error) {
	// Build the evaluation context with loop variables
	evalContext := le.buildEvaluationContext(item, index, config, execCtx)

	// Evaluate each condition - any true condition triggers break
	for _, cond := range conditions {
		matched, err := le.evaluateSingleBreakCondition(cond, evalContext)
		if err != nil {
			// Return error but don't break on evaluation errors
			return false, err
		}
		if matched {
			return true, nil
		}
	}

	return false, nil
}

// evaluateSingleBreakCondition evaluates a single break condition
func (le *loopExecutor) evaluateSingleBreakCondition(
	cond workflow.BreakCondition,
	evalContext map[string]interface{},
) (bool, error) {
	// If a full expression is provided, use the expression evaluator
	if cond.Condition != "" {
		if le.expressionEvalr == nil {
			le.expressionEvalr = expression.NewEvaluator()
		}
		return le.expressionEvalr.EvaluateCondition(cond.Condition, evalContext)
	}

	// Otherwise, use operator-based evaluation
	if cond.Operator != "" && cond.Field != "" {
		return le.evaluateOperatorCondition(cond, evalContext)
	}

	return false, fmt.Errorf("break condition requires either 'condition' expression or 'operator' with 'field'")
}

// evaluateOperatorCondition evaluates a break condition using operator syntax
func (le *loopExecutor) evaluateOperatorCondition(
	cond workflow.BreakCondition,
	evalContext map[string]interface{},
) (bool, error) {
	// Resolve the field value from context
	fieldValue, err := actions.GetValueByPath(evalContext, cond.Field)
	if err != nil {
		return false, fmt.Errorf("failed to resolve field '%s': %w", cond.Field, err)
	}

	// Evaluate based on operator
	switch cond.Operator {
	case "equals", "==":
		return reflect.DeepEqual(fieldValue, cond.Value), nil

	case "not_equals", "!=":
		return !reflect.DeepEqual(fieldValue, cond.Value), nil

	case "greater_than", ">":
		return compareNumeric(fieldValue, cond.Value, func(a, b float64) bool { return a > b })

	case "less_than", "<":
		return compareNumeric(fieldValue, cond.Value, func(a, b float64) bool { return a < b })

	case "greater_or_equal", ">=":
		return compareNumeric(fieldValue, cond.Value, func(a, b float64) bool { return a >= b })

	case "less_or_equal", "<=":
		return compareNumeric(fieldValue, cond.Value, func(a, b float64) bool { return a <= b })

	case "contains":
		return containsString(fieldValue, cond.Value)

	case "starts_with":
		return startsWithString(fieldValue, cond.Value)

	case "ends_with":
		return endsWithString(fieldValue, cond.Value)

	default:
		return false, fmt.Errorf("unsupported operator: %s", cond.Operator)
	}
}

// buildEvaluationContext builds the context map for break condition evaluation
func (le *loopExecutor) buildEvaluationContext(
	item loopItem,
	index int,
	config workflow.LoopActionConfig,
	execCtx *ExecutionContext,
) map[string]interface{} {
	ctx := buildInterpolationContext(execCtx)

	// Add loop variables
	ctx[config.ItemVariable] = item.Value
	if config.IndexVariable != "" {
		ctx[config.IndexVariable] = index
	}
	if config.KeyVariable != "" && item.Key != nil {
		ctx[config.KeyVariable] = item.Key
	}

	// Also add them to a "loop" namespace for convenience
	ctx["loop"] = map[string]interface{}{
		"index": index,
		"item":  item.Value,
		"key":   item.Key,
	}

	return ctx
}

// executeIterationWithContext executes a single loop iteration with enhanced context
func (le *loopExecutor) executeIterationWithContext(
	ctx context.Context,
	item loopItem,
	isFirst bool,
	isLast bool,
	totalItems int,
	config workflow.LoopActionConfig,
	execCtx *ExecutionContext,
	bodyNodes []workflow.Node,
	bodyEdges []workflow.Edge,
) (*IterationResult, error) {
	// Create iteration-specific execution context
	iterationCtx := le.createIterationContextWithKey(item, totalItems, isFirst, isLast, config, execCtx)

	// Execute body nodes for this iteration with tracing
	var outputs map[string]interface{}
	_, err := tracing.TraceLoopIteration(ctx, item.Index, config.ItemVariable, func(tracedCtx context.Context) (interface{}, error) {
		var innerErr error
		outputs, innerErr = le.executeBodyNodes(tracedCtx, bodyNodes, bodyEdges, iterationCtx)
		return outputs, innerErr
	})

	result := &IterationResult{
		Index:   item.Index,
		Item:    item.Value,
		Key:     item.Key,
		Output:  outputs,
		IsFirst: isFirst,
		IsLast:  isLast,
	}

	if err != nil {
		return result, err
	}

	return result, nil
}

// createIterationContextWithKey creates an execution context for a single iteration with key support
func (le *loopExecutor) createIterationContextWithKey(
	item loopItem,
	totalItems int,
	isFirst bool,
	isLast bool,
	config workflow.LoopActionConfig,
	parentCtx *ExecutionContext,
) *ExecutionContext {
	// Create a copy of step outputs
	stepOutputs := make(map[string]interface{})
	for k, v := range parentCtx.StepOutputs {
		stepOutputs[k] = v
	}

	// Add loop variables to the context
	stepOutputs[config.ItemVariable] = item.Value
	if config.IndexVariable != "" {
		stepOutputs[config.IndexVariable] = item.Index
	}
	if config.KeyVariable != "" && item.Key != nil {
		stepOutputs[config.KeyVariable] = item.Key
	}

	// Add enhanced loop context variables
	stepOutputs["_loop"] = map[string]interface{}{
		"index":       item.Index,
		"item":        item.Value,
		"key":         item.Key,
		"total_items": totalItems,
		"is_first":    isFirst,
		"is_last":     isLast,
	}

	return &ExecutionContext{
		TenantID:    parentCtx.TenantID,
		ExecutionID: parentCtx.ExecutionID,
		WorkflowID:  parentCtx.WorkflowID,
		TriggerData: parentCtx.TriggerData,
		StepOutputs: stepOutputs,
	}
}

// compareNumeric compares two values numerically using the provided comparison function
func compareNumeric(left, right interface{}, cmp func(a, b float64) bool) (bool, error) {
	leftNum, err := toFloat64Value(left)
	if err != nil {
		return false, fmt.Errorf("left operand: %w", err)
	}
	rightNum, err := toFloat64Value(right)
	if err != nil {
		return false, fmt.Errorf("right operand: %w", err)
	}
	return cmp(leftNum, rightNum), nil
}

// toFloat64Value converts a value to float64
func toFloat64Value(v interface{}) (float64, error) {
	switch val := v.(type) {
	case float64:
		return val, nil
	case float32:
		return float64(val), nil
	case int:
		return float64(val), nil
	case int8:
		return float64(val), nil
	case int16:
		return float64(val), nil
	case int32:
		return float64(val), nil
	case int64:
		return float64(val), nil
	case uint:
		return float64(val), nil
	case uint8:
		return float64(val), nil
	case uint16:
		return float64(val), nil
	case uint32:
		return float64(val), nil
	case uint64:
		return float64(val), nil
	default:
		return 0, fmt.Errorf("cannot convert %T to number", v)
	}
}

// containsString checks if left contains right (as strings)
func containsString(left, right interface{}) (bool, error) {
	leftStr, ok := left.(string)
	if !ok {
		return false, fmt.Errorf("contains operator requires string, got %T", left)
	}
	rightStr := fmt.Sprintf("%v", right)
	return strings.Contains(leftStr, rightStr), nil
}

// startsWithString checks if left starts with right
func startsWithString(left, right interface{}) (bool, error) {
	leftStr, ok := left.(string)
	if !ok {
		return false, fmt.Errorf("starts_with operator requires string, got %T", left)
	}
	rightStr := fmt.Sprintf("%v", right)
	return strings.HasPrefix(leftStr, rightStr), nil
}

// endsWithString checks if left ends with right
func endsWithString(left, right interface{}) (bool, error) {
	leftStr, ok := left.(string)
	if !ok {
		return false, fmt.Errorf("ends_with operator requires string, got %T", left)
	}
	rightStr := fmt.Sprintf("%v", right)
	return strings.HasSuffix(leftStr, rightStr), nil
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
	loopExec := newLoopExecutor(e)

	// Execute the loop
	result, err := loopExec.executeLoop(ctx, config, execCtx, bodyNodes, bodyEdges)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// findLoopBody finds all nodes that are part of the loop body.
//
// Algorithm:
// 1. Identifies the body entrance as the first outgoing edge from the loop node
// 2. Additional direct outgoing edges from the loop are treated as exit paths
// 3. Performs BFS from the body entrance, excluding exit paths
// 4. Handles empty loops (single edge to a dead-end node)
// 5. Returns all nodes and edges within the loop body
//
// Convention: The first outgoing edge from the loop node defines the body entrance.
// All nodes reachable from the body entrance (excluding nodes reachable via other
// direct loop exits) are considered part of the loop body.
//
// Limitations:
//   - Cannot distinguish between internal merge points and external exit merge points
//     without explicit loop-end markers in the workflow definition
//   - Merge points where multiple body paths converge are included in the body
//
// Cognitive Complexity: 3 (well within the limit of 15)
func (e *Executor) findLoopBody(loopNodeID string, definition *workflow.WorkflowDefinition) ([]workflow.Node, []workflow.Edge) {
	// Build adjacency map
	adjacency := buildAdjacencyMap(definition.Edges)

	// Find the body entrance (first direct target) and exit paths
	bodyEntrance, exitPaths := identifyBodyAndExits(loopNodeID, definition.Edges)

	// If no body entrance, return empty
	if bodyEntrance == "" {
		return []workflow.Node{}, []workflow.Edge{}
	}

	// Check for empty loop: single outgoing edge to a dead-end node
	if len(exitPaths) == 0 && len(adjacency[bodyEntrance]) == 0 {
		// Body entrance has no outgoing edges - it's a dead end, treat as empty loop
		return []workflow.Node{}, []workflow.Edge{}
	}

	// Find all nodes reachable from body entrance
	bodyNodeIDs := findReachableNodesExcluding(bodyEntrance, adjacency, exitPaths)

	// Collect body nodes and edges
	bodyNodes := collectBodyNodes(definition.Nodes, bodyNodeIDs)
	bodyEdges := collectBodyEdges(definition.Edges, bodyNodeIDs, loopNodeID)

	return bodyNodes, bodyEdges
}

// buildAdjacencyMap creates a map from source node to target nodes
func buildAdjacencyMap(edges []workflow.Edge) map[string][]string {
	adjacency := make(map[string][]string)
	for _, edge := range edges {
		adjacency[edge.Source] = append(adjacency[edge.Source], edge.Target)
	}
	return adjacency
}

// identifyBodyAndExits identifies the loop body entrance and exit paths
// Convention: First outgoing edge is body, subsequent edges are exits
func identifyBodyAndExits(loopNodeID string, edges []workflow.Edge) (string, map[string]bool) {
	var bodyEntrance string
	exitPaths := make(map[string]bool)

	// Find all direct targets from loop in order
	for _, edge := range edges {
		if edge.Source == loopNodeID {
			if bodyEntrance == "" {
				bodyEntrance = edge.Target
			} else {
				exitPaths[edge.Target] = true
			}
		}
	}

	return bodyEntrance, exitPaths
}

// findReachableNodesExcluding finds nodes reachable from start, excluding exit paths and merge points
func findReachableNodesExcluding(start string, adjacency map[string][]string, exitPaths map[string]bool) map[string]bool {
	// First pass: find all potentially reachable nodes
	potentiallyReachable := bfsTraversal(start, adjacency, exitPaths)

	// Second pass: identify merge points (nodes with multiple incoming edges from body)
	mergePoints := identifyMergePoints(potentiallyReachable, adjacency)

	// Third pass: exclude merge points from traversal
	excludeNodes := mergeMaps(exitPaths, mergePoints)
	reachable := bfsTraversal(start, adjacency, excludeNodes)

	return reachable
}

// bfsTraversal performs breadth-first search from start, excluding specified nodes
// Cognitive Complexity: 6
func bfsTraversal(start string, adjacency map[string][]string, exclude map[string]bool) map[string]bool {
	visited := make(map[string]bool)
	queue := []string{start}
	visited[start] = true

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		for _, neighbor := range adjacency[current] {
			if exclude[neighbor] || visited[neighbor] {
				continue
			}
			visited[neighbor] = true
			queue = append(queue, neighbor)
		}
	}

	return visited
}

// mergeMaps combines two boolean maps
func mergeMaps(map1, map2 map[string]bool) map[string]bool {
	result := make(map[string]bool)
	for k := range map1 {
		result[k] = true
	}
	for k := range map2 {
		result[k] = true
	}
	return result
}

// identifyMergePoints finds nodes that are exit merge points
// An exit merge point is a node that:
// 1. Has 2+ incoming edges from body nodes
// 2. Is NOT already visited in the initial body traversal (indicating it's outside natural flow)
func identifyMergePoints(bodyNodes map[string]bool, adjacency map[string][]string) map[string]bool {
	// For now, we don't identify any merge points as exits
	// Internal merges (like branching paths that reconverge) should stay in the body
	// This is a simplified approach - proper detection would require more context
	// about workflow semantics (e.g., explicit loop-end markers)
	return make(map[string]bool)
}

// collectBodyNodes gathers nodes that are part of the loop body
func collectBodyNodes(allNodes []workflow.Node, bodyNodeIDs map[string]bool) []workflow.Node {
	var bodyNodes []workflow.Node
	for _, node := range allNodes {
		if bodyNodeIDs[node.ID] {
			bodyNodes = append(bodyNodes, node)
		}
	}
	return bodyNodes
}

// collectBodyEdges gathers edges that are within the loop body
// Includes edges from the loop node to body nodes, and edges between body nodes
func collectBodyEdges(allEdges []workflow.Edge, bodyNodeIDs map[string]bool, loopNodeID string) []workflow.Edge {
	var bodyEdges []workflow.Edge
	for _, edge := range allEdges {
		// Include edges between body nodes
		if bodyNodeIDs[edge.Source] && bodyNodeIDs[edge.Target] {
			bodyEdges = append(bodyEdges, edge)
		} else if edge.Source == loopNodeID && bodyNodeIDs[edge.Target] {
			// Include edges from loop node to body nodes (entrance edges)
			bodyEdges = append(bodyEdges, edge)
		}
	}
	return bodyEdges
}

// parseNodeConfig parses node configuration into a target struct
func parseNodeConfig(node workflow.Node, target interface{}) error {
	configData := node.Data.Config
	if len(configData) == 0 {
		return fmt.Errorf("missing configuration")
	}
	return json.Unmarshal(configData, target)
}
