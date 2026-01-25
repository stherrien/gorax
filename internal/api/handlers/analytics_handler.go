package handlers

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/gorax/gorax/internal/analytics"
	"github.com/gorax/gorax/internal/api/middleware"
	"github.com/gorax/gorax/internal/api/response"
)

// AnalyticsService defines the interface for analytics operations
type AnalyticsService interface {
	GetWorkflowStats(ctx context.Context, tenantID, workflowID string, timeRange analytics.TimeRange) (*analytics.WorkflowStats, error)
	GetTenantOverview(ctx context.Context, tenantID string, timeRange analytics.TimeRange) (*analytics.TenantOverview, error)
	GetExecutionTrends(ctx context.Context, tenantID string, timeRange analytics.TimeRange, granularity analytics.Granularity) (*analytics.ExecutionTrends, error)
	GetTopWorkflows(ctx context.Context, tenantID string, timeRange analytics.TimeRange, limit int) (*analytics.TopWorkflows, error)
	GetErrorBreakdown(ctx context.Context, tenantID string, timeRange analytics.TimeRange) (*analytics.ErrorBreakdown, error)
	GetNodePerformance(ctx context.Context, tenantID, workflowID string) (*analytics.NodePerformance, error)
}

// AnalyticsHandler handles analytics-related HTTP requests
type AnalyticsHandler struct {
	service AnalyticsService
	logger  *slog.Logger
}

// NewAnalyticsHandler creates a new analytics handler
func NewAnalyticsHandler(service AnalyticsService, logger *slog.Logger) *AnalyticsHandler {
	return &AnalyticsHandler{
		service: service,
		logger:  logger,
	}
}

// GetTenantOverview retrieves overall analytics for a tenant
// @Summary Get tenant analytics overview
// @Description Returns aggregated analytics data for the tenant including total executions, success rates, and trends
// @Tags Analytics
// @Accept json
// @Produce json
// @Param start_date query string true "Start date (RFC3339 format)" example(2024-01-01T00:00:00Z)
// @Param end_date query string true "End date (RFC3339 format)" example(2024-01-31T23:59:59Z)
// @Security TenantID
// @Security UserID
// @Success 200 {object} analytics.TenantOverview "Tenant analytics overview"
// @Failure 400 {object} map[string]string "Invalid time range"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/v1/analytics/overview [get]
func (h *AnalyticsHandler) GetTenantOverview(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r)
	if tenantID == "" {
		_ = response.InternalError(w, "tenant ID not found")
		return
	}

	timeRange, err := h.parseTimeRange(r)
	if err != nil {
		_ = response.BadRequest(w, "invalid time range: "+err.Error())
		return
	}

	overview, err := h.service.GetTenantOverview(r.Context(), tenantID, timeRange)
	if err != nil {
		h.logger.Error("failed to get tenant overview",
			"error", err,
			"tenant_id", tenantID,
		)
		_ = response.InternalError(w, "failed to get tenant overview")
		return
	}

	_ = response.OK(w, overview)
}

// GetWorkflowStats retrieves analytics for a specific workflow
// @Summary Get workflow-specific analytics
// @Description Returns detailed analytics for a specific workflow including execution counts, success rate, and average duration
// @Tags Analytics
// @Accept json
// @Produce json
// @Param workflowID path string true "Workflow ID"
// @Param start_date query string true "Start date (RFC3339 format)" example(2024-01-01T00:00:00Z)
// @Param end_date query string true "End date (RFC3339 format)" example(2024-01-31T23:59:59Z)
// @Security TenantID
// @Security UserID
// @Success 200 {object} analytics.WorkflowStats "Workflow analytics"
// @Failure 400 {object} map[string]string "Invalid parameters"
// @Failure 404 {object} map[string]string "Workflow not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/v1/analytics/workflows/{workflowID} [get]
func (h *AnalyticsHandler) GetWorkflowStats(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r)
	if tenantID == "" {
		_ = response.InternalError(w, "tenant ID not found")
		return
	}

	workflowID := chi.URLParam(r, "workflowID")
	if workflowID == "" {
		_ = response.BadRequest(w, "workflow ID is required")
		return
	}

	timeRange, err := h.parseTimeRange(r)
	if err != nil {
		_ = response.BadRequest(w, "invalid time range: "+err.Error())
		return
	}

	stats, err := h.service.GetWorkflowStats(r.Context(), tenantID, workflowID, timeRange)
	if err != nil {
		if err == analytics.ErrNotFound {
			_ = response.NotFound(w, "workflow not found")
			return
		}
		h.logger.Error("failed to get workflow stats",
			"error", err,
			"tenant_id", tenantID,
			"workflow_id", workflowID,
		)
		_ = response.InternalError(w, "failed to get workflow stats")
		return
	}

	_ = response.OK(w, stats)
}

