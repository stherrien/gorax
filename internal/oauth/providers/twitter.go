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
	twitterAuthURL     = "https://api.twitter.com/2/oauth2/authorize"
	twitterTokenURL    = "https://api.twitter.com/2/oauth2/token"
	twitterUserInfoURL = "https://api.twitter.com/2/users/me"
	twitterRevokeURL   = "https://api.twitter.com/2/oauth2/revoke"
)

// TwitterProvider implements OAuth 2.0 for Twitter/X
// Twitter supports OAuth 2.0 with PKCE for enhanced security
type TwitterProvider struct {
	httpClient  *http.Client
	authURL     string
	tokenURL    string
	userInfoURL string
	revokeURL   string
}

// NewTwitterProvider creates a new Twitter OAuth provider
func NewTwitterProvider() *TwitterProvider {
	return &TwitterProvider{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		authURL:     twitterAuthURL,
		tokenURL:    twitterTokenURL,
		userInfoURL: twitterUserInfoURL,
		revokeURL:   twitterRevokeURL,
	}
}

// Key returns the provider key
func (p *TwitterProvider) Key() string {
	return "twitter"
}

// Name returns the provider name
func (p *TwitterProvider) Name() string {
	return "Twitter"
}

// GetAuthURL returns the authorization URL
// Twitter supports PKCE (Proof Key for Code Exchange) for enhanced security
func (p *TwitterProvider) GetAuthURL(clientID, redirectURI, state string, scopes []string, codeChallenge string) string {
	params := url.Values{}
	params.Set("client_id", clientID)
	params.Set("redirect_uri", redirectURI)
	params.Set("state", state)
	params.Set("scope", strings.Join(scopes, " "))
	params.Set("response_type", "code")

	// Twitter supports PKCE
	if codeChallenge != "" {
		params.Set("code_challenge", codeChallenge)
		params.Set("code_challenge_method", "S256")
	}

	return fmt.Sprintf("%s?%s", p.authURL, params.Encode())
}

// ExchangeCode exchanges authorization code for tokens
func (p *TwitterProvider) ExchangeCode(ctx context.Context, clientID, clientSecret, code, redirectURI, codeVerifier string) (*oauth.TokenResponse, error) {
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
func (p *TwitterProvider) RefreshToken(ctx context.Context, clientID, clientSecret, refreshToken string) (*oauth.TokenResponse, error) {
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

// GetUserInfo retrieves user information
func (p *TwitterProvider) GetUserInfo(ctx context.Context, accessToken string) (*oauth.UserInfo, error) {
	// Build URL with user.fields query parameter to get username and name
	userInfoURL := fmt.Sprintf("%s?user.fields=username,name", p.userInfoURL)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, userInfoURL, nil)
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

	// Twitter API v2 wraps user data in a "data" field
	var twitterResponse struct {
		Data struct {
			ID       string `json:"id"`
			Username string `json:"username"`
			Name     string `json:"name"`
		} `json:"data"`
	}

	if err := json.Unmarshal(body, &twitterResponse); err != nil {
		return nil, fmt.Errorf("failed to parse user info response: %w", err)
	}

	return &oauth.UserInfo{
		ID:       twitterResponse.Data.ID,
		Username: twitterResponse.Data.Username,
		Name:     twitterResponse.Data.Name,
	}, nil
}

// RevokeToken revokes a token
func (p *TwitterProvider) RevokeToken(ctx context.Context, clientID, clientSecret, token string) error {
	data := url.Values{}
	data.Set("token", token)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.revokeURL, strings.NewReader(data.Encode()))
	if err != nil {
		return fmt.Errorf("failed to create revoke request: %w", err)
	}

	// Twitter requires basic auth for revocation
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
