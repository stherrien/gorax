package javascript

import (
	"context"
	"log/slog"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	"github.com/gorax/gorax/internal/tracing"
)

// ExecutionLogger handles logging and telemetry for JavaScript execution.
type ExecutionLogger struct {
	logger *slog.Logger
}

// NewExecutionLogger creates a new execution logger.
func NewExecutionLogger(logger *slog.Logger) *ExecutionLogger {
	if logger == nil {
		logger = slog.Default()
	}
	return &ExecutionLogger{logger: logger}
}

// ExecutionLogEntry represents a log entry for script execution.
type ExecutionLogEntry struct {
	// ExecutionID uniquely identifies this execution.
	ExecutionID string `json:"execution_id"`

	// TenantID identifies the tenant.
	TenantID string `json:"tenant_id,omitempty"`

	// WorkflowID identifies the workflow.
	WorkflowID string `json:"workflow_id,omitempty"`

	// NodeID identifies the workflow node.
	NodeID string `json:"node_id,omitempty"`

	// ScriptHash is a hash of the script for identification.
	ScriptHash string `json:"script_hash,omitempty"`

	// StartTime is when execution started.
	StartTime time.Time `json:"start_time"`

	// Duration is how long execution took.
	Duration time.Duration `json:"duration"`

	// Success indicates if execution completed successfully.
	Success bool `json:"success"`

	// Error contains error information if execution failed.
	Error string `json:"error,omitempty"`

	// ErrorPhase indicates which phase the error occurred in.
	ErrorPhase ExecutionPhase `json:"error_phase,omitempty"`

	// ConsoleLogs contains captured console output.
	ConsoleLogs []ConsoleEntry `json:"console_logs,omitempty"`

	// MemoryDelta is the memory change during execution.
	MemoryDelta int64 `json:"memory_delta_bytes,omitempty"`

	// Metadata contains additional context.
	Metadata map[string]any `json:"metadata,omitempty"`
}

// LogExecutionStart logs the start of a script execution.
func (el *ExecutionLogger) LogExecutionStart(entry *ExecutionLogEntry) {
	el.logger.Info("javascript execution started",
		"execution_id", entry.ExecutionID,
		"tenant_id", entry.TenantID,
		"workflow_id", entry.WorkflowID,
		"node_id", entry.NodeID,
		"script_hash", entry.ScriptHash,
	)
}

// LogExecutionComplete logs the completion of a script execution.
func (el *ExecutionLogger) LogExecutionComplete(entry *ExecutionLogEntry) {
	level := slog.LevelInfo
	if !entry.Success {
		level = slog.LevelError
	}

	attrs := []any{
		"execution_id", entry.ExecutionID,
		"tenant_id", entry.TenantID,
		"workflow_id", entry.WorkflowID,
		"node_id", entry.NodeID,
		"duration_ms", entry.Duration.Milliseconds(),
		"success", entry.Success,
		"console_log_count", len(entry.ConsoleLogs),
		"memory_delta_bytes", entry.MemoryDelta,
	}

	if !entry.Success && entry.Error != "" {
		attrs = append(attrs, "error", entry.Error)
		attrs = append(attrs, "error_phase", entry.ErrorPhase)
	}

	el.logger.Log(context.Background(), level, "javascript execution completed", attrs...)
}

// LogSecurityEvent logs a security-related event.
func (el *ExecutionLogger) LogSecurityEvent(
	executionID string,
	eventType string,
	details map[string]any,
) {
	el.logger.Warn("javascript security event",
		"execution_id", executionID,
		"event_type", eventType,
		"details", details,
	)
}

// LogResourceLimitReached logs when a resource limit is hit.
func (el *ExecutionLogger) LogResourceLimitReached(
	executionID string,
	limitType string,
	limitValue any,
	actualValue any,
) {
	el.logger.Warn("javascript resource limit reached",
		"execution_id", executionID,
		"limit_type", limitType,
		"limit_value", limitValue,
		"actual_value", actualValue,
	)
}

// ExecutionTracer handles OpenTelemetry tracing for JavaScript execution.
type ExecutionTracer struct{}

// NewExecutionTracer creates a new execution tracer.
func NewExecutionTracer() *ExecutionTracer {
	return &ExecutionTracer{}
}

// TraceExecution wraps a JavaScript execution with tracing.
func (et *ExecutionTracer) TraceExecution(
	ctx context.Context,
	executionID, tenantID, workflowID, nodeID string,
	fn func(context.Context) (any, error),
) (any, error) {
	ctx, span := tracing.StartSpan(ctx, "javascript.execute")
	defer span.End()

	span.SetAttributes(
		attribute.String("execution_id", executionID),
		attribute.String("tenant_id", tenantID),
		attribute.String("workflow_id", workflowID),
		attribute.String("node_id", nodeID),
		attribute.String("component", "javascript_executor"),
	)

	result, err := fn(ctx)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	span.SetStatus(codes.Ok, "script execution completed")
	return result, nil
}

