package openai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/gorax/gorax/internal/llm"
)

const (
	defaultBaseURL = "https://api.openai.com"
	providerName   = "openai"
)

// Client is an OpenAI API client
type Client struct {
	apiKey       string
	organization string
	baseURL      string
	httpClient   *http.Client
}

// NewClient creates a new OpenAI client
func NewClient(config *llm.ProviderConfig) (*Client, error) {
	if config == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}
	if config.APIKey == "" {
		return nil, llm.ErrInvalidAPIKey
	}

	baseURL := defaultBaseURL
	if config.BaseURL != "" {
		baseURL = config.BaseURL
	}

	timeout := 60 * time.Second
	if config.Timeout > 0 {
		timeout = config.Timeout
	}

	return &Client{
		apiKey:       config.APIKey,
		organization: config.Organization,
		baseURL:      baseURL,
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}, nil
}

// Name returns the provider name
func (c *Client) Name() string {
	return providerName
}

// ChatCompletion performs a chat completion request
func (c *Client) ChatCompletion(ctx context.Context, req *llm.ChatRequest) (*llm.ChatResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	// Convert to OpenAI request format
	apiReq := chatCompletionRequest{
		Model:    req.Model,
		Messages: make([]chatMessage, len(req.Messages)),
	}

	for i, msg := range req.Messages {
		apiReq.Messages[i] = chatMessage{
			Role:    msg.Role,
			Content: msg.Content,
			Name:    msg.Name,
		}
	}

	if req.MaxTokens > 0 {
		apiReq.MaxTokens = req.MaxTokens
	}
	if req.Temperature != nil {
		apiReq.Temperature = req.Temperature
	}
	if req.TopP != nil {
		apiReq.TopP = req.TopP
	}
	if len(req.Stop) > 0 {
		apiReq.Stop = req.Stop
	}
	if req.PresencePenalty != nil {
		apiReq.PresencePenalty = req.PresencePenalty
	}
	if req.FrequencyPenalty != nil {
		apiReq.FrequencyPenalty = req.FrequencyPenalty
	}
	if req.ResponseFormat != nil {
		apiReq.ResponseFormat = &responseFormat{
			Type: req.ResponseFormat.Type,
		}
	}
	if req.User != "" {
		apiReq.User = req.User
	}

	var apiResp chatCompletionResponse
	if err := c.doRequest(ctx, "POST", "/v1/chat/completions", apiReq, &apiResp); err != nil {
		return nil, err
	}

	if len(apiResp.Choices) == 0 {
		return nil, llm.ErrInvalidResponse
	}

	choice := apiResp.Choices[0]
	return &llm.ChatResponse{
		ID:    apiResp.ID,
		Model: apiResp.Model,
		Message: llm.ChatMessage{
			Role:    choice.Message.Role,
			Content: choice.Message.Content,
		},
		FinishReason: choice.FinishReason,
		Usage: llm.TokenUsage{
			PromptTokens:     apiResp.Usage.PromptTokens,
			CompletionTokens: apiResp.Usage.CompletionTokens,
			TotalTokens:      apiResp.Usage.TotalTokens,
		},
	}, nil
}

// GenerateEmbeddings generates embeddings for the given texts
func (c *Client) GenerateEmbeddings(ctx context.Context, req *llm.EmbeddingRequest) (*llm.EmbeddingResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	apiReq := embeddingRequest{
		Model: req.Model,
		Input: req.Texts,
	}
	if req.User != "" {
		apiReq.User = req.User
	}

	var apiResp embeddingResponse
	if err := c.doRequest(ctx, "POST", "/v1/embeddings", apiReq, &apiResp); err != nil {
		return nil, err
	}

	embeddings := make([][]float64, len(apiResp.Data))
	for _, data := range apiResp.Data {
		embeddings[data.Index] = data.Embedding
	}

	return &llm.EmbeddingResponse{
		Model:      apiResp.Model,
		Embeddings: embeddings,
		Usage: llm.TokenUsage{
			PromptTokens: apiResp.Usage.PromptTokens,
			TotalTokens:  apiResp.Usage.TotalTokens,
		},
	}, nil
}

// CountTokens estimates the token count for the given text
// This is an approximation based on OpenAI's tokenizer rules
func (c *Client) CountTokens(text string, model string) (int, error) {
	if text == "" {
		return 0, nil
	}
	// Rough approximation: ~4 characters per token for English text
	// For more accurate counting, use tiktoken library
	return (len(text) + 3) / 4, nil
}

// ListModels returns the available OpenAI models
func (c *Client) ListModels(ctx context.Context) ([]llm.Model, error) {
	return openAIModels, nil
}

// HealthCheck verifies the API connection
func (c *Client) HealthCheck(ctx context.Context) error {
	var result map[string]interface{}
	return c.doRequest(ctx, "GET", "/v1/models", nil, &result)
}

