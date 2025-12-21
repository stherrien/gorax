package integrations

import (
	"context"
	"errors"
	"time"
)

// Common errors
var (
	ErrInvalidConfig     = errors.New("invalid configuration")
	ErrAuthFailed        = errors.New("authentication failed")
	ErrRateLimitExceeded = errors.New("rate limit exceeded")
	ErrNotFound          = errors.New("resource not found")
	ErrPermissionDenied  = errors.New("permission denied")
)

// Action represents an executable integration action
type Action interface {
	// Execute runs the action with the given configuration and input
	Execute(ctx context.Context, config map[string]interface{}, input map[string]interface{}) (map[string]interface{}, error)

	// Validate checks if the configuration is valid
	Validate(config map[string]interface{}) error

	// Name returns the unique name of the action (e.g., "slack:send_message")
	Name() string

	// Description returns a human-readable description of what the action does
	Description() string
}

// Client represents a base integration client
type Client interface {
	// Authenticate verifies credentials and establishes connection
	Authenticate(ctx context.Context) error

	// HealthCheck verifies the connection is still valid
	HealthCheck(ctx context.Context) error
}

// RetryConfig defines retry behavior for actions
type RetryConfig struct {
	MaxAttempts int
	BaseDelay   time.Duration
	MaxDelay    time.Duration
}

// DefaultRetryConfig provides sensible defaults for retries
var DefaultRetryConfig = RetryConfig{
	MaxAttempts: 3,
	BaseDelay:   time.Second,
	MaxDelay:    time.Second * 30,
}

// ActionMetadata contains metadata about an action
type ActionMetadata struct {
	Name        string
	Description string
	Category    string
	ConfigSpec  map[string]FieldSpec
	InputSpec   map[string]FieldSpec
	OutputSpec  map[string]FieldSpec
}

// FieldSpec defines a configuration or data field
type FieldSpec struct {
	Type        string // "string", "number", "boolean", "array", "object"
	Description string
	Required    bool
	Default     interface{}
	Options     []string // For enum fields
	Sensitive   bool     // If true, mask in logs
}

// ExecutionContext provides context for action execution
type ExecutionContext struct {
	ExecutionID string
	StepID      string
	UserID      string
	Credentials map[string]interface{}
}
