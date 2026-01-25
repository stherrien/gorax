package providers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gorax/gorax/internal/oauth"
)

// Auth0Provider implements OAuth 2.0 / OpenID Connect for Auth0
// Auth0 uses tenant-specific domains (e.g., tenant.auth0.com, tenant.us.auth0.com)
// or custom domains (e.g., login.mycompany.com)
// Auth0 supports PKCE for enhanced security
type Auth0Provider struct {
	httpClient  *http.Client
	domain      string
	authURL     string
	tokenURL    string
	userInfoURL string
	revokeURL   string
}

// NewAuth0Provider creates a new Auth0 OAuth provider
// domain should be the Auth0 tenant domain (e.g., "tenant.auth0.com", "tenant.us.auth0.com")
// or a custom domain (e.g., "login.mycompany.com")
func NewAuth0Provider(domain string) *Auth0Provider {
	// Ensure domain doesn't have protocol prefix
	domain = strings.TrimPrefix(domain, "https://")
	domain = strings.TrimPrefix(domain, "http://")
	domain = strings.TrimSuffix(domain, "/")

	return &Auth0Provider{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		domain:      domain,
		authURL:     fmt.Sprintf("https://%s/authorize", domain),
		tokenURL:    fmt.Sprintf("https://%s/oauth/token", domain),
		userInfoURL: fmt.Sprintf("https://%s/userinfo", domain),
		revokeURL:   fmt.Sprintf("https://%s/oauth/revoke", domain),
	}
}

// Key returns the provider key
func (p *Auth0Provider) Key() string {
	return "auth0"
}

// Name returns the provider name
func (p *Auth0Provider) Name() string {
	return "Auth0"
}

// GetAuthURL returns the authorization URL
// Auth0 supports PKCE (Proof Key for Code Exchange) for enhanced security
func (p *Auth0Provider) GetAuthURL(clientID, redirectURI, state string, scopes []string, codeChallenge string) string {
	params := url.Values{}
	params.Set("client_id", clientID)
	params.Set("redirect_uri", redirectURI)
	params.Set("state", state)
	params.Set("scope", strings.Join(scopes, " "))
	params.Set("response_type", "code")

	// Auth0 supports PKCE
	if codeChallenge != "" {
		params.Set("code_challenge", codeChallenge)
		params.Set("code_challenge_method", "S256")
	}

	return fmt.Sprintf("%s?%s", p.authURL, params.Encode())
}

// ExchangeCode exchanges authorization code for tokens
func (p *Auth0Provider) ExchangeCode(ctx context.Context, clientID, clientSecret, code, redirectURI, codeVerifier string) (*oauth.TokenResponse, error) {
	data := url.Values{}
	data.Set("client_id", clientID)
	data.Set("client_secret", clientSecret)
	data.Set("code", code)
	data.Set("redirect_uri", redirectURI)
	data.Set("grant_type", "authorization_code")

	// Include code_verifier for PKCE
	if codeVerifier != "" {
		data.Set("code_verifier", codeVerifier)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.tokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create token request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("token request failed: %w", err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			// Log error but don't override the main error
		}
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read token response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("token request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var tokenResp oauth.TokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return nil, fmt.Errorf("failed to parse token response: %w", err)
	}

	return &tokenResp, nil
}

// RefreshToken refreshes an access token
func (p *Auth0Provider) RefreshToken(ctx context.Context, clientID, clientSecret, refreshToken string) (*oauth.TokenResponse, error) {
	data := url.Values{}
	data.Set("client_id", clientID)
	data.Set("client_secret", clientSecret)
	data.Set("refresh_token", refreshToken)
	data.Set("grant_type", "refresh_token")

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.tokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create refresh request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("refresh request failed: %w", err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			// Log error but don't override the main error
		}
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read refresh response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("refresh request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var tokenResp oauth.TokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return nil, fmt.Errorf("failed to parse refresh response: %w", err)
	}

	return &tokenResp, nil
}

// GetUserInfo retrieves user information using OpenID Connect userinfo endpoint
func (p *Auth0Provider) GetUserInfo(ctx context.Context, accessToken string) (*oauth.UserInfo, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, p.userInfoURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create user info request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))
	req.Header.Set("Accept", "application/json")

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("user info request failed: %w", err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			// Log error but don't override the main error
		}
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read user info response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("user info request failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Auth0 returns OpenID Connect standard userinfo with additional fields
	var auth0User struct {
		Sub      string `json:"sub"`      // User ID (e.g., "auth0|123456", "google-oauth2|123456")
		Nickname string `json:"nickname"` // Username/nickname
		Name     string `json:"name"`     // Full name
		Email    string `json:"email"`    // Email address
	}

	if err := json.Unmarshal(body, &auth0User); err != nil {
		return nil, fmt.Errorf("failed to parse user info response: %w", err)
	}

	return &oauth.UserInfo{
		ID:       auth0User.Sub,
		Username: auth0User.Nickname,
		Name:     auth0User.Name,
		Email:    auth0User.Email,
	}, nil
}

// RevokeToken revokes a token
func (p *Auth0Provider) RevokeToken(ctx context.Context, clientID, clientSecret, token string) error {
	data := url.Values{}
	data.Set("token", token)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.revokeURL, strings.NewReader(data.Encode()))
	if err != nil {
		return fmt.Errorf("failed to create revoke request: %w", err)
	}

	// Auth0 requires basic auth for revocation
	req.SetBasicAuth(clientID, clientSecret)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("revoke request failed: %w", err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			// Log error but don't override the main error
		}
	}()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("revoke request failed with status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}
