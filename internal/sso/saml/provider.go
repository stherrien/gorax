package saml

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"encoding/xml"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/crewjam/saml"
	"github.com/crewjam/saml/samlsp"
)

// ProviderType represents SAML provider type
type ProviderType string

const (
	ProviderTypeSAML ProviderType = "saml"
)

// UserAttributes represents user attributes extracted from SSO provider
type UserAttributes struct {
	ExternalID string            `json:"external_id"`
	Email      string            `json:"email"`
	FirstName  string            `json:"first_name,omitempty"`
	LastName   string            `json:"last_name,omitempty"`
	Groups     []string          `json:"groups,omitempty"`
	Attributes map[string]string `json:"attributes,omitempty"`
}

// Provider implements the SAML 2.0 SSO provider
type Provider struct {
	entityID   string
	config     *Config
	sp         *saml.ServiceProvider
}

// NewProvider creates a new SAML provider
func NewProvider(ctx context.Context, provider *sso.Provider) (*Provider, error) {
	if provider.Type != sso.ProviderTypeSAML {
		return nil, fmt.Errorf("invalid provider type: expected saml, got %s", provider.Type)
	}

	var config sso.SAMLConfig
	if err := json.Unmarshal(provider.Config, &config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal SAML config: %w", err)
	}

	p := &Provider{
		provider: provider,
		config:   &config,
	}

	if err := p.initServiceProvider(ctx); err != nil {
		return nil, fmt.Errorf("failed to initialize service provider: %w", err)
	}

	return p, nil
}

// initServiceProvider initializes the SAML service provider
func (p *Provider) initServiceProvider(ctx context.Context) error {
	acsURL, err := url.Parse(p.config.ACSURL)
	if err != nil {
		return fmt.Errorf("invalid ACS URL: %w", err)
	}

	entityID, err := url.Parse(p.config.EntityID)
	if err != nil {
		return fmt.Errorf("invalid entity ID: %w", err)
	}

	sp := &saml.ServiceProvider{
		EntityID:          entityID.String(),
		AcsURL:            *acsURL,
		MetadataURL:       *entityID,
		AllowIDPInitiated: true,
	}

	// Parse IdP metadata
	if p.config.IdPMetadataXML != "" {
		idpMetadata, err := samlsp.ParseMetadata([]byte(p.config.IdPMetadataXML))
		if err != nil {
			return fmt.Errorf("failed to parse IdP metadata: %w", err)
		}
		sp.IDPMetadata = idpMetadata
	}

	// Set up certificate and private key for request signing
	if p.config.Certificate != "" && p.config.PrivateKey != "" {
		cert, key, err := p.parseCertificateAndKey()
		if err != nil {
			return fmt.Errorf("failed to parse certificate and key: %w", err)
		}
		sp.Certificate = cert
		sp.Key = key
	}

	p.sp = sp
	return nil
}

// parseCertificateAndKey parses the PEM-encoded certificate and private key
func (p *Provider) parseCertificateAndKey() (*x509.Certificate, *rsa.PrivateKey, error) {
	// Parse certificate
	certBlock, _ := pem.Decode([]byte(p.config.Certificate))
	if certBlock == nil {
		return nil, nil, fmt.Errorf("failed to decode certificate PEM")
	}

	cert, err := x509.ParseCertificate(certBlock.Bytes)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse certificate: %w", err)
	}

	// Parse private key
	keyBlock, _ := pem.Decode([]byte(p.config.PrivateKey))
	if keyBlock == nil {
		return nil, nil, fmt.Errorf("failed to decode private key PEM")
	}

	key, err := x509.ParsePKCS1PrivateKey(keyBlock.Bytes)
	if err != nil {
		// Try PKCS8 format
		keyInterface, err := x509.ParsePKCS8PrivateKey(keyBlock.Bytes)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to parse private key: %w", err)
		}
		var ok bool
		key, ok = keyInterface.(*rsa.PrivateKey)
		if !ok {
			return nil, nil, fmt.Errorf("private key is not RSA key")
		}
	}

	return cert, key, nil
}

// GetType returns the provider type
func (p *Provider) GetType() sso.ProviderType {
	return sso.ProviderTypeSAML
}

// InitiateLogin generates the SAML authentication request URL
func (p *Provider) InitiateLogin(ctx context.Context, relayState string) (string, error) {
	if p.sp.IDPMetadata == nil {
		return "", fmt.Errorf("IdP metadata not configured")
	}

	// Create authentication request
	authReq, err := p.sp.MakeAuthenticationRequest(
		p.sp.GetSSOBindingLocation(saml.HTTPRedirectBinding),
		saml.HTTPRedirectBinding,
		saml.HTTPPostBinding,
	)
	if err != nil {
		return "", fmt.Errorf("failed to create authentication request: %w", err)
	}

	// Build redirect URL
	redirectURL, err := authReq.Redirect(relayState, p.sp)
	if err != nil {
		return "", fmt.Errorf("failed to build redirect URL: %w", err)
	}

	return redirectURL.String(), nil
}

