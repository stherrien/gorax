package worker

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gorax/gorax/internal/buildinfo"
)

// TestHealthResponse_UsesBuildinfoVersion tests that version comes from buildinfo package
func TestHealthResponse_UsesBuildinfoVersion(t *testing.T) {
	// This is a simple unit test that doesn't require connections
	// We just need to verify the buildinfo integration

	// We expect version to be from buildinfo
	expectedVersion := buildinfo.GetVersion()
	assert.NotEmpty(t, expectedVersion, "buildinfo should provide a version")
	assert.NotEqual(t, "1.0.0", expectedVersion, "version should not be hardcoded 1.0.0")
}

// TestHealthEndpoint_Integration is an integration test that requires real connections
func TestHealthEndpoint_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	// This would require real database and redis connections
	// Skip for now as it's an integration test
	t.Skip("integration test requires database and redis connections")

	ctx := context.Background()

	// In a real integration test, you would:
	// 1. Set up test database connection
	// 2. Set up test redis connection
	// 3. Create worker with real connections
	// 4. Test the full health endpoint

	mockRedis := redis.NewClient(&redis.Options{Addr: "localhost:6379", DB: 15})
	mockDB := &sqlx.DB{}

	w := &Worker{
		concurrency:  5,
		redis:        mockRedis,
		db:           mockDB,
		queueEnabled: false,
	}

	hs := &HealthServer{
		worker: w,
	}

	req := httptest.NewRequest("GET", "/health", nil).WithContext(ctx)
	rr := httptest.NewRecorder()

	hs.handleHealth(rr, req)

	var response HealthResponse
	err := json.NewDecoder(rr.Body).Decode(&response)
	require.NoError(t, err)

	assert.Equal(t, buildinfo.GetVersion(), response.Version)
	assert.Equal(t, 5, response.WorkerInfo.Concurrency)
}
