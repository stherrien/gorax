package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/gorax/gorax/internal/errortracking"
	"github.com/gorax/gorax/internal/tenant"
)

func TestSentryMiddleware_Success(t *testing.T) {
	tracker := &mockTracker{
		breadcrumbs: []errortracking.Breadcrumb{},
	}

	handler := SentryMiddleware(tracker)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	}))

	req := httptest.NewRequest("GET", "/api/workflows", nil)
	req.Header.Set("X-Request-ID", "req-123")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "success", w.Body.String())

	// Should have added breadcrumb
	assert.Len(t, tracker.breadcrumbs, 1)
	assert.Equal(t, "http", tracker.breadcrumbs[0].Type)
	assert.Equal(t, "GET", tracker.breadcrumbs[0].Data["method"])
}

func TestSentryMiddleware_WithUserContext(t *testing.T) {
	tracker := &mockTracker{
		breadcrumbs: []errortracking.Breadcrumb{},
		users:       []errortracking.User{},
	}

	handler := SentryMiddleware(tracker)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/api/workflows", nil)
	req.RemoteAddr = "192.168.1.1:1234"

	// Add user to context
	user := &User{
		ID:       "user-123",
		Email:    "test@example.com",
		TenantID: "tenant-123",
	}
	ctx := context.WithValue(req.Context(), UserContextKey, user)
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Should have set user
	assert.Len(t, tracker.users, 1)
	assert.Equal(t, "user-123", tracker.users[0].ID)
	assert.Equal(t, "test@example.com", tracker.users[0].Email)
	assert.Equal(t, "192.168.1.1", tracker.users[0].IPAddress)
}

func TestSentryMiddleware_WithTenantContext(t *testing.T) {
	tracker := &mockTracker{
		breadcrumbs: []errortracking.Breadcrumb{},
	}

	handler := SentryMiddleware(tracker)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/api/workflows", nil)

	// Add tenant to context
	ten := &tenant.Tenant{
		ID:   "tenant-123",
		Name: "Test Tenant",
	}
	ctx := context.WithValue(req.Context(), TenantContextKey, ten)
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Breadcrumb should include tenant info
	assert.Len(t, tracker.breadcrumbs, 1)
	assert.Equal(t, "tenant-123", tracker.breadcrumbs[0].Data["tenant_id"])
}

func TestSentryMiddleware_PanicRecovery(t *testing.T) {
	tracker := &mockTracker{
		breadcrumbs: []errortracking.Breadcrumb{},
		errors:      []error{},
	}

	handler := SentryMiddleware(tracker)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("test panic")
	}))

	req := httptest.NewRequest("GET", "/api/workflows", nil)
	w := httptest.NewRecorder()

	// Should not panic
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	// Should have captured panic as error
	assert.Len(t, tracker.errors, 1)
	assert.Contains(t, tracker.errors[0].Error(), "test panic")
}

func TestSentryMiddleware_PanicRecoveryWithContext(t *testing.T) {
	tracker := &mockTracker{
		breadcrumbs: []errortracking.Breadcrumb{},
		errors:      []error{},
		contexts:    []context.Context{},
	}

	handler := SentryMiddleware(tracker)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("test panic with context")
	}))

	req := httptest.NewRequest("GET", "/api/workflows", nil)

	// Add user and tenant to context
	user := &User{ID: "user-123"}
	ten := &tenant.Tenant{ID: "tenant-456"}
	ctx := context.WithValue(req.Context(), UserContextKey, user)
	ctx = context.WithValue(ctx, TenantContextKey, ten)
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Len(t, tracker.errors, 1)
}

