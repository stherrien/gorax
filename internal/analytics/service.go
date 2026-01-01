package analytics

import (
	"context"
	"fmt"
)

// Service handles analytics business logic
type Service struct {
	repo AnalyticsRepository
}

// AnalyticsRepository defines the interface for analytics data access
type AnalyticsRepository interface {
	GetWorkflowStats(ctx context.Context, tenantID, workflowID string, timeRange TimeRange) (*WorkflowStats, error)
	GetTenantOverview(ctx context.Context, tenantID string, timeRange TimeRange) (*TenantOverview, error)
	GetExecutionTrends(ctx context.Context, tenantID string, timeRange TimeRange, granularity Granularity) (*ExecutionTrends, error)
	GetTopWorkflows(ctx context.Context, tenantID string, timeRange TimeRange, limit int) (*TopWorkflows, error)
	GetErrorBreakdown(ctx context.Context, tenantID string, timeRange TimeRange) (*ErrorBreakdown, error)
	GetNodePerformance(ctx context.Context, tenantID, workflowID string) (*NodePerformance, error)
}

// NewService creates a new analytics service
func NewService(repo AnalyticsRepository) *Service {
	return &Service{repo: repo}
}

// GetWorkflowStats retrieves statistics for a specific workflow
func (s *Service) GetWorkflowStats(ctx context.Context, tenantID, workflowID string, timeRange TimeRange) (*WorkflowStats, error) {
	if err := validateTimeRange(timeRange); err != nil {
		return nil, err
	}

	stats, err := s.repo.GetWorkflowStats(ctx, tenantID, workflowID, timeRange)
	if err != nil {
		return nil, fmt.Errorf("get workflow stats: %w", err)
	}

	return stats, nil
}

// GetTenantOverview retrieves overall statistics for a tenant
func (s *Service) GetTenantOverview(ctx context.Context, tenantID string, timeRange TimeRange) (*TenantOverview, error) {
	if err := validateTimeRange(timeRange); err != nil {
		return nil, err
	}

	overview, err := s.repo.GetTenantOverview(ctx, tenantID, timeRange)
	if err != nil {
		return nil, fmt.Errorf("get tenant overview: %w", err)
	}

	return overview, nil
}

// GetExecutionTrends retrieves execution trends over time
func (s *Service) GetExecutionTrends(ctx context.Context, tenantID string, timeRange TimeRange, granularity Granularity) (*ExecutionTrends, error) {
	if err := validateTimeRange(timeRange); err != nil {
		return nil, err
	}

	if err := validateGranularity(granularity); err != nil {
		return nil, err
	}

	trends, err := s.repo.GetExecutionTrends(ctx, tenantID, timeRange, granularity)
	if err != nil {
		return nil, fmt.Errorf("get execution trends: %w", err)
	}

	return trends, nil
}

// GetTopWorkflows retrieves the most frequently executed workflows
func (s *Service) GetTopWorkflows(ctx context.Context, tenantID string, timeRange TimeRange, limit int) (*TopWorkflows, error) {
	if err := validateTimeRange(timeRange); err != nil {
		return nil, err
	}

	if limit <= 0 {
		limit = 10
	}

	if limit > 100 {
		limit = 100
	}

	workflows, err := s.repo.GetTopWorkflows(ctx, tenantID, timeRange, limit)
	if err != nil {
		return nil, fmt.Errorf("get top workflows: %w", err)
	}

	return workflows, nil
}

// GetErrorBreakdown retrieves error analysis
func (s *Service) GetErrorBreakdown(ctx context.Context, tenantID string, timeRange TimeRange) (*ErrorBreakdown, error) {
	if err := validateTimeRange(timeRange); err != nil {
		return nil, err
	}

	breakdown, err := s.repo.GetErrorBreakdown(ctx, tenantID, timeRange)
	if err != nil {
		return nil, fmt.Errorf("get error breakdown: %w", err)
	}

	return breakdown, nil
}

// GetNodePerformance retrieves node-level performance statistics
func (s *Service) GetNodePerformance(ctx context.Context, tenantID, workflowID string) (*NodePerformance, error) {
	performance, err := s.repo.GetNodePerformance(ctx, tenantID, workflowID)
	if err != nil {
		return nil, fmt.Errorf("get node performance: %w", err)
	}

	return performance, nil
}

// validateTimeRange validates the time range
func validateTimeRange(timeRange TimeRange) error {
	if timeRange.StartDate.IsZero() || timeRange.EndDate.IsZero() {
		return fmt.Errorf("start date and end date are required")
	}

	if timeRange.EndDate.Before(timeRange.StartDate) {
		return fmt.Errorf("end date must be after start date")
	}

	return nil
}

// validateGranularity validates the granularity
func validateGranularity(granularity Granularity) error {
	switch granularity {
	case GranularityHour, GranularityDay, GranularityWeek, GranularityMonth:
		return nil
	default:
		return fmt.Errorf("invalid granularity: %s", granularity)
	}
}
