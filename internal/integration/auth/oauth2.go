package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/gorax/gorax/internal/integration"
)

// OAuth2GrantType represents the OAuth2 grant type.
type OAuth2GrantType string

const (
	// GrantTypeClientCredentials is the client credentials grant type.
	GrantTypeClientCredentials OAuth2GrantType = "client_credentials"
	// GrantTypePassword is the resource owner password credentials grant type.
	GrantTypePassword OAuth2GrantType = "password"
	// GrantTypeRefreshToken is the refresh token grant type.
	GrantTypeRefreshToken OAuth2GrantType = "refresh_token"
	// GrantTypeAuthorizationCode is the authorization code grant type.
	GrantTypeAuthorizationCode OAuth2GrantType = "authorization_code"
)

// OAuth2Auth authenticates requests using OAuth2.
type OAuth2Auth struct {
	clientID     string
	clientSecret string
	tokenURL     string
	grantType    OAuth2GrantType
	scopes       []string
	audience     string

	// For password grant
	username string
	password string

	// Token state
	accessToken  string
	refreshToken string
	tokenExpiry  time.Time

	// HTTP client for token requests
	httpClient *http.Client

	mu sync.RWMutex
}

// OAuth2AuthConfig holds configuration for OAuth2 authentication.
type OAuth2AuthConfig struct {
	ClientID     string          `json:"client_id"`
	ClientSecret string          `json:"client_secret"`
	TokenURL     string          `json:"token_url"`
	GrantType    OAuth2GrantType `json:"grant_type"`
	Scopes       []string        `json:"scopes,omitempty"`
	Audience     string          `json:"audience,omitempty"`
	Username     string          `json:"username,omitempty"` // For password grant
	Password     string          `json:"password,omitempty"` // For password grant
	AccessToken  string          `json:"access_token,omitempty"`
	RefreshToken string          `json:"refresh_token,omitempty"`
}

// TokenResponse represents an OAuth2 token response.
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token,omitempty"`
	Scope        string `json:"scope,omitempty"`
}

// NewOAuth2Auth creates a new OAuth2 authenticator.
func NewOAuth2Auth(config *OAuth2AuthConfig) (*OAuth2Auth, error) {
	if config == nil {
		return nil, integration.NewValidationError("config", "config is required", nil)
	}

	auth := &OAuth2Auth{
		clientID:     config.ClientID,
		clientSecret: config.ClientSecret,
		tokenURL:     config.TokenURL,
		grantType:    config.GrantType,
		scopes:       config.Scopes,
		audience:     config.Audience,
		username:     config.Username,
		password:     config.Password,
		accessToken:  config.AccessToken,
		refreshToken: config.RefreshToken,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}

	if config.GrantType == "" {
		auth.grantType = GrantTypeClientCredentials
	}

	return auth, nil
}

// NewOAuth2AuthFromCredentials creates an OAuth2 authenticator from credentials.
func NewOAuth2AuthFromCredentials(creds *integration.Credentials) (*OAuth2Auth, error) {
	if creds == nil || creds.Data == nil {
		return nil, integration.NewValidationError("credentials", "credentials data is required", nil)
	}

	clientID, _ := creds.Data.GetString("client_id")
	clientSecret, _ := creds.Data.GetString("client_secret")
	tokenURL, _ := creds.Data.GetString("token_url")
	grantType, _ := creds.Data.GetString("grant_type")
	audience, _ := creds.Data.GetString("audience")
	username, _ := creds.Data.GetString("username")
	password, _ := creds.Data.GetString("password")
	accessToken, _ := creds.Data.GetString("access_token")
	refreshToken, _ := creds.Data.GetString("refresh_token")

	var scopes []string
	if scopesVal, ok := creds.Data.Get("scopes"); ok {
		if scopesArr, ok := scopesVal.([]any); ok {
			for _, s := range scopesArr {
				if str, ok := s.(string); ok {
					scopes = append(scopes, str)
				}
			}
		}
	}

	config := &OAuth2AuthConfig{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		TokenURL:     tokenURL,
		GrantType:    OAuth2GrantType(grantType),
		Scopes:       scopes,
		Audience:     audience,
		Username:     username,
		Password:     password,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}

	return NewOAuth2Auth(config)
}

