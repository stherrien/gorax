package errortracking

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gorax/gorax/internal/config"
)

func TestInitialize(t *testing.T) {
	tests := []struct {
		name      string
		cfg       config.ObservabilityConfig
		wantError bool
	}{
		{
			name: "successful initialization with valid config",
			cfg: config.ObservabilityConfig{
				SentryEnabled:     true,
				SentryDSN:         "https://examplePublicKey@o0.ingest.sentry.io/0",
				SentryEnvironment: "test",
				SentrySampleRate:  1.0,
			},
			wantError: false,
		},
		{
			name: "disabled sentry skips initialization",
			cfg: config.ObservabilityConfig{
				SentryEnabled: false,
			},
			wantError: false,
		},
		{
			name: "invalid DSN returns error",
			cfg: config.ObservabilityConfig{
				SentryEnabled:     true,
				SentryDSN:         "invalid-dsn",
				SentryEnvironment: "test",
				SentrySampleRate:  1.0,
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Cleanup any previous Sentry client
			sentry.Flush(time.Second)

			tracker, err := Initialize(tt.cfg)
			defer func() {
				if tracker != nil {
					tracker.Close()
				}
			}()

			if tt.wantError {
				assert.Error(t, err)
				assert.Nil(t, tracker)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, tracker)
			}
		})
	}
}

func TestTracker_CaptureError(t *testing.T) {
	// Create tracker with mock transport
	tracker := &Tracker{
		enabled: true,
		client:  &mockSentryHub{},
	}

	tests := []struct {
		name string
		err  error
		ctx  context.Context
	}{
		{
			name: "capture simple error",
			err:  errors.New("test error"),
			ctx:  context.Background(),
		},
		{
			name: "capture error with context values",
			err:  errors.New("context error"),
			ctx:  contextWithValues(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			eventID := tracker.CaptureError(tt.ctx, tt.err)
			assert.NotEmpty(t, eventID)
		})
	}
}

func TestTracker_CaptureErrorWithTags(t *testing.T) {
	tracker := &Tracker{
		enabled: true,
		client:  &mockSentryHub{},
	}

	tags := map[string]string{
		"tenant_id":    "tenant-123",
		"execution_id": "exec-456",
	}

	eventID := tracker.CaptureErrorWithTags(context.Background(), errors.New("tagged error"), tags)
	assert.NotEmpty(t, eventID)
}

func TestTracker_CaptureMessage(t *testing.T) {
	tracker := &Tracker{
		enabled: true,
		client:  &mockSentryHub{},
	}

	tests := []struct {
		name    string
		message string
		level   Level
	}{
		{
			name:    "info level message",
			message: "info message",
			level:   LevelInfo,
		},
		{
			name:    "warning level message",
			message: "warning message",
			level:   LevelWarning,
		},
		{
			name:    "error level message",
			message: "error message",
			level:   LevelError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			eventID := tracker.CaptureMessage(context.Background(), tt.message, tt.level)
			assert.NotEmpty(t, eventID)
		})
	}
}

func TestTracker_AddBreadcrumb(t *testing.T) {
	tracker := &Tracker{
		enabled: true,
		client:  &mockSentryHub{},
	}

	breadcrumb := Breadcrumb{
		Type:     "http",
		Category: "request",
		Message:  "HTTP request started",
		Level:    LevelInfo,
		Data: map[string]interface{}{
			"method": "GET",
			"url":    "/api/workflows",
		},
	}

	// Should not panic
	tracker.AddBreadcrumb(context.Background(), breadcrumb)
}

func TestTracker_SetUser(t *testing.T) {
	tracker := &Tracker{
		enabled: true,
		client:  &mockSentryHub{},
	}

	user := User{
		ID:        "user-123",
		Email:     "test@example.com",
		Username:  "testuser",
		IPAddress: "192.168.1.1",
	}

	// Should not panic
	tracker.SetUser(context.Background(), user)
}

func TestTracker_RecoverPanic(t *testing.T) {
	tracker := &Tracker{
		enabled: true,
		client:  &mockSentryHub{},
	}

	tests := []struct {
		name        string
		panicValue  interface{}
		shouldPanic bool
	}{
		{
			name:        "recover from string panic",
			panicValue:  "panic message",
			shouldPanic: false,
		},
		{
			name:        "recover from error panic",
			panicValue:  errors.New("panic error"),
			shouldPanic: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil && !tt.shouldPanic {
					t.Errorf("unexpected panic: %v", r)
				}
			}()

			defer tracker.RecoverPanic(context.Background())

			if tt.panicValue != nil {
				panic(tt.panicValue)
			}
		})
	}
}

