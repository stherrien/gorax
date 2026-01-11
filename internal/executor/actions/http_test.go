package actions

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gorax/gorax/internal/security"
)

// newTestHTTPAction creates an HTTP action with SSRF protection disabled for testing
func newTestHTTPAction() *HTTPAction {
	validator := security.NewURLValidatorWithConfig(&security.URLValidatorConfig{
		Enabled: false, // Disable for test servers on localhost
	})
	return NewHTTPActionWithValidator(validator)
}

func TestHTTPAction_Execute_GET(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("Expected GET request, got %s", r.Method)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"message": "success",
			"id":      123,
		})
	}))
	defer server.Close()

	action := newTestHTTPAction()
	config := HTTPActionConfig{
		Method: "GET",
		URL:    server.URL,
	}

	input := NewActionInput(config, nil)
	output, err := action.Execute(context.Background(), input)

	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if output == nil {
		t.Fatal("Execute() returned nil output")
	}

	result, ok := output.Data.(*HTTPActionResult)
	if !ok {
		t.Fatal("Output data is not HTTPActionResult")
	}

	if result.StatusCode != http.StatusOK {
		t.Errorf("StatusCode = %d, want %d", result.StatusCode, http.StatusOK)
	}

	bodyMap, ok := result.Body.(map[string]interface{})
	if !ok {
		t.Fatal("Body is not a map")
	}

	if bodyMap["message"] != "success" {
		t.Errorf("Body message = %v, want 'success'", bodyMap["message"])
	}
}

func TestHTTPAction_Execute_POST(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Expected POST request, got %s", r.Method)
		}

		var body map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Errorf("Failed to decode request body: %v", err)
		}

		if body["name"] != "test" {
			t.Errorf("Body name = %v, want 'test'", body["name"])
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":   456,
			"name": body["name"],
		})
	}))
	defer server.Close()

	action := newTestHTTPAction()
	config := HTTPActionConfig{
		Method: "POST",
		URL:    server.URL,
		Body:   json.RawMessage(`{"name": "test"}`),
	}

	input := NewActionInput(config, nil)
	output, err := action.Execute(context.Background(), input)

	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	result := output.Data.(*HTTPActionResult)
	if result.StatusCode != http.StatusCreated {
		t.Errorf("StatusCode = %d, want %d", result.StatusCode, http.StatusCreated)
	}
}

func TestHTTPAction_Execute_WithHeaders(t *testing.T) {
	expectedToken := "secret-token"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader != "Bearer "+expectedToken {
			t.Errorf("Authorization header = %v, want 'Bearer %s'", authHeader, expectedToken)
		}

		customHeader := r.Header.Get("X-Custom-Header")
		if customHeader != "custom-value" {
			t.Errorf("X-Custom-Header = %v, want 'custom-value'", customHeader)
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}))
	defer server.Close()

	action := newTestHTTPAction()
	config := HTTPActionConfig{
		Method: "GET",
		URL:    server.URL,
		Headers: map[string]string{
			"Authorization":   "Bearer " + expectedToken,
			"X-Custom-Header": "custom-value",
		},
	}

	input := NewActionInput(config, nil)
	_, err := action.Execute(context.Background(), input)

	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
}

func TestHTTPAction_Execute_WithInterpolation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]interface{}
		json.NewDecoder(r.Body).Decode(&body)

		if body["user_name"] != "Alice" {
			t.Errorf("Body user_name = %v, want 'Alice'", body["user_name"])
		}

		if body["user_id"] != "100" {
			t.Errorf("Body user_id = %v, want '100'", body["user_id"])
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{"status": "ok"})
	}))
	defer server.Close()

	action := newTestHTTPAction()
	config := HTTPActionConfig{
		Method: "POST",
		URL:    server.URL + "/users",
		Body:   json.RawMessage(`{"user_name": "{{trigger.name}}", "user_id": "{{trigger.id}}"}`),
	}

	execContext := map[string]interface{}{
		"trigger": map[string]interface{}{
			"name": "Alice",
			"id":   100,
		},
	}

	input := NewActionInput(config, execContext)
	output, err := action.Execute(context.Background(), input)

	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	result := output.Data.(*HTTPActionResult)
	if result.StatusCode != http.StatusOK {
		t.Errorf("StatusCode = %d, want %d", result.StatusCode, http.StatusOK)
	}
}

