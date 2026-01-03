package mocks

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync"
	"time"
)

// MockOAuthProvider is a mock OAuth 2.0 identity provider for testing
type MockOAuthProvider struct {
	server       *httptest.Server
	providerKey  string
	clientID     string
	clientSecret string
	users        map[string]*MockUser
	authCodes    map[string]*AuthCodeData
	tokens       map[string]*TokenData
	mu           sync.RWMutex
}

// MockUser represents a mock user in the IdP
type MockUser struct {
	ID       string `json:"id"`
	Email    string `json:"email"`
	Name     string `json:"name"`
	Username string `json:"username"`
	Avatar   string `json:"avatar_url,omitempty"`
}

// AuthCodeData stores auth code information
type AuthCodeData struct {
	Code         string
	UserID       string
	Scopes       []string
	RedirectURI  string
	CreatedAt    time.Time
	CodeVerifier string // For PKCE
}

// TokenData stores token information
type TokenData struct {
	AccessToken  string
	RefreshToken string
	TokenType    string
	ExpiresIn    int
	UserID       string
	Scopes       []string
	CreatedAt    time.Time
}

// NewMockOAuthProvider creates a new mock OAuth provider
func NewMockOAuthProvider(providerKey, clientID, clientSecret string) *MockOAuthProvider {
	mock := &MockOAuthProvider{
		providerKey:  providerKey,
		clientID:     clientID,
		clientSecret: clientSecret,
		users:        make(map[string]*MockUser),
		authCodes:    make(map[string]*AuthCodeData),
		tokens:       make(map[string]*TokenData),
	}

	// Set up HTTP server with routes
	mux := http.NewServeMux()
	mux.HandleFunc("/oauth/authorize", mock.handleAuthorize)
	mux.HandleFunc("/oauth/token", mock.handleToken)
	mux.HandleFunc("/oauth/userinfo", mock.handleUserInfo)
	mux.HandleFunc("/.well-known/openid-configuration", mock.handleDiscovery)

	mock.server = httptest.NewServer(mux)

	return mock
}

// Close shuts down the mock server
func (m *MockOAuthProvider) Close() {
	m.server.Close()
}

// BaseURL returns the base URL of the mock provider
func (m *MockOAuthProvider) BaseURL() string {
	return m.server.URL
}

// AuthURL returns the authorization URL
func (m *MockOAuthProvider) AuthURL() string {
	return m.server.URL + "/oauth/authorize"
}

// TokenURL returns the token URL
func (m *MockOAuthProvider) TokenURL() string {
	return m.server.URL + "/oauth/token"
}

// UserInfoURL returns the user info URL
func (m *MockOAuthProvider) UserInfoURL() string {
	return m.server.URL + "/oauth/userinfo"
}

// AddUser adds a mock user to the provider
func (m *MockOAuthProvider) AddUser(user *MockUser) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.users[user.ID] = user
}

// handleAuthorize handles OAuth authorization requests
func (m *MockOAuthProvider) handleAuthorize(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()

	clientID := query.Get("client_id")
	redirectURI := query.Get("redirect_uri")
	state := query.Get("state")
	scope := query.Get("scope")
	codeChallenge := query.Get("code_challenge")

	// Validate client_id
	if clientID != m.clientID {
		http.Error(w, "invalid client_id", http.StatusBadRequest)
		return
	}

	if redirectURI == "" {
		http.Error(w, "redirect_uri required", http.StatusBadRequest)
		return
	}

	// Get first user (for testing, just use first user)
	m.mu.RLock()
	var user *MockUser
	for _, u := range m.users {
		user = u
		break
	}
	m.mu.RUnlock()

	if user == nil {
		http.Error(w, "no mock users available", http.StatusInternalServerError)
		return
	}

	// Generate authorization code
	code := generateRandomString(32)
	scopes := parseScopes(scope)

	authCodeData := &AuthCodeData{
		Code:        code,
		UserID:      user.ID,
		Scopes:      scopes,
		RedirectURI: redirectURI,
		CreatedAt:   time.Now(),
	}

	// Store PKCE verifier if provided
	if codeChallenge != "" {
		authCodeData.CodeVerifier = codeChallenge
	}

	m.mu.Lock()
	m.authCodes[code] = authCodeData
	m.mu.Unlock()

	// Redirect back to redirect_uri with code and state
	redirectURL, _ := url.Parse(redirectURI)
	q := redirectURL.Query()
	q.Set("code", code)
	if state != "" {
		q.Set("state", state)
	}
	redirectURL.RawQuery = q.Encode()

	http.Redirect(w, r, redirectURL.String(), http.StatusFound)
}

// handleToken handles OAuth token exchange requests
func (m *MockOAuthProvider) handleToken(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "failed to parse form", http.StatusBadRequest)
		return
	}

	grantType := r.Form.Get("grant_type")
	clientID := r.Form.Get("client_id")
	clientSecret := r.Form.Get("client_secret")

	// Validate client credentials
	if clientID != m.clientID || clientSecret != m.clientSecret {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{
			"error":             "invalid_client",
			"error_description": "invalid client credentials",
		})
		return
	}

	switch grantType {
	case "authorization_code":
		m.handleAuthorizationCodeGrant(w, r)
	case "refresh_token":
		m.handleRefreshTokenGrant(w, r)
	default:
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"error":             "unsupported_grant_type",
			"error_description": fmt.Sprintf("grant type %s not supported", grantType),
		})
	}
}

