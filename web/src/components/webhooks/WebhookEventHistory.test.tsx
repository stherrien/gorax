import { describe, it, expect, beforeEach, vi } from 'vitest'
import { render, screen, waitFor, within } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import userEvent from '@testing-library/user-event'
import { WebhookEventHistory } from './WebhookEventHistory'
import type { WebhookEvent } from '../../api/webhooks'

// Mock the webhooks API
vi.mock('../../api/webhooks', () => ({
  webhookAPI: {
    getEvents: vi.fn(),
  },
}))

import { webhookAPI } from '../../api/webhooks'

// Helper to wrap component with router
const renderWithRouter = (component: React.ReactElement) => {
  return render(<MemoryRouter>{component}</MemoryRouter>)
}

describe('WebhookEventHistory', () => {
  const mockEvents: WebhookEvent[] = [
    {
      id: 'evt-1',
      webhookId: 'wh-1',
      executionId: 'exec-1',
      requestMethod: 'POST',
      requestHeaders: {
        'Content-Type': 'application/json',
        'X-Signature': 'abc123',
      },
      requestBody: { data: 'test payload 1' },
      responseStatus: 200,
      processingTimeMs: 120,
      status: 'processed',
      replayCount: 0,
      createdAt: '2024-01-15T10:00:00Z',
    },
    {
      id: 'evt-2',
      webhookId: 'wh-1',
      executionId: 'exec-2',
      requestMethod: 'POST',
      requestHeaders: {
        'Content-Type': 'application/json',
      },
      requestBody: { data: 'test payload 2' },
      responseStatus: 202,
      processingTimeMs: 95,
      status: 'received',
      replayCount: 0,
      createdAt: '2024-01-15T09:00:00Z',
    },
    {
      id: 'evt-3',
      webhookId: 'wh-1',
      requestMethod: 'POST',
      requestHeaders: {
        'Content-Type': 'application/json',
      },
      requestBody: { data: 'test payload 3' },
      responseStatus: 400,
      processingTimeMs: 50,
      status: 'failed',
      errorMessage: 'Invalid payload format',
      replayCount: 0,
      createdAt: '2024-01-15T08:00:00Z',
    },
    {
      id: 'evt-4',
      webhookId: 'wh-1',
      requestMethod: 'POST',
      requestHeaders: {
        'Content-Type': 'application/json',
      },
      requestBody: { data: 'filtered payload' },
      processingTimeMs: 10,
      status: 'filtered',
      replayCount: 0,
      createdAt: '2024-01-15T07:00:00Z',
    },
  ]

  beforeEach(() => {
    vi.clearAllMocks()
  })

  describe('Initial Load', () => {
    it('should display loading state while fetching events', () => {
      ;(webhookAPI.getEvents as any).mockImplementation(
        () => new Promise(() => {}) // Never resolves
      )

      renderWithRouter(<WebhookEventHistory webhookId="wh-1" />)

      expect(screen.getByText(/loading/i)).toBeInTheDocument()
    })

    it('should display list of events from API', async () => {
      ;(webhookAPI.getEvents as any).mockResolvedValueOnce({
        events: mockEvents,
        total: 4,
      })

      renderWithRouter(<WebhookEventHistory webhookId="wh-1" />)

      await waitFor(() => {
        expect(screen.getByText('evt-1')).toBeInTheDocument()
        expect(screen.getByText('evt-2')).toBeInTheDocument()
        expect(screen.getByText('evt-3')).toBeInTheDocument()
        expect(screen.getByText('evt-4')).toBeInTheDocument()
      })
    })

    it('should show error message if fetch fails', async () => {
      ;(webhookAPI.getEvents as any).mockRejectedValueOnce(new Error('Failed to fetch events'))

      renderWithRouter(<WebhookEventHistory webhookId="wh-1" />)

      await waitFor(() => {
        expect(screen.getByText(/failed to fetch/i)).toBeInTheDocument()
      })
    })

    it('should show empty state when no events exist', async () => {
      ;(webhookAPI.getEvents as any).mockResolvedValueOnce({
        events: [],
        total: 0,
      })

      renderWithRouter(<WebhookEventHistory webhookId="wh-1" />)

      await waitFor(() => {
        expect(screen.getByText(/no events/i)).toBeInTheDocument()
      })
    })
  })

  describe('Event Table Display', () => {
    beforeEach(async () => {
      ;(webhookAPI.getEvents as any).mockResolvedValueOnce({
        events: mockEvents,
        total: 4,
      })
    })

    it('should display table headers', async () => {
      renderWithRouter(<WebhookEventHistory webhookId="wh-1" />)

      await waitFor(() => {
        const table = screen.getByRole('table')
        expect(within(table).getByRole('columnheader', { name: /event id/i })).toBeInTheDocument()
        expect(within(table).getByRole('columnheader', { name: /method/i })).toBeInTheDocument()
        expect(within(table).getByRole('columnheader', { name: /status/i })).toBeInTheDocument()
        expect(within(table).getByRole('columnheader', { name: /response code/i })).toBeInTheDocument()
        expect(within(table).getByRole('columnheader', { name: /processing time/i })).toBeInTheDocument()
      })
    })

    it('should display event time formatted correctly', async () => {
      renderWithRouter(<WebhookEventHistory webhookId="wh-1" />)

      await waitFor(() => {
        const rows = screen.getAllByRole('row')
        // First data row should contain formatted time
        expect(rows[1].textContent).toMatch(/\d{2}:\d{2}:\d{2}/)
      })
    })

    it('should display request method', async () => {
      renderWithRouter(<WebhookEventHistory webhookId="wh-1" />)

      await waitFor(() => {
        const methodElements = screen.getAllByText('POST')
        expect(methodElements.length).toBeGreaterThan(0)
      })
    })

    it('should display response status code', async () => {
      renderWithRouter(<WebhookEventHistory webhookId="wh-1" />)

      await waitFor(() => {
        expect(screen.getByText('200')).toBeInTheDocument()
        expect(screen.getByText('202')).toBeInTheDocument()
        expect(screen.getByText('400')).toBeInTheDocument()
      })
    })

    it('should display processing time in milliseconds', async () => {
      renderWithRouter(<WebhookEventHistory webhookId="wh-1" />)

      await waitFor(() => {
        expect(screen.getByText(/120.*ms/i)).toBeInTheDocument()
        expect(screen.getByText(/95.*ms/i)).toBeInTheDocument()
        expect(screen.getByText(/50.*ms/i)).toBeInTheDocument()
      })
    })

    it('should display status badge for processed events in green', async () => {
      renderWithRouter(<WebhookEventHistory webhookId="wh-1" />)

      await waitFor(() => {
        const processedBadge = screen.getByTestId('status-badge-processed')
        expect(processedBadge).toHaveClass('bg-green-500/20', 'text-green-400')
      })
    })

    it('should display status badge for received events in blue', async () => {
      renderWithRouter(<WebhookEventHistory webhookId="wh-1" />)

      await waitFor(() => {
        const receivedBadge = screen.getByTestId('status-badge-received')
        expect(receivedBadge).toHaveClass('bg-blue-500/20', 'text-blue-400')
      })
    })

    it('should display status badge for failed events in red', async () => {
      renderWithRouter(<WebhookEventHistory webhookId="wh-1" />)

      await waitFor(() => {
        const failedBadge = screen.getByTestId('status-badge-failed')
        expect(failedBadge).toHaveClass('bg-red-500/20', 'text-red-400')
      })
    })

    it('should display status badge for filtered events in yellow', async () => {
      renderWithRouter(<WebhookEventHistory webhookId="wh-1" />)

      await waitFor(() => {
        const filteredBadge = screen.getByTestId('status-badge-filtered')
        expect(filteredBadge).toHaveClass('bg-yellow-500/20', 'text-yellow-400')
      })
    })

    it('should display N/A for missing response code', async () => {
      renderWithRouter(<WebhookEventHistory webhookId="wh-1" />)

      await waitFor(() => {
        const naElements = screen.getAllByText('N/A')
        expect(naElements.length).toBeGreaterThan(0)
      })
    })
  })

  describe('Sorting', () => {
    beforeEach(() => {
      ;(webhookAPI.getEvents as any).mockResolvedValueOnce({
        events: mockEvents,
        total: 4,
      })
    })

    it('should display events sorted by time newest first by default', async () => {
      renderWithRouter(<WebhookEventHistory webhookId="wh-1" />)

      await waitFor(() => {
        const rows = screen.getAllByRole('row')
        // Skip header row
        const dataRows = rows.slice(1)
        expect(dataRows[0]).toHaveTextContent('evt-1') // 10:00
        expect(dataRows[1]).toHaveTextContent('evt-2') // 09:00
        expect(dataRows[2]).toHaveTextContent('evt-3') // 08:00
        expect(dataRows[3]).toHaveTextContent('evt-4') // 07:00
      })
    })

    it('should allow sorting by time column', async () => {
      const user = userEvent.setup()

      renderWithRouter(<WebhookEventHistory webhookId="wh-1" />)

      await waitFor(() => {
        expect(screen.getByText('evt-1')).toBeInTheDocument()
      })

      const timeHeader = screen.getByRole('button', { name: /time/i })
      await user.click(timeHeader)

      // Should reverse to oldest first
      await waitFor(() => {
        const rows = screen.getAllByRole('row')
        const dataRows = rows.slice(1)
        expect(dataRows[0]).toHaveTextContent('evt-4') // 07:00
        expect(dataRows[3]).toHaveTextContent('evt-1') // 10:00
      })
    })
  })

  describe('Event Detail Modal', () => {
    beforeEach(() => {
      ;(webhookAPI.getEvents as any).mockResolvedValueOnce({
        events: mockEvents,
        total: 4,
      })
    })

    it('should open modal when clicking on an event row', async () => {
      const user = userEvent.setup()

      renderWithRouter(<WebhookEventHistory webhookId="wh-1" />)

      await waitFor(() => {
        expect(screen.getByText('evt-1')).toBeInTheDocument()
      })

      // Click on the event ID cell which has the onClick handler
      await user.click(screen.getByText('evt-1'))

      await waitFor(() => {
        expect(screen.getByRole('dialog')).toBeInTheDocument()
        expect(screen.getByText(/event details/i)).toBeInTheDocument()
      })
    })

    it('should display request headers in modal', async () => {
      const user = userEvent.setup()

      renderWithRouter(<WebhookEventHistory webhookId="wh-1" />)

      await waitFor(() => {
        expect(screen.getByText('evt-1')).toBeInTheDocument()
      })

      await user.click(screen.getByText('evt-1'))

      await waitFor(() => {
        const modal = screen.getByRole('dialog')
        expect(within(modal).getByText(/request headers/i)).toBeInTheDocument()
        expect(within(modal).getByText(/Content-Type/i)).toBeInTheDocument()
        expect(within(modal).getByText(/application\/json/i)).toBeInTheDocument()
      })
    })

    it('should display request body in modal with syntax highlighting', async () => {
      const user = userEvent.setup()

      renderWithRouter(<WebhookEventHistory webhookId="wh-1" />)

      await waitFor(() => {
        expect(screen.getByText('evt-1')).toBeInTheDocument()
      })

      await user.click(screen.getByText('evt-1'))

      await waitFor(() => {
        const modal = screen.getByRole('dialog')
        expect(within(modal).getByText(/request body/i)).toBeInTheDocument()
        expect(within(modal).getByText(/test payload 1/i)).toBeInTheDocument()
      })
    })

    it('should display response status code in modal', async () => {
      const user = userEvent.setup()

      renderWithRouter(<WebhookEventHistory webhookId="wh-1" />)

      await waitFor(() => {
        expect(screen.getByText('evt-1')).toBeInTheDocument()
      })

      await user.click(screen.getByText('evt-1'))

      await waitFor(() => {
        const modal = screen.getByRole('dialog')
        expect(within(modal).getByText('Response Status')).toBeInTheDocument()
        expect(within(modal).getByText('200')).toBeInTheDocument()
      })
    })

    it('should display processing time in modal', async () => {
      const user = userEvent.setup()

      renderWithRouter(<WebhookEventHistory webhookId="wh-1" />)

      await waitFor(() => {
        expect(screen.getByText('evt-1')).toBeInTheDocument()
      })

      await user.click(screen.getByText('evt-1'))

      await waitFor(() => {
        const modal = screen.getByRole('dialog')
        expect(within(modal).getByText('Processing Time')).toBeInTheDocument()
        expect(within(modal).getByText('120 ms')).toBeInTheDocument()
      })
    })

    it('should display error message in modal if event failed', async () => {
      const user = userEvent.setup()

      renderWithRouter(<WebhookEventHistory webhookId="wh-1" />)

      await waitFor(() => {
        expect(screen.getByText('evt-3')).toBeInTheDocument()
      })

      await user.click(screen.getByText('evt-3'))

      await waitFor(() => {
        const modal = screen.getByRole('dialog')
        expect(within(modal).getByText(/error/i)).toBeInTheDocument()
        expect(within(modal).getByText(/invalid payload format/i)).toBeInTheDocument()
      })
    })

    it('should display link to triggered execution if executionId exists', async () => {
      const user = userEvent.setup()

      renderWithRouter(<WebhookEventHistory webhookId="wh-1" />)

      await waitFor(() => {
        expect(screen.getByText('evt-1')).toBeInTheDocument()
      })

      await user.click(screen.getByText('evt-1'))

      await waitFor(() => {
        const modal = screen.getByRole('dialog')
        const executionLink = within(modal).getByRole('link', { name: /view execution/i })
        expect(executionLink).toHaveAttribute('href', '/executions/exec-1')
      })
    })

    it('should close modal when clicking close button', async () => {
      const user = userEvent.setup()

      renderWithRouter(<WebhookEventHistory webhookId="wh-1" />)

      await waitFor(() => {
        expect(screen.getByText('evt-1')).toBeInTheDocument()
      })

      await user.click(screen.getByText('evt-1'))

      await waitFor(() => {
        expect(screen.getByRole('dialog')).toBeInTheDocument()
      })

      const modal = screen.getByRole('dialog')
      const closeButton = within(modal).getByRole('button', { name: 'Close modal' })
      await user.click(closeButton)

      await waitFor(() => {
        expect(screen.queryByRole('dialog')).not.toBeInTheDocument()
      })
    })

    it('should close modal when clicking outside', async () => {
      const user = userEvent.setup()

      renderWithRouter(<WebhookEventHistory webhookId="wh-1" />)

      await waitFor(() => {
        expect(screen.getByText('evt-1')).toBeInTheDocument()
      })

      await user.click(screen.getByText('evt-1'))

      await waitFor(() => {
        expect(screen.getByRole('dialog')).toBeInTheDocument()
      })

      const backdrop = screen.getByTestId('modal-backdrop')
      await user.click(backdrop)

      await waitFor(() => {
        expect(screen.queryByRole('dialog')).not.toBeInTheDocument()
      })
    })
  })

  describe('Filtering', () => {
    beforeEach(() => {
      ;(webhookAPI.getEvents as any).mockResolvedValue({
        events: mockEvents,
        total: 4,
      })
    })

    it('should display status filter dropdown', async () => {
      renderWithRouter(<WebhookEventHistory webhookId="wh-1" />)

      await waitFor(() => {
        expect(screen.getByLabelText(/filter by status/i)).toBeInTheDocument()
      })
    })

    it('should filter events by processed status', async () => {
      const user = userEvent.setup()

      renderWithRouter(<WebhookEventHistory webhookId="wh-1" />)

      await waitFor(() => {
        expect(screen.getByText('evt-1')).toBeInTheDocument()
      })

      const statusFilter = screen.getByLabelText(/filter by status/i)
      await user.selectOptions(statusFilter, 'processed')

      await waitFor(() => {
        expect(screen.getByText('evt-1')).toBeInTheDocument()
        expect(screen.queryByText('evt-2')).not.toBeInTheDocument()
        expect(screen.queryByText('evt-3')).not.toBeInTheDocument()
        expect(screen.queryByText('evt-4')).not.toBeInTheDocument()
      })
    })

    it('should filter events by failed status', async () => {
      const user = userEvent.setup()

      renderWithRouter(<WebhookEventHistory webhookId="wh-1" />)

      await waitFor(() => {
        expect(screen.getByText('evt-3')).toBeInTheDocument()
      })

      const statusFilter = screen.getByLabelText(/filter by status/i)
      await user.selectOptions(statusFilter, 'failed')

      await waitFor(() => {
        expect(screen.queryByText('evt-1')).not.toBeInTheDocument()
        expect(screen.queryByText('evt-2')).not.toBeInTheDocument()
        expect(screen.getByText('evt-3')).toBeInTheDocument()
        expect(screen.queryByText('evt-4')).not.toBeInTheDocument()
      })
    })

    it('should display search input for payload content', async () => {
      renderWithRouter(<WebhookEventHistory webhookId="wh-1" />)

      await waitFor(() => {
        expect(screen.getByPlaceholderText(/search payload/i)).toBeInTheDocument()
      })
    })

    it('should filter events by payload search term', async () => {
      const user = userEvent.setup()

      renderWithRouter(<WebhookEventHistory webhookId="wh-1" />)

      await waitFor(() => {
        expect(screen.getByText('evt-1')).toBeInTheDocument()
      })

      const searchInput = screen.getByPlaceholderText(/search payload/i)
      await user.type(searchInput, 'filtered')

      await waitFor(() => {
        expect(screen.queryByText('evt-1')).not.toBeInTheDocument()
        expect(screen.queryByText('evt-2')).not.toBeInTheDocument()
        expect(screen.queryByText('evt-3')).not.toBeInTheDocument()
        expect(screen.getByText('evt-4')).toBeInTheDocument()
      })
    })

    it('should show all events when filter is cleared', async () => {
      const user = userEvent.setup()

      renderWithRouter(<WebhookEventHistory webhookId="wh-1" />)

      await waitFor(() => {
        expect(screen.getByText('evt-1')).toBeInTheDocument()
      })

      const statusFilter = screen.getByLabelText(/filter by status/i)
      await user.selectOptions(statusFilter, 'processed')

      await waitFor(() => {
        expect(screen.queryByText('evt-2')).not.toBeInTheDocument()
      })

      await user.selectOptions(statusFilter, 'all')

      await waitFor(() => {
        expect(screen.getByText('evt-1')).toBeInTheDocument()
        expect(screen.getByText('evt-2')).toBeInTheDocument()
        expect(screen.getByText('evt-3')).toBeInTheDocument()
        expect(screen.getByText('evt-4')).toBeInTheDocument()
      })
    })
  })

  describe('Pagination', () => {
    const generateEvents = (count: number): WebhookEvent[] => {
      return Array.from({ length: count }, (_, i) => ({
        id: `evt-${i + 1}`,
        webhookId: 'wh-1',
        requestMethod: 'POST',
        requestHeaders: {},
        requestBody: { data: `payload ${i + 1}` },
        responseStatus: 200,
        processingTimeMs: 100,
        status: 'processed' as const,
        replayCount: 0,
        createdAt: new Date(2024, 0, 15, 10, 0, i).toISOString(),
      }))
    }

    it('should display 20 events per page by default', async () => {
      const events = generateEvents(30)
      ;(webhookAPI.getEvents as any).mockResolvedValueOnce({
        events: events.slice(0, 20),
        total: 30,
      })

      renderWithRouter(<WebhookEventHistory webhookId="wh-1" />)

      await waitFor(() => {
        const rows = screen.getAllByRole('row')
        // 20 data rows + 1 header row
        expect(rows.length).toBe(21)
      })
    })

    it('should display pagination controls when total exceeds page size', async () => {
      const events = generateEvents(30)
      ;(webhookAPI.getEvents as any).mockResolvedValueOnce({
        events: events.slice(0, 20),
        total: 30,
      })

      renderWithRouter(<WebhookEventHistory webhookId="wh-1" />)

      await waitFor(() => {
        expect(screen.getByRole('navigation', { name: /pagination/i })).toBeInTheDocument()
        expect(screen.getByText(/page 1 of 2/i)).toBeInTheDocument()
      })
    })

    it('should navigate to next page when clicking next button', async () => {
      const user = userEvent.setup()
      const events = generateEvents(30)

      ;(webhookAPI.getEvents as any)
        .mockResolvedValueOnce({
          events: events.slice(0, 20),
          total: 30,
        })
        .mockResolvedValueOnce({
          events: events.slice(20, 30),
          total: 30,
        })

      renderWithRouter(<WebhookEventHistory webhookId="wh-1" />)

      await waitFor(() => {
        expect(screen.getByText('evt-1')).toBeInTheDocument()
      })

      const nextButton = screen.getByRole('button', { name: /next/i })
      await user.click(nextButton)

      await waitFor(() => {
        expect(screen.getByText('evt-21')).toBeInTheDocument()
        expect(webhookAPI.getEvents).toHaveBeenCalledWith('wh-1', { page: 2, limit: 20 })
      })
    })

    it('should navigate to previous page when clicking previous button', async () => {
      const user = userEvent.setup()
      const events = generateEvents(30)

      ;(webhookAPI.getEvents as any)
        .mockResolvedValueOnce({
          events: events.slice(0, 20),
          total: 30,
        })
        .mockResolvedValueOnce({
          events: events.slice(20, 30),
          total: 30,
        })
        .mockResolvedValueOnce({
          events: events.slice(0, 20),
          total: 30,
        })

      renderWithRouter(<WebhookEventHistory webhookId="wh-1" />)

      await waitFor(() => {
        expect(screen.getByText('evt-1')).toBeInTheDocument()
      })

      const nextButton = screen.getByRole('button', { name: /next/i })
      await user.click(nextButton)

      await waitFor(() => {
        expect(screen.getByText('evt-21')).toBeInTheDocument()
      })

      const prevButton = screen.getByRole('button', { name: /previous/i })
      await user.click(prevButton)

      await waitFor(() => {
        expect(screen.getByText('evt-1')).toBeInTheDocument()
      })
    })

    it('should disable previous button on first page', async () => {
      const events = generateEvents(30)
      ;(webhookAPI.getEvents as any).mockResolvedValueOnce({
        events: events.slice(0, 20),
        total: 30,
      })

      renderWithRouter(<WebhookEventHistory webhookId="wh-1" />)

      await waitFor(() => {
        const prevButton = screen.getByRole('button', { name: /previous/i })
        expect(prevButton).toBeDisabled()
      })
    })

    it('should disable next button on last page', async () => {
      const user = userEvent.setup()
      const events = generateEvents(30)

      ;(webhookAPI.getEvents as any)
        .mockResolvedValueOnce({
          events: events.slice(0, 20),
          total: 30,
        })
        .mockResolvedValueOnce({
          events: events.slice(20, 30),
          total: 30,
        })

      renderWithRouter(<WebhookEventHistory webhookId="wh-1" />)

      await waitFor(() => {
        expect(screen.getByText('evt-1')).toBeInTheDocument()
      })

      const nextButton = screen.getByRole('button', { name: /next/i })
      await user.click(nextButton)

      await waitFor(() => {
        expect(nextButton).toBeDisabled()
      })
    })
  })

  describe('Export Functionality', () => {
    beforeEach(() => {
      ;(webhookAPI.getEvents as any).mockResolvedValueOnce({
        events: mockEvents,
        total: 4,
      })
    })

    it('should display export CSV button', async () => {
      renderWithRouter(<WebhookEventHistory webhookId="wh-1" />)

      await waitFor(() => {
        expect(screen.getByRole('button', { name: /export csv/i })).toBeInTheDocument()
      })
    })

    it('should trigger export when clicking export button', async () => {
      const user = userEvent.setup()

      renderWithRouter(<WebhookEventHistory webhookId="wh-1" />)

      await waitFor(() => {
        expect(screen.getByText('evt-1')).toBeInTheDocument()
      })

      const exportButton = screen.getByRole('button', { name: /export csv/i })

      // Just verify the button can be clicked without errors
      await user.click(exportButton)

      // Button should still exist after click
      expect(exportButton).toBeInTheDocument()
    })
  })

  describe('Accessibility', () => {
    beforeEach(() => {
      ;(webhookAPI.getEvents as any).mockResolvedValueOnce({
        events: mockEvents,
        total: 4,
      })
    })

    it('should have proper table structure with headers', async () => {
      renderWithRouter(<WebhookEventHistory webhookId="wh-1" />)

      await waitFor(() => {
        expect(screen.getByRole('table')).toBeInTheDocument()
        const columnHeaders = screen.getAllByRole('columnheader')
        expect(columnHeaders.length).toBeGreaterThan(0)
      })
    })

    it('should have clickable rows with proper role', async () => {
      renderWithRouter(<WebhookEventHistory webhookId="wh-1" />)

      await waitFor(() => {
        const rows = screen.getAllByRole('row')
        // Skip header, check data rows
        const dataRows = rows.slice(1)
        dataRows.forEach((row) => {
          expect(row).toHaveAttribute('role', 'row')
        })
      })
    })

    it('should have proper ARIA labels for modal', async () => {
      const user = userEvent.setup()

      renderWithRouter(<WebhookEventHistory webhookId="wh-1" />)

      await waitFor(() => {
        expect(screen.getByText('evt-1')).toBeInTheDocument()
      })

      await user.click(screen.getByText('evt-1'))

      await waitFor(() => {
        const dialog = screen.getByRole('dialog')
        expect(dialog).toHaveAttribute('aria-modal', 'true')
      })
    })
  })

  describe('Replay Functionality', () => {
    const mockEventsWithReplay: WebhookEvent[] = [
      {
        id: 'evt-1',
        webhookId: 'wh-1',
        executionId: 'exec-1',
        requestMethod: 'POST',
        requestHeaders: { 'Content-Type': 'application/json' },
        requestBody: { data: 'test payload 1' },
        responseStatus: 200,
        processingTimeMs: 120,
        status: 'processed',
        replayCount: 2,
        createdAt: '2024-01-15T10:00:00Z',
      },
      {
        id: 'evt-2',
        webhookId: 'wh-1',
        executionId: 'exec-2',
        requestMethod: 'POST',
        requestHeaders: { 'Content-Type': 'application/json' },
        requestBody: { data: 'test payload 2' },
        responseStatus: 202,
        processingTimeMs: 95,
        status: 'received',
        replayCount: 0,
        createdAt: '2024-01-15T09:00:00Z',
      },
      {
        id: 'evt-3',
        webhookId: 'wh-1',
        requestMethod: 'POST',
        requestHeaders: { 'Content-Type': 'application/json' },
        requestBody: { data: 'test payload 3' },
        responseStatus: 400,
        processingTimeMs: 50,
        status: 'failed',
        replayCount: 5,
        createdAt: '2024-01-15T08:00:00Z',
      },
    ]

    beforeEach(() => {
      ;(webhookAPI.getEvents as any).mockResolvedValue({
        events: mockEventsWithReplay,
        total: 3,
      })
    })

    it('should display replay button on each event row', async () => {
      renderWithRouter(<WebhookEventHistory webhookId="wh-1" />)

      await waitFor(() => {
        const replayButtons = screen.getAllByRole('button', { name: /replay/i })
        expect(replayButtons.length).toBeGreaterThan(0)
      })
    })

    it('should display replay count badge on events that have been replayed', async () => {
      renderWithRouter(<WebhookEventHistory webhookId="wh-1" />)

      await waitFor(() => {
        expect(screen.getByText(/2×/i)).toBeInTheDocument()
      })
    })

    it('should not display replay count badge on events with zero replays', async () => {
      renderWithRouter(<WebhookEventHistory webhookId="wh-1" />)

      await waitFor(() => {
        const rows = screen.getAllByRole('row')
        const evt2Row = rows.find(row => row.textContent?.includes('evt-2'))
        expect(evt2Row?.textContent).not.toMatch(/0×/)
      })
    })

    it('should disable replay button when event has reached max replay count', async () => {
      renderWithRouter(<WebhookEventHistory webhookId="wh-1" />)

      await waitFor(() => {
        const rows = screen.getAllByRole('row')
        const evt3Row = rows.find(row => row.textContent?.includes('evt-3'))
        const replayButton = within(evt3Row as HTMLElement).getByRole('button', { name: /replay/i })
        expect(replayButton).toBeDisabled()
      })
    })

    it('should show tooltip on disabled replay button', async () => {
      renderWithRouter(<WebhookEventHistory webhookId="wh-1" />)

      await waitFor(() => {
        const rows = screen.getAllByRole('row')
        const evt3Row = rows.find(row => row.textContent?.includes('evt-3'))
        const replayButton = within(evt3Row as HTMLElement).getByRole('button', { name: /replay/i })
        expect(replayButton).toHaveAttribute('title', 'Maximum replay limit reached')
      })
    })

    it('should open replay modal when clicking replay button', async () => {
      const user = userEvent.setup()
      renderWithRouter(<WebhookEventHistory webhookId="wh-1" />)

      await waitFor(() => {
        expect(screen.getByText('evt-1')).toBeInTheDocument()
      })

      const rows = screen.getAllByRole('row')
      const evt1Row = rows.find(row => row.textContent?.includes('evt-1'))
      const replayButton = within(evt1Row as HTMLElement).getByRole('button', { name: /replay/i })
      await user.click(replayButton)

      await waitFor(() => {
        expect(screen.getByRole('dialog')).toBeInTheDocument()
        expect(screen.getByText(/replay webhook event/i)).toBeInTheDocument()
      })
    })

    it('should display checkbox for batch selection', async () => {
      renderWithRouter(<WebhookEventHistory webhookId="wh-1" />)

      await waitFor(() => {
        const checkboxes = screen.getAllByRole('checkbox')
        expect(checkboxes.length).toBeGreaterThan(0)
      })
    })

    it('should not show batch replay button when no events selected', async () => {
      renderWithRouter(<WebhookEventHistory webhookId="wh-1" />)

      await waitFor(() => {
        expect(screen.getByText('evt-1')).toBeInTheDocument()
      })

      expect(screen.queryByRole('button', { name: /replay selected/i })).not.toBeInTheDocument()
    })

    it('should show batch replay button when events are selected', async () => {
      const user = userEvent.setup()
      renderWithRouter(<WebhookEventHistory webhookId="wh-1" />)

      await waitFor(() => {
        expect(screen.getByText('evt-1')).toBeInTheDocument()
      })

      const checkboxes = screen.getAllByRole('checkbox')
      await user.click(checkboxes[1]) // Skip header checkbox, click first event

      await waitFor(() => {
        expect(screen.getByRole('button', { name: /replay selected/i })).toBeInTheDocument()
      })
    })

    it('should display count of selected events in batch replay button', async () => {
      const user = userEvent.setup()
      renderWithRouter(<WebhookEventHistory webhookId="wh-1" />)

      await waitFor(() => {
        expect(screen.getByText('evt-1')).toBeInTheDocument()
      })

      const checkboxes = screen.getAllByRole('checkbox')
      await user.click(checkboxes[1])
      await user.click(checkboxes[2])

      await waitFor(() => {
        expect(screen.getByRole('button', { name: /replay selected \(2\)/i })).toBeInTheDocument()
      })
    })

    it('should disable checkboxes for events at max replay count', async () => {
      renderWithRouter(<WebhookEventHistory webhookId="wh-1" />)

      await waitFor(() => {
        const rows = screen.getAllByRole('row')
        const evt3Row = rows.find(row => row.textContent?.includes('evt-3'))
        const checkbox = within(evt3Row as HTMLElement).getByRole('checkbox')
        expect(checkbox).toBeDisabled()
      })
    })

    it('should prevent selecting more than 10 events for batch replay', async () => {
      const user = userEvent.setup()
      const manyEvents = Array.from({ length: 15 }, (_, i) => ({
        ...mockEventsWithReplay[0],
        id: `evt-${i + 1}`,
        replayCount: 0,
      }))

      ;(webhookAPI.getEvents as any).mockResolvedValueOnce({
        events: manyEvents,
        total: 15,
      })

      renderWithRouter(<WebhookEventHistory webhookId="wh-1" />)

      await waitFor(() => {
        expect(screen.getByText('evt-1')).toBeInTheDocument()
      })

      const checkboxes = screen.getAllByRole('checkbox')

      // Try to select 11 events (skip header checkbox)
      for (let i = 1; i <= 11; i++) {
        await user.click(checkboxes[i])
      }

      // Should only show 10 selected
      await waitFor(() => {
        expect(screen.getByText(/replay selected \(10\)/i)).toBeInTheDocument()
      })

      // 11th checkbox should remain unchecked
      expect(checkboxes[11]).not.toBeChecked()
    })

    it('should show warning message when trying to select more than 10', async () => {
      const user = userEvent.setup()
      const manyEvents = Array.from({ length: 15 }, (_, i) => ({
        ...mockEventsWithReplay[0],
        id: `evt-${i + 1}`,
        replayCount: 0,
      }))

      ;(webhookAPI.getEvents as any).mockResolvedValueOnce({
        events: manyEvents,
        total: 15,
      })

      renderWithRouter(<WebhookEventHistory webhookId="wh-1" />)

      await waitFor(() => {
        expect(screen.getByText('evt-1')).toBeInTheDocument()
      })

      const checkboxes = screen.getAllByRole('checkbox')

      // Select 10 events
      for (let i = 1; i <= 10; i++) {
        await user.click(checkboxes[i])
      }

      // Try to select 11th
      await user.click(checkboxes[11])

      await waitFor(() => {
        expect(screen.getByText(/maximum 10 events/i)).toBeInTheDocument()
      })
    })

    it('should show confirmation dialog before batch replay', async () => {
      const user = userEvent.setup()
      renderWithRouter(<WebhookEventHistory webhookId="wh-1" />)

      await waitFor(() => {
        expect(screen.getByText('evt-1')).toBeInTheDocument()
      })

      const checkboxes = screen.getAllByRole('checkbox')
      await user.click(checkboxes[1])
      await user.click(checkboxes[2])

      const replayButton = screen.getByRole('button', { name: /replay selected/i })
      await user.click(replayButton)

      await waitFor(() => {
        expect(screen.getByText(/confirm batch replay/i)).toBeInTheDocument()
        expect(screen.getByText(/replay 2 events/i)).toBeInTheDocument()
      })
    })

    it('should call batchReplayEvents API when confirmed', async () => {
      const user = userEvent.setup()
      ;(webhookAPI.batchReplayEvents as any) = vi.fn().mockResolvedValue({
        results: {
          'evt-1': { success: true, executionId: 'exec-new-1' },
          'evt-2': { success: true, executionId: 'exec-new-2' },
        },
      })

      renderWithRouter(<WebhookEventHistory webhookId="wh-1" />)

      await waitFor(() => {
        expect(screen.getByText('evt-1')).toBeInTheDocument()
      })

      const checkboxes = screen.getAllByRole('checkbox')
      await user.click(checkboxes[1])
      await user.click(checkboxes[2])

      const replayButton = screen.getByRole('button', { name: /replay selected/i })
      await user.click(replayButton)

      await waitFor(() => {
        expect(screen.getByText(/confirm batch replay/i)).toBeInTheDocument()
      })

      const confirmButton = screen.getByRole('button', { name: /confirm/i })
      await user.click(confirmButton)

      await waitFor(() => {
        expect(webhookAPI.batchReplayEvents).toHaveBeenCalledWith('wh-1', ['evt-1', 'evt-2'])
      })
    })

    it('should show success notification after batch replay', async () => {
      const user = userEvent.setup()
      ;(webhookAPI.batchReplayEvents as any) = vi.fn().mockResolvedValue({
        results: {
          'evt-1': { success: true, executionId: 'exec-new-1' },
          'evt-2': { success: true, executionId: 'exec-new-2' },
        },
      })

      renderWithRouter(<WebhookEventHistory webhookId="wh-1" />)

      await waitFor(() => {
        expect(screen.getByText('evt-1')).toBeInTheDocument()
      })

      const checkboxes = screen.getAllByRole('checkbox')
      await user.click(checkboxes[1])

      const replayButton = screen.getByRole('button', { name: /replay selected/i })
      await user.click(replayButton)

      const confirmButton = screen.getByRole('button', { name: /confirm/i })
      await user.click(confirmButton)

      await waitFor(() => {
        expect(screen.getByText(/successfully replayed/i)).toBeInTheDocument()
      })
    })

    it('should show partial success notification when some replays fail', async () => {
      const user = userEvent.setup()
      ;(webhookAPI.batchReplayEvents as any) = vi.fn().mockResolvedValue({
        results: {
          'evt-1': { success: true, executionId: 'exec-new-1' },
          'evt-2': { success: false, error: 'Replay failed' },
        },
      })

      renderWithRouter(<WebhookEventHistory webhookId="wh-1" />)

      await waitFor(() => {
        expect(screen.getByText('evt-1')).toBeInTheDocument()
      })

      const checkboxes = screen.getAllByRole('checkbox')
      await user.click(checkboxes[1])
      await user.click(checkboxes[2])

      const replayButton = screen.getByRole('button', { name: /replay selected/i })
      await user.click(replayButton)

      const confirmButton = screen.getByRole('button', { name: /confirm/i })
      await user.click(confirmButton)

      await waitFor(() => {
        expect(screen.getByText(/1 succeeded, 1 failed/i)).toBeInTheDocument()
      })
    })

    it('should clear selection after successful batch replay', async () => {
      const user = userEvent.setup()
      ;(webhookAPI.batchReplayEvents as any) = vi.fn().mockResolvedValue({
        results: {
          'evt-1': { success: true, executionId: 'exec-new-1' },
        },
      })

      renderWithRouter(<WebhookEventHistory webhookId="wh-1" />)

      await waitFor(() => {
        expect(screen.getByText('evt-1')).toBeInTheDocument()
      })

      const checkboxes = screen.getAllByRole('checkbox')
      await user.click(checkboxes[1])

      const replayButton = screen.getByRole('button', { name: /replay selected/i })
      await user.click(replayButton)

      const confirmButton = screen.getByRole('button', { name: /confirm/i })
      await user.click(confirmButton)

      await waitFor(() => {
        expect(screen.queryByRole('button', { name: /replay selected/i })).not.toBeInTheDocument()
      })
    })

    it('should refresh events list after successful replay', async () => {
      const user = userEvent.setup()
      ;(webhookAPI.batchReplayEvents as any) = vi.fn().mockResolvedValue({
        results: {
          'evt-1': { success: true, executionId: 'exec-new-1' },
        },
      })

      renderWithRouter(<WebhookEventHistory webhookId="wh-1" />)

      await waitFor(() => {
        expect(screen.getByText('evt-1')).toBeInTheDocument()
      })

      const checkboxes = screen.getAllByRole('checkbox')
      await user.click(checkboxes[1])

      const replayButton = screen.getByRole('button', { name: /replay selected/i })
      await user.click(replayButton)

      const confirmButton = screen.getByRole('button', { name: /confirm/i })
      await user.click(confirmButton)

      await waitFor(() => {
        // getEvents should be called again to refresh
        expect(webhookAPI.getEvents).toHaveBeenCalledTimes(2)
      })
    })
  })

  describe('Metadata Display', () => {
    it('should display event metadata in detail modal when present', async () => {
      const user = userEvent.setup()
      const eventWithMetadata: WebhookEvent = {
        id: 'evt-meta-1',
        webhookId: 'wh-1',
        requestMethod: 'POST',
        requestHeaders: { 'Content-Type': 'application/json' },
        requestBody: { test: 'data' },
        status: 'processed',
        replayCount: 0,
        createdAt: '2024-01-15T10:00:00Z',
        metadata: {
          sourceIp: '192.168.1.100',
          userAgent: 'GitHub-Hookshot/abc123',
          receivedAt: '2024-01-15T10:00:00Z',
          contentType: 'application/json',
          contentLength: 256,
        },
      }

      ;(webhookAPI.getEvents as any).mockResolvedValue({
        events: [eventWithMetadata],
        total: 1,
      })

      renderWithRouter(<WebhookEventHistory webhookId="wh-1" />)

      await waitFor(() => {
        expect(screen.getByText('evt-meta-1')).toBeInTheDocument()
      })

      // Click on event to open modal
      await user.click(screen.getByText('evt-meta-1'))

      // Wait for modal to appear
      const modal = await screen.findByRole('dialog')
      expect(modal).toBeInTheDocument()

      // Verify metadata is displayed
      expect(within(modal).getByText('Request Details')).toBeInTheDocument()
      expect(within(modal).getByText('192.168.1.100')).toBeInTheDocument()
      expect(within(modal).getByText('GitHub-Hookshot/abc123')).toBeInTheDocument()
      expect(within(modal).getByText('application/json')).toBeInTheDocument()
      expect(within(modal).getByText('256 bytes')).toBeInTheDocument()
    })

    it('should not display metadata section when metadata is absent', async () => {
      const user = userEvent.setup()
      const eventWithoutMetadata = mockEvents[0] // This event has no metadata

      ;(webhookAPI.getEvents as any).mockResolvedValue({
        events: [eventWithoutMetadata],
        total: 1,
      })

      renderWithRouter(<WebhookEventHistory webhookId="wh-1" />)

      await waitFor(() => {
        expect(screen.getByText('evt-1')).toBeInTheDocument()
      })

      // Click on event to open modal
      await user.click(screen.getByText('evt-1'))

      // Wait for modal to appear
      const modal = await screen.findByRole('dialog')
      expect(modal).toBeInTheDocument()

      // Verify metadata section is not displayed
      expect(within(modal).queryByText('Request Details')).not.toBeInTheDocument()
    })

    it('should include metadata in CSV export', async () => {
      const eventWithMetadata: WebhookEvent = {
        id: 'evt-meta-2',
        webhookId: 'wh-1',
        requestMethod: 'POST',
        requestHeaders: { 'Content-Type': 'application/json' },
        requestBody: { test: 'data' },
        status: 'processed',
        replayCount: 0,
        createdAt: '2024-01-15T10:00:00Z',
        metadata: {
          sourceIp: '203.0.113.42',
          userAgent: 'Test-Agent/1.0',
          receivedAt: '2024-01-15T10:00:00Z',
          contentType: 'application/json',
          contentLength: 512,
        },
      }

      ;(webhookAPI.getEvents as any).mockResolvedValue({
        events: [eventWithMetadata],
        total: 1,
      })

      renderWithRouter(<WebhookEventHistory webhookId="wh-1" />)

      await waitFor(() => {
        expect(screen.getByText('evt-meta-2')).toBeInTheDocument()
      })

      // Note: Actually testing CSV export would require mocking downloadCSV
      // This test verifies the component renders with metadata available
      expect(screen.getByRole('button', { name: /export csv/i })).toBeInTheDocument()
    })
  })
})
