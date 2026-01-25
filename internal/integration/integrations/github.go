package integrations

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorax/gorax/internal/integration"
	inthttp "github.com/gorax/gorax/internal/integration/http"
)

const (
	githubAPIBaseURL = "https://api.github.com"
	githubIntName    = "github"
)

// GitHubIntegration provides GitHub API integration capabilities.
type GitHubIntegration struct {
	*integration.BaseIntegration
	client *inthttp.Client
	logger *slog.Logger
}

// GitHubAction represents the available GitHub actions.
type GitHubAction string

const (
	// GitHubActionCreateIssue creates a new issue.
	GitHubActionCreateIssue GitHubAction = "create_issue"
	// GitHubActionUpdateIssue updates an existing issue.
	GitHubActionUpdateIssue GitHubAction = "update_issue"
	// GitHubActionGetIssue gets issue details.
	GitHubActionGetIssue GitHubAction = "get_issue"
	// GitHubActionListIssues lists issues with filters.
	GitHubActionListIssues GitHubAction = "list_issues"
	// GitHubActionCreatePullRequest creates a pull request.
	GitHubActionCreatePullRequest GitHubAction = "create_pull_request"
	// GitHubActionMergePullRequest merges a pull request.
	GitHubActionMergePullRequest GitHubAction = "merge_pull_request"
	// GitHubActionGetPullRequest gets pull request details.
	GitHubActionGetPullRequest GitHubAction = "get_pull_request"
	// GitHubActionCreateBranch creates a new branch.
	GitHubActionCreateBranch GitHubAction = "create_branch"
	// GitHubActionCreateWebhook creates a repository webhook.
	GitHubActionCreateWebhook GitHubAction = "create_webhook"
	// GitHubActionGetRepository gets repository information.
	GitHubActionGetRepository GitHubAction = "get_repository"
	// GitHubActionAddComment adds a comment to an issue or PR.
	GitHubActionAddComment GitHubAction = "add_comment"
	// GitHubActionAddLabel adds labels to an issue or PR.
	GitHubActionAddLabel GitHubAction = "add_label"
)

// NewGitHubIntegration creates a new GitHub integration.
func NewGitHubIntegration(logger *slog.Logger) *GitHubIntegration {
	if logger == nil {
		logger = slog.Default()
	}

	base := integration.NewBaseIntegration(githubIntName, integration.TypeAPI)
	base.SetMetadata(&integration.Metadata{
		Name:        githubIntName,
		DisplayName: "GitHub",
		Description: "Manage repositories, issues, pull requests, and webhooks on GitHub",
		Version:     "1.0.0",
		Category:    "version_control",
		Tags:        []string{"github", "git", "version-control", "issues", "pull-requests"},
		Author:      "Gorax",
	})
	base.SetSchema(buildGitHubSchema())

	client := inthttp.NewClient(
		inthttp.WithBaseURL(githubAPIBaseURL),
		inthttp.WithTimeout(30*time.Second),
		inthttp.WithLogger(logger),
		inthttp.WithRetryConfig(buildGitHubRetryConfig()),
		inthttp.WithHeader("Accept", "application/vnd.github+json"),
		inthttp.WithHeader("X-GitHub-Api-Version", "2022-11-28"),
	)

	return &GitHubIntegration{
		BaseIntegration: base,
		client:          client,
		logger:          logger,
	}
}

