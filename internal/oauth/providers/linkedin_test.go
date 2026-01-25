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

func TestLinkedInProvider_Key(t *testing.T) {
	provider := NewLinkedInProvider()
	assert.Equal(t, "linkedin", provider.Key())
}

func TestLinkedInProvider_Name(t *testing.T) {
	provider := NewLinkedInProvider()
	assert.Equal(t, "LinkedIn", provider.Name())
}

func TestLinkedInProvider_GetAuthURL(t *testing.T) {
	tests := []struct {
		name         string
		clientID     string
		redirectURI  string
		state        string
		scopes       []string
		wantContains []string
	}{
		{
			name:        "basic auth URL without PKCE",
			clientID:    "test-client-id",
			redirectURI: "https://example.com/callback",
			state:       "test-state",
			scopes:      []string{"profile", "email", "openid"},
			wantContains: []string{
				"https://www.linkedin.com/oauth/v2/authorization",
				"client_id=test-client-id",
				"redirect_uri=https%3A%2F%2Fexample.com%2Fcallback",
				"state=test-state",
				"scope=profile+email+openid",
				"response_type=code",
			},
		},
		{
			name:        "with different scopes",
			clientID:    "client123",
			redirectURI: "http://localhost:8080/callback",
			state:       "state123",
			scopes:      []string{"r_liteprofile", "r_emailaddress"},
			wantContains: []string{
				"scope=r_liteprofile+r_emailaddress",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := NewLinkedInProvider()
			authURL := provider.GetAuthURL(tt.clientID, tt.redirectURI, tt.state, tt.scopes, "unused-code-challenge")

			for _, want := range tt.wantContains {
				assert.Contains(t, authURL, want)
			}

			// LinkedIn doesn't support PKCE, so code_challenge should not be in URL
			assert.NotContains(t, authURL, "code_challenge")

			parsedURL, err := url.Parse(authURL)
			require.NoError(t, err)
			assert.Equal(t, "www.linkedin.com", parsedURL.Host)
			assert.Equal(t, "/oauth/v2/authorization", parsedURL.Path)
		})
	}
}

func TestLinkedInProvider_ExchangeCode(t *testing.T) {
	tests := []struct {
		name           string
		serverResponse interface{}
		serverStatus   int
		clientID       string
		clientSecret   string
		code           string
		redirectURI    string
		wantToken      *oauth.TokenResponse
		wantErr        bool
		errContains    string
	}{
		{
			name: "successful token exchange",
			serverResponse: map[string]interface{}{
				"access_token": "linkedin-access-token",
				"expires_in":   5184000,
				"scope":        "profile email openid",
			},
			serverStatus: http.StatusOK,
			clientID:     "test-client",
			clientSecret: "test-secret",
			code:         "auth-code",
			redirectURI:  "https://example.com/callback",
			wantToken: &oauth.TokenResponse{
				AccessToken: "linkedin-access-token",
				ExpiresIn:   5184000,
				Scope:       "profile email openid",
			},
			wantErr: false,
		},
		{
			name: "error response from LinkedIn",
			serverResponse: map[string]interface{}{
				"error":             "invalid_grant",
				"error_description": "Invalid authorization code",
			},
			serverStatus: http.StatusBadRequest,
			clientID:     "test-client",
			clientSecret: "test-secret",
			code:         "invalid-code",
			redirectURI:  "https://example.com/callback",
			wantErr:      true,
			errContains:  "token request failed with status 400",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodPost, r.Method)
				assert.Equal(t, "/oauth/v2/accessToken", r.URL.Path)
				assert.Equal(t, "application/x-www-form-urlencoded", r.Header.Get("Content-Type"))

				err := r.ParseForm()
				require.NoError(t, err)
				assert.Equal(t, tt.clientID, r.FormValue("client_id"))
				assert.Equal(t, tt.clientSecret, r.FormValue("client_secret"))
				assert.Equal(t, tt.code, r.FormValue("code"))
				assert.Equal(t, tt.redirectURI, r.FormValue("redirect_uri"))
				assert.Equal(t, "authorization_code", r.FormValue("grant_type"))

				// LinkedIn doesn't use PKCE, so code_verifier should not be present
				assert.Empty(t, r.FormValue("code_verifier"))

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.serverStatus)
				err = json.NewEncoder(w).Encode(tt.serverResponse)
				require.NoError(t, err)
			}))
			defer server.Close()

			provider := NewLinkedInProvider()
			provider.tokenURL = server.URL + "/oauth/v2/accessToken"

			ctx := context.Background()
			token, err := provider.ExchangeCode(ctx, tt.clientID, tt.clientSecret, tt.code, tt.redirectURI, "ignored-verifier")

			if tt.wantErr {
				require.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
				assert.Nil(t, token)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.wantToken.AccessToken, token.AccessToken)
				assert.Equal(t, tt.wantToken.ExpiresIn, token.ExpiresIn)
				assert.Equal(t, tt.wantToken.Scope, token.Scope)
			}
		})
	}
}

