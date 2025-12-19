import { describe, it, expect, beforeEach, vi } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import userEvent from '@testing-library/user-event'
import Executions from './Executions'
import type { Execution } from '../api/executions'

// Mock the hooks
vi.mock('../hooks/useExecutions', () => ({
  useExecutions: vi.fn(),
}))

import { useExecutions } from '../hooks/useExecutions'

describe('Executions List Integration', () => {
  const mockExecutions: Execution[] = [
    {
      id: 'exec-1',
      workflowId: 'wf-1',
      workflowName: 'Data Pipeline',
      status: 'completed',
      trigger: {
        type: 'webhook',
        source: 'api',
      },
      startedAt: '2025-01-15T10:00:00Z',
      completedAt: '2025-01-15T10:05:00Z',
      duration: 300000,
      stepCount: 5,
      completedSteps: 5,
      failedSteps: 0,
    },
    {
      id: 'exec-2',
      workflowId: 'wf-2',
      workflowName: 'Email Automation',
      status: 'failed',
      trigger: {
        type: 'schedule',
      },
      startedAt: '2025-01-15T10:10:00Z',
      completedAt: '2025-01-15T10:15:00Z',
      duration: 300000,
      stepCount: 8,
      completedSteps: 5,
      failedSteps: 3,
      error: 'SMTP connection failed',
    },
    {
      id: 'exec-3',
      workflowId: 'wf-1',
      workflowName: 'Data Pipeline',
      status: 'running',
      trigger: {
        type: 'manual',
      },
      startedAt: '2025-01-15T10:20:00Z',
      duration: 120000,
      stepCount: 5,
      completedSteps: 3,
      failedSteps: 0,
    },
  ]

  beforeEach(() => {
    vi.clearAllMocks()
  })

  describe('Load executions list', () => {
    it('should display list of executions from API', async () => {
      ;(useExecutions as any).mockReturnValue({
        executions: mockExecutions,
        total: 3,
        loading: false,
        error: null,
        refetch: vi.fn(),
      })

      render(
        <MemoryRouter>
          <Executions />
        </MemoryRouter>
      )

      await waitFor(() => {
        const dataPipelines = screen.getAllByText('Data Pipeline')
        expect(dataPipelines.length).toBeGreaterThan(0)
        expect(screen.getByText('Email Automation')).toBeInTheDocument()
      })
    })

    it('should show loading state while fetching', () => {
      ;(useExecutions as any).mockReturnValue({
        executions: [],
        total: 0,
        loading: true,
        error: null,
        refetch: vi.fn(),
      })

      render(
        <MemoryRouter>
          <Executions />
        </MemoryRouter>
      )

      expect(screen.getByText(/loading/i)).toBeInTheDocument()
    })

    it('should show error message if load fails', () => {
      const error = new Error('Failed to load executions')
      ;(useExecutions as any).mockReturnValue({
        executions: [],
        total: 0,
        loading: false,
        error,
        refetch: vi.fn(),
      })

      render(
        <MemoryRouter>
          <Executions />
        </MemoryRouter>
      )

      const errorTexts = screen.getAllByText(/failed to load/i)
      expect(errorTexts.length).toBeGreaterThan(0)
    })

    it('should show empty state when no executions', () => {
      ;(useExecutions as any).mockReturnValue({
        executions: [],
        total: 0,
        loading: false,
        error: null,
        refetch: vi.fn(),
      })

      render(
        <MemoryRouter>
          <Executions />
        </MemoryRouter>
      )

      expect(screen.getByText(/no executions found/i)).toBeInTheDocument()
    })
  })

  describe('Display execution information', () => {
    it('should display workflow names', async () => {
      ;(useExecutions as any).mockReturnValue({
        executions: mockExecutions,
        total: 3,
        loading: false,
        error: null,
        refetch: vi.fn(),
      })

      render(
        <MemoryRouter>
          <Executions />
        </MemoryRouter>
      )

      await waitFor(() => {
        const dataPipelines = screen.getAllByText('Data Pipeline')
        expect(dataPipelines.length).toBeGreaterThan(0)
        expect(screen.getByText('Email Automation')).toBeInTheDocument()
      })
    })

    it('should display status badges', async () => {
      ;(useExecutions as any).mockReturnValue({
        executions: mockExecutions,
        total: 3,
        loading: false,
        error: null,
        refetch: vi.fn(),
      })

      render(
        <MemoryRouter>
          <Executions />
        </MemoryRouter>
      )

      await waitFor(() => {
        const completedBadges = screen.getAllByText('Completed')
        expect(completedBadges.length).toBeGreaterThan(0)
        const failedBadges = screen.getAllByText('Failed')
        expect(failedBadges.length).toBeGreaterThan(0)
        const runningBadges = screen.getAllByText('Running')
        expect(runningBadges.length).toBeGreaterThan(0)
      })
    })

    it('should display trigger types', async () => {
      ;(useExecutions as any).mockReturnValue({
        executions: mockExecutions,
        total: 3,
        loading: false,
        error: null,
        refetch: vi.fn(),
      })

      render(
        <MemoryRouter>
          <Executions />
        </MemoryRouter>
      )

      await waitFor(() => {
        expect(screen.getByText(/webhook/i)).toBeInTheDocument()
        expect(screen.getByText(/schedule/i)).toBeInTheDocument()
        expect(screen.getByText(/manual/i)).toBeInTheDocument()
      })
    })

    it('should display execution times', async () => {
      ;(useExecutions as any).mockReturnValue({
        executions: mockExecutions,
        total: 3,
        loading: false,
        error: null,
        refetch: vi.fn(),
      })

      render(
        <MemoryRouter>
          <Executions />
        </MemoryRouter>
      )

      await waitFor(() => {
        // Should show relative times like "2 hours ago"
        const timeElements = screen.getAllByText(/ago/)
        expect(timeElements.length).toBeGreaterThan(0)
      })
    })

    it('should display step progress', async () => {
      ;(useExecutions as any).mockReturnValue({
        executions: mockExecutions,
        total: 3,
        loading: false,
        error: null,
        refetch: vi.fn(),
      })

      render(
        <MemoryRouter>
          <Executions />
        </MemoryRouter>
      )

      await waitFor(() => {
        // Should show something like "5/5 steps"
        expect(screen.getByText(/5\/5/)).toBeInTheDocument()
        expect(screen.getByText(/3\/5/)).toBeInTheDocument()
      })
    })
  })

  describe('Filtering', () => {
    it('should have status filter dropdown', () => {
      ;(useExecutions as any).mockReturnValue({
        executions: mockExecutions,
        total: 3,
        loading: false,
        error: null,
        refetch: vi.fn(),
      })

      render(
        <MemoryRouter>
          <Executions />
        </MemoryRouter>
      )

      expect(screen.getByLabelText(/status/i)).toBeInTheDocument()
    })

    it('should filter by status when selected', async () => {
      const user = userEvent.setup()
      const mockRefetch = vi.fn()

      ;(useExecutions as any).mockReturnValue({
        executions: mockExecutions,
        total: 3,
        loading: false,
        error: null,
        refetch: mockRefetch,
      })

      render(
        <MemoryRouter>
          <Executions />
        </MemoryRouter>
      )

      const statusFilter = screen.getByLabelText(/status/i)
      await user.selectOptions(statusFilter, 'failed')

      // Should trigger refetch with new params
      await waitFor(() => {
        expect(useExecutions).toHaveBeenCalledWith(
          expect.objectContaining({
            status: 'failed',
          })
        )
      })
    })

    it('should have workflow filter dropdown', () => {
      ;(useExecutions as any).mockReturnValue({
        executions: mockExecutions,
        total: 3,
        loading: false,
        error: null,
        refetch: vi.fn(),
      })

      render(
        <MemoryRouter>
          <Executions />
        </MemoryRouter>
      )

      expect(screen.getByLabelText(/workflow/i)).toBeInTheDocument()
    })

    it('should have search input for workflow name', () => {
      ;(useExecutions as any).mockReturnValue({
        executions: mockExecutions,
        total: 3,
        loading: false,
        error: null,
        refetch: vi.fn(),
      })

      render(
        <MemoryRouter>
          <Executions />
        </MemoryRouter>
      )

      expect(screen.getByPlaceholderText(/search/i)).toBeInTheDocument()
    })
  })

  describe('Navigation', () => {
    it('should have links to execution detail pages', async () => {
      ;(useExecutions as any).mockReturnValue({
        executions: mockExecutions,
        total: 3,
        loading: false,
        error: null,
        refetch: vi.fn(),
      })

      render(
        <MemoryRouter>
          <Executions />
        </MemoryRouter>
      )

      await waitFor(() => {
        const links = screen.getAllByRole('link')
        const executionLinks = links.filter((link) =>
          link.getAttribute('href')?.startsWith('/executions/')
        )
        expect(executionLinks.length).toBeGreaterThan(0)
      })
    })

    it('should navigate to execution detail when clicked', async () => {
      const user = userEvent.setup()
      ;(useExecutions as any).mockReturnValue({
        executions: mockExecutions,
        total: 3,
        loading: false,
        error: null,
        refetch: vi.fn(),
      })

      render(
        <MemoryRouter>
          <Executions />
        </MemoryRouter>
      )

      await waitFor(() => {
        const links = screen.getAllByRole('link')
        const firstExecutionLink = links.find((link) =>
          link.getAttribute('href')?.includes('exec-1')
        )
        expect(firstExecutionLink).toHaveAttribute('href', '/executions/exec-1')
      })
    })
  })

  describe('Pagination', () => {
    it('should display total count', async () => {
      ;(useExecutions as any).mockReturnValue({
        executions: mockExecutions,
        total: 100,
        loading: false,
        error: null,
        refetch: vi.fn(),
      })

      render(
        <MemoryRouter>
          <Executions />
        </MemoryRouter>
      )

      await waitFor(() => {
        expect(screen.getByText(/100/)).toBeInTheDocument()
      })
    })

    it('should have pagination controls when total > page size', async () => {
      ;(useExecutions as any).mockReturnValue({
        executions: mockExecutions,
        total: 100,
        loading: false,
        error: null,
        refetch: vi.fn(),
      })

      render(
        <MemoryRouter>
          <Executions />
        </MemoryRouter>
      )

      await waitFor(() => {
        expect(screen.getByRole('button', { name: /next/i })).toBeInTheDocument()
      })
    })

    it('should change page when pagination clicked', async () => {
      const user = userEvent.setup()
      ;(useExecutions as any).mockReturnValue({
        executions: mockExecutions,
        total: 100,
        loading: false,
        error: null,
        refetch: vi.fn(),
      })

      render(
        <MemoryRouter>
          <Executions />
        </MemoryRouter>
      )

      const nextButton = await screen.findByRole('button', { name: /next/i })
      await user.click(nextButton)

      await waitFor(() => {
        expect(useExecutions).toHaveBeenCalledWith(
          expect.objectContaining({
            page: 2,
          })
        )
      })
    })
  })
})