// Authenticate adds the OAuth2 access token to the request.
func (a *OAuth2Auth) Authenticate(req *http.Request) error {
	token, err := a.GetAccessToken(req.Context())
	if err != nil {
		return fmt.Errorf("getting access token: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+token)
	return nil
}

// Type returns the credential type.
func (a *OAuth2Auth) Type() integration.CredentialType {
	return integration.CredTypeOAuth2
}

// Validate validates the authenticator's configuration.
func (a *OAuth2Auth) Validate() error {
	// If we have an access token, we might be okay
	a.mu.RLock()
	hasToken := a.accessToken != ""
	a.mu.RUnlock()

	if hasToken {
		return nil
	}

	// Otherwise we need token URL and credentials
	if a.tokenURL == "" {
		return integration.NewValidationError("token_url", "token URL is required", nil)
	}

	switch a.grantType {
	case GrantTypeClientCredentials:
		if a.clientID == "" {
			return integration.NewValidationError("client_id", "client ID is required for client_credentials grant", nil)
		}
		if a.clientSecret == "" {
			return integration.NewValidationError("client_secret", "client secret is required for client_credentials grant", nil)
		}
	case GrantTypePassword:
		if a.username == "" {
			return integration.NewValidationError("username", "username is required for password grant", nil)
		}
		if a.password == "" {
			return integration.NewValidationError("password", "password is required for password grant", nil)
		}
	case GrantTypeRefreshToken:
		a.mu.RLock()
		hasRefresh := a.refreshToken != ""
		a.mu.RUnlock()
		if !hasRefresh {
			return integration.NewValidationError("refresh_token", "refresh token is required for refresh_token grant", nil)
		}
	}

	return nil
}

// NeedsRefresh returns true if the access token needs to be refreshed.
func (a *OAuth2Auth) NeedsRefresh() bool {
	a.mu.RLock()
	defer a.mu.RUnlock()

	if a.accessToken == "" {
		return true
	}

	if a.tokenExpiry.IsZero() {
		return false // Unknown expiry, assume it's valid
	}

	// Refresh if token expires in less than 1 minute
	return time.Until(a.tokenExpiry) < time.Minute
}

// Refresh refreshes the access token.
func (a *OAuth2Auth) Refresh() error {
	return a.refreshToken_(context.Background())
}

// GetAccessToken returns the current access token, refreshing if necessary.
func (a *OAuth2Auth) GetAccessToken(ctx context.Context) (string, error) {
	if !a.NeedsRefresh() {
		a.mu.RLock()
		token := a.accessToken
		a.mu.RUnlock()
		return token, nil
	}

	if err := a.refreshToken_(ctx); err != nil {
		return "", err
	}

	a.mu.RLock()
	token := a.accessToken
	a.mu.RUnlock()
	return token, nil
}

// refreshToken_ performs the token refresh.
func (a *OAuth2Auth) refreshToken_(ctx context.Context) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	// Double-check after acquiring lock
	if a.accessToken != "" && !a.tokenExpiry.IsZero() && time.Until(a.tokenExpiry) >= time.Minute {
		return nil
	}

	// Build token request
	data := url.Values{}

	switch a.grantType {
	case GrantTypeClientCredentials:
		data.Set("grant_type", "client_credentials")
		data.Set("client_id", a.clientID)
		data.Set("client_secret", a.clientSecret)
	case GrantTypePassword:
		data.Set("grant_type", "password")
		data.Set("username", a.username)
		data.Set("password", a.password)
		if a.clientID != "" {
			data.Set("client_id", a.clientID)
		}
		if a.clientSecret != "" {
			data.Set("client_secret", a.clientSecret)
		}
	case GrantTypeRefreshToken:
		data.Set("grant_type", "refresh_token")
		data.Set("refresh_token", a.refreshToken)
		if a.clientID != "" {
			data.Set("client_id", a.clientID)
		}
		if a.clientSecret != "" {
			data.Set("client_secret", a.clientSecret)
		}
	default:
		return integration.NewValidationError("grant_type", "unsupported grant type", a.grantType)
	}

	if len(a.scopes) > 0 {
		data.Set("scope", strings.Join(a.scopes, " "))
	}

	if a.audience != "" {
		data.Set("audience", a.audience)
	}

	// Create request
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, a.tokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return fmt.Errorf("creating token request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Execute request
	resp, err := a.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("executing token request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("reading token response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return integration.NewHTTPError(resp.StatusCode, resp.Status, string(body))
	}

	// Parse response
	var tokenResp TokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return fmt.Errorf("parsing token response: %w", err)
	}

	// Update token state
	a.accessToken = tokenResp.AccessToken
	if tokenResp.RefreshToken != "" {
		a.refreshToken = tokenResp.RefreshToken
	}
	if tokenResp.ExpiresIn > 0 {
		a.tokenExpiry = time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)
	} else {
		a.tokenExpiry = time.Time{} // Unknown expiry
	}

	return nil
}

// SetToken sets the access token directly (useful for pre-existing tokens).
func (a *OAuth2Auth) SetToken(accessToken string, expiresIn int) {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.accessToken = accessToken
	if expiresIn > 0 {
		a.tokenExpiry = time.Now().Add(time.Duration(expiresIn) * time.Second)
	}
}

// SetRefreshToken sets the refresh token.
func (a *OAuth2Auth) SetRefreshToken(refreshToken string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.refreshToken = refreshToken
}

// TokenExpiry returns when the current token expires.
func (a *OAuth2Auth) TokenExpiry() time.Time {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.tokenExpiry
}

// HasRefreshToken returns true if a refresh token is available.
func (a *OAuth2Auth) HasRefreshToken() bool {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.refreshToken != ""
}
