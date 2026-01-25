# Gorax Architecture

> Comprehensive system architecture documentation for the Gorax workflow automation platform.

---

## Table of Contents

1. [System Overview](#system-overview)
2. [Architecture Principles](#architecture-principles)
3. [Component Architecture](#component-architecture)
4. [Data Flow](#data-flow)
5. [Technology Stack](#technology-stack)
6. [Design Patterns](#design-patterns)
7. [Deployment Architecture](#deployment-architecture)
8. [Security Architecture](#security-architecture)
9. [Scalability & Performance](#scalability--performance)

---

## System Overview

Gorax is a workflow automation platform built on clean architecture principles with a multi-tenant, event-driven design. The system enables users to create, manage, and execute complex workflows through a visual editor.

### High-Level Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                         Frontend (React)                     │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐      │
│  │ Workflow     │  │  Dashboard   │  │  Admin       │      │
│  │ Editor       │  │              │  │  Panel       │      │
│  └──────────────┘  └──────────────┘  └──────────────┘      │
└─────────────────────────────────────────────────────────────┘
                              │
                        HTTP/WebSocket
                              │
┌─────────────────────────────────────────────────────────────┐
│                      API Layer (Go)                          │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐      │
│  │ REST API     │  │  GraphQL     │  │  WebSocket   │      │
│  │ Handlers     │  │  Resolvers   │  │  Hub         │      │
│  └──────────────┘  └──────────────┘  └──────────────┘      │
└─────────────────────────────────────────────────────────────┘
                              │
┌─────────────────────────────────────────────────────────────┐
│                     Service Layer (Go)                       │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐      │
│  │ Workflow     │  │  Executor    │  │  Credential  │      │
│  │ Service      │  │  Service     │  │  Service     │      │
│  └──────────────┘  └──────────────┘  └──────────────┘      │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐      │
│  │ Webhook      │  │  Analytics   │  │  Marketplace │      │
│  │ Service      │  │  Service     │  │  Service     │      │
│  └──────────────┘  └──────────────┘  └──────────────┘      │
└─────────────────────────────────────────────────────────────┘
                              │
┌─────────────────────────────────────────────────────────────┐
│                   Repository Layer (Go)                      │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐      │
│  │ PostgreSQL   │  │  Redis       │  │  S3          │      │
│  │ Repositories │  │  Cache       │  │  Storage     │      │
│  └──────────────┘  └──────────────┘  └──────────────┘      │
└─────────────────────────────────────────────────────────────┘
                              │
┌─────────────────────────────────────────────────────────────┐
│                    Infrastructure                            │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐      │
│  │ PostgreSQL   │  │  Redis       │  │  S3/Minio    │      │
│  │ Database     │  │  Cache       │  │  Object      │      │
│  └──────────────┘  └──────────────┘  └──────────────┘      │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐      │
│  │ SQS/Queue    │  │  Prometheus  │  │  Grafana     │      │
│  │ Service      │  │  Metrics     │  │  Dashboard   │      │
│  └──────────────┘  └──────────────┘  └──────────────┘      │
└─────────────────────────────────────────────────────────────┘
```

### Key Characteristics

- **Multi-tenant**: Tenant isolation at database and application layers
- **Event-driven**: Asynchronous workflow execution via message queues
- **Microservices-ready**: Modular architecture supports service extraction
- **Cloud-native**: Designed for containerized deployment (Docker, Kubernetes)
- **Real-time**: WebSocket support for live collaboration and updates
- **Scalable**: Horizontal scaling for API, workers, and database

---

## Architecture Principles

Gorax follows **Clean Architecture** principles (Robert C. Martin) with clear separation of concerns:

### 1. Dependency Rule

Dependencies point **inward**:

```
External → Adapters → Use Cases → Domain
```

- **Domain** (inner): Pure business logic, no external dependencies
- **Use Cases** (service layer): Application-specific business rules
- **Adapters** (repositories, API): Interface with external systems
- **External** (frameworks): Databases, web frameworks, UI

### 2. SOLID Principles

| Principle | Implementation |
|-----------|----------------|
| **Single Responsibility** | Each package/type has one reason to change |
| **Open/Closed** | Extend via interfaces, not modification |
| **Liskov Substitution** | Implementations are interchangeable |
| **Interface Segregation** | Small, focused interfaces |
| **Dependency Inversion** | Depend on abstractions (interfaces) |

### 3. Domain-Driven Design (DDD)

- **Bounded Contexts**: Workflow, Credential, Webhook, Analytics, Marketplace
- **Aggregates**: Workflow (nodes, edges), Credential (secrets), Webhook (filters, events)
- **Value Objects**: NodeID, TenantID, WorkflowID
- **Domain Events**: WorkflowExecuted, WebhookTriggered, CredentialRotated

---

## Component Architecture

### Backend (Go)

```
gorax/
├── cmd/
│   ├── api/              # HTTP API server entry point
│   ├── worker/           # Background job processor
│   └── migrate/          # Database migration tool
│
├── internal/             # Private application code
│   ├── api/              # HTTP handlers and middleware
│   │   ├── handlers/     # REST API handlers
│   │   ├── middleware/   # Auth, CORS, rate limiting
│   │   └── app.go        # HTTP server setup
│   │
│   ├── graphql/          # GraphQL API
│   │   ├── schema.graphql
│   │   ├── resolvers.go
│   │   └── generated/
│   │
│   ├── collaboration/    # Real-time collaboration
│   │   ├── hub.go        # WebSocket hub
│   │   ├── service.go    # Collaboration logic
│   │   └── model.go
│   │
│   ├── workflow/         # Workflow management
│   │   ├── service.go    # Business logic
│   │   ├── repository.go # Data access
│   │   ├── bulk_service.go
│   │   └── model.go
│   │
│   ├── executor/         # Workflow execution engine
│   │   ├── executor.go   # Core execution logic
│   │   ├── visitor.go    # Node traversal (Visitor pattern)
│   │   ├── context.go    # Execution context
│   │   ├── actions/      # Action implementations
│   │   └── expression/   # CEL expression evaluator
│   │
│   ├── credential/       # Credential management
│   │   ├── service.go
│   │   ├── encryption.go # AES-256-GCM encryption
│   │   ├── injector.go   # Template injection
│   │   ├── masker.go     # Secret masking
│   │   └── repository.go
│   │
│   ├── webhook/          # Webhook handling
│   │   ├── service.go
│   │   ├── handler.go    # HTTP webhook receiver
│   │   ├── filter.go     # JSONPath filtering
│   │   ├── replay.go     # Event replay
│   │   └── repository.go
│   │
│   ├── analytics/        # Analytics and metrics
│   │   ├── service.go
│   │   ├── repository.go
│   │   └── model.go
│   │
│   ├── marketplace/      # Template marketplace
│   │   ├── service.go
│   │   ├── repository.go
│   │   └── model.go
│   │
│   ├── worker/           # Background job processing
│   │   ├── worker.go     # Job worker
│   │   ├── queue.go      # SQS integration
│   │   └── registry.go   # Job type registry
│   │
│   ├── rbac/             # Role-Based Access Control
│   │   ├── enforcer.go   # Permission checks
│   │   ├── middleware.go # RBAC middleware
│   │   └── policy.go
│   │
│   ├── database/         # Database layer
│   │   ├── postgres.go   # Connection pool
│   │   ├── tenant_hooks.go # Tenant isolation
│   │   └── migrations/
│   │
│   ├── config/           # Configuration
│   │   ├── config.go
│   │   └── websocket.go
│   │
│   └── llm/              # LLM integrations
│       ├── providers/    # OpenAI, Anthropic, Bedrock
│       ├── client.go
│       └── model.go
│
├── pkg/                  # Public libraries (if extracted)
│
└── migrations/           # SQL migrations
    ├── 001_initial_schema.sql
    ├── 002_add_webhooks.sql
    └── ...
```

### Frontend (React + TypeScript)

```
web/
├── src/
│   ├── components/       # Reusable UI components
│   │   ├── canvas/       # ReactFlow workflow editor
│   │   ├── nodes/        # Custom node types
│   │   ├── collaboration/ # Real-time collaboration UI
│   │   ├── credentials/  # Credential management
│   │   ├── webhooks/     # Webhook configuration
│   │   └── ...
│   │
│   ├── pages/            # Route-level components
│   │   ├── WorkflowEditor.tsx
│   │   ├── WorkflowList.tsx
│   │   ├── Dashboard.tsx
│   │   ├── Analytics.tsx
│   │   ├── Marketplace.tsx
│   │   └── ...
│   │
│   ├── hooks/            # Custom React hooks
│   │   ├── useWorkflows.ts
│   │   ├── useCredentials.ts
│   │   ├── useCollaboration.ts
│   │   ├── useWebSocket.ts
│   │   └── ...
│   │
│   ├── stores/           # Zustand state management
│   │   ├── workflowStore.ts
│   │   ├── credentialStore.ts
│   │   └── ...
│   │
│   ├── api/              # API clients
│   │   ├── client.ts     # Base HTTP client
│   │   ├── workflows.ts
│   │   ├── credentials.ts
│   │   └── ...
│   │
│   ├── types/            # TypeScript types
│   │   ├── workflow.ts
│   │   ├── credential.ts
│   │   ├── collaboration.ts
│   │   └── ...
│   │
│   ├── lib/              # Utilities
│   │   ├── websocket.ts
│   │   ├── validation.ts
│   │   └── ...
│   │
│   └── main.tsx          # Application entry point
│
├── vite.config.ts        # Vite build configuration
├── tsconfig.json         # TypeScript configuration
└── package.json          # Dependencies
```

---

## Data Flow

### 1. Workflow Execution Flow

```
User Action → API Handler → Workflow Service → Queue (SQS)
                                                    ↓
Worker Pool ← Queue Poller ← Queue Message
    ↓
Executor Service
    ↓
┌─────────────────────────────────────────┐
│  Execution Context                      │
│  ┌───────────────────────────────────┐  │
│  │ 1. Load Workflow Graph            │  │
│  │ 2. Resolve Credentials            │  │
│  │ 3. Evaluate Trigger Conditions    │  │
│  │ 4. Traverse Nodes (Visitor)       │  │
│  │    ├─ Action Nodes                │  │
│  │    ├─ Conditional Nodes           │  │
│  │    ├─ Loop Nodes                  │  │
│  │    └─ Parallel Nodes              │  │
│  │ 5. Execute Actions                │  │
│  │ 6. Store Execution Results        │  │
│  └───────────────────────────────────┘  │
└─────────────────────────────────────────┘
    ↓
Result Storage (PostgreSQL)
    ↓
Real-time Update (WebSocket) → Frontend
```

### 2. Webhook Event Flow

```
External System → Webhook Endpoint (/webhooks/:path)
                        ↓
                 Webhook Handler
                        ↓
    ┌───────────────────────────────────┐
    │ 1. Verify Signature (HMAC-SHA256) │
    │ 2. Log Event                      │
    │ 3. Apply Filters (JSONPath)       │
    │ 4. Enqueue Workflow Execution     │
    └───────────────────────────────────┘
                        ↓
                 Queue (SQS)
                        ↓
                 Worker (Executor)
```

### 3. Real-time Collaboration Flow

```
User A (Browser) → WebSocket Connection → Collaboration Hub
                                                ↓
                                    ┌─────────────────────┐
                                    │ Hub maintains:      │
                                    │ - Active connections│
                                    │ - User presence     │
                                    │ - Node locks        │
                                    │ - Cursor positions  │
                                    └─────────────────────┘
                                                ↓
                           Broadcast to User B, User C, ...
                                                ↓
                                         User B (Browser)
                                         User C (Browser)
```

### 4. Credential Injection Flow

```
Workflow Definition: "Bearer {{credentials.api_token}}"
                        ↓
                 Credential Service
                        ↓
    ┌───────────────────────────────────┐
    │ 1. Parse template references      │
    │ 2. Load encrypted credentials     │
    │ 3. Decrypt with AES-256-GCM       │
    │ 4. Inject into template           │
    │ 5. Mask in logs                   │
    └───────────────────────────────────┘
                        ↓
            Action Execution
                        ↓
        Executed Value: "Bearer sk_live_abc123..."
        Logged Value:   "Bearer [REDACTED]"
```

---

## Technology Stack

### Backend

| Component | Technology | Purpose |
|-----------|------------|---------|
| **Language** | Go 1.21+ | High-performance, concurrent backend |
| **Web Framework** | Echo v4 | HTTP router and middleware |
| **GraphQL** | gqlgen | Type-safe GraphQL server |
| **Database** | PostgreSQL 15+ | Primary data store |
| **Cache** | Redis 6+ | Session storage, rate limiting |
| **Queue** | AWS SQS | Asynchronous job processing |
| **ORM** | sql.Database (stdlib) | Direct SQL with migrations |
| **Encryption** | AES-256-GCM | Credential encryption |
| **Validation** | go-playground/validator | Request validation |
| **Testing** | testify, httptest | Unit and integration tests |
| **Migrations** | golang-migrate | Database schema versioning |
| **Observability** | OpenTelemetry | Tracing, metrics, logs |

### Frontend

| Component | Technology | Purpose |
|-----------|------------|---------|
| **Language** | TypeScript 5+ | Type-safe JavaScript |
| **Framework** | React 18 | UI library |
| **Build Tool** | Vite | Fast dev server and bundler |
| **State Management** | Zustand | Lightweight global state |
| **Server State** | TanStack Query | Server state caching and sync |
| **Workflow Editor** | ReactFlow | Node-based graph editor |
| **Forms** | React Hook Form | Form state management |
| **Validation** | Zod | Schema validation |
| **HTTP Client** | Fetch API | HTTP requests |
| **WebSocket** | Native WebSocket | Real-time communication |
| **Testing** | Vitest, Testing Library | Unit and component tests |
| **Styling** | Tailwind CSS | Utility-first CSS |

### Infrastructure

| Component | Technology | Purpose |
|-----------|------------|---------|
| **Containerization** | Docker | Application packaging |
| **Orchestration** | Kubernetes | Container orchestration |
| **Load Balancer** | NGINX / AWS ALB | HTTP load balancing |
| **Object Storage** | S3 / MinIO | File and export storage |
| **Monitoring** | Prometheus | Metrics collection |
| **Visualization** | Grafana | Metrics dashboards |
| **Logging** | Loki | Log aggregation |
| **Tracing** | Jaeger | Distributed tracing |
| **Error Tracking** | Sentry | Error reporting |
| **CI/CD** | GitHub Actions | Automated testing and deployment |

---

## Design Patterns

### Creational Patterns

#### 1. Factory Pattern
**Used in**: Action creation, provider instantiation

```go
// executor/actions/factory.go
func CreateAction(actionType string) (Action, error) {
    switch actionType {
    case "http_request":
        return &HTTPRequestAction{}, nil
    case "send_email":
        return &SendEmailAction{}, nil
    case "chat_completion":
        return &ChatCompletionAction{}, nil
    default:
        return nil, fmt.Errorf("unknown action type: %s", actionType)
    }
}
```

#### 2. Builder Pattern
**Used in**: Workflow definition, query builders

```go
// workflow/builder.go
type WorkflowBuilder struct {
    workflow *Workflow
}

func NewWorkflowBuilder() *WorkflowBuilder {
    return &WorkflowBuilder{
        workflow: &Workflow{Nodes: []Node{}, Edges: []Edge{}},
    }
}

func (b *WorkflowBuilder) WithName(name string) *WorkflowBuilder {
    b.workflow.Name = name
    return b
}

func (b *WorkflowBuilder) AddNode(node Node) *WorkflowBuilder {
    b.workflow.Nodes = append(b.workflow.Nodes, node)
    return b
}

func (b *WorkflowBuilder) Build() *Workflow {
    return b.workflow
}
```

### Structural Patterns

#### 3. Repository Pattern
**Used in**: All data access

```go
// workflow/repository.go
type Repository interface {
    GetByID(ctx context.Context, id string) (*Workflow, error)
    Create(ctx context.Context, workflow *Workflow) error
    Update(ctx context.Context, workflow *Workflow) error
    Delete(ctx context.Context, id string) error
    List(ctx context.Context, tenantID string, opts ListOptions) ([]*Workflow, error)
}

type PostgresRepository struct {
    db *sql.DB
}

func (r *PostgresRepository) GetByID(ctx context.Context, id string) (*Workflow, error) {
    // Implementation
}
```

#### 4. Adapter Pattern
**Used in**: LLM providers, queue implementations

```go
// llm/provider.go
type Provider interface {
    ChatCompletion(ctx context.Context, req ChatRequest) (*ChatResponse, error)
    Embedding(ctx context.Context, req EmbeddingRequest) (*EmbeddingResponse, error)
}

// OpenAI adapter
type OpenAIProvider struct {
    client *openai.Client
}

func (p *OpenAIProvider) ChatCompletion(ctx context.Context, req ChatRequest) (*ChatResponse, error) {
    // Adapt our request format to OpenAI's format
    openAIReq := openai.ChatCompletionRequest{
        Model: req.Model,
        Messages: convertMessages(req.Messages),
    }
    resp, err := p.client.CreateChatCompletion(ctx, openAIReq)
    // Adapt OpenAI response to our format
    return convertResponse(resp), err
}
```

#### 5. Decorator Pattern
**Used in**: Middleware, action wrappers

```go
// api/middleware/logging.go
func LoggingMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()
        next.ServeHTTP(w, r)
        duration := time.Since(start)
        log.Info("request completed",
            "method", r.Method,
            "path", r.URL.Path,
            "duration", duration,
        )
    })
}
```

### Behavioral Patterns

#### 6. Visitor Pattern
**Used in**: Workflow node traversal

```go
// executor/visitor.go
type NodeVisitor interface {
    VisitAction(node *ActionNode) error
    VisitCondition(node *ConditionNode) error
    VisitLoop(node *LoopNode) error
    VisitParallel(node *ParallelNode) error
}

type ExecutorVisitor struct {
    ctx *ExecutionContext
}

func (v *ExecutorVisitor) VisitAction(node *ActionNode) error {
    action, err := CreateAction(node.ActionType)
    if err != nil {
        return err
    }
    result, err := action.Execute(v.ctx, node.Config)
    v.ctx.SetResult(node.ID, result)
    return err
}
```

#### 7. Strategy Pattern
**Used in**: Authentication strategies, execution strategies

```go
// webhook/auth.go
type AuthStrategy interface {
    Verify(r *http.Request, webhook *Webhook) error
}

type SignatureAuthStrategy struct{}

func (s *SignatureAuthStrategy) Verify(r *http.Request, webhook *Webhook) error {
    signature := r.Header.Get("X-Webhook-Signature")
    body, _ := io.ReadAll(r.Body)
    expected := computeHMAC(webhook.Secret, body)
    if !hmac.Equal([]byte(signature), []byte(expected)) {
        return ErrInvalidSignature
    }
    return nil
}
```

#### 8. Observer Pattern
**Used in**: Event notifications, real-time updates

```go
// collaboration/hub.go
type Hub struct {
    clients   map[string]*Client
    broadcast chan Message
}

func (h *Hub) Broadcast(msg Message) {
    h.broadcast <- msg
}

func (h *Hub) Run() {
    for {
        select {
        case msg := <-h.broadcast:
            for _, client := range h.clients {
                client.Send(msg)
            }
        }
    }
}
```

### React Patterns

#### 9. Custom Hooks
**Used in**: Reusable stateful logic

```typescript
// hooks/useWorkflows.ts
export function useWorkflows() {
  const { data, isLoading, error, refetch } = useQuery({
    queryKey: ['workflows'],
    queryFn: fetchWorkflows,
  })

  const createMutation = useMutation({
    mutationFn: createWorkflow,
    onSuccess: () => queryClient.invalidateQueries(['workflows']),
  })

  return {
    workflows: data ?? [],
    loading: isLoading,
    error,
    refetch,
    createWorkflow: createMutation.mutate,
  }
}
```

#### 10. Compound Components
**Used in**: FilterBuilder, NodeConfiguration

```typescript
// components/webhooks/FilterBuilder.tsx
export function FilterBuilder({ webhookId }: Props) {
  return (
    <FilterContext.Provider value={{ webhookId }}>
      <FilterList />
      <AddFilterButton />
    </FilterContext.Provider>
  )
}
```

---

## Deployment Architecture

### Development Environment

```
┌─────────────────────────────────────────┐
│ Developer Machine                       │
│ ┌─────────────┐  ┌─────────────┐       │
│ │   Go API    │  │  React App  │       │
│ │ (localhost) │  │  (Vite Dev) │       │
│ │   :8080     │  │   :5173     │       │
│ └─────────────┘  └─────────────┘       │
│         │                 │             │
│         └────────┬────────┘             │
│                  ↓                      │
│ ┌────────────────────────────────────┐ │
│ │ Local PostgreSQL (Docker)          │ │
│ │ Local Redis (Docker)               │ │
│ └────────────────────────────────────┘ │
└─────────────────────────────────────────┘
```

### Production (Kubernetes)

```
┌────────────────────────── Kubernetes Cluster ─────────────────────────────┐
│                                                                            │
│  ┌──────────────────── Ingress (NGINX) ─────────────────────┐            │
│  │                         SSL/TLS                           │            │
│  │  gorax.example.com → api.gorax.example.com               │            │
│  └────────────────────────┬──────────────────────────────────┘            │
│                           │                                               │
│  ┌────────────────────────┴──────────────────────────────┐               │
│  │                                                        │               │
│  │  ┌──────────────┐             ┌──────────────┐       │               │
│  │  │  Frontend    │             │   API        │       │               │
│  │  │  (StaticFS)  │             │  (3 pods)    │       │               │
│  │  │  (1 pod)     │             │  HPA: 3-10   │       │               │
│  │  └──────────────┘             └──────┬───────┘       │               │
│  │                                       │               │               │
│  │                    ┌──────────────────┴──────────────────────┐       │
│  │                    │                                          │       │
│  │        ┌───────────▼──────────┐              ┌───────────────▼────┐  │
│  │        │  Worker Pool         │              │  Collaboration     │  │
│  │        │  (5 pods)            │              │  Hub (2 pods)      │  │
│  │        │  HPA: 5-20           │              │  Stateful          │  │
│  │        └───────────┬──────────┘              └────────────────────┘  │
│  │                    │                                                  │
│  └────────────────────┼──────────────────────────────────────────────────┘
│                       │                                                  │
│  ┌────────────────────▼──────────────────────────────┐                  │
│  │            Data Layer (StatefulSets)              │                  │
│  │  ┌──────────────┐  ┌──────────────┐  ┌─────────┐ │                  │
│  │  │ PostgreSQL   │  │    Redis     │  │  MinIO  │ │                  │
│  │  │ (Primary +   │  │  (Sentinel)  │  │  (S3)   │ │                  │
│  │  │  2 Replicas) │  │  (3 nodes)   │  │         │ │                  │
│  │  └──────────────┘  └──────────────┘  └─────────┘ │                  │
│  └───────────────────────────────────────────────────┘                  │
│                                                                           │
│  ┌───────────────── Monitoring & Observability ─────────────────┐       │
│  │  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐       │       │
│  │  │ Prometheus   │  │   Grafana    │  │   Jaeger     │       │       │
│  │  └──────────────┘  └──────────────┘  └──────────────┘       │       │
│  └──────────────────────────────────────────────────────────────┘       │
└────────────────────────────────────────────────────────────────────────────┘
                                    │
                    ┌───────────────┴───────────────┐
                    │                               │
            ┌───────▼────────┐          ┌──────────▼─────────┐
            │   AWS SQS      │          │  External Services │
            │   (Queue)      │          │  - Auth (Kratos)   │
            │                │          │  - Sentry          │
            └────────────────┘          └────────────────────┘
```

### High Availability (HA) Configuration

| Component | HA Strategy | Min Instances |
|-----------|-------------|---------------|
| **API** | Stateless pods, HPA | 3 |
| **Workers** | Stateless pods, HPA | 5 |
| **PostgreSQL** | Primary-replica with auto-failover | 1 primary + 2 replicas |
| **Redis** | Sentinel or Cluster mode | 3 nodes |
| **Ingress** | Multi-zone load balancer | 2 |
| **Collaboration Hub** | StatefulSet with sticky sessions | 2 |

---

## Security Architecture

### Defense in Depth

```
┌─────────────────────────────────────────────────────────────┐
│  Layer 1: Network Security                                  │
│  - WAF (Web Application Firewall)                           │
│  - DDoS protection                                          │
│  - TLS 1.3 encryption                                       │
└─────────────────────────────────────────────────────────────┘
                          ↓
┌─────────────────────────────────────────────────────────────┐
│  Layer 2: Authentication & Authorization                    │
│  - JWT tokens (httpOnly cookies)                            │
│  - OAuth 2.0 / OIDC (Ory Kratos)                            │
│  - RBAC (Role-Based Access Control)                         │
│  - Tenant isolation                                         │
└─────────────────────────────────────────────────────────────┘
                          ↓
┌─────────────────────────────────────────────────────────────┐
│  Layer 3: Application Security                              │
│  - Input validation (all requests)                          │
│  - SQL injection prevention (parameterized queries)         │
│  - XSS prevention (output encoding)                         │
│  - CSRF tokens                                              │
│  - Rate limiting                                            │
│  - Security headers (HSTS, CSP, X-Frame-Options)            │
└─────────────────────────────────────────────────────────────┘
                          ↓
┌─────────────────────────────────────────────────────────────┐
│  Layer 4: Data Security                                     │
│  - Encryption at rest (AES-256-GCM)                         │
│  - Encryption in transit (TLS 1.3)                          │
│  - Credential envelope encryption (DEK/KEK)                 │
│  - Secret masking in logs                                   │
│  - Secure key management (KMS)                              │
└─────────────────────────────────────────────────────────────┘
                          ↓
┌─────────────────────────────────────────────────────────────┐
│  Layer 5: Infrastructure Security                           │
│  - Container image scanning                                 │
│  - Pod security policies                                    │
│  - Network policies (namespace isolation)                   │
│  - Secrets management (Kubernetes Secrets)                  │
│  - Audit logging                                            │
└─────────────────────────────────────────────────────────────┘
```

### Credential Encryption Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                  Envelope Encryption                         │
│                                                              │
│  Plaintext Credential                                        │
│         ↓                                                    │
│  ┌─────────────────┐                                        │
│  │ Generate DEK    │ ← Data Encryption Key (random)         │
│  │ (AES-256-GCM)   │                                        │
│  └────────┬────────┘                                        │
│           ↓                                                  │
│  Encrypted Credential                                        │
│           ↓                                                  │
│  ┌─────────────────┐                                        │
│  │ Encrypt DEK     │ ← Key Encryption Key (KMS)             │
│  │ with KEK        │                                        │
│  └────────┬────────┘                                        │
│           ↓                                                  │
│  ┌─────────────────────────────────────────┐               │
│  │ Store in Database:                       │               │
│  │ - Encrypted credential                   │               │
│  │ - Encrypted DEK                          │               │
│  │ - Algorithm metadata                     │               │
│  └──────────────────────────────────────────┘               │
└─────────────────────────────────────────────────────────────┘
```

**Benefits**:
- DEK rotation without re-encrypting data
- KMS manages only KEKs (not individual credentials)
- Performance (symmetric encryption for data)
- Compliance (FIPS 140-2 validated KMS)

---

## Scalability & Performance

### Horizontal Scaling

| Component | Scaling Strategy | Bottleneck Metric |
|-----------|------------------|-------------------|
| **API Pods** | HPA on CPU (>70%) or RPS (>1000) | CPU, Memory |
| **Worker Pods** | HPA on queue depth (>100 msgs) | Queue backlog |
| **PostgreSQL** | Read replicas, connection pooling | Connections, IOPS |
| **Redis** | Cluster mode (sharding) | Memory, Connections |
| **Collaboration Hub** | Sticky sessions, multi-instance | WebSocket connections |

### Database Optimization

**Connection Pooling**:
```go
// database/postgres.go
db, err := sql.Open("postgres", dsn)
db.SetMaxOpenConns(25)
db.SetMaxIdleConns(5)
db.SetConnMaxLifetime(5 * time.Minute)
```

**Read Replicas**:
- Write queries → Primary
- Read queries → Replicas (round-robin)
- Execution results, analytics → Replicas

**Partitioning**:
- `executions` table: Partition by month (range partitioning)
- `webhook_events` table: Partition by month
- Archive old partitions to S3

**Indexing Strategy**:
```sql
-- Composite indexes for common queries
CREATE INDEX idx_workflows_tenant_status ON workflows(tenant_id, status);
CREATE INDEX idx_executions_workflow_created ON executions(workflow_id, created_at DESC);
CREATE INDEX idx_webhook_events_webhook_created ON webhook_events(webhook_id, created_at DESC);

-- Partial indexes for active records
CREATE INDEX idx_workflows_active ON workflows(tenant_id) WHERE deleted_at IS NULL;
```

### Caching Strategy

| Cache Layer | Technology | TTL | Use Case |
|-------------|------------|-----|----------|
| **HTTP Cache** | CDN (CloudFront) | 1 hour | Static assets |
| **Application Cache** | Redis | 5-15 min | Workflow definitions, user sessions |
| **Query Cache** | PostgreSQL | N/A | Materialized views |
| **Client Cache** | TanStack Query | 5 min | API responses (GET) |

**Cache Invalidation**:
- Write operations → Invalidate specific keys
- Workflow updates → Invalidate `workflow:{id}`
- Tenant updates → Invalidate `tenant:{id}:*`

### Performance Targets

| Metric | Target | P95 | P99 |
|--------|--------|-----|-----|
| **API Latency** | < 100ms | < 200ms | < 500ms |
| **Workflow Execution** | < 5s (simple) | < 30s | < 60s |
| **WebSocket Latency** | < 50ms | < 100ms | < 200ms |
| **Database Query** | < 50ms | < 100ms | < 200ms |
| **Queue Processing** | < 1s (per message) | < 5s | < 10s |

### Load Testing Results

**Test Scenario**: 1000 concurrent users, 10,000 workflows/hour

| Metric | Result |
|--------|--------|
| **Throughput** | 2.8 workflows/second |
| **API Requests** | 5000 req/sec |
| **Database Connections** | 120 / 200 max |
| **Queue Depth** | < 50 messages |
| **CPU Usage** | 65% (API), 80% (Workers) |
| **Memory Usage** | 2GB (API), 4GB (Workers) |
| **Error Rate** | 0.01% |

---

## Observability

### Metrics (Prometheus)

**Golden Signals**:
```prometheus
# Latency
http_request_duration_seconds{handler="/api/workflows", method="GET"}

# Traffic
http_requests_total{handler="/api/workflows", method="POST"}

# Errors
http_requests_total{handler="/api/workflows", status="500"}

# Saturation
go_goroutines
process_resident_memory_bytes
```

**Custom Metrics**:
```go
// Workflow execution metrics
workflowExecutionDuration.Observe(duration.Seconds())
workflowExecutionTotal.WithLabelValues(status).Inc()
workflowNodeExecutionTotal.WithLabelValues(nodeType, status).Inc()

// Queue metrics
queueDepth.Set(float64(depth))
queueProcessingDuration.Observe(duration.Seconds())
```

### Logging (Structured)

```go
// internal/logger/logger.go
log.Info("workflow execution started",
    "workflow_id", workflowID,
    "tenant_id", tenantID,
    "execution_id", executionID,
    "trigger", trigger,
)

log.Error("action execution failed",
    "action_id", actionID,
    "action_type", actionType,
    "error", err,
    "duration_ms", durationMs,
)
```

**Log Levels**:
- **DEBUG**: Detailed execution flow
- **INFO**: Normal operations (execution start/end)
- **WARN**: Recoverable errors (retries, deprecated APIs)
- **ERROR**: Failures requiring investigation

### Tracing (OpenTelemetry)

**Trace Context Propagation**:
```go
ctx, span := tracer.Start(ctx, "workflow.Execute")
defer span.End()

span.SetAttributes(
    attribute.String("workflow.id", workflowID),
    attribute.String("tenant.id", tenantID),
)
```

**Trace Sampling**:
- HEAD sampling: 100% for errors, 10% for success
- TAIL sampling: 100% for slow requests (> 1s)

### Alerting Rules

| Alert | Condition | Severity | Action |
|-------|-----------|----------|--------|
| **API Error Rate High** | > 1% for 5 min | Critical | Page on-call |
| **Database Connection Exhausted** | > 90% for 2 min | Critical | Page on-call |
| **Queue Depth Growing** | > 500 for 10 min | Warning | Investigate |
| **Worker CPU High** | > 90% for 5 min | Warning | Scale up |
| **Execution Failures** | > 5% for 10 min | Warning | Review logs |

---

## References

- [Clean Architecture by Robert C. Martin](https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html)
- [Domain-Driven Design by Eric Evans](https://www.domainlanguage.com/ddd/)
- [Twelve-Factor App](https://12factor.net/)
- [Go Project Layout](https://github.com/golang-standards/project-layout)
- [Effective Go](https://golang.org/doc/effective_go.html)
- [React Best Practices](https://react.dev/learn)
- [Kubernetes Best Practices](https://kubernetes.io/docs/concepts/)
- [OWASP Top 10](https://owasp.org/www-project-top-ten/)

---

## Contributing

See [CONTRIBUTING.md](../CONTRIBUTING.md) for development guidelines and [DEVELOPER_GUIDE.md](DEVELOPER_GUIDE.md) for detailed implementation patterns.
