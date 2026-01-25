package handlers

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/gorax/gorax/internal/audit"
)

// AuditService defines the interface for audit service operations
type AuditService interface {
	QueryAuditEvents(ctx context.Context, filter audit.QueryFilter) ([]audit.AuditEvent, int, error)
	GetAuditEvent(ctx context.Context, tenantID, eventID string) (*audit.AuditEvent, error)
	GetAuditStats(ctx context.Context, tenantID string, timeRange audit.TimeRange) (*audit.AuditStats, error)
}

// AuditHandler handles audit log HTTP requests
type AuditHandler struct {
	service AuditService
	logger  *slog.Logger
}

// NewAuditHandler creates a new audit handler
func NewAuditHandler(service AuditService, logger *slog.Logger) *AuditHandler {
	return &AuditHandler{
		service: service,
		logger:  logger,
	}
}

// QueryEvents queries audit events with filters
// GET /api/v1/admin/audit/events
func (h *AuditHandler) QueryEvents(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get tenant ID from context (admin can query across tenants or specific tenant)
	tenantID := r.URL.Query().Get("tenant_id")

	// Parse query parameters
	filter := audit.QueryFilter{
		TenantID:      tenantID,
		UserID:        r.URL.Query().Get("user_id"),
		UserEmail:     r.URL.Query().Get("user_email"),
		Limit:         getIntParam(r, "limit", 50),
		Offset:        getIntParam(r, "offset", 0),
		SortBy:        r.URL.Query().Get("sort_by"),
		SortDirection: r.URL.Query().Get("sort_direction"),
	}

	// Parse array parameters
	if categories := r.URL.Query()["category"]; len(categories) > 0 {
		filter.Categories = make([]audit.Category, len(categories))
		for i, c := range categories {
			filter.Categories[i] = audit.Category(c)
		}
	}

	if eventTypes := r.URL.Query()["event_type"]; len(eventTypes) > 0 {
		filter.EventTypes = make([]audit.EventType, len(eventTypes))
		for i, et := range eventTypes {
			filter.EventTypes[i] = audit.EventType(et)
		}
	}

	if severities := r.URL.Query()["severity"]; len(severities) > 0 {
		filter.Severities = make([]audit.Severity, len(severities))
		for i, s := range severities {
			filter.Severities[i] = audit.Severity(s)
		}
	}

	if statuses := r.URL.Query()["status"]; len(statuses) > 0 {
		filter.Statuses = make([]audit.Status, len(statuses))
		for i, s := range statuses {
			filter.Statuses[i] = audit.Status(s)
		}
	}

	// Parse date range
	if startDate := r.URL.Query().Get("start_date"); startDate != "" {
		if t, err := time.Parse(time.RFC3339, startDate); err == nil {
			filter.StartDate = t
		}
	}

	if endDate := r.URL.Query().Get("end_date"); endDate != "" {
		if t, err := time.Parse(time.RFC3339, endDate); err == nil {
			filter.EndDate = t
		}
	}

	// Query events
	events, total, err := h.service.QueryAuditEvents(ctx, filter)
	if err != nil {
		h.logger.Error("failed to query audit events", "error", err)
		http.Error(w, fmt.Sprintf("Failed to query events: %v", err), http.StatusInternalServerError)
		return
	}

	// Return response
	response := map[string]interface{}{
		"events": events,
		"total":  total,
		"limit":  filter.Limit,
		"offset": filter.Offset,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("failed to encode response", "error", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// GetEvent retrieves a single audit event
// GET /api/v1/admin/audit/events/{id}
func (h *AuditHandler) GetEvent(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	eventID := chi.URLParam(r, "id")
	tenantID := r.URL.Query().Get("tenant_id")

	if eventID == "" {
		http.Error(w, "Event ID is required", http.StatusBadRequest)
		return
	}

	event, err := h.service.GetAuditEvent(ctx, tenantID, eventID)
	if err != nil {
		h.logger.Error("failed to get audit event", "error", err, "event_id", eventID)
		http.Error(w, "Event not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(event); err != nil {
		h.logger.Error("failed to encode response", "error", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// GetStats retrieves audit log statistics
// GET /api/v1/admin/audit/stats
func (h *AuditHandler) GetStats(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID := r.URL.Query().Get("tenant_id")

	// Parse time range
	var timeRange audit.TimeRange
	if startDate := r.URL.Query().Get("start_date"); startDate != "" {
		if t, err := time.Parse(time.RFC3339, startDate); err == nil {
			timeRange.StartDate = t
		} else {
			timeRange.StartDate = time.Now().Add(-24 * time.Hour)
		}
	} else {
		timeRange.StartDate = time.Now().Add(-24 * time.Hour)
	}

	if endDate := r.URL.Query().Get("end_date"); endDate != "" {
		if t, err := time.Parse(time.RFC3339, endDate); err == nil {
			timeRange.EndDate = t
		} else {
			timeRange.EndDate = time.Now()
		}
	} else {
		timeRange.EndDate = time.Now()
	}

	stats, err := h.service.GetAuditStats(ctx, tenantID, timeRange)
	if err != nil {
		h.logger.Error("failed to get audit stats", "error", err)
		http.Error(w, fmt.Sprintf("Failed to get stats: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(stats); err != nil {
		h.logger.Error("failed to encode response", "error", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// ExportEvents exports audit events in CSV or JSON format
// POST /api/v1/admin/audit/export
func (h *AuditHandler) ExportEvents(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse request body
	var req struct {
		TenantID   string             `json:"tenantId"`
		Format     audit.ExportFormat `json:"format"`
		Categories []audit.Category   `json:"categories"`
		EventTypes []audit.EventType  `json:"eventTypes"`
		Severities []audit.Severity   `json:"severities"`
		StartDate  string             `json:"startDate"`
		EndDate    string             `json:"endDate"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
		return
	}

	// Default format to JSON
	if req.Format == "" {
		req.Format = audit.ExportFormatJSON
	}

	// Build filter
	filter := audit.QueryFilter{
		TenantID:   req.TenantID,
		Categories: req.Categories,
		EventTypes: req.EventTypes,
		Severities: req.Severities,
		Limit:      10000, // Max export limit
	}

	if req.StartDate != "" {
		if t, err := time.Parse(time.RFC3339, req.StartDate); err == nil {
			filter.StartDate = t
		}
	}

	if req.EndDate != "" {
		if t, err := time.Parse(time.RFC3339, req.EndDate); err == nil {
			filter.EndDate = t
		}
	}

	// Query events
	events, _, err := h.service.QueryAuditEvents(ctx, filter)
	if err != nil {
		h.logger.Error("failed to query audit events for export", "error", err)
		http.Error(w, fmt.Sprintf("Failed to export events: %v", err), http.StatusInternalServerError)
		return
	}

	// Export based on format
	switch req.Format {
	case audit.ExportFormatCSV:
		h.exportCSV(w, events)
	case audit.ExportFormatJSON:
		h.exportJSON(w, events)
	default:
		http.Error(w, "Unsupported export format", http.StatusBadRequest)
	}
}

func (h *AuditHandler) exportJSON(w http.ResponseWriter, events []audit.AuditEvent) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=audit-export-%s.json", time.Now().Format("2006-01-02")))

	if err := json.NewEncoder(w).Encode(events); err != nil {
		h.logger.Error("failed to encode JSON export", "error", err)
		http.Error(w, "Failed to export events", http.StatusInternalServerError)
		return
	}
}

func (h *AuditHandler) exportCSV(w http.ResponseWriter, events []audit.AuditEvent) {
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=audit-export-%s.csv", time.Now().Format("2006-01-02")))

	writer := csv.NewWriter(w)
	defer writer.Flush()

	// Write header
	header := []string{
		"ID", "Tenant ID", "User ID", "User Email", "Category", "Event Type",
		"Action", "Resource Type", "Resource ID", "Resource Name", "IP Address",
		"Severity", "Status", "Error Message", "Created At",
	}
	if err := writer.Write(header); err != nil {
		h.logger.Error("failed to write CSV header", "error", err)
		return
	}

	// Write rows
	for _, event := range events {
		row := []string{
			event.ID,
			event.TenantID,
			event.UserID,
			event.UserEmail,
			string(event.Category),
			string(event.EventType),
			event.Action,
			event.ResourceType,
			event.ResourceID,
			event.ResourceName,
			event.IPAddress,
			string(event.Severity),
			string(event.Status),
			event.ErrorMessage,
			event.CreatedAt.Format(time.RFC3339),
		}
		if err := writer.Write(row); err != nil {
			h.logger.Error("failed to write CSV row", "error", err)
			return
		}
	}
}

// Helper function to parse integer query parameters
func getIntParam(r *http.Request, key string, defaultValue int) int {
	if val := r.URL.Query().Get(key); val != "" {
		if intVal, err := strconv.Atoi(val); err == nil {
			return intVal
		}
	}
	return defaultValue
}
