package anthropic

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorax/gorax/internal/llm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewClient(t *testing.T) {
	t.Run("valid config", func(t *testing.T) {
		config := &llm.ProviderConfig{
			APIKey: "sk-ant-api123",
		}
		client, err := NewClient(config)
		require.NoError(t, err)
		assert.NotNil(t, client)
		assert.Equal(t, "anthropic", client.Name())
	})

	t.Run("nil config", func(t *testing.T) {
		client, err := NewClient(nil)
		assert.Error(t, err)
		assert.Nil(t, client)
		assert.Contains(t, err.Error(), "config cannot be nil")
	})

	t.Run("empty API key", func(t *testing.T) {
		config := &llm.ProviderConfig{
			APIKey: "",
		}
		client, err := NewClient(config)
		assert.Error(t, err)
		assert.Nil(t, client)
		assert.ErrorIs(t, err, llm.ErrInvalidAPIKey)
	})

	t.Run("custom base URL", func(t *testing.T) {
		config := &llm.ProviderConfig{
			APIKey:  "sk-ant-api123",
			BaseURL: "https://custom.example.com",
		}
		client, err := NewClient(config)
		require.NoError(t, err)
		assert.Equal(t, "https://custom.example.com", client.baseURL)
	})
}

func TestClient_ChatCompletion(t *testing.T) {
	t.Run("successful completion", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Verify request
			assert.Equal(t, "POST", r.Method)
			assert.Equal(t, "/v1/messages", r.URL.Path)
			assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
			assert.Contains(t, r.Header.Get("x-api-key"), "sk-ant-")
			assert.Equal(t, "2023-06-01", r.Header.Get("anthropic-version"))

			// Parse request body
			var req messagesRequest
			err := json.NewDecoder(r.Body).Decode(&req)
			require.NoError(t, err)
			assert.Equal(t, "claude-3-sonnet-20240229", req.Model)
			assert.Len(t, req.Messages, 1)
			assert.Equal(t, "user", req.Messages[0].Role)
			assert.Equal(t, "Hello!", req.Messages[0].Content)

			// Send response
			resp := messagesResponse{
				ID:   "msg_123",
				Type: "message",
				Role: "assistant",
				Content: []contentBlock{
					{Type: "text", Text: "Hello! How can I help you today?"},
				},
				Model:        "claude-3-sonnet-20240229",
				StopReason:   "end_turn",
				StopSequence: nil,
				Usage: anthropicUsage{
					InputTokens:  10,
					OutputTokens: 15,
				},
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		client, err := NewClient(&llm.ProviderConfig{
			APIKey:  "sk-ant-api123",
			BaseURL: server.URL,
		})
		require.NoError(t, err)

		resp, err := client.ChatCompletion(context.Background(), &llm.ChatRequest{
			Model: "claude-3-sonnet-20240229",
			Messages: []llm.ChatMessage{
				{Role: "user", Content: "Hello!"},
			},
		})

		require.NoError(t, err)
		assert.Equal(t, "msg_123", resp.ID)
		assert.Equal(t, "claude-3-sonnet-20240229", resp.Model)
		assert.Equal(t, "assistant", resp.Message.Role)
		assert.Equal(t, "Hello! How can I help you today?", resp.Message.Content)
		assert.Equal(t, "end_turn", resp.FinishReason)
		assert.Equal(t, 10, resp.Usage.PromptTokens)
		assert.Equal(t, 15, resp.Usage.CompletionTokens)
		assert.Equal(t, 25, resp.Usage.TotalTokens)
	})

	t.Run("with system prompt", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var req messagesRequest
			err := json.NewDecoder(r.Body).Decode(&req)
			require.NoError(t, err)

			// System prompt should be extracted to top-level system field
			assert.Equal(t, "You are a helpful assistant.", req.System)
			// Messages should only contain user messages
			assert.Len(t, req.Messages, 1)
			assert.Equal(t, "user", req.Messages[0].Role)

			resp := messagesResponse{
				ID:   "msg_456",
				Type: "message",
				Role: "assistant",
				Content: []contentBlock{
					{Type: "text", Text: "I understand, I'm here to help!"},
				},
				Model:      "claude-3-sonnet-20240229",
				StopReason: "end_turn",
				Usage:      anthropicUsage{InputTokens: 20, OutputTokens: 10},
			}
			json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		client, err := NewClient(&llm.ProviderConfig{
			APIKey:  "sk-ant-api123",
			BaseURL: server.URL,
		})
		require.NoError(t, err)

		resp, err := client.ChatCompletion(context.Background(), &llm.ChatRequest{
			Model: "claude-3-sonnet-20240229",
			Messages: []llm.ChatMessage{
				{Role: "system", Content: "You are a helpful assistant."},
				{Role: "user", Content: "Hello!"},
			},
		})

		require.NoError(t, err)
		assert.Equal(t, "I understand, I'm here to help!", resp.Message.Content)
	})

	t.Run("with max tokens and temperature", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var req messagesRequest
			err := json.NewDecoder(r.Body).Decode(&req)
			require.NoError(t, err)

			assert.Equal(t, 500, req.MaxTokens)
			assert.NotNil(t, req.Temperature)
			assert.Equal(t, 0.7, *req.Temperature)

			resp := messagesResponse{
				ID:         "msg_789",
				Type:       "message",
				Role:       "assistant",
				Content:    []contentBlock{{Type: "text", Text: "Response"}},
				Model:      "claude-3-sonnet-20240229",
				StopReason: "end_turn",
				Usage:      anthropicUsage{InputTokens: 5, OutputTokens: 3},
			}
			json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		client, err := NewClient(&llm.ProviderConfig{
			APIKey:  "sk-ant-api123",
			BaseURL: server.URL,
		})
		require.NoError(t, err)

		temp := 0.7
		resp, err := client.ChatCompletion(context.Background(), &llm.ChatRequest{
			Model:       "claude-3-sonnet-20240229",
			Messages:    []llm.ChatMessage{{Role: "user", Content: "Hi"}},
			MaxTokens:   500,
			Temperature: &temp,
		})

		require.NoError(t, err)
		assert.Equal(t, "Response", resp.Message.Content)
	})

	t.Run("invalid request validation", func(t *testing.T) {
		client, err := NewClient(&llm.ProviderConfig{
			APIKey: "sk-ant-api123",
		})
		require.NoError(t, err)

		// Empty messages
		resp, err := client.ChatCompletion(context.Background(), &llm.ChatRequest{
			Model:    "claude-3-sonnet-20240229",
			Messages: []llm.ChatMessage{},
		})
		assert.Error(t, err)
		assert.Nil(t, resp)
	})

	t.Run("rate limit error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Retry-After", "30")
			w.WriteHeader(http.StatusTooManyRequests)
			resp := errorResponse{
				Type: "error",
				Error: apiError{
					Type:    "rate_limit_error",
					Message: "Rate limit exceeded",
				},
			}
			json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		client, err := NewClient(&llm.ProviderConfig{
			APIKey:  "sk-ant-api123",
			BaseURL: server.URL,
		})
		require.NoError(t, err)

		resp, err := client.ChatCompletion(context.Background(), &llm.ChatRequest{
			Model:    "claude-3-sonnet-20240229",
			Messages: []llm.ChatMessage{{Role: "user", Content: "Hi"}},
		})

		assert.Error(t, err)
		assert.Nil(t, resp)

		var llmErr *llm.LLMError
		require.ErrorAs(t, err, &llmErr)
		assert.True(t, llmErr.Retryable)
		assert.Equal(t, 30, llmErr.RetryAfter)
		assert.Equal(t, http.StatusTooManyRequests, llmErr.HTTPStatus)
	})

	t.Run("authentication error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusUnauthorized)
			resp := errorResponse{
				Type: "error",
				Error: apiError{
					Type:    "authentication_error",
					Message: "Invalid API key",
				},
			}
			json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		client, err := NewClient(&llm.ProviderConfig{
			APIKey:  "invalid-key",
			BaseURL: server.URL,
		})
		require.NoError(t, err)

		resp, err := client.ChatCompletion(context.Background(), &llm.ChatRequest{
			Model:    "claude-3-sonnet-20240229",
			Messages: []llm.ChatMessage{{Role: "user", Content: "Hi"}},
		})

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.True(t, llm.IsAuthError(err))
	})

	t.Run("context length exceeded", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
			resp := errorResponse{
				Type: "error",
				Error: apiError{
					Type:    "invalid_request_error",
					Message: "prompt is too long: 250000 tokens > 200000 maximum",
				},
			}
			json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		client, err := NewClient(&llm.ProviderConfig{
			APIKey:  "sk-ant-api123",
			BaseURL: server.URL,
		})
		require.NoError(t, err)

		resp, err := client.ChatCompletion(context.Background(), &llm.ChatRequest{
			Model:    "claude-3-sonnet-20240229",
			Messages: []llm.ChatMessage{{Role: "user", Content: "Hi"}},
		})

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.True(t, llm.IsContextLengthError(err))
	})

	t.Run("server error (retryable)", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			resp := errorResponse{
				Type: "error",
				Error: apiError{
					Type:    "api_error",
					Message: "Internal server error",
				},
			}
			json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		client, err := NewClient(&llm.ProviderConfig{
			APIKey:  "sk-ant-api123",
			BaseURL: server.URL,
		})
		require.NoError(t, err)

		resp, err := client.ChatCompletion(context.Background(), &llm.ChatRequest{
			Model:    "claude-3-sonnet-20240229",
			Messages: []llm.ChatMessage{{Role: "user", Content: "Hi"}},
		})

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.True(t, llm.IsRetryableError(err))
	})

	t.Run("overloaded error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusServiceUnavailable)
			resp := errorResponse{
				Type: "error",
				Error: apiError{
					Type:    "overloaded_error",
					Message: "Anthropic's API is temporarily overloaded",
				},
			}
			json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		client, err := NewClient(&llm.ProviderConfig{
			APIKey:  "sk-ant-api123",
			BaseURL: server.URL,
		})
		require.NoError(t, err)

		resp, err := client.ChatCompletion(context.Background(), &llm.ChatRequest{
			Model:    "claude-3-sonnet-20240229",
			Messages: []llm.ChatMessage{{Role: "user", Content: "Hi"}},
		})

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.True(t, llm.IsRetryableError(err))
	})
}

