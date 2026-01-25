package actions

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/gorax/gorax/internal/workflow"
)

// ParallelAction executes multiple branches in parallel
type ParallelAction struct {
	// Optional: inject executor for node execution
	// In production, this would be set by the main executor
	nodeExecutor NodeExecutor
}

// NodeExecutor defines the interface for executing workflow nodes
type NodeExecutor interface {
	ExecuteNode(ctx context.Context, nodeID string, input map[string]interface{}) (map[string]interface{}, error)
}

// SetNodeExecutor sets the node executor for this action
func (a *ParallelAction) SetNodeExecutor(executor NodeExecutor) {
	a.nodeExecutor = executor
}

// Execute runs the parallel action
func (a *ParallelAction) Execute(ctx context.Context, input *ActionInput) (*ActionOutput, error) {
	// Parse configuration
	config, ok := input.Config.(workflow.ParallelConfig)
	if !ok {
		return nil, fmt.Errorf("invalid parallel configuration type")
	}

	// Validate configuration
	if err := a.validateConfig(config); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	// Normalize configuration (apply defaults and backward compatibility)
	config = a.normalizeConfig(config)

	// Create execution context with timeout if specified
	execCtx, cancel := a.createExecutionContext(ctx, config)
	defer cancel()

	// Execute parallel branches
	results, err := a.executeBranches(execCtx, config, input.Context)
	if err != nil {
		return nil, err
	}

	// Build output
	output := NewActionOutput(results)
	output.Metadata["branch_count"] = len(config.Branches)
	output.Metadata["wait_mode"] = config.WaitMode
	output.Metadata["failure_mode"] = config.FailureMode
	output.Metadata["max_concurrency"] = config.MaxConcurrency

	return output, nil
}

// validateConfig validates the parallel configuration
func (a *ParallelAction) validateConfig(config workflow.ParallelConfig) error {
	// Validate branches
	if len(config.Branches) == 0 {
		return fmt.Errorf("no branches specified for parallel execution")
	}

	// Validate wait mode
	if config.WaitMode != "" && config.WaitMode != "all" && config.WaitMode != "first" {
		return fmt.Errorf("wait_mode must be 'all' or 'first', got '%s'", config.WaitMode)
	}

	// Validate failure mode
	if config.FailureMode != "" && config.FailureMode != "stop_all" && config.FailureMode != "continue" {
		return fmt.Errorf("failure_mode must be 'stop_all' or 'continue', got '%s'", config.FailureMode)
	}

	// Validate max concurrency
	if config.MaxConcurrency < 0 {
		return fmt.Errorf("max_concurrency must be >= 0, got %d", config.MaxConcurrency)
	}

	// Validate timeout format if specified
	if config.Timeout != "" {
		if _, err := time.ParseDuration(config.Timeout); err != nil {
			return fmt.Errorf("invalid timeout format '%s': %w", config.Timeout, err)
		}
	}

	// Validate branch names and nodes
	for i, branch := range config.Branches {
		if branch.Name == "" {
			return fmt.Errorf("branch %d has empty name", i)
		}
		if len(branch.Nodes) == 0 {
			return fmt.Errorf("branch '%s' has no nodes", branch.Name)
		}
	}

	return nil
}

// normalizeConfig applies defaults and backward compatibility mappings
func (a *ParallelAction) normalizeConfig(config workflow.ParallelConfig) workflow.ParallelConfig {
	// Apply wait mode default
	if config.WaitMode == "" {
		config.WaitMode = "all"
	}

	// Apply failure mode default or map from legacy error strategy
	if config.FailureMode == "" {
		if config.ErrorStrategy == "fail_fast" {
			config.FailureMode = "stop_all"
		} else if config.ErrorStrategy == "wait_all" {
			config.FailureMode = "continue"
		} else {
			config.FailureMode = "stop_all" // default
		}
	}

	return config
}

