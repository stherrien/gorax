package config

import (
	"fmt"
	"log/slog"
	"strings"
)

// Common weak/default passwords and secrets to check for
var weakPasswords = []string{
	"password",
	"secret",
	"changeme",
	"admin",
	"root",
	"postgres",
	"123456",
	"12345678",
	"qwerty",
	"abc123",
	"default",
	"guest",
}

// ValidateForProduction validates that configuration is suitable for production use.
// It checks for insecure settings, weak secrets, and development configurations
// that should never be used in production environments.
func ValidateForProduction(cfg *Config) error {
	var errors []string

	// Validate environment setting
	if err := validateEnvironment(cfg); err != nil {
		errors = append(errors, err.Error())
	}

	// Validate credential security
	if err := validateCredentials(cfg); err != nil {
		errors = append(errors, err.Error())
	}

	// Validate database security
	if err := validateDatabase(cfg); err != nil {
		errors = append(errors, err.Error())
	}

	// Validate external service URLs
	if err := validateServiceURLs(cfg); err != nil {
		errors = append(errors, err.Error())
	}

	// Validate notification configuration
	if err := validateNotifications(cfg); err != nil {
		errors = append(errors, err.Error())
	}

	// Log warnings for optional but recommended settings
	logProductionWarnings(cfg)

	if len(errors) > 0 {
		return fmt.Errorf("production configuration validation failed:\n  - %s", strings.Join(errors, "\n  - "))
	}

	slog.Info("production configuration validated successfully")
	return nil
}

func validateEnvironment(cfg *Config) error {
	if cfg.Server.Env != "production" {
		return fmt.Errorf("APP_ENV must be 'production' in production deployment, got: %s", cfg.Server.Env)
	}
	return nil
}

func validateCredentials(cfg *Config) error {
	// Check if using KMS (preferred for production)
	if cfg.Credential.UseKMS {
		if cfg.Credential.KMSKeyID == "" {
			return fmt.Errorf("KMS is enabled but KMSKeyID is not configured")
		}
		// KMS is being used, so MasterKey validation is not needed
		return nil
	}

	// If not using KMS, validate MasterKey
	if cfg.Credential.MasterKey == "" {
		return fmt.Errorf("credential master key must be configured when KMS is not used")
	}

	// Check for default development key
	if cfg.Credential.MasterKey == "dGhpcy1pcy1hLTMyLWJ5dGUtZGV2LWtleS0xMjM0NTY=" {
		return fmt.Errorf("default development credential master key detected - must use unique production key")
	}

	// Warn if master key is too short (should be at least 32 bytes base64 encoded = ~44 chars)
	if len(cfg.Credential.MasterKey) < 32 {
		return fmt.Errorf("credential master key is too short - minimum 32 characters required")
	}

	// Check for weak/insecure keys (only check common weak patterns)
	if isWeakPassword(cfg.Credential.MasterKey) {
		return fmt.Errorf("weak or insecure credential master key detected - must use strong random key")
	}

	return nil
}

func validateDatabase(cfg *Config) error {
	var errors []string

	// Check for weak database password
	if isWeakPassword(cfg.Database.Password) {
		errors = append(errors, "weak or default database password detected")
	}

	// Require SSL/TLS for database connections
	if cfg.Database.SSLMode == "disable" {
		errors = append(errors, "database SSL must be enabled in production (use 'require', 'verify-ca', or 'verify-full')")
	}

	// Check for localhost in database host (but allow valid hostnames)
	if cfg.Database.Host == "" || containsLocalhostURL(cfg.Database.Host) {
		errors = append(errors, "database host appears to be localhost or empty - use production database host")
	}

	if len(errors) > 0 {
		return fmt.Errorf("%s", strings.Join(errors, "; "))
	}

	return nil
}

