package communication

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEmailRequest_Validate(t *testing.T) {
	tests := []struct {
		name    string
		request *EmailRequest
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid email request",
			request: &EmailRequest{
				From:    "sender@example.com",
				To:      []string{"recipient@example.com"},
				Subject: "Test Subject",
				Body:    "Test body",
			},
			wantErr: false,
		},
		{
			name: "valid with html body only",
			request: &EmailRequest{
				From:     "sender@example.com",
				To:       []string{"recipient@example.com"},
				Subject:  "Test Subject",
				BodyHTML: "<p>Test body</p>",
			},
			wantErr: false,
		},
		{
			name: "missing from address",
			request: &EmailRequest{
				To:      []string{"recipient@example.com"},
				Subject: "Test Subject",
				Body:    "Test body",
			},
			wantErr: true,
			errMsg:  "from address is required",
		},
		{
			name: "missing recipients",
			request: &EmailRequest{
				From:    "sender@example.com",
				Subject: "Test Subject",
				Body:    "Test body",
			},
			wantErr: true,
			errMsg:  "at least one recipient is required",
		},
		{
			name: "missing subject",
			request: &EmailRequest{
				From: "sender@example.com",
				To:   []string{"recipient@example.com"},
				Body: "Test body",
			},
			wantErr: true,
			errMsg:  "subject is required",
		},
		{
			name: "missing body",
			request: &EmailRequest{
				From:    "sender@example.com",
				To:      []string{"recipient@example.com"},
				Subject: "Test Subject",
			},
			wantErr: true,
			errMsg:  "email body is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.request.Validate()
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestSMSRequest_Validate(t *testing.T) {
	tests := []struct {
		name    string
		request *SMSRequest
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid sms request",
			request: &SMSRequest{
				From:    "+1234567890",
				To:      "+0987654321",
				Message: "Test message",
			},
			wantErr: false,
		},
		{
			name: "missing from number",
			request: &SMSRequest{
				To:      "+0987654321",
				Message: "Test message",
			},
			wantErr: true,
			errMsg:  "from number is required",
		},
		{
			name: "missing to number",
			request: &SMSRequest{
				From:    "+1234567890",
				Message: "Test message",
			},
			wantErr: true,
			errMsg:  "to number is required",
		},
		{
			name: "missing message",
			request: &SMSRequest{
				From: "+1234567890",
				To:   "+0987654321",
			},
			wantErr: true,
			errMsg:  "message is required",
		},
		{
			name: "message too long",
			request: &SMSRequest{
				From:    "+1234567890",
				To:      "+0987654321",
				Message: string(make([]byte, 1601)),
			},
			wantErr: true,
			errMsg:  "message exceeds maximum length",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.request.Validate()
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestAttachment_Validate(t *testing.T) {
	tests := []struct {
		name       string
		attachment *Attachment
		wantErr    bool
		errMsg     string
	}{
		{
			name: "valid attachment",
			attachment: &Attachment{
				Filename:    "test.pdf",
				Content:     []byte("test content"),
				ContentType: "application/pdf",
			},
			wantErr: false,
		},
		{
			name: "missing filename",
			attachment: &Attachment{
				Content:     []byte("test content"),
				ContentType: "application/pdf",
			},
			wantErr: true,
			errMsg:  "attachment filename is required",
		},
		{
			name: "missing content",
			attachment: &Attachment{
				Filename:    "test.pdf",
				ContentType: "application/pdf",
			},
			wantErr: true,
			errMsg:  "attachment content is required",
		},
		{
			name: "missing content type",
			attachment: &Attachment{
				Filename: "test.pdf",
				Content:  []byte("test content"),
			},
			wantErr: true,
			errMsg:  "attachment content type is required",
		},
		{
			name: "attachment too large",
			attachment: &Attachment{
				Filename:    "test.pdf",
				Content:     make([]byte, 11*1024*1024),
				ContentType: "application/pdf",
			},
			wantErr: true,
			errMsg:  "attachment size exceeds maximum",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.attachment.Validate()
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
