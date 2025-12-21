# RBAC Implementation Guide - Phase 4.4

## Overview

This document provides implementation details for the Role-Based Access Control (RBAC) system in Gorax.

## Architecture

### Database Schema

The RBAC system uses 5 main tables:

1. **roles** - Role definitions
2. **permissions** - Available permissions (pre-seeded)
3. **role_permissions** - Many-to-many mapping
4. **user_roles** - User role assignments
5. **permission_audit_log** - Audit trail

See `migrations/014_rbac.sql` for complete schema.

### Backend Structure

```
internal/rbac/
├── model.go           # Domain models and constants
├── errors.go          # RBAC-specific errors
├── repository.go      # Data access layer
├── repository_test.go # Repository tests
├── service.go         # Business logic
├── service_test.go    # Service tests
├── middleware.go      # HTTP middleware
└── middleware_test.go # Middleware tests

internal/api/handlers/
└── rbac_handler.go    # HTTP handlers
```

### Frontend Structure

```
web/src/
├── api/rbac.ts                    # API client
├── hooks/useRoles.ts              # React hooks
└── pages/RoleManagement.tsx       # Admin UI
```

## Default Roles

Four system roles are created automatically for each tenant:

### Admin
- All permissions
- Cannot be deleted or modified
- Full system access

### Editor
- workflow: create, read, update, execute
- execution: read, cancel
- credential: create, read, update, delete
- webhook: create, read, update, delete

### Viewer
- workflow: read
- execution: read
- credential: read
- webhook: read
- user: read
- tenant: read

### Operator
- workflow: read, execute
- execution: read, cancel

## Permission Model

Permissions follow the pattern: `resource:action`

### Resources
- workflow
- execution
- credential
- webhook
- user
- tenant

### Actions
- create
- read
- update
- delete
- execute (workflow-specific)
- cancel (execution-specific)
- invite (user-specific)
- manage (user-specific, for role management)

## Usage Examples

### Backend Middleware

```go
import "gorax/internal/rbac"

// Require specific permission
r.With(rbac.RequirePermission(repo, "workflow", "create")).Post("/workflows", handler.Create)

// Require any of multiple permissions
r.With(rbac.RequireAnyPermission(repo,
    rbac.PermissionCheck{Resource: "workflow", Action: "create"},
    rbac.PermissionCheck{Resource: "workflow", Action: "update"},
)).Post("/workflows/:id/duplicate", handler.Duplicate)

// Require all permissions
r.With(rbac.RequireAllPermissions(repo,
    rbac.PermissionCheck{Resource: "workflow", Action: "read"},
    rbac.PermissionCheck{Resource: "execution", Action: "read"},
)).Get("/workflows/:id/executions", handler.ListExecutions)

// Require specific role
r.With(rbac.RequireRole(repo, rbac.RoleAdmin)).Get("/admin/settings", handler.Settings)
```

### Backend Service

```go
// Check permission programmatically
hasPermission, err := rbacService.CheckPermission(ctx, userID, tenantID, "workflow", "delete")
if !hasPermission {
    return ErrPermissionDenied
}

// Get user permissions
permissions, err := rbacService.GetUserPermissions(ctx, userID, tenantID)

// Assign roles to user
err := rbacService.AssignRolesToUser(ctx, userID, tenantID, grantedByUserID, &rbac.AssignRolesRequest{
    RoleIDs: []string{roleID1, roleID2},
})
```

### Frontend Hooks

```typescript
import { useHasPermission, useHasAnyPermission } from '../hooks/useRoles';

// Check single permission
const canCreate = useHasPermission('workflow', 'create');

// Check multiple permissions (OR)
const canEdit = useHasAnyPermission([
  { resource: 'workflow', action: 'create' },
  { resource: 'workflow', action: 'update' },
]);

// Conditional rendering
{canCreate && (
  <button onClick={handleCreate}>Create Workflow</button>
)}
```

### Frontend Role Management

