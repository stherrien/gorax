package handlers

import (
	"context"
	"encoding/json"
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

	"github.com/gorax/gorax/internal/workflow"
)

// MockWorkflowService is a mock implementation of workflow service for testing
type MockWorkflowService struct {
	mock.Mock
}

func (m *MockWorkflowService) ListExecutionsAdvanced(ctx context.Context, tenantID string, filter workflow.ExecutionFilter, cursor string, limit int) (*workflow.ExecutionListResult, error) {
	args := m.Called(ctx, tenantID, filter, cursor, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*workflow.ExecutionListResult), args.Error(1)
}

func (m *MockWorkflowService) GetExecutionWithSteps(ctx context.Context, tenantID, executionID string) (*workflow.ExecutionWithSteps, error) {
	args := m.Called(ctx, tenantID, executionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*workflow.ExecutionWithSteps), args.Error(1)
}

func (m *MockWorkflowService) GetExecutionStats(ctx context.Context, tenantID string, filter workflow.ExecutionFilter) (*workflow.ExecutionStats, error) {
	args := m.Called(ctx, tenantID, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*workflow.ExecutionStats), args.Error(1)
}

func newTestExecutionHandler() (*ExecutionHandler, *MockWorkflowService) {
	mockService := new(MockWorkflowService)
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	handler := NewExecutionHandler(mockService, logger)
	return handler, mockService
}

// addTenantContext adds tenant context to the request for testing

// TestListExecutionsAdvanced_Success tests successful execution listing
func TestListExecutionsAdvanced_Success(t *testing.T) {
	handler, mockService := newTestExecutionHandler()

	now := time.Now()
	executions := []*workflow.Execution{
		{
			ID:          "exec-1",
			TenantID:    "tenant-123",
			WorkflowID:  "workflow-1",
			Status:      "completed",
			TriggerType: "webhook",
			CreatedAt:   now,
		},
	}

	expectedResult := &workflow.ExecutionListResult{
		Data:       executions,
		Cursor:     "next-cursor",
		HasMore:    true,
		TotalCount: 10,
	}

	// Mock the service call (handler passes 0, service applies default 20)
	mockService.On("ListExecutionsAdvanced",
		mock.Anything,
		"tenant-123",
		workflow.ExecutionFilter{},
		"",
		0,
	).Return(expectedResult, nil)

	// Create request
	req := httptest.NewRequest(http.MethodGet, "/api/v1/executions", nil)
	req = addTenantContext(req, "tenant-123")
	w := httptest.NewRecorder()

	// Call handler
	handler.ListExecutionsAdvanced(w, req)

	// Assert response
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	assert.Contains(t, response, "data")
	assert.Contains(t, response, "cursor")
	assert.Contains(t, response, "has_more")
	assert.Contains(t, response, "total_count")

	mockService.AssertExpectations(t)
}

// TestListExecutionsAdvanced_WithFilters tests execution listing with query filters
func TestListExecutionsAdvanced_WithFilters(t *testing.T) {
	handler, mockService := newTestExecutionHandler()

	tests := []struct {
		name           string
		queryParams    string
		expectedFilter workflow.ExecutionFilter
		expectedLimit  int
		expectedCursor string
	}{
		{
			name:        "filter by workflow_id",
			queryParams: "?workflow_id=workflow-1",
			expectedFilter: workflow.ExecutionFilter{
				WorkflowID: "workflow-1",
			},
			expectedLimit:  0,
			expectedCursor: "",
		},
		{
			name:        "filter by status",
			queryParams: "?status=completed",
			expectedFilter: workflow.ExecutionFilter{
				Status: "completed",
			},
			expectedLimit:  0,
			expectedCursor: "",
		},
		{
			name:        "filter by trigger_type",
			queryParams: "?trigger_type=webhook",
			expectedFilter: workflow.ExecutionFilter{
				TriggerType: "webhook",
			},
			expectedLimit:  0,
			expectedCursor: "",
		},
		{
			name:           "with pagination",
			queryParams:    "?limit=50&cursor=abc123",
			expectedFilter: workflow.ExecutionFilter{},
			expectedLimit:  50,
			expectedCursor: "abc123",
		},
		{
			name:        "combined filters",
			queryParams: "?workflow_id=workflow-1&status=completed&limit=10",
			expectedFilter: workflow.ExecutionFilter{
				WorkflowID: "workflow-1",
				Status:     "completed",
			},
			expectedLimit:  10,
			expectedCursor: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expectedResult := &workflow.ExecutionListResult{
				Data:       []*workflow.Execution{},
				Cursor:     "",
				HasMore:    false,
				TotalCount: 0,
			}

			mockService.On("ListExecutionsAdvanced",
				mock.Anything,
				"tenant-123",
				tt.expectedFilter,
				tt.expectedCursor,
				tt.expectedLimit,
			).Return(expectedResult, nil).Once()

			req := httptest.NewRequest(http.MethodGet, "/api/v1/executions"+tt.queryParams, nil)
			req = addTenantContext(req, "tenant-123")
			w := httptest.NewRecorder()

			handler.ListExecutionsAdvanced(w, req)

			assert.Equal(t, http.StatusOK, w.Code)
			mockService.AssertExpectations(t)
		})
	}
}

