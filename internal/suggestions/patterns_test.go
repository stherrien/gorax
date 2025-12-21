package suggestions

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultPatterns(t *testing.T) {
	patterns := DefaultPatterns()

	// Should have multiple default patterns
	assert.NotEmpty(t, patterns)
	assert.GreaterOrEqual(t, len(patterns), 5, "should have at least 5 default patterns")

	// Each pattern should have required fields
	for _, p := range patterns {
		assert.NotEmpty(t, p.Name, "pattern should have a name")
		assert.NotEmpty(t, p.Category, "pattern should have a category")
		assert.NotEmpty(t, p.SuggestionType, "pattern should have a suggestion type")
		assert.NotEmpty(t, p.SuggestionTitle, "pattern should have a suggestion title")
		assert.NotEmpty(t, p.SuggestionDescription, "pattern should have a suggestion description")
		assert.NotEmpty(t, p.SuggestionConfidence, "pattern should have a confidence level")

		// Should have at least one matching criteria
		hasMatchingCriteria := len(p.MessagePatterns) > 0 || len(p.HTTPCodes) > 0
		assert.True(t, hasMatchingCriteria, "pattern %s should have message patterns or HTTP codes", p.Name)
	}
}

func TestPatternMatcher_MatchConnectionRefused(t *testing.T) {
	matcher := NewPatternMatcher(DefaultPatterns())

	tests := []struct {
		name    string
		errMsg  string
		matches bool
	}{
		{"connection refused lowercase", "connection refused", true},
		{"ECONNREFUSED", "ECONNREFUSED: connection error", true},
		{"dial tcp connection refused", "dial tcp 127.0.0.1:8080: connection refused", true},
		{"unrelated error", "file not found", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errCtx := &ErrorContext{
				ExecutionID:  "exec-123",
				WorkflowID:   "wf-456",
				NodeID:       "node-789",
				NodeType:     "action:http",
				ErrorMessage: tt.errMsg,
			}

			matches := matcher.Match(errCtx)
			if tt.matches {
				require.NotEmpty(t, matches, "should match pattern")
				assert.Equal(t, ErrorCategoryNetwork, matches[0].Category)
			}
		})
	}
}

func TestPatternMatcher_MatchHTTPCodes(t *testing.T) {
	matcher := NewPatternMatcher(DefaultPatterns())

	tests := []struct {
		name     string
		httpCode int
		category ErrorCategory
		matches  bool
	}{
		{"401 Unauthorized", 401, ErrorCategoryAuth, true},
		{"403 Forbidden", 403, ErrorCategoryAuth, true},
		{"429 Rate Limit", 429, ErrorCategoryRateLimit, true},
		{"500 Internal Server Error", 500, ErrorCategoryExternal, true},
		{"502 Bad Gateway", 502, ErrorCategoryExternal, true},
		{"503 Service Unavailable", 503, ErrorCategoryExternal, true},
		{"504 Gateway Timeout", 504, ErrorCategoryTimeout, true},
		{"408 Request Timeout", 408, ErrorCategoryTimeout, true},
		{"200 OK", 200, "", false},
		{"201 Created", 201, "", false},
		{"404 Not Found", 404, "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errCtx := &ErrorContext{
				ExecutionID:  "exec-123",
				WorkflowID:   "wf-456",
				NodeID:       "node-789",
				NodeType:     "action:http",
				ErrorMessage: "HTTP error",
				HTTPStatus:   tt.httpCode,
			}

			matches := matcher.Match(errCtx)
			if tt.matches {
				require.NotEmpty(t, matches, "should match pattern for HTTP %d", tt.httpCode)
				assert.Equal(t, tt.category, matches[0].Category)
			} else {
				// Either no matches or no match for this specific code
				for _, m := range matches {
					assert.NotEqual(t, tt.httpCode, m.MatchedHTTPCode)
				}
			}
		})
	}
}

func TestPatternMatcher_MatchTimeout(t *testing.T) {
	matcher := NewPatternMatcher(DefaultPatterns())

	tests := []struct {
		name    string
		errMsg  string
		matches bool
	}{
		{"timeout keyword", "request timeout", true},
		{"timed out", "connection timed out", true},
		{"deadline exceeded", "context deadline exceeded", true},
		{"context deadline", "context deadline exceeded: operation cancelled", true},
		{"unrelated", "invalid json", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errCtx := &ErrorContext{
				ExecutionID:  "exec-123",
				WorkflowID:   "wf-456",
				NodeID:       "node-789",
				NodeType:     "action:http",
				ErrorMessage: tt.errMsg,
			}

			matches := matcher.Match(errCtx)
			if tt.matches {
				require.NotEmpty(t, matches, "should match timeout pattern")
				assert.Equal(t, ErrorCategoryTimeout, matches[0].Category)
			}
		})
	}
}

