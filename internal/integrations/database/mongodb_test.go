package database

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gorax/gorax/internal/credential"
	"github.com/gorax/gorax/internal/executor/actions"
)

// TestMongoFindAction_Execute tests the MongoDB find action
func TestMongoFindAction_Execute(t *testing.T) {
	tests := []struct {
		name           string
		config         MongoFindConfig
		context        map[string]interface{}
		mockCredential *credential.DecryptedValue
		credError      error
		mockDocuments  []map[string]interface{}
		mockError      error
		wantErr        bool
		errorContains  string
		validate       func(t *testing.T, output *actions.ActionOutput)
	}{
		{
			name: "successful find with filter",
			config: MongoFindConfig{
				Collection: "users",
				Filter: map[string]interface{}{
					"active": true,
				},
			},
			context: map[string]interface{}{
				"env": map[string]interface{}{
					"tenant_id": "tenant-123",
				},
				"credential_id": "cred-123",
			},
			mockCredential: &credential.DecryptedValue{
				Value: map[string]interface{}{
					"connection_string": "mongodb://localhost:27017/testdb",
				},
			},
			mockDocuments: []map[string]interface{}{
				{"_id": "1", "name": "John Doe", "active": true},
				{"_id": "2", "name": "Jane Smith", "active": true},
			},
			wantErr: false,
			validate: func(t *testing.T, output *actions.ActionOutput) {
				result, ok := output.Data.(*MongoFindResult)
				require.True(t, ok)
				assert.Equal(t, 2, result.Count)
				assert.Len(t, result.Documents, 2)
				assert.Equal(t, "John Doe", result.Documents[0]["name"])
			},
		},
		{
			name: "find with projection",
			config: MongoFindConfig{
				Collection: "users",
				Filter:     map[string]interface{}{},
				Projection: map[string]interface{}{
					"name":  1,
					"email": 1,
				},
			},
			context: map[string]interface{}{
				"env": map[string]interface{}{
					"tenant_id": "tenant-123",
				},
				"credential_id": "cred-123",
			},
			mockCredential: &credential.DecryptedValue{
				Value: map[string]interface{}{
					"connection_string": "mongodb://localhost:27017/testdb",
				},
			},
			mockDocuments: []map[string]interface{}{
				{"_id": "1", "name": "John Doe", "email": "john@example.com"},
			},
			wantErr: false,
		},
		{
			name: "find with sort and limit",
			config: MongoFindConfig{
				Collection: "users",
				Filter:     map[string]interface{}{},
				Sort: map[string]interface{}{
					"created_at": -1,
				},
				Limit: 10,
			},
			context: map[string]interface{}{
				"env": map[string]interface{}{
					"tenant_id": "tenant-123",
				},
				"credential_id": "cred-123",
			},
			mockCredential: &credential.DecryptedValue{
				Value: map[string]interface{}{
					"connection_string": "mongodb://localhost:27017/testdb",
				},
			},
			mockDocuments: []map[string]interface{}{
				{"_id": "1", "name": "User 1"},
			},
			wantErr: false,
		},
		{
			name: "find with skip and limit for pagination",
			config: MongoFindConfig{
				Collection: "users",
				Filter:     map[string]interface{}{},
				Skip:       20,
				Limit:      10,
			},
			context: map[string]interface{}{
				"env": map[string]interface{}{
					"tenant_id": "tenant-123",
				},
				"credential_id": "cred-123",
			},
			mockCredential: &credential.DecryptedValue{
				Value: map[string]interface{}{
					"connection_string": "mongodb://localhost:27017/testdb",
				},
			},
			mockDocuments: []map[string]interface{}{},
			wantErr:       false,
		},
		{
			name: "missing collection",
			config: MongoFindConfig{
				Collection: "",
			},
			context: map[string]interface{}{
				"env": map[string]interface{}{
					"tenant_id": "tenant-123",
				},
				"credential_id": "cred-123",
			},
			wantErr:       true,
			errorContains: "invalid collection",
		},
		{
			name: "missing connection string",
			config: MongoFindConfig{
				Collection: "users",
			},
			context: map[string]interface{}{
				"env": map[string]interface{}{
					"tenant_id": "tenant-123",
				},
				"credential_id": "cred-123",
			},
			mockCredential: &credential.DecryptedValue{
				Value: map[string]interface{}{},
			},
			wantErr:       true,
			errorContains: "connection_string not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock credential service
			mockCred := &MockCredentialService{
				GetValueFunc: func(ctx context.Context, tenantID, credentialID, userID string) (*credential.DecryptedValue, error) {
					if tt.credError != nil {
						return nil, tt.credError
					}
					return tt.mockCredential, nil
				},
			}

			// Create action with mock client factory
			action := NewMongoFindAction(mockCred)
			if tt.mockDocuments != nil {
				action.clientFactory = func(connStr string) (MongoClient, error) {
					return &MockMongoClient{
						FindFunc: func(ctx context.Context, db, collection string, filter, projection, sort map[string]interface{}, limit, skip int64) ([]map[string]interface{}, error) {
							if tt.mockError != nil {
								return nil, tt.mockError
							}
							return tt.mockDocuments, nil
						},
					}, nil
				}
			}

			// Execute
			input := actions.NewActionInput(tt.config, tt.context)
			ctx := context.Background()
			output, err := action.Execute(ctx, input)

			// Assertions
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, err)
				if tt.validate != nil {
					tt.validate(t, output)
				}
			}
		})
	}
}

