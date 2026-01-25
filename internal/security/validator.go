package security

import (
	"encoding/json"
	"fmt"
	"html"
	"net/mail"
	"path/filepath"
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"
)

// ValidationError represents a validation failure with details
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Code    string `json:"code"`
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

// ValidationErrors holds multiple validation errors
type ValidationErrors struct {
	Errors []ValidationError `json:"errors"`
}

func (e *ValidationErrors) Error() string {
	if len(e.Errors) == 0 {
		return "validation failed"
	}
	if len(e.Errors) == 1 {
		return e.Errors[0].Error()
	}
	var msgs []string
	for _, err := range e.Errors {
		msgs = append(msgs, err.Error())
	}
	return strings.Join(msgs, "; ")
}

// Add adds a validation error
func (e *ValidationErrors) Add(field, message, code string) {
	e.Errors = append(e.Errors, ValidationError{
		Field:   field,
		Message: message,
		Code:    code,
	})
}

// HasErrors returns true if there are validation errors
func (e *ValidationErrors) HasErrors() bool {
	return len(e.Errors) > 0
}

// InputValidator provides methods for validating and sanitizing user input
type InputValidator struct {
	config *ValidatorConfig
}

// ValidatorConfig holds validation configuration
type ValidatorConfig struct {
	// MaxStringLength is the maximum allowed string length (default: 10000)
	MaxStringLength int
	// MaxNameLength is the maximum allowed name/title length (default: 255)
	MaxNameLength int
	// MaxDescriptionLength is the maximum allowed description length (default: 5000)
	MaxDescriptionLength int
	// MinPasswordLength is the minimum required password length (default: 12)
	MinPasswordLength int
	// MaxPasswordLength is the maximum allowed password length (default: 128)
	MaxPasswordLength int
	// AllowedFileTypes are the allowed MIME types for file uploads
	AllowedFileTypes []string
	// MaxFileSize is the maximum allowed file size in bytes (default: 10MB)
	MaxFileSize int64
}

// DefaultValidatorConfig returns a secure default configuration
func DefaultValidatorConfig() *ValidatorConfig {
	return &ValidatorConfig{
		MaxStringLength:      10000,
		MaxNameLength:        255,
		MaxDescriptionLength: 5000,
		MinPasswordLength:    12,
		MaxPasswordLength:    128,
		AllowedFileTypes: []string{
			"application/json",
			"text/plain",
			"text/yaml",
			"application/x-yaml",
		},
		MaxFileSize: 10 * 1024 * 1024, // 10MB
	}
}

// NewInputValidator creates a new input validator
func NewInputValidator() *InputValidator {
	return NewInputValidatorWithConfig(DefaultValidatorConfig())
}

// NewInputValidatorWithConfig creates a new input validator with custom config
func NewInputValidatorWithConfig(config *ValidatorConfig) *InputValidator {
	return &InputValidator{config: config}
}

// --- Email Validation ---

// emailRegex is a simplified RFC 5322 compliant email pattern
// Using a simpler pattern to avoid ReDoS while maintaining good validation
var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

// ValidateEmail validates an email address
func (v *InputValidator) ValidateEmail(email string) error {
	if email == "" {
		return &ValidationError{
			Field:   "email",
			Message: "email is required",
			Code:    "required",
		}
	}

	// Check length first to prevent ReDoS
	if len(email) > 254 {
		return &ValidationError{
			Field:   "email",
			Message: "email exceeds maximum length of 254 characters",
			Code:    "max_length",
		}
	}

	// Use stdlib for parsing to validate structure
	_, err := mail.ParseAddress(email)
	if err != nil {
		return &ValidationError{
			Field:   "email",
			Message: "invalid email format",
			Code:    "invalid_format",
		}
	}

	// Additional regex check for common attack patterns
	if !emailRegex.MatchString(email) {
		return &ValidationError{
			Field:   "email",
			Message: "email contains invalid characters",
			Code:    "invalid_chars",
		}
	}

	return nil
}

// --- Password Validation ---

// PasswordStrength represents password strength levels
type PasswordStrength int

const (
	PasswordStrengthWeak PasswordStrength = iota
	PasswordStrengthFair
	PasswordStrengthGood
	PasswordStrengthStrong
)

