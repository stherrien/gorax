package webhook

import (
	"time"
)

// Webhook represents a webhook configuration
type Webhook struct {
	ID         string    `db:"id" json:"id"`
	TenantID   string    `db:"tenant_id" json:"tenant_id"`
	WorkflowID string    `db:"workflow_id" json:"workflow_id"`
	NodeID     string    `db:"node_id" json:"node_id"`
	Path       string    `db:"path" json:"path"`
	Secret     string    `db:"secret" json:"secret"`
	AuthType   string    `db:"auth_type" json:"auth_type"`
	Enabled    bool      `db:"enabled" json:"enabled"`
	CreatedAt  time.Time `db:"created_at" json:"created_at"`
	UpdatedAt  time.Time `db:"updated_at" json:"updated_at"`
}

// WebhookURL returns the full webhook URL path
func (w *Webhook) WebhookURL() string {
	return "/webhooks/" + w.WorkflowID + "/" + w.ID
}

// AuthType constants
const (
	AuthTypeNone      = "none"
	AuthTypeSignature = "signature"
	AuthTypeBasic     = "basic"
	AuthTypeAPIKey    = "api_key"
)
