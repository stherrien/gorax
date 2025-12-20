package database

import (
	"context"
	"fmt"

	"github.com/gorax/gorax/internal/credential"
	"github.com/gorax/gorax/internal/executor/actions"
	"github.com/gorax/gorax/internal/integrations"
)

// RegisterDatabaseActions registers all database actions with the global registry
func RegisterDatabaseActions(credService credential.Service) error {
	// PostgreSQL actions
	if err := integrations.GlobalRegistry.Register(NewPostgresQueryActionWrapper(credService)); err != nil {
		return fmt.Errorf("failed to register postgres:query: %w", err)
	}
	if err := integrations.GlobalRegistry.Register(NewPostgresStatementActionWrapper(credService)); err != nil {
		return fmt.Errorf("failed to register postgres:statement: %w", err)
	}
	if err := integrations.GlobalRegistry.Register(NewPostgresTransactionActionWrapper(credService)); err != nil {
		return fmt.Errorf("failed to register postgres:transaction: %w", err)
	}

	// MySQL actions
	if err := integrations.GlobalRegistry.Register(NewMySQLQueryActionWrapper(credService)); err != nil {
		return fmt.Errorf("failed to register mysql:query: %w", err)
	}
	if err := integrations.GlobalRegistry.Register(NewMySQLStatementActionWrapper(credService)); err != nil {
		return fmt.Errorf("failed to register mysql:statement: %w", err)
	}
	if err := integrations.GlobalRegistry.Register(NewMySQLTransactionActionWrapper(credService)); err != nil {
		return fmt.Errorf("failed to register mysql:transaction: %w", err)
	}

	// MongoDB actions
	if err := integrations.GlobalRegistry.Register(NewMongoFindActionWrapper(credService)); err != nil {
		return fmt.Errorf("failed to register mongodb:find: %w", err)
	}
	if err := integrations.GlobalRegistry.Register(NewMongoInsertActionWrapper(credService)); err != nil {
		return fmt.Errorf("failed to register mongodb:insert: %w", err)
	}
	if err := integrations.GlobalRegistry.Register(NewMongoUpdateActionWrapper(credService)); err != nil {
		return fmt.Errorf("failed to register mongodb:update: %w", err)
	}
	if err := integrations.GlobalRegistry.Register(NewMongoDeleteActionWrapper(credService)); err != nil {
		return fmt.Errorf("failed to register mongodb:delete: %w", err)
	}
	if err := integrations.GlobalRegistry.Register(NewMongoAggregateActionWrapper(credService)); err != nil {
		return fmt.Errorf("failed to register mongodb:aggregate: %w", err)
	}

	return nil
}

// PostgreSQL Action Wrappers

type postgresQueryActionWrapper struct {
	action *PostgresQueryAction
}

func NewPostgresQueryActionWrapper(credService credential.Service) integrations.Action {
	return &postgresQueryActionWrapper{
		action: NewPostgresQueryAction(credService),
	}
}

func (w *postgresQueryActionWrapper) Execute(ctx context.Context, config map[string]interface{}, input map[string]interface{}) (map[string]interface{}, error) {
	// Convert config to QueryConfig
	queryConfig := QueryConfig{
		Query: getStringFromMap(config, "query"),
		Parameters: getInterfaceSliceFromMap(config, "parameters"),
		Timeout: getIntFromMap(config, "timeout"),
	}

	actionInput := actions.NewActionInput(queryConfig, input)
	output, err := w.action.Execute(ctx, actionInput)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"data": output.Data,
		"metadata": output.Metadata,
	}, nil
}

func (w *postgresQueryActionWrapper) Validate(config map[string]interface{}) error {
	queryConfig := QueryConfig{
		Query: getStringFromMap(config, "query"),
	}
	return queryConfig.Validate()
}

func (w *postgresQueryActionWrapper) Name() string {
	return "postgres:query"
}

func (w *postgresQueryActionWrapper) Description() string {
	return "Execute a SELECT query on PostgreSQL database"
}

