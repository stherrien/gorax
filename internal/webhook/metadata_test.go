package webhook

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestWebhookEvent_WithMetadata tests that WebhookEvent can store and retrieve metadata
func TestWebhookEvent_WithMetadata(t *testing.T) {
	tests := []struct {
		name     string
		metadata *EventMetadata
	}{
		{
			name: "complete metadata",
			metadata: &EventMetadata{
				SourceIP:      "192.168.1.1",
				UserAgent:     "Mozilla/5.0 (compatible; Test/1.0)",
				ReceivedAt:    time.Now(),
				ContentType:   "application/json",
				ContentLength: 256,
			},
		},
		{
			name: "minimal metadata",
			metadata: &EventMetadata{
				SourceIP:   "10.0.0.1",
				ReceivedAt: time.Now(),
			},
		},
		{
			name: "metadata with empty user agent",
			metadata: &EventMetadata{
				SourceIP:      "172.16.0.1",
				UserAgent:     "",
				ReceivedAt:    time.Now(),
				ContentType:   "text/plain",
				ContentLength: 0,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			event := &WebhookEvent{
				ID:             "test-id",
				TenantID:       "tenant-1",
				WebhookID:      "webhook-1",
				RequestMethod:  "POST",
				RequestHeaders: map[string]string{},
				RequestBody:    json.RawMessage(`{}`),
				Status:         EventStatusReceived,
				Metadata:       tt.metadata,
			}

			// Verify metadata is set correctly
			assert.NotNil(t, event.Metadata)
			assert.Equal(t, tt.metadata.SourceIP, event.Metadata.SourceIP)
			assert.Equal(t, tt.metadata.UserAgent, event.Metadata.UserAgent)
			assert.Equal(t, tt.metadata.ContentType, event.Metadata.ContentType)
			assert.Equal(t, tt.metadata.ContentLength, event.Metadata.ContentLength)
			assert.WithinDuration(t, tt.metadata.ReceivedAt, event.Metadata.ReceivedAt, time.Second)
		})
	}
}

// TestEventMetadata_JSONSerialization tests metadata can be marshaled/unmarshaled
func TestEventMetadata_JSONSerialization(t *testing.T) {
	receivedAt := time.Now().UTC()
	original := &EventMetadata{
		SourceIP:      "203.0.113.1",
		UserAgent:     "curl/7.68.0",
		ReceivedAt:    receivedAt,
		ContentType:   "application/json",
		ContentLength: 1024,
	}

	// Marshal to JSON
	data, err := json.Marshal(original)
	assert.NoError(t, err)
	assert.NotEmpty(t, data)

	// Unmarshal from JSON
	var restored EventMetadata
	err = json.Unmarshal(data, &restored)
	assert.NoError(t, err)

	// Verify all fields
	assert.Equal(t, original.SourceIP, restored.SourceIP)
	assert.Equal(t, original.UserAgent, restored.UserAgent)
	assert.Equal(t, original.ContentType, restored.ContentType)
	assert.Equal(t, original.ContentLength, restored.ContentLength)
	assert.WithinDuration(t, original.ReceivedAt, restored.ReceivedAt, time.Second)
}

// TestEventMetadata_ExtractFromRequest tests extracting metadata from HTTP request components
func TestEventMetadata_ExtractFromRequest(t *testing.T) {
	tests := []struct {
		name             string
		remoteAddr       string
		userAgent        string
		contentType      string
		contentLength    int64
		expectedSourceIP string
	}{
		{
			name:             "standard request with port",
			remoteAddr:       "192.168.1.100:54321",
			userAgent:        "PostmanRuntime/7.26.8",
			contentType:      "application/json",
			contentLength:    512,
			expectedSourceIP: "192.168.1.100",
		},
		{
			name:             "IPv6 address",
			remoteAddr:       "[2001:db8::1]:8080",
			userAgent:        "Go-http-client/1.1",
			contentType:      "text/plain",
			contentLength:    128,
			expectedSourceIP: "2001:db8::1",
		},
		{
			name:             "request without user agent",
			remoteAddr:       "10.0.0.5:12345",
			userAgent:        "",
			contentType:      "application/x-www-form-urlencoded",
			contentLength:    0,
			expectedSourceIP: "10.0.0.5",
		},
		{
			name:             "IP without port",
			remoteAddr:       "172.16.0.1",
			userAgent:        "Webhook-Tester",
			contentType:      "application/json",
			contentLength:    256,
			expectedSourceIP: "172.16.0.1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metadata := ExtractMetadataFromRequest(
				tt.remoteAddr,
				tt.userAgent,
				tt.contentType,
				tt.contentLength,
			)

			assert.Equal(t, tt.expectedSourceIP, metadata.SourceIP)
			assert.Equal(t, tt.userAgent, metadata.UserAgent)
			assert.Equal(t, tt.contentType, metadata.ContentType)
			assert.Equal(t, int(tt.contentLength), metadata.ContentLength)
			assert.WithinDuration(t, time.Now(), metadata.ReceivedAt, 2*time.Second)
		})
	}
}

// TestEventMetadata_MakeAvailableInWorkflowContext tests metadata in workflow context
func TestEventMetadata_MakeAvailableInWorkflowContext(t *testing.T) {
	receivedAt := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	event := &WebhookEvent{
		ID:             "event-123",
		TenantID:       "tenant-1",
		WebhookID:      "webhook-1",
		RequestMethod:  "POST",
		RequestHeaders: map[string]string{"Authorization": "Bearer token"},
		RequestBody:    json.RawMessage(`{"data": "test"}`),
		Status:         EventStatusReceived,
		Metadata: &EventMetadata{
			SourceIP:      "198.51.100.42",
			UserAgent:     "GitHub-Hookshot/abc123",
			ReceivedAt:    receivedAt,
			ContentType:   "application/json",
			ContentLength: 100,
		},
	}

	// Convert to workflow context format
	context := event.ToWorkflowContext()

	// Verify metadata is available in context
	assert.NotNil(t, context)
	assert.Contains(t, context, "metadata")

	metadata, ok := context["metadata"].(map[string]interface{})
	assert.True(t, ok, "metadata should be a map")

	assert.Equal(t, "198.51.100.42", metadata["sourceIp"])
	assert.Equal(t, "GitHub-Hookshot/abc123", metadata["userAgent"])
	assert.Equal(t, "application/json", metadata["contentType"])
	assert.Equal(t, 100, metadata["contentLength"])
	assert.Equal(t, receivedAt.Format(time.RFC3339), metadata["receivedAt"])
}
