package ai

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/gorax/gorax/internal/llm"
)

func TestSummarizationConfig_Validate(t *testing.T) {
	t.Run("valid config", func(t *testing.T) {
		config := SummarizationConfig{
			Provider: "openai",
			Model:    "gpt-4o",
			Text:     "This is some text to summarize.",
		}
		err := config.Validate()
		assert.NoError(t, err)
	})

	t.Run("missing provider", func(t *testing.T) {
		config := SummarizationConfig{
			Model: "gpt-4o",
			Text:  "Text",
		}
		err := config.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "provider")
	})

	t.Run("missing model", func(t *testing.T) {
		config := SummarizationConfig{
			Provider: "openai",
			Text:     "Text",
		}
		err := config.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "model")
	})

	t.Run("missing text", func(t *testing.T) {
		config := SummarizationConfig{
			Provider: "openai",
			Model:    "gpt-4o",
		}
		err := config.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "text")
	})
}

func TestClassificationConfig_Validate(t *testing.T) {
	t.Run("valid config", func(t *testing.T) {
		config := ClassificationConfig{
			Provider:   "openai",
			Model:      "gpt-4o",
			Text:       "This is some text to classify.",
			Categories: []string{"positive", "negative", "neutral"},
		}
		err := config.Validate()
		assert.NoError(t, err)
	})

	t.Run("missing categories", func(t *testing.T) {
		config := ClassificationConfig{
			Provider: "openai",
			Model:    "gpt-4o",
			Text:     "Text",
		}
		err := config.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "categories")
	})
}

func TestEntityExtractionConfig_Validate(t *testing.T) {
	t.Run("valid config with entity_types", func(t *testing.T) {
		config := EntityExtractionConfig{
			Provider:    "openai",
			Model:       "gpt-4o",
			Text:        "John Smith works at Acme Corp.",
			EntityTypes: []string{"person", "organization"},
		}
		err := config.Validate()
		assert.NoError(t, err)
	})

	t.Run("valid config with custom_entities", func(t *testing.T) {
		config := EntityExtractionConfig{
			Provider: "openai",
			Model:    "gpt-4o",
			Text:     "The order #12345 is for a Blue Widget.",
			CustomEntities: map[string]string{
				"order_number": "Order ID starting with #",
				"product":      "Product name",
			},
		}
		err := config.Validate()
		assert.NoError(t, err)
	})

	t.Run("missing entity types", func(t *testing.T) {
		config := EntityExtractionConfig{
			Provider: "openai",
			Model:    "gpt-4o",
			Text:     "Text",
		}
		err := config.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "entity_types")
	})
}

func TestEmbeddingConfig_Validate(t *testing.T) {
	t.Run("valid config", func(t *testing.T) {
		config := EmbeddingConfig{
			Provider: "openai",
			Model:    "text-embedding-3-small",
			Texts:    []string{"Hello world", "Goodbye world"},
		}
		err := config.Validate()
		assert.NoError(t, err)
	})

	t.Run("missing texts", func(t *testing.T) {
		config := EmbeddingConfig{
			Provider: "openai",
			Model:    "text-embedding-3-small",
		}
		err := config.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "texts")
	})
}

func TestCountWords(t *testing.T) {
	tests := []struct {
		text     string
		expected int
	}{
		{"hello world", 2},
		{"one", 1},
		{"", 0},
		{"   spaces   between   words   ", 3},
		{"new\nlines\nare\nwords", 4},
		{"tabs\tare\tseparators", 3},
	}

	for _, tt := range tests {
		t.Run(tt.text, func(t *testing.T) {
			count := countWords(tt.text)
			assert.Equal(t, tt.expected, count)
		})
	}
}

func TestNewSummarizationAction(t *testing.T) {
	mockCredSvc := new(mockCredentialService)
	mockRegistry := new(mockProviderRegistry)

	action := NewSummarizationAction(mockCredSvc, mockRegistry)
	assert.NotNil(t, action)
	assert.Equal(t, mockCredSvc, action.credentialService)
	assert.Equal(t, mockRegistry, action.providerRegistry)
}

func TestNewClassificationAction(t *testing.T) {
	mockCredSvc := new(mockCredentialService)
	mockRegistry := new(mockProviderRegistry)

	action := NewClassificationAction(mockCredSvc, mockRegistry)
	assert.NotNil(t, action)
}

func TestNewEntityExtractionAction(t *testing.T) {
	mockCredSvc := new(mockCredentialService)
	mockRegistry := new(mockProviderRegistry)

	action := NewEntityExtractionAction(mockCredSvc, mockRegistry)
	assert.NotNil(t, action)
}

