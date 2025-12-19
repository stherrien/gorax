package schedule

import (
	"time"
)

// Schedule represents a scheduled workflow execution
type Schedule struct {
	ID              string     `db:"id" json:"id"`
	TenantID        string     `db:"tenant_id" json:"tenant_id"`
	WorkflowID      string     `db:"workflow_id" json:"workflow_id"`
	Name            string     `db:"name" json:"name"`
	CronExpression  string     `db:"cron_expression" json:"cron_expression"`
	Timezone        string     `db:"timezone" json:"timezone"`
	Enabled         bool       `db:"enabled" json:"enabled"`
	NextRunAt       *time.Time `db:"next_run_at" json:"next_run_at,omitempty"`
	LastRunAt       *time.Time `db:"last_run_at" json:"last_run_at,omitempty"`
	LastExecutionID *string    `db:"last_execution_id" json:"last_execution_id,omitempty"`
	CreatedBy       string     `db:"created_by" json:"created_by"`
	CreatedAt       time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt       time.Time  `db:"updated_at" json:"updated_at"`
}

// CreateScheduleInput represents input for creating a schedule
type CreateScheduleInput struct {
	Name           string `json:"name" validate:"required,min=1,max=255"`
	CronExpression string `json:"cron_expression" validate:"required"`
	Timezone       string `json:"timezone,omitempty"`
	Enabled        bool   `json:"enabled"`
}

// UpdateScheduleInput represents input for updating a schedule
type UpdateScheduleInput struct {
	Name           *string `json:"name,omitempty"`
	CronExpression *string `json:"cron_expression,omitempty"`
	Timezone       *string `json:"timezone,omitempty"`
	Enabled        *bool   `json:"enabled,omitempty"`
}

// ScheduleWithWorkflow represents a schedule with workflow details
type ScheduleWithWorkflow struct {
	Schedule
	WorkflowName   string `db:"workflow_name" json:"workflow_name"`
	WorkflowStatus string `db:"workflow_status" json:"workflow_status"`
}

// ValidationError represents a validation error
type ValidationError struct {
	Message string
}

func (e *ValidationError) Error() string {
	return e.Message
}
