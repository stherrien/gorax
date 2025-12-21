package webhook

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/google/uuid"

	"github.com/gorax/gorax/internal/workflow"
)

// Service handles webhook business logic
type Service struct {
	repo   *Repository
	logger *slog.Logger
}

// NewService creates a new webhook service
func NewService(repo *Repository, logger *slog.Logger) *Service {
	return &Service{
		repo:   repo,
		logger: logger,
	}
}

// GenerateSecret generates a secure random secret for webhook signing
func (s *Service) GenerateSecret() (string, error) {
	bytes := make([]byte, 32) // 256 bits
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate random secret: %w", err)
	}
	return base64.StdEncoding.EncodeToString(bytes), nil
}

// Create creates a new webhook (original signature for backward compatibility)
func (s *Service) Create(ctx context.Context, tenantID, workflowID, nodeID, authType string) (*Webhook, error) {
	// Generate secret if using signature auth
	secret := ""
	if authType == AuthTypeSignature {
		var err error
		secret, err = s.GenerateSecret()
		if err != nil {
			return nil, err
		}
	}

	webhook, err := s.repo.Create(ctx, tenantID, workflowID, nodeID, secret, authType)
	if err != nil {
		s.logger.Error("failed to create webhook", "error", err, "workflow_id", workflowID)
		return nil, err
	}

	s.logger.Info("webhook created", "webhook_id", webhook.ID, "workflow_id", workflowID)
	return webhook, nil
}

// CreateWithDetails creates a new webhook with full details
func (s *Service) CreateWithDetails(ctx context.Context, tenantID, workflowID, name, path, authType, description string, priority int) (*Webhook, error) {
	// For now, use nodeID as name until schema is updated
	// Generate secret if using signature auth
	secret := ""
	if authType == AuthTypeSignature {
		var err error
		secret, err = s.GenerateSecret()
		if err != nil {
			return nil, err
		}
	}

	// Use name as nodeID temporarily
	webhook, err := s.repo.Create(ctx, tenantID, workflowID, name, secret, authType)
	if err != nil {
		s.logger.Error("failed to create webhook", "error", err, "workflow_id", workflowID)
		return nil, err
	}

	// Set fields that don't exist in DB yet
	webhook.Name = name
	webhook.Description = description
	webhook.Priority = priority

	s.logger.Info("webhook created", "webhook_id", webhook.ID, "workflow_id", workflowID)
	return webhook, nil
}

// GetByWorkflowAndWebhookID retrieves a webhook by workflow and webhook IDs
func (s *Service) GetByWorkflowAndWebhookID(ctx context.Context, workflowID, webhookID string) (*Webhook, error) {
	return s.repo.GetByWorkflowAndWebhookID(ctx, workflowID, webhookID)
}

// GetByWorkflowID retrieves all webhooks for a workflow as workflow.WebhookInfo
func (s *Service) GetByWorkflowID(ctx context.Context, workflowID string) ([]*workflow.WebhookInfo, error) {
	webhooks, err := s.repo.GetByWorkflowID(ctx, workflowID)
	if err != nil {
		return nil, err
	}

	// Convert to workflow.WebhookInfo
	result := make([]*workflow.WebhookInfo, len(webhooks))
	for i, wh := range webhooks {
		result[i] = &workflow.WebhookInfo{
			ID:         wh.ID,
			NodeID:     wh.NodeID,
			WebhookURL: wh.WebhookURL(),
			AuthType:   wh.AuthType,
			Secret:     wh.Secret,
		}
	}

	return result, nil
}

// Delete deletes a webhook
func (s *Service) Delete(ctx context.Context, id string) error {
	err := s.repo.Delete(ctx, id)
	if err != nil {
		s.logger.Error("failed to delete webhook", "error", err, "webhook_id", id)
		return err
	}

	s.logger.Info("webhook deleted", "webhook_id", id)
	return nil
}

