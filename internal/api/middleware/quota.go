package middleware

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/gorax/gorax/internal/tenant"
)

// QuotaTenantService defines the interface for tenant service operations used by quota checker
type QuotaTenantService interface {
	GetWorkflowCount(ctx context.Context, tenantID string) (int, error)
	GetConcurrentExecutions(ctx context.Context, tenantID string) (int, error)
}

// QuotaRedisClient defines the interface for Redis operations used by quota checker
type QuotaRedisClient interface {
	Get(ctx context.Context, key string) *redis.StringCmd
	Incr(ctx context.Context, key string) *redis.IntCmd
	Expire(ctx context.Context, key string, expiration time.Duration) *redis.BoolCmd
	Pipeline() redis.Pipeliner
}

// QuotaChecker handles tenant quota validation
type QuotaChecker struct {
	tenantService QuotaTenantService
	redis         QuotaRedisClient
	logger        *slog.Logger
}

// NewQuotaChecker creates a new quota checker
func NewQuotaChecker(tenantService *tenant.Service, redis *redis.Client, logger *slog.Logger) *QuotaChecker {
	return &QuotaChecker{
		tenantService: tenantService,
		redis:         redis,
		logger:        logger,
	}
}

// CheckQuotas returns middleware that validates tenant quotas before allowing operations
func (qc *QuotaChecker) CheckQuotas() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get tenant from context
			t := GetTenant(r)
			if t == nil {
				http.Error(w, "tenant not found in context", http.StatusInternalServerError)
				return
			}

			// Parse quotas from tenant
			var quotas tenant.TenantQuotas
			if err := json.Unmarshal(t.Quotas, &quotas); err != nil {
				qc.logger.Error("failed to parse tenant quotas", "error", err, "tenant_id", t.ID)
				http.Error(w, "failed to parse quotas", http.StatusInternalServerError)
				return
			}

			// Check which operation is being performed
			operation := qc.detectOperation(r)

			// Validate quotas based on operation
			switch operation {
			case "create_workflow":
				if err := qc.checkWorkflowQuota(r.Context(), t.ID, quotas); err != nil {
					qc.handleQuotaExceeded(w, err)
					return
				}
			case "execute_workflow":
				if err := qc.checkExecutionQuota(r.Context(), t.ID, quotas); err != nil {
					qc.handleQuotaExceeded(w, err)
					return
				}
			case "api_call":
				if err := qc.checkAPIRateLimit(r.Context(), t.ID, quotas); err != nil {
					qc.handleQuotaExceeded(w, err)
					return
				}
			}

			// Track API call
			qc.trackAPICall(r.Context(), t.ID)

			next.ServeHTTP(w, r)
		})
	}
}

// detectOperation determines what operation is being performed based on the request
func (qc *QuotaChecker) detectOperation(r *http.Request) string {
	path := r.URL.Path
	method := r.Method

	if method == "POST" && strings.Contains(path, "/workflows") && !strings.Contains(path, "/execute") {
		return "create_workflow"
	}
	if method == "POST" && strings.Contains(path, "/execute") {
		return "execute_workflow"
	}
	return "api_call"
}

// checkWorkflowQuota validates if tenant can create more workflows
func (qc *QuotaChecker) checkWorkflowQuota(ctx context.Context, tenantID string, quotas tenant.TenantQuotas) error {
	// -1 means unlimited
	if quotas.MaxWorkflows == -1 {
		return nil
	}

	// Get current workflow count from database
	count, err := qc.tenantService.GetWorkflowCount(ctx, tenantID)
	if err != nil {
		qc.logger.Error("failed to get workflow count", "error", err, "tenant_id", tenantID)
		return fmt.Errorf("failed to check quota: %w", err)
	}

	if count >= quotas.MaxWorkflows {
		return fmt.Errorf("workflow quota exceeded: %d/%d workflows used", count, quotas.MaxWorkflows)
	}

	return nil
}

