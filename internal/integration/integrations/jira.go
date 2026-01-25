package integrations

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log/slog"
	"maps"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorax/gorax/internal/integration"
	inthttp "github.com/gorax/gorax/internal/integration/http"
)

const jiraIntName = "jira"

// JiraIntegration provides Jira API integration capabilities.
type JiraIntegration struct {
	*integration.BaseIntegration
	logger *slog.Logger
}

// JiraAction represents the available Jira actions.
type JiraAction string

const (
	// JiraActionCreateIssue creates a new issue.
	JiraActionCreateIssue JiraAction = "create_issue"
	// JiraActionUpdateIssue updates an existing issue.
	JiraActionUpdateIssue JiraAction = "update_issue"
	// JiraActionGetIssue gets issue details.
	JiraActionGetIssue JiraAction = "get_issue"
	// JiraActionSearchIssues searches issues using JQL.
	JiraActionSearchIssues JiraAction = "search_issues"
	// JiraActionTransitionIssue transitions an issue to a new status.
	JiraActionTransitionIssue JiraAction = "transition_issue"
	// JiraActionAddComment adds a comment to an issue.
	JiraActionAddComment JiraAction = "add_comment"
	// JiraActionGetTransitions gets available transitions for an issue.
	JiraActionGetTransitions JiraAction = "get_transitions"
	// JiraActionAssignIssue assigns an issue to a user.
	JiraActionAssignIssue JiraAction = "assign_issue"
	// JiraActionGetProject gets project details.
	JiraActionGetProject JiraAction = "get_project"
	// JiraActionListProjects lists all projects.
	JiraActionListProjects JiraAction = "list_projects"
)

// NewJiraIntegration creates a new Jira integration.
func NewJiraIntegration(logger *slog.Logger) *JiraIntegration {
	if logger == nil {
		logger = slog.Default()
	}

	base := integration.NewBaseIntegration(jiraIntName, integration.TypeAPI)
	base.SetMetadata(&integration.Metadata{
		Name:        jiraIntName,
		DisplayName: "Jira",
		Description: "Manage issues, projects, and workflows in Jira",
		Version:     "1.0.0",
		Category:    "project_management",
		Tags:        []string{"jira", "issues", "project-management", "atlassian"},
		Author:      "Gorax",
	})
	base.SetSchema(buildJiraSchema())

	return &JiraIntegration{
		BaseIntegration: base,
		logger:          logger,
	}
}

// Execute performs a Jira API action.
func (j *JiraIntegration) Execute(ctx context.Context, config *integration.Config, params integration.JSONMap) (*integration.Result, error) {
	start := time.Now()

	action, ok := params.GetString("action")
	if !ok || action == "" {
		err := integration.NewValidationError("action", "action is required", nil)
		return integration.NewErrorResult(err, "VALIDATION_ERROR", time.Since(start).Milliseconds()), err
	}

	// Create client with base URL from config
	client, err := j.createClient(config)
	if err != nil {
		return integration.NewErrorResult(err, "CONFIG_ERROR", time.Since(start).Milliseconds()), err
	}

	// Get auth header
	authHeader, err := j.getAuthHeader(config)
	if err != nil {
		return integration.NewErrorResult(err, "AUTH_ERROR", time.Since(start).Milliseconds()), err
	}

	var result *integration.Result
	switch JiraAction(action) {
	case JiraActionCreateIssue:
		result, err = j.createIssue(ctx, client, authHeader, params, start)
	case JiraActionUpdateIssue:
		result, err = j.updateIssue(ctx, client, authHeader, params, start)
	case JiraActionGetIssue:
		result, err = j.getIssue(ctx, client, authHeader, params, start)
	case JiraActionSearchIssues:
		result, err = j.searchIssues(ctx, client, authHeader, params, start)
	case JiraActionTransitionIssue:
		result, err = j.transitionIssue(ctx, client, authHeader, params, start)
	case JiraActionAddComment:
		result, err = j.addComment(ctx, client, authHeader, params, start)
	case JiraActionGetTransitions:
		result, err = j.getTransitions(ctx, client, authHeader, params, start)
	case JiraActionAssignIssue:
		result, err = j.assignIssue(ctx, client, authHeader, params, start)
	case JiraActionGetProject:
		result, err = j.getProject(ctx, client, authHeader, params, start)
	case JiraActionListProjects:
		result, err = j.listProjects(ctx, client, authHeader, params, start)
	default:
		err = integration.NewValidationError("action", "unsupported action", action)
		result = integration.NewErrorResult(err, "INVALID_ACTION", time.Since(start).Milliseconds())
	}

	if err != nil {
		j.logger.Error("jira action failed",
			"action", action,
			"error", err,
			"duration_ms", time.Since(start).Milliseconds(),
		)
	} else {
		j.logger.Info("jira action completed",
			"action", action,
			"duration_ms", result.Duration,
		)
	}

	return result, err
}

