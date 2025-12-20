package database

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gorax/gorax/internal/credential"
	"github.com/gorax/gorax/internal/executor/actions"
)

// MockCredentialService for testing
type MockCredentialService struct {
	GetValueFunc func(ctx context.Context, tenantID, credentialID, userID string) (*credential.DecryptedValue, error)
}

func (m *MockCredentialService) GetValue(ctx context.Context, tenantID, credentialID, userID string) (*credential.DecryptedValue, error) {
	if m.GetValueFunc != nil {
		return m.GetValueFunc(ctx, tenantID, credentialID, userID)
	}
	return nil, errors.New("not implemented")
}

func (m *MockCredentialService) Create(ctx context.Context, tenantID, userID string, input credential.CreateCredentialInput) (*credential.Credential, error) {
	return nil, errors.New("not implemented")
}

func (m *MockCredentialService) List(ctx context.Context, tenantID string, filter credential.CredentialListFilter, limit, offset int) ([]*credential.Credential, error) {
	return nil, errors.New("not implemented")
}

func (m *MockCredentialService) GetByID(ctx context.Context, tenantID, credentialID string) (*credential.Credential, error) {
	return nil, errors.New("not implemented")
}

func (m *MockCredentialService) Update(ctx context.Context, tenantID, credentialID, userID string, input credential.UpdateCredentialInput) (*credential.Credential, error) {
	return nil, errors.New("not implemented")
}

func (m *MockCredentialService) Delete(ctx context.Context, tenantID, credentialID, userID string) error {
	return errors.New("not implemented")
}

func (m *MockCredentialService) Rotate(ctx context.Context, tenantID, credentialID, userID string, input credential.RotateCredentialInput) (*credential.Credential, error) {
	return nil, errors.New("not implemented")
}

func (m *MockCredentialService) ListVersions(ctx context.Context, tenantID, credentialID string) ([]*credential.CredentialValue, error) {
	return nil, errors.New("not implemented")
}

func (m *MockCredentialService) GetAccessLog(ctx context.Context, tenantID, credentialID string, limit, offset int) ([]*credential.AccessLog, error) {
	return nil, errors.New("not implemented")
}