// Execute performs a GitHub API action.
func (g *GitHubIntegration) Execute(ctx context.Context, config *integration.Config, params integration.JSONMap) (*integration.Result, error) {
	start := time.Now()

	action, ok := params.GetString("action")
	if !ok || action == "" {
		err := integration.NewValidationError("action", "action is required", nil)
		return integration.NewErrorResult(err, "VALIDATION_ERROR", time.Since(start).Milliseconds()), err
	}

	token, err := g.getToken(config)
	if err != nil {
		return integration.NewErrorResult(err, "AUTH_ERROR", time.Since(start).Milliseconds()), err
	}

	var result *integration.Result
	switch GitHubAction(action) {
	case GitHubActionCreateIssue:
		result, err = g.createIssue(ctx, token, params, start)
	case GitHubActionUpdateIssue:
		result, err = g.updateIssue(ctx, token, params, start)
	case GitHubActionGetIssue:
		result, err = g.getIssue(ctx, token, params, start)
	case GitHubActionListIssues:
		result, err = g.listIssues(ctx, token, params, start)
	case GitHubActionCreatePullRequest:
		result, err = g.createPullRequest(ctx, token, params, start)
	case GitHubActionMergePullRequest:
		result, err = g.mergePullRequest(ctx, token, params, start)
	case GitHubActionGetPullRequest:
		result, err = g.getPullRequest(ctx, token, params, start)
	case GitHubActionCreateBranch:
		result, err = g.createBranch(ctx, token, params, start)
	case GitHubActionCreateWebhook:
		result, err = g.createWebhook(ctx, token, params, start)
	case GitHubActionGetRepository:
		result, err = g.getRepository(ctx, token, params, start)
	case GitHubActionAddComment:
		result, err = g.addComment(ctx, token, params, start)
	case GitHubActionAddLabel:
		result, err = g.addLabel(ctx, token, params, start)
	default:
		err = integration.NewValidationError("action", "unsupported action", action)
		result = integration.NewErrorResult(err, "INVALID_ACTION", time.Since(start).Milliseconds())
	}

	if err != nil {
		g.logger.Error("github action failed",
			"action", action,
			"error", err,
			"duration_ms", time.Since(start).Milliseconds(),
		)
	} else {
		g.logger.Info("github action completed",
			"action", action,
			"duration_ms", result.Duration,
		)
	}

	return result, err
}

// Validate validates the integration configuration.
func (g *GitHubIntegration) Validate(config *integration.Config) error {
	if err := g.BaseIntegration.ValidateConfig(config); err != nil {
		return err
	}

	if config.Credentials == nil {
		return integration.NewValidationError("credentials", "credentials are required", nil)
	}

	if _, err := g.getToken(config); err != nil {
		return err
	}

	return nil
}

// getToken extracts the authentication token from credentials.
func (g *GitHubIntegration) getToken(config *integration.Config) (string, error) {
	if config.Credentials == nil || config.Credentials.Data == nil {
		return "", integration.NewValidationError("credentials", "credentials are required", nil)
	}

	// Try different token field names
	for _, key := range []string{"token", "access_token", "personal_access_token"} {
		if token, ok := config.Credentials.Data.GetString(key); ok && token != "" {
			return token, nil
		}
	}

	return "", integration.NewValidationError("token", "GitHub token is required", nil)
}

// getOwnerRepo extracts owner and repo from params.
func (g *GitHubIntegration) getOwnerRepo(params integration.JSONMap) (string, string, error) {
	owner, _ := params.GetString("owner")
	repo, _ := params.GetString("repo")

	// Also support "repository" as "owner/repo" format
	if owner == "" || repo == "" {
		if repository, ok := params.GetString("repository"); ok && repository != "" {
			parts := strings.SplitN(repository, "/", 2)
			if len(parts) == 2 {
				owner = parts[0]
				repo = parts[1]
			}
		}
	}

	if owner == "" {
		return "", "", integration.NewValidationError("owner", "repository owner is required", nil)
	}
	if repo == "" {
		return "", "", integration.NewValidationError("repo", "repository name is required", nil)
	}

	return owner, repo, nil
}

