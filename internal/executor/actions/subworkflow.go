package actions

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/gorax/gorax/internal/workflow"
)

// WorkflowRepository defines the interface for workflow operations needed by SubWorkflowAction
type WorkflowRepository interface {
	GetByID(ctx context.Context, tenantID, workflowID string) (*workflow.Workflow, error)
	CreateExecution(ctx context.Context, execution *workflow.Execution) error
	GetExecutionByID(ctx context.Context, tenantID, executionID string) (*workflow.Execution, error)
}

// WorkflowExecutor defines the interface for executing workflows
type WorkflowExecutor interface {
	Execute(ctx context.Context, execution *workflow.Execution) error
}

// SubWorkflowAction executes another workflow as a sub-workflow
type SubWorkflowAction struct {
	repo     WorkflowRepository
	executor WorkflowExecutor
}

// NewSubWorkflowAction creates a new sub-workflow action
func NewSubWorkflowAction(repo WorkflowRepository, executor WorkflowExecutor) *SubWorkflowAction {
	return &SubWorkflowAction{
		repo:     repo,
		executor: executor,
	}
}

// Execute runs the sub-workflow action
func (a *SubWorkflowAction) Execute(ctx context.Context, input *ActionInput) (*ActionOutput, error) {
	// Parse configuration
	config, err := a.parseConfig(input.Config)
	if err != nil {
		return nil, fmt.Errorf("invalid sub-workflow config: %w", err)
	}

	// Extract execution context from input
	execContext := a.extractExecutionContext(input.Context)

	// Check recursion depth
	if execContext.Depth >= 10 { // MaxSubWorkflowDepth
		return nil, fmt.Errorf("max depth exceeded: current depth %d", execContext.Depth)
	}

	// Check for circular dependencies
	if a.containsWorkflow(execContext.WorkflowChain, config.WorkflowID) {
		return nil, fmt.Errorf("circular workflow dependency detected: %s already in chain %v",
			config.WorkflowID, execContext.WorkflowChain)
	}

	// Load target workflow
	targetWorkflow, err := a.repo.GetByID(ctx, execContext.TenantID, config.WorkflowID)
	if err != nil {
		return nil, fmt.Errorf("failed to load workflow %s: %w", config.WorkflowID, err)
	}

	// Verify workflow is active
	if targetWorkflow.Status != string(workflow.WorkflowStatusActive) {
		return nil, fmt.Errorf("workflow %s is not active (status: %s)", config.WorkflowID, targetWorkflow.Status)
	}

	// Map inputs from parent context to sub-workflow trigger data
	triggerData := a.mapInputs(config.InputMapping, input.Context)

	// Add execution metadata
	triggerData["_parent_execution_id"] = execContext.ExecutionID
	triggerData["_parent_workflow_id"] = execContext.WorkflowID
	triggerData["_depth"] = execContext.Depth + 1

	triggerDataJSON, err := json.Marshal(triggerData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal trigger data: %w", err)
	}

	// Create sub-workflow execution
	executionID := uuid.New().String()
	parentExecID := execContext.ExecutionID
	execution := &workflow.Execution{
		ID:                executionID,
		TenantID:          execContext.TenantID,
		WorkflowID:        config.WorkflowID,
		WorkflowVersion:   targetWorkflow.Version,
		Status:            string(workflow.ExecutionStatusPending),
		TriggerType:       "sub_workflow",
		TriggerData:       (*json.RawMessage)(&triggerDataJSON),
		ParentExecutionID: &parentExecID,
		ExecutionDepth:    execContext.Depth + 1,
		CreatedAt:         time.Now(),
	}

	if err := a.repo.CreateExecution(ctx, execution); err != nil {
		return nil, fmt.Errorf("failed to create sub-workflow execution: %w", err)
	}

	// Execute based on sync/async mode
	if config.WaitForResult {
		// Synchronous execution - wait for completion
		return a.executeSynchronous(ctx, execution, config, input.Context)
	}

	// Asynchronous execution - fire and forget
	return a.executeAsynchronous(ctx, execution)
}

// executeSynchronous executes the sub-workflow and waits for completion
func (a *SubWorkflowAction) executeSynchronous(ctx context.Context, execution *workflow.Execution,
	config *workflow.SubWorkflowConfig, parentContext map[string]interface{}) (*ActionOutput, error) {

	if a.executor == nil {
		return nil, fmt.Errorf("workflow executor not configured for synchronous execution")
	}

	// Create context with timeout if specified
	execCtx := ctx
	if config.TimeoutMs > 0 {
		var cancel context.CancelFunc
		execCtx, cancel = context.WithTimeout(ctx, time.Duration(config.TimeoutMs)*time.Millisecond)
		defer cancel()
	}

	// Execute the workflow
	errChan := make(chan error, 1)
	go func() {
		errChan <- a.executor.Execute(execCtx, execution)
	}()

	// Wait for completion or timeout
	select {
	case err := <-errChan:
		if err != nil {
			return nil, fmt.Errorf("sub-workflow execution failed: %w", err)
		}

		// Load completed execution to get output
		completedExec, err := a.repo.GetExecutionByID(ctx, execution.TenantID, execution.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to load completed execution: %w", err)
		}

		// Parse output data
		var outputData map[string]interface{}
		if completedExec.OutputData != nil {
			if err := json.Unmarshal(*completedExec.OutputData, &outputData); err != nil {
				return nil, fmt.Errorf("failed to parse output data: %w", err)
			}
		}

		// Map outputs back to parent context
		mappedOutput := a.mapOutputs(config.OutputMapping, outputData)

		return NewActionOutput(mappedOutput).
			WithMetadata("execution_id", execution.ID).
			WithMetadata("workflow_id", execution.WorkflowID).
			WithMetadata("status", completedExec.Status), nil

	case <-execCtx.Done():
		return nil, fmt.Errorf("sub-workflow execution timeout after %dms", config.TimeoutMs)
	}
}

