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

### gorax-go-security-expert

**When to use**: Go-specific security audits, SQL injection prevention, context security, goroutine safety, secure coding in Go

**Key focus areas**:
- SQL injection prevention with parameterized queries
- Context timeout and cancellation security
- Goroutine leaks and race conditions
- Input validation and sanitization
- Secure random number generation
- Path traversal prevention
- Command injection prevention
- Crypto/TLS best practices in Go

**Go security patterns**:
```go
// ✅ SECURE: Parameterized queries (prevents SQL injection)
func (r *Repository) GetWorkflow(ctx context.Context, tenantID, workflowID string) (*Workflow, error) {
    query := `SELECT id, name, description FROM workflows WHERE tenant_id = $1 AND id = $2`

    var workflow Workflow
    err := r.db.QueryRowContext(ctx, query, tenantID, workflowID).Scan(
        &workflow.ID,
        &workflow.Name,
        &workflow.Description,
    )
    if err != nil {
        return nil, fmt.Errorf("query workflow: %w", err)
    }

    return &workflow, nil
}

// ❌ INSECURE: String concatenation (SQL injection vulnerability)
func (r *Repository) GetWorkflowInsecure(ctx context.Context, tenantID, workflowID string) (*Workflow, error) {
    query := fmt.Sprintf("SELECT * FROM workflows WHERE tenant_id = '%s' AND id = '%s'", tenantID, workflowID)
    // If workflowID contains "'; DROP TABLE workflows; --", database is compromised
    return r.db.QueryContext(ctx, query)
}

// ✅ SECURE: Context with timeout prevents resource exhaustion
func (s *Service) ProcessWithTimeout(workflowID string) error {
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel() // Always cancel to free resources

    workflow, err := s.repo.GetWorkflow(ctx, workflowID)
    if err != nil {
        if errors.Is(err, context.DeadlineExceeded) {
            return fmt.Errorf("operation timed out: %w", err)
        }
        return err
    }

    return s.execute(ctx, workflow)
}

// ✅ SECURE: Goroutine with proper cancellation
func (s *Service) ProcessConcurrently(ctx context.Context, workflowIDs []string) error {
    g, ctx := errgroup.WithContext(ctx)
    g.SetLimit(10) // Limit concurrent goroutines

    for _, id := range workflowIDs {
        id := id // Capture loop variable
        g.Go(func() error {
            select {
            case <-ctx.Done():
                return ctx.Err() // Check cancellation
            default:
                return s.process(ctx, id)
            }
        })
    }

    return g.Wait()
}

// ❌ INSECURE: Goroutine leak (no cancellation)
func (s *Service) ProcessInsecure(workflowIDs []string) {
    for _, id := range workflowIDs {
        go func(id string) {
            // This goroutine will run forever if blocked
            s.process(context.Background(), id)
        }(id)
    }
}

// ✅ SECURE: Path validation (prevents path traversal)
func (s *Service) ReadFile(filename string) ([]byte, error) {
    // Validate filename
    if strings.Contains(filename, "..") {
        return nil, fmt.Errorf("invalid filename: path traversal attempt")
    }

    // Use filepath.Clean to normalize
    cleanPath := filepath.Clean(filename)

    // Ensure it's within allowed directory
    baseDir := "/var/app/data"
    fullPath := filepath.Join(baseDir, cleanPath)

    if !strings.HasPrefix(fullPath, baseDir) {
        return nil, fmt.Errorf("invalid path: outside allowed directory")
    }

    return os.ReadFile(fullPath)
}

// ❌ INSECURE: No path validation (path traversal vulnerability)
func (s *Service) ReadFileInsecure(filename string) ([]byte, error) {
    // filename = "../../etc/passwd" would expose sensitive files
    return os.ReadFile(filename)
}

// ✅ SECURE: Command execution with validation
func (s *Service) ExecuteCommand(command string, args []string) ([]byte, error) {
    // Allowlist of permitted commands
    allowedCommands := map[string]bool{
        "git":  true,
        "node": true,
        "npm":  true,
    }

    if !allowedCommands[command] {
        return nil, fmt.Errorf("command not allowed: %s", command)
    }

    // Validate arguments (no shell metacharacters)
    for _, arg := range args {
        if strings.ContainsAny(arg, ";|&$`\n") {
            return nil, fmt.Errorf("invalid argument: shell metacharacters detected")
        }
    }

    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    cmd := exec.CommandContext(ctx, command, args...)
    return cmd.Output()
}

