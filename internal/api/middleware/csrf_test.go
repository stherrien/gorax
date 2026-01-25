package middleware

import (
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultCSRFConfig(t *testing.T) {
	cfg := DefaultCSRFConfig()

	assert.True(t, cfg.Enabled)
	assert.Equal(t, 32, cfg.TokenLength)
	assert.Equal(t, "csrf_token", cfg.CookieName)
	assert.Equal(t, "X-CSRF-Token", cfg.HeaderName)
	assert.Equal(t, "_csrf", cfg.FormFieldName)
	assert.Equal(t, "/", cfg.CookiePath)
	assert.Equal(t, 43200, cfg.CookieMaxAge)
	assert.True(t, cfg.CookieSecure)
	assert.Equal(t, http.SameSiteStrictMode, cfg.CookieSameSite)
	assert.Equal(t, []string{"GET", "HEAD", "OPTIONS", "TRACE"}, cfg.ExemptMethods)
	assert.Equal(t, []string{"/health", "/ready", "/webhooks/"}, cfg.ExemptPaths)
}

func TestDevelopmentCSRFConfig(t *testing.T) {
	cfg := DevelopmentCSRFConfig()

	assert.True(t, cfg.Enabled)
	assert.False(t, cfg.CookieSecure)
	assert.Equal(t, http.SameSiteLaxMode, cfg.CookieSameSite)
	// Other values should match default
	assert.Equal(t, 32, cfg.TokenLength)
	assert.Equal(t, "csrf_token", cfg.CookieName)
}

func TestNewCSRFProtection(t *testing.T) {
	tests := []struct {
		name     string
		config   CSRFConfig
		expected CSRFConfig
	}{
		{
			name:   "empty config gets defaults",
			config: CSRFConfig{},
			expected: CSRFConfig{
				TokenLength:    32,
				CookieName:     "csrf_token",
				HeaderName:     "X-CSRF-Token",
				FormFieldName:  "_csrf",
				CookiePath:     "/",
				CookieMaxAge:   43200,
				CookieSameSite: http.SameSiteStrictMode,
				ExemptMethods:  []string{"GET", "HEAD", "OPTIONS", "TRACE"},
			},
		},
		{
			name: "custom values preserved",
			config: CSRFConfig{
				TokenLength:    64,
				CookieName:     "my_csrf",
				HeaderName:     "X-My-CSRF",
				FormFieldName:  "my_csrf_field",
				CookiePath:     "/api",
				CookieMaxAge:   3600,
				CookieSameSite: http.SameSiteLaxMode,
				ExemptMethods:  []string{"GET"},
			},
			expected: CSRFConfig{
				TokenLength:    64,
				CookieName:     "my_csrf",
				HeaderName:     "X-My-CSRF",
				FormFieldName:  "my_csrf_field",
				CookiePath:     "/api",
				CookieMaxAge:   3600,
				CookieSameSite: http.SameSiteLaxMode,
				ExemptMethods:  []string{"GET"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			csrf := NewCSRFProtection(tt.config)

			assert.Equal(t, tt.expected.TokenLength, csrf.config.TokenLength)
			assert.Equal(t, tt.expected.CookieName, csrf.config.CookieName)
			assert.Equal(t, tt.expected.HeaderName, csrf.config.HeaderName)
			assert.Equal(t, tt.expected.FormFieldName, csrf.config.FormFieldName)
			assert.Equal(t, tt.expected.CookiePath, csrf.config.CookiePath)
			assert.Equal(t, tt.expected.CookieMaxAge, csrf.config.CookieMaxAge)
			assert.Equal(t, tt.expected.CookieSameSite, csrf.config.CookieSameSite)
			assert.Equal(t, tt.expected.ExemptMethods, csrf.config.ExemptMethods)
		})
	}
}

func TestCSRFMiddleware_Disabled(t *testing.T) {
	cfg := DefaultCSRFConfig()
	cfg.Enabled = false
	csrf := NewCSRFProtection(cfg)

	called := false
	handler := csrf.Middleware()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))

	// POST without any CSRF token should pass when disabled
	req := httptest.NewRequest(http.MethodPost, "/api/v1/resource", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.True(t, called, "handler should be called when CSRF is disabled")
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestCSRFMiddleware_ExemptMethods(t *testing.T) {
	cfg := DefaultCSRFConfig()
	csrf := NewCSRFProtection(cfg)

	exemptMethods := []string{"GET", "HEAD", "OPTIONS", "TRACE"}

	for _, method := range exemptMethods {
		t.Run(method, func(t *testing.T) {
			called := false
			handler := csrf.Middleware()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				called = true
				w.WriteHeader(http.StatusOK)
			}))

			req := httptest.NewRequest(method, "/api/v1/resource", nil)
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			assert.True(t, called, "handler should be called for exempt method %s", method)
			assert.Equal(t, http.StatusOK, w.Code)

			// Should set CSRF token cookie for safe methods
			cookies := w.Result().Cookies()
			var csrfCookie *http.Cookie
			for _, c := range cookies {
				if c.Name == cfg.CookieName {
					csrfCookie = c
					break
				}
			}
			assert.NotNil(t, csrfCookie, "CSRF cookie should be set for exempt method")
		})
	}
}

