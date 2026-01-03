package oidc

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/gorax/gorax/internal/sso"
	"golang.org/x/oauth2"
)

// Provider implements the OIDC SSO provider
type Provider struct {
	provider     *sso.Provider
	config       *sso.OIDCConfig
	oidcProvider *oidc.Provider
	oauth2Config *oauth2.Config
	verifier     *oidc.IDTokenVerifier
	stateStore   map[string]string // In production, use Redis/database
}

// NewProvider creates a new OIDC provider
func NewProvider(ctx context.Context, provider *sso.Provider) (*Provider, error) {
	if provider.Type != sso.ProviderTypeOIDC {
		return nil, fmt.Errorf("invalid provider type: expected oidc, got %s", provider.Type)
	}

	var config sso.OIDCConfig
	if err := json.Unmarshal(provider.Config, &config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal OIDC config: %w", err)
	}

	p := &Provider{
		provider:   provider,
		config:     &config,
		stateStore: make(map[string]string),
	}

	if err := p.initProvider(ctx); err != nil {
		return nil, fmt.Errorf("failed to initialize OIDC provider: %w", err)
	}

	return p, nil
}

// initProvider initializes the OIDC provider and OAuth2 config
func (p *Provider) initProvider(ctx context.Context) error {
	// Initialize OIDC provider
	oidcProvider, err := oidc.NewProvider(ctx, p.config.DiscoveryURL)
	if err != nil {
		return fmt.Errorf("failed to create OIDC provider: %w", err)
	}
	p.oidcProvider = oidcProvider

	// Set up OAuth2 config
	scopes := p.config.Scopes
	if len(scopes) == 0 {
		scopes = []string{oidc.ScopeOpenID, "profile", "email"}
	}

	p.oauth2Config = &oauth2.Config{
		ClientID:     p.config.ClientID,
		ClientSecret: p.config.ClientSecret,
		RedirectURL:  p.config.RedirectURL,
		Endpoint:     oidcProvider.Endpoint(),
		Scopes:       scopes,
	}

	// Set up ID token verifier
	p.verifier = oidcProvider.Verifier(&oidc.Config{
		ClientID: p.config.ClientID,
	})

	return nil
}

// GetType returns the provider type
func (p *Provider) GetType() sso.ProviderType {
	return sso.ProviderTypeOIDC
}

// InitiateLogin generates the OIDC authorization URL
func (p *Provider) InitiateLogin(ctx context.Context, relayState string) (string, error) {
	// Generate state token for CSRF protection
	state, err := generateState()
	if err != nil {
		return "", fmt.Errorf("failed to generate state: %w", err)
	}

	// Store state with relay state (in production, use Redis with TTL)
	p.stateStore[state] = relayState

	// Generate authorization URL
	authURL := p.oauth2Config.AuthCodeURL(state)

	return authURL, nil
}

// HandleCallback processes the OIDC callback and extracts user attributes
func (p *Provider) HandleCallback(ctx context.Context, r *http.Request) (*sso.UserAttributes, error) {
	// Verify state parameter
	state := r.URL.Query().Get("state")
	if state == "" {
		return nil, fmt.Errorf("state parameter missing")
	}

	// Validate state (CSRF protection)
	if _, ok := p.stateStore[state]; !ok {
		return nil, fmt.Errorf("invalid state parameter")
	}
	delete(p.stateStore, state) // Remove used state

	// Check for error response
	if errResp := r.URL.Query().Get("error"); errResp != "" {
		errDesc := r.URL.Query().Get("error_description")
		return nil, fmt.Errorf("OIDC error: %s - %s", errResp, errDesc)
	}

	// Get authorization code
	code := r.URL.Query().Get("code")
	if code == "" {
		return nil, fmt.Errorf("authorization code missing")
	}

	// Exchange code for token
	oauth2Token, err := p.oauth2Config.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange code for token: %w", err)
	}

	// Extract ID token
	rawIDToken, ok := oauth2Token.Extra("id_token").(string)
	if !ok {
		return nil, fmt.Errorf("id_token not found in token response")
	}

	// Verify ID token
	idToken, err := p.verifier.Verify(ctx, rawIDToken)
	if err != nil {
		return nil, fmt.Errorf("failed to verify ID token: %w", err)
	}

	// Extract claims
	var claims map[string]interface{}
	if err := idToken.Claims(&claims); err != nil {
		return nil, fmt.Errorf("failed to parse claims: %w", err)
	}

	// Get additional user info if configured
	if p.config.UserinfoURL != "" {
		userInfo, err := p.getUserInfo(ctx, oauth2Token)
		if err != nil {
			// Log warning but don't fail - use ID token claims
			fmt.Printf("failed to get user info: %v\n", err)
		} else {
			// Merge userinfo into claims
			for k, v := range userInfo {
				if _, exists := claims[k]; !exists {
					claims[k] = v
				}
			}
		}
	}

	// Extract user attributes
	userAttrs, err := p.extractUserAttributes(claims)
	if err != nil {
		return nil, fmt.Errorf("failed to extract user attributes: %w", err)
	}

	return userAttrs, nil
}

