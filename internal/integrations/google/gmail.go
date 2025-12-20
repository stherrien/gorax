package google

import (
	"context"
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/gorax/gorax/internal/credential"
	"github.com/gorax/gorax/internal/executor/actions"
	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"
)

const (
	gmailSendScope = "https://www.googleapis.com/auth/gmail.send"
	gmailReadScope = "https://www.googleapis.com/auth/gmail.readonly"
)

// GmailSendAction implements the Gmail Send action
type GmailSendAction struct {
	credentialService credential.Service
	baseURL           string // For testing
}

// GmailSendConfig defines the configuration for sending an email
type GmailSendConfig struct {
	To          string   `json:"to"`
	Cc          string   `json:"cc,omitempty"`
	Bcc         string   `json:"bcc,omitempty"`
	Subject     string   `json:"subject"`
	Body        string   `json:"body"`
	IsHTML      bool     `json:"is_html,omitempty"`
	Attachments []string `json:"attachments,omitempty"` // Base64 encoded file data
}

// GmailSendResult represents the result of sending an email
type GmailSendResult struct {
	MessageID string `json:"message_id"`
	ThreadID  string `json:"thread_id"`
}

// Validate validates the Gmail send configuration
func (c *GmailSendConfig) Validate() error {
	if c.To == "" {
		return fmt.Errorf("to is required")
	}
	if c.Subject == "" {
		return fmt.Errorf("subject is required")
	}
	if c.Body == "" {
		return fmt.Errorf("body is required")
	}
	return nil
}

// NewGmailSendAction creates a new Gmail send action
func NewGmailSendAction(credentialService credential.Service) *GmailSendAction {
	return &GmailSendAction{
		credentialService: credentialService,
	}
}

// Execute implements the Action interface
func (a *GmailSendAction) Execute(ctx context.Context, input *actions.ActionInput) (*actions.ActionOutput, error) {
	// Parse config
	config, ok := input.Config.(GmailSendConfig)
	if !ok {
		return nil, fmt.Errorf("invalid config type: expected GmailSendConfig")
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

	// Create OAuth2 token
	token, err := createOAuth2Token(decryptedCred.Value)
	if err != nil {
		return nil, err
	}

	// Create Gmail service
	var gmailService *gmail.Service
	if a.baseURL != "" {
		// For testing with mock server
		gmailService, err = gmail.NewService(ctx, option.WithoutAuthentication(), option.WithEndpoint(a.baseURL))
	} else {
		gmailService, err = gmail.NewService(ctx, createOAuth2Client(ctx, token))
	}
	if err != nil {
		return nil, fmt.Errorf("failed to create Gmail service: %w", err)
	}

	// Create email message
	message, err := createEmailMessage(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create email message: %w", err)
	}

	// Send email
	sentMsg, err := gmailService.Users.Messages.Send("me", message).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to send email: %w", err)
	}

	// Build result
	result := &GmailSendResult{
		MessageID: sentMsg.Id,
		ThreadID:  sentMsg.ThreadId,
	}

	// Create output
	output := actions.NewActionOutput(result)
	output.WithMetadata("message_id", sentMsg.Id)
	output.WithMetadata("thread_id", sentMsg.ThreadId)

	return output, nil
}

// createEmailMessage creates a Gmail message from the config
func createEmailMessage(config GmailSendConfig) (*gmail.Message, error) {
	// Build email headers
	var headers strings.Builder
	headers.WriteString(fmt.Sprintf("To: %s\r\n", config.To))
	if config.Cc != "" {
		headers.WriteString(fmt.Sprintf("Cc: %s\r\n", config.Cc))
	}
	if config.Bcc != "" {
		headers.WriteString(fmt.Sprintf("Bcc: %s\r\n", config.Bcc))
	}
	headers.WriteString(fmt.Sprintf("Subject: %s\r\n", config.Subject))

	if config.IsHTML {
		headers.WriteString("Content-Type: text/html; charset=UTF-8\r\n")
	} else {
		headers.WriteString("Content-Type: text/plain; charset=UTF-8\r\n")
	}
	headers.WriteString("\r\n")
	headers.WriteString(config.Body)

	// Encode message in base64
	rawMessage := base64.URLEncoding.EncodeToString([]byte(headers.String()))

	return &gmail.Message{
		Raw: rawMessage,
	}, nil
}

