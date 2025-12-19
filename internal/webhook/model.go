package webhook

import (
	"encoding/json"
	"time"
)

// Webhook represents a webhook configuration
type Webhook struct {
	ID              string     `db:"id" json:"id"`
	TenantID        string     `db:"tenant_id" json:"tenant_id"`
	WorkflowID      string     `db:"workflow_id" json:"workflow_id"`
	NodeID          string     `db:"node_id" json:"node_id"`
	Name            string     `db:"name" json:"name"`
	Path            string     `db:"path" json:"path"`
	Secret          string     `db:"secret" json:"secret"`
	AuthType        string     `db:"auth_type" json:"auth_type"`
	Description     string     `db:"description" json:"description"`
	Priority        int        `db:"priority" json:"priority"`
	Enabled         bool       `db:"enabled" json:"enabled"`
	TriggerCount    int        `db:"trigger_count" json:"trigger_count"`
	LastTriggeredAt *time.Time `db:"last_triggered_at" json:"last_triggered_at,omitempty"`
	CreatedAt       time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt       time.Time  `db:"updated_at" json:"updated_at"`
}

// WebhookURL returns the full webhook URL path
func (w *Webhook) WebhookURL() string {
	return "/webhooks/" + w.WorkflowID + "/" + w.ID
}

// AuthType constants
const (
	AuthTypeNone      = "none"
	AuthTypeSignature = "signature"
	AuthTypeBasic     = "basic"
	AuthTypeAPIKey    = "api_key"
)

// WebhookEventStatus represents the status of a webhook event
type WebhookEventStatus string

const (
	EventStatusReceived  WebhookEventStatus = "received"
	EventStatusProcessed WebhookEventStatus = "processed"
	EventStatusFiltered  WebhookEventStatus = "filtered"
	EventStatusFailed    WebhookEventStatus = "failed"
)

// WebhookEvent represents a webhook event delivery log
type WebhookEvent struct {
	ID               string             `json:"id" db:"id"`
	TenantID         string             `json:"tenantId" db:"tenant_id"`
	WebhookID        string             `json:"webhookId" db:"webhook_id"`
	ExecutionID      *string            `json:"executionId,omitempty" db:"execution_id"`
	RequestMethod    string             `json:"requestMethod" db:"request_method"`
	RequestHeaders   map[string]string  `json:"requestHeaders" db:"request_headers"`
	RequestBody      json.RawMessage    `json:"requestBody" db:"request_body"`
	ResponseStatus   *int               `json:"responseStatus,omitempty" db:"response_status"`
	ProcessingTimeMs *int               `json:"processingTimeMs,omitempty" db:"processing_time_ms"`
	Status           WebhookEventStatus `json:"status" db:"status"`
	ErrorMessage     *string            `json:"errorMessage,omitempty" db:"error_message"`
	FilteredReason   *string            `json:"filteredReason,omitempty" db:"filtered_reason"`
	ReplayCount      int                `json:"replayCount" db:"replay_count"`
	SourceEventID    *string            `json:"sourceEventId,omitempty" db:"source_event_id"`
	CreatedAt        time.Time          `json:"createdAt" db:"created_at"`
}

// WebhookEventFilter represents filter criteria for querying webhook events
type WebhookEventFilter struct {
	WebhookID string
	Status    *WebhookEventStatus
	StartDate *time.Time
	EndDate   *time.Time
	Limit     int
	Offset    int
}

// FilterOperator represents an operator for filter evaluation
type FilterOperator string

const (
	OpEquals       FilterOperator = "equals"
	OpNotEquals    FilterOperator = "not_equals"
	OpContains     FilterOperator = "contains"
	OpNotContains  FilterOperator = "not_contains"
	OpStartsWith   FilterOperator = "starts_with"
	OpEndsWith     FilterOperator = "ends_with"
	OpRegex        FilterOperator = "regex"
	OpGreaterThan  FilterOperator = "gt"
	OpLessThan     FilterOperator = "lt"
	OpIn           FilterOperator = "in"
	OpNotIn        FilterOperator = "not_in"
	OpExists       FilterOperator = "exists"
	OpNotExists    FilterOperator = "not_exists"
)

// WebhookFilter represents a filter rule for webhook payload evaluation
type WebhookFilter struct {
	ID         string         `json:"id" db:"id"`
	WebhookID  string         `json:"webhookId" db:"webhook_id"`
	FieldPath  string         `json:"fieldPath" db:"field_path"`   // JSON path like "$.data.status"
	Operator   FilterOperator `json:"operator" db:"operator"`
	Value      interface{}    `json:"value" db:"value"`
	LogicGroup int            `json:"logicGroup" db:"logic_group"` // For AND/OR grouping
	Enabled    bool           `json:"enabled" db:"enabled"`
	CreatedAt  time.Time      `json:"createdAt" db:"created_at"`
	UpdatedAt  time.Time      `json:"updatedAt" db:"updated_at"`
}

// FilterResult represents the result of filter evaluation
type FilterResult struct {
	Passed  bool                   `json:"passed"`
	Reason  string                 `json:"reason"`
	Details map[string]interface{} `json:"details"`
}

// TestResult represents the result of testing a webhook
type TestResult struct {
	Success      bool   `json:"success"`
	StatusCode   int    `json:"statusCode"`
	ResponseTime int    `json:"responseTimeMs"`
	ExecutionID  string `json:"executionId,omitempty"`
	Error        string `json:"error,omitempty"`
}

// Event represents a webhook event log entry
type Event struct {
	ID           string     `db:"id" json:"id"`
	WebhookID    string     `db:"webhook_id" json:"webhookId"`
	ExecutionID  string     `db:"execution_id" json:"executionId"`
	Status       string     `db:"status" json:"status"`
	StatusCode   int        `db:"status_code" json:"statusCode"`
	ResponseTime int        `db:"response_time" json:"responseTimeMs"`
	ErrorMessage *string    `db:"error_message" json:"errorMessage,omitempty"`
	CreatedAt    time.Time  `db:"created_at" json:"createdAt"`
}

// ReplayRequest represents a request to replay a webhook event
type ReplayRequest struct {
	EventID         string          `json:"eventId"`
	ModifiedPayload json.RawMessage `json:"modifiedPayload,omitempty"`
}

// BatchReplayRequest represents a request to replay multiple webhook events
type BatchReplayRequest struct {
	EventIDs []string `json:"eventIds"`
}

// ReplayResult represents the result of replaying a webhook event
type ReplayResult struct {
	Success     bool   `json:"success"`
	ExecutionID string `json:"executionId,omitempty"`
	Error       string `json:"error,omitempty"`
}

// BatchReplayResponse represents the response from batch replaying events
type BatchReplayResponse struct {
	Results map[string]*ReplayResult `json:"results"`
}
