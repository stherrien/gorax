package suggestions

import (
	"context"
	"log/slog"
)

// SuggestionServiceConfig holds configuration for the suggestion service
type SuggestionServiceConfig struct {
	// Repository for storing suggestions
	Repository Repository

	// PatternAnalyzer for pattern-based analysis
	PatternAnalyzer Analyzer

	// LLMAnalyzer for LLM-based analysis (optional)
	LLMAnalyzer Analyzer

	// UseLLMForUnmatched enables LLM fallback when patterns don't match
	UseLLMForUnmatched bool

	// Logger for service logging
	Logger *slog.Logger
}

// SuggestionService orchestrates error analysis and suggestion management
type SuggestionService struct {
	repo               Repository
	patternAnalyzer    Analyzer
	llmAnalyzer        Analyzer
	useLLMForUnmatched bool
	logger             *slog.Logger
}

// NewSuggestionService creates a new suggestion service
func NewSuggestionService(config SuggestionServiceConfig) *SuggestionService {
	logger := config.Logger
	if logger == nil {
		logger = slog.Default()
	}

	return &SuggestionService{
		repo:               config.Repository,
		patternAnalyzer:    config.PatternAnalyzer,
		llmAnalyzer:        config.LLMAnalyzer,
		useLLMForUnmatched: config.UseLLMForUnmatched,
		logger:             logger,
	}
}

// AnalyzeError analyzes an execution error and generates suggestions
func (s *SuggestionService) AnalyzeError(ctx context.Context, tenantID string, errCtx *ErrorContext) ([]*Suggestion, error) {
	var allSuggestions []*Suggestion
	seenCategories := make(map[ErrorCategory]bool)

	// First try pattern-based analysis
	if s.patternAnalyzer != nil && s.patternAnalyzer.CanHandle(errCtx) {
		patternSuggestions, err := s.patternAnalyzer.Analyze(ctx, errCtx)
		if err != nil {
			s.logger.Warn("pattern analysis failed", "error", err, "execution_id", errCtx.ExecutionID)
		} else {
			for _, sugg := range patternSuggestions {
				sugg.TenantID = tenantID
				if !seenCategories[sugg.Category] {
					allSuggestions = append(allSuggestions, sugg)
					seenCategories[sugg.Category] = true
				}
			}
		}
	}

	// Use LLM if patterns didn't match or as supplement
	if s.llmAnalyzer != nil && s.useLLMForUnmatched {
		shouldUseLLM := len(allSuggestions) == 0 || s.hasLowConfidenceOnly(allSuggestions)

		if shouldUseLLM && s.llmAnalyzer.CanHandle(errCtx) {
			llmSuggestions, err := s.llmAnalyzer.Analyze(ctx, errCtx)
			if err != nil {
				s.logger.Warn("LLM analysis failed", "error", err, "execution_id", errCtx.ExecutionID)
			} else {
				for _, sugg := range llmSuggestions {
					sugg.TenantID = tenantID
					if !seenCategories[sugg.Category] {
						allSuggestions = append(allSuggestions, sugg)
						seenCategories[sugg.Category] = true
					}
				}
			}
		}
	}

	// Store suggestions in repository
	if len(allSuggestions) > 0 && s.repo != nil {
		if err := s.repo.CreateBatch(ctx, allSuggestions); err != nil {
			s.logger.Error("failed to store suggestions", "error", err, "execution_id", errCtx.ExecutionID)
			// Don't fail the analysis, just log the error
		}
	}

	return allSuggestions, nil
}

// ReanalyzeError deletes existing suggestions and reanalyzes
func (s *SuggestionService) ReanalyzeError(ctx context.Context, tenantID string, errCtx *ErrorContext) ([]*Suggestion, error) {
	// Delete existing suggestions for this execution
	if s.repo != nil {
		if err := s.repo.DeleteByExecutionID(ctx, tenantID, errCtx.ExecutionID); err != nil {
			s.logger.Warn("failed to delete old suggestions", "error", err, "execution_id", errCtx.ExecutionID)
		}
	}

	// Analyze fresh
	return s.AnalyzeError(ctx, tenantID, errCtx)
}

// GetSuggestions retrieves suggestions for an execution
func (s *SuggestionService) GetSuggestions(ctx context.Context, tenantID, executionID string) ([]*Suggestion, error) {
	if s.repo == nil {
		return nil, nil
	}
	return s.repo.GetByExecutionID(ctx, tenantID, executionID)
}

