package formula

import (
	"testing"
	"time"
)

// TestEvaluatorBasicExpressions tests basic expression evaluation
func TestEvaluatorBasicExpressions(t *testing.T) {
	evaluator := NewEvaluator()

	tests := []struct {
		name       string
		expression string
		context    map[string]interface{}
		want       interface{}
		wantErr    bool
	}{
		{
			name:       "simple arithmetic",
			expression: "1 + 2 + 3",
			context:    map[string]interface{}{},
			want:       6,
			wantErr:    false,
		},
		{
			name:       "multiplication and division",
			expression: "10 * 2 / 5",
			context:    map[string]interface{}{},
			want:       4.0,
			wantErr:    false,
		},
		{
			name:       "boolean expression",
			expression: "true && false",
			context:    map[string]interface{}{},
			want:       false,
			wantErr:    false,
		},
		{
			name:       "comparison",
			expression: "5 > 3",
			context:    map[string]interface{}{},
			want:       true,
			wantErr:    false,
		},
		{
			name:       "string concatenation with operator",
			expression: `"hello" + " " + "world"`,
			context:    map[string]interface{}{},
			want:       "hello world",
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := evaluator.Evaluate(tt.expression, tt.context)
			if (err != nil) != tt.wantErr {
				t.Errorf("Evaluate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && result != tt.want {
				t.Errorf("Evaluate() = %v (%T), want %v (%T)", result, result, tt.want, tt.want)
			}
		})
	}
}

// TestEvaluatorVariableSubstitution tests variable reference in expressions
func TestEvaluatorVariableSubstitution(t *testing.T) {
	evaluator := NewEvaluator()

	tests := []struct {
		name       string
		expression string
		context    map[string]interface{}
		want       interface{}
		wantErr    bool
	}{
		{
			name:       "simple variable reference",
			expression: "x + 5",
			context:    map[string]interface{}{"x": 10},
			want:       15,
			wantErr:    false,
		},
		{
			name:       "nested object access",
			expression: "user.age",
			context: map[string]interface{}{
				"user": map[string]interface{}{
					"age": 25,
				},
			},
			want:    25,
			wantErr: false,
		},
		{
			name:       "array access",
			expression: "items[0]",
			context: map[string]interface{}{
				"items": []interface{}{"first", "second", "third"},
			},
			want:    "first",
			wantErr: false,
		},
		{
			name:       "complex nested access",
			expression: "steps.http1.body.users[0].name",
			context: map[string]interface{}{
				"steps": map[string]interface{}{
					"http1": map[string]interface{}{
						"body": map[string]interface{}{
							"users": []interface{}{
								map[string]interface{}{
									"name": "John Doe",
								},
							},
						},
					},
				},
			},
			want:    "John Doe",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := evaluator.Evaluate(tt.expression, tt.context)
			if (err != nil) != tt.wantErr {
				t.Errorf("Evaluate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && result != tt.want {
				t.Errorf("Evaluate() = %v, want %v", result, tt.want)
			}
		})
	}
}

// TestEvaluatorStringFunctions tests string function calls
func TestEvaluatorStringFunctions(t *testing.T) {
	evaluator := NewEvaluator()

	tests := []struct {
		name       string
		expression string
		context    map[string]interface{}
		want       interface{}
		wantErr    bool
	}{
		{
			name:       "upper function",
			expression: `upper("hello world")`,
			context:    map[string]interface{}{},
			want:       "HELLO WORLD",
			wantErr:    false,
		},
		{
			name:       "lower function",
			expression: `lower("HELLO WORLD")`,
			context:    map[string]interface{}{},
			want:       "hello world",
			wantErr:    false,
		},
		{
			name:       "trim function",
			expression: `trim("  hello  ")`,
			context:    map[string]interface{}{},
			want:       "hello",
			wantErr:    false,
		},
		{
			name:       "concat function",
			expression: `concat("hello", " ", "world")`,
			context:    map[string]interface{}{},
			want:       "hello world",
			wantErr:    false,
		},
		{
			name:       "substr function",
			expression: `substr("hello world", 0, 5)`,
			context:    map[string]interface{}{},
			want:       "hello",
			wantErr:    false,
		},
		{
			name:       "chained string functions",
			expression: `upper(trim("  hello  "))`,
			context:    map[string]interface{}{},
			want:       "HELLO",
			wantErr:    false,
		},
		{
			name:       "string function with variable",
			expression: `upper(name)`,
			context:    map[string]interface{}{"name": "john"},
			want:       "JOHN",
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := evaluator.Evaluate(tt.expression, tt.context)
			if (err != nil) != tt.wantErr {
				t.Errorf("Evaluate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && result != tt.want {
				t.Errorf("Evaluate() = %v, want %v", result, tt.want)
			}
		})
	}
}

