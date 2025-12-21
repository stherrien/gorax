package aibuilder

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/gorax/gorax/internal/llm"
)

// MockLLMProvider is a mock implementation of llm.Provider
type MockLLMProvider struct {
	mock.Mock
}

func (m *MockLLMProvider) ChatCompletion(ctx context.Context, req *llm.ChatRequest) (*llm.ChatResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*llm.ChatResponse), args.Error(1)
}

func (m *MockLLMProvider) GenerateEmbeddings(ctx context.Context, req *llm.EmbeddingRequest) (*llm.EmbeddingResponse, error) {
	return nil, nil
}

func (m *MockLLMProvider) CountTokens(text string, model string) (int, error) {
	return 0, nil
}

func (m *MockLLMProvider) ListModels(ctx context.Context) ([]llm.Model, error) {
	return nil, nil
}

func (m *MockLLMProvider) Name() string {
	return "mock"
}

func (m *MockLLMProvider) HealthCheck(ctx context.Context) error {
	return nil
}

func TestNewWorkflowGenerator(t *testing.T) {
	provider := &MockLLMProvider{}
	registry := DefaultNodeRegistry()

	gen := NewWorkflowGenerator(provider, registry, nil)

	assert.NotNil(t, gen)
	assert.Equal(t, provider, gen.provider)
	assert.Equal(t, registry, gen.registry)
}

func TestWorkflowGenerator_Generate(t *testing.T) {
	t.Run("successful generation", func(t *testing.T) {
		provider := &MockLLMProvider{}
		registry := DefaultNodeRegistry()
		gen := NewWorkflowGenerator(provider, registry, nil)

		// Mock LLM response with a valid workflow
		llmResponse := &llm.ChatResponse{
			Message: llm.ChatMessage{
				Role: "assistant",
				Content: `{
					"workflow": {
						"name": "Webhook to Slack",
						"description": "Sends a Slack message when a webhook is received",
						"definition": {
							"nodes": [
								{
									"id": "trigger-1",
									"type": "trigger:webhook",
									"name": "Webhook Trigger",
									"config": {"path": "/webhook"}
								},
								{
									"id": "action-1",
									"type": "slack:send_message",
									"name": "Send Slack Message",
									"config": {"channel": "#alerts", "text": "${steps.trigger-1.body.message}"}
								}
							],
							"edges": [
								{
									"id": "edge-1",
									"source": "trigger-1",
									"target": "action-1"
								}
							]
						}
					},
					"explanation": "This workflow listens for webhook requests and sends the message to Slack."
				}`,
			},
			Usage: llm.TokenUsage{
				PromptTokens:     100,
				CompletionTokens: 200,
			},
		}

		provider.On("ChatCompletion", mock.Anything, mock.Anything).Return(llmResponse, nil)

		request := &BuildRequest{
			Description: "Create a workflow that sends a Slack message when a webhook is received",
		}

		workflow, explanation, err := gen.Generate(context.Background(), request, nil)

		require.NoError(t, err)
		assert.NotNil(t, workflow)
		assert.Equal(t, "Webhook to Slack", workflow.Name)
		assert.NotEmpty(t, explanation)
		assert.Len(t, workflow.Definition.Nodes, 2)
		assert.Len(t, workflow.Definition.Edges, 1)

		provider.AssertExpectations(t)
	})

	t.Run("LLM error", func(t *testing.T) {
		provider := &MockLLMProvider{}
		registry := DefaultNodeRegistry()
		gen := NewWorkflowGenerator(provider, registry, nil)

		provider.On("ChatCompletion", mock.Anything, mock.Anything).
			Return(nil, errors.New("LLM unavailable"))

		request := &BuildRequest{
			Description: "Create a workflow",
		}

		workflow, explanation, err := gen.Generate(context.Background(), request, nil)

		require.Error(t, err)
		assert.Nil(t, workflow)
		assert.Empty(t, explanation)
		assert.Contains(t, err.Error(), "LLM unavailable")

		provider.AssertExpectations(t)
	})

	t.Run("invalid JSON response", func(t *testing.T) {
		provider := &MockLLMProvider{}
		registry := DefaultNodeRegistry()
		gen := NewWorkflowGenerator(provider, registry, nil)

		llmResponse := &llm.ChatResponse{
			Message: llm.ChatMessage{
				Role:    "assistant",
				Content: "This is not valid JSON",
			},
		}

		provider.On("ChatCompletion", mock.Anything, mock.Anything).Return(llmResponse, nil)

		request := &BuildRequest{
			Description: "Create a workflow",
		}

		workflow, explanation, err := gen.Generate(context.Background(), request, nil)

		require.Error(t, err)
		assert.Nil(t, workflow)
		assert.Empty(t, explanation)
		assert.Contains(t, err.Error(), "parse")

		provider.AssertExpectations(t)
	})

	t.Run("invalid workflow structure", func(t *testing.T) {
		provider := &MockLLMProvider{}
		registry := DefaultNodeRegistry()
		gen := NewWorkflowGenerator(provider, registry, nil)

		// Response missing required fields
		llmResponse := &llm.ChatResponse{
			Message: llm.ChatMessage{
				Role: "assistant",
				Content: `{
					"workflow": {
						"name": "",
						"definition": {
							"nodes": []
						}
					}
				}`,
			},
		}

		provider.On("ChatCompletion", mock.Anything, mock.Anything).Return(llmResponse, nil)

		request := &BuildRequest{
			Description: "Create a workflow",
		}

		workflow, explanation, err := gen.Generate(context.Background(), request, nil)

		require.Error(t, err)
		assert.Nil(t, workflow)
		assert.Empty(t, explanation)

		provider.AssertExpectations(t)
	})

	t.Run("with conversation history", func(t *testing.T) {
		provider := &MockLLMProvider{}
		registry := DefaultNodeRegistry()
		gen := NewWorkflowGenerator(provider, registry, nil)

		llmResponse := &llm.ChatResponse{
			Message: llm.ChatMessage{
				Role: "assistant",
				Content: `{
					"workflow": {
						"name": "Updated Workflow",
						"definition": {
							"nodes": [
								{
									"id": "trigger-1",
									"type": "trigger:webhook",
									"name": "Webhook Trigger"
								}
							]
						}
					},
					"explanation": "Updated based on your feedback."
				}`,
			},
		}

		provider.On("ChatCompletion", mock.Anything, mock.MatchedBy(func(req *llm.ChatRequest) bool {
			// Should include history messages
			return len(req.Messages) > 2 // system + history + user
		})).Return(llmResponse, nil)

		request := &BuildRequest{
			Description: "Add error handling",
		}

		history := []ConversationMessage{
			{Role: MessageRoleUser, Content: "Create a webhook workflow"},
			{Role: MessageRoleAssistant, Content: "Here's a webhook workflow..."},
		}

		workflow, _, err := gen.Generate(context.Background(), request, history)

		require.NoError(t, err)
		assert.NotNil(t, workflow)

		provider.AssertExpectations(t)
	})
}

