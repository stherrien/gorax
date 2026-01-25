package suites

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gorax/gorax/tests/integration"
)

// Role represents the role model for tests
type Role struct {
	ID          string       `json:"id"`
	TenantID    string       `json:"tenant_id"`
	Name        string       `json:"name"`
	Description string       `json:"description"`
	IsSystem    bool         `json:"is_system"`
	CreatedAt   string       `json:"created_at"`
	UpdatedAt   string       `json:"updated_at"`
	Permissions []Permission `json:"permissions,omitempty"`
}

// Permission represents a permission for tests
type Permission struct {
	ID          string `json:"id"`
	Resource    string `json:"resource"`
	Action      string `json:"action"`
	Description string `json:"description"`
}

// AuditLog represents an RBAC audit log entry
type AuditLog struct {
	ID         string         `json:"id"`
	TenantID   string         `json:"tenant_id"`
	UserID     string         `json:"user_id"`
	Action     string         `json:"action"`
	TargetType *string        `json:"target_type"`
	TargetID   *string        `json:"target_id"`
	Details    map[string]any `json:"details"`
	CreatedAt  string         `json:"created_at"`
}

// TestRBACVerification_CreateRoleAndAssignPermissions tests the full RBAC lifecycle
func TestRBACVerification_CreateRoleAndAssignPermissions(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ts := integration.SetupTestServer(t)

	// Create test tenant
	tenantID := ts.CreateTestTenant(t, "rbac-test-tenant")
	headers := integration.DefaultTestHeaders(tenantID)

	// Step 1: List all available permissions
	t.Run("Step 1: List all available permissions", func(t *testing.T) {
		resp := ts.MakeRequest(t, http.MethodGet, "/api/v1/permissions", nil, headers)
		integration.AssertStatusCode(t, resp, http.StatusOK)

		var permissions []Permission
		integration.ParseJSONResponse(t, resp, &permissions)
		assert.NotEmpty(t, permissions, "Should have at least some permissions")
	})

	// Step 2: List default roles
	t.Run("Step 2: List default roles", func(t *testing.T) {
		resp := ts.MakeRequest(t, http.MethodGet, "/api/v1/roles", nil, headers)
		integration.AssertStatusCode(t, resp, http.StatusOK)

		var roles []Role
		integration.ParseJSONResponse(t, resp, &roles)
		assert.GreaterOrEqual(t, len(roles), 4, "Should have at least 4 default roles")

		// Verify default roles exist
		roleNames := make(map[string]bool)
		for _, role := range roles {
			roleNames[role.Name] = true
		}
		assert.True(t, roleNames["admin"], "Should have admin role")
		assert.True(t, roleNames["editor"], "Should have editor role")
		assert.True(t, roleNames["viewer"], "Should have viewer role")
		assert.True(t, roleNames["operator"], "Should have operator role")
	})

	// Step 3: Create custom role
	var customRoleID string
	t.Run("Step 3: Create custom role with specific permissions", func(t *testing.T) {
		// First get permission IDs
		resp := ts.MakeRequest(t, http.MethodGet, "/api/v1/permissions", nil, headers)
		require.Equal(t, http.StatusOK, resp.StatusCode)

		var permissions []Permission
		integration.ParseJSONResponse(t, resp, &permissions)

		// Find workflow:read and execution:read permissions
		var permIDs []string
		for _, perm := range permissions {
			if (perm.Resource == "workflow" && perm.Action == "read") ||
				(perm.Resource == "execution" && perm.Action == "read") {
				permIDs = append(permIDs, perm.ID)
			}
		}
		require.GreaterOrEqual(t, len(permIDs), 1, "Should find at least one read permission")

		// Create custom role
		createReq := map[string]any{
			"name":           "Workflow Monitor",
			"description":    "Can only read workflows and executions",
			"permission_ids": permIDs,
		}

		resp = ts.MakeRequest(t, http.MethodPost, "/api/v1/roles", createReq, headers)
		integration.AssertStatusCode(t, resp, http.StatusCreated)

		var role Role
		integration.ParseJSONResponse(t, resp, &role)
		assert.Equal(t, "Workflow Monitor", role.Name)
		assert.False(t, role.IsSystem, "Custom role should not be a system role")
		customRoleID = role.ID
	})

	// Step 4: Get custom role details
	t.Run("Step 4: Get custom role details", func(t *testing.T) {
		resp := ts.MakeRequest(t, http.MethodGet, "/api/v1/roles/"+customRoleID, nil, headers)
		integration.AssertStatusCode(t, resp, http.StatusOK)

		var role Role
		integration.ParseJSONResponse(t, resp, &role)
		assert.Equal(t, customRoleID, role.ID)
		assert.Equal(t, "Workflow Monitor", role.Name)
	})

	// Step 5: Create test user and assign role
	userID := ts.CreateTestUser(t, tenantID, "user@test.com", "user")

	t.Run("Step 5: Assign custom role to user", func(t *testing.T) {
		assignReq := map[string]any{
			"role_ids": []string{customRoleID},
		}

		resp := ts.MakeRequest(t, http.MethodPut, "/api/v1/users/"+userID+"/roles", assignReq, headers)
		integration.AssertStatusCode(t, resp, http.StatusNoContent)
	})

	// Step 6: Verify user has the assigned role
	t.Run("Step 6: Verify user has the assigned role", func(t *testing.T) {
		resp := ts.MakeRequest(t, http.MethodGet, "/api/v1/users/"+userID+"/roles", nil, headers)
		integration.AssertStatusCode(t, resp, http.StatusOK)

		var roles []Role
		integration.ParseJSONResponse(t, resp, &roles)
		assert.GreaterOrEqual(t, len(roles), 1, "User should have at least one role")

		foundRole := false
		for _, role := range roles {
			if role.ID == customRoleID {
				foundRole = true
				break
			}
		}
		assert.True(t, foundRole, "User should have the custom role")
	})

	// Step 7: Verify user's effective permissions
	t.Run("Step 7: Verify user's effective permissions", func(t *testing.T) {
		resp := ts.MakeRequest(t, http.MethodGet, "/api/v1/users/"+userID+"/permissions", nil, headers)
		integration.AssertStatusCode(t, resp, http.StatusOK)

		var permissions []Permission
		integration.ParseJSONResponse(t, resp, &permissions)

		// User should have workflow:read or execution:read
		hasReadPermission := false
		for _, perm := range permissions {
			if perm.Action == "read" {
				hasReadPermission = true
				break
			}
		}
		assert.True(t, hasReadPermission, "User should have read permissions")
	})

	// Step 8: Update custom role
	t.Run("Step 8: Update custom role", func(t *testing.T) {
		newName := "Workflow Monitor v2"
		updateReq := map[string]any{
			"name":        newName,
			"description": "Updated description",
		}

		resp := ts.MakeRequest(t, http.MethodPut, "/api/v1/roles/"+customRoleID, updateReq, headers)
		integration.AssertStatusCode(t, resp, http.StatusNoContent)

		// Verify update
		resp = ts.MakeRequest(t, http.MethodGet, "/api/v1/roles/"+customRoleID, nil, headers)
		integration.AssertStatusCode(t, resp, http.StatusOK)

		var role Role
		integration.ParseJSONResponse(t, resp, &role)
		assert.Equal(t, newName, role.Name)
	})

	// Step 9: Verify audit logs capture changes
	t.Run("Step 9: Verify audit logs capture changes", func(t *testing.T) {
		resp := ts.MakeRequest(t, http.MethodGet, "/api/v1/audit-logs?limit=50", nil, headers)
		integration.AssertStatusCode(t, resp, http.StatusOK)

		var logs []AuditLog
		integration.ParseJSONResponse(t, resp, &logs)

		// Should have logs for role creation, role assignment, and role update
		actionCounts := make(map[string]int)
		for _, log := range logs {
			actionCounts[log.Action]++
		}
		assert.GreaterOrEqual(t, actionCounts["role_created"], 1, "Should have role_created audit log")
	})

	// Step 10: Delete custom role
	t.Run("Step 10: Delete custom role", func(t *testing.T) {
		resp := ts.MakeRequest(t, http.MethodDelete, "/api/v1/roles/"+customRoleID, nil, headers)
		integration.AssertStatusCode(t, resp, http.StatusNoContent)

		// Verify deletion
		resp = ts.MakeRequest(t, http.MethodGet, "/api/v1/roles/"+customRoleID, nil, headers)
		integration.AssertStatusCode(t, resp, http.StatusNotFound)
	})
}

