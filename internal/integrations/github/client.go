package github

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
	// DefaultBaseURL is the base URL for GitHub API
	DefaultBaseURL = "https://api.github.com"

	// DefaultTimeout is the default HTTP client timeout
	DefaultTimeout = 30 * time.Second
)

// Client is a GitHub API client
type Client struct {
	baseURL    string
	token      string
	httpClient *http.Client
}

// NewClient creates a new GitHub API client
func NewClient(token string) (*Client, error) {
	if token == "" {
		return nil, fmt.Errorf("token is required")
	}

	return &Client{
		baseURL: DefaultBaseURL,
		token:   token,
		httpClient: &http.Client{
			Timeout: DefaultTimeout,
		},
	}, nil
}

// Authenticate verifies the token
func (c *Client) Authenticate(ctx context.Context) error {
	var user map[string]interface{}
	if err := c.doRequest(ctx, "GET", "/user", nil, &user); err != nil {
		return fmt.Errorf("authentication failed: %w", err)
	}
	return nil
}

// HealthCheck verifies the connection is still valid
func (c *Client) HealthCheck(ctx context.Context) error {
	return c.Authenticate(ctx)
}

// CreateIssue creates a new GitHub issue
func (c *Client) CreateIssue(ctx context.Context, owner, repo, title, body string, labels []string) (*Issue, error) {
	if owner == "" || repo == "" || title == "" {
		return nil, fmt.Errorf("owner, repo, and title are required")
	}

	payload := map[string]interface{}{
		"title": title,
		"body":  body,
	}
	if len(labels) > 0 {
		payload["labels"] = labels
	}

	var issue Issue
	endpoint := fmt.Sprintf("/repos/%s/%s/issues", owner, repo)
	if err := c.doRequest(ctx, "POST", endpoint, payload, &issue); err != nil {
		return nil, fmt.Errorf("failed to create issue: %w", err)
	}

	return &issue, nil
}

// CreatePRComment creates a comment on a pull request
func (c *Client) CreatePRComment(ctx context.Context, owner, repo string, number int, body string) (*Comment, error) {
	if owner == "" || repo == "" || number == 0 || body == "" {
		return nil, fmt.Errorf("owner, repo, number, and body are required")
	}

	payload := map[string]interface{}{
		"body": body,
	}

	var comment Comment
	endpoint := fmt.Sprintf("/repos/%s/%s/issues/%d/comments", owner, repo, number)
	if err := c.doRequest(ctx, "POST", endpoint, payload, &comment); err != nil {
		return nil, fmt.Errorf("failed to create PR comment: %w", err)
	}

	return &comment, nil
}

// AddLabels adds labels to an issue or PR
func (c *Client) AddLabels(ctx context.Context, owner, repo string, number int, labels []string) error {
	if owner == "" || repo == "" || number == 0 || len(labels) == 0 {
		return fmt.Errorf("owner, repo, number, and labels are required")
	}

	payload := map[string]interface{}{
		"labels": labels,
	}

	endpoint := fmt.Sprintf("/repos/%s/%s/issues/%d/labels", owner, repo, number)
	if err := c.doRequest(ctx, "POST", endpoint, payload, nil); err != nil {
		return fmt.Errorf("failed to add labels: %w", err)
	}

	return nil
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
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Accept", "application/vnd.github+json")
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
	if resp.StatusCode == http.StatusForbidden {
		// Check if it's a rate limit error
		if resp.Header.Get("X-RateLimit-Remaining") == "0" {
			return integrations.ErrRateLimitExceeded
		}
	}

	// Handle authentication errors
	if resp.StatusCode == http.StatusUnauthorized {
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
			return fmt.Errorf("GitHub API error: %s", errResp.Message)
		}
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(respBody))
	}

	// Parse response if result is provided
	if result != nil {
		if err := json.Unmarshal(respBody, result); err != nil {
			return fmt.Errorf("failed to parse response: %w", err)
		}
	}

	return nil
}
