package integration

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"maps"
	"time"
)

// IntegrationType represents the type of integration.
type IntegrationType string

// Integration types.
const (
	TypeHTTP    IntegrationType = "http"
	TypeWebhook IntegrationType = "webhook"
	TypeAPI     IntegrationType = "api"
	TypeCustom  IntegrationType = "custom"
	TypePlugin  IntegrationType = "plugin"
)

// Valid returns whether the integration type is valid.
func (t IntegrationType) Valid() bool {
	switch t {
	case TypeHTTP, TypeWebhook, TypeAPI, TypeCustom, TypePlugin:
		return true
	default:
		return false
	}
}

// String returns the string representation.
func (t IntegrationType) String() string {
	return string(t)
}

// FieldType represents the type of a configuration or input field.
type FieldType string

// Field types.
const (
	FieldTypeString  FieldType = "string"
	FieldTypeNumber  FieldType = "number"
	FieldTypeInteger FieldType = "integer"
	FieldTypeBoolean FieldType = "boolean"
	FieldTypeArray   FieldType = "array"
	FieldTypeObject  FieldType = "object"
	FieldTypeSecret  FieldType = "secret"
)

// Valid returns whether the field type is valid.
func (f FieldType) Valid() bool {
	switch f {
	case FieldTypeString, FieldTypeNumber, FieldTypeInteger, FieldTypeBoolean,
		FieldTypeArray, FieldTypeObject, FieldTypeSecret:
		return true
	default:
		return false
	}
}

// FieldSpec defines the specification for a configuration or input field.
type FieldSpec struct {
	Name        string    `json:"name"`
	Type        FieldType `json:"type"`
	Description string    `json:"description,omitempty"`
	Required    bool      `json:"required"`
	Default     any       `json:"default,omitempty"`
	Options     []string  `json:"options,omitempty"` // For enum fields
	Sensitive   bool      `json:"sensitive"`         // Mask in logs if true
	Pattern     string    `json:"pattern,omitempty"` // Regex pattern for validation
	MinLength   int       `json:"min_length,omitempty"`
	MaxLength   int       `json:"max_length,omitempty"`
	Min         *float64  `json:"min,omitempty"`
	Max         *float64  `json:"max,omitempty"`
}

// Schema defines the configuration schema for an integration.
type Schema struct {
	ConfigSpec map[string]FieldSpec `json:"config_spec"`
	InputSpec  map[string]FieldSpec `json:"input_spec"`
	OutputSpec map[string]FieldSpec `json:"output_spec"`
}

