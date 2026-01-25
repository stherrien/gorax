package middleware

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDefaultRateLimitConfig(t *testing.T) {
	config := DefaultRateLimitConfig()

	assert.Equal(t, int64(60), config.RequestsPerMinute)
	assert.Equal(t, int64(1000), config.RequestsPerHour)
	assert.Equal(t, int64(10000), config.RequestsPerDay)
	assert.Contains(t, config.EnabledForPaths, "/api/")
	assert.Contains(t, config.ExcludedPaths, "/api/health")
	assert.Contains(t, config.ExcludedPaths, "/api/metrics")
}

func TestShouldSkipRateLimit(t *testing.T) {
	config := RateLimitConfig{
		EnabledForPaths: []string{"/api/"},
		ExcludedPaths:   []string{"/api/health", "/api/metrics"},
	}

	tests := []struct {
		name     string
		path     string
		expected bool
	}{
		{
			name:     "API path should not skip",
			path:     "/api/v1/workflows",
			expected: false,
		},
		{
			name:     "health endpoint should skip",
			path:     "/api/health",
			expected: true,
		},
		{
			name:     "metrics endpoint should skip",
			path:     "/api/metrics",
			expected: true,
		},
		{
			name:     "non-api path should skip",
			path:     "/static/css/style.css",
			expected: true,
		},
		{
			name:     "root path should skip",
			path:     "/",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := shouldSkipRateLimit(tt.path, config)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestShouldSkipRateLimit_EmptyConfig(t *testing.T) {
	config := RateLimitConfig{
		EnabledForPaths: []string{},
		ExcludedPaths:   []string{},
	}

	// With empty enabled paths, everything should skip
	assert.True(t, shouldSkipRateLimit("/api/v1/test", config))
}

func TestShouldSkipRateLimit_AllPathsEnabled(t *testing.T) {
	config := RateLimitConfig{
		EnabledForPaths: []string{"/"},
		ExcludedPaths:   []string{"/health"},
	}

	tests := []struct {
		name     string
		path     string
		expected bool
	}{
		{
			name:     "root path should not skip",
			path:     "/api/test",
			expected: false,
		},
		{
			name:     "health path should skip",
			path:     "/health",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := shouldSkipRateLimit(tt.path, config)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMatchPath(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		pattern  string
		expected bool
	}{
		{
			name:     "exact match",
			path:     "/api/health",
			pattern:  "/api/health",
			expected: true,
		},
		{
			name:     "prefix match",
			path:     "/api/v1/workflows",
			pattern:  "/api/",
			expected: true,
		},
		{
			name:     "no match",
			path:     "/static/file.js",
			pattern:  "/api/",
			expected: false,
		},
		{
			name:     "empty pattern",
			path:     "/api/test",
			pattern:  "",
			expected: false,
		},
		{
			name:     "pattern longer than path",
			path:     "/api",
			pattern:  "/api/v1/",
			expected: false,
		},
		{
			name:     "same length, no match",
			path:     "/api",
			pattern:  "/app",
			expected: false,
		},
		{
			name:     "root pattern",
			path:     "/anything",
			pattern:  "/",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := matchPath(tt.path, tt.pattern)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFormatInt64(t *testing.T) {
	tests := []struct {
		name     string
		input    int64
		expected string
	}{
		{
			name:     "positive number",
			input:    100,
			expected: "100",
		},
		{
			name:     "zero",
			input:    0,
			expected: "0",
		},
		{
			name:     "negative number",
			input:    -50,
			expected: "-50",
		},
		{
			name:     "large number",
			input:    1000000,
			expected: "1000000",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatInt64(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}
