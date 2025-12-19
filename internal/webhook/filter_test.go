package webhook

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestFilterEvaluator_Equals tests the equals operator
func TestFilterEvaluator_Equals(t *testing.T) {
	tests := []struct {
		name     string
		filter   *WebhookFilter
		payload  map[string]interface{}
		expected bool
		wantErr  bool
	}{
		{
			name: "string equals match",
			filter: &WebhookFilter{
				FieldPath: "$.status",
				Operator:  OpEquals,
				Value:     "active",
				Enabled:   true,
			},
			payload:  map[string]interface{}{"status": "active"},
			expected: true,
		},
		{
			name: "string equals no match",
			filter: &WebhookFilter{
				FieldPath: "$.status",
				Operator:  OpEquals,
				Value:     "active",
				Enabled:   true,
			},
			payload:  map[string]interface{}{"status": "inactive"},
			expected: false,
		},
		{
			name: "nested field equals",
			filter: &WebhookFilter{
				FieldPath: "$.data.type",
				Operator:  OpEquals,
				Value:     "issue",
				Enabled:   true,
			},
			payload: map[string]interface{}{
				"data": map[string]interface{}{
					"type": "issue",
				},
			},
			expected: true,
		},
		{
			name: "number equals match",
			filter: &WebhookFilter{
				FieldPath: "$.count",
				Operator:  OpEquals,
				Value:     42.0,
				Enabled:   true,
			},
			payload:  map[string]interface{}{"count": 42},
			expected: true,
		},
		{
			name: "boolean equals match",
			filter: &WebhookFilter{
				FieldPath: "$.enabled",
				Operator:  OpEquals,
				Value:     true,
				Enabled:   true,
			},
			payload:  map[string]interface{}{"enabled": true},
			expected: true,
		},
	}

	evaluator := NewFilterEvaluator(nil)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := evaluator.EvaluateSingle(tt.filter, tt.payload)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestFilterEvaluator_NotEquals tests the not equals operator
func TestFilterEvaluator_NotEquals(t *testing.T) {
	tests := []struct {
		name     string
		filter   *WebhookFilter
		payload  map[string]interface{}
		expected bool
	}{
		{
			name: "string not equals match",
			filter: &WebhookFilter{
				FieldPath: "$.status",
				Operator:  OpNotEquals,
				Value:     "inactive",
				Enabled:   true,
			},
			payload:  map[string]interface{}{"status": "active"},
			expected: true,
		},
		{
			name: "string not equals no match",
			filter: &WebhookFilter{
				FieldPath: "$.status",
				Operator:  OpNotEquals,
				Value:     "active",
				Enabled:   true,
			},
			payload:  map[string]interface{}{"status": "active"},
			expected: false,
		},
	}

	evaluator := NewFilterEvaluator(nil)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := evaluator.EvaluateSingle(tt.filter, tt.payload)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestFilterEvaluator_Contains tests the contains operator
func TestFilterEvaluator_Contains(t *testing.T) {
	tests := []struct {
		name     string
		filter   *WebhookFilter
		payload  map[string]interface{}
		expected bool
		wantErr  bool
	}{
		{
			name: "string contains match",
			filter: &WebhookFilter{
				FieldPath: "$.message",
				Operator:  OpContains,
				Value:     "error",
				Enabled:   true,
			},
			payload:  map[string]interface{}{"message": "An error occurred"},
			expected: true,
		},
		{
			name: "string contains no match",
			filter: &WebhookFilter{
				FieldPath: "$.message",
				Operator:  OpContains,
				Value:     "error",
				Enabled:   true,
			},
			payload:  map[string]interface{}{"message": "Success"},
			expected: false,
		},
		{
			name: "case sensitive contains",
			filter: &WebhookFilter{
				FieldPath: "$.message",
				Operator:  OpContains,
				Value:     "Error",
				Enabled:   true,
			},
			payload:  map[string]interface{}{"message": "An error occurred"},
			expected: false,
		},
	}

	evaluator := NewFilterEvaluator(nil)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := evaluator.EvaluateSingle(tt.filter, tt.payload)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestFilterEvaluator_NotContains tests the not contains operator
func TestFilterEvaluator_NotContains(t *testing.T) {
	tests := []struct {
		name     string
		filter   *WebhookFilter
		payload  map[string]interface{}
		expected bool
	}{
		{
			name: "string not contains match",
			filter: &WebhookFilter{
				FieldPath: "$.message",
				Operator:  OpNotContains,
				Value:     "error",
				Enabled:   true,
			},
			payload:  map[string]interface{}{"message": "Success"},
			expected: true,
		},
		{
			name: "string not contains no match",
			filter: &WebhookFilter{
				FieldPath: "$.message",
				Operator:  OpNotContains,
				Value:     "error",
				Enabled:   true,
			},
			payload:  map[string]interface{}{"message": "An error occurred"},
			expected: false,
		},
	}

	evaluator := NewFilterEvaluator(nil)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := evaluator.EvaluateSingle(tt.filter, tt.payload)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestFilterEvaluator_StartsWith tests the starts with operator
func TestFilterEvaluator_StartsWith(t *testing.T) {
	tests := []struct {
		name     string
		filter   *WebhookFilter
		payload  map[string]interface{}
		expected bool
	}{
		{
			name: "string starts with match",
			filter: &WebhookFilter{
				FieldPath: "$.url",
				Operator:  OpStartsWith,
				Value:     "https://",
				Enabled:   true,
			},
			payload:  map[string]interface{}{"url": "https://example.com"},
			expected: true,
		},
		{
			name: "string starts with no match",
			filter: &WebhookFilter{
				FieldPath: "$.url",
				Operator:  OpStartsWith,
				Value:     "https://",
				Enabled:   true,
			},
			payload:  map[string]interface{}{"url": "http://example.com"},
			expected: false,
		},
	}

	evaluator := NewFilterEvaluator(nil)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := evaluator.EvaluateSingle(tt.filter, tt.payload)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestFilterEvaluator_EndsWith tests the ends with operator
func TestFilterEvaluator_EndsWith(t *testing.T) {
	tests := []struct {
		name     string
		filter   *WebhookFilter
		payload  map[string]interface{}
		expected bool
	}{
		{
			name: "string ends with match",
			filter: &WebhookFilter{
				FieldPath: "$.filename",
				Operator:  OpEndsWith,
				Value:     ".pdf",
				Enabled:   true,
			},
			payload:  map[string]interface{}{"filename": "document.pdf"},
			expected: true,
		},
		{
			name: "string ends with no match",
			filter: &WebhookFilter{
				FieldPath: "$.filename",
				Operator:  OpEndsWith,
				Value:     ".pdf",
				Enabled:   true,
			},
			payload:  map[string]interface{}{"filename": "document.txt"},
			expected: false,
		},
	}

	evaluator := NewFilterEvaluator(nil)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := evaluator.EvaluateSingle(tt.filter, tt.payload)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestFilterEvaluator_Regex tests the regex operator
func TestFilterEvaluator_Regex(t *testing.T) {
	tests := []struct {
		name     string
		filter   *WebhookFilter
		payload  map[string]interface{}
		expected bool
		wantErr  bool
	}{
		{
			name: "regex match",
			filter: &WebhookFilter{
				FieldPath: "$.email",
				Operator:  OpRegex,
				Value:     `^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`,
				Enabled:   true,
			},
			payload:  map[string]interface{}{"email": "user@example.com"},
			expected: true,
		},
		{
			name: "regex no match",
			filter: &WebhookFilter{
				FieldPath: "$.email",
				Operator:  OpRegex,
				Value:     `^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`,
				Enabled:   true,
			},
			payload:  map[string]interface{}{"email": "invalid-email"},
			expected: false,
		},
		{
			name: "invalid regex pattern",
			filter: &WebhookFilter{
				FieldPath: "$.email",
				Operator:  OpRegex,
				Value:     `[invalid(`,
				Enabled:   true,
			},
			payload: map[string]interface{}{"email": "user@example.com"},
			wantErr: true,
		},
	}

	evaluator := NewFilterEvaluator(nil)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := evaluator.EvaluateSingle(tt.filter, tt.payload)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestFilterEvaluator_GreaterThan tests the greater than operator
func TestFilterEvaluator_GreaterThan(t *testing.T) {
	tests := []struct {
		name     string
		filter   *WebhookFilter
		payload  map[string]interface{}
		expected bool
		wantErr  bool
	}{
		{
			name: "number greater than match",
			filter: &WebhookFilter{
				FieldPath: "$.count",
				Operator:  OpGreaterThan,
				Value:     10.0,
				Enabled:   true,
			},
			payload:  map[string]interface{}{"count": 15},
			expected: true,
		},
		{
			name: "number greater than no match",
			filter: &WebhookFilter{
				FieldPath: "$.count",
				Operator:  OpGreaterThan,
				Value:     10.0,
				Enabled:   true,
			},
			payload:  map[string]interface{}{"count": 5},
			expected: false,
		},
		{
			name: "number equal not greater",
			filter: &WebhookFilter{
				FieldPath: "$.count",
				Operator:  OpGreaterThan,
				Value:     10.0,
				Enabled:   true,
			},
			payload:  map[string]interface{}{"count": 10},
			expected: false,
		},
	}

	evaluator := NewFilterEvaluator(nil)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := evaluator.EvaluateSingle(tt.filter, tt.payload)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestFilterEvaluator_LessThan tests the less than operator
func TestFilterEvaluator_LessThan(t *testing.T) {
	tests := []struct {
		name     string
		filter   *WebhookFilter
		payload  map[string]interface{}
		expected bool
	}{
		{
			name: "number less than match",
			filter: &WebhookFilter{
				FieldPath: "$.count",
				Operator:  OpLessThan,
				Value:     10.0,
				Enabled:   true,
			},
			payload:  map[string]interface{}{"count": 5},
			expected: true,
		},
		{
			name: "number less than no match",
			filter: &WebhookFilter{
				FieldPath: "$.count",
				Operator:  OpLessThan,
				Value:     10.0,
				Enabled:   true,
			},
			payload:  map[string]interface{}{"count": 15},
			expected: false,
		},
		{
			name: "number equal not less",
			filter: &WebhookFilter{
				FieldPath: "$.count",
				Operator:  OpLessThan,
				Value:     10.0,
				Enabled:   true,
			},
			payload:  map[string]interface{}{"count": 10},
			expected: false,
		},
	}

	evaluator := NewFilterEvaluator(nil)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := evaluator.EvaluateSingle(tt.filter, tt.payload)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestFilterEvaluator_In tests the in operator
func TestFilterEvaluator_In(t *testing.T) {
	tests := []struct {
		name     string
		filter   *WebhookFilter
		payload  map[string]interface{}
		expected bool
		wantErr  bool
	}{
		{
			name: "value in array match",
			filter: &WebhookFilter{
				FieldPath: "$.status",
				Operator:  OpIn,
				Value:     []interface{}{"active", "pending", "completed"},
				Enabled:   true,
			},
			payload:  map[string]interface{}{"status": "active"},
			expected: true,
		},
		{
			name: "value in array no match",
			filter: &WebhookFilter{
				FieldPath: "$.status",
				Operator:  OpIn,
				Value:     []interface{}{"active", "pending", "completed"},
				Enabled:   true,
			},
			payload:  map[string]interface{}{"status": "failed"},
			expected: false,
		},
		{
			name: "number in array match",
			filter: &WebhookFilter{
				FieldPath: "$.code",
				Operator:  OpIn,
				Value:     []interface{}{200.0, 201.0, 204.0},
				Enabled:   true,
			},
			payload:  map[string]interface{}{"code": 200},
			expected: true,
		},
	}

	evaluator := NewFilterEvaluator(nil)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := evaluator.EvaluateSingle(tt.filter, tt.payload)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestFilterEvaluator_NotIn tests the not in operator
func TestFilterEvaluator_NotIn(t *testing.T) {
	tests := []struct {
		name     string
		filter   *WebhookFilter
		payload  map[string]interface{}
		expected bool
	}{
		{
			name: "value not in array match",
			filter: &WebhookFilter{
				FieldPath: "$.status",
				Operator:  OpNotIn,
				Value:     []interface{}{"failed", "error"},
				Enabled:   true,
			},
			payload:  map[string]interface{}{"status": "active"},
			expected: true,
		},
		{
			name: "value not in array no match",
			filter: &WebhookFilter{
				FieldPath: "$.status",
				Operator:  OpNotIn,
				Value:     []interface{}{"active", "pending", "completed"},
				Enabled:   true,
			},
			payload:  map[string]interface{}{"status": "active"},
			expected: false,
		},
	}

	evaluator := NewFilterEvaluator(nil)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := evaluator.EvaluateSingle(tt.filter, tt.payload)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestFilterEvaluator_Exists tests the exists operator
func TestFilterEvaluator_Exists(t *testing.T) {
	tests := []struct {
		name     string
		filter   *WebhookFilter
		payload  map[string]interface{}
		expected bool
	}{
		{
			name: "field exists",
			filter: &WebhookFilter{
				FieldPath: "$.data",
				Operator:  OpExists,
				Enabled:   true,
			},
			payload:  map[string]interface{}{"data": "value"},
			expected: true,
		},
		{
			name: "field does not exist",
			filter: &WebhookFilter{
				FieldPath: "$.missing",
				Operator:  OpExists,
				Enabled:   true,
			},
			payload:  map[string]interface{}{"data": "value"},
			expected: false,
		},
		{
			name: "nested field exists",
			filter: &WebhookFilter{
				FieldPath: "$.data.nested",
				Operator:  OpExists,
				Enabled:   true,
			},
			payload: map[string]interface{}{
				"data": map[string]interface{}{
					"nested": "value",
				},
			},
			expected: true,
		},
		{
			name: "field exists with null value",
			filter: &WebhookFilter{
				FieldPath: "$.data",
				Operator:  OpExists,
				Enabled:   true,
			},
			payload:  map[string]interface{}{"data": nil},
			expected: true,
		},
	}

	evaluator := NewFilterEvaluator(nil)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := evaluator.EvaluateSingle(tt.filter, tt.payload)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestFilterEvaluator_NotExists tests the not exists operator
func TestFilterEvaluator_NotExists(t *testing.T) {
	tests := []struct {
		name     string
		filter   *WebhookFilter
		payload  map[string]interface{}
		expected bool
	}{
		{
			name: "field not exists match",
			filter: &WebhookFilter{
				FieldPath: "$.missing",
				Operator:  OpNotExists,
				Enabled:   true,
			},
			payload:  map[string]interface{}{"data": "value"},
			expected: true,
		},
		{
			name: "field not exists no match",
			filter: &WebhookFilter{
				FieldPath: "$.data",
				Operator:  OpNotExists,
				Enabled:   true,
			},
			payload:  map[string]interface{}{"data": "value"},
			expected: false,
		},
	}

	evaluator := NewFilterEvaluator(nil)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := evaluator.EvaluateSingle(tt.filter, tt.payload)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestFilterEvaluator_JSONPath tests JSON path extraction
func TestFilterEvaluator_JSONPath(t *testing.T) {
	tests := []struct {
		name     string
		filter   *WebhookFilter
		payload  map[string]interface{}
		expected bool
	}{
		{
			name: "simple path",
			filter: &WebhookFilter{
				FieldPath: "$.status",
				Operator:  OpEquals,
				Value:     "active",
				Enabled:   true,
			},
			payload:  map[string]interface{}{"status": "active"},
			expected: true,
		},
		{
			name: "nested path",
			filter: &WebhookFilter{
				FieldPath: "$.data.user.name",
				Operator:  OpEquals,
				Value:     "John",
				Enabled:   true,
			},
			payload: map[string]interface{}{
				"data": map[string]interface{}{
					"user": map[string]interface{}{
						"name": "John",
					},
				},
			},
			expected: true,
		},
		{
			name: "array index path",
			filter: &WebhookFilter{
				FieldPath: "$.items.0.id",
				Operator:  OpEquals,
				Value:     "123",
				Enabled:   true,
			},
			payload: map[string]interface{}{
				"items": []interface{}{
					map[string]interface{}{
						"id": "123",
					},
				},
			},
			expected: true,
		},
		{
			name: "path without $ prefix",
			filter: &WebhookFilter{
				FieldPath: "status",
				Operator:  OpEquals,
				Value:     "active",
				Enabled:   true,
			},
			payload:  map[string]interface{}{"status": "active"},
			expected: true,
		},
	}

	evaluator := NewFilterEvaluator(nil)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := evaluator.EvaluateSingle(tt.filter, tt.payload)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestFilterEvaluator_InvalidPath tests error handling for invalid paths
func TestFilterEvaluator_InvalidPath(t *testing.T) {
	tests := []struct {
		name    string
		filter  *WebhookFilter
		payload map[string]interface{}
		wantErr bool
	}{
		{
			name: "deeply nested missing path",
			filter: &WebhookFilter{
				FieldPath: "$.data.missing.nested",
				Operator:  OpEquals,
				Value:     "value",
				Enabled:   true,
			},
			payload: map[string]interface{}{
				"data": map[string]interface{}{},
			},
			wantErr: false, // Should return false, not error
		},
	}

	evaluator := NewFilterEvaluator(nil)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := evaluator.EvaluateSingle(tt.filter, tt.payload)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.False(t, result)
			}
		})
	}
}

// TestFilterEvaluator_TypeMismatch tests type mismatch handling
func TestFilterEvaluator_TypeMismatch(t *testing.T) {
	tests := []struct {
		name     string
		filter   *WebhookFilter
		payload  map[string]interface{}
		expected bool
		wantErr  bool
	}{
		{
			name: "string operator on number field",
			filter: &WebhookFilter{
				FieldPath: "$.count",
				Operator:  OpContains,
				Value:     "5",
				Enabled:   true,
			},
			payload: map[string]interface{}{"count": 5},
			wantErr: true,
		},
		{
			name: "number operator on string field",
			filter: &WebhookFilter{
				FieldPath: "$.status",
				Operator:  OpGreaterThan,
				Value:     10.0,
				Enabled:   true,
			},
			payload: map[string]interface{}{"status": "active"},
			wantErr: true,
		},
	}

	evaluator := NewFilterEvaluator(nil)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := evaluator.EvaluateSingle(tt.filter, tt.payload)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

// TestFilterEvaluator_DisabledFilter tests that disabled filters are skipped
func TestFilterEvaluator_DisabledFilter(t *testing.T) {
	filter := &WebhookFilter{
		FieldPath: "$.status",
		Operator:  OpEquals,
		Value:     "active",
		Enabled:   false, // Disabled
	}

	payload := map[string]interface{}{"status": "inactive"}

	evaluator := NewFilterEvaluator(nil)
	result, err := evaluator.EvaluateSingle(filter, payload)
	require.NoError(t, err)
	assert.True(t, result, "Disabled filters should always pass")
}

// MockRepository implements Repository interface for testing
type MockRepository struct {
	filters []*WebhookFilter
	getErr  error
}

func (m *MockRepository) GetFiltersByWebhookID(ctx context.Context, webhookID string) ([]*WebhookFilter, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	return m.filters, nil
}

// TestFilterEvaluator_MultipleFilters_AND tests multiple filters with AND logic
func TestFilterEvaluator_MultipleFilters_AND(t *testing.T) {
	tests := []struct {
		name     string
		filters  []*WebhookFilter
		payload  map[string]interface{}
		expected bool
	}{
		{
			name: "all filters pass",
			filters: []*WebhookFilter{
				{
					ID:         "1",
					WebhookID:  "webhook-1",
					FieldPath:  "$.status",
					Operator:   OpEquals,
					Value:      "active",
					LogicGroup: 0,
					Enabled:    true,
				},
				{
					ID:         "2",
					WebhookID:  "webhook-1",
					FieldPath:  "$.priority",
					Operator:   OpGreaterThan,
					Value:      5.0,
					LogicGroup: 0,
					Enabled:    true,
				},
			},
			payload: map[string]interface{}{
				"status":   "active",
				"priority": 10,
			},
			expected: true,
		},
		{
			name: "one filter fails",
			filters: []*WebhookFilter{
				{
					ID:         "1",
					WebhookID:  "webhook-1",
					FieldPath:  "$.status",
					Operator:   OpEquals,
					Value:      "active",
					LogicGroup: 0,
					Enabled:    true,
				},
				{
					ID:         "2",
					WebhookID:  "webhook-1",
					FieldPath:  "$.priority",
					Operator:   OpGreaterThan,
					Value:      5.0,
					LogicGroup: 0,
					Enabled:    true,
				},
			},
			payload: map[string]interface{}{
				"status":   "inactive",
				"priority": 10,
			},
			expected: false,
		},
		{
			name: "all filters fail",
			filters: []*WebhookFilter{
				{
					ID:         "1",
					WebhookID:  "webhook-1",
					FieldPath:  "$.status",
					Operator:   OpEquals,
					Value:      "active",
					LogicGroup: 0,
					Enabled:    true,
				},
				{
					ID:         "2",
					WebhookID:  "webhook-1",
					FieldPath:  "$.priority",
					Operator:   OpGreaterThan,
					Value:      5.0,
					LogicGroup: 0,
					Enabled:    true,
				},
			},
			payload: map[string]interface{}{
				"status":   "inactive",
				"priority": 3,
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &MockRepository{filters: tt.filters}
			evaluator := NewFilterEvaluator(mockRepo)

			result, err := evaluator.Evaluate(context.Background(), "webhook-1", tt.payload)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, result.Passed)
		})
	}
}

// TestFilterEvaluator_MultipleFilters_OR tests multiple filters with OR logic (different logic groups)
func TestFilterEvaluator_MultipleFilters_OR(t *testing.T) {
	tests := []struct {
		name     string
		filters  []*WebhookFilter
		payload  map[string]interface{}
		expected bool
	}{
		{
			name: "one group passes",
			filters: []*WebhookFilter{
				{
					ID:         "1",
					WebhookID:  "webhook-1",
					FieldPath:  "$.status",
					Operator:   OpEquals,
					Value:      "active",
					LogicGroup: 0,
					Enabled:    true,
				},
				{
					ID:         "2",
					WebhookID:  "webhook-1",
					FieldPath:  "$.status",
					Operator:   OpEquals,
					Value:      "pending",
					LogicGroup: 1,
					Enabled:    true,
				},
			},
			payload: map[string]interface{}{
				"status": "active",
			},
			expected: true,
		},
		{
			name: "second group passes",
			filters: []*WebhookFilter{
				{
					ID:         "1",
					WebhookID:  "webhook-1",
					FieldPath:  "$.status",
					Operator:   OpEquals,
					Value:      "active",
					LogicGroup: 0,
					Enabled:    true,
				},
				{
					ID:         "2",
					WebhookID:  "webhook-1",
					FieldPath:  "$.status",
					Operator:   OpEquals,
					Value:      "pending",
					LogicGroup: 1,
					Enabled:    true,
				},
			},
			payload: map[string]interface{}{
				"status": "pending",
			},
			expected: true,
		},
		{
			name: "no groups pass",
			filters: []*WebhookFilter{
				{
					ID:         "1",
					WebhookID:  "webhook-1",
					FieldPath:  "$.status",
					Operator:   OpEquals,
					Value:      "active",
					LogicGroup: 0,
					Enabled:    true,
				},
				{
					ID:         "2",
					WebhookID:  "webhook-1",
					FieldPath:  "$.status",
					Operator:   OpEquals,
					Value:      "pending",
					LogicGroup: 1,
					Enabled:    true,
				},
			},
			payload: map[string]interface{}{
				"status": "failed",
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &MockRepository{filters: tt.filters}
			evaluator := NewFilterEvaluator(mockRepo)

			result, err := evaluator.Evaluate(context.Background(), "webhook-1", tt.payload)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, result.Passed)
		})
	}
}

// TestFilterEvaluator_NoFilters tests evaluation with no filters
func TestFilterEvaluator_NoFilters(t *testing.T) {
	mockRepo := &MockRepository{filters: []*WebhookFilter{}}
	evaluator := NewFilterEvaluator(mockRepo)

	payload := map[string]interface{}{"status": "active"}

	result, err := evaluator.Evaluate(context.Background(), "webhook-1", payload)
	require.NoError(t, err)
	assert.True(t, result.Passed, "No filters should pass by default")
}

// TestFilterEvaluator_ComplexNestedPayload tests complex nested payload evaluation
func TestFilterEvaluator_ComplexNestedPayload(t *testing.T) {
	filter := &WebhookFilter{
		FieldPath: "$.event.data.issue.labels.0.name",
		Operator:  OpEquals,
		Value:     "bug",
		Enabled:   true,
	}

	payload := map[string]interface{}{
		"event": map[string]interface{}{
			"data": map[string]interface{}{
				"issue": map[string]interface{}{
					"labels": []interface{}{
						map[string]interface{}{
							"name": "bug",
						},
					},
				},
			},
		},
	}

	evaluator := NewFilterEvaluator(nil)
	result, err := evaluator.EvaluateSingle(filter, payload)
	require.NoError(t, err)
	assert.True(t, result)
}

// TestFilterEvaluator_AllDisabledFilters tests that all disabled filters pass by default
func TestFilterEvaluator_AllDisabledFilters(t *testing.T) {
	mockRepo := &MockRepository{
		filters: []*WebhookFilter{
			{
				ID:         "1",
				WebhookID:  "webhook-1",
				FieldPath:  "$.status",
				Operator:   OpEquals,
				Value:      "active",
				LogicGroup: 0,
				Enabled:    false,
			},
			{
				ID:         "2",
				WebhookID:  "webhook-1",
				FieldPath:  "$.priority",
				Operator:   OpGreaterThan,
				Value:      5.0,
				LogicGroup: 0,
				Enabled:    false,
			},
		},
	}

	evaluator := NewFilterEvaluator(mockRepo)
	payload := map[string]interface{}{
		"status":   "inactive",
		"priority": 3,
	}

	result, err := evaluator.Evaluate(context.Background(), "webhook-1", payload)
	require.NoError(t, err)
	assert.True(t, result.Passed, "All disabled filters should pass")
	assert.Equal(t, "all filters disabled", result.Reason)
}

// TestFilterEvaluator_RepositoryError tests error handling when repository fails
func TestFilterEvaluator_RepositoryError(t *testing.T) {
	mockRepo := &MockRepository{
		getErr: assert.AnError,
	}

	evaluator := NewFilterEvaluator(mockRepo)
	payload := map[string]interface{}{"status": "active"}

	result, err := evaluator.Evaluate(context.Background(), "webhook-1", payload)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to get filters")
}

// TestFilterEvaluator_MixedLogicGroups tests complex OR logic with multiple groups
func TestFilterEvaluator_MixedLogicGroups(t *testing.T) {
	tests := []struct {
		name     string
		filters  []*WebhookFilter
		payload  map[string]interface{}
		expected bool
	}{
		{
			name: "group 0 with AND passes",
			filters: []*WebhookFilter{
				{
					ID:         "1",
					WebhookID:  "webhook-1",
					FieldPath:  "$.status",
					Operator:   OpEquals,
					Value:      "active",
					LogicGroup: 0,
					Enabled:    true,
				},
				{
					ID:         "2",
					WebhookID:  "webhook-1",
					FieldPath:  "$.priority",
					Operator:   OpGreaterThan,
					Value:      5.0,
					LogicGroup: 0,
					Enabled:    true,
				},
				{
					ID:         "3",
					WebhookID:  "webhook-1",
					FieldPath:  "$.status",
					Operator:   OpEquals,
					Value:      "pending",
					LogicGroup: 1,
					Enabled:    true,
				},
			},
			payload: map[string]interface{}{
				"status":   "active",
				"priority": 10,
			},
			expected: true,
		},
		{
			name: "group 1 passes when group 0 fails",
			filters: []*WebhookFilter{
				{
					ID:         "1",
					WebhookID:  "webhook-1",
					FieldPath:  "$.status",
					Operator:   OpEquals,
					Value:      "active",
					LogicGroup: 0,
					Enabled:    true,
				},
				{
					ID:         "2",
					WebhookID:  "webhook-1",
					FieldPath:  "$.priority",
					Operator:   OpGreaterThan,
					Value:      5.0,
					LogicGroup: 0,
					Enabled:    true,
				},
				{
					ID:         "3",
					WebhookID:  "webhook-1",
					FieldPath:  "$.type",
					Operator:   OpEquals,
					Value:      "urgent",
					LogicGroup: 1,
					Enabled:    true,
				},
			},
			payload: map[string]interface{}{
				"status":   "inactive",
				"priority": 3,
				"type":     "urgent",
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &MockRepository{filters: tt.filters}
			evaluator := NewFilterEvaluator(mockRepo)

			result, err := evaluator.Evaluate(context.Background(), "webhook-1", tt.payload)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, result.Passed)
		})
	}
}

// TestExtractValue tests JSON path extraction edge cases
func TestExtractValue(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		payload  map[string]interface{}
		expected interface{}
		exists   bool
	}{
		{
			name:     "empty path returns whole payload",
			path:     "$",
			payload:  map[string]interface{}{"key": "value"},
			expected: map[string]interface{}{"key": "value"},
			exists:   true,
		},
		{
			name:     "path without dollar sign",
			path:     "status",
			payload:  map[string]interface{}{"status": "active"},
			expected: "active",
			exists:   true,
		},
		{
			name:    "array index out of bounds",
			path:    "$.items.5",
			payload: map[string]interface{}{"items": []interface{}{1, 2, 3}},
			exists:  false,
		},
		{
			name:    "negative array index",
			path:    "$.items.-1",
			payload: map[string]interface{}{"items": []interface{}{1, 2, 3}},
			exists:  false,
		},
		{
			name: "array notation with brackets",
			path: "$.items[1].name",
			payload: map[string]interface{}{
				"items": []interface{}{
					map[string]interface{}{"name": "first"},
					map[string]interface{}{"name": "second"},
				},
			},
			expected: "second",
			exists:   true,
		},
		{
			name:    "array notation without field name",
			path:    "$[0]",
			payload: map[string]interface{}{"0": "value"},
			exists:  false, // JSON path doesn't support root-level array access like this
		},
		{
			name:    "invalid array index",
			path:    "$.items[abc]",
			payload: map[string]interface{}{"items": []interface{}{1, 2, 3}},
			exists:  false,
		},
		{
			name: "deeply nested with array",
			path: "$.data.users.0.addresses.1.city",
			payload: map[string]interface{}{
				"data": map[string]interface{}{
					"users": []interface{}{
						map[string]interface{}{
							"addresses": []interface{}{
								map[string]interface{}{"city": "New York"},
								map[string]interface{}{"city": "Boston"},
							},
						},
					},
				},
			},
			expected: "Boston",
			exists:   true,
		},
		{
			name:    "path to non-map field",
			path:    "$.status.nested",
			payload: map[string]interface{}{"status": "active"},
			exists:  false,
		},
		{
			name:    "path to non-array with index",
			path:    "$.status.0",
			payload: map[string]interface{}{"status": "active"},
			exists:  false,
		},
		{
			name:     "nil value exists",
			path:     "$.value",
			payload:  map[string]interface{}{"value": nil},
			expected: nil,
			exists:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			value, exists := extractValue(tt.path, tt.payload)
			assert.Equal(t, tt.exists, exists, "exists mismatch")
			if tt.exists {
				assert.Equal(t, tt.expected, value, "value mismatch")
			}
		})
	}
}

// TestCompareValues tests type coercion in value comparison
func TestCompareValues(t *testing.T) {
	tests := []struct {
		name     string
		a        interface{}
		b        interface{}
		expected bool
	}{
		{
			name:     "both nil",
			a:        nil,
			b:        nil,
			expected: true,
		},
		{
			name:     "one nil",
			a:        nil,
			b:        "value",
			expected: false,
		},
		{
			name:     "identical strings",
			a:        "test",
			b:        "test",
			expected: true,
		},
		{
			name:     "different strings",
			a:        "test",
			b:        "other",
			expected: false,
		},
		{
			name:     "int and float64 equal",
			a:        42,
			b:        42.0,
			expected: true,
		},
		{
			name:     "int and float64 not equal",
			a:        42,
			b:        43.0,
			expected: false,
		},
		{
			name:     "string number and int",
			a:        "42",
			b:        42,
			expected: true,
		},
		{
			name:     "string number and float",
			a:        "42.5",
			b:        42.5,
			expected: true,
		},
		{
			name:     "boolean true",
			a:        true,
			b:        true,
			expected: true,
		},
		{
			name:     "boolean false",
			a:        false,
			b:        false,
			expected: true,
		},
		{
			name:     "different types to string",
			a:        "true",
			b:        true,
			expected: true, // compareValues falls back to string comparison
		},
		{
			name:     "all numeric types",
			a:        int32(100),
			b:        uint64(100),
			expected: true,
		},
		{
			name:     "float32 and float64",
			a:        float32(3.14),
			b:        float64(3.14),
			expected: false, // float32 to float64 conversion loses precision
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := compareValues(tt.a, tt.b)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestToFloat64 tests numeric type conversions
func TestToFloat64(t *testing.T) {
	tests := []struct {
		name     string
		value    interface{}
		expected float64
		success  bool
	}{
		{
			name:     "float64",
			value:    42.5,
			expected: 42.5,
			success:  true,
		},
		{
			name:     "float32",
			value:    float32(42.5),
			expected: 42.5,
			success:  true,
		},
		{
			name:     "int",
			value:    42,
			expected: 42.0,
			success:  true,
		},
		{
			name:     "int8",
			value:    int8(42),
			expected: 42.0,
			success:  true,
		},
		{
			name:     "int16",
			value:    int16(42),
			expected: 42.0,
			success:  true,
		},
		{
			name:     "int32",
			value:    int32(42),
			expected: 42.0,
			success:  true,
		},
		{
			name:     "int64",
			value:    int64(42),
			expected: 42.0,
			success:  true,
		},
		{
			name:     "uint",
			value:    uint(42),
			expected: 42.0,
			success:  true,
		},
		{
			name:     "uint8",
			value:    uint8(42),
			expected: 42.0,
			success:  true,
		},
		{
			name:     "uint16",
			value:    uint16(42),
			expected: 42.0,
			success:  true,
		},
		{
			name:     "uint32",
			value:    uint32(42),
			expected: 42.0,
			success:  true,
		},
		{
			name:     "uint64",
			value:    uint64(42),
			expected: 42.0,
			success:  true,
		},
		{
			name:     "string number",
			value:    "42.5",
			expected: 42.5,
			success:  true,
		},
		{
			name:    "invalid string",
			value:   "not a number",
			success: false,
		},
		{
			name:    "boolean",
			value:   true,
			success: false,
		},
		{
			name:    "nil",
			value:   nil,
			success: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, success := toFloat64(tt.value)
			assert.Equal(t, tt.success, success)
			if tt.success {
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

// TestFilterEvaluator_TypeCoercion tests numeric type coercion in operators
func TestFilterEvaluator_TypeCoercion(t *testing.T) {
	tests := []struct {
		name     string
		filter   *WebhookFilter
		payload  map[string]interface{}
		expected bool
	}{
		{
			name: "int equals float",
			filter: &WebhookFilter{
				FieldPath: "$.count",
				Operator:  OpEquals,
				Value:     42.0,
				Enabled:   true,
			},
			payload:  map[string]interface{}{"count": 42},
			expected: true,
		},
		{
			name: "string number equals int",
			filter: &WebhookFilter{
				FieldPath: "$.count",
				Operator:  OpEquals,
				Value:     42,
				Enabled:   true,
			},
			payload:  map[string]interface{}{"count": "42"},
			expected: true,
		},
		{
			name: "greater than with mixed types",
			filter: &WebhookFilter{
				FieldPath: "$.count",
				Operator:  OpGreaterThan,
				Value:     10,
				Enabled:   true,
			},
			payload:  map[string]interface{}{"count": 15.5},
			expected: true,
		},
		{
			name: "less than with int32",
			filter: &WebhookFilter{
				FieldPath: "$.count",
				Operator:  OpLessThan,
				Value:     float64(100),
				Enabled:   true,
			},
			payload:  map[string]interface{}{"count": int32(50)},
			expected: true,
		},
	}

	evaluator := NewFilterEvaluator(nil)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := evaluator.EvaluateSingle(tt.filter, tt.payload)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestFilterEvaluator_StringOperatorErrors tests error handling for string operators
func TestFilterEvaluator_StringOperatorErrors(t *testing.T) {
	tests := []struct {
		name     string
		filter   *WebhookFilter
		payload  map[string]interface{}
		wantErr  bool
		errMatch string
	}{
		{
			name: "contains on non-string value",
			filter: &WebhookFilter{
				FieldPath: "$.count",
				Operator:  OpContains,
				Value:     "5",
				Enabled:   true,
			},
			payload:  map[string]interface{}{"count": 5},
			wantErr:  true,
			errMatch: "contains operator requires string value",
		},
		{
			name: "contains with non-string comparison",
			filter: &WebhookFilter{
				FieldPath: "$.message",
				Operator:  OpContains,
				Value:     123,
				Enabled:   true,
			},
			payload:  map[string]interface{}{"message": "test"},
			wantErr:  true,
			errMatch: "contains operator requires string comparison value",
		},
		{
			name: "starts_with on non-string value",
			filter: &WebhookFilter{
				FieldPath: "$.count",
				Operator:  OpStartsWith,
				Value:     "5",
				Enabled:   true,
			},
			payload:  map[string]interface{}{"count": 5},
			wantErr:  true,
			errMatch: "starts_with operator requires string value",
		},
		{
			name: "ends_with on non-string value",
			filter: &WebhookFilter{
				FieldPath: "$.count",
				Operator:  OpEndsWith,
				Value:     "5",
				Enabled:   true,
			},
			payload:  map[string]interface{}{"count": 5},
			wantErr:  true,
			errMatch: "ends_with operator requires string value",
		},
		{
			name: "regex on non-string value",
			filter: &WebhookFilter{
				FieldPath: "$.count",
				Operator:  OpRegex,
				Value:     "\\d+",
				Enabled:   true,
			},
			payload:  map[string]interface{}{"count": 5},
			wantErr:  true,
			errMatch: "regex operator requires string value",
		},
		{
			name: "regex with non-string pattern",
			filter: &WebhookFilter{
				FieldPath: "$.message",
				Operator:  OpRegex,
				Value:     123,
				Enabled:   true,
			},
			payload:  map[string]interface{}{"message": "test"},
			wantErr:  true,
			errMatch: "regex operator requires string pattern",
		},
	}

	evaluator := NewFilterEvaluator(nil)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := evaluator.EvaluateSingle(tt.filter, tt.payload)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMatch != "" {
					assert.Contains(t, err.Error(), tt.errMatch)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestFilterEvaluator_NumericOperatorErrors tests error handling for numeric operators
func TestFilterEvaluator_NumericOperatorErrors(t *testing.T) {
	tests := []struct {
		name     string
		filter   *WebhookFilter
		payload  map[string]interface{}
		wantErr  bool
		errMatch string
	}{
		{
			name: "greater than on non-numeric value",
			filter: &WebhookFilter{
				FieldPath: "$.status",
				Operator:  OpGreaterThan,
				Value:     10,
				Enabled:   true,
			},
			payload:  map[string]interface{}{"status": "active"},
			wantErr:  true,
			errMatch: "greater than operator requires numeric value",
		},
		{
			name: "greater than with non-numeric comparison",
			filter: &WebhookFilter{
				FieldPath: "$.count",
				Operator:  OpGreaterThan,
				Value:     "not a number",
				Enabled:   true,
			},
			payload:  map[string]interface{}{"count": 5},
			wantErr:  true,
			errMatch: "greater than operator requires numeric comparison value",
		},
		{
			name: "less than on non-numeric value",
			filter: &WebhookFilter{
				FieldPath: "$.status",
				Operator:  OpLessThan,
				Value:     10,
				Enabled:   true,
			},
			payload:  map[string]interface{}{"status": "active"},
			wantErr:  true,
			errMatch: "less than operator requires numeric value",
		},
		{
			name: "less than with non-numeric comparison",
			filter: &WebhookFilter{
				FieldPath: "$.count",
				Operator:  OpLessThan,
				Value:     "not a number",
				Enabled:   true,
			},
			payload:  map[string]interface{}{"count": 5},
			wantErr:  true,
			errMatch: "less than operator requires numeric comparison value",
		},
	}

	evaluator := NewFilterEvaluator(nil)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := evaluator.EvaluateSingle(tt.filter, tt.payload)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMatch != "" {
					assert.Contains(t, err.Error(), tt.errMatch)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestFilterEvaluator_InOperatorErrors tests error handling for in/not_in operators
func TestFilterEvaluator_InOperatorErrors(t *testing.T) {
	tests := []struct {
		name     string
		filter   *WebhookFilter
		payload  map[string]interface{}
		wantErr  bool
		errMatch string
	}{
		{
			name: "in operator with non-array value",
			filter: &WebhookFilter{
				FieldPath: "$.status",
				Operator:  OpIn,
				Value:     "active",
				Enabled:   true,
			},
			payload:  map[string]interface{}{"status": "active"},
			wantErr:  true,
			errMatch: "in operator requires array comparison value",
		},
		{
			name: "not_in operator with non-array value",
			filter: &WebhookFilter{
				FieldPath: "$.status",
				Operator:  OpNotIn,
				Value:     "active",
				Enabled:   true,
			},
			payload:  map[string]interface{}{"status": "inactive"},
			wantErr:  true,
			errMatch: "in operator requires array comparison value",
		},
	}

	evaluator := NewFilterEvaluator(nil)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := evaluator.EvaluateSingle(tt.filter, tt.payload)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMatch != "" {
					assert.Contains(t, err.Error(), tt.errMatch)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestFilterEvaluator_UnknownOperator tests handling of unknown operators
func TestFilterEvaluator_UnknownOperator(t *testing.T) {
	filter := &WebhookFilter{
		FieldPath: "$.status",
		Operator:  FilterOperator("unknown"),
		Value:     "active",
		Enabled:   true,
	}

	payload := map[string]interface{}{"status": "active"}

	evaluator := NewFilterEvaluator(nil)
	_, err := evaluator.EvaluateSingle(filter, payload)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown operator")
}

// TestFilterEvaluator_NilValues tests handling of nil values
func TestFilterEvaluator_NilValues(t *testing.T) {
	tests := []struct {
		name     string
		filter   *WebhookFilter
		payload  map[string]interface{}
		expected bool
		wantErr  bool
	}{
		{
			name: "equals with nil value",
			filter: &WebhookFilter{
				FieldPath: "$.data",
				Operator:  OpEquals,
				Value:     nil,
				Enabled:   true,
			},
			payload:  map[string]interface{}{"data": nil},
			expected: true,
		},
		{
			name: "not equals with nil value",
			filter: &WebhookFilter{
				FieldPath: "$.data",
				Operator:  OpNotEquals,
				Value:     nil,
				Enabled:   true,
			},
			payload:  map[string]interface{}{"data": "something"},
			expected: true,
		},
		{
			name: "exists with nil value",
			filter: &WebhookFilter{
				FieldPath: "$.data",
				Operator:  OpExists,
				Enabled:   true,
			},
			payload:  map[string]interface{}{"data": nil},
			expected: true,
		},
		{
			name: "in operator with nil in array",
			filter: &WebhookFilter{
				FieldPath: "$.value",
				Operator:  OpIn,
				Value:     []interface{}{nil, "test", 123},
				Enabled:   true,
			},
			payload:  map[string]interface{}{"value": nil},
			expected: true,
		},
	}

	evaluator := NewFilterEvaluator(nil)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := evaluator.EvaluateSingle(tt.filter, tt.payload)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

// TestFilterEvaluator_EmptyValues tests handling of empty strings and arrays
func TestFilterEvaluator_EmptyValues(t *testing.T) {
	tests := []struct {
		name     string
		filter   *WebhookFilter
		payload  map[string]interface{}
		expected bool
	}{
		{
			name: "empty string equals",
			filter: &WebhookFilter{
				FieldPath: "$.name",
				Operator:  OpEquals,
				Value:     "",
				Enabled:   true,
			},
			payload:  map[string]interface{}{"name": ""},
			expected: true,
		},
		{
			name: "empty string contains",
			filter: &WebhookFilter{
				FieldPath: "$.name",
				Operator:  OpContains,
				Value:     "",
				Enabled:   true,
			},
			payload:  map[string]interface{}{"name": "test"},
			expected: true,
		},
		{
			name: "starts with empty string",
			filter: &WebhookFilter{
				FieldPath: "$.name",
				Operator:  OpStartsWith,
				Value:     "",
				Enabled:   true,
			},
			payload:  map[string]interface{}{"name": "test"},
			expected: true,
		},
		{
			name: "ends with empty string",
			filter: &WebhookFilter{
				FieldPath: "$.name",
				Operator:  OpEndsWith,
				Value:     "",
				Enabled:   true,
			},
			payload:  map[string]interface{}{"name": "test"},
			expected: true,
		},
		{
			name: "in operator with empty array",
			filter: &WebhookFilter{
				FieldPath: "$.status",
				Operator:  OpIn,
				Value:     []interface{}{},
				Enabled:   true,
			},
			payload:  map[string]interface{}{"status": "active"},
			expected: false,
		},
	}

	evaluator := NewFilterEvaluator(nil)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := evaluator.EvaluateSingle(tt.filter, tt.payload)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestFilterEvaluator_CaseSensitivity tests case sensitivity of string operations
func TestFilterEvaluator_CaseSensitivity(t *testing.T) {
	tests := []struct {
		name     string
		filter   *WebhookFilter
		payload  map[string]interface{}
		expected bool
	}{
		{
			name: "equals is case sensitive",
			filter: &WebhookFilter{
				FieldPath: "$.status",
				Operator:  OpEquals,
				Value:     "Active",
				Enabled:   true,
			},
			payload:  map[string]interface{}{"status": "active"},
			expected: false,
		},
		{
			name: "contains is case sensitive",
			filter: &WebhookFilter{
				FieldPath: "$.message",
				Operator:  OpContains,
				Value:     "Error",
				Enabled:   true,
			},
			payload:  map[string]interface{}{"message": "an error occurred"},
			expected: false,
		},
		{
			name: "starts_with is case sensitive",
			filter: &WebhookFilter{
				FieldPath: "$.url",
				Operator:  OpStartsWith,
				Value:     "HTTPS://",
				Enabled:   true,
			},
			payload:  map[string]interface{}{"url": "https://example.com"},
			expected: false,
		},
		{
			name: "regex can be case insensitive",
			filter: &WebhookFilter{
				FieldPath: "$.status",
				Operator:  OpRegex,
				Value:     "(?i)active",
				Enabled:   true,
			},
			payload:  map[string]interface{}{"status": "ACTIVE"},
			expected: true,
		},
	}

	evaluator := NewFilterEvaluator(nil)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := evaluator.EvaluateSingle(tt.filter, tt.payload)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestFilterEvaluator_ComplexArrayPaths tests complex array path scenarios
func TestFilterEvaluator_ComplexArrayPaths(t *testing.T) {
	tests := []struct {
		name     string
		filter   *WebhookFilter
		payload  map[string]interface{}
		expected bool
		wantErr  bool
	}{
		{
			name: "bracket notation array access",
			filter: &WebhookFilter{
				FieldPath: "$.items[0].id",
				Operator:  OpEquals,
				Value:     "123",
				Enabled:   true,
			},
			payload: map[string]interface{}{
				"items": []interface{}{
					map[string]interface{}{"id": "123"},
				},
			},
			expected: true,
		},
		{
			name: "multiple array accesses",
			filter: &WebhookFilter{
				FieldPath: "$.data[0].items[1].name",
				Operator:  OpEquals,
				Value:     "test",
				Enabled:   true,
			},
			payload: map[string]interface{}{
				"data": []interface{}{
					map[string]interface{}{
						"items": []interface{}{
							map[string]interface{}{"name": "first"},
							map[string]interface{}{"name": "test"},
						},
					},
				},
			},
			expected: true,
		},
		{
			name: "array access on non-array field",
			filter: &WebhookFilter{
				FieldPath: "$.status[0]",
				Operator:  OpExists,
				Enabled:   true,
			},
			payload:  map[string]interface{}{"status": "active"},
			expected: false,
		},
	}

	evaluator := NewFilterEvaluator(nil)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := evaluator.EvaluateSingle(tt.filter, tt.payload)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

// TestFilterEvaluator_EdgeCases tests additional edge cases for completeness
func TestFilterEvaluator_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		filter   *WebhookFilter
		payload  map[string]interface{}
		expected bool
		wantErr  bool
	}{
		{
			name: "starts_with with non-string comparison value",
			filter: &WebhookFilter{
				FieldPath: "$.name",
				Operator:  OpStartsWith,
				Value:     123,
				Enabled:   true,
			},
			payload: map[string]interface{}{"name": "test123"},
			wantErr: true,
		},
		{
			name: "ends_with with non-string comparison value",
			filter: &WebhookFilter{
				FieldPath: "$.name",
				Operator:  OpEndsWith,
				Value:     123,
				Enabled:   true,
			},
			payload: map[string]interface{}{"name": "test123"},
			wantErr: true,
		},
		{
			name: "field does not exist for equals operator",
			filter: &WebhookFilter{
				FieldPath: "$.missing",
				Operator:  OpEquals,
				Value:     "value",
				Enabled:   true,
			},
			payload:  map[string]interface{}{"status": "active"},
			expected: false,
		},
		{
			name: "field does not exist for greater than operator",
			filter: &WebhookFilter{
				FieldPath: "$.missing",
				Operator:  OpGreaterThan,
				Value:     10,
				Enabled:   true,
			},
			payload:  map[string]interface{}{"count": 5},
			expected: false,
		},
	}

	evaluator := NewFilterEvaluator(nil)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := evaluator.EvaluateSingle(tt.filter, tt.payload)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

// TestFilterEvaluator_DetailedResults tests that filter results include proper details
func TestFilterEvaluator_DetailedResults(t *testing.T) {
	tests := []struct {
		name           string
		filters        []*WebhookFilter
		payload        map[string]interface{}
		expectedPassed bool
		expectedReason string
	}{
		{
			name:           "no filters configured",
			filters:        []*WebhookFilter{},
			payload:        map[string]interface{}{"status": "active"},
			expectedPassed: true,
			expectedReason: "no filters configured",
		},
		{
			name: "single group passes",
			filters: []*WebhookFilter{
				{
					ID:         "1",
					WebhookID:  "webhook-1",
					FieldPath:  "$.status",
					Operator:   OpEquals,
					Value:      "active",
					LogicGroup: 0,
					Enabled:    true,
				},
			},
			payload:        map[string]interface{}{"status": "active"},
			expectedPassed: true,
			expectedReason: "logic group 0 passed",
		},
		{
			name: "no groups pass with details",
			filters: []*WebhookFilter{
				{
					ID:         "1",
					WebhookID:  "webhook-1",
					FieldPath:  "$.status",
					Operator:   OpEquals,
					Value:      "active",
					LogicGroup: 0,
					Enabled:    true,
				},
				{
					ID:         "2",
					WebhookID:  "webhook-1",
					FieldPath:  "$.type",
					Operator:   OpEquals,
					Value:      "urgent",
					LogicGroup: 1,
					Enabled:    true,
				},
			},
			payload:        map[string]interface{}{"status": "inactive", "type": "normal"},
			expectedPassed: false,
			expectedReason: "no logic groups passed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &MockRepository{filters: tt.filters}
			evaluator := NewFilterEvaluator(mockRepo)

			result, err := evaluator.Evaluate(context.Background(), "webhook-1", tt.payload)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedPassed, result.Passed)
			assert.Contains(t, result.Reason, tt.expectedReason)
			assert.NotNil(t, result.Details)
		})
	}
}
