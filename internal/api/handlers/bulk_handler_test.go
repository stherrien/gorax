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
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/gorax/gorax/internal/api/middleware"
	"github.com/gorax/gorax/internal/schedule"
	"github.com/gorax/gorax/internal/tenant"
)

var testLogger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))

// setTestTenantID adds tenant context to the request for testing
func setTestTenantID(req *http.Request, tenantID string) *http.Request {
	t := &tenant.Tenant{
		ID:     tenantID,
		Status: "active",
	}
	ctx := context.WithValue(req.Context(), middleware.TenantContextKey, t)
	return req.WithContext(ctx)
}

// MockBulkScheduleService mocks schedule bulk operations
type MockBulkScheduleService struct {
	mock.Mock
}

func (m *MockBulkScheduleService) BulkUpdate(ctx context.Context, tenantID string, ids []string, enabled bool) ([]string, []BulkOperationError) {
	args := m.Called(ctx, tenantID, ids, enabled)
	return args.Get(0).([]string), args.Get(1).([]BulkOperationError)
}

func (m *MockBulkScheduleService) BulkDelete(ctx context.Context, tenantID string, ids []string) ([]string, []BulkOperationError) {
	args := m.Called(ctx, tenantID, ids)
	return args.Get(0).([]string), args.Get(1).([]BulkOperationError)
}

func (m *MockBulkScheduleService) ExportSchedules(ctx context.Context, tenantID string, ids []string) ([]*schedule.Schedule, error) {
	args := m.Called(ctx, tenantID, ids)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*schedule.Schedule), args.Error(1)
}

// MockBulkExecutionService mocks execution bulk operations
type MockBulkExecutionService struct {
	mock.Mock
}

func (m *MockBulkExecutionService) BulkDelete(ctx context.Context, tenantID string, ids []string) ([]string, []BulkOperationError) {
	args := m.Called(ctx, tenantID, ids)
	return args.Get(0).([]string), args.Get(1).([]BulkOperationError)
}

func (m *MockBulkExecutionService) BulkRetry(ctx context.Context, tenantID string, ids []string) ([]string, []BulkOperationError) {
	args := m.Called(ctx, tenantID, ids)
	return args.Get(0).([]string), args.Get(1).([]BulkOperationError)
}

