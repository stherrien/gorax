package javascript

import (
	"encoding/json"
	"fmt"

	"github.com/dop251/goja"
)

// ExecutionContext holds the data to be injected into the JavaScript runtime.
type ExecutionContext struct {
	// Trigger contains the data that triggered the workflow.
	Trigger map[string]any `json:"trigger,omitempty"`

	// Steps contains outputs from previous workflow steps.
	Steps map[string]any `json:"steps,omitempty"`

	// Env contains environment variables and execution metadata.
	Env map[string]any `json:"env,omitempty"`

	// Vars contains user-defined workflow variables.
	Vars map[string]any `json:"vars,omitempty"`

	// Input contains direct input for script execution.
	Input map[string]any `json:"input,omitempty"`
}

// NewExecutionContext creates a new empty execution context.
func NewExecutionContext() *ExecutionContext {
	return &ExecutionContext{
		Trigger: make(map[string]any),
		Steps:   make(map[string]any),
		Env:     make(map[string]any),
		Vars:    make(map[string]any),
		Input:   make(map[string]any),
	}
}

// WithTrigger sets the trigger data.
func (c *ExecutionContext) WithTrigger(data map[string]any) *ExecutionContext {
	if data != nil {
		c.Trigger = data
	}
	return c
}

// WithSteps sets the steps data.
func (c *ExecutionContext) WithSteps(steps map[string]any) *ExecutionContext {
	if steps != nil {
		c.Steps = steps
	}
	return c
}

// WithEnv sets the environment data.
func (c *ExecutionContext) WithEnv(env map[string]any) *ExecutionContext {
	if env != nil {
		c.Env = env
	}
	return c
}

// WithVars sets the workflow variables.
func (c *ExecutionContext) WithVars(vars map[string]any) *ExecutionContext {
	if vars != nil {
		c.Vars = vars
	}
	return c
}

// WithInput sets the direct input data.
func (c *ExecutionContext) WithInput(input map[string]any) *ExecutionContext {
	if input != nil {
		c.Input = input
	}
	return c
}

// ToMap converts the context to a map suitable for injection.
func (c *ExecutionContext) ToMap() map[string]any {
	return map[string]any{
		"trigger": c.Trigger,
		"steps":   c.Steps,
		"env":     c.Env,
		"vars":    c.Vars,
		"input":   c.Input,
	}
}

// FromWorkflowContext creates an ExecutionContext from a workflow context map.
func FromWorkflowContext(ctx map[string]any) *ExecutionContext {
	execCtx := NewExecutionContext()

	if trigger, ok := ctx["trigger"].(map[string]any); ok {
		execCtx.Trigger = trigger
	}
	if steps, ok := ctx["steps"].(map[string]any); ok {
		execCtx.Steps = steps
	}
	if env, ok := ctx["env"].(map[string]any); ok {
		execCtx.Env = env
	}
	if vars, ok := ctx["vars"].(map[string]any); ok {
		execCtx.Vars = vars
	}
	if input, ok := ctx["input"].(map[string]any); ok {
		execCtx.Input = input
	}

	return execCtx
}

// ContextInjector handles injecting context data into the JavaScript runtime.
type ContextInjector struct {
	sanitizer *ValueSanitizer
}

// NewContextInjector creates a new context injector.
func NewContextInjector() *ContextInjector {
	return &ContextInjector{
		sanitizer: NewValueSanitizer(),
	}
}

// InjectContext injects the execution context into the JavaScript runtime.
func (ci *ContextInjector) InjectContext(vm *goja.Runtime, ctx *ExecutionContext) error {
	if ctx == nil {
		return ErrNilContext
	}

	// Create context object
	contextObj := vm.NewObject()

	// Inject each section
	if err := ci.injectSection(vm, contextObj, "trigger", ctx.Trigger); err != nil {
		return fmt.Errorf("failed to inject trigger: %w", err)
	}
	if err := ci.injectSection(vm, contextObj, "steps", ctx.Steps); err != nil {
		return fmt.Errorf("failed to inject steps: %w", err)
	}
	if err := ci.injectSection(vm, contextObj, "env", ctx.Env); err != nil {
		return fmt.Errorf("failed to inject env: %w", err)
	}
	if err := ci.injectSection(vm, contextObj, "vars", ctx.Vars); err != nil {
		return fmt.Errorf("failed to inject vars: %w", err)
	}
	if err := ci.injectSection(vm, contextObj, "input", ctx.Input); err != nil {
		return fmt.Errorf("failed to inject input: %w", err)
	}

	// Set context as global variable
	if err := vm.Set("context", contextObj); err != nil {
		return fmt.Errorf("failed to set context global: %w", err)
	}

	// Also set shorthand aliases for convenience
	if err := vm.Set("ctx", contextObj); err != nil {
		return fmt.Errorf("failed to set ctx alias: %w", err)
	}

	return nil
}

