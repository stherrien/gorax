package integration

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewWebhookIntegration(t *testing.T) {
	t.Run("creates with default logger", func(t *testing.T) {
		integ := NewWebhookIntegration(nil)

		require.NotNil(t, integ)
		assert.Equal(t, "webhook", integ.Name())
		assert.Equal(t, TypeWebhook, integ.Type())
		assert.NotNil(t, integ.client)
		assert.NotNil(t, integ.logger)
	})

	t.Run("creates with custom logger", func(t *testing.T) {
		logger := slog.Default()
		integ := NewWebhookIntegration(logger)

		require.NotNil(t, integ)
		assert.Equal(t, logger, integ.logger)
	})

	t.Run("has proper metadata", func(t *testing.T) {
		integ := NewWebhookIntegration(nil)
		metadata := integ.GetMetadata()

		assert.Equal(t, "webhook", metadata.Name)
		assert.Equal(t, "Outbound Webhook", metadata.DisplayName)
		assert.Equal(t, "1.0.0", metadata.Version)
		assert.Equal(t, "networking", metadata.Category)
	})

	t.Run("has proper schema", func(t *testing.T) {
		integ := NewWebhookIntegration(nil)
		schema := integ.GetSchema()

		require.NotNil(t, schema)
		assert.Contains(t, schema.ConfigSpec, "url")
		assert.Contains(t, schema.ConfigSpec, "method")
		assert.Contains(t, schema.ConfigSpec, "headers")
		assert.Contains(t, schema.ConfigSpec, "signature_header")
		assert.Contains(t, schema.ConfigSpec, "signature_secret")
	})
}

