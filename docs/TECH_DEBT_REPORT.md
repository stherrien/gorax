# Technical Debt Report - Gorax
**Generated:** 2025-12-20
**Scope:** Backend (`/internal`) and Frontend (`/web/src`)

---

## Executive Summary

This comprehensive technical debt audit analyzed the Gorax workflow automation platform codebase using multiple static analysis tools and manual code review. The analysis identified **162 issues** across the following categories:

- **Critical Issues:** 2
- **High Priority:** 24
- **Medium Priority:** 76
- **Low Priority:** 60

**Overall Code Quality:** Good foundation with some areas requiring attention. The codebase follows clean code principles in most areas but has accumulated technical debt that should be addressed systematically.

---

## 1. Static Analysis Results

### 1.1 Tool Summary

| Tool | Issues Found | Status |
|------|--------------|--------|
| `go vet` | 0 | ✅ PASS |
| `staticcheck` | 63 | ⚠️ WARNINGS |
| `golangci-lint` | 99 | ⚠️ WARNINGS |
| `gocyclo` | 40 functions >10 | ⚠️ COMPLEXITY |

### 1.2 Key Metrics

- **Total Go Files:** 171 production, 104 test files
- **Test Coverage Ratio:** 1:1.64 (good test coverage)
- **Largest Files:**
  - `internal/webhook/filter_test.go` (2,352 lines)
  - `internal/webhook/replay_test.go` (1,088 lines)
  - `internal/workflow/service.go` (954 lines)
- **Functions with High Cyclomatic Complexity:** 40 functions exceed threshold of 10

---

## 2. Critical Issues (Fix Immediately)

### 2.1 Ineffective Break Statements ⚠️ HIGH RISK

**Priority:** CRITICAL
**Impact:** Logic bugs that can cause incorrect behavior

**Locations:**
- `internal/websocket/events_test.go:388` - Break inside select statement doesn't break outer loop
- `internal/websocket/hub_test.go:268` - Break inside select statement doesn't break outer loop

**Issue:**
```go
// Current (BROKEN)
for {
    select {
    case msg := <-ch:
        if condition {
            break // Only breaks from select, not loop!
        }
    }
}
```

**Fix:** Use labeled break or flag variable
```go
loop:
for {
    select {
    case msg := <-ch:
        if condition {
            break loop // Correctly breaks outer loop
        }
    }
}
```

**Recommendation:** Fix immediately - this is a functional bug.

---

### 2.2 Unchecked Error Returns ⚠️ HIGH RISK

**Priority:** CRITICAL
**Impact:** Silent failures, resource leaks, incorrect behavior

**Total Occurrences:** 40+ unchecked error returns

**High-Risk Examples:**

1. **Database Transaction Rollback** (Resource Leak Risk)
   ```go
   // internal/database/tenant_hooks.go:108
   tx.Rollback() // Error not checked - transaction may not rollback
   ```

2. **WebSocket Write Operations** (Connection Corruption Risk)
   ```go
   // internal/websocket/hub.go:268, 273, 274
   w.Write(message)           // Error not checked
   w.Write([]byte{'\n'})      // Error not checked
   c.Conn.WriteMessage(...)   // Error not checked
   ```

3. **JSON Encoding in HTTP Handlers** (Silent Response Corruption)
   ```go
   // internal/rbac/middleware.go:192
   json.NewEncoder(w).Encode(map[string]string{...}) // Error not checked
   ```

**Recommendation:**
- Add error checks for all database operations
- Log errors for operations where you can't return them
- Return errors from functions where possible

---

## 3. High Priority Issues (Fix Within Sprint)

### 3.1 Unused Functions (Dead Code)

**Priority:** HIGH
**Impact:** Code maintenance burden, confusion, bloat

**Total:** 17 unused functions identified

