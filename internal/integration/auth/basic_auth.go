package auth

import (
	"net/http"

	"github.com/gorax/gorax/internal/integration"
)

// BasicAuth authenticates requests using HTTP Basic Authentication.
type BasicAuth struct {
	username string
	password string
}

// BasicAuthConfig holds configuration for basic authentication.
type BasicAuthConfig struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// NewBasicAuth creates a new basic auth authenticator.
func NewBasicAuth(username, password string) *BasicAuth {
	return &BasicAuth{
		username: username,
		password: password,
	}
}

// NewBasicAuthFromCredentials creates a basic auth authenticator from credentials.
func NewBasicAuthFromCredentials(creds *integration.Credentials) (*BasicAuth, error) {
	if creds == nil || creds.Data == nil {
		return nil, integration.NewValidationError("credentials", "credentials data is required", nil)
	}

	username, ok := creds.Data.GetString("username")
	if !ok || username == "" {
		return nil, integration.NewValidationError("username", "username is required", nil)
	}

	password, _ := creds.Data.GetString("password")
	// Password can be empty for some systems

	return NewBasicAuth(username, password), nil
}

// NewBasicAuthFromConfig creates a basic auth authenticator from config.
func NewBasicAuthFromConfig(config *BasicAuthConfig) (*BasicAuth, error) {
	if config == nil {
		return nil, integration.NewValidationError("config", "config is required", nil)
	}
	if config.Username == "" {
		return nil, integration.NewValidationError("username", "username is required", nil)
	}

	return NewBasicAuth(config.Username, config.Password), nil
}

// Authenticate adds basic auth to the request.
func (a *BasicAuth) Authenticate(req *http.Request) error {
	if err := a.Validate(); err != nil {
		return err
	}

	req.SetBasicAuth(a.username, a.password)
	return nil
}

// Type returns the credential type.
func (a *BasicAuth) Type() integration.CredentialType {
	return integration.CredTypeBasicAuth
}

// Validate validates the authenticator's configuration.
func (a *BasicAuth) Validate() error {
	if a.username == "" {
		return integration.NewValidationError("username", "username cannot be empty", nil)
	}
	// Password can be empty
	return nil
}

// Username returns the username.
func (a *BasicAuth) Username() string {
	return a.username
}