// GetExecutionTrends retrieves execution trends over time
// @Summary Get execution trends
// @Description Returns time-series data showing execution trends with configurable granularity (hour, day, week, month)
// @Tags Analytics
// @Accept json
// @Produce json
// @Param start_date query string true "Start date (RFC3339 format)" example(2024-01-01T00:00:00Z)
// @Param end_date query string true "End date (RFC3339 format)" example(2024-01-31T23:59:59Z)
// @Param granularity query string false "Time granularity (hour, day, week, month)" default(day) Enums(hour, day, week, month)
// @Security TenantID
// @Security UserID
// @Success 200 {object} analytics.ExecutionTrends "Execution trends data"
// @Failure 400 {object} map[string]string "Invalid parameters"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/v1/analytics/trends [get]
func (h *AnalyticsHandler) GetExecutionTrends(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r)
	if tenantID == "" {
		_ = response.InternalError(w, "tenant ID not found")
		return
	}

	timeRange, err := h.parseTimeRange(r)
	if err != nil {
		_ = response.BadRequest(w, "invalid time range: "+err.Error())
		return
	}

	granularity := analytics.Granularity(r.URL.Query().Get("granularity"))
	if granularity == "" {
		granularity = analytics.GranularityDay
	}

	if !isValidGranularity(granularity) {
		_ = response.BadRequest(w, "invalid granularity: must be hour, day, week, or month")
		return
	}

	trends, err := h.service.GetExecutionTrends(r.Context(), tenantID, timeRange, granularity)
	if err != nil {
		h.logger.Error("failed to get execution trends",
			"error", err,
			"tenant_id", tenantID,
		)
		_ = response.InternalError(w, "failed to get execution trends")
		return
	}

	_ = response.OK(w, trends)
}

// GetTopWorkflows retrieves the most frequently executed workflows
// @Summary Get top workflows by execution count
// @Description Returns the most frequently executed workflows ordered by execution count
// @Tags Analytics
// @Accept json
// @Produce json
// @Param start_date query string true "Start date (RFC3339 format)" example(2024-01-01T00:00:00Z)
// @Param end_date query string true "End date (RFC3339 format)" example(2024-01-31T23:59:59Z)
// @Param limit query int false "Maximum number of workflows to return" default(10)
// @Security TenantID
// @Security UserID
// @Success 200 {object} analytics.TopWorkflows "Top workflows"
// @Failure 400 {object} map[string]string "Invalid parameters"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/v1/analytics/top-workflows [get]
func (h *AnalyticsHandler) GetTopWorkflows(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r)
	if tenantID == "" {
		_ = response.InternalError(w, "tenant ID not found")
		return
	}

	timeRange, err := h.parseTimeRange(r)
	if err != nil {
		_ = response.BadRequest(w, "invalid time range: "+err.Error())
		return
	}

	limit := 10
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		parsedLimit, err := strconv.Atoi(limitStr)
		if err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	workflows, err := h.service.GetTopWorkflows(r.Context(), tenantID, timeRange, limit)
	if err != nil {
		h.logger.Error("failed to get top workflows",
			"error", err,
			"tenant_id", tenantID,
		)
		_ = response.InternalError(w, "failed to get top workflows")
		return
	}

	_ = response.OK(w, workflows)
}

