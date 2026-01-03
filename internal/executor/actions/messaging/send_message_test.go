package messaging

import (
	"context"
	"testing"
	"time"

	"github.com/gorax/gorax/internal/credential"
	"github.com/gorax/gorax/internal/messaging"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockCredentialService is a mock implementation of the credential service
type MockCredentialService struct {
	mock.Mock
}

func (m *MockCredentialService) GetCredentialValue(ctx context.Context, tenantID, credentialID string) (map[string]interface{}, error) {
	args := m.Called(ctx, tenantID, credentialID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string]interface{}), args.Error(1)
}

// MockMessageQueue is a mock implementation of the message queue interface
type MockMessageQueue struct {
	mock.Mock
}

func (m *MockMessageQueue) Send(ctx context.Context, destination string, message []byte, attributes map[string]string) error {
	args := m.Called(ctx, destination, message, attributes)
	return args.Error(0)
}

func (m *MockMessageQueue) Receive(ctx context.Context, source string, maxMessages int, waitTime time.Duration) ([]messaging.Message, error) {
	args := m.Called(ctx, source, maxMessages, waitTime)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]messaging.Message), args.Error(1)
}

func (m *MockMessageQueue) Ack(ctx context.Context, message messaging.Message) error {
	args := m.Called(ctx, message)
	return args.Error(0)
}

func (m *MockMessageQueue) Nack(ctx context.Context, message messaging.Message) error {
	args := m.Called(ctx, message)
	return args.Error(0)
}

func (m *MockMessageQueue) GetInfo(ctx context.Context, name string) (*messaging.QueueInfo, error) {
	args := m.Called(ctx, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*messaging.QueueInfo), args.Error(1)
}

func (m *MockMessageQueue) Close() error {
	args := m.Called()
	return args.Error(0)
}

func TestSendMessageConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  SendMessageConfig
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid config",
			config: SendMessageConfig{
				QueueType:    "sqs",
				Destination:  "test-queue",
				Message:      "test message",
				CredentialID: "cred-123",
			},
			wantErr: false,
		},
		{
			name: "missing queue type",
			config: SendMessageConfig{
				Destination:  "test-queue",
				Message:      "test message",
				CredentialID: "cred-123",
			},
			wantErr: true,
			errMsg:  "queue_type is required",
		},
		{
			name: "invalid queue type",
			config: SendMessageConfig{
				QueueType:    "invalid",
				Destination:  "test-queue",
				Message:      "test message",
				CredentialID: "cred-123",
			},
			wantErr: true,
			errMsg:  "unsupported queue_type",
		},
		{
			name: "missing destination",
			config: SendMessageConfig{
				QueueType:    "sqs",
				Message:      "test message",
				CredentialID: "cred-123",
			},
			wantErr: true,
			errMsg:  "destination is required",
		},
		{
			name: "missing message",
			config: SendMessageConfig{
				QueueType:    "sqs",
				Destination:  "test-queue",
				CredentialID: "cred-123",
			},
			wantErr: true,
			errMsg:  "message is required",
		},
		{
			name: "missing credential ID",
			config: SendMessageConfig{
				QueueType:   "sqs",
				Destination: "test-queue",
				Message:     "test message",
			},
			wantErr: true,
			errMsg:  "credential_id is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestSendMessageAction_Execute(t *testing.T) {
	tests := []struct {
		name           string
		config         SendMessageConfig
		input          map[string]interface{}
		setupMocks     func(*MockCredentialService, *MockMessageQueue)
		wantErr        bool
		wantOutput     map[string]interface{}
		wantOutputKeys []string
	}{
		{
			name: "successful SQS send",
			config: SendMessageConfig{
				QueueType:   "sqs",
				Destination: "https://sqs.us-east-1.amazonaws.com/123456789/test-queue",
				Message:     `{"workflow": "test"}`,
				Attributes: map[string]string{
					"priority": "high",
				},
				CredentialID: "cred-123",
			},
			input: map[string]interface{}{
				"tenant_id": "tenant-1",
			},
			setupMocks: func(credSvc *MockCredentialService, queue *MockMessageQueue) {
				credSvc.On("GetCredentialValue", mock.Anything, "tenant-1", "cred-123").
					Return(map[string]interface{}{
						"region":            "us-east-1",
						"access_key_id":     "AKIAIOSFODNN7EXAMPLE",
						"secret_access_key": "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
					}, nil)

				queue.On("Send",
					mock.Anything,
					"https://sqs.us-east-1.amazonaws.com/123456789/test-queue",
					[]byte(`{"workflow": "test"}`),
					map[string]string{"priority": "high"},
				).Return(nil)

				queue.On("Close").Return(nil)
			},
			wantErr:        false,
			wantOutputKeys: []string{"success", "message_id", "queue_type", "destination"},
		},
		{
			name: "successful Kafka send",
			config: SendMessageConfig{
				QueueType:    "kafka",
				Destination:  "test-topic",
				Message:      `{"event": "user_created"}`,
				CredentialID: "cred-456",
			},
			input: map[string]interface{}{
				"tenant_id": "tenant-1",
			},
			setupMocks: func(credSvc *MockCredentialService, queue *MockMessageQueue) {
				credSvc.On("GetCredentialValue", mock.Anything, "tenant-1", "cred-456").
					Return(map[string]interface{}{
						"brokers": []interface{}{"localhost:9092"},
					}, nil)

				queue.On("Send",
					mock.Anything,
					"test-topic",
					[]byte(`{"event": "user_created"}`),
					map[string]string{},
				).Return(nil)

				queue.On("Close").Return(nil)
			},
			wantErr:        false,
			wantOutputKeys: []string{"success", "message_id", "queue_type", "destination"},
		},
		{
			name: "successful RabbitMQ send",
			config: SendMessageConfig{
				QueueType:    "rabbitmq",
				Destination:  "workflow-queue",
				Message:      `{"task": "process"}`,
				CredentialID: "cred-789",
			},
			input: map[string]interface{}{
				"tenant_id": "tenant-1",
			},
			setupMocks: func(credSvc *MockCredentialService, queue *MockMessageQueue) {
				credSvc.On("GetCredentialValue", mock.Anything, "tenant-1", "cred-789").
					Return(map[string]interface{}{
						"url": "amqp://guest:guest@localhost:5672/",
					}, nil)

				queue.On("Send",
					mock.Anything,
					"workflow-queue",
					[]byte(`{"task": "process"}`),
					map[string]string{},
				).Return(nil)

				queue.On("Close").Return(nil)
			},
			wantErr:        false,
			wantOutputKeys: []string{"success", "message_id", "queue_type", "destination"},
		},
		{
			name: "credential not found",
			config: SendMessageConfig{
				QueueType:    "sqs",
				Destination:  "test-queue",
				Message:      "test",
				CredentialID: "cred-999",
			},
			input: map[string]interface{}{
				"tenant_id": "tenant-1",
			},
			setupMocks: func(credSvc *MockCredentialService, queue *MockMessageQueue) {
				credSvc.On("GetCredentialValue", mock.Anything, "tenant-1", "cred-999").
					Return(nil, credential.ErrUnauthorized)
			},
			wantErr: true,
		},
		{
			name: "missing tenant_id in input",
			config: SendMessageConfig{
				QueueType:    "sqs",
				Destination:  "test-queue",
				Message:      "test",
				CredentialID: "cred-123",
			},
			input:   map[string]interface{}{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			credSvc := new(MockCredentialService)
			queueMock := new(MockMessageQueue)

			if tt.setupMocks != nil {
				tt.setupMocks(credSvc, queueMock)
			}

			action := &SendMessageAction{
				config:            tt.config,
				credentialService: credSvc,
				queueFactory: func(ctx context.Context, config messaging.Config) (messaging.MessageQueue, error) {
					return queueMock, nil
				},
			}

			ctx := context.Background()
			output, err := action.Execute(ctx, tt.input)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, output)

				for _, key := range tt.wantOutputKeys {
					assert.Contains(t, output, key)
				}

				if success, ok := output["success"].(bool); ok {
					assert.True(t, success)
				}
			}

			credSvc.AssertExpectations(t)
			queueMock.AssertExpectations(t)
		})
	}
}