// Validate validates the integration configuration.
func (j *JiraIntegration) Validate(config *integration.Config) error {
	if err := j.BaseIntegration.ValidateConfig(config); err != nil {
		return err
	}

	// Validate base URL
	baseURL, ok := config.Settings.GetString("base_url")
	if !ok || baseURL == "" {
		return integration.NewValidationError("base_url", "Jira base URL is required", nil)
	}

	// Validate credentials
	if config.Credentials == nil {
		return integration.NewValidationError("credentials", "credentials are required", nil)
	}

	if _, err := j.getAuthHeader(config); err != nil {
		return err
	}

	return nil
}

// createClient creates an HTTP client for Jira API.
func (j *JiraIntegration) createClient(config *integration.Config) (*inthttp.Client, error) {
	baseURL, ok := config.Settings.GetString("base_url")
	if !ok || baseURL == "" {
		return nil, integration.NewValidationError("base_url", "Jira base URL is required", nil)
	}

	// Ensure base URL ends properly
	baseURL = strings.TrimSuffix(baseURL, "/")
	if !strings.HasSuffix(baseURL, "/rest/api") {
		baseURL = baseURL + "/rest/api/3"
	}

	return inthttp.NewClient(
		inthttp.WithBaseURL(baseURL),
		inthttp.WithTimeout(30*time.Second),
		inthttp.WithLogger(j.logger),
		inthttp.WithRetryConfig(buildJiraRetryConfig()),
		inthttp.WithHeader("Content-Type", "application/json"),
		inthttp.WithHeader("Accept", "application/json"),
	), nil
}

// getAuthHeader generates the authorization header from credentials.
func (j *JiraIntegration) getAuthHeader(config *integration.Config) (string, error) {
	if config.Credentials == nil || config.Credentials.Data == nil {
		return "", integration.NewValidationError("credentials", "credentials are required", nil)
	}

	// Try API token authentication (Cloud)
	if email, ok := config.Credentials.Data.GetString("email"); ok && email != "" {
		if apiToken, ok := config.Credentials.Data.GetString("api_token"); ok && apiToken != "" {
			auth := base64.StdEncoding.EncodeToString([]byte(email + ":" + apiToken))
			return "Basic " + auth, nil
		}
	}

	// Try username/password (Server/Data Center)
	if username, ok := config.Credentials.Data.GetString("username"); ok && username != "" {
		if password, ok := config.Credentials.Data.GetString("password"); ok && password != "" {
			auth := base64.StdEncoding.EncodeToString([]byte(username + ":" + password))
			return "Basic " + auth, nil
		}
	}

	// Try bearer token
	if token, ok := config.Credentials.Data.GetString("token"); ok && token != "" {
		return "Bearer " + token, nil
	}

	return "", integration.NewValidationError("credentials", "valid credentials (email/api_token, username/password, or token) are required", nil)
}

