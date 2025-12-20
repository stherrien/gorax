package bedrock

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
	"github.com/gorax/gorax/internal/llm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockBedrockClient implements the BedrockAPI interface for testing
type mockBedrockClient struct {
	invokeModelFunc func(ctx context.Context, params *bedrockruntime.InvokeModelInput, optFns ...func(*bedrockruntime.Options)) (*bedrockruntime.InvokeModelOutput, error)
}

func (m *mockBedrockClient) InvokeModel(ctx context.Context, params *bedrockruntime.InvokeModelInput, optFns ...func(*bedrockruntime.Options)) (*bedrockruntime.InvokeModelOutput, error) {
	return m.invokeModelFunc(ctx, params, optFns...)
}

func TestNewClient(t *testing.T) {
	t.Run("valid config", func(t *testing.T) {
		config := &llm.ProviderConfig{
			AWSAccessKeyID:     "AKIATEST123",
			AWSSecretAccessKey: "secret123",
			Region:             "us-east-1",
		}
		client, err := NewClient(config)
		require.NoError(t, err)
		assert.NotNil(t, client)
		assert.Equal(t, "bedrock", client.Name())
	})

	t.Run("nil config", func(t *testing.T) {
		client, err := NewClient(nil)
		assert.Error(t, err)
		assert.Nil(t, client)
		assert.Contains(t, err.Error(), "config cannot be nil")
	})

	t.Run("missing region", func(t *testing.T) {
		config := &llm.ProviderConfig{
			AWSAccessKeyID:     "AKIATEST123",
			AWSSecretAccessKey: "secret123",
		}
		client, err := NewClient(config)
		assert.Error(t, err)
		assert.Nil(t, client)
		assert.Contains(t, err.Error(), "region is required")
	})

	t.Run("missing credentials", func(t *testing.T) {
		config := &llm.ProviderConfig{
			Region: "us-east-1",
		}
		// Should not error - will use default credential chain
		client, err := NewClient(config)
		require.NoError(t, err)
		assert.NotNil(t, client)
	})
}

