// Package suggestions provides smart error analysis and fix suggestions
// for workflow executions.
package suggestions

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// ErrorCategory categorizes the type of error for pattern matching
type ErrorCategory string

const (
	ErrorCategoryNetwork   ErrorCategory = "network"
	ErrorCategoryAuth      ErrorCategory = "auth"
	ErrorCategoryData      ErrorCategory = "data"
	ErrorCategoryRateLimit ErrorCategory = "rate_limit"
	ErrorCategoryTimeout   ErrorCategory = "timeout"
	ErrorCategoryConfig    ErrorCategory = "config"
	ErrorCategoryExternal  ErrorCategory = "external_service"
	ErrorCategoryUnknown   ErrorCategory = "unknown"
)

// SuggestionType indicates the type of fix suggested
type SuggestionType string

// #nosec G101 -- these are suggestion type constants, not credentials
const (
	SuggestionTypeRetry        SuggestionType = "retry"
	SuggestionTypeConfigChange SuggestionType = "config_change"
	SuggestionTypeCredential   SuggestionType = "credential_update"
	SuggestionTypeDataFix      SuggestionType = "data_fix"
	SuggestionTypeWorkflowFix  SuggestionType = "workflow_modification"
	SuggestionTypeManual       SuggestionType = "manual_intervention"
)

// SuggestionConfidence indicates how confident the system is in the suggestion
type SuggestionConfidence string

const (
	ConfidenceHigh   SuggestionConfidence = "high"
	ConfidenceMedium SuggestionConfidence = "medium"
	ConfidenceLow    SuggestionConfidence = "low"
)

// SuggestionStatus indicates the current state of a suggestion
type SuggestionStatus string

const (
	StatusPending   SuggestionStatus = "pending"
	StatusApplied   SuggestionStatus = "applied"
	StatusDismissed SuggestionStatus = "dismissed"
)

// SuggestionSource indicates where the suggestion came from
type SuggestionSource string

const (
	SourcePattern SuggestionSource = "pattern"
	SourceLLM     SuggestionSource = "llm"
)

// Suggestion represents a fix recommendation for an execution error
type Suggestion struct {
	ID          string               `json:"id" db:"id"`
	TenantID    string               `json:"tenant_id" db:"tenant_id"`
	ExecutionID string               `json:"execution_id" db:"execution_id"`
	NodeID      string               `json:"node_id" db:"node_id"`
	Category    ErrorCategory        `json:"category" db:"category"`
	Type        SuggestionType       `json:"type" db:"type"`
	Confidence  SuggestionConfidence `json:"confidence" db:"confidence"`
	Title       string               `json:"title" db:"title"`
	Description string               `json:"description" db:"description"`
	Details     string               `json:"details,omitempty" db:"details"`

	// Actionable fix information
	Fix *SuggestionFix `json:"fix,omitempty" db:"fix"`

	// Metadata
	Source      SuggestionSource `json:"source" db:"source"`
	Status      SuggestionStatus `json:"status" db:"status"`
	CreatedAt   time.Time        `json:"created_at" db:"created_at"`
	AppliedAt   *time.Time       `json:"applied_at,omitempty" db:"applied_at"`
	DismissedAt *time.Time       `json:"dismissed_at,omitempty" db:"dismissed_at"`
}

// SuggestionFix contains actionable fix data
type SuggestionFix struct {
	// For config changes
	ConfigPath string      `json:"config_path,omitempty"`
	OldValue   interface{} `json:"old_value,omitempty"`
	NewValue   interface{} `json:"new_value,omitempty"`

	// For retry suggestions
	RetryConfig *RetryConfig `json:"retry_config,omitempty"`

	// General action
	ActionType string      `json:"action_type"`
	ActionData interface{} `json:"action_data,omitempty"`
}

// RetryConfig for retry suggestions
type RetryConfig struct {
	MaxRetries    int     `json:"max_retries"`
	BackoffMs     int     `json:"backoff_ms"`
	BackoffFactor float64 `json:"backoff_factor"`
}

// ErrorContext contains information about the error for analysis
type ErrorContext struct {
	ExecutionID  string                 `json:"execution_id"`
	WorkflowID   string                 `json:"workflow_id"`
	NodeID       string                 `json:"node_id"`
	NodeType     string                 `json:"node_type"`
	ErrorMessage string                 `json:"error_message"`
	ErrorCode    string                 `json:"error_code,omitempty"`
	HTTPStatus   int                    `json:"http_status,omitempty"`
	RetryCount   int                    `json:"retry_count"`
	InputData    map[string]interface{} `json:"input_data,omitempty"`
	NodeConfig   map[string]interface{} `json:"node_config,omitempty"`
	Timestamp    time.Time              `json:"timestamp"`
}

