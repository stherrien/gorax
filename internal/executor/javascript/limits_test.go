package javascript

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultLimits(t *testing.T) {
	limits := DefaultLimits()

	assert.Equal(t, DefaultTimeout, limits.Timeout)
	assert.Equal(t, DefaultMaxCallStackSize, limits.MaxCallStackSize)
	assert.Equal(t, int64(DefaultMaxMemoryMB), limits.MaxMemoryMB)
	assert.Equal(t, DefaultMaxScriptLength, limits.MaxScriptLength)
}

func TestNewLimits(t *testing.T) {
	tests := []struct {
		name            string
		timeout         time.Duration
		maxMemoryMB     int64
		expectedTimeout time.Duration
		expectedMemory  int64
	}{
		{
			name:            "default values",
			timeout:         0,
			maxMemoryMB:     0,
			expectedTimeout: DefaultTimeout,
			expectedMemory:  DefaultMaxMemoryMB,
		},
		{
			name:            "custom timeout",
			timeout:         10 * time.Second,
			maxMemoryMB:     0,
			expectedTimeout: 10 * time.Second,
			expectedMemory:  DefaultMaxMemoryMB,
		},
		{
			name:            "custom memory",
			timeout:         0,
			maxMemoryMB:     256,
			expectedTimeout: DefaultTimeout,
			expectedMemory:  256,
		},
		{
			name:            "timeout exceeds max",
			timeout:         120 * time.Second,
			maxMemoryMB:     0,
			expectedTimeout: MaxTimeout,
			expectedMemory:  DefaultMaxMemoryMB,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			limits := NewLimits(tc.timeout, tc.maxMemoryMB)
			assert.Equal(t, tc.expectedTimeout, limits.Timeout)
			assert.Equal(t, tc.expectedMemory, limits.MaxMemoryMB)
		})
	}
}

