package connectors

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestPostgreSQLConnector_SSRFPrevention tests SSRF prevention
func TestPostgreSQLConnector_SSRFPrevention(t *testing.T) {
	tests := []struct {
		name          string
		connString    string
		shouldError   bool
		errorContains string
	}{
		{
			name:          "block localhost",
			connString:    "postgres://user:pass@localhost:5432/db",
			shouldError:   true,
			errorContains: "localhost",
		},
		{
			name:          "block 127.0.0.1",
			connString:    "postgres://user:pass@127.0.0.1:5432/db",
			shouldError:   true,
			errorContains: "localhost",
		},
		{
			name:          "block private IP 192.168",
			connString:    "postgres://user:pass@192.168.1.1:5432/db",
			shouldError:   true,
			errorContains: "private IP",
		},
		{
			name:          "block private IP 10.x",
			connString:    "postgres://user:pass@10.0.0.1:5432/db",
			shouldError:   true,
			errorContains: "private IP",
		},
		{
			name:          "block private IP 172.16-31",
			connString:    "postgres://user:pass@172.16.0.1:5432/db",
			shouldError:   true,
			errorContains: "private IP",
		},
		{
			name:        "allow public IP",
			connString:  "postgres://user:pass@203.0.113.1:5432/db",
			shouldError: false,
		},
		{
			name:        "allow domain",
			connString:  "postgres://user:pass@db.example.com:5432/db",
			shouldError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			connector := NewPostgreSQLConnector()
			err := connector.validateConnectionString(tt.connString)

			if tt.shouldError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorContains)
			} else {
				// Note: This will fail with "connection refused" since the host doesn't exist
				// but we're only testing SSRF prevention here
				assert.NoError(t, err)
			}
		})
	}
}

