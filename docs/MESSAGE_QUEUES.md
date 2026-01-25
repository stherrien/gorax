# Message Queue Integrations

Gorax supports integration with three major message queue systems: AWS SQS, Apache Kafka, and RabbitMQ. These integrations enable asynchronous workflow communication, event-driven architectures, and inter-service messaging.

## Table of Contents

- [Overview](#overview)
- [Supported Queue Systems](#supported-queue-systems)
- [Configuration](#configuration)
- [Workflow Actions](#workflow-actions)
- [Credential Management](#credential-management)
- [Usage Examples](#usage-examples)
- [Error Handling](#error-handling)
- [Best Practices](#best-practices)
- [Troubleshooting](#troubleshooting)

## Overview

Message queues provide reliable asynchronous communication between services and workflows. Gorax's message queue integrations allow you to:

- **Send messages** to queues/topics from workflows
- **Receive messages** and trigger workflow actions
- **Implement event-driven patterns** with pub/sub messaging
- **Decouple services** for better scalability and reliability
- **Handle high-throughput** message processing

## Supported Queue Systems

### AWS SQS (Simple Queue Service)

- **Type**: Fully managed message queue service
- **Pattern**: Point-to-point (queue-based)
- **Best for**: Simple queuing, reliable message delivery, AWS ecosystem integration
- **Features**: Dead-letter queues, message visibility timeout, long polling

### Apache Kafka

- **Type**: Distributed streaming platform
- **Pattern**: Pub/sub (topic-based)
- **Best for**: Event streaming, high-throughput, log aggregation
- **Features**: Message persistence, consumer groups, partition-based parallelism

### RabbitMQ

- **Type**: Message broker with AMQP protocol
- **Pattern**: Both point-to-point and pub/sub
- **Best for**: Complex routing, flexible messaging patterns, enterprise messaging
- **Features**: Exchanges, routing keys, message acknowledgment, requeue

## Configuration

### Environment Variables

Add these configurations to your `.env` file:

```bash
# AWS SQS Configuration
MQ_SQS_REGION=us-east-1
MQ_SQS_ACCESS_KEY_ID=your-access-key
MQ_SQS_SECRET_ACCESS_KEY=your-secret-key

# Apache Kafka Configuration
MQ_KAFKA_BROKERS=localhost:9092,localhost:9093
MQ_KAFKA_CONSUMER_GROUP=gorax-consumers

# RabbitMQ Configuration
MQ_RABBITMQ_URL=amqp://username:password@hostname:5672/vhost
```

### Credential Setup

Message queue credentials must be created in Gorax before use:

1. Navigate to **Credentials** in the Gorax UI
2. Click **Create Credential**
3. Select the appropriate credential type:
   - `queue_aws_sqs` for AWS SQS
   - `queue_kafka` for Apache Kafka
   - `queue_rabbitmq` for RabbitMQ
4. Fill in the required fields (see [Credential Management](#credential-management))

## Workflow Actions

### Send Message Action

Sends a message to a queue or topic.

**Configuration:**

```json
{
  "type": "send_message",
  "config": {
    "queue_type": "sqs",
    "destination": "https://sqs.us-east-1.amazonaws.com/123456789/my-queue",
    "message": "{\"user_id\": \"{{input.user_id}}\", \"action\": \"created\"}",
    "attributes": {
      "priority": "high",
      "correlation_id": "{{input.correlation_id}}"
    },
    "credential_id": "cred-123"
  }
}
```

**Fields:**

- `queue_type` (required): Type of queue - `sqs`, `kafka`, or `rabbitmq`
- `destination` (required): Queue URL, topic name, or queue name
- `message` (required): Message body (supports template expressions)
- `attributes` (optional): Message attributes/headers
- `credential_id` (required): ID of the queue credential

**Output:**

```json
{
  "success": true,
  "message_id": "msg-abc123",
  "queue_type": "sqs",
  "destination": "https://sqs.us-east-1.amazonaws.com/123456789/my-queue",
  "sent_at": "2025-01-02T10:30:00Z"
}
```

### Receive Message Action

Receives messages from a queue or topic.

**Configuration:**

```json
{
  "type": "receive_message",
  "config": {
    "queue_type": "kafka",
    "source": "user-events",
    "max_messages": 10,
    "wait_time": "5s",
    "delete_after": true,
    "credential_id": "cred-456"
  }
}
```

**Fields:**

- `queue_type` (required): Type of queue - `sqs`, `kafka`, or `rabbitmq`
- `source` (required): Queue URL, topic name, or queue name
- `max_messages` (optional): Maximum messages to receive (default: 10)
- `wait_time` (optional): Wait time for messages (default: 5s)
- `delete_after` (optional): Auto-acknowledge after receiving (default: false)
- `credential_id` (required): ID of the queue credential

**Output:**

```json
{
  "success": true,
  "message_count": 2,
  "messages": [
    {
      "id": "msg-1",
      "body": "{\"user_id\": \"user-123\", \"action\": \"created\"}",
      "attributes": {
        "priority": "high"
      },
      "receipt": "receipt-handle-1",
      "timestamp": "2025-01-02T10:29:55Z"
    },
    {
      "id": "msg-2",
      "body": "{\"user_id\": \"user-456\", \"action\": \"updated\"}",
      "attributes": {},
      "receipt": "receipt-handle-2",
      "timestamp": "2025-01-02T10:29:58Z"
    }
  ],
  "queue_type": "kafka",
  "source": "user-events",
  "received_at": "2025-01-02T10:30:00Z"
}
```

## Credential Management

### AWS SQS Credentials

**Required Fields:**

```json
{
  "region": "us-east-1",
  "access_key_id": "AKIAIOSFODNN7EXAMPLE",
  "secret_access_key": "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY"
}
```

**IAM Permissions Required:**

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "sqs:SendMessage",
        "sqs:ReceiveMessage",
        "sqs:DeleteMessage",
        "sqs:GetQueueAttributes"
      ],
      "Resource": "arn:aws:sqs:us-east-1:123456789:my-queue"
    }
  ]
}
```

### Apache Kafka Credentials

**Required Fields:**

```json
{
  "brokers": ["kafka1.example.com:9092", "kafka2.example.com:9092"],
  "consumer_group": "gorax-workflow-group"
}
```

**Optional Fields:**

```json
{
  "sasl_mechanism": "PLAIN",
  "sasl_username": "user",
  "sasl_password": "password",
  "enable_tls": true
}
```

### RabbitMQ Credentials

**Required Fields:**

```json
{
  "url": "amqp://username:password@rabbitmq.example.com:5672/vhost"
}
```

**URL Format:**

```
amqp://username:password@hostname:port/vhost
```

For TLS:

```
amqps://username:password@hostname:5671/vhost
```

## Usage Examples

### Example 1: Send User Event to SQS

```json
{
  "id": "action-1",
  "type": "send_message",
  "name": "Send User Created Event",
  "config": {
    "queue_type": "sqs",
    "destination": "https://sqs.us-east-1.amazonaws.com/123456789/user-events",
    "message": "{\"event_type\": \"user_created\", \"user_id\": \"{{input.user_id}}\", \"email\": \"{{input.email}}\", \"timestamp\": \"{{input.timestamp}}\"}",
    "attributes": {
      "event_type": "user_created",
      "source": "user-service"
    },
    "credential_id": "sqs-cred-123"
  }
}
```

### Example 2: Process Kafka Stream

```json
{
  "id": "action-2",
  "type": "receive_message",
  "name": "Process Order Events",
  "config": {
    "queue_type": "kafka",
    "source": "orders",
    "max_messages": 100,
    "wait_time": "10s",
    "delete_after": false,
    "credential_id": "kafka-cred-456"
  }
}
```

### Example 3: RabbitMQ Task Queue

```json
{
  "id": "action-3",
  "type": "send_message",
  "name": "Queue Background Job",
  "config": {
    "queue_type": "rabbitmq",
    "destination": "background-tasks",
    "message": "{\"task_type\": \"{{input.task_type}}\", \"payload\": {{input.payload}}}",
    "attributes": {
      "priority": "{{input.priority}}",
      "retry_count": "0"
    },
    "credential_id": "rabbitmq-cred-789"
  }
}
```

### Example 4: Event-Driven Workflow

**Workflow: User Registration with Event Publishing**

```json
{
  "nodes": [
    {
      "id": "trigger-1",
      "type": "webhook",
      "name": "User Registration Webhook"
    },
    {
      "id": "action-1",
      "type": "http",
      "name": "Create User in Database",
      "config": {
        "method": "POST",
        "url": "https://api.example.com/users",
        "body": "{\"email\": \"{{trigger.body.email}}\", \"name\": \"{{trigger.body.name}}\"}"
      }
    },
    {
      "id": "action-2",
      "type": "send_message",
      "name": "Publish User Created Event",
      "config": {
        "queue_type": "kafka",
        "destination": "user-events",
        "message": "{\"event\": \"user_created\", \"user_id\": \"{{action-1.response.id}}\", \"email\": \"{{action-1.response.email}}\"}",
        "attributes": {
          "event_type": "user_created",
          "timestamp": "{{now}}"
        },
        "credential_id": "kafka-cred"
      }
    }
  ],
  "edges": [
    {"from": "trigger-1", "to": "action-1"},
    {"from": "action-1", "to": "action-2"}
  ]
}
```

## Error Handling

### Common Errors

#### Connection Errors

```json
{
  "error": "failed to create queue client: connection refused",
  "code": "CONNECTION_FAILED"
}
```

**Solutions:**
- Verify queue service is running
- Check network connectivity
- Validate credentials
- Verify firewall rules

#### Authentication Errors

```json
{
  "error": "failed to get credential: unauthorized access to credential",
  "code": "CREDENTIAL_NOT_FOUND"
}
```

**Solutions:**
- Verify credential exists
- Check credential permissions
- Ensure credential is active
- Validate credential values

#### Message Send Errors

```json
{
  "error": "failed to send message: queue not found",
  "code": "QUEUE_NOT_FOUND"
}
```

**Solutions:**
- Verify queue/topic exists
- Check destination spelling
- Ensure proper permissions
- Create queue/topic if needed

### Retry Strategy

Message queue actions automatically retry on transient failures:

- **SQS**: Uses exponential backoff with AWS SDK defaults
- **Kafka**: Configurable retry policy in producer settings
- **RabbitMQ**: Connection retry with exponential backoff

## Best Practices

### 1. Use Appropriate Queue Type

- **SQS**: Simple queuing, AWS ecosystem, serverless
- **Kafka**: Event streaming, high throughput, persistence
- **RabbitMQ**: Complex routing, flexible patterns, enterprise

### 2. Message Format

Use JSON for message bodies:

```json
{
  "event_type": "user_created",
  "timestamp": "2025-01-02T10:30:00Z",
  "data": {
    "user_id": "user-123",
    "email": "user@example.com"
  },
  "metadata": {
    "source": "user-service",
    "version": "1.0"
  }
}
```

### 3. Message Attributes

Use attributes for filtering and routing:

```json
{
  "attributes": {
    "event_type": "user_created",
    "priority": "high",
    "source": "user-service",
    "correlation_id": "correlation-123"
  }
}
```

### 4. Error Handling

Always handle message processing errors:

```json
{
  "nodes": [
    {
      "id": "receive-1",
      "type": "receive_message",
      "name": "Receive Messages"
    },
    {
      "id": "process-1",
      "type": "http",
      "name": "Process Message"
    },
    {
      "id": "error-handler",
      "type": "condition",
      "condition": "{{process-1.status}} != 200",
      "onTrue": "dead-letter-queue"
    }
  ]
}
```

### 5. Idempotency

Design message handlers to be idempotent:

- Use unique message IDs
- Check for duplicate processing
- Use database transactions
- Implement idempotency keys

### 6. Monitoring

Track message processing metrics:

- Message send/receive rates
- Processing latency
- Error rates
- Queue depth

### 7. Security

- Use encrypted connections (TLS/SSL)
- Rotate credentials regularly
- Apply least-privilege permissions
- Encrypt sensitive message data

### 8. Performance

**SQS:**
- Use batch operations for multiple messages
- Enable long polling (reduce API calls)
- Set appropriate visibility timeout

**Kafka:**
- Configure consumer groups for parallelism
- Tune batch size and linger time
- Use compression for large messages

**RabbitMQ:**
- Use prefetch limits to control concurrency
- Enable publisher confirms for reliability
- Configure queue limits to prevent memory issues

## Troubleshooting

### SQS Issues

**Problem**: Messages not received

**Solutions:**
- Check visibility timeout
- Verify queue policy
- Check dead-letter queue
- Increase wait time

### Kafka Issues

**Problem**: Consumer lag

**Solutions:**
- Increase consumer instances
- Optimize message processing
- Check partition count
- Monitor broker health

**Problem**: Message offset errors

**Solutions:**
- Reset consumer group offset
- Check offset commit strategy
- Verify Kafka cluster health

### RabbitMQ Issues

**Problem**: Messages stuck in queue

**Solutions:**
- Check consumer status
- Verify queue is not full
- Check message TTL
- Inspect dead-letter exchange

**Problem**: Connection failures

**Solutions:**
- Check connection limits
- Verify credentials
- Check network connectivity
- Review RabbitMQ logs

## Integration Testing

To run integration tests with real queue systems:

```bash
# Set up test environment variables
export TEST_SQS_QUEUE_URL=https://sqs.us-east-1.amazonaws.com/123456789/test-queue
export TEST_KAFKA_BROKERS=localhost:9092
export TEST_RABBITMQ_URL=amqp://guest:guest@localhost:5672/

# Run integration tests
go test -tags=integration ./internal/messaging/...
```

### Docker Compose Setup

Use Docker Compose for local development:

```yaml
version: '3.8'
services:
  localstack:
    image: localstack/localstack
    ports:
      - "4566:4566"
    environment:
      - SERVICES=sqs
      - DEBUG=1

  kafka:
    image: confluentinc/cp-kafka:latest
    ports:
      - "9092:9092"
    environment:
      - KAFKA_ZOOKEEPER_CONNECT=zookeeper:2181
      - KAFKA_ADVERTISED_LISTENERS=PLAINTEXT://localhost:9092

  rabbitmq:
    image: rabbitmq:3-management
    ports:
      - "5672:5672"
      - "15672:15672"
    environment:
      - RABBITMQ_DEFAULT_USER=guest
      - RABBITMQ_DEFAULT_PASS=guest
```

## Next Steps

- Explore [Workflow Examples](./WORKFLOW_EXAMPLES.md)
- Learn about [Event-Driven Patterns](./EVENT_DRIVEN_PATTERNS.md)
- Review [Security Best Practices](./SECURITY.md)
- Check [API Documentation](./API.md)
