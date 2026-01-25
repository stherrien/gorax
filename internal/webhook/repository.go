package webhook

import (
	"context"
	"database/sql"
	"encoding/json"
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

	// Marshal metadata to JSONB if present
	var metadataJSON []byte
	var err error
	if event.Metadata != nil {
		metadataJSON, err = json.Marshal(event.Metadata)
		if err != nil {
			return fmt.Errorf("failed to marshal metadata: %w", err)
		}
	}

	query := `
		INSERT INTO webhook_events (
			id, tenant_id, webhook_id, execution_id, request_method,
			request_headers, request_body, response_status, processing_time_ms,
			status, error_message, filtered_reason, metadata, created_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
	`

	_, err = r.db.ExecContext(
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
		metadataJSON,
		event.CreatedAt,
	)

	if err != nil {
		return err
	}

	return nil
}

// webhookEventDB is the database representation with JSONB as RawMessage
type webhookEventDB struct {
	ID                string             `db:"id"`
	TenantID          string             `db:"tenant_id"`
	WebhookID         string             `db:"webhook_id"`
	ExecutionID       *string            `db:"execution_id"`
	RequestMethod     string             `db:"request_method"`
	RequestHeaders    json.RawMessage    `db:"request_headers"`
	RequestBody       json.RawMessage    `db:"request_body"`
	ResponseStatus    *int               `db:"response_status"`
	ProcessingTimeMs  *int               `db:"processing_time_ms"`
	Status            WebhookEventStatus `db:"status"`
	ErrorMessage      *string            `db:"error_message"`
	FilteredReason    *string            `db:"filtered_reason"`
	ReplayCount       int                `db:"replay_count"`
	SourceEventID     *string            `db:"source_event_id"`
	Metadata          json.RawMessage    `db:"metadata"`
	RetryCount        int                `db:"retry_count"`
	MaxRetries        int                `db:"max_retries"`
	NextRetryAt       *time.Time         `db:"next_retry_at"`
	LastRetryAt       *time.Time         `db:"last_retry_at"`
	RetryError        *string            `db:"retry_error"`
	PermanentlyFailed bool               `db:"permanently_failed"`
	CreatedAt         time.Time          `db:"created_at"`
}

// toWebhookEvent converts webhookEventDB to WebhookEvent
func (db *webhookEventDB) toWebhookEvent() (*WebhookEvent, error) {
	event := &WebhookEvent{
		ID:                db.ID,
		TenantID:          db.TenantID,
		WebhookID:         db.WebhookID,
		ExecutionID:       db.ExecutionID,
		RequestMethod:     db.RequestMethod,
		RequestBody:       db.RequestBody,
		ResponseStatus:    db.ResponseStatus,
		ProcessingTimeMs:  db.ProcessingTimeMs,
		Status:            db.Status,
		ErrorMessage:      db.ErrorMessage,
		FilteredReason:    db.FilteredReason,
		ReplayCount:       db.ReplayCount,
		SourceEventID:     db.SourceEventID,
		RetryCount:        db.RetryCount,
		MaxRetries:        db.MaxRetries,
		NextRetryAt:       db.NextRetryAt,
		LastRetryAt:       db.LastRetryAt,
		RetryError:        db.RetryError,
		PermanentlyFailed: db.PermanentlyFailed,
		CreatedAt:         db.CreatedAt,
	}

	// Unmarshal request headers
	if len(db.RequestHeaders) > 0 {
		if err := json.Unmarshal(db.RequestHeaders, &event.RequestHeaders); err != nil {
			return nil, fmt.Errorf("failed to unmarshal request headers: %w", err)
		}
	}

	// Unmarshal metadata if present
	if len(db.Metadata) > 0 {
		var metadata EventMetadata
		if err := json.Unmarshal(db.Metadata, &metadata); err != nil {
			return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
		}
		event.Metadata = &metadata
	}

	return event, nil
}

