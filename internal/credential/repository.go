package credential

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

// Repository handles credential database operations
type Repository struct {
	db *sqlx.DB
}

// NewRepository creates a new credential repository
func NewRepository(db *sqlx.DB) *Repository {
	return &Repository{db: db}
}

// Create inserts a new credential
func (r *Repository) Create(ctx context.Context, tenantID, createdBy string, cred *Credential) (*Credential, error) {
	if tenantID == "" {
		return nil, ErrInvalidTenantID
	}

	if createdBy == "" {
		return nil, errors.New("created_by cannot be empty")
	}

	if cred == nil {
		return nil, ErrEmptyCredentialData
	}

	if cred.Name == "" {
		return nil, ErrInvalidCredentialName
	}

	// Validate credential type
	if cred.Type != TypeAPIKey && cred.Type != TypeOAuth2 && cred.Type != TypeBasicAuth && cred.Type != TypeBearerToken && cred.Type != TypeCustom {
		return nil, ErrInvalidCredentialType
	}

	// Generate ID if not provided
	if cred.ID == "" {
		cred.ID = uuid.NewString()
	}

	now := time.Now()

	// Set status if not provided
	if cred.Status == "" {
		cred.Status = StatusActive
	}

	// Start transaction for RLS context
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Set tenant context for RLS within transaction using set_config
	_, err = tx.ExecContext(ctx, "SELECT set_config('app.current_tenant_id', $1, true)", tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to set tenant context: %w", err)
	}

	query := `
		INSERT INTO credentials (
			id, tenant_id, name, type, description, status,
			encrypted_dek, ciphertext, nonce, auth_tag, kms_key_id,
			metadata, created_by, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6,
			$7, $8, $9, $10, $11,
			$12, $13, $14, $15
		) RETURNING *
	`

	var created Credential
	err = tx.QueryRowxContext(
		ctx, query,
		cred.ID, tenantID, cred.Name, cred.Type, cred.Description, cred.Status,
		cred.EncryptedDEK, cred.Ciphertext, cred.Nonce, cred.AuthTag, cred.KMSKeyID,
		cred.Metadata, createdBy, now, now,
	).StructScan(&created)

	if err != nil {
		// Check for duplicate name constraint
		if pqErr, ok := err.(*pq.Error); ok {
			if pqErr.Code == "23505" { // unique_violation
				return nil, ErrDuplicateCredential
			}
		}
		return nil, fmt.Errorf("database insert failed: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return &created, nil
}

// GetByID retrieves a credential by ID (tenant-scoped)
func (r *Repository) GetByID(ctx context.Context, tenantID, id string) (*Credential, error) {
	if tenantID == "" {
		return nil, ErrInvalidTenantID
	}

	if id == "" {
		return nil, ErrInvalidCredentialID
	}

	// Start transaction for RLS context
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Set tenant context for RLS within transaction using set_config
	_, err = tx.ExecContext(ctx, "SELECT set_config('app.current_tenant_id', $1, true)", tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to set tenant context: %w", err)
	}

	query := `SELECT * FROM credentials WHERE id = $1 AND tenant_id = $2`

	var cred Credential
	err = tx.GetContext(ctx, &cred, query, id, tenantID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to get credential: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return &cred, nil
}

// GetByName retrieves a credential by name (tenant-scoped)
func (r *Repository) GetByName(ctx context.Context, tenantID, name string) (*Credential, error) {
	if tenantID == "" {
		return nil, ErrInvalidTenantID
	}

	if name == "" {
		return nil, ErrInvalidCredentialName
	}

	// Start transaction for RLS context
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Set tenant context for RLS within transaction using set_config
	_, err = tx.ExecContext(ctx, "SELECT set_config('app.current_tenant_id', $1, true)", tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to set tenant context: %w", err)
	}

	query := `SELECT * FROM credentials WHERE name = $1 AND tenant_id = $2`

	var cred Credential
	err = tx.GetContext(ctx, &cred, query, name, tenantID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to get credential: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return &cred, nil
}

// Update updates a credential
func (r *Repository) Update(ctx context.Context, tenantID, id string, input *UpdateCredentialInput) (*Credential, error) {
	if tenantID == "" {
		return nil, ErrInvalidTenantID
	}

	if id == "" {
		return nil, ErrInvalidCredentialID
	}

	if input == nil {
		return nil, errors.New("update input cannot be nil")
	}

	// First check if credential exists
	_, err := r.GetByID(ctx, tenantID, id)
	if err != nil {
		return nil, err
	}

	// Build dynamic update query
	updates := []string{}
	args := []interface{}{}
	argIndex := 1

	if input.Name != nil && *input.Name != "" {
		updates = append(updates, fmt.Sprintf("name = $%d", argIndex))
		args = append(args, *input.Name)
		argIndex++
	}

	if input.Description != nil {
		updates = append(updates, fmt.Sprintf("description = $%d", argIndex))
		args = append(args, *input.Description)
		argIndex++
	}

	if input.Status != nil {
		updates = append(updates, fmt.Sprintf("status = $%d", argIndex))
		args = append(args, *input.Status)
		argIndex++
	}

	if input.Metadata != nil {
		metadataJSON, err := json.Marshal(input.Metadata)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal metadata: %w", err)
		}
		updates = append(updates, fmt.Sprintf("metadata = $%d", argIndex))
		args = append(args, metadataJSON)
		argIndex++
	}

	// Always update updated_at
	updates = append(updates, fmt.Sprintf("updated_at = $%d", argIndex))
	args = append(args, time.Now())
	argIndex++

	if len(updates) == 1 { // Only updated_at was set
		return nil, errors.New("no fields to update")
	}

	// Add WHERE clause parameters
	args = append(args, id, tenantID)

	// Start transaction for RLS context
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Set tenant context for RLS within transaction using set_config
	_, err = tx.ExecContext(ctx, "SELECT set_config('app.current_tenant_id', $1, true)", tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to set tenant context: %w", err)
	}

	query := fmt.Sprintf(`
		UPDATE credentials
		SET %s
		WHERE id = $%d AND tenant_id = $%d
		RETURNING *
	`, joinUpdates(updates), argIndex, argIndex+1)

	var updated Credential
	err = tx.QueryRowxContext(ctx, query, args...).StructScan(&updated)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to update credential: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return &updated, nil
}

// Delete deletes a credential
func (r *Repository) Delete(ctx context.Context, tenantID, id string) error {
	if tenantID == "" {
		return ErrInvalidTenantID
	}

	if id == "" {
		return ErrInvalidCredentialID
	}

	// Start transaction for RLS context
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Set tenant context for RLS within transaction using set_config
	_, err = tx.ExecContext(ctx, "SELECT set_config('app.current_tenant_id', $1, true)", tenantID)
	if err != nil {
		return fmt.Errorf("failed to set tenant context: %w", err)
	}

	query := `DELETE FROM credentials WHERE id = $1 AND tenant_id = $2`

	result, err := tx.ExecContext(ctx, query, id, tenantID)
	if err != nil {
		return fmt.Errorf("failed to delete credential: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return ErrNotFound
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// List retrieves all credentials for a tenant with optional filtering
func (r *Repository) List(ctx context.Context, tenantID string, filter CredentialListFilter) ([]*Credential, error) {
	if tenantID == "" {
		return nil, ErrInvalidTenantID
	}

	// Start transaction for RLS context
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Set tenant context for RLS within transaction using set_config
	_, err = tx.ExecContext(ctx, "SELECT set_config('app.current_tenant_id', $1, true)", tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to set tenant context: %w", err)
	}

	query := `SELECT * FROM credentials WHERE tenant_id = $1`
	args := []interface{}{tenantID}
	argIndex := 2

	// Apply type filter
	if filter.Type != "" {
		query += fmt.Sprintf(" AND type = $%d", argIndex)
		args = append(args, filter.Type)
		argIndex++
	}

	// Apply status filter
	if filter.Status != "" {
		query += fmt.Sprintf(" AND status = $%d", argIndex)
		args = append(args, filter.Status)
		argIndex++
	}

	// Apply search filter (searches in name and description)
	if filter.Search != "" {
		query += fmt.Sprintf(" AND (name ILIKE $%d OR description ILIKE $%d)", argIndex, argIndex)
		searchPattern := "%" + filter.Search + "%"
		args = append(args, searchPattern)
	}

	query += " ORDER BY created_at DESC"

	var credentials []*Credential
	err = tx.SelectContext(ctx, &credentials, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list credentials: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	if credentials == nil {
		credentials = []*Credential{}
	}

	return credentials, nil
}

// UpdateLastUsedAt updates the last_used_at timestamp
func (r *Repository) UpdateLastUsedAt(ctx context.Context, tenantID, id string) error {
	if tenantID == "" {
		return ErrInvalidTenantID
	}

	if id == "" {
		return ErrInvalidCredentialID
	}

	// Start transaction for RLS context
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Set tenant context for RLS within transaction using set_config
	_, err = tx.ExecContext(ctx, "SELECT set_config('app.current_tenant_id', $1, true)", tenantID)
	if err != nil {
		return fmt.Errorf("failed to set tenant context: %w", err)
	}

	query := `
		UPDATE credentials
		SET last_used_at = $1
		WHERE id = $2 AND tenant_id = $3
	`

	result, err := tx.ExecContext(ctx, query, time.Now(), id, tenantID)
	if err != nil {
		return fmt.Errorf("failed to update last_used_at: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return ErrNotFound
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// LogAccess logs a credential access event
func (r *Repository) LogAccess(ctx context.Context, log *AccessLog) error {
	if log == nil {
		return errors.New("access log cannot be nil")
	}

	if log.CredentialID == "" {
		return ErrInvalidCredentialID
	}

	if log.TenantID == "" {
		return ErrInvalidTenantID
	}

	if log.AccessedBy == "" {
		return errors.New("accessed_by cannot be empty")
	}

	if log.ID == "" {
		log.ID = uuid.NewString()
	}

	// Start transaction for RLS context
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Set tenant context for RLS within transaction using set_config
	_, err = tx.ExecContext(ctx, "SELECT set_config('app.current_tenant_id', $1, true)", log.TenantID)
	if err != nil {
		return fmt.Errorf("failed to set tenant context: %w", err)
	}

	query := `
		INSERT INTO credential_access_log (
			id, credential_id, tenant_id, accessed_by, access_type,
			accessed_at, ip_address, user_agent, success, error_message
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10
		)
	`

	now := time.Now()

	_, err = tx.ExecContext(
		ctx, query,
		log.ID, log.CredentialID, log.TenantID, log.AccessedBy, log.AccessType,
		now, log.IPAddress, log.UserAgent, log.Success, log.ErrorMessage,
	)

	if err != nil {
		return fmt.Errorf("failed to log access: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// GetAccessLogs retrieves access log entries for a credential
func (r *Repository) GetAccessLogs(ctx context.Context, credentialID string, limit, offset int) ([]*AccessLog, error) {
	if credentialID == "" {
		return nil, ErrInvalidCredentialID
	}

	// Set default limit
	if limit <= 0 {
		limit = 50
	}
	if limit > 100 {
		limit = 100
	}

	query := `
		SELECT id, credential_id, tenant_id, accessed_by, access_type,
		       accessed_at, ip_address, user_agent, success, error_message
		FROM credential_access_log
		WHERE credential_id = $1
		ORDER BY accessed_at DESC
		LIMIT $2 OFFSET $3
	`

	var logs []*AccessLog
	err := r.db.SelectContext(ctx, &logs, query, credentialID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get access logs: %w", err)
	}

	if logs == nil {
		logs = []*AccessLog{}
	}

	return logs, nil
}

// joinUpdates joins SQL SET clauses with commas
func joinUpdates(updates []string) string {
	result := ""
	for i, update := range updates {
		if i > 0 {
			result += ", "
		}
		result += update
	}
	return result
}
