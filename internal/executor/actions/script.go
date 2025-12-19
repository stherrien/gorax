package actions

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/dop251/goja"
)

// ScriptAction implements the Action interface for sandboxed JavaScript execution
//
// Security Considerations:
// - Uses goja runtime which is completely sandboxed (no file system, network, or goroutines)
// - Enforces execution timeout to prevent infinite loops
// - Memory limits can be configured (future enhancement with goja runtime options)
// - No access to Go stdlib or external modules unless explicitly provided
// - Scripts run in isolated VM instances (no shared state between executions)
type ScriptAction struct{}

// ScriptActionConfig represents the configuration for a script action
type ScriptActionConfig struct {
	Script      string `json:"script"`                 // JavaScript code to execute
	Timeout     int    `json:"timeout,omitempty"`      // Max execution time in seconds (default: 30)
	MemoryLimit int    `json:"memory_limit,omitempty"` // Max memory in MB (future enhancement)
}

// ScriptActionResult represents the result of a script execution
type ScriptActionResult struct {
	Result interface{} `json:"result"` // The value returned by the script
}

const (
	defaultScriptTimeout = 30 // seconds
)

// Execute implements the Action interface
func (a *ScriptAction) Execute(ctx context.Context, input *ActionInput) (*ActionOutput, error) {
	startTime := time.Now()

	// Parse config
	configBytes, err := json.Marshal(input.Config)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal config: %w", err)
	}

	var config ScriptActionConfig
	if err := json.Unmarshal(configBytes, &config); err != nil {
		return nil, fmt.Errorf("failed to parse script action config: %w", err)
	}

	// Validate script
	if config.Script == "" {
		return nil, fmt.Errorf("script is required")
	}

	// Determine timeout
	timeout := time.Duration(config.Timeout) * time.Second
	if timeout <= 0 {
		timeout = defaultScriptTimeout * time.Second
	}

	// Execute script with timeout
	result, err := a.executeScript(ctx, config.Script, input.Context, timeout)
	if err != nil {
		return nil, err
	}

	// Calculate execution time
	executionTime := time.Since(startTime).Milliseconds()

	// Create output
	output := NewActionOutput(&ScriptActionResult{
		Result: result,
	})
	output.WithMetadata("execution_time_ms", executionTime)

	return output, nil
}

// executeScript executes JavaScript code in a sandboxed environment
func (a *ScriptAction) executeScript(ctx context.Context, script string, execContext map[string]interface{}, timeout time.Duration) (interface{}, error) {
	// Create new goja runtime (isolated sandbox)
	vm := goja.New()

	// Set up timeout mechanism
	timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Channel to receive result or error
	type result struct {
		value interface{}
		err   error
	}
	resultChan := make(chan result, 1)

	// Execute script in goroutine to support timeout
	go func() {
		defer func() {
			if r := recover(); r != nil {
				resultChan <- result{
					err: fmt.Errorf("script panic: %v", r),
				}
			}
		}()

		// Inject context into the VM
		if err := a.injectContext(vm, execContext); err != nil {
			resultChan <- result{err: err}
			return
		}

		// Enable interrupts for timeout support
		vm.SetMaxCallStackSize(1000) // Prevent stack overflow attacks

		// Wrap script in an IIFE (Immediately Invoked Function Expression)
		// This allows 'return' statements to work properly
		wrappedScript := "(function() {\n" + script + "\n})();"

		// Run the script
		val, err := vm.RunString(wrappedScript)
		if err != nil {
			resultChan <- result{err: fmt.Errorf("script execution failed: %w", err)}
			return
		}

		// Export result to Go value
		exported := a.exportValue(val)
		resultChan <- result{value: exported}
	}()

	// Wait for result or timeout
	select {
	case res := <-resultChan:
		return res.value, res.err
	case <-timeoutCtx.Done():
		// Attempt to interrupt the VM
		vm.Interrupt("execution timeout")
		return nil, fmt.Errorf("script execution timeout after %v", timeout)
	}
}

// injectContext injects the execution context into the JavaScript VM
func (a *ScriptAction) injectContext(vm *goja.Runtime, execContext map[string]interface{}) error {
	// Create context object in JavaScript
	contextObj := vm.NewObject()

	// Convert execContext map to goja values
	for key, value := range execContext {
		gojaValue := vm.ToValue(value)
		if err := contextObj.Set(key, gojaValue); err != nil {
			return fmt.Errorf("failed to set context.%s: %w", key, err)
		}
	}

	// Set context as global variable
	if err := vm.Set("context", contextObj); err != nil {
		return fmt.Errorf("failed to set context global: %w", err)
	}

	return nil
}

// exportValue converts a goja.Value to a Go interface{} value
func (a *ScriptAction) exportValue(val goja.Value) interface{} {
	if val == nil || goja.IsUndefined(val) || goja.IsNull(val) {
		return nil
	}

	exported := val.Export()

	// Handle special cases for better JSON serialization
	switch v := exported.(type) {
	case map[string]interface{}:
		// Already correct type
		return v
	case []interface{}:
		// Already correct type
		return v
	case string, int, int32, int64, float32, float64, bool:
		// Primitive types
		return v
	default:
		// For other types, try to convert to primitive
		return exported
	}
}

// Helper function for creating script actions (for testing)
func NewScriptAction() *ScriptAction {
	return &ScriptAction{}
}
