package github

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/gorax/gorax/internal/integrations"
)

// CreateIssueAction implements the github:create_issue action
type CreateIssueAction struct {
	token string
}

// NewCreateIssueAction creates a new CreateIssue action
func NewCreateIssueAction(token string) *CreateIssueAction {
	return &CreateIssueAction{token: token}
}

// Execute implements the Action interface
func (a *CreateIssueAction) Execute(ctx context.Context, config map[string]interface{}, input map[string]interface{}) (map[string]interface{}, error) {
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

	client, err := NewClient(a.token)
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %w", err)
	}

	issue, err := client.CreateIssue(ctx, issueConfig.Owner, issueConfig.Repo, issueConfig.Title, issueConfig.Body, issueConfig.Labels)
	if err != nil {
		return nil, fmt.Errorf("failed to create issue: %w", err)
	}

	return map[string]interface{}{
		"number": issue.Number,
		"title":  issue.Title,
		"url":    issue.URL,
		"state":  issue.State,
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
	return "github:create_issue"
}

// Description implements the Action interface
func (a *CreateIssueAction) Description() string {
	return "Create a new GitHub issue"
}

// CreatePRCommentAction implements the github:create_pr_comment action
type CreatePRCommentAction struct {
	token string
}

// NewCreatePRCommentAction creates a new CreatePRComment action
func NewCreatePRCommentAction(token string) *CreatePRCommentAction {
	return &CreatePRCommentAction{token: token}
}

// Execute implements the Action interface
func (a *CreatePRCommentAction) Execute(ctx context.Context, config map[string]interface{}, input map[string]interface{}) (map[string]interface{}, error) {
	configJSON, err := json.Marshal(config)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal config: %w", err)
	}

	var commentConfig CreatePRCommentConfig
	if err := json.Unmarshal(configJSON, &commentConfig); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	if err := commentConfig.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	client, err := NewClient(a.token)
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %w", err)
	}

	comment, err := client.CreatePRComment(ctx, commentConfig.Owner, commentConfig.Repo, commentConfig.Number, commentConfig.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to create PR comment: %w", err)
	}

	return map[string]interface{}{
		"id":   comment.ID,
		"body": comment.Body,
		"url":  comment.URL,
	}, nil
}

// Validate implements the Action interface
func (a *CreatePRCommentAction) Validate(config map[string]interface{}) error {
	configJSON, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	var commentConfig CreatePRCommentConfig
	if err := json.Unmarshal(configJSON, &commentConfig); err != nil {
		return fmt.Errorf("failed to parse config: %w", err)
	}

	return commentConfig.Validate()
}

// Name implements the Action interface
func (a *CreatePRCommentAction) Name() string {
	return "github:create_pr_comment"
}

// Description implements the Action interface
func (a *CreatePRCommentAction) Description() string {
	return "Create a comment on a GitHub pull request"
}

// AddLabelAction implements the github:add_label action
type AddLabelAction struct {
	token string
}

// NewAddLabelAction creates a new AddLabel action
func NewAddLabelAction(token string) *AddLabelAction {
	return &AddLabelAction{token: token}
}

// Execute implements the Action interface
func (a *AddLabelAction) Execute(ctx context.Context, config map[string]interface{}, input map[string]interface{}) (map[string]interface{}, error) {
	configJSON, err := json.Marshal(config)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal config: %w", err)
	}

	var labelConfig AddLabelConfig
	if err := json.Unmarshal(configJSON, &labelConfig); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	if err := labelConfig.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	client, err := NewClient(a.token)
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %w", err)
	}

	if err := client.AddLabels(ctx, labelConfig.Owner, labelConfig.Repo, labelConfig.Number, labelConfig.Labels); err != nil {
		return nil, fmt.Errorf("failed to add labels: %w", err)
	}

	return map[string]interface{}{
		"success": true,
		"labels":  labelConfig.Labels,
	}, nil
}

// Validate implements the Action interface
func (a *AddLabelAction) Validate(config map[string]interface{}) error {
	configJSON, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	var labelConfig AddLabelConfig
	if err := json.Unmarshal(configJSON, &labelConfig); err != nil {
		return fmt.Errorf("failed to parse config: %w", err)
	}

	return labelConfig.Validate()
}

// Name implements the Action interface
func (a *AddLabelAction) Name() string {
	return "github:add_label"
}

// Description implements the Action interface
func (a *AddLabelAction) Description() string {
	return "Add labels to a GitHub issue or pull request"
}

// Ensure all actions implement the Action interface
var (
	_ integrations.Action = (*CreateIssueAction)(nil)
	_ integrations.Action = (*CreatePRCommentAction)(nil)
	_ integrations.Action = (*AddLabelAction)(nil)
)
