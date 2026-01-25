import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import { MemoryRouter, Routes, Route } from 'react-router-dom'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import WebhookDetail from './WebhookDetail'
import WorkflowEditor from './WorkflowEditor'
import ExecutionDetail from './ExecutionDetail'
import EditSchedule from './EditSchedule'

// Mock API modules
vi.mock('../api/webhooks', () => ({
  webhookAPI: {
    get: vi.fn(),
    getEvents: vi.fn(),
  },
}))

vi.mock('../api/workflows', () => ({
  workflowAPI: {
    get: vi.fn(),
    list: vi.fn(),
    dryRun: vi.fn(),
  },
}))

vi.mock('../api/executions', () => ({
  executionAPI: {
    get: vi.fn(),
    getSteps: vi.fn(),
  },
}))

vi.mock('../api/schedules', () => ({
  scheduleAPI: {
    get: vi.fn(),
  },
}))

// Mock components that may cause issues
vi.mock('../components/webhooks/FilterBuilder', () => ({
  default: () => <div>Filter Builder</div>,
}))

vi.mock('../components/canvas/WorkflowCanvas', () => ({
  default: () => <div>Workflow Canvas</div>,
}))

vi.mock('../components/canvas/NodePalette', () => ({
  default: () => <div>Node Palette</div>,
}))

vi.mock('../components/canvas/PropertyPanel', () => ({
  default: () => <div>Property Panel</div>,
}))

vi.mock('../components/schedule/ScheduleForm', () => ({
  ScheduleForm: () => <div>Schedule Form</div>,
}))

const createTestQueryClient = () =>
  new QueryClient({
    defaultOptions: {
      queries: {
        retry: false,
        gcTime: 0,
      },
    },
  })

describe('Routing Integration Tests - Invalid ID Handling', () => {
  let queryClient: QueryClient

  beforeEach(() => {
    queryClient = createTestQueryClient()
    vi.clearAllMocks()
  })

  describe('WebhookDetail', () => {
    it('should redirect to /webhooks when ID is "new"', async () => {
      const TestComponent = () => {
        return (
          <QueryClientProvider client={queryClient}>
            <MemoryRouter initialEntries={['/webhooks/new']}>
              <Routes>
                <Route path="/webhooks/new" element={<WebhookDetail />} />
                <Route path="/webhooks" element={<div>Webhook List</div>} />
              </Routes>
            </MemoryRouter>
          </QueryClientProvider>
        )
      }

      render(<TestComponent />)

      await waitFor(() => {
        expect(screen.getByText('Webhook List')).toBeInTheDocument()
      })
    })

    it('should redirect to /webhooks when ID is "create"', async () => {
      const TestComponent = () => {
        return (
          <QueryClientProvider client={queryClient}>
            <MemoryRouter initialEntries={['/webhooks/create']}>
              <Routes>
                <Route path="/webhooks/:id" element={<WebhookDetail />} />
                <Route path="/webhooks" element={<div>Webhook List</div>} />
              </Routes>
            </MemoryRouter>
          </QueryClientProvider>
        )
      }

      render(<TestComponent />)

      await waitFor(() => {
        expect(screen.getByText('Webhook List')).toBeInTheDocument()
      })
    })

    it('should redirect to /webhooks when ID is invalid', async () => {
      const TestComponent = () => {
        return (
          <QueryClientProvider client={queryClient}>
            <MemoryRouter initialEntries={['/webhooks/invalid-id']}>
              <Routes>
                <Route path="/webhooks/:id" element={<WebhookDetail />} />
                <Route path="/webhooks" element={<div>Webhook List</div>} />
              </Routes>
            </MemoryRouter>
          </QueryClientProvider>
        )
      }

      render(<TestComponent />)

      await waitFor(() => {
        expect(screen.getByText('Webhook List')).toBeInTheDocument()
      })
    })
  })

  describe('ExecutionDetail', () => {
    it('should redirect to /executions when ID is "new"', async () => {
      const TestComponent = () => {
        return (
          <QueryClientProvider client={queryClient}>
            <MemoryRouter initialEntries={['/executions/new']}>
              <Routes>
                <Route path="/executions/:id" element={<ExecutionDetail />} />
                <Route path="/executions" element={<div>Execution List</div>} />
              </Routes>
            </MemoryRouter>
          </QueryClientProvider>
        )
      }

      render(<TestComponent />)

      await waitFor(() => {
        expect(screen.getByText('Execution List')).toBeInTheDocument()
      })
    })

    it('should redirect to /executions when ID is invalid', async () => {
      const TestComponent = () => {
        return (
          <QueryClientProvider client={queryClient}>
            <MemoryRouter initialEntries={['/executions/not-a-uuid']}>
              <Routes>
                <Route path="/executions/:id" element={<ExecutionDetail />} />
                <Route path="/executions" element={<div>Execution List</div>} />
              </Routes>
            </MemoryRouter>
          </QueryClientProvider>
        )
      }

      render(<TestComponent />)

      await waitFor(() => {
        expect(screen.getByText('Execution List')).toBeInTheDocument()
      })
    })
  })

  describe('EditSchedule', () => {
    it('should redirect to /schedules when ID is "new"', async () => {
      const TestComponent = () => {
        return (
          <QueryClientProvider client={queryClient}>
            <MemoryRouter initialEntries={['/schedules/new/edit']}>
              <Routes>
                <Route path="/schedules/:id/edit" element={<EditSchedule />} />
                <Route path="/schedules" element={<div>Schedule List</div>} />
              </Routes>
            </MemoryRouter>
          </QueryClientProvider>
        )
      }

      render(<TestComponent />)

      await waitFor(() => {
        expect(screen.getByText('Schedule List')).toBeInTheDocument()
      })
    })

    it('should redirect to /schedules when ID is invalid', async () => {
      const TestComponent = () => {
        return (
          <QueryClientProvider client={queryClient}>
            <MemoryRouter initialEntries={['/schedules/bad-id/edit']}>
              <Routes>
                <Route path="/schedules/:id/edit" element={<EditSchedule />} />
                <Route path="/schedules" element={<div>Schedule List</div>} />
              </Routes>
            </MemoryRouter>
          </QueryClientProvider>
        )
      }

      render(<TestComponent />)

      await waitFor(() => {
        expect(screen.getByText('Schedule List')).toBeInTheDocument()
      })
    })
  })

  describe('WorkflowEditor', () => {
    it('should allow "new" as special case for creating workflows', () => {
      const TestComponent = () => {
        return (
          <QueryClientProvider client={queryClient}>
            <MemoryRouter initialEntries={['/workflows/new']}>
              <Routes>
                <Route path="/workflows/:id" element={<WorkflowEditor />} />
              </Routes>
            </MemoryRouter>
          </QueryClientProvider>
        )
      }

      render(<TestComponent />)

      // Should render the editor, not redirect
      expect(screen.getByText('Workflow Canvas')).toBeInTheDocument()
    })
  })
})
