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
	GetReviews(ctx context.Context, templateID string, sortBy marketplace.ReviewSortOption, limit, offset int) ([]*marketplace.TemplateReview, error)
	DeleteReview(ctx context.Context, tenantID, templateID, reviewID string) error
	GetCategories() []string
	VoteReviewHelpful(ctx context.Context, tenantID, userID, reviewID string) error
	UnvoteReviewHelpful(ctx context.Context, tenantID, userID, reviewID string) error
	ReportReview(ctx context.Context, tenantID, userID, reviewID string, input marketplace.ReportReviewInput) error
	GetReviewReports(ctx context.Context, status string, limit, offset int) ([]*marketplace.ReviewReport, error)
	ResolveReviewReport(ctx context.Context, reportID, status, resolvedBy string, notes *string) error
	HideReview(ctx context.Context, reviewID, reason, hiddenBy string) error
	GetRatingDistribution(ctx context.Context, templateID string) (*marketplace.RatingDistribution, error)
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
// @Summary List marketplace templates
// @Description Search and filter marketplace templates by category, tags, rating, and verification status
// @Tags Marketplace
// @Accept json
// @Produce json
// @Param category query string false "Filter by category"
// @Param search query string false "Search query for template name or description"
// @Param tags query string false "Comma-separated list of tags"
// @Param min_rating query number false "Minimum rating (0-5)"
// @Param is_verified query boolean false "Filter by verification status"
// @Param sort_by query string false "Sort field (created_at, updated_at, rating, installs)" Enums(created_at, updated_at, rating, installs)
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Results per page" default(20)
// @Security TenantID
// @Security UserID
// @Success 200 {array} marketplace.MarketplaceTemplate "List of templates"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/v1/marketplace/templates [get]
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
// @Summary Get marketplace template
// @Description Retrieves detailed information about a specific marketplace template
// @Tags Marketplace
// @Accept json
// @Produce json
// @Param id path string true "Template ID"
// @Security TenantID
// @Security UserID
// @Success 200 {object} marketplace.MarketplaceTemplate "Template details"
// @Failure 404 {object} map[string]string "Template not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/v1/marketplace/templates/{id} [get]
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
// @Summary Publish template to marketplace
// @Description Publishes a workflow as a reusable template in the marketplace
// @Tags Marketplace
// @Accept json
// @Produce json
// @Param template body marketplace.PublishTemplateInput true "Template publication data"
// @Security TenantID
// @Security UserID
// @Success 201 {object} marketplace.MarketplaceTemplate "Published template"
// @Failure 400 {object} map[string]string "Invalid request or validation error"
// @Failure 409 {object} map[string]string "Template already exists"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/v1/marketplace/templates [post]
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
// @Summary Install marketplace template
// @Description Installs a marketplace template as a workflow in the tenant's account
// @Tags Marketplace
// @Accept json
// @Produce json
// @Param id path string true "Template ID"
// @Param installation body marketplace.InstallTemplateInput true "Installation configuration"
// @Security TenantID
// @Security UserID
// @Success 200 {object} marketplace.InstallTemplateResult "Installation result with workflow ID"
// @Failure 400 {object} map[string]string "Invalid request"
// @Failure 404 {object} map[string]string "Template not found"
// @Failure 409 {object} map[string]string "Template already installed"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/v1/marketplace/templates/{id}/install [post]
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
// @Summary Rate marketplace template
// @Description Submits or updates a rating and review for a marketplace template
// @Tags Marketplace
// @Accept json
// @Produce json
// @Param id path string true "Template ID"
// @Param rating body marketplace.RateTemplateInput true "Rating data (1-5 stars and optional comment)"
// @Security TenantID
// @Security UserID
// @Success 200 {object} marketplace.TemplateReview "Submitted review"
// @Failure 400 {object} map[string]string "Invalid request"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/v1/marketplace/templates/{id}/rate [post]
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
// @Summary Get template reviews
// @Description Retrieves paginated reviews and ratings for a marketplace template
// @Tags Marketplace
// @Accept json
// @Produce json
// @Param id path string true "Template ID"
// @Param sort query string false "Sort by: recent, helpful, rating_high, rating_low" default(recent)
// @Param limit query int false "Maximum results" default(10)
// @Param offset query int false "Pagination offset" default(0)
// @Security TenantID
// @Security UserID
// @Success 200 {array} marketplace.TemplateReview "List of reviews"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/v1/marketplace/templates/{id}/reviews [get]
func (h *MarketplaceHandler) GetReviews(w http.ResponseWriter, r *http.Request) {
	templateID := chi.URLParam(r, "id")

	sortBy := marketplace.ReviewSortRecent
	if sortStr := r.URL.Query().Get("sort"); sortStr != "" {
		sortBy = marketplace.ReviewSortOption(sortStr)
	}

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

	reviews, err := h.service.GetReviews(r.Context(), templateID, sortBy, limit, offset)
	if err != nil {
		h.respondError(w, http.StatusInternalServerError, "failed to get reviews")
		return
	}

	h.respondJSON(w, http.StatusOK, reviews)
}

