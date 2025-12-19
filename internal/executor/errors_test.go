package executor

import (
	"context"
	"errors"
	"net"
	"net/http"
	"syscall"
	"testing"
)

func TestClassifyError(t *testing.T) {
	tests := []struct {
		name           string
		err            error
		expectedClass  ErrorClassification
	}{
		{
			name:          "nil error",
			err:           nil,
			expectedClass: ErrorClassificationUnknown,
		},
		{
			name:          "context deadline exceeded",
			err:           context.DeadlineExceeded,
			expectedClass: ErrorClassificationTransient,
		},
		{
			name:          "context canceled",
			err:           context.Canceled,
			expectedClass: ErrorClassificationPermanent,
		},
		{
			name:          "connection refused",
			err:           syscall.ECONNREFUSED,
			expectedClass: ErrorClassificationTransient,
		},
		{
			name:          "connection reset",
			err:           syscall.ECONNRESET,
			expectedClass: ErrorClassificationTransient,
		},
		{
			name:          "network unreachable",
			err:           syscall.ENETUNREACH,
			expectedClass: ErrorClassificationTransient,
		},
		{
			name:          "timeout error message",
			err:           errors.New("request timed out"),
			expectedClass: ErrorClassificationTransient,
		},
		{
			name:          "temporary failure message",
			err:           errors.New("temporary failure in name resolution"),
			expectedClass: ErrorClassificationTransient,
		},
		{
			name:          "rate limit error",
			err:           errors.New("rate limit exceeded"),
			expectedClass: ErrorClassificationTransient,
		},
		{
			name:          "service unavailable",
			err:           errors.New("service unavailable"),
			expectedClass: ErrorClassificationTransient,
		},
		{
			name:          "invalid request",
			err:           errors.New("invalid request format"),
			expectedClass: ErrorClassificationPermanent,
		},
		{
			name:          "unauthorized error",
			err:           errors.New("unauthorized access"),
			expectedClass: ErrorClassificationPermanent,
		},
		{
			name:          "not found error",
			err:           errors.New("resource not found"),
			expectedClass: ErrorClassificationPermanent,
		},
		{
			name:          "bad request error",
			err:           errors.New("bad request"),
			expectedClass: ErrorClassificationPermanent,
		},
		{
			name:          "unknown error",
			err:           errors.New("something went wrong"),
			expectedClass: ErrorClassificationUnknown,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			classification := ClassifyError(tt.err)
			if classification != tt.expectedClass {
				t.Errorf("ClassifyError() = %v, want %v", classification, tt.expectedClass)
			}
		})
	}
}

func TestClassifyHTTPStatusCode(t *testing.T) {
	tests := []struct {
		name           string
		statusCode     int
		expectedClass  ErrorClassification
	}{
		{
			name:          "200 OK",
			statusCode:    http.StatusOK,
			expectedClass: ErrorClassificationUnknown,
		},
		{
			name:          "201 Created",
			statusCode:    http.StatusCreated,
			expectedClass: ErrorClassificationUnknown,
		},
		{
			name:          "400 Bad Request",
			statusCode:    http.StatusBadRequest,
			expectedClass: ErrorClassificationPermanent,
		},
		{
			name:          "401 Unauthorized",
			statusCode:    http.StatusUnauthorized,
			expectedClass: ErrorClassificationPermanent,
		},
		{
			name:          "403 Forbidden",
			statusCode:    http.StatusForbidden,
			expectedClass: ErrorClassificationPermanent,
		},
		{
			name:          "404 Not Found",
			statusCode:    http.StatusNotFound,
			expectedClass: ErrorClassificationPermanent,
		},
		{
			name:          "408 Request Timeout",
			statusCode:    http.StatusRequestTimeout,
			expectedClass: ErrorClassificationTransient,
		},
		{
			name:          "409 Conflict",
			statusCode:    http.StatusConflict,
			expectedClass: ErrorClassificationTransient,
		},
		{
			name:          "429 Too Many Requests",
			statusCode:    http.StatusTooManyRequests,
			expectedClass: ErrorClassificationTransient,
		},
		{
			name:          "500 Internal Server Error",
			statusCode:    http.StatusInternalServerError,
			expectedClass: ErrorClassificationTransient,
		},
		{
			name:          "501 Not Implemented",
			statusCode:    http.StatusNotImplemented,
			expectedClass: ErrorClassificationPermanent,
		},
		{
			name:          "502 Bad Gateway",
			statusCode:    http.StatusBadGateway,
			expectedClass: ErrorClassificationTransient,
		},
		{
			name:          "503 Service Unavailable",
			statusCode:    http.StatusServiceUnavailable,
			expectedClass: ErrorClassificationTransient,
		},
		{
			name:          "504 Gateway Timeout",
			statusCode:    http.StatusGatewayTimeout,
			expectedClass: ErrorClassificationTransient,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			classification := ClassifyHTTPStatusCode(tt.statusCode)
			if classification != tt.expectedClass {
				t.Errorf("ClassifyHTTPStatusCode() = %v, want %v", classification, tt.expectedClass)
			}
		})
	}
}

