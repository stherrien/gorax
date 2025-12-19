package handlers

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"

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

	// Look up webhook configuration
	webhookConfig, err := h.webhookService.GetByWorkflowAndWebhookID(r.Context(), workflowID, webhookID)
	if err != nil {
		if err == webhook.ErrNotFound {
			h.respondError(w, http.StatusNotFound, "webhook not found")
			return
		}
		h.logger.Error("failed to get webhook config", "error", err, "webhook_id", webhookID)
		h.respondError(w, http.StatusInternalServerError, "failed to process webhook")
		return
	}

	// Read request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "failed to read request body")
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
			h.respondError(w, http.StatusUnauthorized, "invalid signature")
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
		h.respondError(w, http.StatusInternalServerError, "failed to process trigger data")
		return
	}

	// Execute workflow using tenant ID from webhook config
	execution, err := h.workflowService.Execute(r.Context(), webhookConfig.TenantID, workflowID, "webhook", triggerDataJSON)
	if err != nil {
		if err == workflow.ErrNotFound {
			h.respondError(w, http.StatusNotFound, "workflow not found")
			return
		}
		h.logger.Error("failed to execute workflow from webhook", "error", err, "workflow_id", workflowID)
		h.respondError(w, http.StatusInternalServerError, "failed to execute workflow")
		return
	}

	h.logger.Info("workflow execution triggered", "execution_id", execution.ID, "workflow_id", workflowID)

	// Return execution ID
	h.respondJSON(w, http.StatusAccepted, map[string]interface{}{
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

func (h *WebhookHandler) respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func (h *WebhookHandler) respondError(w http.ResponseWriter, status int, message string) {
	h.respondJSON(w, status, map[string]string{
		"error": message,
	})
}
