package openai

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gorax/gorax/internal/llm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewClient(t *testing.T) {
	t.Run("valid config", func(t *testing.T) {
		config := &llm.ProviderConfig{
			APIKey: "sk-test123",
		}
		client, err := NewClient(config)
		require.NoError(t, err)
		assert.NotNil(t, client)
		assert.Equal(t, "openai", client.Name())
	})

	t.Run("with organization", func(t *testing.T) {
		config := &llm.ProviderConfig{
			APIKey:       "sk-test123",
			Organization: "org-test",
		}
		client, err := NewClient(config)
		require.NoError(t, err)
		assert.NotNil(t, client)
	})

	t.Run("missing API key", func(t *testing.T) {
		config := &llm.ProviderConfig{}
		client, err := NewClient(config)
		assert.Error(t, err)
		assert.ErrorIs(t, err, llm.ErrInvalidAPIKey)
		assert.Nil(t, client)
	})

	t.Run("nil config", func(t *testing.T) {
		client, err := NewClient(nil)
		assert.Error(t, err)
		assert.Nil(t, client)
	})

	t.Run("custom base URL", func(t *testing.T) {
		config := &llm.ProviderConfig{
			APIKey:  "sk-test123",
			BaseURL: "https://custom-api.example.com",
		}
		client, err := NewClient(config)
		require.NoError(t, err)
		assert.NotNil(t, client)
	})
}