// createIssue creates a new GitHub issue.
func (g *GitHubIntegration) createIssue(ctx context.Context, token string, params integration.JSONMap, start time.Time) (*integration.Result, error) {
	owner, repo, err := g.getOwnerRepo(params)
	if err != nil {
		return integration.NewErrorResult(err, "VALIDATION_ERROR", time.Since(start).Milliseconds()), err
	}

	title, ok := params.GetString("title")
	if !ok || title == "" {
		err := integration.NewValidationError("title", "issue title is required", nil)
		return integration.NewErrorResult(err, "VALIDATION_ERROR", time.Since(start).Milliseconds()), err
	}

	payload := map[string]any{
		"title": title,
	}

	if body, ok := params.GetString("body"); ok {
		payload["body"] = body
	}

	if labels, ok := params.Get("labels"); ok {
		payload["labels"] = labels
	}

	if assignees, ok := params.Get("assignees"); ok {
		payload["assignees"] = assignees
	}

	if milestone, ok := params.GetInt("milestone"); ok {
		payload["milestone"] = milestone
	}

	path := fmt.Sprintf("/repos/%s/%s/issues", owner, repo)
	return g.executeGitHubAPI(ctx, http.MethodPost, token, path, payload, start)
}

// updateIssue updates an existing GitHub issue.
func (g *GitHubIntegration) updateIssue(ctx context.Context, token string, params integration.JSONMap, start time.Time) (*integration.Result, error) {
	owner, repo, err := g.getOwnerRepo(params)
	if err != nil {
		return integration.NewErrorResult(err, "VALIDATION_ERROR", time.Since(start).Milliseconds()), err
	}

	issueNumber, ok := params.GetInt("issue_number")
	if !ok {
		err := integration.NewValidationError("issue_number", "issue number is required", nil)
		return integration.NewErrorResult(err, "VALIDATION_ERROR", time.Since(start).Milliseconds()), err
	}

	payload := make(map[string]any)

	if title, ok := params.GetString("title"); ok {
		payload["title"] = title
	}
	if body, ok := params.GetString("body"); ok {
		payload["body"] = body
	}
	if state, ok := params.GetString("state"); ok {
		payload["state"] = state
	}
	if labels, ok := params.Get("labels"); ok {
		payload["labels"] = labels
	}
	if assignees, ok := params.Get("assignees"); ok {
		payload["assignees"] = assignees
	}
	if milestone, ok := params.GetInt("milestone"); ok {
		payload["milestone"] = milestone
	}

	path := fmt.Sprintf("/repos/%s/%s/issues/%d", owner, repo, issueNumber)
	return g.executeGitHubAPI(ctx, http.MethodPatch, token, path, payload, start)
}

// getIssue gets issue details.
func (g *GitHubIntegration) getIssue(ctx context.Context, token string, params integration.JSONMap, start time.Time) (*integration.Result, error) {
	owner, repo, err := g.getOwnerRepo(params)
	if err != nil {
		return integration.NewErrorResult(err, "VALIDATION_ERROR", time.Since(start).Milliseconds()), err
	}

	issueNumber, ok := params.GetInt("issue_number")
	if !ok {
		err := integration.NewValidationError("issue_number", "issue number is required", nil)
		return integration.NewErrorResult(err, "VALIDATION_ERROR", time.Since(start).Milliseconds()), err
	}

	path := fmt.Sprintf("/repos/%s/%s/issues/%d", owner, repo, issueNumber)
	return g.executeGitHubAPI(ctx, http.MethodGet, token, path, nil, start)
}

// listIssues lists issues with filters.
func (g *GitHubIntegration) listIssues(ctx context.Context, token string, params integration.JSONMap, start time.Time) (*integration.Result, error) {
	owner, repo, err := g.getOwnerRepo(params)
	if err != nil {
		return integration.NewErrorResult(err, "VALIDATION_ERROR", time.Since(start).Milliseconds()), err
	}

	queryParams := make(map[string]string)

	if state, ok := params.GetString("state"); ok {
		queryParams["state"] = state
	}
	if labels, ok := params.GetString("labels"); ok {
		queryParams["labels"] = labels
	}
	if sort, ok := params.GetString("sort"); ok {
		queryParams["sort"] = sort
	}
	if direction, ok := params.GetString("direction"); ok {
		queryParams["direction"] = direction
	}
	if perPage, ok := params.GetInt("per_page"); ok {
		queryParams["per_page"] = strconv.Itoa(perPage)
	}
	if page, ok := params.GetInt("page"); ok {
		queryParams["page"] = strconv.Itoa(page)
	}

	path := fmt.Sprintf("/repos/%s/%s/issues", owner, repo)
	return g.executeGitHubAPIWithQuery(ctx, http.MethodGet, token, path, nil, queryParams, start)
}

