package javascript

import (
	"testing"

	"github.com/dop251/goja"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewExecutionContext(t *testing.T) {
	ctx := NewExecutionContext()

	assert.NotNil(t, ctx.Trigger)
	assert.NotNil(t, ctx.Steps)
	assert.NotNil(t, ctx.Env)
	assert.NotNil(t, ctx.Vars)
	assert.NotNil(t, ctx.Input)
}

func TestExecutionContext_WithTrigger(t *testing.T) {
	ctx := NewExecutionContext().
		WithTrigger(map[string]any{
			"event": "webhook",
			"data":  "test",
		})

	assert.Equal(t, "webhook", ctx.Trigger["event"])
	assert.Equal(t, "test", ctx.Trigger["data"])
}

func TestExecutionContext_WithSteps(t *testing.T) {
	ctx := NewExecutionContext().
		WithSteps(map[string]any{
			"step1": map[string]any{"result": "success"},
			"step2": map[string]any{"result": "pending"},
		})

	step1 := ctx.Steps["step1"].(map[string]any)
	assert.Equal(t, "success", step1["result"])
}

func TestExecutionContext_WithEnv(t *testing.T) {
	ctx := NewExecutionContext().
		WithEnv(map[string]any{
			"tenant_id":    "t-123",
			"execution_id": "e-456",
		})

	assert.Equal(t, "t-123", ctx.Env["tenant_id"])
	assert.Equal(t, "e-456", ctx.Env["execution_id"])
}

func TestExecutionContext_WithVars(t *testing.T) {
	ctx := NewExecutionContext().
		WithVars(map[string]any{
			"counter": 42,
			"name":    "test",
		})

	assert.Equal(t, 42, ctx.Vars["counter"])
	assert.Equal(t, "test", ctx.Vars["name"])
}

func TestExecutionContext_WithInput(t *testing.T) {
	ctx := NewExecutionContext().
		WithInput(map[string]any{
			"param1": "value1",
			"param2": 123,
		})

	assert.Equal(t, "value1", ctx.Input["param1"])
	assert.Equal(t, 123, ctx.Input["param2"])
}

func TestExecutionContext_Chaining(t *testing.T) {
	ctx := NewExecutionContext().
		WithTrigger(map[string]any{"a": 1}).
		WithSteps(map[string]any{"b": 2}).
		WithEnv(map[string]any{"c": 3}).
		WithVars(map[string]any{"d": 4}).
		WithInput(map[string]any{"e": 5})

	assert.Equal(t, 1, ctx.Trigger["a"])
	assert.Equal(t, 2, ctx.Steps["b"])
	assert.Equal(t, 3, ctx.Env["c"])
	assert.Equal(t, 4, ctx.Vars["d"])
	assert.Equal(t, 5, ctx.Input["e"])
}

func TestExecutionContext_ToMap(t *testing.T) {
	ctx := NewExecutionContext().
		WithTrigger(map[string]any{"event": "test"}).
		WithEnv(map[string]any{"tenant": "abc"})

	m := ctx.ToMap()

	assert.Contains(t, m, "trigger")
	assert.Contains(t, m, "steps")
	assert.Contains(t, m, "env")
	assert.Contains(t, m, "vars")
	assert.Contains(t, m, "input")

	trigger := m["trigger"].(map[string]any)
	assert.Equal(t, "test", trigger["event"])
}

func TestFromWorkflowContext(t *testing.T) {
	workflowCtx := map[string]any{
		"trigger": map[string]any{"event": "webhook"},
		"steps": map[string]any{
			"step1": map[string]any{"result": "done"},
		},
		"env": map[string]any{"tenant_id": "t-123"},
	}

	ctx := FromWorkflowContext(workflowCtx)

	assert.Equal(t, "webhook", ctx.Trigger["event"])
	step1 := ctx.Steps["step1"].(map[string]any)
	assert.Equal(t, "done", step1["result"])
	assert.Equal(t, "t-123", ctx.Env["tenant_id"])
}

func TestFromWorkflowContext_Empty(t *testing.T) {
	ctx := FromWorkflowContext(nil)

	assert.NotNil(t, ctx.Trigger)
	assert.NotNil(t, ctx.Steps)
	assert.NotNil(t, ctx.Env)
	assert.NotNil(t, ctx.Vars)
	assert.NotNil(t, ctx.Input)
}

func TestContextInjector_InjectContext(t *testing.T) {
	injector := NewContextInjector()
	vm := goja.New()

	ctx := NewExecutionContext().
		WithTrigger(map[string]any{
			"name":  "test",
			"value": 42,
		})

	err := injector.InjectContext(vm, ctx)
	require.NoError(t, err)

	// Verify context is available
	val, err := vm.RunString("context.trigger.name")
	require.NoError(t, err)
	assert.Equal(t, "test", val.Export())

	// Verify ctx alias is available
	val, err = vm.RunString("ctx.trigger.value")
	require.NoError(t, err)
	assert.Equal(t, int64(42), val.Export())
}

func TestContextInjector_InjectContext_NilContext(t *testing.T) {
	injector := NewContextInjector()
	vm := goja.New()

	err := injector.InjectContext(vm, nil)
	assert.ErrorIs(t, err, ErrNilContext)
}

func TestContextInjector_InjectContext_NestedData(t *testing.T) {
	injector := NewContextInjector()
	vm := goja.New()

	ctx := NewExecutionContext().
		WithTrigger(map[string]any{
			"user": map[string]any{
				"id":   123,
				"name": "Alice",
				"tags": []any{"admin", "active"},
			},
		})

	err := injector.InjectContext(vm, ctx)
	require.NoError(t, err)

	// Access nested properties
	val, err := vm.RunString("context.trigger.user.name")
	require.NoError(t, err)
	assert.Equal(t, "Alice", val.Export())

	val, err = vm.RunString("context.trigger.user.tags[0]")
	require.NoError(t, err)
	assert.Equal(t, "admin", val.Export())
}

func TestResultExtractor_ExtractResult(t *testing.T) {
	extractor := NewResultExtractor()
	vm := goja.New()

	tests := []struct {
		name     string
		script   string
		expected any
	}{
		{
			name:     "integer",
			script:   "42",
			expected: int64(42),
		},
		{
			name:     "string",
			script:   `"hello"`,
			expected: "hello",
		},
		{
			name:     "boolean",
			script:   "true",
			expected: true,
		},
		{
			name:     "float",
			script:   "3.14",
			expected: 3.14,
		},
		{
			name:     "null",
			script:   "null",
			expected: nil,
		},
		{
			name:     "undefined",
			script:   "undefined",
			expected: nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			val, err := vm.RunString(tc.script)
			require.NoError(t, err)

			result, err := extractor.ExtractResult(val)
			require.NoError(t, err)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestResultExtractor_ExtractResult_Object(t *testing.T) {
	extractor := NewResultExtractor()
	vm := goja.New()

	val, err := vm.RunString(`({ name: "test", value: 123 })`)
	require.NoError(t, err)

	result, err := extractor.ExtractResult(val)
	require.NoError(t, err)

	m, ok := result.(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "test", m["name"])
	assert.Equal(t, int64(123), m["value"])
}

func TestResultExtractor_ExtractResult_Array(t *testing.T) {
	extractor := NewResultExtractor()
	vm := goja.New()

	val, err := vm.RunString(`[1, 2, 3]`)
	require.NoError(t, err)

	result, err := extractor.ExtractResult(val)
	require.NoError(t, err)

	arr, ok := result.([]any)
	require.True(t, ok)
	assert.Len(t, arr, 3)
	assert.Equal(t, int64(1), arr[0])
}

func TestResultExtractor_ExtractJSON(t *testing.T) {
	extractor := NewResultExtractor()
	vm := goja.New()

	val, err := vm.RunString(`({ a: 1, b: "test" })`)
	require.NoError(t, err)

	jsonBytes, err := extractor.ExtractJSON(val)
	require.NoError(t, err)
	assert.Contains(t, string(jsonBytes), `"a":1`)
	assert.Contains(t, string(jsonBytes), `"b":"test"`)
}

func TestValueSanitizer_Sanitize(t *testing.T) {
	sanitizer := NewValueSanitizer()

	tests := []struct {
		name     string
		input    any
		expected any
	}{
		{
			name:     "string",
			input:    "hello",
			expected: "hello",
		},
		{
			name:     "int",
			input:    42,
			expected: 42,
		},
		{
			name:     "float",
			input:    3.14,
			expected: 3.14,
		},
		{
			name:     "bool",
			input:    true,
			expected: true,
		},
		{
			name:     "nil",
			input:    nil,
			expected: nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := sanitizer.Sanitize(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestValueSanitizer_Sanitize_NestedMap(t *testing.T) {
	sanitizer := NewValueSanitizer()

	input := map[string]any{
		"level1": map[string]any{
			"level2": map[string]any{
				"value": "deep",
			},
		},
	}

	result := sanitizer.Sanitize(input)

	m, ok := result.(map[string]any)
	require.True(t, ok)

	level1 := m["level1"].(map[string]any)
	level2 := level1["level2"].(map[string]any)
	assert.Equal(t, "deep", level2["value"])
}

func TestValueSanitizer_Sanitize_Array(t *testing.T) {
	sanitizer := NewValueSanitizer()

	input := []any{1, "two", 3.0, true}
	result := sanitizer.Sanitize(input)

	arr, ok := result.([]any)
	require.True(t, ok)
	assert.Len(t, arr, 4)
}

func TestVariableScope_SetGet(t *testing.T) {
	scope := NewVariableScope(nil)

	scope.Set("x", 42)
	scope.Set("y", "hello")

	val, ok := scope.Get("x")
	assert.True(t, ok)
	assert.Equal(t, 42, val)

	val, ok = scope.Get("y")
	assert.True(t, ok)
	assert.Equal(t, "hello", val)
}

func TestVariableScope_GetNotFound(t *testing.T) {
	scope := NewVariableScope(nil)

	_, ok := scope.Get("notfound")
	assert.False(t, ok)
}

func TestVariableScope_ParentScope(t *testing.T) {
	parent := NewVariableScope(nil)
	parent.Set("a", 1)
	parent.Set("b", 2)

	child := NewVariableScope(parent)
	child.Set("b", 20) // Override parent
	child.Set("c", 3)

	// Get from child
	val, ok := child.Get("c")
	assert.True(t, ok)
	assert.Equal(t, 3, val)

	// Get overridden value
	val, ok = child.Get("b")
	assert.True(t, ok)
	assert.Equal(t, 20, val)

	// Get from parent
	val, ok = child.Get("a")
	assert.True(t, ok)
	assert.Equal(t, 1, val)
}

func TestVariableScope_GetLocal(t *testing.T) {
	parent := NewVariableScope(nil)
	parent.Set("a", 1)

	child := NewVariableScope(parent)
	child.Set("b", 2)

	// Local variable exists
	val, ok := child.GetLocal("b")
	assert.True(t, ok)
	assert.Equal(t, 2, val)

	// Parent variable not in local scope
	_, ok = child.GetLocal("a")
	assert.False(t, ok)
}

func TestVariableScope_Delete(t *testing.T) {
	scope := NewVariableScope(nil)
	scope.Set("x", 42)

	scope.Delete("x")

	_, ok := scope.Get("x")
	assert.False(t, ok)
}

func TestVariableScope_All(t *testing.T) {
	parent := NewVariableScope(nil)
	parent.Set("a", 1)
	parent.Set("b", 2)

	child := NewVariableScope(parent)
	child.Set("b", 20) // Override
	child.Set("c", 3)

	all := child.All()

	assert.Equal(t, 1, all["a"])
	assert.Equal(t, 20, all["b"]) // Child value
	assert.Equal(t, 3, all["c"])
}
