package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/gorax/gorax/internal/api/middleware"
	"github.com/gorax/gorax/internal/schedule"
	"github.com/gorax/gorax/internal/tenant"
)

// MockScheduleService is a mock implementation of ScheduleService for testing
type MockScheduleService struct {
	mock.Mock
}

func (m *MockScheduleService) Create(ctx context.Context, tenantID, workflowID, userID string, input schedule.CreateScheduleInput) (*schedule.Schedule, error) {
	args := m.Called(ctx, tenantID, workflowID, userID, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*schedule.Schedule), args.Error(1)
}

func (m *MockScheduleService) GetByID(ctx context.Context, tenantID, id string) (*schedule.Schedule, error) {
	args := m.Called(ctx, tenantID, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*schedule.Schedule), args.Error(1)
}

func (m *MockScheduleService) Update(ctx context.Context, tenantID, id string, input schedule.UpdateScheduleInput) (*schedule.Schedule, error) {
	args := m.Called(ctx, tenantID, id, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*schedule.Schedule), args.Error(1)
}

func (m *MockScheduleService) Delete(ctx context.Context, tenantID, id string) error {
	args := m.Called(ctx, tenantID, id)
	return args.Error(0)
}

func (m *MockScheduleService) List(ctx context.Context, tenantID, workflowID string, limit, offset int) ([]*schedule.Schedule, error) {
	args := m.Called(ctx, tenantID, workflowID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*schedule.Schedule), args.Error(1)
}

func (m *MockScheduleService) ListAll(ctx context.Context, tenantID string, limit, offset int) ([]*schedule.ScheduleWithWorkflow, error) {
	args := m.Called(ctx, tenantID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*schedule.ScheduleWithWorkflow), args.Error(1)
}

func (m *MockScheduleService) ParseNextRunTime(expression, timezone string) (time.Time, error) {
	args := m.Called(expression, timezone)
	return args.Get(0).(time.Time), args.Error(1)
}

func (m *MockScheduleService) GetNextRunTimes(expression, timezone string, count int) ([]time.Time, error) {
	args := m.Called(expression, timezone, count)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]time.Time), args.Error(1)
}

func (m *MockScheduleService) ListExecutionLogs(ctx context.Context, tenantID, scheduleID string, limit, offset int) ([]*schedule.ExecutionLog, error) {
	args := m.Called(ctx, tenantID, scheduleID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*schedule.ExecutionLog), args.Error(1)
}

func (m *MockScheduleService) GetExecutionLog(ctx context.Context, tenantID, logID string) (*schedule.ExecutionLog, error) {
	args := m.Called(ctx, tenantID, logID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*schedule.ExecutionLog), args.Error(1)
}

func (m *MockScheduleService) CountExecutionLogs(ctx context.Context, tenantID, scheduleID string) (int, error) {
	args := m.Called(ctx, tenantID, scheduleID)
	return args.Int(0), args.Error(1)
}

func newTestScheduleHandler() (*ScheduleHandler, *MockScheduleService) {
	mockService := new(MockScheduleService)
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	handler := NewScheduleHandler(mockService, logger)
	return handler, mockService
}

func addScheduleContext(req *http.Request, tenantID string, user *middleware.User) *http.Request {
	t := &tenant.Tenant{
		ID:     tenantID,
		Status: "active",
	}
	ctx := context.WithValue(req.Context(), middleware.TenantContextKey, t)
	if user != nil {
		ctx = context.WithValue(ctx, middleware.UserContextKey, user)
	}
	return req.WithContext(ctx)
}

func addScheduleURLParams(req *http.Request, params map[string]string) *http.Request {
	rctx := chi.NewRouteContext()
	for key, value := range params {
		rctx.URLParams.Add(key, value)
	}
	return req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
}

// Test fixtures
func createTestSchedule() *schedule.Schedule {
	now := time.Now()
	return &schedule.Schedule{
		ID:             "sched-123",
		TenantID:       "tenant-123",
		WorkflowID:     "workflow-123",
		Name:           "Test Schedule",
		CronExpression: "0 0 * * *",
		Timezone:       "UTC",
		Enabled:        true,
		OverlapPolicy:  schedule.OverlapPolicySkip,
		NextRunAt:      &now,
		CreatedAt:      now,
		UpdatedAt:      now,
		CreatedBy:      "user-123",
	}
}