// TraceCompilation wraps script compilation with tracing.
func (et *ExecutionTracer) TraceCompilation(
	ctx context.Context,
	scriptLength int,
	fn func(context.Context) error,
) error {
	ctx, span := tracing.StartSpan(ctx, "javascript.compile")
	defer span.End()

	span.SetAttributes(
		attribute.Int("script_length", scriptLength),
		attribute.String("component", "javascript_executor"),
	)

	err := fn(ctx)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}

	span.SetStatus(codes.Ok, "compilation completed")
	return nil
}

// RecordExecutionMetrics records metrics to the current span.
func (et *ExecutionTracer) RecordExecutionMetrics(
	ctx context.Context,
	duration time.Duration,
	memoryDelta int64,
	consoleLogCount int,
) {
	span := trace.SpanFromContext(ctx)
	if !span.SpanContext().IsValid() {
		return
	}

	span.SetAttributes(
		attribute.Int64("execution_duration_ms", duration.Milliseconds()),
		attribute.Int64("memory_delta_bytes", memoryDelta),
		attribute.Int("console_log_count", consoleLogCount),
	)
}

// RecordError records an error to the current span.
func (et *ExecutionTracer) RecordError(ctx context.Context, err error, phase ExecutionPhase) {
	span := trace.SpanFromContext(ctx)
	if !span.SpanContext().IsValid() {
		return
	}

	span.RecordError(err)
	span.SetAttributes(
		attribute.String("error_phase", string(phase)),
	)
}

// AuditLogger handles audit logging for JavaScript execution.
type AuditLogger struct {
	logger *slog.Logger
}

// NewAuditLogger creates a new audit logger.
func NewAuditLogger(logger *slog.Logger) *AuditLogger {
	if logger == nil {
		logger = slog.Default()
	}
	return &AuditLogger{logger: logger}
}

// AuditEvent represents an auditable event.
type AuditEvent struct {
	Timestamp   time.Time      `json:"timestamp"`
	EventType   string         `json:"event_type"`
	ExecutionID string         `json:"execution_id"`
	TenantID    string         `json:"tenant_id"`
	WorkflowID  string         `json:"workflow_id"`
	NodeID      string         `json:"node_id"`
	UserID      string         `json:"user_id,omitempty"`
	Action      string         `json:"action"`
	Outcome     string         `json:"outcome"`
	Details     map[string]any `json:"details,omitempty"`
}

// LogAuditEvent logs an audit event.
func (al *AuditLogger) LogAuditEvent(event *AuditEvent) {
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now().UTC()
	}

	al.logger.Info("javascript audit event",
		"timestamp", event.Timestamp,
		"event_type", event.EventType,
		"execution_id", event.ExecutionID,
		"tenant_id", event.TenantID,
		"workflow_id", event.WorkflowID,
		"node_id", event.NodeID,
		"user_id", event.UserID,
		"action", event.Action,
		"outcome", event.Outcome,
		"details", event.Details,
	)
}

// LogExecutionAttempt logs an execution attempt for audit purposes.
func (al *AuditLogger) LogExecutionAttempt(
	executionID, tenantID, workflowID, nodeID, userID string,
	scriptHash string,
) {
	al.LogAuditEvent(&AuditEvent{
		EventType:   "javascript_execution",
		ExecutionID: executionID,
		TenantID:    tenantID,
		WorkflowID:  workflowID,
		NodeID:      nodeID,
		UserID:      userID,
		Action:      "execute",
		Outcome:     "attempted",
		Details: map[string]any{
			"script_hash": scriptHash,
		},
	})
}

// LogExecutionResult logs the result of an execution for audit purposes.
func (al *AuditLogger) LogExecutionResult(
	executionID, tenantID, workflowID, nodeID string,
	success bool,
	duration time.Duration,
	errorMsg string,
) {
	outcome := "success"
	if !success {
		outcome = "failure"
	}

	details := map[string]any{
		"duration_ms": duration.Milliseconds(),
	}
	if errorMsg != "" {
		details["error"] = errorMsg
	}

	al.LogAuditEvent(&AuditEvent{
		EventType:   "javascript_execution",
		ExecutionID: executionID,
		TenantID:    tenantID,
		WorkflowID:  workflowID,
		NodeID:      nodeID,
		Action:      "complete",
		Outcome:     outcome,
		Details:     details,
	})
}

// LogSecurityViolation logs a security violation for audit purposes.
func (al *AuditLogger) LogSecurityViolation(
	executionID, tenantID, workflowID, nodeID string,
	violationType string,
	details map[string]any,
) {
	al.LogAuditEvent(&AuditEvent{
		EventType:   "javascript_security_violation",
		ExecutionID: executionID,
		TenantID:    tenantID,
		WorkflowID:  workflowID,
		NodeID:      nodeID,
		Action:      "security_violation",
		Outcome:     "blocked",
		Details: map[string]any{
			"violation_type": violationType,
			"details":        details,
		},
	})
}
