// Package http provides a resilient HTTP client with retry logic,
// circuit breaker pattern, and middleware support.
package http

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"maps"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

// Default configuration values.
const (
	DefaultTimeout         = 30 * time.Second
	DefaultMaxRetries      = 3
	DefaultBaseDelay       = 100 * time.Millisecond
	DefaultMaxDelay        = 10 * time.Second
	DefaultMaxIdleConns    = 100
	DefaultIdleConnTimeout = 90 * time.Second
)

// Client is a resilient HTTP client with retry logic and circuit breaker.
type Client struct {
	httpClient     *http.Client
	baseURL        string
	defaultHeaders map[string]string
	middleware     []Middleware
	retryConfig    *RetryConfig
	circuitBreaker *CircuitBreaker
	logger         *slog.Logger
	mu             sync.RWMutex
}

// ClientOption is a function that configures a Client.
type ClientOption func(*Client)

// NewClient creates a new HTTP client with the given options.
func NewClient(opts ...ClientOption) *Client {
	c := &Client{
		httpClient: &http.Client{
			Timeout: DefaultTimeout,
			Transport: &http.Transport{
				MaxIdleConns:        DefaultMaxIdleConns,
				IdleConnTimeout:     DefaultIdleConnTimeout,
				DisableCompression:  false,
				DisableKeepAlives:   false,
				MaxIdleConnsPerHost: 10,
			},
		},
		defaultHeaders: make(map[string]string),
		middleware:     make([]Middleware, 0),
		retryConfig:    DefaultRetryConfig(),
		logger:         slog.Default(),
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

// WithTimeout sets the HTTP client timeout.
func WithTimeout(timeout time.Duration) ClientOption {
	return func(c *Client) {
		c.httpClient.Timeout = timeout
	}
}

// WithBaseURL sets the base URL for all requests.
func WithBaseURL(baseURL string) ClientOption {
	return func(c *Client) {
		c.baseURL = strings.TrimSuffix(baseURL, "/")
	}
}

// WithHeader adds a default header to all requests.
func WithHeader(key, value string) ClientOption {
	return func(c *Client) {
		c.defaultHeaders[key] = value
	}
}

// WithHeaders sets multiple default headers.
func WithHeaders(headers map[string]string) ClientOption {
	return func(c *Client) {
		maps.Copy(c.defaultHeaders, headers)
	}
}

// WithRetryConfig sets the retry configuration.
func WithRetryConfig(config *RetryConfig) ClientOption {
	return func(c *Client) {
		c.retryConfig = config
	}
}

// WithCircuitBreaker enables circuit breaker pattern.
func WithCircuitBreaker(cb *CircuitBreaker) ClientOption {
	return func(c *Client) {
		c.circuitBreaker = cb
	}
}

// WithMiddleware adds middleware to the client.
func WithMiddleware(mw ...Middleware) ClientOption {
	return func(c *Client) {
		c.middleware = append(c.middleware, mw...)
	}
}

// WithLogger sets the logger.
func WithLogger(logger *slog.Logger) ClientOption {
	return func(c *Client) {
		if logger != nil {
			c.logger = logger
		}
	}
}

// WithHTTPClient sets a custom HTTP client.
func WithHTTPClient(client *http.Client) ClientOption {
	return func(c *Client) {
		c.httpClient = client
	}
}

// Request represents an HTTP request.
type Request struct {
	Method      string
	URL         string
	Headers     map[string]string
	QueryParams map[string]string
	Body        any
	RawBody     []byte
}

// Response represents an HTTP response.
type Response struct {
	StatusCode int
	Status     string
	Headers    http.Header
	Body       []byte
	Duration   time.Duration
}

// Do executes an HTTP request with retry logic and circuit breaker.
func (c *Client) Do(ctx context.Context, req *Request) (*Response, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	// Check circuit breaker
	if c.circuitBreaker != nil && !c.circuitBreaker.Allow() {
		return nil, ErrCircuitOpen
	}

	var lastErr error
	var resp *Response

	maxAttempts := 1
	if c.retryConfig != nil {
		maxAttempts = c.retryConfig.MaxRetries
	}

	for attempt := 0; attempt < maxAttempts; attempt++ {
		if attempt > 0 {
			delay := c.calculateDelay(attempt)
			c.logger.Debug("retrying request",
				"attempt", attempt+1,
				"delay", delay,
				"method", req.Method,
				"url", req.URL,
			)
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(delay):
			}
		}

		resp, lastErr = c.doRequest(ctx, req)
		if lastErr == nil {
			if c.circuitBreaker != nil {
				c.circuitBreaker.RecordSuccess()
			}
			return resp, nil
		}

		// Record failure for circuit breaker
		if c.circuitBreaker != nil {
			c.circuitBreaker.RecordFailure()
		}

		// Check if error is retryable
		if !c.isRetryable(lastErr, resp) {
			break
		}

		c.logger.Warn("request failed, will retry",
			"attempt", attempt+1,
			"error", lastErr,
			"method", req.Method,
			"url", req.URL,
		)
	}

	return resp, lastErr
}

// doRequest performs a single HTTP request.
func (c *Client) doRequest(ctx context.Context, req *Request) (*Response, error) {
	start := time.Now()

	// Build URL
	reqURL, err := c.buildURL(req.URL, req.QueryParams)
	if err != nil {
		return nil, fmt.Errorf("building URL: %w", err)
	}

	// Build body
	var bodyReader io.Reader
	if req.RawBody != nil {
		bodyReader = bytes.NewReader(req.RawBody)
	} else if req.Body != nil {
		bodyBytes, err := json.Marshal(req.Body)
		if err != nil {
			return nil, fmt.Errorf("marshaling request body: %w", err)
		}
		bodyReader = bytes.NewReader(bodyBytes)
	}

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, req.Method, reqURL, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	// Set default headers
	c.mu.RLock()
	for k, v := range c.defaultHeaders {
		httpReq.Header.Set(k, v)
	}
	c.mu.RUnlock()

	// Set request-specific headers
	for k, v := range req.Headers {
		httpReq.Header.Set(k, v)
	}

	// Set Content-Type if body is present and not already set
	if req.Body != nil && httpReq.Header.Get("Content-Type") == "" {
		httpReq.Header.Set("Content-Type", "application/json")
	}

	// Apply middleware
	for _, mw := range c.middleware {
		httpReq, err = mw.ProcessRequest(httpReq)
		if err != nil {
			return nil, fmt.Errorf("middleware error: %w", err)
		}
	}

	// Execute request
	httpResp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("executing request: %w", err)
	}
	defer httpResp.Body.Close()

	// Read response body
	body, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body: %w", err)
	}

	duration := time.Since(start)

	resp := &Response{
		StatusCode: httpResp.StatusCode,
		Status:     httpResp.Status,
		Headers:    httpResp.Header,
		Body:       body,
		Duration:   duration,
	}

	// Apply middleware to response
	for _, mw := range c.middleware {
		resp, err = mw.ProcessResponse(resp)
		if err != nil {
			return resp, fmt.Errorf("middleware error: %w", err)
		}
	}

	// Check for HTTP errors
	if httpResp.StatusCode >= 400 {
		return resp, NewHTTPError(httpResp.StatusCode, httpResp.Status, string(body))
	}

	return resp, nil
}

