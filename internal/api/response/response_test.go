package response

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJSON(t *testing.T) {
	tests := []struct {
		name       string
		status     int
		data       any
		wantStatus int
		wantBody   string
	}{
		{
			name:       "success with map",
			status:     http.StatusOK,
			data:       map[string]string{"message": "hello"},
			wantStatus: http.StatusOK,
			wantBody:   `{"message":"hello"}`,
		},
		{
			name:       "created with struct",
			status:     http.StatusCreated,
			data:       struct{ ID string `json:"id"` }{ID: "123"},
			wantStatus: http.StatusCreated,
			wantBody:   `{"id":"123"}`,
		},
		{
			name:       "empty slice",
			status:     http.StatusOK,
			data:       []string{},
			wantStatus: http.StatusOK,
			wantBody:   `[]`,
		},
		{
			name:       "nil data",
			status:     http.StatusOK,
			data:       nil,
			wantStatus: http.StatusOK,
			wantBody:   `null`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()

			err := JSON(w, tt.status, tt.data)

			require.NoError(t, err)
			assert.Equal(t, tt.wantStatus, w.Code)
			assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
			assert.JSONEq(t, tt.wantBody, w.Body.String())
		})
	}
}

func TestError(t *testing.T) {
	tests := []struct {
		name       string
		status     int
		message    string
		code       string
		wantStatus int
	}{
		{
			name:       "bad request",
			status:     http.StatusBadRequest,
			message:    "invalid input",
			code:       "bad_request",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "not found",
			status:     http.StatusNotFound,
			message:    "resource not found",
			code:       "not_found",
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "empty code",
			status:     http.StatusBadRequest,
			message:    "error message",
			code:       "",
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()

			err := Error(w, tt.status, tt.message, tt.code)

			require.NoError(t, err)
			assert.Equal(t, tt.wantStatus, w.Code)

			var resp ErrorResponse
			err = json.NewDecoder(w.Body).Decode(&resp)
			require.NoError(t, err)
			assert.Equal(t, tt.message, resp.Error)
			assert.Equal(t, tt.code, resp.Code)
		})
	}
}

func TestErrorWithDetails(t *testing.T) {
	w := httptest.NewRecorder()
	details := map[string]string{
		"email": "invalid email format",
		"name":  "name is required",
	}

	err := ErrorWithDetails(w, http.StatusUnprocessableEntity, "validation failed", "validation_error", details)

	require.NoError(t, err)
	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)

	var resp ErrorResponse
	err = json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, "validation failed", resp.Error)
	assert.Equal(t, "validation_error", resp.Code)
	assert.Equal(t, details, resp.Details)
}

func TestPaginated(t *testing.T) {
	tests := []struct {
		name   string
		data   any
		limit  int
		offset int
		total  int
	}{
		{
			name:   "first page",
			data:   []string{"a", "b", "c"},
			limit:  10,
			offset: 0,
			total:  100,
		},
		{
			name:   "second page",
			data:   []string{"d", "e"},
			limit:  10,
			offset: 10,
			total:  100,
		},
		{
			name:   "empty result",
			data:   []string{},
			limit:  10,
			offset: 0,
			total:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()

			err := Paginated(w, tt.data, tt.limit, tt.offset, tt.total)

			require.NoError(t, err)
			assert.Equal(t, http.StatusOK, w.Code)

			var resp PaginatedResponse
			err = json.NewDecoder(w.Body).Decode(&resp)
			require.NoError(t, err)
			assert.Equal(t, tt.limit, resp.Limit)
			assert.Equal(t, tt.offset, resp.Offset)
			assert.Equal(t, tt.total, resp.Total)
		})
	}
}

func TestCreated(t *testing.T) {
	w := httptest.NewRecorder()
	data := map[string]string{"id": "new-123"}

	err := Created(w, data)

	require.NoError(t, err)
	assert.Equal(t, http.StatusCreated, w.Code)
	assert.JSONEq(t, `{"id":"new-123"}`, w.Body.String())
}

func TestNoContent(t *testing.T) {
	w := httptest.NewRecorder()

	NoContent(w)

	assert.Equal(t, http.StatusNoContent, w.Code)
	assert.Empty(t, w.Body.String())
}

