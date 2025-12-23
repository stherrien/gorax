package humantask

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockRepository is a mock implementation of Repository
type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) Create(ctx context.Context, task *HumanTask) error {
	args := m.Called(ctx, task)
	if args.Get(0) != nil {
		return args.Error(0)
	}
	task.ID = uuid.New()
	task.CreatedAt = time.Now()
	task.UpdatedAt = time.Now()
	return nil
}

func (m *MockRepository) GetByID(ctx context.Context, id uuid.UUID) (*HumanTask, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*HumanTask), args.Error(1)
}

func (m *MockRepository) List(ctx context.Context, filter TaskFilter) ([]*HumanTask, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*HumanTask), args.Error(1)
}

func (m *MockRepository) Update(ctx context.Context, task *HumanTask) error {
	args := m.Called(ctx, task)
	return args.Error(0)
}

func (m *MockRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockRepository) GetOverdueTasks(ctx context.Context, tenantID uuid.UUID) ([]*HumanTask, error) {
	args := m.Called(ctx, tenantID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*HumanTask), args.Error(1)
}

func (m *MockRepository) CountPendingByAssignee(ctx context.Context, tenantID uuid.UUID, assignee string) (int, error) {
	args := m.Called(ctx, tenantID, assignee)
	return args.Int(0), args.Error(1)
}

// MockNotificationService is a mock implementation of NotificationService
type MockNotificationService struct {
	mock.Mock
}

func (m *MockNotificationService) NotifyTaskAssigned(ctx context.Context, task *HumanTask) error {
	args := m.Called(ctx, task)
	return args.Error(0)
}

func (m *MockNotificationService) NotifyTaskCompleted(ctx context.Context, task *HumanTask) error {
	args := m.Called(ctx, task)
	return args.Error(0)
}

func (m *MockNotificationService) NotifyTaskOverdue(ctx context.Context, task *HumanTask) error {
	args := m.Called(ctx, task)
	return args.Error(0)
}

func TestService_CreateTask(t *testing.T) {
	tests := []struct {
		name    string
		request CreateTaskRequest
		setup   func(*MockRepository, *MockNotificationService)
		wantErr error
	}{
		{
			name: "valid approval task",
			request: CreateTaskRequest{
				ExecutionID: uuid.New(),
				StepID:      "step-1",
				TaskType:    TaskTypeApproval,
				Title:       "Approve deployment",
				Description: "Please approve",
				Assignees:   []string{uuid.New().String(), "admin"},
			},
			setup: func(repo *MockRepository, notif *MockNotificationService) {
				repo.On("Create", mock.Anything, mock.Anything).Return(nil)
				// Notification is sent async in a goroutine, use Maybe() to avoid race conditions
				notif.On("NotifyTaskAssigned", mock.Anything, mock.Anything).Return(nil).Maybe()
			},
			wantErr: nil,
		},
		{
			name: "invalid task type",
			request: CreateTaskRequest{
				ExecutionID: uuid.New(),
				StepID:      "step-1",
				TaskType:    "invalid",
				Title:       "Test",
				Assignees:   []string{uuid.New().String()},
			},
			setup:   func(repo *MockRepository, notif *MockNotificationService) {},
			wantErr: ErrInvalidTaskType,
		},
		{
			name: "missing title",
			request: CreateTaskRequest{
				ExecutionID: uuid.New(),
				StepID:      "step-1",
				TaskType:    TaskTypeApproval,
				Assignees:   []string{uuid.New().String()},
			},
			setup:   func(repo *MockRepository, notif *MockNotificationService) {},
			wantErr: ErrMissingRequiredField,
		},
		{
			name: "empty assignees",
			request: CreateTaskRequest{
				ExecutionID: uuid.New(),
				StepID:      "step-1",
				TaskType:    TaskTypeApproval,
				Title:       "Test",
				Assignees:   []string{},
			},
			setup:   func(repo *MockRepository, notif *MockNotificationService) {},
			wantErr: ErrMissingRequiredField,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := new(MockRepository)
			notif := new(MockNotificationService)
			tt.setup(repo, notif)

			service := NewService(repo, notif)
			ctx := context.Background()
			tenantID := uuid.New()

			task, err := service.CreateTask(ctx, tenantID, tt.request)

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				assert.Nil(t, task)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, task)
				assert.NotEqual(t, uuid.Nil, task.ID)
				assert.Equal(t, tenantID, task.TenantID)
				assert.Equal(t, StatusPending, task.Status)
			}

			repo.AssertExpectations(t)
			notif.AssertExpectations(t)
		})
	}
}