func TestClient_GenerateEmbeddings(t *testing.T) {
	t.Run("embeddings not supported", func(t *testing.T) {
		client, err := NewClient(&llm.ProviderConfig{
			APIKey: "sk-ant-api123",
		})
		require.NoError(t, err)

		resp, err := client.GenerateEmbeddings(context.Background(), &llm.EmbeddingRequest{
			Model: "claude-3-sonnet-20240229",
			Texts: []string{"Hello world"},
		})

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.ErrorIs(t, err, llm.ErrUnsupportedOperation)
	})
}

func TestClient_CountTokens(t *testing.T) {
	t.Run("approximate token count", func(t *testing.T) {
		client, err := NewClient(&llm.ProviderConfig{
			APIKey: "sk-ant-api123",
		})
		require.NoError(t, err)

		count, err := client.CountTokens("Hello, how are you doing today?", "claude-3-sonnet-20240229")
		require.NoError(t, err)
		// Approximate: ~4 chars per token, "Hello, how are you doing today?" = 31 chars ≈ 8 tokens
		assert.Greater(t, count, 0)
	})

	t.Run("empty text", func(t *testing.T) {
		client, err := NewClient(&llm.ProviderConfig{
			APIKey: "sk-ant-api123",
		})
		require.NoError(t, err)

		count, err := client.CountTokens("", "claude-3-sonnet-20240229")
		require.NoError(t, err)
		assert.Equal(t, 0, count)
	})
}

