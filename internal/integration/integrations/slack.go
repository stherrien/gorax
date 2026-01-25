// Package integrations provides implementations of external service integrations
// for the Gorax workflow automation platform.
package integrations

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/gorax/gorax/internal/integration"
	inthttp "github.com/gorax/gorax/internal/integration/http"
)

const (
	slackAPIBaseURL = "https://slack.com/api"
	slackIntName    = "slack"
)

// SlackIntegration provides Slack API integration capabilities.
type SlackIntegration struct {
	*integration.BaseIntegration
	client *inthttp.Client
	logger *slog.Logger
}

// SlackAction represents the available Slack actions.
type SlackAction string

const (
	// SlackActionSendMessage sends a message to a channel.
	SlackActionSendMessage SlackAction = "send_message"
	// SlackActionSendDM sends a direct message to a user.
	SlackActionSendDM SlackAction = "send_dm"
	// SlackActionUpdateMessage updates an existing message.
	SlackActionUpdateMessage SlackAction = "update_message"
	// SlackActionCreateChannel creates a new channel.
	SlackActionCreateChannel SlackAction = "create_channel"
	// SlackActionInviteToChannel invites users to a channel.
	SlackActionInviteToChannel SlackAction = "invite_to_channel"
	// SlackActionSetChannelTopic sets the topic for a channel.
	SlackActionSetChannelTopic SlackAction = "set_channel_topic"
	// SlackActionLookupUser looks up a user by email.
	SlackActionLookupUser SlackAction = "lookup_user"
)

// NewSlackIntegration creates a new Slack integration.
func NewSlackIntegration(logger *slog.Logger) *SlackIntegration {
	if logger == nil {
		logger = slog.Default()
	}

	base := integration.NewBaseIntegration(slackIntName, integration.TypeAPI)
	base.SetMetadata(&integration.Metadata{
		Name:        slackIntName,
		DisplayName: "Slack",
		Description: "Send messages, manage channels, and interact with Slack workspaces",
		Version:     "1.0.0",
		Category:    "messaging",
		Tags:        []string{"slack", "messaging", "chat", "notifications"},
		Author:      "Gorax",
	})
	base.SetSchema(buildSlackSchema())

	client := inthttp.NewClient(
		inthttp.WithBaseURL(slackAPIBaseURL),
		inthttp.WithTimeout(30*time.Second),
		inthttp.WithLogger(logger),
		inthttp.WithRetryConfig(buildSlackRetryConfig()),
	)

	return &SlackIntegration{
		BaseIntegration: base,
		client:          client,
		logger:          logger,
	}
}

// Execute performs a Slack API action.
func (s *SlackIntegration) Execute(ctx context.Context, config *integration.Config, params integration.JSONMap) (*integration.Result, error) {
	start := time.Now()

	// Get action type
	action, ok := params.GetString("action")
	if !ok || action == "" {
		err := integration.NewValidationError("action", "action is required", nil)
		return integration.NewErrorResult(err, "VALIDATION_ERROR", time.Since(start).Milliseconds()), err
	}

	// Get token from credentials
	token, err := s.getToken(config)
	if err != nil {
		return integration.NewErrorResult(err, "AUTH_ERROR", time.Since(start).Milliseconds()), err
	}

	// Execute the appropriate action
	var result *integration.Result
	switch SlackAction(action) {
	case SlackActionSendMessage:
		result, err = s.sendMessage(ctx, token, params, start)
	case SlackActionSendDM:
		result, err = s.sendDM(ctx, token, params, start)
	case SlackActionUpdateMessage:
		result, err = s.updateMessage(ctx, token, params, start)
	case SlackActionCreateChannel:
		result, err = s.createChannel(ctx, token, params, start)
	case SlackActionInviteToChannel:
		result, err = s.inviteToChannel(ctx, token, params, start)
	case SlackActionSetChannelTopic:
		result, err = s.setChannelTopic(ctx, token, params, start)
	case SlackActionLookupUser:
		result, err = s.lookupUser(ctx, token, params, start)
	default:
		err = integration.NewValidationError("action", "unsupported action", action)
		result = integration.NewErrorResult(err, "INVALID_ACTION", time.Since(start).Milliseconds())
	}

	if err != nil {
		s.logger.Error("slack action failed",
			"action", action,
			"error", err,
			"duration_ms", time.Since(start).Milliseconds(),
		)
	} else {
		s.logger.Info("slack action completed",
			"action", action,
			"duration_ms", result.Duration,
		)
	}

	return result, err
}

