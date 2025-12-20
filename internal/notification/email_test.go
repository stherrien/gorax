package notification

import (
	"context"
	"fmt"
	"io"
	"net"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockSMTPServer simulates an SMTP server for testing
type mockSMTPServer struct {
	addr         string
	listener     net.Listener
	messages     []string
	shouldFail   bool
	failAfter    int
	messageCount int
	t            *testing.T
}

func newMockSMTPServer(t *testing.T) *mockSMTPServer {
	return &mockSMTPServer{
		t:        t,
		messages: make([]string, 0),
	}
}

func (m *mockSMTPServer) Start() error {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return err
	}
	m.listener = listener
	m.addr = listener.Addr().String()

	go m.serve()
	return nil
}

func (m *mockSMTPServer) serve() {
	for {
		conn, err := m.listener.Accept()
		if err != nil {
			return
		}
		go m.handleConnection(conn)
	}
}

func (m *mockSMTPServer) handleConnection(conn net.Conn) {
	defer conn.Close()

	// Simple SMTP conversation
	conn.Write([]byte("220 Mock SMTP Server\r\n"))

	buf := make([]byte, 4096)
	for {
		n, err := conn.Read(buf)
		if err != nil {
			if err != io.EOF {
				m.t.Logf("Read error: %v", err)
			}
			return
		}

		line := string(buf[:n])
		m.t.Logf("Received: %s", strings.TrimSpace(line))

		if strings.HasPrefix(line, "EHLO") || strings.HasPrefix(line, "HELO") {
			conn.Write([]byte("250 OK\r\n"))
		} else if strings.HasPrefix(line, "MAIL FROM") {
			if m.shouldFail && m.messageCount >= m.failAfter {
				conn.Write([]byte("451 Temporary failure\r\n"))
			} else {
				conn.Write([]byte("250 OK\r\n"))
			}
		} else if strings.HasPrefix(line, "RCPT TO") {
			conn.Write([]byte("250 OK\r\n"))
		} else if strings.HasPrefix(line, "DATA") {
			conn.Write([]byte("354 Start mail input\r\n"))
		} else if strings.HasPrefix(line, ".") {
			m.messageCount++
			m.messages = append(m.messages, line)
			conn.Write([]byte("250 OK\r\n"))
		} else if strings.HasPrefix(line, "QUIT") {
			conn.Write([]byte("221 Bye\r\n"))
			return
		}
	}
}

func (m *mockSMTPServer) Stop() error {
	if m.listener != nil {
		return m.listener.Close()
	}
	return nil
}

func TestNewEmailSender_SMTP(t *testing.T) {
	config := EmailConfig{
		Provider: EmailProviderSMTP,
		SMTPHost: "smtp.example.com",
		SMTPPort: 587,
		SMTPUser: "user@example.com",
		SMTPPass: "password",
		From:     "noreply@example.com",
	}

	sender, err := NewEmailSender(config)
	require.NoError(t, err)
	require.NotNil(t, sender)
}

func TestNewEmailSender_InvalidProvider(t *testing.T) {
	config := EmailConfig{
		Provider: "invalid",
	}

	sender, err := NewEmailSender(config)
	assert.Error(t, err)
	assert.Nil(t, sender)
}

func TestNewEmailSender_MissingConfig(t *testing.T) {
	tests := []struct {
		name   string
		config EmailConfig
	}{
		{
			name: "missing SMTP host",
			config: EmailConfig{
				Provider: EmailProviderSMTP,
				SMTPPort: 587,
				SMTPUser: "user",
				SMTPPass: "pass",
				From:     "from@example.com",
			},
		},
		{
			name: "missing SMTP port",
			config: EmailConfig{
				Provider: EmailProviderSMTP,
				SMTPHost: "smtp.example.com",
				SMTPUser: "user",
				SMTPPass: "pass",
				From:     "from@example.com",
			},
		},
		{
			name: "missing from address",
			config: EmailConfig{
				Provider: EmailProviderSMTP,
				SMTPHost: "smtp.example.com",
				SMTPPort: 587,
				SMTPUser: "user",
				SMTPPass: "pass",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sender, err := NewEmailSender(tt.config)
			assert.Error(t, err)
			assert.Nil(t, sender)
		})
	}
}

func TestEmailSender_SendHTML(t *testing.T) {
	server := newMockSMTPServer(t)
	require.NoError(t, server.Start())
	defer server.Stop()

	// Wait for server to start
	time.Sleep(100 * time.Millisecond)

	host, port, _ := net.SplitHostPort(server.addr)
	portNum := 0
	_, err := fmt.Sscanf(port, "%d", &portNum)
	require.NoError(t, err)

	config := EmailConfig{
		Provider: EmailProviderSMTP,
		SMTPHost: host,
		SMTPPort: portNum,
		SMTPUser: "user@example.com",
		SMTPPass: "password",
		From:     "noreply@example.com",
		TLS:      false, // Disable TLS for testing
	}

	sender, err := NewEmailSender(config)
	require.NoError(t, err)

	ctx := context.Background()
	email := Email{
		To:       []string{"recipient@example.com"},
		Subject:  "Test Email",
		HTMLBody: "<h1>Hello</h1><p>This is a test email</p>",
		TextBody: "Hello\n\nThis is a test email",
	}

	err = sender.Send(ctx, email)
	assert.NoError(t, err)
	assert.Len(t, server.messages, 1)
}

