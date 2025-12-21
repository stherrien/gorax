package humantask

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// Task types
const (
	TaskTypeApproval = "approval"
	TaskTypeInput    = "input"
	TaskTypeReview   = "review"
)

// Task statuses
const (
	StatusPending   = "pending"
	StatusApproved  = "approved"
	StatusRejected  = "rejected"
	StatusExpired   = "expired"
	StatusCancelled = "cancelled"
)

// Timeout actions
const (
	TimeoutActionAutoApprove = "auto_approve"
	TimeoutActionAutoReject  = "auto_reject"
	TimeoutActionEscalate    = "escalate"
)

// HumanTask represents a human approval/input task in a workflow
type HumanTask struct {
	ID           uuid.UUID       `json:"id" db:"id"`
	TenantID     uuid.UUID       `json:"tenant_id" db:"tenant_id"`
	ExecutionID  uuid.UUID       `json:"execution_id" db:"execution_id"`
	StepID       string          `json:"step_id" db:"step_id"`
	TaskType     string          `json:"task_type" db:"task_type"`
	Title        string          `json:"title" db:"title"`
	Description  string          `json:"description" db:"description"`
	Assignees    json.RawMessage `json:"assignees" db:"assignees"`
	Status       string          `json:"status" db:"status"`
	DueDate      *time.Time      `json:"due_date" db:"due_date"`
	CompletedAt  *time.Time      `json:"completed_at" db:"completed_at"`
	CompletedBy  *uuid.UUID      `json:"completed_by" db:"completed_by"`
	ResponseData json.RawMessage `json:"response_data" db:"response_data"`
	Config       json.RawMessage `json:"config" db:"config"`
	CreatedAt    time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time       `json:"updated_at" db:"updated_at"`
}

// HumanTaskConfig holds configuration for a human task action
type HumanTaskConfig struct {
	TaskType    string        `json:"task_type"`
	Title       string        `json:"title"`
	Description string        `json:"description"`
	Assignees   []string      `json:"assignees"`
	DueDate     time.Duration `json:"due_date"`
	Timeout     time.Duration `json:"timeout"`
	OnTimeout   string        `json:"on_timeout"`
	EscalateTo  []string      `json:"escalate_to"`
	FormFields  []FormField   `json:"form_fields"`
}

// FormField represents a field in an input form
type FormField struct {
	Name     string   `json:"name"`
	Type     string   `json:"type"`
	Label    string   `json:"label"`
	Required bool     `json:"required"`
	Options  []string `json:"options"`
	Default  any      `json:"default"`
}

// CreateTaskRequest represents a request to create a human task
type CreateTaskRequest struct {
	ExecutionID uuid.UUID              `json:"execution_id"`
	StepID      string                 `json:"step_id"`
	TaskType    string                 `json:"task_type"`
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	Assignees   []string               `json:"assignees"`
	DueDate     *time.Time             `json:"due_date"`
	Config      map[string]interface{} `json:"config"`
}

// ApproveTaskRequest represents a request to approve a task
type ApproveTaskRequest struct {
	Comment string                 `json:"comment"`
	Data    map[string]interface{} `json:"data"`
}

// RejectTaskRequest represents a request to reject a task
type RejectTaskRequest struct {
	Reason string                 `json:"reason"`
	Data   map[string]interface{} `json:"data"`
}

// SubmitTaskRequest represents a request to submit an input task
type SubmitTaskRequest struct {
	Data map[string]interface{} `json:"data"`
}

// TaskFilter represents filters for querying tasks
type TaskFilter struct {
	TenantID    uuid.UUID
	ExecutionID *uuid.UUID
	Assignee    *string
	Status      *string
	TaskType    *string
	DueBefore   *time.Time
	DueAfter    *time.Time
	Limit       int
	Offset      int
}

// TaskResponse represents the API response for a task
type TaskResponse struct {
	*HumanTask
	AssigneesList []string               `json:"assignees_list"`
	ConfigData    map[string]interface{} `json:"config_data"`
	ResponseMap   map[string]interface{} `json:"response_map"`
}

