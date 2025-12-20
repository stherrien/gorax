package rbac

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestRequirePermission(t *testing.T) {
	tests := []struct {
		name           string
		resource       string
		action         string
		setupMock      func(*MockRepository)
		setupContext   func(*http.Request) *http.Request
		expectedStatus int
		expectHandler  bool
	}{
		{
			name:     "success - has permission",
			resource: "workflow",
			action:   "create",
			setupMock: func(repo *MockRepository) {
				repo.On("HasPermission", mock.Anything, "user-123", "tenant-123", "workflow", "create").
					Return(true, nil)
			},
			setupContext: func(r *http.Request) *http.Request {
				ctx := context.WithValue(r.Context(), "user_id", "user-123")
				ctx = context.WithValue(ctx, "tenant_id", "tenant-123")
				return r.WithContext(ctx)
			},
			expectedStatus: http.StatusOK,
			expectHandler:  true,
		},
		{
			name:     "permission denied",
			resource: "workflow",
			action:   "delete",
			setupMock: func(repo *MockRepository) {
				repo.On("HasPermission", mock.Anything, "user-123", "tenant-123", "workflow", "delete").
					Return(false, nil)
			},
			setupContext: func(r *http.Request) *http.Request {
				ctx := context.WithValue(r.Context(), "user_id", "user-123")
				ctx = context.WithValue(ctx, "tenant_id", "tenant-123")
				return r.WithContext(ctx)
			},
			expectedStatus: http.StatusForbidden,
			expectHandler:  false,
		},
		{
			name:     "missing user_id",
			resource: "workflow",
			action:   "create",
			setupMock: func(repo *MockRepository) {
				// Should not be called
			},
			setupContext: func(r *http.Request) *http.Request {
				ctx := context.WithValue(r.Context(), "tenant_id", "tenant-123")
				return r.WithContext(ctx)
			},
			expectedStatus: http.StatusUnauthorized,
			expectHandler:  false,
		},
		{
			name:     "missing tenant_id",
			resource: "workflow",
			action:   "create",
			setupMock: func(repo *MockRepository) {
				// Should not be called
			},
			setupContext: func(r *http.Request) *http.Request {
				ctx := context.WithValue(r.Context(), "user_id", "user-123")
				return r.WithContext(ctx)
			},
			expectedStatus: http.StatusBadRequest,
			expectHandler:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := new(MockRepository)
			tt.setupMock(repo)

			handlerCalled := false
			nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				handlerCalled = true
				w.WriteHeader(http.StatusOK)
			})

			middleware := RequirePermission(repo, tt.resource, tt.action)
			handler := middleware(nextHandler)

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			req = tt.setupContext(req)
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.Equal(t, tt.expectHandler, handlerCalled)

			repo.AssertExpectations(t)
		})
	}
}

func TestRequireAnyPermission(t *testing.T) {
	tests := []struct {
		name           string
		permissions    []PermissionCheck
		setupMock      func(*MockRepository)
		setupContext   func(*http.Request) *http.Request
		expectedStatus int
		expectHandler  bool
	}{
		{
			name: "success - has one of multiple permissions",
			permissions: []PermissionCheck{
				{Resource: "workflow", Action: "create"},
				{Resource: "workflow", Action: "update"},
			},
			setupMock: func(repo *MockRepository) {
				repo.On("HasPermission", mock.Anything, "user-123", "tenant-123", "workflow", "create").
					Return(false, nil)
				repo.On("HasPermission", mock.Anything, "user-123", "tenant-123", "workflow", "update").
					Return(true, nil)
			},
			setupContext: func(r *http.Request) *http.Request {
				ctx := context.WithValue(r.Context(), "user_id", "user-123")
				ctx = context.WithValue(ctx, "tenant_id", "tenant-123")
				return r.WithContext(ctx)
			},
			expectedStatus: http.StatusOK,
			expectHandler:  true,
		},
		{
			name: "permission denied - has none",
			permissions: []PermissionCheck{
				{Resource: "workflow", Action: "delete"},
				{Resource: "workflow", Action: "execute"},
			},
			setupMock: func(repo *MockRepository) {
				repo.On("HasPermission", mock.Anything, "user-123", "tenant-123", "workflow", "delete").
					Return(false, nil)
				repo.On("HasPermission", mock.Anything, "user-123", "tenant-123", "workflow", "execute").
					Return(false, nil)
			},
			setupContext: func(r *http.Request) *http.Request {
				ctx := context.WithValue(r.Context(), "user_id", "user-123")
				ctx = context.WithValue(ctx, "tenant_id", "tenant-123")
				return r.WithContext(ctx)
			},
			expectedStatus: http.StatusForbidden,
			expectHandler:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := new(MockRepository)
			tt.setupMock(repo)

			handlerCalled := false
			nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				handlerCalled = true
				w.WriteHeader(http.StatusOK)
			})

			middleware := RequireAnyPermission(repo, tt.permissions...)
			handler := middleware(nextHandler)

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			req = tt.setupContext(req)
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.Equal(t, tt.expectHandler, handlerCalled)

			repo.AssertExpectations(t)
		})
	}
}

