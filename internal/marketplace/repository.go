package marketplace

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

// Repository defines marketplace data access operations
type Repository interface {
	Publish(ctx context.Context, template *MarketplaceTemplate) error
	GetByID(ctx context.Context, id string) (*MarketplaceTemplate, error)
	Search(ctx context.Context, filter SearchFilter) ([]*MarketplaceTemplate, error)
	GetPopular(ctx context.Context, limit int) ([]*MarketplaceTemplate, error)
	GetTrending(ctx context.Context, days, limit int) ([]*MarketplaceTemplate, error)
	GetByAuthor(ctx context.Context, authorID string) ([]*MarketplaceTemplate, error)
	Update(ctx context.Context, id string, input UpdateTemplateInput) error
	IncrementDownloadCount(ctx context.Context, templateID string) error
	CreateInstallation(ctx context.Context, installation *TemplateInstallation) error
	GetInstallation(ctx context.Context, tenantID, templateID string) (*TemplateInstallation, error)
	CreateReview(ctx context.Context, review *TemplateReview) error
	UpdateReview(ctx context.Context, tenantID, reviewID string, rating int, comment string) error
	DeleteReview(ctx context.Context, tenantID, reviewID string) error
	GetReviews(ctx context.Context, templateID string, sortBy ReviewSortOption, limit, offset int) ([]*TemplateReview, error)
	GetUserReview(ctx context.Context, tenantID, templateID string) (*TemplateReview, error)
	UpdateTemplateRating(ctx context.Context, templateID string) error
	VoteReviewHelpful(ctx context.Context, vote *ReviewHelpfulVote) error
	UnvoteReviewHelpful(ctx context.Context, tenantID, userID, reviewID string) error
	HasVotedHelpful(ctx context.Context, tenantID, userID, reviewID string) (bool, error)
	CreateReviewReport(ctx context.Context, report *ReviewReport) error
	GetReviewReports(ctx context.Context, status string, limit, offset int) ([]*ReviewReport, error)
	UpdateReviewReportStatus(ctx context.Context, reportID, status, resolvedBy string, notes *string) error
	HideReview(ctx context.Context, reviewID, reason, hiddenBy string) error
	UnhideReview(ctx context.Context, reviewID string) error
	GetRatingDistribution(ctx context.Context, templateID string) (*RatingDistribution, error)
}

// PostgresRepository implements Repository using PostgreSQL
type PostgresRepository struct {
	db *sqlx.DB
}

// NewRepository creates a new marketplace repository
func NewRepository(db *sqlx.DB) Repository {
	return &PostgresRepository{db: db}
}

// Publish publishes a new marketplace template
func (r *PostgresRepository) Publish(ctx context.Context, template *MarketplaceTemplate) error {
	query := `
		INSERT INTO marketplace_templates (
			name, description, category, definition, tags,
			author_id, author_name, version, source_tenant_id, source_template_id
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id, published_at, updated_at
	`

	err := r.db.QueryRowContext(
		ctx, query,
		template.Name,
		template.Description,
		template.Category,
		template.Definition,
		pq.Array(template.Tags),
		template.AuthorID,
		template.AuthorName,
		template.Version,
		template.SourceTenantID,
		template.SourceTemplateID,
	).Scan(&template.ID, &template.PublishedAt, &template.UpdatedAt)

	if err != nil {
		if isUniqueViolation(err) {
			return fmt.Errorf("template with name %s already exists", template.Name)
		}
		return fmt.Errorf("publish template: %w", err)
	}

	return nil
}

// GetByID retrieves a marketplace template by ID
func (r *PostgresRepository) GetByID(ctx context.Context, id string) (*MarketplaceTemplate, error) {
	query := `
		SELECT id, name, description, category, definition, tags,
			   author_id, author_name, version, download_count,
			   average_rating, total_ratings,
			   rating_1_count, rating_2_count, rating_3_count, rating_4_count, rating_5_count,
			   is_verified,
			   source_tenant_id, source_template_id, published_at, updated_at
		FROM marketplace_templates
		WHERE id = $1
	`

	var template MarketplaceTemplate
	err := r.db.GetContext(ctx, &template, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("template not found")
		}
		return nil, fmt.Errorf("get template: %w", err)
	}

	return &template, nil
}

