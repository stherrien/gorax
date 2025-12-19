package slack

import (
	"context"
	"fmt"
	"strings"

	"github.com/gorax/gorax/internal/credential"
	"github.com/gorax/gorax/internal/executor/actions"
)

// AddReactionAction implements the Slack AddReaction action
type AddReactionAction struct {
	credentialService credential.Service
	baseURL           string // For testing, defaults to DefaultBaseURL
}

// NewAddReactionAction creates a new AddReaction action
func NewAddReactionAction(credentialService credential.Service) *AddReactionAction {
	return &AddReactionAction{
		credentialService: credentialService,
		baseURL:           DefaultBaseURL,
	}
}

// Execute implements the Action interface
func (a *AddReactionAction) Execute(ctx context.Context, input *actions.ActionInput) (*actions.ActionOutput, error) {
	// Parse config
	config, ok := input.Config.(AddReactionConfig)
	if !ok {
		return nil, fmt.Errorf("invalid config type: expected AddReactionConfig")
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

	// Normalize emoji name (remove colons if present)
	emoji := normalizeEmoji(config.Emoji)

	// Add reaction (client already handles "already_reacted" as success)
	err = client.AddReaction(ctx, config.Channel, config.Timestamp, emoji)
	if err != nil {
		return nil, fmt.Errorf("failed to add reaction: %w", err)
	}

	// Build result
	result := &AddReactionResult{
		OK:        true,
		Channel:   config.Channel,
		Timestamp: config.Timestamp,
		Emoji:     emoji,
	}

	// Create output
	output := actions.NewActionOutput(result)
	output.WithMetadata("channel", config.Channel)
	output.WithMetadata("timestamp", config.Timestamp)
	output.WithMetadata("emoji", emoji)

	return output, nil
}

// normalizeEmoji removes leading and trailing colons from emoji names
// Slack accepts emoji names without colons, but users often include them
func normalizeEmoji(emoji string) string {
	return strings.Trim(emoji, ":")
}

// AddReactionResult represents the result of adding a reaction to a Slack message
type AddReactionResult struct {
	OK        bool   `json:"ok"`
	Channel   string `json:"channel"`
	Timestamp string `json:"timestamp"`
	Emoji     string `json:"emoji"`
}
