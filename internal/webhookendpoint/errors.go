package webhookendpoint

import "errors"

// Errors for webhook endpoint operations
var (
	ErrEndpointNotFound   = errors.New("webhook endpoint not found")
	ErrEndpointInactive   = errors.New("webhook endpoint is not active")
	ErrEndpointExpired    = errors.New("webhook endpoint has expired")
	ErrInvalidToken       = errors.New("invalid endpoint token")
	ErrInvalidPayload     = errors.New("invalid webhook payload")
	ErrPayloadValidation  = errors.New("payload validation failed")
	ErrMissingRequiredField = errors.New("missing required field in payload")
	ErrAuthenticationFailed = errors.New("webhook authentication failed")
)
