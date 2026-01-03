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

func TestSalesforceProvider_Key(t *testing.T) {
	provider := NewSalesforceProvider(false)
	assert.Equal(t, "salesforce", provider.Key())
}

func TestSalesforceProvider_Name(t *testing.T) {
	provider := NewSalesforceProvider(false)
	assert.Equal(t, "Salesforce", provider.Name())
}

func TestSalesforceProvider_GetAuthURL(t *testing.T) {
	tests := []struct {
		name          string
		isSandbox     bool
		clientID      string
		redirectURI   string
		state         string
		scopes        []string
		codeChallenge string
		wantContains  []string
		wantHost      string
	}{
		{
			name:          "production with PKCE",
			isSandbox:     false,
			clientID:      "test-client-id",
			redirectURI:   "https://example.com/callback",
			state:         "test-state",
			scopes:        []string{"api", "refresh_token", "openid"},
			codeChallenge: "test-challenge",
			wantContains: []string{
				"https://login.salesforce.com/services/oauth2/authorize",
				"client_id=test-client-id",
				"redirect_uri=https%3A%2F%2Fexample.com%2Fcallback",
				"state=test-state",
				"scope=api+refresh_token+openid",
				"code_challenge=test-challenge",
				"code_challenge_method=S256",
				"response_type=code",
			},
			wantHost: "login.salesforce.com",
		},
		{
			name:          "sandbox environment",
			isSandbox:     true,
			clientID:      "test-client-id",
			redirectURI:   "https://example.com/callback",
			state:         "test-state",
			scopes:        []string{"api", "refresh_token"},
			codeChallenge: "test-challenge",
			wantContains: []string{
				"https://test.salesforce.com/services/oauth2/authorize",
				"client_id=test-client-id",
			},
			wantHost: "test.salesforce.com",
		},
		{
			name:          "with full scope list",
			isSandbox:     false,
			clientID:      "client123",
			redirectURI:   "http://localhost:8080/callback",
			state:         "state123",
			scopes:        []string{"api", "refresh_token", "openid", "profile", "email"},
			codeChallenge: "challenge123",
			wantContains: []string{
				"scope=api+refresh_token+openid+profile+email",
			},
			wantHost: "login.salesforce.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := NewSalesforceProvider(tt.isSandbox)
			authURL := provider.GetAuthURL(tt.clientID, tt.redirectURI, tt.state, tt.scopes, tt.codeChallenge)

			for _, want := range tt.wantContains {
				assert.Contains(t, authURL, want)
			}

			parsedURL, err := url.Parse(authURL)
			require.NoError(t, err)
			assert.Equal(t, tt.wantHost, parsedURL.Host)
			assert.Equal(t, "/services/oauth2/authorize", parsedURL.Path)
		})
	}
}

func TestSalesforceProvider_ExchangeCode(t *testing.T) {
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
				"access_token":  "salesforce-access-token",
				"token_type":    "Bearer",
				"refresh_token": "salesforce-refresh-token",
				"id_token":      "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9...",
				"scope":         "api refresh_token openid profile email",
				"instance_url":  "https://myorg.my.salesforce.com",
				"id":            "https://login.salesforce.com/id/00Dxx0000001gPLEAY/005xx000001SwiUAAS",
			},
			serverStatus: http.StatusOK,
			clientID:     "test-client",
			clientSecret: "test-secret",
			code:         "auth-code",
			redirectURI:  "https://example.com/callback",
			codeVerifier: "verifier123",
			wantToken: &oauth.TokenResponse{
				AccessToken:  "salesforce-access-token",
				TokenType:    "Bearer",
				RefreshToken: "salesforce-refresh-token",
				IDToken:      "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9...",
				Scope:        "api refresh_token openid profile email",
			},
			wantErr: false,
		},
		{
			name: "error response from Salesforce",
			serverResponse: map[string]interface{}{
				"error":             "invalid_grant",
				"error_description": "authentication failure",
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
				assert.Equal(t, "/services/oauth2/token", r.URL.Path)
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

			provider := NewSalesforceProvider(false)
			provider.tokenURL = server.URL + "/services/oauth2/token"

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
				assert.Equal(t, tt.wantToken.RefreshToken, token.RefreshToken)
				assert.Equal(t, tt.wantToken.IDToken, token.IDToken)
				assert.Equal(t, tt.wantToken.Scope, token.Scope)
			}
		})
	}
}

