package connectors

import (
	"context"
	"database/sql"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3" // SQLite driver
)

// SQLiteConnector implements the Connector interface for SQLite
type SQLiteConnector struct {
	db *sqlx.DB
}

// NewSQLiteConnector creates a new SQLite connector
func NewSQLiteConnector() *SQLiteConnector {
	return &SQLiteConnector{}
}

// Connect establishes a connection to SQLite
func (c *SQLiteConnector) Connect(ctx context.Context, connectionString string) error {
	// Validate connection string (file path)
	if err := c.validateConnectionString(connectionString); err != nil {
		return fmt.Errorf("invalid connection string: %w", err)
	}

	// Open connection
	db, err := sqlx.ConnectContext(ctx, "sqlite3", connectionString)
	if err != nil {
		return fmt.Errorf("failed to connect to SQLite: %w", err)
	}

	// Set connection pool settings (SQLite is single-writer)
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	db.SetConnMaxLifetime(0)

	c.db = db
	return nil
}

// Close closes the SQLite connection
func (c *SQLiteConnector) Close() error {
	if c.db != nil {
		return c.db.Close()
	}
	return nil
}

// Ping tests the SQLite connection
func (c *SQLiteConnector) Ping(ctx context.Context) error {
	if c.db == nil {
		return ErrConnectionFailed
	}
	return c.db.PingContext(ctx)
}

// Query executes a SELECT query
func (c *SQLiteConnector) Query(ctx context.Context, input *QueryInput) (*QueryResult, error) {
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
func (c *SQLiteConnector) Execute(ctx context.Context, input *QueryInput) (*QueryResult, error) {
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
func (c *SQLiteConnector) GetDatabaseType() DatabaseType {
	return DatabaseTypeSQLite
}

// validateConnectionString validates the SQLite file path
func (c *SQLiteConnector) validateConnectionString(connStr string) error {
	if connStr == "" {
		return ErrInvalidConnectionString
	}

	// Special case for in-memory database
	if connStr == ":memory:" || connStr == "file::memory:?cache=shared" {
		return nil
	}

	// Validate file path to prevent directory traversal
	cleanPath := filepath.Clean(connStr)

	// Block absolute paths for security (should be relative to a configured data directory)
	if filepath.IsAbs(cleanPath) {
		return fmt.Errorf("absolute paths are not allowed for SQLite databases")
	}

	// Block parent directory references
	if strings.Contains(cleanPath, "..") {
		return fmt.Errorf("directory traversal not allowed in file path")
	}

	// Block system paths
	if strings.HasPrefix(cleanPath, "/etc") ||
		strings.HasPrefix(cleanPath, "/sys") ||
		strings.HasPrefix(cleanPath, "/proc") ||
		strings.HasPrefix(cleanPath, "/dev") {
		return fmt.Errorf("access to system directories not allowed")
	}

	return nil
}

// isSelectQuery checks if the query is a SELECT statement
func (c *SQLiteConnector) isSelectQuery(query string) bool {
	trimmed := strings.TrimSpace(strings.ToUpper(query))
	return strings.HasPrefix(trimmed, "SELECT") || strings.HasPrefix(trimmed, "WITH")
}

// normalizeResult converts byte arrays to strings for JSON serialization
func (c *SQLiteConnector) normalizeResult(result map[string]interface{}) {
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
