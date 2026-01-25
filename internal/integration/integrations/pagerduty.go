package integrations

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/gorax/gorax/internal/integration"
	inthttp "github.com/gorax/gorax/internal/integration/http"
)

const (
	pagerdutyAPIBaseURL   = "https://api.pagerduty.com"
	pagerdutyEventsAPIURL = "https://events.pagerduty.com/v2/enqueue"
	pagerdutyIntName      = "pagerduty"
)

// PagerDutyIntegration provides PagerDuty API integration capabilities.
type PagerDutyIntegration struct {
	*integration.BaseIntegration
	client       *inthttp.Client
	eventsClient *inthttp.Client
	logger       *slog.Logger
}

// PagerDutyAction represents the available PagerDuty actions.
type PagerDutyAction string

const (
	// PagerDutyActionTriggerIncident triggers a new incident via Events API.
	PagerDutyActionTriggerIncident PagerDutyAction = "trigger_incident"
	// PagerDutyActionAcknowledgeIncident acknowledges an incident via Events API.
	PagerDutyActionAcknowledgeIncident PagerDutyAction = "acknowledge_incident"
	// PagerDutyActionResolveIncident resolves an incident via Events API.
	PagerDutyActionResolveIncident PagerDutyAction = "resolve_incident"
	// PagerDutyActionCreateIncident creates an incident via REST API.
	PagerDutyActionCreateIncident PagerDutyAction = "create_incident"
	// PagerDutyActionUpdateIncident updates an incident via REST API.
	PagerDutyActionUpdateIncident PagerDutyAction = "update_incident"
	// PagerDutyActionGetIncident gets incident details.
	PagerDutyActionGetIncident PagerDutyAction = "get_incident"
	// PagerDutyActionListIncidents lists incidents with filters.
	PagerDutyActionListIncidents PagerDutyAction = "list_incidents"
	// PagerDutyActionAddNote adds a note to an incident.
	PagerDutyActionAddNote PagerDutyAction = "add_note"
	// PagerDutyActionGetService gets service details.
	PagerDutyActionGetService PagerDutyAction = "get_service"
	// PagerDutyActionListServices lists all services.
	PagerDutyActionListServices PagerDutyAction = "list_services"
)

// NewPagerDutyIntegration creates a new PagerDuty integration.
func NewPagerDutyIntegration(logger *slog.Logger) *PagerDutyIntegration {
	if logger == nil {
		logger = slog.Default()
	}

	base := integration.NewBaseIntegration(pagerdutyIntName, integration.TypeAPI)
	base.SetMetadata(&integration.Metadata{
		Name:        pagerdutyIntName,
		DisplayName: "PagerDuty",
		Description: "Manage incidents, services, and on-call schedules in PagerDuty",
		Version:     "1.0.0",
		Category:    "incident_management",
		Tags:        []string{"pagerduty", "incidents", "on-call", "alerting"},
		Author:      "Gorax",
	})
	base.SetSchema(buildPagerDutySchema())

	// REST API client
	client := inthttp.NewClient(
		inthttp.WithBaseURL(pagerdutyAPIBaseURL),
		inthttp.WithTimeout(30*time.Second),
		inthttp.WithLogger(logger),
		inthttp.WithRetryConfig(buildPagerDutyRetryConfig()),
		inthttp.WithHeader("Content-Type", "application/json"),
		inthttp.WithHeader("Accept", "application/vnd.pagerduty+json;version=2"),
	)

	// Events API client
	eventsClient := inthttp.NewClient(
		inthttp.WithTimeout(30*time.Second),
		inthttp.WithLogger(logger),
		inthttp.WithRetryConfig(buildPagerDutyRetryConfig()),
		inthttp.WithHeader("Content-Type", "application/json"),
	)

	return &PagerDutyIntegration{
		BaseIntegration: base,
		client:          client,
		eventsClient:    eventsClient,
		logger:          logger,
	}
}

