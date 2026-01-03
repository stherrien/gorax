package oauth

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

// PostgresRepository implements OAuthRepository using PostgreSQL
type PostgresRepository struct {
	db *sqlx.DB
}

// NewPostgresRepository creates a new PostgreSQL OAuth repository
func NewPostgresRepository(db *sqlx.DB) *PostgresRepository {
	return &PostgresRepository{
		db: db,
	}
}

// GetProviderByKey retrieves an OAuth provider by key
func (r *PostgresRepository) GetProviderByKey(ctx context.Context, providerKey string) (*OAuthProvider, error) {
	query := `
		SELECT id, provider_key, name, description, auth_url, token_url, user_info_url,
		       default_scopes, client_id, client_secret_encrypted, client_secret_nonce,
		       client_secret_auth_tag, client_secret_encrypted_dek, client_secret_kms_key_id,
		       status, config, created_at, updated_at
		FROM oauth_providers
		WHERE provider_key = $1 AND status = 'active'
	`

	var provider OAuthProvider
	var config []byte
	var defaultScopes pq.StringArray

	err := r.db.QueryRowContext(ctx, query, providerKey).Scan(
		&provider.ID,
		&provider.ProviderKey,
		&provider.Name,
		&provider.Description,
		&provider.AuthURL,
		&provider.TokenURL,
		&provider.UserInfoURL,
		&defaultScopes,
		&provider.ClientID,
		&provider.ClientSecretEncrypted,
		&provider.ClientSecretNonce,
		&provider.ClientSecretAuthTag,
		&provider.ClientSecretEncDEK,
		&provider.ClientSecretKMSKeyID,
		&provider.Status,
		&config,
		&provider.CreatedAt,
		&provider.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrInvalidProvider
		}
		return nil, fmt.Errorf("failed to get provider: %w", err)
	}

	provider.DefaultScopes = defaultScopes

	// Parse config JSON
	if len(config) > 0 {
		if err := json.Unmarshal(config, &provider.Config); err != nil {
			return nil, fmt.Errorf("failed to unmarshal provider config: %w", err)
		}
	}

	return &provider, nil
}