// ValidatePassword validates password strength
func (v *InputValidator) ValidatePassword(password string) error {
	if password == "" {
		return &ValidationError{
			Field:   "password",
			Message: "password is required",
			Code:    "required",
		}
	}

	if len(password) < v.config.MinPasswordLength {
		return &ValidationError{
			Field:   "password",
			Message: fmt.Sprintf("password must be at least %d characters", v.config.MinPasswordLength),
			Code:    "min_length",
		}
	}

	if len(password) > v.config.MaxPasswordLength {
		return &ValidationError{
			Field:   "password",
			Message: fmt.Sprintf("password must not exceed %d characters", v.config.MaxPasswordLength),
			Code:    "max_length",
		}
	}

	// Check for character variety
	var hasUpper, hasLower, hasDigit, hasSpecial bool
	for _, r := range password {
		switch {
		case unicode.IsUpper(r):
			hasUpper = true
		case unicode.IsLower(r):
			hasLower = true
		case unicode.IsDigit(r):
			hasDigit = true
		case unicode.IsPunct(r) || unicode.IsSymbol(r):
			hasSpecial = true
		}
	}

	// Require at least 3 of 4 character types
	count := 0
	if hasUpper {
		count++
	}
	if hasLower {
		count++
	}
	if hasDigit {
		count++
	}
	if hasSpecial {
		count++
	}

	if count < 3 {
		return &ValidationError{
			Field:   "password",
			Message: "password must contain at least 3 of: uppercase, lowercase, digit, special character",
			Code:    "weak_password",
		}
	}

	return nil
}

// --- Name/String Validation ---

// nameRegex allows alphanumeric, spaces, hyphens, underscores, and common punctuation
var nameRegex = regexp.MustCompile(`^[\p{L}\p{N}\s\-_.,!?'":;()&]+$`)

// ValidateName validates a name field (workflow name, credential name, etc.)
func (v *InputValidator) ValidateName(name, fieldName string) error {
	if name == "" {
		return &ValidationError{
			Field:   fieldName,
			Message: fieldName + " is required",
			Code:    "required",
		}
	}

	if len(name) > v.config.MaxNameLength {
		return &ValidationError{
			Field:   fieldName,
			Message: fmt.Sprintf("%s must not exceed %d characters", fieldName, v.config.MaxNameLength),
			Code:    "max_length",
		}
	}

	// Check for valid UTF-8
	if !utf8.ValidString(name) {
		return &ValidationError{
			Field:   fieldName,
			Message: fieldName + " contains invalid UTF-8 characters",
			Code:    "invalid_encoding",
		}
	}

	// Check for null bytes (injection vector)
	if strings.ContainsRune(name, '\x00') {
		return &ValidationError{
			Field:   fieldName,
			Message: fieldName + " contains invalid characters",
			Code:    "invalid_chars",
		}
	}

	// Validate allowed characters
	if !nameRegex.MatchString(name) {
		return &ValidationError{
			Field:   fieldName,
			Message: fieldName + " contains invalid characters",
			Code:    "invalid_chars",
		}
	}

	return nil
}

// ValidateDescription validates a description field
func (v *InputValidator) ValidateDescription(desc, fieldName string) error {
	// Descriptions can be empty
	if desc == "" {
		return nil
	}

	if len(desc) > v.config.MaxDescriptionLength {
		return &ValidationError{
			Field:   fieldName,
			Message: fmt.Sprintf("%s must not exceed %d characters", fieldName, v.config.MaxDescriptionLength),
			Code:    "max_length",
		}
	}

	// Check for valid UTF-8
	if !utf8.ValidString(desc) {
		return &ValidationError{
			Field:   fieldName,
			Message: fieldName + " contains invalid UTF-8 characters",
			Code:    "invalid_encoding",
		}
	}

	// Check for null bytes
	if strings.ContainsRune(desc, '\x00') {
		return &ValidationError{
			Field:   fieldName,
			Message: fieldName + " contains invalid characters",
			Code:    "invalid_chars",
		}
	}

	return nil
}

// --- ID Validation ---

// uuidRegex validates UUID format
var uuidRegex = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

// ValidateUUID validates a UUID string
func (v *InputValidator) ValidateUUID(id, fieldName string) error {
	if id == "" {
		return &ValidationError{
			Field:   fieldName,
			Message: fieldName + " is required",
			Code:    "required",
		}
	}

	if !uuidRegex.MatchString(id) {
		return &ValidationError{
			Field:   fieldName,
			Message: fieldName + " must be a valid UUID",
			Code:    "invalid_format",
		}
	}

	return nil
}

// --- JSON Validation ---

// ValidateJSONSize checks if JSON payload is within limits
func (v *InputValidator) ValidateJSONSize(data []byte) error {
	if len(data) > v.config.MaxStringLength*10 { // 100KB default
		return &ValidationError{
			Field:   "body",
			Message: "request body exceeds maximum size",
			Code:    "max_size",
		}
	}
	return nil
}

// ValidateJSON validates JSON structure
func (v *InputValidator) ValidateJSON(data []byte) error {
	if err := v.ValidateJSONSize(data); err != nil {
		return err
	}

	if !json.Valid(data) {
		return &ValidationError{
			Field:   "body",
			Message: "invalid JSON format",
			Code:    "invalid_json",
		}
	}

	return nil
}

