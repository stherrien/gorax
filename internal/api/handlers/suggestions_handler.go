package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/gorax/gorax/internal/api/middleware"
	"github.com/gorax/gorax/internal/api/response"
	"github.com/gorax/gorax/internal/suggestions"
)

// Context key type for test fallback
type suggestionsContextKey string

const suggestionsTenantIDKey suggestionsContextKey = "tenant_id"

// SuggestionService interface for the handler
type SuggestionService interface {
	AnalyzeError(ctx context.Context, tenantID string, errCtx *suggestions.ErrorContext) ([]*suggestions.Suggestion, error)
	GetSuggestions(ctx context.Context, tenantID, executionID string) ([]*suggestions.Suggestion, error)
	GetSuggestionByID(ctx context.Context, tenantID, suggestionID string) (*suggestions.Suggestion, error)
	ApplySuggestion(ctx context.Context, tenantID, suggestionID string) error
	DismissSuggestion(ctx context.Context, tenantID, suggestionID string) error
}

// SuggestionsHandler handles suggestion-related HTTP requests
type SuggestionsHandler struct {
	service SuggestionService
	logger  *slog.Logger
}

// NewSuggestionsHandler creates a new suggestions handler
func NewSuggestionsHandler(service SuggestionService, logger *slog.Logger) *SuggestionsHandler {
	if logger == nil {
		logger = slog.Default()
	}
	return &SuggestionsHandler{
		service: service,
		logger:  logger,
	}
}

// AnalyzeRequest is the request body for error analysis
type AnalyzeRequest struct {
	WorkflowID   string         `json:"workflow_id"`
	NodeID       string         `json:"node_id"`
	NodeType     string         `json:"node_type"`
	ErrorMessage string         `json:"error_message"`
	ErrorCode    string         `json:"error_code,omitempty"`
	HTTPStatus   int            `json:"http_status,omitempty"`
	RetryCount   int            `json:"retry_count,omitempty"`
	InputData    map[string]any `json:"input_data,omitempty"`
	NodeConfig   map[string]any `json:"node_config,omitempty"`
}

// List returns all suggestions for an execution
// @Summary List suggestions
// @Description Returns all suggestions for a specific execution
// @Tags Suggestions
// @Accept json
// @Produce json
// @Param executionID path string true "Execution ID"
// @Security TenantID
// @Success 200 {object} map[string]any "List of suggestions"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /executions/{executionID}/suggestions [get]
func (h *SuggestionsHandler) List(w http.ResponseWriter, r *http.Request) {
	tenantID := getSuggestionsTenantID(r)
	executionID := chi.URLParam(r, "executionID")

	suggs, err := h.service.GetSuggestions(r.Context(), tenantID, executionID)
	if err != nil {
		h.logger.Error("failed to list suggestions", "error", err, "execution_id", executionID)
		_ = response.InternalError(w, "failed to list suggestions")
		return
	}

	_ = response.OK(w, map[string]any{
		"data": suggs,
	})
}

// Get returns a single suggestion by ID
// @Summary Get suggestion
// @Description Returns a single suggestion by ID
// @Tags Suggestions
// @Accept json
// @Produce json
// @Param suggestionID path string true "Suggestion ID"
// @Security TenantID
// @Success 200 {object} map[string]any "Suggestion details"
// @Failure 404 {object} map[string]string "Suggestion not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /suggestions/{suggestionID} [get]
func (h *SuggestionsHandler) Get(w http.ResponseWriter, r *http.Request) {
	tenantID := getSuggestionsTenantID(r)
	suggestionID := chi.URLParam(r, "suggestionID")

	sugg, err := h.service.GetSuggestionByID(r.Context(), tenantID, suggestionID)
	if err != nil {
		if errors.Is(err, suggestions.ErrSuggestionNotFound) {
			_ = response.NotFound(w, "suggestion not found")
			return
		}
		h.logger.Error("failed to get suggestion", "error", err, "suggestion_id", suggestionID)
		_ = response.InternalError(w, "failed to get suggestion")
		return
	}

	_ = response.OK(w, map[string]any{
		"data": sugg,
	})
}

