package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/gorax/gorax/internal/api/middleware"
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
func (h *WorkflowHandler) List(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r)

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

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

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

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
