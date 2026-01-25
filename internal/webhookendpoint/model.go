package webhookendpoint

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// Status constants for webhook endpoints
const (
	StatusActive    = "active"
	StatusTriggered = "triggered"
	StatusExpired   = "expired"
	StatusCancelled = "cancelled"
)

// WebhookEndpoint represents a dynamic webhook endpoint created by a workflow action
type WebhookEndpoint struct {
	ID            uuid.UUID       `json:"id" db:"id"`
	TenantID      uuid.UUID       `json:"tenant_id" db:"tenant_id"`
	ExecutionID   uuid.UUID       `json:"execution_id" db:"execution_id"`
	StepID        string          `json:"step_id" db:"step_id"`
	EndpointToken string          `json:"endpoint_token" db:"endpoint_token"`
	Config        json.RawMessage `json:"config" db:"config"`
	IsActive      bool            `json:"is_active" db:"is_active"`
	TriggeredAt   *time.Time      `json:"triggered_at" db:"triggered_at"`
	Payload       json.RawMessage `json:"payload" db:"payload"`
	SourceIP      *string         `json:"source_ip" db:"source_ip"`
	UserAgent     *string         `json:"user_agent" db:"user_agent"`
	ContentType   *string         `json:"content_type" db:"content_type"`
	ExpiresAt     time.Time       `json:"expires_at" db:"expires_at"`
	CreatedAt     time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time       `json:"updated_at" db:"updated_at"`
}

// EndpointConfig holds configuration for a webhook endpoint
type EndpointConfig struct {
	// Timeout in seconds for how long to wait for the webhook
	TimeoutSeconds int `json:"timeout_seconds,omitempty"`
	// ExpectedSchema is a JSON schema to validate incoming payload
	ExpectedSchema json.RawMessage `json:"expected_schema,omitempty"`
	// RequiredFields are fields that must be present in the payload
	RequiredFields []string `json:"required_fields,omitempty"`
	// AuthType is the expected authentication type (none, signature, bearer, api_key)
	AuthType string `json:"auth_type,omitempty"`
	// AuthSecret is the secret used for signature verification (HMAC-SHA256)
	AuthSecret string `json:"auth_secret,omitempty"`
	// Description is a human-readable description of this endpoint
	Description string `json:"description,omitempty"`
}

// CreateEndpointRequest represents a request to create a webhook endpoint
type CreateEndpointRequest struct {
	ExecutionID    uuid.UUID       `json:"execution_id"`
	StepID         string          `json:"step_id"`
	TimeoutSeconds int             `json:"timeout_seconds"`
	Config         *EndpointConfig `json:"config,omitempty"`
}

// TriggerResult represents the result of triggering a webhook endpoint
type TriggerResult struct {
	EndpointID  uuid.UUID       `json:"endpoint_id"`
	ExecutionID uuid.UUID       `json:"execution_id"`
	StepID      string          `json:"step_id"`
	Payload     json.RawMessage `json:"payload"`
	TriggeredAt time.Time       `json:"triggered_at"`
}

// EndpointFilter represents filters for querying webhook endpoints
type EndpointFilter struct {
	TenantID    uuid.UUID
	ExecutionID *uuid.UUID
	IsActive    *bool
	ExpiredOnly bool
	Limit       int
	Offset      int
}

// IsExpired returns true if the endpoint has expired
func (e *WebhookEndpoint) IsExpired() bool {
	return time.Now().After(e.ExpiresAt)
}

// CanTrigger returns true if the endpoint can be triggered
func (e *WebhookEndpoint) CanTrigger() bool {
	return e.IsActive && !e.IsExpired()
}

// Trigger marks the endpoint as triggered with the provided payload
func (e *WebhookEndpoint) Trigger(payload json.RawMessage, sourceIP, userAgent, contentType string) error {
	if !e.CanTrigger() {
		if !e.IsActive {
			return ErrEndpointInactive
		}
		return ErrEndpointExpired
	}

	now := time.Now()
	e.TriggeredAt = &now
	e.Payload = payload
	e.IsActive = false

	if sourceIP != "" {
		e.SourceIP = &sourceIP
	}
	if userAgent != "" {
		e.UserAgent = &userAgent
	}
	if contentType != "" {
		e.ContentType = &contentType
	}

	return nil
}

// Cancel marks the endpoint as cancelled
func (e *WebhookEndpoint) Cancel() error {
	if !e.IsActive {
		return ErrEndpointInactive
	}
	e.IsActive = false
	return nil
}

// Expire marks the endpoint as expired
func (e *WebhookEndpoint) Expire() error {
	if !e.IsActive {
		return ErrEndpointInactive
	}
	e.IsActive = false
	return nil
}

// GetConfig parses and returns the endpoint configuration
func (e *WebhookEndpoint) GetConfig() (*EndpointConfig, error) {
	if len(e.Config) == 0 {
		return &EndpointConfig{}, nil
	}

	var config EndpointConfig
	if err := json.Unmarshal(e.Config, &config); err != nil {
		return nil, err
	}
	return &config, nil
}

// URL returns the webhook URL path for this endpoint
func (e *WebhookEndpoint) URL() string {
	return "/api/v1/webhook-actions/" + e.EndpointToken
}

// ToWorkflowContext converts the endpoint data to a map suitable for workflow context
func (e *WebhookEndpoint) ToWorkflowContext() map[string]interface{} {
	ctx := map[string]interface{}{
		"endpointId":   e.ID.String(),
		"executionId":  e.ExecutionID.String(),
		"stepId":       e.StepID,
		"url":          e.URL(),
		"isActive":     e.IsActive,
		"expiresAt":    e.ExpiresAt.Format(time.RFC3339),
		"createdAt":    e.CreatedAt.Format(time.RFC3339),
		"wasTriggered": e.TriggeredAt != nil,
	}

	if e.TriggeredAt != nil {
		ctx["triggeredAt"] = e.TriggeredAt.Format(time.RFC3339)

		// Parse payload into context
		if len(e.Payload) > 0 {
			var payload interface{}
			if err := json.Unmarshal(e.Payload, &payload); err == nil {
				ctx["payload"] = payload
			} else {
				ctx["payload"] = string(e.Payload)
			}
		}

		// Add metadata
		metadata := make(map[string]interface{})
		if e.SourceIP != nil {
			metadata["sourceIp"] = *e.SourceIP
		}
		if e.UserAgent != nil {
			metadata["userAgent"] = *e.UserAgent
		}
		if e.ContentType != nil {
			metadata["contentType"] = *e.ContentType
		}
		if len(metadata) > 0 {
			ctx["metadata"] = metadata
		}
	}

	return ctx
}