func TestService_GetTask(t *testing.T) {
	tenantID := uuid.New()

	tests := []struct {
		name     string
		taskID   uuid.UUID
		tenantID uuid.UUID
		setup    func(*MockRepository) *HumanTask
		wantErr  error
	}{
		{
			name:     "existing task",
			taskID:   uuid.New(),
			tenantID: tenantID,
			setup: func(repo *MockRepository) *HumanTask {
				task := createMockTask()
				task.TenantID = tenantID // Same tenant ID
				repo.On("GetByID", mock.Anything, mock.Anything).Return(task, nil)
				return task
			},
			wantErr: nil,
		},
		{
			name:     "non-existent task",
			taskID:   uuid.New(),
			tenantID: tenantID,
			setup: func(repo *MockRepository) *HumanTask {
				repo.On("GetByID", mock.Anything, mock.Anything).Return(nil, ErrTaskNotFound)
				return nil
			},
			wantErr: ErrTaskNotFound,
		},
		{
			name:     "wrong tenant",
			taskID:   uuid.New(),
			tenantID: uuid.New(), // Different tenant ID
			setup: func(repo *MockRepository) *HumanTask {
				task := createMockTask()
				task.TenantID = tenantID // Task has different tenant
				repo.On("GetByID", mock.Anything, mock.Anything).Return(task, nil)
				return task
			},
			wantErr: ErrTaskNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := new(MockRepository)
			notif := new(MockNotificationService)
			expectedTask := tt.setup(repo)

			service := NewService(repo, notif)
			ctx := context.Background()

			task, err := service.GetTask(ctx, tt.tenantID, tt.taskID)

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				assert.Nil(t, task)
			} else {
				require.NoError(t, err)
				assert.Equal(t, expectedTask.ID, task.ID)
			}

			repo.AssertExpectations(t)
		})
	}
}

func TestService_ApproveTask(t *testing.T) {
	t.Run("authorized user approves", func(t *testing.T) {
		repo := new(MockRepository)
		notif := new(MockNotificationService)

		userID := uuid.New()
		task := createMockTaskWithAssignees([]string{userID.String()})
		repo.On("GetByID", mock.Anything, task.ID).Return(task, nil)
		repo.On("Update", mock.Anything, mock.Anything).Return(nil)
		// Note: notification is async in goroutine, don't assert
		notif.On("NotifyTaskCompleted", mock.Anything, mock.Anything).Maybe().Return(nil)

		service := NewService(repo, notif)
		ctx := context.Background()

		err := service.ApproveTask(ctx, task.TenantID, task.ID, userID, []string{}, ApproveTaskRequest{Comment: "Looks good"})

		require.NoError(t, err)
		repo.AssertExpectations(t)
	})

	t.Run("user with matching role approves", func(t *testing.T) {
		repo := new(MockRepository)
		notif := new(MockNotificationService)

		userID := uuid.New()
		task := createMockTaskWithAssignees([]string{"admin", "manager"})
		repo.On("GetByID", mock.Anything, task.ID).Return(task, nil)
		repo.On("Update", mock.Anything, mock.Anything).Return(nil)
		notif.On("NotifyTaskCompleted", mock.Anything, mock.Anything).Maybe().Return(nil)

		service := NewService(repo, notif)
		ctx := context.Background()

		err := service.ApproveTask(ctx, task.TenantID, task.ID, userID, []string{"admin"}, ApproveTaskRequest{Comment: "Approved"})

		require.NoError(t, err)
		repo.AssertExpectations(t)
	})

	t.Run("unauthorized user", func(t *testing.T) {
		repo := new(MockRepository)
		notif := new(MockNotificationService)

		task := createMockTaskWithAssignees([]string{uuid.New().String()})
		repo.On("GetByID", mock.Anything, task.ID).Return(task, nil)

		service := NewService(repo, notif)
		ctx := context.Background()

		err := service.ApproveTask(ctx, task.TenantID, task.ID, uuid.New(), []string{}, ApproveTaskRequest{Comment: "Approved"})

		assert.ErrorIs(t, err, ErrUnauthorized)
		repo.AssertExpectations(t)
	})

	t.Run("task already completed", func(t *testing.T) {
		repo := new(MockRepository)
		notif := new(MockNotificationService)

		userID := uuid.New()
		task := createMockTaskWithAssignees([]string{userID.String()})
		task.Status = StatusApproved
		repo.On("GetByID", mock.Anything, task.ID).Return(task, nil)

		service := NewService(repo, notif)
		ctx := context.Background()

		err := service.ApproveTask(ctx, task.TenantID, task.ID, userID, []string{}, ApproveTaskRequest{Comment: "Approved"})

		assert.ErrorIs(t, err, ErrTaskNotPending)
		repo.AssertExpectations(t)
	})
}

