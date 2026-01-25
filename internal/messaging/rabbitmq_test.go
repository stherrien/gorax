package messaging

import (
	"context"
	"testing"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockAMQPChannel is a mock implementation of the AMQP channel interface
type MockAMQPChannel struct {
	mock.Mock
}

func (m *MockAMQPChannel) Publish(exchange, key string, mandatory, immediate bool, msg amqp.Publishing) error {
	args := m.Called(exchange, key, mandatory, immediate, msg)
	return args.Error(0)
}

func (m *MockAMQPChannel) Consume(queue, consumer string, autoAck, exclusive, noLocal, noWait bool, args amqp.Table) (<-chan amqp.Delivery, error) {
	callArgs := m.Called(queue, consumer, autoAck, exclusive, noLocal, noWait, args)
	if callArgs.Get(0) == nil {
		return nil, callArgs.Error(1)
	}
	return callArgs.Get(0).(<-chan amqp.Delivery), callArgs.Error(1)
}

func (m *MockAMQPChannel) Ack(tag uint64, multiple bool) error {
	args := m.Called(tag, multiple)
	return args.Error(0)
}

func (m *MockAMQPChannel) Nack(tag uint64, multiple, requeue bool) error {
	args := m.Called(tag, multiple, requeue)
	return args.Error(0)
}

func (m *MockAMQPChannel) QueueDeclare(name string, durable, autoDelete, exclusive, noWait bool, args amqp.Table) (amqp.Queue, error) {
	callArgs := m.Called(name, durable, autoDelete, exclusive, noWait, args)
	return callArgs.Get(0).(amqp.Queue), callArgs.Error(1)
}

func (m *MockAMQPChannel) Close() error {
	args := m.Called()
	return args.Error(0)
}

func TestRabbitMQQueue_Send(t *testing.T) {
	tests := []struct {
		name        string
		destination string
		message     []byte
		attributes  map[string]string
		mockSetup   func(*MockAMQPChannel)
		wantErr     bool
	}{
		{
			name:        "successful send",
			destination: "test-queue",
			message:     []byte(`{"test": "data"}`),
			attributes: map[string]string{
				"priority": "high",
				"source":   "workflow-1",
			},
			mockSetup: func(m *MockAMQPChannel) {
				m.On("Publish", "", "test-queue", false, false, mock.MatchedBy(func(msg amqp.Publishing) bool {
					return string(msg.Body) == `{"test": "data"}` &&
						msg.Headers["priority"] == "high" &&
						msg.Headers["source"] == "workflow-1"
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
			destination: "test-queue",
			message:     []byte{},
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockChannel := new(MockAMQPChannel)
			if tt.mockSetup != nil {
				tt.mockSetup(mockChannel)
			}

			queue := &RabbitMQQueue{
				channel: mockChannel,
				url:     "amqp://guest:guest@localhost:5672/",
			}

			ctx := context.Background()
			err := queue.Send(ctx, tt.destination, tt.message, tt.attributes)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				mockChannel.AssertExpectations(t)
			}
		})
	}
}

func TestRabbitMQQueue_Receive(t *testing.T) {
	tests := []struct {
		name        string
		source      string
		maxMessages int
		waitTime    time.Duration
		mockSetup   func(*MockAMQPChannel, chan amqp.Delivery)
		deliveries  []amqp.Delivery
		wantMsgs    int
		wantErr     bool
	}{
		{
			name:        "successful receive",
			source:      "test-queue",
			maxMessages: 2,
			waitTime:    5 * time.Second,
			mockSetup: func(m *MockAMQPChannel, deliveryChan chan amqp.Delivery) {
				m.On("Consume", "test-queue", "", false, false, false, false, mock.Anything).
					Return((<-chan amqp.Delivery)(deliveryChan), nil)
			},
			deliveries: []amqp.Delivery{
				{
					MessageId:   "msg-1",
					Body:        []byte(`{"test": "data1"}`),
					DeliveryTag: 1,
					Headers: amqp.Table{
						"priority": "high",
					},
				},
				{
					MessageId:   "msg-2",
					Body:        []byte(`{"test": "data2"}`),
					DeliveryTag: 2,
				},
			},
			wantMsgs: 2,
			wantErr:  false,
		},
		{
			name:        "invalid max messages",
			source:      "test-queue",
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
			mockChannel := new(MockAMQPChannel)
			deliveryChan := make(chan amqp.Delivery, len(tt.deliveries))

			if tt.mockSetup != nil {
				tt.mockSetup(mockChannel, deliveryChan)

				// Send deliveries to channel
				go func() {
					for _, d := range tt.deliveries {
						deliveryChan <- d
					}
				}()
			}

			queue := &RabbitMQQueue{
				channel: mockChannel,
				url:     "amqp://guest:guest@localhost:5672/",
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
					assert.NotEmpty(t, messages[0].Receipt)
				}
				mockChannel.AssertExpectations(t)
			}

			close(deliveryChan)
		})
	}
}

func TestRabbitMQQueue_Ack(t *testing.T) {
	tests := []struct {
		name      string
		message   Message
		mockSetup func(*MockAMQPChannel)
		wantErr   bool
	}{
		{
			name: "successful ack",
			message: Message{
				ID:      "msg-1",
				Receipt: "5", // delivery tag as string
			},
			mockSetup: func(m *MockAMQPChannel) {
				m.On("Ack", uint64(5), false).Return(nil)
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
		{
			name: "invalid receipt",
			message: Message{
				ID:      "msg-1",
				Receipt: "invalid",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockChannel := new(MockAMQPChannel)
			if tt.mockSetup != nil {
				tt.mockSetup(mockChannel)
			}

			queue := &RabbitMQQueue{
				channel: mockChannel,
				url:     "amqp://guest:guest@localhost:5672/",
			}

			ctx := context.Background()
			err := queue.Ack(ctx, tt.message)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				mockChannel.AssertExpectations(t)
			}
		})
	}
}

func TestRabbitMQQueue_Nack(t *testing.T) {
	tests := []struct {
		name      string
		message   Message
		mockSetup func(*MockAMQPChannel)
		wantErr   bool
	}{
		{
			name: "successful nack",
			message: Message{
				ID:      "msg-1",
				Receipt: "5",
			},
			mockSetup: func(m *MockAMQPChannel) {
				m.On("Nack", uint64(5), false, true).Return(nil)
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
			mockChannel := new(MockAMQPChannel)
			if tt.mockSetup != nil {
				tt.mockSetup(mockChannel)
			}

			queue := &RabbitMQQueue{
				channel: mockChannel,
				url:     "amqp://guest:guest@localhost:5672/",
			}

			ctx := context.Background()
			err := queue.Nack(ctx, tt.message)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				mockChannel.AssertExpectations(t)
			}
		})
	}
}

func TestRabbitMQQueue_GetInfo(t *testing.T) {
	tests := []struct {
		name      string
		queueName string
		mockSetup func(*MockAMQPChannel)
		wantErr   bool
	}{
		{
			name:      "successful get info",
			queueName: "test-queue",
			mockSetup: func(m *MockAMQPChannel) {
				m.On("QueueDeclare", "test-queue", true, false, false, false, mock.Anything).
					Return(amqp.Queue{
						Name:      "test-queue",
						Messages:  42,
						Consumers: 2,
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
			mockChannel := new(MockAMQPChannel)
			if tt.mockSetup != nil {
				tt.mockSetup(mockChannel)
			}

			queue := &RabbitMQQueue{
				channel: mockChannel,
				url:     "amqp://guest:guest@localhost:5672/",
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
				mockChannel.AssertExpectations(t)
			}
		})
	}
}

func TestRabbitMQQueue_Close(t *testing.T) {
	mockChannel := new(MockAMQPChannel)
	mockChannel.On("Close").Return(nil)

	queue := &RabbitMQQueue{
		channel: mockChannel,
		url:     "amqp://guest:guest@localhost:5672/",
	}

	err := queue.Close()
	require.NoError(t, err)

	mockChannel.AssertExpectations(t)
}
