package actions

import (
	"context"
	"testing"
)

func testInterfaceEqual(t *testing.T, got, want interface{}) bool {
	t.Helper()
	// Handle float64 vs int comparison
	switch g := got.(type) {
	case float64:
		switch w := want.(type) {
		case int:
			return g == float64(w)
		case float64:
			return g == w
		}
	case int:
		switch w := want.(type) {
		case int:
			return g == w
		case float64:
			return float64(g) == w
		}
	}
	return got == want
}

func TestTransformAction_Execute_WithExpression(t *testing.T) {
	execContext := map[string]interface{}{
		"steps": map[string]interface{}{
			"http-1": map[string]interface{}{
				"body": map[string]interface{}{
					"user": map[string]interface{}{
						"name": "Alice",
						"id":   123,
					},
				},
			},
		},
	}

	action := &TransformAction{}
	config := TransformActionConfig{
		Expression: "steps.http-1.body.user",
	}

	input := NewActionInput(config, execContext)
	output, err := action.Execute(context.Background(), input)

	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	result, ok := output.Data.(map[string]interface{})
	if !ok {
		t.Fatal("Output data is not a map")
	}

	if result["name"] != "Alice" {
		t.Errorf("name = %v, want 'Alice'", result["name"])
	}

	if !testInterfaceEqual(t, result["id"], 123) {
		t.Errorf("id = %v (%T), want 123", result["id"], result["id"])
	}
}

func TestTransformAction_Execute_WithMapping(t *testing.T) {
	execContext := map[string]interface{}{
		"trigger": map[string]interface{}{
			"user": map[string]interface{}{
				"first_name": "John",
				"last_name":  "Doe",
				"email":      "john@example.com",
			},
		},
		"steps": map[string]interface{}{
			"http-1": map[string]interface{}{
				"body": map[string]interface{}{
					"id": 456,
				},
			},
		},
	}

	action := &TransformAction{}
	config := TransformActionConfig{
		Mapping: map[string]string{
			"user_id":    "steps.http-1.body.id",
			"first_name": "trigger.user.first_name",
			"last_name":  "trigger.user.last_name",
			"email":      "trigger.user.email",
		},
	}

	input := NewActionInput(config, execContext)
	output, err := action.Execute(context.Background(), input)

	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	result, ok := output.Data.(map[string]interface{})
	if !ok {
		t.Fatal("Output data is not a map")
	}

	if !testInterfaceEqual(t, result["user_id"], 456) {
		t.Errorf("user_id = %v (%T), want 456", result["user_id"], result["user_id"])
	}

	if result["first_name"] != "John" {
		t.Errorf("first_name = %v, want 'John'", result["first_name"])
	}

	if result["last_name"] != "Doe" {
		t.Errorf("last_name = %v, want 'Doe'", result["last_name"])
	}

	if result["email"] != "john@example.com" {
		t.Errorf("email = %v, want 'john@example.com'", result["email"])
	}
}

func TestTransformAction_Execute_WithArrayMapping(t *testing.T) {
	execContext := map[string]interface{}{
		"steps": map[string]interface{}{
			"http-1": map[string]interface{}{
				"body": map[string]interface{}{
					"users": []interface{}{
						map[string]interface{}{"name": "Alice", "age": 30},
						map[string]interface{}{"name": "Bob", "age": 25},
					},
				},
			},
		},
	}

	action := &TransformAction{}
	config := TransformActionConfig{
		Mapping: map[string]string{
			"first_user":  "steps.http-1.body.users[0].name",
			"first_age":   "steps.http-1.body.users[0].age",
			"second_user": "steps.http-1.body.users[1].name",
		},
	}

	input := NewActionInput(config, execContext)
	output, err := action.Execute(context.Background(), input)

	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	result, ok := output.Data.(map[string]interface{})
	if !ok {
		t.Fatal("Output data is not a map")
	}

	if result["first_user"] != "Alice" {
		t.Errorf("first_user = %v, want 'Alice'", result["first_user"])
	}

	if !testInterfaceEqual(t, result["first_age"], 30) {
		t.Errorf("first_age = %v (%T), want 30", result["first_age"], result["first_age"])
	}

	if result["second_user"] != "Bob" {
		t.Errorf("second_user = %v, want 'Bob'", result["second_user"])
	}
}

func TestTransformAction_Execute_MissingPath(t *testing.T) {
	execContext := map[string]interface{}{
		"trigger": map[string]interface{}{
			"data": "test",
		},
	}

	action := &TransformAction{}
	config := TransformActionConfig{
		Mapping: map[string]string{
			"existing": "trigger.data",
			"missing":  "trigger.nonexistent",
		},
	}

	input := NewActionInput(config, execContext)
	output, err := action.Execute(context.Background(), input)

	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	result, ok := output.Data.(map[string]interface{})
	if !ok {
		t.Fatal("Output data is not a map")
	}

	if result["existing"] != "test" {
		t.Errorf("existing = %v, want 'test'", result["existing"])
	}

	// Missing paths should result in nil
	if result["missing"] != nil {
		t.Errorf("missing = %v, want nil", result["missing"])
	}
}

func TestTransformAction_Execute_InvalidExpression(t *testing.T) {
	execContext := map[string]interface{}{
		"trigger": map[string]interface{}{
			"data": "test",
		},
	}

	action := &TransformAction{}
	config := TransformActionConfig{
		Expression: "trigger.nonexistent.path",
	}

	input := NewActionInput(config, execContext)
	_, err := action.Execute(context.Background(), input)

	if err == nil {
		t.Error("Expected error for invalid expression, got nil")
	}
}

