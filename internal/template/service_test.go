package template

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"log/slog"
	"os"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestService_CreateTemplate(t *testing.T) {
	repo := &MockRepository{}
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	service := NewService(repo, logger)

	ctx := context.Background()
	tenantID := "test-tenant-123"
	userID := "user-123"

	input := CreateTemplateInput{
		Name:        "Test Template",
		Description: "Test description",
		Category:    "security",
		Definition:  json.RawMessage(`{"nodes":[],"edges":[]}`),
		Tags:        []string{"test", "security"},
		IsPublic:    false,
	}

	template, err := service.CreateTemplate(ctx, tenantID, userID, input)
	require.NoError(t, err)
	assert.NotEmpty(t, template.ID)
	assert.Equal(t, input.Name, template.Name)
	assert.Equal(t, input.Category, template.Category)
}

func TestService_CreateTemplate_InvalidInput(t *testing.T) {
	repo := &MockRepository{}
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	service := NewService(repo, logger)

	ctx := context.Background()
	tenantID := "test-tenant-123"
	userID := "user-123"

	tests := []struct {
		name  string
		input CreateTemplateInput
	}{
		{
			name: "empty name",
			input: CreateTemplateInput{
				Name:       "",
				Category:   "security",
				Definition: json.RawMessage(`{"nodes":[],"edges":[]}`),
			},
		},
		{
			name: "empty category",
			input: CreateTemplateInput{
				Name:       "Test",
				Category:   "",
				Definition: json.RawMessage(`{"nodes":[],"edges":[]}`),
			},
		},
		{
			name: "empty definition",
			input: CreateTemplateInput{
				Name:       "Test",
				Category:   "security",
				Definition: json.RawMessage(``),
			},
		},
		{
			name: "invalid JSON definition",
			input: CreateTemplateInput{
				Name:       "Test",
				Category:   "security",
				Definition: json.RawMessage(`{invalid`),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := service.CreateTemplate(ctx, tenantID, userID, tt.input)
			assert.Error(t, err)
		})
	}
}

func TestService_GetTemplate(t *testing.T) {
	repo := &MockRepository{}
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	service := NewService(repo, logger)

	ctx := context.Background()
	tenantID := "test-tenant-123"
	userID := "user-123"

	input := CreateTemplateInput{
		Name:       "Test Template",
		Category:   "security",
		Definition: json.RawMessage(`{"nodes":[],"edges":[]}`),
	}

	created, err := service.CreateTemplate(ctx, tenantID, userID, input)
	require.NoError(t, err)

	found, err := service.GetTemplate(ctx, tenantID, created.ID)
	require.NoError(t, err)
	assert.Equal(t, created.ID, found.ID)
	assert.Equal(t, created.Name, found.Name)
}

func TestService_GetTemplate_NotFound(t *testing.T) {
	repo := &MockRepository{}
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	service := NewService(repo, logger)

	ctx := context.Background()
	tenantID := "test-tenant-123"

	_, err := service.GetTemplate(ctx, tenantID, "nonexistent-id")
	assert.Error(t, err)
}

func TestService_ListTemplates(t *testing.T) {
	repo := &MockRepository{}
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	service := NewService(repo, logger)

	ctx := context.Background()
	tenantID := "test-tenant-123"
	userID := "user-123"

	templates := []CreateTemplateInput{
		{
			Name:       "Security Template",
			Category:   "security",
			Definition: json.RawMessage(`{"nodes":[],"edges":[]}`),
			Tags:       []string{"security"},
		},
		{
			Name:       "Monitoring Template",
			Category:   "monitoring",
			Definition: json.RawMessage(`{"nodes":[],"edges":[]}`),
			Tags:       []string{"monitoring"},
		},
	}

	for _, tmpl := range templates {
		_, err := service.CreateTemplate(ctx, tenantID, userID, tmpl)
		require.NoError(t, err)
	}

	result, err := service.ListTemplates(ctx, tenantID, TemplateFilter{})
	require.NoError(t, err)
	assert.Len(t, result, 2)
}

func TestService_ListTemplates_WithFilter(t *testing.T) {
	repo := &MockRepository{}
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	service := NewService(repo, logger)

	ctx := context.Background()
	tenantID := "test-tenant-123"
	userID := "user-123"

	templates := []CreateTemplateInput{
		{
			Name:       "Security Template",
			Category:   "security",
			Definition: json.RawMessage(`{"nodes":[],"edges":[]}`),
		},
		{
			Name:       "Monitoring Template",
			Category:   "monitoring",
			Definition: json.RawMessage(`{"nodes":[],"edges":[]}`),
		},
	}

	for _, tmpl := range templates {
		_, err := service.CreateTemplate(ctx, tenantID, userID, tmpl)
		require.NoError(t, err)
	}

	filter := TemplateFilter{Category: "security"}
	result, err := service.ListTemplates(ctx, tenantID, filter)
	require.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, "security", result[0].Category)
}