// createIssue creates a new Jira issue.
func (j *JiraIntegration) createIssue(ctx context.Context, client *inthttp.Client, authHeader string, params integration.JSONMap, start time.Time) (*integration.Result, error) {
	project, ok := params.GetString("project")
	if !ok || project == "" {
		err := integration.NewValidationError("project", "project key is required", nil)
		return integration.NewErrorResult(err, "VALIDATION_ERROR", time.Since(start).Milliseconds()), err
	}

	issueType, ok := params.GetString("issue_type")
	if !ok || issueType == "" {
		issueType = "Task" // Default issue type
	}

	summary, ok := params.GetString("summary")
	if !ok || summary == "" {
		err := integration.NewValidationError("summary", "issue summary is required", nil)
		return integration.NewErrorResult(err, "VALIDATION_ERROR", time.Since(start).Milliseconds()), err
	}

	fields := map[string]any{
		"project": map[string]string{
			"key": project,
		},
		"issuetype": map[string]string{
			"name": issueType,
		},
		"summary": summary,
	}

	// Add description if provided
	if description, ok := params.GetString("description"); ok && description != "" {
		// Jira Cloud uses Atlassian Document Format (ADF)
		fields["description"] = map[string]any{
			"type":    "doc",
			"version": 1,
			"content": []map[string]any{
				{
					"type": "paragraph",
					"content": []map[string]any{
						{
							"type": "text",
							"text": description,
						},
					},
				},
			},
		}
	}

	// Add priority if provided
	if priority, ok := params.GetString("priority"); ok && priority != "" {
		fields["priority"] = map[string]string{
			"name": priority,
		}
	}

	// Add assignee if provided
	if assignee, ok := params.GetString("assignee"); ok && assignee != "" {
		fields["assignee"] = map[string]string{
			"accountId": assignee,
		}
	}

	// Add labels if provided
	if labels, ok := params.Get("labels"); ok {
		fields["labels"] = labels
	}

	// Add components if provided
	if components, ok := params.Get("components"); ok {
		if compArr, ok := components.([]any); ok {
			comps := make([]map[string]string, 0, len(compArr))
			for _, c := range compArr {
				if name, ok := c.(string); ok {
					comps = append(comps, map[string]string{"name": name})
				}
			}
			fields["components"] = comps
		}
	}

	// Add custom fields
	if customFields, ok := params.Get("custom_fields"); ok {
		if cfMap, ok := customFields.(map[string]any); ok {
			maps.Copy(fields, cfMap)
		}
	}

	payload := map[string]any{
		"fields": fields,
	}

	return j.executeJiraAPI(ctx, client, http.MethodPost, authHeader, "/issue", payload, nil, start)
}

// updateIssue updates an existing Jira issue.
func (j *JiraIntegration) updateIssue(ctx context.Context, client *inthttp.Client, authHeader string, params integration.JSONMap, start time.Time) (*integration.Result, error) {
	issueKey, ok := params.GetString("issue_key")
	if !ok || issueKey == "" {
		err := integration.NewValidationError("issue_key", "issue key is required", nil)
		return integration.NewErrorResult(err, "VALIDATION_ERROR", time.Since(start).Milliseconds()), err
	}

	fields := make(map[string]any)

	if summary, ok := params.GetString("summary"); ok {
		fields["summary"] = summary
	}

	if description, ok := params.GetString("description"); ok {
		fields["description"] = map[string]any{
			"type":    "doc",
			"version": 1,
			"content": []map[string]any{
				{
					"type": "paragraph",
					"content": []map[string]any{
						{
							"type": "text",
							"text": description,
						},
					},
				},
			},
		}
	}

	if priority, ok := params.GetString("priority"); ok {
		fields["priority"] = map[string]string{
			"name": priority,
		}
	}

	if assignee, ok := params.GetString("assignee"); ok {
		fields["assignee"] = map[string]string{
			"accountId": assignee,
		}
	}

	if labels, ok := params.Get("labels"); ok {
		fields["labels"] = labels
	}

	// Add custom fields
	if customFields, ok := params.Get("custom_fields"); ok {
		if cfMap, ok := customFields.(map[string]any); ok {
			maps.Copy(fields, cfMap)
		}
	}

	payload := map[string]any{
		"fields": fields,
	}

	path := fmt.Sprintf("/issue/%s", issueKey)
	return j.executeJiraAPI(ctx, client, http.MethodPut, authHeader, path, payload, nil, start)
}