// buildURL constructs the full URL with query parameters.
func (c *Client) buildURL(path string, queryParams map[string]string) (string, error) {
	var fullURL string
	if c.baseURL != "" && !strings.HasPrefix(path, "http") {
		fullURL = c.baseURL + "/" + strings.TrimPrefix(path, "/")
	} else {
		fullURL = path
	}

	if len(queryParams) == 0 {
		return fullURL, nil
	}

	u, err := url.Parse(fullURL)
	if err != nil {
		return "", err
	}

	q := u.Query()
	for k, v := range queryParams {
		q.Set(k, v)
	}
	u.RawQuery = q.Encode()

	return u.String(), nil
}

// calculateDelay calculates the delay for retry with exponential backoff and jitter.
func (c *Client) calculateDelay(attempt int) time.Duration {
	if c.retryConfig == nil {
		return DefaultBaseDelay
	}
	return c.retryConfig.CalculateDelay(attempt)
}

// isRetryable determines if an error or response should trigger a retry.
func (c *Client) isRetryable(err error, resp *Response) bool {
	if c.retryConfig != nil && c.retryConfig.ShouldRetry != nil {
		return c.retryConfig.ShouldRetry(err, resp)
	}

	// Default retry logic
	if err != nil {
		return IsRetryableError(err)
	}

	if resp != nil {
		// Retry on server errors and rate limiting
		return resp.StatusCode >= 500 || resp.StatusCode == 429 || resp.StatusCode == 408
	}

	return false
}