// executeAsynchronous starts the sub-workflow execution without waiting
func (a *SubWorkflowAction) executeAsynchronous(ctx context.Context, execution *workflow.Execution) (*ActionOutput, error) {
	// Start execution asynchronously
	if a.executor != nil {
		go func() {
			// Use background context to avoid cancellation
			_ = a.executor.Execute(context.Background(), execution)
		}()
	}

	// Return immediately with execution ID
	return NewActionOutput(map[string]interface{}{
		"execution_id": execution.ID,
		"workflow_id":  execution.WorkflowID,
		"status":       "started",
	}).WithMetadata("async", true), nil
}

// mapInputs maps parent context values to sub-workflow inputs
func (a *SubWorkflowAction) mapInputs(mapping map[string]string, context map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})

	for key, expr := range mapping {
		value := a.evaluateExpression(expr, context)
		if value != nil {
			result[key] = value
		}
	}

	return result
}

// mapOutputs maps sub-workflow outputs to parent context
func (a *SubWorkflowAction) mapOutputs(mapping map[string]string, subOutput map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})

	for key, expr := range mapping {
		// For output mapping, we use subOutput as the context
		value := a.evaluateExpression(expr, subOutput)
		if value != nil {
			result[key] = value
		}
	}

	return result
}

// evaluateExpression evaluates a simple expression like ${trigger.data} or ${output.result}
func (a *SubWorkflowAction) evaluateExpression(expr string, context map[string]interface{}) interface{} {
	// Simple variable interpolation: ${path.to.value}
	if !strings.HasPrefix(expr, "${") || !strings.HasSuffix(expr, "}") {
		return expr // Return literal value
	}

	// Extract path
	path := strings.TrimSuffix(strings.TrimPrefix(expr, "${"), "}")
	parts := strings.Split(path, ".")

	// Navigate the context
	var current interface{} = context
	for _, part := range parts {
		if m, ok := current.(map[string]interface{}); ok {
			current = m[part]
		} else {
			return nil
		}
	}

	return current
}

// parseConfig parses the sub-workflow configuration
func (a *SubWorkflowAction) parseConfig(config interface{}) (*workflow.SubWorkflowConfig, error) {
	// Handle both direct struct and map types
	switch v := config.(type) {
	case *workflow.SubWorkflowConfig:
		return v, nil
	case workflow.SubWorkflowConfig:
		return &v, nil
	case map[string]interface{}:
		// Convert map to struct
		data, err := json.Marshal(v)
		if err != nil {
			return nil, err
		}
		var cfg workflow.SubWorkflowConfig
		if err := json.Unmarshal(data, &cfg); err != nil {
			return nil, err
		}
		return &cfg, nil
	default:
		return nil, fmt.Errorf("unsupported config type: %T", config)
	}
}

// extractExecutionContext extracts execution context from the input context
func (a *SubWorkflowAction) extractExecutionContext(context map[string]interface{}) *executionContextInfo {
	info := &executionContextInfo{
		Depth:         0,
		WorkflowChain: []string{},
	}

	// Check for execution metadata
	if execData, ok := context["_execution"].(map[string]interface{}); ok {
		if depth, ok := execData["depth"].(int); ok {
			info.Depth = depth
		}
		if chain, ok := execData["workflow_chain"].([]string); ok {
			info.WorkflowChain = chain
		}
	}

	// Extract tenant and execution IDs from env
	if envData, ok := context["env"].(map[string]interface{}); ok {
		if tenantID, ok := envData["tenant_id"].(string); ok {
			info.TenantID = tenantID
		}
		if executionID, ok := envData["execution_id"].(string); ok {
			info.ExecutionID = executionID
		}
		if workflowID, ok := envData["workflow_id"].(string); ok {
			info.WorkflowID = workflowID
		}
	}

	return info
}

// containsWorkflow checks if a workflow ID is in the chain
func (a *SubWorkflowAction) containsWorkflow(chain []string, workflowID string) bool {
	for _, id := range chain {
		if id == workflowID {
			return true
		}
	}
	return false
}

// executionContextInfo holds extracted execution context information
type executionContextInfo struct {
	TenantID      string
	ExecutionID   string
	WorkflowID    string
	Depth         int
	WorkflowChain []string
}
