// Package response provides standardized HTTP response helpers.
package response

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

// ErrorCode represents standardized error codes.
type ErrorCode string

const (
	// ErrCodeValidation indicates a validation error.
	ErrCodeValidation ErrorCode = "validation_error"
	// ErrCodeNotFound indicates a resource was not found.
	ErrCodeNotFound ErrorCode = "not_found"
	// ErrCodeBadRequest indicates a bad request.
	ErrCodeBadRequest ErrorCode = "bad_request"
	// ErrCodeUnauthorized indicates unauthorized access.
	ErrCodeUnauthorized ErrorCode = "unauthorized"
	// ErrCodeForbidden indicates forbidden access.
	ErrCodeForbidden ErrorCode = "forbidden"
	// ErrCodeInternal indicates an internal server error.
	ErrCodeInternal ErrorCode = "internal_error"
	// ErrCodeConflict indicates a resource conflict.
	ErrCodeConflict ErrorCode = "conflict"
	// ErrCodeRateLimit indicates rate limiting is active.
	ErrCodeRateLimit ErrorCode = "rate_limit_exceeded"
)

// APIError represents a standardized error response.
type APIError struct {
	Error   string            `json:"error"`
	Code    ErrorCode         `json:"code"`
	Details map[string]string `json:"details,omitempty"`
}

// PaginatedResponse represents a paginated response.
type PaginatedResponse struct {
	Data   any `json:"data"`
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
	Total  int `json:"total"`
}

// DataResponse represents a response with data wrapper.
type DataResponse struct {
	Data any `json:"data"`
}

// JSON sends a JSON response with proper error handling.
// If encoding fails, it logs the error and returns a 500 status.
func JSON(w http.ResponseWriter, logger *slog.Logger, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		if logger != nil {
			logger.Error("failed to encode JSON response", "error", err)
		}
	}
}

// Data sends a JSON response wrapped in a data envelope.
func Data(w http.ResponseWriter, logger *slog.Logger, status int, data any) {
	JSON(w, logger, status, DataResponse{Data: data})
}

// Error sends a standardized error response.
func Error(w http.ResponseWriter, logger *slog.Logger, status int, message string, code ErrorCode) {
	JSON(w, logger, status, APIError{
		Error: message,
		Code:  code,
	})
}

// ErrorWithDetails sends an error response with additional details.
func ErrorWithDetails(w http.ResponseWriter, logger *slog.Logger, status int, message string, code ErrorCode, details map[string]string) {
	JSON(w, logger, status, APIError{
		Error:   message,
		Code:    code,
		Details: details,
	})
}

// Paginated sends paginated data with metadata.
func Paginated(w http.ResponseWriter, logger *slog.Logger, data any, limit, offset, total int) {
	JSON(w, logger, http.StatusOK, PaginatedResponse{
		Data:   data,
		Limit:  limit,
		Offset: offset,
		Total:  total,
	})
}

// PaginatedWithStatus sends paginated data with a custom status code.
func PaginatedWithStatus(w http.ResponseWriter, logger *slog.Logger, status int, data any, limit, offset, total int) {
	JSON(w, logger, status, PaginatedResponse{
		Data:   data,
		Limit:  limit,
		Offset: offset,
		Total:  total,
	})
}

// NoContent sends a 204 No Content response.
func NoContent(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}

// BadRequest sends a 400 Bad Request error.
func BadRequest(w http.ResponseWriter, logger *slog.Logger, message string) {
	Error(w, logger, http.StatusBadRequest, message, ErrCodeBadRequest)
}

// ValidationError sends a 400 validation error with field details.
func ValidationError(w http.ResponseWriter, logger *slog.Logger, message, field string) {
	ErrorWithDetails(w, logger, http.StatusBadRequest, message, ErrCodeValidation, map[string]string{
		"field": field,
	})
}

// NotFound sends a 404 Not Found error.
func NotFound(w http.ResponseWriter, logger *slog.Logger, message string) {
	Error(w, logger, http.StatusNotFound, message, ErrCodeNotFound)
}

// Unauthorized sends a 401 Unauthorized error.
func Unauthorized(w http.ResponseWriter, logger *slog.Logger, message string) {
	Error(w, logger, http.StatusUnauthorized, message, ErrCodeUnauthorized)
}

// Forbidden sends a 403 Forbidden error.
func Forbidden(w http.ResponseWriter, logger *slog.Logger, message string) {
	Error(w, logger, http.StatusForbidden, message, ErrCodeForbidden)
}

// InternalError sends a 500 Internal Server Error.
func InternalError(w http.ResponseWriter, logger *slog.Logger, message string) {
	Error(w, logger, http.StatusInternalServerError, message, ErrCodeInternal)
}

// Conflict sends a 409 Conflict error.
func Conflict(w http.ResponseWriter, logger *slog.Logger, message string) {
	Error(w, logger, http.StatusConflict, message, ErrCodeConflict)
}

// TooManyRequests sends a 429 Too Many Requests error.
func TooManyRequests(w http.ResponseWriter, logger *slog.Logger, message string) {
	Error(w, logger, http.StatusTooManyRequests, message, ErrCodeRateLimit)
}

// Created sends a 201 Created response with data.
func Created(w http.ResponseWriter, logger *slog.Logger, data any) {
	Data(w, logger, http.StatusCreated, data)
}

// OK sends a 200 OK response with data.
func OK(w http.ResponseWriter, logger *slog.Logger, data any) {
	Data(w, logger, http.StatusOK, data)
}
