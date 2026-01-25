package credential

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"

	"github.com/gorax/gorax/internal/metrics"
)

// Repository handles credential database operations
type Repository struct {
	db      *sqlx.DB
	metrics *metrics.Metrics
}

// NewRepository creates a new credential repository
func NewRepository(db *sqlx.DB) *Repository {
	return &Repository{
		db:      db,
		metrics: nil, // Metrics optional for backwards compatibility
	}
}

// NewRepositoryWithMetrics creates a new credential repository with metrics
func NewRepositoryWithMetrics(db *sqlx.DB, m *metrics.Metrics) *Repository {
	return &Repository{
		db:      db,
		metrics: m,
	}
}

// recordQuery records database query metrics
func (r *Repository) recordQuery(operation, table string, start time.Time, err error) {
	if r.metrics != nil {
		duration := time.Since(start).Seconds()
		status := "success"
		if err != nil {
			status = "error"
		}
		r.metrics.RecordDBQuery(operation, table, status, duration)
	}
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
	defer tx.Rollback() //nolint:errcheck // rollback is no-op after commit

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
	defer tx.Rollback() //nolint:errcheck // rollback is no-op after commit

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
	defer tx.Rollback() //nolint:errcheck // rollback is no-op after commit

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
	args := []any{}
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
	defer tx.Rollback() //nolint:errcheck // rollback is no-op after commit

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
	defer tx.Rollback() //nolint:errcheck // rollback is no-op after commit

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
	defer tx.Rollback() //nolint:errcheck // rollback is no-op after commit

	// Set tenant context for RLS within transaction using set_config
	_, err = tx.ExecContext(ctx, "SELECT set_config('app.current_tenant_id', $1, true)", tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to set tenant context: %w", err)
	}

	query := `SELECT * FROM credentials WHERE tenant_id = $1`
	args := []any{tenantID}
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
	defer tx.Rollback() //nolint:errcheck // rollback is no-op after commit

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
	defer tx.Rollback() //nolint:errcheck // rollback is no-op after commit

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
	if len(updates) == 0 {
		return ""
	}
	if len(updates) == 1 {
		return updates[0]
	}

	// Use strings.Join for efficiency
	var builder strings.Builder
	for i, update := range updates {
		if i > 0 {
			builder.WriteString(", ")
		}
		builder.WriteString(update)
	}
	return builder.String()
}

// ListWithPagination retrieves credentials with proper database-level pagination
func (r *Repository) ListWithPagination(ctx context.Context, tenantID string, filter CredentialListFilter, limit, offset int) ([]*Credential, int, error) {
	if tenantID == "" {
		return nil, 0, ErrInvalidTenantID
	}

	// Set default limit
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	// Start transaction for RLS context
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback() //nolint:errcheck // rollback is no-op after commit

	// Set tenant context for RLS within transaction using set_config
	_, err = tx.ExecContext(ctx, "SELECT set_config('app.current_tenant_id', $1, true)", tenantID)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to set tenant context: %w", err)
	}

	// Build base query and count query
	baseWhere := "WHERE tenant_id = $1"
	args := []any{tenantID}
	argIndex := 2

	// Apply type filter
	if filter.Type != "" {
		baseWhere += fmt.Sprintf(" AND type = $%d", argIndex)
		args = append(args, filter.Type)
		argIndex++
	}

	// Apply status filter
	if filter.Status != "" {
		baseWhere += fmt.Sprintf(" AND status = $%d", argIndex)
		args = append(args, filter.Status)
		argIndex++
	}

	// Apply search filter (searches in name and description)
	if filter.Search != "" {
		baseWhere += fmt.Sprintf(" AND (name ILIKE $%d OR description ILIKE $%d)", argIndex, argIndex)
		searchPattern := "%" + filter.Search + "%"
		args = append(args, searchPattern)
		argIndex++
	}

	// Get total count
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM credentials %s", baseWhere)
	var total int
	err = tx.GetContext(ctx, &total, countQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count credentials: %w", err)
	}

	// Get paginated results
	query := fmt.Sprintf(`SELECT * FROM credentials %s ORDER BY created_at DESC LIMIT $%d OFFSET $%d`,
		baseWhere, argIndex, argIndex+1)
	args = append(args, limit, offset)

	var credentials []*Credential
	err = tx.SelectContext(ctx, &credentials, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list credentials: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return nil, 0, fmt.Errorf("failed to commit transaction: %w", err)
	}

	if credentials == nil {
		credentials = []*Credential{}
	}

	return credentials, total, nil
}