// createExecutionContext creates a context with timeout if specified
func (a *ParallelAction) createExecutionContext(ctx context.Context, config workflow.ParallelConfig) (context.Context, context.CancelFunc) {
	if config.Timeout != "" {
		if duration, err := time.ParseDuration(config.Timeout); err == nil {
			return context.WithTimeout(ctx, duration)
		}
	}
	return context.WithCancel(ctx)
}

// executeBranches executes all branches according to configuration
func (a *ParallelAction) executeBranches(ctx context.Context, config workflow.ParallelConfig, execContext map[string]interface{}) (map[string]interface{}, error) {
	// Check for context cancellation before starting
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("context cancelled before execution: %w", err)
	}

	// Determine concurrency limit
	maxConcurrency := config.MaxConcurrency
	if maxConcurrency == 0 {
		maxConcurrency = len(config.Branches)
	}

	// Create coordinator for branch execution
	coordinator := &branchCoordinator{
		branches:     config.Branches,
		waitMode:     config.WaitMode,
		failureMode:  config.FailureMode,
		execContext:  execContext,
		action:       a,
		results:      make(map[string]*branchResult),
		resultsMutex: sync.Mutex{},
	}

	return coordinator.execute(ctx, maxConcurrency)
}

// branchCoordinator coordinates parallel branch execution
type branchCoordinator struct {
	branches     []workflow.ParallelBranch
	waitMode     string
	failureMode  string
	execContext  map[string]interface{}
	action       *ParallelAction
	results      map[string]*branchResult
	resultsMutex sync.Mutex
	wg           sync.WaitGroup
	errChan      chan error
	doneChan     chan *branchResult
}

// branchResult represents the result of a single branch execution
type branchResult struct {
	BranchName string                 `json:"branch_name"`
	Output     map[string]interface{} `json:"output,omitempty"`
	Error      string                 `json:"error,omitempty"`
	DurationMs int64                  `json:"duration_ms"`
}

// execute runs all branches
func (bc *branchCoordinator) execute(ctx context.Context, maxConcurrency int) (map[string]interface{}, error) {
	// Initialize channels
	bc.errChan = make(chan error, len(bc.branches))
	bc.doneChan = make(chan *branchResult, len(bc.branches))

	// Create context for branch execution
	branchCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Start branch execution workers
	sem := make(chan struct{}, maxConcurrency)
	for _, branch := range bc.branches {
		bc.wg.Add(1)
		go bc.executeBranchWorker(branchCtx, branch, sem, cancel)
	}

	// Wait for completion based on wait mode
	go bc.waitForCompletion()

	// Wait for results
	return bc.collectResults(branchCtx, cancel)
}

// executeBranchWorker executes a single branch
func (bc *branchCoordinator) executeBranchWorker(ctx context.Context, branch workflow.ParallelBranch, sem chan struct{}, cancel context.CancelFunc) {
	defer bc.wg.Done()

	// Acquire semaphore
	select {
	case sem <- struct{}{}:
		defer func() { <-sem }()
	case <-ctx.Done():
		return
	}

	// Execute the branch
	startTime := time.Now()
	output, err := bc.executeBranch(ctx, branch)
	duration := time.Since(startTime).Milliseconds()

	// Create result
	result := &branchResult{
		BranchName: branch.Name,
		Output:     output,
		DurationMs: duration,
	}

	if err != nil {
		result.Error = err.Error()

		// Handle error based on failure mode
		if bc.failureMode == "stop_all" {
			bc.errChan <- fmt.Errorf("branch '%s' failed: %w", branch.Name, err)
			cancel() // Cancel all other branches
			return
		}
	}

	// Store result
	bc.resultsMutex.Lock()
	bc.results[branch.Name] = result
	bc.resultsMutex.Unlock()

	// Send completion notification
	bc.doneChan <- result
}

