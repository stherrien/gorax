package database

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/gorax/gorax/internal/credential"
	"github.com/gorax/gorax/internal/executor/actions"
)

const (
	// DefaultMongoTimeout is the default timeout for MongoDB operations
	DefaultMongoTimeout = 30 * time.Second
)

// MongoClient interface abstracts MongoDB operations for testability
type MongoClient interface {
	Find(ctx context.Context, db, collection string, filter, projection, sort map[string]interface{}, limit, skip int64) ([]map[string]interface{}, error)
	Insert(ctx context.Context, db, collection string, documents []map[string]interface{}) ([]interface{}, error)
	Update(ctx context.Context, db, collection string, filter, update map[string]interface{}, upsert, multi bool) (*MongoUpdateResult, error)
	Delete(ctx context.Context, db, collection string, filter map[string]interface{}, multi bool) (int64, error)
	Aggregate(ctx context.Context, db, collection string, pipeline []map[string]interface{}) ([]map[string]interface{}, error)
	Close(ctx context.Context) error
}

// mongoClientImpl implements MongoClient using the official MongoDB driver
type mongoClientImpl struct {
	client *mongo.Client
	dbName string
}

// newMongoClient creates a new MongoDB client
func newMongoClient(connStr string) (MongoClient, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	clientOptions := options.Client().ApplyURI(connStr)
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	// Ping to verify connection
	if err := client.Ping(ctx, nil); err != nil {
		_ = client.Disconnect(ctx) // Best effort disconnect on ping failure
		return nil, fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	// Default database name will be extracted from URI or use "admin"
	// The actual database name is passed in each operation
	return &mongoClientImpl{
		client: client,
		dbName: "admin", // default, will be overridden by operations
	}, nil
}

// Find implements MongoClient.Find
func (m *mongoClientImpl) Find(ctx context.Context, db, collection string, filter, projection, sort map[string]interface{}, limit, skip int64) ([]map[string]interface{}, error) {
	coll := m.client.Database(db).Collection(collection)

	// Build find options
	findOptions := options.Find()
	if projection != nil {
		findOptions.SetProjection(projection)
	}
	if sort != nil {
		findOptions.SetSort(sort)
	}
	if limit > 0 {
		findOptions.SetLimit(limit)
	}
	if skip > 0 {
		findOptions.SetSkip(skip)
	}

	// Execute find
	cursor, err := coll.Find(ctx, filter, findOptions)
	if err != nil {
		return nil, fmt.Errorf("find failed: %w", err)
	}
	defer cursor.Close(ctx)

	// Decode results
	var results []map[string]interface{}
	if err := cursor.All(ctx, &results); err != nil {
		return nil, fmt.Errorf("failed to decode results: %w", err)
	}

	return results, nil
}

// Insert implements MongoClient.Insert
func (m *mongoClientImpl) Insert(ctx context.Context, db, collection string, documents []map[string]interface{}) ([]interface{}, error) {
	coll := m.client.Database(db).Collection(collection)

	// Convert to []interface{} for InsertMany
	docs := make([]interface{}, len(documents))
	for i, doc := range documents {
		docs[i] = doc
	}

	// Execute insert
	result, err := coll.InsertMany(ctx, docs)
	if err != nil {
		return nil, fmt.Errorf("insert failed: %w", err)
	}

	return result.InsertedIDs, nil
}

// Update implements MongoClient.Update
func (m *mongoClientImpl) Update(ctx context.Context, db, collection string, filter, update map[string]interface{}, upsert, multi bool) (*MongoUpdateResult, error) {
	coll := m.client.Database(db).Collection(collection)

	updateOptions := options.Update().SetUpsert(upsert)

	var result *mongo.UpdateResult
	var err error

	if multi {
		result, err = coll.UpdateMany(ctx, filter, update, updateOptions)
	} else {
		result, err = coll.UpdateOne(ctx, filter, update, updateOptions)
	}

	if err != nil {
		return nil, fmt.Errorf("update failed: %w", err)
	}

	upsertedCount := int64(0)
	if result.UpsertedID != nil {
		upsertedCount = 1
	}

	return &MongoUpdateResult{
		MatchedCount:  result.MatchedCount,
		ModifiedCount: result.ModifiedCount,
		UpsertedCount: upsertedCount,
		UpsertedID:    result.UpsertedID,
	}, nil
}

// Delete implements MongoClient.Delete
func (m *mongoClientImpl) Delete(ctx context.Context, db, collection string, filter map[string]interface{}, multi bool) (int64, error) {
	coll := m.client.Database(db).Collection(collection)

	var result *mongo.DeleteResult
	var err error

	if multi {
		result, err = coll.DeleteMany(ctx, filter)
	} else {
		result, err = coll.DeleteOne(ctx, filter)
	}

	if err != nil {
		return 0, fmt.Errorf("delete failed: %w", err)
	}

	return result.DeletedCount, nil
}

// Aggregate implements MongoClient.Aggregate
func (m *mongoClientImpl) Aggregate(ctx context.Context, db, collection string, pipeline []map[string]interface{}) ([]map[string]interface{}, error) {
	coll := m.client.Database(db).Collection(collection)

	// Convert pipeline to []interface{}
	pipelineInterfaces := make([]interface{}, len(pipeline))
	for i, stage := range pipeline {
		pipelineInterfaces[i] = stage
	}

	// Execute aggregation
	cursor, err := coll.Aggregate(ctx, pipelineInterfaces)
	if err != nil {
		return nil, fmt.Errorf("aggregate failed: %w", err)
	}
	defer cursor.Close(ctx)

	// Decode results
	var results []map[string]interface{}
	if err := cursor.All(ctx, &results); err != nil {
		return nil, fmt.Errorf("failed to decode results: %w", err)
	}

	return results, nil
}

// Close implements MongoClient.Close
func (m *mongoClientImpl) Close(ctx context.Context) error {
	return m.client.Disconnect(ctx)
}

// MongoFindAction executes find queries on MongoDB
type MongoFindAction struct {
	credentialService credential.Service
	clientFactory     func(connStr string) (MongoClient, error)
}

// NewMongoFindAction creates a new MongoDB find action
func NewMongoFindAction(credentialService credential.Service) *MongoFindAction {
	return &MongoFindAction{
		credentialService: credentialService,
		clientFactory:     newMongoClient,
	}
}

// Execute implements the Action interface
func (a *MongoFindAction) Execute(ctx context.Context, input *actions.ActionInput) (*actions.ActionOutput, error) {
	// Parse config
	config, ok := input.Config.(MongoFindConfig)
	if !ok {
		return nil, fmt.Errorf("invalid config type: expected MongoFindConfig")
	}

	// Validate config
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	// Get connection string and database name
	connStr, dbName, err := getMongoConnectionDetails(ctx, a.credentialService, input.Context)
	if err != nil {
		return nil, err
	}

	// Create MongoDB client
	client, err := a.clientFactory(connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to create MongoDB client: %w", err)
	}
	defer client.Close(ctx)

	// Apply timeout if specified
	queryCtx := ctx
	if config.Timeout > 0 {
		var cancel context.CancelFunc
		queryCtx, cancel = context.WithTimeout(ctx, time.Duration(config.Timeout)*time.Second)
		defer cancel()
	}

	// Execute find
	documents, err := client.Find(queryCtx, dbName, config.Collection, config.Filter, config.Projection, config.Sort, config.Limit, config.Skip)
	if err != nil {
		return nil, fmt.Errorf("find operation failed: %w", err)
	}

	result := &MongoFindResult{
		Documents: documents,
		Count:     len(documents),
	}

	// Create output
	output := actions.NewActionOutput(result)
	output.WithMetadata("count", result.Count)
	output.WithMetadata("collection", config.Collection)

	return output, nil
}

// MongoInsertAction inserts documents into MongoDB
type MongoInsertAction struct {
	credentialService credential.Service
	clientFactory     func(connStr string) (MongoClient, error)
}

// NewMongoInsertAction creates a new MongoDB insert action
func NewMongoInsertAction(credentialService credential.Service) *MongoInsertAction {
	return &MongoInsertAction{
		credentialService: credentialService,
		clientFactory:     newMongoClient,
	}
}

// Execute implements the Action interface
func (a *MongoInsertAction) Execute(ctx context.Context, input *actions.ActionInput) (*actions.ActionOutput, error) {
	// Parse config
	config, ok := input.Config.(MongoInsertConfig)
	if !ok {
		return nil, fmt.Errorf("invalid config type: expected MongoInsertConfig")
	}

	// Validate config
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	// Get connection string and database name
	connStr, dbName, err := getMongoConnectionDetails(ctx, a.credentialService, input.Context)
	if err != nil {
		return nil, err
	}

	// Create MongoDB client
	client, err := a.clientFactory(connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to create MongoDB client: %w", err)
	}
	defer client.Close(ctx)

	// Apply timeout if specified
	insertCtx := ctx
	if config.Timeout > 0 {
		var cancel context.CancelFunc
		insertCtx, cancel = context.WithTimeout(ctx, time.Duration(config.Timeout)*time.Second)
		defer cancel()
	}

	// Execute insert
	insertedIDs, err := client.Insert(insertCtx, dbName, config.Collection, config.Documents)
	if err != nil {
		return nil, fmt.Errorf("insert operation failed: %w", err)
	}

	result := &MongoInsertResult{
		InsertedCount: len(insertedIDs),
		InsertedIDs:   insertedIDs,
	}

	// Create output
	output := actions.NewActionOutput(result)
	output.WithMetadata("inserted_count", result.InsertedCount)
	output.WithMetadata("collection", config.Collection)

	return output, nil
}

// MongoUpdateAction updates documents in MongoDB
type MongoUpdateAction struct {
	credentialService credential.Service
	clientFactory     func(connStr string) (MongoClient, error)
}

// NewMongoUpdateAction creates a new MongoDB update action
func NewMongoUpdateAction(credentialService credential.Service) *MongoUpdateAction {
	return &MongoUpdateAction{
		credentialService: credentialService,
		clientFactory:     newMongoClient,
	}
}

// Execute implements the Action interface
func (a *MongoUpdateAction) Execute(ctx context.Context, input *actions.ActionInput) (*actions.ActionOutput, error) {
	// Parse config
	config, ok := input.Config.(MongoUpdateConfig)
	if !ok {
		return nil, fmt.Errorf("invalid config type: expected MongoUpdateConfig")
	}

	// Validate config
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	// Get connection string and database name
	connStr, dbName, err := getMongoConnectionDetails(ctx, a.credentialService, input.Context)
	if err != nil {
		return nil, err
	}

	// Create MongoDB client
	client, err := a.clientFactory(connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to create MongoDB client: %w", err)
	}
	defer client.Close(ctx)

	// Apply timeout if specified
	updateCtx := ctx
	if config.Timeout > 0 {
		var cancel context.CancelFunc
		updateCtx, cancel = context.WithTimeout(ctx, time.Duration(config.Timeout)*time.Second)
		defer cancel()
	}

	// Execute update
	result, err := client.Update(updateCtx, dbName, config.Collection, config.Filter, config.Update, config.Upsert, config.Multi)
	if err != nil {
		return nil, fmt.Errorf("update operation failed: %w", err)
	}

	// Create output
	output := actions.NewActionOutput(result)
	output.WithMetadata("matched_count", result.MatchedCount)
	output.WithMetadata("modified_count", result.ModifiedCount)
	output.WithMetadata("collection", config.Collection)

	return output, nil
}

// MongoDeleteAction deletes documents from MongoDB
type MongoDeleteAction struct {
	credentialService credential.Service
	clientFactory     func(connStr string) (MongoClient, error)
}

// NewMongoDeleteAction creates a new MongoDB delete action
func NewMongoDeleteAction(credentialService credential.Service) *MongoDeleteAction {
	return &MongoDeleteAction{
		credentialService: credentialService,
		clientFactory:     newMongoClient,
	}
}

// Execute implements the Action interface
func (a *MongoDeleteAction) Execute(ctx context.Context, input *actions.ActionInput) (*actions.ActionOutput, error) {
	// Parse config
	config, ok := input.Config.(MongoDeleteConfig)
	if !ok {
		return nil, fmt.Errorf("invalid config type: expected MongoDeleteConfig")
	}

	// Validate config
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	// Get connection string and database name
	connStr, dbName, err := getMongoConnectionDetails(ctx, a.credentialService, input.Context)
	if err != nil {
		return nil, err
	}

	// Create MongoDB client
	client, err := a.clientFactory(connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to create MongoDB client: %w", err)
	}
	defer client.Close(ctx)

	// Apply timeout if specified
	deleteCtx := ctx
	if config.Timeout > 0 {
		var cancel context.CancelFunc
		deleteCtx, cancel = context.WithTimeout(ctx, time.Duration(config.Timeout)*time.Second)
		defer cancel()
	}

	// Execute delete
	deletedCount, err := client.Delete(deleteCtx, dbName, config.Collection, config.Filter, config.Multi)
	if err != nil {
		return nil, fmt.Errorf("delete operation failed: %w", err)
	}

	result := &MongoDeleteResult{
		DeletedCount: deletedCount,
	}

	// Create output
	output := actions.NewActionOutput(result)
	output.WithMetadata("deleted_count", deletedCount)
	output.WithMetadata("collection", config.Collection)

	return output, nil
}

// MongoAggregateAction executes aggregation pipelines on MongoDB
type MongoAggregateAction struct {
	credentialService credential.Service
	clientFactory     func(connStr string) (MongoClient, error)
}

// NewMongoAggregateAction creates a new MongoDB aggregate action
func NewMongoAggregateAction(credentialService credential.Service) *MongoAggregateAction {
	return &MongoAggregateAction{
		credentialService: credentialService,
		clientFactory:     newMongoClient,
	}
}

// Execute implements the Action interface
func (a *MongoAggregateAction) Execute(ctx context.Context, input *actions.ActionInput) (*actions.ActionOutput, error) {
	// Parse config
	config, ok := input.Config.(MongoAggregateConfig)
	if !ok {
		return nil, fmt.Errorf("invalid config type: expected MongoAggregateConfig")
	}

	// Validate config
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	// Get connection string and database name
	connStr, dbName, err := getMongoConnectionDetails(ctx, a.credentialService, input.Context)
	if err != nil {
		return nil, err
	}

	// Create MongoDB client
	client, err := a.clientFactory(connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to create MongoDB client: %w", err)
	}
	defer client.Close(ctx)

	// Apply timeout if specified
	aggCtx := ctx
	if config.Timeout > 0 {
		var cancel context.CancelFunc
		aggCtx, cancel = context.WithTimeout(ctx, time.Duration(config.Timeout)*time.Second)
		defer cancel()
	}

	// Execute aggregation
	documents, err := client.Aggregate(aggCtx, dbName, config.Collection, config.Pipeline)
	if err != nil {
		return nil, fmt.Errorf("aggregation failed: %w", err)
	}

	result := &MongoAggregateResult{
		Documents: documents,
		Count:     len(documents),
	}

	// Create output
	output := actions.NewActionOutput(result)
	output.WithMetadata("count", result.Count)
	output.WithMetadata("collection", config.Collection)

	return output, nil
}

// getMongoConnectionDetails retrieves connection string and extracts database name
func getMongoConnectionDetails(ctx context.Context, credService credential.Service, inputCtx map[string]interface{}) (string, string, error) {
	// Extract tenant_id
	tenantID, err := extractString(inputCtx, "env.tenant_id")
	if err != nil {
		return "", "", fmt.Errorf("tenant_id is required in context: %w", err)
	}

	// Extract credential_id
	credentialID, err := extractString(inputCtx, "credential_id")
	if err != nil {
		return "", "", fmt.Errorf("credential_id is required in context: %w", err)
	}

	// Retrieve and decrypt credential
	decryptedCred, err := credService.GetValue(ctx, tenantID, credentialID, "system")
	if err != nil {
		return "", "", fmt.Errorf("failed to retrieve credential: %w", err)
	}

	// Extract connection_string from credential
	connStr, ok := decryptedCred.Value["connection_string"].(string)
	if !ok || connStr == "" {
		return "", "", fmt.Errorf("connection_string not found in credential")
	}

	// Extract database name from credential (optional, falls back to URI)
	dbName, ok := decryptedCred.Value["database"].(string)
	if !ok || dbName == "" {
		// Try to parse from connection string or use default
		dbName = parseDatabaseFromURI(connStr)
	}

	return connStr, dbName, nil
}

// parseDatabaseFromURI extracts database name from MongoDB URI
func parseDatabaseFromURI(uri string) string {
	// Simple parsing - can be enhanced with proper URI parsing
	// Format: mongodb://[username:password@]host[:port]/[database][?options]
	// For now, return a default
	return "admin"
}

// Ensure bson is used (import check)
var _ = bson.M{}
