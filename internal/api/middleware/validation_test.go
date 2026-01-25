package middleware

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDefaultRequestValidationConfig(t *testing.T) {
	cfg := DefaultRequestValidationConfig()

	assert.Equal(t, int64(1024*1024), cfg.MaxBodySize)
	assert.True(t, cfg.ValidateJSON)
	assert.True(t, cfg.SanitizeStrings)
	assert.False(t, cfg.BlockSQLInjection)
	assert.True(t, cfg.BlockXSSPatterns)
}

func TestRequestValidation(t *testing.T) {
	tests := []struct {
		name           string
		config         RequestValidationConfig
		method         string
		contentType    string
		body           string
		queryParams    string
		expectedStatus int
		checkResponse  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name:           "GET request passes without body validation",
			config:         DefaultRequestValidationConfig(),
			method:         http.MethodGet,
			contentType:    "",
			body:           "",
			queryParams:    "",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "POST with valid JSON body",
			config:         DefaultRequestValidationConfig(),
			method:         http.MethodPost,
			contentType:    "application/json",
			body:           `{"name": "test", "value": 123}`,
			queryParams:    "",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "PUT with valid JSON body",
			config:         DefaultRequestValidationConfig(),
			method:         http.MethodPut,
			contentType:    "application/json",
			body:           `{"id": "123", "name": "updated"}`,
			queryParams:    "",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "PATCH with valid JSON body",
			config:         DefaultRequestValidationConfig(),
			method:         http.MethodPatch,
			contentType:    "application/json",
			body:           `{"status": "active"}`,
			queryParams:    "",
			expectedStatus: http.StatusOK,
		},
		{
			name: "body too large by content length",
			config: RequestValidationConfig{
				MaxBodySize:  100,
				ValidateJSON: true,
			},
			method:         http.MethodPost,
			contentType:    "application/json",
			body:           strings.Repeat("x", 150),
			queryParams:    "",
			expectedStatus: http.StatusUnprocessableEntity,
		},
		{
			name:           "POST with invalid JSON",
			config:         DefaultRequestValidationConfig(),
			method:         http.MethodPost,
			contentType:    "application/json",
			body:           `{invalid json}`,
			queryParams:    "",
			expectedStatus: http.StatusUnprocessableEntity,
		},
		{
			name:           "POST with XSS in JSON body",
			config:         DefaultRequestValidationConfig(),
			method:         http.MethodPost,
			contentType:    "application/json",
			body:           `{"name": "<script>alert('xss')</script>"}`,
			queryParams:    "",
			expectedStatus: http.StatusUnprocessableEntity,
		},
		{
			name:           "POST with XSS in nested JSON",
			config:         DefaultRequestValidationConfig(),
			method:         http.MethodPost,
			contentType:    "application/json",
			body:           `{"data": {"nested": "<img src=x onerror=alert(1)>"}}`,
			queryParams:    "",
			expectedStatus: http.StatusUnprocessableEntity,
		},
		{
			name:           "POST with XSS in JSON array",
			config:         DefaultRequestValidationConfig(),
			method:         http.MethodPost,
			contentType:    "application/json",
			body:           `{"items": ["safe", "<script>bad</script>"]}`,
			queryParams:    "",
			expectedStatus: http.StatusUnprocessableEntity,
		},
		{
			name: "XSS validation disabled passes XSS",
			config: RequestValidationConfig{
				MaxBodySize:      1024 * 1024,
				ValidateJSON:     true,
				BlockXSSPatterns: false,
			},
			method:         http.MethodPost,
			contentType:    "application/json",
			body:           `{"name": "<script>alert('xss')</script>"}`,
			queryParams:    "",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "XSS in query parameter blocked",
			config:         DefaultRequestValidationConfig(),
			method:         http.MethodGet,
			contentType:    "",
			body:           "",
			queryParams:    "?search=%3Cscript%3Ealert(1)%3C/script%3E",
			expectedStatus: http.StatusUnprocessableEntity,
		},
		{
			name: "SQL injection in query parameter blocked when enabled",
			config: RequestValidationConfig{
				MaxBodySize:       1024 * 1024,
				BlockSQLInjection: true,
			},
			method:         http.MethodGet,
			contentType:    "",
			body:           "",
			queryParams:    "?id=1%3B%20DROP%20TABLE%20users%3B--",
			expectedStatus: http.StatusUnprocessableEntity,
		},
		{
			name:           "SQL injection in query parameter allowed by default",
			config:         DefaultRequestValidationConfig(), // BlockSQLInjection is false by default
			method:         http.MethodGet,
			contentType:    "",
			body:           "",
			queryParams:    "?id=1%3B%20DROP%20TABLE%20users%3B--",
			expectedStatus: http.StatusOK,
		},
		{
			name: "XSS in query parameter allowed when disabled",
			config: RequestValidationConfig{
				MaxBodySize:      1024 * 1024,
				BlockXSSPatterns: false,
			},
			method:         http.MethodGet,
			contentType:    "",
			body:           "",
			queryParams:    "?search=%3Cscript%3Ealert(1)%3C/script%3E",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "POST without body passes",
			config:         DefaultRequestValidationConfig(),
			method:         http.MethodPost,
			contentType:    "application/json",
			body:           "",
			queryParams:    "",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "POST with non-JSON content type skips JSON validation",
			config:         DefaultRequestValidationConfig(),
			method:         http.MethodPost,
			contentType:    "text/plain",
			body:           "not json but valid plain text",
			queryParams:    "",
			expectedStatus: http.StatusOK,
		},
		{
			name: "JSON validation disabled skips JSON validation",
			config: RequestValidationConfig{
				MaxBodySize:  1024 * 1024,
				ValidateJSON: false,
			},
			method:         http.MethodPost,
			contentType:    "application/json",
			body:           "{invalid json}",
			queryParams:    "",
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request
			var body io.Reader
			if tt.body != "" {
				body = strings.NewReader(tt.body)
			}
			req := httptest.NewRequest(tt.method, "/test"+tt.queryParams, body)
			if tt.contentType != "" {
				req.Header.Set("Content-Type", tt.contentType)
			}
			if tt.body != "" {
				req.ContentLength = int64(len(tt.body))
			}

			rr := httptest.NewRecorder()

			nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})

			middleware := RequestValidation(tt.config)
			middleware(nextHandler).ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code, "Response body: %s", rr.Body.String())

			if tt.checkResponse != nil {
				tt.checkResponse(t, rr)
			}
		})
	}
}

