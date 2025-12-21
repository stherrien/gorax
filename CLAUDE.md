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

### gorax-ai-expert

**When to use**: AI/LLM integrations, chat completions, embeddings, entity extraction, classification, summarization

**Key files**:
- `internal/llm/` - LLM provider abstraction layer
- `internal/llm/providers/anthropic/` - Claude/Anthropic integration
- `internal/llm/providers/openai/` - OpenAI GPT integration
- `internal/llm/providers/bedrock/` - AWS Bedrock integration
- `internal/integrations/ai/` - AI workflow actions
- `internal/quota/ai_tracker.go` - AI usage tracking

**Domain knowledge**:
- Multi-provider abstraction: OpenAI, Anthropic, AWS Bedrock
- Provider interface: `ChatCompletion()`, `GenerateEmbeddings()`, `CountTokens()`, `ListModels()`
- Credential-based authentication via credential service
- Token usage tracking for billing/quotas
- AI actions: ChatCompletion, Embedding, EntityExtraction, Classification, Summarization
- Model capabilities: chat, completion, embedding, vision, function calling, JSON mode
- Cost tracking per 1M tokens (input/output)

**Patterns**:
```go
// Provider interface
type Provider interface {
    ChatCompletion(ctx context.Context, req *ChatRequest) (*ChatResponse, error)
    GenerateEmbeddings(ctx context.Context, req *EmbeddingRequest) (*EmbeddingResponse, error)
    CountTokens(text string, model string) (int, error)
    ListModels(ctx context.Context) ([]Model, error)
}

// AI action execution flow
action.Execute(ctx, input) ->
    retrieveCredential(tenantID, credentialID) ->
    createProvider(providerName, credentials) ->
    buildRequest(config) ->
    provider.ChatCompletion(req) ->
    trackUsage(tokens) ->
    return ActionOutput

// Model capability check
if model.HasCapability(llm.CapabilityFunctionCalling) {
    // Use function calling
}
```

**Configuration**:
```go
ProviderConfig {
    APIKey      string        // Authentication key
    MaxRetries  int           // Default: 3
    Timeout     time.Duration // Default: 60s
    BaseURL     string        // Optional proxy override
}

ChatCompletionConfig {
    Provider         string          // "openai", "anthropic", "bedrock"
    Model           string          // Model ID
    Messages        []ChatMessage   // Conversation history
    SystemPrompt    string          // Optional system prompt
    MaxTokens       int             // Response length limit
    Temperature     *float64        // 0-2 randomness
    TopP            *float64        // 0-1 nucleus sampling
}
```

### gorax-architecture-expert

**When to use**: System architecture design, clean architecture patterns, dependency injection, layering, refactoring large features

**Key files**:
- `internal/` - Clean architecture layers
- `internal/domain/` - Business logic (no external dependencies)
- `internal/service/` - Application services (use case orchestration)
- `internal/api/` - HTTP handlers (presentation layer)
- `internal/repository/` - Data access (infrastructure layer)

**Domain knowledge**:
- Clean Architecture: Domain → Service → API/Handler layers
- Dependency direction: inward (API depends on Service, Service depends on Domain)
- Repository pattern abstracts data access from business logic
- Dependency injection via constructor parameters
- Interface segregation: define interfaces where used, not where implemented
- Domain-driven design: workflows, credentials, webhooks, schedules as domain entities
- Event-driven architecture: WebSocket hub, execution observers
- Queue-based async processing: SQS for workflow executions

**Architecture layers**:
```
┌─────────────────────────────────────┐
│    API Layer (HTTP Handlers)        │  ← internal/api/handlers/
│    - Parse requests                  │
│    - Validate input                  │
│    - Call services                   │
│    - Format responses                │
├─────────────────────────────────────┤
│    Service Layer                     │  ← internal/*/service.go
│    - Business logic orchestration    │
│    - Transaction management          │
│    - Cross-cutting concerns          │
├─────────────────────────────────────┤
│    Domain Layer                      │  ← internal/*/model.go
│    - Core entities                   │
│    - Business rules                  │
│    - Domain interfaces               │
├─────────────────────────────────────┤
│    Infrastructure Layer              │  ← internal/*/repository.go
│    - Database access                 │
│    - External APIs                   │
│    - Queue/messaging                 │
└─────────────────────────────────────┘
```

