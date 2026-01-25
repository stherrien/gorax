package security

import (
	"context"
	"log/slog"
	"net/http"
	"sync"
	"time"
)

// SecurityEventType represents types of security events
type SecurityEventType string

const (
	// Authentication events
	EventAuthSuccess    SecurityEventType = "auth.success"
	EventAuthFailure    SecurityEventType = "auth.failure"
	EventAuthLockout    SecurityEventType = "auth.lockout"
	EventAuthUnlock     SecurityEventType = "auth.unlock"
	EventSessionCreated SecurityEventType = "session.created"
	EventSessionExpired SecurityEventType = "session.expired"
	EventSessionRevoked SecurityEventType = "session.revoked"

	// Authorization events
	EventAuthzDenied   SecurityEventType = "authz.denied"
	EventAuthzElevated SecurityEventType = "authz.elevated"

	// Input validation events
	EventValidationFailed SecurityEventType = "validation.failed"
	EventXSSAttempt       SecurityEventType = "xss.attempt"
	EventSQLiAttempt      SecurityEventType = "sqli.attempt"
	EventCSRFViolation    SecurityEventType = "csrf.violation"

	// Rate limiting events
	EventRateLimited SecurityEventType = "rate.limited"
	EventQuotaExceed SecurityEventType = "quota.exceeded"

	// Suspicious activity events
	EventSuspiciousIP       SecurityEventType = "suspicious.ip"
	EventSuspiciousActivity SecurityEventType = "suspicious.activity"
	EventBruteForce         SecurityEventType = "brute.force"

	// Data access events
	EventSensitiveAccess  SecurityEventType = "sensitive.access"
	EventBulkDataAccess   SecurityEventType = "bulk.data.access"
	EventUnauthorizedData SecurityEventType = "unauthorized.data"

	// Configuration events
	EventConfigChange   SecurityEventType = "config.change"
	EventSecurityPolicy SecurityEventType = "security.policy"
)

// SecurityEventSeverity represents the severity of a security event
type SecurityEventSeverity string

const (
	SeverityInfo     SecurityEventSeverity = "info"
	SeverityWarning  SecurityEventSeverity = "warning"
	SeverityHigh     SecurityEventSeverity = "high"
	SeverityCritical SecurityEventSeverity = "critical"
)

// SecurityEvent represents a security-related event
type SecurityEvent struct {
	ID        string                `json:"id"`
	Type      SecurityEventType     `json:"type"`
	Severity  SecurityEventSeverity `json:"severity"`
	Timestamp time.Time             `json:"timestamp"`
	TenantID  string                `json:"tenant_id,omitempty"`
	UserID    string                `json:"user_id,omitempty"`
	IPAddress string                `json:"ip_address,omitempty"`
	UserAgent string                `json:"user_agent,omitempty"`
	RequestID string                `json:"request_id,omitempty"`
	Resource  string                `json:"resource,omitempty"`
	Action    string                `json:"action,omitempty"`
	Success   bool                  `json:"success"`
	Message   string                `json:"message"`
	Details   map[string]any        `json:"details,omitempty"`
}

// SecurityMonitor provides security event monitoring and alerting
type SecurityMonitor struct {
	logger        *slog.Logger
	eventHandlers []SecurityEventHandler
	eventBuffer   chan SecurityEvent
	bufferSize    int
	stopCh        chan struct{}
	wg            sync.WaitGroup

	// Thresholds for automatic alerts
	thresholds SecurityThresholds
}

// SecurityEventHandler processes security events
type SecurityEventHandler interface {
	Handle(ctx context.Context, event SecurityEvent) error
}

// SecurityThresholds defines thresholds for automatic alerting
type SecurityThresholds struct {
	// MaxFailedAuthPerMinute triggers alert when exceeded
	MaxFailedAuthPerMinute int
	// MaxRateLimitsPerMinute triggers alert when exceeded
	MaxRateLimitsPerMinute int
	// MaxValidationFailuresPerMinute triggers alert when exceeded
	MaxValidationFailuresPerMinute int
}

// DefaultSecurityThresholds returns sensible default thresholds
func DefaultSecurityThresholds() SecurityThresholds {
	return SecurityThresholds{
		MaxFailedAuthPerMinute:         10,
		MaxRateLimitsPerMinute:         100,
		MaxValidationFailuresPerMinute: 50,
	}
}

