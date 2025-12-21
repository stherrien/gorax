package aws

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewLambdaClient(t *testing.T) {
	tests := []struct {
		name      string
		accessKey string
		secretKey string
		region    string
		wantErr   bool
	}{
		{
			name:      "valid credentials",
			accessKey: "test-access-key",
			secretKey: "test-secret-key",
			region:    "us-east-1",
			wantErr:   false,
		},
		{
			name:      "missing access key",
			accessKey: "",
			secretKey: "test-secret-key",
			region:    "us-east-1",
			wantErr:   true,
		},
		{
			name:      "missing secret key",
			accessKey: "test-access-key",
			secretKey: "",
			region:    "us-east-1",
			wantErr:   true,
		},
		{
			name:      "missing region",
			accessKey: "test-access-key",
			secretKey: "test-secret-key",
			region:    "",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewLambdaClient(tt.accessKey, tt.secretKey, tt.region)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, client)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, client)
			}
		})
	}
}

func TestInvokeFunctionAction_Name(t *testing.T) {
	action := NewInvokeFunctionAction("key", "secret", "us-east-1")
	assert.Equal(t, "aws:lambda:invoke", action.Name())
}

func TestInvokeFunctionAction_Description(t *testing.T) {
	action := NewInvokeFunctionAction("key", "secret", "us-east-1")
	assert.NotEmpty(t, action.Description())
}

func TestInvokeFunctionAction_Validate(t *testing.T) {
	action := NewInvokeFunctionAction("key", "secret", "us-east-1")

	tests := []struct {
		name    string
		config  map[string]interface{}
		wantErr bool
	}{
		{
			name: "valid config with sync invocation",
			config: map[string]interface{}{
				"function_name":   "test-function",
				"payload":         map[string]interface{}{"key": "value"},
				"invocation_type": "RequestResponse",
			},
			wantErr: false,
		},
		{
			name: "valid config with async invocation",
			config: map[string]interface{}{
				"function_name":   "test-function",
				"payload":         map[string]interface{}{"key": "value"},
				"invocation_type": "Event",
			},
			wantErr: false,
		},
		{
			name: "valid config with qualifier",
			config: map[string]interface{}{
				"function_name": "test-function",
				"payload":       map[string]interface{}{"key": "value"},
				"qualifier":     "v1",
			},
			wantErr: false,
		},
		{
			name: "valid config without invocation type (defaults to sync)",
			config: map[string]interface{}{
				"function_name": "test-function",
				"payload":       map[string]interface{}{"key": "value"},
			},
			wantErr: false,
		},
		{
			name: "missing function name",
			config: map[string]interface{}{
				"payload": map[string]interface{}{"key": "value"},
			},
			wantErr: true,
		},
		{
			name: "missing payload",
			config: map[string]interface{}{
				"function_name": "test-function",
			},
			wantErr: true,
		},
		{
			name:    "empty config",
			config:  map[string]interface{}{},
			wantErr: true,
		},
		{
			name: "invalid invocation type",
			config: map[string]interface{}{
				"function_name":   "test-function",
				"payload":         map[string]interface{}{"key": "value"},
				"invocation_type": "InvalidType",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := action.Validate(tt.config)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestInvokeFunctionConfigValidation(t *testing.T) {
	tests := []struct {
		name    string
		config  InvokeFunctionConfig
		wantErr bool
	}{
		{
			name: "valid config",
			config: InvokeFunctionConfig{
				FunctionName:   "test-function",
				Payload:        map[string]interface{}{"key": "value"},
				InvocationType: "RequestResponse",
			},
			wantErr: false,
		},
		{
			name: "valid config with Event type",
			config: InvokeFunctionConfig{
				FunctionName:   "test-function",
				Payload:        map[string]interface{}{"key": "value"},
				InvocationType: "Event",
			},
			wantErr: false,
		},
		{
			name: "valid config with DryRun type",
			config: InvokeFunctionConfig{
				FunctionName:   "test-function",
				Payload:        map[string]interface{}{"key": "value"},
				InvocationType: "DryRun",
			},
			wantErr: false,
		},
		{
			name: "valid config without invocation type",
			config: InvokeFunctionConfig{
				FunctionName: "test-function",
				Payload:      map[string]interface{}{"key": "value"},
			},
			wantErr: false,
		},
		{
			name: "empty function name",
			config: InvokeFunctionConfig{
				FunctionName: "",
				Payload:      map[string]interface{}{"key": "value"},
			},
			wantErr: true,
		},
		{
			name: "nil payload",
			config: InvokeFunctionConfig{
				FunctionName: "test-function",
				Payload:      nil,
			},
			wantErr: true,
		},
		{
			name: "invalid invocation type",
			config: InvokeFunctionConfig{
				FunctionName:   "test-function",
				Payload:        map[string]interface{}{"key": "value"},
				InvocationType: "InvalidType",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestLambdaClientValidation(t *testing.T) {
	t.Run("validateLambdaConfig with valid params", func(t *testing.T) {
		err := validateLambdaConfig("key", "secret", "us-east-1")
		assert.NoError(t, err)
	})

	t.Run("validateLambdaConfig with empty access key", func(t *testing.T) {
		err := validateLambdaConfig("", "secret", "us-east-1")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "access key")
	})

	t.Run("validateLambdaConfig with empty secret key", func(t *testing.T) {
		err := validateLambdaConfig("key", "", "us-east-1")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "secret key")
	})

	t.Run("validateLambdaConfig with empty region", func(t *testing.T) {
		err := validateLambdaConfig("key", "secret", "")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "region")
	})
}

func TestLambdaActionExecute_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	action := NewInvokeFunctionAction("key", "secret", "us-east-1")
	config := map[string]interface{}{
		"function_name": "test-function",
		"payload":       map[string]interface{}{"key": "value"},
	}

	_, err := action.Execute(ctx, config, nil)
	require.Error(t, err)
}

func TestInvocationTypeValidation(t *testing.T) {
	validTypes := []string{"RequestResponse", "Event", "DryRun", ""}

	for _, invType := range validTypes {
		t.Run("valid type: "+invType, func(t *testing.T) {
			config := InvokeFunctionConfig{
				FunctionName:   "test-function",
				Payload:        map[string]interface{}{"key": "value"},
				InvocationType: invType,
			}
			err := config.Validate()
			assert.NoError(t, err)
		})
	}

	invalidTypes := []string{"InvalidType", "REQUESTRESPONSE", "event", "dryrun"}

	for _, invType := range invalidTypes {
		t.Run("invalid type: "+invType, func(t *testing.T) {
			config := InvokeFunctionConfig{
				FunctionName:   "test-function",
				Payload:        map[string]interface{}{"key": "value"},
				InvocationType: invType,
			}
			err := config.Validate()
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "invocation_type")
		})
	}
}