// TestListExecutionsAdvanced_InvalidDateRange tests invalid date range handling
func TestListExecutionsAdvanced_InvalidDateRange(t *testing.T) {
	handler, mockService := newTestExecutionHandler()

	// Mock service to return validation error
	mockService.On("ListExecutionsAdvanced",
		mock.Anything,
		"tenant-123",
		mock.Anything,
		"",
		0,
	).Return(nil, &workflow.ValidationError{Message: "invalid filter: end_date must be after start_date"})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/executions?start_date=2024-01-10&end_date=2024-01-01", nil)
	req = addTenantContext(req, "tenant-123")
	w := httptest.NewRecorder()

	handler.ListExecutionsAdvanced(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]string
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)
	assert.Contains(t, response["error"], "invalid filter")
}

// TestGetExecutionWithSteps_Success tests successful retrieval of execution with steps
func TestGetExecutionWithSteps_Success(t *testing.T) {
	handler, mockService := newTestExecutionHandler()

	now := time.Now()
	expectedResult := &workflow.ExecutionWithSteps{
		Execution: &workflow.Execution{
			ID:          "exec-123",
			TenantID:    "tenant-123",
			WorkflowID:  "workflow-1",
			Status:      "completed",
			TriggerType: "webhook",
			CreatedAt:   now,
		},
		Steps: []*workflow.StepExecution{
			{
				ID:          "step-1",
				ExecutionID: "exec-123",
				NodeID:      "node-1",
				NodeType:    "action:http",
				Status:      "completed",
				StartedAt:   &now,
			},
		},
	}

	mockService.On("GetExecutionWithSteps",
		mock.Anything,
		"tenant-123",
		"exec-123",
	).Return(expectedResult, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/executions/exec-123/steps", nil)
	req = addTenantContext(req, "tenant-123")

	// Setup chi router context for URL params
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("executionID", "exec-123")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()

	handler.GetExecutionWithSteps(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	assert.Contains(t, response, "execution")
	assert.Contains(t, response, "steps")

	mockService.AssertExpectations(t)
}

// TestGetExecutionWithSteps_NotFound tests execution not found
func TestGetExecutionWithSteps_NotFound(t *testing.T) {
	handler, mockService := newTestExecutionHandler()

	mockService.On("GetExecutionWithSteps",
		mock.Anything,
		"tenant-123",
		"non-existent",
	).Return(nil, workflow.ErrNotFound)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/executions/non-existent/steps", nil)
	req = addTenantContext(req, "tenant-123")

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("executionID", "non-existent")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()

	handler.GetExecutionWithSteps(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var response map[string]string
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)
	assert.Contains(t, response["error"], "not found")

	mockService.AssertExpectations(t)
}

// TestGetExecutionStats_Success tests successful stats retrieval
func TestGetExecutionStats_Success(t *testing.T) {
	handler, mockService := newTestExecutionHandler()

	expectedStats := &workflow.ExecutionStats{
		TotalCount: 31,
		StatusCounts: map[string]int{
			"pending":   5,
			"running":   3,
			"completed": 20,
			"failed":    2,
			"cancelled": 1,
		},
	}

	mockService.On("GetExecutionStats",
		mock.Anything,
		"tenant-123",
		workflow.ExecutionFilter{},
	).Return(expectedStats, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/executions/stats", nil)
	req = addTenantContext(req, "tenant-123")
	w := httptest.NewRecorder()

	handler.GetExecutionStats(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response workflow.ExecutionStats
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	assert.Equal(t, 31, response.TotalCount)
	assert.Equal(t, 5, response.StatusCounts["pending"])
	assert.Equal(t, 20, response.StatusCounts["completed"])

	mockService.AssertExpectations(t)
}

// TestGetExecutionStats_WithFilters tests stats with filters
func TestGetExecutionStats_WithFilters(t *testing.T) {
	handler, mockService := newTestExecutionHandler()

	expectedStats := &workflow.ExecutionStats{
		TotalCount: 10,
		StatusCounts: map[string]int{
			"pending":   2,
			"running":   1,
			"completed": 7,
			"failed":    0,
			"cancelled": 0,
		},
	}

	mockService.On("GetExecutionStats",
		mock.Anything,
		"tenant-123",
		workflow.ExecutionFilter{
			WorkflowID: "workflow-1",
		},
	).Return(expectedStats, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/executions/stats?workflow_id=workflow-1", nil)
	req = addTenantContext(req, "tenant-123")
	w := httptest.NewRecorder()

	handler.GetExecutionStats(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response workflow.ExecutionStats
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	assert.Equal(t, 10, response.TotalCount)
	assert.Equal(t, 7, response.StatusCounts["completed"])

	mockService.AssertExpectations(t)
}

// TestMissingTenantID tests handler behavior when tenant ID is missing
func TestMissingTenantID(t *testing.T) {
	handler, _ := newTestExecutionHandler()

	tests := []struct {
		name    string
		handler http.HandlerFunc
		url     string
	}{
		{
			name:    "ListExecutionsAdvanced without tenant",
			handler: handler.ListExecutionsAdvanced,
			url:     "/api/v1/executions",
		},
		{
			name:    "GetExecutionStats without tenant",
			handler: handler.GetExecutionStats,
			url:     "/api/v1/executions/stats",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.url, nil)
			// No tenant ID in context
			w := httptest.NewRecorder()

			tt.handler(w, req)

			assert.Equal(t, http.StatusInternalServerError, w.Code)
		})
	}
}