// createPullRequest creates a new pull request.
func (g *GitHubIntegration) createPullRequest(ctx context.Context, token string, params integration.JSONMap, start time.Time) (*integration.Result, error) {
	owner, repo, err := g.getOwnerRepo(params)
	if err != nil {
		return integration.NewErrorResult(err, "VALIDATION_ERROR", time.Since(start).Milliseconds()), err
	}

	title, ok := params.GetString("title")
	if !ok || title == "" {
		err := integration.NewValidationError("title", "pull request title is required", nil)
		return integration.NewErrorResult(err, "VALIDATION_ERROR", time.Since(start).Milliseconds()), err
	}

	head, ok := params.GetString("head")
	if !ok || head == "" {
		err := integration.NewValidationError("head", "head branch is required", nil)
		return integration.NewErrorResult(err, "VALIDATION_ERROR", time.Since(start).Milliseconds()), err
	}

	base, ok := params.GetString("base")
	if !ok || base == "" {
		err := integration.NewValidationError("base", "base branch is required", nil)
		return integration.NewErrorResult(err, "VALIDATION_ERROR", time.Since(start).Milliseconds()), err
	}

	payload := map[string]any{
		"title": title,
		"head":  head,
		"base":  base,
	}

	if body, ok := params.GetString("body"); ok {
		payload["body"] = body
	}
	if draft, ok := params.GetBool("draft"); ok {
		payload["draft"] = draft
	}
	if maintainerCanModify, ok := params.GetBool("maintainer_can_modify"); ok {
		payload["maintainer_can_modify"] = maintainerCanModify
	}

	path := fmt.Sprintf("/repos/%s/%s/pulls", owner, repo)
	return g.executeGitHubAPI(ctx, http.MethodPost, token, path, payload, start)
}

// mergePullRequest merges a pull request.
func (g *GitHubIntegration) mergePullRequest(ctx context.Context, token string, params integration.JSONMap, start time.Time) (*integration.Result, error) {
	owner, repo, err := g.getOwnerRepo(params)
	if err != nil {
		return integration.NewErrorResult(err, "VALIDATION_ERROR", time.Since(start).Milliseconds()), err
	}

	pullNumber, ok := params.GetInt("pull_number")
	if !ok {
		err := integration.NewValidationError("pull_number", "pull request number is required", nil)
		return integration.NewErrorResult(err, "VALIDATION_ERROR", time.Since(start).Milliseconds()), err
	}

	payload := make(map[string]any)

	if commitTitle, ok := params.GetString("commit_title"); ok {
		payload["commit_title"] = commitTitle
	}
	if commitMessage, ok := params.GetString("commit_message"); ok {
		payload["commit_message"] = commitMessage
	}
	if sha, ok := params.GetString("sha"); ok {
		payload["sha"] = sha
	}
	// merge_method: merge, squash, or rebase
	if mergeMethod, ok := params.GetString("merge_method"); ok {
		payload["merge_method"] = mergeMethod
	}

	path := fmt.Sprintf("/repos/%s/%s/pulls/%d/merge", owner, repo, pullNumber)
	return g.executeGitHubAPI(ctx, http.MethodPut, token, path, payload, start)
}