func TestPatternMatcher_MatchRateLimit(t *testing.T) {
	matcher := NewPatternMatcher(DefaultPatterns())

	tests := []struct {
		name    string
		errMsg  string
		matches bool
	}{
		{"rate limit", "rate limit exceeded", true},
		{"too many requests", "too many requests, slow down", true},
		{"throttle", "request throttled", true},
		{"exceeded limit", "api call exceeded limit", true},
		{"unrelated", "connection error", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errCtx := &ErrorContext{
				ExecutionID:  "exec-123",
				WorkflowID:   "wf-456",
				NodeID:       "node-789",
				NodeType:     "action:http",
				ErrorMessage: tt.errMsg,
			}

			matches := matcher.Match(errCtx)
			if tt.matches {
				require.NotEmpty(t, matches, "should match rate limit pattern")
				assert.Equal(t, ErrorCategoryRateLimit, matches[0].Category)
			}
		})
	}
}

func TestPatternMatcher_MatchJSONError(t *testing.T) {
	matcher := NewPatternMatcher(DefaultPatterns())

	tests := []struct {
		name    string
		errMsg  string
		matches bool
	}{
		{"invalid json", "invalid json in response", true},
		{"json parse error", "json: parse error at offset 42", true},
		{"unexpected token", "unexpected token < at position 0", true},
		{"syntax error json", "syntax error in json", true},
		{"invalid character", "invalid character '<' looking for beginning of value", true},
		{"unrelated", "connection refused", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errCtx := &ErrorContext{
				ExecutionID:  "exec-123",
				WorkflowID:   "wf-456",
				NodeID:       "node-789",
				NodeType:     "action:http",
				ErrorMessage: tt.errMsg,
			}

			matches := matcher.Match(errCtx)
			if tt.matches {
				require.NotEmpty(t, matches, "should match JSON error pattern")
				assert.Equal(t, ErrorCategoryData, matches[0].Category)
			}
		})
	}
}

func TestPatternMatcher_MatchDNS(t *testing.T) {
	matcher := NewPatternMatcher(DefaultPatterns())

	tests := []struct {
		name    string
		errMsg  string
		matches bool
	}{
		{"no such host", "no such host: api.example.com", true},
		{"DNS resolution failed", "DNS resolution failed for example.com", true},
		{"ENOTFOUND", "getaddrinfo ENOTFOUND api.test.com", true},
		{"unrelated", "connection refused", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errCtx := &ErrorContext{
				ExecutionID:  "exec-123",
				WorkflowID:   "wf-456",
				NodeID:       "node-789",
				NodeType:     "action:http",
				ErrorMessage: tt.errMsg,
			}

			matches := matcher.Match(errCtx)
			if tt.matches {
				require.NotEmpty(t, matches, "should match DNS pattern")
				assert.Equal(t, ErrorCategoryNetwork, matches[0].Category)
			}
		})
	}
}

func TestPatternMatcher_ToSuggestion(t *testing.T) {
	matcher := NewPatternMatcher(DefaultPatterns())

	errCtx := &ErrorContext{
		ExecutionID:  "exec-123",
		WorkflowID:   "wf-456",
		NodeID:       "node-789",
		NodeType:     "action:http",
		ErrorMessage: "connection refused",
	}

	matches := matcher.Match(errCtx)
	require.NotEmpty(t, matches)

	// Convert to suggestion
	suggestion := matches[0].ToSuggestion("tenant-123", errCtx)

	assert.NotEmpty(t, suggestion.ID)
	assert.Equal(t, "tenant-123", suggestion.TenantID)
	assert.Equal(t, "exec-123", suggestion.ExecutionID)
	assert.Equal(t, "node-789", suggestion.NodeID)
	assert.Equal(t, ErrorCategoryNetwork, suggestion.Category)
	assert.Equal(t, SourcePattern, suggestion.Source)
	assert.Equal(t, StatusPending, suggestion.Status)
	assert.NotNil(t, suggestion.Fix, "should have a fix recommendation")
}

func TestPatternMatcher_MultipleMatches(t *testing.T) {
	matcher := NewPatternMatcher(DefaultPatterns())

	// Error with both message pattern and HTTP code
	errCtx := &ErrorContext{
		ExecutionID:  "exec-123",
		WorkflowID:   "wf-456",
		NodeID:       "node-789",
		NodeType:     "action:http",
		ErrorMessage: "rate limit exceeded",
		HTTPStatus:   429,
	}

	matches := matcher.Match(errCtx)
	// Should match at least once (might dedupe by category)
	require.NotEmpty(t, matches)

	// All matches should be rate limit category
	for _, m := range matches {
		assert.Equal(t, ErrorCategoryRateLimit, m.Category)
	}
}

