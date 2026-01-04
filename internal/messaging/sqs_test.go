package messaging

import (
	"context"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockSQSClient is a mock implementation of the SQS client interface
type MockSQSClient struct {
	mock.Mock
}

func (m *MockSQSClient) SendMessage(input *sqs.SendMessageInput) (*sqs.SendMessageOutput, error) {
	args := m.Called(input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*sqs.SendMessageOutput), args.Error(1)
}

func (m *MockSQSClient) ReceiveMessage(input *sqs.ReceiveMessageInput) (*sqs.ReceiveMessageOutput, error) {
	args := m.Called(input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*sqs.ReceiveMessageOutput), args.Error(1)
}

func (m *MockSQSClient) DeleteMessage(input *sqs.DeleteMessageInput) (*sqs.DeleteMessageOutput, error) {
	args := m.Called(input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*sqs.DeleteMessageOutput), args.Error(1)
}

func (m *MockSQSClient) ChangeMessageVisibility(input *sqs.ChangeMessageVisibilityInput) (*sqs.ChangeMessageVisibilityOutput, error) {
	args := m.Called(input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*sqs.ChangeMessageVisibilityOutput), args.Error(1)
}

func (m *MockSQSClient) GetQueueAttributes(input *sqs.GetQueueAttributesInput) (*sqs.GetQueueAttributesOutput, error) {
	args := m.Called(input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*sqs.GetQueueAttributesOutput), args.Error(1)
}

func TestSQSQueue_Send(t *testing.T) {
	tests := []struct {
		name        string
		destination string
		message     []byte
		attributes  map[string]string
		mockSetup   func(*MockSQSClient)
		wantErr     bool
	}{
		{
			name:        "successful send",
			destination: "https://sqs.us-east-1.amazonaws.com/123456789/test-queue",
			message:     []byte(`{"test": "data"}`),
			attributes: map[string]string{
				"priority": "high",
			},
			mockSetup: func(m *MockSQSClient) {
				m.On("SendMessage", mock.MatchedBy(func(input *sqs.SendMessageInput) bool {
					return *input.QueueUrl == "https://sqs.us-east-1.amazonaws.com/123456789/test-queue" &&
						*input.MessageBody == `{"test": "data"}` &&
						*input.MessageAttributes["priority"].StringValue == "high"
				})).Return(&sqs.SendMessageOutput{
					MessageId: aws.String("msg-123"),
				}, nil)
			},
			wantErr: false,
		},
		{
			name:        "empty destination",
			destination: "",
			message:     []byte("test"),
			wantErr:     true,
		},
		{
			name:        "empty message",
			destination: "https://sqs.us-east-1.amazonaws.com/123456789/test-queue",
			message:     []byte{},
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := new(MockSQSClient)
			if tt.mockSetup != nil {
				tt.mockSetup(mockClient)
			}

			queue := &SQSQueue{
				client: mockClient,
				region: "us-east-1",
			}

			ctx := context.Background()
			err := queue.Send(ctx, tt.destination, tt.message, tt.attributes)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				mockClient.AssertExpectations(t)
			}
		})
	}
}

func TestSQSQueue_Receive(t *testing.T) {
	tests := []struct {
		name        string
		source      string
		maxMessages int
		waitTime    time.Duration
		mockSetup   func(*MockSQSClient)
		wantMsgs    int
		wantErr     bool
	}{
		{
			name:        "successful receive",
			source:      "https://sqs.us-east-1.amazonaws.com/123456789/test-queue",
			maxMessages: 10,
			waitTime:    5 * time.Second,
			mockSetup: func(m *MockSQSClient) {
				m.On("ReceiveMessage", mock.MatchedBy(func(input *sqs.ReceiveMessageInput) bool {
					return *input.QueueUrl == "https://sqs.us-east-1.amazonaws.com/123456789/test-queue" &&
						*input.MaxNumberOfMessages == int64(10) &&
						*input.WaitTimeSeconds == int64(5)
				})).Return(&sqs.ReceiveMessageOutput{
					Messages: []*sqs.Message{
						{
							MessageId:     aws.String("msg-1"),
							Body:          aws.String(`{"test": "data1"}`),
							ReceiptHandle: aws.String("receipt-1"),
							MessageAttributes: map[string]*sqs.MessageAttributeValue{
								"priority": {
									StringValue: aws.String("high"),
								},
							},
						},
						{
							MessageId:     aws.String("msg-2"),
							Body:          aws.String(`{"test": "data2"}`),
							ReceiptHandle: aws.String("receipt-2"),
						},
					},
				}, nil)
			},
			wantMsgs: 2,
			wantErr:  false,
		},
		{
			name:        "empty queue",
			source:      "https://sqs.us-east-1.amazonaws.com/123456789/test-queue",
			maxMessages: 10,
			waitTime:    5 * time.Second,
			mockSetup: func(m *MockSQSClient) {
				m.On("ReceiveMessage", mock.Anything).Return(&sqs.ReceiveMessageOutput{
					Messages: []*sqs.Message{},
				}, nil)
			},
			wantMsgs: 0,
			wantErr:  false,
		},
		{
			name:        "invalid max messages",
			source:      "https://sqs.us-east-1.amazonaws.com/123456789/test-queue",
			maxMessages: 0,
			waitTime:    5 * time.Second,
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := new(MockSQSClient)
			if tt.mockSetup != nil {
				tt.mockSetup(mockClient)
			}

			queue := &SQSQueue{
				client: mockClient,
				region: "us-east-1",
			}

			ctx := context.Background()
			messages, err := queue.Receive(ctx, tt.source, tt.maxMessages, tt.waitTime)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Len(t, messages, tt.wantMsgs)
				if tt.wantMsgs > 0 {
					assert.NotEmpty(t, messages[0].ID)
					assert.NotEmpty(t, messages[0].Body)
					assert.NotEmpty(t, messages[0].Receipt)
				}
				mockClient.AssertExpectations(t)
			}
		})
	}
}

