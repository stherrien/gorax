// Package nodes provides workflow node implementations with a unified interface
package nodes

import (
	"context"
	"encoding/json"

	"github.com/gorax/gorax/internal/workflow"
)

// Node represents a workflow node that can be executed
type Node interface {
	// GetType returns the node type identifier
	GetType() string

	// Validate validates the node configuration
	Validate() error

	// Execute runs the node with the given context and input
	Execute(ctx context.Context, input *NodeInput) (*NodeOutput, error)
}

// NodeInput represents the input data for node execution
type NodeInput struct {
	// Config contains the node-specific configuration
	Config json.RawMessage

	// Context contains execution context data (trigger data, step outputs, env)
	Context map[string]interface{}

	// ExecutionContext contains execution metadata
	ExecutionContext *ExecutionContext
}

// ExecutionContext holds execution metadata for a node
type ExecutionContext struct {
	TenantID          string
	ExecutionID       string
	WorkflowID        string
	Depth             int
	WorkflowChain     []string
	ParentExecutionID string
}

// NodeOutput represents the result of a node execution
type NodeOutput struct {
	// Data contains the output data from the node
	Data interface{}

	// Metadata contains additional information about the execution
	Metadata map[string]interface{}
}

// NewNodeInput creates a new NodeInput
func NewNodeInput(config json.RawMessage, context map[string]interface{}, execCtx *ExecutionContext) *NodeInput {
	if context == nil {
		context = make(map[string]interface{})
	}
	return &NodeInput{
		Config:           config,
		Context:          context,
		ExecutionContext: execCtx,
	}
}

// NewNodeOutput creates a new NodeOutput
func NewNodeOutput(data interface{}) *NodeOutput {
	return &NodeOutput{
		Data:     data,
		Metadata: make(map[string]interface{}),
	}
}

// WithMetadata adds metadata to the output
func (o *NodeOutput) WithMetadata(key string, value interface{}) *NodeOutput {
	o.Metadata[key] = value
	return o
}

// WorkflowRepository defines the interface for workflow operations needed by nodes
type WorkflowRepository interface {
	GetByID(ctx context.Context, tenantID, workflowID string) (*workflow.Workflow, error)
	GetByVersion(ctx context.Context, tenantID, workflowID string, version int) (*workflow.Workflow, error)
	CreateExecution(ctx context.Context, execution *workflow.Execution) error
	GetExecutionByID(ctx context.Context, tenantID, executionID string) (*workflow.Execution, error)
}

// WorkflowExecutor defines the interface for executing workflows
type WorkflowExecutor interface {
	Execute(ctx context.Context, execution *workflow.Execution) error
}
