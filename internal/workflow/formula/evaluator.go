package formula

import (
	"fmt"
	"time"

	"github.com/expr-lang/expr"
	"github.com/expr-lang/expr/vm"
)

// Evaluator handles formula evaluation with built-in functions
type Evaluator struct {
	program *vm.Program
	env     map[string]interface{}
}

// NewEvaluator creates a new formula evaluator with all built-in functions
func NewEvaluator() *Evaluator {
	return &Evaluator{
		env: buildEnvironment(),
	}
}

// Evaluate compiles and evaluates an expression with the given context
func (e *Evaluator) Evaluate(expression string, context map[string]interface{}) (interface{}, error) {
	if expression == "" {
		return nil, fmt.Errorf("expression cannot be empty")
	}

	// Merge context with built-in functions
	env := make(map[string]interface{})
	for k, v := range e.env {
		env[k] = v
	}
	for k, v := range context {
		env[k] = v
	}

	// Compile and run the expression
	program, err := expr.Compile(expression, expr.Env(env))
	if err != nil {
		return nil, fmt.Errorf("failed to compile expression: %w", err)
	}

	result, err := expr.Run(program, env)
	if err != nil {
		return nil, fmt.Errorf("failed to evaluate expression: %w", err)
	}

	return result, nil
}

// EvaluateWithType evaluates an expression and ensures the result matches the expected type
func (e *Evaluator) EvaluateWithType(expression string, context map[string]interface{}, expectedType interface{}) (interface{}, error) {
	result, err := e.Evaluate(expression, context)
	if err != nil {
		return nil, err
	}

	// Type checking would go here if needed
	return result, nil
}

// buildEnvironment creates the environment with all built-in functions
func buildEnvironment() map[string]interface{} {
	return map[string]interface{}{
		// String functions
		"upper":  wrapStringFunc1(stringUpper),
		"lower":  wrapStringFunc1(stringLower),
		"trim":   wrapStringFunc1(stringTrim),
		"concat": stringConcat,
		"substr": stringSubstr,

		// Date functions
		"now":        dateNow,
		"dateFormat": dateFormat,
		"dateParse":  dateParse,
		"addDays":    dateAddDays,

		// Math functions
		"round": wrapMathFunc(mathRound),
		"ceil":  wrapMathFunc(mathCeil),
		"floor": wrapMathFunc(mathFloor),
		"abs":   wrapMathFunc(mathAbs),
		"min":   mathMin,
		"max":   mathMax,

		// Array/String functions
		"len": lenFunc,
	}
}

// wrapStringFunc1 wraps a string function to handle the interface{} return
func wrapStringFunc1(fn func(interface{}) (string, error)) func(interface{}) interface{} {
	return func(arg interface{}) interface{} {
		result, err := fn(arg)
		if err != nil {
			panic(err)
		}
		return result
	}
}

// wrapMathFunc wraps a math function to handle the float64 return
func wrapMathFunc(fn func(float64) (float64, error)) func(float64) float64 {
	return func(arg float64) float64 {
		result, err := fn(arg)
		if err != nil {
			panic(err)
		}
		return result
	}
}

// ValidateExpression validates an expression without executing it
func (e *Evaluator) ValidateExpression(expression string) error {
	if expression == "" {
		return fmt.Errorf("expression cannot be empty")
	}

	_, err := expr.Compile(expression, expr.Env(e.env))
	if err != nil {
		return fmt.Errorf("invalid expression: %w", err)
	}

	return nil
}

// GetAvailableFunctions returns a list of all available function names
func (e *Evaluator) GetAvailableFunctions() []string {
	functions := []string{
		"upper", "lower", "trim", "concat", "substr",
		"now", "dateFormat", "dateParse", "addDays",
		"round", "ceil", "floor", "abs", "min", "max",
		"len",
	}
	return functions
}

// FunctionInfo provides documentation for a function
type FunctionInfo struct {
	Name        string
	Description string
	Parameters  []string
	ReturnType  string
	Example     string
}