// Validate validates the integration configuration.
func (s *SlackIntegration) Validate(config *integration.Config) error {
	if err := s.BaseIntegration.ValidateConfig(config); err != nil {
		return err
	}

	// Validate credentials
	if config.Credentials == nil {
		return integration.NewValidationError("credentials", "credentials are required", nil)
	}

	if _, err := s.getToken(config); err != nil {
		return err
	}

	return nil
}

// getToken extracts the OAuth token from credentials.
func (s *SlackIntegration) getToken(config *integration.Config) (string, error) {
	if config.Credentials == nil || config.Credentials.Data == nil {
		return "", integration.NewValidationError("credentials", "credentials are required", nil)
	}

	token, ok := config.Credentials.Data.GetString("token")
	if !ok || token == "" {
		token, ok = config.Credentials.Data.GetString("access_token")
		if !ok || token == "" {
			return "", integration.NewValidationError("token", "OAuth token is required", nil)
		}
	}

	return token, nil
}

// sendMessage sends a message to a Slack channel.
func (s *SlackIntegration) sendMessage(ctx context.Context, token string, params integration.JSONMap, start time.Time) (*integration.Result, error) {
	channel, ok := params.GetString("channel")
	if !ok || channel == "" {
		err := integration.NewValidationError("channel", "channel is required", nil)
		return integration.NewErrorResult(err, "VALIDATION_ERROR", time.Since(start).Milliseconds()), err
	}

	text, _ := params.GetString("text")
	threadTS, _ := params.GetString("thread_ts")

	payload := map[string]any{
		"channel": channel,
	}

	if text != "" {
		payload["text"] = text
	}

	if threadTS != "" {
		payload["thread_ts"] = threadTS
	}

	// Handle blocks for rich messages
	if blocks, ok := params.Get("blocks"); ok {
		payload["blocks"] = blocks
	}

	// Handle attachments
	if attachments, ok := params.Get("attachments"); ok {
		payload["attachments"] = attachments
	}

	// Handle additional options
	if unfurlLinks, ok := params.GetBool("unfurl_links"); ok {
		payload["unfurl_links"] = unfurlLinks
	}
	if unfurlMedia, ok := params.GetBool("unfurl_media"); ok {
		payload["unfurl_media"] = unfurlMedia
	}

	return s.executeSlackAPI(ctx, token, "chat.postMessage", payload, start)
}

// sendDM sends a direct message to a user.
func (s *SlackIntegration) sendDM(ctx context.Context, token string, params integration.JSONMap, start time.Time) (*integration.Result, error) {
	// Get user ID or email
	userID, hasUserID := params.GetString("user_id")
	email, hasEmail := params.GetString("email")

	if !hasUserID && !hasEmail {
		err := integration.NewValidationError("user_id", "user_id or email is required", nil)
		return integration.NewErrorResult(err, "VALIDATION_ERROR", time.Since(start).Milliseconds()), err
	}

	// Look up user by email if needed
	if !hasUserID && hasEmail {
		lookupResult, err := s.lookupUser(ctx, token, integration.JSONMap{"email": email}, start)
		if err != nil {
			return lookupResult, err
		}
		if data, ok := lookupResult.Data.(map[string]any); ok {
			if user, ok := data["user"].(map[string]any); ok {
				if id, ok := user["id"].(string); ok {
					userID = id
				}
			}
		}
		if userID == "" {
			err := fmt.Errorf("user not found for email: %s", email)
			return integration.NewErrorResult(err, "USER_NOT_FOUND", time.Since(start).Milliseconds()), err
		}
	}

	// Open conversation with user
	convPayload := map[string]any{
		"users": userID,
	}

	convResult, err := s.executeSlackAPI(ctx, token, "conversations.open", convPayload, start)
	if err != nil {
		return convResult, err
	}

	// Extract channel ID from conversation
	var channelID string
	if data, ok := convResult.Data.(map[string]any); ok {
		if channel, ok := data["channel"].(map[string]any); ok {
			if id, ok := channel["id"].(string); ok {
				channelID = id
			}
		}
	}

	if channelID == "" {
		err := fmt.Errorf("failed to open conversation with user: %s", userID)
		return integration.NewErrorResult(err, "CONVERSATION_OPEN_FAILED", time.Since(start).Milliseconds()), err
	}

	// Send message to the DM channel
	params["channel"] = channelID
	return s.sendMessage(ctx, token, params, start)
}

