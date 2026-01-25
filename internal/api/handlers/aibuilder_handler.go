package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"

	"github.com/gorax/gorax/internal/aibuilder"
	"github.com/gorax/gorax/internal/api/middleware"
	"github.com/gorax/gorax/internal/api/response"
)

// Context key types for test fallback
type aibuilderContextKey string

const (
	tenantIDKey aibuilderContextKey = "tenant_id"
	userIDKey   aibuilderContextKey = "user_id"
)

// AIBuilderServiceInterface defines the interface for AI builder service
type AIBuilderServiceInterface interface {
	Generate(ctx context.Context, tenantID, userID string, request *aibuilder.BuildRequest) (*aibuilder.BuildResult, error)
	Refine(ctx context.Context, tenantID string, request *aibuilder.RefineRequest) (*aibuilder.BuildResult, error)
	GetConversation(ctx context.Context, tenantID, conversationID string) (*aibuilder.Conversation, error)
	ListConversations(ctx context.Context, tenantID, userID string) ([]*aibuilder.Conversation, error)
	Apply(ctx context.Context, tenantID, userID string, request *aibuilder.ApplyRequest) (string, error)
	AbandonConversation(ctx context.Context, tenantID, conversationID string) error
}

// AIBuilderHandler handles AI workflow builder HTTP endpoints
type AIBuilderHandler struct {
	service AIBuilderServiceInterface
}

// NewAIBuilderHandler creates a new AI builder handler
func NewAIBuilderHandler(service AIBuilderServiceInterface) *AIBuilderHandler {
	return &AIBuilderHandler{
		service: service,
	}
}

// Generate handles POST /api/v1/ai/workflows/generate
// @Summary Generate a workflow from description
// @Description Generates a new workflow from a natural language description
// @Tags ai-builder
// @Accept json
// @Produce json
// @Param request body aibuilder.BuildRequest true "Build request"
// @Success 200 {object} aibuilder.BuildResult
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/ai/workflows/generate [post]
func (h *AIBuilderHandler) Generate(w http.ResponseWriter, r *http.Request) {
	tenantID := getAIBuilderTenantID(r)
	userID := getAIBuilderUserID(r)

	var request aibuilder.BuildRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		_ = response.BadRequest(w, "invalid request body")
		return
	}

	result, err := h.service.Generate(r.Context(), tenantID, userID, &request)
	if err != nil {
		_ = response.InternalError(w, err.Error())
		return
	}

	_ = response.OK(w, result)
}

// Refine handles POST /api/v1/ai/workflows/refine
// @Summary Refine an existing workflow
// @Description Refines an existing workflow based on user feedback
// @Tags ai-builder
// @Accept json
// @Produce json
// @Param request body aibuilder.RefineRequest true "Refine request"
// @Success 200 {object} aibuilder.BuildResult
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/ai/workflows/refine [post]
func (h *AIBuilderHandler) Refine(w http.ResponseWriter, r *http.Request) {
	tenantID := getAIBuilderTenantID(r)

	var request aibuilder.RefineRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		_ = response.BadRequest(w, "invalid request body")
		return
	}

	result, err := h.service.Refine(r.Context(), tenantID, &request)
	if err != nil {
		_ = response.InternalError(w, err.Error())
		return
	}

	_ = response.OK(w, result)
}

// GetConversation handles GET /api/v1/ai/workflows/conversations/{id}
// @Summary Get a conversation
// @Description Gets a conversation by ID
// @Tags ai-builder
// @Produce json
// @Param id path string true "Conversation ID"
// @Success 200 {object} aibuilder.Conversation
// @Failure 404 {object} ErrorResponse
// @Router /api/v1/ai/workflows/conversations/{id} [get]
func (h *AIBuilderHandler) GetConversation(w http.ResponseWriter, r *http.Request) {
	tenantID := getAIBuilderTenantID(r)
	conversationID := chi.URLParam(r, "id")

	conv, err := h.service.GetConversation(r.Context(), tenantID, conversationID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			_ = response.NotFound(w, "conversation not found")
			return
		}
		_ = response.InternalError(w, err.Error())
		return
	}

	_ = response.OK(w, conv)
}