// TestRBACVerification_SystemRoleProtection tests that system roles cannot be modified
func TestRBACVerification_SystemRoleProtection(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ts := integration.SetupTestServer(t)

	tenantID := ts.CreateTestTenant(t, "rbac-system-test")
	headers := integration.DefaultTestHeaders(tenantID)

	// Get admin role ID
	var adminRoleID string
	t.Run("Get system admin role", func(t *testing.T) {
		resp := ts.MakeRequest(t, http.MethodGet, "/api/v1/roles", nil, headers)
		integration.AssertStatusCode(t, resp, http.StatusOK)

		var roles []Role
		integration.ParseJSONResponse(t, resp, &roles)

		for _, role := range roles {
			if role.Name == "admin" && role.IsSystem {
				adminRoleID = role.ID
				break
			}
		}
		require.NotEmpty(t, adminRoleID, "Should find admin role")
	})

	t.Run("Cannot update system role", func(t *testing.T) {
		updateReq := map[string]any{
			"name":        "Modified Admin",
			"description": "Trying to modify system role",
		}

		resp := ts.MakeRequest(t, http.MethodPut, "/api/v1/roles/"+adminRoleID, updateReq, headers)
		integration.AssertStatusCode(t, resp, http.StatusForbidden)
	})

	t.Run("Cannot delete system role", func(t *testing.T) {
		resp := ts.MakeRequest(t, http.MethodDelete, "/api/v1/roles/"+adminRoleID, nil, headers)
		integration.AssertStatusCode(t, resp, http.StatusForbidden)
	})

	t.Run("Cannot modify system role permissions", func(t *testing.T) {
		permReq := map[string]any{
			"permission_ids": []string{}, // Try to remove all permissions
		}

		resp := ts.MakeRequest(t, http.MethodPut, "/api/v1/roles/"+adminRoleID+"/permissions", permReq, headers)
		integration.AssertStatusCode(t, resp, http.StatusForbidden)
	})
}

