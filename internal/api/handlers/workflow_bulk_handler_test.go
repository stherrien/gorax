package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/gorax/gorax/internal/api/middleware"
	"github.com/gorax/gorax/internal/tenant"
	"github.com/gorax/gorax/internal/workflow"
)

// MockWorkflowBulkService is a mock implementation for testing
type MockWorkflowBulkService struct {
	mock.Mock
}

func (m *MockWorkflowBulkService) BulkDelete(ctx context.Context, tenantID string, workflowIDs []string) workflow.BulkOperationResult {
	args := m.Called(ctx, tenantID, workflowIDs)
	return args.Get(0).(workflow.BulkOperationResult)
}

func (m *MockWorkflowBulkService) BulkEnable(ctx context.Context, tenantID string, workflowIDs []string) workflow.BulkOperationResult {
	args := m.Called(ctx, tenantID, workflowIDs)
	return args.Get(0).(workflow.BulkOperationResult)
}

func (m *MockWorkflowBulkService) BulkDisable(ctx context.Context, tenantID string, workflowIDs []string) workflow.BulkOperationResult {
	args := m.Called(ctx, tenantID, workflowIDs)
	return args.Get(0).(workflow.BulkOperationResult)
}

func (m *MockWorkflowBulkService) BulkExport(ctx context.Context, tenantID string, workflowIDs []string) (workflow.WorkflowExport, workflow.BulkOperationResult) {
	args := m.Called(ctx, tenantID, workflowIDs)
	return args.Get(0).(workflow.WorkflowExport), args.Get(1).(workflow.BulkOperationResult)
}

func (m *MockWorkflowBulkService) BulkClone(ctx context.Context, tenantID, userID string, workflowIDs []string) ([]*workflow.Workflow, workflow.BulkOperationResult) {
	args := m.Called(ctx, tenantID, userID, workflowIDs)
	if args.Get(0) == nil {
		return nil, args.Get(1).(workflow.BulkOperationResult)
	}
	return args.Get(0).([]*workflow.Workflow), args.Get(1).(workflow.BulkOperationResult)
}

func newTestWorkflowBulkHandler() (*WorkflowBulkHandler, *MockWorkflowBulkService) {
	mockService := new(MockWorkflowBulkService)
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	handler := NewWorkflowBulkHandler(mockService, logger)
	return handler, mockService
}

func addTenantID(req *http.Request, tenantID string) *http.Request {
	t := &tenant.Tenant{ID: tenantID, Name: "Test Tenant", Status: "active"}
	ctx := context.WithValue(req.Context(), middleware.TenantContextKey, t)
	return req.WithContext(ctx)
}

func addUserID(req *http.Request, userID string) *http.Request {
	user := &middleware.User{ID: userID, Email: "test@example.com"}
	ctx := context.WithValue(req.Context(), middleware.UserContextKey, user)
	return req.WithContext(ctx)
}

