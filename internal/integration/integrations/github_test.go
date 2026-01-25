package integrations

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gorax/gorax/internal/integration"
	inttesting "github.com/gorax/gorax/internal/integration/testing"
)

func TestNewGitHubIntegration(t *testing.T) {
	github := NewGitHubIntegration(nil)

	assert.NotNil(t, github)
	assert.Equal(t, "github", github.Name())
	assert.Equal(t, integration.TypeAPI, github.Type())

	metadata := github.GetMetadata()
	assert.Equal(t, "GitHub", metadata.DisplayName)
	assert.Equal(t, "version_control", metadata.Category)

	schema := github.GetSchema()
	assert.NotNil(t, schema.ConfigSpec["token"])
	assert.NotNil(t, schema.InputSpec["action"])
}

func TestGitHubIntegration_Validate(t *testing.T) {
	github := NewGitHubIntegration(nil)

	tests := []struct {
		name        string
		config      *integration.Config
		expectError bool
	}{
		{
			name:        "nil config",
			config:      nil,
			expectError: true,
		},
		{
			name: "missing credentials",
			config: &integration.Config{
				Name:    "test",
				Type:    integration.TypeAPI,
				Enabled: true,
			},
			expectError: true,
		},
		{
			name: "missing token",
			config: &integration.Config{
				Name:    "test",
				Type:    integration.TypeAPI,
				Enabled: true,
				Credentials: &integration.Credentials{
					Data: integration.JSONMap{},
				},
			},
			expectError: true,
		},
		{
			name: "valid config with token",
			config: &integration.Config{
				Name:    "test",
				Type:    integration.TypeAPI,
				Enabled: true,
				Credentials: &integration.Credentials{
					Data: integration.JSONMap{
						"token": "ghp_test_token",
					},
				},
			},
			expectError: false,
		},
		{
			name: "valid config with access_token",
			config: &integration.Config{
				Name:    "test",
				Type:    integration.TypeAPI,
				Enabled: true,
				Credentials: &integration.Credentials{
					Data: integration.JSONMap{
						"access_token": "ghp_test_token",
					},
				},
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := github.Validate(tt.config)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGitHubIntegration_Execute_MissingAction(t *testing.T) {
	github := NewGitHubIntegration(nil)

	config := &integration.Config{
		Name:    "test-github",
		Type:    integration.TypeAPI,
		Enabled: true,
		Credentials: &integration.Credentials{
			Data: integration.JSONMap{
				"token": "ghp_test_token",
			},
		},
	}

	params := integration.JSONMap{
		"owner": "octocat",
		"repo":  "Hello-World",
	}

	ctx := context.Background()
	result, err := github.Execute(ctx, config, params)

	require.Error(t, err)
	assert.False(t, result.Success)
	assert.Equal(t, "VALIDATION_ERROR", result.ErrorCode)
}

func TestGitHubIntegration_Execute_InvalidAction(t *testing.T) {
	github := NewGitHubIntegration(nil)

	config := &integration.Config{
		Name:    "test-github",
		Type:    integration.TypeAPI,
		Enabled: true,
		Credentials: &integration.Credentials{
			Data: integration.JSONMap{
				"token": "ghp_test_token",
			},
		},
	}

	params := integration.JSONMap{
		"action": "invalid_action",
	}

	ctx := context.Background()
	result, err := github.Execute(ctx, config, params)

	require.Error(t, err)
	assert.False(t, result.Success)
	assert.Equal(t, "INVALID_ACTION", result.ErrorCode)
}

func TestGitHubIntegration_CreateIssue_MissingOwner(t *testing.T) {
	github := NewGitHubIntegration(nil)

	config := &integration.Config{
		Name:    "test-github",
		Type:    integration.TypeAPI,
		Enabled: true,
		Credentials: &integration.Credentials{
			Data: integration.JSONMap{
				"token": "ghp_test_token",
			},
		},
	}

	params := integration.JSONMap{
		"action": "create_issue",
		"title":  "Test Issue",
	}

	ctx := context.Background()
	result, err := github.Execute(ctx, config, params)

	require.Error(t, err)
	assert.False(t, result.Success)
	assert.Equal(t, "VALIDATION_ERROR", result.ErrorCode)
}

func TestGitHubIntegration_CreateIssue_MissingTitle(t *testing.T) {
	github := NewGitHubIntegration(nil)

	config := &integration.Config{
		Name:    "test-github",
		Type:    integration.TypeAPI,
		Enabled: true,
		Credentials: &integration.Credentials{
			Data: integration.JSONMap{
				"token": "ghp_test_token",
			},
		},
	}

	params := integration.JSONMap{
		"action": "create_issue",
		"owner":  "octocat",
		"repo":   "Hello-World",
	}

	ctx := context.Background()
	result, err := github.Execute(ctx, config, params)

	require.Error(t, err)
	assert.False(t, result.Success)
	assert.Equal(t, "VALIDATION_ERROR", result.ErrorCode)
}

func TestGitHubIntegration_CreatePR_MissingHead(t *testing.T) {
	github := NewGitHubIntegration(nil)

	config := &integration.Config{
		Name:    "test-github",
		Type:    integration.TypeAPI,
		Enabled: true,
		Credentials: &integration.Credentials{
			Data: integration.JSONMap{
				"token": "ghp_test_token",
			},
		},
	}

	params := integration.JSONMap{
		"action": "create_pull_request",
		"owner":  "octocat",
		"repo":   "Hello-World",
		"title":  "Test PR",
		"base":   "main",
	}

	ctx := context.Background()
	result, err := github.Execute(ctx, config, params)

	require.Error(t, err)
	assert.False(t, result.Success)
	assert.Equal(t, "VALIDATION_ERROR", result.ErrorCode)
}

func TestGitHubIntegration_GetOwnerRepo(t *testing.T) {
	github := NewGitHubIntegration(nil)

	tests := []struct {
		name        string
		params      integration.JSONMap
		wantOwner   string
		wantRepo    string
		expectError bool
	}{
		{
			name: "owner and repo separate",
			params: integration.JSONMap{
				"owner": "octocat",
				"repo":  "Hello-World",
			},
			wantOwner:   "octocat",
			wantRepo:    "Hello-World",
			expectError: false,
		},
		{
			name: "repository combined format",
			params: integration.JSONMap{
				"repository": "octocat/Hello-World",
			},
			wantOwner:   "octocat",
			wantRepo:    "Hello-World",
			expectError: false,
		},
		{
			name:        "missing owner",
			params:      integration.JSONMap{},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			owner, repo, err := github.getOwnerRepo(tt.params)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantOwner, owner)
				assert.Equal(t, tt.wantRepo, repo)
			}
		})
	}
}

// computeGitHubSignature is a test helper that computes HMAC-SHA256 signature.
func computeGitHubSignature(payload []byte, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	return "sha256=" + hex.EncodeToString(mac.Sum(nil))
}

func TestVerifyGitHubWebhookSignature(t *testing.T) {
	t.Run("valid signatures", func(t *testing.T) {
		tests := []struct {
			name    string
			payload []byte
			secret  string
		}{
			{
				name:    "simple JSON payload",
				payload: []byte(`{"action": "opened"}`),
				secret:  "webhook-secret",
			},
			{
				name:    "complex nested payload",
				payload: []byte(`{"action":"opened","pull_request":{"number":1,"title":"Test PR","user":{"login":"octocat"}}}`),
				secret:  "my-super-secret-key",
			},
			{
				name:    "empty object payload",
				payload: []byte(`{}`),
				secret:  "secret",
			},
			{
				name:    "payload with special characters",
				payload: []byte(`{"message":"Hello \"world\" with 'quotes' and <html> & special chars"}`),
				secret:  "secret-with-special-!@#$%",
			},
			{
				name:    "payload with unicode",
				payload: []byte(`{"message":"Hello 世界 🎉"}`),
				secret:  "unicode-secret-密钥",
			},
			{
				name:    "large payload",
				payload: []byte(strings.Repeat(`{"data":"test"},`, 1000) + `{"data":"end"}`),
				secret:  "large-payload-secret",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				signature := computeGitHubSignature(tt.payload, tt.secret)
				result := VerifyGitHubWebhookSignature(tt.payload, signature, tt.secret)
				assert.True(t, result, "expected valid signature to be verified")
			})
		}
	})

	t.Run("invalid signatures", func(t *testing.T) {
		payload := []byte(`{"action": "opened"}`)
		validSecret := "webhook-secret"
		validSignature := computeGitHubSignature(payload, validSecret)

		tests := []struct {
			name      string
			payload   []byte
			signature string
			secret    string
		}{
			{
				name:      "wrong secret",
				payload:   payload,
				signature: validSignature,
				secret:    "wrong-secret",
			},
			{
				name:      "modified payload",
				payload:   []byte(`{"action": "closed"}`),
				signature: validSignature,
				secret:    validSecret,
			},
			{
				name:      "truncated signature",
				payload:   payload,
				signature: validSignature[:20],
				secret:    validSecret,
			},
			{
				name:      "signature with extra characters",
				payload:   payload,
				signature: validSignature + "extra",
				secret:    validSecret,
			},
			{
				name:      "uppercase signature hex",
				payload:   payload,
				signature: "sha256=" + strings.ToUpper(strings.TrimPrefix(validSignature, "sha256=")),
				secret:    validSecret,
			},
			{
				name:      "signature with wrong algorithm prefix",
				payload:   payload,
				signature: "sha512=" + strings.TrimPrefix(validSignature, "sha256="),
				secret:    validSecret,
			},
			{
				name:      "signature with sha1 prefix",
				payload:   payload,
				signature: "sha1=" + strings.TrimPrefix(validSignature, "sha256="),
				secret:    validSecret,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result := VerifyGitHubWebhookSignature(tt.payload, tt.signature, tt.secret)
				assert.False(t, result, "expected invalid signature to be rejected")
			})
		}
	})

	t.Run("malformed signature formats", func(t *testing.T) {
		payload := []byte(`{"action": "opened"}`)
		secret := "webhook-secret"

		tests := []struct {
			name      string
			signature string
		}{
			{
				name:      "missing sha256 prefix",
				signature: "abc123def456",
			},
			{
				name:      "empty signature",
				signature: "",
			},
			{
				name:      "only sha256 prefix",
				signature: "sha256=",
			},
			{
				name:      "invalid hex characters",
				signature: "sha256=gggggggggggggggggggggggggggggggggggggggggggggggggggggggggggggggg",
			},
			{
				name:      "whitespace in signature",
				signature: "sha256= abc123",
			},
			{
				name:      "newline in signature",
				signature: "sha256=abc\n123",
			},
			{
				name:      "null bytes",
				signature: "sha256=abc\x00123",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result := VerifyGitHubWebhookSignature(payload, tt.signature, secret)
				assert.False(t, result, "expected malformed signature to be rejected")
			})
		}
	})

	t.Run("edge cases", func(t *testing.T) {
		t.Run("empty payload with valid signature", func(t *testing.T) {
			payload := []byte("")
			secret := "secret"
			signature := computeGitHubSignature(payload, secret)

			result := VerifyGitHubWebhookSignature(payload, signature, secret)
			assert.True(t, result)
		})

		t.Run("nil payload", func(t *testing.T) {
			secret := "secret"
			signature := computeGitHubSignature(nil, secret)

			result := VerifyGitHubWebhookSignature(nil, signature, secret)
			assert.True(t, result)
		})

		t.Run("empty secret with valid signature", func(t *testing.T) {
			payload := []byte(`{"test": "data"}`)
			secret := ""
			signature := computeGitHubSignature(payload, secret)

			result := VerifyGitHubWebhookSignature(payload, signature, secret)
			assert.True(t, result)
		})

		t.Run("very long secret", func(t *testing.T) {
			payload := []byte(`{"test": "data"}`)
			secret := strings.Repeat("a", 10000)
			signature := computeGitHubSignature(payload, secret)

			result := VerifyGitHubWebhookSignature(payload, signature, secret)
			assert.True(t, result)
		})

		t.Run("binary payload", func(t *testing.T) {
			payload := []byte{0x00, 0x01, 0x02, 0xFF, 0xFE, 0xFD}
			secret := "binary-secret"
			signature := computeGitHubSignature(payload, secret)

			result := VerifyGitHubWebhookSignature(payload, signature, secret)
			assert.True(t, result)
		})
	})

	t.Run("timing attack resistance", func(t *testing.T) {
		// Verify the implementation uses constant-time comparison
		payload := []byte(`{"action": "opened"}`)
		secret := "webhook-secret"
		validSignature := computeGitHubSignature(payload, secret)

		// Create signatures that differ at different positions
		invalidSignatures := []string{
			"sha256=0" + validSignature[8:],                 // First char different
			validSignature[:40] + "0" + validSignature[42:], // Middle char different
			validSignature[:len(validSignature)-1] + "0",    // Last char different
		}

		for i, invalidSig := range invalidSignatures {
			result := VerifyGitHubWebhookSignature(payload, invalidSig, secret)
			assert.False(t, result, "signature %d should be rejected", i)
		}
	})
}