// ListConversations handles GET /api/v1/ai/workflows/conversations
// @Summary List conversations
// @Description Lists all conversations for the current user
// @Tags ai-builder
// @Produce json
// @Success 200 {object} map[string]any
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/ai/workflows/conversations [get]
func (h *AIBuilderHandler) ListConversations(w http.ResponseWriter, r *http.Request) {
	tenantID := getAIBuilderTenantID(r)
	userID := getAIBuilderUserID(r)

	convs, err := h.service.ListConversations(r.Context(), tenantID, userID)
	if err != nil {
		_ = response.InternalError(w, err.Error())
		return
	}

	_ = response.OK(w, map[string]any{
		"data": convs,
	})
}

// Apply handles POST /api/v1/ai/workflows/conversations/{id}/apply
// @Summary Apply a generated workflow
// @Description Creates an actual workflow from the generated workflow
// @Tags ai-builder
// @Produce json
// @Param id path string true "Conversation ID"
// @Success 200 {object} map[string]string
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/v1/ai/workflows/conversations/{id}/apply [post]
func (h *AIBuilderHandler) Apply(w http.ResponseWriter, r *http.Request) {
	tenantID := getAIBuilderTenantID(r)
	userID := getAIBuilderUserID(r)
	conversationID := chi.URLParam(r, "id")

	request := &aibuilder.ApplyRequest{
		ConversationID: conversationID,
	}

	// Optional: parse request body for workflow name override
	var bodyRequest struct {
		WorkflowName string `json:"workflow_name,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&bodyRequest); err == nil && bodyRequest.WorkflowName != "" {
		request.WorkflowName = bodyRequest.WorkflowName
	}

	workflowID, err := h.service.Apply(r.Context(), tenantID, userID, request)
	if err != nil {
		if strings.Contains(err.Error(), "no workflow") {
			_ = response.BadRequest(w, err.Error())
			return
		}
		if strings.Contains(err.Error(), "not found") {
			_ = response.NotFound(w, err.Error())
			return
		}
		_ = response.InternalError(w, err.Error())
		return
	}

	_ = response.OK(w, map[string]string{
		"workflow_id": workflowID,
	})
}

// Abandon handles POST /api/v1/ai/workflows/conversations/{id}/abandon
// @Summary Abandon a conversation
// @Description Marks a conversation as abandoned
// @Tags ai-builder
// @Param id path string true "Conversation ID"
// @Success 204 "No Content"
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/v1/ai/workflows/conversations/{id}/abandon [post]
func (h *AIBuilderHandler) Abandon(w http.ResponseWriter, r *http.Request) {
	tenantID := getAIBuilderTenantID(r)
	conversationID := chi.URLParam(r, "id")

	err := h.service.AbandonConversation(r.Context(), tenantID, conversationID)
	if err != nil {
		if strings.Contains(err.Error(), "not active") {
			_ = response.BadRequest(w, err.Error())
			return
		}
		if strings.Contains(err.Error(), "not found") {
			_ = response.NotFound(w, err.Error())
			return
		}
		_ = response.InternalError(w, err.Error())
		return
	}

	response.NoContent(w)
}

// getAIBuilderTenantID extracts tenant ID from request context
func getAIBuilderTenantID(r *http.Request) string {
	// First try middleware package
	tenantID := middleware.GetTenantID(r)
	if tenantID != "" {
		return tenantID
	}

	// Fallback for tests
	if val := r.Context().Value(tenantIDKey); val != nil {
		if str, ok := val.(string); ok {
			return str
		}
	}

	return ""
}

// getAIBuilderUserID extracts user ID from request context
func getAIBuilderUserID(r *http.Request) string {
	// First try middleware package
	userID := middleware.GetUserID(r)
	if userID != "" {
		return userID
	}

	// Fallback for tests
	if val := r.Context().Value(userIDKey); val != nil {
		if str, ok := val.(string); ok {
			return str
		}
	}

	return ""
}