func TestClient_ChatCompletion(t *testing.T) {
	t.Run("successful completion", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Verify request
			assert.Equal(t, "/v1/chat/completions", r.URL.Path)
			assert.Equal(t, "POST", r.Method)
			assert.Contains(t, r.Header.Get("Authorization"), "Bearer sk-test")
			assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

			// Parse request body
			var req chatCompletionRequest
			err := json.NewDecoder(r.Body).Decode(&req)
			require.NoError(t, err)
			assert.Equal(t, "gpt-4", req.Model)
			assert.Len(t, req.Messages, 1)

			// Return response
			resp := chatCompletionResponse{
				ID:    "chatcmpl-123",
				Model: "gpt-4",
				Choices: []chatChoice{
					{
						Index: 0,
						Message: chatMessage{
							Role:    "assistant",
							Content: "Hello! How can I help you?",
						},
						FinishReason: "stop",
					},
				},
				Usage: chatUsage{
					PromptTokens:     10,
					CompletionTokens: 8,
					TotalTokens:      18,
				},
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		client, err := NewClient(&llm.ProviderConfig{
			APIKey:  "sk-test",
			BaseURL: server.URL,
		})
		require.NoError(t, err)

		resp, err := client.ChatCompletion(context.Background(), &llm.ChatRequest{
			Model: "gpt-4",
			Messages: []llm.ChatMessage{
				{Role: "user", Content: "Hello"},
			},
		})

		require.NoError(t, err)
		assert.Equal(t, "chatcmpl-123", resp.ID)
		assert.Equal(t, "gpt-4", resp.Model)
		assert.Equal(t, "assistant", resp.Message.Role)
		assert.Equal(t, "Hello! How can I help you?", resp.Message.Content)
		assert.Equal(t, "stop", resp.FinishReason)
		assert.Equal(t, 10, resp.Usage.PromptTokens)
		assert.Equal(t, 8, resp.Usage.CompletionTokens)
		assert.Equal(t, 18, resp.Usage.TotalTokens)
	})

	t.Run("with organization header", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "org-test123", r.Header.Get("OpenAI-Organization"))

			resp := chatCompletionResponse{
				ID:      "chatcmpl-123",
				Model:   "gpt-4",
				Choices: []chatChoice{{Message: chatMessage{Role: "assistant", Content: "Hi"}, FinishReason: "stop"}},
				Usage:   chatUsage{TotalTokens: 5},
			}
			json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		client, err := NewClient(&llm.ProviderConfig{
			APIKey:       "sk-test",
			Organization: "org-test123",
			BaseURL:      server.URL,
		})
		require.NoError(t, err)

		_, err = client.ChatCompletion(context.Background(), &llm.ChatRequest{
			Model:    "gpt-4",
			Messages: []llm.ChatMessage{{Role: "user", Content: "Hi"}},
		})
		require.NoError(t, err)
	})

	t.Run("with optional parameters", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var req chatCompletionRequest
			json.NewDecoder(r.Body).Decode(&req)

			assert.Equal(t, 100, req.MaxTokens)
			assert.NotNil(t, req.Temperature)
			assert.Equal(t, 0.7, *req.Temperature)
			assert.NotNil(t, req.TopP)
			assert.Equal(t, 0.9, *req.TopP)
			assert.Equal(t, []string{"\n"}, req.Stop)

			resp := chatCompletionResponse{
				ID:      "chatcmpl-123",
				Model:   "gpt-4",
				Choices: []chatChoice{{Message: chatMessage{Role: "assistant", Content: "Response"}, FinishReason: "stop"}},
				Usage:   chatUsage{TotalTokens: 10},
			}
			json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		client, err := NewClient(&llm.ProviderConfig{APIKey: "sk-test", BaseURL: server.URL})
		require.NoError(t, err)

		temp := 0.7
		topP := 0.9
		_, err = client.ChatCompletion(context.Background(), &llm.ChatRequest{
			Model:       "gpt-4",
			Messages:    []llm.ChatMessage{{Role: "user", Content: "Test"}},
			MaxTokens:   100,
			Temperature: &temp,
			TopP:        &topP,
			Stop:        []string{"\n"},
		})
		require.NoError(t, err)
	})

	t.Run("rate limit error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Retry-After", "30")
			w.WriteHeader(http.StatusTooManyRequests)
			json.NewEncoder(w).Encode(errorResponse{
				Error: apiError{
					Message: "Rate limit exceeded",
					Type:    "rate_limit_error",
					Code:    "rate_limit_exceeded",
				},
			})
		}))
		defer server.Close()

		client, err := NewClient(&llm.ProviderConfig{APIKey: "sk-test", BaseURL: server.URL})
		require.NoError(t, err)

		_, err = client.ChatCompletion(context.Background(), &llm.ChatRequest{
			Model:    "gpt-4",
			Messages: []llm.ChatMessage{{Role: "user", Content: "Test"}},
		})

		assert.Error(t, err)
		assert.True(t, llm.IsRateLimitError(err))

		var llmErr *llm.LLMError
		assert.ErrorAs(t, err, &llmErr)
		assert.Equal(t, 30, llmErr.RetryAfter)
		assert.True(t, llmErr.IsRetryable())
	})

	t.Run("authentication error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(errorResponse{
				Error: apiError{
					Message: "Invalid API key",
					Type:    "authentication_error",
					Code:    "invalid_api_key",
				},
			})
		}))
		defer server.Close()

		client, err := NewClient(&llm.ProviderConfig{APIKey: "sk-invalid", BaseURL: server.URL})
		require.NoError(t, err)

		_, err = client.ChatCompletion(context.Background(), &llm.ChatRequest{
			Model:    "gpt-4",
			Messages: []llm.ChatMessage{{Role: "user", Content: "Test"}},
		})

		assert.Error(t, err)
		assert.True(t, llm.IsAuthError(err))
	})

	t.Run("context length error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(errorResponse{
				Error: apiError{
					Message: "This model's maximum context length is 8192 tokens",
					Type:    "invalid_request_error",
					Code:    "context_length_exceeded",
				},
			})
		}))
		defer server.Close()

		client, err := NewClient(&llm.ProviderConfig{APIKey: "sk-test", BaseURL: server.URL})
		require.NoError(t, err)

		_, err = client.ChatCompletion(context.Background(), &llm.ChatRequest{
			Model:    "gpt-4",
			Messages: []llm.ChatMessage{{Role: "user", Content: "Very long text..."}},
		})

		assert.Error(t, err)
		assert.True(t, llm.IsContextLengthError(err))
	})

	t.Run("context cancellation", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(100 * time.Millisecond)
			json.NewEncoder(w).Encode(chatCompletionResponse{})
		}))
		defer server.Close()

		client, err := NewClient(&llm.ProviderConfig{APIKey: "sk-test", BaseURL: server.URL})
		require.NoError(t, err)

		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		_, err = client.ChatCompletion(ctx, &llm.ChatRequest{
			Model:    "gpt-4",
			Messages: []llm.ChatMessage{{Role: "user", Content: "Test"}},
		})

		assert.Error(t, err)
	})

	t.Run("invalid request", func(t *testing.T) {
		client, err := NewClient(&llm.ProviderConfig{APIKey: "sk-test"})
		require.NoError(t, err)

		_, err = client.ChatCompletion(context.Background(), &llm.ChatRequest{
			Model:    "", // Missing model
			Messages: []llm.ChatMessage{{Role: "user", Content: "Test"}},
		})

		assert.Error(t, err)
		assert.ErrorIs(t, err, llm.ErrInvalidModel)
	})
}

