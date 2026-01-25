package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

// MakeHTTPRequest is a helper to make HTTP requests in tests
func MakeHTTPRequest(t *testing.T, client *http.Client, baseURL, method, path string, body any, headers map[string]string) *http.Response {
	t.Helper()

	var bodyReader io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		require.NoError(t, err, "failed to marshal request body")
		bodyReader = bytes.NewReader(jsonBody)
	}

	url := baseURL + path
	req, err := http.NewRequest(method, url, bodyReader)
	require.NoError(t, err, "failed to create request")

	// Set default headers
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	// Set custom headers
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := client.Do(req)
	require.NoError(t, err, "failed to execute request")

	return resp
}

// MakeRawRequest makes an HTTP request with a raw body (not JSON-encoded)
func MakeRawRequest(t *testing.T, client *http.Client, baseURL, method, path string, body []byte, headers map[string]string) *http.Response {
	t.Helper()

	var bodyReader io.Reader
	if body != nil {
		bodyReader = bytes.NewReader(body)
	}

	url := baseURL + path
	req, err := http.NewRequest(method, url, bodyReader)
	require.NoError(t, err, "failed to create request")

	// Set custom headers
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := client.Do(req)
	require.NoError(t, err, "failed to execute request")

	return resp
}

// ParseJSONResponse parses a JSON response into a struct
func ParseJSONResponse(t *testing.T, resp *http.Response, v any) {
	t.Helper()

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err, "failed to read response body")

	err = json.Unmarshal(body, v)
	require.NoError(t, err, fmt.Sprintf("failed to unmarshal response: %s", string(body)))
}

// AssertStatusCode asserts that the response has the expected status code
func AssertStatusCode(t *testing.T, resp *http.Response, expected int) {
	t.Helper()

	if resp.StatusCode != expected {
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected status code %d, got %d. Body: %s", expected, resp.StatusCode, string(body))
	}
}

// GetResponseBody reads the response body as a string
func GetResponseBody(t *testing.T, resp *http.Response) string {
	t.Helper()

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err, "failed to read response body")

	return string(body)
}

// DefaultTestHeaders returns default headers for authenticated requests
func DefaultTestHeaders(tenantID string) map[string]string {
	return map[string]string{
		"X-Tenant-ID":   tenantID,
		"Authorization": "Bearer test-token",
	}
}

// AdminTestHeaders returns headers for admin requests
func AdminTestHeaders() map[string]string {
	return map[string]string{
		"Authorization": "Bearer admin-token",
		"X-User-Role":   "admin",
	}
}
