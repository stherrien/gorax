package integrations

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gorax/gorax/internal/integration"
)

func TestNewPagerDutyIntegration(t *testing.T) {
	pd := NewPagerDutyIntegration(nil)

	assert.NotNil(t, pd)
	assert.Equal(t, "pagerduty", pd.Name())
	assert.Equal(t, integration.TypeAPI, pd.Type())

	metadata := pd.GetMetadata()
	assert.Equal(t, "PagerDuty", metadata.DisplayName)
	assert.Equal(t, "incident_management", metadata.Category)

	schema := pd.GetSchema()
	assert.NotNil(t, schema.ConfigSpec["api_token"])
	assert.NotNil(t, schema.ConfigSpec["routing_key"])
	assert.NotNil(t, schema.InputSpec["action"])
}

func TestPagerDutyIntegration_Validate(t *testing.T) {
	pd := NewPagerDutyIntegration(nil)

	tests := []struct {
		name        string
		config      *integration.Config
		expectError bool
	}{
		{
			name:        "nil config",
			config:      nil,
			expectError: true,
		},
		{
			name: "missing credentials",
			config: &integration.Config{
				Name:    "test",
				Type:    integration.TypeAPI,
				Enabled: true,
			},
			expectError: true,
		},
		{
			name: "missing both api_token and routing_key",
			config: &integration.Config{
				Name:    "test",
				Type:    integration.TypeAPI,
				Enabled: true,
				Credentials: &integration.Credentials{
					Data: integration.JSONMap{},
				},
			},
			expectError: true,
		},
		{
			name: "valid config with api_token",
			config: &integration.Config{
				Name:    "test",
				Type:    integration.TypeAPI,
				Enabled: true,
				Credentials: &integration.Credentials{
					Data: integration.JSONMap{
						"api_token": "test-api-token",
					},
				},
			},
			expectError: false,
		},
		{
			name: "valid config with routing_key",
			config: &integration.Config{
				Name:    "test",
				Type:    integration.TypeAPI,
				Enabled: true,
				Credentials: &integration.Credentials{
					Data: integration.JSONMap{
						"routing_key": "test-routing-key",
					},
				},
			},
			expectError: false,
		},
		{
			name: "valid config with both",
			config: &integration.Config{
				Name:    "test",
				Type:    integration.TypeAPI,
				Enabled: true,
				Credentials: &integration.Credentials{
					Data: integration.JSONMap{
						"api_token":   "test-api-token",
						"routing_key": "test-routing-key",
					},
				},
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := pd.Validate(tt.config)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestPagerDutyIntegration_Execute_MissingAction(t *testing.T) {
	pd := NewPagerDutyIntegration(nil)

	config := &integration.Config{
		Name:    "test-pd",
		Type:    integration.TypeAPI,
		Enabled: true,
		Credentials: &integration.Credentials{
			Data: integration.JSONMap{
				"api_token": "test-token",
			},
		},
	}

	params := integration.JSONMap{
		"incident_id": "P123456",
	}

	ctx := context.Background()
	result, err := pd.Execute(ctx, config, params)

	require.Error(t, err)
	assert.False(t, result.Success)
	assert.Equal(t, "VALIDATION_ERROR", result.ErrorCode)
}

func TestPagerDutyIntegration_Execute_InvalidAction(t *testing.T) {
	pd := NewPagerDutyIntegration(nil)

	config := &integration.Config{
		Name:    "test-pd",
		Type:    integration.TypeAPI,
		Enabled: true,
		Credentials: &integration.Credentials{
			Data: integration.JSONMap{
				"api_token": "test-token",
			},
		},
	}

	params := integration.JSONMap{
		"action": "invalid_action",
	}

	ctx := context.Background()
	result, err := pd.Execute(ctx, config, params)

	require.Error(t, err)
	assert.False(t, result.Success)
	assert.Equal(t, "INVALID_ACTION", result.ErrorCode)
}

func TestPagerDutyIntegration_TriggerIncident_MissingSummary(t *testing.T) {
	pd := NewPagerDutyIntegration(nil)

	config := &integration.Config{
		Name:    "test-pd",
		Type:    integration.TypeAPI,
		Enabled: true,
		Credentials: &integration.Credentials{
			Data: integration.JSONMap{
				"routing_key": "test-routing-key",
			},
		},
	}

	params := integration.JSONMap{
		"action": "trigger_incident",
	}

	ctx := context.Background()
	result, err := pd.Execute(ctx, config, params)

	require.Error(t, err)
	assert.False(t, result.Success)
	assert.Equal(t, "VALIDATION_ERROR", result.ErrorCode)
}

func TestPagerDutyIntegration_TriggerIncident_MissingRoutingKey(t *testing.T) {
	pd := NewPagerDutyIntegration(nil)

	config := &integration.Config{
		Name:    "test-pd",
		Type:    integration.TypeAPI,
		Enabled: true,
		Credentials: &integration.Credentials{
			Data: integration.JSONMap{
				"api_token": "test-token", // No routing key
			},
		},
	}

	params := integration.JSONMap{
		"action":  "trigger_incident",
		"summary": "Test incident",
	}

	ctx := context.Background()
	result, err := pd.Execute(ctx, config, params)

	require.Error(t, err)
	assert.False(t, result.Success)
	assert.Equal(t, "AUTH_ERROR", result.ErrorCode)
}

func TestPagerDutyIntegration_AcknowledgeIncident_MissingDedupKey(t *testing.T) {
	pd := NewPagerDutyIntegration(nil)

	config := &integration.Config{
		Name:    "test-pd",
		Type:    integration.TypeAPI,
		Enabled: true,
		Credentials: &integration.Credentials{
			Data: integration.JSONMap{
				"routing_key": "test-routing-key",
			},
		},
	}

	params := integration.JSONMap{
		"action": "acknowledge_incident",
	}

	ctx := context.Background()
	result, err := pd.Execute(ctx, config, params)

	require.Error(t, err)
	assert.False(t, result.Success)
	assert.Equal(t, "VALIDATION_ERROR", result.ErrorCode)
}

func TestPagerDutyIntegration_ResolveIncident_MissingDedupKey(t *testing.T) {
	pd := NewPagerDutyIntegration(nil)

	config := &integration.Config{
		Name:    "test-pd",
		Type:    integration.TypeAPI,
		Enabled: true,
		Credentials: &integration.Credentials{
			Data: integration.JSONMap{
				"routing_key": "test-routing-key",
			},
		},
	}

	params := integration.JSONMap{
		"action": "resolve_incident",
	}

	ctx := context.Background()
	result, err := pd.Execute(ctx, config, params)

	require.Error(t, err)
	assert.False(t, result.Success)
	assert.Equal(t, "VALIDATION_ERROR", result.ErrorCode)
}

func TestPagerDutyIntegration_CreateIncident_MissingServiceID(t *testing.T) {
	pd := NewPagerDutyIntegration(nil)

	config := &integration.Config{
		Name:    "test-pd",
		Type:    integration.TypeAPI,
		Enabled: true,
		Credentials: &integration.Credentials{
			Data: integration.JSONMap{
				"api_token": "test-token",
			},
		},
	}

	params := integration.JSONMap{
		"action": "create_incident",
		"title":  "Test incident",
	}

	ctx := context.Background()
	result, err := pd.Execute(ctx, config, params)

	require.Error(t, err)
	assert.False(t, result.Success)
	assert.Equal(t, "VALIDATION_ERROR", result.ErrorCode)
}

func TestPagerDutyIntegration_CreateIncident_MissingTitle(t *testing.T) {
	pd := NewPagerDutyIntegration(nil)

	config := &integration.Config{
		Name:    "test-pd",
		Type:    integration.TypeAPI,
		Enabled: true,
		Credentials: &integration.Credentials{
			Data: integration.JSONMap{
				"api_token": "test-token",
			},
		},
	}

	params := integration.JSONMap{
		"action":     "create_incident",
		"service_id": "P123456",
	}

	ctx := context.Background()
	result, err := pd.Execute(ctx, config, params)

	require.Error(t, err)
	assert.False(t, result.Success)
	assert.Equal(t, "VALIDATION_ERROR", result.ErrorCode)
}

func TestPagerDutyIntegration_GetIncident_MissingID(t *testing.T) {
	pd := NewPagerDutyIntegration(nil)

	config := &integration.Config{
		Name:    "test-pd",
		Type:    integration.TypeAPI,
		Enabled: true,
		Credentials: &integration.Credentials{
			Data: integration.JSONMap{
				"api_token": "test-token",
			},
		},
	}

	params := integration.JSONMap{
		"action": "get_incident",
	}

	ctx := context.Background()
	result, err := pd.Execute(ctx, config, params)

	require.Error(t, err)
	assert.False(t, result.Success)
	assert.Equal(t, "VALIDATION_ERROR", result.ErrorCode)
}

func TestPagerDutyIntegration_AddNote_MissingContent(t *testing.T) {
	pd := NewPagerDutyIntegration(nil)

	config := &integration.Config{
		Name:    "test-pd",
		Type:    integration.TypeAPI,
		Enabled: true,
		Credentials: &integration.Credentials{
			Data: integration.JSONMap{
				"api_token": "test-token",
			},
		},
	}

	params := integration.JSONMap{
		"action":      "add_note",
		"incident_id": "P123456",
	}

	ctx := context.Background()
	result, err := pd.Execute(ctx, config, params)

	require.Error(t, err)
	assert.False(t, result.Success)
	assert.Equal(t, "VALIDATION_ERROR", result.ErrorCode)
}

func TestPagerDutyIntegration_GetService_MissingID(t *testing.T) {
	pd := NewPagerDutyIntegration(nil)

	config := &integration.Config{
		Name:    "test-pd",
		Type:    integration.TypeAPI,
		Enabled: true,
		Credentials: &integration.Credentials{
			Data: integration.JSONMap{
				"api_token": "test-token",
			},
		},
	}

	params := integration.JSONMap{
		"action": "get_service",
	}

	ctx := context.Background()
	result, err := pd.Execute(ctx, config, params)

	require.Error(t, err)
	assert.False(t, result.Success)
	assert.Equal(t, "VALIDATION_ERROR", result.ErrorCode)
}

func TestPagerDutyRetryConfig(t *testing.T) {
	config := buildPagerDutyRetryConfig()

	assert.Equal(t, 3, config.MaxRetries)
	assert.Equal(t, 1*time.Second, config.BaseDelay)
	assert.Equal(t, 30*time.Second, config.MaxDelay)
	assert.NotNil(t, config.ShouldRetry)
}

func TestPagerDutySchema(t *testing.T) {
	schema := buildPagerDutySchema()

	// Verify config spec
	assert.Contains(t, schema.ConfigSpec, "api_token")
	assert.Contains(t, schema.ConfigSpec, "routing_key")
	assert.True(t, schema.ConfigSpec["api_token"].Sensitive)
	assert.True(t, schema.ConfigSpec["routing_key"].Sensitive)

	// Verify input spec
	assert.Contains(t, schema.InputSpec, "action")
	assert.True(t, schema.InputSpec["action"].Required)
	assert.Contains(t, schema.InputSpec, "incident_id")
	assert.Contains(t, schema.InputSpec, "service_id")
	assert.Contains(t, schema.InputSpec, "title")
	assert.Contains(t, schema.InputSpec, "summary")
	assert.Contains(t, schema.InputSpec, "severity")
	assert.Contains(t, schema.InputSpec, "dedup_key")

	// Verify severity options
	severitySpec := schema.InputSpec["severity"]
	assert.Contains(t, severitySpec.Options, "critical")
	assert.Contains(t, severitySpec.Options, "error")
	assert.Contains(t, severitySpec.Options, "warning")
	assert.Contains(t, severitySpec.Options, "info")

	// Verify output spec
	assert.Contains(t, schema.OutputSpec, "incident")
	assert.Contains(t, schema.OutputSpec, "incidents")
	assert.Contains(t, schema.OutputSpec, "service")
	assert.Contains(t, schema.OutputSpec, "dedup_key")
}

func TestPagerDutyIntegration_GetRoutingKey(t *testing.T) {
	pd := NewPagerDutyIntegration(nil)

	tests := []struct {
		name        string
		config      *integration.Config
		params      integration.JSONMap
		wantKey     string
		expectError bool
	}{
		{
			name: "routing key from params",
			config: &integration.Config{
				Credentials: &integration.Credentials{
					Data: integration.JSONMap{
						"routing_key": "cred-key",
					},
				},
			},
			params: integration.JSONMap{
				"routing_key": "param-key",
			},
			wantKey:     "param-key",
			expectError: false,
		},
		{
			name: "routing key from credentials",
			config: &integration.Config{
				Credentials: &integration.Credentials{
					Data: integration.JSONMap{
						"routing_key": "cred-key",
					},
				},
			},
			params:      integration.JSONMap{},
			wantKey:     "cred-key",
			expectError: false,
		},
		{
			name: "missing routing key",
			config: &integration.Config{
				Credentials: &integration.Credentials{
					Data: integration.JSONMap{},
				},
			},
			params:      integration.JSONMap{},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key, err := pd.getRoutingKey(tt.config, tt.params)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantKey, key)
			}
		})
	}
}
