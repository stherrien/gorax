package rbac

import (
	"context"
	"fmt"
)

// RepositoryInterface defines the contract for RBAC repository
type RepositoryInterface interface {
	CreateRole(ctx context.Context, role *Role) error
	GetRoleByID(ctx context.Context, roleID, tenantID string) (*Role, error)
	GetRoleByName(ctx context.Context, name, tenantID string) (*Role, error)
	ListRoles(ctx context.Context, tenantID string) ([]*Role, error)
	UpdateRole(ctx context.Context, role *Role) error
	DeleteRole(ctx context.Context, roleID, tenantID string) error
	GetRolePermissions(ctx context.Context, roleID string) ([]Permission, error)
	SetRolePermissions(ctx context.Context, roleID string, permissionIDs []string) error
	GetUserRoles(ctx context.Context, userID, tenantID string) ([]*Role, error)
	AssignRolesToUser(ctx context.Context, userID string, roleIDs []string, grantedBy string) error
	GetUserPermissions(ctx context.Context, userID, tenantID string) ([]Permission, error)
	HasPermission(ctx context.Context, userID, tenantID, resource, action string) (bool, error)
	ListPermissions(ctx context.Context) ([]Permission, error)
	GetPermissionByResourceAction(ctx context.Context, resource, action string) (*Permission, error)
	CreateAuditLog(ctx context.Context, log *AuditLog) error
	GetAuditLogs(ctx context.Context, tenantID string, limit, offset int) ([]*AuditLog, error)
}

// Service handles business logic for RBAC
type Service struct {
	repo RepositoryInterface
}

// NewService creates a new RBAC service
func NewService(repo RepositoryInterface) *Service {
	return &Service{
		repo: repo,
	}
}

// CreateRole creates a new role
func (s *Service) CreateRole(ctx context.Context, tenantID, userID string, req *CreateRoleRequest) (*Role, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	role := &Role{
		TenantID:    tenantID,
		Name:        req.Name,
		Description: req.Description,
		IsSystem:    false,
	}

	if err := s.repo.CreateRole(ctx, role); err != nil {
		return nil, fmt.Errorf("create role: %w", err)
	}

	// Set permissions if provided
	if len(req.PermissionIDs) > 0 {
		if err := s.repo.SetRolePermissions(ctx, role.ID, req.PermissionIDs); err != nil {
			return nil, fmt.Errorf("set role permissions: %w", err)
		}
	}

	// Audit log
	s.auditLog(ctx, tenantID, userID, AuditActionRoleCreated, "role", role.ID, map[string]interface{}{
		"role_name":       role.Name,
		"permission_count": len(req.PermissionIDs),
	})

	return role, nil
}

// GetRole retrieves a role by ID with its permissions
func (s *Service) GetRole(ctx context.Context, roleID, tenantID string) (*Role, error) {
	role, err := s.repo.GetRoleByID(ctx, roleID, tenantID)
	if err != nil {
		return nil, fmt.Errorf("get role: %w", err)
	}

	permissions, err := s.repo.GetRolePermissions(ctx, roleID)
	if err != nil {
		return nil, fmt.Errorf("get role permissions: %w", err)
	}

	role.Permissions = permissions

	return role, nil
}

// ListRoles retrieves all roles for a tenant
func (s *Service) ListRoles(ctx context.Context, tenantID string) ([]*Role, error) {
	roles, err := s.repo.ListRoles(ctx, tenantID)
	if err != nil {
		return nil, fmt.Errorf("list roles: %w", err)
	}

	// Fetch permissions for each role
	for _, role := range roles {
		permissions, err := s.repo.GetRolePermissions(ctx, role.ID)
		if err != nil {
			return nil, fmt.Errorf("get role permissions: %w", err)
		}
		role.Permissions = permissions
	}

	return roles, nil
}

// UpdateRole updates a role
func (s *Service) UpdateRole(ctx context.Context, roleID, tenantID, userID string, req *UpdateRoleRequest) error {
	if err := req.Validate(); err != nil {
		return err
	}

	role, err := s.repo.GetRoleByID(ctx, roleID, tenantID)
	if err != nil {
		return fmt.Errorf("get role: %w", err)
	}

	if role.IsSystem {
		return ErrSystemRoleCannotBeModified
	}

	// Update role fields
	if req.Name != nil {
		role.Name = *req.Name
	}
	if req.Description != nil {
		role.Description = *req.Description
	}

	if err := s.repo.UpdateRole(ctx, role); err != nil {
		return fmt.Errorf("update role: %w", err)
	}

	// Update permissions if provided
	if req.PermissionIDs != nil {
		if err := s.repo.SetRolePermissions(ctx, roleID, req.PermissionIDs); err != nil {
			return fmt.Errorf("set role permissions: %w", err)
		}
	}

	// Audit log
	s.auditLog(ctx, tenantID, userID, AuditActionRoleUpdated, "role", roleID, map[string]interface{}{
		"role_name": role.Name,
	})

	return nil
}

