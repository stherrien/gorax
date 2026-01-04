// Package communication provides interfaces and types for email and SMS communication providers.
package communication

import (
	"context"
	"fmt"
	"time"
)

// EmailProvider defines the interface for email service providers.
type EmailProvider interface {
	SendEmail(ctx context.Context, request *EmailRequest) (*EmailResponse, error)
	SendBulkEmail(ctx context.Context, requests []*EmailRequest) ([]*EmailResponse, error)
}

// SMSProvider defines the interface for SMS service providers.
type SMSProvider interface {
	SendSMS(ctx context.Context, request *SMSRequest) (*SMSResponse, error)
	SendBulkSMS(ctx context.Context, requests []*SMSRequest) ([]*SMSResponse, error)
}

// EmailRequest represents a request to send an email.
type EmailRequest struct {
	From        string
	To          []string
	CC          []string
	BCC         []string
	Subject     string
	Body        string
	BodyHTML    string
	Attachments []Attachment
	ReplyTo     string
	Headers     map[string]string
}

// Validate checks if the email request is valid.
func (r *EmailRequest) Validate() error {
	if r.From == "" {
		return fmt.Errorf("from address is required")
	}
	if len(r.To) == 0 {
		return fmt.Errorf("at least one recipient is required")
	}
	if r.Subject == "" {
		return fmt.Errorf("subject is required")
	}
	if r.Body == "" && r.BodyHTML == "" {
		return fmt.Errorf("email body is required")
	}
	return nil
}

// EmailResponse represents the response from sending an email.
type EmailResponse struct {
	MessageID string
	Status    string
	Error     error
	SentAt    time.Time
}

// SMSRequest represents a request to send an SMS.
type SMSRequest struct {
	From    string
	To      string
	Message string
}

// Validate checks if the SMS request is valid.
func (r *SMSRequest) Validate() error {
	if r.From == "" {
		return fmt.Errorf("from number is required")
	}
	if r.To == "" {
		return fmt.Errorf("to number is required")
	}
	if r.Message == "" {
		return fmt.Errorf("message is required")
	}
	if len(r.Message) > 1600 {
		return fmt.Errorf("message exceeds maximum length of 1600 characters")
	}
	return nil
}

// SMSResponse represents the response from sending an SMS.
type SMSResponse struct {
	MessageID string
	Status    string
	Cost      float64
	Error     error
	SentAt    time.Time
}

// Attachment represents an email attachment.
type Attachment struct {
	Filename    string
	Content     []byte
	ContentType string
}

// Validate checks if the attachment is valid.
func (a *Attachment) Validate() error {
	if a.Filename == "" {
		return fmt.Errorf("attachment filename is required")
	}
	if len(a.Content) == 0 {
		return fmt.Errorf("attachment content is required")
	}
	if a.ContentType == "" {
		return fmt.Errorf("attachment content type is required")
	}
	if len(a.Content) > 10*1024*1024 {
		return fmt.Errorf("attachment size exceeds maximum of 10MB")
	}
	return nil
}

// ProviderType represents the type of communication provider.
type ProviderType string

const (
	ProviderTypeSendGrid    ProviderType = "sendgrid"
	ProviderTypeMailgun     ProviderType = "mailgun"
	ProviderTypeAWSSES      ProviderType = "aws_ses"
	ProviderTypeSMTP        ProviderType = "smtp"
	ProviderTypeTwilio      ProviderType = "twilio"
	ProviderTypeAWSSNS      ProviderType = "aws_sns"
	ProviderTypeMessageBird ProviderType = "messagebird"
)

// CommunicationType represents the type of communication.
type CommunicationType string

const (
	CommunicationTypeEmail CommunicationType = "email"
	CommunicationTypeSMS   CommunicationType = "sms"
)

// MessageStatus represents the status of a sent message.
type MessageStatus string

const (
	MessageStatusSent      MessageStatus = "sent"
	MessageStatusFailed    MessageStatus = "failed"
	MessageStatusQueued    MessageStatus = "queued"
	MessageStatusDelivered MessageStatus = "delivered"
)
