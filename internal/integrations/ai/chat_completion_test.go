package ai

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/gorax/gorax/internal/credential"
	"github.com/gorax/gorax/internal/executor/actions"
	"github.com/gorax/gorax/internal/llm"
)

// mockCredentialService is a mock credential service for testing
type mockCredentialService struct {
	mock.Mock
}

func (m *mockCredentialService) GetValue(ctx context.Context, tenantID, credentialID, userID string) (*credential.DecryptedValue, error) {
	args := m.Called(ctx, tenantID, credentialID, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*credential.DecryptedValue), args.Error(1)
}

// mockProvider is a mock LLM provider for testing
type mockProvider struct {
	mock.Mock
}

func (m *mockProvider) ChatCompletion(ctx context.Context, req *llm.ChatRequest) (*llm.ChatResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*llm.ChatResponse), args.Error(1)
}

func (m *mockProvider) GenerateEmbeddings(ctx context.Context, req *llm.EmbeddingRequest) (*llm.EmbeddingResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*llm.EmbeddingResponse), args.Error(1)
}

func (m *mockProvider) CountTokens(text string, model string) (int, error) {
	args := m.Called(text, model)
	return args.Int(0), args.Error(1)
}

func (m *mockProvider) ListModels(ctx context.Context) ([]llm.Model, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]llm.Model), args.Error(1)
}

func (m *mockProvider) Name() string {
	args := m.Called()
	return args.String(0)
}

