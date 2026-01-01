# Gorax Developer Guide

> A comprehensive guide for developers contributing to the Gorax workflow automation platform.

---

## Table of Contents

1. [Getting Started](#getting-started)
2. [Project Architecture](#project-architecture)
3. [Development Workflow](#development-workflow)
4. [Key Concepts](#key-concepts)
5. [Adding New Features](#adding-new-features)
6. [Testing](#testing)
7. [Debugging](#debugging)
8. [Deployment](#deployment)

---

## Getting Started

### Prerequisites

Before you begin, ensure you have the following installed:

| Tool | Version | Purpose |
|------|---------|---------|
| **Go** | 1.24+ | Backend runtime |
| **Node.js** | 18+ | Frontend tooling |
| **PostgreSQL** | 14+ | Primary database |
| **Redis** | 6+ | Cache and sessions |
| **Docker** | Latest | Container runtime (optional) |
| **Git** | Latest | Version control |

**Optional Tools:**
- `golangci-lint` - Go linting ([install](https://golangci-lint.run/usage/install/))
- `air` - Go hot reload ([install](https://github.com/cosmtrek/air))
- `tmux` - Terminal multiplexer (for `make dev-start`)

### Clone and Setup

```bash
# Clone the repository
git clone https://github.com/stherrien/gorax.git
cd gorax

# Copy environment configuration
cp .env.example .env

# Edit .env with your settings (see below)
```

### Environment Configuration

Edit `.env` to configure your local environment:

```bash
# Application
APP_ENV=development
SERVER_ADDRESS=:8080

# Database
DB_HOST=localhost
DB_PORT=5433
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=gorax
DB_SSLMODE=disable

# Redis
REDIS_ADDRESS=localhost:6379
REDIS_PASSWORD=
REDIS_DB=0

# Observability (optional for development)
METRICS_ENABLED=true
METRICS_PORT=9090
TRACING_ENABLED=false
SENTRY_ENABLED=false

# Credentials (development - use KMS in production)
CREDENTIAL_USE_KMS=false
CREDENTIAL_MASTER_KEY=your_32_byte_base64_key_here  # Generate: openssl rand -base64 32

# CORS (adjust for your frontend port)
CORS_ALLOWED_ORIGINS=http://localhost:5173,http://localhost:3000
```

### Running Locally

#### Option 1: Simple Dev Environment (Recommended)

Start only the dependencies (PostgreSQL + Redis):

```bash
# Start postgres and redis in Docker
make dev-simple

# In another terminal, start the API server
make run-api-dev

# In a third terminal, start the frontend
cd web && npm install && npm run dev
```

Access the application:
- **Frontend:** http://localhost:5173
- **API:** http://localhost:8080
- **Metrics:** http://localhost:9090/metrics (if enabled)

#### Option 2: Tmux Split-Screen Development

Automatically start API + frontend in a split terminal:

```bash
# Start both API and frontend in tmux
make dev-start

# Tmux keyboard shortcuts:
# Ctrl+B then Arrow Keys - Switch between panes
# Ctrl+B then D - Detach (keeps running in background)
# Ctrl+B then Z - Zoom current pane (toggle fullscreen)
# Ctrl+B then X - Kill current pane

# To reattach later
make dev-attach

# To restart everything
make dev-restart

# To stop everything
make kill-all
```

#### Option 3: Full Stack with Docker

Run everything in Docker (including Kratos for auth):

```bash
make docker-up

# View logs
make docker-logs

# Stop everything
make docker-down
```

### Verify Installation

```bash
# Check API health
curl http://localhost:8080/health

# Expected response:
# {"status":"healthy","version":"dev"}
```

---

## Project Architecture

### Directory Structure

```
gorax/
├── cmd/                      # Application entry points
│   ├── api/                  # API server (Chi router)
│   └── worker/               # Background worker (job processing)
│
├── internal/                 # Private application code
│   ├── api/                  # HTTP handlers and routing
│   │   ├── handlers/         # Request handlers
│   │   ├── middleware/       # HTTP middleware
│   │   └── router/           # Route definitions
│   │
│   ├── workflow/             # Core workflow engine
│   │   ├── repository.go     # Data access
│   │   ├── service.go        # Business logic
│   │   └── models.go         # Domain models
│   │
│   ├── executor/             # Workflow execution engine
│   │   ├── actions/          # Action implementations
│   │   ├── expression/       # Expression evaluation (Expr)
│   │   ├── executor.go       # Main execution logic
│   │   └── context.go        # Execution context
│   │
│   ├── integrations/         # Third-party integrations
│   │   ├── slack/            # Slack integration
│   │   ├── github/           # GitHub integration
│   │   ├── jira/             # Jira integration
│   │   ├── pagerduty/        # PagerDuty integration
│   │   └── ai/               # AI integrations (OpenAI, etc.)
│   │
│   ├── webhook/              # Webhook management
│   │   ├── service.go        # Webhook service
│   │   ├── filter.go         # Event filtering
│   │   └── replay.go         # Event replay
│   │
│   ├── schedule/             # Cron scheduling
│   ├── credential/           # Secure credential storage (KMS)
│   ├── tenant/               # Multi-tenancy support
│   ├── rbac/                 # Role-based access control
│   ├── humantask/            # Human approval workflows
│   ├── notification/         # Notification system
│   ├── retention/            # Data retention policies
│   ├── queue/                # Job queue (SQS/Redis)
│   ├── metrics/              # Prometheus metrics
│   ├── tracing/              # OpenTelemetry tracing
│   ├── errortracking/        # Sentry integration
│   └── config/               # Configuration management
│
├── web/                      # React frontend
│   ├── src/
│   │   ├── components/       # Reusable UI components
│   │   │   ├── canvas/       # Workflow visual editor
│   │   │   ├── webhooks/     # Webhook UI components
│   │   │   └── schedule/     # Scheduling UI
│   │   ├── pages/            # Route-level pages
│   │   ├── hooks/            # Custom React hooks
│   │   ├── stores/           # Zustand state stores
│   │   ├── api/              # API client and hooks
│   │   ├── types/            # TypeScript type definitions
│   │   └── utils/            # Utility functions
│   │
│   └── package.json          # Frontend dependencies
│
├── migrations/               # SQL database migrations
├── deployments/              # Deployment configurations
│   └── docker/               # Docker and Kubernetes configs
├── docs/                     # Documentation
├── examples/                 # Example workflows
├── tests/                    # Integration tests
│
├── Makefile                  # Build and dev commands
├── go.mod                    # Go dependencies
├── .env.example              # Environment template
└── CLAUDE.md                 # Coding guidelines (IMPORTANT!)
```

### Backend Architecture

**Stack:**
- **Language:** Go 1.24
- **Web Framework:** Chi (lightweight, composable router)
- **Database:** PostgreSQL 14+ with sqlx
- **Cache:** Redis 6+ (go-redis)
- **Queue:** AWS SQS or Redis Streams
- **ORM:** None - uses sqlx for direct SQL queries

**Key Packages:**

| Package | Purpose | Key Files |
|---------|---------|-----------|
| `internal/workflow` | Workflow CRUD operations | `repository.go`, `service.go` |
| `internal/executor` | Executes workflow steps | `executor.go`, `step_executor.go` |
| `internal/integrations` | External service integrations | `slack/`, `github/`, etc. |
| `internal/api/handlers` | HTTP request handlers | `workflow.go`, `execution.go` |
| `internal/credential` | Encrypted credential storage | `service.go`, `kms.go` |
| `internal/webhook` | Webhook receiver and manager | `service.go`, `filter.go` |
| `internal/schedule` | Cron-based scheduling | `scheduler.go` |
| `internal/tenant` | Multi-tenancy support | `middleware.go`, `context.go` |
| `internal/rbac` | Role-based access control | `enforcer.go` |

**Design Patterns:**
- **Repository Pattern:** Abstracts data access (e.g., `WorkflowRepository`)
- **Service Layer:** Business logic (e.g., `WorkflowService`)
- **Dependency Injection:** Constructor-based (e.g., `NewWorkflowService(repo, cache)`)
- **Middleware Chain:** Chi middleware for auth, logging, metrics
- **Strategy Pattern:** Action executors implement `Action` interface

### Frontend Architecture

**Stack:**
- **Framework:** React 18 with TypeScript
- **Build Tool:** Vite
- **State Management:** Zustand (lightweight, no boilerplate)
- **Server State:** React Query (caching, refetching, background updates)
- **Routing:** React Router v6
- **Workflow Canvas:** ReactFlow (drag-and-drop workflow builder)
- **Styling:** Tailwind CSS
- **Testing:** Vitest + React Testing Library

**Key Concepts:**

| Concept | Implementation | Example |
|---------|----------------|---------|
| **Pages** | Route-level components | `WorkflowList.tsx`, `WorkflowEditor.tsx` |
| **Components** | Reusable UI elements | `WorkflowCanvas.tsx`, `PropertyPanel.tsx` |
| **Hooks** | Custom React hooks | `useWorkflows.ts`, `useWebhooks.ts` |
| **Stores** | Zustand global state | `authStore.ts`, `uiStore.ts` |
| **API Client** | Fetch wrappers | `api/workflows.ts`, `api/executions.ts` |

**State Management Strategy:**
- **Zustand:** Global app state (auth, UI preferences)
- **React Query:** Server data (workflows, executions) with caching
- **useState:** Local component state

---

## Development Workflow

### TDD Approach (MANDATORY)

**All new code MUST follow Test-Driven Development:**

1. **Red:** Write a failing test first
2. **Green:** Write minimal code to pass the test
3. **Refactor:** Clean up while keeping tests green

**Example TDD Workflow:**

```bash
# 1. Write the test first (it will fail)
# File: internal/workflow/service_test.go
func TestCreateWorkflow(t *testing.T) {
    // Test setup...
    workflow, err := service.Create(ctx, input)
    assert.NoError(t, err)
    assert.Equal(t, "My Workflow", workflow.Name)
}

# 2. Run tests (should fail)
go test ./internal/workflow/...

# 3. Implement the feature
# File: internal/workflow/service.go
func (s *Service) Create(ctx context.Context, input CreateInput) (*Workflow, error) {
    // Implementation...
}

# 4. Run tests again (should pass)
go test ./internal/workflow/...

# 5. Refactor if needed (tests should still pass)
```

### Branch Naming Conventions

**Format:** `<ticket-number>-<short-description>`

**Examples:**
```bash
# Good
git checkout -b RFLOW-123-add-slack-integration
git checkout -b RFLOW-456-fix-webhook-replay

# Bad (no ticket number)
git checkout -b add-slack-integration
```

**Important:** Never commit directly to `main` or `dev` branches. Always use feature branches and pull requests.

### Git Workflow

```bash
# 1. Create a feature branch from main
git checkout main
git pull origin main
git checkout -b RFLOW-123-add-feature

# 2. Make changes and commit frequently
git add .
git commit -m "RFLOW-123: Add initial structure for feature"

# 3. Write tests (TDD)
# 4. Implement feature
# 5. Run tests and linting

make test
make lint

# 6. Push and create PR
git push origin RFLOW-123-add-feature

# 7. Open pull request on GitHub
```

### Code Review Checklist

Before submitting a pull request, verify:

- [ ] Tests written FIRST (TDD)
- [ ] All tests pass (`make test`)
- [ ] Linting passes (`make lint`)
- [ ] No commented-out code
- [ ] No debug statements (`console.log`, `fmt.Println`)
- [ ] Functions are small and focused (< 50 lines)
- [ ] No code duplication
- [ ] Error handling is complete
- [ ] Types are explicit (no `any` in TypeScript)
- [ ] Dependencies are injected (not hard-coded)
- [ ] Cognitive complexity < 15 per function
- [ ] Updated documentation if needed

### Running Tests

**Backend (Go):**

```bash
# Run all tests
make test

# Run tests with coverage
make test-coverage

# Run tests for specific package
go test -v ./internal/workflow/...

# Run a specific test
go test -v ./internal/workflow/... -run TestCreateWorkflow

# Run tests with race detection
go test -race ./...

# View coverage report
open coverage.html
```

**Frontend (TypeScript/React):**

```bash
cd web

# Run all tests
npm test

# Run tests in watch mode
npm test -- --watch

# Run tests with coverage
npm run test:coverage

# Run tests with UI
npm run test:ui

# Run a specific test file
npm test -- WorkflowCanvas.test.tsx
```

### Linting and Formatting

**Backend:**

```bash
# Run linter
make lint

# Format code
make fmt

# Fix linting issues automatically (if possible)
golangci-lint run --fix ./...
```

**Frontend:**

```bash
cd web

# Run linter
npm run lint

# Format code
npm run format
```

---

## Key Concepts

### Workflows and Nodes

**Workflow:** A directed graph of nodes that define an automation.

**Node Types:**
- **Trigger:** Starts the workflow (webhook, schedule)
- **Action:** Performs an operation (HTTP request, Slack message)
- **Condition:** Branching logic (if/else)
- **Loop:** Iterates over a list
- **Fork/Join:** Parallel execution
- **Human Task:** Approval workflows
- **Sub-workflow:** Nested workflow execution

**Node Structure:**

```json
{
  "id": "node-1",
  "type": "slack:send_message",
  "config": {
    "channel": "C1234567890",
    "text": "Hello {{trigger.body.user}}"
  },
  "next": "node-2"
}
```

**Key Files:**
- `internal/workflow/models.go` - Workflow data models
- `internal/executor/executor.go` - Execution engine
- `internal/executor/actions/` - Action implementations

### Triggers

**Webhook Trigger:**
```json
{
  "type": "trigger:webhook",
  "config": {
    "path": "/deploy",
    "method": "POST",
    "filters": {
      "headers.x-event-type": "deployment"
    }
  }
}
```

**Schedule Trigger:**
```json
{
  "type": "trigger:schedule",
  "config": {
    "cron": "0 9 * * MON",
    "timezone": "America/New_York"
  }
}
```

**Key Files:**
- `internal/webhook/service.go` - Webhook handling
- `internal/webhook/filter.go` - Event filtering
- `internal/schedule/scheduler.go` - Cron scheduling

### Actions

Actions perform operations in workflows. Each integration provides multiple actions.

**Available Integrations:**

| Integration | Actions | Key Files |
|-------------|---------|-----------|
| **Slack** | send_message, send_dm, add_reaction, update_message | `internal/integrations/slack/` |
| **GitHub** | create_issue, create_pr, add_comment | `internal/integrations/github/` |
| **Jira** | create_issue, transition_issue, add_comment | `internal/integrations/jira/` |
| **PagerDuty** | create_incident, trigger_alert | `internal/integrations/pagerduty/` |
| **HTTP** | request | `internal/executor/actions/http.go` |
| **Transform** | json_path, template | `internal/executor/actions/transform.go` |
| **JavaScript** | execute (Goja sandbox) | `internal/executor/actions/javascript.go` |

**Action Interface:**

```go
type Action interface {
    Execute(ctx context.Context, input ActionInput) (ActionOutput, error)
    Validate(config map[string]any) error
}
```

### Executions and Steps

**Execution:** A single run of a workflow.

**Step:** A single node execution within a workflow run.

**Execution Status:**
- `pending` - Queued, not started
- `running` - Currently executing
- `completed` - Finished successfully
- `failed` - Finished with error
- `cancelled` - Manually cancelled

**Key Files:**
- `internal/execution/repository.go` - Execution storage
- `internal/executor/step_executor.go` - Step execution logic
- `web/src/pages/ExecutionDetail.tsx` - Execution viewer UI

### Credentials and Secrets Management

**Storage:**
- **Development:** AES-256-GCM encryption with master key
- **Production:** AWS KMS for envelope encryption

**Credential Types:**
- `oauth2` - OAuth2 tokens
- `api_key` - Simple API keys
- `basic_auth` - Username/password
- `aws` - AWS credentials

**Usage:**

```go
// Store a credential
cred, err := credService.Create(ctx, CreateCredentialInput{
    Name:     "slack-bot-token",
    Type:     "api_key",
    Provider: "slack",
    Data: map[string]string{
        "token": "xoxb-...",
    },
})

// Use in workflow
config := map[string]any{
    "credential_id": cred.ID,
    "channel": "C1234567890",
}
```

**Key Files:**
- `internal/credential/service.go` - Credential management
- `internal/credential/kms.go` - AWS KMS integration
- `internal/credential/encryption.go` - Encryption logic

### Multi-Tenancy

**Tenant Isolation:**
- All data is scoped to `tenant_id`
- Middleware extracts tenant from auth context
- Database queries automatically filter by tenant

**Tenant Context:**

```go
// Extract tenant from request
tenantID := tenant.FromContext(ctx)

// Query with tenant filter
workflows, err := repo.List(ctx, tenantID, filter)
```

**Key Files:**
- `internal/tenant/middleware.go` - Tenant extraction
- `internal/tenant/context.go` - Context helpers
- `internal/api/middleware/tenant.go` - HTTP middleware

---

## Adding New Features

### Adding a New Integration

**Example:** Add a Discord integration

**1. Create integration package:**

```bash
mkdir -p internal/integrations/discord
touch internal/integrations/discord/{client.go,send_message.go,client_test.go,send_message_test.go}
```

**2. Define the client:**

```go
// internal/integrations/discord/client.go
package discord

import (
    "context"
    "github.com/gorax/gorax/internal/credential"
)

type Client struct {
    token string
}

func NewClient(credService credential.Service, credID string) (*Client, error) {
    cred, err := credService.Get(context.Background(), credID)
    if err != nil {
        return nil, err
    }

    token := cred.Data["token"]
    return &Client{token: token}, nil
}
```

**3. Implement an action:**

```go
// internal/integrations/discord/send_message.go
package discord

import (
    "context"
    "fmt"
)

type SendMessageAction struct {
    client *Client
}

func (a *SendMessageAction) Execute(ctx context.Context, input ActionInput) (ActionOutput, error) {
    channelID := input.Config["channel_id"].(string)
    text := input.Config["text"].(string)

    // Make API request to Discord
    // ...

    return ActionOutput{
        Data: map[string]any{
            "message_id": "123456",
        },
    }, nil
}

func (a *SendMessageAction) Validate(config map[string]any) error {
    if _, ok := config["channel_id"]; !ok {
        return fmt.Errorf("channel_id is required")
    }
    if _, ok := config["text"]; !ok {
        return fmt.Errorf("text is required")
    }
    return nil
}
```

**4. Write tests (TDD!):**

```go
// internal/integrations/discord/send_message_test.go
package discord

import (
    "context"
    "testing"
    "github.com/stretchr/testify/assert"
)

func TestSendMessageAction_Execute(t *testing.T) {
    // Mock setup
    action := &SendMessageAction{client: mockClient}

    input := ActionInput{
        Config: map[string]any{
            "channel_id": "987654321",
            "text": "Hello Discord!",
        },
    }

    output, err := action.Execute(context.Background(), input)

    assert.NoError(t, err)
    assert.NotEmpty(t, output.Data["message_id"])
}
```

**5. Register the action:**

```go
// internal/executor/actions/registry.go
func init() {
    Register("discord:send_message", &discord.SendMessageAction{})
}
```

**6. Add frontend components:**

```typescript
// web/src/components/canvas/nodes/DiscordNode.tsx
export function DiscordNode({ data }: NodeProps) {
  return (
    <div className="node discord-node">
      <div className="node-header">
        <MessageSquare size={16} />
        <span>Discord: Send Message</span>
      </div>
      {/* Configuration UI */}
    </div>
  );
}
```

### Adding a New Node Type

**Example:** Add a "Delay" node

**1. Define the node in executor:**

```go
// internal/executor/actions/delay.go
package actions

import (
    "context"
    "time"
)

type DelayAction struct{}

func (a *DelayAction) Execute(ctx context.Context, input ActionInput) (ActionOutput, error) {
    duration := input.Config["duration"].(string)
    d, err := time.ParseDuration(duration)
    if err != nil {
        return ActionOutput{}, err
    }

    time.Sleep(d)

    return ActionOutput{
        Data: map[string]any{
            "delayed_for": duration,
        },
    }, nil
}
```

**2. Register the action:**

```go
func init() {
    Register("core:delay", &DelayAction{})
}
```

**3. Add frontend component:**

```typescript
// web/src/components/canvas/nodes/DelayNode.tsx
export function DelayNode({ data }: NodeProps) {
  return (
    <div className="node delay-node">
      <Clock size={16} />
      <span>Delay: {data.config.duration}</span>
    </div>
  );
}
```

### Adding API Endpoints

**Example:** Add an endpoint to duplicate a workflow

**1. Create the handler:**

```go
// internal/api/handlers/workflow.go

// POST /api/v1/workflows/:id/duplicate
func (h *WorkflowHandler) Duplicate(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    tenantID := tenant.FromContext(ctx)
    workflowID := chi.URLParam(r, "id")

    // Get original workflow
    original, err := h.service.GetByID(ctx, tenantID, workflowID)
    if err != nil {
        http.Error(w, err.Error(), http.StatusNotFound)
        return
    }

    // Create duplicate
    duplicate := &workflow.Workflow{
        Name:        original.Name + " (Copy)",
        Definition:  original.Definition,
        TenantID:    tenantID,
    }

    created, err := h.service.Create(ctx, duplicate)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    json.NewEncoder(w).Encode(created)
}
```

**2. Add route:**

```go
// internal/api/router/router.go
r.Route("/workflows", func(r chi.Router) {
    r.Post("/{id}/duplicate", handlers.workflowHandler.Duplicate)
})
```

**3. Add API client method:**

```typescript
// web/src/api/workflows.ts
export async function duplicateWorkflow(id: string): Promise<Workflow> {
  const response = await fetch(`/api/v1/workflows/${id}/duplicate`, {
    method: 'POST',
    headers: {
      'X-Tenant-ID': getTenantId(),
    },
  });

  if (!response.ok) {
    throw new Error('Failed to duplicate workflow');
  }

  return response.json();
}
```

**4. Use in frontend:**

```typescript
// web/src/pages/WorkflowList.tsx
import { useMutation } from '@tanstack/react-query';
import { duplicateWorkflow } from '../api/workflows';

function WorkflowList() {
  const duplicateMutation = useMutation({
    mutationFn: duplicateWorkflow,
    onSuccess: () => {
      queryClient.invalidateQueries(['workflows']);
      toast.success('Workflow duplicated!');
    },
  });

  return (
    <button onClick={() => duplicateMutation.mutate(workflow.id)}>
      Duplicate
    </button>
  );
}
```

### Adding Frontend Components

**Pattern: Compound Components**

```typescript
// web/src/components/workflow/WorkflowCard.tsx
import { createContext, useContext } from 'react';

interface WorkflowCardContext {
  workflow: Workflow;
  onEdit: () => void;
  onDelete: () => void;
}

const WorkflowCardContext = createContext<WorkflowCardContext | null>(null);

export function WorkflowCard({ workflow, children }: Props) {
  const [isOpen, setIsOpen] = useState(false);

  return (
    <WorkflowCardContext.Provider value={{ workflow, onEdit, onDelete }}>
      <div className="workflow-card">
        {children}
      </div>
    </WorkflowCardContext.Provider>
  );
}

WorkflowCard.Header = function Header() {
  const { workflow } = useContext(WorkflowCardContext)!;
  return <h3>{workflow.name}</h3>;
};

WorkflowCard.Actions = function Actions() {
  const { onEdit, onDelete } = useContext(WorkflowCardContext)!;
  return (
    <div>
      <button onClick={onEdit}>Edit</button>
      <button onClick={onDelete}>Delete</button>
    </div>
  );
};

// Usage:
<WorkflowCard workflow={workflow}>
  <WorkflowCard.Header />
  <WorkflowCard.Actions />
</WorkflowCard>
```

---

## Testing

### Unit Testing Patterns (Go)

**Table-Driven Tests:**

```go
func TestValidateWorkflow(t *testing.T) {
    tests := []struct {
        name    string
        input   Workflow
        wantErr bool
    }{
        {
            name: "valid workflow",
            input: Workflow{
                Name: "Test Workflow",
                Definition: validDefinition,
            },
            wantErr: false,
        },
        {
            name: "missing name",
            input: Workflow{
                Definition: validDefinition,
            },
            wantErr: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := ValidateWorkflow(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("ValidateWorkflow() error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}
```

**Mocking with Interfaces:**

```go
// internal/workflow/service_test.go
type mockRepository struct {
    workflows []*Workflow
}

func (m *mockRepository) Create(ctx context.Context, w *Workflow) error {
    m.workflows = append(m.workflows, w)
    return nil
}

func TestWorkflowService_Create(t *testing.T) {
    repo := &mockRepository{}
    service := NewService(repo)

    workflow, err := service.Create(ctx, input)

    assert.NoError(t, err)
    assert.Len(t, repo.workflows, 1)
}
```

**Using testify/assert:**

```go
import (
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestExample(t *testing.T) {
    result, err := someFunction()

    require.NoError(t, err)  // Stops test if error
    assert.Equal(t, expected, result)
    assert.NotNil(t, result)
    assert.Contains(t, result.Tags, "tag1")
}
```

### Unit Testing Patterns (React/TypeScript)

**Component Testing:**

```typescript
// web/src/components/WorkflowCard.test.tsx
import { render, screen, fireEvent } from '@testing-library/react';
import { WorkflowCard } from './WorkflowCard';

describe('WorkflowCard', () => {
  const mockWorkflow = {
    id: '1',
    name: 'Test Workflow',
    status: 'active',
  };

  it('renders workflow name', () => {
    render(<WorkflowCard workflow={mockWorkflow} />);
    expect(screen.getByText('Test Workflow')).toBeInTheDocument();
  });

  it('calls onEdit when edit button clicked', () => {
    const onEdit = vi.fn();
    render(<WorkflowCard workflow={mockWorkflow} onEdit={onEdit} />);

    fireEvent.click(screen.getByText('Edit'));
    expect(onEdit).toHaveBeenCalledWith('1');
  });
});
```

**Hook Testing:**

```typescript
// web/src/hooks/useWorkflows.test.ts
import { renderHook, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { useWorkflows } from './useWorkflows';

describe('useWorkflows', () => {
  it('fetches workflows', async () => {
    const queryClient = new QueryClient();
    const wrapper = ({ children }) => (
      <QueryClientProvider client={queryClient}>
        {children}
      </QueryClientProvider>
    );

    const { result } = renderHook(() => useWorkflows(), { wrapper });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data).toHaveLength(3);
  });
});
```

### Integration Testing

**Backend Integration Tests:**

```go
// tests/integration/workflow_test.go
func TestWorkflowIntegration(t *testing.T) {
    // Setup test database
    db := setupTestDB(t)
    defer db.Close()

    // Create services
    repo := workflow.NewRepository(db)
    service := workflow.NewService(repo)

    // Test create workflow
    w, err := service.Create(ctx, input)
    require.NoError(t, err)

    // Test retrieve workflow
    retrieved, err := service.GetByID(ctx, tenantID, w.ID)
    require.NoError(t, err)
    assert.Equal(t, w.Name, retrieved.Name)
}
```

**Frontend Integration Tests:**

```typescript
// web/src/pages/WorkflowEditor.integration.test.tsx
describe('WorkflowEditor Integration', () => {
  it('saves workflow and shows success message', async () => {
    render(<WorkflowEditor />);

    // Edit workflow
    fireEvent.change(screen.getByLabelText('Name'), {
      target: { value: 'New Workflow' },
    });

    // Save
    fireEvent.click(screen.getByText('Save'));

    // Wait for success message
    await waitFor(() => {
      expect(screen.getByText('Workflow saved')).toBeInTheDocument();
    });
  });
});
```

### Mocking Strategies

**Mock HTTP Clients (Go):**

```go
// internal/integrations/slack/client_test.go
type mockHTTPClient struct {
    response *http.Response
    err      error
}

func (m *mockHTTPClient) Do(req *http.Request) (*http.Response, error) {
    return m.response, m.err
}

func TestSlackClient_SendMessage(t *testing.T) {
    mockClient := &mockHTTPClient{
        response: &http.Response{
            StatusCode: 200,
            Body: ioutil.NopCloser(strings.NewReader(`{"ok": true}`)),
        },
    }

    client := &Client{httpClient: mockClient}
    err := client.SendMessage(ctx, "C123", "Hello")

    assert.NoError(t, err)
}
```

**Mock API Calls (TypeScript):**

```typescript
// web/src/api/workflows.test.ts
import { vi } from 'vitest';

global.fetch = vi.fn();

describe('workflows API', () => {
  it('fetches workflows', async () => {
    (fetch as any).mockResolvedValueOnce({
      ok: true,
      json: async () => [{ id: '1', name: 'Workflow 1' }],
    });

    const workflows = await getWorkflows();
    expect(workflows).toHaveLength(1);
  });
});
```

### Running the Test Suite

```bash
# Backend - all tests
make test

# Backend - specific package
go test -v ./internal/workflow/...

# Backend - with coverage
make test-coverage

# Frontend - all tests
cd web && npm test

# Frontend - watch mode
cd web && npm test -- --watch

# Frontend - coverage
cd web && npm run test:coverage
```

---

## Debugging

### Common Issues and Solutions

#### Issue: Database Connection Failed

**Symptoms:**
```
Error: dial tcp [::1]:5432: connect: connection refused
```

**Solutions:**

```bash
# Check if PostgreSQL is running
docker ps | grep postgres

# Start PostgreSQL
make dev-simple

# Check connection manually
psql -h localhost -p 5433 -U postgres -d gorax

# Verify .env settings
DB_HOST=localhost
DB_PORT=5433
DB_USER=postgres
DB_PASSWORD=postgres
```

#### Issue: Redis Connection Failed

**Symptoms:**
```
Error: dial tcp 127.0.0.1:6379: connect: connection refused
```

**Solutions:**

```bash
# Check if Redis is running
docker ps | grep redis

# Start Redis
make dev-simple

# Test connection
redis-cli -h localhost -p 6379 ping
# Expected: PONG
```

#### Issue: Port Already in Use

**Symptoms:**
```
Error: listen tcp :8080: bind: address already in use
```

**Solutions:**

```bash
# Find process using port
lsof -i :8080

# Kill the process
kill -9 <PID>

# Or use the make command
make kill-api
```

#### Issue: Frontend Build Errors

**Symptoms:**
```
Error: Cannot find module 'react'
```

**Solutions:**

```bash
cd web

# Clear and reinstall
rm -rf node_modules package-lock.json
npm cache clean --force
npm install

# If using wrong Node version
nvm install 18
nvm use 18
npm install
```

#### Issue: Migration Errors

**Symptoms:**
```
Error: relation "workflows" does not exist
```

**Solutions:**

```bash
# Check migration status
psql -h localhost -p 5433 -U postgres -d gorax -c "\dt"

# Reset and re-run migrations
make db-reset

# Manually run migrations
psql -h localhost -p 5433 -U postgres -d gorax -f migrations/001_initial_schema.sql
```

#### Issue: Credential Decryption Failed

**Symptoms:**
```
Error: cipher: message authentication failed
```

**Solutions:**

```bash
# Check CREDENTIAL_MASTER_KEY is set correctly in .env
# It must be the same key used to encrypt

# Generate a new key (will invalidate existing credentials)
openssl rand -base64 32

# In production, use AWS KMS instead:
CREDENTIAL_USE_KMS=true
CREDENTIAL_KMS_KEY_ID=alias/gorax-credentials
```

### Logging

**Backend Logging:**

```go
// Use structured logging with slog
import "log/slog"

slog.Info("workflow created",
    "workflow_id", workflow.ID,
    "tenant_id", tenantID,
    "name", workflow.Name)

slog.Error("failed to execute workflow",
    "workflow_id", workflow.ID,
    "error", err)

// Add trace context
slog.InfoContext(ctx, "processing webhook",
    "webhook_id", webhookID,
    "event_type", eventType)
```

**Frontend Logging:**

```typescript
// Development: Use console
console.log('Workflow saved:', workflow);

// Production: Use Sentry
import * as Sentry from '@sentry/react';

Sentry.captureException(error, {
  tags: {
    workflow_id: workflow.id,
  },
});
```

**View Logs:**

```bash
# API server logs (if running via make)
tail -f /tmp/gorax-api.log

# Docker logs
make docker-logs

# Follow specific service
docker logs -f gorax-api

# Tmux: Just look at the pane
make dev-attach
```

### Error Tracking (Sentry)

**Setup:**

```bash
# .env
SENTRY_ENABLED=true
SENTRY_DSN=https://xxx@xxx.ingest.sentry.io/xxx
SENTRY_ENVIRONMENT=development
```

**Backend Error Tracking:**

```go
import "github.com/getsentry/sentry-go"

// Capture error
sentry.CaptureException(err)

// Capture message
sentry.CaptureMessage("Something went wrong")

// Add context
sentry.ConfigureScope(func(scope *sentry.Scope) {
    scope.SetTag("workflow_id", workflowID)
    scope.SetUser(sentry.User{ID: userID})
})
```

**Frontend Error Tracking:**

```typescript
// Automatic error boundary
<Sentry.ErrorBoundary fallback={<ErrorFallback />}>
  <App />
</Sentry.ErrorBoundary>

// Manual capture
Sentry.captureException(error);
```

### Distributed Tracing (OpenTelemetry)

**Enable Tracing:**

```bash
# .env
TRACING_ENABLED=true
TRACING_ENDPOINT=localhost:4317
TRACING_SAMPLE_RATE=1.0
```

**Add Traces:**

```go
import "go.opentelemetry.io/otel"

func (s *Service) Create(ctx context.Context, w *Workflow) error {
    ctx, span := otel.Tracer("workflow").Start(ctx, "workflow.create")
    defer span.End()

    // Add attributes
    span.SetAttributes(
        attribute.String("workflow.name", w.Name),
        attribute.String("tenant.id", w.TenantID),
    )

    // Your logic...

    if err != nil {
        span.RecordError(err)
        return err
    }

    return nil
}
```

**View Traces:**

```bash
# Start Jaeger (included in docker-compose.tracing.yml)
docker compose -f docker-compose.tracing.yml up -d

# Open Jaeger UI
open http://localhost:16686
```

---

## Deployment

### Docker Builds

**Build Images:**

```bash
# Build all images
make docker-build

# Build specific image
docker build -t gorax-api:latest -f Dockerfile .

# Multi-platform build (for ARM/AMD64)
docker buildx build --platform linux/amd64,linux/arm64 -t gorax-api:latest .
```

**Dockerfile Locations:**
- API: `/Dockerfile` (multi-stage build)
- Frontend: Built and served by API in production

### Environment Variables

**Required for Production:**

```bash
# Application
APP_ENV=production
SERVER_ADDRESS=:8080

# Database (MUST use SSL in production)
DB_HOST=your-rds-host.region.rds.amazonaws.com
DB_PORT=5432
DB_USER=gorax
DB_PASSWORD=<strong-password>
DB_NAME=gorax
DB_SSLMODE=require  # REQUIRED in production

# Redis
REDIS_ADDRESS=your-redis-host:6379
REDIS_PASSWORD=<strong-password>

# Credentials (MUST use KMS in production)
CREDENTIAL_USE_KMS=true
CREDENTIAL_KMS_KEY_ID=arn:aws:kms:region:account:key/xxx
CREDENTIAL_KMS_REGION=us-east-1

# AWS
AWS_REGION=us-east-1
AWS_ACCESS_KEY_ID=<your-key>
AWS_SECRET_ACCESS_KEY=<your-secret>

# Observability
METRICS_ENABLED=true
TRACING_ENABLED=true
TRACING_ENDPOINT=otel-collector:4317
SENTRY_ENABLED=true
SENTRY_DSN=https://xxx@xxx.ingest.sentry.io/xxx
SENTRY_ENVIRONMENT=production

# CORS (HTTPS only in production)
CORS_ALLOWED_ORIGINS=https://app.gorax.io,https://admin.gorax.io

# Security Headers (auto-configured for production)
SECURITY_HEADER_ENABLE_HSTS=true
```

**Validation:**

The application validates production configuration at startup. See `.env.example` for all validation rules.

### Database Migrations

**Apply Migrations:**

```bash
# Development
make db-migrate

# Production (use migration tool or manual apply)
psql -h $DB_HOST -U $DB_USER -d $DB_NAME -f migrations/001_initial_schema.sql
psql -h $DB_HOST -U $DB_USER -d $DB_NAME -f migrations/002_webhook_events.sql
# ... etc
```

**Migration Files:**
- Location: `/migrations/`
- Naming: `001_description.sql`, `002_description.sql`, etc.
- Always include rollback scripts: `001_description_rollback.sql`

**Best Practices:**
- Test migrations on staging first
- Backup database before applying
- Run migrations during low-traffic windows
- Monitor for lock contention on large tables

### Production Checklist

Before deploying to production:

- [ ] Set `APP_ENV=production`
- [ ] Use strong, unique passwords for DB and Redis
- [ ] Enable SSL for database (`DB_SSLMODE=require`)
- [ ] Use AWS KMS for credentials (`CREDENTIAL_USE_KMS=true`)
- [ ] Configure CORS with HTTPS origins only
- [ ] Enable observability (metrics, tracing, Sentry)
- [ ] Set up database backups
- [ ] Configure auto-scaling for API and worker
- [ ] Set up monitoring and alerting
- [ ] Test disaster recovery procedures
- [ ] Document runbooks for common issues

---

## Additional Resources

### Documentation

- [API Reference](API_REFERENCE.md) - Complete API documentation
- [Getting Started](getting-started.md) - Quick start guide
- [First Workflow](first-workflow.md) - Build your first workflow
- [Security](SECURITY.md) - Security best practices
- [Observability](observability.md) - Metrics, tracing, and logging
- [RBAC](RBAC_IMPLEMENTATION.md) - Role-based access control

### Coding Standards

- [CLAUDE.md](../CLAUDE.md) - **MANDATORY** coding guidelines
  - TDD approach
  - Clean code principles
  - SOLID principles
  - Cognitive complexity rules
  - Design patterns
  - Available slash commands (`/review`, `/test-plan`, etc.)

### Examples

- [Example Workflows](../examples/) - Sample workflows
- [Integration Examples](../examples/integrations/) - Integration usage

### Community

- [GitHub Discussions](https://github.com/stherrien/gorax/discussions)
- [GitHub Issues](https://github.com/stherrien/gorax/issues)
- Email: shawn@gorax.dev

---

## Quick Command Reference

### Development

```bash
# Start dev environment
make dev-simple           # Postgres + Redis only
make dev-start            # Tmux: API + Frontend side-by-side
make run-api-dev          # API only
cd web && npm run dev     # Frontend only

# Testing
make test                 # Backend tests
make test-coverage        # With coverage
cd web && npm test        # Frontend tests

# Linting
make lint                 # Backend
cd web && npm run lint    # Frontend

# Clean up
make kill-all             # Kill all processes
make docker-down          # Stop Docker containers
```

### Build

```bash
make build                # Build all binaries
make build-api            # API only
make build-worker         # Worker only
make docker-build         # Docker images
```

### Database

```bash
make db-up                # Start postgres + redis
make db-migrate           # Apply migrations
make db-reset             # Reset and re-migrate
```

### Troubleshooting

```bash
make status               # Show running processes
make kill-api             # Kill API server
make kill-web             # Kill web server
make kill-worker          # Kill worker
lsof -i :8080             # Check port usage
docker logs gorax-api     # View API logs
```

---

**Welcome to the Gorax community! Happy coding!**