func createTestScheduleWithWorkflow() *schedule.ScheduleWithWorkflow {
	sched := createTestSchedule()
	return &schedule.ScheduleWithWorkflow{
		Schedule:     *sched,
		WorkflowName: "Test Workflow",
	}
}

func createTestExecutionLog() *schedule.ExecutionLog {
	now := time.Now()
	execID := "exec-123"
	return &schedule.ExecutionLog{
		ID:          "log-123",
		TenantID:    "tenant-123",
		ScheduleID:  "sched-123",
		ExecutionID: &execID,
		Status:      schedule.ExecutionLogStatusCompleted,
		StartedAt:   &now,
		CompletedAt: &now,
		TriggerTime: now,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

// ============================================================================
// Create Handler Tests
// ============================================================================

func TestScheduleHandler_Create(t *testing.T) {
	tests := []struct {
		name           string
		tenantID       string
		workflowID     string
		user           *middleware.User
		body           interface{}
		setupMock      func(*MockScheduleService)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:       "successful creation",
			tenantID:   "tenant-123",
			workflowID: "workflow-123",
			user:       &middleware.User{ID: "user-123"},
			body: schedule.CreateScheduleInput{
				Name:           "Daily Report",
				CronExpression: "0 0 * * *",
				Timezone:       "UTC",
				Enabled:        true,
			},
			setupMock: func(m *MockScheduleService) {
				m.On("Create", mock.Anything, "tenant-123", "workflow-123", "user-123", mock.AnythingOfType("schedule.CreateScheduleInput")).
					Return(createTestSchedule(), nil)
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "invalid request body",
			tenantID:       "tenant-123",
			workflowID:     "workflow-123",
			user:           &middleware.User{ID: "user-123"},
			body:           "invalid json",
			setupMock:      func(m *MockScheduleService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "invalid request body",
		},
		{
			name:       "validation error - invalid cron",
			tenantID:   "tenant-123",
			workflowID: "workflow-123",
			user:       &middleware.User{ID: "user-123"},
			body: schedule.CreateScheduleInput{
				Name:           "Daily Report",
				CronExpression: "invalid cron",
			},
			setupMock: func(m *MockScheduleService) {
				m.On("Create", mock.Anything, "tenant-123", "workflow-123", "user-123", mock.AnythingOfType("schedule.CreateScheduleInput")).
					Return(nil, &schedule.ValidationError{Message: "invalid cron expression"})
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "invalid cron expression",
		},
		{
			name:       "service error",
			tenantID:   "tenant-123",
			workflowID: "workflow-123",
			user:       &middleware.User{ID: "user-123"},
			body: schedule.CreateScheduleInput{
				Name:           "Daily Report",
				CronExpression: "0 0 * * *",
			},
			setupMock: func(m *MockScheduleService) {
				m.On("Create", mock.Anything, "tenant-123", "workflow-123", "user-123", mock.AnythingOfType("schedule.CreateScheduleInput")).
					Return(nil, errors.New("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "failed to create schedule",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, mockService := newTestScheduleHandler()
			tt.setupMock(mockService)

			var body []byte
			var err error
			switch v := tt.body.(type) {
			case string:
				body = []byte(v)
			default:
				body, err = json.Marshal(tt.body)
				require.NoError(t, err)
			}

			req := httptest.NewRequest(http.MethodPost, "/api/v1/workflows/"+tt.workflowID+"/schedules", bytes.NewReader(body))
			req = addScheduleContext(req, tt.tenantID, tt.user)
			req = addScheduleURLParams(req, map[string]string{"workflowID": tt.workflowID})
			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()
			handler.Create(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			if tt.expectedBody != "" {
				assert.Contains(t, rr.Body.String(), tt.expectedBody)
			}
			mockService.AssertExpectations(t)
		})
	}
}

// ============================================================================
// List Handler Tests
// ============================================================================

func TestScheduleHandler_List(t *testing.T) {
	tests := []struct {
		name           string
		tenantID       string
		workflowID     string
		queryParams    string
		setupMock      func(*MockScheduleService)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:        "successful list",
			tenantID:    "tenant-123",
			workflowID:  "workflow-123",
			queryParams: "",
			setupMock: func(m *MockScheduleService) {
				schedules := []*schedule.Schedule{createTestSchedule()}
				m.On("List", mock.Anything, "tenant-123", "workflow-123", 20, 0).Return(schedules, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:        "successful list with pagination",
			tenantID:    "tenant-123",
			workflowID:  "workflow-123",
			queryParams: "?limit=10&offset=5",
			setupMock: func(m *MockScheduleService) {
				schedules := []*schedule.Schedule{createTestSchedule()}
				m.On("List", mock.Anything, "tenant-123", "workflow-123", 10, 5).Return(schedules, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:        "empty list",
			tenantID:    "tenant-123",
			workflowID:  "workflow-123",
			queryParams: "",
			setupMock: func(m *MockScheduleService) {
				m.On("List", mock.Anything, "tenant-123", "workflow-123", 20, 0).Return([]*schedule.Schedule{}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:        "service error",
			tenantID:    "tenant-123",
			workflowID:  "workflow-123",
			queryParams: "",
			setupMock: func(m *MockScheduleService) {
				m.On("List", mock.Anything, "tenant-123", "workflow-123", 20, 0).Return(nil, errors.New("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "failed to list schedules",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, mockService := newTestScheduleHandler()
			tt.setupMock(mockService)

			req := httptest.NewRequest(http.MethodGet, "/api/v1/workflows/"+tt.workflowID+"/schedules"+tt.queryParams, nil)
			req = addScheduleContext(req, tt.tenantID, nil)
			req = addScheduleURLParams(req, map[string]string{"workflowID": tt.workflowID})

			rr := httptest.NewRecorder()
			handler.List(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			if tt.expectedBody != "" {
				assert.Contains(t, rr.Body.String(), tt.expectedBody)
			}
			mockService.AssertExpectations(t)
		})
	}
}

// ============================================================================
// ListAll Handler Tests
// ============================================================================

func TestScheduleHandler_ListAll(t *testing.T) {
	tests := []struct {
		name           string
		tenantID       string
		queryParams    string
		setupMock      func(*MockScheduleService)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:        "successful list all",
			tenantID:    "tenant-123",
			queryParams: "",
			setupMock: func(m *MockScheduleService) {
				schedules := []*schedule.ScheduleWithWorkflow{createTestScheduleWithWorkflow()}
				m.On("ListAll", mock.Anything, "tenant-123", 20, 0).Return(schedules, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:        "successful list all with pagination",
			tenantID:    "tenant-123",
			queryParams: "?limit=50&offset=10",
			setupMock: func(m *MockScheduleService) {
				schedules := []*schedule.ScheduleWithWorkflow{createTestScheduleWithWorkflow()}
				m.On("ListAll", mock.Anything, "tenant-123", 50, 10).Return(schedules, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:        "service error",
			tenantID:    "tenant-123",
			queryParams: "",
			setupMock: func(m *MockScheduleService) {
				m.On("ListAll", mock.Anything, "tenant-123", 20, 0).Return(nil, errors.New("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "failed to list schedules",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, mockService := newTestScheduleHandler()
			tt.setupMock(mockService)

			req := httptest.NewRequest(http.MethodGet, "/api/v1/schedules"+tt.queryParams, nil)
			req = addScheduleContext(req, tt.tenantID, nil)

			rr := httptest.NewRecorder()
			handler.ListAll(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			if tt.expectedBody != "" {
				assert.Contains(t, rr.Body.String(), tt.expectedBody)
			}
			mockService.AssertExpectations(t)
		})
	}
}

// ============================================================================
// Get Handler Tests
// ============================================================================

func TestScheduleHandler_Get(t *testing.T) {
	tests := []struct {
		name           string
		tenantID       string
		scheduleID     string
		setupMock      func(*MockScheduleService)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:       "successful get",
			tenantID:   "tenant-123",
			scheduleID: "sched-123",
			setupMock: func(m *MockScheduleService) {
				m.On("GetByID", mock.Anything, "tenant-123", "sched-123").Return(createTestSchedule(), nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:       "schedule not found",
			tenantID:   "tenant-123",
			scheduleID: "nonexistent",
			setupMock: func(m *MockScheduleService) {
				m.On("GetByID", mock.Anything, "tenant-123", "nonexistent").Return(nil, schedule.ErrNotFound)
			},
			expectedStatus: http.StatusNotFound,
			expectedBody:   "schedule not found",
		},
		{
			name:       "service error",
			tenantID:   "tenant-123",
			scheduleID: "sched-123",
			setupMock: func(m *MockScheduleService) {
				m.On("GetByID", mock.Anything, "tenant-123", "sched-123").Return(nil, errors.New("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "failed to get schedule",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, mockService := newTestScheduleHandler()
			tt.setupMock(mockService)

			req := httptest.NewRequest(http.MethodGet, "/api/v1/schedules/"+tt.scheduleID, nil)
			req = addScheduleContext(req, tt.tenantID, nil)
			req = addScheduleURLParams(req, map[string]string{"scheduleID": tt.scheduleID})

			rr := httptest.NewRecorder()
			handler.Get(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			if tt.expectedBody != "" {
				assert.Contains(t, rr.Body.String(), tt.expectedBody)
			}
			mockService.AssertExpectations(t)
		})
	}
}

// ============================================================================
// Update Handler Tests
// ============================================================================

func TestScheduleHandler_Update(t *testing.T) {
	enabledTrue := true
	tests := []struct {
		name           string
		tenantID       string
		scheduleID     string
		body           interface{}
		setupMock      func(*MockScheduleService)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:       "successful update",
			tenantID:   "tenant-123",
			scheduleID: "sched-123",
			body: schedule.UpdateScheduleInput{
				Name:    stringPtr("Updated Schedule"),
				Enabled: &enabledTrue,
			},
			setupMock: func(m *MockScheduleService) {
				m.On("Update", mock.Anything, "tenant-123", "sched-123", mock.AnythingOfType("schedule.UpdateScheduleInput")).
					Return(createTestSchedule(), nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "invalid request body",
			tenantID:       "tenant-123",
			scheduleID:     "sched-123",
			body:           "invalid json",
			setupMock:      func(m *MockScheduleService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "invalid request body",
		},
		{
			name:       "schedule not found",
			tenantID:   "tenant-123",
			scheduleID: "nonexistent",
			body: schedule.UpdateScheduleInput{
				Name: stringPtr("Updated Schedule"),
			},
			setupMock: func(m *MockScheduleService) {
				m.On("Update", mock.Anything, "tenant-123", "nonexistent", mock.AnythingOfType("schedule.UpdateScheduleInput")).
					Return(nil, schedule.ErrNotFound)
			},
			expectedStatus: http.StatusNotFound,
			expectedBody:   "schedule not found",
		},
		{
			name:       "validation error",
			tenantID:   "tenant-123",
			scheduleID: "sched-123",
			body: schedule.UpdateScheduleInput{
				CronExpression: stringPtr("invalid cron"),
			},
			setupMock: func(m *MockScheduleService) {
				m.On("Update", mock.Anything, "tenant-123", "sched-123", mock.AnythingOfType("schedule.UpdateScheduleInput")).
					Return(nil, &schedule.ValidationError{Message: "invalid cron expression"})
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "invalid cron expression",
		},
		{
			name:       "service error",
			tenantID:   "tenant-123",
			scheduleID: "sched-123",
			body: schedule.UpdateScheduleInput{
				Name: stringPtr("Updated Schedule"),
			},
			setupMock: func(m *MockScheduleService) {
				m.On("Update", mock.Anything, "tenant-123", "sched-123", mock.AnythingOfType("schedule.UpdateScheduleInput")).
					Return(nil, errors.New("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "failed to update schedule",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, mockService := newTestScheduleHandler()
			tt.setupMock(mockService)

			var body []byte
			var err error
			switch v := tt.body.(type) {
			case string:
				body = []byte(v)
			default:
				body, err = json.Marshal(tt.body)
				require.NoError(t, err)
			}

			req := httptest.NewRequest(http.MethodPut, "/api/v1/schedules/"+tt.scheduleID, bytes.NewReader(body))
			req = addScheduleContext(req, tt.tenantID, nil)
			req = addScheduleURLParams(req, map[string]string{"scheduleID": tt.scheduleID})
			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()
			handler.Update(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			if tt.expectedBody != "" {
				assert.Contains(t, rr.Body.String(), tt.expectedBody)
			}
			mockService.AssertExpectations(t)
		})
	}
}

// ============================================================================
// Delete Handler Tests
// ============================================================================

func TestScheduleHandler_Delete(t *testing.T) {
	tests := []struct {
		name           string
		tenantID       string
		scheduleID     string
		setupMock      func(*MockScheduleService)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:       "successful delete",
			tenantID:   "tenant-123",
			scheduleID: "sched-123",
			setupMock: func(m *MockScheduleService) {
				m.On("Delete", mock.Anything, "tenant-123", "sched-123").Return(nil)
			},
			expectedStatus: http.StatusNoContent,
		},
		{
			name:       "schedule not found",
			tenantID:   "tenant-123",
			scheduleID: "nonexistent",
			setupMock: func(m *MockScheduleService) {
				m.On("Delete", mock.Anything, "tenant-123", "nonexistent").Return(schedule.ErrNotFound)
			},
			expectedStatus: http.StatusNotFound,
			expectedBody:   "schedule not found",
		},
		{
			name:       "service error",
			tenantID:   "tenant-123",
			scheduleID: "sched-123",
			setupMock: func(m *MockScheduleService) {
				m.On("Delete", mock.Anything, "tenant-123", "sched-123").Return(errors.New("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "failed to delete schedule",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, mockService := newTestScheduleHandler()
			tt.setupMock(mockService)

			req := httptest.NewRequest(http.MethodDelete, "/api/v1/schedules/"+tt.scheduleID, nil)
			req = addScheduleContext(req, tt.tenantID, nil)
			req = addScheduleURLParams(req, map[string]string{"scheduleID": tt.scheduleID})

			rr := httptest.NewRecorder()
			handler.Delete(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			if tt.expectedBody != "" {
				assert.Contains(t, rr.Body.String(), tt.expectedBody)
			}
			mockService.AssertExpectations(t)
		})
	}
}

// ============================================================================
// ParseCron Handler Tests
// ============================================================================

func TestScheduleHandler_ParseCron(t *testing.T) {
	fixedTime := time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name           string
		body           interface{}
		setupMock      func(*MockScheduleService)
		expectedStatus int
		expectedBody   string
		checkResponse  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name: "successful parse",
			body: map[string]string{
				"cron_expression": "0 0 * * *",
				"timezone":        "UTC",
			},
			setupMock: func(m *MockScheduleService) {
				m.On("ParseNextRunTime", "0 0 * * *", "UTC").Return(fixedTime, nil)
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, rr *httptest.ResponseRecorder) {
				var resp map[string]interface{}
				err := json.Unmarshal(rr.Body.Bytes(), &resp)
				require.NoError(t, err)
				assert.True(t, resp["valid"].(bool))
				assert.NotEmpty(t, resp["next_run"])
			},
		},
		{
			name: "successful parse with default timezone",
			body: map[string]string{
				"cron_expression": "0 0 * * *",
			},
			setupMock: func(m *MockScheduleService) {
				m.On("ParseNextRunTime", "0 0 * * *", "UTC").Return(fixedTime, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "invalid request body",
			body:           "invalid json",
			setupMock:      func(m *MockScheduleService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "invalid request body",
		},
		{
			name: "invalid cron expression",
			body: map[string]string{
				"cron_expression": "invalid cron",
				"timezone":        "UTC",
			},
			setupMock: func(m *MockScheduleService) {
				m.On("ParseNextRunTime", "invalid cron", "UTC").Return(time.Time{}, &schedule.ValidationError{Message: "invalid format"})
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "invalid cron expression",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, mockService := newTestScheduleHandler()
			tt.setupMock(mockService)

			var body []byte
			var err error
			switch v := tt.body.(type) {
			case string:
				body = []byte(v)
			default:
				body, err = json.Marshal(tt.body)
				require.NoError(t, err)
			}

			req := httptest.NewRequest(http.MethodPost, "/api/v1/schedules/parse", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()
			handler.ParseCron(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			if tt.expectedBody != "" {
				assert.Contains(t, rr.Body.String(), tt.expectedBody)
			}
			if tt.checkResponse != nil {
				tt.checkResponse(t, rr)
			}
			mockService.AssertExpectations(t)
		})
	}
}

// ============================================================================
// PreviewSchedule Handler Tests
// ============================================================================

func TestScheduleHandler_PreviewSchedule(t *testing.T) {
	fixedTimes := []time.Time{
		time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC),
		time.Date(2024, 1, 16, 12, 0, 0, 0, time.UTC),
		time.Date(2024, 1, 17, 12, 0, 0, 0, time.UTC),
	}

	tests := []struct {
		name           string
		body           interface{}
		setupMock      func(*MockScheduleService)
		expectedStatus int
		expectedBody   string
		checkResponse  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name: "successful preview",
			body: map[string]interface{}{
				"cron_expression": "0 0 * * *",
				"timezone":        "UTC",
				"count":           3,
			},
			setupMock: func(m *MockScheduleService) {
				m.On("GetNextRunTimes", "0 0 * * *", "UTC", 3).Return(fixedTimes, nil)
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, rr *httptest.ResponseRecorder) {
				var resp map[string]interface{}
				err := json.Unmarshal(rr.Body.Bytes(), &resp)
				require.NoError(t, err)
				assert.True(t, resp["valid"].(bool))
				assert.Equal(t, float64(3), resp["count"])
				assert.Equal(t, "UTC", resp["timezone"])
			},
		},
		{
			name: "successful preview with defaults",
			body: map[string]string{
				"cron_expression": "0 0 * * *",
			},
			setupMock: func(m *MockScheduleService) {
				// Default count is 10, default timezone is UTC
				m.On("GetNextRunTimes", "0 0 * * *", "UTC", 10).Return(fixedTimes, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "count capped at 50",
			body: map[string]interface{}{
				"cron_expression": "0 0 * * *",
				"timezone":        "UTC",
				"count":           100,
			},
			setupMock: func(m *MockScheduleService) {
				// Count should be capped at 50
				m.On("GetNextRunTimes", "0 0 * * *", "UTC", 50).Return(fixedTimes, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "invalid request body",
			body:           "invalid json",
			setupMock:      func(m *MockScheduleService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "invalid request body",
		},
		{
			name: "invalid cron expression",
			body: map[string]string{
				"cron_expression": "invalid cron",
			},
			setupMock: func(m *MockScheduleService) {
				m.On("GetNextRunTimes", "invalid cron", "UTC", 10).Return(nil, &schedule.ValidationError{Message: "invalid format"})
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "invalid cron expression",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, mockService := newTestScheduleHandler()
			tt.setupMock(mockService)

			var body []byte
			var err error
			switch v := tt.body.(type) {
			case string:
				body = []byte(v)
			default:
				body, err = json.Marshal(tt.body)
				require.NoError(t, err)
			}

			req := httptest.NewRequest(http.MethodPost, "/api/v1/schedules/preview", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()
			handler.PreviewSchedule(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			if tt.expectedBody != "" {
				assert.Contains(t, rr.Body.String(), tt.expectedBody)
			}
			if tt.checkResponse != nil {
				tt.checkResponse(t, rr)
			}
			mockService.AssertExpectations(t)
		})
	}
}

// ============================================================================
// ListExecutionHistory Handler Tests
// ============================================================================

func TestScheduleHandler_ListExecutionHistory(t *testing.T) {
	tests := []struct {
		name           string
		tenantID       string
		scheduleID     string
		queryParams    string
		setupMock      func(*MockScheduleService)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:        "successful list",
			tenantID:    "tenant-123",
			scheduleID:  "sched-123",
			queryParams: "",
			setupMock: func(m *MockScheduleService) {
				logs := []*schedule.ExecutionLog{createTestExecutionLog()}
				m.On("ListExecutionLogs", mock.Anything, "tenant-123", "sched-123", 20, 0).Return(logs, nil)
				m.On("CountExecutionLogs", mock.Anything, "tenant-123", "sched-123").Return(1, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:        "successful list with pagination",
			tenantID:    "tenant-123",
			scheduleID:  "sched-123",
			queryParams: "?limit=10&offset=5",
			setupMock: func(m *MockScheduleService) {
				logs := []*schedule.ExecutionLog{createTestExecutionLog()}
				m.On("ListExecutionLogs", mock.Anything, "tenant-123", "sched-123", 10, 5).Return(logs, nil)
				m.On("CountExecutionLogs", mock.Anything, "tenant-123", "sched-123").Return(10, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:        "schedule not found",
			tenantID:    "tenant-123",
			scheduleID:  "nonexistent",
			queryParams: "",
			setupMock: func(m *MockScheduleService) {
				m.On("ListExecutionLogs", mock.Anything, "tenant-123", "nonexistent", 20, 0).Return(nil, schedule.ErrNotFound)
			},
			expectedStatus: http.StatusNotFound,
			expectedBody:   "schedule not found",
		},
		{
			name:        "service error on list",
			tenantID:    "tenant-123",
			scheduleID:  "sched-123",
			queryParams: "",
			setupMock: func(m *MockScheduleService) {
				m.On("ListExecutionLogs", mock.Anything, "tenant-123", "sched-123", 20, 0).Return(nil, errors.New("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "failed to list execution history",
		},
		{
			name:        "count error does not fail request",
			tenantID:    "tenant-123",
			scheduleID:  "sched-123",
			queryParams: "",
			setupMock: func(m *MockScheduleService) {
				logs := []*schedule.ExecutionLog{createTestExecutionLog()}
				m.On("ListExecutionLogs", mock.Anything, "tenant-123", "sched-123", 20, 0).Return(logs, nil)
				m.On("CountExecutionLogs", mock.Anything, "tenant-123", "sched-123").Return(0, errors.New("count error"))
			},
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, mockService := newTestScheduleHandler()
			tt.setupMock(mockService)

			req := httptest.NewRequest(http.MethodGet, "/api/v1/schedules/"+tt.scheduleID+"/history"+tt.queryParams, nil)
			req = addScheduleContext(req, tt.tenantID, nil)
			req = addScheduleURLParams(req, map[string]string{"scheduleID": tt.scheduleID})

			rr := httptest.NewRecorder()
			handler.ListExecutionHistory(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			if tt.expectedBody != "" {
				assert.Contains(t, rr.Body.String(), tt.expectedBody)
			}
			mockService.AssertExpectations(t)
		})
	}
}

// ============================================================================
// GetExecutionLog Handler Tests
// ============================================================================

func TestScheduleHandler_GetExecutionLog(t *testing.T) {
	tests := []struct {
		name           string
		tenantID       string
		logID          string
		setupMock      func(*MockScheduleService)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:     "successful get",
			tenantID: "tenant-123",
			logID:    "log-123",
			setupMock: func(m *MockScheduleService) {
				m.On("GetExecutionLog", mock.Anything, "tenant-123", "log-123").Return(createTestExecutionLog(), nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:     "log not found",
			tenantID: "tenant-123",
			logID:    "nonexistent",
			setupMock: func(m *MockScheduleService) {
				m.On("GetExecutionLog", mock.Anything, "tenant-123", "nonexistent").Return(nil, schedule.ErrNotFound)
			},
			expectedStatus: http.StatusNotFound,
			expectedBody:   "execution log not found",
		},
		{
			name:     "service error",
			tenantID: "tenant-123",
			logID:    "log-123",
			setupMock: func(m *MockScheduleService) {
				m.On("GetExecutionLog", mock.Anything, "tenant-123", "log-123").Return(nil, errors.New("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "failed to get execution log",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, mockService := newTestScheduleHandler()
			tt.setupMock(mockService)

			req := httptest.NewRequest(http.MethodGet, "/api/v1/schedules/logs/"+tt.logID, nil)
			req = addScheduleContext(req, tt.tenantID, nil)
			req = addScheduleURLParams(req, map[string]string{"logID": tt.logID})

			rr := httptest.NewRecorder()
			handler.GetExecutionLog(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			if tt.expectedBody != "" {
				assert.Contains(t, rr.Body.String(), tt.expectedBody)
			}
			mockService.AssertExpectations(t)
		})
	}
}
