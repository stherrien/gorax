package connectors

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq" // PostgreSQL driver
)

// PostgreSQLConnector implements the Connector interface for PostgreSQL
type PostgreSQLConnector struct {
	db *sqlx.DB
}

// NewPostgreSQLConnector creates a new PostgreSQL connector
func NewPostgreSQLConnector() *PostgreSQLConnector {
	return &PostgreSQLConnector{}
}

// Connect establishes a connection to PostgreSQL
func (c *PostgreSQLConnector) Connect(ctx context.Context, connectionString string) error {
	// Validate and sanitize connection string
	if err := c.validateConnectionString(connectionString); err != nil {
		return fmt.Errorf("invalid connection string: %w", err)
	}

	// Open connection
	db, err := sqlx.ConnectContext(ctx, "postgres", connectionString)
	if err != nil {
		return fmt.Errorf("failed to connect to PostgreSQL: %w", err)
	}

	// Set connection pool settings
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)
	db.SetConnMaxIdleTime(1 * time.Minute)

	c.db = db
	return nil
}

// Close closes the PostgreSQL connection
func (c *PostgreSQLConnector) Close() error {
	if c.db != nil {
		return c.db.Close()
	}
	return nil
}

// Ping tests the PostgreSQL connection
func (c *PostgreSQLConnector) Ping(ctx context.Context) error {
	if c.db == nil {
		return ErrConnectionFailed
	}
	return c.db.PingContext(ctx)
}

// Query executes a SELECT query
func (c *PostgreSQLConnector) Query(ctx context.Context, input *QueryInput) (*QueryResult, error) {
	if c.db == nil {
		return nil, ErrConnectionFailed
	}

	// Validate input
	if err := input.Validate(); err != nil {
		return nil, err
	}

	// Set default timeout
	timeout := 30 * time.Second
	if input.Timeout > 0 {
		timeout = time.Duration(input.Timeout) * time.Second
	}

	// Create context with timeout
	queryCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Set default max rows
	maxRows := 1000
	if input.MaxRows > 0 {
		maxRows = input.MaxRows
	}

	// Validate query is SELECT only
	if !c.isSelectQuery(input.Query) {
		return nil, fmt.Errorf("%w: only SELECT queries allowed in Query method", ErrInvalidQuery)
	}

	// Execute query with timing
	startTime := time.Now()
	rows, err := c.db.QueryxContext(queryCtx, input.Query, input.Parameters...)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrQueryFailed, err)
	}
	defer rows.Close()

	// Fetch results
	results := make([]map[string]interface{}, 0, maxRows)
	rowCount := 0

	for rows.Next() {
		if rowCount >= maxRows {
			return nil, fmt.Errorf("%w: query returned more than %d rows", ErrRowLimitExceeded, maxRows)
		}

		result := make(map[string]interface{})
		if err := rows.MapScan(result); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		// Convert byte arrays to strings for JSON serialization
		c.normalizeResult(result)

		results = append(results, result)
		rowCount++
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	executionTime := time.Since(startTime)

	return &QueryResult{
		Rows:         results,
		RowsAffected: rowCount,
		ExecutionMS:  executionTime.Milliseconds(),
		Metadata:     input.Metadata,
	}, nil
}

// Execute executes a query that modifies data
func (c *PostgreSQLConnector) Execute(ctx context.Context, input *QueryInput) (*QueryResult, error) {
	if c.db == nil {
		return nil, ErrConnectionFailed
	}

	// Validate input
	if err := input.Validate(); err != nil {
		return nil, err
	}

	// Set default timeout
	timeout := 30 * time.Second
	if input.Timeout > 0 {
		timeout = time.Duration(input.Timeout) * time.Second
	}

	// Create context with timeout
	queryCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Validate query is not SELECT
	if c.isSelectQuery(input.Query) {
		return nil, fmt.Errorf("%w: SELECT queries not allowed in Execute method, use Query instead", ErrInvalidQuery)
	}

	// Execute query with timing
	startTime := time.Now()
	result, err := c.db.ExecContext(queryCtx, input.Query, input.Parameters...)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrQueryFailed, err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		rowsAffected = 0
	}

	executionTime := time.Since(startTime)

	return &QueryResult{
		RowsAffected: int(rowsAffected),
		ExecutionMS:  executionTime.Milliseconds(),
		Metadata:     input.Metadata,
	}, nil
}

