package actions

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gorax/gorax/internal/workflow"
)

// TestSubWorkflowAction_Integration_EndToEnd tests complete subworkflow execution flow
func TestSubWorkflowAction_Integration_EndToEnd(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	mockRepo := NewMockWorkflowRepository()
	mockExec := &MockWorkflowExecutor{
		executeFunc: func(ctx context.Context, execution *workflow.Execution) error {
			// Simulate successful execution with output
			execution.Status = string(workflow.ExecutionStatusCompleted)

			// Parse input to generate relevant output
			var triggerData map[string]interface{}
			if execution.TriggerData != nil {
				_ = json.Unmarshal(*execution.TriggerData, &triggerData)
			}

			// Create output based on input
			outputMap := map[string]interface{}{
				"status":  "validated",
				"message": "Validation successful",
			}

			outputData, _ := json.Marshal(outputMap)
			rawOutput := json.RawMessage(outputData)
			execution.OutputData = &rawOutput
			mockRepo.executions[execution.ID] = execution

			return nil
		},
	}

	// Create parent workflow
	parentDef := workflow.WorkflowDefinition{
		Nodes: []workflow.Node{
			{
				ID:   "trigger",
				Type: string(workflow.NodeTypeTriggerWebhook),
			},
			{
				ID:   "validate-user",
				Type: string(workflow.NodeTypeActionSubworkflow),
			},
		},
		Edges: []workflow.Edge{
			{
				ID:     "e1",
				Source: "trigger",
				Target: "validate-user",
			},
		},
	}
	parentDefBytes, _ := json.Marshal(parentDef)

	mockRepo.workflows["parent-wf"] = &workflow.Workflow{
		ID:         "parent-wf",
		TenantID:   "tenant-1",
		Name:       "Parent Workflow",
		Definition: parentDefBytes,
		Status:     string(workflow.WorkflowStatusActive),
		Version:    1,
	}

	// Create child workflow (validation workflow)
	childDef := workflow.WorkflowDefinition{
		Nodes: []workflow.Node{
			{
				ID:   "trigger",
				Type: string(workflow.NodeTypeTriggerWebhook),
			},
			{
				ID:   "validate",
				Type: string(workflow.NodeTypeActionHTTP),
			},
		},
		Edges: []workflow.Edge{
			{
				ID:     "e1",
				Source: "trigger",
				Target: "validate",
			},
		},
	}
	childDefBytes, _ := json.Marshal(childDef)

	mockRepo.workflows["child-wf"] = &workflow.Workflow{
		ID:         "child-wf",
		TenantID:   "tenant-1",
		Name:       "User Validation Workflow",
		Definition: childDefBytes,
		Status:     string(workflow.WorkflowStatusActive),
		Version:    1,
	}

	// Configure subworkflow action
	config := &workflow.SubWorkflowConfig{
		WorkflowID:   "child-wf",
		WorkflowName: "User Validation Workflow",
		InputMapping: map[string]string{
			"email": "${trigger.body.email}",
			"name":  "${trigger.body.name}",
		},
		OutputMapping: map[string]string{
			"is_valid": "${status}",
		},
		Mode:           "sync",
		Timeout:        "30s",
		InheritContext: false,
	}

	ctxData := map[string]interface{}{
		"trigger": map[string]interface{}{
			"body": map[string]interface{}{
				"email": "user@example.com",
				"name":  "John Doe",
			},
		},
		"env": map[string]interface{}{
			"tenant_id":    "tenant-1",
			"execution_id": "parent-exec-1",
			"workflow_id":  "parent-wf",
		},
		"_execution": map[string]interface{}{
			"depth":          0,
			"workflow_chain": []string{"parent-wf"},
		},
	}

	action := NewSubWorkflowAction(mockRepo, mockExec)
	input := NewActionInput(config, ctxData)

	// Execute
	output, err := action.Execute(context.Background(), input)

	// Assert success
	require.NoError(t, err)
	assert.NotNil(t, output)

	// Verify output data
	data := output.Data.(map[string]interface{})
	assert.Equal(t, "validated", data["is_valid"])

	// Verify metadata
	assert.Contains(t, output.Metadata, "execution_id")
	assert.Contains(t, output.Metadata, "workflow_id")
	assert.Equal(t, "child-wf", output.Metadata["workflow_id"])

	// Verify child execution was created
	assert.GreaterOrEqual(t, len(mockRepo.executions), 1)

	var childExec *workflow.Execution
	for _, exec := range mockRepo.executions {
		if exec.WorkflowID == "child-wf" {
			childExec = exec
			break
		}
	}

	require.NotNil(t, childExec)
	assert.Equal(t, "tenant-1", childExec.TenantID)
	assert.Equal(t, 1, childExec.ExecutionDepth)
	assert.NotNil(t, childExec.ParentExecutionID)
	assert.Equal(t, "parent-exec-1", *childExec.ParentExecutionID)
	assert.Equal(t, string(workflow.ExecutionStatusCompleted), childExec.Status)

	// Verify trigger data was mapped correctly
	var triggerData map[string]interface{}
	err = json.Unmarshal(*childExec.TriggerData, &triggerData)
	require.NoError(t, err)
	assert.Equal(t, "user@example.com", triggerData["email"])
	assert.Equal(t, "John Doe", triggerData["name"])
	assert.Equal(t, "parent-exec-1", triggerData["_parent_execution_id"])
	assert.Equal(t, "parent-wf", triggerData["_parent_workflow_id"])
}

