package credential

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAPIKeyValidator_Validate(t *testing.T) {
	tests := []struct {
		name    string
		value   map[string]any
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid api key",
			value:   map[string]any{"key": "sk-test-12345"},
			wantErr: false,
		},
		{
			name:    "valid api key with prefix",
			value:   map[string]any{"key": "sk-test-12345", "prefix": "Bearer"},
			wantErr: false,
		},
		{
			name:    "missing key field",
			value:   map[string]any{},
			wantErr: true,
			errMsg:  "api key credential requires 'key' field",
		},
		{
			name:    "empty key value",
			value:   map[string]any{"key": ""},
			wantErr: true,
			errMsg:  "api key 'key' must be a non-empty string",
		},
		{
			name:    "key is not a string",
			value:   map[string]any{"key": 12345},
			wantErr: true,
			errMsg:  "api key 'key' must be a non-empty string",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator := &APIKeyValidator{}
			err := validator.Validate(tt.value)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestOAuth2Validator_Validate(t *testing.T) {
	tests := []struct {
		name    string
		value   map[string]any
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid oauth2 credentials",
			value: map[string]any{
				"client_id":     "my-client-id",
				"client_secret": "my-client-secret",
			},
			wantErr: false,
		},
		{
			name: "valid oauth2 with all fields",
			value: map[string]any{
				"client_id":     "my-client-id",
				"client_secret": "my-client-secret",
				"access_token":  "access-token-123",
				"refresh_token": "refresh-token-456",
				"token_url":     "https://auth.example.com/token",
			},
			wantErr: false,
		},
		{
			name:    "missing client_id",
			value:   map[string]any{"client_secret": "secret"},
			wantErr: true,
			errMsg:  "oauth2 credential requires 'client_id' field",
		},
		{
			name:    "missing client_secret",
			value:   map[string]any{"client_id": "client"},
			wantErr: true,
			errMsg:  "oauth2 credential requires 'client_secret' field",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator := &OAuth2Validator{}
			err := validator.Validate(tt.value)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestBasicAuthValidator_Validate(t *testing.T) {
	tests := []struct {
		name    string
		value   map[string]any
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid basic auth",
			value: map[string]any{
				"username": "admin",
				"password": "secret123",
			},
			wantErr: false,
		},
		{
			name:    "missing username",
			value:   map[string]any{"password": "secret"},
			wantErr: true,
			errMsg:  "basic auth credential requires 'username' field",
		},
		{
			name:    "missing password",
			value:   map[string]any{"username": "admin"},
			wantErr: true,
			errMsg:  "basic auth credential requires 'password' field",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator := &BasicAuthValidator{}
			err := validator.Validate(tt.value)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestBearerTokenValidator_Validate(t *testing.T) {
	tests := []struct {
		name    string
		value   map[string]any
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid bearer token",
			value:   map[string]any{"token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."},
			wantErr: false,
		},
		{
			name:    "missing token",
			value:   map[string]any{},
			wantErr: true,
			errMsg:  "bearer token credential requires 'token' field",
		},
		{
			name:    "empty token",
			value:   map[string]any{"token": ""},
			wantErr: true,
			errMsg:  "bearer token 'token' must be a non-empty string",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator := &BearerTokenValidator{}
			err := validator.Validate(tt.value)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestCustomValidator_Validate(t *testing.T) {
	tests := []struct {
		name    string
		value   map[string]any
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid custom credential",
			value:   map[string]any{"custom_field": "value"},
			wantErr: false,
		},
		{
			name:    "valid custom with multiple fields",
			value:   map[string]any{"field1": "value1", "field2": 123, "nested": map[string]any{"key": "val"}},
			wantErr: false,
		},
		{
			name:    "empty custom credential",
			value:   map[string]any{},
			wantErr: true,
			errMsg:  "custom credential value cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator := &CustomValidator{}
			err := validator.Validate(tt.value)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestDatabasePostgreSQLValidator_Validate(t *testing.T) {
	tests := []struct {
		name    string
		value   map[string]any
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid postgresql credentials",
			value: map[string]any{
				"host":     "localhost",
				"database": "mydb",
				"username": "admin",
				"password": "secret",
			},
			wantErr: false,
		},
		{
			name: "valid postgresql with port",
			value: map[string]any{
				"host":     "localhost",
				"database": "mydb",
				"username": "admin",
				"password": "secret",
				"port":     5432,
				"ssl_mode": "require",
			},
			wantErr: false,
		},
		{
			name: "missing host",
			value: map[string]any{
				"database": "mydb",
				"username": "admin",
				"password": "secret",
			},
			wantErr: true,
			errMsg:  "postgresql credential requires 'host' field",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator := &DatabasePostgreSQLValidator{}
			err := validator.Validate(tt.value)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestAWSSQSValidator_Validate(t *testing.T) {
	tests := []struct {
		name    string
		value   map[string]any
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid aws sqs credentials",
			value: map[string]any{
				"access_key_id":     "AKIAIOSFODNN7EXAMPLE",
				"secret_access_key": "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
				"region":            "us-east-1",
			},
			wantErr: false,
		},
		{
			name: "invalid region format",
			value: map[string]any{
				"access_key_id":     "AKIAIOSFODNN7EXAMPLE",
				"secret_access_key": "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
				"region":            "invalid-region",
			},
			wantErr: true,
			errMsg:  "aws sqs 'region' has invalid format",
		},
		{
			name: "missing access_key_id",
			value: map[string]any{
				"secret_access_key": "secret",
				"region":            "us-east-1",
			},
			wantErr: true,
			errMsg:  "aws sqs credential requires 'access_key_id' field",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator := &AWSSQSValidator{}
			err := validator.Validate(tt.value)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestKafkaValidator_Validate(t *testing.T) {
	tests := []struct {
		name    string
		value   map[string]any
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid kafka with string brokers",
			value:   map[string]any{"brokers": "localhost:9092"},
			wantErr: false,
		},
		{
			name:    "valid kafka with array brokers",
			value:   map[string]any{"brokers": []any{"kafka1:9092", "kafka2:9092"}},
			wantErr: false,
		},
		{
			name: "valid kafka with auth",
			value: map[string]any{
				"brokers":        "localhost:9092",
				"username":       "admin",
				"password":       "secret",
				"sasl_mechanism": "PLAIN",
			},
			wantErr: false,
		},
		{
			name:    "missing brokers",
			value:   map[string]any{},
			wantErr: true,
			errMsg:  "kafka credential requires 'brokers' field",
		},
		{
			name:    "empty brokers string",
			value:   map[string]any{"brokers": ""},
			wantErr: true,
			errMsg:  "kafka 'brokers' must not be empty",
		},
		{
			name:    "empty brokers array",
			value:   map[string]any{"brokers": []any{}},
			wantErr: true,
			errMsg:  "kafka 'brokers' must contain at least one broker",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator := &KafkaValidator{}
			err := validator.Validate(tt.value)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestSendGridValidator_Validate(t *testing.T) {
	tests := []struct {
		name    string
		value   map[string]any
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid sendgrid credentials",
			value:   map[string]any{"api_key": "SG.xxxx.yyyy"},
			wantErr: false,
		},
		{
			name:    "missing api_key",
			value:   map[string]any{},
			wantErr: true,
			errMsg:  "sendgrid credential requires 'api_key' field",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator := &SendGridValidator{}
			err := validator.Validate(tt.value)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestTwilioValidator_Validate(t *testing.T) {
	tests := []struct {
		name    string
		value   map[string]any
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid twilio credentials",
			value: map[string]any{
				"account_sid": "ACxxxxxxxx",
				"auth_token":  "xxxxxxxx",
			},
			wantErr: false,
		},
		{
			name:    "missing account_sid",
			value:   map[string]any{"auth_token": "token"},
			wantErr: true,
			errMsg:  "twilio credential requires 'account_sid' field",
		},
		{
			name:    "missing auth_token",
			value:   map[string]any{"account_sid": "ACxxx"},
			wantErr: true,
			errMsg:  "twilio credential requires 'auth_token' field",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator := &TwilioValidator{}
			err := validator.Validate(tt.value)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestGCSValidator_Validate(t *testing.T) {
	tests := []struct {
		name    string
		value   map[string]any
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid gcs with json string",
			value:   map[string]any{"service_account_json": `{"type":"service_account","project_id":"test"}`},
			wantErr: false,
		},
		{
			name: "valid gcs with json object",
			value: map[string]any{
				"service_account_json": map[string]any{
					"type":       "service_account",
					"project_id": "test",
				},
			},
			wantErr: false,
		},
		{
			name:    "missing service_account_json",
			value:   map[string]any{},
			wantErr: true,
			errMsg:  "gcs credential requires 'service_account_json' field",
		},
		{
			name:    "empty service_account_json string",
			value:   map[string]any{"service_account_json": ""},
			wantErr: true,
			errMsg:  "gcs 'service_account_json' must not be empty",
		},
		{
			name:    "invalid json string",
			value:   map[string]any{"service_account_json": "not-valid-json"},
			wantErr: true,
			errMsg:  "gcs 'service_account_json' must be valid JSON",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator := &GCSValidator{}
			err := validator.Validate(tt.value)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestAzureBlobValidator_Validate(t *testing.T) {
	tests := []struct {
		name    string
		value   map[string]any
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid azure blob with account_key",
			value: map[string]any{
				"account_name": "mystorageaccount",
				"account_key":  "base64encodedkey==",
			},
			wantErr: false,
		},
		{
			name: "valid azure blob with connection_string",
			value: map[string]any{
				"account_name":      "mystorageaccount",
				"connection_string": "DefaultEndpointsProtocol=https;...",
			},
			wantErr: false,
		},
		{
			name: "valid azure blob with sas_token",
			value: map[string]any{
				"account_name": "mystorageaccount",
				"sas_token":    "sv=2021-06-08&ss=bfqt...",
			},
			wantErr: false,
		},
		{
			name:    "missing account_name",
			value:   map[string]any{"account_key": "key"},
			wantErr: true,
			errMsg:  "azure blob credential requires 'account_name' field",
		},
		{
			name:    "missing auth method",
			value:   map[string]any{"account_name": "mystorageaccount"},
			wantErr: true,
			errMsg:  "azure blob credential requires one of: 'account_key', 'connection_string', or 'sas_token'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator := &AzureBlobValidator{}
			err := validator.Validate(tt.value)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestValidateCredentialValue(t *testing.T) {
	tests := []struct {
		name     string
		credType CredentialType
		value    map[string]any
		wantErr  bool
	}{
		{
			name:     "api_key type",
			credType: TypeAPIKey,
			value:    map[string]any{"key": "test-key"},
			wantErr:  false,
		},
		{
			name:     "oauth2 type",
			credType: TypeOAuth2,
			value:    map[string]any{"client_id": "id", "client_secret": "secret"},
			wantErr:  false,
		},
		{
			name:     "basic_auth type",
			credType: TypeBasicAuth,
			value:    map[string]any{"username": "user", "password": "pass"},
			wantErr:  false,
		},
		{
			name:     "custom type with any value",
			credType: TypeCustom,
			value:    map[string]any{"anything": "goes"},
			wantErr:  false,
		},
		{
			name:     "unknown type defaults to custom",
			credType: "unknown_type",
			value:    map[string]any{"field": "value"},
			wantErr:  false,
		},
		{
			name:     "invalid api_key",
			credType: TypeAPIKey,
			value:    map[string]any{},
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateCredentialValue(tt.credType, tt.value)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestGetTypeValidator(t *testing.T) {
	tests := []struct {
		name         string
		credType     CredentialType
		expectedType string
	}{
		{
			name:         "api_key returns APIKeyValidator",
			credType:     TypeAPIKey,
			expectedType: "*credential.APIKeyValidator",
		},
		{
			name:         "oauth2 returns OAuth2Validator",
			credType:     TypeOAuth2,
			expectedType: "*credential.OAuth2Validator",
		},
		{
			name:         "unknown type returns CustomValidator",
			credType:     "unknown",
			expectedType: "*credential.CustomValidator",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator := GetTypeValidator(tt.credType)
			require.NotNil(t, validator)
		})
	}
}

func TestGetCredentialTypeSchema(t *testing.T) {
	schema := GetCredentialTypeSchema(TypeAPIKey)

	assert.Equal(t, "api_key", schema["type"])
	assert.Contains(t, schema, "required_fields")
	assert.Contains(t, schema, "optional_fields")

	requiredFields := schema["required_fields"].([]string)
	assert.Contains(t, requiredFields, "key")
}

func TestGetAllCredentialTypeSchemas(t *testing.T) {
	schemas := GetAllCredentialTypeSchemas()

	// Should have at least the core types
	require.NotEmpty(t, schemas)

	// Check that each schema has required fields
	for _, schema := range schemas {
		assert.Contains(t, schema, "type")
		assert.Contains(t, schema, "required_fields")
		assert.Contains(t, schema, "optional_fields")
	}
}
