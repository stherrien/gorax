package connectors

import (
	"context"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"
)

// DatabaseType represents the type of database
type DatabaseType string

const (
	DatabaseTypePostgreSQL DatabaseType = "postgresql"
	DatabaseTypeMySQL      DatabaseType = "mysql"
	DatabaseTypeSQLite     DatabaseType = "sqlite"
	DatabaseTypeMongoDB    DatabaseType = "mongodb"
)

// ConnectionStatus represents the status of a database connection
type ConnectionStatus string

const (
	ConnectionStatusActive   ConnectionStatus = "active"
	ConnectionStatusInactive ConnectionStatus = "inactive"
	ConnectionStatusError    ConnectionStatus = "error"
)

// QueryType represents the type of query being executed
type QueryType string

const (
	QueryTypeSelect    QueryType = "select"
	QueryTypeInsert    QueryType = "insert"
	QueryTypeUpdate    QueryType = "update"
	QueryTypeDelete    QueryType = "delete"
	QueryTypeFind      QueryType = "find"      // MongoDB
	QueryTypeAggregate QueryType = "aggregate" // MongoDB
)

// Common errors
var (
	ErrInvalidConnectionString = errors.New("invalid connection string")
	ErrConnectionFailed        = errors.New("connection failed")
	ErrQueryTimeout            = errors.New("query timeout")
	ErrQueryFailed             = errors.New("query failed")
	ErrRowLimitExceeded        = errors.New("row limit exceeded")
	ErrUnsupportedDatabase     = errors.New("unsupported database type")
	ErrInvalidQuery            = errors.New("invalid query")
	ErrConnectionNotFound      = errors.New("connection not found")
	ErrUnauthorized            = errors.New("unauthorized access to connection")
)

// ValidationError represents a validation error
type ValidationError struct {
	Message string
}

func (e *ValidationError) Error() string {
	return e.Message
}

// JSONMap is a custom type for storing JSON in PostgreSQL
type JSONMap map[string]interface{}

// Value implements driver.Valuer for database serialization
func (j JSONMap) Value() (driver.Value, error) {
	if j == nil {
		return []byte("{}"), nil
	}
	return json.Marshal(j)
}

// Scan implements sql.Scanner for database deserialization
func (j *JSONMap) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}

	var data []byte
	switch v := value.(type) {
	case []byte:
		data = v
	case string:
		data = []byte(v)
	default:
		return errors.New("unsupported type for JSONMap")
	}

	return json.Unmarshal(data, j)
}

// DatabaseConnection represents a database connection configuration
type DatabaseConnection struct {
	ID           string           `json:"id" db:"id"`
	TenantID     string           `json:"tenant_id" db:"tenant_id"`
	Name         string           `json:"name" db:"name"`
	Description  string           `json:"description" db:"description"`
	DatabaseType DatabaseType     `json:"database_type" db:"database_type"`
	CredentialID string           `json:"credential_id" db:"credential_id"`
	Status       ConnectionStatus `json:"status" db:"status"`

	// Connection configuration (non-sensitive)
	ConnectionConfig JSONMap `json:"connection_config" db:"connection_config"`

	// Test results
	LastTestedAt   *time.Time `json:"last_tested_at,omitempty" db:"last_tested_at"`
	LastTestResult JSONMap    `json:"last_test_result,omitempty" db:"last_test_result"`

	// Usage tracking
	LastUsedAt *time.Time `json:"last_used_at,omitempty" db:"last_used_at"`
	UsageCount int64      `json:"usage_count" db:"usage_count"`
	CreatedBy  string     `json:"created_by" db:"created_by"`
	CreatedAt  time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at" db:"updated_at"`
}

// ConnectionConfig represents non-sensitive connection configuration
type ConnectionConfig struct {
	Host           string `json:"host,omitempty"`
	Port           int    `json:"port,omitempty"`
	Database       string `json:"database,omitempty"`
	SSLMode        string `json:"ssl_mode,omitempty"`
	ReadOnly       bool   `json:"read_only,omitempty"`
	MaxConnections int    `json:"max_connections,omitempty"`
	TimeoutSeconds int    `json:"timeout_seconds,omitempty"`
}

// ToMap converts ConnectionConfig to JSONMap
func (c *ConnectionConfig) ToMap() JSONMap {
	data, _ := json.Marshal(c)
	var result JSONMap
	_ = json.Unmarshal(data, &result)
	return result
}

// FromMap populates ConnectionConfig from JSONMap
func (c *ConnectionConfig) FromMap(m JSONMap) error {
	data, err := json.Marshal(m)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, c)
}

// TestResult represents the result of a connection test
type TestResult struct {
	Success   bool   `json:"success"`
	Error     string `json:"error,omitempty"`
	LatencyMS int64  `json:"latency_ms"`
	TestedAt  string `json:"tested_at"`
}

// ToMap converts TestResult to JSONMap
func (t *TestResult) ToMap() JSONMap {
	data, _ := json.Marshal(t)
	var result JSONMap
	_ = json.Unmarshal(data, &result)
	return result
}

