package quota

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"

	"github.com/gorax/gorax/internal/llm"
)

// AIUsageLog represents a single AI usage log entry
type AIUsageLog struct {
	TenantID        string                 `json:"tenant_id" db:"tenant_id"`
	CredentialID    string                 `json:"credential_id,omitempty" db:"credential_id"`
	Provider        string                 `json:"provider" db:"provider"`
	Model           string                 `json:"model" db:"model"`
	ActionType      string                 `json:"action_type" db:"action_type"`
	ExecutionID     string                 `json:"execution_id,omitempty" db:"execution_id"`
	WorkflowID      string                 `json:"workflow_id,omitempty" db:"workflow_id"`
	Usage           llm.TokenUsage         `json:"usage"`
	Success         bool                   `json:"success" db:"success"`
	ErrorCode       string                 `json:"error_code,omitempty" db:"error_code"`
	ErrorMessage    string                 `json:"error_message,omitempty" db:"error_message"`
	LatencyMS       int                    `json:"latency_ms" db:"latency_ms"`
	RequestMetadata map[string]interface{} `json:"request_metadata,omitempty"`
}

// Validate validates the usage log entry
func (l *AIUsageLog) Validate() error {
	if l.TenantID == "" {
		return fmt.Errorf("tenant_id is required")
	}
	if l.Provider == "" {
		return fmt.Errorf("provider is required")
	}
	if l.Model == "" {
		return fmt.Errorf("model is required")
	}
	if l.ActionType == "" {
		return fmt.Errorf("action_type is required")
	}
	return nil
}

// AIUsageSummary represents aggregated AI usage statistics
type AIUsageSummary struct {
	TotalRequests         int64        `json:"total_requests"`
	TotalPromptTokens     int64        `json:"total_prompt_tokens"`
	TotalCompletionTokens int64        `json:"total_completion_tokens"`
	TotalTokens           int64        `json:"total_tokens"`
	TotalCostCents        int64        `json:"total_cost_cents"`
	ByModel               []ModelUsage `json:"by_model"`
	From                  time.Time    `json:"from"`
	To                    time.Time    `json:"to"`
}

// ModelUsage represents usage for a specific model
type ModelUsage struct {
	Provider              string `json:"provider" db:"provider"`
	Model                 string `json:"model" db:"model"`
	RequestCount          int64  `json:"request_count" db:"request_count"`
	TotalPromptTokens     int64  `json:"total_prompt_tokens" db:"total_prompt_tokens"`
	TotalCompletionTokens int64  `json:"total_completion_tokens" db:"total_completion_tokens"`
	TotalTokens           int64  `json:"total_tokens" db:"total_tokens"`
	TotalCostCents        int64  `json:"total_cost_cents" db:"total_cost_cents"`
}

// AITracker handles AI usage tracking and cost estimation
type AITracker struct {
	db *sqlx.DB
}

// NewAITracker creates a new AI usage tracker
func NewAITracker(db *sqlx.DB) *AITracker {
	return &AITracker{
		db: db,
	}
}

// TrackUsage records an AI usage log entry
func (t *AITracker) TrackUsage(ctx context.Context, log *AIUsageLog) error {
	if err := log.Validate(); err != nil {
		return fmt.Errorf("invalid usage log: %w", err)
	}

	// Estimate cost
	estimatedCost := t.EstimateCost(log.Provider, log.Model, log.Usage)

	// Convert metadata to JSON
	metadataJSON, err := json.Marshal(log.RequestMetadata)
	if err != nil {
		metadataJSON = []byte("{}")
	}

	query := `
		INSERT INTO ai_usage_log (
			tenant_id, credential_id, provider, model, action_type,
			execution_id, workflow_id,
			prompt_tokens, completion_tokens, total_tokens,
			estimated_cost_cents, success, error_code, error_message,
			latency_ms, request_metadata
		) VALUES (
			$1, $2, $3, $4, $5,
			$6, $7,
			$8, $9, $10,
			$11, $12, $13, $14,
			$15, $16
		)
	`

	// Handle nullable fields
	var credentialID, executionID, workflowID, errorCode, errorMessage sql.NullString

	if log.CredentialID != "" {
		credentialID = sql.NullString{String: log.CredentialID, Valid: true}
	}
	if log.ExecutionID != "" {
		executionID = sql.NullString{String: log.ExecutionID, Valid: true}
	}
	if log.WorkflowID != "" {
		workflowID = sql.NullString{String: log.WorkflowID, Valid: true}
	}
	if log.ErrorCode != "" {
		errorCode = sql.NullString{String: log.ErrorCode, Valid: true}
	}
	if log.ErrorMessage != "" {
		errorMessage = sql.NullString{String: log.ErrorMessage, Valid: true}
	}

	_, err = t.db.ExecContext(ctx, query,
		log.TenantID,
		credentialID,
		log.Provider,
		log.Model,
		log.ActionType,
		executionID,
		workflowID,
		log.Usage.PromptTokens,
		log.Usage.CompletionTokens,
		log.Usage.TotalTokens,
		estimatedCost,
		log.Success,
		errorCode,
		errorMessage,
		log.LatencyMS,
		metadataJSON,
	)

	if err != nil {
		return fmt.Errorf("failed to insert usage log: %w", err)
	}

	return nil
}

