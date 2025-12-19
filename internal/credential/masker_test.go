package credential

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMaskString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		secrets  []string
		expected string
	}{
		{
			name:     "masks single secret",
			input:    "API Key: sk_test_abc123xyz",
			secrets:  []string{"sk_test_abc123xyz"},
			expected: "API Key: ***MASKED***",
		},
		{
			name:     "masks multiple secrets",
			input:    "username:admin password:secret123",
			secrets:  []string{"admin", "secret123"},
			expected: "username:***MASKED*** password:***MASKED***",
		},
		{
			name:     "masks secrets in JSON",
			input:    `{"api_key":"sk_live_123","token":"bearer_xyz"}`,
			secrets:  []string{"sk_live_123", "bearer_xyz"},
			expected: `{"api_key":"***MASKED***","token":"***MASKED***"}`,
		},
		{
			name:     "no secrets to mask",
			input:    "public information only",
			secrets:  []string{"secret"},
			expected: "public information only",
		},
		{
			name:     "empty secrets list",
			input:    "contains secret123",
			secrets:  []string{},
			expected: "contains secret123",
		},
		{
			name:     "empty input",
			input:    "",
			secrets:  []string{"secret"},
			expected: "",
		},
		{
			name:     "case sensitive masking",
			input:    "Secret: MyPassword",
			secrets:  []string{"MyPassword"},
			expected: "Secret: ***MASKED***",
		},
		{
			name:     "partial match should not mask",
			input:    "password123456",
			secrets:  []string{"123"},
			expected: "password***MASKED***456",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			masker := NewMasker()
			result := masker.MaskString(tt.input, tt.secrets)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMaskJSON(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]interface{}
		secrets  []string
		expected map[string]interface{}
	}{
		{
			name: "masks string values",
			input: map[string]interface{}{
				"api_key": "sk_test_123",
				"name":    "test",
			},
			secrets: []string{"sk_test_123"},
			expected: map[string]interface{}{
				"api_key": "***MASKED***",
				"name":    "test",
			},
		},
		{
			name: "masks nested objects",
			input: map[string]interface{}{
				"auth": map[string]interface{}{
					"token":    "secret_token",
					"username": "user",
				},
			},
			secrets: []string{"secret_token"},
			expected: map[string]interface{}{
				"auth": map[string]interface{}{
					"token":    "***MASKED***",
					"username": "user",
				},
			},
		},
		{
			name: "masks arrays",
			input: map[string]interface{}{
				"tokens": []interface{}{"token1", "token2", "safe"},
			},
			secrets: []string{"token1", "token2"},
			expected: map[string]interface{}{
				"tokens": []interface{}{"***MASKED***", "***MASKED***", "safe"},
			},
		},
		{
			name: "preserves non-string types",
			input: map[string]interface{}{
				"count":   123,
				"enabled": true,
				"rate":    3.14,
			},
			secrets: []string{"123"},
			expected: map[string]interface{}{
				"count":   123,
				"enabled": true,
				"rate":    3.14,
			},
		},
		{
			name: "deeply nested masking",
			input: map[string]interface{}{
				"level1": map[string]interface{}{
					"level2": map[string]interface{}{
						"level3": map[string]interface{}{
							"secret": "deep_secret",
						},
					},
				},
			},
			secrets: []string{"deep_secret"},
			expected: map[string]interface{}{
				"level1": map[string]interface{}{
					"level2": map[string]interface{}{
						"level3": map[string]interface{}{
							"secret": "***MASKED***",
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			masker := NewMasker()
			result := masker.MaskJSON(tt.input, tt.secrets)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMaskRawJSON(t *testing.T) {
	tests := []struct {
		name        string
		input       json.RawMessage
		secrets     []string
		expectError bool
	}{
		{
			name:        "masks valid JSON object",
			input:       json.RawMessage(`{"api_key":"secret123","name":"test"}`),
			secrets:     []string{"secret123"},
			expectError: false,
		},
		{
			name:        "masks JSON array",
			input:       json.RawMessage(`["secret1","public","secret2"]`),
			secrets:     []string{"secret1", "secret2"},
			expectError: false,
		},
		{
			name:        "handles invalid JSON",
			input:       json.RawMessage(`{invalid json}`),
			secrets:     []string{"secret"},
			expectError: true,
		},
		{
			name:        "handles empty JSON",
			input:       json.RawMessage(`{}`),
			secrets:     []string{"secret"},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			masker := NewMasker()
			result, err := masker.MaskRawJSON(tt.input, tt.secrets)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)

				// Verify secrets are masked in result
				resultStr := string(result)
				for _, secret := range tt.secrets {
					assert.NotContains(t, resultStr, secret)
				}
			}
		})
	}
}

func TestExtractSecrets(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected []string
	}{
		{
			name:     "extracts string",
			input:    "secret123",
			expected: []string{"secret123"},
		},
		{
			name: "extracts from map",
			input: map[string]interface{}{
				"key1": "value1",
				"key2": "value2",
			},
			expected: []string{"value1", "value2"},
		},
		{
			name:     "extracts from array",
			input:    []interface{}{"secret1", "secret2", "secret3"},
			expected: []string{"secret1", "secret2", "secret3"},
		},
		{
			name: "extracts from nested structure",
			input: map[string]interface{}{
				"auth": map[string]interface{}{
					"token": "tok_123",
					"keys":  []interface{}{"key1", "key2"},
				},
			},
			expected: []string{"tok_123", "key1", "key2"},
		},
		{
			name:     "ignores non-string values",
			input:    123,
			expected: []string{},
		},
		{
			name: "filters empty strings",
			input: map[string]interface{}{
				"key1": "",
				"key2": "value",
			},
			expected: []string{"value"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			masker := NewMasker()
			result := masker.ExtractSecrets(tt.input)
			assert.ElementsMatch(t, tt.expected, result)
		})
	}
}

func TestMaskWithCustomMask(t *testing.T) {
	masker := NewMaskerWithMask("[REDACTED]")

	input := "API Key: secret123"
	secrets := []string{"secret123"}
	result := masker.MaskString(input, secrets)

	assert.Equal(t, "API Key: [REDACTED]", result)
	assert.NotContains(t, result, "secret123")
}

func TestConcurrentMasking(t *testing.T) {
	masker := NewMasker()
	secrets := []string{"secret1", "secret2", "secret3"}

	// Test that masker is safe for concurrent use
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func(id int) {
			input := "This contains secret1 and secret2 and secret3"
			result := masker.MaskString(input, secrets)
			assert.NotContains(t, result, "secret1")
			assert.NotContains(t, result, "secret2")
			assert.NotContains(t, result, "secret3")
			done <- true
		}(i)
	}

	for i := 0; i < 10; i++ {
		<-done
	}
}