// TestRBACVerification_RoleValidation tests validation of role creation
func TestRBACVerification_RoleValidation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ts := integration.SetupTestServer(t)

	tenantID := ts.CreateTestTenant(t, "rbac-validation-test")
	userID := ts.CreateTestUser(t, tenantID, "admin@test.com", "admin")
	headers := integration.DefaultTestHeaders(tenantID)

	t.Run("Cannot create role with empty name", func(t *testing.T) {
		createReq := map[string]any{
			"name":        "",
			"description": "Test role",
		}

		resp := ts.MakeRequest(t, http.MethodPost, "/api/v1/roles", createReq, headers)
		integration.AssertStatusCode(t, resp, http.StatusBadRequest)
	})

	t.Run("Cannot create duplicate role name", func(t *testing.T) {
		// Create first role
		createReq := map[string]any{
			"name":        "Unique Role",
			"description": "First role",
		}

		resp := ts.MakeRequest(t, http.MethodPost, "/api/v1/roles", createReq, headers)
		require.Equal(t, http.StatusCreated, resp.StatusCode)
		resp.Body.Close()

		// Try to create duplicate
		resp = ts.MakeRequest(t, http.MethodPost, "/api/v1/roles", createReq, headers)
		integration.AssertStatusCode(t, resp, http.StatusConflict)
	})

	t.Run("Cannot assign empty role list", func(t *testing.T) {
		assignReq := map[string]any{
			"role_ids": []string{},
		}

		resp := ts.MakeRequest(t, http.MethodPut, "/api/v1/users/"+userID+"/roles", assignReq, headers)
		integration.AssertStatusCode(t, resp, http.StatusBadRequest)
	})

	t.Run("Cannot assign non-existent role", func(t *testing.T) {
		assignReq := map[string]any{
			"role_ids": []string{"00000000-0000-0000-0000-000000000000"},
		}

		resp := ts.MakeRequest(t, http.MethodPut, "/api/v1/users/"+userID+"/roles", assignReq, headers)
		integration.AssertStatusCode(t, resp, http.StatusNotFound)
	})
}

// TestRBACVerification_TenantIsolation tests that roles are isolated between tenants
func TestRBACVerification_TenantIsolation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ts := integration.SetupTestServer(t)

	// Create two tenants
	tenant1ID := ts.CreateTestTenant(t, "rbac-tenant-1")
	tenant2ID := ts.CreateTestTenant(t, "rbac-tenant-2")
	headers1 := integration.DefaultTestHeaders(tenant1ID)
	headers2 := integration.DefaultTestHeaders(tenant2ID)

	// Create a role in tenant 1
	var tenant1RoleID string
	t.Run("Create role in tenant 1", func(t *testing.T) {
		createReq := map[string]any{
			"name":        "Tenant 1 Role",
			"description": "Role for tenant 1 only",
		}

		resp := ts.MakeRequest(t, http.MethodPost, "/api/v1/roles", createReq, headers1)
		integration.AssertStatusCode(t, resp, http.StatusCreated)

		var role Role
		integration.ParseJSONResponse(t, resp, &role)
		tenant1RoleID = role.ID
	})

	t.Run("Tenant 1 can see its own role", func(t *testing.T) {
		resp := ts.MakeRequest(t, http.MethodGet, "/api/v1/roles/"+tenant1RoleID, nil, headers1)
		integration.AssertStatusCode(t, resp, http.StatusOK)
	})

	t.Run("Tenant 2 cannot see tenant 1's role", func(t *testing.T) {
		resp := ts.MakeRequest(t, http.MethodGet, "/api/v1/roles/"+tenant1RoleID, nil, headers2)
		integration.AssertStatusCode(t, resp, http.StatusNotFound)
	})

	t.Run("Tenant 2 role list does not include tenant 1's role", func(t *testing.T) {
		resp := ts.MakeRequest(t, http.MethodGet, "/api/v1/roles", nil, headers2)
		integration.AssertStatusCode(t, resp, http.StatusOK)

		var roles []Role
		integration.ParseJSONResponse(t, resp, &roles)

		for _, role := range roles {
			assert.NotEqual(t, tenant1RoleID, role.ID, "Tenant 2 should not see tenant 1's role")
			assert.NotEqual(t, "Tenant 1 Role", role.Name, "Tenant 2 should not see tenant 1's role by name")
		}
	})
}

