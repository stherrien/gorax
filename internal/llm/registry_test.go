package llm

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockProvider implements Provider for testing
type mockProvider struct {
	name string
}

func (p *mockProvider) ChatCompletion(ctx context.Context, req *ChatRequest) (*ChatResponse, error) {
	return &ChatResponse{
		ID:    "test-id",
		Model: req.Model,
		Message: ChatMessage{
			Role:    "assistant",
			Content: "Hello",
		},
		FinishReason: "stop",
	}, nil
}

func (p *mockProvider) GenerateEmbeddings(ctx context.Context, req *EmbeddingRequest) (*EmbeddingResponse, error) {
	return &EmbeddingResponse{
		Model:      req.Model,
		Embeddings: [][]float64{{0.1, 0.2, 0.3}},
	}, nil
}

func (p *mockProvider) CountTokens(text string, model string) (int, error) {
	return len(text) / 4, nil
}

func (p *mockProvider) ListModels(ctx context.Context) ([]Model, error) {
	return []Model{
		{ID: "test-model", Name: "Test Model", Provider: p.name},
	}, nil
}

func (p *mockProvider) Name() string {
	return p.name
}

func (p *mockProvider) HealthCheck(ctx context.Context) error {
	return nil
}

func mockProviderFactory(name string) ProviderFactory {
	return func(config *ProviderConfig) (Provider, error) {
		return &mockProvider{name: name}, nil
	}
}

func TestProviderRegistry_Register(t *testing.T) {
	t.Run("successful registration", func(t *testing.T) {
		registry := NewProviderRegistry()
		err := registry.Register("test", mockProviderFactory("test"))
		assert.NoError(t, err)
		assert.True(t, registry.HasProvider("test"))
	})

	t.Run("empty name", func(t *testing.T) {
		registry := NewProviderRegistry()
		err := registry.Register("", mockProviderFactory("test"))
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "name cannot be empty")
	})

	t.Run("nil factory", func(t *testing.T) {
		registry := NewProviderRegistry()
		err := registry.Register("test", nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "factory cannot be nil")
	})

	t.Run("duplicate registration", func(t *testing.T) {
		registry := NewProviderRegistry()
		err := registry.Register("test", mockProviderFactory("test"))
		require.NoError(t, err)

		err = registry.Register("test", mockProviderFactory("test"))
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "already registered")
	})
}

func TestProviderRegistry_GetProvider(t *testing.T) {
	t.Run("get registered provider", func(t *testing.T) {
		registry := NewProviderRegistry()
		err := registry.Register("openai", mockProviderFactory("openai"))
		require.NoError(t, err)

		provider, err := registry.GetProvider("openai", DefaultProviderConfig())
		require.NoError(t, err)
		assert.NotNil(t, provider)
		assert.Equal(t, "openai", provider.Name())
	})

	t.Run("get unregistered provider", func(t *testing.T) {
		registry := NewProviderRegistry()

		provider, err := registry.GetProvider("unknown", DefaultProviderConfig())
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrProviderNotFound)
		assert.Nil(t, provider)
	})
}

