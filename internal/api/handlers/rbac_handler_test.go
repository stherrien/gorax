package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/gorax/gorax/internal/rbac"
)

// MockRBACRepository implements rbac.RepositoryInterface for testing
type MockRBACRepository struct {
	mock.Mock
}

func (m *MockRBACRepository) CreateRole(ctx context.Context, role *rbac.Role) error {
	args := m.Called(ctx, role)
	// Simulate ID assignment
	if role.ID == "" {
		role.ID = "role-generated-id"
	}
	return args.Error(0)
}

func (m *MockRBACRepository) GetRoleByID(ctx context.Context, roleID, tenantID string) (*rbac.Role, error) {
	args := m.Called(ctx, roleID, tenantID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*rbac.Role), args.Error(1)
}

func (m *MockRBACRepository) GetRoleByName(ctx context.Context, name, tenantID string) (*rbac.Role, error) {
	args := m.Called(ctx, name, tenantID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*rbac.Role), args.Error(1)
}

func (m *MockRBACRepository) ListRoles(ctx context.Context, tenantID string) ([]*rbac.Role, error) {
	args := m.Called(ctx, tenantID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*rbac.Role), args.Error(1)
}

func (m *MockRBACRepository) UpdateRole(ctx context.Context, role *rbac.Role) error {
	args := m.Called(ctx, role)
	return args.Error(0)
}

func (m *MockRBACRepository) DeleteRole(ctx context.Context, roleID, tenantID string) error {
	args := m.Called(ctx, roleID, tenantID)
	return args.Error(0)
}

func (m *MockRBACRepository) GetRolePermissions(ctx context.Context, roleID string) ([]rbac.Permission, error) {
	args := m.Called(ctx, roleID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]rbac.Permission), args.Error(1)
}

func (m *MockRBACRepository) SetRolePermissions(ctx context.Context, roleID string, permissionIDs []string) error {
	args := m.Called(ctx, roleID, permissionIDs)
	return args.Error(0)
}

func (m *MockRBACRepository) GetUserRoles(ctx context.Context, userID, tenantID string) ([]*rbac.Role, error) {
	args := m.Called(ctx, userID, tenantID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*rbac.Role), args.Error(1)
}

func (m *MockRBACRepository) AssignRolesToUser(ctx context.Context, userID string, roleIDs []string, grantedBy string) error {
	args := m.Called(ctx, userID, roleIDs, grantedBy)
	return args.Error(0)
}

func (m *MockRBACRepository) GetUserPermissions(ctx context.Context, userID, tenantID string) ([]rbac.Permission, error) {
	args := m.Called(ctx, userID, tenantID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]rbac.Permission), args.Error(1)
}

func (m *MockRBACRepository) HasPermission(ctx context.Context, userID, tenantID, resource, action string) (bool, error) {
	args := m.Called(ctx, userID, tenantID, resource, action)
	return args.Bool(0), args.Error(1)
}

func (m *MockRBACRepository) ListPermissions(ctx context.Context) ([]rbac.Permission, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]rbac.Permission), args.Error(1)
}

func (m *MockRBACRepository) GetPermissionByResourceAction(ctx context.Context, resource, action string) (*rbac.Permission, error) {
	args := m.Called(ctx, resource, action)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*rbac.Permission), args.Error(1)
}

func (m *MockRBACRepository) CreateAuditLog(ctx context.Context, log *rbac.AuditLog) error {
	args := m.Called(ctx, log)
	return args.Error(0)
}

func (m *MockRBACRepository) GetAuditLogs(ctx context.Context, tenantID string, limit, offset int) ([]*rbac.AuditLog, error) {
	args := m.Called(ctx, tenantID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*rbac.AuditLog), args.Error(1)
}

// Helper to create a new test RBAC handler with mock repository
func newTestRBACHandler() (*RBACHandler, *MockRBACRepository) {
	mockRepo := new(MockRBACRepository)
	service := rbac.NewService(mockRepo)
	handler := NewRBACHandler(service)
	return handler, mockRepo
}

// Helper to add string-based context values that the RBAC handler expects
func addRBACContext(req *http.Request, tenantID, userID string) *http.Request {
	ctx := context.WithValue(req.Context(), "tenant_id", tenantID)
	ctx = context.WithValue(ctx, "user_id", userID)
	return req.WithContext(ctx)
}

