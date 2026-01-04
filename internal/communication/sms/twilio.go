package sms

import (
	"context"
	"fmt"
	"time"

	"github.com/gorax/gorax/internal/communication"
	"github.com/twilio/twilio-go"
	twilioApi "github.com/twilio/twilio-go/rest/api/v2010"
)

// TwilioProvider implements SMSProvider using Twilio API.
type TwilioProvider struct {
	client *twilio.RestClient
}

// NewTwilioProvider creates a new Twilio SMS provider.
func NewTwilioProvider(accountSID, authToken string) *TwilioProvider {
	client := twilio.NewRestClientWithParams(twilio.ClientParams{
		Username: accountSID,
		Password: authToken,
	})

	return &TwilioProvider{
		client: client,
	}
}

// SendSMS sends a single SMS using Twilio.
func (p *TwilioProvider) SendSMS(ctx context.Context, request *communication.SMSRequest) (*communication.SMSResponse, error) {
	if err := request.Validate(); err != nil {
		return nil, fmt.Errorf("invalid SMS request: %w", err)
	}

	params := &twilioApi.CreateMessageParams{}
	params.SetFrom(request.From)
	params.SetTo(request.To)
	params.SetBody(request.Message)

	resp, err := p.client.Api.CreateMessage(params)
	if err != nil {
		return &communication.SMSResponse{
			Status: string(communication.MessageStatusFailed),
			Error:  err,
			SentAt: time.Now(),
		}, fmt.Errorf("failed to send SMS: %w", err)
	}

	// Parse cost if available
	var cost float64
	if resp.Price != nil {
		if _, err := fmt.Sscanf(*resp.Price, "%f", &cost); err != nil {
			cost = 0.0
		}
		// Make cost positive
		if cost < 0 {
			cost = -cost
		}
	}

	status := communication.MessageStatusSent
	if resp.Status != nil {
		switch *resp.Status {
		case "queued", "sending":
			status = communication.MessageStatusQueued
		case "sent", "delivered":
			status = communication.MessageStatusSent
		case "failed", "undelivered":
			status = communication.MessageStatusFailed
		}
	}

	return &communication.SMSResponse{
		MessageID: *resp.Sid,
		Status:    string(status),
		Cost:      cost,
		SentAt:    time.Now(),
	}, nil
}

// SendBulkSMS sends multiple SMS messages using Twilio.
func (p *TwilioProvider) SendBulkSMS(ctx context.Context, requests []*communication.SMSRequest) ([]*communication.SMSResponse, error) {
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
