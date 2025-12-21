package github

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewClient(t *testing.T) {
	client, err := NewClient("test-token")

	require.NoError(t, err)
	assert.NotNil(t, client)
	assert.Equal(t, "test-token", client.token)
}

func TestNewClient_EmptyToken(t *testing.T) {
	_, err := NewClient("")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "token is required")
}

func TestClient_CreateIssue(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/repos/owner/repo/issues", r.URL.Path)
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))

		var req map[string]interface{}
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)

		assert.Equal(t, "Test Issue", req["title"])
		assert.Equal(t, "Test body", req["body"])

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(Issue{
			Number: 123,
			Title:  "Test Issue",
			Body:   "Test body",
			State:  "open",
			URL:    "https://github.com/owner/repo/issues/123",
		})
	}))
	defer server.Close()

	client, err := NewClient("test-token")
	require.NoError(t, err)
	client.baseURL = server.URL

	issue, err := client.CreateIssue(context.Background(), "owner", "repo", "Test Issue", "Test body", nil)

	assert.NoError(t, err)
	require.NotNil(t, issue)
	assert.Equal(t, 123, issue.Number)
	assert.Equal(t, "Test Issue", issue.Title)
}

func TestClient_CreatePRComment(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/repos/owner/repo/issues/123/comments", r.URL.Path)
		assert.Equal(t, "POST", r.Method)

		var req map[string]interface{}
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)

		assert.Equal(t, "Test comment", req["body"])

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(Comment{
			ID:   456,
			Body: "Test comment",
			URL:  "https://github.com/owner/repo/issues/123#issuecomment-456",
		})
	}))
	defer server.Close()

	client, err := NewClient("test-token")
	require.NoError(t, err)
	client.baseURL = server.URL

	comment, err := client.CreatePRComment(context.Background(), "owner", "repo", 123, "Test comment")

	assert.NoError(t, err)
	require.NotNil(t, comment)
	assert.Equal(t, 456, comment.ID)
	assert.Equal(t, "Test comment", comment.Body)
}

func TestClient_AddLabels(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/repos/owner/repo/issues/123/labels", r.URL.Path)
		assert.Equal(t, "POST", r.Method)

		var req map[string]interface{}
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)

		labels := req["labels"].([]interface{})
		assert.Len(t, labels, 2)
		assert.Equal(t, "bug", labels[0])
		assert.Equal(t, "enhancement", labels[1])

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client, err := NewClient("test-token")
	require.NoError(t, err)
	client.baseURL = server.URL

	err = client.AddLabels(context.Background(), "owner", "repo", 123, []string{"bug", "enhancement"})

	assert.NoError(t, err)
}
