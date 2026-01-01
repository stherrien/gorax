package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/gorax/gorax/internal/analytics"
	"github.com/gorax/gorax/internal/api/middleware"
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
// GET /api/v1/analytics/overview
func (h *AnalyticsHandler) GetTenantOverview(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r)
	if tenantID == "" {
		h.respondError(w, http.StatusInternalServerError, "tenant ID not found")
		return
	}

	timeRange, err := h.parseTimeRange(r)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid time range: "+err.Error())
		return
	}

	overview, err := h.service.GetTenantOverview(r.Context(), tenantID, timeRange)
	if err != nil {
		h.logger.Error("failed to get tenant overview",
			"error", err,
			"tenant_id", tenantID,
		)
		h.respondError(w, http.StatusInternalServerError, "failed to get tenant overview")
		return
	}

	h.respondJSON(w, http.StatusOK, overview)
}

// GetWorkflowStats retrieves analytics for a specific workflow
// GET /api/v1/analytics/workflows/:workflowID
func (h *AnalyticsHandler) GetWorkflowStats(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r)
	if tenantID == "" {
		h.respondError(w, http.StatusInternalServerError, "tenant ID not found")
		return
	}

	workflowID := chi.URLParam(r, "workflowID")
	if workflowID == "" {
		h.respondError(w, http.StatusBadRequest, "workflow ID is required")
		return
	}

	timeRange, err := h.parseTimeRange(r)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid time range: "+err.Error())
		return
	}

	stats, err := h.service.GetWorkflowStats(r.Context(), tenantID, workflowID, timeRange)
	if err != nil {
		if err == analytics.ErrNotFound {
			h.respondError(w, http.StatusNotFound, "workflow not found")
			return
		}
		h.logger.Error("failed to get workflow stats",
			"error", err,
			"tenant_id", tenantID,
			"workflow_id", workflowID,
		)
		h.respondError(w, http.StatusInternalServerError, "failed to get workflow stats")
		return
	}

	h.respondJSON(w, http.StatusOK, stats)
}

// GetExecutionTrends retrieves execution trends over time
// GET /api/v1/analytics/trends
func (h *AnalyticsHandler) GetExecutionTrends(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r)
	if tenantID == "" {
		h.respondError(w, http.StatusInternalServerError, "tenant ID not found")
		return
	}

	timeRange, err := h.parseTimeRange(r)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid time range: "+err.Error())
		return
	}

	granularity := analytics.Granularity(r.URL.Query().Get("granularity"))
	if granularity == "" {
		granularity = analytics.GranularityDay
	}

	if !isValidGranularity(granularity) {
		h.respondError(w, http.StatusBadRequest, "invalid granularity: must be hour, day, week, or month")
		return
	}

	trends, err := h.service.GetExecutionTrends(r.Context(), tenantID, timeRange, granularity)
	if err != nil {
		h.logger.Error("failed to get execution trends",
			"error", err,
			"tenant_id", tenantID,
		)
		h.respondError(w, http.StatusInternalServerError, "failed to get execution trends")
		return
	}

	h.respondJSON(w, http.StatusOK, trends)
}

// GetTopWorkflows retrieves the most frequently executed workflows
// GET /api/v1/analytics/top-workflows
func (h *AnalyticsHandler) GetTopWorkflows(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r)
	if tenantID == "" {
		h.respondError(w, http.StatusInternalServerError, "tenant ID not found")
		return
	}

	timeRange, err := h.parseTimeRange(r)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid time range: "+err.Error())
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
		h.respondError(w, http.StatusInternalServerError, "failed to get top workflows")
		return
	}

	h.respondJSON(w, http.StatusOK, workflows)
}

// GetErrorBreakdown retrieves error analysis
// GET /api/v1/analytics/errors
func (h *AnalyticsHandler) GetErrorBreakdown(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r)
	if tenantID == "" {
		h.respondError(w, http.StatusInternalServerError, "tenant ID not found")
		return
	}

	timeRange, err := h.parseTimeRange(r)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid time range: "+err.Error())
		return
	}

	breakdown, err := h.service.GetErrorBreakdown(r.Context(), tenantID, timeRange)
	if err != nil {
		h.logger.Error("failed to get error breakdown",
			"error", err,
			"tenant_id", tenantID,
		)
		h.respondError(w, http.StatusInternalServerError, "failed to get error breakdown")
		return
	}

	h.respondJSON(w, http.StatusOK, breakdown)
}

// GetNodePerformance retrieves node-level performance statistics
// GET /api/v1/analytics/workflows/:workflowID/nodes
func (h *AnalyticsHandler) GetNodePerformance(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r)
	if tenantID == "" {
		h.respondError(w, http.StatusInternalServerError, "tenant ID not found")
		return
	}

	workflowID := chi.URLParam(r, "workflowID")
	if workflowID == "" {
		h.respondError(w, http.StatusBadRequest, "workflow ID is required")
		return
	}

	performance, err := h.service.GetNodePerformance(r.Context(), tenantID, workflowID)
	if err != nil {
		if err == analytics.ErrNotFound {
			h.respondError(w, http.StatusNotFound, "workflow not found")
			return
		}
		h.logger.Error("failed to get node performance",
			"error", err,
			"tenant_id", tenantID,
			"workflow_id", workflowID,
		)
		h.respondError(w, http.StatusInternalServerError, "failed to get node performance")
		return
	}

	h.respondJSON(w, http.StatusOK, performance)
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

// respondJSON sends a JSON response
func (h *AnalyticsHandler) respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		h.logger.Error("failed to encode JSON response", "error", err)
	}
}

// respondError sends an error response
func (h *AnalyticsHandler) respondError(w http.ResponseWriter, status int, message string) {
	h.respondJSON(w, status, map[string]string{"error": message})
}
