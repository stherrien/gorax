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

  describe('Not Found State', () => {
    it('should show not found when webhook is null', () => {
      vi.mocked(useWebhook).mockReturnValue({
        webhook: null,
        loading: false,
        error: null,
        refetch: mockRefetch,
      })

      renderWithRouter()
      expect(screen.getByText(/webhook not found/i)).toBeInTheDocument()
    })
  })

  describe('Copy URL Functionality', () => {
    it('should copy URL to clipboard on success', async () => {
      const user = userEvent.setup()
      const writeTextMock = vi.fn().mockResolvedValue(undefined)
      vi.stubGlobal('navigator', {
        clipboard: { writeText: writeTextMock },
      })

      renderWithRouter()

      const copyButtons = screen.getAllByRole('button', { name: /copy/i })
      await user.click(copyButtons[0])

      await waitFor(() => {
        expect(writeTextMock).toHaveBeenCalledWith('https://api.gorax.io/webhook/api/test')
      })

      expect(screen.getByText('URL copied!')).toBeInTheDocument()

      vi.unstubAllGlobals()
    })

    it('should show error when copy fails', async () => {
      const user = userEvent.setup()
      const writeTextMock = vi.fn().mockRejectedValue(new Error('Copy failed'))
      vi.stubGlobal('navigator', {
        clipboard: { writeText: writeTextMock },
      })

      renderWithRouter()

      const copyButtons = screen.getAllByRole('button', { name: /copy/i })
      await user.click(copyButtons[0])

      await waitFor(() => {
        expect(screen.getByText('Failed to copy')).toBeInTheDocument()
      })

      vi.unstubAllGlobals()
    })
  })

  describe('Secret Regeneration', () => {
    it('should cancel regenerate confirmation', async () => {
      const user = userEvent.setup()
      renderWithRouter()

      await user.click(screen.getByRole('button', { name: /regenerate secret/i }))
      expect(screen.getByText(/are you sure/i)).toBeInTheDocument()

      await user.click(screen.getByRole('button', { name: /cancel/i }))
      expect(screen.queryByText(/are you sure/i)).not.toBeInTheDocument()
    })

    it('should show new secret after regeneration', async () => {
      const user = userEvent.setup()
      mockRegenerateSecret.mockResolvedValue({ secret: 'new-secret-abc123' })

      renderWithRouter()

      await user.click(screen.getByRole('button', { name: /regenerate secret/i }))
      await user.click(screen.getByRole('button', { name: /confirm/i }))

      await waitFor(() => {
        expect(screen.getByText('new-secret-abc123')).toBeInTheDocument()
      })

      expect(screen.getByText(/save this secret/i)).toBeInTheDocument()
    })

    it('should copy new secret to clipboard', async () => {
      const user = userEvent.setup()
      const writeTextMock = vi.fn().mockResolvedValue(undefined)
      vi.stubGlobal('navigator', {
        clipboard: { writeText: writeTextMock },
      })

      mockRegenerateSecret.mockResolvedValue({ secret: 'new-secret-xyz' })

      renderWithRouter()

      await user.click(screen.getByRole('button', { name: /regenerate secret/i }))
      await user.click(screen.getByRole('button', { name: /confirm/i }))

      await waitFor(() => {
        expect(screen.getByText('new-secret-xyz')).toBeInTheDocument()
      })

      // Find the copy button in the new secret section
      const copyButtons = screen.getAllByRole('button', { name: /copy/i })
      await user.click(copyButtons[copyButtons.length - 1])

      await waitFor(() => {
        expect(writeTextMock).toHaveBeenCalledWith('new-secret-xyz')
      })

      expect(screen.getByText('Secret copied!')).toBeInTheDocument()

      vi.unstubAllGlobals()
    })

    it('should dismiss new secret', async () => {
      const user = userEvent.setup()
      mockRegenerateSecret.mockResolvedValue({ secret: 'new-secret-dismiss' })

      renderWithRouter()

      await user.click(screen.getByRole('button', { name: /regenerate secret/i }))
      await user.click(screen.getByRole('button', { name: /confirm/i }))

      await waitFor(() => {
        expect(screen.getByText('new-secret-dismiss')).toBeInTheDocument()
      })

      await user.click(screen.getByRole('button', { name: /dismiss/i }))

      expect(screen.queryByText('new-secret-dismiss')).not.toBeInTheDocument()
    })

    it('should show error when regeneration fails', async () => {
      const user = userEvent.setup()
      mockRegenerateSecret.mockRejectedValue(new Error('Regeneration failed'))

      renderWithRouter()

      await user.click(screen.getByRole('button', { name: /regenerate secret/i }))
      await user.click(screen.getByRole('button', { name: /confirm/i }))

      await waitFor(() => {
        expect(screen.getByText('Regeneration failed')).toBeInTheDocument()
      })
    })

    it('should show Regenerating... while regenerating', () => {
      vi.mocked(useWebhookMutations).mockReturnValue({
        updateWebhook: mockUpdateWebhook,
        regenerateSecret: mockRegenerateSecret,
        testWebhook: mockTestWebhook,
        createWebhook: vi.fn(),
        deleteWebhook: vi.fn(),
        creating: false,
        updating: false,
        deleting: false,
        regenerating: true,
        testing: false,
      })

      renderWithRouter()

      expect(screen.getByRole('button', { name: /regenerating/i })).toBeInTheDocument()
      expect(screen.getByRole('button', { name: /regenerating/i })).toBeDisabled()
    })
  })

  describe('Toggle Error Handling', () => {
    it('should show error when toggle fails', async () => {
      const user = userEvent.setup()
      mockUpdateWebhook.mockRejectedValue(new Error('Update failed'))

      renderWithRouter()

      const toggle = screen.getByRole('switch')
      await user.click(toggle)

      await waitFor(() => {
        expect(screen.getByText('Update failed')).toBeInTheDocument()
      })
    })

    it('should show generic error for non-Error objects', async () => {
      const user = userEvent.setup()
      mockUpdateWebhook.mockRejectedValue('Unknown error')

      renderWithRouter()

      const toggle = screen.getByRole('switch')
      await user.click(toggle)

      await waitFor(() => {
        expect(screen.getByText('Failed to update')).toBeInTheDocument()
      })
    })

    it('should disable toggle while updating', () => {
      vi.mocked(useWebhookMutations).mockReturnValue({
        updateWebhook: mockUpdateWebhook,
        regenerateSecret: mockRegenerateSecret,
        testWebhook: mockTestWebhook,
        createWebhook: vi.fn(),
        deleteWebhook: vi.fn(),
        creating: false,
        updating: true,
        deleting: false,
        regenerating: false,
        testing: false,
      })

      renderWithRouter()

      const toggle = screen.getByRole('switch')
      expect(toggle).toBeDisabled()
    })
  })

  describe('Event History States', () => {
    it('should show loading state for events', () => {
      vi.mocked(useWebhookEvents).mockReturnValue({
        events: [],
        total: 0,
        loading: true,
        error: null,
        refetch: vi.fn(),
      })

      renderWithRouter()
      expect(screen.getByText(/loading events/i)).toBeInTheDocument()
    })

    it('should show empty state when no events', () => {
      vi.mocked(useWebhookEvents).mockReturnValue({
        events: [],
        total: 0,
        loading: false,
        error: null,
        refetch: vi.fn(),
      })

      renderWithRouter()
      expect(screen.getByText(/no events yet/i)).toBeInTheDocument()
    })

    it('should show event processing time', () => {
      renderWithRouter()
      expect(screen.getByText('150ms')).toBeInTheDocument()
    })

    it('should show dash when event has no processing time', () => {
      vi.mocked(useWebhookEvents).mockReturnValue({
        events: [{
          ...mockEvents[0],
          processingTimeMs: undefined,
        }],
        total: 1,
        loading: false,
        error: null,
        refetch: vi.fn(),
      })

      renderWithRouter()
      expect(screen.getByText('-')).toBeInTheDocument()
    })

    it('should display multiple event statuses', () => {
      vi.mocked(useWebhookEvents).mockReturnValue({
        events: [
          { ...mockEvents[0], id: 'event-1', status: 'processed' as const },
          { ...mockEvents[0], id: 'event-2', status: 'failed' as const },
          { ...mockEvents[0], id: 'event-3', status: 'filtered' as const },
        ],
        total: 3,
        loading: false,
        error: null,
        refetch: vi.fn(),
      })

      renderWithRouter()
      expect(screen.getByText('processed')).toBeInTheDocument()
      expect(screen.getByText('failed')).toBeInTheDocument()
      expect(screen.getByText('filtered')).toBeInTheDocument()
    })
  })

  describe('Unknown Workflow', () => {
    it('should show Unknown Workflow when workflow not found', () => {
      vi.mocked(useWorkflows).mockReturnValue({
        workflows: [],
        loading: false,
        error: null,
        refetch: vi.fn(),
      } as any)

      renderWithRouter()
      expect(screen.getByText('Unknown Workflow')).toBeInTheDocument()
    })
  })

  describe('Auth Types', () => {
    it('should show none auth type', () => {
      vi.mocked(useWebhook).mockReturnValue({
        webhook: { ...mockWebhook, authType: 'none' },
        loading: false,
        error: null,
        refetch: mockRefetch,
      })

      renderWithRouter()
      expect(screen.getByText('none')).toBeInTheDocument()
    })

    it('should show api_key auth type', () => {
      vi.mocked(useWebhook).mockReturnValue({
        webhook: { ...mockWebhook, authType: 'api_key' },
        loading: false,
        error: null,
        refetch: mockRefetch,
      })

      renderWithRouter()
      expect(screen.getByText('api_key')).toBeInTheDocument()
    })

    it('should not show regenerate button for non-signature auth', () => {
      vi.mocked(useWebhook).mockReturnValue({
        webhook: { ...mockWebhook, authType: 'none' },
        loading: false,
        error: null,
        refetch: mockRefetch,
      })

      renderWithRouter()
      expect(screen.queryByRole('button', { name: /regenerate secret/i })).not.toBeInTheDocument()
    })
  })

  describe('Priority Levels', () => {
    it('should show Low priority', () => {
      vi.mocked(useWebhook).mockReturnValue({
        webhook: { ...mockWebhook, priority: 0 },
        loading: false,
        error: null,
        refetch: mockRefetch,
      })

      renderWithRouter()
      expect(screen.getByText('Low')).toBeInTheDocument()
    })

    it('should show High priority', () => {
      vi.mocked(useWebhook).mockReturnValue({
        webhook: { ...mockWebhook, priority: 2 },
        loading: false,
        error: null,
        refetch: mockRefetch,
      })

      renderWithRouter()
      expect(screen.getByText('High')).toBeInTheDocument()
    })

    it('should show Critical priority', () => {
      vi.mocked(useWebhook).mockReturnValue({
        webhook: { ...mockWebhook, priority: 3 },
        loading: false,
        error: null,
        refetch: mockRefetch,
      })

      renderWithRouter()
      expect(screen.getByText('Critical')).toBeInTheDocument()
    })

    it('should cap priority at Critical for high values', () => {
      vi.mocked(useWebhook).mockReturnValue({
        webhook: { ...mockWebhook, priority: 10 },
        loading: false,
        error: null,
        refetch: mockRefetch,
      })

      renderWithRouter()
      expect(screen.getByText('Critical')).toBeInTheDocument()
    })
  })

  describe('Edit Link', () => {
    it('should have edit webhook link', () => {
      renderWithRouter()
      expect(screen.getByRole('link', { name: /edit webhook/i })).toHaveAttribute(
        'href',
        '/webhooks/webhook-123/edit'
      )
    })
  })

  describe('Statistics Display', () => {
    it('should display Never for lastTriggeredAt when not set', () => {
      vi.mocked(useWebhook).mockReturnValue({
        webhook: { ...mockWebhook, lastTriggeredAt: undefined },
        loading: false,
        error: null,
        refetch: mockRefetch,
      })

      renderWithRouter()
      expect(screen.getByText('Never')).toBeInTheDocument()
    })
  })

  describe('Error Details', () => {
    it('should display error message details', () => {
      vi.mocked(useWebhook).mockReturnValue({
        webhook: null,
        loading: false,
        error: new Error('Network timeout occurred'),
        refetch: mockRefetch,
      })

      renderWithRouter()
      expect(screen.getByText('Network timeout occurred')).toBeInTheDocument()
    })
  })
})