// Test BulkDelete
func TestBulkDelete(t *testing.T) {
	t.Run("successfully deletes workflows", func(t *testing.T) {
		handler, mockService := newTestWorkflowBulkHandler()

		requestBody := BulkOperationRequest{
			WorkflowIDs: []string{"workflow-1", "workflow-2", "workflow-3"},
		}
		bodyBytes, _ := json.Marshal(requestBody)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/workflows/bulk/delete", bytes.NewReader(bodyBytes))
		req = addTenantID(req, "tenant-123")

		result := workflow.BulkOperationResult{
			SuccessCount: 3,
			Failures:     []workflow.BulkOperationFailure{},
		}

		mockService.On("BulkDelete", mock.Anything, "tenant-123", requestBody.WorkflowIDs).Return(result)

		rr := httptest.NewRecorder()
		handler.BulkDelete(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var response map[string]interface{}
		json.Unmarshal(rr.Body.Bytes(), &response)

		assert.Equal(t, float64(3), response["success_count"])
		assert.Empty(t, response["failures"])
		mockService.AssertExpectations(t)
	})

	t.Run("handles partial failures", func(t *testing.T) {
		handler, mockService := newTestWorkflowBulkHandler()

		requestBody := BulkOperationRequest{
			WorkflowIDs: []string{"workflow-1", "workflow-2"},
		}
		bodyBytes, _ := json.Marshal(requestBody)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/workflows/bulk/delete", bytes.NewReader(bodyBytes))
		req = addTenantID(req, "tenant-123")

		result := workflow.BulkOperationResult{
			SuccessCount: 1,
			Failures: []workflow.BulkOperationFailure{
				{WorkflowID: "workflow-2", Error: "not found"},
			},
		}

		mockService.On("BulkDelete", mock.Anything, "tenant-123", requestBody.WorkflowIDs).Return(result)

		rr := httptest.NewRecorder()
		handler.BulkDelete(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var response workflow.BulkOperationResult
		json.Unmarshal(rr.Body.Bytes(), &response)

		assert.Equal(t, 1, response.SuccessCount)
		assert.Len(t, response.Failures, 1)
		mockService.AssertExpectations(t)
	})

	t.Run("handles invalid request body", func(t *testing.T) {
		handler, _ := newTestWorkflowBulkHandler()

		req := httptest.NewRequest(http.MethodPost, "/api/v1/workflows/bulk/delete", bytes.NewReader([]byte("invalid json")))
		req = addTenantID(req, "tenant-123")

		rr := httptest.NewRecorder()
		handler.BulkDelete(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("handles empty workflow list", func(t *testing.T) {
		handler, _ := newTestWorkflowBulkHandler()

		requestBody := BulkOperationRequest{
			WorkflowIDs: []string{},
		}
		bodyBytes, _ := json.Marshal(requestBody)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/workflows/bulk/delete", bytes.NewReader(bodyBytes))
		req = addTenantID(req, "tenant-123")

		rr := httptest.NewRecorder()
		handler.BulkDelete(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})
}

// Test BulkEnable
func TestBulkEnable(t *testing.T) {
	t.Run("successfully enables workflows", func(t *testing.T) {
		handler, mockService := newTestWorkflowBulkHandler()

		requestBody := BulkOperationRequest{
			WorkflowIDs: []string{"workflow-1", "workflow-2"},
		}
		bodyBytes, _ := json.Marshal(requestBody)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/workflows/bulk/enable", bytes.NewReader(bodyBytes))
		req = addTenantID(req, "tenant-123")

		result := workflow.BulkOperationResult{
			SuccessCount: 2,
			Failures:     []workflow.BulkOperationFailure{},
		}

		mockService.On("BulkEnable", mock.Anything, "tenant-123", requestBody.WorkflowIDs).Return(result)

		rr := httptest.NewRecorder()
		handler.BulkEnable(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		mockService.AssertExpectations(t)
	})
}

// Test BulkDisable
func TestBulkDisable(t *testing.T) {
	t.Run("successfully disables workflows", func(t *testing.T) {
		handler, mockService := newTestWorkflowBulkHandler()

		requestBody := BulkOperationRequest{
			WorkflowIDs: []string{"workflow-1", "workflow-2"},
		}
		bodyBytes, _ := json.Marshal(requestBody)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/workflows/bulk/disable", bytes.NewReader(bodyBytes))
		req = addTenantID(req, "tenant-123")

		result := workflow.BulkOperationResult{
			SuccessCount: 2,
			Failures:     []workflow.BulkOperationFailure{},
		}

		mockService.On("BulkDisable", mock.Anything, "tenant-123", requestBody.WorkflowIDs).Return(result)

		rr := httptest.NewRecorder()
		handler.BulkDisable(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		mockService.AssertExpectations(t)
	})
}

// Test BulkExport
func TestBulkExport(t *testing.T) {
	t.Run("successfully exports workflows", func(t *testing.T) {
		handler, mockService := newTestWorkflowBulkHandler()

		requestBody := BulkOperationRequest{
			WorkflowIDs: []string{"workflow-1", "workflow-2"},
		}
		bodyBytes, _ := json.Marshal(requestBody)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/workflows/bulk/export", bytes.NewReader(bodyBytes))
		req = addTenantID(req, "tenant-123")

		definition := json.RawMessage(`{"nodes":[],"edges":[]}`)
		export := workflow.WorkflowExport{
			Workflows: []workflow.WorkflowExportItem{
				{
					ID:          "workflow-1",
					Name:        "Test Workflow 1",
					Description: "Description 1",
					Definition:  definition,
					Status:      "active",
					Version:     1,
				},
				{
					ID:          "workflow-2",
					Name:        "Test Workflow 2",
					Description: "Description 2",
					Definition:  definition,
					Status:      "active",
					Version:     2,
				},
			},
			Version: "1.0",
		}

		result := workflow.BulkOperationResult{
			SuccessCount: 2,
			Failures:     []workflow.BulkOperationFailure{},
		}

		mockService.On("BulkExport", mock.Anything, "tenant-123", requestBody.WorkflowIDs).Return(export, result)

		rr := httptest.NewRecorder()
		handler.BulkExport(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var response map[string]interface{}
		json.Unmarshal(rr.Body.Bytes(), &response)

		assert.NotNil(t, response["export"])
		assert.NotNil(t, response["result"])
		mockService.AssertExpectations(t)
	})

	t.Run("handles export with failures", func(t *testing.T) {
		handler, mockService := newTestWorkflowBulkHandler()

		requestBody := BulkOperationRequest{
			WorkflowIDs: []string{"workflow-1", "workflow-2"},
		}
		bodyBytes, _ := json.Marshal(requestBody)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/workflows/bulk/export", bytes.NewReader(bodyBytes))
		req = addTenantID(req, "tenant-123")

		definition := json.RawMessage(`{"nodes":[],"edges":[]}`)
		export := workflow.WorkflowExport{
			Workflows: []workflow.WorkflowExportItem{
				{
					ID:         "workflow-1",
					Name:       "Test Workflow 1",
					Definition: definition,
					Status:     "active",
					Version:    1,
				},
			},
			Version: "1.0",
		}

		result := workflow.BulkOperationResult{
			SuccessCount: 1,
			Failures: []workflow.BulkOperationFailure{
				{WorkflowID: "workflow-2", Error: "not found"},
			},
		}

		mockService.On("BulkExport", mock.Anything, "tenant-123", requestBody.WorkflowIDs).Return(export, result)

		rr := httptest.NewRecorder()
		handler.BulkExport(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		mockService.AssertExpectations(t)
	})
}

// Test BulkClone
func TestBulkClone(t *testing.T) {
	t.Run("successfully clones workflows", func(t *testing.T) {
		handler, mockService := newTestWorkflowBulkHandler()

		requestBody := BulkOperationRequest{
			WorkflowIDs: []string{"workflow-1", "workflow-2"},
		}
		bodyBytes, _ := json.Marshal(requestBody)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/workflows/bulk/clone", bytes.NewReader(bodyBytes))
		req = addTenantID(req, "tenant-123")
		req = addUserID(req, "user-456")

		definition := json.RawMessage(`{"nodes":[],"edges":[]}`)
		clones := []*workflow.Workflow{
			{
				ID:          "cloned-1",
				Name:        "Test Workflow 1 (Copy)",
				Description: "Description 1",
				Definition:  definition,
				Status:      "draft",
				Version:     1,
			},
			{
				ID:          "cloned-2",
				Name:        "Test Workflow 2 (Copy)",
				Description: "Description 2",
				Definition:  definition,
				Status:      "draft",
				Version:     1,
			},
		}

		result := workflow.BulkOperationResult{
			SuccessCount: 2,
			Failures:     []workflow.BulkOperationFailure{},
		}

		mockService.On("BulkClone", mock.Anything, "tenant-123", "user-456", requestBody.WorkflowIDs).Return(clones, result)

		rr := httptest.NewRecorder()
		handler.BulkClone(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var response map[string]interface{}
		json.Unmarshal(rr.Body.Bytes(), &response)

		assert.NotNil(t, response["clones"])
		assert.NotNil(t, response["result"])
		mockService.AssertExpectations(t)
	})

	t.Run("handles clone failures", func(t *testing.T) {
		handler, mockService := newTestWorkflowBulkHandler()

		requestBody := BulkOperationRequest{
			WorkflowIDs: []string{"workflow-1", "workflow-2"},
		}
		bodyBytes, _ := json.Marshal(requestBody)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/workflows/bulk/clone", bytes.NewReader(bodyBytes))
		req = addTenantID(req, "tenant-123")
		req = addUserID(req, "user-456")

		definition := json.RawMessage(`{"nodes":[],"edges":[]}`)
		clones := []*workflow.Workflow{
			{
				ID:         "cloned-1",
				Name:       "Test Workflow 1 (Copy)",
				Definition: definition,
				Status:     "draft",
				Version:    1,
			},
		}

		result := workflow.BulkOperationResult{
			SuccessCount: 1,
			Failures: []workflow.BulkOperationFailure{
				{WorkflowID: "workflow-2", Error: "not found"},
			},
		}

		mockService.On("BulkClone", mock.Anything, "tenant-123", "user-456", requestBody.WorkflowIDs).Return(clones, result)

		rr := httptest.NewRecorder()
		handler.BulkClone(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		mockService.AssertExpectations(t)
	})

	t.Run("handles missing user ID", func(t *testing.T) {
		handler, _ := newTestWorkflowBulkHandler()

		requestBody := BulkOperationRequest{
			WorkflowIDs: []string{"workflow-1"},
		}
		bodyBytes, _ := json.Marshal(requestBody)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/workflows/bulk/clone", bytes.NewReader(bodyBytes))
		req = addTenantID(req, "tenant-123")
		// No user ID added

		rr := httptest.NewRecorder()
		handler.BulkClone(rr, req)

		assert.Equal(t, http.StatusUnauthorized, rr.Code)
	})
}

// Test validation
func TestBulkOperationValidation(t *testing.T) {
	t.Run("rejects request with nil workflow IDs", func(t *testing.T) {
		handler, _ := newTestWorkflowBulkHandler()

		requestBody := BulkOperationRequest{
			WorkflowIDs: nil,
		}
		bodyBytes, _ := json.Marshal(requestBody)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/workflows/bulk/delete", bytes.NewReader(bodyBytes))
		req = addTenantID(req, "tenant-123")

		rr := httptest.NewRecorder()
		handler.BulkDelete(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)

		var response map[string]string
		json.Unmarshal(rr.Body.Bytes(), &response)
		assert.Contains(t, response["error"], "workflow_ids")
	})

	t.Run("rejects request with too many workflows", func(t *testing.T) {
		handler, _ := newTestWorkflowBulkHandler()

		// Create more than 100 workflow IDs
		workflowIDs := make([]string, 101)
		for i := 0; i < 101; i++ {
			workflowIDs[i] = "workflow-" + string(rune(i))
		}

		requestBody := BulkOperationRequest{
			WorkflowIDs: workflowIDs,
		}
		bodyBytes, _ := json.Marshal(requestBody)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/workflows/bulk/delete", bytes.NewReader(bodyBytes))
		req = addTenantID(req, "tenant-123")

		rr := httptest.NewRecorder()
		handler.BulkDelete(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)

		var response map[string]string
		json.Unmarshal(rr.Body.Bytes(), &response)
		assert.Contains(t, response["error"], "100")
	})
}

// Integration Tests - Full Request/Response Cycle

func TestBulkOperationIntegration_DeleteWithMixedResults(t *testing.T) {
	t.Run("processes mixed success and failure results", func(t *testing.T) {
		handler, mockService := newTestWorkflowBulkHandler()

		requestBody := BulkOperationRequest{
			WorkflowIDs: []string{"wf-1", "wf-2", "wf-3", "wf-4", "wf-5"},
		}
		bodyBytes, _ := json.Marshal(requestBody)

		result := workflow.BulkOperationResult{
			SuccessCount: 3,
			Failures: []workflow.BulkOperationFailure{
				{WorkflowID: "wf-2", Error: "workflow is currently executing"},
				{WorkflowID: "wf-4", Error: "insufficient permissions"},
			},
		}

		mockService.On("BulkDelete", mock.Anything, "tenant-integration", requestBody.WorkflowIDs).Return(result)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/workflows/bulk/delete", bytes.NewReader(bodyBytes))
		req = addTenantID(req, "tenant-integration")

		rr := httptest.NewRecorder()
		handler.BulkDelete(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))

		var response workflow.BulkOperationResult
		json.Unmarshal(rr.Body.Bytes(), &response)

		assert.Equal(t, 3, response.SuccessCount)
		assert.Len(t, response.Failures, 2)
		assert.Equal(t, "wf-2", response.Failures[0].WorkflowID)
		assert.Equal(t, "workflow is currently executing", response.Failures[0].Error)
		assert.Equal(t, "wf-4", response.Failures[1].WorkflowID)
		assert.Equal(t, "insufficient permissions", response.Failures[1].Error)
		mockService.AssertExpectations(t)
	})

	t.Run("handles total failure scenario", func(t *testing.T) {
		handler, mockService := newTestWorkflowBulkHandler()

		requestBody := BulkOperationRequest{
			WorkflowIDs: []string{"wf-nonexistent-1", "wf-nonexistent-2"},
		}
		bodyBytes, _ := json.Marshal(requestBody)

		result := workflow.BulkOperationResult{
			SuccessCount: 0,
			Failures: []workflow.BulkOperationFailure{
				{WorkflowID: "wf-nonexistent-1", Error: "workflow not found"},
				{WorkflowID: "wf-nonexistent-2", Error: "workflow not found"},
			},
		}

		mockService.On("BulkDelete", mock.Anything, "tenant-123", requestBody.WorkflowIDs).Return(result)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/workflows/bulk/delete", bytes.NewReader(bodyBytes))
		req = addTenantID(req, "tenant-123")

		rr := httptest.NewRecorder()
		handler.BulkDelete(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var response workflow.BulkOperationResult
		json.Unmarshal(rr.Body.Bytes(), &response)

		assert.Equal(t, 0, response.SuccessCount)
		assert.Len(t, response.Failures, 2)
		mockService.AssertExpectations(t)
	})
}

func TestBulkOperationIntegration_EnableDisable_StatusTransitions(t *testing.T) {
	t.Run("enables multiple disabled workflows", func(t *testing.T) {
		handler, mockService := newTestWorkflowBulkHandler()

		requestBody := BulkOperationRequest{
			WorkflowIDs: []string{"wf-disabled-1", "wf-disabled-2", "wf-disabled-3"},
		}
		bodyBytes, _ := json.Marshal(requestBody)

		result := workflow.BulkOperationResult{
			SuccessCount: 3,
			Failures:     []workflow.BulkOperationFailure{},
		}

		mockService.On("BulkEnable", mock.Anything, "tenant-status", requestBody.WorkflowIDs).Return(result)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/workflows/bulk/enable", bytes.NewReader(bodyBytes))
		req = addTenantID(req, "tenant-status")

		rr := httptest.NewRecorder()
		handler.BulkEnable(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var response workflow.BulkOperationResult
		json.Unmarshal(rr.Body.Bytes(), &response)

		assert.Equal(t, 3, response.SuccessCount)
		assert.Empty(t, response.Failures)
		mockService.AssertExpectations(t)
	})

	t.Run("disables active workflows with partial success", func(t *testing.T) {
		handler, mockService := newTestWorkflowBulkHandler()

		requestBody := BulkOperationRequest{
			WorkflowIDs: []string{"wf-active-1", "wf-active-2", "wf-running-3"},
		}
		bodyBytes, _ := json.Marshal(requestBody)

		result := workflow.BulkOperationResult{
			SuccessCount: 2,
			Failures: []workflow.BulkOperationFailure{
				{WorkflowID: "wf-running-3", Error: "cannot disable workflow with active executions"},
			},
		}

		mockService.On("BulkDisable", mock.Anything, "tenant-123", requestBody.WorkflowIDs).Return(result)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/workflows/bulk/disable", bytes.NewReader(bodyBytes))
		req = addTenantID(req, "tenant-123")

		rr := httptest.NewRecorder()
		handler.BulkDisable(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var response workflow.BulkOperationResult
		json.Unmarshal(rr.Body.Bytes(), &response)

		assert.Equal(t, 2, response.SuccessCount)
		assert.Len(t, response.Failures, 1)
		assert.Contains(t, response.Failures[0].Error, "active executions")
		mockService.AssertExpectations(t)
	})
}

func TestBulkOperationIntegration_Export_CompleteWorkflowData(t *testing.T) {
	t.Run("exports multiple workflows with full definitions", func(t *testing.T) {
		handler, mockService := newTestWorkflowBulkHandler()

		requestBody := BulkOperationRequest{
			WorkflowIDs: []string{"wf-export-1", "wf-export-2", "wf-export-3"},
		}
		bodyBytes, _ := json.Marshal(requestBody)

		definition1 := json.RawMessage(`{"nodes":[{"id":"1","type":"trigger:webhook"}],"edges":[]}`)
		definition2 := json.RawMessage(`{"nodes":[{"id":"1","type":"trigger:schedule"},{"id":"2","type":"action:http"}],"edges":[{"from":"1","to":"2"}]}`)
		definition3 := json.RawMessage(`{"nodes":[{"id":"1","type":"trigger:webhook"},{"id":"2","type":"action:transform"},{"id":"3","type":"action:http"}],"edges":[{"from":"1","to":"2"},{"from":"2","to":"3"}]}`)

		export := workflow.WorkflowExport{
			Workflows: []workflow.WorkflowExportItem{
				{
					ID:          "wf-export-1",
					Name:        "Simple Webhook Workflow",
					Description: "Basic webhook trigger",
					Definition:  definition1,
					Status:      "active",
					Version:     1,
				},
				{
					ID:          "wf-export-2",
					Name:        "Scheduled HTTP Workflow",
					Description: "Runs on schedule and makes HTTP call",
					Definition:  definition2,
					Status:      "active",
					Version:     2,
				},
				{
					ID:          "wf-export-3",
					Name:        "Complex Transformation Pipeline",
					Description: "Multi-step transformation workflow",
					Definition:  definition3,
					Status:      "active",
					Version:     3,
				},
			},
			Version:    "1.0",
			ExportedAt: time.Now(),
		}

		result := workflow.BulkOperationResult{
			SuccessCount: 3,
			Failures:     []workflow.BulkOperationFailure{},
		}

		mockService.On("BulkExport", mock.Anything, "tenant-export", requestBody.WorkflowIDs).Return(export, result)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/workflows/bulk/export", bytes.NewReader(bodyBytes))
		req = addTenantID(req, "tenant-export")

		rr := httptest.NewRecorder()
		handler.BulkExport(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var response map[string]interface{}
		json.Unmarshal(rr.Body.Bytes(), &response)

		assert.NotNil(t, response["export"])
		assert.NotNil(t, response["result"])

		exportData := response["export"].(map[string]interface{})
		workflows := exportData["workflows"].([]interface{})
		assert.Len(t, workflows, 3)

		firstWorkflow := workflows[0].(map[string]interface{})
		assert.Equal(t, "wf-export-1", firstWorkflow["id"])
		assert.Equal(t, "Simple Webhook Workflow", firstWorkflow["name"])

		resultData := response["result"].(map[string]interface{})
		assert.Equal(t, float64(3), resultData["success_count"])
		mockService.AssertExpectations(t)
	})

	t.Run("exports with partial failures", func(t *testing.T) {
		handler, mockService := newTestWorkflowBulkHandler()

		requestBody := BulkOperationRequest{
			WorkflowIDs: []string{"wf-1", "wf-missing", "wf-3"},
		}
		bodyBytes, _ := json.Marshal(requestBody)

		definition := json.RawMessage(`{"nodes":[],"edges":[]}`)
		export := workflow.WorkflowExport{
			Workflows: []workflow.WorkflowExportItem{
				{ID: "wf-1", Name: "Workflow 1", Definition: definition, Status: "active", Version: 1},
				{ID: "wf-3", Name: "Workflow 3", Definition: definition, Status: "active", Version: 1},
			},
			Version: "1.0",
		}

		result := workflow.BulkOperationResult{
			SuccessCount: 2,
			Failures: []workflow.BulkOperationFailure{
				{WorkflowID: "wf-missing", Error: "workflow not found"},
			},
		}

		mockService.On("BulkExport", mock.Anything, "tenant-123", requestBody.WorkflowIDs).Return(export, result)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/workflows/bulk/export", bytes.NewReader(bodyBytes))
		req = addTenantID(req, "tenant-123")

		rr := httptest.NewRecorder()
		handler.BulkExport(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var response map[string]interface{}
		json.Unmarshal(rr.Body.Bytes(), &response)

		exportData := response["export"].(map[string]interface{})
		workflows := exportData["workflows"].([]interface{})
		assert.Len(t, workflows, 2)

		resultData := response["result"].(map[string]interface{})
		assert.Equal(t, float64(2), resultData["success_count"])
		failures := resultData["failures"].([]interface{})
		assert.Len(t, failures, 1)
		mockService.AssertExpectations(t)
	})
}

func TestBulkOperationIntegration_Clone_FullWorkflowCopy(t *testing.T) {
	t.Run("clones workflows with name suffixes", func(t *testing.T) {
		handler, mockService := newTestWorkflowBulkHandler()

		requestBody := BulkOperationRequest{
			WorkflowIDs: []string{"wf-original-1", "wf-original-2"},
		}
		bodyBytes, _ := json.Marshal(requestBody)

		definition := json.RawMessage(`{
			"nodes": [
				{"id": "1", "type": "trigger:webhook", "config": {"path": "/webhook"}},
				{"id": "2", "type": "action:http", "config": {"url": "https://api.example.com"}}
			],
			"edges": [{"from": "1", "to": "2"}]
		}`)

		clones := []*workflow.Workflow{
			{
				ID:          "wf-clone-1",
				Name:        "Original Workflow 1 (Copy)",
				Description: "Cloned from original workflow",
				Definition:  definition,
				Status:      "draft",
				Version:     1,
				TenantID:    "tenant-clone",
				CreatedBy:   "user-clone",
			},
			{
				ID:          "wf-clone-2",
				Name:        "Original Workflow 2 (Copy)",
				Description: "Cloned from original workflow",
				Definition:  definition,
				Status:      "draft",
				Version:     1,
				TenantID:    "tenant-clone",
				CreatedBy:   "user-clone",
			},
		}

		result := workflow.BulkOperationResult{
			SuccessCount: 2,
			Failures:     []workflow.BulkOperationFailure{},
		}

		mockService.On("BulkClone", mock.Anything, "tenant-clone", "user-clone", requestBody.WorkflowIDs).Return(clones, result)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/workflows/bulk/clone", bytes.NewReader(bodyBytes))
		req = addTenantID(req, "tenant-clone")
		req = addUserID(req, "user-clone")

		rr := httptest.NewRecorder()
		handler.BulkClone(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var response map[string]interface{}
		json.Unmarshal(rr.Body.Bytes(), &response)

		assert.NotNil(t, response["clones"])
		assert.NotNil(t, response["result"])

		clonesData := response["clones"].([]interface{})
		assert.Len(t, clonesData, 2)

		firstClone := clonesData[0].(map[string]interface{})
		assert.Equal(t, "wf-clone-1", firstClone["id"])
		assert.Contains(t, firstClone["name"], "(Copy)")
		assert.Equal(t, "draft", firstClone["status"])

		resultData := response["result"].(map[string]interface{})
		assert.Equal(t, float64(2), resultData["success_count"])
		mockService.AssertExpectations(t)
	})

	t.Run("handles clone failures for specific workflows", func(t *testing.T) {
		handler, mockService := newTestWorkflowBulkHandler()

		requestBody := BulkOperationRequest{
			WorkflowIDs: []string{"wf-valid", "wf-locked", "wf-deleted"},
		}
		bodyBytes, _ := json.Marshal(requestBody)

		definition := json.RawMessage(`{"nodes":[],"edges":[]}`)
		clones := []*workflow.Workflow{
			{
				ID:         "wf-clone-valid",
				Name:       "Valid Workflow (Copy)",
				Definition: definition,
				Status:     "draft",
				Version:    1,
			},
		}

		result := workflow.BulkOperationResult{
			SuccessCount: 1,
			Failures: []workflow.BulkOperationFailure{
				{WorkflowID: "wf-locked", Error: "workflow is locked for editing"},
				{WorkflowID: "wf-deleted", Error: "workflow not found"},
			},
		}

		mockService.On("BulkClone", mock.Anything, "tenant-123", "user-123", requestBody.WorkflowIDs).Return(clones, result)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/workflows/bulk/clone", bytes.NewReader(bodyBytes))
		req = addTenantID(req, "tenant-123")
		req = addUserID(req, "user-123")

		rr := httptest.NewRecorder()
		handler.BulkClone(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var response map[string]interface{}
		json.Unmarshal(rr.Body.Bytes(), &response)

		clonesData := response["clones"].([]interface{})
		assert.Len(t, clonesData, 1)

		resultData := response["result"].(map[string]interface{})
		assert.Equal(t, float64(1), resultData["success_count"])
		failures := resultData["failures"].([]interface{})
		assert.Len(t, failures, 2)
		mockService.AssertExpectations(t)
	})
}

func TestBulkOperationIntegration_EdgeCases(t *testing.T) {
	t.Run("handles single workflow ID", func(t *testing.T) {
		handler, mockService := newTestWorkflowBulkHandler()

		requestBody := BulkOperationRequest{
			WorkflowIDs: []string{"wf-single"},
		}
		bodyBytes, _ := json.Marshal(requestBody)

		result := workflow.BulkOperationResult{
			SuccessCount: 1,
			Failures:     []workflow.BulkOperationFailure{},
		}

		mockService.On("BulkDelete", mock.Anything, "tenant-123", requestBody.WorkflowIDs).Return(result)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/workflows/bulk/delete", bytes.NewReader(bodyBytes))
		req = addTenantID(req, "tenant-123")

		rr := httptest.NewRecorder()
		handler.BulkDelete(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		mockService.AssertExpectations(t)
	})

	t.Run("handles maximum allowed workflows (100)", func(t *testing.T) {
		handler, mockService := newTestWorkflowBulkHandler()

		workflowIDs := make([]string, 100)
		for i := 0; i < 100; i++ {
			workflowIDs[i] = fmt.Sprintf("wf-%d", i)
		}

		requestBody := BulkOperationRequest{
			WorkflowIDs: workflowIDs,
		}
		bodyBytes, _ := json.Marshal(requestBody)

		result := workflow.BulkOperationResult{
			SuccessCount: 100,
			Failures:     []workflow.BulkOperationFailure{},
		}

		mockService.On("BulkDelete", mock.Anything, "tenant-123", workflowIDs).Return(result)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/workflows/bulk/delete", bytes.NewReader(bodyBytes))
		req = addTenantID(req, "tenant-123")

		rr := httptest.NewRecorder()
		handler.BulkDelete(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var response workflow.BulkOperationResult
		json.Unmarshal(rr.Body.Bytes(), &response)
		assert.Equal(t, 100, response.SuccessCount)
		mockService.AssertExpectations(t)
	})
}