// NewSecurityMonitor creates a new security monitor
func NewSecurityMonitor(logger *slog.Logger, bufferSize int) *SecurityMonitor {
	if bufferSize <= 0 {
		bufferSize = 1000
	}

	m := &SecurityMonitor{
		logger:      logger,
		eventBuffer: make(chan SecurityEvent, bufferSize),
		bufferSize:  bufferSize,
		stopCh:      make(chan struct{}),
		thresholds:  DefaultSecurityThresholds(),
	}

	// Start processing goroutine
	m.wg.Add(1)
	go m.processEvents()

	return m
}

// AddHandler adds an event handler
func (m *SecurityMonitor) AddHandler(handler SecurityEventHandler) {
	m.eventHandlers = append(m.eventHandlers, handler)
}

// RecordEvent records a security event
func (m *SecurityMonitor) RecordEvent(event SecurityEvent) {
	// Set timestamp if not set
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}

	// Log immediately based on severity
	m.logEvent(event)

	// Send to buffer for async processing
	select {
	case m.eventBuffer <- event:
		// Event buffered successfully
	default:
		// Buffer full - log warning and drop event
		m.logger.Warn("security event buffer full, event dropped",
			"event_type", event.Type,
			"severity", event.Severity,
		)
	}
}

// RecordAuthFailure is a convenience method for auth failures
func (m *SecurityMonitor) RecordAuthFailure(r *http.Request, userID, reason string) {
	m.RecordEvent(SecurityEvent{
		Type:      EventAuthFailure,
		Severity:  SeverityWarning,
		UserID:    userID,
		IPAddress: getClientIP(r),
		UserAgent: r.UserAgent(),
		RequestID: getRequestID(r),
		Success:   false,
		Message:   reason,
		Details: map[string]any{
			"path":   r.URL.Path,
			"method": r.Method,
		},
	})
}

// RecordAuthSuccess is a convenience method for auth successes
func (m *SecurityMonitor) RecordAuthSuccess(r *http.Request, userID string) {
	m.RecordEvent(SecurityEvent{
		Type:      EventAuthSuccess,
		Severity:  SeverityInfo,
		UserID:    userID,
		IPAddress: getClientIP(r),
		UserAgent: r.UserAgent(),
		RequestID: getRequestID(r),
		Success:   true,
		Message:   "Authentication successful",
	})
}

// RecordValidationFailure records a validation failure event
func (m *SecurityMonitor) RecordValidationFailure(r *http.Request, field, reason string) {
	m.RecordEvent(SecurityEvent{
		Type:      EventValidationFailed,
		Severity:  SeverityWarning,
		IPAddress: getClientIP(r),
		UserAgent: r.UserAgent(),
		RequestID: getRequestID(r),
		Success:   false,
		Message:   reason,
		Details: map[string]any{
			"field":  field,
			"path":   r.URL.Path,
			"method": r.Method,
		},
	})
}

// RecordXSSAttempt records a potential XSS attempt
func (m *SecurityMonitor) RecordXSSAttempt(r *http.Request, payload string) {
	m.RecordEvent(SecurityEvent{
		Type:      EventXSSAttempt,
		Severity:  SeverityHigh,
		IPAddress: getClientIP(r),
		UserAgent: r.UserAgent(),
		RequestID: getRequestID(r),
		Success:   false,
		Message:   "Potential XSS attack blocked",
		Details: map[string]any{
			"path":    r.URL.Path,
			"method":  r.Method,
			"payload": truncateString(payload, 200),
		},
	})
}

// RecordCSRFViolation records a CSRF violation
func (m *SecurityMonitor) RecordCSRFViolation(r *http.Request) {
	m.RecordEvent(SecurityEvent{
		Type:      EventCSRFViolation,
		Severity:  SeverityHigh,
		IPAddress: getClientIP(r),
		UserAgent: r.UserAgent(),
		RequestID: getRequestID(r),
		Success:   false,
		Message:   "CSRF token validation failed",
		Details: map[string]any{
			"path":   r.URL.Path,
			"method": r.Method,
			"origin": r.Header.Get("Origin"),
		},
	})
}