**Patterns**:
```go
// Service layer with dependency injection
type WorkflowService struct {
    repo      WorkflowRepository    // Interface, not concrete
    executor  Executor
    publisher Publisher
}

func NewWorkflowService(repo WorkflowRepository, executor Executor, publisher Publisher) *WorkflowService {
    return &WorkflowService{
        repo:      repo,
        executor:  executor,
        publisher: publisher,
    }
}

// Interface defined where used (service layer), implemented in infrastructure
type WorkflowRepository interface {
    Create(ctx context.Context, workflow *Workflow) error
    GetByID(ctx context.Context, id string) (*Workflow, error)
    Update(ctx context.Context, workflow *Workflow) error
    Delete(ctx context.Context, id string) error
}

// Handler delegates to service
func (h *WorkflowHandler) CreateWorkflow(w http.ResponseWriter, r *http.Request) {
    req := parseRequest(r)
    workflow, err := h.service.Create(r.Context(), req)
    if err != nil {
        respondError(w, err)
        return
    }
    respondJSON(w, workflow)
}
```

**Design principles**:
- Single Responsibility: Each package has one reason to change
- Open/Closed: Extend via interfaces, don't modify existing code
- Liskov Substitution: Implementations must honor interface contracts
- Interface Segregation: Many small interfaces > one large interface
- Dependency Inversion: Depend on abstractions (interfaces), not concretions

### gorax-go-expert

**When to use**: Go best practices, error handling, concurrency, testing, performance optimization, idiomatic Go

**Key patterns**:
- Error wrapping: `fmt.Errorf("operation failed: %w", err)`
- Context propagation for cancellation and deadlines
- Goroutines with `sync.WaitGroup` or `errgroup.Group`
- Table-driven tests with subtests
- Interface-based design for testability
- Struct embedding for composition over inheritance

**Best practices**:
```go
// Error handling with context
func (s *Service) Process(ctx context.Context, id string) error {
    item, err := s.repo.Get(ctx, id)
    if err != nil {
        return fmt.Errorf("failed to get item %s: %w", id, err)
    }

    if err := s.validate(item); err != nil {
        return fmt.Errorf("validation failed: %w", err)
    }

    return nil
}

// Concurrency with errgroup
func (s *Service) ProcessBatch(ctx context.Context, ids []string) error {
    g, ctx := errgroup.WithContext(ctx)
    g.SetLimit(10) // Max 10 concurrent goroutines

    for _, id := range ids {
        id := id // Capture loop variable
        g.Go(func() error {
            return s.Process(ctx, id)
        })
    }

    return g.Wait()
}

// Table-driven tests
func TestValidate(t *testing.T) {
    tests := []struct {
        name    string
        input   Input
        wantErr bool
    }{
        {
            name:    "valid input",
            input:   Input{Value: "test"},
            wantErr: false,
        },
        {
            name:    "empty value",
            input:   Input{Value: ""},
            wantErr: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := Validate(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}

// Dependency injection with interfaces
type Service struct {
    repo Repository // Interface, not struct
}

func NewService(repo Repository) *Service {
    return &Service{repo: repo}
}

// Mock for testing
type MockRepository struct {
    GetFunc func(ctx context.Context, id string) (*Item, error)
}

func (m *MockRepository) Get(ctx context.Context, id string) (*Item, error) {
    return m.GetFunc(ctx, id)
}
```

**Common anti-patterns to avoid**:
- Ignoring errors: `_ = doSomething()`
- Naked returns in long functions
- Global variables for state
- Mutex without defer unlock
- Goroutine leaks (no cancellation)
- Interface pollution (interfaces with 10+ methods)

### gorax-react-expert

**When to use**: React components, hooks, state management, performance optimization, component patterns

**Key files**:
- `web/src/components/` - Reusable UI components
- `web/src/pages/` - Route-level components
- `web/src/hooks/` - Custom React hooks
- `web/src/stores/` - Zustand state stores
- `web/src/api/` - TanStack Query hooks

**Domain knowledge**:
- Functional components only (no class components)
- Custom hooks for reusable logic
- Zustand for global state (canvas, credentials, auth)
- TanStack Query for server state (workflows, executions, caching)
- Compound component pattern for flexible APIs
- Render props for component composition
- Error boundaries for graceful error handling

