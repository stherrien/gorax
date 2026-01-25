package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/gorax/gorax/internal/humantask"
)

// MockHumanTaskService is a mock implementation of humantask.Service
type MockHumanTaskService struct {
	mock.Mock
}

func (m *MockHumanTaskService) CreateTask(ctx context.Context, tenantID uuid.UUID, req humantask.CreateTaskRequest) (*humantask.HumanTask, error) {
	args := m.Called(ctx, tenantID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*humantask.HumanTask), args.Error(1)
}

func (m *MockHumanTaskService) GetTask(ctx context.Context, tenantID uuid.UUID, taskID uuid.UUID) (*humantask.HumanTask, error) {
	args := m.Called(ctx, tenantID, taskID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*humantask.HumanTask), args.Error(1)
}

func (m *MockHumanTaskService) ListTasks(ctx context.Context, filter humantask.TaskFilter) ([]*humantask.HumanTask, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*humantask.HumanTask), args.Error(1)
}

func (m *MockHumanTaskService) ApproveTask(ctx context.Context, tenantID uuid.UUID, taskID uuid.UUID,
	userID uuid.UUID, roles []string, req humantask.ApproveTaskRequest) error {
	args := m.Called(ctx, tenantID, taskID, userID, roles, req)
	return args.Error(0)
}

func (m *MockHumanTaskService) RejectTask(ctx context.Context, tenantID uuid.UUID, taskID uuid.UUID,
	userID uuid.UUID, roles []string, req humantask.RejectTaskRequest) error {
	args := m.Called(ctx, tenantID, taskID, userID, roles, req)
	return args.Error(0)
}

func (m *MockHumanTaskService) SubmitTask(ctx context.Context, tenantID uuid.UUID, taskID uuid.UUID,
	userID uuid.UUID, roles []string, req humantask.SubmitTaskRequest) error {
	args := m.Called(ctx, tenantID, taskID, userID, roles, req)
	return args.Error(0)
}

func (m *MockHumanTaskService) ProcessOverdueTasks(ctx context.Context, tenantID uuid.UUID) error {
	args := m.Called(ctx, tenantID)
	return args.Error(0)
}

func (m *MockHumanTaskService) CancelTasksByExecution(ctx context.Context, tenantID uuid.UUID, executionID uuid.UUID) error {
	args := m.Called(ctx, tenantID, executionID)
	return args.Error(0)
}

func (m *MockHumanTaskService) GetEscalationHistory(ctx context.Context, tenantID uuid.UUID, taskID uuid.UUID) (*humantask.EscalationHistory, error) {
	args := m.Called(ctx, tenantID, taskID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*humantask.EscalationHistory), args.Error(1)
}

func (m *MockHumanTaskService) UpdateEscalationConfig(ctx context.Context, tenantID uuid.UUID, taskID uuid.UUID, req humantask.UpdateEscalationRequest) error {
	args := m.Called(ctx, tenantID, taskID, req)
	return args.Error(0)
}

func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	return gin.New()
}

func TestHumanTaskHandler_ListTasks(t *testing.T) {
	tests := []struct {
		name           string
		queryParams    string
		setup          func(*MockHumanTaskService)
		expectedStatus int
		expectedCount  int
	}{
		{
			name:        "list all tasks",
			queryParams: "",
			setup: func(svc *MockHumanTaskService) {
				tasks := []*humantask.HumanTask{
					createTestHumanTask(),
					createTestHumanTask(),
				}
				svc.On("ListTasks", mock.Anything, mock.Anything).Return(tasks, nil)
			},
			expectedStatus: http.StatusOK,
			expectedCount:  2,
		},
		{
			name:        "filter by status",
			queryParams: "?status=pending",
			setup: func(svc *MockHumanTaskService) {
				tasks := []*humantask.HumanTask{createTestHumanTask()}
				svc.On("ListTasks", mock.Anything, mock.MatchedBy(func(f humantask.TaskFilter) bool {
					return f.Status != nil && *f.Status == humantask.StatusPending
				})).Return(tasks, nil)
			},
			expectedStatus: http.StatusOK,
			expectedCount:  1,
		},
		{
			name:        "filter by assignee",
			queryParams: "?assignee=" + uuid.New().String(),
			setup: func(svc *MockHumanTaskService) {
				tasks := []*humantask.HumanTask{createTestHumanTask()}
				svc.On("ListTasks", mock.Anything, mock.Anything).Return(tasks, nil)
			},
			expectedStatus: http.StatusOK,
			expectedCount:  1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := new(MockHumanTaskService)
			tt.setup(svc)

			router := setupTestRouter()
			handler := NewHumanTaskHandler(svc)
			router.GET("/api/v1/tasks", setTestContext(), handler.ListTasks)

			req := httptest.NewRequest(http.MethodGet, "/api/v1/tasks"+tt.queryParams, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedStatus == http.StatusOK {
				var response struct {
					Tasks []*humantask.TaskResponse `json:"tasks"`
				}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.Len(t, response.Tasks, tt.expectedCount)
			}

			svc.AssertExpectations(t)
		})
	}
}

