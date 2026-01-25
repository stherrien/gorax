package javascript

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExecutionError_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      *ExecutionError
		expected string
	}{
		{
			name: "without line",
			err: &ExecutionError{
				Err:   errors.New("test error"),
				Phase: PhaseExecution,
			},
			expected: "execution error: test error",
		},
		{
			name: "with line",
			err: &ExecutionError{
				Err:        errors.New("syntax error"),
				Phase:      PhaseCompilation,
				ScriptLine: 10,
			},
			expected: "compilation error at line 10: syntax error",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, tc.err.Error())
		})
	}
}

func TestExecutionError_Unwrap(t *testing.T) {
	innerErr := errors.New("inner error")
	execErr := NewExecutionError(innerErr, PhaseExecution)

	unwrapped := execErr.Unwrap()
	assert.Equal(t, innerErr, unwrapped)
}

func TestNewExecutionError(t *testing.T) {
	err := errors.New("test")
	execErr := NewExecutionError(err, PhaseValidation)

	assert.Equal(t, err, execErr.Err)
	assert.Equal(t, PhaseValidation, execErr.Phase)
	assert.NotNil(t, execErr.Details)
	assert.Equal(t, 0, execErr.ScriptLine)
}

func TestExecutionError_WithLine(t *testing.T) {
	execErr := NewExecutionError(errors.New("test"), PhaseCompilation).
		WithLine(42)

	assert.Equal(t, 42, execErr.ScriptLine)
}

func TestExecutionError_WithDetail(t *testing.T) {
	execErr := NewExecutionError(errors.New("test"), PhaseExecution).
		WithDetail("key1", "value1").
		WithDetail("key2", 123)

	assert.Equal(t, "value1", execErr.Details["key1"])
	assert.Equal(t, 123, execErr.Details["key2"])
}

func TestIsTimeout(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "timeout error",
			err:      ErrTimeout,
			expected: true,
		},
		{
			name:     "wrapped timeout error",
			err:      errors.Join(errors.New("wrapper"), ErrTimeout),
			expected: true,
		},
		{
			name:     "other error",
			err:      errors.New("some error"),
			expected: false,
		},
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := IsTimeout(tc.err)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestIsMemoryLimit(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "memory limit error",
			err:      ErrMemoryLimitExceeded,
			expected: true,
		},
		{
			name:     "other error",
			err:      ErrTimeout,
			expected: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := IsMemoryLimit(tc.err)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestIsSandboxViolation(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "sandbox violation",
			err:      ErrSandboxViolation,
			expected: true,
		},
		{
			name:     "forbidden global",
			err:      ErrForbiddenGlobal,
			expected: true,
		},
		{
			name:     "forbidden operation",
			err:      ErrForbiddenOperation,
			expected: true,
		},
		{
			name:     "timeout error",
			err:      ErrTimeout,
			expected: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := IsSandboxViolation(tc.err)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestIsResourceLimit(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "timeout",
			err:      ErrTimeout,
			expected: true,
		},
		{
			name:     "memory limit",
			err:      ErrMemoryLimitExceeded,
			expected: true,
		},
		{
			name:     "stack overflow",
			err:      ErrStackOverflow,
			expected: true,
		},
		{
			name:     "script too large",
			err:      ErrScriptTooLarge,
			expected: true,
		},
		{
			name:     "other error",
			err:      ErrEmptyScript,
			expected: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := IsResourceLimit(tc.err)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestWrapCompilationError(t *testing.T) {
	innerErr := errors.New("syntax error at line 5")
	wrapped := WrapCompilationError(innerErr)

	execErr, ok := wrapped.(*ExecutionError)
	assert.True(t, ok)
	assert.Equal(t, PhaseCompilation, execErr.Phase)
	assert.ErrorIs(t, wrapped, ErrCompilation)
}

func TestWrapExecutionError(t *testing.T) {
	innerErr := errors.New("undefined is not a function")
	wrapped := WrapExecutionError(innerErr)

	execErr, ok := wrapped.(*ExecutionError)
	assert.True(t, ok)
	assert.Equal(t, PhaseExecution, execErr.Phase)
	assert.ErrorIs(t, wrapped, ErrExecution)
}

func TestWrapValidationError(t *testing.T) {
	innerErr := errors.New("script is empty")
	wrapped := WrapValidationError(innerErr)

	execErr, ok := wrapped.(*ExecutionError)
	assert.True(t, ok)
	assert.Equal(t, PhaseValidation, execErr.Phase)
}

func TestExecutionPhaseConstants(t *testing.T) {
	assert.Equal(t, ExecutionPhase("validation"), PhaseValidation)
	assert.Equal(t, ExecutionPhase("compilation"), PhaseCompilation)
	assert.Equal(t, ExecutionPhase("execution"), PhaseExecution)
	assert.Equal(t, ExecutionPhase("extraction"), PhaseExtraction)
}
