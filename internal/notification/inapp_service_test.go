package notification

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockInAppRepository is a mock implementation of the repository
type MockInAppRepository struct {
	mock.Mock
}

func (m *MockInAppRepository) Create(ctx context.Context, notif *InAppNotification) error {
	args := m.Called(ctx, notif)
	return args.Error(0)
}

func (m *MockInAppRepository) GetByID(ctx context.Context, id uuid.UUID) (*InAppNotification, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*InAppNotification), args.Error(1)
}

func (m *MockInAppRepository) ListByUser(ctx context.Context, userID string, limit, offset int) ([]*InAppNotification, error) {
	args := m.Called(ctx, userID, limit, offset)
	return args.Get(0).([]*InAppNotification), args.Error(1)
}

func (m *MockInAppRepository) ListUnread(ctx context.Context, userID string, limit, offset int) ([]*InAppNotification, error) {
	args := m.Called(ctx, userID, limit, offset)
	return args.Get(0).([]*InAppNotification), args.Error(1)
}

func (m *MockInAppRepository) CountUnread(ctx context.Context, userID string) (int, error) {
	args := m.Called(ctx, userID)
	return args.Int(0), args.Error(1)
}

func (m *MockInAppRepository) MarkAsRead(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockInAppRepository) MarkAllAsRead(ctx context.Context, userID string) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *MockInAppRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockInAppRepository) BulkCreate(ctx context.Context, notifications []*InAppNotification) error {
	args := m.Called(ctx, notifications)
	return args.Error(0)
}

// MockWebSocketHub is a mock WebSocket hub
type MockWebSocketHub struct {
	mock.Mock
	broadcasts []BroadcastData
}

type BroadcastData struct {
	Room    string
	Message []byte
}

func (m *MockWebSocketHub) BroadcastToRoom(room string, message []byte) {
	m.broadcasts = append(m.broadcasts, BroadcastData{Room: room, Message: message})
	m.Called(room, message)
}

func TestInAppService_Create(t *testing.T) {
	mockRepo := new(MockInAppRepository)
	mockHub := new(MockWebSocketHub)

	service := NewInAppService(mockRepo, mockHub)

	tenantID := uuid.New()
	userID := "user-123"

	notif := &InAppNotification{
		TenantID: tenantID,
		UserID:   userID,
		Title:    "Test Notification",
		Message:  "This is a test",
		Type:     NotificationTypeInfo,
	}

	mockRepo.On("Create", mock.Anything, notif).Return(nil)
	mockHub.On("BroadcastToRoom", mock.Anything, mock.Anything).Return()

	err := service.Create(context.Background(), notif)
	require.NoError(t, err)

	mockRepo.AssertExpectations(t)
	mockHub.AssertExpectations(t)

	// Verify broadcast was called with correct room
	assert.Len(t, mockHub.broadcasts, 1)
	assert.Equal(t, "notifications:user-123", mockHub.broadcasts[0].Room)
}

func TestInAppService_GetByID(t *testing.T) {
	mockRepo := new(MockInAppRepository)
	mockHub := new(MockWebSocketHub)

	service := NewInAppService(mockRepo, mockHub)

	notifID := uuid.New()
	expectedNotif := &InAppNotification{
		ID:      notifID,
		Title:   "Test",
		Message: "Test message",
	}

	mockRepo.On("GetByID", mock.Anything, notifID).Return(expectedNotif, nil)

	notif, err := service.GetByID(context.Background(), notifID)
	require.NoError(t, err)
	assert.Equal(t, expectedNotif, notif)

	mockRepo.AssertExpectations(t)
}

func TestInAppService_ListByUser(t *testing.T) {
	mockRepo := new(MockInAppRepository)
	mockHub := new(MockWebSocketHub)

	service := NewInAppService(mockRepo, mockHub)

	userID := "user-123"
	expectedNotifs := []*InAppNotification{
		{ID: uuid.New(), Title: "Notif 1"},
		{ID: uuid.New(), Title: "Notif 2"},
	}

	mockRepo.On("ListByUser", mock.Anything, userID, 10, 0).Return(expectedNotifs, nil)

	notifs, err := service.ListByUser(context.Background(), userID, 10, 0)
	require.NoError(t, err)
	assert.Len(t, notifs, 2)

	mockRepo.AssertExpectations(t)
}

func TestInAppService_ListUnread(t *testing.T) {
	mockRepo := new(MockInAppRepository)
	mockHub := new(MockWebSocketHub)

	service := NewInAppService(mockRepo, mockHub)

	userID := "user-123"
	expectedNotifs := []*InAppNotification{
		{ID: uuid.New(), Title: "Unread 1", IsRead: false},
	}

	mockRepo.On("ListUnread", mock.Anything, userID, 10, 0).Return(expectedNotifs, nil)

	notifs, err := service.ListUnread(context.Background(), userID, 10, 0)
	require.NoError(t, err)
	assert.Len(t, notifs, 1)
	assert.False(t, notifs[0].IsRead)

	mockRepo.AssertExpectations(t)
}

func TestInAppService_CountUnread(t *testing.T) {
	mockRepo := new(MockInAppRepository)
	mockHub := new(MockWebSocketHub)

	service := NewInAppService(mockRepo, mockHub)

	userID := "user-123"

	mockRepo.On("CountUnread", mock.Anything, userID).Return(5, nil)

	count, err := service.CountUnread(context.Background(), userID)
	require.NoError(t, err)
	assert.Equal(t, 5, count)

	mockRepo.AssertExpectations(t)
}

