package integration

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewBaseIntegration(t *testing.T) {
	t.Run("creates integration with defaults", func(t *testing.T) {
		base := NewBaseIntegration("test-integration", TypeHTTP)

		assert.Equal(t, "test-integration", base.Name())
		assert.Equal(t, TypeHTTP, base.Type())
		assert.NotNil(t, base.GetMetadata())
		assert.NotNil(t, base.GetSchema())
	})

	t.Run("has default metadata", func(t *testing.T) {
		base := NewBaseIntegration("my-integration", TypeAPI)
		metadata := base.GetMetadata()

		assert.Equal(t, "my-integration", metadata.Name)
		assert.Equal(t, "1.0.0", metadata.Version)
	})

	t.Run("has empty schema specs", func(t *testing.T) {
		base := NewBaseIntegration("my-integration", TypeAPI)
		schema := base.GetSchema()

		assert.NotNil(t, schema.ConfigSpec)
		assert.NotNil(t, schema.InputSpec)
		assert.NotNil(t, schema.OutputSpec)
		assert.Empty(t, schema.ConfigSpec)
		assert.Empty(t, schema.InputSpec)
		assert.Empty(t, schema.OutputSpec)
	})
}

func TestBaseIntegration_SetSchema(t *testing.T) {
	base := NewBaseIntegration("test", TypeHTTP)

	customSchema := &Schema{
		ConfigSpec: map[string]FieldSpec{
			"url": {Name: "url", Type: FieldTypeString, Required: true},
		},
		InputSpec: map[string]FieldSpec{
			"data": {Name: "data", Type: FieldTypeObject},
		},
		OutputSpec: map[string]FieldSpec{
			"result": {Name: "result", Type: FieldTypeString},
		},
	}

	base.SetSchema(customSchema)

	schema := base.GetSchema()
	assert.Equal(t, customSchema, schema)
	assert.Contains(t, schema.ConfigSpec, "url")
}

func TestBaseIntegration_SetMetadata(t *testing.T) {
	base := NewBaseIntegration("test", TypeHTTP)

	customMetadata := &Metadata{
		Name:        "test",
		DisplayName: "Test Integration",
		Description: "A test integration",
		Version:     "2.0.0",
		Author:      "Test Author",
		Category:    "testing",
	}

	base.SetMetadata(customMetadata)

	metadata := base.GetMetadata()
	assert.Equal(t, customMetadata, metadata)
	assert.Equal(t, "Test Integration", metadata.DisplayName)
	assert.Equal(t, "2.0.0", metadata.Version)
}

func TestBaseIntegration_ValidateConfig(t *testing.T) {
	base := NewBaseIntegration("test", TypeHTTP)

	t.Run("validates valid config", func(t *testing.T) {
		config := &Config{
			Name:     "test-config",
			Type:     TypeHTTP,
			Settings: JSONMap{},
		}

		err := base.ValidateConfig(config)
		require.NoError(t, err)
	})

	t.Run("fails with nil config", func(t *testing.T) {
		err := base.ValidateConfig(nil)
		require.Error(t, err)
	})

	t.Run("fails with empty name", func(t *testing.T) {
		config := &Config{
			Name:     "",
			Type:     TypeHTTP,
			Settings: JSONMap{},
		}

		err := base.ValidateConfig(config)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "name")
	})

	t.Run("fails with invalid type", func(t *testing.T) {
		config := &Config{
			Name:     "test",
			Type:     "invalid-type",
			Settings: JSONMap{},
		}

		err := base.ValidateConfig(config)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "type")
	})
}