// updateMessage updates an existing Slack message.
func (s *SlackIntegration) updateMessage(ctx context.Context, token string, params integration.JSONMap, start time.Time) (*integration.Result, error) {
	channel, ok := params.GetString("channel")
	if !ok || channel == "" {
		err := integration.NewValidationError("channel", "channel is required", nil)
		return integration.NewErrorResult(err, "VALIDATION_ERROR", time.Since(start).Milliseconds()), err
	}

	ts, ok := params.GetString("ts")
	if !ok || ts == "" {
		err := integration.NewValidationError("ts", "message timestamp (ts) is required", nil)
		return integration.NewErrorResult(err, "VALIDATION_ERROR", time.Since(start).Milliseconds()), err
	}

	text, _ := params.GetString("text")

	payload := map[string]any{
		"channel": channel,
		"ts":      ts,
	}

	if text != "" {
		payload["text"] = text
	}

	if blocks, ok := params.Get("blocks"); ok {
		payload["blocks"] = blocks
	}

	if attachments, ok := params.Get("attachments"); ok {
		payload["attachments"] = attachments
	}

	return s.executeSlackAPI(ctx, token, "chat.update", payload, start)
}

// createChannel creates a new Slack channel.
func (s *SlackIntegration) createChannel(ctx context.Context, token string, params integration.JSONMap, start time.Time) (*integration.Result, error) {
	name, ok := params.GetString("name")
	if !ok || name == "" {
		err := integration.NewValidationError("name", "channel name is required", nil)
		return integration.NewErrorResult(err, "VALIDATION_ERROR", time.Since(start).Milliseconds()), err
	}

	payload := map[string]any{
		"name": name,
	}

	if isPrivate, ok := params.GetBool("is_private"); ok {
		payload["is_private"] = isPrivate
	}

	return s.executeSlackAPI(ctx, token, "conversations.create", payload, start)
}

// inviteToChannel invites users to a channel.
func (s *SlackIntegration) inviteToChannel(ctx context.Context, token string, params integration.JSONMap, start time.Time) (*integration.Result, error) {
	channel, ok := params.GetString("channel")
	if !ok || channel == "" {
		err := integration.NewValidationError("channel", "channel is required", nil)
		return integration.NewErrorResult(err, "VALIDATION_ERROR", time.Since(start).Milliseconds()), err
	}

	// Get users - can be a single user ID or comma-separated list
	users, ok := params.GetString("users")
	if !ok || users == "" {
		// Try to get as array
		if usersArr, ok := params.Get("users"); ok {
			if arr, ok := usersArr.([]any); ok {
				userIDs := make([]string, 0, len(arr))
				for _, u := range arr {
					if str, ok := u.(string); ok {
						userIDs = append(userIDs, str)
					}
				}
				users = strings.Join(userIDs, ",")
			}
		}
	}

	if users == "" {
		err := integration.NewValidationError("users", "users is required", nil)
		return integration.NewErrorResult(err, "VALIDATION_ERROR", time.Since(start).Milliseconds()), err
	}

	payload := map[string]any{
		"channel": channel,
		"users":   users,
	}

	return s.executeSlackAPI(ctx, token, "conversations.invite", payload, start)
}