// Search searches for marketplace templates with filters
func (r *PostgresRepository) Search(ctx context.Context, filter SearchFilter) ([]*MarketplaceTemplate, error) {
	query := `
		SELECT id, name, description, category, definition, tags,
			   author_id, author_name, version, download_count,
			   average_rating, total_ratings,
			   rating_1_count, rating_2_count, rating_3_count, rating_4_count, rating_5_count,
			   is_verified,
			   source_tenant_id, source_template_id, published_at, updated_at
		FROM marketplace_templates
		WHERE 1=1
	`

	args := []interface{}{}
	argCount := 0

	if filter.Category != "" {
		argCount++
		query += fmt.Sprintf(" AND category = $%d", argCount)
		args = append(args, filter.Category)
	}

	if len(filter.Tags) > 0 {
		argCount++
		query += fmt.Sprintf(" AND tags && $%d", argCount)
		args = append(args, pq.Array(filter.Tags))
	}

	if filter.SearchQuery != "" {
		argCount++
		searchPattern := "%" + filter.SearchQuery + "%"
		query += fmt.Sprintf(" AND (name ILIKE $%d OR description ILIKE $%d)", argCount, argCount)
		args = append(args, searchPattern)
	}

	if filter.MinRating != nil {
		argCount++
		query += fmt.Sprintf(" AND average_rating >= $%d", argCount)
		args = append(args, *filter.MinRating)
	}

	if filter.IsVerified != nil {
		argCount++
		query += fmt.Sprintf(" AND is_verified = $%d", argCount)
		args = append(args, *filter.IsVerified)
	}

	query += r.buildOrderBy(filter.SortBy)

	if filter.Limit > 0 {
		argCount++
		query += fmt.Sprintf(" LIMIT $%d", argCount)
		args = append(args, filter.Limit)
	}

	if filter.Page > 0 && filter.Limit > 0 {
		argCount++
		offset := filter.Page * filter.Limit
		query += fmt.Sprintf(" OFFSET $%d", argCount)
		args = append(args, offset)
	}

	var templates []*MarketplaceTemplate
	err := r.db.SelectContext(ctx, &templates, query, args...)
	if err != nil {
		return nil, fmt.Errorf("search templates: %w", err)
	}

	return templates, nil
}

// GetPopular retrieves the most popular templates
func (r *PostgresRepository) GetPopular(ctx context.Context, limit int) ([]*MarketplaceTemplate, error) {
	query := `
		SELECT id, name, description, category, definition, tags,
			   author_id, author_name, version, download_count,
			   average_rating, total_ratings, is_verified,
			   source_tenant_id, source_template_id, published_at, updated_at
		FROM marketplace_templates
		ORDER BY download_count DESC, average_rating DESC
		LIMIT $1
	`

	var templates []*MarketplaceTemplate
	err := r.db.SelectContext(ctx, &templates, query, limit)
	if err != nil {
		return nil, fmt.Errorf("get popular templates: %w", err)
	}

	return templates, nil
}

