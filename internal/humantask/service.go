package humantask

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// NotificationService defines the interface for sending notifications
type NotificationService interface {
	NotifyTaskAssigned(ctx context.Context, task *HumanTask) error
	NotifyTaskCompleted(ctx context.Context, task *HumanTask) error
	NotifyTaskOverdue(ctx context.Context, task *HumanTask) error
	NotifyTaskEscalated(ctx context.Context, task *HumanTask, escalation *TaskEscalation) error
}

// Service defines the human task business logic
type Service interface {
	CreateTask(ctx context.Context, tenantID uuid.UUID, req CreateTaskRequest) (*HumanTask, error)
	GetTask(ctx context.Context, tenantID uuid.UUID, taskID uuid.UUID) (*HumanTask, error)
	ListTasks(ctx context.Context, filter TaskFilter) ([]*HumanTask, error)
	ApproveTask(ctx context.Context, tenantID uuid.UUID, taskID uuid.UUID, userID uuid.UUID, roles []string, req ApproveTaskRequest) error
	RejectTask(ctx context.Context, tenantID uuid.UUID, taskID uuid.UUID, userID uuid.UUID, roles []string, req RejectTaskRequest) error
	SubmitTask(ctx context.Context, tenantID uuid.UUID, taskID uuid.UUID, userID uuid.UUID, roles []string, req SubmitTaskRequest) error
	ProcessOverdueTasks(ctx context.Context, tenantID uuid.UUID) error
	CancelTasksByExecution(ctx context.Context, tenantID uuid.UUID, executionID uuid.UUID) error

	// Escalation methods
	GetEscalationHistory(ctx context.Context, tenantID uuid.UUID, taskID uuid.UUID) (*EscalationHistory, error)
	UpdateEscalationConfig(ctx context.Context, tenantID uuid.UUID, taskID uuid.UUID, req UpdateEscalationRequest) error
}

type service struct {
	repo         Repository
	notification NotificationService
}

// NewService creates a new human task service
func NewService(repo Repository, notification NotificationService) Service {
	return &service{
		repo:         repo,
		notification: notification,
	}
}

func (s *service) CreateTask(ctx context.Context, tenantID uuid.UUID, req CreateTaskRequest) (*HumanTask, error) {
	if err := s.validateCreateRequest(req); err != nil {
		return nil, err
	}

	assigneesJSON, err := json.Marshal(req.Assignees)
	if err != nil {
		return nil, fmt.Errorf("marshal assignees: %w", err)
	}

	configJSON, err := json.Marshal(req.Config)
	if err != nil {
		return nil, fmt.Errorf("marshal config: %w", err)
	}

	task := &HumanTask{
		TenantID:    tenantID,
		ExecutionID: req.ExecutionID,
		StepID:      req.StepID,
		TaskType:    req.TaskType,
		Title:       req.Title,
		Description: req.Description,
		Assignees:   assigneesJSON,
		Status:      StatusPending,
		DueDate:     req.DueDate,
		Config:      configJSON,
	}

	// Set max escalation level from config if present
	escConfig, _ := ParseEscalationConfig(configJSON)
	if escConfig != nil && escConfig.Enabled {
		task.MaxEscalationLevel = escConfig.GetMaxLevel()
	}

	if err := s.repo.Create(ctx, task); err != nil {
		return nil, fmt.Errorf("create task: %w", err)
	}

	// Send notification asynchronously
	go func() {
		if err := s.notification.NotifyTaskAssigned(context.Background(), task); err != nil {
			// Log error but don't fail the task creation
			fmt.Printf("failed to send notification: %v\n", err)
		}
	}()

	return task, nil
}

func (s *service) GetTask(ctx context.Context, tenantID uuid.UUID, taskID uuid.UUID) (*HumanTask, error) {
	task, err := s.repo.GetByID(ctx, taskID)
	if err != nil {
		return nil, err
	}

	if task.TenantID != tenantID {
		return nil, ErrTaskNotFound
	}

	return task, nil
}

func (s *service) ListTasks(ctx context.Context, filter TaskFilter) ([]*HumanTask, error) {
	return s.repo.List(ctx, filter)
}

func (s *service) ApproveTask(ctx context.Context, tenantID uuid.UUID, taskID uuid.UUID,
	userID uuid.UUID, roles []string, req ApproveTaskRequest) error {

	task, err := s.getAndValidateTask(ctx, tenantID, taskID, userID, roles)
	if err != nil {
		return err
	}

	data := map[string]any{
		"comment": req.Comment,
	}
	for k, v := range req.Data {
		data[k] = v
	}

	if err := task.Approve(userID, data); err != nil {
		return err
	}

	if err := s.repo.Update(ctx, task); err != nil {
		return fmt.Errorf("update task: %w", err)
	}

	// Complete any active escalations
	if err := s.repo.CompleteEscalationsByTaskID(ctx, taskID, &userID); err != nil {
		fmt.Printf("failed to complete escalations: %v\n", err)
	}

	// Send notification asynchronously
	go func() {
		if err := s.notification.NotifyTaskCompleted(context.Background(), task); err != nil {
			fmt.Printf("failed to send notification: %v\n", err)
		}
	}()

	return nil
}

