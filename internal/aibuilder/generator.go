package aibuilder

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/gorax/gorax/internal/llm"
)

// GeneratorConfig holds configuration for the workflow generator
type GeneratorConfig struct {
	Model       string  `json:"model"`
	MaxTokens   int     `json:"max_tokens"`
	Temperature float64 `json:"temperature"`
}

// DefaultGeneratorConfig returns default generator configuration
func DefaultGeneratorConfig() *GeneratorConfig {
	return &GeneratorConfig{
		Model:       "gpt-4",
		MaxTokens:   4096,
		Temperature: 0.7,
	}
}

// WorkflowGenerator generates workflows from natural language using LLM
type WorkflowGenerator struct {
	provider llm.Provider
	registry *NodeRegistry
	config   *GeneratorConfig
}

// NewWorkflowGenerator creates a new workflow generator
func NewWorkflowGenerator(provider llm.Provider, registry *NodeRegistry, config *GeneratorConfig) *WorkflowGenerator {
	if config == nil {
		config = DefaultGeneratorConfig()
	}
	return &WorkflowGenerator{
		provider: provider,
		registry: registry,
		config:   config,
	}
}

// Generate generates a workflow from a build request
func (g *WorkflowGenerator) Generate(ctx context.Context, request *BuildRequest, history []ConversationMessage) (*GeneratedWorkflow, string, error) {
	prompt := g.buildGeneratePrompt(request)
	messages := g.buildMessages(prompt, history)

	temp := g.config.Temperature
	llmReq := &llm.ChatRequest{
		Model:       g.config.Model,
		Messages:    messages,
		MaxTokens:   g.config.MaxTokens,
		Temperature: &temp,
		ResponseFormat: &llm.ResponseFormat{
			Type: llm.ResponseFormatJSON,
		},
	}

	resp, err := g.provider.ChatCompletion(ctx, llmReq)
	if err != nil {
		return nil, "", fmt.Errorf("LLM error: %w", err)
	}

	workflow, explanation, err := g.parseResponse(resp.Message.Content)
	if err != nil {
		return nil, "", fmt.Errorf("failed to parse response: %w", err)
	}

	if err := g.validateWorkflow(workflow); err != nil {
		return nil, "", fmt.Errorf("invalid workflow: %w", err)
	}

	g.assignPositions(workflow)

	return workflow, explanation, nil
}

// Refine refines an existing workflow based on user feedback
func (g *WorkflowGenerator) Refine(ctx context.Context, workflow *GeneratedWorkflow, feedback string, history []ConversationMessage) (*GeneratedWorkflow, string, error) {
	if workflow == nil {
		return nil, "", errors.New("workflow is required")
	}
	if feedback == "" {
		return nil, "", errors.New("feedback is required")
	}

	prompt := g.buildRefinePrompt(workflow, feedback)
	messages := g.buildMessages(prompt, history)

	temp := g.config.Temperature
	llmReq := &llm.ChatRequest{
		Model:       g.config.Model,
		Messages:    messages,
		MaxTokens:   g.config.MaxTokens,
		Temperature: &temp,
		ResponseFormat: &llm.ResponseFormat{
			Type: llm.ResponseFormatJSON,
		},
	}

	resp, err := g.provider.ChatCompletion(ctx, llmReq)
	if err != nil {
		return nil, "", fmt.Errorf("LLM error: %w", err)
	}

	refined, explanation, err := g.parseResponse(resp.Message.Content)
	if err != nil {
		return nil, "", fmt.Errorf("failed to parse response: %w", err)
	}

	if err := g.validateWorkflow(refined); err != nil {
		return nil, "", fmt.Errorf("invalid workflow: %w", err)
	}

	g.assignPositions(refined)

	return refined, explanation, nil
}