func TestCSRFMiddleware_ExemptPaths(t *testing.T) {
	cfg := DefaultCSRFConfig()
	cfg.ExemptPaths = []string{"/health", "/ready", "/webhooks/"}
	csrf := NewCSRFProtection(cfg)

	tests := []struct {
		path   string
		exempt bool
	}{
		{"/health", true},
		{"/ready", true},
		{"/webhooks/github", true},
		{"/webhooks/", true},
		{"/api/v1/workflows", false},
		{"/api/v1/health-check", false}, // not a prefix match
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			called := false
			handler := csrf.Middleware()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				called = true
				w.WriteHeader(http.StatusOK)
			}))

			req := httptest.NewRequest(http.MethodPost, tt.path, nil)
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			if tt.exempt {
				assert.True(t, called, "handler should be called for exempt path %s", tt.path)
				assert.Equal(t, http.StatusOK, w.Code)
			} else {
				// Non-exempt paths without CSRF token should fail
				assert.False(t, called, "handler should not be called for non-exempt path %s", tt.path)
				assert.Equal(t, http.StatusForbidden, w.Code)
			}
		})
	}
}

func TestCSRFMiddleware_ValidToken(t *testing.T) {
	cfg := DefaultCSRFConfig()
	csrf := NewCSRFProtection(cfg)

	// First, get a valid token
	token := csrf.generateToken()

	tests := []struct {
		name       string
		setupToken func(req *http.Request, token string)
	}{
		{
			name: "token in header",
			setupToken: func(req *http.Request, token string) {
				req.Header.Set(cfg.HeaderName, token)
			},
		},
		{
			name: "token in form field",
			setupToken: func(req *http.Request, token string) {
				req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
				req.Body = nil // Will be set from PostForm
				req.PostForm = make(map[string][]string)
				req.PostForm.Set(cfg.FormFieldName, token)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			called := false
			handler := csrf.Middleware()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				called = true
				w.WriteHeader(http.StatusOK)
			}))

			req := httptest.NewRequest(http.MethodPost, "/api/v1/resource", nil)
			req.AddCookie(&http.Cookie{
				Name:  cfg.CookieName,
				Value: token,
			})
			tt.setupToken(req, token)
			req.Host = "localhost"

			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)

			assert.True(t, called, "handler should be called with valid CSRF token")
			assert.Equal(t, http.StatusOK, w.Code)
		})
	}
}

func TestCSRFMiddleware_InvalidToken(t *testing.T) {
	cfg := DefaultCSRFConfig()
	csrf := NewCSRFProtection(cfg)

	validToken := csrf.generateToken()

	tests := []struct {
		name          string
		cookieToken   string
		headerToken   string
		expectedError string
	}{
		{
			name:          "missing cookie token",
			cookieToken:   "",
			headerToken:   validToken,
			expectedError: "CSRF token validation failed",
		},
		{
			name:          "missing header token",
			cookieToken:   validToken,
			headerToken:   "",
			expectedError: "CSRF token validation failed",
		},
		{
			name:          "mismatched tokens",
			cookieToken:   validToken,
			headerToken:   base64.URLEncoding.EncodeToString([]byte("different-token")),
			expectedError: "CSRF token validation failed",
		},
		{
			name:          "invalid base64 in cookie",
			cookieToken:   "not-valid-base64!!!",
			headerToken:   validToken,
			expectedError: "CSRF token validation failed",
		},
		{
			name:          "invalid base64 in header",
			cookieToken:   validToken,
			headerToken:   "not-valid-base64!!!",
			expectedError: "CSRF token validation failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			called := false
			handler := csrf.Middleware()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				called = true
				w.WriteHeader(http.StatusOK)
			}))

			req := httptest.NewRequest(http.MethodPost, "/api/v1/resource", nil)
			if tt.cookieToken != "" {
				req.AddCookie(&http.Cookie{
					Name:  cfg.CookieName,
					Value: tt.cookieToken,
				})
			}
			if tt.headerToken != "" {
				req.Header.Set(cfg.HeaderName, tt.headerToken)
			}

			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)

			assert.False(t, called, "handler should not be called with invalid CSRF token")
			assert.Equal(t, http.StatusForbidden, w.Code)
		})
	}
}