func TestInAppService_MarkAsRead(t *testing.T) {
	mockRepo := new(MockInAppRepository)
	mockHub := new(MockWebSocketHub)

	service := NewInAppService(mockRepo, mockHub)

	notifID := uuid.New()
	userID := "user-123"

	mockRepo.On("MarkAsRead", mock.Anything, notifID).Return(nil)
	mockHub.On("BroadcastToRoom", mock.Anything, mock.Anything).Return()

	err := service.MarkAsRead(context.Background(), notifID, userID)
	require.NoError(t, err)

	mockRepo.AssertExpectations(t)
	mockHub.AssertExpectations(t)
}

func TestInAppService_MarkAllAsRead(t *testing.T) {
	mockRepo := new(MockInAppRepository)
	mockHub := new(MockWebSocketHub)

	service := NewInAppService(mockRepo, mockHub)

	userID := "user-123"

	mockRepo.On("MarkAllAsRead", mock.Anything, userID).Return(nil)
	mockHub.On("BroadcastToRoom", mock.Anything, mock.Anything).Return()

	err := service.MarkAllAsRead(context.Background(), userID)
	require.NoError(t, err)

	mockRepo.AssertExpectations(t)
	mockHub.AssertExpectations(t)
}

func TestInAppService_Delete(t *testing.T) {
	mockRepo := new(MockInAppRepository)
	mockHub := new(MockWebSocketHub)

	service := NewInAppService(mockRepo, mockHub)

	notifID := uuid.New()

	mockRepo.On("Delete", mock.Anything, notifID).Return(nil)

	err := service.Delete(context.Background(), notifID)
	require.NoError(t, err)

	mockRepo.AssertExpectations(t)
}

func TestInAppService_NotifyTaskAssigned(t *testing.T) {
	mockRepo := new(MockInAppRepository)
	mockHub := new(MockWebSocketHub)

	service := NewInAppService(mockRepo, mockHub)

	tenantID := uuid.New()
	userID := "user-123"
	taskTitle := "Review PR"
	taskURL := "https://app.example.com/tasks/123"

	mockRepo.On("Create", mock.Anything, mock.MatchedBy(func(n *InAppNotification) bool {
		return n.TenantID == tenantID &&
			n.UserID == userID &&
			n.Type == NotificationTypeInfo &&
			n.Title == "Task Assigned" &&
			n.Link == taskURL
	})).Return(nil)

	mockHub.On("BroadcastToRoom", mock.Anything, mock.Anything).Return()

	err := service.NotifyTaskAssigned(context.Background(), tenantID, userID, taskTitle, taskURL)
	require.NoError(t, err)

	mockRepo.AssertExpectations(t)
	mockHub.AssertExpectations(t)
}

func TestInAppService_NotifyTaskCompleted(t *testing.T) {
	mockRepo := new(MockInAppRepository)
	mockHub := new(MockWebSocketHub)

	service := NewInAppService(mockRepo, mockHub)

	tenantID := uuid.New()
	userID := "user-123"
	taskTitle := "Review PR"
	status := "approved"

	mockRepo.On("Create", mock.Anything, mock.MatchedBy(func(n *InAppNotification) bool {
		return n.TenantID == tenantID &&
			n.UserID == userID &&
			n.Type == NotificationTypeSuccess &&
			n.Title == "Task Completed"
	})).Return(nil)

	mockHub.On("BroadcastToRoom", mock.Anything, mock.Anything).Return()

	err := service.NotifyTaskCompleted(context.Background(), tenantID, userID, taskTitle, status)
	require.NoError(t, err)

	mockRepo.AssertExpectations(t)
	mockHub.AssertExpectations(t)
}

func TestInAppService_NotifyTaskOverdue(t *testing.T) {
	mockRepo := new(MockInAppRepository)
	mockHub := new(MockWebSocketHub)

	service := NewInAppService(mockRepo, mockHub)

	tenantID := uuid.New()
	userID := "user-123"
	taskTitle := "Sign contract"
	taskURL := "https://app.example.com/tasks/456"

	mockRepo.On("Create", mock.Anything, mock.MatchedBy(func(n *InAppNotification) bool {
		return n.TenantID == tenantID &&
			n.UserID == userID &&
			n.Type == NotificationTypeWarning &&
			n.Title == "Task Overdue"
	})).Return(nil)

	mockHub.On("BroadcastToRoom", mock.Anything, mock.Anything).Return()

	err := service.NotifyTaskOverdue(context.Background(), tenantID, userID, taskTitle, taskURL)
	require.NoError(t, err)

	mockRepo.AssertExpectations(t)
	mockHub.AssertExpectations(t)
}

func TestInAppService_NotifyWorkflowExecution(t *testing.T) {
	mockRepo := new(MockInAppRepository)
	mockHub := new(MockWebSocketHub)

	service := NewInAppService(mockRepo, mockHub)

	tenantID := uuid.New()
	userID := "user-123"
	workflowName := "Data Pipeline"
	status := "failed"
	errorMsg := "Connection timeout"
	executionURL := "https://app.example.com/executions/789"

	mockRepo.On("Create", mock.Anything, mock.MatchedBy(func(n *InAppNotification) bool {
		return n.TenantID == tenantID &&
			n.UserID == userID &&
			n.Type == NotificationTypeError &&
			n.Title == "Workflow Failed"
	})).Return(nil)

	mockHub.On("BroadcastToRoom", mock.Anything, mock.Anything).Return()

	err := service.NotifyWorkflowExecution(context.Background(), tenantID, userID, workflowName, status, errorMsg, executionURL)
	require.NoError(t, err)

	mockRepo.AssertExpectations(t)
	mockHub.AssertExpectations(t)
}