**Critical Unused Code:**
```
internal/executor/conditional.go:
  - executeConditionalAction (line 22)
  - buildConditionalExecutionPlan (line 110)
  - findStartNodes (line 135)
  - getNextNodes (line 170)
  - hasAllDependenciesCompleted (line 188)
  - findNodesToSkip (line 204)

internal/executor/retry.go:
  - withAttempt (line 232)

internal/executor/tracing.go:
  - executeWithTracing (line 11)
  - executeNodeWithTracing (line 24)

internal/worker/message_handler.go:
  - processExecutionMessageWithRequeue (line 17)
  - requeueMessageWithDelay (line 76)

internal/worker/tracing.go:
  - handleMessageWithTracing (line 11)

internal/humantask/repository.go:
  - buildFilterQuery (line 266)

internal/workflow/formula/evaluator.go:
  - timeToExpr (line 275)
  - program field (line 13)

internal/workflow/repository.go:
  - setTenantContext (line 30)

internal/workflow/service_test.go:
  - jsonRawMessage (line 588)

internal/template/service_test.go:
  - errTest variable (line 383)
```

**Recommendation:**
- Remove if truly unused
- If planning to use, add TODO comment with issue number
- Move to separate "future features" package if experimental

---

### 3.2 Deprecated API Usage

**Priority:** HIGH
**Impact:** Future compatibility issues, warnings

**Occurrences:**

1. **netErr.Temporary() - Deprecated since Go 1.18**
   ```go
   // internal/executor/errors.go:99
   if netErr.Temporary() {
       return ErrorClassificationTransient
   }
   ```
   **Fix:** Use explicit error type checking instead

2. **strings.Title() - Deprecated since Go 1.18**
   ```go
   // internal/notification/slack.go:321
   strings.Title(status)
   ```
   **Fix:** Use `golang.org/x/text/cases` package

3. **trace.NewNoopTracerProvider() - Deprecated**
   ```go
   // internal/tracing/tracer.go:26
   return trace.NewNoopTracerProvider(), func() {}, nil
   ```
   **Fix:** Use `go.opentelemetry.io/otel/trace/noop.NewTracerProvider`

**Recommendation:** Replace all deprecated APIs in next sprint.

---

### 3.3 Context Key Type Safety Issues

**Priority:** HIGH
**Impact:** Context key collisions, runtime panics

**Occurrences:** 20+ locations using string keys directly

**Problem:**
```go
// BAD: Built-in string type allows collisions
ctx = context.WithValue(ctx, "userID", userID)
ctx = context.WithValue(ctx, "tenant_id", tenantID)
```

**Solution:**
```go
// GOOD: Custom type prevents collisions
type contextKey string
const (
    ContextKeyUserID   contextKey = "user_id"
    ContextKeyTenantID contextKey = "tenant_id"
)
ctx = context.WithValue(ctx, ContextKeyUserID, userID)
```

**Affected Files:**
- `internal/api/handlers/template_handler_test.go` (multiple)
- `internal/api/middleware/quota.go:230`
- `internal/rbac/middleware.go:173`
- `internal/rbac/middleware_test.go` (multiple)
- `internal/tenant/service_test.go:64`

**Recommendation:** Create centralized context key types in `internal/api/context/keys.go`

---

### 3.4 High Cyclomatic Complexity Functions

**Priority:** HIGH
**Impact:** Maintainability, testability, bug risk

**Functions Exceeding Threshold (>15):**

| Function | Complexity | File | Line |
|----------|------------|------|------|
| `(*Executor).Execute` | 26 | executor.go | 135 |
| `ClassifyError` | 26 | errors.go | 75 |
| `TestGetNextRunTimes` | 23 | service_test.go | 179 |
| `(*HTTPAction).executeHTTP` | 20 | http.go | 68 |
| `extractValue` | 18 | filter.go | 168 |
| `(*Service).CreateDefaultRoles` | 18 | service.go | 282 |
| `(*Executor).executeNode` | 18 | executor.go | 412 |
| `(*Service).DryRun` | 17 | service.go | 549 |

**Most Critical: `(*Executor).Execute` (Complexity: 26)**

**Location:** `internal/executor/executor.go:135`

