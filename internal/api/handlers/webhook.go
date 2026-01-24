package handlers

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/gorax/gorax/internal/api/response"
	"github.com/gorax/gorax/internal/webhook"
	"github.com/gorax/gorax/internal/workflow"
)

// WebhookHandler handles incoming webhook requests
type WebhookHandler struct {
	workflowService *workflow.Service
	webhookService  *webhook.Service
	logger          *slog.Logger
}

// NewWebhookHandler creates a new webhook handler
func NewWebhookHandler(workflowService *workflow.Service, webhookService *webhook.Service, logger *slog.Logger) *WebhookHandler {
	return &WebhookHandler{
		workflowService: workflowService,
		webhookService:  webhookService,
		logger:          logger,
	}
}

// Handle processes incoming webhook requests
func (h *WebhookHandler) Handle(w http.ResponseWriter, r *http.Request) {
	workflowID := chi.URLParam(r, "workflowID")
	webhookID := chi.URLParam(r, "webhookID")

	h.logger.Info("webhook received", "workflow_id", workflowID, "webhook_id", webhookID)

	// Capture metadata from request
	metadata := webhook.ExtractMetadataFromRequest(
		r.RemoteAddr,
		r.Header.Get("User-Agent"),
		r.Header.Get("Content-Type"),
		r.ContentLength,
	)

	// Look up webhook configuration
	webhookConfig, err := h.webhookService.GetByWorkflowAndWebhookID(r.Context(), workflowID, webhookID)
	if err != nil {
		if err == webhook.ErrNotFound {
			_ = response.NotFound(w, "webhook not found")
			return
		}
		h.logger.Error("failed to get webhook config", "error", err, "webhook_id", webhookID)
		_ = response.InternalError(w, "failed to process webhook")
		return
	}

	// Read request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		_ = response.BadRequest(w, "failed to read request body")
		return
	}

	// Verify signature if required
	if webhookConfig.AuthType == webhook.AuthTypeSignature {
		signature := r.Header.Get("X-Webhook-Signature")
		if signature == "" {
			// Also check for X-Hub-Signature-256 (GitHub style)
			signature = r.Header.Get("X-Hub-Signature-256")
		}

		if !h.webhookService.VerifySignature(body, signature, webhookConfig.Secret) {
			h.logger.Warn("webhook signature verification failed", "webhook_id", webhookID)
			_ = response.Unauthorized(w, "invalid signature")
			return
		}
	}

	// Build trigger data
	triggerData := map[string]interface{}{
		"method":  r.Method,
		"headers": flattenHeaders(r.Header),
		"query":   flattenQuery(r.URL.Query()),
		"body":    json.RawMessage(body),
	}

	triggerDataJSON, err := json.Marshal(triggerData)
	if err != nil {
		_ = response.InternalError(w, "failed to process trigger data")
		return
	}

	// Execute workflow using tenant ID from webhook config
	execution, err := h.workflowService.Execute(r.Context(), webhookConfig.TenantID, workflowID, "webhook", triggerDataJSON)
	if err != nil {
		if err == workflow.ErrNotFound {
			_ = response.NotFound(w, "workflow not found")
			return
		}
		h.logger.Error("failed to execute workflow from webhook", "error", err, "workflow_id", workflowID)

		// Log failed event with metadata
		h.logWebhookEvent(r.Context(), webhookConfig, r, body, nil, webhook.EventStatusFailed, metadata, stringPtr(err.Error()))

		_ = response.InternalError(w, "failed to execute workflow")
		return
	}

	h.logger.Info("workflow execution triggered", "execution_id", execution.ID, "workflow_id", workflowID)

	// Log successful event with metadata
	h.logWebhookEvent(r.Context(), webhookConfig, r, body, &execution.ID, webhook.EventStatusProcessed, metadata, nil)

	// Return execution ID
	_ = response.JSON(w, http.StatusAccepted, map[string]any{
		"execution_id": execution.ID,
		"status":       execution.Status,
	})
}

func flattenHeaders(headers http.Header) map[string]string {
	result := make(map[string]string)
	for key, values := range headers {
		if len(values) > 0 {
			result[key] = values[0]
		}
	}
	return result
}

func flattenQuery(query map[string][]string) map[string]string {
	result := make(map[string]string)
	for key, values := range query {
		if len(values) > 0 {
			result[key] = values[0]
		}
	}
	return result
}

// logWebhookEvent logs a webhook event with metadata
func (h *WebhookHandler) logWebhookEvent(
	ctx context.Context,
	webhookConfig *webhook.Webhook,
	r *http.Request,
	body []byte,
	executionID *string,
	status webhook.WebhookEventStatus,
	metadata *webhook.EventMetadata,
	errorMsg *string,
) {
	event := &webhook.WebhookEvent{
		TenantID:       webhookConfig.TenantID,
		WebhookID:      webhookConfig.ID,
		ExecutionID:    executionID,
		RequestMethod:  r.Method,
		RequestHeaders: flattenHeaders(r.Header),
		RequestBody:    json.RawMessage(body),
		Status:         status,
		ErrorMessage:   errorMsg,
		Metadata:       metadata,
	}

	// Log the event (best effort - don't fail request if logging fails)
	if err := h.webhookService.LogEvent(ctx, event); err != nil {
		h.logger.Error("failed to log webhook event", "error", err, "webhook_id", webhookConfig.ID)
	}
}

// stringPtr returns a pointer to the given string
func stringPtr(s string) *string {
	return &s
}
