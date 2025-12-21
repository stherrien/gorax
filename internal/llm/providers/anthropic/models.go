package anthropic

import "github.com/gorax/gorax/internal/llm"

// Messages API request/response types

type messagesRequest struct {
	Model         string           `json:"model"`
	Messages      []messageContent `json:"messages"`
	System        string           `json:"system,omitempty"`
	MaxTokens     int              `json:"max_tokens"`
	Temperature   *float64         `json:"temperature,omitempty"`
	TopP          *float64         `json:"top_p,omitempty"`
	TopK          *int             `json:"top_k,omitempty"`
	StopSequences []string         `json:"stop_sequences,omitempty"`
	Stream        bool             `json:"stream,omitempty"`
	Metadata      *requestMetadata `json:"metadata,omitempty"`
}

type messageContent struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type requestMetadata struct {
	UserID string `json:"user_id,omitempty"`
}

type messagesResponse struct {
	ID           string         `json:"id"`
	Type         string         `json:"type"`
	Role         string         `json:"role"`
	Content      []contentBlock `json:"content"`
	Model        string         `json:"model"`
	StopReason   string         `json:"stop_reason"`
	StopSequence *string        `json:"stop_sequence,omitempty"`
	Usage        anthropicUsage `json:"usage"`
}

type contentBlock struct {
	Type string `json:"type"`
	Text string `json:"text,omitempty"`
}

type anthropicUsage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}

// Error response types

type errorResponse struct {
	Type  string   `json:"type"`
	Error apiError `json:"error"`
}

type apiError struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

// Predefined Anthropic models with capabilities and pricing
var anthropicModels = []llm.Model{
	{
		ID:              "claude-3-opus-20240229",
		Name:            "Claude 3 Opus",
		Provider:        providerName,
		MaxTokens:       4096,
		ContextWindow:   200000,
		InputCostPer1M:  1500,
		OutputCostPer1M: 7500,
		Capabilities:    []string{llm.CapabilityChat, llm.CapabilityFunction, llm.CapabilityVision},
	},
	{
		ID:              "claude-3-sonnet-20240229",
		Name:            "Claude 3 Sonnet",
		Provider:        providerName,
		MaxTokens:       4096,
		ContextWindow:   200000,
		InputCostPer1M:  300,
		OutputCostPer1M: 1500,
		Capabilities:    []string{llm.CapabilityChat, llm.CapabilityFunction, llm.CapabilityVision},
	},
	{
		ID:              "claude-3-haiku-20240307",
		Name:            "Claude 3 Haiku",
		Provider:        providerName,
		MaxTokens:       4096,
		ContextWindow:   200000,
		InputCostPer1M:  25,
		OutputCostPer1M: 125,
		Capabilities:    []string{llm.CapabilityChat, llm.CapabilityFunction, llm.CapabilityVision},
	},
	{
		ID:              "claude-3-5-sonnet-20241022",
		Name:            "Claude 3.5 Sonnet",
		Provider:        providerName,
		MaxTokens:       8192,
		ContextWindow:   200000,
		InputCostPer1M:  300,
		OutputCostPer1M: 1500,
		Capabilities:    []string{llm.CapabilityChat, llm.CapabilityFunction, llm.CapabilityVision},
	},
	{
		ID:              "claude-3-5-haiku-20241022",
		Name:            "Claude 3.5 Haiku",
		Provider:        providerName,
		MaxTokens:       8192,
		ContextWindow:   200000,
		InputCostPer1M:  100,
		OutputCostPer1M: 500,
		Capabilities:    []string{llm.CapabilityChat, llm.CapabilityFunction, llm.CapabilityVision},
	},
}
