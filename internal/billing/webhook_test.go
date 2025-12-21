package billing

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewWebhookNotifier(t *testing.T) {
	notifier := NewWebhookNotifier()
	assert.NotNil(t, notifier)
}

func TestWebhookNotifier_RegisterWebhook(t *testing.T) {
	notifier := NewWebhookNotifier()
	ctx := context.Background()

	config := WebhookConfig{
		TenantID:  "tenant-1",
		URL:       "https://example.com/webhook",
		Secret:    "secret123",
		Events:    []string{"quota.threshold", "execution.completed"},
		Active:    true,
		CreatedAt: time.Now(),
	}

	err := notifier.RegisterWebhook(ctx, config)
	assert.NoError(t, err)

	// Verify webhook is registered
	webhooks := notifier.GetWebhooks(ctx, "tenant-1")
	assert.Len(t, webhooks, 1)
	assert.Equal(t, config.URL, webhooks[0].URL)
}

func TestWebhookNotifier_UnregisterWebhook(t *testing.T) {
	notifier := NewWebhookNotifier()
	ctx := context.Background()

	config := WebhookConfig{
		TenantID: "tenant-1",
		URL:      "https://example.com/webhook",
		Secret:   "secret123",
		Events:   []string{"quota.threshold"},
		Active:   true,
	}

	err := notifier.RegisterWebhook(ctx, config)
	require.NoError(t, err)

	err = notifier.UnregisterWebhook(ctx, "tenant-1", config.URL)
	assert.NoError(t, err)

	webhooks := notifier.GetWebhooks(ctx, "tenant-1")
	assert.Len(t, webhooks, 0)
}

func TestWebhookNotifier_SendQuotaThreshold(t *testing.T) {
	tests := []struct {
		name           string
		threshold      int
		current        int64
		limit          int64
		expectWebhook  bool
		webhookActive  bool
		expectedStatus int
	}{
		{
			name:           "80% threshold reached",
			threshold:      80,
			current:        80,
			limit:          100,
			expectWebhook:  true,
			webhookActive:  true,
			expectedStatus: http.StatusOK,
		},
		{
			name:          "webhook inactive",
			threshold:     80,
			current:       80,
			limit:         100,
			expectWebhook: false,
			webhookActive: false,
		},
		{
			name:           "100% threshold reached",
			threshold:      100,
			current:        100,
			limit:          100,
			expectWebhook:  true,
			webhookActive:  true,
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test server with thread-safe variables
			var called atomic.Bool
			var receivedPayload WebhookPayload
			var payloadMu sync.Mutex

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				called.Store(true)

				// Verify headers
				assert.NotEmpty(t, r.Header.Get("X-Webhook-Signature"))
				assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

				// Decode payload
				payloadMu.Lock()
				err := json.NewDecoder(r.Body).Decode(&receivedPayload)
				payloadMu.Unlock()
				assert.NoError(t, err)

				w.WriteHeader(tt.expectedStatus)
			}))
			defer server.Close()

			notifier := NewWebhookNotifier()
			ctx := context.Background()

			// Register webhook
			config := WebhookConfig{
				TenantID: "tenant-1",
				URL:      server.URL,
				Secret:   "secret123",
				Events:   []string{"quota.threshold"},
				Active:   tt.webhookActive,
			}

			err := notifier.RegisterWebhook(ctx, config)
			require.NoError(t, err)

			// Send notification
			err = notifier.SendQuotaThreshold(ctx, "tenant-1", tt.threshold, tt.current, tt.limit)
			assert.NoError(t, err)

			// Wait a bit for async delivery
			time.Sleep(100 * time.Millisecond)

			// Verify webhook was called if expected
			assert.Equal(t, tt.expectWebhook, called.Load())

			if tt.expectWebhook {
				payloadMu.Lock()
				assert.Equal(t, "quota.threshold", receivedPayload.Event)
				assert.Equal(t, "tenant-1", receivedPayload.TenantID)

				data := receivedPayload.Data.(map[string]interface{})
				assert.Equal(t, float64(tt.threshold), data["threshold_percent"])
				assert.Equal(t, float64(tt.current), data["current"])
				assert.Equal(t, float64(tt.limit), data["limit"])
				payloadMu.Unlock()
			}
		})
	}
}