func TestClient_ChatCompletion_Claude(t *testing.T) {
	t.Run("successful Claude completion", func(t *testing.T) {
		mockClient := &mockBedrockClient{
			invokeModelFunc: func(ctx context.Context, params *bedrockruntime.InvokeModelInput, optFns ...func(*bedrockruntime.Options)) (*bedrockruntime.InvokeModelOutput, error) {
				// Verify model ID
				assert.Equal(t, "anthropic.claude-3-sonnet-20240229-v1:0", *params.ModelId)

				// Parse request body
				var req claudeRequest
				err := json.Unmarshal(params.Body, &req)
				require.NoError(t, err)
				assert.Equal(t, "Hello!", req.Messages[0].Content)

				// Return mock response
				resp := claudeResponse{
					ID:   "msg_123",
					Type: "message",
					Role: "assistant",
					Content: []claudeContent{
						{Type: "text", Text: "Hello! How can I help you today?"},
					},
					Model:      "claude-3-sonnet-20240229",
					StopReason: "end_turn",
					Usage: claudeUsage{
						InputTokens:  10,
						OutputTokens: 15,
					},
				}
				respBody, _ := json.Marshal(resp)
				return &bedrockruntime.InvokeModelOutput{
					Body: respBody,
				}, nil
			},
		}

		client := &Client{
			region:    "us-east-1",
			apiClient: mockClient,
		}

		resp, err := client.ChatCompletion(context.Background(), &llm.ChatRequest{
			Model: "anthropic.claude-3-sonnet-20240229-v1:0",
			Messages: []llm.ChatMessage{
				{Role: "user", Content: "Hello!"},
			},
		})

		require.NoError(t, err)
		assert.Equal(t, "msg_123", resp.ID)
		assert.Equal(t, "Hello! How can I help you today?", resp.Message.Content)
		assert.Equal(t, "assistant", resp.Message.Role)
		assert.Equal(t, 10, resp.Usage.PromptTokens)
		assert.Equal(t, 15, resp.Usage.CompletionTokens)
	})

	t.Run("with system prompt", func(t *testing.T) {
		mockClient := &mockBedrockClient{
			invokeModelFunc: func(ctx context.Context, params *bedrockruntime.InvokeModelInput, optFns ...func(*bedrockruntime.Options)) (*bedrockruntime.InvokeModelOutput, error) {
				var req claudeRequest
				err := json.Unmarshal(params.Body, &req)
				require.NoError(t, err)

				// System prompt should be in top-level field
				assert.Equal(t, "You are a helpful assistant.", req.System)
				// Messages should only contain user message
				assert.Len(t, req.Messages, 1)
				assert.Equal(t, "user", req.Messages[0].Role)

				resp := claudeResponse{
					ID:         "msg_456",
					Type:       "message",
					Role:       "assistant",
					Content:    []claudeContent{{Type: "text", Text: "I'm here to help!"}},
					Model:      "claude-3-sonnet-20240229",
					StopReason: "end_turn",
					Usage:      claudeUsage{InputTokens: 20, OutputTokens: 10},
				}
				respBody, _ := json.Marshal(resp)
				return &bedrockruntime.InvokeModelOutput{Body: respBody}, nil
			},
		}

		client := &Client{
			region:    "us-east-1",
			apiClient: mockClient,
		}

		resp, err := client.ChatCompletion(context.Background(), &llm.ChatRequest{
			Model: "anthropic.claude-3-sonnet-20240229-v1:0",
			Messages: []llm.ChatMessage{
				{Role: "system", Content: "You are a helpful assistant."},
				{Role: "user", Content: "Hello!"},
			},
		})

		require.NoError(t, err)
		assert.Equal(t, "I'm here to help!", resp.Message.Content)
	})

	t.Run("with max tokens and temperature", func(t *testing.T) {
		mockClient := &mockBedrockClient{
			invokeModelFunc: func(ctx context.Context, params *bedrockruntime.InvokeModelInput, optFns ...func(*bedrockruntime.Options)) (*bedrockruntime.InvokeModelOutput, error) {
				var req claudeRequest
				err := json.Unmarshal(params.Body, &req)
				require.NoError(t, err)

				assert.Equal(t, 500, req.MaxTokens)
				assert.Equal(t, 0.7, req.Temperature)

				resp := claudeResponse{
					ID:         "msg_789",
					Type:       "message",
					Role:       "assistant",
					Content:    []claudeContent{{Type: "text", Text: "Response"}},
					Model:      "claude-3-sonnet-20240229",
					StopReason: "end_turn",
					Usage:      claudeUsage{InputTokens: 5, OutputTokens: 3},
				}
				respBody, _ := json.Marshal(resp)
				return &bedrockruntime.InvokeModelOutput{Body: respBody}, nil
			},
		}

		client := &Client{
			region:    "us-east-1",
			apiClient: mockClient,
		}

		temp := 0.7
		resp, err := client.ChatCompletion(context.Background(), &llm.ChatRequest{
			Model:       "anthropic.claude-3-sonnet-20240229-v1:0",
			Messages:    []llm.ChatMessage{{Role: "user", Content: "Hi"}},
			MaxTokens:   500,
			Temperature: &temp,
		})

		require.NoError(t, err)
		assert.Equal(t, "Response", resp.Message.Content)
	})
}

func TestClient_ChatCompletion_Titan(t *testing.T) {
	t.Run("successful Titan completion", func(t *testing.T) {
		mockClient := &mockBedrockClient{
			invokeModelFunc: func(ctx context.Context, params *bedrockruntime.InvokeModelInput, optFns ...func(*bedrockruntime.Options)) (*bedrockruntime.InvokeModelOutput, error) {
				// Verify model ID
				assert.Equal(t, "amazon.titan-text-express-v1", *params.ModelId)

				// Parse request body - Titan uses different format
				var req titanRequest
				err := json.Unmarshal(params.Body, &req)
				require.NoError(t, err)
				assert.Contains(t, req.InputText, "Hello!")

				// Return mock response
				resp := titanResponse{
					Results: []titanResult{
						{
							TokenCount:       5,
							OutputText:       "Hello! How may I assist you?",
							CompletionReason: "FINISH",
						},
					},
					InputTextTokenCount: 10,
				}
				respBody, _ := json.Marshal(resp)
				return &bedrockruntime.InvokeModelOutput{
					Body: respBody,
				}, nil
			},
		}

		client := &Client{
			region:    "us-east-1",
			apiClient: mockClient,
		}

		resp, err := client.ChatCompletion(context.Background(), &llm.ChatRequest{
			Model: "amazon.titan-text-express-v1",
			Messages: []llm.ChatMessage{
				{Role: "user", Content: "Hello!"},
			},
		})

		require.NoError(t, err)
		assert.Equal(t, "Hello! How may I assist you?", resp.Message.Content)
		assert.Equal(t, "assistant", resp.Message.Role)
		assert.Equal(t, 10, resp.Usage.PromptTokens)
		assert.Equal(t, 5, resp.Usage.CompletionTokens)
	})
}