func TestValidateSchema(t *testing.T) {
	t.Run("validates data against spec", func(t *testing.T) {
		spec := map[string]FieldSpec{
			"url":  {Name: "url", Type: FieldTypeString, Required: true},
			"port": {Name: "port", Type: FieldTypeInteger, Required: false},
		}

		data := JSONMap{
			"url":  "https://example.com",
			"port": float64(8080),
		}

		err := ValidateSchema(data, spec)
		require.NoError(t, err)
	})

	t.Run("fails with missing required field", func(t *testing.T) {
		spec := map[string]FieldSpec{
			"url": {Name: "url", Type: FieldTypeString, Required: true},
		}

		data := JSONMap{}

		err := ValidateSchema(data, spec)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "required")
	})

	t.Run("accepts missing optional field", func(t *testing.T) {
		spec := map[string]FieldSpec{
			"url": {Name: "url", Type: FieldTypeString, Required: false},
		}

		data := JSONMap{}

		err := ValidateSchema(data, spec)
		require.NoError(t, err)
	})

	t.Run("validates string type", func(t *testing.T) {
		spec := map[string]FieldSpec{
			"name": {Name: "name", Type: FieldTypeString, Required: true},
		}

		// Valid
		err := ValidateSchema(JSONMap{"name": "test"}, spec)
		require.NoError(t, err)

		// Invalid
		err = ValidateSchema(JSONMap{"name": 123}, spec)
		require.Error(t, err)
	})

	t.Run("validates number type", func(t *testing.T) {
		spec := map[string]FieldSpec{
			"value": {Name: "value", Type: FieldTypeNumber, Required: true},
		}

		// Valid float64
		err := ValidateSchema(JSONMap{"value": float64(3.14)}, spec)
		require.NoError(t, err)

		// Valid float32
		err = ValidateSchema(JSONMap{"value": float32(3.14)}, spec)
		require.NoError(t, err)

		// Valid int
		err = ValidateSchema(JSONMap{"value": 42}, spec)
		require.NoError(t, err)

		// Valid int64
		err = ValidateSchema(JSONMap{"value": int64(42)}, spec)
		require.NoError(t, err)

		// Valid int32
		err = ValidateSchema(JSONMap{"value": int32(42)}, spec)
		require.NoError(t, err)

		// Invalid
		err = ValidateSchema(JSONMap{"value": "string"}, spec)
		require.Error(t, err)
	})

	t.Run("validates integer type", func(t *testing.T) {
		spec := map[string]FieldSpec{
			"count": {Name: "count", Type: FieldTypeInteger, Required: true},
		}

		// Valid int
		err := ValidateSchema(JSONMap{"count": 42}, spec)
		require.NoError(t, err)

		// Valid int64
		err = ValidateSchema(JSONMap{"count": int64(42)}, spec)
		require.NoError(t, err)

		// Valid int32
		err = ValidateSchema(JSONMap{"count": int32(42)}, spec)
		require.NoError(t, err)

		// Valid float64 that is whole number
		err = ValidateSchema(JSONMap{"count": float64(42)}, spec)
		require.NoError(t, err)

		// Invalid float64 with decimals
		err = ValidateSchema(JSONMap{"count": float64(42.5)}, spec)
		require.Error(t, err)

		// Invalid string
		err = ValidateSchema(JSONMap{"count": "42"}, spec)
		require.Error(t, err)
	})

	t.Run("validates boolean type", func(t *testing.T) {
		spec := map[string]FieldSpec{
			"enabled": {Name: "enabled", Type: FieldTypeBoolean, Required: true},
		}

		// Valid
		err := ValidateSchema(JSONMap{"enabled": true}, spec)
		require.NoError(t, err)

		err = ValidateSchema(JSONMap{"enabled": false}, spec)
		require.NoError(t, err)

		// Invalid
		err = ValidateSchema(JSONMap{"enabled": "true"}, spec)
		require.Error(t, err)
	})

	t.Run("validates array type", func(t *testing.T) {
		spec := map[string]FieldSpec{
			"items": {Name: "items", Type: FieldTypeArray, Required: true},
		}

		// Valid
		err := ValidateSchema(JSONMap{"items": []any{"a", "b", "c"}}, spec)
		require.NoError(t, err)

		// Invalid
		err = ValidateSchema(JSONMap{"items": "not an array"}, spec)
		require.Error(t, err)
	})

	t.Run("validates object type", func(t *testing.T) {
		spec := map[string]FieldSpec{
			"config": {Name: "config", Type: FieldTypeObject, Required: true},
		}

		// Valid map[string]any
		err := ValidateSchema(JSONMap{"config": map[string]any{"key": "value"}}, spec)
		require.NoError(t, err)

		// Valid JSONMap
		err = ValidateSchema(JSONMap{"config": JSONMap{"key": "value"}}, spec)
		require.NoError(t, err)

		// Invalid
		err = ValidateSchema(JSONMap{"config": "not an object"}, spec)
		require.Error(t, err)
	})

	t.Run("validates secret type as string", func(t *testing.T) {
		spec := map[string]FieldSpec{
			"api_key": {Name: "api_key", Type: FieldTypeSecret, Required: true},
		}

		// Valid
		err := ValidateSchema(JSONMap{"api_key": "secret-key-123"}, spec)
		require.NoError(t, err)

		// Invalid
		err = ValidateSchema(JSONMap{"api_key": 123}, spec)
		require.Error(t, err)
	})

	t.Run("validates required field cannot be null", func(t *testing.T) {
		spec := map[string]FieldSpec{
			"name": {Name: "name", Type: FieldTypeString, Required: true},
		}

		err := ValidateSchema(JSONMap{"name": nil}, spec)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "null")
	})

	t.Run("accepts null for optional field", func(t *testing.T) {
		spec := map[string]FieldSpec{
			"name": {Name: "name", Type: FieldTypeString, Required: false},
		}

		err := ValidateSchema(JSONMap{"name": nil}, spec)
		require.NoError(t, err)
	})
}

