package billing

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"sync"
	"time"
)

var (
	// ErrInvalidConfig is returned when webhook config is invalid
	ErrInvalidConfig = errors.New("invalid webhook configuration")
	// ErrWebhookNotFound is returned when webhook is not found
	ErrWebhookNotFound = errors.New("webhook not found")
)

// WebhookConfig holds webhook configuration
type WebhookConfig struct {
	TenantID   string    `json:"tenant_id"`
	URL        string    `json:"url"`
	Secret     string    `json:"secret"`
	Events     []string  `json:"events"`
	Active     bool      `json:"active"`
	MaxRetries int       `json:"max_retries"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// Validate validates webhook configuration
func (c *WebhookConfig) Validate() error {
	if c.TenantID == "" {
		return fmt.Errorf("%w: tenant_id is required", ErrInvalidConfig)
	}

	if c.URL == "" {
		return fmt.Errorf("%w: url is required", ErrInvalidConfig)
	}

	// Validate URL format
	parsedURL, err := url.Parse(c.URL)
	if err != nil {
		return fmt.Errorf("%w: invalid url format: %v", ErrInvalidConfig, err)
	}
	if parsedURL.Scheme == "" || parsedURL.Host == "" {
		return fmt.Errorf("%w: url must have scheme and host", ErrInvalidConfig)
	}

	if len(c.Events) == 0 {
		return fmt.Errorf("%w: at least one event is required", ErrInvalidConfig)
	}

	return nil
}

// WebhookPayload represents the payload sent to webhooks
type WebhookPayload struct {
	Event     string      `json:"event"`
	TenantID  string      `json:"tenant_id"`
	Timestamp time.Time   `json:"timestamp"`
	Data      interface{} `json:"data"`
}

// ExecutionCompletedPayload represents execution completion data
type ExecutionCompletedPayload struct {
	ExecutionID string `json:"execution_id"`
	WorkflowID  string `json:"workflow_id"`
	Status      string `json:"status"`
	Duration    int64  `json:"duration_ms"`
	StepCount   int    `json:"step_count"`
}

// DeliveryLog represents a webhook delivery attempt
type DeliveryLog struct {
	ID         string    `json:"id"`
	TenantID   string    `json:"tenant_id"`
	Event      string    `json:"event"`
	URL        string    `json:"url"`
	StatusCode int       `json:"status_code"`
	Response   string    `json:"response"`
	Attempts   int       `json:"attempts"`
	CreatedAt  time.Time `json:"created_at"`
	Success    bool      `json:"success"`
}

// WebhookNotifier handles webhook notifications
type WebhookNotifier struct {
	webhooks     map[string][]WebhookConfig
	deliveryLogs map[string][]DeliveryLog
	mu           sync.RWMutex
	client       *http.Client
	logger       *slog.Logger
	retryBackoff time.Duration // For testing
}

// NewWebhookNotifier creates a new webhook notifier
func NewWebhookNotifier() *WebhookNotifier {
	return &WebhookNotifier{
		webhooks:     make(map[string][]WebhookConfig),
		deliveryLogs: make(map[string][]DeliveryLog),
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		logger:       slog.Default(),
		retryBackoff: time.Second, // Default 1 second backoff
	}
}

// RegisterWebhook registers a webhook for a tenant
func (n *WebhookNotifier) RegisterWebhook(ctx context.Context, config WebhookConfig) error {
	if err := config.Validate(); err != nil {
		return err
	}

	// Set defaults
	if config.MaxRetries == 0 {
		config.MaxRetries = 3
	}
	if config.CreatedAt.IsZero() {
		config.CreatedAt = time.Now()
	}
	config.UpdatedAt = time.Now()

	n.mu.Lock()
	defer n.mu.Unlock()

	n.webhooks[config.TenantID] = append(n.webhooks[config.TenantID], config)

	return nil
}

// UnregisterWebhook removes a webhook for a tenant
func (n *WebhookNotifier) UnregisterWebhook(ctx context.Context, tenantID, url string) error {
	n.mu.Lock()
	defer n.mu.Unlock()

	webhooks := n.webhooks[tenantID]
	filtered := make([]WebhookConfig, 0)

	for _, wh := range webhooks {
		if wh.URL != url {
			filtered = append(filtered, wh)
		}
	}

	if len(filtered) == len(webhooks) {
		return ErrWebhookNotFound
	}

	n.webhooks[tenantID] = filtered

	return nil
}

// GetWebhooks returns all webhooks for a tenant
func (n *WebhookNotifier) GetWebhooks(ctx context.Context, tenantID string) []WebhookConfig {
	n.mu.RLock()
	defer n.mu.RUnlock()

	webhooks := n.webhooks[tenantID]
	result := make([]WebhookConfig, len(webhooks))
	copy(result, webhooks)

	return result
}

// SendQuotaThreshold sends a quota threshold notification
func (n *WebhookNotifier) SendQuotaThreshold(ctx context.Context, tenantID string, thresholdPercent int, current, limit int64) error {
	payload := WebhookPayload{
		Event:     "quota.threshold",
		TenantID:  tenantID,
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"threshold_percent": thresholdPercent,
			"current":           current,
			"limit":             limit,
			"percent_used":      (float64(current) / float64(limit)) * 100,
		},
	}

	return n.send(ctx, tenantID, payload)
}

// SendExecutionCompleted sends an execution completed notification
func (n *WebhookNotifier) SendExecutionCompleted(ctx context.Context, tenantID string, execPayload ExecutionCompletedPayload) error {
	payload := WebhookPayload{
		Event:     "execution.completed",
		TenantID:  tenantID,
		Timestamp: time.Now(),
		Data:      execPayload,
	}

	return n.send(ctx, tenantID, payload)
}

// send delivers payload to registered webhooks
func (n *WebhookNotifier) send(ctx context.Context, tenantID string, payload WebhookPayload) error {
	webhooks := n.GetWebhooks(ctx, tenantID)

	for _, webhook := range webhooks {
		if !webhook.Active {
			continue
		}

		// Check if webhook is subscribed to this event
		if !n.isSubscribed(webhook, payload.Event) {
			continue
		}

		// Deliver asynchronously
		go n.deliver(ctx, webhook, payload)
	}

	return nil
}

// isSubscribed checks if webhook is subscribed to an event
func (n *WebhookNotifier) isSubscribed(webhook WebhookConfig, event string) bool {
	for _, e := range webhook.Events {
		if e == event {
			return true
		}
	}
	return false
}

// deliver delivers payload to a webhook with retry logic
func (n *WebhookNotifier) deliver(ctx context.Context, webhook WebhookConfig, payload WebhookPayload) {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		n.logger.Error("failed to marshal webhook payload",
			"error", err,
			"tenant_id", webhook.TenantID,
		)
		return
	}

	var lastStatusCode int
	var lastResponse string
	var success bool

	for attempt := 1; attempt <= webhook.MaxRetries; attempt++ {
		statusCode, response, err := n.deliverOnce(ctx, webhook, payloadBytes)
		lastStatusCode = statusCode
		lastResponse = response

		if err == nil && statusCode >= 200 && statusCode < 300 {
			success = true
			break
		}

		if attempt < webhook.MaxRetries {
			// Exponential backoff
			backoff := time.Duration(attempt) * n.retryBackoff
			time.Sleep(backoff)
		}

		n.logger.Warn("webhook delivery attempt failed",
			"attempt", attempt,
			"tenant_id", webhook.TenantID,
			"url", webhook.URL,
			"status", statusCode,
			"error", err,
		)
	}

	// Log delivery
	log := DeliveryLog{
		ID:         fmt.Sprintf("%d", time.Now().UnixNano()),
		TenantID:   webhook.TenantID,
		Event:      payload.Event,
		URL:        webhook.URL,
		StatusCode: lastStatusCode,
		Response:   lastResponse,
		Attempts:   webhook.MaxRetries,
		CreatedAt:  time.Now(),
		Success:    success,
	}

	n.mu.Lock()
	n.deliveryLogs[webhook.TenantID] = append(n.deliveryLogs[webhook.TenantID], log)
	n.mu.Unlock()
}

// deliverOnce attempts a single webhook delivery
func (n *WebhookNotifier) deliverOnce(ctx context.Context, webhook WebhookConfig, payloadBytes []byte) (int, string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, webhook.URL, bytes.NewReader(payloadBytes))
	if err != nil {
		return 0, "", fmt.Errorf("failed to create request: %w", err)
	}

	// Add headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "gorax-webhook/1.0")

	// Add signature
	signature := computeSignature(payloadBytes, webhook.Secret)
	req.Header.Set("X-Webhook-Signature", signature)

	resp, err := n.client.Do(req)
	if err != nil {
		return 0, "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Read response (limited)
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return resp.StatusCode, "", fmt.Errorf("failed to read response body: %w", err)
	}
	response := string(bodyBytes)

	if len(response) > 1000 {
		response = response[:1000]
	}

	return resp.StatusCode, response, nil
}

// GetDeliveryLogs returns delivery logs for a tenant
func (n *WebhookNotifier) GetDeliveryLogs(ctx context.Context, tenantID string, limit int) []DeliveryLog {
	n.mu.RLock()
	defer n.mu.RUnlock()

	logs := n.deliveryLogs[tenantID]
	if len(logs) == 0 {
		return []DeliveryLog{}
	}

	// Return most recent logs
	start := len(logs) - limit
	if start < 0 {
		start = 0
	}

	result := make([]DeliveryLog, len(logs)-start)
	copy(result, logs[start:])

	// Reverse to show most recent first
	for i, j := 0, len(result)-1; i < j; i, j = i+1, j-1 {
		result[i], result[j] = result[j], result[i]
	}

	return result
}

// computeSignature computes HMAC-SHA256 signature
func computeSignature(payload []byte, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	signature := hex.EncodeToString(mac.Sum(nil))
	return fmt.Sprintf("sha256=%s", signature)
}
