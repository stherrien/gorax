package integration

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewHTTPIntegration(t *testing.T) {
	t.Run("creates with default logger", func(t *testing.T) {
		integ := NewHTTPIntegration(nil)

		require.NotNil(t, integ)
		assert.Equal(t, "http", integ.Name())
		assert.Equal(t, TypeHTTP, integ.Type())
		assert.NotNil(t, integ.client)
		assert.NotNil(t, integ.logger)
	})

	t.Run("creates with custom logger", func(t *testing.T) {
		logger := slog.Default()
		integ := NewHTTPIntegration(logger)

		require.NotNil(t, integ)
		assert.Equal(t, logger, integ.logger)
	})

	t.Run("has proper metadata", func(t *testing.T) {
		integ := NewHTTPIntegration(nil)
		metadata := integ.GetMetadata()

		assert.Equal(t, "http", metadata.Name)
		assert.Equal(t, "HTTP Request", metadata.DisplayName)
		assert.Equal(t, "1.0.0", metadata.Version)
		assert.Equal(t, "networking", metadata.Category)
	})

	t.Run("has proper schema", func(t *testing.T) {
		integ := NewHTTPIntegration(nil)
		schema := integ.GetSchema()

		require.NotNil(t, schema)
		assert.Contains(t, schema.ConfigSpec, "url")
		assert.Contains(t, schema.ConfigSpec, "method")
		assert.Contains(t, schema.ConfigSpec, "headers")
		assert.Contains(t, schema.ConfigSpec, "body")
		assert.Contains(t, schema.ConfigSpec, "timeout")
	})
}

func TestHTTPIntegration_Execute(t *testing.T) {
	t.Run("executes GET request", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodGet, r.Method)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"status": "ok"}`))
		}))
		defer server.Close()

		integ := NewHTTPIntegration(nil)
		config := &Config{
			Name:    "test",
			Type:    TypeHTTP,
			Enabled: true,
			Settings: JSONMap{
				"url":    server.URL,
				"method": "GET",
			},
		}

		result, err := integ.Execute(context.Background(), config, nil)

		require.NoError(t, err)
		require.NotNil(t, result)
		assert.True(t, result.Success)
		assert.Equal(t, http.StatusOK, result.StatusCode)
	})

	t.Run("executes POST request with body", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodPost, r.Method)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{"id": "123"}`))
		}))
		defer server.Close()

		integ := NewHTTPIntegration(nil)
		config := &Config{
			Name:    "test",
			Type:    TypeHTTP,
			Enabled: true,
			Settings: JSONMap{
				"url":    server.URL,
				"method": "POST",
				"body": map[string]any{
					"name": "test",
				},
			},
		}

		result, err := integ.Execute(context.Background(), config, nil)

		require.NoError(t, err)
		require.NotNil(t, result)
		assert.True(t, result.Success)
		assert.Equal(t, http.StatusCreated, result.StatusCode)
	})

	t.Run("includes custom headers", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "Bearer token123", r.Header.Get("Authorization"))
			assert.Equal(t, "custom-value", r.Header.Get("X-Custom-Header"))
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		integ := NewHTTPIntegration(nil)
		config := &Config{
			Name:    "test",
			Type:    TypeHTTP,
			Enabled: true,
			Settings: JSONMap{
				"url":    server.URL,
				"method": "GET",
				"headers": map[string]any{
					"Authorization":   "Bearer token123",
					"X-Custom-Header": "custom-value",
				},
			},
		}

		result, err := integ.Execute(context.Background(), config, nil)

		require.NoError(t, err)
		assert.True(t, result.Success)
	})

	t.Run("includes query parameters", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "value1", r.URL.Query().Get("param1"))
			assert.Equal(t, "value2", r.URL.Query().Get("param2"))
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		integ := NewHTTPIntegration(nil)
		config := &Config{
			Name:    "test",
			Type:    TypeHTTP,
			Enabled: true,
			Settings: JSONMap{
				"url":    server.URL,
				"method": "GET",
				"query_params": map[string]any{
					"param1": "value1",
					"param2": "value2",
				},
			},
		}

		result, err := integ.Execute(context.Background(), config, nil)

		require.NoError(t, err)
		assert.True(t, result.Success)
	})

	t.Run("handles non-2xx response", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`{"error": "not found"}`))
		}))
		defer server.Close()

		integ := NewHTTPIntegration(nil)
		config := &Config{
			Name:    "test",
			Type:    TypeHTTP,
			Enabled: true,
			Settings: JSONMap{
				"url":    server.URL,
				"method": "GET",
			},
		}

		result, err := integ.Execute(context.Background(), config, nil)

		require.Error(t, err)
		require.NotNil(t, result)
		assert.False(t, result.Success)
	})

	t.Run("uses custom success codes", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusAccepted)
		}))
		defer server.Close()

		integ := NewHTTPIntegration(nil)
		config := &Config{
			Name:    "test",
			Type:    TypeHTTP,
			Enabled: true,
			Settings: JSONMap{
				"url":           server.URL,
				"method":        "POST",
				"success_codes": []any{float64(200), float64(202)},
			},
		}

		result, err := integ.Execute(context.Background(), config, nil)

		require.NoError(t, err)
		assert.True(t, result.Success)
	})

	t.Run("processes URL template", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/users/123", r.URL.Path)
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		integ := NewHTTPIntegration(nil)
		config := &Config{
			Name:    "test",
			Type:    TypeHTTP,
			Enabled: true,
			Settings: JSONMap{
				"url":    server.URL + "/users/{{.user_id}}",
				"method": "GET",
			},
		}

		result, err := integ.Execute(context.Background(), config, JSONMap{"user_id": "123"})

		require.NoError(t, err)
		assert.True(t, result.Success)
	})

	t.Run("processes body template", func(t *testing.T) {
		var receivedBody map[string]any
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			json.NewDecoder(r.Body).Decode(&receivedBody)
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		integ := NewHTTPIntegration(nil)
		config := &Config{
			Name:    "test",
			Type:    TypeHTTP,
			Enabled: true,
			Settings: JSONMap{
				"url":           server.URL,
				"method":        "POST",
				"body_template": `{"name": "{{.name}}"}`,
			},
		}

		result, err := integ.Execute(context.Background(), config, JSONMap{"name": "John"})

		require.NoError(t, err)
		assert.True(t, result.Success)
	})

	t.Run("handles different response types", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/plain")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("plain text response"))
		}))
		defer server.Close()

		integ := NewHTTPIntegration(nil)
		config := &Config{
			Name:    "test",
			Type:    TypeHTTP,
			Enabled: true,
			Settings: JSONMap{
				"url":           server.URL,
				"method":        "GET",
				"response_type": "text",
			},
		}

		result, err := integ.Execute(context.Background(), config, nil)

		require.NoError(t, err)
		assert.True(t, result.Success)
		data := result.Data.(JSONMap)
		assert.Equal(t, "plain text response", data["body"])
	})
}

