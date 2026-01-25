package integrations

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gorax/gorax/internal/integration"
)

func TestNewJiraIntegration(t *testing.T) {
	jira := NewJiraIntegration(nil)

	assert.NotNil(t, jira)
	assert.Equal(t, "jira", jira.Name())
	assert.Equal(t, integration.TypeAPI, jira.Type())

	metadata := jira.GetMetadata()
	assert.Equal(t, "Jira", metadata.DisplayName)
	assert.Equal(t, "project_management", metadata.Category)

	schema := jira.GetSchema()
	assert.NotNil(t, schema.ConfigSpec["base_url"])
	assert.NotNil(t, schema.InputSpec["action"])
}

func TestJiraIntegration_Validate(t *testing.T) {
	jira := NewJiraIntegration(nil)

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
			name: "missing base_url",
			config: &integration.Config{
				Name:     "test",
				Type:     integration.TypeAPI,
				Enabled:  true,
				Settings: integration.JSONMap{},
			},
			expectError: true,
		},
		{
			name: "missing credentials",
			config: &integration.Config{
				Name:    "test",
				Type:    integration.TypeAPI,
				Enabled: true,
				Settings: integration.JSONMap{
					"base_url": "https://example.atlassian.net",
				},
			},
			expectError: true,
		},
		{
			name: "valid config with email/api_token",
			config: &integration.Config{
				Name:    "test",
				Type:    integration.TypeAPI,
				Enabled: true,
				Settings: integration.JSONMap{
					"base_url": "https://example.atlassian.net",
				},
				Credentials: &integration.Credentials{
					Data: integration.JSONMap{
						"email":     "user@example.com",
						"api_token": "test-api-token",
					},
				},
			},
			expectError: false,
		},
		{
			name: "valid config with username/password",
			config: &integration.Config{
				Name:    "test",
				Type:    integration.TypeAPI,
				Enabled: true,
				Settings: integration.JSONMap{
					"base_url": "https://jira.example.com",
				},
				Credentials: &integration.Credentials{
					Data: integration.JSONMap{
						"username": "admin",
						"password": "secret",
					},
				},
			},
			expectError: false,
		},
		{
			name: "valid config with bearer token",
			config: &integration.Config{
				Name:    "test",
				Type:    integration.TypeAPI,
				Enabled: true,
				Settings: integration.JSONMap{
					"base_url": "https://example.atlassian.net",
				},
				Credentials: &integration.Credentials{
					Data: integration.JSONMap{
						"token": "bearer-token",
					},
				},
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := jira.Validate(tt.config)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestJiraIntegration_Execute_MissingAction(t *testing.T) {
	jira := NewJiraIntegration(nil)

	config := &integration.Config{
		Name:    "test-jira",
		Type:    integration.TypeAPI,
		Enabled: true,
		Settings: integration.JSONMap{
			"base_url": "https://example.atlassian.net",
		},
		Credentials: &integration.Credentials{
			Data: integration.JSONMap{
				"email":     "user@example.com",
				"api_token": "test-token",
			},
		},
	}

	params := integration.JSONMap{
		"project": "TEST",
	}

	ctx := context.Background()
	result, err := jira.Execute(ctx, config, params)

	require.Error(t, err)
	assert.False(t, result.Success)
	assert.Equal(t, "VALIDATION_ERROR", result.ErrorCode)
}

func TestJiraIntegration_Execute_InvalidAction(t *testing.T) {
	jira := NewJiraIntegration(nil)

	config := &integration.Config{
		Name:    "test-jira",
		Type:    integration.TypeAPI,
		Enabled: true,
		Settings: integration.JSONMap{
			"base_url": "https://example.atlassian.net",
		},
		Credentials: &integration.Credentials{
			Data: integration.JSONMap{
				"email":     "user@example.com",
				"api_token": "test-token",
			},
		},
	}

	params := integration.JSONMap{
		"action": "invalid_action",
	}

	ctx := context.Background()
	result, err := jira.Execute(ctx, config, params)

	require.Error(t, err)
	assert.False(t, result.Success)
	assert.Equal(t, "INVALID_ACTION", result.ErrorCode)
}

func TestJiraIntegration_CreateIssue_MissingProject(t *testing.T) {
	jira := NewJiraIntegration(nil)

	config := &integration.Config{
		Name:    "test-jira",
		Type:    integration.TypeAPI,
		Enabled: true,
		Settings: integration.JSONMap{
			"base_url": "https://example.atlassian.net",
		},
		Credentials: &integration.Credentials{
			Data: integration.JSONMap{
				"email":     "user@example.com",
				"api_token": "test-token",
			},
		},
	}

	params := integration.JSONMap{
		"action":  "create_issue",
		"summary": "Test Issue",
	}

	ctx := context.Background()
	result, err := jira.Execute(ctx, config, params)

	require.Error(t, err)
	assert.False(t, result.Success)
	assert.Equal(t, "VALIDATION_ERROR", result.ErrorCode)
}

func TestJiraIntegration_CreateIssue_MissingSummary(t *testing.T) {
	jira := NewJiraIntegration(nil)

	config := &integration.Config{
		Name:    "test-jira",
		Type:    integration.TypeAPI,
		Enabled: true,
		Settings: integration.JSONMap{
			"base_url": "https://example.atlassian.net",
		},
		Credentials: &integration.Credentials{
			Data: integration.JSONMap{
				"email":     "user@example.com",
				"api_token": "test-token",
			},
		},
	}

	params := integration.JSONMap{
		"action":  "create_issue",
		"project": "TEST",
	}

	ctx := context.Background()
	result, err := jira.Execute(ctx, config, params)

	require.Error(t, err)
	assert.False(t, result.Success)
	assert.Equal(t, "VALIDATION_ERROR", result.ErrorCode)
}

func TestJiraIntegration_GetIssue_MissingKey(t *testing.T) {
	jira := NewJiraIntegration(nil)

	config := &integration.Config{
		Name:    "test-jira",
		Type:    integration.TypeAPI,
		Enabled: true,
		Settings: integration.JSONMap{
			"base_url": "https://example.atlassian.net",
		},
		Credentials: &integration.Credentials{
			Data: integration.JSONMap{
				"email":     "user@example.com",
				"api_token": "test-token",
			},
		},
	}

	params := integration.JSONMap{
		"action": "get_issue",
	}

	ctx := context.Background()
	result, err := jira.Execute(ctx, config, params)

	require.Error(t, err)
	assert.False(t, result.Success)
	assert.Equal(t, "VALIDATION_ERROR", result.ErrorCode)
}

func TestJiraIntegration_SearchIssues_MissingJQL(t *testing.T) {
	jira := NewJiraIntegration(nil)

	config := &integration.Config{
		Name:    "test-jira",
		Type:    integration.TypeAPI,
		Enabled: true,
		Settings: integration.JSONMap{
			"base_url": "https://example.atlassian.net",
		},
		Credentials: &integration.Credentials{
			Data: integration.JSONMap{
				"email":     "user@example.com",
				"api_token": "test-token",
			},
		},
	}

	params := integration.JSONMap{
		"action": "search_issues",
	}

	ctx := context.Background()
	result, err := jira.Execute(ctx, config, params)

	require.Error(t, err)
	assert.False(t, result.Success)
	assert.Equal(t, "VALIDATION_ERROR", result.ErrorCode)
}

func TestJiraIntegration_TransitionIssue_MissingTransition(t *testing.T) {
	jira := NewJiraIntegration(nil)

	config := &integration.Config{
		Name:    "test-jira",
		Type:    integration.TypeAPI,
		Enabled: true,
		Settings: integration.JSONMap{
			"base_url": "https://example.atlassian.net",
		},
		Credentials: &integration.Credentials{
			Data: integration.JSONMap{
				"email":     "user@example.com",
				"api_token": "test-token",
			},
		},
	}

	params := integration.JSONMap{
		"action":    "transition_issue",
		"issue_key": "TEST-123",
	}

	ctx := context.Background()
	result, err := jira.Execute(ctx, config, params)

	require.Error(t, err)
	assert.False(t, result.Success)
	assert.Equal(t, "VALIDATION_ERROR", result.ErrorCode)
}

func TestJiraIntegration_AddComment_MissingBody(t *testing.T) {
	jira := NewJiraIntegration(nil)

	config := &integration.Config{
		Name:    "test-jira",
		Type:    integration.TypeAPI,
		Enabled: true,
		Settings: integration.JSONMap{
			"base_url": "https://example.atlassian.net",
		},
		Credentials: &integration.Credentials{
			Data: integration.JSONMap{
				"email":     "user@example.com",
				"api_token": "test-token",
			},
		},
	}

	params := integration.JSONMap{
		"action":    "add_comment",
		"issue_key": "TEST-123",
	}

	ctx := context.Background()
	result, err := jira.Execute(ctx, config, params)

	require.Error(t, err)
	assert.False(t, result.Success)
	assert.Equal(t, "VALIDATION_ERROR", result.ErrorCode)
}

func TestJiraRetryConfig(t *testing.T) {
	config := buildJiraRetryConfig()

	assert.Equal(t, 3, config.MaxRetries)
	assert.Equal(t, 1*time.Second, config.BaseDelay)
	assert.Equal(t, 30*time.Second, config.MaxDelay)
	assert.NotNil(t, config.ShouldRetry)
}

func TestJiraSchema(t *testing.T) {
	schema := buildJiraSchema()

	// Verify config spec
	assert.Contains(t, schema.ConfigSpec, "base_url")
	assert.Contains(t, schema.ConfigSpec, "email")
	assert.Contains(t, schema.ConfigSpec, "api_token")
	assert.True(t, schema.ConfigSpec["api_token"].Sensitive)

	// Verify input spec
	assert.Contains(t, schema.InputSpec, "action")
	assert.True(t, schema.InputSpec["action"].Required)
	assert.Contains(t, schema.InputSpec, "project")
	assert.Contains(t, schema.InputSpec, "issue_key")
	assert.Contains(t, schema.InputSpec, "summary")
	assert.Contains(t, schema.InputSpec, "description")
	assert.Contains(t, schema.InputSpec, "jql")
	assert.Contains(t, schema.InputSpec, "transition_id")

	// Verify output spec
	assert.Contains(t, schema.OutputSpec, "id")
	assert.Contains(t, schema.OutputSpec, "key")
	assert.Contains(t, schema.OutputSpec, "issues")
}

func TestJiraIntegration_GetAuthHeader(t *testing.T) {
	jira := NewJiraIntegration(nil)

	tests := []struct {
		name        string
		config      *integration.Config
		wantPrefix  string
		expectError bool
	}{
		{
			name: "basic auth with email/api_token",
			config: &integration.Config{
				Credentials: &integration.Credentials{
					Data: integration.JSONMap{
						"email":     "user@example.com",
						"api_token": "token123",
					},
				},
			},
			wantPrefix:  "Basic ",
			expectError: false,
		},
		{
			name: "basic auth with username/password",
			config: &integration.Config{
				Credentials: &integration.Credentials{
					Data: integration.JSONMap{
						"username": "admin",
						"password": "secret",
					},
				},
			},
			wantPrefix:  "Basic ",
			expectError: false,
		},
		{
			name: "bearer token",
			config: &integration.Config{
				Credentials: &integration.Credentials{
					Data: integration.JSONMap{
						"token": "bearer-token-123",
					},
				},
			},
			wantPrefix:  "Bearer ",
			expectError: false,
		},
		{
			name: "missing credentials",
			config: &integration.Config{
				Credentials: nil,
			},
			expectError: true,
		},
		{
			name: "empty credentials data",
			config: &integration.Config{
				Credentials: &integration.Credentials{
					Data: integration.JSONMap{},
				},
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			header, err := jira.getAuthHeader(tt.config)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.True(t, len(header) > len(tt.wantPrefix))
				assert.Equal(t, tt.wantPrefix, header[:len(tt.wantPrefix)])
			}
		})
	}
}