func TestLinkedInProvider_RefreshToken(t *testing.T) {
	provider := NewLinkedInProvider()
	ctx := context.Background()

	// LinkedIn OAuth 2.0 does not support refresh tokens
	_, err := provider.RefreshToken(ctx, "client-id", "client-secret", "refresh-token")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "does not support refresh tokens")
}

func TestLinkedInProvider_GetUserInfo(t *testing.T) {
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
				"sub":   "abc123xyz",
				"name":  "Test User",
				"email": "testuser@example.com",
			},
			serverStatus: http.StatusOK,
			accessToken:  "valid-token",
			wantUserInfo: &oauth.UserInfo{
				ID:    "abc123xyz",
				Name:  "Test User",
				Email: "testuser@example.com",
			},
			wantErr: false,
		},
		{
			name: "error response from LinkedIn",
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
				assert.Equal(t, "/v2/userinfo", r.URL.Path)
				assert.Equal(t, "Bearer "+tt.accessToken, r.Header.Get("Authorization"))

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.serverStatus)
				err := json.NewEncoder(w).Encode(tt.serverResponse)
				require.NoError(t, err)
			}))
			defer server.Close()

			provider := NewLinkedInProvider()
			provider.userInfoURL = server.URL + "/v2/userinfo"

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
				assert.Equal(t, tt.wantUserInfo.Name, userInfo.Name)
				assert.Equal(t, tt.wantUserInfo.Email, userInfo.Email)
			}
		})
	}
}

func TestLinkedInProvider_RevokeToken(t *testing.T) {
	provider := NewLinkedInProvider()
	ctx := context.Background()

	// LinkedIn OAuth 2.0 does not support token revocation
	err := provider.RevokeToken(ctx, "client-id", "client-secret", "token")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "does not support token revocation")
}

func TestLinkedInProvider_URLEncoding(t *testing.T) {
	provider := NewLinkedInProvider()

	authURL := provider.GetAuthURL(
		"client-id",
		"https://example.com/callback?foo=bar",
		"state-with-special-chars-!@#$%",
		[]string{"profile", "email"},
		"unused-challenge",
	)

	parsedURL, err := url.Parse(authURL)
	require.NoError(t, err)

	params := parsedURL.Query()
	assert.Equal(t, "https://example.com/callback?foo=bar", params.Get("redirect_uri"))
	assert.Equal(t, "state-with-special-chars-!@#$%", params.Get("state"))
}

func TestLinkedInProvider_ContextCancellation(t *testing.T) {
	provider := NewLinkedInProvider()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := provider.ExchangeCode(ctx, "id", "secret", "code", "uri", "verifier")
	assert.Error(t, err)
	assert.Contains(t, strings.ToLower(err.Error()), "context")

	_, err = provider.GetUserInfo(ctx, "token")
	assert.Error(t, err)
	assert.Contains(t, strings.ToLower(err.Error()), "context")
}
