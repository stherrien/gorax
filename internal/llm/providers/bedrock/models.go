package bedrock

import "github.com/gorax/gorax/internal/llm"

// Claude (Anthropic on Bedrock) request/response types

type claudeRequest struct {
	AnthropicVersion string          `json:"anthropic_version"`
	Messages         []claudeMessage `json:"messages"`
	System           string          `json:"system,omitempty"`
	MaxTokens        int             `json:"max_tokens"`
	Temperature      float64         `json:"temperature,omitempty"`
	TopP             float64         `json:"top_p,omitempty"`
	TopK             int             `json:"top_k,omitempty"`
	StopSequences    []string        `json:"stop_sequences,omitempty"`
}

type claudeMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type claudeResponse struct {
	ID           string         `json:"id"`
	Type         string         `json:"type"`
	Role         string         `json:"role"`
	Content      []claudeContent `json:"content"`
	Model        string         `json:"model"`
	StopReason   string         `json:"stop_reason"`
	StopSequence *string        `json:"stop_sequence,omitempty"`
	Usage        claudeUsage    `json:"usage"`
}

type claudeContent struct {
	Type string `json:"type"`
	Text string `json:"text,omitempty"`
}

type claudeUsage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}

// Amazon Titan text request/response types

type titanRequest struct {
	InputText            string                `json:"inputText"`
	TextGenerationConfig titanGenerationConfig `json:"textGenerationConfig,omitempty"`
}

type titanGenerationConfig struct {
	MaxTokenCount int      `json:"maxTokenCount,omitempty"`
	Temperature   float64  `json:"temperature,omitempty"`
	TopP          float64  `json:"topP,omitempty"`
	StopSequences []string `json:"stopSequences,omitempty"`
}

type titanResponse struct {
	InputTextTokenCount int           `json:"inputTextTokenCount"`
	Results             []titanResult `json:"results"`
}

type titanResult struct {
	TokenCount       int    `json:"tokenCount"`
	OutputText       string `json:"outputText"`
	CompletionReason string `json:"completionReason"`
}

// Amazon Titan embeddings request/response types

type titanEmbedRequest struct {
	InputText string `json:"inputText"`
}

type titanEmbedResponse struct {
	Embedding           []float64 `json:"embedding"`
	InputTextTokenCount int       `json:"inputTextTokenCount"`
}

// Predefined Bedrock models with capabilities and pricing
var bedrockModels = []llm.Model{
	// Claude models via Bedrock
	{
		ID:              "anthropic.claude-3-opus-20240229-v1:0",
		Name:            "Claude 3 Opus (Bedrock)",
		Provider:        providerName,
		MaxTokens:       4096,
		ContextWindow:   200000,
		InputCostPer1M:  1500,
		OutputCostPer1M: 7500,
		Capabilities:    []string{llm.CapabilityChat, llm.CapabilityFunction, llm.CapabilityVision},
	},
	{
		ID:              "anthropic.claude-3-sonnet-20240229-v1:0",
		Name:            "Claude 3 Sonnet (Bedrock)",
		Provider:        providerName,
		MaxTokens:       4096,
		ContextWindow:   200000,
		InputCostPer1M:  300,
		OutputCostPer1M: 1500,
		Capabilities:    []string{llm.CapabilityChat, llm.CapabilityFunction, llm.CapabilityVision},
	},
	{
		ID:              "anthropic.claude-3-haiku-20240307-v1:0",
		Name:            "Claude 3 Haiku (Bedrock)",
		Provider:        providerName,
		MaxTokens:       4096,
		ContextWindow:   200000,
		InputCostPer1M:  25,
		OutputCostPer1M: 125,
		Capabilities:    []string{llm.CapabilityChat, llm.CapabilityFunction, llm.CapabilityVision},
	},
	{
		ID:              "anthropic.claude-3-5-sonnet-20241022-v2:0",
		Name:            "Claude 3.5 Sonnet (Bedrock)",
		Provider:        providerName,
		MaxTokens:       8192,
		ContextWindow:   200000,
		InputCostPer1M:  300,
		OutputCostPer1M: 1500,
		Capabilities:    []string{llm.CapabilityChat, llm.CapabilityFunction, llm.CapabilityVision},
	},
	// Amazon Titan models
	{
		ID:              "amazon.titan-text-express-v1",
		Name:            "Titan Text Express",
		Provider:        providerName,
		MaxTokens:       8000,
		ContextWindow:   8000,
		InputCostPer1M:  80,
		OutputCostPer1M: 240,
		Capabilities:    []string{llm.CapabilityChat},
	},
	{
		ID:              "amazon.titan-text-lite-v1",
		Name:            "Titan Text Lite",
		Provider:        providerName,
		MaxTokens:       4000,
		ContextWindow:   4000,
		InputCostPer1M:  15,
		OutputCostPer1M: 20,
		Capabilities:    []string{llm.CapabilityChat},
	},
	{
		ID:              "amazon.titan-embed-text-v2:0",
		Name:            "Titan Embeddings V2",
		Provider:        providerName,
		MaxTokens:       0,
		ContextWindow:   8192,
		InputCostPer1M:  2,
		OutputCostPer1M: 0,
		Capabilities:    []string{llm.CapabilityEmbedding},
	},
}
