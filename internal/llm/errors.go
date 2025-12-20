package llm

import (
	"errors"
	"fmt"
)

// Common LLM errors
var (
	// Authentication errors
	ErrInvalidAPIKey = errors.New("invalid or missing API key")
	ErrUnauthorized  = errors.New("unauthorized: check API key and permissions")

	// Request validation errors
	ErrInvalidModel       = errors.New("invalid or unsupported model")
	ErrEmptyMessages      = errors.New("messages cannot be empty")
	ErrEmptyTexts         = errors.New("texts cannot be empty")
	ErrInvalidRole        = errors.New("invalid message role: must be system, user, or assistant")
	ErrInvalidTemperature = errors.New("temperature must be between 0 and 2")
	ErrInvalidTopP        = errors.New("top_p must be between 0 and 1")

	// Limit errors
	ErrRateLimitExceeded  = errors.New("rate limit exceeded")
	ErrContextTooLong     = errors.New("input exceeds model context window")
	ErrTokenLimitExceeded = errors.New("token limit exceeded")
	ErrQuotaExceeded      = errors.New("API quota exceeded")

	// Provider errors
	ErrProviderUnavailable = errors.New("provider temporarily unavailable")
	ErrProviderTimeout     = errors.New("provider request timed out")
	ErrProviderNotFound    = errors.New("provider not found in registry")

	// Content errors
	ErrContentFiltered = errors.New("content filtered by provider safety measures")
	ErrInvalidResponse = errors.New("invalid response from provider")

	// Capability errors
	ErrUnsupportedOperation = errors.New("operation not supported by this provider")
)

// LLMError wraps provider errors with additional context
type LLMError struct {
	// Provider is the LLM provider name (e.g., "openai", "anthropic")
	Provider string

	// Code is the provider-specific error code
	Code string

	// Message is a human-readable error description
	Message string

	// HTTPStatus is the HTTP status code if applicable
	HTTPStatus int

	// RetryAfter is seconds to wait before retrying (for rate limits)
	RetryAfter int

	// Retryable indicates if the error is transient and the request can be retried
	Retryable bool

	// Cause is the underlying error
	Cause error
}

// Error implements the error interface
func (e *LLMError) Error() string {
	if e.RetryAfter > 0 {
		return fmt.Sprintf("%s error [%s]: %s (retry after %ds)", e.Provider, e.Code, e.Message, e.RetryAfter)
	}
	if e.Code != "" {
		return fmt.Sprintf("%s error [%s]: %s", e.Provider, e.Code, e.Message)
	}
	return fmt.Sprintf("%s error: %s", e.Provider, e.Message)
}

// Unwrap returns the underlying cause for errors.Is/errors.As
func (e *LLMError) Unwrap() error {
	return e.Cause
}

// IsRetryable returns true if the error is transient and can be retried
func (e *LLMError) IsRetryable() bool {
	return e.Retryable
}

// NewLLMError creates a new LLMError
func NewLLMError(provider, code, message string, cause error) *LLMError {
	return &LLMError{
		Provider: provider,
		Code:     code,
		Message:  message,
		Cause:    cause,
	}
}

// NewRetryableLLMError creates a new retryable LLMError
func NewRetryableLLMError(provider, code, message string, retryAfter int) *LLMError {
	return &LLMError{
		Provider:   provider,
		Code:       code,
		Message:    message,
		RetryAfter: retryAfter,
		Retryable:  true,
		Cause:      ErrRateLimitExceeded,
	}
}

// IsRateLimitError checks if an error is a rate limit error
func IsRateLimitError(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, ErrRateLimitExceeded) {
		return true
	}
	var llmErr *LLMError
	if errors.As(err, &llmErr) {
		return llmErr.Code == "rate_limit_exceeded" ||
			llmErr.Code == "429" ||
			llmErr.HTTPStatus == 429
	}
	return false
}

// IsAuthError checks if an error is an authentication error
func IsAuthError(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, ErrInvalidAPIKey) || errors.Is(err, ErrUnauthorized) {
		return true
	}
	var llmErr *LLMError
	if errors.As(err, &llmErr) {
		return llmErr.Code == "invalid_api_key" ||
			llmErr.Code == "unauthorized" ||
			llmErr.HTTPStatus == 401 ||
			llmErr.HTTPStatus == 403
	}
	return false
}

// IsContextLengthError checks if an error is due to context length
func IsContextLengthError(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, ErrContextTooLong) || errors.Is(err, ErrTokenLimitExceeded) {
		return true
	}
	var llmErr *LLMError
	if errors.As(err, &llmErr) {
		return llmErr.Code == "context_length_exceeded" ||
			llmErr.Code == "max_tokens_exceeded"
	}
	return false
}

// IsRetryableError checks if an error should be retried
func IsRetryableError(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, ErrRateLimitExceeded) || errors.Is(err, ErrProviderUnavailable) || errors.Is(err, ErrProviderTimeout) {
		return true
	}
	var llmErr *LLMError
	if errors.As(err, &llmErr) {
		return llmErr.Retryable
	}
	return false
}

// GetRetryAfter returns the retry-after duration in seconds, or 0 if not set
func GetRetryAfter(err error) int {
	var llmErr *LLMError
	if errors.As(err, &llmErr) {
		return llmErr.RetryAfter
	}
	return 0
}