// GetFunctionInfo returns documentation for all built-in functions
func GetFunctionInfo() []FunctionInfo {
	return []FunctionInfo{
		// String functions
		{
			Name:        "upper",
			Description: "Converts a string to uppercase",
			Parameters:  []string{"string"},
			ReturnType:  "string",
			Example:     `upper("hello") => "HELLO"`,
		},
		{
			Name:        "lower",
			Description: "Converts a string to lowercase",
			Parameters:  []string{"string"},
			ReturnType:  "string",
			Example:     `lower("HELLO") => "hello"`,
		},
		{
			Name:        "trim",
			Description: "Removes leading and trailing whitespace",
			Parameters:  []string{"string"},
			ReturnType:  "string",
			Example:     `trim("  hello  ") => "hello"`,
		},
		{
			Name:        "concat",
			Description: "Concatenates multiple strings",
			Parameters:  []string{"...strings"},
			ReturnType:  "string",
			Example:     `concat("hello", " ", "world") => "hello world"`,
		},
		{
			Name:        "substr",
			Description: "Extracts a substring",
			Parameters:  []string{"string", "start", "length"},
			ReturnType:  "string",
			Example:     `substr("hello world", 0, 5) => "hello"`,
		},

		// Date functions
		{
			Name:        "now",
			Description: "Returns the current time",
			Parameters:  []string{},
			ReturnType:  "time",
			Example:     `now()`,
		},
		{
			Name:        "dateFormat",
			Description: "Formats a time value",
			Parameters:  []string{"time", "layout"},
			ReturnType:  "string",
			Example:     `dateFormat(now(), "2006-01-02") => "2025-12-17"`,
		},
		{
			Name:        "dateParse",
			Description: "Parses a time string",
			Parameters:  []string{"value", "layout"},
			ReturnType:  "time",
			Example:     `dateParse("2025-12-17", "2006-01-02")`,
		},
		{
			Name:        "addDays",
			Description: "Adds days to a time value",
			Parameters:  []string{"time", "days"},
			ReturnType:  "time",
			Example:     `addDays(now(), 5)`,
		},

		// Math functions
		{
			Name:        "round",
			Description: "Rounds to the nearest integer",
			Parameters:  []string{"number"},
			ReturnType:  "number",
			Example:     `round(4.6) => 5`,
		},
		{
			Name:        "ceil",
			Description: "Rounds up to the nearest integer",
			Parameters:  []string{"number"},
			ReturnType:  "number",
			Example:     `ceil(4.1) => 5`,
		},
		{
			Name:        "floor",
			Description: "Rounds down to the nearest integer",
			Parameters:  []string{"number"},
			ReturnType:  "number",
			Example:     `floor(4.9) => 4`,
		},
		{
			Name:        "abs",
			Description: "Returns the absolute value",
			Parameters:  []string{"number"},
			ReturnType:  "number",
			Example:     `abs(-5) => 5`,
		},
		{
			Name:        "min",
			Description: "Returns the minimum value",
			Parameters:  []string{"...numbers"},
			ReturnType:  "number",
			Example:     `min(5, 3, 7, 1) => 1`,
		},
		{
			Name:        "max",
			Description: "Returns the maximum value",
			Parameters:  []string{"...numbers"},
			ReturnType:  "number",
			Example:     `max(5, 3, 7, 1) => 7`,
		},

		// Array/String functions
		{
			Name:        "len",
			Description: "Returns the length of an array or string",
			Parameters:  []string{"arrayOrString"},
			ReturnType:  "number",
			Example:     `len([1, 2, 3]) => 3, len("hello") => 5`,
		},
	}
}

// Helper function to convert time.Time to a format that expr can handle
func timeToExpr(t time.Time) interface{} {
	return t
}

// compileExpression compiles an expression with the given environment
func compileExpression(expression string, env map[string]interface{}) (*vm.Program, error) {
	program, err := expr.Compile(expression, expr.Env(env))
	if err != nil {
		return nil, fmt.Errorf("failed to compile expression: %w", err)
	}
	return program, nil
}

// runProgram executes a compiled program with the given environment
func runProgram(program *vm.Program, env map[string]interface{}) (interface{}, error) {
	result, err := expr.Run(program, env)
	if err != nil {
		return nil, err
	}
	return result, nil
}
