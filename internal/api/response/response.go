// Package response provides standardized HTTP response utilities for API handlers.
// It centralizes JSON response handling with proper error handling and consistent formats.
package response

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

// ErrorResponse represents a standardized error response structure.
type ErrorResponse struct {
	Error   string            `json:"error"`
	Code    string            `json:"code,omitempty"`
	Details map[string]string `json:"details,omitempty"`
}

// PaginatedResponse wraps paginated data with metadata.
type PaginatedResponse struct {
	Data   any `json:"data"`
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
	Total  int `json:"total"`
}

// JSON sends a JSON response with the given status code and data.
// It properly handles encoding errors by logging them and returning an error.
func JSON(w http.ResponseWriter, status int, data any) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		// Log the error - at this point headers are already sent,
		// so we can't change the response, but we should log the failure
		slog.Error("failed to encode JSON response", "error", err)
		return err
	}

	return nil
}

// Error sends a standardized error response with the given status code and message.
// The code parameter is optional and provides a machine-readable error identifier.
func Error(w http.ResponseWriter, status int, message, code string) error {
	resp := ErrorResponse{
		Error: message,
		Code:  code,
	}
	return JSON(w, status, resp)
}

// ErrorWithDetails sends a standardized error response with additional field-level details.
// Useful for validation errors where multiple fields may have issues.
func ErrorWithDetails(w http.ResponseWriter, status int, message, code string, details map[string]string) error {
	resp := ErrorResponse{
		Error:   message,
		Code:    code,
		Details: details,
	}
	return JSON(w, status, resp)
}

// Paginated sends paginated data with metadata.
// The data parameter should be the slice of items, and limit/offset/total provide pagination info.
func Paginated(w http.ResponseWriter, data any, limit, offset, total int) error {
	resp := PaginatedResponse{
		Data:   data,
		Limit:  limit,
		Offset: offset,
		Total:  total,
	}
	return JSON(w, http.StatusOK, resp)
}

// Created sends a 201 Created response with the given data.
func Created(w http.ResponseWriter, data any) error {
	return JSON(w, http.StatusCreated, data)
}

// NoContent sends a 204 No Content response.
func NoContent(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}

// BadRequest sends a 400 Bad Request error response.
func BadRequest(w http.ResponseWriter, message string) error {
	return Error(w, http.StatusBadRequest, message, "bad_request")
}

// Unauthorized sends a 401 Unauthorized error response.
func Unauthorized(w http.ResponseWriter, message string) error {
	return Error(w, http.StatusUnauthorized, message, "unauthorized")
}

// Forbidden sends a 403 Forbidden error response.
func Forbidden(w http.ResponseWriter, message string) error {
	return Error(w, http.StatusForbidden, message, "forbidden")
}

// NotFound sends a 404 Not Found error response.
func NotFound(w http.ResponseWriter, message string) error {
	return Error(w, http.StatusNotFound, message, "not_found")
}

// Conflict sends a 409 Conflict error response.
func Conflict(w http.ResponseWriter, message string) error {
	return Error(w, http.StatusConflict, message, "conflict")
}

// UnprocessableEntity sends a 422 Unprocessable Entity error response.
// Useful for validation errors.
func UnprocessableEntity(w http.ResponseWriter, message string) error {
	return Error(w, http.StatusUnprocessableEntity, message, "unprocessable_entity")
}

// ValidationError sends a 422 Unprocessable Entity error with field-level details.
func ValidationError(w http.ResponseWriter, message string, details map[string]string) error {
	return ErrorWithDetails(w, http.StatusUnprocessableEntity, message, "validation_error", details)
}

// InternalError sends a 500 Internal Server Error response.
// The actual error is not exposed to the client for security reasons.
func InternalError(w http.ResponseWriter, message string) error {
	return Error(w, http.StatusInternalServerError, message, "internal_error")
}

// ServiceUnavailable sends a 503 Service Unavailable error response.
func ServiceUnavailable(w http.ResponseWriter, message string) error {
	return Error(w, http.StatusServiceUnavailable, message, "service_unavailable")
}
