package workflow

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestDryRun_ValidWorkflow tests dry-run with a valid workflow
func TestDryRun_ValidWorkflow(t *testing.T) {
	service, mockRepo := newTestService()
	ctx := context.Background()
	tenantID := "tenant-123"
	workflowID := "workflow-123"

	definition := WorkflowDefinition{
		Nodes: []Node{
			{
				ID:   "trigger-1",
				Type: string(NodeTypeTriggerWebhook),
				Data: NodeData{
					Name: "Webhook Trigger",
					Config: json.RawMessage(`{
						"auth_type": "signature"
					}`),
				},
			},
			{
				ID:   "http-1",
				Type: string(NodeTypeActionHTTP),
				Data: NodeData{
					Name: "HTTP Request",
					Config: json.RawMessage(`{
						"method": "POST",
						"url": "https://api.example.com/data",
						"body": {"data": "${trigger.payload}"}
					}`),
				},
			},
			{
				ID:   "transform-1",
				Type: string(NodeTypeActionTransform),
				Data: NodeData{
					Name: "Transform Data",
					Config: json.RawMessage(`{
						"mapping": {
							"result": "${steps.http-1.body.result}"
						}
					}`),
				},
			},
		},
		Edges: []Edge{
			{ID: "e1", Source: "trigger-1", Target: "http-1"},
			{ID: "e2", Source: "http-1", Target: "transform-1"},
		},
	}

	definitionJSON, _ := json.Marshal(definition)
	workflow := &Workflow{
		ID:         workflowID,
		TenantID:   tenantID,
		Name:       "Test Workflow",
		Status:     string(WorkflowStatusActive),
		Definition: definitionJSON,
		Version:    1,
	}

	mockRepo.On("GetByID", ctx, tenantID, workflowID).Return(workflow, nil)

	testData := map[string]interface{}{
		"payload": map[string]interface{}{
			"id":   "123",
			"name": "test",
		},
	}

	result, err := service.DryRun(ctx, tenantID, workflowID, testData)

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, result.Valid)
	assert.Len(t, result.ExecutionOrder, 3)
	assert.Equal(t, []string{"trigger-1", "http-1", "transform-1"}, result.ExecutionOrder)
	assert.Len(t, result.Errors, 0)
	assert.NotNil(t, result.VariableMapping)
	assert.Contains(t, result.VariableMapping, "trigger.payload")
	assert.Contains(t, result.VariableMapping, "steps.http-1")
	mockRepo.AssertExpectations(t)
}

// TestDryRun_InvalidNodeConfiguration tests dry-run with invalid node config
func TestDryRun_InvalidNodeConfiguration(t *testing.T) {
	service, mockRepo := newTestService()
	ctx := context.Background()
	tenantID := "tenant-123"
	workflowID := "workflow-123"

	definition := WorkflowDefinition{
		Nodes: []Node{
			{
				ID:   "trigger-1",
				Type: string(NodeTypeTriggerWebhook),
				Data: NodeData{
					Name:   "Webhook Trigger",
					Config: json.RawMessage(`{}`),
				},
			},
			{
				ID:   "http-1",
				Type: string(NodeTypeActionHTTP),
				Data: NodeData{
					Name: "HTTP Request",
					Config: json.RawMessage(`{
						"method": "",
						"url": ""
					}`),
				},
			},
		},
		Edges: []Edge{
			{ID: "e1", Source: "trigger-1", Target: "http-1"},
		},
	}

	definitionJSON, _ := json.Marshal(definition)
	workflow := &Workflow{
		ID:         workflowID,
		TenantID:   tenantID,
		Name:       "Invalid Workflow",
		Status:     string(WorkflowStatusActive),
		Definition: definitionJSON,
		Version:    1,
	}

	mockRepo.On("GetByID", ctx, tenantID, workflowID).Return(workflow, nil)

	result, err := service.DryRun(ctx, tenantID, workflowID, nil)

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.False(t, result.Valid)
	assert.Greater(t, len(result.Errors), 0)
	assert.Equal(t, "http-1", result.Errors[0].NodeID)
	mockRepo.AssertExpectations(t)
}

