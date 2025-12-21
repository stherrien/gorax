package errortracking

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/getsentry/sentry-go"

	"github.com/gorax/gorax/internal/config"
)

// Level represents the severity level
type Level string

const (
	LevelDebug   Level = "debug"
	LevelInfo    Level = "info"
	LevelWarning Level = "warning"
	LevelError   Level = "error"
	LevelFatal   Level = "fatal"
)

// ErrPanic represents a recovered panic
type ErrPanic struct {
	Message string
}

func (e ErrPanic) Error() string {
	return fmt.Sprintf("panic: %s", e.Message)
}

// Tracker wraps Sentry SDK for error tracking
type Tracker struct {
	enabled bool
	client  sentryHub
}

// sentryHub is an interface that matches the key methods we use from *sentry.Hub
// This allows us to mock Sentry in tests
type sentryHub interface {
	CaptureException(exception error) *sentry.EventID
	CaptureMessage(message string) *sentry.EventID
	AddBreadcrumb(breadcrumb *sentry.Breadcrumb, hint *sentry.BreadcrumbHint)
	ConfigureScope(f func(*sentry.Scope))
	WithScope(f func(*sentry.Scope))
	Flush(timeout time.Duration) bool
	Recover(err interface{}) *sentry.EventID
}

// Scope wraps Sentry scope for testing
type Scope = sentry.Scope

// User represents user information for Sentry
type User struct {
	ID        string
	Email     string
	Username  string
	IPAddress string
}

// Breadcrumb represents a breadcrumb for Sentry
type Breadcrumb struct {
	Type      string
	Category  string
	Message   string
	Level     Level
	Data      map[string]interface{}
	Timestamp time.Time
}

