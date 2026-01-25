package oauth

import "context"

// TokenResponse represents an OAuth token response
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in,omitempty"`
	RefreshToken string `json:"refresh_token,omitempty"`
	Scope        string `json:"scope,omitempty"`
	IDToken      string `json:"id_token,omitempty"`
}

// UserInfo represents OAuth provider user information
type UserInfo struct {
	ID       string `json:"id"`
	Username string `json:"username,omitempty"`
	Email    string `json:"email,omitempty"`
	Name     string `json:"name,omitempty"`
}

// Provider defines the OAuth provider interface
type Provider interface {
	// Key returns the provider key (github, google, slack, etc.)
	Key() string

	// Name returns the provider display name
	Name() string

	// GetAuthURL returns the authorization URL
	GetAuthURL(clientID, redirectURI, state string, scopes []string, codeChallenge string) string

	// ExchangeCode exchanges authorization code for tokens
	ExchangeCode(ctx context.Context, clientID, clientSecret, code, redirectURI, codeVerifier string) (*TokenResponse, error)

	// RefreshToken refreshes an access token
	RefreshToken(ctx context.Context, clientID, clientSecret, refreshToken string) (*TokenResponse, error)

	// GetUserInfo retrieves user information
	GetUserInfo(ctx context.Context, accessToken string) (*UserInfo, error)

	// RevokeToken revokes a token (if supported)
	RevokeToken(ctx context.Context, clientID, clientSecret, token string) error
}
