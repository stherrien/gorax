package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/gorax/gorax/internal/api/middleware"
	"github.com/gorax/gorax/internal/validation"
	"github.com/gorax/gorax/internal/workflow"
)

// WorkflowHandler handles workflow-related HTTP requests
type WorkflowHandler struct {
	service *workflow.Service
	logger  *slog.Logger
}

// NewWorkflowHandler creates a new workflow handler
func NewWorkflowHandler(service *workflow.Service, logger *slog.Logger) *WorkflowHandler {
	return &WorkflowHandler{
		service: service,
		logger:  logger,
	}
}

// List returns all workflows for the tenant
// @Summary List workflows
// @Description Returns a paginated list of workflows for the authenticated tenant
// @Tags Workflows
// @Accept json
// @Produce json
// @Param limit query int false "Maximum number of results" default(20)
// @Param offset query int false "Offset for pagination" default(0)
// @Security TenantID
// @Security UserID
// @Success 200 {object} map[string]interface{} "List of workflows"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /workflows [get]
func (h *WorkflowHandler) List(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r)

	limit, _ := validation.ParsePaginationLimit(
		r.URL.Query().Get("limit"),
		validation.DefaultPaginationLimit,
		validation.MaxPaginationLimit,
	)
	offset, _ := validation.ParsePaginationOffset(r.URL.Query().Get("offset"))

	workflows, err := h.service.List(r.Context(), tenantID, limit, offset)
	if err != nil {
		h.respondError(w, http.StatusInternalServerError, "failed to list workflows")
		return
	}

	h.respondJSON(w, http.StatusOK, map[string]interface{}{
		"data":   workflows,
		"limit":  limit,
		"offset": offset,
	})
}

// Create creates a new workflow
// @Summary Create workflow
// @Description Creates a new workflow with the provided configuration
// @Tags Workflows
// @Accept json
// @Produce json
// @Param workflow body workflow.CreateWorkflowInput true "Workflow creation input"
// @Security TenantID
// @Security UserID
// @Success 201 {object} map[string]interface{} "Created workflow"
// @Failure 400 {object} map[string]string "Invalid request or validation error"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /workflows [post]
func (h *WorkflowHandler) Create(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r)
	user := middleware.GetUser(r)

	var input workflow.CreateWorkflowInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	wf, err := h.service.Create(r.Context(), tenantID, user.ID, input)
	if err != nil {
		if _, ok := err.(*workflow.ValidationError); ok {
			h.respondError(w, http.StatusBadRequest, err.Error())
			return
		}
		h.respondError(w, http.StatusInternalServerError, "failed to create workflow")
		return
	}

	h.respondJSON(w, http.StatusCreated, map[string]interface{}{
		"data": wf,
	})
}

// Get retrieves a single workflow
// @Summary Get workflow
// @Description Retrieves a workflow by ID
// @Tags Workflows
// @Accept json
// @Produce json
// @Param workflowID path string true "Workflow ID"
// @Security TenantID
// @Security UserID
// @Success 200 {object} map[string]interface{} "Workflow details"
// @Failure 404 {object} map[string]string "Workflow not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /workflows/{workflowID} [get]
func (h *WorkflowHandler) Get(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r)
	workflowID := chi.URLParam(r, "workflowID")

	wf, err := h.service.GetByID(r.Context(), tenantID, workflowID)
	if err != nil {
		if err == workflow.ErrNotFound {
			h.respondError(w, http.StatusNotFound, "workflow not found")
			return
		}
		h.respondError(w, http.StatusInternalServerError, "failed to get workflow")
		return
	}

	h.respondJSON(w, http.StatusOK, map[string]interface{}{
		"data": wf,
	})
}