func TestWebhookNotifier_SendExecutionCompleted(t *testing.T) {
	var called atomic.Bool
	var receivedPayload WebhookPayload
	var payloadMu sync.Mutex

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called.Store(true)
		payloadMu.Lock()
		err := json.NewDecoder(r.Body).Decode(&receivedPayload)
		payloadMu.Unlock()
		assert.NoError(t, err)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	notifier := NewWebhookNotifier()
	ctx := context.Background()

	config := WebhookConfig{
		TenantID: "tenant-1",
		URL:      server.URL,
		Secret:   "secret123",
		Events:   []string{"execution.completed"},
		Active:   true,
	}

	err := notifier.RegisterWebhook(ctx, config)
	require.NoError(t, err)

	payload := ExecutionCompletedPayload{
		ExecutionID: "exec-1",
		WorkflowID:  "workflow-1",
		Status:      "success",
		Duration:    1500,
		StepCount:   5,
	}

	err = notifier.SendExecutionCompleted(ctx, "tenant-1", payload)
	assert.NoError(t, err)

	time.Sleep(100 * time.Millisecond)

	assert.True(t, called.Load())
	payloadMu.Lock()
	assert.Equal(t, "execution.completed", receivedPayload.Event)
	payloadMu.Unlock()
}

func TestWebhookNotifier_RetryLogic(t *testing.T) {
	var attempts atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		count := attempts.Add(1)
		if count < 3 {
			w.WriteHeader(http.StatusInternalServerError)
		} else {
			w.WriteHeader(http.StatusOK)
		}
	}))
	defer server.Close()

	notifier := NewWebhookNotifier()
	notifier.retryBackoff = 10 * time.Millisecond // Fast retries for testing
	ctx := context.Background()

	config := WebhookConfig{
		TenantID:   "tenant-1",
		URL:        server.URL,
		Secret:     "secret123",
		Events:     []string{"quota.threshold"},
		Active:     true,
		MaxRetries: 3,
	}

	err := notifier.RegisterWebhook(ctx, config)
	require.NoError(t, err)

	err = notifier.SendQuotaThreshold(ctx, "tenant-1", 80, 80, 100)
	assert.NoError(t, err)

	// Wait for retries (exponential backoff: 10ms + 20ms + 30ms + buffer)
	time.Sleep(200 * time.Millisecond)

	// Should have retried until success
	assert.GreaterOrEqual(t, int(attempts.Load()), 3)
}

func TestWebhookNotifier_SignatureVerification(t *testing.T) {
	var receivedSignature atomic.Value

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedSignature.Store(r.Header.Get("X-Webhook-Signature"))
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	notifier := NewWebhookNotifier()
	ctx := context.Background()

	secret := "my-secret-key"
	config := WebhookConfig{
		TenantID: "tenant-1",
		URL:      server.URL,
		Secret:   secret,
		Events:   []string{"quota.threshold"},
		Active:   true,
	}

	err := notifier.RegisterWebhook(ctx, config)
	require.NoError(t, err)

	err = notifier.SendQuotaThreshold(ctx, "tenant-1", 80, 80, 100)
	assert.NoError(t, err)

	time.Sleep(100 * time.Millisecond)

	// Verify signature was sent
	sig := receivedSignature.Load().(string)
	assert.NotEmpty(t, sig)

	// Verify signature format (should be sha256=<hex>)
	assert.Contains(t, sig, "sha256=")
}

