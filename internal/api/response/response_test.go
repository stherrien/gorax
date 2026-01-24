package response

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJSON(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	tests := []struct {
		name       string
		status     int
		data       any
		wantStatus int
		wantBody   map[string]any
	}{
		{
			name:       "success response",
			status:     http.StatusOK,
			data:       map[string]string{"message": "hello"},
			wantStatus: http.StatusOK,
			wantBody:   map[string]any{"message": "hello"},
		},
		{
			name:       "created response",
			status:     http.StatusCreated,
			data:       map[string]int{"id": 123},
			wantStatus: http.StatusCreated,
			wantBody:   map[string]any{"id": float64(123)},
		},
		{
			name:       "empty response",
			status:     http.StatusOK,
			data:       map[string]any{},
			wantStatus: http.StatusOK,
			wantBody:   map[string]any{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()

			JSON(w, logger, tt.status, tt.data)

			assert.Equal(t, tt.wantStatus, w.Code)
			assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

			var got map[string]any
			err := json.NewDecoder(w.Body).Decode(&got)
			require.NoError(t, err)
			assert.Equal(t, tt.wantBody, got)
		})
	}
}

func TestJSON_NilLogger(t *testing.T) {
	w := httptest.NewRecorder()

	// Should not panic with nil logger
	JSON(w, nil, http.StatusOK, map[string]string{"test": "value"})

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestData(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	tests := []struct {
		name       string
		status     int
		data       any
		wantStatus int
	}{
		{
			name:       "wraps data in envelope",
			status:     http.StatusOK,
			data:       map[string]string{"id": "abc"},
			wantStatus: http.StatusOK,
		},
		{
			name:       "created with data",
			status:     http.StatusCreated,
			data:       []string{"a", "b", "c"},
			wantStatus: http.StatusCreated,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()

			Data(w, logger, tt.status, tt.data)

			assert.Equal(t, tt.wantStatus, w.Code)

			var got DataResponse
			err := json.NewDecoder(w.Body).Decode(&got)
			require.NoError(t, err)
			assert.NotNil(t, got.Data)
		})
	}
}

func TestError(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	tests := []struct {
		name        string
		status      int
		message     string
		code        ErrorCode
		wantStatus  int
		wantMessage string
		wantCode    string
	}{
		{
			name:        "bad request",
			status:      http.StatusBadRequest,
			message:     "invalid input",
			code:        ErrCodeBadRequest,
			wantStatus:  http.StatusBadRequest,
			wantMessage: "invalid input",
			wantCode:    "bad_request",
		},
		{
			name:        "not found",
			status:      http.StatusNotFound,
			message:     "resource not found",
			code:        ErrCodeNotFound,
			wantStatus:  http.StatusNotFound,
			wantMessage: "resource not found",
			wantCode:    "not_found",
		},
		{
			name:        "internal error",
			status:      http.StatusInternalServerError,
			message:     "something went wrong",
			code:        ErrCodeInternal,
			wantStatus:  http.StatusInternalServerError,
			wantMessage: "something went wrong",
			wantCode:    "internal_error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()

			Error(w, logger, tt.status, tt.message, tt.code)

			assert.Equal(t, tt.wantStatus, w.Code)

			var got APIError
			err := json.NewDecoder(w.Body).Decode(&got)
			require.NoError(t, err)
			assert.Equal(t, tt.wantMessage, got.Error)
			assert.Equal(t, ErrorCode(tt.wantCode), got.Code)
		})
	}
}

func TestErrorWithDetails(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	w := httptest.NewRecorder()

	details := map[string]string{
		"field": "email",
		"hint":  "must be a valid email",
	}

	ErrorWithDetails(w, logger, http.StatusBadRequest, "validation failed", ErrCodeValidation, details)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var got APIError
	err := json.NewDecoder(w.Body).Decode(&got)
	require.NoError(t, err)
	assert.Equal(t, "validation failed", got.Error)
	assert.Equal(t, ErrCodeValidation, got.Code)
	assert.Equal(t, "email", got.Details["field"])
	assert.Equal(t, "must be a valid email", got.Details["hint"])
}

func TestPaginated(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	tests := []struct {
		name       string
		data       any
		limit      int
		offset     int
		total      int
		wantLimit  int
		wantOffset int
		wantTotal  int
	}{
		{
			name:       "full page",
			data:       []string{"a", "b", "c"},
			limit:      10,
			offset:     0,
			total:      3,
			wantLimit:  10,
			wantOffset: 0,
			wantTotal:  3,
		},
		{
			name:       "second page",
			data:       []string{"d", "e"},
			limit:      10,
			offset:     10,
			total:      15,
			wantLimit:  10,
			wantOffset: 10,
			wantTotal:  15,
		},
		{
			name:       "empty page",
			data:       []string{},
			limit:      10,
			offset:     100,
			total:      50,
			wantLimit:  10,
			wantOffset: 100,
			wantTotal:  50,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()

			Paginated(w, logger, tt.data, tt.limit, tt.offset, tt.total)

			assert.Equal(t, http.StatusOK, w.Code)

			var got PaginatedResponse
			err := json.NewDecoder(w.Body).Decode(&got)
			require.NoError(t, err)
			assert.Equal(t, tt.wantLimit, got.Limit)
			assert.Equal(t, tt.wantOffset, got.Offset)
			assert.Equal(t, tt.wantTotal, got.Total)
		})
	}
}