// ListProviders lists all available OAuth providers
func (r *PostgresRepository) ListProviders(ctx context.Context) ([]*OAuthProvider, error) {
	query := `
		SELECT id, provider_key, name, description, auth_url, token_url, user_info_url,
		       default_scopes, client_id, status, config, created_at, updated_at
		FROM oauth_providers
		WHERE status = 'active'
		ORDER BY name
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list providers: %w", err)
	}
	defer rows.Close()

	var providers []*OAuthProvider
	for rows.Next() {
		var provider OAuthProvider
		var config []byte
		var defaultScopes pq.StringArray

		err := rows.Scan(
			&provider.ID,
			&provider.ProviderKey,
			&provider.Name,
			&provider.Description,
			&provider.AuthURL,
			&provider.TokenURL,
			&provider.UserInfoURL,
			&defaultScopes,
			&provider.ClientID,
			&provider.Status,
			&config,
			&provider.CreatedAt,
			&provider.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan provider: %w", err)
		}

		provider.DefaultScopes = defaultScopes

		// Parse config JSON
		if len(config) > 0 {
			if err := json.Unmarshal(config, &provider.Config); err != nil {
				return nil, fmt.Errorf("failed to unmarshal provider config: %w", err)
			}
		}

		providers = append(providers, &provider)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating providers: %w", err)
	}

	return providers, nil
}

// CreateConnection creates a new OAuth connection
func (r *PostgresRepository) CreateConnection(ctx context.Context, conn *OAuthConnection) error {
	query := `
		INSERT INTO oauth_connections (
			id, user_id, tenant_id, provider_key, provider_user_id, provider_username, provider_email,
			access_token_encrypted, access_token_nonce, access_token_auth_tag, access_token_encrypted_dek, access_token_kms_key_id,
			refresh_token_encrypted, refresh_token_nonce, refresh_token_auth_tag, refresh_token_encrypted_dek, refresh_token_kms_key_id,
			token_expiry, scopes, status, raw_token_response, metadata
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22
		)
		ON CONFLICT (user_id, tenant_id, provider_key)
		DO UPDATE SET
			provider_user_id = EXCLUDED.provider_user_id,
			provider_username = EXCLUDED.provider_username,
			provider_email = EXCLUDED.provider_email,
			access_token_encrypted = EXCLUDED.access_token_encrypted,
			access_token_nonce = EXCLUDED.access_token_nonce,
			access_token_auth_tag = EXCLUDED.access_token_auth_tag,
			access_token_encrypted_dek = EXCLUDED.access_token_encrypted_dek,
			access_token_kms_key_id = EXCLUDED.access_token_kms_key_id,
			refresh_token_encrypted = EXCLUDED.refresh_token_encrypted,
			refresh_token_nonce = EXCLUDED.refresh_token_nonce,
			refresh_token_auth_tag = EXCLUDED.refresh_token_auth_tag,
			refresh_token_encrypted_dek = EXCLUDED.refresh_token_encrypted_dek,
			refresh_token_kms_key_id = EXCLUDED.refresh_token_kms_key_id,
			token_expiry = EXCLUDED.token_expiry,
			scopes = EXCLUDED.scopes,
			status = EXCLUDED.status,
			raw_token_response = EXCLUDED.raw_token_response,
			metadata = EXCLUDED.metadata,
			updated_at = NOW()
	`

	rawTokenJSON, err := json.Marshal(conn.RawTokenResponse)
	if err != nil {
		return fmt.Errorf("failed to marshal raw token response: %w", err)
	}

	metadataJSON, err := json.Marshal(conn.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	_, err = r.db.ExecContext(ctx, query,
		conn.ID,
		conn.UserID,
		conn.TenantID,
		conn.ProviderKey,
		conn.ProviderUserID,
		conn.ProviderUsername,
		conn.ProviderEmail,
		conn.AccessTokenEncrypted,
		conn.AccessTokenNonce,
		conn.AccessTokenAuthTag,
		conn.AccessTokenEncDEK,
		conn.AccessTokenKMSKeyID,
		conn.RefreshTokenEncrypted,
		conn.RefreshTokenNonce,
		conn.RefreshTokenAuthTag,
		conn.RefreshTokenEncDEK,
		conn.RefreshTokenKMSKeyID,
		conn.TokenExpiry,
		pq.Array(conn.Scopes),
		conn.Status,
		rawTokenJSON,
		metadataJSON,
	)

	if err != nil {
		return fmt.Errorf("failed to create connection: %w", err)
	}

	return nil
}

// GetConnection retrieves an OAuth connection by ID
func (r *PostgresRepository) GetConnection(ctx context.Context, id string) (*OAuthConnection, error) {
	query := `
		SELECT id, user_id, tenant_id, provider_key, provider_user_id, provider_username, provider_email,
		       access_token_encrypted, access_token_nonce, access_token_auth_tag, access_token_encrypted_dek, access_token_kms_key_id,
		       refresh_token_encrypted, refresh_token_nonce, refresh_token_auth_tag, refresh_token_encrypted_dek, refresh_token_kms_key_id,
		       token_expiry, scopes, status, created_at, updated_at, last_used_at, last_refresh_at,
		       raw_token_response, metadata
		FROM oauth_connections
		WHERE id = $1
	`

	return r.scanConnection(ctx, query, id)
}

// GetConnectionByUserProvider retrieves an OAuth connection by user, tenant, and provider
func (r *PostgresRepository) GetConnectionByUserProvider(ctx context.Context, userID, tenantID, providerKey string) (*OAuthConnection, error) {
	query := `
		SELECT id, user_id, tenant_id, provider_key, provider_user_id, provider_username, provider_email,
		       access_token_encrypted, access_token_nonce, access_token_auth_tag, access_token_encrypted_dek, access_token_kms_key_id,
		       refresh_token_encrypted, refresh_token_nonce, refresh_token_auth_tag, refresh_token_encrypted_dek, refresh_token_kms_key_id,
		       token_expiry, scopes, status, created_at, updated_at, last_used_at, last_refresh_at,
		       raw_token_response, metadata
		FROM oauth_connections
		WHERE user_id = $1 AND tenant_id = $2 AND provider_key = $3
	`

	return r.scanConnection(ctx, query, userID, tenantID, providerKey)
}

// scanConnection is a helper to scan a connection from a query
func (r *PostgresRepository) scanConnection(ctx context.Context, query string, args ...interface{}) (*OAuthConnection, error) {
	var conn OAuthConnection
	var scopes pq.StringArray
	var rawTokenJSON, metadataJSON []byte

	err := r.db.QueryRowContext(ctx, query, args...).Scan(
		&conn.ID,
		&conn.UserID,
		&conn.TenantID,
		&conn.ProviderKey,
		&conn.ProviderUserID,
		&conn.ProviderUsername,
		&conn.ProviderEmail,
		&conn.AccessTokenEncrypted,
		&conn.AccessTokenNonce,
		&conn.AccessTokenAuthTag,
		&conn.AccessTokenEncDEK,
		&conn.AccessTokenKMSKeyID,
		&conn.RefreshTokenEncrypted,
		&conn.RefreshTokenNonce,
		&conn.RefreshTokenAuthTag,
		&conn.RefreshTokenEncDEK,
		&conn.RefreshTokenKMSKeyID,
		&conn.TokenExpiry,
		&scopes,
		&conn.Status,
		&conn.CreatedAt,
		&conn.UpdatedAt,
		&conn.LastUsedAt,
		&conn.LastRefreshAt,
		&rawTokenJSON,
		&metadataJSON,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrConnectionNotFound
		}
		return nil, fmt.Errorf("failed to get connection: %w", err)
	}

	conn.Scopes = scopes

	// Parse raw token response
	if len(rawTokenJSON) > 0 {
		if err := json.Unmarshal(rawTokenJSON, &conn.RawTokenResponse); err != nil {
			return nil, fmt.Errorf("failed to unmarshal raw token response: %w", err)
		}
	}

	// Parse metadata
	if len(metadataJSON) > 0 {
		if err := json.Unmarshal(metadataJSON, &conn.Metadata); err != nil {
			return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
		}
	}

	return &conn, nil
}

// ListConnectionsByUser lists all OAuth connections for a user
func (r *PostgresRepository) ListConnectionsByUser(ctx context.Context, userID, tenantID string) ([]*OAuthConnection, error) {
	query := `
		SELECT id, user_id, tenant_id, provider_key, provider_user_id, provider_username, provider_email,
		       token_expiry, scopes, status, created_at, updated_at, last_used_at, last_refresh_at
		FROM oauth_connections
		WHERE user_id = $1 AND tenant_id = $2
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, userID, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to list connections: %w", err)
	}
	defer rows.Close()

	var connections []*OAuthConnection
	for rows.Next() {
		var conn OAuthConnection
		var scopes pq.StringArray

		err := rows.Scan(
			&conn.ID,
			&conn.UserID,
			&conn.TenantID,
			&conn.ProviderKey,
			&conn.ProviderUserID,
			&conn.ProviderUsername,
			&conn.ProviderEmail,
			&conn.TokenExpiry,
			&scopes,
			&conn.Status,
			&conn.CreatedAt,
			&conn.UpdatedAt,
			&conn.LastUsedAt,
			&conn.LastRefreshAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan connection: %w", err)
		}

		conn.Scopes = scopes
		connections = append(connections, &conn)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating connections: %w", err)
	}

	return connections, nil
}

