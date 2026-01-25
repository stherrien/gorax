package http

import (
	"log/slog"
	"net/http"
	"time"
)

// Middleware defines an interface for HTTP request/response middleware.
type Middleware interface {
	// ProcessRequest modifies the request before it is sent.
	ProcessRequest(req *http.Request) (*http.Request, error)

	// ProcessResponse modifies the response after it is received.
	ProcessResponse(resp *Response) (*Response, error)
}

// MiddlewareFunc is a function that implements the Middleware interface.
type MiddlewareFunc struct {
	RequestFn  func(req *http.Request) (*http.Request, error)
	ResponseFn func(resp *Response) (*Response, error)
}

// ProcessRequest implements Middleware.
func (m *MiddlewareFunc) ProcessRequest(req *http.Request) (*http.Request, error) {
	if m.RequestFn != nil {
		return m.RequestFn(req)
	}
	return req, nil
}

// ProcessResponse implements Middleware.
func (m *MiddlewareFunc) ProcessResponse(resp *Response) (*Response, error) {
	if m.ResponseFn != nil {
		return m.ResponseFn(resp)
	}
	return resp, nil
}

// LoggingMiddleware logs HTTP requests and responses.
type LoggingMiddleware struct {
	logger    *slog.Logger
	logBody   bool
	startTime time.Time
}

// NewLoggingMiddleware creates a new logging middleware.
func NewLoggingMiddleware(logger *slog.Logger, logBody bool) *LoggingMiddleware {
	if logger == nil {
		logger = slog.Default()
	}
	return &LoggingMiddleware{
		logger:  logger,
		logBody: logBody,
	}
}

// ProcessRequest logs the outgoing request.
func (m *LoggingMiddleware) ProcessRequest(req *http.Request) (*http.Request, error) {
	m.startTime = time.Now()
	m.logger.Debug("HTTP request",
		"method", req.Method,
		"url", req.URL.String(),
		"headers", sanitizeHeaders(req.Header),
	)
	return req, nil
}

// ProcessResponse logs the incoming response.
func (m *LoggingMiddleware) ProcessResponse(resp *Response) (*Response, error) {
	duration := time.Since(m.startTime)

	attrs := []any{
		"status_code", resp.StatusCode,
		"status", resp.Status,
		"duration_ms", duration.Milliseconds(),
	}

	if m.logBody {
		attrs = append(attrs, "body_length", len(resp.Body))
	}

	m.logger.Debug("HTTP response", attrs...)
	return resp, nil
}

// sanitizeHeaders removes sensitive headers for logging.
func sanitizeHeaders(headers http.Header) map[string]string {
	result := make(map[string]string)
	sensitiveHeaders := map[string]bool{
		"Authorization":   true,
		"X-Api-Key":       true,
		"X-Auth-Token":    true,
		"Cookie":          true,
		"Set-Cookie":      true,
		"X-Access-Token":  true,
		"X-Refresh-Token": true,
	}

	for key, values := range headers {
		if sensitiveHeaders[key] {
			result[key] = "[REDACTED]"
		} else if len(values) > 0 {
			result[key] = values[0]
		}
	}
	return result
}

// HeadersMiddleware adds custom headers to requests.
type HeadersMiddleware struct {
	headers map[string]string
}

// NewHeadersMiddleware creates a new headers middleware.
func NewHeadersMiddleware(headers map[string]string) *HeadersMiddleware {
	return &HeadersMiddleware{headers: headers}
}

// ProcessRequest adds headers to the request.
func (m *HeadersMiddleware) ProcessRequest(req *http.Request) (*http.Request, error) {
	for key, value := range m.headers {
		req.Header.Set(key, value)
	}
	return req, nil
}

// ProcessResponse returns the response unchanged.
func (m *HeadersMiddleware) ProcessResponse(resp *Response) (*Response, error) {
	return resp, nil
}

// UserAgentMiddleware sets the User-Agent header.
type UserAgentMiddleware struct {
	userAgent string
}

// NewUserAgentMiddleware creates a new user-agent middleware.
func NewUserAgentMiddleware(userAgent string) *UserAgentMiddleware {
	return &UserAgentMiddleware{userAgent: userAgent}
}

// ProcessRequest sets the User-Agent header.
func (m *UserAgentMiddleware) ProcessRequest(req *http.Request) (*http.Request, error) {
	req.Header.Set("User-Agent", m.userAgent)
	return req, nil
}

// ProcessResponse returns the response unchanged.
func (m *UserAgentMiddleware) ProcessResponse(resp *Response) (*Response, error) {
	return resp, nil
}

// ContentTypeMiddleware ensures Content-Type is set for requests with bodies.
type ContentTypeMiddleware struct {
	contentType string
}

// NewContentTypeMiddleware creates a new content-type middleware.
func NewContentTypeMiddleware(contentType string) *ContentTypeMiddleware {
	return &ContentTypeMiddleware{contentType: contentType}
}

// ProcessRequest sets the Content-Type header if not already set.
func (m *ContentTypeMiddleware) ProcessRequest(req *http.Request) (*http.Request, error) {
	if req.Body != nil && req.Header.Get("Content-Type") == "" {
		req.Header.Set("Content-Type", m.contentType)
	}
	return req, nil
}

// ProcessResponse returns the response unchanged.
func (m *ContentTypeMiddleware) ProcessResponse(resp *Response) (*Response, error) {
	return resp, nil
}

// ChainMiddleware creates a middleware that chains multiple middleware together.
func ChainMiddleware(middlewares ...Middleware) Middleware {
	return &chainedMiddleware{middlewares: middlewares}
}

type chainedMiddleware struct {
	middlewares []Middleware
}

// ProcessRequest applies all middleware request processors in order.
func (c *chainedMiddleware) ProcessRequest(req *http.Request) (*http.Request, error) {
	var err error
	for _, mw := range c.middlewares {
		req, err = mw.ProcessRequest(req)
		if err != nil {
			return req, err
		}
	}
	return req, nil
}

// ProcessResponse applies all middleware response processors in reverse order.
func (c *chainedMiddleware) ProcessResponse(resp *Response) (*Response, error) {
	var err error
	for i := len(c.middlewares) - 1; i >= 0; i-- {
		resp, err = c.middlewares[i].ProcessResponse(resp)
		if err != nil {
			return resp, err
		}
	}
	return resp, nil
}