func TestClient_ChatCompletion_Validation(t *testing.T) {
	client := &Client{
		region:    "us-east-1",
		apiClient: &mockBedrockClient{},
	}

	t.Run("empty messages", func(t *testing.T) {
		resp, err := client.ChatCompletion(context.Background(), &llm.ChatRequest{
			Model:    "anthropic.claude-3-sonnet-20240229-v1:0",
			Messages: []llm.ChatMessage{},
		})
		assert.Error(t, err)
		assert.Nil(t, resp)
	})

	t.Run("empty model", func(t *testing.T) {
		resp, err := client.ChatCompletion(context.Background(), &llm.ChatRequest{
			Model:    "",
			Messages: []llm.ChatMessage{{Role: "user", Content: "Hi"}},
		})
		assert.Error(t, err)
		assert.Nil(t, resp)
	})
}

func TestClient_GenerateEmbeddings(t *testing.T) {
	t.Run("Titan embeddings", func(t *testing.T) {
		mockClient := &mockBedrockClient{
			invokeModelFunc: func(ctx context.Context, params *bedrockruntime.InvokeModelInput, optFns ...func(*bedrockruntime.Options)) (*bedrockruntime.InvokeModelOutput, error) {
				assert.Equal(t, "amazon.titan-embed-text-v2:0", *params.ModelId)

				var req titanEmbedRequest
				err := json.Unmarshal(params.Body, &req)
				require.NoError(t, err)
				assert.Equal(t, "Hello world", req.InputText)

				resp := titanEmbedResponse{
					Embedding:           []float64{0.1, 0.2, 0.3, 0.4},
					InputTextTokenCount: 2,
				}
				respBody, _ := json.Marshal(resp)
				return &bedrockruntime.InvokeModelOutput{Body: respBody}, nil
			},
		}

		client := &Client{
			region:    "us-east-1",
			apiClient: mockClient,
		}

		resp, err := client.GenerateEmbeddings(context.Background(), &llm.EmbeddingRequest{
			Model: "amazon.titan-embed-text-v2:0",
			Texts: []string{"Hello world"},
		})

		require.NoError(t, err)
		assert.Len(t, resp.Embeddings, 1)
		assert.Equal(t, []float64{0.1, 0.2, 0.3, 0.4}, resp.Embeddings[0])
	})

	t.Run("multiple texts", func(t *testing.T) {
		callCount := 0
		mockClient := &mockBedrockClient{
			invokeModelFunc: func(ctx context.Context, params *bedrockruntime.InvokeModelInput, optFns ...func(*bedrockruntime.Options)) (*bedrockruntime.InvokeModelOutput, error) {
				callCount++
				resp := titanEmbedResponse{
					Embedding:           []float64{float64(callCount) * 0.1, float64(callCount) * 0.2},
					InputTextTokenCount: 2,
				}
				respBody, _ := json.Marshal(resp)
				return &bedrockruntime.InvokeModelOutput{Body: respBody}, nil
			},
		}

		client := &Client{
			region:    "us-east-1",
			apiClient: mockClient,
		}

		resp, err := client.GenerateEmbeddings(context.Background(), &llm.EmbeddingRequest{
			Model: "amazon.titan-embed-text-v2:0",
			Texts: []string{"Hello", "World"},
		})

		require.NoError(t, err)
		assert.Len(t, resp.Embeddings, 2)
		assert.Equal(t, 2, callCount) // One call per text
	})

	t.Run("unsupported Claude embeddings", func(t *testing.T) {
		client := &Client{
			region:    "us-east-1",
			apiClient: &mockBedrockClient{},
		}

		resp, err := client.GenerateEmbeddings(context.Background(), &llm.EmbeddingRequest{
			Model: "anthropic.claude-3-sonnet-20240229-v1:0",
			Texts: []string{"Hello"},
		})

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.ErrorIs(t, err, llm.ErrUnsupportedOperation)
	})
}