func TestSalesforceProvider_RefreshToken(t *testing.T) {
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
				"access_token": "new-salesforce-access-token",
				"token_type":   "Bearer",
				"id_token":     "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9...",
				"scope":        "api refresh_token openid",
				"instance_url": "https://myorg.my.salesforce.com",
				"id":           "https://login.salesforce.com/id/00Dxx0000001gPLEAY/005xx000001SwiUAAS",
			},
			serverStatus: http.StatusOK,
			clientID:     "test-client",
			clientSecret: "test-secret",
			refreshToken: "old-refresh-token",
			wantToken: &oauth.TokenResponse{
				AccessToken: "new-salesforce-access-token",
				TokenType:   "Bearer",
				IDToken:     "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9...",
				Scope:       "api refresh_token openid",
			},
			wantErr: false,
		},
		{
			name: "error response from Salesforce",
			serverResponse: map[string]interface{}{
				"error":             "invalid_grant",
				"error_description": "expired access/refresh token",
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
				assert.Equal(t, "/services/oauth2/token", r.URL.Path)

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

			provider := NewSalesforceProvider(false)
			provider.tokenURL = server.URL + "/services/oauth2/token"

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
			}
		})
	}
}

func TestSalesforceProvider_GetUserInfo(t *testing.T) {
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
				"sub":                "https://login.salesforce.com/id/00Dxx0000001gPLEAY/005xx000001SwiUAAS",
				"user_id":            "005xx000001SwiUAAS",
				"organization_id":    "00Dxx0000001gPLEAY",
				"name":               "Test User",
				"email":              "testuser@example.com",
				"email_verified":     true,
				"preferred_username": "testuser@example.com",
			},
			serverStatus: http.StatusOK,
			accessToken:  "valid-token",
			wantUserInfo: &oauth.UserInfo{
				ID:       "005xx000001SwiUAAS",
				Username: "testuser@example.com",
				Email:    "testuser@example.com",
				Name:     "Test User",
			},
			wantErr: false,
		},
		{
			name: "error response from Salesforce",
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
				assert.Equal(t, "/services/oauth2/userinfo", r.URL.Path)
				assert.Equal(t, "Bearer "+tt.accessToken, r.Header.Get("Authorization"))

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.serverStatus)
				err := json.NewEncoder(w).Encode(tt.serverResponse)
				require.NoError(t, err)
			}))
			defer server.Close()

			provider := NewSalesforceProvider(false)
			provider.userInfoURL = server.URL + "/services/oauth2/userinfo"

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
				assert.Equal(t, tt.wantUserInfo.Email, userInfo.Email)
				assert.Equal(t, tt.wantUserInfo.Name, userInfo.Name)
			}
		})
	}
}

func TestSalesforceProvider_RevokeToken(t *testing.T) {
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
			name:         "error response from Salesforce",
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
				assert.Equal(t, "/services/oauth2/revoke", r.URL.Path)

				err := r.ParseForm()
				require.NoError(t, err)
				assert.Equal(t, tt.accessToken, r.FormValue("token"))

				w.WriteHeader(tt.serverStatus)
			}))
			defer server.Close()

			provider := NewSalesforceProvider(false)
			provider.revokeURL = server.URL + "/services/oauth2/revoke"

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

func TestSalesforceProvider_SandboxVsProduction(t *testing.T) {
	// Test production URLs
	prodProvider := NewSalesforceProvider(false)
	prodAuthURL := prodProvider.GetAuthURL("client", "redirect", "state", []string{"api"}, "challenge")
	assert.Contains(t, prodAuthURL, "https://login.salesforce.com")

	// Test sandbox URLs
	sandboxProvider := NewSalesforceProvider(true)
	sandboxAuthURL := sandboxProvider.GetAuthURL("client", "redirect", "state", []string{"api"}, "challenge")
	assert.Contains(t, sandboxAuthURL, "https://test.salesforce.com")
}

func TestSalesforceProvider_URLEncoding(t *testing.T) {
	provider := NewSalesforceProvider(false)

	authURL := provider.GetAuthURL(
		"client-id",
		"https://example.com/callback?foo=bar",
		"state-with-special-chars-!@#$%",
		[]string{"api", "refresh_token"},
		"challenge123",
	)

	parsedURL, err := url.Parse(authURL)
	require.NoError(t, err)

	params := parsedURL.Query()
	assert.Equal(t, "https://example.com/callback?foo=bar", params.Get("redirect_uri"))
	assert.Equal(t, "state-with-special-chars-!@#$%", params.Get("state"))
}

func TestSalesforceProvider_ContextCancellation(t *testing.T) {
	provider := NewSalesforceProvider(false)

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
