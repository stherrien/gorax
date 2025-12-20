# Database Integrations

This package provides secure database connector integrations for the Gorax workflow automation platform.

## Supported Databases

### PostgreSQL
- **postgres:query** - Execute SELECT queries
- **postgres:statement** - Execute INSERT/UPDATE/DELETE statements
- **postgres:transaction** - Execute multiple statements in a transaction

### MySQL
- **mysql:query** - Execute SELECT queries
- **mysql:statement** - Execute INSERT/UPDATE/DELETE statements
- **mysql:transaction** - Execute multiple statements in a transaction

### MongoDB
- **mongodb:find** - Find documents in a collection
- **mongodb:insert** - Insert documents into a collection
- **mongodb:update** - Update documents in a collection
- **mongodb:delete** - Delete documents from a collection
- **mongodb:aggregate** - Execute aggregation pipelines

## Security Features

All database connectors implement the following security features:

1. **Secure Credential Management**: Connection strings are stored encrypted in the credential vault
2. **Parameterized Queries**: All SQL queries use parameterized statements to prevent SQL injection
3. **Connection Pooling**: Proper connection pool configuration for performance and resource management
4. **TLS Support**: Connection strings can specify TLS/SSL parameters
5. **Timeout Handling**: All operations support configurable timeouts

## Usage Examples

### PostgreSQL Query

```json
{
  "action": "postgres:query",
  "config": {
    "query": "SELECT id, name, email FROM users WHERE active = $1",
    "parameters": [true],
    "timeout": 30
  },
  "context": {
    "credential_id": "postgres-prod-cred"
  }
}
```

### MySQL Statement

```json
{
  "action": "mysql:statement",
  "config": {
    "statement": "INSERT INTO users (name, email) VALUES (?, ?)",
    "parameters": ["John Doe", "john@example.com"],
    "timeout": 30
  },
  "context": {
    "credential_id": "mysql-prod-cred"
  }
}
```

### PostgreSQL Transaction

```json
{
  "action": "postgres:transaction",
  "config": {
    "statements": [
      {
        "statement": "INSERT INTO orders (user_id, total) VALUES ($1, $2)",
        "parameters": [123, 100.50]
      },
      {
        "statement": "UPDATE users SET total_spent = total_spent + $1 WHERE id = $2",
        "parameters": [100.50, 123]
      }
    ],
    "timeout": 60
  },
  "context": {
    "credential_id": "postgres-prod-cred"
  }
}
```

### MongoDB Find

```json
{
  "action": "mongodb:find",
  "config": {
    "collection": "users",
    "filter": {"active": true},
    "projection": {"name": 1, "email": 1},
    "sort": {"created_at": -1},
    "limit": 10
  },
  "context": {
    "credential_id": "mongodb-prod-cred"
  }
}
```

### MongoDB Aggregate

```json
{
  "action": "mongodb:aggregate",
  "config": {
    "collection": "orders",
    "pipeline": [
      {"$match": {"status": "completed"}},
      {"$group": {
        "_id": "$user_id",
        "total": {"$sum": "$amount"}
      }},
      {"$sort": {"total": -1}}
    ]
  },
  "context": {
    "credential_id": "mongodb-prod-cred"
  }
}
```

## Credential Configuration

### PostgreSQL/MySQL Credentials

Store the following in the credential vault:

```json
{
  "connection_string": "postgresql://user:password@localhost:5432/dbname?sslmode=require"
}
```

For MySQL:
```json
{
  "connection_string": "user:password@tcp(localhost:3306)/dbname?parseTime=true&tls=true"
}
```

### MongoDB Credentials

```json
{
  "connection_string": "mongodb://user:password@localhost:27017/dbname?authSource=admin&tls=true",
  "database": "dbname"
}
```

## Testing

All connectors include comprehensive tests using mocks and sqlmock:

```bash
go test ./internal/integrations/database/...
```

Test coverage includes:
- Successful operations
- Error handling
- Connection failures
- SQL injection prevention
- Transaction rollback
- Context timeout handling
- Connection pooling

## Implementation Details

### Connection Pooling

All SQL connectors use connection pooling with the following defaults:
- **MaxOpenConns**: 10
- **MaxIdleConns**: 5
- **ConnMaxLifetime**: 1 hour

### Timeout Handling

Each operation supports an optional `timeout` parameter (in seconds). If not specified, the context timeout is used.

### Error Handling

All errors are wrapped with context information and returned to the workflow execution engine for proper handling and retry logic.

## Dependencies

- `github.com/lib/pq` - PostgreSQL driver
- `github.com/go-sql-driver/mysql` - MySQL driver
- `go.mongodb.org/mongo-driver` - MongoDB driver
- `github.com/DATA-DOG/go-sqlmock` - SQL mocking for tests

## Architecture

The database integrations follow the repository pattern:

1. **Models** (`models.go`) - Configuration and result types
2. **Connectors** (`postgres.go`, `mysql.go`, `mongodb.go`) - Database-specific implementations
3. **Actions** (`actions.go`) - Action wrappers for integration registry
4. **Tests** (`*_test.go`) - Comprehensive test suites following TDD

All integrations are registered with the global integration registry and can be used in workflow definitions.