// Execute performs a PagerDuty API action.
func (p *PagerDutyIntegration) Execute(ctx context.Context, config *integration.Config, params integration.JSONMap) (*integration.Result, error) {
	start := time.Now()

	action, ok := params.GetString("action")
	if !ok || action == "" {
		err := integration.NewValidationError("action", "action is required", nil)
		return integration.NewErrorResult(err, "VALIDATION_ERROR", time.Since(start).Milliseconds()), err
	}

	var result *integration.Result
	var err error

	switch PagerDutyAction(action) {
	case PagerDutyActionTriggerIncident:
		result, err = p.triggerIncident(ctx, config, params, start)
	case PagerDutyActionAcknowledgeIncident:
		result, err = p.acknowledgeIncident(ctx, config, params, start)
	case PagerDutyActionResolveIncident:
		result, err = p.resolveIncident(ctx, config, params, start)
	case PagerDutyActionCreateIncident:
		result, err = p.createIncident(ctx, config, params, start)
	case PagerDutyActionUpdateIncident:
		result, err = p.updateIncident(ctx, config, params, start)
	case PagerDutyActionGetIncident:
		result, err = p.getIncident(ctx, config, params, start)
	case PagerDutyActionListIncidents:
		result, err = p.listIncidents(ctx, config, params, start)
	case PagerDutyActionAddNote:
		result, err = p.addNote(ctx, config, params, start)
	case PagerDutyActionGetService:
		result, err = p.getService(ctx, config, params, start)
	case PagerDutyActionListServices:
		result, err = p.listServices(ctx, config, params, start)
	default:
		err = integration.NewValidationError("action", "unsupported action", action)
		result = integration.NewErrorResult(err, "INVALID_ACTION", time.Since(start).Milliseconds())
	}

	if err != nil {
		p.logger.Error("pagerduty action failed",
			"action", action,
			"error", err,
			"duration_ms", time.Since(start).Milliseconds(),
		)
	} else {
		p.logger.Info("pagerduty action completed",
			"action", action,
			"duration_ms", result.Duration,
		)
	}

	return result, err
}

// Validate validates the integration configuration.
func (p *PagerDutyIntegration) Validate(config *integration.Config) error {
	if err := p.BaseIntegration.ValidateConfig(config); err != nil {
		return err
	}

	if config.Credentials == nil {
		return integration.NewValidationError("credentials", "credentials are required", nil)
	}

	// Check for either API token or routing key
	hasAPIToken := false
	hasRoutingKey := false

	if config.Credentials.Data != nil {
		if token, ok := config.Credentials.Data.GetString("api_token"); ok && token != "" {
			hasAPIToken = true
		}
		if key, ok := config.Credentials.Data.GetString("routing_key"); ok && key != "" {
			hasRoutingKey = true
		}
	}

	if !hasAPIToken && !hasRoutingKey {
		return integration.NewValidationError("credentials", "api_token or routing_key is required", nil)
	}

	return nil
}

// getAPIToken extracts the API token from credentials.
func (p *PagerDutyIntegration) getAPIToken(config *integration.Config) (string, error) {
	if config.Credentials == nil || config.Credentials.Data == nil {
		return "", integration.NewValidationError("credentials", "credentials are required", nil)
	}

	token, ok := config.Credentials.Data.GetString("api_token")
	if !ok || token == "" {
		return "", integration.NewValidationError("api_token", "API token is required for this action", nil)
	}

	return token, nil
}

// getRoutingKey extracts the routing key from credentials or params.
func (p *PagerDutyIntegration) getRoutingKey(config *integration.Config, params integration.JSONMap) (string, error) {
	// First check params
	if key, ok := params.GetString("routing_key"); ok && key != "" {
		return key, nil
	}

	// Then check credentials
	if config.Credentials != nil && config.Credentials.Data != nil {
		if key, ok := config.Credentials.Data.GetString("routing_key"); ok && key != "" {
			return key, nil
		}
	}

	return "", integration.NewValidationError("routing_key", "routing key is required for Events API actions", nil)
}

