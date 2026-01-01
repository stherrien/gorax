package handlers

import (
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

	"github.com/gorax/gorax/internal/analytics"
)

// MockAnalyticsService is a mock implementation of analytics service
type MockAnalyticsService struct {
	mock.Mock
}

func (m *MockAnalyticsService) GetWorkflowStats(ctx context.Context, tenantID, workflowID string, timeRange analytics.TimeRange) (*analytics.WorkflowStats, error) {
	args := m.Called(ctx, tenantID, workflowID, timeRange)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*analytics.WorkflowStats), args.Error(1)
}

func (m *MockAnalyticsService) GetTenantOverview(ctx context.Context, tenantID string, timeRange analytics.TimeRange) (*analytics.TenantOverview, error) {
	args := m.Called(ctx, tenantID, timeRange)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*analytics.TenantOverview), args.Error(1)
}

func (m *MockAnalyticsService) GetExecutionTrends(ctx context.Context, tenantID string, timeRange analytics.TimeRange, granularity analytics.Granularity) (*analytics.ExecutionTrends, error) {
	args := m.Called(ctx, tenantID, timeRange, granularity)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*analytics.ExecutionTrends), args.Error(1)
}

func (m *MockAnalyticsService) GetTopWorkflows(ctx context.Context, tenantID string, timeRange analytics.TimeRange, limit int) (*analytics.TopWorkflows, error) {
	args := m.Called(ctx, tenantID, timeRange, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*analytics.TopWorkflows), args.Error(1)
}

func (m *MockAnalyticsService) GetErrorBreakdown(ctx context.Context, tenantID string, timeRange analytics.TimeRange) (*analytics.ErrorBreakdown, error) {
	args := m.Called(ctx, tenantID, timeRange)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*analytics.ErrorBreakdown), args.Error(1)
}

func (m *MockAnalyticsService) GetNodePerformance(ctx context.Context, tenantID, workflowID string) (*analytics.NodePerformance, error) {
	args := m.Called(ctx, tenantID, workflowID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*analytics.NodePerformance), args.Error(1)
}

func newTestAnalyticsHandler() (*AnalyticsHandler, *MockAnalyticsService) {
	mockService := new(MockAnalyticsService)
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	handler := NewAnalyticsHandler(mockService, logger)
	return handler, mockService
}

// addTenantContext adds tenant context to request for testing

