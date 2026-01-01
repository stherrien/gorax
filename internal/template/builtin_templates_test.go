package template

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetBuiltinTemplates(t *testing.T) {
	templates := GetBuiltinTemplates()

	assert.NotEmpty(t, templates, "should have built-in templates")
	assert.GreaterOrEqual(t, len(templates), 24, "should have at least 24 templates")

	// Check that all templates have required fields
	for _, tmpl := range templates {
		assert.NotEmpty(t, tmpl.Name, "template should have a name")
		assert.NotEmpty(t, tmpl.Description, "template should have a description")
		assert.NotEmpty(t, tmpl.Category, "template should have a category")
		assert.NotEmpty(t, tmpl.Tags, "template should have tags")
		assert.NotEmpty(t, tmpl.Definition, "template should have a definition")
		assert.True(t, tmpl.IsPublic, "built-in templates should be public")
		assert.Nil(t, tmpl.TenantID, "built-in templates should not have tenant_id")
		assert.Equal(t, "system", tmpl.CreatedBy, "built-in templates should be created by system")
	}
}

func TestDevOpsTemplates(t *testing.T) {
	templates := GetBuiltinTemplates()
	devOpsTemplates := filterByCategory(templates, string(CategoryDevOps))

	assert.GreaterOrEqual(t, len(devOpsTemplates), 3, "should have at least 3 DevOps templates")

	t.Run("CI/CD Notification Template", func(t *testing.T) {
		tmpl := findTemplateByName(devOpsTemplates, "CI/CD Pipeline Notification")
		require.NotNil(t, tmpl, "CI/CD notification template should exist")

		assert.Contains(t, tmpl.Description, "GitHub", "should mention GitHub")
		assert.Contains(t, tmpl.Description, "Slack", "should mention Slack")
		assert.Contains(t, tmpl.Tags, "cicd")
		assert.Contains(t, tmpl.Tags, "slack")
		assert.Contains(t, tmpl.Tags, "github")

		// Validate definition structure
		var def map[string]interface{}
		err := json.Unmarshal(tmpl.Definition, &def)
		require.NoError(t, err, "definition should be valid JSON")

		nodes, ok := def["nodes"].([]interface{})
		require.True(t, ok, "definition should have nodes array")
		assert.GreaterOrEqual(t, len(nodes), 2, "should have at least trigger and action nodes")

		edges, ok := def["edges"].([]interface{})
		require.True(t, ok, "definition should have edges array")
		assert.NotEmpty(t, edges, "should have edges connecting nodes")
	})

	t.Run("Deployment Approval Template", func(t *testing.T) {
		tmpl := findTemplateByName(devOpsTemplates, "Deployment Approval Workflow")
		require.NotNil(t, tmpl, "deployment approval template should exist")

		assert.Contains(t, tmpl.Tags, "approval")
		assert.Contains(t, tmpl.Tags, "deployment")
	})

	t.Run("Infrastructure Alert Handler", func(t *testing.T) {
		tmpl := findTemplateByName(devOpsTemplates, "Infrastructure Alert Handler")
		require.NotNil(t, tmpl, "infrastructure alert template should exist")

		assert.Contains(t, tmpl.Tags, "monitoring")
		assert.Contains(t, tmpl.Tags, "alerts")
	})
}

func TestBusinessProcessTemplates(t *testing.T) {
	templates := GetBuiltinTemplates()

	// Business templates might be under "integration" or custom "business" category
	businessTemplates := make([]*Template, 0)
	for _, tmpl := range templates {
		for _, tag := range tmpl.Tags {
			if tag == "business" || tag == "onboarding" || tag == "approval" || tag == "document" {
				businessTemplates = append(businessTemplates, tmpl)
				break
			}
		}
	}

	assert.GreaterOrEqual(t, len(businessTemplates), 3, "should have at least 3 business process templates")

	t.Run("Customer Onboarding Template", func(t *testing.T) {
		tmpl := findTemplateByName(templates, "Customer Onboarding Workflow")
		require.NotNil(t, tmpl, "customer onboarding template should exist")

		assert.Contains(t, tmpl.Tags, "onboarding")
		assert.Contains(t, tmpl.Tags, "business")
	})

	t.Run("Approval Request Template", func(t *testing.T) {
		tmpl := findTemplateByName(templates, "Approval Request Workflow")
		require.NotNil(t, tmpl, "approval request template should exist")

		assert.Contains(t, tmpl.Tags, "approval")
	})

	t.Run("Document Processing Template", func(t *testing.T) {
		tmpl := findTemplateByName(templates, "Document Processing Pipeline")
		require.NotNil(t, tmpl, "document processing template should exist")

		assert.Contains(t, tmpl.Tags, "document")
	})
}

