package http

import (
	"errors"
	"fmt"
)

// Sentinel errors for HTTP operations.
var (
	// ErrCircuitOpen indicates the circuit breaker is open.
	ErrCircuitOpen = errors.New("circuit breaker is open")

	// ErrRateLimited indicates the request was rate limited.
	ErrRateLimited = errors.New("rate limited")

	// ErrTimeout indicates the request timed out.
	ErrTimeout = errors.New("request timed out")
)

// HTTPError represents an HTTP-specific error with status code.
type HTTPError struct {
	StatusCode int
	Status     string
	Body       string
	Err        error
}

// Error returns the error message.
func (e *HTTPError) Error() string {
	if e.Body != "" {
		return fmt.Sprintf("HTTP %d %s: %s", e.StatusCode, e.Status, e.Body)
	}
	return fmt.Sprintf("HTTP %d %s", e.StatusCode, e.Status)
}

// Unwrap returns the underlying error.
func (e *HTTPError) Unwrap() error {
	return e.Err
}

// IsRetryable returns whether the HTTP error is retryable.
func (e *HTTPError) IsRetryable() bool {
	// Retry on server errors and rate limiting
	return e.StatusCode >= 500 || e.StatusCode == 429 || e.StatusCode == 408
}

// IsClientError returns whether the HTTP error is a client error (4xx).
func (e *HTTPError) IsClientError() bool {
	return e.StatusCode >= 400 && e.StatusCode < 500
}

// IsServerError returns whether the HTTP error is a server error (5xx).
func (e *HTTPError) IsServerError() bool {
	return e.StatusCode >= 500
}

// NewHTTPError creates a new HTTPError.
func NewHTTPError(statusCode int, status, body string) *HTTPError {
	return &HTTPError{
		StatusCode: statusCode,
		Status:     status,
		Body:       body,
	}
}

// IsRetryableError checks if an error is retryable.
func IsRetryableError(err error) bool {
	if err == nil {
		return false
	}

	// Check for specific retryable sentinel errors
	if errors.Is(err, ErrTimeout) || errors.Is(err, ErrRateLimited) {
		return true
	}

	// Check for HTTPError
	var httpErr *HTTPError
	if errors.As(err, &httpErr) {
		return httpErr.IsRetryable()
	}

	return false
}
