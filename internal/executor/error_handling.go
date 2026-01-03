package executor

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/gorax/gorax/internal/executor/actions"
	"github.com/gorax/gorax/internal/workflow"
)

// executeTryAction executes a try/catch/finally error handling node
func (e *Executor) executeTryAction(ctx context.Context, node workflow.Node, execCtx *ExecutionContext) (interface{}, error) {
	// Parse try configuration
	var config workflow.TryConfig
	if err := json.Unmarshal(node.Data.Config, &config); err != nil {
		return nil, fmt.Errorf("failed to parse try config: %w", err)
	}

	// Create helper functions for the try action
	executeNode := func(ctx context.Context, nodeID string) (interface{}, error) {
		// This would need to execute a node by ID from the workflow definition
		// For now, return error indicating this needs workflow definition access
		return nil, fmt.Errorf("try action requires workflow definition access - not yet implemented")
	}

	getNode := func(nodeID string) (*workflow.Node, error) {
		return nil, fmt.Errorf("get node not yet implemented")
	}

	// Create try action
	tryAction := actions.NewTryAction(executeNode, getNode)

	// Build input data
	inputData := buildInputData(execCtx)

	// Execute try action
	actionInput := actions.NewActionInput(config, inputData)
	output, err := tryAction.Execute(ctx, actionInput)

	if err != nil {
		return nil, err
	}

	return output.Data, nil
}

// executeRetryAction executes a retry error handling node
func (e *Executor) executeRetryAction(ctx context.Context, node workflow.Node, execCtx *ExecutionContext) (interface{}, error) {
	// Parse retry configuration
	var config workflow.RetryNodeConfig
	if err := json.Unmarshal(node.Data.Config, &config); err != nil {
		return nil, fmt.Errorf("failed to parse retry config: %w", err)
	}

	// Create helper function for the retry action
	executeNode := func(ctx context.Context, nodeID string) (interface{}, error) {
		// This would need to execute a node by ID from the workflow definition
		// For now, return error indicating this needs workflow definition access
		return nil, fmt.Errorf("retry action requires workflow definition access - not yet implemented")
	}

	// Create retry action
	retryAction := actions.NewRetryAction(executeNode)

	// Build input data with retry_node_id in context
	inputData := buildInputData(execCtx)

	// Get the node ID to retry from the config or context
	// This would typically come from the workflow graph
	retryNodeID := node.ID + "-target" // Placeholder
	inputData["retry_node_id"] = retryNodeID

	// Execute retry action
	actionInput := actions.NewActionInput(config, inputData)
	output, err := retryAction.Execute(ctx, actionInput)

	if err != nil {
		return nil, err
	}

	return output.Data, nil
}

// executeCircuitBreakerAction executes a circuit breaker error handling node
func (e *Executor) executeCircuitBreakerAction(ctx context.Context, node workflow.Node, execCtx *ExecutionContext) (interface{}, error) {
	// Parse circuit breaker configuration
	var config workflow.CircuitBreakerConfig
	if err := json.Unmarshal(node.Data.Config, &config); err != nil {
		return nil, fmt.Errorf("failed to parse circuit breaker config: %w", err)
	}

	// Get or create circuit breaker
	breakerName := config.Name
	if breakerName == "" {
		breakerName = node.ID
	}

	breaker := e.circuitBreakers.GetOrCreate(breakerName)

	// Execute the protected operation
	result, err := breaker.ExecuteWithResult(ctx, func(ctx context.Context) (interface{}, error) {
		// This would need to execute the wrapped node
		// For now, return a placeholder
		return map[string]interface{}{
			"circuit_breaker": breakerName,
			"state":           breaker.GetState().String(),
		}, nil
	})

	if err != nil {
		return nil, fmt.Errorf("circuit breaker operation failed: %w", err)
	}

	return result, nil
}
