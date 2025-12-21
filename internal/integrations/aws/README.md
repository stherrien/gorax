# AWS Integrations for Gorax

This package provides AWS service integrations for the Gorax workflow automation platform. All integrations use AWS SDK v2 and support credential injection from the credential vault.

## Supported Services

### S3 (Simple Storage Service)

#### `aws:s3:list_buckets`
Lists all S3 buckets in the AWS account.

**Configuration:** None required

**Output:**
```json
{
  "buckets": [
    {
      "name": "bucket-name",
      "creation_date": "2024-01-01T00:00:00Z"
    }
  ],
  "count": 1
}
```

#### `aws:s3:get_object`
Retrieves an object from an S3 bucket.

**Configuration:**
```json
{
  "bucket": "my-bucket",
  "key": "path/to/file.txt"
}
```

**Output:**
```json
{
  "body": "file contents",
  "content_type": "text/plain",
  "content_length": 1234,
  "last_modified": "2024-01-01T00:00:00Z",
  "etag": "\"abc123\""
}
```

#### `aws:s3:put_object`
Uploads an object to an S3 bucket.

**Configuration:**
```json
{
  "bucket": "my-bucket",
  "key": "path/to/file.txt",
  "body": "file contents",
  "content_type": "text/plain"
}
```

**Output:**
```json
{
  "etag": "\"abc123\"",
  "success": true
}
```

#### `aws:s3:delete_object`
Deletes an object from an S3 bucket.

**Configuration:**
```json
{
  "bucket": "my-bucket",
  "key": "path/to/file.txt"
}
```

**Output:**
```json
{
  "success": true,
  "bucket": "my-bucket",
  "key": "path/to/file.txt"
}
```

---

### SNS (Simple Notification Service)

#### `aws:sns:publish`
Publishes a message to an SNS topic or endpoint.

**Configuration:**
```json
{
  "topic_arn": "arn:aws:sns:us-east-1:123456789012:my-topic",
  "message": "Hello, World!",
  "subject": "Test Message",
  "attributes": {
    "key1": "value1",
    "key2": "value2"
  }
}
```

**Alternative with target ARN:**
```json
{
  "target_arn": "arn:aws:sns:us-east-1:123456789012:endpoint/test",
  "message": "Direct message"
}
```

**Output:**
```json
{
  "message_id": "12345678-1234-1234-1234-123456789012",
  "success": true
}
```

---

### SQS (Simple Queue Service)

#### `aws:sqs:send_message`
Sends a message to an SQS queue (supports FIFO queues).

**Configuration:**
```json
{
  "queue_url": "https://sqs.us-east-1.amazonaws.com/123456789012/my-queue",
  "message_body": "Hello, World!",
  "delay_seconds": 0,
  "message_group_id": "group1",
  "attributes": {
    "key1": "value1"
  }
}
```

**Output:**
```json
{
  "message_id": "12345678-1234-1234-1234-123456789012",
  "success": true
}
```

#### `aws:sqs:receive_message`
Receives messages from an SQS queue.

**Configuration:**
```json
{
  "queue_url": "https://sqs.us-east-1.amazonaws.com/123456789012/my-queue",
  "max_messages": 10,
  "wait_time_seconds": 20,
  "visibility_timeout": 30
}
```

**Output:**
```json
{
  "messages": [
    {
      "message_id": "12345678-1234-1234-1234-123456789012",
      "body": "message content",
      "receipt_handle": "receipt-handle-string",
      "attributes": {
        "key1": "value1"
      }
    }
  ],
  "count": 1
}
```

#### `aws:sqs:delete_message`
Deletes a message from an SQS queue.

**Configuration:**
```json
{
  "queue_url": "https://sqs.us-east-1.amazonaws.com/123456789012/my-queue",
  "receipt_handle": "receipt-handle-from-receive"
}
```

**Output:**
```json
{
  "success": true
}
```

---

### Lambda

#### `aws:lambda:invoke`
Invokes an AWS Lambda function synchronously or asynchronously.

**Configuration (Synchronous):**
```json
{
  "function_name": "my-function",
  "payload": {
    "key": "value",
    "another_key": 123
  },
  "invocation_type": "RequestResponse",
  "qualifier": "v1"
}
```

**Configuration (Asynchronous):**
```json
{
  "function_name": "my-function",
  "payload": {
    "key": "value"
  },
  "invocation_type": "Event"
}
```

**Output (Sync):**
```json
{
  "status_code": 200,
  "success": true,
  "payload": {
    "result": "function response"
  },
  "executed_version": "$LATEST"
}
```

**Output (Async):**
```json
{
  "status_code": 202,
  "success": true,
  "invocation_type": "async"
}
```

**Invocation Types:**
- `RequestResponse` (default): Synchronous invocation with response payload
- `Event`: Asynchronous invocation (fire and forget)
- `DryRun`: Validate parameters without executing

---

## Usage Example

### Registering AWS Actions

```go
import (
    "github.com/gorax/gorax/internal/integrations"
    "github.com/gorax/gorax/internal/integrations/aws"
)

func main() {
    registry := integrations.NewRegistry()

    // Register all AWS actions with credentials from vault
    err := aws.RegisterAWSActions(
        registry,
        "aws-access-key",
        "aws-secret-key",
        "us-east-1",
    )
    if err != nil {
        log.Fatal(err)
    }
}
```

### Using in Workflows

In your workflow definition:

```json
{
  "id": "step-1",
  "type": "action",
  "action_type": "aws:s3:get_object",
  "config": {
    "bucket": "{{credentials.aws_bucket}}",
    "key": "data/input.json"
  }
}
```

## Credential Management

All AWS integrations support credential injection from the Gorax credential vault. Store your AWS credentials securely:

1. Create AWS credentials in the vault:
   - `aws_access_key_id`
   - `aws_secret_access_key`
   - `aws_region`

2. Reference credentials in workflow configurations:
   ```json
   {
     "config": {
       "bucket": "{{credentials.aws_bucket}}"
     }
   }
   ```

## Testing

Run all AWS integration tests:

```bash
go test ./internal/integrations/aws/... -v
```

Run with coverage:

```bash
go test ./internal/integrations/aws/... -cover
```

## Architecture

- **Client Layer**: Wraps AWS SDK v2 clients (S3Client, SNSClient, SQSClient, LambdaClient)
- **Action Layer**: Implements the `integrations.Action` interface for each operation
- **Validation**: Comprehensive configuration validation before execution
- **Error Handling**: AWS SDK errors are wrapped with context
- **Testing**: TDD approach with unit tests for all actions

## Dependencies

- `github.com/aws/aws-sdk-go-v2` - AWS SDK for Go v2
- `github.com/aws/aws-sdk-go-v2/service/s3`
- `github.com/aws/aws-sdk-go-v2/service/sns`
- `github.com/aws/aws-sdk-go-v2/service/sqs`
- `github.com/aws/aws-sdk-go-v2/service/lambda`

## Contributing

When adding new AWS service integrations:

1. Follow TDD - write tests first
2. Implement the `integrations.Action` interface
3. Add comprehensive validation
4. Update `RegisterAWSActions()` and `GetAWSActionsList()`
5. Add documentation to this README
6. Ensure test coverage > 50%