func TestIntegrationTemplates(t *testing.T) {
	templates := GetBuiltinTemplates()
	integrationTemplates := filterByCategory(templates, string(CategoryIntegration))

	assert.GreaterOrEqual(t, len(integrationTemplates), 3, "should have at least 3 integration templates")

	t.Run("Slack to Jira Template", func(t *testing.T) {
		tmpl := findTemplateByName(templates, "Slack to Jira Ticket Creator")
		require.NotNil(t, tmpl, "Slack to Jira template should exist")

		assert.Contains(t, tmpl.Tags, "slack")
		assert.Contains(t, tmpl.Tags, "jira")
	})

	t.Run("Email to Task Template", func(t *testing.T) {
		tmpl := findTemplateByName(templates, "Email to Task Converter")
		require.NotNil(t, tmpl, "email to task template should exist")

		assert.Contains(t, tmpl.Tags, "email")
	})

	t.Run("Calendar Event Reminder", func(t *testing.T) {
		tmpl := findTemplateByName(templates, "Calendar Event Reminder")
		require.NotNil(t, tmpl, "calendar reminder template should exist")

		assert.Contains(t, tmpl.Tags, "calendar")
	})
}

func TestDataProcessingTemplates(t *testing.T) {
	templates := GetBuiltinTemplates()
	dataTemplates := filterByCategory(templates, string(CategoryDataOps))

	assert.GreaterOrEqual(t, len(dataTemplates), 3, "should have at least 3 data processing templates")

	t.Run("Data Sync Template", func(t *testing.T) {
		tmpl := findTemplateByName(templates, "Data Sync Workflow")
		require.NotNil(t, tmpl, "data sync template should exist")

		assert.Contains(t, tmpl.Tags, "sync")
		assert.Contains(t, tmpl.Tags, "database")
	})

	t.Run("Report Generation Template", func(t *testing.T) {
		tmpl := findTemplateByName(templates, "Report Generation Workflow")
		require.NotNil(t, tmpl, "report generation template should exist")

		assert.Contains(t, tmpl.Tags, "report")
	})

	t.Run("Data Validation Template", func(t *testing.T) {
		tmpl := findTemplateByName(templates, "Data Validation Pipeline")
		require.NotNil(t, tmpl, "data validation template should exist")

		assert.Contains(t, tmpl.Tags, "validation")
	})
}

func TestTemplateDefinitionStructure(t *testing.T) {
	templates := GetBuiltinTemplates()

	for _, tmpl := range templates {
		t.Run(tmpl.Name, func(t *testing.T) {
			var def map[string]interface{}
			err := json.Unmarshal(tmpl.Definition, &def)
			require.NoError(t, err, "definition should be valid JSON")

			// Check required fields
			nodes, ok := def["nodes"].([]interface{})
			require.True(t, ok, "definition should have nodes array")
			assert.NotEmpty(t, nodes, "should have at least one node")

			edges, ok := def["edges"].([]interface{})
			require.True(t, ok, "definition should have edges array")

			// If there are multiple nodes, there should be edges
			if len(nodes) > 1 {
				assert.NotEmpty(t, edges, "multi-node workflow should have edges")
			}

			// Validate each node has required fields
			for i, nodeData := range nodes {
				node, ok := nodeData.(map[string]interface{})
				require.True(t, ok, "node %d should be an object", i)

				_, hasID := node["id"]
				assert.True(t, hasID, "node %d should have id", i)

				_, hasType := node["type"]
				assert.True(t, hasType, "node %d should have type", i)

				_, hasData := node["data"]
				assert.True(t, hasData, "node %d should have data", i)

				_, hasPosition := node["position"]
				assert.True(t, hasPosition, "node %d should have position", i)
			}

			// Validate each edge has required fields
			for i, edgeData := range edges {
				edge, ok := edgeData.(map[string]interface{})
				require.True(t, ok, "edge %d should be an object", i)

				_, hasID := edge["id"]
				assert.True(t, hasID, "edge %d should have id", i)

				_, hasSource := edge["source"]
				assert.True(t, hasSource, "edge %d should have source", i)

				_, hasTarget := edge["target"]
				assert.True(t, hasTarget, "edge %d should have target", i)
			}
		})
	}
}