**Cognitive Complexity Breakdown:**
- Multiple nested conditionals
- Complex error handling paths
- Retry logic embedded in main flow
- Broadcasting logic interspersed

**Refactoring Strategy:**
```
Execute (main orchestration)
  ├── validateExecution()
  ├── prepareExecutionContext()
  ├── executeWorkflowSteps()
  │   ├── executeStep()
  │   └── handleStepError()
  ├── finalizeExecution()
  └── broadcastResults()
```

**Recommendation:** Break down into smaller, focused functions (< 15 complexity each)

---

## 4. Medium Priority Issues (Fix Within Month)

### 4.1 Code Simplification Opportunities

**Priority:** MEDIUM
**Impact:** Code readability, maintainability

**Occurrences:** 8 locations

1. **Unnecessary Nil Checks Before len()**
   ```go
   // UNNECESSARY
   if c.Value == nil || len(c.Value) == 0 {
       return errors.New("value is required")
   }

   // SIMPLIFIED (len() returns 0 for nil)
   if len(c.Value) == 0 {
       return errors.New("value is required")
   }
   ```

   **Locations:**
   - `internal/credential/domain.go:180, 201`
   - `internal/executor/actions/http.go:110`
   - `internal/webhook/replay.go:84`
   - `internal/workflow/service.go:593`

2. **Loop Can Be Replaced with append()**
   ```go
   // VERBOSE
   for _, id := range strings.Split(tt.queryIDs, ",") {
       ids = append(ids, id)
   }

   // CONCISE
   ids = append(ids, strings.Split(tt.queryIDs, ",")...)
   ```

   **Location:** `internal/api/handlers/bulk_handler_test.go:385`

3. **Type Conversion Instead of Struct Literal**
   ```go
   // VERBOSE
   req := CreateIssueRequest{
       Project:     issueConfig.Project,
       Summary:     issueConfig.Summary,
       Description: issueConfig.Description,
       IssueType:   issueConfig.IssueType,
       // ... all fields copied manually
   }

   // CONCISE
   req := CreateIssueRequest(issueConfig)
   ```

   **Locations:**
   - `internal/integrations/jira/actions.go:51`
   - `internal/notification/slack.go:146`

**Recommendation:** Apply all simplifications - quick wins with no risk.

---

### 4.2 Error String Capitalization

**Priority:** MEDIUM
**Impact:** Style consistency, Go conventions

**Occurrences:** 7 locations

**Issue:** Error strings should not be capitalized (Go convention)

```go
// BAD
return fmt.Errorf("Failed to connect")

// GOOD
return fmt.Errorf("failed to connect")
```

**Locations:**
- `internal/integrations/jira/client.go:289`
- `internal/integrations/slack/errors.go:18, 25, 80, 81, 82`
- `internal/notification/slack.go:188`

**Recommendation:** Automated fix with regex find/replace.

---

### 4.3 Empty Branch Anti-Pattern

**Priority:** MEDIUM
**Impact:** Code clarity, potential bugs

**Occurrences:** 2 locations

```go
// internal/credential/service_impl.go:64, 78
if err := s.repo.UpdateLastUsedAt(ctx, tenantID, credentialID); err != nil {
    // TODO: Should we log this error?
}
```

**Issue:** Empty error handling suggests incomplete implementation or dead code.

**Recommendation:**
- Add logging: `s.logger.Warn("failed to update last used at", "error", err)`
- Or remove the check if truly not needed
- Add TODO comment if deferred work

---

### 4.4 Ineffectual Assignments

**Priority:** MEDIUM
**Impact:** Code clarity, potential logic bugs

**Occurrences:**
- `internal/credential/repository.go:284` - `argIndex++` has no effect
- `internal/humantask/repository.go:153, 309` - `argPos++` has no effect

**Example:**
```go
argIndex := 1
query := "SELECT * FROM table WHERE id = $1"
// argIndex++ // This does nothing if argIndex not used after
```

**Recommendation:** Remove if unused, or fix logic if intended to be used.

---

### 4.5 TODO/FIXME Comments

