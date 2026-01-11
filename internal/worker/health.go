package worker

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/gorax/gorax/internal/buildinfo"
)

// HealthServer provides health check endpoints for the worker
type HealthServer struct {
	worker *Worker
	server *http.Server
	ready  atomic.Bool
}

// HealthResponse represents the health check response
type HealthResponse struct {
	Status      string            `json:"status"`
	Timestamp   time.Time         `json:"timestamp"`
	Version     string            `json:"version"`
	WorkerInfo  WorkerInfo        `json:"worker_info"`
	Connections ConnectionsHealth `json:"connections"`
}

// WorkerInfo contains worker statistics
type WorkerInfo struct {
	Concurrency      int   `json:"concurrency"`
	ActiveExecutions int32 `json:"active_executions"`
	ProcessedTotal   int64 `json:"processed_total"`
	FailedTotal      int64 `json:"failed_total"`
}

// ConnectionsHealth contains connection status
type ConnectionsHealth struct {
	Database string `json:"database"`
	Redis    string `json:"redis"`
	Queue    string `json:"queue"`
}

// NewHealthServer creates a new health check server
func NewHealthServer(worker *Worker, port string) *HealthServer {
	hs := &HealthServer{
		worker: worker,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/health/live", hs.handleLiveness)
	mux.HandleFunc("/health/ready", hs.handleReadiness)
	mux.HandleFunc("/health", hs.handleHealth)

	hs.server = &http.Server{
		Addr:         ":" + port,
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	return hs
}

// Start starts the health check server
func (hs *HealthServer) Start() error {
	hs.worker.logger.Info("starting health check server", "port", hs.server.Addr)
	hs.ready.Store(true)
	return hs.server.ListenAndServe()
}

// Shutdown gracefully shuts down the health server
func (hs *HealthServer) Shutdown(ctx context.Context) error {
	hs.ready.Store(false)
	return hs.server.Shutdown(ctx)
}

// SetReady sets the ready state
func (hs *HealthServer) SetReady(ready bool) {
	hs.ready.Store(ready)
}

// handleLiveness handles Kubernetes liveness probe
// Returns 200 if worker process is alive
func (hs *HealthServer) handleLiveness(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(map[string]string{
		"status": "alive",
		"time":   time.Now().Format(time.RFC3339),
	}); err != nil {
		slog.Error("failed to encode liveness response", "error", err)
	}
}

// handleReadiness handles Kubernetes readiness probe
// Returns 200 if worker is ready to process work
func (hs *HealthServer) handleReadiness(w http.ResponseWriter, r *http.Request) {
	if !hs.ready.Load() {
		w.WriteHeader(http.StatusServiceUnavailable)
		if err := json.NewEncoder(w).Encode(map[string]string{
			"status": "not_ready",
			"time":   time.Now().Format(time.RFC3339),
		}); err != nil {
			slog.Error("failed to encode readiness not_ready response", "error", err)
		}
		return
	}

	// Check if worker can accept more work
	// Convert concurrency to int32 safely (concurrency is always a small positive config value)
	concurrencyInt32 := safeIntToInt32(hs.worker.concurrency)
	if hs.worker.getActiveExecutions() >= concurrencyInt32 {
		w.WriteHeader(http.StatusServiceUnavailable)
		if err := json.NewEncoder(w).Encode(map[string]string{
			"status": "at_capacity",
			"time":   time.Now().Format(time.RFC3339),
		}); err != nil {
			slog.Error("failed to encode readiness at_capacity response", "error", err)
		}
		return
	}

	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(map[string]string{
		"status": "ready",
		"time":   time.Now().Format(time.RFC3339),
	}); err != nil {
		slog.Error("failed to encode readiness ready response", "error", err)
	}
}

// handleHealth provides detailed health information
func (hs *HealthServer) handleHealth(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	response := HealthResponse{
		Status:    "healthy",
		Timestamp: time.Now(),
		Version:   buildinfo.GetVersion(),
		WorkerInfo: WorkerInfo{
			Concurrency:      hs.worker.concurrency,
			ActiveExecutions: hs.worker.getActiveExecutions(),
			ProcessedTotal:   hs.worker.getProcessedCount(),
			FailedTotal:      hs.worker.getFailedCount(),
		},
		Connections: ConnectionsHealth{
			Database: hs.checkDatabase(ctx),
			Redis:    hs.checkRedis(ctx),
			Queue:    hs.checkQueue(ctx),
		},
	}

	// If any connection is unhealthy, set overall status to unhealthy
	if response.Connections.Database != "ok" || response.Connections.Redis != "ok" || response.Connections.Queue != "ok" {
		response.Status = "unhealthy"
		w.WriteHeader(http.StatusServiceUnavailable)
	} else {
		w.WriteHeader(http.StatusOK)
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		slog.Error("failed to encode health response", "error", err)
	}
}

// checkDatabase checks database connectivity
func (hs *HealthServer) checkDatabase(ctx context.Context) string {
	if err := hs.worker.db.PingContext(ctx); err != nil {
		return "error: " + err.Error()
	}
	return "ok"
}

// checkRedis checks Redis connectivity
func (hs *HealthServer) checkRedis(ctx context.Context) string {
	if err := hs.worker.redis.Ping(ctx).Err(); err != nil {
		return "error: " + err.Error()
	}
	return "ok"
}

// checkQueue checks queue connectivity and health
func (hs *HealthServer) checkQueue(ctx context.Context) string {
	// If queue is not enabled, report as ok (not applicable)
	if !hs.worker.queueEnabled {
		return "ok"
	}

	// If queue is enabled but client is nil, report error
	if hs.worker.sqsClient == nil {
		return "error: queue client not initialized"
	}

	// Check SQS queue health by attempting to get queue attributes
	if err := hs.worker.sqsClient.HealthCheck(ctx); err != nil {
		return "error: " + err.Error()
	}

	return "ok"
}

// safeIntToInt32 safely converts int to int32 with bounds checking
// Returns 0 for negative values, maxInt32 for values exceeding int32 range
func safeIntToInt32(val int) int32 {
	const maxInt32 = 1<<31 - 1 // 2147483647
	if val < 0 {
		return 0
	}
	if val > maxInt32 {
		return maxInt32
	}
	return int32(val)
}
