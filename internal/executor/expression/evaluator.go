package expression

import (
	"fmt"
	"reflect"

	"github.com/expr-lang/expr"
	"github.com/expr-lang/expr/vm"
)

// Evaluator evaluates boolean expressions with support for operators
type Evaluator struct {
	parser *Parser
}

// NewEvaluator creates a new expression evaluator
func NewEvaluator() *Evaluator {
	return &Evaluator{
		parser: NewParser(),
	}
}

// EvaluateCondition evaluates a boolean condition expression
// Returns true/false based on the evaluation result
func (e *Evaluator) EvaluateCondition(expression string, context map[string]interface{}) (bool, error) {
	if expression == "" {
		return false, fmt.Errorf("empty expression")
	}

	// Parse the expression
	parsed, err := e.parser.Parse(expression)
	if err != nil {
		return false, fmt.Errorf("failed to parse expression: %w", err)
	}

	// Prepare the expression for evaluation
	exprContent := parsed.Content

	// Compile and evaluate directly with context (expr library handles variable resolution)
	// No need to resolve template variables manually
	program, err := expr.Compile(exprContent, expr.Env(context), expr.AsBool())
	if err != nil {
		return false, fmt.Errorf("failed to compile expression: %w", err)
	}

	// Execute the compiled expression
	result, err := expr.Run(program, context)
	if err != nil {
		return false, fmt.Errorf("failed to evaluate expression: %w", err)
	}

	// Convert result to boolean
	boolResult, ok := result.(bool)
	if !ok {
		return false, fmt.Errorf("expression did not evaluate to boolean, got %T", result)
	}

	return boolResult, nil
}

// Evaluate evaluates any expression and returns the result
// This is more flexible than EvaluateCondition and can return any type
func (e *Evaluator) Evaluate(expression string, context map[string]interface{}) (interface{}, error) {
	if expression == "" {
		return nil, fmt.Errorf("empty expression")
	}

	// Parse the expression
	parsed, err := e.parser.Parse(expression)
	if err != nil {
		return nil, fmt.Errorf("failed to parse expression: %w", err)
	}

	// Prepare the expression for evaluation
	exprContent := parsed.Content

	// Compile and evaluate directly with context (expr library handles variable resolution)
	program, err := expr.Compile(exprContent, expr.Env(context))
	if err != nil {
		return nil, fmt.Errorf("failed to compile expression: %w", err)
	}

	// Execute the compiled expression
	result, err := expr.Run(program, context)
	if err != nil {
		return nil, fmt.Errorf("failed to evaluate expression: %w", err)
	}

	return result, nil
}

// EvaluateWithProgram evaluates a pre-compiled expression program
// This is more efficient when evaluating the same expression multiple times
func (e *Evaluator) EvaluateWithProgram(program *vm.Program, context map[string]interface{}) (interface{}, error) {
	result, err := expr.Run(program, context)
	if err != nil {
		return nil, fmt.Errorf("failed to evaluate expression: %w", err)
	}
	return result, nil
}

// CompileExpression compiles an expression for later evaluation
// This is useful for caching compiled expressions
func (e *Evaluator) CompileExpression(expression string, context map[string]interface{}) (*vm.Program, error) {
	if expression == "" {
		return nil, fmt.Errorf("empty expression")
	}

	// Parse the expression
	parsed, err := e.parser.Parse(expression)
	if err != nil {
		return nil, fmt.Errorf("failed to parse expression: %w", err)
	}

	// Compile the expression directly (expr library handles variable resolution at runtime)
	program, err := expr.Compile(parsed.Content, expr.Env(context))
	if err != nil {
		return nil, fmt.Errorf("failed to compile expression: %w", err)
	}

	return program, nil
}

// ValidateCondition validates that an expression is a valid boolean condition
func (e *Evaluator) ValidateCondition(expression string) error {
	if expression == "" {
		return fmt.Errorf("expression cannot be empty")
	}

	// Parse the expression
	parsed, err := e.parser.Parse(expression)
	if err != nil {
		return fmt.Errorf("failed to parse expression: %w", err)
	}

	// Validate basic syntax
	if err := e.parser.ValidateExpression(expression); err != nil {
		return err
	}

	// Try to compile with a mock context to catch syntax errors
	mockContext := map[string]interface{}{
		"steps": map[string]interface{}{
			"test": map[string]interface{}{
				"status": "success",
				"output": map[string]interface{}{
					"count": 10,
					"data":  []interface{}{1, 2, 3},
				},
			},
		},
		"trigger": map[string]interface{}{
			"body": map[string]interface{}{
				"field": "value",
			},
		},
		"env": map[string]interface{}{
			"tenant_id": "test-tenant",
		},
	}

	// Try to compile the expression with mock context
	_, err = expr.Compile(parsed.Content, expr.Env(mockContext), expr.AsBool())
	if err != nil {
		return fmt.Errorf("invalid condition expression: %w", err)
	}

	return nil
}

