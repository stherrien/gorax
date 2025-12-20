package middleware

import (
	"fmt"
	"net/http"
)

// SecurityHeadersConfig holds configuration for security headers
type SecurityHeadersConfig struct {
	// EnableHSTS controls whether to set Strict-Transport-Security header
	EnableHSTS bool
	// HSTSMaxAge is the max-age value for HSTS in seconds (default: 31536000 = 1 year)
	HSTSMaxAge int
	// CSPDirectives is the Content-Security-Policy directive (default: "default-src 'self'")
	CSPDirectives string
	// FrameOptions controls X-Frame-Options header (default: "DENY", can be "SAMEORIGIN")
	FrameOptions string
}

// DefaultSecurityHeadersConfig returns a secure default configuration
func DefaultSecurityHeadersConfig() SecurityHeadersConfig {
	return SecurityHeadersConfig{
		EnableHSTS:    true,
		HSTSMaxAge:    31536000, // 1 year
		CSPDirectives: "default-src 'self'",
		FrameOptions:  "DENY",
	}
}

// DevelopmentSecurityHeadersConfig returns a configuration suitable for development
func DevelopmentSecurityHeadersConfig() SecurityHeadersConfig {
	return SecurityHeadersConfig{
		EnableHSTS:    false, // Disable HSTS in development
		HSTSMaxAge:    31536000,
		CSPDirectives: "default-src 'self' 'unsafe-inline' 'unsafe-eval'; connect-src 'self' ws: wss:",
		FrameOptions:  "SAMEORIGIN",
	}
}

// ProductionSecurityHeadersConfig returns a strict configuration for production
func ProductionSecurityHeadersConfig() SecurityHeadersConfig {
	return SecurityHeadersConfig{
		EnableHSTS:    true,
		HSTSMaxAge:    63072000, // 2 years
		CSPDirectives: "default-src 'self'; script-src 'self'; style-src 'self'; connect-src 'self' wss:",
		FrameOptions:  "DENY",
	}
}

// SecurityHeaders adds security headers to HTTP responses
func SecurityHeaders(cfg SecurityHeadersConfig) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Set X-Content-Type-Options header
			w.Header().Set("X-Content-Type-Options", "nosniff")

			// Set X-Frame-Options header
			if cfg.FrameOptions != "" {
				w.Header().Set("X-Frame-Options", cfg.FrameOptions)
			}

			// Set X-XSS-Protection header
			w.Header().Set("X-XSS-Protection", "1; mode=block")

			// Set Content-Security-Policy header
			if cfg.CSPDirectives != "" {
				w.Header().Set("Content-Security-Policy", cfg.CSPDirectives)
			}

			// Set Referrer-Policy header
			w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")

			// Set Permissions-Policy header
			w.Header().Set("Permissions-Policy", "geolocation=(), microphone=(), camera=()")

			// Set Strict-Transport-Security header (HSTS)
			if cfg.EnableHSTS {
				hstsValue := fmt.Sprintf("max-age=%d; includeSubDomains", cfg.HSTSMaxAge)
				w.Header().Set("Strict-Transport-Security", hstsValue)
			}

			// Continue with the request
			next.ServeHTTP(w, r)
		})
	}
}
