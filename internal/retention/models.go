// Package retention provides automated data lifecycle management for the Gorax platform.
package retention

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// ResourceType defines the types of resources that can have retention policies.
type ResourceType string

const (
	ResourceTypeWorkflows      ResourceType = "workflows"
	ResourceTypeExecutions     ResourceType = "executions"
	ResourceTypeLogs           ResourceType = "logs"
	ResourceTypeAuditLogs      ResourceType = "audit_logs"
	ResourceTypeWebhookEvents  ResourceType = "webhook_events"
	ResourceTypeScheduleEvents ResourceType = "schedule_events"
)

// AllResourceTypes returns all valid resource types.
func AllResourceTypes() []ResourceType {
	return []ResourceType{
		ResourceTypeWorkflows,
		ResourceTypeExecutions,
		ResourceTypeLogs,
		ResourceTypeAuditLogs,
		ResourceTypeWebhookEvents,
		ResourceTypeScheduleEvents,
	}
}

// IsValidResourceType checks if a resource type is valid.
func IsValidResourceType(rt ResourceType) bool {
	switch rt {
	case ResourceTypeWorkflows, ResourceTypeExecutions, ResourceTypeLogs,
		ResourceTypeAuditLogs, ResourceTypeWebhookEvents, ResourceTypeScheduleEvents:
		return true
	}
	return false
}

// PolicyStatus defines the status of a retention policy.
type PolicyStatus string

const (
	PolicyStatusActive    PolicyStatus = "active"
	PolicyStatusInactive  PolicyStatus = "inactive"
	PolicyStatusSuspended PolicyStatus = "suspended"
)

// IsValidPolicyStatus checks if a policy status is valid.
func IsValidPolicyStatus(status PolicyStatus) bool {
	switch status {
	case PolicyStatusActive, PolicyStatusInactive, PolicyStatusSuspended:
		return true
	}
	return false
}

// MinRetentionDays is the minimum allowed retention period for safety.
const MinRetentionDays = 7

// MaxRetentionDays is the maximum allowed retention period.
const MaxRetentionDays = 3650 // 10 years

// AdvancedRetentionPolicy defines a comprehensive data retention policy for a tenant.
// This is used for resource-type-specific policies with rules-based filtering.
type AdvancedRetentionPolicy struct {
	ID            string          `db:"id" json:"id"`
	TenantID      string          `db:"tenant_id" json:"tenant_id"`
	Name          string          `db:"name" json:"name"`
	Description   string          `db:"description" json:"description"`
	ResourceType  ResourceType    `db:"resource_type" json:"resource_type"`
	RetentionDays int             `db:"retention_days" json:"retention_days"`
	Status        PolicyStatus    `db:"status" json:"status"`
	Priority      int             `db:"priority" json:"priority"`
	Rules         json.RawMessage `db:"rules" json:"rules,omitempty"`
	CreatedBy     string          `db:"created_by" json:"created_by"`
	CreatedAt     time.Time       `db:"created_at" json:"created_at"`
	UpdatedAt     time.Time       `db:"updated_at" json:"updated_at"`
}

// IsActive returns true if the policy is active.
func (p *AdvancedRetentionPolicy) IsActive() bool {
	return p.Status == PolicyStatusActive
}

// GetRules parses and returns the retention rules.
func (p *AdvancedRetentionPolicy) GetRules() (*RetentionRules, error) {
	if len(p.Rules) == 0 {
		return &RetentionRules{}, nil
	}
	var rules RetentionRules
	if err := json.Unmarshal(p.Rules, &rules); err != nil {
		return nil, fmt.Errorf("parse retention rules: %w", err)
	}
	return &rules, nil
}

// SetRules serializes and stores retention rules.
func (p *AdvancedRetentionPolicy) SetRules(rules *RetentionRules) error {
	if rules == nil {
		p.Rules = nil
		return nil
	}
	data, err := json.Marshal(rules)
	if err != nil {
		return fmt.Errorf("marshal retention rules: %w", err)
	}
	p.Rules = data
	return nil
}

// Validate checks if the advanced retention policy is valid.
func (p *AdvancedRetentionPolicy) Validate() error {
	if p.TenantID == "" {
		return ErrInvalidTenantID
	}
	if p.Name == "" {
		return &ValidationError{Message: "name is required"}
	}
	if len(p.Name) > 255 {
		return &ValidationError{Message: "name must be at most 255 characters"}
	}
	if !IsValidResourceType(p.ResourceType) {
		return &ValidationError{Message: fmt.Sprintf("invalid resource type: %s", p.ResourceType)}
	}
	if p.RetentionDays < MinRetentionDays {
		return &ValidationError{Message: fmt.Sprintf("retention days must be at least %d", MinRetentionDays)}
	}
	if p.RetentionDays > MaxRetentionDays {
		return &ValidationError{Message: fmt.Sprintf("retention days must be at most %d", MaxRetentionDays)}
	}
	if !IsValidPolicyStatus(p.Status) {
		return &ValidationError{Message: fmt.Sprintf("invalid status: %s", p.Status)}
	}
	return nil
}

