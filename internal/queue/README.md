# Queue Package

This package provides AWS SQS integration for the gorax project, enabling horizontal scaling of workflow executions through queue-based processing.

## Features

- **AWS SQS Integration**: Uses AWS SDK v2 for Go with support for custom endpoints (LocalStack)
- **Publisher/Consumer Pattern**: Separate components for publishing and consuming messages
- **Dead-Letter Queue Support**: Automatic handling of failed messages with configurable retry limits
- **Message Serialization**: JSON-based message format with validation
- **Queue Depth Monitoring**: Real-time metrics collection for queue health monitoring
- **Concurrent Processing**: Configurable worker pool with per-tenant concurrency limits
- **Long Polling**: Efficient message retrieval using SQS long polling

## Components

### SQS Client (`sqs.go`)
Low-level wrapper around AWS SQS with the following features:
- Send single or batch messages
- Receive messages with long polling
- Delete messages (single or batch)
- Change message visibility
- Get queue attributes and metrics
- Support for LocalStack via custom endpoint

### Message Format (`message.go`)
Defines the execution message structure:
```go
type ExecutionMessage struct {
    ExecutionID     string
    TenantID        string
    WorkflowID      string
    WorkflowVersion int
    TriggerType     string
    TriggerData     json.RawMessage
    EnqueuedAt      time.Time
    RetryCount      int
    CorrelationID   string
}
```

### Publisher (`publisher.go`)
Handles publishing execution messages to SQS:
- Single message publishing
- Batch publishing (up to 10 messages)
- Automatic validation
- Metrics tracking (published, failed counts)

### Consumer (`consumer.go`)
Consumes and processes messages from SQS:
- Configurable concurrency and polling settings
- Automatic retry handling
- Message validation
- Visibility timeout management
- Metrics tracking (received, processed, failed counts)

### Metrics Collector (`metrics.go`)
Collects queue metrics for monitoring:
- Queue depth (visible, in-flight, delayed messages)
- Dead-letter queue depth
- Health status checks
- Periodic metric collection

## Configuration

### Environment Variables

```bash
# Enable queue-based processing
QUEUE_ENABLED=true

# AWS Configuration
AWS_REGION=us-east-1
AWS_ACCESS_KEY_ID=your-access-key
AWS_SECRET_ACCESS_KEY=your-secret-key
AWS_ENDPOINT=http://localhost:4566  # For LocalStack
AWS_SQS_QUEUE_URL=https://sqs.us-east-1.amazonaws.com/123456789/gorax-executions
AWS_SQS_DLQ_URL=https://sqs.us-east-1.amazonaws.com/123456789/gorax-executions-dlq

# Queue Settings
QUEUE_MAX_MESSAGES=10           # Max messages per poll (1-10)
QUEUE_WAIT_TIME_SECONDS=20      # Long polling wait time (0-20)
QUEUE_VISIBILITY_TIMEOUT=30     # Message visibility timeout (seconds)
QUEUE_MAX_RETRIES=3             # Max retries before DLQ
QUEUE_PROCESS_TIMEOUT=300       # Max processing time per message (seconds)
QUEUE_POLL_INTERVAL=1           # Interval between polls when no messages (seconds)
QUEUE_CONCURRENT_WORKERS=10     # Number of concurrent message processors
QUEUE_DELETE_AFTER_PROCESS=true # Auto-delete messages after successful processing
```

## Usage

### Setting Up the Publisher (API Server)

```go
import (
    "github.com/gorax/gorax/internal/queue"
)

// Create SQS client
sqsClient, err := queue.NewSQSClient(ctx, queue.SQSConfig{
    QueueURL:        cfg.AWS.SQSQueueURL,
    DLQueueURL:      cfg.AWS.SQSDLQueueURL,
    Region:          cfg.AWS.Region,
    AccessKeyID:     cfg.AWS.AccessKeyID,
    SecretAccessKey: cfg.AWS.SecretAccessKey,
    Endpoint:        cfg.AWS.Endpoint, // For LocalStack
}, logger)

// Create publisher
publisher := queue.NewPublisher(sqsClient, logger)

// Create adapter for workflow service
publisherAdapter := queue.NewPublisherAdapter(publisher, logger)

// Set publisher in workflow service
workflowService.SetQueuePublisher(publisherAdapter)
```

### Setting Up the Consumer (Worker)

The worker automatically configures the consumer when `QUEUE_ENABLED=true`. See `internal/worker/worker.go` for implementation details.

### Publishing a Message

```go
// Create execution message
msg := queue.NewExecutionMessage(
    executionID,
    tenantID,
    workflowID,
    workflowVersion,
    triggerType,
    triggerData,
)

// Publish to queue
err := publisher.PublishExecution(ctx, msg)
```

### Collecting Metrics

