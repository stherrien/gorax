// Package javascript provides a sandboxed JavaScript execution engine for workflow automation.
package javascript

import (
	"context"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

// Default resource limits.
const (
	DefaultTimeout          = 30 * time.Second
	DefaultMaxCallStackSize = 1000
	DefaultMaxMemoryMB      = 128
	DefaultMaxScriptLength  = 1024 * 1024 // 1MB
	MaxTimeout              = 60 * time.Second
)

// Limits defines resource constraints for JavaScript execution.
type Limits struct {
	// Timeout is the maximum execution time.
	Timeout time.Duration

	// MaxCallStackSize prevents stack overflow attacks.
	MaxCallStackSize int

	// MaxMemoryMB limits memory usage (soft limit, monitored).
	MaxMemoryMB int64

	// MaxScriptLength limits script size.
	MaxScriptLength int
}

// DefaultLimits returns the default resource limits.
func DefaultLimits() *Limits {
	return &Limits{
		Timeout:          DefaultTimeout,
		MaxCallStackSize: DefaultMaxCallStackSize,
		MaxMemoryMB:      DefaultMaxMemoryMB,
		MaxScriptLength:  DefaultMaxScriptLength,
	}
}

// NewLimits creates limits with optional overrides.
func NewLimits(timeout time.Duration, maxMemoryMB int64) *Limits {
	limits := DefaultLimits()

	if timeout > 0 {
		if timeout > MaxTimeout {
			timeout = MaxTimeout
		}
		limits.Timeout = timeout
	}

	if maxMemoryMB > 0 {
		limits.MaxMemoryMB = maxMemoryMB
	}

	return limits
}

// Validate checks if the limits are valid.
func (l *Limits) Validate() error {
	if l.Timeout <= 0 {
		return ErrInvalidTimeout
	}
	if l.Timeout > MaxTimeout {
		return ErrTimeoutExceedsMax
	}
	if l.MaxCallStackSize <= 0 {
		return ErrInvalidStackSize
	}
	if l.MaxMemoryMB <= 0 {
		return ErrInvalidMemoryLimit
	}
	if l.MaxScriptLength <= 0 {
		return ErrInvalidScriptLength
	}
	return nil
}

// ResourceMonitor tracks resource usage during script execution.
type ResourceMonitor struct {
	mu             sync.RWMutex
	startTime      time.Time
	startMemory    uint64
	limits         *Limits
	interrupted    atomic.Bool
	interruptChan  chan struct{}
	stopMonitoring chan struct{}
}

// NewResourceMonitor creates a new resource monitor.
func NewResourceMonitor(limits *Limits) *ResourceMonitor {
	if limits == nil {
		limits = DefaultLimits()
	}
	return &ResourceMonitor{
		limits:         limits,
		interruptChan:  make(chan struct{}),
		stopMonitoring: make(chan struct{}),
	}
}

// Start begins resource monitoring.
func (r *ResourceMonitor) Start() {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.startTime = time.Now()
	r.startMemory = getCurrentMemoryUsage()
	r.interrupted.Store(false)
}

// Stop ends resource monitoring.
func (r *ResourceMonitor) Stop() {
	close(r.stopMonitoring)
}

// InterruptChan returns a channel that is closed when execution should stop.
func (r *ResourceMonitor) InterruptChan() <-chan struct{} {
	return r.interruptChan
}

// Interrupt signals that execution should stop.
func (r *ResourceMonitor) Interrupt() {
	if r.interrupted.CompareAndSwap(false, true) {
		close(r.interruptChan)
	}
}

// IsInterrupted returns whether execution has been interrupted.
func (r *ResourceMonitor) IsInterrupted() bool {
	return r.interrupted.Load()
}

// CheckTimeout returns an error if the timeout has been exceeded.
func (r *ResourceMonitor) CheckTimeout() error {
	r.mu.RLock()
	elapsed := time.Since(r.startTime)
	timeout := r.limits.Timeout
	r.mu.RUnlock()

	if elapsed > timeout {
		r.Interrupt()
		return ErrTimeout
	}
	return nil
}

// CheckMemory returns an error if memory limits are exceeded.
func (r *ResourceMonitor) CheckMemory() error {
	currentMemory := getCurrentMemoryUsage()

	r.mu.RLock()
	startMem := r.startMemory
	maxMemMB := r.limits.MaxMemoryMB
	r.mu.RUnlock()

	// Calculate memory increase since start
	memIncreaseMB := int64((currentMemory - startMem) / (1024 * 1024))

	if memIncreaseMB > maxMemMB {
		r.Interrupt()
		return ErrMemoryLimitExceeded
	}
	return nil
}

// MonitorWithContext runs periodic checks in a goroutine until context is done.
func (r *ResourceMonitor) MonitorWithContext(ctx context.Context, checkInterval time.Duration) {
	ticker := time.NewTicker(checkInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			r.Interrupt()
			return
		case <-r.stopMonitoring:
			return
		case <-ticker.C:
			if err := r.CheckTimeout(); err != nil {
				return
			}
			if err := r.CheckMemory(); err != nil {
				return
			}
		}
	}
}

// GetElapsedTime returns the time elapsed since monitoring started.
func (r *ResourceMonitor) GetElapsedTime() time.Duration {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return time.Since(r.startTime)
}

// GetMemoryDelta returns the memory change since monitoring started.
func (r *ResourceMonitor) GetMemoryDelta() int64 {
	currentMemory := getCurrentMemoryUsage()
	r.mu.RLock()
	startMem := r.startMemory
	r.mu.RUnlock()
	return int64(currentMemory - startMem)
}

// GetLimits returns a copy of the current limits.
func (r *ResourceMonitor) GetLimits() *Limits {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return &Limits{
		Timeout:          r.limits.Timeout,
		MaxCallStackSize: r.limits.MaxCallStackSize,
		MaxMemoryMB:      r.limits.MaxMemoryMB,
		MaxScriptLength:  r.limits.MaxScriptLength,
	}
}

// getCurrentMemoryUsage returns the current memory allocation in bytes.
func getCurrentMemoryUsage() uint64 {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	return memStats.Alloc
}

// ValidateScriptLength checks if the script is within size limits.
func ValidateScriptLength(script string, maxLength int) error {
	if maxLength <= 0 {
		maxLength = DefaultMaxScriptLength
	}
	if len(script) > maxLength {
		return ErrScriptTooLarge
	}
	return nil
}
