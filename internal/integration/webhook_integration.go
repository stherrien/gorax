package integration

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log/slog"
	"maps"
	"strconv"
	"text/template"
	"time"

	inthttp "github.com/gorax/gorax/internal/integration/http"
)

// WebhookIntegration provides outbound webhook integration capabilities.
type WebhookIntegration struct {
	*BaseIntegration
	client *inthttp.Client
	logger *slog.Logger
}

// WebhookConfig holds configuration for webhook execution.
type WebhookConfig struct {
	URL              string            `json:"url"`
	Method           string            `json:"method,omitempty"` // Default: POST
	Headers          map[string]string `json:"headers,omitempty"`
	PayloadTemplate  string            `json:"payload_template,omitempty"`
	Payload          any               `json:"payload,omitempty"`
	ContentType      string            `json:"content_type,omitempty"`     // Default: application/json
	SignatureHeader  string            `json:"signature_header,omitempty"` // e.g., "X-Signature-256"
	SignatureSecret  string            `json:"signature_secret,omitempty"` // Secret for HMAC signature
	Timeout          int               `json:"timeout,omitempty"`          // seconds
	RetryOnFailure   bool              `json:"retry_on_failure,omitempty"`
	IncludeTimestamp bool              `json:"include_timestamp,omitempty"`
}

// WebhookResult extends Result with webhook-specific fields.
type WebhookResult struct {
	*Result
	WebhookID    string `json:"webhook_id,omitempty"`
	DeliveryID   string `json:"delivery_id,omitempty"`
	Signature    string `json:"signature,omitempty"`
	AttemptCount int    `json:"attempt_count,omitempty"`
}

// NewWebhookIntegration creates a new webhook integration.
func NewWebhookIntegration(logger *slog.Logger) *WebhookIntegration {
	if logger == nil {
		logger = slog.Default()
	}

	base := NewBaseIntegration("webhook", TypeWebhook)
	base.SetMetadata(&Metadata{
		Name:        "webhook",
		DisplayName: "Outbound Webhook",
		Description: "Send outbound webhook calls to external endpoints",
		Version:     "1.0.0",
		Category:    "networking",
	})
	base.SetSchema(&Schema{
		ConfigSpec: map[string]FieldSpec{
			"url": {
				Name:        "url",
				Type:        FieldTypeString,
				Description: "Webhook destination URL",
				Required:    true,
			},
			"method": {
				Name:        "method",
				Type:        FieldTypeString,
				Description: "HTTP method (default: POST)",
				Required:    false,
				Options:     []string{"POST", "PUT", "PATCH"},
			},
			"headers": {
				Name:        "headers",
				Type:        FieldTypeObject,
				Description: "Custom headers to include",
				Required:    false,
			},
			"payload_template": {
				Name:        "payload_template",
				Type:        FieldTypeString,
				Description: "Go template for payload generation",
				Required:    false,
			},
			"signature_header": {
				Name:        "signature_header",
				Type:        FieldTypeString,
				Description: "Header name for HMAC signature",
				Required:    false,
			},
			"signature_secret": {
				Name:        "signature_secret",
				Type:        FieldTypeSecret,
				Description: "Secret key for HMAC signature",
				Required:    false,
				Sensitive:   true,
			},
		},
		InputSpec: map[string]FieldSpec{
			"event": {
				Name:        "event",
				Type:        FieldTypeObject,
				Description: "Event data to send",
				Required:    true,
			},
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
			"delivery_id": {
				Name:        "delivery_id",
				Type:        FieldTypeString,
				Description: "Unique delivery identifier",
			},
			"signature": {
				Name:        "signature",
				Type:        FieldTypeString,
				Description: "HMAC signature (if configured)",
			},
		},
	})

	client := inthttp.NewClient(
		inthttp.WithTimeout(30*time.Second),
		inthttp.WithLogger(logger),
		inthttp.WithRetryConfig(inthttp.DefaultRetryConfig()),
	)

	return &WebhookIntegration{
		BaseIntegration: base,
		client:          client,
		logger:          logger,
	}
}

