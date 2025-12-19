package slack

import (
	"context"
	"fmt"

	"github.com/gorax/gorax/internal/credential"
	"github.com/gorax/gorax/internal/executor/actions"
)

// SendMessageAction implements the Slack SendMessage action
type SendMessageAction struct {
	credentialService credential.Service
	baseURL           string // For testing, defaults to DefaultBaseURL
}

// NewSendMessageAction creates a new SendMessage action
func NewSendMessageAction(credentialService credential.Service) *SendMessageAction {
	return &SendMessageAction{
		credentialService: credentialService,
		baseURL:           DefaultBaseURL,
	}
}

// Execute implements the Action interface
func (a *SendMessageAction) Execute(ctx context.Context, input *actions.ActionInput) (*actions.ActionOutput, error) {
	// Parse config
	config, ok := input.Config.(SendMessageConfig)
	if !ok {
		return nil, fmt.Errorf("invalid config type: expected SendMessageConfig")
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

	// Build SendMessage request
	req := &SendMessageRequest{
		Channel:        config.Channel,
		Text:           config.Text,
		Blocks:         config.Blocks,
		ThreadTS:       config.ThreadTS,
		ReplyBroadcast: config.ReplyBroadcast,
		IconEmoji:      config.IconEmoji,
		Username:       config.Username,
	}

	// Set unfurl options if provided
	if config.UnfurlLinks != nil {
		req.UnfurlLinks = *config.UnfurlLinks
	}
	if config.UnfurlMedia != nil {
		req.UnfurlMedia = *config.UnfurlMedia
	}

	// Send message
	resp, err := client.SendMessage(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to send message: %w", err)
	}

	// Build result
	result := &SendMessageResult{
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

	return output, nil
}

// SendMessageResult represents the result of sending a Slack message
type SendMessageResult struct {
	OK        bool     `json:"ok"`
	Channel   string   `json:"channel"`
	Timestamp string   `json:"timestamp"`
	Message   *Message `json:"message,omitempty"`
}

// extractString extracts a string value from a nested map using dot notation
// e.g., "env.tenant_id" from map["env"]["tenant_id"]
func extractString(data map[string]interface{}, path string) (string, error) {
	// Simple implementation for now - can be enhanced with proper JSONPath later
	keys := parsePath(path)
	current := data

	for i, key := range keys {
		if i == len(keys)-1 {
			// Last key - should be the value
			if val, ok := current[key]; ok {
				if str, ok := val.(string); ok {
					return str, nil
				}
				return "", fmt.Errorf("value at '%s' is not a string", path)
			}
			return "", fmt.Errorf("key '%s' not found in context", path)
		}

		// Intermediate key - should be a map
		if val, ok := current[key]; ok {
			if m, ok := val.(map[string]interface{}); ok {
				current = m
			} else {
				return "", fmt.Errorf("value at '%s' is not a map", key)
			}
		} else {
			return "", fmt.Errorf("key '%s' not found in context", key)
		}
	}

	return "", fmt.Errorf("failed to extract value from path '%s'", path)
}

// parsePath splits a dot-notation path into keys
// e.g., "env.tenant_id" -> ["env", "tenant_id"]
func parsePath(path string) []string {
	// Simple split by dot - can be enhanced to handle escaped dots later
	result := []string{}
	current := ""

	for _, ch := range path {
		if ch == '.' {
			if current != "" {
				result = append(result, current)
				current = ""
			}
		} else {
			current += string(ch)
		}
	}

	if current != "" {
		result = append(result, current)
	}

	return result
}