// ❌ INSECURE: Shell injection vulnerability
func (s *Service) ExecuteCommandInsecure(command string) ([]byte, error) {
    // command = "ls; rm -rf /" would delete everything
    cmd := exec.Command("sh", "-c", command)
    return cmd.Output()
}

// ✅ SECURE: Cryptographically secure random generation
func GenerateSecureToken() (string, error) {
    b := make([]byte, 32)
    if _, err := rand.Read(b); err != nil {
        return "", fmt.Errorf("failed to generate random bytes: %w", err)
    }
    return base64.URLEncoding.EncodeToString(b), nil
}

// ❌ INSECURE: Predictable random (math/rand is not cryptographically secure)
func GenerateInsecureToken() string {
    return fmt.Sprintf("%d", mathrand.Int63())
}

// ✅ SECURE: Input validation with bounds checking
func ValidatePaginationParams(limit, offset string) (int, int, error) {
    const maxLimit = 1000
    const maxOffset = 1000000

    limitInt, err := strconv.Atoi(limit)
    if err != nil || limitInt < 1 || limitInt > maxLimit {
        return 0, 0, fmt.Errorf("invalid limit: must be between 1 and %d", maxLimit)
    }

    offsetInt, err := strconv.Atoi(offset)
    if err != nil || offsetInt < 0 || offsetInt > maxOffset {
        return 0, 0, fmt.Errorf("invalid offset: must be between 0 and %d", maxOffset)
    }

    return limitInt, offsetInt, nil
}

// ✅ SECURE: TLS configuration
func SecureTLSConfig() *tls.Config {
    return &tls.Config{
        MinVersion:               tls.VersionTLS13,
        CurvePreferences:         []tls.CurveID{tls.CurveP521, tls.CurveP384, tls.CurveP256},
        PreferServerCipherSuites: true,
        CipherSuites: []uint16{
            tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
            tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
        },
    }
}
```

**Go security checklist**:
- [ ] All SQL queries use parameterized statements ($1, $2 placeholders)
- [ ] Context with timeout for all external calls (DB, HTTP, etc.)
- [ ] Goroutines have cancellation via context
- [ ] errgroup.Group with SetLimit() for bounded concurrency
- [ ] File paths validated against traversal (no "..")
- [ ] Commands use exec.Command with args array (not shell)
- [ ] crypto/rand used for tokens/keys (not math/rand)
- [ ] Input validation with bounds checking
- [ ] Errors wrapped with context, never expose internal details
- [ ] Defer statements for cleanup (Close, cancel, Unlock)
- [ ] Race detector used in tests: `go test -race ./...`
- [ ] TLS 1.3+ for all external connections

### gorax-react-security-expert

**When to use**: React-specific security audits, XSS prevention, CSRF protection, secure state management, API security in React

**Key focus areas**:
- XSS prevention (dangerouslySetInnerHTML avoidance)
- CSRF token handling
- Secure authentication state management
- API token storage (not in localStorage)
- Content Security Policy compliance
- Third-party dependency security
- Secure routing and authorization

**React security patterns**:
```typescript
// ✅ SECURE: React escapes content by default
function WorkflowName({ name }: { name: string }) {
  return <h1>{name}</h1> // Automatically escaped
}

// ❌ INSECURE: dangerouslySetInnerHTML (XSS vulnerability)
function WorkflowNameInsecure({ html }: { html: string }) {
  // If html = "<img src=x onerror='alert(document.cookie)'/>", XSS occurs
  return <h1 dangerouslySetInnerHTML={{ __html: html }} />
}

// ✅ SECURE: Sanitize before rendering HTML (if absolutely necessary)
import DOMPurify from 'dompurify'