func TestRequestValidation_BodyAvailableAfterValidation(t *testing.T) {
	cfg := DefaultRequestValidationConfig()
	originalBody := `{"name": "test"}`

	req := httptest.NewRequest(http.MethodPost, "/test", strings.NewReader(originalBody))
	req.Header.Set("Content-Type", "application/json")
	req.ContentLength = int64(len(originalBody))

	rr := httptest.NewRecorder()

	var capturedBody []byte
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedBody, _ = io.ReadAll(r.Body)
		w.WriteHeader(http.StatusOK)
	})

	middleware := RequestValidation(cfg)
	middleware(nextHandler).ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, originalBody, string(capturedBody))
}

func TestContainsXSSInJSON(t *testing.T) {
	tests := []struct {
		name     string
		json     string
		expected bool
	}{
		{
			name:     "clean simple object",
			json:     `{"name": "John"}`,
			expected: false,
		},
		{
			name:     "script tag in string",
			json:     `{"name": "<script>alert(1)</script>"}`,
			expected: true,
		},
		{
			name:     "nested XSS",
			json:     `{"outer": {"inner": "<img src=x onerror=alert(1)>"}}`,
			expected: true,
		},
		{
			name:     "XSS in array",
			json:     `{"items": ["safe", "<script>bad</script>"]}`,
			expected: true,
		},
		{
			name:     "clean nested object",
			json:     `{"user": {"name": "John", "age": 30}}`,
			expected: false,
		},
		{
			name:     "clean array",
			json:     `{"tags": ["one", "two", "three"]}`,
			expected: false,
		},
		{
			name:     "invalid JSON returns false",
			json:     `{invalid}`,
			expected: false,
		},
		{
			name:     "number value",
			json:     `{"count": 42}`,
			expected: false,
		},
		{
			name:     "boolean value",
			json:     `{"active": true}`,
			expected: false,
		},
		{
			name:     "null value",
			json:     `{"data": null}`,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := containsXSSInJSON([]byte(tt.json))
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestContainsXSSInValue(t *testing.T) {
	tests := []struct {
		name     string
		value    any
		expected bool
	}{
		{
			name:     "clean string",
			value:    "hello world",
			expected: false,
		},
		{
			name:     "XSS string",
			value:    "<script>alert(1)</script>",
			expected: true,
		},
		{
			name:     "number",
			value:    float64(42),
			expected: false,
		},
		{
			name:     "boolean",
			value:    true,
			expected: false,
		},
		{
			name:     "nil",
			value:    nil,
			expected: false,
		},
		{
			name:     "clean map",
			value:    map[string]any{"key": "value"},
			expected: false,
		},
		{
			name:     "map with XSS",
			value:    map[string]any{"key": "<script>bad</script>"},
			expected: true,
		},
		{
			name:     "clean slice",
			value:    []any{"one", "two"},
			expected: false,
		},
		{
			name:     "slice with XSS",
			value:    []any{"one", "<script>bad</script>"},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := containsXSSInValue(tt.value)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestValidateUUID(t *testing.T) {
	tests := []struct {
		name           string
		paramNames     []string
		expectedStatus int
	}{
		{
			name:           "no parameters to validate",
			paramNames:     []string{},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "parameter name provided but empty value",
			paramNames:     []string{"id"},
			expectedStatus: http.StatusOK, // Empty param is allowed to pass through
		},
		{
			name:           "multiple parameters to validate",
			paramNames:     []string{"id", "tenantId", "workflowId"},
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			rr := httptest.NewRecorder()

			nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})

			middleware := ValidateUUID(tt.paramNames...)
			middleware(nextHandler).ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
		})
	}
}

func TestValidateUUID_Integration(t *testing.T) {
	// This test validates the middleware initialization and execution flow
	// The actual UUID extraction is done via chi.URLParam in production

	middleware := ValidateUUID("id", "workflowId")

	req := httptest.NewRequest(http.MethodGet, "/test/123/workflow/456", nil)
	rr := httptest.NewRecorder()

	called := false
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	})

	middleware(nextHandler).ServeHTTP(rr, req)

	// Since extractPathParam returns empty string, validation is skipped
	assert.True(t, called, "next handler should be called")
	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestSanitizeQueryParams(t *testing.T) {
	tests := []struct {
		name            string
		queryParams     string
		checkRequest    func(*testing.T, *http.Request)
		expectedStatus  int
	}{
		{
			name:           "no query params",
			queryParams:    "",
			expectedStatus: http.StatusOK,
			checkRequest: func(t *testing.T, r *http.Request) {
				assert.Empty(t, r.URL.RawQuery)
			},
		},
		{
			name:           "simple query params",
			queryParams:    "?name=test&value=123",
			expectedStatus: http.StatusOK,
			checkRequest: func(t *testing.T, r *http.Request) {
				assert.Contains(t, r.URL.RawQuery, "name=test")
				assert.Contains(t, r.URL.RawQuery, "value=123")
			},
		},
		{
			name:           "query params with multiple values",
			queryParams:    "?tag=one&tag=two",
			expectedStatus: http.StatusOK,
			checkRequest: func(t *testing.T, r *http.Request) {
				values := r.URL.Query()
				assert.Len(t, values["tag"], 2)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/test"+tt.queryParams, nil)
			rr := httptest.NewRecorder()

			var capturedReq *http.Request
			nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				capturedReq = r
				w.WriteHeader(http.StatusOK)
			})

			middleware := SanitizeQueryParams()
			middleware(nextHandler).ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)

			if tt.checkRequest != nil && capturedReq != nil {
				tt.checkRequest(t, capturedReq)
			}
		})
	}
}

