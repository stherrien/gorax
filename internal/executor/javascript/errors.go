package javascript

import (
	"errors"
	"fmt"
)

// Sentinel errors for JavaScript execution.
var (
	// Resource limit errors.
	ErrTimeout             = errors.New("script execution timeout")
	ErrMemoryLimitExceeded = errors.New("memory limit exceeded")
	ErrScriptTooLarge      = errors.New("script size exceeds maximum limit")
	ErrStackOverflow       = errors.New("stack overflow detected")

	// Validation errors.
	ErrEmptyScript         = errors.New("script is empty")
	ErrInvalidTimeout      = errors.New("timeout must be positive")
	ErrTimeoutExceedsMax   = errors.New("timeout exceeds maximum allowed")
	ErrInvalidStackSize    = errors.New("max call stack size must be positive")
	ErrInvalidMemoryLimit  = errors.New("max memory must be positive")
	ErrInvalidScriptLength = errors.New("max script length must be positive")

	// Sandbox errors.
	ErrSandboxViolation   = errors.New("sandbox security violation")
	ErrForbiddenGlobal    = errors.New("access to forbidden global")
	ErrForbiddenOperation = errors.New("forbidden operation attempted")

	// VM pool errors.
	ErrPoolExhausted = errors.New("VM pool exhausted")
	ErrPoolClosed    = errors.New("VM pool is closed")

	// Context errors.
	ErrNilContext       = errors.New("execution context is nil")
	ErrContextInjection = errors.New("failed to inject context")

	// Execution errors.
	ErrCompilation      = errors.New("script compilation failed")
	ErrExecution        = errors.New("script execution failed")
	ErrResultExtraction = errors.New("failed to extract result")
	ErrInterrupted      = errors.New("execution was interrupted")
)

// ExecutionError wraps execution errors with additional context.
type ExecutionError struct {
	// Err is the underlying error.
	Err error

	// Phase indicates when the error occurred.
	Phase ExecutionPhase

	// ScriptLine is the line number where the error occurred (if available).
	ScriptLine int

	// Details contains additional context about the error.
	Details map[string]interface{}
}

// ExecutionPhase represents the phase of execution where an error occurred.
type ExecutionPhase string

const (
	PhaseValidation  ExecutionPhase = "validation"
	PhaseCompilation ExecutionPhase = "compilation"
	PhaseExecution   ExecutionPhase = "execution"
	PhaseExtraction  ExecutionPhase = "extraction"
)

// Error implements the error interface.
func (e *ExecutionError) Error() string {
	if e.ScriptLine > 0 {
		return fmt.Sprintf("%s error at line %d: %v", e.Phase, e.ScriptLine, e.Err)
	}
	return fmt.Sprintf("%s error: %v", e.Phase, e.Err)
}

// Unwrap returns the underlying error.
func (e *ExecutionError) Unwrap() error {
	return e.Err
}

// NewExecutionError creates a new execution error.
func NewExecutionError(err error, phase ExecutionPhase) *ExecutionError {
	return &ExecutionError{
		Err:     err,
		Phase:   phase,
		Details: make(map[string]interface{}),
	}
}

// WithLine adds line number information to the error.
func (e *ExecutionError) WithLine(line int) *ExecutionError {
	e.ScriptLine = line
	return e
}

// WithDetail adds a detail to the error.
func (e *ExecutionError) WithDetail(key string, value interface{}) *ExecutionError {
	e.Details[key] = value
	return e
}

// IsTimeout checks if the error is a timeout error.
func IsTimeout(err error) bool {
	return errors.Is(err, ErrTimeout)
}

// IsMemoryLimit checks if the error is a memory limit error.
func IsMemoryLimit(err error) bool {
	return errors.Is(err, ErrMemoryLimitExceeded)
}

// IsSandboxViolation checks if the error is a sandbox violation.
func IsSandboxViolation(err error) bool {
	return errors.Is(err, ErrSandboxViolation) ||
		errors.Is(err, ErrForbiddenGlobal) ||
		errors.Is(err, ErrForbiddenOperation)
}

// IsResourceLimit checks if the error is any resource limit error.
func IsResourceLimit(err error) bool {
	return IsTimeout(err) ||
		IsMemoryLimit(err) ||
		errors.Is(err, ErrStackOverflow) ||
		errors.Is(err, ErrScriptTooLarge)
}

// WrapCompilationError wraps a compilation error with context.
func WrapCompilationError(err error) error {
	return NewExecutionError(
		fmt.Errorf("%w: %v", ErrCompilation, err),
		PhaseCompilation,
	)
}

// WrapExecutionError wraps a runtime error with context.
func WrapExecutionError(err error) error {
	return NewExecutionError(
		fmt.Errorf("%w: %v", ErrExecution, err),
		PhaseExecution,
	)
}

// WrapValidationError wraps a validation error with context.
func WrapValidationError(err error) error {
	return NewExecutionError(
		fmt.Errorf("validation failed: %w", err),
		PhaseValidation,
	)
}
