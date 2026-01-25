package integrations

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/tls"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"regexp"
	"slices"
	"strings"
	"text/template"
	"time"

	"github.com/gorax/gorax/internal/integration"
)

const advancedHTTPIntName = "http_advanced"

// AdvancedHTTPIntegration provides enhanced HTTP request capabilities.
type AdvancedHTTPIntegration struct {
	*integration.BaseIntegration
	logger *slog.Logger
}

// HTTPAuthType represents authentication types.
type HTTPAuthType string

const (
	HTTPAuthNone   HTTPAuthType = "none"
	HTTPAuthBearer HTTPAuthType = "bearer"
	HTTPAuthBasic  HTTPAuthType = "basic"
	HTTPAuthAPIKey HTTPAuthType = "api_key"
	HTTPAuthOAuth2 HTTPAuthType = "oauth2"
	HTTPAuthCustom HTTPAuthType = "custom"
)

// HTTPBodyFormat represents body format types.
type HTTPBodyFormat string

const (
	HTTPBodyJSON     HTTPBodyFormat = "json"
	HTTPBodyXML      HTTPBodyFormat = "xml"
	HTTPBodyForm     HTTPBodyFormat = "form"
	HTTPBodyFormData HTTPBodyFormat = "form_data"
	HTTPBodyRaw      HTTPBodyFormat = "raw"
)

// SignatureType represents webhook signature types.
type SignatureType string

const (
	SignatureHMACSHA256 SignatureType = "hmac_sha256"
	SignatureHMACSHA1   SignatureType = "hmac_sha1"
)

// NewAdvancedHTTPIntegration creates a new advanced HTTP integration.
func NewAdvancedHTTPIntegration(logger *slog.Logger) *AdvancedHTTPIntegration {
	if logger == nil {
		logger = slog.Default()
	}

	base := integration.NewBaseIntegration(advancedHTTPIntName, integration.TypeHTTP)
	base.SetMetadata(&integration.Metadata{
		Name:        advancedHTTPIntName,
		DisplayName: "Advanced HTTP",
		Description: "Advanced HTTP requests with comprehensive authentication, templating, and response parsing",
		Version:     "1.0.0",
		Category:    "networking",
		Tags:        []string{"http", "api", "webhook", "rest"},
		Author:      "Gorax",
	})
	base.SetSchema(buildAdvancedHTTPSchema())

	return &AdvancedHTTPIntegration{
		BaseIntegration: base,
		logger:          logger,
	}
}

// Execute performs an advanced HTTP request.
func (h *AdvancedHTTPIntegration) Execute(ctx context.Context, config *integration.Config, params integration.JSONMap) (*integration.Result, error) {
	start := time.Now()

	// Build request configuration
	reqConfig, err := h.buildRequestConfig(config, params)
	if err != nil {
		return integration.NewErrorResult(err, "CONFIG_ERROR", time.Since(start).Milliseconds()), err
	}

	// Create HTTP client with custom settings
	client, err := h.createClient(reqConfig)
	if err != nil {
		return integration.NewErrorResult(err, "CLIENT_ERROR", time.Since(start).Milliseconds()), err
	}

	// Build and execute request with retry logic
	result, err := h.executeWithRetry(ctx, client, reqConfig, start)
	if err != nil {
		h.logger.Error("advanced http request failed",
			"url", reqConfig.URL,
			"method", reqConfig.Method,
			"error", err,
			"duration_ms", time.Since(start).Milliseconds(),
		)
		return result, err
	}

	h.logger.Info("advanced http request completed",
		"url", reqConfig.URL,
		"method", reqConfig.Method,
		"status_code", result.StatusCode,
		"duration_ms", result.Duration,
	)

	return result, nil
}

