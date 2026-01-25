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

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gorax/gorax/internal/api/middleware"
	"github.com/gorax/gorax/internal/api/response"
	"github.com/gorax/gorax/internal/tenant"
	"github.com/gorax/gorax/internal/workflow"
)

// mockDryRunService is a simple mock that implements DryRun
type mockDryRunService struct {
	dryRunFunc func(ctx context.Context, tenantID, workflowID string, testData map[string]interface{}) (*workflow.DryRunResult, error)
}

func (m *mockDryRunService) DryRun(ctx context.Context, tenantID, workflowID string, testData map[string]interface{}) (*workflow.DryRunResult, error) {
	if m.dryRunFunc != nil {
		return m.dryRunFunc(ctx, tenantID, workflowID, testData)
	}
	return nil, errors.New("not implemented")
}

// workflowHandlerWithMock creates a handler with a mocked DryRun
type workflowHandlerWithMock struct {
	*WorkflowHandler
	mockService *mockDryRunService
}

func (h *workflowHandlerWithMock) DryRun(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r)
	workflowID := chi.URLParam(r, "workflowID")

	var input workflow.DryRunInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		_ = response.BadRequest(w, "invalid request body")
		return
	}

	result, err := h.mockService.DryRun(r.Context(), tenantID, workflowID, input.TestData)
	if err != nil {
		if err == workflow.ErrNotFound {
			_ = response.NotFound(w, "workflow not found")
			return
		}
		if _, ok := err.(*workflow.ValidationError); ok {
			_ = response.BadRequest(w, err.Error())
			return
		}
		_ = response.InternalError(w, "failed to perform dry-run")
		return
	}

	_ = response.OK(w, map[string]any{
		"data": result,
	})
}

func newTestWorkflowHandlerForDryRun(mockService *mockDryRunService) *workflowHandlerWithMock {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	baseHandler := &WorkflowHandler{
		service: nil,
		logger:  logger,
	}
	return &workflowHandlerWithMock{
		WorkflowHandler: baseHandler,
		mockService:     mockService,
	}
}

func addTenantContextForDryRun(req *http.Request, tenantID string) *http.Request {
	t := &tenant.Tenant{
		ID:     tenantID,
		Status: "active",
	}
	ctx := context.WithValue(req.Context(), middleware.TenantContextKey, t)
	return req.WithContext(ctx)
}

// TestDryRun_Success tests successful dry-run
func TestDryRun_Success(t *testing.T) {
	testData := map[string]interface{}{
		"payload": map[string]interface{}{
			"id":   "123",
			"name": "test",
		},
	}

	expectedResult := &workflow.DryRunResult{
		Valid:          true,
		ExecutionOrder: []string{"trigger-1", "http-1", "transform-1"},
		VariableMapping: map[string]string{
			"trigger.payload": "test_data",
			"steps.http-1":    "node:http-1",
		},
		Warnings: []workflow.DryRunWarning{},
		Errors:   []workflow.DryRunError{},
	}

	mockService := &mockDryRunService{
		dryRunFunc: func(ctx context.Context, tenantID, workflowID string, data map[string]interface{}) (*workflow.DryRunResult, error) {
			assert.Equal(t, "tenant-123", tenantID)
			assert.Equal(t, "workflow-123", workflowID)
			return expectedResult, nil
		},
	}

	handler := newTestWorkflowHandlerForDryRun(mockService)

	requestBody, _ := json.Marshal(map[string]interface{}{
		"test_data": testData,
	})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/workflows/workflow-123/dry-run", bytes.NewBuffer(requestBody))
	req = addTenantContextForDryRun(req, "tenant-123")

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("workflowID", "workflow-123")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()
	handler.DryRun(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	data, ok := response["data"].(map[string]interface{})
	require.True(t, ok)
	assert.True(t, data["valid"].(bool))
}

// TestDryRun_InvalidWorkflow tests dry-run with invalid workflow
func TestDryRun_InvalidWorkflow(t *testing.T) {
	expectedResult := &workflow.DryRunResult{
		Valid:          false,
		ExecutionOrder: []string{"trigger-1", "http-1"},
		VariableMapping: map[string]string{
			"trigger": "trigger_data",
		},
		Warnings: []workflow.DryRunWarning{},
		Errors: []workflow.DryRunError{
			{
				NodeID:  "http-1",
				Field:   "method",
				Message: "HTTP method is required",
			},
		},
	}

	mockService := &mockDryRunService{
		dryRunFunc: func(ctx context.Context, tenantID, workflowID string, data map[string]interface{}) (*workflow.DryRunResult, error) {
			return expectedResult, nil
		},
	}

	handler := newTestWorkflowHandlerForDryRun(mockService)

	requestBody, _ := json.Marshal(map[string]interface{}{})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/workflows/workflow-123/dry-run", bytes.NewBuffer(requestBody))
	req = addTenantContextForDryRun(req, "tenant-123")

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("workflowID", "workflow-123")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()
	handler.DryRun(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	data, ok := response["data"].(map[string]interface{})
	require.True(t, ok)
	assert.False(t, data["valid"].(bool))
}

// TestDryRun_WorkflowNotFound tests dry-run with non-existent workflow
func TestDryRun_WorkflowNotFound(t *testing.T) {
	mockService := &mockDryRunService{
		dryRunFunc: func(ctx context.Context, tenantID, workflowID string, data map[string]interface{}) (*workflow.DryRunResult, error) {
			return nil, workflow.ErrNotFound
		},
	}

	handler := newTestWorkflowHandlerForDryRun(mockService)

	requestBody, _ := json.Marshal(map[string]interface{}{})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/workflows/nonexistent/dry-run", bytes.NewBuffer(requestBody))
	req = addTenantContextForDryRun(req, "tenant-123")

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("workflowID", "nonexistent")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()
	handler.DryRun(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

// TestDryRun_ServiceError tests dry-run with service error
func TestDryRun_ServiceError(t *testing.T) {
	mockService := &mockDryRunService{
		dryRunFunc: func(ctx context.Context, tenantID, workflowID string, data map[string]interface{}) (*workflow.DryRunResult, error) {
			return nil, errors.New("internal service error")
		},
	}

	handler := newTestWorkflowHandlerForDryRun(mockService)

	requestBody, _ := json.Marshal(map[string]interface{}{})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/workflows/workflow-123/dry-run", bytes.NewBuffer(requestBody))
	req = addTenantContextForDryRun(req, "tenant-123")

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("workflowID", "workflow-123")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()
	handler.DryRun(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// TestDryRun_InvalidRequestBody tests dry-run with invalid request body
func TestDryRun_InvalidRequestBody(t *testing.T) {
	mockService := &mockDryRunService{}
	handler := newTestWorkflowHandlerForDryRun(mockService)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/workflows/workflow-123/dry-run", bytes.NewBuffer([]byte("invalid json")))
	req = addTenantContextForDryRun(req, "tenant-123")

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("workflowID", "workflow-123")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()
	handler.DryRun(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}