function SanitizedContent({ html }: { html: string }) {
  const clean = DOMPurify.sanitize(html, {
    ALLOWED_TAGS: ['b', 'i', 'em', 'strong', 'p'],
    ALLOWED_ATTR: []
  })
  return <div dangerouslySetInnerHTML={{ __html: clean }} />
}

// ✅ SECURE: Auth token in httpOnly cookie (set by backend)
// NO localStorage/sessionStorage for sensitive tokens
function useAuth() {
  // Token is automatically sent with requests via cookie
  const { data: user } = useQuery({
    queryKey: ['user'],
    queryFn: () => api.auth.getCurrentUser(), // Cookie sent automatically
  })

  return { user, isAuthenticated: !!user }
}

// ❌ INSECURE: Token in localStorage (vulnerable to XSS)
function useAuthInsecure() {
  const [token, setToken] = useState(localStorage.getItem('auth_token'))
  // If XSS exists, attacker can: localStorage.getItem('auth_token')
}

// ✅ SECURE: CSRF token from meta tag
function useCSRFToken() {
  return useMemo(() => {
    const meta = document.querySelector('meta[name="csrf-token"]')
    return meta?.getAttribute('content') || ''
  }, [])
}

// API client with CSRF protection
const apiClient = axios.create({
  baseURL: '/api',
  withCredentials: true, // Send cookies
})

apiClient.interceptors.request.use((config) => {
  const csrfToken = document.querySelector('meta[name="csrf-token"]')?.getAttribute('content')
  if (csrfToken) {
    config.headers['X-CSRF-Token'] = csrfToken
  }
  return config
})

// ✅ SECURE: Protected route with authorization
function ProtectedRoute({ children, requiredPermission }: Props) {
  const { user, isLoading } = useAuth()
  const hasPermission = user?.permissions?.includes(requiredPermission)

  if (isLoading) return <Loading />
  if (!user) return <Navigate to="/login" />
  if (!hasPermission) return <Navigate to="/unauthorized" />

  return <>{children}</>
}

// ❌ INSECURE: Client-side only authorization (can be bypassed)
function InsecureRoute({ children }: Props) {
  const isAdmin = localStorage.getItem('isAdmin') === 'true'
  // Attacker can: localStorage.setItem('isAdmin', 'true')
  return isAdmin ? <>{children}</> : <Navigate to="/" />
}

// ✅ SECURE: Input validation before API call
function WorkflowForm() {
  const [name, setName] = useState('')
  const createWorkflow = useCreateWorkflow()

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()

    // Client-side validation (UX + first line of defense)
    if (!name || name.length > 255) {
      toast.error('Name must be 1-255 characters')
      return
    }

    // Remove any HTML tags
    const sanitizedName = name.replace(/<[^>]*>/g, '')

    createWorkflow.mutate({ name: sanitizedName })
  }

  return (
    <form onSubmit={handleSubmit}>
      <input
        value={name}
        onChange={(e) => setName(e.target.value)}
        maxLength={255}
      />
      <button type="submit">Create</button>
    </form>
  )
}

// ✅ SECURE: Avoid exposing sensitive data in URL
function WorkflowEditor() {
  const { id } = useParams() // Public workflow ID only
  const { data: workflow } = useWorkflow(id)

  // Credentials/secrets fetched separately, never in URL
  const { data: credentials } = useCredentials(workflow?.tenantId)
}

// ❌ INSECURE: Sensitive data in URL
function InsecureEditor() {
  // URL: /editor?apiKey=sk_live_xyz123
  const [searchParams] = useSearchParams()
  const apiKey = searchParams.get('apiKey') // Exposed in browser history/logs
}

// ✅ SECURE: Secure dependency management
// package.json with exact versions (no ^ or ~)
{
  "dependencies": {
    "react": "18.2.0",  // ✅ Exact version
    "axios": "1.6.0"
  }
}

// Run security audits regularly
// npm audit
// npm audit fix

// ✅ SECURE: Content Security Policy compliant
// No inline scripts, no eval()
function SecureComponent() {
  // ❌ Avoid: onclick="handleClick()"
  // ✅ Use: React event handlers
  return <button onClick={handleClick}>Click</button>
}

// ✅ SECURE: Safe navigation
import { useNavigate } from 'react-router-dom'

