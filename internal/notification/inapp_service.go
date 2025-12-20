package notification

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
)

// WebSocketBroadcaster defines the interface for WebSocket broadcasting
type WebSocketBroadcaster interface {
	BroadcastToRoom(room string, message []byte)
}

// InAppRepositoryInterface defines the interface for in-app notification repository
type InAppRepositoryInterface interface {
	Create(ctx context.Context, notif *InAppNotification) error
	GetByID(ctx context.Context, id uuid.UUID) (*InAppNotification, error)
	ListByUser(ctx context.Context, userID string, limit, offset int) ([]*InAppNotification, error)
	ListUnread(ctx context.Context, userID string, limit, offset int) ([]*InAppNotification, error)
	CountUnread(ctx context.Context, userID string) (int, error)
	MarkAsRead(ctx context.Context, id uuid.UUID) error
	MarkAllAsRead(ctx context.Context, userID string) error
	Delete(ctx context.Context, id uuid.UUID) error
	BulkCreate(ctx context.Context, notifications []*InAppNotification) error
}

// InAppService handles in-app notification operations
type InAppService struct {
	repo InAppRepositoryInterface
	hub  WebSocketBroadcaster
}

// NewInAppService creates a new in-app notification service
func NewInAppService(repo InAppRepositoryInterface, hub WebSocketBroadcaster) *InAppService {
	return &InAppService{
		repo: repo,
		hub:  hub,
	}
}

// Create creates a new in-app notification and broadcasts it via WebSocket
func (s *InAppService) Create(ctx context.Context, notif *InAppNotification) error {
	// Create notification in database
	if err := s.repo.Create(ctx, notif); err != nil {
		return fmt.Errorf("failed to create notification: %w", err)
	}

	// Broadcast to user's WebSocket room
	s.broadcastNotification(notif)

	return nil
}

// GetByID retrieves a notification by ID
func (s *InAppService) GetByID(ctx context.Context, id uuid.UUID) (*InAppNotification, error) {
	return s.repo.GetByID(ctx, id)
}

// ListByUser lists notifications for a user
func (s *InAppService) ListByUser(ctx context.Context, userID string, limit, offset int) ([]*InAppNotification, error) {
	return s.repo.ListByUser(ctx, userID, limit, offset)
}

// ListUnread lists unread notifications for a user
func (s *InAppService) ListUnread(ctx context.Context, userID string, limit, offset int) ([]*InAppNotification, error) {
	return s.repo.ListUnread(ctx, userID, limit, offset)
}

// CountUnread counts unread notifications for a user
func (s *InAppService) CountUnread(ctx context.Context, userID string) (int, error) {
	return s.repo.CountUnread(ctx, userID)
}

// MarkAsRead marks a notification as read and broadcasts the update
func (s *InAppService) MarkAsRead(ctx context.Context, id uuid.UUID, userID string) error {
	if err := s.repo.MarkAsRead(ctx, id); err != nil {
		return fmt.Errorf("failed to mark notification as read: %w", err)
	}

	// Broadcast read status update
	s.broadcastReadUpdate(userID, id)

	return nil
}

// MarkAllAsRead marks all notifications as read for a user
func (s *InAppService) MarkAllAsRead(ctx context.Context, userID string) error {
	if err := s.repo.MarkAllAsRead(ctx, userID); err != nil {
		return fmt.Errorf("failed to mark all notifications as read: %w", err)
	}

	// Broadcast bulk read update
	s.broadcastBulkReadUpdate(userID)

	return nil
}

// Delete deletes a notification
func (s *InAppService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}

// NotifyTaskAssigned creates a task assigned notification
func (s *InAppService) NotifyTaskAssigned(ctx context.Context, tenantID uuid.UUID, userID, taskTitle, taskURL string) error {
	notif := &InAppNotification{
		TenantID: tenantID,
		UserID:   userID,
		Title:    "Task Assigned",
		Message:  fmt.Sprintf("You have been assigned a new task: %s", taskTitle),
		Type:     NotificationTypeInfo,
		Link:     taskURL,
		Metadata: map[string]interface{}{
			"task_title": taskTitle,
			"event_type": "task_assigned",
		},
	}

	return s.Create(ctx, notif)
}

