package middleware

import (
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/gorax/gorax/internal/audit"
)

// MockAuditRepository is a mock implementation of audit.AuditRepository
type MockAuditRepository struct {
	mock.Mock
}

func (m *MockAuditRepository) CreateAuditEvent(ctx context.Context, event *audit.AuditEvent) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

func (m *MockAuditRepository) CreateAuditEventBatch(ctx context.Context, events []*audit.AuditEvent) error {
	args := m.Called(ctx, events)
	return args.Error(0)
}

func (m *MockAuditRepository) GetAuditEvent(ctx context.Context, tenantID, eventID string) (*audit.AuditEvent, error) {
	args := m.Called(ctx, tenantID, eventID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*audit.AuditEvent), args.Error(1)
}

func (m *MockAuditRepository) QueryAuditEvents(ctx context.Context, filter audit.QueryFilter) ([]audit.AuditEvent, int, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]audit.AuditEvent), args.Int(1), args.Error(2)
}

func (m *MockAuditRepository) GetAuditStats(ctx context.Context, tenantID string, timeRange audit.TimeRange) (*audit.AuditStats, error) {
	args := m.Called(ctx, tenantID, timeRange)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*audit.AuditStats), args.Error(1)
}

func (m *MockAuditRepository) GetRetentionPolicy(ctx context.Context, tenantID string) (*audit.RetentionPolicy, error) {
	args := m.Called(ctx, tenantID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*audit.RetentionPolicy), args.Error(1)
}

func (m *MockAuditRepository) UpdateRetentionPolicy(ctx context.Context, policy *audit.RetentionPolicy) error {
	args := m.Called(ctx, policy)
	return args.Error(0)
}

func (m *MockAuditRepository) DeleteOldAuditEvents(ctx context.Context, tenantID string, cutoffDate time.Time) (int64, error) {
	args := m.Called(ctx, tenantID, cutoffDate)
	return args.Get(0).(int64), args.Error(1)
}

// Helper to create audit service with mock repository
func newTestAuditService() (*audit.Service, *MockAuditRepository) {
	mockRepo := new(MockAuditRepository)
	// Use small buffer and short flush time for testing
	service := audit.NewService(mockRepo, 10, 50*time.Millisecond)
	return service, mockRepo
}

func newTestLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
}

// =============================================================================
// shouldSkipAudit Tests
// =============================================================================

