package middleware

import (
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/gorax/gorax/internal/ratelimit"
)

// RateLimitConfig holds configuration for rate limiting
type RateLimitConfig struct {
	RequestsPerMinute int64 // Max requests per minute
	RequestsPerHour   int64 // Max requests per hour
	RequestsPerDay    int64 // Max requests per day
	EnabledForPaths   []string
	ExcludedPaths     []string
}

// DefaultRateLimitConfig returns sensible defaults
func DefaultRateLimitConfig() RateLimitConfig {
	return RateLimitConfig{
		RequestsPerMinute: 60,
		RequestsPerHour:   1000,
		RequestsPerDay:    10000,
		EnabledForPaths:   []string{"/api/"},
		ExcludedPaths:     []string{"/api/health", "/api/metrics"},
	}
}

// RateLimitMiddleware creates a middleware that enforces rate limits per tenant
func RateLimitMiddleware(redisClient *redis.Client, config RateLimitConfig, logger *slog.Logger) func(http.Handler) http.Handler {
	limiter := ratelimit.NewSlidingWindowLimiter(redisClient)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get tenant from context
			tenant := GetTenant(r)
			if tenant == nil {
				// No tenant, skip rate limiting (e.g., public endpoints)
				next.ServeHTTP(w, r)
				return
			}

			tenantID := tenant.ID

			// Check if path should be rate limited
			if shouldSkipRateLimit(r.URL.Path, config) {
				next.ServeHTTP(w, r)
				return
			}

			ctx := r.Context()

			// Check rate limits in order: minute, hour, day
			limits := []struct {
				name   string
				limit  int64
				window time.Duration
			}{
				{"per_minute", config.RequestsPerMinute, time.Minute},
				{"per_hour", config.RequestsPerHour, time.Hour},
				{"per_day", config.RequestsPerDay, 24 * time.Hour},
			}

			for _, limit := range limits {
				if limit.limit <= 0 {
					continue // Skip disabled limits
				}

				allowed, err := limiter.Allow(ctx, tenantID, limit.limit, limit.window)
				if err != nil {
					logger.Error("rate limit check failed",
						"error", err,
						"tenant_id", tenantID,
						"limit", limit.name,
					)
					// On error, allow the request (fail open)
					next.ServeHTTP(w, r)
					return
				}

				if !allowed {
					logger.Warn("rate limit exceeded",
						"tenant_id", tenantID,
						"limit", limit.name,
						"path", r.URL.Path,
					)

					// Get current usage for headers
					usage, _ := limiter.GetUsage(ctx, tenantID, limit.window)

					// Set rate limit headers
					w.Header().Set("X-RateLimit-Limit", formatInt64(limit.limit))
					w.Header().Set("X-RateLimit-Remaining", "0")
					w.Header().Set("X-RateLimit-Used", formatInt64(usage))
					w.Header().Set("X-RateLimit-Window", limit.window.String())
					w.Header().Set("Retry-After", formatInt64(int64(limit.window.Seconds())))

					http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
					return
				}
			}

			// All limits passed - add usage headers
			w.Header().Set("X-RateLimit-Limit-Minute", formatInt64(config.RequestsPerMinute))
			w.Header().Set("X-RateLimit-Limit-Hour", formatInt64(config.RequestsPerHour))
			w.Header().Set("X-RateLimit-Limit-Day", formatInt64(config.RequestsPerDay))

			next.ServeHTTP(w, r)
		})
	}
}

// shouldSkipRateLimit checks if a path should skip rate limiting
func shouldSkipRateLimit(path string, config RateLimitConfig) bool {
	// Check excluded paths
	for _, excluded := range config.ExcludedPaths {
		if matchPath(path, excluded) {
			return true
		}
	}

	// Check if path is in enabled paths
	for _, enabled := range config.EnabledForPaths {
		if matchPath(path, enabled) {
			return false // Should be rate limited
		}
	}

	// Not in enabled paths, skip
	return true
}

// matchPath checks if a path matches a pattern (simple prefix matching)
func matchPath(path, pattern string) bool {
	if len(pattern) == 0 {
		return false
	}

	// Simple prefix match
	if len(path) >= len(pattern) {
		return path[:len(pattern)] == pattern
	}

	return false
}

// formatInt64 converts int64 to string for headers
func formatInt64(n int64) string {
	return strconv.FormatInt(n, 10)
}
