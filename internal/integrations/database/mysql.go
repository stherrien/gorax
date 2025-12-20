package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"

	"github.com/gorax/gorax/internal/credential"
	"github.com/gorax/gorax/internal/executor/actions"
)

// MySQLQueryAction executes SELECT queries on MySQL
type MySQLQueryAction struct {
	credentialService credential.Service
	dbFactory         func(connStr string) (*sql.DB, error)
}

// NewMySQLQueryAction creates a new MySQL query action
func NewMySQLQueryAction(credentialService credential.Service) *MySQLQueryAction {
	return &MySQLQueryAction{
		credentialService: credentialService,
		dbFactory:         newMySQLDB,
	}
}

// Execute implements the Action interface
func (a *MySQLQueryAction) Execute(ctx context.Context, input *actions.ActionInput) (*actions.ActionOutput, error) {
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

// MySQLStatementAction executes INSERT/UPDATE/DELETE statements on MySQL
type MySQLStatementAction struct {
	credentialService credential.Service
	dbFactory         func(connStr string) (*sql.DB, error)
}

// NewMySQLStatementAction creates a new MySQL statement action
func NewMySQLStatementAction(credentialService credential.Service) *MySQLStatementAction {
	return &MySQLStatementAction{
		credentialService: credentialService,
		dbFactory:         newMySQLDB,
	}
}

// Execute implements the Action interface
func (a *MySQLStatementAction) Execute(ctx context.Context, input *actions.ActionInput) (*actions.ActionOutput, error) {
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

	// Get last insert ID
	lastInsertID, _ := sqlResult.LastInsertId()

	result := &StatementResult{
		RowsAffected: rowsAffected,
		LastInsertID: lastInsertID,
	}

	// Create output
	output := actions.NewActionOutput(result)
	output.WithMetadata("rows_affected", rowsAffected)
	if lastInsertID > 0 {
		output.WithMetadata("last_insert_id", lastInsertID)
	}

	return output, nil
}

// MySQLTransactionAction executes multiple statements in a transaction
type MySQLTransactionAction struct {
	credentialService credential.Service
	dbFactory         func(connStr string) (*sql.DB, error)
}

// NewMySQLTransactionAction creates a new MySQL transaction action
func NewMySQLTransactionAction(credentialService credential.Service) *MySQLTransactionAction {
	return &MySQLTransactionAction{
		credentialService: credentialService,
		dbFactory:         newMySQLDB,
	}
}

// Execute implements the Action interface
func (a *MySQLTransactionAction) Execute(ctx context.Context, input *actions.ActionInput) (*actions.ActionOutput, error) {
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

// newMySQLDB creates a new MySQL database connection with proper pooling
func newMySQLDB(connStr string) (*sql.DB, error) {
	db, err := sql.Open("mysql", connStr)
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