// injectSection injects a single section into the context object.
func (ci *ContextInjector) injectSection(vm *goja.Runtime, contextObj *goja.Object, name string, data map[string]any) error {
	sectionObj := vm.NewObject()

	for key, value := range data {
		sanitized := ci.sanitizer.Sanitize(value)
		gojaValue := vm.ToValue(sanitized)
		if err := sectionObj.Set(key, gojaValue); err != nil {
			return fmt.Errorf("failed to set %s.%s: %w", name, key, err)
		}
	}

	if err := contextObj.Set(name, sectionObj); err != nil {
		return fmt.Errorf("failed to set context.%s: %w", name, err)
	}

	return nil
}

// ResultExtractor handles extracting results from the JavaScript runtime.
type ResultExtractor struct {
	sanitizer *ValueSanitizer
}

// NewResultExtractor creates a new result extractor.
func NewResultExtractor() *ResultExtractor {
	return &ResultExtractor{
		sanitizer: NewValueSanitizer(),
	}
}

// ExtractResult converts a Goja value to a Go value.
func (re *ResultExtractor) ExtractResult(val goja.Value) (any, error) {
	if val == nil || goja.IsUndefined(val) || goja.IsNull(val) {
		return nil, nil
	}

	exported := val.Export()
	return re.sanitizer.Sanitize(exported), nil
}

// ExtractJSON converts a Goja value to JSON bytes.
func (re *ResultExtractor) ExtractJSON(val goja.Value) ([]byte, error) {
	result, err := re.ExtractResult(val)
	if err != nil {
		return nil, err
	}

	return json.Marshal(result)
}

// ValueSanitizer ensures values are safe for JSON serialization and Go usage.
type ValueSanitizer struct {
	maxDepth int
}

// NewValueSanitizer creates a new value sanitizer.
func NewValueSanitizer() *ValueSanitizer {
	return &ValueSanitizer{
		maxDepth: 100, // Prevent circular reference infinite recursion
	}
}

// Sanitize converts a value to a safe, serializable form.
func (vs *ValueSanitizer) Sanitize(value any) any {
	return vs.sanitizeWithDepth(value, 0)
}

func (vs *ValueSanitizer) sanitizeWithDepth(value any, depth int) any {
	if depth > vs.maxDepth {
		return nil // Prevent infinite recursion
	}

	if value == nil {
		return nil
	}

	switch v := value.(type) {
	case string, bool, int, int8, int16, int32, int64,
		uint, uint8, uint16, uint32, uint64,
		float32, float64:
		return v

	case map[string]any:
		result := make(map[string]any, len(v))
		for key, val := range v {
			result[key] = vs.sanitizeWithDepth(val, depth+1)
		}
		return result

	case []any:
		result := make([]any, len(v))
		for i, val := range v {
			result[i] = vs.sanitizeWithDepth(val, depth+1)
		}
		return result

	case []map[string]any:
		result := make([]any, len(v))
		for i, val := range v {
			result[i] = vs.sanitizeWithDepth(val, depth+1)
		}
		return result

	default:
		// Try to convert via JSON round-trip for complex types
		jsonBytes, err := json.Marshal(v)
		if err != nil {
			return fmt.Sprintf("%v", v)
		}

		var result any
		if err := json.Unmarshal(jsonBytes, &result); err != nil {
			return fmt.Sprintf("%v", v)
		}
		return vs.sanitizeWithDepth(result, depth+1)
	}
}

// VariableScope manages variable scoping and isolation.
type VariableScope struct {
	parent *VariableScope
	vars   map[string]any
}

// NewVariableScope creates a new variable scope.
func NewVariableScope(parent *VariableScope) *VariableScope {
	return &VariableScope{
		parent: parent,
		vars:   make(map[string]any),
	}
}

// Set sets a variable in the current scope.
func (s *VariableScope) Set(name string, value any) {
	s.vars[name] = value
}

// Get retrieves a variable, searching parent scopes if not found.
func (s *VariableScope) Get(name string) (any, bool) {
	if val, ok := s.vars[name]; ok {
		return val, true
	}
	if s.parent != nil {
		return s.parent.Get(name)
	}
	return nil, false
}

// GetLocal retrieves a variable from only the current scope.
func (s *VariableScope) GetLocal(name string) (any, bool) {
	val, ok := s.vars[name]
	return val, ok
}

// Delete removes a variable from the current scope.
func (s *VariableScope) Delete(name string) {
	delete(s.vars, name)
}

// All returns all variables in scope, including parent scopes.
func (s *VariableScope) All() map[string]any {
	result := make(map[string]any)

	// First add parent variables
	if s.parent != nil {
		for k, v := range s.parent.All() {
			result[k] = v
		}
	}

	// Then override with local variables
	for k, v := range s.vars {
		result[k] = v
	}

	return result
}
