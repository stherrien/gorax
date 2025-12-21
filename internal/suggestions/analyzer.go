package suggestions

import (
	"context"
	"regexp"
)

// PatternAnalyzer analyzes errors using pattern matching
type PatternAnalyzer struct {
	matcher  *PatternMatcher
	tenantID string
}

// NewPatternAnalyzer creates a new pattern analyzer with default patterns
func NewPatternAnalyzer(patterns []*BuiltinPattern) *PatternAnalyzer {
	if patterns == nil {
		patterns = DefaultPatterns()
	}
	return &PatternAnalyzer{
		matcher: NewPatternMatcher(patterns),
	}
}

// NewPatternAnalyzerWithTenant creates a pattern analyzer with a tenant ID
func NewPatternAnalyzerWithTenant(patterns []*BuiltinPattern, tenantID string) *PatternAnalyzer {
	analyzer := NewPatternAnalyzer(patterns)
	analyzer.tenantID = tenantID
	return analyzer
}

// Name returns the analyzer name
func (a *PatternAnalyzer) Name() string {
	return "pattern"
}

// CanHandle returns true if this analyzer can handle the error context
func (a *PatternAnalyzer) CanHandle(errCtx *ErrorContext) bool {
	if errCtx == nil {
		return false
	}
	// Can handle if there's an error message or HTTP status
	return errCtx.ErrorMessage != "" || errCtx.HTTPStatus > 0
}

// Analyze analyzes the error context and returns suggestions
func (a *PatternAnalyzer) Analyze(ctx context.Context, errCtx *ErrorContext) ([]*Suggestion, error) {
	if !a.CanHandle(errCtx) {
		return nil, nil
	}

	matches := a.matcher.Match(errCtx)
	if len(matches) == 0 {
		return nil, nil
	}

	suggestions := make([]*Suggestion, 0, len(matches))
	for _, match := range matches {
		suggestion := match.ToSuggestion(a.tenantID, errCtx)
		suggestions = append(suggestions, suggestion)
	}

	return suggestions, nil
}

// PatternRepository interface for database pattern access
type PatternRepository interface {
	GetActivePatterns(ctx context.Context, tenantID string) ([]*ErrorPattern, error)
}

// DatabasePatternAnalyzer loads patterns from database and matches
type DatabasePatternAnalyzer struct {
	repo     PatternRepository
	tenantID string
}

// NewDatabasePatternAnalyzer creates a database pattern analyzer
func NewDatabasePatternAnalyzer(repo PatternRepository, tenantID string) *DatabasePatternAnalyzer {
	return &DatabasePatternAnalyzer{
		repo:     repo,
		tenantID: tenantID,
	}
}

// Name returns the analyzer name
func (a *DatabasePatternAnalyzer) Name() string {
	return "database_pattern"
}

// CanHandle returns true if this analyzer can handle the error context
func (a *DatabasePatternAnalyzer) CanHandle(errCtx *ErrorContext) bool {
	if errCtx == nil {
		return false
	}
	return errCtx.ErrorMessage != "" || errCtx.HTTPStatus > 0
}

// Analyze loads patterns from database and analyzes the error context
func (a *DatabasePatternAnalyzer) Analyze(ctx context.Context, errCtx *ErrorContext) ([]*Suggestion, error) {
	if !a.CanHandle(errCtx) {
		return nil, nil
	}

	// Load patterns from database
	patterns, err := a.repo.GetActivePatterns(ctx, a.tenantID)
	if err != nil {
		return nil, err
	}

	if len(patterns) == 0 {
		return nil, nil
	}

	// Convert database patterns to builtin patterns
	builtinPatterns := convertToBuiltinPatterns(patterns)

	// Create matcher and find matches
	matcher := NewPatternMatcher(builtinPatterns)
	matches := matcher.Match(errCtx)

	if len(matches) == 0 {
		return nil, nil
	}

	suggestions := make([]*Suggestion, 0, len(matches))
	for _, match := range matches {
		suggestion := match.ToSuggestion(a.tenantID, errCtx)
		suggestions = append(suggestions, suggestion)
	}

	return suggestions, nil
}

// convertToBuiltinPatterns converts database ErrorPatterns to BuiltinPatterns
func convertToBuiltinPatterns(patterns []*ErrorPattern) []*BuiltinPattern {
	result := make([]*BuiltinPattern, 0, len(patterns))

	for _, p := range patterns {
		if !p.IsActive {
			continue
		}

		builtin := &BuiltinPattern{
			Name:                  p.Name,
			Category:              p.Category,
			MessagePatterns:       p.Patterns,
			HTTPCodes:             p.HTTPCodes,
			NodeTypes:             p.NodeTypes,
			SuggestionType:        p.SuggestionType,
			SuggestionTitle:       p.SuggestionTitle,
			SuggestionDescription: p.SuggestionDescription,
			SuggestionConfidence:  p.SuggestionConfidence,
			FixTemplate:           p.FixTemplate,
			Priority:              p.Priority,
		}
		result = append(result, builtin)
	}

	return result
}

