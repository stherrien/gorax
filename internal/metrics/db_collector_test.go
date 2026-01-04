package metrics

import (
	"context"
	"database/sql"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	_ "github.com/mattn/go-sqlite3"
)

func TestNewDBStatsCollector(t *testing.T) {
	// Given: metrics and db connection
	m := NewMetrics()
	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)
	defer db.Close()

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	// When: creating a new collector
	collector := NewDBStatsCollector(m, db, "main", logger)

	// Then: collector should be initialized
	assert.NotNil(t, collector)
	assert.Equal(t, "main", collector.poolName)
}

func TestDBStatsCollector_CollectOnce(t *testing.T) {
	// Given: metrics and db connection with registry
	m := NewMetrics()
	registry := prometheus.NewRegistry()
	err := m.Register(registry)
	require.NoError(t, err)

	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)
	defer db.Close()

	// Set pool limits
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	collector := NewDBStatsCollector(m, db, "main", logger)

	// When: collecting stats once
	collector.collectOnce()

	// Then: metrics should be recorded
	metrics, err := registry.Gather()
	require.NoError(t, err)

	foundOpen := false
	foundIdle := false
	foundInUse := false

	for _, metric := range metrics {
		switch metric.GetName() {
		case "gorax_db_connections_open":
			foundOpen = true
		case "gorax_db_connections_idle":
			foundIdle = true
		case "gorax_db_connections_in_use":
			foundInUse = true
		}
	}

	assert.True(t, foundOpen, "db connections open gauge should be present")
	assert.True(t, foundIdle, "db connections idle gauge should be present")
	assert.True(t, foundInUse, "db connections in use gauge should be present")
}

func TestDBStatsCollector_Start(t *testing.T) {
	// Given: metrics and db connection
	m := NewMetrics()
	registry := prometheus.NewRegistry()
	err := m.Register(registry)
	require.NoError(t, err)

	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)
	defer db.Close()

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	collector := NewDBStatsCollector(m, db, "main", logger)

	// When: starting collector with short interval
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	go collector.Start(ctx, 25*time.Millisecond)

	// Wait for at least one collection cycle
	time.Sleep(50 * time.Millisecond)

	// Then: metrics should be collected
	metrics, err := registry.Gather()
	require.NoError(t, err)
	assert.NotEmpty(t, metrics)
}

func TestDBStatsCollector_Stop(t *testing.T) {
	// Given: running collector
	m := NewMetrics()
	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)
	defer db.Close()

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	collector := NewDBStatsCollector(m, db, "main", logger)

	ctx := context.Background()
	go collector.Start(ctx, 100*time.Millisecond)

	// Wait for collector to start
	time.Sleep(50 * time.Millisecond)

	// When: stopping collector
	collector.Stop()

	// Then: collector should stop without error
	time.Sleep(50 * time.Millisecond)
	// If we get here without hanging, the test passes
}
