package suggestions

import (
	"regexp"
	"sort"
	"strings"
)

// BuiltinPattern defines a pattern for matching errors
type BuiltinPattern struct {
	Name                  string
	Category              ErrorCategory
	MessagePatterns       []string // Regex patterns to match against error message
	HTTPCodes             []int    // HTTP status codes to match
	NodeTypes             []string // Optional: only match these node types (empty = all)
	SuggestionType        SuggestionType
	SuggestionTitle       string
	SuggestionDescription string
	SuggestionConfidence  SuggestionConfidence
	FixTemplate           *SuggestionFix
	Priority              int // Higher priority patterns are checked first
}

// PatternMatch represents a matched pattern with context
type PatternMatch struct {
	PatternName     string
	Category        ErrorCategory
	Type            SuggestionType
	Confidence      SuggestionConfidence
	Title           string
	Description     string
	MatchedPattern  string // The regex pattern that matched
	MatchedHTTPCode int    // The HTTP code that matched (0 if none)
	FixTemplate     *SuggestionFix
	Priority        int
}

// ToSuggestion converts a pattern match to a suggestion
func (m *PatternMatch) ToSuggestion(tenantID string, errCtx *ErrorContext) *Suggestion {
	suggestion := NewSuggestion(
		tenantID,
		errCtx.ExecutionID,
		errCtx.NodeID,
		m.Category,
		m.Type,
		m.Confidence,
		m.Title,
		m.Description,
		SourcePattern,
	)

	// Apply fix template if present
	if m.FixTemplate != nil {
		suggestion.Fix = m.FixTemplate
	}

	return suggestion
}

// PatternMatcher matches errors against patterns
type PatternMatcher struct {
	patterns        []*BuiltinPattern
	compiledRegexes map[string][]*regexp.Regexp
}

// NewPatternMatcher creates a new pattern matcher
func NewPatternMatcher(patterns []*BuiltinPattern) *PatternMatcher {
	pm := &PatternMatcher{
		patterns:        patterns,
		compiledRegexes: make(map[string][]*regexp.Regexp),
	}

	// Pre-compile all regex patterns (case-insensitive)
	for _, p := range patterns {
		var compiled []*regexp.Regexp
		for _, pattern := range p.MessagePatterns {
			// Make pattern case-insensitive
			re, err := regexp.Compile("(?i)" + pattern)
			if err == nil {
				compiled = append(compiled, re)
			}
		}
		pm.compiledRegexes[p.Name] = compiled
	}

	// Sort patterns by priority (highest first)
	sort.Slice(pm.patterns, func(i, j int) bool {
		return pm.patterns[i].Priority > pm.patterns[j].Priority
	})

	return pm
}

// Match finds all patterns that match the given error context
func (pm *PatternMatcher) Match(errCtx *ErrorContext) []*PatternMatch {
	var matches []*PatternMatch
	seenCategories := make(map[ErrorCategory]bool)

	for _, pattern := range pm.patterns {
		match := pm.matchPattern(pattern, errCtx)
		if match != nil {
			// Deduplicate by category (keep highest priority)
			if !seenCategories[match.Category] {
				matches = append(matches, match)
				seenCategories[match.Category] = true
			}
		}
	}

	return matches
}

func (pm *PatternMatcher) matchPattern(pattern *BuiltinPattern, errCtx *ErrorContext) *PatternMatch {
	// Check node type filter if specified
	if len(pattern.NodeTypes) > 0 {
		matched := false
		for _, nodeType := range pattern.NodeTypes {
			if nodeType == errCtx.NodeType {
				matched = true
				break
			}
		}
		if !matched {
			return nil
		}
	}

	var matchedPattern string
	var matchedHTTPCode int

	// Check HTTP codes
	if errCtx.HTTPStatus > 0 && len(pattern.HTTPCodes) > 0 {
		for _, code := range pattern.HTTPCodes {
			if code == errCtx.HTTPStatus {
				matchedHTTPCode = code
				break
			}
		}
	}

	// Check message patterns
	if errCtx.ErrorMessage != "" {
		regexes := pm.compiledRegexes[pattern.Name]
		for i, re := range regexes {
			if re.MatchString(errCtx.ErrorMessage) {
				matchedPattern = pattern.MessagePatterns[i]
				break
			}
		}
	}

	// Must match at least one criterion
	if matchedPattern == "" && matchedHTTPCode == 0 {
		return nil
	}

	return &PatternMatch{
		PatternName:     pattern.Name,
		Category:        pattern.Category,
		Type:            pattern.SuggestionType,
		Confidence:      pattern.SuggestionConfidence,
		Title:           pattern.SuggestionTitle,
		Description:     pattern.SuggestionDescription,
		MatchedPattern:  matchedPattern,
		MatchedHTTPCode: matchedHTTPCode,
		FixTemplate:     pattern.FixTemplate,
		Priority:        pattern.Priority,
	}
}