// handleAuthorizationCodeGrant handles authorization code grant
func (m *MockOAuthProvider) handleAuthorizationCodeGrant(w http.ResponseWriter, r *http.Request) {
	code := r.Form.Get("code")
	redirectURI := r.Form.Get("redirect_uri")

	m.mu.RLock()
	authCodeData, exists := m.authCodes[code]
	m.mu.RUnlock()

	if !exists {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"error":             "invalid_grant",
			"error_description": "invalid authorization code",
		})
		return
	}

	// Validate redirect URI
	if authCodeData.RedirectURI != redirectURI {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"error":             "invalid_grant",
			"error_description": "redirect_uri mismatch",
		})
		return
	}

	// Check code expiration (10 minutes)
	if time.Since(authCodeData.CreatedAt) > 10*time.Minute {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"error":             "invalid_grant",
			"error_description": "authorization code expired",
		})
		return
	}

	// Generate tokens
	accessToken := generateRandomString(64)
	refreshToken := generateRandomString(64)

	tokenData := &TokenData{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    3600,
		UserID:       authCodeData.UserID,
		Scopes:       authCodeData.Scopes,
		CreatedAt:    time.Now(),
	}

	m.mu.Lock()
	m.tokens[accessToken] = tokenData
	delete(m.authCodes, code) // Code is single-use
	m.mu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
		"token_type":    "Bearer",
		"expires_in":    3600,
		"scope":         joinScopes(authCodeData.Scopes),
	})
}

// handleRefreshTokenGrant handles refresh token grant
func (m *MockOAuthProvider) handleRefreshTokenGrant(w http.ResponseWriter, r *http.Request) {
	refreshToken := r.Form.Get("refresh_token")

	// Find token by refresh token
	m.mu.RLock()
	var oldTokenData *TokenData
	for _, tokenData := range m.tokens {
		if tokenData.RefreshToken == refreshToken {
			oldTokenData = tokenData
			break
		}
	}
	m.mu.RUnlock()

	if oldTokenData == nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"error":             "invalid_grant",
			"error_description": "invalid refresh token",
		})
		return
	}

	// Generate new access token
	newAccessToken := generateRandomString(64)

	newTokenData := &TokenData{
		AccessToken:  newAccessToken,
		RefreshToken: refreshToken, // Reuse refresh token
		TokenType:    "Bearer",
		ExpiresIn:    3600,
		UserID:       oldTokenData.UserID,
		Scopes:       oldTokenData.Scopes,
		CreatedAt:    time.Now(),
	}

	m.mu.Lock()
	m.tokens[newAccessToken] = newTokenData
	m.mu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"access_token":  newAccessToken,
		"refresh_token": refreshToken,
		"token_type":    "Bearer",
		"expires_in":    3600,
		"scope":         joinScopes(oldTokenData.Scopes),
	})
}

// handleUserInfo handles user info requests
func (m *MockOAuthProvider) handleUserInfo(w http.ResponseWriter, r *http.Request) {
	// Extract access token from Authorization header
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "unauthorized",
		})
		return
	}

	// Parse Bearer token
	var accessToken string
	if _, err := fmt.Sscanf(authHeader, "Bearer %s", &accessToken); err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "invalid_token",
		})
		return
	}

	// Find token
	m.mu.RLock()
	tokenData, exists := m.tokens[accessToken]
	m.mu.RUnlock()

	if !exists {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "invalid_token",
		})
		return
	}

	// Get user
	m.mu.RLock()
	user, exists := m.users[tokenData.UserID]
	m.mu.RUnlock()

	if !exists {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "user_not_found",
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

// handleDiscovery handles OIDC discovery
func (m *MockOAuthProvider) handleDiscovery(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"issuer":                                m.server.URL,
		"authorization_endpoint":                m.server.URL + "/oauth/authorize",
		"token_endpoint":                        m.server.URL + "/oauth/token",
		"userinfo_endpoint":                     m.server.URL + "/oauth/userinfo",
		"jwks_uri":                              m.server.URL + "/oauth/jwks",
		"response_types_supported":              []string{"code"},
		"subject_types_supported":               []string{"public"},
		"id_token_signing_alg_values_supported": []string{"RS256"},
		"scopes_supported":                      []string{"openid", "profile", "email"},
		"token_endpoint_auth_methods_supported": []string{"client_secret_basic", "client_secret_post"},
	})
}

// Helper functions

func generateRandomString(length int) string {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		panic(err)
	}
	return base64.URLEncoding.EncodeToString(bytes)[:length]
}

func parseScopes(scopeStr string) []string {
	if scopeStr == "" {
		return []string{}
	}
	return splitAndTrim(scopeStr, " ")
}

func joinScopes(scopes []string) string {
	return joinStrings(scopes, " ")
}

func splitAndTrim(s, sep string) []string {
	parts := splitString(s, sep)
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := trimSpace(part)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

func splitString(s, sep string) []string {
	if s == "" {
		return []string{}
	}
	// Simple split implementation
	var result []string
	start := 0
	for i := 0; i < len(s); i++ {
		if i+len(sep) <= len(s) && s[i:i+len(sep)] == sep {
			result = append(result, s[start:i])
			start = i + len(sep)
			i += len(sep) - 1
		}
	}
	result = append(result, s[start:])
	return result
}

func trimSpace(s string) string {
	start := 0
	end := len(s)
	for start < end && (s[start] == ' ' || s[start] == '\t' || s[start] == '\n' || s[start] == '\r') {
		start++
	}
	for end > start && (s[end-1] == ' ' || s[end-1] == '\t' || s[end-1] == '\n' || s[end-1] == '\r') {
		end--
	}
	return s[start:end]
}

func joinStrings(strs []string, sep string) string {
	if len(strs) == 0 {
		return ""
	}
	result := strs[0]
	for i := 1; i < len(strs); i++ {
		result += sep + strs[i]
	}
	return result
}