func TestEnrichSentryContext(t *testing.T) {
	tests := []struct {
		name           string
		req            *http.Request
		setupContext   func(*http.Request) *http.Request
		expectedTags   map[string]string
		expectedExtras map[string]interface{}
	}{
		{
			name: "basic request context",
			req:  httptest.NewRequest("GET", "/api/workflows", nil),
			setupContext: func(r *http.Request) *http.Request {
				return r
			},
			expectedTags: map[string]string{
				"http.method": "GET",
				"http.path":   "/api/workflows",
			},
			expectedExtras: map[string]interface{}{
				"http.url": "/api/workflows",
			},
		},
		{
			name: "with user context",
			req:  httptest.NewRequest("POST", "/api/executions", nil),
			setupContext: func(r *http.Request) *http.Request {
				user := &User{
					ID:       "user-123",
					Email:    "test@example.com",
					TenantID: "tenant-456",
				}
				ctx := context.WithValue(r.Context(), UserContextKey, user)
				return r.WithContext(ctx)
			},
			expectedTags: map[string]string{
				"http.method": "POST",
				"http.path":   "/api/executions",
				"user.id":     "user-123",
				"tenant.id":   "tenant-456",
			},
			expectedExtras: map[string]interface{}{
				"user.email": "test@example.com",
			},
		},
		{
			name: "with tenant context",
			req:  httptest.NewRequest("GET", "/api/workflows", nil),
			setupContext: func(r *http.Request) *http.Request {
				ten := &tenant.Tenant{
					ID:   "tenant-789",
					Name: "Test Tenant",
				}
				ctx := context.WithValue(r.Context(), TenantContextKey, ten)
				return r.WithContext(ctx)
			},
			expectedTags: map[string]string{
				"tenant.id": "tenant-789",
			},
			expectedExtras: map[string]interface{}{
				"tenant.name": "Test Tenant",
			},
		},
		{
			name: "with request ID",
			req:  httptest.NewRequest("GET", "/api/workflows", nil),
			setupContext: func(r *http.Request) *http.Request {
				r.Header.Set("X-Request-ID", "req-123")
				return r
			},
			expectedTags: map[string]string{
				"request.id": "req-123",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := tt.setupContext(tt.req)

			tags, extras := enrichSentryContext(req)

			for key, expectedValue := range tt.expectedTags {
				assert.Equal(t, expectedValue, tags[key], "tag %s mismatch", key)
			}

			for key, expectedValue := range tt.expectedExtras {
				assert.Equal(t, expectedValue, extras[key], "extra %s mismatch", key)
			}
		})
	}
}

func TestExtractIPAddress(t *testing.T) {
	tests := []struct {
		name         string
		remoteAddr   string
		forwardedFor string
		realIP       string
		expectedIP   string
	}{
		{
			name:       "from RemoteAddr",
			remoteAddr: "192.168.1.1:1234",
			expectedIP: "192.168.1.1",
		},
		{
			name:         "from X-Forwarded-For",
			remoteAddr:   "10.0.0.1:5678",
			forwardedFor: "203.0.113.1, 198.51.100.1",
			expectedIP:   "203.0.113.1",
		},
		{
			name:       "from X-Real-IP",
			remoteAddr: "10.0.0.1:5678",
			realIP:     "203.0.113.1",
			expectedIP: "203.0.113.1",
		},
		{
			name:       "IPv6 address",
			remoteAddr: "[2001:db8::1]:8080",
			expectedIP: "2001:db8::1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/", nil)
			req.RemoteAddr = tt.remoteAddr

			if tt.forwardedFor != "" {
				req.Header.Set("X-Forwarded-For", tt.forwardedFor)
			}
			if tt.realIP != "" {
				req.Header.Set("X-Real-IP", tt.realIP)
			}

			ip := extractIPAddress(req)
			assert.Equal(t, tt.expectedIP, ip)
		})
	}
}

// Mock tracker for testing

type mockTracker struct {
	breadcrumbs []errortracking.Breadcrumb
	users       []errortracking.User
	contexts    []context.Context
	errors      []error
}

func (m *mockTracker) AddBreadcrumb(ctx context.Context, breadcrumb errortracking.Breadcrumb) {
	m.breadcrumbs = append(m.breadcrumbs, breadcrumb)
	m.contexts = append(m.contexts, ctx)
}

func (m *mockTracker) SetUser(ctx context.Context, user errortracking.User) {
	m.users = append(m.users, user)
	m.contexts = append(m.contexts, ctx)
}

func (m *mockTracker) CaptureError(ctx context.Context, err error) string {
	m.errors = append(m.errors, err)
	m.contexts = append(m.contexts, ctx)
	return "mock-event-id"
}

func (m *mockTracker) CaptureErrorWithTags(ctx context.Context, err error, tags map[string]string) string {
	m.errors = append(m.errors, err)
	m.contexts = append(m.contexts, ctx)
	return "mock-event-id"
}

func (m *mockTracker) CaptureMessage(ctx context.Context, message string, level errortracking.Level) string {
	return "mock-event-id"
}

func (m *mockTracker) WithScope(ctx context.Context, f func(*errortracking.Scope)) {
	// No-op for testing
}

func (m *mockTracker) Flush(timeout interface{}) {
	// No-op for testing
}

func (m *mockTracker) Close() {
	// No-op for testing
}
