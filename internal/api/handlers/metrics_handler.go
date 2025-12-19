package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gorax/gorax/internal/api/middleware"
	"github.com/gorax/gorax/internal/workflow"
)

// MetricsHandler handles metrics-related HTTP requests
type MetricsHandler struct {
	repo *workflow.Repository
}

// NewMetricsHandler creates a new metrics handler
func NewMetricsHandler(repo *workflow.Repository) *MetricsHandler {
	return &MetricsHandler{
		repo: repo,
	}
}

// GetExecutionTrends returns execution counts grouped by time period
// GET /api/v1/metrics/trends?groupBy=day&days=7&startDate=2024-01-01&endDate=2024-01-31
func (h *MetricsHandler) GetExecutionTrends(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r)
	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant_id required")
		return
	}

	groupBy := r.URL.Query().Get("groupBy")
	if groupBy == "" {
		groupBy = "day"
	}
	if groupBy != "hour" && groupBy != "day" {
		h.respondError(w, http.StatusBadRequest, "groupBy must be 'hour' or 'day'")
		return
	}

	startDate, endDate, err := h.parseDateRange(r)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	trends, err := h.repo.GetExecutionTrends(r.Context(), tenantID, startDate, endDate, groupBy)
	if err != nil {
		h.respondError(w, http.StatusInternalServerError, "failed to get execution trends")
		return
	}

	h.respondJSON(w, http.StatusOK, map[string]interface{}{
		"trends":    trends,
		"startDate": startDate.Format(time.RFC3339),
		"endDate":   endDate.Format(time.RFC3339),
		"groupBy":   groupBy,
	})
}

// GetDurationStats returns duration statistics by workflow
// GET /api/v1/metrics/duration?days=30&startDate=2024-01-01&endDate=2024-01-31
func (h *MetricsHandler) GetDurationStats(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r)
	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant_id required")
		return
	}

	startDate, endDate, err := h.parseDateRange(r)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	stats, err := h.repo.GetDurationStats(r.Context(), tenantID, startDate, endDate)
	if err != nil {
		h.respondError(w, http.StatusInternalServerError, "failed to get duration stats")
		return
	}

	h.respondJSON(w, http.StatusOK, map[string]interface{}{
		"stats":     stats,
		"startDate": startDate.Format(time.RFC3339),
		"endDate":   endDate.Format(time.RFC3339),
	})
}

// GetTopFailures returns workflows with the most failures
// GET /api/v1/metrics/failures?limit=10&days=30
func (h *MetricsHandler) GetTopFailures(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r)
	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant_id required")
		return
	}

	limit := 10
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		parsedLimit, err := strconv.Atoi(limitStr)
		if err != nil || parsedLimit <= 0 || parsedLimit > 100 {
			h.respondError(w, http.StatusBadRequest, "limit must be between 1 and 100")
			return
		}
		limit = parsedLimit
	}

	startDate, endDate, err := h.parseDateRange(r)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	failures, err := h.repo.GetTopFailures(r.Context(), tenantID, startDate, endDate, limit)
	if err != nil {
		h.respondError(w, http.StatusInternalServerError, "failed to get top failures")
		return
	}

	h.respondJSON(w, http.StatusOK, map[string]interface{}{
		"failures":  failures,
		"startDate": startDate.Format(time.RFC3339),
		"endDate":   endDate.Format(time.RFC3339),
		"limit":     limit,
	})
}

// GetTriggerBreakdown returns execution distribution by trigger type
// GET /api/v1/metrics/trigger-breakdown?days=30
func (h *MetricsHandler) GetTriggerBreakdown(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r)
	if tenantID == "" {
		h.respondError(w, http.StatusUnauthorized, "tenant_id required")
		return
	}

	startDate, endDate, err := h.parseDateRange(r)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	breakdown, err := h.repo.GetTriggerTypeBreakdown(r.Context(), tenantID, startDate, endDate)
	if err != nil {
		h.respondError(w, http.StatusInternalServerError, "failed to get trigger breakdown")
		return
	}

	h.respondJSON(w, http.StatusOK, map[string]interface{}{
		"breakdown": breakdown,
		"startDate": startDate.Format(time.RFC3339),
		"endDate":   endDate.Format(time.RFC3339),
	})
}

// parseDateRange parses start and end dates from query parameters
// Supports: days (default 7), or explicit startDate/endDate
func (h *MetricsHandler) parseDateRange(r *http.Request) (time.Time, time.Time, error) {
	now := time.Now()
	endDate := now

	// Check for explicit dates first
	if startDateStr := r.URL.Query().Get("startDate"); startDateStr != "" {
		parsedStart, err := parseDate(startDateStr)
		if err != nil {
			return time.Time{}, time.Time{}, err
		}

		if endDateStr := r.URL.Query().Get("endDate"); endDateStr != "" {
			parsedEnd, err := parseDate(endDateStr)
			if err != nil {
				return time.Time{}, time.Time{}, err
			}
			return parsedStart, parsedEnd, nil
		}

		// If only startDate provided, use now as endDate
		return parsedStart, endDate, nil
	}

	// Use days parameter (default 7)
	days := 7
	if daysStr := r.URL.Query().Get("days"); daysStr != "" {
		parsedDays, err := strconv.Atoi(daysStr)
		if err != nil || parsedDays <= 0 {
			return time.Time{}, time.Time{}, fmt.Errorf("invalid days parameter")
		}
		days = parsedDays
	}

	startDate := now.Add(-time.Duration(days) * 24 * time.Hour)
	return startDate, endDate, nil
}

// parseDate parses a date string in multiple formats
func parseDate(dateStr string) (time.Time, error) {
	formats := []string{
		"2006-01-02",
		time.RFC3339,
		"2006-01-02T15:04:05Z",
		"2006-01-02 15:04:05",
	}

	for _, format := range formats {
		if t, err := time.Parse(format, dateStr); err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("invalid date format: %s", dateStr)
}

// Helper methods
func (h *MetricsHandler) respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		// Log error but can't change response at this point
		_ = err
	}
}

func (h *MetricsHandler) respondError(w http.ResponseWriter, status int, message string) {
	h.respondJSON(w, status, map[string]string{
		"error": message,
	})
}
