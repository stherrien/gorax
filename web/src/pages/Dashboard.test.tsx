import { describe, it, expect, beforeEach, vi } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import Dashboard from './Dashboard'
import type { DashboardStats, Execution } from '../api/executions'

// Mock the hooks
vi.mock('../hooks/useExecutions', () => ({
  useDashboardStats: vi.fn(),
  useRecentExecutions: vi.fn(),
}))

vi.mock('../hooks/useWorkflows', () => ({
  useWorkflows: vi.fn(),
}))

import { useDashboardStats, useRecentExecutions } from '../hooks/useExecutions'
import { useWorkflows } from '../hooks/useWorkflows'

describe('Dashboard Integration', () => {
  const mockStats: DashboardStats = {
    totalExecutions: 1000,
    executionsToday: 847,
    failedToday: 3,
    successRateToday: 99.6,
    averageDuration: 45000,
    activeWorkflows: 12,
  }

  const mockRecentExecutions: Execution[] = [
    {
      id: 'exec-1',
      workflowId: 'wf-1',
      workflowName: 'Hello World Workflow',
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
      workflowName: 'Data Processing Pipeline',
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
      error: 'HTTP request failed',
    },
  ]

  beforeEach(() => {
    vi.clearAllMocks()

    // Default mock returns
    ;(useWorkflows as any).mockReturnValue({
      workflows: [],
      total: 0,
      loading: false,
      error: null,
      refetch: vi.fn(),
    })
  })

  describe('Load dashboard data', () => {
    it('should display dashboard statistics from API', async () => {
      ;(useDashboardStats as any).mockReturnValue({
        stats: mockStats,
        loading: false,
        error: null,
        refetch: vi.fn(),
      })

      ;(useRecentExecutions as any).mockReturnValue({
        executions: [],
        loading: false,
        error: null,
        refetch: vi.fn(),
      })

      render(
        <MemoryRouter>
          <Dashboard />
        </MemoryRouter>
      )

      await waitFor(() => {
        expect(screen.getByText('12')).toBeInTheDocument() // Active Workflows
        expect(screen.getByText('847')).toBeInTheDocument() // Executions Today
        expect(screen.getByText('3')).toBeInTheDocument() // Failed Executions
      })
    })

    it('should show loading state while fetching stats', () => {
      ;(useDashboardStats as any).mockReturnValue({
        stats: null,
        loading: true,
        error: null,
        refetch: vi.fn(),
      })

      ;(useRecentExecutions as any).mockReturnValue({
        executions: [],
        loading: true,
        error: null,
        refetch: vi.fn(),
      })

      render(
        <MemoryRouter>
          <Dashboard />
        </MemoryRouter>
      )

      expect(screen.getByText(/loading/i)).toBeInTheDocument()
    })

    it('should show error message if stats fetch fails', () => {
      const error = new Error('Failed to load dashboard stats')
      ;(useDashboardStats as any).mockReturnValue({
        stats: null,
        loading: false,
        error,
        refetch: vi.fn(),
      })

      ;(useRecentExecutions as any).mockReturnValue({
        executions: [],
        loading: false,
        error: null,
        refetch: vi.fn(),
      })

      render(
        <MemoryRouter>
          <Dashboard />
        </MemoryRouter>
      )

      const errorTexts = screen.getAllByText(/failed to load/i)
      expect(errorTexts.length).toBeGreaterThan(0)
    })
  })

  describe('Recent executions', () => {
    it('should display recent executions from API', async () => {
      ;(useDashboardStats as any).mockReturnValue({
        stats: mockStats,
        loading: false,
        error: null,
        refetch: vi.fn(),
      })

      ;(useRecentExecutions as any).mockReturnValue({
        executions: mockRecentExecutions,
        loading: false,
        error: null,
        refetch: vi.fn(),
      })

      render(
        <MemoryRouter>
          <Dashboard />
        </MemoryRouter>
      )

      await waitFor(() => {
        expect(screen.getByText('Hello World Workflow')).toBeInTheDocument()
        expect(screen.getByText('Data Processing Pipeline')).toBeInTheDocument()
      })
    })

    it('should display execution status badges', async () => {
      ;(useDashboardStats as any).mockReturnValue({
        stats: mockStats,
        loading: false,
        error: null,
        refetch: vi.fn(),
      })

      ;(useRecentExecutions as any).mockReturnValue({
        executions: mockRecentExecutions,
        loading: false,
        error: null,
        refetch: vi.fn(),
      })

      render(
        <MemoryRouter>
          <Dashboard />
        </MemoryRouter>
      )

      await waitFor(() => {
        expect(screen.getByText('Completed')).toBeInTheDocument()
        expect(screen.getByText('Failed')).toBeInTheDocument()
      })
    })

    it('should show empty state when no recent executions', async () => {
      ;(useDashboardStats as any).mockReturnValue({
        stats: mockStats,
        loading: false,
        error: null,
        refetch: vi.fn(),
      })

      ;(useRecentExecutions as any).mockReturnValue({
        executions: [],
        loading: false,
        error: null,
        refetch: vi.fn(),
      })

      render(
        <MemoryRouter>
          <Dashboard />
        </MemoryRouter>
      )

      await waitFor(() => {
        expect(screen.getByText(/no recent executions/i)).toBeInTheDocument()
      })
    })

    it('should display trigger type for executions', async () => {
      ;(useDashboardStats as any).mockReturnValue({
        stats: mockStats,
        loading: false,
        error: null,
        refetch: vi.fn(),
      })

      ;(useRecentExecutions as any).mockReturnValue({
        executions: mockRecentExecutions,
        loading: false,
        error: null,
        refetch: vi.fn(),
      })

      render(
        <MemoryRouter>
          <Dashboard />
        </MemoryRouter>
      )

      await waitFor(() => {
        expect(screen.getByText(/webhook/i)).toBeInTheDocument()
        expect(screen.getByText(/schedule/i)).toBeInTheDocument()
      })
    })
  })

  describe('Navigation', () => {
    it('should have quick action links', () => {
      ;(useDashboardStats as any).mockReturnValue({
        stats: mockStats,
        loading: false,
        error: null,
        refetch: vi.fn(),
      })

      ;(useRecentExecutions as any).mockReturnValue({
        executions: [],
        loading: false,
        error: null,
        refetch: vi.fn(),
      })

      render(
        <MemoryRouter>
          <Dashboard />
        </MemoryRouter>
      )

      const createButton = screen.getByRole('link', { name: /create workflow/i })
      expect(createButton).toHaveAttribute('href', '/workflows/new')

      const viewWorkflowsButton = screen.getByRole('link', { name: /view all workflows/i })
      expect(viewWorkflowsButton).toHaveAttribute('href', '/workflows')

      const viewExecutionsButton = screen.getByRole('link', { name: /view executions/i })
      expect(viewExecutionsButton).toHaveAttribute('href', '/executions')
    })
  })
})
