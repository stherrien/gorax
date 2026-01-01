package handlers

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/go-playground/validator/v10"

	"github.com/gorax/gorax/internal/api/middleware"
	"github.com/gorax/gorax/internal/workflow"
)

// WorkflowBulkService defines the interface for bulk workflow operations
type WorkflowBulkService interface {
	BulkDelete(ctx context.Context, tenantID string, workflowIDs []string) workflow.BulkOperationResult
	BulkEnable(ctx context.Context, tenantID string, workflowIDs []string) workflow.BulkOperationResult
	BulkDisable(ctx context.Context, tenantID string, workflowIDs []string) workflow.BulkOperationResult
	BulkExport(ctx context.Context, tenantID string, workflowIDs []string) (workflow.WorkflowExport, workflow.BulkOperationResult)
	BulkClone(ctx context.Context, tenantID, userID string, workflowIDs []string) ([]*workflow.Workflow, workflow.BulkOperationResult)
}

// WorkflowBulkHandler handles bulk workflow operations
type WorkflowBulkHandler struct {
	service  WorkflowBulkService
	logger   *slog.Logger
	validate *validator.Validate
}

// NewWorkflowBulkHandler creates a new bulk workflow handler
func NewWorkflowBulkHandler(service WorkflowBulkService, logger *slog.Logger) *WorkflowBulkHandler {
	return &WorkflowBulkHandler{
		service:  service,
		logger:   logger,
		validate: validator.New(),
	}
}

// BulkOperationRequest represents a bulk operation request
type BulkOperationRequest struct {
	WorkflowIDs []string `json:"workflow_ids" validate:"required,min=1,max=100,dive,required"`
}

// BulkDelete handles bulk delete requests
// @Summary Bulk delete workflows
// @Description Deletes multiple workflows in a single operation
// @Tags Workflows
// @Accept json
// @Produce json
// @Param request body BulkOperationRequest true "Bulk delete request"
// @Security TenantID
// @Security UserID
// @Success 200 {object} workflow.BulkOperationResult
// @Failure 400 {object} map[string]string "Invalid request"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /workflows/bulk/delete [post]
func (h *WorkflowBulkHandler) BulkDelete(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r)

	var input BulkOperationRequest
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.validate.Struct(input); err != nil {
		h.respondError(w, http.StatusBadRequest, "workflow_ids is required and must contain 1-100 workflow IDs")
		return
	}

	result := h.service.BulkDelete(r.Context(), tenantID, input.WorkflowIDs)

	h.respondJSON(w, http.StatusOK, result)
}

// BulkEnable handles bulk enable requests
// @Summary Bulk enable workflows
// @Description Enables multiple workflows in a single operation
// @Tags Workflows
// @Accept json
// @Produce json
// @Param request body BulkOperationRequest true "Bulk enable request"
// @Security TenantID
// @Security UserID
// @Success 200 {object} workflow.BulkOperationResult
// @Failure 400 {object} map[string]string "Invalid request"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /workflows/bulk/enable [post]
func (h *WorkflowBulkHandler) BulkEnable(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r)

	var input BulkOperationRequest
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.validate.Struct(input); err != nil {
		h.respondError(w, http.StatusBadRequest, "workflow_ids is required and must contain 1-100 workflow IDs")
		return
	}

	result := h.service.BulkEnable(r.Context(), tenantID, input.WorkflowIDs)

	h.respondJSON(w, http.StatusOK, result)
}

// BulkDisable handles bulk disable requests
// @Summary Bulk disable workflows
// @Description Disables multiple workflows in a single operation
// @Tags Workflows
// @Accept json
// @Produce json
// @Param request body BulkOperationRequest true "Bulk disable request"
// @Security TenantID
// @Security UserID
// @Success 200 {object} workflow.BulkOperationResult
// @Failure 400 {object} map[string]string "Invalid request"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /workflows/bulk/disable [post]
func (h *WorkflowBulkHandler) BulkDisable(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r)

	var input BulkOperationRequest
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.validate.Struct(input); err != nil {
		h.respondError(w, http.StatusBadRequest, "workflow_ids is required and must contain 1-100 workflow IDs")
		return
	}

	result := h.service.BulkDisable(r.Context(), tenantID, input.WorkflowIDs)

	h.respondJSON(w, http.StatusOK, result)
}

// BulkExport handles bulk export requests
// @Summary Bulk export workflows
// @Description Exports multiple workflows as JSON in a single operation
// @Tags Workflows
// @Accept json
// @Produce json
// @Param request body BulkOperationRequest true "Bulk export request"
// @Security TenantID
// @Security UserID
// @Success 200 {object} map[string]interface{} "Export data and result"
// @Failure 400 {object} map[string]string "Invalid request"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /workflows/bulk/export [post]
func (h *WorkflowBulkHandler) BulkExport(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r)

	var input BulkOperationRequest
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.validate.Struct(input); err != nil {
		h.respondError(w, http.StatusBadRequest, "workflow_ids is required and must contain 1-100 workflow IDs")
		return
	}

	export, result := h.service.BulkExport(r.Context(), tenantID, input.WorkflowIDs)

	h.respondJSON(w, http.StatusOK, map[string]interface{}{
		"export": export,
		"result": result,
	})
}

// BulkClone handles bulk clone requests
// @Summary Bulk clone workflows
// @Description Clones multiple workflows in a single operation
// @Tags Workflows
// @Accept json
// @Produce json
// @Param request body BulkOperationRequest true "Bulk clone request"
// @Security TenantID
// @Security UserID
// @Success 200 {object} map[string]interface{} "Cloned workflows and result"
// @Failure 400 {object} map[string]string "Invalid request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /workflows/bulk/clone [post]
func (h *WorkflowBulkHandler) BulkClone(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r)
	userID := middleware.GetUserID(r)

	if userID == "" {
		h.respondError(w, http.StatusUnauthorized, "user ID is required")
		return
	}

	var input BulkOperationRequest
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.validate.Struct(input); err != nil {
		h.respondError(w, http.StatusBadRequest, "workflow_ids is required and must contain 1-100 workflow IDs")
		return
	}

	clones, result := h.service.BulkClone(r.Context(), tenantID, userID, input.WorkflowIDs)

	h.respondJSON(w, http.StatusOK, map[string]interface{}{
		"clones": clones,
		"result": result,
	})
}

// Helper methods

func (h *WorkflowBulkHandler) respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func (h *WorkflowBulkHandler) respondError(w http.ResponseWriter, status int, message string) {
	h.respondJSON(w, status, map[string]string{
		"error": message,
	})
}