type postgresStatementActionWrapper struct {
	action *PostgresStatementAction
}

func NewPostgresStatementActionWrapper(credService credential.Service) integrations.Action {
	return &postgresStatementActionWrapper{
		action: NewPostgresStatementAction(credService),
	}
}

func (w *postgresStatementActionWrapper) Execute(ctx context.Context, config map[string]interface{}, input map[string]interface{}) (map[string]interface{}, error) {
	stmtConfig := StatementConfig{
		Statement: getStringFromMap(config, "statement"),
		Parameters: getInterfaceSliceFromMap(config, "parameters"),
		Timeout: getIntFromMap(config, "timeout"),
	}

	actionInput := actions.NewActionInput(stmtConfig, input)
	output, err := w.action.Execute(ctx, actionInput)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"data": output.Data,
		"metadata": output.Metadata,
	}, nil
}

func (w *postgresStatementActionWrapper) Validate(config map[string]interface{}) error {
	stmtConfig := StatementConfig{
		Statement: getStringFromMap(config, "statement"),
	}
	return stmtConfig.Validate()
}

func (w *postgresStatementActionWrapper) Name() string {
	return "postgres:statement"
}

func (w *postgresStatementActionWrapper) Description() string {
	return "Execute an INSERT/UPDATE/DELETE statement on PostgreSQL database"
}

type postgresTransactionActionWrapper struct {
	action *PostgresTransactionAction
}

func NewPostgresTransactionActionWrapper(credService credential.Service) integrations.Action {
	return &postgresTransactionActionWrapper{
		action: NewPostgresTransactionAction(credService),
	}
}

func (w *postgresTransactionActionWrapper) Execute(ctx context.Context, config map[string]interface{}, input map[string]interface{}) (map[string]interface{}, error) {
	txConfig := TransactionConfig{
		Statements: parseTransactionStatements(config),
		Timeout: getIntFromMap(config, "timeout"),
	}

	actionInput := actions.NewActionInput(txConfig, input)
	output, err := w.action.Execute(ctx, actionInput)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"data": output.Data,
		"metadata": output.Metadata,
	}, nil
}

func (w *postgresTransactionActionWrapper) Validate(config map[string]interface{}) error {
	txConfig := TransactionConfig{
		Statements: parseTransactionStatements(config),
	}
	return txConfig.Validate()
}

func (w *postgresTransactionActionWrapper) Name() string {
	return "postgres:transaction"
}

func (w *postgresTransactionActionWrapper) Description() string {
	return "Execute multiple statements in a PostgreSQL transaction"
}

// MySQL Action Wrappers (similar structure to PostgreSQL)

type mysqlQueryActionWrapper struct {
	action *MySQLQueryAction
}

func NewMySQLQueryActionWrapper(credService credential.Service) integrations.Action {
	return &mysqlQueryActionWrapper{
		action: NewMySQLQueryAction(credService),
	}
}

func (w *mysqlQueryActionWrapper) Execute(ctx context.Context, config map[string]interface{}, input map[string]interface{}) (map[string]interface{}, error) {
	queryConfig := QueryConfig{
		Query: getStringFromMap(config, "query"),
		Parameters: getInterfaceSliceFromMap(config, "parameters"),
		Timeout: getIntFromMap(config, "timeout"),
	}

	actionInput := actions.NewActionInput(queryConfig, input)
	output, err := w.action.Execute(ctx, actionInput)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"data": output.Data,
		"metadata": output.Metadata,
	}, nil
}

func (w *mysqlQueryActionWrapper) Validate(config map[string]interface{}) error {
	queryConfig := QueryConfig{
		Query: getStringFromMap(config, "query"),
	}
	return queryConfig.Validate()
}

func (w *mysqlQueryActionWrapper) Name() string {
	return "mysql:query"
}

func (w *mysqlQueryActionWrapper) Description() string {
	return "Execute a SELECT query on MySQL database"
}

type mysqlStatementActionWrapper struct {
	action *MySQLStatementAction
}