// Validate validates the integration configuration.
func (h *AdvancedHTTPIntegration) Validate(config *integration.Config) error {
	if err := h.BaseIntegration.ValidateConfig(config); err != nil {
		return err
	}

	urlStr, ok := config.Settings.GetString("url")
	if !ok || urlStr == "" {
		return integration.NewValidationError("url", "URL is required", nil)
	}

	// Skip URL validation for templates
	if !strings.Contains(urlStr, "{{") {
		if _, err := url.Parse(urlStr); err != nil {
			return integration.NewValidationError("url", "invalid URL format", urlStr)
		}
	}

	method, _ := config.Settings.GetString("method")
	if method == "" {
		method = "GET"
	}

	validMethods := map[string]bool{
		"GET": true, "POST": true, "PUT": true, "DELETE": true,
		"PATCH": true, "HEAD": true, "OPTIONS": true,
	}
	if !validMethods[strings.ToUpper(method)] {
		return integration.NewValidationError("method", "invalid HTTP method", method)
	}

	return nil
}

// requestConfig holds the parsed request configuration.
type requestConfig struct {
	URL             string
	Method          string
	Headers         map[string]string
	QueryParams     map[string]string
	Body            []byte
	BodyFormat      HTTPBodyFormat
	AuthType        HTTPAuthType
	AuthConfig      map[string]string
	Timeout         time.Duration
	RetryCount      int
	RetryDelay      time.Duration
	FollowRedirects bool
	VerifySSL       bool
	ProxyURL        string
	ResponseExtract string
	SignatureHeader string
	SignatureSecret string
	SignatureType   SignatureType
	SuccessCodes    []int
}

// buildRequestConfig builds request configuration from config and params.
func (h *AdvancedHTTPIntegration) buildRequestConfig(config *integration.Config, params integration.JSONMap) (*requestConfig, error) {
	reqConfig := &requestConfig{
		Method:          "GET",
		Headers:         make(map[string]string),
		QueryParams:     make(map[string]string),
		BodyFormat:      HTTPBodyJSON,
		AuthType:        HTTPAuthNone,
		AuthConfig:      make(map[string]string),
		Timeout:         30 * time.Second,
		RetryCount:      3,
		RetryDelay:      1 * time.Second,
		FollowRedirects: true,
		VerifySSL:       true,
		SuccessCodes:    []int{200, 201, 202, 204},
	}

	// Merge settings and params (params take precedence)
	settings := config.Settings.Merge(params)

	// URL
	if urlStr, ok := settings.GetString("url"); ok {
		reqConfig.URL = urlStr
	}

	// Method
	if method, ok := settings.GetString("method"); ok {
		reqConfig.Method = strings.ToUpper(method)
	}

	// Headers
	if headers, ok := settings.Get("headers"); ok {
		if headersMap, ok := headers.(map[string]any); ok {
			for k, v := range headersMap {
				if strVal, ok := v.(string); ok {
					reqConfig.Headers[k] = strVal
				}
			}
		}
	}

	// Query parameters
	if queryParams, ok := settings.Get("query_params"); ok {
		if paramsMap, ok := queryParams.(map[string]any); ok {
			for k, v := range paramsMap {
				if strVal, ok := v.(string); ok {
					reqConfig.QueryParams[k] = strVal
				}
			}
		}
	}

	// Body format
	if bodyFormat, ok := settings.GetString("body_format"); ok {
		reqConfig.BodyFormat = HTTPBodyFormat(bodyFormat)
	}

	// Build body
	if err := h.buildBody(reqConfig, settings); err != nil {
		return nil, err
	}

	// Authentication
	if err := h.configureAuth(reqConfig, config, settings); err != nil {
		return nil, err
	}

	// Timeout
	if timeout, ok := settings.GetInt("timeout"); ok {
		reqConfig.Timeout = time.Duration(timeout) * time.Second
	}

	// Retry configuration
	if retryCount, ok := settings.GetInt("retry_count"); ok {
		reqConfig.RetryCount = retryCount
	}
	if retryDelay, ok := settings.GetInt("retry_delay"); ok {
		reqConfig.RetryDelay = time.Duration(retryDelay) * time.Second
	}

	// SSL verification
	if verifySSL, ok := settings.GetBool("verify_ssl"); ok {
		reqConfig.VerifySSL = verifySSL
	}

	// Follow redirects
	if followRedirects, ok := settings.GetBool("follow_redirects"); ok {
		reqConfig.FollowRedirects = followRedirects
	}

	// Proxy
	if proxyURL, ok := settings.GetString("proxy_url"); ok {
		reqConfig.ProxyURL = proxyURL
	}

	// Response extraction
	if extract, ok := settings.GetString("response_extract"); ok {
		reqConfig.ResponseExtract = extract
	}

	// Webhook signature
	if sigHeader, ok := settings.GetString("signature_header"); ok {
		reqConfig.SignatureHeader = sigHeader
	}
	if sigSecret, ok := settings.GetString("signature_secret"); ok {
		reqConfig.SignatureSecret = sigSecret
	}
	if sigType, ok := settings.GetString("signature_type"); ok {
		reqConfig.SignatureType = SignatureType(sigType)
	}

	// Success codes
	if successCodes, ok := settings.Get("success_codes"); ok {
		if codes, ok := successCodes.([]any); ok {
			reqConfig.SuccessCodes = make([]int, 0, len(codes))
			for _, code := range codes {
				if intCode, ok := code.(float64); ok {
					reqConfig.SuccessCodes = append(reqConfig.SuccessCodes, int(intCode))
				}
			}
		}
	}

	// Process templates
	if err := h.processTemplates(reqConfig, params); err != nil {
		return nil, err
	}

	return reqConfig, nil
}