// DeleteRole deletes a role
func (s *Service) DeleteRole(ctx context.Context, roleID, tenantID, userID string) error {
	role, err := s.repo.GetRoleByID(ctx, roleID, tenantID)
	if err != nil {
		return fmt.Errorf("get role: %w", err)
	}

	if role.IsSystem {
		return ErrSystemRoleCannotBeDeleted
	}

	if err := s.repo.DeleteRole(ctx, roleID, tenantID); err != nil {
		return fmt.Errorf("delete role: %w", err)
	}

	// Audit log
	s.auditLog(ctx, tenantID, userID, AuditActionRoleDeleted, "role", roleID, map[string]interface{}{
		"role_name": role.Name,
	})

	return nil
}

// GetRolePermissions retrieves permissions for a role
func (s *Service) GetRolePermissions(ctx context.Context, roleID, tenantID string) ([]Permission, error) {
	// Verify role exists and belongs to tenant
	_, err := s.repo.GetRoleByID(ctx, roleID, tenantID)
	if err != nil {
		return nil, fmt.Errorf("get role: %w", err)
	}

	permissions, err := s.repo.GetRolePermissions(ctx, roleID)
	if err != nil {
		return nil, fmt.Errorf("get role permissions: %w", err)
	}

	return permissions, nil
}

// UpdateRolePermissions updates permissions for a role
func (s *Service) UpdateRolePermissions(ctx context.Context, roleID, tenantID, userID string, permissionIDs []string) error {
	role, err := s.repo.GetRoleByID(ctx, roleID, tenantID)
	if err != nil {
		return fmt.Errorf("get role: %w", err)
	}

	if role.IsSystem {
		return ErrSystemRoleCannotBeModified
	}

	if err := s.repo.SetRolePermissions(ctx, roleID, permissionIDs); err != nil {
		return fmt.Errorf("set role permissions: %w", err)
	}

	// Audit log
	s.auditLog(ctx, tenantID, userID, AuditActionPermissionGranted, "role", roleID, map[string]interface{}{
		"role_name":       role.Name,
		"permission_count": len(permissionIDs),
	})

	return nil
}

// GetUserRoles retrieves roles assigned to a user
func (s *Service) GetUserRoles(ctx context.Context, userID, tenantID string) ([]*Role, error) {
	roles, err := s.repo.GetUserRoles(ctx, userID, tenantID)
	if err != nil {
		return nil, fmt.Errorf("get user roles: %w", err)
	}

	return roles, nil
}

// AssignRolesToUser assigns roles to a user
func (s *Service) AssignRolesToUser(ctx context.Context, userID, tenantID, grantedBy string, req *AssignRolesRequest) error {
	if err := req.Validate(); err != nil {
		return err
	}

	// Verify all roles exist and belong to tenant
	for _, roleID := range req.RoleIDs {
		_, err := s.repo.GetRoleByID(ctx, roleID, tenantID)
		if err != nil {
			return fmt.Errorf("get role %s: %w", roleID, err)
		}
	}

	if err := s.repo.AssignRolesToUser(ctx, userID, req.RoleIDs, grantedBy); err != nil {
		return fmt.Errorf("assign roles to user: %w", err)
	}

	// Audit log
	s.auditLog(ctx, tenantID, grantedBy, AuditActionUserRoleAssigned, "user", userID, map[string]interface{}{
		"role_count": len(req.RoleIDs),
		"role_ids":   req.RoleIDs,
	})

	return nil
}

// GetUserPermissions retrieves all permissions for a user
func (s *Service) GetUserPermissions(ctx context.Context, userID, tenantID string) ([]Permission, error) {
	permissions, err := s.repo.GetUserPermissions(ctx, userID, tenantID)
	if err != nil {
		return nil, fmt.Errorf("get user permissions: %w", err)
	}

	return permissions, nil
}

// CheckPermission checks if a user has a specific permission
func (s *Service) CheckPermission(ctx context.Context, userID, tenantID, resource, action string) (bool, error) {
	has, err := s.repo.HasPermission(ctx, userID, tenantID, resource, action)
	if err != nil {
		return false, fmt.Errorf("check permission: %w", err)
	}

	return has, nil
}

// ListPermissions retrieves all available permissions
func (s *Service) ListPermissions(ctx context.Context) ([]Permission, error) {
	permissions, err := s.repo.ListPermissions(ctx)
	if err != nil {
		return nil, fmt.Errorf("list permissions: %w", err)
	}

	return permissions, nil
}