func TestGetTenantOverview_Success(t *testing.T) {
	handler, mockService := newTestAnalyticsHandler()

	expectedOverview := &analytics.TenantOverview{
		TotalExecutions:      500,
		SuccessfulExecutions: 450,
		FailedExecutions:     50,
		SuccessRate:          0.90,
		AvgDurationMs:        2000,
		ActiveWorkflows:      10,
		TotalWorkflows:       15,
	}

	mockService.On("GetTenantOverview",
		mock.Anything,
		"tenant-123",
		mock.AnythingOfType("analytics.TimeRange"),
	).Return(expectedOverview, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/analytics/overview?start_date=2024-01-01T00:00:00Z&end_date=2024-01-31T23:59:59Z", nil)
	req = addTenantContext(req, "tenant-123")
	w := httptest.NewRecorder()

	handler.GetTenantOverview(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response analytics.TenantOverview
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)
	assert.Equal(t, expectedOverview.TotalExecutions, response.TotalExecutions)
	assert.Equal(t, expectedOverview.SuccessRate, response.SuccessRate)
	mockService.AssertExpectations(t)
}

func TestGetTenantOverview_MissingDates(t *testing.T) {
	handler, _ := newTestAnalyticsHandler()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/analytics/overview", nil)
	req = addTenantContext(req, "tenant-123")
	w := httptest.NewRecorder()

	handler.GetTenantOverview(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetWorkflowStats_Success(t *testing.T) {
	handler, mockService := newTestAnalyticsHandler()

	expectedStats := &analytics.WorkflowStats{
		WorkflowID:     "workflow-1",
		WorkflowName:   "Test Workflow",
		ExecutionCount: 100,
		SuccessCount:   90,
		FailureCount:   10,
		SuccessRate:    0.90,
		AvgDurationMs:  1500,
	}

	mockService.On("GetWorkflowStats",
		mock.Anything,
		"tenant-123",
		"workflow-1",
		mock.AnythingOfType("analytics.TimeRange"),
	).Return(expectedStats, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/analytics/workflows/workflow-1?start_date=2024-01-01T00:00:00Z&end_date=2024-01-31T23:59:59Z", nil)
	req = addTenantContext(req, "tenant-123")

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("workflowID", "workflow-1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()

	handler.GetWorkflowStats(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response analytics.WorkflowStats
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)
	assert.Equal(t, "workflow-1", response.WorkflowID)
	assert.Equal(t, 100, response.ExecutionCount)
	mockService.AssertExpectations(t)
}

func TestGetExecutionTrends_Success(t *testing.T) {
	handler, mockService := newTestAnalyticsHandler()

	expectedTrends := &analytics.ExecutionTrends{
		Granularity: analytics.GranularityDay,
		StartDate:   time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		EndDate:     time.Date(2024, 1, 7, 23, 59, 59, 0, time.UTC),
		DataPoints: []analytics.TimeSeriesPoint{
			{
				Timestamp:      time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				ExecutionCount: 50,
				SuccessCount:   45,
				FailureCount:   5,
				SuccessRate:    0.90,
				AvgDurationMs:  1500,
			},
		},
	}

	mockService.On("GetExecutionTrends",
		mock.Anything,
		"tenant-123",
		mock.AnythingOfType("analytics.TimeRange"),
		analytics.GranularityDay,
	).Return(expectedTrends, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/analytics/trends?start_date=2024-01-01T00:00:00Z&end_date=2024-01-07T23:59:59Z&granularity=day", nil)
	req = addTenantContext(req, "tenant-123")
	w := httptest.NewRecorder()

	handler.GetExecutionTrends(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response analytics.ExecutionTrends
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)
	assert.Equal(t, analytics.GranularityDay, response.Granularity)
	assert.Len(t, response.DataPoints, 1)
	mockService.AssertExpectations(t)
}

func TestGetExecutionTrends_InvalidGranularity(t *testing.T) {
	handler, _ := newTestAnalyticsHandler()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/analytics/trends?start_date=2024-01-01T00:00:00Z&end_date=2024-01-07T23:59:59Z&granularity=invalid", nil)
	req = addTenantContext(req, "tenant-123")
	w := httptest.NewRecorder()

	handler.GetExecutionTrends(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetTopWorkflows_Success(t *testing.T) {
	handler, mockService := newTestAnalyticsHandler()

	expectedWorkflows := &analytics.TopWorkflows{
		Workflows: []analytics.TopWorkflow{
			{
				WorkflowID:     "workflow-1",
				WorkflowName:   "Top Workflow",
				ExecutionCount: 500,
				SuccessRate:    0.95,
				AvgDurationMs:  1000,
			},
		},
		Total: 1,
	}

	mockService.On("GetTopWorkflows",
		mock.Anything,
		"tenant-123",
		mock.AnythingOfType("analytics.TimeRange"),
		10,
	).Return(expectedWorkflows, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/analytics/top-workflows?start_date=2024-01-01T00:00:00Z&end_date=2024-01-31T23:59:59Z&limit=10", nil)
	req = addTenantContext(req, "tenant-123")
	w := httptest.NewRecorder()

	handler.GetTopWorkflows(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response analytics.TopWorkflows
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)
	assert.Len(t, response.Workflows, 1)
	assert.Equal(t, 500, response.Workflows[0].ExecutionCount)
	mockService.AssertExpectations(t)
}

func TestGetErrorBreakdown_Success(t *testing.T) {
	handler, mockService := newTestAnalyticsHandler()

	expectedBreakdown := &analytics.ErrorBreakdown{
		TotalErrors: 40,
		ErrorsByType: []analytics.ErrorInfo{
			{
				ErrorMessage: "Connection timeout",
				ErrorCount:   20,
				WorkflowID:   "workflow-1",
				WorkflowName: "Test Workflow",
				Percentage:   50.0,
			},
		},
	}

	mockService.On("GetErrorBreakdown",
		mock.Anything,
		"tenant-123",
		mock.AnythingOfType("analytics.TimeRange"),
	).Return(expectedBreakdown, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/analytics/errors?start_date=2024-01-01T00:00:00Z&end_date=2024-01-31T23:59:59Z", nil)
	req = addTenantContext(req, "tenant-123")
	w := httptest.NewRecorder()

	handler.GetErrorBreakdown(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response analytics.ErrorBreakdown
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)
	assert.Equal(t, 40, response.TotalErrors)
	assert.Len(t, response.ErrorsByType, 1)
	mockService.AssertExpectations(t)
}

func TestGetNodePerformance_Success(t *testing.T) {
	handler, mockService := newTestAnalyticsHandler()

	expectedPerformance := &analytics.NodePerformance{
		WorkflowID:   "workflow-1",
		WorkflowName: "Test Workflow",
		Nodes: []analytics.NodeStats{
			{
				NodeID:         "node-1",
				NodeType:       "action:http",
				ExecutionCount: 100,
				SuccessCount:   95,
				FailureCount:   5,
				SuccessRate:    0.95,
				AvgDurationMs:  500,
			},
		},
	}

	mockService.On("GetNodePerformance",
		mock.Anything,
		"tenant-123",
		"workflow-1",
	).Return(expectedPerformance, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/analytics/workflows/workflow-1/nodes", nil)
	req = addTenantContext(req, "tenant-123")

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("workflowID", "workflow-1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()

	handler.GetNodePerformance(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response analytics.NodePerformance
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)
	assert.Equal(t, "workflow-1", response.WorkflowID)
	assert.Len(t, response.Nodes, 1)
	mockService.AssertExpectations(t)
}

func TestGetNodePerformance_WorkflowNotFound(t *testing.T) {
	handler, mockService := newTestAnalyticsHandler()

	mockService.On("GetNodePerformance",
		mock.Anything,
		"tenant-123",
		"nonexistent",
	).Return(nil, analytics.ErrNotFound)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/analytics/workflows/nonexistent/nodes", nil)
	req = addTenantContext(req, "tenant-123")

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("workflowID", "nonexistent")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()

	handler.GetNodePerformance(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	mockService.AssertExpectations(t)
}

func TestGetTenantOverview_ServiceError(t *testing.T) {
	handler, mockService := newTestAnalyticsHandler()

	mockService.On("GetTenantOverview",
		mock.Anything,
		"tenant-123",
		mock.AnythingOfType("analytics.TimeRange"),
	).Return(nil, errors.New("database error"))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/analytics/overview?start_date=2024-01-01T00:00:00Z&end_date=2024-01-31T23:59:59Z", nil)
	req = addTenantContext(req, "tenant-123")
	w := httptest.NewRecorder()

	handler.GetTenantOverview(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockService.AssertExpectations(t)
}

func TestGetTenantOverview_NoTenantContext(t *testing.T) {
	handler, _ := newTestAnalyticsHandler()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/analytics/overview?start_date=2024-01-01T00:00:00Z&end_date=2024-01-31T23:59:59Z", nil)
	w := httptest.NewRecorder()

	handler.GetTenantOverview(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// Integration Tests - Full Request/Response Cycle

func TestAnalyticsIntegration_GetTenantOverview_FullCycle(t *testing.T) {
	t.Run("successful overview retrieval with all fields", func(t *testing.T) {
		handler, mockService := newTestAnalyticsHandler()

		expectedOverview := &analytics.TenantOverview{
			TotalExecutions:      1000,
			SuccessfulExecutions: 850,
			FailedExecutions:     150,
			SuccessRate:          0.85,
			AvgDurationMs:        2500,
			ActiveWorkflows:      25,
			TotalWorkflows:       30,
		}

		mockService.On("GetTenantOverview",
			mock.Anything,
			"tenant-integration-123",
			mock.MatchedBy(func(tr analytics.TimeRange) bool {
				return tr.StartDate.Year() == 2024 && tr.EndDate.Year() == 2024
			}),
		).Return(expectedOverview, nil)

		req := httptest.NewRequest(
			http.MethodGet,
			"/api/v1/analytics/overview?start_date=2024-01-01T00:00:00Z&end_date=2024-12-31T23:59:59Z",
			nil,
		)
		req = addTenantContext(req, "tenant-integration-123")
		w := httptest.NewRecorder()

		handler.GetTenantOverview(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

		var response analytics.TenantOverview
		err := json.NewDecoder(w.Body).Decode(&response)
		require.NoError(t, err)
		assert.Equal(t, 1000, response.TotalExecutions)
		assert.Equal(t, 850, response.SuccessfulExecutions)
		assert.Equal(t, 150, response.FailedExecutions)
		assert.Equal(t, 0.85, response.SuccessRate)
		assert.Equal(t, int64(2500), response.AvgDurationMs)
		assert.Equal(t, 25, response.ActiveWorkflows)
		assert.Equal(t, 30, response.TotalWorkflows)
		mockService.AssertExpectations(t)
	})

	t.Run("handles malformed date formats", func(t *testing.T) {
		handler, _ := newTestAnalyticsHandler()

		req := httptest.NewRequest(
			http.MethodGet,
			"/api/v1/analytics/overview?start_date=invalid-date&end_date=2024-12-31",
			nil,
		)
		req = addTenantContext(req, "tenant-123")
		w := httptest.NewRecorder()

		handler.GetTenantOverview(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		var response map[string]string
		json.NewDecoder(w.Body).Decode(&response)
		assert.Contains(t, response["error"], "invalid")
	})

	t.Run("handles missing start_date parameter", func(t *testing.T) {
		handler, _ := newTestAnalyticsHandler()

		req := httptest.NewRequest(
			http.MethodGet,
			"/api/v1/analytics/overview?end_date=2024-12-31T23:59:59Z",
			nil,
		)
		req = addTenantContext(req, "tenant-123")
		w := httptest.NewRecorder()

		handler.GetTenantOverview(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestAnalyticsIntegration_GetExecutionTrends_AllGranularities(t *testing.T) {
	tests := []struct {
		name        string
		granularity string
		expected    analytics.Granularity
	}{
		{
			name:        "hourly granularity",
			granularity: "hour",
			expected:    analytics.GranularityHour,
		},
		{
			name:        "daily granularity",
			granularity: "day",
			expected:    analytics.GranularityDay,
		},
		{
			name:        "weekly granularity",
			granularity: "week",
			expected:    analytics.GranularityWeek,
		},
		{
			name:        "monthly granularity",
			granularity: "month",
			expected:    analytics.GranularityMonth,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, mockService := newTestAnalyticsHandler()

			expectedTrends := &analytics.ExecutionTrends{
				Granularity: tt.expected,
				StartDate:   time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				EndDate:     time.Date(2024, 1, 31, 23, 59, 59, 0, time.UTC),
				DataPoints: []analytics.TimeSeriesPoint{
					{
						Timestamp:      time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
						ExecutionCount: 100,
						SuccessCount:   90,
						FailureCount:   10,
						SuccessRate:    0.90,
						AvgDurationMs:  1500,
					},
					{
						Timestamp:      time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
						ExecutionCount: 120,
						SuccessCount:   110,
						FailureCount:   10,
						SuccessRate:    0.92,
						AvgDurationMs:  1400,
					},
				},
			}

			mockService.On("GetExecutionTrends",
				mock.Anything,
				"tenant-123",
				mock.AnythingOfType("analytics.TimeRange"),
				tt.expected,
			).Return(expectedTrends, nil)

			req := httptest.NewRequest(
				http.MethodGet,
				"/api/v1/analytics/trends?start_date=2024-01-01T00:00:00Z&end_date=2024-01-31T23:59:59Z&granularity="+tt.granularity,
				nil,
			)
			req = addTenantContext(req, "tenant-123")
			w := httptest.NewRecorder()

			handler.GetExecutionTrends(w, req)

			assert.Equal(t, http.StatusOK, w.Code)

			var response analytics.ExecutionTrends
			err := json.NewDecoder(w.Body).Decode(&response)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, response.Granularity)
			assert.Len(t, response.DataPoints, 2)
			assert.Equal(t, 100, response.DataPoints[0].ExecutionCount)
			assert.Equal(t, 120, response.DataPoints[1].ExecutionCount)
			mockService.AssertExpectations(t)
		})
	}
}

func TestAnalyticsIntegration_GetTopWorkflows_WithPagination(t *testing.T) {
	t.Run("retrieves top 5 workflows", func(t *testing.T) {
		handler, mockService := newTestAnalyticsHandler()

		expectedWorkflows := &analytics.TopWorkflows{
			Workflows: []analytics.TopWorkflow{
				{WorkflowID: "wf-1", WorkflowName: "Top 1", ExecutionCount: 1000, SuccessRate: 0.99, AvgDurationMs: 800},
				{WorkflowID: "wf-2", WorkflowName: "Top 2", ExecutionCount: 900, SuccessRate: 0.98, AvgDurationMs: 850},
				{WorkflowID: "wf-3", WorkflowName: "Top 3", ExecutionCount: 800, SuccessRate: 0.97, AvgDurationMs: 900},
				{WorkflowID: "wf-4", WorkflowName: "Top 4", ExecutionCount: 700, SuccessRate: 0.96, AvgDurationMs: 950},
				{WorkflowID: "wf-5", WorkflowName: "Top 5", ExecutionCount: 600, SuccessRate: 0.95, AvgDurationMs: 1000},
			},
			Total: 5,
		}

		mockService.On("GetTopWorkflows",
			mock.Anything,
			"tenant-123",
			mock.AnythingOfType("analytics.TimeRange"),
			5,
		).Return(expectedWorkflows, nil)

		req := httptest.NewRequest(
			http.MethodGet,
			"/api/v1/analytics/top-workflows?start_date=2024-01-01T00:00:00Z&end_date=2024-01-31T23:59:59Z&limit=5",
			nil,
		)
		req = addTenantContext(req, "tenant-123")
		w := httptest.NewRecorder()

		handler.GetTopWorkflows(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response analytics.TopWorkflows
		err := json.NewDecoder(w.Body).Decode(&response)
		require.NoError(t, err)
		assert.Len(t, response.Workflows, 5)
		assert.Equal(t, 5, response.Total)
		assert.Equal(t, 1000, response.Workflows[0].ExecutionCount)
		mockService.AssertExpectations(t)
	})

	t.Run("defaults to limit 10 when not specified", func(t *testing.T) {
		handler, mockService := newTestAnalyticsHandler()

		mockService.On("GetTopWorkflows",
			mock.Anything,
			"tenant-123",
			mock.AnythingOfType("analytics.TimeRange"),
			10,
		).Return(&analytics.TopWorkflows{Workflows: []analytics.TopWorkflow{}, Total: 0}, nil)

		req := httptest.NewRequest(
			http.MethodGet,
			"/api/v1/analytics/top-workflows?start_date=2024-01-01T00:00:00Z&end_date=2024-01-31T23:59:59Z",
			nil,
		)
		req = addTenantContext(req, "tenant-123")
		w := httptest.NewRecorder()

		handler.GetTopWorkflows(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		mockService.AssertExpectations(t)
	})
}

func TestAnalyticsIntegration_GetErrorBreakdown_DetailedErrors(t *testing.T) {
	handler, mockService := newTestAnalyticsHandler()

	expectedBreakdown := &analytics.ErrorBreakdown{
		TotalErrors: 250,
		ErrorsByType: []analytics.ErrorInfo{
			{
				ErrorMessage: "Connection timeout",
				ErrorCount:   100,
				WorkflowID:   "wf-1",
				WorkflowName: "HTTP Workflow",
				Percentage:   40.0,
			},
			{
				ErrorMessage: "Invalid credentials",
				ErrorCount:   80,
				WorkflowID:   "wf-2",
				WorkflowName: "Auth Workflow",
				Percentage:   32.0,
			},
			{
				ErrorMessage: "Rate limit exceeded",
				ErrorCount:   70,
				WorkflowID:   "wf-3",
				WorkflowName: "API Workflow",
				Percentage:   28.0,
			},
		},
	}

	mockService.On("GetErrorBreakdown",
		mock.Anything,
		"tenant-123",
		mock.AnythingOfType("analytics.TimeRange"),
	).Return(expectedBreakdown, nil)

	req := httptest.NewRequest(
		http.MethodGet,
		"/api/v1/analytics/errors?start_date=2024-01-01T00:00:00Z&end_date=2024-01-31T23:59:59Z",
		nil,
	)
	req = addTenantContext(req, "tenant-123")
	w := httptest.NewRecorder()

	handler.GetErrorBreakdown(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response analytics.ErrorBreakdown
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)
	assert.Equal(t, 250, response.TotalErrors)
	assert.Len(t, response.ErrorsByType, 3)
	assert.Equal(t, "Connection timeout", response.ErrorsByType[0].ErrorMessage)
	assert.Equal(t, 100, response.ErrorsByType[0].ErrorCount)
	assert.Equal(t, 40.0, response.ErrorsByType[0].Percentage)
	mockService.AssertExpectations(t)
}

func TestAnalyticsIntegration_GetNodePerformance_ComplexWorkflow(t *testing.T) {
	t.Run("retrieves performance for multi-node workflow", func(t *testing.T) {
		handler, mockService := newTestAnalyticsHandler()

		expectedPerformance := &analytics.NodePerformance{
			WorkflowID:   "wf-complex-123",
			WorkflowName: "Complex Multi-Step Workflow",
			Nodes: []analytics.NodeStats{
				{
					NodeID:         "node-trigger",
					NodeType:       "trigger:webhook",
					ExecutionCount: 1000,
					SuccessCount:   1000,
					FailureCount:   0,
					SuccessRate:    1.0,
					AvgDurationMs:  50,
				},
				{
					NodeID:         "node-http",
					NodeType:       "action:http",
					ExecutionCount: 1000,
					SuccessCount:   950,
					FailureCount:   50,
					SuccessRate:    0.95,
					AvgDurationMs:  1200,
				},
				{
					NodeID:         "node-transform",
					NodeType:       "action:transform",
					ExecutionCount: 950,
					SuccessCount:   950,
					FailureCount:   0,
					SuccessRate:    1.0,
					AvgDurationMs:  100,
				},
			},
		}

		mockService.On("GetNodePerformance",
			mock.Anything,
			"tenant-123",
			"wf-complex-123",
		).Return(expectedPerformance, nil)

		req := httptest.NewRequest(
			http.MethodGet,
			"/api/v1/analytics/workflows/wf-complex-123/nodes",
			nil,
		)
		req = addTenantContext(req, "tenant-123")

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("workflowID", "wf-complex-123")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()

		handler.GetNodePerformance(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response analytics.NodePerformance
		err := json.NewDecoder(w.Body).Decode(&response)
		require.NoError(t, err)
		assert.Equal(t, "wf-complex-123", response.WorkflowID)
		assert.Equal(t, "Complex Multi-Step Workflow", response.WorkflowName)
		assert.Len(t, response.Nodes, 3)
		assert.Equal(t, "node-trigger", response.Nodes[0].NodeID)
		assert.Equal(t, 1.0, response.Nodes[0].SuccessRate)
		assert.Equal(t, 0.95, response.Nodes[1].SuccessRate)
		mockService.AssertExpectations(t)
	})

	t.Run("handles workflow with no executions", func(t *testing.T) {
		handler, mockService := newTestAnalyticsHandler()

		expectedPerformance := &analytics.NodePerformance{
			WorkflowID:   "wf-no-exec",
			WorkflowName: "Unused Workflow",
			Nodes:        []analytics.NodeStats{},
		}

		mockService.On("GetNodePerformance",
			mock.Anything,
			"tenant-123",
			"wf-no-exec",
		).Return(expectedPerformance, nil)

		req := httptest.NewRequest(
			http.MethodGet,
			"/api/v1/analytics/workflows/wf-no-exec/nodes",
			nil,
		)
		req = addTenantContext(req, "tenant-123")

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("workflowID", "wf-no-exec")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()

		handler.GetNodePerformance(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response analytics.NodePerformance
		err := json.NewDecoder(w.Body).Decode(&response)
		require.NoError(t, err)
		assert.Len(t, response.Nodes, 0)
		mockService.AssertExpectations(t)
	})
}
