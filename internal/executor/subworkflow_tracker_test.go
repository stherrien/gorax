package executor

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewSubworkflowTracker(t *testing.T) {
	tracker := NewSubworkflowTracker("parent-exec-1", 0)

	assert.NotNil(t, tracker)
	assert.Equal(t, "parent-exec-1", tracker.parentExecID)
	assert.Equal(t, 0, tracker.depth)
	assert.Equal(t, MaxSubworkflowDepth, tracker.maxDepth)
	assert.Empty(t, tracker.visitedWorkflows)
	assert.Empty(t, tracker.workflowChain)
}

func TestSubworkflowTracker_WithWorkflowChain(t *testing.T) {
	tracker := NewSubworkflowTracker("parent-exec-1", 2)
	chain := []string{"wf-root", "wf-child"}

	tracker.WithWorkflowChain(chain)

	assert.Equal(t, 2, len(tracker.workflowChain))
	assert.Equal(t, chain, tracker.workflowChain)
	assert.True(t, tracker.visitedWorkflows["wf-root"])
	assert.True(t, tracker.visitedWorkflows["wf-child"])

	// Verify original chain is not mutated
	chain[0] = "modified"
	assert.Equal(t, "wf-root", tracker.workflowChain[0])
}

func TestSubworkflowTracker_WithMaxDepth(t *testing.T) {
	tracker := NewSubworkflowTracker("parent-exec-1", 0)
	tracker.WithMaxDepth(5)

	assert.Equal(t, 5, tracker.maxDepth)
}

func TestSubworkflowTracker_CanExecute_Success(t *testing.T) {
	tracker := NewSubworkflowTracker("parent-exec-1", 0)

	err := tracker.CanExecute("wf-new")

	assert.NoError(t, err)
}

func TestSubworkflowTracker_CanExecute_MaxDepthExceeded(t *testing.T) {
	tracker := NewSubworkflowTracker("parent-exec-1", 10)

	err := tracker.CanExecute("wf-new")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "maximum subworkflow depth exceeded")
	assert.Contains(t, err.Error(), "current depth 10")
}

func TestSubworkflowTracker_CanExecute_CircularDependency(t *testing.T) {
	tracker := NewSubworkflowTracker("parent-exec-1", 2)
	tracker.WithWorkflowChain([]string{"wf-root", "wf-child"})

	// Try to execute wf-child again (circular)
	err := tracker.CanExecute("wf-child")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "circular workflow dependency detected")
	assert.Contains(t, err.Error(), "wf-child")
}

func TestSubworkflowTracker_CanExecute_CircularDependencyRoot(t *testing.T) {
	tracker := NewSubworkflowTracker("parent-exec-1", 1)
	tracker.WithWorkflowChain([]string{"wf-root"})

	// Try to execute wf-root again (circular back to root)
	err := tracker.CanExecute("wf-root")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "circular workflow dependency detected")
	assert.Contains(t, err.Error(), "wf-root")
}

func TestSubworkflowTracker_AddToChain(t *testing.T) {
	tracker := NewSubworkflowTracker("parent-exec-1", 0)
	tracker.WithWorkflowChain([]string{"wf-root"})

	tracker.AddToChain("wf-child")

	assert.True(t, tracker.visitedWorkflows["wf-child"])
	assert.Equal(t, []string{"wf-root", "wf-child"}, tracker.workflowChain)
}

func TestSubworkflowTracker_GetChain(t *testing.T) {
	tracker := NewSubworkflowTracker("parent-exec-1", 0)
	tracker.WithWorkflowChain([]string{"wf-root", "wf-child"})

	chain := tracker.GetChain()

	assert.Equal(t, []string{"wf-root", "wf-child"}, chain)

	// Verify returned chain is a copy
	chain[0] = "modified"
	assert.Equal(t, "wf-root", tracker.workflowChain[0])
}

func TestSubworkflowTracker_GetDepth(t *testing.T) {
	tracker := NewSubworkflowTracker("parent-exec-1", 3)

	assert.Equal(t, 3, tracker.GetDepth())
}

func TestSubworkflowTracker_NextDepth(t *testing.T) {
	tracker := NewSubworkflowTracker("parent-exec-1", 3)

	assert.Equal(t, 4, tracker.NextDepth())
	// Verify original depth unchanged
	assert.Equal(t, 3, tracker.GetDepth())
}

func TestSubworkflowTracker_ComplexChain(t *testing.T) {
	// Simulate a complex workflow chain: A -> B -> C -> D
	tracker := NewSubworkflowTracker("exec-a", 0)
	tracker.WithWorkflowChain([]string{"wf-a"})

	// Can execute B
	err := tracker.CanExecute("wf-b")
	assert.NoError(t, err)
	tracker.AddToChain("wf-b")

	// Update depth for next level
	tracker.depth = 1

	// Can execute C
	err = tracker.CanExecute("wf-c")
	assert.NoError(t, err)
	tracker.AddToChain("wf-c")

	// Update depth for next level
	tracker.depth = 2

	// Can execute D
	err = tracker.CanExecute("wf-d")
	assert.NoError(t, err)
	tracker.AddToChain("wf-d")

	// Cannot execute A again (circular)
	tracker.depth = 3
	err = tracker.CanExecute("wf-a")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "circular")

	// Cannot execute B again (circular)
	err = tracker.CanExecute("wf-b")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "circular")

	// Verify chain
	assert.Equal(t, []string{"wf-a", "wf-b", "wf-c", "wf-d"}, tracker.GetChain())
}

func TestSubworkflowTracker_CustomMaxDepth(t *testing.T) {
	tracker := NewSubworkflowTracker("parent-exec-1", 4)
	tracker.WithMaxDepth(5)

	// Should succeed at depth 4 with max 5
	err := tracker.CanExecute("wf-new")
	assert.NoError(t, err)

	// Should fail at depth 5 with max 5
	tracker.depth = 5
	err = tracker.CanExecute("wf-new")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "maximum subworkflow depth exceeded")
}

func TestSubworkflowTracker_EdgeCaseEmptyWorkflowID(t *testing.T) {
	tracker := NewSubworkflowTracker("parent-exec-1", 0)

	// Empty workflow ID should be allowed (treated as unique)
	err := tracker.CanExecute("")
	assert.NoError(t, err)

	tracker.AddToChain("")

	// But adding it again should detect circular
	err = tracker.CanExecute("")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "circular")
}

func TestSubworkflowTracker_ThreadSafety(t *testing.T) {
	// This test verifies that getting a copy of the chain doesn't cause issues
	tracker := NewSubworkflowTracker("parent-exec-1", 0)
	tracker.WithWorkflowChain([]string{"wf-1"})

	chain1 := tracker.GetChain()
	tracker.AddToChain("wf-2")
	chain2 := tracker.GetChain()

	// Chains should be independent
	assert.Equal(t, []string{"wf-1"}, chain1)
	assert.Equal(t, []string{"wf-1", "wf-2"}, chain2)
}
