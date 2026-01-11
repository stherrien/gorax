package executor

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/gorax/gorax/internal/executor/actions"
	"github.com/gorax/gorax/internal/workflow"
)

const (
	// ErrorStrategyFailFast stops execution on first error
	ErrorStrategyFailFast = "fail_fast"
	// ErrorStrategyWaitAll waits for all branches even if errors occur
	ErrorStrategyWaitAll = "wait_all"
)

// ParallelResult represents the result of parallel execution
type ParallelResult struct {
	BranchCount   int                    `json:"branch_count"`
	BranchResults []BranchResult         `json:"branch_results"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
}

// BranchResult represents the result of a single parallel branch
type BranchResult struct {
	BranchIndex int                    `json:"branch_index"`
	Output      map[string]interface{} `json:"output,omitempty"`
	Error       *string                `json:"error,omitempty"`
	DurationMs  int64                  `json:"duration_ms"`
}

// parallelExecutor handles parallel execution logic
type parallelExecutor struct {
	mainExecutor *Executor
}

// executeParallel executes multiple branches in parallel
func (pe *parallelExecutor) executeParallel(
	ctx context.Context,
	config workflow.ParallelConfig,
	execCtx *ExecutionContext,
	branchNodes [][]workflow.Node,
) (interface{}, error) {
	// Validate configuration
	if err := pe.validateConfig(config); err != nil {
		return nil, fmt.Errorf("invalid parallel configuration: %w", err)
	}

	// Determine error handling strategy
	errorStrategy := config.ErrorStrategy
	if errorStrategy == "" {
		errorStrategy = ErrorStrategyFailFast
	}

	// Determine concurrency limit
	maxConcurrency := config.MaxConcurrency
	if maxConcurrency == 0 {
		maxConcurrency = len(branchNodes) // unlimited = all at once
	}

	// Check context before starting
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	// Execute branches with concurrency control
	results, err := pe.executeBranches(ctx, branchNodes, execCtx, maxConcurrency, errorStrategy)
	if err != nil {
		return nil, err
	}

	// Build parallel result
	parallelResult := &ParallelResult{
		BranchCount:   len(branchNodes),
		BranchResults: results,
		Metadata: map[string]interface{}{
			"error_strategy":  errorStrategy,
			"max_concurrency": config.MaxConcurrency,
		},
	}

	return parallelResult, nil
}

// validateConfig validates parallel configuration
func (pe *parallelExecutor) validateConfig(config workflow.ParallelConfig) error {
	// Validate error strategy
	if config.ErrorStrategy != "" &&
		config.ErrorStrategy != ErrorStrategyFailFast &&
		config.ErrorStrategy != ErrorStrategyWaitAll {
		return fmt.Errorf("error_strategy must be 'fail_fast' or 'wait_all', got '%s'", config.ErrorStrategy)
	}

	// Validate max concurrency
	if config.MaxConcurrency < 0 {
		return fmt.Errorf("max_concurrency must be >= 0, got %d", config.MaxConcurrency)
	}

	return nil
}

// executeBranches executes branches with concurrency control
func (pe *parallelExecutor) executeBranches(
	ctx context.Context,
	branchNodes [][]workflow.Node,
	execCtx *ExecutionContext,
	maxConcurrency int,
	errorStrategy string,
) ([]BranchResult, error) {
	if len(branchNodes) == 0 {
		return []BranchResult{}, nil
	}

	executor := &branchExecutionCoordinator{
		branchNodes:   branchNodes,
		execCtx:       execCtx,
		errorStrategy: errorStrategy,
		pe:            pe,
	}

	return executor.execute(ctx, maxConcurrency)
}

// branchExecutionCoordinator coordinates parallel branch execution
type branchExecutionCoordinator struct {
	branchNodes   [][]workflow.Node
	execCtx       *ExecutionContext
	errorStrategy string
	pe            *parallelExecutor
	results       []BranchResult
	mu            sync.Mutex
	wg            sync.WaitGroup
	errChan       chan error
	doneChan      chan struct{}
}

// execute runs the parallel branches
func (bec *branchExecutionCoordinator) execute(ctx context.Context, maxConcurrency int) ([]BranchResult, error) {
	bec.initializeChannels(len(bec.branchNodes))
	defer bec.closeChannels()

	branchCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	bec.startBranches(branchCtx, maxConcurrency, cancel)
	return bec.waitForCompletion(ctx, cancel)
}

// initializeChannels sets up channels and results
func (bec *branchExecutionCoordinator) initializeChannels(branchCount int) {
	bec.results = make([]BranchResult, branchCount)
	bec.errChan = make(chan error, branchCount)
	bec.doneChan = make(chan struct{})
}

// closeChannels cleans up channels
func (bec *branchExecutionCoordinator) closeChannels() {
	close(bec.errChan)
}

// startBranches launches goroutines for each branch
func (bec *branchExecutionCoordinator) startBranches(ctx context.Context, maxConcurrency int, cancel context.CancelFunc) {
	sem := make(chan struct{}, maxConcurrency)

	for i := 0; i < len(bec.branchNodes); i++ {
		bec.wg.Add(1)
		go bec.executeBranchWorker(ctx, i, sem, cancel)
	}

	go bec.waitForAllBranches()
}

// executeBranchWorker runs a single branch
func (bec *branchExecutionCoordinator) executeBranchWorker(ctx context.Context, branchIndex int, sem chan struct{}, cancel context.CancelFunc) {
	defer bec.wg.Done()

	if !bec.acquireSemaphore(ctx, sem) {
		return
	}
	defer func() { <-sem }()

	result := bec.pe.executeBranch(ctx, branchIndex, bec.branchNodes[branchIndex], bec.execCtx)
	bec.storeResult(branchIndex, result)
	bec.handleBranchError(result, branchIndex, cancel)
}

// acquireSemaphore attempts to acquire semaphore or returns false if cancelled
func (bec *branchExecutionCoordinator) acquireSemaphore(ctx context.Context, sem chan struct{}) bool {
	select {
	case sem <- struct{}{}:
		return true
	case <-ctx.Done():
		return false
	}
}

// storeResult safely stores branch result
func (bec *branchExecutionCoordinator) storeResult(branchIndex int, result BranchResult) {
	bec.mu.Lock()
	defer bec.mu.Unlock()
	bec.results[branchIndex] = result
}

// handleBranchError handles errors based on strategy
func (bec *branchExecutionCoordinator) handleBranchError(result BranchResult, branchIndex int, cancel context.CancelFunc) {
	if result.Error == nil {
		return
	}

	if bec.errorStrategy == ErrorStrategyFailFast {
		bec.errChan <- fmt.Errorf("branch %d failed: %s", branchIndex, *result.Error)
		cancel()
	}
}

// waitForAllBranches waits for completion and signals done
func (bec *branchExecutionCoordinator) waitForAllBranches() {
	bec.wg.Wait()
	close(bec.doneChan)
}

// waitForCompletion waits for all branches or early termination
func (bec *branchExecutionCoordinator) waitForCompletion(ctx context.Context, cancel context.CancelFunc) ([]BranchResult, error) {
	select {
	case <-bec.doneChan:
		return bec.checkWaitAllErrors()
	case err := <-bec.errChan:
		return bec.handleFailFastError(err)
	case <-ctx.Done():
		return bec.handleContextCancelled(cancel)
	}
}

// checkWaitAllErrors checks for errors in wait-all mode
func (bec *branchExecutionCoordinator) checkWaitAllErrors() ([]BranchResult, error) {
	if bec.errorStrategy != ErrorStrategyWaitAll {
		return bec.results, nil
	}

	for i, result := range bec.results {
		if result.Error != nil {
			return bec.results, fmt.Errorf("branch %d failed: %s", i, *result.Error)
		}
	}

	return bec.results, nil
}

// handleFailFastError handles fail-fast errors
func (bec *branchExecutionCoordinator) handleFailFastError(err error) ([]BranchResult, error) {
	bec.wg.Wait()
	if bec.errorStrategy == ErrorStrategyFailFast {
		return bec.results, err
	}
	return bec.results, nil
}

// handleContextCancelled handles context cancellation
func (bec *branchExecutionCoordinator) handleContextCancelled(cancel context.CancelFunc) ([]BranchResult, error) {
	cancel()
	bec.wg.Wait()
	return bec.results, context.Canceled
}

// executeBranch executes a single branch
func (pe *parallelExecutor) executeBranch(
	ctx context.Context,
	branchIndex int,
	nodes []workflow.Node,
	execCtx *ExecutionContext,
) BranchResult {
	// Create branch-specific execution context
	branchCtx := pe.createBranchContext(execCtx)

	// Execute branch nodes
	output, duration, err := pe.executeBranchNodes(ctx, nodes, branchCtx)

	result := BranchResult{
		BranchIndex: branchIndex,
		Output:      output,
		DurationMs:  duration,
	}

	if err != nil {
		errStr := err.Error()
		result.Error = &errStr
	}

	return result
}

// createBranchContext creates an isolated execution context for a branch
func (pe *parallelExecutor) createBranchContext(parentCtx *ExecutionContext) *ExecutionContext {
	// Create a copy of step outputs
	stepOutputs := make(map[string]interface{})
	for k, v := range parentCtx.StepOutputs {
		stepOutputs[k] = v
	}

	return &ExecutionContext{
		TenantID:         parentCtx.TenantID,
		ExecutionID:      parentCtx.ExecutionID,
		WorkflowID:       parentCtx.WorkflowID,
		TriggerData:      parentCtx.TriggerData,
		StepOutputs:      stepOutputs,
		CredentialValues: parentCtx.CredentialValues,
	}
}

// executeBranchNodes executes all nodes in a branch
func (pe *parallelExecutor) executeBranchNodes(
	ctx context.Context,
	nodes []workflow.Node,
	branchCtx *ExecutionContext,
) (map[string]interface{}, int64, error) {
	startTime := getCurrentTimeMs()

	// If no nodes in branch, return empty output
	if len(nodes) == 0 {
		return make(map[string]interface{}), 0, nil
	}

	// Execute nodes sequentially within the branch
	outputs := make(map[string]interface{})
	for _, node := range nodes {
		// Check context cancellation
		if err := ctx.Err(); err != nil {
			return outputs, getCurrentTimeMs() - startTime, err
		}

		// Execute the node
		var output interface{}
		var execErr error

		// Use the main executor if available
		if pe.mainExecutor != nil {
			output, execErr = pe.mainExecutor.executeNode(ctx, node, branchCtx)
		} else {
			// For testing without full executor, return mock output
			output = map[string]interface{}{"status": "executed"}
			execErr = nil
		}

		if execErr != nil {
			return outputs, getCurrentTimeMs() - startTime, fmt.Errorf("node %s failed: %w", node.ID, execErr)
		}

		// Store output for downstream nodes
		branchCtx.StepOutputs[node.ID] = output
		outputs[node.ID] = output
	}

	duration := getCurrentTimeMs() - startTime
	return outputs, duration, nil
}

// executeParallelAction is the main entry point for parallel execution
func (e *Executor) executeParallelAction(
	ctx context.Context,
	node workflow.Node,
	execCtx *ExecutionContext,
	definition *workflow.WorkflowDefinition,
) (interface{}, error) {
	// Parse parallel configuration
	var config workflow.ParallelConfig
	if err := parseNodeConfig(node, &config); err != nil {
		return nil, fmt.Errorf("failed to parse parallel configuration: %w", err)
	}

	// If config has named branches, use new action-based implementation
	if len(config.Branches) > 0 {
		return e.executeParallelActionV2(ctx, config, execCtx)
	}

	// Otherwise, use legacy implementation for backward compatibility
	return e.executeParallelActionLegacy(ctx, node.ID, config, execCtx, definition)
}

// executeParallelActionV2 uses the new action-based parallel execution with named branches
func (e *Executor) executeParallelActionV2(
	ctx context.Context,
	config workflow.ParallelConfig,
	execCtx *ExecutionContext,
) (interface{}, error) {
	// Create parallel action with node executor
	action := &actions.ParallelAction{}
	action.SetNodeExecutor(&executorNodeAdapter{
		executor: e,
		execCtx:  execCtx,
	})

	// Create action input
	input := actions.NewActionInput(config, execCtx.StepOutputs)

	// Execute the action
	output, err := action.Execute(ctx, input)
	if err != nil {
		return nil, err
	}

	return output.Data, nil
}

// executeParallelActionLegacy maintains backward compatibility with old parallel node format
func (e *Executor) executeParallelActionLegacy(
	ctx context.Context,
	nodeID string,
	config workflow.ParallelConfig,
	execCtx *ExecutionContext,
	definition *workflow.WorkflowDefinition,
) (interface{}, error) {
	// Find parallel branches (groups of nodes connected to this parallel node)
	branchNodes := e.findParallelBranches(nodeID, definition)

	if len(branchNodes) == 0 {
		return nil, fmt.Errorf("parallel node has no branches to execute")
	}

	// Create parallel executor with reference to main executor
	parallelExec := &parallelExecutor{
		mainExecutor: e,
	}

	// Execute the parallel branches
	result, err := parallelExec.executeParallel(ctx, config, execCtx, branchNodes)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// findParallelBranches finds all parallel branches connected to the parallel node
func (e *Executor) findParallelBranches(parallelNodeID string, definition *workflow.WorkflowDefinition) [][]workflow.Node {
	// Find all direct children of the parallel node
	var branchRootIDs []string
	for _, edge := range definition.Edges {
		if edge.Source == parallelNodeID {
			branchRootIDs = append(branchRootIDs, edge.Target)
		}
	}

	// For each branch root, collect all nodes in that branch
	var branches [][]workflow.Node
	nodeMap := buildNodeMap(definition.Nodes)

	for _, rootID := range branchRootIDs {
		// For simplicity, each direct child is considered a separate branch
		// In a more complex implementation, you might want to traverse the graph
		// to collect all nodes until a merge point
		if node, exists := nodeMap[rootID]; exists {
			branches = append(branches, []workflow.Node{node})
		}
	}

	return branches
}

// getCurrentTimeMs returns current time in milliseconds
func getCurrentTimeMs() int64 {
	return time.Now().UnixMilli()
}

// executorNodeAdapter adapts the Executor to the NodeExecutor interface
type executorNodeAdapter struct {
	executor *Executor
	execCtx  *ExecutionContext
}

// ExecuteNode executes a single node by ID
func (a *executorNodeAdapter) ExecuteNode(ctx context.Context, nodeID string, input map[string]interface{}) (map[string]interface{}, error) {
	// This is a simplified adapter - in production, you'd need to:
	// 1. Look up the node by ID from the workflow definition
	// 2. Execute the node with the executor
	// 3. Return the output
	//
	// For now, we'll execute using the existing executeNode method if available
	// This requires access to the node object, which we don't have from just the ID
	//
	// A better approach would be to store the workflow definition in the adapter
	// or have a node registry that can look up nodes by ID

	// Return a placeholder - this will be enhanced in integration
	output := make(map[string]interface{})
	output["nodeId"] = nodeID
	output["status"] = "executed"
	output["input"] = input

	return output, nil
}