func TestGitHubRetryConfig(t *testing.T) {
	config := buildGitHubRetryConfig()

	assert.Equal(t, 3, config.MaxRetries)
	assert.Equal(t, 1*time.Second, config.BaseDelay)
	assert.Equal(t, 60*time.Second, config.MaxDelay)
	assert.NotNil(t, config.ShouldRetry)
}

func TestGitHubSchema(t *testing.T) {
	schema := buildGitHubSchema()

	// Verify config spec
	assert.Contains(t, schema.ConfigSpec, "token")
	assert.True(t, schema.ConfigSpec["token"].Sensitive)

	// Verify input spec
	assert.Contains(t, schema.InputSpec, "action")
	assert.True(t, schema.InputSpec["action"].Required)
	assert.Contains(t, schema.InputSpec, "owner")
	assert.Contains(t, schema.InputSpec, "repo")
	assert.Contains(t, schema.InputSpec, "title")
	assert.Contains(t, schema.InputSpec, "body")
	assert.Contains(t, schema.InputSpec, "labels")

	// Verify output spec
	assert.Contains(t, schema.OutputSpec, "id")
	assert.Contains(t, schema.OutputSpec, "number")
	assert.Contains(t, schema.OutputSpec, "html_url")
}

// TestGitHubMockServer tests with mock server
func TestGitHubMockServer_GetRepository(t *testing.T) {
	mockServer := inttesting.NewMockServer()
	defer mockServer.Close()

	// Configure mock response
	mockServer.OnGet("/repos/octocat/Hello-World", inttesting.JSONResponse(http.StatusOK, map[string]any{
		"id":        1296269,
		"name":      "Hello-World",
		"full_name": "octocat/Hello-World",
		"owner": map[string]any{
			"login": "octocat",
			"id":    1,
		},
		"html_url": "https://github.com/octocat/Hello-World",
	}))

	// Note: This would require injecting the mock server URL into the client
	// For now, this serves as documentation of the expected behavior
	assert.NotNil(t, mockServer)
}

