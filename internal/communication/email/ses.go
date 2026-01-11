package email

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"mime/multipart"
	"mime/quotedprintable"
	"net/textproto"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ses"

	"github.com/gorax/gorax/internal/communication"
)

// SESProvider implements EmailProvider using AWS SES.
type SESProvider struct {
	client *ses.SES
}

// NewSESProvider creates a new AWS SES email provider.
func NewSESProvider(region string) (*SESProvider, error) {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(region),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create AWS session: %w", err)
	}

	return &SESProvider{
		client: ses.New(sess),
	}, nil
}

// SendEmail sends a single email using AWS SES.
func (p *SESProvider) SendEmail(ctx context.Context, request *communication.EmailRequest) (*communication.EmailResponse, error) {
	if err := request.Validate(); err != nil {
		return nil, fmt.Errorf("invalid email request: %w", err)
	}

	// Use SendRawEmail if attachments are present
	if len(request.Attachments) > 0 {
		return p.sendRawEmail(ctx, request)
	}

	return p.sendSimpleEmail(ctx, request)
}

// SendBulkEmail sends multiple emails using AWS SES.
func (p *SESProvider) SendBulkEmail(ctx context.Context, requests []*communication.EmailRequest) ([]*communication.EmailResponse, error) {
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

// sendSimpleEmail sends a simple email without attachments.
func (p *SESProvider) sendSimpleEmail(ctx context.Context, request *communication.EmailRequest) (*communication.EmailResponse, error) {
	input := &ses.SendEmailInput{
		Source: aws.String(request.From),
		Destination: &ses.Destination{
			ToAddresses: aws.StringSlice(request.To),
		},
		Message: &ses.Message{
			Subject: &ses.Content{
				Charset: aws.String("UTF-8"),
				Data:    aws.String(request.Subject),
			},
		},
	}

	// Add CC recipients
	if len(request.CC) > 0 {
		input.Destination.CcAddresses = aws.StringSlice(request.CC)
	}

	// Add BCC recipients
	if len(request.BCC) > 0 {
		input.Destination.BccAddresses = aws.StringSlice(request.BCC)
	}

	// Add body content
	body := &ses.Body{}
	if request.Body != "" {
		body.Text = &ses.Content{
			Charset: aws.String("UTF-8"),
			Data:    aws.String(request.Body),
		}
	}
	if request.BodyHTML != "" {
		body.Html = &ses.Content{
			Charset: aws.String("UTF-8"),
			Data:    aws.String(request.BodyHTML),
		}
	}
	input.Message.Body = body

	// Add reply-to
	if request.ReplyTo != "" {
		input.ReplyToAddresses = aws.StringSlice([]string{request.ReplyTo})
	}

	result, err := p.client.SendEmailWithContext(ctx, input)
	if err != nil {
		return &communication.EmailResponse{
			Status: string(communication.MessageStatusFailed),
			Error:  err,
			SentAt: time.Now(),
		}, fmt.Errorf("failed to send email: %w", err)
	}

	return &communication.EmailResponse{
		MessageID: aws.StringValue(result.MessageId),
		Status:    string(communication.MessageStatusSent),
		SentAt:    time.Now(),
	}, nil
}

// sendRawEmail sends an email with attachments using raw email format.
func (p *SESProvider) sendRawEmail(ctx context.Context, request *communication.EmailRequest) (*communication.EmailResponse, error) {
	rawMessage, err := p.buildRawMessage(request)
	if err != nil {
		return nil, fmt.Errorf("failed to build raw message: %w", err)
	}

	input := &ses.SendRawEmailInput{
		Source: aws.String(request.From),
		Destinations: aws.StringSlice(append(
			append(request.To, request.CC...),
			request.BCC...,
		)),
		RawMessage: &ses.RawMessage{
			Data: rawMessage,
		},
	}

	result, err := p.client.SendRawEmailWithContext(ctx, input)
	if err != nil {
		return &communication.EmailResponse{
			Status: string(communication.MessageStatusFailed),
			Error:  err,
			SentAt: time.Now(),
		}, fmt.Errorf("failed to send raw email: %w", err)
	}

	return &communication.EmailResponse{
		MessageID: aws.StringValue(result.MessageId),
		Status:    string(communication.MessageStatusSent),
		SentAt:    time.Now(),
	}, nil
}

// buildRawMessage builds a MIME multipart message with attachments.
func (p *SESProvider) buildRawMessage(request *communication.EmailRequest) ([]byte, error) {
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

	// Create multipart writer
	writer := multipart.NewWriter(buf)
	buf.WriteString(fmt.Sprintf("Content-Type: multipart/mixed; boundary=%s\r\n\r\n", writer.Boundary()))

	// Write text body
	if request.Body != "" {
		textHeader := textproto.MIMEHeader{}
		textHeader.Set("Content-Type", "text/plain; charset=UTF-8")
		textHeader.Set("Content-Transfer-Encoding", "quoted-printable")

		textPart, err := writer.CreatePart(textHeader)
		if err != nil {
			return nil, fmt.Errorf("failed to create text part: %w", err)
		}

		qpWriter := quotedprintable.NewWriter(textPart)
		if _, err := qpWriter.Write([]byte(request.Body)); err != nil {
			return nil, fmt.Errorf("failed to write text body: %w", err)
		}
		if err := qpWriter.Close(); err != nil {
			return nil, fmt.Errorf("failed to close qp writer: %w", err)
		}
	}

	// Write HTML body
	if request.BodyHTML != "" {
		htmlHeader := textproto.MIMEHeader{}
		htmlHeader.Set("Content-Type", "text/html; charset=UTF-8")
		htmlHeader.Set("Content-Transfer-Encoding", "quoted-printable")

		htmlPart, err := writer.CreatePart(htmlHeader)
		if err != nil {
			return nil, fmt.Errorf("failed to create html part: %w", err)
		}

		qpWriter := quotedprintable.NewWriter(htmlPart)
		if _, err := qpWriter.Write([]byte(request.BodyHTML)); err != nil {
			return nil, fmt.Errorf("failed to write html body: %w", err)
		}
		if err := qpWriter.Close(); err != nil {
			return nil, fmt.Errorf("failed to close qp writer: %w", err)
		}
	}

	// Write attachments
	for _, att := range request.Attachments {
		if err := att.Validate(); err != nil {
			return nil, fmt.Errorf("invalid attachment: %w", err)
		}

		attHeader := textproto.MIMEHeader{}
		attHeader.Set("Content-Type", att.ContentType)
		attHeader.Set("Content-Transfer-Encoding", "base64")
		attHeader.Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", att.Filename))

		attPart, err := writer.CreatePart(attHeader)
		if err != nil {
			return nil, fmt.Errorf("failed to create attachment part: %w", err)
		}

		encoded := base64.StdEncoding.EncodeToString(att.Content)
		if _, err := attPart.Write([]byte(encoded)); err != nil {
			return nil, fmt.Errorf("failed to write attachment: %w", err)
		}
	}

	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("failed to close writer: %w", err)
	}

	return buf.Bytes(), nil
}