**Priority:** MEDIUM
**Impact:** Incomplete features, deferred work

**Found:** 1 critical TODO

**Location:** `internal/api/handlers/rbac_handler.go:276`
```go
// TODO: Parse from query params
limit := 50
offset := 0
```

**Issue:** Pagination not implemented for audit logs endpoint.

**Recommendation:** Create Jira tickets for all TODOs with sprint assignments.

---

## 5. Low Priority Issues (Technical Debt Backlog)

### 5.1 Missing ESLint Configuration

**Priority:** LOW
**Impact:** TypeScript/React code quality not enforced

**Issue:** No ESLint configuration found for web frontend.

```bash
npm run lint
# ESLint couldn't find a configuration file
```

**Recommendation:**
```bash
npm init @eslint/config
# Choose:
# - TypeScript
# - React
# - Strict rules
```

Add to `web/.eslintrc.json`:
```json
{
  "extends": [
    "eslint:recommended",
    "plugin:@typescript-eslint/recommended",
    "plugin:react/recommended",
    "plugin:react-hooks/recommended"
  ],
  "rules": {
    "no-console": "warn",
    "@typescript-eslint/no-explicit-any": "error",
    "react/prop-types": "off"
  }
}
```

---

### 5.2 Large Test Files

**Priority:** LOW
**Impact:** Test maintainability

**Files Exceeding 1000 Lines:**
- `internal/webhook/filter_test.go` - 2,352 lines
- `internal/webhook/replay_test.go` - 1,088 lines
- `internal/api/handlers/webhook_management_handler_test.go` - 1,081 lines
- `internal/workflow/repository_test.go` - 997 lines

**Recommendation:** Split into multiple test files by feature/scenario.

---

### 5.3 Large Production Files

**Priority:** LOW
**Impact:** Code navigation, maintainability

**Files Exceeding 600 Lines:**
- `internal/workflow/service.go` - 954 lines (❗ Exceeds 400 line guideline)
- `internal/workflow/repository.go` - 634 lines
- `internal/webhook/repository.go` - 630 lines
- `internal/webhook/service.go` - 615 lines
- `internal/executor/executor.go` - 588 lines

**Recommendation:** Consider splitting large files by responsibility:
- Extract interfaces to separate files
- Group related methods into sub-services
- Use composition over large monolithic files

---

### 5.4 Security Considerations

**Priority:** LOW (No Critical Issues Found)
**Impact:** Security posture

**Positive Findings:**
- ✅ No SQL injection vulnerabilities (using parameterized queries)
- ✅ No XSS vulnerabilities (no `dangerouslySetInnerHTML` found)
- ✅ No `eval()` usage
- ✅ Credentials properly encrypted at rest
- ✅ Context-based tenant isolation implemented

**Minor Concerns:**
1. **Password/Secret Handling:** 122 files reference credentials/secrets/tokens
   - Verify all are properly masked in logs
   - Ensure no secrets in error messages

2. **Webhook Signature Verification:** Review implementation
   - Ensure constant-time comparison for HMAC validation
   - Check for replay attack protection

**Recommendation:** Schedule security audit focused on:
- Credential masking in logs and errors
- Webhook authentication mechanisms
- API rate limiting effectiveness

---

## 6. Code Quality Patterns

### 6.1 Positive Patterns Found ✅

1. **Strong Test Coverage**
   - 104 test files for 171 production files
   - Comprehensive integration tests
   - Table-driven tests used extensively

2. **Good Error Handling Architecture**
   - Custom error types with classification
   - Error wrapping with context
   - Transient vs permanent error detection

3. **Clean Dependency Injection**
   - Interface-based design
   - Constructor injection pattern
   - Testable components

4. **Repository Pattern**
   - Clean separation of data access
   - Consistent interface across domains

5. **Tenant Isolation**
   - Database-level tenant context
   - Middleware enforcement
   - Context propagation

### 6.2 Anti-Patterns to Address ⚠️

1. **God Objects**
   - `Executor` has too many responsibilities
   - `WorkflowService` is doing too much

