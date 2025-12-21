package expression

import (
	"testing"
)

func TestParser_Parse(t *testing.T) {
	parser := NewParser()

	tests := []struct {
		name         string
		expr         string
		wantTemplate bool
		wantContent  string
		wantErr      bool
	}{
		{
			name:         "template syntax",
			expr:         "{{steps.step1.status}} == \"success\"",
			wantTemplate: true,
			wantContent:  "steps.step1.status}} == \"success",
			wantErr:      false,
		},
		{
			name:         "raw expression",
			expr:         "steps.step1.output.count > 10",
			wantTemplate: false,
			wantContent:  "steps.step1.output.count > 10",
			wantErr:      false,
		},
		{
			name:    "empty expression",
			expr:    "",
			wantErr: true,
		},
		{
			name:         "complex condition",
			expr:         "{{steps.http.status}} == 200 && {{trigger.body.type}} == \"webhook\"",
			wantTemplate: true,
			wantContent:  "steps.http.status}} == 200 && {{trigger.body.type}} == \"webhook",
			wantErr:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parser.Parse(tt.expr)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err == nil {
				if got.IsTemplate != tt.wantTemplate {
					t.Errorf("Parse() IsTemplate = %v, want %v", got.IsTemplate, tt.wantTemplate)
				}
			}
		})
	}
}

func TestParser_ExtractPaths(t *testing.T) {
	parser := NewParser()

	tests := []struct {
		name      string
		expr      string
		wantPaths []string
	}{
		{
			name:      "single path",
			expr:      "{{steps.step1.output}}",
			wantPaths: []string{"steps.step1.output"},
		},
		{
			name:      "multiple paths",
			expr:      "{{steps.step1.status}} == \"success\" && {{trigger.body.count}} > 10",
			wantPaths: []string{"steps.step1.status", "trigger.body.count"},
		},
		{
			name:      "path with array index",
			expr:      "{{steps.step1.output[0].name}}",
			wantPaths: []string{"steps.step1.output[0].name"},
		},
		{
			name:      "env variable",
			expr:      "{{env.tenant_id}} == \"test\"",
			wantPaths: []string{"env.tenant_id"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parser.ExtractPaths(tt.expr)
			if len(got) != len(tt.wantPaths) {
				t.Errorf("ExtractPaths() = %v, want %v", got, tt.wantPaths)
			}
		})
	}
}

func TestParser_ValidateExpression(t *testing.T) {
	parser := NewParser()

	tests := []struct {
		name    string
		expr    string
		wantErr bool
	}{
		{
			name:    "valid expression",
			expr:    "{{steps.step1.status}} == \"success\"",
			wantErr: false,
		},
		{
			name:    "unbalanced parentheses",
			expr:    "((steps.step1.count > 10)",
			wantErr: true,
		},
		{
			name:    "unbalanced brackets",
			expr:    "steps.step1.output[0",
			wantErr: true,
		},
		{
			name:    "empty expression",
			expr:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := parser.ValidateExpression(tt.expr)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateExpression() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGetValueByPath(t *testing.T) {
	data := map[string]interface{}{
		"steps": map[string]interface{}{
			"step1": map[string]interface{}{
				"status": "success",
				"output": map[string]interface{}{
					"count": 42,
					"users": []interface{}{
						map[string]interface{}{"name": "Alice", "age": 30},
						map[string]interface{}{"name": "Bob", "age": 25},
					},
				},
			},
		},
		"trigger": map[string]interface{}{
			"body": map[string]interface{}{
				"type": "webhook",
			},
		},
	}

	tests := []struct {
		name    string
		path    string
		want    interface{}
		wantErr bool
	}{
		{
			name:    "simple path",
			path:    "steps.step1.status",
			want:    "success",
			wantErr: false,
		},
		{
			name:    "nested path",
			path:    "steps.step1.output.count",
			want:    42,
			wantErr: false,
		},
		{
			name:    "array access",
			path:    "steps.step1.output.users[0].name",
			want:    "Alice",
			wantErr: false,
		},
		{
			name:    "array access second element",
			path:    "steps.step1.output.users[1].age",
			want:    25,
			wantErr: false,
		},
		{
			name:    "non-existent path",
			path:    "steps.step2.status",
			want:    nil,
			wantErr: true,
		},
		{
			name:    "out of bounds array",
			path:    "steps.step1.output.users[5].name",
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetValueByPath(data, tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetValueByPath() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("GetValueByPath() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestResolveTemplateVariables(t *testing.T) {
	context := map[string]interface{}{
		"steps": map[string]interface{}{
			"step1": map[string]interface{}{
				"status": "success",
				"count":  42,
			},
		},
		"trigger": map[string]interface{}{
			"body": map[string]interface{}{
				"user": "Alice",
			},
		},
		"env": map[string]interface{}{
			"tenant_id": "tenant-123",
		},
	}

	tests := []struct {
		name     string
		template string
		want     string
		wantErr  bool
	}{
		{
			name:     "single variable",
			template: "Status is {{steps.step1.status}}",
			want:     "Status is success",
			wantErr:  false,
		},
		{
			name:     "multiple variables",
			template: "User {{trigger.body.user}} has count {{steps.step1.count}}",
			want:     "User Alice has count 42",
			wantErr:  false,
		},
		{
			name:     "env variable",
			template: "Tenant: {{env.tenant_id}}",
			want:     "Tenant: tenant-123",
			wantErr:  false,
		},
		{
			name:     "no variables",
			template: "Plain text",
			want:     "Plain text",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ResolveTemplateVariables(tt.template, context)
			if (err != nil) != tt.wantErr {
				t.Errorf("ResolveTemplateVariables() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ResolveTemplateVariables() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBuildContext(t *testing.T) {
	trigger := map[string]interface{}{"type": "webhook"}
	steps := map[string]interface{}{"step1": map[string]interface{}{"status": "success"}}
	env := map[string]interface{}{"tenant_id": "test"}

	ctx := BuildContext(trigger, steps, env)

	if ctx["trigger"] == nil {
		t.Error("BuildContext() trigger is nil")
	}
	if ctx["steps"] == nil {
		t.Error("BuildContext() steps is nil")
	}
	if ctx["env"] == nil {
		t.Error("BuildContext() env is nil")
	}

	// Test with nil inputs
	ctx2 := BuildContext(nil, nil, nil)
	if ctx2["trigger"] == nil {
		t.Error("BuildContext() should create empty trigger map")
	}
	if ctx2["steps"] == nil {
		t.Error("BuildContext() should create empty steps map")
	}
	if ctx2["env"] == nil {
		t.Error("BuildContext() should create empty env map")
	}
}