func TestGetTemplateByName(t *testing.T) {
	t.Run("existing template", func(t *testing.T) {
		tmpl := GetTemplateByName("CI/CD Pipeline Notification")
		require.NotNil(t, tmpl, "should find existing template")
		assert.Equal(t, "CI/CD Pipeline Notification", tmpl.Name)
	})

	t.Run("non-existing template", func(t *testing.T) {
		tmpl := GetTemplateByName("Non Existing Template")
		assert.Nil(t, tmpl, "should return nil for non-existing template")
	})
}

func TestGetTemplatesByCategory(t *testing.T) {
	t.Run("devops category", func(t *testing.T) {
		templates := GetTemplatesByCategory(string(CategoryDevOps))
		assert.NotEmpty(t, templates, "should have DevOps templates")

		for _, tmpl := range templates {
			assert.Equal(t, string(CategoryDevOps), tmpl.Category)
		}
	})

	t.Run("integration category", func(t *testing.T) {
		templates := GetTemplatesByCategory(string(CategoryIntegration))
		assert.NotEmpty(t, templates, "should have integration templates")

		for _, tmpl := range templates {
			assert.Equal(t, string(CategoryIntegration), tmpl.Category)
		}
	})

	t.Run("non-existing category", func(t *testing.T) {
		templates := GetTemplatesByCategory("non-existing")
		assert.Empty(t, templates, "should return empty array for non-existing category")
	})
}

func TestGetTemplatesByTag(t *testing.T) {
	t.Run("slack tag", func(t *testing.T) {
		templates := GetTemplatesByTag("slack")
		assert.NotEmpty(t, templates, "should have templates with slack tag")

		for _, tmpl := range templates {
			assert.Contains(t, tmpl.Tags, "slack")
		}
	})

	t.Run("github tag", func(t *testing.T) {
		templates := GetTemplatesByTag("github")
		assert.NotEmpty(t, templates, "should have templates with github tag")

		for _, tmpl := range templates {
			assert.Contains(t, tmpl.Tags, "github")
		}
	})

	t.Run("non-existing tag", func(t *testing.T) {
		templates := GetTemplatesByTag("non-existing-tag-xyz")
		assert.Empty(t, templates, "should return empty array for non-existing tag")
	})
}

func TestSeedBuiltinTemplates(t *testing.T) {
	mockRepo := &mockTemplateRepository{
		templates: make(map[string]*Template),
	}

	logger := slog.Default()
	service := NewService(mockRepo, logger)

	err := SeedBuiltinTemplates(service, "test-tenant")
	require.NoError(t, err, "seeding should succeed")

	// Verify all templates were seeded
	templates := GetBuiltinTemplates()
	assert.Len(t, mockRepo.templates, len(templates), "all templates should be seeded")

	// Verify templates are stored correctly
	for _, originalTmpl := range templates {
		storedTmpl, exists := mockRepo.templates[originalTmpl.Name]
		require.True(t, exists, "template %s should be stored", originalTmpl.Name)
		assert.Equal(t, originalTmpl.Name, storedTmpl.Name)
		assert.Equal(t, originalTmpl.Category, storedTmpl.Category)
		assert.Equal(t, originalTmpl.Description, storedTmpl.Description)
		assert.Equal(t, originalTmpl.Tags, storedTmpl.Tags)
		assert.Equal(t, "system", storedTmpl.CreatedBy)
	}
}

func TestSeedBuiltinTemplates_SkipsExisting(t *testing.T) {
	mockRepo := &mockTemplateRepository{
		templates: make(map[string]*Template),
	}

	logger := slog.Default()
	service := NewService(mockRepo, logger)

	// First seed
	err := SeedBuiltinTemplates(service, "test-tenant")
	require.NoError(t, err, "first seeding should succeed")

	originalCount := len(mockRepo.templates)

	// Second seed (should not duplicate)
	err = SeedBuiltinTemplates(service, "test-tenant")
	require.NoError(t, err, "second seeding should succeed")

	// Count should be the same (no duplicates)
	assert.Equal(t, originalCount, len(mockRepo.templates), "should not create duplicates")
}