// TestMongoInsertAction_Execute tests the MongoDB insert action
func TestMongoInsertAction_Execute(t *testing.T) {
	tests := []struct {
		name            string
		config          MongoInsertConfig
		context         map[string]interface{}
		mockCredential  *credential.DecryptedValue
		mockInsertedIDs []interface{}
		wantErr         bool
		errorContains   string
		validate        func(t *testing.T, output *actions.ActionOutput)
	}{
		{
			name: "successful insert one document",
			config: MongoInsertConfig{
				Collection: "users",
				Documents: []map[string]interface{}{
					{"name": "John Doe", "email": "john@example.com"},
				},
			},
			context: map[string]interface{}{
				"env": map[string]interface{}{
					"tenant_id": "tenant-123",
				},
				"credential_id": "cred-123",
			},
			mockCredential: &credential.DecryptedValue{
				Value: map[string]interface{}{
					"connection_string": "mongodb://localhost:27017/testdb",
				},
			},
			mockInsertedIDs: []interface{}{"507f1f77bcf86cd799439011"},
			wantErr:         false,
			validate: func(t *testing.T, output *actions.ActionOutput) {
				result, ok := output.Data.(*MongoInsertResult)
				require.True(t, ok)
				assert.Equal(t, 1, result.InsertedCount)
				assert.Len(t, result.InsertedIDs, 1)
			},
		},
		{
			name: "successful insert multiple documents",
			config: MongoInsertConfig{
				Collection: "users",
				Documents: []map[string]interface{}{
					{"name": "User 1", "email": "user1@example.com"},
					{"name": "User 2", "email": "user2@example.com"},
					{"name": "User 3", "email": "user3@example.com"},
				},
			},
			context: map[string]interface{}{
				"env": map[string]interface{}{
					"tenant_id": "tenant-123",
				},
				"credential_id": "cred-123",
			},
			mockCredential: &credential.DecryptedValue{
				Value: map[string]interface{}{
					"connection_string": "mongodb://localhost:27017/testdb",
				},
			},
			mockInsertedIDs: []interface{}{"id1", "id2", "id3"},
			wantErr:         false,
			validate: func(t *testing.T, output *actions.ActionOutput) {
				result, ok := output.Data.(*MongoInsertResult)
				require.True(t, ok)
				assert.Equal(t, 3, result.InsertedCount)
			},
		},
		{
			name: "missing collection",
			config: MongoInsertConfig{
				Collection: "",
				Documents:  []map[string]interface{}{{"name": "Test"}},
			},
			context: map[string]interface{}{
				"env": map[string]interface{}{
					"tenant_id": "tenant-123",
				},
				"credential_id": "cred-123",
			},
			wantErr:       true,
			errorContains: "invalid collection",
		},
		{
			name: "empty documents",
			config: MongoInsertConfig{
				Collection: "users",
				Documents:  []map[string]interface{}{},
			},
			context: map[string]interface{}{
				"env": map[string]interface{}{
					"tenant_id": "tenant-123",
				},
				"credential_id": "cred-123",
			},
			wantErr:       true,
			errorContains: "at least one document is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock credential service
			mockCred := &MockCredentialService{
				GetValueFunc: func(ctx context.Context, tenantID, credentialID, userID string) (*credential.DecryptedValue, error) {
					return tt.mockCredential, nil
				},
			}

			// Create action with mock client factory
			action := NewMongoInsertAction(mockCred)
			if tt.mockInsertedIDs != nil {
				action.clientFactory = func(connStr string) (MongoClient, error) {
					return &MockMongoClient{
						InsertFunc: func(ctx context.Context, db, collection string, documents []map[string]interface{}) ([]interface{}, error) {
							return tt.mockInsertedIDs, nil
						},
					}, nil
				}
			}

			// Execute
			input := actions.NewActionInput(tt.config, tt.context)
			ctx := context.Background()
			output, err := action.Execute(ctx, input)

			// Assertions
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, err)
				if tt.validate != nil {
					tt.validate(t, output)
				}
			}
		})
	}
}

