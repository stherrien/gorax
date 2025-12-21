package aibuilder

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNodeCategory(t *testing.T) {
	tests := []struct {
		name     string
		category NodeCategory
		valid    bool
	}{
		{"trigger", NodeCategoryTrigger, true},
		{"action", NodeCategoryAction, true},
		{"control", NodeCategoryControl, true},
		{"integration", NodeCategoryIntegration, true},
		{"invalid", NodeCategory("invalid"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.valid, tt.category.IsValid())
		})
	}
}

func TestNodeTemplate_Validate(t *testing.T) {
	tests := []struct {
		name     string
		template NodeTemplate
		wantErr  bool
		errMsg   string
	}{
		{
			name: "valid template",
			template: NodeTemplate{
				NodeType:       "trigger:webhook",
				Name:           "Webhook Trigger",
				Description:    "Starts workflow on webhook",
				Category:       NodeCategoryTrigger,
				LLMDescription: "Use this to start a workflow when a webhook is received",
			},
			wantErr: false,
		},
		{
			name: "empty node type",
			template: NodeTemplate{
				NodeType:       "",
				Name:           "Webhook Trigger",
				Description:    "Starts workflow on webhook",
				Category:       NodeCategoryTrigger,
				LLMDescription: "Use this for webhooks",
			},
			wantErr: true,
			errMsg:  "node_type is required",
		},
		{
			name: "empty name",
			template: NodeTemplate{
				NodeType:       "trigger:webhook",
				Name:           "",
				Description:    "Starts workflow on webhook",
				Category:       NodeCategoryTrigger,
				LLMDescription: "Use this for webhooks",
			},
			wantErr: true,
			errMsg:  "name is required",
		},
		{
			name: "empty description",
			template: NodeTemplate{
				NodeType:       "trigger:webhook",
				Name:           "Webhook Trigger",
				Description:    "",
				Category:       NodeCategoryTrigger,
				LLMDescription: "Use this for webhooks",
			},
			wantErr: true,
			errMsg:  "description is required",
		},
		{
			name: "invalid category",
			template: NodeTemplate{
				NodeType:       "trigger:webhook",
				Name:           "Webhook Trigger",
				Description:    "Starts workflow on webhook",
				Category:       NodeCategory("invalid"),
				LLMDescription: "Use this for webhooks",
			},
			wantErr: true,
			errMsg:  "invalid category",
		},
		{
			name: "empty LLM description",
			template: NodeTemplate{
				NodeType:       "trigger:webhook",
				Name:           "Webhook Trigger",
				Description:    "Starts workflow on webhook",
				Category:       NodeCategoryTrigger,
				LLMDescription: "",
			},
			wantErr: true,
			errMsg:  "llm_description is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.template.Validate()
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestNewNodeRegistry(t *testing.T) {
	registry := NewNodeRegistry()

	assert.NotNil(t, registry)
	assert.NotNil(t, registry.templates)
	assert.NotNil(t, registry.byCategory)
}

func TestNodeRegistry_Register(t *testing.T) {
	registry := NewNodeRegistry()

	template := NodeTemplate{
		NodeType:       "trigger:webhook",
		Name:           "Webhook Trigger",
		Description:    "Starts workflow on webhook",
		Category:       NodeCategoryTrigger,
		LLMDescription: "Use this for webhooks",
	}

	err := registry.Register(template)
	require.NoError(t, err)

	// Should be able to get the template
	got, exists := registry.Get("trigger:webhook")
	assert.True(t, exists)
	assert.Equal(t, template.Name, got.Name)
}

func TestNodeRegistry_Register_Duplicate(t *testing.T) {
	registry := NewNodeRegistry()

	template := NodeTemplate{
		NodeType:       "trigger:webhook",
		Name:           "Webhook Trigger",
		Description:    "Starts workflow on webhook",
		Category:       NodeCategoryTrigger,
		LLMDescription: "Use this for webhooks",
	}

	err := registry.Register(template)
	require.NoError(t, err)

	// Registering again should fail
	err = registry.Register(template)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "already registered")
}

func TestNodeRegistry_Register_Invalid(t *testing.T) {
	registry := NewNodeRegistry()

	template := NodeTemplate{
		NodeType: "", // Invalid - empty
		Name:     "Test",
	}

	err := registry.Register(template)
	require.Error(t, err)
}

func TestNodeRegistry_Get(t *testing.T) {
	registry := NewNodeRegistry()

	template := NodeTemplate{
		NodeType:       "action:http",
		Name:           "HTTP Request",
		Description:    "Makes HTTP calls",
		Category:       NodeCategoryAction,
		LLMDescription: "Use this for API calls",
	}
	registry.Register(template)

	t.Run("existing template", func(t *testing.T) {
		got, exists := registry.Get("action:http")
		assert.True(t, exists)
		assert.Equal(t, "HTTP Request", got.Name)
	})

	t.Run("non-existing template", func(t *testing.T) {
		got, exists := registry.Get("action:nonexistent")
		assert.False(t, exists)
		assert.Nil(t, got)
	})
}

func TestNodeRegistry_GetAll(t *testing.T) {
	registry := NewNodeRegistry()

	templates := []NodeTemplate{
		{NodeType: "trigger:webhook", Name: "Webhook", Description: "Webhook trigger", Category: NodeCategoryTrigger, LLMDescription: "Webhook"},
		{NodeType: "action:http", Name: "HTTP", Description: "HTTP action", Category: NodeCategoryAction, LLMDescription: "HTTP"},
		{NodeType: "control:if", Name: "If", Description: "Conditional", Category: NodeCategoryControl, LLMDescription: "If"},
	}

	for _, t := range templates {
		registry.Register(t)
	}

	all := registry.GetAll()
	assert.Len(t, all, 3)
}

func TestNodeRegistry_GetByCategory(t *testing.T) {
	registry := NewNodeRegistry()

	templates := []NodeTemplate{
		{NodeType: "trigger:webhook", Name: "Webhook", Description: "Webhook trigger", Category: NodeCategoryTrigger, LLMDescription: "Webhook"},
		{NodeType: "trigger:schedule", Name: "Schedule", Description: "Schedule trigger", Category: NodeCategoryTrigger, LLMDescription: "Schedule"},
		{NodeType: "action:http", Name: "HTTP", Description: "HTTP action", Category: NodeCategoryAction, LLMDescription: "HTTP"},
		{NodeType: "control:if", Name: "If", Description: "Conditional", Category: NodeCategoryControl, LLMDescription: "If"},
	}

	for _, tmpl := range templates {
		registry.Register(tmpl)
	}

	t.Run("triggers", func(t *testing.T) {
		triggers := registry.GetByCategory(NodeCategoryTrigger)
		assert.Len(t, triggers, 2)
	})

	t.Run("actions", func(t *testing.T) {
		actions := registry.GetByCategory(NodeCategoryAction)
		assert.Len(t, actions, 1)
	})

	t.Run("controls", func(t *testing.T) {
		controls := registry.GetByCategory(NodeCategoryControl)
		assert.Len(t, controls, 1)
	})

	t.Run("integrations (empty)", func(t *testing.T) {
		integrations := registry.GetByCategory(NodeCategoryIntegration)
		assert.Len(t, integrations, 0)
	})
}

func TestNodeRegistry_BuildLLMContext(t *testing.T) {
	registry := NewNodeRegistry()

	templates := []NodeTemplate{
		{
			NodeType:       "trigger:webhook",
			Name:           "Webhook Trigger",
			Description:    "Starts workflow on webhook",
			Category:       NodeCategoryTrigger,
			LLMDescription: "Use this to start a workflow when an HTTP webhook is received",
			ExampleConfig:  map[string]interface{}{"path": "/my-webhook"},
		},
		{
			NodeType:       "action:http",
			Name:           "HTTP Request",
			Description:    "Makes HTTP calls",
			Category:       NodeCategoryAction,
			LLMDescription: "Use this to call REST APIs",
			ExampleConfig:  map[string]interface{}{"method": "POST", "url": "https://api.example.com"},
		},
	}

	for _, tmpl := range templates {
		registry.Register(tmpl)
	}

	context := registry.BuildLLMContext()

	assert.Contains(t, context, "trigger:webhook")
	assert.Contains(t, context, "action:http")
	assert.Contains(t, context, "Webhook Trigger")
	assert.Contains(t, context, "HTTP Request")
	assert.Contains(t, context, "Use this to start a workflow")
	assert.Contains(t, context, "Use this to call REST APIs")
}

func TestNodeRegistry_GetNodeTypes(t *testing.T) {
	registry := NewNodeRegistry()

	templates := []NodeTemplate{
		{NodeType: "trigger:webhook", Name: "Webhook", Description: "Webhook trigger", Category: NodeCategoryTrigger, LLMDescription: "Webhook"},
		{NodeType: "action:http", Name: "HTTP", Description: "HTTP action", Category: NodeCategoryAction, LLMDescription: "HTTP"},
	}

	for _, tmpl := range templates {
		registry.Register(tmpl)
	}

	types := registry.GetNodeTypes()
	assert.Len(t, types, 2)
	assert.Contains(t, types, "trigger:webhook")
	assert.Contains(t, types, "action:http")
}

func TestDefaultNodeRegistry(t *testing.T) {
	registry := DefaultNodeRegistry()

	assert.NotNil(t, registry)

	// Should have default triggers
	triggers := registry.GetByCategory(NodeCategoryTrigger)
	assert.GreaterOrEqual(t, len(triggers), 2, "Should have at least webhook and schedule triggers")

	// Should have default actions
	actions := registry.GetByCategory(NodeCategoryAction)
	assert.GreaterOrEqual(t, len(actions), 4, "Should have at least http, transform, code, formula actions")

	// Should have control flow nodes
	controls := registry.GetByCategory(NodeCategoryControl)
	assert.GreaterOrEqual(t, len(controls), 3, "Should have at least if, loop, delay controls")

	// Verify specific templates exist
	webhook, exists := registry.Get("trigger:webhook")
	assert.True(t, exists)
	assert.Equal(t, "Webhook Trigger", webhook.Name)

	http, exists := registry.Get("action:http")
	assert.True(t, exists)
	assert.Equal(t, "HTTP Request", http.Name)

	ifNode, exists := registry.Get("control:if")
	assert.True(t, exists)
	assert.Equal(t, "Conditional (If/Else)", ifNode.Name)
}

func TestNodeTemplate_ToGeneratedNode(t *testing.T) {
	template := NodeTemplate{
		NodeType:      "trigger:webhook",
		Name:          "Webhook Trigger",
		Description:   "Starts workflow on webhook",
		Category:      NodeCategoryTrigger,
		ExampleConfig: map[string]interface{}{"path": "/webhook"},
	}

	node := template.ToGeneratedNode("node-123")

	assert.Equal(t, "node-123", node.ID)
	assert.Equal(t, "trigger:webhook", node.Type)
	assert.Equal(t, "Webhook Trigger", node.Name)
	assert.NotNil(t, node.Config)
}

func TestNodeRegistry_IsValidType(t *testing.T) {
	registry := NewNodeRegistry()

	template := NodeTemplate{
		NodeType:       "trigger:webhook",
		Name:           "Webhook Trigger",
		Description:    "Starts workflow",
		Category:       NodeCategoryTrigger,
		LLMDescription: "Use for webhooks",
	}
	registry.Register(template)

	assert.True(t, registry.IsValidType("trigger:webhook"))
	assert.False(t, registry.IsValidType("trigger:unknown"))
}
