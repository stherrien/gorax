# Database Connectors for Gorax

## Overview

Gorax supports connecting to external databases from workflows, enabling workflows to query and manipulate data across PostgreSQL, MySQL, SQLite, and MongoDB databases.

## Features

- **Multiple Database Support**: PostgreSQL, MySQL, SQLite, MongoDB
- **Security First**: SQL injection prevention, SSRF protection, parameterized queries only
- **Resource Limits**: Automatic timeout and row limits
- **Connection Pooling**: Efficient connection management
- **Credential Management**: Secure storage of database credentials using envelope encryption

## Security Features

### SQL Injection Prevention

All database actions **require parameterized queries**. String interpolation is not supported for security reasons.

**PostgreSQL Example (SECURE):**
```sql
SELECT * FROM users WHERE email = $1 AND status = $2
```
Parameters: `["user@example.com", "active"]`

**MySQL Example (SECURE):**
```sql
SELECT * FROM orders WHERE user_id = ? AND created_at > ?
```
Parameters: `[123, "2024-01-01"]`

### SSRF Protection

Database connectors automatically block connections to:
- Localhost (`127.0.0.1`, `::1`, `localhost`)
- Private IP ranges (`192.168.x.x`, `10.x.x.x`, `172.16-31.x.x`)

This prevents Server-Side Request Forgery (SSRF) attacks that could access internal services.

### Path Traversal Prevention (SQLite)

SQLite connections block:
- Absolute paths
- Parent directory references (`..`)
- System directories (`/etc`, `/sys`, `/proc`, `/dev`)

## Database Types

### PostgreSQL

**Connection String Format:**
```
postgres://username:password@host:port/database?sslmode=require
```

**Example Credential Value:**
```json
{
  "connection_string": "postgres://myuser:mypassword@db.example.com:5432/myapp?sslmode=require"
}
```

**Parameterized Queries:** Use `$1`, `$2`, `$3`, etc.

### MySQL

**Connection String Format:**
```
username:password@tcp(host:port)/database?charset=utf8mb4&parseTime=True
```

**Example Credential Value:**
```json
{
  "connection_string": "root:mypassword@tcp(db.example.com:3306)/myapp?charset=utf8mb4"
}
```

**Parameterized Queries:** Use `?` placeholders

### SQLite

**Connection String Format:**
```
data/myapp.db
```
or
```
:memory:
```

**Example Credential Value:**
```json
{
  "connection_string": "data/workflows.db"
}
```

**Parameterized Queries:** Use `?` placeholders

**Security Note:** Only relative paths are allowed. Absolute paths are blocked.

### MongoDB

**Connection String Format:**
```
mongodb://username:password@host:port/database
mongodb+srv://username:password@cluster.mongodb.net/database
```

**Example Credential Value:**
```json
{
  "connection_string": "mongodb+srv://myuser:mypassword@cluster0.mongodb.net/myapp"
}
```

**Query Format:** JSON-based queries

## Creating Database Credentials

1. **Navigate to Credentials** in the Gorax UI
2. **Create New Credential**
3. **Select Type:**
   - `database_postgresql`
   - `database_mysql`
   - `database_sqlite`
   - `database_mongodb`
4. **Provide Connection String** in the value field:
   ```json
   {
     "connection_string": "your-connection-string-here"
   }
   ```

## Using Database Actions in Workflows

### SQL Query Action (SELECT)

Use this for **read-only** queries that return data.

**Action Type:** `sql_query`

**Configuration:**
```json
{
  "connection_string": "{{credentials.my_postgres_db.connection_string}}",
  "database_type": "postgresql",
  "query": "SELECT id, name, email FROM users WHERE status = $1 AND created_at > $2",
  "parameters": ["active", "2024-01-01"],
  "timeout": 30,
  "max_rows": 1000
}
```

**Output:**
```json
{
  "rows": [
    {"id": 1, "name": "John Doe", "email": "john@example.com"},
    {"id": 2, "name": "Jane Smith", "email": "jane@example.com"}
  ],
  "rows_affected": 2,
  "execution_ms": 45
}
```

### SQL Execute Action (INSERT/UPDATE/DELETE)

Use this for queries that **modify data**.

**Action Type:** `sql_execute`