// buildGeneratePrompt creates the prompt for workflow generation
func (g *WorkflowGenerator) buildGeneratePrompt(request *BuildRequest) string {
	var sb strings.Builder

	sb.WriteString("Create a workflow based on the following description:\n\n")
	sb.WriteString(request.Description)
	sb.WriteString("\n\n")

	if request.Context != nil {
		if len(request.Context.AvailableCredentials) > 0 {
			sb.WriteString("Available credentials: ")
			sb.WriteString(strings.Join(request.Context.AvailableCredentials, ", "))
			sb.WriteString("\n")
		}
		if len(request.Context.AvailableIntegrations) > 0 {
			sb.WriteString("Available integrations: ")
			sb.WriteString(strings.Join(request.Context.AvailableIntegrations, ", "))
			sb.WriteString("\n")
		}
	}

	if request.Constraints != nil {
		if request.Constraints.MaxNodes > 0 {
			sb.WriteString(fmt.Sprintf("Maximum nodes allowed: %d\n", request.Constraints.MaxNodes))
		}
		if len(request.Constraints.AllowedTypes) > 0 {
			sb.WriteString("Allowed node types: ")
			sb.WriteString(strings.Join(request.Constraints.AllowedTypes, ", "))
			sb.WriteString("\n")
		}
	}

	return sb.String()
}

// buildRefinePrompt creates the prompt for workflow refinement
func (g *WorkflowGenerator) buildRefinePrompt(workflow *GeneratedWorkflow, feedback string) string {
	workflowJSON, _ := json.MarshalIndent(workflow, "", "  ") //nolint:errcheck // known struct type

	return fmt.Sprintf(`Refine the following workflow based on user feedback:

Current workflow:
%s

User feedback:
%s

Please update the workflow according to the feedback and return the refined workflow.`, string(workflowJSON), feedback)
}

// buildMessages constructs the LLM messages array
func (g *WorkflowGenerator) buildMessages(prompt string, history []ConversationMessage) []llm.ChatMessage {
	messages := []llm.ChatMessage{
		llm.SystemMessage(g.buildSystemPrompt()),
	}

	// Add conversation history
	for _, msg := range history {
		role := string(msg.Role)
		messages = append(messages, llm.NewChatMessage(role, msg.Content))
	}

	// Add the current user message
	messages = append(messages, llm.UserMessage(prompt))

	return messages
}

// buildSystemPrompt creates the system prompt with node registry context
func (g *WorkflowGenerator) buildSystemPrompt() string {
	nodeContext := g.registry.BuildLLMContext()

	return fmt.Sprintf(`You are an AI workflow builder assistant. Your task is to create workflow definitions based on user descriptions.

%s

Response Format:
You must respond with valid JSON in this exact format:
{
  "workflow": {
    "name": "Workflow Name",
    "description": "Brief description of what the workflow does",
    "definition": {
      "nodes": [
        {
          "id": "unique-node-id",
          "type": "node:type",
          "name": "Human readable name",
          "description": "Optional description",
          "config": { ... node configuration ... }
        }
      ],
      "edges": [
        {
          "id": "unique-edge-id",
          "source": "source-node-id",
          "target": "target-node-id",
          "label": "optional label for conditional edges"
        }
      ]
    }
  },
  "explanation": "Brief explanation of how the workflow works"
}

Guidelines:
1. Always start with a trigger node (trigger:webhook or trigger:schedule)
2. Use meaningful, descriptive node IDs (e.g., "trigger-1", "send-slack-message")
3. Connect nodes with edges in logical execution order
4. Use ${steps.nodeId.output.field} syntax for referencing data from previous steps
5. For conditional nodes (control:if), create edges with labels "true" and "false"
6. Keep workflows simple and focused on the user's requirements
7. Include appropriate error handling where needed`, nodeContext)
}

// parseResponse parses the LLM response into a workflow
func (g *WorkflowGenerator) parseResponse(content string) (*GeneratedWorkflow, string, error) {
	// Extract JSON from response (handles code blocks)
	jsonStr := extractJSONFromResponse(content)

	var response struct {
		Workflow    *GeneratedWorkflow `json:"workflow"`
		Explanation string             `json:"explanation"`
	}

	if err := json.Unmarshal([]byte(jsonStr), &response); err != nil {
		return nil, "", fmt.Errorf("failed to parse JSON: %w", err)
	}

	if response.Workflow == nil {
		return nil, "", errors.New("response missing workflow field")
	}

	return response.Workflow, response.Explanation, nil
}

