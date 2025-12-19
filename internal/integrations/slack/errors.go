package slack

import (
	"errors"
	"fmt"
)

// Common validation errors
var (
	ErrChannelRequired        = errors.New("channel is required")
	ErrUserRequired           = errors.New("user is required")
	ErrTextOrBlocksRequired   = errors.New("either text or blocks must be provided")
	ErrTextTooLong            = errors.New("text exceeds 40,000 character limit")
	ErrTimestampRequired      = errors.New("timestamp is required")
	ErrEmojiRequired          = errors.New("emoji is required")
	ErrInvalidCredential      = errors.New("invalid or missing Slack credential")
	ErrInvalidToken           = errors.New("invalid or expired access token")
	ErrRateLimitExceeded      = errors.New("Slack API rate limit exceeded")
	ErrChannelNotFound        = errors.New("channel not found")
	ErrMessageNotFound        = errors.New("message not found")
	ErrUserNotFound           = errors.New("user not found")
	ErrUnauthorized           = errors.New("unauthorized: missing required scopes")
	ErrChannelArchived        = errors.New("cannot post to archived channel")
	ErrRestrictedAction       = errors.New("action restricted in this channel")
	ErrAccountInactive        = errors.New("Slack account inactive")
	ErrInvalidAuth            = errors.New("invalid authentication")
	ErrMissingScope           = errors.New("missing required OAuth scope")
	ErrTokenRevoked           = errors.New("access token has been revoked")
)

// SlackError represents a Slack API error
type SlackError struct {
	ErrorCode string
	Message   string
	RetryAfter int // Seconds to wait before retrying (for rate limits)
}

func (e *SlackError) Error() string {
	if e.RetryAfter > 0 {
		return fmt.Sprintf("slack error %s: %s (retry after %ds)", e.ErrorCode, e.Message, e.RetryAfter)
	}
	return fmt.Sprintf("slack error %s: %s", e.ErrorCode, e.Message)
}

// IsRetryable returns true if the error is transient and should be retried
func (e *SlackError) IsRetryable() bool {
	retryableCodes := map[string]bool{
		"rate_limited":       true,
		"service_unavailable": true,
		"internal_error":     true,
		"timeout":            true,
		"fatal_error":        false, // Actually not retryable
	}
	return retryableCodes[e.ErrorCode]
}

// ParseSlackError converts a Slack error code to a structured error
func ParseSlackError(errorCode string) error {
	errorMap := map[string]error{
		"invalid_auth":          ErrInvalidAuth,
		"token_revoked":         ErrTokenRevoked,
		"token_expired":         ErrInvalidToken,
		"account_inactive":      ErrAccountInactive,
		"channel_not_found":     ErrChannelNotFound,
		"not_in_channel":        ErrChannelNotFound,
		"is_archived":           ErrChannelArchived,
		"msg_too_long":          ErrTextTooLong,
		"no_text":               ErrTextOrBlocksRequired,
		"rate_limited":          ErrRateLimitExceeded,
		"message_not_found":     ErrMessageNotFound,
		"cant_update_message":   ErrUnauthorized,
		"edit_window_closed":    ErrRestrictedAction,
		"user_not_found":        ErrUserNotFound,
		"users_not_found":       ErrUserNotFound,
		"missing_scope":         ErrMissingScope,
		"restricted_action":     ErrRestrictedAction,
		"not_authed":            ErrInvalidAuth,
		"invalid_arguments":     errors.New("invalid arguments provided"),
		"invalid_name":          errors.New("invalid arguments provided"),
		"internal_error":        errors.New("Slack internal error"),
		"fatal_error":           errors.New("Slack fatal error"),
		"service_unavailable":   errors.New("Slack service temporarily unavailable"),
	}

	if err, ok := errorMap[errorCode]; ok {
		return &SlackError{
			ErrorCode: errorCode,
			Message:   err.Error(),
		}
	}

	return &SlackError{
		ErrorCode: errorCode,
		Message:   fmt.Sprintf("unknown Slack error: %s", errorCode),
	}
}

// ValidationError represents a configuration validation error
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation error on field '%s': %s", e.Field, e.Message)
}

// RateLimitError represents a rate limit error with retry information
type RateLimitError struct {
	RetryAfter int    // Seconds to wait
	Message    string
}

func (e *RateLimitError) Error() string {
	return fmt.Sprintf("rate limit exceeded: %s (retry after %ds)", e.Message, e.RetryAfter)
}