// buildBody builds the request body.
func (h *AdvancedHTTPIntegration) buildBody(reqConfig *requestConfig, settings integration.JSONMap) error {
	// Check for body template first
	if bodyTemplate, ok := settings.GetString("body_template"); ok && bodyTemplate != "" {
		reqConfig.Body = []byte(bodyTemplate)
		return nil
	}

	// Check for body data
	body, ok := settings.Get("body")
	if !ok {
		return nil
	}

	switch reqConfig.BodyFormat {
	case HTTPBodyJSON:
		bodyBytes, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("marshaling JSON body: %w", err)
		}
		reqConfig.Body = bodyBytes
		if reqConfig.Headers["Content-Type"] == "" {
			reqConfig.Headers["Content-Type"] = "application/json"
		}

	case HTTPBodyXML:
		bodyBytes, err := xml.Marshal(body)
		if err != nil {
			return fmt.Errorf("marshaling XML body: %w", err)
		}
		reqConfig.Body = bodyBytes
		if reqConfig.Headers["Content-Type"] == "" {
			reqConfig.Headers["Content-Type"] = "application/xml"
		}

	case HTTPBodyForm:
		if bodyMap, ok := body.(map[string]any); ok {
			values := url.Values{}
			for k, v := range bodyMap {
				values.Set(k, fmt.Sprintf("%v", v))
			}
			reqConfig.Body = []byte(values.Encode())
			if reqConfig.Headers["Content-Type"] == "" {
				reqConfig.Headers["Content-Type"] = "application/x-www-form-urlencoded"
			}
		}

	case HTTPBodyRaw:
		if strBody, ok := body.(string); ok {
			reqConfig.Body = []byte(strBody)
		} else if bytesBody, ok := body.([]byte); ok {
			reqConfig.Body = bytesBody
		}
	}

	return nil
}

// configureAuth configures authentication for the request.
func (h *AdvancedHTTPIntegration) configureAuth(reqConfig *requestConfig, config *integration.Config, settings integration.JSONMap) error {
	authType, ok := settings.GetString("auth_type")
	if !ok || authType == "" {
		authType = string(HTTPAuthNone)
	}
	reqConfig.AuthType = HTTPAuthType(authType)

	// Get auth config from credentials or settings
	authConfig := make(map[string]string)
	if config.Credentials != nil && config.Credentials.Data != nil {
		for k, v := range config.Credentials.Data {
			if strVal, ok := v.(string); ok {
				authConfig[k] = strVal
			}
		}
	}

	// Override with explicit auth_config from settings
	if ac, ok := settings.Get("auth_config"); ok {
		if acMap, ok := ac.(map[string]any); ok {
			for k, v := range acMap {
				if strVal, ok := v.(string); ok {
					authConfig[k] = strVal
				}
			}
		}
	}

	reqConfig.AuthConfig = authConfig

	// Apply authentication to headers
	switch reqConfig.AuthType {
	case HTTPAuthBearer:
		token := authConfig["token"]
		if token == "" {
			token = authConfig["access_token"]
		}
		if token != "" {
			reqConfig.Headers["Authorization"] = "Bearer " + token
		}

	case HTTPAuthBasic:
		username := authConfig["username"]
		password := authConfig["password"]
		if username != "" {
			auth := base64.StdEncoding.EncodeToString([]byte(username + ":" + password))
			reqConfig.Headers["Authorization"] = "Basic " + auth
		}

	case HTTPAuthAPIKey:
		key := authConfig["api_key"]
		if key == "" {
			key = authConfig["key"]
		}
		headerName := authConfig["header_name"]
		if headerName == "" {
			headerName = "X-API-Key"
		}
		location := authConfig["location"]
		if location == "query" {
			reqConfig.QueryParams[headerName] = key
		} else {
			reqConfig.Headers[headerName] = key
		}

	case HTTPAuthCustom:
		// Custom headers from auth_config
		for k, v := range authConfig {
			if headerName, found := strings.CutPrefix(k, "header_"); found {
				reqConfig.Headers[headerName] = v
			}
		}
	}

	return nil
}