func TestRequireAllPermissions(t *testing.T) {
	tests := []struct {
		name           string
		permissions    []PermissionCheck
		setupMock      func(*MockRepository)
		setupContext   func(*http.Request) *http.Request
		expectedStatus int
		expectHandler  bool
	}{
		{
			name: "success - has all permissions",
			permissions: []PermissionCheck{
				{Resource: "workflow", Action: "read"},
				{Resource: "workflow", Action: "update"},
			},
			setupMock: func(repo *MockRepository) {
				repo.On("HasPermission", mock.Anything, "user-123", "tenant-123", "workflow", "read").
					Return(true, nil)
				repo.On("HasPermission", mock.Anything, "user-123", "tenant-123", "workflow", "update").
					Return(true, nil)
			},
			setupContext: func(r *http.Request) *http.Request {
				ctx := context.WithValue(r.Context(), "user_id", "user-123")
				ctx = context.WithValue(ctx, "tenant_id", "tenant-123")
				return r.WithContext(ctx)
			},
			expectedStatus: http.StatusOK,
			expectHandler:  true,
		},
		{
			name: "permission denied - missing one",
			permissions: []PermissionCheck{
				{Resource: "workflow", Action: "read"},
				{Resource: "workflow", Action: "delete"},
			},
			setupMock: func(repo *MockRepository) {
				repo.On("HasPermission", mock.Anything, "user-123", "tenant-123", "workflow", "read").
					Return(true, nil)
				repo.On("HasPermission", mock.Anything, "user-123", "tenant-123", "workflow", "delete").
					Return(false, nil)
			},
			setupContext: func(r *http.Request) *http.Request {
				ctx := context.WithValue(r.Context(), "user_id", "user-123")
				ctx = context.WithValue(ctx, "tenant_id", "tenant-123")
				return r.WithContext(ctx)
			},
			expectedStatus: http.StatusForbidden,
			expectHandler:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := new(MockRepository)
			tt.setupMock(repo)

			handlerCalled := false
			nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				handlerCalled = true
				w.WriteHeader(http.StatusOK)
			})

			middleware := RequireAllPermissions(repo, tt.permissions...)
			handler := middleware(nextHandler)

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			req = tt.setupContext(req)
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.Equal(t, tt.expectHandler, handlerCalled)

			repo.AssertExpectations(t)
		})
	}
}

func TestRequireRole(t *testing.T) {
	tests := []struct {
		name           string
		roleName       string
		setupMock      func(*MockRepository)
		setupContext   func(*http.Request) *http.Request
		expectedStatus int
		expectHandler  bool
	}{
		{
			name:     "success - has role",
			roleName: "admin",
			setupMock: func(repo *MockRepository) {
				roles := []*Role{
					{ID: "role-1", Name: "admin"},
				}
				repo.On("GetUserRoles", mock.Anything, "user-123", "tenant-123").
					Return(roles, nil)
			},
			setupContext: func(r *http.Request) *http.Request {
				ctx := context.WithValue(r.Context(), "user_id", "user-123")
				ctx = context.WithValue(ctx, "tenant_id", "tenant-123")
				return r.WithContext(ctx)
			},
			expectedStatus: http.StatusOK,
			expectHandler:  true,
		},
		{
			name:     "permission denied - does not have role",
			roleName: "admin",
			setupMock: func(repo *MockRepository) {
				roles := []*Role{
					{ID: "role-1", Name: "viewer"},
				}
				repo.On("GetUserRoles", mock.Anything, "user-123", "tenant-123").
					Return(roles, nil)
			},
			setupContext: func(r *http.Request) *http.Request {
				ctx := context.WithValue(r.Context(), "user_id", "user-123")
				ctx = context.WithValue(ctx, "tenant_id", "tenant-123")
				return r.WithContext(ctx)
			},
			expectedStatus: http.StatusForbidden,
			expectHandler:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := new(MockRepository)
			tt.setupMock(repo)

			handlerCalled := false
			nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				handlerCalled = true
				w.WriteHeader(http.StatusOK)
			})

			middleware := RequireRole(repo, tt.roleName)
			handler := middleware(nextHandler)

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			req = tt.setupContext(req)
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.Equal(t, tt.expectHandler, handlerCalled)

			repo.AssertExpectations(t)
		})
	}
}