func TestWorkflowGenerator_Refine(t *testing.T) {
	t.Run("successful refinement", func(t *testing.T) {
		provider := &MockLLMProvider{}
		registry := DefaultNodeRegistry()
		gen := NewWorkflowGenerator(provider, registry, nil)

		existingWorkflow := &GeneratedWorkflow{
			Name:        "Original Workflow",
			Description: "Original description",
			Definition: &WorkflowDefinition{
				Nodes: []GeneratedNode{
					{ID: "trigger-1", Type: "trigger:webhook", Name: "Webhook Trigger"},
				},
				Edges: []GeneratedEdge{},
			},
		}

		llmResponse := &llm.ChatResponse{
			Message: llm.ChatMessage{
				Role: "assistant",
				Content: `{
					"workflow": {
						"name": "Refined Workflow",
						"description": "Added error handling",
						"definition": {
							"nodes": [
								{
									"id": "trigger-1",
									"type": "trigger:webhook",
									"name": "Webhook Trigger"
								},
								{
									"id": "action-1",
									"type": "action:http",
									"name": "API Call"
								}
							],
							"edges": [
								{
									"id": "edge-1",
									"source": "trigger-1",
									"target": "action-1"
								}
							]
						}
					},
					"explanation": "Added an HTTP action for API calls."
				}`,
			},
		}

		provider.On("ChatCompletion", mock.Anything, mock.Anything).Return(llmResponse, nil)

		refined, explanation, err := gen.Refine(context.Background(), existingWorkflow, "Add an HTTP action", nil)

		require.NoError(t, err)
		assert.NotNil(t, refined)
		assert.Equal(t, "Refined Workflow", refined.Name)
		assert.Len(t, refined.Definition.Nodes, 2)
		assert.Contains(t, explanation, "HTTP action")

		provider.AssertExpectations(t)
	})

	t.Run("nil workflow", func(t *testing.T) {
		provider := &MockLLMProvider{}
		registry := DefaultNodeRegistry()
		gen := NewWorkflowGenerator(provider, registry, nil)

		refined, explanation, err := gen.Refine(context.Background(), nil, "Add error handling", nil)

		require.Error(t, err)
		assert.Nil(t, refined)
		assert.Empty(t, explanation)
		assert.Contains(t, err.Error(), "workflow is required")
	})

	t.Run("empty feedback", func(t *testing.T) {
		provider := &MockLLMProvider{}
		registry := DefaultNodeRegistry()
		gen := NewWorkflowGenerator(provider, registry, nil)

		existingWorkflow := &GeneratedWorkflow{
			Name: "Test",
			Definition: &WorkflowDefinition{
				Nodes: []GeneratedNode{
					{ID: "n1", Type: "trigger:webhook", Name: "Trigger"},
				},
			},
		}

		refined, explanation, err := gen.Refine(context.Background(), existingWorkflow, "", nil)

		require.Error(t, err)
		assert.Nil(t, refined)
		assert.Empty(t, explanation)
		assert.Contains(t, err.Error(), "feedback is required")
	})
}