// GmailReadAction implements the Gmail Read action
type GmailReadAction struct {
	credentialService credential.Service
	baseURL           string // For testing
}

// GmailReadConfig defines the configuration for reading emails
type GmailReadConfig struct {
	Query      string `json:"query"`
	MaxResults int64  `json:"max_results,omitempty"`
}

// GmailMessage represents an email message
type GmailMessage struct {
	ID       string            `json:"id"`
	ThreadID string            `json:"thread_id"`
	Snippet  string            `json:"snippet"`
	From     string            `json:"from"`
	To       string            `json:"to"`
	Subject  string            `json:"subject"`
	Date     string            `json:"date"`
	Body     string            `json:"body"`
	Labels   []string          `json:"labels"`
	Headers  map[string]string `json:"headers"`
}

// GmailReadResult represents the result of reading emails
type GmailReadResult struct {
	Messages []GmailMessage `json:"messages"`
	Count    int            `json:"count"`
}

// NewGmailReadAction creates a new Gmail read action
func NewGmailReadAction(credentialService credential.Service) *GmailReadAction {
	return &GmailReadAction{
		credentialService: credentialService,
	}
}

// Execute implements the Action interface
func (a *GmailReadAction) Execute(ctx context.Context, input *actions.ActionInput) (*actions.ActionOutput, error) {
	// Parse config
	config, ok := input.Config.(GmailReadConfig)
	if !ok {
		return nil, fmt.Errorf("invalid config type: expected GmailReadConfig")
	}

	// Set default max results
	if config.MaxResults == 0 {
		config.MaxResults = 10
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

	// Create OAuth2 token
	token, err := createOAuth2Token(decryptedCred.Value)
	if err != nil {
		return nil, err
	}

	// Create Gmail service
	var gmailService *gmail.Service
	if a.baseURL != "" {
		gmailService, err = gmail.NewService(ctx, option.WithoutAuthentication(), option.WithEndpoint(a.baseURL))
	} else {
		gmailService, err = gmail.NewService(ctx, createOAuth2Client(ctx, token))
	}
	if err != nil {
		return nil, fmt.Errorf("failed to create Gmail service: %w", err)
	}

	// List messages
	listCall := gmailService.Users.Messages.List("me")
	if config.Query != "" {
		listCall = listCall.Q(config.Query)
	}
	listCall = listCall.MaxResults(config.MaxResults)

	response, err := listCall.Do()
	if err != nil {
		return nil, fmt.Errorf("failed to list emails: %w", err)
	}

	// Parse messages
	messages := make([]GmailMessage, 0, len(response.Messages))
	for _, msg := range response.Messages {
		// Get full message details
		fullMsg, err := gmailService.Users.Messages.Get("me", msg.Id).Do()
		if err != nil {
			// Skip messages that fail to fetch
			continue
		}

		gmailMsg := parseGmailMessage(fullMsg)
		messages = append(messages, gmailMsg)
	}

	// Build result
	result := &GmailReadResult{
		Messages: messages,
		Count:    len(messages),
	}

	// Create output
	output := actions.NewActionOutput(result)
	output.WithMetadata("count", len(messages))

	return output, nil
}

// parseGmailMessage parses a Gmail message into our simplified format
func parseGmailMessage(msg *gmail.Message) GmailMessage {
	gmailMsg := GmailMessage{
		ID:       msg.Id,
		ThreadID: msg.ThreadId,
		Snippet:  msg.Snippet,
		Labels:   msg.LabelIds,
		Headers:  make(map[string]string),
	}

	// Parse headers if payload exists
	if msg.Payload != nil && msg.Payload.Headers != nil {
		for _, header := range msg.Payload.Headers {
			gmailMsg.Headers[header.Name] = header.Value
			switch header.Name {
			case "From":
				gmailMsg.From = header.Value
			case "To":
				gmailMsg.To = header.Value
			case "Subject":
				gmailMsg.Subject = header.Value
			case "Date":
				gmailMsg.Date = header.Value
			}
		}

		// Extract body
		if msg.Payload.Body != nil && msg.Payload.Body.Data != "" {
			data, err := base64.URLEncoding.DecodeString(msg.Payload.Body.Data)
			if err == nil {
				gmailMsg.Body = string(data)
			}
		}
	}

	return gmailMsg
}
