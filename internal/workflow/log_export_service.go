package workflow

import (
	"context"
	"errors"
)

// LogExportService handles log export operations
type LogExportService struct {
	repo     RepositoryInterface
	exporter *LogExporter
}

// NewLogExportService creates a new log export service
func NewLogExportService(repo RepositoryInterface) *LogExportService {
	return &LogExportService{
		repo:     repo,
		exporter: NewLogExporter(),
	}
}

// GetExecutionWithSteps retrieves an execution and its step executions
func (s *LogExportService) GetExecutionWithSteps(ctx context.Context, tenantID, executionID string) (*Execution, []*StepExecution, error) {
	execution, err := s.repo.GetExecutionByID(ctx, tenantID, executionID)
	if err != nil {
		return nil, nil, err
	}

	steps, err := s.repo.GetStepExecutionsByExecutionID(ctx, executionID)
	if err != nil {
		return nil, nil, err
	}

	return execution, steps, nil
}

// ExportLogs exports logs in the specified format
func (s *LogExportService) ExportLogs(execution *Execution, steps []*StepExecution, format string) ([]byte, string, error) {
	switch format {
	case "txt":
		return s.exporter.ExportTXT(execution, steps), "text/plain", nil
	case "json":
		return s.exporter.ExportJSON(execution, steps), "application/json", nil
	case "csv":
		return s.exporter.ExportCSV(execution, steps), "text/csv", nil
	default:
		return nil, "", errors.New("unsupported format")
	}
}
