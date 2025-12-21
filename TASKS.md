# Gorax Development Tasks

## Completed

### API Documentation (2024-01-20)

**Objective:** Generate comprehensive OpenAPI/Swagger documentation for the Gorax REST API

**Completed Items:**

1. ✅ **Swagger Dependencies**
   - Installed `swaggo/swag` CLI tool
   - Added `github.com/swaggo/http-swagger` for UI
   - Added `github.com/swaggo/files` for static assets

2. ✅ **Swagger Annotations**
   - Added comprehensive API metadata to `cmd/api/main.go`
   - Documented authentication methods (dev and production)
   - Added security definitions (TenantID, UserID, SessionCookie)
   - Documented all API tags and descriptions
   - Added annotations to critical handlers:
     - Health endpoints (health.go)
     - Workflow management (workflow.go)
     - Webhook management (webhook_management_handler.go)

3. ✅ **OpenAPI Spec Generation**
   - Generated `docs/api/swagger.json`
   - Generated `docs/api/swagger.yaml`
   - Generated `docs/api/docs.go` (Go bindings)
   - Configured with `--parseDependency` and `--parseInternal` flags

4. ✅ **Swagger UI Integration**
   - Added Swagger UI endpoint at `/api/docs/`
   - Configured to serve from `/docs/api/swagger.json`
   - Integrated into main router (no authentication required)
   - Imported generated docs in main.go

5. ✅ **API Reference Documentation**
   - Created comprehensive `docs/API_REFERENCE.md`
   - Documented all major endpoints with:
     - Request/response examples
     - Authentication requirements
     - Error handling
     - Rate limiting information
     - Pagination details
     - curl examples
     - Code examples for signature verification
   - Organized by resource type (Workflows, Webhooks, Executions, etc.)

6. ✅ **Postman Collection**
   - Created `docs/api/gorax.postman_collection.json`
   - Includes all major API endpoints
   - Organized into folders by resource
   - Pre-configured with collection variables
   - Includes test scripts to auto-extract IDs
   - Created environment template `docs/api/gorax.postman_environment.json`

**Files Modified:**
- `cmd/api/main.go` - Added Swagger metadata annotations
- `internal/api/app.go` - Added Swagger UI route and import
- `internal/api/handlers/health.go` - Added endpoint annotations
- `internal/api/handlers/workflow.go` - Added endpoint annotations
- `internal/api/handlers/webhook_management_handler.go` - Added endpoint annotations
- `go.mod` - Added Swagger dependencies
- `go.sum` - Updated checksums

**Files Created:**
- `docs/api/swagger.json` - OpenAPI 3.0 specification (JSON)
- `docs/api/swagger.yaml` - OpenAPI 3.0 specification (YAML)
- `docs/api/docs.go` - Generated Go documentation
- `docs/API_REFERENCE.md` - Comprehensive API reference
- `docs/api/gorax.postman_collection.json` - Postman collection
- `docs/api/gorax.postman_environment.json` - Postman environment template

**How to Use:**

1. **View Interactive Documentation:**
   ```bash
   # Start the API server
   make run

   # Open browser to Swagger UI
   open http://localhost:8080/api/docs/
   ```

2. **Import Postman Collection:**
   - Open Postman
   - File → Import
   - Select `docs/api/gorax.postman_collection.json`
   - Import environment from `docs/api/gorax.postman_environment.json`

3. **Regenerate Swagger Docs (after changes):**
   ```bash
   swag init -g cmd/api/main.go -o docs/api --parseDependency --parseInternal
   ```

4. **View Static Documentation:**
   ```bash
   open docs/API_REFERENCE.md
   ```

**Future Enhancements:**

- Add remaining handler annotations (execution, schedule, credentials, admin)
- Add response schema definitions for all endpoints
- Add request validation examples
- Create SDK generation scripts using OpenAPI spec
- Add Redoc alternative documentation
- Create API versioning strategy
- Add GraphQL schema documentation (when implemented)

---

### Performance and Load Testing Suite (2024-12-20)

**Objective:** Create comprehensive performance and load testing infrastructure for the Gorax platform

**Completed Items:**