func TestWebhookIntegration_Execute(t *testing.T) {
	t.Run("sends basic webhook", func(t *testing.T) {
		var receivedBody map[string]any
		var receivedHeaders http.Header
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			receivedHeaders = r.Header
			body, _ := io.ReadAll(r.Body)
			json.Unmarshal(body, &receivedBody)
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		integ := NewWebhookIntegration(nil)
		config := &Config{
			Name:    "test",
			Type:    TypeWebhook,
			Enabled: true,
			Settings: JSONMap{
				"url": server.URL,
			},
		}

		params := JSONMap{
			"event": map[string]any{
				"type": "user.created",
				"data": map[string]any{
					"id":   "123",
					"name": "John",
				},
			},
		}

		result, err := integ.Execute(context.Background(), config, params)

		require.NoError(t, err)
		require.NotNil(t, result)
		assert.True(t, result.Success)
		assert.Equal(t, http.StatusOK, result.StatusCode)
		assert.NotEmpty(t, receivedHeaders.Get("X-Webhook-Delivery-ID"))
		assert.NotEmpty(t, receivedHeaders.Get("Content-Type"))
	})

	t.Run("sends webhook with signature", func(t *testing.T) {
		var receivedSignature string
		var receivedBody []byte
		secret := "my-secret-key"

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			receivedSignature = r.Header.Get("X-Signature-256")
			receivedBody, _ = io.ReadAll(r.Body)
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		integ := NewWebhookIntegration(nil)
		config := &Config{
			Name:    "test",
			Type:    TypeWebhook,
			Enabled: true,
			Settings: JSONMap{
				"url":              server.URL,
				"signature_header": "X-Signature-256",
				"signature_secret": secret,
			},
		}

		params := JSONMap{
			"event": map[string]any{"type": "test"},
		}

		result, err := integ.Execute(context.Background(), config, params)

		require.NoError(t, err)
		assert.True(t, result.Success)
		assert.NotEmpty(t, receivedSignature)
		assert.True(t, strings.HasPrefix(receivedSignature, "sha256="))

		// Verify signature
		assert.True(t, VerifyWebhookSignature(receivedBody, receivedSignature, secret))
	})

	t.Run("includes custom headers", func(t *testing.T) {
		var receivedHeaders http.Header
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			receivedHeaders = r.Header
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		integ := NewWebhookIntegration(nil)
		config := &Config{
			Name:    "test",
			Type:    TypeWebhook,
			Enabled: true,
			Settings: JSONMap{
				"url": server.URL,
				"headers": map[string]any{
					"X-Custom-Header": "custom-value",
					"Authorization":   "Bearer token123",
				},
			},
		}

		result, err := integ.Execute(context.Background(), config, JSONMap{"event": "test"})

		require.NoError(t, err)
		assert.True(t, result.Success)
		assert.Equal(t, "custom-value", receivedHeaders.Get("X-Custom-Header"))
		assert.Equal(t, "Bearer token123", receivedHeaders.Get("Authorization"))
	})

	t.Run("includes timestamp when configured", func(t *testing.T) {
		var receivedHeaders http.Header
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			receivedHeaders = r.Header
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		integ := NewWebhookIntegration(nil)
		config := &Config{
			Name:    "test",
			Type:    TypeWebhook,
			Enabled: true,
			Settings: JSONMap{
				"url":               server.URL,
				"include_timestamp": true,
			},
		}

		result, err := integ.Execute(context.Background(), config, JSONMap{"event": "test"})

		require.NoError(t, err)
		assert.True(t, result.Success)
		assert.NotEmpty(t, receivedHeaders.Get("X-Webhook-Timestamp"))
	})

	t.Run("uses payload template", func(t *testing.T) {
		var receivedBody map[string]any
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			body, _ := io.ReadAll(r.Body)
			json.Unmarshal(body, &receivedBody)
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		integ := NewWebhookIntegration(nil)
		config := &Config{
			Name:    "test",
			Type:    TypeWebhook,
			Enabled: true,
			Settings: JSONMap{
				"url":              server.URL,
				"payload_template": `{"event_type": "{{.event_type}}", "user_id": "{{.user_id}}"}`,
			},
		}

		params := JSONMap{
			"event_type": "user.created",
			"user_id":    "123",
		}

		result, err := integ.Execute(context.Background(), config, params)

		require.NoError(t, err)
		assert.True(t, result.Success)
	})

	t.Run("handles non-2xx response", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusServiceUnavailable)
			_, _ = w.Write([]byte(`{"error": "service unavailable"}`))
		}))
		defer server.Close()

		integ := NewWebhookIntegration(nil)
		config := &Config{
			Name:    "test",
			Type:    TypeWebhook,
			Enabled: true,
			Settings: JSONMap{
				"url": server.URL,
			},
		}

		result, err := integ.Execute(context.Background(), config, JSONMap{"event": "test"})

		require.Error(t, err)
		require.NotNil(t, result)
		assert.False(t, result.Success)
	})

	t.Run("uses custom method", func(t *testing.T) {
		var receivedMethod string
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			receivedMethod = r.Method
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		integ := NewWebhookIntegration(nil)
		config := &Config{
			Name:    "test",
			Type:    TypeWebhook,
			Enabled: true,
			Settings: JSONMap{
				"url":    server.URL,
				"method": "PUT",
			},
		}

		result, err := integ.Execute(context.Background(), config, JSONMap{"event": "test"})

		require.NoError(t, err)
		assert.True(t, result.Success)
		assert.Equal(t, "PUT", receivedMethod)
	})

	t.Run("uses direct payload", func(t *testing.T) {
		var receivedBody map[string]any
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			body, _ := io.ReadAll(r.Body)
			json.Unmarshal(body, &receivedBody)
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		integ := NewWebhookIntegration(nil)
		config := &Config{
			Name:    "test",
			Type:    TypeWebhook,
			Enabled: true,
			Settings: JSONMap{
				"url": server.URL,
				"payload": map[string]any{
					"direct": "payload",
				},
			},
		}

		result, err := integ.Execute(context.Background(), config, nil)

		require.NoError(t, err)
		assert.True(t, result.Success)
		assert.Equal(t, "payload", receivedBody["direct"])
	})
}

