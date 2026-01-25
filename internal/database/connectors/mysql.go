package connectors

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql" // MySQL driver
	"github.com/jmoiron/sqlx"
)

// MySQLConnector implements the Connector interface for MySQL
type MySQLConnector struct {
	db *sqlx.DB
}

// NewMySQLConnector creates a new MySQL connector
func NewMySQLConnector() *MySQLConnector {
	return &MySQLConnector{}
}

// Connect establishes a connection to MySQL
func (c *MySQLConnector) Connect(ctx context.Context, connectionString string) error {
	// Validate and sanitize connection string
	if err := c.validateConnectionString(connectionString); err != nil {
		return fmt.Errorf("invalid connection string: %w", err)
	}

	// Open connection
	db, err := sqlx.ConnectContext(ctx, "mysql", connectionString)
	if err != nil {
		return fmt.Errorf("failed to connect to MySQL: %w", err)
	}

	// Set connection pool settings
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)
	db.SetConnMaxIdleTime(1 * time.Minute)

	c.db = db
	return nil
}

// Close closes the MySQL connection
func (c *MySQLConnector) Close() error {
	if c.db != nil {
		return c.db.Close()
	}
	return nil
}

// Ping tests the MySQL connection
func (c *MySQLConnector) Ping(ctx context.Context) error {
	if c.db == nil {
		return ErrConnectionFailed
	}
	return c.db.PingContext(ctx)
}

// Query executes a SELECT query
func (c *MySQLConnector) Query(ctx context.Context, input *QueryInput) (*QueryResult, error) {
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
func (c *MySQLConnector) Execute(ctx context.Context, input *QueryInput) (*QueryResult, error) {
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
func (c *MySQLConnector) GetDatabaseType() DatabaseType {
	return DatabaseTypeMySQL
}

// validateConnectionString validates the MySQL connection string
func (c *MySQLConnector) validateConnectionString(connStr string) error {
	if connStr == "" {
		return ErrInvalidConnectionString
	}

	// Parse connection string (format: username:password@tcp(host:port)/database)
	if !strings.Contains(connStr, "@tcp(") && !strings.Contains(connStr, "@unix(") {
		return fmt.Errorf("%w: MySQL connection string must contain @tcp() or @unix()", ErrInvalidConnectionString)
	}

	// Extract host for validation
	host := c.extractHost(connStr)
	if host != "" {
		if err := c.validateHost(host); err != nil {
			return err
		}
	}

	return nil
}

// extractHost extracts the host from MySQL connection string
func (c *MySQLConnector) extractHost(connStr string) string {
	// Format: username:password@tcp(host:port)/database
	startIdx := strings.Index(connStr, "@tcp(")
	if startIdx == -1 {
		return ""
	}

	startIdx += 5 // Skip "@tcp("
	endIdx := strings.Index(connStr[startIdx:], ")")
	if endIdx == -1 {
		return ""
	}

	hostPort := connStr[startIdx : startIdx+endIdx]
	// Split host:port
	parts := strings.Split(hostPort, ":")
	if len(parts) > 0 {
		return parts[0]
	}

	return ""
}

// validateHost validates the database host to prevent SSRF
func (c *MySQLConnector) validateHost(host string) error {
	// Parse if URL encoded
	if decodedHost, err := url.QueryUnescape(host); err == nil {
		host = decodedHost
	}

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
func (c *MySQLConnector) isSelectQuery(query string) bool {
	trimmed := strings.TrimSpace(strings.ToUpper(query))
	return strings.HasPrefix(trimmed, "SELECT") || strings.HasPrefix(trimmed, "WITH")
}

// normalizeResult converts byte arrays to strings for JSON serialization
func (c *MySQLConnector) normalizeResult(result map[string]interface{}) {
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
