package email

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"mime/multipart"
	"mime/quotedprintable"
	"net/smtp"
	"net/textproto"
	"strings"
	"time"

	"github.com/gorax/gorax/internal/communication"
)

// SMTPProvider implements EmailProvider using SMTP protocol.
type SMTPProvider struct {
	host     string
	port     int
	username string
	password string
	useTLS   bool
}

// NewSMTPProvider creates a new SMTP email provider.
func NewSMTPProvider(host string, port int, username, password string, useTLS bool) *SMTPProvider {
	return &SMTPProvider{
		host:     host,
		port:     port,
		username: username,
		password: password,
		useTLS:   useTLS,
	}
}

// SendEmail sends a single email using SMTP.
func (p *SMTPProvider) SendEmail(ctx context.Context, request *communication.EmailRequest) (*communication.EmailResponse, error) {
	if err := request.Validate(); err != nil {
		return nil, fmt.Errorf("invalid email request: %w", err)
	}

	message, err := p.buildMessage(request)
	if err != nil {
		return nil, fmt.Errorf("failed to build message: %w", err)
	}

	addr := fmt.Sprintf("%s:%d", p.host, p.port)
	auth := smtp.PlainAuth("", p.username, p.password, p.host)

	recipients := append(request.To, request.CC...)
	recipients = append(recipients, request.BCC...)

	var sendErr error
	if p.useTLS {
		sendErr = p.sendWithTLS(addr, auth, request.From, recipients, message)
	} else {
		sendErr = smtp.SendMail(addr, auth, request.From, recipients, message)
	}

	if sendErr != nil {
		return &communication.EmailResponse{
			Status: string(communication.MessageStatusFailed),
			Error:  sendErr,
			SentAt: time.Now(),
		}, fmt.Errorf("failed to send email: %w", sendErr)
	}

	return &communication.EmailResponse{
		MessageID: fmt.Sprintf("%d@%s", time.Now().UnixNano(), p.host),
		Status:    string(communication.MessageStatusSent),
		SentAt:    time.Now(),
	}, nil
}

// SendBulkEmail sends multiple emails using SMTP.
func (p *SMTPProvider) SendBulkEmail(ctx context.Context, requests []*communication.EmailRequest) ([]*communication.EmailResponse, error) {
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

// sendWithTLS sends email using explicit TLS.
func (p *SMTPProvider) sendWithTLS(addr string, auth smtp.Auth, from string, to []string, msg []byte) error {
	tlsConfig := &tls.Config{
		ServerName: p.host,
		MinVersion: tls.VersionTLS12,
	}

	conn, err := tls.Dial("tcp", addr, tlsConfig)
	if err != nil {
		return fmt.Errorf("failed to dial TLS: %w", err)
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, p.host)
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}
	defer client.Close()

	if err := client.Auth(auth); err != nil {
		return fmt.Errorf("authentication failed: %w", err)
	}

	if err := client.Mail(from); err != nil {
		return fmt.Errorf("failed to set sender: %w", err)
	}

	for _, recipient := range to {
		if err := client.Rcpt(recipient); err != nil {
			return fmt.Errorf("failed to set recipient: %w", err)
		}
	}

	writer, err := client.Data()
	if err != nil {
		return fmt.Errorf("failed to get data writer: %w", err)
	}

	if _, err := writer.Write(msg); err != nil {
		return fmt.Errorf("failed to write message: %w", err)
	}

	if err := writer.Close(); err != nil {
		return fmt.Errorf("failed to close writer: %w", err)
	}

	return client.Quit()
}