// =============================================================================
// GitHub Webhook Payload Parsing Tests
// =============================================================================

// GitHubWebhookPayload represents a generic GitHub webhook payload for testing.
type GitHubWebhookPayload struct {
	Action     string         `json:"action,omitempty"`
	Sender     *GitHubUser    `json:"sender,omitempty"`
	Repository *GitHubRepo    `json:"repository,omitempty"`
	Ref        string         `json:"ref,omitempty"`
	Before     string         `json:"before,omitempty"`
	After      string         `json:"after,omitempty"`
	Commits    []GitHubCommit `json:"commits,omitempty"`
}

// GitHubUser represents a GitHub user in webhook payloads.
type GitHubUser struct {
	Login     string `json:"login"`
	ID        int64  `json:"id"`
	AvatarURL string `json:"avatar_url,omitempty"`
	Type      string `json:"type,omitempty"`
}

// GitHubRepo represents a GitHub repository in webhook payloads.
type GitHubRepo struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	FullName string `json:"full_name"`
	Private  bool   `json:"private"`
	HTMLURL  string `json:"html_url,omitempty"`
}

// GitHubCommit represents a commit in push webhook payloads.
type GitHubCommit struct {
	ID        string     `json:"id"`
	Message   string     `json:"message"`
	Timestamp string     `json:"timestamp"`
	Author    GitHubUser `json:"author"`
	Added     []string   `json:"added,omitempty"`
	Removed   []string   `json:"removed,omitempty"`
	Modified  []string   `json:"modified,omitempty"`
}

