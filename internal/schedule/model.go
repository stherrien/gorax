package schedule

import (
	"slices"
	"time"
)

// OverlapPolicy defines how to handle overlapping executions
type OverlapPolicy string

const (
	// OverlapPolicySkip skips the new execution if previous is still running
	OverlapPolicySkip OverlapPolicy = "skip"
	// OverlapPolicyQueue queues the new execution for when the current completes
	OverlapPolicyQueue OverlapPolicy = "queue"
	// OverlapPolicyTerminate terminates the current execution and starts new one
	OverlapPolicyTerminate OverlapPolicy = "terminate"
)

// ValidOverlapPolicies contains all valid overlap policy values
var ValidOverlapPolicies = []OverlapPolicy{
	OverlapPolicySkip,
	OverlapPolicyQueue,
	OverlapPolicyTerminate,
}

// IsValid checks if the overlap policy is valid
func (p OverlapPolicy) IsValid() bool {
	return slices.Contains(ValidOverlapPolicies, p)
}

// Schedule represents a scheduled workflow execution
type Schedule struct {
	ID                 string        `db:"id" json:"id"`
	TenantID           string        `db:"tenant_id" json:"tenant_id"`
	WorkflowID         string        `db:"workflow_id" json:"workflow_id"`
	Name               string        `db:"name" json:"name"`
	CronExpression     string        `db:"cron_expression" json:"cron_expression"`
	Timezone           string        `db:"timezone" json:"timezone"`
	OverlapPolicy      OverlapPolicy `db:"overlap_policy" json:"overlap_policy"`
	Enabled            bool          `db:"enabled" json:"enabled"`
	NextRunAt          *time.Time    `db:"next_run_at" json:"next_run_at,omitempty"`
	LastRunAt          *time.Time    `db:"last_run_at" json:"last_run_at,omitempty"`
	LastExecutionID    *string       `db:"last_execution_id" json:"last_execution_id,omitempty"`
	RunningExecutionID *string       `db:"running_execution_id" json:"running_execution_id,omitempty"`
	CreatedBy          string        `db:"created_by" json:"created_by"`
	CreatedAt          time.Time     `db:"created_at" json:"created_at"`
	UpdatedAt          time.Time     `db:"updated_at" json:"updated_at"`
}

// CreateScheduleInput represents input for creating a schedule
type CreateScheduleInput struct {
	Name           string        `json:"name" validate:"required,min=1,max=255"`
	CronExpression string        `json:"cron_expression" validate:"required"`
	Timezone       string        `json:"timezone,omitempty"`
	OverlapPolicy  OverlapPolicy `json:"overlap_policy,omitempty"`
	Enabled        bool          `json:"enabled"`
}

// UpdateScheduleInput represents input for updating a schedule
type UpdateScheduleInput struct {
	Name           *string        `json:"name,omitempty"`
	CronExpression *string        `json:"cron_expression,omitempty"`
	Timezone       *string        `json:"timezone,omitempty"`
	OverlapPolicy  *OverlapPolicy `json:"overlap_policy,omitempty"`
	Enabled        *bool          `json:"enabled,omitempty"`
}

// ScheduleWithWorkflow represents a schedule with workflow details
type ScheduleWithWorkflow struct {
	Schedule
	WorkflowName   string `db:"workflow_name" json:"workflow_name"`
	WorkflowStatus string `db:"workflow_status" json:"workflow_status"`
}

// ExecutionLogStatus represents the status of a schedule execution
type ExecutionLogStatus string

const (
	ExecutionLogStatusPending    ExecutionLogStatus = "pending"
	ExecutionLogStatusRunning    ExecutionLogStatus = "running"
	ExecutionLogStatusCompleted  ExecutionLogStatus = "completed"
	ExecutionLogStatusFailed     ExecutionLogStatus = "failed"
	ExecutionLogStatusSkipped    ExecutionLogStatus = "skipped"
	ExecutionLogStatusTerminated ExecutionLogStatus = "terminated"
)

// ExecutionLog represents a record of a schedule execution attempt
type ExecutionLog struct {
	ID            string             `db:"id" json:"id"`
	TenantID      string             `db:"tenant_id" json:"tenant_id"`
	ScheduleID    string             `db:"schedule_id" json:"schedule_id"`
	ExecutionID   *string            `db:"execution_id" json:"execution_id,omitempty"`
	Status        ExecutionLogStatus `db:"status" json:"status"`
	StartedAt     *time.Time         `db:"started_at" json:"started_at,omitempty"`
	CompletedAt   *time.Time         `db:"completed_at" json:"completed_at,omitempty"`
	ErrorMessage  *string            `db:"error_message" json:"error_message,omitempty"`
	TriggerTime   time.Time          `db:"trigger_time" json:"trigger_time"`
	SkippedReason *string            `db:"skipped_reason" json:"skipped_reason,omitempty"`
	CreatedAt     time.Time          `db:"created_at" json:"created_at"`
	UpdatedAt     time.Time          `db:"updated_at" json:"updated_at"`
}

// ExecutionLogListParams represents parameters for listing execution logs
type ExecutionLogListParams struct {
	ScheduleID string
	Status     *ExecutionLogStatus
	Limit      int
	Offset     int
}

// ScheduleExecution represents a workflow execution triggered by a schedule
type ScheduleExecution struct {
	ID          string     `db:"id" json:"id"`
	ScheduleID  string     `db:"schedule_id" json:"schedule_id"`
	ExecutionID string     `db:"execution_id" json:"execution_id"`
	Status      string     `db:"status" json:"status"`
	StartedAt   *time.Time `db:"started_at" json:"started_at,omitempty"`
	CompletedAt *time.Time `db:"completed_at" json:"completed_at,omitempty"`
	CreatedAt   time.Time  `db:"created_at" json:"created_at"`
}

// ValidationError represents a validation error
type ValidationError struct {
	Message string
}

func (e *ValidationError) Error() string {
	return e.Message
}