// buildMessage builds an RFC 5322 compliant email message.
func (p *SMTPProvider) buildMessage(request *communication.EmailRequest) ([]byte, error) {
	buf := new(bytes.Buffer)

	// Write headers
	buf.WriteString(fmt.Sprintf("From: %s\r\n", request.From))
	buf.WriteString(fmt.Sprintf("To: %s\r\n", strings.Join(request.To, ", ")))
	if len(request.CC) > 0 {
		buf.WriteString(fmt.Sprintf("Cc: %s\r\n", strings.Join(request.CC, ", ")))
	}
	if request.ReplyTo != "" {
		buf.WriteString(fmt.Sprintf("Reply-To: %s\r\n", request.ReplyTo))
	}
	buf.WriteString(fmt.Sprintf("Subject: %s\r\n", request.Subject))

	// Add custom headers
	for key, value := range request.Headers {
		buf.WriteString(fmt.Sprintf("%s: %s\r\n", key, value))
	}

	buf.WriteString("MIME-Version: 1.0\r\n")

	// Create multipart if needed
	if len(request.Attachments) > 0 || (request.Body != "" && request.BodyHTML != "") {
		writer := multipart.NewWriter(buf)
		buf.WriteString(fmt.Sprintf("Content-Type: multipart/mixed; boundary=%s\r\n\r\n", writer.Boundary()))

		// Write text body
		if request.Body != "" {
			if err := p.writeTextPart(writer, request.Body); err != nil {
				return nil, err
			}
		}

		// Write HTML body
		if request.BodyHTML != "" {
			if err := p.writeHTMLPart(writer, request.BodyHTML); err != nil {
				return nil, err
			}
		}

		// Write attachments
		for _, att := range request.Attachments {
			if err := att.Validate(); err != nil {
				return nil, fmt.Errorf("invalid attachment: %w", err)
			}
			if err := p.writeAttachment(writer, att); err != nil {
				return nil, err
			}
		}

		if err := writer.Close(); err != nil {
			return nil, fmt.Errorf("failed to close writer: %w", err)
		}
	} else {
		// Simple message without multipart
		if request.BodyHTML != "" {
			buf.WriteString("Content-Type: text/html; charset=UTF-8\r\n")
			buf.WriteString("Content-Transfer-Encoding: quoted-printable\r\n\r\n")
			buf.WriteString(request.BodyHTML)
		} else {
			buf.WriteString("Content-Type: text/plain; charset=UTF-8\r\n")
			buf.WriteString("Content-Transfer-Encoding: quoted-printable\r\n\r\n")
			buf.WriteString(request.Body)
		}
	}

	return buf.Bytes(), nil
}

func (p *SMTPProvider) writeTextPart(writer *multipart.Writer, text string) error {
	header := textproto.MIMEHeader{}
	header.Set("Content-Type", "text/plain; charset=UTF-8")
	header.Set("Content-Transfer-Encoding", "quoted-printable")

	part, err := writer.CreatePart(header)
	if err != nil {
		return fmt.Errorf("failed to create text part: %w", err)
	}

	qpWriter := quotedprintable.NewWriter(part)
	if _, err := qpWriter.Write([]byte(text)); err != nil {
		return fmt.Errorf("failed to write text: %w", err)
	}
	return qpWriter.Close()
}

func (p *SMTPProvider) writeHTMLPart(writer *multipart.Writer, html string) error {
	header := textproto.MIMEHeader{}
	header.Set("Content-Type", "text/html; charset=UTF-8")
	header.Set("Content-Transfer-Encoding", "quoted-printable")

	part, err := writer.CreatePart(header)
	if err != nil {
		return fmt.Errorf("failed to create html part: %w", err)
	}

	qpWriter := quotedprintable.NewWriter(part)
	if _, err := qpWriter.Write([]byte(html)); err != nil {
		return fmt.Errorf("failed to write html: %w", err)
	}
	return qpWriter.Close()
}

func (p *SMTPProvider) writeAttachment(writer *multipart.Writer, att communication.Attachment) error {
	header := textproto.MIMEHeader{}
	header.Set("Content-Type", att.ContentType)
	header.Set("Content-Transfer-Encoding", "base64")
	header.Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", att.Filename))

	part, err := writer.CreatePart(header)
	if err != nil {
		return fmt.Errorf("failed to create attachment part: %w", err)
	}

	encoded := base64.StdEncoding.EncodeToString(att.Content)
	if _, err := part.Write([]byte(encoded)); err != nil {
		return fmt.Errorf("failed to write attachment: %w", err)
	}

	return nil
}