// GetEventByID retrieves a webhook event by ID with tenant isolation
func (r *Repository) GetEventByID(ctx context.Context, tenantID, eventID string) (*WebhookEvent, error) {
	query := `SELECT * FROM webhook_events WHERE id = $1 AND tenant_id = $2`

	var eventDB webhookEventDB
	err := r.db.GetContext(ctx, &eventDB, query, eventID, tenantID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	return eventDB.toWebhookEvent()
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

	var eventsDB []webhookEventDB
	err = r.db.SelectContext(ctx, &eventsDB, query, args...)
	if err != nil {
		return nil, 0, err
	}

	// Convert database structs to domain structs
	events := make([]*WebhookEvent, 0, len(eventsDB))
	for _, eventDB := range eventsDB {
		event, err := eventDB.toWebhookEvent()
		if err != nil {
			return nil, 0, fmt.Errorf("failed to convert event: %w", err)
		}
		events = append(events, event)
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

// DeleteOldEvents deletes webhook events older than the retention period in batches
func (r *Repository) DeleteOldEvents(ctx context.Context, retentionPeriod time.Duration, batchSize int) (int, error) {
	cutoffTime := time.Now().Add(-retentionPeriod)

	query := `
		DELETE FROM webhook_events
		WHERE id IN (
			SELECT id FROM webhook_events
			WHERE created_at < $1
			ORDER BY created_at
			LIMIT $2
		)
	`

	result, err := r.db.ExecContext(ctx, query, cutoffTime, batchSize)
	if err != nil {
		return 0, fmt.Errorf("delete old events: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("get rows affected: %w", err)
	}

	return int(rows), nil
}

// MarkEventForRetry marks a webhook event for retry with exponential backoff
func (r *Repository) MarkEventForRetry(ctx context.Context, eventID string, errorMsg string) error {
	query := `SELECT mark_webhook_event_for_retry($1, $2)`

	_, err := r.db.ExecContext(ctx, query, eventID, errorMsg)
	if err != nil {
		return fmt.Errorf("mark event for retry: %w", err)
	}

	return nil
}

// MarkEventAsNonRetryable marks a webhook event as permanently failed without retry
func (r *Repository) MarkEventAsNonRetryable(ctx context.Context, eventID string, errorMsg string) error {
	query := `
		UPDATE webhook_events
		SET
			permanently_failed = true,
			retry_error = $2,
			status = 'failed',
			next_retry_at = NULL
		WHERE id = $1
	`

	result, err := r.db.ExecContext(ctx, query, eventID, errorMsg)
	if err != nil {
		return fmt.Errorf("mark event as non-retryable: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("get rows affected: %w", err)
	}

	if rows == 0 {
		return ErrNotFound
	}

	return nil
}

// GetEventsForRetry retrieves webhook events that are ready for retry
func (r *Repository) GetEventsForRetry(ctx context.Context, batchSize int) ([]*WebhookEvent, error) {
	query := `SELECT * FROM get_webhook_events_for_retry($1)`

	var eventsDB []webhookEventDB
	err := r.db.SelectContext(ctx, &eventsDB, query, batchSize)
	if err != nil {
		return nil, fmt.Errorf("get events for retry: %w", err)
	}

	events := make([]*WebhookEvent, 0, len(eventsDB))
	for _, eventDB := range eventsDB {
		event, err := eventDB.toWebhookEvent()
		if err != nil {
			return nil, fmt.Errorf("convert event: %w", err)
		}
		events = append(events, event)
	}

	return events, nil
}

// MarkEventRetrySucceeded marks a retry attempt as successful
func (r *Repository) MarkEventRetrySucceeded(ctx context.Context, eventID string, executionID string, processingTimeMs int) error {
	query := `
		UPDATE webhook_events
		SET
			status = 'processed',
			execution_id = $2,
			processing_time_ms = $3,
			next_retry_at = NULL
		WHERE id = $1
	`

	result, err := r.db.ExecContext(ctx, query, eventID, executionID, processingTimeMs)
	if err != nil {
		return fmt.Errorf("mark retry succeeded: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("get rows affected: %w", err)
	}

	if rows == 0 {
		return ErrNotFound
	}

	return nil
}

// GetRetryStatistics retrieves retry statistics for a webhook
func (r *Repository) GetRetryStatistics(ctx context.Context, webhookID string) (*RetryStatistics, error) {
	query := `SELECT * FROM webhook_retry_stats WHERE webhook_id = $1`

	var stats RetryStatistics
	err := r.db.GetContext(ctx, &stats, query, webhookID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// Return empty stats if no retry data exists
			return &RetryStatistics{
				WebhookID: webhookID,
			}, nil
		}
		return nil, fmt.Errorf("get retry statistics: %w", err)
	}

	return &stats, nil
}