func TestService_UpdateTemplate(t *testing.T) {
	repo := &MockRepository{}
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	service := NewService(repo, logger)

	ctx := context.Background()
	tenantID := "test-tenant-123"
	userID := "user-123"

	input := CreateTemplateInput{
		Name:       "Original Name",
		Category:   "security",
		Definition: json.RawMessage(`{"nodes":[],"edges":[]}`),
	}

	created, err := service.CreateTemplate(ctx, tenantID, userID, input)
	require.NoError(t, err)

	update := UpdateTemplateInput{
		Name:        "Updated Name",
		Description: "Updated description",
	}

	err = service.UpdateTemplate(ctx, tenantID, created.ID, update)
	require.NoError(t, err)

	updated, err := service.GetTemplate(ctx, tenantID, created.ID)
	require.NoError(t, err)
	assert.Equal(t, "Updated Name", updated.Name)
	assert.Equal(t, "Updated description", updated.Description)
}

func TestService_DeleteTemplate(t *testing.T) {
	repo := &MockRepository{}
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	service := NewService(repo, logger)

	ctx := context.Background()
	tenantID := "test-tenant-123"
	userID := "user-123"

	input := CreateTemplateInput{
		Name:       "To Delete",
		Category:   "security",
		Definition: json.RawMessage(`{"nodes":[],"edges":[]}`),
	}

	created, err := service.CreateTemplate(ctx, tenantID, userID, input)
	require.NoError(t, err)

	err = service.DeleteTemplate(ctx, tenantID, created.ID)
	require.NoError(t, err)

	_, err = service.GetTemplate(ctx, tenantID, created.ID)
	assert.Error(t, err)
}

func TestService_CreateFromWorkflow(t *testing.T) {
	repo := &MockRepository{}
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	service := NewService(repo, logger)

	ctx := context.Background()
	tenantID := "test-tenant-123"
	userID := "user-123"

	workflowDef := json.RawMessage(`{
		"nodes": [{"id": "1", "type": "trigger:webhook"}],
		"edges": []
	}`)

	input := CreateTemplateFromWorkflowInput{
		WorkflowID:  "workflow-123",
		Name:        "Workflow Template",
		Description: "From workflow",
		Category:    "integration",
		Definition:  workflowDef,
		Tags:        []string{"workflow"},
	}

	template, err := service.CreateFromWorkflow(ctx, tenantID, userID, input)
	require.NoError(t, err)
	assert.NotEmpty(t, template.ID)
	assert.Equal(t, input.Name, template.Name)
}

func TestService_InstantiateTemplate(t *testing.T) {
	repo := &MockRepository{}
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	service := NewService(repo, logger)

	ctx := context.Background()
	tenantID := "test-tenant-123"
	userID := "user-123"

	input := CreateTemplateInput{
		Name:     "Test Template",
		Category: "security",
		Definition: json.RawMessage(`{
			"nodes": [{"id": "1", "type": "trigger:webhook", "data": {"name": "Webhook"}}],
			"edges": []
		}`),
	}

	created, err := service.CreateTemplate(ctx, tenantID, userID, input)
	require.NoError(t, err)

	instantiateInput := InstantiateTemplateInput{
		WorkflowName: "New Workflow",
	}

	result, err := service.InstantiateTemplate(ctx, tenantID, created.ID, instantiateInput)
	require.NoError(t, err)
	assert.Equal(t, "New Workflow", result.WorkflowName)
	assert.NotEmpty(t, result.Definition)
}

func TestService_InstantiateTemplate_NotFound(t *testing.T) {
	repo := &MockRepository{}
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	service := NewService(repo, logger)

	ctx := context.Background()
	tenantID := "test-tenant-123"

	instantiateInput := InstantiateTemplateInput{
		WorkflowName: "New Workflow",
	}

	_, err := service.InstantiateTemplate(ctx, tenantID, "nonexistent-id", instantiateInput)
	assert.Error(t, err)
}

func TestService_ValidateDefinition(t *testing.T) {
	repo := &MockRepository{}
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	service := NewService(repo, logger)

	tests := []struct {
		name       string
		definition json.RawMessage
		wantError  bool
	}{
		{
			name:       "valid definition",
			definition: json.RawMessage(`{"nodes":[],"edges":[]}`),
			wantError:  false,
		},
		{
			name:       "empty definition",
			definition: json.RawMessage(``),
			wantError:  true,
		},
		{
			name:       "invalid JSON",
			definition: json.RawMessage(`{invalid`),
			wantError:  true,
		},
		{
			name:       "missing nodes",
			definition: json.RawMessage(`{"edges":[]}`),
			wantError:  true,
		},
		{
			name:       "missing edges",
			definition: json.RawMessage(`{"nodes":[]}`),
			wantError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.validateDefinition(tt.definition)
			if tt.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// Helper error for testing
var errTest = errors.New("test error")