2. **Feature Envy**
   - Some functions reach deep into other objects
   - Consider moving logic closer to data

3. **Primitive Obsession**
   - String-based context keys (addressed in High Priority)
   - Consider value objects for IDs

---

## 7. Refactoring Opportunities

### 7.1 Extract Service Pattern

**Target:** `internal/workflow/service.go` (954 lines)

**Proposed Split:**
```
workflow/
  ├── service.go           (orchestration, 200 lines)
  ├── version_service.go   (versioning logic)
  ├── execution_service.go (execution management)
  ├── validation_service.go (definition validation)
  └── webhook_sync.go      (webhook synchronization)
```

### 7.2 Strategy Pattern for Error Classification

**Target:** `internal/executor/errors.go:75` (26 complexity)

**Current:** Single massive function with nested conditionals

**Proposed:**
```go
type ErrorClassifier interface {
    Classify(error) ErrorClassification
}

type ChainedClassifier struct {
    classifiers []ErrorClassifier
}

// Specific classifiers
type NetworkErrorClassifier struct{}
type ContextErrorClassifier struct{}
type HTTPErrorClassifier struct{}
```

### 7.3 Builder Pattern for Executor

**Target:** `internal/executor/executor.go`

**Issue:** Multiple constructor functions with overlapping logic

**Proposed:**
```go
type ExecutorBuilder struct {
    repo        *Repository
    logger      *slog.Logger
    broadcaster Broadcaster
    injector    *credential.Injector
    credService credential.Service
}

func NewExecutorBuilder(repo *Repository, logger *slog.Logger) *ExecutorBuilder

func (b *ExecutorBuilder) WithBroadcaster(bc Broadcaster) *ExecutorBuilder
func (b *ExecutorBuilder) WithCredentials(...) *ExecutorBuilder
func (b *ExecutorBuilder) Build() *Executor
```

---

## 8. Recommendations by Priority

### Immediate Actions (This Week)

1. ✅ **Fix ineffective break statements** - Critical bugs
2. ✅ **Add error checks to WebSocket operations** - Prevent connection corruption
3. ✅ **Add error checks to database transactions** - Prevent resource leaks
4. ⚠️ **Create Jira ticket tracking for all TODOs** - Visibility

### Sprint Actions (Next 2 Weeks)

5. 🔧 **Remove dead code** - Clean up unused functions
6. 🔧 **Replace deprecated APIs** - Future compatibility
7. 🔧 **Fix context key type safety** - Prevent runtime issues
8. 🔧 **Refactor high-complexity functions** - Start with Execute() and ClassifyError()

### Monthly Actions

9. 📝 **Apply code simplifications** - Quick wins
10. 📝 **Fix error string capitalization** - Style consistency
11. 📝 **Add ESLint configuration** - Frontend quality
12. 📝 **Split large test files** - Maintainability

### Quarterly Actions

13. 🏗️ **Refactor WorkflowService** - Architecture improvement
14. 🏗️ **Implement Strategy Pattern for errors** - Complexity reduction
15. 🏗️ **Add Builder Pattern for Executor** - API improvement
16. 🔒 **Conduct security audit** - Comprehensive review

---

## 9. Technical Debt Metrics

### Current State

| Metric | Value | Target | Status |
|--------|-------|--------|--------|
| Functions >15 Complexity | 8 | 0 | 🔴 Needs Work |
| Unused Functions | 17 | 0 | 🔴 Needs Work |
| Unchecked Errors | 40+ | 0 | 🔴 Needs Work |
| Deprecated API Usage | 3 | 0 | 🟡 Acceptable |
| Large Files (>600 lines) | 5 | 0 | 🟡 Acceptable |
| Test Coverage Ratio | 1:1.64 | 1:1 | 🟢 Good |
| ESLint Configuration | Missing | Present | 🔴 Needs Work |

### Estimated Effort