// DatabaseConnectionQuery represents a query execution log entry
type DatabaseConnectionQuery struct {
	ID              string    `json:"id" db:"id"`
	ConnectionID    string    `json:"connection_id" db:"connection_id"`
	TenantID        string    `json:"tenant_id" db:"tenant_id"`
	WorkflowID      *string   `json:"workflow_id,omitempty" db:"workflow_id"`
	ExecutionID     *string   `json:"execution_id,omitempty" db:"execution_id"`
	QueryType       QueryType `json:"query_type" db:"query_type"`
	QueryHash       string    `json:"query_hash,omitempty" db:"query_hash"`
	RowsAffected    int       `json:"rows_affected" db:"rows_affected"`
	ExecutionTimeMS int       `json:"execution_time_ms" db:"execution_time_ms"`
	Success         bool      `json:"success" db:"success"`
	ErrorMessage    string    `json:"error_message,omitempty" db:"error_message"`
	ExecutedBy      string    `json:"executed_by" db:"executed_by"`
	ExecutedAt      time.Time `json:"executed_at" db:"executed_at"`
}

// CreateConnectionInput represents input for creating a connection
type CreateConnectionInput struct {
	Name             string           `json:"name"`
	Description      string           `json:"description"`
	DatabaseType     DatabaseType     `json:"database_type"`
	CredentialID     string           `json:"credential_id"`
	ConnectionConfig ConnectionConfig `json:"connection_config"`
}

// Validate validates CreateConnectionInput
func (c *CreateConnectionInput) Validate() error {
	if c.Name == "" {
		return &ValidationError{Message: "name is required"}
	}
	if len(c.Name) > 255 {
		return &ValidationError{Message: "name must be less than 255 characters"}
	}
	if c.DatabaseType == "" {
		return &ValidationError{Message: "database_type is required"}
	}
	if c.DatabaseType != DatabaseTypePostgreSQL &&
		c.DatabaseType != DatabaseTypeMySQL &&
		c.DatabaseType != DatabaseTypeSQLite &&
		c.DatabaseType != DatabaseTypeMongoDB {
		return &ValidationError{Message: "invalid database type"}
	}
	if c.CredentialID == "" {
		return &ValidationError{Message: "credential_id is required"}
	}

	// Validate connection config based on database type
	if c.DatabaseType != DatabaseTypeSQLite {
		if c.ConnectionConfig.Host == "" {
			return &ValidationError{Message: "host is required for non-SQLite databases"}
		}
		if c.ConnectionConfig.Port == 0 {
			return &ValidationError{Message: "port is required for non-SQLite databases"}
		}
		if c.ConnectionConfig.Database == "" {
			return &ValidationError{Message: "database name is required"}
		}
	}

	return nil
}

// UpdateConnectionInput represents input for updating a connection
type UpdateConnectionInput struct {
	Name             *string           `json:"name,omitempty"`
	Description      *string           `json:"description,omitempty"`
	Status           *ConnectionStatus `json:"status,omitempty"`
	ConnectionConfig *ConnectionConfig `json:"connection_config,omitempty"`
}

// Validate validates UpdateConnectionInput
func (u *UpdateConnectionInput) Validate() error {
	if u.Name != nil && len(*u.Name) > 255 {
		return &ValidationError{Message: "name must be less than 255 characters"}
	}
	if u.Status != nil {
		if *u.Status != ConnectionStatusActive &&
			*u.Status != ConnectionStatusInactive &&
			*u.Status != ConnectionStatusError {
			return &ValidationError{Message: "invalid status"}
		}
	}
	return nil
}

// ConnectionListFilter represents filters for listing connections
type ConnectionListFilter struct {
	DatabaseType DatabaseType     `json:"database_type,omitempty"`
	Status       ConnectionStatus `json:"status,omitempty"`
	Search       string           `json:"search,omitempty"`
}

// QueryInput represents input for executing a query
type QueryInput struct {
	Query      string                 `json:"query"`
	Parameters []interface{}          `json:"parameters,omitempty"`
	Timeout    int                    `json:"timeout,omitempty"`  // seconds, default 30
	MaxRows    int                    `json:"max_rows,omitempty"` // default 1000
	Metadata   map[string]interface{} `json:"metadata,omitempty"` // workflow context
}

// Validate validates QueryInput
func (q *QueryInput) Validate() error {
	if q.Query == "" {
		return &ValidationError{Message: "query is required"}
	}
	if q.Timeout < 0 || q.Timeout > 300 {
		return &ValidationError{Message: "timeout must be between 0 and 300 seconds"}
	}
	if q.MaxRows < 0 || q.MaxRows > 10000 {
		return &ValidationError{Message: "max_rows must be between 0 and 10000"}
	}
	return nil
}

// QueryResult represents the result of a query execution
type QueryResult struct {
	Rows         []map[string]interface{} `json:"rows,omitempty"`
	RowsAffected int                      `json:"rows_affected"`
	ExecutionMS  int64                    `json:"execution_ms"`
	Metadata     map[string]interface{}   `json:"metadata,omitempty"`
}

// Connector defines the interface for database connectors
type Connector interface {
	// Connect establishes a connection to the database
	Connect(ctx context.Context, connectionString string) error

	// Close closes the connection
	Close() error

	// Ping tests the connection
	Ping(ctx context.Context) error

	// Query executes a SELECT query and returns results
	Query(ctx context.Context, input *QueryInput) (*QueryResult, error)

	// Execute executes a query that modifies data (INSERT, UPDATE, DELETE)
	Execute(ctx context.Context, input *QueryInput) (*QueryResult, error)

	// GetDatabaseType returns the database type
	GetDatabaseType() DatabaseType
}
