package providers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gorax/gorax/internal/oauth"
)

func TestTwitterProvider_Key(t *testing.T) {
	provider := NewTwitterProvider()
	assert.Equal(t, "twitter", provider.Key())
}

func TestTwitterProvider_Name(t *testing.T) {
	provider := NewTwitterProvider()
	assert.Equal(t, "Twitter", provider.Name())
}

func TestTwitterProvider_GetAuthURL(t *testing.T) {
	tests := []struct {
		name          string
		clientID      string
		redirectURI   string
		state         string
		scopes        []string
		codeChallenge string
		wantContains  []string
	}{
		{
			name:          "basic auth URL with PKCE",
			clientID:      "test-client-id",
			redirectURI:   "https://example.com/callback",
			state:         "test-state",
			scopes:        []string{"tweet.read", "users.read", "offline.access"},
			codeChallenge: "test-challenge",
			wantContains: []string{
				"https://api.twitter.com/2/oauth2/authorize",
				"client_id=test-client-id",
				"redirect_uri=https%3A%2F%2Fexample.com%2Fcallback",
				"state=test-state",
				"scope=tweet.read+users.read+offline.access",
				"code_challenge=test-challenge",
				"code_challenge_method=S256",
				"response_type=code",
			},
		},
		{
			name:          "with different scopes",
			clientID:      "client123",
			redirectURI:   "http://localhost:8080/callback",
			state:         "state123",
			scopes:        []string{"tweet.write", "follows.read"},
			codeChallenge: "challenge123",
			wantContains: []string{
				"scope=tweet.write+follows.read",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := NewTwitterProvider()
			authURL := provider.GetAuthURL(tt.clientID, tt.redirectURI, tt.state, tt.scopes, tt.codeChallenge)

			for _, want := range tt.wantContains {
				assert.Contains(t, authURL, want)
			}

			// Parse URL to validate structure
			parsedURL, err := url.Parse(authURL)
			require.NoError(t, err)
			assert.Equal(t, "api.twitter.com", parsedURL.Host)
			assert.Equal(t, "/2/oauth2/authorize", parsedURL.Path)
		})
	}
}

func TestTwitterProvider_ExchangeCode(t *testing.T) {
	tests := []struct {
		name           string
		serverResponse interface{}
		serverStatus   int
		clientID       string
		clientSecret   string
		code           string
		redirectURI    string
		codeVerifier   string
		wantToken      *oauth.TokenResponse
		wantErr        bool
		errContains    string
	}{
		{
			name: "successful token exchange with refresh token",
			serverResponse: map[string]interface{}{
				"access_token":  "twitter-access-token",
				"token_type":    "bearer",
				"expires_in":    7200,
				"refresh_token": "twitter-refresh-token",
				"scope":         "tweet.read users.read offline.access",
			},
			serverStatus: http.StatusOK,
			clientID:     "test-client",
			clientSecret: "test-secret",
			code:         "auth-code",
			redirectURI:  "https://example.com/callback",
			codeVerifier: "verifier123",
			wantToken: &oauth.TokenResponse{
				AccessToken:  "twitter-access-token",
				TokenType:    "bearer",
				ExpiresIn:    7200,
				RefreshToken: "twitter-refresh-token",
				Scope:        "tweet.read users.read offline.access",
			},
			wantErr: false,
		},
		{
			name: "successful token exchange without refresh token",
			serverResponse: map[string]interface{}{
				"access_token": "twitter-access-token",
				"token_type":   "bearer",
				"expires_in":   7200,
			},
			serverStatus: http.StatusOK,
			clientID:     "test-client",
			clientSecret: "test-secret",
			code:         "auth-code",
			redirectURI:  "https://example.com/callback",
			codeVerifier: "verifier123",
			wantToken: &oauth.TokenResponse{
				AccessToken: "twitter-access-token",
				TokenType:   "bearer",
				ExpiresIn:   7200,
			},
			wantErr: false,
		},
		{
			name: "error response from Twitter",
			serverResponse: map[string]interface{}{
				"error":             "invalid_grant",
				"error_description": "Invalid authorization code",
			},
			serverStatus: http.StatusBadRequest,
			clientID:     "test-client",
			clientSecret: "test-secret",
			code:         "invalid-code",
			redirectURI:  "https://example.com/callback",
			codeVerifier: "verifier123",
			wantErr:      true,
			errContains:  "token request failed with status 400",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Validate request method and path
				assert.Equal(t, http.MethodPost, r.Method)
				assert.Equal(t, "/2/oauth2/token", r.URL.Path)

				// Validate headers
				assert.Equal(t, "application/x-www-form-urlencoded", r.Header.Get("Content-Type"))

				// Parse form data
				err := r.ParseForm()
				require.NoError(t, err)
				assert.Equal(t, tt.clientID, r.FormValue("client_id"))
				assert.Equal(t, tt.clientSecret, r.FormValue("client_secret"))
				assert.Equal(t, tt.code, r.FormValue("code"))
				assert.Equal(t, tt.redirectURI, r.FormValue("redirect_uri"))
				assert.Equal(t, "authorization_code", r.FormValue("grant_type"))
				assert.Equal(t, tt.codeVerifier, r.FormValue("code_verifier"))

				// Send response
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.serverStatus)
				err = json.NewEncoder(w).Encode(tt.serverResponse)
				require.NoError(t, err)
			}))
			defer server.Close()

			// Create provider with custom token URL
			provider := NewTwitterProvider()
			provider.tokenURL = server.URL + "/2/oauth2/token"

			// Execute test
			ctx := context.Background()
			token, err := provider.ExchangeCode(ctx, tt.clientID, tt.clientSecret, tt.code, tt.redirectURI, tt.codeVerifier)

			if tt.wantErr {
				require.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
				assert.Nil(t, token)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.wantToken.AccessToken, token.AccessToken)
				assert.Equal(t, tt.wantToken.TokenType, token.TokenType)
				assert.Equal(t, tt.wantToken.ExpiresIn, token.ExpiresIn)
				assert.Equal(t, tt.wantToken.RefreshToken, token.RefreshToken)
				assert.Equal(t, tt.wantToken.Scope, token.Scope)
			}
		})
	}
}