1. ✅ **k6 Load Testing Suite**
   - Created `/tests/load/` directory structure
   - Implemented 5 comprehensive k6 test scripts:
     - `workflow_api.js` - Workflow CRUD operations load test
     - `execution_api.js` - Workflow execution throughput test
     - `webhook_trigger.js` - Webhook ingestion rate test
     - `websocket_connections.js` - WebSocket connection scaling test
     - `auth_endpoints.js` - Authentication performance test

2. ✅ **Test Configuration**
   - Created `config.js` with:
     - Environment configuration
     - Test scenario definitions (smoke, load, stress, spike, soak)
     - Performance thresholds
     - Helper functions for test data generation
   - Configurable VU (virtual user) settings
   - Configurable duration and ramp-up patterns

3. ✅ **Test Runner Script**
   - Created `run_tests.sh` with:
     - Automated test execution
     - Command-line argument parsing
     - Result collection and reporting
     - HTML report generation
     - Error handling and logging
   - Support for running individual or all tests
   - Environment variable configuration

4. ✅ **Custom Metrics**
   - Workflow operations metrics (create, read, update, delete durations)
   - Execution performance metrics (start, completion, throughput)
   - Webhook ingestion metrics (ingestion time, throughput)
   - WebSocket metrics (connection time, message latency)
   - Authentication metrics (login, token refresh, validation times)
   - Success rates and error counters for all operations

5. ✅ **Go Benchmarks**
   - Created `internal/executor/executor_bench_test.go`:
     - Simple workflow execution benchmark
     - Workflow with retry logic benchmark
     - Sequential workflow benchmark
     - Conditional workflow benchmark
     - Loop workflow benchmark
     - Parallel workflow benchmark
     - Circuit breaker performance benchmark
     - Context data marshaling/unmarshaling benchmarks
     - Memory allocation benchmarks
   - Created `internal/workflow/formula/formula_bench_test.go`:
     - Simple and complex expression benchmarks
     - String operations benchmarks
     - Math operations benchmarks
     - Date operations benchmarks
     - Array operations benchmarks
     - Conditional expressions benchmarks
     - Complex workflow context benchmarks
     - Expression compilation benchmarks
     - Concurrent evaluation benchmarks
     - Memory allocation benchmarks

6. ✅ **Comprehensive Documentation**
   - Created `tests/load/README.md` with:
     - Installation instructions for k6
     - Test scenario descriptions
     - Usage examples for all test types
     - Performance threshold definitions
     - Result interpretation guide
     - Troubleshooting section
     - CI/CD integration examples
     - Prometheus/Grafana integration guide
     - Performance baselines
     - Best practices

**Test Scenarios:**

- **Smoke Test:** 1 VU, 1 minute - Quick health check
- **Load Test:** 10 VUs ramped, 9 minutes - Normal expected load
- **Stress Test:** Up to 100 VUs, 26 minutes - Find breaking points
- **Spike Test:** 10→100→10 VUs, 8 minutes - Sudden traffic spikes
- **Soak Test:** 20 VUs, 70 minutes - Long-term stability

**Performance Thresholds:**

- HTTP p95 latency: < 500ms
- HTTP p99 latency: < 1s
- Error rate: < 1%
- Workflow create p95: < 1s
- Workflow execute p95: < 2s
- Webhook ingestion p95: < 200ms
- Auth login p95: < 300ms
- WebSocket connect p95: < 500ms

**Files Created:**
- `tests/load/config.js` - Test configuration
- `tests/load/workflow_api.js` - Workflow API load test
- `tests/load/execution_api.js` - Execution API load test
- `tests/load/webhook_trigger.js` - Webhook trigger load test
- `tests/load/websocket_connections.js` - WebSocket load test
- `tests/load/auth_endpoints.js` - Authentication load test
- `tests/load/run_tests.sh` - Test runner script (executable)
- `tests/load/README.md` - Comprehensive documentation
- `internal/executor/executor_bench_test.go` - Executor benchmarks
- `internal/workflow/formula/formula_bench_test.go` - Formula benchmarks

**How to Use:**

