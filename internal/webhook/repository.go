package webhook

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

var (
	ErrNotFound = errors.New("webhook not found")
)

// Repository handles webhook database operations
type Repository struct {
	db *sqlx.DB
}

// NewRepository creates a new webhook repository
func NewRepository(db *sqlx.DB) *Repository {
	return &Repository{db: db}
}

// Create creates a new webhook
func (r *Repository) Create(ctx context.Context, tenantID, workflowID, nodeID, secret, authType string) (*Webhook, error) {
	id := uuid.New().String()
	now := time.Now()

	// Generate unique path: /webhooks/{workflowID}/{webhookID}
	path := "/webhooks/" + workflowID + "/" + id

	query := `
		INSERT INTO webhooks (id, tenant_id, workflow_id, node_id, path, secret, auth_type, enabled, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING *
	`

	var webhook Webhook
	err := r.db.QueryRowxContext(
		ctx, query,
		id, tenantID, workflowID, nodeID, path, secret, authType, true, now, now,
	).StructScan(&webhook)

	if err != nil {
		return nil, err
	}

	return &webhook, nil
}

// GetByID retrieves a webhook by ID
func (r *Repository) GetByID(ctx context.Context, id string) (*Webhook, error) {
	query := `SELECT * FROM webhooks WHERE id = $1`

	var webhook Webhook
	err := r.db.GetContext(ctx, &webhook, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	return &webhook, nil
}

// GetByWorkflowAndWebhookID retrieves a webhook by workflow ID and webhook ID
func (r *Repository) GetByWorkflowAndWebhookID(ctx context.Context, workflowID, webhookID string) (*Webhook, error) {
	query := `SELECT * FROM webhooks WHERE workflow_id = $1 AND id = $2 AND enabled = true`

	var webhook Webhook
	err := r.db.GetContext(ctx, &webhook, query, workflowID, webhookID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	return &webhook, nil
}

// GetByWorkflowID retrieves all webhooks for a workflow
func (r *Repository) GetByWorkflowID(ctx context.Context, workflowID string) ([]*Webhook, error) {
	query := `SELECT * FROM webhooks WHERE workflow_id = $1`

	var webhooks []*Webhook
	err := r.db.SelectContext(ctx, &webhooks, query, workflowID)
	if err != nil {
		return nil, err
	}

	return webhooks, nil
}

// Delete deletes a webhook
func (r *Repository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM webhooks WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return ErrNotFound
	}

	return nil
}

// DeleteByWorkflowID deletes all webhooks for a workflow
func (r *Repository) DeleteByWorkflowID(ctx context.Context, workflowID string) error {
	query := `DELETE FROM webhooks WHERE workflow_id = $1`

	_, err := r.db.ExecContext(ctx, query, workflowID)
	return err
}

// UpdateEnabled updates the enabled status of a webhook
func (r *Repository) UpdateEnabled(ctx context.Context, id string, enabled bool) error {
	query := `UPDATE webhooks SET enabled = $2, updated_at = $3 WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id, enabled, time.Now())
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return ErrNotFound
	}

	return nil
}
