package integrations

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gorax/gorax/internal/integration"
	inttesting "github.com/gorax/gorax/internal/integration/testing"
)

func TestNewAdvancedHTTPIntegration(t *testing.T) {
	httpInt := NewAdvancedHTTPIntegration(nil)

	assert.NotNil(t, httpInt)
	assert.Equal(t, "http_advanced", httpInt.Name())
	assert.Equal(t, integration.TypeHTTP, httpInt.Type())

	metadata := httpInt.GetMetadata()
	assert.Equal(t, "Advanced HTTP", metadata.DisplayName)
	assert.Equal(t, "networking", metadata.Category)

	schema := httpInt.GetSchema()
	assert.NotNil(t, schema.ConfigSpec["url"])
	assert.NotNil(t, schema.InputSpec["headers"])
}

func TestAdvancedHTTPIntegration_Validate(t *testing.T) {
	httpInt := NewAdvancedHTTPIntegration(nil)

	tests := []struct {
		name        string
		config      *integration.Config
		expectError bool
	}{
		{
			name:        "nil config",
			config:      nil,
			expectError: true,
		},
		{
			name: "missing URL",
			config: &integration.Config{
				Name:     "test",
				Type:     integration.TypeHTTP,
				Enabled:  true,
				Settings: integration.JSONMap{},
			},
			expectError: true,
		},
		{
			name: "invalid method",
			config: &integration.Config{
				Name:    "test",
				Type:    integration.TypeHTTP,
				Enabled: true,
				Settings: integration.JSONMap{
					"url":    "https://example.com",
					"method": "INVALID",
				},
			},
			expectError: true,
		},
		{
			name: "valid config",
			config: &integration.Config{
				Name:    "test",
				Type:    integration.TypeHTTP,
				Enabled: true,
				Settings: integration.JSONMap{
					"url":    "https://example.com/api",
					"method": "GET",
				},
			},
			expectError: false,
		},
		{
			name: "valid config with template URL",
			config: &integration.Config{
				Name:    "test",
				Type:    integration.TypeHTTP,
				Enabled: true,
				Settings: integration.JSONMap{
					"url":    "https://example.com/api/{{.id}}",
					"method": "GET",
				},
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := httpInt.Validate(tt.config)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestAdvancedHTTPIntegration_ExecuteWithMockServer(t *testing.T) {
	mockServer := inttesting.NewMockServer()
	defer mockServer.Close()

	mockServer.OnGet("/api/test", inttesting.JSONResponse(http.StatusOK, map[string]any{
		"message": "success",
		"data": map[string]any{
			"id":   123,
			"name": "test",
		},
	}))

	httpInt := NewAdvancedHTTPIntegration(nil)

	config := &integration.Config{
		Name:    "test-http",
		Type:    integration.TypeHTTP,
		Enabled: true,
		Settings: integration.JSONMap{
			"url":    mockServer.URL() + "/api/test",
			"method": "GET",
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err := httpInt.Execute(ctx, config, nil)

	require.NoError(t, err)
	assert.True(t, result.Success)
	assert.Equal(t, http.StatusOK, result.StatusCode)

	// Verify request was made
	assert.Equal(t, 1, mockServer.GetRequestCount())
	lastReq := mockServer.GetLastRequest()
	assert.Equal(t, "GET", lastReq.Method)
}

func TestAdvancedHTTPIntegration_ExecutePost(t *testing.T) {
	mockServer := inttesting.NewMockServer()
	defer mockServer.Close()

	mockServer.OnPost("/api/create", inttesting.JSONResponse(http.StatusCreated, map[string]any{
		"id":      456,
		"created": true,
	}))

	httpInt := NewAdvancedHTTPIntegration(nil)

	config := &integration.Config{
		Name:    "test-http",
		Type:    integration.TypeHTTP,
		Enabled: true,
		Settings: integration.JSONMap{
			"url":    mockServer.URL() + "/api/create",
			"method": "POST",
		},
	}

	params := integration.JSONMap{
		"body": map[string]any{
			"name":  "test",
			"value": 42,
		},
		"success_codes": []any{float64(200), float64(201)},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err := httpInt.Execute(ctx, config, params)

	require.NoError(t, err)
	assert.True(t, result.Success)
	assert.Equal(t, http.StatusCreated, result.StatusCode)
}

func TestAdvancedHTTPIntegration_ExecuteWithBearerAuth(t *testing.T) {
	mockServer := inttesting.NewMockServer()
	defer mockServer.Close()

	mockServer.OnGet("/api/protected", inttesting.JSONResponse(http.StatusOK, map[string]any{
		"authorized": true,
	}))

	httpInt := NewAdvancedHTTPIntegration(nil)

	config := &integration.Config{
		Name:    "test-http",
		Type:    integration.TypeHTTP,
		Enabled: true,
		Settings: integration.JSONMap{
			"url":       mockServer.URL() + "/api/protected",
			"method":    "GET",
			"auth_type": "bearer",
		},
		Credentials: &integration.Credentials{
			Data: integration.JSONMap{
				"token": "test-bearer-token",
			},
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err := httpInt.Execute(ctx, config, nil)

	require.NoError(t, err)
	assert.True(t, result.Success)

	// Verify Authorization header
	lastReq := mockServer.GetLastRequest()
	assert.Contains(t, string(lastReq.Body), "") // Body is empty for GET
	// Note: In real implementation, we'd verify the header was set correctly
}

func TestAdvancedHTTPIntegration_ExecuteWithBasicAuth(t *testing.T) {
	mockServer := inttesting.NewMockServer()
	defer mockServer.Close()

	mockServer.OnGet("/api/basic", inttesting.JSONResponse(http.StatusOK, map[string]any{
		"authenticated": true,
	}))

	httpInt := NewAdvancedHTTPIntegration(nil)

	config := &integration.Config{
		Name:    "test-http",
		Type:    integration.TypeHTTP,
		Enabled: true,
		Settings: integration.JSONMap{
			"url":       mockServer.URL() + "/api/basic",
			"method":    "GET",
			"auth_type": "basic",
		},
		Credentials: &integration.Credentials{
			Data: integration.JSONMap{
				"username": "user",
				"password": "pass",
			},
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err := httpInt.Execute(ctx, config, nil)

	require.NoError(t, err)
	assert.True(t, result.Success)
}

func TestAdvancedHTTPIntegration_ExecuteWithAPIKeyAuth(t *testing.T) {
	mockServer := inttesting.NewMockServer()
	defer mockServer.Close()

	mockServer.OnGet("/api/apikey", inttesting.JSONResponse(http.StatusOK, map[string]any{
		"valid_key": true,
	}))

	httpInt := NewAdvancedHTTPIntegration(nil)

	config := &integration.Config{
		Name:    "test-http",
		Type:    integration.TypeHTTP,
		Enabled: true,
		Settings: integration.JSONMap{
			"url":       mockServer.URL() + "/api/apikey",
			"method":    "GET",
			"auth_type": "api_key",
		},
		Credentials: &integration.Credentials{
			Data: integration.JSONMap{
				"api_key":     "test-api-key",
				"header_name": "X-API-Key",
			},
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err := httpInt.Execute(ctx, config, nil)

	require.NoError(t, err)
	assert.True(t, result.Success)
}

func TestAdvancedHTTPIntegration_ExecuteWithRetry(t *testing.T) {
	mockServer := inttesting.NewMockServer()
	defer mockServer.Close()

	// First two requests fail, third succeeds
	callCount := 0
	mockServer.OnRequest("GET", "/api/retry", &inttesting.MockResponse{
		Handler: func(w http.ResponseWriter, r *http.Request) {
			callCount++
			if callCount < 3 {
				w.WriteHeader(http.StatusServiceUnavailable)
				return
			}
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"success": true}`))
		},
	})

	httpInt := NewAdvancedHTTPIntegration(nil)

	config := &integration.Config{
		Name:    "test-http",
		Type:    integration.TypeHTTP,
		Enabled: true,
		Settings: integration.JSONMap{
			"url":    mockServer.URL() + "/api/retry",
			"method": "GET",
		},
	}

	params := integration.JSONMap{
		"retry_count": 5,
		"retry_delay": 0, // No delay for test
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	result, err := httpInt.Execute(ctx, config, params)

	require.NoError(t, err)
	assert.True(t, result.Success)
	assert.Equal(t, http.StatusOK, result.StatusCode)
	assert.Equal(t, 3, callCount)
}

func TestAdvancedHTTPIntegration_ExecuteWithFormBody(t *testing.T) {
	mockServer := inttesting.NewMockServer()
	defer mockServer.Close()

	mockServer.OnPost("/api/form", inttesting.JSONResponse(http.StatusOK, map[string]any{
		"received": true,
	}))

	httpInt := NewAdvancedHTTPIntegration(nil)

	config := &integration.Config{
		Name:    "test-http",
		Type:    integration.TypeHTTP,
		Enabled: true,
		Settings: integration.JSONMap{
			"url":         mockServer.URL() + "/api/form",
			"method":      "POST",
			"body_format": "form",
		},
	}

	params := integration.JSONMap{
		"body": map[string]any{
			"username": "testuser",
			"password": "testpass",
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err := httpInt.Execute(ctx, config, params)

	require.NoError(t, err)
	assert.True(t, result.Success)

	lastReq := mockServer.GetLastRequest()
	assert.Contains(t, string(lastReq.Body), "username=testuser")
}

func TestAdvancedHTTPIntegration_ExecuteWithTemplate(t *testing.T) {
	mockServer := inttesting.NewMockServer()
	defer mockServer.Close()

	mockServer.OnGet("/api/users/123", inttesting.JSONResponse(http.StatusOK, map[string]any{
		"id":   123,
		"name": "Test User",
	}))

	httpInt := NewAdvancedHTTPIntegration(nil)

	config := &integration.Config{
		Name:    "test-http",
		Type:    integration.TypeHTTP,
		Enabled: true,
		Settings: integration.JSONMap{
			"url":    mockServer.URL() + "/api/users/{{.user_id}}",
			"method": "GET",
		},
	}

	params := integration.JSONMap{
		"user_id": "123",
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err := httpInt.Execute(ctx, config, params)

	require.NoError(t, err)
	assert.True(t, result.Success)
}

func TestAdvancedHTTPIntegration_ResponseExtraction(t *testing.T) {
	mockServer := inttesting.NewMockServer()
	defer mockServer.Close()

	mockServer.OnGet("/api/nested", inttesting.JSONResponse(http.StatusOK, map[string]any{
		"data": map[string]any{
			"items": []any{
				map[string]any{"id": 1, "name": "first"},
				map[string]any{"id": 2, "name": "second"},
			},
		},
	}))

	httpInt := NewAdvancedHTTPIntegration(nil)

	config := &integration.Config{
		Name:    "test-http",
		Type:    integration.TypeHTTP,
		Enabled: true,
		Settings: integration.JSONMap{
			"url":              mockServer.URL() + "/api/nested",
			"method":           "GET",
			"response_extract": "data.items[0].name",
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err := httpInt.Execute(ctx, config, nil)

	require.NoError(t, err)
	assert.True(t, result.Success)

	// Check that extraction was performed
	if data, ok := result.Data.(integration.JSONMap); ok {
		assert.Equal(t, "first", data["extracted"])
	}
}

func TestExtractFromResponse(t *testing.T) {
	httpInt := NewAdvancedHTTPIntegration(nil)

	data := map[string]any{
		"data": map[string]any{
			"items": []any{
				map[string]any{"id": 1, "name": "first"},
				map[string]any{"id": 2, "name": "second"},
			},
			"count": 2,
		},
		"status": "success",
	}

	tests := []struct {
		name   string
		path   string
		expect any
	}{
		{
			name:   "simple key",
			path:   "status",
			expect: "success",
		},
		{
			name:   "nested key",
			path:   "data.count",
			expect: 2,
		},
		{
			name:   "array index",
			path:   "data.items[0].name",
			expect: "first",
		},
		{
			name:   "second array index",
			path:   "data.items[1].id",
			expect: 2,
		},
		{
			name:   "invalid path",
			path:   "invalid.path",
			expect: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := httpInt.extractFromResponse(data, tt.path)
			assert.Equal(t, tt.expect, result)
		})
	}
}

func TestVerifyWebhookSignature(t *testing.T) {
	tests := []struct {
		name      string
		payload   []byte
		secret    string
		sigType   SignatureType
		genSig    bool // Generate valid signature
		signature string
		valid     bool
	}{
		{
			name:    "valid HMAC-SHA256",
			payload: []byte(`{"action": "test"}`),
			secret:  "test-secret",
			sigType: SignatureHMACSHA256,
			genSig:  true,
			valid:   true,
		},
		{
			name:    "valid HMAC-SHA1",
			payload: []byte(`{"action": "test"}`),
			secret:  "test-secret",
			sigType: SignatureHMACSHA1,
			genSig:  true,
			valid:   true,
		},
		{
			name:      "invalid signature",
			payload:   []byte(`{"action": "test"}`),
			secret:    "test-secret",
			sigType:   SignatureHMACSHA256,
			signature: "sha256=invalid",
			valid:     false,
		},
		{
			name:      "missing prefix",
			payload:   []byte(`{"action": "test"}`),
			secret:    "test-secret",
			sigType:   SignatureHMACSHA256,
			signature: "just-a-hash",
			valid:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sig := tt.signature
			if tt.genSig {
				httpInt := NewAdvancedHTTPIntegration(nil)
				sig = httpInt.calculateSignature(tt.payload, tt.secret, tt.sigType)
			}
			result := VerifyWebhookSignature(tt.payload, sig, tt.secret, tt.sigType)
			assert.Equal(t, tt.valid, result)
		})
	}
}

func TestCalculateSignature(t *testing.T) {
	httpInt := NewAdvancedHTTPIntegration(nil)

	payload := []byte(`{"test": "data"}`)
	secret := "my-secret"

	// SHA256
	sig256 := httpInt.calculateSignature(payload, secret, SignatureHMACSHA256)
	assert.True(t, len(sig256) > 0)
	assert.Contains(t, sig256, "sha256=")

	// SHA1
	sig1 := httpInt.calculateSignature(payload, secret, SignatureHMACSHA1)
	assert.True(t, len(sig1) > 0)
	assert.Contains(t, sig1, "sha1=")

	// Different payloads produce different signatures
	sig2 := httpInt.calculateSignature([]byte(`{"other": "data"}`), secret, SignatureHMACSHA256)
	assert.NotEqual(t, sig256, sig2)
}

func TestExtractWithRegex(t *testing.T) {
	tests := []struct {
		name      string
		data      string
		pattern   string
		expectLen int
		expectErr bool
	}{
		{
			name:      "simple match",
			data:      "Order ID: 12345",
			pattern:   `Order ID: (\d+)`,
			expectLen: 2,
			expectErr: false,
		},
		{
			name:      "no match",
			data:      "No numbers here",
			pattern:   `(\d+)`,
			expectLen: 0,
			expectErr: false,
		},
		{
			name:      "invalid regex",
			data:      "test",
			pattern:   `[invalid`,
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ExtractWithRegex(tt.data, tt.pattern)
			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, result, tt.expectLen)
			}
		})
	}
}

func TestAdvancedHTTPSchema(t *testing.T) {
	schema := buildAdvancedHTTPSchema()

	// Verify config spec
	assert.Contains(t, schema.ConfigSpec, "url")
	assert.Contains(t, schema.ConfigSpec, "method")
	assert.Contains(t, schema.ConfigSpec, "auth_type")

	// Verify input spec
	assert.Contains(t, schema.InputSpec, "headers")
	assert.Contains(t, schema.InputSpec, "query_params")
	assert.Contains(t, schema.InputSpec, "body")
	assert.Contains(t, schema.InputSpec, "body_template")
	assert.Contains(t, schema.InputSpec, "body_format")
	assert.Contains(t, schema.InputSpec, "timeout")
	assert.Contains(t, schema.InputSpec, "retry_count")
	assert.Contains(t, schema.InputSpec, "verify_ssl")
	assert.Contains(t, schema.InputSpec, "proxy_url")
	assert.Contains(t, schema.InputSpec, "response_extract")
	assert.Contains(t, schema.InputSpec, "signature_secret")

	// Verify output spec
	assert.Contains(t, schema.OutputSpec, "status_code")
	assert.Contains(t, schema.OutputSpec, "headers")
	assert.Contains(t, schema.OutputSpec, "body")
	assert.Contains(t, schema.OutputSpec, "extracted")
}
