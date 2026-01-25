package javascript

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/dop251/goja"
	"github.com/google/uuid"
)

// Engine provides a sandboxed JavaScript execution environment.
type Engine struct {
	pool           *VMPool
	sandbox        *Sandbox
	injector       *ContextInjector
	extractor      *ResultExtractor
	logger         *ExecutionLogger
	auditLogger    *AuditLogger
	tracer         *ExecutionTracer
	limits         *Limits
	consoleCapture bool
}

// EngineConfig holds configuration for the JavaScript engine.
type EngineConfig struct {
	// Limits defines resource constraints.
	Limits *Limits

	// SandboxConfig defines sandbox settings.
	SandboxConfig *SandboxConfig

	// PoolSize is the number of VMs to keep in the pool.
	PoolSize int

	// Logger is the structured logger for execution events.
	Logger *slog.Logger

	// EnableConsoleCapture enables capturing console output.
	EnableConsoleCapture bool
}

// DefaultEngineConfig returns the default engine configuration.
func DefaultEngineConfig() *EngineConfig {
	return &EngineConfig{
		Limits:               DefaultLimits(),
		SandboxConfig:        DefaultSandboxConfig(),
		PoolSize:             10,
		Logger:               slog.Default(),
		EnableConsoleCapture: true,
	}
}

// NewEngine creates a new JavaScript execution engine.
func NewEngine(config *EngineConfig) (*Engine, error) {
	if config == nil {
		config = DefaultEngineConfig()
	}

	if config.Limits == nil {
		config.Limits = DefaultLimits()
	}

	if err := config.Limits.Validate(); err != nil {
		return nil, fmt.Errorf("invalid limits: %w", err)
	}

	sandbox := NewSandbox(config.SandboxConfig)
	pool := NewVMPool(config.PoolSize, sandbox, config.Limits)

	return &Engine{
		pool:           pool,
		sandbox:        sandbox,
		injector:       NewContextInjector(),
		extractor:      NewResultExtractor(),
		logger:         NewExecutionLogger(config.Logger),
		auditLogger:    NewAuditLogger(config.Logger),
		tracer:         NewExecutionTracer(),
		limits:         config.Limits,
		consoleCapture: config.EnableConsoleCapture,
	}, nil
}

// ExecuteConfig holds configuration for a single script execution.
type ExecuteConfig struct {
	// Script is the JavaScript code to execute.
	Script string

	// Context is the execution context data.
	Context *ExecutionContext

	// Timeout overrides the default timeout for this execution.
	Timeout time.Duration

	// ExecutionID is a unique identifier for this execution.
	ExecutionID string

	// TenantID identifies the tenant.
	TenantID string

	// WorkflowID identifies the workflow.
	WorkflowID string

	// NodeID identifies the workflow node.
	NodeID string

	// UserID identifies the user who triggered execution.
	UserID string
}

// ExecuteResult holds the result of a script execution.
type ExecuteResult struct {
	// Result is the value returned by the script.
	Result any

	// ConsoleLogs contains captured console output.
	ConsoleLogs []ConsoleEntry

	// Duration is how long execution took.
	Duration time.Duration

	// MemoryDelta is the memory change during execution.
	MemoryDelta int64

	// ExecutionID is the unique identifier for this execution.
	ExecutionID string
}

// Execute executes a JavaScript script in a sandboxed environment.
func (e *Engine) Execute(ctx context.Context, config *ExecuteConfig) (*ExecuteResult, error) {
	if config.ExecutionID == "" {
		config.ExecutionID = uuid.New().String()
	}

	scriptHash := hashScript(config.Script)

	// Log execution attempt
	e.auditLogger.LogExecutionAttempt(
		config.ExecutionID,
		config.TenantID,
		config.WorkflowID,
		config.NodeID,
		config.UserID,
		scriptHash,
	)

	// Trace execution
	result, err := e.tracer.TraceExecution(
		ctx,
		config.ExecutionID,
		config.TenantID,
		config.WorkflowID,
		config.NodeID,
		func(ctx context.Context) (any, error) {
			return e.executeInternal(ctx, config, scriptHash)
		},
	)

	if err != nil {
		return nil, err
	}

	return result.(*ExecuteResult), nil
}