**INSERT Example:**
```json
{
  "connection_string": "{{credentials.my_mysql_db.connection_string}}",
  "database_type": "mysql",
  "query": "INSERT INTO orders (user_id, product_id, quantity) VALUES (?, ?, ?)",
  "parameters": [123, 456, 2],
  "timeout": 30
}
```

**UPDATE Example:**
```json
{
  "connection_string": "{{credentials.my_postgres_db.connection_string}}",
  "database_type": "postgresql",
  "query": "UPDATE users SET status = $1 WHERE id = $2",
  "parameters": ["inactive", 123],
  "timeout": 30
}
```

**DELETE Example:**
```json
{
  "connection_string": "{{credentials.my_mysql_db.connection_string}}",
  "database_type": "mysql",
  "query": "DELETE FROM temp_data WHERE created_at < ?",
  "parameters": ["2024-01-01"],
  "timeout": 30
}
```

**Output:**
```json
{
  "rows_affected": 5,
  "execution_ms": 120
}
```

### MongoDB Action

**Action Type:** `mongodb`

**Find Example:**
```json
{
  "connection_string": "{{credentials.my_mongodb.connection_string}}",
  "operation": "find",
  "collection": "users",
  "filter": {"status": "active"},
  "sort": {"created_at": -1},
  "projection": {"password": 0},
  "max_rows": 100
}
```

**Insert One Example:**
```json
{
  "connection_string": "{{credentials.my_mongodb.connection_string}}",
  "operation": "insertOne",
  "collection": "users",
  "document": {
    "name": "John Doe",
    "email": "john@example.com",
    "status": "active",
    "created_at": "{{workflow.timestamp}}"
  }
}
```

**Update One Example:**
```json
{
  "connection_string": "{{credentials.my_mongodb.connection_string}}",
  "operation": "updateOne",
  "collection": "users",
  "filter": {"email": "john@example.com"},
  "update": {"$set": {"status": "inactive"}}
}
```

**Delete Many Example:**
```json
{
  "connection_string": "{{credentials.my_mongodb.connection_string}}",
  "operation": "deleteMany",
  "collection": "logs",
  "filter": {"created_at": {"$lt": "2024-01-01"}}
}
```

**Supported Operations:**
- `find` - Query documents
- `insertOne` - Insert single document
- `insertMany` - Insert multiple documents
- `updateOne` - Update single document
- `updateMany` - Update multiple documents
- `deleteOne` - Delete single document
- `deleteMany` - Delete multiple documents

## Resource Limits

### Timeouts

- **Default:** 30 seconds
- **Maximum:** 300 seconds (5 minutes)
- **Configurable** per query

### Row Limits

- **Default:** 1,000 rows
- **Maximum:** 10,000 rows
- **Configurable** per query

Queries that exceed the row limit will fail with an error.

### Connection Pooling

- **PostgreSQL/MySQL:** Max 25 connections, 5 idle
- **SQLite:** Max 1 connection (single-writer)
- **MongoDB:** Max 25 connections, 5 idle

## Best Practices

### 1. Always Use Parameterized Queries

**BAD (INSECURE):**
```sql
SELECT * FROM users WHERE email = '${user_input}'
```

**GOOD (SECURE):**
```sql
SELECT * FROM users WHERE email = $1
```
Parameters: `["${user_input}"]`

### 2. Limit Result Sets

Always set appropriate `max_rows` limits:
```json
{
  "max_rows": 100
}
```

### 3. Set Appropriate Timeouts

For long-running queries, increase timeout:
```json
{
  "timeout": 120
}
```

### 4. Use Read-Only Credentials

When possible, use database credentials with read-only permissions for SELECT queries.

### 5. Monitor Query Performance

Check execution times in workflow logs to identify slow queries.

## Error Handling

### Common Errors

**Connection Failed:**
```
failed to connect to database: connection refused
```
- Check connection string
- Verify database is accessible
- Check firewall rules

**Query Timeout:**
```
query timeout exceeded
```
- Increase timeout value
- Optimize query with indexes
- Reduce result set size

**Row Limit Exceeded:**
```
query returned more than 1000 rows
```
- Increase `max_rows` (up to 10,000)
- Add LIMIT clause to query
- Add more specific WHERE conditions