// Execute sends a webhook to the configured endpoint.
func (w *WebhookIntegration) Execute(ctx context.Context, config *Config, params JSONMap) (*Result, error) {
	start := time.Now()

	// Extract webhook-specific config
	webhookConfig, err := w.parseConfig(config)
	if err != nil {
		return NewErrorResult(err, "INVALID_CONFIG", time.Since(start).Milliseconds()), err
	}

	// Generate delivery ID
	deliveryID := generateDeliveryID()

	// Build payload
	payload, err := w.buildPayload(webhookConfig, params)
	if err != nil {
		return NewErrorResult(err, "PAYLOAD_BUILD_FAILED", time.Since(start).Milliseconds()), err
	}

	// Serialize payload
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return NewErrorResult(err, "PAYLOAD_SERIALIZE_FAILED", time.Since(start).Milliseconds()), err
	}

	// Calculate signature if configured
	var signature string
	if webhookConfig.SignatureSecret != "" {
		signature = w.calculateSignature(payloadBytes, webhookConfig.SignatureSecret)
	}

	// Build request
	req := w.buildRequest(webhookConfig, payloadBytes, signature, deliveryID)

	// Execute request
	resp, err := w.client.Do(ctx, req)
	if err != nil {
		w.logger.Error("webhook delivery failed",
			"error", err,
			"url", webhookConfig.URL,
			"delivery_id", deliveryID,
		)
		return NewErrorResult(err, "DELIVERY_FAILED", time.Since(start).Milliseconds()), err
	}

	// Check response status
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		err := NewHTTPError(resp.StatusCode, resp.Status, string(resp.Body))
		return NewErrorResult(err, "WEBHOOK_REJECTED", time.Since(start).Milliseconds()), err
	}

	result := &Result{
		Success:    true,
		StatusCode: resp.StatusCode,
		Duration:   time.Since(start).Milliseconds(),
		ExecutedAt: time.Now().UTC(),
		Data: JSONMap{
			"delivery_id": deliveryID,
			"status_code": resp.StatusCode,
			"signature":   signature,
		},
	}

	w.logger.Info("webhook delivered successfully",
		"url", webhookConfig.URL,
		"delivery_id", deliveryID,
		"status_code", resp.StatusCode,
		"duration_ms", result.Duration,
	)

	return result, nil
}

// Validate validates the webhook configuration.
func (w *WebhookIntegration) Validate(config *Config) error {
	if err := w.BaseIntegration.ValidateConfig(config); err != nil {
		return err
	}

	// Validate URL
	url, ok := config.Settings.GetString("url")
	if !ok || url == "" {
		return NewValidationError("url", "webhook URL is required", nil)
	}

	// Validate signature configuration
	signatureHeader, hasHeader := config.Settings.GetString("signature_header")
	signatureSecret, hasSecret := config.Settings.GetString("signature_secret")
	if (signatureHeader != "" && !hasSecret) || (signatureSecret != "" && !hasHeader) {
		return NewValidationError("signature", "both signature_header and signature_secret must be provided together", nil)
	}

	return nil
}

// parseConfig extracts webhook configuration from the integration config.
func (w *WebhookIntegration) parseConfig(config *Config) (*WebhookConfig, error) {
	webhookConfig := &WebhookConfig{
		Method:      "POST",
		ContentType: "application/json",
	}

	if url, ok := config.Settings.GetString("url"); ok {
		webhookConfig.URL = url
	}

	if method, ok := config.Settings.GetString("method"); ok {
		webhookConfig.Method = method
	}

	if headers, ok := config.Settings.Get("headers"); ok {
		if headersMap, ok := headers.(map[string]any); ok {
			webhookConfig.Headers = make(map[string]string)
			for k, v := range headersMap {
				if strVal, ok := v.(string); ok {
					webhookConfig.Headers[k] = strVal
				}
			}
		}
	}

	if payloadTemplate, ok := config.Settings.GetString("payload_template"); ok {
		webhookConfig.PayloadTemplate = payloadTemplate
	}

	if payload, ok := config.Settings.Get("payload"); ok {
		webhookConfig.Payload = payload
	}

	if contentType, ok := config.Settings.GetString("content_type"); ok {
		webhookConfig.ContentType = contentType
	}

	if signatureHeader, ok := config.Settings.GetString("signature_header"); ok {
		webhookConfig.SignatureHeader = signatureHeader
	}

	if signatureSecret, ok := config.Settings.GetString("signature_secret"); ok {
		webhookConfig.SignatureSecret = signatureSecret
	}

	if timeout, ok := config.Settings.GetInt("timeout"); ok {
		webhookConfig.Timeout = timeout
	}

	if retryOnFailure, ok := config.Settings.GetBool("retry_on_failure"); ok {
		webhookConfig.RetryOnFailure = retryOnFailure
	}

	if includeTimestamp, ok := config.Settings.GetBool("include_timestamp"); ok {
		webhookConfig.IncludeTimestamp = includeTimestamp
	}

	return webhookConfig, nil
}

// buildPayload builds the webhook payload.
func (w *WebhookIntegration) buildPayload(config *WebhookConfig, params JSONMap) (any, error) {
	// Use provided payload if available
	if config.Payload != nil {
		return config.Payload, nil
	}

	// Use template if available
	if config.PayloadTemplate != "" {
		return w.executeTemplate(config.PayloadTemplate, params)
	}

	// Use params directly
	if params != nil {
		if event, ok := params.Get("event"); ok {
			return event, nil
		}
		return params, nil
	}

	return JSONMap{}, nil
}