// DeleteByWorkflowID deletes all webhooks for a workflow
func (s *Service) DeleteByWorkflowID(ctx context.Context, workflowID string) error {
	err := s.repo.DeleteByWorkflowID(ctx, workflowID)
	if err != nil {
		s.logger.Error("failed to delete webhooks", "error", err, "workflow_id", workflowID)
		return err
	}

	s.logger.Info("webhooks deleted for workflow", "workflow_id", workflowID)
	return nil
}

// VerifySignature verifies the HMAC signature of a webhook request
// Signature format: sha256=<hex encoded hmac>
func (s *Service) VerifySignature(payload []byte, signature string, secret string) bool {
	if signature == "" || secret == "" {
		return false
	}

	// Remove 'sha256=' prefix if present
	if len(signature) > 7 && signature[:7] == "sha256=" {
		signature = signature[7:]
	}

	// Calculate expected signature
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	expectedMAC := mac.Sum(nil)
	expectedSignature := hex.EncodeToString(expectedMAC)

	// Compare signatures (constant time comparison)
	return hmac.Equal([]byte(signature), []byte(expectedSignature))
}

// GenerateSignature generates an HMAC signature for testing purposes
func (s *Service) GenerateSignature(payload []byte, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	signature := hex.EncodeToString(mac.Sum(nil))
	return "sha256=" + signature
}

// SyncWorkflowWebhooks syncs webhooks for a workflow based on its definition
// This should be called when a workflow is created or updated
func (s *Service) SyncWorkflowWebhooks(ctx context.Context, tenantID, workflowID string, webhookNodes []workflow.WebhookNodeConfig) error {
	// Get existing webhooks
	existing, err := s.repo.GetByWorkflowID(ctx, workflowID)
	if err != nil {
		return err
	}

	// Create a map of existing webhooks by node_id
	existingMap := make(map[string]*Webhook)
	for _, wh := range existing {
		existingMap[wh.NodeID] = wh
	}

	// Track which webhooks should exist
	shouldExist := make(map[string]bool)

	// Create or update webhooks for each webhook node
	for _, nodeConfig := range webhookNodes {
		shouldExist[nodeConfig.NodeID] = true

		// If webhook doesn't exist, create it
		if _, exists := existingMap[nodeConfig.NodeID]; !exists {
			authType := nodeConfig.AuthType
			if authType == "" {
				authType = AuthTypeSignature // Default to signature
			}

			_, err := s.Create(ctx, tenantID, workflowID, nodeConfig.NodeID, authType)
			if err != nil {
				s.logger.Error("failed to create webhook during sync", "error", err, "node_id", nodeConfig.NodeID)
				return err
			}
		}
	}

	// Delete webhooks that no longer exist in the workflow definition
	for nodeID, webhook := range existingMap {
		if !shouldExist[nodeID] {
			if err := s.repo.Delete(ctx, webhook.ID); err != nil {
				s.logger.Error("failed to delete orphaned webhook", "error", err, "webhook_id", webhook.ID)
				// Continue even if deletion fails
			}
		}
	}

	return nil
}

// List retrieves all webhooks for a tenant with pagination
func (s *Service) List(ctx context.Context, tenantID string, limit, offset int) ([]*Webhook, int, error) {
	webhooks, total, err := s.repo.List(ctx, tenantID, limit, offset)
	if err != nil {
		s.logger.Error("failed to list webhooks", "error", err, "tenant_id", tenantID)
		return nil, 0, err
	}
	return webhooks, total, nil
}

// GetByID retrieves a webhook by ID with tenant isolation
func (s *Service) GetByID(ctx context.Context, tenantID, webhookID string) (*Webhook, error) {
	webhook, err := s.repo.GetByIDAndTenant(ctx, webhookID, tenantID)
	if err != nil {
		s.logger.Error("failed to get webhook", "error", err, "webhook_id", webhookID)
		return nil, err
	}
	return webhook, nil
}

// Update updates a webhook
func (s *Service) Update(ctx context.Context, tenantID, webhookID, name, authType, description string, priority int, enabled bool) (*Webhook, error) {
	// Verify webhook belongs to tenant
	_, err := s.repo.GetByIDAndTenant(ctx, webhookID, tenantID)
	if err != nil {
		return nil, err
	}

	webhook, err := s.repo.Update(ctx, webhookID, name, authType, description, priority, enabled)
	if err != nil {
		s.logger.Error("failed to update webhook", "error", err, "webhook_id", webhookID)
		return nil, err
	}

	s.logger.Info("webhook updated", "webhook_id", webhookID)
	return webhook, nil
}

