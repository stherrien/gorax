package auth

import (
	"net/http"

	"github.com/gorax/gorax/internal/integration"
)

// APIKeyLocation specifies where to place the API key.
type APIKeyLocation string

const (
	// APIKeyLocationHeader places the API key in a header.
	APIKeyLocationHeader APIKeyLocation = "header"
	// APIKeyLocationQuery places the API key in a query parameter.
	APIKeyLocationQuery APIKeyLocation = "query"
)

// APIKeyAuth authenticates requests using an API key.
type APIKeyAuth struct {
	key      string
	name     string // Header name or query parameter name
	location APIKeyLocation
}

// APIKeyAuthConfig holds configuration for API key authentication.
type APIKeyAuthConfig struct {
	Key      string         `json:"key"`
	Name     string         `json:"name"`     // e.g., "X-API-Key" or "api_key"
	Location APIKeyLocation `json:"location"` // "header" or "query"
}

// NewAPIKeyAuth creates a new API key authenticator.
func NewAPIKeyAuth(key, name string, location APIKeyLocation) *APIKeyAuth {
	if location == "" {
		location = APIKeyLocationHeader
	}
	if name == "" {
		name = "X-API-Key"
	}
	return &APIKeyAuth{
		key:      key,
		name:     name,
		location: location,
	}
}

// NewAPIKeyAuthFromCredentials creates an API key authenticator from credentials.
func NewAPIKeyAuthFromCredentials(creds *integration.Credentials) (*APIKeyAuth, error) {
	if creds == nil || creds.Data == nil {
		return nil, integration.NewValidationError("credentials", "credentials data is required", nil)
	}

	key, ok := creds.Data.GetString("key")
	if !ok || key == "" {
		return nil, integration.NewValidationError("key", "API key is required", nil)
	}

	name, _ := creds.Data.GetString("name")
	if name == "" {
		name = "X-API-Key"
	}

	locationStr, _ := creds.Data.GetString("location")
	location := APIKeyLocation(locationStr)
	if location == "" {
		location = APIKeyLocationHeader
	}

	return NewAPIKeyAuth(key, name, location), nil
}

// NewAPIKeyAuthFromConfig creates an API key authenticator from config.
func NewAPIKeyAuthFromConfig(config *APIKeyAuthConfig) (*APIKeyAuth, error) {
	if config == nil {
		return nil, integration.NewValidationError("config", "config is required", nil)
	}
	if config.Key == "" {
		return nil, integration.NewValidationError("key", "API key is required", nil)
	}

	return NewAPIKeyAuth(config.Key, config.Name, config.Location), nil
}

// Authenticate adds the API key to the request.
func (a *APIKeyAuth) Authenticate(req *http.Request) error {
	if err := a.Validate(); err != nil {
		return err
	}

	switch a.location {
	case APIKeyLocationQuery:
		q := req.URL.Query()
		q.Set(a.name, a.key)
		req.URL.RawQuery = q.Encode()
	case APIKeyLocationHeader:
		fallthrough
	default:
		req.Header.Set(a.name, a.key)
	}

	return nil
}

// Type returns the credential type.
func (a *APIKeyAuth) Type() integration.CredentialType {
	return integration.CredTypeAPIKey
}

// Validate validates the authenticator's configuration.
func (a *APIKeyAuth) Validate() error {
	if a.key == "" {
		return integration.NewValidationError("key", "API key cannot be empty", nil)
	}
	if a.name == "" {
		return integration.NewValidationError("name", "key name cannot be empty", nil)
	}
	if a.location != APIKeyLocationHeader && a.location != APIKeyLocationQuery {
		return integration.NewValidationError("location", "invalid API key location", a.location)
	}
	return nil
}

// Key returns the API key (for testing/debugging - use with caution).
func (a *APIKeyAuth) Key() string {
	return a.key
}

// Name returns the header/parameter name.
func (a *APIKeyAuth) Name() string {
	return a.name
}

// Location returns the API key location.
func (a *APIKeyAuth) Location() APIKeyLocation {
	return a.location
}
