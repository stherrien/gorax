// Package auth provides authentication patterns for integration requests.
package auth

import (
	"net/http"

	"github.com/gorax/gorax/internal/integration"
)

// Authenticator is the interface for authenticating HTTP requests.
type Authenticator interface {
	// Authenticate adds authentication credentials to the HTTP request.
	Authenticate(req *http.Request) error

	// Type returns the authentication type.
	Type() integration.CredentialType

	// Validate validates the authenticator's configuration.
	Validate() error
}

// RefreshableAuthenticator extends Authenticator with credential refresh capability.
type RefreshableAuthenticator interface {
	Authenticator

	// NeedsRefresh returns true if the credentials need to be refreshed.
	NeedsRefresh() bool

	// Refresh refreshes the credentials.
	Refresh() error
}

// AuthenticatorFactory creates authenticators from credentials.
type AuthenticatorFactory struct{}

// NewAuthenticatorFactory creates a new AuthenticatorFactory.
func NewAuthenticatorFactory() *AuthenticatorFactory {
	return &AuthenticatorFactory{}
}

// Create creates an Authenticator from credentials.
func (f *AuthenticatorFactory) Create(creds *integration.Credentials) (Authenticator, error) {
	if creds == nil {
		return nil, integration.NewValidationError("credentials", "credentials cannot be nil", nil)
	}

	switch creds.Type {
	case integration.CredTypeAPIKey:
		return NewAPIKeyAuthFromCredentials(creds)
	case integration.CredTypeBearerToken:
		return NewBearerTokenAuthFromCredentials(creds)
	case integration.CredTypeBasicAuth:
		return NewBasicAuthFromCredentials(creds)
	case integration.CredTypeOAuth2:
		return NewOAuth2AuthFromCredentials(creds)
	default:
		return nil, integration.NewValidationError("type", "unsupported credential type", creds.Type)
	}
}

// CreateFromConfig creates an Authenticator from integration configuration.
func (f *AuthenticatorFactory) CreateFromConfig(config *integration.Config) (Authenticator, error) {
	if config == nil || config.Credentials == nil {
		return nil, nil // No authentication configured
	}
	return f.Create(config.Credentials)
}

// NoAuth is an authenticator that does nothing (no authentication).
type NoAuth struct{}

// NewNoAuth creates a new NoAuth authenticator.
func NewNoAuth() *NoAuth {
	return &NoAuth{}
}

// Authenticate does nothing.
func (a *NoAuth) Authenticate(_ *http.Request) error {
	return nil
}

// Type returns the credential type.
func (a *NoAuth) Type() integration.CredentialType {
	return integration.CredTypeCustom
}

// Validate always returns nil.
func (a *NoAuth) Validate() error {
	return nil
}