// DeleteReview deletes a review
// @Summary Delete template review
// @Description Deletes a review (only allowed for review author or admin)
// @Tags Marketplace
// @Accept json
// @Produce json
// @Param id path string true "Template ID"
// @Param reviewId path string true "Review ID"
// @Security TenantID
// @Security UserID
// @Success 204 "Review deleted successfully"
// @Failure 404 {object} map[string]string "Review not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/v1/marketplace/templates/{id}/reviews/{reviewId} [delete]
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
// @Summary Get trending templates
// @Description Returns templates that are currently trending based on recent installs and views
// @Tags Marketplace
// @Accept json
// @Produce json
// @Param limit query int false "Maximum results" default(10)
// @Security TenantID
// @Security UserID
// @Success 200 {array} marketplace.MarketplaceTemplate "Trending templates"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/v1/marketplace/trending [get]
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
// @Summary Get popular templates
// @Description Returns the most popular templates based on total installs and ratings
// @Tags Marketplace
// @Accept json
// @Produce json
// @Param limit query int false "Maximum results" default(10)
// @Security TenantID
// @Security UserID
// @Success 200 {array} marketplace.MarketplaceTemplate "Popular templates"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/v1/marketplace/popular [get]
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
// @Summary Get template categories
// @Description Returns the list of all available template categories
// @Tags Marketplace
// @Accept json
// @Produce json
// @Success 200 {array} string "List of category names"
// @Router /api/v1/marketplace/categories [get]
func (h *MarketplaceHandler) GetCategories(w http.ResponseWriter, r *http.Request) {
	categories := h.service.GetCategories()
	h.respondJSON(w, http.StatusOK, categories)
}

