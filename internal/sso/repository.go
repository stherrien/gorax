package sso

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

// Repository handles SSO provider database operations
type Repository interface {
	// Provider operations
	CreateProvider(ctx context.Context, provider *Provider) error
	GetProvider(ctx context.Context, id uuid.UUID) (*Provider, error)
	GetProviderByTenant(ctx context.Context, tenantID uuid.UUID, id uuid.UUID) (*Provider, error)
	ListProviders(ctx context.Context, tenantID uuid.UUID) ([]*Provider, error)
	UpdateProvider(ctx context.Context, provider *Provider) error
	DeleteProvider(ctx context.Context, id uuid.UUID) error
	GetProviderByDomain(ctx context.Context, domain string) (*Provider, error)

	// Connection operations
	CreateConnection(ctx context.Context, conn *Connection) error
	GetConnection(ctx context.Context, userID, providerID uuid.UUID) (*Connection, error)
	GetConnectionByExternalID(ctx context.Context, providerID uuid.UUID, externalID string) (*Connection, error)
	UpdateConnection(ctx context.Context, conn *Connection) error
	DeleteConnection(ctx context.Context, id uuid.UUID) error
	ListConnectionsByUser(ctx context.Context, userID uuid.UUID) ([]*Connection, error)

	// Login event operations
	CreateLoginEvent(ctx context.Context, event *LoginEvent) error
	ListLoginEvents(ctx context.Context, providerID uuid.UUID, limit int) ([]*LoginEvent, error)
}

// PostgresRepository implements Repository for PostgreSQL
type PostgresRepository struct {
	db *sqlx.DB
}

// NewRepository creates a new SSO repository
func NewRepository(db *sqlx.DB) Repository {
	return &PostgresRepository{db: db}
}

// CreateProvider creates a new SSO provider
func (r *PostgresRepository) CreateProvider(ctx context.Context, provider *Provider) error {
	query := `
		INSERT INTO sso_providers (
			id, tenant_id, name, provider_type, enabled, enforce_sso,
			config, domains, created_by, updated_by
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10
		)
		RETURNING created_at, updated_at
	`

	if provider.ID == uuid.Nil {
		provider.ID = uuid.New()
	}

	err := r.db.QueryRowContext(
		ctx, query,
		provider.ID, provider.TenantID, provider.Name, provider.Type,
		provider.Enabled, provider.EnforceSSO, provider.Config,
		pq.Array(provider.Domains), provider.CreatedBy, provider.UpdatedBy,
	).Scan(&provider.CreatedAt, &provider.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create SSO provider: %w", err)
	}

	return nil
}

// GetProvider retrieves an SSO provider by ID
func (r *PostgresRepository) GetProvider(ctx context.Context, id uuid.UUID) (*Provider, error) {
	query := `
		SELECT id, tenant_id, name, provider_type, enabled, enforce_sso,
		       config, domains, created_at, updated_at, created_by, updated_by
		FROM sso_providers
		WHERE id = $1
	`

	var provider Provider
	var domains pq.StringArray

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&provider.ID, &provider.TenantID, &provider.Name, &provider.Type,
		&provider.Enabled, &provider.EnforceSSO, &provider.Config, &domains,
		&provider.CreatedAt, &provider.UpdatedAt, &provider.CreatedBy, &provider.UpdatedBy,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("SSO provider not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get SSO provider: %w", err)
	}

	provider.Domains = domains

	return &provider, nil
}

// GetProviderByTenant retrieves an SSO provider by ID and tenant
func (r *PostgresRepository) GetProviderByTenant(ctx context.Context, tenantID, id uuid.UUID) (*Provider, error) {
	query := `
		SELECT id, tenant_id, name, provider_type, enabled, enforce_sso,
		       config, domains, created_at, updated_at, created_by, updated_by
		FROM sso_providers
		WHERE id = $1 AND tenant_id = $2
	`

	var provider Provider
	var domains pq.StringArray

	err := r.db.QueryRowContext(ctx, query, id, tenantID).Scan(
		&provider.ID, &provider.TenantID, &provider.Name, &provider.Type,
		&provider.Enabled, &provider.EnforceSSO, &provider.Config, &domains,
		&provider.CreatedAt, &provider.UpdatedAt, &provider.CreatedBy, &provider.UpdatedBy,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("SSO provider not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get SSO provider: %w", err)
	}

	provider.Domains = domains

	return &provider, nil
}

