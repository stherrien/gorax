// Package handlers contains HTTP request handlers for the API.
package handlers

// ErrorResponse represents an API error response
// @Description API error response
type ErrorResponse struct {
	// Error message
	Error string `json:"error" example:"workflow must have at least one trigger"`
	// HTTP status code
	Code int `json:"code,omitempty" example:"400"`
	// Additional details about the error
	Details map[string]interface{} `json:"details,omitempty"`
}

// SuccessResponse represents a generic success response
// @Description Generic success response
type SuccessResponse struct {
	// Success message
	Message string `json:"message" example:"Operation completed successfully"`
	// Unique identifier of the created/updated resource
	ID string `json:"id,omitempty" example:"550e8400-e29b-41d4-a716-446655440000"`
}

// PaginatedResponse represents a paginated list response
// @Description Paginated list response metadata
type PaginatedResponse struct {
	// Total number of items
	Total int `json:"total" example:"100"`
	// Current page number
	Page int `json:"page" example:"1"`
	// Items per page
	Limit int `json:"limit" example:"10"`
	// Total number of pages
	TotalPages int `json:"total_pages" example:"10"`
}

// Note: HealthResponse is defined in health.go