// Get performs a GET request.
func (c *Client) Get(ctx context.Context, path string, opts ...RequestOption) (*Response, error) {
	req := &Request{
		Method:      http.MethodGet,
		URL:         path,
		Headers:     make(map[string]string),
		QueryParams: make(map[string]string),
	}
	for _, opt := range opts {
		opt(req)
	}
	return c.Do(ctx, req)
}

// Post performs a POST request.
func (c *Client) Post(ctx context.Context, path string, body any, opts ...RequestOption) (*Response, error) {
	req := &Request{
		Method:      http.MethodPost,
		URL:         path,
		Body:        body,
		Headers:     make(map[string]string),
		QueryParams: make(map[string]string),
	}
	for _, opt := range opts {
		opt(req)
	}
	return c.Do(ctx, req)
}

// Put performs a PUT request.
func (c *Client) Put(ctx context.Context, path string, body any, opts ...RequestOption) (*Response, error) {
	req := &Request{
		Method:      http.MethodPut,
		URL:         path,
		Body:        body,
		Headers:     make(map[string]string),
		QueryParams: make(map[string]string),
	}
	for _, opt := range opts {
		opt(req)
	}
	return c.Do(ctx, req)
}

// Patch performs a PATCH request.
func (c *Client) Patch(ctx context.Context, path string, body any, opts ...RequestOption) (*Response, error) {
	req := &Request{
		Method:      http.MethodPatch,
		URL:         path,
		Body:        body,
		Headers:     make(map[string]string),
		QueryParams: make(map[string]string),
	}
	for _, opt := range opts {
		opt(req)
	}
	return c.Do(ctx, req)
}

// Delete performs a DELETE request.
func (c *Client) Delete(ctx context.Context, path string, opts ...RequestOption) (*Response, error) {
	req := &Request{
		Method:      http.MethodDelete,
		URL:         path,
		Headers:     make(map[string]string),
		QueryParams: make(map[string]string),
	}
	for _, opt := range opts {
		opt(req)
	}
	return c.Do(ctx, req)
}

// RequestOption is a function that configures a Request.
type RequestOption func(*Request)

// WithRequestHeader adds a header to the request.
func WithRequestHeader(key, value string) RequestOption {
	return func(r *Request) {
		if r.Headers == nil {
			r.Headers = make(map[string]string)
		}
		r.Headers[key] = value
	}
}

// WithRequestHeaders sets multiple headers on the request.
func WithRequestHeaders(headers map[string]string) RequestOption {
	return func(r *Request) {
		if r.Headers == nil {
			r.Headers = make(map[string]string)
		}
		maps.Copy(r.Headers, headers)
	}
}

// WithQueryParam adds a query parameter to the request.
func WithQueryParam(key, value string) RequestOption {
	return func(r *Request) {
		if r.QueryParams == nil {
			r.QueryParams = make(map[string]string)
		}
		r.QueryParams[key] = value
	}
}

// WithQueryParams sets multiple query parameters on the request.
func WithQueryParams(params map[string]string) RequestOption {
	return func(r *Request) {
		if r.QueryParams == nil {
			r.QueryParams = make(map[string]string)
		}
		maps.Copy(r.QueryParams, params)
	}
}

// WithRawBody sets the raw body bytes on the request.
func WithRawBody(body []byte) RequestOption {
	return func(r *Request) {
		r.RawBody = body
	}
}

// JSON attempts to unmarshal the response body into the given interface.
func (r *Response) JSON(v any) error {
	return json.Unmarshal(r.Body, v)
}

// String returns the response body as a string.
func (r *Response) String() string {
	return string(r.Body)
}

// IsSuccess returns true if the response status code indicates success (2xx).
func (r *Response) IsSuccess() bool {
	return r.StatusCode >= 200 && r.StatusCode < 300
}

// IsError returns true if the response status code indicates an error (4xx or 5xx).
func (r *Response) IsError() bool {
	return r.StatusCode >= 400
}
