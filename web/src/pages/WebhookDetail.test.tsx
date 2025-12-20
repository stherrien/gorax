import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import { MemoryRouter, Route, Routes } from 'react-router-dom'
import userEvent from '@testing-library/user-event'
import WebhookDetail from './WebhookDetail'
import { useWebhook, useWebhookEvents, useWebhookMutations } from '../hooks/useWebhooks'
import { useWorkflows } from '../hooks/useWorkflows'

vi.mock('../hooks/useWebhooks')
vi.mock('../hooks/useWorkflows')

const mockWebhook = {
  id: 'webhook-123',
  tenantId: 'tenant-1',
  workflowId: 'workflow-456',
  name: 'Test Webhook',
  path: '/api/test',
  authType: 'signature' as const,
  enabled: true,
  priority: 1,
  triggerCount: 42,
  lastTriggeredAt: '2024-01-15T10:30:00Z',
  createdAt: '2024-01-01T00:00:00Z',
  updatedAt: '2024-01-15T10:30:00Z',
  url: 'https://api.gorax.io/webhook/api/test',
}

const mockEvents = [
  {
    id: 'event-1',
    webhookId: 'webhook-123',
    executionId: 'exec-1',
    requestMethod: 'POST',
    requestHeaders: { 'Content-Type': 'application/json' },
    requestBody: { test: 'data' },
    responseStatus: 200,
    processingTimeMs: 150,
    status: 'processed' as const,
    createdAt: '2024-01-15T10:30:00Z',
  },
]

const mockWorkflows = [
  { id: 'workflow-456', name: 'My Workflow', status: 'active' },
]

const renderWithRouter = (webhookId: string = 'webhook-123') => {
  return render(
    <MemoryRouter initialEntries={[`/webhooks/${webhookId}`]}>
      <Routes>
        <Route path="/webhooks/:id" element={<WebhookDetail />} />
      </Routes>
    </MemoryRouter>
  )
}

describe('WebhookDetail', () => {
  const mockRefetch = vi.fn()
  const mockUpdateWebhook = vi.fn()
  const mockRegenerateSecret = vi.fn()
  const mockTestWebhook = vi.fn()

  beforeEach(() => {
    vi.clearAllMocks()

    vi.mocked(useWebhook).mockReturnValue({
      webhook: mockWebhook,
      loading: false,
      error: null,
      refetch: mockRefetch,
    })

    vi.mocked(useWebhookEvents).mockReturnValue({
      events: mockEvents,
      total: 1,
      loading: false,
      error: null,
      refetch: vi.fn(),
    })

    vi.mocked(useWebhookMutations).mockReturnValue({
      updateWebhook: mockUpdateWebhook,
      regenerateSecret: mockRegenerateSecret,
      testWebhook: mockTestWebhook,
      createWebhook: vi.fn(),
      deleteWebhook: vi.fn(),
      creating: false,
      updating: false,
      deleting: false,
      regenerating: false,
      testing: false,
    })

    vi.mocked(useWorkflows).mockReturnValue({
      workflows: mockWorkflows,
      loading: false,
      error: null,
      refetch: vi.fn(),
    } as any)
  })

  describe('Loading State', () => {
    it('should show loading state', () => {
      vi.mocked(useWebhook).mockReturnValue({
        webhook: null,
        loading: true,
        error: null,
        refetch: mockRefetch,
      })

      renderWithRouter()
      expect(screen.getByText(/loading/i)).toBeInTheDocument()
    })
  })

  describe('Error State', () => {
    it('should show error state', () => {
      vi.mocked(useWebhook).mockReturnValue({
        webhook: null,
        loading: false,
        error: new Error('Failed to fetch'),
        refetch: mockRefetch,
      })

      renderWithRouter()
      expect(screen.getByText(/Failed to fetch webhook/i)).toBeInTheDocument()
    })
  })

  describe('Webhook Details Display', () => {
    it('should display webhook name', () => {
      renderWithRouter()
      expect(screen.getByText('Test Webhook')).toBeInTheDocument()
    })

    it('should display webhook path', () => {
      renderWithRouter()
      expect(screen.getByText('/api/test')).toBeInTheDocument()
    })

    it('should display webhook URL', () => {
      renderWithRouter()
      expect(screen.getByText(/api\.gorax\.io/)).toBeInTheDocument()
    })

    it('should display auth type', () => {
      renderWithRouter()
      expect(screen.getByText('signature')).toBeInTheDocument()
    })

    it('should display enabled status', () => {
      renderWithRouter()
      expect(screen.getByRole('switch')).toBeInTheDocument()
    })

    it('should display trigger count', () => {
      renderWithRouter()
      expect(screen.getByText('42')).toBeInTheDocument()
    })

    it('should display linked workflow', () => {
      renderWithRouter()
      expect(screen.getByText('My Workflow')).toBeInTheDocument()
    })
  })

  describe('URL Copy Functionality', () => {
    it('should have copy URL button', () => {
      renderWithRouter()
      expect(screen.getByRole('button', { name: /copy/i })).toBeInTheDocument()
    })
  })

  describe('Secret Management', () => {
    it('should have regenerate secret button', () => {
      renderWithRouter()
      expect(screen.getByRole('button', { name: /regenerate secret/i })).toBeInTheDocument()
    })

    it('should show confirmation before regenerating', async () => {
      const user = userEvent.setup()
      renderWithRouter()

      await user.click(screen.getByRole('button', { name: /regenerate secret/i }))
      expect(screen.getByText(/are you sure/i)).toBeInTheDocument()
    })

    it('should call regenerateSecret on confirmation', async () => {
      const user = userEvent.setup()
      mockRegenerateSecret.mockResolvedValue({ secret: 'new-secret-123' })

      renderWithRouter()

      await user.click(screen.getByRole('button', { name: /regenerate secret/i }))
      await user.click(screen.getByRole('button', { name: /confirm/i }))

      await waitFor(() => {
        expect(mockRegenerateSecret).toHaveBeenCalledWith('webhook-123')
      })
    })
  })

  describe('Toggle Enabled', () => {
    it('should toggle enabled status', async () => {
      const user = userEvent.setup()
      mockUpdateWebhook.mockResolvedValue({ ...mockWebhook, enabled: false })

      renderWithRouter()

      const toggle = screen.getByRole('switch')
      await user.click(toggle)

      await waitFor(() => {
        expect(mockUpdateWebhook).toHaveBeenCalledWith('webhook-123', { enabled: false })
      })
    })
  })

  describe('Event History', () => {
    it('should display event history section', () => {
      renderWithRouter()
      expect(screen.getByText(/event history/i)).toBeInTheDocument()
    })

    it('should display events', () => {
      renderWithRouter()
      expect(screen.getByText('processed')).toBeInTheDocument()
    })
  })

  describe('Back Navigation', () => {
    it('should have back to list link', () => {
      renderWithRouter()
      expect(screen.getByRole('link', { name: /back/i })).toHaveAttribute('href', '/webhooks')
    })
  })
})