func TestHTTPIntegration_Validate(t *testing.T) {
	integ := NewHTTPIntegration(nil)

	t.Run("validates valid config", func(t *testing.T) {
		config := &Config{
			Name:    "test",
			Type:    TypeHTTP,
			Enabled: true,
			Settings: JSONMap{
				"url":    "https://api.example.com/data",
				"method": "GET",
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
			Name:    "test",
			Type:    TypeHTTP,
			Enabled: true,
			Settings: JSONMap{
				"method": "GET",
			},
		}

		err := integ.Validate(config)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "URL")
	})

	t.Run("fails with missing method", func(t *testing.T) {
		config := &Config{
			Name:    "test",
			Type:    TypeHTTP,
			Enabled: true,
			Settings: JSONMap{
				"url": "https://api.example.com",
			},
		}

		err := integ.Validate(config)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "method")
	})

	t.Run("fails with invalid method", func(t *testing.T) {
		config := &Config{
			Name:    "test",
			Type:    TypeHTTP,
			Enabled: true,
			Settings: JSONMap{
				"url":    "https://api.example.com",
				"method": "INVALID",
			},
		}

		err := integ.Validate(config)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "method")
	})

	t.Run("accepts all valid methods", func(t *testing.T) {
		methods := []string{"GET", "POST", "PUT", "DELETE", "PATCH", "HEAD", "OPTIONS"}

		for _, method := range methods {
			config := &Config{
				Name:    "test",
				Type:    TypeHTTP,
				Enabled: true,
				Settings: JSONMap{
					"url":    "https://api.example.com",
					"method": method,
				},
			}

			err := integ.Validate(config)
			require.NoError(t, err, "expected no error for method: %s", method)
		}
	})

	t.Run("skips URL validation for templates", func(t *testing.T) {
		config := &Config{
			Name:    "test",
			Type:    TypeHTTP,
			Enabled: true,
			Settings: JSONMap{
				"url":    "{{.base_url}}/api/users",
				"method": "GET",
			},
		}

		err := integ.Validate(config)
		require.NoError(t, err)
	})
}

func TestHTTPIntegration_isSuccessStatus(t *testing.T) {
	integ := NewHTTPIntegration(nil)

	t.Run("default success range 2xx", func(t *testing.T) {
		assert.True(t, integ.isSuccessStatus(200, nil))
		assert.True(t, integ.isSuccessStatus(201, nil))
		assert.True(t, integ.isSuccessStatus(204, nil))
		assert.True(t, integ.isSuccessStatus(299, nil))
		assert.False(t, integ.isSuccessStatus(199, nil))
		assert.False(t, integ.isSuccessStatus(300, nil))
		assert.False(t, integ.isSuccessStatus(404, nil))
		assert.False(t, integ.isSuccessStatus(500, nil))
	})

	t.Run("custom success codes", func(t *testing.T) {
		successCodes := []int{200, 201, 404}
		assert.True(t, integ.isSuccessStatus(200, successCodes))
		assert.True(t, integ.isSuccessStatus(201, successCodes))
		assert.True(t, integ.isSuccessStatus(404, successCodes))
		assert.False(t, integ.isSuccessStatus(202, successCodes))
		assert.False(t, integ.isSuccessStatus(500, successCodes))
	})
}