// TestMongoUpdateAction_Execute tests the MongoDB update action
func TestMongoUpdateAction_Execute(t *testing.T) {
	tests := []struct {
		name           string
		config         MongoUpdateConfig
		context        map[string]interface{}
		mockCredential *credential.DecryptedValue
		mockResult     *MongoUpdateResult
		wantErr        bool
		errorContains  string
		validate       func(t *testing.T, output *actions.ActionOutput)
	}{
		{
			name: "successful update one document",
			config: MongoUpdateConfig{
				Collection: "users",
				Filter:     map[string]interface{}{"email": "john@example.com"},
				Update:     map[string]interface{}{"$set": map[string]interface{}{"active": false}},
				Multi:      false,
			},
			context: map[string]interface{}{
				"env": map[string]interface{}{
					"tenant_id": "tenant-123",
				},
				"credential_id": "cred-123",
			},
			mockCredential: &credential.DecryptedValue{
				Value: map[string]interface{}{
					"connection_string": "mongodb://localhost:27017/testdb",
				},
			},
			mockResult: &MongoUpdateResult{
				MatchedCount:  1,
				ModifiedCount: 1,
			},
			wantErr: false,
			validate: func(t *testing.T, output *actions.ActionOutput) {
				result, ok := output.Data.(*MongoUpdateResult)
				require.True(t, ok)
				assert.Equal(t, int64(1), result.MatchedCount)
				assert.Equal(t, int64(1), result.ModifiedCount)
			},
		},
		{
			name: "successful update multiple documents",
			config: MongoUpdateConfig{
				Collection: "users",
				Filter:     map[string]interface{}{"active": true},
				Update:     map[string]interface{}{"$set": map[string]interface{}{"verified": true}},
				Multi:      true,
			},
			context: map[string]interface{}{
				"env": map[string]interface{}{
					"tenant_id": "tenant-123",
				},
				"credential_id": "cred-123",
			},
			mockCredential: &credential.DecryptedValue{
				Value: map[string]interface{}{
					"connection_string": "mongodb://localhost:27017/testdb",
				},
			},
			mockResult: &MongoUpdateResult{
				MatchedCount:  5,
				ModifiedCount: 5,
			},
			wantErr: false,
			validate: func(t *testing.T, output *actions.ActionOutput) {
				result, ok := output.Data.(*MongoUpdateResult)
				require.True(t, ok)
				assert.Equal(t, int64(5), result.ModifiedCount)
			},
		},
		{
			name: "upsert creates new document",
			config: MongoUpdateConfig{
				Collection: "users",
				Filter:     map[string]interface{}{"email": "new@example.com"},
				Update:     map[string]interface{}{"$set": map[string]interface{}{"name": "New User"}},
				Upsert:     true,
			},
			context: map[string]interface{}{
				"env": map[string]interface{}{
					"tenant_id": "tenant-123",
				},
				"credential_id": "cred-123",
			},
			mockCredential: &credential.DecryptedValue{
				Value: map[string]interface{}{
					"connection_string": "mongodb://localhost:27017/testdb",
				},
			},
			mockResult: &MongoUpdateResult{
				MatchedCount:  0,
				ModifiedCount: 0,
				UpsertedCount: 1,
				UpsertedID:    "507f1f77bcf86cd799439011",
			},
			wantErr: false,
			validate: func(t *testing.T, output *actions.ActionOutput) {
				result, ok := output.Data.(*MongoUpdateResult)
				require.True(t, ok)
				assert.Equal(t, int64(1), result.UpsertedCount)
				assert.NotNil(t, result.UpsertedID)
			},
		},
		{
			name: "missing filter",
			config: MongoUpdateConfig{
				Collection: "users",
				Update:     map[string]interface{}{"$set": map[string]interface{}{"active": false}},
			},
			context: map[string]interface{}{
				"env": map[string]interface{}{
					"tenant_id": "tenant-123",
				},
				"credential_id": "cred-123",
			},
			wantErr:       true,
			errorContains: "invalid filter",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock credential service
			mockCred := &MockCredentialService{
				GetValueFunc: func(ctx context.Context, tenantID, credentialID, userID string) (*credential.DecryptedValue, error) {
					return tt.mockCredential, nil
				},
			}

			// Create action with mock client factory
			action := NewMongoUpdateAction(mockCred)
			if tt.mockResult != nil {
				action.clientFactory = func(connStr string) (MongoClient, error) {
					return &MockMongoClient{
						UpdateFunc: func(ctx context.Context, db, collection string, filter, update map[string]interface{}, upsert, multi bool) (*MongoUpdateResult, error) {
							return tt.mockResult, nil
						},
					}, nil
				}
			}

			// Execute
			input := actions.NewActionInput(tt.config, tt.context)
			ctx := context.Background()
			output, err := action.Execute(ctx, input)

			// Assertions
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, err)
				if tt.validate != nil {
					tt.validate(t, output)
				}
			}
		})
	}
}

