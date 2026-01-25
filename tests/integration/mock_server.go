package integration

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"time"
)

// MockServer provides a configurable HTTP server for testing external API calls
type MockServer struct {
	Server          *httptest.Server
	mu              sync.RWMutex
	requests        []RecordedRequest
	responses       map[string]MockResponse
	defaultResponse *MockResponse
	authConfig      *AuthConfig
	delayMs         int
	shutdownChan    chan struct{}
}

// RecordedRequest stores details of requests made to the mock server
type RecordedRequest struct {
	Method      string
	Path        string
	Headers     map[string]string
	Body        []byte
	QueryParams map[string][]string
	ReceivedAt  time.Time
}

// MockResponse defines a response to return for a specific path
type MockResponse struct {
	StatusCode  int
	Body        any
	Headers     map[string]string
	DelayMs     int
	Error       bool // If true, return connection error
	ErrorMsg    string
	ContentType string
}

// AuthConfig defines authentication requirements for the mock server
type AuthConfig struct {
	Type            string            // "bearer", "basic", "api_key", "signature"
	Token           string            // For bearer auth
	Username        string            // For basic auth
	Password        string            // For basic auth
	APIKey          string            // For API key auth
	APIKeyHeader    string            // Header name for API key (default: X-API-Key)
	SignatureKey    string            // For HMAC signature verification
	SignatureHeader string            // Header name for signature (default: X-Signature)
	AdditionalAuth  map[string]string // Additional headers to validate
}

// NewMockServer creates a new mock HTTP server
func NewMockServer() *MockServer {
	ms := &MockServer{
		requests:     make([]RecordedRequest, 0),
		responses:    make(map[string]MockResponse),
		shutdownChan: make(chan struct{}),
		defaultResponse: &MockResponse{
			StatusCode:  200,
			Body:        map[string]string{"status": "ok"},
			ContentType: "application/json",
		},
	}

	ms.Server = httptest.NewServer(http.HandlerFunc(ms.handler))
	return ms
}

// NewMockServerTLS creates a new mock HTTPS server
func NewMockServerTLS() *MockServer {
	ms := &MockServer{
		requests:     make([]RecordedRequest, 0),
		responses:    make(map[string]MockResponse),
		shutdownChan: make(chan struct{}),
		defaultResponse: &MockResponse{
			StatusCode:  200,
			Body:        map[string]string{"status": "ok"},
			ContentType: "application/json",
		},
	}

	ms.Server = httptest.NewTLSServer(http.HandlerFunc(ms.handler))
	return ms
}

