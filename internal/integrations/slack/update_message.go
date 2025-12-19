package slack

import (
	"context"
	"fmt"

	"github.com/gorax/gorax/internal/credential"
	"github.com/gorax/gorax/internal/executor/actions"
)

// UpdateMessageAction implements the Slack UpdateMessage action
type UpdateMessageAction struct {
	credentialService credential.Service
	baseURL           string // For testing, defaults to DefaultBaseURL
}

// NewUpdateMessageAction creates a new UpdateMessage action
func NewUpdateMessageAction(credentialService credential.Service) *UpdateMessageAction {
	return &UpdateMessageAction{
		credentialService: credentialService,
		baseURL:           DefaultBaseURL,
	}
}

// Execute implements the Action interface
func (a *UpdateMessageAction) Execute(ctx context.Context, input *actions.ActionInput) (*actions.ActionOutput, error) {
	// Parse config
	config, ok := input.Config.(UpdateMessageConfig)
	if !ok {
		return nil, fmt.Errorf("invalid config type: expected UpdateMessageConfig")
	}

	// Validate config
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	// Extract tenant_id and credential_id from context
	tenantID, err := extractString(input.Context, "env.tenant_id")
	if err != nil {
		return nil, fmt.Errorf("tenant_id is required in context: %w", err)
	}

	credentialID, err := extractString(input.Context, "credential_id")
	if err != nil {
		return nil, fmt.Errorf("credential_id is required in context: %w", err)
	}

	// Retrieve and decrypt credential
	decryptedCred, err := a.credentialService.GetValue(ctx, tenantID, credentialID, "system")
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve credential: %w", err)
	}

	// Extract access_token from credential
	accessToken, ok := decryptedCred.Value["access_token"].(string)
	if !ok || accessToken == "" {
		return nil, fmt.Errorf("access_token not found in credential")
	}

	// Create Slack client
	client, err := NewClient(accessToken)
	if err != nil {
		return nil, fmt.Errorf("failed to create Slack client: %w", err)
	}

	// Override base URL if set (for testing)
	if a.baseURL != "" {
		client.baseURL = a.baseURL
	}

	// Build UpdateMessage request
	req := &UpdateMessageRequest{
		Channel: config.Channel,
		TS:      config.TS,
		Text:    config.Text,
		Blocks:  config.Blocks,
	}

	// Update message
	resp, err := client.UpdateMessage(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to update message: %w", err)
	}

	// Build result
	result := &UpdateMessageResult{
		OK:        resp.OK,
		Channel:   resp.Channel,
		Timestamp: resp.TS,
		Message:   &resp.Message,
	}

	// Create output
	output := actions.NewActionOutput(result)
	output.WithMetadata("channel", resp.Channel)
	output.WithMetadata("timestamp", resp.TS)
	output.WithMetadata("message_type", resp.Message.Type)
	output.WithMetadata("updated", true)

	return output, nil
}

// UpdateMessageResult represents the result of updating a Slack message
type UpdateMessageResult struct {
	OK        bool     `json:"ok"`
	Channel   string   `json:"channel"`
	Timestamp string   `json:"timestamp"`
	Message   *Message `json:"message,omitempty"`
}