// triggerIncident triggers a new incident via Events API v2.
func (p *PagerDutyIntegration) triggerIncident(ctx context.Context, config *integration.Config, params integration.JSONMap, start time.Time) (*integration.Result, error) {
	routingKey, err := p.getRoutingKey(config, params)
	if err != nil {
		return integration.NewErrorResult(err, "AUTH_ERROR", time.Since(start).Milliseconds()), err
	}

	summary, ok := params.GetString("summary")
	if !ok || summary == "" {
		err := integration.NewValidationError("summary", "summary is required", nil)
		return integration.NewErrorResult(err, "VALIDATION_ERROR", time.Since(start).Milliseconds()), err
	}

	severity, _ := params.GetString("severity")
	if severity == "" {
		severity = "error"
	}

	source, _ := params.GetString("source")
	if source == "" {
		source = "gorax"
	}

	// Generate dedup key if not provided
	dedupKey, _ := params.GetString("dedup_key")
	if dedupKey == "" {
		// Create deterministic dedup key from summary and source
		hash := sha256.Sum256([]byte(summary + source))
		dedupKey = hex.EncodeToString(hash[:16])
	}

	payload := map[string]any{
		"routing_key":  routingKey,
		"event_action": "trigger",
		"dedup_key":    dedupKey,
		"payload": map[string]any{
			"summary":  summary,
			"severity": severity,
			"source":   source,
		},
	}

	// Add optional fields
	payloadData := payload["payload"].(map[string]any)

	if component, ok := params.GetString("component"); ok {
		payloadData["component"] = component
	}
	if group, ok := params.GetString("group"); ok {
		payloadData["group"] = group
	}
	if class, ok := params.GetString("class"); ok {
		payloadData["class"] = class
	}
	if customDetails, ok := params.Get("custom_details"); ok {
		payloadData["custom_details"] = customDetails
	}

	// Add images if provided
	if images, ok := params.Get("images"); ok {
		payload["images"] = images
	}

	// Add links if provided
	if links, ok := params.Get("links"); ok {
		payload["links"] = links
	}

	return p.executeEventsAPI(ctx, payload, start)
}

// acknowledgeIncident acknowledges an incident via Events API v2.
func (p *PagerDutyIntegration) acknowledgeIncident(ctx context.Context, config *integration.Config, params integration.JSONMap, start time.Time) (*integration.Result, error) {
	routingKey, err := p.getRoutingKey(config, params)
	if err != nil {
		return integration.NewErrorResult(err, "AUTH_ERROR", time.Since(start).Milliseconds()), err
	}

	dedupKey, ok := params.GetString("dedup_key")
	if !ok || dedupKey == "" {
		err := integration.NewValidationError("dedup_key", "dedup_key is required to acknowledge incident", nil)
		return integration.NewErrorResult(err, "VALIDATION_ERROR", time.Since(start).Milliseconds()), err
	}

	payload := map[string]any{
		"routing_key":  routingKey,
		"event_action": "acknowledge",
		"dedup_key":    dedupKey,
	}

	return p.executeEventsAPI(ctx, payload, start)
}

// resolveIncident resolves an incident via Events API v2.
func (p *PagerDutyIntegration) resolveIncident(ctx context.Context, config *integration.Config, params integration.JSONMap, start time.Time) (*integration.Result, error) {
	routingKey, err := p.getRoutingKey(config, params)
	if err != nil {
		return integration.NewErrorResult(err, "AUTH_ERROR", time.Since(start).Milliseconds()), err
	}

	dedupKey, ok := params.GetString("dedup_key")
	if !ok || dedupKey == "" {
		err := integration.NewValidationError("dedup_key", "dedup_key is required to resolve incident", nil)
		return integration.NewErrorResult(err, "VALIDATION_ERROR", time.Since(start).Milliseconds()), err
	}

	payload := map[string]any{
		"routing_key":  routingKey,
		"event_action": "resolve",
		"dedup_key":    dedupKey,
	}

	return p.executeEventsAPI(ctx, payload, start)
}