// Analyze analyzes an execution error and generates suggestions
// @Summary Analyze error
// @Description Analyzes an execution error and generates fix suggestions
// @Tags Suggestions
// @Accept json
// @Produce json
// @Param executionID path string true "Execution ID"
// @Param request body AnalyzeRequest true "Error context for analysis"
// @Security TenantID
// @Success 200 {object} map[string]any "Generated suggestions"
// @Failure 400 {object} map[string]string "Invalid request"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /executions/{executionID}/analyze [post]
func (h *SuggestionsHandler) Analyze(w http.ResponseWriter, r *http.Request) {
	tenantID := getSuggestionsTenantID(r)
	executionID := chi.URLParam(r, "executionID")

	var req AnalyzeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		_ = response.BadRequest(w, "invalid request body")
		return
	}

	errCtx := &suggestions.ErrorContext{
		ExecutionID:  executionID,
		WorkflowID:   req.WorkflowID,
		NodeID:       req.NodeID,
		NodeType:     req.NodeType,
		ErrorMessage: req.ErrorMessage,
		ErrorCode:    req.ErrorCode,
		HTTPStatus:   req.HTTPStatus,
		RetryCount:   req.RetryCount,
		InputData:    req.InputData,
		NodeConfig:   req.NodeConfig,
	}

	suggs, err := h.service.AnalyzeError(r.Context(), tenantID, errCtx)
	if err != nil {
		h.logger.Error("failed to analyze error", "error", err, "execution_id", executionID)
		_ = response.InternalError(w, "failed to analyze error")
		return
	}

	_ = response.OK(w, map[string]any{
		"data": suggs,
	})
}

// Apply marks a suggestion as applied
// @Summary Apply suggestion
// @Description Marks a suggestion as applied (user accepted the fix)
// @Tags Suggestions
// @Accept json
// @Produce json
// @Param suggestionID path string true "Suggestion ID"
// @Security TenantID
// @Success 200 {object} map[string]any "Success message"
// @Failure 404 {object} map[string]string "Suggestion not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /suggestions/{suggestionID}/apply [post]
func (h *SuggestionsHandler) Apply(w http.ResponseWriter, r *http.Request) {
	tenantID := getSuggestionsTenantID(r)
	suggestionID := chi.URLParam(r, "suggestionID")

	err := h.service.ApplySuggestion(r.Context(), tenantID, suggestionID)
	if err != nil {
		if errors.Is(err, suggestions.ErrSuggestionNotFound) {
			_ = response.NotFound(w, "suggestion not found")
			return
		}
		h.logger.Error("failed to apply suggestion", "error", err, "suggestion_id", suggestionID)
		_ = response.InternalError(w, "failed to apply suggestion")
		return
	}

	_ = response.OK(w, map[string]any{
		"message": "suggestion applied",
	})
}

// Dismiss marks a suggestion as dismissed
// @Summary Dismiss suggestion
// @Description Marks a suggestion as dismissed (user rejected the fix)
// @Tags Suggestions
// @Accept json
// @Produce json
// @Param suggestionID path string true "Suggestion ID"
// @Security TenantID
// @Success 200 {object} map[string]any "Success message"
// @Failure 404 {object} map[string]string "Suggestion not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /suggestions/{suggestionID}/dismiss [post]
func (h *SuggestionsHandler) Dismiss(w http.ResponseWriter, r *http.Request) {
	tenantID := getSuggestionsTenantID(r)
	suggestionID := chi.URLParam(r, "suggestionID")

	err := h.service.DismissSuggestion(r.Context(), tenantID, suggestionID)
	if err != nil {
		if errors.Is(err, suggestions.ErrSuggestionNotFound) {
			_ = response.NotFound(w, "suggestion not found")
			return
		}
		h.logger.Error("failed to dismiss suggestion", "error", err, "suggestion_id", suggestionID)
		_ = response.InternalError(w, "failed to dismiss suggestion")
		return
	}

	_ = response.OK(w, map[string]any{
		"message": "suggestion dismissed",
	})
}

// getSuggestionsTenantID retrieves the tenant ID from the request context
// It first tries the middleware package, then falls back to direct context lookup
func getSuggestionsTenantID(r *http.Request) string {
	// First try the middleware package function
	tenantID := middleware.GetTenantID(r)
	if tenantID != "" {
		return tenantID
	}

	// Fallback to direct context lookup (for tests)
	if val := r.Context().Value(suggestionsTenantIDKey); val != nil {
		if str, ok := val.(string); ok {
			return str
		}
	}

	return ""
}
