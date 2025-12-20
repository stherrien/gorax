package rbac

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestDB(t *testing.T) (*sqlx.DB, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	sqlxDB := sqlx.NewDb(db, "sqlmock")
	return sqlxDB, mock
}

func TestRepository_CreateRole(t *testing.T) {
	db, mock := setupTestDB(t)
	defer db.Close()

	repo := NewRepository(db)
	ctx := context.Background()
	tenantID := "tenant-123"

	tests := []struct {
		name      string
		role      *Role
		setupMock func()
		wantErr   bool
	}{
		{
			name: "success",
			role: &Role{
				TenantID:    tenantID,
				Name:        "editor",
				Description: "Editor role",
				IsSystem:    false,
			},
			setupMock: func() {
				mock.ExpectQuery(`INSERT INTO roles`).
					WithArgs(tenantID, "editor", "Editor role", false).
					WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at"}).
						AddRow("role-123", time.Now(), time.Now()))
			},
			wantErr: false,
		},
		{
			name: "duplicate role name",
			role: &Role{
				TenantID:    tenantID,
				Name:        "admin",
				Description: "Admin role",
			},
			setupMock: func() {
				mock.ExpectQuery(`INSERT INTO roles`).
					WithArgs(tenantID, "admin", "Admin role", false).
					WillReturnError(ErrRoleAlreadyExists)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()
			err := repo.CreateRole(ctx, tt.role)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, tt.role.ID)
			}
		})
	}
}

func TestRepository_GetRoleByID(t *testing.T) {
	db, mock := setupTestDB(t)
	defer db.Close()

	repo := NewRepository(db)
	ctx := context.Background()
	roleID := "role-123"
	tenantID := "tenant-123"

	tests := []struct {
		name      string
		roleID    string
		tenantID  string
		setupMock func()
		wantErr   bool
	}{
		{
			name:     "success",
			roleID:   roleID,
			tenantID: tenantID,
			setupMock: func() {
				rows := sqlmock.NewRows([]string{"id", "tenant_id", "name", "description", "is_system", "created_at", "updated_at"}).
					AddRow(roleID, tenantID, "admin", "Admin role", true, time.Now(), time.Now())
				mock.ExpectQuery(`SELECT (.+) FROM roles WHERE id = \$1 AND tenant_id = \$2`).
					WithArgs(roleID, tenantID).
					WillReturnRows(rows)
			},
			wantErr: false,
		},
		{
			name:     "role not found",
			roleID:   "nonexistent",
			tenantID: tenantID,
			setupMock: func() {
				mock.ExpectQuery(`SELECT (.+) FROM roles WHERE id = \$1 AND tenant_id = \$2`).
					WithArgs("nonexistent", tenantID).
					WillReturnError(ErrRoleNotFound)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()
			role, err := repo.GetRoleByID(ctx, tt.roleID, tt.tenantID)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, role)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, role)
				assert.Equal(t, tt.roleID, role.ID)
			}
		})
	}
}

func TestRepository_ListRoles(t *testing.T) {
	db, mock := setupTestDB(t)
	defer db.Close()

	repo := NewRepository(db)
	ctx := context.Background()
	tenantID := "tenant-123"

	t.Run("success", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "tenant_id", "name", "description", "is_system", "created_at", "updated_at"}).
			AddRow("role-1", tenantID, "admin", "Admin role", true, time.Now(), time.Now()).
			AddRow("role-2", tenantID, "editor", "Editor role", false, time.Now(), time.Now())

		mock.ExpectQuery(`SELECT (.+) FROM roles WHERE tenant_id = \$1`).
			WithArgs(tenantID).
			WillReturnRows(rows)

		roles, err := repo.ListRoles(ctx, tenantID)
		assert.NoError(t, err)
		assert.Len(t, roles, 2)
	})
}

