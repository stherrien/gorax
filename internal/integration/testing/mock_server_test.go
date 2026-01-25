package testing

import (
	"bytes"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewMockServer(t *testing.T) {
	ms := NewMockServer()
	defer ms.Close()

	require.NotNil(t, ms)
	assert.NotEmpty(t, ms.URL())
	assert.NotNil(t, ms.server)
	assert.NotNil(t, ms.responses)
	assert.Empty(t, ms.requests)
}

func TestMockServer_OnRequest(t *testing.T) {
	ms := NewMockServer()
	defer ms.Close()

	response := &MockResponse{
		StatusCode: http.StatusOK,
		Body:       map[string]string{"status": "ok"},
	}

	ms.OnRequest("GET", "/test", response)

	resp, err := http.Get(ms.URL() + "/test")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestMockServer_OnGet(t *testing.T) {
	ms := NewMockServer()
	defer ms.Close()

	ms.OnGet("/users", JSONResponse(http.StatusOK, map[string]string{"name": "John"}))

	resp, err := http.Get(ms.URL() + "/users")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestMockServer_OnPost(t *testing.T) {
	ms := NewMockServer()
	defer ms.Close()

	ms.OnPost("/users", JSONResponse(http.StatusCreated, map[string]string{"id": "123"}))

	resp, err := http.Post(ms.URL()+"/users", "application/json", bytes.NewReader([]byte(`{"name": "John"}`)))
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusCreated, resp.StatusCode)
}

func TestMockServer_OnPut(t *testing.T) {
	ms := NewMockServer()
	defer ms.Close()

	ms.OnPut("/users/123", JSONResponse(http.StatusOK, map[string]string{"updated": "true"}))

	req, _ := http.NewRequest(http.MethodPut, ms.URL()+"/users/123", bytes.NewReader([]byte(`{"name": "Jane"}`)))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestMockServer_OnDelete(t *testing.T) {
	ms := NewMockServer()
	defer ms.Close()

	ms.OnDelete("/users/123", JSONResponse(http.StatusNoContent, nil))

	req, _ := http.NewRequest(http.MethodDelete, ms.URL()+"/users/123", nil)

	client := &http.Client{}
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusNoContent, resp.StatusCode)
}

func TestMockServer_OnPatch(t *testing.T) {
	ms := NewMockServer()
	defer ms.Close()

	ms.OnPatch("/users/123", JSONResponse(http.StatusOK, map[string]string{"patched": "true"}))

	req, _ := http.NewRequest(http.MethodPatch, ms.URL()+"/users/123", bytes.NewReader([]byte(`{"name": "Jane"}`)))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestMockServer_RecordRequests(t *testing.T) {
	ms := NewMockServer()
	defer ms.Close()

	ms.OnGet("/test", JSONResponse(http.StatusOK, nil))

	// Make multiple requests
	for i := 0; i < 3; i++ {
		resp, err := http.Get(ms.URL() + "/test")
		require.NoError(t, err)
		resp.Body.Close()
	}

	assert.Equal(t, 3, ms.GetRequestCount())

	requests := ms.GetRequests()
	assert.Len(t, requests, 3)

	lastRequest := ms.GetLastRequest()
	require.NotNil(t, lastRequest)
	assert.Equal(t, "GET", lastRequest.Method)
	assert.Equal(t, "/test", lastRequest.URL)
}

func TestMockServer_GetLastRequest_Empty(t *testing.T) {
	ms := NewMockServer()
	defer ms.Close()

	assert.Nil(t, ms.GetLastRequest())
}

func TestMockServer_ClearRequests(t *testing.T) {
	ms := NewMockServer()
	defer ms.Close()

	ms.OnGet("/test", JSONResponse(http.StatusOK, nil))

	resp, _ := http.Get(ms.URL() + "/test")
	resp.Body.Close()

	assert.Equal(t, 1, ms.GetRequestCount())

	ms.ClearRequests()
	assert.Equal(t, 0, ms.GetRequestCount())
}

func TestMockServer_Reset(t *testing.T) {
	ms := NewMockServer()
	defer ms.Close()

	ms.OnGet("/test", JSONResponse(http.StatusOK, nil))
	resp, _ := http.Get(ms.URL() + "/test")
	resp.Body.Close()

	ms.Reset()

	assert.Equal(t, 0, ms.GetRequestCount())

	// Response should no longer be configured
	resp, err := http.Get(ms.URL() + "/test")
	require.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	resp.Body.Close()
}

func TestMockServer_NotFound(t *testing.T) {
	ms := NewMockServer()
	defer ms.Close()

	resp, err := http.Get(ms.URL() + "/nonexistent")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestMockServer_CustomHandler(t *testing.T) {
	ms := NewMockServer()
	defer ms.Close()

	ms.OnGet("/custom", &MockResponse{
		Handler: func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Custom-Header", "custom-value")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("custom response"))
		},
	})

	resp, err := http.Get(ms.URL() + "/custom")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "custom-value", resp.Header.Get("X-Custom-Header"))

	body, _ := io.ReadAll(resp.Body)
	assert.Equal(t, "custom response", string(body))
}

func TestMockServer_DelayedResponse(t *testing.T) {
	ms := NewMockServer()
	defer ms.Close()

	ms.OnGet("/slow", DelayedResponse(http.StatusOK, "slow response", 100*time.Millisecond))

	start := time.Now()
	resp, err := http.Get(ms.URL() + "/slow")
	duration := time.Since(start)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.GreaterOrEqual(t, duration, 100*time.Millisecond)
}

func TestMockServer_CustomHeaders(t *testing.T) {
	ms := NewMockServer()
	defer ms.Close()

	ms.OnGet("/headers", &MockResponse{
		StatusCode: http.StatusOK,
		Headers: map[string]string{
			"X-Custom-1":   "value1",
			"X-Custom-2":   "value2",
			"Content-Type": "text/plain",
		},
		Body: "test",
	})

	resp, err := http.Get(ms.URL() + "/headers")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, "value1", resp.Header.Get("X-Custom-1"))
	assert.Equal(t, "value2", resp.Header.Get("X-Custom-2"))
	assert.Equal(t, "text/plain", resp.Header.Get("Content-Type"))
}

func TestMockServer_ByteBody(t *testing.T) {
	ms := NewMockServer()
	defer ms.Close()

	ms.OnGet("/bytes", &MockResponse{
		StatusCode: http.StatusOK,
		Body:       []byte("byte content"),
	})

	resp, err := http.Get(ms.URL() + "/bytes")
	require.NoError(t, err)
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	assert.Equal(t, "byte content", string(body))
}

func TestMockServer_StringBody(t *testing.T) {
	ms := NewMockServer()
	defer ms.Close()

	ms.OnGet("/string", &MockResponse{
		StatusCode: http.StatusOK,
		Body:       "string content",
	})

	resp, err := http.Get(ms.URL() + "/string")
	require.NoError(t, err)
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	assert.Equal(t, "string content", string(body))
}

func TestMockServer_DefaultStatusCode(t *testing.T) {
	ms := NewMockServer()
	defer ms.Close()

	ms.OnGet("/default", &MockResponse{
		Body: "test",
	})

	resp, err := http.Get(ms.URL() + "/default")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestMockServer_RecordsRequestBody(t *testing.T) {
	ms := NewMockServer()
	defer ms.Close()

	ms.OnPost("/data", JSONResponse(http.StatusOK, nil))

	requestBody := []byte(`{"key": "value"}`)
	resp, err := http.Post(ms.URL()+"/data", "application/json", bytes.NewReader(requestBody))
	require.NoError(t, err)
	resp.Body.Close()

	lastRequest := ms.GetLastRequest()
	require.NotNil(t, lastRequest)
	assert.Equal(t, requestBody, lastRequest.Body)
	assert.Equal(t, "application/json", lastRequest.Headers.Get("Content-Type"))
}

func TestMockServer_RecordsRequestTimestamp(t *testing.T) {
	ms := NewMockServer()
	defer ms.Close()

	ms.OnGet("/timestamp", JSONResponse(http.StatusOK, nil))

	before := time.Now()
	resp, _ := http.Get(ms.URL() + "/timestamp")
	resp.Body.Close()
	after := time.Now()

	lastRequest := ms.GetLastRequest()
	require.NotNil(t, lastRequest)
	assert.True(t, lastRequest.ReceivedAt.After(before) || lastRequest.ReceivedAt.Equal(before))
	assert.True(t, lastRequest.ReceivedAt.Before(after) || lastRequest.ReceivedAt.Equal(after))
}

func TestJSONResponse(t *testing.T) {
	response := JSONResponse(http.StatusCreated, map[string]string{"id": "123"})

	assert.Equal(t, http.StatusCreated, response.StatusCode)
	assert.Equal(t, "application/json", response.Headers["Content-Type"])
	assert.NotNil(t, response.Body)
}

func TestTextResponse(t *testing.T) {
	response := TextResponse(http.StatusOK, "hello world")

	assert.Equal(t, http.StatusOK, response.StatusCode)
	assert.Equal(t, "text/plain", response.Headers["Content-Type"])
	assert.Equal(t, "hello world", response.Body)
}

func TestErrorResponse(t *testing.T) {
	response := ErrorResponse(http.StatusBadRequest, "invalid input")

	assert.Equal(t, http.StatusBadRequest, response.StatusCode)
	assert.Equal(t, "application/json", response.Headers["Content-Type"])

	body := response.Body.(map[string]string)
	assert.Equal(t, "invalid input", body["error"])
}

func TestDelayedResponse_Factory(t *testing.T) {
	response := DelayedResponse(http.StatusOK, "delayed", 50*time.Millisecond)

	assert.Equal(t, http.StatusOK, response.StatusCode)
	assert.Equal(t, "delayed", response.Body)
	assert.Equal(t, 50*time.Millisecond, response.Delay)
}
