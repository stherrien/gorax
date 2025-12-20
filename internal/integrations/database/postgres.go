package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"

	"github.com/gorax/gorax/internal/credential"
	"github.com/gorax/gorax/internal/executor/actions"
)

const (
	// DefaultMaxOpenConns is the default maximum number of open connections
	DefaultMaxOpenConns = 10

	// DefaultMaxIdleConns is the default maximum number of idle connections
	DefaultMaxIdleConns = 5

	// DefaultConnMaxLifetime is the default maximum lifetime of a connection
	DefaultConnMaxLifetime = time.Hour
)

// PostgresQueryAction executes SELECT queries on PostgreSQL
type PostgresQueryAction struct {
	credentialService credential.Service
	dbFactory         func(connStr string) (*sql.DB, error)
}

// NewPostgresQueryAction creates a new PostgreSQL query action
func NewPostgresQueryAction(credentialService credential.Service) *PostgresQueryAction {
	return &PostgresQueryAction{
		credentialService: credentialService,
		dbFactory:         newPostgresDB,
	}
}

// Execute implements the Action interface
func (a *PostgresQueryAction) Execute(ctx context.Context, input *actions.ActionInput) (*actions.ActionOutput, error) {
	// Parse config
	config, ok := input.Config.(QueryConfig)
	if !ok {
		return nil, fmt.Errorf("invalid config type: expected QueryConfig")
	}

	// Validate config
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	// Get connection string from credentials
	connStr, err := getConnectionString(ctx, a.credentialService, input.Context)
	if err != nil {
		return nil, err
	}

	// Create database connection
	db, err := a.dbFactory(connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}
	defer db.Close()

	// Apply timeout if specified
	queryCtx := ctx
	if config.Timeout > 0 {
		var cancel context.CancelFunc
		queryCtx, cancel = context.WithTimeout(ctx, time.Duration(config.Timeout)*time.Second)
		defer cancel()
	}

	// Execute query
	rows, err := db.QueryContext(queryCtx, config.Query, config.Parameters...)
	if err != nil {
		return nil, fmt.Errorf("query execution failed: %w", err)
	}
	defer rows.Close()

	// Get column names
	columns, err := rows.Columns()
	if err != nil {
		return nil, fmt.Errorf("failed to get columns: %w", err)
	}

	// Scan results
	result := &QueryResult{
		Rows:        make([]map[string]interface{}, 0),
		ColumnNames: columns,
	}

	for rows.Next() {
		// Create slice for row values
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		// Scan row
		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		// Convert to map
		rowMap := make(map[string]interface{})
		for i, col := range columns {
			val := values[i]
			// Convert []byte to string for better JSON serialization
			if b, ok := val.([]byte); ok {
				rowMap[col] = string(b)
			} else {
				rowMap[col] = val
			}
		}
		result.Rows = append(result.Rows, rowMap)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	result.RowCount = len(result.Rows)

	// Create output
	output := actions.NewActionOutput(result)
	output.WithMetadata("row_count", result.RowCount)
	output.WithMetadata("column_count", len(columns))

	return output, nil
}

// PostgresStatementAction executes INSERT/UPDATE/DELETE statements on PostgreSQL
type PostgresStatementAction struct {
	credentialService credential.Service
	dbFactory         func(connStr string) (*sql.DB, error)
}

// NewPostgresStatementAction creates a new PostgreSQL statement action
func NewPostgresStatementAction(credentialService credential.Service) *PostgresStatementAction {
	return &PostgresStatementAction{
		credentialService: credentialService,
		dbFactory:         newPostgresDB,
	}
}

// Execute implements the Action interface
func (a *PostgresStatementAction) Execute(ctx context.Context, input *actions.ActionInput) (*actions.ActionOutput, error) {
	// Parse config
	config, ok := input.Config.(StatementConfig)
	if !ok {
		return nil, fmt.Errorf("invalid config type: expected StatementConfig")
	}

	// Validate config
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	// Get connection string from credentials
	connStr, err := getConnectionString(ctx, a.credentialService, input.Context)
	if err != nil {
		return nil, err
	}

	// Create database connection
	db, err := a.dbFactory(connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}
	defer db.Close()

	// Apply timeout if specified
	execCtx := ctx
	if config.Timeout > 0 {
		var cancel context.CancelFunc
		execCtx, cancel = context.WithTimeout(ctx, time.Duration(config.Timeout)*time.Second)
		defer cancel()
	}

	// Execute statement
	sqlResult, err := db.ExecContext(execCtx, config.Statement, config.Parameters...)
	if err != nil {
		return nil, fmt.Errorf("statement execution failed: %w", err)
	}

	// Get rows affected
	rowsAffected, err := sqlResult.RowsAffected()
	if err != nil {
		return nil, fmt.Errorf("failed to get rows affected: %w", err)
	}

	// Get last insert ID (may not be supported by all drivers)
	lastInsertID, _ := sqlResult.LastInsertId()

	result := &StatementResult{
		RowsAffected: rowsAffected,
		LastInsertID: lastInsertID,
	}

	// Create output
	output := actions.NewActionOutput(result)
	output.WithMetadata("rows_affected", rowsAffected)

	return output, nil
}

// PostgresTransactionAction executes multiple statements in a transaction
type PostgresTransactionAction struct {
	credentialService credential.Service
	dbFactory         func(connStr string) (*sql.DB, error)
}

// NewPostgresTransactionAction creates a new PostgreSQL transaction action
func NewPostgresTransactionAction(credentialService credential.Service) *PostgresTransactionAction {
	return &PostgresTransactionAction{
		credentialService: credentialService,
		dbFactory:         newPostgresDB,
	}
}

// Execute implements the Action interface
func (a *PostgresTransactionAction) Execute(ctx context.Context, input *actions.ActionInput) (*actions.ActionOutput, error) {
	// Parse config
	config, ok := input.Config.(TransactionConfig)
	if !ok {
		return nil, fmt.Errorf("invalid config type: expected TransactionConfig")
	}

	// Validate config
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	// Get connection string from credentials
	connStr, err := getConnectionString(ctx, a.credentialService, input.Context)
	if err != nil {
		return nil, err
	}

	// Create database connection
	db, err := a.dbFactory(connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}
	defer db.Close()

	// Apply timeout if specified
	txCtx := ctx
	if config.Timeout > 0 {
		var cancel context.CancelFunc
		txCtx, cancel = context.WithTimeout(ctx, time.Duration(config.Timeout)*time.Second)
		defer cancel()
	}

	// Begin transaction
	tx, err := db.BeginTx(txCtx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}

	// Track results
	result := &TransactionResult{
		Committed:     false,
		StatementsRun: 0,
		TotalAffected: 0,
	}

	// Execute statements
	for i, stmt := range config.Statements {
		sqlResult, err := tx.ExecContext(txCtx, stmt.Statement, stmt.Parameters...)
		if err != nil {
			// Rollback on error
			if rbErr := tx.Rollback(); rbErr != nil {
				return nil, fmt.Errorf("statement %d failed: %w (rollback error: %v)", i, err, rbErr)
			}
			return nil, fmt.Errorf("statement %d failed: %w (transaction rolled back)", i, err)
		}

		rowsAffected, _ := sqlResult.RowsAffected()
		result.TotalAffected += rowsAffected
		result.StatementsRun++
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	result.Committed = true

	// Create output
	output := actions.NewActionOutput(result)
	output.WithMetadata("statements_run", result.StatementsRun)
	output.WithMetadata("total_affected", result.TotalAffected)

	return output, nil
}

// newPostgresDB creates a new PostgreSQL database connection with proper pooling
func newPostgresDB(connStr string) (*sql.DB, error) {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(DefaultMaxOpenConns)
	db.SetMaxIdleConns(DefaultMaxIdleConns)
	db.SetConnMaxLifetime(DefaultConnMaxLifetime)

	// Verify connection
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return db, nil
}

// getConnectionString retrieves and validates connection string from credentials
func getConnectionString(ctx context.Context, credService credential.Service, inputCtx map[string]interface{}) (string, error) {
	// Extract tenant_id
	tenantID, err := extractString(inputCtx, "env.tenant_id")
	if err != nil {
		return "", fmt.Errorf("tenant_id is required in context: %w", err)
	}

	// Extract credential_id
	credentialID, err := extractString(inputCtx, "credential_id")
	if err != nil {
		return "", fmt.Errorf("credential_id is required in context: %w", err)
	}

	// Retrieve and decrypt credential
	decryptedCred, err := credService.GetValue(ctx, tenantID, credentialID, "system")
	if err != nil {
		return "", fmt.Errorf("failed to retrieve credential: %w", err)
	}

	// Extract connection_string from credential
	connStr, ok := decryptedCred.Value["connection_string"].(string)
	if !ok || connStr == "" {
		return "", fmt.Errorf("connection_string not found in credential")
	}

	return connStr, nil
}

// extractString extracts a string value from a nested map using dot notation
func extractString(data map[string]interface{}, path string) (string, error) {
	keys := parsePath(path)
	current := data

	for i, key := range keys {
		if i == len(keys)-1 {
			// Last key - should be the value
			if val, ok := current[key]; ok {
				if str, ok := val.(string); ok {
					return str, nil
				}
				return "", fmt.Errorf("value at '%s' is not a string", path)
			}
			return "", fmt.Errorf("key '%s' not found in context", path)
		}

		// Intermediate key - should be a map
		if val, ok := current[key]; ok {
			if m, ok := val.(map[string]interface{}); ok {
				current = m
			} else {
				return "", fmt.Errorf("value at '%s' is not a map", key)
			}
		} else {
			return "", fmt.Errorf("key '%s' not found in context", key)
		}
	}

	return "", fmt.Errorf("failed to extract value from path '%s'", path)
}

// parsePath splits a dot-notation path into keys
func parsePath(path string) []string {
	result := []string{}
	current := ""

	for _, ch := range path {
		if ch == '.' {
			if current != "" {
				result = append(result, current)
				current = ""
			}
		} else {
			current += string(ch)
		}
	}

	if current != "" {
		result = append(result, current)
	}

	return result
}
