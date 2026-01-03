package connectors

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MongoDBConnector implements the Connector interface for MongoDB
type MongoDBConnector struct {
	client   *mongo.Client
	database string
}

// NewMongoDBConnector creates a new MongoDB connector
func NewMongoDBConnector() *MongoDBConnector {
	return &MongoDBConnector{}
}

// Connect establishes a connection to MongoDB
func (c *MongoDBConnector) Connect(ctx context.Context, connectionString string) error {
	// Validate connection string
	if err := c.validateConnectionString(connectionString); err != nil {
		return fmt.Errorf("invalid connection string: %w", err)
	}

	// Extract database name from connection string
	c.database = c.extractDatabase(connectionString)
	if c.database == "" {
		return fmt.Errorf("%w: database name required in connection string", ErrInvalidConnectionString)
	}

	// Set client options with timeouts
	clientOptions := options.Client().
		ApplyURI(connectionString).
		SetMaxPoolSize(25).
		SetMinPoolSize(5).
		SetMaxConnIdleTime(5 * time.Minute).
		SetConnectTimeout(10 * time.Second).
		SetServerSelectionTimeout(10 * time.Second)

	// Connect to MongoDB
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	// Ping to verify connection
	if err := client.Ping(ctx, nil); err != nil {
		_ = client.Disconnect(ctx)
		return fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	c.client = client
	return nil
}

// Close closes the MongoDB connection
func (c *MongoDBConnector) Close() error {
	if c.client != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		return c.client.Disconnect(ctx)
	}
	return nil
}

// Ping tests the MongoDB connection
func (c *MongoDBConnector) Ping(ctx context.Context) error {
	if c.client == nil {
		return ErrConnectionFailed
	}
	return c.client.Ping(ctx, nil)
}

// Query executes a MongoDB find query
func (c *MongoDBConnector) Query(ctx context.Context, input *QueryInput) (*QueryResult, error) {
	if c.client == nil {
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
	maxRows := int64(1000)
	if input.MaxRows > 0 {
		maxRows = int64(input.MaxRows)
	}

	// Parse MongoDB query
	// Expected format: {"collection": "users", "filter": {"status": "active"}, "sort": {"created_at": -1}}
	var queryDoc struct {
		Collection string                 `json:"collection"`
		Filter     map[string]interface{} `json:"filter"`
		Sort       map[string]interface{} `json:"sort,omitempty"`
		Projection map[string]interface{} `json:"projection,omitempty"`
	}

	if err := json.Unmarshal([]byte(input.Query), &queryDoc); err != nil {
		return nil, fmt.Errorf("%w: invalid MongoDB query format: %v", ErrInvalidQuery, err)
	}

	if queryDoc.Collection == "" {
		return nil, fmt.Errorf("%w: collection is required", ErrInvalidQuery)
	}

	// Get collection
	collection := c.client.Database(c.database).Collection(queryDoc.Collection)

	// Build find options
	findOptions := options.Find().SetLimit(maxRows)
	if queryDoc.Sort != nil {
		findOptions.SetSort(queryDoc.Sort)
	}
	if queryDoc.Projection != nil {
		findOptions.SetProjection(queryDoc.Projection)
	}

	// Convert filter to BSON
	filterBSON, err := c.toBSON(queryDoc.Filter)
	if err != nil {
		return nil, fmt.Errorf("failed to convert filter to BSON: %w", err)
	}

	// Execute query with timing
	startTime := time.Now()
	cursor, err := collection.Find(queryCtx, filterBSON, findOptions)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrQueryFailed, err)
	}
	defer cursor.Close(queryCtx)

	// Fetch results
	results := make([]map[string]interface{}, 0)
	if err := cursor.All(queryCtx, &results); err != nil {
		return nil, fmt.Errorf("failed to decode results: %w", err)
	}

	executionTime := time.Since(startTime)

	return &QueryResult{
		Rows:         results,
		RowsAffected: len(results),
		ExecutionMS:  executionTime.Milliseconds(),
		Metadata:     input.Metadata,
	}, nil
}

