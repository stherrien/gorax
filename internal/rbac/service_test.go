package rbac

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockRepository is a mock implementation of the repository
type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) CreateRole(ctx context.Context, role *Role) error {
	args := m.Called(ctx, role)
	return args.Error(0)
}

func (m *MockRepository) GetRoleByID(ctx context.Context, roleID, tenantID string) (*Role, error) {
	args := m.Called(ctx, roleID, tenantID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Role), args.Error(1)
}

func (m *MockRepository) GetRoleByName(ctx context.Context, name, tenantID string) (*Role, error) {
	args := m.Called(ctx, name, tenantID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Role), args.Error(1)
}

func (m *MockRepository) ListRoles(ctx context.Context, tenantID string) ([]*Role, error) {
	args := m.Called(ctx, tenantID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*Role), args.Error(1)
}

func (m *MockRepository) UpdateRole(ctx context.Context, role *Role) error {
	args := m.Called(ctx, role)
	return args.Error(0)
}

func (m *MockRepository) DeleteRole(ctx context.Context, roleID, tenantID string) error {
	args := m.Called(ctx, roleID, tenantID)
	return args.Error(0)
}

func (m *MockRepository) GetRolePermissions(ctx context.Context, roleID string) ([]Permission, error) {
	args := m.Called(ctx, roleID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]Permission), args.Error(1)
}

func (m *MockRepository) SetRolePermissions(ctx context.Context, roleID string, permissionIDs []string) error {
	args := m.Called(ctx, roleID, permissionIDs)
	return args.Error(0)
}

func (m *MockRepository) GetUserRoles(ctx context.Context, userID, tenantID string) ([]*Role, error) {
	args := m.Called(ctx, userID, tenantID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*Role), args.Error(1)
}

func (m *MockRepository) AssignRolesToUser(ctx context.Context, userID string, roleIDs []string, grantedBy string) error {
	args := m.Called(ctx, userID, roleIDs, grantedBy)
	return args.Error(0)
}

func (m *MockRepository) GetUserPermissions(ctx context.Context, userID, tenantID string) ([]Permission, error) {
	args := m.Called(ctx, userID, tenantID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]Permission), args.Error(1)
}

func (m *MockRepository) HasPermission(ctx context.Context, userID, tenantID, resource, action string) (bool, error) {
	args := m.Called(ctx, userID, tenantID, resource, action)
	return args.Bool(0), args.Error(1)
}

func (m *MockRepository) ListPermissions(ctx context.Context) ([]Permission, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]Permission), args.Error(1)
}

func (m *MockRepository) GetPermissionByResourceAction(ctx context.Context, resource, action string) (*Permission, error) {
	args := m.Called(ctx, resource, action)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Permission), args.Error(1)
}

func (m *MockRepository) CreateAuditLog(ctx context.Context, log *AuditLog) error {
	args := m.Called(ctx, log)
	return args.Error(0)
}

func (m *MockRepository) GetAuditLogs(ctx context.Context, tenantID string, limit, offset int) ([]*AuditLog, error) {
	args := m.Called(ctx, tenantID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*AuditLog), args.Error(1)
}

