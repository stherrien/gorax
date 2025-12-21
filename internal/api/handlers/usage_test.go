package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gorax/gorax/internal/api/middleware"
	"github.com/gorax/gorax/internal/quota"
	"github.com/gorax/gorax/internal/tenant"
)

type mockUsageService struct {
	currentUsage    *UsageResponse
	usageHistory    []quota.UsageByDate
	rateLimitHits   int64
	getUsageErr     error
	getHistoryErr   error
	getRateLimitErr error
}

func (m *mockUsageService) GetCurrentUsage(ctx context.Context, tenantID string) (*UsageResponse, error) {
	if m.getUsageErr != nil {
		return nil, m.getUsageErr
	}
	return m.currentUsage, nil
}

func (m *mockUsageService) GetUsageHistory(ctx context.Context, tenantID string, startDate, endDate time.Time, page, limit int) ([]quota.UsageByDate, int, error) {
	if m.getHistoryErr != nil {
		return nil, 0, m.getHistoryErr
	}
	return m.usageHistory, len(m.usageHistory), nil
}

func (m *mockUsageService) GetRateLimitHits(ctx context.Context, tenantID string, period quota.Period) (int64, error) {
	if m.getRateLimitErr != nil {
		return 0, m.getRateLimitErr
	}
	return m.rateLimitHits, nil
}

func TestUsageHandler_GetCurrentUsage(t *testing.T) {
	tests := []struct {
		name           string
		tenantID       string
		mockUsage      *UsageResponse
		mockError      error
		expectedStatus int
		validateBody   bool
	}{
		{
			name:     "successful retrieval",
			tenantID: "tenant-1",
			mockUsage: &UsageResponse{
				TenantID: "tenant-1",
				CurrentPeriod: PeriodUsage{
					WorkflowExecutions: 50,
					StepExecutions:     200,
					Period:             "daily",
				},
				Quotas: QuotaInfo{
					MaxExecutionsPerDay: 100,
					ExecutionsRemaining: 50,
					QuotaPercentUsed:    50.0,
				},
			},
			expectedStatus: http.StatusOK,
			validateBody:   true,
		},
		{
			name:           "service error",
			tenantID:       "tenant-2",
			mockError:      assert.AnError,
			expectedStatus: http.StatusInternalServerError,
			validateBody:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &mockUsageService{
				currentUsage: tt.mockUsage,
				getUsageErr:  tt.mockError,
			}

			handler := NewUsageHandler(mockService)

			// Create request
			req := httptest.NewRequest(http.MethodGet, "/api/tenants/"+tt.tenantID+"/usage", nil)

			// Add tenant to context
			testTenant := &tenant.Tenant{
				ID:   tt.tenantID,
				Name: "Test Tenant",
			}
			ctx := context.WithValue(req.Context(), middleware.TenantContextKey, testTenant)
			req = req.WithContext(ctx)

			// Add URL params
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", tt.tenantID)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			// Record response
			rr := httptest.NewRecorder()

			// Execute
			handler.GetCurrentUsage(rr, req)

			// Assert status
			assert.Equal(t, tt.expectedStatus, rr.Code)

			// Validate body if expected
			if tt.validateBody {
				var response UsageResponse
				err := json.NewDecoder(rr.Body).Decode(&response)
				require.NoError(t, err)

				assert.Equal(t, tt.mockUsage.TenantID, response.TenantID)
				assert.Equal(t, tt.mockUsage.CurrentPeriod.WorkflowExecutions, response.CurrentPeriod.WorkflowExecutions)
				assert.Equal(t, tt.mockUsage.Quotas.MaxExecutionsPerDay, response.Quotas.MaxExecutionsPerDay)
			}
		})
	}
}