// getIssue gets issue details.
func (j *JiraIntegration) getIssue(ctx context.Context, client *inthttp.Client, authHeader string, params integration.JSONMap, start time.Time) (*integration.Result, error) {
	issueKey, ok := params.GetString("issue_key")
	if !ok || issueKey == "" {
		err := integration.NewValidationError("issue_key", "issue key is required", nil)
		return integration.NewErrorResult(err, "VALIDATION_ERROR", time.Since(start).Milliseconds()), err
	}

	queryParams := make(map[string]string)

	if fields, ok := params.GetString("fields"); ok {
		queryParams["fields"] = fields
	}

	if expand, ok := params.GetString("expand"); ok {
		queryParams["expand"] = expand
	}

	path := fmt.Sprintf("/issue/%s", issueKey)
	return j.executeJiraAPI(ctx, client, http.MethodGet, authHeader, path, nil, queryParams, start)
}

// searchIssues searches issues using JQL.
func (j *JiraIntegration) searchIssues(ctx context.Context, client *inthttp.Client, authHeader string, params integration.JSONMap, start time.Time) (*integration.Result, error) {
	jql, ok := params.GetString("jql")
	if !ok || jql == "" {
		err := integration.NewValidationError("jql", "JQL query is required", nil)
		return integration.NewErrorResult(err, "VALIDATION_ERROR", time.Since(start).Milliseconds()), err
	}

	payload := map[string]any{
		"jql": jql,
	}

	if fields, ok := params.Get("fields"); ok {
		payload["fields"] = fields
	} else {
		payload["fields"] = []string{"summary", "status", "assignee", "priority", "created", "updated"}
	}

	if startAt, ok := params.GetInt("start_at"); ok {
		payload["startAt"] = startAt
	}

	if maxResults, ok := params.GetInt("max_results"); ok {
		payload["maxResults"] = maxResults
	} else {
		payload["maxResults"] = 50
	}

	if expand, ok := params.Get("expand"); ok {
		payload["expand"] = expand
	}

	return j.executeJiraAPI(ctx, client, http.MethodPost, authHeader, "/search", payload, nil, start)
}

// transitionIssue transitions an issue to a new status.
func (j *JiraIntegration) transitionIssue(ctx context.Context, client *inthttp.Client, authHeader string, params integration.JSONMap, start time.Time) (*integration.Result, error) {
	issueKey, ok := params.GetString("issue_key")
	if !ok || issueKey == "" {
		err := integration.NewValidationError("issue_key", "issue key is required", nil)
		return integration.NewErrorResult(err, "VALIDATION_ERROR", time.Since(start).Milliseconds()), err
	}

	transitionID, ok := params.GetString("transition_id")
	if !ok || transitionID == "" {
		// Try to get by name
		transitionName, hasName := params.GetString("transition_name")
		if !hasName || transitionName == "" {
			err := integration.NewValidationError("transition_id", "transition_id or transition_name is required", nil)
			return integration.NewErrorResult(err, "VALIDATION_ERROR", time.Since(start).Milliseconds()), err
		}

		// Look up transition ID by name
		transitions, err := j.getTransitions(ctx, client, authHeader, params, start)
		if err != nil {
			return transitions, err
		}

		if data, ok := transitions.Data.(map[string]any); ok {
			if transArr, ok := data["transitions"].([]any); ok {
				for _, t := range transArr {
					if trans, ok := t.(map[string]any); ok {
						if name, ok := trans["name"].(string); ok {
							if strings.EqualFold(name, transitionName) {
								if id, ok := trans["id"].(string); ok {
									transitionID = id
									break
								}
							}
						}
					}
				}
			}
		}

		if transitionID == "" {
			err := fmt.Errorf("transition not found: %s", transitionName)
			return integration.NewErrorResult(err, "TRANSITION_NOT_FOUND", time.Since(start).Milliseconds()), err
		}
	}

	payload := map[string]any{
		"transition": map[string]string{
			"id": transitionID,
		},
	}

	// Add comment if provided
	if comment, ok := params.GetString("comment"); ok && comment != "" {
		payload["update"] = map[string]any{
			"comment": []map[string]any{
				{
					"add": map[string]any{
						"body": map[string]any{
							"type":    "doc",
							"version": 1,
							"content": []map[string]any{
								{
									"type": "paragraph",
									"content": []map[string]any{
										{
											"type": "text",
											"text": comment,
										},
									},
								},
							},
						},
					},
				},
			},
		}
	}

	// Add resolution if provided
	if resolution, ok := params.GetString("resolution"); ok && resolution != "" {
		if fields, ok := payload["fields"].(map[string]any); ok {
			fields["resolution"] = map[string]string{"name": resolution}
		} else {
			payload["fields"] = map[string]any{
				"resolution": map[string]string{"name": resolution},
			}
		}
	}

	path := fmt.Sprintf("/issue/%s/transitions", issueKey)
	return j.executeJiraAPI(ctx, client, http.MethodPost, authHeader, path, payload, nil, start)
}