// DefaultPatterns returns the built-in error patterns
func DefaultPatterns() []*BuiltinPattern {
	return []*BuiltinPattern{
		// Network errors
		connectionRefusedPattern(),
		dnsResolutionPattern(),

		// Authentication errors
		auth401Pattern(),
		auth403Pattern(),

		// Rate limiting
		rateLimitPattern(),

		// Timeout errors
		timeoutPattern(),

		// Data/parsing errors
		jsonParsePattern(),
		validationErrorPattern(),

		// Server errors
		serverError500Pattern(),
		serverError502503Pattern(),
	}
}

func connectionRefusedPattern() *BuiltinPattern {
	return &BuiltinPattern{
		Name:     "connection_refused",
		Category: ErrorCategoryNetwork,
		MessagePatterns: []string{
			"connection refused",
			"ECONNREFUSED",
			`dial tcp.*connection refused`,
		},
		SuggestionType:        SuggestionTypeRetry,
		SuggestionTitle:       "Connection Refused",
		SuggestionDescription: "The target service is not accepting connections. This may be a temporary issue.",
		SuggestionConfidence:  ConfidenceHigh,
		FixTemplate: &SuggestionFix{
			ActionType: "retry_with_backoff",
			RetryConfig: &RetryConfig{
				MaxRetries:    5,
				BackoffMs:     2000,
				BackoffFactor: 2.0,
			},
		},
		Priority: 100,
	}
}

func dnsResolutionPattern() *BuiltinPattern {
	return &BuiltinPattern{
		Name:     "dns_resolution",
		Category: ErrorCategoryNetwork,
		MessagePatterns: []string{
			"no such host",
			"DNS resolution failed",
			"getaddrinfo ENOTFOUND",
		},
		SuggestionType:        SuggestionTypeConfigChange,
		SuggestionTitle:       "DNS Resolution Failed",
		SuggestionDescription: "The hostname could not be resolved. Check if the URL is correct.",
		SuggestionConfidence:  ConfidenceHigh,
		FixTemplate: &SuggestionFix{
			ActionType: "config_change",
			ConfigPath: "url",
		},
		Priority: 100,
	}
}

func auth401Pattern() *BuiltinPattern {
	return &BuiltinPattern{
		Name:                  "auth_401",
		Category:              ErrorCategoryAuth,
		HTTPCodes:             []int{401},
		SuggestionType:        SuggestionTypeCredential,
		SuggestionTitle:       "Authentication Failed",
		SuggestionDescription: "The credentials used for this action are invalid or expired. Please update your credentials.",
		SuggestionConfidence:  ConfidenceHigh,
		FixTemplate: &SuggestionFix{
			ActionType: "credential_update",
		},
		Priority: 100,
	}
}

func auth403Pattern() *BuiltinPattern {
	return &BuiltinPattern{
		Name:                  "auth_403",
		Category:              ErrorCategoryAuth,
		HTTPCodes:             []int{403},
		SuggestionType:        SuggestionTypeCredential,
		SuggestionTitle:       "Access Forbidden",
		SuggestionDescription: "The credentials do not have permission for this operation. Check the credential permissions.",
		SuggestionConfidence:  ConfidenceHigh,
		FixTemplate: &SuggestionFix{
			ActionType: "credential_update",
		},
		Priority: 100,
	}
}

func rateLimitPattern() *BuiltinPattern {
	return &BuiltinPattern{
		Name:     "rate_limit",
		Category: ErrorCategoryRateLimit,
		MessagePatterns: []string{
			"rate limit",
			"too many requests",
			"throttle",
			`exceeded.*limit`,
		},
		HTTPCodes:             []int{429},
		SuggestionType:        SuggestionTypeConfigChange,
		SuggestionTitle:       "Rate Limit Exceeded",
		SuggestionDescription: "The API rate limit was exceeded. Consider adding delays between requests or reducing request frequency.",
		SuggestionConfidence:  ConfidenceHigh,
		FixTemplate: &SuggestionFix{
			ActionType: "config_change",
			ConfigPath: "rate_limit",
			NewValue: map[string]interface{}{
				"delay_ms":       1000,
				"max_concurrent": 1,
			},
		},
		Priority: 100,
	}
}

func timeoutPattern() *BuiltinPattern {
	return &BuiltinPattern{
		Name:     "timeout",
		Category: ErrorCategoryTimeout,
		MessagePatterns: []string{
			"timeout",
			"timed out",
			"deadline exceeded",
			"context deadline exceeded",
		},
		HTTPCodes:             []int{504, 408},
		SuggestionType:        SuggestionTypeConfigChange,
		SuggestionTitle:       "Request Timeout",
		SuggestionDescription: "The request took too long to complete. Consider increasing the timeout value.",
		SuggestionConfidence:  ConfidenceHigh,
		FixTemplate: &SuggestionFix{
			ActionType: "config_change",
			ConfigPath: "timeout",
			NewValue:   60,
		},
		Priority: 100,
	}
}

