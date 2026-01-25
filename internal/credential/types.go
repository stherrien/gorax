package credential

import (
	"encoding/json"
	"fmt"
	"regexp"
	"time"
)

// CredentialTypeValidator provides type-specific validation for credentials
type CredentialTypeValidator interface {
	// Validate validates the credential value for the specific type
	Validate(value map[string]any) error
	// RequiredFields returns the list of required fields for this credential type
	RequiredFields() []string
	// OptionalFields returns the list of optional fields for this credential type
	OptionalFields() []string
}

// typeValidators maps credential types to their validators
var typeValidators = map[CredentialType]CredentialTypeValidator{
	TypeAPIKey:             &APIKeyValidator{},
	TypeOAuth2:             &OAuth2Validator{},
	TypeBasicAuth:          &BasicAuthValidator{},
	TypeBearerToken:        &BearerTokenValidator{},
	TypeCustom:             &CustomValidator{},
	TypeDatabasePostgreSQL: &DatabasePostgreSQLValidator{},
	TypeDatabaseMySQL:      &DatabaseMySQLValidator{},
	TypeDatabaseSQLite:     &DatabaseSQLiteValidator{},
	TypeDatabaseMongoDB:    &DatabaseMongoDBValidator{},
	TypeQueueAWSSQS:        &AWSSQSValidator{},
	TypeQueueKafka:         &KafkaValidator{},
	TypeQueueRabbitMQ:      &RabbitMQValidator{},
	TypeEmailSendGrid:      &SendGridValidator{},
	TypeEmailMailgun:       &MailgunValidator{},
	TypeEmailAWSSES:        &AWSSESValidator{},
	TypeEmailSMTP:          &SMTPValidator{},
	TypeSMSTwilio:          &TwilioValidator{},
	TypeSMSAWSSNS:          &AWSSNSValidator{},
	TypeSMSMessageBird:     &MessageBirdValidator{},
	TypeStorageAWSS3:       &AWSS3Validator{},
	TypeStorageGCS:         &GCSValidator{},
	TypeStorageAzureBlob:   &AzureBlobValidator{},
}

// GetTypeValidator returns the validator for a credential type
func GetTypeValidator(credType CredentialType) CredentialTypeValidator {
	if validator, ok := typeValidators[credType]; ok {
		return validator
	}
	return &CustomValidator{} // Default to custom validator
}

// ValidateCredentialValue validates credential value based on type
func ValidateCredentialValue(credType CredentialType, value map[string]any) error {
	validator := GetTypeValidator(credType)
	return validator.Validate(value)
}

// --- API Key Credential ---

// APIKeyCredential represents an API key credential structure
type APIKeyCredential struct {
	Key    string `json:"key"`
	Prefix string `json:"prefix,omitempty"` // Optional prefix (e.g., "Bearer", "Api-Key")
}

// APIKeyValidator validates API key credentials
type APIKeyValidator struct{}

func (v *APIKeyValidator) RequiredFields() []string {
	return []string{"key"}
}

func (v *APIKeyValidator) OptionalFields() []string {
	return []string{"prefix"}
}

func (v *APIKeyValidator) Validate(value map[string]any) error {
	key, ok := value["key"]
	if !ok {
		return &ValidationError{Message: "api key credential requires 'key' field"}
	}
	keyStr, ok := key.(string)
	if !ok || keyStr == "" {
		return &ValidationError{Message: "api key 'key' must be a non-empty string"}
	}
	return nil
}

// --- OAuth2 Credential ---

// OAuth2Credential represents an OAuth2 credential structure
type OAuth2Credential struct {
	ClientID     string     `json:"client_id"`
	ClientSecret string     `json:"client_secret"`
	AccessToken  string     `json:"access_token,omitempty"`
	RefreshToken string     `json:"refresh_token,omitempty"`
	TokenURL     string     `json:"token_url,omitempty"`
	Scopes       []string   `json:"scopes,omitempty"`
	ExpiresAt    *time.Time `json:"expires_at,omitempty"`
}

// OAuth2Validator validates OAuth2 credentials
type OAuth2Validator struct{}

func (v *OAuth2Validator) RequiredFields() []string {
	return []string{"client_id", "client_secret"}
}

func (v *OAuth2Validator) OptionalFields() []string {
	return []string{"access_token", "refresh_token", "token_url", "scopes", "expires_at"}
}

