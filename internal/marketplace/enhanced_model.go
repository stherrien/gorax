package marketplace

import "errors"

// EnhancedSearchFilter represents enhanced search filters for marketplace templates
type EnhancedSearchFilter struct {
	CategoryIDs []string `json:"category_ids,omitempty"` // Filter by category IDs
	Tags        []string `json:"tags,omitempty"`
	SearchQuery string   `json:"search_query,omitempty"`
	MinRating   *float64 `json:"min_rating,omitempty"`
	IsVerified  *bool    `json:"is_verified,omitempty"`
	IsFeatured  *bool    `json:"is_featured,omitempty"` // Filter by featured status
	SortBy      string   `json:"sort_by,omitempty"`     // popular, recent, rating, name, relevance
	Page        int      `json:"page,omitempty"`
	Limit       int      `json:"limit,omitempty"`
}

// Validate validates the enhanced search filter
func (f EnhancedSearchFilter) Validate() error {
	if f.MinRating != nil && (*f.MinRating < 0 || *f.MinRating > 5) {
		return errors.New("min_rating must be between 0 and 5")
	}
	validSortBy := map[string]bool{
		"popular":   true,
		"recent":    true,
		"rating":    true,
		"name":      true,
		"relevance": true,
	}
	if f.SortBy != "" && !validSortBy[f.SortBy] {
		return errors.New("sort_by must be one of: popular, recent, rating, name, relevance")
	}
	if f.Page < 0 {
		return errors.New("page must be non-negative")
	}
	if f.Limit < 0 {
		return errors.New("limit must be non-negative")
	}
	if f.Limit > 100 {
		return errors.New("limit must be 100 or less")
	}
	return nil
}

// ToSearchFilter converts EnhancedSearchFilter to legacy SearchFilter for backward compatibility
func (f EnhancedSearchFilter) ToSearchFilter() SearchFilter {
	var category string
	if len(f.CategoryIDs) > 0 {
		category = f.CategoryIDs[0] // Use first category for legacy support
	}
	return SearchFilter{
		Category:    category,
		Tags:        f.Tags,
		SearchQuery: f.SearchQuery,
		MinRating:   f.MinRating,
		IsVerified:  f.IsVerified,
		SortBy:      f.SortBy,
		Page:        f.Page,
		Limit:       f.Limit,
	}
}

// MarketplaceTemplateWithCategories extends MarketplaceTemplate with category relationships
type MarketplaceTemplateWithCategories struct {
	MarketplaceTemplate
	Categories []Category `json:"categories"`
}
