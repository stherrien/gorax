package eventtypes

import (
	"context"
	"log/slog"
)

// Service provides event type business logic
type Service struct {
	repo   *Repository
	logger *slog.Logger
}

// NewService creates a new event type service
func NewService(repo *Repository, logger *slog.Logger) *Service {
	return &Service{
		repo:   repo,
		logger: logger,
	}
}

// ListEventTypes returns all event types from the registry
func (s *Service) ListEventTypes(ctx context.Context) ([]EventType, error) {
	return s.repo.ListAll(ctx)
}
