package auth

import (
	"net/http"

	"github.com/gorax/gorax/internal/integration"
)

// BearerTokenAuth authenticates requests using a bearer token.
type BearerTokenAuth struct {
	token  string
	scheme string // Default: "Bearer"
}

// BearerTokenAuthConfig holds configuration for bearer token authentication.
type BearerTokenAuthConfig struct {
	Token  string `json:"token"`
	Scheme string `json:"scheme,omitempty"` // Default: "Bearer"
}

// NewBearerTokenAuth creates a new bearer token authenticator.
func NewBearerTokenAuth(token string) *BearerTokenAuth {
	return &BearerTokenAuth{
		token:  token,
		scheme: "Bearer",
	}
}

// NewBearerTokenAuthWithScheme creates a new bearer token authenticator with a custom scheme.
func NewBearerTokenAuthWithScheme(token, scheme string) *BearerTokenAuth {
	if scheme == "" {
		scheme = "Bearer"
	}
	return &BearerTokenAuth{
		token:  token,
		scheme: scheme,
	}
}

// NewBearerTokenAuthFromCredentials creates a bearer token authenticator from credentials.
func NewBearerTokenAuthFromCredentials(creds *integration.Credentials) (*BearerTokenAuth, error) {
	if creds == nil || creds.Data == nil {
		return nil, integration.NewValidationError("credentials", "credentials data is required", nil)
	}

	token, ok := creds.Data.GetString("token")
	if !ok || token == "" {
		return nil, integration.NewValidationError("token", "token is required", nil)
	}

	scheme, _ := creds.Data.GetString("scheme")
	if scheme == "" {
		scheme = "Bearer"
	}

	return NewBearerTokenAuthWithScheme(token, scheme), nil
}

// NewBearerTokenAuthFromConfig creates a bearer token authenticator from config.
func NewBearerTokenAuthFromConfig(config *BearerTokenAuthConfig) (*BearerTokenAuth, error) {
	if config == nil {
		return nil, integration.NewValidationError("config", "config is required", nil)
	}
	if config.Token == "" {
		return nil, integration.NewValidationError("token", "token is required", nil)
	}

	return NewBearerTokenAuthWithScheme(config.Token, config.Scheme), nil
}

// Authenticate adds the bearer token to the request.
func (a *BearerTokenAuth) Authenticate(req *http.Request) error {
	if err := a.Validate(); err != nil {
		return err
	}

	req.Header.Set("Authorization", a.scheme+" "+a.token)
	return nil
}

// Type returns the credential type.
func (a *BearerTokenAuth) Type() integration.CredentialType {
	return integration.CredTypeBearerToken
}

// Validate validates the authenticator's configuration.
func (a *BearerTokenAuth) Validate() error {
	if a.token == "" {
		return integration.NewValidationError("token", "token cannot be empty", nil)
	}
	return nil
}

// Token returns the token (for testing/debugging - use with caution).
func (a *BearerTokenAuth) Token() string {
	return a.token
}

// Scheme returns the authentication scheme.
func (a *BearerTokenAuth) Scheme() string {
	return a.scheme
}
