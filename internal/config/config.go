package config

import (
	"fmt"
	"os"
	"strconv"
)

// Config holds all application configuration
type Config struct {
	Server        ServerConfig
	Database      DatabaseConfig
	Redis         RedisConfig
	Kratos        KratosConfig
	Worker        WorkerConfig
	AWS           AWSConfig
	Queue         QueueConfig
	Credential    CredentialConfig
	Cleanup       CleanupConfig
	Retention     RetentionConfig
	Observability ObservabilityConfig
	Notification  NotificationConfig
}

// CredentialConfig holds credential vault configuration
type CredentialConfig struct {
	// MasterKey is the 32-byte encryption key for credentials (base64 encoded)
	// In production, this should come from a secure secret manager
	MasterKey string
	// UseKMS indicates whether to use AWS KMS for key management
	UseKMS bool
	// KMSKeyID is the AWS KMS key ID for production encryption
	KMSKeyID string
}

// ServerConfig holds HTTP server configuration
type ServerConfig struct {
	Address string
	Env     string
}

// DatabaseConfig holds PostgreSQL configuration
type DatabaseConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  string
}

// ConnectionString returns the PostgreSQL connection string
func (d DatabaseConfig) ConnectionString() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		d.Host, d.Port, d.User, d.Password, d.DBName, d.SSLMode,
	)
}

// RedisConfig holds Redis configuration
type RedisConfig struct {
	Address  string
	Password string
	DB       int
}

// KratosConfig holds Ory Kratos configuration
type KratosConfig struct {
	PublicURL string
	AdminURL  string
}

// WorkerConfig holds worker configuration
type WorkerConfig struct {
	Concurrency             int
	MaxConcurrencyPerTenant int
	HealthPort              string
	QueueURL                string
}

// AWSConfig holds AWS configuration
type AWSConfig struct {
	Region          string
	AccessKeyID     string
	SecretAccessKey string
	Endpoint        string // For LocalStack or custom endpoints
	S3Bucket        string
	SQSQueueURL     string
	SQSDLQueueURL   string // Dead-letter queue URL
}

// QueueConfig holds queue-specific configuration
type QueueConfig struct {
	Enabled            bool
	MaxMessages        int32
	WaitTimeSeconds    int32
	VisibilityTimeout  int32
	MaxRetries         int
	ProcessTimeout     int // in seconds
	PollInterval       int // in seconds
	ConcurrentWorkers  int
	DeleteAfterProcess bool
}

// CleanupConfig holds webhook event cleanup configuration
type CleanupConfig struct {
	// Enabled indicates whether cleanup is enabled
	Enabled bool
	// RetentionDays is the number of days to retain webhook events (default: 30)
	RetentionDays int
	// BatchSize is the number of events to delete per batch (default: 1000)
	BatchSize int
	// Schedule is the cron schedule for cleanup (default: "0 0 * * *" - daily at midnight)
	Schedule string
}

// RetentionConfig holds execution retention policy configuration
type RetentionConfig struct {
	// Enabled indicates whether retention cleanup is enabled
	Enabled bool
	// DefaultRetentionDays is the default retention period in days (default: 90)
	DefaultRetentionDays int
	// BatchSize is the number of executions to delete per batch (default: 1000)
	BatchSize int
	// RunInterval is how often to run cleanup (default: 24h)
	RunInterval string
	// EnableAuditLog enables audit logging of cleanup operations
	EnableAuditLog bool
}

// ObservabilityConfig holds observability configuration
type ObservabilityConfig struct {
	// Metrics configuration
	MetricsEnabled bool
	MetricsPort    string

	// Tracing configuration
	TracingEnabled     bool
	TracingEndpoint    string // OTLP endpoint (e.g., "localhost:4317")
	TracingSampleRate  float64
	TracingServiceName string

	// Error tracking configuration
	SentryEnabled     bool
	SentryDSN         string
	SentryEnvironment string
	SentrySampleRate  float64
}