// Helper functions

// mockTemplateRepository for testing seeding
type mockTemplateRepository struct {
	templates map[string]*Template
}

func (m *mockTemplateRepository) Create(ctx context.Context, tenantID string, template *Template) error {
	// Check if template already exists
	if _, exists := m.templates[template.Name]; exists {
		return fmt.Errorf("template with name %s already exists", template.Name)
	}

	// Generate ID if not set
	if template.ID == "" {
		template.ID = "tmpl-" + template.Name
	}

	m.templates[template.Name] = template
	return nil
}

func (m *mockTemplateRepository) GetByID(ctx context.Context, tenantID, id string) (*Template, error) {
	for _, tmpl := range m.templates {
		if tmpl.ID == id {
			return tmpl, nil
		}
	}
	return nil, fmt.Errorf("template not found")
}

func (m *mockTemplateRepository) List(ctx context.Context, tenantID string, filter TemplateFilter) ([]*Template, error) {
	result := make([]*Template, 0, len(m.templates))
	for _, tmpl := range m.templates {
		result = append(result, tmpl)
	}
	return result, nil
}

func (m *mockTemplateRepository) Update(ctx context.Context, tenantID, id string, input UpdateTemplateInput) error {
	for _, tmpl := range m.templates {
		if tmpl.ID == id {
			if input.Name != "" {
				delete(m.templates, tmpl.Name)
				tmpl.Name = input.Name
				m.templates[tmpl.Name] = tmpl
			}
			return nil
		}
	}
	return fmt.Errorf("template not found")
}

func (m *mockTemplateRepository) Delete(ctx context.Context, tenantID, id string) error {
	for name, tmpl := range m.templates {
		if tmpl.ID == id {
			delete(m.templates, name)
			return nil
		}
	}
	return fmt.Errorf("template not found")
}

func (m *mockTemplateRepository) IncrementUsageCount(ctx context.Context, id string) error {
	for _, tmpl := range m.templates {
		if tmpl.ID == id {
			tmpl.UsageCount++
			return nil
		}
	}
	return fmt.Errorf("template not found")
}

func filterByCategory(templates []*Template, category string) []*Template {
	result := make([]*Template, 0)
	for _, tmpl := range templates {
		if tmpl.Category == category {
			result = append(result, tmpl)
		}
	}
	return result
}

func findTemplateByName(templates []*Template, name string) *Template {
	for _, tmpl := range templates {
		if tmpl.Name == name {
			return tmpl
		}
	}
	return nil
}

// Tests for new templates