func TestTracker_Flush(t *testing.T) {
	tracker := &Tracker{
		enabled: true,
		client:  &mockSentryHub{},
	}

	// Should not panic
	tracker.Flush(2 * time.Second)
}

func TestTracker_Close(t *testing.T) {
	tracker := &Tracker{
		enabled: true,
		client:  &mockSentryHub{},
	}

	// Should not panic
	tracker.Close()
}

func TestTracker_WithScope(t *testing.T) {
	tracker := &Tracker{
		enabled: true,
		client:  &mockSentryHub{},
	}

	captured := false
	tracker.WithScope(context.Background(), func(scope *Scope) {
		scope.SetTag("test", "value")
		scope.SetExtra("key", "value")
		captured = true
	})

	assert.True(t, captured)
}

func TestEnrichContext(t *testing.T) {
	tests := []struct {
		name     string
		ctx      context.Context
		wantTags map[string]string
	}{
		{
			name: "extract tenant and user from context",
			ctx:  contextWithValues(),
			wantTags: map[string]string{
				"tenant_id": "test-tenant",
				"user_id":   "test-user",
			},
		},
		{
			name:     "empty context returns empty tags",
			ctx:      context.Background(),
			wantTags: map[string]string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tags := enrichContext(tt.ctx)
			for key, value := range tt.wantTags {
				assert.Equal(t, value, tags[key])
			}
		})
	}
}

func TestExtractRequestContext(t *testing.T) {
	req, err := http.NewRequest("GET", "/api/workflows", nil)
	require.NoError(t, err)

	req.Header.Set("User-Agent", "test-agent")
	req.Header.Set("X-Request-ID", "req-123")
	req.RemoteAddr = "192.168.1.1:1234"

	data := extractRequestContext(req)

	assert.Equal(t, "GET", data["method"])
	assert.Equal(t, "/api/workflows", data["url"])
	assert.Equal(t, "test-agent", data["user_agent"])
	assert.Contains(t, data, "headers")
}

func TestLevelConversion(t *testing.T) {
	tests := []struct {
		level    Level
		expected sentry.Level
	}{
		{LevelDebug, sentry.LevelDebug},
		{LevelInfo, sentry.LevelInfo},
		{LevelWarning, sentry.LevelWarning},
		{LevelError, sentry.LevelError},
		{LevelFatal, sentry.LevelFatal},
	}

	for _, tt := range tests {
		t.Run(string(tt.level), func(t *testing.T) {
			result := toSentryLevel(tt.level)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Helper functions and mocks

func contextWithValues() context.Context {
	ctx := context.Background()
	ctx = context.WithValue(ctx, "tenant_id", "test-tenant")
	ctx = context.WithValue(ctx, "user_id", "test-user")
	ctx = context.WithValue(ctx, "execution_id", "exec-123")
	return ctx
}

// mockSentryHub implements a minimal Hub interface for testing
type mockSentryHub struct{}

func (m *mockSentryHub) CaptureException(exception error) *sentry.EventID {
	id := sentry.EventID("mock-event-id")
	return &id
}

func (m *mockSentryHub) CaptureMessage(message string) *sentry.EventID {
	id := sentry.EventID("mock-event-id")
	return &id
}

func (m *mockSentryHub) AddBreadcrumb(breadcrumb *sentry.Breadcrumb, hint *sentry.BreadcrumbHint) {
}

func (m *mockSentryHub) ConfigureScope(f func(*sentry.Scope)) {
	scope := sentry.NewScope()
	f(scope)
}

func (m *mockSentryHub) WithScope(f func(*sentry.Scope)) {
	scope := sentry.NewScope()
	f(scope)
}

func (m *mockSentryHub) Flush(timeout time.Duration) bool {
	return true
}

func (m *mockSentryHub) Clone() *sentry.Hub {
	return nil
}

func (m *mockSentryHub) Recover(err interface{}) *sentry.EventID {
	id := sentry.EventID("mock-event-id")
	return &id
}
