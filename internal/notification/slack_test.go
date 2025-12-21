package notification

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewSlackNotifier(t *testing.T) {
	config := SlackConfig{
		WebhookURL: "https://hooks.slack.com/services/T00000000/B00000000/XXXXXXXXXXXX",
	}

	notifier, err := NewSlackNotifier(config)
	require.NoError(t, err)
	require.NotNil(t, notifier)
}

func TestNewSlackNotifier_InvalidConfig(t *testing.T) {
	tests := []struct {
		name   string
		config SlackConfig
	}{
		{
			name:   "missing webhook URL",
			config: SlackConfig{},
		},
		{
			name: "invalid webhook URL",
			config: SlackConfig{
				WebhookURL: "not-a-url",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			notifier, err := NewSlackNotifier(tt.config)
			assert.Error(t, err)
			assert.Nil(t, notifier)
		})
	}
}

func TestSlackNotifier_SendText(t *testing.T) {
	var receivedPayload slackWebhookPayload
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		err := json.NewDecoder(r.Body).Decode(&receivedPayload)
		require.NoError(t, err)

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	}))
	defer server.Close()

	config := SlackConfig{
		WebhookURL: server.URL,
	}

	notifier, err := NewSlackNotifier(config)
	require.NoError(t, err)

	msg := SlackMessage{
		Text: "Test notification",
	}

	err = notifier.Send(context.Background(), msg)
	assert.NoError(t, err)
	assert.Equal(t, "Test notification", receivedPayload.Text)
}

func TestSlackNotifier_SendWithBlocks(t *testing.T) {
	var receivedPayload slackWebhookPayload
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := json.NewDecoder(r.Body).Decode(&receivedPayload)
		require.NoError(t, err)

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	}))
	defer server.Close()

	config := SlackConfig{
		WebhookURL: server.URL,
	}

	notifier, err := NewSlackNotifier(config)
	require.NoError(t, err)

	msg := SlackMessage{
		Text: "Fallback text",
		Blocks: []SlackBlock{
			{
				Type: "header",
				Text: &SlackText{
					Type: "plain_text",
					Text: "Task Assigned",
				},
			},
			{
				Type: "section",
				Text: &SlackText{
					Type: "mrkdwn",
					Text: "*Priority:* High\n*Due:* Tomorrow",
				},
			},
		},
	}

	err = notifier.Send(context.Background(), msg)
	assert.NoError(t, err)
	assert.Equal(t, "Fallback text", receivedPayload.Text)
	assert.Len(t, receivedPayload.Blocks, 2)
	assert.Equal(t, "header", receivedPayload.Blocks[0].Type)
}

func TestSlackNotifier_SendWithChannel(t *testing.T) {
	var receivedPayload slackWebhookPayload
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := json.NewDecoder(r.Body).Decode(&receivedPayload)
		require.NoError(t, err)

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	}))
	defer server.Close()

	config := SlackConfig{
		WebhookURL: server.URL,
	}

	notifier, err := NewSlackNotifier(config)
	require.NoError(t, err)

	msg := SlackMessage{
		Text:    "Test notification",
		Channel: "#alerts",
	}

	err = notifier.Send(context.Background(), msg)
	assert.NoError(t, err)
	assert.Equal(t, "#alerts", receivedPayload.Channel)
}

func TestSlackNotifier_RateLimit(t *testing.T) {
	requestCount := 0
	var mu sync.Mutex

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		requestCount++
		count := requestCount
		mu.Unlock()

		// Rate limit the first 2 requests
		if count <= 2 {
			w.Header().Set("Retry-After", "1")
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	}))
	defer server.Close()

	config := SlackConfig{
		WebhookURL: server.URL,
		MaxRetries: 3,
		RetryDelay: 100 * time.Millisecond,
	}

	notifier, err := NewSlackNotifier(config)
	require.NoError(t, err)

	msg := SlackMessage{
		Text: "Test notification",
	}

	err = notifier.Send(context.Background(), msg)
	assert.NoError(t, err)
	assert.Equal(t, 3, requestCount)
}

func TestSlackNotifier_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("internal server error"))
	}))
	defer server.Close()

	config := SlackConfig{
		WebhookURL: server.URL,
		MaxRetries: 2,
		RetryDelay: 10 * time.Millisecond,
	}

	notifier, err := NewSlackNotifier(config)
	require.NoError(t, err)

	msg := SlackMessage{
		Text: "Test notification",
	}

	err = notifier.Send(context.Background(), msg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "500")
}