// getPullRequest gets pull request details.
func (g *GitHubIntegration) getPullRequest(ctx context.Context, token string, params integration.JSONMap, start time.Time) (*integration.Result, error) {
	owner, repo, err := g.getOwnerRepo(params)
	if err != nil {
		return integration.NewErrorResult(err, "VALIDATION_ERROR", time.Since(start).Milliseconds()), err
	}

	pullNumber, ok := params.GetInt("pull_number")
	if !ok {
		err := integration.NewValidationError("pull_number", "pull request number is required", nil)
		return integration.NewErrorResult(err, "VALIDATION_ERROR", time.Since(start).Milliseconds()), err
	}

	path := fmt.Sprintf("/repos/%s/%s/pulls/%d", owner, repo, pullNumber)
	return g.executeGitHubAPI(ctx, http.MethodGet, token, path, nil, start)
}

// createBranch creates a new branch.
func (g *GitHubIntegration) createBranch(ctx context.Context, token string, params integration.JSONMap, start time.Time) (*integration.Result, error) {
	owner, repo, err := g.getOwnerRepo(params)
	if err != nil {
		return integration.NewErrorResult(err, "VALIDATION_ERROR", time.Since(start).Milliseconds()), err
	}

	branchName, ok := params.GetString("branch")
	if !ok || branchName == "" {
		err := integration.NewValidationError("branch", "branch name is required", nil)
		return integration.NewErrorResult(err, "VALIDATION_ERROR", time.Since(start).Milliseconds()), err
	}

	// Get the SHA to branch from
	sha, hasSHA := params.GetString("sha")
	fromBranch, hasFromBranch := params.GetString("from_branch")

	if !hasSHA && !hasFromBranch {
		// Default to main branch
		fromBranch = "main"
		hasFromBranch = true
	}

	// If we need to look up the SHA from a branch
	if !hasSHA && hasFromBranch {
		refPath := fmt.Sprintf("/repos/%s/%s/git/ref/heads/%s", owner, repo, fromBranch)
		refResult, err := g.executeGitHubAPI(ctx, http.MethodGet, token, refPath, nil, start)
		if err != nil {
			return refResult, err
		}

		if data, ok := refResult.Data.(map[string]any); ok {
			if obj, ok := data["object"].(map[string]any); ok {
				if s, ok := obj["sha"].(string); ok {
					sha = s
				}
			}
		}

		if sha == "" {
			err := fmt.Errorf("could not find SHA for branch: %s", fromBranch)
			return integration.NewErrorResult(err, "BRANCH_NOT_FOUND", time.Since(start).Milliseconds()), err
		}
	}

	payload := map[string]any{
		"ref": "refs/heads/" + branchName,
		"sha": sha,
	}

	path := fmt.Sprintf("/repos/%s/%s/git/refs", owner, repo)
	return g.executeGitHubAPI(ctx, http.MethodPost, token, path, payload, start)
}

// createWebhook creates a repository webhook.
func (g *GitHubIntegration) createWebhook(ctx context.Context, token string, params integration.JSONMap, start time.Time) (*integration.Result, error) {
	owner, repo, err := g.getOwnerRepo(params)
	if err != nil {
		return integration.NewErrorResult(err, "VALIDATION_ERROR", time.Since(start).Milliseconds()), err
	}

	url, ok := params.GetString("url")
	if !ok || url == "" {
		err := integration.NewValidationError("url", "webhook URL is required", nil)
		return integration.NewErrorResult(err, "VALIDATION_ERROR", time.Since(start).Milliseconds()), err
	}

	config := map[string]any{
		"url":          url,
		"content_type": "json",
	}

	if secret, ok := params.GetString("secret"); ok {
		config["secret"] = secret
	}

	if insecureSSL, ok := params.GetString("insecure_ssl"); ok {
		config["insecure_ssl"] = insecureSSL
	}

	payload := map[string]any{
		"name":   "web",
		"config": config,
		"active": true,
	}

	if events, ok := params.Get("events"); ok {
		payload["events"] = events
	} else {
		payload["events"] = []string{"push", "pull_request"}
	}

	path := fmt.Sprintf("/repos/%s/%s/hooks", owner, repo)
	return g.executeGitHubAPI(ctx, http.MethodPost, token, path, payload, start)
}