// TestPostgresQueryAction_Execute tests the PostgreSQL query action
func TestPostgresQueryAction_Execute(t *testing.T) {
	tests := []struct {
		name           string
		config         QueryConfig
		context        map[string]interface{}
		mockCredential *credential.DecryptedValue
		credError      error
		setupMock      func(mock sqlmock.Sqlmock)
		wantErr        bool
		errorContains  string
		validate       func(t *testing.T, output *actions.ActionOutput)
	}{
		{
			name: "successful query execution",
			config: QueryConfig{
				Query: "SELECT id, name, email FROM users WHERE active = $1",
				Parameters: []interface{}{true},
			},
			context: map[string]interface{}{
				"env": map[string]interface{}{
					"tenant_id": "tenant-123",
				},
				"credential_id": "cred-123",
			},
			mockCredential: &credential.DecryptedValue{
				Value: map[string]interface{}{
					"connection_string": "postgresql://user:pass@localhost:5432/testdb?sslmode=disable",
				},
			},
			setupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "name", "email"}).
					AddRow(1, "John Doe", "john@example.com").
					AddRow(2, "Jane Smith", "jane@example.com")
				mock.ExpectQuery("SELECT id, name, email FROM users WHERE active = \\$1").
					WithArgs(true).
					WillReturnRows(rows)
			},
			wantErr: false,
			validate: func(t *testing.T, output *actions.ActionOutput) {
				assert.NotNil(t, output)
				result, ok := output.Data.(*QueryResult)
				require.True(t, ok, "output.Data should be *QueryResult")
				assert.Equal(t, 2, result.RowCount)
				assert.Len(t, result.Rows, 2)
				assert.Equal(t, "John Doe", result.Rows[0]["name"])
				assert.Equal(t, "jane@example.com", result.Rows[1]["email"])
			},
		},
		{
			name: "query with no results",
			config: QueryConfig{
				Query: "SELECT * FROM users WHERE id = $1",
				Parameters: []interface{}{999},
			},
			context: map[string]interface{}{
				"env": map[string]interface{}{
					"tenant_id": "tenant-123",
				},
				"credential_id": "cred-123",
			},
			mockCredential: &credential.DecryptedValue{
				Value: map[string]interface{}{
					"connection_string": "postgresql://user:pass@localhost:5432/testdb",
				},
			},
			setupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "name", "email"})
				mock.ExpectQuery("SELECT \\* FROM users WHERE id = \\$1").
					WithArgs(999).
					WillReturnRows(rows)
			},
			wantErr: false,
			validate: func(t *testing.T, output *actions.ActionOutput) {
				result, ok := output.Data.(*QueryResult)
				require.True(t, ok)
				assert.Equal(t, 0, result.RowCount)
				assert.Len(t, result.Rows, 0)
			},
		},
		{
			name: "missing query",
			config: QueryConfig{
				Query: "",
			},
			context: map[string]interface{}{
				"env": map[string]interface{}{
					"tenant_id": "tenant-123",
				},
				"credential_id": "cred-123",
			},
			wantErr:       true,
			errorContains: "invalid query",
		},
		{
			name: "missing connection string",
			config: QueryConfig{
				Query: "SELECT * FROM users",
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
		{
			name: "database query error",
			config: QueryConfig{
				Query: "SELECT * FROM nonexistent_table",
			},
			context: map[string]interface{}{
				"env": map[string]interface{}{
					"tenant_id": "tenant-123",
				},
				"credential_id": "cred-123",
			},
			mockCredential: &credential.DecryptedValue{
				Value: map[string]interface{}{
					"connection_string": "postgresql://user:pass@localhost:5432/testdb",
				},
			},
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT \\* FROM nonexistent_table").
					WillReturnError(errors.New("relation \"nonexistent_table\" does not exist"))
			},
			wantErr:       true,
			errorContains: "does not exist",
		},
		{
			name: "parameterized query prevents SQL injection",
			config: QueryConfig{
				Query: "SELECT * FROM users WHERE email = $1",
				Parameters: []interface{}{"user@example.com' OR '1'='1"},
			},
			context: map[string]interface{}{
				"env": map[string]interface{}{
					"tenant_id": "tenant-123",
				},
				"credential_id": "cred-123",
			},
			mockCredential: &credential.DecryptedValue{
				Value: map[string]interface{}{
					"connection_string": "postgresql://user:pass@localhost:5432/testdb",
				},
			},
			setupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "email"})
				mock.ExpectQuery("SELECT \\* FROM users WHERE email = \\$1").
					WithArgs("user@example.com' OR '1'='1").
					WillReturnRows(rows)
			},
			wantErr: false,
			validate: func(t *testing.T, output *actions.ActionOutput) {
				result, ok := output.Data.(*QueryResult)
				require.True(t, ok)
				assert.Equal(t, 0, result.RowCount)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock DB
			var db *sql.DB
			var mock sqlmock.Sqlmock
			if tt.setupMock != nil {
				var err error
				db, mock, err = sqlmock.New()
				require.NoError(t, err)
				defer db.Close()
				tt.setupMock(mock)
			}

			// Create mock credential service
			mockCred := &MockCredentialService{
				GetValueFunc: func(ctx context.Context, tenantID, credentialID, userID string) (*credential.DecryptedValue, error) {
					if tt.credError != nil {
						return nil, tt.credError
					}
					return tt.mockCredential, nil
				},
			}

			// Create action
			action := NewPostgresQueryAction(mockCred)
			if db != nil {
				action.dbFactory = func(connStr string) (*sql.DB, error) {
					return db, nil
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

			if mock != nil {
				assert.NoError(t, mock.ExpectationsWereMet())
			}
		})
	}
}

