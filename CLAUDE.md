# gorax Development Guidelines

## User Preferences
- **Always show next items to work on** at the end of each response when completing tasks

---

## Development Principles

### Test-Driven Development (TDD)
**MANDATORY**: All new code MUST follow TDD:
1. **Red**: Write a failing test first
2. **Green**: Write minimal code to pass the test
3. **Refactor**: Clean up while keeping tests green

### Clean Code & SOLID
- **Meaningful Names**: Variables, functions, classes reveal intent
- **Small Functions**: One thing well (< 20 lines preferred)
- **DRY**: Extract common logic into reusable functions
- **YAGNI**: Don't add functionality until needed
- **Single Responsibility**: One reason to change
- **Open/Closed**: Open for extension, closed for modification
- **Dependency Inversion**: Depend on abstractions, not concretions

### Code Quality Rules
| Metric | Limit |
|--------|-------|
| Cognitive Complexity | < 15 |
| Cyclomatic Complexity | < 10 |
| Function Length | < 50 lines |
| File Length | < 400 lines |
| Parameters | < 4 (use objects for more) |

## Cognitive Complexity Calculation

SonarQube rules (keep under 15):

**+1 for each**: `if`, `else if`, `else`, `switch`, `case`, `for`, `while`, `catch`, labeled `break`/`continue`, binary logical operator sequences, each nesting level

**Example** (complexity = 7 → refactored to 3):
```go
// BAD: complexity = 7
func process(items []Item) error {
    for _, item := range items {          // +1 (loop)
        if item.IsValid() {               // +2 (if + nesting)
            if item.NeedsProcessing() {   // +3 (if + nesting)
                if err := item.Process(); err != nil { // +1
                    return err
                }
            }
        }
    }
    return nil
}

// GOOD: complexity = 3 (extract to helper)
func process(items []Item) error {
    for _, item := range items {          // +1
        if err := processItem(item); err != nil { // +2
            return err
        }
    }
    return nil
}

func processItem(item Item) error {
    if !item.IsValid() || !item.NeedsProcessing() { // +1
        return nil
    }
    return item.Process()
}
```

## Project Structure

### Go
```
internal/           # Private application code
  domain/          # Business logic, no external dependencies
  service/         # Application services
  repository/      # Data access
  api/             # HTTP handlers
cmd/               # Application entry points
```

**Patterns**: Interfaces for dependencies, constructor injection, table-driven tests, error wrapping (`fmt.Errorf("op failed: %w", err)`)

### TypeScript/React
```
src/
  components/      # Reusable UI components
  pages/           # Route-level components
  hooks/           # Custom React hooks
  stores/          # Zustand state stores
  api/             # API client and hooks
  types/           # TypeScript type definitions
```

**Patterns**: Functional components, custom hooks, Zustand for global state, TanStack Query for server state

## Git Workflow

**MANDATORY**: Follow Git Flow:
1. **Never commit directly to `main` or `dev`**
2. Create branches off `dev`: `git checkout -b <ticket>-<short-description>`
3. Branch naming: `<ticket>-<description>` (e.g., `RFLOW-123-add-webhook-auth`)
4. Squash commits before merging
5. Create PR to merge back into `dev`

## Code Review Checklist
- [ ] Tests written FIRST (TDD)
- [ ] All tests pass
- [ ] No commented-out code or debug statements
- [ ] Functions small and focused
- [ ] No code duplication
- [ ] Error handling complete
- [ ] Types explicit (no `any` in TS)
- [ ] Dependencies injected

## Commands Available

| Command | Description |
|---------|-------------|
| `/review` | Review code for SOLID, DRY, complexity |
| `/test-plan` | Generate comprehensive test plan |
| `/refactor` | Suggest refactoring opportunities |
| `/tdd` | Start TDD workflow for a feature |
| `/architecture` | Review architectural design |
| `/gen-tests` | Generate test code for a file/function |

---

## Design Patterns Reference

### Creational
- **Factory**: Centralize complex object creation logic
- **Builder**: Construct complex objects with optional parameters (fluent API)

### Structural
- **Repository**: Abstract data access from business logic
- **Adapter**: Convert one interface to another
- **Decorator**: Add behavior dynamically (logging, retry, caching)

### Behavioral
- **Strategy**: Interchangeable algorithms (retry strategies, validators)
- **Observer**: Event-driven notifications (execution events)
- **Command**: Encapsulate requests as objects (undo/redo)

### React Patterns
- **Custom Hooks**: Extract reusable stateful logic
- **Compound Components**: Flexible, related components sharing state
- **Render Props**: Component composition via function props

---

## Gorax Expert Agents

Use the Task tool with appropriate `subagent_type` for domain-specific tasks.

### gorax-workflow-expert
**Use for**: Workflow execution, actions, triggers, expressions, node graph traversal

**Key files**: `internal/executor/`, `internal/workflow/`, `internal/executor/expression/`, `internal/executor/actions/`

**Domain**: Visitor pattern for traversal, Action interface (`Execute(ctx, input) (output, error)`), CEL expressions, JSON graph storage, node types (trigger, action, condition, loop, parallel)