func TestEncodeQuery(t *testing.T) {
	tests := []struct {
		name     string
		values   map[string][]string
		contains []string
	}{
		{
			name:     "empty map",
			values:   map[string][]string{},
			contains: []string{},
		},
		{
			name: "single key value",
			values: map[string][]string{
				"key": {"value"},
			},
			contains: []string{"key=value"},
		},
		{
			name: "multiple values for same key",
			values: map[string][]string{
				"tag": {"one", "two"},
			},
			contains: []string{"tag=one", "tag=two"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := encodeQuery(tt.values)
			for _, expected := range tt.contains {
				assert.Contains(t, result, expected)
			}
		})
	}
}

func TestValidationError(t *testing.T) {
	err := errValidationFailed
	assert.Equal(t, "validation failed", err.Error())
}

func TestRequestValidation_BodyExceedsLimitDuringRead(t *testing.T) {
	cfg := RequestValidationConfig{
		MaxBodySize:  50,
		ValidateJSON: true,
	}

	// Create a body that's larger than max size with proper ContentLength
	largeBody := strings.Repeat("x", 100)
	req := httptest.NewRequest(http.MethodPost, "/test", strings.NewReader(largeBody))
	req.Header.Set("Content-Type", "application/json")
	req.ContentLength = int64(len(largeBody)) // Set content length to trigger body validation

	rr := httptest.NewRecorder()

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	middleware := RequestValidation(cfg)
	middleware(nextHandler).ServeHTTP(rr, req)

	// Should fail because body exceeds limit
	assert.Equal(t, http.StatusUnprocessableEntity, rr.Code)
}

func TestRespondValidationError(t *testing.T) {
	rr := httptest.NewRecorder()
	respondValidationError(rr, "test error message", "test_field")

	assert.Equal(t, http.StatusUnprocessableEntity, rr.Code)
	assert.Contains(t, rr.Body.String(), "test error message")
}

func TestExtractPathParam(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/test/123", nil)
	result := extractPathParam(req, "id")
	// The function always returns empty string by design
	assert.Empty(t, result)
}

func TestRequestValidation_NilBody(t *testing.T) {
	cfg := DefaultRequestValidationConfig()

	req := httptest.NewRequest(http.MethodPost, "/test", nil)
	req.Header.Set("Content-Type", "application/json")
	req.Body = nil
	req.ContentLength = 0

	rr := httptest.NewRecorder()

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	middleware := RequestValidation(cfg)
	middleware(nextHandler).ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestRequestValidation_ContentLengthZero(t *testing.T) {
	cfg := DefaultRequestValidationConfig()

	req := httptest.NewRequest(http.MethodPost, "/test", bytes.NewReader([]byte{}))
	req.Header.Set("Content-Type", "application/json")
	req.ContentLength = 0

	rr := httptest.NewRecorder()

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	middleware := RequestValidation(cfg)
	middleware(nextHandler).ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
}