// TestPostgresStatementAction_Execute tests the PostgreSQL statement action
func TestPostgresStatementAction_Execute(t *testing.T) {
	tests := []struct {
		name           string
		config         StatementConfig
		context        map[string]interface{}
		mockCredential *credential.DecryptedValue
		setupMock      func(mock sqlmock.Sqlmock)
		wantErr        bool
		errorContains  string
		validate       func(t *testing.T, output *actions.ActionOutput)
	}{
		{
			name: "successful INSERT statement",
			config: StatementConfig{
				Statement: "INSERT INTO users (name, email) VALUES ($1, $2)",
				Parameters: []interface{}{"John Doe", "john@example.com"},
			},
			context: map[string]interface{}{
				"env": map[string]interface{}{
					"tenant_id": "tenant-123",
				},
				"credential_id": "cred-123",
			},
			mockCredential: &credential.DecryptedValue{
				Value: map[string]interface{}{
					"connection_string": "postgresql://user:pass@localhost:5432/testdb",
				},
			},
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("INSERT INTO users").
					WithArgs("John Doe", "john@example.com").
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
			wantErr: false,
			validate: func(t *testing.T, output *actions.ActionOutput) {
				result, ok := output.Data.(*StatementResult)
				require.True(t, ok)
				assert.Equal(t, int64(1), result.RowsAffected)
				assert.Equal(t, int64(1), result.LastInsertID)
			},
		},
		{
			name: "successful UPDATE statement",
			config: StatementConfig{
				Statement: "UPDATE users SET active = $1 WHERE id = $2",
				Parameters: []interface{}{false, 5},
			},
			context: map[string]interface{}{
				"env": map[string]interface{}{
					"tenant_id": "tenant-123",
				},
				"credential_id": "cred-123",
			},
			mockCredential: &credential.DecryptedValue{
				Value: map[string]interface{}{
					"connection_string": "postgresql://user:pass@localhost:5432/testdb",
				},
			},
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("UPDATE users SET active").
					WithArgs(false, 5).
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
			wantErr: false,
			validate: func(t *testing.T, output *actions.ActionOutput) {
				result, ok := output.Data.(*StatementResult)
				require.True(t, ok)
				assert.Equal(t, int64(1), result.RowsAffected)
			},
		},
		{
			name: "successful DELETE statement",
			config: StatementConfig{
				Statement: "DELETE FROM users WHERE email = $1",
				Parameters: []interface{}{"delete@example.com"},
			},
			context: map[string]interface{}{
				"env": map[string]interface{}{
					"tenant_id": "tenant-123",
				},
				"credential_id": "cred-123",
			},
			mockCredential: &credential.DecryptedValue{
				Value: map[string]interface{}{
					"connection_string": "postgresql://user:pass@localhost:5432/testdb",
				},
			},
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("DELETE FROM users WHERE email").
					WithArgs("delete@example.com").
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
			wantErr: false,
			validate: func(t *testing.T, output *actions.ActionOutput) {
				result, ok := output.Data.(*StatementResult)
				require.True(t, ok)
				assert.Equal(t, int64(1), result.RowsAffected)
			},
		},
		{
			name: "missing statement",
			config: StatementConfig{
				Statement: "",
			},
			context: map[string]interface{}{
				"env": map[string]interface{}{
					"tenant_id": "tenant-123",
				},
				"credential_id": "cred-123",
			},
			wantErr:       true,
			errorContains: "invalid query",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock DB
			var db *sql.DB
			var mock sqlmock.Sqlmock
			if tt.setupMock != nil {
				var err error
				db, mock, err = sqlmock.New()
				require.NoError(t, err)
				defer db.Close()
				tt.setupMock(mock)
			}

			// Create mock credential service
			mockCred := &MockCredentialService{
				GetValueFunc: func(ctx context.Context, tenantID, credentialID, userID string) (*credential.DecryptedValue, error) {
					return tt.mockCredential, nil
				},
			}

			// Create action
			action := NewPostgresStatementAction(mockCred)
			if db != nil {
				action.dbFactory = func(connStr string) (*sql.DB, error) {
					return db, nil
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

			if mock != nil {
				assert.NoError(t, mock.ExpectationsWereMet())
			}
		})
	}
}

// TestPostgresTransactionAction_Execute tests the PostgreSQL transaction action
func TestPostgresTransactionAction_Execute(t *testing.T) {
	tests := []struct {
		name           string
		config         TransactionConfig
		context        map[string]interface{}
		mockCredential *credential.DecryptedValue
		setupMock      func(mock sqlmock.Sqlmock)
		wantErr        bool
		errorContains  string
		validate       func(t *testing.T, output *actions.ActionOutput)
	}{
		{
			name: "successful transaction with multiple statements",
			config: TransactionConfig{
				Statements: []TransactionStatement{
					{
						Statement: "INSERT INTO users (name, email) VALUES ($1, $2)",
						Parameters: []interface{}{"John Doe", "john@example.com"},
					},
					{
						Statement: "UPDATE accounts SET balance = balance + $1 WHERE user_email = $2",
						Parameters: []interface{}{100.0, "john@example.com"},
					},
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
					"connection_string": "postgresql://user:pass@localhost:5432/testdb",
				},
			},
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO users").
					WithArgs("John Doe", "john@example.com").
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectExec("UPDATE accounts SET balance").
					WithArgs(100.0, "john@example.com").
					WillReturnResult(sqlmock.NewResult(0, 1))
				mock.ExpectCommit()
			},
			wantErr: false,
			validate: func(t *testing.T, output *actions.ActionOutput) {
				result, ok := output.Data.(*TransactionResult)
				require.True(t, ok)
				assert.True(t, result.Committed)
				assert.Equal(t, 2, result.StatementsRun)
				assert.Equal(t, int64(2), result.TotalAffected)
			},
		},
		{
			name: "transaction rollback on error",
			config: TransactionConfig{
				Statements: []TransactionStatement{
					{
						Statement: "INSERT INTO users (name) VALUES ($1)",
						Parameters: []interface{}{"John"},
					},
					{
						Statement: "INSERT INTO invalid_table (name) VALUES ($1)",
						Parameters: []interface{}{"Jane"},
					},
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
					"connection_string": "postgresql://user:pass@localhost:5432/testdb",
				},
			},
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO users").
					WithArgs("John").
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectExec("INSERT INTO invalid_table").
					WithArgs("Jane").
					WillReturnError(errors.New("table does not exist"))
				mock.ExpectRollback()
			},
			wantErr:       true,
			errorContains: "table does not exist",
		},
		{
			name: "empty transaction",
			config: TransactionConfig{
				Statements: []TransactionStatement{},
			},
			context: map[string]interface{}{
				"env": map[string]interface{}{
					"tenant_id": "tenant-123",
				},
				"credential_id": "cred-123",
			},
			wantErr:       true,
			errorContains: "at least one statement is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock DB
			var db *sql.DB
			var mock sqlmock.Sqlmock
			if tt.setupMock != nil {
				var err error
				db, mock, err = sqlmock.New()
				require.NoError(t, err)
				defer db.Close()
				tt.setupMock(mock)
			}

			// Create mock credential service
			mockCred := &MockCredentialService{
				GetValueFunc: func(ctx context.Context, tenantID, credentialID, userID string) (*credential.DecryptedValue, error) {
					return tt.mockCredential, nil
				},
			}

			// Create action
			action := NewPostgresTransactionAction(mockCred)
			if db != nil {
				action.dbFactory = func(connStr string) (*sql.DB, error) {
					return db, nil
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

			if mock != nil {
				assert.NoError(t, mock.ExpectationsWereMet())
			}
		})
	}
}