func TestTwitterProvider_RefreshToken(t *testing.T) {
	tests := []struct {
		name           string
		serverResponse interface{}
		serverStatus   int
		clientID       string
		clientSecret   string
		refreshToken   string
		wantToken      *oauth.TokenResponse
		wantErr        bool
		errContains    string
	}{
		{
			name: "successful token refresh",
			serverResponse: map[string]interface{}{
				"access_token":  "new-twitter-access-token",
				"token_type":    "bearer",
				"expires_in":    7200,
				"refresh_token": "new-twitter-refresh-token",
			},
			serverStatus: http.StatusOK,
			clientID:     "test-client",
			clientSecret: "test-secret",
			refreshToken: "old-refresh-token",
			wantToken: &oauth.TokenResponse{
				AccessToken:  "new-twitter-access-token",
				TokenType:    "bearer",
				ExpiresIn:    7200,
				RefreshToken: "new-twitter-refresh-token",
			},
			wantErr: false,
		},
		{
			name: "error response from Twitter",
			serverResponse: map[string]interface{}{
				"error":             "invalid_grant",
				"error_description": "Invalid refresh token",
			},
			serverStatus: http.StatusBadRequest,
			clientID:     "test-client",
			clientSecret: "test-secret",
			refreshToken: "invalid-token",
			wantErr:      true,
			errContains:  "refresh request failed with status 400",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodPost, r.Method)
				assert.Equal(t, "/2/oauth2/token", r.URL.Path)

				// Parse form data
				err := r.ParseForm()
				require.NoError(t, err)
				assert.Equal(t, tt.clientID, r.FormValue("client_id"))
				assert.Equal(t, tt.clientSecret, r.FormValue("client_secret"))
				assert.Equal(t, tt.refreshToken, r.FormValue("refresh_token"))
				assert.Equal(t, "refresh_token", r.FormValue("grant_type"))

				// Send response
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.serverStatus)
				err = json.NewEncoder(w).Encode(tt.serverResponse)
				require.NoError(t, err)
			}))
			defer server.Close()

			provider := NewTwitterProvider()
			provider.tokenURL = server.URL + "/2/oauth2/token"

			ctx := context.Background()
			token, err := provider.RefreshToken(ctx, tt.clientID, tt.clientSecret, tt.refreshToken)

			if tt.wantErr {
				require.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.wantToken.AccessToken, token.AccessToken)
				assert.Equal(t, tt.wantToken.TokenType, token.TokenType)
				assert.Equal(t, tt.wantToken.ExpiresIn, token.ExpiresIn)
			}
		})
	}
}

