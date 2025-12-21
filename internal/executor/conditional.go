package executor

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/gorax/gorax/internal/executor/expression"
	"github.com/gorax/gorax/internal/workflow"
)

// ConditionalBranchResult represents the result of evaluating a conditional node
type ConditionalBranchResult struct {
	Condition     string
	Result        bool
	TakenBranch   string // "true" or "false"
	NextNodes     []string
	StopExecution bool
}

// executeConditionalAction executes a conditional (if/else) action
func (e *Executor) executeConditionalAction(ctx context.Context, node workflow.Node, execCtx *ExecutionContext, definition *workflow.WorkflowDefinition) (*ConditionalBranchResult, error) {
	// Extract config from node data
	configData := node.Data.Config
	if len(configData) == 0 {
		return nil, fmt.Errorf("missing config for conditional action")
	}

	// Parse node config
	var config workflow.ConditionalActionConfig
	if err := json.Unmarshal(configData, &config); err != nil {
		return nil, fmt.Errorf("failed to parse conditional action config: %w", err)
	}

	if config.Condition == "" {
		return nil, fmt.Errorf("condition expression is required")
	}

	// Build evaluation context
	evalContext := expression.BuildContext(
		execCtx.TriggerData,
		execCtx.StepOutputs,
		map[string]interface{}{
			"tenant_id":    execCtx.TenantID,
			"execution_id": execCtx.ExecutionID,
			"workflow_id":  execCtx.WorkflowID,
		},
	)

	// Evaluate the condition
	evaluator := expression.NewEvaluator()
	result, err := evaluator.EvaluateCondition(config.Condition, evalContext)
	if err != nil {
		return nil, fmt.Errorf("failed to evaluate condition: %w", err)
	}

	e.logger.Info("conditional expression evaluated",
		"node_id", node.ID,
		"condition", config.Condition,
		"result", result,
	)

	// Determine which branch to take
	branchResult := &ConditionalBranchResult{
		Condition:   config.Condition,
		Result:      result,
		TakenBranch: "false",
		NextNodes:   []string{},
	}

	if result {
		branchResult.TakenBranch = "true"
		branchResult.StopExecution = config.StopOnTrue

		// Find nodes connected via "true" branch
		branchResult.NextNodes = e.findConditionalBranch(node.ID, "true", definition.Edges)
	} else {
		branchResult.TakenBranch = "false"
		branchResult.StopExecution = config.StopOnFalse

		// Find nodes connected via "false" branch
		branchResult.NextNodes = e.findConditionalBranch(node.ID, "false", definition.Edges)
	}

	e.logger.Info("conditional branch determined",
		"node_id", node.ID,
		"taken_branch", branchResult.TakenBranch,
		"next_nodes", branchResult.NextNodes,
		"stop_execution", branchResult.StopExecution,
	)

	return branchResult, nil
}

// findConditionalBranch finds the target nodes for a specific branch label
func (e *Executor) findConditionalBranch(sourceNodeID string, branchLabel string, edges []workflow.Edge) []string {
	var targets []string

	for _, edge := range edges {
		if edge.Source == sourceNodeID && edge.Label == branchLabel {
			targets = append(targets, edge.Target)
		}
	}

	return targets
}

// buildConditionalExecutionPlan builds an execution plan that respects conditional branches
// Instead of simple topological sort, this builds a plan that can handle branching
func (e *Executor) buildConditionalExecutionPlan(nodes []workflow.Node, edges []workflow.Edge) (*ExecutionPlan, error) {
	plan := &ExecutionPlan{
		AllNodes:         buildNodeMap(nodes),
		Edges:            edges,
		ConditionalNodes: make(map[string]bool),
	}

	// Identify conditional nodes
	for _, node := range nodes {
		if node.Type == string(workflow.NodeTypeControlIf) {
			plan.ConditionalNodes[node.ID] = true
		}
	}

	// Find start nodes (nodes with no incoming edges or trigger nodes)
	plan.StartNodes = e.findStartNodes(nodes, edges)

	if len(plan.StartNodes) == 0 {
		return nil, fmt.Errorf("no start nodes found in workflow")
	}

	return plan, nil
}