func TestHTTPAction_Execute_BasicAuth(t *testing.T) {
	expectedUsername := "admin"
	expectedPassword := "secret"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		username, password, ok := r.BasicAuth()
		if !ok {
			t.Error("Basic auth not provided")
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		if username != expectedUsername {
			t.Errorf("Username = %v, want %v", username, expectedUsername)
		}

		if password != expectedPassword {
			t.Errorf("Password = %v, want %v", password, expectedPassword)
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Authenticated"))
	}))
	defer server.Close()

	action := newTestHTTPAction()
	config := HTTPActionConfig{
		Method: "GET",
		URL:    server.URL,
		Auth: &HTTPAuth{
			Type:     "basic",
			Username: expectedUsername,
			Password: expectedPassword,
		},
	}

	input := NewActionInput(config, nil)
	output, err := action.Execute(context.Background(), input)

	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	result := output.Data.(*HTTPActionResult)
	if result.StatusCode != http.StatusOK {
		t.Errorf("StatusCode = %d, want %d", result.StatusCode, http.StatusOK)
	}
}

func TestHTTPAction_Execute_BearerAuth(t *testing.T) {
	expectedToken := "my-bearer-token"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		expected := "Bearer " + expectedToken
		if authHeader != expected {
			t.Errorf("Authorization = %v, want %v", authHeader, expected)
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Authenticated"))
	}))
	defer server.Close()

	action := newTestHTTPAction()
	config := HTTPActionConfig{
		Method: "GET",
		URL:    server.URL,
		Auth: &HTTPAuth{
			Type:  "bearer",
			Token: expectedToken,
		},
	}

	input := NewActionInput(config, nil)
	output, err := action.Execute(context.Background(), input)

	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	result := output.Data.(*HTTPActionResult)
	if result.StatusCode != http.StatusOK {
		t.Errorf("StatusCode = %d, want %d", result.StatusCode, http.StatusOK)
	}
}

func TestHTTPAction_Execute_APIKeyAuth(t *testing.T) {
	expectedAPIKey := "my-api-key"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		apiKey := r.Header.Get("X-API-Key")
		if apiKey != expectedAPIKey {
			t.Errorf("X-API-Key = %v, want %v", apiKey, expectedAPIKey)
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Authenticated"))
	}))
	defer server.Close()

	action := newTestHTTPAction()
	config := HTTPActionConfig{
		Method: "GET",
		URL:    server.URL,
		Auth: &HTTPAuth{
			Type:   "api_key",
			APIKey: expectedAPIKey,
		},
	}

	input := NewActionInput(config, nil)
	output, err := action.Execute(context.Background(), input)

	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	result := output.Data.(*HTTPActionResult)
	if result.StatusCode != http.StatusOK {
		t.Errorf("StatusCode = %d, want %d", result.StatusCode, http.StatusOK)
	}
}

func TestHTTPAction_Execute_CustomAPIKeyHeader(t *testing.T) {
	expectedAPIKey := "my-api-key"
	customHeader := "X-Custom-API-Key"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		apiKey := r.Header.Get(customHeader)
		if apiKey != expectedAPIKey {
			t.Errorf("%s = %v, want %v", customHeader, apiKey, expectedAPIKey)
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Authenticated"))
	}))
	defer server.Close()

	action := newTestHTTPAction()
	config := HTTPActionConfig{
		Method: "GET",
		URL:    server.URL,
		Auth: &HTTPAuth{
			Type:   "api_key",
			APIKey: expectedAPIKey,
			Header: customHeader,
		},
	}

	input := NewActionInput(config, nil)
	output, err := action.Execute(context.Background(), input)

	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	result := output.Data.(*HTTPActionResult)
	if result.StatusCode != http.StatusOK {
		t.Errorf("StatusCode = %d, want %d", result.StatusCode, http.StatusOK)
	}
}