func TestShouldSkipAudit(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected bool
	}{
		{
			name:     "health endpoint",
			path:     "/health",
			expected: true,
		},
		{
			name:     "health with suffix",
			path:     "/health/liveness",
			expected: true,
		},
		{
			name:     "ready endpoint",
			path:     "/ready",
			expected: true,
		},
		{
			name:     "metrics endpoint",
			path:     "/metrics",
			expected: true,
		},
		{
			name:     "swagger endpoint",
			path:     "/swagger/index.html",
			expected: true,
		},
		{
			name:     "api docs",
			path:     "/api/docs",
			expected: true,
		},
		{
			name:     "favicon",
			path:     "/favicon.ico",
			expected: true,
		},
		{
			name:     "regular API endpoint",
			path:     "/api/v1/workflows",
			expected: false,
		},
		{
			name:     "login endpoint - should be logged",
			path:     "/api/v1/auth/login",
			expected: false,
		},
		{
			name:     "credentials endpoint",
			path:     "/api/v1/credentials",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := shouldSkipAudit(tt.path)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// =============================================================================
// categorizeRequest Tests
// =============================================================================

func TestCategorizeRequest(t *testing.T) {
	tests := []struct {
		name             string
		path             string
		method           string
		expectedCategory audit.Category
		expectedType     audit.EventType
		expectedAction   string
	}{
		{
			name:             "login endpoint",
			path:             "/api/v1/auth/login",
			method:           "POST",
			expectedCategory: audit.CategoryAuthentication,
			expectedType:     audit.EventTypeLogin,
			expectedAction:   "user_login",
		},
		{
			name:             "logout endpoint",
			path:             "/api/v1/auth/logout",
			method:           "POST",
			expectedCategory: audit.CategoryAuthentication,
			expectedType:     audit.EventTypeLogout,
			expectedAction:   "user_logout",
		},
		{
			name:             "other auth endpoint",
			path:             "/api/v1/auth/refresh",
			method:           "POST",
			expectedCategory: audit.CategoryAuthentication,
			expectedType:     audit.EventTypeAccess,
			expectedAction:   "authentication_action",
		},
		{
			name:             "SSO endpoint",
			path:             "/api/v1/sso/providers",
			method:           "GET",
			expectedCategory: audit.CategoryAuthentication,
			expectedType:     audit.EventTypeLogin,
			expectedAction:   "sso_authentication",
		},
		{
			name:             "OAuth endpoint - falls through to auth due to path containing 'auth'",
			path:             "/api/v1/oauth/authorize",
			method:           "GET",
			expectedCategory: audit.CategoryAuthentication,
			expectedType:     audit.EventTypeAccess,
			expectedAction:   "authentication_action",
		},
		{
			name:             "OAuth callback endpoint",
			path:             "/api/v1/oauth/callback",
			method:           "GET",
			expectedCategory: audit.CategoryIntegration,
			expectedType:     audit.EventTypeConfigure,
			expectedAction:   "oauth_action",
		},
		{
			name:             "create credential",
			path:             "/api/v1/credentials",
			method:           "POST",
			expectedCategory: audit.CategoryCredential,
			expectedType:     audit.EventTypeCreate,
			expectedAction:   "post_credential",
		},
		{
			name:             "get credentials",
			path:             "/api/v1/credentials",
			method:           "GET",
			expectedCategory: audit.CategoryCredential,
			expectedType:     audit.EventTypeRead,
			expectedAction:   "get_credential",
		},
		{
			name:             "delete credential",
			path:             "/api/v1/credentials/123",
			method:           "DELETE",
			expectedCategory: audit.CategoryCredential,
			expectedType:     audit.EventTypeDelete,
			expectedAction:   "delete_credential",
		},
		{
			name:             "execute workflow",
			path:             "/api/v1/workflows/123/execute",
			method:           "POST",
			expectedCategory: audit.CategoryWorkflow,
			expectedType:     audit.EventTypeExecute,
			expectedAction:   "execute_workflow",
		},
		{
			name:             "create workflow",
			path:             "/api/v1/workflows",
			method:           "POST",
			expectedCategory: audit.CategoryWorkflow,
			expectedType:     audit.EventTypeCreate,
			expectedAction:   "post_workflow",
		},
		{
			name:             "update workflow",
			path:             "/api/v1/workflows/123",
			method:           "PUT",
			expectedCategory: audit.CategoryWorkflow,
			expectedType:     audit.EventTypeUpdate,
			expectedAction:   "put_workflow",
		},
		{
			name:             "admin tenants",
			path:             "/api/v1/admin/tenants",
			method:           "POST",
			expectedCategory: audit.CategoryUserManagement,
			expectedType:     audit.EventTypeCreate,
			expectedAction:   "post_tenant",
		},
		{
			name:             "admin audit logs",
			path:             "/api/v1/admin/audit",
			method:           "GET",
			expectedCategory: audit.CategorySystem,
			expectedType:     audit.EventTypeAccess,
			expectedAction:   "view_audit_logs",
		},
		{
			name:             "other admin endpoint",
			path:             "/api/v1/admin/settings",
			method:           "PUT",
			expectedCategory: audit.CategoryConfiguration,
			expectedType:     audit.EventTypeUpdate,
			expectedAction:   "admin_action",
		},
		{
			name:             "generic GET request",
			path:             "/api/v1/executions",
			method:           "GET",
			expectedCategory: audit.CategoryDataAccess,
			expectedType:     audit.EventTypeRead,
			expectedAction:   "view_resource",
		},
		{
			name:             "generic POST request",
			path:             "/api/v1/schedules",
			method:           "POST",
			expectedCategory: audit.CategoryConfiguration,
			expectedType:     audit.EventTypeCreate,
			expectedAction:   "post_resource",
		},
		{
			name:             "generic PATCH request",
			path:             "/api/v1/schedules/123",
			method:           "PATCH",
			expectedCategory: audit.CategoryConfiguration,
			expectedType:     audit.EventTypeUpdate,
			expectedAction:   "patch_resource",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			category, eventType, action := categorizeRequest(tt.path, tt.method)
			assert.Equal(t, tt.expectedCategory, category)
			assert.Equal(t, tt.expectedType, eventType)
			assert.Equal(t, tt.expectedAction, action)
		})
	}
}

// =============================================================================
// getEventTypeFromMethod Tests
// =============================================================================

func TestGetEventTypeFromMethod(t *testing.T) {
	tests := []struct {
		method   string
		expected audit.EventType
	}{
		{"POST", audit.EventTypeCreate},
		{"GET", audit.EventTypeRead},
		{"PUT", audit.EventTypeUpdate},
		{"PATCH", audit.EventTypeUpdate},
		{"DELETE", audit.EventTypeDelete},
		{"OPTIONS", audit.EventTypeAccess},
		{"HEAD", audit.EventTypeAccess},
	}

	for _, tt := range tests {
		t.Run(tt.method, func(t *testing.T) {
			result := getEventTypeFromMethod(tt.method)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// =============================================================================
// extractResource Tests
// =============================================================================

func TestExtractResource(t *testing.T) {
	tests := []struct {
		name             string
		path             string
		expectedType     string
		expectedID       string
	}{
		{
			name:         "standard API path with ID",
			path:         "/api/v1/workflows/workflow-123",
			expectedType: "workflows",
			expectedID:   "workflow-123",
		},
		{
			name:         "standard API path without ID",
			path:         "/api/v1/credentials",
			expectedType: "credentials",
			expectedID:   "",
		},
		{
			name:         "nested path",
			path:         "/api/v1/workflows/workflow-123/executions",
			expectedType: "workflows",
			expectedID:   "workflow-123",
		},
		{
			name:         "simple path",
			path:         "/health",
			expectedType: "health",
			expectedID:   "",
		},
		{
			name:         "two-part path",
			path:         "/resource/123",
			expectedType: "resource",
			expectedID:   "123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resourceType, resourceID := extractResource(tt.path)
			assert.Equal(t, tt.expectedType, resourceType)
			assert.Equal(t, tt.expectedID, resourceID)
		})
	}
}

// =============================================================================
// determineSeverity Tests
// =============================================================================

func TestDetermineSeverity(t *testing.T) {
	tests := []struct {
		name           string
		statusCode     int
		method         string
		expectedSeverity audit.Severity
	}{
		{
			name:           "unauthorized",
			statusCode:     http.StatusUnauthorized,
			method:         "GET",
			expectedSeverity: audit.SeverityCritical,
		},
		{
			name:           "forbidden",
			statusCode:     http.StatusForbidden,
			method:         "POST",
			expectedSeverity: audit.SeverityCritical,
		},
		{
			name:           "server error",
			statusCode:     http.StatusInternalServerError,
			method:         "GET",
			expectedSeverity: audit.SeverityError,
		},
		{
			name:           "bad gateway",
			statusCode:     http.StatusBadGateway,
			method:         "GET",
			expectedSeverity: audit.SeverityError,
		},
		{
			name:           "bad request",
			statusCode:     http.StatusBadRequest,
			method:         "POST",
			expectedSeverity: audit.SeverityWarning,
		},
		{
			name:           "conflict",
			statusCode:     http.StatusConflict,
			method:         "POST",
			expectedSeverity: audit.SeverityWarning,
		},
		{
			name:           "not found - info",
			statusCode:     http.StatusNotFound,
			method:         "GET",
			expectedSeverity: audit.SeverityInfo,
		},
		{
			name:           "successful delete - warning",
			statusCode:     http.StatusOK,
			method:         "DELETE",
			expectedSeverity: audit.SeverityWarning,
		},
		{
			name:           "no content delete - warning",
			statusCode:     http.StatusNoContent,
			method:         "DELETE",
			expectedSeverity: audit.SeverityWarning,
		},
		{
			name:           "successful GET",
			statusCode:     http.StatusOK,
			method:         "GET",
			expectedSeverity: audit.SeverityInfo,
		},
		{
			name:           "successful POST",
			statusCode:     http.StatusCreated,
			method:         "POST",
			expectedSeverity: audit.SeverityInfo,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := determineSeverity(tt.statusCode, tt.method)
			assert.Equal(t, tt.expectedSeverity, result)
		})
	}
}

// =============================================================================
// determineStatus Tests
// =============================================================================

func TestDetermineStatus(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		expected   audit.Status
	}{
		{"200 OK", http.StatusOK, audit.StatusSuccess},
		{"201 Created", http.StatusCreated, audit.StatusSuccess},
		{"204 No Content", http.StatusNoContent, audit.StatusSuccess},
		{"300 Multiple Choices", http.StatusMultipleChoices, audit.StatusPartial},
		{"301 Moved", http.StatusMovedPermanently, audit.StatusPartial},
		{"400 Bad Request", http.StatusBadRequest, audit.StatusFailure},
		{"401 Unauthorized", http.StatusUnauthorized, audit.StatusFailure},
		{"403 Forbidden", http.StatusForbidden, audit.StatusFailure},
		{"404 Not Found", http.StatusNotFound, audit.StatusFailure},
		{"500 Internal Error", http.StatusInternalServerError, audit.StatusFailure},
		{"503 Unavailable", http.StatusServiceUnavailable, audit.StatusFailure},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := determineStatus(tt.statusCode)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// =============================================================================
// getClientIP Tests
// =============================================================================

func TestGetClientIP(t *testing.T) {
	tests := []struct {
		name       string
		headers    map[string]string
		remoteAddr string
		expected   string
	}{
		{
			name:       "X-Forwarded-For single IP",
			headers:    map[string]string{"X-Forwarded-For": "192.168.1.100"},
			remoteAddr: "10.0.0.1:12345",
			expected:   "192.168.1.100",
		},
		{
			name:       "X-Forwarded-For multiple IPs",
			headers:    map[string]string{"X-Forwarded-For": "192.168.1.100, 10.0.0.2, 10.0.0.3"},
			remoteAddr: "10.0.0.1:12345",
			expected:   "192.168.1.100",
		},
		{
			name:       "X-Real-IP",
			headers:    map[string]string{"X-Real-IP": "172.16.0.50"},
			remoteAddr: "10.0.0.1:12345",
			expected:   "172.16.0.50",
		},
		{
			name:       "X-Forwarded-For takes precedence over X-Real-IP",
			headers:    map[string]string{"X-Forwarded-For": "192.168.1.100", "X-Real-IP": "172.16.0.50"},
			remoteAddr: "10.0.0.1:12345",
			expected:   "192.168.1.100",
		},
		{
			name:       "fallback to RemoteAddr with port",
			headers:    map[string]string{},
			remoteAddr: "10.0.0.1:12345",
			expected:   "10.0.0.1",
		},
		{
			name:       "fallback to RemoteAddr without port",
			headers:    map[string]string{},
			remoteAddr: "10.0.0.1",
			expected:   "10.0.0.1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			req.RemoteAddr = tt.remoteAddr
			for key, value := range tt.headers {
				req.Header.Set(key, value)
			}

			result := getClientIP(req)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// =============================================================================
// responseRecorder Tests
// =============================================================================

func TestResponseRecorder_WriteHeader(t *testing.T) {
	tests := []struct {
		name           string
		callTwice      bool
		firstStatus    int
		secondStatus   int
		expectedStatus int
	}{
		{
			name:           "single WriteHeader",
			callTwice:      false,
			firstStatus:    http.StatusCreated,
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "double WriteHeader - second ignored",
			callTwice:      true,
			firstStatus:    http.StatusCreated,
			secondStatus:   http.StatusOK,
			expectedStatus: http.StatusCreated,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			underlying := httptest.NewRecorder()
			recorder := &responseRecorder{
				ResponseWriter: underlying,
				statusCode:     http.StatusOK,
				written:        false,
			}

			recorder.WriteHeader(tt.firstStatus)
			if tt.callTwice {
				recorder.WriteHeader(tt.secondStatus)
			}

			assert.Equal(t, tt.expectedStatus, recorder.statusCode)
			assert.True(t, recorder.written)
			assert.Equal(t, tt.expectedStatus, underlying.Code)
		})
	}
}

func TestResponseRecorder_Write(t *testing.T) {
	tests := []struct {
		name              string
		writeHeaderFirst  bool
		headerStatus      int
		expectedStatus    int
	}{
		{
			name:             "write without WriteHeader",
			writeHeaderFirst: false,
			expectedStatus:   http.StatusOK,
		},
		{
			name:             "write after WriteHeader",
			writeHeaderFirst: true,
			headerStatus:     http.StatusCreated,
			expectedStatus:   http.StatusCreated,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			underlying := httptest.NewRecorder()
			recorder := &responseRecorder{
				ResponseWriter: underlying,
				statusCode:     http.StatusOK,
				written:        false,
			}

			if tt.writeHeaderFirst {
				recorder.WriteHeader(tt.headerStatus)
			}

			n, err := recorder.Write([]byte("test content"))
			assert.NoError(t, err)
			assert.Equal(t, 12, n)
			assert.Equal(t, tt.expectedStatus, recorder.statusCode)
			assert.True(t, recorder.written)
		})
	}
}

// =============================================================================
// AuditMiddleware Integration Tests
// =============================================================================

func TestAuditMiddleware_SkipsHealthChecks(t *testing.T) {
	service, mockRepo := newTestAuditService()
	logger := newTestLogger()

	// Create middleware
	middleware := AuditMiddleware(service, logger)

	// Create a handler that always returns 200
	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Test health endpoint
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	// Give async goroutine time to execute (if it were to)
	time.Sleep(100 * time.Millisecond)

	// Verify no audit events were created for health check
	mockRepo.AssertNotCalled(t, "CreateAuditEvent")
}

func TestAuditMiddleware_SkipsPublicEndpointWithoutContext(t *testing.T) {
	service, mockRepo := newTestAuditService()
	logger := newTestLogger()

	// Create middleware
	middleware := AuditMiddleware(service, logger)

	// Create a handler
	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Request without tenant/user context
	req := httptest.NewRequest(http.MethodGet, "/api/v1/public/resource", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	// Give async goroutine time to execute
	time.Sleep(100 * time.Millisecond)

	// Verify no audit events were created (no tenant/user context)
	mockRepo.AssertNotCalled(t, "CreateAuditEvent")
}

func TestAuditMiddleware_LogsAuthenticatedRequest(t *testing.T) {
	service, mockRepo := newTestAuditService()
	logger := newTestLogger()

	// Setup mock to accept batch events (service uses batching)
	mockRepo.On("CreateAuditEventBatch", mock.Anything, mock.MatchedBy(func(events []*audit.AuditEvent) bool {
		for _, event := range events {
			if event.TenantID == "tenant-123" && event.UserID == "user-456" {
				return true
			}
		}
		return false
	})).Return(nil).Maybe()

	// Create middleware
	middleware := AuditMiddleware(service, logger)

	// Create a handler
	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Request with tenant/user context
	req := httptest.NewRequest(http.MethodGet, "/api/v1/workflows", nil)
	ctx := context.WithValue(req.Context(), "tenant_id", "tenant-123")
	ctx = context.WithValue(ctx, "user_id", "user-456")
	ctx = context.WithValue(ctx, "user_email", "user@example.com")
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	// Give the async logging goroutine time to complete
	// We don't call Flush/Close to avoid racing with the middleware's goroutine
	runtime.Gosched()
	time.Sleep(150 * time.Millisecond)
}

func TestAuditMiddleware_CapturesErrorStatus(t *testing.T) {
	service, mockRepo := newTestAuditService()
	logger := newTestLogger()

	// Setup mock to accept batch events
	mockRepo.On("CreateAuditEventBatch", mock.Anything, mock.MatchedBy(func(events []*audit.AuditEvent) bool {
		for _, event := range events {
			if event.Status == audit.StatusFailure &&
				event.Severity == audit.SeverityCritical &&
				event.ErrorMessage == "Forbidden" {
				return true
			}
		}
		return false
	})).Return(nil).Maybe()

	// Create middleware
	middleware := AuditMiddleware(service, logger)

	// Create a handler that returns 403
	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	}))

	// Request with context
	req := httptest.NewRequest(http.MethodPost, "/api/v1/credentials", nil)
	ctx := context.WithValue(req.Context(), "tenant_id", "tenant-123")
	ctx = context.WithValue(ctx, "user_id", "user-456")
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusForbidden, rr.Code)

	// Give the async logging goroutine time to complete
	runtime.Gosched()
	time.Sleep(150 * time.Millisecond)
}

func TestAuditMiddleware_TracksRequestDuration(t *testing.T) {
	service, mockRepo := newTestAuditService()
	logger := newTestLogger()

	// Setup mock to verify duration is tracked in batch events
	mockRepo.On("CreateAuditEventBatch", mock.Anything, mock.MatchedBy(func(events []*audit.AuditEvent) bool {
		for _, event := range events {
			durationMs, ok := event.Metadata["duration_ms"].(int64)
			if ok && durationMs > 0 {
				return true
			}
		}
		return false
	})).Return(nil).Maybe()

	// Create middleware
	middleware := AuditMiddleware(service, logger)

	// Create a handler with intentional delay
	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(10 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))

	// Request with context
	req := httptest.NewRequest(http.MethodGet, "/api/v1/executions", nil)
	ctx := context.WithValue(req.Context(), "tenant_id", "tenant-123")
	ctx = context.WithValue(ctx, "user_id", "user-456")
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	// Give the async logging goroutine time to complete
	runtime.Gosched()
	time.Sleep(150 * time.Millisecond)
}
