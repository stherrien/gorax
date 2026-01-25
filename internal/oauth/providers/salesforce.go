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
	salesforceProductionAuthURL     = "https://login.salesforce.com/services/oauth2/authorize"
	salesforceProductionTokenURL    = "https://login.salesforce.com/services/oauth2/token"
	salesforceProductionUserInfoURL = "https://login.salesforce.com/services/oauth2/userinfo"
	salesforceProductionRevokeURL   = "https://login.salesforce.com/services/oauth2/revoke"

	salesforceSandboxAuthURL     = "https://test.salesforce.com/services/oauth2/authorize"
	salesforceSandboxTokenURL    = "https://test.salesforce.com/services/oauth2/token"
	salesforceSandboxUserInfoURL = "https://test.salesforce.com/services/oauth2/userinfo"
	salesforceSandboxRevokeURL   = "https://test.salesforce.com/services/oauth2/revoke"
)

// SalesforceProvider implements OAuth 2.0 for Salesforce
// Supports both production and sandbox environments
// Salesforce supports PKCE for enhanced security
type SalesforceProvider struct {
	httpClient  *http.Client
	authURL     string
	tokenURL    string
	userInfoURL string
	revokeURL   string
	isSandbox   bool
}

// NewSalesforceProvider creates a new Salesforce OAuth provider
// isSandbox determines whether to use sandbox (test.salesforce.com) or production (login.salesforce.com) URLs
func NewSalesforceProvider(isSandbox bool) *SalesforceProvider {
	var authURL, tokenURL, userInfoURL, revokeURL string

	if isSandbox {
		authURL = salesforceSandboxAuthURL
		tokenURL = salesforceSandboxTokenURL
		userInfoURL = salesforceSandboxUserInfoURL
		revokeURL = salesforceSandboxRevokeURL
	} else {
		authURL = salesforceProductionAuthURL
		tokenURL = salesforceProductionTokenURL
		userInfoURL = salesforceProductionUserInfoURL
		revokeURL = salesforceProductionRevokeURL
	}

	return &SalesforceProvider{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		authURL:     authURL,
		tokenURL:    tokenURL,
		userInfoURL: userInfoURL,
		revokeURL:   revokeURL,
		isSandbox:   isSandbox,
	}
}

// Key returns the provider key
func (p *SalesforceProvider) Key() string {
	return "salesforce"
}

// Name returns the provider name
func (p *SalesforceProvider) Name() string {
	return "Salesforce"
}

// GetAuthURL returns the authorization URL
// Salesforce supports PKCE (Proof Key for Code Exchange) for enhanced security
func (p *SalesforceProvider) GetAuthURL(clientID, redirectURI, state string, scopes []string, codeChallenge string) string {
	params := url.Values{}
	params.Set("client_id", clientID)
	params.Set("redirect_uri", redirectURI)
	params.Set("state", state)
	params.Set("scope", strings.Join(scopes, " "))
	params.Set("response_type", "code")

	// Salesforce supports PKCE
	if codeChallenge != "" {
		params.Set("code_challenge", codeChallenge)
		params.Set("code_challenge_method", "S256")
	}

	return fmt.Sprintf("%s?%s", p.authURL, params.Encode())
}

// ExchangeCode exchanges authorization code for tokens
func (p *SalesforceProvider) ExchangeCode(ctx context.Context, clientID, clientSecret, code, redirectURI, codeVerifier string) (*oauth.TokenResponse, error) {
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
func (p *SalesforceProvider) RefreshToken(ctx context.Context, clientID, clientSecret, refreshToken string) (*oauth.TokenResponse, error) {
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
func (p *SalesforceProvider) GetUserInfo(ctx context.Context, accessToken string) (*oauth.UserInfo, error) {
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

	// Salesforce returns OpenID Connect standard userinfo with additional fields
	var salesforceUser struct {
		Sub               string `json:"sub"`                // User identity URL
		UserID            string `json:"user_id"`            // Salesforce user ID
		OrganizationID    string `json:"organization_id"`    // Salesforce org ID
		Name              string `json:"name"`               // Full name
		Email             string `json:"email"`              // Email address
		EmailVerified     bool   `json:"email_verified"`     // Email verification status
		PreferredUsername string `json:"preferred_username"` // Username (usually email)
	}

	if err := json.Unmarshal(body, &salesforceUser); err != nil {
		return nil, fmt.Errorf("failed to parse user info response: %w", err)
	}

	return &oauth.UserInfo{
		ID:       salesforceUser.UserID,
		Username: salesforceUser.PreferredUsername,
		Email:    salesforceUser.Email,
		Name:     salesforceUser.Name,
	}, nil
}

// RevokeToken revokes a token
func (p *SalesforceProvider) RevokeToken(ctx context.Context, clientID, clientSecret, token string) error {
	data := url.Values{}
	data.Set("token", token)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.revokeURL, strings.NewReader(data.Encode()))
	if err != nil {
		return fmt.Errorf("failed to create revoke request: %w", err)
	}

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
