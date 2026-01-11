package middleware

import (
	"bytes"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestStructuredLogger(t *testing.T) {
	tests := []struct {
		name          string
		statusCode    int
		path          string
		expectLogged  bool
		expectedLevel string
		expectedMsg   string
	}{
		{
			name:          "successful request logged at debug",
			statusCode:    200,
			path:          "/api/v1/workflows",
			expectLogged:  true,
			expectedLevel: "DEBUG",
			expectedMsg:   "http request",
		},
		{
			name:          "client error logged at warn",
			statusCode:    400,
			path:          "/api/v1/workflows",
			expectLogged:  true,
			expectedLevel: "WARN",
			expectedMsg:   "http client error",
		},
		{
			name:          "not found logged at warn",
			statusCode:    404,
			path:          "/api/v1/workflows/notfound",
			expectLogged:  true,
			expectedLevel: "WARN",
			expectedMsg:   "http client error",
		},
		{
			name:          "server error logged at error",
			statusCode:    500,
			path:          "/api/v1/workflows",
			expectLogged:  true,
			expectedLevel: "ERROR",
			expectedMsg:   "http server error",
		},
		{
			name:          "health check not logged",
			statusCode:    200,
			path:          "/health",
			expectLogged:  false,
			expectedLevel: "",
			expectedMsg:   "",
		},
		{
			name:          "ready check not logged",
			statusCode:    200,
			path:          "/ready",
			expectLogged:  false,
			expectedLevel: "",
			expectedMsg:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a buffer to capture log output
			var buf bytes.Buffer

			// Create a logger that writes to the buffer
			logger := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{
				Level: slog.LevelDebug, // Set to DEBUG to capture all levels
			}))

			// Create test handler that returns the specified status code
			testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
			})

			// Wrap with our logging middleware
			middleware := StructuredLogger(logger)
			handler := middleware(testHandler)

			// Create test request
			req := httptest.NewRequest("GET", tt.path, nil)
			rec := httptest.NewRecorder()

			// Execute request
			handler.ServeHTTP(rec, req)

			// Check logs
			logOutput := buf.String()

			if tt.expectLogged {
				if !strings.Contains(logOutput, tt.expectedLevel) {
					t.Errorf("expected log level %q, but got: %s", tt.expectedLevel, logOutput)
				}
				if !strings.Contains(logOutput, tt.expectedMsg) {
					t.Errorf("expected log message to contain %q, but got: %s", tt.expectedMsg, logOutput)
				}
				if !strings.Contains(logOutput, tt.path) {
					t.Errorf("expected log to contain path %q, but got: %s", tt.path, logOutput)
				}
			} else {
				if logOutput != "" {
					t.Errorf("expected no logs for %s, but got: %s", tt.path, logOutput)
				}
			}
		})
	}
}

func TestStructuredLoggerWithConfig(t *testing.T) {
	tests := []struct {
		name         string
		logLevel     slog.Level
		statusCode   int
		expectLogged bool
	}{
		{
			name:         "debug level shows successful requests",
			logLevel:     slog.LevelDebug,
			statusCode:   200,
			expectLogged: true,
		},
		{
			name:         "info level hides debug successful requests",
			logLevel:     slog.LevelInfo,
			statusCode:   200,
			expectLogged: false,
		},
		{
			name:         "info level shows errors",
			logLevel:     slog.LevelInfo,
			statusCode:   500,
			expectLogged: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer

			// Create logger with specified minimum level
			logger := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{
				Level: tt.logLevel,
			}))

			testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
			})

			// Create middleware with DEBUG for HTTP logs
			middleware := StructuredLoggerWithConfig(logger, HTTPLoggerConfig{
				LogLevel: slog.LevelDebug,
			})
			handler := middleware(testHandler)

			req := httptest.NewRequest("GET", "/api/v1/test", nil)
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			logOutput := buf.String()

			if tt.expectLogged && logOutput == "" {
				t.Errorf("expected log output, but got none")
			} else if !tt.expectLogged && logOutput != "" {
				t.Errorf("expected no log output, but got: %s", logOutput)
			}
		})
	}
}

func TestShouldSkipLogging(t *testing.T) {
	tests := []struct {
		name string
		path string
		want bool
	}{
		{"health endpoint", "/health", true},
		{"ready endpoint", "/ready", true},
		{"favicon", "/favicon.ico", true},
		{"api endpoint", "/api/v1/workflows", false},
		{"root", "/", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := shouldSkipLogging(tt.path); got != tt.want {
				t.Errorf("shouldSkipLogging(%q) = %v, want %v", tt.path, got, tt.want)
			}
		})
	}
}