// getRepository gets repository information.
func (g *GitHubIntegration) getRepository(ctx context.Context, token string, params integration.JSONMap, start time.Time) (*integration.Result, error) {
	owner, repo, err := g.getOwnerRepo(params)
	if err != nil {
		return integration.NewErrorResult(err, "VALIDATION_ERROR", time.Since(start).Milliseconds()), err
	}

	path := fmt.Sprintf("/repos/%s/%s", owner, repo)
	return g.executeGitHubAPI(ctx, http.MethodGet, token, path, nil, start)
}

// addComment adds a comment to an issue or PR.
func (g *GitHubIntegration) addComment(ctx context.Context, token string, params integration.JSONMap, start time.Time) (*integration.Result, error) {
	owner, repo, err := g.getOwnerRepo(params)
	if err != nil {
		return integration.NewErrorResult(err, "VALIDATION_ERROR", time.Since(start).Milliseconds()), err
	}

	issueNumber, ok := params.GetInt("issue_number")
	if !ok {
		err := integration.NewValidationError("issue_number", "issue/PR number is required", nil)
		return integration.NewErrorResult(err, "VALIDATION_ERROR", time.Since(start).Milliseconds()), err
	}

	body, ok := params.GetString("body")
	if !ok || body == "" {
		err := integration.NewValidationError("body", "comment body is required", nil)
		return integration.NewErrorResult(err, "VALIDATION_ERROR", time.Since(start).Milliseconds()), err
	}

	payload := map[string]any{
		"body": body,
	}

	path := fmt.Sprintf("/repos/%s/%s/issues/%d/comments", owner, repo, issueNumber)
	return g.executeGitHubAPI(ctx, http.MethodPost, token, path, payload, start)
}

// addLabel adds labels to an issue or PR.
func (g *GitHubIntegration) addLabel(ctx context.Context, token string, params integration.JSONMap, start time.Time) (*integration.Result, error) {
	owner, repo, err := g.getOwnerRepo(params)
	if err != nil {
		return integration.NewErrorResult(err, "VALIDATION_ERROR", time.Since(start).Milliseconds()), err
	}

	issueNumber, ok := params.GetInt("issue_number")
	if !ok {
		err := integration.NewValidationError("issue_number", "issue/PR number is required", nil)
		return integration.NewErrorResult(err, "VALIDATION_ERROR", time.Since(start).Milliseconds()), err
	}

	labels, ok := params.Get("labels")
	if !ok {
		err := integration.NewValidationError("labels", "labels are required", nil)
		return integration.NewErrorResult(err, "VALIDATION_ERROR", time.Since(start).Milliseconds()), err
	}

	payload := map[string]any{
		"labels": labels,
	}

	path := fmt.Sprintf("/repos/%s/%s/issues/%d/labels", owner, repo, issueNumber)
	return g.executeGitHubAPI(ctx, http.MethodPost, token, path, payload, start)
}

// executeGitHubAPI executes a GitHub API request.
func (g *GitHubIntegration) executeGitHubAPI(ctx context.Context, method, token, path string, payload map[string]any, start time.Time) (*integration.Result, error) {
	return g.executeGitHubAPIWithQuery(ctx, method, token, path, payload, nil, start)
}

