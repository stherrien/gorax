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

Before implementing any feature or fix:
- Write unit tests that define expected behavior
- Tests must fail initially (proving they test something)
- Only then write implementation code

### Clean Code Principles

- **Meaningful Names**: Variables, functions, and classes should reveal intent
- **Small Functions**: Each function should do ONE thing well (< 20 lines preferred)
- **No Comments for Bad Code**: Rewrite unclear code instead of commenting it
- **DRY (Don't Repeat Yourself)**: Extract common logic into reusable functions
- **YAGNI**: Don't add functionality until it's needed

### SOLID Principles

1. **Single Responsibility**: A class/module should have only one reason to change
2. **Open/Closed**: Open for extension, closed for modification
3. **Liskov Substitution**: Subtypes must be substitutable for their base types
4. **Interface Segregation**: Many specific interfaces > one general interface
5. **Dependency Inversion**: Depend on abstractions, not concretions

### Code Quality Rules

- **Cognitive Complexity**: Keep function complexity under 15 (SonarQube metric)
- **Cyclomatic Complexity**: Max 10 per function
- **Max Function Length**: 50 lines (excluding tests)
- **Max File Length**: 400 lines (split if larger)
- **Max Parameters**: 4 per function (use objects for more)

## Go-Specific Guidelines

### Project Structure
```
internal/         # Private application code
  domain/        # Business logic, no external dependencies
  service/       # Application services
  repository/    # Data access
  api/           # HTTP handlers
cmd/             # Application entry points
pkg/             # Public libraries (if any)
```

### Patterns
- Use interfaces for dependencies (enables testing)
- Constructor injection for dependencies
- Table-driven tests
- Error wrapping with context: `fmt.Errorf("operation failed: %w", err)`

### Testing
- Test files: `*_test.go` in same package
- Use `testify/assert` for assertions
- Mock external dependencies
- Aim for 80%+ coverage on business logic

## TypeScript/React Guidelines

### Project Structure
```
src/
  components/    # Reusable UI components
  pages/         # Route-level components
  hooks/         # Custom React hooks
  stores/        # Zustand state stores
  api/           # API client and hooks
  utils/         # Pure utility functions
  types/         # TypeScript type definitions
```

### Patterns
- Functional components only (no class components)
- Custom hooks for reusable logic
- Zustand for global state (keep stores small and focused)
- React Query for server state (caching, refetching)

### Testing
- Test files: `*.test.ts` or `*.test.tsx`
- Use Vitest + React Testing Library
- Test behavior, not implementation
- Mock API calls

## Git Workflow

**MANDATORY**: Follow Git Flow process:

1. **Never commit directly to `main` or `dev` branches** - Always create a feature branch first
2. **Create branches off `dev`** before making any commits:
   ```bash
   git checkout dev
   git pull origin dev
   git checkout -b <ticket>-<short-description>
   ```
3. **Branch naming**: `<ticket>-<short-description>` (e.g., `RFLOW-123-add-webhook-auth`)
4. Write meaningful commit messages
5. Squash commits before merging
6. Create PR to merge back into `dev`

**Before committing, always verify:**
```bash
git branch  # Ensure you're NOT on main or dev
```

## Code Review Checklist

Before submitting code, verify:

- [ ] Tests written FIRST (TDD)
- [ ] All tests pass
- [ ] No commented-out code
- [ ] No console.log/fmt.Println debugging statements
- [ ] Functions are small and focused
- [ ] No code duplication
- [ ] Error handling is complete
- [ ] Types are explicit (no `any` in TypeScript)
- [ ] Dependencies are injected (not hard-coded)

## Cognitive Complexity Calculation

SonarQube cognitive complexity rules (avoid exceeding 15):

**+1 for each**:
- `if`, `else if`, `else`
- `switch`, `case`
- `for`, `while`, `do while`
- `catch`
- `break`, `continue` to a label
- Sequences of binary logical operators
- Each level of nesting

**Example** (complexity = 7):
```go
func process(items []Item) error {        // +0
    for _, item := range items {          // +1 (loop)
        if item.IsValid() {               // +2 (if + nesting)
            if item.NeedsProcessing() {   // +3 (if + nesting)
                err := item.Process()
                if err != nil {           // +1 (if, same nesting)
                    return err
                }
            }
        }
    }
    return nil
}
```

**Refactored** (complexity = 3):
```go
func process(items []Item) error {
    for _, item := range items {          // +1
        if err := processItem(item); err != nil {  // +2
            return err
        }
    }
    return nil
}

func processItem(item Item) error {
    if !item.IsValid() || !item.NeedsProcessing() {  // +1
        return nil
    }
    return item.Process()
}
```

## Design Patterns

### Creational Patterns

#### Factory Pattern
Use when object creation logic is complex or needs to be centralized.

```go
// Go Example
type ActionFactory struct{}

func (f *ActionFactory) Create(actionType string) (Action, error) {
    switch actionType {
    case "http":
        return NewHTTPAction(), nil
    case "transform":
        return NewTransformAction(), nil
    default:
        return nil, fmt.Errorf("unknown action type: %s", actionType)
    }
}
```

```typescript
// TypeScript Example
class NodeFactory {
  static create(type: NodeType): WorkflowNode {
    switch (type) {
      case 'trigger': return new TriggerNode();
      case 'action': return new ActionNode();
      default: throw new Error(`Unknown node type: ${type}`);
    }
  }
}
```

#### Builder Pattern
Use when constructing complex objects with many optional parameters.

```go
type WorkflowBuilder struct {
    workflow *Workflow
}

func NewWorkflowBuilder() *WorkflowBuilder {
    return &WorkflowBuilder{workflow: &Workflow{}}
}

func (b *WorkflowBuilder) WithName(name string) *WorkflowBuilder {
    b.workflow.Name = name
    return b
}

func (b *WorkflowBuilder) WithTrigger(trigger Trigger) *WorkflowBuilder {
    b.workflow.Trigger = trigger
    return b
}

func (b *WorkflowBuilder) Build() (*Workflow, error) {
    if err := b.workflow.Validate(); err != nil {
        return nil, err
    }
    return b.workflow, nil
}
```

### Structural Patterns

#### Repository Pattern
Abstracts data access logic from business logic.

```go
type WorkflowRepository interface {
    Create(ctx context.Context, workflow *Workflow) error
    GetByID(ctx context.Context, id string) (*Workflow, error)
    Update(ctx context.Context, workflow *Workflow) error
    Delete(ctx context.Context, id string) error
    List(ctx context.Context, filter WorkflowFilter) ([]*Workflow, error)
}

type PostgresWorkflowRepository struct {
    db *sqlx.DB
}

func (r *PostgresWorkflowRepository) GetByID(ctx context.Context, id string) (*Workflow, error) {
    var workflow Workflow
    err := r.db.GetContext(ctx, &workflow, "SELECT * FROM workflows WHERE id = $1", id)
    if err != nil {
        return nil, fmt.Errorf("get workflow: %w", err)
    }
    return &workflow, nil
}
```

#### Adapter Pattern
Converts one interface to another expected by clients.

```go
// External API response
type ExternalWebhookPayload struct {
    EventType string `json:"event_type"`
    Data      any    `json:"data"`
}

// Internal domain model
type TriggerEvent struct {
    Type    string
    Payload map[string]any
}

// Adapter
type WebhookAdapter struct{}

func (a *WebhookAdapter) ToTriggerEvent(payload ExternalWebhookPayload) TriggerEvent {
    return TriggerEvent{
        Type:    payload.EventType,
        Payload: payload.Data.(map[string]any),
    }
}
```

#### Decorator Pattern
Add behavior to objects dynamically.

```go
type Action interface {
    Execute(ctx context.Context, input map[string]any) (map[string]any, error)
}

// Base action
type HTTPAction struct{}

func (a *HTTPAction) Execute(ctx context.Context, input map[string]any) (map[string]any, error) {
    // Make HTTP request
    return result, nil
}

// Decorator: adds logging
type LoggingAction struct {
    wrapped Action
    logger  *slog.Logger
}

func (a *LoggingAction) Execute(ctx context.Context, input map[string]any) (map[string]any, error) {
    a.logger.Info("executing action", "input", input)
    result, err := a.wrapped.Execute(ctx, input)
    a.logger.Info("action completed", "error", err)
    return result, err
}

// Decorator: adds retry
type RetryAction struct {
    wrapped    Action
    maxRetries int
}

func (a *RetryAction) Execute(ctx context.Context, input map[string]any) (map[string]any, error) {
    var lastErr error
    for i := 0; i < a.maxRetries; i++ {
        result, err := a.wrapped.Execute(ctx, input)
        if err == nil {
            return result, nil
        }
        lastErr = err
        time.Sleep(time.Second * time.Duration(i+1))
    }
    return nil, fmt.Errorf("max retries exceeded: %w", lastErr)
}
```

### Behavioral Patterns

#### Strategy Pattern
Define a family of algorithms and make them interchangeable.

```go
type RetryStrategy interface {
    NextDelay(attempt int) time.Duration
    ShouldRetry(err error, attempt int) bool
}

type ExponentialBackoff struct {
    BaseDelay   time.Duration
    MaxDelay    time.Duration
    MaxAttempts int
}

func (s *ExponentialBackoff) NextDelay(attempt int) time.Duration {
    delay := s.BaseDelay * time.Duration(1<<attempt)
    if delay > s.MaxDelay {
        return s.MaxDelay
    }
    return delay
}

func (s *ExponentialBackoff) ShouldRetry(err error, attempt int) bool {
    return attempt < s.MaxAttempts && isRetryable(err)
}

type LinearBackoff struct {
    Delay       time.Duration
    MaxAttempts int
}

func (s *LinearBackoff) NextDelay(attempt int) time.Duration {
    return s.Delay
}
```

#### Observer Pattern (Event-Driven)
Notify multiple objects about state changes.

```go
type ExecutionEvent struct {
    Type        string
    ExecutionID string
    StepID      string
    Status      string
    Timestamp   time.Time
}

type ExecutionObserver interface {
    OnEvent(event ExecutionEvent)
}

type ExecutionNotifier struct {
    observers []ExecutionObserver
    mu        sync.RWMutex
}

func (n *ExecutionNotifier) Subscribe(observer ExecutionObserver) {
    n.mu.Lock()
    defer n.mu.Unlock()
    n.observers = append(n.observers, observer)
}

func (n *ExecutionNotifier) Notify(event ExecutionEvent) {
    n.mu.RLock()
    defer n.mu.RUnlock()
    for _, observer := range n.observers {
        go observer.OnEvent(event)
    }
}
```

#### Command Pattern
Encapsulate requests as objects.

```go
type Command interface {
    Execute(ctx context.Context) error
    Undo(ctx context.Context) error
}

type CreateWorkflowCommand struct {
    repo     WorkflowRepository
    workflow *Workflow
}

func (c *CreateWorkflowCommand) Execute(ctx context.Context) error {
    return c.repo.Create(ctx, c.workflow)
}

func (c *CreateWorkflowCommand) Undo(ctx context.Context) error {
    return c.repo.Delete(ctx, c.workflow.ID)
}
```

### React/TypeScript Patterns

#### Custom Hook Pattern
Extract reusable stateful logic.

```typescript
function useWorkflow(id: string) {
  const [workflow, setWorkflow] = useState<Workflow | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<Error | null>(null);

  useEffect(() => {
    const fetchWorkflow = async () => {
      try {
        setLoading(true);
        const data = await api.workflows.get(id);
        setWorkflow(data);
      } catch (err) {
        setError(err as Error);
      } finally {
        setLoading(false);
      }
    };
    fetchWorkflow();
  }, [id]);

  const save = async (updates: Partial<Workflow>) => {
    const updated = await api.workflows.update(id, updates);
    setWorkflow(updated);
    return updated;
  };

  return { workflow, loading, error, save };
}
```

#### Compound Component Pattern
Create flexible, related components.

```typescript
interface NodePanelContextValue {
  expanded: boolean;
  toggle: () => void;
}

const NodePanelContext = createContext<NodePanelContextValue | null>(null);

function NodePanel({ children }: { children: React.ReactNode }) {
  const [expanded, setExpanded] = useState(false);
  const toggle = () => setExpanded(!expanded);

  return (
    <NodePanelContext.Provider value={{ expanded, toggle }}>
      <div className="node-panel">{children}</div>
    </NodePanelContext.Provider>
  );
}

NodePanel.Header = function Header({ children }: { children: React.ReactNode }) {
  const { toggle } = useContext(NodePanelContext)!;
  return <div onClick={toggle}>{children}</div>;
};

NodePanel.Content = function Content({ children }: { children: React.ReactNode }) {
  const { expanded } = useContext(NodePanelContext)!;
  return expanded ? <div>{children}</div> : null;
};

// Usage
<NodePanel>
  <NodePanel.Header>Click to expand</NodePanel.Header>
  <NodePanel.Content>Panel content here</NodePanel.Content>
</NodePanel>
```

#### Render Props / Children as Function
Share code between components using a prop whose value is a function.

```typescript
interface DataFetcherProps<T> {
  url: string;
  children: (data: T | null, loading: boolean, error: Error | null) => React.ReactNode;
}

function DataFetcher<T>({ url, children }: DataFetcherProps<T>) {
  const [data, setData] = useState<T | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<Error | null>(null);

  useEffect(() => {
    fetch(url)
      .then(res => res.json())
      .then(setData)
      .catch(setError)
      .finally(() => setLoading(false));
  }, [url]);

  return <>{children(data, loading, error)}</>;
}

// Usage
<DataFetcher<Workflow[]> url="/api/workflows">
  {(workflows, loading, error) => {
    if (loading) return <Spinner />;
    if (error) return <Error message={error.message} />;
    return <WorkflowList workflows={workflows!} />;
  }}
</DataFetcher>
```

## Commands Available

| Command | Description | Usage |
|---------|-------------|-------|
| `/review` | Review code for SOLID, DRY, complexity issues | `/review internal/workflow/service.go` |
| `/test-plan` | Generate comprehensive test plan | `/test-plan workflow execution feature` |
| `/refactor` | Suggest refactoring opportunities | `/refactor internal/api/handlers/` |
| `/tdd` | Start TDD workflow for a feature | `/tdd implement webhook signature verification` |
| `/architecture` | Review architectural design | `/architecture internal/workflow/` |
| `/gen-tests` | Generate test code for a file/function | `/gen-tests internal/workflow/service.go` |

## Quick Reference

### Before Writing Code
1. Check TASKS.md for current task
2. Run `/test-plan` to design tests
3. Run `/tdd` to start implementation

### During Code Review
1. Run `/review` on changed files
2. Run `/architecture` for structural changes
3. Check cognitive complexity < 15

### When Refactoring
1. Ensure tests exist first
2. Run `/refactor` for suggestions
3. Make small, incremental changes
4. Run tests after each change

## Gorax Expert Agents

Use these specialized agents for domain-specific tasks. Launch with the Task tool using the appropriate `subagent_type`.

### gorax-workflow-expert

**When to use**: Workflow execution, actions, triggers, expressions, node graph traversal

**Key files**:
- `internal/executor/` - Execution engine, action handlers
- `internal/workflow/` - Workflow models, repository, service
- `internal/executor/expression/` - CEL expression evaluation
- `internal/executor/actions/` - HTTP, Transform, Condition actions

**Domain knowledge**:
- Executor uses visitor pattern for node traversal
- Actions implement `Action` interface with `Execute(ctx, input) (output, error)`
- Expressions use cel-go for dynamic evaluation
- Workflows stored as JSON graph with nodes and edges
- Execution creates step records for each node processed
- Triggers: webhook, schedule, manual
- Node types: trigger, action, condition, loop, parallel

**Patterns**:
```go
// Action execution flow
type Action interface {
    Execute(ctx context.Context, input map[string]any) (map[string]any, error)
}

// Node traversal
executor.Visit(node) -> resolveInputs() -> executeAction() -> storeStepResult()
```

### gorax-credential-expert

**When to use**: Credential encryption, injection, masking, security, audit logging

**Key files**:
- `internal/credential/encryption.go` - Envelope encryption (DEK/KEK)
- `internal/credential/injector.go` - Credential injection into workflows
- `internal/credential/masker.go` - Output masking
- `internal/credential/service_impl.go` - CRUD operations
- `internal/credential/repository.go` - Database operations

**Domain knowledge**:
- Envelope encryption: DEK encrypts data, KEK (master key or KMS) encrypts DEK
- AES-256-GCM for symmetric encryption
- Credentials referenced as `{{credentials.name}}` in workflow configs
- Injector extracts references, decrypts values, replaces in config
- Masker replaces sensitive values with `[REDACTED]` in outputs
- Access logging tracks all credential reads/updates/rotations
- SimpleEncryptionService for dev, KMS for production

**Patterns**:
```go
// Credential injection flow
injector.InjectCredentials(config, ctx) ->
    ExtractCredentialReferences(config) ->
    getCredentialValue(name) ->  // decrypt
    injectValues(config, credentials) ->
    return InjectResult{Config, Values}

// Masking flow
masker.MaskOutput(output, credentialValues) -> replaces all occurrences
```

### gorax-worker-expert

**When to use**: Queue processing, SQS, job orchestration, concurrency, retries

**Key files**:
- `internal/worker/worker.go` - Main worker loop, message processing
- `internal/worker/health.go` - Health checks
- `internal/queue/` - SQS client, message types
- `internal/config/config.go` - Worker and queue configuration

**Domain knowledge**:
- Workers poll SQS for execution messages
- Tenant concurrency limits prevent resource exhaustion
- Dead-letter queue (DLQ) for failed messages after max retries
- Visibility timeout prevents duplicate processing
- Health endpoint reports worker status, queue depth
- Graceful shutdown waits for in-flight executions

**Configuration**:
```go
WorkerConfig {
    Concurrency             int  // Max parallel executions
    MaxConcurrencyPerTenant int  // Per-tenant limit
    HealthPort              string
    QueueURL                string
}

QueueConfig {
    MaxMessages        int32  // Batch size
    WaitTimeSeconds    int32  // Long polling
    VisibilityTimeout  int32  // Processing window
    MaxRetries         int    // Before DLQ
}
```

**Patterns**:
```go
// Worker loop
for {
    messages := queue.ReceiveMessages()
    for _, msg := range messages {
        if !tenantLimiter.Allow(msg.TenantID) {
            requeue(msg)  // Tenant at capacity
            continue
        }
        go processMessage(msg)
    }
}
```

### gorax-frontend-expert

**When to use**: React canvas editor, workflow visualization, state management, API integration

**Key files**:
- `web/src/components/canvas/` - ReactFlow canvas, nodes, edges
- `web/src/pages/WorkflowEditor.tsx` - Main editor page
- `web/src/stores/` - Zustand state stores
- `web/src/hooks/` - Custom React hooks
- `web/src/api/` - API client with TanStack Query

**Domain knowledge**:
- ReactFlow for node-based workflow canvas
- Custom node types: TriggerNode, ActionNode, ConditionNode
- Zustand for canvas state (nodes, edges, selection)
- TanStack Query for server state (workflows, executions)
- Property panel for node configuration
- Real-time execution updates via WebSocket

**Patterns**:
```typescript
// Canvas state (Zustand)
const useCanvasStore = create<CanvasState>((set) => ({
  nodes: [],
  edges: [],
  addNode: (node) => set((state) => ({ nodes: [...state.nodes, node] })),
  updateNode: (id, data) => set((state) => ({
    nodes: state.nodes.map(n => n.id === id ? { ...n, data } : n)
  })),
}));

// API hooks (TanStack Query)
const useWorkflow = (id: string) => useQuery({
  queryKey: ['workflow', id],
  queryFn: () => api.workflows.get(id),
});

const useUpdateWorkflow = () => useMutation({
  mutationFn: (data) => api.workflows.update(data.id, data),
  onSuccess: () => queryClient.invalidateQueries(['workflow']),
});
```

### gorax-webhook-expert

**When to use**: Webhook handling, event filtering, replay, signature verification

**Key files**:
- `internal/webhook/` - Webhook models, service, repository
- `internal/webhook/filter.go` - Event filtering with JSONPath
- `internal/webhook/replay.go` - Event replay service
- `internal/api/handlers/webhook*.go` - HTTP handlers

**Domain knowledge**:
- Webhooks trigger workflow executions
- Secret-based signature verification (HMAC-SHA256)
- Filters use JSONPath expressions with operators (eq, ne, contains, etc.)
- Event history stored for replay and debugging
- Replay re-executes workflow with historical payload

**Patterns**:
```go
// Webhook handling flow
handler.Handle(request) ->
    verifySignature(secret, payload) ->
    storeEvent(webhookID, payload, headers) ->
    evaluateFilters(filters, payload) ->
    if passesFilters: triggerWorkflow(workflowID, payload)

// Filter evaluation
filter.Evaluate(payload) ->
    extractValue(jsonPath) ->
    applyOperator(operator, value, expected) ->
    return bool
```

### Using Expert Agents

Launch an expert agent for complex domain-specific tasks:

```
Task tool with:
- subagent_type: "gorax-workflow-expert"
- prompt: "Implement parallel node execution with proper error handling..."
```

The agent will have context about patterns, file locations, and domain concepts specific to that area of the codebase.
