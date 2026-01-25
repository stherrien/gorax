package notification

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// SlackConfig holds Slack notification configuration
type SlackConfig struct {
	WebhookURL string
	MaxRetries int
	RetryDelay time.Duration
	Timeout    time.Duration
}

// SlackMessage represents a Slack message
type SlackMessage struct {
	Text    string       `json:"text"`
	Blocks  []SlackBlock `json:"blocks,omitempty"`
	Channel string       `json:"channel,omitempty"`
}

// SlackBlock represents a Slack Block Kit block
type SlackBlock struct {
	Type      string         `json:"type"`
	Text      *SlackText     `json:"text,omitempty"`
	Fields    []SlackText    `json:"fields,omitempty"`
	Accessory *SlackElement  `json:"accessory,omitempty"`
	Elements  []SlackElement `json:"elements,omitempty"`
}

// SlackText represents text in a Slack block
type SlackText struct {
	Type  string `json:"type"`
	Text  string `json:"text"`
	Emoji bool   `json:"emoji,omitempty"`
}

// SlackElement represents an element in a Slack block
type SlackElement struct {
	Type  string `json:"type"`
	Text  string `json:"text,omitempty"`
	URL   string `json:"url,omitempty"`
	Value string `json:"value,omitempty"`
}

// slackWebhookPayload is the internal payload structure
type slackWebhookPayload struct {
	Text    string       `json:"text"`
	Blocks  []SlackBlock `json:"blocks,omitempty"`
	Channel string       `json:"channel,omitempty"`
}

// SlackNotifier sends Slack notifications
type SlackNotifier struct {
	config     SlackConfig
	httpClient *http.Client
}

// NewSlackNotifier creates a new Slack notifier
func NewSlackNotifier(cfg SlackConfig) (*SlackNotifier, error) {
	if err := validateSlackConfig(cfg); err != nil {
		return nil, err
	}

	timeout := cfg.Timeout
	if timeout == 0 {
		timeout = 30 * time.Second
	}

	return &SlackNotifier{
		config: cfg,
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}, nil
}

