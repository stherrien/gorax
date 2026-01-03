package middleware

import (
	"context"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/gorax/gorax/internal/audit"
)

// responseRecorder wraps http.ResponseWriter to capture status code
type responseRecorder struct {
	http.ResponseWriter
	statusCode int
	written    bool
}

func (r *responseRecorder) WriteHeader(statusCode int) {
	if !r.written {
		r.statusCode = statusCode
		r.written = true
		r.ResponseWriter.WriteHeader(statusCode)
	}
}

func (r *responseRecorder) Write(b []byte) (int, error) {
	if !r.written {
		r.statusCode = http.StatusOK
		r.written = true
	}
	return r.ResponseWriter.Write(b)
}

// AuditMiddleware creates middleware that logs all API requests to the audit service
func AuditMiddleware(auditService *audit.Service, logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			startTime := time.Now()

			// Skip audit logging for health checks and static assets
			if shouldSkipAudit(r.URL.Path) {
				next.ServeHTTP(w, r)
				return
			}

			// Create response recorder to capture status code
			recorder := &responseRecorder{
				ResponseWriter: w,
				statusCode:     http.StatusOK,
				written:        false,
			}

			// Execute request
			next.ServeHTTP(recorder, r)

			// Log audit event asynchronously
			duration := time.Since(startTime)
			go logAuditEvent(r.Context(), auditService, r, recorder.statusCode, duration, logger)
		})
	}
}

func shouldSkipAudit(path string) bool {
	skipPaths := []string{
		"/health",
		"/ready",
		"/metrics",
		"/api/docs",
		"/swagger",
		"/favicon.ico",
	}

	for _, skip := range skipPaths {
		if strings.HasPrefix(path, skip) {
			return true
		}
	}

	return false
}

func logAuditEvent(
	ctx context.Context,
	auditService *audit.Service,
	r *http.Request,
	statusCode int,
	duration time.Duration,
	logger *slog.Logger,
) {
	// Extract context values
	tenantID, _ := ctx.Value("tenant_id").(string)
	userID, _ := ctx.Value("user_id").(string)
	userEmail, _ := ctx.Value("user_email").(string)

	// If no tenant/user in context, skip audit (likely public endpoint)
	if tenantID == "" && userID == "" {
		return
	}

	// Determine event category and type based on path and method
	category, eventType, action := categorizeRequest(r.URL.Path, r.Method)

	// Determine resource type and ID from URL
	resourceType, resourceID := extractResource(r.URL.Path)

	// Determine severity and status based on response code
	severity := determineSeverity(statusCode, r.Method)
	status := determineStatus(statusCode)

	// Create audit event
	event := &audit.AuditEvent{
		ID:           uuid.New().String(),
		TenantID:     tenantID,
		UserID:       userID,
		UserEmail:    userEmail,
		Category:     category,
		EventType:    eventType,
		Action:       action,
		ResourceType: resourceType,
		ResourceID:   resourceID,
		IPAddress:    getClientIP(r),
		UserAgent:    r.UserAgent(),
		Severity:     severity,
		Status:       status,
		Metadata: map[string]interface{}{
			"method":       r.Method,
			"path":         r.URL.Path,
			"query":        r.URL.RawQuery,
			"status_code":  statusCode,
			"duration_ms":  duration.Milliseconds(),
			"content_type": r.Header.Get("Content-Type"),
		},
		CreatedAt: time.Now(),
	}

	// Log error message for failed requests
	if statusCode >= 400 {
		event.ErrorMessage = http.StatusText(statusCode)
	}

	// Log the event (async)
	if err := auditService.LogEvent(ctx, event); err != nil {
		logger.Error("failed to log audit event", "error", err, "event_id", event.ID)
	}
}