// addComment adds a comment to an issue.
func (j *JiraIntegration) addComment(ctx context.Context, client *inthttp.Client, authHeader string, params integration.JSONMap, start time.Time) (*integration.Result, error) {
	issueKey, ok := params.GetString("issue_key")
	if !ok || issueKey == "" {
		err := integration.NewValidationError("issue_key", "issue key is required", nil)
		return integration.NewErrorResult(err, "VALIDATION_ERROR", time.Since(start).Milliseconds()), err
	}

	body, ok := params.GetString("body")
	if !ok || body == "" {
		err := integration.NewValidationError("body", "comment body is required", nil)
		return integration.NewErrorResult(err, "VALIDATION_ERROR", time.Since(start).Milliseconds()), err
	}

	payload := map[string]any{
		"body": map[string]any{
			"type":    "doc",
			"version": 1,
			"content": []map[string]any{
				{
					"type": "paragraph",
					"content": []map[string]any{
						{
							"type": "text",
							"text": body,
						},
					},
				},
			},
		},
	}

	path := fmt.Sprintf("/issue/%s/comment", issueKey)
	return j.executeJiraAPI(ctx, client, http.MethodPost, authHeader, path, payload, nil, start)
}

// getTransitions gets available transitions for an issue.
func (j *JiraIntegration) getTransitions(ctx context.Context, client *inthttp.Client, authHeader string, params integration.JSONMap, start time.Time) (*integration.Result, error) {
	issueKey, ok := params.GetString("issue_key")
	if !ok || issueKey == "" {
		err := integration.NewValidationError("issue_key", "issue key is required", nil)
		return integration.NewErrorResult(err, "VALIDATION_ERROR", time.Since(start).Milliseconds()), err
	}

	path := fmt.Sprintf("/issue/%s/transitions", issueKey)
	return j.executeJiraAPI(ctx, client, http.MethodGet, authHeader, path, nil, nil, start)
}

// assignIssue assigns an issue to a user.
func (j *JiraIntegration) assignIssue(ctx context.Context, client *inthttp.Client, authHeader string, params integration.JSONMap, start time.Time) (*integration.Result, error) {
	issueKey, ok := params.GetString("issue_key")
	if !ok || issueKey == "" {
		err := integration.NewValidationError("issue_key", "issue key is required", nil)
		return integration.NewErrorResult(err, "VALIDATION_ERROR", time.Since(start).Milliseconds()), err
	}

	payload := make(map[string]any)

	// Support account ID (Cloud) or name (Server)
	if accountID, ok := params.GetString("account_id"); ok && accountID != "" {
		payload["accountId"] = accountID
	} else if name, ok := params.GetString("name"); ok && name != "" {
		payload["name"] = name
	} else {
		// Unassign
		payload["accountId"] = nil
	}

	path := fmt.Sprintf("/issue/%s/assignee", issueKey)
	return j.executeJiraAPI(ctx, client, http.MethodPut, authHeader, path, payload, nil, start)
}