func (s *service) RejectTask(ctx context.Context, tenantID uuid.UUID, taskID uuid.UUID,
	userID uuid.UUID, roles []string, req RejectTaskRequest) error {

	task, err := s.getAndValidateTask(ctx, tenantID, taskID, userID, roles)
	if err != nil {
		return err
	}

	data := map[string]any{
		"reason": req.Reason,
	}
	for k, v := range req.Data {
		data[k] = v
	}

	if err := task.Reject(userID, data); err != nil {
		return err
	}

	if err := s.repo.Update(ctx, task); err != nil {
		return fmt.Errorf("update task: %w", err)
	}

	// Complete any active escalations
	if err := s.repo.CompleteEscalationsByTaskID(ctx, taskID, &userID); err != nil {
		fmt.Printf("failed to complete escalations: %v\n", err)
	}

	// Send notification asynchronously
	go func() {
		if err := s.notification.NotifyTaskCompleted(context.Background(), task); err != nil {
			fmt.Printf("failed to send notification: %v\n", err)
		}
	}()

	return nil
}

func (s *service) SubmitTask(ctx context.Context, tenantID uuid.UUID, taskID uuid.UUID,
	userID uuid.UUID, roles []string, req SubmitTaskRequest) error {

	task, err := s.getAndValidateTask(ctx, tenantID, taskID, userID, roles)
	if err != nil {
		return err
	}

	if err := task.Submit(userID, req.Data); err != nil {
		return err
	}

	if err := s.repo.Update(ctx, task); err != nil {
		return fmt.Errorf("update task: %w", err)
	}

	// Complete any active escalations
	if err := s.repo.CompleteEscalationsByTaskID(ctx, taskID, &userID); err != nil {
		fmt.Printf("failed to complete escalations: %v\n", err)
	}

	// Send notification asynchronously
	go func() {
		if err := s.notification.NotifyTaskCompleted(context.Background(), task); err != nil {
			fmt.Printf("failed to send notification: %v\n", err)
		}
	}()

	return nil
}

func (s *service) ProcessOverdueTasks(ctx context.Context, tenantID uuid.UUID) error {
	tasks, err := s.repo.GetOverdueTasks(ctx, tenantID)
	if err != nil {
		return fmt.Errorf("get overdue tasks: %w", err)
	}

	for _, task := range tasks {
		if err := s.handleOverdueTask(ctx, task); err != nil {
			fmt.Printf("failed to handle overdue task %s: %v\n", task.ID, err)
			continue
		}
	}

	return nil
}

func (s *service) CancelTasksByExecution(ctx context.Context, tenantID uuid.UUID, executionID uuid.UUID) error {
	filter := TaskFilter{
		TenantID:    tenantID,
		ExecutionID: &executionID,
		Status:      strPtr(StatusPending),
		Limit:       1000,
	}

	tasks, err := s.repo.List(ctx, filter)
	if err != nil {
		return fmt.Errorf("list tasks: %w", err)
	}

	for _, task := range tasks {
		if err := task.Cancel(); err != nil {
			continue
		}

		if err := s.repo.Update(ctx, task); err != nil {
			fmt.Printf("failed to cancel task %s: %v\n", task.ID, err)
		}
	}

	return nil
}

// Private helper methods

func (s *service) validateCreateRequest(req CreateTaskRequest) error {
	if req.TaskType != TaskTypeApproval &&
		req.TaskType != TaskTypeInput &&
		req.TaskType != TaskTypeReview {
		return ErrInvalidTaskType
	}

	if req.Title == "" {
		return ErrMissingRequiredField
	}

	if len(req.Assignees) == 0 {
		return ErrMissingRequiredField
	}

	return nil
}

func (s *service) getAndValidateTask(ctx context.Context, tenantID uuid.UUID, taskID uuid.UUID,
	userID uuid.UUID, roles []string) (*HumanTask, error) {

	task, err := s.repo.GetByID(ctx, taskID)
	if err != nil {
		return nil, err
	}

	if task.TenantID != tenantID {
		return nil, ErrTaskNotFound
	}

	if !task.IsPending() {
		return nil, ErrTaskNotPending
	}

	if !task.CanBeCompletedBy(userID, roles) {
		return nil, ErrUnauthorized
	}

	return task, nil
}