func addRBACTenantContext(req *http.Request, tenantID string) *http.Request {
	ctx := context.WithValue(req.Context(), "tenant_id", tenantID)
	return req.WithContext(ctx)
}

// Helper to add chi URL params
func addURLParam(req *http.Request, key, value string) *http.Request {
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add(key, value)
	return req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
}

func TestRBACHandler_ListRoles(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name           string
		tenantID       string
		mockRoles      []*rbac.Role
		mockError      error
		expectedStatus int
		checkResponse  func(t *testing.T, body []byte)
	}{
		{
			name:     "success",
			tenantID: "tenant-123",
			mockRoles: []*rbac.Role{
				{
					ID:          "role-1",
					TenantID:    "tenant-123",
					Name:        "admin",
					Description: "Administrator role",
					IsSystem:    true,
					CreatedAt:   now,
					UpdatedAt:   now,
				},
				{
					ID:          "role-2",
					TenantID:    "tenant-123",
					Name:        "editor",
					Description: "Editor role",
					IsSystem:    false,
					CreatedAt:   now,
					UpdatedAt:   now,
				},
			},
			mockError:      nil,
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, body []byte) {
				var roles []rbac.Role
				err := json.Unmarshal(body, &roles)
				require.NoError(t, err)
				assert.Len(t, roles, 2)
				assert.Equal(t, "admin", roles[0].Name)
			},
		},
		{
			name:           "empty list",
			tenantID:       "tenant-123",
			mockRoles:      []*rbac.Role{},
			mockError:      nil,
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, body []byte) {
				var roles []rbac.Role
				err := json.Unmarshal(body, &roles)
				require.NoError(t, err)
				assert.Len(t, roles, 0)
			},
		},
		{
			name:           "service error",
			tenantID:       "tenant-123",
			mockRoles:      nil,
			mockError:      assert.AnError,
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, mockRepo := newTestRBACHandler()

			mockRepo.On("ListRoles", mock.Anything, tt.tenantID).Return(tt.mockRoles, tt.mockError)
			// For each role, mock GetRolePermissions
			if tt.mockRoles != nil {
				for _, role := range tt.mockRoles {
					mockRepo.On("GetRolePermissions", mock.Anything, role.ID).Return([]rbac.Permission{}, nil)
				}
			}

			req := httptest.NewRequest(http.MethodGet, "/api/v1/roles", nil)
			req = addRBACTenantContext(req, tt.tenantID)
			w := httptest.NewRecorder()

			handler.ListRoles(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.checkResponse != nil {
				tt.checkResponse(t, w.Body.Bytes())
			}
		})
	}
}

