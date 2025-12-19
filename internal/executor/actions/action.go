package actions

import (
	"context"
)

// Action represents a workflow action that can be executed
type Action interface {
	// Execute runs the action with the given context and input
	// Returns the action output or an error if execution fails
	Execute(ctx context.Context, input *ActionInput) (*ActionOutput, error)
}

// ActionInput represents the input data for an action execution
type ActionInput struct {
	// Config contains the action-specific configuration
	Config interface{}
	// Context contains data from previous steps and trigger
	Context map[string]interface{}
}

// ActionOutput represents the result of an action execution
type ActionOutput struct {
	// Data contains the output data from the action
	Data interface{}
	// Metadata contains additional information about the execution
	Metadata map[string]interface{}
}

// ActionFactory is a function that creates an action instance
type ActionFactory func() Action

// NewActionInput creates a new ActionInput
func NewActionInput(config interface{}, context map[string]interface{}) *ActionInput {
	if context == nil {
		context = make(map[string]interface{})
	}
	return &ActionInput{
		Config:  config,
		Context: context,
	}
}

// NewActionOutput creates a new ActionOutput
func NewActionOutput(data interface{}) *ActionOutput {
	return &ActionOutput{
		Data:     data,
		Metadata: make(map[string]interface{}),
	}
}

// WithMetadata adds metadata to the output
func (o *ActionOutput) WithMetadata(key string, value interface{}) *ActionOutput {
	o.Metadata[key] = value
	return o
}
