package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/gorax/gorax/docs/api" // Import generated Swagger docs
	"github.com/gorax/gorax/internal/api"
	"github.com/gorax/gorax/internal/config"
	"github.com/gorax/gorax/internal/tracing"
)

// @title Gorax Workflow Automation API
// @version 1.0
// @description REST API for Gorax - a powerful workflow automation platform with webhooks, scheduling, and real-time execution monitoring.
// @description
// @description ## Authentication
// @description All API endpoints (except /health, /ready, and webhook triggers) require authentication.
// @description In development mode, use the X-User-ID header. In production, use Ory Kratos session cookies.
// @description
// @description ## Multi-tenancy
// @description All authenticated endpoints require a valid X-Tenant-ID header to identify the tenant context.
// @description
// @description ## Rate Limiting
// @description API requests are rate-limited based on tenant quotas. Check response headers for limit information.
// @description
// @description ## Error Handling
// @description Errors follow a consistent format with appropriate HTTP status codes and error messages.

// @contact.name Gorax Support
// @contact.url https://github.com/gorax/gorax
// @contact.email support@gorax.io

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8080
// @BasePath /api/v1

// @securityDefinitions.apikey TenantID
// @in header
// @name X-Tenant-ID
// @description Tenant identifier for multi-tenant isolation

// @securityDefinitions.apikey UserID
// @in header
// @name X-User-ID
// @description User identifier (development mode only)

// @securityDefinitions.apikey SessionCookie
// @in cookie
// @name ory_kratos_session
// @description Ory Kratos session cookie (production mode)

// @tag.name Health
// @tag.description Health check and readiness endpoints

// @tag.name Workflows
// @tag.description Workflow management and execution

// @tag.name Webhooks
// @tag.description Webhook configuration and management

// @tag.name Executions
// @tag.description Workflow execution history and monitoring

// @tag.name Schedules
// @tag.description Scheduled workflow triggers

// @tag.name Credentials
// @tag.description Secure credential management

// @tag.name Tenants
// @tag.description Tenant administration (admin only)

// @tag.name Metrics
// @tag.description Execution metrics and analytics

// @tag.name WebSocket
// @tag.description Real-time execution updates

// @tag.name Event Types
// @tag.description Event type registry for webhooks

// @tag.name Analytics
// @tag.description Workflow execution analytics and reporting

// @tag.name Marketplace
// @tag.description Template marketplace for sharing and installing workflows

// @tag.name RBAC
// @tag.description Role-based access control and permissions management

func main() {
	// Load configuration first (we need it to configure logging)
	cfg, err := config.Load()
	if err != nil {
		// Use default logger for startup errors
		slog.Error("failed to load configuration", "error", err)
		os.Exit(1)
	}

	// Parse log level from configuration
	logLevel := parseLogLevel(cfg.Log.Level)

	// Initialize structured logger with configured level
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: logLevel,
	}))
	slog.SetDefault(logger)

	// Validate production configuration
	// This prevents the application from starting with insecure development settings
	// in production environments. Checks for weak passwords, localhost URLs, disabled SSL, etc.
	if cfg.Server.Env == "production" {
		if err := config.ValidateForProduction(cfg); err != nil {
			slog.Error("production configuration validation failed", "error", err)
			os.Exit(1)
		}
	}

	// Initialize tracing
	tracingCleanup, err := tracing.InitGlobalTracer(context.Background(), &cfg.Observability)
	if err != nil {
		slog.Error("failed to initialize tracing", "error", err)
		os.Exit(1)
	}
	defer tracingCleanup()

	if cfg.Observability.TracingEnabled {
		slog.Info("distributed tracing enabled",
			"endpoint", cfg.Observability.TracingEndpoint,
			"service_name", cfg.Observability.TracingServiceName,
			"sample_rate", cfg.Observability.TracingSampleRate,
		)
	}

	// Initialize application
	app, err := api.NewApp(cfg, logger)
	if err != nil {
		slog.Error("failed to initialize application", "error", err)
		os.Exit(1)
	}
	defer app.Close()

	// Create HTTP server
	server := &http.Server{
		Addr:         cfg.Server.Address,
		Handler:      app.Router(),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine
	go func() {
		slog.Info("starting API server", "address", cfg.Server.Address)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("shutting down server...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		slog.Error("server forced to shutdown", "error", err)
	}

	slog.Info("server stopped")
}

// parseLogLevel converts string log level to slog.Level
func parseLogLevel(level string) slog.Level {
	switch level {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		// Default to info if invalid level specified
		return slog.LevelInfo
	}
}