// RetentionRules defines optional filtering rules for a retention policy.
type RetentionRules struct {
	WorkflowID   string            `json:"workflow_id,omitempty"`
	WorkflowIDs  []string          `json:"workflow_ids,omitempty"`
	Status       string            `json:"status,omitempty"`
	Statuses     []string          `json:"statuses,omitempty"`
	Tags         map[string]string `json:"tags,omitempty"`
	ExcludeTags  map[string]string `json:"exclude_tags,omitempty"`
	MinPriority  *int              `json:"min_priority,omitempty"`
	MaxPriority  *int              `json:"max_priority,omitempty"`
	CreatedAfter *time.Time        `json:"created_after,omitempty"`
}

// IsEmpty returns true if no rules are defined.
func (r *RetentionRules) IsEmpty() bool {
	return r.WorkflowID == "" &&
		len(r.WorkflowIDs) == 0 &&
		r.Status == "" &&
		len(r.Statuses) == 0 &&
		len(r.Tags) == 0 &&
		len(r.ExcludeTags) == 0 &&
		r.MinPriority == nil &&
		r.MaxPriority == nil &&
		r.CreatedAfter == nil
}

// CreatePolicyInput is the input for creating a retention policy.
type CreatePolicyInput struct {
	Name          string          `json:"name" validate:"required,min=1,max=255"`
	Description   string          `json:"description" validate:"max=1000"`
	ResourceType  ResourceType    `json:"resource_type" validate:"required"`
	RetentionDays int             `json:"retention_days" validate:"required,min=7,max=3650"`
	Priority      int             `json:"priority" validate:"min=0,max=100"`
	Rules         *RetentionRules `json:"rules,omitempty"`
}

// Validate validates the create policy input.
func (i *CreatePolicyInput) Validate() error {
	if i.Name == "" {
		return &ValidationError{Message: "name is required"}
	}
	if len(i.Name) > 255 {
		return &ValidationError{Message: "name must be at most 255 characters"}
	}
	if !IsValidResourceType(i.ResourceType) {
		return &ValidationError{Message: fmt.Sprintf("invalid resource type: %s", i.ResourceType)}
	}
	if i.RetentionDays < MinRetentionDays {
		return &ValidationError{Message: fmt.Sprintf("retention days must be at least %d", MinRetentionDays)}
	}
	if i.RetentionDays > MaxRetentionDays {
		return &ValidationError{Message: fmt.Sprintf("retention days must be at most %d", MaxRetentionDays)}
	}
	if i.Priority < 0 || i.Priority > 100 {
		return &ValidationError{Message: "priority must be between 0 and 100"}
	}
	return nil
}

// UpdatePolicyInput is the input for updating a retention policy.
type UpdatePolicyInput struct {
	Name          *string         `json:"name,omitempty"`
	Description   *string         `json:"description,omitempty"`
	RetentionDays *int            `json:"retention_days,omitempty"`
	Status        *PolicyStatus   `json:"status,omitempty"`
	Priority      *int            `json:"priority,omitempty"`
	Rules         *RetentionRules `json:"rules,omitempty"`
}

// Validate validates the update policy input.
func (i *UpdatePolicyInput) Validate() error {
	if i.Name != nil {
		if *i.Name == "" {
			return &ValidationError{Message: "name cannot be empty"}
		}
		if len(*i.Name) > 255 {
			return &ValidationError{Message: "name must be at most 255 characters"}
		}
	}
	if i.RetentionDays != nil {
		if *i.RetentionDays < MinRetentionDays {
			return &ValidationError{Message: fmt.Sprintf("retention days must be at least %d", MinRetentionDays)}
		}
		if *i.RetentionDays > MaxRetentionDays {
			return &ValidationError{Message: fmt.Sprintf("retention days must be at most %d", MaxRetentionDays)}
		}
	}
	if i.Status != nil && !IsValidPolicyStatus(*i.Status) {
		return &ValidationError{Message: fmt.Sprintf("invalid status: %s", *i.Status)}
	}
	if i.Priority != nil && (*i.Priority < 0 || *i.Priority > 100) {
		return &ValidationError{Message: "priority must be between 0 and 100"}
	}
	return nil
}