// GetUsage retrieves aggregated usage statistics for a tenant
func (t *AITracker) GetUsage(ctx context.Context, tenantID string, from, to time.Time) (*AIUsageSummary, error) {
	if tenantID == "" {
		return nil, ErrInvalidTenantID
	}

	query := `
		SELECT
			provider,
			model,
			COUNT(*) as request_count,
			COALESCE(SUM(prompt_tokens), 0) as total_prompt_tokens,
			COALESCE(SUM(completion_tokens), 0) as total_completion_tokens,
			COALESCE(SUM(total_tokens), 0) as total_tokens,
			COALESCE(SUM(estimated_cost_cents), 0) as total_cost_cents
		FROM ai_usage_log
		WHERE tenant_id = $1
		  AND created_at >= $2
		  AND created_at <= $3
		GROUP BY provider, model
		ORDER BY total_tokens DESC
	`

	var models []ModelUsage
	err := t.db.SelectContext(ctx, &models, query, tenantID, from, to)
	if err != nil {
		return nil, fmt.Errorf("failed to query usage: %w", err)
	}

	// Calculate totals
	summary := &AIUsageSummary{
		ByModel: models,
		From:    from,
		To:      to,
	}

	for _, m := range models {
		summary.TotalRequests += m.RequestCount
		summary.TotalPromptTokens += m.TotalPromptTokens
		summary.TotalCompletionTokens += m.TotalCompletionTokens
		summary.TotalTokens += m.TotalTokens
		summary.TotalCostCents += m.TotalCostCents
	}

	return summary, nil
}

// EstimateCost estimates the cost in USD cents for a given usage
func (t *AITracker) EstimateCost(provider, model string, usage llm.TokenUsage) int {
	pricing := getModelPricing(provider, model)

	// Calculate cost: (tokens / 1,000,000) * cost_per_million
	inputCost := int64(usage.PromptTokens) * int64(pricing.InputCostPer1M) / 1000000
	outputCost := int64(usage.CompletionTokens) * int64(pricing.OutputCostPer1M) / 1000000

	return int(inputCost + outputCost)
}

// modelPricing holds pricing info for cost estimation
type modelPricing struct {
	InputCostPer1M  int
	OutputCostPer1M int
}

// getModelPricing returns pricing for a model (fallback to reasonable defaults)
func getModelPricing(provider, model string) modelPricing {
	// Normalize model name for lookup
	normalizedModel := strings.ToLower(model)

	// OpenAI models
	if provider == "openai" || strings.HasPrefix(normalizedModel, "gpt") {
		if strings.Contains(normalizedModel, "gpt-4o-mini") {
			return modelPricing{InputCostPer1M: 15, OutputCostPer1M: 60}
		}
		if strings.Contains(normalizedModel, "gpt-4o") {
			return modelPricing{InputCostPer1M: 500, OutputCostPer1M: 1500}
		}
		if strings.Contains(normalizedModel, "gpt-4-turbo") {
			return modelPricing{InputCostPer1M: 1000, OutputCostPer1M: 3000}
		}
		if strings.Contains(normalizedModel, "gpt-4") {
			return modelPricing{InputCostPer1M: 3000, OutputCostPer1M: 6000}
		}
		if strings.Contains(normalizedModel, "gpt-3.5") {
			return modelPricing{InputCostPer1M: 50, OutputCostPer1M: 150}
		}
	}

	// Anthropic models
	if provider == "anthropic" || strings.Contains(normalizedModel, "claude") {
		if strings.Contains(normalizedModel, "opus") {
			return modelPricing{InputCostPer1M: 1500, OutputCostPer1M: 7500}
		}
		if strings.Contains(normalizedModel, "sonnet") {
			return modelPricing{InputCostPer1M: 300, OutputCostPer1M: 1500}
		}
		if strings.Contains(normalizedModel, "haiku") {
			return modelPricing{InputCostPer1M: 25, OutputCostPer1M: 125}
		}
	}

	// Bedrock models (same as above but with anthropic. prefix)
	if provider == "bedrock" {
		if strings.Contains(normalizedModel, "claude") {
			if strings.Contains(normalizedModel, "opus") {
				return modelPricing{InputCostPer1M: 1500, OutputCostPer1M: 7500}
			}
			if strings.Contains(normalizedModel, "sonnet") {
				return modelPricing{InputCostPer1M: 300, OutputCostPer1M: 1500}
			}
			if strings.Contains(normalizedModel, "haiku") {
				return modelPricing{InputCostPer1M: 25, OutputCostPer1M: 125}
			}
		}
		if strings.Contains(normalizedModel, "titan") {
			if strings.Contains(normalizedModel, "express") {
				return modelPricing{InputCostPer1M: 80, OutputCostPer1M: 240}
			}
			if strings.Contains(normalizedModel, "lite") {
				return modelPricing{InputCostPer1M: 15, OutputCostPer1M: 20}
			}
			if strings.Contains(normalizedModel, "embed") {
				return modelPricing{InputCostPer1M: 2, OutputCostPer1M: 0}
			}
		}
	}

	// Default pricing (reasonable middle ground)
	return modelPricing{InputCostPer1M: 100, OutputCostPer1M: 300}
}
