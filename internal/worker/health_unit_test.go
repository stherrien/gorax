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