// doRequest performs an HTTP request to the OpenAI API
func (c *Client) doRequest(ctx context.Context, method, path string, body interface{}, result interface{}) error {
	var bodyReader io.Reader
	if body != nil {
		bodyJSON, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("failed to marshal request: %w", err)
		}
		bodyReader = bytes.NewReader(bodyJSON)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, bodyReader)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")
	if c.organization != "" {
		req.Header.Set("OpenAI-Organization", c.organization)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	// Handle errors
	if resp.StatusCode >= 400 {
		return c.parseError(resp.StatusCode, resp.Header, respBody)
	}

	// Parse successful response
	if result != nil && len(respBody) > 0 {
		if err := json.Unmarshal(respBody, result); err != nil {
			return fmt.Errorf("failed to parse response: %w", err)
		}
	}

	return nil
}

// parseError converts an HTTP error response to an LLMError
func (c *Client) parseError(statusCode int, headers http.Header, body []byte) error {
	var errResp errorResponse
	if err := json.Unmarshal(body, &errResp); err != nil {
		return &llm.LLMError{
			Provider:   providerName,
			Code:       fmt.Sprintf("%d", statusCode),
			Message:    string(body),
			HTTPStatus: statusCode,
		}
	}

	llmErr := &llm.LLMError{
		Provider:   providerName,
		Code:       errResp.Error.Code,
		Message:    errResp.Error.Message,
		HTTPStatus: statusCode,
	}

	// Handle rate limiting
	if statusCode == http.StatusTooManyRequests {
		llmErr.Retryable = true
		llmErr.Cause = llm.ErrRateLimitExceeded
		if retryAfter := headers.Get("Retry-After"); retryAfter != "" {
			if seconds, err := strconv.Atoi(retryAfter); err == nil {
				llmErr.RetryAfter = seconds
			}
		}
		return llmErr
	}

	// Handle authentication errors
	if statusCode == http.StatusUnauthorized || statusCode == http.StatusForbidden {
		llmErr.Cause = llm.ErrInvalidAPIKey
		return llmErr
	}

	// Handle context length errors
	if errResp.Error.Code == "context_length_exceeded" {
		llmErr.Cause = llm.ErrContextTooLong
		return llmErr
	}

	// Handle server errors (retryable)
	if statusCode >= 500 {
		llmErr.Retryable = true
		llmErr.Cause = llm.ErrProviderUnavailable
		return llmErr
	}

	return llmErr
}

// Predefined OpenAI models with capabilities and pricing
var openAIModels = []llm.Model{
	{
		ID:            "gpt-4-turbo",
		Name:          "GPT-4 Turbo",
		Provider:      providerName,
		MaxTokens:     4096,
		ContextWindow: 128000,
		InputCostPer1M:  1000,
		OutputCostPer1M: 3000,
		Capabilities:  []string{llm.CapabilityChat, llm.CapabilityFunction, llm.CapabilityJSON, llm.CapabilityVision},
	},
	{
		ID:            "gpt-4o",
		Name:          "GPT-4o",
		Provider:      providerName,
		MaxTokens:     4096,
		ContextWindow: 128000,
		InputCostPer1M:  500,
		OutputCostPer1M: 1500,
		Capabilities:  []string{llm.CapabilityChat, llm.CapabilityFunction, llm.CapabilityJSON, llm.CapabilityVision},
	},
	{
		ID:            "gpt-4o-mini",
		Name:          "GPT-4o Mini",
		Provider:      providerName,
		MaxTokens:     16384,
		ContextWindow: 128000,
		InputCostPer1M:  15,
		OutputCostPer1M: 60,
		Capabilities:  []string{llm.CapabilityChat, llm.CapabilityFunction, llm.CapabilityJSON, llm.CapabilityVision},
	},
	{
		ID:            "gpt-4",
		Name:          "GPT-4",
		Provider:      providerName,
		MaxTokens:     4096,
		ContextWindow: 8192,
		InputCostPer1M:  3000,
		OutputCostPer1M: 6000,
		Capabilities:  []string{llm.CapabilityChat, llm.CapabilityFunction, llm.CapabilityJSON},
	},
	{
		ID:            "gpt-3.5-turbo",
		Name:          "GPT-3.5 Turbo",
		Provider:      providerName,
		MaxTokens:     4096,
		ContextWindow: 16385,
		InputCostPer1M:  50,
		OutputCostPer1M: 150,
		Capabilities:  []string{llm.CapabilityChat, llm.CapabilityFunction, llm.CapabilityJSON},
	},
	{
		ID:            "text-embedding-3-small",
		Name:          "Text Embedding 3 Small",
		Provider:      providerName,
		MaxTokens:     0,
		ContextWindow: 8191,
		InputCostPer1M:  2,
		OutputCostPer1M: 0,
		Capabilities:  []string{llm.CapabilityEmbedding},
	},
	{
		ID:            "text-embedding-3-large",
		Name:          "Text Embedding 3 Large",
		Provider:      providerName,
		MaxTokens:     0,
		ContextWindow: 8191,
		InputCostPer1M:  13,
		OutputCostPer1M: 0,
		Capabilities:  []string{llm.CapabilityEmbedding},
	},
}

// RegisterWithGlobal registers the OpenAI provider with the global registry
func RegisterWithGlobal() error {
	return llm.RegisterProvider(providerName, func(config *llm.ProviderConfig) (llm.Provider, error) {
		return NewClient(config)
	})
}