// GetSuggestionByID retrieves a single suggestion
func (s *SuggestionService) GetSuggestionByID(ctx context.Context, tenantID, suggestionID string) (*Suggestion, error) {
	if s.repo == nil {
		return nil, ErrSuggestionNotFound
	}
	return s.repo.GetByID(ctx, tenantID, suggestionID)
}

// ApplySuggestion marks a suggestion as applied
func (s *SuggestionService) ApplySuggestion(ctx context.Context, tenantID, suggestionID string) error {
	if s.repo == nil {
		return ErrSuggestionNotFound
	}
	return s.repo.UpdateStatus(ctx, tenantID, suggestionID, StatusApplied)
}

// DismissSuggestion marks a suggestion as dismissed
func (s *SuggestionService) DismissSuggestion(ctx context.Context, tenantID, suggestionID string) error {
	if s.repo == nil {
		return ErrSuggestionNotFound
	}
	return s.repo.UpdateStatus(ctx, tenantID, suggestionID, StatusDismissed)
}

// DeleteSuggestion deletes a suggestion
func (s *SuggestionService) DeleteSuggestion(ctx context.Context, tenantID, suggestionID string) error {
	if s.repo == nil {
		return ErrSuggestionNotFound
	}
	return s.repo.Delete(ctx, tenantID, suggestionID)
}

// DeleteByExecutionID deletes all suggestions for an execution
func (s *SuggestionService) DeleteByExecutionID(ctx context.Context, tenantID, executionID string) error {
	if s.repo == nil {
		return nil
	}
	return s.repo.DeleteByExecutionID(ctx, tenantID, executionID)
}

func (s *SuggestionService) hasLowConfidenceOnly(suggestions []*Suggestion) bool {
	for _, sugg := range suggestions {
		if sugg.Confidence == ConfidenceHigh || sugg.Confidence == ConfidenceMedium {
			return false
		}
	}
	return true
}

// GetPendingSuggestions returns only pending suggestions for an execution
func (s *SuggestionService) GetPendingSuggestions(ctx context.Context, tenantID, executionID string) ([]*Suggestion, error) {
	all, err := s.GetSuggestions(ctx, tenantID, executionID)
	if err != nil {
		return nil, err
	}

	var pending []*Suggestion
	for _, sugg := range all {
		if sugg.Status == StatusPending {
			pending = append(pending, sugg)
		}
	}

	return pending, nil
}

// GetSuggestionStats returns statistics about suggestions for an execution
type SuggestionStats struct {
	Total     int            `json:"total"`
	Pending   int            `json:"pending"`
	Applied   int            `json:"applied"`
	Dismissed int            `json:"dismissed"`
	BySource  map[string]int `json:"by_source"`
	ByConfidence map[string]int `json:"by_confidence"`
}

// GetStats returns statistics about suggestions for an execution
func (s *SuggestionService) GetStats(ctx context.Context, tenantID, executionID string) (*SuggestionStats, error) {
	suggestions, err := s.GetSuggestions(ctx, tenantID, executionID)
	if err != nil {
		return nil, err
	}

	stats := &SuggestionStats{
		Total:        len(suggestions),
		BySource:     make(map[string]int),
		ByConfidence: make(map[string]int),
	}

	for _, sugg := range suggestions {
		switch sugg.Status {
		case StatusPending:
			stats.Pending++
		case StatusApplied:
			stats.Applied++
		case StatusDismissed:
			stats.Dismissed++
		}

		stats.BySource[string(sugg.Source)]++
		stats.ByConfidence[string(sugg.Confidence)]++
	}

	return stats, nil
}

// QuickService creates a simple service with pattern analysis only
func QuickService(repo Repository) *SuggestionService {
	return NewSuggestionService(SuggestionServiceConfig{
		Repository:      repo,
		PatternAnalyzer: NewPatternAnalyzer(nil),
	})
}

// FullService creates a service with both pattern and LLM analysis
func FullService(repo Repository, llmAnalyzer Analyzer) *SuggestionService {
	return NewSuggestionService(SuggestionServiceConfig{
		Repository:         repo,
		PatternAnalyzer:    NewPatternAnalyzer(nil),
		LLMAnalyzer:        llmAnalyzer,
		UseLLMForUnmatched: true,
	})
}
