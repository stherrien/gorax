package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/gorax/gorax/internal/humantask"
)

// HumanTaskHandler handles HTTP requests for human tasks
type HumanTaskHandler struct {
	service humantask.Service
}

// NewHumanTaskHandler creates a new human task handler
func NewHumanTaskHandler(service humantask.Service) *HumanTaskHandler {
	return &HumanTaskHandler{
		service: service,
	}
}

// ListTasks godoc
// @Summary List human tasks
// @Description List human tasks with optional filters
// @Tags tasks
// @Accept json
// @Produce json
// @Param status query string false "Task status"
// @Param task_type query string false "Task type"
// @Param assignee query string false "Assignee user ID or role"
// @Param execution_id query string false "Execution ID"
// @Param limit query int false "Limit" default(20)
// @Param offset query int false "Offset" default(0)
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/tasks [get]
func (h *HumanTaskHandler) ListTasks(c *gin.Context) {
	tenantID, err := getTenantID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	filter := humantask.TaskFilter{
		TenantID: tenantID,
		Limit:    20,
		Offset:   0,
	}

	// Parse query parameters
	if status := c.Query("status"); status != "" {
		filter.Status = &status
	}

	if taskType := c.Query("task_type"); taskType != "" {
		filter.TaskType = &taskType
	}

	if assignee := c.Query("assignee"); assignee != "" {
		filter.Assignee = &assignee
	}

	if executionIDStr := c.Query("execution_id"); executionIDStr != "" {
		executionID, err := uuid.Parse(executionIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid execution_id"})
			return
		}
		filter.ExecutionID = &executionID
	}

	if limitStr := c.Query("limit"); limitStr != "" {
		limit, err := strconv.Atoi(limitStr)
		if err == nil && limit > 0 {
			filter.Limit = limit
		}
	}

	if offsetStr := c.Query("offset"); offsetStr != "" {
		offset, err := strconv.Atoi(offsetStr)
		if err == nil && offset >= 0 {
			filter.Offset = offset
		}
	}

	tasks, err := h.service.ListTasks(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Convert to response format
	responses := make([]*humantask.TaskResponse, len(tasks))
	for i, task := range tasks {
		resp, err := task.ToResponse()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to convert task"})
			return
		}
		responses[i] = resp
	}

	c.JSON(http.StatusOK, gin.H{
		"tasks": responses,
		"count": len(responses),
	})
}

// GetTask godoc
// @Summary Get a human task
// @Description Get a human task by ID
// @Tags tasks
// @Accept json
// @Produce json
// @Param id path string true "Task ID"
// @Success 200 {object} humantask.TaskResponse
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/tasks/{id} [get]
func (h *HumanTaskHandler) GetTask(c *gin.Context) {
	tenantID, err := getTenantID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	taskID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid task ID"})
		return
	}

	task, err := h.service.GetTask(c.Request.Context(), tenantID, taskID)
	if err != nil {
		if errors.Is(err, humantask.ErrTaskNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "task not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	response, err := task.ToResponse()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to convert task"})
		return
	}

	c.JSON(http.StatusOK, response)
}

// ApproveTask godoc
// @Summary Approve a task
// @Description Approve a human task
// @Tags tasks
// @Accept json
// @Produce json
// @Param id path string true "Task ID"
// @Param request body humantask.ApproveTaskRequest true "Approve request"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 409 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/tasks/{id}/approve [post]
func (h *HumanTaskHandler) ApproveTask(c *gin.Context) {
	tenantID, userID, roles, err := getAuthContext(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	taskID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid task ID"})
		return
	}

	var req humantask.ApproveTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err = h.service.ApproveTask(c.Request.Context(), tenantID, taskID, userID, roles, req)
	if err != nil {
		statusCode := getErrorStatusCode(err)
		c.JSON(statusCode, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "task approved successfully"})
}

// RejectTask godoc
// @Summary Reject a task
// @Description Reject a human task
// @Tags tasks
// @Accept json
// @Produce json
// @Param id path string true "Task ID"
// @Param request body humantask.RejectTaskRequest true "Reject request"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 409 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/tasks/{id}/reject [post]
func (h *HumanTaskHandler) RejectTask(c *gin.Context) {
	tenantID, userID, roles, err := getAuthContext(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	taskID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid task ID"})
		return
	}

	var req humantask.RejectTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err = h.service.RejectTask(c.Request.Context(), tenantID, taskID, userID, roles, req)
	if err != nil {
		statusCode := getErrorStatusCode(err)
		c.JSON(statusCode, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "task rejected successfully"})
}

