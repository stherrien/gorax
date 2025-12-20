package middleware

import (
	"context"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/gorax/gorax/internal/errortracking"
)

// sentryTracker defines the interface for error tracking operations
// This allows us to mock the tracker in tests
type sentryTracker interface {
	AddBreadcrumb(ctx context.Context, breadcrumb errortracking.Breadcrumb)
	SetUser(ctx context.Context, user errortracking.User)
	CaptureErrorWithTags(ctx context.Context, err error, tags map[string]string) string
}

// SentryMiddleware enriches requests with Sentry context and handles panic recovery
func SentryMiddleware(tracker sentryTracker) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Recover from panics and report to Sentry
			defer func() {
				if rvr := recover(); rvr != nil {
					ctx := r.Context()

					// Convert panic to error
					var err error
					switch v := rvr.(type) {
					case error:
						err = v
					case string:
						err = errortracking.ErrPanic{Message: v}
					default:
						err = errortracking.ErrPanic{Message: "unknown panic"}
					}

					// Build context tags
					tags := make(map[string]string)
					tags["panic"] = "true"

					if user := GetUser(r); user != nil {
						tags["user_id"] = user.ID
						tags["tenant_id"] = user.TenantID
					}

					if tenant := GetTenant(r); tenant != nil {
						tags["tenant_id"] = tenant.ID
					}

					// Capture the panic as an error
					tracker.CaptureErrorWithTags(ctx, err, tags)

					// Return 500 error
					http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				}
			}()

			// Enrich Sentry context with request information
			enrichRequestContext(r, tracker)

			// Add HTTP request breadcrumb
			addRequestBreadcrumb(r, tracker)

			// Continue with the request
			next.ServeHTTP(w, r)
		})
	}
}

// enrichRequestContext extracts and sets user/tenant information in Sentry
func enrichRequestContext(r *http.Request, tracker sentryTracker) {
	ctx := r.Context()

	// Set user if available
	if user := GetUser(r); user != nil {
		tracker.SetUser(ctx, errortracking.User{
			ID:        user.ID,
			Email:     user.Email,
			Username:  user.Email, // Use email as username if no separate username
			IPAddress: extractIPAddress(r),
		})
	}
}

// addRequestBreadcrumb adds HTTP request information as a breadcrumb
func addRequestBreadcrumb(r *http.Request, tracker sentryTracker) {
	ctx := r.Context()

	// Build breadcrumb data
	data := map[string]interface{}{
		"method":      r.Method,
		"url":         r.URL.String(),
		"path":        r.URL.Path,
		"user_agent":  r.UserAgent(),
		"remote_addr": r.RemoteAddr,
	}

	// Add request ID if available
	if requestID := middleware.GetReqID(ctx); requestID != "" {
		data["request_id"] = requestID
	}

	// Add tenant ID if available
	if tenant := GetTenant(r); tenant != nil {
		data["tenant_id"] = tenant.ID
	}

	// Add user ID if available
	if user := GetUser(r); user != nil {
		data["user_id"] = user.ID
	}

	breadcrumb := errortracking.Breadcrumb{
		Type:      "http",
		Category:  "request",
		Message:   r.Method + " " + r.URL.Path,
		Level:     errortracking.LevelInfo,
		Data:      data,
		Timestamp: time.Now(),
	}

	tracker.AddBreadcrumb(ctx, breadcrumb)
}

// enrichSentryContext extracts context information for Sentry tags and extras
// This is used by handlers to manually enrich Sentry events
func enrichSentryContext(r *http.Request) (tags map[string]string, extras map[string]interface{}) {
	tags = make(map[string]string)
	extras = make(map[string]interface{})

	// HTTP context
	tags["http.method"] = r.Method
	tags["http.path"] = r.URL.Path
	extras["http.url"] = r.URL.String()

	// Request ID
	if requestID := r.Header.Get("X-Request-ID"); requestID != "" {
		tags["request.id"] = requestID
	}

	// User context
	if user := GetUser(r); user != nil {
		tags["user.id"] = user.ID
		extras["user.email"] = user.Email

		if user.TenantID != "" {
			tags["tenant.id"] = user.TenantID
		}
	}

	// Tenant context
	if tenant := GetTenant(r); tenant != nil {
		tags["tenant.id"] = tenant.ID
		extras["tenant.name"] = tenant.Name
	}

	return tags, extras
}

// extractIPAddress extracts the real client IP address from the request
func extractIPAddress(r *http.Request) string {
	// Try X-Real-IP first
	if ip := r.Header.Get("X-Real-IP"); ip != "" {
		return ip
	}

	// Try X-Forwarded-For
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		// X-Forwarded-For can contain multiple IPs, take the first one
		parts := strings.Split(forwarded, ",")
		if len(parts) > 0 {
			return strings.TrimSpace(parts[0])
		}
	}

	// Fall back to RemoteAddr
	// RemoteAddr includes port, so we need to strip it
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}

	return ip
}
