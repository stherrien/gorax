package email

import (
	"context"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"

	"github.com/gorax/gorax/internal/communication"
)

// SendGridProvider implements EmailProvider using SendGrid API.
type SendGridProvider struct {
	client *sendgrid.Client
	apiKey string
}

// NewSendGridProvider creates a new SendGrid email provider.
func NewSendGridProvider(apiKey string) *SendGridProvider {
	return &SendGridProvider{
		client: sendgrid.NewSendClient(apiKey),
		apiKey: apiKey,
	}
}

// SendEmail sends a single email using SendGrid.
func (p *SendGridProvider) SendEmail(ctx context.Context, request *communication.EmailRequest) (*communication.EmailResponse, error) {
	if err := request.Validate(); err != nil {
		return nil, fmt.Errorf("invalid email request: %w", err)
	}

	message, err := p.buildMessage(request)
	if err != nil {
		return nil, fmt.Errorf("failed to build message: %w", err)
	}

	response, err := p.client.SendWithContext(ctx, message)
	if err != nil {
		return &communication.EmailResponse{
			Status: string(communication.MessageStatusFailed),
			Error:  err,
			SentAt: time.Now(),
		}, fmt.Errorf("failed to send email: %w", err)
	}

	if response.StatusCode >= 400 {
		return &communication.EmailResponse{
			Status: string(communication.MessageStatusFailed),
			Error:  fmt.Errorf("sendgrid returned error: %d - %s", response.StatusCode, response.Body),
			SentAt: time.Now(),
		}, fmt.Errorf("sendgrid API error: status=%d", response.StatusCode)
	}

	return &communication.EmailResponse{
		MessageID: response.Headers["X-Message-Id"][0],
		Status:    string(communication.MessageStatusSent),
		SentAt:    time.Now(),
	}, nil
}

// SendBulkEmail sends multiple emails using SendGrid.
func (p *SendGridProvider) SendBulkEmail(ctx context.Context, requests []*communication.EmailRequest) ([]*communication.EmailResponse, error) {
	responses := make([]*communication.EmailResponse, len(requests))
	var firstErr error

	for i, req := range requests {
		resp, err := p.SendEmail(ctx, req)
		responses[i] = resp
		if err != nil && firstErr == nil {
			firstErr = err
		}
	}

	return responses, firstErr
}

// buildMessage constructs a SendGrid mail message from an EmailRequest.
func (p *SendGridProvider) buildMessage(request *communication.EmailRequest) (*mail.SGMailV3, error) {
	from := mail.NewEmail("", request.From)
	subject := request.Subject

	message := mail.NewV3Mail()
	message.SetFrom(from)
	message.Subject = subject

	// Add personalization for recipients
	personalization := mail.NewPersonalization()

	for _, to := range request.To {
		personalization.AddTos(mail.NewEmail("", to))
	}

	for _, cc := range request.CC {
		personalization.AddCCs(mail.NewEmail("", cc))
	}

	for _, bcc := range request.BCC {
		personalization.AddBCCs(mail.NewEmail("", bcc))
	}

	message.AddPersonalizations(personalization)

	// Add content
	if request.Body != "" {
		message.AddContent(mail.NewContent("text/plain", request.Body))
	}

	if request.BodyHTML != "" {
		message.AddContent(mail.NewContent("text/html", request.BodyHTML))
	}

	// Add reply-to
	if request.ReplyTo != "" {
		message.SetReplyTo(mail.NewEmail("", request.ReplyTo))
	}

	// Add custom headers
	if len(request.Headers) > 0 {
		personalization.Headers = request.Headers
	}

	// Add attachments
	for _, att := range request.Attachments {
		if err := att.Validate(); err != nil {
			return nil, fmt.Errorf("invalid attachment: %w", err)
		}

		attachment := mail.NewAttachment()
		attachment.SetFilename(att.Filename)
		attachment.SetContent(base64.StdEncoding.EncodeToString(att.Content))
		attachment.SetType(att.ContentType)
		attachment.SetDisposition("attachment")

		message.AddAttachment(attachment)
	}

	return message, nil
}