// DeleteByID deletes a webhook with tenant isolation
func (s *Service) DeleteByID(ctx context.Context, tenantID, webhookID string) error {
	// Verify webhook belongs to tenant
	_, err := s.repo.GetByIDAndTenant(ctx, webhookID, tenantID)
	if err != nil {
		return err
	}

	err = s.repo.Delete(ctx, webhookID)
	if err != nil {
		s.logger.Error("failed to delete webhook", "error", err, "webhook_id", webhookID)
		return err
	}

	s.logger.Info("webhook deleted", "webhook_id", webhookID)
	return nil
}

// RegenerateSecret regenerates the secret for a webhook
func (s *Service) RegenerateSecret(ctx context.Context, tenantID, webhookID string) (*Webhook, error) {
	// Verify webhook belongs to tenant
	_, err := s.repo.GetByIDAndTenant(ctx, webhookID, tenantID)
	if err != nil {
		return nil, err
	}

	// Generate new secret
	secret, err := s.GenerateSecret()
	if err != nil {
		return nil, err
	}

	webhook, err := s.repo.UpdateSecret(ctx, webhookID, secret)
	if err != nil {
		s.logger.Error("failed to regenerate secret", "error", err, "webhook_id", webhookID)
		return nil, err
	}

	s.logger.Info("webhook secret regenerated", "webhook_id", webhookID)
	return webhook, nil
}

// TestWebhook tests a webhook with sample payload
func (s *Service) TestWebhook(ctx context.Context, tenantID, webhookID, method string, headers map[string]string, body json.RawMessage) (*TestResult, error) {
	// Verify webhook belongs to tenant
	webhook, err := s.repo.GetByIDAndTenant(ctx, webhookID, tenantID)
	if err != nil {
		return nil, err
	}

	// This is a simulation - in real implementation would trigger actual workflow
	result := &TestResult{
		Success:      true,
		StatusCode:   200,
		ResponseTime: 150,
		ExecutionID:  "test-execution-id",
	}

	s.logger.Info("webhook tested", "webhook_id", webhookID, "workflow_id", webhook.WorkflowID)
	return result, nil
}

// GetEventHistory retrieves webhook event history
func (s *Service) GetEventHistory(ctx context.Context, tenantID, webhookID string, limit, offset int) ([]*Event, int, error) {
	// Verify webhook belongs to tenant
	_, err := s.repo.GetByIDAndTenant(ctx, webhookID, tenantID)
	if err != nil {
		return nil, 0, err
	}

	// Note: This returns empty for now - would need to query webhook_events table
	// For MVP, returning empty slice
	events := []*Event{}
	total := 0

	return events, total, nil
}

// LogEvent logs a webhook event to the database
func (s *Service) LogEvent(ctx context.Context, event *WebhookEvent) error {
	err := s.repo.CreateEvent(ctx, event)
	if err != nil {
		s.logger.Error("failed to log webhook event",
			"error", err,
			"webhook_id", event.WebhookID,
			"status", event.Status)
		return fmt.Errorf("failed to log webhook event: %w", err)
	}

	s.logger.Debug("webhook event logged",
		"event_id", event.ID,
		"webhook_id", event.WebhookID,
		"status", event.Status)

	return nil
}

// GetEvents retrieves webhook events with pagination
func (s *Service) GetEvents(ctx context.Context, tenantID, webhookID string, limit, offset int) ([]*WebhookEvent, int, error) {
	filter := WebhookEventFilter{
		WebhookID: webhookID,
		Limit:     limit,
		Offset:    offset,
	}

	events, total, err := s.repo.ListEvents(ctx, tenantID, filter)
	if err != nil {
		s.logger.Error("failed to get webhook events",
			"error", err,
			"webhook_id", webhookID)
		return nil, 0, fmt.Errorf("failed to get webhook events: %w", err)
	}

	return events, total, nil
}