func TestClient_ListModels(t *testing.T) {
	t.Run("returns predefined models", func(t *testing.T) {
		client, err := NewClient(&llm.ProviderConfig{
			APIKey: "sk-ant-api123",
		})
		require.NoError(t, err)

		models, err := client.ListModels(context.Background())
		require.NoError(t, err)
		assert.NotEmpty(t, models)

		// Check that Claude 3 models are included
		modelIDs := make([]string, len(models))
		for i, m := range models {
			modelIDs[i] = m.ID
		}
		assert.Contains(t, modelIDs, "claude-3-opus-20240229")
		assert.Contains(t, modelIDs, "claude-3-sonnet-20240229")
		assert.Contains(t, modelIDs, "claude-3-haiku-20240307")
		assert.Contains(t, modelIDs, "claude-3-5-sonnet-20241022")

		// Verify model attributes
		for _, model := range models {
			assert.Equal(t, "anthropic", model.Provider)
			assert.Greater(t, model.ContextWindow, 0)
			assert.Contains(t, model.Capabilities, llm.CapabilityChat)
		}
	})
}

func TestClient_HealthCheck(t *testing.T) {
	t.Run("healthy", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Verify it's making a valid request
			assert.Equal(t, "POST", r.Method)
			assert.Equal(t, "/v1/messages", r.URL.Path)

			resp := messagesResponse{
				ID:         "msg_health",
				Type:       "message",
				Role:       "assistant",
				Content:    []contentBlock{{Type: "text", Text: "OK"}},
				Model:      "claude-3-haiku-20240307",
				StopReason: "end_turn",
				Usage:      anthropicUsage{InputTokens: 1, OutputTokens: 1},
			}
			json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		client, err := NewClient(&llm.ProviderConfig{
			APIKey:  "sk-ant-api123",
			BaseURL: server.URL,
		})
		require.NoError(t, err)

		err = client.HealthCheck(context.Background())
		assert.NoError(t, err)
	})

	t.Run("unhealthy", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusServiceUnavailable)
		}))
		defer server.Close()

		client, err := NewClient(&llm.ProviderConfig{
			APIKey:  "sk-ant-api123",
			BaseURL: server.URL,
		})
		require.NoError(t, err)

		err = client.HealthCheck(context.Background())
		assert.Error(t, err)
	})
}

func TestClient_Name(t *testing.T) {
	client, err := NewClient(&llm.ProviderConfig{
		APIKey: "sk-ant-api123",
	})
	require.NoError(t, err)
	assert.Equal(t, "anthropic", client.Name())
}

func TestRegisterWithGlobal(t *testing.T) {
	// Reset global registry for clean test
	originalRegistry := llm.GlobalProviderRegistry
	llm.GlobalProviderRegistry = llm.NewProviderRegistry()
	defer func() { llm.GlobalProviderRegistry = originalRegistry }()

	err := RegisterWithGlobal()
	require.NoError(t, err)

	assert.True(t, llm.GlobalProviderRegistry.HasProvider("anthropic"))
}
