package executor

import (
	"fmt"
)

// MaxSubworkflowDepth defines the maximum nesting depth for subworkflows
const MaxSubworkflowDepth = 10

// SubworkflowTracker tracks subworkflow execution to prevent circular dependencies
// and enforce depth limits
type SubworkflowTracker struct {
	parentExecID     string
	childExecID      string
	depth            int
	maxDepth         int
	visitedWorkflows map[string]bool
	workflowChain    []string
}

// NewSubworkflowTracker creates a new subworkflow tracker
func NewSubworkflowTracker(parentExecID string, depth int) *SubworkflowTracker {
	return &SubworkflowTracker{
		parentExecID:     parentExecID,
		depth:            depth,
		maxDepth:         MaxSubworkflowDepth,
		visitedWorkflows: make(map[string]bool),
		workflowChain:    make([]string, 0),
	}
}

// WithWorkflowChain sets the workflow chain for tracking
func (t *SubworkflowTracker) WithWorkflowChain(chain []string) *SubworkflowTracker {
	t.workflowChain = append([]string{}, chain...) // Copy to avoid mutation
	for _, wfID := range chain {
		t.visitedWorkflows[wfID] = true
	}
	return t
}

// WithMaxDepth sets a custom maximum depth
func (t *SubworkflowTracker) WithMaxDepth(maxDepth int) *SubworkflowTracker {
	t.maxDepth = maxDepth
	return t
}

// CanExecute checks if a workflow can be executed without violating constraints
func (t *SubworkflowTracker) CanExecute(workflowID string) error {
	// Check depth limit
	if t.depth >= t.maxDepth {
		return fmt.Errorf("maximum subworkflow depth exceeded: current depth %d, max depth %d",
			t.depth, t.maxDepth)
	}

	// Check for circular dependency
	if t.visitedWorkflows[workflowID] {
		return fmt.Errorf("circular workflow dependency detected: workflow %s already in execution chain %v",
			workflowID, t.workflowChain)
	}

	return nil
}

// AddToChain adds a workflow to the execution chain
func (t *SubworkflowTracker) AddToChain(workflowID string) {
	t.visitedWorkflows[workflowID] = true
	t.workflowChain = append(t.workflowChain, workflowID)
}

// GetChain returns the current workflow chain
func (t *SubworkflowTracker) GetChain() []string {
	return append([]string{}, t.workflowChain...) // Return copy
}

// GetDepth returns the current execution depth
func (t *SubworkflowTracker) GetDepth() int {
	return t.depth
}

// NextDepth returns the depth for the next subworkflow
func (t *SubworkflowTracker) NextDepth() int {
	return t.depth + 1
}
