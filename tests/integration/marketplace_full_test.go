package integration

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMarketplace_FullWorkflow tests the complete marketplace workflow
func TestMarketplace_FullWorkflow(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	ts := SetupTestServer(t)
	defer ts.Cleanup()

	// Create test tenant and user
	tenantID := ts.CreateTestTenant(t, "Test Tenant")
	userID := ts.CreateTestUser(t, tenantID, "user@example.com", "user")

	headers := DefaultTestHeaders(tenantID)

	// Step 1: Get categories
	t.Run("GetCategories", func(t *testing.T) {
		resp := ts.MakeRequest(t, http.MethodGet, "/api/v1/marketplace/categories", nil, headers)
		AssertStatusCode(t, resp, http.StatusOK)

		var categories []map[string]interface{}
		ParseJSONResponse(t, resp, &categories)
		assert.NotEmpty(t, categories, "categories should not be empty")
		t.Logf("✓ Retrieved %d categories", len(categories))
	})

	// Step 2: Publish a template
	var templateID string
	t.Run("PublishTemplate", func(t *testing.T) {
		publishPayload := map[string]interface{}{
			"name":        "Test Workflow Template",
			"description": "A comprehensive test workflow template for integration testing purposes",
			"category":    "automation",
			"version":     "1.0.0",
			"tags":        []string{"test", "automation", "webhook"},
			"definition": map[string]interface{}{
				"nodes": []map[string]interface{}{
					{
						"id":   "1",
						"type": "trigger",
						"data": map[string]interface{}{
							"nodeType": "webhook",
							"name":     "Webhook Trigger",
						},
					},
					{
						"id":   "2",
						"type": "action",
						"data": map[string]interface{}{
							"nodeType": "http",
							"name":     "HTTP Request",
						},
					},
				},
				"edges": []map[string]interface{}{
					{
						"id":     "e1",
						"source": "1",
						"target": "2",
					},
				},
			},
		}

		resp := ts.MakeRequest(t, http.MethodPost, "/api/v1/marketplace/templates", publishPayload, headers)
		AssertStatusCode(t, resp, http.StatusCreated)

		var template map[string]interface{}
		ParseJSONResponse(t, resp, &template)

		templateID = template["id"].(string)
		assert.NotEmpty(t, templateID)
		assert.Equal(t, "Test Workflow Template", template["name"])
		assert.Equal(t, userID, template["author_id"])
		assert.Equal(t, 0.0, template["average_rating"])
		t.Logf("✓ Template published: %s", templateID)
	})

	// Step 3: Search for the template
	t.Run("SearchTemplates", func(t *testing.T) {
		resp := ts.MakeRequest(t, http.MethodGet, "/api/v1/marketplace/templates?category=automation", nil, headers)
		AssertStatusCode(t, resp, http.StatusOK)

		var result map[string]interface{}
		ParseJSONResponse(t, resp, &result)

		templates := result["templates"].([]interface{})
		assert.NotEmpty(t, templates)

		found := false
		for _, tmpl := range templates {
			t := tmpl.(map[string]interface{})
			if t["id"].(string) == templateID {
				found = true
				break
			}
		}
		assert.True(t, found, "published template should be in search results")
		t.Logf("✓ Template found in search")
	})

	// Step 4: Get template details
	t.Run("GetTemplateDetails", func(t *testing.T) {
		resp := ts.MakeRequest(t, http.MethodGet, "/api/v1/marketplace/templates/"+templateID, nil, headers)
		AssertStatusCode(t, resp, http.StatusOK)

		var template map[string]interface{}
		ParseJSONResponse(t, resp, &template)

		assert.Equal(t, templateID, template["id"])
		assert.Equal(t, "Test Workflow Template", template["name"])
		assert.NotNil(t, template["definition"])
		t.Logf("✓ Template details retrieved")
	})

	// Step 5: Rate the template
	t.Run("RateTemplate", func(t *testing.T) {
		ratingPayload := map[string]interface{}{
			"rating":  5,
			"comment": "Excellent template! Works perfectly for my use case.",
		}

		resp := ts.MakeRequest(t, http.MethodPost, "/api/v1/marketplace/templates/"+templateID+"/rate", ratingPayload, headers)
		AssertStatusCode(t, resp, http.StatusCreated)

		var review map[string]interface{}
		ParseJSONResponse(t, resp, &review)

		assert.Equal(t, float64(5), review["rating"])
		assert.Equal(t, "Excellent template! Works perfectly for my use case.", review["comment"])
		t.Logf("✓ Template rated: 5 stars")

		// Verify rating is reflected in template
		resp = ts.MakeRequest(t, http.MethodGet, "/api/v1/marketplace/templates/"+templateID, nil, headers)
		AssertStatusCode(t, resp, http.StatusOK)

		var updatedTemplate map[string]interface{}
		ParseJSONResponse(t, resp, &updatedTemplate)

		assert.Equal(t, 5.0, updatedTemplate["average_rating"])
		assert.Equal(t, float64(1), updatedTemplate["total_ratings"])
		t.Logf("✓ Rating reflected in template")
	})

	// Step 6: Get reviews
	t.Run("GetReviews", func(t *testing.T) {
		resp := ts.MakeRequest(t, http.MethodGet, "/api/v1/marketplace/templates/"+templateID+"/reviews", nil, headers)
		AssertStatusCode(t, resp, http.StatusOK)

		var reviews []map[string]interface{}
		ParseJSONResponse(t, resp, &reviews)

		assert.Len(t, reviews, 1)
		assert.Equal(t, float64(5), reviews[0]["rating"])
		t.Logf("✓ Retrieved %d reviews", len(reviews))
	})

	// Step 7: Install the template
	var workflowID string
	t.Run("InstallTemplate", func(t *testing.T) {
		installPayload := map[string]interface{}{
			"workflow_name": "My Installed Workflow",
		}

		resp := ts.MakeRequest(t, http.MethodPost, "/api/v1/marketplace/templates/"+templateID+"/install", installPayload, headers)
		AssertStatusCode(t, resp, http.StatusCreated)

		var result map[string]interface{}
		ParseJSONResponse(t, resp, &result)

		workflowID = result["workflow_id"].(string)
		assert.NotEmpty(t, workflowID)
		assert.Equal(t, "My Installed Workflow", result["workflow_name"])
		t.Logf("✓ Template installed as workflow: %s", workflowID)

		// Verify download count incremented
		resp = ts.MakeRequest(t, http.MethodGet, "/api/v1/marketplace/templates/"+templateID, nil, headers)
		AssertStatusCode(t, resp, http.StatusOK)

		var updatedTemplate map[string]interface{}
		ParseJSONResponse(t, resp, &updatedTemplate)

		assert.Equal(t, float64(1), updatedTemplate["download_count"])
		t.Logf("✓ Download count incremented")
	})

	// Step 8: Try to install again (should fail)
	t.Run("PreventDuplicateInstallation", func(t *testing.T) {
		installPayload := map[string]interface{}{
			"workflow_name": "Duplicate Installation",
		}

		resp := ts.MakeRequest(t, http.MethodPost, "/api/v1/marketplace/templates/"+templateID+"/install", installPayload, headers)
		AssertStatusCode(t, resp, http.StatusConflict)
		t.Logf("✓ Duplicate installation prevented")
	})

	// Step 9: Get popular templates
	t.Run("GetPopularTemplates", func(t *testing.T) {
		resp := ts.MakeRequest(t, http.MethodGet, "/api/v1/marketplace/popular?limit=10", nil, headers)
		AssertStatusCode(t, resp, http.StatusOK)

		var templates []map[string]interface{}
		ParseJSONResponse(t, resp, &templates)

		assert.NotEmpty(t, templates)
		// Our template should be in the list
		found := false
		for _, tmpl := range templates {
			if tmpl["id"].(string) == templateID {
				found = true
				assert.Equal(t, float64(1), tmpl["download_count"])
				break
			}
		}
		assert.True(t, found, "installed template should be in popular list")
		t.Logf("✓ Retrieved %d popular templates", len(templates))
	})

	// Step 10: Get trending templates
	t.Run("GetTrendingTemplates", func(t *testing.T) {
		resp := ts.MakeRequest(t, http.MethodGet, "/api/v1/marketplace/trending?limit=10", nil, headers)
		AssertStatusCode(t, resp, http.StatusOK)

		var templates []map[string]interface{}
		ParseJSONResponse(t, resp, &templates)

		assert.NotEmpty(t, templates)
		t.Logf("✓ Retrieved %d trending templates", len(templates))
	})
}