// RecordRateLimited records a rate limit event
func (m *SecurityMonitor) RecordRateLimited(r *http.Request, limit string) {
	m.RecordEvent(SecurityEvent{
		Type:      EventRateLimited,
		Severity:  SeverityWarning,
		IPAddress: getClientIP(r),
		UserAgent: r.UserAgent(),
		RequestID: getRequestID(r),
		Success:   false,
		Message:   "Rate limit exceeded",
		Details: map[string]any{
			"path":  r.URL.Path,
			"limit": limit,
		},
	})
}

// RecordSensitiveAccess records access to sensitive data
func (m *SecurityMonitor) RecordSensitiveAccess(r *http.Request, userID, resource string) {
	m.RecordEvent(SecurityEvent{
		Type:      EventSensitiveAccess,
		Severity:  SeverityInfo,
		UserID:    userID,
		IPAddress: getClientIP(r),
		UserAgent: r.UserAgent(),
		RequestID: getRequestID(r),
		Resource:  resource,
		Success:   true,
		Message:   "Sensitive resource accessed",
	})
}

func (m *SecurityMonitor) logEvent(event SecurityEvent) {
	level := slog.LevelInfo
	switch event.Severity {
	case SeverityWarning:
		level = slog.LevelWarn
	case SeverityHigh, SeverityCritical:
		level = slog.LevelError
	}

	m.logger.Log(context.Background(), level, "security event",
		"event_type", event.Type,
		"severity", event.Severity,
		"user_id", event.UserID,
		"ip_address", event.IPAddress,
		"success", event.Success,
		"message", event.Message,
	)
}

func (m *SecurityMonitor) processEvents() {
	defer m.wg.Done()

	for {
		select {
		case event := <-m.eventBuffer:
			// Process through all handlers
			ctx := context.Background()
			for _, handler := range m.eventHandlers {
				if err := handler.Handle(ctx, event); err != nil {
					m.logger.Error("security event handler failed",
						"error", err,
						"event_type", event.Type,
					)
				}
			}
		case <-m.stopCh:
			// Drain remaining events
			for {
				select {
				case event := <-m.eventBuffer:
					ctx := context.Background()
					for _, handler := range m.eventHandlers {
						_ = handler.Handle(ctx, event)
					}
				default:
					return
				}
			}
		}
	}
}

// Close stops the security monitor
func (m *SecurityMonitor) Close() {
	close(m.stopCh)
	m.wg.Wait()
}

// Helper functions

func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header (may be spoofed)
	xff := r.Header.Get("X-Forwarded-For")
	if xff != "" {
		// Take the first IP in the chain
		return xff
	}

	// Check X-Real-IP header
	xri := r.Header.Get("X-Real-IP")
	if xri != "" {
		return xri
	}

	// Fall back to RemoteAddr
	return r.RemoteAddr
}

func getRequestID(r *http.Request) string {
	return r.Header.Get("X-Request-ID")
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// --- Log-based Security Event Handler ---

// LoggingEventHandler logs security events to structured logger
type LoggingEventHandler struct {
	logger *slog.Logger
}

// NewLoggingEventHandler creates a new logging event handler
func NewLoggingEventHandler(logger *slog.Logger) *LoggingEventHandler {
	return &LoggingEventHandler{logger: logger}
}

// Handle logs the security event
func (h *LoggingEventHandler) Handle(_ context.Context, event SecurityEvent) error {
	h.logger.Info("security_event",
		"id", event.ID,
		"type", event.Type,
		"severity", event.Severity,
		"timestamp", event.Timestamp,
		"tenant_id", event.TenantID,
		"user_id", event.UserID,
		"ip_address", event.IPAddress,
		"request_id", event.RequestID,
		"success", event.Success,
		"message", event.Message,
	)
	return nil
}

// --- Metrics Event Handler ---

// MetricsEventHandler updates metrics based on security events
type MetricsEventHandler struct {
	counters map[SecurityEventType]int
	mu       sync.Mutex
}

// NewMetricsEventHandler creates a new metrics event handler
func NewMetricsEventHandler() *MetricsEventHandler {
	return &MetricsEventHandler{
		counters: make(map[SecurityEventType]int),
	}
}

// Handle updates metrics for the security event
func (h *MetricsEventHandler) Handle(_ context.Context, event SecurityEvent) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.counters[event.Type]++
	return nil
}

// GetCount returns the count for an event type
func (h *MetricsEventHandler) GetCount(eventType SecurityEventType) int {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.counters[eventType]
}
