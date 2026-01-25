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
	slackAuthURL     = "https://slack.com/oauth/v2/authorize"
	slackTokenURL    = "https://slack.com/api/oauth.v2.access"
	slackUserInfoURL = "https://slack.com/api/users.identity"
	slackRevokeURL   = "https://slack.com/api/auth.revoke"
)

// SlackProvider implements OAuth for Slack
type SlackProvider struct {
	httpClient *http.Client
}

// NewSlackProvider creates a new Slack OAuth provider
func NewSlackProvider() *SlackProvider {
	return &SlackProvider{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Key returns the provider key
func (p *SlackProvider) Key() string {
	return "slack"
}

// Name returns the provider name
func (p *SlackProvider) Name() string {
	return "Slack"
}

// GetAuthURL returns the authorization URL
func (p *SlackProvider) GetAuthURL(clientID, redirectURI, state string, scopes []string, codeChallenge string) string {
	params := url.Values{}
	params.Set("client_id", clientID)
	params.Set("redirect_uri", redirectURI)
	params.Set("state", state)
	params.Set("scope", strings.Join(scopes, ","))
	params.Set("user_scope", "identity.basic,identity.email")

	// Slack supports PKCE
	if codeChallenge != "" {
		params.Set("code_challenge", codeChallenge)
		params.Set("code_challenge_method", "S256")
	}

	return fmt.Sprintf("%s?%s", slackAuthURL, params.Encode())
}

// ExchangeCode exchanges authorization code for tokens
func (p *SlackProvider) ExchangeCode(ctx context.Context, clientID, clientSecret, code, redirectURI, codeVerifier string) (*oauth.TokenResponse, error) {
	data := url.Values{}
	data.Set("client_id", clientID)
	data.Set("client_secret", clientSecret)
	data.Set("code", code)
	data.Set("redirect_uri", redirectURI)

	// Include code_verifier for PKCE
	if codeVerifier != "" {
		data.Set("code_verifier", codeVerifier)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, slackTokenURL, strings.NewReader(data.Encode()))
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

	var slackResp struct {
		OK          bool   `json:"ok"`
		Error       string `json:"error,omitempty"`
		AccessToken string `json:"access_token"`
		TokenType   string `json:"token_type"`
		Scope       string `json:"scope"`
		BotUserID   string `json:"bot_user_id,omitempty"`
		AppID       string `json:"app_id"`
		Team        struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"team"`
		AuthedUser struct {
			ID          string `json:"id"`
			Scope       string `json:"scope"`
			AccessToken string `json:"access_token"`
			TokenType   string `json:"token_type"`
		} `json:"authed_user"`
	}

	if err := json.Unmarshal(body, &slackResp); err != nil {
		return nil, fmt.Errorf("failed to parse token response: %w", err)
	}

	if !slackResp.OK {
		return nil, fmt.Errorf("slack API error: %s", slackResp.Error)
	}

	return &oauth.TokenResponse{
		AccessToken: slackResp.AccessToken,
		TokenType:   slackResp.TokenType,
		Scope:       slackResp.Scope,
	}, nil
}

// RefreshToken refreshes an access token
// Note: Slack doesn't support refresh tokens - tokens don't expire unless revoked
func (p *SlackProvider) RefreshToken(ctx context.Context, clientID, clientSecret, refreshToken string) (*oauth.TokenResponse, error) {
	return nil, fmt.Errorf("Slack OAuth does not support refresh tokens")
}

// GetUserInfo retrieves user information
func (p *SlackProvider) GetUserInfo(ctx context.Context, accessToken string) (*oauth.UserInfo, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, slackUserInfoURL, nil)
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

	var slackResp struct {
		OK    bool   `json:"ok"`
		Error string `json:"error,omitempty"`
		User  struct {
			ID    string `json:"id"`
			Name  string `json:"name"`
			Email string `json:"email"`
		} `json:"user"`
	}

	if err := json.Unmarshal(body, &slackResp); err != nil {
		return nil, fmt.Errorf("failed to parse user info response: %w", err)
	}

	if !slackResp.OK {
		return nil, fmt.Errorf("slack API error: %s", slackResp.Error)
	}

	return &oauth.UserInfo{
		ID:       slackResp.User.ID,
		Username: slackResp.User.Name,
		Email:    slackResp.User.Email,
		Name:     slackResp.User.Name,
	}, nil
}

// RevokeToken revokes a token
func (p *SlackProvider) RevokeToken(ctx context.Context, clientID, clientSecret, token string) error {
	data := url.Values{}
	data.Set("token", token)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, slackRevokeURL, strings.NewReader(data.Encode()))
	if err != nil {
		return fmt.Errorf("failed to create revoke request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("revoke request failed: %w", err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			// Log error but don't override the main error
		}
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read revoke response: %w", err)
	}

	var slackResp struct {
		OK    bool   `json:"ok"`
		Error string `json:"error,omitempty"`
	}

	if err := json.Unmarshal(body, &slackResp); err != nil {
		return fmt.Errorf("failed to parse revoke response: %w", err)
	}

	if !slackResp.OK {
		return fmt.Errorf("slack revoke error: %s", slackResp.Error)
	}

	return nil
}