// GetTrending retrieves trending templates
func (r *PostgresRepository) GetTrending(ctx context.Context, days, limit int) ([]*MarketplaceTemplate, error) {
	query := `
		SELECT mt.id, mt.name, mt.description, mt.category, mt.definition, mt.tags,
			   mt.author_id, mt.author_name, mt.version, mt.download_count,
			   mt.average_rating, mt.total_ratings, mt.is_verified,
			   mt.source_tenant_id, mt.source_template_id, mt.published_at, mt.updated_at,
			   COUNT(mi.id) as recent_installs
		FROM marketplace_templates mt
		LEFT JOIN marketplace_installations mi ON mt.id = mi.template_id
			AND mi.installed_at >= NOW() - INTERVAL '1 day' * $1
		GROUP BY mt.id, mt.name, mt.description, mt.category, mt.definition, mt.tags,
				 mt.author_id, mt.author_name, mt.version, mt.download_count,
				 mt.average_rating, mt.total_ratings, mt.is_verified,
				 mt.source_tenant_id, mt.source_template_id, mt.published_at, mt.updated_at
		ORDER BY recent_installs DESC, mt.average_rating DESC
		LIMIT $2
	`

	var templates []*MarketplaceTemplate
	err := r.db.SelectContext(ctx, &templates, query, days, limit)
	if err != nil {
		return nil, fmt.Errorf("get trending templates: %w", err)
	}

	return templates, nil
}

// GetByAuthor retrieves templates by author
func (r *PostgresRepository) GetByAuthor(ctx context.Context, authorID string) ([]*MarketplaceTemplate, error) {
	query := `
		SELECT id, name, description, category, definition, tags,
			   author_id, author_name, version, download_count,
			   average_rating, total_ratings, is_verified,
			   source_tenant_id, source_template_id, published_at, updated_at
		FROM marketplace_templates
		WHERE author_id = $1
		ORDER BY published_at DESC
	`

	var templates []*MarketplaceTemplate
	err := r.db.SelectContext(ctx, &templates, query, authorID)
	if err != nil {
		return nil, fmt.Errorf("get templates by author: %w", err)
	}

	return templates, nil
}

// Update updates a marketplace template
func (r *PostgresRepository) Update(ctx context.Context, id string, input UpdateTemplateInput) error {
	updates := []string{}
	args := []interface{}{}
	argCount := 0

	if input.Name != "" {
		argCount++
		updates = append(updates, fmt.Sprintf("name = $%d", argCount))
		args = append(args, input.Name)
	}

	if input.Description != "" {
		argCount++
		updates = append(updates, fmt.Sprintf("description = $%d", argCount))
		args = append(args, input.Description)
	}

	if input.Category != "" {
		argCount++
		updates = append(updates, fmt.Sprintf("category = $%d", argCount))
		args = append(args, input.Category)
	}

	if input.Definition != nil {
		argCount++
		updates = append(updates, fmt.Sprintf("definition = $%d", argCount))
		args = append(args, input.Definition)
	}

	if input.Tags != nil {
		argCount++
		updates = append(updates, fmt.Sprintf("tags = $%d", argCount))
		args = append(args, pq.Array(input.Tags))
	}

	if input.Version != "" {
		argCount++
		updates = append(updates, fmt.Sprintf("version = $%d", argCount))
		args = append(args, input.Version)
	}

	if len(updates) == 0 {
		return nil
	}

	argCount++
	args = append(args, id)

	query := fmt.Sprintf(`
		UPDATE marketplace_templates
		SET %s, updated_at = NOW()
		WHERE id = $%d
	`, strings.Join(updates, ", "), argCount)

	result, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("update template: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("template not found")
	}

	return nil
}

// IncrementDownloadCount increments the download count for a template
func (r *PostgresRepository) IncrementDownloadCount(ctx context.Context, templateID string) error {
	query := `
		UPDATE marketplace_templates
		SET download_count = download_count + 1
		WHERE id = $1
	`

	_, err := r.db.ExecContext(ctx, query, templateID)
	if err != nil {
		return fmt.Errorf("increment download count: %w", err)
	}

	return nil
}

// CreateInstallation creates a new template installation record
func (r *PostgresRepository) CreateInstallation(ctx context.Context, installation *TemplateInstallation) error {
	query := `
		INSERT INTO marketplace_installations (
			template_id, tenant_id, user_id, workflow_id, installed_version
		) VALUES ($1, $2, $3, $4, $5)
		RETURNING id, installed_at
	`

	err := r.db.QueryRowContext(
		ctx, query,
		installation.TemplateID,
		installation.TenantID,
		installation.UserID,
		installation.WorkflowID,
		installation.InstalledVersion,
	).Scan(&installation.ID, &installation.InstalledAt)

	if err != nil {
		return fmt.Errorf("create installation: %w", err)
	}

	return nil
}

