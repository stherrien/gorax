package llm

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestChatRequest_Validate(t *testing.T) {
	validTemp := 0.7
	invalidTempHigh := 2.5
	invalidTempLow := -0.5
	validTopP := 0.9
	invalidTopP := 1.5

	tests := []struct {
		name    string
		request *ChatRequest
		wantErr error
	}{
		{
			name: "valid request",
			request: &ChatRequest{
				Model: "gpt-4",
				Messages: []ChatMessage{
					{Role: "user", Content: "Hello"},
				},
			},
			wantErr: nil,
		},
		{
			name: "valid request with all fields",
			request: &ChatRequest{
				Model: "gpt-4",
				Messages: []ChatMessage{
					{Role: "system", Content: "You are helpful"},
					{Role: "user", Content: "Hello"},
				},
				MaxTokens:   100,
				Temperature: &validTemp,
				TopP:        &validTopP,
				Stop:        []string{"\n"},
			},
			wantErr: nil,
		},
		{
			name: "missing model",
			request: &ChatRequest{
				Messages: []ChatMessage{
					{Role: "user", Content: "Hello"},
				},
			},
			wantErr: ErrInvalidModel,
		},
		{
			name: "empty messages",
			request: &ChatRequest{
				Model:    "gpt-4",
				Messages: []ChatMessage{},
			},
			wantErr: ErrEmptyMessages,
		},
		{
			name: "nil messages",
			request: &ChatRequest{
				Model: "gpt-4",
			},
			wantErr: ErrEmptyMessages,
		},
		{
			name: "invalid role - empty",
			request: &ChatRequest{
				Model: "gpt-4",
				Messages: []ChatMessage{
					{Role: "", Content: "Hello"},
				},
			},
			wantErr: ErrInvalidRole,
		},
		{
			name: "invalid role - unknown",
			request: &ChatRequest{
				Model: "gpt-4",
				Messages: []ChatMessage{
					{Role: "bot", Content: "Hello"},
				},
			},
			wantErr: ErrInvalidRole,
		},
		{
			name: "temperature too high",
			request: &ChatRequest{
				Model: "gpt-4",
				Messages: []ChatMessage{
					{Role: "user", Content: "Hello"},
				},
				Temperature: &invalidTempHigh,
			},
			wantErr: ErrInvalidTemperature,
		},
		{
			name: "temperature too low",
			request: &ChatRequest{
				Model: "gpt-4",
				Messages: []ChatMessage{
					{Role: "user", Content: "Hello"},
				},
				Temperature: &invalidTempLow,
			},
			wantErr: ErrInvalidTemperature,
		},
		{
			name: "top_p invalid",
			request: &ChatRequest{
				Model: "gpt-4",
				Messages: []ChatMessage{
					{Role: "user", Content: "Hello"},
				},
				TopP: &invalidTopP,
			},
			wantErr: ErrInvalidTopP,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.request.Validate()
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestEmbeddingRequest_Validate(t *testing.T) {
	tests := []struct {
		name    string
		request *EmbeddingRequest
		wantErr error
	}{
		{
			name: "valid request",
			request: &EmbeddingRequest{
				Model: "text-embedding-3-small",
				Texts: []string{"Hello world"},
			},
			wantErr: nil,
		},
		{
			name: "valid request with multiple texts",
			request: &EmbeddingRequest{
				Model: "text-embedding-3-small",
				Texts: []string{"Hello", "World", "Test"},
			},
			wantErr: nil,
		},
		{
			name: "missing model",
			request: &EmbeddingRequest{
				Texts: []string{"Hello"},
			},
			wantErr: ErrInvalidModel,
		},
		{
			name: "empty texts",
			request: &EmbeddingRequest{
				Model: "text-embedding-3-small",
				Texts: []string{},
			},
			wantErr: ErrEmptyTexts,
		},
		{
			name: "nil texts",
			request: &EmbeddingRequest{
				Model: "text-embedding-3-small",
			},
			wantErr: ErrEmptyTexts,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.request.Validate()
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestMessageHelpers(t *testing.T) {
	t.Run("SystemMessage", func(t *testing.T) {
		msg := SystemMessage("You are helpful")
		assert.Equal(t, "system", msg.Role)
		assert.Equal(t, "You are helpful", msg.Content)
	})

	t.Run("UserMessage", func(t *testing.T) {
		msg := UserMessage("Hello")
		assert.Equal(t, "user", msg.Role)
		assert.Equal(t, "Hello", msg.Content)
	})

	t.Run("AssistantMessage", func(t *testing.T) {
		msg := AssistantMessage("Hi there")
		assert.Equal(t, "assistant", msg.Role)
		assert.Equal(t, "Hi there", msg.Content)
	})

	t.Run("NewChatMessage", func(t *testing.T) {
		msg := NewChatMessage("user", "Test content")
		assert.Equal(t, "user", msg.Role)
		assert.Equal(t, "Test content", msg.Content)
	})
}

func TestModel_HasCapability(t *testing.T) {
	model := Model{
		ID:           "gpt-4",
		Name:         "GPT-4",
		Provider:     "openai",
		Capabilities: []string{CapabilityChat, CapabilityFunction, CapabilityJSON},
	}

	t.Run("has chat capability", func(t *testing.T) {
		assert.True(t, model.HasCapability(CapabilityChat))
	})

	t.Run("has function capability", func(t *testing.T) {
		assert.True(t, model.HasCapability(CapabilityFunction))
	})

	t.Run("does not have embedding capability", func(t *testing.T) {
		assert.False(t, model.HasCapability(CapabilityEmbedding))
	})

	t.Run("does not have vision capability", func(t *testing.T) {
		assert.False(t, model.HasCapability(CapabilityVision))
	})
}

func TestDefaultProviderConfig(t *testing.T) {
	config := DefaultProviderConfig()

	require.NotNil(t, config)
	assert.Equal(t, 3, config.MaxRetries)
	assert.NotZero(t, config.Timeout)
}
