package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"slices"
	"strings"
	"text/template"
	"time"

	inthttp "github.com/gorax/gorax/internal/integration/http"
)

// HTTPIntegration provides HTTP request integration capabilities.
type HTTPIntegration struct {
	*BaseIntegration
	client *inthttp.Client
	logger *slog.Logger
}

// HTTPIntegrationConfig holds configuration for HTTP integration execution.
type HTTPIntegrationConfig struct {
	URL          string            `json:"url"`
	Method       string            `json:"method"`
	Headers      map[string]string `json:"headers,omitempty"`
	QueryParams  map[string]string `json:"query_params,omitempty"`
	Body         any               `json:"body,omitempty"`
	BodyTemplate string            `json:"body_template,omitempty"`
	Timeout      int               `json:"timeout,omitempty"`       // seconds
	ResponseType string            `json:"response_type,omitempty"` // json, text, binary
	SuccessCodes []int             `json:"success_codes,omitempty"` // Custom success status codes
	ExtractPath  string            `json:"extract_path,omitempty"`  // JSONPath for response extraction
}

// NewHTTPIntegration creates a new HTTP integration.
func NewHTTPIntegration(logger *slog.Logger) *HTTPIntegration {
	if logger == nil {
		logger = slog.Default()
	}

	base := NewBaseIntegration("http", TypeHTTP)
	base.SetMetadata(&Metadata{
		Name:        "http",
		DisplayName: "HTTP Request",
		Description: "Execute HTTP requests to external APIs",
		Version:     "1.0.0",
		Category:    "networking",
	})
	base.SetSchema(&Schema{
		ConfigSpec: map[string]FieldSpec{
			"url": {
				Name:        "url",
				Type:        FieldTypeString,
				Description: "Target URL for the request",
				Required:    true,
			},
			"method": {
				Name:        "method",
				Type:        FieldTypeString,
				Description: "HTTP method (GET, POST, PUT, DELETE, PATCH)",
				Required:    true,
				Options:     []string{"GET", "POST", "PUT", "DELETE", "PATCH", "HEAD", "OPTIONS"},
			},
			"headers": {
				Name:        "headers",
				Type:        FieldTypeObject,
				Description: "Request headers",
				Required:    false,
			},
			"body": {
				Name:        "body",
				Type:        FieldTypeObject,
				Description: "Request body (for POST, PUT, PATCH)",
				Required:    false,
			},
			"timeout": {
				Name:        "timeout",
				Type:        FieldTypeInteger,
				Description: "Request timeout in seconds",
				Required:    false,
			},
		},
		InputSpec: map[string]FieldSpec{
			"context": {
				Name:        "context",
				Type:        FieldTypeObject,
				Description: "Workflow context for template substitution",
				Required:    false,
			},
		},
		OutputSpec: map[string]FieldSpec{
			"status_code": {
				Name:        "status_code",
				Type:        FieldTypeInteger,
				Description: "HTTP response status code",
			},
			"headers": {
				Name:        "headers",
				Type:        FieldTypeObject,
				Description: "Response headers",
			},
			"body": {
				Name:        "body",
				Type:        FieldTypeObject,
				Description: "Response body (parsed JSON or text)",
			},
		},
	})

	client := inthttp.NewClient(
		inthttp.WithTimeout(30*time.Second),
		inthttp.WithLogger(logger),
		inthttp.WithRetryConfig(inthttp.DefaultRetryConfig()),
	)

	return &HTTPIntegration{
		BaseIntegration: base,
		client:          client,
		logger:          logger,
	}
}

// Execute performs an HTTP request based on the configuration.
func (h *HTTPIntegration) Execute(ctx context.Context, config *Config, params JSONMap) (*Result, error) {
	start := time.Now()

	// Extract HTTP-specific config
	httpConfig, err := h.parseConfig(config, params)
	if err != nil {
		return NewErrorResult(err, "INVALID_CONFIG", time.Since(start).Milliseconds()), err
	}

	// Build request
	req, err := h.buildRequest(httpConfig, params)
	if err != nil {
		return NewErrorResult(err, "REQUEST_BUILD_FAILED", time.Since(start).Milliseconds()), err
	}

	// Execute request
	resp, err := h.client.Do(ctx, req)
	if err != nil {
		h.logger.Error("HTTP request failed",
			"error", err,
			"url", httpConfig.URL,
			"method", httpConfig.Method,
		)
		return NewErrorResult(err, "REQUEST_FAILED", time.Since(start).Milliseconds()), err
	}

	// Check response status
	if !h.isSuccessStatus(resp.StatusCode, httpConfig.SuccessCodes) {
		err := NewHTTPError(resp.StatusCode, resp.Status, string(resp.Body))
		return NewErrorResult(err, "HTTP_ERROR", time.Since(start).Milliseconds()), err
	}

	// Parse response
	result := h.buildResult(resp, httpConfig, start)

	h.logger.Info("HTTP request completed",
		"url", httpConfig.URL,
		"method", httpConfig.Method,
		"status_code", resp.StatusCode,
		"duration_ms", result.Duration,
	)

	return result, nil
}

