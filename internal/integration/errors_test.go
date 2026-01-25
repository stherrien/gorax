package integration

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExecutionError(t *testing.T) {
	tests := []struct {
		name      string
		err       *ExecutionError
		wantMsg   string
		wantRetry bool
	}{
		{
			name: "retryable error",
			err: &ExecutionError{
				IntegrationName: "test",
				Operation:       "execute",
				Err:             errors.New("timeout"),
				Retryable:       true,
			},
			wantMsg:   "integration test: execute failed: timeout",
			wantRetry: true,
		},
		{
			name: "non-retryable error",
			err: &ExecutionError{
				IntegrationName: "test",
				Operation:       "validate",
				Err:             errors.New("invalid config"),
				Retryable:       false,
			},
			wantMsg:   "integration test: validate failed: invalid config",
			wantRetry: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.wantMsg, tt.err.Error())
			assert.Equal(t, tt.wantRetry, tt.err.IsRetryable())
			assert.NotNil(t, tt.err.Unwrap())
		})
	}
}

func TestValidationError(t *testing.T) {
	tests := []struct {
		name    string
		err     *ValidationError
		wantMsg string
	}{
		{
			name: "with field",
			err: &ValidationError{
				Field:   "url",
				Message: "URL is required",
			},
			wantMsg: "validation error for field 'url': URL is required",
		},
		{
			name: "without field",
			err: &ValidationError{
				Message: "general validation error",
			},
			wantMsg: "validation error: general validation error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.wantMsg, tt.err.Error())
		})
	}
}

func TestHTTPError(t *testing.T) {
	tests := []struct {
		name          string
		err           *HTTPError
		wantRetryable bool
		wantClient    bool
		wantServer    bool
	}{
		{
			name:          "server error 500",
			err:           &HTTPError{StatusCode: 500, Status: "Internal Server Error"},
			wantRetryable: true,
			wantClient:    false,
			wantServer:    true,
		},
		{
			name:          "rate limit 429",
			err:           &HTTPError{StatusCode: 429, Status: "Too Many Requests"},
			wantRetryable: true,
			wantClient:    true,
			wantServer:    false,
		},
		{
			name:          "client error 400",
			err:           &HTTPError{StatusCode: 400, Status: "Bad Request"},
			wantRetryable: false,
			wantClient:    true,
			wantServer:    false,
		},
		{
			name:          "not found 404",
			err:           &HTTPError{StatusCode: 404, Status: "Not Found"},
			wantRetryable: false,
			wantClient:    true,
			wantServer:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.wantRetryable, tt.err.IsRetryable())
			assert.Equal(t, tt.wantClient, tt.err.IsClientError())
			assert.Equal(t, tt.wantServer, tt.err.IsServerError())
		})
	}
}

func TestIsRetryableError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "nil error",
			err:  nil,
			want: false,
		},
		{
			name: "timeout sentinel",
			err:  ErrTimeout,
			want: true,
		},
		{
			name: "rate limited sentinel",
			err:  ErrRateLimited,
			want: true,
		},
		{
			name: "execution error retryable",
			err:  &ExecutionError{Retryable: true},
			want: true,
		},
		{
			name: "execution error not retryable",
			err:  &ExecutionError{Retryable: false},
			want: false,
		},
		{
			name: "http error 500",
			err:  &HTTPError{StatusCode: 500},
			want: true,
		},
		{
			name: "http error 400",
			err:  &HTTPError{StatusCode: 400},
			want: false,
		},
		{
			name: "wrapped timeout",
			err:  errors.New("connection timeout"),
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, IsRetryableError(tt.err))
		})
	}
}

func TestIsPermanentError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "nil error",
			err:  nil,
			want: false,
		},
		{
			name: "invalid config",
			err:  ErrInvalidConfig,
			want: true,
		},
		{
			name: "unauthorized",
			err:  ErrUnauthorized,
			want: true,
		},
		{
			name: "http 401",
			err:  &HTTPError{StatusCode: 401},
			want: true,
		},
		{
			name: "http 429 not permanent",
			err:  &HTTPError{StatusCode: 429},
			want: false,
		},
		{
			name: "timeout not permanent",
			err:  ErrTimeout,
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, IsPermanentError(tt.err))
		})
	}
}

func TestErrorConstructors(t *testing.T) {
	t.Run("NewExecutionError", func(t *testing.T) {
		err := NewExecutionError("test", "execute", errors.New("failed"), true)
		assert.Equal(t, "test", err.IntegrationName)
		assert.Equal(t, "execute", err.Operation)
		assert.True(t, err.Retryable)
	})

	t.Run("NewValidationError", func(t *testing.T) {
		err := NewValidationError("field", "message", "value")
		assert.Equal(t, "field", err.Field)
		assert.Equal(t, "message", err.Message)
		assert.Equal(t, "value", err.Value)
	})

	t.Run("NewHTTPError", func(t *testing.T) {
		err := NewHTTPError(500, "Internal Server Error", "body")
		assert.Equal(t, 500, err.StatusCode)
		assert.Equal(t, "Internal Server Error", err.Status)
		assert.Equal(t, "body", err.Body)
	})

	t.Run("NewAuthError", func(t *testing.T) {
		err := NewAuthError("bearer", "token expired", nil)
		assert.Equal(t, "bearer", err.AuthType)
		assert.Equal(t, "token expired", err.Message)
	})

	t.Run("NewPluginError", func(t *testing.T) {
		err := NewPluginError("plugin1", "load", "failed to load", nil)
		assert.Equal(t, "plugin1", err.PluginName)
		assert.Equal(t, "load", err.Operation)
	})
}