// setChannelTopic sets the topic for a channel.
func (s *SlackIntegration) setChannelTopic(ctx context.Context, token string, params integration.JSONMap, start time.Time) (*integration.Result, error) {
	channel, ok := params.GetString("channel")
	if !ok || channel == "" {
		err := integration.NewValidationError("channel", "channel is required", nil)
		return integration.NewErrorResult(err, "VALIDATION_ERROR", time.Since(start).Milliseconds()), err
	}

	topic, ok := params.GetString("topic")
	if !ok {
		err := integration.NewValidationError("topic", "topic is required", nil)
		return integration.NewErrorResult(err, "VALIDATION_ERROR", time.Since(start).Milliseconds()), err
	}

	payload := map[string]any{
		"channel": channel,
		"topic":   topic,
	}

	return s.executeSlackAPI(ctx, token, "conversations.setTopic", payload, start)
}

// lookupUser looks up a user by email.
func (s *SlackIntegration) lookupUser(ctx context.Context, token string, params integration.JSONMap, start time.Time) (*integration.Result, error) {
	email, ok := params.GetString("email")
	if !ok || email == "" {
		err := integration.NewValidationError("email", "email is required", nil)
		return integration.NewErrorResult(err, "VALIDATION_ERROR", time.Since(start).Milliseconds()), err
	}

	payload := map[string]any{
		"email": email,
	}

	return s.executeSlackAPI(ctx, token, "users.lookupByEmail", payload, start)
}

// executeSlackAPI executes a Slack API method.
func (s *SlackIntegration) executeSlackAPI(ctx context.Context, token, method string, payload map[string]any, start time.Time) (*integration.Result, error) {
	resp, err := s.client.Post(ctx, method,
		payload,
		inthttp.WithRequestHeader("Authorization", "Bearer "+token),
		inthttp.WithRequestHeader("Content-Type", "application/json; charset=utf-8"),
	)
	if err != nil {
		return integration.NewErrorResult(err, "API_ERROR", time.Since(start).Milliseconds()), err
	}

	// Parse response
	var slackResp map[string]any
	if err := json.Unmarshal(resp.Body, &slackResp); err != nil {
		return integration.NewErrorResult(err, "PARSE_ERROR", time.Since(start).Milliseconds()), err
	}

	// Check for Slack API errors
	ok, _ := slackResp["ok"].(bool)
	if !ok {
		errMsg := "unknown error"
		if e, hasErr := slackResp["error"].(string); hasErr {
			errMsg = e
		}
		err := fmt.Errorf("slack API error: %s", errMsg)

		// Handle rate limiting
		if errMsg == "ratelimited" {
			retryAfter := 1
			if headers := resp.Headers.Get("Retry-After"); headers != "" {
				fmt.Sscanf(headers, "%d", &retryAfter)
			}
			err = fmt.Errorf("%w: retry after %d seconds", integration.ErrRateLimited, retryAfter)
		}

		result := integration.NewErrorResult(err, "SLACK_ERROR", time.Since(start).Milliseconds())
		result.Data = slackResp
		return result, err
	}

	return &integration.Result{
		Success:    true,
		Data:       slackResp,
		StatusCode: http.StatusOK,
		Duration:   time.Since(start).Milliseconds(),
		ExecutedAt: time.Now().UTC(),
	}, nil
}

