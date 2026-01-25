//go:build integration
// +build integration

package messaging

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSQSIntegration tests SQS queue integration with LocalStack or real AWS
func TestSQSIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	region := os.Getenv("TEST_SQS_REGION")
	if region == "" {
		region = "us-east-1"
	}

	ctx := context.Background()
	config := Config{
		Type:   QueueTypeSQS,
		Region: region,
	}

	queue, err := NewSQSQueue(ctx, config)
	require.NoError(t, err)
	defer queue.Close()

	// Test queue URL (use LocalStack or real AWS queue)
	queueURL := os.Getenv("TEST_SQS_QUEUE_URL")
	if queueURL == "" {
		t.Skip("TEST_SQS_QUEUE_URL not set")
	}

	// Set queue URL for ack/nack
	queue.SetQueueURL(queueURL)

	t.Run("send and receive message", func(t *testing.T) {
		// Send a test message
		testMessage := []byte(`{"test": "integration", "timestamp": "` + time.Now().Format(time.RFC3339) + `"}`)
		attributes := map[string]string{
			"test": "true",
			"env":  "integration",
		}

		err := queue.Send(ctx, queueURL, testMessage, attributes)
		require.NoError(t, err)

		// Wait a bit for message to be available
		time.Sleep(2 * time.Second)

		// Receive the message
		messages, err := queue.Receive(ctx, queueURL, 10, 5*time.Second)
		require.NoError(t, err)
		assert.NotEmpty(t, messages)

		// Verify message content
		found := false
		var receivedMsg Message
		for _, msg := range messages {
			if string(msg.Body) == string(testMessage) {
				found = true
				receivedMsg = msg
				break
			}
		}
		assert.True(t, found, "Test message not found in received messages")

		if found {
			assert.Equal(t, "true", receivedMsg.Attributes["test"])
			assert.Equal(t, "integration", receivedMsg.Attributes["env"])

			// Acknowledge the message
			err = queue.Ack(ctx, receivedMsg)
			require.NoError(t, err)
		}
	})
}

// TestKafkaIntegration tests Kafka queue integration
func TestKafkaIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	brokers := os.Getenv("TEST_KAFKA_BROKERS")
	if brokers == "" {
		brokers = "localhost:9092"
	}

	ctx := context.Background()
	config := Config{
		Type:    QueueTypeKafka,
		Brokers: []string{brokers},
	}

	queue, err := NewKafkaQueue(ctx, config)
	require.NoError(t, err)
	defer queue.Close()

	// Set consumer group
	kafkaQueue := queue.(*KafkaQueue)
	kafkaQueue.SetConsumerGroup("gorax-integration-test")

	topic := "test-topic-integration"

	t.Run("send and receive message", func(t *testing.T) {
		// Send a test message
		testMessage := []byte(`{"test": "kafka-integration", "timestamp": "` + time.Now().Format(time.RFC3339) + `"}`)
		attributes := map[string]string{
			"test":   "true",
			"source": "integration-test",
		}

		err := queue.Send(ctx, topic, testMessage, attributes)
		require.NoError(t, err)

		// Receive the message
		messages, err := queue.Receive(ctx, topic, 10, 5*time.Second)
		require.NoError(t, err)
		assert.NotEmpty(t, messages)

		// Verify message content
		msg := messages[0]
		assert.Equal(t, string(testMessage), string(msg.Body))
		assert.Equal(t, "true", msg.Attributes["test"])
		assert.Equal(t, "integration-test", msg.Attributes["source"])

		// Acknowledge the message
		err = queue.Ack(ctx, msg)
		require.NoError(t, err)
	})
}

