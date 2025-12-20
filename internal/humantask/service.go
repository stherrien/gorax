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

	data := map[string]interface{}{
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

	data := map[string]interface{}{
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
	var config HumanTaskConfig
	if len(task.Config) > 0 {
		if err := json.Unmarshal(task.Config, &config); err != nil {
			return fmt.Errorf("unmarshal config: %w", err)
		}
	}

	switch config.OnTimeout {
	case TimeoutActionAutoApprove:
		if err := task.Approve(uuid.Nil, map[string]interface{}{
			"comment": "Auto-approved due to timeout",
		}); err != nil {
			return err
		}

	case TimeoutActionAutoReject:
		if err := task.Reject(uuid.Nil, map[string]interface{}{
			"reason": "Auto-rejected due to timeout",
		}); err != nil {
			return err
		}

	case TimeoutActionEscalate:
		// Update assignees to escalation list
		if len(config.EscalateTo) > 0 {
			escalateJSON, err := json.Marshal(config.EscalateTo)
			if err != nil {
				return err
			}
			task.Assignees = escalateJSON

			// Update due date
			if config.Timeout > 0 {
				newDueDate := time.Now().Add(config.Timeout)
				task.DueDate = &newDueDate
			}
		}

	default:
		// Mark as expired
		if err := task.Expire(); err != nil {
			return err
		}
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

func strPtr(s string) *string {
	return &s
}
