package messaging

import (
	"context"
	"testing"
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockKafkaWriter is a mock implementation of the Kafka writer interface
type MockKafkaWriter struct {
	mock.Mock
}

func (m *MockKafkaWriter) WriteMessages(ctx context.Context, msgs ...kafka.Message) error {
	args := m.Called(ctx, msgs)
	return args.Error(0)
}

func (m *MockKafkaWriter) Close() error {
	args := m.Called()
	return args.Error(0)
}

// MockKafkaReader is a mock implementation of the Kafka reader interface
type MockKafkaReader struct {
	mock.Mock
}

func (m *MockKafkaReader) FetchMessage(ctx context.Context) (kafka.Message, error) {
	args := m.Called(ctx)
	return args.Get(0).(kafka.Message), args.Error(1)
}

func (m *MockKafkaReader) CommitMessages(ctx context.Context, msgs ...kafka.Message) error {
	args := m.Called(ctx, msgs)
	return args.Error(0)
}

func (m *MockKafkaReader) Close() error {
	args := m.Called()
	return args.Error(0)
}

func TestKafkaQueue_Send(t *testing.T) {
	tests := []struct {
		name        string
		destination string
		message     []byte
		attributes  map[string]string
		mockSetup   func(*MockKafkaWriter)
		wantErr     bool
	}{
		{
			name:        "successful send",
			destination: "test-topic",
			message:     []byte(`{"test": "data"}`),
			attributes: map[string]string{
				"priority": "high",
				"source":   "workflow-1",
			},
			mockSetup: func(m *MockKafkaWriter) {
				m.On("WriteMessages", mock.Anything, mock.MatchedBy(func(msgs []kafka.Message) bool {
					if len(msgs) != 1 {
						return false
					}
					msg := msgs[0]
					return msg.Topic == "test-topic" &&
						string(msg.Value) == `{"test": "data"}` &&
						len(msg.Headers) == 2
				})).Return(nil)
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
			destination: "test-topic",
			message:     []byte{},
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockWriter := new(MockKafkaWriter)
			if tt.mockSetup != nil {
				tt.mockSetup(mockWriter)
			}

			queue := &KafkaQueue{
				writer:  mockWriter,
				brokers: []string{"localhost:9092"},
			}

			ctx := context.Background()
			err := queue.Send(ctx, tt.destination, tt.message, tt.attributes)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				mockWriter.AssertExpectations(t)
			}
		})
	}
}

func TestKafkaQueue_Receive(t *testing.T) {
	tests := []struct {
		name        string
		source      string
		maxMessages int
		waitTime    time.Duration
		mockSetup   func(*MockKafkaReader)
		wantMsgs    int
		wantErr     bool
	}{
		{
			name:        "successful receive",
			source:      "test-topic",
			maxMessages: 2,
			waitTime:    5 * time.Second,
			mockSetup: func(m *MockKafkaReader) {
				m.On("FetchMessage", mock.Anything).Return(kafka.Message{
					Topic: "test-topic",
					Value: []byte(`{"test": "data1"}`),
					Headers: []kafka.Header{
						{Key: "priority", Value: []byte("high")},
					},
				}, nil).Once()
				m.On("FetchMessage", mock.Anything).Return(kafka.Message{
					Topic: "test-topic",
					Value: []byte(`{"test": "data2"}`),
				}, nil).Once()
			},
			wantMsgs: 2,
			wantErr:  false,
		},
		{
			name:        "invalid max messages",
			source:      "test-topic",
			maxMessages: 0,
			waitTime:    5 * time.Second,
			wantErr:     true,
		},
		{
			name:        "empty source",
			source:      "",
			maxMessages: 1,
			waitTime:    5 * time.Second,
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockReader := new(MockKafkaReader)
			if tt.mockSetup != nil {
				tt.mockSetup(mockReader)
			}

			queue := &KafkaQueue{
				reader:      mockReader,
				brokers:     []string{"localhost:9092"},
				pendingMsgs: make(map[string]kafka.Message),
			}

			ctx := context.Background()
			messages, err := queue.Receive(ctx, tt.source, tt.maxMessages, tt.waitTime)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Len(t, messages, tt.wantMsgs)
				if tt.wantMsgs > 0 {
					assert.NotEmpty(t, messages[0].Body)
				}
				mockReader.AssertExpectations(t)
			}
		})
	}
}

func TestKafkaQueue_Ack(t *testing.T) {
	tests := []struct {
		name      string
		message   Message
		mockSetup func(*MockKafkaReader)
		wantErr   bool
	}{
		{
			name: "successful ack",
			message: Message{
				ID:      "test-topic-0-123",
				Body:    []byte("test"),
				Receipt: "kafka-msg-ref",
			},
			mockSetup: func(m *MockKafkaReader) {
				m.On("CommitMessages", mock.Anything, mock.Anything).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "missing receipt",
			message: Message{
				ID:   "test-topic-0-123",
				Body: []byte("test"),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockReader := new(MockKafkaReader)
			if tt.mockSetup != nil {
				tt.mockSetup(mockReader)
			}

			queue := &KafkaQueue{
				reader:        mockReader,
				brokers:       []string{"localhost:9092"},
				pendingMsgs:   make(map[string]kafka.Message),
				consumerGroup: "test-group",
			}

			if tt.message.Receipt != "" {
				queue.pendingMsgs[tt.message.Receipt] = kafka.Message{
					Topic: "test-topic",
					Value: tt.message.Body,
				}
			}

			ctx := context.Background()
			err := queue.Ack(ctx, tt.message)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				mockReader.AssertExpectations(t)
			}
		})
	}
}

func TestKafkaQueue_Nack(t *testing.T) {
	tests := []struct {
		name    string
		message Message
		wantErr bool
	}{
		{
			name: "nack not supported",
			message: Message{
				ID:      "test-topic-0-123",
				Body:    []byte("test"),
				Receipt: "kafka-msg-ref",
			},
			wantErr: false, // Kafka doesn't support nack, but we don't error
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			queue := &KafkaQueue{
				brokers: []string{"localhost:9092"},
			}

			ctx := context.Background()
			err := queue.Nack(ctx, tt.message)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestKafkaQueue_GetInfo(t *testing.T) {
	tests := []struct {
		name      string
		topicName string
		wantErr   bool
	}{
		{
			name:      "get info returns basic data",
			topicName: "test-topic",
			wantErr:   false,
		},
		{
			name:      "empty topic name",
			topicName: "",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			queue := &KafkaQueue{
				brokers: []string{"localhost:9092"},
			}

			ctx := context.Background()
			info, err := queue.GetInfo(ctx, tt.topicName)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, info)
				assert.Equal(t, tt.topicName, info.Name)
			}
		})
	}
}

func TestKafkaQueue_Close(t *testing.T) {
	mockWriter := new(MockKafkaWriter)
	mockReader := new(MockKafkaReader)

	mockWriter.On("Close").Return(nil)
	mockReader.On("Close").Return(nil)

	queue := &KafkaQueue{
		writer:  mockWriter,
		reader:  mockReader,
		brokers: []string{"localhost:9092"},
	}

	err := queue.Close()
	require.NoError(t, err)

	mockWriter.AssertExpectations(t)
	mockReader.AssertExpectations(t)
}