func TestRepository_UpdateRole(t *testing.T) {
	db, mock := setupTestDB(t)
	defer db.Close()

	repo := NewRepository(db)
	ctx := context.Background()
	roleID := "role-123"
	tenantID := "tenant-123"

	t.Run("success", func(t *testing.T) {
		role := &Role{
			ID:          roleID,
			TenantID:    tenantID,
			Name:        "updated-editor",
			Description: "Updated editor role",
			IsSystem:    false,
		}

		mock.ExpectExec(`UPDATE roles SET`).
			WithArgs("updated-editor", "Updated editor role", roleID, tenantID).
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := repo.UpdateRole(ctx, role)
		assert.NoError(t, err)
	})

	t.Run("role not found", func(t *testing.T) {
		role := &Role{
			ID:       "nonexistent",
			TenantID: tenantID,
			Name:     "test",
		}

		mock.ExpectExec(`UPDATE roles SET`).
			WithArgs("test", "", "nonexistent", tenantID).
			WillReturnResult(sqlmock.NewResult(0, 0))

		err := repo.UpdateRole(ctx, role)
		assert.Error(t, err)
		assert.Equal(t, ErrRoleNotFound, err)
	})
}

func TestRepository_DeleteRole(t *testing.T) {
	db, mock := setupTestDB(t)
	defer db.Close()

	repo := NewRepository(db)
	ctx := context.Background()
	roleID := "role-123"
	tenantID := "tenant-123"

	t.Run("success", func(t *testing.T) {
		mock.ExpectExec(`DELETE FROM roles WHERE id = \$1 AND tenant_id = \$2`).
			WithArgs(roleID, tenantID).
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := repo.DeleteRole(ctx, roleID, tenantID)
		assert.NoError(t, err)
	})

	t.Run("role not found", func(t *testing.T) {
		mock.ExpectExec(`DELETE FROM roles WHERE id = \$1 AND tenant_id = \$2`).
			WithArgs("nonexistent", tenantID).
			WillReturnResult(sqlmock.NewResult(0, 0))

		err := repo.DeleteRole(ctx, "nonexistent", tenantID)
		assert.Error(t, err)
		assert.Equal(t, ErrRoleNotFound, err)
	})
}

func TestRepository_GetRolePermissions(t *testing.T) {
	db, mock := setupTestDB(t)
	defer db.Close()

	repo := NewRepository(db)
	ctx := context.Background()
	roleID := "role-123"

	t.Run("success", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "resource", "action", "description", "created_at"}).
			AddRow("perm-1", "workflow", "create", "Create workflows", time.Now()).
			AddRow("perm-2", "workflow", "read", "View workflows", time.Now())

		mock.ExpectQuery(`SELECT (.+) FROM permissions p`).
			WithArgs(roleID).
			WillReturnRows(rows)

		perms, err := repo.GetRolePermissions(ctx, roleID)
		assert.NoError(t, err)
		assert.Len(t, perms, 2)
	})
}

func TestRepository_SetRolePermissions(t *testing.T) {
	db, mock := setupTestDB(t)
	defer db.Close()

	repo := NewRepository(db)
	ctx := context.Background()
	roleID := "role-123"
	permissionIDs := []string{"perm-1", "perm-2"}

	t.Run("success", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectExec(`DELETE FROM role_permissions WHERE role_id = \$1`).
			WithArgs(roleID).
			WillReturnResult(sqlmock.NewResult(0, 0))

		mock.ExpectExec(`INSERT INTO role_permissions`).
			WithArgs(roleID, "perm-1").
			WillReturnResult(sqlmock.NewResult(0, 1))

		mock.ExpectExec(`INSERT INTO role_permissions`).
			WithArgs(roleID, "perm-2").
			WillReturnResult(sqlmock.NewResult(0, 1))

		mock.ExpectCommit()

		err := repo.SetRolePermissions(ctx, roleID, permissionIDs)
		assert.NoError(t, err)
	})
}

func TestRepository_GetUserRoles(t *testing.T) {
	db, mock := setupTestDB(t)
	defer db.Close()

	repo := NewRepository(db)
	ctx := context.Background()
	userID := "user-123"
	tenantID := "tenant-123"

	t.Run("success", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "tenant_id", "name", "description", "is_system", "created_at", "updated_at"}).
			AddRow("role-1", tenantID, "admin", "Admin role", true, time.Now(), time.Now())

		mock.ExpectQuery(`SELECT (.+) FROM roles r`).
			WithArgs(userID, tenantID).
			WillReturnRows(rows)

		roles, err := repo.GetUserRoles(ctx, userID, tenantID)
		assert.NoError(t, err)
		assert.Len(t, roles, 1)
	})
}

