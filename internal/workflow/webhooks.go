package workflow

import "context"

// WebhookService interface to avoid circular dependency
type WebhookService interface {
	SyncWorkflowWebhooks(ctx context.Context, tenantID, workflowID string, webhookNodes []WebhookNodeConfig) error
	DeleteByWorkflowID(ctx context.Context, workflowID string) error
	GetByWorkflowID(ctx context.Context, workflowID string) ([]*WebhookInfo, error)
}

// WebhookNodeConfig represents a webhook trigger node configuration
type WebhookNodeConfig struct {
	NodeID   string
	AuthType string
}

// WebhookInfo contains webhook information for responses
type WebhookInfo struct {
	ID         string
	NodeID     string
	WebhookURL string
	AuthType   string
	Secret     string
}
