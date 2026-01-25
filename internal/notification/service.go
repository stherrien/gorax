package notification

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"sync"

	"github.com/google/uuid"

	"github.com/gorax/gorax/internal/humantask"
)

// getAssignees extracts assignees from the task's Assignees JSON field
func getAssignees(task *humantask.HumanTask) []string {
	if len(task.Assignees) == 0 {
		return nil
	}
	var assignees []string
	if err := json.Unmarshal(task.Assignees, &assignees); err != nil {
		return nil
	}
	return assignees
}

// getCompletedByString converts CompletedBy UUID pointer to string
func getCompletedByString(task *humantask.HumanTask) string {
	if task.CompletedBy == nil {
		return ""
	}
	return task.CompletedBy.String()
}

// Config holds notification service configuration
type Config struct {
	// Enabled channels
	EnableEmail bool
	EnableSlack bool
	EnableInApp bool

	// Email configuration
	Email EmailConfig

	// Slack configuration
	Slack SlackConfig

	// Default channels for notifications
	DefaultChannels []string // email, slack, inapp
}

// Service implements notification delivery across multiple channels
type Service struct {
	logger        *slog.Logger
	config        Config
	emailSender   *EmailSender
	slackNotifier *SlackNotifier
	inAppService  *InAppService
}

// NewService creates a new notification service
func NewService(logger *slog.Logger, config Config, inAppService *InAppService) (*Service, error) {
	service := &Service{
		logger:       logger,
		config:       config,
		inAppService: inAppService,
	}

	// Initialize email sender if enabled
	if config.EnableEmail {
		emailSender, err := NewEmailSender(config.Email)
		if err != nil {
			return nil, fmt.Errorf("failed to initialize email sender: %w", err)
		}
		service.emailSender = emailSender
	}

	// Initialize Slack notifier if enabled
	if config.EnableSlack {
		slackNotifier, err := NewSlackNotifier(config.Slack)
		if err != nil {
			return nil, fmt.Errorf("failed to initialize Slack notifier: %w", err)
		}
		service.slackNotifier = slackNotifier
	}

	return service, nil
}

// NotifyTaskAssigned sends a notification when a task is assigned
func (s *Service) NotifyTaskAssigned(ctx context.Context, task *humantask.HumanTask) error {
	assignees := getAssignees(task)

	s.logger.Info("task assigned notification",
		"task_id", task.ID,
		"task_type", task.TaskType,
		"title", task.Title,
		"assignees", assignees,
	)

	if len(assignees) == 0 {
		return nil
	}

	// Send notifications across all enabled channels asynchronously
	var wg sync.WaitGroup
	errors := make(chan error, 3)

	taskURL := fmt.Sprintf("/tasks/%s", task.ID)
	dueStr := "Not set"
	if task.DueDate != nil {
		dueStr = task.DueDate.Format("2006-01-02 15:04")
	}

	// In-app notification - notify each assignee
	if s.config.EnableInApp && s.inAppService != nil {
		for _, assignee := range assignees {
			assignee := assignee // capture for goroutine
			wg.Add(1)
			go func() {
				defer wg.Done()
				if err := s.inAppService.NotifyTaskAssigned(ctx, task.TenantID, assignee, task.Title, taskURL); err != nil {
					s.logger.Error("failed to send in-app notification", "error", err, "assignee", assignee)
					errors <- err
				}
			}()
		}
	}

	// Email notification - send to all assignees that look like emails
	if s.config.EnableEmail && s.emailSender != nil {
		wg.Add(1)
		go func() {
			defer wg.Done()
			email := Email{
				To:      assignees,
				Subject: fmt.Sprintf("Task Assigned: %s", task.Title),
				HTMLBody: fmt.Sprintf(`
					<h2>Task Assigned</h2>
					<p>You have been assigned a new task:</p>
					<h3>%s</h3>
					<p><strong>Type:</strong> %s</p>
					<p><strong>Due:</strong> %s</p>
					<p><a href="%s">View Task</a></p>
				`, task.Title, task.TaskType, dueStr, taskURL),
				TextBody: fmt.Sprintf("Task Assigned: %s\nType: %s\nDue: %s\nView: %s",
					task.Title, task.TaskType, dueStr, taskURL),
			}

			if err := s.emailSender.Send(ctx, email); err != nil {
				s.logger.Error("failed to send email notification", "error", err)
				errors <- err
			}
		}()
	}

	// Slack notification
	if s.config.EnableSlack && s.slackNotifier != nil {
		wg.Add(1)
		go func() {
			defer wg.Done()
			msg := BuildTaskAssignedMessage(assignees[0], task.Title, taskURL)

			if err := s.slackNotifier.Send(ctx, msg); err != nil {
				s.logger.Error("failed to send Slack notification", "error", err)
				errors <- err
			}
		}()
	}

	// Wait for all notifications to complete
	wg.Wait()
	close(errors)

	// Log any errors but don't fail the operation
	for err := range errors {
		s.logger.Warn("notification delivery failed", "error", err)
	}

	return nil
}

