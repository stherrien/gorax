package executor

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/gorax/gorax/internal/executor/actions"
	"github.com/gorax/gorax/internal/integrations/slack"
	"github.com/gorax/gorax/internal/workflow"
)

// executeSlackSendMessageAction executes a Slack send message action
func (e *Executor) executeSlackSendMessageAction(ctx context.Context, node workflow.Node, execCtx *ExecutionContext) (interface{}, error) {
	// Check if credential service is available
	if e.credentialService == nil {
		return nil, fmt.Errorf("credential service not available for Slack actions")
	}

	// Extract config from node data
	configData := node.Data.Config
	if len(configData) == 0 {
		return nil, fmt.Errorf("missing config for Slack send message action")
	}

	// Parse node config
	var config slack.SendMessageConfig
	if err := json.Unmarshal(configData, &config); err != nil {
		return nil, fmt.Errorf("failed to parse Slack send message config: %w", err)
	}

	// Create Slack action
	action := slack.NewSendMessageAction(e.credentialService)

	// Build execution context
	context := buildInterpolationContext(execCtx)

	// Create action input
	input := actions.NewActionInput(config, context)

	// Execute action
	output, err := action.Execute(ctx, input)
	if err != nil {
		return nil, err
	}

	return output.Data, nil
}

// executeSlackSendDMAction executes a Slack send direct message action
func (e *Executor) executeSlackSendDMAction(ctx context.Context, node workflow.Node, execCtx *ExecutionContext) (interface{}, error) {
	// Check if credential service is available
	if e.credentialService == nil {
		return nil, fmt.Errorf("credential service not available for Slack actions")
	}

	// Extract config from node data
	configData := node.Data.Config
	if len(configData) == 0 {
		return nil, fmt.Errorf("missing config for Slack send DM action")
	}

	// Parse node config
	var config slack.SendDMConfig
	if err := json.Unmarshal(configData, &config); err != nil {
		return nil, fmt.Errorf("failed to parse Slack send DM config: %w", err)
	}

	// Create Slack action
	action := slack.NewSendDMAction(e.credentialService)

	// Build execution context
	context := buildInterpolationContext(execCtx)

	// Create action input
	input := actions.NewActionInput(config, context)

	// Execute action
	output, err := action.Execute(ctx, input)
	if err != nil {
		return nil, err
	}

	return output.Data, nil
}

// executeSlackUpdateMessageAction executes a Slack update message action
func (e *Executor) executeSlackUpdateMessageAction(ctx context.Context, node workflow.Node, execCtx *ExecutionContext) (interface{}, error) {
	// Check if credential service is available
	if e.credentialService == nil {
		return nil, fmt.Errorf("credential service not available for Slack actions")
	}

	// Extract config from node data
	configData := node.Data.Config
	if len(configData) == 0 {
		return nil, fmt.Errorf("missing config for Slack update message action")
	}

	// Parse node config
	var config slack.UpdateMessageConfig
	if err := json.Unmarshal(configData, &config); err != nil {
		return nil, fmt.Errorf("failed to parse Slack update message config: %w", err)
	}

	// Create Slack action
	action := slack.NewUpdateMessageAction(e.credentialService)

	// Build execution context
	context := buildInterpolationContext(execCtx)

	// Create action input
	input := actions.NewActionInput(config, context)

	// Execute action
	output, err := action.Execute(ctx, input)
	if err != nil {
		return nil, err
	}

	return output.Data, nil
}

// executeSlackAddReactionAction executes a Slack add reaction action
func (e *Executor) executeSlackAddReactionAction(ctx context.Context, node workflow.Node, execCtx *ExecutionContext) (interface{}, error) {
	// Check if credential service is available
	if e.credentialService == nil {
		return nil, fmt.Errorf("credential service not available for Slack actions")
	}

	// Extract config from node data
	configData := node.Data.Config
	if len(configData) == 0 {
		return nil, fmt.Errorf("missing config for Slack add reaction action")
	}

	// Parse node config
	var config slack.AddReactionConfig
	if err := json.Unmarshal(configData, &config); err != nil {
		return nil, fmt.Errorf("failed to parse Slack add reaction config: %w", err)
	}

	// Create Slack action
	action := slack.NewAddReactionAction(e.credentialService)

	// Build execution context
	context := buildInterpolationContext(execCtx)

	// Create action input
	input := actions.NewActionInput(config, context)

	// Execute action
	output, err := action.Execute(ctx, input)
	if err != nil {
		return nil, err
	}

	return output.Data, nil
}
