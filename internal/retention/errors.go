package retention

import "errors"

// Sentinel errors for the retention package.
var (
	ErrNotFound           = errors.New("retention policy not found")
	ErrInvalidTenantID    = errors.New("invalid tenant ID")
	ErrDuplicatePolicy    = errors.New("retention policy with this name already exists")
	ErrPolicyInUse        = errors.New("retention policy is currently in use")
	ErrMinRetentionPeriod = errors.New("retention period is below minimum threshold")
	ErrCleanupInProgress  = errors.New("cleanup is already in progress")
	ErrExecutionNotFound  = errors.New("cleanup execution not found")
)

// ValidationError represents a validation error.
type ValidationError struct {
	Message string
	Field   string
}

func (e *ValidationError) Error() string {
	if e.Field != "" {
		return e.Field + ": " + e.Message
	}
	return e.Message
}

// ConflictError represents a policy conflict error.
type ConflictError struct {
	Message    string
	PolicyID   string
	ConflictID string
}

func (e *ConflictError) Error() string {
	return e.Message
}

// ReferentialIntegrityError represents a referential integrity violation.
type ReferentialIntegrityError struct {
	ResourceType string
	ResourceID   string
	References   []string
}

func (e *ReferentialIntegrityError) Error() string {
	return "cannot delete resource: has active references"
}