function SafeRedirect() {
  const navigate = useNavigate()

  // Validate redirect URL
  const handleRedirect = (url: string) => {
    const allowed = ['/dashboard', '/workflows', '/settings']
    if (allowed.includes(url)) {
      navigate(url)
    }
  }
}

// ❌ INSECURE: Open redirect vulnerability
function InsecureRedirect() {
  const [searchParams] = useSearchParams()
  const redirect = searchParams.get('redirect')
  // redirect=https://evil.com would redirect user away
  window.location.href = redirect
}
```

**React security checklist**:
- [ ] Never use dangerouslySetInnerHTML without DOMPurify
- [ ] Auth tokens in httpOnly cookies (not localStorage)
- [ ] CSRF token included in POST/PUT/DELETE requests
- [ ] Authorization checks on both client AND server
- [ ] Input validation before API calls
- [ ] No sensitive data in URLs or browser history
- [ ] Exact dependency versions in package.json
- [ ] Regular npm audit runs
- [ ] CSP compliant (no inline scripts)
- [ ] Validate redirect URLs (whitelist)
- [ ] Use HTTPS only (enforced)
- [ ] Secrets never committed to git
- [ ] React.StrictMode enabled in development

### gorax-go-testing-expert

**When to use**: Writing Go tests, table-driven tests, mocking, benchmarks, integration tests, test coverage improvement

**Key expertise**:
- Table-driven test patterns
- testify/assert and testify/mock usage
- Test fixtures and test data management
- Integration test setup/teardown
- Benchmark tests for performance
- Race condition detection
- Test coverage analysis

**Go testing patterns**:
```go
// ✅ BEST PRACTICE: Table-driven tests with subtests
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
                Description: "Valid workflow",
                Nodes: []Node{
                    {ID: "1", Type: "trigger"},
                    {ID: "2", Type: "action"},
                },
                Edges: []Edge{{From: "1", To: "2"}},
            },
            wantErr: false,
        },
        {
            name: "missing name",
            input: &Workflow{
                Nodes: []Node{{ID: "1", Type: "trigger"}},
            },
            wantErr: true,
            errMsg:  "name is required",
        },
        {
            name: "no trigger node",
            input: &Workflow{
                Name:  "Test",
                Nodes: []Node{{ID: "1", Type: "action"}},
            },
            wantErr: true,
            errMsg:  "workflow must have a trigger node",
        },
        {
            name: "cycle detected",
            input: &Workflow{
                Name: "Cyclic",
                Nodes: []Node{
                    {ID: "1", Type: "trigger"},
                    {ID: "2", Type: "action"},
                },
                Edges: []Edge{
                    {From: "1", To: "2"},
                    {From: "2", To: "1"}, // Creates cycle
                },
            },
            wantErr: true,
            errMsg:  "cycle detected",
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

// ✅ BEST PRACTICE: Mock with testify/mock
type MockRepository struct {
    mock.Mock
}

func (m *MockRepository) GetWorkflow(ctx context.Context, tenantID, id string) (*Workflow, error) {
    args := m.Called(ctx, tenantID, id)
    if args.Get(0) == nil {
        return nil, args.Error(1)
    }
    return args.Get(0).(*Workflow), args.Error(1)
}

func TestWorkflowService_Execute(t *testing.T) {
    // Arrange
    mockRepo := new(MockRepository)
    mockExecutor := new(MockExecutor)
    service := NewWorkflowService(mockRepo, mockExecutor)

    ctx := context.Background()
    tenantID := "tenant-123"
    workflowID := "wf-456"

    expectedWorkflow := &Workflow{
        ID:   workflowID,
        Name: "Test Workflow",
        Nodes: []Node{
            {ID: "1", Type: "trigger"},
            {ID: "2", Type: "action"},
        },
    }

    // Setup mock expectations
    mockRepo.On("GetWorkflow", ctx, tenantID, workflowID).
        Return(expectedWorkflow, nil).
        Once()

    mockExecutor.On("Execute", ctx, expectedWorkflow).
        Return(&ExecutionResult{Success: true}, nil).
        Once()

    // Act
    result, err := service.Execute(ctx, tenantID, workflowID)

    // Assert
    require.NoError(t, err)
    assert.NotNil(t, result)
    assert.True(t, result.Success)

    // Verify all expectations were met
    mockRepo.AssertExpectations(t)
    mockExecutor.AssertExpectations(t)
}

// ✅ BEST PRACTICE: Test helper functions
func setupTestDB(t *testing.T) *sql.DB {
    t.Helper()

    db, err := sql.Open("postgres", "postgresql://localhost/test?sslmode=disable")
    require.NoError(t, err)

    // Run migrations
    err = runMigrations(db)
    require.NoError(t, err)

    // Cleanup on test completion
    t.Cleanup(func() {
        db.Exec("TRUNCATE TABLE workflows CASCADE")
        db.Close()
    })

    return db
}

// ✅ BEST PRACTICE: Integration test with real dependencies
func TestWorkflowRepository_Integration(t *testing.T) {
    if testing.Short() {
        t.Skip("skipping integration test in short mode")
    }

    db := setupTestDB(t)
    repo := NewWorkflowRepository(db)

    ctx := context.Background()
    tenantID := "tenant-123"

    t.Run("create and retrieve workflow", func(t *testing.T) {
        // Create
        input := CreateWorkflowInput{
            Name:        "Integration Test",
            Description: "Test workflow",
        }

        created, err := repo.Create(ctx, tenantID, "user-1", input)
        require.NoError(t, err)
        assert.NotEmpty(t, created.ID)
        assert.Equal(t, input.Name, created.Name)

        // Retrieve
        retrieved, err := repo.GetByID(ctx, tenantID, created.ID)
        require.NoError(t, err)
        assert.Equal(t, created.ID, retrieved.ID)
        assert.Equal(t, created.Name, retrieved.Name)
    })

    t.Run("list with pagination", func(t *testing.T) {
        // Create multiple workflows
        for i := 0; i < 5; i++ {
            _, err := repo.Create(ctx, tenantID, "user-1", CreateWorkflowInput{
                Name: fmt.Sprintf("Workflow %d", i),
            })
            require.NoError(t, err)
        }

        // List with limit
        workflows, err := repo.List(ctx, tenantID, 3, 0)
        require.NoError(t, err)
        assert.Len(t, workflows, 3)
    })
}

// ✅ BEST PRACTICE: Benchmark test
func BenchmarkFormulaEvaluator(b *testing.B) {
    evaluator := NewFormulaEvaluator()

    tests := []struct {
        name    string
        formula string
        data    map[string]interface{}
    }{
        {
            name:    "simple arithmetic",
            formula: "a + b * c",
            data:    map[string]interface{}{"a": 1, "b": 2, "c": 3},
        },
        {
            name:    "array operations",
            formula: "sum(map(items, item => item.value))",
            data: map[string]interface{}{
                "items": []map[string]interface{}{
                    {"value": 10},
                    {"value": 20},
                    {"value": 30},
                },
            },
        },
        {
            name:    "complex nested",
            formula: "filter(map(items, i => i.value * 2), v => v > 20)",
            data: map[string]interface{}{
                "items": makeTestItems(100),
            },
        },
    }

    for _, tt := range tests {
        b.Run(tt.name, func(b *testing.B) {
            b.ResetTimer()
            for i := 0; i < b.N; i++ {
                _, _ = evaluator.Evaluate(tt.formula, tt.data)
            }
        })
    }
}

// ✅ BEST PRACTICE: Test with timeout
func TestSlowOperation(t *testing.T) {
    if testing.Short() {
        t.Skip("skipping slow test")
    }

    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    result, err := slowOperation(ctx)
    require.NoError(t, err)
    assert.NotNil(t, result)
}

// ✅ BEST PRACTICE: Race condition detection
func TestConcurrentAccess(t *testing.T) {
    cache := NewCache()

    var wg sync.WaitGroup
    for i := 0; i < 100; i++ {
        wg.Add(1)
        go func(val int) {
            defer wg.Done()
            cache.Set(fmt.Sprintf("key-%d", val), val)
            cache.Get(fmt.Sprintf("key-%d", val))
        }(i)
    }

    wg.Wait()

    // Run with: go test -race ./...
}

// ✅ BEST PRACTICE: Error case testing
func TestWorkflowService_HandleErrors(t *testing.T) {
    mockRepo := new(MockRepository)
    service := NewWorkflowService(mockRepo)

    tests := []struct {
        name        string
        setupMock   func()
        expectedErr string
    }{
        {
            name: "repository error",
            setupMock: func() {
                mockRepo.On("GetWorkflow", mock.Anything, mock.Anything, mock.Anything).
                    Return(nil, errors.New("database error"))
            },
            expectedErr: "database error",
        },
        {
            name: "not found error",
            setupMock: func() {
                mockRepo.On("GetWorkflow", mock.Anything, mock.Anything, mock.Anything).
                    Return(nil, ErrNotFound)
            },
            expectedErr: "not found",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            mockRepo.ExpectedCalls = nil // Reset mock
            tt.setupMock()

            _, err := service.Execute(context.Background(), "tenant", "wf")

            require.Error(t, err)
            assert.Contains(t, err.Error(), tt.expectedErr)
        })
    }
}