// Config holds the configuration for an integration instance.
type Config struct {
	ID          string          `json:"id"`
	Name        string          `json:"name"`
	Type        IntegrationType `json:"type"`
	Description string          `json:"description,omitempty"`
	Version     string          `json:"version,omitempty"`
	Enabled     bool            `json:"enabled"`
	Settings    JSONMap         `json:"settings"`
	Credentials *Credentials    `json:"credentials,omitempty"`
	Metadata    JSONMap         `json:"metadata,omitempty"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
}

// Credentials holds credential information for an integration.
type Credentials struct {
	ID        string         `json:"id"`
	Type      CredentialType `json:"type"`
	Name      string         `json:"name,omitempty"`
	Data      JSONMap        `json:"data,omitempty"`      // Decrypted data (in memory only)
	Encrypted *EncryptedData `json:"encrypted,omitempty"` // Encrypted storage
	ExpiresAt *time.Time     `json:"expires_at,omitempty"`
	RefreshAt *time.Time     `json:"refresh_at,omitempty"`
	Metadata  JSONMap        `json:"metadata,omitempty"`
}

// CredentialType represents the type of credential.
type CredentialType string

// Credential types.
const (
	CredTypeAPIKey      CredentialType = "api_key"
	CredTypeBearerToken CredentialType = "bearer_token"
	CredTypeBasicAuth   CredentialType = "basic_auth"
	CredTypeOAuth2      CredentialType = "oauth2"
	CredTypeCustom      CredentialType = "custom"
)

// Valid returns whether the credential type is valid.
func (c CredentialType) Valid() bool {
	switch c {
	case CredTypeAPIKey, CredTypeBearerToken, CredTypeBasicAuth, CredTypeOAuth2, CredTypeCustom:
		return true
	default:
		return false
	}
}

// EncryptedData holds encrypted credential data.
type EncryptedData struct {
	EncryptedDEK []byte `json:"encrypted_dek"`
	Ciphertext   []byte `json:"ciphertext"`
	Nonce        []byte `json:"nonce"`
	AuthTag      []byte `json:"auth_tag,omitempty"`
	KMSKeyID     string `json:"kms_key_id,omitempty"`
}

// Result represents the result of an integration execution.
type Result struct {
	Success    bool      `json:"success"`
	Data       any       `json:"data,omitempty"`
	Error      string    `json:"error,omitempty"`
	ErrorCode  string    `json:"error_code,omitempty"`
	StatusCode int       `json:"status_code,omitempty"`
	Headers    JSONMap   `json:"headers,omitempty"`
	Duration   int64     `json:"duration_ms"`
	RetryCount int       `json:"retry_count,omitempty"`
	Metadata   JSONMap   `json:"metadata,omitempty"`
	ExecutedAt time.Time `json:"executed_at"`
}

// NewSuccessResult creates a successful result.
func NewSuccessResult(data any, duration int64) *Result {
	return &Result{
		Success:    true,
		Data:       data,
		Duration:   duration,
		ExecutedAt: time.Now().UTC(),
	}
}

// NewErrorResult creates an error result.
func NewErrorResult(err error, errorCode string, duration int64) *Result {
	errMsg := ""
	if err != nil {
		errMsg = err.Error()
	}
	return &Result{
		Success:    false,
		Error:      errMsg,
		ErrorCode:  errorCode,
		Duration:   duration,
		ExecutedAt: time.Now().UTC(),
	}
}

// ExecutionContext provides context for integration execution.
type ExecutionContext struct {
	ExecutionID string  `json:"execution_id"`
	WorkflowID  string  `json:"workflow_id,omitempty"`
	StepID      string  `json:"step_id,omitempty"`
	TenantID    string  `json:"tenant_id"`
	UserID      string  `json:"user_id,omitempty"`
	TraceID     string  `json:"trace_id,omitempty"`
	SpanID      string  `json:"span_id,omitempty"`
	Metadata    JSONMap `json:"metadata,omitempty"`
}

// Metadata holds integration metadata.
type Metadata struct {
	Name        string            `json:"name"`
	DisplayName string            `json:"display_name,omitempty"`
	Description string            `json:"description,omitempty"`
	Version     string            `json:"version"`
	Author      string            `json:"author,omitempty"`
	License     string            `json:"license,omitempty"`
	Homepage    string            `json:"homepage,omitempty"`
	Tags        []string          `json:"tags,omitempty"`
	Category    string            `json:"category,omitempty"`
	Schema      *Schema           `json:"schema,omitempty"`
	Permissions []string          `json:"permissions,omitempty"`
	Extra       map[string]string `json:"extra,omitempty"`
}

// JSONMap is a map that can be stored in and retrieved from databases.
type JSONMap map[string]any

// Value implements driver.Valuer for database serialization.
func (j JSONMap) Value() (driver.Value, error) {
	if j == nil {
		return []byte("{}"), nil
	}
	return json.Marshal(j)
}

// Scan implements sql.Scanner for database deserialization.
func (j *JSONMap) Scan(value any) error {
	if value == nil {
		*j = nil
		return nil
	}

	var data []byte
	switch v := value.(type) {
	case []byte:
		data = v
	case string:
		data = []byte(v)
	default:
		return errors.New("unsupported type for JSONMap")
	}

	return json.Unmarshal(data, j)
}

// Get retrieves a value from the map with type assertion.
func (j JSONMap) Get(key string) (any, bool) {
	if j == nil {
		return nil, false
	}
	v, ok := j[key]
	return v, ok
}

// GetString retrieves a string value from the map.
func (j JSONMap) GetString(key string) (string, bool) {
	v, ok := j.Get(key)
	if !ok {
		return "", false
	}
	s, ok := v.(string)
	return s, ok
}

// GetInt retrieves an integer value from the map.
func (j JSONMap) GetInt(key string) (int, bool) {
	v, ok := j.Get(key)
	if !ok {
		return 0, false
	}
	switch n := v.(type) {
	case int:
		return n, true
	case int64:
		return int(n), true
	case float64:
		return int(n), true
	default:
		return 0, false
	}
}

// GetBool retrieves a boolean value from the map.
func (j JSONMap) GetBool(key string) (bool, bool) {
	v, ok := j.Get(key)
	if !ok {
		return false, false
	}
	b, ok := v.(bool)
	return b, ok
}

// Merge merges another map into this one.
func (j JSONMap) Merge(other JSONMap) JSONMap {
	result := make(JSONMap, len(j)+len(other))
	maps.Copy(result, j)
	maps.Copy(result, other)
	return result
}
