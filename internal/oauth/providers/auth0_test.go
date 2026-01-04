package providers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/gorax/gorax/internal/oauth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuth0Provider_Key(t *testing.T) {
	provider := NewAuth0Provider("tenant.auth0.com")
	assert.Equal(t, "auth0", provider.Key())
}

func TestAuth0Provider_Name(t *testing.T) {
	provider := NewAuth0Provider("tenant.auth0.com")
	assert.Equal(t, "Auth0", provider.Name())
}

func TestAuth0Provider_GetAuthURL(t *testing.T) {
	tests := []struct {
		name          string
		domain        string
		clientID      string
		redirectURI   string
		state         string
		scopes        []string
		codeChallenge string
		wantContains  []string
		wantHost      string
	}{
		{
			name:          "basic auth URL with PKCE",
			domain:        "tenant.auth0.com",
			clientID:      "test-client-id",
			redirectURI:   "https://example.com/callback",
			state:         "test-state",
			scopes:        []string{"openid", "profile", "email", "offline_access"},
			codeChallenge: "test-challenge",
			wantContains: []string{
				"https://tenant.auth0.com/authorize",
				"client_id=test-client-id",
				"redirect_uri=https%3A%2F%2Fexample.com%2Fcallback",
				"state=test-state",
				"scope=openid+profile+email+offline_access",
				"code_challenge=test-challenge",
				"code_challenge_method=S256",
				"response_type=code",
			},
			wantHost: "tenant.auth0.com",
		},
		{
			name:          "custom domain",
			domain:        "login.mycompany.com",
			clientID:      "client123",
			redirectURI:   "http://localhost:8080/callback",
			state:         "state123",
			scopes:        []string{"openid", "email"},
			codeChallenge: "challenge123",
			wantContains: []string{
				"https://login.mycompany.com/authorize",
				"scope=openid+email",
			},
			wantHost: "login.mycompany.com",
		},
		{
			name:          "regional tenant",
			domain:        "tenant.us.auth0.com",
			clientID:      "client-abc",
			redirectURI:   "https://app.example.com/callback",
			state:         "state-xyz",
			scopes:        []string{"openid", "profile"},
			codeChallenge: "challenge-xyz",
			wantContains: []string{
				"https://tenant.us.auth0.com/authorize",
			},
			wantHost: "tenant.us.auth0.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := NewAuth0Provider(tt.domain)
			authURL := provider.GetAuthURL(tt.clientID, tt.redirectURI, tt.state, tt.scopes, tt.codeChallenge)

			for _, want := range tt.wantContains {
				assert.Contains(t, authURL, want)
			}

			parsedURL, err := url.Parse(authURL)
			require.NoError(t, err)
			assert.Equal(t, tt.wantHost, parsedURL.Host)
			assert.Equal(t, "/authorize", parsedURL.Path)
		})
	}
}

func TestAuth0Provider_ExchangeCode(t *testing.T) {
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
				"access_token":  "auth0-access-token",
				"token_type":    "Bearer",
				"expires_in":    86400,
				"refresh_token": "auth0-refresh-token",
				"id_token":      "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9...",
				"scope":         "openid profile email offline_access",
			},
			serverStatus: http.StatusOK,
			clientID:     "test-client",
			clientSecret: "test-secret",
			code:         "auth-code",
			redirectURI:  "https://example.com/callback",
			codeVerifier: "verifier123",
			wantToken: &oauth.TokenResponse{
				AccessToken:  "auth0-access-token",
				TokenType:    "Bearer",
				ExpiresIn:    86400,
				RefreshToken: "auth0-refresh-token",
				IDToken:      "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9...",
				Scope:        "openid profile email offline_access",
			},
			wantErr: false,
		},
		{
			name: "successful token exchange without refresh token",
			serverResponse: map[string]interface{}{
				"access_token": "auth0-access-token",
				"token_type":   "Bearer",
				"expires_in":   86400,
				"id_token":     "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9...",
			},
			serverStatus: http.StatusOK,
			clientID:     "test-client",
			clientSecret: "test-secret",
			code:         "auth-code",
			redirectURI:  "https://example.com/callback",
			codeVerifier: "verifier123",
			wantToken: &oauth.TokenResponse{
				AccessToken: "auth0-access-token",
				TokenType:   "Bearer",
				ExpiresIn:   86400,
				IDToken:     "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9...",
			},
			wantErr: false,
		},
		{
			name: "error response from Auth0",
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
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodPost, r.Method)
				assert.Equal(t, "/oauth/token", r.URL.Path)
				assert.Equal(t, "application/x-www-form-urlencoded", r.Header.Get("Content-Type"))

				err := r.ParseForm()
				require.NoError(t, err)
				assert.Equal(t, tt.clientID, r.FormValue("client_id"))
				assert.Equal(t, tt.clientSecret, r.FormValue("client_secret"))
				assert.Equal(t, tt.code, r.FormValue("code"))
				assert.Equal(t, tt.redirectURI, r.FormValue("redirect_uri"))
				assert.Equal(t, "authorization_code", r.FormValue("grant_type"))
				assert.Equal(t, tt.codeVerifier, r.FormValue("code_verifier"))

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.serverStatus)
				err = json.NewEncoder(w).Encode(tt.serverResponse)
				require.NoError(t, err)
			}))
			defer server.Close()

			// Extract host from server URL
			serverURL, err := url.Parse(server.URL)
			require.NoError(t, err)

			provider := NewAuth0Provider(serverURL.Host)
			provider.tokenURL = server.URL + "/oauth/token"

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
				assert.Equal(t, tt.wantToken.IDToken, token.IDToken)
			}
		})
	}
}