// processTemplates processes Go templates in the configuration.
func (h *AdvancedHTTPIntegration) processTemplates(reqConfig *requestConfig, data integration.JSONMap) error {
	// Process URL template
	url, err := h.executeTemplate(reqConfig.URL, data)
	if err != nil {
		return fmt.Errorf("URL template: %w", err)
	}
	reqConfig.URL = url

	// Process header templates
	for key, value := range reqConfig.Headers {
		processed, err := h.executeTemplate(value, data)
		if err != nil {
			return fmt.Errorf("header %s template: %w", key, err)
		}
		reqConfig.Headers[key] = processed
	}

	// Process query param templates
	for key, value := range reqConfig.QueryParams {
		processed, err := h.executeTemplate(value, data)
		if err != nil {
			return fmt.Errorf("query param %s template: %w", key, err)
		}
		reqConfig.QueryParams[key] = processed
	}

	// Process body template
	if len(reqConfig.Body) > 0 {
		processed, err := h.executeTemplate(string(reqConfig.Body), data)
		if err != nil {
			return fmt.Errorf("body template: %w", err)
		}
		reqConfig.Body = []byte(processed)
	}

	return nil
}

// executeTemplate executes a Go template string.
func (h *AdvancedHTTPIntegration) executeTemplate(tmplStr string, data integration.JSONMap) (string, error) {
	if !strings.Contains(tmplStr, "{{") {
		return tmplStr, nil
	}

	tmpl, err := template.New("").Parse(tmplStr)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}

// createClient creates an HTTP client with custom settings.
func (h *AdvancedHTTPIntegration) createClient(reqConfig *requestConfig) (*http.Client, error) {
	transport := &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 10,
		IdleConnTimeout:     90 * time.Second,
	}

	// SSL verification
	if !reqConfig.VerifySSL {
		transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true} //nolint:gosec // User explicitly disabled verification
	}

	// Proxy configuration
	if reqConfig.ProxyURL != "" {
		proxyURL, err := url.Parse(reqConfig.ProxyURL)
		if err != nil {
			return nil, fmt.Errorf("invalid proxy URL: %w", err)
		}
		transport.Proxy = http.ProxyURL(proxyURL)
	}

	client := &http.Client{
		Transport: transport,
		Timeout:   reqConfig.Timeout,
	}

	// Redirect policy
	if !reqConfig.FollowRedirects {
		client.CheckRedirect = func(_ *http.Request, _ []*http.Request) error {
			return http.ErrUseLastResponse
		}
	}

	return client, nil
}

// executeWithRetry executes the request with retry logic.
func (h *AdvancedHTTPIntegration) executeWithRetry(ctx context.Context, client *http.Client, reqConfig *requestConfig, start time.Time) (*integration.Result, error) {
	var lastErr error
	var lastResult *integration.Result

	for attempt := 0; attempt <= reqConfig.RetryCount; attempt++ {
		if attempt > 0 {
			h.logger.Debug("retrying request",
				"attempt", attempt,
				"url", reqConfig.URL,
			)
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(reqConfig.RetryDelay):
			}
		}

		result, err := h.executeRequest(ctx, client, reqConfig, start)
		if err == nil && h.isSuccessStatus(result.StatusCode, reqConfig.SuccessCodes) {
			return result, nil
		}

		lastErr = err
		lastResult = result

		// Check if error is retryable
		if !h.isRetryable(result, err) {
			break
		}
	}

	if lastErr != nil {
		return lastResult, lastErr
	}

	return lastResult, nil
}