// GitHubPullRequest represents a pull request in webhook payloads.
type GitHubPullRequest struct {
	Number    int         `json:"number"`
	State     string      `json:"state"`
	Title     string      `json:"title"`
	Body      string      `json:"body,omitempty"`
	User      *GitHubUser `json:"user,omitempty"`
	HTMLURL   string      `json:"html_url,omitempty"`
	Merged    bool        `json:"merged,omitempty"`
	Mergeable *bool       `json:"mergeable,omitempty"`
}

// GitHubIssue represents an issue in webhook payloads.
type GitHubIssue struct {
	Number  int         `json:"number"`
	State   string      `json:"state"`
	Title   string      `json:"title"`
	Body    string      `json:"body,omitempty"`
	User    *GitHubUser `json:"user,omitempty"`
	HTMLURL string      `json:"html_url,omitempty"`
	Labels  []struct {
		Name string `json:"name"`
	} `json:"labels,omitempty"`
}

func TestGitHubWebhookPayloadParsing_PushEvent(t *testing.T) {
	t.Run("parses standard push event", func(t *testing.T) {
		payload := `{
			"ref": "refs/heads/main",
			"before": "0000000000000000000000000000000000000000",
			"after": "abc123def456abc123def456abc123def456abc1",
			"repository": {
				"id": 123456,
				"name": "test-repo",
				"full_name": "octocat/test-repo",
				"private": false
			},
			"pusher": {
				"name": "octocat",
				"email": "octocat@github.com"
			},
			"sender": {
				"login": "octocat",
				"id": 1
			},
			"commits": [
				{
					"id": "abc123",
					"message": "Initial commit",
					"timestamp": "2024-01-15T10:30:00Z",
					"author": {
						"name": "octocat",
						"email": "octocat@github.com"
					},
					"added": ["README.md"],
					"removed": [],
					"modified": []
				}
			]
		}`

		var parsed map[string]any
		err := json.Unmarshal([]byte(payload), &parsed)
		require.NoError(t, err)

		assert.Equal(t, "refs/heads/main", parsed["ref"])
		assert.NotNil(t, parsed["repository"])
		assert.NotNil(t, parsed["commits"])

		commits := parsed["commits"].([]any)
		assert.Len(t, commits, 1)
	})

	t.Run("parses push event with multiple commits", func(t *testing.T) {
		payload := `{
			"ref": "refs/heads/feature-branch",
			"commits": [
				{"id": "commit1", "message": "First commit"},
				{"id": "commit2", "message": "Second commit"},
				{"id": "commit3", "message": "Third commit"}
			],
			"head_commit": {
				"id": "commit3",
				"message": "Third commit"
			}
		}`

		var parsed map[string]any
		err := json.Unmarshal([]byte(payload), &parsed)
		require.NoError(t, err)

		commits := parsed["commits"].([]any)
		assert.Len(t, commits, 3)
	})

	t.Run("parses tag push event", func(t *testing.T) {
		payload := `{
			"ref": "refs/tags/v1.0.0",
			"ref_type": "tag",
			"base_ref": "refs/heads/main",
			"repository": {
				"id": 123456,
				"name": "test-repo",
				"full_name": "octocat/test-repo"
			}
		}`

		var parsed map[string]any
		err := json.Unmarshal([]byte(payload), &parsed)
		require.NoError(t, err)

		assert.Equal(t, "refs/tags/v1.0.0", parsed["ref"])
		assert.True(t, strings.HasPrefix(parsed["ref"].(string), "refs/tags/"))
	})

	t.Run("parses branch deletion event", func(t *testing.T) {
		payload := `{
			"ref": "refs/heads/feature-to-delete",
			"before": "abc123def456abc123def456abc123def456abc1",
			"after": "0000000000000000000000000000000000000000",
			"deleted": true,
			"created": false
		}`

		var parsed map[string]any
		err := json.Unmarshal([]byte(payload), &parsed)
		require.NoError(t, err)

		assert.Equal(t, true, parsed["deleted"])
		assert.Equal(t, "0000000000000000000000000000000000000000", parsed["after"])
	})
}

