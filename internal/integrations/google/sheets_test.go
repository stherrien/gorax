package google

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gorax/gorax/internal/credential"
	"github.com/gorax/gorax/internal/executor/actions"
)

func TestSheetsReadAction_Execute(t *testing.T) {
	tests := []struct {
		name           string
		config         SheetsReadConfig
		context        map[string]interface{}
		mockCredential *credential.DecryptedValue
		sheetsResponse interface{}
		sheetsStatus   int
		wantErr        bool
		errorContains  string
		validate       func(t *testing.T, output *actions.ActionOutput)
	}{
		{
			name: "successful range read",
			config: SheetsReadConfig{
				SpreadsheetID: "test-spreadsheet-id",
				Range:         "Sheet1!A1:B2",
			},
			context: map[string]interface{}{
				"env": map[string]interface{}{
					"tenant_id": "tenant-123",
				},
				"credential_id": "cred-123",
			},
			mockCredential: &credential.DecryptedValue{
				Value: map[string]interface{}{
					"access_token": "test-access-token",
				},
			},
			sheetsResponse: map[string]interface{}{
				"range": "Sheet1!A1:B2",
				"values": []interface{}{
					[]interface{}{"Header1", "Header2"},
					[]interface{}{"Value1", "Value2"},
				},
			},
			sheetsStatus: http.StatusOK,
			wantErr:      false,
			validate: func(t *testing.T, output *actions.ActionOutput) {
				assert.NotNil(t, output)
				result, ok := output.Data.(*SheetsReadResult)
				assert.True(t, ok)
				assert.Len(t, result.Values, 2)
				assert.Equal(t, "Sheet1!A1:B2", result.Range)
			},
		},
		{
			name: "missing spreadsheet_id",
			config: SheetsReadConfig{
				Range: "Sheet1!A1:B2",
			},
			context: map[string]interface{}{
				"env": map[string]interface{}{
					"tenant_id": "tenant-123",
				},
				"credential_id": "cred-123",
			},
			wantErr:       true,
			errorContains: "spreadsheet_id is required",
		},
		{
			name: "missing range",
			config: SheetsReadConfig{
				SpreadsheetID: "test-spreadsheet-id",
			},
			context: map[string]interface{}{
				"env": map[string]interface{}{
					"tenant_id": "tenant-123",
				},
				"credential_id": "cred-123",
			},
			wantErr:       true,
			errorContains: "range is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var server *httptest.Server
			if tt.sheetsResponse != nil {
				server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(tt.sheetsStatus)
					json.NewEncoder(w).Encode(tt.sheetsResponse)
				}))
				defer server.Close()
			}

			mockCred := &MockCredentialService{
				GetValueFunc: func(ctx context.Context, tenantID, credentialID, userID string) (*credential.DecryptedValue, error) {
					return tt.mockCredential, nil
				},
			}

			action := NewSheetsReadAction(mockCred)
			if server != nil {
				action.baseURL = server.URL
			}

			input := actions.NewActionInput(tt.config, tt.context)
			ctx := context.Background()
			output, err := action.Execute(ctx, input)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, err)
				if tt.validate != nil {
					tt.validate(t, output)
				}
			}
		})
	}
}

func TestSheetsWriteAction_Execute(t *testing.T) {
	tests := []struct {
		name           string
		config         SheetsWriteConfig
		context        map[string]interface{}
		mockCredential *credential.DecryptedValue
		wantErr        bool
		errorContains  string
	}{
		{
			name: "valid write config",
			config: SheetsWriteConfig{
				SpreadsheetID: "test-spreadsheet-id",
				Range:         "Sheet1!A1:B2",
				Values: [][]interface{}{
					{"Header1", "Header2"},
					{"Value1", "Value2"},
				},
			},
			context: map[string]interface{}{
				"env": map[string]interface{}{
					"tenant_id": "tenant-123",
				},
				"credential_id": "cred-123",
			},
			mockCredential: &credential.DecryptedValue{
				Value: map[string]interface{}{
					"access_token": "test-access-token",
				},
			},
			wantErr: false,
		},
		{
			name: "missing values",
			config: SheetsWriteConfig{
				SpreadsheetID: "test-spreadsheet-id",
				Range:         "Sheet1!A1:B2",
			},
			context: map[string]interface{}{
				"env": map[string]interface{}{
					"tenant_id": "tenant-123",
				},
				"credential_id": "cred-123",
			},
			wantErr:       true,
			errorContains: "values are required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockCred := &MockCredentialService{
				GetValueFunc: func(ctx context.Context, tenantID, credentialID, userID string) (*credential.DecryptedValue, error) {
					return tt.mockCredential, nil
				},
			}

			action := NewSheetsWriteAction(mockCred)

			input := actions.NewActionInput(tt.config, tt.context)
			ctx := context.Background()
			_, err := action.Execute(ctx, input)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			}
		})
	}
}

func TestSheetsReadConfig_Validate(t *testing.T) {
	tests := []struct {
		name      string
		config    SheetsReadConfig
		wantErr   bool
		errString string
	}{
		{
			name: "valid config",
			config: SheetsReadConfig{
				SpreadsheetID: "test-id",
				Range:         "Sheet1!A1:B2",
			},
			wantErr: false,
		},
		{
			name: "missing spreadsheet_id",
			config: SheetsReadConfig{
				Range: "Sheet1!A1:B2",
			},
			wantErr:   true,
			errString: "spreadsheet_id is required",
		},
		{
			name: "missing range",
			config: SheetsReadConfig{
				SpreadsheetID: "test-id",
			},
			wantErr:   true,
			errString: "range is required",
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