func (v *OAuth2Validator) Validate(value map[string]any) error {
	clientID, ok := value["client_id"]
	if !ok {
		return &ValidationError{Message: "oauth2 credential requires 'client_id' field"}
	}
	if _, ok := clientID.(string); !ok {
		return &ValidationError{Message: "oauth2 'client_id' must be a string"}
	}

	clientSecret, ok := value["client_secret"]
	if !ok {
		return &ValidationError{Message: "oauth2 credential requires 'client_secret' field"}
	}
	if _, ok := clientSecret.(string); !ok {
		return &ValidationError{Message: "oauth2 'client_secret' must be a string"}
	}

	return nil
}

// --- Basic Auth Credential ---

// BasicAuthCredential represents a basic auth credential structure
type BasicAuthCredential struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// BasicAuthValidator validates basic auth credentials
type BasicAuthValidator struct{}

func (v *BasicAuthValidator) RequiredFields() []string {
	return []string{"username", "password"}
}

func (v *BasicAuthValidator) OptionalFields() []string {
	return []string{}
}

func (v *BasicAuthValidator) Validate(value map[string]any) error {
	username, ok := value["username"]
	if !ok {
		return &ValidationError{Message: "basic auth credential requires 'username' field"}
	}
	if _, ok := username.(string); !ok {
		return &ValidationError{Message: "basic auth 'username' must be a string"}
	}

	password, ok := value["password"]
	if !ok {
		return &ValidationError{Message: "basic auth credential requires 'password' field"}
	}
	if _, ok := password.(string); !ok {
		return &ValidationError{Message: "basic auth 'password' must be a string"}
	}

	return nil
}

// --- Bearer Token Credential ---