func NewMySQLStatementActionWrapper(credService credential.Service) integrations.Action {
	return &mysqlStatementActionWrapper{
		action: NewMySQLStatementAction(credService),
	}
}

func (w *mysqlStatementActionWrapper) Execute(ctx context.Context, config map[string]interface{}, input map[string]interface{}) (map[string]interface{}, error) {
	stmtConfig := StatementConfig{
		Statement: getStringFromMap(config, "statement"),
		Parameters: getInterfaceSliceFromMap(config, "parameters"),
		Timeout: getIntFromMap(config, "timeout"),
	}

	actionInput := actions.NewActionInput(stmtConfig, input)
	output, err := w.action.Execute(ctx, actionInput)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"data": output.Data,
		"metadata": output.Metadata,
	}, nil
}

func (w *mysqlStatementActionWrapper) Validate(config map[string]interface{}) error {
	stmtConfig := StatementConfig{
		Statement: getStringFromMap(config, "statement"),
	}
	return stmtConfig.Validate()
}

func (w *mysqlStatementActionWrapper) Name() string {
	return "mysql:statement"
}

func (w *mysqlStatementActionWrapper) Description() string {
	return "Execute an INSERT/UPDATE/DELETE statement on MySQL database"
}

type mysqlTransactionActionWrapper struct {
	action *MySQLTransactionAction
}

func NewMySQLTransactionActionWrapper(credService credential.Service) integrations.Action {
	return &mysqlTransactionActionWrapper{
		action: NewMySQLTransactionAction(credService),
	}
}

func (w *mysqlTransactionActionWrapper) Execute(ctx context.Context, config map[string]interface{}, input map[string]interface{}) (map[string]interface{}, error) {
	txConfig := TransactionConfig{
		Statements: parseTransactionStatements(config),
		Timeout: getIntFromMap(config, "timeout"),
	}

	actionInput := actions.NewActionInput(txConfig, input)
	output, err := w.action.Execute(ctx, actionInput)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"data": output.Data,
		"metadata": output.Metadata,
	}, nil
}

func (w *mysqlTransactionActionWrapper) Validate(config map[string]interface{}) error {
	txConfig := TransactionConfig{
		Statements: parseTransactionStatements(config),
	}
	return txConfig.Validate()
}

func (w *mysqlTransactionActionWrapper) Name() string {
	return "mysql:transaction"
}

func (w *mysqlTransactionActionWrapper) Description() string {
	return "Execute multiple statements in a MySQL transaction"
}

// MongoDB Action Wrappers

type mongoFindActionWrapper struct {
	action *MongoFindAction
}

func NewMongoFindActionWrapper(credService credential.Service) integrations.Action {
	return &mongoFindActionWrapper{
		action: NewMongoFindAction(credService),
	}
}

func (w *mongoFindActionWrapper) Execute(ctx context.Context, config map[string]interface{}, input map[string]interface{}) (map[string]interface{}, error) {
	findConfig := MongoFindConfig{
		Collection: getStringFromMap(config, "collection"),
		Filter: getMapFromMap(config, "filter"),
		Projection: getMapFromMap(config, "projection"),
		Sort: getMapFromMap(config, "sort"),
		Limit: int64(getIntFromMap(config, "limit")),
		Skip: int64(getIntFromMap(config, "skip")),
		Timeout: getIntFromMap(config, "timeout"),
	}

	actionInput := actions.NewActionInput(findConfig, input)
	output, err := w.action.Execute(ctx, actionInput)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"data": output.Data,
		"metadata": output.Metadata,
	}, nil
}

func (w *mongoFindActionWrapper) Validate(config map[string]interface{}) error {
	findConfig := MongoFindConfig{
		Collection: getStringFromMap(config, "collection"),
	}
	return findConfig.Validate()
}

func (w *mongoFindActionWrapper) Name() string {
	return "mongodb:find"
}

func (w *mongoFindActionWrapper) Description() string {
	return "Find documents in a MongoDB collection"
}

type mongoInsertActionWrapper struct {
	action *MongoInsertAction
}

