package javascript

import (
	"fmt"
	"strings"

	"github.com/dop251/goja"
)

// ForbiddenGlobals lists globals that must be removed from the sandbox.
var ForbiddenGlobals = []string{
	// Node.js specific
	"require",
	"module",
	"exports",
	"__dirname",
	"__filename",
	"process",
	"Buffer",
	"global",
	"globalThis",

	// Browser specific
	"window",
	"document",
	"location",
	"navigator",
	"fetch",
	"XMLHttpRequest",
	"WebSocket",

	// Dangerous operations
	"eval",
	"Function",
}

// AllowedGlobals lists globals that are safe to use.
var AllowedGlobals = []string{
	// Standard built-ins
	"Object",
	"Array",
	"String",
	"Number",
	"Boolean",
	"Math",
	"Date",
	"JSON",
	"RegExp",
	"Error",
	"TypeError",
	"RangeError",
	"SyntaxError",
	"ReferenceError",
	"URIError",
	"EvalError",
	"parseInt",
	"parseFloat",
	"isNaN",
	"isFinite",
	"encodeURI",
	"decodeURI",
	"encodeURIComponent",
	"decodeURIComponent",
	"undefined",
	"NaN",
	"Infinity",
	"console",
	"Map",
	"Set",
	"WeakMap",
	"WeakSet",
	"Promise",
	"Symbol",
	"Proxy",
	"Reflect",
	"ArrayBuffer",
	"DataView",
	"Int8Array",
	"Uint8Array",
	"Uint8ClampedArray",
	"Int16Array",
	"Uint16Array",
	"Int32Array",
	"Uint32Array",
	"Float32Array",
	"Float64Array",
	"BigInt",
	"BigInt64Array",
	"BigUint64Array",
}

// SandboxConfig holds configuration for the sandbox environment.
type SandboxConfig struct {
	// DisableEval prevents use of eval() and new Function().
	DisableEval bool

	// MaxCallStackSize limits recursion depth.
	MaxCallStackSize int

	// AllowedGlobals overrides the default allowed globals.
	AllowedGlobals []string

	// ForbiddenGlobals extends the default forbidden globals.
	ForbiddenGlobals []string

	// CustomGlobals adds custom globals to the sandbox.
	CustomGlobals map[string]interface{}
}

// DefaultSandboxConfig returns the default sandbox configuration.
func DefaultSandboxConfig() *SandboxConfig {
	return &SandboxConfig{
		DisableEval:      true,
		MaxCallStackSize: DefaultMaxCallStackSize,
		AllowedGlobals:   AllowedGlobals,
		ForbiddenGlobals: ForbiddenGlobals,
		CustomGlobals:    make(map[string]interface{}),
	}
}

// Sandbox provides a secure execution environment for JavaScript.
type Sandbox struct {
	config *SandboxConfig
}

// NewSandbox creates a new sandbox with the given configuration.
func NewSandbox(config *SandboxConfig) *Sandbox {
	if config == nil {
		config = DefaultSandboxConfig()
	}
	return &Sandbox{config: config}
}

// ApplyToRuntime applies sandbox restrictions to a Goja runtime.
func (s *Sandbox) ApplyToRuntime(vm *goja.Runtime) error {
	// Set call stack limit to prevent stack overflow attacks
	vm.SetMaxCallStackSize(s.config.MaxCallStackSize)

	// Remove forbidden globals
	if err := s.removeForbiddenGlobals(vm); err != nil {
		return fmt.Errorf("failed to remove forbidden globals: %w", err)
	}

	// Add safe console implementation
	if err := s.addSafeConsole(vm); err != nil {
		return fmt.Errorf("failed to add safe console: %w", err)
	}

	// Add custom globals
	if err := s.addCustomGlobals(vm); err != nil {
		return fmt.Errorf("failed to add custom globals: %w", err)
	}

	return nil
}

// removeForbiddenGlobals removes dangerous globals from the runtime.
func (s *Sandbox) removeForbiddenGlobals(vm *goja.Runtime) error {
	// Combine default and custom forbidden globals
	forbidden := make(map[string]struct{})
	for _, name := range ForbiddenGlobals {
		forbidden[name] = struct{}{}
	}
	for _, name := range s.config.ForbiddenGlobals {
		forbidden[name] = struct{}{}
	}

	// Set forbidden globals to undefined to effectively remove them
	for name := range forbidden {
		// Check if the global exists before trying to remove it
		val := vm.Get(name)
		if val != nil && !goja.IsUndefined(val) {
			// Set to undefined to remove access
			if err := vm.Set(name, goja.Undefined()); err != nil {
				// Continue on error - some globals may not be settable
				continue
			}
		}
	}

	return nil
}