// getProject gets project details.
func (j *JiraIntegration) getProject(ctx context.Context, client *inthttp.Client, authHeader string, params integration.JSONMap, start time.Time) (*integration.Result, error) {
	projectKey, ok := params.GetString("project_key")
	if !ok || projectKey == "" {
		err := integration.NewValidationError("project_key", "project key is required", nil)
		return integration.NewErrorResult(err, "VALIDATION_ERROR", time.Since(start).Milliseconds()), err
	}

	queryParams := make(map[string]string)
	if expand, ok := params.GetString("expand"); ok {
		queryParams["expand"] = expand
	}

	path := fmt.Sprintf("/project/%s", projectKey)
	return j.executeJiraAPI(ctx, client, http.MethodGet, authHeader, path, nil, queryParams, start)
}

// listProjects lists all projects.
func (j *JiraIntegration) listProjects(ctx context.Context, client *inthttp.Client, authHeader string, params integration.JSONMap, start time.Time) (*integration.Result, error) {
	queryParams := make(map[string]string)

	if startAt, ok := params.GetInt("start_at"); ok {
		queryParams["startAt"] = strconv.Itoa(startAt)
	}
	if maxResults, ok := params.GetInt("max_results"); ok {
		queryParams["maxResults"] = strconv.Itoa(maxResults)
	}
	if expand, ok := params.GetString("expand"); ok {
		queryParams["expand"] = expand
	}

	return j.executeJiraAPI(ctx, client, http.MethodGet, authHeader, "/project", nil, queryParams, start)
}

