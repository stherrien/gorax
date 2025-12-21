package jira

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/gorax/gorax/internal/integrations"
)

// CreateIssueAction implements the jira:create_issue action
type CreateIssueAction struct {
	baseURL  string
	email    string
	apiToken string
}

// NewCreateIssueAction creates a new CreateIssue action
func NewCreateIssueAction(baseURL, email, apiToken string) *CreateIssueAction {
	return &CreateIssueAction{
		baseURL:  baseURL,
		email:    email,
		apiToken: apiToken,
	}
}

// Execute implements the Action interface
func (a *CreateIssueAction) Execute(ctx context.Context, config map[string]interface{}, input map[string]interface{}) (map[string]interface{}, error) {
	// Parse config
	configJSON, err := json.Marshal(config)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal config: %w", err)
	}

	var issueConfig CreateIssueConfig
	if err := json.Unmarshal(configJSON, &issueConfig); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	if err := issueConfig.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	// Create client
	client, err := NewClient(a.baseURL, a.email, a.apiToken)
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %w", err)
	}

	// Create issue
	req := CreateIssueRequest{
		Project:     issueConfig.Project,
		IssueType:   issueConfig.IssueType,
		Summary:     issueConfig.Summary,
		Description: issueConfig.Description,
		Priority:    issueConfig.Priority,
		Assignee:    issueConfig.Assignee,
		Labels:      issueConfig.Labels,
		Components:  issueConfig.Components,
	}

	issue, err := client.CreateIssue(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to create issue: %w", err)
	}

	return map[string]interface{}{
		"id":   issue.ID,
		"key":  issue.Key,
		"self": issue.Self,
	}, nil
}

// Validate implements the Action interface
func (a *CreateIssueAction) Validate(config map[string]interface{}) error {
	configJSON, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	var issueConfig CreateIssueConfig
	if err := json.Unmarshal(configJSON, &issueConfig); err != nil {
		return fmt.Errorf("failed to parse config: %w", err)
	}

	return issueConfig.Validate()
}

// Name implements the Action interface
func (a *CreateIssueAction) Name() string {
	return "jira:create_issue"
}

// Description implements the Action interface
func (a *CreateIssueAction) Description() string {
	return "Create a new Jira issue"
}

// UpdateIssueAction implements the jira:update_issue action
type UpdateIssueAction struct {
	baseURL  string
	email    string
	apiToken string
}

// NewUpdateIssueAction creates a new UpdateIssue action
func NewUpdateIssueAction(baseURL, email, apiToken string) *UpdateIssueAction {
	return &UpdateIssueAction{
		baseURL:  baseURL,
		email:    email,
		apiToken: apiToken,
	}
}

// Execute implements the Action interface
func (a *UpdateIssueAction) Execute(ctx context.Context, config map[string]interface{}, input map[string]interface{}) (map[string]interface{}, error) {
	configJSON, err := json.Marshal(config)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal config: %w", err)
	}

	var updateConfig UpdateIssueConfig
	if err := json.Unmarshal(configJSON, &updateConfig); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	if err := updateConfig.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	client, err := NewClient(a.baseURL, a.email, a.apiToken)
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %w", err)
	}

	if err := client.UpdateIssue(ctx, updateConfig.IssueKey, updateConfig.Fields); err != nil {
		return nil, fmt.Errorf("failed to update issue: %w", err)
	}

	return map[string]interface{}{
		"success":   true,
		"issue_key": updateConfig.IssueKey,
	}, nil
}

// Validate implements the Action interface
func (a *UpdateIssueAction) Validate(config map[string]interface{}) error {
	configJSON, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	var updateConfig UpdateIssueConfig
	if err := json.Unmarshal(configJSON, &updateConfig); err != nil {
		return fmt.Errorf("failed to parse config: %w", err)
	}

	return updateConfig.Validate()
}

// Name implements the Action interface
func (a *UpdateIssueAction) Name() string {
	return "jira:update_issue"
}

// Description implements the Action interface
func (a *UpdateIssueAction) Description() string {
	return "Update an existing Jira issue"
}

// AddCommentAction implements the jira:add_comment action
type AddCommentAction struct {
	baseURL  string
	email    string
	apiToken string
}

// NewAddCommentAction creates a new AddComment action
func NewAddCommentAction(baseURL, email, apiToken string) *AddCommentAction {
	return &AddCommentAction{
		baseURL:  baseURL,
		email:    email,
		apiToken: apiToken,
	}
}

// Execute implements the Action interface
func (a *AddCommentAction) Execute(ctx context.Context, config map[string]interface{}, input map[string]interface{}) (map[string]interface{}, error) {
	configJSON, err := json.Marshal(config)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal config: %w", err)
	}

	var commentConfig AddCommentConfig
	if err := json.Unmarshal(configJSON, &commentConfig); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	if err := commentConfig.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	client, err := NewClient(a.baseURL, a.email, a.apiToken)
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %w", err)
	}

	comment, err := client.AddComment(ctx, commentConfig.IssueKey, commentConfig.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to add comment: %w", err)
	}

	return map[string]interface{}{
		"id":   comment.ID,
		"body": comment.Body,
	}, nil
}