func TestSlackNotifier_ContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	config := SlackConfig{
		WebhookURL: server.URL,
	}

	notifier, err := NewSlackNotifier(config)
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	msg := SlackMessage{
		Text: "Test notification",
	}

	err = notifier.Send(ctx, msg)
	assert.Error(t, err)
	assert.Equal(t, context.Canceled, err)
}

func TestSlackNotifier_ValidationError(t *testing.T) {
	config := SlackConfig{
		WebhookURL: "https://hooks.slack.com/services/test",
	}

	notifier, err := NewSlackNotifier(config)
	require.NoError(t, err)

	msg := SlackMessage{
		// No text or blocks
	}

	err = notifier.Send(context.Background(), msg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "text or blocks")
}

func TestBuildTaskAssignedMessage(t *testing.T) {
	msg := BuildTaskAssignedMessage("John Doe", "Approve deployment", "https://app.example.com/tasks/123")

	assert.NotEmpty(t, msg.Text)
	assert.NotEmpty(t, msg.Blocks)
	assert.Contains(t, msg.Text, "John Doe")
	assert.Contains(t, msg.Text, "Approve deployment")
}

func TestBuildTaskCompletedMessage(t *testing.T) {
	msg := BuildTaskCompletedMessage("Review PR", "approved", "Jane Smith")

	assert.NotEmpty(t, msg.Text)
	assert.NotEmpty(t, msg.Blocks)
	assert.Contains(t, msg.Text, "Review PR")
	assert.Contains(t, msg.Text, "Jane Smith")
	// Status is shown in the blocks section, not in the summary text
	assert.Len(t, msg.Blocks, 2)
}

func TestBuildTaskOverdueMessage(t *testing.T) {
	dueDate := time.Now().Add(-24 * time.Hour)
	msg := BuildTaskOverdueMessage("Sign contract", dueDate, "https://app.example.com/tasks/456")

	assert.NotEmpty(t, msg.Text)
	assert.NotEmpty(t, msg.Blocks)
	assert.Contains(t, msg.Text, "Sign contract")
	assert.Contains(t, msg.Text, "overdue")
}

func TestBuildWorkflowExecutionMessage(t *testing.T) {
	msg := BuildWorkflowExecutionMessage("Data Pipeline", "failed", "Connection timeout", "https://app.example.com/executions/789")

	assert.NotEmpty(t, msg.Text)
	assert.NotEmpty(t, msg.Blocks)
	assert.Contains(t, msg.Text, "Data Pipeline")
	assert.Contains(t, msg.Text, "failed")
}

func TestSlackNotifier_ConcurrentRequests(t *testing.T) {
	requestCount := 0
	var mu sync.Mutex

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		requestCount++
		mu.Unlock()

		time.Sleep(10 * time.Millisecond) // Simulate processing
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	}))
	defer server.Close()

	config := SlackConfig{
		WebhookURL: server.URL,
	}

	notifier, err := NewSlackNotifier(config)
	require.NoError(t, err)

	// Send 10 concurrent requests
	var wg sync.WaitGroup
	errors := make([]error, 10)

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()

			msg := SlackMessage{
				Text: "Test notification",
			}

			errors[index] = notifier.Send(context.Background(), msg)
		}(i)
	}

	wg.Wait()

	// All requests should succeed
	for i, err := range errors {
		assert.NoError(t, err, "Request %d failed", i)
	}

	assert.Equal(t, 10, requestCount)
}

func TestSlackNotifier_Timeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(5 * time.Second) // Longer than timeout
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	config := SlackConfig{
		WebhookURL: server.URL,
		Timeout:    100 * time.Millisecond,
	}

	notifier, err := NewSlackNotifier(config)
	require.NoError(t, err)

	msg := SlackMessage{
		Text: "Test notification",
	}

	err = notifier.Send(context.Background(), msg)
	assert.Error(t, err)
	// Error can contain "timeout" or "deadline exceeded" depending on the error source
	errStr := err.Error()
	assert.True(t, strings.Contains(errStr, "timeout") || strings.Contains(errStr, "deadline exceeded"),
		"expected timeout or deadline exceeded error, got: %s", errStr)
}