func TestClient_CountTokens(t *testing.T) {
	client := &Client{region: "us-east-1"}

	t.Run("approximate count", func(t *testing.T) {
		count, err := client.CountTokens("Hello, how are you doing today?", "anthropic.claude-3-sonnet-20240229-v1:0")
		require.NoError(t, err)
		assert.Greater(t, count, 0)
	})

	t.Run("empty text", func(t *testing.T) {
		count, err := client.CountTokens("", "anthropic.claude-3-sonnet-20240229-v1:0")
		require.NoError(t, err)
		assert.Equal(t, 0, count)
	})
}

func TestClient_ListModels(t *testing.T) {
	client := &Client{region: "us-east-1"}

	models, err := client.ListModels(context.Background())
	require.NoError(t, err)
	assert.NotEmpty(t, models)

	// Check for expected models
	modelIDs := make([]string, len(models))
	for i, m := range models {
		modelIDs[i] = m.ID
	}
	assert.Contains(t, modelIDs, "anthropic.claude-3-opus-20240229-v1:0")
	assert.Contains(t, modelIDs, "anthropic.claude-3-sonnet-20240229-v1:0")
	assert.Contains(t, modelIDs, "amazon.titan-text-express-v1")
	assert.Contains(t, modelIDs, "amazon.titan-embed-text-v2:0")

	// Verify all models have bedrock provider
	for _, model := range models {
		assert.Equal(t, "bedrock", model.Provider)
	}
}

func TestClient_Name(t *testing.T) {
	client := &Client{region: "us-east-1"}
	assert.Equal(t, "bedrock", client.Name())
}

func TestClient_HealthCheck(t *testing.T) {
	t.Run("healthy", func(t *testing.T) {
		mockClient := &mockBedrockClient{
			invokeModelFunc: func(ctx context.Context, params *bedrockruntime.InvokeModelInput, optFns ...func(*bedrockruntime.Options)) (*bedrockruntime.InvokeModelOutput, error) {
				resp := claudeResponse{
					ID:         "msg_health",
					Type:       "message",
					Role:       "assistant",
					Content:    []claudeContent{{Type: "text", Text: "OK"}},
					Model:      "claude-3-haiku-20240307",
					StopReason: "end_turn",
					Usage:      claudeUsage{InputTokens: 1, OutputTokens: 1},
				}
				respBody, _ := json.Marshal(resp)
				return &bedrockruntime.InvokeModelOutput{Body: respBody}, nil
			},
		}

		client := &Client{
			region:    "us-east-1",
			apiClient: mockClient,
		}

		err := client.HealthCheck(context.Background())
		assert.NoError(t, err)
	})
}

func TestRegisterWithGlobal(t *testing.T) {
	// Reset global registry for clean test
	originalRegistry := llm.GlobalProviderRegistry
	llm.GlobalProviderRegistry = llm.NewProviderRegistry()
	defer func() { llm.GlobalProviderRegistry = originalRegistry }()

	err := RegisterWithGlobal()
	require.NoError(t, err)

	assert.True(t, llm.GlobalProviderRegistry.HasProvider("bedrock"))
}

func TestModelTypeDetection(t *testing.T) {
	tests := []struct {
		modelID  string
		isClaude bool
		isTitan  bool
	}{
		{"anthropic.claude-3-opus-20240229-v1:0", true, false},
		{"anthropic.claude-3-sonnet-20240229-v1:0", true, false},
		{"anthropic.claude-3-haiku-20240307-v1:0", true, false},
		{"amazon.titan-text-express-v1", false, true},
		{"amazon.titan-text-lite-v1", false, true},
		{"amazon.titan-embed-text-v2:0", false, true},
	}

	for _, tt := range tests {
		t.Run(tt.modelID, func(t *testing.T) {
			assert.Equal(t, tt.isClaude, isClaudeModel(tt.modelID))
			assert.Equal(t, tt.isTitan, isTitanModel(tt.modelID))
		})
	}
}