func (s *service) handleOverdueTask(ctx context.Context, task *HumanTask) error {
	// Parse escalation config
	escConfig, err := ParseEscalationConfig(task.Config)
	if err != nil {
		return fmt.Errorf("parse escalation config: %w", err)
	}

	// Check if escalation is configured
	if escConfig != nil && escConfig.Enabled {
		return s.handleEscalation(ctx, task, escConfig)
	}

	// Fall back to legacy timeout handling
	return s.handleLegacyTimeout(ctx, task)
}

func (s *service) handleEscalation(ctx context.Context, task *HumanTask, config *EscalationConfig) error {
	nextLevel := task.EscalationLevel + 1
	levelConfig := config.GetLevelConfig(nextLevel)

	// Check if we can escalate to next level
	if levelConfig != nil {
		return s.escalateToLevel(ctx, task, levelConfig, config)
	}

	// No more escalation levels - apply final action
	return s.applyFinalAction(ctx, task, config)
}

func (s *service) escalateToLevel(ctx context.Context, task *HumanTask, levelConfig *EscalationLevel, escConfig *EscalationConfig) error {
	// Get current assignees before escalation
	var currentAssignees []string
	if err := json.Unmarshal(task.Assignees, &currentAssignees); err != nil {
		currentAssignees = []string{}
	}

	// Determine new due date
	var newDueDate *time.Time
	if levelConfig.TimeoutMinutes > 0 {
		due := time.Now().Add(time.Duration(levelConfig.TimeoutMinutes) * time.Minute)
		newDueDate = &due
	}

	// Update task with new assignees and escalation level
	if err := task.Escalate(levelConfig.BackupApprovers, newDueDate); err != nil {
		return fmt.Errorf("escalate task: %w", err)
	}

	// Mark any active escalations as superseded
	activeEsc, _ := s.repo.GetActiveEscalation(ctx, task.ID)
	if activeEsc != nil {
		activeEsc.Supersede()
		if err := s.repo.UpdateEscalation(ctx, activeEsc); err != nil {
			fmt.Printf("failed to supersede escalation: %v\n", err)
		}
	}

	// Create escalation record
	escalatedFrom, _ := json.Marshal(currentAssignees)
	escalatedTo, _ := json.Marshal(levelConfig.BackupApprovers)
	timeoutMinutes := levelConfig.TimeoutMinutes

	escalation := &TaskEscalation{
		TaskID:           task.ID,
		EscalationLevel:  levelConfig.Level,
		EscalatedFrom:    escalatedFrom,
		EscalatedTo:      escalatedTo,
		EscalationReason: EscalationReasonTimeout,
		TimeoutMinutes:   &timeoutMinutes,
		Status:           EscalationStatusActive,
		Metadata:         json.RawMessage("{}"),
	}

	if err := s.repo.CreateEscalation(ctx, escalation); err != nil {
		return fmt.Errorf("create escalation: %w", err)
	}

	// Update task in database
	if err := s.repo.Update(ctx, task); err != nil {
		return fmt.Errorf("update task: %w", err)
	}

	// Send escalation notification
	if escConfig.NotifyOnEscalate {
		go func() {
			if err := s.notification.NotifyTaskEscalated(context.Background(), task, escalation); err != nil {
				fmt.Printf("failed to send escalation notification: %v\n", err)
			}
		}()
	}

	return nil
}

func (s *service) applyFinalAction(ctx context.Context, task *HumanTask, config *EscalationConfig) error {
	var currentAssignees []string
	if err := json.Unmarshal(task.Assignees, &currentAssignees); err != nil {
		currentAssignees = []string{}
	}

	var autoAction string

	switch config.FinalAction {
	case TimeoutActionAutoApprove:
		autoAction = TimeoutActionAutoApprove
		if err := task.Approve(uuid.Nil, map[string]any{
			"comment":     "Auto-approved due to timeout after all escalation levels exhausted",
			"auto_action": true,
		}); err != nil {
			return err
		}

	case TimeoutActionAutoReject:
		autoAction = TimeoutActionAutoReject
		if err := task.Reject(uuid.Nil, map[string]any{
			"reason":      "Auto-rejected due to timeout after all escalation levels exhausted",
			"auto_action": true,
		}); err != nil {
			return err
		}

	default:
		// Mark as expired
		if err := task.Expire(); err != nil {
			return err
		}
	}

	// Create final escalation record with auto action
	escalatedFrom, _ := json.Marshal(currentAssignees)

	escalation := &TaskEscalation{
		TaskID:           task.ID,
		EscalationLevel:  task.EscalationLevel + 1,
		EscalatedFrom:    escalatedFrom,
		EscalatedTo:      json.RawMessage("[]"),
		EscalationReason: EscalationReasonTimeout,
		AutoActionTaken:  &autoAction,
		Status:           EscalationStatusCompleted,
		Metadata:         json.RawMessage("{}"),
	}

	now := time.Now()
	escalation.CompletedAt = &now

	if err := s.repo.CreateEscalation(ctx, escalation); err != nil {
		fmt.Printf("failed to create final escalation: %v\n", err)
	}

	// Complete any active escalations
	if err := s.repo.CompleteEscalationsByTaskID(ctx, task.ID, nil); err != nil {
		fmt.Printf("failed to complete escalations: %v\n", err)
	}

	if err := s.repo.Update(ctx, task); err != nil {
		return fmt.Errorf("update task: %w", err)
	}

	// Send notification
	go func() {
		if err := s.notification.NotifyTaskCompleted(context.Background(), task); err != nil {
			fmt.Printf("failed to send notification: %v\n", err)
		}
	}()

	return nil
}

