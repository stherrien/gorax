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

const (
	linkedinAuthURL     = "https://www.linkedin.com/oauth/v2/authorization"
	linkedinTokenURL    = "https://www.linkedin.com/oauth/v2/accessToken"
	linkedinUserInfoURL = "https://api.linkedin.com/v2/userinfo"
)

// LinkedInProvider implements OAuth 2.0 for LinkedIn
// Note: LinkedIn does not support PKCE or refresh tokens
type LinkedInProvider struct {
	httpClient  *http.Client
	authURL     string
	tokenURL    string
	userInfoURL string
}

// NewLinkedInProvider creates a new LinkedIn OAuth provider
func NewLinkedInProvider() *LinkedInProvider {
	return &LinkedInProvider{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		authURL:     linkedinAuthURL,
		tokenURL:    linkedinTokenURL,
		userInfoURL: linkedinUserInfoURL,
	}
}

// Key returns the provider key
func (p *LinkedInProvider) Key() string {
	return "linkedin"
}

// Name returns the provider name
func (p *LinkedInProvider) Name() string {
	return "LinkedIn"
}

// GetAuthURL returns the authorization URL
// Note: LinkedIn does not support PKCE (codeChallenge parameter is ignored)
func (p *LinkedInProvider) GetAuthURL(clientID, redirectURI, state string, scopes []string, codeChallenge string) string {
	params := url.Values{}
	params.Set("client_id", clientID)
	params.Set("redirect_uri", redirectURI)
	params.Set("state", state)
	params.Set("scope", strings.Join(scopes, " "))
	params.Set("response_type", "code")

	// Note: LinkedIn doesn't support PKCE (code_challenge parameter)
	// Accepting codeChallenge for interface consistency, but not using it
	_ = codeChallenge

	return fmt.Sprintf("%s?%s", p.authURL, params.Encode())
}

// ExchangeCode exchanges authorization code for tokens
func (p *LinkedInProvider) ExchangeCode(ctx context.Context, clientID, clientSecret, code, redirectURI, codeVerifier string) (*oauth.TokenResponse, error) {
	// Note: codeVerifier is not used as LinkedIn doesn't support PKCE
	_ = codeVerifier

	data := url.Values{}
	data.Set("client_id", clientID)
	data.Set("client_secret", clientSecret)
	data.Set("code", code)
	data.Set("redirect_uri", redirectURI)
	data.Set("grant_type", "authorization_code")

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
// Note: LinkedIn OAuth 2.0 does not support refresh tokens
func (p *LinkedInProvider) RefreshToken(ctx context.Context, clientID, clientSecret, refreshToken string) (*oauth.TokenResponse, error) {
	return nil, fmt.Errorf("LinkedIn OAuth 2.0 does not support refresh tokens")
}

// GetUserInfo retrieves user information using OpenID Connect userinfo endpoint
func (p *LinkedInProvider) GetUserInfo(ctx context.Context, accessToken string) (*oauth.UserInfo, error) {
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

	// LinkedIn returns OpenID Connect standard userinfo
	var linkedinUser struct {
		Sub   string `json:"sub"`   // User ID
		Name  string `json:"name"`  // Full name
		Email string `json:"email"` // Email address
	}

	if err := json.Unmarshal(body, &linkedinUser); err != nil {
		return nil, fmt.Errorf("failed to parse user info response: %w", err)
	}

	return &oauth.UserInfo{
		ID:    linkedinUser.Sub,
		Name:  linkedinUser.Name,
		Email: linkedinUser.Email,
	}, nil
}

// RevokeToken revokes a token
// Note: LinkedIn OAuth 2.0 does not support token revocation
func (p *LinkedInProvider) RevokeToken(ctx context.Context, clientID, clientSecret, token string) error {
	return fmt.Errorf("LinkedIn OAuth 2.0 does not support token revocation")
}