### gorax-credential-expert
**Use for**: Encryption, injection, masking, security, audit logging

**Key files**: `internal/credential/encryption.go`, `internal/credential/injector.go`, `internal/credential/masker.go`

**Domain**: Envelope encryption (DEK/KEK), AES-256-GCM, `{{credentials.name}}` references, `[REDACTED]` masking, KMS for production

### gorax-worker-expert
**Use for**: Queue processing, SQS, job orchestration, concurrency, retries

**Key files**: `internal/worker/`, `internal/queue/`

**Domain**: SQS polling, tenant concurrency limits, DLQ for failures, visibility timeout, graceful shutdown

### gorax-frontend-expert
**Use for**: React canvas editor, workflow visualization, state management

**Key files**: `web/src/components/canvas/`, `web/src/pages/WorkflowEditor.tsx`, `web/src/stores/`, `web/src/hooks/`

**Domain**: ReactFlow canvas, custom nodes (Trigger, Action, Condition), Zustand for canvas state, TanStack Query for server state, WebSocket for real-time updates

### gorax-webhook-expert
**Use for**: Webhook handling, event filtering, replay, signature verification

**Key files**: `internal/webhook/`, `internal/api/handlers/webhook*.go`

**Domain**: HMAC-SHA256 signatures, JSONPath filters, event history, replay functionality

### gorax-ai-expert
**Use for**: AI/LLM integrations, chat completions, embeddings, entity extraction

**Key files**: `internal/llm/`, `internal/llm/providers/`, `internal/integrations/ai/`

**Domain**: Multi-provider (OpenAI, Anthropic, Bedrock), Provider interface, credential-based auth, token tracking, AI actions (ChatCompletion, Embedding, EntityExtraction)

### gorax-architecture-expert
**Use for**: System design, clean architecture, dependency injection, layering

**Domain**: Clean Architecture layers (API → Service → Domain → Repository), inward dependencies, interface segregation, event-driven architecture

### gorax-go-expert
**Use for**: Go best practices, error handling, concurrency, testing

**Domain**: Error wrapping, context propagation, errgroup for concurrency, table-driven tests, interface-based design

**Anti-patterns**: Ignoring errors, naked returns, global state, mutex without defer, goroutine leaks, interface pollution

### gorax-react-expert
**Use for**: React components, hooks, state management, performance

**Domain**: Functional components, custom hooks, Zustand/TanStack Query, compound components, error boundaries

**Performance**: useMemo, useCallback, React.memo

### gorax-security-expert
**Use for**: Security audits, auth/authz, encryption, OWASP compliance

**Key files**: `internal/api/middleware/auth.go`, `internal/api/middleware/security_headers.go`, `internal/rbac/`, `internal/credential/encryption.go`

**Security layers**: Headers → CORS → Rate Limit → Auth → Tenant → RBAC → Validation → Encryption

**OWASP mitigations**: RBAC (access control), AES-256-GCM (crypto), parameterized queries (injection), security headers (misconfiguration), Kratos (auth), signature verification (integrity)

### gorax-qa-expert
**Use for**: Test strategy, coverage analysis, integration/E2E testing

**Key files**: `internal/**/*_test.go`, `web/src/**/*.test.tsx`

**Domain**: Go (testify/assert, testify/mock, table-driven), React (Vitest, Testing Library, MSW), AAA pattern, 80%+ coverage on business logic

### gorax-reactflow-expert
**Use for**: ReactFlow canvas, node-based editors, DAG validation

**Key files**: `web/src/components/canvas/`, `web/src/components/nodes/`, `web/src/utils/dagValidation.ts`

**Domain**: ReactFlow v12, custom nodes, cycle detection, edge validation, drag-and-drop

### gorax-go-security-expert
**Use for**: Go-specific security, SQL injection prevention, goroutine safety

**Key patterns**:
- Parameterized queries: `$1, $2` placeholders (never string concat)
- Context with timeout for all external calls
- `errgroup.Group` with `SetLimit()` for bounded concurrency
- Path validation against traversal (`..`)
- `crypto/rand` for tokens (not `math/rand`)
- `exec.Command` with args array (not shell)

### gorax-react-security-expert
**Use for**: React security, XSS prevention, CSRF, secure state

**Key patterns**:
- Never `dangerouslySetInnerHTML` without DOMPurify
- Auth tokens in httpOnly cookies (not localStorage)
- CSRF token in POST/PUT/DELETE
- Server-side authorization (client-side is bypass-able)
- Validate redirect URLs (whitelist)

### gorax-go-testing-expert
**Use for**: Go tests, table-driven tests, mocking, benchmarks

**Commands**: `go test -cover ./...`, `go test -race ./...`, `go test -bench=. -benchmem ./...`

### gorax-react-testing-expert
**Use for**: React tests, component/hook testing, MSW mocking

**Patterns**: `render()`, `screen.getByRole()`, `userEvent.setup()`, `waitFor()`, `renderHook()`, MSW for API mocking