// MarkEventProcessed marks a webhook event as processed
func (s *Service) MarkEventProcessed(ctx context.Context, eventID, executionID string, processingTimeMs int) error {
	// First update the status
	err := s.repo.UpdateEventStatus(ctx, eventID, EventStatusProcessed, nil)
	if err != nil {
		s.logger.Error("failed to mark event as processed",
			"error", err,
			"event_id", eventID)
		return fmt.Errorf("failed to mark event as processed: %w", err)
	}

	// Then update execution ID and processing time
	query := `UPDATE webhook_events SET execution_id = $2, processing_time_ms = $3 WHERE id = $1`
	_, err = s.repo.db.ExecContext(ctx, query, eventID, executionID, processingTimeMs)
	if err != nil {
		s.logger.Error("failed to update event execution details",
			"error", err,
			"event_id", eventID)
		return fmt.Errorf("failed to update event execution details: %w", err)
	}

	s.logger.Info("webhook event marked as processed",
		"event_id", eventID,
		"execution_id", executionID,
		"processing_time_ms", processingTimeMs)

	return nil
}

// MarkEventFailed marks a webhook event as failed
func (s *Service) MarkEventFailed(ctx context.Context, eventID string, errorMsg string) error {
	err := s.repo.UpdateEventStatus(ctx, eventID, EventStatusFailed, &errorMsg)
	if err != nil {
		s.logger.Error("failed to mark event as failed",
			"error", err,
			"event_id", eventID)
		return fmt.Errorf("failed to mark event as failed: %w", err)
	}

	s.logger.Info("webhook event marked as failed",
		"event_id", eventID,
		"error", errorMsg)

	return nil
}

// MarkEventFiltered marks a webhook event as filtered
func (s *Service) MarkEventFiltered(ctx context.Context, eventID string, reason string) error {
	// Update status to filtered
	err := s.repo.UpdateEventStatus(ctx, eventID, EventStatusFiltered, nil)
	if err != nil {
		s.logger.Error("failed to mark event as filtered",
			"error", err,
			"event_id", eventID)
		return fmt.Errorf("failed to mark event as filtered: %w", err)
	}

	// Update filtered reason
	query := `UPDATE webhook_events SET filtered_reason = $2 WHERE id = $1`
	_, err = s.repo.db.ExecContext(ctx, query, eventID, reason)
	if err != nil {
		s.logger.Error("failed to update filtered reason",
			"error", err,
			"event_id", eventID)
		return fmt.Errorf("failed to update filtered reason: %w", err)
	}

	s.logger.Info("webhook event marked as filtered",
		"event_id", eventID,
		"reason", reason)

	return nil
}

// ListFilters retrieves all filters for a webhook
func (s *Service) ListFilters(ctx context.Context, tenantID, webhookID string) ([]*WebhookFilter, error) {
	// Verify webhook exists and belongs to tenant
	wh, err := s.repo.GetByID(ctx, webhookID)
	if err != nil {
		return nil, ErrNotFound
	}
	if wh.TenantID != tenantID {
		return nil, ErrNotFound
	}

	filters, err := s.repo.GetFiltersByWebhookID(ctx, webhookID)
	if err != nil {
		s.logger.Error("failed to list filters", "error", err, "webhook_id", webhookID)
		return nil, fmt.Errorf("failed to list filters: %w", err)
	}

	return filters, nil
}

// GetFilter retrieves a single filter by ID
func (s *Service) GetFilter(ctx context.Context, tenantID, webhookID, filterID string) (*WebhookFilter, error) {
	// Verify webhook exists and belongs to tenant
	wh, err := s.repo.GetByID(ctx, webhookID)
	if err != nil {
		return nil, ErrNotFound
	}
	if wh.TenantID != tenantID {
		return nil, ErrNotFound
	}

	filter, err := s.repo.GetFilterByID(ctx, filterID)
	if err != nil {
		return nil, ErrNotFound
	}

	// Verify filter belongs to webhook
	if filter.WebhookID != webhookID {
		return nil, ErrNotFound
	}

	return filter, nil
}