func (s *service) handleLegacyTimeout(ctx context.Context, task *HumanTask) error {
	var config HumanTaskConfig
	if len(task.Config) > 0 {
		if err := json.Unmarshal(task.Config, &config); err != nil {
			return fmt.Errorf("unmarshal config: %w", err)
		}
	}

	switch config.OnTimeout {
	case TimeoutActionAutoApprove:
		if err := task.Approve(uuid.Nil, map[string]any{
			"comment": "Auto-approved due to timeout",
		}); err != nil {
			return err
		}

	case TimeoutActionAutoReject:
		if err := task.Reject(uuid.Nil, map[string]any{
			"reason": "Auto-rejected due to timeout",
		}); err != nil {
			return err
		}

	case TimeoutActionEscalate:
		// Legacy single-level escalation
		if len(config.EscalateTo) > 0 {
			escalateJSON, err := json.Marshal(config.EscalateTo)
			if err != nil {
				return err
			}
			task.Assignees = escalateJSON
			task.EscalationLevel = 1

			if config.Timeout > 0 {
				newDueDate := time.Now().Add(config.Timeout)
				task.DueDate = &newDueDate
			}

			now := time.Now()
			task.LastEscalatedAt = &now
		}

	default:
		if err := task.Expire(); err != nil {
			return err
		}
	}

	if err := s.repo.Update(ctx, task); err != nil {
		return fmt.Errorf("update task: %w", err)
	}

	go func() {
		if err := s.notification.NotifyTaskCompleted(context.Background(), task); err != nil {
			fmt.Printf("failed to send notification: %v\n", err)
		}
	}()

	return nil
}

// GetEscalationHistory returns the escalation history for a task
func (s *service) GetEscalationHistory(ctx context.Context, tenantID uuid.UUID, taskID uuid.UUID) (*EscalationHistory, error) {
	task, err := s.repo.GetByID(ctx, taskID)
	if err != nil {
		return nil, err
	}

	if task.TenantID != tenantID {
		return nil, ErrTaskNotFound
	}

	escalations, err := s.repo.GetEscalationsByTaskID(ctx, taskID)
	if err != nil {
		return nil, fmt.Errorf("get escalations: %w", err)
	}

	summaries := make([]EscalationSummary, len(escalations))
	for i, esc := range escalations {
		summaries[i] = esc.ToSummary()
	}

	return &EscalationHistory{
		TaskID:      taskID,
		Escalations: summaries,
	}, nil
}

// UpdateEscalationConfig updates the escalation configuration for a task
func (s *service) UpdateEscalationConfig(ctx context.Context, tenantID uuid.UUID, taskID uuid.UUID, req UpdateEscalationRequest) error {
	task, err := s.repo.GetByID(ctx, taskID)
	if err != nil {
		return err
	}

	if task.TenantID != tenantID {
		return ErrTaskNotFound
	}

	if !task.IsPending() {
		return ErrTaskNotPending
	}

	// Update config
	var configMap map[string]any
	if len(task.Config) > 0 {
		if err := json.Unmarshal(task.Config, &configMap); err != nil {
			configMap = make(map[string]any)
		}
	} else {
		configMap = make(map[string]any)
	}

	configMap["escalation"] = req.Config

	configJSON, err := json.Marshal(configMap)
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}

	task.Config = configJSON
	task.MaxEscalationLevel = req.Config.GetMaxLevel()

	if err := s.repo.Update(ctx, task); err != nil {
		return fmt.Errorf("update task: %w", err)
	}

	return nil
}

func strPtr(s string) *string {
	return &s
}