// handler processes incoming requests
func (ms *MockServer) handler(w http.ResponseWriter, r *http.Request) {
	// Record the request
	ms.recordRequest(r)

	// Apply global delay if set
	if ms.delayMs > 0 {
		time.Sleep(time.Duration(ms.delayMs) * time.Millisecond)
	}

	// Check authentication
	if ms.authConfig != nil {
		if !ms.validateAuth(r) {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
	}

	// Find response for this path
	response := ms.getResponse(r.Method, r.URL.Path)

	// Apply response-specific delay
	if response.DelayMs > 0 {
		time.Sleep(time.Duration(response.DelayMs) * time.Millisecond)
	}

	// Check for simulated error
	if response.Error {
		// Close connection without sending response
		hj, ok := w.(http.Hijacker)
		if ok {
			conn, _, err := hj.Hijack()
			if err == nil {
				conn.Close()
				return
			}
		}
		http.Error(w, response.ErrorMsg, http.StatusInternalServerError)
		return
	}

	// Set headers
	if response.ContentType != "" {
		w.Header().Set("Content-Type", response.ContentType)
	} else {
		w.Header().Set("Content-Type", "application/json")
	}
	for k, v := range response.Headers {
		w.Header().Set(k, v)
	}

	// Write response
	w.WriteHeader(response.StatusCode)

	switch body := response.Body.(type) {
	case string:
		w.Write([]byte(body))
	case []byte:
		w.Write(body)
	default:
		json.NewEncoder(w).Encode(body)
	}
}

// recordRequest stores the request details
func (ms *MockServer) recordRequest(r *http.Request) {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	// Read body
	bodyBytes, _ := io.ReadAll(r.Body)
	r.Body.Close()
	// Restore body for potential re-reading
	r.Body = io.NopCloser(strings.NewReader(string(bodyBytes)))

	// Extract headers
	headers := make(map[string]string)
	for k, v := range r.Header {
		if len(v) > 0 {
			headers[k] = v[0]
		}
	}

	ms.requests = append(ms.requests, RecordedRequest{
		Method:      r.Method,
		Path:        r.URL.Path,
		Headers:     headers,
		Body:        bodyBytes,
		QueryParams: r.URL.Query(),
		ReceivedAt:  time.Now(),
	})
}

// validateAuth checks if the request has valid authentication
func (ms *MockServer) validateAuth(r *http.Request) bool {
	switch strings.ToLower(ms.authConfig.Type) {
	case "bearer":
		auth := r.Header.Get("Authorization")
		return auth == "Bearer "+ms.authConfig.Token

	case "basic":
		auth := r.Header.Get("Authorization")
		expected := base64.StdEncoding.EncodeToString(
			[]byte(ms.authConfig.Username + ":" + ms.authConfig.Password),
		)
		return auth == "Basic "+expected

	case "api_key":
		header := ms.authConfig.APIKeyHeader
		if header == "" {
			header = "X-API-Key"
		}
		return r.Header.Get(header) == ms.authConfig.APIKey

	case "signature":
		return ms.validateSignature(r)

	default:
		return true
	}
}

// validateSignature verifies HMAC signature
func (ms *MockServer) validateSignature(r *http.Request) bool {
	header := ms.authConfig.SignatureHeader
	if header == "" {
		header = "X-Signature"
	}

	signature := r.Header.Get(header)
	if signature == "" {
		return false
	}

	// Read body for signature calculation
	bodyBytes, _ := io.ReadAll(r.Body)
	r.Body.Close()
	r.Body = io.NopCloser(strings.NewReader(string(bodyBytes)))

	// Calculate expected signature
	mac := hmac.New(sha256.New, []byte(ms.authConfig.SignatureKey))
	mac.Write(bodyBytes)
	expected := base64.StdEncoding.EncodeToString(mac.Sum(nil))

	return signature == expected || signature == fmt.Sprintf("sha256=%s", expected)
}

// getResponse finds the appropriate response for the request
func (ms *MockServer) getResponse(method, path string) MockResponse {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	// Try exact match with method
	key := method + ":" + path
	if resp, ok := ms.responses[key]; ok {
		return resp
	}

	// Try path only
	if resp, ok := ms.responses[path]; ok {
		return resp
	}

	// Try wildcard match
	for pattern, resp := range ms.responses {
		if prefix, found := strings.CutSuffix(pattern, "*"); found {
			if strings.HasPrefix(path, prefix) {
				return resp
			}
		}
	}

	return *ms.defaultResponse
}

// SetResponse configures a response for a specific path
func (ms *MockServer) SetResponse(path string, response MockResponse) {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	ms.responses[path] = response
}

// SetMethodResponse configures a response for a specific method and path
func (ms *MockServer) SetMethodResponse(method, path string, response MockResponse) {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	ms.responses[method+":"+path] = response
}

// SetDefaultResponse sets the default response for unmatched paths
func (ms *MockServer) SetDefaultResponse(response MockResponse) {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	ms.defaultResponse = &response
}

// SetAuth configures authentication requirements
func (ms *MockServer) SetAuth(config *AuthConfig) {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	ms.authConfig = config
}

// SetGlobalDelay sets a delay for all requests
func (ms *MockServer) SetGlobalDelay(delayMs int) {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	ms.delayMs = delayMs
}

// GetRequests returns all recorded requests
func (ms *MockServer) GetRequests() []RecordedRequest {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	result := make([]RecordedRequest, len(ms.requests))
	copy(result, ms.requests)
	return result
}

// GetLastRequest returns the most recent request
func (ms *MockServer) GetLastRequest() *RecordedRequest {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	if len(ms.requests) == 0 {
		return nil
	}
	req := ms.requests[len(ms.requests)-1]
	return &req
}

// GetRequestCount returns the number of requests received
func (ms *MockServer) GetRequestCount() int {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	return len(ms.requests)
}

// GetRequestsForPath returns requests for a specific path
func (ms *MockServer) GetRequestsForPath(path string) []RecordedRequest {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	var result []RecordedRequest
	for _, req := range ms.requests {
		if req.Path == path {
			result = append(result, req)
		}
	}
	return result
}

// ClearRequests clears all recorded requests
func (ms *MockServer) ClearRequests() {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	ms.requests = make([]RecordedRequest, 0)
}

// ClearResponses clears all configured responses
func (ms *MockServer) ClearResponses() {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	ms.responses = make(map[string]MockResponse)
}

// Reset clears all requests and responses
func (ms *MockServer) Reset() {
	ms.ClearRequests()
	ms.ClearResponses()
	ms.authConfig = nil
	ms.delayMs = 0
}

// URL returns the base URL of the mock server
func (ms *MockServer) URL() string {
	return ms.Server.URL
}

// Close shuts down the mock server
func (ms *MockServer) Close() {
	close(ms.shutdownChan)
	ms.Server.Close()
}

// MockSlackAPI creates a mock server that simulates Slack API
func NewMockSlackAPI() *MockServer {
	ms := NewMockServer()

	// Chat.postMessage
	ms.SetMethodResponse("POST", "/api/chat.postMessage", MockResponse{
		StatusCode: 200,
		Body: map[string]any{
			"ok":      true,
			"channel": "C123456",
			"ts":      "1234567890.123456",
			"message": map[string]any{
				"text": "Message sent",
			},
		},
	})

	// Chat.update
	ms.SetMethodResponse("POST", "/api/chat.update", MockResponse{
		StatusCode: 200,
		Body: map[string]any{
			"ok":      true,
			"channel": "C123456",
			"ts":      "1234567890.123456",
		},
	})

	// Reactions.add
	ms.SetMethodResponse("POST", "/api/reactions.add", MockResponse{
		StatusCode: 200,
		Body: map[string]any{
			"ok": true,
		},
	})

	// Conversations.open (for DMs)
	ms.SetMethodResponse("POST", "/api/conversations.open", MockResponse{
		StatusCode: 200,
		Body: map[string]any{
			"ok": true,
			"channel": map[string]any{
				"id": "D123456",
			},
		},
	})

	return ms
}

// NewMockWebhookEndpoint creates a mock server for webhook testing
func NewMockWebhookEndpoint() *MockServer {
	ms := NewMockServer()

	ms.SetDefaultResponse(MockResponse{
		StatusCode: 200,
		Body: map[string]any{
			"received": true,
		},
	})

	return ms
}

// RequestBodyAs unmarshals the recorded request body into the provided struct
func (r *RecordedRequest) RequestBodyAs(v any) error {
	return json.Unmarshal(r.Body, v)
}

// HasHeader checks if a header exists with the expected value
func (r *RecordedRequest) HasHeader(key, value string) bool {
	return r.Headers[key] == value
}

// HasHeaderContaining checks if a header contains a substring
func (r *RecordedRequest) HasHeaderContaining(key, substr string) bool {
	return strings.Contains(r.Headers[key], substr)
}

// QueryParam returns a query parameter value
func (r *RecordedRequest) QueryParam(key string) string {
	values := r.QueryParams[key]
	if len(values) > 0 {
		return values[0]
	}
	return ""
}
