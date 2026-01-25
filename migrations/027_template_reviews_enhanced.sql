-- Migration: 027_template_reviews_enhanced.sql
-- Description: Enhance template reviews with helpful votes, reports, and moderation

-- Add helpful_count and moderation fields to reviews table
ALTER TABLE marketplace_reviews
ADD COLUMN IF NOT EXISTS helpful_count INTEGER DEFAULT 0,
ADD COLUMN IF NOT EXISTS is_hidden BOOLEAN DEFAULT FALSE,
ADD COLUMN IF NOT EXISTS hidden_reason TEXT,
ADD COLUMN IF NOT EXISTS hidden_at TIMESTAMP,
ADD COLUMN IF NOT EXISTS hidden_by VARCHAR(255),
ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMP;

-- Add rating distribution columns to templates table
ALTER TABLE marketplace_templates
ADD COLUMN IF NOT EXISTS rating_1_count INTEGER DEFAULT 0,
ADD COLUMN IF NOT EXISTS rating_2_count INTEGER DEFAULT 0,
ADD COLUMN IF NOT EXISTS rating_3_count INTEGER DEFAULT 0,
ADD COLUMN IF NOT EXISTS rating_4_count INTEGER DEFAULT 0,
ADD COLUMN IF NOT EXISTS rating_5_count INTEGER DEFAULT 0;

-- Create review_helpful_votes table to track who found reviews helpful
CREATE TABLE IF NOT EXISTS review_helpful_votes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    review_id UUID NOT NULL REFERENCES marketplace_reviews(id) ON DELETE CASCADE,
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    user_id VARCHAR(255) NOT NULL,
    voted_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT unique_user_review_vote UNIQUE (review_id, tenant_id, user_id)
);

-- Create review_reports table for moderation
CREATE TABLE IF NOT EXISTS review_reports (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    review_id UUID NOT NULL REFERENCES marketplace_reviews(id) ON DELETE CASCADE,
    reporter_tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    reporter_user_id VARCHAR(255) NOT NULL,
    reason VARCHAR(50) NOT NULL,
    details TEXT,
    status VARCHAR(20) DEFAULT 'pending',
    resolved_at TIMESTAMP,
    resolved_by VARCHAR(255),
    resolution_notes TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT valid_report_reason CHECK (reason IN ('spam', 'inappropriate', 'offensive', 'misleading', 'other')),
    CONSTRAINT valid_report_status CHECK (status IN ('pending', 'reviewed', 'actioned', 'dismissed'))
);

-- Add constraints for rating distribution counts
ALTER TABLE marketplace_templates
ADD CONSTRAINT valid_rating_1_count CHECK (rating_1_count >= 0),
ADD CONSTRAINT valid_rating_2_count CHECK (rating_2_count >= 0),
ADD CONSTRAINT valid_rating_3_count CHECK (rating_3_count >= 0),
ADD CONSTRAINT valid_rating_4_count CHECK (rating_4_count >= 0),
ADD CONSTRAINT valid_rating_5_count CHECK (rating_5_count >= 0);

-- Add constraint for helpful_count
ALTER TABLE marketplace_reviews
ADD CONSTRAINT valid_helpful_count CHECK (helpful_count >= 0);

-- Create indexes for performance
CREATE INDEX idx_review_helpful_votes_review ON review_helpful_votes(review_id);
CREATE INDEX idx_review_helpful_votes_user ON review_helpful_votes(tenant_id, user_id);

CREATE INDEX idx_review_reports_review ON review_reports(review_id);
CREATE INDEX idx_review_reports_status ON review_reports(status);
CREATE INDEX idx_review_reports_created_at ON review_reports(created_at DESC);

CREATE INDEX idx_marketplace_reviews_helpful_count ON marketplace_reviews(helpful_count DESC);
CREATE INDEX idx_marketplace_reviews_is_hidden ON marketplace_reviews(is_hidden) WHERE is_hidden = false;
CREATE INDEX idx_marketplace_reviews_deleted_at ON marketplace_reviews(deleted_at) WHERE deleted_at IS NULL;

-- Function to update review helpful count
CREATE OR REPLACE FUNCTION update_review_helpful_count()
RETURNS TRIGGER AS $$
BEGIN
    IF TG_OP = 'INSERT' THEN
        UPDATE marketplace_reviews
        SET helpful_count = helpful_count + 1
        WHERE id = NEW.review_id;
    ELSIF TG_OP = 'DELETE' THEN
        UPDATE marketplace_reviews
        SET helpful_count = GREATEST(helpful_count - 1, 0)
        WHERE id = OLD.review_id;
    END IF;
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

-- Trigger to automatically update helpful count
CREATE TRIGGER review_helpful_votes_count
    AFTER INSERT OR DELETE ON review_helpful_votes
    FOR EACH ROW
    EXECUTE FUNCTION update_review_helpful_count();