// executeGitHubAPIWithQuery executes a GitHub API request with query parameters.
func (g *GitHubIntegration) executeGitHubAPIWithQuery(ctx context.Context, method, token, path string, payload map[string]any, queryParams map[string]string, start time.Time) (*integration.Result, error) {
	opts := []inthttp.RequestOption{
		inthttp.WithRequestHeader("Authorization", "Bearer "+token),
	}

	if queryParams != nil {
		opts = append(opts, inthttp.WithQueryParams(queryParams))
	}

	var resp *inthttp.Response
	var err error

	switch method {
	case http.MethodGet:
		resp, err = g.client.Get(ctx, path, opts...)
	case http.MethodPost:
		resp, err = g.client.Post(ctx, path, payload, opts...)
	case http.MethodPatch:
		resp, err = g.client.Patch(ctx, path, payload, opts...)
	case http.MethodPut:
		resp, err = g.client.Put(ctx, path, payload, opts...)
	case http.MethodDelete:
		resp, err = g.client.Delete(ctx, path, opts...)
	default:
		err = fmt.Errorf("unsupported HTTP method: %s", method)
	}

	if err != nil {
		return integration.NewErrorResult(err, "API_ERROR", time.Since(start).Milliseconds()), err
	}

	// Parse response
	var data any
	if len(resp.Body) > 0 {
		if err := json.Unmarshal(resp.Body, &data); err != nil {
			// Return raw response if not JSON
			data = string(resp.Body)
		}
	}

	// Handle pagination headers
	metadata := integration.JSONMap{}
	if link := resp.Headers.Get("Link"); link != "" {
		metadata["link"] = link
	}
	if rateLimit := resp.Headers.Get("X-RateLimit-Remaining"); rateLimit != "" {
		metadata["rate_limit_remaining"] = rateLimit
	}
	if rateReset := resp.Headers.Get("X-RateLimit-Reset"); rateReset != "" {
		metadata["rate_limit_reset"] = rateReset
	}

	return &integration.Result{
		Success:    resp.IsSuccess(),
		Data:       data,
		StatusCode: resp.StatusCode,
		Duration:   time.Since(start).Milliseconds(),
		Metadata:   metadata,
		ExecutedAt: time.Now().UTC(),
	}, nil
}

// VerifyGitHubWebhookSignature verifies a GitHub webhook signature.
func VerifyGitHubWebhookSignature(payload []byte, signature, secret string) bool {
	if !strings.HasPrefix(signature, "sha256=") {
		return false
	}

	expectedMAC := strings.TrimPrefix(signature, "sha256=")

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	actualMAC := hex.EncodeToString(mac.Sum(nil))

	return hmac.Equal([]byte(expectedMAC), []byte(actualMAC))
}

