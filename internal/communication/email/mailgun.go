package email

import (
	"context"
	"fmt"
	"time"

	"github.com/gorax/gorax/internal/communication"
	"github.com/mailgun/mailgun-go/v4"
)

// MailgunProvider implements EmailProvider using Mailgun API.
type MailgunProvider struct {
	client mailgun.Mailgun
	domain string
}

// NewMailgunProvider creates a new Mailgun email provider.
func NewMailgunProvider(domain, apiKey string) *MailgunProvider {
	mg := mailgun.NewMailgun(domain, apiKey)

	return &MailgunProvider{
		client: mg,
		domain: domain,
	}
}

// SendEmail sends a single email using Mailgun.
func (p *MailgunProvider) SendEmail(ctx context.Context, request *communication.EmailRequest) (*communication.EmailResponse, error) {
	if err := request.Validate(); err != nil {
		return nil, fmt.Errorf("invalid email request: %w", err)
	}

	message := p.buildMessage(request)

	resp, id, err := p.client.Send(ctx, message)
	if err != nil {
		return &communication.EmailResponse{
			Status: string(communication.MessageStatusFailed),
			Error:  err,
			SentAt: time.Now(),
		}, fmt.Errorf("failed to send email: %w", err)
	}

	if resp != "Queued. Thank you." {
		return &communication.EmailResponse{
			MessageID: id,
			Status:    string(communication.MessageStatusFailed),
			Error:     fmt.Errorf("unexpected response: %s", resp),
			SentAt:    time.Now(),
		}, fmt.Errorf("mailgun error: %s", resp)
	}

	return &communication.EmailResponse{
		MessageID: id,
		Status:    string(communication.MessageStatusQueued),
		SentAt:    time.Now(),
	}, nil
}

// SendBulkEmail sends multiple emails using Mailgun.
func (p *MailgunProvider) SendBulkEmail(ctx context.Context, requests []*communication.EmailRequest) ([]*communication.EmailResponse, error) {
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

// buildMessage constructs a Mailgun message from an EmailRequest.
func (p *MailgunProvider) buildMessage(request *communication.EmailRequest) *mailgun.Message {
	message := p.client.NewMessage(
		request.From,
		request.Subject,
		request.Body,
		request.To...,
	)

	// Add CC recipients
	for _, cc := range request.CC {
		message.AddCC(cc)
	}

	// Add BCC recipients
	for _, bcc := range request.BCC {
		message.AddBCC(bcc)
	}

	// Add HTML body if present
	if request.BodyHTML != "" {
		message.SetHtml(request.BodyHTML)
	}

	// Add reply-to
	if request.ReplyTo != "" {
		message.SetReplyTo(request.ReplyTo)
	}

	// Add custom headers
	for key, value := range request.Headers {
		message.AddHeader(key, value)
	}

	// Add attachments
	for _, att := range request.Attachments {
		message.AddBufferAttachment(att.Filename, att.Content)
	}

	return message
}