func TestRepository_AssignRolesToUser(t *testing.T) {
	db, mock := setupTestDB(t)
	defer db.Close()

	repo := NewRepository(db)
	ctx := context.Background()
	userID := "user-123"
	roleIDs := []string{"role-1", "role-2"}
	grantedBy := "admin-123"

	t.Run("success", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectExec(`DELETE FROM user_roles WHERE user_id = \$1`).
			WithArgs(userID).
			WillReturnResult(sqlmock.NewResult(0, 0))

		mock.ExpectExec(`INSERT INTO user_roles`).
			WithArgs(userID, "role-1", grantedBy).
			WillReturnResult(sqlmock.NewResult(0, 1))

		mock.ExpectExec(`INSERT INTO user_roles`).
			WithArgs(userID, "role-2", grantedBy).
			WillReturnResult(sqlmock.NewResult(0, 1))

		mock.ExpectCommit()

		err := repo.AssignRolesToUser(ctx, userID, roleIDs, grantedBy)
		assert.NoError(t, err)
	})
}

func TestRepository_GetUserPermissions(t *testing.T) {
	db, mock := setupTestDB(t)
	defer db.Close()

	repo := NewRepository(db)
	ctx := context.Background()
	userID := "user-123"
	tenantID := "tenant-123"

	t.Run("success", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "resource", "action", "description", "created_at"}).
			AddRow("perm-1", "workflow", "create", "Create workflows", time.Now()).
			AddRow("perm-2", "workflow", "read", "View workflows", time.Now())

		mock.ExpectQuery(`SELECT DISTINCT (.+) FROM permissions p`).
			WithArgs(userID, tenantID).
			WillReturnRows(rows)

		perms, err := repo.GetUserPermissions(ctx, userID, tenantID)
		assert.NoError(t, err)
		assert.Len(t, perms, 2)
	})
}

func TestRepository_HasPermission(t *testing.T) {
	db, mock := setupTestDB(t)
	defer db.Close()

	repo := NewRepository(db)
	ctx := context.Background()
	userID := "user-123"
	tenantID := "tenant-123"

	tests := []struct {
		name      string
		resource  string
		action    string
		setupMock func()
		want      bool
		wantErr   bool
	}{
		{
			name:     "has permission",
			resource: "workflow",
			action:   "create",
			setupMock: func() {
				rows := sqlmock.NewRows([]string{"count"}).AddRow(1)
				mock.ExpectQuery(`SELECT COUNT\(\*\) FROM permissions p`).
					WithArgs(userID, tenantID, "workflow", "create").
					WillReturnRows(rows)
			},
			want:    true,
			wantErr: false,
		},
		{
			name:     "no permission",
			resource: "workflow",
			action:   "delete",
			setupMock: func() {
				rows := sqlmock.NewRows([]string{"count"}).AddRow(0)
				mock.ExpectQuery(`SELECT COUNT\(\*\) FROM permissions p`).
					WithArgs(userID, tenantID, "workflow", "delete").
					WillReturnRows(rows)
			},
			want:    false,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()
			has, err := repo.HasPermission(ctx, userID, tenantID, tt.resource, tt.action)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, has)
			}
		})
	}
}

func TestRepository_ListPermissions(t *testing.T) {
	db, mock := setupTestDB(t)
	defer db.Close()

	repo := NewRepository(db)
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "resource", "action", "description", "created_at"}).
			AddRow("perm-1", "workflow", "create", "Create workflows", time.Now()).
			AddRow("perm-2", "workflow", "read", "View workflows", time.Now())

		mock.ExpectQuery(`SELECT (.+) FROM permissions`).
			WillReturnRows(rows)

		perms, err := repo.ListPermissions(ctx)
		assert.NoError(t, err)
		assert.Len(t, perms, 2)
	})
}

func TestRepository_CreateAuditLog(t *testing.T) {
	db, mock := setupTestDB(t)
	defer db.Close()

	repo := NewRepository(db)
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		log := &AuditLog{
			TenantID:   "tenant-123",
			UserID:     "user-123",
			Action:     AuditActionRoleCreated,
			TargetType: stringPtr("role"),
			TargetID:   stringPtr("role-123"),
			Details:    map[string]interface{}{"role_name": "editor"},
		}

		mock.ExpectExec(`INSERT INTO permission_audit_log`).
			WithArgs(log.TenantID, log.UserID, log.Action, log.TargetType, log.TargetID, sqlmock.AnyArg()).
			WillReturnResult(sqlmock.NewResult(1, 1))

		err := repo.CreateAuditLog(ctx, log)
		assert.NoError(t, err)
	})
}

func stringPtr(s string) *string {
	return &s
}