func (m *mockProvider) HealthCheck(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

// mockProviderRegistry is a mock registry for testing
type mockProviderRegistry struct {
	mock.Mock
}

func (m *mockProviderRegistry) GetProviderFromCredential(name string, credValue map[string]interface{}) (llm.Provider, error) {
	args := m.Called(name, credValue)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(llm.Provider), args.Error(1)
}

func TestChatCompletionAction_Execute(t *testing.T) {
	t.Run("successful completion with OpenAI", func(t *testing.T) {
		// Setup mocks
		mockCredSvc := new(mockCredentialService)
		mockRegistry := new(mockProviderRegistry)
		mockProv := new(mockProvider)

		// Setup credential return
		mockCredSvc.On("GetValue", mock.Anything, "tenant-123", "cred-456", "system").Return(
			&credential.DecryptedValue{
				Value: map[string]interface{}{
					"api_key": "sk-test123",
				},
			}, nil,
		)

		// Setup registry to return mock provider
		mockRegistry.On("GetProviderFromCredential", "openai", map[string]interface{}{
			"api_key": "sk-test123",
		}).Return(mockProv, nil)

		// Setup provider response
		mockProv.On("ChatCompletion", mock.Anything, mock.MatchedBy(func(req *llm.ChatRequest) bool {
			return req.Model == "gpt-4o" &&
				len(req.Messages) == 2 &&
				req.Messages[0].Role == "system" &&
				req.Messages[1].Content == "Hello!"
		})).Return(&llm.ChatResponse{
			ID:    "chatcmpl-123",
			Model: "gpt-4o",
			Message: llm.ChatMessage{
				Role:    "assistant",
				Content: "Hello! How can I help you today?",
			},
			FinishReason: "stop",
			Usage: llm.TokenUsage{
				PromptTokens:     20,
				CompletionTokens: 10,
				TotalTokens:      30,
			},
		}, nil)

		// Create action
		action := NewChatCompletionAction(mockCredSvc, mockRegistry)

		// Build input
		config := ChatCompletionConfig{
			Provider:     "openai",
			Model:        "gpt-4o",
			SystemPrompt: "You are a helpful assistant.",
			Messages: []llm.ChatMessage{
				{Role: "user", Content: "Hello!"},
			},
		}

		input := actions.NewActionInput(config, map[string]interface{}{
			"env": map[string]interface{}{
				"tenant_id": "tenant-123",
			},
			"credential_id": "cred-456",
		})

		// Execute
		output, err := action.Execute(context.Background(), input)

		// Assert
		require.NoError(t, err)
		require.NotNil(t, output)

		result, ok := output.Data.(*ChatCompletionResult)
		require.True(t, ok)
		assert.Equal(t, "chatcmpl-123", result.ID)
		assert.Equal(t, "Hello! How can I help you today?", result.Content)
		assert.Equal(t, "assistant", result.Role)
		assert.Equal(t, "stop", result.FinishReason)
		assert.Equal(t, 20, result.Usage.PromptTokens)
		assert.Equal(t, 10, result.Usage.CompletionTokens)
		assert.Equal(t, 30, result.Usage.TotalTokens)

		// Verify metadata
		assert.Equal(t, "gpt-4o", output.Metadata["model"])
		assert.Equal(t, "openai", output.Metadata["provider"])
		assert.Equal(t, 30, output.Metadata["total_tokens"])

		mockCredSvc.AssertExpectations(t)
		mockRegistry.AssertExpectations(t)
		mockProv.AssertExpectations(t)
	})

	t.Run("with temperature and max_tokens", func(t *testing.T) {
		mockCredSvc := new(mockCredentialService)
		mockRegistry := new(mockProviderRegistry)
		mockProv := new(mockProvider)

		mockCredSvc.On("GetValue", mock.Anything, "tenant-123", "cred-456", "system").Return(
			&credential.DecryptedValue{
				Value: map[string]interface{}{"api_key": "sk-test123"},
			}, nil,
		)

		mockRegistry.On("GetProviderFromCredential", "openai", mock.Anything).Return(mockProv, nil)

		mockProv.On("ChatCompletion", mock.Anything, mock.MatchedBy(func(req *llm.ChatRequest) bool {
			return req.MaxTokens == 500 &&
				req.Temperature != nil &&
				*req.Temperature == 0.7
		})).Return(&llm.ChatResponse{
			ID:           "chatcmpl-456",
			Model:        "gpt-4o",
			Message:      llm.ChatMessage{Role: "assistant", Content: "Response"},
			FinishReason: "stop",
			Usage:        llm.TokenUsage{PromptTokens: 10, CompletionTokens: 5, TotalTokens: 15},
		}, nil)

		action := NewChatCompletionAction(mockCredSvc, mockRegistry)

		temp := 0.7
		config := ChatCompletionConfig{
			Provider:    "openai",
			Model:       "gpt-4o",
			Messages:    []llm.ChatMessage{{Role: "user", Content: "Hi"}},
			MaxTokens:   500,
			Temperature: &temp,
		}

		input := actions.NewActionInput(config, map[string]interface{}{
			"env":           map[string]interface{}{"tenant_id": "tenant-123"},
			"credential_id": "cred-456",
		})

		output, err := action.Execute(context.Background(), input)

		require.NoError(t, err)
		require.NotNil(t, output)

		mockProv.AssertExpectations(t)
	})

	t.Run("with Anthropic provider", func(t *testing.T) {
		mockCredSvc := new(mockCredentialService)
		mockRegistry := new(mockProviderRegistry)
		mockProv := new(mockProvider)

		mockCredSvc.On("GetValue", mock.Anything, "tenant-123", "cred-456", "system").Return(
			&credential.DecryptedValue{
				Value: map[string]interface{}{"api_key": "sk-ant-test123"},
			}, nil,
		)

		mockRegistry.On("GetProviderFromCredential", "anthropic", map[string]interface{}{
			"api_key": "sk-ant-test123",
		}).Return(mockProv, nil)

		mockProv.On("ChatCompletion", mock.Anything, mock.MatchedBy(func(req *llm.ChatRequest) bool {
			return req.Model == "claude-3-sonnet-20240229"
		})).Return(&llm.ChatResponse{
			ID:           "msg_123",
			Model:        "claude-3-sonnet-20240229",
			Message:      llm.ChatMessage{Role: "assistant", Content: "Hello from Claude!"},
			FinishReason: "end_turn",
			Usage:        llm.TokenUsage{PromptTokens: 15, CompletionTokens: 8, TotalTokens: 23},
		}, nil)

		action := NewChatCompletionAction(mockCredSvc, mockRegistry)

		config := ChatCompletionConfig{
			Provider: "anthropic",
			Model:    "claude-3-sonnet-20240229",
			Messages: []llm.ChatMessage{{Role: "user", Content: "Hello!"}},
		}

		input := actions.NewActionInput(config, map[string]interface{}{
			"env":           map[string]interface{}{"tenant_id": "tenant-123"},
			"credential_id": "cred-456",
		})

		output, err := action.Execute(context.Background(), input)

		require.NoError(t, err)
		result := output.Data.(*ChatCompletionResult)
		assert.Equal(t, "Hello from Claude!", result.Content)
		assert.Equal(t, "anthropic", output.Metadata["provider"])
	})

	t.Run("missing tenant_id", func(t *testing.T) {
		mockCredSvc := new(mockCredentialService)
		mockRegistry := new(mockProviderRegistry)

		action := NewChatCompletionAction(mockCredSvc, mockRegistry)

		config := ChatCompletionConfig{
			Provider: "openai",
			Model:    "gpt-4o",
			Messages: []llm.ChatMessage{{Role: "user", Content: "Hi"}},
		}

		input := actions.NewActionInput(config, map[string]interface{}{
			"credential_id": "cred-456",
			// Missing env.tenant_id
		})

		output, err := action.Execute(context.Background(), input)

		assert.Error(t, err)
		assert.Nil(t, output)
		assert.Contains(t, err.Error(), "tenant_id")
	})

	t.Run("missing credential_id", func(t *testing.T) {
		mockCredSvc := new(mockCredentialService)
		mockRegistry := new(mockProviderRegistry)

		action := NewChatCompletionAction(mockCredSvc, mockRegistry)

		config := ChatCompletionConfig{
			Provider: "openai",
			Model:    "gpt-4o",
			Messages: []llm.ChatMessage{{Role: "user", Content: "Hi"}},
		}

		input := actions.NewActionInput(config, map[string]interface{}{
			"env": map[string]interface{}{"tenant_id": "tenant-123"},
			// Missing credential_id
		})

		output, err := action.Execute(context.Background(), input)

		assert.Error(t, err)
		assert.Nil(t, output)
		assert.Contains(t, err.Error(), "credential_id")
	})

	t.Run("credential retrieval failure", func(t *testing.T) {
		mockCredSvc := new(mockCredentialService)
		mockRegistry := new(mockProviderRegistry)

		mockCredSvc.On("GetValue", mock.Anything, "tenant-123", "cred-456", "system").Return(
			nil, assert.AnError,
		)

		action := NewChatCompletionAction(mockCredSvc, mockRegistry)

		config := ChatCompletionConfig{
			Provider: "openai",
			Model:    "gpt-4o",
			Messages: []llm.ChatMessage{{Role: "user", Content: "Hi"}},
		}

		input := actions.NewActionInput(config, map[string]interface{}{
			"env":           map[string]interface{}{"tenant_id": "tenant-123"},
			"credential_id": "cred-456",
		})

		output, err := action.Execute(context.Background(), input)

		assert.Error(t, err)
		assert.Nil(t, output)
		assert.Contains(t, err.Error(), "credential")
	})

	t.Run("provider creation failure", func(t *testing.T) {
		mockCredSvc := new(mockCredentialService)
		mockRegistry := new(mockProviderRegistry)

		mockCredSvc.On("GetValue", mock.Anything, "tenant-123", "cred-456", "system").Return(
			&credential.DecryptedValue{
				Value: map[string]interface{}{"api_key": "invalid"},
			}, nil,
		)

		mockRegistry.On("GetProviderFromCredential", "openai", mock.Anything).Return(
			nil, llm.ErrInvalidAPIKey,
		)

		action := NewChatCompletionAction(mockCredSvc, mockRegistry)

		config := ChatCompletionConfig{
			Provider: "openai",
			Model:    "gpt-4o",
			Messages: []llm.ChatMessage{{Role: "user", Content: "Hi"}},
		}

		input := actions.NewActionInput(config, map[string]interface{}{
			"env":           map[string]interface{}{"tenant_id": "tenant-123"},
			"credential_id": "cred-456",
		})

		output, err := action.Execute(context.Background(), input)

		assert.Error(t, err)
		assert.Nil(t, output)
	})

	t.Run("LLM API failure", func(t *testing.T) {
		mockCredSvc := new(mockCredentialService)
		mockRegistry := new(mockProviderRegistry)
		mockProv := new(mockProvider)

		mockCredSvc.On("GetValue", mock.Anything, "tenant-123", "cred-456", "system").Return(
			&credential.DecryptedValue{
				Value: map[string]interface{}{"api_key": "sk-test123"},
			}, nil,
		)

		mockRegistry.On("GetProviderFromCredential", "openai", mock.Anything).Return(mockProv, nil)

		mockProv.On("ChatCompletion", mock.Anything, mock.Anything).Return(
			nil, llm.ErrRateLimitExceeded,
		)

		action := NewChatCompletionAction(mockCredSvc, mockRegistry)

		config := ChatCompletionConfig{
			Provider: "openai",
			Model:    "gpt-4o",
			Messages: []llm.ChatMessage{{Role: "user", Content: "Hi"}},
		}

		input := actions.NewActionInput(config, map[string]interface{}{
			"env":           map[string]interface{}{"tenant_id": "tenant-123"},
			"credential_id": "cred-456",
		})

		output, err := action.Execute(context.Background(), input)

		assert.Error(t, err)
		assert.Nil(t, output)
		assert.ErrorIs(t, err, llm.ErrRateLimitExceeded)
	})

	t.Run("invalid config type", func(t *testing.T) {
		mockCredSvc := new(mockCredentialService)
		mockRegistry := new(mockProviderRegistry)

		action := NewChatCompletionAction(mockCredSvc, mockRegistry)

		input := actions.NewActionInput("invalid config", map[string]interface{}{
			"env":           map[string]interface{}{"tenant_id": "tenant-123"},
			"credential_id": "cred-456",
		})

		output, err := action.Execute(context.Background(), input)

		assert.Error(t, err)
		assert.Nil(t, output)
		assert.Contains(t, err.Error(), "invalid config type")
	})
}

func TestChatCompletionConfig_Validate(t *testing.T) {
	t.Run("valid config", func(t *testing.T) {
		config := ChatCompletionConfig{
			Provider: "openai",
			Model:    "gpt-4o",
			Messages: []llm.ChatMessage{
				{Role: "user", Content: "Hello"},
			},
		}
		err := config.Validate()
		assert.NoError(t, err)
	})

	t.Run("missing provider", func(t *testing.T) {
		config := ChatCompletionConfig{
			Model:    "gpt-4o",
			Messages: []llm.ChatMessage{{Role: "user", Content: "Hello"}},
		}
		err := config.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "provider")
	})

	t.Run("missing model", func(t *testing.T) {
		config := ChatCompletionConfig{
			Provider: "openai",
			Messages: []llm.ChatMessage{{Role: "user", Content: "Hello"}},
		}
		err := config.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "model")
	})

	t.Run("empty messages", func(t *testing.T) {
		config := ChatCompletionConfig{
			Provider: "openai",
			Model:    "gpt-4o",
			Messages: []llm.ChatMessage{},
		}
		err := config.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "messages")
	})

	t.Run("messages nil", func(t *testing.T) {
		config := ChatCompletionConfig{
			Provider: "openai",
			Model:    "gpt-4o",
		}
		err := config.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "messages")
	})

	t.Run("valid with system prompt only", func(t *testing.T) {
		// System prompt alone is valid if we have at least one user message
		config := ChatCompletionConfig{
			Provider:     "openai",
			Model:        "gpt-4o",
			SystemPrompt: "You are helpful",
			Messages:     []llm.ChatMessage{{Role: "user", Content: "Hello"}},
		}
		err := config.Validate()
		assert.NoError(t, err)
	})
}
