package google

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gorax/gorax/internal/credential"
	"github.com/gorax/gorax/internal/executor/actions"
)

func TestCalendarCreateConfig_Validate(t *testing.T) {
	tests := []struct {
		name      string
		config    CalendarCreateConfig
		wantErr   bool
		errString string
	}{
		{
			name: "valid config",
			config: CalendarCreateConfig{
				CalendarID: "primary",
				Summary:    "Test Event",
				StartTime:  "2024-01-01T10:00:00Z",
				EndTime:    "2024-01-01T11:00:00Z",
			},
			wantErr: false,
		},
		{
			name: "missing calendar_id",
			config: CalendarCreateConfig{
				Summary:   "Test Event",
				StartTime: "2024-01-01T10:00:00Z",
				EndTime:   "2024-01-01T11:00:00Z",
			},
			wantErr:   true,
			errString: "calendar_id is required",
		},
		{
			name: "missing summary",
			config: CalendarCreateConfig{
				CalendarID: "primary",
				StartTime:  "2024-01-01T10:00:00Z",
				EndTime:    "2024-01-01T11:00:00Z",
			},
			wantErr:   true,
			errString: "summary is required",
		},
		{
			name: "missing start_time",
			config: CalendarCreateConfig{
				CalendarID: "primary",
				Summary:    "Test Event",
				EndTime:    "2024-01-01T11:00:00Z",
			},
			wantErr:   true,
			errString: "start_time is required",
		},
		{
			name: "missing end_time",
			config: CalendarCreateConfig{
				CalendarID: "primary",
				Summary:    "Test Event",
				StartTime:  "2024-01-01T10:00:00Z",
			},
			wantErr:   true,
			errString: "end_time is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errString)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestCalendarListConfig_Validate(t *testing.T) {
	tests := []struct {
		name      string
		config    CalendarListConfig
		wantErr   bool
		errString string
	}{
		{
			name: "valid config",
			config: CalendarListConfig{
				CalendarID: "primary",
			},
			wantErr: false,
		},
		{
			name:      "missing calendar_id",
			config:    CalendarListConfig{},
			wantErr:   true,
			errString: "calendar_id is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errString)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestCalendarDeleteConfig_Validate(t *testing.T) {
	tests := []struct {
		name      string
		config    CalendarDeleteConfig
		wantErr   bool
		errString string
	}{
		{
			name: "valid config",
			config: CalendarDeleteConfig{
				CalendarID: "primary",
				EventID:    "event-123",
			},
			wantErr: false,
		},
		{
			name: "missing calendar_id",
			config: CalendarDeleteConfig{
				EventID: "event-123",
			},
			wantErr:   true,
			errString: "calendar_id is required",
		},
		{
			name: "missing event_id",
			config: CalendarDeleteConfig{
				CalendarID: "primary",
			},
			wantErr:   true,
			errString: "event_id is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errString)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestCalendarCreateAction_Execute(t *testing.T) {
	mockCred := &MockCredentialService{
		GetValueFunc: func(ctx context.Context, tenantID, credentialID, userID string) (*credential.DecryptedValue, error) {
			return &credential.DecryptedValue{
				Value: map[string]interface{}{
					"access_token": "test-access-token",
				},
			}, nil
		},
	}

	action := NewCalendarCreateAction(mockCred)

	config := CalendarCreateConfig{
		CalendarID: "primary",
		Summary:    "Test Event",
		StartTime:  "2024-01-01T10:00:00Z",
		EndTime:    "2024-01-01T11:00:00Z",
	}

	input := actions.NewActionInput(config, map[string]interface{}{
		"env": map[string]interface{}{
			"tenant_id": "tenant-123",
		},
		"credential_id": "cred-123",
	})

	ctx := context.Background()
	_, err := action.Execute(ctx, input)

	// Expect error since we don't have a real Calendar API
	assert.Error(t, err)
}

func TestCalendarListAction_Execute(t *testing.T) {
	mockCred := &MockCredentialService{
		GetValueFunc: func(ctx context.Context, tenantID, credentialID, userID string) (*credential.DecryptedValue, error) {
			return &credential.DecryptedValue{
				Value: map[string]interface{}{
					"access_token": "test-access-token",
				},
			}, nil
		},
	}

	action := NewCalendarListAction(mockCred)

	config := CalendarListConfig{
		CalendarID: "primary",
		MaxResults: 10,
	}

	input := actions.NewActionInput(config, map[string]interface{}{
		"env": map[string]interface{}{
			"tenant_id": "tenant-123",
		},
		"credential_id": "cred-123",
	})

	ctx := context.Background()
	_, err := action.Execute(ctx, input)

	// Expect error since we don't have a real Calendar API
	assert.Error(t, err)
}