// TestRBACVerification_CurrentUserPermissions tests the /me/permissions endpoint
func TestRBACVerification_CurrentUserPermissions(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ts := integration.SetupTestServer(t)

	tenantID := ts.CreateTestTenant(t, "rbac-me-test")
	headers := integration.DefaultTestHeaders(tenantID)

	t.Run("Get current user permissions", func(t *testing.T) {
		resp := ts.MakeRequest(t, http.MethodGet, "/api/v1/me/permissions", nil, headers)
		integration.AssertStatusCode(t, resp, http.StatusOK)

		var permissions []Permission
		integration.ParseJSONResponse(t, resp, &permissions)
		// Just verify the endpoint works - permissions may vary
		assert.NotNil(t, permissions)
	})
}

// TestRBACVerification_RolePermissionsManagement tests permission management for roles
func TestRBACVerification_RolePermissionsManagement(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ts := integration.SetupTestServer(t)

	tenantID := ts.CreateTestTenant(t, "rbac-perms-test")
	headers := integration.DefaultTestHeaders(tenantID)

	// Create a custom role
	var roleID string
	t.Run("Create custom role", func(t *testing.T) {
		createReq := map[string]any{
			"name":        "Permission Test Role",
			"description": "Role for permission testing",
		}

		resp := ts.MakeRequest(t, http.MethodPost, "/api/v1/roles", createReq, headers)
		integration.AssertStatusCode(t, resp, http.StatusCreated)

		var role Role
		integration.ParseJSONResponse(t, resp, &role)
		roleID = role.ID
	})

	t.Run("Get role permissions initially empty", func(t *testing.T) {
		resp := ts.MakeRequest(t, http.MethodGet, "/api/v1/roles/"+roleID+"/permissions", nil, headers)
		integration.AssertStatusCode(t, resp, http.StatusOK)

		var permissions []Permission
		integration.ParseJSONResponse(t, resp, &permissions)
		assert.Empty(t, permissions, "New role should have no permissions")
	})

	// Get some permission IDs to assign
	var permIDs []string
	t.Run("Get available permissions", func(t *testing.T) {
		resp := ts.MakeRequest(t, http.MethodGet, "/api/v1/permissions", nil, headers)
		integration.AssertStatusCode(t, resp, http.StatusOK)

		var permissions []Permission
		integration.ParseJSONResponse(t, resp, &permissions)
		require.NotEmpty(t, permissions)

		// Take first 2 permissions
		for i, perm := range permissions {
			if i >= 2 {
				break
			}
			permIDs = append(permIDs, perm.ID)
		}
	})

	t.Run("Update role permissions", func(t *testing.T) {
		updateReq := map[string]any{
			"permission_ids": permIDs,
		}

		resp := ts.MakeRequest(t, http.MethodPut, "/api/v1/roles/"+roleID+"/permissions", updateReq, headers)
		integration.AssertStatusCode(t, resp, http.StatusNoContent)
	})

	t.Run("Verify permissions were assigned", func(t *testing.T) {
		resp := ts.MakeRequest(t, http.MethodGet, "/api/v1/roles/"+roleID+"/permissions", nil, headers)
		integration.AssertStatusCode(t, resp, http.StatusOK)

		var permissions []Permission
		integration.ParseJSONResponse(t, resp, &permissions)
		assert.Equal(t, len(permIDs), len(permissions), "Role should have the assigned permissions")
	})

	// Cleanup
	t.Run("Delete role", func(t *testing.T) {
		resp := ts.MakeRequest(t, http.MethodDelete, "/api/v1/roles/"+roleID, nil, headers)
		integration.AssertStatusCode(t, resp, http.StatusNoContent)
	})
}