func TestTwitterProvider_GetUserInfo(t *testing.T) {
	tests := []struct {
		name           string
		serverResponse interface{}
		serverStatus   int
		accessToken    string
		wantUserInfo   *oauth.UserInfo
		wantErr        bool
		errContains    string
	}{
		{
			name: "successful user info retrieval",
			serverResponse: map[string]interface{}{
				"data": map[string]interface{}{
					"id":       "123456789",
					"username": "testuser",
					"name":     "Test User",
				},
			},
			serverStatus: http.StatusOK,
			accessToken:  "valid-token",
			wantUserInfo: &oauth.UserInfo{
				ID:       "123456789",
				Username: "testuser",
				Name:     "Test User",
			},
			wantErr: false,
		},
		{
			name: "error response from Twitter",
			serverResponse: map[string]interface{}{
				"error": "Unauthorized",
			},
			serverStatus: http.StatusUnauthorized,
			accessToken:  "invalid-token",
			wantErr:      true,
			errContains:  "user info request failed with status 401",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodGet, r.Method)
				assert.Equal(t, "/2/users/me", r.URL.Path)
				assert.Equal(t, "Bearer "+tt.accessToken, r.Header.Get("Authorization"))

				// Check for user.fields query parameter
				assert.Contains(t, r.URL.Query().Get("user.fields"), "username")

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.serverStatus)
				err := json.NewEncoder(w).Encode(tt.serverResponse)
				require.NoError(t, err)
			}))
			defer server.Close()

			provider := NewTwitterProvider()
			provider.userInfoURL = server.URL + "/2/users/me"

			ctx := context.Background()
			userInfo, err := provider.GetUserInfo(ctx, tt.accessToken)

			if tt.wantErr {
				require.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.wantUserInfo.ID, userInfo.ID)
				assert.Equal(t, tt.wantUserInfo.Username, userInfo.Username)
				assert.Equal(t, tt.wantUserInfo.Name, userInfo.Name)
			}
		})
	}
}

func TestTwitterProvider_RevokeToken(t *testing.T) {
	tests := []struct {
		name         string
		serverStatus int
		accessToken  string
		wantErr      bool
		errContains  string
	}{
		{
			name:         "successful token revocation",
			serverStatus: http.StatusOK,
			accessToken:  "valid-token",
			wantErr:      false,
		},
		{
			name:         "error response from Twitter",
			serverStatus: http.StatusBadRequest,
			accessToken:  "invalid-token",
			wantErr:      true,
			errContains:  "revoke request failed with status 400",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodPost, r.Method)
				assert.Equal(t, "/2/oauth2/revoke", r.URL.Path)

				// Parse form data
				err := r.ParseForm()
				require.NoError(t, err)
				assert.Equal(t, tt.accessToken, r.FormValue("token"))

				// Basic auth should be present
				username, password, ok := r.BasicAuth()
				assert.True(t, ok)
				assert.NotEmpty(t, username)
				assert.NotEmpty(t, password)

				w.WriteHeader(tt.serverStatus)
			}))
			defer server.Close()

			provider := NewTwitterProvider()
			provider.revokeURL = server.URL + "/2/oauth2/revoke"

			ctx := context.Background()
			err := provider.RevokeToken(ctx, "client-id", "client-secret", tt.accessToken)

			if tt.wantErr {
				require.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestTwitterProvider_URLEncoding(t *testing.T) {
	provider := NewTwitterProvider()

	// Test special characters in scopes and state
	authURL := provider.GetAuthURL(
		"client-id",
		"https://example.com/callback?foo=bar",
		"state-with-special-chars-!@#$%",
		[]string{"tweet.read", "users.read"},
		"challenge123",
	)

	// Should properly encode redirect URI and state
	parsedURL, err := url.Parse(authURL)
	require.NoError(t, err)

	params := parsedURL.Query()
	assert.Equal(t, "https://example.com/callback?foo=bar", params.Get("redirect_uri"))
	assert.Equal(t, "state-with-special-chars-!@#$%", params.Get("state"))
}

func TestTwitterProvider_ContextCancellation(t *testing.T) {
	provider := NewTwitterProvider()

	// Create a canceled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// All operations should fail with context cancellation
	_, err := provider.ExchangeCode(ctx, "id", "secret", "code", "uri", "verifier")
	assert.Error(t, err)
	assert.Contains(t, strings.ToLower(err.Error()), "context")

	_, err = provider.RefreshToken(ctx, "id", "secret", "token")
	assert.Error(t, err)
	assert.Contains(t, strings.ToLower(err.Error()), "context")

	_, err = provider.GetUserInfo(ctx, "token")
	assert.Error(t, err)
	assert.Contains(t, strings.ToLower(err.Error()), "context")

	err = provider.RevokeToken(ctx, "id", "secret", "token")
	assert.Error(t, err)
	assert.Contains(t, strings.ToLower(err.Error()), "context")
}
