package javascript

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEngine_Execute_BasicScript(t *testing.T) {
	engine, err := NewEngine(nil)
	require.NoError(t, err)
	defer engine.Close()

	config := &ExecuteConfig{
		Script: "return 42;",
	}

	result, err := engine.Execute(context.Background(), config)
	require.NoError(t, err)
	require.NotNil(t, result)

	// Goja returns int64 for integers
	assert.Equal(t, int64(42), result.Result)
}

func TestEngine_Execute_StringReturn(t *testing.T) {
	engine, err := NewEngine(nil)
	require.NoError(t, err)
	defer engine.Close()

	config := &ExecuteConfig{
		Script: `return "Hello, World!";`,
	}

	result, err := engine.Execute(context.Background(), config)
	require.NoError(t, err)
	assert.Equal(t, "Hello, World!", result.Result)
}

func TestEngine_Execute_ObjectReturn(t *testing.T) {
	engine, err := NewEngine(nil)
	require.NoError(t, err)
	defer engine.Close()

	config := &ExecuteConfig{
		Script: `
			return {
				name: "Alice",
				age: 30,
				active: true
			};
		`,
	}

	result, err := engine.Execute(context.Background(), config)
	require.NoError(t, err)

	objMap, ok := result.Result.(map[string]any)
	require.True(t, ok, "Result should be a map")
	assert.Equal(t, "Alice", objMap["name"])
	assert.Equal(t, int64(30), objMap["age"])
	assert.Equal(t, true, objMap["active"])
}

func TestEngine_Execute_ArrayReturn(t *testing.T) {
	engine, err := NewEngine(nil)
	require.NoError(t, err)
	defer engine.Close()

	config := &ExecuteConfig{
		Script: "return [1, 2, 3, 4, 5];",
	}

	result, err := engine.Execute(context.Background(), config)
	require.NoError(t, err)

	arr, ok := result.Result.([]any)
	require.True(t, ok, "Result should be an array")
	assert.Len(t, arr, 5)
	assert.Equal(t, int64(1), arr[0])
}

func TestEngine_Execute_WithContext(t *testing.T) {
	engine, err := NewEngine(nil)
	require.NoError(t, err)
	defer engine.Close()

	execCtx := NewExecutionContext().
		WithTrigger(map[string]any{
			"name": "Bob",
			"id":   100,
		}).
		WithSteps(map[string]any{
			"step1": map[string]any{
				"result": "success",
			},
		})

	config := &ExecuteConfig{
		Script: `
			var userName = context.trigger.name;
			var userId = context.trigger.id;
			var previousResult = context.steps.step1.result;

			return {
				user: userName,
				id: userId,
				previous: previousResult
			};
		`,
		Context: execCtx,
	}

	result, err := engine.Execute(context.Background(), config)
	require.NoError(t, err)

	objMap, ok := result.Result.(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "Bob", objMap["user"])
	assert.Equal(t, int64(100), objMap["id"])
	assert.Equal(t, "success", objMap["previous"])
}

func TestEngine_Execute_WithEnvContext(t *testing.T) {
	engine, err := NewEngine(nil)
	require.NoError(t, err)
	defer engine.Close()

	execCtx := NewExecutionContext().
		WithEnv(map[string]any{
			"tenant_id":    "tenant-123",
			"execution_id": "exec-456",
			"workflow_id":  "wf-789",
		})

	config := &ExecuteConfig{
		Script: `
			return {
				tenant: context.env.tenant_id,
				execution: context.env.execution_id,
				workflow: context.env.workflow_id
			};
		`,
		Context: execCtx,
	}

	result, err := engine.Execute(context.Background(), config)
	require.NoError(t, err)

	objMap, ok := result.Result.(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "tenant-123", objMap["tenant"])
	assert.Equal(t, "exec-456", objMap["execution"])
	assert.Equal(t, "wf-789", objMap["workflow"])
}

