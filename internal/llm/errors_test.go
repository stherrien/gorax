package llm

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLLMError_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      *LLMError
		expected string
	}{
		{
			name: "basic error",
			err: &LLMError{
				Provider: "openai",
				Code:     "invalid_request",
				Message:  "invalid parameters",
			},
			expected: "openai error [invalid_request]: invalid parameters",
		},
		{
			name: "error with retry after",
			err: &LLMError{
				Provider:   "anthropic",
				Code:       "rate_limited",
				Message:    "too many requests",
				RetryAfter: 30,
			},
			expected: "anthropic error [rate_limited]: too many requests (retry after 30s)",
		},
		{
			name: "error without code",
			err: &LLMError{
				Provider: "bedrock",
				Message:  "service unavailable",
			},
			expected: "bedrock error: service unavailable",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.err.Error())
		})
	}
}

func TestLLMError_Unwrap(t *testing.T) {
	cause := errors.New("underlying error")
	err := &LLMError{
		Provider: "openai",
		Code:     "test",
		Message:  "test error",
		Cause:    cause,
	}

	assert.Equal(t, cause, err.Unwrap())
	assert.True(t, errors.Is(err, cause))
}

func TestLLMError_IsRetryable(t *testing.T) {
	t.Run("retryable error", func(t *testing.T) {
		err := &LLMError{
			Provider:  "openai",
			Retryable: true,
		}
		assert.True(t, err.IsRetryable())
	})

	t.Run("non-retryable error", func(t *testing.T) {
		err := &LLMError{
			Provider:  "openai",
			Retryable: false,
		}
		assert.False(t, err.IsRetryable())
	})
}

func TestNewLLMError(t *testing.T) {
	cause := errors.New("test cause")
	err := NewLLMError("openai", "invalid_request", "test message", cause)

	assert.Equal(t, "openai", err.Provider)
	assert.Equal(t, "invalid_request", err.Code)
	assert.Equal(t, "test message", err.Message)
	assert.Equal(t, cause, err.Cause)
	assert.False(t, err.Retryable)
}

func TestNewRetryableLLMError(t *testing.T) {
	err := NewRetryableLLMError("anthropic", "rate_limited", "slow down", 60)

	assert.Equal(t, "anthropic", err.Provider)
	assert.Equal(t, "rate_limited", err.Code)
	assert.Equal(t, "slow down", err.Message)
	assert.Equal(t, 60, err.RetryAfter)
	assert.True(t, err.Retryable)
	assert.True(t, errors.Is(err, ErrRateLimitExceeded))
}

func TestIsRateLimitError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
		{
			name:     "ErrRateLimitExceeded",
			err:      ErrRateLimitExceeded,
			expected: true,
		},
		{
			name: "LLMError with rate_limit_exceeded code",
			err: &LLMError{
				Provider: "openai",
				Code:     "rate_limit_exceeded",
			},
			expected: true,
		},
		{
			name: "LLMError with 429 code",
			err: &LLMError{
				Provider: "openai",
				Code:     "429",
			},
			expected: true,
		},
		{
			name: "LLMError with 429 HTTP status",
			err: &LLMError{
				Provider:   "anthropic",
				HTTPStatus: 429,
			},
			expected: true,
		},
		{
			name: "LLMError with different error",
			err: &LLMError{
				Provider: "openai",
				Code:     "invalid_request",
			},
			expected: false,
		},
		{
			name:     "generic error",
			err:      errors.New("some error"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, IsRateLimitError(tt.err))
		})
	}
}

func TestIsAuthError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
		{
			name:     "ErrInvalidAPIKey",
			err:      ErrInvalidAPIKey,
			expected: true,
		},
		{
			name:     "ErrUnauthorized",
			err:      ErrUnauthorized,
			expected: true,
		},
		{
			name: "LLMError with invalid_api_key code",
			err: &LLMError{
				Provider: "openai",
				Code:     "invalid_api_key",
			},
			expected: true,
		},
		{
			name: "LLMError with 401 status",
			err: &LLMError{
				Provider:   "anthropic",
				HTTPStatus: 401,
			},
			expected: true,
		},
		{
			name: "LLMError with 403 status",
			err: &LLMError{
				Provider:   "bedrock",
				HTTPStatus: 403,
			},
			expected: true,
		},
		{
			name: "LLMError with different error",
			err: &LLMError{
				Provider: "openai",
				Code:     "rate_limited",
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, IsAuthError(tt.err))
		})
	}
}

func TestIsContextLengthError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
		{
			name:     "ErrContextTooLong",
			err:      ErrContextTooLong,
			expected: true,
		},
		{
			name:     "ErrTokenLimitExceeded",
			err:      ErrTokenLimitExceeded,
			expected: true,
		},
		{
			name: "LLMError with context_length_exceeded code",
			err: &LLMError{
				Provider: "openai",
				Code:     "context_length_exceeded",
			},
			expected: true,
		},
		{
			name: "LLMError with max_tokens_exceeded code",
			err: &LLMError{
				Provider: "anthropic",
				Code:     "max_tokens_exceeded",
			},
			expected: true,
		},
		{
			name: "different error",
			err: &LLMError{
				Provider: "openai",
				Code:     "invalid_request",
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, IsContextLengthError(tt.err))
		})
	}
}

func TestIsRetryableError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
		{
			name:     "ErrRateLimitExceeded",
			err:      ErrRateLimitExceeded,
			expected: true,
		},
		{
			name:     "ErrProviderUnavailable",
			err:      ErrProviderUnavailable,
			expected: true,
		},
		{
			name:     "ErrProviderTimeout",
			err:      ErrProviderTimeout,
			expected: true,
		},
		{
			name: "LLMError marked retryable",
			err: &LLMError{
				Provider:  "openai",
				Retryable: true,
			},
			expected: true,
		},
		{
			name: "LLMError not retryable",
			err: &LLMError{
				Provider:  "openai",
				Retryable: false,
			},
			expected: false,
		},
		{
			name:     "ErrInvalidAPIKey - not retryable",
			err:      ErrInvalidAPIKey,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, IsRetryableError(tt.err))
		})
	}
}

func TestGetRetryAfter(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected int
	}{
		{
			name:     "nil error",
			err:      nil,
			expected: 0,
		},
		{
			name: "LLMError with retry after",
			err: &LLMError{
				Provider:   "openai",
				RetryAfter: 30,
			},
			expected: 30,
		},
		{
			name: "LLMError without retry after",
			err: &LLMError{
				Provider: "openai",
			},
			expected: 0,
		},
		{
			name:     "generic error",
			err:      errors.New("some error"),
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, GetRetryAfter(tt.err))
		})
	}
}