// createIncident creates an incident via REST API.
func (p *PagerDutyIntegration) createIncident(ctx context.Context, config *integration.Config, params integration.JSONMap, start time.Time) (*integration.Result, error) {
	token, err := p.getAPIToken(config)
	if err != nil {
		return integration.NewErrorResult(err, "AUTH_ERROR", time.Since(start).Milliseconds()), err
	}

	serviceID, ok := params.GetString("service_id")
	if !ok || serviceID == "" {
		err := integration.NewValidationError("service_id", "service_id is required", nil)
		return integration.NewErrorResult(err, "VALIDATION_ERROR", time.Since(start).Milliseconds()), err
	}

	title, ok := params.GetString("title")
	if !ok || title == "" {
		err := integration.NewValidationError("title", "incident title is required", nil)
		return integration.NewErrorResult(err, "VALIDATION_ERROR", time.Since(start).Milliseconds()), err
	}

	incident := map[string]any{
		"type":  "incident",
		"title": title,
		"service": map[string]string{
			"id":   serviceID,
			"type": "service_reference",
		},
	}

	// Add optional fields
	if urgency, ok := params.GetString("urgency"); ok {
		incident["urgency"] = urgency
	}

	if body, ok := params.GetString("body"); ok {
		incident["body"] = map[string]string{
			"type":    "incident_body",
			"details": body,
		}
	}

	if escalationPolicyID, ok := params.GetString("escalation_policy_id"); ok {
		incident["escalation_policy"] = map[string]string{
			"id":   escalationPolicyID,
			"type": "escalation_policy_reference",
		}
	}

	if assignees, ok := params.Get("assignee_ids"); ok {
		if ids, ok := assignees.([]any); ok {
			assignments := make([]map[string]any, 0, len(ids))
			for _, id := range ids {
				if userID, ok := id.(string); ok {
					assignments = append(assignments, map[string]any{
						"assignee": map[string]string{
							"id":   userID,
							"type": "user_reference",
						},
					})
				}
			}
			incident["assignments"] = assignments
		}
	}

	payload := map[string]any{
		"incident": incident,
	}

	// Get from email for request header
	fromEmail, _ := params.GetString("from_email")
	if fromEmail == "" {
		if config.Credentials != nil && config.Credentials.Data != nil {
			fromEmail, _ = config.Credentials.Data.GetString("from_email")
		}
	}

	return p.executeRESTAPI(ctx, http.MethodPost, token, fromEmail, "/incidents", payload, nil, start)
}

// updateIncident updates an incident via REST API.
func (p *PagerDutyIntegration) updateIncident(ctx context.Context, config *integration.Config, params integration.JSONMap, start time.Time) (*integration.Result, error) {
	token, err := p.getAPIToken(config)
	if err != nil {
		return integration.NewErrorResult(err, "AUTH_ERROR", time.Since(start).Milliseconds()), err
	}

	incidentID, ok := params.GetString("incident_id")
	if !ok || incidentID == "" {
		err := integration.NewValidationError("incident_id", "incident_id is required", nil)
		return integration.NewErrorResult(err, "VALIDATION_ERROR", time.Since(start).Milliseconds()), err
	}

	incident := map[string]any{
		"id":   incidentID,
		"type": "incident",
	}

	if title, ok := params.GetString("title"); ok {
		incident["title"] = title
	}

	if status, ok := params.GetString("status"); ok {
		incident["status"] = status
	}

	if urgency, ok := params.GetString("urgency"); ok {
		incident["urgency"] = urgency
	}

	if resolution, ok := params.GetString("resolution"); ok {
		incident["resolution"] = resolution
	}

	payload := map[string]any{
		"incident": incident,
	}

	fromEmail, _ := params.GetString("from_email")
	if fromEmail == "" {
		if config.Credentials != nil && config.Credentials.Data != nil {
			fromEmail, _ = config.Credentials.Data.GetString("from_email")
		}
	}

	path := fmt.Sprintf("/incidents/%s", incidentID)
	return p.executeRESTAPI(ctx, http.MethodPut, token, fromEmail, path, payload, nil, start)
}

// getIncident gets incident details.
func (p *PagerDutyIntegration) getIncident(ctx context.Context, config *integration.Config, params integration.JSONMap, start time.Time) (*integration.Result, error) {
	token, err := p.getAPIToken(config)
	if err != nil {
		return integration.NewErrorResult(err, "AUTH_ERROR", time.Since(start).Milliseconds()), err
	}

	incidentID, ok := params.GetString("incident_id")
	if !ok || incidentID == "" {
		err := integration.NewValidationError("incident_id", "incident_id is required", nil)
		return integration.NewErrorResult(err, "VALIDATION_ERROR", time.Since(start).Milliseconds()), err
	}

	path := fmt.Sprintf("/incidents/%s", incidentID)
	return p.executeRESTAPI(ctx, http.MethodGet, token, "", path, nil, nil, start)
}