func TestHTTPAction_Execute_Timeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	action := newTestHTTPAction()
	config := HTTPActionConfig{
		Method:  "GET",
		URL:     server.URL,
		Timeout: 1, // 1 second - should not timeout
	}

	input := NewActionInput(config, nil)
	_, err := action.Execute(context.Background(), input)

	if err != nil {
		t.Errorf("Expected no timeout, got error: %v", err)
	}

	// Test with short timeout that should fail
	config.Timeout = 0 // This will still use default 30s, let's use context
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	input2 := NewActionInput(config, nil)
	_, err = action.Execute(ctx, input2)

	if err == nil {
		t.Error("Expected timeout error, got nil")
	}
}

func TestHTTPAction_Execute_InvalidMethod(t *testing.T) {
	action := newTestHTTPAction()
	config := HTTPActionConfig{
		Method: "INVALID",
		URL:    "http://example.com",
	}

	input := NewActionInput(config, nil)
	_, err := action.Execute(context.Background(), input)

	if err == nil {
		t.Error("Expected error for invalid HTTP method, got nil")
	}
}

func TestHTTPAction_Execute_MissingURL(t *testing.T) {
	action := newTestHTTPAction()
	config := HTTPActionConfig{
		Method: "GET",
		URL:    "",
	}

	input := NewActionInput(config, nil)
	_, err := action.Execute(context.Background(), input)

	if err == nil {
		t.Error("Expected error for missing URL, got nil")
	}
}

func TestHTTPAction_Execute_NonJSONResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Plain text response"))
	}))
	defer server.Close()

	action := newTestHTTPAction()
	config := HTTPActionConfig{
		Method: "GET",
		URL:    server.URL,
	}

	input := NewActionInput(config, nil)
	output, err := action.Execute(context.Background(), input)

	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	result := output.Data.(*HTTPActionResult)
	body, ok := result.Body.(string)
	if !ok {
		t.Fatal("Body should be string for non-JSON response")
	}

	if body != "Plain text response" {
		t.Errorf("Body = %v, want 'Plain text response'", body)
	}
}

func TestIsValidHTTPMethod(t *testing.T) {
	tests := []struct {
		method string
		want   bool
	}{
		{"GET", true},
		{"POST", true},
		{"PUT", true},
		{"DELETE", true},
		{"PATCH", true},
		{"HEAD", true},
		{"OPTIONS", true},
		{"INVALID", false},
		{"get", false}, // lowercase not valid
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.method, func(t *testing.T) {
			got := isValidHTTPMethod(tt.method)
			if got != tt.want {
				t.Errorf("isValidHTTPMethod(%s) = %v, want %v", tt.method, got, tt.want)
			}
		})
	}
}

func TestExecuteHTTP_LegacyFunction(t *testing.T) {
	// Note: The legacy function ExecuteHTTP uses the default validator with SSRF enabled.
	// This test uses a non-loopback URL pattern to test the function works.
	// In real usage, SSRF protection would block loopback addresses.

	// Test that the function validates URLs properly
	config := HTTPActionConfig{
		Method: "GET",
		URL:    "http://127.0.0.1/test", // This should be blocked by SSRF
	}

	_, err := ExecuteHTTP(context.Background(), config, nil)

	// Should get SSRF protection error
	if err == nil {
		t.Error("Expected SSRF protection error for loopback address")
	}

	if err != nil && !contains(err.Error(), "SSRF protection") {
		t.Errorf("Expected SSRF protection error, got: %v", err)
	}
}

// SSRF Protection Tests

func TestHTTPAction_SSRFProtection_BlocksLoopback(t *testing.T) {
	action := NewHTTPAction()
	config := HTTPActionConfig{
		Method: "GET",
		URL:    "http://127.0.0.1:8080/admin",
	}

	input := NewActionInput(config, nil)
	_, err := action.Execute(context.Background(), input)

	if err == nil {
		t.Error("Expected SSRF protection to block loopback address")
	}

	if err != nil && !contains(err.Error(), "SSRF protection") {
		t.Errorf("Expected SSRF protection error, got: %v", err)
	}
}