func TestEngine_Execute_ComplexLogic(t *testing.T) {
	engine, err := NewEngine(nil)
	require.NoError(t, err)
	defer engine.Close()

	execCtx := NewExecutionContext().
		WithTrigger(map[string]any{
			"numbers": []any{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
		})

	config := &ExecuteConfig{
		Script: `
			var numbers = context.trigger.numbers;
			var sum = 0;
			for (var i = 0; i < numbers.length; i++) {
				sum += numbers[i];
			}

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
		Context: execCtx,
	}

	result, err := engine.Execute(context.Background(), config)
	require.NoError(t, err)

	objMap, ok := result.Result.(map[string]any)
	require.True(t, ok)
	assert.Equal(t, int64(55), objMap["sum"])

	evens, ok := objMap["evens"].([]any)
	require.True(t, ok)
	assert.Len(t, evens, 5)

	assert.Equal(t, int64(10), objMap["count"])
}

func TestEngine_Execute_Timeout(t *testing.T) {
	engine, err := NewEngine(&EngineConfig{
		Limits: NewLimits(1*time.Second, 128),
	})
	require.NoError(t, err)
	defer engine.Close()

	config := &ExecuteConfig{
		Script: `
			while (true) {
				// Infinite loop
			}
		`,
		Timeout: 100 * time.Millisecond,
	}

	_, err = engine.Execute(context.Background(), config)
	require.Error(t, err)
	assert.True(t, IsTimeout(err) || strings.Contains(err.Error(), "timeout"))
}

func TestEngine_Execute_ContextTimeout(t *testing.T) {
	engine, err := NewEngine(nil)
	require.NoError(t, err)
	defer engine.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	config := &ExecuteConfig{
		Script: `
			var sum = 0;
			for (var i = 0; i < 1000000000; i++) {
				sum += i;
			}
			return sum;
		`,
	}

	_, err = engine.Execute(ctx, config)
	require.Error(t, err)
}

func TestEngine_Execute_SyntaxError(t *testing.T) {
	engine, err := NewEngine(nil)
	require.NoError(t, err)
	defer engine.Close()

	config := &ExecuteConfig{
		Script: `
			return {
				invalid syntax here
			};
		`,
	}

	_, err = engine.Execute(context.Background(), config)
	require.Error(t, err)
}

func TestEngine_Execute_RuntimeError(t *testing.T) {
	engine, err := NewEngine(nil)
	require.NoError(t, err)
	defer engine.Close()

	config := &ExecuteConfig{
		Script: `
			return undefinedVariable.property;
		`,
	}

	_, err = engine.Execute(context.Background(), config)
	require.Error(t, err)
}

func TestEngine_Execute_EmptyScript(t *testing.T) {
	engine, err := NewEngine(nil)
	require.NoError(t, err)
	defer engine.Close()

	config := &ExecuteConfig{
		Script: "",
	}

	_, err = engine.Execute(context.Background(), config)
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrEmptyScript)
}

func TestEngine_Execute_NoReturn(t *testing.T) {
	engine, err := NewEngine(nil)
	require.NoError(t, err)
	defer engine.Close()

	config := &ExecuteConfig{
		Script: `
			var x = 42;
		`,
	}

	result, err := engine.Execute(context.Background(), config)
	require.NoError(t, err)
	// JavaScript returns undefined when no explicit return
	assert.Nil(t, result.Result)
}

func TestEngine_Execute_JSONOperations(t *testing.T) {
	engine, err := NewEngine(nil)
	require.NoError(t, err)
	defer engine.Close()

	config := &ExecuteConfig{
		Script: `
			var data = {
				name: "Test",
				value: 123
			};

			var jsonString = JSON.stringify(data);
			var parsed = JSON.parse(jsonString);

			return {
				original: data,
				stringified: jsonString,
				parsed: parsed
			};
		`,
	}

	result, err := engine.Execute(context.Background(), config)
	require.NoError(t, err)

	objMap, ok := result.Result.(map[string]any)
	require.True(t, ok)
	assert.NotNil(t, objMap["stringified"])

	parsed, ok := objMap["parsed"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "Test", parsed["name"])
}

func TestEngine_Execute_MathOperations(t *testing.T) {
	engine, err := NewEngine(nil)
	require.NoError(t, err)
	defer engine.Close()

	config := &ExecuteConfig{
		Script: `
			return {
				sqrt: Math.sqrt(16),
				pow: Math.pow(2, 3),
				round: Math.round(3.7),
				max: Math.max(1, 5, 3),
				min: Math.min(1, 5, 3)
			};
		`,
	}

	result, err := engine.Execute(context.Background(), config)
	require.NoError(t, err)

	objMap, ok := result.Result.(map[string]any)
	require.True(t, ok)

	// Goja returns int64 for integer results from Math functions
	// Check numeric equality regardless of type
	assert.EqualValues(t, 4, objMap["sqrt"])
	assert.EqualValues(t, 8, objMap["pow"])
	assert.EqualValues(t, 4, objMap["round"])
	assert.EqualValues(t, 5, objMap["max"])
	assert.EqualValues(t, 1, objMap["min"])
}

func TestEngine_Execute_ConsoleCapture(t *testing.T) {
	engine, err := NewEngine(&EngineConfig{
		EnableConsoleCapture: true,
	})
	require.NoError(t, err)
	defer engine.Close()

	config := &ExecuteConfig{
		Script: `
			console.log("Hello", "World");
			console.warn("This is a warning");
			console.error("This is an error");
			return 42;
		`,
	}

	result, err := engine.Execute(context.Background(), config)
	require.NoError(t, err)
	assert.Equal(t, int64(42), result.Result)

	require.Len(t, result.ConsoleLogs, 3)
	assert.Equal(t, "info", result.ConsoleLogs[0].Level)
	assert.Contains(t, result.ConsoleLogs[0].Message, "Hello")
	assert.Equal(t, "warn", result.ConsoleLogs[1].Level)
	assert.Equal(t, "error", result.ConsoleLogs[2].Level)
}

func TestEngine_Execute_ExecutionMetadata(t *testing.T) {
	engine, err := NewEngine(nil)
	require.NoError(t, err)
	defer engine.Close()

	config := &ExecuteConfig{
		Script:      "return 42;",
		ExecutionID: "test-exec-123",
		TenantID:    "tenant-abc",
		WorkflowID:  "workflow-xyz",
		NodeID:      "node-001",
	}

	result, err := engine.Execute(context.Background(), config)
	require.NoError(t, err)
	assert.Equal(t, "test-exec-123", result.ExecutionID)
	assert.Greater(t, result.Duration, time.Duration(0))
}

func TestEngine_Validate_ValidScript(t *testing.T) {
	engine, err := NewEngine(nil)
	require.NoError(t, err)
	defer engine.Close()

	err = engine.Validate("return 42;")
	assert.NoError(t, err)
}

func TestEngine_Validate_EmptyScript(t *testing.T) {
	engine, err := NewEngine(nil)
	require.NoError(t, err)
	defer engine.Close()

	err = engine.Validate("")
	assert.Error(t, err)
}

func TestEngine_Compile_ValidScript(t *testing.T) {
	engine, err := NewEngine(nil)
	require.NoError(t, err)
	defer engine.Close()

	err = engine.Compile("return { x: 1, y: 2 };")
	assert.NoError(t, err)
}

func TestEngine_Compile_SyntaxError(t *testing.T) {
	engine, err := NewEngine(nil)
	require.NoError(t, err)
	defer engine.Close()

	err = engine.Compile("return { invalid syntax };")
	assert.Error(t, err)
}

func TestEngine_Execute_CtxAlias(t *testing.T) {
	engine, err := NewEngine(nil)
	require.NoError(t, err)
	defer engine.Close()

	execCtx := NewExecutionContext().
		WithTrigger(map[string]any{
			"value": 42,
		})

	config := &ExecuteConfig{
		Script: `
			// Test that ctx is available as alias for context
			return ctx.trigger.value;
		`,
		Context: execCtx,
	}

	result, err := engine.Execute(context.Background(), config)
	require.NoError(t, err)
	assert.Equal(t, int64(42), result.Result)
}

// Security tests

func TestEngine_Execute_NoRequire(t *testing.T) {
	engine, err := NewEngine(nil)
	require.NoError(t, err)
	defer engine.Close()

	config := &ExecuteConfig{
		Script: `
			var fs = require('fs');
			return "accessed";
		`,
	}

	_, err = engine.Execute(context.Background(), config)
	require.Error(t, err)
}

func TestEngine_Execute_NoProcess(t *testing.T) {
	engine, err := NewEngine(nil)
	require.NoError(t, err)
	defer engine.Close()

	config := &ExecuteConfig{
		Script: `
			return process.env;
		`,
	}

	_, err = engine.Execute(context.Background(), config)
	require.Error(t, err)
}

func TestEngine_Execute_NoGlobal(t *testing.T) {
	engine, err := NewEngine(nil)
	require.NoError(t, err)
	defer engine.Close()

	config := &ExecuteConfig{
		Script: `
			return global.setTimeout;
		`,
	}

	_, err = engine.Execute(context.Background(), config)
	require.Error(t, err)
}

func TestEngine_Validate_ForbiddenEval(t *testing.T) {
	engine, err := NewEngine(nil)
	require.NoError(t, err)
	defer engine.Close()

	err = engine.Validate(`eval("alert(1)")`)
	assert.Error(t, err)
	assert.True(t, IsSandboxViolation(err))
}

func TestEngine_Validate_ForbiddenNewFunction(t *testing.T) {
	engine, err := NewEngine(nil)
	require.NoError(t, err)
	defer engine.Close()

	err = engine.Validate(`new Function("return 1")()`)
	assert.Error(t, err)
	assert.True(t, IsSandboxViolation(err))
}

// Benchmark tests

func BenchmarkEngine_Execute_Simple(b *testing.B) {
	engine, err := NewEngine(nil)
	require.NoError(b, err)
	defer engine.Close()

	config := &ExecuteConfig{
		Script: "return 42;",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := engine.Execute(context.Background(), config)
		if err != nil {
			b.Fatalf("Execute() error = %v", err)
		}
	}
}

func BenchmarkEngine_Execute_WithContext(b *testing.B) {
	engine, err := NewEngine(nil)
	require.NoError(b, err)
	defer engine.Close()

	execCtx := NewExecutionContext().
		WithTrigger(map[string]any{
			"numbers": []any{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
		})

	config := &ExecuteConfig{
		Script: `
			var sum = 0;
			var numbers = context.trigger.numbers;
			for (var i = 0; i < numbers.length; i++) {
				sum += numbers[i];
			}
			return sum;
		`,
		Context: execCtx,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := engine.Execute(context.Background(), config)
		if err != nil {
			b.Fatalf("Execute() error = %v", err)
		}
	}
}