func NewMongoInsertActionWrapper(credService credential.Service) integrations.Action {
	return &mongoInsertActionWrapper{
		action: NewMongoInsertAction(credService),
	}
}

func (w *mongoInsertActionWrapper) Execute(ctx context.Context, config map[string]interface{}, input map[string]interface{}) (map[string]interface{}, error) {
	insertConfig := MongoInsertConfig{
		Collection: getStringFromMap(config, "collection"),
		Documents: parseDocuments(config),
		Timeout: getIntFromMap(config, "timeout"),
	}

	actionInput := actions.NewActionInput(insertConfig, input)
	output, err := w.action.Execute(ctx, actionInput)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"data": output.Data,
		"metadata": output.Metadata,
	}, nil
}

func (w *mongoInsertActionWrapper) Validate(config map[string]interface{}) error {
	insertConfig := MongoInsertConfig{
		Collection: getStringFromMap(config, "collection"),
		Documents: parseDocuments(config),
	}
	return insertConfig.Validate()
}

func (w *mongoInsertActionWrapper) Name() string {
	return "mongodb:insert"
}

func (w *mongoInsertActionWrapper) Description() string {
	return "Insert documents into a MongoDB collection"
}

type mongoUpdateActionWrapper struct {
	action *MongoUpdateAction
}

func NewMongoUpdateActionWrapper(credService credential.Service) integrations.Action {
	return &mongoUpdateActionWrapper{
		action: NewMongoUpdateAction(credService),
	}
}

func (w *mongoUpdateActionWrapper) Execute(ctx context.Context, config map[string]interface{}, input map[string]interface{}) (map[string]interface{}, error) {
	updateConfig := MongoUpdateConfig{
		Collection: getStringFromMap(config, "collection"),
		Filter: getMapFromMap(config, "filter"),
		Update: getMapFromMap(config, "update"),
		Upsert: getBoolFromMap(config, "upsert"),
		Multi: getBoolFromMap(config, "multi"),
		Timeout: getIntFromMap(config, "timeout"),
	}

	actionInput := actions.NewActionInput(updateConfig, input)
	output, err := w.action.Execute(ctx, actionInput)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"data": output.Data,
		"metadata": output.Metadata,
	}, nil
}

func (w *mongoUpdateActionWrapper) Validate(config map[string]interface{}) error {
	updateConfig := MongoUpdateConfig{
		Collection: getStringFromMap(config, "collection"),
		Filter: getMapFromMap(config, "filter"),
		Update: getMapFromMap(config, "update"),
	}
	return updateConfig.Validate()
}

func (w *mongoUpdateActionWrapper) Name() string {
	return "mongodb:update"
}

func (w *mongoUpdateActionWrapper) Description() string {
	return "Update documents in a MongoDB collection"
}

type mongoDeleteActionWrapper struct {
	action *MongoDeleteAction
}

func NewMongoDeleteActionWrapper(credService credential.Service) integrations.Action {
	return &mongoDeleteActionWrapper{
		action: NewMongoDeleteAction(credService),
	}
}

func (w *mongoDeleteActionWrapper) Execute(ctx context.Context, config map[string]interface{}, input map[string]interface{}) (map[string]interface{}, error) {
	deleteConfig := MongoDeleteConfig{
		Collection: getStringFromMap(config, "collection"),
		Filter: getMapFromMap(config, "filter"),
		Multi: getBoolFromMap(config, "multi"),
		Timeout: getIntFromMap(config, "timeout"),
	}

	actionInput := actions.NewActionInput(deleteConfig, input)
	output, err := w.action.Execute(ctx, actionInput)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"data": output.Data,
		"metadata": output.Metadata,
	}, nil
}

func (w *mongoDeleteActionWrapper) Validate(config map[string]interface{}) error {
	deleteConfig := MongoDeleteConfig{
		Collection: getStringFromMap(config, "collection"),
		Filter: getMapFromMap(config, "filter"),
	}
	return deleteConfig.Validate()
}

func (w *mongoDeleteActionWrapper) Name() string {
	return "mongodb:delete"
}

