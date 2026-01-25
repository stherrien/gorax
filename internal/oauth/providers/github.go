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
	githubAuthURL     = "https://github.com/login/oauth/authorize"
	githubTokenURL    = "https://github.com/login/oauth/access_token"
	githubUserInfoURL = "https://api.github.com/user"
	githubRevokeURL   = "https://api.github.com/applications/%s/token"
)

// GitHubProvider implements OAuth for GitHub
type GitHubProvider struct {
	httpClient *http.Client
}

// NewGitHubProvider creates a new GitHub OAuth provider
func NewGitHubProvider() *GitHubProvider {
	return &GitHubProvider{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Key returns the provider key
func (p *GitHubProvider) Key() string {
	return "github"
}

// Name returns the provider name
func (p *GitHubProvider) Name() string {
	return "GitHub"
}

// GetAuthURL returns the authorization URL
func (p *GitHubProvider) GetAuthURL(clientID, redirectURI, state string, scopes []string, codeChallenge string) string {
	params := url.Values{}
	params.Set("client_id", clientID)
	params.Set("redirect_uri", redirectURI)
	params.Set("state", state)
	params.Set("scope", strings.Join(scopes, " "))

	// Note: GitHub doesn't support PKCE (code_challenge parameter)
	// Accepting codeChallenge for interface consistency, but not using it
	_ = codeChallenge

	return fmt.Sprintf("%s?%s", githubAuthURL, params.Encode())
}

// ExchangeCode exchanges authorization code for tokens
func (p *GitHubProvider) ExchangeCode(ctx context.Context, clientID, clientSecret, code, redirectURI, codeVerifier string) (*oauth.TokenResponse, error) {
	data := url.Values{}
	data.Set("client_id", clientID)
	data.Set("client_secret", clientSecret)
	data.Set("code", code)
	data.Set("redirect_uri", redirectURI)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, githubTokenURL, strings.NewReader(data.Encode()))
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
// Note: GitHub doesn't support refresh tokens for OAuth Apps
func (p *GitHubProvider) RefreshToken(ctx context.Context, clientID, clientSecret, refreshToken string) (*oauth.TokenResponse, error) {
	return nil, fmt.Errorf("GitHub OAuth Apps do not support refresh tokens")
}

// GetUserInfo retrieves user information
func (p *GitHubProvider) GetUserInfo(ctx context.Context, accessToken string) (*oauth.UserInfo, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, githubUserInfoURL, nil)
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

	var githubUser struct {
		ID    int64  `json:"id"`
		Login string `json:"login"`
		Email string `json:"email"`
		Name  string `json:"name"`
	}

	if err := json.Unmarshal(body, &githubUser); err != nil {
		return nil, fmt.Errorf("failed to parse user info response: %w", err)
	}

	return &oauth.UserInfo{
		ID:       fmt.Sprintf("%d", githubUser.ID),
		Username: githubUser.Login,
		Email:    githubUser.Email,
		Name:     githubUser.Name,
	}, nil
}

// RevokeToken revokes a token
func (p *GitHubProvider) RevokeToken(ctx context.Context, clientID, clientSecret, token string) error {
	revokeURL := fmt.Sprintf(githubRevokeURL, clientID)

	data := map[string]string{
		"access_token": token,
	}
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal revoke request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, revokeURL, strings.NewReader(string(jsonData)))
	if err != nil {
		return fmt.Errorf("failed to create revoke request: %w", err)
	}

	// GitHub requires basic auth for revocation
	req.SetBasicAuth(clientID, clientSecret)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("revoke request failed: %w", err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			// Log error but don't override the main error
		}
	}()

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("revoke request failed with status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}
