package jira

import "fmt"

// CreateIssueRequest represents a request to create a new issue
type CreateIssueRequest struct {
	Project     string   `json:"project"`
	IssueType   string   `json:"issue_type"`
	Summary     string   `json:"summary"`
	Description string   `json:"description"`
	Priority    string   `json:"priority,omitempty"`
	Assignee    string   `json:"assignee,omitempty"`
	Labels      []string `json:"labels,omitempty"`
	Components  []string `json:"components,omitempty"`
}

// Validate validates the create issue request
func (r *CreateIssueRequest) Validate() error {
	if r.Project == "" {
		return fmt.Errorf("project is required")
	}
	if r.IssueType == "" {
		return fmt.Errorf("issue type is required")
	}
	if r.Summary == "" {
		return fmt.Errorf("summary is required")
	}
	if len(r.Summary) > 255 {
		return fmt.Errorf("summary must be less than 255 characters")
	}
	return nil
}

// Issue represents a Jira issue
type Issue struct {
	ID     string                 `json:"id"`
	Key    string                 `json:"key"`
	Self   string                 `json:"self"`
	Fields map[string]interface{} `json:"fields,omitempty"`
}

// Comment represents a Jira comment
type Comment struct {
	ID      string `json:"id"`
	Body    string `json:"body"`
	Author  User   `json:"author,omitempty"`
	Created string `json:"created,omitempty"`
	Updated string `json:"updated,omitempty"`
}

// User represents a Jira user
type User struct {
	AccountID    string `json:"accountId"`
	EmailAddress string `json:"emailAddress,omitempty"`
	DisplayName  string `json:"displayName,omitempty"`
	Active       bool   `json:"active,omitempty"`
}

// Transition represents an available issue transition
type Transition struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	To   struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"to"`
}

// SearchResult represents the result of a JQL search
type SearchResult struct {
	Total      int     `json:"total"`
	StartAt    int     `json:"startAt"`
	MaxResults int     `json:"maxResults"`
	Issues     []Issue `json:"issues"`
}

// ErrorResponse represents a Jira error response
type ErrorResponse struct {
	ErrorMessages []string          `json:"errorMessages,omitempty"`
	Errors        map[string]string `json:"errors,omitempty"`
}

// CreateIssueConfig represents configuration for CreateIssue action
type CreateIssueConfig struct {
	Project     string   `json:"project"`
	IssueType   string   `json:"issue_type"`
	Summary     string   `json:"summary"`
	Description string   `json:"description"`
	Priority    string   `json:"priority,omitempty"`
	Assignee    string   `json:"assignee,omitempty"`
	Labels      []string `json:"labels,omitempty"`
	Components  []string `json:"components,omitempty"`
}

// UpdateIssueConfig represents configuration for UpdateIssue action
type UpdateIssueConfig struct {
	IssueKey string                 `json:"issue_key"`
	Fields   map[string]interface{} `json:"fields"`
}

// AddCommentConfig represents configuration for AddComment action
type AddCommentConfig struct {
	IssueKey string `json:"issue_key"`
	Body     string `json:"body"`
}

// TransitionIssueConfig represents configuration for TransitionIssue action
type TransitionIssueConfig struct {
	IssueKey       string `json:"issue_key"`
	TransitionName string `json:"transition_name"`
}

// SearchIssuesConfig represents configuration for SearchIssues action
type SearchIssuesConfig struct {
	JQL        string `json:"jql"`
	MaxResults int    `json:"max_results,omitempty"`
	StartAt    int    `json:"start_at,omitempty"`
}

// Validate validates CreateIssueConfig
func (c *CreateIssueConfig) Validate() error {
	req := CreateIssueRequest{
		Project:     c.Project,
		IssueType:   c.IssueType,
		Summary:     c.Summary,
		Description: c.Description,
	}
	return req.Validate()
}

// Validate validates UpdateIssueConfig
func (c *UpdateIssueConfig) Validate() error {
	if c.IssueKey == "" {
		return fmt.Errorf("issue_key is required")
	}
	if len(c.Fields) == 0 {
		return fmt.Errorf("fields are required")
	}
	return nil
}

// Validate validates AddCommentConfig
func (c *AddCommentConfig) Validate() error {
	if c.IssueKey == "" {
		return fmt.Errorf("issue_key is required")
	}
	if c.Body == "" {
		return fmt.Errorf("body is required")
	}
	return nil
}

// Validate validates TransitionIssueConfig
func (c *TransitionIssueConfig) Validate() error {
	if c.IssueKey == "" {
		return fmt.Errorf("issue_key is required")
	}
	if c.TransitionName == "" {
		return fmt.Errorf("transition_name is required")
	}
	return nil
}

// Validate validates SearchIssuesConfig
func (c *SearchIssuesConfig) Validate() error {
	if c.JQL == "" {
		return fmt.Errorf("jql is required")
	}
	if c.MaxResults == 0 {
		c.MaxResults = 50 // Default
	}
	if c.MaxResults > 100 {
		return fmt.Errorf("max_results cannot exceed 100")
	}
	return nil
}
