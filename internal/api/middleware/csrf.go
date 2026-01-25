package middleware

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gorax/gorax/internal/api/response"
)

// CSRFConfig holds CSRF protection configuration
type CSRFConfig struct {
	// Enabled controls whether CSRF protection is active
	Enabled bool
	// SecretKey is used for signing tokens (should be 32+ bytes)
	SecretKey []byte
	// TokenLength is the length of random token bytes (default: 32)
	TokenLength int
	// CookieName is the name of the CSRF cookie (default: "csrf_token")
	CookieName string
	// HeaderName is the name of the CSRF header (default: "X-CSRF-Token")
	HeaderName string
	// FormFieldName is the name of the CSRF form field (default: "_csrf")
	FormFieldName string
	// CookiePath is the path for the CSRF cookie (default: "/")
	CookiePath string
	// CookieMaxAge is the max age of the cookie in seconds (default: 12 hours)
	CookieMaxAge int
	// CookieSecure sets the Secure flag on the cookie (default: true in production)
	CookieSecure bool
	// CookieSameSite sets the SameSite attribute (default: Strict)
	CookieSameSite http.SameSite
	// TrustedOrigins are allowed origins for cross-origin requests
	TrustedOrigins []string
	// ExemptPaths are paths that don't require CSRF validation
	ExemptPaths []string
	// ExemptMethods are HTTP methods that don't require CSRF validation
	ExemptMethods []string
}

// DefaultCSRFConfig returns a secure default CSRF configuration
func DefaultCSRFConfig() CSRFConfig {
	return CSRFConfig{
		Enabled:        true,
		TokenLength:    32,
		CookieName:     "csrf_token",
		HeaderName:     "X-CSRF-Token",
		FormFieldName:  "_csrf",
		CookiePath:     "/",
		CookieMaxAge:   43200, // 12 hours
		CookieSecure:   true,
		CookieSameSite: http.SameSiteStrictMode,
		ExemptMethods:  []string{"GET", "HEAD", "OPTIONS", "TRACE"},
		ExemptPaths:    []string{"/health", "/ready", "/webhooks/"},
	}
}

// DevelopmentCSRFConfig returns CSRF config suitable for development
func DevelopmentCSRFConfig() CSRFConfig {
	cfg := DefaultCSRFConfig()
	cfg.CookieSecure = false
	cfg.CookieSameSite = http.SameSiteLaxMode
	return cfg
}

// CSRFProtection provides CSRF protection middleware
type CSRFProtection struct {
	config      CSRFConfig
	tokenCache  sync.Map // Cache of valid tokens
	cleanupOnce sync.Once
}

// NewCSRFProtection creates a new CSRF protection middleware
func NewCSRFProtection(config CSRFConfig) *CSRFProtection {
	if config.TokenLength == 0 {
		config.TokenLength = 32
	}
	if config.CookieName == "" {
		config.CookieName = "csrf_token"
	}
	if config.HeaderName == "" {
		config.HeaderName = "X-CSRF-Token"
	}
	if config.FormFieldName == "" {
		config.FormFieldName = "_csrf"
	}
	if config.CookiePath == "" {
		config.CookiePath = "/"
	}
	if config.CookieMaxAge == 0 {
		config.CookieMaxAge = 43200
	}
	if config.CookieSameSite == 0 {
		config.CookieSameSite = http.SameSiteStrictMode
	}
	if len(config.ExemptMethods) == 0 {
		config.ExemptMethods = []string{"GET", "HEAD", "OPTIONS", "TRACE"}
	}

	csrf := &CSRFProtection{
		config: config,
	}

	// Start cleanup goroutine
	csrf.cleanupOnce.Do(func() {
		go csrf.cleanupExpiredTokens()
	})

	return csrf
}

