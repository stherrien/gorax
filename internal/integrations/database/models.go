package database

import "errors"

// Common errors
var (
	ErrInvalidConnectionString = errors.New("invalid connection string")
	ErrQueryFailed             = errors.New("query execution failed")
	ErrNoRowsAffected          = errors.New("no rows affected")
	ErrInvalidQuery            = errors.New("invalid query")
	ErrInvalidParams           = errors.New("invalid parameters")
	ErrConnectionFailed        = errors.New("connection failed")
	ErrTransactionFailed       = errors.New("transaction failed")
	ErrInvalidCollection       = errors.New("invalid collection name")
	ErrInvalidFilter           = errors.New("invalid filter")
	ErrInvalidDocument         = errors.New("invalid document")
)

// QueryConfig represents configuration for SQL query actions
type QueryConfig struct {
	Query      string                 `json:"query"`
	Parameters []interface{}          `json:"parameters,omitempty"`
	Timeout    int                    `json:"timeout,omitempty"` // seconds
}

// Validate checks if query config is valid
func (c *QueryConfig) Validate() error {
	if c.Query == "" {
		return ErrInvalidQuery
	}
	return nil
}

// StatementConfig represents configuration for SQL statement actions (INSERT/UPDATE/DELETE)
type StatementConfig struct {
	Statement  string                 `json:"statement"`
	Parameters []interface{}          `json:"parameters,omitempty"`
	Timeout    int                    `json:"timeout,omitempty"` // seconds
}

// Validate checks if statement config is valid
func (c *StatementConfig) Validate() error {
	if c.Statement == "" {
		return ErrInvalidQuery
	}
	return nil
}

// TransactionConfig represents configuration for transaction actions
type TransactionConfig struct {
	Statements []TransactionStatement `json:"statements"`
	Timeout    int                    `json:"timeout,omitempty"` // seconds
}

// TransactionStatement represents a single statement in a transaction
type TransactionStatement struct {
	Statement  string        `json:"statement"`
	Parameters []interface{} `json:"parameters,omitempty"`
}

// Validate checks if transaction config is valid
func (c *TransactionConfig) Validate() error {
	if len(c.Statements) == 0 {
		return errors.New("at least one statement is required")
	}
	for _, stmt := range c.Statements {
		if stmt.Statement == "" {
			return ErrInvalidQuery
		}
	}
	return nil
}

// QueryResult represents the result of a query
type QueryResult struct {
	Rows         []map[string]interface{} `json:"rows"`
	RowCount     int                      `json:"row_count"`
	ColumnNames  []string                 `json:"column_names,omitempty"`
}

// StatementResult represents the result of a statement execution
type StatementResult struct {
	RowsAffected int64  `json:"rows_affected"`
	LastInsertID int64  `json:"last_insert_id,omitempty"`
}

// TransactionResult represents the result of a transaction
type TransactionResult struct {
	Committed      bool              `json:"committed"`
	StatementsRun  int               `json:"statements_run"`
	TotalAffected  int64             `json:"total_affected"`
}

// MongoFindConfig represents configuration for MongoDB find operations
type MongoFindConfig struct {
	Collection string                 `json:"collection"`
	Filter     map[string]interface{} `json:"filter,omitempty"`
	Projection map[string]interface{} `json:"projection,omitempty"`
	Sort       map[string]interface{} `json:"sort,omitempty"`
	Limit      int64                  `json:"limit,omitempty"`
	Skip       int64                  `json:"skip,omitempty"`
	Timeout    int                    `json:"timeout,omitempty"` // seconds
}

// Validate checks if find config is valid
func (c *MongoFindConfig) Validate() error {
	if c.Collection == "" {
		return ErrInvalidCollection
	}
	return nil
}

// MongoInsertConfig represents configuration for MongoDB insert operations
type MongoInsertConfig struct {
	Collection string                   `json:"collection"`
	Documents  []map[string]interface{} `json:"documents"`
	Timeout    int                      `json:"timeout,omitempty"` // seconds
}

// Validate checks if insert config is valid
func (c *MongoInsertConfig) Validate() error {
	if c.Collection == "" {
		return ErrInvalidCollection
	}
	if len(c.Documents) == 0 {
		return errors.New("at least one document is required")
	}
	return nil
}

// MongoUpdateConfig represents configuration for MongoDB update operations
type MongoUpdateConfig struct {
	Collection string                 `json:"collection"`
	Filter     map[string]interface{} `json:"filter"`
	Update     map[string]interface{} `json:"update"`
	Upsert     bool                   `json:"upsert,omitempty"`
	Multi      bool                   `json:"multi,omitempty"` // Update multiple documents
	Timeout    int                    `json:"timeout,omitempty"` // seconds
}

// Validate checks if update config is valid
func (c *MongoUpdateConfig) Validate() error {
	if c.Collection == "" {
		return ErrInvalidCollection
	}
	if c.Filter == nil || len(c.Filter) == 0 {
		return ErrInvalidFilter
	}
	if c.Update == nil || len(c.Update) == 0 {
		return errors.New("update document is required")
	}
	return nil
}

// MongoDeleteConfig represents configuration for MongoDB delete operations
type MongoDeleteConfig struct {
	Collection string                 `json:"collection"`
	Filter     map[string]interface{} `json:"filter"`
	Multi      bool                   `json:"multi,omitempty"` // Delete multiple documents
	Timeout    int                    `json:"timeout,omitempty"` // seconds
}

// Validate checks if delete config is valid
func (c *MongoDeleteConfig) Validate() error {
	if c.Collection == "" {
		return ErrInvalidCollection
	}
	if c.Filter == nil || len(c.Filter) == 0 {
		return ErrInvalidFilter
	}
	return nil
}

// MongoAggregateConfig represents configuration for MongoDB aggregation operations
type MongoAggregateConfig struct {
	Collection string                   `json:"collection"`
	Pipeline   []map[string]interface{} `json:"pipeline"`
	Timeout    int                      `json:"timeout,omitempty"` // seconds
}

// Validate checks if aggregate config is valid
func (c *MongoAggregateConfig) Validate() error {
	if c.Collection == "" {
		return ErrInvalidCollection
	}
	if len(c.Pipeline) == 0 {
		return errors.New("aggregation pipeline is required")
	}
	return nil
}

// MongoFindResult represents the result of a MongoDB find operation
type MongoFindResult struct {
	Documents []map[string]interface{} `json:"documents"`
	Count     int                      `json:"count"`
}

// MongoInsertResult represents the result of a MongoDB insert operation
type MongoInsertResult struct {
	InsertedCount int                    `json:"inserted_count"`
	InsertedIDs   []interface{}          `json:"inserted_ids,omitempty"`
}

// MongoUpdateResult represents the result of a MongoDB update operation
type MongoUpdateResult struct {
	MatchedCount  int64       `json:"matched_count"`
	ModifiedCount int64       `json:"modified_count"`
	UpsertedCount int64       `json:"upserted_count"`
	UpsertedID    interface{} `json:"upserted_id,omitempty"`
}

// MongoDeleteResult represents the result of a MongoDB delete operation
type MongoDeleteResult struct {
	DeletedCount int64 `json:"deleted_count"`
}

// MongoAggregateResult represents the result of a MongoDB aggregation operation
type MongoAggregateResult struct {
	Documents []map[string]interface{} `json:"documents"`
	Count     int                      `json:"count"`
}