// EvaluateBooleanExpression is a convenience method for evaluating simple boolean expressions
// It handles common comparison operators and logical operators
func (e *Evaluator) EvaluateBooleanExpression(left interface{}, operator string, right interface{}) (bool, error) {
	switch operator {
	case "==", "equals":
		return compareEqual(left, right), nil
	case "!=", "not_equals":
		return !compareEqual(left, right), nil
	case ">", "greater_than":
		return compareGreater(left, right)
	case ">=", "greater_or_equal":
		result, err := compareGreater(left, right)
		if err != nil {
			return false, err
		}
		return result || compareEqual(left, right), nil
	case "<", "less_than":
		return compareLess(left, right)
	case "<=", "less_or_equal":
		result, err := compareLess(left, right)
		if err != nil {
			return false, err
		}
		return result || compareEqual(left, right), nil
	case "contains":
		return compareContains(left, right)
	case "starts_with":
		return compareStartsWith(left, right)
	case "ends_with":
		return compareEndsWith(left, right)
	default:
		return false, fmt.Errorf("unsupported operator: %s", operator)
	}
}

// Helper comparison functions

func compareEqual(left, right interface{}) bool {
	return reflect.DeepEqual(left, right)
}

func compareGreater(left, right interface{}) (bool, error) {
	leftNum, err := toFloat64(left)
	if err != nil {
		return false, err
	}
	rightNum, err := toFloat64(right)
	if err != nil {
		return false, err
	}
	return leftNum > rightNum, nil
}

func compareLess(left, right interface{}) (bool, error) {
	leftNum, err := toFloat64(left)
	if err != nil {
		return false, err
	}
	rightNum, err := toFloat64(right)
	if err != nil {
		return false, err
	}
	return leftNum < rightNum, nil
}

func compareContains(haystack, needle interface{}) (bool, error) {
	haystackStr, ok := haystack.(string)
	if !ok {
		return false, fmt.Errorf("contains operator requires string haystack, got %T", haystack)
	}
	needleStr := fmt.Sprintf("%v", needle)
	return contains(haystackStr, needleStr), nil
}

func compareStartsWith(str, prefix interface{}) (bool, error) {
	strVal, ok := str.(string)
	if !ok {
		return false, fmt.Errorf("starts_with operator requires string, got %T", str)
	}
	prefixStr := fmt.Sprintf("%v", prefix)
	return startsWith(strVal, prefixStr), nil
}

func compareEndsWith(str, suffix interface{}) (bool, error) {
	strVal, ok := str.(string)
	if !ok {
		return false, fmt.Errorf("ends_with operator requires string, got %T", str)
	}
	suffixStr := fmt.Sprintf("%v", suffix)
	return endsWith(strVal, suffixStr), nil
}

// Helper conversion functions

func toFloat64(v interface{}) (float64, error) {
	switch val := v.(type) {
	case float64:
		return val, nil
	case float32:
		return float64(val), nil
	case int:
		return float64(val), nil
	case int8:
		return float64(val), nil
	case int16:
		return float64(val), nil
	case int32:
		return float64(val), nil
	case int64:
		return float64(val), nil
	case uint:
		return float64(val), nil
	case uint8:
		return float64(val), nil
	case uint16:
		return float64(val), nil
	case uint32:
		return float64(val), nil
	case uint64:
		return float64(val), nil
	default:
		return 0, fmt.Errorf("cannot convert %T to number", v)
	}
}

// String helper functions
func contains(s, substr string) bool {
	return len(s) >= len(substr) && indexOf(s, substr) >= 0
}

func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

func startsWith(s, prefix string) bool {
	return len(s) >= len(prefix) && s[:len(prefix)] == prefix
}

func endsWith(s, suffix string) bool {
	return len(s) >= len(suffix) && s[len(s)-len(suffix):] == suffix
}
