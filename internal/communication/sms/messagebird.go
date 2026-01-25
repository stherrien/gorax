package sms

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/gorax/gorax/internal/communication"
)

// MessageBirdProvider implements SMSProvider using MessageBird API.
type MessageBirdProvider struct {
	apiKey string
	client *http.Client
}

// NewMessageBirdProvider creates a new MessageBird SMS provider.
func NewMessageBirdProvider(apiKey string) *MessageBirdProvider {
	return &MessageBirdProvider{
		apiKey: apiKey,
		client: &http.Client{Timeout: 30 * time.Second},
	}
}

type messageBirdRequest struct {
	Recipients []string `json:"recipients"`
	Originator string   `json:"originator"`
	Body       string   `json:"body"`
}

type messageBirdResponse struct {
	ID         string `json:"id"`
	Recipients struct {
		TotalCount          int `json:"totalCount"`
		TotalSentCount      int `json:"totalSentCount"`
		TotalDeliveredCount int `json:"totalDeliveredCount"`
	} `json:"recipients"`
}

// SendSMS sends a single SMS using MessageBird.
func (p *MessageBirdProvider) SendSMS(ctx context.Context, request *communication.SMSRequest) (*communication.SMSResponse, error) {
	if err := request.Validate(); err != nil {
		return nil, fmt.Errorf("invalid SMS request: %w", err)
	}

	// Create request payload
	payload := messageBirdRequest{
		Recipients: []string{request.To},
		Originator: request.From,
		Body:       request.Message,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, "POST", "https://rest.messagebird.com/messages", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Authorization", fmt.Sprintf("AccessKey %s", p.apiKey))
	httpReq.Header.Set("Content-Type", "application/json")

	// Send request
	resp, err := p.client.Do(httpReq)
	if err != nil {
		return &communication.SMSResponse{
			Status: string(communication.MessageStatusFailed),
			Error:  err,
			SentAt: time.Now(),
		}, fmt.Errorf("failed to send SMS: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return &communication.SMSResponse{
			Status: string(communication.MessageStatusFailed),
			Error:  fmt.Errorf("messagebird API error: %d - %s", resp.StatusCode, string(body)),
			SentAt: time.Now(),
		}, fmt.Errorf("messagebird API error: status=%d", resp.StatusCode)
	}

	var mbResp messageBirdResponse
	if err := json.Unmarshal(body, &mbResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	status := communication.MessageStatusSent
	if mbResp.Recipients.TotalSentCount == 0 {
		status = communication.MessageStatusFailed
	} else if mbResp.Recipients.TotalSentCount < mbResp.Recipients.TotalCount {
		status = communication.MessageStatusQueued
	}

	return &communication.SMSResponse{
		MessageID: mbResp.ID,
		Status:    string(status),
		Cost:      0.01, // Approximate cost per message
		SentAt:    time.Now(),
	}, nil
}

// SendBulkSMS sends multiple SMS messages using MessageBird.
func (p *MessageBirdProvider) SendBulkSMS(ctx context.Context, requests []*communication.SMSRequest) ([]*communication.SMSResponse, error) {
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
