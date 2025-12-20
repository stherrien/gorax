package aws

import (
	"testing"

	"github.com/gorax/gorax/internal/integrations"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRegisterAWSActions(t *testing.T) {
	registry := integrations.NewRegistry()

	err := RegisterAWSActions(registry, "test-key", "test-secret", "us-east-1")
	require.NoError(t, err)

	// Verify all actions are registered
	expectedActions := GetAWSActionsList()
	registeredActions := registry.List()

	assert.Len(t, registeredActions, len(expectedActions))

	for _, actionName := range expectedActions {
		t.Run("verify_"+actionName, func(t *testing.T) {
			action, err := registry.Get(actionName)
			assert.NoError(t, err)
			assert.NotNil(t, action)
			assert.Equal(t, actionName, action.Name())
		})
	}
}

func TestGetAWSActionsList(t *testing.T) {
	actions := GetAWSActionsList()

	assert.NotEmpty(t, actions)
	assert.Len(t, actions, 9) // 4 S3 + 1 SNS + 3 SQS + 1 Lambda

	// Verify expected action names are present
	expectedPrefixes := map[string]int{
		"aws:s3:":     4,
		"aws:sns:":    1,
		"aws:sqs:":    3,
		"aws:lambda:": 1,
	}

	for prefix, expectedCount := range expectedPrefixes {
		count := 0
		for _, action := range actions {
			if len(action) >= len(prefix) && action[:len(prefix)] == prefix {
				count++
			}
		}
		assert.Equal(t, expectedCount, count, "Expected %d actions with prefix %s", expectedCount, prefix)
	}
}

func TestRegisterAWSActions_DoubleRegistration(t *testing.T) {
	registry := integrations.NewRegistry()

	// First registration should succeed
	err := RegisterAWSActions(registry, "test-key", "test-secret", "us-east-1")
	require.NoError(t, err)

	// Second registration should fail (duplicate action names)
	err = RegisterAWSActions(registry, "test-key", "test-secret", "us-east-1")
	assert.Error(t, err)
}

func TestAWSActionsImplementInterface(t *testing.T) {
	registry := integrations.NewRegistry()
	err := RegisterAWSActions(registry, "test-key", "test-secret", "us-east-1")
	require.NoError(t, err)

	for _, actionName := range GetAWSActionsList() {
		t.Run(actionName, func(t *testing.T) {
			action, err := registry.Get(actionName)
			require.NoError(t, err)
			require.NotNil(t, action)

			// Verify interface methods are callable
			assert.Equal(t, actionName, action.Name())
			assert.NotEmpty(t, action.Description())

			// Validate should be callable
			err = action.Validate(map[string]interface{}{})
			// Don't assert error - some actions require config, some don't
			_ = err
		})
	}
}