// Update updates a workflow
// @Summary Update workflow
// @Description Updates an existing workflow
// @Tags Workflows
// @Accept json
// @Produce json
// @Param workflowID path string true "Workflow ID"
// @Param workflow body workflow.UpdateWorkflowInput true "Workflow update input"
// @Security TenantID
// @Security UserID
// @Success 200 {object} map[string]interface{} "Updated workflow"
// @Failure 400 {object} map[string]string "Invalid request or validation error"
// @Failure 404 {object} map[string]string "Workflow not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /workflows/{workflowID} [put]
func (h *WorkflowHandler) Update(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r)
	workflowID := chi.URLParam(r, "workflowID")

	var input workflow.UpdateWorkflowInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	wf, err := h.service.Update(r.Context(), tenantID, workflowID, input)
	if err != nil {
		if err == workflow.ErrNotFound {
			h.respondError(w, http.StatusNotFound, "workflow not found")
			return
		}
		if _, ok := err.(*workflow.ValidationError); ok {
			h.respondError(w, http.StatusBadRequest, err.Error())
			return
		}
		h.respondError(w, http.StatusInternalServerError, "failed to update workflow")
		return
	}

	h.respondJSON(w, http.StatusOK, map[string]interface{}{
		"data": wf,
	})
}

// Delete deletes a workflow
// @Summary Delete workflow
// @Description Deletes a workflow by ID
// @Tags Workflows
// @Accept json
// @Produce json
// @Param workflowID path string true "Workflow ID"
// @Security TenantID
// @Security UserID
// @Success 204 "Workflow deleted successfully"
// @Failure 404 {object} map[string]string "Workflow not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /workflows/{workflowID} [delete]
func (h *WorkflowHandler) Delete(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r)
	workflowID := chi.URLParam(r, "workflowID")

	err := h.service.Delete(r.Context(), tenantID, workflowID)
	if err != nil {
		if err == workflow.ErrNotFound {
			h.respondError(w, http.StatusNotFound, "workflow not found")
			return
		}
		h.respondError(w, http.StatusInternalServerError, "failed to delete workflow")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Execute triggers a workflow execution
// @Summary Execute workflow
// @Description Triggers a manual execution of a workflow
// @Tags Workflows
// @Accept json
// @Produce json
// @Param workflowID path string true "Workflow ID"
// @Param triggerData body object false "Optional trigger data"
// @Security TenantID
// @Security UserID
// @Success 202 {object} map[string]interface{} "Execution started"
// @Failure 404 {object} map[string]string "Workflow not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /workflows/{workflowID}/execute [post]
func (h *WorkflowHandler) Execute(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r)
	workflowID := chi.URLParam(r, "workflowID")

	// Parse trigger data from request body
	var triggerData json.RawMessage
	if r.Body != nil {
		json.NewDecoder(r.Body).Decode(&triggerData)
	}

	execution, err := h.service.Execute(r.Context(), tenantID, workflowID, "manual", triggerData)
	if err != nil {
		if err == workflow.ErrNotFound {
			h.respondError(w, http.StatusNotFound, "workflow not found")
			return
		}
		h.respondError(w, http.StatusInternalServerError, "failed to execute workflow")
		return
	}

	h.respondJSON(w, http.StatusAccepted, map[string]interface{}{
		"data": execution,
	})
}

// ListExecutions returns executions for the tenant
func (h *WorkflowHandler) ListExecutions(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r)
	workflowID := r.URL.Query().Get("workflow_id")

	limit, _ := validation.ParsePaginationLimit(
		r.URL.Query().Get("limit"),
		validation.DefaultPaginationLimit,
		validation.MaxPaginationLimit,
	)
	offset, _ := validation.ParsePaginationOffset(r.URL.Query().Get("offset"))

	executions, err := h.service.ListExecutions(r.Context(), tenantID, workflowID, limit, offset)
	if err != nil {
		h.respondError(w, http.StatusInternalServerError, "failed to list executions")
		return
	}

	h.respondJSON(w, http.StatusOK, map[string]interface{}{
		"data":   executions,
		"limit":  limit,
		"offset": offset,
	})
}

// GetExecution retrieves a single execution
func (h *WorkflowHandler) GetExecution(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r)
	executionID := chi.URLParam(r, "executionID")

	execution, err := h.service.GetExecution(r.Context(), tenantID, executionID)
	if err != nil {
		if err == workflow.ErrNotFound {
			h.respondError(w, http.StatusNotFound, "execution not found")
			return
		}
		h.respondError(w, http.StatusInternalServerError, "failed to get execution")
		return
	}

	h.respondJSON(w, http.StatusOK, map[string]interface{}{
		"data": execution,
	})
}