// listIncidents lists incidents with filters.
func (p *PagerDutyIntegration) listIncidents(ctx context.Context, config *integration.Config, params integration.JSONMap, start time.Time) (*integration.Result, error) {
	token, err := p.getAPIToken(config)
	if err != nil {
		return integration.NewErrorResult(err, "AUTH_ERROR", time.Since(start).Milliseconds()), err
	}

	queryParams := make(map[string]string)

	if statuses, ok := params.Get("statuses"); ok {
		if arr, ok := statuses.([]any); ok {
			for i, s := range arr {
				if str, ok := s.(string); ok {
					queryParams[fmt.Sprintf("statuses[%d]", i)] = str
				}
			}
		}
	}

	if serviceIDs, ok := params.Get("service_ids"); ok {
		if arr, ok := serviceIDs.([]any); ok {
			for i, s := range arr {
				if str, ok := s.(string); ok {
					queryParams[fmt.Sprintf("service_ids[%d]", i)] = str
				}
			}
		}
	}

	if since, ok := params.GetString("since"); ok {
		queryParams["since"] = since
	}

	if until, ok := params.GetString("until"); ok {
		queryParams["until"] = until
	}

	if urgencies, ok := params.Get("urgencies"); ok {
		if arr, ok := urgencies.([]any); ok {
			for i, u := range arr {
				if str, ok := u.(string); ok {
					queryParams[fmt.Sprintf("urgencies[%d]", i)] = str
				}
			}
		}
	}

	if limit, ok := params.GetInt("limit"); ok {
		queryParams["limit"] = strconv.Itoa(limit)
	}

	if offset, ok := params.GetInt("offset"); ok {
		queryParams["offset"] = strconv.Itoa(offset)
	}

	if sortBy, ok := params.GetString("sort_by"); ok {
		queryParams["sort_by"] = sortBy
	}

	return p.executeRESTAPI(ctx, http.MethodGet, token, "", "/incidents", nil, queryParams, start)
}

// addNote adds a note to an incident.
func (p *PagerDutyIntegration) addNote(ctx context.Context, config *integration.Config, params integration.JSONMap, start time.Time) (*integration.Result, error) {
	token, err := p.getAPIToken(config)
	if err != nil {
		return integration.NewErrorResult(err, "AUTH_ERROR", time.Since(start).Milliseconds()), err
	}

	incidentID, ok := params.GetString("incident_id")
	if !ok || incidentID == "" {
		err := integration.NewValidationError("incident_id", "incident_id is required", nil)
		return integration.NewErrorResult(err, "VALIDATION_ERROR", time.Since(start).Milliseconds()), err
	}

	content, ok := params.GetString("content")
	if !ok || content == "" {
		err := integration.NewValidationError("content", "note content is required", nil)
		return integration.NewErrorResult(err, "VALIDATION_ERROR", time.Since(start).Milliseconds()), err
	}

	payload := map[string]any{
		"note": map[string]string{
			"content": content,
		},
	}

	fromEmail, _ := params.GetString("from_email")
	if fromEmail == "" {
		if config.Credentials != nil && config.Credentials.Data != nil {
			fromEmail, _ = config.Credentials.Data.GetString("from_email")
		}
	}

	path := fmt.Sprintf("/incidents/%s/notes", incidentID)
	return p.executeRESTAPI(ctx, http.MethodPost, token, fromEmail, path, payload, nil, start)
}

// getService gets service details.
func (p *PagerDutyIntegration) getService(ctx context.Context, config *integration.Config, params integration.JSONMap, start time.Time) (*integration.Result, error) {
	token, err := p.getAPIToken(config)
	if err != nil {
		return integration.NewErrorResult(err, "AUTH_ERROR", time.Since(start).Milliseconds()), err
	}

	serviceID, ok := params.GetString("service_id")
	if !ok || serviceID == "" {
		err := integration.NewValidationError("service_id", "service_id is required", nil)
		return integration.NewErrorResult(err, "VALIDATION_ERROR", time.Since(start).Milliseconds()), err
	}

	path := fmt.Sprintf("/services/%s", serviceID)
	return p.executeRESTAPI(ctx, http.MethodGet, token, "", path, nil, nil, start)
}

