package worker

import (
	"context"
	"encoding/json"
	"net/http"
	"sync/atomic"
	"time"
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
	Concurrency       int   `json:"concurrency"`
	ActiveExecutions  int32 `json:"active_executions"`
	ProcessedTotal    int64 `json:"processed_total"`
	FailedTotal       int64 `json:"failed_total"`
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
	json.NewEncoder(w).Encode(map[string]string{
		"status": "alive",
		"time":   time.Now().Format(time.RFC3339),
	})
}

// handleReadiness handles Kubernetes readiness probe
// Returns 200 if worker is ready to process work
func (hs *HealthServer) handleReadiness(w http.ResponseWriter, r *http.Request) {
	if !hs.ready.Load() {
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(map[string]string{
			"status": "not_ready",
			"time":   time.Now().Format(time.RFC3339),
		})
		return
	}

	// Check if worker can accept more work
	if hs.worker.getActiveExecutions() >= int32(hs.worker.concurrency) {
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(map[string]string{
			"status": "at_capacity",
			"time":   time.Now().Format(time.RFC3339),
		})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status": "ready",
		"time":   time.Now().Format(time.RFC3339),
	})
}

// handleHealth provides detailed health information
func (hs *HealthServer) handleHealth(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	response := HealthResponse{
		Status:    "healthy",
		Timestamp: time.Now(),
		Version:   "1.0.0", // TODO: Get from build info
		WorkerInfo: WorkerInfo{
			Concurrency:      hs.worker.concurrency,
			ActiveExecutions: hs.worker.getActiveExecutions(),
			ProcessedTotal:   hs.worker.getProcessedCount(),
			FailedTotal:      hs.worker.getFailedCount(),
		},
		Connections: ConnectionsHealth{
			Database: hs.checkDatabase(ctx),
			Redis:    hs.checkRedis(ctx),
			Queue:    "ok", // TODO: Check queue connection
		},
	}

	// If any connection is unhealthy, set overall status to unhealthy
	if response.Connections.Database != "ok" || response.Connections.Redis != "ok" {
		response.Status = "unhealthy"
		w.WriteHeader(http.StatusServiceUnavailable)
	} else {
		w.WriteHeader(http.StatusOK)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
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