func TestClient_GenerateEmbeddings(t *testing.T) {
	t.Run("successful embedding", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/v1/embeddings", r.URL.Path)
			assert.Equal(t, "POST", r.Method)

			var req embeddingRequest
			json.NewDecoder(r.Body).Decode(&req)
			assert.Equal(t, "text-embedding-3-small", req.Model)
			assert.Equal(t, []string{"Hello world"}, req.Input)

			resp := embeddingResponse{
				Model: "text-embedding-3-small",
				Data: []embeddingData{
					{
						Index:     0,
						Embedding: []float64{0.1, 0.2, 0.3, 0.4, 0.5},
					},
				},
				Usage: embeddingUsage{
					PromptTokens: 2,
					TotalTokens:  2,
				},
			}
			json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		client, err := NewClient(&llm.ProviderConfig{APIKey: "sk-test", BaseURL: server.URL})
		require.NoError(t, err)

		resp, err := client.GenerateEmbeddings(context.Background(), &llm.EmbeddingRequest{
			Model: "text-embedding-3-small",
			Texts: []string{"Hello world"},
		})

		require.NoError(t, err)
		assert.Equal(t, "text-embedding-3-small", resp.Model)
		assert.Len(t, resp.Embeddings, 1)
		assert.Len(t, resp.Embeddings[0], 5)
		assert.Equal(t, 2, resp.Usage.PromptTokens)
	})

	t.Run("multiple texts", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var req embeddingRequest
			json.NewDecoder(r.Body).Decode(&req)
			assert.Len(t, req.Input, 3)

			resp := embeddingResponse{
				Model: "text-embedding-3-small",
				Data: []embeddingData{
					{Index: 0, Embedding: []float64{0.1, 0.2}},
					{Index: 1, Embedding: []float64{0.3, 0.4}},
					{Index: 2, Embedding: []float64{0.5, 0.6}},
				},
				Usage: embeddingUsage{PromptTokens: 6, TotalTokens: 6},
			}
			json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		client, err := NewClient(&llm.ProviderConfig{APIKey: "sk-test", BaseURL: server.URL})
		require.NoError(t, err)

		resp, err := client.GenerateEmbeddings(context.Background(), &llm.EmbeddingRequest{
			Model: "text-embedding-3-small",
			Texts: []string{"First", "Second", "Third"},
		})

		require.NoError(t, err)
		assert.Len(t, resp.Embeddings, 3)
	})

	t.Run("invalid request", func(t *testing.T) {
		client, err := NewClient(&llm.ProviderConfig{APIKey: "sk-test"})
		require.NoError(t, err)

		_, err = client.GenerateEmbeddings(context.Background(), &llm.EmbeddingRequest{
			Model: "text-embedding-3-small",
			Texts: []string{}, // Empty texts
		})

		assert.Error(t, err)
		assert.ErrorIs(t, err, llm.ErrEmptyTexts)
	})
}

func TestClient_ListModels(t *testing.T) {
	t.Run("returns predefined models", func(t *testing.T) {
		client, err := NewClient(&llm.ProviderConfig{APIKey: "sk-test"})
		require.NoError(t, err)

		models, err := client.ListModels(context.Background())
		require.NoError(t, err)
		assert.NotEmpty(t, models)

		// Verify GPT-4 is in the list
		var foundGPT4 bool
		for _, m := range models {
			if m.ID == "gpt-4" {
				foundGPT4 = true
				assert.Equal(t, "openai", m.Provider)
				assert.Greater(t, m.ContextWindow, 0)
				assert.True(t, m.HasCapability(llm.CapabilityChat))
			}
		}
		assert.True(t, foundGPT4, "GPT-4 should be in model list")
	})
}

func TestClient_HealthCheck(t *testing.T) {
	t.Run("successful health check", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/v1/models", r.URL.Path)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": []map[string]string{{"id": "gpt-4"}},
			})
		}))
		defer server.Close()

		client, err := NewClient(&llm.ProviderConfig{APIKey: "sk-test", BaseURL: server.URL})
		require.NoError(t, err)

		err = client.HealthCheck(context.Background())
		assert.NoError(t, err)
	})

	t.Run("failed health check", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusUnauthorized)
		}))
		defer server.Close()

		client, err := NewClient(&llm.ProviderConfig{APIKey: "sk-invalid", BaseURL: server.URL})
		require.NoError(t, err)

		err = client.HealthCheck(context.Background())
		assert.Error(t, err)
	})
}

func TestClient_CountTokens(t *testing.T) {
	client, err := NewClient(&llm.ProviderConfig{APIKey: "sk-test"})
	require.NoError(t, err)

	t.Run("count tokens for text", func(t *testing.T) {
		count, err := client.CountTokens("Hello, how are you?", "gpt-4")
		require.NoError(t, err)
		assert.Greater(t, count, 0)
	})

	t.Run("empty text", func(t *testing.T) {
		count, err := client.CountTokens("", "gpt-4")
		require.NoError(t, err)
		assert.Equal(t, 0, count)
	})
}
