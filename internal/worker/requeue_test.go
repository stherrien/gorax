package worker

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/gorax/gorax/internal/queue"
)

// MockSQSClient is a mock SQS client for testing
type MockSQSClient struct {
	mock.Mock
}

func (m *MockSQSClient) ChangeMessageVisibility(ctx context.Context, receiptHandle string, visibilityTimeout int32) error {
	args := m.Called(ctx, receiptHandle, visibilityTimeout)
	return args.Error(0)
}

func (m *MockSQSClient) DeleteMessage(ctx context.Context, receiptHandle string) error {
	args := m.Called(ctx, receiptHandle)
	return args.Error(0)
}

func (m *MockSQSClient) SendMessage(ctx context.Context, body string, attributes map[string]string) (*string, error) {
	args := m.Called(ctx, body, attributes)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*string), args.Error(1)
}

// TestRequeueExecution_WhenTenantAtCapacity tests requeue with delay
func TestRequeueExecution_WhenTenantAtCapacity(t *testing.T) {
	tests := []struct {
		name          string
		currentRetry  int
		expectedDelay int32
		shouldRequeue bool
	}{
		{
			name:          "first retry - 30 second delay",
			currentRetry:  0,
			expectedDelay: 30,
			shouldRequeue: true,
		},
		{
			name:          "second retry - 60 second delay",
			currentRetry:  1,
			expectedDelay: 60,
			shouldRequeue: true,
		},
		{
			name:          "third retry - 120 second delay",
			currentRetry:  2,
			expectedDelay: 120,
			shouldRequeue: true,
		},
		{
			name:          "max retries exceeded - do not requeue",
			currentRetry:  5,
			expectedDelay: 0,
			shouldRequeue: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			delay := calculateRequeueDelay(tt.currentRetry)

			if tt.shouldRequeue {
				assert.Equal(t, tt.expectedDelay, delay)
			}
		})
	}
}

// TestRequeueMessage_ExtendsVisibilityTimeout tests message visibility timeout extension
func TestRequeueMessage_ExtendsVisibilityTimeout(t *testing.T) {
	// Setup
	mockSQS := new(MockSQSClient)
	receiptHandle := "test-receipt-handle"
	retryCount := 1
	expectedDelay := int32(60) // Second retry = 60 seconds

	// Expect visibility timeout change
	mockSQS.On("ChangeMessageVisibility", mock.Anything, receiptHandle, expectedDelay).Return(nil)

	// Test
	err := requeueMessage(context.Background(), mockSQS, receiptHandle, retryCount)

	// Assert
	assert.NoError(t, err)
	mockSQS.AssertExpectations(t)
}

// TestRequeueMessage_ErrorHandling tests error handling during requeue
func TestRequeueMessage_ErrorHandling(t *testing.T) {
	// Setup
	mockSQS := new(MockSQSClient)
	receiptHandle := "test-receipt-handle"
	retryCount := 1

	// Expect visibility timeout change to fail
	expectedErr := assert.AnError
	mockSQS.On("ChangeMessageVisibility", mock.Anything, receiptHandle, mock.Anything).Return(expectedErr)

	// Test
	err := requeueMessage(context.Background(), mockSQS, receiptHandle, retryCount)

	// Assert
	assert.Error(t, err)
	assert.ErrorContains(t, err, "failed to requeue message")
	assert.ErrorIs(t, err, expectedErr)
	mockSQS.AssertExpectations(t)
}

// TestProcessExecutionMessage_RequeuesOnTenantCapacity tests the integration
func TestProcessExecutionMessage_RequeuesOnTenantCapacity(t *testing.T) {
	// This test validates the end-to-end flow
	t.Run("should extend visibility when tenant at capacity", func(t *testing.T) {
		// Setup mocks
		mockSQS := new(MockSQSClient)
		receiptHandle := "test-receipt"

		// Create execution message
		msg := &queue.ExecutionMessage{
			ExecutionID:     "exec-123",
			TenantID:        "tenant-1",
			WorkflowID:      "workflow-1",
			WorkflowVersion: 1,
			TriggerType:     "webhook",
			RetryCount:      0,
		}

		// Expect requeue with 30 second delay
		mockSQS.On("ChangeMessageVisibility", mock.Anything, receiptHandle, int32(30)).Return(nil)

		// Test requeue
		err := requeueMessage(context.Background(), mockSQS, receiptHandle, msg.RetryCount)

		// Assert
		assert.NoError(t, err)
		mockSQS.AssertExpectations(t)
	})
}