```typescript
import { useRoles, useCreateRole, useUpdateRolePermissions } from '../hooks/useRoles';

function RoleManagement() {
  const { data: roles } = useRoles();
  const createRole = useCreateRole();
  const updatePermissions = useUpdateRolePermissions(roleId);

  const handleCreateRole = async () => {
    await createRole.mutateAsync({
      name: 'custom-role',
      description: 'Custom role description',
      permission_ids: selectedPermissionIds,
    });
  };

  const handleUpdatePermissions = async () => {
    await updatePermissions.mutateAsync(selectedPermissionIds);
  };
}
```

## API Endpoints

### Roles
- `GET /api/v1/roles` - List all roles
- `POST /api/v1/roles` - Create role (requires `user:manage`)
- `GET /api/v1/roles/:id` - Get role details
- `PUT /api/v1/roles/:id` - Update role (requires `user:manage`)
- `DELETE /api/v1/roles/:id` - Delete role (requires `user:manage`)

### Role Permissions
- `GET /api/v1/roles/:id/permissions` - Get role permissions
- `PUT /api/v1/roles/:id/permissions` - Update role permissions (requires `user:manage`)

### User Roles
- `GET /api/v1/users/:id/roles` - Get user's roles
- `PUT /api/v1/users/:id/roles` - Assign roles to user (requires `user:manage`)
- `GET /api/v1/users/:id/permissions` - Get user's effective permissions
- `GET /api/v1/me/permissions` - Get current user's permissions

### Permissions
- `GET /api/v1/permissions` - List all available permissions

### Audit Logs
- `GET /api/v1/audit-logs` - Get audit logs (requires `user:manage`)

## Integration Steps

### 1. Run Migration

```bash
# Apply the migration
make migrate-up
```

### 2. Update Tenant Service

Modify `internal/tenant/service.go` to create default roles on tenant creation:

```go
import "gorax/internal/rbac"

func (s *Service) CreateTenant(ctx context.Context, req *CreateTenantRequest) (*Tenant, error) {
    // ... existing tenant creation code ...

    // Create default RBAC roles
    if err := s.rbacService.CreateDefaultRoles(ctx, tenant.ID); err != nil {
        return nil, fmt.Errorf("create default roles: %w", err)
    }

    // Assign admin role to tenant owner
    adminRole, err := s.rbacRepo.GetRoleByName(ctx, rbac.RoleAdmin, tenant.ID)
    if err != nil {
        return nil, fmt.Errorf("get admin role: %w", err)
    }

    err = s.rbacService.AssignRolesToUser(ctx, ownerUserID, tenant.ID, ownerUserID, &rbac.AssignRolesRequest{
        RoleIDs: []string{adminRole.ID},
    })
    if err != nil {
        return nil, fmt.Errorf("assign admin role: %w", err)
    }

    return tenant, nil
}
```

### 3. Add Routes

Update `internal/api/app.go`:

```go
func (app *App) setupRoutes() {
    rbacHandler := handlers.NewRBACHandler(app.rbacService)

    // RBAC routes (require user:manage permission)
    app.router.Route("/api/v1/roles", func(r chi.Router) {
        r.With(rbac.RequirePermission(app.rbacRepo, "user", "manage")).Get("/", rbacHandler.ListRoles)
        r.With(rbac.RequirePermission(app.rbacRepo, "user", "manage")).Post("/", rbacHandler.CreateRole)
        r.With(rbac.RequirePermission(app.rbacRepo, "user", "manage")).Get("/{id}", rbacHandler.GetRole)
        r.With(rbac.RequirePermission(app.rbacRepo, "user", "manage")).Put("/{id}", rbacHandler.UpdateRole)
        r.With(rbac.RequirePermission(app.rbacRepo, "user", "manage")).Delete("/{id}", rbacHandler.DeleteRole)
        r.With(rbac.RequirePermission(app.rbacRepo, "user", "manage")).Get("/{id}/permissions", rbacHandler.GetRolePermissions)
        r.With(rbac.RequirePermission(app.rbacRepo, "user", "manage")).Put("/{id}/permissions", rbacHandler.UpdateRolePermissions)
    })

    app.router.Get("/api/v1/permissions", rbacHandler.ListPermissions)

    // Protected workflow routes
    app.router.Route("/api/v1/workflows", func(r chi.Router) {
        r.With(rbac.RequirePermission(app.rbacRepo, "workflow", "read")).Get("/", workflowHandler.List)
        r.With(rbac.RequirePermission(app.rbacRepo, "workflow", "create")).Post("/", workflowHandler.Create)
        r.With(rbac.RequirePermission(app.rbacRepo, "workflow", "read")).Get("/{id}", workflowHandler.Get)
        r.With(rbac.RequirePermission(app.rbacRepo, "workflow", "update")).Put("/{id}", workflowHandler.Update)
        r.With(rbac.RequirePermission(app.rbacRepo, "workflow", "delete")).Delete("/{id}", workflowHandler.Delete)
        r.With(rbac.RequirePermission(app.rbacRepo, "workflow", "execute")).Post("/{id}/execute", workflowHandler.Execute)
    })
}
```

### 4. Add Frontend Route

Update `web/src/App.tsx`:

```typescript
import RoleManagement from './pages/RoleManagement';

function App() {
  return (
    <Routes>
      {/* ... existing routes ... */}
      <Route path="/roles" element={<RoleManagement />} />
    </Routes>
  );
}
```

## Testing

### Unit Tests

```bash
# Backend tests
go test ./internal/rbac/... -v

# Frontend tests
cd web && npm test
```

### Integration Testing

1. Create a test tenant
2. Verify default roles are created
3. Create a test user with "viewer" role
4. Attempt to create a workflow (should fail with 403)
5. Update user to "editor" role
6. Attempt to create a workflow (should succeed)
7. Attempt to delete workflow (should fail - editor can't delete)

## Security Considerations

1. **System Roles**: Cannot be modified or deleted to prevent privilege escalation
2. **Permission Checks**: Always run on the backend; frontend checks are for UX only
3. **Audit Logging**: All permission changes are logged
4. **Tenant Isolation**: All queries include tenant_id to prevent cross-tenant access
5. **Context Requirements**: Middleware requires both user_id and tenant_id in context

## Performance Optimization

### Caching Recommendations

Consider caching user permissions in Redis:

```go
// Cache key: "user_permissions:{tenant_id}:{user_id}"
// TTL: 5 minutes
// Invalidate on: role assignment, permission changes
```

### Database Indexes

All required indexes are created in the migration:
- `roles(tenant_id, name)`
- `permissions(resource, action)`
- `user_roles(user_id)`
- `role_permissions(role_id)`

## Troubleshooting

### Permission Denied Errors

1. Check user has assigned roles: `GET /api/v1/users/:id/roles`
2. Check role permissions: `GET /api/v1/roles/:id/permissions`
3. Check audit logs: `GET /api/v1/audit-logs`

### System Role Not Found

Ensure default roles are created during tenant setup. Run manually if needed:

```go
err := rbacService.CreateDefaultRoles(ctx, tenantID)
```

### Frontend Permission Checks Not Working

Verify `getCurrentUserPermissions` is called and cached by React Query. Check browser network tab for API calls.

## Future Enhancements

1. **Resource-Level Permissions**: Add `resource_id` to permissions for fine-grained control
2. **Permission Groups**: Create permission bundles for common use cases
3. **Time-Based Permissions**: Add expiration dates to role assignments
4. **API Key Permissions**: Extend RBAC to API keys
5. **Permission Inheritance**: Support hierarchical roles

## Related Documentation

- [Database Schema](../migrations/014_rbac.sql)
- [API Documentation](./API.md)
- [Security Best Practices](./SECURITY.md)
