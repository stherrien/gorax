package jira

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
	client, err := NewClient("https://example.atlassian.net", "user@example.com", "api-token")

	require.NoError(t, err)
	assert.NotNil(t, client)
	assert.Equal(t, "https://example.atlassian.net", client.baseURL)
	assert.Equal(t, "user@example.com", client.email)
	assert.Equal(t, "api-token", client.apiToken)
}

func TestNewClient_EmptyBaseURL(t *testing.T) {
	_, err := NewClient("", "user@example.com", "api-token")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "base URL is required")
}

func TestNewClient_EmptyCredentials(t *testing.T) {
	_, err := NewClient("https://example.atlassian.net", "", "api-token")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "email is required")

	_, err = NewClient("https://example.atlassian.net", "user@example.com", "")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "API token is required")
}

func TestClient_Authenticate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/rest/api/3/myself", r.URL.Path)

		// Check Basic Auth
		user, pass, ok := r.BasicAuth()
		assert.True(t, ok)
		assert.Equal(t, "user@example.com", user)
		assert.Equal(t, "api-token", pass)

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"accountId": "123456",
			"emailAddress": "user@example.com",
		})
	}))
	defer server.Close()

	client, err := NewClient(server.URL, "user@example.com", "api-token")
	require.NoError(t, err)

	err = client.Authenticate(context.Background())
	assert.NoError(t, err)
}

func TestClient_AuthenticateFailed(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"errorMessages": []string{"Invalid credentials"},
		})
	}))
	defer server.Close()

	client, err := NewClient(server.URL, "user@example.com", "invalid-token")
	require.NoError(t, err)

	err = client.Authenticate(context.Background())
	assert.Error(t, err)
}

func TestClient_CreateIssue(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/rest/api/3/issue", r.URL.Path)
		assert.Equal(t, "POST", r.Method)

		var req map[string]interface{}
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)

		fields := req["fields"].(map[string]interface{})
		assert.Equal(t, "Test Issue", fields["summary"])
		assert.Equal(t, "Bug", fields["issuetype"].(map[string]interface{})["name"])

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":  "10001",
			"key": "TEST-1",
			"self": "https://example.atlassian.net/rest/api/3/issue/10001",
		})
	}))
	defer server.Close()

	client, err := NewClient(server.URL, "user@example.com", "api-token")
	require.NoError(t, err)

	req := CreateIssueRequest{
		Project:     "TEST",
		IssueType:   "Bug",
		Summary:     "Test Issue",
		Description: "Test description",
	}

	issue, err := client.CreateIssue(context.Background(), req)

	assert.NoError(t, err)
	require.NotNil(t, issue)
	assert.Equal(t, "10001", issue.ID)
	assert.Equal(t, "TEST-1", issue.Key)
}

func TestClient_UpdateIssue(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/rest/api/3/issue/TEST-1", r.URL.Path)
		assert.Equal(t, "PUT", r.Method)

		var req map[string]interface{}
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)

		fields := req["fields"].(map[string]interface{})
		assert.Equal(t, "Updated Summary", fields["summary"])

		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client, err := NewClient(server.URL, "user@example.com", "api-token")
	require.NoError(t, err)

	err = client.UpdateIssue(context.Background(), "TEST-1", map[string]interface{}{
		"summary": "Updated Summary",
	})

	assert.NoError(t, err)
}

func TestClient_AddComment(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/rest/api/3/issue/TEST-1/comment", r.URL.Path)
		assert.Equal(t, "POST", r.Method)

		var req map[string]interface{}
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)

		assert.Equal(t, "Test comment", req["body"])

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":   "10100",
			"body": "Test comment",
		})
	}))
	defer server.Close()

	client, err := NewClient(server.URL, "user@example.com", "api-token")
	require.NoError(t, err)

	comment, err := client.AddComment(context.Background(), "TEST-1", "Test comment")

	assert.NoError(t, err)
	assert.NotNil(t, comment)
	assert.Equal(t, "10100", comment.ID)
}

func TestClient_TransitionIssue(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/rest/api/3/issue/TEST-1/transitions" && r.Method == "GET" {
			// Return available transitions
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"transitions": []map[string]interface{}{
					{
						"id":   "21",
						"name": "Done",
					},
				},
			})
			return
		}

		if r.URL.Path == "/rest/api/3/issue/TEST-1/transitions" && r.Method == "POST" {
			var req map[string]interface{}
			err := json.NewDecoder(r.Body).Decode(&req)
			require.NoError(t, err)

			transition := req["transition"].(map[string]interface{})
			assert.Equal(t, "21", transition["id"])

			w.WriteHeader(http.StatusNoContent)
			return
		}

		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client, err := NewClient(server.URL, "user@example.com", "api-token")
	require.NoError(t, err)

	err = client.TransitionIssue(context.Background(), "TEST-1", "Done")

	assert.NoError(t, err)
}

func TestClient_SearchIssues(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/rest/api/3/search", r.URL.Path)
		assert.Equal(t, "POST", r.Method)

		var req map[string]interface{}
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)

		assert.Equal(t, "project = TEST AND status = Open", req["jql"])

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"total": 2,
			"issues": []map[string]interface{}{
				{
					"id":  "10001",
					"key": "TEST-1",
				},
				{
					"id":  "10002",
					"key": "TEST-2",
				},
			},
		})
	}))
	defer server.Close()

	client, err := NewClient(server.URL, "user@example.com", "api-token")
	require.NoError(t, err)

	result, err := client.SearchIssues(context.Background(), "project = TEST AND status = Open", 50, 0)

	assert.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, 2, result.Total)
	assert.Len(t, result.Issues, 2)
	assert.Equal(t, "TEST-1", result.Issues[0].Key)
}
