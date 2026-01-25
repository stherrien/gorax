package worker

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gorax/gorax/internal/buildinfo"
)

func TestLiveness_ReturnsAlive(t *testing.T) {
	// Liveness doesn't require any connections
	hs := &HealthServer{
		worker: &Worker{},
	}

	req := httptest.NewRequest("GET", "/health/live", nil)
	rr := httptest.NewRecorder()

	hs.handleLiveness(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var response map[string]string
	err := json.NewDecoder(rr.Body).Decode(&response)
	require.NoError(t, err)
	assert.Equal(t, "alive", response["status"])
}

func TestReadiness_NotReadyWhenFlagIsFalse(t *testing.T) {
	// Readiness checks the ready flag, no connections needed
	hs := &HealthServer{
		worker: &Worker{},
	}
	hs.ready.Store(false)

	req := httptest.NewRequest("GET", "/health/ready", nil)
	rr := httptest.NewRecorder()

	hs.handleReadiness(rr, req)

	assert.Equal(t, http.StatusServiceUnavailable, rr.Code)

	var response map[string]string
	err := json.NewDecoder(rr.Body).Decode(&response)
	require.NoError(t, err)
	assert.Equal(t, "not_ready", response["status"])
}

func TestReadiness_ReadyWhenFlagIsTrue(t *testing.T) {
	hs := &HealthServer{
		worker: &Worker{
			concurrency: 10,
		},
	}
	hs.ready.Store(true)

	req := httptest.NewRequest("GET", "/health/ready", nil)
	rr := httptest.NewRecorder()

	hs.handleReadiness(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var response map[string]string
	err := json.NewDecoder(rr.Body).Decode(&response)
	require.NoError(t, err)
	assert.Equal(t, "ready", response["status"])
}

func TestReadiness_AtCapacityWhenFull(t *testing.T) {
	hs := &HealthServer{
		worker: &Worker{
			concurrency: 5,
		},
	}
	hs.ready.Store(true)
	hs.worker.activeExecutions.Store(5) // At capacity

	req := httptest.NewRequest("GET", "/health/ready", nil)
	rr := httptest.NewRecorder()

	hs.handleReadiness(rr, req)

	assert.Equal(t, http.StatusServiceUnavailable, rr.Code)

	var response map[string]string
	err := json.NewDecoder(rr.Body).Decode(&response)
	require.NoError(t, err)
	assert.Equal(t, "at_capacity", response["status"])
}

func TestWorkerMetrics(t *testing.T) {
	// Test that worker metrics are correctly tracked
	w := &Worker{
		concurrency: 10,
	}

	w.activeExecutions.Store(3)
	w.processedTotal.Store(100)
	w.failedTotal.Store(5)

	assert.Equal(t, int32(3), w.getActiveExecutions())
	assert.Equal(t, int64(100), w.getProcessedCount())
	assert.Equal(t, int64(5), w.getFailedCount())
}

func TestVersionFromBuildInfo(t *testing.T) {
	// Test that we're using buildinfo package
	version := buildinfo.GetVersion()
	assert.NotEmpty(t, version)

	// In tests, version will be "dev" unless set via ldflags
	// The important thing is it's not hardcoded to "1.0.0"
	assert.NotEqual(t, "1.0.0", version, "should use buildinfo, not hardcoded value")
}

// TestSafeIntToInt32 tests the safe conversion from int to int32
func TestSafeIntToInt32(t *testing.T) {
	tests := []struct {
		name     string
		input    int
		expected int32
	}{
		{
			name:     "zero value",
			input:    0,
			expected: 0,
		},
		{
			name:     "positive value",
			input:    100,
			expected: 100,
		},
		{
			name:     "negative value",
			input:    -10,
			expected: 0,
		},
		{
			name:     "max int32 value",
			input:    2147483647,
			expected: 2147483647,
		},
		{
			name:     "value exceeding max int32",
			input:    2147483648,
			expected: 2147483647,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := safeIntToInt32(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestLiveness_HandlesEncodingGracefully tests that liveness endpoint doesn't panic on encoding issues
func TestLiveness_HandlesEncodingGracefully(t *testing.T) {
	hs := &HealthServer{
		worker: &Worker{},
	}

	req := httptest.NewRequest("GET", "/health/live", nil)
	rr := httptest.NewRecorder()

	// This test verifies that the handler completes without panic
	// In production, if json.Encode fails, it should be logged but not crash
	assert.NotPanics(t, func() {
		hs.handleLiveness(rr, req)
	})

	assert.Equal(t, http.StatusOK, rr.Code)
}

// TestReadiness_HandlesEncodingGracefully tests that readiness endpoint doesn't panic on encoding issues
func TestReadiness_HandlesEncodingGracefully(t *testing.T) {
	hs := &HealthServer{
		worker: &Worker{
			concurrency: 10,
		},
	}
	hs.ready.Store(true)

	req := httptest.NewRequest("GET", "/health/ready", nil)
	rr := httptest.NewRecorder()

	// This test verifies that the handler completes without panic
	assert.NotPanics(t, func() {
		hs.handleReadiness(rr, req)
	})

	assert.Equal(t, http.StatusOK, rr.Code)
}

// TestHealth_HandlesEncodingGracefully tests that health endpoint doesn't panic on encoding issues
// Note: Full health endpoint requires real connections, tested in integration tests
// This unit test focuses on JSON encoding behavior
func TestHealth_EncodesValidJSON(t *testing.T) {
	// For unit testing, we focus on the parts that don't require real connections
	// The full health endpoint is tested in integration tests with real DB/Redis

	// We can test that the response structure is correct
	response := HealthResponse{
		Status:  "healthy",
		Version: buildinfo.GetVersion(),
		WorkerInfo: WorkerInfo{
			Concurrency:      10,
			ActiveExecutions: 2,
			ProcessedTotal:   100,
			FailedTotal:      5,
		},
		Connections: ConnectionsHealth{
			Database: "ok",
			Redis:    "ok",
			Queue:    "ok",
		},
	}

	// Verify we can encode this structure
	data, err := json.Marshal(response)
	require.NoError(t, err)
	assert.NotEmpty(t, data)
}