// CreateDefaultRoles creates the default roles for a tenant
func (s *Service) CreateDefaultRoles(ctx context.Context, tenantID string) error {
	allPermissions, err := s.repo.ListPermissions(ctx)
	if err != nil {
		return fmt.Errorf("list permissions: %w", err)
	}

	// Create permission maps for easy filtering
	permMap := make(map[string]string) // "resource:action" -> permissionID
	for _, perm := range allPermissions {
		key := fmt.Sprintf("%s:%s", perm.Resource, perm.Action)
		permMap[key] = perm.ID
	}

	// Helper function to get permission IDs
	getPermIDs := func(patterns []string) []string {
		var ids []string
		for _, pattern := range patterns {
			if id, ok := permMap[pattern]; ok {
				ids = append(ids, id)
			}
		}
		return ids
	}

	// Admin role - all permissions
	adminRole := &Role{
		TenantID:    tenantID,
		Name:        RoleAdmin,
		Description: "Full access to all resources",
		IsSystem:    true,
	}
	if err := s.repo.CreateRole(ctx, adminRole); err != nil && err != ErrRoleAlreadyExists {
		return fmt.Errorf("create admin role: %w", err)
	}
	var allPermIDs []string
	for _, perm := range allPermissions {
		allPermIDs = append(allPermIDs, perm.ID)
	}
	if err := s.repo.SetRolePermissions(ctx, adminRole.ID, allPermIDs); err != nil {
		return fmt.Errorf("set admin permissions: %w", err)
	}

	// Editor role - can create/edit workflows and credentials
	editorRole := &Role{
		TenantID:    tenantID,
		Name:        RoleEditor,
		Description: "Create and edit workflows, credentials, and webhooks",
		IsSystem:    true,
	}
	if err := s.repo.CreateRole(ctx, editorRole); err != nil && err != ErrRoleAlreadyExists {
		return fmt.Errorf("create editor role: %w", err)
	}
	editorPerms := getPermIDs([]string{
		"workflow:create", "workflow:read", "workflow:update", "workflow:execute",
		"execution:read", "execution:cancel",
		"credential:create", "credential:read", "credential:update", "credential:delete",
		"webhook:create", "webhook:read", "webhook:update", "webhook:delete",
	})
	if err := s.repo.SetRolePermissions(ctx, editorRole.ID, editorPerms); err != nil {
		return fmt.Errorf("set editor permissions: %w", err)
	}

	// Viewer role - read-only access
	viewerRole := &Role{
		TenantID:    tenantID,
		Name:        RoleViewer,
		Description: "Read-only access to workflows and executions",
		IsSystem:    true,
	}
	if err := s.repo.CreateRole(ctx, viewerRole); err != nil && err != ErrRoleAlreadyExists {
		return fmt.Errorf("create viewer role: %w", err)
	}
	viewerPerms := getPermIDs([]string{
		"workflow:read",
		"execution:read",
		"credential:read",
		"webhook:read",
		"user:read",
		"tenant:read",
	})
	if err := s.repo.SetRolePermissions(ctx, viewerRole.ID, viewerPerms); err != nil {
		return fmt.Errorf("set viewer permissions: %w", err)
	}

	// Operator role - execute workflows and view results
	operatorRole := &Role{
		TenantID:    tenantID,
		Name:        RoleOperator,
		Description: "Execute workflows and manage executions",
		IsSystem:    true,
	}
	if err := s.repo.CreateRole(ctx, operatorRole); err != nil && err != ErrRoleAlreadyExists {
		return fmt.Errorf("create operator role: %w", err)
	}
	operatorPerms := getPermIDs([]string{
		"workflow:read", "workflow:execute",
		"execution:read", "execution:cancel",
	})
	if err := s.repo.SetRolePermissions(ctx, operatorRole.ID, operatorPerms); err != nil {
		return fmt.Errorf("set operator permissions: %w", err)
	}

	return nil
}

// GetAuditLogs retrieves audit logs for a tenant
func (s *Service) GetAuditLogs(ctx context.Context, tenantID string, limit, offset int) ([]*AuditLog, error) {
	logs, err := s.repo.GetAuditLogs(ctx, tenantID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("get audit logs: %w", err)
	}

	return logs, nil
}

// auditLog creates an audit log entry (non-blocking)
func (s *Service) auditLog(ctx context.Context, tenantID, userID, action, targetType, targetID string, details map[string]interface{}) {
	log := &AuditLog{
		TenantID:   tenantID,
		UserID:     userID,
		Action:     action,
		TargetType: &targetType,
		TargetID:   &targetID,
		Details:    details,
	}

	// Best effort - don't fail if audit logging fails
	_ = s.repo.CreateAuditLog(ctx, log)
}