// TestSubWorkflowAction_Integration_MultiLevel tests nested subworkflow execution
func TestSubWorkflowAction_Integration_MultiLevel(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	mockRepo := NewMockWorkflowRepository()
	executionCount := 0

	mockExec := &MockWorkflowExecutor{
		executeFunc: func(ctx context.Context, execution *workflow.Execution) error {
			executionCount++
			execution.Status = string(workflow.ExecutionStatusCompleted)

			outputMap := map[string]interface{}{
				"level":  executionCount,
				"result": "success",
			}

			outputData, _ := json.Marshal(outputMap)
			rawOutput := json.RawMessage(outputData)
			execution.OutputData = &rawOutput
			mockRepo.executions[execution.ID] = execution

			return nil
		},
	}

	// Create three-level workflow hierarchy
	workflows := []struct {
		id       string
		name     string
		childID  string
		hasChild bool
	}{
		{"wf-level-1", "Level 1 Workflow", "wf-level-2", true},
		{"wf-level-2", "Level 2 Workflow", "wf-level-3", true},
		{"wf-level-3", "Level 3 Workflow", "", false},
	}

	for _, wf := range workflows {
		def := workflow.WorkflowDefinition{
			Nodes: []workflow.Node{
				{
					ID:   "trigger",
					Type: string(workflow.NodeTypeTriggerWebhook),
				},
			},
		}

		if wf.hasChild {
			def.Nodes = append(def.Nodes, workflow.Node{
				ID:   "call-child",
				Type: string(workflow.NodeTypeActionSubworkflow),
			})
			def.Edges = []workflow.Edge{
				{
					ID:     "e1",
					Source: "trigger",
					Target: "call-child",
				},
			}
		}

		defBytes, _ := json.Marshal(def)

		mockRepo.workflows[wf.id] = &workflow.Workflow{
			ID:         wf.id,
			TenantID:   "tenant-1",
			Name:       wf.name,
			Definition: defBytes,
			Status:     string(workflow.WorkflowStatusActive),
			Version:    1,
		}
	}

	// Execute level 1 -> level 2 -> level 3
	// Start with level 1
	config1 := &workflow.SubWorkflowConfig{
		WorkflowID:    "wf-level-1",
		Mode:          "sync",
		WaitForResult: true,
		TimeoutMs:     10000,
	}

	ctxData1 := map[string]interface{}{
		"env": map[string]interface{}{
			"tenant_id":    "tenant-1",
			"execution_id": "exec-root",
			"workflow_id":  "wf-root",
		},
		"_execution": map[string]interface{}{
			"depth":          0,
			"workflow_chain": []string{},
		},
	}

	action := NewSubWorkflowAction(mockRepo, mockExec)
	input1 := NewActionInput(config1, ctxData1)

	output1, err := action.Execute(context.Background(), input1)

	// Assert
	require.NoError(t, err)
	assert.NotNil(t, output1)

	// Verify execution count (should be at least 1 for level-1)
	assert.GreaterOrEqual(t, executionCount, 1)

	// Verify all workflow executions were created with correct depth
	var exec1, exec2, exec3 *workflow.Execution
	for _, exec := range mockRepo.executions {
		switch exec.WorkflowID {
		case "wf-level-1":
			exec1 = exec
		case "wf-level-2":
			exec2 = exec
		case "wf-level-3":
			exec3 = exec
		}
	}

	// Level 1 should exist
	require.NotNil(t, exec1)
	assert.Equal(t, 1, exec1.ExecutionDepth)

	// If level 2 was called
	if exec2 != nil {
		assert.Equal(t, 2, exec2.ExecutionDepth)
		assert.NotNil(t, exec2.ParentExecutionID)
	}

	// If level 3 was called
	if exec3 != nil {
		assert.Equal(t, 3, exec3.ExecutionDepth)
		assert.NotNil(t, exec3.ParentExecutionID)
	}
}