// buildGitHubSchema builds the schema for the GitHub integration.
func buildGitHubSchema() *integration.Schema {
	return &integration.Schema{
		ConfigSpec: map[string]integration.FieldSpec{
			"token": {
				Name:        "token",
				Type:        integration.FieldTypeSecret,
				Description: "GitHub personal access token or App installation token",
				Required:    true,
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
					string(GitHubActionCreateIssue),
					string(GitHubActionUpdateIssue),
					string(GitHubActionGetIssue),
					string(GitHubActionListIssues),
					string(GitHubActionCreatePullRequest),
					string(GitHubActionMergePullRequest),
					string(GitHubActionGetPullRequest),
					string(GitHubActionCreateBranch),
					string(GitHubActionCreateWebhook),
					string(GitHubActionGetRepository),
					string(GitHubActionAddComment),
					string(GitHubActionAddLabel),
				},
			},
			"owner": {
				Name:        "owner",
				Type:        integration.FieldTypeString,
				Description: "Repository owner (username or organization)",
				Required:    false,
			},
			"repo": {
				Name:        "repo",
				Type:        integration.FieldTypeString,
				Description: "Repository name",
				Required:    false,
			},
			"repository": {
				Name:        "repository",
				Type:        integration.FieldTypeString,
				Description: "Repository in owner/repo format",
				Required:    false,
			},
			"title": {
				Name:        "title",
				Type:        integration.FieldTypeString,
				Description: "Issue or PR title",
				Required:    false,
			},
			"body": {
				Name:        "body",
				Type:        integration.FieldTypeString,
				Description: "Issue, PR, or comment body",
				Required:    false,
			},
			"labels": {
				Name:        "labels",
				Type:        integration.FieldTypeArray,
				Description: "Array of label names",
				Required:    false,
			},
			"assignees": {
				Name:        "assignees",
				Type:        integration.FieldTypeArray,
				Description: "Array of assignee usernames",
				Required:    false,
			},
			"milestone": {
				Name:        "milestone",
				Type:        integration.FieldTypeInteger,
				Description: "Milestone number",
				Required:    false,
			},
			"issue_number": {
				Name:        "issue_number",
				Type:        integration.FieldTypeInteger,
				Description: "Issue or PR number",
				Required:    false,
			},
			"pull_number": {
				Name:        "pull_number",
				Type:        integration.FieldTypeInteger,
				Description: "Pull request number",
				Required:    false,
			},
			"head": {
				Name:        "head",
				Type:        integration.FieldTypeString,
				Description: "Head branch for PR",
				Required:    false,
			},
			"base": {
				Name:        "base",
				Type:        integration.FieldTypeString,
				Description: "Base branch for PR",
				Required:    false,
			},
			"branch": {
				Name:        "branch",
				Type:        integration.FieldTypeString,
				Description: "Branch name to create",
				Required:    false,
			},
			"from_branch": {
				Name:        "from_branch",
				Type:        integration.FieldTypeString,
				Description: "Source branch to create from",
				Required:    false,
			},
			"sha": {
				Name:        "sha",
				Type:        integration.FieldTypeString,
				Description: "Git SHA to use",
				Required:    false,
			},
			"merge_method": {
				Name:        "merge_method",
				Type:        integration.FieldTypeString,
				Description: "Merge method (merge, squash, rebase)",
				Required:    false,
				Options:     []string{"merge", "squash", "rebase"},
			},
			"state": {
				Name:        "state",
				Type:        integration.FieldTypeString,
				Description: "Issue state (open, closed)",
				Required:    false,
				Options:     []string{"open", "closed", "all"},
			},
			"url": {
				Name:        "url",
				Type:        integration.FieldTypeString,
				Description: "Webhook URL",
				Required:    false,
			},
			"secret": {
				Name:        "secret",
				Type:        integration.FieldTypeSecret,
				Description: "Webhook secret",
				Required:    false,
				Sensitive:   true,
			},
			"events": {
				Name:        "events",
				Type:        integration.FieldTypeArray,
				Description: "Webhook events to subscribe to",
				Required:    false,
			},
		},
		OutputSpec: map[string]integration.FieldSpec{
			"id": {
				Name:        "id",
				Type:        integration.FieldTypeInteger,
				Description: "Resource ID",
			},
			"number": {
				Name:        "number",
				Type:        integration.FieldTypeInteger,
				Description: "Issue/PR number",
			},
			"html_url": {
				Name:        "html_url",
				Type:        integration.FieldTypeString,
				Description: "URL to the resource",
			},
			"state": {
				Name:        "state",
				Type:        integration.FieldTypeString,
				Description: "Resource state",
			},
		},
	}
}

// buildGitHubRetryConfig builds retry configuration for GitHub API.
func buildGitHubRetryConfig() *inthttp.RetryConfig {
	return &inthttp.RetryConfig{
		MaxRetries:   3,
		BaseDelay:    1 * time.Second,
		MaxDelay:     60 * time.Second,
		Multiplier:   2.0,
		JitterFactor: 0.1,
		ShouldRetry: func(err error, resp *inthttp.Response) bool {
			if err != nil {
				return inthttp.IsRetryableError(err)
			}
			if resp == nil {
				return false
			}
			// Retry on rate limiting (403 with rate limit header), 429, and server errors
			if resp.StatusCode == 403 {
				if resp.Headers.Get("X-RateLimit-Remaining") == "0" {
					return true
				}
			}
			return resp.StatusCode == 429 || resp.StatusCode >= 500
		},
	}
}
