package expression

import (
	"testing"
)

func TestEvaluator_EvaluateCondition(t *testing.T) {
	evaluator := NewEvaluator()

	context := map[string]interface{}{
		"steps": map[string]interface{}{
			"step1": map[string]interface{}{
				"status": "success",
				"output": map[string]interface{}{
					"count": 42,
					"value": 100,
				},
			},
			"step2": map[string]interface{}{
				"status": "failed",
			},
		},
		"trigger": map[string]interface{}{
			"body": map[string]interface{}{
				"type":   "webhook",
				"action": "create",
			},
		},
		"env": map[string]interface{}{
			"tenant_id": "tenant-123",
		},
	}

	tests := []struct {
		name       string
		expression string
		want       bool
		wantErr    bool
	}{
		{
			name:       "simple equality",
			expression: "steps.step1.status == \"success\"",
			want:       true,
			wantErr:    false,
		},
		{
			name:       "simple inequality",
			expression: "steps.step2.status != \"success\"",
			want:       true,
			wantErr:    false,
		},
		{
			name:       "greater than",
			expression: "steps.step1.output.count > 10",
			want:       true,
			wantErr:    false,
		},
		{
			name:       "less than",
			expression: "steps.step1.output.count < 50",
			want:       true,
			wantErr:    false,
		},
		{
			name:       "greater or equal",
			expression: "steps.step1.output.count >= 42",
			want:       true,
			wantErr:    false,
		},
		{
			name:       "less or equal",
			expression: "steps.step1.output.count <= 42",
			want:       true,
			wantErr:    false,
		},
		{
			name:       "logical AND true",
			expression: "steps.step1.status == \"success\" && steps.step1.output.count > 10",
			want:       true,
			wantErr:    false,
		},
		{
			name:       "logical AND false",
			expression: "steps.step1.status == \"success\" && steps.step1.output.count > 100",
			want:       false,
			wantErr:    false,
		},
		{
			name:       "logical OR true",
			expression: "steps.step1.status == \"success\" || steps.step2.status == \"success\"",
			want:       true,
			wantErr:    false,
		},
		{
			name:       "logical OR false",
			expression: "steps.step1.status == \"failed\" || steps.step2.status == \"success\"",
			want:       false,
			wantErr:    false,
		},
		{
			name:       "logical NOT",
			expression: "!(steps.step1.status == \"failed\")",
			want:       true,
			wantErr:    false,
		},
		{
			name:       "complex condition",
			expression: "(steps.step1.status == \"success\" && steps.step1.output.count > 10) || steps.step2.status == \"success\"",
			want:       true,
			wantErr:    false,
		},
		{
			name:       "trigger data access",
			expression: "trigger.body.type == \"webhook\"",
			want:       true,
			wantErr:    false,
		},
		{
			name:       "env variable access",
			expression: "env.tenant_id == \"tenant-123\"",
			want:       true,
			wantErr:    false,
		},
		{
			name:       "empty expression",
			expression: "",
			want:       false,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := evaluator.EvaluateCondition(tt.expression, context)
			if (err != nil) != tt.wantErr {
				t.Errorf("EvaluateCondition() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("EvaluateCondition() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEvaluator_EvaluateWithTemplate(t *testing.T) {
	evaluator := NewEvaluator()

	context := map[string]interface{}{
		"steps": map[string]interface{}{
			"step1": map[string]interface{}{
				"status": "success",
				"output": map[string]interface{}{
					"count": 42,
				},
			},
		},
	}

	tests := []struct {
		name       string
		expression string
		want       bool
		wantErr    bool
	}{
		{
			name:       "template syntax wrapping entire expression",
			expression: "{{steps.step1.status == \"success\"}}",
			want:       true,
			wantErr:    false,
		},
		{
			name:       "template syntax with comparison wrapped",
			expression: "{{steps.step1.output.count > 10}}",
			want:       true,
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := evaluator.EvaluateCondition(tt.expression, context)
			if (err != nil) != tt.wantErr {
				t.Errorf("EvaluateCondition() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("EvaluateCondition() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEvaluator_Evaluate(t *testing.T) {
	evaluator := NewEvaluator()

	context := map[string]interface{}{
		"steps": map[string]interface{}{
			"step1": map[string]interface{}{
				"output": map[string]interface{}{
					"count": 42,
					"name":  "test",
				},
			},
		},
	}

	tests := []struct {
		name       string
		expression string
		want       interface{}
		wantErr    bool
	}{
		{
			name:       "return number",
			expression: "steps.step1.output.count",
			want:       42,
			wantErr:    false,
		},
		{
			name:       "return string",
			expression: "steps.step1.output.name",
			want:       "test",
			wantErr:    false,
		},
		{
			name:       "arithmetic",
			expression: "steps.step1.output.count * 2",
			want:       84,
			wantErr:    false,
		},
		{
			name:       "boolean result",
			expression: "steps.step1.output.count > 10",
			want:       true,
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := evaluator.Evaluate(tt.expression, context)
			if (err != nil) != tt.wantErr {
				t.Errorf("Evaluate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				// For numeric comparisons, handle type differences
				switch v := tt.want.(type) {
				case int:
					gotFloat, ok := got.(float64)
					if ok && int(gotFloat) != v {
						t.Errorf("Evaluate() = %v, want %v", got, tt.want)
					}
				default:
					if got != tt.want {
						t.Errorf("Evaluate() = %v, want %v", got, tt.want)
					}
				}
			}
		})
	}
}

func TestEvaluator_ValidateCondition(t *testing.T) {
	evaluator := NewEvaluator()

	tests := []struct {
		name    string
		expr    string
		wantErr bool
	}{
		{
			name:    "valid simple condition",
			expr:    "steps.step1.status == \"success\"",
			wantErr: false,
		},
		{
			name:    "valid complex condition",
			expr:    "(steps.step1.count > 10 && trigger.body.type == \"webhook\") || env.tenant_id == \"test\"",
			wantErr: false,
		},
		{
			name:    "empty expression",
			expr:    "",
			wantErr: true,
		},
		{
			name:    "invalid syntax",
			expr:    "steps.step1.status ==",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := evaluator.ValidateCondition(tt.expr)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateCondition() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestEvaluator_EvaluateBooleanExpression(t *testing.T) {
	evaluator := NewEvaluator()

	tests := []struct {
		name     string
		left     interface{}
		operator string
		right    interface{}
		want     bool
		wantErr  bool
	}{
		{
			name:     "equals string",
			left:     "success",
			operator: "==",
			right:    "success",
			want:     true,
			wantErr:  false,
		},
		{
			name:     "not equals",
			left:     "success",
			operator: "!=",
			right:    "failed",
			want:     true,
			wantErr:  false,
		},
		{
			name:     "greater than",
			left:     42,
			operator: ">",
			right:    10,
			want:     true,
			wantErr:  false,
		},
		{
			name:     "less than",
			left:     10,
			operator: "<",
			right:    42,
			want:     true,
			wantErr:  false,
		},
		{
			name:     "greater or equal",
			left:     42,
			operator: ">=",
			right:    42,
			want:     true,
			wantErr:  false,
		},
		{
			name:     "less or equal",
			left:     10,
			operator: "<=",
			right:    42,
			want:     true,
			wantErr:  false,
		},
		{
			name:     "contains",
			left:     "hello world",
			operator: "contains",
			right:    "world",
			want:     true,
			wantErr:  false,
		},
		{
			name:     "starts with",
			left:     "hello world",
			operator: "starts_with",
			right:    "hello",
			want:     true,
			wantErr:  false,
		},
		{
			name:     "ends with",
			left:     "hello world",
			operator: "ends_with",
			right:    "world",
			want:     true,
			wantErr:  false,
		},
		{
			name:     "unsupported operator",
			left:     "test",
			operator: "invalid",
			right:    "test",
			want:     false,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := evaluator.EvaluateBooleanExpression(tt.left, tt.operator, tt.right)
			if (err != nil) != tt.wantErr {
				t.Errorf("EvaluateBooleanExpression() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("EvaluateBooleanExpression() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEvaluator_CompileAndRun(t *testing.T) {
	evaluator := NewEvaluator()

	context := map[string]interface{}{
		"steps": map[string]interface{}{
			"step1": map[string]interface{}{
				"count": 42,
			},
		},
	}

	// Compile expression
	program, err := evaluator.CompileExpression("steps.step1.count > 10", context)
	if err != nil {
		t.Fatalf("CompileExpression() error = %v", err)
	}

	// Run compiled expression multiple times
	for i := 0; i < 3; i++ {
		result, err := evaluator.EvaluateWithProgram(program, context)
		if err != nil {
			t.Errorf("EvaluateWithProgram() iteration %d error = %v", i, err)
		}

		boolResult, ok := result.(bool)
		if !ok {
			t.Errorf("Expected bool result, got %T", result)
		}

		if !boolResult {
			t.Errorf("Expected true, got false")
		}
	}
}
