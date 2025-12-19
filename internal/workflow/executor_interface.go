package workflow

import "context"

// ExecutorInterface defines the interface for workflow execution
type ExecutorInterface interface {
	Execute(ctx context.Context, execution *Execution) error
}
