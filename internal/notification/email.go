package notification

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"net/smtp"
	"strings"
	"text/template"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ses"
	"github.com/aws/aws-sdk-go-v2/service/ses/types"
)

// EmailProvider represents email delivery provider
type EmailProvider string

const (
	EmailProviderSMTP EmailProvider = "smtp"
	EmailProviderSES  EmailProvider = "ses"
)

// EmailConfig holds email configuration
type EmailConfig struct {
	Provider EmailProvider
	From     string

	// SMTP settings
	SMTPHost string
	SMTPPort int
	SMTPUser string
	SMTPPass string
	TLS      bool

	// AWS SES settings
	AWSRegion          string
	AWSAccessKeyID     string
	AWSSecretAccessKey string

	// Retry settings
	MaxRetries int
	RetryDelay time.Duration
}

// Email represents an email message
type Email struct {
	To       []string
	CC       []string
	BCC      []string
	Subject  string
	HTMLBody string
	TextBody string
}

// EmailTemplate represents an email template
type EmailTemplate struct {
	Subject  string
	HTMLBody string
	TextBody string
}

// EmailSender sends email notifications
type EmailSender struct {
	config    EmailConfig
	sesClient *ses.Client
}

// NewEmailSender creates a new email sender
func NewEmailSender(cfg EmailConfig) (*EmailSender, error) {
	if err := validateEmailConfig(cfg); err != nil {
		return nil, err
	}

	sender := &EmailSender{
		config: cfg,
	}

	// Initialize SES client if using SES
	if cfg.Provider == EmailProviderSES {
		awsCfg, err := config.LoadDefaultConfig(context.Background(),
			config.WithRegion(cfg.AWSRegion),
		)
		if err != nil {
			return nil, fmt.Errorf("failed to load AWS config: %w", err)
		}

		sender.sesClient = ses.NewFromConfig(awsCfg)
	}

	return sender, nil
}

// Send sends an email
func (s *EmailSender) Send(ctx context.Context, email Email) error {
	// Check context
	if err := ctx.Err(); err != nil {
		return err
	}

	// Validate email
	if err := validateEmail(email); err != nil {
		return err
	}

	switch s.config.Provider {
	case EmailProviderSMTP:
		return s.sendViaSMTP(ctx, email)
	case EmailProviderSES:
		return s.sendViaSES(ctx, email)
	default:
		return fmt.Errorf("unsupported email provider: %s", s.config.Provider)
	}
}