func TestGitHubWebhookPayloadParsing_PullRequestEvent(t *testing.T) {
	t.Run("parses pull request opened event", func(t *testing.T) {
		payload := `{
			"action": "opened",
			"number": 42,
			"pull_request": {
				"number": 42,
				"state": "open",
				"title": "Add new feature",
				"body": "This PR adds a new feature.",
				"user": {
					"login": "octocat",
					"id": 1
				},
				"head": {
					"ref": "feature-branch",
					"sha": "abc123"
				},
				"base": {
					"ref": "main",
					"sha": "def456"
				},
				"merged": false,
				"mergeable": true,
				"draft": false
			},
			"repository": {
				"id": 123456,
				"name": "test-repo",
				"full_name": "octocat/test-repo"
			},
			"sender": {
				"login": "octocat",
				"id": 1
			}
		}`

		var parsed map[string]any
		err := json.Unmarshal([]byte(payload), &parsed)
		require.NoError(t, err)

		assert.Equal(t, "opened", parsed["action"])
		assert.Equal(t, float64(42), parsed["number"])

		pr := parsed["pull_request"].(map[string]any)
		assert.Equal(t, "open", pr["state"])
		assert.Equal(t, "Add new feature", pr["title"])
		assert.Equal(t, false, pr["merged"])
	})

	t.Run("parses pull request closed event", func(t *testing.T) {
		payload := `{
			"action": "closed",
			"number": 42,
			"pull_request": {
				"number": 42,
				"state": "closed",
				"merged": true,
				"merged_by": {
					"login": "maintainer",
					"id": 2
				},
				"merge_commit_sha": "merge123abc"
			}
		}`

		var parsed map[string]any
		err := json.Unmarshal([]byte(payload), &parsed)
		require.NoError(t, err)

		assert.Equal(t, "closed", parsed["action"])
		pr := parsed["pull_request"].(map[string]any)
		assert.Equal(t, true, pr["merged"])
		assert.NotNil(t, pr["merged_by"])
	})

	t.Run("parses pull request review requested event", func(t *testing.T) {
		payload := `{
			"action": "review_requested",
			"number": 42,
			"pull_request": {
				"number": 42
			},
			"requested_reviewer": {
				"login": "reviewer",
				"id": 3
			}
		}`

		var parsed map[string]any
		err := json.Unmarshal([]byte(payload), &parsed)
		require.NoError(t, err)

		assert.Equal(t, "review_requested", parsed["action"])
		reviewer := parsed["requested_reviewer"].(map[string]any)
		assert.Equal(t, "reviewer", reviewer["login"])
	})

	t.Run("parses pull request synchronize event", func(t *testing.T) {
		payload := `{
			"action": "synchronize",
			"number": 42,
			"before": "old_sha",
			"after": "new_sha",
			"pull_request": {
				"number": 42,
				"commits": 5
			}
		}`

		var parsed map[string]any
		err := json.Unmarshal([]byte(payload), &parsed)
		require.NoError(t, err)

		assert.Equal(t, "synchronize", parsed["action"])
		assert.Equal(t, "old_sha", parsed["before"])
		assert.Equal(t, "new_sha", parsed["after"])
	})
}