func TestPaginatedWithStatus(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	w := httptest.NewRecorder()

	PaginatedWithStatus(w, logger, http.StatusCreated, []string{"a"}, 10, 0, 1)

	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestNoContent(t *testing.T) {
	w := httptest.NewRecorder()

	NoContent(w)

	assert.Equal(t, http.StatusNoContent, w.Code)
	assert.Empty(t, w.Body.String())
}

func TestConvenienceErrors(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	tests := []struct {
		name       string
		fn         func(http.ResponseWriter, *slog.Logger, string)
		wantStatus int
		wantCode   ErrorCode
	}{
		{
			name:       "BadRequest",
			fn:         BadRequest,
			wantStatus: http.StatusBadRequest,
			wantCode:   ErrCodeBadRequest,
		},
		{
			name:       "NotFound",
			fn:         NotFound,
			wantStatus: http.StatusNotFound,
			wantCode:   ErrCodeNotFound,
		},
		{
			name:       "Unauthorized",
			fn:         Unauthorized,
			wantStatus: http.StatusUnauthorized,
			wantCode:   ErrCodeUnauthorized,
		},
		{
			name:       "Forbidden",
			fn:         Forbidden,
			wantStatus: http.StatusForbidden,
			wantCode:   ErrCodeForbidden,
		},
		{
			name:       "InternalError",
			fn:         InternalError,
			wantStatus: http.StatusInternalServerError,
			wantCode:   ErrCodeInternal,
		},
		{
			name:       "Conflict",
			fn:         Conflict,
			wantStatus: http.StatusConflict,
			wantCode:   ErrCodeConflict,
		},
		{
			name:       "TooManyRequests",
			fn:         TooManyRequests,
			wantStatus: http.StatusTooManyRequests,
			wantCode:   ErrCodeRateLimit,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()

			tt.fn(w, logger, "test message")

			assert.Equal(t, tt.wantStatus, w.Code)

			var got APIError
			err := json.NewDecoder(w.Body).Decode(&got)
			require.NoError(t, err)
			assert.Equal(t, "test message", got.Error)
			assert.Equal(t, tt.wantCode, got.Code)
		})
	}
}

func TestValidationError(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	w := httptest.NewRecorder()

	ValidationError(w, logger, "invalid email format", "email")

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var got APIError
	err := json.NewDecoder(w.Body).Decode(&got)
	require.NoError(t, err)
	assert.Equal(t, "invalid email format", got.Error)
	assert.Equal(t, ErrCodeValidation, got.Code)
	assert.Equal(t, "email", got.Details["field"])
}

func TestCreated(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	w := httptest.NewRecorder()

	Created(w, logger, map[string]string{"id": "new-id"})

	assert.Equal(t, http.StatusCreated, w.Code)

	var got DataResponse
	err := json.NewDecoder(w.Body).Decode(&got)
	require.NoError(t, err)
	assert.NotNil(t, got.Data)
}

func TestOK(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	w := httptest.NewRecorder()

	OK(w, logger, map[string]string{"status": "success"})

	assert.Equal(t, http.StatusOK, w.Code)

	var got DataResponse
	err := json.NewDecoder(w.Body).Decode(&got)
	require.NoError(t, err)
	assert.NotNil(t, got.Data)
}

func TestErrorCodes(t *testing.T) {
	// Verify error code constants have expected string values
	assert.Equal(t, ErrorCode("validation_error"), ErrCodeValidation)
	assert.Equal(t, ErrorCode("not_found"), ErrCodeNotFound)
	assert.Equal(t, ErrorCode("bad_request"), ErrCodeBadRequest)
	assert.Equal(t, ErrorCode("unauthorized"), ErrCodeUnauthorized)
	assert.Equal(t, ErrorCode("forbidden"), ErrCodeForbidden)
	assert.Equal(t, ErrorCode("internal_error"), ErrCodeInternal)
	assert.Equal(t, ErrorCode("conflict"), ErrCodeConflict)
	assert.Equal(t, ErrorCode("rate_limit_exceeded"), ErrCodeRateLimit)
}

func TestJSON_ContentTypeAlwaysSet(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	w := httptest.NewRecorder()

	JSON(w, logger, http.StatusOK, nil)

	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
}

func TestPaginatedResponse_OmitsZeroTotal(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	w := httptest.NewRecorder()

	// When total is 0, it should still be included (valid count)
	Paginated(w, logger, []string{}, 10, 0, 0)

	body := w.Body.String()
	assert.Contains(t, body, `"total":0`)
}

func TestAPIError_OmitsEmptyDetails(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	w := httptest.NewRecorder()

	Error(w, logger, http.StatusBadRequest, "error", ErrCodeBadRequest)

	body := w.Body.String()
	assert.NotContains(t, body, "details")
}