// DryRun performs a dry-run validation of a workflow
// @Summary Dry-run workflow
// @Description Validates a workflow without executing it, useful for testing
// @Tags Workflows
// @Accept json
// @Produce json
// @Param workflowID path string true "Workflow ID"
// @Param input body workflow.DryRunInput true "Dry-run input with test data"
// @Security TenantID
// @Security UserID
// @Success 200 {object} map[string]interface{} "Dry-run result"
// @Failure 400 {object} map[string]string "Invalid request or validation error"
// @Failure 404 {object} map[string]string "Workflow not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /workflows/{workflowID}/dry-run [post]
func (h *WorkflowHandler) DryRun(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r)
	workflowID := chi.URLParam(r, "workflowID")

	var input workflow.DryRunInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	result, err := h.service.DryRun(r.Context(), tenantID, workflowID, input.TestData)
	if err != nil {
		if err == workflow.ErrNotFound {
			h.respondError(w, http.StatusNotFound, "workflow not found")
			return
		}
		if _, ok := err.(*workflow.ValidationError); ok {
			h.respondError(w, http.StatusBadRequest, err.Error())
			return
		}
		h.respondError(w, http.StatusInternalServerError, "failed to perform dry-run")
		return
	}

	h.respondJSON(w, http.StatusOK, map[string]interface{}{
		"data": result,
	})
}

// Helper methods

func (h *WorkflowHandler) respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func (h *WorkflowHandler) respondError(w http.ResponseWriter, status int, message string) {
	h.respondJSON(w, status, map[string]string{
		"error": message,
	})
}

// ListVersions retrieves all versions for a workflow
func (h *WorkflowHandler) ListVersions(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r)
	workflowID := chi.URLParam(r, "workflowID")

	// Verify workflow belongs to tenant
	_, err := h.service.GetByID(r.Context(), tenantID, workflowID)
	if err != nil {
		if err == workflow.ErrNotFound {
			h.respondError(w, http.StatusNotFound, "workflow not found")
			return
		}
		h.respondError(w, http.StatusInternalServerError, "failed to get workflow")
		return
	}

	versions, err := h.service.ListWorkflowVersions(r.Context(), workflowID)
	if err != nil {
		h.respondError(w, http.StatusInternalServerError, "failed to list workflow versions")
		return
	}

	h.respondJSON(w, http.StatusOK, map[string]interface{}{
		"data": versions,
	})
}

// GetVersion retrieves a specific version of a workflow
func (h *WorkflowHandler) GetVersion(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r)
	workflowID := chi.URLParam(r, "workflowID")
	versionStr := chi.URLParam(r, "version")

	version, err := strconv.Atoi(versionStr)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid version number")
		return
	}

	// Verify workflow belongs to tenant
	_, err = h.service.GetByID(r.Context(), tenantID, workflowID)
	if err != nil {
		if err == workflow.ErrNotFound {
			h.respondError(w, http.StatusNotFound, "workflow not found")
			return
		}
		h.respondError(w, http.StatusInternalServerError, "failed to get workflow")
		return
	}

	versionData, err := h.service.GetWorkflowVersion(r.Context(), workflowID, version)
	if err != nil {
		if err == workflow.ErrNotFound {
			h.respondError(w, http.StatusNotFound, "version not found")
			return
		}
		h.respondError(w, http.StatusInternalServerError, "failed to get workflow version")
		return
	}

	h.respondJSON(w, http.StatusOK, map[string]interface{}{
		"data": versionData,
	})
}

// RestoreVersion restores a workflow to a previous version
func (h *WorkflowHandler) RestoreVersion(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r)
	workflowID := chi.URLParam(r, "workflowID")
	versionStr := chi.URLParam(r, "version")

	version, err := strconv.Atoi(versionStr)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid version number")
		return
	}

	restoredWorkflow, err := h.service.RestoreWorkflowVersion(r.Context(), tenantID, workflowID, version)
	if err != nil {
		if err == workflow.ErrNotFound {
			h.respondError(w, http.StatusNotFound, "workflow or version not found")
			return
		}
		h.respondError(w, http.StatusInternalServerError, "failed to restore workflow version")
		return
	}

	h.respondJSON(w, http.StatusOK, map[string]interface{}{
		"data": restoredWorkflow,
	})
}
