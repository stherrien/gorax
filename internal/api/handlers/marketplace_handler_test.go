package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/gorax/gorax/internal/api/middleware"
	"github.com/gorax/gorax/internal/marketplace"
	"github.com/gorax/gorax/internal/tenant"
)

// MockMarketplaceService is a mock implementation of MarketplaceService
type MockMarketplaceService struct {
	mock.Mock
}

func (m *MockMarketplaceService) PublishTemplate(ctx context.Context, userID, userName string, input marketplace.PublishTemplateInput) (*marketplace.MarketplaceTemplate, error) {
	args := m.Called(ctx, userID, userName, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*marketplace.MarketplaceTemplate), args.Error(1)
}

func (m *MockMarketplaceService) GetTemplate(ctx context.Context, id string) (*marketplace.MarketplaceTemplate, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*marketplace.MarketplaceTemplate), args.Error(1)
}

func (m *MockMarketplaceService) SearchTemplates(ctx context.Context, filter marketplace.SearchFilter) ([]*marketplace.MarketplaceTemplate, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).([]*marketplace.MarketplaceTemplate), args.Error(1)
}

func (m *MockMarketplaceService) GetTrending(ctx context.Context, limit int) ([]*marketplace.MarketplaceTemplate, error) {
	args := m.Called(ctx, limit)
	return args.Get(0).([]*marketplace.MarketplaceTemplate), args.Error(1)
}

func (m *MockMarketplaceService) GetPopular(ctx context.Context, limit int) ([]*marketplace.MarketplaceTemplate, error) {
	args := m.Called(ctx, limit)
	return args.Get(0).([]*marketplace.MarketplaceTemplate), args.Error(1)
}

func (m *MockMarketplaceService) InstallTemplate(ctx context.Context, tenantID, userID, templateID string, input marketplace.InstallTemplateInput) (*marketplace.InstallTemplateResult, error) {
	args := m.Called(ctx, tenantID, userID, templateID, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*marketplace.InstallTemplateResult), args.Error(1)
}

func (m *MockMarketplaceService) RateTemplate(ctx context.Context, tenantID, userID, userName, templateID string, input marketplace.RateTemplateInput) (*marketplace.TemplateReview, error) {
	args := m.Called(ctx, tenantID, userID, userName, templateID, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*marketplace.TemplateReview), args.Error(1)
}

func (m *MockMarketplaceService) GetReviews(ctx context.Context, templateID string, limit, offset int) ([]*marketplace.TemplateReview, error) {
	args := m.Called(ctx, templateID, limit, offset)
	return args.Get(0).([]*marketplace.TemplateReview), args.Error(1)
}

func (m *MockMarketplaceService) DeleteReview(ctx context.Context, tenantID, templateID, reviewID string) error {
	args := m.Called(ctx, tenantID, templateID, reviewID)
	return args.Error(0)
}

func (m *MockMarketplaceService) GetCategories() []string {
	args := m.Called()
	return args.Get(0).([]string)
}

