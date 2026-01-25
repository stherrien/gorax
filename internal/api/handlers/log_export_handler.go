package handlers

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"

	"github.com/gorax/gorax/internal/api/middleware"
	"github.com/gorax/gorax/internal/api/response"
	"github.com/gorax/gorax/internal/workflow"
)

// LogExportHandler handles log export HTTP requests
type LogExportHandler struct {
	service LogExportService
	logger  *slog.Logger
}

// LogExportService defines the interface for log export operations
type LogExportService interface {
	GetExecutionWithSteps(ctx context.Context, tenantID, executionID string) (*workflow.Execution, []*workflow.StepExecution, error)
	ExportLogs(execution *workflow.Execution, steps []*workflow.StepExecution, format string) ([]byte, string, error)
}

// NewLogExportHandler creates a new log export handler
func NewLogExportHandler(service LogExportService, logger *slog.Logger) *LogExportHandler {
	return &LogExportHandler{
		service: service,
		logger:  logger,
	}
}

// ExportExecutionLogs exports execution logs in specified format
// GET /api/v1/executions/{id}/logs/export?format={txt|json|csv}
func (h *LogExportHandler) ExportExecutionLogs(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r)
	executionID := chi.URLParam(r, "id")
	format := getExportFormat(r)

	if !isValidFormat(format) {
		_ = response.BadRequest(w, "invalid format: must be txt, json, or csv")
		return
	}

	execution, steps, err := h.service.GetExecutionWithSteps(r.Context(), tenantID, executionID)
	if err != nil {
		h.handleExportError(w, err)
		return
	}

	data, contentType, err := h.service.ExportLogs(execution, steps, format)
	if err != nil {
		_ = response.InternalError(w, "failed to export logs")
		return
	}

	filename := createFilename(executionID, format)
	setDownloadHeaders(w, contentType, filename)

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
}

// getExportFormat retrieves the export format from query parameters
func getExportFormat(r *http.Request) string {
	format := r.URL.Query().Get("format")
	if format == "" {
		return "txt"
	}
	return strings.ToLower(format)
}

// isValidFormat checks if the format is valid
func isValidFormat(format string) bool {
	validFormats := map[string]bool{
		"txt":  true,
		"json": true,
		"csv":  true,
	}
	return validFormats[format]
}

// createFilename creates a filename for the export
func createFilename(executionID, format string) string {
	return fmt.Sprintf("%s.%s", executionID, format)
}

// setDownloadHeaders sets response headers for file download
func setDownloadHeaders(w http.ResponseWriter, contentType, filename string) {
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
}

// handleExportError handles errors from GetExecutionWithSteps
func (h *LogExportHandler) handleExportError(w http.ResponseWriter, err error) {
	if strings.Contains(err.Error(), "not found") {
		_ = response.NotFound(w, "execution not found")
		return
	}
	_ = response.InternalError(w, "failed to retrieve execution")
}
