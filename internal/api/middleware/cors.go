package middleware

import (
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/gorax/gorax/internal/config"
)

// ValidateCORSConfig validates CORS configuration for the given environment
func ValidateCORSConfig(cfg config.CORSConfig, env string) error {
	// Validate that at least one origin is specified
	if len(cfg.AllowedOrigins) == 0 {
		return fmt.Errorf("at least one allowed origin must be specified")
	}

	// Validate that at least one method is specified
	if len(cfg.AllowedMethods) == 0 {
		return fmt.Errorf("at least one allowed method must be specified")
	}

	// Validate max age is non-negative
	if cfg.MaxAge < 0 {
		return fmt.Errorf("max age must be non-negative")
	}

	// Production-specific validations
	if env == "production" {
		for _, origin := range cfg.AllowedOrigins {
			// Reject wildcard origins
			if origin == "*" {
				return fmt.Errorf("wildcard (*) origin not allowed in production")
			}

			// Reject localhost origins
			if strings.Contains(origin, "localhost") || strings.Contains(origin, "127.0.0.1") {
				return fmt.Errorf("localhost origins not allowed in production: %s", origin)
			}

			// Warn if HTTP is used (should be HTTPS in production)
			if strings.HasPrefix(origin, "http://") {
				slog.Warn("HTTP origin detected in production (HTTPS recommended)", "origin", origin)
			}
		}

		// Warn if credentials are allowed with multiple origins
		if cfg.AllowCredentials && len(cfg.AllowedOrigins) > 2 {
			slog.Warn("credentials allowed with multiple origins may be a security risk",
				"origin_count", len(cfg.AllowedOrigins))
		}
	}

	return nil
}

// NewCORSMiddleware creates a CORS middleware with the given configuration
func NewCORSMiddleware(cfg config.CORSConfig, env string) (func(http.Handler) http.Handler, error) {
	// Validate configuration
	if err := ValidateCORSConfig(cfg, env); err != nil {
		return nil, fmt.Errorf("invalid CORS configuration: %w", err)
	}

	// Build allowed origins map for fast lookup
	allowedOriginsMap := make(map[string]bool)
	hasWildcard := false
	for _, origin := range cfg.AllowedOrigins {
		if origin == "*" {
			hasWildcard = true
		}
		allowedOriginsMap[origin] = true
	}

	// Build header strings
	allowedMethods := strings.Join(cfg.AllowedMethods, ", ")
	allowedHeaders := strings.Join(cfg.AllowedHeaders, ", ")
	exposedHeaders := strings.Join(cfg.ExposedHeaders, ", ")
	maxAge := fmt.Sprintf("%d", cfg.MaxAge)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")

			// Check if origin is allowed
			isAllowed := hasWildcard || allowedOriginsMap[origin]

			if isAllowed {
				// Set CORS headers for allowed origins
				if hasWildcard {
					w.Header().Set("Access-Control-Allow-Origin", "*")
				} else {
					w.Header().Set("Access-Control-Allow-Origin", origin)
				}

				if cfg.AllowCredentials {
					w.Header().Set("Access-Control-Allow-Credentials", "true")
				}

				if len(cfg.ExposedHeaders) > 0 {
					w.Header().Set("Access-Control-Expose-Headers", exposedHeaders)
				}

				// Handle preflight requests
				if r.Method == http.MethodOptions {
					w.Header().Set("Access-Control-Allow-Methods", allowedMethods)
					w.Header().Set("Access-Control-Allow-Headers", allowedHeaders)
					if cfg.MaxAge > 0 {
						w.Header().Set("Access-Control-Max-Age", maxAge)
					}
					w.WriteHeader(http.StatusOK)
					return
				}
			}

			// Continue to next handler
			next.ServeHTTP(w, r)
		})
	}, nil
}
