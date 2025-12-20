package webhook

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
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

// CreateEvent creates a new webhook event
func (r *Repository) CreateEvent(ctx context.Context, event *WebhookEvent) error {
	event.ID = uuid.New().String()
	event.CreatedAt = time.Now()

	query := `
		INSERT INTO webhook_events (
			id, tenant_id, webhook_id, execution_id, request_method,
			request_headers, request_body, response_status, processing_time_ms,
			status, error_message, filtered_reason, created_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	`

	_, err := r.db.ExecContext(
		ctx, query,
		event.ID,
		event.TenantID,
		event.WebhookID,
		event.ExecutionID,
		event.RequestMethod,
		event.RequestHeaders,
		event.RequestBody,
		event.ResponseStatus,
		event.ProcessingTimeMs,
		event.Status,
		event.ErrorMessage,
		event.FilteredReason,
		event.CreatedAt,
	)

	if err != nil {
		return err
	}

	return nil
}

// GetEventByID retrieves a webhook event by ID with tenant isolation
func (r *Repository) GetEventByID(ctx context.Context, tenantID, eventID string) (*WebhookEvent, error) {
	query := `SELECT * FROM webhook_events WHERE id = $1 AND tenant_id = $2`

	var event WebhookEvent
	err := r.db.GetContext(ctx, &event, query, eventID, tenantID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	return &event, nil
}

// ListEvents retrieves webhook events with filtering
func (r *Repository) ListEvents(ctx context.Context, tenantID string, filter WebhookEventFilter) ([]*WebhookEvent, int, error) {
	// Build the WHERE clause dynamically
	whereConditions := []string{"tenant_id = $1", "webhook_id = $2"}
	args := []interface{}{tenantID, filter.WebhookID}
	argCount := 2

	if filter.Status != nil {
		argCount++
		whereConditions = append(whereConditions, fmt.Sprintf("status = $%d", argCount))
		args = append(args, *filter.Status)
	}

	if filter.StartDate != nil {
		argCount++
		whereConditions = append(whereConditions, fmt.Sprintf("created_at >= $%d", argCount))
		args = append(args, *filter.StartDate)
	}

	if filter.EndDate != nil {
		argCount++
		whereConditions = append(whereConditions, fmt.Sprintf("created_at <= $%d", argCount))
		args = append(args, *filter.EndDate)
	}

	whereClause := "WHERE " + whereConditions[0]
	for i := 1; i < len(whereConditions); i++ {
		whereClause += " AND " + whereConditions[i]
	}

	// Get total count
	countQuery := "SELECT COUNT(*) FROM webhook_events " + whereClause
	var total int
	err := r.db.GetContext(ctx, &total, countQuery, args...)
	if err != nil {
		return nil, 0, err
	}

	// Set default limit and offset
	if filter.Limit == 0 {
		filter.Limit = 50
	}

	// Get events with pagination
	query := fmt.Sprintf(`
		SELECT * FROM webhook_events
		%s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d`,
		whereClause, argCount+1, argCount+2)

	args = append(args, filter.Limit, filter.Offset)

	var events []*WebhookEvent
	err = r.db.SelectContext(ctx, &events, query, args...)
	if err != nil {
		return nil, 0, err
	}

	return events, total, nil
}

// UpdateEventStatus updates the status of a webhook event
func (r *Repository) UpdateEventStatus(ctx context.Context, eventID string, status WebhookEventStatus, errorMsg *string) error {
	query := `UPDATE webhook_events SET status = $2, error_message = $3 WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, eventID, status, errorMsg)
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

// IncrementTriggerCount increments the trigger count for a webhook
func (r *Repository) IncrementTriggerCount(ctx context.Context, webhookID string) error {
	query := `UPDATE webhooks SET trigger_count = trigger_count + 1 WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, webhookID)
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

// UpdateLastTriggeredAt updates the last triggered timestamp for a webhook
func (r *Repository) UpdateLastTriggeredAt(ctx context.Context, webhookID string) error {
	query := `UPDATE webhooks SET last_triggered_at = $2 WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, webhookID, time.Now())
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

// List retrieves all webhooks for a tenant with pagination
func (r *Repository) List(ctx context.Context, tenantID string, limit, offset int) ([]*Webhook, int, error) {
	if limit == 0 {
		limit = 20
	}

	countQuery := `SELECT COUNT(*) FROM webhooks WHERE tenant_id = $1`
	var total int
	err := r.db.GetContext(ctx, &total, countQuery, tenantID)
	if err != nil {
		return nil, 0, err
	}

	query := `
		SELECT * FROM webhooks
		WHERE tenant_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	var webhooks []*Webhook
	err = r.db.SelectContext(ctx, &webhooks, query, tenantID, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	return webhooks, total, nil
}

// Update updates webhook fields
func (r *Repository) Update(ctx context.Context, id, name, authType, description string, priority int, enabled bool) (*Webhook, error) {
	query := `
		UPDATE webhooks
		SET name = $2, auth_type = $3, description = $4, priority = $5, enabled = $6, updated_at = $7
		WHERE id = $1
		RETURNING *
	`

	var webhook Webhook
	err := r.db.QueryRowxContext(
		ctx, query,
		id, name, authType, description, priority, enabled, time.Now(),
	).StructScan(&webhook)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	return &webhook, nil
}

// UpdateSecret updates the webhook secret
func (r *Repository) UpdateSecret(ctx context.Context, id, secret string) (*Webhook, error) {
	query := `
		UPDATE webhooks
		SET secret = $2, updated_at = $3
		WHERE id = $1
		RETURNING *
	`

	var webhook Webhook
	err := r.db.QueryRowxContext(
		ctx, query,
		id, secret, time.Now(),
	).StructScan(&webhook)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	return &webhook, nil
}

// GetByIDAndTenant retrieves a webhook by ID and tenant ID
func (r *Repository) GetByIDAndTenant(ctx context.Context, id, tenantID string) (*Webhook, error) {
	query := `SELECT * FROM webhooks WHERE id = $1 AND tenant_id = $2`

	var webhook Webhook
	err := r.db.GetContext(ctx, &webhook, query, id, tenantID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	return &webhook, nil
}

// GetFiltersByWebhookID retrieves all filters for a webhook
func (r *Repository) GetFiltersByWebhookID(ctx context.Context, webhookID string) ([]*WebhookFilter, error) {
	query := `SELECT * FROM webhook_filters WHERE webhook_id = $1 ORDER BY logic_group, created_at`

	var filters []*WebhookFilter
	err := r.db.SelectContext(ctx, &filters, query, webhookID)
	if err != nil {
		return nil, err
	}

	return filters, nil
}

// GetFilterByID retrieves a single filter by ID
func (r *Repository) GetFilterByID(ctx context.Context, filterID string) (*WebhookFilter, error) {
	query := `SELECT * FROM webhook_filters WHERE id = $1`

	var filter WebhookFilter
	err := r.db.GetContext(ctx, &filter, query, filterID)
	if err != nil {
		return nil, err
	}

	return &filter, nil
}

// CreateFilter creates a new webhook filter
func (r *Repository) CreateFilter(ctx context.Context, filter *WebhookFilter) (*WebhookFilter, error) {
	query := `
		INSERT INTO webhook_filters (id, webhook_id, field_path, operator, value, logic_group, enabled, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, NOW(), NOW())
		RETURNING *`

	var created WebhookFilter
	err := r.db.GetContext(ctx, &created, query,
		filter.ID,
		filter.WebhookID,
		filter.FieldPath,
		filter.Operator,
		filter.Value,
		filter.LogicGroup,
		filter.Enabled,
	)
	if err != nil {
		return nil, err
	}

	return &created, nil
}

// UpdateFilter updates an existing webhook filter
func (r *Repository) UpdateFilter(ctx context.Context, filter *WebhookFilter) (*WebhookFilter, error) {
	query := `
		UPDATE webhook_filters
		SET field_path = $2, operator = $3, value = $4, logic_group = $5, enabled = $6, updated_at = NOW()
		WHERE id = $1
		RETURNING *`

	var updated WebhookFilter
	err := r.db.GetContext(ctx, &updated, query,
		filter.ID,
		filter.FieldPath,
		filter.Operator,
		filter.Value,
		filter.LogicGroup,
		filter.Enabled,
	)
	if err != nil {
		return nil, err
	}

	return &updated, nil
}

// DeleteFilter deletes a webhook filter
func (r *Repository) DeleteFilter(ctx context.Context, filterID string) error {
	query := `DELETE FROM webhook_filters WHERE id = $1`
	result, err := r.db.ExecContext(ctx, query, filterID)
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

// DeleteFiltersByWebhookID deletes all filters for a webhook
func (r *Repository) DeleteFiltersByWebhookID(ctx context.Context, webhookID string) error {
	query := `DELETE FROM webhook_filters WHERE webhook_id = $1`
	_, err := r.db.ExecContext(ctx, query, webhookID)
	return err
}