func TestService_CreateRole(t *testing.T) {
	ctx := context.Background()
	tenantID := "tenant-123"
	userID := "user-123"

	tests := []struct {
		name      string
		req       *CreateRoleRequest
		setupMock func(*MockRepository)
		wantErr   error
	}{
		{
			name: "success",
			req: &CreateRoleRequest{
				Name:          "editor",
				Description:   "Editor role",
				PermissionIDs: []string{"perm-1", "perm-2"},
			},
			setupMock: func(repo *MockRepository) {
				repo.On("CreateRole", ctx, mock.MatchedBy(func(r *Role) bool {
					return r.Name == "editor" && r.TenantID == tenantID
				})).Return(nil).Run(func(args mock.Arguments) {
					role := args.Get(1).(*Role)
					role.ID = "role-123"
					role.CreatedAt = time.Now()
					role.UpdatedAt = time.Now()
				})
				repo.On("SetRolePermissions", ctx, "role-123", []string{"perm-1", "perm-2"}).Return(nil)
				repo.On("CreateAuditLog", ctx, mock.AnythingOfType("*rbac.AuditLog")).Return(nil)
			},
			wantErr: nil,
		},
		{
			name: "invalid name",
			req: &CreateRoleRequest{
				Name: "",
			},
			setupMock: func(repo *MockRepository) {},
			wantErr:   ErrInvalidRoleName,
		},
		{
			name: "role already exists",
			req: &CreateRoleRequest{
				Name:        "admin",
				Description: "Admin role",
			},
			setupMock: func(repo *MockRepository) {
				repo.On("CreateRole", ctx, mock.Anything).Return(ErrRoleAlreadyExists)
			},
			wantErr: ErrRoleAlreadyExists,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := new(MockRepository)
			tt.setupMock(repo)

			svc := NewService(repo)
			role, err := svc.CreateRole(ctx, tenantID, userID, tt.req)

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				assert.Nil(t, role)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, role)
				assert.Equal(t, tt.req.Name, role.Name)
			}

			repo.AssertExpectations(t)
		})
	}
}

func TestService_GetRole(t *testing.T) {
	ctx := context.Background()
	roleID := "role-123"
	tenantID := "tenant-123"

	t.Run("success with permissions", func(t *testing.T) {
		repo := new(MockRepository)
		expectedRole := &Role{
			ID:       roleID,
			TenantID: tenantID,
			Name:     "admin",
		}
		expectedPerms := []Permission{
			{ID: "perm-1", Resource: "workflow", Action: "create"},
		}

		repo.On("GetRoleByID", ctx, roleID, tenantID).Return(expectedRole, nil)
		repo.On("GetRolePermissions", ctx, roleID).Return(expectedPerms, nil)

		svc := NewService(repo)
		role, err := svc.GetRole(ctx, roleID, tenantID)

		assert.NoError(t, err)
		assert.NotNil(t, role)
		assert.Equal(t, roleID, role.ID)
		assert.Len(t, role.Permissions, 1)

		repo.AssertExpectations(t)
	})

	t.Run("role not found", func(t *testing.T) {
		repo := new(MockRepository)
		repo.On("GetRoleByID", ctx, "nonexistent", tenantID).Return(nil, ErrRoleNotFound)

		svc := NewService(repo)
		role, err := svc.GetRole(ctx, "nonexistent", tenantID)

		assert.ErrorIs(t, err, ErrRoleNotFound)
		assert.Nil(t, role)

		repo.AssertExpectations(t)
	})
}

func TestService_UpdateRole(t *testing.T) {
	ctx := context.Background()
	roleID := "role-123"
	tenantID := "tenant-123"
	userID := "user-123"

	tests := []struct {
		name      string
		req       *UpdateRoleRequest
		setupMock func(*MockRepository)
		wantErr   error
	}{
		{
			name: "success - update name and permissions",
			req: &UpdateRoleRequest{
				Name:          stringPtr("new-editor"),
				Description:   stringPtr("Updated editor"),
				PermissionIDs: []string{"perm-1", "perm-2"},
			},
			setupMock: func(repo *MockRepository) {
				existingRole := &Role{
					ID:       roleID,
					TenantID: tenantID,
					Name:     "editor",
					IsSystem: false,
				}
				repo.On("GetRoleByID", ctx, roleID, tenantID).Return(existingRole, nil)
				repo.On("UpdateRole", ctx, mock.MatchedBy(func(r *Role) bool {
					return r.Name == "new-editor"
				})).Return(nil)
				repo.On("SetRolePermissions", ctx, roleID, []string{"perm-1", "perm-2"}).Return(nil)
				repo.On("CreateAuditLog", ctx, mock.AnythingOfType("*rbac.AuditLog")).Return(nil)
			},
			wantErr: nil,
		},
		{
			name: "cannot modify system role",
			req: &UpdateRoleRequest{
				Name: stringPtr("new-admin"),
			},
			setupMock: func(repo *MockRepository) {
				systemRole := &Role{
					ID:       roleID,
					TenantID: tenantID,
					Name:     "admin",
					IsSystem: true,
				}
				repo.On("GetRoleByID", ctx, roleID, tenantID).Return(systemRole, nil)
			},
			wantErr: ErrSystemRoleCannotBeModified,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := new(MockRepository)
			tt.setupMock(repo)

			svc := NewService(repo)
			err := svc.UpdateRole(ctx, roleID, tenantID, userID, tt.req)

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				assert.NoError(t, err)
			}

			repo.AssertExpectations(t)
		})
	}
}