// TestMarketplace_ReviewManagement tests review operations
func TestMarketplace_ReviewManagement(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	ts := SetupTestServer(t)
	defer ts.Cleanup()

	// Create test tenant and users
	tenantID := ts.CreateTestTenant(t, "Test Tenant")
	user1ID := ts.CreateTestUser(t, tenantID, "user1@example.com", "user")
	user2ID := ts.CreateTestUser(t, tenantID, "user2@example.com", "user")

	headers1 := DefaultTestHeaders(tenantID)
	headers2 := DefaultTestHeaders(tenantID)
	headers2["X-User-ID"] = user2ID

	// Publish a template
	publishPayload := map[string]interface{}{
		"name":        "Review Test Template",
		"description": "Template for testing review functionality",
		"category":    "integration",
		"version":     "1.0.0",
		"definition": map[string]interface{}{
			"nodes": []map[string]interface{}{
				{"id": "1", "type": "trigger"},
			},
			"edges": []map[string]interface{}{},
		},
	}

	resp := ts.MakeRequest(t, http.MethodPost, "/api/v1/marketplace/templates", publishPayload, headers1)
	AssertStatusCode(t, resp, http.StatusCreated)

	var template map[string]interface{}
	ParseJSONResponse(t, resp, &template)
	templateID := template["id"].(string)

	// Add multiple reviews
	t.Run("AddMultipleReviews", func(t *testing.T) {
		reviews := []struct {
			rating  int
			comment string
			headers map[string]string
		}{
			{5, "Excellent!", headers1},
			{4, "Very good", headers2},
		}

		for i, rev := range reviews {
			ratingPayload := map[string]interface{}{
				"rating":  rev.rating,
				"comment": rev.comment,
			}

			resp := ts.MakeRequest(t, http.MethodPost, "/api/v1/marketplace/templates/"+templateID+"/rate", ratingPayload, rev.headers)
			AssertStatusCode(t, resp, http.StatusCreated)
			t.Logf("✓ Review %d added", i+1)
		}

		// Verify average rating: (5 + 4) / 2 = 4.5
		resp := ts.MakeRequest(t, http.MethodGet, "/api/v1/marketplace/templates/"+templateID, nil, headers1)
		AssertStatusCode(t, resp, http.StatusOK)

		var updatedTemplate map[string]interface{}
		ParseJSONResponse(t, resp, &updatedTemplate)

		assert.Equal(t, 4.5, updatedTemplate["average_rating"])
		assert.Equal(t, float64(2), updatedTemplate["total_ratings"])
		t.Logf("✓ Average rating: %.1f", updatedTemplate["average_rating"])
	})

	// Update a review
	t.Run("UpdateReview", func(t *testing.T) {
		updatePayload := map[string]interface{}{
			"rating":  3,
			"comment": "Changed my mind, it's just okay",
		}

		resp := ts.MakeRequest(t, http.MethodPost, "/api/v1/marketplace/templates/"+templateID+"/rate", updatePayload, headers1)
		AssertStatusCode(t, resp, http.StatusOK)

		// Verify new average: (3 + 4) / 2 = 3.5
		resp = ts.MakeRequest(t, http.MethodGet, "/api/v1/marketplace/templates/"+templateID, nil, headers1)
		AssertStatusCode(t, resp, http.StatusOK)

		var updatedTemplate map[string]interface{}
		ParseJSONResponse(t, resp, &updatedTemplate)

		assert.Equal(t, 3.5, updatedTemplate["average_rating"])
		assert.Equal(t, float64(2), updatedTemplate["total_ratings"])
		t.Logf("✓ Review updated, new average: %.1f", updatedTemplate["average_rating"])
	})
}