func jsonParsePattern() *BuiltinPattern {
	return &BuiltinPattern{
		Name:     "json_parse",
		Category: ErrorCategoryData,
		MessagePatterns: []string{
			"invalid json",
			`json.*parse error`,
			"unexpected token",
			`syntax error.*json`,
			"invalid character",
		},
		SuggestionType:        SuggestionTypeDataFix,
		SuggestionTitle:       "Invalid JSON Data",
		SuggestionDescription: "The data format is invalid JSON. Check the input data structure and ensure it is valid JSON.",
		SuggestionConfidence:  ConfidenceMedium,
		FixTemplate: &SuggestionFix{
			ActionType: "data_fix",
		},
		Priority: 90,
	}
}

func validationErrorPattern() *BuiltinPattern {
	return &BuiltinPattern{
		Name:     "validation_error",
		Category: ErrorCategoryData,
		MessagePatterns: []string{
			`validation.*failed`,
			"required field",
			`invalid.*format`,
			`must be.*type`,
		},
		HTTPCodes:             []int{400, 422},
		SuggestionType:        SuggestionTypeDataFix,
		SuggestionTitle:       "Data Validation Failed",
		SuggestionDescription: "The input data does not meet validation requirements. Check the data format and required fields.",
		SuggestionConfidence:  ConfidenceMedium,
		FixTemplate: &SuggestionFix{
			ActionType: "data_fix",
		},
		Priority: 90,
	}
}

func serverError500Pattern() *BuiltinPattern {
	return &BuiltinPattern{
		Name:                  "server_error_500",
		Category:              ErrorCategoryExternal,
		HTTPCodes:             []int{500},
		SuggestionType:        SuggestionTypeRetry,
		SuggestionTitle:       "Internal Server Error",
		SuggestionDescription: "The external service returned an internal error. This is usually a temporary issue.",
		SuggestionConfidence:  ConfidenceMedium,
		FixTemplate: &SuggestionFix{
			ActionType: "retry_with_backoff",
			RetryConfig: &RetryConfig{
				MaxRetries:    3,
				BackoffMs:     5000,
				BackoffFactor: 2.0,
			},
		},
		Priority: 80,
	}
}

func serverError502503Pattern() *BuiltinPattern {
	return &BuiltinPattern{
		Name:                  "server_error_502_503",
		Category:              ErrorCategoryExternal,
		HTTPCodes:             []int{502, 503},
		SuggestionType:        SuggestionTypeRetry,
		SuggestionTitle:       "Service Unavailable",
		SuggestionDescription: "The external service is temporarily unavailable. This is usually a temporary issue.",
		SuggestionConfidence:  ConfidenceHigh,
		FixTemplate: &SuggestionFix{
			ActionType: "retry_with_backoff",
			RetryConfig: &RetryConfig{
				MaxRetries:    5,
				BackoffMs:     3000,
				BackoffFactor: 2.0,
			},
		},
		Priority: 80,
	}
}

// GetPatternByName returns a pattern by name
func GetPatternByName(patterns []*BuiltinPattern, name string) *BuiltinPattern {
	for _, p := range patterns {
		if p.Name == name {
			return p
		}
	}
	return nil
}

// FilterPatternsByCategory returns patterns matching the given category
func FilterPatternsByCategory(patterns []*BuiltinPattern, category ErrorCategory) []*BuiltinPattern {
	var result []*BuiltinPattern
	for _, p := range patterns {
		if p.Category == category {
			result = append(result, p)
		}
	}
	return result
}

// CategoryFromHTTPCode determines the error category from an HTTP status code
func CategoryFromHTTPCode(code int) ErrorCategory {
	switch {
	case code == 401 || code == 403:
		return ErrorCategoryAuth
	case code == 429:
		return ErrorCategoryRateLimit
	case code == 408 || code == 504:
		return ErrorCategoryTimeout
	case code >= 500:
		return ErrorCategoryExternal
	case code >= 400:
		return ErrorCategoryData
	default:
		return ErrorCategoryUnknown
	}
}

// CategoryFromMessage attempts to determine category from error message
func CategoryFromMessage(msg string) ErrorCategory {
	lowerMsg := strings.ToLower(msg)

	switch {
	case containsAny(lowerMsg, "connection refused", "econnrefused", "no such host", "dns"):
		return ErrorCategoryNetwork
	case containsAny(lowerMsg, "unauthorized", "authentication", "invalid credentials", "token expired"):
		return ErrorCategoryAuth
	case containsAny(lowerMsg, "rate limit", "too many requests", "throttle"):
		return ErrorCategoryRateLimit
	case containsAny(lowerMsg, "timeout", "timed out", "deadline exceeded"):
		return ErrorCategoryTimeout
	case containsAny(lowerMsg, "invalid json", "parse error", "validation", "required field"):
		return ErrorCategoryData
	case containsAny(lowerMsg, "service unavailable", "bad gateway", "internal server error"):
		return ErrorCategoryExternal
	default:
		return ErrorCategoryUnknown
	}
}

func containsAny(s string, substrs ...string) bool {
	for _, substr := range substrs {
		if strings.Contains(s, substr) {
			return true
		}
	}
	return false
}