// CompositeAnalyzer combines multiple analyzers
type CompositeAnalyzer struct {
	analyzers []Analyzer
}

// NewCompositeAnalyzer creates a composite analyzer from multiple analyzers
func NewCompositeAnalyzer(analyzers ...Analyzer) *CompositeAnalyzer {
	return &CompositeAnalyzer{
		analyzers: analyzers,
	}
}

// Name returns the analyzer name
func (a *CompositeAnalyzer) Name() string {
	return "composite"
}

// CanHandle returns true if any analyzer can handle the error context
func (a *CompositeAnalyzer) CanHandle(errCtx *ErrorContext) bool {
	for _, analyzer := range a.analyzers {
		if analyzer.CanHandle(errCtx) {
			return true
		}
	}
	return false
}

// Analyze runs all analyzers and combines results
func (a *CompositeAnalyzer) Analyze(ctx context.Context, errCtx *ErrorContext) ([]*Suggestion, error) {
	var allSuggestions []*Suggestion
	seenCategories := make(map[ErrorCategory]bool)

	for _, analyzer := range a.analyzers {
		if !analyzer.CanHandle(errCtx) {
			continue
		}

		suggestions, err := analyzer.Analyze(ctx, errCtx)
		if err != nil {
			// Log error but continue with other analyzers
			continue
		}

		// Deduplicate by category across analyzers
		for _, s := range suggestions {
			if !seenCategories[s.Category] {
				allSuggestions = append(allSuggestions, s)
				seenCategories[s.Category] = true
			}
		}
	}

	return allSuggestions, nil
}

// AddAnalyzer adds an analyzer to the composite
func (a *CompositeAnalyzer) AddAnalyzer(analyzer Analyzer) {
	a.analyzers = append(a.analyzers, analyzer)
}

// AnalyzeWithDetails enriches suggestions with additional context
func AnalyzeWithDetails(suggestion *Suggestion, errCtx *ErrorContext) *Suggestion {
	// Add error message as details if not already set
	if suggestion.Details == "" && errCtx.ErrorMessage != "" {
		suggestion.Details = "Original error: " + errCtx.ErrorMessage
	}

	// Add HTTP status info if relevant
	if errCtx.HTTPStatus > 0 && suggestion.Details != "" {
		suggestion.Details += " (HTTP " + string(rune('0'+errCtx.HTTPStatus/100)) +
			string(rune('0'+(errCtx.HTTPStatus/10)%10)) +
			string(rune('0'+errCtx.HTTPStatus%10)) + ")"
	}

	return suggestion
}

// QuickAnalyze performs a quick pattern-based analysis without database
func QuickAnalyze(errCtx *ErrorContext) ([]*Suggestion, error) {
	analyzer := NewPatternAnalyzer(nil)
	return analyzer.Analyze(context.Background(), errCtx)
}

// ClassifyError classifies an error into a category
func ClassifyError(errCtx *ErrorContext) ErrorCategory {
	// First check HTTP status
	if errCtx.HTTPStatus > 0 {
		category := CategoryFromHTTPCode(errCtx.HTTPStatus)
		if category != ErrorCategoryUnknown {
			return category
		}
	}

	// Then check message patterns
	if errCtx.ErrorMessage != "" {
		return CategoryFromMessage(errCtx.ErrorMessage)
	}

	return ErrorCategoryUnknown
}

// IsRetryable determines if an error is likely retryable
func IsRetryable(errCtx *ErrorContext) bool {
	category := ClassifyError(errCtx)

	switch category {
	case ErrorCategoryNetwork, ErrorCategoryTimeout, ErrorCategoryExternal, ErrorCategoryRateLimit:
		return true
	default:
		return false
	}
}

// SuggestedRetryDelay returns a suggested retry delay based on error type
func SuggestedRetryDelay(errCtx *ErrorContext, attemptNumber int) int {
	category := ClassifyError(errCtx)

	baseDelay := 1000 // 1 second
	switch category {
	case ErrorCategoryRateLimit:
		baseDelay = 5000 // 5 seconds for rate limits
	case ErrorCategoryTimeout:
		baseDelay = 2000 // 2 seconds for timeouts
	case ErrorCategoryExternal:
		baseDelay = 3000 // 3 seconds for external errors
	}

	// Exponential backoff with max of 60 seconds
	delay := baseDelay * (1 << attemptNumber)
	if delay > 60000 {
		delay = 60000
	}

	return delay
}

// compilePatterns compiles regex patterns for matching
func compilePatterns(patterns []string) []*regexp.Regexp {
	compiled := make([]*regexp.Regexp, 0, len(patterns))
	for _, pattern := range patterns {
		re, err := regexp.Compile("(?i)" + pattern)
		if err == nil {
			compiled = append(compiled, re)
		}
	}
	return compiled
}