**Patterns**:
```typescript
// Custom hook for data fetching with TanStack Query
function useWorkflow(id: string) {
  return useQuery({
    queryKey: ['workflow', id],
    queryFn: () => api.workflows.get(id),
    staleTime: 5 * 60 * 1000, // 5 minutes
  })
}

function useUpdateWorkflow() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (data: UpdateWorkflowRequest) =>
      api.workflows.update(data.id, data),
    onSuccess: (_, variables) => {
      // Invalidate and refetch
      queryClient.invalidateQueries(['workflow', variables.id])
    },
  })
}

// Zustand store for local state
import { create } from 'zustand'

interface CanvasState {
  nodes: Node[]
  edges: Edge[]
  selectedNode: Node | null
  addNode: (node: Node) => void
  updateNode: (id: string, data: Partial<Node['data']>) => void
  selectNode: (id: string | null) => void
}

const useCanvasStore = create<CanvasState>((set) => ({
  nodes: [],
  edges: [],
  selectedNode: null,

  addNode: (node) =>
    set((state) => ({ nodes: [...state.nodes, node] })),

  updateNode: (id, data) =>
    set((state) => ({
      nodes: state.nodes.map((n) =>
        n.id === id ? { ...n, data: { ...n.data, ...data } } : n
      ),
    })),

  selectNode: (id) =>
    set((state) => ({
      selectedNode: state.nodes.find((n) => n.id === id) || null,
    })),
}))

// Compound component pattern
interface PanelContextValue {
  expanded: boolean
  toggle: () => void
}

const PanelContext = createContext<PanelContextValue | null>(null)

function Panel({ children }: { children: React.ReactNode }) {
  const [expanded, setExpanded] = useState(false)
  const toggle = useCallback(() => setExpanded((e) => !e), [])

  return (
    <PanelContext.Provider value={{ expanded, toggle }}>
      <div className="panel">{children}</div>
    </PanelContext.Provider>
  )
}

Panel.Header = function Header({ children }: { children: React.ReactNode }) {
  const context = useContext(PanelContext)
  if (!context) throw new Error('Panel.Header must be within Panel')

  return (
    <div onClick={context.toggle} className="panel-header">
      {children}
    </div>
  )
}

Panel.Content = function Content({ children }: { children: React.ReactNode }) {
  const context = useContext(PanelContext)
  if (!context) throw new Error('Panel.Content must be within Panel')

  return context.expanded ? <div className="panel-content">{children}</div> : null
}

// Usage
<Panel>
  <Panel.Header>Click to expand</Panel.Header>
  <Panel.Content>Panel content here</Panel.Content>
</Panel>
```

**Performance optimization**:
```typescript
// Memoize expensive computations
const sortedItems = useMemo(() => {
  return items.sort((a, b) => a.name.localeCompare(b.name))
}, [items])

// Memoize callbacks to prevent child re-renders
const handleClick = useCallback((id: string) => {
  console.log('Clicked', id)
}, [])

// Memoize components
const MemoizedChild = memo(ChildComponent, (prev, next) => {
  return prev.id === next.id && prev.data === next.data
})
```

### gorax-security-expert

**When to use**: Security audits, vulnerability assessment, authentication/authorization, encryption, input validation, OWASP compliance

**Key files**:
- `internal/api/middleware/auth.go` - Ory Kratos authentication
- `internal/api/middleware/security_headers.go` - Security headers (CSP, HSTS, etc.)
- `internal/api/middleware/ratelimit.go` - Rate limiting per tenant
- `internal/api/middleware/cors.go` - CORS configuration
- `internal/rbac/` - Role-based access control
- `internal/credential/encryption.go` - Envelope encryption (AES-256-GCM)
- `internal/credential/masker.go` - Sensitive data masking
- `internal/validation/` - Input validation

**Domain knowledge**:
- **Authentication**: Ory Kratos session validation, JWT bearer tokens, cookie-based sessions
- **Authorization**: RBAC with roles, permissions, and resource-action pairs
- **Encryption**: Envelope encryption (DEK/KEK), AES-256-GCM, KMS integration
- **Multi-tenancy**: Tenant isolation, per-tenant rate limits, tenant context validation
- **Rate limiting**: Sliding window algorithm (per minute/hour/day), Redis-backed
- **Security headers**: CSP, HSTS, X-Frame-Options, X-Content-Type-Options, Referrer-Policy
- **Credential masking**: Redact sensitive values from logs and API responses
- **Input validation**: SQL injection prevention, XSS protection, command injection prevention
- **OWASP Top 10**: Addressed through middleware, encryption, validation

