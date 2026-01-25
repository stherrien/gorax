package middleware

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"
)

// SecurityHeadersConfig holds configuration for security headers
type SecurityHeadersConfig struct {
	// EnableHSTS controls whether to set Strict-Transport-Security header
	EnableHSTS bool
	// HSTSMaxAge is the max-age value for HSTS in seconds (default: 31536000 = 1 year)
	HSTSMaxAge int
	// HSTSPreload enables HSTS preload list eligibility
	HSTSPreload bool
	// CSPDirectives is the Content-Security-Policy directive (default: "default-src 'self'")
	CSPDirectives string
	// CSPReportOnly sets headers in report-only mode (no enforcement)
	CSPReportOnly bool
	// CSPReportURI is the URI to report CSP violations to
	CSPReportURI string
	// EnableCSPNonce enables nonce generation for inline scripts
	EnableCSPNonce bool
	// FrameOptions controls X-Frame-Options header (default: "DENY", can be "SAMEORIGIN")
	FrameOptions string
	// ReferrerPolicy controls the Referrer-Policy header (default: "strict-origin-when-cross-origin")
	ReferrerPolicy string
	// PermissionsPolicy controls the Permissions-Policy header
	PermissionsPolicy string
	// CrossOriginOpenerPolicy controls Cross-Origin-Opener-Policy header
	CrossOriginOpenerPolicy string
	// CrossOriginEmbedderPolicy controls Cross-Origin-Embedder-Policy header
	CrossOriginEmbedderPolicy string
	// CrossOriginResourcePolicy controls Cross-Origin-Resource-Policy header
	CrossOriginResourcePolicy string
}

// CSPNonceContextKey is the context key for CSP nonce
type cspNonceContextKeyType struct{}

var cspNonceContextKey = cspNonceContextKeyType{}

// GetCSPNonce retrieves the CSP nonce from the request context
func GetCSPNonce(r *http.Request) string {
	if nonce, ok := r.Context().Value(cspNonceContextKey).(string); ok {
		return nonce
	}
	return ""
}

// DefaultSecurityHeadersConfig returns a secure default configuration
func DefaultSecurityHeadersConfig() SecurityHeadersConfig {
	return SecurityHeadersConfig{
		EnableHSTS:                true,
		HSTSMaxAge:                31536000, // 1 year
		CSPDirectives:             "default-src 'self'",
		FrameOptions:              "DENY",
		ReferrerPolicy:            "strict-origin-when-cross-origin",
		PermissionsPolicy:         "geolocation=(), microphone=(), camera=(), payment=(), usb=(), magnetometer=(), gyroscope=(), accelerometer=()",
		CrossOriginOpenerPolicy:   "same-origin",
		CrossOriginResourcePolicy: "same-origin",
	}
}

// DevelopmentSecurityHeadersConfig returns a configuration suitable for development
func DevelopmentSecurityHeadersConfig() SecurityHeadersConfig {
	return SecurityHeadersConfig{
		EnableHSTS:                false, // Disable HSTS in development
		HSTSMaxAge:                31536000,
		CSPDirectives:             "default-src 'self' 'unsafe-inline' 'unsafe-eval'; connect-src 'self' ws: wss: http://localhost:*; img-src 'self' data: blob:",
		FrameOptions:              "SAMEORIGIN",
		ReferrerPolicy:            "strict-origin-when-cross-origin",
		PermissionsPolicy:         "geolocation=(), microphone=(), camera=()",
		CrossOriginOpenerPolicy:   "same-origin-allow-popups", // Allow OAuth popups
		CrossOriginResourcePolicy: "cross-origin",             // Allow cross-origin in dev
	}
}

// ProductionSecurityHeadersConfig returns a strict configuration for production
func ProductionSecurityHeadersConfig() SecurityHeadersConfig {
	return SecurityHeadersConfig{
		EnableHSTS:                true,
		HSTSMaxAge:                63072000, // 2 years
		HSTSPreload:               true,
		EnableCSPNonce:            true, // Use nonce for inline scripts/styles instead of unsafe-inline
		CSPDirectives:             "default-src 'self'; script-src 'self'; style-src 'self'; img-src 'self' data: https:; connect-src 'self' wss:; font-src 'self'; object-src 'none'; base-uri 'self'; form-action 'self'; frame-ancestors 'none'; upgrade-insecure-requests",
		FrameOptions:              "DENY",
		ReferrerPolicy:            "strict-origin-when-cross-origin",
		PermissionsPolicy:         "geolocation=(), microphone=(), camera=(), payment=(), usb=(), magnetometer=(), gyroscope=(), accelerometer=(), interest-cohort=()",
		CrossOriginOpenerPolicy:   "same-origin",
		CrossOriginEmbedderPolicy: "require-corp",
		CrossOriginResourcePolicy: "same-origin",
	}
}

