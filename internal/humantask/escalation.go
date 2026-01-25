package humantask

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// Escalation reasons
const (
	EscalationReasonTimeout = "timeout"
	EscalationReasonManual  = "manual"
)

// Escalation statuses
const (
	EscalationStatusActive     = "active"
	EscalationStatusSuperseded = "superseded"
	EscalationStatusCompleted  = "completed"
)

// TaskEscalation represents an escalation event for a human task
type TaskEscalation struct {
	ID               uuid.UUID       `json:"id" db:"id"`
	TaskID           uuid.UUID       `json:"task_id" db:"task_id"`
	EscalationLevel  int             `json:"escalation_level" db:"escalation_level"`
	EscalatedAt      time.Time       `json:"escalated_at" db:"escalated_at"`
	EscalatedFrom    json.RawMessage `json:"escalated_from" db:"escalated_from"`
	EscalatedTo      json.RawMessage `json:"escalated_to" db:"escalated_to"`
	EscalationReason string          `json:"escalation_reason" db:"escalation_reason"`
	TimeoutMinutes   *int            `json:"timeout_minutes" db:"timeout_minutes"`
	AutoActionTaken  *string         `json:"auto_action_taken" db:"auto_action_taken"`
	Status           string          `json:"status" db:"status"`
	CompletedAt      *time.Time      `json:"completed_at" db:"completed_at"`
	CompletedBy      *uuid.UUID      `json:"completed_by" db:"completed_by"`
	Metadata         json.RawMessage `json:"metadata" db:"metadata"`
	CreatedAt        time.Time       `json:"created_at" db:"created_at"`
}

// EscalationConfig defines the escalation configuration for a task
type EscalationConfig struct {
	Enabled          bool              `json:"enabled"`
	Levels           []EscalationLevel `json:"levels"`
	FinalAction      string            `json:"final_action"` // auto_approve, auto_reject, fail
	NotifyOnEscalate bool              `json:"notify_on_escalate"`
}

// EscalationLevel defines a single escalation level configuration
type EscalationLevel struct {
	Level           int      `json:"level"`
	TimeoutMinutes  int      `json:"timeout_minutes"`
	BackupApprovers []string `json:"backup_approvers"` // User IDs or roles
	NotifyOriginal  bool     `json:"notify_original"`  // Whether to notify original assignees
}

// EscalationHistory represents the API response for escalation history
type EscalationHistory struct {
	TaskID      uuid.UUID           `json:"task_id"`
	Escalations []EscalationSummary `json:"escalations"`
}

// EscalationSummary provides a summary of an escalation event
type EscalationSummary struct {
	ID             uuid.UUID  `json:"id"`
	Level          int        `json:"level"`
	EscalatedAt    time.Time  `json:"escalated_at"`
	FromAssignees  []string   `json:"from_assignees"`
	ToAssignees    []string   `json:"to_assignees"`
	Reason         string     `json:"reason"`
	TimeoutMinutes *int       `json:"timeout_minutes,omitempty"`
	AutoAction     *string    `json:"auto_action,omitempty"`
	Status         string     `json:"status"`
	CompletedAt    *time.Time `json:"completed_at,omitempty"`
}

// UpdateEscalationRequest represents a request to update escalation settings
type UpdateEscalationRequest struct {
	Config EscalationConfig `json:"config"`
}

// GetFromAssignees returns the escalated_from field as a string slice
func (e *TaskEscalation) GetFromAssignees() []string {
	if len(e.EscalatedFrom) == 0 {
		return nil
	}
	var assignees []string
	if err := json.Unmarshal(e.EscalatedFrom, &assignees); err != nil {
		return nil
	}
	return assignees
}

// GetToAssignees returns the escalated_to field as a string slice
func (e *TaskEscalation) GetToAssignees() []string {
	if len(e.EscalatedTo) == 0 {
		return nil
	}
	var assignees []string
	if err := json.Unmarshal(e.EscalatedTo, &assignees); err != nil {
		return nil
	}
	return assignees
}

// ToSummary converts a TaskEscalation to an EscalationSummary
func (e *TaskEscalation) ToSummary() EscalationSummary {
	return EscalationSummary{
		ID:             e.ID,
		Level:          e.EscalationLevel,
		EscalatedAt:    e.EscalatedAt,
		FromAssignees:  e.GetFromAssignees(),
		ToAssignees:    e.GetToAssignees(),
		Reason:         e.EscalationReason,
		TimeoutMinutes: e.TimeoutMinutes,
		AutoAction:     e.AutoActionTaken,
		Status:         e.Status,
		CompletedAt:    e.CompletedAt,
	}
}

// Complete marks the escalation as completed
func (e *TaskEscalation) Complete(userID *uuid.UUID) {
	e.Status = EscalationStatusCompleted
	now := time.Now()
	e.CompletedAt = &now
	e.CompletedBy = userID
}

// Supersede marks the escalation as superseded by a higher level
func (e *TaskEscalation) Supersede() {
	e.Status = EscalationStatusSuperseded
}

// ParseEscalationConfig extracts EscalationConfig from task config
func ParseEscalationConfig(configJSON json.RawMessage) (*EscalationConfig, error) {
	if len(configJSON) == 0 {
		return nil, nil
	}

	var config struct {
		Escalation *EscalationConfig `json:"escalation"`
	}

	if err := json.Unmarshal(configJSON, &config); err != nil {
		return nil, err
	}

	return config.Escalation, nil
}

// GetEscalationLevelConfig returns the configuration for a specific escalation level
func (ec *EscalationConfig) GetLevelConfig(level int) *EscalationLevel {
	for i := range ec.Levels {
		if ec.Levels[i].Level == level {
			return &ec.Levels[i]
		}
	}
	return nil
}

// GetMaxLevel returns the maximum escalation level configured
func (ec *EscalationConfig) GetMaxLevel() int {
	maxLevel := 0
	for _, level := range ec.Levels {
		if level.Level > maxLevel {
			maxLevel = level.Level
		}
	}
	return maxLevel
}

// GetNextLevel returns the next escalation level configuration
func (ec *EscalationConfig) GetNextLevel(currentLevel int) *EscalationLevel {
	return ec.GetLevelConfig(currentLevel + 1)
}
