package messaging

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
)

// SQSClient defines the interface for SQS operations
type SQSClient interface {
	SendMessage(input *sqs.SendMessageInput) (*sqs.SendMessageOutput, error)
	ReceiveMessage(input *sqs.ReceiveMessageInput) (*sqs.ReceiveMessageOutput, error)
	DeleteMessage(input *sqs.DeleteMessageInput) (*sqs.DeleteMessageOutput, error)
	ChangeMessageVisibility(input *sqs.ChangeMessageVisibilityInput) (*sqs.ChangeMessageVisibilityOutput, error)
	GetQueueAttributes(input *sqs.GetQueueAttributesInput) (*sqs.GetQueueAttributesOutput, error)
}

// SQSQueue implements MessageQueue for AWS SQS
type SQSQueue struct {
	client   SQSClient
	region   string
	queueURL string
}

// NewSQSQueue creates a new AWS SQS queue client
func NewSQSQueue(ctx context.Context, config Config) (*SQSQueue, error) {
	if config.Region == "" {
		return nil, fmt.Errorf("region is required for SQS")
	}

	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(config.Region),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create AWS session: %w", err)
	}

	return &SQSQueue{
		client: sqs.New(sess),
		region: config.Region,
	}, nil
}

// Send sends a message to an SQS queue
func (q *SQSQueue) Send(ctx context.Context, destination string, message []byte, attributes map[string]string) error {
	if destination == "" {
		return fmt.Errorf("destination queue URL is required")
	}

	if len(message) == 0 {
		return fmt.Errorf("message body cannot be empty")
	}

	// Convert attributes to SQS message attributes
	msgAttributes := make(map[string]*sqs.MessageAttributeValue)
	for key, value := range attributes {
		msgAttributes[key] = &sqs.MessageAttributeValue{
			DataType:    aws.String("String"),
			StringValue: aws.String(value),
		}
	}

	input := &sqs.SendMessageInput{
		QueueUrl:          aws.String(destination),
		MessageBody:       aws.String(string(message)),
		MessageAttributes: msgAttributes,
	}

	_, err := q.client.SendMessage(input)
	if err != nil {
		return fmt.Errorf("failed to send message to SQS: %w", err)
	}

	return nil
}

// Receive receives messages from an SQS queue
func (q *SQSQueue) Receive(ctx context.Context, source string, maxMessages int, waitTime time.Duration) ([]Message, error) {
	if source == "" {
		return nil, fmt.Errorf("source queue URL is required")
	}

	if maxMessages <= 0 {
		return nil, fmt.Errorf("maxMessages must be greater than 0")
	}

	// SQS has a maximum of 10 messages per request
	if maxMessages > 10 {
		maxMessages = 10
	}

	waitSeconds := int64(waitTime.Seconds())
	if waitSeconds > 20 {
		waitSeconds = 20 // SQS maximum
	}

	input := &sqs.ReceiveMessageInput{
		QueueUrl:              aws.String(source),
		MaxNumberOfMessages:   aws.Int64(int64(maxMessages)),
		WaitTimeSeconds:       aws.Int64(waitSeconds),
		MessageAttributeNames: []*string{aws.String("All")},
	}

	output, err := q.client.ReceiveMessage(input)
	if err != nil {
		return nil, fmt.Errorf("failed to receive messages from SQS: %w", err)
	}

	messages := make([]Message, 0, len(output.Messages))
	for _, sqsMsg := range output.Messages {
		// Convert SQS message attributes to map
		attributes := make(map[string]string)
		for key, attr := range sqsMsg.MessageAttributes {
			if attr.StringValue != nil {
				attributes[key] = *attr.StringValue
			}
		}

		msg := Message{
			ID:         aws.StringValue(sqsMsg.MessageId),
			Body:       []byte(aws.StringValue(sqsMsg.Body)),
			Attributes: attributes,
			Receipt:    aws.StringValue(sqsMsg.ReceiptHandle),
			Timestamp:  time.Now(), // SQS doesn't provide original timestamp easily
		}
		messages = append(messages, msg)
	}

	return messages, nil
}

// Ack acknowledges a message by deleting it from the queue
func (q *SQSQueue) Ack(ctx context.Context, message Message) error {
	if message.Receipt == "" {
		return fmt.Errorf("message receipt handle is required")
	}

	if q.queueURL == "" {
		return fmt.Errorf("queue URL not set")
	}

	input := &sqs.DeleteMessageInput{
		QueueUrl:      aws.String(q.queueURL),
		ReceiptHandle: aws.String(message.Receipt),
	}

	_, err := q.client.DeleteMessage(input)
	if err != nil {
		return fmt.Errorf("failed to delete message from SQS: %w", err)
	}

	return nil
}

// Nack negatively acknowledges a message by making it immediately visible
func (q *SQSQueue) Nack(ctx context.Context, message Message) error {
	if message.Receipt == "" {
		return fmt.Errorf("message receipt handle is required")
	}

	if q.queueURL == "" {
		return fmt.Errorf("queue URL not set")
	}

	input := &sqs.ChangeMessageVisibilityInput{
		QueueUrl:          aws.String(q.queueURL),
		ReceiptHandle:     aws.String(message.Receipt),
		VisibilityTimeout: aws.Int64(0), // Make immediately visible
	}

	_, err := q.client.ChangeMessageVisibility(input)
	if err != nil {
		return fmt.Errorf("failed to change message visibility in SQS: %w", err)
	}

	return nil
}

// GetInfo retrieves information about an SQS queue
func (q *SQSQueue) GetInfo(ctx context.Context, name string) (*QueueInfo, error) {
	if name == "" {
		return nil, fmt.Errorf("queue name/URL is required")
	}

	input := &sqs.GetQueueAttributesInput{
		QueueUrl: aws.String(name),
		AttributeNames: []*string{
			aws.String("ApproximateNumberOfMessages"),
			aws.String("CreatedTimestamp"),
		},
	}

	output, err := q.client.GetQueueAttributes(input)
	if err != nil {
		return nil, fmt.Errorf("failed to get queue attributes: %w", err)
	}

	info := &QueueInfo{
		Name: name,
	}

	if countStr, ok := output.Attributes["ApproximateNumberOfMessages"]; ok && countStr != nil {
		count, _ := strconv.Atoi(*countStr)
		info.ApproximateCount = count
	}

	if timestampStr, ok := output.Attributes["CreatedTimestamp"]; ok && timestampStr != nil {
		timestamp, _ := strconv.ParseInt(*timestampStr, 10, 64)
		info.CreatedAt = time.Unix(timestamp, 0)
	}

	return info, nil
}

// Close closes the SQS client connection
func (q *SQSQueue) Close() error {
	// AWS SDK doesn't require explicit connection closing
	return nil
}

// SetQueueURL sets the queue URL for ack/nack operations
func (q *SQSQueue) SetQueueURL(url string) {
	q.queueURL = url
}