// buildSlackSchema builds the schema for the Slack integration.
func buildSlackSchema() *integration.Schema {
	return &integration.Schema{
		ConfigSpec: map[string]integration.FieldSpec{
			"token": {
				Name:        "token",
				Type:        integration.FieldTypeSecret,
				Description: "Slack OAuth token (Bot or User token)",
				Required:    true,
				Sensitive:   true,
			},
		},
		InputSpec: map[string]integration.FieldSpec{
			"action": {
				Name:        "action",
				Type:        integration.FieldTypeString,
				Description: "Action to perform",
				Required:    true,
				Options: []string{
					string(SlackActionSendMessage),
					string(SlackActionSendDM),
					string(SlackActionUpdateMessage),
					string(SlackActionCreateChannel),
					string(SlackActionInviteToChannel),
					string(SlackActionSetChannelTopic),
					string(SlackActionLookupUser),
				},
			},
			"channel": {
				Name:        "channel",
				Type:        integration.FieldTypeString,
				Description: "Channel ID or name (for send_message, update_message, invite_to_channel, set_channel_topic)",
				Required:    false,
			},
			"text": {
				Name:        "text",
				Type:        integration.FieldTypeString,
				Description: "Message text",
				Required:    false,
			},
			"blocks": {
				Name:        "blocks",
				Type:        integration.FieldTypeArray,
				Description: "Block Kit blocks for rich messages",
				Required:    false,
			},
			"attachments": {
				Name:        "attachments",
				Type:        integration.FieldTypeArray,
				Description: "Legacy attachments",
				Required:    false,
			},
			"thread_ts": {
				Name:        "thread_ts",
				Type:        integration.FieldTypeString,
				Description: "Thread timestamp for replying to a thread",
				Required:    false,
			},
			"ts": {
				Name:        "ts",
				Type:        integration.FieldTypeString,
				Description: "Message timestamp (for update_message)",
				Required:    false,
			},
			"user_id": {
				Name:        "user_id",
				Type:        integration.FieldTypeString,
				Description: "User ID (for send_dm)",
				Required:    false,
			},
			"email": {
				Name:        "email",
				Type:        integration.FieldTypeString,
				Description: "User email (for send_dm, lookup_user)",
				Required:    false,
			},
			"users": {
				Name:        "users",
				Type:        integration.FieldTypeString,
				Description: "Comma-separated user IDs (for invite_to_channel)",
				Required:    false,
			},
			"name": {
				Name:        "name",
				Type:        integration.FieldTypeString,
				Description: "Channel name (for create_channel)",
				Required:    false,
			},
			"is_private": {
				Name:        "is_private",
				Type:        integration.FieldTypeBoolean,
				Description: "Whether the channel is private (for create_channel)",
				Required:    false,
			},
			"topic": {
				Name:        "topic",
				Type:        integration.FieldTypeString,
				Description: "Channel topic (for set_channel_topic)",
				Required:    false,
			},
		},
		OutputSpec: map[string]integration.FieldSpec{
			"ok": {
				Name:        "ok",
				Type:        integration.FieldTypeBoolean,
				Description: "Whether the API call was successful",
			},
			"channel": {
				Name:        "channel",
				Type:        integration.FieldTypeObject,
				Description: "Channel information",
			},
			"message": {
				Name:        "message",
				Type:        integration.FieldTypeObject,
				Description: "Message information",
			},
			"ts": {
				Name:        "ts",
				Type:        integration.FieldTypeString,
				Description: "Message timestamp",
			},
			"user": {
				Name:        "user",
				Type:        integration.FieldTypeObject,
				Description: "User information",
			},
		},
	}
}

// buildSlackRetryConfig builds retry configuration for Slack API.
func buildSlackRetryConfig() *inthttp.RetryConfig {
	return &inthttp.RetryConfig{
		MaxRetries:   3,
		BaseDelay:    500 * time.Millisecond,
		MaxDelay:     30 * time.Second,
		Multiplier:   2.0,
		JitterFactor: 0.1,
		ShouldRetry: func(err error, resp *inthttp.Response) bool {
			if err != nil {
				return inthttp.IsRetryableError(err)
			}
			if resp == nil {
				return false
			}
			// Retry on rate limiting (429) and server errors (5xx)
			return resp.StatusCode == 429 || resp.StatusCode >= 500
		},
	}
}