// CreateFilter creates a new webhook filter
func (s *Service) CreateFilter(ctx context.Context, tenantID, webhookID string, filter *WebhookFilter) (*WebhookFilter, error) {
	// Verify webhook exists and belongs to tenant
	wh, err := s.repo.GetByID(ctx, webhookID)
	if err != nil {
		return nil, ErrNotFound
	}
	if wh.TenantID != tenantID {
		return nil, ErrNotFound
	}

	// Set webhook ID and generate filter ID
	filter.WebhookID = webhookID
	filter.ID = uuid.NewString()

	created, err := s.repo.CreateFilter(ctx, filter)
	if err != nil {
		s.logger.Error("failed to create filter", "error", err, "webhook_id", webhookID)
		return nil, fmt.Errorf("failed to create filter: %w", err)
	}

	s.logger.Info("filter created",
		"filter_id", created.ID,
		"webhook_id", webhookID,
		"field_path", filter.FieldPath,
		"operator", filter.Operator)

	return created, nil
}

// UpdateFilter updates an existing webhook filter
func (s *Service) UpdateFilter(ctx context.Context, tenantID, webhookID, filterID string, filter *WebhookFilter) (*WebhookFilter, error) {
	// Verify webhook exists and belongs to tenant
	wh, err := s.repo.GetByID(ctx, webhookID)
	if err != nil {
		return nil, ErrNotFound
	}
	if wh.TenantID != tenantID {
		return nil, ErrNotFound
	}

	// Verify filter exists and belongs to webhook
	existing, err := s.repo.GetFilterByID(ctx, filterID)
	if err != nil {
		return nil, ErrNotFound
	}
	if existing.WebhookID != webhookID {
		return nil, ErrNotFound
	}

	// Preserve ID and webhook ID
	filter.ID = filterID
	filter.WebhookID = webhookID

	updated, err := s.repo.UpdateFilter(ctx, filter)
	if err != nil {
		s.logger.Error("failed to update filter", "error", err, "filter_id", filterID)
		return nil, fmt.Errorf("failed to update filter: %w", err)
	}

	s.logger.Info("filter updated",
		"filter_id", filterID,
		"webhook_id", webhookID)

	return updated, nil
}

// DeleteFilter deletes a webhook filter
func (s *Service) DeleteFilter(ctx context.Context, tenantID, webhookID, filterID string) error {
	// Verify webhook exists and belongs to tenant
	wh, err := s.repo.GetByID(ctx, webhookID)
	if err != nil {
		return ErrNotFound
	}
	if wh.TenantID != tenantID {
		return ErrNotFound
	}

	// Verify filter exists and belongs to webhook
	existing, err := s.repo.GetFilterByID(ctx, filterID)
	if err != nil {
		return ErrNotFound
	}
	if existing.WebhookID != webhookID {
		return ErrNotFound
	}

	err = s.repo.DeleteFilter(ctx, filterID)
	if err != nil {
		s.logger.Error("failed to delete filter", "error", err, "filter_id", filterID)
		return fmt.Errorf("failed to delete filter: %w", err)
	}

	s.logger.Info("filter deleted",
		"filter_id", filterID,
		"webhook_id", webhookID)

	return nil
}

// TestFilters tests a set of filters against a sample payload
func (s *Service) TestFilters(ctx context.Context, tenantID, webhookID string, payload map[string]interface{}) (*FilterResult, error) {
	// Verify webhook exists and belongs to tenant
	wh, err := s.repo.GetByID(ctx, webhookID)
	if err != nil {
		return nil, ErrNotFound
	}
	if wh.TenantID != tenantID {
		return nil, ErrNotFound
	}

	// Create filter evaluator and evaluate
	evaluator := NewFilterEvaluator(s.repo)
	result, err := evaluator.Evaluate(ctx, webhookID, payload)
	if err != nil {
		s.logger.Error("failed to test filters", "error", err, "webhook_id", webhookID)
		return nil, fmt.Errorf("failed to test filters: %w", err)
	}

	return result, nil
}