// GetDatabaseType returns the database type
func (c *PostgreSQLConnector) GetDatabaseType() DatabaseType {
	return DatabaseTypePostgreSQL
}

// validateConnectionString validates the PostgreSQL connection string
func (c *PostgreSQLConnector) validateConnectionString(connStr string) error {
	// Parse connection string
	if connStr == "" {
		return ErrInvalidConnectionString
	}

	// Check for basic validity (postgres:// or postgresql://)
	if !strings.HasPrefix(connStr, "postgres://") && !strings.HasPrefix(connStr, "postgresql://") {
		// Try parsing as key=value format
		if !strings.Contains(connStr, "host=") {
			return fmt.Errorf("%w: must start with postgres:// or postgresql:// or contain host=", ErrInvalidConnectionString)
		}
	} else {
		// Parse as URL
		u, err := url.Parse(connStr)
		if err != nil {
			return fmt.Errorf("%w: %v", ErrInvalidConnectionString, err)
		}

		// Check for SSRF - prevent connections to localhost, private IPs
		if err := c.validateHost(u.Hostname()); err != nil {
			return err
		}
	}

	return nil
}

// validateHost validates the database host to prevent SSRF
func (c *PostgreSQLConnector) validateHost(host string) error {
	// Block localhost and loopback addresses
	if host == "localhost" || host == "127.0.0.1" || host == "::1" {
		return fmt.Errorf("connections to localhost are not allowed for security reasons")
	}

	// Block private IP ranges (basic check)
	if strings.HasPrefix(host, "192.168.") ||
		strings.HasPrefix(host, "10.") ||
		strings.HasPrefix(host, "172.16.") ||
		strings.HasPrefix(host, "172.17.") ||
		strings.HasPrefix(host, "172.18.") ||
		strings.HasPrefix(host, "172.19.") ||
		strings.HasPrefix(host, "172.20.") ||
		strings.HasPrefix(host, "172.21.") ||
		strings.HasPrefix(host, "172.22.") ||
		strings.HasPrefix(host, "172.23.") ||
		strings.HasPrefix(host, "172.24.") ||
		strings.HasPrefix(host, "172.25.") ||
		strings.HasPrefix(host, "172.26.") ||
		strings.HasPrefix(host, "172.27.") ||
		strings.HasPrefix(host, "172.28.") ||
		strings.HasPrefix(host, "172.29.") ||
		strings.HasPrefix(host, "172.30.") ||
		strings.HasPrefix(host, "172.31.") {
		return fmt.Errorf("connections to private IP addresses are not allowed for security reasons")
	}

	return nil
}

// isSelectQuery checks if the query is a SELECT statement
func (c *PostgreSQLConnector) isSelectQuery(query string) bool {
	trimmed := strings.TrimSpace(strings.ToUpper(query))
	return strings.HasPrefix(trimmed, "SELECT") || strings.HasPrefix(trimmed, "WITH")
}

// normalizeResult converts byte arrays to strings for JSON serialization
func (c *PostgreSQLConnector) normalizeResult(result map[string]interface{}) {
	for key, value := range result {
		switch v := value.(type) {
		case []byte:
			result[key] = string(v)
		case sql.NullString:
			if v.Valid {
				result[key] = v.String
			} else {
				result[key] = nil
			}
		case sql.NullInt64:
			if v.Valid {
				result[key] = v.Int64
			} else {
				result[key] = nil
			}
		case sql.NullFloat64:
			if v.Valid {
				result[key] = v.Float64
			} else {
				result[key] = nil
			}
		case sql.NullBool:
			if v.Valid {
				result[key] = v.Bool
			} else {
				result[key] = nil
			}
		case sql.NullTime:
			if v.Valid {
				result[key] = v.Time
			} else {
				result[key] = nil
			}
		}
	}
}
