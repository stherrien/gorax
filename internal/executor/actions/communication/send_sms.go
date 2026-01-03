package communication

import (
	"context"
	"fmt"

	"github.com/gorax/gorax/internal/communication"
	"github.com/gorax/gorax/internal/communication/sms"
	"github.com/gorax/gorax/internal/credential"
)

// SendSMSAction executes SMS sending operations.
type SendSMSAction struct {
	config            SendSMSConfig
	credentialService credential.Service
}

// SendSMSConfig represents the configuration for the SendSMS action.
type SendSMSConfig struct {
	Provider     string `json:"provider"` // twilio, aws_sns, messagebird
	From         string `json:"from"`
	To           string `json:"to"`
	Message      string `json:"message"`
	CredentialID string `json:"credential_id"`
}

// NewSendSMSAction creates a new SendSMS action.
func NewSendSMSAction(config SendSMSConfig, credService credential.Service) *SendSMSAction {
	return &SendSMSAction{
		config:            config,
		credentialService: credService,
	}
}

// Execute sends an SMS using the configured provider.
// TODO: This action needs integration with the workflow executor context to properly
// retrieve credentials with tenant/user information. The full implementation is ready
// but commented out until the execution context integration is complete.
func (a *SendSMSAction) Execute(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
	return nil, fmt.Errorf("SMS sending action requires execution context integration (see TODO in source)")

	/*
		// Full implementation (to be uncommented after execution context integration):

		// Get tenant and user from execution context
		tenantID := ctx.Value("tenantID").(string)
		userID := ctx.Value("userID").(string)

		// Get credential
		cred, err := a.credentialService.GetValue(ctx, tenantID, a.config.CredentialID, userID)
		if err != nil {
			return nil, fmt.Errorf("failed to get credential: %w", err)
		}

		// Create provider
		provider, err := a.createProvider(cred.Value)
		if err != nil {
			return nil, fmt.Errorf("failed to create SMS provider: %w", err)
		}

		// Build SMS request
		request := &communication.SMSRequest{
			From:    a.config.From,
			To:      a.config.To,
			Message: a.config.Message,
		}

		// Send SMS
		response, err := provider.SendSMS(ctx, request)
		if err != nil {
			return map[string]interface{}{
				"success":    false,
				"error":      err.Error(),
				"message_id": "",
			}, fmt.Errorf("failed to send SMS: %w", err)
		}

		return map[string]interface{}{
			"success":    true,
			"message_id": response.MessageID,
			"status":     response.Status,
			"cost":       response.Cost,
			"sent_at":    response.SentAt,
		}, nil
	*/
}

// createProvider creates an SMS provider based on the configuration.
func (a *SendSMSAction) createProvider(credValue map[string]interface{}) (communication.SMSProvider, error) {
	switch a.config.Provider {
	case "twilio":
		accountSID, ok := credValue["account_sid"].(string)
		if !ok {
			return nil, fmt.Errorf("twilio account_sid not found in credential")
		}
		authToken, ok := credValue["auth_token"].(string)
		if !ok {
			return nil, fmt.Errorf("twilio auth_token not found in credential")
		}
		return sms.NewTwilioProvider(accountSID, authToken), nil

	case "aws_sns":
		region, ok := credValue["region"].(string)
		if !ok {
			region = "us-east-1" // Default region
		}
		return sms.NewSNSProvider(region)

	case "messagebird":
		apiKey, ok := credValue["api_key"].(string)
		if !ok {
			return nil, fmt.Errorf("messagebird api_key not found in credential")
		}
		return sms.NewMessageBirdProvider(apiKey), nil

	default:
		return nil, fmt.Errorf("unsupported SMS provider: %s", a.config.Provider)
	}
}

// Name returns the action name.
func (a *SendSMSAction) Name() string {
	return "send_sms"
}

// Validate validates the action configuration.
func (a *SendSMSAction) Validate() error {
	if a.config.Provider == "" {
		return fmt.Errorf("provider is required")
	}
	if a.config.From == "" {
		return fmt.Errorf("from number is required")
	}
	if a.config.To == "" {
		return fmt.Errorf("to number is required")
	}
	if a.config.Message == "" {
		return fmt.Errorf("message is required")
	}
	if a.config.CredentialID == "" {
		return fmt.Errorf("credential_id is required")
	}
	return nil
}
