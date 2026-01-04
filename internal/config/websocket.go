package config

import (
	"net/http"
	"net/url"
	"strings"
)

// WebSocketConfig holds WebSocket configuration
type WebSocketConfig struct {
	// AllowedOrigins is the list of allowed origins for WebSocket connections
	// Supports exact matches, wildcard subdomains (*.example.com), and wildcard ports (localhost:*)
	// Examples:
	//   - "http://localhost:3000" - exact match
	//   - "https://*.example.com" - any subdomain of example.com
	//   - "http://localhost:*" - any port on localhost
	AllowedOrigins []string

	// MaxMessageSize is the maximum size of a WebSocket message in bytes (default: 512KB)
	MaxMessageSize int64

	// MaxConnectionsPerWorkflow is the maximum number of concurrent connections per workflow (default: 50)
	MaxConnectionsPerWorkflow int

	// ConnectionsPerTenantPerMinute is the rate limit for new WebSocket connections per tenant (default: 60)
	ConnectionsPerTenantPerMinute int
}

// ValidateOrigin validates if the request origin is allowed based on the configuration
func ValidateOrigin(config WebSocketConfig, r *http.Request) bool {
	origin := r.Header.Get("Origin")
	if origin == "" {
		return false
	}

	// If no origins configured, deny all
	if len(config.AllowedOrigins) == 0 {
		return false
	}

	// Parse the request origin
	originURL, err := url.Parse(origin)
	if err != nil {
		return false
	}

	// Check each allowed origin
	for _, allowed := range config.AllowedOrigins {
		if matchOrigin(allowed, originURL) {
			return true
		}
	}

	return false
}

// matchOrigin checks if the origin matches the allowed pattern
func matchOrigin(allowed string, origin *url.URL) bool {
	// Check for wildcard port before parsing (e.g., http://localhost:*)
	if strings.Contains(allowed, ":*") {
		// Extract scheme and host from allowed pattern
		schemeEnd := strings.Index(allowed, "://")
		if schemeEnd == -1 {
			return false
		}
		scheme := allowed[:schemeEnd]
		hostWithWildcard := allowed[schemeEnd+3:]

		// Protocol must match
		if scheme != origin.Scheme {
			return false
		}

		// Extract host without port from pattern
		allowedHost := strings.TrimSuffix(hostWithWildcard, ":*")
		allowedHost = strings.ToLower(allowedHost)

		// Extract host without port from origin
		originHost := origin.Host
		if colonIdx := strings.Index(originHost, ":"); colonIdx != -1 {
			originHost = originHost[:colonIdx]
		}
		originHost = strings.ToLower(originHost)

		return allowedHost == originHost
	}

	// Parse allowed origin
	allowedURL, err := url.Parse(allowed)
	if err != nil {
		return false
	}

	// Protocol must match (http vs https)
	if allowedURL.Scheme != origin.Scheme {
		return false
	}

	// Check host with wildcard support
	if !matchHost(allowedURL.Host, origin.Host) {
		return false
	}

	return true
}

// matchHost checks if the origin host matches the allowed host pattern
// Supports wildcards for subdomains (*.example.com)
func matchHost(allowed, origin string) bool {
	// Case-insensitive comparison
	allowed = strings.ToLower(allowed)
	origin = strings.ToLower(origin)

	// Exact match
	if allowed == origin {
		return true
	}

	// Check for wildcard subdomain (e.g., *.example.com)
	if strings.HasPrefix(allowed, "*.") {
		baseDomain := strings.TrimPrefix(allowed, "*.")

		// Extract domain from origin (remove port if present)
		originHost := origin
		if colonIdx := strings.Index(origin, ":"); colonIdx != -1 {
			originHost = origin[:colonIdx]
		}

		// Must end with the base domain
		if !strings.HasSuffix(originHost, baseDomain) {
			return false
		}

		// Must not be the base domain itself (*.example.com should not match example.com)
		if originHost == baseDomain {
			return false
		}

		// Must have a subdomain (at least one more dot)
		// e.g., app.example.com or dev.app.example.com should match *.example.com
		prefix := strings.TrimSuffix(originHost, baseDomain)
		return len(prefix) > 0 && strings.HasSuffix(prefix, ".")
	}

	return false
}

// CheckOrigin returns a function suitable for websocket.Upgrader.CheckOrigin
func (c WebSocketConfig) CheckOrigin() func(r *http.Request) bool {
	return func(r *http.Request) bool {
		return ValidateOrigin(c, r)
	}
}

// loadWebSocketConfig loads WebSocket configuration from environment variables
func loadWebSocketConfig() WebSocketConfig {
	return WebSocketConfig{
		AllowedOrigins: getEnvAsSlice("WEBSOCKET_ALLOWED_ORIGINS", []string{
			"http://localhost:3000",
			"http://localhost:5173",
			"http://localhost:5174",
		}),
		MaxMessageSize:                getEnvAsInt64("WEBSOCKET_MAX_MESSAGE_SIZE", 512*1024),          // 512KB default
		MaxConnectionsPerWorkflow:     getEnvAsInt("WEBSOCKET_MAX_CONNECTIONS_PER_WORKFLOW", 50),      // 50 users per workflow
		ConnectionsPerTenantPerMinute: getEnvAsInt("WEBSOCKET_CONNECTIONS_PER_TENANT_PER_MINUTE", 60), // 60 connections/min
	}
}

// DefaultWebSocketConfig returns sensible defaults for WebSocket configuration
func DefaultWebSocketConfig() WebSocketConfig {
	return WebSocketConfig{
		AllowedOrigins: []string{
			"http://localhost:3000",
			"http://localhost:5173",
			"http://localhost:5174",
		},
		MaxMessageSize:                512 * 1024, // 512KB
		MaxConnectionsPerWorkflow:     50,
		ConnectionsPerTenantPerMinute: 60,
	}
}