// Execute executes a MongoDB command that modifies data
func (c *MongoDBConnector) Execute(ctx context.Context, input *QueryInput) (*QueryResult, error) {
	if c.client == nil {
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

	// Parse MongoDB command
	// Expected format: {"operation": "insertOne|updateOne|deleteOne", "collection": "users", "document": {...}, "filter": {...}}
	var commandDoc struct {
		Operation  string                 `json:"operation"`
		Collection string                 `json:"collection"`
		Document   map[string]interface{} `json:"document,omitempty"`
		Filter     map[string]interface{} `json:"filter,omitempty"`
		Update     map[string]interface{} `json:"update,omitempty"`
	}

	if err := json.Unmarshal([]byte(input.Query), &commandDoc); err != nil {
		return nil, fmt.Errorf("%w: invalid MongoDB command format: %v", ErrInvalidQuery, err)
	}

	if commandDoc.Collection == "" {
		return nil, fmt.Errorf("%w: collection is required", ErrInvalidQuery)
	}

	// Get collection
	collection := c.client.Database(c.database).Collection(commandDoc.Collection)

	// Execute command with timing
	startTime := time.Now()
	var rowsAffected int64

	switch commandDoc.Operation {
	case "insertOne":
		if commandDoc.Document == nil {
			return nil, fmt.Errorf("%w: document is required for insertOne", ErrInvalidQuery)
		}
		docBSON, err := c.toBSON(commandDoc.Document)
		if err != nil {
			return nil, fmt.Errorf("failed to convert document to BSON: %w", err)
		}
		_, err = collection.InsertOne(queryCtx, docBSON)
		if err != nil {
			return nil, fmt.Errorf("%w: %v", ErrQueryFailed, err)
		}
		rowsAffected = 1

	case "insertMany":
		if commandDoc.Document == nil {
			return nil, fmt.Errorf("%w: document is required for insertMany", ErrInvalidQuery)
		}
		// Document should be an array
		docs, ok := commandDoc.Document["documents"].([]interface{})
		if !ok {
			return nil, fmt.Errorf("%w: documents array is required for insertMany", ErrInvalidQuery)
		}
		result, err := collection.InsertMany(queryCtx, docs)
		if err != nil {
			return nil, fmt.Errorf("%w: %v", ErrQueryFailed, err)
		}
		rowsAffected = int64(len(result.InsertedIDs))

	case "updateOne":
		if commandDoc.Filter == nil || commandDoc.Update == nil {
			return nil, fmt.Errorf("%w: filter and update are required for updateOne", ErrInvalidQuery)
		}
		filterBSON, err := c.toBSON(commandDoc.Filter)
		if err != nil {
			return nil, fmt.Errorf("failed to convert filter to BSON: %w", err)
		}
		updateBSON, err := c.toBSON(commandDoc.Update)
		if err != nil {
			return nil, fmt.Errorf("failed to convert update to BSON: %w", err)
		}
		result, err := collection.UpdateOne(queryCtx, filterBSON, updateBSON)
		if err != nil {
			return nil, fmt.Errorf("%w: %v", ErrQueryFailed, err)
		}
		rowsAffected = result.ModifiedCount

	case "updateMany":
		if commandDoc.Filter == nil || commandDoc.Update == nil {
			return nil, fmt.Errorf("%w: filter and update are required for updateMany", ErrInvalidQuery)
		}
		filterBSON, err := c.toBSON(commandDoc.Filter)
		if err != nil {
			return nil, fmt.Errorf("failed to convert filter to BSON: %w", err)
		}
		updateBSON, err := c.toBSON(commandDoc.Update)
		if err != nil {
			return nil, fmt.Errorf("failed to convert update to BSON: %w", err)
		}
		result, err := collection.UpdateMany(queryCtx, filterBSON, updateBSON)
		if err != nil {
			return nil, fmt.Errorf("%w: %v", ErrQueryFailed, err)
		}
		rowsAffected = result.ModifiedCount

	case "deleteOne":
		if commandDoc.Filter == nil {
			return nil, fmt.Errorf("%w: filter is required for deleteOne", ErrInvalidQuery)
		}
		filterBSON, err := c.toBSON(commandDoc.Filter)
		if err != nil {
			return nil, fmt.Errorf("failed to convert filter to BSON: %w", err)
		}
		result, err := collection.DeleteOne(queryCtx, filterBSON)
		if err != nil {
			return nil, fmt.Errorf("%w: %v", ErrQueryFailed, err)
		}
		rowsAffected = result.DeletedCount

	case "deleteMany":
		if commandDoc.Filter == nil {
			return nil, fmt.Errorf("%w: filter is required for deleteMany", ErrInvalidQuery)
		}
		filterBSON, err := c.toBSON(commandDoc.Filter)
		if err != nil {
			return nil, fmt.Errorf("failed to convert filter to BSON: %w", err)
		}
		result, err := collection.DeleteMany(queryCtx, filterBSON)
		if err != nil {
			return nil, fmt.Errorf("%w: %v", ErrQueryFailed, err)
		}
		rowsAffected = result.DeletedCount

	default:
		return nil, fmt.Errorf("%w: unsupported operation: %s", ErrInvalidQuery, commandDoc.Operation)
	}

	executionTime := time.Since(startTime)

	return &QueryResult{
		RowsAffected: int(rowsAffected),
		ExecutionMS:  executionTime.Milliseconds(),
		Metadata:     input.Metadata,
	}, nil
}

// GetDatabaseType returns the database type
func (c *MongoDBConnector) GetDatabaseType() DatabaseType {
	return DatabaseTypeMongoDB
}

// validateConnectionString validates the MongoDB connection string
func (c *MongoDBConnector) validateConnectionString(connStr string) error {
	if connStr == "" {
		return ErrInvalidConnectionString
	}

	// Check for basic validity (mongodb:// or mongodb+srv://)
	if !strings.HasPrefix(connStr, "mongodb://") && !strings.HasPrefix(connStr, "mongodb+srv://") {
		return fmt.Errorf("%w: must start with mongodb:// or mongodb+srv://", ErrInvalidConnectionString)
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

// extractHost extracts the host from MongoDB connection string
func (c *MongoDBConnector) extractHost(connStr string) string {
	// Format: mongodb://username:password@host:port/database
	// or mongodb+srv://username:password@cluster.mongodb.net/database
	startIdx := strings.Index(connStr, "@")
	if startIdx == -1 {
		// No auth, format: mongodb://host:port/database
		startIdx = strings.Index(connStr, "://")
		if startIdx == -1 {
			return ""
		}
		startIdx += 3
	} else {
		startIdx += 1
	}

	endIdx := strings.Index(connStr[startIdx:], "/")
	if endIdx == -1 {
		endIdx = strings.Index(connStr[startIdx:], "?")
	}
	if endIdx == -1 {
		endIdx = len(connStr[startIdx:])
	}

	hostPort := connStr[startIdx : startIdx+endIdx]
	// Split host:port
	parts := strings.Split(hostPort, ":")
	if len(parts) > 0 {
		return parts[0]
	}

	return ""
}

// extractDatabase extracts the database name from MongoDB connection string
func (c *MongoDBConnector) extractDatabase(connStr string) string {
	// Format: mongodb://host:port/database or mongodb+srv://host/database
	startIdx := strings.LastIndex(connStr, "/")
	if startIdx == -1 {
		return ""
	}
	startIdx++

	endIdx := strings.Index(connStr[startIdx:], "?")
	if endIdx == -1 {
		return connStr[startIdx:]
	}

	return connStr[startIdx : startIdx+endIdx]
}

// validateHost validates the database host to prevent SSRF
func (c *MongoDBConnector) validateHost(host string) error {
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

// toBSON converts a map to BSON
func (c *MongoDBConnector) toBSON(m map[string]interface{}) (bson.M, error) {
	result := bson.M{}
	for k, v := range m {
		result[k] = v
	}
	return result, nil
}