func TestTransformAction_Execute_WithDefault(t *testing.T) {
	execContext := map[string]interface{}{
		"trigger": map[string]interface{}{
			"data": "test",
		},
	}

	action := &TransformAction{}
	config := TransformActionConfig{
		Expression: "trigger.nonexistent",
		Default:    "default_value",
	}

	input := NewActionInput(config, execContext)
	output, err := action.Execute(context.Background(), input)

	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	result, ok := output.Data.(string)
	if !ok {
		t.Fatal("Output data is not a string")
	}

	if result != "default_value" {
		t.Errorf("result = %v, want 'default_value'", result)
	}
}

func TestTransformAction_Execute_NoConfig(t *testing.T) {
	execContext := map[string]interface{}{
		"trigger": map[string]interface{}{
			"data": "test",
		},
	}

	action := &TransformAction{}
	config := TransformActionConfig{}

	input := NewActionInput(config, execContext)
	output, err := action.Execute(context.Background(), input)

	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	// Should return the input context when no transformation is specified
	result, ok := output.Data.(map[string]interface{})
	if !ok {
		t.Fatal("Output data is not a map")
	}

	triggerData, ok := result["trigger"].(map[string]interface{})
	if !ok {
		t.Fatal("trigger data is not a map")
	}

	if triggerData["data"] != "test" {
		t.Errorf("trigger.data = %v, want 'test'", triggerData["data"])
	}
}

func TestTransformAction_Execute_ComplexMapping(t *testing.T) {
	execContext := map[string]interface{}{
		"trigger": map[string]interface{}{
			"webhook": map[string]interface{}{
				"headers": map[string]interface{}{
					"user-agent": "TestAgent/1.0",
				},
				"body": map[string]interface{}{
					"event": "user.created",
					"data": map[string]interface{}{
						"user": map[string]interface{}{
							"id":         789,
							"email":      "test@example.com",
							"created_at": "2024-01-01T00:00:00Z",
						},
					},
				},
			},
		},
		"steps": map[string]interface{}{
			"validate": map[string]interface{}{
				"valid": true,
			},
		},
	}

	action := &TransformAction{}
	config := TransformActionConfig{
		Mapping: map[string]string{
			"event_type": "trigger.webhook.body.event",
			"user_id":    "trigger.webhook.body.data.user.id",
			"email":      "trigger.webhook.body.data.user.email",
			"is_valid":   "steps.validate.valid",
			"user_agent": "trigger.webhook.headers.user-agent",
			"timestamp":  "trigger.webhook.body.data.user.created_at",
		},
	}

	input := NewActionInput(config, execContext)
	output, err := action.Execute(context.Background(), input)

	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	result, ok := output.Data.(map[string]interface{})
	if !ok {
		t.Fatal("Output data is not a map")
	}

	tests := []struct {
		key  string
		want interface{}
	}{
		{"event_type", "user.created"},
		{"email", "test@example.com"},
		{"is_valid", true},
		{"user_agent", "TestAgent/1.0"},
		{"timestamp", "2024-01-01T00:00:00Z"},
	}

	for _, tt := range tests {
		if result[tt.key] != tt.want {
			t.Errorf("%s = %v, want %v", tt.key, result[tt.key], tt.want)
		}
	}

	// Check user_id separately (int/float64 comparison)
	if !testInterfaceEqual(t, result["user_id"], 789) {
		t.Errorf("user_id = %v (%T), want 789", result["user_id"], result["user_id"])
	}
}

func TestTransformAction_ExecuteExpression(t *testing.T) {
	execContext := map[string]interface{}{
		"data": map[string]interface{}{
			"value": 42,
		},
	}

	action := &TransformAction{}

	result, err := action.executeExpression("data.value", execContext)
	if err != nil {
		t.Fatalf("executeExpression() error = %v", err)
	}

	if !testInterfaceEqual(t, result, 42) {
		t.Errorf("result = %v (%T), want 42", result, result)
	}
}

func TestTransformAction_ExecuteMapping(t *testing.T) {
	execContext := map[string]interface{}{
		"source": map[string]interface{}{
			"a": 1,
			"b": 2,
		},
	}

	action := &TransformAction{}
	mapping := map[string]string{
		"field_a": "source.a",
		"field_b": "source.b",
	}

	result, err := action.executeMapping(mapping, execContext)
	if err != nil {
		t.Fatalf("executeMapping() error = %v", err)
	}

	if !testInterfaceEqual(t, result["field_a"], 1) {
		t.Errorf("field_a = %v (%T), want 1", result["field_a"], result["field_a"])
	}

	if !testInterfaceEqual(t, result["field_b"], 2) {
		t.Errorf("field_b = %v (%T), want 2", result["field_b"], result["field_b"])
	}
}

func TestExecuteTransform_LegacyFunction(t *testing.T) {
	execContext := map[string]interface{}{
		"data": map[string]interface{}{
			"value": "test",
		},
	}

	config := TransformActionConfig{
		Expression: "data.value",
	}

	result, err := ExecuteTransform(context.Background(), config, execContext)

	if err != nil {
		t.Fatalf("ExecuteTransform() error = %v", err)
	}

	if result != "test" {
		t.Errorf("result = %v, want 'test'", result)
	}
}

func TestTransformAction_Execute_EmptyMapping(t *testing.T) {
	execContext := map[string]interface{}{
		"data": "test",
	}

	action := &TransformAction{}
	config := TransformActionConfig{
		Mapping: map[string]string{},
	}

	input := NewActionInput(config, execContext)
	output, err := action.Execute(context.Background(), input)

	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	// Empty mapping returns the context itself (no transformation specified)
	result, ok := output.Data.(map[string]interface{})
	if !ok {
		t.Fatal("Output data is not a map")
	}

	// When mapping is empty but mapping key exists, it returns context
	if result["data"] != "test" {
		t.Errorf("Expected context data, got %v", result)
	}
}
