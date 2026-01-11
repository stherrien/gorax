# Frontend Development Best Practices for Gorax

## Critical Rules (MUST FOLLOW)

### 1. Server State Management

**❌ NEVER DO THIS:**
```typescript
const [workflows, setWorkflows] = useState<Workflow[]>([])
const [loading, setLoading] = useState(true)

useEffect(() => {
  async function fetch() {
    const data = await api.list()
    setWorkflows(data)
    setLoading(false)
  }
  fetch()
}, [])  // Missing dependencies, error handling, etc.
```

**✅ ALWAYS DO THIS:**
```typescript
import { useQuery } from '@tanstack/react-query'

export function useWorkflows(params?: Params) {
  return useQuery({
    queryKey: ['workflows', params],
    queryFn: () => workflowAPI.list(params),
    staleTime: 30000,  // Cache for 30 seconds
  })
}

// Usage
const { data, isLoading, error } = useWorkflows()
```

### 2. Dependency Arrays

**❌ NEVER DO THIS:**
```typescript
const params = { page: 1, limit: 20 }  // New object every render!

useEffect(() => {
  fetch(params)
}, [params])  // INFINITE LOOP!

const callback = useCallback(() => {
  doSomething(params)
}, [params])  // INFINITE LOOP!
```

**✅ ALWAYS DO THIS:**
```typescript
// Option A: Use primitive values
useEffect(() => {
  fetch({ page, limit })
}, [page, limit])  // Primitive dependencies

// Option B: useMemo for objects
const params = useMemo(() => ({ page, limit }), [page, limit])
useEffect(() => {
  fetch(params)
}, [params])  // Now stable reference

// Option C: TanStack Query (BEST)
useQuery({
  queryKey: ['data', { page, limit }],  // Query handles this correctly
  queryFn: () => fetch({ page, limit }),
})
```

### 3. useCallback Usage

**❌ DON'T OVERUSE:**
```typescript
// Every function wrapped unnecessarily
const handleClick = useCallback(() => {
  doSimpleThing()
}, [])

const handleChange = useCallback((e) => {
  setValue(e.target.value)
}, [])

// Component full of useless useCallbacks
```

**✅ USE ONLY WHEN NEEDED:**
```typescript
// Use for: expensive operations, dependencies for other hooks
const handleSubmit = useCallback(async () => {
  await expensiveOperation()
}, [dependency])

// DON'T use for: simple inline functions
const handleClick = () => doSimpleThing()  // Just inline it!
```

### 4. Data Fetching

**❌ NEVER:**
- Fetch in useEffect
- Manual loading state for server data
- Multiple useStates for one query

**✅ ALWAYS:**
- Use TanStack Query
- Let React Query handle loading/error states
- Use mutations for create/update/delete

### 5. Forms

**❌ NEVER:**
```typescript
const [field1, setField1] = useState('')
const [field2, setField2] = useState('')
const [field3, setField3] = useState('')
// Too many useState calls
```

**✅ ALWAYS:**
```typescript
const [formData, setFormData] = useState({ field1: '', field2: '', field3: '' })

// Or use React Hook Form:
import { useForm } from 'react-hook-form'
const { register, handleSubmit } = useForm()
```

---

## Pre-Commit Checklist

```bash
# REQUIRED before every frontend commit:

# 1. Type check
npx tsc --noEmit

# 2. Lint
npm run lint

# 3. Manual browser test
# - Load the page
# - Check Network tab (no infinite loops)
# - Check Console (no React warnings)
# - Test all interactions
# - Test loading states
# - Test error states

# 4. Anti-pattern check
./scripts/frontend-quality-check.sh
```

---

## React Query Patterns

### List Query
```typescript
export function useWorkflows(params?: Params) {
  return useQuery({
    queryKey: ['workflows', params],
    queryFn: () => workflowAPI.list(params),
    staleTime: 30000,
  })
}
```

### Single Item Query
```typescript
export function useWorkflow(id: string) {
  return useQuery({
    queryKey: ['workflow', id],
    queryFn: () => workflowAPI.get(id),
    enabled: !!id,
  })
}
```

### Mutations
```typescript
export function useCreateWorkflow() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (input: Input) => workflowAPI.create(input),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['workflows'] })
    },
  })
}
```

---

## Common Mistakes & Fixes

| Mistake | Fix |
|---------|-----|
| Object in deps | Use primitives or useMemo |
| Fetch in useEffect | Use useQuery |
| Manual loading state | Use isLoading from useQuery |
| Inline setState | Batch updates or use single state |
| Missing error boundary | Wrap features |
| Not testing in browser | ALWAYS test manually |

---

## Testing React Query Components

```typescript
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { renderHook, waitFor } from '@testing-library/react'

function createWrapper() {
  const queryClient = new QueryClient({
    defaultOptions: {
      queries: { retry: false },
    },
  })
  return ({ children }: { children: React.ReactNode }) => (
    <QueryClientProvider client={queryClient}>
      {children}
    </QueryClientProvider>
  )
}

test('should fetch workflows', async () => {
  const { result } = renderHook(() => useWorkflows(), {
    wrapper: createWrapper(),
  })

  await waitFor(() => expect(result.current.isSuccess).toBe(true))
  expect(result.current.data).toBeDefined()
})
```

---

## Agent Instructions for Frontend

When asking AI agents to build frontend:

```
Build React component following these STRICT rules:

1. Use TanStack Query for ALL server state (import from @tanstack/react-query)
2. NEVER use object/array literals in dependency arrays
3. Test the component manually in browser before finishing
4. Verify API endpoints exist and work
5. Follow patterns from web/src/hooks/useWorkflows.ts (the fixed version)
6. Use TypeScript strict mode
7. Include error boundaries
8. Test loading and error states
9. Check Network tab for infinite loops
10. Run: npx tsc --noEmit before completing

Do NOT generate code without following these rules.
```

---

## Summary

The frontend issues stemmed from:
1. Manual state management for server data (should use TanStack Query)
2. Object references in dependency arrays (causes infinite loops)
3. Not testing in browser before committing
4. Generating too many components without verifying patterns

**All fixed now** with TanStack Query migration. Going forward, follow this guide and the checklist to prevent these issues.