func TestService_RejectTask(t *testing.T) {
	userID := uuid.New()
	task := createMockTaskWithAssignees([]string{userID.String()})

	repo := new(MockRepository)
	notif := new(MockNotificationService)

	repo.On("GetByID", mock.Anything, task.ID).Return(task, nil)
	repo.On("Update", mock.Anything, mock.Anything).Return(nil)
	notif.On("NotifyTaskCompleted", mock.Anything, mock.Anything).Maybe().Return(nil)

	service := NewService(repo, notif)
	ctx := context.Background()

	request := RejectTaskRequest{Reason: "Not ready"}
	err := service.RejectTask(ctx, task.TenantID, task.ID, userID, []string{}, request)

	require.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestService_SubmitTask(t *testing.T) {
	userID := uuid.New()
	task := createMockTaskWithAssignees([]string{userID.String()})
	task.TaskType = TaskTypeInput

	repo := new(MockRepository)
	notif := new(MockNotificationService)

	repo.On("GetByID", mock.Anything, task.ID).Return(task, nil)
	repo.On("Update", mock.Anything, mock.Anything).Return(nil)
	notif.On("NotifyTaskCompleted", mock.Anything, mock.Anything).Maybe().Return(nil)

	service := NewService(repo, notif)
	ctx := context.Background()

	request := SubmitTaskRequest{
		Data: map[string]interface{}{
			"field1": "value1",
			"field2": 42,
		},
	}

	err := service.SubmitTask(ctx, task.TenantID, task.ID, userID, []string{}, request)

	require.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestService_ListTasks(t *testing.T) {
	tenantID := uuid.New()
	userID := uuid.New()

	tasks := []*HumanTask{
		createMockTaskWithAssignees([]string{userID.String()}),
		createMockTaskWithAssignees([]string{userID.String()}),
	}

	repo := new(MockRepository)
	notif := new(MockNotificationService)

	repo.On("List", mock.Anything, mock.MatchedBy(func(f TaskFilter) bool {
		return f.TenantID == tenantID && *f.Assignee == userID.String()
	})).Return(tasks, nil)

	service := NewService(repo, notif)
	ctx := context.Background()

	userIDStr := userID.String()
	filter := TaskFilter{
		TenantID: tenantID,
		Assignee: &userIDStr,
		Limit:    10,
	}

	result, err := service.ListTasks(ctx, filter)

	require.NoError(t, err)
	assert.Len(t, result, 2)
	repo.AssertExpectations(t)
}

func TestService_ProcessOverdueTasks(t *testing.T) {
	tenantID := uuid.New()

	// Task with auto-approve on timeout
	config1, _ := json.Marshal(HumanTaskConfig{
		OnTimeout: TimeoutActionAutoApprove,
	})
	task1 := createMockTask()
	task1.Config = config1
	task1.Status = StatusPending

	// Task with auto-reject on timeout
	config2, _ := json.Marshal(HumanTaskConfig{
		OnTimeout: TimeoutActionAutoReject,
	})
	task2 := createMockTask()
	task2.Config = config2
	task2.Status = StatusPending

	overdueTasks := []*HumanTask{task1, task2}

	repo := new(MockRepository)
	notif := new(MockNotificationService)

	repo.On("GetOverdueTasks", mock.Anything, tenantID).Return(overdueTasks, nil)
	repo.On("Update", mock.Anything, mock.Anything).Return(nil).Times(2)
	notif.On("NotifyTaskCompleted", mock.Anything, mock.Anything).Maybe().Return(nil)

	service := NewService(repo, notif)
	ctx := context.Background()

	err := service.ProcessOverdueTasks(ctx, tenantID)

	require.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestService_CancelTasksByExecution(t *testing.T) {
	tenantID := uuid.New()
	executionID := uuid.New()

	tasks := []*HumanTask{
		createMockTask(),
		createMockTask(),
	}

	repo := new(MockRepository)
	notif := new(MockNotificationService)

	repo.On("List", mock.Anything, mock.MatchedBy(func(f TaskFilter) bool {
		return f.TenantID == tenantID && *f.ExecutionID == executionID
	})).Return(tasks, nil)
	repo.On("Update", mock.Anything, mock.Anything).Return(nil).Times(2)

	service := NewService(repo, notif)
	ctx := context.Background()

	err := service.CancelTasksByExecution(ctx, tenantID, executionID)

	require.NoError(t, err)
	repo.AssertExpectations(t)
}

// Helper functions

func createMockTask() *HumanTask {
	assignees, _ := json.Marshal([]string{uuid.New().String()})
	config, _ := json.Marshal(map[string]interface{}{})

	return &HumanTask{
		ID:          uuid.New(),
		TenantID:    uuid.New(),
		ExecutionID: uuid.New(),
		StepID:      "step-1",
		TaskType:    TaskTypeApproval,
		Title:       "Test task",
		Description: "Test description",
		Assignees:   assignees,
		Status:      StatusPending,
		Config:      config,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
}

func createMockTaskWithAssignees(assignees []string) *HumanTask {
	task := createMockTask()
	assigneesJSON, _ := json.Marshal(assignees)
	task.Assignees = assigneesJSON
	return task
}