// executeBranch executes nodes in a single branch sequentially
func (bc *branchCoordinator) executeBranch(ctx context.Context, branch workflow.ParallelBranch) (map[string]interface{}, error) {
	output := make(map[string]interface{})

	// Create branch-specific context (copy of execution context)
	branchContext := bc.copyContext(bc.execContext)

	// Execute nodes sequentially within the branch
	for _, nodeID := range branch.Nodes {
		// Check for cancellation
		if err := ctx.Err(); err != nil {
			return output, err
		}

		// Execute node
		var nodeOutput map[string]interface{}
		var err error

		if bc.action.nodeExecutor != nil {
			nodeOutput, err = bc.action.nodeExecutor.ExecuteNode(ctx, nodeID, branchContext)
		} else {
			// Mock execution for unit tests without executor
			nodeOutput = map[string]interface{}{
				"nodeId": nodeID,
				"status": "executed",
			}
		}

		if err != nil {
			return output, fmt.Errorf("node '%s' failed: %w", nodeID, err)
		}

		// Store output for downstream nodes in branch
		output[nodeID] = nodeOutput
		branchContext[nodeID] = nodeOutput
	}

	return output, nil
}

// copyContext creates a copy of the execution context
func (bc *branchCoordinator) copyContext(original map[string]interface{}) map[string]interface{} {
	copy := make(map[string]interface{}, len(original))
	for k, v := range original {
		copy[k] = v
	}
	return copy
}

// waitForCompletion waits for all branches to complete
func (bc *branchCoordinator) waitForCompletion() {
	bc.wg.Wait()
	close(bc.doneChan)
	close(bc.errChan)
}

// collectResults collects results based on wait mode
func (bc *branchCoordinator) collectResults(ctx context.Context, cancel context.CancelFunc) (map[string]interface{}, error) {
	if bc.waitMode == "first" {
		return bc.collectFirstResult(ctx, cancel)
	}
	return bc.collectAllResults(ctx)
}

// collectFirstResult waits for first branch to complete
func (bc *branchCoordinator) collectFirstResult(ctx context.Context, cancel context.CancelFunc) (map[string]interface{}, error) {
	select {
	case result := <-bc.doneChan:
		if result == nil {
			return nil, fmt.Errorf("no branches completed successfully")
		}
		cancel() // Cancel remaining branches
		bc.wg.Wait()

		if result.Error != "" {
			return nil, fmt.Errorf("first branch failed: %s", result.Error)
		}

		return map[string]interface{}{
			"first_completed": result.BranchName,
			"result":          result,
		}, nil

	case err := <-bc.errChan:
		cancel()
		bc.wg.Wait()
		return nil, err

	case <-ctx.Done():
		cancel()
		bc.wg.Wait()
		return nil, ctx.Err()
	}
}

// collectAllResults waits for all branches to complete
func (bc *branchCoordinator) collectAllResults(ctx context.Context) (map[string]interface{}, error) {
	// Wait for all branches to complete or error
	var firstError error

	for {
		select {
		case result, ok := <-bc.doneChan:
			if !ok {
				// All branches completed
				return bc.buildFinalResults(firstError)
			}
			if result != nil && result.Error != "" && firstError == nil {
				firstError = fmt.Errorf("branch '%s' failed: %s", result.BranchName, result.Error)
			}

		case err := <-bc.errChan:
			if firstError == nil {
				firstError = err
			}
			// Continue waiting for other branches if failure mode is continue
			if bc.failureMode == "stop_all" {
				bc.wg.Wait()
				return bc.buildFinalResults(firstError)
			}

		case <-ctx.Done():
			bc.wg.Wait()
			return nil, ctx.Err()
		}
	}
}

// buildFinalResults builds the final result map
func (bc *branchCoordinator) buildFinalResults(err error) (map[string]interface{}, error) {
	bc.resultsMutex.Lock()
	defer bc.resultsMutex.Unlock()

	results := make(map[string]interface{})
	branchResults := make([]interface{}, 0, len(bc.results))

	for _, result := range bc.results {
		branchResults = append(branchResults, result)
	}

	results["branches"] = branchResults
	results["total_branches"] = len(bc.branches)
	results["completed_branches"] = len(bc.results)

	// If failure mode is "continue", don't return error even if some branches failed
	if bc.failureMode == "continue" {
		return results, nil
	}

	return results, err
}