func TestMarketplaceHandler_ListTemplates(t *testing.T) {
	service := new(MockMarketplaceService)
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	handler := NewMarketplaceHandler(service, logger)

	templates := []*marketplace.MarketplaceTemplate{
		{ID: "1", Name: "Template 1"},
		{ID: "2", Name: "Template 2"},
	}

	service.On("SearchTemplates", mock.Anything, mock.AnythingOfType("marketplace.SearchFilter")).Return(templates, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/marketplace/templates", nil)
	w := httptest.NewRecorder()

	handler.ListTemplates(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var response []*marketplace.MarketplaceTemplate
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)
	assert.Len(t, response, 2)
}

func TestMarketplaceHandler_GetTemplate(t *testing.T) {
	service := new(MockMarketplaceService)
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	handler := NewMarketplaceHandler(service, logger)

	template := &marketplace.MarketplaceTemplate{
		ID:   "template-1",
		Name: "Test Template",
	}

	service.On("GetTemplate", mock.Anything, "template-1").Return(template, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/marketplace/templates/template-1", nil)
	w := httptest.NewRecorder()

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "template-1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	handler.GetTemplate(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var response marketplace.MarketplaceTemplate
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)
	assert.Equal(t, "template-1", response.ID)
}

func TestPublishTemplate(t *testing.T) {
	service := new(MockMarketplaceService)
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	handler := NewMarketplaceHandler(service, logger)

	definition := json.RawMessage(`{"nodes":[],"edges":[]}`)
	input := marketplace.PublishTemplateInput{
		Name:        "Test Template",
		Description: "This is a test template description that is long enough",
		Category:    "automation",
		Definition:  definition,
		Version:     "1.0.0",
	}

	template := &marketplace.MarketplaceTemplate{
		ID:   "template-1",
		Name: input.Name,
	}

	service.On("PublishTemplate", mock.Anything, "user-1", "Test User", input).Return(template, nil)

	body, _ := json.Marshal(input)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/marketplace/templates", bytes.NewReader(body))
	w := httptest.NewRecorder()

	user := &middleware.User{
		ID:       "user-1",
		Email:    "test@example.com",
		TenantID: "tenant-1",
	}
	ctx := context.WithValue(req.Context(), middleware.UserContextKey, user)
	ctx = context.WithValue(ctx, "user_name", "Test User")
	req = req.WithContext(ctx)

	handler.PublishTemplate(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestInstallTemplate(t *testing.T) {
	service := new(MockMarketplaceService)
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	handler := NewMarketplaceHandler(service, logger)

	input := marketplace.InstallTemplateInput{
		WorkflowName: "My Workflow",
	}

	result := &marketplace.InstallTemplateResult{
		WorkflowID:   "workflow-1",
		WorkflowName: "My Workflow",
	}

	service.On("InstallTemplate", mock.Anything, "tenant-1", "user-1", "template-1", input).Return(result, nil)

	body, _ := json.Marshal(input)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/marketplace/templates/template-1/install", bytes.NewReader(body))
	w := httptest.NewRecorder()

	ten := &tenant.Tenant{ID: "tenant-1", Name: "Test Tenant"}
	user := &middleware.User{ID: "user-1", Email: "test@example.com", TenantID: "tenant-1"}
	ctx := context.WithValue(req.Context(), middleware.TenantContextKey, ten)
	ctx = context.WithValue(ctx, middleware.UserContextKey, user)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "template-1")
	ctx = context.WithValue(ctx, chi.RouteCtxKey, rctx)
	req = req.WithContext(ctx)

	handler.InstallTemplate(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRateTemplate(t *testing.T) {
	service := new(MockMarketplaceService)
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	handler := NewMarketplaceHandler(service, logger)

	input := marketplace.RateTemplateInput{
		Rating:  5,
		Comment: "Great template!",
	}

	review := &marketplace.TemplateReview{
		ID:     "review-1",
		Rating: 5,
	}

	service.On("RateTemplate", mock.Anything, "tenant-1", "user-1", "Test User", "template-1", input).Return(review, nil)

	body, _ := json.Marshal(input)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/marketplace/templates/template-1/rate", bytes.NewReader(body))
	w := httptest.NewRecorder()

	ten := &tenant.Tenant{ID: "tenant-1", Name: "Test Tenant"}
	user := &middleware.User{ID: "user-1", Email: "test@example.com", TenantID: "tenant-1"}
	ctx := context.WithValue(req.Context(), middleware.TenantContextKey, ten)
	ctx = context.WithValue(ctx, middleware.UserContextKey, user)
	ctx = context.WithValue(ctx, "user_name", "Test User")
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "template-1")
	ctx = context.WithValue(ctx, chi.RouteCtxKey, rctx)
	req = req.WithContext(ctx)

	handler.RateTemplate(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestGetTrending(t *testing.T) {
	service := new(MockMarketplaceService)
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	handler := NewMarketplaceHandler(service, logger)

	templates := []*marketplace.MarketplaceTemplate{
		{ID: "1", Name: "Trending 1"},
	}

	service.On("GetTrending", mock.Anything, 10).Return(templates, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/marketplace/trending", nil)
	w := httptest.NewRecorder()

	handler.GetTrending(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestGetCategories(t *testing.T) {
	service := new(MockMarketplaceService)
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	handler := NewMarketplaceHandler(service, logger)

	categories := []string{"security", "automation"}
	service.On("GetCategories").Return(categories)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/marketplace/categories", nil)
	w := httptest.NewRecorder()

	handler.GetCategories(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var response []string
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)
	assert.Len(t, response, 2)
}

// Integration Tests - Full Request/Response Cycle

func TestMarketplaceIntegration_ListTemplates_WithComplexFilters(t *testing.T) {
	t.Run("filters by category and tags", func(t *testing.T) {
		service := new(MockMarketplaceService)
		logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
		handler := NewMarketplaceHandler(service, logger)

		templates := []*marketplace.MarketplaceTemplate{
			{
				ID:            "tpl-1",
				Name:          "Security Workflow",
				Description:   "Advanced security workflow",
				Category:      "security",
				Tags:          []string{"webhook", "api"},
				IsVerified:    true,
				AverageRating: 4.8,
			},
			{
				ID:            "tpl-2",
				Name:          "Auth Flow",
				Description:   "OAuth authentication flow",
				Category:      "security",
				Tags:          []string{"oauth", "api"},
				IsVerified:    true,
				AverageRating: 4.5,
			},
		}

		service.On("SearchTemplates", mock.Anything, mock.MatchedBy(func(filter marketplace.SearchFilter) bool {
			return filter.Category == "security" &&
				len(filter.Tags) == 2 &&
				filter.Tags[0] == "api" &&
				filter.Tags[1] == "webhook" &&
				*filter.IsVerified == true &&
				*filter.MinRating == 4.0
		})).Return(templates, nil)

		req := httptest.NewRequest(
			http.MethodGet,
			"/api/v1/marketplace/templates?category=security&tags=api,webhook&is_verified=true&min_rating=4.0",
			nil,
		)
		w := httptest.NewRecorder()

		handler.ListTemplates(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

		var response []*marketplace.MarketplaceTemplate
		err := json.NewDecoder(w.Body).Decode(&response)
		require.NoError(t, err)
		assert.Len(t, response, 2)
		assert.Equal(t, "Security Workflow", response[0].Name)
		assert.True(t, response[0].IsVerified)
		service.AssertExpectations(t)
	})

	t.Run("handles pagination parameters", func(t *testing.T) {
		service := new(MockMarketplaceService)
		logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
		handler := NewMarketplaceHandler(service, logger)

		service.On("SearchTemplates", mock.Anything, mock.MatchedBy(func(filter marketplace.SearchFilter) bool {
			return filter.Page == 2 && filter.Limit == 20
		})).Return([]*marketplace.MarketplaceTemplate{}, nil)

		req := httptest.NewRequest(
			http.MethodGet,
			"/api/v1/marketplace/templates?page=2&limit=20",
			nil,
		)
		w := httptest.NewRecorder()

		handler.ListTemplates(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		service.AssertExpectations(t)
	})

	t.Run("handles search query", func(t *testing.T) {
		service := new(MockMarketplaceService)
		logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
		handler := NewMarketplaceHandler(service, logger)

		templates := []*marketplace.MarketplaceTemplate{
			{ID: "tpl-1", Name: "Slack Integration", Description: "Send messages to Slack"},
		}

		service.On("SearchTemplates", mock.Anything, mock.MatchedBy(func(filter marketplace.SearchFilter) bool {
			return filter.SearchQuery == "slack"
		})).Return(templates, nil)

		req := httptest.NewRequest(
			http.MethodGet,
			"/api/v1/marketplace/templates?search=slack",
			nil,
		)
		w := httptest.NewRecorder()

		handler.ListTemplates(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response []*marketplace.MarketplaceTemplate
		err := json.NewDecoder(w.Body).Decode(&response)
		require.NoError(t, err)
		assert.Len(t, response, 1)
		assert.Contains(t, response[0].Name, "Slack")
		service.AssertExpectations(t)
	})
}

func TestMarketplaceIntegration_PublishTemplate_ValidationScenarios(t *testing.T) {
	t.Run("publishes valid template with all fields", func(t *testing.T) {
		service := new(MockMarketplaceService)
		logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
		handler := NewMarketplaceHandler(service, logger)

		definition := json.RawMessage(`{
			"nodes": [
				{"id": "1", "type": "trigger:webhook"},
				{"id": "2", "type": "action:http"}
			],
			"edges": [{"from": "1", "to": "2"}]
		}`)

		input := marketplace.PublishTemplateInput{
			Name:        "Complete Workflow Template",
			Description: "This is a comprehensive workflow template with all features enabled for testing purposes",
			Category:    "automation",
			Definition:  definition,
			Version:     "1.0.0",
			Tags:        []string{"webhook", "http", "integration"},
		}

		expected := &marketplace.MarketplaceTemplate{
			ID:            "tpl-new-123",
			Name:          input.Name,
			Description:   input.Description,
			Category:      input.Category,
			Tags:          input.Tags,
			Version:       input.Version,
			AuthorID:      "user-1",
			AuthorName:    "Test Publisher",
			IsVerified:    false,
			AverageRating: 0.0,
		}

		service.On("PublishTemplate",
			mock.Anything,
			"user-1",
			"Test Publisher",
			mock.MatchedBy(func(i marketplace.PublishTemplateInput) bool {
				return i.Name == input.Name &&
					i.Description == input.Description &&
					i.Category == input.Category
			}),
		).Return(expected, nil)

		body, _ := json.Marshal(input)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/marketplace/templates", bytes.NewReader(body))
		user := &middleware.User{ID: "user-1", Email: "test@example.com", TenantID: "tenant-1"}
		ctx := context.WithValue(req.Context(), middleware.UserContextKey, user)
		ctx = context.WithValue(ctx, "user_name", "Test Publisher")
		req = req.WithContext(ctx)

		w := httptest.NewRecorder()

		handler.PublishTemplate(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)

		var response marketplace.MarketplaceTemplate
		err := json.NewDecoder(w.Body).Decode(&response)
		require.NoError(t, err)
		assert.Equal(t, "tpl-new-123", response.ID)
		assert.Equal(t, "Complete Workflow Template", response.Name)
		assert.Equal(t, "user-1", response.AuthorID)
		assert.Equal(t, "Test Publisher", response.AuthorName)
		assert.Len(t, response.Tags, 3)
		service.AssertExpectations(t)
	})

	t.Run("rejects duplicate template name", func(t *testing.T) {
		service := new(MockMarketplaceService)
		logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
		handler := NewMarketplaceHandler(service, logger)

		definition := json.RawMessage(`{"nodes":[],"edges":[]}`)
		input := marketplace.PublishTemplateInput{
			Name:        "Existing Template",
			Description: "This template name already exists in the marketplace",
			Category:    "automation",
			Definition:  definition,
			Version:     "1.0.0",
		}

		service.On("PublishTemplate",
			mock.Anything,
			mock.Anything,
			mock.Anything,
			mock.MatchedBy(func(i marketplace.PublishTemplateInput) bool {
				return i.Name == input.Name
			}),
		).Return(nil, errors.New("template with this name already exists"))

		body, _ := json.Marshal(input)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/marketplace/templates", bytes.NewReader(body))
		user := &middleware.User{ID: "user-1"}
		ctx := context.WithValue(req.Context(), middleware.UserContextKey, user)
		ctx = context.WithValue(ctx, "user_name", "User")
		req = req.WithContext(ctx)

		w := httptest.NewRecorder()

		handler.PublishTemplate(w, req)

		assert.Equal(t, http.StatusConflict, w.Code)

		var response map[string]string
		json.NewDecoder(w.Body).Decode(&response)
		assert.Contains(t, response["error"], "already exists")
		service.AssertExpectations(t)
	})
}

func TestMarketplaceIntegration_InstallTemplate_FullFlow(t *testing.T) {
	t.Run("installs template and creates workflow", func(t *testing.T) {
		service := new(MockMarketplaceService)
		logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
		handler := NewMarketplaceHandler(service, logger)

		input := marketplace.InstallTemplateInput{
			WorkflowName: "My Custom Workflow",
		}

		result := &marketplace.InstallTemplateResult{
			WorkflowID:   "wf-installed-123",
			WorkflowName: "My Custom Workflow",
		}

		service.On("InstallTemplate",
			mock.Anything,
			"tenant-install-123",
			"user-install-456",
			"tpl-source-456",
			input,
		).Return(result, nil)

		body, _ := json.Marshal(input)
		req := httptest.NewRequest(
			http.MethodPost,
			"/api/v1/marketplace/templates/tpl-source-456/install",
			bytes.NewReader(body),
		)

		ten := &tenant.Tenant{ID: "tenant-install-123", Name: "Install Tenant"}
		user := &middleware.User{ID: "user-install-456", TenantID: "tenant-install-123"}
		ctx := context.WithValue(req.Context(), middleware.TenantContextKey, ten)
		ctx = context.WithValue(ctx, middleware.UserContextKey, user)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", "tpl-source-456")
		ctx = context.WithValue(ctx, chi.RouteCtxKey, rctx)
		req = req.WithContext(ctx)

		w := httptest.NewRecorder()

		handler.InstallTemplate(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response marketplace.InstallTemplateResult
		err := json.NewDecoder(w.Body).Decode(&response)
		require.NoError(t, err)
		assert.Equal(t, "wf-installed-123", response.WorkflowID)
		assert.Equal(t, "My Custom Workflow", response.WorkflowName)
		service.AssertExpectations(t)
	})

	t.Run("handles template not found", func(t *testing.T) {
		service := new(MockMarketplaceService)
		logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
		handler := NewMarketplaceHandler(service, logger)

		input := marketplace.InstallTemplateInput{WorkflowName: "Test"}

		service.On("InstallTemplate",
			mock.Anything,
			mock.Anything,
			mock.Anything,
			"nonexistent-template",
			input,
		).Return(nil, errors.New("template not found"))

		body, _ := json.Marshal(input)
		req := httptest.NewRequest(
			http.MethodPost,
			"/api/v1/marketplace/templates/nonexistent-template/install",
			bytes.NewReader(body),
		)

		ten := &tenant.Tenant{ID: "tenant-1"}
		user := &middleware.User{ID: "user-1"}
		ctx := context.WithValue(req.Context(), middleware.TenantContextKey, ten)
		ctx = context.WithValue(ctx, middleware.UserContextKey, user)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", "nonexistent-template")
		ctx = context.WithValue(ctx, chi.RouteCtxKey, rctx)
		req = req.WithContext(ctx)

		w := httptest.NewRecorder()

		handler.InstallTemplate(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		service.AssertExpectations(t)
	})

	t.Run("handles already installed template", func(t *testing.T) {
		service := new(MockMarketplaceService)
		logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
		handler := NewMarketplaceHandler(service, logger)

		input := marketplace.InstallTemplateInput{WorkflowName: "Duplicate"}

		service.On("InstallTemplate",
			mock.Anything,
			mock.Anything,
			mock.Anything,
			"tpl-1",
			input,
		).Return(nil, errors.New("template already installed"))

		body, _ := json.Marshal(input)
		req := httptest.NewRequest(
			http.MethodPost,
			"/api/v1/marketplace/templates/tpl-1/install",
			bytes.NewReader(body),
		)

		ten := &tenant.Tenant{ID: "tenant-1"}
		user := &middleware.User{ID: "user-1"}
		ctx := context.WithValue(req.Context(), middleware.TenantContextKey, ten)
		ctx = context.WithValue(ctx, middleware.UserContextKey, user)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", "tpl-1")
		ctx = context.WithValue(ctx, chi.RouteCtxKey, rctx)
		req = req.WithContext(ctx)

		w := httptest.NewRecorder()

		handler.InstallTemplate(w, req)

		assert.Equal(t, http.StatusConflict, w.Code)
		service.AssertExpectations(t)
	})
}

func TestMarketplaceIntegration_RateTemplate_ReviewFlow(t *testing.T) {
	t.Run("submits rating with comment", func(t *testing.T) {
		service := new(MockMarketplaceService)
		logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
		handler := NewMarketplaceHandler(service, logger)

		input := marketplace.RateTemplateInput{
			Rating:  5,
			Comment: "Excellent template! Saved me hours of work. Highly recommended for API integrations.",
		}

		review := &marketplace.TemplateReview{
			ID:         "review-123",
			TemplateID: "tpl-456",
			UserID:     "user-789",
			UserName:   "Happy User",
			Rating:     5,
			Comment:    input.Comment,
			CreatedAt:  time.Now(),
		}

		service.On("RateTemplate",
			mock.Anything,
			"tenant-1",
			"user-789",
			"Happy User",
			"tpl-456",
			input,
		).Return(review, nil)

		body, _ := json.Marshal(input)
		req := httptest.NewRequest(
			http.MethodPost,
			"/api/v1/marketplace/templates/tpl-456/rate",
			bytes.NewReader(body),
		)

		ten := &tenant.Tenant{ID: "tenant-1"}
		user := &middleware.User{ID: "user-789", TenantID: "tenant-1"}
		ctx := context.WithValue(req.Context(), middleware.TenantContextKey, ten)
		ctx = context.WithValue(ctx, middleware.UserContextKey, user)
		ctx = context.WithValue(ctx, "user_name", "Happy User")
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", "tpl-456")
		ctx = context.WithValue(ctx, chi.RouteCtxKey, rctx)
		req = req.WithContext(ctx)

		w := httptest.NewRecorder()

		handler.RateTemplate(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response marketplace.TemplateReview
		err := json.NewDecoder(w.Body).Decode(&response)
		require.NoError(t, err)
		assert.Equal(t, "review-123", response.ID)
		assert.Equal(t, 5, response.Rating)
		assert.Contains(t, response.Comment, "Excellent")
		service.AssertExpectations(t)
	})

	t.Run("retrieves paginated reviews", func(t *testing.T) {
		service := new(MockMarketplaceService)
		logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
		handler := NewMarketplaceHandler(service, logger)

		reviews := []*marketplace.TemplateReview{
			{ID: "rev-1", Rating: 5, Comment: "Great!"},
			{ID: "rev-2", Rating: 4, Comment: "Good"},
			{ID: "rev-3", Rating: 5, Comment: "Awesome"},
		}

		service.On("GetReviews",
			mock.Anything,
			"tpl-popular",
			15,
			10,
		).Return(reviews, nil)

		req := httptest.NewRequest(
			http.MethodGet,
			"/api/v1/marketplace/templates/tpl-popular/reviews?limit=15&offset=10",
			nil,
		)

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", "tpl-popular")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()

		handler.GetReviews(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response []*marketplace.TemplateReview
		err := json.NewDecoder(w.Body).Decode(&response)
		require.NoError(t, err)
		assert.Len(t, response, 3)
		service.AssertExpectations(t)
	})
}

func TestMarketplaceIntegration_TrendingAndPopular(t *testing.T) {
	t.Run("retrieves trending templates with custom limit", func(t *testing.T) {
		service := new(MockMarketplaceService)
		logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
		handler := NewMarketplaceHandler(service, logger)

		templates := []*marketplace.MarketplaceTemplate{
			{ID: "tpl-1", Name: "Trending 1", DownloadCount: 500, AverageRating: 4.9},
			{ID: "tpl-2", Name: "Trending 2", DownloadCount: 450, AverageRating: 4.8},
			{ID: "tpl-3", Name: "Trending 3", DownloadCount: 400, AverageRating: 4.7},
		}

		service.On("GetTrending", mock.Anything, 3).Return(templates, nil)

		req := httptest.NewRequest(
			http.MethodGet,
			"/api/v1/marketplace/trending?limit=3",
			nil,
		)
		w := httptest.NewRecorder()

		handler.GetTrending(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response []*marketplace.MarketplaceTemplate
		err := json.NewDecoder(w.Body).Decode(&response)
		require.NoError(t, err)
		assert.Len(t, response, 3)
		assert.Equal(t, 500, response[0].DownloadCount)
		service.AssertExpectations(t)
	})

	t.Run("retrieves popular templates", func(t *testing.T) {
		service := new(MockMarketplaceService)
		logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
		handler := NewMarketplaceHandler(service, logger)

		templates := []*marketplace.MarketplaceTemplate{
			{ID: "tpl-pop-1", Name: "Popular 1", DownloadCount: 1000, AverageRating: 5.0},
			{ID: "tpl-pop-2", Name: "Popular 2", DownloadCount: 900, AverageRating: 4.9},
		}

		service.On("GetPopular", mock.Anything, 10).Return(templates, nil)

		req := httptest.NewRequest(
			http.MethodGet,
			"/api/v1/marketplace/popular",
			nil,
		)
		w := httptest.NewRecorder()

		handler.GetPopular(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response []*marketplace.MarketplaceTemplate
		err := json.NewDecoder(w.Body).Decode(&response)
		require.NoError(t, err)
		assert.Len(t, response, 2)
		service.AssertExpectations(t)
	})
}
