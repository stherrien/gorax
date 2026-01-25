package integration

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIntegrationType_Valid(t *testing.T) {
	tests := []struct {
		name     string
		intType  IntegrationType
		expected bool
	}{
		{"http", TypeHTTP, true},
		{"webhook", TypeWebhook, true},
		{"api", TypeAPI, true},
		{"custom", TypeCustom, true},
		{"plugin", TypePlugin, true},
		{"invalid", IntegrationType("invalid"), false},
		{"empty", IntegrationType(""), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.intType.Valid())
		})
	}
}

func TestFieldType_Valid(t *testing.T) {
	tests := []struct {
		name      string
		fieldType FieldType
		expected  bool
	}{
		{"string", FieldTypeString, true},
		{"number", FieldTypeNumber, true},
		{"integer", FieldTypeInteger, true},
		{"boolean", FieldTypeBoolean, true},
		{"array", FieldTypeArray, true},
		{"object", FieldTypeObject, true},
		{"secret", FieldTypeSecret, true},
		{"invalid", FieldType("invalid"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.fieldType.Valid())
		})
	}
}

func TestCredentialType_Valid(t *testing.T) {
	tests := []struct {
		name     string
		credType CredentialType
		expected bool
	}{
		{"api_key", CredTypeAPIKey, true},
		{"bearer_token", CredTypeBearerToken, true},
		{"basic_auth", CredTypeBasicAuth, true},
		{"oauth2", CredTypeOAuth2, true},
		{"custom", CredTypeCustom, true},
		{"invalid", CredentialType("invalid"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.credType.Valid())
		})
	}
}

func TestJSONMap(t *testing.T) {
	t.Run("Get", func(t *testing.T) {
		m := JSONMap{"key": "value", "num": 42}

		v, ok := m.Get("key")
		assert.True(t, ok)
		assert.Equal(t, "value", v)

		_, ok = m.Get("missing")
		assert.False(t, ok)
	})

	t.Run("GetString", func(t *testing.T) {
		m := JSONMap{"key": "value", "num": 42}

		v, ok := m.GetString("key")
		assert.True(t, ok)
		assert.Equal(t, "value", v)

		_, ok = m.GetString("num")
		assert.False(t, ok)
	})

	t.Run("GetInt", func(t *testing.T) {
		m := JSONMap{"int": 42, "float": 3.14, "str": "text"}

		v, ok := m.GetInt("int")
		assert.True(t, ok)
		assert.Equal(t, 42, v)

		v, ok = m.GetInt("float")
		assert.True(t, ok)
		assert.Equal(t, 3, v)

		_, ok = m.GetInt("str")
		assert.False(t, ok)
	})

	t.Run("GetBool", func(t *testing.T) {
		m := JSONMap{"flag": true, "str": "true"}

		v, ok := m.GetBool("flag")
		assert.True(t, ok)
		assert.True(t, v)

		_, ok = m.GetBool("str")
		assert.False(t, ok)
	})

	t.Run("Merge", func(t *testing.T) {
		m1 := JSONMap{"a": 1, "b": 2}
		m2 := JSONMap{"b": 3, "c": 4}

		merged := m1.Merge(m2)

		assert.Equal(t, 1, merged["a"])
		assert.Equal(t, 3, merged["b"]) // m2 value should override
		assert.Equal(t, 4, merged["c"])
	})

	t.Run("Value and Scan", func(t *testing.T) {
		m := JSONMap{"key": "value", "nested": JSONMap{"inner": 1}}

		// Test Value
		val, err := m.Value()
		require.NoError(t, err)

		// Test Scan from bytes
		var scanned JSONMap
		err = scanned.Scan(val)
		require.NoError(t, err)
		assert.Equal(t, "value", scanned["key"])

		// Test Scan from string
		var scanned2 JSONMap
		err = scanned2.Scan(`{"key": "value"}`)
		require.NoError(t, err)
		assert.Equal(t, "value", scanned2["key"])

		// Test Scan nil
		var scanned3 JSONMap
		err = scanned3.Scan(nil)
		require.NoError(t, err)
		assert.Nil(t, scanned3)
	})

	t.Run("Nil map operations", func(t *testing.T) {
		var m JSONMap

		_, ok := m.Get("key")
		assert.False(t, ok)

		_, ok = m.GetString("key")
		assert.False(t, ok)

		_, ok = m.GetInt("key")
		assert.False(t, ok)

		_, ok = m.GetBool("key")
		assert.False(t, ok)

		val, err := m.Value()
		require.NoError(t, err)
		assert.Equal(t, []byte("{}"), val)
	})
}