func TestWebhookNotifier_DeliveryLog(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	notifier := NewWebhookNotifier()
	ctx := context.Background()

	config := WebhookConfig{
		TenantID: "tenant-1",
		URL:      server.URL,
		Secret:   "secret123",
		Events:   []string{"quota.threshold"},
		Active:   true,
	}

	err := notifier.RegisterWebhook(ctx, config)
	require.NoError(t, err)

	err = notifier.SendQuotaThreshold(ctx, "tenant-1", 80, 80, 100)
	assert.NoError(t, err)

	time.Sleep(100 * time.Millisecond)

	// Get delivery logs
	logs := notifier.GetDeliveryLogs(ctx, "tenant-1", 10)
	assert.NotEmpty(t, logs)
	assert.Equal(t, "quota.threshold", logs[0].Event)
	assert.Equal(t, http.StatusOK, logs[0].StatusCode)
}

func TestWebhookNotifier_FilterByEvent(t *testing.T) {
	var quotaCalled atomic.Bool
	var executionCalled atomic.Bool

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var payload WebhookPayload
		json.NewDecoder(r.Body).Decode(&payload)

		if payload.Event == "quota.threshold" {
			quotaCalled.Store(true)
		} else if payload.Event == "execution.completed" {
			executionCalled.Store(true)
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	notifier := NewWebhookNotifier()
	ctx := context.Background()

	// Register webhook that only listens to quota events
	config := WebhookConfig{
		TenantID: "tenant-1",
		URL:      server.URL,
		Secret:   "secret123",
		Events:   []string{"quota.threshold"},
		Active:   true,
	}

	err := notifier.RegisterWebhook(ctx, config)
	require.NoError(t, err)

	// Send quota event (should be delivered)
	err = notifier.SendQuotaThreshold(ctx, "tenant-1", 80, 80, 100)
	assert.NoError(t, err)

	time.Sleep(100 * time.Millisecond)

	// Send execution event (should NOT be delivered)
	execPayload := ExecutionCompletedPayload{
		ExecutionID: "exec-1",
		WorkflowID:  "workflow-1",
		Status:      "success",
	}
	err = notifier.SendExecutionCompleted(ctx, "tenant-1", execPayload)
	assert.NoError(t, err)

	time.Sleep(100 * time.Millisecond)

	assert.True(t, quotaCalled.Load(), "quota webhook should have been called")
	assert.False(t, executionCalled.Load(), "execution webhook should NOT have been called")
}

func TestWebhookConfig_Validation(t *testing.T) {
	tests := []struct {
		name    string
		config  WebhookConfig
		wantErr bool
	}{
		{
			name: "valid config",
			config: WebhookConfig{
				TenantID: "tenant-1",
				URL:      "https://example.com/webhook",
				Secret:   "secret123",
				Events:   []string{"quota.threshold"},
				Active:   true,
			},
			wantErr: false,
		},
		{
			name: "missing tenant ID",
			config: WebhookConfig{
				URL:    "https://example.com/webhook",
				Secret: "secret123",
				Events: []string{"quota.threshold"},
			},
			wantErr: true,
		},
		{
			name: "invalid URL",
			config: WebhookConfig{
				TenantID: "tenant-1",
				URL:      "not-a-url",
				Secret:   "secret123",
				Events:   []string{"quota.threshold"},
			},
			wantErr: true,
		},
		{
			name: "no events",
			config: WebhookConfig{
				TenantID: "tenant-1",
				URL:      "https://example.com/webhook",
				Secret:   "secret123",
				Events:   []string{},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestComputeSignature(t *testing.T) {
	secret := "my-secret"
	payload := []byte(`{"event":"test","data":{}}`)

	sig1 := computeSignature(payload, secret)
	sig2 := computeSignature(payload, secret)

	// Same input should produce same signature
	assert.Equal(t, sig1, sig2)

	// Should start with sha256=
	assert.Contains(t, sig1, "sha256=")

	// Different payload should produce different signature
	differentPayload := []byte(`{"event":"different","data":{}}`)
	sig3 := computeSignature(differentPayload, secret)
	assert.NotEqual(t, sig1, sig3)
}
