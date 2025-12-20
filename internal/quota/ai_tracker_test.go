package quota

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gorax/gorax/internal/llm"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAITracker_TrackUsage(t *testing.T) {
	t.Run("successful tracking", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		sqlxDB := sqlx.NewDb(db, "postgres")
		tracker := NewAITracker(sqlxDB)

		mock.ExpectExec(`INSERT INTO ai_usage_log`).
			WithArgs(
				"tenant-123",
				"cred-456",
				"openai",
				"gpt-4o",
				"chat_completion",
				sqlmock.AnyArg(), // execution_id
				sqlmock.AnyArg(), // workflow_id
				100,              // prompt_tokens
				50,               // completion_tokens
				150,              // total_tokens
				sqlmock.AnyArg(), // estimated_cost_cents
				true,             // success
				sqlmock.AnyArg(), // error_code
				sqlmock.AnyArg(), // error_message
				250,              // latency_ms
				sqlmock.AnyArg(), // request_metadata
			).
			WillReturnResult(sqlmock.NewResult(1, 1))

		err = tracker.TrackUsage(context.Background(), &AIUsageLog{
			TenantID:     "tenant-123",
			CredentialID: "cred-456",
			Provider:     "openai",
			Model:        "gpt-4o",
			ActionType:   "chat_completion",
			Usage: llm.TokenUsage{
				PromptTokens:     100,
				CompletionTokens: 50,
				TotalTokens:      150,
			},
			Success:   true,
			LatencyMS: 250,
		})

		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("track failure with error", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		sqlxDB := sqlx.NewDb(db, "postgres")
		tracker := NewAITracker(sqlxDB)

		mock.ExpectExec(`INSERT INTO ai_usage_log`).
			WithArgs(
				"tenant-123",
				"cred-456",
				"openai",
				"gpt-4o",
				"chat_completion",
				sqlmock.AnyArg(), // execution_id
				sqlmock.AnyArg(), // workflow_id
				100,              // prompt_tokens (estimated)
				0,                // completion_tokens
				100,              // total_tokens
				sqlmock.AnyArg(), // estimated_cost_cents
				false,            // success
				"rate_limit",     // error_code
				"Rate limit exceeded",
				0,                // latency_ms
				sqlmock.AnyArg(), // request_metadata
			).
			WillReturnResult(sqlmock.NewResult(1, 1))

		err = tracker.TrackUsage(context.Background(), &AIUsageLog{
			TenantID:     "tenant-123",
			CredentialID: "cred-456",
			Provider:     "openai",
			Model:        "gpt-4o",
			ActionType:   "chat_completion",
			Usage: llm.TokenUsage{
				PromptTokens: 100,
				TotalTokens:  100,
			},
			Success:      false,
			ErrorCode:    "rate_limit",
			ErrorMessage: "Rate limit exceeded",
		})

		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("with execution context", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		sqlxDB := sqlx.NewDb(db, "postgres")
		tracker := NewAITracker(sqlxDB)

		mock.ExpectExec(`INSERT INTO ai_usage_log`).
			WithArgs(
				"tenant-123",
				"cred-456",
				"anthropic",
				"claude-3-sonnet-20240229",
				"summarization",
				"exec-789",        // execution_id
				"workflow-abc",    // workflow_id
				200,               // prompt_tokens
				80,                // completion_tokens
				280,               // total_tokens
				sqlmock.AnyArg(),  // estimated_cost_cents
				true,              // success
				sqlmock.AnyArg(),  // error_code
				sqlmock.AnyArg(),  // error_message
				500,               // latency_ms
				sqlmock.AnyArg(),  // request_metadata
			).
			WillReturnResult(sqlmock.NewResult(1, 1))

		err = tracker.TrackUsage(context.Background(), &AIUsageLog{
			TenantID:     "tenant-123",
			CredentialID: "cred-456",
			Provider:     "anthropic",
			Model:        "claude-3-sonnet-20240229",
			ActionType:   "summarization",
			ExecutionID:  "exec-789",
			WorkflowID:   "workflow-abc",
			Usage: llm.TokenUsage{
				PromptTokens:     200,
				CompletionTokens: 80,
				TotalTokens:      280,
			},
			Success:   true,
			LatencyMS: 500,
		})

		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestAITracker_GetUsage(t *testing.T) {
	t.Run("get monthly usage summary", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		sqlxDB := sqlx.NewDb(db, "postgres")
		tracker := NewAITracker(sqlxDB)

		rows := sqlmock.NewRows([]string{
			"provider", "model", "request_count",
			"total_prompt_tokens", "total_completion_tokens", "total_tokens", "total_cost_cents",
		}).
			AddRow("openai", "gpt-4o", 100, 10000, 5000, 15000, 750).
			AddRow("anthropic", "claude-3-sonnet-20240229", 50, 8000, 3000, 11000, 500)

		from := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
		to := time.Date(2024, 1, 31, 23, 59, 59, 0, time.UTC)

		mock.ExpectQuery(`SELECT provider, model`).
			WithArgs("tenant-123", from, to).
			WillReturnRows(rows)

		summary, err := tracker.GetUsage(context.Background(), "tenant-123", from, to)

		require.NoError(t, err)
		require.NotNil(t, summary)
		assert.Len(t, summary.ByModel, 2)
		assert.Equal(t, int64(150), summary.TotalRequests)
		assert.Equal(t, int64(18000), summary.TotalPromptTokens)
		assert.Equal(t, int64(8000), summary.TotalCompletionTokens)
		assert.Equal(t, int64(26000), summary.TotalTokens)
		assert.Equal(t, int64(1250), summary.TotalCostCents)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("empty usage", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		sqlxDB := sqlx.NewDb(db, "postgres")
		tracker := NewAITracker(sqlxDB)

		rows := sqlmock.NewRows([]string{
			"provider", "model", "request_count",
			"total_prompt_tokens", "total_completion_tokens", "total_tokens", "total_cost_cents",
		})

		from := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
		to := time.Date(2024, 1, 31, 23, 59, 59, 0, time.UTC)

		mock.ExpectQuery(`SELECT provider, model`).
			WithArgs("tenant-123", from, to).
			WillReturnRows(rows)

		summary, err := tracker.GetUsage(context.Background(), "tenant-123", from, to)

		require.NoError(t, err)
		require.NotNil(t, summary)
		assert.Empty(t, summary.ByModel)
		assert.Equal(t, int64(0), summary.TotalRequests)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestAITracker_EstimateCost(t *testing.T) {
	tracker := NewAITracker(nil)

	tests := []struct {
		name        string
		provider    string
		model       string
		usage       llm.TokenUsage
		expectedMin int // minimum expected cost in cents
		expectedMax int // maximum expected cost in cents
	}{
		{
			name:        "OpenAI GPT-4o",
			provider:    "openai",
			model:       "gpt-4o",
			usage:       llm.TokenUsage{PromptTokens: 1000, CompletionTokens: 500, TotalTokens: 1500},
			expectedMin: 0, // Small usage, might round to 0
			expectedMax: 10,
		},
		{
			name:        "Anthropic Claude Sonnet",
			provider:    "anthropic",
			model:       "claude-3-sonnet-20240229",
			usage:       llm.TokenUsage{PromptTokens: 1000, CompletionTokens: 500, TotalTokens: 1500},
			expectedMin: 0,
			expectedMax: 10,
		},
		{
			name:        "Bedrock Claude",
			provider:    "bedrock",
			model:       "anthropic.claude-3-sonnet-20240229-v1:0",
			usage:       llm.TokenUsage{PromptTokens: 1000, CompletionTokens: 500, TotalTokens: 1500},
			expectedMin: 0,
			expectedMax: 10,
		},
		{
			name:        "Unknown model uses default",
			provider:    "unknown",
			model:       "unknown-model",
			usage:       llm.TokenUsage{PromptTokens: 1000, CompletionTokens: 500, TotalTokens: 1500},
			expectedMin: 0,
			expectedMax: 10,
		},
		{
			name:        "Large volume calculation",
			provider:    "openai",
			model:       "gpt-4o",
			usage:       llm.TokenUsage{PromptTokens: 1000000, CompletionTokens: 500000, TotalTokens: 1500000},
			expectedMin: 500,
			expectedMax: 2000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cost := tracker.EstimateCost(tt.provider, tt.model, tt.usage)
			assert.GreaterOrEqual(t, cost, tt.expectedMin)
			assert.LessOrEqual(t, cost, tt.expectedMax)
		})
	}
}

func TestAIUsageLog_Validate(t *testing.T) {
	t.Run("valid log", func(t *testing.T) {
		log := &AIUsageLog{
			TenantID:   "tenant-123",
			Provider:   "openai",
			Model:      "gpt-4o",
			ActionType: "chat_completion",
		}
		err := log.Validate()
		assert.NoError(t, err)
	})

	t.Run("missing tenant_id", func(t *testing.T) {
		log := &AIUsageLog{
			Provider:   "openai",
			Model:      "gpt-4o",
			ActionType: "chat_completion",
		}
		err := log.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "tenant_id")
	})

	t.Run("missing provider", func(t *testing.T) {
		log := &AIUsageLog{
			TenantID:   "tenant-123",
			Model:      "gpt-4o",
			ActionType: "chat_completion",
		}
		err := log.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "provider")
	})

	t.Run("missing model", func(t *testing.T) {
		log := &AIUsageLog{
			TenantID:   "tenant-123",
			Provider:   "openai",
			ActionType: "chat_completion",
		}
		err := log.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "model")
	})

	t.Run("missing action_type", func(t *testing.T) {
		log := &AIUsageLog{
			TenantID: "tenant-123",
			Provider: "openai",
			Model:    "gpt-4o",
		}
		err := log.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "action_type")
	})
}