// checkExecutionQuota validates if tenant can execute more workflows today
func (qc *QuotaChecker) checkExecutionQuota(ctx context.Context, tenantID string, quotas tenant.TenantQuotas) error {
	// -1 means unlimited
	if quotas.MaxExecutionsPerDay == -1 {
		return nil
	}

	// Check concurrent executions
	if quotas.MaxConcurrentExecutions > 0 {
		concurrent, err := qc.getConcurrentExecutions(ctx, tenantID)
		if err != nil {
			qc.logger.Error("failed to get concurrent executions", "error", err, "tenant_id", tenantID)
		} else if concurrent >= quotas.MaxConcurrentExecutions {
			return fmt.Errorf("concurrent execution quota exceeded: %d/%d concurrent executions", concurrent, quotas.MaxConcurrentExecutions)
		}
	}

	// Get today's execution count from Redis
	key := fmt.Sprintf("quota:executions:daily:%s:%s", tenantID, time.Now().Format("2006-01-02"))
	count, err := qc.redis.Get(ctx, key).Int()
	if err != nil && err != redis.Nil {
		qc.logger.Error("failed to get execution count from Redis", "error", err, "tenant_id", tenantID)
		return fmt.Errorf("failed to check quota: %w", err)
	}

	if count >= quotas.MaxExecutionsPerDay {
		return fmt.Errorf("daily execution quota exceeded: %d/%d executions used today", count, quotas.MaxExecutionsPerDay)
	}

	// Increment counter with expiration (48 hours to handle timezone differences)
	pipe := qc.redis.Pipeline()
	pipe.Incr(ctx, key)
	pipe.Expire(ctx, key, 48*time.Hour)
	_, err = pipe.Exec(ctx)
	if err != nil {
		qc.logger.Warn("failed to increment execution counter", "error", err, "tenant_id", tenantID)
	}

	return nil
}

// checkAPIRateLimit validates API rate limits
func (qc *QuotaChecker) checkAPIRateLimit(ctx context.Context, tenantID string, quotas tenant.TenantQuotas) error {
	// -1 means unlimited
	if quotas.MaxAPICallsPerMinute == -1 {
		return nil
	}

	// Use sliding window rate limiting with Redis
	key := fmt.Sprintf("quota:api:minute:%s", tenantID)
	now := time.Now().Unix()
	windowStart := now - 60

	// Remove old entries and count current
	pipe := qc.redis.Pipeline()
	pipe.ZRemRangeByScore(ctx, key, "0", fmt.Sprintf("%d", windowStart))
	pipe.ZCard(ctx, key)
	results, err := pipe.Exec(ctx)
	if err != nil {
		qc.logger.Error("failed to check rate limit", "error", err, "tenant_id", tenantID)
		return fmt.Errorf("failed to check rate limit: %w", err)
	}

	count := results[1].(*redis.IntCmd).Val()
	if int(count) >= quotas.MaxAPICallsPerMinute {
		return fmt.Errorf("API rate limit exceeded: %d/%d calls per minute", count, quotas.MaxAPICallsPerMinute)
	}

	// Add current request
	pipe = qc.redis.Pipeline()
	pipe.ZAdd(ctx, key, redis.Z{Score: float64(now), Member: fmt.Sprintf("%d", now)})
	pipe.Expire(ctx, key, 2*time.Minute)
	_, err = pipe.Exec(ctx)
	if err != nil {
		qc.logger.Warn("failed to record API call", "error", err, "tenant_id", tenantID)
	}

	return nil
}

// getConcurrentExecutions returns the number of currently running executions
func (qc *QuotaChecker) getConcurrentExecutions(ctx context.Context, tenantID string) (int, error) {
	return qc.tenantService.GetConcurrentExecutions(ctx, tenantID)
}

// trackAPICall tracks API call for analytics (non-blocking)
func (qc *QuotaChecker) trackAPICall(ctx context.Context, tenantID string) {
	key := fmt.Sprintf("analytics:api:daily:%s:%s", tenantID, time.Now().Format("2006-01-02"))
	err := qc.redis.Incr(ctx, key).Err()
	if err != nil {
		qc.logger.Debug("failed to track API call", "error", err, "tenant_id", tenantID)
	}
	qc.redis.Expire(ctx, key, 90*24*time.Hour) // 90 days retention
}

// handleQuotaExceeded returns a 429 response with quota information
func (qc *QuotaChecker) handleQuotaExceeded(w http.ResponseWriter, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Retry-After", "3600") // Suggest retry after 1 hour
	w.WriteHeader(http.StatusTooManyRequests)

	response := map[string]interface{}{
		"error":       "quota_exceeded",
		"message":     err.Error(),
		"retry_after": 3600,
	}
	json.NewEncoder(w).Encode(response)
}

// QuotaExempt returns middleware that bypasses quota checks for specific routes
func QuotaExempt() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Mark request as quota-exempt in context
			ctx := context.WithValue(r.Context(), "quota_exempt", true)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
