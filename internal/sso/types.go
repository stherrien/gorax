package sso

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
)

// ProviderType represents the type of SSO provider
type ProviderType string

const (
	ProviderTypeSAML ProviderType = "saml"
	ProviderTypeOIDC ProviderType = "oidc"
)

// LoginStatus represents the status of an SSO login attempt
type LoginStatus string

const (
	LoginStatusSuccess LoginStatus = "success"
	LoginStatusFailure LoginStatus = "failure"
	LoginStatusError   LoginStatus = "error"
)

// UserAttributes represents user attributes extracted from SSO provider
type UserAttributes struct {
	ExternalID string            `json:"external_id"`
	Email      string            `json:"email"`
	FirstName  string            `json:"first_name,omitempty"`
	LastName   string            `json:"last_name,omitempty"`
	Groups     []string          `json:"groups,omitempty"`
	Attributes map[string]string `json:"attributes,omitempty"`
}

// SSOProvider defines the interface for SSO authentication providers
type SSOProvider interface {
	// GetType returns the provider type (SAML or OIDC)
	GetType() ProviderType

	// InitiateLogin generates the SSO login URL and any necessary state
	InitiateLogin(ctx context.Context, relayState string) (redirectURL string, err error)

	// HandleCallback processes the SSO callback and extracts user attributes
	HandleCallback(ctx context.Context, r *http.Request) (*UserAttributes, error)

	// GetMetadata returns provider metadata (for SAML SP metadata)
	GetMetadata(ctx context.Context) (string, error)

	// Validate validates the provider configuration
	Validate(ctx context.Context) error
}

// Provider represents an SSO provider configuration
type Provider struct {
	ID         uuid.UUID       `json:"id" db:"id"`
	TenantID   uuid.UUID       `json:"tenant_id" db:"tenant_id"`
	Name       string          `json:"name" db:"name"`
	Type       ProviderType    `json:"provider_type" db:"provider_type"`
	Enabled    bool            `json:"enabled" db:"enabled"`
	EnforceSSO bool            `json:"enforce_sso" db:"enforce_sso"`
	Config     json.RawMessage `json:"config" db:"config"`
	Domains    []string        `json:"domains" db:"domains"`
	CreatedAt  time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time       `json:"updated_at" db:"updated_at"`
	CreatedBy  *uuid.UUID      `json:"created_by,omitempty" db:"created_by"`
	UpdatedBy  *uuid.UUID      `json:"updated_by,omitempty" db:"updated_by"`
}

// Connection represents a user's connection to an SSO provider
type Connection struct {
	ID          uuid.UUID       `json:"id" db:"id"`
	UserID      uuid.UUID       `json:"user_id" db:"user_id"`
	ProviderID  uuid.UUID       `json:"sso_provider_id" db:"sso_provider_id"`
	ExternalID  string          `json:"external_id" db:"external_id"`
	Attributes  json.RawMessage `json:"attributes,omitempty" db:"attributes"`
	LastLoginAt *time.Time      `json:"last_login_at,omitempty" db:"last_login_at"`
	CreatedAt   time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at" db:"updated_at"`
}

// LoginEvent represents an SSO login event for audit purposes
type LoginEvent struct {
	ID           uuid.UUID   `json:"id" db:"id"`
	ProviderID   uuid.UUID   `json:"sso_provider_id" db:"sso_provider_id"`
	UserID       *uuid.UUID  `json:"user_id,omitempty" db:"user_id"`
	ExternalID   string      `json:"external_id" db:"external_id"`
	Status       LoginStatus `json:"status" db:"status"`
	ErrorMessage *string     `json:"error_message,omitempty" db:"error_message"`
	IPAddress    *string     `json:"ip_address,omitempty" db:"ip_address"`
	UserAgent    *string     `json:"user_agent,omitempty" db:"user_agent"`
	CreatedAt    time.Time   `json:"created_at" db:"created_at"`
}

// AuthenticationRequest represents an SSO authentication request
type AuthenticationRequest struct {
	ProviderID uuid.UUID `json:"provider_id"`
	RelayState string    `json:"relay_state,omitempty"`
}

// AuthenticationResponse represents the result of SSO authentication
type AuthenticationResponse struct {
	UserAttributes UserAttributes `json:"user_attributes"`
	SessionToken   string         `json:"session_token"`
	ExpiresAt      time.Time      `json:"expires_at"`
}

// CreateProviderRequest represents a request to create an SSO provider
type CreateProviderRequest struct {
	Name       string          `json:"name"`
	Type       ProviderType    `json:"provider_type"`
	Enabled    bool            `json:"enabled"`
	EnforceSSO bool            `json:"enforce_sso"`
	Config     json.RawMessage `json:"config"`
	Domains    []string        `json:"domains"`
}

// UpdateProviderRequest represents a request to update an SSO provider
type UpdateProviderRequest struct {
	Name       *string         `json:"name,omitempty"`
	Enabled    *bool           `json:"enabled,omitempty"`
	EnforceSSO *bool           `json:"enforce_sso,omitempty"`
	Config     json.RawMessage `json:"config,omitempty"`
	Domains    []string        `json:"domains,omitempty"`
}

// ProviderFactory creates SSO providers from configuration
type ProviderFactory interface {
	CreateProvider(ctx context.Context, provider *Provider) (SSOProvider, error)
}