**Security layers**:
```
┌─────────────────────────────────────┐
│    Security Headers Middleware      │  ← CSP, HSTS, X-Frame-Options
├─────────────────────────────────────┤
│    CORS Middleware                  │  ← Origin validation
├─────────────────────────────────────┤
│    Rate Limit Middleware            │  ← DDoS protection, abuse prevention
├─────────────────────────────────────┤
│    Authentication Middleware        │  ← Kratos session validation
├─────────────────────────────────────┤
│    Tenant Middleware                │  ← Tenant isolation
├─────────────────────────────────────┤
│    RBAC Middleware                  │  ← Permission checks
├─────────────────────────────────────┤
│    Handlers (Input Validation)      │  ← Sanitize, validate inputs
├─────────────────────────────────────┤
│    Services (Business Logic)        │  ← Secure operations
├─────────────────────────────────────┤
│    Encryption Layer                 │  ← Envelope encryption
└─────────────────────────────────────┘
```

**Patterns**:
```go
// Authentication with Ory Kratos
func KratosAuth(cfg config.KratosConfig) func(next http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            sessionToken := extractSessionToken(r) // Cookie or Bearer token
            if sessionToken == "" {
                http.Error(w, "unauthorized", http.StatusUnauthorized)
                return
            }

            user, err := validateKratosSession(cfg.PublicURL, sessionToken)
            if err != nil {
                http.Error(w, "invalid session", http.StatusUnauthorized)
                return
            }

            ctx := context.WithValue(r.Context(), UserContextKey, user)
            next.ServeHTTP(w, r.WithContext(ctx))
        })
    }
}

// RBAC permission check
func RequirePermission(repo Repository, resource, action string) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            userID := GetUserID(r)
            tenantID := GetTenantID(r)

            hasPermission, err := repo.HasPermission(r.Context(), userID, tenantID, resource, action)
            if err != nil || !hasPermission {
                http.Error(w, "forbidden", http.StatusForbidden)
                return
            }

            next.ServeHTTP(w, r)
        })
    }
}

// Rate limiting (sliding window)
func RateLimitMiddleware(redisClient *redis.Client, config RateLimitConfig) func(http.Handler) http.Handler {
    limiter := ratelimit.NewSlidingWindowLimiter(redisClient)

    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            tenantID := GetTenantID(r)

            allowed, err := limiter.Allow(r.Context(), tenantID, config.RequestsPerMinute, time.Minute)
            if err != nil {
                // Fail open on error (allow request)
                next.ServeHTTP(w, r)
                return
            }

            if !allowed {
                w.Header().Set("Retry-After", "60")
                http.Error(w, "rate limit exceeded", http.StatusTooManyRequests)
                return
            }

            next.ServeHTTP(w, r)
        })
    }
}

// Envelope encryption (DEK + KEK)
func (s *EncryptionService) Encrypt(ctx context.Context, data *CredentialData) ([]byte, []byte, error) {
    // 1. Serialize data
    plaintext, err := json.Marshal(data)
    if err != nil {
        return nil, nil, err
    }

    // 2. Generate data encryption key (DEK) via KMS
    plainKey, encryptedKey, err := s.kmsClient.GenerateDataKey(ctx, keyID, nil)
    if err != nil {
        return nil, nil, err
    }
    defer ClearKey(plainKey) // Zero out key from memory

    // 3. Encrypt data with DEK using AES-256-GCM
    encryptedData, err := s.encryptWithAESGCM(plaintext, plainKey)
    if err != nil {
        return nil, nil, err
    }

    // Return encrypted data and encrypted DEK
    return encryptedData, encryptedKey, nil
}

// Security headers
func SecurityHeaders(cfg SecurityHeadersConfig) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            // Prevent MIME sniffing
            w.Header().Set("X-Content-Type-Options", "nosniff")

            // Clickjacking protection
            w.Header().Set("X-Frame-Options", cfg.FrameOptions) // "DENY" or "SAMEORIGIN"

            // XSS protection
            w.Header().Set("X-XSS-Protection", "1; mode=block")

            // Content Security Policy
            w.Header().Set("Content-Security-Policy", cfg.CSPDirectives)

            // Referrer policy
            w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")

            // HSTS (force HTTPS)
            if cfg.EnableHSTS {
                w.Header().Set("Strict-Transport-Security",
                    fmt.Sprintf("max-age=%d; includeSubDomains", cfg.HSTSMaxAge))
            }

            // Permissions policy
            w.Header().Set("Permissions-Policy", "geolocation=(), microphone=(), camera=()")

            next.ServeHTTP(w, r)
        })
    }
}

// Credential masking (prevent leaks in logs/responses)
func (m *Masker) MaskOutput(output map[string]interface{}, credentialValues []string) map[string]interface{} {
    masked := make(map[string]interface{})

    for key, value := range output {
        switch v := value.(type) {
        case string:
            masked[key] = m.maskString(v, credentialValues)
        case map[string]interface{}:
            masked[key] = m.MaskOutput(v, credentialValues) // Recursive
        case []interface{}:
            masked[key] = m.maskArray(v, credentialValues)
        default:
            masked[key] = v
        }
    }

    return masked
}

func (m *Masker) maskString(s string, credentialValues []string) string {
    for _, credValue := range credentialValues {
        if credValue != "" && strings.Contains(s, credValue) {
            s = strings.ReplaceAll(s, credValue, "[REDACTED]")
        }
    }
    return s
}
```