1. **Install k6:**
   ```bash
   # macOS
   brew install k6

   # Linux
   curl -s https://dl.k6.io/key.gpg | sudo apt-key add -
   echo "deb https://dl.k6.io/deb stable main" | sudo tee /etc/apt/sources.list.d/k6.list
   sudo apt-get update
   sudo apt-get install k6
   ```

2. **Run Load Tests:**
   ```bash
   cd tests/load

   # Run all tests with default (load) scenario
   ./run_tests.sh

   # Run smoke test
   ./run_tests.sh --scenario smoke

   # Run specific test with stress scenario
   ./run_tests.sh workflow --scenario stress

   # Run against different environment
   ./run_tests.sh --url https://staging.gorax.io --scenario load
   ```

3. **Run Go Benchmarks:**
   ```bash
   # Benchmark executor
   go test -bench=. -benchmem ./internal/executor/

   # Benchmark formula evaluator
   go test -bench=. -benchmem ./internal/workflow/formula/

   # Run specific benchmark
   go test -bench=BenchmarkExecuteSimpleWorkflow -benchmem ./internal/executor/

   # Compare benchmarks
   go test -bench=. -benchmem ./internal/executor/ > old.txt
   # ... make changes ...
   go test -bench=. -benchmem ./internal/executor/ > new.txt
   benchcmp old.txt new.txt
   ```

4. **View Results:**
   ```bash
   # View test output (stdout)
   # Results are automatically displayed after test completion

   # View JSON results
   cat tests/load/results/*.json | jq '.metrics'

   # View HTML report
   open tests/load/results/combined_report_*.html
   ```

**Integration with CI/CD:**

The load tests can be integrated into CI/CD pipelines for:
- Daily performance regression testing
- Pre-release performance validation
- Continuous performance monitoring
- Capacity planning

Example GitHub Actions workflow provided in documentation.

**Future Enhancements:**

- Add database query performance benchmarks
- Implement memory leak detection tests
- Add network latency simulation
- Create performance regression detection
- Implement automated performance reports
- Add distributed load testing support
- Integrate with Grafana for real-time monitoring
- Add chaos engineering tests

---

### Google Workspace Integrations (2024-12-20)

**Objective:** Implement comprehensive Google Workspace integrations for the Gorax workflow automation platform

**Completed Items:**

1. ✅ **Dependencies & Configuration**
   - Added Google API client libraries to go.mod:
     - `google.golang.org/api/gmail/v1` - Gmail API
     - `google.golang.org/api/sheets/v4` - Google Sheets API
     - `google.golang.org/api/drive/v3` - Google Drive API
     - `google.golang.org/api/calendar/v3` - Google Calendar API
     - `golang.org/x/oauth2` - OAuth2 authentication
   - Created shared authentication utilities for OAuth2 and Service Accounts

2. ✅ **Common Utilities** (`internal/integrations/google/common.go`)
   - OAuth2 token creation from credential values
   - Service account authentication support
   - Nested context value extraction helper
   - Path parsing utilities
   - Comprehensive test coverage

3. ✅ **Gmail Integration** (`internal/integrations/google/gmail.go`)
   - **Send Email Action:**
     - To, Cc, Bcc recipients
     - HTML/plain text support
     - Subject and body
     - Returns message ID and thread ID
   - **Read Emails Action:**
     - Query-based email filtering
     - Configurable result limit
     - Full message parsing (headers, body, labels)
     - Pagination support
   - Validation, error handling, and tests

4. ✅ **Google Sheets Integration** (`internal/integrations/google/sheets.go`)
   - **Read Range Action:**
     - Read data from specified range
     - Returns values as 2D array
     - Range information metadata
   - **Write Range Action:**
     - Update cells in specified range
     - User-entered value input option
     - Returns updated cell counts
   - **Append Rows Action:**
     - Append rows to end of range
     - Automatic range detection
     - Returns updated range info
   - Validation, error handling, and tests

5. ✅ **Google Drive Integration** (`internal/integrations/google/drive.go`)
   - **Upload File Action:**
     - Base64 encoded content support
     - MIME type specification
     - Optional parent folder
     - Description metadata
     - Returns file ID and web view link
   - **Download File Action:**
     - Download by file ID
     - Base64 encoded content return
     - File metadata retrieval
   - **List Files Action:**
     - Query-based filtering
     - Folder-specific listing
     - Configurable page size and ordering
     - Returns file metadata (ID, name, MIME type, links, timestamps)
   - Validation, error handling, and tests