func TestWorkflowGenerator_BuildPrompt(t *testing.T) {
	provider := &MockLLMProvider{}
	registry := DefaultNodeRegistry()
	gen := NewWorkflowGenerator(provider, registry, nil)

	t.Run("basic prompt", func(t *testing.T) {
		request := &BuildRequest{
			Description: "Create a webhook to Slack workflow",
		}

		prompt := gen.buildGeneratePrompt(request)

		assert.Contains(t, prompt, "Create a webhook to Slack workflow")
		assert.Contains(t, prompt, "workflow")
	})

	t.Run("prompt with context", func(t *testing.T) {
		request := &BuildRequest{
			Description: "Create a workflow that uses Slack",
			Context: &BuildContext{
				AvailableCredentials:  []string{"slack", "email"},
				AvailableIntegrations: []string{"slack", "email"},
			},
		}

		prompt := gen.buildGeneratePrompt(request)

		assert.Contains(t, prompt, "Create a workflow that uses Slack")
		assert.Contains(t, prompt, "slack")
	})

	t.Run("prompt with constraints", func(t *testing.T) {
		request := &BuildRequest{
			Description: "Create a simple workflow",
			Constraints: &BuildConstraints{
				MaxNodes:     5,
				AllowedTypes: []string{"trigger:webhook", "action:http"},
			},
		}

		prompt := gen.buildGeneratePrompt(request)

		assert.Contains(t, prompt, "Create a simple workflow")
	})
}

func TestWorkflowGenerator_ParseResponse(t *testing.T) {
	provider := &MockLLMProvider{}
	registry := DefaultNodeRegistry()
	gen := NewWorkflowGenerator(provider, registry, nil)

	t.Run("valid response", func(t *testing.T) {
		response := `{
			"workflow": {
				"name": "Test Workflow",
				"description": "A test",
				"definition": {
					"nodes": [
						{"id": "n1", "type": "trigger:webhook", "name": "Trigger"}
					],
					"edges": []
				}
			},
			"explanation": "Here's your workflow"
		}`

		workflow, explanation, err := gen.parseResponse(response)

		require.NoError(t, err)
		assert.Equal(t, "Test Workflow", workflow.Name)
		assert.Equal(t, "Here's your workflow", explanation)
	})

	t.Run("response with markdown code block", func(t *testing.T) {
		response := "```json\n" + `{
			"workflow": {
				"name": "Test Workflow",
				"definition": {
					"nodes": [
						{"id": "n1", "type": "trigger:webhook", "name": "Trigger"}
					]
				}
			},
			"explanation": "Here's your workflow"
		}` + "\n```"

		workflow, explanation, err := gen.parseResponse(response)

		require.NoError(t, err)
		assert.Equal(t, "Test Workflow", workflow.Name)
		assert.NotEmpty(t, explanation)
	})

	t.Run("invalid JSON", func(t *testing.T) {
		response := "This is not JSON"

		workflow, explanation, err := gen.parseResponse(response)

		require.Error(t, err)
		assert.Nil(t, workflow)
		assert.Empty(t, explanation)
	})

	t.Run("missing workflow field", func(t *testing.T) {
		response := `{"explanation": "No workflow here"}`

		workflow, explanation, err := gen.parseResponse(response)

		require.Error(t, err)
		assert.Nil(t, workflow)
		assert.Empty(t, explanation)
	})
}

