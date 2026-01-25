package executor

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/gorax/gorax/internal/executor/actions"
	"github.com/gorax/gorax/internal/executor/javascript"
	"github.com/gorax/gorax/internal/tracing"
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

	// Execute HTTP request through circuit breaker with tracing
	result, err := tracing.TraceHTTPAction(ctx, config.Method, config.URL, func(tracedCtx context.Context) (interface{}, error) {
		return circuitBreaker.ExecuteWithResult(tracedCtx, func(reqCtx context.Context) (interface{}, error) {
			return actions.ExecuteHTTP(reqCtx, config, execContext)
		})
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

// ScriptActionConfig represents the configuration for a code action node
type ScriptActionConfig struct {
	Script  string `json:"script"`
	Timeout int    `json:"timeout,omitempty"` // Timeout in seconds
}

// executeCodeAction executes a custom code/script action using the sandboxed JavaScript engine
func (e *Executor) executeCodeAction(ctx context.Context, node workflow.Node, execCtx *ExecutionContext) (interface{}, error) {
	// Extract config from node data
	configData := node.Data.Config
	if len(configData) == 0 {
		return nil, fmt.Errorf("missing config for code action")
	}

	// Parse the script configuration
	var scriptConfig ScriptActionConfig
	if err := json.Unmarshal(configData, &scriptConfig); err != nil {
		return nil, fmt.Errorf("failed to parse code action config: %w", err)
	}

	// Validate script
	if scriptConfig.Script == "" {
		return nil, fmt.Errorf("script is required for code action")
	}

	// Use the JavaScript engine if available, otherwise fall back to legacy action
	if e.jsEngine != nil {
		return e.executeWithJSEngine(ctx, scriptConfig, node.ID, execCtx)
	}

	// Fallback to legacy ScriptAction
	return e.executeWithLegacyAction(ctx, configData, execCtx)
}

// executeWithJSEngine executes script using the new comprehensive JavaScript engine
func (e *Executor) executeWithJSEngine(ctx context.Context, config ScriptActionConfig, nodeID string, execCtx *ExecutionContext) (interface{}, error) {
	// Build JavaScript execution context
	jsCtx := javascript.NewExecutionContext().
		WithTrigger(execCtx.TriggerData).
		WithSteps(execCtx.StepOutputs).
		WithEnv(map[string]any{
			"tenant_id":    execCtx.TenantID,
			"execution_id": execCtx.ExecutionID,
			"workflow_id":  execCtx.WorkflowID,
		})

	// Determine timeout
	var timeout time.Duration
	if config.Timeout > 0 {
		timeout = time.Duration(config.Timeout) * time.Second
	}

	// Create execution config
	executeConfig := &javascript.ExecuteConfig{
		Script:      config.Script,
		Context:     jsCtx,
		Timeout:     timeout,
		ExecutionID: execCtx.ExecutionID,
		TenantID:    execCtx.TenantID,
		WorkflowID:  execCtx.WorkflowID,
		NodeID:      nodeID,
		UserID:      execCtx.GetUserID(),
	}

	// Execute the script
	result, err := e.jsEngine.Execute(ctx, executeConfig)
	if err != nil {
		return nil, fmt.Errorf("javascript execution failed: %w", err)
	}

	// Log console output if any
	if len(result.ConsoleLogs) > 0 {
		e.logger.Debug("javascript console output",
			"execution_id", execCtx.ExecutionID,
			"node_id", nodeID,
			"log_count", len(result.ConsoleLogs),
		)
	}

	return result.Result, nil
}

// executeWithLegacyAction falls back to the legacy ScriptAction implementation
func (e *Executor) executeWithLegacyAction(ctx context.Context, configData json.RawMessage, execCtx *ExecutionContext) (interface{}, error) {
	action := &actions.ScriptAction{}
	context := buildInterpolationContext(execCtx)
	input := actions.NewActionInput(configData, context)

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