// TestEvaluatorMathFunctions tests mathematical function calls
func TestEvaluatorMathFunctions(t *testing.T) {
	evaluator := NewEvaluator()

	tests := []struct {
		name       string
		expression string
		context    map[string]interface{}
		want       interface{}
		wantErr    bool
	}{
		{
			name:       "round function",
			expression: "round(4.6)",
			context:    map[string]interface{}{},
			want:       5.0,
			wantErr:    false,
		},
		{
			name:       "ceil function",
			expression: "ceil(4.1)",
			context:    map[string]interface{}{},
			want:       5.0,
			wantErr:    false,
		},
		{
			name:       "floor function",
			expression: "floor(4.9)",
			context:    map[string]interface{}{},
			want:       4.0,
			wantErr:    false,
		},
		{
			name:       "abs function",
			expression: "abs(-5)",
			context:    map[string]interface{}{},
			want:       5.0,
			wantErr:    false,
		},
		{
			name:       "min function",
			expression: "min(5, 3, 7, 1)",
			context:    map[string]interface{}{},
			want:       1.0,
			wantErr:    false,
		},
		{
			name:       "max function",
			expression: "max(5, 3, 7, 1)",
			context:    map[string]interface{}{},
			want:       7.0,
			wantErr:    false,
		},
		{
			name:       "nested math functions",
			expression: "round(abs(-4.6))",
			context:    map[string]interface{}{},
			want:       5.0,
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := evaluator.Evaluate(tt.expression, tt.context)
			if (err != nil) != tt.wantErr {
				t.Errorf("Evaluate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && result != tt.want {
				t.Errorf("Evaluate() = %v, want %v", result, tt.want)
			}
		})
	}
}

// TestEvaluatorDateFunctions tests date function calls
func TestEvaluatorDateFunctions(t *testing.T) {
	evaluator := NewEvaluator()

	t.Run("now function returns current time", func(t *testing.T) {
		before := time.Now()
		result, err := evaluator.Evaluate("now()", map[string]interface{}{})
		after := time.Now()

		if err != nil {
			t.Errorf("Evaluate() error = %v", err)
			return
		}

		resultTime, ok := result.(time.Time)
		if !ok {
			t.Errorf("Expected time.Time, got %T", result)
			return
		}

		if resultTime.Before(before) || resultTime.After(after) {
			t.Errorf("now() returned time outside expected range")
		}
	})

	t.Run("dateFormat function", func(t *testing.T) {
		fixedTime := time.Date(2025, 12, 17, 10, 30, 0, 0, time.UTC)
		context := map[string]interface{}{
			"myTime": fixedTime,
		}

		result, err := evaluator.Evaluate(`dateFormat(myTime, "2006-01-02")`, context)
		if err != nil {
			t.Errorf("Evaluate() error = %v", err)
			return
		}

		want := "2025-12-17"
		if result != want {
			t.Errorf("dateFormat() = %v, want %v", result, want)
		}
	})

	t.Run("dateParse function", func(t *testing.T) {
		result, err := evaluator.Evaluate(`dateParse("2025-12-17T10:30:00Z", "2006-01-02T15:04:05Z07:00")`, map[string]interface{}{})
		if err != nil {
			t.Errorf("Evaluate() error = %v", err)
			return
		}

		resultTime, ok := result.(time.Time)
		if !ok {
			t.Errorf("Expected time.Time, got %T", result)
			return
		}

		expectedTime := time.Date(2025, 12, 17, 10, 30, 0, 0, time.UTC)
		if !resultTime.Equal(expectedTime) {
			t.Errorf("dateParse() = %v, want %v", resultTime, expectedTime)
		}
	})

	t.Run("addDays function", func(t *testing.T) {
		fixedTime := time.Date(2025, 12, 17, 10, 30, 0, 0, time.UTC)
		context := map[string]interface{}{
			"myTime": fixedTime,
		}

		result, err := evaluator.Evaluate("addDays(myTime, 5)", context)
		if err != nil {
			t.Errorf("Evaluate() error = %v", err)
			return
		}

		resultTime, ok := result.(time.Time)
		if !ok {
			t.Errorf("Expected time.Time, got %T", result)
			return
		}

		expectedTime := fixedTime.AddDate(0, 0, 5)
		if !resultTime.Equal(expectedTime) {
			t.Errorf("addDays() = %v, want %v", resultTime, expectedTime)
		}
	})
}