func TestCSRFMiddleware_OriginValidation(t *testing.T) {
	tests := []struct {
		name           string
		trustedOrigins []string
		origin         string
		referer        string
		host           string
		expectPass     bool
	}{
		{
			name:           "same origin - no Origin header",
			trustedOrigins: nil,
			origin:         "",
			referer:        "",
			host:           "localhost",
			expectPass:     true, // browsers don't send Origin for same-origin
		},
		{
			name:           "same origin via referer",
			trustedOrigins: nil,
			origin:         "",
			referer:        "http://localhost/page",
			host:           "localhost",
			expectPass:     true,
		},
		{
			name:           "cross origin without trusted list",
			trustedOrigins: nil,
			origin:         "http://evil.com",
			referer:        "",
			host:           "localhost",
			expectPass:     false,
		},
		{
			name:           "cross origin with trusted list - allowed",
			trustedOrigins: []string{"http://trusted.com", "http://localhost"},
			origin:         "http://trusted.com",
			referer:        "",
			host:           "api.example.com",
			expectPass:     true,
		},
		{
			name:           "cross origin with trusted list - not allowed",
			trustedOrigins: []string{"http://trusted.com"},
			origin:         "http://evil.com",
			referer:        "",
			host:           "api.example.com",
			expectPass:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := DefaultCSRFConfig()
			cfg.TrustedOrigins = tt.trustedOrigins
			csrf := NewCSRFProtection(cfg)

			// Generate valid token
			token := csrf.generateToken()

			called := false
			handler := csrf.Middleware()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				called = true
				w.WriteHeader(http.StatusOK)
			}))

			req := httptest.NewRequest(http.MethodPost, "/api/v1/resource", nil)
			req.Host = tt.host
			req.AddCookie(&http.Cookie{
				Name:  cfg.CookieName,
				Value: token,
			})
			req.Header.Set(cfg.HeaderName, token)

			if tt.origin != "" {
				req.Header.Set("Origin", tt.origin)
			}
			if tt.referer != "" {
				req.Header.Set("Referer", tt.referer)
			}

			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)

			if tt.expectPass {
				assert.True(t, called, "handler should be called")
				assert.Equal(t, http.StatusOK, w.Code)
			} else {
				assert.False(t, called, "handler should not be called")
				assert.Equal(t, http.StatusForbidden, w.Code)
			}
		})
	}
}

func TestCSRFMiddleware_CookieAlreadyExists(t *testing.T) {
	cfg := DefaultCSRFConfig()
	csrf := NewCSRFProtection(cfg)

	existingToken := csrf.generateToken()

	handler := csrf.Middleware()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/resource", nil)
	req.AddCookie(&http.Cookie{
		Name:  cfg.CookieName,
		Value: existingToken,
	})

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Should not set a new cookie if one already exists
	cookies := w.Result().Cookies()
	for _, c := range cookies {
		if c.Name == cfg.CookieName {
			t.Error("should not set new CSRF cookie when one already exists")
		}
	}
}

func TestCSRF_GenerateToken(t *testing.T) {
	t.Run("without secret key", func(t *testing.T) {
		cfg := DefaultCSRFConfig()
		csrf := NewCSRFProtection(cfg)

		token1 := csrf.generateToken()
		token2 := csrf.generateToken()

		// Tokens should be base64 encoded
		_, err := base64.URLEncoding.DecodeString(token1)
		assert.NoError(t, err)

		// Each token should be unique
		assert.NotEqual(t, token1, token2)

		// Token should be cached
		_, exists := csrf.tokenCache.Load(token1)
		assert.True(t, exists, "token should be cached")
	})

	t.Run("with secret key", func(t *testing.T) {
		cfg := DefaultCSRFConfig()
		cfg.SecretKey = []byte("super-secret-key-for-signing-tokens")
		csrf := NewCSRFProtection(cfg)

		token := csrf.generateToken()

		// Token should be base64 encoded
		decoded, err := base64.URLEncoding.DecodeString(token)
		assert.NoError(t, err)

		// With HMAC signing, the token should be 32 bytes (sha256 output)
		assert.Equal(t, 32, len(decoded))
	})
}