// Validate implements the Action interface
func (a *AddCommentAction) Validate(config map[string]interface{}) error {
	configJSON, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	var commentConfig AddCommentConfig
	if err := json.Unmarshal(configJSON, &commentConfig); err != nil {
		return fmt.Errorf("failed to parse config: %w", err)
	}

	return commentConfig.Validate()
}

// Name implements the Action interface
func (a *AddCommentAction) Name() string {
	return "jira:add_comment"
}

// Description implements the Action interface
func (a *AddCommentAction) Description() string {
	return "Add a comment to a Jira issue"
}

// TransitionIssueAction implements the jira:transition_issue action
type TransitionIssueAction struct {
	baseURL  string
	email    string
	apiToken string
}

// NewTransitionIssueAction creates a new TransitionIssue action
func NewTransitionIssueAction(baseURL, email, apiToken string) *TransitionIssueAction {
	return &TransitionIssueAction{
		baseURL:  baseURL,
		email:    email,
		apiToken: apiToken,
	}
}

// Execute implements the Action interface
func (a *TransitionIssueAction) Execute(ctx context.Context, config map[string]interface{}, input map[string]interface{}) (map[string]interface{}, error) {
	configJSON, err := json.Marshal(config)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal config: %w", err)
	}

	var transitionConfig TransitionIssueConfig
	if err := json.Unmarshal(configJSON, &transitionConfig); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	if err := transitionConfig.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	client, err := NewClient(a.baseURL, a.email, a.apiToken)
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %w", err)
	}

	if err := client.TransitionIssue(ctx, transitionConfig.IssueKey, transitionConfig.TransitionName); err != nil {
		return nil, fmt.Errorf("failed to transition issue: %w", err)
	}

	return map[string]interface{}{
		"success":         true,
		"issue_key":       transitionConfig.IssueKey,
		"transition_name": transitionConfig.TransitionName,
	}, nil
}

// Validate implements the Action interface
func (a *TransitionIssueAction) Validate(config map[string]interface{}) error {
	configJSON, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	var transitionConfig TransitionIssueConfig
	if err := json.Unmarshal(configJSON, &transitionConfig); err != nil {
		return fmt.Errorf("failed to parse config: %w", err)
	}

	return transitionConfig.Validate()
}

// Name implements the Action interface
func (a *TransitionIssueAction) Name() string {
	return "jira:transition_issue"
}

// Description implements the Action interface
func (a *TransitionIssueAction) Description() string {
	return "Transition a Jira issue to a new status"
}

// SearchIssuesAction implements the jira:search_issues action
type SearchIssuesAction struct {
	baseURL  string
	email    string
	apiToken string
}

// NewSearchIssuesAction creates a new SearchIssues action
func NewSearchIssuesAction(baseURL, email, apiToken string) *SearchIssuesAction {
	return &SearchIssuesAction{
		baseURL:  baseURL,
		email:    email,
		apiToken: apiToken,
	}
}

// Execute implements the Action interface
func (a *SearchIssuesAction) Execute(ctx context.Context, config map[string]interface{}, input map[string]interface{}) (map[string]interface{}, error) {
	configJSON, err := json.Marshal(config)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal config: %w", err)
	}

	var searchConfig SearchIssuesConfig
	if err := json.Unmarshal(configJSON, &searchConfig); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	if err := searchConfig.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	client, err := NewClient(a.baseURL, a.email, a.apiToken)
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %w", err)
	}

	result, err := client.SearchIssues(ctx, searchConfig.JQL, searchConfig.MaxResults, searchConfig.StartAt)
	if err != nil {
		return nil, fmt.Errorf("failed to search issues: %w", err)
	}

	// Convert result to map
	resultJSON, err := json.Marshal(result)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal result: %w", err)
	}

	var resultMap map[string]interface{}
	if err := json.Unmarshal(resultJSON, &resultMap); err != nil {
		return nil, fmt.Errorf("failed to unmarshal result: %w", err)
	}

	return resultMap, nil
}

// Validate implements the Action interface
func (a *SearchIssuesAction) Validate(config map[string]interface{}) error {
	configJSON, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	var searchConfig SearchIssuesConfig
	if err := json.Unmarshal(configJSON, &searchConfig); err != nil {
		return fmt.Errorf("failed to parse config: %w", err)
	}

	return searchConfig.Validate()
}

// Name implements the Action interface
func (a *SearchIssuesAction) Name() string {
	return "jira:search_issues"
}

// Description implements the Action interface
func (a *SearchIssuesAction) Description() string {
	return "Search for Jira issues using JQL"
}

// Ensure all actions implement the Action interface
var (
	_ integrations.Action = (*CreateIssueAction)(nil)
	_ integrations.Action = (*UpdateIssueAction)(nil)
	_ integrations.Action = (*AddCommentAction)(nil)
	_ integrations.Action = (*TransitionIssueAction)(nil)
	_ integrations.Action = (*SearchIssuesAction)(nil)
)
