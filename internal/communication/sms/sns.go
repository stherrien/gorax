package sms

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/gorax/gorax/internal/communication"
)

// SNSProvider implements SMSProvider using AWS SNS.
type SNSProvider struct {
	client *sns.SNS
}

// NewSNSProvider creates a new AWS SNS SMS provider.
func NewSNSProvider(region string) (*SNSProvider, error) {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(region),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create AWS session: %w", err)
	}

	return &SNSProvider{
		client: sns.New(sess),
	}, nil
}

// SendSMS sends a single SMS using AWS SNS.
func (p *SNSProvider) SendSMS(ctx context.Context, request *communication.SMSRequest) (*communication.SMSResponse, error) {
	if err := request.Validate(); err != nil {
		return nil, fmt.Errorf("invalid SMS request: %w", err)
	}

	input := &sns.PublishInput{
		Message:     aws.String(request.Message),
		PhoneNumber: aws.String(request.To),
		MessageAttributes: map[string]*sns.MessageAttributeValue{
			"AWS.SNS.SMS.SenderID": {
				DataType:    aws.String("String"),
				StringValue: aws.String(request.From),
			},
			"AWS.SNS.SMS.SMSType": {
				DataType:    aws.String("String"),
				StringValue: aws.String("Transactional"),
			},
		},
	}

	result, err := p.client.PublishWithContext(ctx, input)
	if err != nil {
		return &communication.SMSResponse{
			Status: string(communication.MessageStatusFailed),
			Error:  err,
			SentAt: time.Now(),
		}, fmt.Errorf("failed to send SMS: %w", err)
	}

	return &communication.SMSResponse{
		MessageID: aws.StringValue(result.MessageId),
		Status:    string(communication.MessageStatusSent),
		SentAt:    time.Now(),
	}, nil
}

// SendBulkSMS sends multiple SMS messages using AWS SNS.
func (p *SNSProvider) SendBulkSMS(ctx context.Context, requests []*communication.SMSRequest) ([]*communication.SMSResponse, error) {
	responses := make([]*communication.SMSResponse, len(requests))
	var firstErr error

	for i, req := range requests {
		resp, err := p.SendSMS(ctx, req)
		responses[i] = resp
		if err != nil && firstErr == nil {
			firstErr = err
		}
	}

	return responses, firstErr
}
