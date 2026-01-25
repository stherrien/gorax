package security

import (
	"encoding/json"
	"html"
	"regexp"
	"strings"
	"unicode"
)

// OutputSanitizer provides methods for sanitizing output data
type OutputSanitizer struct{}

// NewOutputSanitizer creates a new output sanitizer
func NewOutputSanitizer() *OutputSanitizer {
	return &OutputSanitizer{}
}

// SanitizeForHTML escapes HTML special characters
func (s *OutputSanitizer) SanitizeForHTML(input string) string {
	return html.EscapeString(input)
}

// SanitizeForJSON ensures string is safe for JSON output
func (s *OutputSanitizer) SanitizeForJSON(input string) string {
	// Remove null bytes
	input = strings.ReplaceAll(input, "\x00", "")

	// Remove other control characters except newlines and tabs
	var result strings.Builder
	for _, r := range input {
		if r == '\n' || r == '\r' || r == '\t' || !unicode.IsControl(r) {
			result.WriteRune(r)
		}
	}

	return result.String()
}

// SanitizeStruct sanitizes all string fields in a struct for JSON output
func (s *OutputSanitizer) SanitizeStruct(v any) (any, error) {
	// Marshal to JSON
	data, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}

	// Unmarshal to map for manipulation
	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		// Try as array
		var arr []any
		if err := json.Unmarshal(data, &arr); err != nil {
			return v, nil // Return original if not map or array
		}
		return s.sanitizeSlice(arr), nil
	}

	return s.sanitizeMap(m), nil
}

func (s *OutputSanitizer) sanitizeMap(m map[string]any) map[string]any {
	result := make(map[string]any)
	for key, value := range m {
		result[key] = s.sanitizeValue(value)
	}
	return result
}

func (s *OutputSanitizer) sanitizeSlice(arr []any) []any {
	result := make([]any, len(arr))
	for i, value := range arr {
		result[i] = s.sanitizeValue(value)
	}
	return result
}

func (s *OutputSanitizer) sanitizeValue(v any) any {
	switch val := v.(type) {
	case string:
		return s.SanitizeForJSON(val)
	case map[string]any:
		return s.sanitizeMap(val)
	case []any:
		return s.sanitizeSlice(val)
	default:
		return v
	}
}

// --- Search Query Sanitization ---

// SearchQuerySanitizer provides methods for sanitizing search queries
type SearchQuerySanitizer struct {
	maxLength int
}

// NewSearchQuerySanitizer creates a new search query sanitizer
func NewSearchQuerySanitizer(maxLength int) *SearchQuerySanitizer {
	if maxLength <= 0 {
		maxLength = 500
	}
	return &SearchQuerySanitizer{maxLength: maxLength}
}

// Sanitize sanitizes a search query string
func (s *SearchQuerySanitizer) Sanitize(query string) string {
	// Remove null bytes
	query = strings.ReplaceAll(query, "\x00", "")

	// Trim whitespace
	query = strings.TrimSpace(query)

	// Limit length
	if len(query) > s.maxLength {
		query = query[:s.maxLength]
	}

	return query
}

// SanitizeForLike sanitizes a search query for use in SQL LIKE patterns
// This escapes SQL LIKE wildcards while allowing the search to work
func (s *SearchQuerySanitizer) SanitizeForLike(query string) string {
	query = s.Sanitize(query)

	// Escape SQL LIKE special characters
	query = strings.ReplaceAll(query, "\\", "\\\\")
	query = strings.ReplaceAll(query, "%", "\\%")
	query = strings.ReplaceAll(query, "_", "\\_")

	return query
}

// --- Identifier Sanitization ---

// identifierRegex matches valid SQL identifiers
var identifierRegex = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)

// ValidateIdentifier checks if a string is a valid SQL identifier
func ValidateIdentifier(identifier string) bool {
	// Check length
	if len(identifier) == 0 || len(identifier) > 63 {
		return false
	}

	// Check format
	return identifierRegex.MatchString(identifier)
}

// SanitizeIdentifier sanitizes a SQL identifier (column/table name)
// Returns empty string if invalid
func SanitizeIdentifier(identifier string) string {
	if ValidateIdentifier(identifier) {
		return identifier
	}
	return ""
}

// AllowedSortColumns validates a sort column against an allowlist
func AllowedSortColumns(column string, allowlist []string) bool {
	for _, allowed := range allowlist {
		if strings.EqualFold(column, allowed) {
			return true
		}
	}
	return false
}

// AllowedSortDirections validates a sort direction
func AllowedSortDirections(direction string) bool {
	upper := strings.ToUpper(direction)
	return upper == "ASC" || upper == "DESC"
}

// --- Log Sanitization ---

// LogSanitizer provides methods for sanitizing data before logging
type LogSanitizer struct {
	sensitiveFields []string
	redactValue     string
}

// NewLogSanitizer creates a new log sanitizer
func NewLogSanitizer() *LogSanitizer {
	return &LogSanitizer{
		sensitiveFields: []string{
			"password", "secret", "token", "api_key", "apikey",
			"credential", "private_key", "privatekey", "auth",
			"authorization", "bearer", "cookie", "session",
			"access_token", "refresh_token", "id_token",
			"client_secret", "client_id",
		},
		redactValue: "[REDACTED]",
	}
}

// SanitizeForLog sanitizes a map for logging (redacts sensitive fields)
func (s *LogSanitizer) SanitizeForLog(data map[string]any) map[string]any {
	result := make(map[string]any)
	for key, value := range data {
		if s.isSensitiveField(key) {
			result[key] = s.redactValue
		} else if m, ok := value.(map[string]any); ok {
			result[key] = s.SanitizeForLog(m)
		} else {
			result[key] = value
		}
	}
	return result
}

func (s *LogSanitizer) isSensitiveField(field string) bool {
	lower := strings.ToLower(field)
	for _, sensitive := range s.sensitiveFields {
		if strings.Contains(lower, sensitive) {
			return true
		}
	}
	return false
}

// MaskString masks a string, showing only first and last characters
func MaskString(s string, visibleChars int) string {
	if len(s) <= visibleChars*2 {
		return strings.Repeat("*", len(s))
	}
	return s[:visibleChars] + strings.Repeat("*", len(s)-visibleChars*2) + s[len(s)-visibleChars:]
}

// MaskEmail masks an email address
func MaskEmail(email string) string {
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return strings.Repeat("*", len(email))
	}

	local := parts[0]
	domain := parts[1]

	maskedLocal := MaskString(local, 2)
	return maskedLocal + "@" + domain
}

// --- JSON Value Sanitization ---

// SanitizeJSONValue recursively sanitizes JSON values to remove potential XSS
func SanitizeJSONValue(v any) any {
	switch val := v.(type) {
	case string:
		// Remove potential XSS vectors
		if ContainsXSSPattern(val) {
			return SanitizeHTML(val)
		}
		return val
	case map[string]any:
		result := make(map[string]any)
		for k, value := range val {
			result[k] = SanitizeJSONValue(value)
		}
		return result
	case []any:
		result := make([]any, len(val))
		for i, value := range val {
			result[i] = SanitizeJSONValue(value)
		}
		return result
	default:
		return v
	}
}
