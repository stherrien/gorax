package github

import "fmt"

// CreateIssueConfig represents configuration for CreateIssue action
type CreateIssueConfig struct {
	Owner  string   `json:"owner"`
	Repo   string   `json:"repo"`
	Title  string   `json:"title"`
	Body   string   `json:"body"`
	Labels []string `json:"labels,omitempty"`
}

// Validate validates CreateIssueConfig
func (c *CreateIssueConfig) Validate() error {
	if c.Owner == "" {
		return fmt.Errorf("owner is required")
	}
	if c.Repo == "" {
		return fmt.Errorf("repo is required")
	}
	if c.Title == "" {
		return fmt.Errorf("title is required")
	}
	return nil
}

// CreatePRCommentConfig represents configuration for CreatePRComment action
type CreatePRCommentConfig struct {
	Owner  string `json:"owner"`
	Repo   string `json:"repo"`
	Number int    `json:"number"`
	Body   string `json:"body"`
}

// Validate validates CreatePRCommentConfig
func (c *CreatePRCommentConfig) Validate() error {
	if c.Owner == "" {
		return fmt.Errorf("owner is required")
	}
	if c.Repo == "" {
		return fmt.Errorf("repo is required")
	}
	if c.Number == 0 {
		return fmt.Errorf("PR number is required")
	}
	if c.Body == "" {
		return fmt.Errorf("body is required")
	}
	return nil
}

// AddLabelConfig represents configuration for AddLabel action
type AddLabelConfig struct {
	Owner  string   `json:"owner"`
	Repo   string   `json:"repo"`
	Number int      `json:"number"`
	Labels []string `json:"labels"`
}

// Validate validates AddLabelConfig
func (c *AddLabelConfig) Validate() error {
	if c.Owner == "" {
		return fmt.Errorf("owner is required")
	}
	if c.Repo == "" {
		return fmt.Errorf("repo is required")
	}
	if c.Number == 0 {
		return fmt.Errorf("issue number is required")
	}
	if len(c.Labels) == 0 {
		return fmt.Errorf("at least one label is required")
	}
	return nil
}

// Issue represents a GitHub issue
type Issue struct {
	Number int    `json:"number"`
	Title  string `json:"title"`
	Body   string `json:"body"`
	State  string `json:"state"`
	URL    string `json:"html_url"`
}

// Comment represents a GitHub comment
type Comment struct {
	ID   int    `json:"id"`
	Body string `json:"body"`
	URL  string `json:"html_url"`
}

// ErrorResponse represents a GitHub error response
type ErrorResponse struct {
	Message          string `json:"message"`
	DocumentationURL string `json:"documentation_url,omitempty"`
}