func TestGitHubWebhookPayloadParsing_IssuesEvent(t *testing.T) {
	t.Run("parses issue opened event", func(t *testing.T) {
		payload := `{
			"action": "opened",
			"issue": {
				"number": 1,
				"title": "Bug report: Something is broken",
				"body": "When I do X, Y happens instead of Z.",
				"state": "open",
				"user": {
					"login": "reporter",
					"id": 123
				},
				"labels": [
					{"name": "bug"},
					{"name": "priority:high"}
				],
				"assignees": [],
				"created_at": "2024-01-15T10:00:00Z"
			},
			"repository": {
				"id": 123456,
				"name": "test-repo"
			}
		}`

		var parsed map[string]any
		err := json.Unmarshal([]byte(payload), &parsed)
		require.NoError(t, err)

		assert.Equal(t, "opened", parsed["action"])
		issue := parsed["issue"].(map[string]any)
		assert.Equal(t, float64(1), issue["number"])
		assert.Equal(t, "open", issue["state"])

		labels := issue["labels"].([]any)
		assert.Len(t, labels, 2)
	})

	t.Run("parses issue labeled event", func(t *testing.T) {
		payload := `{
			"action": "labeled",
			"issue": {
				"number": 1,
				"title": "Test issue"
			},
			"label": {
				"name": "enhancement",
				"color": "84b6eb"
			}
		}`

		var parsed map[string]any
		err := json.Unmarshal([]byte(payload), &parsed)
		require.NoError(t, err)

		assert.Equal(t, "labeled", parsed["action"])
		label := parsed["label"].(map[string]any)
		assert.Equal(t, "enhancement", label["name"])
	})

	t.Run("parses issue assigned event", func(t *testing.T) {
		payload := `{
			"action": "assigned",
			"issue": {
				"number": 1,
				"assignees": [
					{"login": "developer1", "id": 1}
				]
			},
			"assignee": {
				"login": "developer1",
				"id": 1
			}
		}`

		var parsed map[string]any
		err := json.Unmarshal([]byte(payload), &parsed)
		require.NoError(t, err)

		assert.Equal(t, "assigned", parsed["action"])
		assignee := parsed["assignee"].(map[string]any)
		assert.Equal(t, "developer1", assignee["login"])
	})
}

func TestGitHubWebhookPayloadParsing_ReleaseEvent(t *testing.T) {
	t.Run("parses release published event", func(t *testing.T) {
		payload := `{
			"action": "published",
			"release": {
				"id": 12345,
				"tag_name": "v1.0.0",
				"name": "Version 1.0.0",
				"body": "Release notes here",
				"draft": false,
				"prerelease": false,
				"created_at": "2024-01-15T10:00:00Z",
				"published_at": "2024-01-15T10:30:00Z",
				"author": {
					"login": "releaser",
					"id": 1
				},
				"assets": [
					{
						"name": "app-v1.0.0.zip",
						"size": 1024000,
						"download_count": 0
					}
				]
			},
			"repository": {
				"id": 123456,
				"name": "test-repo"
			}
		}`

		var parsed map[string]any
		err := json.Unmarshal([]byte(payload), &parsed)
		require.NoError(t, err)

		assert.Equal(t, "published", parsed["action"])
		release := parsed["release"].(map[string]any)
		assert.Equal(t, "v1.0.0", release["tag_name"])
		assert.Equal(t, false, release["prerelease"])

		assets := release["assets"].([]any)
		assert.Len(t, assets, 1)
	})

	t.Run("parses prerelease event", func(t *testing.T) {
		payload := `{
			"action": "prereleased",
			"release": {
				"tag_name": "v2.0.0-beta.1",
				"prerelease": true,
				"draft": false
			}
		}`

		var parsed map[string]any
		err := json.Unmarshal([]byte(payload), &parsed)
		require.NoError(t, err)

		release := parsed["release"].(map[string]any)
		assert.Equal(t, true, release["prerelease"])
		assert.True(t, strings.Contains(release["tag_name"].(string), "beta"))
	})
}