func TestHTTPAction_SSRFProtection_BlocksLocalhost(t *testing.T) {
	action := NewHTTPAction()
	config := HTTPActionConfig{
		Method: "GET",
		URL:    "http://localhost/api",
	}

	input := NewActionInput(config, nil)
	_, err := action.Execute(context.Background(), input)

	if err == nil {
		t.Error("Expected SSRF protection to block localhost")
	}

	if err != nil && !contains(err.Error(), "SSRF protection") {
		t.Errorf("Expected SSRF protection error, got: %v", err)
	}
}

func TestHTTPAction_SSRFProtection_BlocksPrivateIP(t *testing.T) {
	tests := []struct {
		name string
		url  string
	}{
		{"10.x.x.x", "http://10.0.0.1/api"},
		{"172.16.x.x", "http://172.16.0.1/api"},
		{"192.168.x.x", "http://192.168.1.1/api"},
	}

	action := NewHTTPAction()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := HTTPActionConfig{
				Method: "GET",
				URL:    tt.url,
			}

			input := NewActionInput(config, nil)
			_, err := action.Execute(context.Background(), input)

			if err == nil {
				t.Errorf("Expected SSRF protection to block private IP: %s", tt.url)
			}

			if err != nil && !contains(err.Error(), "SSRF protection") {
				t.Errorf("Expected SSRF protection error, got: %v", err)
			}
		})
	}
}

func TestHTTPAction_SSRFProtection_BlocksAWSMetadata(t *testing.T) {
	action := NewHTTPAction()
	config := HTTPActionConfig{
		Method: "GET",
		URL:    "http://169.254.169.254/latest/meta-data/",
	}

	input := NewActionInput(config, nil)
	_, err := action.Execute(context.Background(), input)

	if err == nil {
		t.Error("Expected SSRF protection to block AWS metadata service")
	}

	if err != nil && !contains(err.Error(), "SSRF protection") {
		t.Errorf("Expected SSRF protection error, got: %v", err)
	}
}

func TestHTTPAction_SSRFProtection_BlocksFileScheme(t *testing.T) {
	action := NewHTTPAction()
	config := HTTPActionConfig{
		Method: "GET",
		URL:    "file:///etc/passwd",
	}

	input := NewActionInput(config, nil)
	_, err := action.Execute(context.Background(), input)

	if err == nil {
		t.Error("Expected SSRF protection to block file:// scheme")
	}

	if err != nil && !contains(err.Error(), "SSRF protection") {
		t.Errorf("Expected SSRF protection error, got: %v", err)
	}
}

func TestHTTPAction_SSRFProtection_AllowsPublicURL(t *testing.T) {
	// This test verifies that when SSRF protection is disabled,
	// local test servers work. In production with SSRF enabled,
	// only real public IPs would be allowed.
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}))
	defer server.Close()

	// Use test action with SSRF disabled since httptest uses 127.0.0.1
	action := newTestHTTPAction()
	config := HTTPActionConfig{
		Method: "GET",
		URL:    server.URL,
	}

	input := NewActionInput(config, nil)
	output, err := action.Execute(context.Background(), input)

	if err != nil {
		t.Fatalf("Should allow URLs when SSRF protection is disabled, got error: %v", err)
	}

	if output == nil {
		t.Fatal("Expected output, got nil")
	}

	result := output.Data.(*HTTPActionResult)
	if result.StatusCode != http.StatusOK {
		t.Errorf("StatusCode = %d, want %d", result.StatusCode, http.StatusOK)
	}
}

func TestHTTPAction_SSRFProtection_WithInterpolation(t *testing.T) {
	action := NewHTTPAction()
	config := HTTPActionConfig{
		Method: "GET",
		URL:    "http://{{host}}/api",
	}

	// Test with blocked host via interpolation
	execContext := map[string]interface{}{
		"host": "127.0.0.1",
	}

	input := NewActionInput(config, execContext)
	_, err := action.Execute(context.Background(), input)

	if err == nil {
		t.Error("Expected SSRF protection to block interpolated loopback address")
	}

	if err != nil && !contains(err.Error(), "SSRF protection") {
		t.Errorf("Expected SSRF protection error, got: %v", err)
	}
}

// Helper function for string containment
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 || indexOf(s, substr) >= 0)
}

func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