func TestProviderRegistry_GetProviderFromCredential(t *testing.T) {
	t.Run("openai credentials", func(t *testing.T) {
		registry := NewProviderRegistry()
		var receivedConfig *ProviderConfig

		factory := func(config *ProviderConfig) (Provider, error) {
			receivedConfig = config
			return &mockProvider{name: "openai"}, nil
		}
		err := registry.Register("openai", factory)
		require.NoError(t, err)

		credentials := map[string]interface{}{
			"api_key":      "sk-test123",
			"organization": "org-test",
		}

		provider, err := registry.GetProviderFromCredential("openai", credentials)
		require.NoError(t, err)
		assert.NotNil(t, provider)
		assert.Equal(t, "sk-test123", receivedConfig.APIKey)
		assert.Equal(t, "org-test", receivedConfig.Organization)
	})

	t.Run("bedrock credentials", func(t *testing.T) {
		registry := NewProviderRegistry()
		var receivedConfig *ProviderConfig

		factory := func(config *ProviderConfig) (Provider, error) {
			receivedConfig = config
			return &mockProvider{name: "bedrock"}, nil
		}
		err := registry.Register("bedrock", factory)
		require.NoError(t, err)

		credentials := map[string]interface{}{
			"access_key_id":     "AKIATEST",
			"secret_access_key": "secret123",
			"region":            "us-east-1",
		}

		provider, err := registry.GetProviderFromCredential("bedrock", credentials)
		require.NoError(t, err)
		assert.NotNil(t, provider)
		assert.Equal(t, "AKIATEST", receivedConfig.AWSAccessKeyID)
		assert.Equal(t, "secret123", receivedConfig.AWSSecretAccessKey)
		assert.Equal(t, "us-east-1", receivedConfig.Region)
	})

	t.Run("custom base url", func(t *testing.T) {
		registry := NewProviderRegistry()
		var receivedConfig *ProviderConfig

		factory := func(config *ProviderConfig) (Provider, error) {
			receivedConfig = config
			return &mockProvider{name: "openai"}, nil
		}
		err := registry.Register("openai", factory)
		require.NoError(t, err)

		credentials := map[string]interface{}{
			"api_key":  "sk-test123",
			"base_url": "https://my-proxy.example.com",
		}

		provider, err := registry.GetProviderFromCredential("openai", credentials)
		require.NoError(t, err)
		assert.NotNil(t, provider)
		assert.Equal(t, "https://my-proxy.example.com", receivedConfig.BaseURL)
	})

	t.Run("unregistered provider", func(t *testing.T) {
		registry := NewProviderRegistry()

		provider, err := registry.GetProviderFromCredential("unknown", map[string]interface{}{})
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrProviderNotFound)
		assert.Nil(t, provider)
	})
}

func TestProviderRegistry_ListProviders(t *testing.T) {
	registry := NewProviderRegistry()

	// Initially empty
	providers := registry.ListProviders()
	assert.Empty(t, providers)

	// Register some providers
	err := registry.Register("openai", mockProviderFactory("openai"))
	require.NoError(t, err)
	err = registry.Register("anthropic", mockProviderFactory("anthropic"))
	require.NoError(t, err)

	providers = registry.ListProviders()
	assert.Len(t, providers, 2)
	assert.Contains(t, providers, "openai")
	assert.Contains(t, providers, "anthropic")
}

func TestProviderRegistry_HasProvider(t *testing.T) {
	registry := NewProviderRegistry()

	assert.False(t, registry.HasProvider("openai"))

	err := registry.Register("openai", mockProviderFactory("openai"))
	require.NoError(t, err)

	assert.True(t, registry.HasProvider("openai"))
	assert.False(t, registry.HasProvider("anthropic"))
}

func TestProviderRegistry_Unregister(t *testing.T) {
	registry := NewProviderRegistry()

	err := registry.Register("openai", mockProviderFactory("openai"))
	require.NoError(t, err)
	assert.True(t, registry.HasProvider("openai"))

	registry.Unregister("openai")
	assert.False(t, registry.HasProvider("openai"))

	// Unregister non-existent provider should not panic
	registry.Unregister("unknown")
}

func TestProviderConstants(t *testing.T) {
	assert.Equal(t, "openai", ProviderOpenAI)
	assert.Equal(t, "anthropic", ProviderAnthropic)
	assert.Equal(t, "bedrock", ProviderBedrock)
	assert.Equal(t, "azure_openai", ProviderAzure)
}

func TestGlobalRegistry(t *testing.T) {
	// Reset global registry for clean test
	originalRegistry := GlobalProviderRegistry
	GlobalProviderRegistry = NewProviderRegistry()
	defer func() { GlobalProviderRegistry = originalRegistry }()

	// Test RegisterProvider
	err := RegisterProvider("test", mockProviderFactory("test"))
	require.NoError(t, err)

	// Test GetGlobalProvider
	provider, err := GetGlobalProvider("test", DefaultProviderConfig())
	require.NoError(t, err)
	assert.NotNil(t, provider)
	assert.Equal(t, "test", provider.Name())
}

// TestProviderRegistry_Concurrency tests thread safety
func TestProviderRegistry_Concurrency(t *testing.T) {
	registry := NewProviderRegistry()

	// Register initial provider
	err := registry.Register("initial", mockProviderFactory("initial"))
	require.NoError(t, err)

	done := make(chan bool)

	// Concurrent reads
	for i := 0; i < 10; i++ {
		go func() {
			for j := 0; j < 100; j++ {
				registry.HasProvider("initial")
				registry.ListProviders()
				_, _ = registry.GetProvider("initial", DefaultProviderConfig())
			}
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}
}
