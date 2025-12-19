package slack

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"
)

const (
	// DefaultBaseURL is the base URL for Slack API
	DefaultBaseURL = "https://slack.com/api"

	// DefaultTimeout is the default HTTP client timeout
	DefaultTimeout = 30 * time.Second

	// DefaultMaxRetries is the default number of retry attempts
	DefaultMaxRetries = 3

	// DefaultRetryDelay is the default delay between retries
	DefaultRetryDelay = 1 * time.Second
)

// Client is a Slack API client
type Client struct {
	accessToken string
	baseURL     string
	httpClient  *http.Client
	maxRetries  int
	retryDelay  time.Duration
}

// NewClient creates a new Slack API client
func NewClient(accessToken string) (*Client, error) {
	if accessToken == "" {
		return nil, ErrInvalidToken
	}

	return &Client{
		accessToken: accessToken,
		baseURL:     DefaultBaseURL,
		httpClient: &http.Client{
			Timeout: DefaultTimeout,
		},
		maxRetries: DefaultMaxRetries,
		retryDelay: DefaultRetryDelay,
	}, nil
}

// WithBaseURL sets a custom base URL (useful for testing)
func (c *Client) WithBaseURL(baseURL string) *Client {
	c.baseURL = baseURL
	return c
}

// WithTimeout sets a custom timeout
func (c *Client) WithTimeout(timeout time.Duration) *Client {
	c.httpClient.Timeout = timeout
	return c
}

// WithMaxRetries sets the maximum number of retry attempts
func (c *Client) WithMaxRetries(maxRetries int) *Client {
	c.maxRetries = maxRetries
	return c
}

// SendMessage sends a message to a Slack channel
func (c *Client) SendMessage(ctx context.Context, req *SendMessageRequest) (*MessageResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("request cannot be nil")
	}

	var resp MessageResponse
	if err := c.doRequest(ctx, "POST", "/chat.postMessage", req, &resp); err != nil {
		return nil, err
	}

	if !resp.OK {
		return nil, ParseSlackError(resp.Error)
	}

	return &resp, nil
}

// UpdateMessage updates an existing message
func (c *Client) UpdateMessage(ctx context.Context, req *UpdateMessageRequest) (*MessageResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("request cannot be nil")
	}

	var resp MessageResponse
	if err := c.doRequest(ctx, "POST", "/chat.update", req, &resp); err != nil {
		return nil, err
	}

	if !resp.OK {
		return nil, ParseSlackError(resp.Error)
	}

	return &resp, nil
}

// DeleteMessage deletes a message
func (c *Client) DeleteMessage(ctx context.Context, channel, timestamp string) error {
	req := map[string]string{
		"channel": channel,
		"ts":      timestamp,
	}

	var resp APIResponse
	if err := c.doRequest(ctx, "POST", "/chat.delete", req, &resp); err != nil {
		return err
	}

	if !resp.OK {
		return ParseSlackError(resp.Error)
	}

	return nil
}

// AddReaction adds a reaction emoji to a message
func (c *Client) AddReaction(ctx context.Context, channel, timestamp, emoji string) error {
	req := map[string]string{
		"channel":   channel,
		"timestamp": timestamp,
		"name":      emoji,
	}

	var resp APIResponse
	if err := c.doRequest(ctx, "POST", "/reactions.add", req, &resp); err != nil {
		return err
	}

	// Treat "already_reacted" as success
	if !resp.OK && resp.Error != "already_reacted" {
		return ParseSlackError(resp.Error)
	}

	return nil
}

// RemoveReaction removes a reaction emoji from a message
func (c *Client) RemoveReaction(ctx context.Context, channel, timestamp, emoji string) error {
	req := map[string]string{
		"channel":   channel,
		"timestamp": timestamp,
		"name":      emoji,
	}

	var resp APIResponse
	if err := c.doRequest(ctx, "POST", "/reactions.remove", req, &resp); err != nil {
		return err
	}

	if !resp.OK {
		return ParseSlackError(resp.Error)
	}

	return nil
}

// GetUserByEmail looks up a user by their email address
func (c *Client) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	if email == "" {
		return nil, ErrUserRequired
	}

	var resp UserByEmailResponse
	endpoint := fmt.Sprintf("/users.lookupByEmail?email=%s", email)
	if err := c.doRequest(ctx, "GET", endpoint, nil, &resp); err != nil {
		return nil, err
	}

	if !resp.OK {
		return nil, ParseSlackError(resp.Error)
	}

	return &resp.User, nil
}

// GetUserByID looks up a user by their ID
func (c *Client) GetUserByID(ctx context.Context, userID string) (*User, error) {
	if userID == "" {
		return nil, ErrUserRequired
	}

	req := map[string]string{
		"user": userID,
	}

	var resp struct {
		OK    bool   `json:"ok"`
		User  User   `json:"user"`
		Error string `json:"error,omitempty"`
	}

	if err := c.doRequest(ctx, "POST", "/users.info", req, &resp); err != nil {
		return nil, err
	}

	if !resp.OK {
		return nil, ParseSlackError(resp.Error)
	}

	return &resp.User, nil
}