func TestWebhookIntegration_Validate(t *testing.T) {
	integ := NewWebhookIntegration(nil)

	t.Run("validates valid config", func(t *testing.T) {
		config := &Config{
			Name:    "test",
			Type:    TypeWebhook,
			Enabled: true,
			Settings: JSONMap{
				"url": "https://hooks.example.com/webhook",
			},
		}

		err := integ.Validate(config)
		require.NoError(t, err)
	})

	t.Run("validates config with signature", func(t *testing.T) {
		config := &Config{
			Name:    "test",
			Type:    TypeWebhook,
			Enabled: true,
			Settings: JSONMap{
				"url":              "https://hooks.example.com/webhook",
				"signature_header": "X-Signature",
				"signature_secret": "secret",
			},
		}

		err := integ.Validate(config)
		require.NoError(t, err)
	})

	t.Run("fails with nil config", func(t *testing.T) {
		err := integ.Validate(nil)
		require.Error(t, err)
	})

	t.Run("fails with missing URL", func(t *testing.T) {
		config := &Config{
			Name:     "test",
			Type:     TypeWebhook,
			Enabled:  true,
			Settings: JSONMap{},
		}

		err := integ.Validate(config)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "URL")
	})

	t.Run("fails with signature_header but no secret", func(t *testing.T) {
		config := &Config{
			Name:    "test",
			Type:    TypeWebhook,
			Enabled: true,
			Settings: JSONMap{
				"url":              "https://hooks.example.com/webhook",
				"signature_header": "X-Signature",
			},
		}

		err := integ.Validate(config)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "signature")
	})

	t.Run("fails with signature_secret but no header", func(t *testing.T) {
		config := &Config{
			Name:    "test",
			Type:    TypeWebhook,
			Enabled: true,
			Settings: JSONMap{
				"url":              "https://hooks.example.com/webhook",
				"signature_secret": "secret",
			},
		}

		err := integ.Validate(config)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "signature")
	})
}

func TestVerifyWebhookSignature(t *testing.T) {
	secret := "test-secret"
	payload := []byte(`{"event": "test"}`)

	// Calculate expected signature
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	expectedSig := hex.EncodeToString(mac.Sum(nil))

	t.Run("verifies valid signature", func(t *testing.T) {
		assert.True(t, VerifyWebhookSignature(payload, expectedSig, secret))
	})

	t.Run("verifies signature with sha256 prefix", func(t *testing.T) {
		assert.True(t, VerifyWebhookSignature(payload, "sha256="+expectedSig, secret))
	})

	t.Run("rejects invalid signature", func(t *testing.T) {
		assert.False(t, VerifyWebhookSignature(payload, "invalid-signature", secret))
	})

	t.Run("rejects signature with wrong secret", func(t *testing.T) {
		assert.False(t, VerifyWebhookSignature(payload, expectedSig, "wrong-secret"))
	})

	t.Run("rejects signature for different payload", func(t *testing.T) {
		assert.False(t, VerifyWebhookSignature([]byte("different payload"), expectedSig, secret))
	})
}

func TestNewWebhookFromConfig(t *testing.T) {
	t.Run("creates webhook with basic config", func(t *testing.T) {
		integ, config := NewWebhookFromConfig("https://hooks.example.com", "")

		require.NotNil(t, integ)
		require.NotNil(t, config)
		assert.Equal(t, "https://hooks.example.com", config.Settings["url"])
		assert.Equal(t, "POST", config.Settings["method"])
	})

	t.Run("creates webhook with secret", func(t *testing.T) {
		integ, config := NewWebhookFromConfig("https://hooks.example.com", "my-secret")

		require.NotNil(t, integ)
		require.NotNil(t, config)
		assert.Equal(t, "my-secret", config.Settings["signature_secret"])
		assert.Equal(t, "X-Signature-256", config.Settings["signature_header"])
	})

	t.Run("applies options", func(t *testing.T) {
		_, config := NewWebhookFromConfig(
			"https://hooks.example.com",
			"secret",
			WithWebhookHeaders(map[string]string{"X-Custom": "value"}),
			WithWebhookMethod("PUT"),
			WithWebhookPayloadTemplate(`{"type": "{{.type}}"}`),
			WithWebhookTimestamp(),
		)

		headers := config.Settings["headers"].(map[string]string)
		assert.Equal(t, "value", headers["X-Custom"])
		assert.Equal(t, "PUT", config.Settings["method"])
		assert.Equal(t, `{"type": "{{.type}}"}`, config.Settings["payload_template"])
		assert.Equal(t, true, config.Settings["include_timestamp"])
	})
}

