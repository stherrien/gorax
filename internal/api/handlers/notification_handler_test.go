package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/gorax/gorax/internal/notification"
)

// MockInAppNotificationService is a mock implementation
type MockInAppNotificationService struct {
	mock.Mock
}

func (m *MockInAppNotificationService) ListByUser(ctx interface{}, userID string, limit, offset int) ([]*notification.InAppNotification, error) {
	args := m.Called(ctx, userID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*notification.InAppNotification), args.Error(1)
}

func (m *MockInAppNotificationService) ListUnread(ctx interface{}, userID string, limit, offset int) ([]*notification.InAppNotification, error) {
	args := m.Called(ctx, userID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*notification.InAppNotification), args.Error(1)
}

func (m *MockInAppNotificationService) CountUnread(ctx interface{}, userID string) (int, error) {
	args := m.Called(ctx, userID)
	return args.Int(0), args.Error(1)
}

func (m *MockInAppNotificationService) GetByID(ctx interface{}, id uuid.UUID) (*notification.InAppNotification, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*notification.InAppNotification), args.Error(1)
}

func (m *MockInAppNotificationService) MarkAsRead(ctx interface{}, id uuid.UUID, userID string) error {
	args := m.Called(ctx, id, userID)
	return args.Error(0)
}

func (m *MockInAppNotificationService) MarkAllAsRead(ctx interface{}, userID string) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *MockInAppNotificationService) Delete(ctx interface{}, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func createTestNotification() *notification.InAppNotification {
	return &notification.InAppNotification{
		ID:       uuid.New(),
		TenantID: uuid.New(),
		UserID:   "user-123",
		Title:    "Test Notification",
		Message:  "This is a test notification",
		Type:     notification.NotificationTypeInfo,
		IsRead:   false,
	}
}

func TestNotificationHandler_ListNotifications(t *testing.T) {
	gin.SetMode(gin.TestMode)

	svc := new(MockInAppNotificationService)
	handler := NewNotificationHandler(svc)

	router := gin.New()
	router.GET("/api/v1/notifications", handler.ListNotifications)

	notifications := []*notification.InAppNotification{
		createTestNotification(),
		createTestNotification(),
	}

	svc.On("ListByUser", mock.Anything, "user-123", 20, 0).Return(notifications, nil)

	req, _ := http.NewRequest("GET", "/api/v1/notifications", nil)
	req.Header.Set("X-User-ID", "user-123")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response []notification.InAppNotification
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Len(t, response, 2)

	svc.AssertExpectations(t)
}

func TestNotificationHandler_ListUnreadNotifications(t *testing.T) {
	gin.SetMode(gin.TestMode)

	svc := new(MockInAppNotificationService)
	handler := NewNotificationHandler(svc)

	router := gin.New()
	router.GET("/api/v1/notifications/unread", handler.ListUnreadNotifications)

	notifications := []*notification.InAppNotification{
		createTestNotification(),
	}

	svc.On("ListUnread", mock.Anything, "user-123", 20, 0).Return(notifications, nil)

	req, _ := http.NewRequest("GET", "/api/v1/notifications/unread", nil)
	req.Header.Set("X-User-ID", "user-123")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response []notification.InAppNotification
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Len(t, response, 1)

	svc.AssertExpectations(t)
}

func TestNotificationHandler_GetUnreadCount(t *testing.T) {
	gin.SetMode(gin.TestMode)

	svc := new(MockInAppNotificationService)
	handler := NewNotificationHandler(svc)

	router := gin.New()
	router.GET("/api/v1/notifications/count", handler.GetUnreadCount)

	svc.On("CountUnread", mock.Anything, "user-123").Return(5, nil)

	req, _ := http.NewRequest("GET", "/api/v1/notifications/count", nil)
	req.Header.Set("X-User-ID", "user-123")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]int
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, 5, response["count"])

	svc.AssertExpectations(t)
}

func TestNotificationHandler_MarkAsRead(t *testing.T) {
	gin.SetMode(gin.TestMode)

	svc := new(MockInAppNotificationService)
	handler := NewNotificationHandler(svc)

	router := gin.New()
	router.PATCH("/api/v1/notifications/:id/read", handler.MarkAsRead)

	notifID := uuid.New()
	svc.On("MarkAsRead", mock.Anything, notifID, "user-123").Return(nil)

	req, _ := http.NewRequest("PATCH", "/api/v1/notifications/"+notifID.String()+"/read", nil)
	req.Header.Set("X-User-ID", "user-123")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	svc.AssertExpectations(t)
}

func TestNotificationHandler_MarkAllAsRead(t *testing.T) {
	gin.SetMode(gin.TestMode)

	svc := new(MockInAppNotificationService)
	handler := NewNotificationHandler(svc)

	router := gin.New()
	router.POST("/api/v1/notifications/read-all", handler.MarkAllAsRead)

	svc.On("MarkAllAsRead", mock.Anything, "user-123").Return(nil)

	req, _ := http.NewRequest("POST", "/api/v1/notifications/read-all", nil)
	req.Header.Set("X-User-ID", "user-123")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	svc.AssertExpectations(t)
}

func TestNotificationHandler_DeleteNotification(t *testing.T) {
	gin.SetMode(gin.TestMode)

	svc := new(MockInAppNotificationService)
	handler := NewNotificationHandler(svc)

	router := gin.New()
	router.DELETE("/api/v1/notifications/:id", handler.DeleteNotification)

	notifID := uuid.New()
	// Handler calls GetByID first to verify ownership
	svc.On("GetByID", mock.Anything, notifID).Return(&notification.InAppNotification{
		ID:     notifID,
		UserID: "user-123",
	}, nil)
	svc.On("Delete", mock.Anything, notifID).Return(nil)

	req, _ := http.NewRequest("DELETE", "/api/v1/notifications/"+notifID.String(), nil)
	req.Header.Set("X-User-ID", "user-123")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	svc.AssertExpectations(t)
}

func TestNotificationHandler_MissingUserID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	svc := new(MockInAppNotificationService)
	handler := NewNotificationHandler(svc)

	router := gin.New()
	router.GET("/api/v1/notifications", handler.ListNotifications)

	req, _ := http.NewRequest("GET", "/api/v1/notifications", nil)
	// No X-User-ID header

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestNotificationHandler_InvalidUUID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	svc := new(MockInAppNotificationService)
	handler := NewNotificationHandler(svc)

	router := gin.New()
	router.PATCH("/api/v1/notifications/:id/read", handler.MarkAsRead)

	req, _ := http.NewRequest("PATCH", "/api/v1/notifications/invalid-uuid/read", nil)
	req.Header.Set("X-User-ID", "user-123")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}