// mockIntegrationImpl is a mock implementation of Integration for testing
type mockIntegrationImpl struct {
	*BaseIntegration
	executeFunc  func(ctx context.Context, config *Config, params JSONMap) (*Result, error)
	validateFunc func(config *Config) error
}

func (m *mockIntegrationImpl) Execute(ctx context.Context, config *Config, params JSONMap) (*Result, error) {
	if m.executeFunc != nil {
		return m.executeFunc(ctx, config, params)
	}
	return NewSuccessResult(nil, 0), nil
}

func (m *mockIntegrationImpl) Validate(config *Config) error {
	if m.validateFunc != nil {
		return m.validateFunc(config)
	}
	return m.BaseIntegration.ValidateConfig(config)
}

func TestIntegrationInterface(t *testing.T) {
	t.Run("integration implements interface", func(t *testing.T) {
		base := NewBaseIntegration("test", TypeHTTP)
		integ := &mockIntegrationImpl{BaseIntegration: base}

		// Verify it implements the interface
		var _ Integration = integ

		assert.Equal(t, "test", integ.Name())
		assert.Equal(t, TypeHTTP, integ.Type())
		assert.NotNil(t, integ.GetSchema())
		assert.NotNil(t, integ.GetMetadata())
	})

	t.Run("custom execute function", func(t *testing.T) {
		base := NewBaseIntegration("test", TypeHTTP)
		integ := &mockIntegrationImpl{
			BaseIntegration: base,
			executeFunc: func(ctx context.Context, config *Config, params JSONMap) (*Result, error) {
				return NewSuccessResult(JSONMap{"custom": "result"}, 100), nil
			},
		}

		result, err := integ.Execute(context.Background(), nil, nil)
		require.NoError(t, err)
		assert.True(t, result.Success)

		data := result.Data.(JSONMap)
		assert.Equal(t, "result", data["custom"])
	})

	t.Run("custom validate function", func(t *testing.T) {
		base := NewBaseIntegration("test", TypeHTTP)
		integ := &mockIntegrationImpl{
			BaseIntegration: base,
			validateFunc: func(config *Config) error {
				requireAuth, ok := config.Settings.GetBool("require_auth")
				if ok && requireAuth {
					if _, hasKey := config.Settings.GetString("api_key"); !hasKey {
						return NewValidationError("api_key", "required when auth is enabled", nil)
					}
				}
				return nil
			},
		}

		// Valid without auth
		config := &Config{Name: "test", Type: TypeHTTP, Settings: JSONMap{}}
		err := integ.Validate(config)
		require.NoError(t, err)

		// Invalid with auth but no key
		config = &Config{
			Name: "test",
			Type: TypeHTTP,
			Settings: JSONMap{
				"require_auth": true,
			},
		}
		err = integ.Validate(config)
		require.Error(t, err)

		// Valid with auth and key
		config = &Config{
			Name: "test",
			Type: TypeHTTP,
			Settings: JSONMap{
				"require_auth": true,
				"api_key":      "key123",
			},
		}
		err = integ.Validate(config)
		require.NoError(t, err)
	})
}

// mockLifecycleIntegration tests LifecycleAware interface
type mockLifecycleIntegration struct {
	*BaseIntegration
	initCalled     bool
	shutdownCalled bool
}

func (m *mockLifecycleIntegration) Initialize(ctx context.Context) error {
	m.initCalled = true
	return nil
}

func (m *mockLifecycleIntegration) Shutdown(ctx context.Context) error {
	m.shutdownCalled = true
	return nil
}

func (m *mockLifecycleIntegration) Execute(ctx context.Context, config *Config, params JSONMap) (*Result, error) {
	return NewSuccessResult(nil, 0), nil
}

func (m *mockLifecycleIntegration) Validate(config *Config) error {
	return m.BaseIntegration.ValidateConfig(config)
}