// BearerTokenCredential represents a bearer token credential structure
type BearerTokenCredential struct {
	Token     string     `json:"token"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
}

// BearerTokenValidator validates bearer token credentials
type BearerTokenValidator struct{}

func (v *BearerTokenValidator) RequiredFields() []string {
	return []string{"token"}
}

func (v *BearerTokenValidator) OptionalFields() []string {
	return []string{"expires_at"}
}

func (v *BearerTokenValidator) Validate(value map[string]any) error {
	token, ok := value["token"]
	if !ok {
		return &ValidationError{Message: "bearer token credential requires 'token' field"}
	}
	tokenStr, ok := token.(string)
	if !ok || tokenStr == "" {
		return &ValidationError{Message: "bearer token 'token' must be a non-empty string"}
	}
	return nil
}

// --- Custom Credential ---

// CustomValidator validates custom credentials (allows any structure)
type CustomValidator struct{}

func (v *CustomValidator) RequiredFields() []string {
	return []string{}
}

func (v *CustomValidator) OptionalFields() []string {
	return []string{"*"}
}

func (v *CustomValidator) Validate(value map[string]any) error {
	// Custom credentials allow any structure, just verify it's not empty
	if len(value) == 0 {
		return &ValidationError{Message: "custom credential value cannot be empty"}
	}
	return nil
}

// --- Database PostgreSQL Credential ---

// DatabasePostgreSQLValidator validates PostgreSQL database credentials
type DatabasePostgreSQLValidator struct{}

func (v *DatabasePostgreSQLValidator) RequiredFields() []string {
	return []string{"host", "database", "username", "password"}
}

func (v *DatabasePostgreSQLValidator) OptionalFields() []string {
	return []string{"port", "ssl_mode"}
}

func (v *DatabasePostgreSQLValidator) Validate(value map[string]any) error {
	return validateDatabaseCredential(value, "postgresql")
}

// --- Database MySQL Credential ---

// DatabaseMySQLValidator validates MySQL database credentials
type DatabaseMySQLValidator struct{}

func (v *DatabaseMySQLValidator) RequiredFields() []string {
	return []string{"host", "database", "username", "password"}
}

func (v *DatabaseMySQLValidator) OptionalFields() []string {
	return []string{"port"}
}

func (v *DatabaseMySQLValidator) Validate(value map[string]any) error {
	return validateDatabaseCredential(value, "mysql")
}

// --- Database SQLite Credential ---

// DatabaseSQLiteValidator validates SQLite database credentials
type DatabaseSQLiteValidator struct{}

func (v *DatabaseSQLiteValidator) RequiredFields() []string {
	return []string{"path"}
}

func (v *DatabaseSQLiteValidator) OptionalFields() []string {
	return []string{"password"}
}

func (v *DatabaseSQLiteValidator) Validate(value map[string]any) error {
	path, ok := value["path"]
	if !ok {
		return &ValidationError{Message: "sqlite credential requires 'path' field"}
	}
	if _, ok := path.(string); !ok {
		return &ValidationError{Message: "sqlite 'path' must be a string"}
	}
	return nil
}

// --- Database MongoDB Credential ---

// DatabaseMongoDBValidator validates MongoDB database credentials
type DatabaseMongoDBValidator struct{}

func (v *DatabaseMongoDBValidator) RequiredFields() []string {
	return []string{"connection_string"}
}

func (v *DatabaseMongoDBValidator) OptionalFields() []string {
	return []string{"database", "username", "password"}
}

func (v *DatabaseMongoDBValidator) Validate(value map[string]any) error {
	connStr, ok := value["connection_string"]
	if !ok {
		return &ValidationError{Message: "mongodb credential requires 'connection_string' field"}
	}
	if _, ok := connStr.(string); !ok {
		return &ValidationError{Message: "mongodb 'connection_string' must be a string"}
	}
	return nil
}

// --- AWS SQS Credential ---

// AWSSQSValidator validates AWS SQS credentials
type AWSSQSValidator struct{}

func (v *AWSSQSValidator) RequiredFields() []string {
	return []string{"access_key_id", "secret_access_key", "region"}
}

func (v *AWSSQSValidator) OptionalFields() []string {
	return []string{"queue_url", "session_token"}
}

func (v *AWSSQSValidator) Validate(value map[string]any) error {
	return validateAWSCredential(value, "sqs")
}

// --- Kafka Credential ---

// KafkaValidator validates Kafka credentials
type KafkaValidator struct{}

func (v *KafkaValidator) RequiredFields() []string {
	return []string{"brokers"}
}

func (v *KafkaValidator) OptionalFields() []string {
	return []string{"username", "password", "sasl_mechanism", "tls_enabled"}
}

func (v *KafkaValidator) Validate(value map[string]any) error {
	brokers, ok := value["brokers"]
	if !ok {
		return &ValidationError{Message: "kafka credential requires 'brokers' field"}
	}
	// brokers can be a string or array of strings
	switch b := brokers.(type) {
	case string:
		if b == "" {
			return &ValidationError{Message: "kafka 'brokers' must not be empty"}
		}
	case []any:
		if len(b) == 0 {
			return &ValidationError{Message: "kafka 'brokers' must contain at least one broker"}
		}
	default:
		return &ValidationError{Message: "kafka 'brokers' must be a string or array"}
	}
	return nil
}

// --- RabbitMQ Credential ---

// RabbitMQValidator validates RabbitMQ credentials
type RabbitMQValidator struct{}

func (v *RabbitMQValidator) RequiredFields() []string {
	return []string{"host", "username", "password"}
}

func (v *RabbitMQValidator) OptionalFields() []string {
	return []string{"port", "vhost", "tls_enabled"}
}

func (v *RabbitMQValidator) Validate(value map[string]any) error {
	required := []string{"host", "username", "password"}
	for _, field := range required {
		if _, ok := value[field]; !ok {
			return &ValidationError{Message: fmt.Sprintf("rabbitmq credential requires '%s' field", field)}
		}
	}
	return nil
}

// --- SendGrid Credential ---

// SendGridValidator validates SendGrid credentials
type SendGridValidator struct{}

func (v *SendGridValidator) RequiredFields() []string {
	return []string{"api_key"}
}

func (v *SendGridValidator) OptionalFields() []string {
	return []string{"from_email", "from_name"}
}

func (v *SendGridValidator) Validate(value map[string]any) error {
	return validateAPIKeyCredential(value, "sendgrid")
}

// --- Mailgun Credential ---

// MailgunValidator validates Mailgun credentials
type MailgunValidator struct{}

func (v *MailgunValidator) RequiredFields() []string {
	return []string{"api_key", "domain"}
}

func (v *MailgunValidator) OptionalFields() []string {
	return []string{"from_email", "from_name"}
}

func (v *MailgunValidator) Validate(value map[string]any) error {
	if err := validateAPIKeyCredential(value, "mailgun"); err != nil {
		return err
	}
	if _, ok := value["domain"]; !ok {
		return &ValidationError{Message: "mailgun credential requires 'domain' field"}
	}
	return nil
}

// --- AWS SES Credential ---

// AWSSESValidator validates AWS SES credentials
type AWSSESValidator struct{}

func (v *AWSSESValidator) RequiredFields() []string {
	return []string{"access_key_id", "secret_access_key", "region"}
}

func (v *AWSSESValidator) OptionalFields() []string {
	return []string{"from_email", "session_token"}
}

func (v *AWSSESValidator) Validate(value map[string]any) error {
	return validateAWSCredential(value, "ses")
}

// --- SMTP Credential ---

// SMTPValidator validates SMTP credentials
type SMTPValidator struct{}

func (v *SMTPValidator) RequiredFields() []string {
	return []string{"host", "username", "password"}
}

func (v *SMTPValidator) OptionalFields() []string {
	return []string{"port", "from_email", "tls_enabled", "start_tls"}
}

func (v *SMTPValidator) Validate(value map[string]any) error {
	required := []string{"host", "username", "password"}
	for _, field := range required {
		if _, ok := value[field]; !ok {
			return &ValidationError{Message: fmt.Sprintf("smtp credential requires '%s' field", field)}
		}
	}
	return nil
}

// --- Twilio Credential ---

// TwilioValidator validates Twilio credentials
type TwilioValidator struct{}

func (v *TwilioValidator) RequiredFields() []string {
	return []string{"account_sid", "auth_token"}
}

func (v *TwilioValidator) OptionalFields() []string {
	return []string{"from_number", "messaging_service_sid"}
}

func (v *TwilioValidator) Validate(value map[string]any) error {
	required := []string{"account_sid", "auth_token"}
	for _, field := range required {
		val, ok := value[field]
		if !ok {
			return &ValidationError{Message: fmt.Sprintf("twilio credential requires '%s' field", field)}
		}
		if _, ok := val.(string); !ok {
			return &ValidationError{Message: fmt.Sprintf("twilio '%s' must be a string", field)}
		}
	}
	return nil
}

// --- AWS SNS Credential ---

// AWSSNSValidator validates AWS SNS credentials
type AWSSNSValidator struct{}

func (v *AWSSNSValidator) RequiredFields() []string {
	return []string{"access_key_id", "secret_access_key", "region"}
}

func (v *AWSSNSValidator) OptionalFields() []string {
	return []string{"topic_arn", "session_token"}
}

func (v *AWSSNSValidator) Validate(value map[string]any) error {
	return validateAWSCredential(value, "sns")
}

// --- MessageBird Credential ---

// MessageBirdValidator validates MessageBird credentials
type MessageBirdValidator struct{}

func (v *MessageBirdValidator) RequiredFields() []string {
	return []string{"api_key"}
}

func (v *MessageBirdValidator) OptionalFields() []string {
	return []string{"originator"}
}

func (v *MessageBirdValidator) Validate(value map[string]any) error {
	return validateAPIKeyCredential(value, "messagebird")
}

// --- AWS S3 Credential ---

// AWSS3Validator validates AWS S3 credentials
type AWSS3Validator struct{}

func (v *AWSS3Validator) RequiredFields() []string {
	return []string{"access_key_id", "secret_access_key", "region"}
}

func (v *AWSS3Validator) OptionalFields() []string {
	return []string{"bucket", "session_token", "endpoint"}
}

func (v *AWSS3Validator) Validate(value map[string]any) error {
	return validateAWSCredential(value, "s3")
}

// --- Google Cloud Storage Credential ---

// GCSValidator validates Google Cloud Storage credentials
type GCSValidator struct{}

func (v *GCSValidator) RequiredFields() []string {
	return []string{"service_account_json"}
}

func (v *GCSValidator) OptionalFields() []string {
	return []string{"bucket", "project_id"}
}

func (v *GCSValidator) Validate(value map[string]any) error {
	saJSON, ok := value["service_account_json"]
	if !ok {
		return &ValidationError{Message: "gcs credential requires 'service_account_json' field"}
	}

	// service_account_json can be a string (JSON) or a map
	switch sa := saJSON.(type) {
	case string:
		if sa == "" {
			return &ValidationError{Message: "gcs 'service_account_json' must not be empty"}
		}
		// Validate it's valid JSON
		var js json.RawMessage
		if err := json.Unmarshal([]byte(sa), &js); err != nil {
			return &ValidationError{Message: "gcs 'service_account_json' must be valid JSON"}
		}
	case map[string]any:
		if len(sa) == 0 {
			return &ValidationError{Message: "gcs 'service_account_json' must not be empty"}
		}
	default:
		return &ValidationError{Message: "gcs 'service_account_json' must be a JSON string or object"}
	}
	return nil
}

// --- Azure Blob Storage Credential ---

// AzureBlobValidator validates Azure Blob Storage credentials
type AzureBlobValidator struct{}

func (v *AzureBlobValidator) RequiredFields() []string {
	return []string{"account_name"}
}

func (v *AzureBlobValidator) OptionalFields() []string {
	return []string{"account_key", "connection_string", "sas_token", "container"}
}

func (v *AzureBlobValidator) Validate(value map[string]any) error {
	accountName, ok := value["account_name"]
	if !ok {
		return &ValidationError{Message: "azure blob credential requires 'account_name' field"}
	}
	if _, ok := accountName.(string); !ok {
		return &ValidationError{Message: "azure blob 'account_name' must be a string"}
	}

	// Must have at least one auth method
	hasKey := value["account_key"] != nil
	hasConnStr := value["connection_string"] != nil
	hasSAS := value["sas_token"] != nil

	if !hasKey && !hasConnStr && !hasSAS {
		return &ValidationError{Message: "azure blob credential requires one of: 'account_key', 'connection_string', or 'sas_token'"}
	}

	return nil
}

// --- Helper validation functions ---

// validateDatabaseCredential validates common database credential fields
func validateDatabaseCredential(value map[string]any, dbType string) error {
	required := []string{"host", "database", "username", "password"}
	for _, field := range required {
		val, ok := value[field]
		if !ok {
			return &ValidationError{Message: fmt.Sprintf("%s credential requires '%s' field", dbType, field)}
		}
		if _, ok := val.(string); !ok {
			return &ValidationError{Message: fmt.Sprintf("%s '%s' must be a string", dbType, field)}
		}
	}
	return nil
}

// validateAWSCredential validates common AWS credential fields
func validateAWSCredential(value map[string]any, service string) error {
	required := []string{"access_key_id", "secret_access_key", "region"}
	for _, field := range required {
		val, ok := value[field]
		if !ok {
			return &ValidationError{Message: fmt.Sprintf("aws %s credential requires '%s' field", service, field)}
		}
		if _, ok := val.(string); !ok {
			return &ValidationError{Message: fmt.Sprintf("aws %s '%s' must be a string", service, field)}
		}
	}

	// Validate region format
	region, _ := value["region"].(string)
	regionRegex := regexp.MustCompile(`^[a-z]{2}-[a-z]+-\d+$`)
	if !regionRegex.MatchString(region) {
		return &ValidationError{Message: fmt.Sprintf("aws %s 'region' has invalid format (expected: us-east-1, eu-west-2, etc.)", service)}
	}

	return nil
}

// validateAPIKeyCredential validates simple API key credential
func validateAPIKeyCredential(value map[string]any, service string) error {
	apiKey, ok := value["api_key"]
	if !ok {
		return &ValidationError{Message: fmt.Sprintf("%s credential requires 'api_key' field", service)}
	}
	keyStr, ok := apiKey.(string)
	if !ok || keyStr == "" {
		return &ValidationError{Message: fmt.Sprintf("%s 'api_key' must be a non-empty string", service)}
	}
	return nil
}

// GetCredentialTypeSchema returns the schema information for a credential type
func GetCredentialTypeSchema(credType CredentialType) map[string]any {
	validator := GetTypeValidator(credType)
	return map[string]any{
		"type":            string(credType),
		"required_fields": validator.RequiredFields(),
		"optional_fields": validator.OptionalFields(),
	}
}

// GetAllCredentialTypeSchemas returns schemas for all supported credential types
func GetAllCredentialTypeSchemas() []map[string]any {
	var schemas []map[string]any
	for credType := range typeValidators {
		schemas = append(schemas, GetCredentialTypeSchema(credType))
	}
	return schemas
}