// Initialize sets up Sentry error tracking
func Initialize(cfg config.ObservabilityConfig) (*Tracker, error) {
	tracker := &Tracker{
		enabled: cfg.SentryEnabled,
	}

	// If not enabled, return early with disabled tracker
	if !cfg.SentryEnabled {
		return tracker, nil
	}

	// Initialize Sentry SDK
	err := sentry.Init(sentry.ClientOptions{
		Dsn:              cfg.SentryDSN,
		Environment:      cfg.SentryEnvironment,
		TracesSampleRate: cfg.SentrySampleRate,
		AttachStacktrace: true,
		BeforeSend: func(event *sentry.Event, hint *sentry.EventHint) *sentry.Event {
			// Custom event processing if needed
			return event
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Sentry: %w", err)
	}

	// Use the current hub
	tracker.client = sentry.CurrentHub()

	return tracker, nil
}

// CaptureError captures an error and sends it to Sentry
func (t *Tracker) CaptureError(ctx context.Context, err error) string {
	if !t.enabled || err == nil {
		return ""
	}

	// Enrich with context
	tags := enrichContext(ctx)

	var eventID *sentry.EventID
	t.client.WithScope(func(scope *sentry.Scope) {
		// Add context tags
		for key, value := range tags {
			scope.SetTag(key, value)
		}

		// Capture the error
		eventID = t.client.CaptureException(err)
	})

	if eventID != nil {
		return string(*eventID)
	}
	return ""
}

// CaptureErrorWithTags captures an error with custom tags
func (t *Tracker) CaptureErrorWithTags(ctx context.Context, err error, tags map[string]string) string {
	if !t.enabled || err == nil {
		return ""
	}

	// Merge context and custom tags
	contextTags := enrichContext(ctx)
	for key, value := range tags {
		contextTags[key] = value
	}

	var eventID *sentry.EventID
	t.client.WithScope(func(scope *sentry.Scope) {
		// Add all tags
		for key, value := range contextTags {
			scope.SetTag(key, value)
		}

		// Capture the error
		eventID = t.client.CaptureException(err)
	})

	if eventID != nil {
		return string(*eventID)
	}
	return ""
}

// CaptureMessage captures a message with a specific level
func (t *Tracker) CaptureMessage(ctx context.Context, message string, level Level) string {
	if !t.enabled {
		return ""
	}

	tags := enrichContext(ctx)

	var eventID *sentry.EventID
	t.client.WithScope(func(scope *sentry.Scope) {
		// Set level
		scope.SetLevel(toSentryLevel(level))

		// Add context tags
		for key, value := range tags {
			scope.SetTag(key, value)
		}

		// Capture the message
		eventID = t.client.CaptureMessage(message)
	})

	if eventID != nil {
		return string(*eventID)
	}
	return ""
}

// AddBreadcrumb adds a breadcrumb to the current scope
func (t *Tracker) AddBreadcrumb(ctx context.Context, breadcrumb Breadcrumb) {
	if !t.enabled {
		return
	}

	sentryBreadcrumb := &sentry.Breadcrumb{
		Type:      breadcrumb.Type,
		Category:  breadcrumb.Category,
		Message:   breadcrumb.Message,
		Level:     toSentryLevel(breadcrumb.Level),
		Data:      breadcrumb.Data,
		Timestamp: breadcrumb.Timestamp,
	}

	if sentryBreadcrumb.Timestamp.IsZero() {
		sentryBreadcrumb.Timestamp = time.Now()
	}

	t.client.AddBreadcrumb(sentryBreadcrumb, nil)
}

// SetUser sets the user context for error tracking
func (t *Tracker) SetUser(ctx context.Context, user User) {
	if !t.enabled {
		return
	}

	t.client.ConfigureScope(func(scope *sentry.Scope) {
		scope.SetUser(sentry.User{
			ID:        user.ID,
			Email:     user.Email,
			Username:  user.Username,
			IPAddress: user.IPAddress,
		})
	})
}

// RecoverPanic recovers from a panic and reports it to Sentry
func (t *Tracker) RecoverPanic(ctx context.Context) {
	if !t.enabled {
		return
	}

	if err := recover(); err != nil {
		tags := enrichContext(ctx)

		t.client.WithScope(func(scope *sentry.Scope) {
			// Add context tags
			for key, value := range tags {
				scope.SetTag(key, value)
			}

			// Capture the panic
			t.client.Recover(err)
		})

		// Flush immediately for panics
		t.client.Flush(2 * time.Second)
	}
}

// WithScope executes a function with a new Sentry scope
func (t *Tracker) WithScope(ctx context.Context, f func(*Scope)) {
	if !t.enabled {
		return
	}

	t.client.WithScope(func(scope *sentry.Scope) {
		// Enrich scope with context
		tags := enrichContext(ctx)
		for key, value := range tags {
			scope.SetTag(key, value)
		}

		// Execute user function
		f(scope)
	})
}

// Flush waits until the underlying client sends any buffered events
func (t *Tracker) Flush(timeout time.Duration) {
	if !t.enabled {
		return
	}

	t.client.Flush(timeout)
}

// Close flushes and closes the Sentry client
func (t *Tracker) Close() {
	if !t.enabled {
		return
	}

	t.client.Flush(5 * time.Second)
}

// enrichContext extracts relevant information from context
func enrichContext(ctx context.Context) map[string]string {
	tags := make(map[string]string)

	// Extract tenant ID
	if tenantID, ok := ctx.Value("tenant_id").(string); ok && tenantID != "" {
		tags["tenant_id"] = tenantID
	}

	// Extract user ID
	if userID, ok := ctx.Value("user_id").(string); ok && userID != "" {
		tags["user_id"] = userID
	}

	// Extract execution ID
	if executionID, ok := ctx.Value("execution_id").(string); ok && executionID != "" {
		tags["execution_id"] = executionID
	}

	// Extract workflow ID
	if workflowID, ok := ctx.Value("workflow_id").(string); ok && workflowID != "" {
		tags["workflow_id"] = workflowID
	}

	// Extract request ID
	if requestID, ok := ctx.Value("request_id").(string); ok && requestID != "" {
		tags["request_id"] = requestID
	}

	return tags
}

// extractRequestContext extracts HTTP request information
func extractRequestContext(r *http.Request) map[string]interface{} {
	data := make(map[string]interface{})

	data["method"] = r.Method
	data["url"] = r.URL.String()
	data["user_agent"] = r.UserAgent()
	data["remote_addr"] = r.RemoteAddr

	// Extract selected headers
	headers := make(map[string]string)
	for key, values := range r.Header {
		if len(values) > 0 {
			// Include common headers, exclude sensitive ones
			switch key {
			case "User-Agent", "Content-Type", "Accept", "X-Request-ID",
				"X-Tenant-ID", "Referer", "Origin":
				headers[key] = values[0]
			}
		}
	}
	data["headers"] = headers

	return data
}

// toSentryLevel converts our Level to sentry.Level
func toSentryLevel(level Level) sentry.Level {
	switch level {
	case LevelDebug:
		return sentry.LevelDebug
	case LevelInfo:
		return sentry.LevelInfo
	case LevelWarning:
		return sentry.LevelWarning
	case LevelError:
		return sentry.LevelError
	case LevelFatal:
		return sentry.LevelFatal
	default:
		return sentry.LevelError
	}
}