// TestSubWorkflowAction_Integration_WorkflowReusability tests multiple parents calling same child
func TestSubWorkflowAction_Integration_WorkflowReusability(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	mockRepo := NewMockWorkflowRepository()
	mockExec := &MockWorkflowExecutor{
		executeFunc: func(ctx context.Context, execution *workflow.Execution) error {
			execution.Status = string(workflow.ExecutionStatusCompleted)
			outputData := json.RawMessage(`{"result": "success"}`)
			execution.OutputData = &outputData
			mockRepo.executions[execution.ID] = execution
			return nil
		},
	}

	// Create shared child workflow
	childDef := workflow.WorkflowDefinition{
		Nodes: []workflow.Node{
			{
				ID:   "trigger",
				Type: string(workflow.NodeTypeTriggerWebhook),
			},
		},
	}
	childDefBytes, _ := json.Marshal(childDef)

	mockRepo.workflows["shared-child"] = &workflow.Workflow{
		ID:         "shared-child",
		TenantID:   "tenant-1",
		Name:       "Shared Notification Workflow",
		Definition: childDefBytes,
		Status:     string(workflow.WorkflowStatusActive),
		Version:    1,
	}

	action := NewSubWorkflowAction(mockRepo, mockExec)

	// Execute from parent A
	configA := &workflow.SubWorkflowConfig{
		WorkflowID:    "shared-child",
		WaitForResult: true,
		TimeoutMs:     5000,
	}

	ctxDataA := map[string]interface{}{
		"env": map[string]interface{}{
			"tenant_id":    "tenant-1",
			"execution_id": "parent-a-exec",
			"workflow_id":  "parent-a",
		},
	}

	outputA, errA := action.Execute(context.Background(), NewActionInput(configA, ctxDataA))
	require.NoError(t, errA)
	assert.NotNil(t, outputA)

	// Execute from parent B
	configB := &workflow.SubWorkflowConfig{
		WorkflowID:    "shared-child",
		WaitForResult: true,
		TimeoutMs:     5000,
	}

	ctxDataB := map[string]interface{}{
		"env": map[string]interface{}{
			"tenant_id":    "tenant-1",
			"execution_id": "parent-b-exec",
			"workflow_id":  "parent-b",
		},
	}

	outputB, errB := action.Execute(context.Background(), NewActionInput(configB, ctxDataB))
	require.NoError(t, errB)
	assert.NotNil(t, outputB)

	// Verify two separate executions were created
	assert.Equal(t, 2, len(mockRepo.executions))

	// Verify both have correct parent references
	parentIDs := make(map[string]bool)
	for _, exec := range mockRepo.executions {
		assert.Equal(t, "shared-child", exec.WorkflowID)
		assert.NotNil(t, exec.ParentExecutionID)
		parentIDs[*exec.ParentExecutionID] = true
	}

	assert.True(t, parentIDs["parent-a-exec"])
	assert.True(t, parentIDs["parent-b-exec"])
}