func TestResult(t *testing.T) {
	t.Run("NewSuccessResult", func(t *testing.T) {
		data := map[string]string{"key": "value"}
		result := NewSuccessResult(data, 100)

		assert.True(t, result.Success)
		assert.Equal(t, data, result.Data)
		assert.Equal(t, int64(100), result.Duration)
		assert.NotZero(t, result.ExecutedAt)
	})

	t.Run("NewErrorResult", func(t *testing.T) {
		err := ErrTimeout
		result := NewErrorResult(err, "TIMEOUT", 50)

		assert.False(t, result.Success)
		assert.Equal(t, "operation timed out", result.Error)
		assert.Equal(t, "TIMEOUT", result.ErrorCode)
		assert.Equal(t, int64(50), result.Duration)
	})

	t.Run("NewErrorResult with nil error", func(t *testing.T) {
		result := NewErrorResult(nil, "CODE", 10)

		assert.False(t, result.Success)
		assert.Empty(t, result.Error)
	})
}

func TestConfig(t *testing.T) {
	t.Run("JSON serialization", func(t *testing.T) {
		now := time.Now().UTC()
		config := &Config{
			ID:          "test-id",
			Name:        "test-config",
			Type:        TypeHTTP,
			Description: "Test configuration",
			Version:     "1.0.0",
			Enabled:     true,
			Settings:    JSONMap{"url": "https://example.com"},
			Credentials: &Credentials{
				ID:   "cred-id",
				Type: CredTypeAPIKey,
				Data: JSONMap{"key": "secret"},
			},
			CreatedAt: now,
			UpdatedAt: now,
		}

		// Serialize
		data, err := json.Marshal(config)
		require.NoError(t, err)

		// Deserialize
		var decoded Config
		err = json.Unmarshal(data, &decoded)
		require.NoError(t, err)

		assert.Equal(t, config.ID, decoded.ID)
		assert.Equal(t, config.Name, decoded.Name)
		assert.Equal(t, config.Type, decoded.Type)
		assert.True(t, decoded.Enabled)
	})
}

func TestSchema(t *testing.T) {
	schema := &Schema{
		ConfigSpec: map[string]FieldSpec{
			"url": {
				Name:        "url",
				Type:        FieldTypeString,
				Required:    true,
				Description: "Target URL",
			},
			"timeout": {
				Name:        "timeout",
				Type:        FieldTypeInteger,
				Required:    false,
				Description: "Timeout in seconds",
			},
		},
		InputSpec: map[string]FieldSpec{
			"data": {
				Name: "data",
				Type: FieldTypeObject,
			},
		},
		OutputSpec: map[string]FieldSpec{
			"response": {
				Name: "response",
				Type: FieldTypeObject,
			},
		},
	}

	assert.Len(t, schema.ConfigSpec, 2)
	assert.Len(t, schema.InputSpec, 1)
	assert.Len(t, schema.OutputSpec, 1)
	assert.True(t, schema.ConfigSpec["url"].Required)
}

func TestMetadata(t *testing.T) {
	metadata := &Metadata{
		Name:        "test-integration",
		DisplayName: "Test Integration",
		Description: "A test integration",
		Version:     "1.0.0",
		Author:      "Test Author",
		License:     "MIT",
		Homepage:    "https://example.com",
		Tags:        []string{"test", "example"},
		Category:    "testing",
		Permissions: []string{"http:read", "http:write"},
	}

	assert.Equal(t, "test-integration", metadata.Name)
	assert.Len(t, metadata.Tags, 2)
	assert.Len(t, metadata.Permissions, 2)
}
