# Database Connectors Package

## Overview

This package provides secure database connectivity for Gorax workflows, supporting PostgreSQL, MySQL, SQLite, and MongoDB.

## Security Features

- **SQL Injection Prevention**: Parameterized queries only
- **SSRF Protection**: Blocks connections to localhost and private IPs
- **Path Traversal Prevention**: SQLite path validation
- **Resource Limits**: Timeout and row limits enforced
- **Connection Pooling**: Efficient connection management

## Usage

### Creating a Connector

```go
factory := connectors.NewConnectorFactory()
connector, err := factory.CreateConnector(connectors.DatabaseTypePostgreSQL)
if err != nil {
    return err
}
defer connector.Close()
```

### Connecting to Database

```go
ctx := context.Background()
connString := "postgres://user:pass@db.example.com:5432/mydb"
if err := connector.Connect(ctx, connString); err != nil {
    return err
}
```

### Executing Queries

```go
// SELECT query
input := &connectors.QueryInput{
    Query:      "SELECT * FROM users WHERE status = $1",
    Parameters: []interface{}{"active"},
    Timeout:    30,
    MaxRows:    1000,
}

result, err := connector.Query(ctx, input)
if err != nil {
    return err
}

// Access results
for _, row := range result.Rows {
    fmt.Printf("User: %v\n", row)
}
```

### Executing Commands

```go
// INSERT/UPDATE/DELETE
input := &connectors.QueryInput{
    Query:      "INSERT INTO logs (message, level) VALUES ($1, $2)",
    Parameters: []interface{}{"Test message", "info"},
    Timeout:    30,
}

result, err := connector.Execute(ctx, input)
if err != nil {
    return err
}

fmt.Printf("Rows affected: %d\n", result.RowsAffected)
```

## Connector Types

### PostgreSQL

```go
connector := connectors.NewPostgreSQLConnector()
```

**Connection String:**
```
postgres://username:password@host:port/database?sslmode=require
```

**Parameterized Queries:** `$1`, `$2`, `$3`

### MySQL

```go
connector := connectors.NewMySQLConnector()
```

**Connection String:**
```
username:password@tcp(host:port)/database
```

**Parameterized Queries:** `?` placeholders

### SQLite

```go
connector := connectors.NewSQLiteConnector()
```

**Connection String:**
```
data/mydb.db
:memory:
```

**Parameterized Queries:** `?` placeholders

**Note:** Only relative paths allowed for security.

### MongoDB

```go
connector := connectors.NewMongoDBConnector()
```

**Connection String:**
```
mongodb://username:password@host:port/database
mongodb+srv://username:password@cluster.mongodb.net/database
```

**Query Format:** JSON-based

```go
input := &connectors.QueryInput{
    Query: `{"collection": "users", "filter": {"status": "active"}}`,
}
```

## Testing

Run security tests:

```bash
go test ./internal/database/connectors/... -v
```

All tests validate:
- SSRF prevention
- Path traversal prevention
- Query type detection
- Parameter handling
- Resource limit enforcement

## Files

- `types.go` - Core types and interfaces
- `factory.go` - Connector factory
- `postgresql.go` - PostgreSQL implementation
- `mysql.go` - MySQL implementation
- `sqlite.go` - SQLite implementation
- `mongodb.go` - MongoDB implementation
- `security_test.go` - Security validation tests

## Error Handling

Common errors:
- `ErrInvalidConnectionString` - Invalid connection format
- `ErrConnectionFailed` - Connection failed
- `ErrQueryTimeout` - Query exceeded timeout
- `ErrQueryFailed` - Query execution failed
- `ErrRowLimitExceeded` - Result set too large
- `ErrUnsupportedDatabase` - Database type not supported
- `ErrInvalidQuery` - Query validation failed

## Configuration Limits

| Limit | Default | Maximum |
|-------|---------|---------|
| Query Timeout | 30s | 300s |
| Max Rows | 1,000 | 10,000 |
| Connection Pool (SQL) | 25 | - |
| Idle Connections | 5 | - |
| Connection Lifetime | 5 min | - |

## Security Considerations

1. **Never use string interpolation** for queries
2. **Always use parameterized queries** with placeholders
3. **Set appropriate timeouts** to prevent resource exhaustion
4. **Limit result sets** with max_rows
5. **Use least privilege credentials** (read-only when possible)
6. **Monitor query performance** and execution times

## Architecture

```
Connector Interface
    ├── Connect(ctx, connectionString)
    ├── Close()
    ├── Ping(ctx)
    ├── Query(ctx, input)    // SELECT only
    ├── Execute(ctx, input)  // INSERT/UPDATE/DELETE
    └── GetDatabaseType()

Implementations:
    ├── PostgreSQLConnector
    ├── MySQLConnector
    ├── SQLiteConnector
    └── MongoDBConnector
```

## Integration with Workflow Actions

This package is used by:
- `internal/executor/actions/database/sql_query.go`
- `internal/executor/actions/database/sql_execute.go`
- `internal/executor/actions/database/mongodb.go`

Actions retrieve credentials and pass connection strings to connectors.