// Validate validates the integration configuration.
func (h *HTTPIntegration) Validate(config *Config) error {
	if err := h.BaseIntegration.ValidateConfig(config); err != nil {
		return err
	}

	// Validate URL
	urlStr, ok := config.Settings.GetString("url")
	if !ok || urlStr == "" {
		return NewValidationError("url", "URL is required", nil)
	}

	if !strings.HasPrefix(urlStr, "{{") { // Skip validation for templates
		if _, err := url.Parse(urlStr); err != nil {
			return NewValidationError("url", "invalid URL format", urlStr)
		}
	}

	// Validate method
	method, ok := config.Settings.GetString("method")
	if !ok || method == "" {
		return NewValidationError("method", "HTTP method is required", nil)
	}

	validMethods := map[string]bool{
		"GET": true, "POST": true, "PUT": true, "DELETE": true,
		"PATCH": true, "HEAD": true, "OPTIONS": true,
	}
	if !validMethods[strings.ToUpper(method)] {
		return NewValidationError("method", "invalid HTTP method", method)
	}

	return nil
}

// parseConfig extracts HTTP configuration from the integration config.
func (h *HTTPIntegration) parseConfig(config *Config, params JSONMap) (*HTTPIntegrationConfig, error) {
	httpConfig := &HTTPIntegrationConfig{
		ResponseType: "json",
		SuccessCodes: []int{200, 201, 202, 204},
	}

	if urlStr, ok := config.Settings.GetString("url"); ok {
		httpConfig.URL = urlStr
	}

	if method, ok := config.Settings.GetString("method"); ok {
		httpConfig.Method = strings.ToUpper(method)
	}

	if headers, ok := config.Settings.Get("headers"); ok {
		if headersMap, ok := headers.(map[string]any); ok {
			httpConfig.Headers = make(map[string]string)
			for k, v := range headersMap {
				if strVal, ok := v.(string); ok {
					httpConfig.Headers[k] = strVal
				}
			}
		}
	}

	if queryParams, ok := config.Settings.Get("query_params"); ok {
		if paramsMap, ok := queryParams.(map[string]any); ok {
			httpConfig.QueryParams = make(map[string]string)
			for k, v := range paramsMap {
				if strVal, ok := v.(string); ok {
					httpConfig.QueryParams[k] = strVal
				}
			}
		}
	}

	if body, ok := config.Settings.Get("body"); ok {
		httpConfig.Body = body
	}

	if bodyTemplate, ok := config.Settings.GetString("body_template"); ok {
		httpConfig.BodyTemplate = bodyTemplate
	}

	if timeout, ok := config.Settings.GetInt("timeout"); ok {
		httpConfig.Timeout = timeout
	}

	if responseType, ok := config.Settings.GetString("response_type"); ok {
		httpConfig.ResponseType = responseType
	}

	if successCodes, ok := config.Settings.Get("success_codes"); ok {
		if codesArr, ok := successCodes.([]any); ok {
			httpConfig.SuccessCodes = make([]int, 0, len(codesArr))
			for _, code := range codesArr {
				if intCode, ok := code.(float64); ok {
					httpConfig.SuccessCodes = append(httpConfig.SuccessCodes, int(intCode))
				}
			}
		}
	}

	if extractPath, ok := config.Settings.GetString("extract_path"); ok {
		httpConfig.ExtractPath = extractPath
	}

	// Process templates
	if err := h.processTemplates(httpConfig, params); err != nil {
		return nil, fmt.Errorf("processing templates: %w", err)
	}

	return httpConfig, nil
}

// processTemplates processes template strings in the configuration.
func (h *HTTPIntegration) processTemplates(config *HTTPIntegrationConfig, params JSONMap) error {
	// Process URL template
	url, err := h.executeTemplate(config.URL, params)
	if err != nil {
		return fmt.Errorf("URL template: %w", err)
	}
	config.URL = url

	// Process header templates
	for key, value := range config.Headers {
		processed, err := h.executeTemplate(value, params)
		if err != nil {
			return fmt.Errorf("header %s template: %w", key, err)
		}
		config.Headers[key] = processed
	}

	// Process query param templates
	for key, value := range config.QueryParams {
		processed, err := h.executeTemplate(value, params)
		if err != nil {
			return fmt.Errorf("query param %s template: %w", key, err)
		}
		config.QueryParams[key] = processed
	}

	// Process body template
	if config.BodyTemplate != "" {
		processed, err := h.executeTemplate(config.BodyTemplate, params)
		if err != nil {
			return fmt.Errorf("body template: %w", err)
		}

		// Try to parse as JSON
		var body any
		if err := json.Unmarshal([]byte(processed), &body); err != nil {
			// Use as raw string
			config.Body = processed
		} else {
			config.Body = body
		}
	}

	return nil
}