func TestHumanTaskHandler_GetTask(t *testing.T) {
	taskID := uuid.New()

	tests := []struct {
		name           string
		taskID         string
		setup          func(*MockHumanTaskService)
		expectedStatus int
	}{
		{
			name:   "get existing task",
			taskID: taskID.String(),
			setup: func(svc *MockHumanTaskService) {
				task := createTestHumanTask()
				task.ID = taskID
				svc.On("GetTask", mock.Anything, mock.Anything, taskID).Return(task, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:   "task not found",
			taskID: uuid.New().String(),
			setup: func(svc *MockHumanTaskService) {
				svc.On("GetTask", mock.Anything, mock.Anything, mock.Anything).
					Return(nil, humantask.ErrTaskNotFound)
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "invalid task ID",
			taskID:         "invalid",
			setup:          func(svc *MockHumanTaskService) {},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := new(MockHumanTaskService)
			tt.setup(svc)

			router := setupTestRouter()
			handler := NewHumanTaskHandler(svc)
			router.GET("/api/v1/tasks/:id", setTestContext(), handler.GetTask)

			req := httptest.NewRequest(http.MethodGet, "/api/v1/tasks/"+tt.taskID, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			svc.AssertExpectations(t)
		})
	}
}

func TestHumanTaskHandler_ApproveTask(t *testing.T) {
	taskID := uuid.New()

	tests := []struct {
		name           string
		taskID         string
		body           humantask.ApproveTaskRequest
		setup          func(*MockHumanTaskService)
		expectedStatus int
	}{
		{
			name:   "approve task successfully",
			taskID: taskID.String(),
			body: humantask.ApproveTaskRequest{
				Comment: "Looks good",
			},
			setup: func(svc *MockHumanTaskService) {
				svc.On("ApproveTask", mock.Anything, mock.Anything, taskID,
					mock.Anything, mock.Anything, mock.Anything).Return(nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:   "unauthorized user",
			taskID: taskID.String(),
			body: humantask.ApproveTaskRequest{
				Comment: "Approved",
			},
			setup: func(svc *MockHumanTaskService) {
				svc.On("ApproveTask", mock.Anything, mock.Anything, mock.Anything,
					mock.Anything, mock.Anything, mock.Anything).
					Return(humantask.ErrUnauthorized)
			},
			expectedStatus: http.StatusForbidden,
		},
		{
			name:   "task not pending",
			taskID: taskID.String(),
			body: humantask.ApproveTaskRequest{
				Comment: "Approved",
			},
			setup: func(svc *MockHumanTaskService) {
				svc.On("ApproveTask", mock.Anything, mock.Anything, mock.Anything,
					mock.Anything, mock.Anything, mock.Anything).
					Return(humantask.ErrTaskNotPending)
			},
			expectedStatus: http.StatusConflict,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := new(MockHumanTaskService)
			tt.setup(svc)

			router := setupTestRouter()
			handler := NewHumanTaskHandler(svc)
			router.POST("/api/v1/tasks/:id/approve", setTestContext(), handler.ApproveTask)

			bodyBytes, _ := json.Marshal(tt.body)
			req := httptest.NewRequest(http.MethodPost, "/api/v1/tasks/"+tt.taskID+"/approve",
				bytes.NewReader(bodyBytes))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			svc.AssertExpectations(t)
		})
	}
}

func TestHumanTaskHandler_RejectTask(t *testing.T) {
	taskID := uuid.New()

	svc := new(MockHumanTaskService)
	svc.On("RejectTask", mock.Anything, mock.Anything, taskID,
		mock.Anything, mock.Anything, mock.Anything).Return(nil)

	router := setupTestRouter()
	handler := NewHumanTaskHandler(svc)
	router.POST("/api/v1/tasks/:id/reject", setTestContext(), handler.RejectTask)

	body := humantask.RejectTaskRequest{
		Reason: "Not ready",
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/tasks/"+taskID.String()+"/reject",
		bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	svc.AssertExpectations(t)
}

func TestHumanTaskHandler_SubmitTask(t *testing.T) {
	taskID := uuid.New()

	svc := new(MockHumanTaskService)
	svc.On("SubmitTask", mock.Anything, mock.Anything, taskID,
		mock.Anything, mock.Anything, mock.Anything).Return(nil)

	router := setupTestRouter()
	handler := NewHumanTaskHandler(svc)
	router.POST("/api/v1/tasks/:id/submit", setTestContext(), handler.SubmitTask)

	body := humantask.SubmitTaskRequest{
		Data: map[string]interface{}{
			"field1": "value1",
			"field2": 42,
		},
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/tasks/"+taskID.String()+"/submit",
		bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	svc.AssertExpectations(t)
}

func TestHumanTaskHandler_GetEscalationHistory(t *testing.T) {
	taskID := uuid.New()

	tests := []struct {
		name           string
		taskID         string
		setup          func(*MockHumanTaskService)
		expectedStatus int
	}{
		{
			name:   "get escalation history successfully",
			taskID: taskID.String(),
			setup: func(svc *MockHumanTaskService) {
				history := &humantask.EscalationHistory{
					TaskID:      taskID,
					Escalations: []humantask.EscalationSummary{},
				}
				svc.On("GetEscalationHistory", mock.Anything, mock.Anything, taskID).Return(history, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:   "task not found",
			taskID: uuid.New().String(),
			setup: func(svc *MockHumanTaskService) {
				svc.On("GetEscalationHistory", mock.Anything, mock.Anything, mock.Anything).
					Return(nil, humantask.ErrTaskNotFound)
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "invalid task ID",
			taskID:         "invalid",
			setup:          func(svc *MockHumanTaskService) {},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := new(MockHumanTaskService)
			tt.setup(svc)

			router := setupTestRouter()
			handler := NewHumanTaskHandler(svc)
			router.GET("/api/v1/tasks/:id/escalations", setTestContext(), handler.GetEscalationHistory)

			req := httptest.NewRequest(http.MethodGet, "/api/v1/tasks/"+tt.taskID+"/escalations", nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			svc.AssertExpectations(t)
		})
	}
}

func TestHumanTaskHandler_UpdateEscalationConfig(t *testing.T) {
	taskID := uuid.New()

	tests := []struct {
		name           string
		taskID         string
		body           humantask.UpdateEscalationRequest
		setup          func(*MockHumanTaskService)
		expectedStatus int
	}{
		{
			name:   "update escalation config successfully",
			taskID: taskID.String(),
			body: humantask.UpdateEscalationRequest{
				Config: humantask.EscalationConfig{
					Enabled: true,
					Levels: []humantask.EscalationLevel{
						{
							Level:           1,
							TimeoutMinutes:  60,
							BackupApprovers: []string{"backup@example.com"},
						},
					},
				},
			},
			setup: func(svc *MockHumanTaskService) {
				svc.On("UpdateEscalationConfig", mock.Anything, mock.Anything, taskID, mock.Anything).Return(nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:   "task not found",
			taskID: uuid.New().String(),
			body: humantask.UpdateEscalationRequest{
				Config: humantask.EscalationConfig{
					Enabled: true,
					Levels: []humantask.EscalationLevel{
						{Level: 1, TimeoutMinutes: 60, BackupApprovers: []string{"backup@example.com"}},
					},
				},
			},
			setup: func(svc *MockHumanTaskService) {
				svc.On("UpdateEscalationConfig", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(humantask.ErrTaskNotFound)
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:   "task not pending",
			taskID: taskID.String(),
			body: humantask.UpdateEscalationRequest{
				Config: humantask.EscalationConfig{
					Enabled: true,
					Levels: []humantask.EscalationLevel{
						{Level: 1, TimeoutMinutes: 60, BackupApprovers: []string{"backup@example.com"}},
					},
				},
			},
			setup: func(svc *MockHumanTaskService) {
				svc.On("UpdateEscalationConfig", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(humantask.ErrTaskNotPending)
			},
			expectedStatus: http.StatusConflict,
		},
		{
			name:           "invalid task ID",
			taskID:         "invalid",
			body:           humantask.UpdateEscalationRequest{},
			setup:          func(svc *MockHumanTaskService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:   "enabled without levels",
			taskID: taskID.String(),
			body: humantask.UpdateEscalationRequest{
				Config: humantask.EscalationConfig{
					Enabled: true,
					Levels:  []humantask.EscalationLevel{},
				},
			},
			setup:          func(svc *MockHumanTaskService) {},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := new(MockHumanTaskService)
			tt.setup(svc)

			router := setupTestRouter()
			handler := NewHumanTaskHandler(svc)
			router.PUT("/api/v1/tasks/:id/escalation", setTestContext(), handler.UpdateEscalationConfig)

			bodyBytes, _ := json.Marshal(tt.body)
			req := httptest.NewRequest(http.MethodPut, "/api/v1/tasks/"+tt.taskID+"/escalation",
				bytes.NewReader(bodyBytes))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			svc.AssertExpectations(t)
		})
	}
}

// Helper functions

func createTestHumanTask() *humantask.HumanTask {
	assignees, _ := json.Marshal([]string{uuid.New().String()})
	config, _ := json.Marshal(map[string]interface{}{})

	return &humantask.HumanTask{
		ID:          uuid.New(),
		TenantID:    uuid.New(),
		ExecutionID: uuid.New(),
		StepID:      "step-1",
		TaskType:    humantask.TaskTypeApproval,
		Title:       "Test task",
		Description: "Test description",
		Assignees:   assignees,
		Status:      humantask.StatusPending,
		Config:      config,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
}

func setTestContext() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Set tenant ID and user ID in context for testing
		c.Set("tenant_id", uuid.New().String())
		c.Set("user_id", uuid.New().String())
		c.Set("user_roles", []string{"admin"})
		c.Next()
	}
}