// UpdateConnection updates an OAuth connection
func (r *PostgresRepository) UpdateConnection(ctx context.Context, conn *OAuthConnection) error {
	query := `
		UPDATE oauth_connections
		SET provider_user_id = $1,
		    provider_username = $2,
		    provider_email = $3,
		    access_token_encrypted = $4,
		    access_token_nonce = $5,
		    access_token_auth_tag = $6,
		    access_token_encrypted_dek = $7,
		    access_token_kms_key_id = $8,
		    refresh_token_encrypted = $9,
		    refresh_token_nonce = $10,
		    refresh_token_auth_tag = $11,
		    refresh_token_encrypted_dek = $12,
		    refresh_token_kms_key_id = $13,
		    token_expiry = $14,
		    scopes = $15,
		    status = $16,
		    last_used_at = $17,
		    last_refresh_at = $18,
		    metadata = $19,
		    updated_at = NOW()
		WHERE id = $20
	`

	metadataJSON, err := json.Marshal(conn.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	result, err := r.db.ExecContext(ctx, query,
		conn.ProviderUserID,
		conn.ProviderUsername,
		conn.ProviderEmail,
		conn.AccessTokenEncrypted,
		conn.AccessTokenNonce,
		conn.AccessTokenAuthTag,
		conn.AccessTokenEncDEK,
		conn.AccessTokenKMSKeyID,
		conn.RefreshTokenEncrypted,
		conn.RefreshTokenNonce,
		conn.RefreshTokenAuthTag,
		conn.RefreshTokenEncDEK,
		conn.RefreshTokenKMSKeyID,
		conn.TokenExpiry,
		pq.Array(conn.Scopes),
		conn.Status,
		conn.LastUsedAt,
		conn.LastRefreshAt,
		metadataJSON,
		conn.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update connection: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return ErrConnectionNotFound
	}

	return nil
}

// DeleteConnection deletes an OAuth connection
func (r *PostgresRepository) DeleteConnection(ctx context.Context, id string) error {
	query := `DELETE FROM oauth_connections WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete connection: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return ErrConnectionNotFound
	}

	return nil
}

// CreateState creates a new OAuth state
func (r *PostgresRepository) CreateState(ctx context.Context, state *OAuthState) error {
	query := `
		INSERT INTO oauth_states (
			state, user_id, tenant_id, provider_key, redirect_uri, code_verifier,
			scopes, metadata, created_at, expires_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	metadataJSON, err := json.Marshal(state.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	_, err = r.db.ExecContext(ctx, query,
		state.State,
		state.UserID,
		state.TenantID,
		state.ProviderKey,
		state.RedirectURI,
		state.CodeVerifier,
		pq.Array(state.Scopes),
		metadataJSON,
		state.CreatedAt,
		state.ExpiresAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create state: %w", err)
	}

	return nil
}

// GetState retrieves an OAuth state
func (r *PostgresRepository) GetState(ctx context.Context, stateStr string) (*OAuthState, error) {
	query := `
		SELECT state, user_id, tenant_id, provider_key, redirect_uri, code_verifier,
		       scopes, metadata, created_at, expires_at, used
		FROM oauth_states
		WHERE state = $1
	`

	var state OAuthState
	var scopes pq.StringArray
	var metadataJSON []byte

	err := r.db.QueryRowContext(ctx, query, stateStr).Scan(
		&state.State,
		&state.UserID,
		&state.TenantID,
		&state.ProviderKey,
		&state.RedirectURI,
		&state.CodeVerifier,
		&scopes,
		&metadataJSON,
		&state.CreatedAt,
		&state.ExpiresAt,
		&state.Used,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrInvalidState
		}
		return nil, fmt.Errorf("failed to get state: %w", err)
	}

	state.Scopes = scopes

	// Parse metadata
	if len(metadataJSON) > 0 {
		if err := json.Unmarshal(metadataJSON, &state.Metadata); err != nil {
			return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
		}
	}

	return &state, nil
}

// MarkStateUsed marks an OAuth state as used
func (r *PostgresRepository) MarkStateUsed(ctx context.Context, stateStr string) error {
	query := `UPDATE oauth_states SET used = TRUE WHERE state = $1`

	_, err := r.db.ExecContext(ctx, query, stateStr)
	if err != nil {
		return fmt.Errorf("failed to mark state as used: %w", err)
	}

	return nil
}

// DeleteExpiredStates deletes expired OAuth states
func (r *PostgresRepository) DeleteExpiredStates(ctx context.Context) (int, error) {
	query := `DELETE FROM oauth_states WHERE expires_at < NOW()`

	result, err := r.db.ExecContext(ctx, query)
	if err != nil {
		return 0, fmt.Errorf("failed to delete expired states: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to get rows affected: %w", err)
	}

	return int(rowsAffected), nil
}

// CreateLog creates an OAuth connection log entry
func (r *PostgresRepository) CreateLog(ctx context.Context, log *OAuthConnectionLog) error {
	query := `
		INSERT INTO oauth_connection_logs (
			id, connection_id, user_id, tenant_id, action, success, error_message, metadata
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	metadataJSON, err := json.Marshal(log.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	_, err = r.db.ExecContext(ctx, query,
		log.ID,
		log.ConnectionID,
		log.UserID,
		log.TenantID,
		log.Action,
		log.Success,
		log.ErrorMessage,
		metadataJSON,
	)

	if err != nil {
		return fmt.Errorf("failed to create log: %w", err)
	}

	return nil
}
