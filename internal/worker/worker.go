package worker

import (
	"context"
	"log/slog"
	"sync"
	"sync/atomic"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/redis/go-redis/v9"

	"github.com/gorax/gorax/internal/config"
	"github.com/gorax/gorax/internal/executor"
	"github.com/gorax/gorax/internal/queue"
	"github.com/gorax/gorax/internal/workflow"
)

// Worker processes workflow executions
type Worker struct {
	config   *config.Config
	logger   *slog.Logger
	db       *sqlx.DB
	redis    *redis.Client
	executor *executor.Executor
	workflowRepo *workflow.Repository

	// Queue-based processing
	queueConsumer  *queue.Consumer
	queueEnabled   bool

	concurrency      int
	concurrencyLimit *TenantConcurrencyLimiter
	wg               sync.WaitGroup

	// Metrics
	activeExecutions atomic.Int32
	processedTotal   atomic.Int64
	failedTotal      atomic.Int64
}

// New creates a new worker instance
func New(cfg *config.Config, logger *slog.Logger) (*Worker, error) {
	// Initialize database connection
	db, err := sqlx.Connect("postgres", cfg.Database.ConnectionString())
	if err != nil {
		return nil, err
	}

	// Initialize Redis client
	redisClient := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Address,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})

	// Initialize workflow repository
	workflowRepo := workflow.NewRepository(db)

	// Initialize executor
	exec := executor.New(workflowRepo, logger)

	// Initialize tenant concurrency limiter
	// Default to 10 concurrent executions per tenant if not configured
	maxPerTenant := 10
	if cfg.Worker.MaxConcurrencyPerTenant > 0 {
		maxPerTenant = cfg.Worker.MaxConcurrencyPerTenant
	}
	concurrencyLimit := NewTenantConcurrencyLimiter(redisClient, maxPerTenant)

	w := &Worker{
		config:           cfg,
		logger:           logger,
		db:               db,
		redis:            redisClient,
		executor:         exec,
		workflowRepo:     workflowRepo,
		concurrency:      cfg.Worker.Concurrency,
		concurrencyLimit: concurrencyLimit,
		queueEnabled:     cfg.Queue.Enabled,
	}

	// Initialize queue consumer if enabled
	if cfg.Queue.Enabled {
		if cfg.AWS.SQSQueueURL == "" {
			return nil, ErrMissingQueueURL
		}

		// Create SQS client
		sqsClient, err := queue.NewSQSClient(context.Background(), queue.SQSConfig{
			QueueURL:        cfg.AWS.SQSQueueURL,
			DLQueueURL:      cfg.AWS.SQSDLQueueURL,
			Region:          cfg.AWS.Region,
			AccessKeyID:     cfg.AWS.AccessKeyID,
			SecretAccessKey: cfg.AWS.SecretAccessKey,
			Endpoint:        cfg.AWS.Endpoint,
		}, logger)
		if err != nil {
			return nil, err
		}

		// Create message handler
		handler := func(ctx context.Context, msg *queue.ExecutionMessage) error {
			return w.processExecutionMessage(ctx, msg)
		}

		// Create consumer config
		consumerConfig := queue.ConsumerConfig{
			MaxMessages:        cfg.Queue.MaxMessages,
			WaitTimeSeconds:    cfg.Queue.WaitTimeSeconds,
			VisibilityTimeout:  cfg.Queue.VisibilityTimeout,
			MaxRetries:         cfg.Queue.MaxRetries,
			ProcessTimeout:     time.Duration(cfg.Queue.ProcessTimeout) * time.Second,
			PollInterval:       time.Duration(cfg.Queue.PollInterval) * time.Second,
			ConcurrentWorkers:  cfg.Queue.ConcurrentWorkers,
			DeleteAfterProcess: cfg.Queue.DeleteAfterProcess,
		}

		// Create consumer
		w.queueConsumer = queue.NewConsumer(sqsClient, handler, consumerConfig, logger)
		logger.Info("queue consumer initialized", "queue_url", cfg.AWS.SQSQueueURL)
	}

	return w, nil
}

// Start begins processing jobs
func (w *Worker) Start(ctx context.Context) error {
	if w.queueEnabled && w.queueConsumer != nil {
		// Use queue-based processing
		w.logger.Info("starting queue-based worker", "queue_enabled", true)
		return w.queueConsumer.Start(ctx)
	}

	// Fallback to polling-based processing (backward compatibility)
	w.logger.Info("starting worker pool", "concurrency", w.concurrency, "queue_enabled", false)

	// Start worker goroutines
	for i := 0; i < w.concurrency; i++ {
		w.wg.Add(1)
		go w.processLoop(ctx, i)
	}

	// Wait for context cancellation
	<-ctx.Done()
	return ctx.Err()
}

