package notification

import (
	"context"
	"log/slog"

	"github.com/gorax/gorax/internal/humantask"
)

// Service implements notification delivery
type Service struct {
	logger *slog.Logger
	// Add email client, Slack client, etc. here in the future
}

// NewService creates a new notification service
func NewService(logger *slog.Logger) *Service {
	return &Service{
		logger: logger,
	}
}

// NotifyTaskAssigned sends a notification when a task is assigned
func (s *Service) NotifyTaskAssigned(ctx context.Context, task *humantask.HumanTask) error {
	s.logger.Info("task assigned notification",
		"task_id", task.ID,
		"task_type", task.TaskType,
		"title", task.Title,
	)

	// TODO: Implement actual notification delivery
	// - Send email to assignees
	// - Send Slack message
	// - Create in-app notification

	return nil
}

// NotifyTaskCompleted sends a notification when a task is completed
func (s *Service) NotifyTaskCompleted(ctx context.Context, task *humantask.HumanTask) error {
	s.logger.Info("task completed notification",
		"task_id", task.ID,
		"status", task.Status,
		"completed_by", task.CompletedBy,
	)

	// TODO: Implement actual notification delivery

	return nil
}

// NotifyTaskOverdue sends a notification when a task is overdue
func (s *Service) NotifyTaskOverdue(ctx context.Context, task *humantask.HumanTask) error {
	s.logger.Warn("task overdue notification",
		"task_id", task.ID,
		"due_date", task.DueDate,
	)

	// TODO: Implement actual notification delivery

	return nil
}

// SendEmail sends an email notification (placeholder)
func (s *Service) SendEmail(to []string, subject string, body string) error {
	s.logger.Info("sending email",
		"to", to,
		"subject", subject,
	)

	// TODO: Implement actual email sending using SMTP or email service

	return nil
}

// SendSlackMessage sends a Slack message (placeholder)
func (s *Service) SendSlackMessage(channel string, message string) error {
	s.logger.Info("sending slack message",
		"channel", channel,
		"message", message,
	)

	// TODO: Implement actual Slack integration

	return nil
}

// CreateInAppNotification creates an in-app notification (placeholder)
func (s *Service) CreateInAppNotification(userID string, title string, message string) error {
	s.logger.Info("creating in-app notification",
		"user_id", userID,
		"title", title,
	)

	// TODO: Store in notifications table for in-app display

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

// NewNoOpService returns a no-op notification service for testing
func NewNoOpService() *NoOpNotificationService {
	return &NoOpNotificationService{}
}