func categorizeRequest(path, method string) (audit.Category, audit.EventType, string) {
	// Authentication endpoints
	if strings.Contains(path, "/login") || strings.Contains(path, "/logout") || strings.Contains(path, "/auth") {
		if strings.Contains(path, "/login") {
			return audit.CategoryAuthentication, audit.EventTypeLogin, "user_login"
		}
		if strings.Contains(path, "/logout") {
			return audit.CategoryAuthentication, audit.EventTypeLogout, "user_logout"
		}
		return audit.CategoryAuthentication, audit.EventTypeAccess, "authentication_action"
	}

	// SSO endpoints
	if strings.Contains(path, "/sso") {
		return audit.CategoryAuthentication, audit.EventTypeLogin, "sso_authentication"
	}

	// OAuth endpoints
	if strings.Contains(path, "/oauth") {
		return audit.CategoryIntegration, audit.EventTypeConfigure, "oauth_action"
	}

	// Credential endpoints
	if strings.Contains(path, "/credentials") {
		eventType := getEventTypeFromMethod(method)
		return audit.CategoryCredential, eventType, strings.ToLower(method) + "_credential"
	}

	// Workflow endpoints
	if strings.Contains(path, "/workflows") {
		if strings.Contains(path, "/execute") {
			return audit.CategoryWorkflow, audit.EventTypeExecute, "execute_workflow"
		}
		eventType := getEventTypeFromMethod(method)
		return audit.CategoryWorkflow, eventType, strings.ToLower(method) + "_workflow"
	}

	// Admin endpoints
	if strings.Contains(path, "/admin") {
		if strings.Contains(path, "/tenants") {
			eventType := getEventTypeFromMethod(method)
			return audit.CategoryUserManagement, eventType, strings.ToLower(method) + "_tenant"
		}
		if strings.Contains(path, "/audit") {
			return audit.CategorySystem, audit.EventTypeAccess, "view_audit_logs"
		}
		return audit.CategoryConfiguration, getEventTypeFromMethod(method), "admin_action"
	}

	// Data access (GET requests)
	if method == "GET" {
		return audit.CategoryDataAccess, audit.EventTypeRead, "view_resource"
	}

	// Configuration changes
	eventType := getEventTypeFromMethod(method)
	return audit.CategoryConfiguration, eventType, strings.ToLower(method) + "_resource"
}

func getEventTypeFromMethod(method string) audit.EventType {
	switch method {
	case "POST":
		return audit.EventTypeCreate
	case "GET":
		return audit.EventTypeRead
	case "PUT", "PATCH":
		return audit.EventTypeUpdate
	case "DELETE":
		return audit.EventTypeDelete
	default:
		return audit.EventTypeAccess
	}
}

func extractResource(path string) (resourceType, resourceID string) {
	parts := strings.Split(strings.Trim(path, "/"), "/")

	// Find resource type (usually after /api/v1/)
	for i, part := range parts {
		if part == "v1" && i+1 < len(parts) {
			resourceType = parts[i+1]
			// Check if there's an ID after the resource type
			if i+2 < len(parts) && !strings.Contains(parts[i+2], "?") {
				resourceID = parts[i+2]
			}
			break
		}
	}

	// If no resource type found, use the first non-api part
	if resourceType == "" && len(parts) > 0 {
		resourceType = parts[0]
		if len(parts) > 1 {
			resourceID = parts[1]
		}
	}

	return resourceType, resourceID
}

func determineSeverity(statusCode int, method string) audit.Severity {
	// Critical for failed authentication or authorization
	if statusCode == http.StatusUnauthorized || statusCode == http.StatusForbidden {
		return audit.SeverityCritical
	}

	// Error for server errors
	if statusCode >= 500 {
		return audit.SeverityError
	}

	// Warning for client errors (except 404)
	if statusCode >= 400 && statusCode != http.StatusNotFound {
		return audit.SeverityWarning
	}

	// Warning for destructive operations (DELETE, even if successful)
	if method == "DELETE" {
		return audit.SeverityWarning
	}

	// Info for everything else
	return audit.SeverityInfo
}

func determineStatus(statusCode int) audit.Status {
	if statusCode >= 200 && statusCode < 300 {
		return audit.StatusSuccess
	}
	if statusCode >= 400 {
		return audit.StatusFailure
	}
	return audit.StatusPartial
}

func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header first (proxy/load balancer)
	xff := r.Header.Get("X-Forwarded-For")
	if xff != "" {
		// X-Forwarded-For can contain multiple IPs, take the first one
		ips := strings.Split(xff, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}

	// Check X-Real-IP header
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}

	// Fall back to RemoteAddr
	ip := r.RemoteAddr
	// Remove port if present
	if idx := strings.LastIndex(ip, ":"); idx != -1 {
		ip = ip[:idx]
	}

	return ip
}