// addSafeConsole adds a safe console implementation that captures output.
func (s *Sandbox) addSafeConsole(vm *goja.Runtime) error {
	console := vm.NewObject()

	// No-op implementations for console methods
	// Actual logging is handled by ConsoleCapture wrapper
	noop := func(call goja.FunctionCall) goja.Value {
		return goja.Undefined()
	}

	methods := []string{"log", "info", "warn", "error", "debug", "trace"}
	for _, method := range methods {
		if err := console.Set(method, noop); err != nil {
			return fmt.Errorf("failed to set console.%s: %w", method, err)
		}
	}

	if err := vm.Set("console", console); err != nil {
		return fmt.Errorf("failed to set console: %w", err)
	}

	return nil
}

// addCustomGlobals adds custom globals to the runtime.
func (s *Sandbox) addCustomGlobals(vm *goja.Runtime) error {
	for name, value := range s.config.CustomGlobals {
		if err := vm.Set(name, value); err != nil {
			return fmt.Errorf("failed to set custom global %s: %w", name, err)
		}
	}
	return nil
}

// ValidateScript checks a script for forbidden patterns.
func (s *Sandbox) ValidateScript(script string) error {
	// Check for potentially dangerous patterns
	dangerousPatterns := []string{
		"new Function",
		"eval(",
		"constructor[",
		".constructor(",
		"__proto__",
		"prototype.constructor",
	}

	scriptLower := strings.ToLower(script)
	for _, pattern := range dangerousPatterns {
		if strings.Contains(scriptLower, strings.ToLower(pattern)) {
			return fmt.Errorf("%w: script contains forbidden pattern: %s",
				ErrForbiddenOperation, pattern)
		}
	}

	return nil
}

// ConsoleCapture captures console output from script execution.
type ConsoleCapture struct {
	logs []ConsoleEntry
}

// ConsoleEntry represents a single console log entry.
type ConsoleEntry struct {
	Level   string        `json:"level"`
	Message string        `json:"message"`
	Args    []interface{} `json:"args,omitempty"`
}

// NewConsoleCapture creates a new console capture instance.
func NewConsoleCapture() *ConsoleCapture {
	return &ConsoleCapture{
		logs: make([]ConsoleEntry, 0),
	}
}

// InstallInRuntime installs capturing console methods in the runtime.
func (c *ConsoleCapture) InstallInRuntime(vm *goja.Runtime) error {
	console := vm.NewObject()

	makeLogger := func(level string) func(goja.FunctionCall) goja.Value {
		return func(call goja.FunctionCall) goja.Value {
			args := make([]interface{}, len(call.Arguments))
			for i, arg := range call.Arguments {
				args[i] = arg.Export()
			}

			message := formatConsoleArgs(args)
			c.logs = append(c.logs, ConsoleEntry{
				Level:   level,
				Message: message,
				Args:    args,
			})
			return goja.Undefined()
		}
	}

	methods := map[string]string{
		"log":   "info",
		"info":  "info",
		"warn":  "warn",
		"error": "error",
		"debug": "debug",
		"trace": "trace",
	}

	for method, level := range methods {
		if err := console.Set(method, makeLogger(level)); err != nil {
			return fmt.Errorf("failed to set console.%s: %w", method, err)
		}
	}

	if err := vm.Set("console", console); err != nil {
		return fmt.Errorf("failed to set console: %w", err)
	}

	return nil
}

// GetLogs returns all captured console logs.
func (c *ConsoleCapture) GetLogs() []ConsoleEntry {
	return c.logs
}

// Clear clears all captured logs.
func (c *ConsoleCapture) Clear() {
	c.logs = c.logs[:0]
}

// formatConsoleArgs formats console arguments into a string.
func formatConsoleArgs(args []interface{}) string {
	if len(args) == 0 {
		return ""
	}

	parts := make([]string, len(args))
	for i, arg := range args {
		parts[i] = fmt.Sprintf("%v", arg)
	}
	return strings.Join(parts, " ")
}