func TestBulkUpdateSchedules(t *testing.T) {
	tests := []struct {
		name           string
		tenantID       string
		requestBody    interface{}
		mockSuccess    []string
		mockFailed     []BulkOperationError
		expectedStatus int
		expectedError  string
	}{
		{
			name:     "successful bulk enable",
			tenantID: "tenant1",
			requestBody: map[string]interface{}{
				"ids":    []string{"sched1", "sched2"},
				"action": "enable",
			},
			mockSuccess:    []string{"sched1", "sched2"},
			mockFailed:     []BulkOperationError{},
			expectedStatus: http.StatusOK,
		},
		{
			name:     "partial success bulk disable",
			tenantID: "tenant1",
			requestBody: map[string]interface{}{
				"ids":    []string{"sched1", "sched2", "sched3"},
				"action": "disable",
			},
			mockSuccess: []string{"sched1", "sched2"},
			mockFailed: []BulkOperationError{
				{ID: "sched3", Error: "schedule not found"},
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:     "invalid action",
			tenantID: "tenant1",
			requestBody: map[string]interface{}{
				"ids":    []string{"sched1"},
				"action": "invalid",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "invalid action",
		},
		{
			name:     "empty ids",
			tenantID: "tenant1",
			requestBody: map[string]interface{}{
				"ids":    []string{},
				"action": "enable",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "at least one schedule ID is required",
		},
		{
			name:     "missing action",
			tenantID: "tenant1",
			requestBody: map[string]interface{}{
				"ids": []string{"sched1"},
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "action is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockBulkScheduleService)
			handler := NewBulkHandler(mockService, nil, testLogger)

			// Set up mock expectations if valid request
			if tt.expectedStatus == http.StatusOK {
				enabled := tt.requestBody.(map[string]interface{})["action"] == "enable"
				ids := tt.requestBody.(map[string]interface{})["ids"].([]string)
				mockService.On("BulkUpdate", mock.Anything, tt.tenantID, ids, enabled).
					Return(tt.mockSuccess, tt.mockFailed)
			}

			body, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest(http.MethodPatch, "/api/v1/schedules/bulk", bytes.NewReader(body))
			req = setTestTenantID(req, tt.tenantID)
			w := httptest.NewRecorder()

			handler.BulkUpdateSchedules(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedError != "" {
				var response map[string]interface{}
				json.NewDecoder(w.Body).Decode(&response)
				assert.Contains(t, response["error"], tt.expectedError)
			} else {
				var response BulkOperationResult
				json.NewDecoder(w.Body).Decode(&response)
				assert.Equal(t, tt.mockSuccess, response.Success)
				assert.Equal(t, tt.mockFailed, response.Failed)
			}

			mockService.AssertExpectations(t)
		})
	}
}

func TestBulkDeleteSchedules(t *testing.T) {
	tests := []struct {
		name           string
		tenantID       string
		requestBody    interface{}
		mockSuccess    []string
		mockFailed     []BulkOperationError
		expectedStatus int
		expectedError  string
	}{
		{
			name:     "successful bulk delete",
			tenantID: "tenant1",
			requestBody: map[string]interface{}{
				"ids":    []string{"sched1", "sched2"},
				"action": "delete",
			},
			mockSuccess:    []string{"sched1", "sched2"},
			mockFailed:     []BulkOperationError{},
			expectedStatus: http.StatusOK,
		},
		{
			name:     "partial failure",
			tenantID: "tenant1",
			requestBody: map[string]interface{}{
				"ids":    []string{"sched1", "sched2"},
				"action": "delete",
			},
			mockSuccess: []string{"sched1"},
			mockFailed: []BulkOperationError{
				{ID: "sched2", Error: "schedule not found"},
			},
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockBulkScheduleService)
			handler := NewBulkHandler(mockService, nil, testLogger)

			ids := tt.requestBody.(map[string]interface{})["ids"].([]string)
			mockService.On("BulkDelete", mock.Anything, tt.tenantID, ids).
				Return(tt.mockSuccess, tt.mockFailed)

			body, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest(http.MethodPatch, "/api/v1/schedules/bulk", bytes.NewReader(body))
			req = setTestTenantID(req, tt.tenantID)
			w := httptest.NewRecorder()

			handler.BulkUpdateSchedules(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			mockService.AssertExpectations(t)
		})
	}
}

func TestBulkDeleteExecutions(t *testing.T) {
	tests := []struct {
		name           string
		tenantID       string
		ids            []string
		mockSuccess    []string
		mockFailed     []BulkOperationError
		expectedStatus int
	}{
		{
			name:           "successful bulk delete",
			tenantID:       "tenant1",
			ids:            []string{"exec1", "exec2"},
			mockSuccess:    []string{"exec1", "exec2"},
			mockFailed:     []BulkOperationError{},
			expectedStatus: http.StatusOK,
		},
		{
			name:        "partial failure",
			tenantID:    "tenant1",
			ids:         []string{"exec1", "exec2"},
			mockSuccess: []string{"exec1"},
			mockFailed: []BulkOperationError{
				{ID: "exec2", Error: "execution not found"},
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "empty ids",
			tenantID:       "tenant1",
			ids:            []string{},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockBulkExecutionService)
			handler := NewBulkHandler(nil, mockService, testLogger)

			if tt.expectedStatus == http.StatusOK {
				mockService.On("BulkDelete", mock.Anything, tt.tenantID, tt.ids).
					Return(tt.mockSuccess, tt.mockFailed)
			}

			body, _ := json.Marshal(map[string]interface{}{"ids": tt.ids})
			req := httptest.NewRequest(http.MethodDelete, "/api/v1/executions/bulk", bytes.NewReader(body))
			req = setTestTenantID(req, tt.tenantID)
			w := httptest.NewRecorder()

			handler.BulkDeleteExecutions(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			mockService.AssertExpectations(t)
		})
	}
}

func TestBulkRetryExecutions(t *testing.T) {
	tests := []struct {
		name           string
		tenantID       string
		ids            []string
		mockSuccess    []string
		mockFailed     []BulkOperationError
		expectedStatus int
	}{
		{
			name:           "successful bulk retry",
			tenantID:       "tenant1",
			ids:            []string{"exec1", "exec2"},
			mockSuccess:    []string{"exec1", "exec2"},
			mockFailed:     []BulkOperationError{},
			expectedStatus: http.StatusOK,
		},
		{
			name:        "partial failure",
			tenantID:    "tenant1",
			ids:         []string{"exec1", "exec2"},
			mockSuccess: []string{"exec1"},
			mockFailed: []BulkOperationError{
				{ID: "exec2", Error: "execution not in failed state"},
			},
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockBulkExecutionService)
			handler := NewBulkHandler(nil, mockService, testLogger)

			mockService.On("BulkRetry", mock.Anything, tt.tenantID, tt.ids).
				Return(tt.mockSuccess, tt.mockFailed)

			body, _ := json.Marshal(map[string]interface{}{"ids": tt.ids})
			req := httptest.NewRequest(http.MethodPost, "/api/v1/executions/bulk/retry", bytes.NewReader(body))
			req = setTestTenantID(req, tt.tenantID)
			w := httptest.NewRecorder()

			handler.BulkRetryExecutions(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			mockService.AssertExpectations(t)
		})
	}
}

func TestExportSchedules(t *testing.T) {
	tests := []struct {
		name           string
		tenantID       string
		queryIDs       string
		mockSchedules  []*schedule.Schedule
		mockError      error
		expectedStatus int
	}{
		{
			name:     "export specific schedules",
			tenantID: "tenant1",
			queryIDs: "sched1,sched2",
			mockSchedules: []*schedule.Schedule{
				{ID: "sched1", Name: "Schedule 1"},
				{ID: "sched2", Name: "Schedule 2"},
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "export all schedules",
			tenantID:       "tenant1",
			queryIDs:       "",
			mockSchedules:  []*schedule.Schedule{{ID: "sched1"}},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "service error",
			tenantID:       "tenant1",
			queryIDs:       "sched1",
			mockError:      errors.New("database error"),
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockBulkScheduleService)
			handler := NewBulkHandler(mockService, nil, testLogger)

			// Parse the query IDs to match what the handler will parse
			var ids []string
			if tt.queryIDs != "" {
				ids = strings.Split(tt.queryIDs, ",")
			}

			mockService.On("ExportSchedules", mock.Anything, tt.tenantID, ids).
				Return(tt.mockSchedules, tt.mockError)

			req := httptest.NewRequest(http.MethodGet, "/api/v1/schedules/export?ids="+tt.queryIDs, nil)
			req = setTestTenantID(req, tt.tenantID)
			w := httptest.NewRecorder()

			handler.ExportSchedules(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedStatus == http.StatusOK {
				var exported []schedule.Schedule
				json.NewDecoder(w.Body).Decode(&exported)
				assert.Equal(t, len(tt.mockSchedules), len(exported))
			}

			mockService.AssertExpectations(t)
		})
	}
}