**OWASP Top 10 mitigations**:
1. **Broken Access Control**: RBAC middleware, tenant isolation, permission checks
2. **Cryptographic Failures**: AES-256-GCM, envelope encryption, KMS, TLS
3. **Injection**: Prepared statements (SQL), input validation, parameterized queries
4. **Insecure Design**: Clean architecture, defense in depth, secure defaults
5. **Security Misconfiguration**: Security headers, HSTS, CSP, secure cookies
6. **Vulnerable Components**: Dependabot, regular updates, minimal dependencies
7. **Authentication Failures**: Ory Kratos, session management, MFA support
8. **Software/Data Integrity**: Signature verification (webhooks), checksums
9. **Logging Failures**: Structured logging, credential masking, audit trails
10. **SSRF**: URL validation, allowlist for external requests

**Security checklist**:
- [ ] Authentication required for all sensitive endpoints
- [ ] Authorization (RBAC) checks on resource access
- [ ] Rate limiting enabled per tenant
- [ ] Input validation on all user inputs
- [ ] SQL queries use parameterized statements
- [ ] Credentials encrypted at rest (envelope encryption)
- [ ] Sensitive data masked in logs and responses
- [ ] Security headers set (CSP, HSTS, X-Frame-Options)
- [ ] CORS configured with specific origins
- [ ] TLS/HTTPS enforced in production
- [ ] Secrets never hardcoded or committed to git
- [ ] Audit logging for sensitive operations
- [ ] Session timeouts configured
- [ ] Multi-tenancy isolation enforced

**Common vulnerabilities to avoid**:
- SQL injection: Always use parameterized queries, never string concatenation
- XSS: Sanitize output, use CSP, escape user input
- CSRF: Use CSRF tokens, validate origin headers
- Command injection: Never pass user input to shell commands
- Path traversal: Validate file paths, use allowlist
- Insecure deserialization: Validate input before unmarshaling
- Information disclosure: Mask credentials, sanitize error messages
- Weak crypto: Use AES-256-GCM, not ECB mode; use strong random sources
- Broken auth: Validate sessions, enforce timeouts, use secure cookies
- Missing rate limits: Prevent brute force, DDoS attacks

### gorax-qa-expert

**When to use**: Test strategy, test coverage analysis, integration testing, E2E testing, test automation, quality assurance

**Key files**:
- Go tests: `internal/**/*_test.go` (138+ test files)
- React tests: `web/src/**/*.test.tsx` (60+ test files)
- Integration tests: `internal/credential/integration_test.go`, `web/src/pages/*.integration.test.tsx`
- Test utilities: `internal/testutil/`, `web/src/test/`

**Domain knowledge**:
- **Go testing**: `testing` package, `testify/assert`, `testify/mock`, table-driven tests
- **React testing**: Vitest, React Testing Library, `@testing-library/user-event`
- **Test types**: Unit, integration, E2E, benchmark, performance
- **Test patterns**: AAA (Arrange-Act-Assert), Given-When-Then, table-driven
- **Mocking**: Mock repositories, mock external services, mock HTTP clients
- **Coverage**: 80%+ on business logic, lower tolerance on UI
- **TDD workflow**: Red-Green-Refactor

**Testing pyramid**:
```
┌─────────────────────────────────────┐
│         E2E Tests (Few)              │  ← Playwright, full system
├─────────────────────────────────────┤
│    Integration Tests (Some)          │  ← API + DB, component + hooks
├─────────────────────────────────────┤
│      Unit Tests (Many)               │  ← Functions, services, components
└─────────────────────────────────────┘
```