// findStartNodes finds all nodes that can start execution
func (e *Executor) findStartNodes(nodes []workflow.Node, edges []workflow.Edge) []string {
	inDegree := make(map[string]int)

	// Initialize in-degree for all nodes
	for _, node := range nodes {
		inDegree[node.ID] = 0
	}

	// Count incoming edges
	for _, edge := range edges {
		inDegree[edge.Target]++
	}

	// Find nodes with no incoming edges
	var startNodes []string
	for _, node := range nodes {
		if inDegree[node.ID] == 0 {
			startNodes = append(startNodes, node.ID)
		}
	}

	return startNodes
}

// ExecutionPlan represents a conditional execution plan
type ExecutionPlan struct {
	AllNodes         map[string]workflow.Node
	Edges            []workflow.Edge
	StartNodes       []string
	ConditionalNodes map[string]bool
}

// getNextNodes returns the next nodes to execute based on current node
// For conditional nodes, this will be determined at runtime
// For regular nodes, this returns all connected nodes
func (plan *ExecutionPlan) getNextNodes(nodeID string, conditionalResult *ConditionalBranchResult) []string {
	// If this was a conditional node and we have a result, use it
	if conditionalResult != nil && plan.ConditionalNodes[nodeID] {
		return conditionalResult.NextNodes
	}

	// For non-conditional nodes, return all targets
	var targets []string
	for _, edge := range plan.Edges {
		if edge.Source == nodeID {
			targets = append(targets, edge.Target)
		}
	}

	return targets
}

// hasAllDependenciesCompleted checks if all dependencies of a node are completed
func (plan *ExecutionPlan) hasAllDependenciesCompleted(nodeID string, completedNodes map[string]bool, skippedNodes map[string]bool) bool {
	// Find all nodes that have edges pointing to this node
	for _, edge := range plan.Edges {
		if edge.Target == nodeID {
			sourceNode := edge.Source
			// Dependency must be either completed or skipped
			if !completedNodes[sourceNode] && !skippedNodes[sourceNode] {
				return false
			}
		}
	}
	return true
}

// findNodesToSkip finds all nodes that should be skipped based on conditional branching
// This performs a traversal from the conditional node along the non-taken branch
func (e *Executor) findNodesToSkip(conditionalNodeID string, takenBranch string, edges []workflow.Edge, allNodes map[string]workflow.Node) map[string]bool {
	skipped := make(map[string]bool)

	// Find the branch that was NOT taken
	notTakenBranch := "false"
	if takenBranch == "false" {
		notTakenBranch = "true"
	}

	// Find all nodes reachable from the non-taken branch
	// These should be marked as skipped
	notTakenTargets := []string{}
	for _, edge := range edges {
		if edge.Source == conditionalNodeID && edge.Label == notTakenBranch {
			notTakenTargets = append(notTakenTargets, edge.Target)
		}
	}

	// BFS to find all nodes reachable from non-taken branch
	visited := make(map[string]bool)
	queue := notTakenTargets

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		if visited[current] {
			continue
		}

		visited[current] = true
		skipped[current] = true

		// Add all children of this node to the queue
		// But only if they are exclusively reachable from this branch
		for _, edge := range edges {
			if edge.Source == current {
				// Check if this target has other incoming edges from nodes not in skipped set
				hasOtherParents := false
				for _, checkEdge := range edges {
					if checkEdge.Target == edge.Target && checkEdge.Source != current && !skipped[checkEdge.Source] {
						hasOtherParents = true
						break
					}
				}

				// Only skip if this node has no other non-skipped parents
				if !hasOtherParents {
					queue = append(queue, edge.Target)
				}
			}
		}
	}

	return skipped
}