func TestNewHTTPIntegrationFromRequest(t *testing.T) {
	t.Run("creates integration with config", func(t *testing.T) {
		integ, config := NewHTTPIntegrationFromRequest("GET", "https://api.example.com/users")

		require.NotNil(t, integ)
		require.NotNil(t, config)
		assert.Equal(t, "GET", config.Settings["method"])
		assert.Equal(t, "https://api.example.com/users", config.Settings["url"])
	})

	t.Run("applies options", func(t *testing.T) {
		integ, config := NewHTTPIntegrationFromRequest(
			"POST",
			"https://api.example.com/users",
			WithHTTPHeaders(map[string]string{"Authorization": "Bearer token"}),
			WithHTTPBody(map[string]string{"name": "John"}),
			WithHTTPQueryParams(map[string]string{"limit": "10"}),
			WithHTTPTimeout(30),
		)

		require.NotNil(t, integ)
		require.NotNil(t, config)

		headers := config.Settings["headers"].(map[string]string)
		assert.Equal(t, "Bearer token", headers["Authorization"])

		body := config.Settings["body"].(map[string]string)
		assert.Equal(t, "John", body["name"])

		params := config.Settings["query_params"].(map[string]string)
		assert.Equal(t, "10", params["limit"])

		assert.Equal(t, 30, config.Settings["timeout"])
	})
}

func TestQuickHTTP(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status": "ok"}`))
	}))
	defer server.Close()

	t.Run("QuickHTTP", func(t *testing.T) {
		result, err := QuickHTTP(context.Background(), "GET", server.URL)
		require.NoError(t, err)
		assert.True(t, result.Success)
	})

	t.Run("QuickGet", func(t *testing.T) {
		result, err := QuickGet(context.Background(), server.URL)
		require.NoError(t, err)
		assert.True(t, result.Success)
	})

	t.Run("QuickPost", func(t *testing.T) {
		result, err := QuickPost(context.Background(), server.URL, map[string]string{"key": "value"})
		require.NoError(t, err)
		assert.True(t, result.Success)
	})
}

func TestHTTPIntegrationConfig(t *testing.T) {
	config := HTTPIntegrationConfig{
		URL:          "https://api.example.com",
		Method:       "POST",
		Headers:      map[string]string{"Content-Type": "application/json"},
		QueryParams:  map[string]string{"key": "value"},
		Body:         map[string]string{"data": "test"},
		BodyTemplate: `{"name": "{{.name}}"}`,
		Timeout:      30,
		ResponseType: "json",
		SuccessCodes: []int{200, 201},
		ExtractPath:  "$.data",
	}

	data, err := json.Marshal(config)
	require.NoError(t, err)

	var decoded HTTPIntegrationConfig
	require.NoError(t, json.Unmarshal(data, &decoded))

	assert.Equal(t, config.URL, decoded.URL)
	assert.Equal(t, config.Method, decoded.Method)
	assert.Equal(t, config.Timeout, decoded.Timeout)
	assert.Equal(t, config.ResponseType, decoded.ResponseType)
}

func TestHTTPIntegration_buildResult(t *testing.T) {
	integ := NewHTTPIntegration(nil)

	t.Run("builds result with JSON response", func(t *testing.T) {
		// This is an internal method so we test it indirectly through Execute
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("X-Request-ID", "123")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"key": "value"}`))
		}))
		defer server.Close()

		config := &Config{
			Name:    "test",
			Type:    TypeHTTP,
			Enabled: true,
			Settings: JSONMap{
				"url":    server.URL,
				"method": "GET",
			},
		}

		result, err := integ.Execute(context.Background(), config, nil)
		require.NoError(t, err)

		data := result.Data.(JSONMap)
		assert.Equal(t, http.StatusOK, data["status_code"])
		assert.NotNil(t, data["headers"])
		assert.NotNil(t, data["body"])
	})
}

func TestHTTPIntegration_ExecuteTimeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	integ := NewHTTPIntegration(nil)
	config := &Config{
		Name:    "test",
		Type:    TypeHTTP,
		Enabled: true,
		Settings: JSONMap{
			"url":     server.URL,
			"method":  "GET",
			"timeout": 1,
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	result, err := integ.Execute(ctx, config, nil)

	require.Error(t, err)
	require.NotNil(t, result)
	assert.False(t, result.Success)
}
