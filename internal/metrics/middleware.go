package metrics

import (
	"net/http"
	"regexp"
	"strconv"
	"time"
)

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

// WriteHeader captures the status code before writing it
func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// Write ensures we have a status code even if WriteHeader wasn't called
func (rw *responseWriter) Write(b []byte) (int, error) {
	return rw.ResponseWriter.Write(b)
}

// HTTPMetricsMiddleware creates middleware that records HTTP metrics
func HTTPMetricsMiddleware(m *Metrics) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Wrap response writer to capture status code
			rw := &responseWriter{
				ResponseWriter: w,
				statusCode:     http.StatusOK, // Default status
			}

			// Call next handler
			next.ServeHTTP(rw, r)

			// Record metrics
			duration := time.Since(start).Seconds()
			path := normalizeHTTPPath(r.URL.Path)
			status := strconv.Itoa(rw.statusCode)

			m.RecordHTTPRequest(r.Method, path, status, duration)
		})
	}
}

// normalizeHTTPPath normalizes URL paths to reduce cardinality
// Replaces UUIDs and numeric IDs with `:id` placeholder
func normalizeHTTPPath(path string) string {
	// Replace short hyphenated IDs first (like wf-123, hook-456)
	// This must come before numeric replacement to avoid partial matches
	shortIDRegex := regexp.MustCompile(`/[a-zA-Z]+-[0-9a-zA-Z]+(/|$)`)
	for shortIDRegex.MatchString(path) {
		path = shortIDRegex.ReplaceAllString(path, "/:id$1")
	}

	// Replace UUIDs (8-4-4-4-12 hex format)
	uuidRegex := regexp.MustCompile(`/[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}(/|$)`)
	path = uuidRegex.ReplaceAllString(path, "/:id$1")

	// Replace numeric-only IDs (but not version numbers like v1, v2)
	// Match /123/ or /123$ but not /v1
	numericIDRegex := regexp.MustCompile(`/(\d{2,})(/|$)`)
	path = numericIDRegex.ReplaceAllString(path, "/:id$2")

	return path
}
