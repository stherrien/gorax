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

func TestCreateIssueAction_Execute(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/rest/api/3/issue" {
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"id":   "10001",
				"key":  "TEST-1",
				"self": "https://example.atlassian.net/rest/api/3/issue/10001",
			})
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	action := NewCreateIssueAction(server.URL, "test@example.com", "test-token")

	config := CreateIssueConfig{
		Project:     "TEST",
		IssueType:   "Bug",
		Summary:     "Test issue",
		Description: "Test description",
	}

	result, err := action.Execute(context.Background(), map[string]interface{}{
		"project":     config.Project,
		"issue_type":  config.IssueType,
		"summary":     config.Summary,
		"description": config.Description,
	}, map[string]interface{}{})

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "10001", result["id"])
	assert.Equal(t, "TEST-1", result["key"])
}

func TestCreateIssueAction_Validate(t *testing.T) {
	action := NewCreateIssueAction("https://example.atlassian.net", "test@example.com", "test-token")

	tests := []struct {
		name    string
		config  map[string]interface{}
		wantErr bool
	}{
		{
			name: "valid config",
			config: map[string]interface{}{
				"project":     "TEST",
				"issue_type":  "Bug",
				"summary":     "Test issue",
				"description": "Test description",
			},
			wantErr: false,
		},
		{
			name: "missing project",
			config: map[string]interface{}{
				"issue_type":  "Bug",
				"summary":     "Test issue",
				"description": "Test description",
			},
			wantErr: true,
		},
		{
			name: "missing summary",
			config: map[string]interface{}{
				"project":     "TEST",
				"issue_type":  "Bug",
				"description": "Test description",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := action.Validate(tt.config)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestUpdateIssueAction_Execute(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/rest/api/3/issue/TEST-1" {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	action := NewUpdateIssueAction(server.URL, "test@example.com", "test-token")

	result, err := action.Execute(context.Background(), map[string]interface{}{
		"issue_key": "TEST-1",
		"fields": map[string]interface{}{
			"summary": "Updated summary",
		},
	}, map[string]interface{}{})

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, true, result["success"])
}

func TestAddCommentAction_Execute(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/rest/api/3/issue/TEST-1/comment" {
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"id":   "10100",
				"body": "Test comment",
			})
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	action := NewAddCommentAction(server.URL, "test@example.com", "test-token")

	result, err := action.Execute(context.Background(), map[string]interface{}{
		"issue_key": "TEST-1",
		"body":      "Test comment",
	}, map[string]interface{}{})

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "10100", result["id"])
}

func TestTransitionIssueAction_Execute(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/rest/api/3/issue/TEST-1/transitions" && r.Method == "GET" {
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
			w.WriteHeader(http.StatusNoContent)
			return
		}

		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	action := NewTransitionIssueAction(server.URL, "test@example.com", "test-token")

	result, err := action.Execute(context.Background(), map[string]interface{}{
		"issue_key":       "TEST-1",
		"transition_name": "Done",
	}, map[string]interface{}{})

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, true, result["success"])
}

func TestSearchIssuesAction_Execute(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/rest/api/3/search" {
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
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	action := NewSearchIssuesAction(server.URL, "test@example.com", "test-token")

	result, err := action.Execute(context.Background(), map[string]interface{}{
		"jql":         "project = TEST",
		"max_results": 50,
		"start_at":    0,
	}, map[string]interface{}{})

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, float64(2), result["total"])
	assert.NotNil(t, result["issues"])
}