// TestMongoDeleteAction_Execute tests the MongoDB delete action
func TestMongoDeleteAction_Execute(t *testing.T) {
	tests := []struct {
		name           string
		config         MongoDeleteConfig
		context        map[string]interface{}
		mockCredential *credential.DecryptedValue
		mockDeleted    int64
		wantErr        bool
		errorContains  string
		validate       func(t *testing.T, output *actions.ActionOutput)
	}{
		{
			name: "successful delete one document",
			config: MongoDeleteConfig{
				Collection: "users",
				Filter:     map[string]interface{}{"_id": "507f1f77bcf86cd799439011"},
				Multi:      false,
			},
			context: map[string]interface{}{
				"env": map[string]interface{}{
					"tenant_id": "tenant-123",
				},
				"credential_id": "cred-123",
			},
			mockCredential: &credential.DecryptedValue{
				Value: map[string]interface{}{
					"connection_string": "mongodb://localhost:27017/testdb",
				},
			},
			mockDeleted: 1,
			wantErr:     false,
			validate: func(t *testing.T, output *actions.ActionOutput) {
				result, ok := output.Data.(*MongoDeleteResult)
				require.True(t, ok)
				assert.Equal(t, int64(1), result.DeletedCount)
			},
		},
		{
			name: "successful delete multiple documents",
			config: MongoDeleteConfig{
				Collection: "users",
				Filter:     map[string]interface{}{"active": false},
				Multi:      true,
			},
			context: map[string]interface{}{
				"env": map[string]interface{}{
					"tenant_id": "tenant-123",
				},
				"credential_id": "cred-123",
			},
			mockCredential: &credential.DecryptedValue{
				Value: map[string]interface{}{
					"connection_string": "mongodb://localhost:27017/testdb",
				},
			},
			mockDeleted: 10,
			wantErr:     false,
			validate: func(t *testing.T, output *actions.ActionOutput) {
				result, ok := output.Data.(*MongoDeleteResult)
				require.True(t, ok)
				assert.Equal(t, int64(10), result.DeletedCount)
			},
		},
		{
			name: "missing filter",
			config: MongoDeleteConfig{
				Collection: "users",
			},
			context: map[string]interface{}{
				"env": map[string]interface{}{
					"tenant_id": "tenant-123",
				},
				"credential_id": "cred-123",
			},
			wantErr:       true,
			errorContains: "invalid filter",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock credential service
			mockCred := &MockCredentialService{
				GetValueFunc: func(ctx context.Context, tenantID, credentialID, userID string) (*credential.DecryptedValue, error) {
					return tt.mockCredential, nil
				},
			}

			// Create action with mock client factory
			action := NewMongoDeleteAction(mockCred)
			if tt.mockDeleted > 0 {
				action.clientFactory = func(connStr string) (MongoClient, error) {
					return &MockMongoClient{
						DeleteFunc: func(ctx context.Context, db, collection string, filter map[string]interface{}, multi bool) (int64, error) {
							return tt.mockDeleted, nil
						},
					}, nil
				}
			}

			// Execute
			input := actions.NewActionInput(tt.config, tt.context)
			ctx := context.Background()
			output, err := action.Execute(ctx, input)

			// Assertions
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, err)
				if tt.validate != nil {
					tt.validate(t, output)
				}
			}
		})
	}
}

