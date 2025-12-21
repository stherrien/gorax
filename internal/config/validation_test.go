package config

import (
	"strings"
	"testing"
)

func TestValidateForProduction(t *testing.T) {
	tests := []struct {
		name        string
		config      *Config
		expectError bool
		errorMsg    string
	}{
		{
			name: "reject development environment",
			config: &Config{
				Server: ServerConfig{
					Env: "development",
				},
			},
			expectError: true,
			errorMsg:    "APP_ENV must be 'production' in production deployment",
		},
		{
			name: "reject default credential master key",
			config: &Config{
				Server: ServerConfig{
					Env: "production",
				},
				Credential: CredentialConfig{
					MasterKey: "dGhpcy1pcy1hLTMyLWJ5dGUtZGV2LWtleS0xMjM0NTY=", // Default dev key
				},
			},
			expectError: true,
			errorMsg:    "default development credential master key detected",
		},
		{
			name: "reject weak credential master key",
			config: &Config{
				Server: ServerConfig{
					Env: "production",
				},
				Database: DatabaseConfig{
					Host:     "db.production.example.com",
					Password: "secure-password-123",
					SSLMode:  "require",
				},
				Kratos: KratosConfig{
					PublicURL: "https://kratos.example.com",
					AdminURL:  "https://kratos-admin.example.com",
				},
				Credential: CredentialConfig{
					MasterKey: "changeme",
				},
			},
			expectError: true,
			errorMsg:    "credential master key is too short",
		},
		{
			name: "reject insecure DB password",
			config: &Config{
				Server: ServerConfig{
					Env: "production",
				},
				Database: DatabaseConfig{
					Password: "postgres",
				},
				Credential: CredentialConfig{
					MasterKey: "secure-random-key-32-bytes-long-abc123=",
				},
			},
			expectError: true,
			errorMsg:    "weak or default database password",
		},
		{
			name: "reject localhost in Kratos URL",
			config: &Config{
				Server: ServerConfig{
					Env: "production",
				},
				Database: DatabaseConfig{
					Host:     "db.production.example.com",
					Password: "secure-db-password-123",
					SSLMode:  "require",
				},
				Kratos: KratosConfig{
					PublicURL: "http://localhost:4433",
					AdminURL:  "http://localhost:4434",
				},
				Credential: CredentialConfig{
					MasterKey: "secure-random-key-32-bytes-long-abc123=",
				},
			},
			expectError: true,
			errorMsg:    "localhost URL detected in Kratos",
		},
		{
			name: "reject disabled SSL for database",
			config: &Config{
				Server: ServerConfig{
					Env: "production",
				},
				Database: DatabaseConfig{
					Password: "secure-db-password-123",
					SSLMode:  "disable",
				},
				Kratos: KratosConfig{
					PublicURL: "https://kratos.example.com",
					AdminURL:  "https://kratos-admin.example.com",
				},
				Credential: CredentialConfig{
					MasterKey: "secure-random-key-32-bytes-long-abc123=",
				},
			},
			expectError: true,
			errorMsg:    "database SSL must be enabled",
		},
		{
			name: "reject insecure HTTP in Kratos URLs",
			config: &Config{
				Server: ServerConfig{
					Env: "production",
				},
				Database: DatabaseConfig{
					Password: "secure-db-password-123",
					SSLMode:  "require",
				},
				Kratos: KratosConfig{
					PublicURL: "http://kratos.example.com",
					AdminURL:  "http://kratos-admin.example.com",
				},
				Credential: CredentialConfig{
					MasterKey: "secure-random-key-32-bytes-long-abc123=",
				},
			},
			expectError: true,
			errorMsg:    "insecure HTTP protocol",
		},
		{
			name: "warn on missing Sentry DSN when enabled",
			config: &Config{
				Server: ServerConfig{
					Env: "production",
				},
				Database: DatabaseConfig{
					Host:     "db.production.example.com",
					Password: "secure-db-password-123",
					SSLMode:  "require",
				},
				Kratos: KratosConfig{
					PublicURL: "https://kratos.example.com",
					AdminURL:  "https://kratos-admin.example.com",
				},
				Credential: CredentialConfig{
					MasterKey: "secure-random-key-32-bytes-long-abc123=",
				},
				Observability: ObservabilityConfig{
					SentryEnabled: true,
					SentryDSN:     "",
				},
			},
			expectError: false, // Warning, not error
		},
		{
			name: "valid production configuration",
			config: &Config{
				Server: ServerConfig{
					Env: "production",
				},
				Database: DatabaseConfig{
					Host:     "db.example.com",
					Password: "secure-db-password-123",
					SSLMode:  "require",
				},
				Redis: RedisConfig{
					Address:  "redis.example.com:6379",
					Password: "secure-redis-password",
				},
				Kratos: KratosConfig{
					PublicURL: "https://kratos.example.com",
					AdminURL:  "https://kratos-admin.example.com",
				},
				Credential: CredentialConfig{
					MasterKey: "secure-random-key-32-bytes-long-abc123=",
				},
				Observability: ObservabilityConfig{
					SentryEnabled:     true,
					SentryDSN:         "https://abc123@sentry.io/123456",
					SentryEnvironment: "production",
				},
			},
			expectError: false,
		},
		{
			name: "valid production with KMS",
			config: &Config{
				Server: ServerConfig{
					Env: "production",
				},
				Database: DatabaseConfig{
					Host:     "db.production.example.com",
					Password: "secure-db-password-123",
					SSLMode:  "verify-full",
				},
				Kratos: KratosConfig{
					PublicURL: "https://kratos.example.com",
					AdminURL:  "https://kratos-admin.example.com",
				},
				Credential: CredentialConfig{
					UseKMS:   true,
					KMSKeyID: "arn:aws:kms:us-east-1:123456789:key/abc-123",
				},
			},
			expectError: false,
		},
		{
			name: "reject empty SMTP password when email enabled",
			config: &Config{
				Server: ServerConfig{
					Env: "production",
				},
				Database: DatabaseConfig{
					Password: "secure-db-password-123",
					SSLMode:  "require",
				},
				Kratos: KratosConfig{
					PublicURL: "https://kratos.example.com",
					AdminURL:  "https://kratos-admin.example.com",
				},
				Credential: CredentialConfig{
					MasterKey: "secure-random-key-32-bytes-long-abc123=",
				},
				Notification: NotificationConfig{
					EnableEmail:   true,
					EmailProvider: "smtp",
					SMTPPass:      "",
				},
			},
			expectError: true,
			errorMsg:    "SMTP password must be configured",
		},
		{
			name: "reject default email sender",
			config: &Config{
				Server: ServerConfig{
					Env: "production",
				},
				Database: DatabaseConfig{
					Password: "secure-db-password-123",
					SSLMode:  "require",
				},
				Kratos: KratosConfig{
					PublicURL: "https://kratos.example.com",
					AdminURL:  "https://kratos-admin.example.com",
				},
				Credential: CredentialConfig{
					MasterKey: "secure-random-key-32-bytes-long-abc123=",
				},
				Notification: NotificationConfig{
					EnableEmail:   true,
					EmailProvider: "smtp",
					EmailFrom:     "noreply@example.com",
					SMTPPass:      "secure-smtp-password",
				},
			},
			expectError: true,
			errorMsg:    "default email sender address",
		},
		{
			name: "accept AWS SES without SMTP password",
			config: &Config{
				Server: ServerConfig{
					Env: "production",
				},
				Database: DatabaseConfig{
					Host:     "db.production.example.com",
					Password: "secure-db-password-123",
					SSLMode:  "require",
				},
				Kratos: KratosConfig{
					PublicURL: "https://kratos.example.com",
					AdminURL:  "https://kratos-admin.example.com",
				},
				Credential: CredentialConfig{
					MasterKey: "secure-random-key-32-bytes-long-abc123=",
				},
				Notification: NotificationConfig{
					EnableEmail:   true,
					EmailProvider: "ses",
					EmailFrom:     "noreply@mycompany.com",
					SESRegion:     "us-east-1",
				},
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateForProduction(tt.config)

			if tt.expectError {
				if err == nil {
					t.Error("expected error but got nil")
					return
				}
				if !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("expected error containing %q, got %q", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("expected no error but got: %v", err)
				}
			}
		})
	}
}