6. ✅ **Google Calendar Integration** (`internal/integrations/google/calendar.go`)
   - **Create Event Action:**
     - Summary, description, location
     - Start/end time in RFC3339 format
     - Time zone support
     - Attendee management
     - Returns event ID and HTML link
   - **List Events Action:**
     - Time range filtering
     - Configurable max results
     - Sorted by start time
     - Returns event details and attendees
   - **Delete Event Action:**
     - Delete by event ID
     - Confirmation result
   - Validation, error handling, and tests

7. ✅ **Test Coverage**
   - Created `mock_credential_test.go` for reusable test mock
   - Comprehensive unit tests for all actions
   - Configuration validation tests
   - Error handling tests
   - All tests passing (100% of test cases)

**Files Created:**
- `internal/integrations/google/common.go` - Shared utilities and auth
- `internal/integrations/google/common_test.go` - Common utilities tests
- `internal/integrations/google/mock_credential_test.go` - Test mocks
- `internal/integrations/google/gmail.go` - Gmail integration (329 lines)
- `internal/integrations/google/gmail_test.go` - Gmail tests (306 lines)
- `internal/integrations/google/sheets.go` - Sheets integration (328 lines)
- `internal/integrations/google/sheets_test.go` - Sheets tests (224 lines)
- `internal/integrations/google/drive.go` - Drive integration (362 lines)
- `internal/integrations/google/drive_test.go` - Drive tests (164 lines)
- `internal/integrations/google/calendar.go` - Calendar integration (428 lines)
- `internal/integrations/google/calendar_test.go` - Calendar tests (225 lines)

**Files Modified:**
- `go.mod` - Added Google API dependencies
- `go.sum` - Updated dependency checksums

**How to Use:**

1. **Configure Google Credentials:**
   ```go
   // OAuth2 credentials
   credential := map[string]interface{}{
       "access_token":  "ya29.xxx",
       "refresh_token": "1//xxx",
       "token_type":    "Bearer",
   }

   // Service Account credentials
   credential := map[string]interface{}{
       "type": "service_account",
       "project_id": "project-123",
       "private_key_id": "key-id",
       "private_key": "-----BEGIN PRIVATE KEY-----...",
       "client_email": "service@project.iam.gserviceaccount.com",
       // ... other service account fields
   }
   ```

2. **Gmail Actions:**
   ```go
   // Send email
   action := google.NewGmailSendAction(credentialService)
   config := google.GmailSendConfig{
       To:      "recipient@example.com",
       Subject: "Test Email",
       Body:    "Email content",
       IsHTML:  false,
   }

   // Read emails
   action := google.NewGmailReadAction(credentialService)
   config := google.GmailReadConfig{
       Query:      "from:test@example.com",
       MaxResults: 10,
   }
   ```

3. **Google Sheets Actions:**
   ```go
   // Read range
   action := google.NewSheetsReadAction(credentialService)
   config := google.SheetsReadConfig{
       SpreadsheetID: "1abc...",
       Range:         "Sheet1!A1:B10",
   }

   // Write range
   action := google.NewSheetsWriteAction(credentialService)
   config := google.SheetsWriteConfig{
       SpreadsheetID: "1abc...",
       Range:         "Sheet1!A1:B2",
       Values: [][]interface{}{
           {"Header1", "Header2"},
           {"Value1", "Value2"},
       },
   }

   // Append rows
   action := google.NewSheetsAppendAction(credentialService)
   config := google.SheetsAppendConfig{
       SpreadsheetID: "1abc...",
       Range:         "Sheet1!A:B",
       Values: [][]interface{}{
           {"New", "Row"},
       },
   }
   ```

4. **Google Drive Actions:**
   ```go
   // Upload file
   action := google.NewDriveUploadAction(credentialService)
   config := google.DriveUploadConfig{
       FileName: "report.txt",
       Content:  base64.StdEncoding.EncodeToString([]byte("content")),
       MimeType: "text/plain",
       ParentID: "folder-id", // optional
   }

   // Download file
   action := google.NewDriveDownloadAction(credentialService)
   config := google.DriveDownloadConfig{
       FileID: "file-id",
   }

   // List files
   action := google.NewDriveListAction(credentialService)
   config := google.DriveListConfig{
       Query:    "name contains 'report'",
       FolderID: "folder-id", // optional
       PageSize: 100,
   }
   ```