// IsPending returns true if the task is pending
func (t *HumanTask) IsPending() bool {
	return t.Status == StatusPending
}

// IsCompleted returns true if the task is completed (approved, rejected, or expired)
func (t *HumanTask) IsCompleted() bool {
	return t.Status == StatusApproved ||
		t.Status == StatusRejected ||
		t.Status == StatusExpired ||
		t.Status == StatusCancelled
}

// IsOverdue returns true if the task is pending and past due date
func (t *HumanTask) IsOverdue() bool {
	return t.IsPending() && t.DueDate != nil && time.Now().After(*t.DueDate)
}

// CanBeCompletedBy checks if a user can complete the task
func (t *HumanTask) CanBeCompletedBy(userID uuid.UUID, roles []string) bool {
	if !t.IsPending() {
		return false
	}

	var assignees []string
	if err := json.Unmarshal(t.Assignees, &assignees); err != nil {
		return false
	}

	// Check if user ID matches
	userIDStr := userID.String()
	for _, assignee := range assignees {
		if assignee == userIDStr {
			return true
		}
	}

	// Check if any role matches
	for _, assignee := range assignees {
		for _, role := range roles {
			if assignee == role {
				return true
			}
		}
	}

	return false
}

// Approve marks the task as approved
func (t *HumanTask) Approve(userID uuid.UUID, data map[string]interface{}) error {
	if !t.IsPending() {
		return ErrTaskNotPending
	}

	t.Status = StatusApproved
	now := time.Now()
	t.CompletedAt = &now
	t.CompletedBy = &userID

	responseData, err := json.Marshal(data)
	if err != nil {
		return err
	}
	t.ResponseData = responseData

	return nil
}

// Reject marks the task as rejected
func (t *HumanTask) Reject(userID uuid.UUID, data map[string]interface{}) error {
	if !t.IsPending() {
		return ErrTaskNotPending
	}

	t.Status = StatusRejected
	now := time.Now()
	t.CompletedAt = &now
	t.CompletedBy = &userID

	responseData, err := json.Marshal(data)
	if err != nil {
		return err
	}
	t.ResponseData = responseData

	return nil
}

// Submit marks the task as completed with input data
func (t *HumanTask) Submit(userID uuid.UUID, data map[string]interface{}) error {
	if !t.IsPending() {
		return ErrTaskNotPending
	}

	t.Status = StatusApproved
	now := time.Now()
	t.CompletedAt = &now
	t.CompletedBy = &userID

	responseData, err := json.Marshal(data)
	if err != nil {
		return err
	}
	t.ResponseData = responseData

	return nil
}

// Expire marks the task as expired
func (t *HumanTask) Expire() error {
	if !t.IsPending() {
		return ErrTaskNotPending
	}

	t.Status = StatusExpired
	now := time.Now()
	t.CompletedAt = &now

	return nil
}

// Cancel marks the task as cancelled
func (t *HumanTask) Cancel() error {
	if !t.IsPending() {
		return ErrTaskNotPending
	}

	t.Status = StatusCancelled
	now := time.Now()
	t.CompletedAt = &now

	return nil
}

// ToResponse converts a HumanTask to a TaskResponse
func (t *HumanTask) ToResponse() (*TaskResponse, error) {
	var assigneesList []string
	if err := json.Unmarshal(t.Assignees, &assigneesList); err != nil {
		return nil, err
	}

	var configData map[string]interface{}
	if len(t.Config) > 0 {
		if err := json.Unmarshal(t.Config, &configData); err != nil {
			return nil, err
		}
	}

	var responseMap map[string]interface{}
	if len(t.ResponseData) > 0 {
		if err := json.Unmarshal(t.ResponseData, &responseMap); err != nil {
			return nil, err
		}
	}

	return &TaskResponse{
		HumanTask:     t,
		AssigneesList: assigneesList,
		ConfigData:    configData,
		ResponseMap:   responseMap,
	}, nil
}