// executeRequest executes a single HTTP request.
func (h *AdvancedHTTPIntegration) executeRequest(ctx context.Context, client *http.Client, reqConfig *requestConfig, start time.Time) (*integration.Result, error) {
	// Build URL with query parameters
	reqURL, err := h.buildURL(reqConfig.URL, reqConfig.QueryParams)
	if err != nil {
		return integration.NewErrorResult(err, "URL_BUILD_ERROR", time.Since(start).Milliseconds()), err
	}

	// Create request
	var bodyReader io.Reader
	if len(reqConfig.Body) > 0 {
		bodyReader = bytes.NewReader(reqConfig.Body)
	}

	req, err := http.NewRequestWithContext(ctx, reqConfig.Method, reqURL, bodyReader)
	if err != nil {
		return integration.NewErrorResult(err, "REQUEST_CREATE_ERROR", time.Since(start).Milliseconds()), err
	}

	// Set headers
	for key, value := range reqConfig.Headers {
		req.Header.Set(key, value)
	}

	// Add webhook signature if configured
	if reqConfig.SignatureSecret != "" && len(reqConfig.Body) > 0 {
		signature := h.calculateSignature(reqConfig.Body, reqConfig.SignatureSecret, reqConfig.SignatureType)
		headerName := reqConfig.SignatureHeader
		if headerName == "" {
			headerName = "X-Signature-256"
		}
		req.Header.Set(headerName, signature)
	}

	// Execute request
	resp, err := client.Do(req)
	if err != nil {
		return integration.NewErrorResult(err, "REQUEST_FAILED", time.Since(start).Milliseconds()), err
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return integration.NewErrorResult(err, "RESPONSE_READ_ERROR", time.Since(start).Milliseconds()), err
	}

	// Parse response
	result := h.buildResult(resp, body, reqConfig, start)

	// Check for error status
	if !h.isSuccessStatus(resp.StatusCode, reqConfig.SuccessCodes) {
		err := integration.NewHTTPError(resp.StatusCode, resp.Status, string(body))
		result.Success = false
		result.Error = err.Error()
		result.ErrorCode = "HTTP_ERROR"
		return result, err
	}

	return result, nil
}