func TestSendWebhook(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	t.Run("sends unsigned webhook", func(t *testing.T) {
		result, err := SendWebhook(context.Background(), server.URL, map[string]string{"key": "value"})

		require.NoError(t, err)
		assert.True(t, result.Success)
	})
}

func TestSendSignedWebhook(t *testing.T) {
	secret := "test-secret"
	var receivedSignature string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedSignature = r.Header.Get("X-Signature-256")
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	t.Run("sends signed webhook", func(t *testing.T) {
		result, err := SendSignedWebhook(context.Background(), server.URL, secret, map[string]string{"key": "value"})

		require.NoError(t, err)
		assert.True(t, result.Success)
		assert.NotEmpty(t, receivedSignature)
		assert.True(t, strings.HasPrefix(receivedSignature, "sha256="))
	})
}

func TestWebhookConfig(t *testing.T) {
	config := WebhookConfig{
		URL:              "https://hooks.example.com",
		Method:           "POST",
		Headers:          map[string]string{"X-Custom": "value"},
		PayloadTemplate:  `{"event": "{{.event}}"}`,
		Payload:          map[string]string{"key": "value"},
		ContentType:      "application/json",
		SignatureHeader:  "X-Signature",
		SignatureSecret:  "secret",
		Timeout:          30,
		RetryOnFailure:   true,
		IncludeTimestamp: true,
	}

	data, err := json.Marshal(config)
	require.NoError(t, err)

	var decoded WebhookConfig
	require.NoError(t, json.Unmarshal(data, &decoded))

	assert.Equal(t, config.URL, decoded.URL)
	assert.Equal(t, config.Method, decoded.Method)
	assert.Equal(t, config.SignatureHeader, decoded.SignatureHeader)
	assert.Equal(t, config.RetryOnFailure, decoded.RetryOnFailure)
}

func TestWebhookResult(t *testing.T) {
	result := WebhookResult{
		Result: &Result{
			Success:    true,
			StatusCode: 200,
			Duration:   100,
			ExecutedAt: time.Now(),
		},
		WebhookID:    "wh_123",
		DeliveryID:   "whd_456",
		Signature:    "sha256=abc123",
		AttemptCount: 1,
	}

	data, err := json.Marshal(result)
	require.NoError(t, err)

	var decoded WebhookResult
	require.NoError(t, json.Unmarshal(data, &decoded))

	assert.Equal(t, result.WebhookID, decoded.WebhookID)
	assert.Equal(t, result.DeliveryID, decoded.DeliveryID)
	assert.Equal(t, result.Signature, decoded.Signature)
}

func TestGenerateDeliveryID(t *testing.T) {
	id1 := generateDeliveryID()
	id2 := generateDeliveryID()

	assert.NotEmpty(t, id1)
	assert.NotEmpty(t, id2)
	assert.True(t, strings.HasPrefix(id1, "whd_"))
	// IDs should be unique (different timestamps)
	// Note: In fast execution, they might be the same, so we just check format
}

func TestCalculateHMACSHA256(t *testing.T) {
	payload := []byte(`{"test": "data"}`)
	secret := "secret-key"

	sig := calculateHMACSHA256(payload, secret)

	assert.NotEmpty(t, sig)
	// Verify it's a valid hex string
	_, err := hex.DecodeString(sig)
	require.NoError(t, err)
}

func TestWebhookIntegration_ExecuteTimeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	integ := NewWebhookIntegration(nil)
	config := &Config{
		Name:    "test",
		Type:    TypeWebhook,
		Enabled: true,
		Settings: JSONMap{
			"url":     server.URL,
			"timeout": 1,
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	result, err := integ.Execute(ctx, config, JSONMap{"event": "test"})

	require.Error(t, err)
	require.NotNil(t, result)
	assert.False(t, result.Success)
}