// validateWorkflow validates the generated workflow structure
func (g *WorkflowGenerator) validateWorkflow(workflow *GeneratedWorkflow) error {
	if err := workflow.Validate(); err != nil {
		return err
	}

	// Validate node types against registry
	for _, node := range workflow.Definition.Nodes {
		if !g.registry.IsValidType(node.Type) {
			return fmt.Errorf("unknown node type: %s", node.Type)
		}
	}

	return nil
}

// assignPositions assigns canvas positions to nodes for visualization
func (g *WorkflowGenerator) assignPositions(workflow *GeneratedWorkflow) {
	if workflow == nil || workflow.Definition == nil {
		return
	}

	const (
		startX   = 250.0
		startY   = 50.0
		spacingX = 300.0 // Horizontal spacing between parallel branches
		spacingY = 150.0 // Vertical spacing between levels
	)

	// Build graph structure
	nodeIndex := make(map[string]int)
	for i, node := range workflow.Definition.Nodes {
		nodeIndex[node.ID] = i
	}

	// Build adjacency list (forward edges)
	adjList := make(map[string][]string)
	inDegree := make(map[string]int)

	for _, node := range workflow.Definition.Nodes {
		inDegree[node.ID] = 0
	}

	for _, edge := range workflow.Definition.Edges {
		adjList[edge.Source] = append(adjList[edge.Source], edge.Target)
		inDegree[edge.Target]++
	}

	// Perform topological sort to determine levels (BFS-based)
	levels := make(map[string]int)
	queue := []string{}

	// Find root nodes (in-degree = 0)
	for nodeID := range inDegree {
		if inDegree[nodeID] == 0 {
			levels[nodeID] = 0
			queue = append(queue, nodeID)
		}
	}

	// BFS to assign levels
	for len(queue) > 0 {
		currentID := queue[0]
		queue = queue[1:]

		currentLevel := levels[currentID]

		for _, neighborID := range adjList[currentID] {
			// Assign level only if not assigned or needs update
			if existingLevel, exists := levels[neighborID]; !exists || currentLevel+1 > existingLevel {
				levels[neighborID] = currentLevel + 1
			}

			inDegree[neighborID]--
			if inDegree[neighborID] == 0 {
				queue = append(queue, neighborID)
			}
		}
	}

	// Nodes that aren't connected (isolated nodes) get level 0
	for _, node := range workflow.Definition.Nodes {
		if _, exists := levels[node.ID]; !exists {
			levels[node.ID] = 0
		}
	}

	// Group nodes by level
	levelNodes := make(map[int][]string)
	maxLevel := 0
	for nodeID, level := range levels {
		levelNodes[level] = append(levelNodes[level], nodeID)
		if level > maxLevel {
			maxLevel = level
		}
	}

	// Position nodes: center each level horizontally
	for level := 0; level <= maxLevel; level++ {
		nodes := levelNodes[level]
		nodesCount := len(nodes)

		// Calculate starting X to center the nodes at this level
		totalWidth := float64(nodesCount-1) * spacingX
		levelStartX := startX - (totalWidth / 2)

		for i, nodeID := range nodes {
			idx := nodeIndex[nodeID]
			workflow.Definition.Nodes[idx].Position = &NodePosition{
				X: levelStartX + float64(i)*spacingX,
				Y: startY + float64(level)*spacingY,
			}
		}
	}
}

// extractJSONFromResponse extracts JSON from an LLM response that may contain markdown
func extractJSONFromResponse(content string) string {
	content = strings.TrimSpace(content)

	// Try to extract from markdown code block
	codeBlockRegex := regexp.MustCompile("```(?:json)?\\s*([\\s\\S]*?)```")
	matches := codeBlockRegex.FindStringSubmatch(content)
	if len(matches) > 1 {
		return strings.TrimSpace(matches[1])
	}

	// Try to find JSON object in the content
	start := strings.Index(content, "{")
	end := strings.LastIndex(content, "}")
	if start >= 0 && end > start {
		return content[start : end+1]
	}

	return content
}