func TestBadRequest(t *testing.T) {
	w := httptest.NewRecorder()

	err := BadRequest(w, "invalid input")

	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp ErrorResponse
	err = json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, "invalid input", resp.Error)
	assert.Equal(t, "bad_request", resp.Code)
}

func TestUnauthorized(t *testing.T) {
	w := httptest.NewRecorder()

	err := Unauthorized(w, "authentication required")

	require.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var resp ErrorResponse
	err = json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, "authentication required", resp.Error)
	assert.Equal(t, "unauthorized", resp.Code)
}

func TestForbidden(t *testing.T) {
	w := httptest.NewRecorder()

	err := Forbidden(w, "access denied")

	require.NoError(t, err)
	assert.Equal(t, http.StatusForbidden, w.Code)

	var resp ErrorResponse
	err = json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, "access denied", resp.Error)
	assert.Equal(t, "forbidden", resp.Code)
}

func TestNotFound(t *testing.T) {
	w := httptest.NewRecorder()

	err := NotFound(w, "resource not found")

	require.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, w.Code)

	var resp ErrorResponse
	err = json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, "resource not found", resp.Error)
	assert.Equal(t, "not_found", resp.Code)
}

func TestConflict(t *testing.T) {
	w := httptest.NewRecorder()

	err := Conflict(w, "resource already exists")

	require.NoError(t, err)
	assert.Equal(t, http.StatusConflict, w.Code)

	var resp ErrorResponse
	err = json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, "resource already exists", resp.Error)
	assert.Equal(t, "conflict", resp.Code)
}

func TestUnprocessableEntity(t *testing.T) {
	w := httptest.NewRecorder()

	err := UnprocessableEntity(w, "validation failed")

	require.NoError(t, err)
	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)

	var resp ErrorResponse
	err = json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, "validation failed", resp.Error)
	assert.Equal(t, "unprocessable_entity", resp.Code)
}

func TestValidationError(t *testing.T) {
	w := httptest.NewRecorder()
	details := map[string]string{
		"email": "invalid format",
	}

	err := ValidationError(w, "validation failed", details)

	require.NoError(t, err)
	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)

	var resp ErrorResponse
	err = json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, "validation failed", resp.Error)
	assert.Equal(t, "validation_error", resp.Code)
	assert.Equal(t, details, resp.Details)
}

func TestInternalError(t *testing.T) {
	w := httptest.NewRecorder()

	err := InternalError(w, "something went wrong")

	require.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var resp ErrorResponse
	err = json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, "something went wrong", resp.Error)
	assert.Equal(t, "internal_error", resp.Code)
}

func TestServiceUnavailable(t *testing.T) {
	w := httptest.NewRecorder()

	err := ServiceUnavailable(w, "service is down")

	require.NoError(t, err)
	assert.Equal(t, http.StatusServiceUnavailable, w.Code)

	var resp ErrorResponse
	err = json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, "service is down", resp.Error)
	assert.Equal(t, "service_unavailable", resp.Code)
}

// TestJSONWithUnencodableData tests error handling for data that cannot be encoded.
func TestJSONWithUnencodableData(t *testing.T) {
	w := httptest.NewRecorder()

	// Channels cannot be JSON encoded
	ch := make(chan int)
	err := JSON(w, http.StatusOK, ch)

	// Should return an error
	assert.Error(t, err)
	// Status is already written before encoding starts
	assert.Equal(t, http.StatusOK, w.Code)
}

// TestContentTypeHeader ensures Content-Type is always set correctly.
func TestContentTypeHeader(t *testing.T) {
	tests := []struct {
		name string
		fn   func(w http.ResponseWriter) error
	}{
		{
			name: "JSON",
			fn:   func(w http.ResponseWriter) error { return JSON(w, 200, nil) },
		},
		{
			name: "Error",
			fn:   func(w http.ResponseWriter) error { return Error(w, 400, "error", "code") },
		},
		{
			name: "Paginated",
			fn:   func(w http.ResponseWriter) error { return Paginated(w, []string{}, 10, 0, 0) },
		},
		{
			name: "Created",
			fn:   func(w http.ResponseWriter) error { return Created(w, nil) },
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			_ = tt.fn(w)
			assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
		})
	}
}
