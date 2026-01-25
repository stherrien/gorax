package javascript

import (
	"testing"

	"github.com/dop251/goja"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSandbox_ApplyToRuntime(t *testing.T) {
	sandbox := NewSandbox(nil)
	vm := goja.New()

	err := sandbox.ApplyToRuntime(vm)
	require.NoError(t, err)

	// Verify console is available
	consoleVal := vm.Get("console")
	require.NotNil(t, consoleVal)
	assert.False(t, goja.IsUndefined(consoleVal))
}

func TestSandbox_RemovesForbiddenGlobals(t *testing.T) {
	// Note: Goja doesn't have Node.js globals by default (no require, process, etc.)
	// This test verifies that if globals are set, they are properly removed
	vm := goja.New()

	// Manually set some globals that should be removed
	_ = vm.Set("testForbidden1", "value1")
	_ = vm.Set("testForbidden2", "value2")

	// Create a config that includes these test globals as forbidden
	config := &SandboxConfig{
		DisableEval:      true,
		MaxCallStackSize: DefaultMaxCallStackSize,
		AllowedGlobals:   AllowedGlobals,
		ForbiddenGlobals: []string{"testForbidden1", "testForbidden2"},
		CustomGlobals:    make(map[string]any),
	}
	sandbox := NewSandbox(config)

	err := sandbox.ApplyToRuntime(vm)
	require.NoError(t, err)

	// Verify the test globals are removed
	val := vm.Get("testForbidden1")
	assert.True(t, goja.IsUndefined(val), "testForbidden1 should be undefined")

	val = vm.Get("testForbidden2")
	assert.True(t, goja.IsUndefined(val), "testForbidden2 should be undefined")
}

func TestSandbox_PreservesAllowedGlobals(t *testing.T) {
	sandbox := NewSandbox(nil)
	vm := goja.New()

	err := sandbox.ApplyToRuntime(vm)
	require.NoError(t, err)

	// Check allowed globals are still available
	allowedTests := []string{
		"Object",
		"Array",
		"String",
		"Number",
		"Boolean",
		"Math",
		"Date",
		"JSON",
		"RegExp",
		"parseInt",
		"parseFloat",
	}

	for _, name := range allowedTests {
		t.Run(name, func(t *testing.T) {
			val := vm.Get(name)
			assert.False(t, goja.IsUndefined(val), "%s should be defined", name)
		})
	}
}

func TestSandbox_ValidateScript_AllowsValidScript(t *testing.T) {
	sandbox := NewSandbox(nil)

	validScripts := []struct {
		name   string
		script string
	}{
		{"simple return", "return 42;"},
		{"variable declaration", "var x = 1; return x;"},
		{"function definition", "function add(a, b) { return a + b; } return add(1, 2);"},
		{"object literal", "return { name: 'test' };"},
		{"array literal", "return [1, 2, 3];"},
		{"math operation", "return Math.sqrt(16);"},
	}

	for _, tc := range validScripts {
		t.Run(tc.name, func(t *testing.T) {
			err := sandbox.ValidateScript(tc.script)
			assert.NoError(t, err)
		})
	}
}

func TestSandbox_ValidateScript_RejectsForbiddenPatterns(t *testing.T) {
	sandbox := NewSandbox(nil)

	forbiddenScripts := []struct {
		name   string
		script string
	}{
		{"eval", `eval("alert(1)")`},
		{"new Function", `new Function("return 1")()`},
		{"constructor access", `var x = {}; x.constructor("alert(1)")()`},
		{"__proto__", `obj.__proto__.constructor`},
	}

	for _, tc := range forbiddenScripts {
		t.Run(tc.name, func(t *testing.T) {
			err := sandbox.ValidateScript(tc.script)
			assert.Error(t, err)
			assert.True(t, IsSandboxViolation(err))
		})
	}
}

func TestSandbox_CustomGlobals(t *testing.T) {
	config := &SandboxConfig{
		DisableEval:      true,
		MaxCallStackSize: 1000,
		AllowedGlobals:   AllowedGlobals,
		ForbiddenGlobals: ForbiddenGlobals,
		CustomGlobals: map[string]any{
			"customValue": 42,
			"customFunc": func(x int) int {
				return x * 2
			},
		},
	}

	sandbox := NewSandbox(config)
	vm := goja.New()

	err := sandbox.ApplyToRuntime(vm)
	require.NoError(t, err)

	// Check custom value
	val := vm.Get("customValue")
	require.False(t, goja.IsUndefined(val))
	assert.Equal(t, int64(42), val.Export())
}

func TestSandbox_MaxCallStackSize(t *testing.T) {
	config := &SandboxConfig{
		MaxCallStackSize: 100,
	}

	sandbox := NewSandbox(config)
	vm := goja.New()

	err := sandbox.ApplyToRuntime(vm)
	require.NoError(t, err)

	// This should panic/error due to stack overflow
	script := `
		function recurse(n) {
			if (n > 0) return recurse(n - 1);
			return 0;
		}
		recurse(1000);
	`

	_, err = vm.RunString(script)
	assert.Error(t, err)
}

func TestConsoleCapture_CapturesLogs(t *testing.T) {
	capture := NewConsoleCapture()
	vm := goja.New()

	err := capture.InstallInRuntime(vm)
	require.NoError(t, err)

	_, err = vm.RunString(`
		console.log("test message");
		console.warn("warning message");
		console.error("error message");
	`)
	require.NoError(t, err)

	logs := capture.GetLogs()
	require.Len(t, logs, 3)

	assert.Equal(t, "info", logs[0].Level)
	assert.Contains(t, logs[0].Message, "test message")

	assert.Equal(t, "warn", logs[1].Level)
	assert.Contains(t, logs[1].Message, "warning message")

	assert.Equal(t, "error", logs[2].Level)
	assert.Contains(t, logs[2].Message, "error message")
}

func TestConsoleCapture_CapturesMultipleArgs(t *testing.T) {
	capture := NewConsoleCapture()
	vm := goja.New()

	err := capture.InstallInRuntime(vm)
	require.NoError(t, err)

	_, err = vm.RunString(`
		console.log("Hello", "World", 123);
	`)
	require.NoError(t, err)

	logs := capture.GetLogs()
	require.Len(t, logs, 1)

	assert.Contains(t, logs[0].Message, "Hello")
	assert.Contains(t, logs[0].Message, "World")
	assert.Contains(t, logs[0].Message, "123")
	assert.Len(t, logs[0].Args, 3)
}

func TestConsoleCapture_Clear(t *testing.T) {
	capture := NewConsoleCapture()
	vm := goja.New()

	err := capture.InstallInRuntime(vm)
	require.NoError(t, err)

	_, err = vm.RunString(`console.log("test");`)
	require.NoError(t, err)

	require.Len(t, capture.GetLogs(), 1)

	capture.Clear()
	assert.Len(t, capture.GetLogs(), 0)
}

func TestDefaultSandboxConfig(t *testing.T) {
	config := DefaultSandboxConfig()

	assert.True(t, config.DisableEval)
	assert.Equal(t, DefaultMaxCallStackSize, config.MaxCallStackSize)
	assert.NotEmpty(t, config.AllowedGlobals)
	assert.NotEmpty(t, config.ForbiddenGlobals)
	assert.NotNil(t, config.CustomGlobals)
}