// GetInstallation retrieves an installation record
func (r *PostgresRepository) GetInstallation(ctx context.Context, tenantID, templateID string) (*TemplateInstallation, error) {
	query := `
		SELECT id, template_id, tenant_id, user_id, workflow_id,
			   installed_version, installed_at
		FROM marketplace_installations
		WHERE tenant_id = $1 AND template_id = $2
		ORDER BY installed_at DESC
		LIMIT 1
	`

	var installation TemplateInstallation
	err := r.db.GetContext(ctx, &installation, query, tenantID, templateID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("installation not found")
		}
		return nil, fmt.Errorf("get installation: %w", err)
	}

	return &installation, nil
}

// CreateReview creates a new template review
func (r *PostgresRepository) CreateReview(ctx context.Context, review *TemplateReview) error {
	query := `
		INSERT INTO marketplace_reviews (
			template_id, tenant_id, user_id, user_name, rating, comment
		) VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at, updated_at
	`

	err := r.db.QueryRowContext(
		ctx, query,
		review.TemplateID,
		review.TenantID,
		review.UserID,
		review.UserName,
		review.Rating,
		review.Comment,
	).Scan(&review.ID, &review.CreatedAt, &review.UpdatedAt)

	if err != nil {
		if isUniqueViolation(err) {
			return fmt.Errorf("you have already reviewed this template")
		}
		return fmt.Errorf("create review: %w", err)
	}

	return nil
}

