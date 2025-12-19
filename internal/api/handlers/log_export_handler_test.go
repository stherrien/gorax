package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/gorax/gorax/internal/api/middleware"
	"github.com/gorax/gorax/internal/tenant"
	"github.com/gorax/gorax/internal/workflow"
)

// MockLogExportService is a mock for the log export service
type MockLogExportService struct {
	mock.Mock
}

func (m *MockLogExportService) GetExecutionWithSteps(ctx context.Context, tenantID, executionID string) (*workflow.Execution, []*workflow.StepExecution, error) {
	args := m.Called(ctx, tenantID, executionID)
	if args.Get(0) == nil {
		return nil, nil, args.Error(2)
	}
	return args.Get(0).(*workflow.Execution), args.Get(1).([]*workflow.StepExecution), args.Error(2)
}

func (m *MockLogExportService) ExportLogs(execution *workflow.Execution, steps []*workflow.StepExecution, format string) ([]byte, string, error) {
	args := m.Called(execution, steps, format)
	if args.Get(0) == nil {
		return nil, "", args.Error(2)
	}
	return args.Get(0).([]byte), args.String(1), args.Error(2)
}

func TestLogExportHandler_ExportExecutionLogs(t *testing.T) {
	now := time.Now()

	execution := &workflow.Execution{
		ID:         "exec-123",
		WorkflowID: "wf-456",
		TenantID:   "tenant-1",
		Status:     "completed",
		StartedAt:  &now,
	}

	steps := []*workflow.StepExecution{
		{
			ID:         "step-1",
			NodeID:     "node-1",
			NodeType:   "action:http",
			Status:     "completed",
			StartedAt:  &now,
			DurationMs: ptrInt(1000),
		},
	}

	t.Run("exports logs in TXT format", func(t *testing.T) {
		mockService := new(MockLogExportService)
		handler := NewLogExportHandler(mockService, nil)

		mockService.On("GetExecutionWithSteps", mock.Anything, "tenant-1", "exec-123").
			Return(execution, steps, nil)
		mockService.On("ExportLogs", execution, steps, "txt").
			Return([]byte("Log content"), "text/plain", nil)

		req := createLogExportRequest(t, "exec-123", "txt")
		rr := httptest.NewRecorder()

		handler.ExportExecutionLogs(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Equal(t, "text/plain", rr.Header().Get("Content-Type"))
		assert.Contains(t, rr.Header().Get("Content-Disposition"), "attachment")
		assert.Contains(t, rr.Header().Get("Content-Disposition"), "exec-123.txt")
		assert.Equal(t, "Log content", rr.Body.String())

		mockService.AssertExpectations(t)
	})

	t.Run("exports logs in JSON format", func(t *testing.T) {
		mockService := new(MockLogExportService)
		handler := NewLogExportHandler(mockService, nil)

		jsonData := []byte(`{"execution_id":"exec-123"}`)
		mockService.On("GetExecutionWithSteps", mock.Anything, "tenant-1", "exec-123").
			Return(execution, steps, nil)
		mockService.On("ExportLogs", execution, steps, "json").
			Return(jsonData, "application/json", nil)

		req := createLogExportRequest(t, "exec-123", "json")
		rr := httptest.NewRecorder()

		handler.ExportExecutionLogs(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))
		assert.Contains(t, rr.Header().Get("Content-Disposition"), "exec-123.json")

		mockService.AssertExpectations(t)
	})

	t.Run("exports logs in CSV format", func(t *testing.T) {
		mockService := new(MockLogExportService)
		handler := NewLogExportHandler(mockService, nil)

		csvData := []byte("step_id,node_id,status\nstep-1,node-1,completed")
		mockService.On("GetExecutionWithSteps", mock.Anything, "tenant-1", "exec-123").
			Return(execution, steps, nil)
		mockService.On("ExportLogs", execution, steps, "csv").
			Return(csvData, "text/csv", nil)

		req := createLogExportRequest(t, "exec-123", "csv")
		rr := httptest.NewRecorder()

		handler.ExportExecutionLogs(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Equal(t, "text/csv", rr.Header().Get("Content-Type"))
		assert.Contains(t, rr.Header().Get("Content-Disposition"), "exec-123.csv")

		mockService.AssertExpectations(t)
	})

	t.Run("defaults to TXT format when format not specified", func(t *testing.T) {
		mockService := new(MockLogExportService)
		handler := NewLogExportHandler(mockService, nil)

		mockService.On("GetExecutionWithSteps", mock.Anything, "tenant-1", "exec-123").
			Return(execution, steps, nil)
		mockService.On("ExportLogs", execution, steps, "txt").
			Return([]byte("Log content"), "text/plain", nil)

		req := createLogExportRequestNoFormat(t, "exec-123")
		rr := httptest.NewRecorder()

		handler.ExportExecutionLogs(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		mockService.AssertExpectations(t)
	})

	t.Run("returns 400 for invalid format", func(t *testing.T) {
		mockService := new(MockLogExportService)
		handler := NewLogExportHandler(mockService, nil)

		req := createLogExportRequest(t, "exec-123", "invalid")
		rr := httptest.NewRecorder()

		handler.ExportExecutionLogs(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)

		var response map[string]string
		err := json.Unmarshal(rr.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Contains(t, response["error"], "invalid format")
	})

	t.Run("returns 404 when execution not found", func(t *testing.T) {
		mockService := new(MockLogExportService)
		handler := NewLogExportHandler(mockService, nil)

		mockService.On("GetExecutionWithSteps", mock.Anything, "tenant-1", "exec-999").
			Return(nil, nil, errors.New("execution not found"))

		req := createLogExportRequest(t, "exec-999", "txt")
		rr := httptest.NewRecorder()

		handler.ExportExecutionLogs(rr, req)

		assert.Equal(t, http.StatusNotFound, rr.Code)
		mockService.AssertExpectations(t)
	})

	t.Run("returns 500 on export error", func(t *testing.T) {
		mockService := new(MockLogExportService)
		handler := NewLogExportHandler(mockService, nil)

		mockService.On("GetExecutionWithSteps", mock.Anything, "tenant-1", "exec-123").
			Return(execution, steps, nil)
		mockService.On("ExportLogs", execution, steps, "txt").
			Return(nil, "", errors.New("export failed"))

		req := createLogExportRequest(t, "exec-123", "txt")
		rr := httptest.NewRecorder()

		handler.ExportExecutionLogs(rr, req)

		assert.Equal(t, http.StatusInternalServerError, rr.Code)
		mockService.AssertExpectations(t)
	})
}

func TestLogExportHandler_ContentDisposition(t *testing.T) {
	tests := []struct {
		name           string
		executionID    string
		format         string
		expectedSuffix string
	}{
		{"TXT format", "exec-123", "txt", "exec-123.txt"},
		{"JSON format", "exec-456", "json", "exec-456.json"},
		{"CSV format", "exec-789", "csv", "exec-789.csv"},
		{"long execution ID", "exec-abc-def-123-456", "txt", "exec-abc-def-123-456.txt"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockLogExportService)
			handler := NewLogExportHandler(mockService, nil)

			execution := &workflow.Execution{
				ID:       tt.executionID,
				TenantID: "tenant-1",
				Status:   "completed",
			}

			mockService.On("GetExecutionWithSteps", mock.Anything, "tenant-1", tt.executionID).
				Return(execution, []*workflow.StepExecution{}, nil)
			mockService.On("ExportLogs", execution, mock.Anything, tt.format).
				Return([]byte("content"), "text/plain", nil)

			req := createLogExportRequest(t, tt.executionID, tt.format)
			rr := httptest.NewRecorder()

			handler.ExportExecutionLogs(rr, req)

			assert.Equal(t, http.StatusOK, rr.Code)
			assert.Contains(t, rr.Header().Get("Content-Disposition"), tt.expectedSuffix)
		})
	}
}

// Helper functions
func createLogExportRequest(t *testing.T, executionID, format string) *http.Request {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/executions/"+executionID+"/logs/export?format="+format, nil)
	req = addLogExportTenantCtx(req)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", executionID)
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	return req
}

func createLogExportRequestNoFormat(t *testing.T, executionID string) *http.Request {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/executions/"+executionID+"/logs/export", nil)
	req = addLogExportTenantCtx(req)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", executionID)
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	return req
}

func addLogExportTenantCtx(req *http.Request) *http.Request {
	t := &tenant.Tenant{
		ID:     "tenant-1",
		Status: "active",
	}
	ctx := context.WithValue(req.Context(), middleware.TenantContextKey, t)
	return req.WithContext(ctx)
}

func ptrInt(i int) *int {
	return &i
}
