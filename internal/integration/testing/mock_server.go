// Package testing provides utilities for testing integrations.
package testing

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"time"
)

// MockServer is a test HTTP server that records requests and returns configured responses.
type MockServer struct {
	server    *httptest.Server
	responses map[string]*MockResponse
	requests  []*RecordedRequest
	mu        sync.RWMutex
}

// MockResponse defines the response for a mock endpoint.
type MockResponse struct {
	StatusCode int
	Headers    map[string]string
	Body       any
	Delay      time.Duration
	Error      error
	Handler    http.HandlerFunc // Custom handler for complex scenarios
}

// RecordedRequest represents a request that was received by the mock server.
type RecordedRequest struct {
	Method     string
	URL        string
	Headers    http.Header
	Body       []byte
	ReceivedAt time.Time
}

// NewMockServer creates a new mock server.
func NewMockServer() *MockServer {
	ms := &MockServer{
		responses: make(map[string]*MockResponse),
		requests:  make([]*RecordedRequest, 0),
	}

	ms.server = httptest.NewServer(http.HandlerFunc(ms.handle))
	return ms
}

// URL returns the base URL of the mock server.
func (ms *MockServer) URL() string {
	return ms.server.URL
}

// Close shuts down the mock server.
func (ms *MockServer) Close() {
	ms.server.Close()
}

// OnRequest configures a response for a specific method and path.
func (ms *MockServer) OnRequest(method, path string, response *MockResponse) {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	key := method + ":" + path
	ms.responses[key] = response
}

// OnGet configures a response for GET requests to the path.
func (ms *MockServer) OnGet(path string, response *MockResponse) {
	ms.OnRequest(http.MethodGet, path, response)
}

// OnPost configures a response for POST requests to the path.
func (ms *MockServer) OnPost(path string, response *MockResponse) {
	ms.OnRequest(http.MethodPost, path, response)
}

// OnPut configures a response for PUT requests to the path.
func (ms *MockServer) OnPut(path string, response *MockResponse) {
	ms.OnRequest(http.MethodPut, path, response)
}

// OnDelete configures a response for DELETE requests to the path.
func (ms *MockServer) OnDelete(path string, response *MockResponse) {
	ms.OnRequest(http.MethodDelete, path, response)
}

// OnPatch configures a response for PATCH requests to the path.
func (ms *MockServer) OnPatch(path string, response *MockResponse) {
	ms.OnRequest(http.MethodPatch, path, response)
}

// GetRequests returns all recorded requests.
func (ms *MockServer) GetRequests() []*RecordedRequest {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	result := make([]*RecordedRequest, len(ms.requests))
	copy(result, ms.requests)
	return result
}

// GetRequestCount returns the number of recorded requests.
func (ms *MockServer) GetRequestCount() int {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	return len(ms.requests)
}

// GetLastRequest returns the most recent recorded request.
func (ms *MockServer) GetLastRequest() *RecordedRequest {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	if len(ms.requests) == 0 {
		return nil
	}
	return ms.requests[len(ms.requests)-1]
}

// ClearRequests clears all recorded requests.
func (ms *MockServer) ClearRequests() {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	ms.requests = make([]*RecordedRequest, 0)
}

// Reset clears all responses and requests.
func (ms *MockServer) Reset() {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	ms.responses = make(map[string]*MockResponse)
	ms.requests = make([]*RecordedRequest, 0)
}

// handle processes incoming requests.
func (ms *MockServer) handle(w http.ResponseWriter, r *http.Request) {
	// Record the request
	body, _ := io.ReadAll(r.Body)
	recorded := &RecordedRequest{
		Method:     r.Method,
		URL:        r.URL.String(),
		Headers:    r.Header,
		Body:       body,
		ReceivedAt: time.Now(),
	}

	ms.mu.Lock()
	ms.requests = append(ms.requests, recorded)
	ms.mu.Unlock()

	// Find matching response
	ms.mu.RLock()
	key := r.Method + ":" + r.URL.Path
	response, exists := ms.responses[key]
	ms.mu.RUnlock()

	if !exists {
		// Try with just the path (any method)
		ms.mu.RLock()
		response, exists = ms.responses["*:"+r.URL.Path]
		ms.mu.RUnlock()
	}

	if !exists {
		http.NotFound(w, r)
		return
	}

	// Handle custom handler
	if response.Handler != nil {
		response.Handler(w, r)
		return
	}

	// Apply delay
	if response.Delay > 0 {
		time.Sleep(response.Delay)
	}

	// Set headers
	for key, value := range response.Headers {
		w.Header().Set(key, value)
	}

	// Set default content type if not set
	if w.Header().Get("Content-Type") == "" {
		w.Header().Set("Content-Type", "application/json")
	}

	// Write status code
	statusCode := response.StatusCode
	if statusCode == 0 {
		statusCode = http.StatusOK
	}
	w.WriteHeader(statusCode)

	// Write body
	if response.Body != nil {
		switch body := response.Body.(type) {
		case []byte:
			_, _ = w.Write(body)
		case string:
			_, _ = w.Write([]byte(body))
		default:
			data, _ := json.Marshal(body)
			_, _ = w.Write(data)
		}
	}
}

// JSONResponse creates a mock response with JSON body.
func JSONResponse(statusCode int, body any) *MockResponse {
	return &MockResponse{
		StatusCode: statusCode,
		Body:       body,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
	}
}

// TextResponse creates a mock response with text body.
func TextResponse(statusCode int, body string) *MockResponse {
	return &MockResponse{
		StatusCode: statusCode,
		Body:       body,
		Headers: map[string]string{
			"Content-Type": "text/plain",
		},
	}
}

// ErrorResponse creates an error response.
func ErrorResponse(statusCode int, message string) *MockResponse {
	return &MockResponse{
		StatusCode: statusCode,
		Body: map[string]string{
			"error": message,
		},
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
	}
}

// DelayedResponse creates a response with a delay.
func DelayedResponse(statusCode int, body any, delay time.Duration) *MockResponse {
	return &MockResponse{
		StatusCode: statusCode,
		Body:       body,
		Delay:      delay,
	}
}