// UpdateReview updates an existing review
func (r *PostgresRepository) UpdateReview(ctx context.Context, tenantID, reviewID string, rating int, comment string) error {
	query := `
		UPDATE marketplace_reviews
		SET rating = $1, comment = $2, updated_at = NOW()
		WHERE id = $3 AND tenant_id = $4
	`

	result, err := r.db.ExecContext(ctx, query, rating, comment, reviewID, tenantID)
	if err != nil {
		return fmt.Errorf("update review: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("review not found")
	}

	return nil
}

// DeleteReview deletes a review
func (r *PostgresRepository) DeleteReview(ctx context.Context, tenantID, reviewID string) error {
	query := `DELETE FROM marketplace_reviews WHERE id = $1 AND tenant_id = $2`

	result, err := r.db.ExecContext(ctx, query, reviewID, tenantID)
	if err != nil {
		return fmt.Errorf("delete review: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("review not found")
	}

	return nil
}

// GetReviews retrieves reviews for a template
func (r *PostgresRepository) GetReviews(ctx context.Context, templateID string, sortBy ReviewSortOption, limit, offset int) ([]*TemplateReview, error) {
	orderClause := buildReviewOrderClause(sortBy)

	query := fmt.Sprintf(`
		SELECT id, template_id, tenant_id, user_id, user_name,
			   rating, comment, helpful_count, is_hidden,
			   hidden_reason, hidden_at, hidden_by, deleted_at,
			   created_at, updated_at
		FROM marketplace_reviews
		WHERE template_id = $1
		  AND deleted_at IS NULL
		  AND is_hidden = false
		%s
		LIMIT $2 OFFSET $3
	`, orderClause)

	var reviews []*TemplateReview
	err := r.db.SelectContext(ctx, &reviews, query, templateID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("get reviews: %w", err)
	}

	return reviews, nil
}

// buildReviewOrderClause builds the ORDER BY clause for review queries
func buildReviewOrderClause(sortBy ReviewSortOption) string {
	switch sortBy {
	case ReviewSortHelpful:
		return "ORDER BY helpful_count DESC, created_at DESC"
	case ReviewSortRatingH:
		return "ORDER BY rating DESC, created_at DESC"
	case ReviewSortRatingL:
		return "ORDER BY rating ASC, created_at DESC"
	case ReviewSortRecent:
		fallthrough
	default:
		return "ORDER BY created_at DESC"
	}
}

// GetUserReview retrieves a user's review for a template
func (r *PostgresRepository) GetUserReview(ctx context.Context, tenantID, templateID string) (*TemplateReview, error) {
	query := `
		SELECT id, template_id, tenant_id, user_id, user_name,
			   rating, comment, created_at, updated_at
		FROM marketplace_reviews
		WHERE tenant_id = $1 AND template_id = $2
	`

	var review TemplateReview
	err := r.db.GetContext(ctx, &review, query, tenantID, templateID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("review not found")
		}
		return nil, fmt.Errorf("get user review: %w", err)
	}

	return &review, nil
}

// UpdateTemplateRating recalculates and updates the template's average rating
func (r *PostgresRepository) UpdateTemplateRating(ctx context.Context, templateID string) error {
	query := `
		UPDATE marketplace_templates
		SET
			average_rating = (
				SELECT COALESCE(AVG(rating), 0)
				FROM marketplace_reviews
				WHERE template_id = $1
			),
			total_ratings = (
				SELECT COUNT(*)
				FROM marketplace_reviews
				WHERE template_id = $1
			)
		WHERE id = $1
	`

	_, err := r.db.ExecContext(ctx, query, templateID)
	if err != nil {
		return fmt.Errorf("update template rating: %w", err)
	}

	return nil
}

// VoteReviewHelpful records a helpful vote for a review
func (r *PostgresRepository) VoteReviewHelpful(ctx context.Context, vote *ReviewHelpfulVote) error {
	query := `
		INSERT INTO review_helpful_votes (review_id, tenant_id, user_id)
		VALUES ($1, $2, $3)
		RETURNING id, voted_at
	`

	err := r.db.QueryRowContext(ctx, query, vote.ReviewID, vote.TenantID, vote.UserID).
		Scan(&vote.ID, &vote.VotedAt)
	if err != nil {
		if isUniqueViolation(err) {
			return fmt.Errorf("you have already voted this review as helpful")
		}
		return fmt.Errorf("vote review helpful: %w", err)
	}

	return nil
}

// UnvoteReviewHelpful removes a helpful vote for a review
func (r *PostgresRepository) UnvoteReviewHelpful(ctx context.Context, tenantID, userID, reviewID string) error {
	query := `DELETE FROM review_helpful_votes WHERE review_id = $1 AND tenant_id = $2 AND user_id = $3`

	result, err := r.db.ExecContext(ctx, query, reviewID, tenantID, userID)
	if err != nil {
		return fmt.Errorf("unvote review helpful: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("vote not found")
	}

	return nil
}

// HasVotedHelpful checks if a user has voted a review as helpful
func (r *PostgresRepository) HasVotedHelpful(ctx context.Context, tenantID, userID, reviewID string) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM review_helpful_votes
			WHERE review_id = $1 AND tenant_id = $2 AND user_id = $3
		)
	`

	var exists bool
	err := r.db.QueryRowContext(ctx, query, reviewID, tenantID, userID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("check helpful vote: %w", err)
	}

	return exists, nil
}

// CreateReviewReport creates a new review report
func (r *PostgresRepository) CreateReviewReport(ctx context.Context, report *ReviewReport) error {
	query := `
		INSERT INTO review_reports (
			review_id, reporter_tenant_id, reporter_user_id, reason, details
		) VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at, status
	`

	err := r.db.QueryRowContext(
		ctx, query,
		report.ReviewID,
		report.ReporterTenantID,
		report.ReporterUserID,
		report.Reason,
		report.Details,
	).Scan(&report.ID, &report.CreatedAt, &report.Status)

	if err != nil {
		return fmt.Errorf("create review report: %w", err)
	}

	return nil
}

// GetReviewReports retrieves review reports filtered by status
func (r *PostgresRepository) GetReviewReports(ctx context.Context, status string, limit, offset int) ([]*ReviewReport, error) {
	query := `
		SELECT id, review_id, reporter_tenant_id, reporter_user_id,
			   reason, details, status, resolved_at, resolved_by,
			   resolution_notes, created_at
		FROM review_reports
		WHERE ($1 = '' OR status = $1)
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	var reports []*ReviewReport
	err := r.db.SelectContext(ctx, &reports, query, status, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("get review reports: %w", err)
	}

	return reports, nil
}