func TestRBACHandler_CreateRole(t *testing.T) {
	tests := []struct {
		name           string
		tenantID       string
		userID         string
		requestBody    rbac.CreateRoleRequest
		mockError      error
		expectedStatus int
	}{
		{
			name:     "success",
			tenantID: "tenant-123",
			userID:   "user-123",
			requestBody: rbac.CreateRoleRequest{
				Name:        "custom-role",
				Description: "A custom role",
			},
			mockError:      nil,
			expectedStatus: http.StatusCreated,
		},
		{
			name:     "invalid name - empty",
			tenantID: "tenant-123",
			userID:   "user-123",
			requestBody: rbac.CreateRoleRequest{
				Name:        "",
				Description: "A custom role",
			},
			mockError:      nil, // Validation happens before mock is called
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:     "role already exists",
			tenantID: "tenant-123",
			userID:   "user-123",
			requestBody: rbac.CreateRoleRequest{
				Name:        "existing-role",
				Description: "An existing role",
			},
			mockError:      rbac.ErrRoleAlreadyExists,
			expectedStatus: http.StatusConflict,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, mockRepo := newTestRBACHandler()

			// Only set up mock if validation passes
			if tt.requestBody.Name != "" {
				mockRepo.On("CreateRole", mock.Anything, mock.AnythingOfType("*rbac.Role")).Return(tt.mockError)
				// Mock audit log (best effort)
				mockRepo.On("CreateAuditLog", mock.Anything, mock.Anything).Return(nil).Maybe()
			}

			body, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest(http.MethodPost, "/api/v1/roles", bytes.NewReader(body))
			req = addRBACContext(req, tt.tenantID, tt.userID)
			w := httptest.NewRecorder()

			handler.CreateRole(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestRBACHandler_CreateRole_InvalidJSON(t *testing.T) {
	handler, _ := newTestRBACHandler()

	req := httptest.NewRequest(http.MethodPost, "/api/v1/roles", bytes.NewReader([]byte("invalid-json")))
	req = addRBACContext(req, "tenant-123", "user-123")
	w := httptest.NewRecorder()

	handler.CreateRole(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestRBACHandler_GetRole(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name           string
		tenantID       string
		roleID         string
		mockRole       *rbac.Role
		mockError      error
		expectedStatus int
	}{
		{
			name:     "success",
			tenantID: "tenant-123",
			roleID:   "role-123",
			mockRole: &rbac.Role{
				ID:          "role-123",
				TenantID:    "tenant-123",
				Name:        "admin",
				Description: "Administrator role",
				IsSystem:    true,
				CreatedAt:   now,
				UpdatedAt:   now,
			},
			mockError:      nil,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "not found",
			tenantID:       "tenant-123",
			roleID:         "non-existent",
			mockRole:       nil,
			mockError:      rbac.ErrRoleNotFound,
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "internal error",
			tenantID:       "tenant-123",
			roleID:         "role-123",
			mockRole:       nil,
			mockError:      assert.AnError,
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, mockRepo := newTestRBACHandler()

			mockRepo.On("GetRoleByID", mock.Anything, tt.roleID, tt.tenantID).Return(tt.mockRole, tt.mockError)
			if tt.mockRole != nil {
				mockRepo.On("GetRolePermissions", mock.Anything, tt.roleID).Return([]rbac.Permission{}, nil)
			}

			req := httptest.NewRequest(http.MethodGet, "/api/v1/roles/"+tt.roleID, nil)
			req = addRBACTenantContext(req, tt.tenantID)
			req = addURLParam(req, "id", tt.roleID)
			w := httptest.NewRecorder()

			handler.GetRole(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestRBACHandler_UpdateRole(t *testing.T) {
	now := time.Now()
	newName := "updated-role"

	tests := []struct {
		name           string
		tenantID       string
		userID         string
		roleID         string
		requestBody    rbac.UpdateRoleRequest
		mockRole       *rbac.Role
		mockGetError   error
		mockError      error
		expectedStatus int
	}{
		{
			name:     "success",
			tenantID: "tenant-123",
			userID:   "user-123",
			roleID:   "role-123",
			requestBody: rbac.UpdateRoleRequest{
				Name: &newName,
			},
			mockRole: &rbac.Role{
				ID:        "role-123",
				TenantID:  "tenant-123",
				Name:      "old-name",
				IsSystem:  false,
				CreatedAt: now,
				UpdatedAt: now,
			},
			mockGetError:   nil,
			mockError:      nil,
			expectedStatus: http.StatusNoContent,
		},
		{
			name:     "role not found",
			tenantID: "tenant-123",
			userID:   "user-123",
			roleID:   "non-existent",
			requestBody: rbac.UpdateRoleRequest{
				Name: &newName,
			},
			mockRole:       nil,
			mockGetError:   rbac.ErrRoleNotFound,
			expectedStatus: http.StatusNotFound,
		},
		{
			name:     "system role cannot be modified",
			tenantID: "tenant-123",
			userID:   "user-123",
			roleID:   "role-123",
			requestBody: rbac.UpdateRoleRequest{
				Name: &newName,
			},
			mockRole: &rbac.Role{
				ID:       "role-123",
				TenantID: "tenant-123",
				Name:     "admin",
				IsSystem: true,
			},
			mockGetError:   nil,
			expectedStatus: http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, mockRepo := newTestRBACHandler()

			mockRepo.On("GetRoleByID", mock.Anything, tt.roleID, tt.tenantID).Return(tt.mockRole, tt.mockGetError)
			if tt.mockRole != nil && !tt.mockRole.IsSystem {
				mockRepo.On("UpdateRole", mock.Anything, mock.AnythingOfType("*rbac.Role")).Return(tt.mockError)
				mockRepo.On("CreateAuditLog", mock.Anything, mock.Anything).Return(nil).Maybe()
			}

			body, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest(http.MethodPut, "/api/v1/roles/"+tt.roleID, bytes.NewReader(body))
			req = addRBACContext(req, tt.tenantID, tt.userID)
			req = addURLParam(req, "id", tt.roleID)
			w := httptest.NewRecorder()

			handler.UpdateRole(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestRBACHandler_UpdateRole_InvalidJSON(t *testing.T) {
	handler, _ := newTestRBACHandler()

	req := httptest.NewRequest(http.MethodPut, "/api/v1/roles/role-123", bytes.NewReader([]byte("invalid")))
	req = addRBACContext(req, "tenant-123", "user-123")
	req = addURLParam(req, "id", "role-123")
	w := httptest.NewRecorder()

	handler.UpdateRole(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestRBACHandler_DeleteRole(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name           string
		tenantID       string
		userID         string
		roleID         string
		mockRole       *rbac.Role
		mockGetError   error
		mockError      error
		expectedStatus int
	}{
		{
			name:     "success",
			tenantID: "tenant-123",
			userID:   "user-123",
			roleID:   "role-123",
			mockRole: &rbac.Role{
				ID:        "role-123",
				TenantID:  "tenant-123",
				Name:      "custom-role",
				IsSystem:  false,
				CreatedAt: now,
				UpdatedAt: now,
			},
			mockGetError:   nil,
			mockError:      nil,
			expectedStatus: http.StatusNoContent,
		},
		{
			name:           "role not found",
			tenantID:       "tenant-123",
			userID:         "user-123",
			roleID:         "non-existent",
			mockRole:       nil,
			mockGetError:   rbac.ErrRoleNotFound,
			expectedStatus: http.StatusNotFound,
		},
		{
			name:     "system role cannot be deleted",
			tenantID: "tenant-123",
			userID:   "user-123",
			roleID:   "role-123",
			mockRole: &rbac.Role{
				ID:       "role-123",
				TenantID: "tenant-123",
				Name:     "admin",
				IsSystem: true,
			},
			mockGetError:   nil,
			expectedStatus: http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, mockRepo := newTestRBACHandler()

			mockRepo.On("GetRoleByID", mock.Anything, tt.roleID, tt.tenantID).Return(tt.mockRole, tt.mockGetError)
			if tt.mockRole != nil && !tt.mockRole.IsSystem {
				mockRepo.On("DeleteRole", mock.Anything, tt.roleID, tt.tenantID).Return(tt.mockError)
				mockRepo.On("CreateAuditLog", mock.Anything, mock.Anything).Return(nil).Maybe()
			}

			req := httptest.NewRequest(http.MethodDelete, "/api/v1/roles/"+tt.roleID, nil)
			req = addRBACContext(req, tt.tenantID, tt.userID)
			req = addURLParam(req, "id", tt.roleID)
			w := httptest.NewRecorder()

			handler.DeleteRole(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestRBACHandler_GetRolePermissions(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name            string
		tenantID        string
		roleID          string
		mockRole        *rbac.Role
		mockPermissions []rbac.Permission
		mockGetError    error
		mockPermError   error
		expectedStatus  int
	}{
		{
			name:     "success",
			tenantID: "tenant-123",
			roleID:   "role-123",
			mockRole: &rbac.Role{
				ID:       "role-123",
				TenantID: "tenant-123",
			},
			mockPermissions: []rbac.Permission{
				{ID: "perm-1", Resource: "workflow", Action: "read", CreatedAt: now},
				{ID: "perm-2", Resource: "workflow", Action: "execute", CreatedAt: now},
			},
			mockGetError:   nil,
			mockPermError:  nil,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "role not found",
			tenantID:       "tenant-123",
			roleID:         "non-existent",
			mockRole:       nil,
			mockGetError:   rbac.ErrRoleNotFound,
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, mockRepo := newTestRBACHandler()

			mockRepo.On("GetRoleByID", mock.Anything, tt.roleID, tt.tenantID).Return(tt.mockRole, tt.mockGetError)
			if tt.mockRole != nil {
				mockRepo.On("GetRolePermissions", mock.Anything, tt.roleID).Return(tt.mockPermissions, tt.mockPermError)
			}

			req := httptest.NewRequest(http.MethodGet, "/api/v1/roles/"+tt.roleID+"/permissions", nil)
			req = addRBACTenantContext(req, tt.tenantID)
			req = addURLParam(req, "id", tt.roleID)
			w := httptest.NewRecorder()

			handler.GetRolePermissions(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestRBACHandler_UpdateRolePermissions(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name           string
		tenantID       string
		userID         string
		roleID         string
		permissionIDs  []string
		mockRole       *rbac.Role
		mockGetError   error
		mockError      error
		expectedStatus int
	}{
		{
			name:          "success",
			tenantID:      "tenant-123",
			userID:        "user-123",
			roleID:        "role-123",
			permissionIDs: []string{"perm-1", "perm-2"},
			mockRole: &rbac.Role{
				ID:        "role-123",
				TenantID:  "tenant-123",
				Name:      "custom-role",
				IsSystem:  false,
				CreatedAt: now,
				UpdatedAt: now,
			},
			mockGetError:   nil,
			mockError:      nil,
			expectedStatus: http.StatusNoContent,
		},
		{
			name:           "role not found",
			tenantID:       "tenant-123",
			userID:         "user-123",
			roleID:         "non-existent",
			permissionIDs:  []string{"perm-1"},
			mockRole:       nil,
			mockGetError:   rbac.ErrRoleNotFound,
			expectedStatus: http.StatusNotFound,
		},
		{
			name:          "system role cannot be modified",
			tenantID:      "tenant-123",
			userID:        "user-123",
			roleID:        "role-123",
			permissionIDs: []string{"perm-1"},
			mockRole: &rbac.Role{
				ID:       "role-123",
				TenantID: "tenant-123",
				Name:     "admin",
				IsSystem: true,
			},
			mockGetError:   nil,
			expectedStatus: http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, mockRepo := newTestRBACHandler()

			mockRepo.On("GetRoleByID", mock.Anything, tt.roleID, tt.tenantID).Return(tt.mockRole, tt.mockGetError)
			if tt.mockRole != nil && !tt.mockRole.IsSystem {
				mockRepo.On("SetRolePermissions", mock.Anything, tt.roleID, tt.permissionIDs).Return(tt.mockError)
				mockRepo.On("CreateAuditLog", mock.Anything, mock.Anything).Return(nil).Maybe()
			}

			body, _ := json.Marshal(map[string][]string{"permission_ids": tt.permissionIDs})
			req := httptest.NewRequest(http.MethodPut, "/api/v1/roles/"+tt.roleID+"/permissions", bytes.NewReader(body))
			req = addRBACContext(req, tt.tenantID, tt.userID)
			req = addURLParam(req, "id", tt.roleID)
			w := httptest.NewRecorder()

			handler.UpdateRolePermissions(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestRBACHandler_UpdateRolePermissions_InvalidJSON(t *testing.T) {
	handler, _ := newTestRBACHandler()

	req := httptest.NewRequest(http.MethodPut, "/api/v1/roles/role-123/permissions", bytes.NewReader([]byte("invalid")))
	req = addRBACContext(req, "tenant-123", "user-123")
	req = addURLParam(req, "id", "role-123")
	w := httptest.NewRecorder()

	handler.UpdateRolePermissions(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestRBACHandler_GetUserRoles(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name           string
		tenantID       string
		userID         string
		mockRoles      []*rbac.Role
		mockError      error
		expectedStatus int
	}{
		{
			name:     "success",
			tenantID: "tenant-123",
			userID:   "user-456",
			mockRoles: []*rbac.Role{
				{ID: "role-1", TenantID: "tenant-123", Name: "admin", CreatedAt: now, UpdatedAt: now},
			},
			mockError:      nil,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "no roles",
			tenantID:       "tenant-123",
			userID:         "user-456",
			mockRoles:      []*rbac.Role{},
			mockError:      nil,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "error",
			tenantID:       "tenant-123",
			userID:         "user-456",
			mockRoles:      nil,
			mockError:      assert.AnError,
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, mockRepo := newTestRBACHandler()

			mockRepo.On("GetUserRoles", mock.Anything, tt.userID, tt.tenantID).Return(tt.mockRoles, tt.mockError)

			req := httptest.NewRequest(http.MethodGet, "/api/v1/users/"+tt.userID+"/roles", nil)
			req = addRBACTenantContext(req, tt.tenantID)
			req = addURLParam(req, "id", tt.userID)
			w := httptest.NewRecorder()

			handler.GetUserRoles(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestRBACHandler_AssignUserRoles(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name           string
		tenantID       string
		grantedBy      string
		targetUserID   string
		roleIDs        []string
		mockRoles      []*rbac.Role
		mockError      error
		expectedStatus int
	}{
		{
			name:         "success",
			tenantID:     "tenant-123",
			grantedBy:    "admin-user",
			targetUserID: "user-456",
			roleIDs:      []string{"role-1", "role-2"},
			mockRoles: []*rbac.Role{
				{ID: "role-1", TenantID: "tenant-123", Name: "editor", CreatedAt: now, UpdatedAt: now},
				{ID: "role-2", TenantID: "tenant-123", Name: "viewer", CreatedAt: now, UpdatedAt: now},
			},
			mockError:      nil,
			expectedStatus: http.StatusNoContent,
		},
		{
			name:           "no roles provided",
			tenantID:       "tenant-123",
			grantedBy:      "admin-user",
			targetUserID:   "user-456",
			roleIDs:        []string{},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "role not found",
			tenantID:       "tenant-123",
			grantedBy:      "admin-user",
			targetUserID:   "user-456",
			roleIDs:        []string{"non-existent"},
			mockRoles:      nil,
			mockError:      rbac.ErrRoleNotFound,
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, mockRepo := newTestRBACHandler()

			// Mock role lookups
			if len(tt.roleIDs) > 0 && tt.name != "no roles provided" {
				for i, roleID := range tt.roleIDs {
					if tt.mockRoles != nil && i < len(tt.mockRoles) {
						mockRepo.On("GetRoleByID", mock.Anything, roleID, tt.tenantID).Return(tt.mockRoles[i], nil)
					} else {
						mockRepo.On("GetRoleByID", mock.Anything, roleID, tt.tenantID).Return(nil, tt.mockError)
					}
				}
			}

			if tt.mockRoles != nil && tt.mockError == nil {
				mockRepo.On("AssignRolesToUser", mock.Anything, tt.targetUserID, tt.roleIDs, tt.grantedBy).Return(nil)
				mockRepo.On("CreateAuditLog", mock.Anything, mock.Anything).Return(nil).Maybe()
			}

			body, _ := json.Marshal(rbac.AssignRolesRequest{RoleIDs: tt.roleIDs})
			req := httptest.NewRequest(http.MethodPut, "/api/v1/users/"+tt.targetUserID+"/roles", bytes.NewReader(body))
			req = addRBACContext(req, tt.tenantID, tt.grantedBy)
			req = addURLParam(req, "id", tt.targetUserID)
			w := httptest.NewRecorder()

			handler.AssignUserRoles(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestRBACHandler_AssignUserRoles_InvalidJSON(t *testing.T) {
	handler, _ := newTestRBACHandler()

	req := httptest.NewRequest(http.MethodPut, "/api/v1/users/user-123/roles", bytes.NewReader([]byte("invalid")))
	req = addRBACContext(req, "tenant-123", "admin-user")
	req = addURLParam(req, "id", "user-123")
	w := httptest.NewRecorder()

	handler.AssignUserRoles(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestRBACHandler_ListPermissions(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name            string
		mockPermissions []rbac.Permission
		mockError       error
		expectedStatus  int
	}{
		{
			name: "success",
			mockPermissions: []rbac.Permission{
				{ID: "perm-1", Resource: "workflow", Action: "read", Description: "Read workflows", CreatedAt: now},
				{ID: "perm-2", Resource: "workflow", Action: "create", Description: "Create workflows", CreatedAt: now},
			},
			mockError:      nil,
			expectedStatus: http.StatusOK,
		},
		{
			name:            "error",
			mockPermissions: nil,
			mockError:       assert.AnError,
			expectedStatus:  http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, mockRepo := newTestRBACHandler()

			mockRepo.On("ListPermissions", mock.Anything).Return(tt.mockPermissions, tt.mockError)

			req := httptest.NewRequest(http.MethodGet, "/api/v1/permissions", nil)
			w := httptest.NewRecorder()

			handler.ListPermissions(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestRBACHandler_GetUserPermissions(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name            string
		tenantID        string
		userID          string
		mockPermissions []rbac.Permission
		mockError       error
		expectedStatus  int
	}{
		{
			name:     "success",
			tenantID: "tenant-123",
			userID:   "user-456",
			mockPermissions: []rbac.Permission{
				{ID: "perm-1", Resource: "workflow", Action: "read", CreatedAt: now},
			},
			mockError:      nil,
			expectedStatus: http.StatusOK,
		},
		{
			name:            "error",
			tenantID:        "tenant-123",
			userID:          "user-456",
			mockPermissions: nil,
			mockError:       assert.AnError,
			expectedStatus:  http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, mockRepo := newTestRBACHandler()

			mockRepo.On("GetUserPermissions", mock.Anything, tt.userID, tt.tenantID).Return(tt.mockPermissions, tt.mockError)

			req := httptest.NewRequest(http.MethodGet, "/api/v1/users/"+tt.userID+"/permissions", nil)
			req = addRBACTenantContext(req, tt.tenantID)
			req = addURLParam(req, "id", tt.userID)
			w := httptest.NewRecorder()

			handler.GetUserPermissions(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestRBACHandler_GetCurrentUserPermissions(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name            string
		tenantID        string
		userID          string
		mockPermissions []rbac.Permission
		mockError       error
		expectedStatus  int
	}{
		{
			name:     "success",
			tenantID: "tenant-123",
			userID:   "current-user-123",
			mockPermissions: []rbac.Permission{
				{ID: "perm-1", Resource: "workflow", Action: "read", CreatedAt: now},
				{ID: "perm-2", Resource: "workflow", Action: "execute", CreatedAt: now},
			},
			mockError:      nil,
			expectedStatus: http.StatusOK,
		},
		{
			name:            "error",
			tenantID:        "tenant-123",
			userID:          "current-user-123",
			mockPermissions: nil,
			mockError:       assert.AnError,
			expectedStatus:  http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, mockRepo := newTestRBACHandler()

			mockRepo.On("GetUserPermissions", mock.Anything, tt.userID, tt.tenantID).Return(tt.mockPermissions, tt.mockError)

			req := httptest.NewRequest(http.MethodGet, "/api/v1/me/permissions", nil)
			req = addRBACContext(req, tt.tenantID, tt.userID)
			w := httptest.NewRecorder()

			handler.GetCurrentUserPermissions(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestRBACHandler_GetAuditLogs(t *testing.T) {
	now := time.Now()
	targetType := "role"
	targetID := "role-123"

	tests := []struct {
		name           string
		tenantID       string
		queryParams    string
		expectedLimit  int
		expectedOffset int
		mockLogs       []*rbac.AuditLog
		mockError      error
		expectedStatus int
	}{
		{
			name:           "success with defaults",
			tenantID:       "tenant-123",
			queryParams:    "",
			expectedLimit:  50,
			expectedOffset: 0,
			mockLogs: []*rbac.AuditLog{
				{
					ID:         "log-1",
					TenantID:   "tenant-123",
					UserID:     "user-123",
					Action:     "role_created",
					TargetType: &targetType,
					TargetID:   &targetID,
					CreatedAt:  now,
				},
			},
			mockError:      nil,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "with pagination",
			tenantID:       "tenant-123",
			queryParams:    "?limit=10&offset=20",
			expectedLimit:  10,
			expectedOffset: 20,
			mockLogs:       []*rbac.AuditLog{},
			mockError:      nil,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "limit capped at 100",
			tenantID:       "tenant-123",
			queryParams:    "?limit=200",
			expectedLimit:  50, // invalid limit, use default
			expectedOffset: 0,
			mockLogs:       []*rbac.AuditLog{},
			mockError:      nil,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "invalid limit uses default",
			tenantID:       "tenant-123",
			queryParams:    "?limit=invalid",
			expectedLimit:  50,
			expectedOffset: 0,
			mockLogs:       []*rbac.AuditLog{},
			mockError:      nil,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "error",
			tenantID:       "tenant-123",
			queryParams:    "",
			expectedLimit:  50,
			expectedOffset: 0,
			mockLogs:       nil,
			mockError:      assert.AnError,
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, mockRepo := newTestRBACHandler()

			mockRepo.On("GetAuditLogs", mock.Anything, tt.tenantID, tt.expectedLimit, tt.expectedOffset).Return(tt.mockLogs, tt.mockError)

			req := httptest.NewRequest(http.MethodGet, "/api/v1/audit-logs"+tt.queryParams, nil)
			req = addRBACTenantContext(req, tt.tenantID)
			w := httptest.NewRecorder()

			handler.GetAuditLogs(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}
