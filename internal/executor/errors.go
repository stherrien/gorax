package executor

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"
	"syscall"
)

// ErrorClassification defines the type of error for retry handling
type ErrorClassification int

const (
	// ErrorClassificationUnknown is for errors that cannot be classified
	ErrorClassificationUnknown ErrorClassification = iota
	// ErrorClassificationTransient are temporary errors that can be retried
	ErrorClassificationTransient
	// ErrorClassificationPermanent are permanent errors that should not be retried
	ErrorClassificationPermanent
)

// ExecutionError wraps an error with classification and context
type ExecutionError struct {
	// Original error
	Err error
	// Classification of the error
	Classification ErrorClassification
	// NodeID where the error occurred
	NodeID string
	// NodeType where the error occurred
	NodeType string
	// Additional context
	Context map[string]interface{}
	// Retry count when this error occurred
	RetryCount int
}

// Error implements the error interface
func (e *ExecutionError) Error() string {
	return fmt.Sprintf("execution error in node %s (%s): %v", e.NodeID, e.NodeType, e.Err)
}

// Unwrap implements the errors.Unwrap interface
func (e *ExecutionError) Unwrap() error {
	return e.Err
}

// IsRetryable returns true if the error is transient and can be retried
func (e *ExecutionError) IsRetryable() bool {
	return e.Classification == ErrorClassificationTransient
}

// NewExecutionError creates a new ExecutionError with automatic classification
func NewExecutionError(err error, nodeID, nodeType string, retryCount int) *ExecutionError {
	return &ExecutionError{
		Err:            err,
		Classification: ClassifyError(err),
		NodeID:         nodeID,
		NodeType:       nodeType,
		Context:        make(map[string]interface{}),
		RetryCount:     retryCount,
	}
}

// WithContext adds context to the error
func (e *ExecutionError) WithContext(key string, value interface{}) *ExecutionError {
	e.Context[key] = value
	return e
}

// ClassifyError determines if an error is transient or permanent
func ClassifyError(err error) ErrorClassification {
	if err == nil {
		return ErrorClassificationUnknown
	}

	// Check for timeout errors
	if errors.Is(err, syscall.ETIMEDOUT) {
		return ErrorClassificationTransient
	}

	// Check for context errors
	if errors.Is(err, context.DeadlineExceeded) {
		return ErrorClassificationTransient
	}
	if errors.Is(err, context.Canceled) {
		return ErrorClassificationPermanent // User canceled, don't retry
	}

	// Check for network errors
	var netErr net.Error
	if errors.As(err, &netErr) {
		if netErr.Timeout() {
			return ErrorClassificationTransient
		}
		if netErr.Temporary() {
			return ErrorClassificationTransient
		}
	}

	// Check for DNS errors
	var dnsErr *net.DNSError
	if errors.As(err, &dnsErr) {
		if dnsErr.IsTimeout {
			return ErrorClassificationTransient
		}
		// DNS not found is permanent
		if dnsErr.IsNotFound {
			return ErrorClassificationPermanent
		}
		// Temporary DNS issues can be retried
		if dnsErr.Temporary() {
			return ErrorClassificationTransient
		}
	}

	// Check for connection errors
	var opErr *net.OpError
	if errors.As(err, &opErr) {
		// Connection refused might be temporary (service starting up)
		if errors.Is(opErr.Err, syscall.ECONNREFUSED) {
			return ErrorClassificationTransient
		}
		// Connection reset can be transient
		if errors.Is(opErr.Err, syscall.ECONNRESET) {
			return ErrorClassificationTransient
		}
		// Network unreachable is transient
		if errors.Is(opErr.Err, syscall.ENETUNREACH) {
			return ErrorClassificationTransient
		}
		// Host unreachable is transient
		if errors.Is(opErr.Err, syscall.EHOSTUNREACH) {
			return ErrorClassificationTransient
		}
	}

	// Check for syscall errors
	if errors.Is(err, syscall.ECONNREFUSED) ||
		errors.Is(err, syscall.ECONNRESET) ||
		errors.Is(err, syscall.ECONNABORTED) ||
		errors.Is(err, syscall.ENETUNREACH) ||
		errors.Is(err, syscall.EHOSTUNREACH) {
		return ErrorClassificationTransient
	}

	// Check error message for common patterns
	errMsg := strings.ToLower(err.Error())

	// Transient error patterns
	transientPatterns := []string{
		"timeout",
		"timed out",
		"temporary failure",
		"connection refused",
		"connection reset",
		"connection aborted",
		"network is unreachable",
		"host is unreachable",
		"too many connections",
		"service unavailable",
		"rate limit exceeded",
		"throttle",
		"try again",
		"temporarily unavailable",
		"gateway timeout",
		"bad gateway",
	}

	for _, pattern := range transientPatterns {
		if strings.Contains(errMsg, pattern) {
			return ErrorClassificationTransient
		}
	}

	// Permanent error patterns
	permanentPatterns := []string{
		"invalid",
		"malformed",
		"parse error",
		"syntax error",
		"unauthorized",
		"forbidden",
		"not found",
		"bad request",
		"authentication failed",
		"permission denied",
		"access denied",
		"unsupported",
		"not implemented",
		"method not allowed",
		"conflict",
		"precondition failed",
		"unprocessable entity",
		"payload too large",
		"uri too long",
		"expectation failed",
	}

	for _, pattern := range permanentPatterns {
		if strings.Contains(errMsg, pattern) {
			return ErrorClassificationPermanent
		}
	}

	// Default to unknown classification
	return ErrorClassificationUnknown
}

// ClassifyHTTPStatusCode classifies an HTTP status code
func ClassifyHTTPStatusCode(statusCode int) ErrorClassification {
	// 2xx and 3xx are successful, not errors
	if statusCode >= 200 && statusCode < 400 {
		return ErrorClassificationUnknown
	}

	// 4xx Client Errors - mostly permanent
	if statusCode >= 400 && statusCode < 500 {
		switch statusCode {
		case http.StatusRequestTimeout: // 408
			return ErrorClassificationTransient
		case http.StatusTooManyRequests: // 429
			return ErrorClassificationTransient
		case http.StatusConflict: // 409 - might be transient depending on use case
			return ErrorClassificationTransient
		default:
			return ErrorClassificationPermanent
		}
	}

	// 5xx Server Errors - mostly transient
	if statusCode >= 500 && statusCode < 600 {
		switch statusCode {
		case http.StatusNotImplemented: // 501
			return ErrorClassificationPermanent
		case http.StatusHTTPVersionNotSupported: // 505
			return ErrorClassificationPermanent
		default:
			// 500, 502, 503, 504 and others are transient
			return ErrorClassificationTransient
		}
	}

	return ErrorClassificationUnknown
}

// ShouldRetry determines if an operation should be retried based on the error and retry count
func ShouldRetry(err error, currentRetry, maxRetries int) bool {
	if err == nil {
		return false
	}

	if currentRetry >= maxRetries {
		return false
	}

	classification := ClassifyError(err)
	return classification == ErrorClassificationTransient
}

// WrapError wraps an error with execution context
func WrapError(err error, nodeID, nodeType string, retryCount int) error {
	if err == nil {
		return nil
	}

	// If already an ExecutionError, update retry count
	var execErr *ExecutionError
	if errors.As(err, &execErr) {
		execErr.RetryCount = retryCount
		return execErr
	}

	return NewExecutionError(err, nodeID, nodeType, retryCount)
}