// SubmitTask godoc
// @Summary Submit an input task
// @Description Submit data for an input task
// @Tags tasks
// @Accept json
// @Produce json
// @Param id path string true "Task ID"
// @Param request body humantask.SubmitTaskRequest true "Submit request"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 409 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/tasks/{id}/submit [post]
func (h *HumanTaskHandler) SubmitTask(c *gin.Context) {
	tenantID, userID, roles, err := getAuthContext(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	taskID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid task ID"})
		return
	}

	var req humantask.SubmitTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err = h.service.SubmitTask(c.Request.Context(), tenantID, taskID, userID, roles, req)
	if err != nil {
		statusCode := getErrorStatusCode(err)
		c.JSON(statusCode, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "task submitted successfully"})
}

// GetEscalationHistory godoc
// @Summary Get escalation history for a task
// @Description Get the escalation history for a human task
// @Tags tasks
// @Accept json
// @Produce json
// @Param id path string true "Task ID"
// @Success 200 {object} humantask.EscalationHistory
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/tasks/{id}/escalations [get]
func (h *HumanTaskHandler) GetEscalationHistory(c *gin.Context) {
	tenantID, err := getTenantID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	taskID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid task ID"})
		return
	}

	history, err := h.service.GetEscalationHistory(c.Request.Context(), tenantID, taskID)
	if err != nil {
		if errors.Is(err, humantask.ErrTaskNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "task not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, history)
}

// UpdateEscalationConfig godoc
// @Summary Update escalation configuration for a task
// @Description Update the escalation configuration for a pending human task
// @Tags tasks
// @Accept json
// @Produce json
// @Param id path string true "Task ID"
// @Param request body humantask.UpdateEscalationRequest true "Escalation config"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 409 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/tasks/{id}/escalation [put]
func (h *HumanTaskHandler) UpdateEscalationConfig(c *gin.Context) {
	tenantID, err := getTenantID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	taskID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid task ID"})
		return
	}

	var req humantask.UpdateEscalationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate escalation config
	if len(req.Config.Levels) == 0 && req.Config.Enabled {
		c.JSON(http.StatusBadRequest, gin.H{"error": "at least one escalation level is required when escalation is enabled"})
		return
	}

	err = h.service.UpdateEscalationConfig(c.Request.Context(), tenantID, taskID, req)
	if err != nil {
		statusCode := getErrorStatusCode(err)
		c.JSON(statusCode, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "escalation configuration updated successfully"})
}

// Helper functions

func getTenantID(c *gin.Context) (uuid.UUID, error) {
	tenantIDStr, exists := c.Get("tenant_id")
	if !exists {
		return uuid.Nil, errors.New("tenant_id not found in context")
	}

	tenantID, err := uuid.Parse(tenantIDStr.(string))
	if err != nil {
		return uuid.Nil, errors.New("invalid tenant_id")
	}

	return tenantID, nil
}

func getUserID(c *gin.Context) (uuid.UUID, error) {
	userIDStr, exists := c.Get("user_id")
	if !exists {
		return uuid.Nil, errors.New("user_id not found in context")
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		return uuid.Nil, errors.New("invalid user_id")
	}

	return userID, nil
}

func getUserRoles(c *gin.Context) []string {
	roles, exists := c.Get("user_roles")
	if !exists {
		return []string{}
	}

	if roleSlice, ok := roles.([]string); ok {
		return roleSlice
	}

	return []string{}
}

func getAuthContext(c *gin.Context) (tenantID uuid.UUID, userID uuid.UUID, roles []string, err error) {
	tenantID, err = getTenantID(c)
	if err != nil {
		return
	}

	userID, err = getUserID(c)
	if err != nil {
		return
	}

	roles = getUserRoles(c)
	return
}

func getErrorStatusCode(err error) int {
	switch {
	case errors.Is(err, humantask.ErrTaskNotFound):
		return http.StatusNotFound
	case errors.Is(err, humantask.ErrUnauthorized):
		return http.StatusForbidden
	case errors.Is(err, humantask.ErrTaskNotPending):
		return http.StatusConflict
	case errors.Is(err, humantask.ErrInvalidTaskType),
		errors.Is(err, humantask.ErrInvalidStatus),
		errors.Is(err, humantask.ErrMissingRequiredField):
		return http.StatusBadRequest
	default:
		return http.StatusInternalServerError
	}
}