// NotifyTaskCompleted sends a notification when a task is completed
func (s *Service) NotifyTaskCompleted(ctx context.Context, task *humantask.HumanTask) error {
	completedBy := getCompletedByString(task)

	s.logger.Info("task completed notification",
		"task_id", task.ID,
		"status", task.Status,
		"completed_by", completedBy,
	)

	var wg sync.WaitGroup
	errors := make(chan error, 3)

	// Slack notification
	if s.config.EnableSlack && s.slackNotifier != nil {
		wg.Add(1)
		go func() {
			defer wg.Done()
			msg := BuildTaskCompletedMessage(task.Title, task.Status, completedBy)

			if err := s.slackNotifier.Send(ctx, msg); err != nil {
				s.logger.Error("failed to send Slack notification", "error", err)
				errors <- err
			}
		}()
	}

	wg.Wait()
	close(errors)

	for err := range errors {
		s.logger.Warn("notification delivery failed", "error", err)
	}

	return nil
}

// NotifyTaskOverdue sends a notification when a task is overdue
func (s *Service) NotifyTaskOverdue(ctx context.Context, task *humantask.HumanTask) error {
	assignees := getAssignees(task)

	s.logger.Warn("task overdue notification",
		"task_id", task.ID,
		"due_date", task.DueDate,
		"assignees", assignees,
	)

	if len(assignees) == 0 {
		return nil
	}

	var wg sync.WaitGroup
	errors := make(chan error, 3)

	taskURL := fmt.Sprintf("/tasks/%s", task.ID)
	dueStr := "Not set"
	if task.DueDate != nil {
		dueStr = task.DueDate.Format("2006-01-02 15:04")
	}

	// In-app notification - notify each assignee
	if s.config.EnableInApp && s.inAppService != nil {
		for _, assignee := range assignees {
			assignee := assignee // capture for goroutine
			wg.Add(1)
			go func() {
				defer wg.Done()
				if err := s.inAppService.NotifyTaskOverdue(ctx, task.TenantID, assignee, task.Title, taskURL); err != nil {
					s.logger.Error("failed to send in-app notification", "error", err, "assignee", assignee)
					errors <- err
				}
			}()
		}
	}

	// Email notification
	if s.config.EnableEmail && s.emailSender != nil {
		wg.Add(1)
		go func() {
			defer wg.Done()
			email := Email{
				To:      assignees,
				Subject: fmt.Sprintf("Task Overdue: %s", task.Title),
				HTMLBody: fmt.Sprintf(`
					<h2>⚠️ Task Overdue</h2>
					<p>The following task is overdue and requires your attention:</p>
					<h3>%s</h3>
					<p><strong>Due date:</strong> %s</p>
					<p><a href="%s">View Task</a></p>
				`, task.Title, dueStr, taskURL),
				TextBody: fmt.Sprintf("Task Overdue: %s\nDue date: %s\nView: %s",
					task.Title, dueStr, taskURL),
			}

			if err := s.emailSender.Send(ctx, email); err != nil {
				s.logger.Error("failed to send email notification", "error", err)
				errors <- err
			}
		}()
	}

	// Slack notification
	if s.config.EnableSlack && s.slackNotifier != nil && task.DueDate != nil {
		wg.Add(1)
		go func() {
			defer wg.Done()
			msg := BuildTaskOverdueMessage(task.Title, *task.DueDate, taskURL)

			if err := s.slackNotifier.Send(ctx, msg); err != nil {
				s.logger.Error("failed to send Slack notification", "error", err)
				errors <- err
			}
		}()
	}

	wg.Wait()
	close(errors)

	for err := range errors {
		s.logger.Warn("notification delivery failed", "error", err)
	}

	return nil
}