func TestAuth0Provider_RefreshToken(t *testing.T) {
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
				"access_token": "new-auth0-access-token",
				"token_type":   "Bearer",
				"expires_in":   86400,
				"id_token":     "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9...",
				"scope":        "openid profile email",
			},
			serverStatus: http.StatusOK,
			clientID:     "test-client",
			clientSecret: "test-secret",
			refreshToken: "old-refresh-token",
			wantToken: &oauth.TokenResponse{
				AccessToken: "new-auth0-access-token",
				TokenType:   "Bearer",
				ExpiresIn:   86400,
				IDToken:     "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9...",
				Scope:       "openid profile email",
			},
			wantErr: false,
		},
		{
			name: "error response from Auth0",
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
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodPost, r.Method)
				assert.Equal(t, "/oauth/token", r.URL.Path)

				err := r.ParseForm()
				require.NoError(t, err)
				assert.Equal(t, tt.clientID, r.FormValue("client_id"))
				assert.Equal(t, tt.clientSecret, r.FormValue("client_secret"))
				assert.Equal(t, tt.refreshToken, r.FormValue("refresh_token"))
				assert.Equal(t, "refresh_token", r.FormValue("grant_type"))

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.serverStatus)
				err = json.NewEncoder(w).Encode(tt.serverResponse)
				require.NoError(t, err)
			}))
			defer server.Close()

			serverURL, err := url.Parse(server.URL)
			require.NoError(t, err)

			provider := NewAuth0Provider(serverURL.Host)
			provider.tokenURL = server.URL + "/oauth/token"

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
			}
		})
	}
}

func TestAuth0Provider_GetUserInfo(t *testing.T) {
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
				"sub":   "auth0|123456789",
				"name":  "Test User",
				"email": "testuser@example.com",
			},
			serverStatus: http.StatusOK,
			accessToken:  "valid-token",
			wantUserInfo: &oauth.UserInfo{
				ID:    "auth0|123456789",
				Name:  "Test User",
				Email: "testuser@example.com",
			},
			wantErr: false,
		},
		{
			name: "successful user info with nickname",
			serverResponse: map[string]interface{}{
				"sub":      "google-oauth2|123456",
				"nickname": "testuser",
				"name":     "Test User",
				"email":    "testuser@example.com",
			},
			serverStatus: http.StatusOK,
			accessToken:  "valid-token",
			wantUserInfo: &oauth.UserInfo{
				ID:       "google-oauth2|123456",
				Username: "testuser",
				Name:     "Test User",
				Email:    "testuser@example.com",
			},
			wantErr: false,
		},
		{
			name: "error response from Auth0",
			serverResponse: map[string]interface{}{
				"error": "invalid_token",
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
				assert.Equal(t, "/userinfo", r.URL.Path)
				assert.Equal(t, "Bearer "+tt.accessToken, r.Header.Get("Authorization"))

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.serverStatus)
				err := json.NewEncoder(w).Encode(tt.serverResponse)
				require.NoError(t, err)
			}))
			defer server.Close()

			serverURL, err := url.Parse(server.URL)
			require.NoError(t, err)

			provider := NewAuth0Provider(serverURL.Host)
			provider.userInfoURL = server.URL + "/userinfo"

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
				assert.Equal(t, tt.wantUserInfo.Email, userInfo.Email)
			}
		})
	}
}

func TestAuth0Provider_RevokeToken(t *testing.T) {
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
			name:         "error response from Auth0",
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
				assert.Equal(t, "/oauth/revoke", r.URL.Path)

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

			serverURL, err := url.Parse(server.URL)
			require.NoError(t, err)

			provider := NewAuth0Provider(serverURL.Host)
			provider.revokeURL = server.URL + "/oauth/revoke"

			ctx := context.Background()
			err = provider.RevokeToken(ctx, "client-id", "client-secret", tt.accessToken)

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

func TestAuth0Provider_DomainVariations(t *testing.T) {
	tests := []struct {
		name       string
		domain     string
		wantInAuth string
	}{
		{
			name:       "standard tenant",
			domain:     "tenant.auth0.com",
			wantInAuth: "https://tenant.auth0.com",
		},
		{
			name:       "regional tenant",
			domain:     "tenant.us.auth0.com",
			wantInAuth: "https://tenant.us.auth0.com",
		},
		{
			name:       "custom domain",
			domain:     "login.mycompany.com",
			wantInAuth: "https://login.mycompany.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := NewAuth0Provider(tt.domain)
			authURL := provider.GetAuthURL("client", "redirect", "state", []string{"openid"}, "challenge")
			assert.Contains(t, authURL, tt.wantInAuth)
		})
	}
}

func TestAuth0Provider_URLEncoding(t *testing.T) {
	provider := NewAuth0Provider("tenant.auth0.com")

	authURL := provider.GetAuthURL(
		"client-id",
		"https://example.com/callback?foo=bar",
		"state-with-special-chars-!@#$%",
		[]string{"openid", "profile"},
		"challenge123",
	)

	parsedURL, err := url.Parse(authURL)
	require.NoError(t, err)

	params := parsedURL.Query()
	assert.Equal(t, "https://example.com/callback?foo=bar", params.Get("redirect_uri"))
	assert.Equal(t, "state-with-special-chars-!@#$%", params.Get("state"))
}

func TestAuth0Provider_ContextCancellation(t *testing.T) {
	provider := NewAuth0Provider("tenant.auth0.com")

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

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
