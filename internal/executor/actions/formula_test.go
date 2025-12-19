package actions

import (
	"context"
	"testing"
)

func TestFormulaAction_Execute_BasicExpression(t *testing.T) {
	execContext := map[string]interface{}{
		"x": 10,
		"y": 5,
	}

	action := &FormulaAction{}
	config := FormulaActionConfig{
		Expression: "x + y",
	}

	input := NewActionInput(config, execContext)
	output, err := action.Execute(context.Background(), input)

	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if output.Data != 15 {
		t.Errorf("Execute() result = %v, want 15", output.Data)
	}
}

func TestFormulaAction_Execute_StringFunctions(t *testing.T) {
	tests := []struct {
		name       string
		expression string
		context    map[string]interface{}
		want       interface{}
	}{
		{
			name:       "upper function",
			expression: `upper("hello")`,
			context:    map[string]interface{}{},
			want:       "HELLO",
		},
		{
			name:       "lower function",
			expression: `lower("WORLD")`,
			context:    map[string]interface{}{},
			want:       "world",
		},
		{
			name:       "concat function",
			expression: `concat("hello", " ", "world")`,
			context:    map[string]interface{}{},
			want:       "hello world",
		},
		{
			name:       "trim function",
			expression: `trim("  test  ")`,
			context:    map[string]interface{}{},
			want:       "test",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			action := &FormulaAction{}
			config := FormulaActionConfig{
				Expression: tt.expression,
			}

			input := NewActionInput(config, tt.context)
			output, err := action.Execute(context.Background(), input)

			if err != nil {
				t.Fatalf("Execute() error = %v", err)
			}

			if output.Data != tt.want {
				t.Errorf("Execute() result = %v, want %v", output.Data, tt.want)
			}
		})
	}
}

func TestFormulaAction_Execute_MathFunctions(t *testing.T) {
	tests := []struct {
		name       string
		expression string
		context    map[string]interface{}
		want       float64
	}{
		{
			name:       "round function",
			expression: "round(4.6)",
			context:    map[string]interface{}{},
			want:       5.0,
		},
		{
			name:       "ceil function",
			expression: "ceil(4.1)",
			context:    map[string]interface{}{},
			want:       5.0,
		},
		{
			name:       "floor function",
			expression: "floor(4.9)",
			context:    map[string]interface{}{},
			want:       4.0,
		},
		{
			name:       "abs function",
			expression: "abs(-5)",
			context:    map[string]interface{}{},
			want:       5.0,
		},
		{
			name:       "min function",
			expression: "min(5, 3, 7, 1)",
			context:    map[string]interface{}{},
			want:       1.0,
		},
		{
			name:       "max function",
			expression: "max(5, 3, 7, 1)",
			context:    map[string]interface{}{},
			want:       7.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			action := &FormulaAction{}
			config := FormulaActionConfig{
				Expression: tt.expression,
			}

			input := NewActionInput(config, tt.context)
			output, err := action.Execute(context.Background(), input)

			if err != nil {
				t.Fatalf("Execute() error = %v", err)
			}

			if output.Data != tt.want {
				t.Errorf("Execute() result = %v, want %v", output.Data, tt.want)
			}
		})
	}
}

func TestFormulaAction_Execute_WithExecutionContext(t *testing.T) {
	execContext := map[string]interface{}{
		"steps": map[string]interface{}{
			"http-1": map[string]interface{}{
				"body": map[string]interface{}{
					"users": []interface{}{
						map[string]interface{}{
							"name": "John Doe",
							"age":  30,
						},
						map[string]interface{}{
							"name": "Jane Smith",
							"age":  25,
						},
					},
				},
			},
		},
	}

	tests := []struct {
		name       string
		expression string
		want       interface{}
	}{
		{
			name:       "access nested field",
			expression: "steps['http-1'].body.users[0].name",
			want:       "John Doe",
		},
		{
			name:       "transform with function",
			expression: "upper(steps['http-1'].body.users[1].name)",
			want:       "JANE SMITH",
		},
		{
			name:       "array length",
			expression: "len(steps['http-1'].body.users)",
			want:       2,
		},
		{
			name:       "conditional expression",
			expression: "steps['http-1'].body.users[0].age > 25 ? 'senior' : 'junior'",
			want:       "senior",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			action := &FormulaAction{}
			config := FormulaActionConfig{
				Expression: tt.expression,
			}

			input := NewActionInput(config, execContext)
			output, err := action.Execute(context.Background(), input)

			if err != nil {
				t.Fatalf("Execute() error = %v", err)
			}

			if output.Data != tt.want {
				t.Errorf("Execute() result = %v, want %v", output.Data, tt.want)
			}
		})
	}
}

func TestFormulaAction_Execute_ComplexExpressions(t *testing.T) {
	execContext := map[string]interface{}{
		"price":    10.99,
		"quantity": 3,
		"taxRate":  0.08,
		"customer": map[string]interface{}{
			"firstName": "john",
			"lastName":  "Doe",
		},
	}

	tests := []struct {
		name       string
		expression string
		want       interface{}
	}{
		{
			name:       "calculate total with tax",
			expression: "round((price * quantity) * (1 + taxRate))",
			want:       36.0,
		},
		{
			name:       "format customer name",
			expression: `concat(upper(substr(customer.firstName, 0, 1)), ". ", customer.lastName)`,
			want:       "J. Doe",
		},
		{
			name:       "nested functions",
			expression: "upper(trim(concat('  ', customer.firstName, '  ')))",
			want:       "JOHN",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			action := &FormulaAction{}
			config := FormulaActionConfig{
				Expression: tt.expression,
			}

			input := NewActionInput(config, execContext)
			output, err := action.Execute(context.Background(), input)

			if err != nil {
				t.Fatalf("Execute() error = %v", err)
			}

			if output.Data != tt.want {
				t.Errorf("Execute() result = %v, want %v", output.Data, tt.want)
			}
		})
	}
}

func TestFormulaAction_Execute_ErrorCases(t *testing.T) {
	tests := []struct {
		name       string
		expression string
		context    map[string]interface{}
		wantErr    bool
	}{
		{
			name:       "empty expression",
			expression: "",
			context:    map[string]interface{}{},
			wantErr:    true,
		},
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
			name:       "invalid function call",
			expression: "undefinedFunction(5)",
			context:    map[string]interface{}{},
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			action := &FormulaAction{}
			config := FormulaActionConfig{
				Expression: tt.expression,
			}

			input := NewActionInput(config, tt.context)
			_, err := action.Execute(context.Background(), input)

			if (err != nil) != tt.wantErr {
				t.Errorf("Execute() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestFormulaAction_Execute_InvalidConfig(t *testing.T) {
	action := &FormulaAction{}

	// Test with invalid config type
	input := NewActionInput("invalid config", map[string]interface{}{})
	_, err := action.Execute(context.Background(), input)

	if err == nil {
		t.Error("Execute() should return error for invalid config")
	}
}