func TestIsWeakPassword(t *testing.T) {
	tests := []struct {
		name     string
		password string
		expected bool
	}{
		{"empty password", "", true},
		{"weak password - password", "password", true},
		{"weak password - secret", "secret", true},
		{"weak password - changeme", "changeme", true},
		{"weak password - admin", "admin", true},
		{"weak password - 123456", "123456", true},
		{"weak password - postgres", "postgres", true},
		{"weak password - root", "root", true},
		{"short password", "short", true},
		{"strong password", "secure-random-password-with-length-123", false},
		{"secure UUID-like", "550e8400-e29b-41d4-a716-446655440000", false},
		{"secure random string", "JKH87dfkjh234KJHdfkj", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isWeakPassword(tt.password)
			if result != tt.expected {
				t.Errorf("isWeakPassword(%q) = %v, expected %v", tt.password, result, tt.expected)
			}
		})
	}
}

func TestContainsLocalhostURL(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		expected bool
	}{
		{"localhost with http", "http://localhost:8080", true},
		{"localhost with https", "https://localhost:8443", true},
		{"localhost without port", "http://localhost", true},
		{"127.0.0.1 IPv4", "http://127.0.0.1:8080", true},
		{"0.0.0.0 IPv4", "http://0.0.0.0:8080", true},
		{"IPv6 localhost", "http://[::1]:8080", true},
		{"valid domain", "https://api.example.com", false},
		{"valid subdomain", "https://service.production.example.com", false},
		{"empty string", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := containsLocalhostURL(tt.url)
			if result != tt.expected {
				t.Errorf("containsLocalhostURL(%q) = %v, expected %v", tt.url, result, tt.expected)
			}
		})
	}
}