func validateServiceURLs(cfg *Config) error {
	var errors []string

	// Check Kratos URLs
	if containsLocalhostURL(cfg.Kratos.PublicURL) {
		errors = append(errors, "localhost URL detected in Kratos PublicURL")
	}
	if containsLocalhostURL(cfg.Kratos.AdminURL) {
		errors = append(errors, "localhost URL detected in Kratos AdminURL")
	}

	// Require HTTPS for Kratos in production
	if !strings.HasPrefix(cfg.Kratos.PublicURL, "https://") {
		errors = append(errors, "insecure HTTP protocol in Kratos PublicURL - must use HTTPS in production")
	}
	if !strings.HasPrefix(cfg.Kratos.AdminURL, "https://") {
		errors = append(errors, "insecure HTTP protocol in Kratos AdminURL - must use HTTPS in production")
	}

	// Check Redis for localhost
	if containsLocalhostURL(cfg.Redis.Address) {
		errors = append(errors, "localhost detected in Redis address - use production Redis host")
	}

	// Check tracing endpoint if enabled
	if cfg.Observability.TracingEnabled && containsLocalhostURL(cfg.Observability.TracingEndpoint) {
		errors = append(errors, "localhost detected in tracing endpoint")
	}

	if len(errors) > 0 {
		return fmt.Errorf("%s", strings.Join(errors, "; "))
	}

	return nil
}

func validateNotifications(cfg *Config) error {
	if !cfg.Notification.EnableEmail {
		return nil
	}

	var errors []string

	// Check email configuration
	if cfg.Notification.EmailFrom == "noreply@example.com" {
		errors = append(errors, "default email sender address detected - must configure valid sender")
	}

	// Validate SMTP settings if using SMTP provider
	if cfg.Notification.EmailProvider == "smtp" {
		if cfg.Notification.SMTPHost == "" {
			errors = append(errors, "SMTP host must be configured when email is enabled")
		}
		if cfg.Notification.SMTPPass == "" {
			errors = append(errors, "SMTP password must be configured when using SMTP provider")
		}
		if isWeakPassword(cfg.Notification.SMTPPass) {
			errors = append(errors, "weak SMTP password detected")
		}
		if !cfg.Notification.SMTPTLS {
			errors = append(errors, "SMTP TLS must be enabled in production")
		}
	}

	// Validate Slack configuration if enabled
	if cfg.Notification.EnableSlack {
		if cfg.Notification.SlackWebhookURL == "" {
			errors = append(errors, "Slack webhook URL must be configured when Slack notifications are enabled")
		}
		if containsLocalhostURL(cfg.Notification.SlackWebhookURL) {
			errors = append(errors, "localhost detected in Slack webhook URL")
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("%s", strings.Join(errors, "; "))
	}

	return nil
}

func logProductionWarnings(cfg *Config) {
	// Check for missing but recommended settings
	if cfg.Observability.SentryEnabled && cfg.Observability.SentryDSN == "" {
		slog.Warn("Sentry error tracking is enabled but DSN is not configured")
	}

	if cfg.Observability.SentryEnabled && cfg.Observability.SentryEnvironment != "production" {
		slog.Warn("Sentry environment should be 'production'", "current", cfg.Observability.SentryEnvironment)
	}

	if !cfg.Observability.TracingEnabled {
		slog.Warn("distributed tracing is disabled - consider enabling for production observability")
	}

	if !cfg.Observability.MetricsEnabled {
		slog.Warn("metrics collection is disabled - consider enabling for production monitoring")
	}

	if cfg.Redis.Password == "" {
		slog.Warn("Redis password is not set - ensure Redis is secured by other means")
	}

	if !cfg.Retention.Enabled {
		slog.Warn("execution retention cleanup is disabled - database may grow indefinitely")
	}

	if !cfg.Cleanup.Enabled {
		slog.Warn("webhook event cleanup is disabled - database may grow indefinitely")
	}
}

// isWeakPassword checks if a password matches common weak passwords or patterns
func isWeakPassword(password string) bool {
	if password == "" {
		return true
	}

	// Check length
	if len(password) < 8 {
		return true
	}

	// Check against common weak passwords (exact match or if the password IS the weak word)
	lowerPassword := strings.ToLower(password)
	for _, weak := range weakPasswords {
		if lowerPassword == weak {
			return true
		}
	}

	return false
}

// containsLocalhostURL checks if a URL or host string contains localhost references
func containsLocalhostURL(url string) bool {
	if url == "" {
		return false
	}

	lowerURL := strings.ToLower(url)

	// Check for localhost
	if strings.Contains(lowerURL, "localhost") {
		return true
	}

	// Check for IPv4 loopback
	if strings.Contains(lowerURL, "127.0.0.1") || strings.Contains(lowerURL, "0.0.0.0") {
		return true
	}

	// Check for IPv6 loopback
	if strings.Contains(lowerURL, "::1") || strings.Contains(lowerURL, "[::1]") {
		return true
	}

	return false
}