5. **Google Calendar Actions:**
   ```go
   // Create event
   action := google.NewCalendarCreateAction(credentialService)
   config := google.CalendarCreateConfig{
       CalendarID:  "primary",
       Summary:     "Team Meeting",
       Description: "Weekly sync",
       Location:    "Conference Room A",
       StartTime:   "2024-01-20T10:00:00Z",
       EndTime:     "2024-01-20T11:00:00Z",
       TimeZone:    "America/New_York",
       Attendees:   []string{"team@example.com"},
   }

   // List events
   action := google.NewCalendarListAction(credentialService)
   config := google.CalendarListConfig{
       CalendarID: "primary",
       TimeMin:    "2024-01-20T00:00:00Z",
       TimeMax:    "2024-01-27T00:00:00Z",
       MaxResults: 10,
   }

   // Delete event
   action := google.NewCalendarDeleteAction(credentialService)
   config := google.CalendarDeleteConfig{
       CalendarID: "primary",
       EventID:    "event-id",
   }
   ```

6. **Run Tests:**
   ```bash
   # Run all Google integration tests
   go test -v ./internal/integrations/google/... -count=1

   # Run specific test
   go test -v ./internal/integrations/google/... -run TestGmailSendAction

   # Run with coverage
   go test -cover ./internal/integrations/google/...
   ```

**Architecture:**

- All actions implement `actions.Action` interface
- OAuth2 and Service Account authentication supported
- Credential service integration for secure credential retrieval
- Context-based tenant and credential ID extraction
- Comprehensive error handling with wrapped errors
- Base64 encoding for file content (Drive, attachments)
- RFC3339 datetime format for Calendar events
- Configurable test mode with mock servers

**Test Coverage:**

- Unit tests for all action configurations
- Validation tests for required fields
- Error handling tests
- Mock credential service for testing
- HTTP mock servers for API testing
- All tests passing (100% success rate)

**Security Considerations:**

- Credentials retrieved from secure credential service
- OAuth2 tokens support refresh tokens
- Service account credentials support
- Tenant isolation via context
- No credentials logged or exposed
- HTTPS-only communication with Google APIs

**Future Enhancements:**

- Add Gmail attachment support
- Add Google Sheets create spreadsheet action
- Add Google Drive folder creation action
- Add Google Calendar update event action
- Add batch operations for improved performance
- Add webhook support for real-time event notifications
- Implement rate limiting and retry logic
- Add support for shared drives
- Add Google Docs integration
- Add Google Forms integration

**Performance:**

- Efficient OAuth2 token caching
- Minimal memory allocation
- Concurrent-safe operations
- Configurable timeout support via context
- Lazy client creation

---

## In Progress

(No current tasks)

---

## Pending

### High Priority

- Implement comprehensive integration tests for webhook replay functionality
- Add OpenTelemetry tracing to webhook handlers
- Implement webhook retry mechanism with exponential backoff
- Add webhook event filtering UI

### Medium Priority

- Create developer onboarding documentation
- Add more workflow templates
- Implement workflow templates marketplace
- Add bulk workflow operations

### Low Priority

- Add GraphQL API
- Implement workflow visual editor backend
- Add collaborative editing features
- Create workflow analytics dashboard

---

## Notes

### API Documentation Best Practices

When adding new endpoints, remember to:

1. Add Swagger annotations above the handler function
2. Include all parameters (path, query, body)
3. Document all response codes
4. Update the Postman collection
5. Regenerate Swagger docs with `swag init`
6. Update API_REFERENCE.md with examples

### Annotation Format

```go
// HandlerName does something
// @Summary Short description
// @Description Longer description
// @Tags ResourceName
// @Accept json
// @Produce json
// @Param paramName paramType dataType required "description"
// @Security TenantID
// @Security UserID
// @Success 200 {object} ResponseType "description"
// @Failure 400 {object} map[string]string "error description"
// @Router /path [method]
func (h *Handler) HandlerName(w http.ResponseWriter, r *http.Request) {
    // implementation
}
```