**Go testing patterns**:
```go
// Table-driven test (preferred pattern)
func TestValidateWorkflow(t *testing.T) {
    tests := []struct {
        name    string
        input   *Workflow
        wantErr bool
        errMsg  string
    }{
        {
            name: "valid workflow",
            input: &Workflow{
                Name:        "Test Workflow",
                Description: "Test",
                Nodes:       []Node{{Type: "trigger"}},
            },
            wantErr: false,
        },
        {
            name: "empty name",
            input: &Workflow{
                Name:  "",
                Nodes: []Node{{Type: "trigger"}},
            },
            wantErr: true,
            errMsg:  "name is required",
        },
        {
            name: "no trigger node",
            input: &Workflow{
                Name:  "Test",
                Nodes: []Node{{Type: "action"}},
            },
            wantErr: true,
            errMsg:  "workflow must have a trigger",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := ValidateWorkflow(tt.input)

            if tt.wantErr {
                require.Error(t, err)
                assert.Contains(t, err.Error(), tt.errMsg)
            } else {
                require.NoError(t, err)
            }
        })
    }
}

// Mock usage with testify/mock
func TestWorkflowService_Create(t *testing.T) {
    // Arrange
    mockRepo := new(MockRepository)
    mockExecutor := new(MockExecutor)
    service := NewWorkflowService(mockRepo, mockExecutor)

    input := CreateWorkflowInput{Name: "Test Workflow"}
    expectedWorkflow := &Workflow{
        ID:   "wf-123",
        Name: "Test Workflow",
    }

    mockRepo.On("Create", mock.Anything, "tenant-1", "user-1", input).
        Return(expectedWorkflow, nil)

    // Act
    result, err := service.Create(context.Background(), "tenant-1", "user-1", input)

    // Assert
    require.NoError(t, err)
    assert.Equal(t, expectedWorkflow.ID, result.ID)
    assert.Equal(t, expectedWorkflow.Name, result.Name)
    mockRepo.AssertExpectations(t)
}

// Integration test (with real dependencies)
func TestWorkflowRepository_Integration(t *testing.T) {
    if testing.Short() {
        t.Skip("skipping integration test")
    }

    // Setup test database
    db := setupTestDB(t)
    defer db.Close()

    repo := NewWorkflowRepository(db)

    // Test create
    workflow, err := repo.Create(context.Background(), "tenant-1", "user-1", CreateWorkflowInput{
        Name:        "Integration Test",
        Description: "Test workflow",
    })
    require.NoError(t, err)
    assert.NotEmpty(t, workflow.ID)

    // Test get
    retrieved, err := repo.GetByID(context.Background(), "tenant-1", workflow.ID)
    require.NoError(t, err)
    assert.Equal(t, workflow.Name, retrieved.Name)
}

// Benchmark test
func BenchmarkFormulaEvaluation(b *testing.B) {
    evaluator := NewFormulaEvaluator()
    formula := "sum(map(items, item => item.value))"
    data := map[string]interface{}{
        "items": []map[string]interface{}{
            {"value": 10},
            {"value": 20},
            {"value": 30},
        },
    }

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _, _ = evaluator.Evaluate(formula, data)
    }
}
```

