package llm

// ChatMessage represents a single message in a conversation
type ChatMessage struct {
	// Role is the message author: "system", "user", or "assistant"
	Role string `json:"role"`

	// Content is the message text
	Content string `json:"content"`

	// Name is an optional name for the participant (for multi-user conversations)
	Name string `json:"name,omitempty"`
}

// ChatRequest represents a chat completion request
type ChatRequest struct {
	// Model is the model ID to use (e.g., "gpt-4", "claude-3-opus")
	Model string `json:"model"`

	// Messages is the conversation history
	Messages []ChatMessage `json:"messages"`

	// MaxTokens limits the response length (optional, uses model default if 0)
	MaxTokens int `json:"max_tokens,omitempty"`

	// Temperature controls randomness (0.0 to 2.0, default varies by provider)
	Temperature *float64 `json:"temperature,omitempty"`

	// TopP is nucleus sampling parameter (0.0 to 1.0)
	TopP *float64 `json:"top_p,omitempty"`

	// Stop sequences that will halt generation
	Stop []string `json:"stop,omitempty"`

	// PresencePenalty penalizes new tokens based on presence in text so far (-2.0 to 2.0)
	PresencePenalty *float64 `json:"presence_penalty,omitempty"`

	// FrequencyPenalty penalizes new tokens based on frequency in text so far (-2.0 to 2.0)
	FrequencyPenalty *float64 `json:"frequency_penalty,omitempty"`

	// ResponseFormat specifies the output format
	ResponseFormat *ResponseFormat `json:"response_format,omitempty"`

	// User is an optional unique identifier for the end-user (for abuse monitoring)
	User string `json:"user,omitempty"`

	// Metadata holds provider-specific options
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// ResponseFormat specifies the output format for structured responses
type ResponseFormat struct {
	// Type is "text" or "json_object"
	Type string `json:"type"`

	// JSONSchema is an optional JSON schema for structured output (provider-specific support)
	JSONSchema string `json:"json_schema,omitempty"`
}

// ResponseFormatText is the default text response format
const ResponseFormatText = "text"

// ResponseFormatJSON requests JSON object output
const ResponseFormatJSON = "json_object"

// ChatResponse represents the response from a chat completion
type ChatResponse struct {
	// ID is a unique identifier for this completion
	ID string `json:"id"`

	// Model is the actual model used (may differ from requested for aliases)
	Model string `json:"model"`

	// Message is the assistant's response
	Message ChatMessage `json:"message"`

	// FinishReason indicates why generation stopped: "stop", "length", "content_filter"
	FinishReason string `json:"finish_reason"`

	// Usage contains token consumption statistics
	Usage TokenUsage `json:"usage"`
}

// TokenUsage tracks token consumption for billing and monitoring
type TokenUsage struct {
	// PromptTokens is the number of tokens in the input
	PromptTokens int `json:"prompt_tokens"`

	// CompletionTokens is the number of tokens in the output
	CompletionTokens int `json:"completion_tokens"`

	// TotalTokens is the sum of prompt and completion tokens
	TotalTokens int `json:"total_tokens"`
}

// EmbeddingRequest represents a request to generate embeddings
type EmbeddingRequest struct {
	// Model is the embedding model to use
	Model string `json:"model"`

	// Texts is the list of texts to embed
	Texts []string `json:"texts"`

	// User is an optional unique identifier for the end-user
	User string `json:"user,omitempty"`
}

// EmbeddingResponse contains the generated embeddings
type EmbeddingResponse struct {
	// Model is the actual model used
	Model string `json:"model"`

	// Embeddings is a list of embedding vectors, one per input text
	Embeddings [][]float64 `json:"embeddings"`

	// Usage contains token consumption statistics
	Usage TokenUsage `json:"usage"`
}

// Validate validates a ChatRequest
func (r *ChatRequest) Validate() error {
	if r.Model == "" {
		return ErrInvalidModel
	}
	if len(r.Messages) == 0 {
		return ErrEmptyMessages
	}
	for _, msg := range r.Messages {
		if msg.Role == "" {
			return ErrInvalidRole
		}
		if msg.Role != "system" && msg.Role != "user" && msg.Role != "assistant" {
			return ErrInvalidRole
		}
	}
	if r.Temperature != nil && (*r.Temperature < 0 || *r.Temperature > 2) {
		return ErrInvalidTemperature
	}
	if r.TopP != nil && (*r.TopP < 0 || *r.TopP > 1) {
		return ErrInvalidTopP
	}
	return nil
}

// Validate validates an EmbeddingRequest
func (r *EmbeddingRequest) Validate() error {
	if r.Model == "" {
		return ErrInvalidModel
	}
	if len(r.Texts) == 0 {
		return ErrEmptyTexts
	}
	return nil
}

// NewChatMessage creates a new chat message
func NewChatMessage(role, content string) ChatMessage {
	return ChatMessage{Role: role, Content: content}
}

// SystemMessage creates a system message
func SystemMessage(content string) ChatMessage {
	return ChatMessage{Role: "system", Content: content}
}

// UserMessage creates a user message
func UserMessage(content string) ChatMessage {
	return ChatMessage{Role: "user", Content: content}
}

// AssistantMessage creates an assistant message
func AssistantMessage(content string) ChatMessage {
	return ChatMessage{Role: "assistant", Content: content}
}
