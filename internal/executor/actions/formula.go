package actions

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/gorax/gorax/internal/workflow/formula"
)

// FormulaAction implements the Action interface for formula evaluation
type FormulaAction struct {
	evaluator FormulaEvaluator
}

// FormulaEvaluator interface for formula evaluation (allows dependency injection)
type FormulaEvaluator interface {
	Evaluate(expression string, context map[string]interface{}) (interface{}, error)
}

// FormulaActionConfig represents the configuration for a formula action
type FormulaActionConfig struct {
	// Expression is the formula expression to evaluate
	Expression string `json:"expression"`
	// OutputVariable is the name to store the result (optional, defaults to "result")
	OutputVariable string `json:"output_variable,omitempty"`
}

// SetEvaluator sets the formula evaluator for this action
func (a *FormulaAction) SetEvaluator(evaluator FormulaEvaluator) {
	a.evaluator = evaluator
}

// Execute implements the Action interface
func (a *FormulaAction) Execute(ctx context.Context, input *ActionInput) (*ActionOutput, error) {
	// Initialize evaluator with default if not already set
	if a.evaluator == nil {
		a.evaluator = formula.NewEvaluator()
	}

	// Parse config
	configBytes, err := json.Marshal(input.Config)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal config: %w", err)
	}

	var config FormulaActionConfig
	if err := json.Unmarshal(configBytes, &config); err != nil {
		return nil, fmt.Errorf("failed to parse formula action config: %w", err)
	}

	// Validate expression
	if config.Expression == "" {
		return nil, fmt.Errorf("expression cannot be empty")
	}

	// Evaluate the formula with execution context
	result, err := a.evaluator.Evaluate(config.Expression, input.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to evaluate formula: %w", err)
	}

	// Create output
	output := NewActionOutput(result)

	// Add metadata
	output.WithMetadata("expression", config.Expression)
	output.WithMetadata("output_variable", config.OutputVariable)

	return output, nil
}