// listServices lists all services.
func (p *PagerDutyIntegration) listServices(ctx context.Context, config *integration.Config, params integration.JSONMap, start time.Time) (*integration.Result, error) {
	token, err := p.getAPIToken(config)
	if err != nil {
		return integration.NewErrorResult(err, "AUTH_ERROR", time.Since(start).Milliseconds()), err
	}

	queryParams := make(map[string]string)

	if query, ok := params.GetString("query"); ok {
		queryParams["query"] = query
	}

	if limit, ok := params.GetInt("limit"); ok {
		queryParams["limit"] = strconv.Itoa(limit)
	}

	if offset, ok := params.GetInt("offset"); ok {
		queryParams["offset"] = strconv.Itoa(offset)
	}

	return p.executeRESTAPI(ctx, http.MethodGet, token, "", "/services", nil, queryParams, start)
}

// executeEventsAPI executes a PagerDuty Events API request.
func (p *PagerDutyIntegration) executeEventsAPI(ctx context.Context, payload map[string]any, start time.Time) (*integration.Result, error) {
	resp, err := p.eventsClient.Post(ctx, pagerdutyEventsAPIURL, payload)
	if err != nil {
		return integration.NewErrorResult(err, "API_ERROR", time.Since(start).Milliseconds()), err
	}

	var data map[string]any
	if err := json.Unmarshal(resp.Body, &data); err != nil {
		data = map[string]any{"response": string(resp.Body)}
	}

	return &integration.Result{
		Success:    resp.IsSuccess(),
		Data:       data,
		StatusCode: resp.StatusCode,
		Duration:   time.Since(start).Milliseconds(),
		ExecutedAt: time.Now().UTC(),
	}, nil
}

// executeRESTAPI executes a PagerDuty REST API request.
func (p *PagerDutyIntegration) executeRESTAPI(ctx context.Context, method, token, fromEmail, path string, payload map[string]any, queryParams map[string]string, start time.Time) (*integration.Result, error) {
	opts := []inthttp.RequestOption{
		inthttp.WithRequestHeader("Authorization", "Token token="+token),
	}

	if fromEmail != "" {
		opts = append(opts, inthttp.WithRequestHeader("From", fromEmail))
	}

	if queryParams != nil {
		opts = append(opts, inthttp.WithQueryParams(queryParams))
	}

	var resp *inthttp.Response
	var err error

	switch method {
	case http.MethodGet:
		resp, err = p.client.Get(ctx, path, opts...)
	case http.MethodPost:
		resp, err = p.client.Post(ctx, path, payload, opts...)
	case http.MethodPut:
		resp, err = p.client.Put(ctx, path, payload, opts...)
	case http.MethodDelete:
		resp, err = p.client.Delete(ctx, path, opts...)
	default:
		err = fmt.Errorf("unsupported HTTP method: %s", method)
	}

	if err != nil {
		return integration.NewErrorResult(err, "API_ERROR", time.Since(start).Milliseconds()), err
	}

	var data any
	if len(resp.Body) > 0 {
		if jsonErr := json.Unmarshal(resp.Body, &data); jsonErr != nil {
			data = string(resp.Body)
		}
	}

	return &integration.Result{
		Success:    resp.IsSuccess(),
		Data:       data,
		StatusCode: resp.StatusCode,
		Duration:   time.Since(start).Milliseconds(),
		ExecutedAt: time.Now().UTC(),
	}, nil
}