// SecurityHeaders adds security headers to HTTP responses
func SecurityHeaders(cfg SecurityHeadersConfig) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Generate CSP nonce if enabled
			var nonce string
			if cfg.EnableCSPNonce {
				nonce = generateCSPNonce()
				ctx := context.WithValue(r.Context(), cspNonceContextKey, nonce)
				r = r.WithContext(ctx)
			}

			// Set X-Content-Type-Options header (prevents MIME sniffing)
			w.Header().Set("X-Content-Type-Options", "nosniff")

			// Set X-Frame-Options header (clickjacking protection)
			if cfg.FrameOptions != "" {
				w.Header().Set("X-Frame-Options", cfg.FrameOptions)
			}

			// Set X-XSS-Protection header (legacy XSS protection)
			// Note: CSP is the modern replacement, but this provides defense-in-depth
			w.Header().Set("X-XSS-Protection", "1; mode=block")

			// Set Content-Security-Policy header
			if cfg.CSPDirectives != "" {
				csp := cfg.CSPDirectives

				// Add nonce to script-src and style-src if enabled
				if nonce != "" {
					csp = addNonceToCSP(csp, nonce)
				}

				// Add report URI if configured
				if cfg.CSPReportURI != "" {
					csp = csp + "; report-uri " + cfg.CSPReportURI
				}

				if cfg.CSPReportOnly {
					w.Header().Set("Content-Security-Policy-Report-Only", csp)
				} else {
					w.Header().Set("Content-Security-Policy", csp)
				}
			}

			// Set Referrer-Policy header
			referrerPolicy := cfg.ReferrerPolicy
			if referrerPolicy == "" {
				referrerPolicy = "strict-origin-when-cross-origin"
			}
			w.Header().Set("Referrer-Policy", referrerPolicy)

			// Set Permissions-Policy header (feature policy)
			permissionsPolicy := cfg.PermissionsPolicy
			if permissionsPolicy == "" {
				permissionsPolicy = "geolocation=(), microphone=(), camera=()"
			}
			w.Header().Set("Permissions-Policy", permissionsPolicy)

			// Set Cross-Origin-Opener-Policy header
			if cfg.CrossOriginOpenerPolicy != "" {
				w.Header().Set("Cross-Origin-Opener-Policy", cfg.CrossOriginOpenerPolicy)
			}

			// Set Cross-Origin-Embedder-Policy header
			if cfg.CrossOriginEmbedderPolicy != "" {
				w.Header().Set("Cross-Origin-Embedder-Policy", cfg.CrossOriginEmbedderPolicy)
			}

			// Set Cross-Origin-Resource-Policy header
			if cfg.CrossOriginResourcePolicy != "" {
				w.Header().Set("Cross-Origin-Resource-Policy", cfg.CrossOriginResourcePolicy)
			}

			// Set Strict-Transport-Security header (HSTS)
			if cfg.EnableHSTS {
				hstsValue := fmt.Sprintf("max-age=%d; includeSubDomains", cfg.HSTSMaxAge)
				if cfg.HSTSPreload {
					hstsValue += "; preload"
				}
				w.Header().Set("Strict-Transport-Security", hstsValue)
			}

			// Set X-DNS-Prefetch-Control header
			w.Header().Set("X-DNS-Prefetch-Control", "off")

			// Set X-Download-Options header (IE)
			w.Header().Set("X-Download-Options", "noopen")

			// Set X-Permitted-Cross-Domain-Policies header (Adobe products)
			w.Header().Set("X-Permitted-Cross-Domain-Policies", "none")

			// Continue with the request
			next.ServeHTTP(w, r)
		})
	}
}

// generateCSPNonce generates a random nonce for CSP
func generateCSPNonce() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return ""
	}
	return base64.StdEncoding.EncodeToString(b)
}

// addNonceToCSP adds nonce to script-src and style-src directives
func addNonceToCSP(csp, nonce string) string {
	nonceValue := "'nonce-" + nonce + "'"

	// Parse and modify directives
	directives := strings.Split(csp, ";")
	for i, directive := range directives {
		directive = strings.TrimSpace(directive)
		lower := strings.ToLower(directive)

		if strings.HasPrefix(lower, "script-src") {
			directives[i] = directive + " " + nonceValue
		} else if strings.HasPrefix(lower, "style-src") {
			directives[i] = directive + " " + nonceValue
		}
	}

	return strings.Join(directives, "; ")
}
