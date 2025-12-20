package pagerduty

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/gorax/gorax/internal/integrations"
)

const (
	// DefaultBaseURL is the base URL for PagerDuty API
	DefaultBaseURL = "https://api.pagerduty.com"

	// DefaultTimeout is the default HTTP client timeout
	DefaultTimeout = 30 * time.Second
)

// Client is a PagerDuty API client
type Client struct {
	baseURL    string
	apiKey     string
	email      string // Required for some operations
	httpClient *http.Client
}

// NewClient creates a new PagerDuty API client
func NewClient(apiKey, email string) (*Client, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("API key is required")
	}
	if email == "" {
		return nil, fmt.Errorf("email is required")
	}

	return &Client{
		baseURL: DefaultBaseURL,
		apiKey:  apiKey,
		email:   email,
		httpClient: &http.Client{
			Timeout: DefaultTimeout,
		},
	}, nil
}

// Authenticate verifies the API key
func (c *Client) Authenticate(ctx context.Context) error {
	var result map[string]interface{}
	if err := c.doRequest(ctx, "GET", "/users", nil, &result); err != nil {
		return fmt.Errorf("authentication failed: %w", err)
	}
	return nil
}

// HealthCheck verifies the connection is still valid
func (c *Client) HealthCheck(ctx context.Context) error {
	return c.Authenticate(ctx)
}

// CreateIncident creates a new incident
func (c *Client) CreateIncident(ctx context.Context, title, serviceID, urgency, body, incidentKey string) (*Incident, error) {
	if title == "" || serviceID == "" {
		return nil, fmt.Errorf("title and service are required")
	}

	if urgency == "" {
		urgency = "high"
	}

	payload := map[string]interface{}{
		"incident": map[string]interface{}{
			"type":  "incident",
			"title": title,
			"service": map[string]interface{}{
				"id":   serviceID,
				"type": "service_reference",
			},
			"urgency": urgency,
		},
	}

	if body != "" {
		payload["incident"].(map[string]interface{})["body"] = map[string]interface{}{
			"type":    "incident_body",
			"details": body,
		}
	}

	if incidentKey != "" {
		payload["incident"].(map[string]interface{})["incident_key"] = incidentKey
	}

	var resp struct {
		Incident Incident `json:"incident"`
	}

	if err := c.doRequest(ctx, "POST", "/incidents", payload, &resp); err != nil {
		return nil, fmt.Errorf("failed to create incident: %w", err)
	}

	return &resp.Incident, nil
}

// AcknowledgeIncident acknowledges an incident
func (c *Client) AcknowledgeIncident(ctx context.Context, incidentID string) error {
	return c.updateIncidentStatus(ctx, incidentID, "acknowledged")
}

// ResolveIncident resolves an incident
func (c *Client) ResolveIncident(ctx context.Context, incidentID string) error {
	return c.updateIncidentStatus(ctx, incidentID, "resolved")
}

// updateIncidentStatus updates the status of an incident
func (c *Client) updateIncidentStatus(ctx context.Context, incidentID, status string) error {
	if incidentID == "" {
		return fmt.Errorf("incident ID is required")
	}

	payload := map[string]interface{}{
		"incident": map[string]interface{}{
			"type":   "incident_reference",
			"status": status,
		},
	}

	endpoint := fmt.Sprintf("/incidents/%s", incidentID)
	if err := c.doRequest(ctx, "PUT", endpoint, payload, nil); err != nil {
		return fmt.Errorf("failed to update incident status: %w", err)
	}

	return nil
}

// AddNote adds a note to an incident
func (c *Client) AddNote(ctx context.Context, incidentID, content string) (*Note, error) {
	if incidentID == "" || content == "" {
		return nil, fmt.Errorf("incident ID and content are required")
	}

	payload := map[string]interface{}{
		"note": map[string]interface{}{
			"content": content,
		},
	}

	var resp struct {
		Note Note `json:"note"`
	}

	endpoint := fmt.Sprintf("/incidents/%s/notes", incidentID)
	if err := c.doRequest(ctx, "POST", endpoint, payload, &resp); err != nil {
		return nil, fmt.Errorf("failed to add note: %w", err)
	}

	return &resp.Note, nil
}

// doRequest performs an HTTP request with error handling
func (c *Client) doRequest(ctx context.Context, method, endpoint string, body interface{}, result interface{}) error {
	return integrations.WithRetry(ctx, integrations.DefaultRetryConfig, func() error {
		return c.doRequestOnce(ctx, method, endpoint, body, result)
	})
}

// doRequestOnce performs a single HTTP request
func (c *Client) doRequestOnce(ctx context.Context, method, endpoint string, body interface{}, result interface{}) error {
	url := c.baseURL + endpoint

	var bodyReader io.Reader
	if body != nil {
		bodyJSON, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("failed to marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(bodyJSON)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Authorization", "Token token="+c.apiKey)
	req.Header.Set("Accept", "application/vnd.pagerduty+json;version=2")
	req.Header.Set("From", c.email)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	// Handle rate limiting
	if resp.StatusCode == http.StatusTooManyRequests {
		return integrations.ErrRateLimitExceeded
	}

	// Handle authentication errors
	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		return integrations.ErrAuthFailed
	}

	// Handle 404
	if resp.StatusCode == http.StatusNotFound {
		return integrations.ErrNotFound
	}

	// Handle other errors
	if resp.StatusCode >= 400 {
		var errResp ErrorResponse
		if err := json.Unmarshal(respBody, &errResp); err == nil {
			return fmt.Errorf("PagerDuty API error: %s", errResp.Error.Message)
		}
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(respBody))
	}

	// Parse response if result is provided
	if result != nil && resp.StatusCode != http.StatusNoContent {
		if err := json.Unmarshal(respBody, result); err != nil {
			return fmt.Errorf("failed to parse response: %w", err)
		}
	}

	return nil
}
