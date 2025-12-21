package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Config holds all application configuration
type Config struct {
	Server         ServerConfig
	Database       DatabaseConfig
	Redis          RedisConfig
	Kratos         KratosConfig
	Worker         WorkerConfig
	AWS            AWSConfig
	Queue          QueueConfig
	Credential     CredentialConfig
	Cleanup        CleanupConfig
	Retention      RetentionConfig
	Observability  ObservabilityConfig
	Notification   NotificationConfig
	CORS           CORSConfig
	SecurityHeader SecurityHeaderConfig
	AIBuilder      AIBuilderConfig
}

// AIBuilderConfig holds AI Workflow Builder configuration
type AIBuilderConfig struct {
	// Enabled indicates whether the AI Builder feature is enabled
	Enabled bool
	// Provider is the LLM provider to use (openai, anthropic, bedrock)
	Provider string
	// APIKey is the API key for the LLM provider
	// For AWS Bedrock, use AWS credentials from AWSConfig instead
	APIKey string
	// Model is the model name/ID to use for generation
	Model string
	// MaxTokens is the maximum number of tokens for completion
	MaxTokens int
	// Temperature controls randomness in generation (0.0 to 1.0)
	Temperature float64
}

// CredentialConfig holds credential vault configuration
type CredentialConfig struct {
	// MasterKey is the 32-byte encryption key for credentials (base64 encoded)
	// In production, this should come from a secure secret manager
	MasterKey string
	// UseKMS indicates whether to use AWS KMS for key management
	UseKMS bool
	// KMSKeyID is the AWS KMS key ID or alias for production encryption
	KMSKeyID string
	// KMSRegion is the AWS region for KMS operations (defaults to AWS_REGION if not set)
	KMSRegion string
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

// CORSConfig holds CORS configuration
type CORSConfig struct {
	// AllowedOrigins is the list of allowed origins for CORS
	// Development: Can include localhost origins
	// Production: Must use HTTPS origins only, no localhost
	AllowedOrigins []string
	// AllowedMethods is the list of allowed HTTP methods
	AllowedMethods []string
	// AllowedHeaders is the list of allowed HTTP headers
	AllowedHeaders []string
	// ExposedHeaders is the list of headers exposed to the client
	ExposedHeaders []string
	// AllowCredentials indicates whether credentials are allowed
	AllowCredentials bool
	// MaxAge is the preflight cache duration in seconds
	MaxAge int
}

// SecurityHeaderConfig holds security headers configuration
type SecurityHeaderConfig struct {
	// EnableHSTS controls whether to set Strict-Transport-Security header
	EnableHSTS bool
	// HSTSMaxAge is the max-age value for HSTS in seconds
	HSTSMaxAge int
	// CSPDirectives is the Content-Security-Policy directive
	CSPDirectives string
	// FrameOptions controls X-Frame-Options header (DENY or SAMEORIGIN)
	FrameOptions string
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
			// KMSRegion defaults to AWS_REGION if not explicitly set
			KMSRegion: getEnvWithFallback("CREDENTIAL_KMS_REGION", "AWS_REGION", "us-east-1"),
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
		CORS:           loadCORSConfig(),
		SecurityHeader: loadSecurityHeaderConfig(),
		AIBuilder: AIBuilderConfig{
			Enabled:     getEnvAsBool("AI_BUILDER_ENABLED", false),
			Provider:    getEnv("AI_BUILDER_PROVIDER", "openai"),
			APIKey:      getEnv("AI_BUILDER_API_KEY", ""),
			Model:       getEnv("AI_BUILDER_MODEL", "gpt-4"),
			MaxTokens:   getEnvAsInt("AI_BUILDER_MAX_TOKENS", 4096),
			Temperature: getEnvAsFloat("AI_BUILDER_TEMPERATURE", 0.7),
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

func getEnvAsSlice(key string, defaultValue []string) []string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	// Split by comma and trim whitespace
	parts := strings.Split(value, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

// getEnvWithFallback gets an environment variable with a fallback to another env var
func getEnvWithFallback(key, fallbackKey, defaultValue string) string {
	// Try primary key first
	if value := os.Getenv(key); value != "" {
		return value
	}
	// Try fallback key
	if value := os.Getenv(fallbackKey); value != "" {
		return value
	}
	// Return default
	return defaultValue
}

func loadCORSConfig() CORSConfig {
	return CORSConfig{
		AllowedOrigins: getEnvAsSlice("CORS_ALLOWED_ORIGINS", []string{
			"http://localhost:5173",
			"http://localhost:5174",
			"http://localhost:3000",
		}),
		AllowedMethods: getEnvAsSlice("CORS_ALLOWED_METHODS", []string{
			"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH",
		}),
		AllowedHeaders: getEnvAsSlice("CORS_ALLOWED_HEADERS", []string{
			"Accept", "Authorization", "Content-Type", "X-Tenant-ID",
		}),
		ExposedHeaders: getEnvAsSlice("CORS_EXPOSED_HEADERS", []string{
			"Link",
		}),
		AllowCredentials: getEnvAsBool("CORS_ALLOW_CREDENTIALS", true),
		MaxAge:           getEnvAsInt("CORS_MAX_AGE", 300),
	}
}

func loadSecurityHeaderConfig() SecurityHeaderConfig {
	env := getEnv("APP_ENV", "development")

	// Default values based on environment
	var defaultEnableHSTS bool
	var defaultHSTSMaxAge int
	var defaultCSPDirectives string
	var defaultFrameOptions string

	if env == "production" {
		defaultEnableHSTS = true
		defaultHSTSMaxAge = 63072000 // 2 years
		defaultCSPDirectives = "default-src 'self'; script-src 'self'; style-src 'self'; connect-src 'self' wss:"
		defaultFrameOptions = "DENY"
	} else {
		defaultEnableHSTS = false // Disable HSTS in development
		defaultHSTSMaxAge = 31536000
		defaultCSPDirectives = "default-src 'self' 'unsafe-inline' 'unsafe-eval'; connect-src 'self' ws: wss:"
		defaultFrameOptions = "SAMEORIGIN"
	}

	return SecurityHeaderConfig{
		EnableHSTS:    getEnvAsBool("SECURITY_HEADER_ENABLE_HSTS", defaultEnableHSTS),
		HSTSMaxAge:    getEnvAsInt("SECURITY_HEADER_HSTS_MAX_AGE", defaultHSTSMaxAge),
		CSPDirectives: getEnv("SECURITY_HEADER_CSP_DIRECTIVES", defaultCSPDirectives),
		FrameOptions:  getEnv("SECURITY_HEADER_FRAME_OPTIONS", defaultFrameOptions),
	}
}
