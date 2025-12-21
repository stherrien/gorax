package aibuilder

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
)

// NodeCategory represents the category of a node type
type NodeCategory string

const (
	NodeCategoryTrigger     NodeCategory = "trigger"
	NodeCategoryAction      NodeCategory = "action"
	NodeCategoryControl     NodeCategory = "control"
	NodeCategoryIntegration NodeCategory = "integration"
)

// IsValid checks if the node category is valid
func (c NodeCategory) IsValid() bool {
	switch c {
	case NodeCategoryTrigger, NodeCategoryAction, NodeCategoryControl, NodeCategoryIntegration:
		return true
	default:
		return false
	}
}

// NodeTemplate represents a template for a node type
type NodeTemplate struct {
	ID             string                 `json:"id,omitempty" db:"id"`
	TenantID       *string                `json:"tenant_id,omitempty" db:"tenant_id"`
	NodeType       string                 `json:"node_type" db:"node_type"`
	Name           string                 `json:"name" db:"name"`
	Description    string                 `json:"description" db:"description"`
	Category       NodeCategory           `json:"category" db:"category"`
	ConfigSchema   map[string]interface{} `json:"config_schema,omitempty" db:"config_schema"`
	ExampleConfig  map[string]interface{} `json:"example_config,omitempty" db:"example_config"`
	LLMDescription string                 `json:"llm_description" db:"llm_description"`
	IsActive       bool                   `json:"is_active" db:"is_active"`
}

// Validate validates the node template
func (t *NodeTemplate) Validate() error {
	if t.NodeType == "" {
		return errors.New("node_type is required")
	}
	if t.Name == "" {
		return errors.New("name is required")
	}
	if t.Description == "" {
		return errors.New("description is required")
	}
	if !t.Category.IsValid() {
		return errors.New("invalid category")
	}
	if t.LLMDescription == "" {
		return errors.New("llm_description is required")
	}
	return nil
}

// ToGeneratedNode converts the template to a generated node with the given ID
func (t *NodeTemplate) ToGeneratedNode(nodeID string) *GeneratedNode {
	var config json.RawMessage
	if t.ExampleConfig != nil {
		configBytes, _ := json.Marshal(t.ExampleConfig)
		config = configBytes
	}

	return &GeneratedNode{
		ID:          nodeID,
		Type:        t.NodeType,
		Name:        t.Name,
		Description: t.Description,
		Config:      config,
	}
}

// NodeRegistry manages available node types for workflow generation
type NodeRegistry struct {
	templates  map[string]*NodeTemplate
	byCategory map[NodeCategory][]*NodeTemplate
	mu         sync.RWMutex
}

// NewNodeRegistry creates a new empty node registry
func NewNodeRegistry() *NodeRegistry {
	return &NodeRegistry{
		templates:  make(map[string]*NodeTemplate),
		byCategory: make(map[NodeCategory][]*NodeTemplate),
	}
}

// Register adds a node template to the registry
func (r *NodeRegistry) Register(template NodeTemplate) error {
	if err := template.Validate(); err != nil {
		return err
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.templates[template.NodeType]; exists {
		return fmt.Errorf("node type %s already registered", template.NodeType)
	}

	t := &template
	r.templates[template.NodeType] = t
	r.byCategory[template.Category] = append(r.byCategory[template.Category], t)

	return nil
}

// Get returns a node template by type
func (r *NodeRegistry) Get(nodeType string) (*NodeTemplate, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	t, exists := r.templates[nodeType]
	return t, exists
}

// GetAll returns all registered node templates
func (r *NodeRegistry) GetAll() []*NodeTemplate {
	r.mu.RLock()
	defer r.mu.RUnlock()

	templates := make([]*NodeTemplate, 0, len(r.templates))
	for _, t := range r.templates {
		templates = append(templates, t)
	}
	return templates
}

// GetByCategory returns all templates in a category
func (r *NodeRegistry) GetByCategory(category NodeCategory) []*NodeTemplate {
	r.mu.RLock()
	defer r.mu.RUnlock()

	templates := r.byCategory[category]
	if templates == nil {
		return []*NodeTemplate{}
	}
	return templates
}

// GetNodeTypes returns all registered node type identifiers
func (r *NodeRegistry) GetNodeTypes() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	types := make([]string, 0, len(r.templates))
	for nodeType := range r.templates {
		types = append(types, nodeType)
	}
	return types
}

