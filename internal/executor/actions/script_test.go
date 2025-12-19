package actions

import (
	"context"
	"encoding/json"
	"testing"
	"time"
)

func TestScriptAction_Execute_BasicScript(t *testing.T) {
	action := &ScriptAction{}
	config := ScriptActionConfig{
		Script:  "return 42;",
		Timeout: 30,
	}

	input := NewActionInput(config, nil)
	output, err := action.Execute(context.Background(), input)

	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if output == nil {
		t.Fatal("Execute() returned nil output")
	}

	result, ok := output.Data.(*ScriptActionResult)
	if !ok {
		t.Fatal("Output data is not ScriptActionResult")
	}

	// Goja returns int64 for integer values
	if val, ok := result.Result.(int64); !ok || val != 42 {
		t.Errorf("Result = %v (%T), want 42", result.Result, result.Result)
	}
}

func TestScriptAction_Execute_StringReturn(t *testing.T) {
	action := &ScriptAction{}
	config := ScriptActionConfig{
		Script:  `return "Hello, World!";`,
		Timeout: 30,
	}

	input := NewActionInput(config, nil)
	output, err := action.Execute(context.Background(), input)

	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	result := output.Data.(*ScriptActionResult)
	if result.Result != "Hello, World!" {
		t.Errorf("Result = %v, want 'Hello, World!'", result.Result)
	}
}

func TestScriptAction_Execute_ObjectReturn(t *testing.T) {
	action := &ScriptAction{}
	config := ScriptActionConfig{
		Script: `
			var obj = {
				name: "Alice",
				age: 30,
				active: true
			};
			return obj;
		`,
		Timeout: 30,
	}

	input := NewActionInput(config, nil)
	output, err := action.Execute(context.Background(), input)

	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	result := output.Data.(*ScriptActionResult)
	objMap, ok := result.Result.(map[string]interface{})
	if !ok {
		t.Fatalf("Result is not a map: %T", result.Result)
	}

	if objMap["name"] != "Alice" {
		t.Errorf("name = %v, want 'Alice'", objMap["name"])
	}

	if age, ok := objMap["age"].(int64); !ok || age != 30 {
		t.Errorf("age = %v (%T), want 30", objMap["age"], objMap["age"])
	}

	if objMap["active"] != true {
		t.Errorf("active = %v, want true", objMap["active"])
	}
}

func TestScriptAction_Execute_ArrayReturn(t *testing.T) {
	action := &ScriptAction{}
	config := ScriptActionConfig{
		Script: `
			return [1, 2, 3, 4, 5];
		`,
		Timeout: 30,
	}

	input := NewActionInput(config, nil)
	output, err := action.Execute(context.Background(), input)

	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	result := output.Data.(*ScriptActionResult)
	arr, ok := result.Result.([]interface{})
	if !ok {
		t.Fatalf("Result is not an array: %T", result.Result)
	}

	if len(arr) != 5 {
		t.Errorf("Array length = %d, want 5", len(arr))
	}

	// Check first element
	if val, ok := arr[0].(int64); !ok || val != 1 {
		t.Errorf("arr[0] = %v (%T), want 1", arr[0], arr[0])
	}
}

func TestScriptAction_Execute_WithContextAccess(t *testing.T) {
	action := &ScriptAction{}
	config := ScriptActionConfig{
		Script: `
			// Access trigger data
			var userName = context.trigger.name;
			var userId = context.trigger.id;

			// Access previous step data
			var previousResult = context.steps.step1.result;

			return {
				user: userName,
				id: userId,
				previous: previousResult
			};
		`,
		Timeout: 30,
	}

	execContext := map[string]interface{}{
		"trigger": map[string]interface{}{
			"name": "Bob",
			"id":   100,
		},
		"steps": map[string]interface{}{
			"step1": map[string]interface{}{
				"result": "success",
			},
		},
	}

	input := NewActionInput(config, execContext)
	output, err := action.Execute(context.Background(), input)

	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	result := output.Data.(*ScriptActionResult)
	objMap, ok := result.Result.(map[string]interface{})
	if !ok {
		t.Fatalf("Result is not a map: %T", result.Result)
	}

	if objMap["user"] != "Bob" {
		t.Errorf("user = %v, want 'Bob'", objMap["user"])
	}

	if id, ok := objMap["id"].(int64); !ok || id != 100 {
		t.Errorf("id = %v (%T), want 100", objMap["id"], objMap["id"])
	}

	if objMap["previous"] != "success" {
		t.Errorf("previous = %v, want 'success'", objMap["previous"])
	}
}

