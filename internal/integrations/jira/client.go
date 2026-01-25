package jira

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
	// DefaultTimeout is the default HTTP client timeout
	DefaultTimeout = 30 * time.Second
)

// Client is a Jira API client
type Client struct {
	baseURL    string
	email      string
	apiToken   string
	httpClient *http.Client
}

// NewClient creates a new Jira API client
func NewClient(baseURL, email, apiToken string) (*Client, error) {
	if baseURL == "" {
		return nil, fmt.Errorf("base URL is required")
	}
	if email == "" {
		return nil, fmt.Errorf("email is required")
	}
	if apiToken == "" {
		return nil, fmt.Errorf("API token is required")
	}

	return &Client{
		baseURL:  baseURL,
		email:    email,
		apiToken: apiToken,
		httpClient: &http.Client{
			Timeout: DefaultTimeout,
		},
	}, nil
}

// Authenticate verifies the credentials
func (c *Client) Authenticate(ctx context.Context) error {
	var user User
	if err := c.doRequest(ctx, "GET", "/rest/api/3/myself", nil, &user); err != nil {
		return fmt.Errorf("authentication failed: %w", err)
	}
	return nil
}

// HealthCheck verifies the connection is still valid
func (c *Client) HealthCheck(ctx context.Context) error {
	return c.Authenticate(ctx)
}

// CreateIssue creates a new Jira issue
func (c *Client) CreateIssue(ctx context.Context, req CreateIssueRequest) (*Issue, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	fields := map[string]interface{}{
		"project": map[string]interface{}{
			"key": req.Project,
		},
		"issuetype": map[string]interface{}{
			"name": req.IssueType,
		},
		"summary":     req.Summary,
		"description": req.Description,
	}

	// Add optional fields
	if req.Priority != "" {
		fields["priority"] = map[string]interface{}{"name": req.Priority}
	}
	if req.Assignee != "" {
		fields["assignee"] = map[string]interface{}{"accountId": req.Assignee}
	}
	if len(req.Labels) > 0 {
		fields["labels"] = req.Labels
	}
	if len(req.Components) > 0 {
		components := make([]map[string]interface{}, len(req.Components))
		for i, comp := range req.Components {
			components[i] = map[string]interface{}{"name": comp}
		}
		fields["components"] = components
	}

	payload := map[string]interface{}{
		"fields": fields,
	}

	var resp struct {
		ID   string `json:"id"`
		Key  string `json:"key"`
		Self string `json:"self"`
	}

	if err := c.doRequest(ctx, "POST", "/rest/api/3/issue", payload, &resp); err != nil {
		return nil, fmt.Errorf("failed to create issue: %w", err)
	}

	return &Issue{
		ID:   resp.ID,
		Key:  resp.Key,
		Self: resp.Self,
	}, nil
}

// UpdateIssue updates an existing issue
func (c *Client) UpdateIssue(ctx context.Context, issueKey string, fields map[string]interface{}) error {
	if issueKey == "" {
		return fmt.Errorf("issue key is required")
	}

	payload := map[string]interface{}{
		"fields": fields,
	}

	endpoint := fmt.Sprintf("/rest/api/3/issue/%s", issueKey)
	if err := c.doRequest(ctx, "PUT", endpoint, payload, nil); err != nil {
		return fmt.Errorf("failed to update issue: %w", err)
	}

	return nil
}

// AddComment adds a comment to an issue
func (c *Client) AddComment(ctx context.Context, issueKey, body string) (*Comment, error) {
	if issueKey == "" {
		return nil, fmt.Errorf("issue key is required")
	}
	if body == "" {
		return nil, fmt.Errorf("comment body is required")
	}

	payload := map[string]interface{}{
		"body": body,
	}

	var comment Comment
	endpoint := fmt.Sprintf("/rest/api/3/issue/%s/comment", issueKey)
	if err := c.doRequest(ctx, "POST", endpoint, payload, &comment); err != nil {
		return nil, fmt.Errorf("failed to add comment: %w", err)
	}

	return &comment, nil
}

// TransitionIssue transitions an issue to a new status
func (c *Client) TransitionIssue(ctx context.Context, issueKey, transitionName string) error {
	if issueKey == "" {
		return fmt.Errorf("issue key is required")
	}
	if transitionName == "" {
		return fmt.Errorf("transition name is required")
	}

	// Get available transitions
	var transitionsResp struct {
		Transitions []Transition `json:"transitions"`
	}

	endpoint := fmt.Sprintf("/rest/api/3/issue/%s/transitions", issueKey)
	if err := c.doRequest(ctx, "GET", endpoint, nil, &transitionsResp); err != nil {
		return fmt.Errorf("failed to get transitions: %w", err)
	}

	// Find the transition ID by name
	var transitionID string
	for _, t := range transitionsResp.Transitions {
		if t.Name == transitionName {
			transitionID = t.ID
			break
		}
	}

	if transitionID == "" {
		return fmt.Errorf("transition '%s' not found", transitionName)
	}

	// Perform the transition
	payload := map[string]interface{}{
		"transition": map[string]interface{}{
			"id": transitionID,
		},
	}

	if err := c.doRequest(ctx, "POST", endpoint, payload, nil); err != nil {
		return fmt.Errorf("failed to transition issue: %w", err)
	}

	return nil
}

// SearchIssues searches for issues using JQL
func (c *Client) SearchIssues(ctx context.Context, jql string, maxResults, startAt int) (*SearchResult, error) {
	if jql == "" {
		return nil, fmt.Errorf("JQL query is required")
	}

	payload := map[string]interface{}{
		"jql":        jql,
		"maxResults": maxResults,
		"startAt":    startAt,
	}

	var result SearchResult
	if err := c.doRequest(ctx, "POST", "/rest/api/3/search", payload, &result); err != nil {
		return nil, fmt.Errorf("failed to search issues: %w", err)
	}

	return &result, nil
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

	// Set Basic Auth
	req.SetBasicAuth(c.email, c.apiToken)

	// Set headers
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Accept", "application/json")

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
			return fmt.Errorf("jira API error: %v", errResp.ErrorMessages)
		}
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(respBody))
	}

	// For 204 No Content, don't try to parse response
	if resp.StatusCode == http.StatusNoContent {
		return nil
	}

	// Parse response if result is provided
	if result != nil {
		if err := json.Unmarshal(respBody, result); err != nil {
			return fmt.Errorf("failed to parse response: %w", err)
		}
	}

	return nil
}
