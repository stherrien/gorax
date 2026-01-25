package rbac

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

// Repository handles database operations for RBAC
type Repository struct {
	db *sqlx.DB
}

// NewRepository creates a new RBAC repository
func NewRepository(db *sqlx.DB) *Repository {
	return &Repository{db: db}
}

// CreateRole creates a new role
func (r *Repository) CreateRole(ctx context.Context, role *Role) error {
	query := `
		INSERT INTO roles (tenant_id, name, description, is_system)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at, updated_at
	`

	err := r.db.QueryRowContext(ctx, query,
		role.TenantID,
		role.Name,
		role.Description,
		role.IsSystem,
	).Scan(&role.ID, &role.CreatedAt, &role.UpdatedAt)

	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			return ErrRoleAlreadyExists
		}
		return fmt.Errorf("create role: %w", err)
	}

	return nil
}

// GetRoleByID retrieves a role by ID
func (r *Repository) GetRoleByID(ctx context.Context, roleID, tenantID string) (*Role, error) {
	var role Role
	query := `
		SELECT id, tenant_id, name, description, is_system, created_at, updated_at
		FROM roles
		WHERE id = $1 AND tenant_id = $2
	`

	err := r.db.GetContext(ctx, &role, query, roleID, tenantID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrRoleNotFound
		}
		return nil, fmt.Errorf("get role by id: %w", err)
	}

	return &role, nil
}

// GetRoleByName retrieves a role by name
func (r *Repository) GetRoleByName(ctx context.Context, name, tenantID string) (*Role, error) {
	var role Role
	query := `
		SELECT id, tenant_id, name, description, is_system, created_at, updated_at
		FROM roles
		WHERE name = $1 AND tenant_id = $2
	`

	err := r.db.GetContext(ctx, &role, query, name, tenantID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrRoleNotFound
		}
		return nil, fmt.Errorf("get role by name: %w", err)
	}

	return &role, nil
}

// ListRoles retrieves all roles for a tenant
func (r *Repository) ListRoles(ctx context.Context, tenantID string) ([]*Role, error) {
	var roles []*Role
	query := `
		SELECT id, tenant_id, name, description, is_system, created_at, updated_at
		FROM roles
		WHERE tenant_id = $1
		ORDER BY is_system DESC, name ASC
	`

	err := r.db.SelectContext(ctx, &roles, query, tenantID)
	if err != nil {
		return nil, fmt.Errorf("list roles: %w", err)
	}

	return roles, nil
}

// UpdateRole updates a role
func (r *Repository) UpdateRole(ctx context.Context, role *Role) error {
	query := `
		UPDATE roles
		SET name = $1, description = $2
		WHERE id = $3 AND tenant_id = $4
	`

	result, err := r.db.ExecContext(ctx, query,
		role.Name,
		role.Description,
		role.ID,
		role.TenantID,
	)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			return ErrRoleAlreadyExists
		}
		return fmt.Errorf("update role: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("get rows affected: %w", err)
	}

	if rows == 0 {
		return ErrRoleNotFound
	}

	return nil
}