func TestService_DeleteRole(t *testing.T) {
	ctx := context.Background()
	roleID := "role-123"
	tenantID := "tenant-123"
	userID := "user-123"

	tests := []struct {
		name      string
		setupMock func(*MockRepository)
		wantErr   error
	}{
		{
			name: "success",
			setupMock: func(repo *MockRepository) {
				role := &Role{
					ID:       roleID,
					TenantID: tenantID,
					Name:     "editor",
					IsSystem: false,
				}
				repo.On("GetRoleByID", ctx, roleID, tenantID).Return(role, nil)
				repo.On("DeleteRole", ctx, roleID, tenantID).Return(nil)
				repo.On("CreateAuditLog", ctx, mock.AnythingOfType("*rbac.AuditLog")).Return(nil)
			},
			wantErr: nil,
		},
		{
			name: "cannot delete system role",
			setupMock: func(repo *MockRepository) {
				systemRole := &Role{
					ID:       roleID,
					TenantID: tenantID,
					Name:     "admin",
					IsSystem: true,
				}
				repo.On("GetRoleByID", ctx, roleID, tenantID).Return(systemRole, nil)
			},
			wantErr: ErrSystemRoleCannotBeDeleted,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := new(MockRepository)
			tt.setupMock(repo)

			svc := NewService(repo)
			err := svc.DeleteRole(ctx, roleID, tenantID, userID)

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				assert.NoError(t, err)
			}

			repo.AssertExpectations(t)
		})
	}
}

func TestService_AssignRolesToUser(t *testing.T) {
	ctx := context.Background()
	userID := "user-123"
	tenantID := "tenant-123"
	grantedBy := "admin-123"

	tests := []struct {
		name      string
		req       *AssignRolesRequest
		setupMock func(*MockRepository)
		wantErr   error
	}{
		{
			name: "success",
			req: &AssignRolesRequest{
				RoleIDs: []string{"role-1", "role-2"},
			},
			setupMock: func(repo *MockRepository) {
				repo.On("GetRoleByID", ctx, "role-1", tenantID).Return(&Role{ID: "role-1", TenantID: tenantID}, nil)
				repo.On("GetRoleByID", ctx, "role-2", tenantID).Return(&Role{ID: "role-2", TenantID: tenantID}, nil)
				repo.On("AssignRolesToUser", ctx, userID, []string{"role-1", "role-2"}, grantedBy).Return(nil)
				repo.On("CreateAuditLog", ctx, mock.AnythingOfType("*rbac.AuditLog")).Return(nil)
			},
			wantErr: nil,
		},
		{
			name: "no roles provided",
			req: &AssignRolesRequest{
				RoleIDs: []string{},
			},
			setupMock: func(repo *MockRepository) {},
			wantErr:   ErrNoRolesProvided,
		},
		{
			name: "role not found",
			req: &AssignRolesRequest{
				RoleIDs: []string{"nonexistent"},
			},
			setupMock: func(repo *MockRepository) {
				repo.On("GetRoleByID", ctx, "nonexistent", tenantID).Return(nil, ErrRoleNotFound)
			},
			wantErr: ErrRoleNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := new(MockRepository)
			tt.setupMock(repo)

			svc := NewService(repo)
			err := svc.AssignRolesToUser(ctx, userID, tenantID, grantedBy, tt.req)

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				assert.NoError(t, err)
			}

			repo.AssertExpectations(t)
		})
	}
}