// processLoop is the main processing loop for a worker
func (w *Worker) processLoop(ctx context.Context, workerID int) {
	defer w.wg.Done()

	w.logger.Info("worker started", "worker_id", workerID)

	for {
		select {
		case <-ctx.Done():
			w.logger.Info("worker stopping", "worker_id", workerID)
			return
		default:
			// Poll for pending executions
			execution, err := w.pollExecution(ctx)
			if err != nil {
				// No work available, wait a bit
				continue
			}

			// Process the execution
			if err := w.processExecution(ctx, execution); err != nil {
				w.logger.Error("execution failed", "error", err, "execution_id", execution.ID)
			}
		}
	}
}

// pollExecution polls for pending executions
func (w *Worker) pollExecution(ctx context.Context) (*workflow.Execution, error) {
	// TODO: Implement polling from queue (SQS) or database
	// For now, this is a placeholder that returns an error to indicate no work

	// In production, this would:
	// 1. Receive message from SQS queue
	// 2. Parse execution ID from message
	// 3. Load execution from database
	// 4. Return execution for processing

	return nil, ErrNoWork
}

// processExecution processes a single execution
func (w *Worker) processExecution(ctx context.Context, execution *workflow.Execution) error {
	w.logger.Info("processing execution", "execution_id", execution.ID, "workflow_id", execution.WorkflowID, "tenant_id", execution.TenantID)

	// Try to acquire tenant concurrency slot
	acquired, err := w.concurrencyLimit.Acquire(ctx, execution.TenantID, execution.ID)
	if err != nil {
		w.logger.Error("failed to acquire tenant concurrency slot", "error", err, "tenant_id", execution.TenantID)
		return err
	}

	if !acquired {
		w.logger.Warn("tenant at concurrency limit, requeueing execution",
			"tenant_id", execution.TenantID,
			"execution_id", execution.ID,
			"max_concurrent", w.concurrencyLimit.GetMaxPerTenant(),
		)
		// TODO: Requeue the message with delay
		return ErrTenantAtCapacity
	}

	// Release the slot when done
	defer func() {
		if err := w.concurrencyLimit.Release(ctx, execution.TenantID, execution.ID); err != nil {
			w.logger.Error("failed to release tenant concurrency slot", "error", err, "tenant_id", execution.TenantID)
		}
	}()

	// Track active executions
	w.activeExecutions.Add(1)
	defer w.activeExecutions.Add(-1)

	// Execute the workflow
	err = w.executor.Execute(ctx, execution)
	if err != nil {
		w.failedTotal.Add(1)
		return err
	}

	w.logger.Info("execution completed", "execution_id", execution.ID)
	w.processedTotal.Add(1)
	return nil
}

// processExecutionMessage processes an execution message from the queue
func (w *Worker) processExecutionMessage(ctx context.Context, msg *queue.ExecutionMessage) error {
	w.logger.Info("processing execution message",
		"execution_id", msg.ExecutionID,
		"workflow_id", msg.WorkflowID,
		"tenant_id", msg.TenantID,
	)

	// Load execution from database
	execution, err := w.workflowRepo.GetExecutionByID(ctx, msg.TenantID, msg.ExecutionID)
	if err != nil {
		w.logger.Error("failed to load execution",
			"error", err,
			"execution_id", msg.ExecutionID,
		)
		return err
	}

	// Process the execution
	if err := w.processExecution(ctx, execution); err != nil {
		w.logger.Error("execution processing failed",
			"error", err,
			"execution_id", msg.ExecutionID,
		)
		return err
	}

	return nil
}

// Wait waits for all workers to finish
func (w *Worker) Wait() {
	w.wg.Wait()
}

// Close cleans up worker resources
func (w *Worker) Close() error {
	if w.db != nil {
		w.db.Close()
	}
	if w.redis != nil {
		w.redis.Close()
	}
	return nil
}

// getActiveExecutions returns the current number of active executions
func (w *Worker) getActiveExecutions() int32 {
	return w.activeExecutions.Load()
}

// getProcessedCount returns the total number of processed executions
func (w *Worker) getProcessedCount() int64 {
	return w.processedTotal.Load()
}

// getFailedCount returns the total number of failed executions
func (w *Worker) getFailedCount() int64 {
	return w.failedTotal.Load()
}

// Custom errors
type WorkerError struct {
	Message string
}

func (e WorkerError) Error() string {
	return e.Message
}

var (
	ErrNoWork           = WorkerError{Message: "no work available"}
	ErrTenantAtCapacity = WorkerError{Message: "tenant at concurrency capacity"}
	ErrMissingQueueURL  = WorkerError{Message: "queue URL is required when queue is enabled"}
)