// SendEmail sends an email notification
func (s *Service) SendEmail(ctx context.Context, to []string, subject string, htmlBody, textBody string) error {
	if !s.config.EnableEmail || s.emailSender == nil {
		s.logger.Debug("email disabled, skipping")
		return nil
	}

	email := Email{
		To:       to,
		Subject:  subject,
		HTMLBody: htmlBody,
		TextBody: textBody,
	}

	return s.emailSender.Send(ctx, email)
}

// SendSlackMessage sends a Slack message
func (s *Service) SendSlackMessage(ctx context.Context, message SlackMessage) error {
	if !s.config.EnableSlack || s.slackNotifier == nil {
		s.logger.Debug("Slack disabled, skipping")
		return nil
	}

	return s.slackNotifier.Send(ctx, message)
}

// CreateInAppNotification creates an in-app notification
func (s *Service) CreateInAppNotification(ctx context.Context, tenantID uuid.UUID, userID, title, message string, notifType NotificationType) error {
	if !s.config.EnableInApp || s.inAppService == nil {
		s.logger.Debug("in-app notifications disabled, skipping")
		return nil
	}

	notif := &InAppNotification{
		TenantID: tenantID,
		UserID:   userID,
		Title:    title,
		Message:  message,
		Type:     notifType,
	}

	return s.inAppService.Create(ctx, notif)
}