// TestDryRun_MissingVariableReference tests dry-run with undefined variable reference
func TestDryRun_MissingVariableReference(t *testing.T) {
	service, mockRepo := newTestService()
	ctx := context.Background()
	tenantID := "tenant-123"
	workflowID := "workflow-123"

	definition := WorkflowDefinition{
		Nodes: []Node{
			{
				ID:   "trigger-1",
				Type: string(NodeTypeTriggerWebhook),
				Data: NodeData{
					Name:   "Webhook Trigger",
					Config: json.RawMessage(`{}`),
				},
			},
			{
				ID:   "transform-1",
				Type: string(NodeTypeActionTransform),
				Data: NodeData{
					Name: "Transform Data",
					Config: json.RawMessage(`{
						"mapping": {
							"result": "${steps.nonexistent.data}"
						}
					}`),
				},
			},
		},
		Edges: []Edge{
			{ID: "e1", Source: "trigger-1", Target: "transform-1"},
		},
	}

	definitionJSON, _ := json.Marshal(definition)
	workflow := &Workflow{
		ID:         workflowID,
		TenantID:   tenantID,
		Name:       "Workflow with Missing Ref",
		Status:     string(WorkflowStatusActive),
		Definition: definitionJSON,
		Version:    1,
	}

	mockRepo.On("GetByID", ctx, tenantID, workflowID).Return(workflow, nil)

	result, err := service.DryRun(ctx, tenantID, workflowID, nil)

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.False(t, result.Valid)
	assert.Greater(t, len(result.Errors), 0)
	foundError := false
	for _, e := range result.Errors {
		if e.NodeID == "transform-1" && e.Field == "mapping" {
			assert.Contains(t, e.Message, "nonexistent")
			foundError = true
		}
	}
	assert.True(t, foundError, "Expected error about missing variable reference")
	mockRepo.AssertExpectations(t)
}

// TestDryRun_CircularDependency tests dry-run with circular dependency
func TestDryRun_CircularDependency(t *testing.T) {
	service, mockRepo := newTestService()
	ctx := context.Background()
	tenantID := "tenant-123"
	workflowID := "workflow-123"

	definition := WorkflowDefinition{
		Nodes: []Node{
			{
				ID:   "node-1",
				Type: string(NodeTypeActionTransform),
				Data: NodeData{
					Name:   "Node 1",
					Config: json.RawMessage(`{}`),
				},
			},
			{
				ID:   "node-2",
				Type: string(NodeTypeActionTransform),
				Data: NodeData{
					Name:   "Node 2",
					Config: json.RawMessage(`{}`),
				},
			},
			{
				ID:   "node-3",
				Type: string(NodeTypeActionTransform),
				Data: NodeData{
					Name:   "Node 3",
					Config: json.RawMessage(`{}`),
				},
			},
		},
		Edges: []Edge{
			{ID: "e1", Source: "node-1", Target: "node-2"},
			{ID: "e2", Source: "node-2", Target: "node-3"},
			{ID: "e3", Source: "node-3", Target: "node-1"}, // Creates cycle
		},
	}

	definitionJSON, _ := json.Marshal(definition)
	workflow := &Workflow{
		ID:         workflowID,
		TenantID:   tenantID,
		Name:       "Circular Workflow",
		Status:     string(WorkflowStatusActive),
		Definition: definitionJSON,
		Version:    1,
	}

	mockRepo.On("GetByID", ctx, tenantID, workflowID).Return(workflow, nil)

	result, err := service.DryRun(ctx, tenantID, workflowID, nil)

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.False(t, result.Valid)
	assert.Greater(t, len(result.Errors), 0)
	foundCycleError := false
	for _, e := range result.Errors {
		if e.Field == "edges" {
			assert.Contains(t, e.Message, "cycle")
			foundCycleError = true
		}
	}
	assert.True(t, foundCycleError, "Expected error about circular dependency")
	mockRepo.AssertExpectations(t)
}