// TestMongoAggregateAction_Execute tests the MongoDB aggregate action
func TestMongoAggregateAction_Execute(t *testing.T) {
	tests := []struct {
		name           string
		config         MongoAggregateConfig
		context        map[string]interface{}
		mockCredential *credential.DecryptedValue
		mockDocuments  []map[string]interface{}
		wantErr        bool
		errorContains  string
		validate       func(t *testing.T, output *actions.ActionOutput)
	}{
		{
			name: "successful aggregation pipeline",
			config: MongoAggregateConfig{
				Collection: "orders",
				Pipeline: []map[string]interface{}{
					{"$match": map[string]interface{}{"status": "completed"}},
					{"$group": map[string]interface{}{
						"_id":   "$user_id",
						"total": map[string]interface{}{"$sum": "$amount"},
					}},
					{"$sort": map[string]interface{}{"total": -1}},
				},
			},
			context: map[string]interface{}{
				"env": map[string]interface{}{
					"tenant_id": "tenant-123",
				},
				"credential_id": "cred-123",
			},
			mockCredential: &credential.DecryptedValue{
				Value: map[string]interface{}{
					"connection_string": "mongodb://localhost:27017/testdb",
				},
			},
			mockDocuments: []map[string]interface{}{
				{"_id": "user1", "total": 500.0},
				{"_id": "user2", "total": 300.0},
			},
			wantErr: false,
			validate: func(t *testing.T, output *actions.ActionOutput) {
				result, ok := output.Data.(*MongoAggregateResult)
				require.True(t, ok)
				assert.Equal(t, 2, result.Count)
				assert.Len(t, result.Documents, 2)
			},
		},
		{
			name: "empty pipeline",
			config: MongoAggregateConfig{
				Collection: "users",
				Pipeline:   []map[string]interface{}{},
			},
			context: map[string]interface{}{
				"env": map[string]interface{}{
					"tenant_id": "tenant-123",
				},
				"credential_id": "cred-123",
			},
			wantErr:       true,
			errorContains: "aggregation pipeline is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock credential service
			mockCred := &MockCredentialService{
				GetValueFunc: func(ctx context.Context, tenantID, credentialID, userID string) (*credential.DecryptedValue, error) {
					return tt.mockCredential, nil
				},
			}

			// Create action with mock client factory
			action := NewMongoAggregateAction(mockCred)
			if tt.mockDocuments != nil {
				action.clientFactory = func(connStr string) (MongoClient, error) {
					return &MockMongoClient{
						AggregateFunc: func(ctx context.Context, db, collection string, pipeline []map[string]interface{}) ([]map[string]interface{}, error) {
							return tt.mockDocuments, nil
						},
					}, nil
				}
			}

			// Execute
			input := actions.NewActionInput(tt.config, tt.context)
			ctx := context.Background()
			output, err := action.Execute(ctx, input)

			// Assertions
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, err)
				if tt.validate != nil {
					tt.validate(t, output)
				}
			}
		})
	}
}

