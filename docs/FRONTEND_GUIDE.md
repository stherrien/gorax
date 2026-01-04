# Gorax Frontend Development Guide

A comprehensive guide to developing the Gorax React application.

## Table of Contents

1. [Overview](#overview)
2. [Getting Started](#getting-started)
3. [Architecture & Patterns](#architecture--patterns)
4. [Core Technologies](#core-technologies)
5. [Component Development](#component-development)
6. [State Management](#state-management)
7. [API Integration](#api-integration)
8. [Testing](#testing)
9. [Performance Optimization](#performance-optimization)
10. [Common Patterns](#common-patterns)
11. [Build & Deployment](#build--deployment)
12. [Troubleshooting](#troubleshooting)

---

## Overview

### Frontend Stack

Gorax uses a modern, type-safe React stack:

- **React 18.3.1** - UI library with concurrent rendering
- **TypeScript 5.2** - Type-safe JavaScript
- **Vite 7.3** - Fast build tool and dev server
- **Tailwind CSS 3.4** - Utility-first CSS framework
- **ReactFlow 12** - Node-based workflow canvas
- **Zustand 4.5** - Lightweight state management
- **TanStack Query 5.90** - Server state management
- **React Router 6.26** - Client-side routing
- **Vitest 4.0** - Fast unit testing framework
- **Testing Library** - Component testing utilities

### Project Structure

```
web/
├── src/
│   ├── api/              # API client and request functions
│   ├── components/       # Reusable UI components
│   │   ├── canvas/      # ReactFlow workflow canvas
│   │   ├── nodes/       # Custom ReactFlow nodes
│   │   ├── common/      # Shared UI components
│   │   ├── ui/          # Base UI primitives
│   │   └── */           # Feature-specific components
│   ├── contexts/        # React Context providers
│   ├── hooks/           # Custom React hooks
│   ├── lib/             # Utility libraries
│   ├── pages/           # Route-level page components
│   ├── stores/          # Zustand state stores
│   ├── styles/          # Global styles
│   ├── types/           # TypeScript type definitions
│   ├── utils/           # Helper functions
│   ├── App.tsx          # Root application component
│   └── main.tsx         # Application entry point
├── public/              # Static assets
├── package.json         # Dependencies and scripts
├── vite.config.ts       # Vite configuration
├── tailwind.config.js   # Tailwind CSS configuration
├── tsconfig.json        # TypeScript configuration
└── .eslintrc.cjs        # ESLint configuration
```

### Key Features

- **Workflow Canvas Editor** - Drag-and-drop visual workflow builder
- **Real-time Collaboration** - Multi-user editing with WebSocket
- **Execution Monitoring** - Live workflow execution tracking
- **Credential Management** - Secure credential storage UI
- **Schedule Management** - Cron-based workflow scheduling
- **Webhook Management** - Webhook testing and replay
- **Analytics Dashboard** - Workflow metrics and insights
- **Marketplace** - Browse and install workflow templates

---

## Getting Started

### Prerequisites

- **Node.js 18+** (LTS recommended)
- **npm 9+** or **yarn 1.22+**
- **Go 1.21+** (for backend)

### Setup

1. **Clone the repository**:
   ```bash
   git clone <repo-url>
   cd gorax/web
   ```

2. **Install dependencies**:
   ```bash
   npm install
   ```

3. **Configure environment**:
   Create `.env` file (optional):
   ```env
   VITE_API_URL=http://localhost:8080
   VITE_SENTRY_DSN=
   VITE_SENTRY_ENABLED=false
   ```

4. **Start the development server**:
   ```bash
   npm run dev
   ```
   The app will be available at `http://localhost:5173`

5. **Start the backend** (in another terminal):
   ```bash
   cd ..
   make run
   ```
   The API will be available at `http://localhost:8080`

### Development Commands

```bash
npm run dev           # Start dev server (port 5173)
npm run build         # Build for production
npm run preview       # Preview production build
npm run lint          # Lint TypeScript and React code
npm test              # Run tests in watch mode
npm run test:ui       # Run tests with UI
npm run test:coverage # Generate coverage report
```

### IDE Setup

**VSCode Extensions** (recommended):
- ESLint
- Prettier
- Tailwind CSS IntelliSense
- TypeScript Vue Plugin (for better TS support)

**VSCode Settings** (`.vscode/settings.json`):
```json
{
  "editor.formatOnSave": true,
  "editor.codeActionsOnSave": {
    "source.fixAll.eslint": true
  },
  "typescript.tsdk": "node_modules/typescript/lib"
}
```

---

## Architecture & Patterns

### Clean Architecture Principles

The frontend follows clean architecture principles:

1. **UI Layer** (`components/`, `pages/`) - Presentation logic
2. **Application Layer** (`hooks/`) - Business logic and orchestration
3. **Data Layer** (`api/`, `stores/`) - Data access and state
4. **Domain Layer** (`types/`) - Business models and interfaces

### Component Structure

**Functional Components Only** - No class components:
```typescript
// Good: Functional component
export function MyComponent({ prop1, prop2 }: MyComponentProps) {
  return <div>{prop1}</div>
}

// Bad: Class component (deprecated)
export class MyComponent extends React.Component {
  render() { return <div>{this.props.prop1}</div> }
}
```

**Component Patterns**:
- **Presentational Components** - Pure UI, no side effects
- **Container Components** - Data fetching and business logic
- **Compound Components** - Related components sharing state
- **Render Props** - Dynamic component composition

### File Naming Conventions

```
ComponentName.tsx           # Component implementation
ComponentName.test.tsx      # Component tests
ComponentName.example.tsx   # Usage examples (Storybook-like)
index.ts                    # Public exports
```

### Code Organization

**Group by Feature, Not by Type**:
```
✅ Good:
components/
  credentials/
    CredentialForm.tsx
    CredentialList.tsx
    CredentialPicker.tsx
  workflows/
    WorkflowCanvas.tsx
    WorkflowList.tsx

❌ Bad:
forms/
  CredentialForm.tsx
  WorkflowForm.tsx
lists/
  CredentialList.tsx
  WorkflowList.tsx
```

### Dependency Flow

```
Pages → Hooks → API/Stores → Types
  ↓       ↓         ↓
Components ←←←←←←←←←
```

**Rules**:
- Pages can use components, hooks, API, stores
- Hooks can use API, stores, other hooks
- Components should be presentational (props-driven)
- No circular dependencies

---

## Core Technologies

### React 18 Features

**Concurrent Rendering**:
```typescript
import { Suspense } from 'react'

function App() {
  return (
    <Suspense fallback={<LoadingSpinner />}>
      <WorkflowList />
    </Suspense>
  )
}
```

**Automatic Batching**:
```typescript
// Multiple state updates are automatically batched
function handleClick() {
  setCount(c => c + 1)
  setFlag(f => !f)
  // Only triggers ONE re-render
}
```

**Transitions** (for non-urgent updates):
```typescript
import { useTransition } from 'react'

function SearchBox() {
  const [isPending, startTransition] = useTransition()
  const [query, setQuery] = useState('')

  const handleChange = (e) => {
    startTransition(() => {
      setQuery(e.target.value) // Non-urgent update
    })
  }

  return (
    <input value={query} onChange={handleChange} />
  )
}
```

### TypeScript Best Practices

**1. Use Explicit Types**:
```typescript
// ✅ Good: Explicit types
interface WorkflowListProps {
  workflows: Workflow[]
  onSelect: (id: string) => void
  loading?: boolean
}

// ❌ Bad: Implicit any
function WorkflowList({ workflows, onSelect }) { }
```

**2. Avoid `any`**:
```typescript
// ✅ Good: Use specific types or unknown
function processData(data: unknown) {
  if (typeof data === 'object' && data !== null) {
    // Type narrowing
  }
}

// ❌ Bad: Using any
function processData(data: any) { }
```

**3. Use Discriminated Unions**:
```typescript
type ApiResponse<T> =
  | { status: 'loading' }
  | { status: 'success'; data: T }
  | { status: 'error'; error: string }

function useApiData() {
  const [state, setState] = useState<ApiResponse<Workflow>>({ status: 'loading' })

  if (state.status === 'success') {
    // TypeScript knows state.data exists
    return state.data
  }
}
```

**4. Use Type Guards**:
```typescript
function isWorkflow(obj: any): obj is Workflow {
  return obj && typeof obj === 'object' && 'id' in obj && 'definition' in obj
}
```

### Tailwind CSS Patterns

**Component-Specific Styles**:
```typescript
// Use composition over duplication
const buttonClasses = {
  base: 'px-4 py-2 rounded-md font-medium transition-colors',
  primary: 'bg-primary-600 text-white hover:bg-primary-700',
  secondary: 'bg-gray-200 text-gray-800 hover:bg-gray-300',
  danger: 'bg-red-600 text-white hover:bg-red-700',
}

function Button({ variant = 'primary', children }: ButtonProps) {
  return (
    <button className={`${buttonClasses.base} ${buttonClasses[variant]}`}>
      {children}
    </button>
  )
}
```

**Responsive Design**:
```typescript
<div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
  {/* Responsive grid: 1 col on mobile, 2 on tablet, 3 on desktop */}
</div>
```

**Dark Mode Support**:
```typescript
// Use context for theme
import { useThemeContext } from '@/contexts/ThemeContext'

function MyComponent() {
  const { isDark } = useThemeContext()

  return (
    <div className={isDark ? 'bg-gray-900 text-white' : 'bg-white text-gray-900'}>
      Content
    </div>
  )
}
```

### ReactFlow for Workflow Canvas

**Basic Setup**:
```typescript
import { ReactFlow, Background, Controls, MiniMap } from '@xyflow/react'
import '@xyflow/react/dist/style.css'

function WorkflowCanvas() {
  const [nodes, setNodes, onNodesChange] = useNodesState([])
  const [edges, setEdges, onEdgesChange] = useEdgesState([])

  return (
    <ReactFlow
      nodes={nodes}
      edges={edges}
      onNodesChange={onNodesChange}
      onEdgesChange={onEdgesChange}
      nodeTypes={customNodeTypes}
    >
      <Background />
      <Controls />
      <MiniMap />
    </ReactFlow>
  )
}
```

**Custom Node Types**:
```typescript
// Define custom node component
function TriggerNode({ data }: NodeProps) {
  return (
    <div className="px-4 py-2 border-2 border-blue-500 rounded-lg bg-white">
      <Handle type="source" position={Position.Bottom} />
      <div className="font-medium">{data.label}</div>
    </div>
  )
}

// Register node type
const nodeTypes = {
  trigger: TriggerNode,
  action: ActionNode,
  condition: ConditionNode,
}
```

---

## Component Development

### Creating a New Component

**1. Define Props Interface**:
```typescript
// components/workflows/WorkflowCard.tsx
interface WorkflowCardProps {
  workflow: Workflow
  onEdit?: (id: string) => void
  onDelete?: (id: string) => void
  onExecute?: (id: string) => void
  className?: string
}
```

**2. Implement Component**:
```typescript
export function WorkflowCard({
  workflow,
  onEdit,
  onDelete,
  onExecute,
  className = '',
}: WorkflowCardProps) {
  return (
    <div className={`border rounded-lg p-4 ${className}`}>
      <h3 className="text-lg font-semibold">{workflow.name}</h3>
      <p className="text-gray-600 text-sm">{workflow.description}</p>

      <div className="mt-4 flex gap-2">
        {onEdit && <button onClick={() => onEdit(workflow.id)}>Edit</button>}
        {onExecute && <button onClick={() => onExecute(workflow.id)}>Run</button>}
        {onDelete && <button onClick={() => onDelete(workflow.id)}>Delete</button>}
      </div>
    </div>
  )
}
```

**3. Write Tests**:
```typescript
// components/workflows/WorkflowCard.test.tsx
import { describe, it, expect, vi } from 'vitest'
import { render, screen, fireEvent } from '@testing-library/react'
import { WorkflowCard } from './WorkflowCard'

describe('WorkflowCard', () => {
  const mockWorkflow = {
    id: 'wf-1',
    name: 'Test Workflow',
    description: 'Test description',
    status: 'active',
  }

  it('renders workflow information', () => {
    render(<WorkflowCard workflow={mockWorkflow} />)
    expect(screen.getByText('Test Workflow')).toBeInTheDocument()
    expect(screen.getByText('Test description')).toBeInTheDocument()
  })

  it('calls onEdit when edit button is clicked', () => {
    const onEdit = vi.fn()
    render(<WorkflowCard workflow={mockWorkflow} onEdit={onEdit} />)

    fireEvent.click(screen.getByText('Edit'))
    expect(onEdit).toHaveBeenCalledWith('wf-1')
  })
})
```

### Component Patterns

**Compound Components** (for flexible APIs):
```typescript
// Usage
<Schedule>
  <Schedule.Header />
  <Schedule.Calendar events={events} />
  <Schedule.List events={events} />
</Schedule>

// Implementation
function Schedule({ children }: { children: React.ReactNode }) {
  const [view, setView] = useState<'calendar' | 'list'>('calendar')

  return (
    <ScheduleContext.Provider value={{ view, setView }}>
      <div className="schedule-container">{children}</div>
    </ScheduleContext.Provider>
  )
}

Schedule.Header = function Header() {
  const { view, setView } = useScheduleContext()
  return (
    <div>
      <button onClick={() => setView('calendar')}>Calendar</button>
      <button onClick={() => setView('list')}>List</button>
    </div>
  )
}
```

**Render Props** (for dynamic composition):
```typescript
interface DataTableProps<T> {
  data: T[]
  loading: boolean
  renderRow: (item: T, index: number) => React.ReactNode
  renderEmpty?: () => React.ReactNode
}

function DataTable<T>({ data, loading, renderRow, renderEmpty }: DataTableProps<T>) {
  if (loading) return <Spinner />
  if (data.length === 0) return renderEmpty ? renderEmpty() : <EmptyState />

  return (
    <div className="space-y-2">
      {data.map((item, index) => (
        <div key={index}>{renderRow(item, index)}</div>
      ))}
    </div>
  )
}

// Usage
<DataTable
  data={workflows}
  loading={loading}
  renderRow={(workflow) => <WorkflowCard workflow={workflow} />}
  renderEmpty={() => <div>No workflows found</div>}
/>
```

### Styling Guidelines

**1. Use Tailwind Utility Classes**:
```typescript
// ✅ Good: Utility classes
<button className="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700">
  Click me
</button>

// ❌ Bad: Inline styles
<button style={{ padding: '8px 16px', backgroundColor: 'blue' }}>
  Click me
</button>
```

**2. Extract Common Patterns**:
```typescript
// Create reusable class name helpers
const cardStyles = 'border rounded-lg p-4 shadow-sm hover:shadow-md transition-shadow'

<div className={cardStyles}>Content</div>
```

**3. Conditional Classes**:
```typescript
// Use template literals for conditional classes
<div className={`
  base-class
  ${isActive ? 'bg-blue-600' : 'bg-gray-200'}
  ${isDark ? 'text-white' : 'text-gray-900'}
`}>
  Content
</div>

// Or use a helper library like clsx
import clsx from 'clsx'

<div className={clsx(
  'base-class',
  isActive && 'bg-blue-600',
  !isActive && 'bg-gray-200'
)}>
```

### Accessibility

**1. Semantic HTML**:
```typescript
// ✅ Good: Semantic elements
<button onClick={handleClick}>Click me</button>

// ❌ Bad: Non-semantic elements
<div onClick={handleClick}>Click me</div>
```

**2. ARIA Labels**:
```typescript
<button
  aria-label="Delete workflow"
  onClick={handleDelete}
>
  <TrashIcon />
</button>
```

**3. Keyboard Navigation**:
```typescript
<div
  role="button"
  tabIndex={0}
  onKeyDown={(e) => {
    if (e.key === 'Enter' || e.key === ' ') {
      handleClick()
    }
  }}
  onClick={handleClick}
>
  Custom button
</div>
```

---

## State Management

### When to Use What

| State Type | Solution | Example |
|------------|----------|---------|
| Local component state | `useState` | Form inputs, toggles |
| Shared UI state | Zustand | Canvas state, theme |
| Server data | TanStack Query | Workflows, executions |
| URL state | React Router | Filters, pagination |
| Form state | React Hook Form | Complex forms |

### Zustand (Global State)

**Creating a Store**:
```typescript
// stores/workflowStore.ts
import { create } from 'zustand'

interface WorkflowState {
  nodes: Node[]
  edges: Edge[]
  selectedNode: string | null
  isDirty: boolean

  setNodes: (nodes: Node[]) => void
  setEdges: (edges: Edge[]) => void
  addNode: (node: Node) => void
  selectNode: (id: string | null) => void
  setDirty: (dirty: boolean) => void
}

export const useWorkflowStore = create<WorkflowState>((set, get) => ({
  nodes: [],
  edges: [],
  selectedNode: null,
  isDirty: false,

  setNodes: (nodes) => set({ nodes, isDirty: true }),
  setEdges: (edges) => set({ edges, isDirty: true }),

  addNode: (node) => set({
    nodes: [...get().nodes, node],
    isDirty: true,
  }),

  selectNode: (id) => set({ selectedNode: id }),
  setDirty: (dirty) => set({ isDirty: dirty }),
}))
```

**Using the Store**:
```typescript
function WorkflowEditor() {
  // Subscribe to specific state
  const nodes = useWorkflowStore((state) => state.nodes)
  const addNode = useWorkflowStore((state) => state.addNode)

  // Or subscribe to multiple
  const { nodes, edges, addNode } = useWorkflowStore()

  return <div>{/* ... */}</div>
}
```

**Performance Optimization**:
```typescript
// ✅ Good: Selective subscription
const nodes = useWorkflowStore((state) => state.nodes)

// ❌ Bad: Subscribes to entire store
const store = useWorkflowStore()
```

### TanStack Query (Server State)

**Basic Query**:
```typescript
import { useQuery } from '@tanstack/react-query'
import { workflowAPI } from '@/api/workflows'

function WorkflowList() {
  const { data, isLoading, error, refetch } = useQuery({
    queryKey: ['workflows'],
    queryFn: () => workflowAPI.list(),
    staleTime: 5 * 60 * 1000, // 5 minutes
  })

  if (isLoading) return <LoadingSpinner />
  if (error) return <ErrorMessage error={error} />

  return (
    <div>
      {data?.workflows.map(wf => (
        <WorkflowCard key={wf.id} workflow={wf} />
      ))}
      <button onClick={() => refetch()}>Refresh</button>
    </div>
  )
}
```

**Mutations**:
```typescript
import { useMutation, useQueryClient } from '@tanstack/react-query'

function CreateWorkflowButton() {
  const queryClient = useQueryClient()

  const createMutation = useMutation({
    mutationFn: (data: WorkflowCreateInput) => workflowAPI.create(data),
    onSuccess: () => {
      // Invalidate and refetch
      queryClient.invalidateQueries({ queryKey: ['workflows'] })
    },
  })

  const handleCreate = () => {
    createMutation.mutate({
      name: 'New Workflow',
      definition: { nodes: [], edges: [] },
    })
  }

  return (
    <button onClick={handleCreate} disabled={createMutation.isPending}>
      {createMutation.isPending ? 'Creating...' : 'Create'}
    </button>
  )
}
```

**Optimistic Updates**:
```typescript
const updateMutation = useMutation({
  mutationFn: (update: WorkflowUpdateInput) => workflowAPI.update(id, update),
  onMutate: async (newData) => {
    // Cancel outgoing refetches
    await queryClient.cancelQueries({ queryKey: ['workflow', id] })

    // Snapshot previous value
    const previous = queryClient.getQueryData(['workflow', id])

    // Optimistically update
    queryClient.setQueryData(['workflow', id], (old: Workflow) => ({
      ...old,
      ...newData,
    }))

    return { previous }
  },
  onError: (err, newData, context) => {
    // Rollback on error
    queryClient.setQueryData(['workflow', id], context?.previous)
  },
  onSettled: () => {
    queryClient.invalidateQueries({ queryKey: ['workflow', id] })
  },
})
```

**Query Keys Best Practices**:
```typescript
// Create query key factories
const workflowKeys = {
  all: ['workflows'] as const,
  lists: () => [...workflowKeys.all, 'list'] as const,
  list: (filters: WorkflowListParams) => [...workflowKeys.lists(), filters] as const,
  details: () => [...workflowKeys.all, 'detail'] as const,
  detail: (id: string) => [...workflowKeys.details(), id] as const,
}

// Usage
useQuery({
  queryKey: workflowKeys.detail(workflowId),
  queryFn: () => workflowAPI.get(workflowId),
})

// Invalidate all workflows
queryClient.invalidateQueries({ queryKey: workflowKeys.all })
```

### Custom Hooks Pattern

**Encapsulate Logic**:
```typescript
// hooks/useWorkflow.ts
export function useWorkflow(id: string | null) {
  const { data: workflow, isLoading, error } = useQuery({
    queryKey: ['workflow', id],
    queryFn: () => workflowAPI.get(id!),
    enabled: !!id,
  })

  const updateMutation = useMutation({
    mutationFn: (updates: WorkflowUpdateInput) => workflowAPI.update(id!, updates),
  })

  const deleteMutation = useMutation({
    mutationFn: () => workflowAPI.delete(id!),
  })

  return {
    workflow,
    isLoading,
    error,
    update: updateMutation.mutate,
    delete: deleteMutation.mutate,
    isUpdating: updateMutation.isPending,
    isDeleting: deleteMutation.isPending,
  }
}
```

---

## API Integration

### API Client

The centralized API client handles authentication, error handling, and retries:

```typescript
// api/client.ts
import { APIClient } from './client'

export const apiClient = new APIClient(import.meta.env.VITE_API_URL || '')

// Automatic features:
// - Bearer token authentication from localStorage
// - Tenant ID header injection
// - JSON request/response handling
// - Exponential backoff retry on 5xx errors
// - Request timeout support
```

### Creating API Functions

**1. Define Types**:
```typescript
// types/analytics.ts
export interface WorkflowMetrics {
  total_executions: number
  success_rate: number
  avg_duration_ms: number
  error_rate: number
}

export interface TimeSeriesPoint {
  timestamp: string
  count: number
}
```

**2. Implement API Functions**:
```typescript
// api/analytics.ts
import { apiClient } from './client'
import type { WorkflowMetrics, TimeSeriesPoint } from '@/types/analytics'

export const analyticsAPI = {
  async getMetrics(
    startDate: string,
    endDate: string
  ): Promise<WorkflowMetrics> {
    return await apiClient.get('/api/v1/analytics/metrics', {
      params: { start_date: startDate, end_date: endDate },
    })
  },

  async getTimeSeries(
    workflowId: string,
    interval: '1h' | '1d' | '1w'
  ): Promise<TimeSeriesPoint[]> {
    const response = await apiClient.get(
      `/api/v1/analytics/workflows/${workflowId}/timeseries`,
      { params: { interval } }
    )
    return response.data
  },
}
```

**3. Create Custom Hook**:
```typescript
// hooks/useAnalytics.ts
import { useQuery } from '@tanstack/react-query'
import { analyticsAPI } from '@/api/analytics'

export function useWorkflowMetrics(startDate: string, endDate: string) {
  return useQuery({
    queryKey: ['analytics', 'metrics', startDate, endDate],
    queryFn: () => analyticsAPI.getMetrics(startDate, endDate),
    staleTime: 2 * 60 * 1000, // 2 minutes
  })
}

export function useWorkflowTimeSeries(
  workflowId: string,
  interval: '1h' | '1d' | '1w'
) {
  return useQuery({
    queryKey: ['analytics', 'timeseries', workflowId, interval],
    queryFn: () => analyticsAPI.getTimeSeries(workflowId, interval),
    refetchInterval: 30000, // Refetch every 30 seconds
  })
}
```

### Error Handling

**Centralized Error Types**:
```typescript
// api/client.ts exports these
export class APIError extends Error {
  constructor(message: string, public status?: number, public response?: any)
}
export class AuthError extends APIError {}
export class NotFoundError extends APIError {}
export class ValidationError extends APIError {}
```

**Handling Errors in Components**:
```typescript
import { AuthError, NotFoundError } from '@/api/client'

function WorkflowDetail({ id }: { id: string }) {
  const { data, error } = useQuery({
    queryKey: ['workflow', id],
    queryFn: () => workflowAPI.get(id),
  })

  if (error) {
    if (error instanceof NotFoundError) {
      return <NotFoundPage message="Workflow not found" />
    }
    if (error instanceof AuthError) {
      return <RedirectToLogin />
    }
    return <ErrorMessage error={error} />
  }

  return <div>{/* render workflow */}</div>
}
```

### Loading States

**Skeleton Screens**:
```typescript
function WorkflowList() {
  const { data, isLoading } = useQuery({
    queryKey: ['workflows'],
    queryFn: () => workflowAPI.list(),
  })

  if (isLoading) {
    return (
      <div className="space-y-4">
        {[1, 2, 3].map(i => (
          <div key={i} className="animate-pulse">
            <div className="h-8 bg-gray-200 rounded w-1/4 mb-2" />
            <div className="h-4 bg-gray-200 rounded w-1/2" />
          </div>
        ))}
      </div>
    )
  }

  return <div>{/* actual content */}</div>
}
```

---

## Testing

### Test Structure

Follow the **AAA** pattern (Arrange, Act, Assert):

```typescript
import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import { WorkflowCard } from './WorkflowCard'

describe('WorkflowCard', () => {
  // Arrange: Setup
  const mockWorkflow = {
    id: 'wf-1',
    name: 'Test Workflow',
    status: 'active',
  }

  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('calls onExecute when run button is clicked', async () => {
    // Arrange
    const onExecute = vi.fn()
    render(<WorkflowCard workflow={mockWorkflow} onExecute={onExecute} />)

    // Act
    fireEvent.click(screen.getByRole('button', { name: /run/i }))

    // Assert
    await waitFor(() => {
      expect(onExecute).toHaveBeenCalledWith('wf-1')
    })
  })
})
```

### Component Testing

**Rendering**:
```typescript
import { render, screen } from '@testing-library/react'

it('renders component', () => {
  render(<MyComponent prop="value" />)
  expect(screen.getByText('value')).toBeInTheDocument()
})
```

**User Interactions**:
```typescript
import userEvent from '@testing-library/user-event'

it('handles user input', async () => {
  const user = userEvent.setup()
  render(<SearchBox onSearch={mockSearch} />)

  const input = screen.getByRole('textbox')
  await user.type(input, 'query')
  await user.click(screen.getByRole('button', { name: /search/i }))

  expect(mockSearch).toHaveBeenCalledWith('query')
})
```

**Testing with Context**:
```typescript
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { ThemeProvider } from '@/contexts/ThemeContext'

function renderWithProviders(ui: React.ReactElement) {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  })

  return render(
    <QueryClientProvider client={queryClient}>
      <ThemeProvider>
        {ui}
      </ThemeProvider>
    </QueryClientProvider>
  )
}

it('fetches and displays data', async () => {
  renderWithProviders(<WorkflowList />)
  expect(await screen.findByText('Test Workflow')).toBeInTheDocument()
})
```

### Hook Testing

```typescript
import { renderHook, waitFor } from '@testing-library/react'
import { useWorkflow } from '@/hooks/useWorkflow'

it('fetches workflow data', async () => {
  const { result } = renderHook(() => useWorkflow('wf-1'))

  expect(result.current.isLoading).toBe(true)

  await waitFor(() => {
    expect(result.current.isLoading).toBe(false)
    expect(result.current.workflow).toBeDefined()
  })
})
```

### Mocking API Calls

**Mock fetch globally**:
```typescript
// test/setup.ts
global.fetch = vi.fn()

// In test file
beforeEach(() => {
  (global.fetch as any).mockResolvedValue({
    ok: true,
    json: async () => ({ data: mockData }),
  })
})
```

**Mock API client**:
```typescript
vi.mock('@/api/workflows', () => ({
  workflowAPI: {
    list: vi.fn().mockResolvedValue({ workflows: [], total: 0 }),
    get: vi.fn().mockResolvedValue(mockWorkflow),
    create: vi.fn().mockResolvedValue(mockWorkflow),
  },
}))
```

### Coverage Requirements

**Target Coverage**:
- UI Components: 60%+ (focus on critical paths)
- Business Logic Hooks: 80%+
- Utility Functions: 90%+

**Check Coverage**:
```bash
npm run test:coverage
```

**Coverage Report** (in `coverage/index.html`):
- View line-by-line coverage
- Identify untested branches
- Track coverage trends

---

## Performance Optimization

### Code Splitting

**Route-based Splitting**:
```typescript
import { lazy, Suspense } from 'react'

const Dashboard = lazy(() => import('./pages/Dashboard'))
const WorkflowEditor = lazy(() => import('./pages/WorkflowEditor'))

function App() {
  return (
    <Suspense fallback={<LoadingSpinner />}>
      <Routes>
        <Route path="/" element={<Dashboard />} />
        <Route path="/workflows/:id" element={<WorkflowEditor />} />
      </Routes>
    </Suspense>
  )
}
```

**Component-based Splitting**:
```typescript
const HeavyChart = lazy(() => import('./components/HeavyChart'))

function Dashboard() {
  return (
    <div>
      <Suspense fallback={<div>Loading chart...</div>}>
        <HeavyChart data={data} />
      </Suspense>
    </div>
  )
}
```

### Memoization

**useMemo** (expensive calculations):
```typescript
function WorkflowList({ workflows }: { workflows: Workflow[] }) {
  const sortedWorkflows = useMemo(() => {
    return [...workflows].sort((a, b) =>
      a.name.localeCompare(b.name)
    )
  }, [workflows])

  return <div>{/* render sortedWorkflows */}</div>
}
```

**useCallback** (stable function references):
```typescript
function Parent() {
  const [count, setCount] = useState(0)

  // Stable reference (won't cause Child to re-render)
  const handleClick = useCallback(() => {
    console.log('clicked')
  }, [])

  return <Child onClick={handleClick} />
}

const Child = React.memo(({ onClick }: { onClick: () => void }) => {
  return <button onClick={onClick}>Click</button>
})
```

**React.memo** (prevent re-renders):
```typescript
// Component only re-renders if props change
export const WorkflowCard = React.memo(
  ({ workflow, onEdit }: WorkflowCardProps) => {
    return <div>{/* ... */}</div>
  },
  (prevProps, nextProps) => {
    // Custom comparison (optional)
    return prevProps.workflow.id === nextProps.workflow.id
  }
)
```

### Bundle Size Optimization

**Analyze Bundle**:
```bash
npm run build
# Check output for chunk sizes
```

**Lazy Load Heavy Dependencies**:
```typescript
// Instead of:
import Recharts from 'recharts'

// Use dynamic import:
const loadRecharts = () => import('recharts')

function Chart() {
  const [Recharts, setRecharts] = useState<any>(null)

  useEffect(() => {
    loadRecharts().then(module => setRecharts(module))
  }, [])

  if (!Recharts) return <Spinner />
  return <Recharts.LineChart {...props} />
}
```

### List Virtualization

For long lists (1000+ items), use virtualization:

```typescript
import { useVirtualizer } from '@tanstack/react-virtual'

function LargeList({ items }: { items: any[] }) {
  const parentRef = useRef<HTMLDivElement>(null)

  const virtualizer = useVirtualizer({
    count: items.length,
    getScrollElement: () => parentRef.current,
    estimateSize: () => 50, // Estimated item height
  })

  return (
    <div ref={parentRef} style={{ height: '400px', overflow: 'auto' }}>
      <div
        style={{
          height: `${virtualizer.getTotalSize()}px`,
          position: 'relative',
        }}
      >
        {virtualizer.getVirtualItems().map(virtualItem => (
          <div
            key={virtualItem.key}
            style={{
              position: 'absolute',
              top: 0,
              left: 0,
              width: '100%',
              transform: `translateY(${virtualItem.start}px)`,
            }}
          >
            {items[virtualItem.index].name}
          </div>
        ))}
      </div>
    </div>
  )
}
```

---

## Common Patterns

### Form Handling

**Simple Forms** (Controlled Components):
```typescript
function LoginForm() {
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [errors, setErrors] = useState<Record<string, string>>({})

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()

    const newErrors: Record<string, string> = {}
    if (!email) newErrors.email = 'Email is required'
    if (!password) newErrors.password = 'Password is required'

    if (Object.keys(newErrors).length > 0) {
      setErrors(newErrors)
      return
    }

    // Submit
    login({ email, password })
  }

  return (
    <form onSubmit={handleSubmit}>
      <input
        type="email"
        value={email}
        onChange={(e) => setEmail(e.target.value)}
      />
      {errors.email && <span>{errors.email}</span>}

      <input
        type="password"
        value={password}
        onChange={(e) => setPassword(e.target.value)}
      />
      {errors.password && <span>{errors.password}</span>}

      <button type="submit">Login</button>
    </form>
  )
}
```

**Complex Forms** (React Hook Form + Zod):
```typescript
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'

const workflowSchema = z.object({
  name: z.string().min(1, 'Name is required'),
  description: z.string().optional(),
  status: z.enum(['draft', 'active', 'inactive']),
})

type WorkflowFormData = z.infer<typeof workflowSchema>

function WorkflowForm({ onSubmit }: { onSubmit: (data: WorkflowFormData) => void }) {
  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<WorkflowFormData>({
    resolver: zodResolver(workflowSchema),
  })

  return (
    <form onSubmit={handleSubmit(onSubmit)}>
      <input {...register('name')} />
      {errors.name && <span>{errors.name.message}</span>}

      <textarea {...register('description')} />

      <select {...register('status')}>
        <option value="draft">Draft</option>
        <option value="active">Active</option>
      </select>

      <button type="submit">Save</button>
    </form>
  )
}
```

### Modal Dialogs

```typescript
import { useState } from 'react'

function DeleteConfirmModal({
  isOpen,
  onClose,
  onConfirm,
  itemName,
}: {
  isOpen: boolean
  onClose: () => void
  onConfirm: () => void
  itemName: string
}) {
  if (!isOpen) return null

  return (
    <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
      <div className="bg-white rounded-lg p-6 max-w-md">
        <h2 className="text-xl font-bold mb-4">Confirm Deletion</h2>
        <p>Are you sure you want to delete "{itemName}"?</p>

        <div className="mt-6 flex gap-3 justify-end">
          <button onClick={onClose}>Cancel</button>
          <button onClick={onConfirm} className="bg-red-600 text-white">
            Delete
          </button>
        </div>
      </div>
    </div>
  )
}

// Usage
function WorkflowList() {
  const [deleteModalOpen, setDeleteModalOpen] = useState(false)
  const [selectedWorkflow, setSelectedWorkflow] = useState<Workflow | null>(null)

  const handleDelete = (workflow: Workflow) => {
    setSelectedWorkflow(workflow)
    setDeleteModalOpen(true)
  }

  const confirmDelete = () => {
    if (selectedWorkflow) {
      workflowAPI.delete(selectedWorkflow.id)
      setDeleteModalOpen(false)
    }
  }

  return (
    <>
      {/* workflow list */}
      <DeleteConfirmModal
        isOpen={deleteModalOpen}
        onClose={() => setDeleteModalOpen(false)}
        onConfirm={confirmDelete}
        itemName={selectedWorkflow?.name || ''}
      />
    </>
  )
}
```

### Data Tables

```typescript
function WorkflowTable({ workflows }: { workflows: Workflow[] }) {
  const [sortField, setSortField] = useState<keyof Workflow>('name')
  const [sortDirection, setSortDirection] = useState<'asc' | 'desc'>('asc')

  const sortedWorkflows = useMemo(() => {
    return [...workflows].sort((a, b) => {
      const aVal = a[sortField]
      const bVal = b[sortField]

      if (typeof aVal === 'string' && typeof bVal === 'string') {
        return sortDirection === 'asc'
          ? aVal.localeCompare(bVal)
          : bVal.localeCompare(aVal)
      }

      return 0
    })
  }, [workflows, sortField, sortDirection])

  const handleSort = (field: keyof Workflow) => {
    if (field === sortField) {
      setSortDirection(d => d === 'asc' ? 'desc' : 'asc')
    } else {
      setSortField(field)
      setSortDirection('asc')
    }
  }

  return (
    <table className="w-full">
      <thead>
        <tr>
          <th onClick={() => handleSort('name')}>
            Name {sortField === 'name' && (sortDirection === 'asc' ? '↑' : '↓')}
          </th>
          <th onClick={() => handleSort('status')}>Status</th>
          <th>Actions</th>
        </tr>
      </thead>
      <tbody>
        {sortedWorkflows.map(wf => (
          <tr key={wf.id}>
            <td>{wf.name}</td>
            <td>{wf.status}</td>
            <td>{/* actions */}</td>
          </tr>
        ))}
      </tbody>
    </table>
  )
}
```

### Real-time Updates (WebSocket)

```typescript
import { useEffect, useState } from 'react'

function useWebSocket(url: string) {
  const [data, setData] = useState<any>(null)
  const [connected, setConnected] = useState(false)

  useEffect(() => {
    const ws = new WebSocket(url)

    ws.onopen = () => setConnected(true)
    ws.onclose = () => setConnected(false)
    ws.onmessage = (event) => setData(JSON.parse(event.data))

    return () => ws.close()
  }, [url])

  return { data, connected }
}

// Usage
function ExecutionMonitor({ executionId }: { executionId: string }) {
  const { data, connected } = useWebSocket(
    `ws://localhost:8080/api/v1/executions/${executionId}/stream`
  )

  if (!connected) return <div>Connecting...</div>

  return (
    <div>
      <div>Status: {data?.status}</div>
      <div>Progress: {data?.progress}%</div>
    </div>
  )
}
```

---

## Build & Deployment

### Production Build

```bash
# Build for production
npm run build

# Output: dist/
# - index.html
# - assets/
#   - index-[hash].js
#   - index-[hash].css
#   - vendor-react-[hash].js
#   - vendor-query-[hash].js
```

### Environment Variables

Create `.env.production`:
```env
VITE_API_URL=https://api.production.com
VITE_SENTRY_DSN=https://...
VITE_SENTRY_ENABLED=true
```

Access in code:
```typescript
const apiUrl = import.meta.env.VITE_API_URL
```

### Vite Configuration

```typescript
// vite.config.ts
export default defineConfig({
  plugins: [react()],
  server: {
    port: 5173,
    proxy: {
      '/api': 'http://localhost:8080',
    },
  },
  build: {
    sourcemap: true,
    chunkSizeWarningLimit: 1000,
    rollupOptions: {
      output: {
        manualChunks: {
          'vendor-react': ['react', 'react-dom', 'react-router-dom'],
          'vendor-query': ['@tanstack/react-query', 'zustand'],
          'vendor-ui': ['@xyflow/react', 'recharts'],
        },
      },
    },
  },
})
```

### Docker Deployment

```dockerfile
FROM node:18-alpine as build

WORKDIR /app
COPY package*.json ./
RUN npm ci
COPY . .
RUN npm run build

FROM nginx:alpine
COPY --from=build /app/dist /usr/share/nginx/html
COPY nginx.conf /etc/nginx/conf.d/default.conf
EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]
```

---

## Troubleshooting

### Common Issues

**1. "Cannot find module" errors**:
```bash
# Clear node_modules and reinstall
rm -rf node_modules package-lock.json
npm install
```

**2. Vite dev server not hot-reloading**:
- Check that files are in `src/` directory
- Ensure file extensions are `.tsx` or `.ts`
- Restart dev server

**3. TypeScript errors**:
```bash
# Check types without building
npx tsc --noEmit

# Update TypeScript definitions
npm update @types/react @types/react-dom
```

**4. Tailwind classes not applying**:
- Check `tailwind.config.js` content paths
- Ensure `@tailwind` directives are in `index.css`
- Clear browser cache

**5. Tests failing with "fetch is not defined"**:
```typescript
// Add to test/setup.ts
global.fetch = vi.fn()
```

### Debugging Tips

**React DevTools**:
- Install React DevTools browser extension
- Inspect component props and state
- Profile component renders

**Network Tab**:
- Monitor API requests
- Check request/response payloads
- Identify slow requests

**Console Errors**:
```typescript
// Add error boundaries to catch errors
import { ErrorBoundary } from 'react-error-boundary'

function ErrorFallback({ error }: { error: Error }) {
  return (
    <div>
      <h1>Something went wrong</h1>
      <pre>{error.message}</pre>
    </div>
  )
}

<ErrorBoundary FallbackComponent={ErrorFallback}>
  <App />
</ErrorBoundary>
```

---

## Additional Resources

### Documentation

- [React Docs](https://react.dev)
- [TypeScript Handbook](https://www.typescriptlang.org/docs/)
- [TanStack Query](https://tanstack.com/query/latest)
- [Zustand](https://docs.pmnd.rs/zustand/getting-started/introduction)
- [ReactFlow](https://reactflow.dev)
- [Tailwind CSS](https://tailwindcss.com/docs)
- [Vitest](https://vitest.dev)

### Code Examples

Check existing components for patterns:
- `web/src/components/credentials/CredentialForm.tsx` - Form validation
- `web/src/components/canvas/WorkflowCanvas.tsx` - ReactFlow integration
- `web/src/hooks/useCollaboration.ts` - WebSocket hooks
- `web/src/stores/workflowStore.ts` - Zustand store

### Getting Help

- Check `CLAUDE.md` for project-specific guidelines
- Review existing tests for usage examples
- Ask in team channels for architecture questions

---

## Next Steps

After reading this guide, you should be able to:

1. Set up the development environment
2. Create new components following established patterns
3. Integrate with backend APIs using TanStack Query
4. Manage state with Zustand and TanStack Query
5. Write comprehensive tests with Vitest
6. Optimize performance with code splitting and memoization
7. Debug common issues

**Recommended Learning Path**:

1. Build a simple page component
2. Add API integration with TanStack Query
3. Create a custom hook
4. Add Zustand store for shared state
5. Write comprehensive tests
6. Add performance optimizations

For more specific topics, refer to the relevant section above or check the project-specific documentation in `/docs`.
