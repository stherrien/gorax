package database

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/gorax/gorax/internal/database/connectors"
	"github.com/gorax/gorax/internal/executor/actions"
)

// MongoDBAction implements the Action interface for executing MongoDB operations
type MongoDBAction struct {
	connectorFactory *connectors.ConnectorFactory
}

// NewMongoDBAction creates a new MongoDB action
func NewMongoDBAction() *MongoDBAction {
	return &MongoDBAction{
		connectorFactory: connectors.NewConnectorFactory(),
	}
}

// MongoDBConfig represents the configuration for a MongoDB action
type MongoDBConfig struct {
	ConnectionString string                 `json:"connection_string"`    // MongoDB connection string from credential
	Operation        string                 `json:"operation"`            // find, insertOne, updateOne, deleteOne, etc.
	Collection       string                 `json:"collection"`           // Collection name
	Filter           map[string]interface{} `json:"filter,omitempty"`     // Filter for find/update/delete
	Document         map[string]interface{} `json:"document,omitempty"`   // Document for insert
	Update           map[string]interface{} `json:"update,omitempty"`     // Update document
	Sort             map[string]interface{} `json:"sort,omitempty"`       // Sort for find
	Projection       map[string]interface{} `json:"projection,omitempty"` // Projection for find
	Timeout          int                    `json:"timeout"`              // Timeout in seconds (default: 30)
	MaxRows          int                    `json:"max_rows,omitempty"`   // Max rows for find (default: 1000)
}

// Execute implements the Action interface
func (a *MongoDBAction) Execute(ctx context.Context, input *actions.ActionInput) (*actions.ActionOutput, error) {
	// Parse config
	configBytes, err := json.Marshal(input.Config)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal config: %w", err)
	}

	var config MongoDBConfig
	if err := json.Unmarshal(configBytes, &config); err != nil {
		return nil, fmt.Errorf("failed to parse MongoDB action config: %w", err)
	}

	// Validate config
	if err := a.validateConfig(&config); err != nil {
		return nil, err
	}

	// Create connector
	connector, err := a.connectorFactory.CreateConnector(connectors.DatabaseTypeMongoDB)
	if err != nil {
		return nil, fmt.Errorf("failed to create connector: %w", err)
	}

	// Connect to database
	if err := connector.Connect(ctx, config.ConnectionString); err != nil {
		return nil, fmt.Errorf("failed to connect to MongoDB: %w", err)
	}
	defer connector.Close()

	// Determine if this is a read or write operation
	isReadOperation := config.Operation == "find"

	// Prepare query/command
	var queryJSON string
	if isReadOperation {
		// Build query for find operation
		queryDoc := map[string]interface{}{
			"collection": config.Collection,
			"filter":     config.Filter,
		}
		if config.Sort != nil {
			queryDoc["sort"] = config.Sort
		}
		if config.Projection != nil {
			queryDoc["projection"] = config.Projection
		}

		queryBytes, err := json.Marshal(queryDoc)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal query: %w", err)
		}
		queryJSON = string(queryBytes)
	} else {
		// Build command for write operations
		commandDoc := map[string]interface{}{
			"operation":  config.Operation,
			"collection": config.Collection,
		}
		if config.Filter != nil {
			commandDoc["filter"] = config.Filter
		}
		if config.Document != nil {
			commandDoc["document"] = config.Document
		}
		if config.Update != nil {
			commandDoc["update"] = config.Update
		}

		commandBytes, err := json.Marshal(commandDoc)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal command: %w", err)
		}
		queryJSON = string(commandBytes)
	}

	// Prepare query input
	queryInput := &connectors.QueryInput{
		Query:   queryJSON,
		Timeout: config.Timeout,
		MaxRows: config.MaxRows,
		Metadata: map[string]interface{}{
			"action_type": "mongodb",
			"operation":   config.Operation,
			"collection":  config.Collection,
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

	// Execute operation
	var result *connectors.QueryResult
	if isReadOperation {
		result, err = connector.Query(ctx, queryInput)
	} else {
		result, err = connector.Execute(ctx, queryInput)
	}

	if err != nil {
		return nil, fmt.Errorf("operation failed: %w", err)
	}

	// Return output
	output := actions.NewActionOutput(result)
	if isReadOperation {
		output.WithMetadata("documents_count", len(result.Rows))
	} else {
		output.WithMetadata("modified_count", result.RowsAffected)
	}
	output.WithMetadata("execution_ms", result.ExecutionMS)

	return output, nil
}

// validateConfig validates the MongoDB configuration
func (a *MongoDBAction) validateConfig(config *MongoDBConfig) error {
	if config.ConnectionString == "" {
		return fmt.Errorf("connection_string is required")
	}

	if config.Operation == "" {
		return fmt.Errorf("operation is required")
	}

	// Validate operation type
	validOperations := []string{
		"find", "insertOne", "insertMany",
		"updateOne", "updateMany",
		"deleteOne", "deleteMany",
	}
	isValidOp := false
	for _, op := range validOperations {
		if config.Operation == op {
			isValidOp = true
			break
		}
	}
	if !isValidOp {
		return fmt.Errorf("invalid operation: %s", config.Operation)
	}

	if config.Collection == "" {
		return fmt.Errorf("collection is required")
	}

	// Validate operation-specific requirements
	switch config.Operation {
	case "find":
		// Filter is optional for find
		if config.Filter == nil {
			config.Filter = make(map[string]interface{})
		}
	case "insertOne", "insertMany":
		if config.Document == nil {
			return fmt.Errorf("document is required for insert operations")
		}
	case "updateOne", "updateMany":
		if config.Filter == nil {
			return fmt.Errorf("filter is required for update operations")
		}
		if config.Update == nil {
			return fmt.Errorf("update is required for update operations")
		}
	case "deleteOne", "deleteMany":
		if config.Filter == nil {
			return fmt.Errorf("filter is required for delete operations")
		}
	}

	// Set defaults
	if config.Timeout == 0 {
		config.Timeout = 30
	}
	if config.MaxRows == 0 {
		config.MaxRows = 1000
	}

	// Validate ranges
	if config.Timeout < 1 || config.Timeout > 300 {
		return fmt.Errorf("timeout must be between 1 and 300 seconds")
	}
	if config.MaxRows < 1 || config.MaxRows > 10000 {
		return fmt.Errorf("max_rows must be between 1 and 10000")
	}

	return nil
}