// OpenConversation opens a direct message or multi-person direct message
func (c *Client) OpenConversation(ctx context.Context, users []string) (*Conversation, error) {
	if len(users) == 0 {
		return nil, ErrUserRequired
	}

	req := &OpenConversationRequest{
		Users: users,
	}

	var resp OpenConversationResponse
	if err := c.doRequest(ctx, "POST", "/conversations.open", req, &resp); err != nil {
		return nil, err
	}

	if !resp.OK {
		return nil, ParseSlackError(resp.Error)
	}

	return &resp.Channel, nil
}

// ListChannels lists channels in the workspace
func (c *Client) ListChannels(ctx context.Context, types []string) ([]*Conversation, error) {
	req := map[string]interface{}{
		"types": types,
		"limit": 100,
	}

	var resp ListChannelsResponse
	if err := c.doRequest(ctx, "POST", "/conversations.list", req, &resp); err != nil {
		return nil, err
	}

	if !resp.OK {
		return nil, ParseSlackError(resp.Error)
	}

	return resp.Channels, nil
}

// doRequest performs an HTTP request with retry logic
func (c *Client) doRequest(ctx context.Context, method, endpoint string, body interface{}, result interface{}) error {
	var lastErr error

	for attempt := 0; attempt <= c.maxRetries; attempt++ {
		// Check context before attempting
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// Attempt the request
		err := c.doRequestOnce(ctx, method, endpoint, body, result)
		if err == nil {
			return nil // Success
		}

		lastErr = err

		// Check if error is retryable
		slackErr, ok := err.(*SlackError)
		if !ok {
			return err // Not a Slack error, don't retry
		}

		if !slackErr.IsRetryable() {
			return err // Not retryable
		}

		// Don't retry on last attempt
		if attempt == c.maxRetries {
			break
		}

		// Calculate retry delay
		retryDelay := c.retryDelay
		if slackErr.RetryAfter > 0 {
			retryDelay = time.Duration(slackErr.RetryAfter) * time.Second
		}

		// Wait before retrying
		select {
		case <-time.After(retryDelay):
			// Continue to next attempt
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	return lastErr
}

// doRequestOnce performs a single HTTP request
func (c *Client) doRequestOnce(ctx context.Context, method, endpoint string, body interface{}, result interface{}) error {
	// Build URL
	url := c.baseURL + endpoint

	// Prepare request body
	var bodyReader io.Reader
	if body != nil {
		bodyJSON, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("failed to marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(bodyJSON)
	}

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Authorization", "Bearer "+c.accessToken)
	if body != nil {
		req.Header.Set("Content-Type", "application/json; charset=utf-8")
	}

	// Execute request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	// Handle rate limiting
	if resp.StatusCode == http.StatusTooManyRequests {
		retryAfter := 1 // Default to 1 second
		if retryHeader := resp.Header.Get("Retry-After"); retryHeader != "" {
			if seconds, err := strconv.Atoi(retryHeader); err == nil {
				retryAfter = seconds
			}
		}

		return &SlackError{
			ErrorCode:  "rate_limited",
			Message:    "Slack API rate limit exceeded",
			RetryAfter: retryAfter,
		}
	}

	// Handle other HTTP errors
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(respBody))
	}

	// Parse response
	if err := json.Unmarshal(respBody, result); err != nil {
		return fmt.Errorf("failed to parse response: %w (body: %s)", err, string(respBody))
	}

	return nil
}

// ExchangeCode exchanges an OAuth2 authorization code for an access token
func (c *Client) ExchangeCode(ctx context.Context, clientID, clientSecret, code, redirectURI string) (*OAuthResponse, error) {
	req := map[string]string{
		"client_id":     clientID,
		"client_secret": clientSecret,
		"code":          code,
		"redirect_uri":  redirectURI,
	}

	var resp OAuthResponse
	if err := c.doRequest(ctx, "POST", "/oauth.v2.access", req, &resp); err != nil {
		return nil, err
	}

	if !resp.OK {
		if resp.Error != "" {
			return nil, fmt.Errorf("OAuth error: %s (%s)", resp.Error, resp.ErrorDescription)
		}
		return nil, fmt.Errorf("OAuth exchange failed")
	}

	return &resp, nil
}

// RefreshToken refreshes an expired access token using a refresh token
func (c *Client) RefreshToken(ctx context.Context, clientID, clientSecret, refreshToken string) (*OAuthResponse, error) {
	req := map[string]string{
		"client_id":     clientID,
		"client_secret": clientSecret,
		"refresh_token": refreshToken,
		"grant_type":    "refresh_token",
	}

	var resp OAuthResponse
	if err := c.doRequest(ctx, "POST", "/oauth.v2.access", req, &resp); err != nil {
		return nil, err
	}

	if !resp.OK {
		if resp.Error != "" {
			return nil, fmt.Errorf("OAuth refresh error: %s", resp.Error)
		}
		return nil, fmt.Errorf("OAuth refresh failed")
	}

	return &resp, nil
}