// ✅ BEST PRACTICE: Test coverage helper
func TestMain(m *testing.M) {
    // Setup
    setupTestEnvironment()

    // Run tests
    code := m.Run()

    // Cleanup
    teardownTestEnvironment()

    os.Exit(code)
}
```

**Go testing commands**:
```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Run with race detector
go test -race ./...

# Run only short tests
go test -short ./...

# Run specific test
go test -run TestValidateWorkflow ./internal/workflow

# Verbose output
go test -v ./...

# Benchmark tests
go test -bench=. ./...

# Benchmark with memory stats
go test -bench=. -benchmem ./...
```

**Go testing checklist**:
- [ ] All functions have unit tests
- [ ] Table-driven tests for multiple cases
- [ ] Mocks for external dependencies
- [ ] Integration tests for database/API
- [ ] Benchmark tests for performance-critical code
- [ ] Tests run with -race flag
- [ ] 80%+ code coverage on business logic
- [ ] Error cases tested
- [ ] Edge cases tested (nil, empty, boundary)
- [ ] Tests use t.Helper() for test utilities
- [ ] Tests are fast (< 100ms for unit tests)
- [ ] Integration tests skip with -short flag

### gorax-react-testing-expert

**When to use**: Writing React tests, component testing, hook testing, integration tests with React Testing Library, mocking API calls

**Key expertise**:
- Vitest and React Testing Library
- Component behavior testing
- Custom hook testing
- User interaction simulation
- API mocking with MSW
- Accessibility testing
- Test organization and patterns

**React testing patterns**:
```typescript
// ✅ BEST PRACTICE: Component testing with React Testing Library
import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import WorkflowList from './WorkflowList'