// buildPagerDutySchema builds the schema for the PagerDuty integration.
func buildPagerDutySchema() *integration.Schema {
	return &integration.Schema{
		ConfigSpec: map[string]integration.FieldSpec{
			"api_token": {
				Name:        "api_token",
				Type:        integration.FieldTypeSecret,
				Description: "PagerDuty API token for REST API",
				Required:    false,
				Sensitive:   true,
			},
			"routing_key": {
				Name:        "routing_key",
				Type:        integration.FieldTypeSecret,
				Description: "Integration key for Events API v2",
				Required:    false,
				Sensitive:   true,
			},
			"from_email": {
				Name:        "from_email",
				Type:        integration.FieldTypeString,
				Description: "Email address for REST API requests",
				Required:    false,
			},
		},
		InputSpec: map[string]integration.FieldSpec{
			"action": {
				Name:        "action",
				Type:        integration.FieldTypeString,
				Description: "Action to perform",
				Required:    true,
				Options: []string{
					string(PagerDutyActionTriggerIncident),
					string(PagerDutyActionAcknowledgeIncident),
					string(PagerDutyActionResolveIncident),
					string(PagerDutyActionCreateIncident),
					string(PagerDutyActionUpdateIncident),
					string(PagerDutyActionGetIncident),
					string(PagerDutyActionListIncidents),
					string(PagerDutyActionAddNote),
					string(PagerDutyActionGetService),
					string(PagerDutyActionListServices),
				},
			},
			"incident_id": {
				Name:        "incident_id",
				Type:        integration.FieldTypeString,
				Description: "Incident ID",
				Required:    false,
			},
			"service_id": {
				Name:        "service_id",
				Type:        integration.FieldTypeString,
				Description: "Service ID",
				Required:    false,
			},
			"title": {
				Name:        "title",
				Type:        integration.FieldTypeString,
				Description: "Incident title",
				Required:    false,
			},
			"summary": {
				Name:        "summary",
				Type:        integration.FieldTypeString,
				Description: "Event summary (for Events API)",
				Required:    false,
			},
			"severity": {
				Name:        "severity",
				Type:        integration.FieldTypeString,
				Description: "Severity level (critical, error, warning, info)",
				Required:    false,
				Options:     []string{"critical", "error", "warning", "info"},
			},
			"source": {
				Name:        "source",
				Type:        integration.FieldTypeString,
				Description: "Event source",
				Required:    false,
			},
			"dedup_key": {
				Name:        "dedup_key",
				Type:        integration.FieldTypeString,
				Description: "Deduplication key for Events API",
				Required:    false,
			},
			"urgency": {
				Name:        "urgency",
				Type:        integration.FieldTypeString,
				Description: "Urgency level (high, low)",
				Required:    false,
				Options:     []string{"high", "low"},
			},
			"status": {
				Name:        "status",
				Type:        integration.FieldTypeString,
				Description: "Incident status (triggered, acknowledged, resolved)",
				Required:    false,
				Options:     []string{"triggered", "acknowledged", "resolved"},
			},
			"body": {
				Name:        "body",
				Type:        integration.FieldTypeString,
				Description: "Incident body/details",
				Required:    false,
			},
			"content": {
				Name:        "content",
				Type:        integration.FieldTypeString,
				Description: "Note content",
				Required:    false,
			},
			"custom_details": {
				Name:        "custom_details",
				Type:        integration.FieldTypeObject,
				Description: "Custom details for the event",
				Required:    false,
			},
			"from_email": {
				Name:        "from_email",
				Type:        integration.FieldTypeString,
				Description: "Email address for REST API requests",
				Required:    false,
			},
		},
		OutputSpec: map[string]integration.FieldSpec{
			"incident": {
				Name:        "incident",
				Type:        integration.FieldTypeObject,
				Description: "Incident details",
			},
			"incidents": {
				Name:        "incidents",
				Type:        integration.FieldTypeArray,
				Description: "List of incidents",
			},
			"service": {
				Name:        "service",
				Type:        integration.FieldTypeObject,
				Description: "Service details",
			},
			"services": {
				Name:        "services",
				Type:        integration.FieldTypeArray,
				Description: "List of services",
			},
			"dedup_key": {
				Name:        "dedup_key",
				Type:        integration.FieldTypeString,
				Description: "Deduplication key (Events API response)",
			},
			"status": {
				Name:        "status",
				Type:        integration.FieldTypeString,
				Description: "Response status",
			},
		},
	}
}

// buildPagerDutyRetryConfig builds retry configuration for PagerDuty API.
func buildPagerDutyRetryConfig() *inthttp.RetryConfig {
	return &inthttp.RetryConfig{
		MaxRetries:   3,
		BaseDelay:    1 * time.Second,
		MaxDelay:     30 * time.Second,
		Multiplier:   2.0,
		JitterFactor: 0.1,
		ShouldRetry: func(err error, resp *inthttp.Response) bool {
			if err != nil {
				return inthttp.IsRetryableError(err)
			}
			if resp == nil {
				return false
			}
			return resp.StatusCode == 429 || resp.StatusCode >= 500
		},
	}
}