// TestMarketplace_SearchAndFiltering tests search and filtering capabilities
func TestMarketplace_SearchAndFiltering(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	ts := SetupTestServer(t)
	defer ts.Cleanup()

	tenantID := ts.CreateTestTenant(t, "Test Tenant")
	userID := ts.CreateTestUser(t, tenantID, "user@example.com", "user")
	headers := DefaultTestHeaders(tenantID)

	// Publish templates with different categories and tags
	templates := []struct {
		name     string
		category string
		tags     []string
	}{
		{"Webhook Handler", "integration", []string{"webhook", "http"}},
		{"Data Processor", "automation", []string{"data", "transform"}},
		{"API Integration", "integration", []string{"api", "http"}},
	}

	for _, tmpl := range templates {
		publishPayload := map[string]interface{}{
			"name":        tmpl.name,
			"description": "Test template for " + tmpl.name,
			"category":    tmpl.category,
			"version":     "1.0.0",
			"tags":        tmpl.tags,
			"definition": map[string]interface{}{
				"nodes": []map[string]interface{}{{"id": "1", "type": "trigger"}},
				"edges": []map[string]interface{}{},
			},
		}

		resp := ts.MakeRequest(t, http.MethodPost, "/api/v1/marketplace/templates", publishPayload, headers)
		AssertStatusCode(t, resp, http.StatusCreated)
	}

	t.Logf("✓ Published %d test templates", len(templates))

	// Test category filtering
	t.Run("FilterByCategory", func(t *testing.T) {
		resp := ts.MakeRequest(t, http.MethodGet, "/api/v1/marketplace/templates?category=integration", nil, headers)
		AssertStatusCode(t, resp, http.StatusOK)

		var result map[string]interface{}
		ParseJSONResponse(t, resp, &result)

		templates := result["templates"].([]interface{})
		assert.GreaterOrEqual(t, len(templates), 2, "should find at least 2 integration templates")

		for _, tmpl := range templates {
			t := tmpl.(map[string]interface{})
			assert.Equal(t, "integration", t["category"])
		}
		t.Logf("✓ Found %d templates in 'integration' category", len(templates))
	})

	// Test tag filtering
	t.Run("FilterByTag", func(t *testing.T) {
		resp := ts.MakeRequest(t, http.MethodGet, "/api/v1/marketplace/templates?tag=http", nil, headers)
		AssertStatusCode(t, resp, http.StatusOK)

		var result map[string]interface{}
		ParseJSONResponse(t, resp, &result)

		templates := result["templates"].([]interface{})
		assert.GreaterOrEqual(t, len(templates), 2, "should find at least 2 templates with 'http' tag")
		t.Logf("✓ Found %d templates with 'http' tag", len(templates))
	})

	// Test search query
	t.Run("SearchByQuery", func(t *testing.T) {
		resp := ts.MakeRequest(t, http.MethodGet, "/api/v1/marketplace/templates?q=webhook", nil, headers)
		AssertStatusCode(t, resp, http.StatusOK)

		var result map[string]interface{}
		ParseJSONResponse(t, resp, &result)

		templates := result["templates"].([]interface{})
		assert.GreaterOrEqual(t, len(templates), 1, "should find at least 1 template matching 'webhook'")
		t.Logf("✓ Found %d templates matching 'webhook'", len(templates))
	})
}