describe('WorkflowList', () => {
  const mockWorkflows = [
    { id: '1', name: 'Workflow 1', status: 'active' },
    { id: '2', name: 'Workflow 2', status: 'draft' },
  ]

  beforeEach(() => {
    vi.clearAllMocks()
  })

  describe('rendering', () => {
    it('should render all workflows', () => {
      render(<WorkflowList workflows={mockWorkflows} />)

      expect(screen.getByText('Workflow 1')).toBeInTheDocument()
      expect(screen.getByText('Workflow 2')).toBeInTheDocument()
    })

    it('should show empty state when no workflows', () => {
      render(<WorkflowList workflows={[]} />)

      expect(screen.getByText(/no workflows found/i)).toBeInTheDocument()
    })

    it('should show loading spinner', () => {
      render(<WorkflowList workflows={[]} loading />)

      expect(screen.getByRole('progressbar')).toBeInTheDocument()
      expect(screen.queryByText(/no workflows/i)).not.toBeInTheDocument()
    })

    it('should display error message', () => {
      const error = 'Failed to load workflows'
      render(<WorkflowList workflows={[]} error={error} />)

      expect(screen.getByText(error)).toBeInTheDocument()
      expect(screen.getByRole('alert')).toBeInTheDocument()
    })
  })

  describe('user interactions', () => {
    it('should handle workflow selection', async () => {
      const user = userEvent.setup()
      const onSelect = vi.fn()

      render(<WorkflowList workflows={mockWorkflows} onSelect={onSelect} />)

      await user.click(screen.getByText('Workflow 1'))

      expect(onSelect).toHaveBeenCalledWith(mockWorkflows[0])
      expect(onSelect).toHaveBeenCalledTimes(1)
    })

    it('should handle delete with confirmation', async () => {
      const user = userEvent.setup()
      const onDelete = vi.fn()

      render(<WorkflowList workflows={mockWorkflows} onDelete={onDelete} />)

      // Click delete button
      await user.click(screen.getAllByRole('button', { name: /delete/i })[0])

      // Confirm dialog appears
      expect(screen.getByText(/are you sure/i)).toBeInTheDocument()

      // Click confirm
      await user.click(screen.getByRole('button', { name: /confirm/i }))

      expect(onDelete).toHaveBeenCalledWith('1')
    })

    it('should filter workflows by search', async () => {
      const user = userEvent.setup()

      render(<WorkflowList workflows={mockWorkflows} />)

      const searchInput = screen.getByPlaceholderText(/search/i)
      await user.type(searchInput, 'Workflow 1')

      expect(screen.getByText('Workflow 1')).toBeInTheDocument()
      expect(screen.queryByText('Workflow 2')).not.toBeInTheDocument()
    })
  })

  describe('accessibility', () => {
    it('should have accessible list', () => {
      render(<WorkflowList workflows={mockWorkflows} />)

      expect(screen.getByRole('list')).toBeInTheDocument()
      expect(screen.getAllByRole('listitem')).toHaveLength(2)
    })

    it('should have accessible buttons', () => {
      render(<WorkflowList workflows={mockWorkflows} />)

      const buttons = screen.getAllByRole('button')
      buttons.forEach(button => {
        expect(button).toHaveAccessibleName()
      })
    })
  })
})

