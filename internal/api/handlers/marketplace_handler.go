package handlers

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"

	"github.com/gorax/gorax/internal/api/middleware"
	"github.com/gorax/gorax/internal/marketplace"
)

// MarketplaceHandler handles marketplace HTTP requests
type MarketplaceHandler struct {
	service  MarketplaceService
	logger   *slog.Logger
	validate *validator.Validate
}

// MarketplaceService defines the interface for marketplace business logic
type MarketplaceService interface {
	PublishTemplate(ctx context.Context, userID, userName string, input marketplace.PublishTemplateInput) (*marketplace.MarketplaceTemplate, error)
	GetTemplate(ctx context.Context, id string) (*marketplace.MarketplaceTemplate, error)
	SearchTemplates(ctx context.Context, filter marketplace.SearchFilter) ([]*marketplace.MarketplaceTemplate, error)
	GetTrending(ctx context.Context, limit int) ([]*marketplace.MarketplaceTemplate, error)
	GetPopular(ctx context.Context, limit int) ([]*marketplace.MarketplaceTemplate, error)
	InstallTemplate(ctx context.Context, tenantID, userID, templateID string, input marketplace.InstallTemplateInput) (*marketplace.InstallTemplateResult, error)
	RateTemplate(ctx context.Context, tenantID, userID, userName, templateID string, input marketplace.RateTemplateInput) (*marketplace.TemplateReview, error)
	GetReviews(ctx context.Context, templateID string, limit, offset int) ([]*marketplace.TemplateReview, error)
	DeleteReview(ctx context.Context, tenantID, templateID, reviewID string) error
	GetCategories() []string
}

// NewMarketplaceHandler creates a new marketplace handler
func NewMarketplaceHandler(service MarketplaceService, logger *slog.Logger) *MarketplaceHandler {
	return &MarketplaceHandler{
		service:  service,
		logger:   logger,
		validate: validator.New(),
	}
}

// ListTemplates returns all marketplace templates with optional filters
func (h *MarketplaceHandler) ListTemplates(w http.ResponseWriter, r *http.Request) {
	filter := marketplace.SearchFilter{
		Category:    r.URL.Query().Get("category"),
		SearchQuery: r.URL.Query().Get("search"),
		SortBy:      r.URL.Query().Get("sort_by"),
	}

	if tags := r.URL.Query().Get("tags"); tags != "" {
		filter.Tags = strings.Split(tags, ",")
	}

	if minRating := r.URL.Query().Get("min_rating"); minRating != "" {
		if rating, err := strconv.ParseFloat(minRating, 64); err == nil {
			filter.MinRating = &rating
		}
	}

	if isVerified := r.URL.Query().Get("is_verified"); isVerified != "" {
		verified := isVerified == "true"
		filter.IsVerified = &verified
	}

	if page := r.URL.Query().Get("page"); page != "" {
		if p, err := strconv.Atoi(page); err == nil {
			filter.Page = p
		}
	}

	if limit := r.URL.Query().Get("limit"); limit != "" {
		if l, err := strconv.Atoi(limit); err == nil {
			filter.Limit = l
		}
	}

	templates, err := h.service.SearchTemplates(r.Context(), filter)
	if err != nil {
		h.respondError(w, http.StatusInternalServerError, "failed to list templates")
		return
	}

	h.respondJSON(w, http.StatusOK, templates)
}

// GetTemplate retrieves a single marketplace template
func (h *MarketplaceHandler) GetTemplate(w http.ResponseWriter, r *http.Request) {
	templateID := chi.URLParam(r, "id")

	template, err := h.service.GetTemplate(r.Context(), templateID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			h.respondError(w, http.StatusNotFound, "template not found")
			return
		}
		h.respondError(w, http.StatusInternalServerError, "failed to get template")
		return
	}

	h.respondJSON(w, http.StatusOK, template)
}

