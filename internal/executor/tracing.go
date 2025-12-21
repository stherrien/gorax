package executor

import (
	"context"

	"github.com/gorax/gorax/internal/tracing"
	"github.com/gorax/gorax/internal/workflow"
)

// executeWithTracing wraps workflow execution with distributed tracing
func (e *Executor) executeWithTracing(ctx context.Context, execution *workflow.Execution) error {
	return tracing.TraceWorkflowExecution(
		ctx,
		execution.TenantID,
		execution.WorkflowID,
		execution.ID,
		func(ctx context.Context) error {
			return e.Execute(ctx, execution)
		},
	)
}

// executeNodeWithTracing wraps node execution with distributed tracing
func (e *Executor) executeNodeWithTracing(ctx context.Context, node workflow.Node, execCtx *ExecutionContext) (interface{}, error) {
	return tracing.TraceStepExecution(
		ctx,
		execCtx.TenantID,
		execCtx.WorkflowID,
		execCtx.ExecutionID,
		node.ID,
		node.Type,
		func(ctx context.Context) (interface{}, error) {
			return e.executeNodeWithTracking(ctx, node, execCtx)
		},
	)
}