// ✅ BEST PRACTICE: Custom hook testing
import { renderHook, waitFor } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { useWorkflow } from './useWorkflow'
import * as api from '@/api'

describe('useWorkflow', () => {
  const createWrapper = () => {
    const queryClient = new QueryClient({
      defaultOptions: {
        queries: { retry: false, cacheTime: 0 },
      },
    })

    return ({ children }: { children: React.ReactNode }) => (
      <QueryClientProvider client={queryClient}>
        {children}
      </QueryClientProvider>
    )
  }

  it('should fetch workflow data', async () => {
    const mockWorkflow = { id: 'wf-123', name: 'Test' }
    vi.spyOn(api.workflows, 'get').mockResolvedValue(mockWorkflow)

    const { result } = renderHook(() => useWorkflow('wf-123'), {
      wrapper: createWrapper(),
    })

    expect(result.current.isLoading).toBe(true)

    await waitFor(() => {
      expect(result.current.isSuccess).toBe(true)
    })

    expect(result.current.data).toEqual(mockWorkflow)
    expect(api.workflows.get).toHaveBeenCalledWith('wf-123')
  })

  it('should handle fetch errors', async () => {
    const error = new Error('Not found')
    vi.spyOn(api.workflows, 'get').mockRejectedValue(error)

    const { result } = renderHook(() => useWorkflow('invalid'), {
      wrapper: createWrapper(),
    })

    await waitFor(() => {
      expect(result.current.isError).toBe(true)
    })

    expect(result.current.error).toEqual(error)
  })

  it('should refetch on demand', async () => {
    vi.spyOn(api.workflows, 'get').mockResolvedValue({ id: 'wf-123', name: 'Test' })

    const { result } = renderHook(() => useWorkflow('wf-123'), {
      wrapper: createWrapper(),
    })

    await waitFor(() => expect(result.current.isSuccess).toBe(true))

    // Trigger refetch
    result.current.refetch()

    await waitFor(() => {
      expect(api.workflows.get).toHaveBeenCalledTimes(2)
    })
  })
})

// ✅ BEST PRACTICE: MSW for API mocking
import { setupServer } from 'msw/node'
import { http, HttpResponse } from 'msw'

