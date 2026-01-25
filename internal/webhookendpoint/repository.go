package webhookendpoint

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

// Repository defines the interface for webhook endpoint data access
type Repository interface {
	Create(ctx context.Context, endpoint *WebhookEndpoint) error
	GetByID(ctx context.Context, id uuid.UUID) (*WebhookEndpoint, error)
	GetByToken(ctx context.Context, token string) (*WebhookEndpoint, error)
	GetByExecutionAndStep(ctx context.Context, executionID uuid.UUID, stepID string) (*WebhookEndpoint, error)
	List(ctx context.Context, filter EndpointFilter) ([]*WebhookEndpoint, error)
	Update(ctx context.Context, endpoint *WebhookEndpoint) error
	Delete(ctx context.Context, id uuid.UUID) error
	GetExpiredEndpoints(ctx context.Context, limit int) ([]*WebhookEndpoint, error)
	DeactivateByExecution(ctx context.Context, executionID uuid.UUID) error
	DeleteExpired(ctx context.Context, olderThan int) (int64, error)
}

type repository struct {
	db *sqlx.DB
}

// NewRepository creates a new webhook endpoint repository
func NewRepository(db *sqlx.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Create(ctx context.Context, endpoint *WebhookEndpoint) error {
	query := `
		INSERT INTO webhook_endpoints (
			tenant_id, execution_id, step_id, endpoint_token, config,
			is_active, expires_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7
		) RETURNING id, created_at, updated_at
	`

	err := r.db.QueryRowxContext(
		ctx, query,
		endpoint.TenantID,
		endpoint.ExecutionID,
		endpoint.StepID,
		endpoint.EndpointToken,
		endpoint.Config,
		endpoint.IsActive,
		endpoint.ExpiresAt,
	).Scan(&endpoint.ID, &endpoint.CreatedAt, &endpoint.UpdatedAt)

	if err != nil {
		return fmt.Errorf("create webhook endpoint: %w", err)
	}

	return nil
}

func (r *repository) GetByID(ctx context.Context, id uuid.UUID) (*WebhookEndpoint, error) {
	var endpoint WebhookEndpoint

	query := `
		SELECT id, tenant_id, execution_id, step_id, endpoint_token, config,
			is_active, triggered_at, payload, source_ip, user_agent, content_type,
			expires_at, created_at, updated_at
		FROM webhook_endpoints
		WHERE id = $1
	`

	err := r.db.GetContext(ctx, &endpoint, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrEndpointNotFound
		}
		return nil, fmt.Errorf("get webhook endpoint by id: %w", err)
	}

	return &endpoint, nil
}

func (r *repository) GetByToken(ctx context.Context, token string) (*WebhookEndpoint, error) {
	var endpoint WebhookEndpoint

	query := `
		SELECT id, tenant_id, execution_id, step_id, endpoint_token, config,
			is_active, triggered_at, payload, source_ip, user_agent, content_type,
			expires_at, created_at, updated_at
		FROM webhook_endpoints
		WHERE endpoint_token = $1
	`

	err := r.db.GetContext(ctx, &endpoint, query, token)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrEndpointNotFound
		}
		return nil, fmt.Errorf("get webhook endpoint by token: %w", err)
	}

	return &endpoint, nil
}

func (r *repository) GetByExecutionAndStep(ctx context.Context, executionID uuid.UUID, stepID string) (*WebhookEndpoint, error) {
	var endpoint WebhookEndpoint

	query := `
		SELECT id, tenant_id, execution_id, step_id, endpoint_token, config,
			is_active, triggered_at, payload, source_ip, user_agent, content_type,
			expires_at, created_at, updated_at
		FROM webhook_endpoints
		WHERE execution_id = $1 AND step_id = $2
		ORDER BY created_at DESC
		LIMIT 1
	`

	err := r.db.GetContext(ctx, &endpoint, query, executionID, stepID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrEndpointNotFound
		}
		return nil, fmt.Errorf("get webhook endpoint by execution and step: %w", err)
	}

	return &endpoint, nil
}