func TestLimits_Validate(t *testing.T) {
	tests := []struct {
		name    string
		limits  *Limits
		wantErr error
	}{
		{
			name:    "valid limits",
			limits:  DefaultLimits(),
			wantErr: nil,
		},
		{
			name: "zero timeout",
			limits: &Limits{
				Timeout:          0,
				MaxCallStackSize: 1000,
				MaxMemoryMB:      128,
				MaxScriptLength:  1024,
			},
			wantErr: ErrInvalidTimeout,
		},
		{
			name: "negative timeout",
			limits: &Limits{
				Timeout:          -1 * time.Second,
				MaxCallStackSize: 1000,
				MaxMemoryMB:      128,
				MaxScriptLength:  1024,
			},
			wantErr: ErrInvalidTimeout,
		},
		{
			name: "timeout exceeds max",
			limits: &Limits{
				Timeout:          120 * time.Second,
				MaxCallStackSize: 1000,
				MaxMemoryMB:      128,
				MaxScriptLength:  1024,
			},
			wantErr: ErrTimeoutExceedsMax,
		},
		{
			name: "zero stack size",
			limits: &Limits{
				Timeout:          30 * time.Second,
				MaxCallStackSize: 0,
				MaxMemoryMB:      128,
				MaxScriptLength:  1024,
			},
			wantErr: ErrInvalidStackSize,
		},
		{
			name: "zero memory",
			limits: &Limits{
				Timeout:          30 * time.Second,
				MaxCallStackSize: 1000,
				MaxMemoryMB:      0,
				MaxScriptLength:  1024,
			},
			wantErr: ErrInvalidMemoryLimit,
		},
		{
			name: "zero script length",
			limits: &Limits{
				Timeout:          30 * time.Second,
				MaxCallStackSize: 1000,
				MaxMemoryMB:      128,
				MaxScriptLength:  0,
			},
			wantErr: ErrInvalidScriptLength,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.limits.Validate()
			if tc.wantErr != nil {
				assert.ErrorIs(t, err, tc.wantErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestResourceMonitor_Start(t *testing.T) {
	monitor := NewResourceMonitor(DefaultLimits())

	monitor.Start()

	assert.False(t, monitor.IsInterrupted())
	assert.Greater(t, monitor.GetElapsedTime(), time.Duration(0))
}

func TestResourceMonitor_CheckTimeout(t *testing.T) {
	limits := &Limits{
		Timeout:          100 * time.Millisecond,
		MaxCallStackSize: 1000,
		MaxMemoryMB:      128,
		MaxScriptLength:  1024,
	}
	monitor := NewResourceMonitor(limits)
	monitor.Start()

	// Should not timeout immediately
	err := monitor.CheckTimeout()
	assert.NoError(t, err)

	// Wait for timeout
	time.Sleep(150 * time.Millisecond)

	err = monitor.CheckTimeout()
	assert.ErrorIs(t, err, ErrTimeout)
	assert.True(t, monitor.IsInterrupted())
}

func TestResourceMonitor_Interrupt(t *testing.T) {
	monitor := NewResourceMonitor(DefaultLimits())
	monitor.Start()

	assert.False(t, monitor.IsInterrupted())

	monitor.Interrupt()

	assert.True(t, monitor.IsInterrupted())

	// Should be able to receive from interrupt channel
	select {
	case <-monitor.InterruptChan():
		// Expected
	default:
		t.Error("Expected interrupt channel to be closed")
	}
}

func TestResourceMonitor_GetElapsedTime(t *testing.T) {
	monitor := NewResourceMonitor(DefaultLimits())
	monitor.Start()

	time.Sleep(50 * time.Millisecond)

	elapsed := monitor.GetElapsedTime()
	assert.GreaterOrEqual(t, elapsed, 50*time.Millisecond)
}

func TestResourceMonitor_GetLimits(t *testing.T) {
	limits := &Limits{
		Timeout:          10 * time.Second,
		MaxCallStackSize: 500,
		MaxMemoryMB:      64,
		MaxScriptLength:  2048,
	}
	monitor := NewResourceMonitor(limits)

	retrieved := monitor.GetLimits()

	assert.Equal(t, limits.Timeout, retrieved.Timeout)
	assert.Equal(t, limits.MaxCallStackSize, retrieved.MaxCallStackSize)
	assert.Equal(t, limits.MaxMemoryMB, retrieved.MaxMemoryMB)
	assert.Equal(t, limits.MaxScriptLength, retrieved.MaxScriptLength)
}

func TestResourceMonitor_MonitorWithContext(t *testing.T) {
	limits := &Limits{
		Timeout:          1 * time.Second,
		MaxCallStackSize: 1000,
		MaxMemoryMB:      128,
		MaxScriptLength:  1024,
	}
	monitor := NewResourceMonitor(limits)
	monitor.Start()

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	done := make(chan struct{})
	go func() {
		monitor.MonitorWithContext(ctx, 10*time.Millisecond)
		close(done)
	}()

	<-done

	assert.True(t, monitor.IsInterrupted())
}

func TestValidateScriptLength(t *testing.T) {
	tests := []struct {
		name      string
		script    string
		maxLength int
		wantErr   bool
	}{
		{
			name:      "within limit",
			script:    "return 42;",
			maxLength: 100,
			wantErr:   false,
		},
		{
			name:      "at limit",
			script:    "x",
			maxLength: 1,
			wantErr:   false,
		},
		{
			name:      "exceeds limit",
			script:    "return 42;",
			maxLength: 5,
			wantErr:   true,
		},
		{
			name:      "default limit",
			script:    "return 42;",
			maxLength: 0,
			wantErr:   false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateScriptLength(tc.script, tc.maxLength)
			if tc.wantErr {
				assert.ErrorIs(t, err, ErrScriptTooLarge)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGetCurrentMemoryUsage(t *testing.T) {
	mem := getCurrentMemoryUsage()
	assert.Greater(t, mem, uint64(0))
}

func TestResourceMonitor_GetMemoryDelta(t *testing.T) {
	monitor := NewResourceMonitor(DefaultLimits())
	monitor.Start()

	// The delta might be small or even negative due to GC
	// Just verify it doesn't panic and returns a value
	delta := monitor.GetMemoryDelta()
	require.NotPanics(t, func() {
		_ = delta
	})
}