// TestDryRun_MissingCredential tests dry-run with missing credential reference
func TestDryRun_MissingCredential(t *testing.T) {
	service, mockRepo := newTestService()
	ctx := context.Background()
	tenantID := "tenant-123"
	workflowID := "workflow-123"

	definition := WorkflowDefinition{
		Nodes: []Node{
			{
				ID:   "trigger-1",
				Type: string(NodeTypeTriggerWebhook),
				Data: NodeData{
					Name:   "Webhook Trigger",
					Config: json.RawMessage(`{}`),
				},
			},
			{
				ID:   "http-1",
				Type: string(NodeTypeActionHTTP),
				Data: NodeData{
					Name: "HTTP Request with Credential",
					Config: json.RawMessage(`{
						"method": "POST",
						"url": "https://api.example.com/data",
						"credential_id": "cred-nonexistent"
					}`),
				},
			},
		},
		Edges: []Edge{
			{ID: "e1", Source: "trigger-1", Target: "http-1"},
		},
	}

	definitionJSON, _ := json.Marshal(definition)
	workflow := &Workflow{
		ID:         workflowID,
		TenantID:   tenantID,
		Name:       "Workflow with Missing Credential",
		Status:     string(WorkflowStatusActive),
		Definition: definitionJSON,
		Version:    1,
	}

	mockRepo.On("GetByID", ctx, tenantID, workflowID).Return(workflow, nil)

	result, err := service.DryRun(ctx, tenantID, workflowID, nil)

	require.NoError(t, err)
	assert.NotNil(t, result)
	// Should have a warning about credential reference
	foundWarning := false
	for _, w := range result.Warnings {
		if w.NodeID == "http-1" && w.Message != "" {
			assert.Contains(t, w.Message, "credential")
			foundWarning = true
		}
	}
	assert.True(t, foundWarning, "Expected warning about credential reference")
	mockRepo.AssertExpectations(t)
}

// TestDryRun_WorkflowNotFound tests dry-run with non-existent workflow
func TestDryRun_WorkflowNotFound(t *testing.T) {
	service, mockRepo := newTestService()
	ctx := context.Background()
	tenantID := "tenant-123"
	workflowID := "nonexistent"

	mockRepo.On("GetByID", ctx, tenantID, workflowID).Return(nil, ErrNotFound)

	result, err := service.DryRun(ctx, tenantID, workflowID, nil)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, ErrNotFound, err)
	mockRepo.AssertExpectations(t)
}

// TestDryRun_InvalidWorkflowDefinition tests dry-run with corrupt definition JSON
func TestDryRun_InvalidWorkflowDefinition(t *testing.T) {
	service, mockRepo := newTestService()
	ctx := context.Background()
	tenantID := "tenant-123"
	workflowID := "workflow-123"

	workflow := &Workflow{
		ID:         workflowID,
		TenantID:   tenantID,
		Name:       "Invalid Workflow",
		Status:     string(WorkflowStatusActive),
		Definition: json.RawMessage(`{invalid json`),
		Version:    1,
	}

	mockRepo.On("GetByID", ctx, tenantID, workflowID).Return(workflow, nil)

	result, err := service.DryRun(ctx, tenantID, workflowID, nil)

	assert.Error(t, err)
	assert.Nil(t, result)
	mockRepo.AssertExpectations(t)
}

