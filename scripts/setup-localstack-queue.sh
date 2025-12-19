#!/bin/bash
set -e

# Setup LocalStack SQS queues for local development
# This script creates the necessary SQS queues for the rflow queue system

ENDPOINT_URL="${AWS_ENDPOINT:-http://localhost:4566}"
REGION="${AWS_REGION:-us-east-1}"
MAIN_QUEUE_NAME="${MAIN_QUEUE_NAME:-rflow-executions}"
DLQ_NAME="${DLQ_NAME:-rflow-executions-dlq}"
MAX_RECEIVE_COUNT="${MAX_RECEIVE_COUNT:-3}"

echo "Setting up LocalStack SQS queues..."
echo "Endpoint: $ENDPOINT_URL"
echo "Region: $REGION"
echo "Main Queue: $MAIN_QUEUE_NAME"
echo "DLQ: $DLQ_NAME"
echo "Max Receive Count: $MAX_RECEIVE_COUNT"
echo ""

# Wait for LocalStack to be ready
echo "Waiting for LocalStack to be ready..."
max_attempts=30
attempt=0
until curl -s "$ENDPOINT_URL" > /dev/null 2>&1 || [ $attempt -eq $max_attempts ]; do
  echo "  Attempt $((attempt+1))/$max_attempts..."
  sleep 1
  attempt=$((attempt+1))
done

if [ $attempt -eq $max_attempts ]; then
  echo "ERROR: LocalStack is not responding at $ENDPOINT_URL"
  echo "Please make sure LocalStack is running:"
  echo "  docker run -d --name localstack -p 4566:4566 -e SERVICES=sqs localstack/localstack"
  exit 1
fi

echo "LocalStack is ready!"
echo ""

# Create dead-letter queue first
echo "Creating dead-letter queue: $DLQ_NAME"
DLQ_URL=$(aws --endpoint-url="$ENDPOINT_URL" --region="$REGION" sqs create-queue \
  --queue-name "$DLQ_NAME" \
  --output text \
  --query 'QueueUrl' 2>&1)

if [ $? -eq 0 ]; then
  echo "  DLQ created: $DLQ_URL"
else
  echo "  DLQ may already exist, attempting to get URL..."
  DLQ_URL=$(aws --endpoint-url="$ENDPOINT_URL" --region="$REGION" sqs get-queue-url \
    --queue-name "$DLQ_NAME" \
    --output text \
    --query 'QueueUrl')
  echo "  DLQ URL: $DLQ_URL"
fi

# Get DLQ ARN
echo "Getting DLQ ARN..."
DLQ_ARN=$(aws --endpoint-url="$ENDPOINT_URL" --region="$REGION" sqs get-queue-attributes \
  --queue-url "$DLQ_URL" \
  --attribute-names QueueArn \
  --output text \
  --query 'Attributes.QueueArn')
echo "  DLQ ARN: $DLQ_ARN"
echo ""

# Create main queue
echo "Creating main queue: $MAIN_QUEUE_NAME"
MAIN_QUEUE_URL=$(aws --endpoint-url="$ENDPOINT_URL" --region="$REGION" sqs create-queue \
  --queue-name "$MAIN_QUEUE_NAME" \
  --attributes VisibilityTimeout=30 \
  --output text \
  --query 'QueueUrl' 2>&1)

if [ $? -eq 0 ]; then
  echo "  Main queue created: $MAIN_QUEUE_URL"
else
  echo "  Main queue may already exist, attempting to get URL..."
  MAIN_QUEUE_URL=$(aws --endpoint-url="$ENDPOINT_URL" --region="$REGION" sqs get-queue-url \
    --queue-name "$MAIN_QUEUE_NAME" \
    --output text \
    --query 'QueueUrl')
  echo "  Main queue URL: $MAIN_QUEUE_URL"
fi
echo ""

# Configure redrive policy (DLQ association)
echo "Configuring redrive policy..."
REDRIVE_POLICY="{\"deadLetterTargetArn\":\"$DLQ_ARN\",\"maxReceiveCount\":\"$MAX_RECEIVE_COUNT\"}"
aws --endpoint-url="$ENDPOINT_URL" --region="$REGION" sqs set-queue-attributes \
  --queue-url "$MAIN_QUEUE_URL" \
  --attributes "RedrivePolicy=$REDRIVE_POLICY"
echo "  Redrive policy configured (max receive count: $MAX_RECEIVE_COUNT)"
echo ""

# Display queue information
echo "Queue setup complete!"
echo ""
echo "=========================================="
echo "Queue Configuration"
echo "=========================================="
echo "Main Queue URL:"
echo "  $MAIN_QUEUE_URL"
echo ""
echo "Dead-Letter Queue URL:"
echo "  $DLQ_URL"
echo ""
echo "=========================================="
echo "Environment Variables"
echo "=========================================="
echo "Add these to your .env file:"
echo ""
echo "AWS_ENDPOINT=$ENDPOINT_URL"
echo "AWS_REGION=$REGION"
echo "AWS_ACCESS_KEY_ID=test"
echo "AWS_SECRET_ACCESS_KEY=test"
echo "AWS_SQS_QUEUE_URL=$MAIN_QUEUE_URL"
echo "AWS_SQS_DLQ_URL=$DLQ_URL"
echo "QUEUE_ENABLED=true"
echo ""
echo "=========================================="
echo "Verification"
echo "=========================================="

# List queues
echo "Available queues:"
aws --endpoint-url="$ENDPOINT_URL" --region="$REGION" sqs list-queues

# Get queue attributes
echo ""
echo "Main queue attributes:"
aws --endpoint-url="$ENDPOINT_URL" --region="$REGION" sqs get-queue-attributes \
  --queue-url "$MAIN_QUEUE_URL" \
  --attribute-names All \
  --output table

echo ""
echo "Setup complete! You can now start the rflow worker with queue support enabled."
echo ""
echo "To start LocalStack (if not already running):"
echo "  docker run -d --name localstack -p 4566:4566 -e SERVICES=sqs localstack/localstack"
echo ""
echo "To test sending a message:"
echo "  aws --endpoint-url=$ENDPOINT_URL sqs send-message \\"
echo "    --queue-url $MAIN_QUEUE_URL \\"
echo "    --message-body '{\"execution_id\":\"test-123\",\"tenant_id\":\"tenant-1\",\"workflow_id\":\"wf-1\",\"workflow_version\":1,\"trigger_type\":\"manual\"}'"
echo ""
echo "To view messages:"
echo "  aws --endpoint-url=$ENDPOINT_URL sqs receive-message --queue-url $MAIN_QUEUE_URL"
echo ""
echo "To purge queues (clear all messages):"
echo "  aws --endpoint-url=$ENDPOINT_URL sqs purge-queue --queue-url $MAIN_QUEUE_URL"
echo "  aws --endpoint-url=$ENDPOINT_URL sqs purge-queue --queue-url $DLQ_URL"
