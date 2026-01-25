package http

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClient_NewClient(t *testing.T) {
	t.Run("default configuration", func(t *testing.T) {
		client := NewClient()
		assert.NotNil(t, client)
		assert.NotNil(t, client.httpClient)
		assert.NotNil(t, client.retryConfig)
	})

	t.Run("with options", func(t *testing.T) {
		client := NewClient(
			WithTimeout(5*time.Second),
			WithBaseURL("https://api.example.com"),
			WithHeader("X-Custom", "value"),
		)

		assert.Equal(t, 5*time.Second, client.httpClient.Timeout)
		assert.Equal(t, "https://api.example.com", client.baseURL)
		assert.Equal(t, "value", client.defaultHeaders["X-Custom"])
	})

	t.Run("with multiple headers", func(t *testing.T) {
		headers := map[string]string{
			"X-Header-1": "value1",
			"X-Header-2": "value2",
		}
		client := NewClient(WithHeaders(headers))

		assert.Equal(t, "value1", client.defaultHeaders["X-Header-1"])
		assert.Equal(t, "value2", client.defaultHeaders["X-Header-2"])
	})
}

func TestClient_Do(t *testing.T) {
	t.Run("successful GET request", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodGet, r.Method)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"status": "ok"}`))
		}))
		defer server.Close()

		client := NewClient()
		resp, err := client.Get(context.Background(), server.URL)

		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.True(t, resp.IsSuccess())
	})

	t.Run("successful POST request", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodPost, r.Method)
			assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
			w.WriteHeader(http.StatusCreated)
		}))
		defer server.Close()

		client := NewClient()
		resp, err := client.Post(context.Background(), server.URL, map[string]string{"key": "value"})

		require.NoError(t, err)
		assert.Equal(t, http.StatusCreated, resp.StatusCode)
	})

	t.Run("request with headers", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "custom-value", r.Header.Get("X-Custom-Header"))
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		client := NewClient(WithHeader("X-Custom-Header", "custom-value"))
		_, err := client.Get(context.Background(), server.URL)

		require.NoError(t, err)
	})

	t.Run("request with query params", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "value1", r.URL.Query().Get("param1"))
			assert.Equal(t, "value2", r.URL.Query().Get("param2"))
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		client := NewClient()
		_, err := client.Get(context.Background(), server.URL,
			WithQueryParam("param1", "value1"),
			WithQueryParam("param2", "value2"),
		)

		require.NoError(t, err)
	})

	t.Run("HTTP error response", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(`{"error": "bad request"}`))
		}))
		defer server.Close()

		client := NewClient(WithRetryConfig(NoRetry()))
		resp, err := client.Get(context.Background(), server.URL)

		require.Error(t, err)
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
		assert.True(t, resp.IsError())

		var httpErr *HTTPError
		assert.ErrorAs(t, err, &httpErr)
		assert.Equal(t, http.StatusBadRequest, httpErr.StatusCode)
	})

	t.Run("context cancellation", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(100 * time.Millisecond)
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		client := NewClient()
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		_, err := client.Get(ctx, server.URL)
		assert.Error(t, err)
	})
}

func TestClient_Retry(t *testing.T) {
	t.Run("retries on 500 error", func(t *testing.T) {
		var attempts int32

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			count := atomic.AddInt32(&attempts, 1)
			if count < 3 {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		retryConfig := NewRetryConfigBuilder().
			WithMaxRetries(3).
			WithBaseDelay(1 * time.Millisecond).
			Build()

		client := NewClient(WithRetryConfig(retryConfig))
		resp, err := client.Get(context.Background(), server.URL)

		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Equal(t, int32(3), atomic.LoadInt32(&attempts))
	})

	t.Run("no retry on 400 error", func(t *testing.T) {
		var attempts int32

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			atomic.AddInt32(&attempts, 1)
			w.WriteHeader(http.StatusBadRequest)
		}))
		defer server.Close()

		retryConfig := NewRetryConfigBuilder().
			WithMaxRetries(3).
			Build()

		client := NewClient(WithRetryConfig(retryConfig))
		_, _ = client.Get(context.Background(), server.URL)

		assert.Equal(t, int32(1), atomic.LoadInt32(&attempts))
	})

	t.Run("max retries exceeded", func(t *testing.T) {
		var attempts int32

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			atomic.AddInt32(&attempts, 1)
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer server.Close()

		retryConfig := NewRetryConfigBuilder().
			WithMaxRetries(3).
			WithBaseDelay(1 * time.Millisecond).
			Build()

		client := NewClient(WithRetryConfig(retryConfig))
		_, err := client.Get(context.Background(), server.URL)

		assert.Error(t, err)
		assert.Equal(t, int32(3), atomic.LoadInt32(&attempts))
	})
}

func TestClient_CircuitBreaker(t *testing.T) {
	t.Run("circuit opens after failures", func(t *testing.T) {
		var attempts int32

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			atomic.AddInt32(&attempts, 1)
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer server.Close()

		cb := NewCircuitBreaker(&CircuitBreakerConfig{
			FailureThreshold: 3,
			Timeout:          1 * time.Second,
		})

		client := NewClient(
			WithCircuitBreaker(cb),
			WithRetryConfig(NoRetry()),
		)

		// Make requests until circuit opens
		for i := 0; i < 5; i++ {
			_, _ = client.Get(context.Background(), server.URL)
		}

		// Circuit should be open now
		assert.Equal(t, StateOpen, cb.State())

		// Requests should be blocked
		_, err := client.Get(context.Background(), server.URL)
		assert.ErrorIs(t, err, ErrCircuitOpen)
	})

	t.Run("circuit closes after success", func(t *testing.T) {
		failureCount := int32(3)
		var attempts int32

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			count := atomic.AddInt32(&attempts, 1)
			if count <= failureCount {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		cb := NewCircuitBreaker(&CircuitBreakerConfig{
			FailureThreshold:   3,
			SuccessThreshold:   1,
			Timeout:            10 * time.Millisecond,
			HalfOpenMaxAllowed: 1,
		})

		client := NewClient(
			WithCircuitBreaker(cb),
			WithRetryConfig(NoRetry()),
		)

		// Trigger failures to open circuit
		for i := 0; i < 3; i++ {
			_, _ = client.Get(context.Background(), server.URL)
		}
		assert.Equal(t, StateOpen, cb.State())

		// Wait for timeout
		time.Sleep(20 * time.Millisecond)

		// Should transition to half-open and then closed
		_, err := client.Get(context.Background(), server.URL)
		require.NoError(t, err)
		assert.Equal(t, StateClosed, cb.State())
	})
}

func TestClient_Methods(t *testing.T) {
	tests := []struct {
		name       string
		method     string
		clientCall func(*Client, context.Context, string) (*Response, error)
	}{
		{
			name:   "GET",
			method: http.MethodGet,
			clientCall: func(c *Client, ctx context.Context, url string) (*Response, error) {
				return c.Get(ctx, url)
			},
		},
		{
			name:   "DELETE",
			method: http.MethodDelete,
			clientCall: func(c *Client, ctx context.Context, url string) (*Response, error) {
				return c.Delete(ctx, url)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, tt.method, r.Method)
				w.WriteHeader(http.StatusOK)
			}))
			defer server.Close()

			client := NewClient()
			resp, err := tt.clientCall(client, context.Background(), server.URL)

			require.NoError(t, err)
			assert.Equal(t, http.StatusOK, resp.StatusCode)
		})
	}
}

func TestClient_MethodsWithBody(t *testing.T) {
	tests := []struct {
		name       string
		method     string
		clientCall func(*Client, context.Context, string, any) (*Response, error)
	}{
		{
			name:   "POST",
			method: http.MethodPost,
			clientCall: func(c *Client, ctx context.Context, url string, body any) (*Response, error) {
				return c.Post(ctx, url, body)
			},
		},
		{
			name:   "PUT",
			method: http.MethodPut,
			clientCall: func(c *Client, ctx context.Context, url string, body any) (*Response, error) {
				return c.Put(ctx, url, body)
			},
		},
		{
			name:   "PATCH",
			method: http.MethodPatch,
			clientCall: func(c *Client, ctx context.Context, url string, body any) (*Response, error) {
				return c.Patch(ctx, url, body)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, tt.method, r.Method)
				assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
				w.WriteHeader(http.StatusOK)
			}))
			defer server.Close()

			client := NewClient()
			resp, err := tt.clientCall(client, context.Background(), server.URL, map[string]string{"key": "value"})

			require.NoError(t, err)
			assert.Equal(t, http.StatusOK, resp.StatusCode)
		})
	}
}

func TestResponse_JSON(t *testing.T) {
	resp := &Response{
		StatusCode: 200,
		Body:       []byte(`{"key": "value", "number": 42}`),
	}

	var data map[string]any
	err := resp.JSON(&data)

	require.NoError(t, err)
	assert.Equal(t, "value", data["key"])
	assert.Equal(t, float64(42), data["number"])
}

func TestResponse_String(t *testing.T) {
	resp := &Response{
		Body: []byte("Hello, World!"),
	}

	assert.Equal(t, "Hello, World!", resp.String())
}

func TestBuildURL(t *testing.T) {
	tests := []struct {
		name        string
		baseURL     string
		path        string
		queryParams map[string]string
		expected    string
	}{
		{
			name:     "full URL without base",
			baseURL:  "",
			path:     "https://api.example.com/resource",
			expected: "https://api.example.com/resource",
		},
		{
			name:     "path with base URL",
			baseURL:  "https://api.example.com",
			path:     "/resource",
			expected: "https://api.example.com/resource",
		},
		{
			name:        "with query params",
			baseURL:     "https://api.example.com",
			path:        "/search",
			queryParams: map[string]string{"q": "test", "limit": "10"},
			expected:    "https://api.example.com/search?limit=10&q=test",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewClient(WithBaseURL(tt.baseURL))
			url, err := client.buildURL(tt.path, tt.queryParams)

			require.NoError(t, err)
			assert.Equal(t, tt.expected, url)
		})
	}
}