// TestRabbitMQIntegration tests RabbitMQ queue integration
func TestRabbitMQIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	url := os.Getenv("TEST_RABBITMQ_URL")
	if url == "" {
		url = "amqp://guest:guest@localhost:5672/"
	}

	ctx := context.Background()
	config := Config{
		Type: QueueTypeRabbitMQ,
		URL:  url,
	}

	queue, err := NewRabbitMQQueue(ctx, config)
	require.NoError(t, err)
	defer queue.Close()

	queueName := "gorax-integration-test"

	t.Run("send and receive message", func(t *testing.T) {
		// Send a test message
		testMessage := []byte(`{"test": "rabbitmq-integration", "timestamp": "` + time.Now().Format(time.RFC3339) + `"}`)
		attributes := map[string]string{
			"test":     "true",
			"priority": "high",
		}

		err := queue.Send(ctx, queueName, testMessage, attributes)
		require.NoError(t, err)

		// Receive the message
		messages, err := queue.Receive(ctx, queueName, 10, 5*time.Second)
		require.NoError(t, err)
		assert.NotEmpty(t, messages)

		// Verify message content
		msg := messages[0]
		assert.Equal(t, string(testMessage), string(msg.Body))
		assert.Equal(t, "true", msg.Attributes["test"])
		assert.Equal(t, "high", msg.Attributes["priority"])

		// Acknowledge the message
		err = queue.Ack(ctx, msg)
		require.NoError(t, err)
	})

	t.Run("nack and requeue message", func(t *testing.T) {
		// Send a test message
		testMessage := []byte(`{"test": "nack-test"}`)

		err := queue.Send(ctx, queueName, testMessage, nil)
		require.NoError(t, err)

		// Receive the message
		messages, err := queue.Receive(ctx, queueName, 1, 5*time.Second)
		require.NoError(t, err)
		require.Len(t, messages, 1)

		msg := messages[0]

		// Nack the message (should requeue)
		err = queue.Nack(ctx, msg)
		require.NoError(t, err)

		// Receive again to verify it was requeued
		messages, err = queue.Receive(ctx, queueName, 1, 5*time.Second)
		require.NoError(t, err)
		assert.NotEmpty(t, messages)

		// Clean up - acknowledge the message
		err = queue.Ack(ctx, messages[0])
		require.NoError(t, err)
	})
}

// TestQueueInfo tests getting queue information
func TestQueueInfo(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()

	t.Run("SQS queue info", func(t *testing.T) {
		queueURL := os.Getenv("TEST_SQS_QUEUE_URL")
		if queueURL == "" {
			t.Skip("TEST_SQS_QUEUE_URL not set")
		}

		config := Config{
			Type:   QueueTypeSQS,
			Region: "us-east-1",
		}

		queue, err := NewSQSQueue(ctx, config)
		require.NoError(t, err)
		defer queue.Close()

		info, err := queue.GetInfo(ctx, queueURL)
		require.NoError(t, err)
		assert.NotNil(t, info)
		assert.Equal(t, queueURL, info.Name)
	})

	t.Run("Kafka topic info", func(t *testing.T) {
		brokers := os.Getenv("TEST_KAFKA_BROKERS")
		if brokers == "" {
			t.Skip("TEST_KAFKA_BROKERS not set")
		}

		config := Config{
			Type:    QueueTypeKafka,
			Brokers: []string{brokers},
		}

		queue, err := NewKafkaQueue(ctx, config)
		require.NoError(t, err)
		defer queue.Close()

		info, err := queue.GetInfo(ctx, "test-topic")
		require.NoError(t, err)
		assert.NotNil(t, info)
		assert.Equal(t, "test-topic", info.Name)
	})

	t.Run("RabbitMQ queue info", func(t *testing.T) {
		url := os.Getenv("TEST_RABBITMQ_URL")
		if url == "" {
			t.Skip("TEST_RABBITMQ_URL not set")
		}

		config := Config{
			Type: QueueTypeRabbitMQ,
			URL:  url,
		}

		queue, err := NewRabbitMQQueue(ctx, config)
		require.NoError(t, err)
		defer queue.Close()

		info, err := queue.GetInfo(ctx, "gorax-integration-test")
		require.NoError(t, err)
		assert.NotNil(t, info)
		assert.Equal(t, "gorax-integration-test", info.Name)
	})
}
