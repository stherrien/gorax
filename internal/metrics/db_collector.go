package metrics

import (
	"context"
	"database/sql"
	"log/slog"
	"time"
)

// DBStatsCollector periodically collects database connection pool statistics
type DBStatsCollector struct {
	metrics  *Metrics
	db       *sql.DB
	poolName string
	logger   *slog.Logger
	stopCh   chan struct{}
}

// NewDBStatsCollector creates a new database statistics collector
func NewDBStatsCollector(metrics *Metrics, db *sql.DB, poolName string, logger *slog.Logger) *DBStatsCollector {
	return &DBStatsCollector{
		metrics:  metrics,
		db:       db,
		poolName: poolName,
		logger:   logger,
		stopCh:   make(chan struct{}),
	}
}

// Start begins collecting database statistics at regular intervals
func (c *DBStatsCollector) Start(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	// Collect initial statistics
	c.collectOnce()

	for {
		select {
		case <-ctx.Done():
			return
		case <-c.stopCh:
			return
		case <-ticker.C:
			c.collectOnce()
		}
	}
}

// Stop stops the database statistics collector
func (c *DBStatsCollector) Stop() {
	close(c.stopCh)
}

// collectOnce performs a single collection cycle
func (c *DBStatsCollector) collectOnce() {
	stats := c.db.Stats()

	c.metrics.SetDBConnectionPoolStats(
		c.poolName,
		stats.OpenConnections,
		stats.Idle,
		stats.InUse,
	)

	c.logger.Debug("database connection pool stats collected",
		"pool", c.poolName,
		"open", stats.OpenConnections,
		"idle", stats.Idle,
		"in_use", stats.InUse,
		"max_open", stats.MaxOpenConnections,
		"wait_count", stats.WaitCount,
		"wait_duration", stats.WaitDuration,
	)
}