func TestService_CheckPermission(t *testing.T) {
	ctx := context.Background()
	userID := "user-123"
	tenantID := "tenant-123"

	tests := []struct {
		name      string
		resource  string
		action    string
		setupMock func(*MockRepository)
		want      bool
		wantErr   bool
	}{
		{
			name:     "has permission",
			resource: "workflow",
			action:   "create",
			setupMock: func(repo *MockRepository) {
				repo.On("HasPermission", ctx, userID, tenantID, "workflow", "create").Return(true, nil)
			},
			want:    true,
			wantErr: false,
		},
		{
			name:     "no permission",
			resource: "workflow",
			action:   "delete",
			setupMock: func(repo *MockRepository) {
				repo.On("HasPermission", ctx, userID, tenantID, "workflow", "delete").Return(false, nil)
			},
			want:    false,
			wantErr: false,
		},
		{
			name:     "error checking permission",
			resource: "workflow",
			action:   "create",
			setupMock: func(repo *MockRepository) {
				repo.On("HasPermission", ctx, userID, tenantID, "workflow", "create").
					Return(false, errors.New("database error"))
			},
			want:    false,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := new(MockRepository)
			tt.setupMock(repo)

			svc := NewService(repo)
			has, err := svc.CheckPermission(ctx, userID, tenantID, tt.resource, tt.action)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, has)
			}

			repo.AssertExpectations(t)
		})
	}
}

func TestService_CreateDefaultRoles(t *testing.T) {
	ctx := context.Background()
	tenantID := "tenant-123"

	t.Run("success", func(t *testing.T) {
		repo := new(MockRepository)

		// Get all permissions
		allPerms := []Permission{
			{ID: "perm-1", Resource: "workflow", Action: "create"},
			{ID: "perm-2", Resource: "workflow", Action: "read"},
			{ID: "perm-3", Resource: "workflow", Action: "delete"},
		}
		repo.On("ListPermissions", ctx).Return(allPerms, nil)

		// Create roles
		repo.On("CreateRole", ctx, mock.MatchedBy(func(r *Role) bool {
			return r.Name == RoleAdmin && r.IsSystem
		})).Return(nil).Run(func(args mock.Arguments) {
			args.Get(1).(*Role).ID = "role-admin"
		})
		repo.On("CreateRole", ctx, mock.MatchedBy(func(r *Role) bool {
			return r.Name == RoleEditor && r.IsSystem
		})).Return(nil).Run(func(args mock.Arguments) {
			args.Get(1).(*Role).ID = "role-editor"
		})
		repo.On("CreateRole", ctx, mock.MatchedBy(func(r *Role) bool {
			return r.Name == RoleViewer && r.IsSystem
		})).Return(nil).Run(func(args mock.Arguments) {
			args.Get(1).(*Role).ID = "role-viewer"
		})
		repo.On("CreateRole", ctx, mock.MatchedBy(func(r *Role) bool {
			return r.Name == RoleOperator && r.IsSystem
		})).Return(nil).Run(func(args mock.Arguments) {
			args.Get(1).(*Role).ID = "role-operator"
		})

		// Set permissions
		repo.On("SetRolePermissions", ctx, mock.Anything, mock.Anything).Return(nil).Times(4)

		svc := NewService(repo)
		err := svc.CreateDefaultRoles(ctx, tenantID)

		assert.NoError(t, err)
		repo.AssertExpectations(t)
	})
}
