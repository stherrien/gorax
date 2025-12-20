package llm

import (
	"fmt"
	"sync"
)

// ProviderRegistry manages available LLM providers
type ProviderRegistry struct {
	factories map[string]ProviderFactory
	mu        sync.RWMutex
}

// NewProviderRegistry creates a new provider registry
func NewProviderRegistry() *ProviderRegistry {
	return &ProviderRegistry{
		factories: make(map[string]ProviderFactory),
	}
}

// Register adds a provider factory to the registry
func (r *ProviderRegistry) Register(name string, factory ProviderFactory) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if name == "" {
		return fmt.Errorf("provider name cannot be empty")
	}
	if factory == nil {
		return fmt.Errorf("provider factory cannot be nil")
	}
	if _, exists := r.factories[name]; exists {
		return fmt.Errorf("provider %s already registered", name)
	}

	r.factories[name] = factory
	return nil
}

// GetProvider creates a provider instance with the given configuration
func (r *ProviderRegistry) GetProvider(name string, config *ProviderConfig) (Provider, error) {
	r.mu.RLock()
	factory, exists := r.factories[name]
	r.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("%w: %s", ErrProviderNotFound, name)
	}

	return factory(config)
}

// GetProviderFromCredential creates a provider using credential values
func (r *ProviderRegistry) GetProviderFromCredential(name string, credentialValue map[string]interface{}) (Provider, error) {
	config := DefaultProviderConfig()

	// Extract API key
	if apiKey, ok := credentialValue["api_key"].(string); ok {
		config.APIKey = apiKey
	}

	// Extract organization (OpenAI)
	if org, ok := credentialValue["organization"].(string); ok {
		config.Organization = org
	}

	// Extract AWS credentials (Bedrock)
	if accessKey, ok := credentialValue["access_key_id"].(string); ok {
		config.AWSAccessKeyID = accessKey
	}
	if secretKey, ok := credentialValue["secret_access_key"].(string); ok {
		config.AWSSecretAccessKey = secretKey
	}
	if region, ok := credentialValue["region"].(string); ok {
		config.Region = region
	}

	// Extract base URL override
	if baseURL, ok := credentialValue["base_url"].(string); ok {
		config.BaseURL = baseURL
	}

	return r.GetProvider(name, config)
}

// ListProviders returns all registered provider names
func (r *ProviderRegistry) ListProviders() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.factories))
	for name := range r.factories {
		names = append(names, name)
	}
	return names
}

// HasProvider checks if a provider is registered
func (r *ProviderRegistry) HasProvider(name string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	_, exists := r.factories[name]
	return exists
}

// Unregister removes a provider from the registry
func (r *ProviderRegistry) Unregister(name string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.factories, name)
}

// Provider name constants
const (
	ProviderOpenAI    = "openai"
	ProviderAnthropic = "anthropic"
	ProviderBedrock   = "bedrock"
	ProviderAzure     = "azure_openai"
)

// Global registry instance (can be overridden for testing)
var GlobalProviderRegistry = NewProviderRegistry()

// RegisterProvider registers a provider with the global registry
func RegisterProvider(name string, factory ProviderFactory) error {
	return GlobalProviderRegistry.Register(name, factory)
}

// GetGlobalProvider gets a provider from the global registry
func GetGlobalProvider(name string, config *ProviderConfig) (Provider, error) {
	return GlobalProviderRegistry.GetProvider(name, config)
}
