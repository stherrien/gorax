package communication

import (
	"context"
	"encoding/base64"
	"fmt"

	"github.com/gorax/gorax/internal/communication"
	"github.com/gorax/gorax/internal/communication/email"
	"github.com/gorax/gorax/internal/credential"
)

// SendEmailAction executes email sending operations.
type SendEmailAction struct {
	config            SendEmailConfig
	credentialService credential.Service
}

// SendEmailConfig represents the configuration for the SendEmail action.
type SendEmailConfig struct {
	Provider     string              `json:"provider"` // sendgrid, mailgun, aws_ses, smtp
	From         string              `json:"from"`
	To           []string            `json:"to"`
	CC           []string            `json:"cc,omitempty"`
	BCC          []string            `json:"bcc,omitempty"`
	Subject      string              `json:"subject"`
	Body         string              `json:"body,omitempty"`
	BodyHTML     string              `json:"body_html,omitempty"`
	Attachments  []AttachmentConfig  `json:"attachments,omitempty"`
	ReplyTo      string              `json:"reply_to,omitempty"`
	Headers      map[string]string   `json:"headers,omitempty"`
	CredentialID string              `json:"credential_id"`
	SMTPConfig   *SMTPProviderConfig `json:"smtp_config,omitempty"` // Required for SMTP provider
}

// AttachmentConfig represents an email attachment configuration.
type AttachmentConfig struct {
	Filename    string `json:"filename"`
	Content     string `json:"content"` // Base64 encoded content
	ContentType string `json:"content_type"`
}

// SMTPProviderConfig contains SMTP-specific configuration.
type SMTPProviderConfig struct {
	Host   string `json:"host"`
	Port   int    `json:"port"`
	UseTLS bool   `json:"use_tls"`
}

// NewSendEmailAction creates a new SendEmail action.
func NewSendEmailAction(config SendEmailConfig, credService credential.Service) *SendEmailAction {
	return &SendEmailAction{
		config:            config,
		credentialService: credService,
	}
}

// Execute sends an email using the configured provider.
// TODO: This action needs integration with the workflow executor context to properly
// retrieve credentials with tenant/user information. The full implementation is ready
// but commented out until the execution context integration is complete.
func (a *SendEmailAction) Execute(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
	return nil, fmt.Errorf("email sending action requires execution context integration (see TODO in source)")

	/*
		// Full implementation (to be uncommented after execution context integration):

		// Get tenant and user from execution context
		tenantID := ctx.Value("tenantID").(string)
		userID := ctx.Value("userID").(string)

		// Get credential
		cred, err := a.credentialService.GetValue(ctx, tenantID, a.config.CredentialID, userID)
		if err != nil {
			return nil, fmt.Errorf("failed to get credential: %w", err)
		}

		// Create provider
		provider, err := a.createProvider(cred.Value)
		if err != nil {
			return nil, fmt.Errorf("failed to create email provider: %w", err)
		}

		// Build email request
		request, err := a.buildRequest()
		if err != nil {
			return nil, fmt.Errorf("failed to build email request: %w", err)
		}

		// Send email
		response, err := provider.SendEmail(ctx, request)
		if err != nil {
			return map[string]interface{}{
				"success":    false,
				"error":      err.Error(),
				"message_id": "",
			}, fmt.Errorf("failed to send email: %w", err)
		}

		return map[string]interface{}{
			"success":    true,
			"message_id": response.MessageID,
			"status":     response.Status,
			"sent_at":    response.SentAt,
		}, nil
	*/
}

// createProvider creates an email provider based on the configuration.
func (a *SendEmailAction) createProvider(credValue map[string]interface{}) (communication.EmailProvider, error) {
	switch a.config.Provider {
	case "sendgrid":
		apiKey, ok := credValue["api_key"].(string)
		if !ok {
			return nil, fmt.Errorf("sendgrid api_key not found in credential")
		}
		return email.NewSendGridProvider(apiKey), nil

	case "mailgun":
		domain, ok := credValue["domain"].(string)
		if !ok {
			return nil, fmt.Errorf("mailgun domain not found in credential")
		}
		apiKey, ok := credValue["api_key"].(string)
		if !ok {
			return nil, fmt.Errorf("mailgun api_key not found in credential")
		}
		return email.NewMailgunProvider(domain, apiKey), nil

	case "aws_ses":
		region, ok := credValue["region"].(string)
		if !ok {
			region = "us-east-1" // Default region
		}
		return email.NewSESProvider(region)

	case "smtp":
		if a.config.SMTPConfig == nil {
			return nil, fmt.Errorf("smtp_config is required for SMTP provider")
		}
		username, ok := credValue["username"].(string)
		if !ok {
			return nil, fmt.Errorf("smtp username not found in credential")
		}
		password, ok := credValue["password"].(string)
		if !ok {
			return nil, fmt.Errorf("smtp password not found in credential")
		}
		return email.NewSMTPProvider(
			a.config.SMTPConfig.Host,
			a.config.SMTPConfig.Port,
			username,
			password,
			a.config.SMTPConfig.UseTLS,
		), nil

	default:
		return nil, fmt.Errorf("unsupported email provider: %s", a.config.Provider)
	}
}

// buildRequest builds an EmailRequest from the action configuration.
func (a *SendEmailAction) buildRequest() (*communication.EmailRequest, error) {
	request := &communication.EmailRequest{
		From:     a.config.From,
		To:       a.config.To,
		CC:       a.config.CC,
		BCC:      a.config.BCC,
		Subject:  a.config.Subject,
		Body:     a.config.Body,
		BodyHTML: a.config.BodyHTML,
		ReplyTo:  a.config.ReplyTo,
		Headers:  a.config.Headers,
	}

	// Parse attachments
	for _, attConfig := range a.config.Attachments {
		content, err := base64.StdEncoding.DecodeString(attConfig.Content)
		if err != nil {
			return nil, fmt.Errorf("failed to decode attachment %s: %w", attConfig.Filename, err)
		}

		request.Attachments = append(request.Attachments, communication.Attachment{
			Filename:    attConfig.Filename,
			Content:     content,
			ContentType: attConfig.ContentType,
		})
	}

	return request, nil
}

// Name returns the action name.
func (a *SendEmailAction) Name() string {
	return "send_email"
}

// Validate validates the action configuration.
func (a *SendEmailAction) Validate() error {
	if a.config.Provider == "" {
		return fmt.Errorf("provider is required")
	}
	if a.config.From == "" {
		return fmt.Errorf("from address is required")
	}
	if len(a.config.To) == 0 {
		return fmt.Errorf("at least one recipient is required")
	}
	if a.config.Subject == "" {
		return fmt.Errorf("subject is required")
	}
	if a.config.Body == "" && a.config.BodyHTML == "" {
		return fmt.Errorf("email body is required")
	}
	if a.config.CredentialID == "" {
		return fmt.Errorf("credential_id is required")
	}
	if a.config.Provider == "smtp" && a.config.SMTPConfig == nil {
		return fmt.Errorf("smtp_config is required for SMTP provider")
	}
	return nil
}
