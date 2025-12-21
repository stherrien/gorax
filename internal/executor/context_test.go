package executor

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExecutionContext_UserID(t *testing.T) {
	tests := []struct {
		name           string
		triggerType    string
		triggerData    map[string]interface{}
		expectedUserID string
	}{
		{
			name:        "manual trigger with user ID in trigger data",
			triggerType: "manual",
			triggerData: map[string]interface{}{
				"user_id": "user-123",
			},
			expectedUserID: "user-123",
		},
		{
			name:        "webhook trigger with user ID in auth context",
			triggerType: "webhook",
			triggerData: map[string]interface{}{
				"_auth": map[string]interface{}{
					"user_id": "user-456",
				},
			},
			expectedUserID: "user-456",
		},
		{
			name:           "scheduled trigger defaults to system",
			triggerType:    "schedule",
			triggerData:    map[string]interface{}{},
			expectedUserID: "system",
		},
		{
			name:           "no user ID defaults to system",
			triggerType:    "manual",
			triggerData:    map[string]interface{}{},
			expectedUserID: "system",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := &ExecutionContext{
				TriggerType: tt.triggerType,
				TriggerData: tt.triggerData,
			}

			userID := ctx.GetUserID()
			assert.Equal(t, tt.expectedUserID, userID)
		})
	}
}

func TestExecutionContext_SetUserID(t *testing.T) {
	ctx := &ExecutionContext{
		TriggerData: make(map[string]interface{}),
	}

	// Initially should be system
	assert.Equal(t, "system", ctx.GetUserID())

	// Set user ID
	ctx.SetUserID("user-789")
	assert.Equal(t, "user-789", ctx.GetUserID())

	// Should be stored in context
	assert.Equal(t, "user-789", ctx.UserID)
}
