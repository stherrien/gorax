package slack

import (
	"context"
	"fmt"
	"strings"

	"github.com/gorax/gorax/internal/credential"
	"github.com/gorax/gorax/internal/executor/actions"
)

// SendDMAction implements the Slack SendDM (Direct Message) action
type SendDMAction struct {
	credentialService credential.Service
	baseURL           string // For testing, defaults to DefaultBaseURL
}

// NewSendDMAction creates a new SendDM action
func NewSendDMAction(credentialService credential.Service) *SendDMAction {
	return &SendDMAction{
		credentialService: credentialService,
		baseURL:           DefaultBaseURL,
	}
}

// Execute implements the Action interface
func (a *SendDMAction) Execute(ctx context.Context, input *actions.ActionInput) (*actions.ActionOutput, error) {
	// Parse config
	config, ok := input.Config.(SendDMConfig)
	if !ok {
		return nil, fmt.Errorf("invalid config type: expected SendDMConfig")
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

	// Resolve user identifier to user ID
	userID, err := a.resolveUserID(ctx, client, config.User)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve user: %w", err)
	}

	// Open DM conversation
	conversation, err := client.OpenConversation(ctx, []string{userID})
	if err != nil {
		return nil, fmt.Errorf("failed to open conversation: %w", err)
	}

	// Build SendMessage request
	req := &SendMessageRequest{
		Channel: conversation.ID,
		Text:    config.Text,
		Blocks:  config.Blocks,
	}

	// Send message
	resp, err := client.SendMessage(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to send message: %w", err)
	}

	// Build result
	result := &SendDMResult{
		OK:        resp.OK,
		UserID:    userID,
		Channel:   resp.Channel,
		Timestamp: resp.TS,
		Message:   &resp.Message,
	}

	// Create output
	output := actions.NewActionOutput(result)
	output.WithMetadata("user_id", userID)
	output.WithMetadata("channel", resp.Channel)
	output.WithMetadata("timestamp", resp.TS)
	output.WithMetadata("message_type", resp.Message.Type)

	return output, nil
}

// resolveUserID resolves a user identifier (email or user ID) to a Slack user ID
func (a *SendDMAction) resolveUserID(ctx context.Context, client *Client, userIdentifier string) (string, error) {
	// Check if it's an email
	if isEmail(userIdentifier) {
		// Look up user by email
		user, err := client.GetUserByEmail(ctx, userIdentifier)
		if err != nil {
			return "", fmt.Errorf("failed to lookup user by email: %w", err)
		}
		return user.ID, nil
	}

	// Assume it's a user ID
	return userIdentifier, nil
}

// isEmail checks if a string is an email address
func isEmail(s string) bool {
	// Simple email validation - contains @ and has parts before and after
	if !strings.Contains(s, "@") {
		return false
	}

	parts := strings.Split(s, "@")
	if len(parts) != 2 {
		return false
	}

	// Check that both parts are non-empty
	if len(parts[0]) == 0 || len(parts[1]) == 0 {
		return false
	}

	// Check that domain part contains a dot
	if !strings.Contains(parts[1], ".") {
		return false
	}

	return true
}

// SendDMResult represents the result of sending a direct message
type SendDMResult struct {
	OK        bool     `json:"ok"`
	UserID    string   `json:"user_id"`
	Channel   string   `json:"channel"`
	Timestamp string   `json:"timestamp"`
	Message   *Message `json:"message,omitempty"`
}