// IsValidType checks if a node type is registered
func (r *NodeRegistry) IsValidType(nodeType string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	_, exists := r.templates[nodeType]
	return exists
}

// BuildLLMContext generates context text for LLM prompts
func (r *NodeRegistry) BuildLLMContext() string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var sb strings.Builder

	sb.WriteString("Available Node Types:\n\n")

	categories := []NodeCategory{
		NodeCategoryTrigger,
		NodeCategoryAction,
		NodeCategoryControl,
		NodeCategoryIntegration,
	}

	for _, category := range categories {
		templates := r.byCategory[category]
		if len(templates) == 0 {
			continue
		}

		sb.WriteString(fmt.Sprintf("## %s Nodes\n\n", strings.Title(string(category))))

		for _, t := range templates {
			sb.WriteString(fmt.Sprintf("### %s (%s)\n", t.Name, t.NodeType))
			sb.WriteString(fmt.Sprintf("%s\n\n", t.LLMDescription))

			if t.ExampleConfig != nil {
				configJSON, _ := json.MarshalIndent(t.ExampleConfig, "", "  ")
				sb.WriteString(fmt.Sprintf("Example config:\n```json\n%s\n```\n\n", string(configJSON)))
			}
		}
	}

	return sb.String()
}

// DefaultNodeRegistry creates a registry with all default node types
func DefaultNodeRegistry() *NodeRegistry {
	registry := NewNodeRegistry()

	// Triggers
	registry.Register(NodeTemplate{
		NodeType:    "trigger:webhook",
		Name:        "Webhook Trigger",
		Description: "Starts workflow when an HTTP webhook is received",
		Category:    NodeCategoryTrigger,
		ConfigSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"path":      map[string]interface{}{"type": "string"},
				"auth_type": map[string]interface{}{"type": "string", "enum": []string{"none", "basic", "signature", "api_key"}},
				"secret":    map[string]interface{}{"type": "string"},
			},
		},
		ExampleConfig: map[string]interface{}{
			"path":      "/my-webhook",
			"auth_type": "signature",
			"secret":    "my-secret",
		},
		LLMDescription: "Use this to start a workflow when an external system sends an HTTP POST request. " +
			"Configure authentication (none, basic auth, signature verification, or API key) for security. " +
			"The webhook payload will be available in subsequent steps via ${steps.trigger.body}.",
		IsActive: true,
	})

	registry.Register(NodeTemplate{
		NodeType:    "trigger:schedule",
		Name:        "Schedule Trigger",
		Description: "Starts workflow on a cron schedule",
		Category:    NodeCategoryTrigger,
		ConfigSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"cron":     map[string]interface{}{"type": "string"},
				"timezone": map[string]interface{}{"type": "string"},
			},
		},
		ExampleConfig: map[string]interface{}{
			"cron":     "0 9 * * 1-5",
			"timezone": "America/New_York",
		},
		LLMDescription: "Use this to run a workflow automatically at scheduled times. " +
			"Specify a cron expression (e.g., '0 9 * * 1-5' for 9am weekdays) and optional timezone.",
		IsActive: true,
	})

	// Actions
	registry.Register(NodeTemplate{
		NodeType:    "action:http",
		Name:        "HTTP Request",
		Description: "Makes HTTP API calls to external services",
		Category:    NodeCategoryAction,
		ConfigSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"method":  map[string]interface{}{"type": "string", "enum": []string{"GET", "POST", "PUT", "PATCH", "DELETE"}},
				"url":     map[string]interface{}{"type": "string"},
				"headers": map[string]interface{}{"type": "object"},
				"body":    map[string]interface{}{"type": "object"},
				"timeout": map[string]interface{}{"type": "integer"},
			},
		},
		ExampleConfig: map[string]interface{}{
			"method":  "POST",
			"url":     "https://api.example.com/data",
			"headers": map[string]interface{}{"Content-Type": "application/json"},
			"body":    map[string]interface{}{"key": "${steps.trigger.body.value}"},
			"timeout": 30,
		},
		LLMDescription: "Use this to call REST APIs. Configure the HTTP method, URL, headers, and request body. " +
			"Supports template variables like ${steps.trigger.body.data} for dynamic values from previous steps.",
		IsActive: true,
	})

	registry.Register(NodeTemplate{
		NodeType:    "action:transform",
		Name:        "Transform Data",
		Description: "Transforms and maps data between steps",
		Category:    NodeCategoryAction,
		ConfigSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"expression": map[string]interface{}{"type": "string"},
				"mapping":    map[string]interface{}{"type": "object"},
			},
		},
		ExampleConfig: map[string]interface{}{
			"mapping": map[string]interface{}{
				"user_name":  "${steps.trigger.body.user.name}",
				"user_email": "${steps.trigger.body.user.email}",
			},
		},
		LLMDescription: "Use this to reshape, filter, or combine data. Create mappings to extract and rename fields " +
			"from previous steps. Use ${steps.nodeName.output.field} syntax for variable references.",
		IsActive: true,
	})

	registry.Register(NodeTemplate{
		NodeType:    "action:code",
		Name:        "JavaScript Code",
		Description: "Executes custom JavaScript code",
		Category:    NodeCategoryAction,
		ConfigSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"script":  map[string]interface{}{"type": "string"},
				"timeout": map[string]interface{}{"type": "integer"},
			},
		},
		ExampleConfig: map[string]interface{}{
			"script":  "const input = context.steps.transform.output;\nreturn { processed: input.data.map(x => x * 2) };",
			"timeout": 30,
		},
		LLMDescription: "Use this for complex logic that cannot be expressed with other nodes. " +
			"Write JavaScript code that returns an object. Access previous steps via context.steps.nodeName.output.",
		IsActive: true,
	})

	registry.Register(NodeTemplate{
		NodeType:    "action:formula",
		Name:        "Formula",
		Description: "Evaluates mathematical or logical expressions",
		Category:    NodeCategoryAction,
		ConfigSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"expression":      map[string]interface{}{"type": "string"},
				"output_variable": map[string]interface{}{"type": "string"},
			},
		},
		ExampleConfig: map[string]interface{}{
			"expression":      "${steps.data.output.price} * ${steps.data.output.quantity} * 1.1",
			"output_variable": "total_with_tax",
		},
		LLMDescription: "Use this for calculations. Write expressions using ${variable} syntax. " +
			"Supports math operators (+, -, *, /), comparisons (==, !=, <, >), and logical operators (&&, ||, !).",
		IsActive: true,
	})

	registry.Register(NodeTemplate{
		NodeType:    "action:email",
		Name:        "Send Email",
		Description: "Sends email notifications",
		Category:    NodeCategoryAction,
		ConfigSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"to":      map[string]interface{}{"type": "string"},
				"subject": map[string]interface{}{"type": "string"},
				"body":    map[string]interface{}{"type": "string"},
				"cc":      map[string]interface{}{"type": "string"},
				"bcc":     map[string]interface{}{"type": "string"},
			},
		},
		ExampleConfig: map[string]interface{}{
			"to":      "user@example.com",
			"subject": "Workflow Alert",
			"body":    "Data processed: ${steps.process.output.count} items",
		},
		LLMDescription: "Use this to send email notifications. Configure recipients, subject, and body. " +
			"Use template variables like ${steps.nodeName.output.field} for dynamic content.",
		IsActive: true,
	})

	// Control flow
	registry.Register(NodeTemplate{
		NodeType:    "control:if",
		Name:        "Conditional (If/Else)",
		Description: "Branches workflow based on a condition",
		Category:    NodeCategoryControl,
		ConfigSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"condition":   map[string]interface{}{"type": "string"},
				"description": map[string]interface{}{"type": "string"},
			},
		},
		ExampleConfig: map[string]interface{}{
			"condition":   "${steps.data.output.status} == \"approved\"",
			"description": "Check if status is approved",
		},
		LLMDescription: "Use this to branch workflow logic based on conditions. Write a condition that evaluates to true/false. " +
			"Connect 'true' and 'false' edges to different downstream nodes for conditional branching.",
		IsActive: true,
	})

	registry.Register(NodeTemplate{
		NodeType:    "control:loop",
		Name:        "Loop (For Each)",
		Description: "Iterates over an array executing child nodes for each item",
		Category:    NodeCategoryControl,
		ConfigSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"source":         map[string]interface{}{"type": "string"},
				"item_variable":  map[string]interface{}{"type": "string"},
				"index_variable": map[string]interface{}{"type": "string"},
				"max_iterations": map[string]interface{}{"type": "integer"},
			},
		},
		ExampleConfig: map[string]interface{}{
			"source":         "${steps.data.output.items}",
			"item_variable":  "item",
			"index_variable": "index",
			"max_iterations": 100,
		},
		LLMDescription: "Use this to process each item in an array. Specify the source array and variable names. " +
			"Nodes inside the loop can access ${loop.item} for the current item and ${loop.index} for the index.",
		IsActive: true,
	})

	registry.Register(NodeTemplate{
		NodeType:    "control:delay",
		Name:        "Delay",
		Description: "Pauses workflow execution for a specified duration",
		Category:    NodeCategoryControl,
		ConfigSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"duration": map[string]interface{}{"type": "string"},
			},
		},
		ExampleConfig: map[string]interface{}{
			"duration": "5s",
		},
		LLMDescription: "Use this to pause the workflow. Specify duration as '5s' (seconds), '2m' (minutes), or '1h' (hours). " +
			"Can also use variables: '${steps.config.output.delay}'.",
		IsActive: true,
	})

	registry.Register(NodeTemplate{
		NodeType:    "control:parallel",
		Name:        "Parallel",
		Description: "Executes multiple branches in parallel",
		Category:    NodeCategoryControl,
		ConfigSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"error_strategy":  map[string]interface{}{"type": "string", "enum": []string{"fail_fast", "wait_all"}},
				"max_concurrency": map[string]interface{}{"type": "integer"},
			},
		},
		ExampleConfig: map[string]interface{}{
			"error_strategy":  "fail_fast",
			"max_concurrency": 5,
		},
		LLMDescription: "Use this to run multiple branches simultaneously. Choose 'fail_fast' to stop on first error " +
			"or 'wait_all' to complete all branches regardless of errors.",
		IsActive: true,
	})

	// Integrations
	registry.Register(NodeTemplate{
		NodeType:    "slack:send_message",
		Name:        "Slack: Send Message",
		Description: "Sends a message to a Slack channel",
		Category:    NodeCategoryIntegration,
		ConfigSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"channel": map[string]interface{}{"type": "string"},
				"text":    map[string]interface{}{"type": "string"},
				"blocks":  map[string]interface{}{"type": "array"},
			},
		},
		ExampleConfig: map[string]interface{}{
			"channel": "#general",
			"text":    "New alert: ${steps.data.output.message}",
		},
		LLMDescription: "Use this to send messages to Slack channels. Specify channel name (e.g., #general) or ID " +
			"and message text. Supports Block Kit for rich formatting.",
		IsActive: true,
	})

	registry.Register(NodeTemplate{
		NodeType:    "slack:send_dm",
		Name:        "Slack: Send Direct Message",
		Description: "Sends a direct message to a Slack user",
		Category:    NodeCategoryIntegration,
		ConfigSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"user_id": map[string]interface{}{"type": "string"},
				"text":    map[string]interface{}{"type": "string"},
			},
		},
		ExampleConfig: map[string]interface{}{
			"user_id": "${steps.lookup.output.slack_user_id}",
			"text":    "You have a new task assigned",
		},
		LLMDescription: "Use this to send direct messages to Slack users. Specify user ID (not username) and message text.",
		IsActive: true,
	})

	return registry
}