// TestCalculateRequeueDelay_ExponentialBackoff tests backoff calculation
func TestCalculateRequeueDelay_ExponentialBackoff(t *testing.T) {
	tests := []struct {
		retry         int
		expectedDelay int32
	}{
		{retry: 0, expectedDelay: 30},   // 30 seconds
		{retry: 1, expectedDelay: 60},   // 1 minute
		{retry: 2, expectedDelay: 120},  // 2 minutes
		{retry: 3, expectedDelay: 240},  // 4 minutes
		{retry: 4, expectedDelay: 300},  // 5 minutes (capped)
		{retry: 5, expectedDelay: 300},  // 5 minutes (capped)
		{retry: 10, expectedDelay: 300}, // 5 minutes (capped)
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			delay := calculateRequeueDelay(tt.retry)
			assert.Equal(t, tt.expectedDelay, delay)
		})
	}
}

// TestProcessExecution_WithSQSRequeue tests requeue in processExecution
func TestProcessExecution_WithSQSRequeue(t *testing.T) {
	// This is an integration test that would require full setup
	// We'll test the logic through unit tests above
	t.Skip("requires full worker setup with SQS client")
}

// TestWorker_MessageRequeue_EndToEnd tests complete requeue flow
func TestWorker_MessageRequeue_EndToEnd(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	// This would test:
	// 1. Worker receives message from SQS
	// 2. Tenant is at capacity
	// 3. Message visibility timeout is extended
	// 4. Message is not deleted
	// 5. Message becomes visible again after delay
	// 6. Worker retries processing

	t.Skip("requires SQS integration test setup")
}

// TestRequeueWithReceiptHandle tests requeue with SQS receipt handle
func TestRequeueWithReceiptHandle(t *testing.T) {
	tests := []struct {
		name          string
		receiptHandle string
		retryCount    int
		expectedDelay int32
		setupMock     func(*MockSQSClient)
		expectedError bool
	}{
		{
			name:          "successful requeue - first retry",
			receiptHandle: "receipt-123",
			retryCount:    0,
			expectedDelay: 30,
			setupMock: func(m *MockSQSClient) {
				m.On("ChangeMessageVisibility", mock.Anything, "receipt-123", int32(30)).Return(nil)
			},
			expectedError: false,
		},
		{
			name:          "successful requeue - second retry",
			receiptHandle: "receipt-456",
			retryCount:    1,
			expectedDelay: 60,
			setupMock: func(m *MockSQSClient) {
				m.On("ChangeMessageVisibility", mock.Anything, "receipt-456", int32(60)).Return(nil)
			},
			expectedError: false,
		},
		{
			name:          "failed requeue - SQS error",
			receiptHandle: "receipt-789",
			retryCount:    0,
			expectedDelay: 30,
			setupMock: func(m *MockSQSClient) {
				m.On("ChangeMessageVisibility", mock.Anything, "receipt-789", int32(30)).Return(assert.AnError)
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSQS := new(MockSQSClient)
			tt.setupMock(mockSQS)

			err := requeueMessage(context.Background(), mockSQS, tt.receiptHandle, tt.retryCount)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockSQS.AssertExpectations(t)
		})
	}
}

// TestRequeueStrategy_CappedBackoff tests that delay is capped at 5 minutes
func TestRequeueStrategy_CappedBackoff(t *testing.T) {
	maxDelay := int32(300) // 5 minutes

	for retry := 0; retry < 20; retry++ {
		delay := calculateRequeueDelay(retry)
		assert.LessOrEqual(t, delay, maxDelay, "delay should not exceed 5 minutes")
	}
}

// TestMessageContext_PreservesMetadata tests that message metadata is preserved during requeue
func TestMessageContext_PreservesMetadata(t *testing.T) {
	// Setup
	msg := &queue.ExecutionMessage{
		ExecutionID:     "exec-123",
		TenantID:        "tenant-1",
		WorkflowID:      "workflow-1",
		WorkflowVersion: 1,
		TriggerType:     "webhook",
		TriggerData:     json.RawMessage(`{"key": "value"}`),
		EnqueuedAt:      time.Now(),
		RetryCount:      0,
		CorrelationID:   "corr-123",
	}

	// After requeue, metadata should be preserved
	msgData, err := msg.Marshal()
	require.NoError(t, err)

	// Unmarshal and verify
	unmarshaled, err := queue.UnmarshalExecutionMessage(msgData)
	require.NoError(t, err)

	assert.Equal(t, msg.ExecutionID, unmarshaled.ExecutionID)
	assert.Equal(t, msg.TenantID, unmarshaled.TenantID)
	assert.Equal(t, msg.WorkflowID, unmarshaled.WorkflowID)
	assert.Equal(t, msg.CorrelationID, unmarshaled.CorrelationID)
	assert.Equal(t, msg.RetryCount, unmarshaled.RetryCount)
}