// executeTemplate executes a Go template string.
func (w *WebhookIntegration) executeTemplate(tmplStr string, data JSONMap) (any, error) {
	tmpl, err := template.New("").Parse(tmplStr)
	if err != nil {
		return nil, fmt.Errorf("parsing template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return nil, fmt.Errorf("executing template: %w", err)
	}

	// Try to parse as JSON
	var result any
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		// Return as raw string
		return buf.String(), nil
	}

	return result, nil
}

// buildRequest builds the HTTP request for the webhook.
func (w *WebhookIntegration) buildRequest(config *WebhookConfig, payload []byte, signature, deliveryID string) *inthttp.Request {
	headers := make(map[string]string)

	// Copy custom headers
	maps.Copy(headers, config.Headers)

	// Set standard headers
	headers["Content-Type"] = config.ContentType
	headers["X-Webhook-Delivery-ID"] = deliveryID

	// Add signature header if configured
	if config.SignatureHeader != "" && signature != "" {
		headers[config.SignatureHeader] = "sha256=" + signature
	}

	// Add timestamp if configured
	if config.IncludeTimestamp {
		headers["X-Webhook-Timestamp"] = strconv.FormatInt(time.Now().Unix(), 10)
	}

	return &inthttp.Request{
		Method:  config.Method,
		URL:     config.URL,
		Headers: headers,
		RawBody: payload,
	}
}

// calculateSignature calculates the HMAC-SHA256 signature.
func (w *WebhookIntegration) calculateSignature(payload []byte, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	return hex.EncodeToString(mac.Sum(nil))
}

// generateDeliveryID generates a unique delivery ID.
func generateDeliveryID() string {
	// Simple timestamp-based ID
	// In production, use UUID or similar
	return fmt.Sprintf("whd_%d", time.Now().UnixNano())
}

// VerifyWebhookSignature verifies an incoming webhook signature.
// This is a utility function for receiving webhooks.
func VerifyWebhookSignature(payload []byte, signature, secret string) bool {
	// Remove "sha256=" prefix if present
	if len(signature) > 7 && signature[:7] == "sha256=" {
		signature = signature[7:]
	}

	expectedSignature := calculateHMACSHA256(payload, secret)
	return hmac.Equal([]byte(signature), []byte(expectedSignature))
}

// calculateHMACSHA256 calculates the HMAC-SHA256 signature.
func calculateHMACSHA256(payload []byte, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	return hex.EncodeToString(mac.Sum(nil))
}

// NewWebhookFromConfig creates a WebhookIntegration with pre-configured settings.
func NewWebhookFromConfig(url, secret string, opts ...WebhookOption) (*WebhookIntegration, *Config) {
	integration := NewWebhookIntegration(nil)

	config := &Config{
		Name:    "webhook",
		Type:    TypeWebhook,
		Enabled: true,
		Settings: JSONMap{
			"url":              url,
			"method":           "POST",
			"signature_header": "X-Signature-256",
			"signature_secret": secret,
		},
	}

	for _, opt := range opts {
		opt(config)
	}

	return integration, config
}

// WebhookOption configures a webhook.
type WebhookOption func(*Config)

// WithWebhookHeaders sets custom headers.
func WithWebhookHeaders(headers map[string]string) WebhookOption {
	return func(c *Config) {
		c.Settings["headers"] = headers
	}
}

// WithWebhookMethod sets the HTTP method.
func WithWebhookMethod(method string) WebhookOption {
	return func(c *Config) {
		c.Settings["method"] = method
	}
}

// WithWebhookPayloadTemplate sets a payload template.
func WithWebhookPayloadTemplate(template string) WebhookOption {
	return func(c *Config) {
		c.Settings["payload_template"] = template
	}
}

// WithWebhookTimestamp enables timestamp header.
func WithWebhookTimestamp() WebhookOption {
	return func(c *Config) {
		c.Settings["include_timestamp"] = true
	}
}

// SendWebhook is a convenience function to send a webhook.
func SendWebhook(ctx context.Context, url string, payload any, opts ...WebhookOption) (*Result, error) {
	integration, config := NewWebhookFromConfig(url, "", opts...)
	params := JSONMap{"event": payload}
	return integration.Execute(ctx, config, params)
}

// SendSignedWebhook sends a webhook with HMAC signature.
func SendSignedWebhook(ctx context.Context, url, secret string, payload any, opts ...WebhookOption) (*Result, error) {
	integration, config := NewWebhookFromConfig(url, secret, opts...)
	params := JSONMap{"event": payload}
	return integration.Execute(ctx, config, params)
}