// HandleCallback processes the SAML assertion and extracts user attributes
func (p *Provider) HandleCallback(ctx context.Context, r *http.Request) (*sso.UserAttributes, error) {
	if err := r.ParseForm(); err != nil {
		return nil, fmt.Errorf("failed to parse form: %w", err)
	}

	// Get SAML response from form
	samlResponse := r.PostForm.Get("SAMLResponse")
	if samlResponse == "" {
		return nil, fmt.Errorf("SAMLResponse parameter not found")
	}

	// Decode base64 response
	responseData, err := base64.StdEncoding.DecodeString(samlResponse)
	if err != nil {
		return nil, fmt.Errorf("failed to decode SAML response: %w", err)
	}

	// Parse and validate assertion
	assertion, err := p.validateAssertion(ctx, responseData)
	if err != nil {
		return nil, fmt.Errorf("failed to validate assertion: %w", err)
	}

	// Extract user attributes
	userAttrs, err := p.extractUserAttributes(assertion)
	if err != nil {
		return nil, fmt.Errorf("failed to extract user attributes: %w", err)
	}

	return userAttrs, nil
}

// validateAssertion validates the SAML assertion
func (p *Provider) validateAssertion(ctx context.Context, responseData []byte) (*saml.Assertion, error) {
	var response saml.Response
	if err := xml.Unmarshal(responseData, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal SAML response: %w", err)
	}

	// Verify response signature
	if err := p.sp.ValidateEncodedResponse(&response); err != nil {
		return nil, fmt.Errorf("invalid SAML response signature: %w", err)
	}

	// Get assertion
	if len(response.EncryptedAssertions) > 0 {
		return nil, fmt.Errorf("encrypted assertions not supported")
	}

	if len(response.Assertions) == 0 {
		return nil, fmt.Errorf("no assertions in response")
	}

	assertion := response.Assertions[0]

	// Validate assertion
	now := time.Now()

	// Check NotBefore
	if assertion.Conditions.NotBefore != nil && now.Before(*assertion.Conditions.NotBefore) {
		return nil, fmt.Errorf("assertion not yet valid")
	}

	// Check NotOnOrAfter
	if assertion.Conditions.NotOnOrAfter != nil && now.After(*assertion.Conditions.NotOnOrAfter) {
		return nil, fmt.Errorf("assertion expired")
	}

	// Check audience
	if len(assertion.Conditions.AudienceRestrictions) > 0 {
		validAudience := false
		for _, ar := range assertion.Conditions.AudienceRestrictions {
			for _, audience := range ar.Audience {
				if audience.Value == p.config.EntityID {
					validAudience = true
					break
				}
			}
		}
		if !validAudience {
			return nil, fmt.Errorf("invalid audience")
		}
	}

	return &assertion, nil
}

// extractUserAttributes extracts user attributes from SAML assertion
func (p *Provider) extractUserAttributes(assertion *saml.Assertion) (*sso.UserAttributes, error) {
	attrs := &sso.UserAttributes{
		Attributes: make(map[string]string),
	}

	// Get NameID as external ID
	if assertion.Subject != nil && assertion.Subject.NameID != nil {
		attrs.ExternalID = assertion.Subject.NameID.Value
	}

	// Map attributes
	for _, attrStatement := range assertion.AttributeStatements {
		for _, attr := range attrStatement.Attributes {
			if len(attr.Values) == 0 {
				continue
			}

			attrName := attr.Name
			attrValue := attr.Values[0].Value

			// Store in raw attributes
			attrs.Attributes[attrName] = attrValue

			// Map to standard fields
			if mappedName, ok := p.config.AttributeMapping[attrName]; ok {
				switch mappedName {
				case "email":
					attrs.Email = attrValue
				case "first_name":
					attrs.FirstName = attrValue
				case "last_name":
					attrs.LastName = attrValue
				case "groups":
					// Handle groups (can be multi-valued)
					for _, val := range attr.Values {
						attrs.Groups = append(attrs.Groups, val.Value)
					}
				}
			}
		}
	}

	// Use NameID as email if not mapped
	if attrs.Email == "" && attrs.ExternalID != "" {
		attrs.Email = attrs.ExternalID
	}

	if attrs.Email == "" {
		return nil, fmt.Errorf("email attribute not found in assertion")
	}

	return attrs, nil
}

// GetMetadata returns the SAML service provider metadata
func (p *Provider) GetMetadata(ctx context.Context) (string, error) {
	metadata := p.sp.Metadata()
	xmlData, err := xml.MarshalIndent(metadata, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal metadata: %w", err)
	}

	return xml.Header + string(xmlData), nil
}

// Validate validates the SAML provider configuration
func (p *Provider) Validate(ctx context.Context) error {
	if p.config.EntityID == "" {
		return fmt.Errorf("entity ID is required")
	}

	if p.config.ACSURL == "" {
		return fmt.Errorf("ACS URL is required")
	}

	if p.config.IdPMetadataXML == "" && p.config.IdPMetadataURL == "" {
		return fmt.Errorf("IdP metadata is required")
	}

	if p.config.SignAuthnRequests {
		if p.config.Certificate == "" || p.config.PrivateKey == "" {
			return fmt.Errorf("certificate and private key required for signing")
		}
	}

	// Validate attribute mapping
	if len(p.config.AttributeMapping) == 0 {
		return fmt.Errorf("attribute mapping is required")
	}

	return nil
}
