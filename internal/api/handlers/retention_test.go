package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/gorax/gorax/internal/api/middleware"
	"github.com/gorax/gorax/internal/retention"
	"github.com/gorax/gorax/internal/tenant"
)

// MockRetentionService is a mock implementation of RetentionService for testing
type MockRetentionService struct {
	mock.Mock
}

func (m *MockRetentionService) GetRetentionPolicy(ctx context.Context, tenantID string) (*retention.RetentionPolicy, error) {
	args := m.Called(ctx, tenantID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*retention.RetentionPolicy), args.Error(1)
}

func (m *MockRetentionService) CleanupOldExecutions(ctx context.Context, tenantID string) (*retention.CleanupResult, error) {
	args := m.Called(ctx, tenantID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*retention.CleanupResult), args.Error(1)
}

func (m *MockRetentionService) CleanupAllTenants(ctx context.Context) (*retention.CleanupResult, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*retention.CleanupResult), args.Error(1)
}

// MockRetentionRepository is a mock implementation of RetentionRepository for testing
type MockRetentionRepository struct {
	mock.Mock
}

func (m *MockRetentionRepository) SetRetentionPolicy(ctx context.Context, tenantID string, retentionDays int, enabled bool) error {
	args := m.Called(ctx, tenantID, retentionDays, enabled)
	return args.Error(0)
}

func newTestRetentionHandler() (*RetentionHandler, *MockRetentionService, *MockRetentionRepository) {
	mockService := new(MockRetentionService)
	mockRepo := new(MockRetentionRepository)
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	handler := NewRetentionHandler(mockService, mockRepo, logger)
	return handler, mockService, mockRepo
}

func addRetentionContext(req *http.Request, tenantID string) *http.Request {
	t := &tenant.Tenant{
		ID:     tenantID,
		Status: "active",
	}
	ctx := context.WithValue(req.Context(), middleware.TenantContextKey, t)
	return req.WithContext(ctx)
}

func addRetentionURLParams(req *http.Request, params map[string]string) *http.Request {
	rctx := chi.NewRouteContext()
	for key, value := range params {
		rctx.URLParams.Add(key, value)
	}
	return req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
}

// Test fixtures
func createTestRetentionPolicy(tenantID string) *retention.RetentionPolicy {
	return &retention.RetentionPolicy{
		TenantID:      tenantID,
		RetentionDays: 90,
		Enabled:       true,
	}
}

func createTestCleanupResult() *retention.CleanupResult {
	return &retention.CleanupResult{
		ExecutionsDeleted:     10,
		StepExecutionsDeleted: 50,
		ExecutionsArchived:    10,
		BatchesProcessed:      1,
	}
}

// ============================================================================
// GetPolicy Handler Tests
// ============================================================================