func TestLifecycleAwareInterface(t *testing.T) {
	base := NewBaseIntegration("test", TypeHTTP)
	integ := &mockLifecycleIntegration{BaseIntegration: base}

	// Verify it implements both interfaces
	var _ Integration = integ
	var _ LifecycleAware = integ

	t.Run("initialize is called", func(t *testing.T) {
		err := integ.Initialize(context.Background())
		require.NoError(t, err)
		assert.True(t, integ.initCalled)
	})

	t.Run("shutdown is called", func(t *testing.T) {
		err := integ.Shutdown(context.Background())
		require.NoError(t, err)
		assert.True(t, integ.shutdownCalled)
	})
}

// mockHealthCheckIntegration tests HealthCheckable interface
type mockHealthCheckIntegration struct {
	*BaseIntegration
	healthy bool
}

func (m *mockHealthCheckIntegration) HealthCheck(ctx context.Context) error {
	if !m.healthy {
		return NewExecutionError("test", "health_check", ErrNotFound, false)
	}
	return nil
}

func (m *mockHealthCheckIntegration) Execute(ctx context.Context, config *Config, params JSONMap) (*Result, error) {
	return NewSuccessResult(nil, 0), nil
}

func (m *mockHealthCheckIntegration) Validate(config *Config) error {
	return m.BaseIntegration.ValidateConfig(config)
}

func TestHealthCheckableInterface(t *testing.T) {
	base := NewBaseIntegration("test", TypeHTTP)
	integ := &mockHealthCheckIntegration{BaseIntegration: base, healthy: true}

	// Verify it implements both interfaces
	var _ Integration = integ
	var _ HealthCheckable = integ

	t.Run("healthy check passes", func(t *testing.T) {
		integ.healthy = true
		err := integ.HealthCheck(context.Background())
		require.NoError(t, err)
	})

	t.Run("unhealthy check fails", func(t *testing.T) {
		integ.healthy = false
		err := integ.HealthCheck(context.Background())
		require.Error(t, err)
	})
}

// mockRefreshableIntegration tests Refreshable interface
type mockRefreshableIntegration struct {
	*BaseIntegration
}

func (m *mockRefreshableIntegration) RefreshCredentials(ctx context.Context, creds *Credentials) (*Credentials, error) {
	newCreds := &Credentials{
		ID:   creds.ID,
		Type: creds.Type,
		Data: JSONMap{
			"access_token": "new-token",
		},
	}
	return newCreds, nil
}

func (m *mockRefreshableIntegration) Execute(ctx context.Context, config *Config, params JSONMap) (*Result, error) {
	return NewSuccessResult(nil, 0), nil
}

func (m *mockRefreshableIntegration) Validate(config *Config) error {
	return m.BaseIntegration.ValidateConfig(config)
}

func TestRefreshableInterface(t *testing.T) {
	base := NewBaseIntegration("test", TypeHTTP)
	integ := &mockRefreshableIntegration{BaseIntegration: base}

	// Verify it implements both interfaces
	var _ Integration = integ
	var _ Refreshable = integ

	t.Run("refresh credentials", func(t *testing.T) {
		oldCreds := &Credentials{
			ID:   "cred-123",
			Type: CredTypeOAuth2,
			Data: JSONMap{
				"access_token": "old-token",
			},
		}

		newCreds, err := integ.RefreshCredentials(context.Background(), oldCreds)
		require.NoError(t, err)
		assert.Equal(t, "cred-123", newCreds.ID)
		assert.Equal(t, "new-token", newCreds.Data["access_token"])
	})
}

// mockConfigurableIntegration tests Configurable interface
type mockConfigurableIntegration struct {
	*BaseIntegration
	currentConfig *Config
}

func (m *mockConfigurableIntegration) UpdateConfig(ctx context.Context, config *Config) error {
	m.currentConfig = config
	return nil
}

func (m *mockConfigurableIntegration) Execute(ctx context.Context, config *Config, params JSONMap) (*Result, error) {
	return NewSuccessResult(nil, 0), nil
}

func (m *mockConfigurableIntegration) Validate(config *Config) error {
	return m.BaseIntegration.ValidateConfig(config)
}

func TestConfigurableInterface(t *testing.T) {
	base := NewBaseIntegration("test", TypeHTTP)
	integ := &mockConfigurableIntegration{BaseIntegration: base}

	// Verify it implements both interfaces
	var _ Integration = integ
	var _ Configurable = integ

	t.Run("update config", func(t *testing.T) {
		config := &Config{
			Name: "updated-config",
			Type: TypeHTTP,
			Settings: JSONMap{
				"url": "https://new-url.com",
			},
		}

		err := integ.UpdateConfig(context.Background(), config)
		require.NoError(t, err)
		assert.Equal(t, config, integ.currentConfig)
	})
}