func (e *Engine) executeInternal(ctx context.Context, config *ExecuteConfig, scriptHash string) (*ExecuteResult, error) {
	startTime := time.Now()

	// Create log entry
	logEntry := &ExecutionLogEntry{
		ExecutionID: config.ExecutionID,
		TenantID:    config.TenantID,
		WorkflowID:  config.WorkflowID,
		NodeID:      config.NodeID,
		ScriptHash:  scriptHash,
		StartTime:   startTime,
	}

	e.logger.LogExecutionStart(logEntry)

	// Validate script
	if err := e.validateScript(config.Script); err != nil {
		logEntry.Duration = time.Since(startTime)
		logEntry.Success = false
		logEntry.Error = err.Error()
		logEntry.ErrorPhase = PhaseValidation
		e.logger.LogExecutionComplete(logEntry)
		return nil, err
	}

	// Determine timeout
	timeout := e.limits.Timeout
	if config.Timeout > 0 && config.Timeout <= MaxTimeout {
		timeout = config.Timeout
	}

	// Create timeout context
	execCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Get VM from pool
	vm, err := e.pool.Get(execCtx)
	if err != nil {
		logEntry.Duration = time.Since(startTime)
		logEntry.Success = false
		logEntry.Error = err.Error()
		logEntry.ErrorPhase = PhaseExecution
		e.logger.LogExecutionComplete(logEntry)
		return nil, fmt.Errorf("failed to get VM: %w", err)
	}
	defer e.pool.Put(vm)

	// Set up console capture
	var consoleCapture *ConsoleCapture
	if e.consoleCapture {
		consoleCapture = NewConsoleCapture()
		if err := consoleCapture.InstallInRuntime(vm); err != nil {
			return nil, fmt.Errorf("failed to install console capture: %w", err)
		}
	}

	// Set up resource monitoring
	monitor := NewResourceMonitor(&Limits{
		Timeout:          timeout,
		MaxCallStackSize: e.limits.MaxCallStackSize,
		MaxMemoryMB:      e.limits.MaxMemoryMB,
	})
	monitor.Start()
	defer monitor.Stop()

	// Start monitoring in background
	go monitor.MonitorWithContext(execCtx, 100*time.Millisecond)

	// Inject context
	if config.Context != nil {
		if err := e.injector.InjectContext(vm, config.Context); err != nil {
			logEntry.Duration = time.Since(startTime)
			logEntry.Success = false
			logEntry.Error = err.Error()
			logEntry.ErrorPhase = PhaseExecution
			e.logger.LogExecutionComplete(logEntry)
			return nil, fmt.Errorf("failed to inject context: %w", err)
		}
	}

	// Execute script
	result, execErr := e.runScript(execCtx, vm, config.Script, monitor)

	// Collect metrics
	duration := time.Since(startTime)
	memoryDelta := monitor.GetMemoryDelta()

	// Build result
	execResult := &ExecuteResult{
		Duration:    duration,
		MemoryDelta: memoryDelta,
		ExecutionID: config.ExecutionID,
	}

	if consoleCapture != nil {
		execResult.ConsoleLogs = consoleCapture.GetLogs()
	}

	// Update log entry
	logEntry.Duration = duration
	logEntry.MemoryDelta = memoryDelta
	if consoleCapture != nil {
		logEntry.ConsoleLogs = consoleCapture.GetLogs()
	}

	if execErr != nil {
		logEntry.Success = false
		logEntry.Error = execErr.Error()
		logEntry.ErrorPhase = PhaseExecution
		e.logger.LogExecutionComplete(logEntry)

		// Log audit result
		e.auditLogger.LogExecutionResult(
			config.ExecutionID,
			config.TenantID,
			config.WorkflowID,
			config.NodeID,
			false,
			duration,
			execErr.Error(),
		)

		return nil, execErr
	}

	// Extract result
	extractedResult, err := e.extractor.ExtractResult(result)
	if err != nil {
		logEntry.Success = false
		logEntry.Error = err.Error()
		logEntry.ErrorPhase = PhaseExtraction
		e.logger.LogExecutionComplete(logEntry)
		return nil, fmt.Errorf("failed to extract result: %w", err)
	}

	execResult.Result = extractedResult
	logEntry.Success = true
	e.logger.LogExecutionComplete(logEntry)

	// Log audit result
	e.auditLogger.LogExecutionResult(
		config.ExecutionID,
		config.TenantID,
		config.WorkflowID,
		config.NodeID,
		true,
		duration,
		"",
	)

	// Record metrics to trace
	e.tracer.RecordExecutionMetrics(ctx, duration, memoryDelta, len(execResult.ConsoleLogs))

	return execResult, nil
}

func (e *Engine) validateScript(script string) error {
	if script == "" {
		return WrapValidationError(ErrEmptyScript)
	}

	if err := ValidateScriptLength(script, e.limits.MaxScriptLength); err != nil {
		return WrapValidationError(err)
	}

	if err := e.sandbox.ValidateScript(script); err != nil {
		return WrapValidationError(err)
	}

	return nil
}