// NotificationConfig holds notification configuration
type NotificationConfig struct {
	// Enabled channels
	EnableEmail bool
	EnableSlack bool
	EnableInApp bool

	// Email settings
	EmailProvider          string // smtp or ses
	EmailFrom              string
	SMTPHost               string
	SMTPPort               int
	SMTPUser               string
	SMTPPass               string
	SMTPTLS                bool
	EmailMaxRetries        int
	EmailRetryDelaySeconds int

	// AWS SES settings
	SESRegion string

	// Slack settings
	SlackWebhookURL        string
	SlackMaxRetries        int
	SlackRetryDelaySeconds int
	SlackTimeoutSeconds    int

	// In-app notification settings
	InAppRetentionDays int
}

// Load reads configuration from environment variables
func Load() (*Config, error) {
	cfg := &Config{
		Server: ServerConfig{
			Address: getEnv("SERVER_ADDRESS", ":8080"),
			Env:     getEnv("APP_ENV", "development"),
		},
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnvAsInt("DB_PORT", 5433),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", "postgres"),
			DBName:   getEnv("DB_NAME", "gorax"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},
		Redis: RedisConfig{
			Address:  getEnv("REDIS_ADDRESS", "localhost:6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       getEnvAsInt("REDIS_DB", 0),
		},
		Kratos: KratosConfig{
			PublicURL: getEnv("KRATOS_PUBLIC_URL", "http://localhost:4433"),
			AdminURL:  getEnv("KRATOS_ADMIN_URL", "http://localhost:4434"),
		},
		Worker: WorkerConfig{
			Concurrency:             getEnvAsInt("WORKER_CONCURRENCY", 10),
			MaxConcurrencyPerTenant: getEnvAsInt("WORKER_MAX_CONCURRENCY_PER_TENANT", 10),
			HealthPort:              getEnv("WORKER_HEALTH_PORT", "8081"),
			QueueURL:                getEnv("WORKER_QUEUE_URL", ""),
		},
		AWS: AWSConfig{
			Region:          getEnv("AWS_REGION", "us-east-1"),
			AccessKeyID:     getEnv("AWS_ACCESS_KEY_ID", ""),
			SecretAccessKey: getEnv("AWS_SECRET_ACCESS_KEY", ""),
			Endpoint:        getEnv("AWS_ENDPOINT", ""), // For LocalStack
			S3Bucket:        getEnv("AWS_S3_BUCKET", "gorax-artifacts"),
			SQSQueueURL:     getEnv("AWS_SQS_QUEUE_URL", ""),
			SQSDLQueueURL:   getEnv("AWS_SQS_DLQ_URL", ""),
		},
		Queue: QueueConfig{
			Enabled:            getEnvAsBool("QUEUE_ENABLED", false),
			MaxMessages:        int32(getEnvAsInt("QUEUE_MAX_MESSAGES", 10)),
			WaitTimeSeconds:    int32(getEnvAsInt("QUEUE_WAIT_TIME_SECONDS", 20)),
			VisibilityTimeout:  int32(getEnvAsInt("QUEUE_VISIBILITY_TIMEOUT", 30)),
			MaxRetries:         getEnvAsInt("QUEUE_MAX_RETRIES", 3),
			ProcessTimeout:     getEnvAsInt("QUEUE_PROCESS_TIMEOUT", 300), // 5 minutes
			PollInterval:       getEnvAsInt("QUEUE_POLL_INTERVAL", 1),
			ConcurrentWorkers:  getEnvAsInt("QUEUE_CONCURRENT_WORKERS", 10),
			DeleteAfterProcess: getEnvAsBool("QUEUE_DELETE_AFTER_PROCESS", true),
		},
		Credential: CredentialConfig{
			// Default development key (32 bytes base64 encoded) - DO NOT USE IN PRODUCTION
			MasterKey: getEnv("CREDENTIAL_MASTER_KEY", "dGhpcy1pcy1hLTMyLWJ5dGUtZGV2LWtleS0xMjM0NTY="),
			UseKMS:    getEnvAsBool("CREDENTIAL_USE_KMS", false),
			KMSKeyID:  getEnv("CREDENTIAL_KMS_KEY_ID", ""),
		},
		Cleanup: CleanupConfig{
			Enabled:       getEnvAsBool("CLEANUP_ENABLED", true),
			RetentionDays: getEnvAsInt("CLEANUP_RETENTION_DAYS", 30),
			BatchSize:     getEnvAsInt("CLEANUP_BATCH_SIZE", 1000),
			Schedule:      getEnv("CLEANUP_SCHEDULE", "0 0 * * *"), // Daily at midnight
		},
		Retention: RetentionConfig{
			Enabled:              getEnvAsBool("RETENTION_ENABLED", true),
			DefaultRetentionDays: getEnvAsInt("RETENTION_DEFAULT_DAYS", 90),
			BatchSize:            getEnvAsInt("RETENTION_BATCH_SIZE", 1000),
			RunInterval:          getEnv("RETENTION_RUN_INTERVAL", "24h"),
			EnableAuditLog:       getEnvAsBool("RETENTION_ENABLE_AUDIT_LOG", true),
		},
		Observability: ObservabilityConfig{
			MetricsEnabled:     getEnvAsBool("METRICS_ENABLED", true),
			MetricsPort:        getEnv("METRICS_PORT", "9090"),
			TracingEnabled:     getEnvAsBool("TRACING_ENABLED", false),
			TracingEndpoint:    getEnv("TRACING_ENDPOINT", "localhost:4317"),
			TracingSampleRate:  getEnvAsFloat("TRACING_SAMPLE_RATE", 1.0),
			TracingServiceName: getEnv("TRACING_SERVICE_NAME", "gorax"),
			SentryEnabled:      getEnvAsBool("SENTRY_ENABLED", false),
			SentryDSN:          getEnv("SENTRY_DSN", ""),
			SentryEnvironment:  getEnv("SENTRY_ENVIRONMENT", "development"),
			SentrySampleRate:   getEnvAsFloat("SENTRY_SAMPLE_RATE", 1.0),
		},
		Notification: NotificationConfig{
			EnableEmail:            getEnvAsBool("NOTIFICATION_ENABLE_EMAIL", false),
			EnableSlack:            getEnvAsBool("NOTIFICATION_ENABLE_SLACK", false),
			EnableInApp:            getEnvAsBool("NOTIFICATION_ENABLE_INAPP", true),
			EmailProvider:          getEnv("NOTIFICATION_EMAIL_PROVIDER", "smtp"),
			EmailFrom:              getEnv("NOTIFICATION_EMAIL_FROM", "noreply@example.com"),
			SMTPHost:               getEnv("NOTIFICATION_SMTP_HOST", ""),
			SMTPPort:               getEnvAsInt("NOTIFICATION_SMTP_PORT", 587),
			SMTPUser:               getEnv("NOTIFICATION_SMTP_USER", ""),
			SMTPPass:               getEnv("NOTIFICATION_SMTP_PASS", ""),
			SMTPTLS:                getEnvAsBool("NOTIFICATION_SMTP_TLS", true),
			EmailMaxRetries:        getEnvAsInt("NOTIFICATION_EMAIL_MAX_RETRIES", 3),
			EmailRetryDelaySeconds: getEnvAsInt("NOTIFICATION_EMAIL_RETRY_DELAY_SECONDS", 1),
			SESRegion:              getEnv("NOTIFICATION_SES_REGION", "us-east-1"),
			SlackWebhookURL:        getEnv("NOTIFICATION_SLACK_WEBHOOK_URL", ""),
			SlackMaxRetries:        getEnvAsInt("NOTIFICATION_SLACK_MAX_RETRIES", 3),
			SlackRetryDelaySeconds: getEnvAsInt("NOTIFICATION_SLACK_RETRY_DELAY_SECONDS", 1),
			SlackTimeoutSeconds:    getEnvAsInt("NOTIFICATION_SLACK_TIMEOUT_SECONDS", 30),
			InAppRetentionDays:     getEnvAsInt("NOTIFICATION_INAPP_RETENTION_DAYS", 90),
		},
	}

	return cfg, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvAsBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}

func getEnvAsFloat(key string, defaultValue float64) float64 {
	if value := os.Getenv(key); value != "" {
		if floatValue, err := strconv.ParseFloat(value, 64); err == nil {
			return floatValue
		}
	}
	return defaultValue
}