// CleanupExecution records a cleanup job execution.
type CleanupExecution struct {
	ID             string          `db:"id" json:"id"`
	TenantID       string          `db:"tenant_id" json:"tenant_id"`
	PolicyID       string          `db:"policy_id" json:"policy_id"`
	ResourceType   ResourceType    `db:"resource_type" json:"resource_type"`
	Status         ExecutionStatus `db:"status" json:"status"`
	StartedAt      time.Time       `db:"started_at" json:"started_at"`
	CompletedAt    *time.Time      `db:"completed_at" json:"completed_at,omitempty"`
	RecordsScanned int64           `db:"records_scanned" json:"records_scanned"`
	RecordsDeleted int64           `db:"records_deleted" json:"records_deleted"`
	RecordsFailed  int64           `db:"records_failed" json:"records_failed"`
	DryRun         bool            `db:"dry_run" json:"dry_run"`
	ErrorMessage   string          `db:"error_message" json:"error_message,omitempty"`
	Details        json.RawMessage `db:"details" json:"details,omitempty"`
}

// ExecutionStatus defines the status of a cleanup execution.
type ExecutionStatus string

const (
	ExecutionStatusPending   ExecutionStatus = "pending"
	ExecutionStatusRunning   ExecutionStatus = "running"
	ExecutionStatusCompleted ExecutionStatus = "completed"
	ExecutionStatusFailed    ExecutionStatus = "failed"
	ExecutionStatusCancelled ExecutionStatus = "cancelled"
)

// IsValidExecutionStatus checks if an execution status is valid.
func IsValidExecutionStatus(status ExecutionStatus) bool {
	switch status {
	case ExecutionStatusPending, ExecutionStatusRunning, ExecutionStatusCompleted,
		ExecutionStatusFailed, ExecutionStatusCancelled:
		return true
	}
	return false
}

// NewCleanupExecution creates a new cleanup execution record.
func NewCleanupExecution(tenantID, policyID string, resourceType ResourceType, dryRun bool) *CleanupExecution {
	return &CleanupExecution{
		ID:           uuid.NewString(),
		TenantID:     tenantID,
		PolicyID:     policyID,
		ResourceType: resourceType,
		Status:       ExecutionStatusPending,
		StartedAt:    time.Now(),
		DryRun:       dryRun,
	}
}

// MarkRunning marks the execution as running.
func (e *CleanupExecution) MarkRunning() {
	e.Status = ExecutionStatusRunning
}

// MarkCompleted marks the execution as completed.
func (e *CleanupExecution) MarkCompleted() {
	e.Status = ExecutionStatusCompleted
	now := time.Now()
	e.CompletedAt = &now
}

// MarkFailed marks the execution as failed with an error message.
func (e *CleanupExecution) MarkFailed(err error) {
	e.Status = ExecutionStatusFailed
	now := time.Now()
	e.CompletedAt = &now
	if err != nil {
		e.ErrorMessage = err.Error()
	}
}

// Duration returns the execution duration.
func (e *CleanupExecution) Duration() time.Duration {
	end := time.Now()
	if e.CompletedAt != nil {
		end = *e.CompletedAt
	}
	return end.Sub(e.StartedAt)
}

// CleanupStats aggregates cleanup statistics.
type CleanupStats struct {
	TotalExecutions int64            `json:"total_executions"`
	SuccessfulCount int64            `json:"successful_count"`
	FailedCount     int64            `json:"failed_count"`
	TotalDeleted    int64            `json:"total_deleted"`
	TotalScanned    int64            `json:"total_scanned"`
	AvgDurationMs   float64          `json:"avg_duration_ms"`
	ByResourceType  map[string]int64 `json:"by_resource_type"`
	LastExecutionAt *time.Time       `json:"last_execution_at,omitempty"`
	LastSuccessAt   *time.Time       `json:"last_success_at,omitempty"`
}

// CleanupPreview provides an estimate of what would be deleted.
type CleanupPreview struct {
	PolicyID       string       `json:"policy_id"`
	ResourceType   ResourceType `json:"resource_type"`
	RetentionDays  int          `json:"retention_days"`
	CutoffDate     time.Time    `json:"cutoff_date"`
	EstimatedCount int64        `json:"estimated_count"`
	SampleRecords  []string     `json:"sample_records,omitempty"`
	Message        string       `json:"message,omitempty"`
}

// PolicyListFilter defines filters for listing retention policies.
type PolicyListFilter struct {
	ResourceType *ResourceType `json:"resource_type,omitempty"`
	Status       *PolicyStatus `json:"status,omitempty"`
	Search       string        `json:"search,omitempty"`
}

// ExecutionListFilter defines filters for listing cleanup executions.
type ExecutionListFilter struct {
	PolicyID     string           `json:"policy_id,omitempty"`
	ResourceType *ResourceType    `json:"resource_type,omitempty"`
	Status       *ExecutionStatus `json:"status,omitempty"`
	StartDate    *time.Time       `json:"start_date,omitempty"`
	EndDate      *time.Time       `json:"end_date,omitempty"`
	DryRun       *bool            `json:"dry_run,omitempty"`
}