func TestNewTemplates(t *testing.T) {
	templates := GetBuiltinTemplates()

	t.Run("Slack Alert Notification Template", func(t *testing.T) {
		tmpl := findTemplateByName(templates, "Slack Alert Notification")
		require.NotNil(t, tmpl, "Slack alert notification template should exist")

		assert.Contains(t, tmpl.Description, "Slack", "should mention Slack")
		assert.Contains(t, tmpl.Tags, "slack")
		assert.Contains(t, tmpl.Tags, "alert")
		assert.Contains(t, tmpl.Tags, "notification")
		assert.Equal(t, string(CategoryIntegration), tmpl.Category)

		// Validate definition structure
		var def map[string]interface{}
		err := json.Unmarshal(tmpl.Definition, &def)
		require.NoError(t, err, "definition should be valid JSON")

		nodes, ok := def["nodes"].([]interface{})
		require.True(t, ok, "definition should have nodes array")
		assert.GreaterOrEqual(t, len(nodes), 3, "should have at least trigger, transform, and action nodes")

		edges, ok := def["edges"].([]interface{})
		require.True(t, ok, "definition should have edges array")
		assert.NotEmpty(t, edges, "should have edges connecting nodes")
	})

	t.Run("GitHub PR Automation Template", func(t *testing.T) {
		tmpl := findTemplateByName(templates, "GitHub PR Automation")
		require.NotNil(t, tmpl, "GitHub PR automation template should exist")

		assert.Contains(t, tmpl.Description, "GitHub", "should mention GitHub")
		assert.Contains(t, tmpl.Tags, "github")
		assert.Contains(t, tmpl.Tags, "pr")
		assert.Contains(t, tmpl.Tags, "automation")
		assert.Equal(t, string(CategoryDevOps), tmpl.Category)

		var def map[string]interface{}
		err := json.Unmarshal(tmpl.Definition, &def)
		require.NoError(t, err, "definition should be valid JSON")

		nodes := def["nodes"].([]interface{})
		assert.GreaterOrEqual(t, len(nodes), 4, "should have multiple nodes for PR processing")
	})

	t.Run("Data Backup Workflow Template", func(t *testing.T) {
		tmpl := findTemplateByName(templates, "Data Backup Workflow")
		require.NotNil(t, tmpl, "data backup workflow template should exist")

		assert.Contains(t, tmpl.Description, "backup", "should mention backup")
		assert.Contains(t, tmpl.Tags, "backup")
		assert.Contains(t, tmpl.Tags, "schedule")
		assert.Contains(t, tmpl.Tags, "dataops")
		assert.Equal(t, string(CategoryDataOps), tmpl.Category)

		var def map[string]interface{}
		err := json.Unmarshal(tmpl.Definition, &def)
		require.NoError(t, err, "definition should be valid JSON")

		// Should have a schedule trigger
		nodes := def["nodes"].([]interface{})
		hasScheduleTrigger := false
		for _, nodeData := range nodes {
			node := nodeData.(map[string]interface{})
			if nodeType, ok := node["type"].(string); ok && nodeType == "trigger:schedule" {
				hasScheduleTrigger = true
				break
			}
		}
		assert.True(t, hasScheduleTrigger, "backup workflow should have schedule trigger")
	})

	t.Run("Error Monitoring Workflow Template", func(t *testing.T) {
		tmpl := findTemplateByName(templates, "Error Monitoring Workflow")
		require.NotNil(t, tmpl, "error monitoring workflow template should exist")

		assert.Contains(t, tmpl.Description, "error", "should mention error")
		assert.Contains(t, tmpl.Tags, "error")
		assert.Contains(t, tmpl.Tags, "monitoring")
		assert.Contains(t, tmpl.Tags, "alert")
		assert.Equal(t, string(CategoryMonitoring), tmpl.Category)

		var def map[string]interface{}
		err := json.Unmarshal(tmpl.Definition, &def)
		require.NoError(t, err, "definition should be valid JSON")

		nodes := def["nodes"].([]interface{})
		assert.GreaterOrEqual(t, len(nodes), 3, "should have trigger, condition, and alert nodes")
	})

	t.Run("Multi-Step User Onboarding Template", func(t *testing.T) {
		tmpl := findTemplateByName(templates, "Multi-Step User Onboarding")
		require.NotNil(t, tmpl, "multi-step user onboarding template should exist")

		assert.Contains(t, tmpl.Description, "onboarding", "should mention onboarding")
		assert.Contains(t, tmpl.Tags, "onboarding")
		assert.Contains(t, tmpl.Tags, "user")
		assert.Contains(t, tmpl.Tags, "business")
		assert.Equal(t, string(CategoryIntegration), tmpl.Category)

		var def map[string]interface{}
		err := json.Unmarshal(tmpl.Definition, &def)
		require.NoError(t, err, "definition should be valid JSON")

		nodes := def["nodes"].([]interface{})
		assert.GreaterOrEqual(t, len(nodes), 5, "should have multiple steps for onboarding")
	})

	t.Run("API Health Check Workflow Template", func(t *testing.T) {
		tmpl := findTemplateByName(templates, "API Health Check Workflow")
		require.NotNil(t, tmpl, "API health check workflow template should exist")

		assert.Contains(t, tmpl.Description, "health", "should mention health check")
		assert.Contains(t, tmpl.Tags, "health")
		assert.Contains(t, tmpl.Tags, "monitoring")
		assert.Contains(t, tmpl.Tags, "api")
		assert.Equal(t, string(CategoryMonitoring), tmpl.Category)

		var def map[string]interface{}
		err := json.Unmarshal(tmpl.Definition, &def)
		require.NoError(t, err, "definition should be valid JSON")

		// Should have schedule trigger for periodic checks
		nodes := def["nodes"].([]interface{})
		hasScheduleTrigger := false
		for _, nodeData := range nodes {
			node := nodeData.(map[string]interface{})
			if nodeType, ok := node["type"].(string); ok && nodeType == "trigger:schedule" {
				hasScheduleTrigger = true
				break
			}
		}
		assert.True(t, hasScheduleTrigger, "health check should have schedule trigger")
	})

	t.Run("Scheduled Report Generation Template", func(t *testing.T) {
		tmpl := findTemplateByName(templates, "Scheduled Report Generation")
		require.NotNil(t, tmpl, "scheduled report generation template should exist")

		assert.Contains(t, tmpl.Description, "report", "should mention report")
		assert.Contains(t, tmpl.Tags, "report")
		assert.Contains(t, tmpl.Tags, "schedule")
		assert.Contains(t, tmpl.Tags, "analytics")
		assert.Equal(t, string(CategoryDataOps), tmpl.Category)

		var def map[string]interface{}
		err := json.Unmarshal(tmpl.Definition, &def)
		require.NoError(t, err, "definition should be valid JSON")

		nodes := def["nodes"].([]interface{})
		assert.GreaterOrEqual(t, len(nodes), 4, "should have data fetching, transformation, and notification")
	})
}

