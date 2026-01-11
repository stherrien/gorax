package database

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/gorax/gorax/internal/database/connectors"
	"github.com/gorax/gorax/internal/executor/actions"
)

// SQLExecuteAction implements the Action interface for executing SQL data modification queries
type SQLExecuteAction struct {
	connectorFactory *connectors.ConnectorFactory
}

// NewSQLExecuteAction creates a new SQL execute action
func NewSQLExecuteAction() *SQLExecuteAction {
	return &SQLExecuteAction{
		connectorFactory: connectors.NewConnectorFactory(),
	}
}

// SQLExecuteConfig represents the configuration for an SQL execute action
type SQLExecuteConfig struct {
	ConnectionString string        `json:"connection_string"` // Connection string from credential
	DatabaseType     string        `json:"database_type"`     // postgresql, mysql, sqlite
	Query            string        `json:"query"`             // SQL query (INSERT, UPDATE, DELETE)
	Parameters       []interface{} `json:"parameters"`        // Query parameters
	Timeout          int           `json:"timeout"`           // Timeout in seconds (default: 30)
}

// Execute implements the Action interface
func (a *SQLExecuteAction) Execute(ctx context.Context, input *actions.ActionInput) (*actions.ActionOutput, error) {
	// Parse config
	configBytes, err := json.Marshal(input.Config)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal config: %w", err)
	}

	var config SQLExecuteConfig
	if err := json.Unmarshal(configBytes, &config); err != nil {
		return nil, fmt.Errorf("failed to parse SQL execute action config: %w", err)
	}

	// Validate config
	if err := a.validateConfig(&config); err != nil {
		return nil, err
	}

	// Create connector
	connector, err := a.connectorFactory.CreateConnector(connectors.DatabaseType(config.DatabaseType))
	if err != nil {
		return nil, fmt.Errorf("failed to create connector: %w", err)
	}

	// Connect to database
	if err := connector.Connect(ctx, config.ConnectionString); err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}
	defer connector.Close()

	// Prepare query input
	queryInput := &connectors.QueryInput{
		Query:      config.Query,
		Parameters: config.Parameters,
		Timeout:    config.Timeout,
		Metadata: map[string]interface{}{
			"action_type":   "sql_execute",
			"database_type": config.DatabaseType,
		},
	}

	// Merge workflow context metadata
	if input.Context != nil {
		if workflowID, ok := input.Context["workflow_id"].(string); ok {
			queryInput.Metadata["workflow_id"] = workflowID
		}
		if executionID, ok := input.Context["execution_id"].(string); ok {
			queryInput.Metadata["execution_id"] = executionID
		}
	}

	// Execute query
	result, err := connector.Execute(ctx, queryInput)
	if err != nil {
		return nil, fmt.Errorf("query execution failed: %w", err)
	}

	// Return output
	output := actions.NewActionOutput(result)
	output.WithMetadata("rows_affected", result.RowsAffected)
	output.WithMetadata("execution_ms", result.ExecutionMS)

	return output, nil
}

// validateConfig validates the SQL execute configuration
func (a *SQLExecuteAction) validateConfig(config *SQLExecuteConfig) error {
	if config.ConnectionString == "" {
		return fmt.Errorf("connection_string is required")
	}

	if config.DatabaseType == "" {
		return fmt.Errorf("database_type is required")
	}

	// Validate database type
	dbType := connectors.DatabaseType(config.DatabaseType)
	if dbType != connectors.DatabaseTypePostgreSQL &&
		dbType != connectors.DatabaseTypeMySQL &&
		dbType != connectors.DatabaseTypeSQLite {
		return fmt.Errorf("invalid database_type for SQL queries: %s (must be postgresql, mysql, or sqlite)", config.DatabaseType)
	}

	if config.Query == "" {
		return fmt.Errorf("query is required")
	}

	// Set defaults
	if config.Timeout == 0 {
		config.Timeout = 30
	}

	// Validate ranges
	if config.Timeout < 1 || config.Timeout > 300 {
		return fmt.Errorf("timeout must be between 1 and 300 seconds")
	}

	return nil
}