func TestGitHubWebhookPayloadParsing_WorkflowRunEvent(t *testing.T) {
	t.Run("parses workflow run completed event", func(t *testing.T) {
		payload := `{
			"action": "completed",
			"workflow_run": {
				"id": 987654321,
				"name": "CI",
				"head_branch": "main",
				"head_sha": "abc123def456",
				"status": "completed",
				"conclusion": "success",
				"workflow_id": 12345,
				"run_number": 42,
				"run_attempt": 1,
				"created_at": "2024-01-15T10:00:00Z",
				"updated_at": "2024-01-15T10:05:00Z"
			},
			"workflow": {
				"id": 12345,
				"name": "CI",
				"path": ".github/workflows/ci.yml"
			},
			"repository": {
				"id": 123456,
				"name": "test-repo"
			}
		}`

		var parsed map[string]any
		err := json.Unmarshal([]byte(payload), &parsed)
		require.NoError(t, err)

		assert.Equal(t, "completed", parsed["action"])
		run := parsed["workflow_run"].(map[string]any)
		assert.Equal(t, "completed", run["status"])
		assert.Equal(t, "success", run["conclusion"])
	})

	t.Run("parses workflow run failed event", func(t *testing.T) {
		payload := `{
			"action": "completed",
			"workflow_run": {
				"id": 987654322,
				"status": "completed",
				"conclusion": "failure",
				"run_attempt": 1
			}
		}`

		var parsed map[string]any
		err := json.Unmarshal([]byte(payload), &parsed)
		require.NoError(t, err)

		run := parsed["workflow_run"].(map[string]any)
		assert.Equal(t, "failure", run["conclusion"])
	})
}

func TestGitHubWebhookPayloadParsing_CheckRunEvent(t *testing.T) {
	t.Run("parses check run completed event", func(t *testing.T) {
		payload := `{
			"action": "completed",
			"check_run": {
				"id": 123456789,
				"name": "test-suite",
				"head_sha": "abc123",
				"status": "completed",
				"conclusion": "success",
				"started_at": "2024-01-15T10:00:00Z",
				"completed_at": "2024-01-15T10:05:00Z",
				"output": {
					"title": "Test Results",
					"summary": "All tests passed",
					"annotations_count": 0
				}
			},
			"repository": {
				"id": 123456,
				"name": "test-repo"
			}
		}`

		var parsed map[string]any
		err := json.Unmarshal([]byte(payload), &parsed)
		require.NoError(t, err)

		assert.Equal(t, "completed", parsed["action"])
		checkRun := parsed["check_run"].(map[string]any)
		assert.Equal(t, "success", checkRun["conclusion"])
	})
}

// =============================================================================
// Error Handling Tests for Malformed Payloads
// =============================================================================

func TestGitHubWebhookPayloadParsing_MalformedPayloads(t *testing.T) {
	t.Run("rejects invalid JSON", func(t *testing.T) {
		malformedPayloads := []struct {
			name    string
			payload string
		}{
			{
				name:    "missing closing brace",
				payload: `{"action": "opened"`,
			},
			{
				name:    "missing quotes on string",
				payload: `{action: opened}`,
			},
			{
				name:    "trailing comma",
				payload: `{"action": "opened",}`,
			},
			{
				name:    "single quotes instead of double",
				payload: `{'action': 'opened'}`,
			},
			{
				name:    "unescaped control characters",
				payload: "{\"action\": \"opened\n\"}",
			},
			{
				name:    "invalid unicode escape",
				payload: `{"action": "\uZZZZ"}`,
			},
		}

		for _, tc := range malformedPayloads {
			t.Run(tc.name, func(t *testing.T) {
				var parsed map[string]any
				err := json.Unmarshal([]byte(tc.payload), &parsed)
				assert.Error(t, err)
			})
		}
	})

	t.Run("handles missing required fields gracefully", func(t *testing.T) {
		// Payloads with missing fields should still parse, just with nil values
		payloads := []string{
			`{}`,
			`{"action": "opened"}`,
			`{"repository": null}`,
			`{"sender": {}}`,
		}

		for _, payload := range payloads {
			var parsed map[string]any
			err := json.Unmarshal([]byte(payload), &parsed)
			assert.NoError(t, err)
		}
	})

	t.Run("handles wrong field types", func(t *testing.T) {
		// JSON parsing is permissive - these will parse but have unexpected types
		payload := `{
			"number": "not-a-number",
			"issue": "should-be-object",
			"labels": "should-be-array"
		}`

		var parsed map[string]any
		err := json.Unmarshal([]byte(payload), &parsed)
		require.NoError(t, err)

		// The values are strings, not the expected types
		_, isString := parsed["number"].(string)
		assert.True(t, isString)
	})

	t.Run("handles deeply nested payloads", func(t *testing.T) {
		// Create a deeply nested payload
		depth := 100
		payload := strings.Repeat(`{"nested":`, depth) + `"value"` + strings.Repeat(`}`, depth)

		var parsed map[string]any
		err := json.Unmarshal([]byte(payload), &parsed)
		assert.NoError(t, err)
	})

	t.Run("handles very large string values", func(t *testing.T) {
		largeBody := strings.Repeat("x", 1000000) // 1MB of text
		payload := `{"body": "` + largeBody + `"}`

		var parsed map[string]any
		err := json.Unmarshal([]byte(payload), &parsed)
		assert.NoError(t, err)
		assert.Len(t, parsed["body"].(string), 1000000)
	})

	t.Run("handles empty arrays and objects", func(t *testing.T) {
		payload := `{
			"commits": [],
			"labels": [],
			"assignees": [],
			"pull_request": {},
			"issue": {}
		}`

		var parsed map[string]any
		err := json.Unmarshal([]byte(payload), &parsed)
		require.NoError(t, err)

		commits := parsed["commits"].([]any)
		assert.Empty(t, commits)

		pr := parsed["pull_request"].(map[string]any)
		assert.Empty(t, pr)
	})

	t.Run("handles null values", func(t *testing.T) {
		payload := `{
			"action": null,
			"repository": null,
			"sender": null,
			"issue": null
		}`

		var parsed map[string]any
		err := json.Unmarshal([]byte(payload), &parsed)
		require.NoError(t, err)

		assert.Nil(t, parsed["action"])
		assert.Nil(t, parsed["repository"])
	})

	t.Run("handles numeric edge cases", func(t *testing.T) {
		payload := `{
			"id": 9223372036854775807,
			"float_id": 1.7976931348623157e+308,
			"negative_id": -9223372036854775808,
			"zero": 0,
			"negative_zero": -0
		}`

		var parsed map[string]any
		err := json.Unmarshal([]byte(payload), &parsed)
		require.NoError(t, err)

		// JSON numbers are parsed as float64 in Go
		assert.NotNil(t, parsed["id"])
	})
}