func TestWorkflowGenerator_ValidateGeneratedWorkflow(t *testing.T) {
	provider := &MockLLMProvider{}
	registry := DefaultNodeRegistry()
	gen := NewWorkflowGenerator(provider, registry, nil)

	t.Run("valid workflow", func(t *testing.T) {
		workflow := &GeneratedWorkflow{
			Name: "Test",
			Definition: &WorkflowDefinition{
				Nodes: []GeneratedNode{
					{ID: "n1", Type: "trigger:webhook", Name: "Trigger"},
					{ID: "n2", Type: "action:http", Name: "HTTP"},
				},
				Edges: []GeneratedEdge{
					{ID: "e1", Source: "n1", Target: "n2"},
				},
			},
		}

		err := gen.validateWorkflow(workflow)
		require.NoError(t, err)
	})

	t.Run("unknown node type", func(t *testing.T) {
		workflow := &GeneratedWorkflow{
			Name: "Test",
			Definition: &WorkflowDefinition{
				Nodes: []GeneratedNode{
					{ID: "n1", Type: "trigger:unknown", Name: "Unknown"},
				},
			},
		}

		err := gen.validateWorkflow(workflow)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "unknown node type")
	})
}

func TestGeneratorConfig_Defaults(t *testing.T) {
	config := DefaultGeneratorConfig()

	assert.Equal(t, "gpt-4", config.Model)
	assert.Greater(t, config.MaxTokens, 0)
	assert.GreaterOrEqual(t, config.Temperature, 0.0)
	assert.LessOrEqual(t, config.Temperature, 1.0)
}

func TestWorkflowGenerator_AssignPositions(t *testing.T) {
	provider := &MockLLMProvider{}
	registry := DefaultNodeRegistry()
	gen := NewWorkflowGenerator(provider, registry, nil)

	workflow := &GeneratedWorkflow{
		Name: "Test",
		Definition: &WorkflowDefinition{
			Nodes: []GeneratedNode{
				{ID: "n1", Type: "trigger:webhook", Name: "Trigger"},
				{ID: "n2", Type: "action:http", Name: "HTTP"},
				{ID: "n3", Type: "slack:send_message", Name: "Slack"},
			},
			Edges: []GeneratedEdge{
				{ID: "e1", Source: "n1", Target: "n2"},
				{ID: "e2", Source: "n2", Target: "n3"},
			},
		},
	}

	gen.assignPositions(workflow)

	// All nodes should have positions
	for _, node := range workflow.Definition.Nodes {
		assert.NotNil(t, node.Position, "Node %s should have position", node.ID)
	}

	// First node should be at starting Y position (50)
	assert.Equal(t, float64(50), workflow.Definition.Nodes[0].Position.Y)

	// Second node should be below first
	assert.Greater(t, workflow.Definition.Nodes[1].Position.Y, workflow.Definition.Nodes[0].Position.Y)

	// Third node should be below second
	assert.Greater(t, workflow.Definition.Nodes[2].Position.Y, workflow.Definition.Nodes[1].Position.Y)
}

func TestExtractJSONFromResponse(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "plain JSON",
			input:    `{"key": "value"}`,
			expected: `{"key": "value"}`,
		},
		{
			name:     "JSON in code block",
			input:    "```json\n{\"key\": \"value\"}\n```",
			expected: `{"key": "value"}`,
		},
		{
			name:     "JSON in plain code block",
			input:    "```\n{\"key\": \"value\"}\n```",
			expected: `{"key": "value"}`,
		},
		{
			name:     "JSON with surrounding text",
			input:    "Here's the workflow:\n{\"key\": \"value\"}\nLet me know if you need changes.",
			expected: `{"key": "value"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractJSONFromResponse(tt.input)
			// Parse both to compare as JSON
			var expected, actual interface{}
			json.Unmarshal([]byte(tt.expected), &expected)
			json.Unmarshal([]byte(result), &actual)
			assert.Equal(t, expected, actual)
		})
	}
}