// Middleware returns the CSRF protection middleware handler
func (c *CSRFProtection) Middleware() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip if CSRF protection is disabled
			if !c.config.Enabled {
				next.ServeHTTP(w, r)
				return
			}

			// Skip exempt methods (safe methods)
			if c.isExemptMethod(r.Method) {
				// Set CSRF token cookie for safe methods
				c.setCSRFCookie(w, r)
				next.ServeHTTP(w, r)
				return
			}

			// Skip exempt paths
			if c.isExemptPath(r.URL.Path) {
				next.ServeHTTP(w, r)
				return
			}

			// Validate CSRF token
			if !c.validateCSRFToken(r) {
				c.respondCSRFError(w, "CSRF token validation failed")
				return
			}

			// Validate origin header for additional protection
			if !c.validateOrigin(r) {
				c.respondCSRFError(w, "Origin validation failed")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// isExemptMethod checks if the HTTP method is exempt from CSRF validation
func (c *CSRFProtection) isExemptMethod(method string) bool {
	for _, m := range c.config.ExemptMethods {
		if strings.EqualFold(method, m) {
			return true
		}
	}
	return false
}

// isExemptPath checks if the path is exempt from CSRF validation
func (c *CSRFProtection) isExemptPath(path string) bool {
	for _, exemptPath := range c.config.ExemptPaths {
		if strings.HasPrefix(path, exemptPath) {
			return true
		}
	}
	return false
}

// validateCSRFToken validates the CSRF token from the request
func (c *CSRFProtection) validateCSRFToken(r *http.Request) bool {
	// Get token from cookie
	cookieToken := c.getTokenFromCookie(r)
	if cookieToken == "" {
		return false
	}

	// Get token from request (header or form)
	requestToken := c.getTokenFromRequest(r)
	if requestToken == "" {
		return false
	}

	// Compare tokens using constant-time comparison
	return c.verifyToken(cookieToken, requestToken)
}

// getTokenFromCookie extracts the CSRF token from the cookie
func (c *CSRFProtection) getTokenFromCookie(r *http.Request) string {
	cookie, err := r.Cookie(c.config.CookieName)
	if err != nil {
		return ""
	}
	return cookie.Value
}

// getTokenFromRequest extracts the CSRF token from header or form
func (c *CSRFProtection) getTokenFromRequest(r *http.Request) string {
	// First check header
	token := r.Header.Get(c.config.HeaderName)
	if token != "" {
		return token
	}

	// Then check form field
	if err := r.ParseForm(); err == nil {
		token = r.FormValue(c.config.FormFieldName)
		if token != "" {
			return token
		}
	}

	return ""
}

// verifyToken compares two tokens using constant-time comparison
func (c *CSRFProtection) verifyToken(cookieToken, requestToken string) bool {
	// Decode tokens
	cookieBytes, err := base64.URLEncoding.DecodeString(cookieToken)
	if err != nil {
		return false
	}
	requestBytes, err := base64.URLEncoding.DecodeString(requestToken)
	if err != nil {
		return false
	}

	// Use HMAC comparison if secret key is configured
	if len(c.config.SecretKey) > 0 {
		// The cookie token is the signed version
		// The request token should match when verified
		return hmac.Equal(cookieBytes, requestBytes)
	}

	// Simple constant-time comparison
	return hmac.Equal(cookieBytes, requestBytes)
}

// validateOrigin validates the Origin header against trusted origins
func (c *CSRFProtection) validateOrigin(r *http.Request) bool {
	origin := r.Header.Get("Origin")
	if origin == "" {
		// No Origin header - check Referer as fallback
		referer := r.Header.Get("Referer")
		if referer == "" {
			// No origin information - allow for same-origin requests
			// (browsers don't send Origin for same-origin)
			return true
		}
		origin = referer
	}

	// If no trusted origins configured, allow same-origin
	if len(c.config.TrustedOrigins) == 0 {
		host := r.Host
		return strings.Contains(origin, host)
	}

	// Check against trusted origins
	for _, trusted := range c.config.TrustedOrigins {
		if strings.HasPrefix(origin, trusted) {
			return true
		}
	}

	return false
}

// setCSRFCookie sets the CSRF token cookie
func (c *CSRFProtection) setCSRFCookie(w http.ResponseWriter, r *http.Request) {
	// Check if cookie already exists
	if _, err := r.Cookie(c.config.CookieName); err == nil {
		return // Cookie already set
	}

	// Generate new token
	token := c.generateToken()

	// Set cookie
	http.SetCookie(w, &http.Cookie{
		Name:     c.config.CookieName,
		Value:    token,
		Path:     c.config.CookiePath,
		MaxAge:   c.config.CookieMaxAge,
		HttpOnly: false, // Must be readable by JavaScript
		Secure:   c.config.CookieSecure,
		SameSite: c.config.CookieSameSite,
	})

	// Also set in response header for SPA convenience
	w.Header().Set(c.config.HeaderName, token)
}

// generateToken generates a new CSRF token
func (c *CSRFProtection) generateToken() string {
	// Generate random bytes
	bytes := make([]byte, c.config.TokenLength)
	if _, err := rand.Read(bytes); err != nil {
		// Fallback to less secure but still random
		return base64.URLEncoding.EncodeToString([]byte(time.Now().String()))
	}

	// If secret key is configured, sign the token
	if len(c.config.SecretKey) > 0 {
		h := hmac.New(sha256.New, c.config.SecretKey)
		h.Write(bytes)
		bytes = h.Sum(nil)
	}

	token := base64.URLEncoding.EncodeToString(bytes)

	// Cache the token with expiry
	c.tokenCache.Store(token, time.Now().Add(time.Duration(c.config.CookieMaxAge)*time.Second))

	return token
}

// cleanupExpiredTokens periodically removes expired tokens from cache
func (c *CSRFProtection) cleanupExpiredTokens() {
	ticker := time.NewTicker(15 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		now := time.Now()
		c.tokenCache.Range(func(key, value any) bool {
			if expiry, ok := value.(time.Time); ok {
				if now.After(expiry) {
					c.tokenCache.Delete(key)
				}
			}
			return true
		})
	}
}

// respondCSRFError sends a CSRF error response
func (c *CSRFProtection) respondCSRFError(w http.ResponseWriter, message string) {
	_ = response.Forbidden(w, message)
}

// GetToken is a handler that returns a new CSRF token
// Useful for SPAs that need to fetch a token via API
func (c *CSRFProtection) GetToken(w http.ResponseWriter, r *http.Request) {
	token := c.generateToken()

	// Set cookie
	http.SetCookie(w, &http.Cookie{
		Name:     c.config.CookieName,
		Value:    token,
		Path:     c.config.CookiePath,
		MaxAge:   c.config.CookieMaxAge,
		HttpOnly: false,
		Secure:   c.config.CookieSecure,
		SameSite: c.config.CookieSameSite,
	})

	_ = response.JSON(w, http.StatusOK, map[string]string{
		"token": token,
	})
}
