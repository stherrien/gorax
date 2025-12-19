package webhook

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"log/slog"

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

// Create creates a new webhook
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