// ListProviders retrieves all SSO providers for a tenant
func (r *PostgresRepository) ListProviders(ctx context.Context, tenantID uuid.UUID) ([]*Provider, error) {
	query := `
		SELECT id, tenant_id, name, provider_type, enabled, enforce_sso,
		       config, domains, created_at, updated_at, created_by, updated_by
		FROM sso_providers
		WHERE tenant_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to list SSO providers: %w", err)
	}
	defer rows.Close()

	var providers []*Provider
	for rows.Next() {
		var provider Provider
		var domains pq.StringArray

		err := rows.Scan(
			&provider.ID, &provider.TenantID, &provider.Name, &provider.Type,
			&provider.Enabled, &provider.EnforceSSO, &provider.Config, &domains,
			&provider.CreatedAt, &provider.UpdatedAt, &provider.CreatedBy, &provider.UpdatedBy,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan SSO provider: %w", err)
		}

		provider.Domains = domains
		providers = append(providers, &provider)
	}

	return providers, nil
}

// UpdateProvider updates an existing SSO provider
func (r *PostgresRepository) UpdateProvider(ctx context.Context, provider *Provider) error {
	query := `
		UPDATE sso_providers
		SET name = $1, enabled = $2, enforce_sso = $3, config = $4,
		    domains = $5, updated_by = $6, updated_at = NOW()
		WHERE id = $7
		RETURNING updated_at
	`

	err := r.db.QueryRowContext(
		ctx, query,
		provider.Name, provider.Enabled, provider.EnforceSSO,
		provider.Config, pq.Array(provider.Domains),
		provider.UpdatedBy, provider.ID,
	).Scan(&provider.UpdatedAt)

	if err == sql.ErrNoRows {
		return fmt.Errorf("SSO provider not found")
	}
	if err != nil {
		return fmt.Errorf("failed to update SSO provider: %w", err)
	}

	return nil
}

// DeleteProvider deletes an SSO provider
func (r *PostgresRepository) DeleteProvider(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM sso_providers WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete SSO provider: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("SSO provider not found")
	}

	return nil
}

// GetProviderByDomain retrieves an SSO provider by email domain
func (r *PostgresRepository) GetProviderByDomain(ctx context.Context, domain string) (*Provider, error) {
	query := `
		SELECT id, tenant_id, name, provider_type, enabled, enforce_sso,
		       config, domains, created_at, updated_at, created_by, updated_by
		FROM sso_providers
		WHERE enabled = true AND $1 = ANY(domains)
		ORDER BY created_at DESC
		LIMIT 1
	`

	var provider Provider
	var domains pq.StringArray

	err := r.db.QueryRowContext(ctx, query, domain).Scan(
		&provider.ID, &provider.TenantID, &provider.Name, &provider.Type,
		&provider.Enabled, &provider.EnforceSSO, &provider.Config, &domains,
		&provider.CreatedAt, &provider.UpdatedAt, &provider.CreatedBy, &provider.UpdatedBy,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("SSO provider not found for domain")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get SSO provider by domain: %w", err)
	}

	provider.Domains = domains

	return &provider, nil
}

// CreateConnection creates a new SSO connection
func (r *PostgresRepository) CreateConnection(ctx context.Context, conn *Connection) error {
	query := `
		INSERT INTO sso_connections (
			id, user_id, sso_provider_id, external_id, attributes, last_login_at
		) VALUES (
			$1, $2, $3, $4, $5, $6
		)
		RETURNING created_at, updated_at
	`

	if conn.ID == uuid.Nil {
		conn.ID = uuid.New()
	}

	err := r.db.QueryRowContext(
		ctx, query,
		conn.ID, conn.UserID, conn.ProviderID, conn.ExternalID,
		conn.Attributes, conn.LastLoginAt,
	).Scan(&conn.CreatedAt, &conn.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create SSO connection: %w", err)
	}

	return nil
}

// GetConnection retrieves an SSO connection by user and provider
func (r *PostgresRepository) GetConnection(ctx context.Context, userID, providerID uuid.UUID) (*Connection, error) {
	query := `
		SELECT id, user_id, sso_provider_id, external_id, attributes,
		       last_login_at, created_at, updated_at
		FROM sso_connections
		WHERE user_id = $1 AND sso_provider_id = $2
	`

	var conn Connection
	err := r.db.GetContext(ctx, &conn, query, userID, providerID)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("SSO connection not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get SSO connection: %w", err)
	}

	return &conn, nil
}

// GetConnectionByExternalID retrieves an SSO connection by external ID
func (r *PostgresRepository) GetConnectionByExternalID(ctx context.Context, providerID uuid.UUID, externalID string) (*Connection, error) {
	query := `
		SELECT id, user_id, sso_provider_id, external_id, attributes,
		       last_login_at, created_at, updated_at
		FROM sso_connections
		WHERE sso_provider_id = $1 AND external_id = $2
	`

	var conn Connection
	err := r.db.GetContext(ctx, &conn, query, providerID, externalID)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get SSO connection: %w", err)
	}

	return &conn, nil
}

// UpdateConnection updates an existing SSO connection
func (r *PostgresRepository) UpdateConnection(ctx context.Context, conn *Connection) error {
	query := `
		UPDATE sso_connections
		SET attributes = $1, last_login_at = $2, updated_at = NOW()
		WHERE id = $3
		RETURNING updated_at
	`

	err := r.db.QueryRowContext(
		ctx, query,
		conn.Attributes, conn.LastLoginAt, conn.ID,
	).Scan(&conn.UpdatedAt)

	if err == sql.ErrNoRows {
		return fmt.Errorf("SSO connection not found")
	}
	if err != nil {
		return fmt.Errorf("failed to update SSO connection: %w", err)
	}

	return nil
}

// DeleteConnection deletes an SSO connection
func (r *PostgresRepository) DeleteConnection(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM sso_connections WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete SSO connection: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("SSO connection not found")
	}

	return nil
}

// ListConnectionsByUser retrieves all SSO connections for a user
func (r *PostgresRepository) ListConnectionsByUser(ctx context.Context, userID uuid.UUID) ([]*Connection, error) {
	query := `
		SELECT id, user_id, sso_provider_id, external_id, attributes,
		       last_login_at, created_at, updated_at
		FROM sso_connections
		WHERE user_id = $1
		ORDER BY created_at DESC
	`

	var connections []*Connection
	err := r.db.SelectContext(ctx, &connections, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to list SSO connections: %w", err)
	}

	return connections, nil
}

// CreateLoginEvent creates a new SSO login event
func (r *PostgresRepository) CreateLoginEvent(ctx context.Context, event *LoginEvent) error {
	query := `
		INSERT INTO sso_login_events (
			id, sso_provider_id, user_id, external_id, status,
			error_message, ip_address, user_agent
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8
		)
		RETURNING created_at
	`

	if event.ID == uuid.Nil {
		event.ID = uuid.New()
	}

	err := r.db.QueryRowContext(
		ctx, query,
		event.ID, event.ProviderID, event.UserID, event.ExternalID,
		event.Status, event.ErrorMessage, event.IPAddress, event.UserAgent,
	).Scan(&event.CreatedAt)

	if err != nil {
		return fmt.Errorf("failed to create SSO login event: %w", err)
	}

	return nil
}

// ListLoginEvents retrieves recent login events for a provider
func (r *PostgresRepository) ListLoginEvents(ctx context.Context, providerID uuid.UUID, limit int) ([]*LoginEvent, error) {
	if limit <= 0 {
		limit = 100
	}

	query := `
		SELECT id, sso_provider_id, user_id, external_id, status,
		       error_message, ip_address, user_agent, created_at
		FROM sso_login_events
		WHERE sso_provider_id = $1
		ORDER BY created_at DESC
		LIMIT $2
	`

	var events []*LoginEvent
	err := r.db.SelectContext(ctx, &events, query, providerID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to list SSO login events: %w", err)
	}

	return events, nil
}

// MaskSensitiveConfig masks sensitive data in provider config for display
func MaskSensitiveConfig(providerType ProviderType, config json.RawMessage) (json.RawMessage, error) {
	var configMap map[string]interface{}
	if err := json.Unmarshal(config, &configMap); err != nil {
		return nil, err
	}

	// Mask sensitive fields
	sensitiveFields := []string{"client_secret", "private_key", "certificate"}
	for _, field := range sensitiveFields {
		if _, exists := configMap[field]; exists {
			configMap[field] = "[REDACTED]"
		}
	}

	return json.Marshal(configMap)
}
