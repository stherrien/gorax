package executor

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/gorax/gorax/internal/executor/actions"
	"github.com/gorax/gorax/internal/workflow"
)

// executeHTTPAction executes an HTTP request action with circuit breaker
func (e *Executor) executeHTTPAction(ctx context.Context, node workflow.Node, execCtx *ExecutionContext) (interface{}, error) {
	// Extract config from node data
	configData := node.Data.Config
	if len(configData) == 0 {
		return nil, fmt.Errorf("missing config for HTTP action")
	}

	// Parse node config
	var config actions.HTTPActionConfig
	if err := json.Unmarshal(configData, &config); err != nil {
		return nil, fmt.Errorf("failed to parse HTTP action config: %w", err)
	}

	// Build execution context for interpolation
	execContext := buildInterpolationContext(execCtx)

	// Get or create circuit breaker for this URL/host
	circuitBreakerName := fmt.Sprintf("http:%s", config.URL)
	circuitBreaker := e.circuitBreakers.GetOrCreate(circuitBreakerName)

	// Execute HTTP request through circuit breaker
	result, err := circuitBreaker.ExecuteWithResult(ctx, func(reqCtx context.Context) (interface{}, error) {
		return actions.ExecuteHTTP(reqCtx, config, execContext)
	})

	if err != nil {
		return nil, err
	}

	return result, nil
}

// executeTransformAction executes a data transformation action
func (e *Executor) executeTransformAction(ctx context.Context, node workflow.Node, execCtx *ExecutionContext) (interface{}, error) {
	// Extract config from node data
	configData := node.Data.Config
	if len(configData) == 0 {
		return nil, fmt.Errorf("missing config for transform action")
	}

	// Parse node config
	var config actions.TransformActionConfig
	if err := json.Unmarshal(configData, &config); err != nil {
		return nil, fmt.Errorf("failed to parse transform action config: %w", err)
	}

	// Build execution context for interpolation
	context := buildInterpolationContext(execCtx)

	// Execute transformation
	result, err := actions.ExecuteTransform(ctx, config, context)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// executeFormulaAction executes a formula evaluation action
func (e *Executor) executeFormulaAction(ctx context.Context, node workflow.Node, execCtx *ExecutionContext) (interface{}, error) {
	// Extract config from node data
	configData := node.Data.Config
	if len(configData) == 0 {
		return nil, fmt.Errorf("missing config for formula action")
	}

	// Create formula action with evaluator from executor if available
	action := &actions.FormulaAction{}
	if e.formulaEvaluator != nil {
		action.SetEvaluator(e.formulaEvaluator)
	}

	// Build execution context
	context := buildInterpolationContext(execCtx)

	// Create action input
	input := actions.NewActionInput(configData, context)

	// Execute formula action
	output, err := action.Execute(ctx, input)
	if err != nil {
		return nil, err
	}

	return output.Data, nil
}

// executeCodeAction executes a custom code/script action
func (e *Executor) executeCodeAction(ctx context.Context, node workflow.Node, execCtx *ExecutionContext) (interface{}, error) {
	// Extract config from node data
	configData := node.Data.Config
	if len(configData) == 0 {
		return nil, fmt.Errorf("missing config for code action")
	}

	// Create script action
	action := &actions.ScriptAction{}

	// Build execution context
	context := buildInterpolationContext(execCtx)

	// Create action input
	input := actions.NewActionInput(configData, context)

	// Execute script action
	output, err := action.Execute(ctx, input)
	if err != nil {
		return nil, err
	}

	return output.Data, nil
}

// buildInterpolationContext builds the context map for template interpolation
func buildInterpolationContext(execCtx *ExecutionContext) map[string]interface{} {
	return map[string]interface{}{
		"trigger": execCtx.TriggerData,
		"steps":   execCtx.StepOutputs,
		"env": map[string]interface{}{
			"tenant_id":    execCtx.TenantID,
			"execution_id": execCtx.ExecutionID,
			"workflow_id":  execCtx.WorkflowID,
		},
	}
}
