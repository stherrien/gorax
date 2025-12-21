package database

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gorax/gorax/internal/credential"
	"github.com/gorax/gorax/internal/executor/actions"
)

// TestMySQLQueryAction_Execute tests the MySQL query action
func TestMySQLQueryAction_Execute(t *testing.T) {
	tests := []struct {
		name           string
		config         QueryConfig
		context        map[string]interface{}
		mockCredential *credential.DecryptedValue
		setupMock      func(mock sqlmock.Sqlmock)
		wantErr        bool
		errorContains  string
		validate       func(t *testing.T, output *actions.ActionOutput)
	}{
		{
			name: "successful query execution",
			config: QueryConfig{
				Query:      "SELECT id, name, email FROM users WHERE active = ?",
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
					"connection_string": "root:password@tcp(localhost:3306)/testdb?parseTime=true",
				},
			},
			setupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "name", "email"}).
					AddRow(1, "John Doe", "john@example.com").
					AddRow(2, "Jane Smith", "jane@example.com")
				mock.ExpectQuery("SELECT id, name, email FROM users WHERE active = \\?").
					WithArgs(true).
					WillReturnRows(rows)
			},
			wantErr: false,
			validate: func(t *testing.T, output *actions.ActionOutput) {
				assert.NotNil(t, output)
				result, ok := output.Data.(*QueryResult)
				require.True(t, ok)
				assert.Equal(t, 2, result.RowCount)
				assert.Len(t, result.Rows, 2)
				assert.Equal(t, "John Doe", result.Rows[0]["name"])
			},
		},
		{
			name: "query with LIMIT",
			config: QueryConfig{
				Query:      "SELECT * FROM users LIMIT ?",
				Parameters: []interface{}{10},
			},
			context: map[string]interface{}{
				"env": map[string]interface{}{
					"tenant_id": "tenant-123",
				},
				"credential_id": "cred-123",
			},
			mockCredential: &credential.DecryptedValue{
				Value: map[string]interface{}{
					"connection_string": "root:password@tcp(localhost:3306)/testdb",
				},
			},
			setupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id"}).AddRow(1)
				mock.ExpectQuery("SELECT \\* FROM users LIMIT \\?").
					WithArgs(10).
					WillReturnRows(rows)
			},
			wantErr: false,
		},
		{
			name: "query with datetime columns",
			config: QueryConfig{
				Query:      "SELECT id, created_at, updated_at FROM users WHERE id = ?",
				Parameters: []interface{}{1},
			},
			context: map[string]interface{}{
				"env": map[string]interface{}{
					"tenant_id": "tenant-123",
				},
				"credential_id": "cred-123",
			},
			mockCredential: &credential.DecryptedValue{
				Value: map[string]interface{}{
					"connection_string": "root:password@tcp(localhost:3306)/testdb?parseTime=true",
				},
			},
			setupMock: func(mock sqlmock.Sqlmock) {
				now := time.Now()
				rows := sqlmock.NewRows([]string{"id", "created_at", "updated_at"}).
					AddRow(1, now, now)
				mock.ExpectQuery("SELECT id, created_at, updated_at FROM users WHERE id = \\?").
					WithArgs(1).
					WillReturnRows(rows)
			},
			wantErr: false,
			validate: func(t *testing.T, output *actions.ActionOutput) {
				result, ok := output.Data.(*QueryResult)
				require.True(t, ok)
				assert.Equal(t, 1, result.RowCount)
				assert.NotNil(t, result.Rows[0]["created_at"])
			},
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
			action := NewMySQLQueryAction(mockCred)
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

// TestMySQLStatementAction_Execute tests the MySQL statement action
func TestMySQLStatementAction_Execute(t *testing.T) {
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
			name: "successful INSERT with AUTO_INCREMENT",
			config: StatementConfig{
				Statement:  "INSERT INTO users (name, email) VALUES (?, ?)",
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
					"connection_string": "root:password@tcp(localhost:3306)/testdb",
				},
			},
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("INSERT INTO users").
					WithArgs("John Doe", "john@example.com").
					WillReturnResult(sqlmock.NewResult(42, 1))
			},
			wantErr: false,
			validate: func(t *testing.T, output *actions.ActionOutput) {
				result, ok := output.Data.(*StatementResult)
				require.True(t, ok)
				assert.Equal(t, int64(1), result.RowsAffected)
				assert.Equal(t, int64(42), result.LastInsertID)
			},
		},
		{
			name: "successful UPDATE with multiple rows",
			config: StatementConfig{
				Statement:  "UPDATE users SET active = ? WHERE created_at < ?",
				Parameters: []interface{}{false, "2020-01-01"},
			},
			context: map[string]interface{}{
				"env": map[string]interface{}{
					"tenant_id": "tenant-123",
				},
				"credential_id": "cred-123",
			},
			mockCredential: &credential.DecryptedValue{
				Value: map[string]interface{}{
					"connection_string": "root:password@tcp(localhost:3306)/testdb",
				},
			},
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("UPDATE users SET active").
					WithArgs(false, "2020-01-01").
					WillReturnResult(sqlmock.NewResult(0, 5))
			},
			wantErr: false,
			validate: func(t *testing.T, output *actions.ActionOutput) {
				result, ok := output.Data.(*StatementResult)
				require.True(t, ok)
				assert.Equal(t, int64(5), result.RowsAffected)
			},
		},
		{
			name: "INSERT with ON DUPLICATE KEY UPDATE",
			config: StatementConfig{
				Statement:  "INSERT INTO users (id, name, email) VALUES (?, ?, ?) ON DUPLICATE KEY UPDATE name = VALUES(name), email = VALUES(email)",
				Parameters: []interface{}{1, "Updated Name", "updated@example.com"},
			},
			context: map[string]interface{}{
				"env": map[string]interface{}{
					"tenant_id": "tenant-123",
				},
				"credential_id": "cred-123",
			},
			mockCredential: &credential.DecryptedValue{
				Value: map[string]interface{}{
					"connection_string": "root:password@tcp(localhost:3306)/testdb",
				},
			},
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("INSERT INTO users").
					WithArgs(1, "Updated Name", "updated@example.com").
					WillReturnResult(sqlmock.NewResult(0, 2)) // 2 affected rows for update
			},
			wantErr: false,
			validate: func(t *testing.T, output *actions.ActionOutput) {
				result, ok := output.Data.(*StatementResult)
				require.True(t, ok)
				assert.Equal(t, int64(2), result.RowsAffected)
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
					return tt.mockCredential, nil
				},
			}

			// Create action
			action := NewMySQLStatementAction(mockCred)
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