// VoteReviewHelpful marks a review as helpful
// @Summary Vote review as helpful
// @Description Marks a review as helpful by the current user
// @Tags Marketplace
// @Accept json
// @Produce json
// @Param reviewId path string true "Review ID"
// @Security TenantID
// @Security UserID
// @Success 204 "Vote recorded successfully"
// @Failure 409 {object} map[string]string "Already voted"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/v1/marketplace/reviews/{reviewId}/helpful [post]
func (h *MarketplaceHandler) VoteReviewHelpful(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r)
	userID := middleware.GetUserID(r)
	reviewID := chi.URLParam(r, "reviewId")

	if err := h.service.VoteReviewHelpful(r.Context(), tenantID, userID, reviewID); err != nil {
		if strings.Contains(err.Error(), "already voted") {
			h.respondError(w, http.StatusConflict, "already voted helpful")
			return
		}
		h.respondError(w, http.StatusInternalServerError, "failed to vote review helpful")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// UnvoteReviewHelpful removes a helpful vote from a review
// @Summary Remove helpful vote
// @Description Removes the current user's helpful vote from a review
// @Tags Marketplace
// @Accept json
// @Produce json
// @Param reviewId path string true "Review ID"
// @Security TenantID
// @Security UserID
// @Success 204 "Vote removed successfully"
// @Failure 404 {object} map[string]string "Vote not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/v1/marketplace/reviews/{reviewId}/helpful [delete]
func (h *MarketplaceHandler) UnvoteReviewHelpful(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r)
	userID := middleware.GetUserID(r)
	reviewID := chi.URLParam(r, "reviewId")

	if err := h.service.UnvoteReviewHelpful(r.Context(), tenantID, userID, reviewID); err != nil {
		if strings.Contains(err.Error(), "not found") {
			h.respondError(w, http.StatusNotFound, "vote not found")
			return
		}
		h.respondError(w, http.StatusInternalServerError, "failed to unvote review helpful")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ReportReview reports a review for moderation
// @Summary Report review
// @Description Reports a review for inappropriate content or behavior
// @Tags Marketplace
// @Accept json
// @Produce json
// @Param reviewId path string true "Review ID"
// @Param report body marketplace.ReportReviewInput true "Report details"
// @Security TenantID
// @Security UserID
// @Success 204 "Review reported successfully"
// @Failure 400 {object} map[string]string "Invalid request"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/v1/marketplace/reviews/{reviewId}/report [post]
func (h *MarketplaceHandler) ReportReview(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r)
	userID := middleware.GetUserID(r)
	reviewID := chi.URLParam(r, "reviewId")

	var input marketplace.ReportReviewInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.validate.Struct(input); err != nil {
		h.respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.service.ReportReview(r.Context(), tenantID, userID, reviewID, input); err != nil {
		if strings.Contains(err.Error(), "invalid") {
			h.respondError(w, http.StatusBadRequest, err.Error())
			return
		}
		h.respondError(w, http.StatusInternalServerError, "failed to report review")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// GetRatingDistribution retrieves the rating distribution for a template
// @Summary Get rating distribution
// @Description Returns the distribution of ratings (1-5 stars) for a template
// @Tags Marketplace
// @Accept json
// @Produce json
// @Param id path string true "Template ID"
// @Success 200 {object} marketplace.RatingDistribution "Rating distribution"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/v1/marketplace/templates/{id}/rating-distribution [get]
func (h *MarketplaceHandler) GetRatingDistribution(w http.ResponseWriter, r *http.Request) {
	templateID := chi.URLParam(r, "id")

	distribution, err := h.service.GetRatingDistribution(r.Context(), templateID)
	if err != nil {
		h.respondError(w, http.StatusInternalServerError, "failed to get rating distribution")
		return
	}

	h.respondJSON(w, http.StatusOK, distribution)
}

// GetReviewReports retrieves review reports for admin moderation
// @Summary Get review reports (admin only)
// @Description Returns a list of review reports for moderation
// @Tags Marketplace
// @Accept json
// @Produce json
// @Param status query string false "Filter by status: pending, reviewed, actioned, dismissed"
// @Param limit query int false "Maximum results" default(20)
// @Param offset query int false "Pagination offset" default(0)
// @Security TenantID
// @Security UserID
// @Success 200 {array} marketplace.ReviewReport "List of reports"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/v1/marketplace/admin/review-reports [get]
func (h *MarketplaceHandler) GetReviewReports(w http.ResponseWriter, r *http.Request) {
	status := r.URL.Query().Get("status")

	limit := 20
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

	reports, err := h.service.GetReviewReports(r.Context(), status, limit, offset)
	if err != nil {
		h.respondError(w, http.StatusInternalServerError, "failed to get review reports")
		return
	}

	h.respondJSON(w, http.StatusOK, reports)
}

// ResolveReviewReportInput represents input for resolving a review report
type ResolveReviewReportInput struct {
	Status string  `json:"status" validate:"required,oneof=reviewed actioned dismissed"`
	Notes  *string `json:"notes,omitempty"`
}

// ResolveReviewReport resolves a review report
// @Summary Resolve review report (admin only)
// @Description Resolves a review report with a status and optional notes
// @Tags Marketplace
// @Accept json
// @Produce json
// @Param reportId path string true "Report ID"
// @Param resolution body ResolveReviewReportInput true "Resolution details"
// @Security TenantID
// @Security UserID
// @Success 204 "Report resolved successfully"
// @Failure 400 {object} map[string]string "Invalid request"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/v1/marketplace/admin/review-reports/{reportId} [put]
func (h *MarketplaceHandler) ResolveReviewReport(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	reportID := chi.URLParam(r, "reportId")

	var input ResolveReviewReportInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.validate.Struct(input); err != nil {
		h.respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.service.ResolveReviewReport(r.Context(), reportID, input.Status, userID, input.Notes); err != nil {
		if strings.Contains(err.Error(), "invalid") {
			h.respondError(w, http.StatusBadRequest, err.Error())
			return
		}
		h.respondError(w, http.StatusInternalServerError, "failed to resolve review report")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// HideReviewInput represents input for hiding a review
type HideReviewInput struct {
	Reason string `json:"reason" validate:"required,min=1,max=500"`
}

// HideReview hides a review (admin/moderator only)
// @Summary Hide review (admin only)
// @Description Hides a review from public view
// @Tags Marketplace
// @Accept json
// @Produce json
// @Param reviewId path string true "Review ID"
// @Param input body HideReviewInput true "Hide reason"
// @Security TenantID
// @Security UserID
// @Success 204 "Review hidden successfully"
// @Failure 400 {object} map[string]string "Invalid request"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/v1/marketplace/admin/reviews/{reviewId}/hide [put]
func (h *MarketplaceHandler) HideReview(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	reviewID := chi.URLParam(r, "reviewId")

	var input HideReviewInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.validate.Struct(input); err != nil {
		h.respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.service.HideReview(r.Context(), reviewID, input.Reason, userID); err != nil {
		if strings.Contains(err.Error(), "not found") {
			h.respondError(w, http.StatusNotFound, "review not found")
			return
		}
		h.respondError(w, http.StatusInternalServerError, "failed to hide review")
		return
	}

	w.WriteHeader(http.StatusNoContent)
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
