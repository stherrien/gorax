package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/gorax/gorax/internal/api/middleware"
	"github.com/gorax/gorax/internal/quota"
	"github.com/gorax/gorax/internal/tenant"
)

// UsageService defines the interface for usage operations
type UsageService interface {
	GetCurrentUsage(ctx context.Context, tenantID string) (*UsageResponse, error)
	GetUsageHistory(ctx context.Context, tenantID string, startDate, endDate time.Time, page, limit int) ([]quota.UsageByDate, int, error)
	GetRateLimitHits(ctx context.Context, tenantID string, period quota.Period) (int64, error)
}

// UsageHandler handles usage-related HTTP requests
type UsageHandler struct {
	service UsageService
	logger  *slog.Logger
}

// NewUsageHandler creates a new usage handler
func NewUsageHandler(service UsageService) *UsageHandler {
	return &UsageHandler{
		service: service,
		logger:  slog.Default(),
	}
}

// UsageResponse represents current usage statistics
type UsageResponse struct {
	TenantID      string        `json:"tenant_id"`
	CurrentPeriod PeriodUsage   `json:"current_period"`
	MonthToDate   PeriodUsage   `json:"month_to_date"`
	Quotas        QuotaInfo     `json:"quotas"`
	RateLimits    RateLimitInfo `json:"rate_limits"`
}

// PeriodUsage represents usage for a specific period
type PeriodUsage struct {
	WorkflowExecutions int64  `json:"workflow_executions"`
	StepExecutions     int64  `json:"step_executions"`
	Period             string `json:"period"`
}

// QuotaInfo represents quota information
type QuotaInfo struct {
	MaxExecutionsPerDay     int     `json:"max_executions_per_day"`
	MaxExecutionsPerMonth   int     `json:"max_executions_per_month"`
	ExecutionsRemaining     int64   `json:"executions_remaining"`
	QuotaPercentUsed        float64 `json:"quota_percent_used"`
	MaxConcurrentExecutions int     `json:"max_concurrent_executions"`
	MaxWorkflows            int     `json:"max_workflows"`
}

// RateLimitInfo represents rate limit information
type RateLimitInfo struct {
	RequestsPerMinute int   `json:"requests_per_minute"`
	RequestsPerHour   int   `json:"requests_per_hour"`
	RequestsPerDay    int   `json:"requests_per_day"`
	HitsToday         int64 `json:"hits_today"`
}

// UsageHistoryResponse represents historical usage data
type UsageHistoryResponse struct {
	Usage     []quota.UsageByDate `json:"usage"`
	Total     int                 `json:"total"`
	Page      int                 `json:"page"`
	Limit     int                 `json:"limit"`
	StartDate string              `json:"start_date"`
	EndDate   string              `json:"end_date"`
}