const handlers = [
  http.get('/api/workflows/:id', ({ params }) => {
    const { id } = params
    return HttpResponse.json({
      id,
      name: 'Test Workflow',
      nodes: [],
      edges: [],
    })
  }),

  http.post('/api/workflows', async ({ request }) => {
    const body = await request.json()
    return HttpResponse.json({
      id: 'wf-new',
      ...body,
    }, { status: 201 })
  }),

  http.delete('/api/workflows/:id', () => {
    return new HttpResponse(null, { status: 204 })
  }),
]

const server = setupServer(...handlers)

beforeAll(() => server.listen())
afterEach(() => server.resetHandlers())
afterAll(() => server.close())

describe('WorkflowEditor integration', () => {
  it('should load workflow from API', async () => {
    render(<WorkflowEditor id="wf-123" />)

    await waitFor(() => {
      expect(screen.getByText('Test Workflow')).toBeInTheDocument()
    })
  })

  it('should handle create workflow', async () => {
    const user = userEvent.setup()

    render(<WorkflowEditor />)

    await user.type(screen.getByLabelText(/name/i), 'New Workflow')
    await user.click(screen.getByRole('button', { name: /create/i }))

    await waitFor(() => {
      expect(screen.getByText(/workflow created/i)).toBeInTheDocument()
    })
  })

  it('should handle API errors', async () => {
    server.use(
      http.get('/api/workflows/:id', () => {
        return HttpResponse.json(
          { error: 'Not found' },
          { status: 404 }
        )
      })
    )

    render(<WorkflowEditor id="invalid" />)

    await waitFor(() => {
      expect(screen.getByText(/not found/i)).toBeInTheDocument()
    })
  })
})

// ✅ BEST PRACTICE: Testing with providers
import { MemoryRouter } from 'react-router-dom'
import { AuthProvider } from '@/contexts/AuthContext'

function renderWithProviders(
  ui: React.ReactElement,
  {
    initialRoute = '/',
    user = null,
    ...renderOptions
  } = {}
) {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  })

  function Wrapper({ children }: { children: React.ReactNode }) {
    return (
      <QueryClientProvider client={queryClient}>
        <AuthProvider initialUser={user}>
          <MemoryRouter initialEntries={[initialRoute]}>
            {children}
          </MemoryRouter>
        </AuthProvider>
      </QueryClientProvider>
    )
  }

  return render(ui, { wrapper: Wrapper, ...renderOptions })
}

// Usage
it('should show authenticated content', () => {
  const user = { id: '1', name: 'Test User' }

  renderWithProviders(<Dashboard />, { user })

  expect(screen.getByText(/welcome, test user/i)).toBeInTheDocument()
})

// ✅ BEST PRACTICE: Snapshot testing (use sparingly)
it('should match snapshot', () => {
  const { container } = render(<WorkflowCard workflow={mockWorkflow} />)
  expect(container.firstChild).toMatchSnapshot()
})

// Update snapshots with: npm test -- -u
```

**React testing commands**:
```bash
# Run all tests
npm test

# Run in watch mode
npm test -- --watch

# Run with coverage
npm test -- --coverage

# Update snapshots
npm test -- -u

# Run specific test file
npm test WorkflowList

# Run tests matching pattern
npm test -- --grep "user interactions"

# UI mode (interactive)
npm test -- --ui
```

**React testing checklist**:
- [ ] All components have tests
- [ ] User interactions tested
- [ ] Loading states tested
- [ ] Error states tested
- [ ] Accessibility checked (roles, labels)
- [ ] Custom hooks tested
- [ ] API calls mocked with MSW
- [ ] Provider contexts mocked
- [ ] Edge cases tested (empty, error, loading)
- [ ] Use screen.getByRole over getByTestId
- [ ] Async operations use waitFor
- [ ] No hardcoded timeouts (use waitFor)
- [ ] Tests are isolated and independent

### Using Expert Agents

Launch an expert agent for complex domain-specific tasks:

```
Task tool with:
- subagent_type: "gorax-workflow-expert"
- prompt: "Implement parallel node execution with proper error handling..."
```

The agent will have context about patterns, file locations, and domain concepts specific to that area of the codebase.