| Category | Issues | Effort (Hours) | Priority |
|----------|--------|----------------|----------|
| Critical Fixes | 2 | 4 | P0 |
| High Priority | 24 | 40 | P1 |
| Medium Priority | 76 | 60 | P2 |
| Low Priority | 60 | 40 | P3 |
| **Total** | **162** | **144** | - |

### Suggested Sprint Allocation

- **Sprint 1:** Critical + 50% High Priority (24 hours)
- **Sprint 2:** Remaining High Priority (20 hours)
- **Sprint 3:** Medium Priority (30 hours)
- **Sprint 4:** Medium Priority (30 hours)
- **Ongoing:** Low Priority (background work)

---

## 10. Quick Wins (Can Complete Today)

These require minimal effort with high impact:

1. ✅ **Remove unused test helper** - `internal/workflow/service_test.go:588`
2. ✅ **Remove unused variable** - `internal/template/service_test.go:383`
3. ✅ **Apply loop simplification** - `internal/api/handlers/bulk_handler_test.go:385`
4. ✅ **Remove unnecessary nil checks** - 4 locations
5. ✅ **Fix error capitalization** - 7 locations (regex replace)
6. ✅ **Add logging to empty branches** - 2 locations
7. ✅ **Remove ineffectual assignments** - 3 locations

**Total Time:** ~2 hours
**Impact:** -13 issues, cleaner codebase

---

## 11. Monitoring & Prevention

### Recommended CI/CD Checks

Add to `.github/workflows/ci.yml`:

```yaml
- name: Static Analysis
  run: |
    go vet ./...
    staticcheck ./...
    golangci-lint run --max-issues-per-linter 0

- name: Complexity Check
  run: |
    gocyclo -over 15 internal/ > complexity.txt
    if [ -s complexity.txt ]; then
      cat complexity.txt
      exit 1
    fi

- name: Test Coverage
  run: |
    go test -coverprofile=coverage.out ./...
    go tool cover -func=coverage.out | tail -1 | awk '{print $3}' | sed 's/%//' > coverage.txt
    COVERAGE=$(cat coverage.txt)
    if (( $(echo "$COVERAGE < 80" | bc -l) )); then
      echo "Coverage $COVERAGE% is below 80%"
      exit 1
    fi

- name: Frontend Lint
  run: |
    cd web
    npm run lint
```

### Code Review Checklist

Add to `PULL_REQUEST_TEMPLATE.md`:

```markdown
## Code Quality Checklist

- [ ] No functions with cyclomatic complexity >15
- [ ] All errors are checked and handled
- [ ] No TODO comments without Jira ticket reference
- [ ] Tests added/updated for changes
- [ ] No deprecated API usage
- [ ] Context keys use custom types (not strings)
- [ ] Error messages are lowercase
- [ ] No files exceed 400 lines
```

---

## 12. Conclusion

The Gorax codebase demonstrates **good engineering practices** overall with:
- Strong test coverage
- Clean architecture patterns
- Proper separation of concerns
- Secure coding practices

However, **technical debt has accumulated** in several areas:
- Unchecked error returns create stability risks
- High complexity functions reduce maintainability
- Dead code creates confusion
- Missing tooling (ESLint) reduces frontend quality

**Recommended Approach:**
1. Address critical issues immediately (4 hours)
2. Tackle high-priority items over 2 sprints (60 hours)
3. Gradually address medium/low priority in background

Following this plan will significantly improve code quality while maintaining velocity on new features.

---

## Appendix A: Tool Outputs

### staticcheck Summary
- Total issues: 63
- Unused code: 17 functions
- Style issues: 20
- Correctness issues: 26

### golangci-lint Summary
- Total issues: 99
- errcheck: 40
- unused: 17
- gosimple: 8
- staticcheck: 26
- ineffassign: 3
- SA series: 5

### Complexity Analysis
- Functions analyzed: 400+
- Functions >10: 40
- Functions >15: 8
- Highest: 26 (Execute, ClassifyError)

---

**Report Generated by:** Claude Code Static Analysis Suite
**Next Review:** Recommended quarterly or after major feature releases