func TestUsageHandler_GetUsageHistory(t *testing.T) {
	tests := []struct {
		name           string
		tenantID       string
		queryParams    map[string]string
		mockHistory    []quota.UsageByDate
		mockError      error
		expectedStatus int
		validateBody   bool
	}{
		{
			name:     "successful retrieval with date range",
			tenantID: "tenant-1",
			queryParams: map[string]string{
				"start_date": "2024-01-01",
				"end_date":   "2024-01-31",
			},
			mockHistory: []quota.UsageByDate{
				{
					Date:               "2024-01-01",
					WorkflowExecutions: 10,
					StepExecutions:     50,
				},
				{
					Date:               "2024-01-02",
					WorkflowExecutions: 15,
					StepExecutions:     75,
				},
			},
			expectedStatus: http.StatusOK,
			validateBody:   true,
		},
		{
			name:     "default date range (last 30 days)",
			tenantID: "tenant-1",
			mockHistory: []quota.UsageByDate{
				{
					Date:               time.Now().Format("2006-01-02"),
					WorkflowExecutions: 5,
					StepExecutions:     25,
				},
			},
			expectedStatus: http.StatusOK,
			validateBody:   true,
		},
		{
			name:     "pagination parameters",
			tenantID: "tenant-1",
			queryParams: map[string]string{
				"page":  "2",
				"limit": "10",
			},
			mockHistory:    []quota.UsageByDate{},
			expectedStatus: http.StatusOK,
			validateBody:   true,
		},
		{
			name:     "invalid date format",
			tenantID: "tenant-1",
			queryParams: map[string]string{
				"start_date": "invalid",
			},
			expectedStatus: http.StatusBadRequest,
			validateBody:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &mockUsageService{
				usageHistory:  tt.mockHistory,
				getHistoryErr: tt.mockError,
			}

			handler := NewUsageHandler(mockService)

			// Build URL with query params
			url := "/api/tenants/" + tt.tenantID + "/usage/history"
			if len(tt.queryParams) > 0 {
				url += "?"
				first := true
				for k, v := range tt.queryParams {
					if !first {
						url += "&"
					}
					url += k + "=" + v
					first = false
				}
			}

			req := httptest.NewRequest(http.MethodGet, url, nil)

			// Add tenant to context
			testTenant := &tenant.Tenant{
				ID:   tt.tenantID,
				Name: "Test Tenant",
			}
			ctx := context.WithValue(req.Context(), middleware.TenantContextKey, testTenant)
			req = req.WithContext(ctx)

			// Add URL params
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", tt.tenantID)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			// Record response
			rr := httptest.NewRecorder()

			// Execute
			handler.GetUsageHistory(rr, req)

			// Assert status
			assert.Equal(t, tt.expectedStatus, rr.Code)

			// Validate body if expected
			if tt.validateBody {
				var response UsageHistoryResponse
				err := json.NewDecoder(rr.Body).Decode(&response)
				require.NoError(t, err)

				assert.Equal(t, len(tt.mockHistory), len(response.Usage))
				if len(tt.mockHistory) > 0 {
					assert.Equal(t, tt.mockHistory[0].Date, response.Usage[0].Date)
				}
			}
		})
	}
}

func TestUsageHandler_NoTenantInContext(t *testing.T) {
	mockService := &mockUsageService{}
	handler := NewUsageHandler(mockService)

	req := httptest.NewRequest(http.MethodGet, "/api/tenants/tenant-1/usage", nil)
	rr := httptest.NewRecorder()

	handler.GetCurrentUsage(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
}

func TestUsageResponse_QuotaCalculations(t *testing.T) {
	tests := []struct {
		name                string
		current             int64
		quota               int64
		expectedRemaining   int64
		expectedPercentUsed float64
	}{
		{
			name:                "half quota used",
			current:             50,
			quota:               100,
			expectedRemaining:   50,
			expectedPercentUsed: 50.0,
		},
		{
			name:                "quota exceeded",
			current:             150,
			quota:               100,
			expectedRemaining:   0,
			expectedPercentUsed: 150.0,
		},
		{
			name:                "unlimited quota",
			current:             999,
			quota:               -1,
			expectedRemaining:   -1,
			expectedPercentUsed: 0.0,
		},
		{
			name:                "zero usage",
			current:             0,
			quota:               100,
			expectedRemaining:   100,
			expectedPercentUsed: 0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			remaining := tt.quota - tt.current
			if tt.quota == -1 {
				remaining = -1
			} else if remaining < 0 {
				remaining = 0
			}

			var percentUsed float64
			if tt.quota > 0 {
				percentUsed = (float64(tt.current) / float64(tt.quota)) * 100
			}

			assert.Equal(t, tt.expectedRemaining, remaining)
			assert.InDelta(t, tt.expectedPercentUsed, percentUsed, 0.01)
		})
	}
}

func TestParseDateRange(t *testing.T) {
	tests := []struct {
		name      string
		startStr  string
		endStr    string
		wantError bool
	}{
		{
			name:      "valid date range",
			startStr:  "2024-01-01",
			endStr:    "2024-01-31",
			wantError: false,
		},
		{
			name:      "invalid start date",
			startStr:  "invalid",
			endStr:    "2024-01-31",
			wantError: true,
		},
		{
			name:      "invalid end date",
			startStr:  "2024-01-01",
			endStr:    "invalid",
			wantError: true,
		},
		{
			name:      "empty dates (should use defaults)",
			startStr:  "",
			endStr:    "",
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			start, end, err := parseDateRange(tt.startStr, tt.endStr)

			if tt.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.True(t, start.Before(end) || start.Equal(end))
			}
		})
	}
}