func TestNewEmbeddingAction(t *testing.T) {
	mockCredSvc := new(mockCredentialService)
	mockRegistry := new(mockProviderRegistry)

	action := NewEmbeddingAction(mockCredSvc, mockRegistry)
	assert.NotNil(t, action)
}

func TestClassificationAction_parseResponse(t *testing.T) {
	action := &ClassificationAction{}

	t.Run("valid JSON response", func(t *testing.T) {
		content := `{"category": "positive", "confidence": 0.95, "reasoning": "The text expresses positive sentiment."}`
		result, err := action.parseResponse(content, false)
		assert.NoError(t, err)
		assert.Equal(t, "positive", result.Category)
		assert.Equal(t, 0.95, result.Confidence)
		assert.Equal(t, "The text expresses positive sentiment.", result.Reasoning)
	})

	t.Run("JSON with markdown code block", func(t *testing.T) {
		content := "```json\n{\"category\": \"negative\", \"confidence\": 0.8, \"reasoning\": \"Negative words detected.\"}\n```"
		result, err := action.parseResponse(content, false)
		assert.NoError(t, err)
		assert.Equal(t, "negative", result.Category)
	})

	t.Run("multi-label response", func(t *testing.T) {
		content := `{"categories": ["urgent", "customer-support"], "confidence": 0.9, "reasoning": "Multiple categories apply."}`
		result, err := action.parseResponse(content, true)
		assert.NoError(t, err)
		assert.Equal(t, "urgent", result.Category) // First category becomes primary
		assert.Contains(t, result.Categories, "urgent")
		assert.Contains(t, result.Categories, "customer-support")
	})
}

func TestEntityExtractionAction_parseResponse(t *testing.T) {
	action := &EntityExtractionAction{}

	t.Run("valid JSON response", func(t *testing.T) {
		content := `{
			"entities": [
				{"type": "person", "value": "John Smith", "confidence": 0.95},
				{"type": "organization", "value": "Acme Corp", "confidence": 0.9}
			]
		}`
		result, err := action.parseResponse(content)
		assert.NoError(t, err)
		assert.Equal(t, 2, result.EntityCount)
		assert.Len(t, result.Entities, 2)
		assert.Equal(t, "person", result.Entities[0].Type)
		assert.Equal(t, "John Smith", result.Entities[0].Value)
	})

	t.Run("empty entities", func(t *testing.T) {
		content := `{"entities": []}`
		result, err := action.parseResponse(content)
		assert.NoError(t, err)
		assert.Equal(t, 0, result.EntityCount)
	})
}

func TestSummarizationAction_buildChatRequest(t *testing.T) {
	action := &SummarizationAction{}

	t.Run("basic request", func(t *testing.T) {
		config := &SummarizationConfig{
			Model: "gpt-4o",
			Text:  "Some long text to summarize.",
		}
		req := action.buildChatRequest(config)
		assert.Equal(t, "gpt-4o", req.Model)
		assert.Len(t, req.Messages, 2)
		assert.Equal(t, "system", req.Messages[0].Role)
		assert.Equal(t, "user", req.Messages[1].Role)
		assert.Contains(t, req.Messages[1].Content, "Some long text to summarize.")
	})

	t.Run("with max_length", func(t *testing.T) {
		config := &SummarizationConfig{
			Model:     "gpt-4o",
			Text:      "Text",
			MaxLength: 100,
		}
		req := action.buildChatRequest(config)
		assert.Contains(t, req.Messages[1].Content, "100 words")
	})

	t.Run("with bullet format", func(t *testing.T) {
		config := &SummarizationConfig{
			Model:  "gpt-4o",
			Text:   "Text",
			Format: "bullets",
		}
		req := action.buildChatRequest(config)
		assert.Contains(t, req.Messages[0].Content, "bullet points")
	})

	t.Run("with focus area", func(t *testing.T) {
		config := &SummarizationConfig{
			Model: "gpt-4o",
			Text:  "Text",
			Focus: "financial metrics",
		}
		req := action.buildChatRequest(config)
		assert.Contains(t, req.Messages[0].Content, "financial metrics")
	})
}

func TestEmbeddingResult(t *testing.T) {
	result := &EmbeddingResult{
		Embeddings: [][]float64{
			{0.1, 0.2, 0.3},
			{0.4, 0.5, 0.6},
		},
		Dimensions: 3,
		Count:      2,
		Usage: llm.TokenUsage{
			PromptTokens: 10,
			TotalTokens:  10,
		},
	}

	assert.Equal(t, 2, result.Count)
	assert.Equal(t, 3, result.Dimensions)
	assert.Len(t, result.Embeddings, 2)
}
