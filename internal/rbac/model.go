package rbac

import (
	"time"
)

// Role represents a role in the system
type Role struct {
	ID          string       `json:"id" db:"id"`
	TenantID    string       `json:"tenant_id" db:"tenant_id"`
	Name        string       `json:"name" db:"name"`
	Description string       `json:"description" db:"description"`
	IsSystem    bool         `json:"is_system" db:"is_system"`
	CreatedAt   time.Time    `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at" db:"updated_at"`
	Permissions []Permission `json:"permissions,omitempty" db:"-"`
}

// Permission represents a permission in the system
type Permission struct {
	ID          string    `json:"id" db:"id"`
	Resource    string    `json:"resource" db:"resource"`
	Action      string    `json:"action" db:"action"`
	Description string    `json:"description" db:"description"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

// UserRole represents the assignment of a role to a user
type UserRole struct {
	UserID    string    `json:"user_id" db:"user_id"`
	RoleID    string    `json:"role_id" db:"role_id"`
	GrantedBy *string   `json:"granted_by" db:"granted_by"`
	GrantedAt time.Time `json:"granted_at" db:"granted_at"`
}

// AuditLog represents a permission-related audit log entry
type AuditLog struct {
	ID         string                 `json:"id" db:"id"`
	TenantID   string                 `json:"tenant_id" db:"tenant_id"`
	UserID     string                 `json:"user_id" db:"user_id"`
	Action     string                 `json:"action" db:"action"`
	TargetType *string                `json:"target_type" db:"target_type"`
	TargetID   *string                `json:"target_id" db:"target_id"`
	Details    map[string]interface{} `json:"details" db:"details"`
	CreatedAt  time.Time              `json:"created_at" db:"created_at"`
}

// Default role names
const (
	RoleAdmin    = "admin"
	RoleEditor   = "editor"
	RoleViewer   = "viewer"
	RoleOperator = "operator"
)

// Audit action types
const (
	AuditActionRoleCreated       = "role_created"
	AuditActionRoleUpdated       = "role_updated"
	AuditActionRoleDeleted       = "role_deleted"
	AuditActionPermissionGranted = "permission_granted"
	AuditActionPermissionRevoked = "permission_revoked"
	AuditActionUserRoleAssigned  = "user_role_assigned"
	AuditActionUserRoleRemoved   = "user_role_removed"
)

// CreateRoleRequest represents a request to create a role
type CreateRoleRequest struct {
	Name          string   `json:"name" validate:"required,min=1,max=100"`
	Description   string   `json:"description"`
	PermissionIDs []string `json:"permission_ids"`
}

// UpdateRoleRequest represents a request to update a role
type UpdateRoleRequest struct {
	Name          *string  `json:"name" validate:"omitempty,min=1,max=100"`
	Description   *string  `json:"description"`
	PermissionIDs []string `json:"permission_ids"`
}

// AssignRolesRequest represents a request to assign roles to a user
type AssignRolesRequest struct {
	RoleIDs []string `json:"role_ids" validate:"required,min=1"`
}

// Validate validates the CreateRoleRequest
func (r *CreateRoleRequest) Validate() error {
	if r.Name == "" {
		return ErrInvalidRoleName
	}
	return nil
}

// Validate validates the UpdateRoleRequest
func (r *UpdateRoleRequest) Validate() error {
	if r.Name != nil && *r.Name == "" {
		return ErrInvalidRoleName
	}
	return nil
}

// Validate validates the AssignRolesRequest
func (r *AssignRolesRequest) Validate() error {
	if len(r.RoleIDs) == 0 {
		return ErrNoRolesProvided
	}
	return nil
}