// Analyzer interface for error analysis
type Analyzer interface {
	// Analyze examines an error context and returns suggestions
	Analyze(ctx context.Context, errCtx *ErrorContext) ([]*Suggestion, error)
	// CanHandle returns true if this analyzer can handle the given error context
	CanHandle(errCtx *ErrorContext) bool
	// Name returns the analyzer name for logging/debugging
	Name() string
}

// Service interface for suggestion operations
type Service interface {
	// AnalyzeError analyzes an execution error and generates suggestions
	AnalyzeError(ctx context.Context, tenantID string, errCtx *ErrorContext) ([]*Suggestion, error)
	// GetSuggestions retrieves suggestions for an execution
	GetSuggestions(ctx context.Context, tenantID, executionID string) ([]*Suggestion, error)
	// GetSuggestionByID retrieves a single suggestion
	GetSuggestionByID(ctx context.Context, tenantID, suggestionID string) (*Suggestion, error)
	// ApplySuggestion marks a suggestion as applied
	ApplySuggestion(ctx context.Context, tenantID, suggestionID string) error
	// DismissSuggestion marks a suggestion as dismissed
	DismissSuggestion(ctx context.Context, tenantID, suggestionID string) error
}

// Repository interface for suggestion persistence
type Repository interface {
	// Create creates a new suggestion
	Create(ctx context.Context, suggestion *Suggestion) error
	// CreateBatch creates multiple suggestions
	CreateBatch(ctx context.Context, suggestions []*Suggestion) error
	// GetByID retrieves a suggestion by ID
	GetByID(ctx context.Context, tenantID, id string) (*Suggestion, error)
	// GetByExecutionID retrieves suggestions for an execution
	GetByExecutionID(ctx context.Context, tenantID, executionID string) ([]*Suggestion, error)
	// UpdateStatus updates the status of a suggestion
	UpdateStatus(ctx context.Context, tenantID, id string, status SuggestionStatus) error
	// Delete deletes a suggestion
	Delete(ctx context.Context, tenantID, id string) error
	// DeleteByExecutionID deletes all suggestions for an execution
	DeleteByExecutionID(ctx context.Context, tenantID, executionID string) error
}

// NewErrorContext creates a new ErrorContext with the given parameters
func NewErrorContext(executionID, workflowID, nodeID, nodeType, errorMessage string) *ErrorContext {
	return &ErrorContext{
		ExecutionID:  executionID,
		WorkflowID:   workflowID,
		NodeID:       nodeID,
		NodeType:     nodeType,
		ErrorMessage: errorMessage,
		Timestamp:    time.Now(),
	}
}

// NewSuggestion creates a new Suggestion with a generated ID
func NewSuggestion(
	tenantID, executionID, nodeID string,
	category ErrorCategory,
	suggType SuggestionType,
	confidence SuggestionConfidence,
	title, description string,
	source SuggestionSource,
) *Suggestion {
	return &Suggestion{
		ID:          uuid.New().String(),
		TenantID:    tenantID,
		ExecutionID: executionID,
		NodeID:      nodeID,
		Category:    category,
		Type:        suggType,
		Confidence:  confidence,
		Title:       title,
		Description: description,
		Source:      source,
		Status:      StatusPending,
		CreatedAt:   time.Now(),
	}
}

// WithFix adds a fix to the suggestion and returns it for chaining
func (s *Suggestion) WithFix(fix *SuggestionFix) *Suggestion {
	s.Fix = fix
	return s
}

// WithDetails adds details to the suggestion and returns it for chaining
func (s *Suggestion) WithDetails(details string) *Suggestion {
	s.Details = details
	return s
}

// IsPending returns true if the suggestion is still pending
func (s *Suggestion) IsPending() bool {
	return s.Status == StatusPending
}

// IsApplied returns true if the suggestion has been applied
func (s *Suggestion) IsApplied() bool {
	return s.Status == StatusApplied
}

// IsDismissed returns true if the suggestion has been dismissed
func (s *Suggestion) IsDismissed() bool {
	return s.Status == StatusDismissed
}

// NewRetryFix creates a SuggestionFix for retry with backoff
func NewRetryFix(maxRetries, backoffMs int, backoffFactor float64) *SuggestionFix {
	return &SuggestionFix{
		ActionType: "retry_with_backoff",
		RetryConfig: &RetryConfig{
			MaxRetries:    maxRetries,
			BackoffMs:     backoffMs,
			BackoffFactor: backoffFactor,
		},
	}
}

// NewConfigChangeFix creates a SuggestionFix for config change
func NewConfigChangeFix(configPath string, oldValue, newValue interface{}) *SuggestionFix {
	return &SuggestionFix{
		ActionType: "config_change",
		ConfigPath: configPath,
		OldValue:   oldValue,
		NewValue:   newValue,
	}
}

// NewCredentialFix creates a SuggestionFix for credential update
func NewCredentialFix(credentialID string) *SuggestionFix {
	return &SuggestionFix{
		ActionType: "credential_update",
		ActionData: map[string]interface{}{
			"credential_id": credentialID,
		},
	}
}