func TestPatternMatcher_Priority(t *testing.T) {
	// Create patterns with different priorities
	patterns := []*BuiltinPattern{
		{
			Name:                  "low-priority",
			Category:              ErrorCategoryNetwork,
			MessagePatterns:       []string{"error"},
			SuggestionType:        SuggestionTypeManual,
			SuggestionTitle:       "Generic Error",
			SuggestionDescription: "A generic error occurred",
			SuggestionConfidence:  ConfidenceLow,
			Priority:              10,
		},
		{
			Name:                  "high-priority",
			Category:              ErrorCategoryNetwork,
			MessagePatterns:       []string{"connection refused"},
			SuggestionType:        SuggestionTypeRetry,
			SuggestionTitle:       "Connection Refused",
			SuggestionDescription: "Specific connection error",
			SuggestionConfidence:  ConfidenceHigh,
			Priority:              100,
		},
	}

	matcher := NewPatternMatcher(patterns)

	errCtx := &ErrorContext{
		ExecutionID:  "exec-123",
		WorkflowID:   "wf-456",
		NodeID:       "node-789",
		NodeType:     "action:http",
		ErrorMessage: "connection refused error",
	}

	matches := matcher.Match(errCtx)
	require.NotEmpty(t, matches)

	// High priority should come first
	assert.Equal(t, "high-priority", matches[0].PatternName)
}

func TestPatternMatcher_NodeTypeFilter(t *testing.T) {
	// Pattern that only applies to HTTP actions
	patterns := []*BuiltinPattern{
		{
			Name:                  "http-only",
			Category:              ErrorCategoryNetwork,
			MessagePatterns:       []string{"connection error"},
			NodeTypes:             []string{"action:http", "action:webhook"},
			SuggestionType:        SuggestionTypeRetry,
			SuggestionTitle:       "HTTP Connection Error",
			SuggestionDescription: "HTTP-specific error",
			SuggestionConfidence:  ConfidenceHigh,
			Priority:              100,
		},
	}

	matcher := NewPatternMatcher(patterns)

	// Should match HTTP action
	httpCtx := &ErrorContext{
		ExecutionID:  "exec-123",
		WorkflowID:   "wf-456",
		NodeID:       "node-789",
		NodeType:     "action:http",
		ErrorMessage: "connection error",
	}

	matches := matcher.Match(httpCtx)
	require.NotEmpty(t, matches, "should match HTTP action")

	// Should NOT match transform action
	transformCtx := &ErrorContext{
		ExecutionID:  "exec-123",
		WorkflowID:   "wf-456",
		NodeID:       "node-789",
		NodeType:     "action:transform",
		ErrorMessage: "connection error",
	}

	matches = matcher.Match(transformCtx)
	assert.Empty(t, matches, "should not match transform action")
}

func TestPatternMatcher_CaseInsensitive(t *testing.T) {
	matcher := NewPatternMatcher(DefaultPatterns())

	tests := []struct {
		name   string
		errMsg string
	}{
		{"lowercase", "connection refused"},
		{"uppercase", "CONNECTION REFUSED"},
		{"mixed case", "Connection Refused"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errCtx := &ErrorContext{
				ExecutionID:  "exec-123",
				WorkflowID:   "wf-456",
				NodeID:       "node-789",
				NodeType:     "action:http",
				ErrorMessage: tt.errMsg,
			}

			matches := matcher.Match(errCtx)
			require.NotEmpty(t, matches, "should match case-insensitively")
		})
	}
}

func TestPatternMatch_FixTemplates(t *testing.T) {
	matcher := NewPatternMatcher(DefaultPatterns())

	tests := []struct {
		name          string
		errMsg        string
		httpCode      int
		expectFixType string
		expectRetry   bool
		expectConfig  bool
	}{
		{
			name:          "connection refused should have retry fix",
			errMsg:        "connection refused",
			expectFixType: "retry_with_backoff",
			expectRetry:   true,
		},
		{
			name:          "401 should have credential fix",
			httpCode:      401,
			expectFixType: "credential_update",
		},
		{
			name:          "429 should have config fix",
			errMsg:        "rate limit",
			httpCode:      429,
			expectFixType: "config_change",
			expectConfig:  true,
		},
		{
			name:          "timeout should have config fix",
			errMsg:        "context deadline exceeded",
			expectFixType: "config_change",
			expectConfig:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errCtx := &ErrorContext{
				ExecutionID:  "exec-123",
				WorkflowID:   "wf-456",
				NodeID:       "node-789",
				NodeType:     "action:http",
				ErrorMessage: tt.errMsg,
				HTTPStatus:   tt.httpCode,
			}

			matches := matcher.Match(errCtx)
			require.NotEmpty(t, matches)

			suggestion := matches[0].ToSuggestion("tenant-123", errCtx)
			require.NotNil(t, suggestion.Fix)
			assert.Equal(t, tt.expectFixType, suggestion.Fix.ActionType)

			if tt.expectRetry {
				assert.NotNil(t, suggestion.Fix.RetryConfig)
			}
			if tt.expectConfig {
				assert.NotEmpty(t, suggestion.Fix.ConfigPath)
			}
		})
	}
}
