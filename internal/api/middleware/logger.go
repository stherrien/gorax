package middleware

import (
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5/middleware"
)

// HTTPLoggerConfig holds configuration for HTTP logging
type HTTPLoggerConfig struct {
	// LogLevel is the log level for successful HTTP requests (2xx)
	// Typically "debug" in development to reduce noise, "info" in production
	LogLevel slog.Level
}

// StructuredLogger returns a middleware that logs requests with slog
// Uses configured log level for successful requests, WARN for client errors, ERROR for server errors
func StructuredLogger(logger *slog.Logger) func(next http.Handler) http.Handler {
	// Default to DEBUG for HTTP access logs to reduce noise
	return StructuredLoggerWithConfig(logger, HTTPLoggerConfig{
		LogLevel: slog.LevelDebug,
	})
}

// StructuredLoggerWithConfig returns a middleware with custom logging configuration
func StructuredLoggerWithConfig(logger *slog.Logger, config HTTPLoggerConfig) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

			defer func() {
				status := ww.Status()
				duration := time.Since(start)

				// Skip logging for health checks at any level
				if shouldSkipLogging(r.URL.Path) {
					return
				}

				// Prepare common log attributes
				attrs := []any{
					"method", r.Method,
					"path", r.URL.Path,
					"status", status,
					"bytes", ww.BytesWritten(),
					"duration_ms", duration.Milliseconds(),
					"request_id", middleware.GetReqID(r.Context()),
					"remote_addr", r.RemoteAddr,
					"user_agent", r.UserAgent(),
				}

				// Log at different levels based on response status
				if status >= 500 {
					// Server errors (5xx) - ERROR level
					logger.Error("http server error", attrs...)
				} else if status >= 400 {
					// Client errors (4xx) - WARN level
					logger.Warn("http client error", attrs...)
				} else {
					// Success (2xx, 3xx) - Use configured level (typically DEBUG)
					logger.Log(r.Context(), config.LogLevel, "http request", attrs...)
				}
			}()

			next.ServeHTTP(ww, r)
		})
	}
}

// shouldSkipLogging returns true for paths that should not be logged at any level
func shouldSkipLogging(path string) bool {
	noisyPaths := []string{
		"/health",
		"/ready",
		"/favicon.ico",
	}

	for _, noisy := range noisyPaths {
		if strings.HasPrefix(path, noisy) {
			return true
		}
	}

	return false
}