// NotifyTaskEscalated sends a notification when a task is escalated to backup approvers
func (s *Service) NotifyTaskEscalated(ctx context.Context, task *humantask.HumanTask, escalation *humantask.TaskEscalation) error {
	newAssignees := escalation.GetToAssignees()
	previousAssignees := escalation.GetFromAssignees()

	s.logger.Warn("task escalated notification",
		"task_id", task.ID,
		"escalation_level", escalation.EscalationLevel,
		"from_assignees", previousAssignees,
		"to_assignees", newAssignees,
	)

	if len(newAssignees) == 0 {
		return nil
	}

	var wg sync.WaitGroup
	errors := make(chan error, 3)

	taskURL := fmt.Sprintf("/tasks/%s", task.ID)
	dueStr := "Not set"
	if task.DueDate != nil {
		dueStr = task.DueDate.Format("2006-01-02 15:04")
	}

	// In-app notification to new assignees
	if s.config.EnableInApp && s.inAppService != nil {
		for _, assignee := range newAssignees {
			assignee := assignee
			wg.Add(1)
			go func() {
				defer wg.Done()
				title := fmt.Sprintf("Task Escalated: %s", task.Title)
				message := fmt.Sprintf("This task has been escalated to you (Level %d). Previous assignees did not respond.", escalation.EscalationLevel)
				if err := s.inAppService.Create(ctx, &InAppNotification{
					TenantID: task.TenantID,
					UserID:   assignee,
					Title:    title,
					Message:  message,
					Type:     NotificationTypeWarning,
					Metadata: map[string]any{
						"task_id":          task.ID.String(),
						"escalation_level": escalation.EscalationLevel,
						"task_url":         taskURL,
					},
				}); err != nil {
					s.logger.Error("failed to send in-app notification", "error", err, "assignee", assignee)
					errors <- err
				}
			}()
		}
	}

	// Email notification to new assignees
	if s.config.EnableEmail && s.emailSender != nil {
		wg.Add(1)
		go func() {
			defer wg.Done()
			email := Email{
				To:      newAssignees,
				Subject: fmt.Sprintf("Task Escalated: %s (Level %d)", task.Title, escalation.EscalationLevel),
				HTMLBody: fmt.Sprintf(`
					<h2>Task Escalated</h2>
					<p>A task has been escalated to you because the previous assignees did not respond in time.</p>
					<h3>%s</h3>
					<p><strong>Escalation Level:</strong> %d</p>
					<p><strong>Type:</strong> %s</p>
					<p><strong>New Due Date:</strong> %s</p>
					<p><a href="%s">View Task</a></p>
				`, task.Title, escalation.EscalationLevel, task.TaskType, dueStr, taskURL),
				TextBody: fmt.Sprintf("Task Escalated: %s\nLevel: %d\nType: %s\nDue: %s\nView: %s",
					task.Title, escalation.EscalationLevel, task.TaskType, dueStr, taskURL),
			}

			if err := s.emailSender.Send(ctx, email); err != nil {
				s.logger.Error("failed to send email notification", "error", err)
				errors <- err
			}
		}()
	}

	// Slack notification
	if s.config.EnableSlack && s.slackNotifier != nil {
		wg.Add(1)
		go func() {
			defer wg.Done()
			msg := BuildTaskEscalatedMessage(task.Title, escalation.EscalationLevel, dueStr, taskURL)

			if err := s.slackNotifier.Send(ctx, msg); err != nil {
				s.logger.Error("failed to send Slack notification", "error", err)
				errors <- err
			}
		}()
	}

	wg.Wait()
	close(errors)

	for err := range errors {
		s.logger.Warn("notification delivery failed", "error", err)
	}

	return nil
}

// NotifyWorkflowExecution sends notifications for workflow execution events
func (s *Service) NotifyWorkflowExecution(ctx context.Context, tenantID uuid.UUID, userID, workflowName, status, errorMsg, executionURL string) error {
	var wg sync.WaitGroup
	errors := make(chan error, 3)

	// In-app notification
	if s.config.EnableInApp && s.inAppService != nil {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := s.inAppService.NotifyWorkflowExecution(ctx, tenantID, userID, workflowName, status, errorMsg, executionURL); err != nil {
				s.logger.Error("failed to send in-app notification", "error", err)
				errors <- err
			}
		}()
	}

	// Slack notification (only for failures)
	if s.config.EnableSlack && s.slackNotifier != nil && status == "failed" {
		wg.Add(1)
		go func() {
			defer wg.Done()
			msg := BuildWorkflowExecutionMessage(workflowName, status, errorMsg, executionURL)

			if err := s.slackNotifier.Send(ctx, msg); err != nil {
				s.logger.Error("failed to send Slack notification", "error", err)
				errors <- err
			}
		}()
	}

	wg.Wait()
	close(errors)

	for err := range errors {
		s.logger.Warn("notification delivery failed", "error", err)
	}

	return nil
}

// NoOpNotificationService is a notification service that does nothing
type NoOpNotificationService struct{}

func (n *NoOpNotificationService) NotifyTaskAssigned(ctx context.Context, task *humantask.HumanTask) error {
	return nil
}

func (n *NoOpNotificationService) NotifyTaskCompleted(ctx context.Context, task *humantask.HumanTask) error {
	return nil
}

func (n *NoOpNotificationService) NotifyTaskOverdue(ctx context.Context, task *humantask.HumanTask) error {
	return nil
}

func (n *NoOpNotificationService) NotifyTaskEscalated(ctx context.Context, task *humantask.HumanTask, escalation *humantask.TaskEscalation) error {
	return nil
}

// NewNoOpService returns a no-op notification service for testing
func NewNoOpService() *NoOpNotificationService {
	return &NoOpNotificationService{}
}