// GetCurrentUsage returns current usage statistics for a tenant
func (h *UsageHandler) GetCurrentUsage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get tenant from context
	t := middleware.GetTenant(r)
	if t == nil {
		http.Error(w, "tenant not found in context", http.StatusInternalServerError)
		return
	}

	tenantID := chi.URLParam(r, "id")
	if tenantID != t.ID {
		http.Error(w, "unauthorized", http.StatusForbidden)
		return
	}

	usage, err := h.service.GetCurrentUsage(ctx, tenantID)
	if err != nil {
		h.logger.Error("failed to get current usage",
			"error", err,
			"tenant_id", tenantID,
		)
		http.Error(w, "failed to get usage", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(usage)
}

// GetUsageHistory returns historical usage data for a tenant
func (h *UsageHandler) GetUsageHistory(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get tenant from context
	t := middleware.GetTenant(r)
	if t == nil {
		http.Error(w, "tenant not found in context", http.StatusInternalServerError)
		return
	}

	tenantID := chi.URLParam(r, "id")
	if tenantID != t.ID {
		http.Error(w, "unauthorized", http.StatusForbidden)
		return
	}

	// Parse query parameters
	startDateStr := r.URL.Query().Get("start_date")
	endDateStr := r.URL.Query().Get("end_date")

	startDate, endDate, err := parseDateRange(startDateStr, endDateStr)
	if err != nil {
		http.Error(w, fmt.Sprintf("invalid date format: %v", err), http.StatusBadRequest)
		return
	}

	page := parseIntParam(r.URL.Query().Get("page"), 1)
	limit := parseIntParam(r.URL.Query().Get("limit"), 30)

	// Validate limits
	if limit > 100 {
		limit = 100
	}
	if limit < 1 {
		limit = 30
	}
	if page < 1 {
		page = 1
	}

	usage, total, err := h.service.GetUsageHistory(ctx, tenantID, startDate, endDate, page, limit)
	if err != nil {
		h.logger.Error("failed to get usage history",
			"error", err,
			"tenant_id", tenantID,
		)
		http.Error(w, "failed to get usage history", http.StatusInternalServerError)
		return
	}

	response := UsageHistoryResponse{
		Usage:     usage,
		Total:     total,
		Page:      page,
		Limit:     limit,
		StartDate: startDate.Format("2006-01-02"),
		EndDate:   endDate.Format("2006-01-02"),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// parseDateRange parses start and end date strings
func parseDateRange(startStr, endStr string) (time.Time, time.Time, error) {
	var startDate, endDate time.Time
	var err error

	if startStr == "" {
		// Default to last 30 days
		endDate = time.Now()
		startDate = endDate.AddDate(0, 0, -30)
	} else {
		startDate, err = time.Parse("2006-01-02", startStr)
		if err != nil {
			return time.Time{}, time.Time{}, fmt.Errorf("invalid start_date format: %w", err)
		}

		if endStr == "" {
			endDate = time.Now()
		} else {
			endDate, err = time.Parse("2006-01-02", endStr)
			if err != nil {
				return time.Time{}, time.Time{}, fmt.Errorf("invalid end_date format: %w", err)
			}
		}
	}

	return startDate, endDate, nil
}

// parseIntParam parses an integer parameter with a default value
func parseIntParam(s string, defaultValue int) int {
	if s == "" {
		return defaultValue
	}

	val, err := strconv.Atoi(s)
	if err != nil {
		return defaultValue
	}

	return val
}

// UsageServiceImpl implements UsageService
type UsageServiceImpl struct {
	tracker       *quota.Tracker
	tenantService *tenant.Service
	logger        *slog.Logger
}

// NewUsageService creates a new usage service
func NewUsageService(tracker *quota.Tracker, tenantService *tenant.Service, logger *slog.Logger) *UsageServiceImpl {
	return &UsageServiceImpl{
		tracker:       tracker,
		tenantService: tenantService,
		logger:        logger,
	}
}

// GetCurrentUsage returns current usage statistics
func (s *UsageServiceImpl) GetCurrentUsage(ctx context.Context, tenantID string) (*UsageResponse, error) {
	// Get tenant to retrieve quotas
	t, err := s.tenantService.GetByID(ctx, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get tenant: %w", err)
	}

	// Parse quotas
	var quotas tenant.TenantQuotas
	if err := json.Unmarshal(t.Quotas, &quotas); err != nil {
		return nil, fmt.Errorf("failed to parse quotas: %w", err)
	}

	// Get current usage
	dailyWorkflow, err := s.tracker.GetWorkflowExecutions(ctx, tenantID, quota.PeriodDaily)
	if err != nil {
		return nil, fmt.Errorf("failed to get daily workflow executions: %w", err)
	}

	dailySteps, err := s.tracker.GetStepExecutions(ctx, tenantID, quota.PeriodDaily)
	if err != nil {
		return nil, fmt.Errorf("failed to get daily step executions: %w", err)
	}

	monthlyWorkflow, err := s.tracker.GetWorkflowExecutions(ctx, tenantID, quota.PeriodMonthly)
	if err != nil {
		return nil, fmt.Errorf("failed to get monthly workflow executions: %w", err)
	}

	monthlySteps, err := s.tracker.GetStepExecutions(ctx, tenantID, quota.PeriodMonthly)
	if err != nil {
		return nil, fmt.Errorf("failed to get monthly step executions: %w", err)
	}

	// Calculate remaining quota
	var remaining int64
	var percentUsed float64

	if quotas.MaxExecutionsPerDay == -1 {
		remaining = -1
		percentUsed = 0.0
	} else {
		remaining = int64(quotas.MaxExecutionsPerDay) - dailyWorkflow
		if remaining < 0 {
			remaining = 0
		}
		percentUsed = (float64(dailyWorkflow) / float64(quotas.MaxExecutionsPerDay)) * 100
	}

	// Get rate limit hits
	rateLimitHits, err := s.GetRateLimitHits(ctx, tenantID, quota.PeriodDaily)
	if err != nil {
		s.logger.Warn("failed to get rate limit hits", "error", err)
		rateLimitHits = 0
	}

	return &UsageResponse{
		TenantID: tenantID,
		CurrentPeriod: PeriodUsage{
			WorkflowExecutions: dailyWorkflow,
			StepExecutions:     dailySteps,
			Period:             "daily",
		},
		MonthToDate: PeriodUsage{
			WorkflowExecutions: monthlyWorkflow,
			StepExecutions:     monthlySteps,
			Period:             "monthly",
		},
		Quotas: QuotaInfo{
			MaxExecutionsPerDay:     quotas.MaxExecutionsPerDay,
			MaxExecutionsPerMonth:   -1, // Not currently tracked
			ExecutionsRemaining:     remaining,
			QuotaPercentUsed:        percentUsed,
			MaxConcurrentExecutions: quotas.MaxConcurrentExecutions,
			MaxWorkflows:            quotas.MaxWorkflows,
		},
		RateLimits: RateLimitInfo{
			RequestsPerMinute: quotas.MaxAPICallsPerMinute,
			RequestsPerHour:   -1, // Not currently set
			RequestsPerDay:    -1, // Not currently set
			HitsToday:         rateLimitHits,
		},
	}, nil
}

// GetUsageHistory returns historical usage data
func (s *UsageServiceImpl) GetUsageHistory(ctx context.Context, tenantID string, startDate, endDate time.Time, page, limit int) ([]quota.UsageByDate, int, error) {
	usage, err := s.tracker.GetUsageByDateRange(ctx, tenantID, startDate, endDate)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get usage history: %w", err)
	}

	// Apply pagination
	total := len(usage)
	start := (page - 1) * limit
	end := start + limit

	if start >= total {
		return []quota.UsageByDate{}, total, nil
	}

	if end > total {
		end = total
	}

	return usage[start:end], total, nil
}

// GetRateLimitHits returns rate limit hits for a period
func (s *UsageServiceImpl) GetRateLimitHits(ctx context.Context, tenantID string, period quota.Period) (int64, error) {
	// This would require additional tracking in the rate limiter
	// For now, return 0
	return 0, nil
}