// Tests for additional workflow templates

func TestDataETLSyncTemplate(t *testing.T) {
	templates := GetBuiltinTemplates()
	tmpl := findTemplateByName(templates, "Data ETL Sync Workflow")
	require.NotNil(t, tmpl, "Data ETL Sync workflow template should exist")

	assert.Contains(t, tmpl.Description, "ETL", "should mention ETL")
	assert.Contains(t, tmpl.Tags, "etl")
	assert.Contains(t, tmpl.Tags, "sync")
	assert.Contains(t, tmpl.Tags, "dataops")
	assert.Equal(t, string(CategoryDataOps), tmpl.Category)

	// Validate definition structure
	var def map[string]interface{}
	err := json.Unmarshal(tmpl.Definition, &def)
	require.NoError(t, err, "definition should be valid JSON")

	nodes, ok := def["nodes"].([]interface{})
	require.True(t, ok, "definition should have nodes array")
	assert.GreaterOrEqual(t, len(nodes), 5, "should have extract, transform, load, validate, and notification nodes")

	edges, ok := def["edges"].([]interface{})
	require.True(t, ok, "definition should have edges array")
	assert.NotEmpty(t, edges, "should have edges connecting nodes")

	// Should have schedule trigger for periodic sync
	hasScheduleTrigger := false
	for _, nodeData := range nodes {
		node := nodeData.(map[string]interface{})
		if nodeType, ok := node["type"].(string); ok && nodeType == "trigger:schedule" {
			hasScheduleTrigger = true
			break
		}
	}
	assert.True(t, hasScheduleTrigger, "ETL workflow should have schedule trigger")
}

func TestScheduledReportingWorkflowTemplate(t *testing.T) {
	templates := GetBuiltinTemplates()
	tmpl := findTemplateByName(templates, "Scheduled Reporting Workflow")
	require.NotNil(t, tmpl, "Scheduled Reporting workflow template should exist")

	assert.Contains(t, tmpl.Description, "report", "should mention reporting")
	assert.Contains(t, tmpl.Tags, "report")
	assert.Contains(t, tmpl.Tags, "schedule")
	assert.Contains(t, tmpl.Tags, "analytics")
	assert.Equal(t, string(CategoryDataOps), tmpl.Category)

	var def map[string]interface{}
	err := json.Unmarshal(tmpl.Definition, &def)
	require.NoError(t, err, "definition should be valid JSON")

	nodes := def["nodes"].([]interface{})
	assert.GreaterOrEqual(t, len(nodes), 5, "should have trigger, data fetch, transform, report generation, and notification")

	// Should have schedule trigger
	hasScheduleTrigger := false
	for _, nodeData := range nodes {
		node := nodeData.(map[string]interface{})
		if nodeType, ok := node["type"].(string); ok && nodeType == "trigger:schedule" {
			hasScheduleTrigger = true
			break
		}
	}
	assert.True(t, hasScheduleTrigger, "reporting workflow should have schedule trigger")
}