// PublishTemplate publishes a new template to the marketplace
func (h *MarketplaceHandler) PublishTemplate(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	userName := h.getUserName(r)

	var input marketplace.PublishTemplateInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.validate.Struct(input); err != nil {
		h.respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	template, err := h.service.PublishTemplate(r.Context(), userID, userName, input)
	if err != nil {
		if strings.Contains(err.Error(), "already exists") {
			h.respondError(w, http.StatusConflict, "template with this name already exists")
			return
		}
		if strings.Contains(err.Error(), "invalid") {
			h.respondError(w, http.StatusBadRequest, err.Error())
			return
		}
		h.respondError(w, http.StatusInternalServerError, "failed to publish template")
		return
	}

	h.respondJSON(w, http.StatusCreated, template)
}

// InstallTemplate installs a template as a workflow
func (h *MarketplaceHandler) InstallTemplate(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r)
	userID := middleware.GetUserID(r)
	templateID := chi.URLParam(r, "id")

	var input marketplace.InstallTemplateInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.validate.Struct(input); err != nil {
		h.respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	result, err := h.service.InstallTemplate(r.Context(), tenantID, userID, templateID, input)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			h.respondError(w, http.StatusNotFound, "template not found")
			return
		}
		if strings.Contains(err.Error(), "already installed") {
			h.respondError(w, http.StatusConflict, "template already installed")
			return
		}
		h.respondError(w, http.StatusInternalServerError, "failed to install template")
		return
	}

	h.respondJSON(w, http.StatusOK, result)
}

// RateTemplate adds or updates a rating for a template
func (h *MarketplaceHandler) RateTemplate(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r)
	userID := middleware.GetUserID(r)
	userName := h.getUserName(r)
	templateID := chi.URLParam(r, "id")

	var input marketplace.RateTemplateInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.validate.Struct(input); err != nil {
		h.respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	review, err := h.service.RateTemplate(r.Context(), tenantID, userID, userName, templateID, input)
	if err != nil {
		if strings.Contains(err.Error(), "invalid") {
			h.respondError(w, http.StatusBadRequest, err.Error())
			return
		}
		h.respondError(w, http.StatusInternalServerError, "failed to rate template")
		return
	}

	h.respondJSON(w, http.StatusOK, review)
}

// GetReviews retrieves reviews for a template
func (h *MarketplaceHandler) GetReviews(w http.ResponseWriter, r *http.Request) {
	templateID := chi.URLParam(r, "id")

	limit := 10
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil {
			limit = l
		}
	}

	offset := 0
	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil {
			offset = o
		}
	}

	reviews, err := h.service.GetReviews(r.Context(), templateID, limit, offset)
	if err != nil {
		h.respondError(w, http.StatusInternalServerError, "failed to get reviews")
		return
	}

	h.respondJSON(w, http.StatusOK, reviews)
}

// DeleteReview deletes a review
func (h *MarketplaceHandler) DeleteReview(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r)
	templateID := chi.URLParam(r, "id")
	reviewID := chi.URLParam(r, "reviewId")

	if err := h.service.DeleteReview(r.Context(), tenantID, templateID, reviewID); err != nil {
		if strings.Contains(err.Error(), "not found") {
			h.respondError(w, http.StatusNotFound, "review not found")
			return
		}
		h.respondError(w, http.StatusInternalServerError, "failed to delete review")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// GetTrending retrieves trending templates
func (h *MarketplaceHandler) GetTrending(w http.ResponseWriter, r *http.Request) {
	limit := 10
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil {
			limit = l
		}
	}

	templates, err := h.service.GetTrending(r.Context(), limit)
	if err != nil {
		h.respondError(w, http.StatusInternalServerError, "failed to get trending templates")
		return
	}

	h.respondJSON(w, http.StatusOK, templates)
}

// GetPopular retrieves popular templates
func (h *MarketplaceHandler) GetPopular(w http.ResponseWriter, r *http.Request) {
	limit := 10
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil {
			limit = l
		}
	}

	templates, err := h.service.GetPopular(r.Context(), limit)
	if err != nil {
		h.respondError(w, http.StatusInternalServerError, "failed to get popular templates")
		return
	}

	h.respondJSON(w, http.StatusOK, templates)
}

// GetCategories retrieves all available template categories
func (h *MarketplaceHandler) GetCategories(w http.ResponseWriter, r *http.Request) {
	categories := h.service.GetCategories()
	h.respondJSON(w, http.StatusOK, categories)
}

func (h *MarketplaceHandler) getUserName(r *http.Request) string {
	if userName, ok := r.Context().Value("user_name").(string); ok {
		return userName
	}
	return "Unknown User"
}

func (h *MarketplaceHandler) respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		h.logger.Error("failed to encode response", "error", err)
	}
}

func (h *MarketplaceHandler) respondError(w http.ResponseWriter, status int, message string) {
	h.respondJSON(w, status, map[string]string{"error": message})
}