// NotifyTaskCompleted creates a task completed notification
func (s *InAppService) NotifyTaskCompleted(ctx context.Context, tenantID uuid.UUID, userID, taskTitle, status string) error {
	notifType := NotificationTypeSuccess
	if status == "rejected" {
		notifType = NotificationTypeWarning
	}

	notif := &InAppNotification{
		TenantID: tenantID,
		UserID:   userID,
		Title:    "Task Completed",
		Message:  fmt.Sprintf("Task '%s' has been completed with status: %s", taskTitle, status),
		Type:     notifType,
		Metadata: map[string]interface{}{
			"task_title": taskTitle,
			"status":     status,
			"event_type": "task_completed",
		},
	}

	return s.Create(ctx, notif)
}

// NotifyTaskOverdue creates a task overdue notification
func (s *InAppService) NotifyTaskOverdue(ctx context.Context, tenantID uuid.UUID, userID, taskTitle, taskURL string) error {
	notif := &InAppNotification{
		TenantID: tenantID,
		UserID:   userID,
		Title:    "Task Overdue",
		Message:  fmt.Sprintf("Task '%s' is overdue and requires your attention", taskTitle),
		Type:     NotificationTypeWarning,
		Link:     taskURL,
		Metadata: map[string]interface{}{
			"task_title": taskTitle,
			"event_type": "task_overdue",
		},
	}

	return s.Create(ctx, notif)
}

// NotifyWorkflowExecution creates a workflow execution notification
func (s *InAppService) NotifyWorkflowExecution(ctx context.Context, tenantID uuid.UUID, userID, workflowName, status, errorMsg, executionURL string) error {
	var (
		title   string
		message string
		notifType NotificationType
	)

	switch status {
	case "completed", "success":
		title = "Workflow Completed"
		message = fmt.Sprintf("Workflow '%s' completed successfully", workflowName)
		notifType = NotificationTypeSuccess
	case "failed":
		title = "Workflow Failed"
		message = fmt.Sprintf("Workflow '%s' failed", workflowName)
		if errorMsg != "" {
			message += fmt.Sprintf(": %s", errorMsg)
		}
		notifType = NotificationTypeError
	case "running":
		title = "Workflow Running"
		message = fmt.Sprintf("Workflow '%s' is now running", workflowName)
		notifType = NotificationTypeInfo
	default:
		title = "Workflow Status Update"
		message = fmt.Sprintf("Workflow '%s' status: %s", workflowName, status)
		notifType = NotificationTypeInfo
	}

	metadata := map[string]interface{}{
		"workflow_name": workflowName,
		"status":        status,
		"event_type":    "workflow_execution",
	}

	if errorMsg != "" {
		metadata["error"] = errorMsg
	}

	notif := &InAppNotification{
		TenantID: tenantID,
		UserID:   userID,
		Title:    title,
		Message:  message,
		Type:     notifType,
		Link:     executionURL,
		Metadata: metadata,
	}

	return s.Create(ctx, notif)
}

// broadcastNotification sends a notification to the user via WebSocket
func (s *InAppService) broadcastNotification(notif *InAppNotification) {
	if s.hub == nil {
		return
	}

	// Build room name for user
	room := fmt.Sprintf("notifications:%s", notif.UserID)

	// Marshal notification to JSON
	data, err := json.Marshal(map[string]interface{}{
		"type":         "notification",
		"action":       "created",
		"notification": notif,
	})

	if err != nil {
		// Log error but don't fail the operation
		return
	}

	// Broadcast to room
	s.hub.BroadcastToRoom(room, data)
}

// broadcastReadUpdate broadcasts a read status update
func (s *InAppService) broadcastReadUpdate(userID string, notifID uuid.UUID) {
	if s.hub == nil {
		return
	}

	room := fmt.Sprintf("notifications:%s", userID)

	data, err := json.Marshal(map[string]interface{}{
		"type":            "notification",
		"action":          "read",
		"notification_id": notifID,
	})

	if err != nil {
		return
	}

	s.hub.BroadcastToRoom(room, data)
}

// broadcastBulkReadUpdate broadcasts a bulk read status update
func (s *InAppService) broadcastBulkReadUpdate(userID string) {
	if s.hub == nil {
		return
	}

	room := fmt.Sprintf("notifications:%s", userID)

	data, err := json.Marshal(map[string]interface{}{
		"type":   "notification",
		"action": "read_all",
	})

	if err != nil {
		return
	}

	s.hub.BroadcastToRoom(room, data)
}

// CreateBulk creates multiple notifications in bulk
func (s *InAppService) CreateBulk(ctx context.Context, notifications []*InAppNotification) error {
	if err := s.repo.BulkCreate(ctx, notifications); err != nil {
		return fmt.Errorf("failed to create bulk notifications: %w", err)
	}

	// Broadcast each notification
	for _, notif := range notifications {
		s.broadcastNotification(notif)
	}

	return nil
}
