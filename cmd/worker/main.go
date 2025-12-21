package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"

	"github.com/gorax/gorax/internal/config"
	"github.com/gorax/gorax/internal/schedule"
	"github.com/gorax/gorax/internal/tracing"
	"github.com/gorax/gorax/internal/webhook"
	"github.com/gorax/gorax/internal/worker"
	"github.com/gorax/gorax/internal/workflow"
)

func main() {
	// Initialize structured logger
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load configuration", "error", err)
		os.Exit(1)
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

	// Create cancellable context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize database connection for scheduler
	db, err := sqlx.Connect("postgres", cfg.Database.ConnectionString())
	if err != nil {
		slog.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	// Initialize repositories
	workflowRepo := workflow.NewRepository(db)
	scheduleRepo := schedule.NewRepository(db)

	// Initialize workflow service
	workflowService := workflow.NewService(workflowRepo, logger)

	// Initialize schedule service
	scheduleService := schedule.NewService(scheduleRepo, logger)

	// Create workflow getter adapter
	workflowGetter := &workflowServiceAdapter{workflowService: workflowService}
	scheduleService.SetWorkflowService(workflowGetter)

	// Create workflow executor adapter for scheduler
	executorAdapter := schedule.NewWorkflowServiceAdapter(func(ctx context.Context, tenantID, workflowID, triggerType string, triggerData []byte) (string, error) {
		execution, err := workflowService.Execute(ctx, tenantID, workflowID, triggerType, triggerData)
		if err != nil {
			return "", err
		}
		return execution.ID, nil
	})

	// Initialize scheduler
	scheduler := schedule.NewScheduler(scheduleService, executorAdapter, logger)

	// Initialize webhook cleanup if enabled
	var cleanupScheduler *webhook.CleanupScheduler
	if cfg.Cleanup.Enabled {
		webhookRepo := webhook.NewRepository(db)
		retentionPeriod := time.Duration(cfg.Cleanup.RetentionDays) * 24 * time.Hour
		cleanupService := webhook.NewCleanupService(webhookRepo, cfg.Cleanup.BatchSize, retentionPeriod)
		cleanupScheduler = webhook.NewCleanupScheduler(cleanupService, cfg.Cleanup.Schedule, logger)
	}

	// Initialize worker
	w, err := worker.New(cfg, logger)
	if err != nil {
		slog.Error("failed to initialize worker", "error", err)
		os.Exit(1)
	}
	defer w.Close()

	// Start health check server
	healthServer := worker.NewHealthServer(w, cfg.Worker.HealthPort)
	go func() {
		if err := healthServer.Start(); err != nil && err != http.ErrServerClosed {
			slog.Error("health server error", "error", err)
		}
	}()
	defer func() {
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer shutdownCancel()
		healthServer.Shutdown(shutdownCtx)
	}()

	// Start scheduler in goroutine
	go func() {
		slog.Info("starting workflow scheduler")
		if err := scheduler.Start(ctx); err != nil {
			slog.Error("scheduler error", "error", err)
		}
	}()

	// Start cleanup scheduler if enabled
	if cleanupScheduler != nil {
		go func() {
			slog.Info("starting cleanup scheduler",
				"retention_days", cfg.Cleanup.RetentionDays,
				"batch_size", cfg.Cleanup.BatchSize,
				"schedule", cfg.Cleanup.Schedule,
			)
			if err := cleanupScheduler.Start(ctx); err != nil {
				slog.Error("cleanup scheduler error", "error", err)
			}
		}()
	}

	// Start worker in goroutine
	go func() {
		slog.Info("starting workflow worker", "concurrency", cfg.Worker.Concurrency)
		if err := w.Start(ctx); err != nil {
			slog.Error("worker error", "error", err)
			cancel()
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("shutting down worker, scheduler, and cleanup scheduler...")
	cancel()

	// Stop scheduler
	scheduler.Stop()

	// Stop cleanup scheduler if enabled
	if cleanupScheduler != nil {
		cleanupScheduler.Stop()
	}

	// Wait for worker to finish current jobs
	w.Wait()

	slog.Info("worker, scheduler, and cleanup scheduler stopped")
}

// workflowServiceAdapter adapts workflow.Service to schedule.WorkflowGetter interface
type workflowServiceAdapter struct {
	workflowService *workflow.Service
}

func (w *workflowServiceAdapter) GetByID(ctx context.Context, tenantID, id string) (interface{}, error) {
	return w.workflowService.GetByID(ctx, tenantID, id)
}
