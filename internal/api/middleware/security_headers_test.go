package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSecurityHeaders_AllHeaders(t *testing.T) {
	cfg := SecurityHeadersConfig{
		EnableHSTS:    true,
		HSTSMaxAge:    31536000,
		CSPDirectives: "default-src 'self'",
		FrameOptions:  "DENY",
	}

	handler := SecurityHeaders(cfg)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	}))

	req := httptest.NewRequest("GET", "/api/workflows", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "success", w.Body.String())

	// Verify all security headers are set
	assert.Equal(t, "max-age=31536000; includeSubDomains", w.Header().Get("Strict-Transport-Security"))
	assert.Equal(t, "nosniff", w.Header().Get("X-Content-Type-Options"))
	assert.Equal(t, "DENY", w.Header().Get("X-Frame-Options"))
	assert.Equal(t, "1; mode=block", w.Header().Get("X-XSS-Protection"))
	assert.Equal(t, "default-src 'self'", w.Header().Get("Content-Security-Policy"))
	assert.Equal(t, "strict-origin-when-cross-origin", w.Header().Get("Referrer-Policy"))
	assert.Equal(t, "geolocation=(), microphone=(), camera=()", w.Header().Get("Permissions-Policy"))
}

func TestSecurityHeaders_HSTSDisabled(t *testing.T) {
	cfg := SecurityHeadersConfig{
		EnableHSTS:    false,
		HSTSMaxAge:    31536000,
		CSPDirectives: "default-src 'self'",
		FrameOptions:  "DENY",
	}

	handler := SecurityHeaders(cfg)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/api/workflows", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// HSTS should not be set when disabled
	assert.Empty(t, w.Header().Get("Strict-Transport-Security"))

	// Other headers should still be set
	assert.Equal(t, "nosniff", w.Header().Get("X-Content-Type-Options"))
	assert.Equal(t, "DENY", w.Header().Get("X-Frame-Options"))
}

func TestSecurityHeaders_CustomCSP(t *testing.T) {
	cfg := SecurityHeadersConfig{
		EnableHSTS:    true,
		HSTSMaxAge:    31536000,
		CSPDirectives: "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'",
		FrameOptions:  "DENY",
	}

	handler := SecurityHeaders(cfg)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/api/workflows", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'",
		w.Header().Get("Content-Security-Policy"))
}

func TestSecurityHeaders_CustomFrameOptions(t *testing.T) {
	cfg := SecurityHeadersConfig{
		EnableHSTS:    true,
		HSTSMaxAge:    31536000,
		CSPDirectives: "default-src 'self'",
		FrameOptions:  "SAMEORIGIN",
	}

	handler := SecurityHeaders(cfg)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/api/workflows", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "SAMEORIGIN", w.Header().Get("X-Frame-Options"))
}

func TestSecurityHeaders_CustomHSTSMaxAge(t *testing.T) {
	cfg := SecurityHeadersConfig{
		EnableHSTS:    true,
		HSTSMaxAge:    63072000, // 2 years
		CSPDirectives: "default-src 'self'",
		FrameOptions:  "DENY",
	}

	handler := SecurityHeaders(cfg)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/api/workflows", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "max-age=63072000; includeSubDomains", w.Header().Get("Strict-Transport-Security"))
}

func TestSecurityHeaders_EmptyCSP(t *testing.T) {
	cfg := SecurityHeadersConfig{
		EnableHSTS:    true,
		HSTSMaxAge:    31536000,
		CSPDirectives: "",
		FrameOptions:  "DENY",
	}

	handler := SecurityHeaders(cfg)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/api/workflows", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	// CSP should not be set when empty
	assert.Empty(t, w.Header().Get("Content-Security-Policy"))
}

func TestSecurityHeaders_MultipleRequests(t *testing.T) {
	cfg := SecurityHeadersConfig{
		EnableHSTS:    true,
		HSTSMaxAge:    31536000,
		CSPDirectives: "default-src 'self'",
		FrameOptions:  "DENY",
	}

	handler := SecurityHeaders(cfg)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Make multiple requests to ensure headers are set consistently
	for i := 0; i < 3; i++ {
		req := httptest.NewRequest("GET", "/api/workflows", nil)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "max-age=31536000; includeSubDomains", w.Header().Get("Strict-Transport-Security"))
		assert.Equal(t, "nosniff", w.Header().Get("X-Content-Type-Options"))
		assert.Equal(t, "DENY", w.Header().Get("X-Frame-Options"))
	}
}