func TestCSRF_VerifyToken(t *testing.T) {
	cfg := DefaultCSRFConfig()
	csrf := NewCSRFProtection(cfg)

	tests := []struct {
		name         string
		cookieToken  string
		requestToken string
		expectValid  bool
	}{
		{
			name:         "matching tokens",
			cookieToken:  base64.URLEncoding.EncodeToString([]byte("same-token")),
			requestToken: base64.URLEncoding.EncodeToString([]byte("same-token")),
			expectValid:  true,
		},
		{
			name:         "different tokens",
			cookieToken:  base64.URLEncoding.EncodeToString([]byte("token-a")),
			requestToken: base64.URLEncoding.EncodeToString([]byte("token-b")),
			expectValid:  false,
		},
		{
			name:         "invalid cookie base64",
			cookieToken:  "not-valid-base64!!!",
			requestToken: base64.URLEncoding.EncodeToString([]byte("token")),
			expectValid:  false,
		},
		{
			name:         "invalid request base64",
			cookieToken:  base64.URLEncoding.EncodeToString([]byte("token")),
			requestToken: "not-valid-base64!!!",
			expectValid:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := csrf.verifyToken(tt.cookieToken, tt.requestToken)
			assert.Equal(t, tt.expectValid, result)
		})
	}
}

func TestCSRF_IsExemptMethod(t *testing.T) {
	cfg := DefaultCSRFConfig()
	csrf := NewCSRFProtection(cfg)

	tests := []struct {
		method string
		exempt bool
	}{
		{"GET", true},
		{"get", true}, // case insensitive
		{"HEAD", true},
		{"OPTIONS", true},
		{"TRACE", true},
		{"POST", false},
		{"PUT", false},
		{"DELETE", false},
		{"PATCH", false},
	}

	for _, tt := range tests {
		t.Run(tt.method, func(t *testing.T) {
			result := csrf.isExemptMethod(tt.method)
			assert.Equal(t, tt.exempt, result)
		})
	}
}

func TestCSRF_IsExemptPath(t *testing.T) {
	cfg := DefaultCSRFConfig()
	cfg.ExemptPaths = []string{"/health", "/ready", "/webhooks/"}
	csrf := NewCSRFProtection(cfg)

	tests := []struct {
		path   string
		exempt bool
	}{
		{"/health", true},
		{"/health/check", true}, // prefix match
		{"/ready", true},
		{"/ready/", true},
		{"/webhooks/", true},
		{"/webhooks/github", true},
		{"/api/v1/workflows", false},
		{"/api/health", false}, // health is not a prefix here
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			result := csrf.isExemptPath(tt.path)
			assert.Equal(t, tt.exempt, result)
		})
	}
}

func TestCSRF_GetToken(t *testing.T) {
	cfg := DefaultCSRFConfig()
	csrf := NewCSRFProtection(cfg)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/csrf-token", nil)
	w := httptest.NewRecorder()

	csrf.GetToken(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Should set cookie
	cookies := w.Result().Cookies()
	var csrfCookie *http.Cookie
	for _, c := range cookies {
		if c.Name == cfg.CookieName {
			csrfCookie = c
			break
		}
	}
	require.NotNil(t, csrfCookie, "CSRF cookie should be set")
	assert.False(t, csrfCookie.HttpOnly, "CSRF cookie must be readable by JavaScript")
	assert.Equal(t, cfg.CookiePath, csrfCookie.Path)
	assert.Equal(t, cfg.CookieMaxAge, csrfCookie.MaxAge)

	// Response should contain the token
	assert.Contains(t, w.Body.String(), "token")
	assert.Contains(t, w.Body.String(), csrfCookie.Value)
}

func TestCSRF_GetTokenFromRequest(t *testing.T) {
	cfg := DefaultCSRFConfig()
	csrf := NewCSRFProtection(cfg)

	t.Run("from header", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/", nil)
		req.Header.Set(cfg.HeaderName, "header-token")

		token := csrf.getTokenFromRequest(req)
		assert.Equal(t, "header-token", token)
	})

	t.Run("from form field", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("_csrf=form-token"))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		token := csrf.getTokenFromRequest(req)
		assert.Equal(t, "form-token", token)
	})

	t.Run("header takes precedence", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("_csrf=form-token"))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.Header.Set(cfg.HeaderName, "header-token")

		token := csrf.getTokenFromRequest(req)
		assert.Equal(t, "header-token", token)
	})

	t.Run("no token", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/", nil)

		token := csrf.getTokenFromRequest(req)
		assert.Empty(t, token)
	})
}