**SQL Injection Attempt:**
```
only SELECT queries allowed in Query method
```
- Use `sql_execute` for INSERT/UPDATE/DELETE
- Ensure correct action type

**SSRF Prevention:**
```
connections to localhost are not allowed for security reasons
```
- Cannot connect to internal/private IPs
- Use publicly accessible databases only

## Migration Schema

The database connector system requires running migration `029_database_connectors.sql`:

```sql
-- Database connection types
CREATE TYPE database_type AS ENUM ('postgresql', 'mysql', 'sqlite', 'mongodb');

-- Database connections table
CREATE TABLE database_connections (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL,
    name VARCHAR(255) NOT NULL,
    database_type database_type NOT NULL,
    credential_id UUID NOT NULL,
    connection_config JSONB NOT NULL DEFAULT '{}',
    status VARCHAR(50) NOT NULL DEFAULT 'active',
    last_tested_at TIMESTAMPTZ,
    last_used_at TIMESTAMPTZ,
    usage_count BIGINT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

## Testing

### Unit Tests

Security tests validate:
- SSRF prevention for all database types
- Path traversal prevention for SQLite
- Parameterized query support
- Query type detection (SELECT vs INSERT/UPDATE/DELETE)
- Row limit enforcement
- Timeout validation

Run tests:
```bash
go test ./internal/database/connectors/... -v
```

### Integration Tests

Integration tests would require actual database instances. Use Docker for testing:

```bash
# PostgreSQL
docker run -d -p 5432:5432 -e POSTGRES_PASSWORD=testpass postgres:15

# MySQL
docker run -d -p 3306:3306 -e MYSQL_ROOT_PASSWORD=testpass mysql:8

# MongoDB
docker run -d -p 27017:27017 mongo:7
```

## Architecture

### Components

1. **Connectors** (`internal/database/connectors/`)
   - `types.go` - Core types and interfaces
   - `postgresql.go` - PostgreSQL implementation
   - `mysql.go` - MySQL implementation
   - `sqlite.go` - SQLite implementation
   - `mongodb.go` - MongoDB implementation
   - `factory.go` - Connector factory

2. **Actions** (`internal/executor/actions/database/`)
   - `sql_query.go` - SELECT query action
   - `sql_execute.go` - INSERT/UPDATE/DELETE action
   - `mongodb.go` - MongoDB operations action

3. **Credentials** (`internal/credential/`)
   - Added database credential types
   - Envelope encryption for connection strings

4. **Database Schema** (`migrations/029_database_connectors.sql`)
   - Connection metadata storage
   - Query audit logs

### Security Layers

1. **Connection String Validation**
   - SSRF protection
   - Path traversal prevention

2. **Query Validation**
   - Query type detection
   - Timeout enforcement
   - Row limit enforcement

3. **Parameterized Queries**
   - SQL injection prevention
   - Parameters passed separately

4. **Credential Encryption**
   - Envelope encryption at rest
   - Secure retrieval through credential service

## Future Enhancements

Potential future features:
- [ ] Connection manager API endpoints
- [ ] Connection testing UI
- [ ] Query builder UI component
- [ ] Visual database action nodes for workflow canvas
- [ ] Connection health monitoring
- [ ] Query performance analytics
- [ ] Read-only connection mode enforcement
- [ ] Query result caching
- [ ] Transaction support
- [ ] Stored procedure execution
- [ ] Additional database types (Oracle, SQL Server, Redshift)

## Troubleshooting

### Cannot Connect to Database

1. Verify connection string format
2. Check database is running and accessible
3. Verify credentials are correct
4. Check firewall/network rules
5. Review SSRF protection isn't blocking legitimate hosts

### Slow Query Performance

1. Check query execution time in logs
2. Add database indexes on filtered columns
3. Reduce result set with WHERE clauses
4. Increase timeout if needed
5. Use EXPLAIN to analyze query plan

### Memory Issues

1. Reduce `max_rows` limit
2. Add pagination to queries
3. Use streaming for large result sets
4. Close connections properly

## Support

For issues or questions:
- Check logs for detailed error messages
- Review security tests for examples
- Consult database-specific documentation
- Verify connection strings match expected format