// Send sends a Slack notification
func (s *SlackNotifier) Send(ctx context.Context, msg SlackMessage) error {
	// Check context
	if err := ctx.Err(); err != nil {
		return err
	}

	// Validate message
	if err := validateSlackMessage(msg); err != nil {
		return err
	}

	maxRetries := s.config.MaxRetries
	if maxRetries == 0 {
		maxRetries = 3
	}

	retryDelay := s.config.RetryDelay
	if retryDelay == 0 {
		retryDelay = time.Second
	}

	var lastErr error
	for attempt := 0; attempt <= maxRetries; attempt++ {
		// Check context
		if err := ctx.Err(); err != nil {
			return err
		}

		err := s.sendOnce(ctx, msg)
		if err == nil {
			return nil
		}

		lastErr = err

		// Don't retry on last attempt
		if attempt == maxRetries {
			break
		}

		// Check if error is retryable
		if !isSlackRetryableError(err) {
			return err
		}

		// Wait before retry
		select {
		case <-time.After(retryDelay):
			// Continue to next attempt
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	return fmt.Errorf("failed to send Slack notification after %d attempts: %w", maxRetries+1, lastErr)
}

// sendOnce performs a single send attempt
func (s *SlackNotifier) sendOnce(ctx context.Context, msg SlackMessage) error {
	// Build payload
	payload := slackWebhookPayload{
		Text:    msg.Text,
		Blocks:  msg.Blocks,
		Channel: msg.Channel,
	}

	// Marshal to JSON
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	// Create request
	req, err := http.NewRequestWithContext(ctx, "POST", s.config.WebhookURL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// Send request
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// Handle rate limiting
	if resp.StatusCode == http.StatusTooManyRequests {
		retryAfter := 1 // Default to 1 second
		if retryHeader := resp.Header.Get("Retry-After"); retryHeader != "" {
			if seconds, err := strconv.Atoi(retryHeader); err == nil {
				retryAfter = seconds
			}
		}

		time.Sleep(time.Duration(retryAfter) * time.Second)
		return fmt.Errorf("rate limited, retry after %d seconds", retryAfter)
	}

	// Check response status
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Slack webhook returned status %d", resp.StatusCode)
	}

	return nil
}

// BuildTaskAssignedMessage builds a Slack message for task assignment
func BuildTaskAssignedMessage(assignee, taskTitle, taskURL string) SlackMessage {
	return SlackMessage{
		Text: fmt.Sprintf("Task assigned to %s: %s", assignee, taskTitle),
		Blocks: []SlackBlock{
			{
				Type: "header",
				Text: &SlackText{
					Type: "plain_text",
					Text: "📋 Task Assigned",
				},
			},
			{
				Type: "section",
				Text: &SlackText{
					Type: "mrkdwn",
					Text: fmt.Sprintf("*Assignee:* %s\n*Task:* %s", assignee, taskTitle),
				},
			},
			{
				Type: "actions",
				Elements: []SlackElement{
					{
						Type: "button",
						Text: "View Task",
						URL:  taskURL,
					},
				},
			},
		},
	}
}

// BuildTaskCompletedMessage builds a Slack message for task completion
func BuildTaskCompletedMessage(taskTitle, status, completedBy string) SlackMessage {
	emoji := "✅"
	if status == "rejected" {
		emoji = "❌"
	}

	return SlackMessage{
		Text: fmt.Sprintf("Task %s completed: %s (by %s)", emoji, taskTitle, completedBy),
		Blocks: []SlackBlock{
			{
				Type: "header",
				Text: &SlackText{
					Type: "plain_text",
					Text: fmt.Sprintf("%s Task Completed", emoji),
				},
			},
			{
				Type: "section",
				Text: &SlackText{
					Type: "mrkdwn",
					Text: fmt.Sprintf("*Task:* %s\n*Status:* %s\n*Completed by:* %s", taskTitle, status, completedBy),
				},
			},
		},
	}
}

// BuildTaskOverdueMessage builds a Slack message for overdue task
func BuildTaskOverdueMessage(taskTitle string, dueDate time.Time, taskURL string) SlackMessage {
	overdueDuration := time.Since(dueDate)
	overdueDays := int(overdueDuration.Hours() / 24)

	overdueText := fmt.Sprintf("%d days", overdueDays)
	if overdueDays == 0 {
		overdueHours := int(overdueDuration.Hours())
		if overdueHours == 0 {
			overdueText = "less than an hour"
		} else {
			overdueText = fmt.Sprintf("%d hours", overdueHours)
		}
	}

	return SlackMessage{
		Text: fmt.Sprintf("⚠️ Task overdue: %s (overdue by %s)", taskTitle, overdueText),
		Blocks: []SlackBlock{
			{
				Type: "header",
				Text: &SlackText{
					Type: "plain_text",
					Text: "⚠️ Task Overdue",
				},
			},
			{
				Type: "section",
				Text: &SlackText{
					Type: "mrkdwn",
					Text: fmt.Sprintf("*Task:* %s\n*Due date:* %s\n*Overdue by:* %s", taskTitle, dueDate.Format("2006-01-02 15:04"), overdueText),
				},
			},
			{
				Type: "actions",
				Elements: []SlackElement{
					{
						Type: "button",
						Text: "View Task",
						URL:  taskURL,
					},
				},
			},
		},
	}
}

// BuildTaskEscalatedMessage builds a Slack message for task escalation
func BuildTaskEscalatedMessage(taskTitle string, level int, dueDate, taskURL string) SlackMessage {
	return SlackMessage{
		Text: fmt.Sprintf("⚠️ Task Escalated (Level %d): %s", level, taskTitle),
		Blocks: []SlackBlock{
			{
				Type: "header",
				Text: &SlackText{
					Type: "plain_text",
					Text: "⚠️ Task Escalated",
				},
			},
			{
				Type: "section",
				Text: &SlackText{
					Type: "mrkdwn",
					Text: fmt.Sprintf("*Task:* %s\n*Escalation Level:* %d\n*New Due Date:* %s", taskTitle, level, dueDate),
				},
			},
			{
				Type: "context",
				Elements: []SlackElement{
					{
						Type: "mrkdwn",
						Text: "Previous assignees did not respond in time. Task has been escalated to backup approvers.",
					},
				},
			},
			{
				Type: "actions",
				Elements: []SlackElement{
					{
						Type: "button",
						Text: "View Task",
						URL:  taskURL,
					},
				},
			},
		},
	}
}

// BuildWorkflowExecutionMessage builds a Slack message for workflow execution
func BuildWorkflowExecutionMessage(workflowName, status, errorMsg, executionURL string) SlackMessage {
	emoji := "✅"

	switch status {
	case "failed":
		emoji = "❌"
	case "running":
		emoji = "▶️"
	case "pending":
		emoji = "⏳"
	}

	text := fmt.Sprintf("%s Workflow %s: %s", emoji, status, workflowName)

	blocks := []SlackBlock{
		{
			Type: "header",
			Text: &SlackText{
				Type: "plain_text",
				Text: fmt.Sprintf("%s Workflow %s", emoji, strings.Title(status)),
			},
		},
		{
			Type: "section",
			Text: &SlackText{
				Type: "mrkdwn",
				Text: fmt.Sprintf("*Workflow:* %s\n*Status:* %s", workflowName, status),
			},
		},
	}

	if errorMsg != "" {
		blocks = append(blocks, SlackBlock{
			Type: "section",
			Text: &SlackText{
				Type: "mrkdwn",
				Text: fmt.Sprintf("*Error:* ```%s```", errorMsg),
			},
		})
	}

	blocks = append(blocks, SlackBlock{
		Type: "actions",
		Elements: []SlackElement{
			{
				Type: "button",
				Text: "View Execution",
				URL:  executionURL,
			},
		},
	})

	return SlackMessage{
		Text:   text,
		Blocks: blocks,
	}
}

// validateSlackConfig validates Slack configuration
func validateSlackConfig(cfg SlackConfig) error {
	if cfg.WebhookURL == "" {
		return fmt.Errorf("webhook URL is required")
	}

	if !strings.HasPrefix(cfg.WebhookURL, "http://") && !strings.HasPrefix(cfg.WebhookURL, "https://") {
		return fmt.Errorf("invalid webhook URL: must start with http:// or https://")
	}

	return nil
}

// validateSlackMessage validates a Slack message
func validateSlackMessage(msg SlackMessage) error {
	if msg.Text == "" && len(msg.Blocks) == 0 {
		return fmt.Errorf("message must have text or blocks")
	}

	return nil
}

// isSlackRetryableError checks if an error is retryable
func isSlackRetryableError(err error) bool {
	errStr := err.Error()

	// Network errors
	if strings.Contains(errStr, "connection refused") ||
		strings.Contains(errStr, "timeout") ||
		strings.Contains(errStr, "temporary") ||
		strings.Contains(errStr, "rate limited") ||
		strings.Contains(errStr, "500") ||
		strings.Contains(errStr, "502") ||
		strings.Contains(errStr, "503") ||
		strings.Contains(errStr, "504") {
		return true
	}

	return false
}