// MockMongoClient implements MongoClient for testing
type MockMongoClient struct {
	FindFunc      func(ctx context.Context, db, collection string, filter, projection, sort map[string]interface{}, limit, skip int64) ([]map[string]interface{}, error)
	InsertFunc    func(ctx context.Context, db, collection string, documents []map[string]interface{}) ([]interface{}, error)
	UpdateFunc    func(ctx context.Context, db, collection string, filter, update map[string]interface{}, upsert, multi bool) (*MongoUpdateResult, error)
	DeleteFunc    func(ctx context.Context, db, collection string, filter map[string]interface{}, multi bool) (int64, error)
	AggregateFunc func(ctx context.Context, db, collection string, pipeline []map[string]interface{}) ([]map[string]interface{}, error)
	CloseFunc     func(ctx context.Context) error
}

func (m *MockMongoClient) Find(ctx context.Context, db, collection string, filter, projection, sort map[string]interface{}, limit, skip int64) ([]map[string]interface{}, error) {
	if m.FindFunc != nil {
		return m.FindFunc(ctx, db, collection, filter, projection, sort, limit, skip)
	}
	return nil, errors.New("not implemented")
}

func (m *MockMongoClient) Insert(ctx context.Context, db, collection string, documents []map[string]interface{}) ([]interface{}, error) {
	if m.InsertFunc != nil {
		return m.InsertFunc(ctx, db, collection, documents)
	}
	return nil, errors.New("not implemented")
}

func (m *MockMongoClient) Update(ctx context.Context, db, collection string, filter, update map[string]interface{}, upsert, multi bool) (*MongoUpdateResult, error) {
	if m.UpdateFunc != nil {
		return m.UpdateFunc(ctx, db, collection, filter, update, upsert, multi)
	}
	return nil, errors.New("not implemented")
}

func (m *MockMongoClient) Delete(ctx context.Context, db, collection string, filter map[string]interface{}, multi bool) (int64, error) {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(ctx, db, collection, filter, multi)
	}
	return 0, errors.New("not implemented")
}

func (m *MockMongoClient) Aggregate(ctx context.Context, db, collection string, pipeline []map[string]interface{}) ([]map[string]interface{}, error) {
	if m.AggregateFunc != nil {
		return m.AggregateFunc(ctx, db, collection, pipeline)
	}
	return nil, errors.New("not implemented")
}

func (m *MockMongoClient) Close(ctx context.Context) error {
	if m.CloseFunc != nil {
		return m.CloseFunc(ctx)
	}
	return nil
}