// executeJiraAPI executes a Jira API request.
func (j *JiraIntegration) executeJiraAPI(ctx context.Context, client *inthttp.Client, method, authHeader, path string, payload map[string]any, queryParams map[string]string, start time.Time) (*integration.Result, error) {
	opts := []inthttp.RequestOption{
		inthttp.WithRequestHeader("Authorization", authHeader),
	}

	if queryParams != nil {
		opts = append(opts, inthttp.WithQueryParams(queryParams))
	}

	var resp *inthttp.Response
	var err error

	switch method {
	case http.MethodGet:
		resp, err = client.Get(ctx, path, opts...)
	case http.MethodPost:
		resp, err = client.Post(ctx, path, payload, opts...)
	case http.MethodPut:
		resp, err = client.Put(ctx, path, payload, opts...)
	case http.MethodDelete:
		resp, err = client.Delete(ctx, path, opts...)
	default:
		err = fmt.Errorf("unsupported HTTP method: %s", method)
	}

	if err != nil {
		// Check for Jira-specific error format
		var jiraErr map[string]any
		if resp != nil && len(resp.Body) > 0 {
			if jsonErr := json.Unmarshal(resp.Body, &jiraErr); jsonErr == nil {
				if errors, ok := jiraErr["errorMessages"].([]any); ok && len(errors) > 0 {
					err = fmt.Errorf("jira error: %v", errors)
				} else if errMap, ok := jiraErr["errors"].(map[string]any); ok {
					err = fmt.Errorf("jira validation errors: %v", errMap)
				}
			}
		}
		return integration.NewErrorResult(err, "API_ERROR", time.Since(start).Milliseconds()), err
	}

	// Parse response
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

// buildJiraSchema builds the schema for the Jira integration.
func buildJiraSchema() *integration.Schema {
	return &integration.Schema{
		ConfigSpec: map[string]integration.FieldSpec{
			"base_url": {
				Name:        "base_url",
				Type:        integration.FieldTypeString,
				Description: "Jira instance URL (e.g., https://your-domain.atlassian.net)",
				Required:    true,
			},
			"email": {
				Name:        "email",
				Type:        integration.FieldTypeString,
				Description: "Email address for API authentication (Cloud)",
				Required:    false,
			},
			"api_token": {
				Name:        "api_token",
				Type:        integration.FieldTypeSecret,
				Description: "API token for authentication (Cloud)",
				Required:    false,
				Sensitive:   true,
			},
			"username": {
				Name:        "username",
				Type:        integration.FieldTypeString,
				Description: "Username for API authentication (Server/Data Center)",
				Required:    false,
			},
			"password": {
				Name:        "password",
				Type:        integration.FieldTypeSecret,
				Description: "Password for authentication (Server/Data Center)",
				Required:    false,
				Sensitive:   true,
			},
		},
		InputSpec: map[string]integration.FieldSpec{
			"action": {
				Name:        "action",
				Type:        integration.FieldTypeString,
				Description: "Action to perform",
				Required:    true,
				Options: []string{
					string(JiraActionCreateIssue),
					string(JiraActionUpdateIssue),
					string(JiraActionGetIssue),
					string(JiraActionSearchIssues),
					string(JiraActionTransitionIssue),
					string(JiraActionAddComment),
					string(JiraActionGetTransitions),
					string(JiraActionAssignIssue),
					string(JiraActionGetProject),
					string(JiraActionListProjects),
				},
			},
			"project": {
				Name:        "project",
				Type:        integration.FieldTypeString,
				Description: "Project key (e.g., PROJ)",
				Required:    false,
			},
			"issue_key": {
				Name:        "issue_key",
				Type:        integration.FieldTypeString,
				Description: "Issue key (e.g., PROJ-123)",
				Required:    false,
			},
			"issue_type": {
				Name:        "issue_type",
				Type:        integration.FieldTypeString,
				Description: "Issue type (e.g., Task, Bug, Story)",
				Required:    false,
			},
			"summary": {
				Name:        "summary",
				Type:        integration.FieldTypeString,
				Description: "Issue summary/title",
				Required:    false,
			},
			"description": {
				Name:        "description",
				Type:        integration.FieldTypeString,
				Description: "Issue description",
				Required:    false,
			},
			"priority": {
				Name:        "priority",
				Type:        integration.FieldTypeString,
				Description: "Issue priority (e.g., High, Medium, Low)",
				Required:    false,
			},
			"assignee": {
				Name:        "assignee",
				Type:        integration.FieldTypeString,
				Description: "Assignee account ID",
				Required:    false,
			},
			"labels": {
				Name:        "labels",
				Type:        integration.FieldTypeArray,
				Description: "Array of label names",
				Required:    false,
			},
			"components": {
				Name:        "components",
				Type:        integration.FieldTypeArray,
				Description: "Array of component names",
				Required:    false,
			},
			"custom_fields": {
				Name:        "custom_fields",
				Type:        integration.FieldTypeObject,
				Description: "Custom field values (key: field ID, value: field value)",
				Required:    false,
			},
			"jql": {
				Name:        "jql",
				Type:        integration.FieldTypeString,
				Description: "JQL query for searching issues",
				Required:    false,
			},
			"transition_id": {
				Name:        "transition_id",
				Type:        integration.FieldTypeString,
				Description: "Transition ID for changing issue status",
				Required:    false,
			},
			"transition_name": {
				Name:        "transition_name",
				Type:        integration.FieldTypeString,
				Description: "Transition name for changing issue status",
				Required:    false,
			},
			"body": {
				Name:        "body",
				Type:        integration.FieldTypeString,
				Description: "Comment body text",
				Required:    false,
			},
			"comment": {
				Name:        "comment",
				Type:        integration.FieldTypeString,
				Description: "Comment to add during transition",
				Required:    false,
			},
			"resolution": {
				Name:        "resolution",
				Type:        integration.FieldTypeString,
				Description: "Resolution name for closing issues",
				Required:    false,
			},
		},
		OutputSpec: map[string]integration.FieldSpec{
			"id": {
				Name:        "id",
				Type:        integration.FieldTypeString,
				Description: "Issue ID",
			},
			"key": {
				Name:        "key",
				Type:        integration.FieldTypeString,
				Description: "Issue key",
			},
			"self": {
				Name:        "self",
				Type:        integration.FieldTypeString,
				Description: "API URL to the issue",
			},
			"issues": {
				Name:        "issues",
				Type:        integration.FieldTypeArray,
				Description: "Array of issues (for search)",
			},
			"total": {
				Name:        "total",
				Type:        integration.FieldTypeInteger,
				Description: "Total number of results",
			},
		},
	}
}

// buildJiraRetryConfig builds retry configuration for Jira API.
func buildJiraRetryConfig() *inthttp.RetryConfig {
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