// GetErrorBreakdown retrieves error analysis
// @Summary Get error breakdown and analysis
// @Description Returns categorized error data including error types, frequencies, and affected workflows
// @Tags Analytics
// @Accept json
// @Produce json
// @Param start_date query string true "Start date (RFC3339 format)" example(2024-01-01T00:00:00Z)
// @Param end_date query string true "End date (RFC3339 format)" example(2024-01-31T23:59:59Z)
// @Security TenantID
// @Security UserID
// @Success 200 {object} analytics.ErrorBreakdown "Error breakdown data"
// @Failure 400 {object} map[string]string "Invalid parameters"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/v1/analytics/errors [get]
func (h *AnalyticsHandler) GetErrorBreakdown(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r)
	if tenantID == "" {
		_ = response.InternalError(w, "tenant ID not found")
		return
	}

	timeRange, err := h.parseTimeRange(r)
	if err != nil {
		_ = response.BadRequest(w, "invalid time range: "+err.Error())
		return
	}

	breakdown, err := h.service.GetErrorBreakdown(r.Context(), tenantID, timeRange)
	if err != nil {
		h.logger.Error("failed to get error breakdown",
			"error", err,
			"tenant_id", tenantID,
		)
		_ = response.InternalError(w, "failed to get error breakdown")
		return
	}

	_ = response.OK(w, breakdown)
}

// GetNodePerformance retrieves node-level performance statistics
// @Summary Get node performance metrics
// @Description Returns performance statistics for individual nodes within a workflow including execution time and success rate
// @Tags Analytics
// @Accept json
// @Produce json
// @Param workflowID path string true "Workflow ID"
// @Security TenantID
// @Security UserID
// @Success 200 {object} analytics.NodePerformance "Node performance metrics"
// @Failure 400 {object} map[string]string "Invalid workflow ID"
// @Failure 404 {object} map[string]string "Workflow not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/v1/analytics/workflows/{workflowID}/nodes [get]
func (h *AnalyticsHandler) GetNodePerformance(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r)
	if tenantID == "" {
		_ = response.InternalError(w, "tenant ID not found")
		return
	}

	workflowID := chi.URLParam(r, "workflowID")
	if workflowID == "" {
		_ = response.BadRequest(w, "workflow ID is required")
		return
	}

	performance, err := h.service.GetNodePerformance(r.Context(), tenantID, workflowID)
	if err != nil {
		if err == analytics.ErrNotFound {
			_ = response.NotFound(w, "workflow not found")
			return
		}
		h.logger.Error("failed to get node performance",
			"error", err,
			"tenant_id", tenantID,
			"workflow_id", workflowID,
		)
		_ = response.InternalError(w, "failed to get node performance")
		return
	}

	_ = response.OK(w, performance)
}

// parseTimeRange parses start_date and end_date from query parameters
func (h *AnalyticsHandler) parseTimeRange(r *http.Request) (analytics.TimeRange, error) {
	startDateStr := r.URL.Query().Get("start_date")
	endDateStr := r.URL.Query().Get("end_date")

	if startDateStr == "" || endDateStr == "" {
		return analytics.TimeRange{}, errors.New("start_date and end_date are required")
	}

	startDate, err := time.Parse(time.RFC3339, startDateStr)
	if err != nil {
		return analytics.TimeRange{}, errors.New("invalid start_date format, use RFC3339")
	}

	endDate, err := time.Parse(time.RFC3339, endDateStr)
	if err != nil {
		return analytics.TimeRange{}, errors.New("invalid end_date format, use RFC3339")
	}

	return analytics.TimeRange{
		StartDate: startDate,
		EndDate:   endDate,
	}, nil
}

// isValidGranularity checks if the granularity is valid
func isValidGranularity(granularity analytics.Granularity) bool {
	switch granularity {
	case analytics.GranularityHour, analytics.GranularityDay, analytics.GranularityWeek, analytics.GranularityMonth:
		return true
	default:
		return false
	}
}
