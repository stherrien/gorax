package executor

import (
	"context"
	"fmt"
	"time"

	"github.com/gorax/gorax/internal/workflow"
)

const (
	// JoinStrategyWaitAll waits for all branches to complete
	JoinStrategyWaitAll = "wait_all"
	// JoinStrategyWaitN waits for N branches to complete
	JoinStrategyWaitN = "wait_n"
	// OnTimeoutFail fails the join when timeout occurs
	OnTimeoutFail = "fail"
	// OnTimeoutContinue continues with partial results when timeout occurs
	OnTimeoutContinue = "continue"
)

// ForkResult represents the result of a fork execution
type ForkResult struct {
	BranchCount int                    `json:"branch_count"`
	BranchIDs   []string               `json:"branch_ids"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// JoinResult represents the result of a join execution
type JoinResult struct {
	CompletedBranches int                    `json:"completed_branches"`
	BranchOutputs     map[string]interface{} `json:"branch_outputs"`
	TimedOut          bool                   `json:"timed_out"`
	Metadata          map[string]interface{} `json:"metadata,omitempty"`
}

// forkExecutor handles fork execution logic
type forkExecutor struct {
	mainExecutor *Executor
}

// executeFork executes a fork node that splits into multiple branches
func (fe *forkExecutor) executeFork(
	ctx context.Context,
	config workflow.ForkConfig,
	execCtx *ExecutionContext,
) (interface{}, error) {
	// Validate configuration
	if err := fe.validateConfig(config); err != nil {
		return nil, fmt.Errorf("invalid fork configuration: %w", err)
	}

	// Generate branch IDs
	branchIDs := make([]string, config.BranchCount)
	for i := 0; i < config.BranchCount; i++ {
		branchIDs[i] = fmt.Sprintf("branch_%d", i)
	}

	// Create fork result
	result := &ForkResult{
		BranchCount: config.BranchCount,
		BranchIDs:   branchIDs,
		Metadata: map[string]interface{}{
			"execution_id": execCtx.ExecutionID,
			"fork_time":    time.Now().Unix(),
		},
	}

	return result, nil
}

// validateConfig validates fork configuration
func (fe *forkExecutor) validateConfig(config workflow.ForkConfig) error {
	if config.BranchCount <= 0 {
		return fmt.Errorf("branch_count must be greater than 0, got %d", config.BranchCount)
	}
	return nil
}

// joinExecutor handles join execution logic
type joinExecutor struct {
	mainExecutor *Executor
}

// executeJoin executes a join node that synchronizes multiple branches
func (je *joinExecutor) executeJoin(
	ctx context.Context,
	config workflow.JoinConfig,
	execCtx *ExecutionContext,
	branchIDs []string,
) (interface{}, error) {
	// Validate configuration
	if err := je.validateConfig(config, len(branchIDs)); err != nil {
		return nil, fmt.Errorf("invalid join configuration: %w", err)
	}

	// Determine join strategy
	strategy := config.JoinStrategy
	if strategy == "" {
		strategy = JoinStrategyWaitAll
	}

	// Determine timeout
	timeoutDuration := time.Duration(0)
	if config.TimeoutMs > 0 {
		timeoutDuration = time.Duration(config.TimeoutMs) * time.Millisecond
	}

	// Wait for branches based on strategy
	var result *JoinResult
	var err error

	switch strategy {
	case JoinStrategyWaitAll:
		result, err = je.waitForAllBranches(ctx, execCtx, branchIDs, timeoutDuration, config.OnTimeout)
	case JoinStrategyWaitN:
		result, err = je.waitForNBranches(ctx, execCtx, branchIDs, config.RequiredCount, timeoutDuration, config.OnTimeout)
	default:
		return nil, fmt.Errorf("unknown join_strategy: %s", strategy)
	}

	if err != nil {
		return nil, err
	}

	// Add metadata
	result.Metadata = map[string]interface{}{
		"strategy":       strategy,
		"required_count": config.RequiredCount,
		"timeout_ms":     config.TimeoutMs,
	}

	return result, nil
}

// validateConfig validates join configuration
func (je *joinExecutor) validateConfig(config workflow.JoinConfig, branchCount int) error {
	// Validate strategy
	if config.JoinStrategy != "" &&
		config.JoinStrategy != JoinStrategyWaitAll &&
		config.JoinStrategy != JoinStrategyWaitN {
		return fmt.Errorf("join_strategy must be 'wait_all' or 'wait_n', got '%s'", config.JoinStrategy)
	}

	// Validate wait_n specific config
	if config.JoinStrategy == JoinStrategyWaitN {
		if config.RequiredCount <= 0 {
			return fmt.Errorf("required_count must be greater than 0 for wait_n strategy, got %d", config.RequiredCount)
		}
		if config.RequiredCount > branchCount {
			return fmt.Errorf("required_count (%d) cannot exceed total branches (%d)", config.RequiredCount, branchCount)
		}
	}

	// Validate timeout config
	if config.TimeoutMs < 0 {
		return fmt.Errorf("timeout_ms must be non-negative, got %d", config.TimeoutMs)
	}

	// Validate on_timeout
	if config.OnTimeout != "" &&
		config.OnTimeout != OnTimeoutFail &&
		config.OnTimeout != OnTimeoutContinue {
		return fmt.Errorf("on_timeout must be 'fail' or 'continue', got '%s'", config.OnTimeout)
	}

	return nil
}

// waitForAllBranches waits for all branches to complete
func (je *joinExecutor) waitForAllBranches(
	ctx context.Context,
	execCtx *ExecutionContext,
	branchIDs []string,
	timeout time.Duration,
	onTimeout string,
) (*JoinResult, error) {
	// If no branches, return immediately
	if len(branchIDs) == 0 {
		return &JoinResult{
			CompletedBranches: 0,
			BranchOutputs:     make(map[string]interface{}),
			TimedOut:          false,
		}, nil
	}

	// Create context with timeout if specified
	waitCtx := ctx
	if timeout > 0 {
		var cancel context.CancelFunc
		waitCtx, cancel = context.WithTimeout(ctx, timeout)
		defer cancel()
	}

	// Collect branch outputs
	branchOutputs := make(map[string]interface{})
	completedCount := 0

	// First, collect all currently completed branches
	for _, branchID := range branchIDs {
		if output, exists := execCtx.StepOutputs[branchID]; exists {
			branchOutputs[branchID] = output
			completedCount++
		}
	}

	// Check if all branches are complete
	if completedCount == len(branchIDs) {
		return &JoinResult{
			CompletedBranches: completedCount,
			BranchOutputs:     branchOutputs,
			TimedOut:          false,
		}, nil
	}

	// Wait for remaining branches
	select {
	case <-waitCtx.Done():
		// Timeout occurred
		if onTimeout == OnTimeoutContinue {
			// Continue with partial results
			return &JoinResult{
				CompletedBranches: completedCount,
				BranchOutputs:     branchOutputs,
				TimedOut:          true,
			}, nil
		}
		// Fail on timeout
		return nil, fmt.Errorf("join timeout: waiting for branches")
	default:
		// All branches complete (shouldn't reach here)
		return &JoinResult{
			CompletedBranches: completedCount,
			BranchOutputs:     branchOutputs,
			TimedOut:          false,
		}, nil
	}
}

// waitForNBranches waits for N branches to complete
func (je *joinExecutor) waitForNBranches(
	ctx context.Context,
	execCtx *ExecutionContext,
	branchIDs []string,
	requiredCount int,
	timeout time.Duration,
	onTimeout string,
) (*JoinResult, error) {
	// If no branches, return immediately
	if len(branchIDs) == 0 {
		return &JoinResult{
			CompletedBranches: 0,
			BranchOutputs:     make(map[string]interface{}),
			TimedOut:          false,
		}, nil
	}

	// Create context with timeout if specified
	waitCtx := ctx
	if timeout > 0 {
		var cancel context.CancelFunc
		waitCtx, cancel = context.WithTimeout(ctx, timeout)
		defer cancel()
	}

	// Collect branch outputs until we have enough
	branchOutputs := make(map[string]interface{})
	completedCount := 0

	for _, branchID := range branchIDs {
		// Check if we've reached required count
		if completedCount >= requiredCount {
			break
		}

		// Check for timeout
		select {
		case <-waitCtx.Done():
			// Timeout occurred
			if onTimeout == OnTimeoutContinue {
				// Continue with partial results
				return &JoinResult{
					CompletedBranches: completedCount,
					BranchOutputs:     branchOutputs,
					TimedOut:          true,
				}, nil
			}
			// Fail on timeout
			return nil, fmt.Errorf("join timeout: waiting for %d branches, got %d", requiredCount, completedCount)
		default:
			// Check if branch has completed
			if output, exists := execCtx.StepOutputs[branchID]; exists {
				branchOutputs[branchID] = output
				completedCount++
			}
		}
	}

	return &JoinResult{
		CompletedBranches: completedCount,
		BranchOutputs:     branchOutputs,
		TimedOut:          false,
	}, nil
}

// executeForkAction is the main entry point for fork execution
func (e *Executor) executeForkAction(
	ctx context.Context,
	node workflow.Node,
	execCtx *ExecutionContext,
) (interface{}, error) {
	// Parse fork configuration
	var config workflow.ForkConfig
	if err := parseNodeConfig(node, &config); err != nil {
		return nil, fmt.Errorf("failed to parse fork configuration: %w", err)
	}

	// Create fork executor
	forkExec := &forkExecutor{
		mainExecutor: e,
	}

	// Execute the fork
	result, err := forkExec.executeFork(ctx, config, execCtx)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// executeJoinAction is the main entry point for join execution
func (e *Executor) executeJoinAction(
	ctx context.Context,
	node workflow.Node,
	execCtx *ExecutionContext,
	definition *workflow.WorkflowDefinition,
) (interface{}, error) {
	// Parse join configuration
	var config workflow.JoinConfig
	if err := parseNodeConfig(node, &config); err != nil {
		return nil, fmt.Errorf("failed to parse join configuration: %w", err)
	}

	// Find incoming branches to this join node
	branchIDs := e.findIncomingBranches(node.ID, definition)

	// Create join executor
	joinExec := &joinExecutor{
		mainExecutor: e,
	}

	// Execute the join
	result, err := joinExec.executeJoin(ctx, config, execCtx, branchIDs)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// findIncomingBranches finds all nodes that connect to the join node
func (e *Executor) findIncomingBranches(joinNodeID string, definition *workflow.WorkflowDefinition) []string {
	var branchIDs []string
	for _, edge := range definition.Edges {
		if edge.Target == joinNodeID {
			branchIDs = append(branchIDs, edge.Source)
		}
	}
	return branchIDs
}