// DeleteRole deletes a role
func (r *Repository) DeleteRole(ctx context.Context, roleID, tenantID string) error {
	query := `DELETE FROM roles WHERE id = $1 AND tenant_id = $2`

	result, err := r.db.ExecContext(ctx, query, roleID, tenantID)
	if err != nil {
		return fmt.Errorf("delete role: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("get rows affected: %w", err)
	}

	if rows == 0 {
		return ErrRoleNotFound
	}

	return nil
}

// GetRolePermissions retrieves all permissions for a role
func (r *Repository) GetRolePermissions(ctx context.Context, roleID string) ([]Permission, error) {
	var permissions []Permission
	query := `
		SELECT p.id, p.resource, p.action, p.description, p.created_at
		FROM permissions p
		INNER JOIN role_permissions rp ON rp.permission_id = p.id
		WHERE rp.role_id = $1
		ORDER BY p.resource, p.action
	`

	err := r.db.SelectContext(ctx, &permissions, query, roleID)
	if err != nil {
		return nil, fmt.Errorf("get role permissions: %w", err)
	}

	return permissions, nil
}

// SetRolePermissions sets the permissions for a role (replaces all existing)
func (r *Repository) SetRolePermissions(ctx context.Context, roleID string, permissionIDs []string) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback() //nolint:errcheck // rollback is no-op after commit

	// Delete existing permissions
	_, err = tx.ExecContext(ctx, `DELETE FROM role_permissions WHERE role_id = $1`, roleID)
	if err != nil {
		return fmt.Errorf("delete existing permissions: %w", err)
	}

	// Insert new permissions
	for _, permID := range permissionIDs {
		_, err = tx.ExecContext(ctx,
			`INSERT INTO role_permissions (role_id, permission_id) VALUES ($1, $2)`,
			roleID, permID,
		)
		if err != nil {
			return fmt.Errorf("insert permission: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}

// GetUserRoles retrieves all roles assigned to a user
func (r *Repository) GetUserRoles(ctx context.Context, userID, tenantID string) ([]*Role, error) {
	var roles []*Role
	query := `
		SELECT r.id, r.tenant_id, r.name, r.description, r.is_system, r.created_at, r.updated_at
		FROM roles r
		INNER JOIN user_roles ur ON ur.role_id = r.id
		WHERE ur.user_id = $1 AND r.tenant_id = $2
		ORDER BY r.name
	`

	err := r.db.SelectContext(ctx, &roles, query, userID, tenantID)
	if err != nil {
		return nil, fmt.Errorf("get user roles: %w", err)
	}

	return roles, nil
}

// AssignRolesToUser assigns roles to a user (replaces all existing)
func (r *Repository) AssignRolesToUser(ctx context.Context, userID string, roleIDs []string, grantedBy string) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback() //nolint:errcheck // rollback is no-op after commit

	// Delete existing roles
	_, err = tx.ExecContext(ctx, `DELETE FROM user_roles WHERE user_id = $1`, userID)
	if err != nil {
		return fmt.Errorf("delete existing roles: %w", err)
	}

	// Insert new roles
	for _, roleID := range roleIDs {
		_, err = tx.ExecContext(ctx,
			`INSERT INTO user_roles (user_id, role_id, granted_by) VALUES ($1, $2, $3)`,
			userID, roleID, grantedBy,
		)
		if err != nil {
			return fmt.Errorf("insert role: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}

// GetUserPermissions retrieves all permissions for a user across all their roles
func (r *Repository) GetUserPermissions(ctx context.Context, userID, tenantID string) ([]Permission, error) {
	var permissions []Permission
	query := `
		SELECT DISTINCT p.id, p.resource, p.action, p.description, p.created_at
		FROM permissions p
		INNER JOIN role_permissions rp ON rp.permission_id = p.id
		INNER JOIN user_roles ur ON ur.role_id = rp.role_id
		INNER JOIN roles r ON r.id = ur.role_id
		WHERE ur.user_id = $1 AND r.tenant_id = $2
		ORDER BY p.resource, p.action
	`

	err := r.db.SelectContext(ctx, &permissions, query, userID, tenantID)
	if err != nil {
		return nil, fmt.Errorf("get user permissions: %w", err)
	}

	return permissions, nil
}

// HasPermission checks if a user has a specific permission
func (r *Repository) HasPermission(ctx context.Context, userID, tenantID, resource, action string) (bool, error) {
	var count int
	query := `
		SELECT COUNT(*)
		FROM permissions p
		INNER JOIN role_permissions rp ON rp.permission_id = p.id
		INNER JOIN user_roles ur ON ur.role_id = rp.role_id
		INNER JOIN roles r ON r.id = ur.role_id
		WHERE ur.user_id = $1
		AND r.tenant_id = $2
		AND p.resource = $3
		AND p.action = $4
	`

	err := r.db.GetContext(ctx, &count, query, userID, tenantID, resource, action)
	if err != nil {
		return false, fmt.Errorf("check permission: %w", err)
	}

	return count > 0, nil
}

// ListPermissions retrieves all available permissions
func (r *Repository) ListPermissions(ctx context.Context) ([]Permission, error) {
	var permissions []Permission
	query := `
		SELECT id, resource, action, description, created_at
		FROM permissions
		ORDER BY resource, action
	`

	err := r.db.SelectContext(ctx, &permissions, query)
	if err != nil {
		return nil, fmt.Errorf("list permissions: %w", err)
	}

	return permissions, nil
}

// GetPermissionByResourceAction retrieves a permission by resource and action
func (r *Repository) GetPermissionByResourceAction(ctx context.Context, resource, action string) (*Permission, error) {
	var permission Permission
	query := `
		SELECT id, resource, action, description, created_at
		FROM permissions
		WHERE resource = $1 AND action = $2
	`

	err := r.db.GetContext(ctx, &permission, query, resource, action)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrPermissionNotFound
		}
		return nil, fmt.Errorf("get permission: %w", err)
	}

	return &permission, nil
}

// CreateAuditLog creates an audit log entry
func (r *Repository) CreateAuditLog(ctx context.Context, log *AuditLog) error {
	detailsJSON, err := json.Marshal(log.Details)
	if err != nil {
		return fmt.Errorf("marshal details: %w", err)
	}

	query := `
		INSERT INTO permission_audit_log (tenant_id, user_id, action, target_type, target_id, details)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	_, err = r.db.ExecContext(ctx, query,
		log.TenantID,
		log.UserID,
		log.Action,
		log.TargetType,
		log.TargetID,
		detailsJSON,
	)
	if err != nil {
		return fmt.Errorf("create audit log: %w", err)
	}

	return nil
}

// GetAuditLogs retrieves audit logs for a tenant
func (r *Repository) GetAuditLogs(ctx context.Context, tenantID string, limit, offset int) ([]*AuditLog, error) {
	var logs []*AuditLog
	query := `
		SELECT id, tenant_id, user_id, action, target_type, target_id, details, created_at
		FROM permission_audit_log
		WHERE tenant_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	err := r.db.SelectContext(ctx, &logs, query, tenantID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("get audit logs: %w", err)
	}

	return logs, nil
}