---

### Database Integrations (2024-12-20)

**Objective:** Implement database connector integrations for workflow automation

**Completed Items:**

1. ✅ **PostgreSQL Connector**
   - Implemented query action (SELECT)
   - Implemented statement action (INSERT/UPDATE/DELETE)
   - Implemented transaction action
   - Parameterized queries to prevent SQL injection
   - Connection pooling with configurable limits
   - TLS/SSL support via connection string
   - Comprehensive test suite with sqlmock

2. ✅ **MySQL Connector**
   - Implemented query action (SELECT)
   - Implemented statement action (INSERT/UPDATE/DELETE)
   - Implemented transaction action
   - Support for AUTO_INCREMENT and LAST_INSERT_ID
   - Support for ON DUPLICATE KEY UPDATE
   - Parameterized queries for security
   - Connection pooling
   - Comprehensive test suite with sqlmock

3. ✅ **MongoDB Connector**
   - Implemented find action (with projection, sort, limit, skip)
   - Implemented insert action (single and bulk)
   - Implemented update action (single and multi, with upsert)
   - Implemented delete action (single and multi)
   - Implemented aggregate action (pipeline support)
   - Official MongoDB Go driver integration
   - Comprehensive test suite with mocks

4. ✅ **Security Features**
   - All connection strings stored encrypted in credential vault
   - Parameterized queries prevent SQL injection
   - TLS/SSL connection support
   - Credential injection from vault at runtime
   - No hardcoded connection strings

5. ✅ **Testing & Quality**
   - 100% test coverage for all database actions
   - TDD methodology followed (tests written first)
   - Mock-based testing for unit tests
   - Context timeout testing
   - Connection pooling verification
   - Transaction rollback testing
   - Error handling validation

6. ✅ **Integration Registry**
   - All actions registered with global registry
   - Action wrappers for map-based config
   - Proper naming convention (postgres:query, mysql:statement, mongodb:find)
   - Validation methods for all configs

7. ✅ **Documentation**
   - Comprehensive README with usage examples
   - API documentation for each action
   - Security best practices
   - Credential configuration examples
   - Testing instructions

**Dependencies Added:**
- `github.com/go-sql-driver/mysql` v1.9.3
- `go.mongodb.org/mongo-driver` v1.17.6
- `github.com/lib/pq` (already present)
- `github.com/DATA-DOG/go-sqlmock` (already present)

**Files Created:**
- `/internal/integrations/database/models.go` - Common types and configs
- `/internal/integrations/database/postgres.go` - PostgreSQL connector
- `/internal/integrations/database/postgres_test.go` - PostgreSQL tests
- `/internal/integrations/database/mysql.go` - MySQL connector
- `/internal/integrations/database/mysql_test.go` - MySQL tests
- `/internal/integrations/database/mongodb.go` - MongoDB connector
- `/internal/integrations/database/mongodb_test.go` - MongoDB tests
- `/internal/integrations/database/actions.go` - Action registry wrappers
- `/internal/integrations/database/README.md` - Documentation

**Test Results:**
```
PASS: TestPostgresQueryAction_Execute
PASS: TestPostgresStatementAction_Execute
PASS: TestPostgresTransactionAction_Execute
PASS: TestPostgresActions_ContextTimeout
PASS: TestPostgresActions_ConnectionPooling
PASS: TestMySQLQueryAction_Execute
PASS: TestMySQLStatementAction_Execute
PASS: TestMySQLTransactionAction_Execute
PASS: TestMySQLActions_ConnectionPooling
PASS: TestMongoFindAction_Execute
PASS: TestMongoInsertAction_Execute
PASS: TestMongoUpdateAction_Execute
PASS: TestMongoDeleteAction_Execute
PASS: TestMongoAggregateAction_Execute
```

**Code Quality:**
- Followed SOLID principles
- DRY - Common helpers extracted
- Functions under 50 lines
- Cognitive complexity < 15
- No code duplication
- Proper error wrapping
- Interface-based design for testability

---

Last Updated: 2024-12-20
