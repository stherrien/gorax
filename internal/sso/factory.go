package sso

import (
	"context"
	"fmt"

	"github.com/gorax/gorax/internal/sso/oidc"
	"github.com/gorax/gorax/internal/sso/saml"
)

// DefaultProviderFactory implements ProviderFactory
type DefaultProviderFactory struct{}

// NewProviderFactory creates a new provider factory
func NewProviderFactory() ProviderFactory {
	return &DefaultProviderFactory{}
}

// CreateProvider creates an SSO provider based on the provider type
func (f *DefaultProviderFactory) CreateProvider(ctx context.Context, provider *Provider) (SSOProvider, error) {
	if provider == nil {
		return nil, fmt.Errorf("provider cannot be nil")
	}

	switch provider.Type {
	case ProviderTypeSAML:
		return saml.NewProvider(ctx, provider)
	case ProviderTypeOIDC:
		return oidc.NewProvider(ctx, provider)
	default:
		return nil, fmt.Errorf("unsupported provider type: %s", provider.Type)
	}
}
