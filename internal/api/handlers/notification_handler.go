package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorax/gorax/internal/notification"
)

// InAppNotificationService defines the interface for in-app notification operations
type InAppNotificationService interface {
	ListByUser(ctx interface{}, userID string, limit, offset int) ([]*notification.InAppNotification, error)
	ListUnread(ctx interface{}, userID string, limit, offset int) ([]*notification.InAppNotification, error)
	CountUnread(ctx interface{}, userID string) (int, error)
	GetByID(ctx interface{}, id uuid.UUID) (*notification.InAppNotification, error)
	MarkAsRead(ctx interface{}, id uuid.UUID, userID string) error
	MarkAllAsRead(ctx interface{}, userID string) error
	Delete(ctx interface{}, id uuid.UUID) error
}

// NotificationHandler handles HTTP requests for notifications
type NotificationHandler struct {
	service InAppNotificationService
}

// NewNotificationHandler creates a new notification handler
func NewNotificationHandler(service InAppNotificationService) *NotificationHandler {
	return &NotificationHandler{
		service: service,
	}
}

// ListNotifications lists all notifications for the current user
// GET /api/v1/notifications
func (h *NotificationHandler) ListNotifications(c *gin.Context) {
	userID := c.GetHeader("X-User-ID")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	// Parse pagination parameters
	limit := 20
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	offset := 0
	if offsetStr := c.Query("offset"); offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	notifications, err := h.service.ListByUser(c.Request.Context(), userID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve notifications"})
		return
	}

	c.JSON(http.StatusOK, notifications)
}

// ListUnreadNotifications lists unread notifications for the current user
// GET /api/v1/notifications/unread
func (h *NotificationHandler) ListUnreadNotifications(c *gin.Context) {
	userID := c.GetHeader("X-User-ID")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	// Parse pagination parameters
	limit := 20
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	offset := 0
	if offsetStr := c.Query("offset"); offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	notifications, err := h.service.ListUnread(c.Request.Context(), userID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve notifications"})
		return
	}

	c.JSON(http.StatusOK, notifications)
}

// GetUnreadCount returns the count of unread notifications
// GET /api/v1/notifications/count
func (h *NotificationHandler) GetUnreadCount(c *gin.Context) {
	userID := c.GetHeader("X-User-ID")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	count, err := h.service.CountUnread(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to count notifications"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"count": count})
}

// GetNotification retrieves a single notification by ID
// GET /api/v1/notifications/:id
func (h *NotificationHandler) GetNotification(c *gin.Context) {
	userID := c.GetHeader("X-User-ID")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid notification ID"})
		return
	}

	notification, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "notification not found"})
		return
	}

	// Verify the notification belongs to the user
	if notification.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	c.JSON(http.StatusOK, notification)
}

// MarkAsRead marks a notification as read
// PATCH /api/v1/notifications/:id/read
func (h *NotificationHandler) MarkAsRead(c *gin.Context) {
	userID := c.GetHeader("X-User-ID")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid notification ID"})
		return
	}

	err = h.service.MarkAsRead(c.Request.Context(), id, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to mark notification as read"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "notification marked as read"})
}

// MarkAllAsRead marks all notifications as read for the current user
// POST /api/v1/notifications/read-all
func (h *NotificationHandler) MarkAllAsRead(c *gin.Context) {
	userID := c.GetHeader("X-User-ID")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	err := h.service.MarkAllAsRead(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to mark notifications as read"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "all notifications marked as read"})
}

// DeleteNotification deletes a notification
// DELETE /api/v1/notifications/:id
func (h *NotificationHandler) DeleteNotification(c *gin.Context) {
	userID := c.GetHeader("X-User-ID")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid notification ID"})
		return
	}

	// Verify ownership before deletion
	notification, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "notification not found"})
		return
	}

	if notification.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	err = h.service.Delete(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete notification"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "notification deleted"})
}

// RegisterRoutes registers notification routes to the router
func (h *NotificationHandler) RegisterRoutes(router *gin.RouterGroup) {
	notifications := router.Group("/notifications")
	{
		notifications.GET("", h.ListNotifications)
		notifications.GET("/unread", h.ListUnreadNotifications)
		notifications.GET("/count", h.GetUnreadCount)
		notifications.GET("/:id", h.GetNotification)
		notifications.PATCH("/:id/read", h.MarkAsRead)
		notifications.POST("/read-all", h.MarkAllAsRead)
		notifications.DELETE("/:id", h.DeleteNotification)
	}
}