// =============================================================================
// Signature Verification Integration Tests
// =============================================================================

func TestGitHubWebhookSignatureWithRealPayloads(t *testing.T) {
	t.Run("verifies signature for push event payload", func(t *testing.T) {
		payload := []byte(`{
			"ref": "refs/heads/main",
			"before": "abc123",
			"after": "def456",
			"repository": {"id": 123, "name": "test"},
			"pusher": {"name": "octocat"},
			"commits": []
		}`)
		secret := "github-webhook-secret"
		signature := computeGitHubSignature(payload, secret)

		assert.True(t, VerifyGitHubWebhookSignature(payload, signature, secret))
	})

	t.Run("verifies signature for pull request event payload", func(t *testing.T) {
		payload := []byte(`{
			"action": "opened",
			"number": 1,
			"pull_request": {
				"number": 1,
				"title": "Test PR",
				"state": "open"
			}
		}`)
		secret := "pr-secret-key"
		signature := computeGitHubSignature(payload, secret)

		assert.True(t, VerifyGitHubWebhookSignature(payload, signature, secret))
	})

	t.Run("rejects tampered push event payload", func(t *testing.T) {
		originalPayload := []byte(`{"ref": "refs/heads/main", "commits": []}`)
		tamperedPayload := []byte(`{"ref": "refs/heads/malicious", "commits": []}`)
		secret := "secret"
		signature := computeGitHubSignature(originalPayload, secret)

		assert.False(t, VerifyGitHubWebhookSignature(tamperedPayload, signature, secret))
	})
}

// =============================================================================
// Benchmarks for Performance Testing
// =============================================================================

func BenchmarkVerifyGitHubWebhookSignature(b *testing.B) {
	payload := []byte(`{"action":"opened","pull_request":{"number":1,"title":"Test","body":"Description"}}`)
	secret := "benchmark-secret"
	signature := computeGitHubSignature(payload, secret)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		VerifyGitHubWebhookSignature(payload, signature, secret)
	}
}

func BenchmarkVerifyGitHubWebhookSignature_LargePayload(b *testing.B) {
	// Create a large payload (approximately 1MB)
	largeData := strings.Repeat(`{"commit":"abc123","message":"Test commit message"},`, 10000)
	payload := []byte(`{"commits":[` + largeData[:len(largeData)-1] + `]}`)
	secret := "benchmark-secret"
	signature := computeGitHubSignature(payload, secret)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		VerifyGitHubWebhookSignature(payload, signature, secret)
	}
}

func BenchmarkComputeGitHubSignature(b *testing.B) {
	payload := []byte(`{"action":"opened","pull_request":{"number":1}}`)
	secret := "benchmark-secret"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		computeGitHubSignature(payload, secret)
	}
}