func TestShouldRetry(t *testing.T) {
	tests := []struct {
		name         string
		err          error
		currentRetry int
		maxRetries   int
		shouldRetry  bool
	}{
		{
			name:         "nil error should not retry",
			err:          nil,
			currentRetry: 0,
			maxRetries:   3,
			shouldRetry:  false,
		},
		{
			name:         "max retries exceeded should not retry",
			err:          errors.New("timeout"),
			currentRetry: 3,
			maxRetries:   3,
			shouldRetry:  false,
		},
		{
			name:         "transient error should retry",
			err:          errors.New("connection refused"),
			currentRetry: 0,
			maxRetries:   3,
			shouldRetry:  true,
		},
		{
			name:         "permanent error should not retry",
			err:          errors.New("invalid request"),
			currentRetry: 0,
			maxRetries:   3,
			shouldRetry:  false,
		},
		{
			name:         "unknown error should not retry",
			err:          errors.New("something went wrong"),
			currentRetry: 0,
			maxRetries:   3,
			shouldRetry:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			shouldRetry := ShouldRetry(tt.err, tt.currentRetry, tt.maxRetries)
			if shouldRetry != tt.shouldRetry {
				t.Errorf("ShouldRetry() = %v, want %v", shouldRetry, tt.shouldRetry)
			}
		})
	}
}

func TestExecutionError(t *testing.T) {
	err := errors.New("connection timeout")
	nodeID := "node-123"
	nodeType := "action:http"
	retryCount := 2

	execErr := NewExecutionError(err, nodeID, nodeType, retryCount)

	if execErr.Err != err {
		t.Errorf("ExecutionError.Err = %v, want %v", execErr.Err, err)
	}
	if execErr.NodeID != nodeID {
		t.Errorf("ExecutionError.NodeID = %v, want %v", execErr.NodeID, nodeID)
	}
	if execErr.NodeType != nodeType {
		t.Errorf("ExecutionError.NodeType = %v, want %v", execErr.NodeType, nodeType)
	}
	if execErr.RetryCount != retryCount {
		t.Errorf("ExecutionError.RetryCount = %v, want %v", execErr.RetryCount, retryCount)
	}
	if execErr.Classification != ErrorClassificationTransient {
		t.Errorf("ExecutionError.Classification = %v, want %v", execErr.Classification, ErrorClassificationTransient)
	}
}

func TestExecutionErrorIsRetryable(t *testing.T) {
	tests := []struct {
		name          string
		classification ErrorClassification
		isRetryable   bool
	}{
		{
			name:          "transient error is retryable",
			classification: ErrorClassificationTransient,
			isRetryable:   true,
		},
		{
			name:          "permanent error is not retryable",
			classification: ErrorClassificationPermanent,
			isRetryable:   false,
		},
		{
			name:          "unknown error is not retryable",
			classification: ErrorClassificationUnknown,
			isRetryable:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			execErr := &ExecutionError{
				Err:            errors.New("test error"),
				Classification: tt.classification,
				NodeID:         "node-123",
				NodeType:       "action:http",
			}

			if execErr.IsRetryable() != tt.isRetryable {
				t.Errorf("ExecutionError.IsRetryable() = %v, want %v", execErr.IsRetryable(), tt.isRetryable)
			}
		})
	}
}

func TestWrapError(t *testing.T) {
	err := errors.New("test error")
	nodeID := "node-123"
	nodeType := "action:http"
	retryCount := 1

	wrappedErr := WrapError(err, nodeID, nodeType, retryCount)

	execErr, ok := wrappedErr.(*ExecutionError)
	if !ok {
		t.Fatalf("WrapError() did not return *ExecutionError")
	}

	if execErr.NodeID != nodeID {
		t.Errorf("ExecutionError.NodeID = %v, want %v", execErr.NodeID, nodeID)
	}
	if execErr.RetryCount != retryCount {
		t.Errorf("ExecutionError.RetryCount = %v, want %v", execErr.RetryCount, retryCount)
	}
}

func TestExecutionErrorWithContext(t *testing.T) {
	execErr := NewExecutionError(errors.New("test error"), "node-123", "action:http", 0)

	execErr.WithContext("key1", "value1")
	execErr.WithContext("key2", 123)

	if execErr.Context["key1"] != "value1" {
		t.Errorf("Context[key1] = %v, want 'value1'", execErr.Context["key1"])
	}
	if execErr.Context["key2"] != 123 {
		t.Errorf("Context[key2] = %v, want 123", execErr.Context["key2"])
	}
}

func TestClassifyNetworkErrors(t *testing.T) {
	// Create a timeout error
	timeoutErr := &net.OpError{
		Op:  "dial",
		Net: "tcp",
		Err: &timeoutError{},
	}

	classification := ClassifyError(timeoutErr)
	if classification != ErrorClassificationTransient {
		t.Errorf("timeout error should be transient, got %v", classification)
	}

	// Create a DNS error
	dnsErr := &net.DNSError{
		Err:       "no such host",
		IsTimeout: true,
	}

	classification = ClassifyError(dnsErr)
	if classification != ErrorClassificationTransient {
		t.Errorf("DNS timeout error should be transient, got %v", classification)
	}
}

// Mock timeout error for testing
type timeoutError struct{}

func (e *timeoutError) Error() string   { return "timeout" }
func (e *timeoutError) Timeout() bool   { return true }
func (e *timeoutError) Temporary() bool { return true }