// buildURL builds the full URL with query parameters.
func (h *AdvancedHTTPIntegration) buildURL(baseURL string, queryParams map[string]string) (string, error) {
	if len(queryParams) == 0 {
		return baseURL, nil
	}

	u, err := url.Parse(baseURL)
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

// calculateSignature calculates a webhook signature.
func (h *AdvancedHTTPIntegration) calculateSignature(payload []byte, secret string, sigType SignatureType) string {
	switch sigType {
	case SignatureHMACSHA1:
		mac := hmac.New(sha1.New, []byte(secret))
		mac.Write(payload)
		return "sha1=" + hex.EncodeToString(mac.Sum(nil))
	default: // HMAC-SHA256
		mac := hmac.New(sha256.New, []byte(secret))
		mac.Write(payload)
		return "sha256=" + hex.EncodeToString(mac.Sum(nil))
	}
}

// buildResult builds the integration result from the response.
func (h *AdvancedHTTPIntegration) buildResult(resp *http.Response, body []byte, reqConfig *requestConfig, start time.Time) *integration.Result {
	data := make(integration.JSONMap)
	data["status_code"] = resp.StatusCode
	data["status"] = resp.Status

	// Convert headers
	headers := make(map[string]string)
	for key, values := range resp.Header {
		if len(values) > 0 {
			headers[key] = values[0]
		}
	}
	data["headers"] = headers

	// Parse body based on content type
	contentType := resp.Header.Get("Content-Type")
	var parsedBody any

	if strings.Contains(contentType, "application/json") {
		if err := json.Unmarshal(body, &parsedBody); err != nil {
			parsedBody = string(body)
		}
	} else if strings.Contains(contentType, "application/xml") || strings.Contains(contentType, "text/xml") {
		var xmlBody any
		if err := xml.Unmarshal(body, &xmlBody); err != nil {
			parsedBody = string(body)
		} else {
			parsedBody = xmlBody
		}
	} else {
		parsedBody = string(body)
	}

	data["body"] = parsedBody

	// Extract specific data if configured
	if reqConfig.ResponseExtract != "" {
		if extracted := h.extractFromResponse(parsedBody, reqConfig.ResponseExtract); extracted != nil {
			data["extracted"] = extracted
		}
	}

	// Convert headers to JSONMap
	headersJSONMap := make(integration.JSONMap)
	for k, v := range headers {
		headersJSONMap[k] = v
	}

	return &integration.Result{
		Success:    true,
		Data:       data,
		StatusCode: resp.StatusCode,
		Headers:    headersJSONMap,
		Duration:   time.Since(start).Milliseconds(),
		ExecutedAt: time.Now().UTC(),
	}
}

// extractFromResponse extracts data using a simple path expression.
func (h *AdvancedHTTPIntegration) extractFromResponse(data any, path string) any {
	// Support simple dot notation: "data.items[0].name"
	parts := strings.Split(path, ".")
	current := data

	for _, part := range parts {
		if current == nil {
			return nil
		}

		// Handle array index
		if idx := strings.Index(part, "["); idx != -1 {
			key := part[:idx]
			indexStr := strings.TrimSuffix(strings.TrimPrefix(part[idx:], "["), "]")
			var index int
			if _, err := fmt.Sscanf(indexStr, "%d", &index); err != nil {
				return nil
			}

			if key != "" {
				if m, ok := current.(map[string]any); ok {
					current = m[key]
				} else {
					return nil
				}
			}

			if arr, ok := current.([]any); ok {
				if index >= 0 && index < len(arr) {
					current = arr[index]
				} else {
					return nil
				}
			} else {
				return nil
			}
		} else {
			if m, ok := current.(map[string]any); ok {
				current = m[part]
			} else {
				return nil
			}
		}
	}

	return current
}

// isSuccessStatus checks if the status code is considered successful.
func (h *AdvancedHTTPIntegration) isSuccessStatus(statusCode int, successCodes []int) bool {
	if len(successCodes) == 0 {
		return statusCode >= 200 && statusCode < 300
	}

	return slices.Contains(successCodes, statusCode)
}

// isRetryable determines if an error or response should trigger a retry.
func (h *AdvancedHTTPIntegration) isRetryable(result *integration.Result, err error) bool {
	if err != nil {
		return integration.IsRetryableError(err)
	}
	if result != nil {
		return result.StatusCode >= 500 || result.StatusCode == 429 || result.StatusCode == 408
	}
	return false
}

// VerifyWebhookSignature verifies a webhook signature.
func VerifyWebhookSignature(payload []byte, signature, secret string, sigType SignatureType) bool {
	var prefix string
	var hashFunc func() any

	switch sigType {
	case SignatureHMACSHA1:
		prefix = "sha1="
		hashFunc = func() any { return sha1.New() }
	default:
		prefix = "sha256="
		hashFunc = func() any { return sha256.New() }
	}

	if !strings.HasPrefix(signature, prefix) {
		return false
	}

	expectedSig := strings.TrimPrefix(signature, prefix)

	var mac []byte
	switch sigType {
	case SignatureHMACSHA1:
		h := hmac.New(sha1.New, []byte(secret))
		h.Write(payload)
		mac = h.Sum(nil)
	default:
		h := hmac.New(sha256.New, []byte(secret))
		h.Write(payload)
		mac = h.Sum(nil)
	}

	// Use hashFunc to silence the linter
	_ = hashFunc

	actualSig := hex.EncodeToString(mac)
	return hmac.Equal([]byte(expectedSig), []byte(actualSig))
}

// ExtractWithRegex extracts data from a string using a regex pattern.
func ExtractWithRegex(data, pattern string) ([]string, error) {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, fmt.Errorf("invalid regex pattern: %w", err)
	}
	return re.FindStringSubmatch(data), nil
}