// CredentialVersion represents a historical version of a credential
type CredentialVersion struct {
	ID             string     `json:"id" db:"id"`
	CredentialID   string     `json:"credential_id" db:"credential_id"`
	TenantID       string     `json:"tenant_id" db:"tenant_id"`
	Version        int        `json:"version" db:"version"`
	EncryptedDEK   []byte     `json:"-" db:"encrypted_dek"`
	Ciphertext     []byte     `json:"-" db:"ciphertext"`
	Nonce          []byte     `json:"-" db:"nonce"`
	AuthTag        []byte     `json:"-" db:"auth_tag"`
	KMSKeyID       string     `json:"-" db:"kms_key_id"`
	IsActive       bool       `json:"is_active" db:"is_active"`
	CreatedBy      string     `json:"created_by" db:"created_by"`
	CreatedAt      time.Time  `json:"created_at" db:"created_at"`
	DeactivatedAt  *time.Time `json:"deactivated_at,omitempty" db:"deactivated_at"`
	DeactivatedBy  *string    `json:"deactivated_by,omitempty" db:"deactivated_by"`
	RotationReason *string    `json:"rotation_reason,omitempty" db:"rotation_reason"`
}

// CreateVersion creates a new credential version during rotation
func (r *Repository) CreateVersion(ctx context.Context, tenantID string, cred *Credential, createdBy string, reason string) (*CredentialVersion, error) {
	if tenantID == "" {
		return nil, ErrInvalidTenantID
	}

	if cred == nil {
		return nil, ErrEmptyCredentialData
	}

	// Start transaction for RLS context
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback() //nolint:errcheck // rollback is no-op after commit

	// Set tenant context for RLS within transaction using set_config
	_, err = tx.ExecContext(ctx, "SELECT set_config('app.current_tenant_id', $1, true)", tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to set tenant context: %w", err)
	}

	// Get current max version for this credential
	var maxVersion int
	err = tx.GetContext(ctx, &maxVersion, `
		SELECT COALESCE(MAX(version), 0) FROM credential_versions
		WHERE credential_id = $1 AND tenant_id = $2
	`, cred.ID, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get max version: %w", err)
	}

	newVersion := maxVersion + 1
	now := time.Now()

	// Deactivate all existing versions for this credential
	_, err = tx.ExecContext(ctx, `
		UPDATE credential_versions
		SET is_active = false, deactivated_at = $1, deactivated_by = $2
		WHERE credential_id = $3 AND tenant_id = $4 AND is_active = true
	`, now, createdBy, cred.ID, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to deactivate old versions: %w", err)
	}

	// Create new version
	version := &CredentialVersion{
		ID:           uuid.NewString(),
		CredentialID: cred.ID,
		TenantID:     tenantID,
		Version:      newVersion,
		EncryptedDEK: cred.EncryptedDEK,
		Ciphertext:   cred.Ciphertext,
		Nonce:        cred.Nonce,
		AuthTag:      cred.AuthTag,
		KMSKeyID:     cred.KMSKeyID,
		IsActive:     true,
		CreatedBy:    createdBy,
		CreatedAt:    now,
	}

	if reason != "" {
		version.RotationReason = &reason
	}

	query := `
		INSERT INTO credential_versions (
			id, credential_id, tenant_id, version,
			encrypted_dek, ciphertext, nonce, auth_tag, kms_key_id,
			is_active, created_by, created_at, rotation_reason
		) VALUES (
			$1, $2, $3, $4,
			$5, $6, $7, $8, $9,
			$10, $11, $12, $13
		) RETURNING *
	`

	err = tx.QueryRowxContext(
		ctx, query,
		version.ID, version.CredentialID, version.TenantID, version.Version,
		version.EncryptedDEK, version.Ciphertext, version.Nonce, version.AuthTag, version.KMSKeyID,
		version.IsActive, version.CreatedBy, version.CreatedAt, version.RotationReason,
	).StructScan(version)
	if err != nil {
		return nil, fmt.Errorf("failed to create version: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return version, nil
}

// GetVersions retrieves all versions of a credential
func (r *Repository) GetVersions(ctx context.Context, tenantID, credentialID string) ([]*CredentialVersion, error) {
	if tenantID == "" {
		return nil, ErrInvalidTenantID
	}

	if credentialID == "" {
		return nil, ErrInvalidCredentialID
	}

	// Start transaction for RLS context
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback() //nolint:errcheck // rollback is no-op after commit

	// Set tenant context for RLS within transaction using set_config
	_, err = tx.ExecContext(ctx, "SELECT set_config('app.current_tenant_id', $1, true)", tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to set tenant context: %w", err)
	}

	query := `
		SELECT id, credential_id, tenant_id, version, is_active,
		       created_by, created_at, deactivated_at, deactivated_by, rotation_reason
		FROM credential_versions
		WHERE credential_id = $1 AND tenant_id = $2
		ORDER BY version DESC
	`

	var versions []*CredentialVersion
	err = tx.SelectContext(ctx, &versions, query, credentialID, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get versions: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	if versions == nil {
		versions = []*CredentialVersion{}
	}

	return versions, nil
}

// GetActiveVersion retrieves the current active version of a credential
func (r *Repository) GetActiveVersion(ctx context.Context, tenantID, credentialID string) (*CredentialVersion, error) {
	if tenantID == "" {
		return nil, ErrInvalidTenantID
	}

	if credentialID == "" {
		return nil, ErrInvalidCredentialID
	}

	// Start transaction for RLS context
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback() //nolint:errcheck // rollback is no-op after commit

	// Set tenant context for RLS within transaction using set_config
	_, err = tx.ExecContext(ctx, "SELECT set_config('app.current_tenant_id', $1, true)", tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to set tenant context: %w", err)
	}

	query := `
		SELECT * FROM credential_versions
		WHERE credential_id = $1 AND tenant_id = $2 AND is_active = true
		ORDER BY version DESC
		LIMIT 1
	`

	var version CredentialVersion
	err = tx.GetContext(ctx, &version, query, credentialID, tenantID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to get active version: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return &version, nil
}

// GetExpiredCredentials retrieves credentials that have expired or will expire within the given duration
func (r *Repository) GetExpiredCredentials(ctx context.Context, tenantID string, withinDuration time.Duration) ([]*Credential, error) {
	if tenantID == "" {
		return nil, ErrInvalidTenantID
	}

	// Start transaction for RLS context
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback() //nolint:errcheck // rollback is no-op after commit

	// Set tenant context for RLS within transaction using set_config
	_, err = tx.ExecContext(ctx, "SELECT set_config('app.current_tenant_id', $1, true)", tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to set tenant context: %w", err)
	}

	expirationThreshold := time.Now().Add(withinDuration)

	query := `
		SELECT * FROM credentials
		WHERE tenant_id = $1
		AND expires_at IS NOT NULL
		AND expires_at <= $2
		AND status = 'active'
		ORDER BY expires_at ASC
	`

	var credentials []*Credential
	err = tx.SelectContext(ctx, &credentials, query, tenantID, expirationThreshold)
	if err != nil {
		return nil, fmt.Errorf("failed to get expired credentials: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	if credentials == nil {
		credentials = []*Credential{}
	}

	return credentials, nil
}

// ValidateAndGet retrieves a credential by name after validation (implements RepositoryInterface for Injector)
func (r *Repository) ValidateAndGet(ctx context.Context, tenantID, name string) (*Credential, error) {
	cred, err := r.GetByName(ctx, tenantID, name)
	if err != nil {
		return nil, err
	}

	// Validate credential is active and not expired
	if cred.Status != StatusActive {
		return nil, fmt.Errorf("credential '%s' is not active (status: %s)", name, cred.Status)
	}

	if cred.IsExpired() {
		return nil, fmt.Errorf("credential '%s' has expired", name)
	}

	return cred, nil
}

// UpdateAccessTime updates the last_used_at timestamp (alias for UpdateLastUsedAt to satisfy RepositoryInterface)
func (r *Repository) UpdateAccessTime(ctx context.Context, tenantID, credentialID string) error {
	return r.UpdateLastUsedAt(ctx, tenantID, credentialID)
}

// RotateCredential performs a credential rotation with proper version tracking
func (r *Repository) RotateCredential(ctx context.Context, tenantID, credentialID, userID string, newCred *Credential, reason string) (*Credential, error) {
	if tenantID == "" {
		return nil, ErrInvalidTenantID
	}

	if credentialID == "" {
		return nil, ErrInvalidCredentialID
	}

	if newCred == nil {
		return nil, ErrEmptyCredentialData
	}

	// Start transaction for RLS context
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback() //nolint:errcheck // rollback is no-op after commit

	// Set tenant context for RLS within transaction using set_config
	_, err = tx.ExecContext(ctx, "SELECT set_config('app.current_tenant_id', $1, true)", tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to set tenant context: %w", err)
	}

	// Get existing credential to verify it exists
	var existing Credential
	err = tx.GetContext(ctx, &existing, `SELECT * FROM credentials WHERE id = $1 AND tenant_id = $2`, credentialID, tenantID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to get credential: %w", err)
	}

	// Archive current version before updating
	var maxVersion int
	err = tx.GetContext(ctx, &maxVersion, `
		SELECT COALESCE(MAX(version), 0) FROM credential_versions
		WHERE credential_id = $1 AND tenant_id = $2
	`, credentialID, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get max version: %w", err)
	}

	now := time.Now()
	newVersion := maxVersion + 1

	// Deactivate all existing versions
	_, err = tx.ExecContext(ctx, `
		UPDATE credential_versions
		SET is_active = false, deactivated_at = $1, deactivated_by = $2
		WHERE credential_id = $3 AND tenant_id = $4 AND is_active = true
	`, now, userID, credentialID, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to deactivate old versions: %w", err)
	}

	// Create new version entry
	versionID := uuid.NewString()
	var reasonPtr *string
	if reason != "" {
		reasonPtr = &reason
	}

	_, err = tx.ExecContext(ctx, `
		INSERT INTO credential_versions (
			id, credential_id, tenant_id, version,
			encrypted_dek, ciphertext, nonce, auth_tag, kms_key_id,
			is_active, created_by, created_at, rotation_reason
		) VALUES (
			$1, $2, $3, $4,
			$5, $6, $7, $8, $9,
			$10, $11, $12, $13
		)
	`, versionID, credentialID, tenantID, newVersion,
		newCred.EncryptedDEK, newCred.Ciphertext, newCred.Nonce, newCred.AuthTag, newCred.KMSKeyID,
		true, userID, now, reasonPtr)
	if err != nil {
		return nil, fmt.Errorf("failed to create version: %w", err)
	}

	// Update the main credential record with new encrypted values
	query := `
		UPDATE credentials
		SET encrypted_dek = $1, ciphertext = $2, nonce = $3, auth_tag = $4,
		    kms_key_id = $5, updated_at = $6
		WHERE id = $7 AND tenant_id = $8
		RETURNING *
	`

	var updated Credential
	err = tx.QueryRowxContext(ctx, query,
		newCred.EncryptedDEK, newCred.Ciphertext, newCred.Nonce, newCred.AuthTag,
		newCred.KMSKeyID, now,
		credentialID, tenantID,
	).StructScan(&updated)
	if err != nil {
		return nil, fmt.Errorf("failed to update credential: %w", err)
	}

	// Record rotation in credential_rotations table
	_, err = tx.ExecContext(ctx, `
		INSERT INTO credential_rotations (
			credential_id, previous_key_id, new_key_id, rotated_by, rotated_at, reason
		) VALUES ($1, $2, $3, $4, $5, $6)
	`, credentialID, existing.KMSKeyID, newCred.KMSKeyID, userID, now, reason)
	if err != nil {
		return nil, fmt.Errorf("failed to record rotation: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return &updated, nil
}