func TestMultiStepApprovalWorkflowTemplate(t *testing.T) {
	templates := GetBuiltinTemplates()
	tmpl := findTemplateByName(templates, "Multi-Step Approval Workflow")
	require.NotNil(t, tmpl, "Multi-Step Approval workflow template should exist")

	assert.Contains(t, tmpl.Description, "approval", "should mention approval")
	assert.Contains(t, tmpl.Tags, "approval")
	assert.Contains(t, tmpl.Tags, "business")
	assert.Contains(t, tmpl.Tags, "workflow")
	assert.Equal(t, string(CategoryIntegration), tmpl.Category)

	var def map[string]interface{}
	err := json.Unmarshal(tmpl.Definition, &def)
	require.NoError(t, err, "definition should be valid JSON")

	nodes := def["nodes"].([]interface{})
	assert.GreaterOrEqual(t, len(nodes), 6, "should have multiple approval steps with notifications")

	// Should have webhook trigger
	hasWebhookTrigger := false
	for _, nodeData := range nodes {
		node := nodeData.(map[string]interface{})
		if nodeType, ok := node["type"].(string); ok && nodeType == "trigger:webhook" {
			hasWebhookTrigger = true
			break
		}
	}
	assert.True(t, hasWebhookTrigger, "approval workflow should have webhook trigger")
}

func TestErrorNotificationWorkflowTemplate(t *testing.T) {
	templates := GetBuiltinTemplates()
	tmpl := findTemplateByName(templates, "Error Notification Workflow")
	require.NotNil(t, tmpl, "Error Notification workflow template should exist")

	assert.Contains(t, tmpl.Description, "error", "should mention error")
	assert.Contains(t, tmpl.Tags, "error")
	assert.Contains(t, tmpl.Tags, "notification")
	assert.Contains(t, tmpl.Tags, "monitoring")
	assert.Equal(t, string(CategoryMonitoring), tmpl.Category)

	var def map[string]interface{}
	err := json.Unmarshal(tmpl.Definition, &def)
	require.NoError(t, err, "definition should be valid JSON")

	nodes := def["nodes"].([]interface{})
	assert.GreaterOrEqual(t, len(nodes), 4, "should have trigger, enrichment, severity check, and notifications")

	// Should have condition node for severity checking
	hasConditionNode := false
	for _, nodeData := range nodes {
		node := nodeData.(map[string]interface{})
		if nodeType, ok := node["type"].(string); ok && nodeType == "control:if" {
			hasConditionNode = true
			break
		}
	}
	assert.True(t, hasConditionNode, "error notification should have condition for severity")
}

func TestAPIOrchestrationWorkflowTemplate(t *testing.T) {
	templates := GetBuiltinTemplates()
	tmpl := findTemplateByName(templates, "API Orchestration Workflow")
	require.NotNil(t, tmpl, "API Orchestration workflow template should exist")

	assert.Contains(t, tmpl.Description, "API", "should mention API")
	assert.Contains(t, tmpl.Tags, "api")
	assert.Contains(t, tmpl.Tags, "orchestration")
	assert.Contains(t, tmpl.Tags, "integration")
	assert.Equal(t, string(CategoryIntegration), tmpl.Category)

	var def map[string]interface{}
	err := json.Unmarshal(tmpl.Definition, &def)
	require.NoError(t, err, "definition should be valid JSON")

	nodes := def["nodes"].([]interface{})
	assert.GreaterOrEqual(t, len(nodes), 5, "should have trigger and multiple API calls")

	// Should have multiple HTTP action nodes
	httpActionCount := 0
	for _, nodeData := range nodes {
		node := nodeData.(map[string]interface{})
		if nodeType, ok := node["type"].(string); ok && nodeType == "action:http" {
			httpActionCount++
		}
	}
	assert.GreaterOrEqual(t, httpActionCount, 3, "orchestration should have multiple HTTP actions")
}

func TestAllNewTemplatesCount(t *testing.T) {
	templates := GetBuiltinTemplates()

	// With the 5 new templates, we should have at least 24 templates total
	assert.GreaterOrEqual(t, len(templates), 24, "should have at least 24 templates with new additions")
}
