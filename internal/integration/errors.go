// Package integration provides a comprehensive framework for secure,
// extensible connections to external services in the Gorax platform.
package integration

import (
	"errors"
	"fmt"
)

// Sentinel errors for common integration failures.
var (
	// ErrNotFound indicates the requested integration was not found.
	ErrNotFound = errors.New("integration not found")

	// ErrAlreadyRegistered indicates the integration is already registered.
	ErrAlreadyRegistered = errors.New("integration already registered")

	// ErrInvalidConfig indicates the integration configuration is invalid.
	ErrInvalidConfig = errors.New("invalid integration configuration")

	// ErrInvalidCredentials indicates the credentials are invalid or missing.
	ErrInvalidCredentials = errors.New("invalid credentials")

	// ErrExecutionFailed indicates the integration execution failed.
	ErrExecutionFailed = errors.New("integration execution failed")

	// ErrTimeout indicates the integration operation timed out.
	ErrTimeout = errors.New("operation timed out")

	// ErrCircuitOpen indicates the circuit breaker is open.
	ErrCircuitOpen = errors.New("circuit breaker is open")

	// ErrRateLimited indicates the request was rate limited.
	ErrRateLimited = errors.New("rate limited")

	// ErrUnauthorized indicates authentication failed.
	ErrUnauthorized = errors.New("unauthorized")

	// ErrForbidden indicates authorization failed.
	ErrForbidden = errors.New("forbidden")

	// ErrPluginInvalid indicates the plugin failed validation.
	ErrPluginInvalid = errors.New("plugin validation failed")

	// ErrPluginNotLoaded indicates the plugin is not loaded.
	ErrPluginNotLoaded = errors.New("plugin not loaded")

	// ErrSchemaValidation indicates schema validation failed.
	ErrSchemaValidation = errors.New("schema validation failed")
)

// ExecutionError wraps errors that occur during integration execution.
type ExecutionError struct {
	IntegrationName string
	Operation       string
	Err             error
	Retryable       bool
}

// Error returns the error message.
func (e *ExecutionError) Error() string {
	return fmt.Sprintf("integration %s: %s failed: %v", e.IntegrationName, e.Operation, e.Err)
}

// Unwrap returns the underlying error.
func (e *ExecutionError) Unwrap() error {
	return e.Err
}

// IsRetryable returns whether the error is retryable.
func (e *ExecutionError) IsRetryable() bool {
	return e.Retryable
}

// ValidationError represents configuration or input validation errors.
type ValidationError struct {
	Field   string
	Message string
	Value   any
}

// Error returns the error message.
func (e *ValidationError) Error() string {
	if e.Field != "" {
		return fmt.Sprintf("validation error for field '%s': %s", e.Field, e.Message)
	}
	return fmt.Sprintf("validation error: %s", e.Message)
}

// ConfigError represents configuration-related errors.
type ConfigError struct {
	ConfigKey string
	Message   string
	Err       error
}

// Error returns the error message.
func (e *ConfigError) Error() string {
	if e.ConfigKey != "" {
		return fmt.Sprintf("config error for '%s': %s", e.ConfigKey, e.Message)
	}
	return fmt.Sprintf("config error: %s", e.Message)
}

// Unwrap returns the underlying error.
func (e *ConfigError) Unwrap() error {
	return e.Err
}

// HTTPError represents HTTP-specific errors with status code.
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

// AuthError represents authentication-related errors.
type AuthError struct {
	AuthType string
	Message  string
	Err      error
}

// Error returns the error message.
func (e *AuthError) Error() string {
	return fmt.Sprintf("auth error (%s): %s", e.AuthType, e.Message)
}

// Unwrap returns the underlying error.
func (e *AuthError) Unwrap() error {
	return e.Err
}

// PluginError represents plugin-related errors.
type PluginError struct {
	PluginName string
	Operation  string
	Message    string
	Err        error
}

// Error returns the error message.
func (e *PluginError) Error() string {
	return fmt.Sprintf("plugin %s: %s: %s", e.PluginName, e.Operation, e.Message)
}

// Unwrap returns the underlying error.
func (e *PluginError) Unwrap() error {
	return e.Err
}

// NewExecutionError creates a new ExecutionError.
func NewExecutionError(name, operation string, err error, retryable bool) *ExecutionError {
	return &ExecutionError{
		IntegrationName: name,
		Operation:       operation,
		Err:             err,
		Retryable:       retryable,
	}
}

// NewValidationError creates a new ValidationError.
func NewValidationError(field, message string, value any) *ValidationError {
	return &ValidationError{
		Field:   field,
		Message: message,
		Value:   value,
	}
}

// NewConfigError creates a new ConfigError.
func NewConfigError(key, message string, err error) *ConfigError {
	return &ConfigError{
		ConfigKey: key,
		Message:   message,
		Err:       err,
	}
}

// NewHTTPError creates a new HTTPError.
func NewHTTPError(statusCode int, status, body string) *HTTPError {
	return &HTTPError{
		StatusCode: statusCode,
		Status:     status,
		Body:       body,
	}
}

// NewAuthError creates a new AuthError.
func NewAuthError(authType, message string, err error) *AuthError {
	return &AuthError{
		AuthType: authType,
		Message:  message,
		Err:      err,
	}
}

// NewPluginError creates a new PluginError.
func NewPluginError(name, operation, message string, err error) *PluginError {
	return &PluginError{
		PluginName: name,
		Operation:  operation,
		Message:    message,
		Err:        err,
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

	// Check for ExecutionError
	var execErr *ExecutionError
	if errors.As(err, &execErr) {
		return execErr.IsRetryable()
	}

	// Check for HTTPError
	var httpErr *HTTPError
	if errors.As(err, &httpErr) {
		return httpErr.IsRetryable()
	}

	return false
}

// IsTransientError checks if an error is transient (temporary).
func IsTransientError(err error) bool {
	return IsRetryableError(err)
}

// IsPermanentError checks if an error is permanent (non-recoverable).
func IsPermanentError(err error) bool {
	if err == nil {
		return false
	}

	// Check for permanent sentinel errors
	if errors.Is(err, ErrInvalidConfig) ||
		errors.Is(err, ErrInvalidCredentials) ||
		errors.Is(err, ErrUnauthorized) ||
		errors.Is(err, ErrForbidden) ||
		errors.Is(err, ErrSchemaValidation) {
		return true
	}

	// Check for HTTPError client errors (except rate limiting)
	var httpErr *HTTPError
	if errors.As(err, &httpErr) {
		return httpErr.IsClientError() && httpErr.StatusCode != 429
	}

	return false
}