// executeTemplate executes a Go template string.
func (h *HTTPIntegration) executeTemplate(tmplStr string, data JSONMap) (string, error) {
	if !strings.Contains(tmplStr, "{{") {
		return tmplStr, nil
	}

	tmpl, err := template.New("").Parse(tmplStr)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}

// buildRequest builds an HTTP request from the configuration.
func (h *HTTPIntegration) buildRequest(config *HTTPIntegrationConfig, _ JSONMap) (*inthttp.Request, error) {
	req := &inthttp.Request{
		Method:      config.Method,
		URL:         config.URL,
		Headers:     config.Headers,
		QueryParams: config.QueryParams,
	}

	// Set body
	if config.Body != nil {
		req.Body = config.Body
	}

	return req, nil
}

// isSuccessStatus checks if the status code is considered successful.
func (h *HTTPIntegration) isSuccessStatus(statusCode int, successCodes []int) bool {
	if len(successCodes) == 0 {
		return statusCode >= 200 && statusCode < 300
	}

	return slices.Contains(successCodes, statusCode)
}

// buildResult builds the integration result from the HTTP response.
func (h *HTTPIntegration) buildResult(resp *inthttp.Response, config *HTTPIntegrationConfig, start time.Time) *Result {
	data := make(JSONMap)
	data["status_code"] = resp.StatusCode
	data["status"] = resp.Status

	// Convert headers
	headers := make(map[string]string)
	for key, values := range resp.Headers {
		if len(values) > 0 {
			headers[key] = values[0]
		}
	}
	data["headers"] = headers

	// Parse response body
	switch config.ResponseType {
	case "json":
		var body any
		if err := json.Unmarshal(resp.Body, &body); err != nil {
			data["body"] = string(resp.Body)
		} else {
			data["body"] = body
		}
	case "text":
		data["body"] = string(resp.Body)
	case "binary":
		data["body"] = resp.Body
	default:
		// Try JSON first, fall back to text
		var body any
		if err := json.Unmarshal(resp.Body, &body); err != nil {
			data["body"] = string(resp.Body)
		} else {
			data["body"] = body
		}
	}

	return &Result{
		Success:    true,
		Data:       data,
		StatusCode: resp.StatusCode,
		Duration:   time.Since(start).Milliseconds(),
		ExecutedAt: time.Now().UTC(),
	}
}

// NewHTTPIntegrationFromRequest creates an HTTPIntegration configured for a single request.
func NewHTTPIntegrationFromRequest(method, targetURL string, opts ...HTTPRequestOption) (*HTTPIntegration, *Config) {
	integration := NewHTTPIntegration(nil)

	config := &Config{
		Name:    "http-request",
		Type:    TypeHTTP,
		Enabled: true,
		Settings: JSONMap{
			"method": method,
			"url":    targetURL,
		},
	}

	for _, opt := range opts {
		opt(config)
	}

	return integration, config
}

// HTTPRequestOption configures an HTTP request.
type HTTPRequestOption func(*Config)

// WithHTTPHeaders sets request headers.
func WithHTTPHeaders(headers map[string]string) HTTPRequestOption {
	return func(c *Config) {
		c.Settings["headers"] = headers
	}
}

// WithHTTPBody sets the request body.
func WithHTTPBody(body any) HTTPRequestOption {
	return func(c *Config) {
		c.Settings["body"] = body
	}
}

// WithHTTPQueryParams sets query parameters.
func WithHTTPQueryParams(params map[string]string) HTTPRequestOption {
	return func(c *Config) {
		c.Settings["query_params"] = params
	}
}

// WithHTTPTimeout sets the request timeout.
func WithHTTPTimeout(seconds int) HTTPRequestOption {
	return func(c *Config) {
		c.Settings["timeout"] = seconds
	}
}

// QuickHTTP performs a quick HTTP request.
func QuickHTTP(ctx context.Context, method, url string, opts ...HTTPRequestOption) (*Result, error) {
	integration, config := NewHTTPIntegrationFromRequest(method, url, opts...)
	return integration.Execute(ctx, config, nil)
}

// QuickGet performs a quick GET request.
func QuickGet(ctx context.Context, url string, opts ...HTTPRequestOption) (*Result, error) {
	return QuickHTTP(ctx, http.MethodGet, url, opts...)
}

// QuickPost performs a quick POST request.
func QuickPost(ctx context.Context, url string, body any, opts ...HTTPRequestOption) (*Result, error) {
	opts = append(opts, WithHTTPBody(body))
	return QuickHTTP(ctx, http.MethodPost, url, opts...)
}