func (r *repository) List(ctx context.Context, filter EndpointFilter) ([]*WebhookEndpoint, error) {
	var endpoints []*WebhookEndpoint

	query := `
		SELECT id, tenant_id, execution_id, step_id, endpoint_token, config,
			is_active, triggered_at, payload, source_ip, user_agent, content_type,
			expires_at, created_at, updated_at
		FROM webhook_endpoints
		WHERE tenant_id = $1
	`

	args := []interface{}{filter.TenantID}
	argPos := 2

	if filter.ExecutionID != nil {
		query += fmt.Sprintf(" AND execution_id = $%d", argPos)
		args = append(args, *filter.ExecutionID)
		argPos++
	}

	if filter.IsActive != nil {
		query += fmt.Sprintf(" AND is_active = $%d", argPos)
		args = append(args, *filter.IsActive)
		argPos++
	}

	if filter.ExpiredOnly {
		query += " AND expires_at < NOW()"
	}

	query += " ORDER BY created_at DESC"

	if filter.Limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argPos)
		args = append(args, filter.Limit)
		argPos++
	}

	if filter.Offset > 0 {
		query += fmt.Sprintf(" OFFSET $%d", argPos)
		args = append(args, filter.Offset)
	}

	err := r.db.SelectContext(ctx, &endpoints, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list webhook endpoints: %w", err)
	}

	return endpoints, nil
}

func (r *repository) Update(ctx context.Context, endpoint *WebhookEndpoint) error {
	query := `
		UPDATE webhook_endpoints
		SET is_active = $1,
			triggered_at = $2,
			payload = $3,
			source_ip = $4,
			user_agent = $5,
			content_type = $6,
			config = $7
		WHERE id = $8
	`

	result, err := r.db.ExecContext(
		ctx, query,
		endpoint.IsActive,
		endpoint.TriggeredAt,
		endpoint.Payload,
		endpoint.SourceIP,
		endpoint.UserAgent,
		endpoint.ContentType,
		endpoint.Config,
		endpoint.ID,
	)
	if err != nil {
		return fmt.Errorf("update webhook endpoint: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("check rows affected: %w", err)
	}

	if rows == 0 {
		return ErrEndpointNotFound
	}

	return nil
}

func (r *repository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM webhook_endpoints WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("delete webhook endpoint: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("check rows affected: %w", err)
	}

	if rows == 0 {
		return ErrEndpointNotFound
	}

	return nil
}

func (r *repository) GetExpiredEndpoints(ctx context.Context, limit int) ([]*WebhookEndpoint, error) {
	var endpoints []*WebhookEndpoint

	query := `
		SELECT id, tenant_id, execution_id, step_id, endpoint_token, config,
			is_active, triggered_at, payload, source_ip, user_agent, content_type,
			expires_at, created_at, updated_at
		FROM webhook_endpoints
		WHERE is_active = TRUE AND expires_at < NOW()
		ORDER BY expires_at ASC
		LIMIT $1
	`

	err := r.db.SelectContext(ctx, &endpoints, query, limit)
	if err != nil {
		return nil, fmt.Errorf("get expired endpoints: %w", err)
	}

	return endpoints, nil
}

func (r *repository) DeactivateByExecution(ctx context.Context, executionID uuid.UUID) error {
	query := `
		UPDATE webhook_endpoints
		SET is_active = FALSE
		WHERE execution_id = $1 AND is_active = TRUE
	`

	_, err := r.db.ExecContext(ctx, query, executionID)
	if err != nil {
		return fmt.Errorf("deactivate endpoints by execution: %w", err)
	}

	return nil
}

func (r *repository) DeleteExpired(ctx context.Context, olderThanDays int) (int64, error) {
	query := `
		DELETE FROM webhook_endpoints
		WHERE is_active = FALSE
			AND (
				triggered_at < NOW() - INTERVAL '1 day' * $1
				OR (triggered_at IS NULL AND expires_at < NOW() - INTERVAL '1 day' * $1)
			)
	`

	result, err := r.db.ExecContext(ctx, query, olderThanDays)
	if err != nil {
		return 0, fmt.Errorf("delete expired endpoints: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("check rows affected: %w", err)
	}

	return rows, nil
}