func TestEmailSender_SendMultipleRecipients(t *testing.T) {
	server := newMockSMTPServer(t)
	require.NoError(t, server.Start())
	defer server.Stop()

	time.Sleep(100 * time.Millisecond)

	host, port, _ := net.SplitHostPort(server.addr)
	portNum := 0
	_, err := fmt.Sscanf(port, "%d", &portNum)
	require.NoError(t, err)

	config := EmailConfig{
		Provider: EmailProviderSMTP,
		SMTPHost: host,
		SMTPPort: portNum,
		From:     "noreply@example.com",
		TLS:      false,
	}

	sender, err := NewEmailSender(config)
	require.NoError(t, err)

	ctx := context.Background()
	email := Email{
		To:       []string{"recipient1@example.com", "recipient2@example.com"},
		Subject:  "Test Email",
		HTMLBody: "<h1>Hello</h1>",
		TextBody: "Hello",
	}

	err = sender.Send(ctx, email)
	assert.NoError(t, err)
}

func TestEmailSender_SendWithRetry(t *testing.T) {
	server := newMockSMTPServer(t)
	server.shouldFail = true
	server.failAfter = 2 // Fail first 2 attempts, succeed on 3rd

	require.NoError(t, server.Start())
	defer server.Stop()

	time.Sleep(100 * time.Millisecond)

	host, port, _ := net.SplitHostPort(server.addr)
	portNum := 0
	_, err := fmt.Sscanf(port, "%d", &portNum)
	require.NoError(t, err)

	config := EmailConfig{
		Provider:   EmailProviderSMTP,
		SMTPHost:   host,
		SMTPPort:   portNum,
		From:       "noreply@example.com",
		TLS:        false,
		MaxRetries: 3,
		RetryDelay: 100 * time.Millisecond,
	}

	sender, err := NewEmailSender(config)
	require.NoError(t, err)

	ctx := context.Background()
	email := Email{
		To:       []string{"recipient@example.com"},
		Subject:  "Test Email",
		HTMLBody: "<h1>Hello</h1>",
		TextBody: "Hello",
	}

	err = sender.Send(ctx, email)
	assert.NoError(t, err)
	assert.Equal(t, 3, server.messageCount)
}

func TestEmailSender_ValidationErrors(t *testing.T) {
	config := EmailConfig{
		Provider: EmailProviderSMTP,
		SMTPHost: "smtp.example.com",
		SMTPPort: 587,
		From:     "noreply@example.com",
	}

	sender, err := NewEmailSender(config)
	require.NoError(t, err)

	tests := []struct {
		name  string
		email Email
	}{
		{
			name: "no recipients",
			email: Email{
				Subject:  "Test",
				HTMLBody: "Test",
			},
		},
		{
			name: "no subject",
			email: Email{
				To:       []string{"test@example.com"},
				HTMLBody: "Test",
			},
		},
		{
			name: "no body",
			email: Email{
				To:      []string{"test@example.com"},
				Subject: "Test",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := sender.Send(context.Background(), tt.email)
			assert.Error(t, err)
		})
	}
}

func TestEmailSender_ContextCancellation(t *testing.T) {
	config := EmailConfig{
		Provider: EmailProviderSMTP,
		SMTPHost: "smtp.example.com",
		SMTPPort: 587,
		From:     "noreply@example.com",
		TLS:      true,
	}

	sender, err := NewEmailSender(config)
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	email := Email{
		To:       []string{"test@example.com"},
		Subject:  "Test",
		HTMLBody: "Test",
	}

	err = sender.Send(ctx, email)
	assert.Error(t, err)
	assert.Equal(t, context.Canceled, err)
}

func TestTemplateRendering(t *testing.T) {
	template := EmailTemplate{
		Subject:  "Welcome {{.Name}}",
		HTMLBody: "<h1>Hello {{.Name}}</h1><p>Your email is {{.Email}}</p>",
		TextBody: "Hello {{.Name}}\n\nYour email is {{.Email}}",
	}

	data := map[string]interface{}{
		"Name":  "John Doe",
		"Email": "john@example.com",
	}

	email, err := RenderTemplate(template, data)
	require.NoError(t, err)

	assert.Equal(t, "Welcome John Doe", email.Subject)
	assert.Contains(t, email.HTMLBody, "Hello John Doe")
	assert.Contains(t, email.HTMLBody, "john@example.com")
	assert.Contains(t, email.TextBody, "Hello John Doe")
}

func TestTemplateRenderingError(t *testing.T) {
	// Use invalid template syntax to trigger a parse error
	template := EmailTemplate{
		Subject:  "Welcome {{.Name}",  // Missing closing braces
		HTMLBody: "<h1>Hello</h1>",
		TextBody: "Hello",
	}

	data := map[string]interface{}{
		"Name": "John Doe",
	}

	_, err := RenderTemplate(template, data)
	assert.Error(t, err)
}