func TestRetentionHandler_GetPolicy(t *testing.T) {
	tests := []struct {
		name           string
		tenantID       string
		setupMock      func(*MockRetentionService, *MockRetentionRepository)
		expectedStatus int
		expectedBody   string
		checkResponse  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name:     "successful get policy",
			tenantID: "tenant-123",
			setupMock: func(ms *MockRetentionService, mr *MockRetentionRepository) {
				ms.On("GetRetentionPolicy", mock.Anything, "tenant-123").
					Return(createTestRetentionPolicy("tenant-123"), nil)
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, rr *httptest.ResponseRecorder) {
				var resp map[string]interface{}
				err := json.Unmarshal(rr.Body.Bytes(), &resp)
				require.NoError(t, err)
				data := resp["data"].(map[string]interface{})
				assert.Equal(t, "tenant-123", data["tenant_id"])
				assert.Equal(t, float64(90), data["retention_days"])
				assert.True(t, data["retention_enabled"].(bool))
			},
		},
		{
			name:     "missing tenant context",
			tenantID: "", // Empty tenant ID
			setupMock: func(ms *MockRetentionService, mr *MockRetentionRepository) {
				// No mock setup needed - should fail before service call
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "tenant context required",
		},
		{
			name:     "service error",
			tenantID: "tenant-123",
			setupMock: func(ms *MockRetentionService, mr *MockRetentionRepository) {
				ms.On("GetRetentionPolicy", mock.Anything, "tenant-123").
					Return(nil, errors.New("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "failed to get retention policy",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, mockService, mockRepo := newTestRetentionHandler()
			tt.setupMock(mockService, mockRepo)

			req := httptest.NewRequest(http.MethodGet, "/api/v1/retention/policy", nil)
			if tt.tenantID != "" {
				req = addRetentionContext(req, tt.tenantID)
			}

			rr := httptest.NewRecorder()
			handler.GetPolicy(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			if tt.expectedBody != "" {
				assert.Contains(t, rr.Body.String(), tt.expectedBody)
			}
			if tt.checkResponse != nil {
				tt.checkResponse(t, rr)
			}
			mockService.AssertExpectations(t)
			mockRepo.AssertExpectations(t)
		})
	}
}

// ============================================================================
// UpdatePolicy Handler Tests
// ============================================================================

func TestRetentionHandler_UpdatePolicy(t *testing.T) {
	tests := []struct {
		name           string
		tenantID       string
		body           interface{}
		setupMock      func(*MockRetentionService, *MockRetentionRepository)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:     "successful update",
			tenantID: "tenant-123",
			body: UpdateRetentionPolicyInput{
				RetentionDays: 30,
				Enabled:       true,
			},
			setupMock: func(ms *MockRetentionService, mr *MockRetentionRepository) {
				mr.On("SetRetentionPolicy", mock.Anything, "tenant-123", 30, true).Return(nil)
				ms.On("GetRetentionPolicy", mock.Anything, "tenant-123").
					Return(&retention.RetentionPolicy{
						TenantID:      "tenant-123",
						RetentionDays: 30,
						Enabled:       true,
					}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:     "missing tenant context",
			tenantID: "",
			body: UpdateRetentionPolicyInput{
				RetentionDays: 30,
				Enabled:       true,
			},
			setupMock:      func(ms *MockRetentionService, mr *MockRetentionRepository) {},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "tenant context required",
		},
		{
			name:           "invalid request body",
			tenantID:       "tenant-123",
			body:           "invalid json",
			setupMock:      func(ms *MockRetentionService, mr *MockRetentionRepository) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "invalid request body",
		},
		{
			name:     "retention days too low",
			tenantID: "tenant-123",
			body: UpdateRetentionPolicyInput{
				RetentionDays: 3, // Less than MinRetentionDays (7)
				Enabled:       true,
			},
			setupMock:      func(ms *MockRetentionService, mr *MockRetentionRepository) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "retention days must be at least 7",
		},
		{
			name:     "retention days too high",
			tenantID: "tenant-123",
			body: UpdateRetentionPolicyInput{
				RetentionDays: 5000, // More than MaxRetentionDays (3650)
				Enabled:       true,
			},
			setupMock:      func(ms *MockRetentionService, mr *MockRetentionRepository) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "retention days must be at most 3650",
		},
		{
			name:     "tenant not found",
			tenantID: "tenant-123",
			body: UpdateRetentionPolicyInput{
				RetentionDays: 30,
				Enabled:       true,
			},
			setupMock: func(ms *MockRetentionService, mr *MockRetentionRepository) {
				mr.On("SetRetentionPolicy", mock.Anything, "tenant-123", 30, true).
					Return(retention.ErrNotFound)
			},
			expectedStatus: http.StatusNotFound,
			expectedBody:   "tenant not found",
		},
		{
			name:     "repository error",
			tenantID: "tenant-123",
			body: UpdateRetentionPolicyInput{
				RetentionDays: 30,
				Enabled:       true,
			},
			setupMock: func(ms *MockRetentionService, mr *MockRetentionRepository) {
				mr.On("SetRetentionPolicy", mock.Anything, "tenant-123", 30, true).
					Return(errors.New("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "failed to update retention policy",
		},
		{
			name:     "service error on get after update",
			tenantID: "tenant-123",
			body: UpdateRetentionPolicyInput{
				RetentionDays: 30,
				Enabled:       true,
			},
			setupMock: func(ms *MockRetentionService, mr *MockRetentionRepository) {
				mr.On("SetRetentionPolicy", mock.Anything, "tenant-123", 30, true).Return(nil)
				ms.On("GetRetentionPolicy", mock.Anything, "tenant-123").
					Return(nil, errors.New("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "failed to get updated retention policy",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, mockService, mockRepo := newTestRetentionHandler()
			tt.setupMock(mockService, mockRepo)

			var body []byte
			var err error
			switch v := tt.body.(type) {
			case string:
				body = []byte(v)
			default:
				body, err = json.Marshal(tt.body)
				require.NoError(t, err)
			}

			req := httptest.NewRequest(http.MethodPut, "/api/v1/retention/policy", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			if tt.tenantID != "" {
				req = addRetentionContext(req, tt.tenantID)
			}

			rr := httptest.NewRecorder()
			handler.UpdatePolicy(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			if tt.expectedBody != "" {
				assert.Contains(t, rr.Body.String(), tt.expectedBody)
			}
			mockService.AssertExpectations(t)
			mockRepo.AssertExpectations(t)
		})
	}
}

// ============================================================================
// TriggerCleanup Handler Tests
// ============================================================================

func TestRetentionHandler_TriggerCleanup(t *testing.T) {
	tests := []struct {
		name           string
		tenantID       string
		setupMock      func(*MockRetentionService, *MockRetentionRepository)
		expectedStatus int
		expectedBody   string
		checkResponse  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name:     "successful cleanup",
			tenantID: "tenant-123",
			setupMock: func(ms *MockRetentionService, mr *MockRetentionRepository) {
				ms.On("CleanupOldExecutions", mock.Anything, "tenant-123").
					Return(createTestCleanupResult(), nil)
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, rr *httptest.ResponseRecorder) {
				var resp map[string]interface{}
				err := json.Unmarshal(rr.Body.Bytes(), &resp)
				require.NoError(t, err)
				data := resp["data"].(map[string]interface{})
				assert.Equal(t, float64(10), data["executions_deleted"])
				assert.Equal(t, float64(50), data["step_executions_deleted"])
				assert.Equal(t, float64(10), data["executions_archived"])
			},
		},
		{
			name:     "missing tenant context",
			tenantID: "",
			setupMock: func(ms *MockRetentionService, mr *MockRetentionRepository) {
				// No mock setup needed
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "tenant context required",
		},
		{
			name:     "cleanup error",
			tenantID: "tenant-123",
			setupMock: func(ms *MockRetentionService, mr *MockRetentionRepository) {
				ms.On("CleanupOldExecutions", mock.Anything, "tenant-123").
					Return(nil, errors.New("cleanup failed"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "cleanup failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, mockService, mockRepo := newTestRetentionHandler()
			tt.setupMock(mockService, mockRepo)

			req := httptest.NewRequest(http.MethodPost, "/api/v1/retention/cleanup", nil)
			if tt.tenantID != "" {
				req = addRetentionContext(req, tt.tenantID)
			}

			rr := httptest.NewRecorder()
			handler.TriggerCleanup(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			if tt.expectedBody != "" {
				assert.Contains(t, rr.Body.String(), tt.expectedBody)
			}
			if tt.checkResponse != nil {
				tt.checkResponse(t, rr)
			}
			mockService.AssertExpectations(t)
			mockRepo.AssertExpectations(t)
		})
	}
}

// ============================================================================
// AdminGetPolicy Handler Tests
// ============================================================================

func TestRetentionHandler_AdminGetPolicy(t *testing.T) {
	tests := []struct {
		name           string
		tenantID       string
		setupMock      func(*MockRetentionService, *MockRetentionRepository)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:     "successful get policy",
			tenantID: "tenant-123",
			setupMock: func(ms *MockRetentionService, mr *MockRetentionRepository) {
				ms.On("GetRetentionPolicy", mock.Anything, "tenant-123").
					Return(createTestRetentionPolicy("tenant-123"), nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:     "missing tenant ID parameter",
			tenantID: "",
			setupMock: func(ms *MockRetentionService, mr *MockRetentionRepository) {
				// No mock setup needed
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "tenant ID is required",
		},
		{
			name:     "tenant not found",
			tenantID: "nonexistent",
			setupMock: func(ms *MockRetentionService, mr *MockRetentionRepository) {
				ms.On("GetRetentionPolicy", mock.Anything, "nonexistent").
					Return(nil, retention.ErrNotFound)
			},
			expectedStatus: http.StatusNotFound,
			expectedBody:   "tenant not found",
		},
		{
			name:     "service error",
			tenantID: "tenant-123",
			setupMock: func(ms *MockRetentionService, mr *MockRetentionRepository) {
				ms.On("GetRetentionPolicy", mock.Anything, "tenant-123").
					Return(nil, errors.New("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "failed to get retention policy",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, mockService, mockRepo := newTestRetentionHandler()
			tt.setupMock(mockService, mockRepo)

			req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/tenants/"+tt.tenantID+"/retention", nil)
			if tt.tenantID != "" {
				req = addRetentionURLParams(req, map[string]string{"tenantID": tt.tenantID})
			}

			rr := httptest.NewRecorder()
			handler.AdminGetPolicy(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			if tt.expectedBody != "" {
				assert.Contains(t, rr.Body.String(), tt.expectedBody)
			}
			mockService.AssertExpectations(t)
			mockRepo.AssertExpectations(t)
		})
	}
}

// ============================================================================
// AdminUpdatePolicy Handler Tests
// ============================================================================

func TestRetentionHandler_AdminUpdatePolicy(t *testing.T) {
	tests := []struct {
		name           string
		tenantID       string
		body           interface{}
		setupMock      func(*MockRetentionService, *MockRetentionRepository)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:     "successful update",
			tenantID: "tenant-123",
			body: UpdateRetentionPolicyInput{
				RetentionDays: 60,
				Enabled:       false,
			},
			setupMock: func(ms *MockRetentionService, mr *MockRetentionRepository) {
				mr.On("SetRetentionPolicy", mock.Anything, "tenant-123", 60, false).Return(nil)
				ms.On("GetRetentionPolicy", mock.Anything, "tenant-123").
					Return(&retention.RetentionPolicy{
						TenantID:      "tenant-123",
						RetentionDays: 60,
						Enabled:       false,
					}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:     "missing tenant ID parameter",
			tenantID: "",
			body: UpdateRetentionPolicyInput{
				RetentionDays: 60,
				Enabled:       true,
			},
			setupMock:      func(ms *MockRetentionService, mr *MockRetentionRepository) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "tenant ID is required",
		},
		{
			name:           "invalid request body",
			tenantID:       "tenant-123",
			body:           "invalid json",
			setupMock:      func(ms *MockRetentionService, mr *MockRetentionRepository) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "invalid request body",
		},
		{
			name:     "retention days too low",
			tenantID: "tenant-123",
			body: UpdateRetentionPolicyInput{
				RetentionDays: 5,
				Enabled:       true,
			},
			setupMock:      func(ms *MockRetentionService, mr *MockRetentionRepository) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "retention days must be at least 7",
		},
		{
			name:     "retention days too high",
			tenantID: "tenant-123",
			body: UpdateRetentionPolicyInput{
				RetentionDays: 4000,
				Enabled:       true,
			},
			setupMock:      func(ms *MockRetentionService, mr *MockRetentionRepository) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "retention days must be at most 3650",
		},
		{
			name:     "tenant not found",
			tenantID: "nonexistent",
			body: UpdateRetentionPolicyInput{
				RetentionDays: 60,
				Enabled:       true,
			},
			setupMock: func(ms *MockRetentionService, mr *MockRetentionRepository) {
				mr.On("SetRetentionPolicy", mock.Anything, "nonexistent", 60, true).
					Return(retention.ErrNotFound)
			},
			expectedStatus: http.StatusNotFound,
			expectedBody:   "tenant not found",
		},
		{
			name:     "repository error",
			tenantID: "tenant-123",
			body: UpdateRetentionPolicyInput{
				RetentionDays: 60,
				Enabled:       true,
			},
			setupMock: func(ms *MockRetentionService, mr *MockRetentionRepository) {
				mr.On("SetRetentionPolicy", mock.Anything, "tenant-123", 60, true).
					Return(errors.New("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "failed to update retention policy",
		},
		{
			name:     "service error on get after update",
			tenantID: "tenant-123",
			body: UpdateRetentionPolicyInput{
				RetentionDays: 60,
				Enabled:       true,
			},
			setupMock: func(ms *MockRetentionService, mr *MockRetentionRepository) {
				mr.On("SetRetentionPolicy", mock.Anything, "tenant-123", 60, true).Return(nil)
				ms.On("GetRetentionPolicy", mock.Anything, "tenant-123").
					Return(nil, errors.New("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "failed to get updated retention policy",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, mockService, mockRepo := newTestRetentionHandler()
			tt.setupMock(mockService, mockRepo)

			var body []byte
			var err error
			switch v := tt.body.(type) {
			case string:
				body = []byte(v)
			default:
				body, err = json.Marshal(tt.body)
				require.NoError(t, err)
			}

			req := httptest.NewRequest(http.MethodPut, "/api/v1/admin/tenants/"+tt.tenantID+"/retention", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			if tt.tenantID != "" {
				req = addRetentionURLParams(req, map[string]string{"tenantID": tt.tenantID})
			}

			rr := httptest.NewRecorder()
			handler.AdminUpdatePolicy(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			if tt.expectedBody != "" {
				assert.Contains(t, rr.Body.String(), tt.expectedBody)
			}
			mockService.AssertExpectations(t)
			mockRepo.AssertExpectations(t)
		})
	}
}

// ============================================================================
// AdminTriggerCleanup Handler Tests
// ============================================================================

func TestRetentionHandler_AdminTriggerCleanup(t *testing.T) {
	tests := []struct {
		name           string
		tenantID       string
		setupMock      func(*MockRetentionService, *MockRetentionRepository)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:     "successful cleanup",
			tenantID: "tenant-123",
			setupMock: func(ms *MockRetentionService, mr *MockRetentionRepository) {
				ms.On("CleanupOldExecutions", mock.Anything, "tenant-123").
					Return(createTestCleanupResult(), nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:     "missing tenant ID parameter",
			tenantID: "",
			setupMock: func(ms *MockRetentionService, mr *MockRetentionRepository) {
				// No mock setup needed
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "tenant ID is required",
		},
		{
			name:     "cleanup error",
			tenantID: "tenant-123",
			setupMock: func(ms *MockRetentionService, mr *MockRetentionRepository) {
				ms.On("CleanupOldExecutions", mock.Anything, "tenant-123").
					Return(nil, errors.New("cleanup failed"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "cleanup failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, mockService, mockRepo := newTestRetentionHandler()
			tt.setupMock(mockService, mockRepo)

			req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/tenants/"+tt.tenantID+"/retention/cleanup", nil)
			if tt.tenantID != "" {
				req = addRetentionURLParams(req, map[string]string{"tenantID": tt.tenantID})
			}

			rr := httptest.NewRecorder()
			handler.AdminTriggerCleanup(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			if tt.expectedBody != "" {
				assert.Contains(t, rr.Body.String(), tt.expectedBody)
			}
			mockService.AssertExpectations(t)
			mockRepo.AssertExpectations(t)
		})
	}
}

// ============================================================================
// AdminTriggerAllTenantsCleanup Handler Tests
// ============================================================================

func TestRetentionHandler_AdminTriggerAllTenantsCleanup(t *testing.T) {
	tests := []struct {
		name           string
		setupMock      func(*MockRetentionService, *MockRetentionRepository)
		expectedStatus int
		expectedBody   string
		checkResponse  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name: "successful cleanup all tenants",
			setupMock: func(ms *MockRetentionService, mr *MockRetentionRepository) {
				ms.On("CleanupAllTenants", mock.Anything).
					Return(&retention.CleanupResult{
						ExecutionsDeleted:     100,
						StepExecutionsDeleted: 500,
						ExecutionsArchived:    100,
						BatchesProcessed:      10,
					}, nil)
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, rr *httptest.ResponseRecorder) {
				var resp map[string]interface{}
				err := json.Unmarshal(rr.Body.Bytes(), &resp)
				require.NoError(t, err)
				data := resp["data"].(map[string]interface{})
				assert.Equal(t, float64(100), data["executions_deleted"])
				assert.Equal(t, float64(500), data["step_executions_deleted"])
			},
		},
		{
			name: "cleanup error",
			setupMock: func(ms *MockRetentionService, mr *MockRetentionRepository) {
				ms.On("CleanupAllTenants", mock.Anything).
					Return(nil, errors.New("cleanup failed for all tenants"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "cleanup failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, mockService, mockRepo := newTestRetentionHandler()
			tt.setupMock(mockService, mockRepo)

			req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/retention/cleanup-all", nil)

			rr := httptest.NewRecorder()
			handler.AdminTriggerAllTenantsCleanup(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			if tt.expectedBody != "" {
				assert.Contains(t, rr.Body.String(), tt.expectedBody)
			}
			if tt.checkResponse != nil {
				tt.checkResponse(t, rr)
			}
			mockService.AssertExpectations(t)
			mockRepo.AssertExpectations(t)
		})
	}
}
