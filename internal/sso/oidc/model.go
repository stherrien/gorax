package oidc

import "encoding/json"

// Config represents OIDC provider configuration
type Config struct {
	ClientID         string            `json:"client_id"`
	ClientSecret     string            `json:"client_secret"`
	DiscoveryURL     string            `json:"discovery_url"`
	AuthorizationURL string            `json:"authorization_url,omitempty"`
	TokenURL         string            `json:"token_url,omitempty"`
	UserinfoURL      string            `json:"userinfo_url,omitempty"`
	JWKSURL          string            `json:"jwks_url,omitempty"`
	RedirectURL      string            `json:"redirect_url"`
	Scopes           []string          `json:"scopes"`
	AttributeMapping map[string]string `json:"attribute_mapping"`
}

// ParseConfig parses raw JSON into OIDC config
func ParseConfig(raw json.RawMessage) (*Config, error) {
	var config Config
	if err := json.Unmarshal(raw, &config); err != nil {
		return nil, err
	}
	return &config, nil
}