func (e *Engine) runScript(ctx context.Context, vm *goja.Runtime, script string, monitor *ResourceMonitor) (goja.Value, error) {
	// Wrap script in IIFE for proper return handling
	wrappedScript := "(function() {\n" + script + "\n})();"

	// Channel for result
	type result struct {
		value goja.Value
		err   error
	}
	resultChan := make(chan result, 1)

	// Execute in goroutine for timeout support
	go func() {
		defer func() {
			if r := recover(); r != nil {
				resultChan <- result{
					err: fmt.Errorf("script panic: %v", r),
				}
			}
		}()

		val, err := vm.RunString(wrappedScript)
		if err != nil {
			resultChan <- result{err: WrapExecutionError(err)}
			return
		}
		resultChan <- result{value: val}
	}()

	// Wait for result or timeout
	select {
	case res := <-resultChan:
		return res.value, res.err
	case <-ctx.Done():
		vm.Interrupt("execution timeout")
		return nil, ErrTimeout
	case <-monitor.InterruptChan():
		vm.Interrupt("resource limit exceeded")
		if monitor.IsInterrupted() {
			if err := monitor.CheckTimeout(); err != nil {
				return nil, err
			}
			if err := monitor.CheckMemory(); err != nil {
				return nil, err
			}
		}
		return nil, ErrInterrupted
	}
}

// Validate validates a script without executing it.
func (e *Engine) Validate(script string) error {
	return e.validateScript(script)
}

// Compile pre-compiles a script and returns any compilation errors.
func (e *Engine) Compile(script string) error {
	if err := e.validateScript(script); err != nil {
		return err
	}

	// Try to compile in a temporary VM
	vm := goja.New()
	wrappedScript := "(function() {\n" + script + "\n});"

	_, err := goja.Compile("script", wrappedScript, false)
	if err != nil {
		return WrapCompilationError(err)
	}

	// Also try to parse in the VM to catch more errors
	_, err = vm.RunString(wrappedScript)
	if err != nil {
		return WrapCompilationError(err)
	}

	return nil
}

// Close shuts down the engine and releases resources.
func (e *Engine) Close() error {
	e.pool.Close()
	return nil
}

// VMPool manages a pool of Goja runtime instances.
type VMPool struct {
	pool    chan *goja.Runtime
	sandbox *Sandbox
	limits  *Limits
	mu      sync.Mutex
	closed  bool
}

// NewVMPool creates a new VM pool.
func NewVMPool(size int, sandbox *Sandbox, limits *Limits) *VMPool {
	if size <= 0 {
		size = 10
	}

	pool := &VMPool{
		pool:    make(chan *goja.Runtime, size),
		sandbox: sandbox,
		limits:  limits,
	}

	// Pre-warm the pool
	for i := 0; i < size; i++ {
		vm := pool.createVM()
		if vm != nil {
			pool.pool <- vm
		}
	}

	return pool
}

func (p *VMPool) createVM() *goja.Runtime {
	vm := goja.New()

	// Apply sandbox restrictions
	if p.sandbox != nil {
		if err := p.sandbox.ApplyToRuntime(vm); err != nil {
			return nil
		}
	}

	// Set call stack limit
	if p.limits != nil {
		vm.SetMaxCallStackSize(p.limits.MaxCallStackSize)
	}

	return vm
}

// Get retrieves a VM from the pool or creates a new one.
func (p *VMPool) Get(ctx context.Context) (*goja.Runtime, error) {
	p.mu.Lock()
	if p.closed {
		p.mu.Unlock()
		return nil, ErrPoolClosed
	}
	p.mu.Unlock()

	select {
	case vm := <-p.pool:
		return vm, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		// Pool empty, create new VM
		vm := p.createVM()
		if vm == nil {
			return nil, fmt.Errorf("failed to create VM")
		}
		return vm, nil
	}
}

// Put returns a VM to the pool.
func (p *VMPool) Put(vm *goja.Runtime) {
	p.mu.Lock()
	if p.closed {
		p.mu.Unlock()
		return
	}
	p.mu.Unlock()

	// Reset the VM before returning to pool
	// Create fresh VM instead of reusing to avoid state leakage
	freshVM := p.createVM()
	if freshVM == nil {
		return
	}

	select {
	case p.pool <- freshVM:
		// Successfully returned to pool
	default:
		// Pool is full, discard
	}
}

// Close closes the pool and releases all VMs.
func (p *VMPool) Close() {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.closed {
		return
	}

	p.closed = true
	close(p.pool)
}

// Size returns the current number of VMs in the pool.
func (p *VMPool) Size() int {
	return len(p.pool)
}

// hashScript creates a SHA256 hash of the script for identification.
func hashScript(script string) string {
	hash := sha256.Sum256([]byte(script))
	return hex.EncodeToString(hash[:8]) // Use first 8 bytes for brevity
}
