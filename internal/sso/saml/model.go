package saml

import "encoding/json"

// Config represents SAML provider configuration
type Config struct {
	EntityID          string            `json:"entity_id"`
	ACSURL            string            `json:"acs_url"`
	IdPMetadataURL    string            `json:"idp_metadata_url,omitempty"`
	IdPMetadataXML    string            `json:"idp_metadata_xml,omitempty"`
	IdPEntityID       string            `json:"idp_entity_id"`
	IdPSSOURL         string            `json:"idp_sso_url"`
	Certificate       string            `json:"certificate,omitempty"`
	PrivateKey        string            `json:"private_key,omitempty"`
	AttributeMapping  map[string]string `json:"attribute_mapping"`
	SignAuthnRequests bool              `json:"sign_authn_requests"`
}

// ParseConfig parses raw JSON into SAML config
func ParseConfig(raw json.RawMessage) (*Config, error) {
	var config Config
	if err := json.Unmarshal(raw, &config); err != nil {
		return nil, err
	}
	return &config, nil
}
