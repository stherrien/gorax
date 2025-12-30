package credential

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"
)

// JSONMap is a custom type for storing JSON in PostgreSQL
// Implements driver.Valuer and sql.Scanner for automatic serialization
type JSONMap map[string]interface{}

// Value implements driver.Valuer for database serialization
func (j JSONMap) Value() (driver.Value, error) {
	if j == nil {
		return []byte("{}"), nil // Return empty JSON object instead of NULL
	}
	return json.Marshal(j)
}

// Scan implements sql.Scanner for database deserialization
func (j *JSONMap) Scan(value interface{}) error {
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

// Common errors
var (
	ErrUnauthorized  = errors.New("unauthorized access to credential")
	ErrInvalidInput  = errors.New("invalid input")
	ErrAlreadyExists = errors.New("credential already exists")
)

// ValidationError represents a validation error
type ValidationError struct {
	Message string
}

func (e *ValidationError) Error() string {
	return e.Message
}

// CredentialType represents the type of credential
type CredentialType string

const (
	TypeAPIKey      CredentialType = "api_key"
	TypeOAuth2      CredentialType = "oauth2"
	TypeBasicAuth   CredentialType = "basic_auth"
	TypeBearerToken CredentialType = "bearer_token"
	TypeCustom      CredentialType = "custom"
)

// CredentialStatus represents the status of a credential
type CredentialStatus string

const (
	StatusActive   CredentialStatus = "active"
	StatusInactive CredentialStatus = "inactive"
	StatusRevoked  CredentialStatus = "revoked"
)

// AccessType constants
const (
	AccessTypeRead   = "read"
	AccessTypeUpdate = "update"
	AccessTypeRotate = "rotate"
	AccessTypeDelete = "delete"
)

// Credential represents a credential in the system
// Values are encrypted at rest and should never be returned in API responses except through /value endpoint
type Credential struct {
	ID          string           `json:"id" db:"id"`
	TenantID    string           `json:"tenant_id" db:"tenant_id"`
	Name        string           `json:"name" db:"name"`
	Description string           `json:"description" db:"description"`
	Type        CredentialType   `json:"type" db:"type"`
	Status      CredentialStatus `json:"status" db:"status"`
	CreatedBy   string           `json:"created_by" db:"created_by"`
	CreatedAt   time.Time        `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time        `json:"updated_at" db:"updated_at"`
	LastUsedAt  *time.Time       `json:"last_used_at,omitempty" db:"last_used_at"`
	ExpiresAt   *time.Time       `json:"expires_at,omitempty" db:"expires_at"`

	// Envelope encryption fields
	EncryptedDEK []byte `json:"-" db:"encrypted_dek"` // Never serialize
	Ciphertext   []byte `json:"-" db:"ciphertext"`    // Never serialize
	Nonce        []byte `json:"-" db:"nonce"`         // Never serialize
	AuthTag      []byte `json:"-" db:"auth_tag"`      // Never serialize
	KMSKeyID     string `json:"-" db:"kms_key_id"`    // Never serialize

	// Metadata stored as JSON
	Metadata JSONMap `json:"metadata,omitempty" db:"metadata"`
}

// IsExpired checks if the credential has expired
func (c *Credential) IsExpired() bool {
	if c.ExpiresAt == nil {
		return false
	}
	return time.Now().After(*c.ExpiresAt)
}

// CredentialData represents plaintext credential data before encryption
// This is used internally by the encryption service
type CredentialData struct {
	Value map[string]interface{} `json:"value"`
}

// EncryptedSecret represents an encrypted credential value using envelope encryption
type EncryptedSecret struct {
	EncryptedDEK []byte `json:"encrypted_dek"` // Data Encryption Key encrypted by KMS
	Ciphertext   []byte `json:"ciphertext"`    // Data encrypted by DEK
	Nonce        []byte `json:"nonce"`         // Nonce for GCM
	AuthTag      []byte `json:"auth_tag"`      // Authentication tag for GCM
	KMSKeyID     string `json:"kms_key_id"`    // KMS key ID used to encrypt DEK
}

// CredentialValue represents the encrypted value of a credential
// This is stored separately and only returned through secure endpoints
type CredentialValue struct {
	ID             string    `json:"id" db:"id"`
	CredentialID   string    `json:"credential_id" db:"credential_id"`
	Version        int       `json:"version" db:"version"`
	EncryptedValue string    `json:"-" db:"encrypted_value"` // Never serialize
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
	CreatedBy      string    `json:"created_by" db:"created_by"`
	IsActive       bool      `json:"is_active" db:"is_active"`
}

// AccessLog represents an access log entry for credential usage
type AccessLog struct {
	ID           string    `json:"id" db:"id"`
	CredentialID string    `json:"credential_id" db:"credential_id"`
	TenantID     string    `json:"tenant_id" db:"tenant_id"`
	AccessedBy   string    `json:"accessed_by" db:"accessed_by"`
	AccessType   string    `json:"access_type" db:"access_type"` // "read", "update", "rotate", "delete"
	AccessedAt   time.Time `json:"accessed_at" db:"accessed_at"`
	IPAddress    string    `json:"ip_address,omitempty" db:"ip_address"`
	UserAgent    string    `json:"user_agent,omitempty" db:"user_agent"`
	Success      bool      `json:"success" db:"success"`
	ErrorMessage string    `json:"error_message,omitempty" db:"error_message"`
}

// CreateCredentialInput represents input for creating a credential
type CreateCredentialInput struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Type        CredentialType         `json:"type"`
	Value       map[string]interface{} `json:"value"` // Will be encrypted
	ExpiresAt   *time.Time             `json:"expires_at,omitempty"`
	Metadata    JSONMap                `json:"metadata,omitempty"`
}

// UpdateCredentialInput represents input for updating credential metadata
// Note: This does NOT update the value - use Rotate for that
type UpdateCredentialInput struct {
	Name        *string           `json:"name,omitempty"`
	Description *string           `json:"description,omitempty"`
	Status      *CredentialStatus `json:"status,omitempty"`
	ExpiresAt   *time.Time        `json:"expires_at,omitempty"`
	Metadata    JSONMap           `json:"metadata,omitempty"`
}

// RotateCredentialInput represents input for rotating a credential value
type RotateCredentialInput struct {
	Value map[string]interface{} `json:"value"` // New value to encrypt
}

// CredentialListFilter represents filters for listing credentials
type CredentialListFilter struct {
	Type   CredentialType   `json:"type,omitempty"`
	Status CredentialStatus `json:"status,omitempty"`
	Search string           `json:"search,omitempty"` // Search in name/description
}

// DecryptedValue represents a decrypted credential value
// Only used internally and in the /value endpoint response
type DecryptedValue struct {
	Version   int                    `json:"version"`
	Value     map[string]interface{} `json:"value"`
	CreatedAt time.Time              `json:"created_at"`
}

// Validate validates CreateCredentialInput
func (c *CreateCredentialInput) Validate() error {
	if c.Name == "" {
		return &ValidationError{Message: "name is required"}
	}
	if len(c.Name) > 255 {
		return &ValidationError{Message: "name must be less than 255 characters"}
	}
	if c.Type == "" {
		return &ValidationError{Message: "type is required"}
	}
	if c.Type != TypeAPIKey && c.Type != TypeOAuth2 && c.Type != TypeBasicAuth && c.Type != TypeBearerToken && c.Type != TypeCustom {
		return &ValidationError{Message: "invalid credential type"}
	}
	if len(c.Value) == 0 {
		return &ValidationError{Message: "value is required"}
	}
	return nil
}

// Validate validates UpdateCredentialInput
func (u *UpdateCredentialInput) Validate() error {
	if u.Name != nil && len(*u.Name) > 255 {
		return &ValidationError{Message: "name must be less than 255 characters"}
	}
	if u.Status != nil {
		if *u.Status != StatusActive && *u.Status != StatusInactive && *u.Status != StatusRevoked {
			return &ValidationError{Message: "invalid status"}
		}
	}
	return nil
}

// Validate validates RotateCredentialInput
func (r *RotateCredentialInput) Validate() error {
	if len(r.Value) == 0 {
		return &ValidationError{Message: "value is required"}
	}
	return nil
}