// --- Sanitization Functions ---

// SanitizeString removes or escapes potentially dangerous characters
func SanitizeString(s string) string {
	// Remove null bytes
	s = strings.ReplaceAll(s, "\x00", "")

	// Trim whitespace
	s = strings.TrimSpace(s)

	return s
}

// SanitizeHTML escapes HTML entities to prevent XSS
func SanitizeHTML(s string) string {
	return html.EscapeString(s)
}

// SanitizePath prevents path traversal attacks
func SanitizePath(path string) (string, error) {
	// Remove null bytes
	path = strings.ReplaceAll(path, "\x00", "")

	// Clean the path
	cleaned := filepath.Clean(path)

	// Check for path traversal
	if strings.Contains(cleaned, "..") {
		return "", &ValidationError{
			Field:   "path",
			Message: "path traversal not allowed",
			Code:    "path_traversal",
		}
	}

	// Ensure path doesn't start with /
	if strings.HasPrefix(cleaned, "/") || strings.HasPrefix(cleaned, "\\") {
		return "", &ValidationError{
			Field:   "path",
			Message: "absolute paths not allowed",
			Code:    "absolute_path",
		}
	}

	return cleaned, nil
}

// SanitizeSearchQuery sanitizes search input
func SanitizeSearchQuery(query string) string {
	// Remove null bytes
	query = strings.ReplaceAll(query, "\x00", "")

	// Trim and limit length
	query = strings.TrimSpace(query)
	if len(query) > 500 {
		query = query[:500]
	}

	return query
}

// --- SQL Injection Prevention ---

// Contains characters that could be used for SQL injection
// This is a defense-in-depth measure; parameterized queries are the primary protection
var sqlInjectionPatterns = []string{
	"--",
	";--",
	";",
	"/*",
	"*/",
	"@@",
	"char(",
	"nchar(",
	"varchar(",
	"nvarchar(",
	"alter",
	"begin",
	"cast(",
	"create",
	"cursor",
	"declare",
	"delete",
	"drop",
	"exec(",
	"execute(",
	"fetch",
	"insert",
	"kill",
	"select",
	"sys.",
	"sysobjects",
	"syscolumns",
	"table",
	"update",
	"union",
}

// ContainsSQLInjection checks if a string contains potential SQL injection patterns
// This is defense-in-depth; always use parameterized queries
func ContainsSQLInjection(s string) bool {
	lower := strings.ToLower(s)
	for _, pattern := range sqlInjectionPatterns {
		if strings.Contains(lower, pattern) {
			return true
		}
	}
	return false
}

// --- Command Injection Prevention ---

// shellMetaChars are characters that have special meaning in shell contexts
var shellMetaChars = []string{
	"|", "&", ";", "$", ">", "<", "`", "\\", "'", "\"",
	"\n", "\r", "(", ")", "{", "}", "[", "]", "!", "~",
}

// ContainsShellMetaChars checks if a string contains shell metacharacters
func ContainsShellMetaChars(s string) bool {
	for _, char := range shellMetaChars {
		if strings.Contains(s, char) {
			return true
		}
	}
	return false
}

// --- XSS Prevention ---

// xssPatterns contains common XSS attack patterns
var xssPatterns = []string{
	"<script",
	"</script",
	"javascript:",
	"vbscript:",
	"onload=",
	"onerror=",
	"onclick=",
	"onmouseover=",
	"onfocus=",
	"onblur=",
	"<iframe",
	"<object",
	"<embed",
	"<svg",
	"<math",
	"expression(",
	"url(",
	"data:",
}

// ContainsXSSPattern checks if a string contains potential XSS patterns
func ContainsXSSPattern(s string) bool {
	lower := strings.ToLower(s)
	for _, pattern := range xssPatterns {
		if strings.Contains(lower, pattern) {
			return true
		}
	}
	return false
}

// --- Webhook Signature Validation ---

// WebhookSignatureValidator validates webhook signatures
type WebhookSignatureValidator struct{}

// ValidateHMACSHA256 validates an HMAC-SHA256 signature
// The actual crypto operations are in the webhook package; this is for validation logic
func (v *WebhookSignatureValidator) ValidateSignatureFormat(signature string) error {
	if signature == "" {
		return &ValidationError{
			Field:   "signature",
			Message: "signature is required",
			Code:    "required",
		}
	}

	// HMAC-SHA256 produces 64 hex characters
	if len(signature) != 64 {
		return &ValidationError{
			Field:   "signature",
			Message: "invalid signature format",
			Code:    "invalid_format",
		}
	}

	// Must be hex characters only
	for _, c := range signature {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
			return &ValidationError{
				Field:   "signature",
				Message: "signature contains invalid characters",
				Code:    "invalid_chars",
			}
		}
	}

	return nil
}