**React testing patterns**:
```typescript
// Component unit test
import { describe, it, expect, vi } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import WorkflowList from './WorkflowList'

describe('WorkflowList', () => {
  it('should render workflows', () => {
    const workflows = [
      { id: '1', name: 'Workflow 1' },
      { id: '2', name: 'Workflow 2' },
    ]

    render(<WorkflowList workflows={workflows} />)

    expect(screen.getByText('Workflow 1')).toBeInTheDocument()
    expect(screen.getByText('Workflow 2')).toBeInTheDocument()
  })

  it('should handle workflow selection', async () => {
    const user = userEvent.setup()
    const onSelect = vi.fn()
    const workflows = [{ id: '1', name: 'Workflow 1' }]

    render(<WorkflowList workflows={workflows} onSelect={onSelect} />)

    await user.click(screen.getByText('Workflow 1'))

    expect(onSelect).toHaveBeenCalledWith(workflows[0])
  })

  it('should show loading state', () => {
    render(<WorkflowList workflows={[]} loading />)

    expect(screen.getByRole('progressbar')).toBeInTheDocument()
  })

  it('should show error message', () => {
    render(<WorkflowList workflows={[]} error="Failed to load" />)

    expect(screen.getByText(/failed to load/i)).toBeInTheDocument()
  })
})

// Hook testing
import { renderHook, waitFor } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { useWorkflow } from './useWorkflow'

describe('useWorkflow', () => {
  const createWrapper = () => {
    const queryClient = new QueryClient({
      defaultOptions: { queries: { retry: false } },
    })
    return ({ children }) => (
      <QueryClientProvider client={queryClient}>
        {children}
      </QueryClientProvider>
    )
  }

  it('should fetch workflow data', async () => {
    const { result } = renderHook(() => useWorkflow('wf-123'), {
      wrapper: createWrapper(),
    })

    await waitFor(() => expect(result.current.isSuccess).toBe(true))

    expect(result.current.data).toEqual({
      id: 'wf-123',
      name: 'Test Workflow',
    })
  })

  it('should handle errors', async () => {
    // Mock API to return error
    vi.mocked(api.workflows.get).mockRejectedValue(new Error('Not found'))

    const { result } = renderHook(() => useWorkflow('invalid'), {
      wrapper: createWrapper(),
    })

    await waitFor(() => expect(result.current.isError).toBe(true))

    expect(result.current.error).toBeDefined()
  })
})

// Integration test with MSW (Mock Service Worker)
import { setupServer } from 'msw/node'
import { rest } from 'msw'

const server = setupServer(
  rest.get('/api/workflows/:id', (req, res, ctx) => {
    const { id } = req.params
    return res(
      ctx.json({
        id,
        name: 'Test Workflow',
        nodes: [],
        edges: [],
      })
    )
  })
)

beforeAll(() => server.listen())
afterEach(() => server.resetHandlers())
afterAll(() => server.close())

it('should load workflow from API', async () => {
  render(<WorkflowEditor id="wf-123" />)

  await waitFor(() => {
    expect(screen.getByText('Test Workflow')).toBeInTheDocument()
  })
})
```

**Test coverage targets**:
- **Business logic**: 80-90% coverage (services, domain models)
- **API handlers**: 70-80% coverage
- **Utilities**: 90%+ coverage
- **UI components**: 60-70% coverage (focus on critical paths)
- **Integration points**: 100% of critical flows

**Testing best practices**:
1. **AAA pattern**: Arrange (setup), Act (execute), Assert (verify)
2. **One assertion per test**: Keep tests focused and debuggable
3. **Test behavior, not implementation**: Avoid testing internal details
4. **Use descriptive test names**: `should_return_error_when_name_is_empty`
5. **Isolate tests**: Each test should be independent
6. **Mock external dependencies**: Network, database, filesystem
7. **Test edge cases**: Empty inputs, nil values, boundary conditions
8. **Performance tests**: Benchmark critical paths
9. **Test failure paths**: Errors, timeouts, invalid inputs
10. **Keep tests fast**: Unit tests < 100ms, integration tests < 1s

**Common test smells**:
- Flaky tests (intermittent failures)
- Slow tests (> 1s for unit tests)
- Tests that depend on execution order
- Tests with hardcoded sleep/timeouts
- Tests that test multiple things
- Unclear test names (test1, test2)
- Missing cleanup (defer, afterEach)
- Over-mocking (mocking too much)

**QA checklist**:
- [ ] All new code has unit tests (80%+ coverage)
- [ ] Critical paths have integration tests
- [ ] Edge cases tested (empty, nil, boundary)
- [ ] Error cases tested
- [ ] Happy path tested
- [ ] API contract tests written
- [ ] Performance benchmarks for critical code
- [ ] Tests pass in CI/CD
- [ ] No flaky tests
- [ ] Tests run fast (< 5s for unit suite)

### gorax-reactflow-expert

**When to use**: ReactFlow canvas, node-based editors, workflow visualization, custom nodes, DAG validation

**Key files**:
- `web/src/components/canvas/WorkflowCanvas.tsx` - Main canvas component
- `web/src/components/nodes/` - Custom node types
- `web/src/utils/dagValidation.ts` - Cycle detection
- `web/src/components/canvas/PropertyPanel.tsx` - Node configuration
- `web/src/components/canvas/NodePalette.tsx` - Drag-and-drop palette