func TestSQSQueue_Ack(t *testing.T) {
	tests := []struct {
		name      string
		message   Message
		mockSetup func(*MockSQSClient)
		wantErr   bool
	}{
		{
			name: "successful ack",
			message: Message{
				ID:      "msg-1",
				Receipt: "receipt-1",
			},
			mockSetup: func(m *MockSQSClient) {
				m.On("DeleteMessage", mock.MatchedBy(func(input *sqs.DeleteMessageInput) bool {
					return *input.ReceiptHandle == "receipt-1"
				})).Return(&sqs.DeleteMessageOutput{}, nil)
			},
			wantErr: false,
		},
		{
			name: "missing receipt",
			message: Message{
				ID: "msg-1",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := new(MockSQSClient)
			if tt.mockSetup != nil {
				tt.mockSetup(mockClient)
			}

			queue := &SQSQueue{
				client:   mockClient,
				region:   "us-east-1",
				queueURL: "https://sqs.us-east-1.amazonaws.com/123456789/test-queue",
			}

			ctx := context.Background()
			err := queue.Ack(ctx, tt.message)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				mockClient.AssertExpectations(t)
			}
		})
	}
}

func TestSQSQueue_Nack(t *testing.T) {
	tests := []struct {
		name      string
		message   Message
		mockSetup func(*MockSQSClient)
		wantErr   bool
	}{
		{
			name: "successful nack",
			message: Message{
				ID:      "msg-1",
				Receipt: "receipt-1",
			},
			mockSetup: func(m *MockSQSClient) {
				m.On("ChangeMessageVisibility", mock.MatchedBy(func(input *sqs.ChangeMessageVisibilityInput) bool {
					return *input.ReceiptHandle == "receipt-1" &&
						*input.VisibilityTimeout == int64(0)
				})).Return(&sqs.ChangeMessageVisibilityOutput{}, nil)
			},
			wantErr: false,
		},
		{
			name: "missing receipt",
			message: Message{
				ID: "msg-1",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := new(MockSQSClient)
			if tt.mockSetup != nil {
				tt.mockSetup(mockClient)
			}

			queue := &SQSQueue{
				client:   mockClient,
				region:   "us-east-1",
				queueURL: "https://sqs.us-east-1.amazonaws.com/123456789/test-queue",
			}

			ctx := context.Background()
			err := queue.Nack(ctx, tt.message)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				mockClient.AssertExpectations(t)
			}
		})
	}
}

func TestSQSQueue_GetInfo(t *testing.T) {
	tests := []struct {
		name      string
		queueName string
		mockSetup func(*MockSQSClient)
		wantErr   bool
	}{
		{
			name:      "successful get info",
			queueName: "https://sqs.us-east-1.amazonaws.com/123456789/test-queue",
			mockSetup: func(m *MockSQSClient) {
				m.On("GetQueueAttributes", mock.MatchedBy(func(input *sqs.GetQueueAttributesInput) bool {
					return *input.QueueUrl == "https://sqs.us-east-1.amazonaws.com/123456789/test-queue"
				})).Return(&sqs.GetQueueAttributesOutput{
					Attributes: map[string]*string{
						"ApproximateNumberOfMessages": aws.String("42"),
						"CreatedTimestamp":            aws.String("1640000000"),
					},
				}, nil)
			},
			wantErr: false,
		},
		{
			name:      "empty queue name",
			queueName: "",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := new(MockSQSClient)
			if tt.mockSetup != nil {
				tt.mockSetup(mockClient)
			}

			queue := &SQSQueue{
				client: mockClient,
				region: "us-east-1",
			}

			ctx := context.Background()
			info, err := queue.GetInfo(ctx, tt.queueName)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, info)
				assert.Equal(t, tt.queueName, info.Name)
				assert.Equal(t, 42, info.ApproximateCount)
				mockClient.AssertExpectations(t)
			}
		})
	}
}