```go
// Create metrics collector
metricsCollector := queue.NewMetricsCollector(sqsClient, 30*time.Second, logger)

// Start collecting metrics
go metricsCollector.Start(ctx)

// Get current metrics
metrics := metricsCollector.GetMetrics()
queueDepth := metricsCollector.GetQueueDepth()
dlqDepth := metricsCollector.GetDLQDepth()

// Check health
healthy := metricsCollector.IsHealthy(1000) // Max queue depth: 1000
status := metricsCollector.GetHealthStatus(1000)
```

## LocalStack Setup

For local development, use LocalStack to simulate AWS SQS:

```bash
# Start LocalStack
docker run -d \
  --name localstack \
  -p 4566:4566 \
  -e SERVICES=sqs \
  localstack/localstack

# Create main queue
aws --endpoint-url=http://localhost:4566 sqs create-queue \
  --queue-name gorax-executions \
  --attributes VisibilityTimeout=30

# Create dead-letter queue
aws --endpoint-url=http://localhost:4566 sqs create-queue \
  --queue-name gorax-executions-dlq

# Configure DLQ on main queue
aws --endpoint-url=http://localhost:4566 sqs set-queue-attributes \
  --queue-url http://localhost:4566/000000000000/gorax-executions \
  --attributes RedrivePolicy='{"deadLetterTargetArn":"arn:aws:sqs:us-east-1:000000000000:gorax-executions-dlq","maxReceiveCount":"3"}'

# Set environment variables
export AWS_ENDPOINT=http://localhost:4566
export AWS_REGION=us-east-1
export AWS_ACCESS_KEY_ID=test
export AWS_SECRET_ACCESS_KEY=test
export AWS_SQS_QUEUE_URL=http://localhost:4566/000000000000/gorax-executions
export AWS_SQS_DLQ_URL=http://localhost:4566/000000000000/gorax-executions-dlq
export QUEUE_ENABLED=true
```

## Production Deployment

### AWS SQS Queue Setup

1. **Create Main Queue**:
   - Type: Standard Queue
   - Visibility Timeout: 30 seconds (should match processing time)
   - Message Retention: 4 days
   - Delivery Delay: 0 seconds
   - Maximum Message Size: 256 KB

2. **Create Dead-Letter Queue**:
   - Type: Standard Queue
   - Same settings as main queue

3. **Configure Redrive Policy**:
   - Set maxReceiveCount to 3 (or desired retry limit)
   - Point to DLQ ARN

4. **Set Up IAM Permissions**:
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
           "sqs:ChangeMessageVisibility",
           "sqs:GetQueueAttributes"
         ],
         "Resource": [
           "arn:aws:sqs:*:*:gorax-executions",
           "arn:aws:sqs:*:*:gorax-executions-dlq"
         ]
       }
     ]
   }
   ```

### Monitoring

- **CloudWatch Metrics**: Monitor queue depth, age of oldest message, and DLQ messages
- **Alarms**: Set up alarms for high queue depth and DLQ messages
- **Application Metrics**: Use the built-in metrics collector for application-level monitoring

### Scaling

- **Horizontal Scaling**: Scale worker pods based on queue depth
- **Auto-Scaling**: Configure HPA (Horizontal Pod Autoscaler) to scale based on custom metrics
- **Visibility Timeout**: Adjust based on average execution time (should be 2-3x average time)

## Testing

Run tests:
```bash
go test ./internal/queue/...
```

Test with coverage:
```bash
go test -cover ./internal/queue/...
```

## Error Handling

The queue system handles errors as follows:

1. **Transient Errors**: Message becomes visible again after visibility timeout
2. **Retry Logic**: SQS automatically retries up to `maxReceiveCount` times
3. **Dead-Letter Queue**: Messages that exceed retry limit are sent to DLQ
4. **Processing Errors**: Worker logs errors and returns them to trigger retry

## Best Practices

1. **Visibility Timeout**: Set to 2-3x the average execution time
2. **Batch Processing**: Use batch operations when publishing multiple messages
3. **Long Polling**: Use 20 second wait time to reduce empty receives
4. **Monitoring**: Monitor queue depth and DLQ messages regularly
5. **Idempotency**: Ensure execution processing is idempotent (handle duplicate messages)
6. **Error Classification**: Distinguish between retryable and fatal errors

## Troubleshooting

### Messages Not Processing
- Check that `QUEUE_ENABLED=true`
- Verify SQS queue URL is correct
- Check AWS credentials and permissions
- Review worker logs for errors

### High Queue Depth
- Scale worker pods horizontally
- Increase concurrent workers per pod
- Check for slow executions blocking workers
- Verify tenant concurrency limits aren't too restrictive

### Messages in DLQ
- Review DLQ messages to identify patterns
- Check application logs for processing errors
- Verify message format is correct
- Consider increasing retry limit if needed

### Performance Issues
- Increase batch size for publishing
- Adjust visibility timeout based on execution time
- Use long polling (20 seconds)
- Scale workers based on queue depth