// sendViaSMTP sends email via SMTP with retry logic
func (s *EmailSender) sendViaSMTP(ctx context.Context, email Email) error {
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

		err := s.sendSMTPOnce(email)
		if err == nil {
			return nil
		}

		lastErr = err

		// Don't retry on last attempt
		if attempt == maxRetries {
			break
		}

		// Check if error is retryable
		if !isRetryableError(err) {
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

	return fmt.Errorf("failed to send email after %d attempts: %w", maxRetries+1, lastErr)
}

// sendSMTPOnce performs a single SMTP send attempt
func (s *EmailSender) sendSMTPOnce(email Email) error {
	// Build message
	msg := s.buildMessage(email)

	// SMTP address
	addr := fmt.Sprintf("%s:%d", s.config.SMTPHost, s.config.SMTPPort)

	// Prepare auth
	var auth smtp.Auth
	if s.config.SMTPUser != "" && s.config.SMTPPass != "" {
		auth = smtp.PlainAuth("", s.config.SMTPUser, s.config.SMTPPass, s.config.SMTPHost)
	}

	// Build recipient list
	recipients := append([]string{}, email.To...)
	recipients = append(recipients, email.CC...)
	recipients = append(recipients, email.BCC...)

	// Send with or without TLS
	if s.config.TLS {
		return s.sendWithTLS(addr, auth, recipients, msg)
	}

	return smtp.SendMail(addr, auth, s.config.From, recipients, msg)
}

// sendWithTLS sends email using explicit TLS
func (s *EmailSender) sendWithTLS(addr string, auth smtp.Auth, recipients []string, msg []byte) error {
	// Connect to server
	conn, err := tls.Dial("tcp", addr, &tls.Config{
		ServerName: s.config.SMTPHost,
	})
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	defer conn.Close()

	// Create SMTP client
	client, err := smtp.NewClient(conn, s.config.SMTPHost)
	if err != nil {
		return fmt.Errorf("failed to create SMTP client: %w", err)
	}
	defer client.Quit()

	// Authenticate
	if auth != nil {
		if err := client.Auth(auth); err != nil {
			return fmt.Errorf("authentication failed: %w", err)
		}
	}

	// Set sender
	if err := client.Mail(s.config.From); err != nil {
		return fmt.Errorf("failed to set sender: %w", err)
	}

	// Set recipients
	for _, recipient := range recipients {
		if err := client.Rcpt(recipient); err != nil {
			return fmt.Errorf("failed to add recipient %s: %w", recipient, err)
		}
	}

	// Send data
	writer, err := client.Data()
	if err != nil {
		return fmt.Errorf("failed to create data writer: %w", err)
	}

	_, err = writer.Write(msg)
	if err != nil {
		writer.Close()
		return fmt.Errorf("failed to write message: %w", err)
	}

	err = writer.Close()
	if err != nil {
		return fmt.Errorf("failed to close data writer: %w", err)
	}

	return nil
}

// sendViaSES sends email via AWS SES
func (s *EmailSender) sendViaSES(ctx context.Context, email Email) error {
	input := &ses.SendEmailInput{
		Source: aws.String(s.config.From),
		Destination: &types.Destination{
			ToAddresses:  email.To,
			CcAddresses:  email.CC,
			BccAddresses: email.BCC,
		},
		Message: &types.Message{
			Subject: &types.Content{
				Data:    aws.String(email.Subject),
				Charset: aws.String("UTF-8"),
			},
			Body: &types.Body{},
		},
	}

	if email.HTMLBody != "" {
		input.Message.Body.Html = &types.Content{
			Data:    aws.String(email.HTMLBody),
			Charset: aws.String("UTF-8"),
		}
	}

	if email.TextBody != "" {
		input.Message.Body.Text = &types.Content{
			Data:    aws.String(email.TextBody),
			Charset: aws.String("UTF-8"),
		}
	}

	_, err := s.sesClient.SendEmail(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to send email via SES: %w", err)
	}

	return nil
}

// buildMessage builds MIME email message
func (s *EmailSender) buildMessage(email Email) []byte {
	var buf bytes.Buffer

	// Headers
	buf.WriteString(fmt.Sprintf("From: %s\r\n", s.config.From))
	buf.WriteString(fmt.Sprintf("To: %s\r\n", strings.Join(email.To, ", ")))

	if len(email.CC) > 0 {
		buf.WriteString(fmt.Sprintf("Cc: %s\r\n", strings.Join(email.CC, ", ")))
	}

	buf.WriteString(fmt.Sprintf("Subject: %s\r\n", email.Subject))
	buf.WriteString("MIME-Version: 1.0\r\n")

	// Multipart if both HTML and text
	if email.HTMLBody != "" && email.TextBody != "" {
		boundary := "----=_Part_0_12345678.12345678"
		buf.WriteString(fmt.Sprintf("Content-Type: multipart/alternative; boundary=\"%s\"\r\n\r\n", boundary))

		// Text part
		buf.WriteString(fmt.Sprintf("--%s\r\n", boundary))
		buf.WriteString("Content-Type: text/plain; charset=UTF-8\r\n\r\n")
		buf.WriteString(email.TextBody)
		buf.WriteString("\r\n\r\n")

		// HTML part
		buf.WriteString(fmt.Sprintf("--%s\r\n", boundary))
		buf.WriteString("Content-Type: text/html; charset=UTF-8\r\n\r\n")
		buf.WriteString(email.HTMLBody)
		buf.WriteString("\r\n\r\n")

		buf.WriteString(fmt.Sprintf("--%s--\r\n", boundary))
	} else if email.HTMLBody != "" {
		buf.WriteString("Content-Type: text/html; charset=UTF-8\r\n\r\n")
		buf.WriteString(email.HTMLBody)
	} else {
		buf.WriteString("Content-Type: text/plain; charset=UTF-8\r\n\r\n")
		buf.WriteString(email.TextBody)
	}

	return buf.Bytes()
}

// RenderTemplate renders an email template with data
func RenderTemplate(tmpl EmailTemplate, data map[string]interface{}) (Email, error) {
	email := Email{}

	// Render subject
	subjectTmpl, err := template.New("subject").Parse(tmpl.Subject)
	if err != nil {
		return email, fmt.Errorf("failed to parse subject template: %w", err)
	}

	var subjectBuf bytes.Buffer
	if err := subjectTmpl.Execute(&subjectBuf, data); err != nil {
		return email, fmt.Errorf("failed to render subject: %w", err)
	}
	email.Subject = subjectBuf.String()

	// Render HTML body
	if tmpl.HTMLBody != "" {
		htmlTmpl, err := template.New("html").Parse(tmpl.HTMLBody)
		if err != nil {
			return email, fmt.Errorf("failed to parse HTML template: %w", err)
		}

		var htmlBuf bytes.Buffer
		if err := htmlTmpl.Execute(&htmlBuf, data); err != nil {
			return email, fmt.Errorf("failed to render HTML body: %w", err)
		}
		email.HTMLBody = htmlBuf.String()
	}

	// Render text body
	if tmpl.TextBody != "" {
		textTmpl, err := template.New("text").Parse(tmpl.TextBody)
		if err != nil {
			return email, fmt.Errorf("failed to parse text template: %w", err)
		}

		var textBuf bytes.Buffer
		if err := textTmpl.Execute(&textBuf, data); err != nil {
			return email, fmt.Errorf("failed to render text body: %w", err)
		}
		email.TextBody = textBuf.String()
	}

	return email, nil
}

// validateEmailConfig validates email configuration
func validateEmailConfig(cfg EmailConfig) error {
	if cfg.From == "" {
		return fmt.Errorf("from address is required")
	}

	switch cfg.Provider {
	case EmailProviderSMTP:
		if cfg.SMTPHost == "" {
			return fmt.Errorf("SMTP host is required")
		}
		if cfg.SMTPPort == 0 {
			return fmt.Errorf("SMTP port is required")
		}
	case EmailProviderSES:
		if cfg.AWSRegion == "" {
			return fmt.Errorf("AWS region is required for SES")
		}
	default:
		return fmt.Errorf("unsupported email provider: %s", cfg.Provider)
	}

	return nil
}

// validateEmail validates email message
func validateEmail(email Email) error {
	if len(email.To) == 0 {
		return fmt.Errorf("at least one recipient is required")
	}

	if email.Subject == "" {
		return fmt.Errorf("subject is required")
	}

	if email.HTMLBody == "" && email.TextBody == "" {
		return fmt.Errorf("email body is required (HTML or text)")
	}

	return nil
}

// isRetryableError checks if an error is retryable
func isRetryableError(err error) bool {
	errStr := err.Error()

	// Network errors
	if strings.Contains(errStr, "connection refused") ||
		strings.Contains(errStr, "timeout") ||
		strings.Contains(errStr, "temporary failure") ||
		strings.Contains(errStr, "i/o timeout") {
		return true
	}

	return false
}