func TestSendMessageAction_ExecuteWithExpressions(t *testing.T) {
	credSvc := new(MockCredentialService)
	queueMock := new(MockMessageQueue)

	config := SendMessageConfig{
		QueueType:    "sqs",
		Destination:  "{{input.queue_url}}",
		Message:      `{"user_id": "{{input.user_id}}", "status": "{{input.status}}"}`,
		CredentialID: "cred-123",
		Attributes: map[string]string{
			"correlation_id": "{{input.correlation_id}}",
		},
	}

	input := map[string]interface{}{
		"tenant_id":      "tenant-1",
		"queue_url":      "https://sqs.us-east-1.amazonaws.com/123456789/test-queue",
		"user_id":        "user-456",
		"status":         "active",
		"correlation_id": "corr-789",
	}

	credSvc.On("GetCredentialValue", mock.Anything, "tenant-1", "cred-123").
		Return(map[string]interface{}{
			"region":            "us-east-1",
			"access_key_id":     "AKIAIOSFODNN7EXAMPLE",
			"secret_access_key": "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
		}, nil)

	queueMock.On("Send",
		mock.Anything,
		"https://sqs.us-east-1.amazonaws.com/123456789/test-queue",
		mock.MatchedBy(func(msg []byte) bool {
			return string(msg) == `{"user_id": "user-456", "status": "active"}`
		}),
		map[string]string{"correlation_id": "corr-789"},
	).Return(nil)

	queueMock.On("Close").Return(nil)

	action := &SendMessageAction{
		config:            config,
		credentialService: credSvc,
		queueFactory: func(ctx context.Context, config messaging.Config) (messaging.MessageQueue, error) {
			return queueMock, nil
		},
	}

	ctx := context.Background()
	output, err := action.Execute(ctx, input)

	require.NoError(t, err)
	assert.NotNil(t, output)
	assert.True(t, output["success"].(bool))

	credSvc.AssertExpectations(t)
	queueMock.AssertExpectations(t)
}