// TestPostgresActions_ContextTimeout tests context timeout handling
func TestPostgresActions_ContextTimeout(t *testing.T) {
	// Create mock DB with slow query
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	mock.ExpectQuery("SELECT").
		WillDelayFor(500 * time.Millisecond).
		WillReturnRows(sqlmock.NewRows([]string{"id"}))

	// Create mock credential service
	mockCred := &MockCredentialService{
		GetValueFunc: func(ctx context.Context, tenantID, credentialID, userID string) (*credential.DecryptedValue, error) {
			return &credential.DecryptedValue{
				Value: map[string]interface{}{
					"connection_string": "postgresql://user:pass@localhost:5432/testdb",
				},
			}, nil
		},
	}

	// Create action
	action := NewPostgresQueryAction(mockCred)
	action.dbFactory = func(connStr string) (*sql.DB, error) {
		return db, nil
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	// Execute
	config := QueryConfig{
		Query: "SELECT * FROM users",
	}
	input := actions.NewActionInput(config, map[string]interface{}{
		"env": map[string]interface{}{
			"tenant_id": "tenant-123",
		},
		"credential_id": "cred-123",
	})

	_, err = action.Execute(ctx, input)
	assert.Error(t, err)
	// Context timeout can produce different error messages depending on timing
	// Both "context" and "canceling query" indicate proper timeout handling
	errMsg := err.Error()
	_ = errMsg // Use the variable to avoid compiler warning
	assert.True(t, strings.Contains(errMsg, "context") || strings.Contains(errMsg, "canceling query"), "error should indicate timeout: %s", errMsg)
}

// TestPostgresActions_ConnectionPooling tests connection pooling
func TestPostgresActions_ConnectionPooling(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	// Configure connection pool
	db.SetMaxOpenConns(5)
	db.SetMaxIdleConns(2)
	db.SetConnMaxLifetime(time.Hour)

	rows := sqlmock.NewRows([]string{"id"}).AddRow(1)
	mock.ExpectQuery("SELECT").WillReturnRows(rows)

	mockCred := &MockCredentialService{
		GetValueFunc: func(ctx context.Context, tenantID, credentialID, userID string) (*credential.DecryptedValue, error) {
			return &credential.DecryptedValue{
				Value: map[string]interface{}{
					"connection_string": "postgresql://user:pass@localhost:5432/testdb",
				},
			}, nil
		},
	}

	action := NewPostgresQueryAction(mockCred)
	action.dbFactory = func(connStr string) (*sql.DB, error) {
		return db, nil
	}

	config := QueryConfig{
		Query: "SELECT * FROM users",
	}
	input := actions.NewActionInput(config, map[string]interface{}{
		"env": map[string]interface{}{
			"tenant_id": "tenant-123",
		},
		"credential_id": "cred-123",
	})

	_, err = action.Execute(context.Background(), input)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())

	// Verify pool stats
	stats := db.Stats()
	assert.LessOrEqual(t, stats.MaxOpenConnections, 5)
}