// getUserInfo fetches additional user information from userinfo endpoint
func (p *Provider) getUserInfo(ctx context.Context, token *oauth2.Token) (map[string]interface{}, error) {
	userInfo, err := p.oidcProvider.UserInfo(ctx, oauth2.StaticTokenSource(token))
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}

	var claims map[string]interface{}
	if err := userInfo.Claims(&claims); err != nil {
		return nil, fmt.Errorf("failed to parse user info claims: %w", err)
	}

	return claims, nil
}

// extractUserAttributes extracts user attributes from OIDC claims
func (p *Provider) extractUserAttributes(claims map[string]interface{}) (*sso.UserAttributes, error) {
	attrs := &sso.UserAttributes{
		Attributes: make(map[string]string),
	}

	// Store all claims as strings
	for key, val := range claims {
		if strVal, ok := val.(string); ok {
			attrs.Attributes[key] = strVal
		}
	}

	// Extract subject as external ID
	if sub, ok := claims["sub"].(string); ok {
		attrs.ExternalID = sub
	} else {
		return nil, fmt.Errorf("sub claim not found")
	}

	// Map standard OIDC claims
	if email, ok := claims["email"].(string); ok {
		attrs.Email = email
	}

	if givenName, ok := claims["given_name"].(string); ok {
		attrs.FirstName = givenName
	}

	if familyName, ok := claims["family_name"].(string); ok {
		attrs.LastName = familyName
	}

	// Handle groups claim (can be string or array)
	if groups, ok := claims["groups"]; ok {
		switch v := groups.(type) {
		case []interface{}:
			for _, group := range v {
				if groupStr, ok := group.(string); ok {
					attrs.Groups = append(attrs.Groups, groupStr)
				}
			}
		case []string:
			attrs.Groups = v
		case string:
			attrs.Groups = []string{v}
		}
	}

	// Apply custom attribute mapping
	for claimName, mappedName := range p.config.AttributeMapping {
		if val, ok := claims[claimName]; ok {
			strVal := fmt.Sprintf("%v", val)
			switch mappedName {
			case "email":
				attrs.Email = strVal
			case "first_name":
				attrs.FirstName = strVal
			case "last_name":
				attrs.LastName = strVal
			case "groups":
				if groupVal, ok := val.([]interface{}); ok {
					attrs.Groups = nil
					for _, g := range groupVal {
						attrs.Groups = append(attrs.Groups, fmt.Sprintf("%v", g))
					}
				}
			}
		}
	}

	if attrs.Email == "" {
		return nil, fmt.Errorf("email claim not found")
	}

	return attrs, nil
}

// GetMetadata returns empty string (OIDC doesn't have SP metadata like SAML)
func (p *Provider) GetMetadata(ctx context.Context) (string, error) {
	// OIDC discovery is handled by the IdP's .well-known endpoint
	return "", nil
}

// Validate validates the OIDC provider configuration
func (p *Provider) Validate(ctx context.Context) error {
	if p.config.ClientID == "" {
		return fmt.Errorf("client ID is required")
	}

	if p.config.ClientSecret == "" {
		return fmt.Errorf("client secret is required")
	}

	if p.config.DiscoveryURL == "" {
		return fmt.Errorf("discovery URL is required")
	}

	if p.config.RedirectURL == "" {
		return fmt.Errorf("redirect URL is required")
	}

	// Test discovery endpoint
	_, err := oidc.NewProvider(ctx, p.config.DiscoveryURL)
	if err != nil {
		return fmt.Errorf("failed to fetch OIDC discovery: %w", err)
	}

	return nil
}

// generateState generates a random state token for CSRF protection
func generateState() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	state := base64.URLEncoding.EncodeToString(b)
	return strings.TrimRight(state, "="), nil
}