// UpdateReviewReportStatus updates the status of a review report
func (r *PostgresRepository) UpdateReviewReportStatus(ctx context.Context, reportID, status, resolvedBy string, notes *string) error {
	query := `
		UPDATE review_reports
		SET status = $1, resolved_at = NOW(), resolved_by = $2, resolution_notes = $3
		WHERE id = $4
	`

	result, err := r.db.ExecContext(ctx, query, status, resolvedBy, notes, reportID)
	if err != nil {
		return fmt.Errorf("update report status: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("report not found")
	}

	return nil
}

// HideReview hides a review (moderation)
func (r *PostgresRepository) HideReview(ctx context.Context, reviewID, reason, hiddenBy string) error {
	query := `
		UPDATE marketplace_reviews
		SET is_hidden = true, hidden_reason = $1, hidden_at = NOW(), hidden_by = $2
		WHERE id = $3
	`

	result, err := r.db.ExecContext(ctx, query, reason, hiddenBy, reviewID)
	if err != nil {
		return fmt.Errorf("hide review: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("review not found")
	}

	return nil
}

// UnhideReview unhides a review
func (r *PostgresRepository) UnhideReview(ctx context.Context, reviewID string) error {
	query := `
		UPDATE marketplace_reviews
		SET is_hidden = false, hidden_reason = NULL, hidden_at = NULL, hidden_by = NULL
		WHERE id = $1
	`

	result, err := r.db.ExecContext(ctx, query, reviewID)
	if err != nil {
		return fmt.Errorf("unhide review: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("review not found")
	}

	return nil
}

// GetRatingDistribution retrieves the rating distribution for a template
func (r *PostgresRepository) GetRatingDistribution(ctx context.Context, templateID string) (*RatingDistribution, error) {
	query := `
		SELECT rating_1_count, rating_2_count, rating_3_count, rating_4_count, rating_5_count,
			   total_ratings, average_rating
		FROM marketplace_templates
		WHERE id = $1
	`

	var dist RatingDistribution
	err := r.db.QueryRowContext(ctx, query, templateID).Scan(
		&dist.Rating1Count,
		&dist.Rating2Count,
		&dist.Rating3Count,
		&dist.Rating4Count,
		&dist.Rating5Count,
		&dist.TotalRatings,
		&dist.AverageRating,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("template not found")
		}
		return nil, fmt.Errorf("get rating distribution: %w", err)
	}

	if dist.TotalRatings > 0 {
		total := float64(dist.TotalRatings)
		dist.Rating1Percent = float64(dist.Rating1Count) / total * 100
		dist.Rating2Percent = float64(dist.Rating2Count) / total * 100
		dist.Rating3Percent = float64(dist.Rating3Count) / total * 100
		dist.Rating4Percent = float64(dist.Rating4Count) / total * 100
		dist.Rating5Percent = float64(dist.Rating5Count) / total * 100
	}

	return &dist, nil
}

func (r *PostgresRepository) buildOrderBy(sortBy string) string {
	switch sortBy {
	case "popular":
		return " ORDER BY download_count DESC, average_rating DESC"
	case "recent":
		return " ORDER BY published_at DESC"
	case "rating":
		return " ORDER BY average_rating DESC, total_ratings DESC"
	default:
		return " ORDER BY published_at DESC"
	}
}

func isUniqueViolation(err error) bool {
	if pqErr, ok := err.(*pq.Error); ok {
		return pqErr.Code == "23505"
	}
	return false
}
