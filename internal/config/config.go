package config

import (
	"fmt"
	"os"
	"strconv"
)

// Config holds all application configuration
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Redis    RedisConfig
	Kratos   KratosConfig
	Worker   WorkerConfig
	AWS      AWSConfig
	Queue    QueueConfig
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
