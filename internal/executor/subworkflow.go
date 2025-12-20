package executor

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/gorax/gorax/internal/executor/actions"
	"github.com/gorax/gorax/internal/workflow"
)

// workflowRepositoryAdapter adapts the workflow.Repository to the actions.WorkflowRepository interface
type workflowRepositoryAdapter struct {
	repo *workflow.Repository
}

func (a *workflowRepositoryAdapter) GetByID(ctx context.Context, tenantID, workflowID string) (*workflow.Workflow, error) {
	return a.repo.GetByID(ctx, tenantID, workflowID)
}

func (a *workflowRepositoryAdapter) CreateExecution(ctx context.Context, execution *workflow.Execution) error {
	var triggerData []byte
	if execution.TriggerData != nil {
		triggerData = []byte(*execution.TriggerData)
	}

	// Create execution using the repository's method
	created, err := a.repo.CreateExecution(ctx, execution.TenantID, execution.WorkflowID,
		execution.WorkflowVersion, execution.TriggerType, triggerData)
	if err != nil {
		return err
	}

	// Update the execution with the created ID
	execution.ID = created.ID
	execution.CreatedAt = created.CreatedAt
	return nil
}

func (a *workflowRepositoryAdapter) GetExecutionByID(ctx context.Context, tenantID, executionID string) (*workflow.Execution, error) {
	return a.repo.GetExecutionByID(ctx, tenantID, executionID)
}

// executeSubWorkflowAction executes a sub-workflow action
func (e *Executor) executeSubWorkflowAction(ctx context.Context, node workflow.Node, execCtx *ExecutionContext) (interface{}, error) {
	// Check recursion depth
	if execCtx.Depth >= MaxSubWorkflowDepth {
		return nil, fmt.Errorf("max sub-workflow depth exceeded: %d", execCtx.Depth)
	}

	// Parse sub-workflow configuration
	var config workflow.SubWorkflowConfig
	if err := json.Unmarshal(node.Data.Config, &config); err != nil {
		return nil, fmt.Errorf("invalid sub-workflow config: %w", err)
	}

	// Check for circular dependencies
	for _, wfID := range execCtx.WorkflowChain {
		if wfID == config.WorkflowID {
			return nil, fmt.Errorf("circular workflow dependency detected: %s", config.WorkflowID)
		}
	}

	// Build context for sub-workflow action
	actionContext := buildInputData(execCtx)

	// Add execution metadata for recursion protection
	actionContext["_execution"] = map[string]interface{}{
		"depth":          execCtx.Depth,
		"workflow_chain": execCtx.WorkflowChain,
	}

	// Create adapter for the repository
	repoAdapter := &workflowRepositoryAdapter{repo: e.repo}

	// Create sub-workflow action
	subWorkflowAction := actions.NewSubWorkflowAction(repoAdapter, e)
	actionInput := actions.NewActionInput(&config, actionContext)

	// Execute sub-workflow
	output, err := subWorkflowAction.Execute(ctx, actionInput)
	if err != nil {
		return nil, fmt.Errorf("sub-workflow execution failed: %w", err)
	}

	return output.Data, nil
}
