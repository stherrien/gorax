package anthropic

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorax/gorax/internal/llm"
)

const (
	defaultBaseURL   = "https://api.anthropic.com"
	providerName     = "anthropic"
	anthropicVersion = "2023-06-01"
	defaultMaxTokens = 4096
)

// Client is an Anthropic API client
type Client struct {
	apiKey     string
	baseURL    string
	httpClient *http.Client
}

// NewClient creates a new Anthropic client
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
		apiKey:  config.APIKey,
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}, nil
}

// Name returns the provider name
func (c *Client) Name() string {
	return providerName
}

// ChatCompletion performs a chat completion request using the Messages API
func (c *Client) ChatCompletion(ctx context.Context, req *llm.ChatRequest) (*llm.ChatResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	apiReq := c.buildMessagesRequest(req)

	var apiResp messagesResponse
	if err := c.doRequest(ctx, "POST", "/v1/messages", apiReq, &apiResp); err != nil {
		return nil, err
	}

	return c.convertResponse(&apiResp), nil
}

// buildMessagesRequest converts the generic chat request to Anthropic's format
func (c *Client) buildMessagesRequest(req *llm.ChatRequest) *messagesRequest {
	apiReq := &messagesRequest{
		Model:     req.Model,
		MaxTokens: defaultMaxTokens,
	}

	// Extract system prompt and convert messages
	var messages []messageContent
	for _, msg := range req.Messages {
		if msg.Role == "system" {
			apiReq.System = msg.Content
			continue
		}
		messages = append(messages, messageContent{
			Role:    msg.Role,
			Content: msg.Content,
		})
	}
	apiReq.Messages = messages

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
		apiReq.StopSequences = req.Stop
	}
	if req.User != "" {
		apiReq.Metadata = &requestMetadata{
			UserID: req.User,
		}
	}

	return apiReq
}

// convertResponse converts the Anthropic response to our generic format
func (c *Client) convertResponse(apiResp *messagesResponse) *llm.ChatResponse {
	// Extract text from content blocks
	var content strings.Builder
	for _, block := range apiResp.Content {
		if block.Type == "text" {
			content.WriteString(block.Text)
		}
	}

	return &llm.ChatResponse{
		ID:    apiResp.ID,
		Model: apiResp.Model,
		Message: llm.ChatMessage{
			Role:    apiResp.Role,
			Content: content.String(),
		},
		FinishReason: apiResp.StopReason,
		Usage: llm.TokenUsage{
			PromptTokens:     apiResp.Usage.InputTokens,
			CompletionTokens: apiResp.Usage.OutputTokens,
			TotalTokens:      apiResp.Usage.InputTokens + apiResp.Usage.OutputTokens,
		},
	}
}

// GenerateEmbeddings is not supported by Anthropic
func (c *Client) GenerateEmbeddings(ctx context.Context, req *llm.EmbeddingRequest) (*llm.EmbeddingResponse, error) {
	return nil, llm.ErrUnsupportedOperation
}

// CountTokens estimates the token count for the given text
// This is an approximation based on Anthropic's tokenizer rules
func (c *Client) CountTokens(text string, model string) (int, error) {
	if text == "" {
		return 0, nil
	}
	// Rough approximation: ~4 characters per token for English text
	// For more accurate counting, use Anthropic's token counting API
	return (len(text) + 3) / 4, nil
}

// ListModels returns the available Anthropic models
func (c *Client) ListModels(ctx context.Context) ([]llm.Model, error) {
	return anthropicModels, nil
}

// HealthCheck verifies the API connection by making a minimal request
func (c *Client) HealthCheck(ctx context.Context) error {
	req := &llm.ChatRequest{
		Model: "claude-3-haiku-20240307", // Use cheapest model
		Messages: []llm.ChatMessage{
			{Role: "user", Content: "hi"},
		},
		MaxTokens: 1,
	}
	_, err := c.ChatCompletion(ctx, req)
	return err
}

// doRequest performs an HTTP request to the Anthropic API
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

	req.Header.Set("x-api-key", c.apiKey)
	req.Header.Set("anthropic-version", anthropicVersion)
	req.Header.Set("Content-Type", "application/json")

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
		Code:       errResp.Error.Type,
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

	// Handle context length errors (check message content)
	if strings.Contains(errResp.Error.Message, "tokens") && strings.Contains(errResp.Error.Message, "maximum") {
		llmErr.Cause = llm.ErrContextTooLong
		return llmErr
	}

	// Handle overloaded errors (503)
	if statusCode == http.StatusServiceUnavailable || errResp.Error.Type == "overloaded_error" {
		llmErr.Retryable = true
		llmErr.Cause = llm.ErrProviderUnavailable
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

// RegisterWithGlobal registers the Anthropic provider with the global registry
func RegisterWithGlobal() error {
	return llm.RegisterProvider(providerName, func(config *llm.ProviderConfig) (llm.Provider, error) {
		return NewClient(config)
	})
}