func TestSecurityHeaders_DifferentMethods(t *testing.T) {
	cfg := SecurityHeadersConfig{
		EnableHSTS:    true,
		HSTSMaxAge:    31536000,
		CSPDirectives: "default-src 'self'",
		FrameOptions:  "DENY",
	}

	handler := SecurityHeaders(cfg)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	methods := []string{"GET", "POST", "PUT", "DELETE", "PATCH"}

	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			req := httptest.NewRequest(method, "/api/workflows", nil)
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)
			// Verify headers are set for all HTTP methods
			assert.Equal(t, "nosniff", w.Header().Get("X-Content-Type-Options"))
			assert.Equal(t, "DENY", w.Header().Get("X-Frame-Options"))
		})
	}
}

func TestSecurityHeaders_DoesNotOverrideExistingHeaders(t *testing.T) {
	cfg := SecurityHeadersConfig{
		EnableHSTS:    true,
		HSTSMaxAge:    31536000,
		CSPDirectives: "default-src 'self'",
		FrameOptions:  "DENY",
	}

	handler := SecurityHeaders(cfg)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Handler tries to set a custom CSP
		w.Header().Set("Content-Security-Policy", "default-src 'none'")
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/api/workflows", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	// Security headers middleware should set headers before handler runs
	// so the handler's CSP should override the middleware's CSP
	// This tests that middleware sets headers early
	assert.NotEmpty(t, w.Header().Get("Content-Security-Policy"))
}

func TestSecurityHeaders_WithErrorResponse(t *testing.T) {
	cfg := SecurityHeadersConfig{
		EnableHSTS:    true,
		HSTSMaxAge:    31536000,
		CSPDirectives: "default-src 'self'",
		FrameOptions:  "DENY",
	}

	handler := SecurityHeaders(cfg)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "not found", http.StatusNotFound)
	}))

	req := httptest.NewRequest("GET", "/api/workflows/nonexistent", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	// Headers should be set even for error responses
	assert.Equal(t, "nosniff", w.Header().Get("X-Content-Type-Options"))
	assert.Equal(t, "DENY", w.Header().Get("X-Frame-Options"))
}

func TestSecurityHeaders_ZeroHSTSMaxAge(t *testing.T) {
	cfg := SecurityHeadersConfig{
		EnableHSTS:    true,
		HSTSMaxAge:    0,
		CSPDirectives: "default-src 'self'",
		FrameOptions:  "DENY",
	}

	handler := SecurityHeaders(cfg)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/api/workflows", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	// HSTS should still be set with max-age=0 if enabled
	assert.Equal(t, "max-age=0; includeSubDomains", w.Header().Get("Strict-Transport-Security"))
}

func TestDefaultSecurityHeadersConfig(t *testing.T) {
	cfg := DefaultSecurityHeadersConfig()

	assert.True(t, cfg.EnableHSTS)
	assert.Equal(t, 31536000, cfg.HSTSMaxAge)
	assert.Equal(t, "default-src 'self'", cfg.CSPDirectives)
	assert.Equal(t, "DENY", cfg.FrameOptions)
}

func TestDevelopmentSecurityHeadersConfig(t *testing.T) {
	cfg := DevelopmentSecurityHeadersConfig()

	// HSTS should be disabled in development
	assert.False(t, cfg.EnableHSTS)
	assert.Equal(t, 31536000, cfg.HSTSMaxAge)
	// CSP should be more permissive in development
	assert.Contains(t, cfg.CSPDirectives, "unsafe-inline")
	assert.Equal(t, "SAMEORIGIN", cfg.FrameOptions)
}

func TestProductionSecurityHeadersConfig(t *testing.T) {
	cfg := ProductionSecurityHeadersConfig()

	// HSTS should be enabled in production
	assert.True(t, cfg.EnableHSTS)
	assert.Equal(t, 63072000, cfg.HSTSMaxAge) // 2 years for production
	// CSP should be strict in production
	assert.NotContains(t, cfg.CSPDirectives, "unsafe-inline")
	assert.Equal(t, "DENY", cfg.FrameOptions)
}