func TestCSRF_GetTokenFromCookie(t *testing.T) {
	cfg := DefaultCSRFConfig()
	csrf := NewCSRFProtection(cfg)

	t.Run("cookie exists", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/", nil)
		req.AddCookie(&http.Cookie{
			Name:  cfg.CookieName,
			Value: "cookie-token",
		})

		token := csrf.getTokenFromCookie(req)
		assert.Equal(t, "cookie-token", token)
	})

	t.Run("cookie missing", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/", nil)

		token := csrf.getTokenFromCookie(req)
		assert.Empty(t, token)
	})
}

func TestCSRF_ValidateOrigin(t *testing.T) {
	t.Run("no origin or referer - same origin assumed", func(t *testing.T) {
		cfg := DefaultCSRFConfig()
		csrf := NewCSRFProtection(cfg)

		req := httptest.NewRequest(http.MethodPost, "/", nil)
		req.Host = "localhost"

		result := csrf.validateOrigin(req)
		assert.True(t, result)
	})

	t.Run("origin matches host", func(t *testing.T) {
		cfg := DefaultCSRFConfig()
		csrf := NewCSRFProtection(cfg)

		req := httptest.NewRequest(http.MethodPost, "/", nil)
		req.Host = "localhost:8080"
		req.Header.Set("Origin", "http://localhost:8080")

		result := csrf.validateOrigin(req)
		assert.True(t, result)
	})

	t.Run("origin does not match host", func(t *testing.T) {
		cfg := DefaultCSRFConfig()
		csrf := NewCSRFProtection(cfg)

		req := httptest.NewRequest(http.MethodPost, "/", nil)
		req.Host = "localhost:8080"
		req.Header.Set("Origin", "http://evil.com")

		result := csrf.validateOrigin(req)
		assert.False(t, result)
	})

	t.Run("trusted origins configured", func(t *testing.T) {
		cfg := DefaultCSRFConfig()
		cfg.TrustedOrigins = []string{"http://app.example.com", "https://admin.example.com"}
		csrf := NewCSRFProtection(cfg)

		req := httptest.NewRequest(http.MethodPost, "/", nil)
		req.Host = "api.example.com"
		req.Header.Set("Origin", "http://app.example.com")

		result := csrf.validateOrigin(req)
		assert.True(t, result)
	})
}

func TestCSRF_ConcurrentAccess(t *testing.T) {
	cfg := DefaultCSRFConfig()
	csrf := NewCSRFProtection(cfg)

	// Test concurrent token generation
	done := make(chan bool, 100)

	for range 100 {
		go func() {
			token := csrf.generateToken()
			assert.NotEmpty(t, token)

			// Token should be cached
			_, exists := csrf.tokenCache.Load(token)
			assert.True(t, exists)

			done <- true
		}()
	}

	// Wait for all goroutines
	for range 100 {
		<-done
	}
}

func TestCSRFMiddleware_AllHTTPMethods(t *testing.T) {
	cfg := DefaultCSRFConfig()
	csrf := NewCSRFProtection(cfg)

	stateMutatingMethods := []string{"POST", "PUT", "DELETE", "PATCH"}

	for _, method := range stateMutatingMethods {
		t.Run(method+" without token fails", func(t *testing.T) {
			called := false
			handler := csrf.Middleware()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				called = true
			}))

			req := httptest.NewRequest(method, "/api/v1/resource", nil)
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			assert.False(t, called)
			assert.Equal(t, http.StatusForbidden, w.Code)
		})

		t.Run(method+" with valid token passes", func(t *testing.T) {
			token := csrf.generateToken()

			called := false
			handler := csrf.Middleware()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				called = true
				w.WriteHeader(http.StatusOK)
			}))

			req := httptest.NewRequest(method, "/api/v1/resource", nil)
			req.Host = "localhost"
			req.AddCookie(&http.Cookie{
				Name:  cfg.CookieName,
				Value: token,
			})
			req.Header.Set(cfg.HeaderName, token)

			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)

			assert.True(t, called)
			assert.Equal(t, http.StatusOK, w.Code)
		})
	}
}