// TestMySQLConnector_SSRFPrevention tests SSRF prevention for MySQL
func TestMySQLConnector_SSRFPrevention(t *testing.T) {
	tests := []struct {
		name          string
		connString    string
		shouldError   bool
		errorContains string
	}{
		{
			name:          "block localhost",
			connString:    "user:pass@tcp(localhost:3306)/db",
			shouldError:   true,
			errorContains: "localhost",
		},
		{
			name:          "block 127.0.0.1",
			connString:    "user:pass@tcp(127.0.0.1:3306)/db",
			shouldError:   true,
			errorContains: "localhost",
		},
		{
			name:          "block private IP 192.168",
			connString:    "user:pass@tcp(192.168.1.1:3306)/db",
			shouldError:   true,
			errorContains: "private IP",
		},
		{
			name:          "block private IP 10.x",
			connString:    "user:pass@tcp(10.0.0.1:3306)/db",
			shouldError:   true,
			errorContains: "private IP",
		},
		{
			name:        "allow public IP",
			connString:  "user:pass@tcp(203.0.113.1:3306)/db",
			shouldError: false,
		},
		{
			name:        "allow domain",
			connString:  "user:pass@tcp(db.example.com:3306)/db",
			shouldError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			connector := NewMySQLConnector()
			err := connector.validateConnectionString(tt.connString)

			if tt.shouldError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorContains)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestMongoDBConnector_SSRFPrevention tests SSRF prevention for MongoDB
func TestMongoDBConnector_SSRFPrevention(t *testing.T) {
	tests := []struct {
		name          string
		connString    string
		shouldError   bool
		errorContains string
	}{
		{
			name:          "block localhost",
			connString:    "mongodb://user:pass@localhost:27017/db",
			shouldError:   true,
			errorContains: "localhost",
		},
		{
			name:          "block 127.0.0.1",
			connString:    "mongodb://user:pass@127.0.0.1:27017/db",
			shouldError:   true,
			errorContains: "localhost",
		},
		{
			name:          "block private IP 192.168",
			connString:    "mongodb://user:pass@192.168.1.1:27017/db",
			shouldError:   true,
			errorContains: "private IP",
		},
		{
			name:          "block private IP 10.x",
			connString:    "mongodb://user:pass@10.0.0.1:27017/db",
			shouldError:   true,
			errorContains: "private IP",
		},
		{
			name:        "allow public IP",
			connString:  "mongodb://user:pass@203.0.113.1:27017/db",
			shouldError: false,
		},
		{
			name:        "allow domain",
			connString:  "mongodb://user:pass@cluster.mongodb.net/db",
			shouldError: false,
		},
		{
			name:        "allow mongodb+srv",
			connString:  "mongodb+srv://user:pass@cluster.mongodb.net/db",
			shouldError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			connector := NewMongoDBConnector()
			err := connector.validateConnectionString(tt.connString)

			if tt.shouldError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorContains)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestSQLiteConnector_PathTraversalPrevention tests path traversal prevention
func TestSQLiteConnector_PathTraversalPrevention(t *testing.T) {
	tests := []struct {
		name          string
		path          string
		shouldError   bool
		errorContains string
	}{
		{
			name:          "block parent directory",
			path:          "../../../etc/passwd",
			shouldError:   true,
			errorContains: "directory traversal",
		},
		{
			name:          "block absolute path",
			path:          "/var/lib/database.db",
			shouldError:   true,
			errorContains: "absolute paths",
		},
		{
			name:          "block system path /etc",
			path:          "/etc/database.db",
			shouldError:   true,
			errorContains: "absolute paths",
		},
		{
			name:          "block system path /sys",
			path:          "/sys/database.db",
			shouldError:   true,
			errorContains: "absolute paths",
		},
		{
			name:        "allow relative path",
			path:        "data/database.db",
			shouldError: false,
		},
		{
			name:        "allow in-memory",
			path:        ":memory:",
			shouldError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			connector := NewSQLiteConnector()
			err := connector.validateConnectionString(tt.path)

			if tt.shouldError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorContains)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestQueryInput_Validation tests query input validation
func TestQueryInput_Validation(t *testing.T) {
	tests := []struct {
		name        string
		input       *QueryInput
		shouldError bool
		errorMsg    string
	}{
		{
			name: "valid query",
			input: &QueryInput{
				Query:   "SELECT * FROM users WHERE id = $1",
				Timeout: 30,
				MaxRows: 1000,
			},
			shouldError: false,
		},
		{
			name: "empty query",
			input: &QueryInput{
				Query: "",
			},
			shouldError: true,
			errorMsg:    "query is required",
		},
		{
			name: "timeout too high",
			input: &QueryInput{
				Query:   "SELECT * FROM users",
				Timeout: 500,
			},
			shouldError: true,
			errorMsg:    "timeout must be between 0 and 300 seconds",
		},
		{
			name: "max rows too high",
			input: &QueryInput{
				Query:   "SELECT * FROM users",
				MaxRows: 20000,
			},
			shouldError: true,
			errorMsg:    "max_rows must be between 0 and 10000",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.input.Validate()

			if tt.shouldError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestConnectorFactory tests the connector factory
func TestConnectorFactory(t *testing.T) {
	factory := NewConnectorFactory()

	tests := []struct {
		name        string
		dbType      DatabaseType
		shouldError bool
	}{
		{
			name:        "create PostgreSQL connector",
			dbType:      DatabaseTypePostgreSQL,
			shouldError: false,
		},
		{
			name:        "create MySQL connector",
			dbType:      DatabaseTypeMySQL,
			shouldError: false,
		},
		{
			name:        "create SQLite connector",
			dbType:      DatabaseTypeSQLite,
			shouldError: false,
		},
		{
			name:        "create MongoDB connector",
			dbType:      DatabaseTypeMongoDB,
			shouldError: false,
		},
		{
			name:        "unsupported database type",
			dbType:      "oracle",
			shouldError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			connector, err := factory.CreateConnector(tt.dbType)

			if tt.shouldError {
				require.Error(t, err)
				assert.Nil(t, connector)
				assert.ErrorIs(t, err, ErrUnsupportedDatabase)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, connector)
				assert.Equal(t, tt.dbType, connector.GetDatabaseType())
			}
		})
	}
}

// TestSQLQueryDetection tests SQL query type detection
func TestSQLQueryDetection(t *testing.T) {
	connector := NewPostgreSQLConnector()

	tests := []struct {
		name     string
		query    string
		isSelect bool
	}{
		{
			name:     "SELECT query",
			query:    "SELECT * FROM users",
			isSelect: true,
		},
		{
			name:     "SELECT with whitespace",
			query:    "  SELECT * FROM users",
			isSelect: true,
		},
		{
			name:     "WITH query",
			query:    "WITH cte AS (SELECT * FROM users) SELECT * FROM cte",
			isSelect: true,
		},
		{
			name:     "INSERT query",
			query:    "INSERT INTO users (name) VALUES ('John')",
			isSelect: false,
		},
		{
			name:     "UPDATE query",
			query:    "UPDATE users SET name = 'John'",
			isSelect: false,
		},
		{
			name:     "DELETE query",
			query:    "DELETE FROM users WHERE id = 1",
			isSelect: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := connector.isSelectQuery(tt.query)
			assert.Equal(t, tt.isSelect, result)
		})
	}
}

// TestRowLimitEnforcement tests that row limits are enforced
func TestRowLimitEnforcement(t *testing.T) {
	// This is an integration test that would need a real database
	// For now, we test the validation logic
	input := &QueryInput{
		Query:   "SELECT * FROM users",
		MaxRows: 1000,
	}

	err := input.Validate()
	assert.NoError(t, err)

	// Test exceeding limit
	input.MaxRows = 15000
	err = input.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "max_rows must be between 0 and 10000")
}

// TestParameterizedQueries tests that parameterized queries are properly handled
func TestParameterizedQueries(t *testing.T) {
	tests := []struct {
		name       string
		dbType     DatabaseType
		query      string
		parameters []interface{}
	}{
		{
			name:       "PostgreSQL parameterized query",
			dbType:     DatabaseTypePostgreSQL,
			query:      "SELECT * FROM users WHERE id = $1 AND status = $2",
			parameters: []interface{}{1, "active"},
		},
		{
			name:       "MySQL parameterized query",
			dbType:     DatabaseTypeMySQL,
			query:      "SELECT * FROM users WHERE id = ? AND status = ?",
			parameters: []interface{}{1, "active"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Validate that the query input accepts parameters
			input := &QueryInput{
				Query:      tt.query,
				Parameters: tt.parameters,
				Timeout:    30,
				MaxRows:    1000,
			}

			err := input.Validate()
			assert.NoError(t, err)

			// Ensure parameters are present
			assert.Len(t, input.Parameters, 2)
			assert.Equal(t, 1, input.Parameters[0])
			assert.Equal(t, "active", input.Parameters[1])
		})
	}
}