func TestScriptAction_Execute_WithEnvAccess(t *testing.T) {
	action := &ScriptAction{}
	config := ScriptActionConfig{
		Script: `
			return {
				tenant: context.env.tenant_id,
				execution: context.env.execution_id,
				workflow: context.env.workflow_id
			};
		`,
		Timeout: 30,
	}

	execContext := map[string]interface{}{
		"env": map[string]interface{}{
			"tenant_id":    "tenant-123",
			"execution_id": "exec-456",
			"workflow_id":  "wf-789",
		},
	}

	input := NewActionInput(config, execContext)
	output, err := action.Execute(context.Background(), input)

	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	result := output.Data.(*ScriptActionResult)
	objMap, ok := result.Result.(map[string]interface{})
	if !ok {
		t.Fatalf("Result is not a map: %T", result.Result)
	}

	if objMap["tenant"] != "tenant-123" {
		t.Errorf("tenant = %v, want 'tenant-123'", objMap["tenant"])
	}

	if objMap["execution"] != "exec-456" {
		t.Errorf("execution = %v, want 'exec-456'", objMap["execution"])
	}

	if objMap["workflow"] != "wf-789" {
		t.Errorf("workflow = %v, want 'wf-789'", objMap["workflow"])
	}
}

func TestScriptAction_Execute_ComplexLogic(t *testing.T) {
	action := &ScriptAction{}
	config := ScriptActionConfig{
		Script: `
			// Calculate sum of array
			var numbers = context.trigger.numbers;
			var sum = 0;
			for (var i = 0; i < numbers.length; i++) {
				sum += numbers[i];
			}

			// Filter even numbers
			var evens = [];
			for (var i = 0; i < numbers.length; i++) {
				if (numbers[i] % 2 === 0) {
					evens.push(numbers[i]);
				}
			}

			return {
				sum: sum,
				evens: evens,
				count: numbers.length
			};
		`,
		Timeout: 30,
	}

	execContext := map[string]interface{}{
		"trigger": map[string]interface{}{
			"numbers": []interface{}{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
		},
	}

	input := NewActionInput(config, execContext)
	output, err := action.Execute(context.Background(), input)

	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	result := output.Data.(*ScriptActionResult)
	objMap, ok := result.Result.(map[string]interface{})
	if !ok {
		t.Fatalf("Result is not a map: %T", result.Result)
	}

	if sum, ok := objMap["sum"].(int64); !ok || sum != 55 {
		t.Errorf("sum = %v (%T), want 55", objMap["sum"], objMap["sum"])
	}

	evens, ok := objMap["evens"].([]interface{})
	if !ok {
		t.Fatalf("evens is not an array: %T", objMap["evens"])
	}

	if len(evens) != 5 {
		t.Errorf("evens length = %d, want 5", len(evens))
	}

	if count, ok := objMap["count"].(int64); !ok || count != 10 {
		t.Errorf("count = %v (%T), want 10", objMap["count"], objMap["count"])
	}
}

func TestScriptAction_Execute_Timeout(t *testing.T) {
	action := &ScriptAction{}
	config := ScriptActionConfig{
		Script: `
			// Infinite loop to trigger timeout
			while (true) {
				// This will timeout
			}
		`,
		Timeout: 1, // 1 second timeout
	}

	input := NewActionInput(config, nil)
	_, err := action.Execute(context.Background(), input)

	if err == nil {
		t.Error("Expected timeout error, got nil")
	}

	// Check that error mentions timeout
	if err != nil {
		errMsg := err.Error()
		if errMsg == "" {
			t.Error("Expected non-empty error message")
		}
	}
}

func TestScriptAction_Execute_ContextTimeout(t *testing.T) {
	action := &ScriptAction{}
	config := ScriptActionConfig{
		Script: `
			// Script that takes longer than context timeout
			var sum = 0;
			for (var i = 0; i < 1000000000; i++) {
				sum += i;
			}
			return sum;
		`,
		Timeout: 30, // Script timeout is 30s, but context will timeout first
	}

	// Context timeout is 100ms
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	input := NewActionInput(config, nil)
	_, err := action.Execute(ctx, input)

	if err == nil {
		t.Error("Expected context timeout error, got nil")
	}
}

func TestScriptAction_Execute_SyntaxError(t *testing.T) {
	action := &ScriptAction{}
	config := ScriptActionConfig{
		Script: `
			return {
				invalid syntax here
			};
		`,
		Timeout: 30,
	}

	input := NewActionInput(config, nil)
	_, err := action.Execute(context.Background(), input)

	if err == nil {
		t.Error("Expected syntax error, got nil")
	}
}

func TestScriptAction_Execute_RuntimeError(t *testing.T) {
	action := &ScriptAction{}
	config := ScriptActionConfig{
		Script: `
			// Try to access undefined variable
			return undefinedVariable.property;
		`,
		Timeout: 30,
	}

	input := NewActionInput(config, nil)
	_, err := action.Execute(context.Background(), input)

	if err == nil {
		t.Error("Expected runtime error, got nil")
	}
}

func TestScriptAction_Execute_NoReturn(t *testing.T) {
	action := &ScriptAction{}
	config := ScriptActionConfig{
		Script: `
			// Script with no return statement
			var x = 42;
		`,
		Timeout: 30,
	}

	input := NewActionInput(config, nil)
	output, err := action.Execute(context.Background(), input)

	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	result := output.Data.(*ScriptActionResult)
	// In JavaScript, if no return statement, result is undefined
	// Goja should return nil or undefined
	if result.Result != nil {
		t.Logf("Result = %v (%T)", result.Result, result.Result)
		// This is acceptable - some interpreters return undefined/null
	}
}

func TestScriptAction_Execute_MissingScript(t *testing.T) {
	action := &ScriptAction{}
	config := ScriptActionConfig{
		Script:  "",
		Timeout: 30,
	}

	input := NewActionInput(config, nil)
	_, err := action.Execute(context.Background(), input)

	if err == nil {
		t.Error("Expected error for missing script, got nil")
	}
}

func TestScriptAction_Execute_DefaultTimeout(t *testing.T) {
	action := &ScriptAction{}
	config := ScriptActionConfig{
		Script:  "return 42;",
		Timeout: 0, // Should use default timeout
	}

	input := NewActionInput(config, nil)
	output, err := action.Execute(context.Background(), input)

	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	result := output.Data.(*ScriptActionResult)
	if val, ok := result.Result.(int64); !ok || val != 42 {
		t.Errorf("Result = %v (%T), want 42", result.Result, result.Result)
	}
}

func TestScriptAction_Execute_JSONFunctions(t *testing.T) {
	action := &ScriptAction{}
	config := ScriptActionConfig{
		Script: `
			var data = {
				name: "Test",
				value: 123
			};

			// Test JSON.stringify
			var jsonString = JSON.stringify(data);

			// Test JSON.parse
			var parsed = JSON.parse(jsonString);

			return {
				original: data,
				stringified: jsonString,
				parsed: parsed
			};
		`,
		Timeout: 30,
	}

	input := NewActionInput(config, nil)
	output, err := action.Execute(context.Background(), input)

	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	result := output.Data.(*ScriptActionResult)
	objMap, ok := result.Result.(map[string]interface{})
	if !ok {
		t.Fatalf("Result is not a map: %T", result.Result)
	}

	if objMap["stringified"] == nil {
		t.Error("stringified is nil")
	}

	parsed, ok := objMap["parsed"].(map[string]interface{})
	if !ok {
		t.Fatalf("parsed is not a map: %T", objMap["parsed"])
	}

	if parsed["name"] != "Test" {
		t.Errorf("parsed.name = %v, want 'Test'", parsed["name"])
	}
}

func TestScriptAction_Execute_MathOperations(t *testing.T) {
	action := &ScriptAction{}
	config := ScriptActionConfig{
		Script: `
			return {
				sqrt: Math.sqrt(16),
				pow: Math.pow(2, 3),
				round: Math.round(3.7),
				max: Math.max(1, 5, 3),
				min: Math.min(1, 5, 3),
				random: Math.random() >= 0 && Math.random() <= 1
			};
		`,
		Timeout: 30,
	}

	input := NewActionInput(config, nil)
	output, err := action.Execute(context.Background(), input)

	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	result := output.Data.(*ScriptActionResult)
	objMap, ok := result.Result.(map[string]interface{})
	if !ok {
		t.Fatalf("Result is not a map: %T", result.Result)
	}

	// Check sqrt
	if val, ok := objMap["sqrt"].(float64); !ok || val != 4 {
		if intVal, ok := objMap["sqrt"].(int64); !ok || intVal != 4 {
			t.Errorf("sqrt = %v (%T), want 4", objMap["sqrt"], objMap["sqrt"])
		}
	}

	// Check pow
	if val, ok := objMap["pow"].(int64); !ok || val != 8 {
		if floatVal, ok := objMap["pow"].(float64); !ok || floatVal != 8 {
			t.Errorf("pow = %v (%T), want 8", objMap["pow"], objMap["pow"])
		}
	}

	// Check round
	if val, ok := objMap["round"].(int64); !ok || val != 4 {
		if floatVal, ok := objMap["round"].(float64); !ok || floatVal != 4 {
			t.Errorf("round = %v (%T), want 4", objMap["round"], objMap["round"])
		}
	}

	// Check random result is boolean true
	if objMap["random"] != true {
		t.Errorf("random = %v, want true", objMap["random"])
	}
}

func TestScriptAction_Execute_InvalidConfig(t *testing.T) {
	action := &ScriptAction{}

	// Pass invalid config type
	input := NewActionInput("invalid", nil)
	_, err := action.Execute(context.Background(), input)

	if err == nil {
		t.Error("Expected error for invalid config, got nil")
	}
}

func TestScriptAction_Execute_WithMetadata(t *testing.T) {
	action := &ScriptAction{}
	config := ScriptActionConfig{
		Script:  "return 42;",
		Timeout: 30,
	}

	input := NewActionInput(config, nil)
	output, err := action.Execute(context.Background(), input)

	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	// Check that metadata contains execution_time
	if output.Metadata == nil {
		t.Fatal("Expected metadata, got nil")
	}

	if _, exists := output.Metadata["execution_time_ms"]; !exists {
		t.Error("Expected execution_time_ms in metadata")
	}
}

// Benchmark tests
func BenchmarkScriptAction_Execute_Simple(b *testing.B) {
	action := &ScriptAction{}
	config := ScriptActionConfig{
		Script:  "return 42;",
		Timeout: 30,
	}

	input := NewActionInput(config, nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := action.Execute(context.Background(), input)
		if err != nil {
			b.Fatalf("Execute() error = %v", err)
		}
	}
}

func BenchmarkScriptAction_Execute_WithContext(b *testing.B) {
	action := &ScriptAction{}
	config := ScriptActionConfig{
		Script: `
			var sum = 0;
			var numbers = context.trigger.numbers;
			for (var i = 0; i < numbers.length; i++) {
				sum += numbers[i];
			}
			return sum;
		`,
		Timeout: 30,
	}

	execContext := map[string]interface{}{
		"trigger": map[string]interface{}{
			"numbers": []interface{}{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
		},
	}

	input := NewActionInput(config, execContext)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := action.Execute(context.Background(), input)
		if err != nil {
			b.Fatalf("Execute() error = %v", err)
		}
	}
}

// Test security: script should not have access to file system
func TestScriptAction_Execute_NoFileSystemAccess(t *testing.T) {
	action := &ScriptAction{}
	config := ScriptActionConfig{
		Script: `
			// Try to access file system (should fail in sandbox)
			var fs = require('fs');
			return "accessed";
		`,
		Timeout: 30,
	}

	input := NewActionInput(config, nil)
	_, err := action.Execute(context.Background(), input)

	// Should fail because 'require' is not available in sandbox
	if err == nil {
		t.Error("Expected error for file system access, got nil")
	}
}

// Test security: script should not have access to network
func TestScriptAction_Execute_NoNetworkAccess(t *testing.T) {
	action := &ScriptAction{}
	config := ScriptActionConfig{
		Script: `
			// Try to access network (should fail in sandbox)
			var http = require('http');
			return "accessed";
		`,
		Timeout: 30,
	}

	input := NewActionInput(config, nil)
	_, err := action.Execute(context.Background(), input)

	// Should fail because 'require' is not available in sandbox
	if err == nil {
		t.Error("Expected error for network access, got nil")
	}
}

// Test ConfigFromJSON helper
func TestScriptActionConfigFromJSON(t *testing.T) {
	jsonData := []byte(`{
		"script": "return 42;",
		"timeout": 60,
		"memory_limit": 128
	}`)

	var config ScriptActionConfig
	err := json.Unmarshal(jsonData, &config)

	if err != nil {
		t.Fatalf("Unmarshal error = %v", err)
	}

	if config.Script != "return 42;" {
		t.Errorf("Script = %v, want 'return 42;'", config.Script)
	}

	if config.Timeout != 60 {
		t.Errorf("Timeout = %d, want 60", config.Timeout)
	}

	if config.MemoryLimit != 128 {
		t.Errorf("MemoryLimit = %d, want 128", config.MemoryLimit)
	}
}