// TestDryRun_ConditionalNode tests dry-run with conditional branching
func TestDryRun_ConditionalNode(t *testing.T) {
	service, mockRepo := newTestService()
	ctx := context.Background()
	tenantID := "tenant-123"
	workflowID := "workflow-123"

	definition := WorkflowDefinition{
		Nodes: []Node{
			{
				ID:   "trigger-1",
				Type: string(NodeTypeTriggerWebhook),
				Data: NodeData{
					Name:   "Webhook Trigger",
					Config: json.RawMessage(`{}`),
				},
			},
			{
				ID:   "if-1",
				Type: string(NodeTypeControlIf),
				Data: NodeData{
					Name: "Check Condition",
					Config: json.RawMessage(`{
						"condition": "${trigger.status} == 'success'"
					}`),
				},
			},
			{
				ID:   "success-1",
				Type: string(NodeTypeActionTransform),
				Data: NodeData{
					Name: "Success Path",
					Config: json.RawMessage(`{
						"mapping": {"message": "Success"}
					}`),
				},
			},
			{
				ID:   "failure-1",
				Type: string(NodeTypeActionTransform),
				Data: NodeData{
					Name: "Failure Path",
					Config: json.RawMessage(`{
						"mapping": {"message": "Failure"}
					}`),
				},
			},
		},
		Edges: []Edge{
			{ID: "e1", Source: "trigger-1", Target: "if-1"},
			{ID: "e2", Source: "if-1", Target: "success-1", Label: "true"},
			{ID: "e3", Source: "if-1", Target: "failure-1", Label: "false"},
		},
	}

	definitionJSON, _ := json.Marshal(definition)
	workflow := &Workflow{
		ID:         workflowID,
		TenantID:   tenantID,
		Name:       "Conditional Workflow",
		Status:     string(WorkflowStatusActive),
		Definition: definitionJSON,
		Version:    1,
	}

	mockRepo.On("GetByID", ctx, tenantID, workflowID).Return(workflow, nil)

	result, err := service.DryRun(ctx, tenantID, workflowID, nil)

	require.NoError(t, err)
	assert.NotNil(t, result)
	// Both branches should be included in the execution order
	assert.Contains(t, result.ExecutionOrder, "if-1")
	assert.Contains(t, result.ExecutionOrder, "success-1")
	assert.Contains(t, result.ExecutionOrder, "failure-1")
	mockRepo.AssertExpectations(t)
}

// TestDryRun_LoopNode tests dry-run with loop node
func TestDryRun_LoopNode(t *testing.T) {
	service, mockRepo := newTestService()
	ctx := context.Background()
	tenantID := "tenant-123"
	workflowID := "workflow-123"

	definition := WorkflowDefinition{
		Nodes: []Node{
			{
				ID:   "trigger-1",
				Type: string(NodeTypeTriggerWebhook),
				Data: NodeData{
					Name:   "Webhook Trigger",
					Config: json.RawMessage(`{}`),
				},
			},
			{
				ID:   "loop-1",
				Type: string(NodeTypeControlLoop),
				Data: NodeData{
					Name: "Process Items",
					Config: json.RawMessage(`{
						"source": "${trigger.items}",
						"item_variable": "item",
						"max_iterations": 100
					}`),
				},
			},
			{
				ID:   "transform-1",
				Type: string(NodeTypeActionTransform),
				Data: NodeData{
					Name: "Transform Item",
					Config: json.RawMessage(`{
						"mapping": {
							"processed": "${item.value}"
						}
					}`),
				},
			},
		},
		Edges: []Edge{
			{ID: "e1", Source: "trigger-1", Target: "loop-1"},
			{ID: "e2", Source: "loop-1", Target: "transform-1"},
		},
	}

	definitionJSON, _ := json.Marshal(definition)
	workflow := &Workflow{
		ID:         workflowID,
		TenantID:   tenantID,
		Name:       "Loop Workflow",
		Status:     string(WorkflowStatusActive),
		Definition: definitionJSON,
		Version:    1,
	}

	mockRepo.On("GetByID", ctx, tenantID, workflowID).Return(workflow, nil)

	testData := map[string]interface{}{
		"items": []interface{}{
			map[string]interface{}{"value": 1},
			map[string]interface{}{"value": 2},
		},
	}

	result, err := service.DryRun(ctx, tenantID, workflowID, testData)

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Contains(t, result.ExecutionOrder, "loop-1")
	assert.Contains(t, result.VariableMapping, "item")
	mockRepo.AssertExpectations(t)
}