**Domain knowledge**:
- ReactFlow v12 (@xyflow/react)
- Custom node types: TriggerNode, ActionNode, ConditionNode, LoopNode, ParallelNode
- DAG validation with cycle detection
- Controlled vs uncontrolled mode
- Node state management via Zustand
- Edge validation before connection
- Custom handles for input/output ports
- Node data synchronization with backend workflow definition

**Patterns**:
```typescript
// Canvas setup with state management
function WorkflowCanvas({ initialNodes, initialEdges }: Props) {
  const [nodes, setNodes, onNodesChange] = useNodesState(initialNodes)
  const [edges, setEdges, onEdgesChange] = useEdgesState(initialEdges)
  const { screenToFlowPosition } = useReactFlow()

  // Validate before connecting
  const onConnect = useCallback((connection: Connection) => {
    const newEdge: Edge = {
      id: `e${connection.source}-${connection.target}`,
      source: connection.source!,
      target: connection.target!,
    }

    // Check for cycles
    const cycles = detectCycles(nodes, [...edges, newEdge])
    if (cycles.length > 0) {
      alert('Cannot create cycle')
      return
    }

    setEdges((eds) => addEdge(connection, eds))
  }, [nodes, edges, setEdges])

  return (
    <ReactFlow
      nodes={nodes}
      edges={edges}
      onNodesChange={onNodesChange}
      onEdgesChange={onEdgesChange}
      onConnect={onConnect}
      nodeTypes={nodeTypes}
      fitView
    >
      <Background />
      <Controls />
      <MiniMap />
    </ReactFlow>
  )
}

// Custom node type
function ActionNode({ data, id }: NodeProps<ActionNodeData>) {
  const updateNode = useCanvasStore((state) => state.updateNode)

  return (
    <div className="action-node">
      <Handle type="target" position={Position.Top} />

      <div className="node-header">
        <IconComponent icon={data.actionType} />
        <span>{data.label}</span>
      </div>

      <div className="node-body">
        {data.config && (
          <ConfigPreview config={data.config} />
        )}
      </div>

      <Handle type="source" position={Position.Bottom} />
    </div>
  )
}

// DAG validation (cycle detection)
function detectCycles(nodes: Node[], edges: Edge[]): string[][] {
  const graph = new Map<string, string[]>()

  // Build adjacency list
  for (const edge of edges) {
    if (!graph.has(edge.source)) {
      graph.set(edge.source, [])
    }
    graph.get(edge.source)!.push(edge.target)
  }

  const visited = new Set<string>()
  const recStack = new Set<string>()
  const cycles: string[][] = []

  function dfs(nodeId: string, path: string[]): void {
    visited.add(nodeId)
    recStack.add(nodeId)
    path.push(nodeId)

    const neighbors = graph.get(nodeId) || []
    for (const neighbor of neighbors) {
      if (!visited.has(neighbor)) {
        dfs(neighbor, [...path])
      } else if (recStack.has(neighbor)) {
        // Found cycle
        const cycleStart = path.indexOf(neighbor)
        cycles.push([...path.slice(cycleStart), neighbor])
      }
    }

    recStack.delete(nodeId)
  }

  for (const node of nodes) {
    if (!visited.has(node.id)) {
      dfs(node.id, [])
    }
  }

  return cycles
}

// Node palette with drag-and-drop
function NodePalette() {
  const onDragStart = (event: React.DragEvent, nodeType: string) => {
    event.dataTransfer.setData('application/reactflow', nodeType)
    event.dataTransfer.effectAllowed = 'move'
  }

  return (
    <aside className="node-palette">
      <div
        draggable
        onDragStart={(e) => onDragStart(e, 'action')}
        className="palette-item"
      >
        Action Node
      </div>
      <div
        draggable
        onDragStart={(e) => onDragStart(e, 'condition')}
        className="palette-item"
      >
        Condition Node
      </div>
    </aside>
  )
}
```

**Common patterns**:
- Controlled mode: sync nodes/edges with external state (Zustand)
- Edge validation: prevent cycles, enforce DAG structure
- Custom nodes: use data prop for configuration, handles for connections
- Node selection: track selected node for property panel
- Canvas events: onNodeClick, onEdgeClick, onNodeDragStop
- Position calculation: use `screenToFlowPosition()` for drop zones

### Using Expert Agents

Launch an expert agent for complex domain-specific tasks:

```
Task tool with:
- subagent_type: "gorax-workflow-expert"
- prompt: "Implement parallel node execution with proper error handling..."
```

The agent will have context about patterns, file locations, and domain concepts specific to that area of the codebase.
