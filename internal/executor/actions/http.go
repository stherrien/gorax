package actions

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gorax/gorax/internal/security"
)

// HTTPAction implements the Action interface for HTTP requests
type HTTPAction struct {
	urlValidator *security.URLValidator
}

// NewHTTPAction creates a new HTTP action with default URL validator
func NewHTTPAction() *HTTPAction {
	return &HTTPAction{
		urlValidator: security.NewURLValidator(),
	}
}

// NewHTTPActionWithValidator creates a new HTTP action with custom URL validator
func NewHTTPActionWithValidator(validator *security.URLValidator) *HTTPAction {
	return &HTTPAction{
		urlValidator: validator,
	}
}

// HTTPActionConfig represents the configuration for an HTTP action
type HTTPActionConfig struct {
	Method          string            `json:"method"`
	URL             string            `json:"url"`
	Headers         map[string]string `json:"headers,omitempty"`
	Body            json.RawMessage   `json:"body,omitempty"`
	Timeout         int               `json:"timeout,omitempty"`          // seconds
	Auth            *HTTPAuth         `json:"auth,omitempty"`             // authentication config
	FollowRedirects bool              `json:"follow_redirects,omitempty"` // default: true
}

// HTTPAuth represents HTTP authentication configuration
type HTTPAuth struct {
	Type     string `json:"type"`               // basic, bearer, api_key
	Username string `json:"username,omitempty"` // for basic auth
	Password string `json:"password,omitempty"` // for basic auth
	Token    string `json:"token,omitempty"`    // for bearer auth
	APIKey   string `json:"api_key,omitempty"`  // for api_key auth
	Header   string `json:"header,omitempty"`   // header name for api_key
}

// HTTPActionResult represents the result of an HTTP action
type HTTPActionResult struct {
	StatusCode int               `json:"status_code"`
	Headers    map[string]string `json:"headers"`
	Body       interface{}       `json:"body"`
}

// Execute implements the Action interface
func (a *HTTPAction) Execute(ctx context.Context, input *ActionInput) (*ActionOutput, error) {
	// Parse config
	configBytes, err := json.Marshal(input.Config)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal config: %w", err)
	}

	var config HTTPActionConfig
	if err := json.Unmarshal(configBytes, &config); err != nil {
		return nil, fmt.Errorf("failed to parse HTTP action config: %w", err)
	}

	// Execute HTTP request
	result, err := a.executeHTTP(ctx, config, input.Context)
	if err != nil {
		return nil, err
	}

	return NewActionOutput(result), nil
}

// executeHTTP executes an HTTP request
func (a *HTTPAction) executeHTTP(ctx context.Context, config HTTPActionConfig, execContext map[string]interface{}) (*HTTPActionResult, error) {
	// Validate method
	method := strings.ToUpper(config.Method)
	if method == "" {
		method = "GET"
	}
	if !isValidHTTPMethod(method) {
		return nil, fmt.Errorf("invalid HTTP method: %s", method)
	}

	// Default timeout
	timeout := 30 * time.Second
	if config.Timeout > 0 {
		timeout = time.Duration(config.Timeout) * time.Second
	}

	// Create context with timeout
	timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Configure HTTP client
	client := &http.Client{
		Timeout: timeout,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if !config.FollowRedirects && len(via) > 0 {
				return http.ErrUseLastResponse
			}
			if len(via) >= 10 {
				return fmt.Errorf("too many redirects")
			}
			return nil
		},
	}

	// Interpolate URL
	url := InterpolateString(config.URL, execContext)
	if url == "" {
		return nil, fmt.Errorf("URL is required")
	}

	// Validate URL to prevent SSRF attacks
	if a.urlValidator != nil {
		if err := a.urlValidator.ValidateURL(url); err != nil {
			return nil, fmt.Errorf("SSRF protection: %w", err)
		}
	}

	// Prepare request body
	var bodyReader io.Reader
	if config.Body != nil && len(config.Body) > 0 {
		interpolatedBody := InterpolateJSON(config.Body, execContext)
		bodyBytes, err := json.Marshal(interpolatedBody)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(bodyBytes)
	}

	// Create request
	req, err := http.NewRequestWithContext(timeoutCtx, method, url, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set default content type for requests with body
	if bodyReader != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	// Set custom headers
	for key, value := range config.Headers {
		req.Header.Set(key, InterpolateString(value, execContext))
	}

	// Apply authentication
	if err := a.applyAuth(req, config.Auth, execContext); err != nil {
		return nil, fmt.Errorf("failed to apply authentication: %w", err)
	}

	// Execute request
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Parse response body
	var parsedBody interface{}
	contentType := resp.Header.Get("Content-Type")
	if strings.Contains(contentType, "application/json") {
		if err := json.Unmarshal(respBody, &parsedBody); err != nil {
			// If JSON parsing fails, use raw string
			parsedBody = string(respBody)
		}
	} else {
		parsedBody = string(respBody)
	}

	// Build response headers map
	respHeaders := make(map[string]string)
	for key := range resp.Header {
		respHeaders[key] = resp.Header.Get(key)
	}

	return &HTTPActionResult{
		StatusCode: resp.StatusCode,
		Headers:    respHeaders,
		Body:       parsedBody,
	}, nil
}

// applyAuth applies authentication to the request
func (a *HTTPAction) applyAuth(req *http.Request, auth *HTTPAuth, context map[string]interface{}) error {
	if auth == nil {
		return nil
	}

	switch strings.ToLower(auth.Type) {
	case "basic":
		username := InterpolateString(auth.Username, context)
		password := InterpolateString(auth.Password, context)
		req.SetBasicAuth(username, password)

	case "bearer":
		token := InterpolateString(auth.Token, context)
		if token == "" {
			return fmt.Errorf("bearer token is required")
		}
		req.Header.Set("Authorization", "Bearer "+token)

	case "api_key":
		apiKey := InterpolateString(auth.APIKey, context)
		if apiKey == "" {
			return fmt.Errorf("API key is required")
		}
		header := auth.Header
		if header == "" {
			header = "X-API-Key"
		}
		req.Header.Set(header, apiKey)

	case "":
		// No auth
		return nil

	default:
		return fmt.Errorf("unsupported auth type: %s", auth.Type)
	}

	return nil
}

// isValidHTTPMethod checks if the method is valid
func isValidHTTPMethod(method string) bool {
	validMethods := []string{"GET", "POST", "PUT", "DELETE", "PATCH", "HEAD", "OPTIONS"}
	for _, m := range validMethods {
		if method == m {
			return true
		}
	}
	return false
}

// Legacy function for backward compatibility
func ExecuteHTTP(ctx context.Context, config HTTPActionConfig, context map[string]interface{}) (*HTTPActionResult, error) {
	action := NewHTTPAction()
	input := NewActionInput(config, context)
	output, err := action.Execute(ctx, input)
	if err != nil {
		return nil, err
	}
	result, ok := output.Data.(*HTTPActionResult)
	if !ok {
		return nil, fmt.Errorf("unexpected output type")
	}
	return result, nil
}
