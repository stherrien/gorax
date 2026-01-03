package sso

import (
	"context"
	"fmt"
)

// DefaultProviderFactory implements ProviderFactory
type DefaultProviderFactory struct{}

// NewProviderFactory creates a new provider factory
func NewProviderFactory() ProviderFactory {
	return &DefaultProviderFactory{}
}

// CreateProvider creates an SSO provider based on the provider type
// Note: Actual provider creation is delegated to avoid import cycles
func (f *DefaultProviderFactory) CreateProvider(ctx context.Context, provider *Provider) (SSOProvider, error) {
	if provider == nil {
		return nil, fmt.Errorf("provider cannot be nil")
	}

	// To avoid import cycles, provider creation is handled by createProviderFunc
	// which is injected when the factory is initialized
	return nil, fmt.Errorf("provider factory must be initialized with createProviderFunc")
}

// CreateProviderFunc is a function type for creating providers
type CreateProviderFunc func(ctx context.Context, provider *Provider) (SSOProvider, error)