// TestMySQLTransactionAction_Execute tests the MySQL transaction action
func TestMySQLTransactionAction_Execute(t *testing.T) {
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
			name: "successful transaction with BEGIN and COMMIT",
			config: TransactionConfig{
				Statements: []TransactionStatement{
					{
						Statement:  "INSERT INTO orders (user_id, total) VALUES (?, ?)",
						Parameters: []interface{}{1, 100.50},
					},
					{
						Statement:  "UPDATE users SET total_spent = total_spent + ? WHERE id = ?",
						Parameters: []interface{}{100.50, 1},
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
					"connection_string": "root:password@tcp(localhost:3306)/testdb",
				},
			},
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO orders").
					WithArgs(1, 100.50).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectExec("UPDATE users SET total_spent").
					WithArgs(100.50, 1).
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
			name: "transaction rollback with foreign key constraint violation",
			config: TransactionConfig{
				Statements: []TransactionStatement{
					{
						Statement:  "INSERT INTO orders (user_id, total) VALUES (?, ?)",
						Parameters: []interface{}{999, 50.0}, // Non-existent user
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
					"connection_string": "root:password@tcp(localhost:3306)/testdb",
				},
			},
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO orders").
					WithArgs(999, 50.0).
					WillReturnError(errors.New("foreign key constraint fails"))
				mock.ExpectRollback()
			},
			wantErr:       true,
			errorContains: "foreign key constraint fails",
		},
		{
			name: "transaction with deadlock detection",
			config: TransactionConfig{
				Statements: []TransactionStatement{
					{
						Statement:  "UPDATE accounts SET balance = balance - ? WHERE id = ?",
						Parameters: []interface{}{100.0, 1},
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
					"connection_string": "root:password@tcp(localhost:3306)/testdb",
				},
			},
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec("UPDATE accounts SET balance").
					WithArgs(100.0, 1).
					WillReturnError(errors.New("deadlock found when trying to get lock"))
				mock.ExpectRollback()
			},
			wantErr:       true,
			errorContains: "deadlock",
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
			action := NewMySQLTransactionAction(mockCred)
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

// TestMySQLActions_ConnectionPooling tests MySQL connection pooling
func TestMySQLActions_ConnectionPooling(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	// Configure connection pool
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(time.Hour)

	rows := sqlmock.NewRows([]string{"id"}).AddRow(1)
	mock.ExpectQuery("SELECT").WillReturnRows(rows)

	mockCred := &MockCredentialService{
		GetValueFunc: func(ctx context.Context, tenantID, credentialID, userID string) (*credential.DecryptedValue, error) {
			return &credential.DecryptedValue{
				Value: map[string]interface{}{
					"connection_string": "root:password@tcp(localhost:3306)/testdb",
				},
			}, nil
		},
	}

	action := NewMySQLQueryAction(mockCred)
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
	assert.LessOrEqual(t, stats.MaxOpenConnections, 10)
}