-- Function to update template rating distribution
CREATE OR REPLACE FUNCTION update_template_rating_distribution()
RETURNS TRIGGER AS $$
DECLARE
    old_rating INTEGER;
    new_rating INTEGER;
    tmpl_id UUID;
BEGIN
    IF TG_OP = 'INSERT' THEN
        new_rating := NEW.rating;
        tmpl_id := NEW.template_id;

        -- Increment the appropriate rating count
        UPDATE marketplace_templates
        SET
            rating_1_count = CASE WHEN new_rating = 1 THEN rating_1_count + 1 ELSE rating_1_count END,
            rating_2_count = CASE WHEN new_rating = 2 THEN rating_2_count + 1 ELSE rating_2_count END,
            rating_3_count = CASE WHEN new_rating = 3 THEN rating_3_count + 1 ELSE rating_3_count END,
            rating_4_count = CASE WHEN new_rating = 4 THEN rating_4_count + 1 ELSE rating_4_count END,
            rating_5_count = CASE WHEN new_rating = 5 THEN rating_5_count + 1 ELSE rating_5_count END
        WHERE id = tmpl_id;

    ELSIF TG_OP = 'UPDATE' AND OLD.rating != NEW.rating THEN
        old_rating := OLD.rating;
        new_rating := NEW.rating;
        tmpl_id := NEW.template_id;

        -- Decrement old rating, increment new rating
        UPDATE marketplace_templates
        SET
            rating_1_count = CASE
                WHEN old_rating = 1 THEN GREATEST(rating_1_count - 1, 0)
                WHEN new_rating = 1 THEN rating_1_count + 1
                ELSE rating_1_count
            END,
            rating_2_count = CASE
                WHEN old_rating = 2 THEN GREATEST(rating_2_count - 1, 0)
                WHEN new_rating = 2 THEN rating_2_count + 1
                ELSE rating_2_count
            END,
            rating_3_count = CASE
                WHEN old_rating = 3 THEN GREATEST(rating_3_count - 1, 0)
                WHEN new_rating = 3 THEN rating_3_count + 1
                ELSE rating_3_count
            END,
            rating_4_count = CASE
                WHEN old_rating = 4 THEN GREATEST(rating_4_count - 1, 0)
                WHEN new_rating = 4 THEN rating_4_count + 1
                ELSE rating_4_count
            END,
            rating_5_count = CASE
                WHEN old_rating = 5 THEN GREATEST(rating_5_count - 1, 0)
                WHEN new_rating = 5 THEN rating_5_count + 1
                ELSE rating_5_count
            END
        WHERE id = tmpl_id;

    ELSIF TG_OP = 'DELETE' THEN
        old_rating := OLD.rating;
        tmpl_id := OLD.template_id;

        -- Decrement the appropriate rating count
        UPDATE marketplace_templates
        SET
            rating_1_count = CASE WHEN old_rating = 1 THEN GREATEST(rating_1_count - 1, 0) ELSE rating_1_count END,
            rating_2_count = CASE WHEN old_rating = 2 THEN GREATEST(rating_2_count - 1, 0) ELSE rating_2_count END,
            rating_3_count = CASE WHEN old_rating = 3 THEN GREATEST(rating_3_count - 1, 0) ELSE rating_3_count END,
            rating_4_count = CASE WHEN old_rating = 4 THEN GREATEST(rating_4_count - 1, 0) ELSE rating_4_count END,
            rating_5_count = CASE WHEN old_rating = 5 THEN GREATEST(rating_5_count - 1, 0) ELSE rating_5_count END
        WHERE id = tmpl_id;
    END IF;

    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

-- Trigger to automatically update rating distribution
CREATE TRIGGER marketplace_reviews_rating_distribution
    AFTER INSERT OR UPDATE OR DELETE ON marketplace_reviews
    FOR EACH ROW
    EXECUTE FUNCTION update_template_rating_distribution();

-- Comments
COMMENT ON TABLE review_helpful_votes IS 'Tracks users who found reviews helpful';
COMMENT ON TABLE review_reports IS 'Stores reports of inappropriate reviews for moderation';
COMMENT ON COLUMN marketplace_reviews.helpful_count IS 'Number of users who found this review helpful';
COMMENT ON COLUMN marketplace_reviews.is_hidden IS 'Whether the review is hidden by moderators';
COMMENT ON COLUMN marketplace_reviews.deleted_at IS 'Soft delete timestamp';
COMMENT ON COLUMN marketplace_templates.rating_1_count IS 'Count of 1-star ratings';
COMMENT ON COLUMN marketplace_templates.rating_2_count IS 'Count of 2-star ratings';
COMMENT ON COLUMN marketplace_templates.rating_3_count IS 'Count of 3-star ratings';
COMMENT ON COLUMN marketplace_templates.rating_4_count IS 'Count of 4-star ratings';
COMMENT ON COLUMN marketplace_templates.rating_5_count IS 'Count of 5-star ratings';