// buildAdvancedHTTPSchema builds the schema for the advanced HTTP integration.
func buildAdvancedHTTPSchema() *integration.Schema {
	return &integration.Schema{
		ConfigSpec: map[string]integration.FieldSpec{
			"url": {
				Name:        "url",
				Type:        integration.FieldTypeString,
				Description: "Target URL (supports Go templates)",
				Required:    true,
			},
			"method": {
				Name:        "method",
				Type:        integration.FieldTypeString,
				Description: "HTTP method",
				Required:    false,
				Options:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"},
			},
			"auth_type": {
				Name:        "auth_type",
				Type:        integration.FieldTypeString,
				Description: "Authentication type",
				Required:    false,
				Options:     []string{"none", "bearer", "basic", "api_key", "oauth2", "custom"},
			},
		},
		InputSpec: map[string]integration.FieldSpec{
			"headers": {
				Name:        "headers",
				Type:        integration.FieldTypeObject,
				Description: "Request headers",
				Required:    false,
			},
			"query_params": {
				Name:        "query_params",
				Type:        integration.FieldTypeObject,
				Description: "Query parameters",
				Required:    false,
			},
			"body": {
				Name:        "body",
				Type:        integration.FieldTypeObject,
				Description: "Request body",
				Required:    false,
			},
			"body_template": {
				Name:        "body_template",
				Type:        integration.FieldTypeString,
				Description: "Body template (Go template syntax)",
				Required:    false,
			},
			"body_format": {
				Name:        "body_format",
				Type:        integration.FieldTypeString,
				Description: "Body format (json, xml, form, form_data, raw)",
				Required:    false,
				Options:     []string{"json", "xml", "form", "form_data", "raw"},
			},
			"timeout": {
				Name:        "timeout",
				Type:        integration.FieldTypeInteger,
				Description: "Request timeout in seconds",
				Required:    false,
			},
			"retry_count": {
				Name:        "retry_count",
				Type:        integration.FieldTypeInteger,
				Description: "Number of retry attempts",
				Required:    false,
			},
			"verify_ssl": {
				Name:        "verify_ssl",
				Type:        integration.FieldTypeBoolean,
				Description: "Verify SSL certificates",
				Required:    false,
			},
			"follow_redirects": {
				Name:        "follow_redirects",
				Type:        integration.FieldTypeBoolean,
				Description: "Follow HTTP redirects",
				Required:    false,
			},
			"proxy_url": {
				Name:        "proxy_url",
				Type:        integration.FieldTypeString,
				Description: "Proxy URL for the request",
				Required:    false,
			},
			"response_extract": {
				Name:        "response_extract",
				Type:        integration.FieldTypeString,
				Description: "Path expression to extract from response (e.g., data.items[0].id)",
				Required:    false,
			},
			"signature_secret": {
				Name:        "signature_secret",
				Type:        integration.FieldTypeSecret,
				Description: "Secret for webhook signature",
				Required:    false,
				Sensitive:   true,
			},
			"signature_header": {
				Name:        "signature_header",
				Type:        integration.FieldTypeString,
				Description: "Header name for webhook signature",
				Required:    false,
			},
			"signature_type": {
				Name:        "signature_type",
				Type:        integration.FieldTypeString,
				Description: "Signature algorithm (hmac_sha256, hmac_sha1)",
				Required:    false,
				Options:     []string{"hmac_sha256", "hmac_sha1"},
			},
			"success_codes": {
				Name:        "success_codes",
				Type:        integration.FieldTypeArray,
				Description: "HTTP status codes considered successful",
				Required:    false,
			},
		},
		OutputSpec: map[string]integration.FieldSpec{
			"status_code": {
				Name:        "status_code",
				Type:        integration.FieldTypeInteger,
				Description: "HTTP response status code",
			},
			"headers": {
				Name:        "headers",
				Type:        integration.FieldTypeObject,
				Description: "Response headers",
			},
			"body": {
				Name:        "body",
				Type:        integration.FieldTypeObject,
				Description: "Response body (parsed)",
			},
			"extracted": {
				Name:        "extracted",
				Type:        integration.FieldTypeObject,
				Description: "Extracted data from response",
			},
		},
	}
}