func (w *mongoDeleteActionWrapper) Description() string {
	return "Delete documents from a MongoDB collection"
}

type mongoAggregateActionWrapper struct {
	action *MongoAggregateAction
}

func NewMongoAggregateActionWrapper(credService credential.Service) integrations.Action {
	return &mongoAggregateActionWrapper{
		action: NewMongoAggregateAction(credService),
	}
}

func (w *mongoAggregateActionWrapper) Execute(ctx context.Context, config map[string]interface{}, input map[string]interface{}) (map[string]interface{}, error) {
	aggConfig := MongoAggregateConfig{
		Collection: getStringFromMap(config, "collection"),
		Pipeline: parsePipeline(config),
		Timeout: getIntFromMap(config, "timeout"),
	}

	actionInput := actions.NewActionInput(aggConfig, input)
	output, err := w.action.Execute(ctx, actionInput)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"data": output.Data,
		"metadata": output.Metadata,
	}, nil
}

func (w *mongoAggregateActionWrapper) Validate(config map[string]interface{}) error {
	aggConfig := MongoAggregateConfig{
		Collection: getStringFromMap(config, "collection"),
		Pipeline: parsePipeline(config),
	}
	return aggConfig.Validate()
}

func (w *mongoAggregateActionWrapper) Name() string {
	return "mongodb:aggregate"
}

func (w *mongoAggregateActionWrapper) Description() string {
	return "Execute an aggregation pipeline on a MongoDB collection"
}

// Helper functions for parsing config maps

func getStringFromMap(m map[string]interface{}, key string) string {
	if val, ok := m[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}

func getIntFromMap(m map[string]interface{}, key string) int {
	if val, ok := m[key]; ok {
		switch v := val.(type) {
		case int:
			return v
		case int64:
			return int(v)
		case float64:
			return int(v)
		}
	}
	return 0
}

func getBoolFromMap(m map[string]interface{}, key string) bool {
	if val, ok := m[key]; ok {
		if b, ok := val.(bool); ok {
			return b
		}
	}
	return false
}

func getMapFromMap(m map[string]interface{}, key string) map[string]interface{} {
	if val, ok := m[key]; ok {
		if subMap, ok := val.(map[string]interface{}); ok {
			return subMap
		}
	}
	return nil
}

func getInterfaceSliceFromMap(m map[string]interface{}, key string) []interface{} {
	if val, ok := m[key]; ok {
		if slice, ok := val.([]interface{}); ok {
			return slice
		}
	}
	return nil
}

func parseTransactionStatements(config map[string]interface{}) []TransactionStatement {
	var statements []TransactionStatement
	if stmtsVal, ok := config["statements"]; ok {
		if stmtsSlice, ok := stmtsVal.([]interface{}); ok {
			for _, s := range stmtsSlice {
				if stmtMap, ok := s.(map[string]interface{}); ok {
					stmt := TransactionStatement{
						Statement: getStringFromMap(stmtMap, "statement"),
						Parameters: getInterfaceSliceFromMap(stmtMap, "parameters"),
					}
					statements = append(statements, stmt)
				}
			}
		}
	}
	return statements
}

func parseDocuments(config map[string]interface{}) []map[string]interface{} {
	var documents []map[string]interface{}
	if docsVal, ok := config["documents"]; ok {
		if docsSlice, ok := docsVal.([]interface{}); ok {
			for _, d := range docsSlice {
				if docMap, ok := d.(map[string]interface{}); ok {
					documents = append(documents, docMap)
				}
			}
		}
	}
	return documents
}

func parsePipeline(config map[string]interface{}) []map[string]interface{} {
	var pipeline []map[string]interface{}
	if pipelineVal, ok := config["pipeline"]; ok {
		if pipelineSlice, ok := pipelineVal.([]interface{}); ok {
			for _, p := range pipelineSlice {
				if stageMap, ok := p.(map[string]interface{}); ok {
					pipeline = append(pipeline, stageMap)
				}
			}
		}
	}
	return pipeline
}