// TestEvaluatorArrayFunctions tests array function calls
func TestEvaluatorArrayFunctions(t *testing.T) {
	evaluator := NewEvaluator()

	t.Run("len function with array", func(t *testing.T) {
		context := map[string]interface{}{
			"items": []interface{}{1, 2, 3, 4, 5},
		}

		result, err := evaluator.Evaluate("len(items)", context)
		if err != nil {
			t.Errorf("Evaluate() error = %v", err)
			return
		}

		if result != 5 {
			t.Errorf("len() = %v, want 5", result)
		}
	})

	t.Run("len with empty array", func(t *testing.T) {
		context := map[string]interface{}{
			"items": []interface{}{},
		}

		result, err := evaluator.Evaluate("len(items)", context)
		if err != nil {
			t.Errorf("Evaluate() error = %v", err)
			return
		}

		if result != 0 {
			t.Errorf("len() = %v, want 0", result)
		}
	})

	t.Run("len with string", func(t *testing.T) {
		context := map[string]interface{}{
			"name": "Alexander",
		}

		result, err := evaluator.Evaluate("len(name)", context)
		if err != nil {
			t.Errorf("Evaluate() error = %v", err)
			return
		}

		if result != 9 {
			t.Errorf("len() = %v, want 9", result)
		}
	})
}

// TestEvaluatorComplexExpressions tests complex real-world expressions
func TestEvaluatorComplexExpressions(t *testing.T) {
	evaluator := NewEvaluator()

	tests := []struct {
		name       string
		expression string
		context    map[string]interface{}
		want       interface{}
		wantErr    bool
	}{
		{
			name:       "conditional with string function",
			expression: `len(name) > 5 ? upper(name) : lower(name)`,
			context:    map[string]interface{}{"name": "Alexander"},
			want:       "ALEXANDER",
			wantErr:    false,
		},
		{
			name:       "math with variables",
			expression: `round((price * quantity) * (1 + taxRate))`,
			context: map[string]interface{}{
				"price":    10.99,
				"quantity": 3,
				"taxRate":  0.08,
			},
			want:    36.0,
			wantErr: false,
		},
		{
			name:       "string manipulation pipeline",
			expression: `concat(upper(substr(firstName, 0, 1)), ". ", lastName)`,
			context: map[string]interface{}{
				"firstName": "john",
				"lastName":  "Doe",
			},
			want:    "J. Doe",
			wantErr: false,
		},
		{
			name:       "nested object with function",
			expression: `upper(steps.http1.body.status)`,
			context: map[string]interface{}{
				"steps": map[string]interface{}{
					"http1": map[string]interface{}{
						"body": map[string]interface{}{
							"status": "success",
						},
					},
				},
			},
			want:    "SUCCESS",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := evaluator.Evaluate(tt.expression, tt.context)
			if (err != nil) != tt.wantErr {
				t.Errorf("Evaluate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && result != tt.want {
				t.Errorf("Evaluate() = %v, want %v", result, tt.want)
			}
		})
	}
}

// TestEvaluatorErrorCases tests error handling
func TestEvaluatorErrorCases(t *testing.T) {
	evaluator := NewEvaluator()

	tests := []struct {
		name       string
		expression string
		context    map[string]interface{}
		wantErr    bool
	}{
		{
			name:       "invalid syntax",
			expression: "1 + * 2",
			context:    map[string]interface{}{},
			wantErr:    true,
		},
		{
			name:       "undefined variable",
			expression: "undefinedVar + 5",
			context:    map[string]interface{}{},
			wantErr:    true,
		},
		{
			name:       "type mismatch",
			expression: `"string" - 5`,
			context:    map[string]interface{}{},
			wantErr:    true,
		},
		{
			name:       "undefined function",
			expression: "undefinedFunction(5)",
			context:    map[string]interface{}{},
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := evaluator.Evaluate(tt.expression, tt.context)
			if (err != nil) != tt.wantErr {
				t.Errorf("Evaluate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
