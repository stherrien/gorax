import { describe, it, expect, beforeEach, vi } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import { MemoryRouter, Route, Routes } from 'react-router-dom'
import userEvent from '@testing-library/user-event'
import ExecutionDetail from './ExecutionDetail'
import type { Execution, ExecutionStep } from '../api/executions'

// Mock the hooks
vi.mock('../hooks/useExecutions', () => ({
  useExecution: vi.fn(),
}))

// Mock the API
vi.mock('../api/executions', () => ({
  executionAPI: {
    getSteps: vi.fn(),
    cancel: vi.fn(),
    retry: vi.fn(),
  },
}))

import { useExecution } from '../hooks/useExecutions'
import { executionAPI } from '../api/executions'

// Valid RFC 4122 UUIDs for testing
const executionId = '11111111-1111-4111-8111-111111111111'
const workflowId = '22222222-2222-4222-8222-222222222222'

describe('ExecutionDetail Integration', () => {
  const mockExecution: Execution = {
    id: executionId,
    workflowId: workflowId,
    workflowName: 'Test Workflow',
    status: 'completed',
    trigger: {
      type: 'webhook',
      source: 'api',
    },
    startedAt: '2025-01-15T10:00:00Z',
    completedAt: '2025-01-15T10:05:00Z',
    duration: 300000,
    stepCount: 3,
    completedSteps: 3,
    failedSteps: 0,
  }

  const mockSteps: ExecutionStep[] = [
    {
      id: 'step-1',
      executionId: executionId,
      nodeId: 'node-1',
      nodeName: 'Webhook Trigger',
      status: 'completed',
      startedAt: '2025-01-15T10:00:00Z',
      completedAt: '2025-01-15T10:00:10Z',
      duration: 10000,
      input: { method: 'POST', body: { test: 'data' } },
      output: { statusCode: 200, body: { test: 'data' } },
    },
    {
      id: 'step-2',
      executionId: executionId,
      nodeId: 'node-2',
      nodeName: 'HTTP Request',
      status: 'completed',
      startedAt: '2025-01-15T10:00:10Z',
      completedAt: '2025-01-15T10:00:50Z',
      duration: 40000,
      input: { url: 'https://api.example.com', method: 'GET' },
      output: { statusCode: 200, data: { result: 'success' } },
    },
    {
      id: 'step-3',
      executionId: executionId,
      nodeId: 'node-3',
      nodeName: 'Transform Data',
      status: 'completed',
      startedAt: '2025-01-15T10:00:50Z',
      completedAt: '2025-01-15T10:05:00Z',
      duration: 250000,
      input: { data: { result: 'success' } },
      output: { transformed: true, result: 'SUCCESS' },
    },
  ]

  beforeEach(() => {
    vi.clearAllMocks()
  })

  describe('Load execution details', () => {
    it('should display execution metadata from API', async () => {
      (useExecution as any).mockReturnValue({
        execution: mockExecution,
        loading: false,
        error: null,
        refetch: vi.fn(),
      })

      ;(executionAPI.getSteps as any).mockResolvedValueOnce({ steps: mockSteps })

      render(
        <MemoryRouter initialEntries={[`/executions/${executionId}`]}>
          <Routes>
            <Route path="/executions/:id" element={<ExecutionDetail />} />
          </Routes>
        </MemoryRouter>
      )

      await waitFor(() => {
        expect(screen.getByText('Test Workflow')).toBeInTheDocument()
        const completedBadges = screen.getAllByText('Completed')
        expect(completedBadges.length).toBeGreaterThan(0)
      })
    })

    it('should show loading state while fetching', () => {
      (useExecution as any).mockReturnValue({
        execution: null,
        loading: true,
        error: null,
        refetch: vi.fn(),
      })

      render(
        <MemoryRouter initialEntries={[`/executions/${executionId}`]}>
          <Routes>
            <Route path="/executions/:id" element={<ExecutionDetail />} />
          </Routes>
        </MemoryRouter>
      )

      expect(screen.getByText(/loading/i)).toBeInTheDocument()
    })

    it('should show error message if load fails', () => {
      const error = new Error('Failed to load execution')
      ;(useExecution as any).mockReturnValue({
        execution: null,
        loading: false,
        error,
        refetch: vi.fn(),
      })

      render(
        <MemoryRouter initialEntries={[`/executions/${executionId}`]}>
          <Routes>
            <Route path="/executions/:id" element={<ExecutionDetail />} />
          </Routes>
        </MemoryRouter>
      )

      const errorTexts = screen.getAllByText(/failed to load/i)
      expect(errorTexts.length).toBeGreaterThan(0)
    })
  })

  describe('Execution steps', () => {
    it('should display all execution steps', async () => {
      (useExecution as any).mockReturnValue({
        execution: mockExecution,
        loading: false,
        error: null,
        refetch: vi.fn(),
      })

      ;(executionAPI.getSteps as any).mockResolvedValueOnce({ steps: mockSteps })

      render(
        <MemoryRouter initialEntries={[`/executions/${executionId}`]}>
          <Routes>
            <Route path="/executions/:id" element={<ExecutionDetail />} />
          </Routes>
        </MemoryRouter>
      )

      await waitFor(() => {
        expect(screen.getByText('Webhook Trigger')).toBeInTheDocument()
        expect(screen.getByText('HTTP Request')).toBeInTheDocument()
        expect(screen.getByText('Transform Data')).toBeInTheDocument()
      })
    })

    it('should display step status badges', async () => {
      (useExecution as any).mockReturnValue({
        execution: mockExecution,
        loading: false,
        error: null,
        refetch: vi.fn(),
      })

      ;(executionAPI.getSteps as any).mockResolvedValueOnce({ steps: mockSteps })

      render(
        <MemoryRouter initialEntries={[`/executions/${executionId}`]}>
          <Routes>
            <Route path="/executions/:id" element={<ExecutionDetail />} />
          </Routes>
        </MemoryRouter>
      )

      await waitFor(() => {
        const completedBadges = screen.getAllByText('Completed')
        expect(completedBadges.length).toBeGreaterThan(0)
      })
    })

    it('should display step duration', async () => {
      (useExecution as any).mockReturnValue({
        execution: mockExecution,
        loading: false,
        error: null,
        refetch: vi.fn(),
      })

      ;(executionAPI.getSteps as any).mockResolvedValueOnce({ steps: mockSteps })

      render(
        <MemoryRouter initialEntries={[`/executions/${executionId}`]}>
          <Routes>
            <Route path="/executions/:id" element={<ExecutionDetail />} />
          </Routes>
        </MemoryRouter>
      )

      await waitFor(() => {
        // 10.0s, 40.0s, 250.0s
        expect(screen.getByText('10.0s')).toBeInTheDocument()
        expect(screen.getByText('40.0s')).toBeInTheDocument()
      })
    })

    it('should display failed step with error message', async () => {
      const user = userEvent.setup()
      const failedExecution: Execution = {
        ...mockExecution,
        status: 'failed',
        completedSteps: 1,
        failedSteps: 1,
      }

      const failedSteps: ExecutionStep[] = [
        mockSteps[0],
        {
          ...mockSteps[1],
          status: 'failed',
          error: 'HTTP request timeout',
        },
      ]

      ;(useExecution as any).mockReturnValue({
        execution: failedExecution,
        loading: false,
        error: null,
        refetch: vi.fn(),
      })

      ;(executionAPI.getSteps as any).mockResolvedValueOnce({ steps: failedSteps })

      render(
        <MemoryRouter initialEntries={[`/executions/${executionId}`]}>
          <Routes>
            <Route path="/executions/:id" element={<ExecutionDetail />} />
          </Routes>
        </MemoryRouter>
      )

      await waitFor(() => {
        expect(screen.getByText('Failed')).toBeInTheDocument()
      })

      // Click on the failed step to expand it
      const failedStep = screen.getByText('HTTP Request')
      await user.click(failedStep)

      // Now the error should be visible
      await waitFor(() => {
        expect(screen.getByText(/HTTP request timeout/i)).toBeInTheDocument()
      })
    })
  })

  describe('Execution actions', () => {
    it('should show cancel button for running execution', async () => {
      const runningExecution: Execution = {
        ...mockExecution,
        status: 'running',
        completedAt: undefined,
      }

      ;(useExecution as any).mockReturnValue({
        execution: runningExecution,
        loading: false,
        error: null,
        refetch: vi.fn(),
      })

      ;(executionAPI.getSteps as any).mockResolvedValueOnce({ steps: mockSteps.slice(0, 2) })

      render(
        <MemoryRouter initialEntries={[`/executions/${executionId}`]}>
          <Routes>
            <Route path="/executions/:id" element={<ExecutionDetail />} />
          </Routes>
        </MemoryRouter>
      )

      await waitFor(() => {
        expect(screen.getByRole('button', { name: /cancel/i })).toBeInTheDocument()
      })
    })

    it('should cancel execution when cancel button clicked', async () => {
      const user = userEvent.setup()
      const runningExecution: Execution = {
        ...mockExecution,
        status: 'running',
        completedAt: undefined,
      }

      const refetch = vi.fn()

      ;(useExecution as any).mockReturnValue({
        execution: runningExecution,
        loading: false,
        error: null,
        refetch,
      })

      ;(executionAPI.getSteps as any).mockResolvedValueOnce({ steps: [] })
      ;(executionAPI.cancel as any).mockResolvedValueOnce({
        ...runningExecution,
        status: 'cancelled',
      })

      render(
        <MemoryRouter initialEntries={[`/executions/${executionId}`]}>
          <Routes>
            <Route path="/executions/:id" element={<ExecutionDetail />} />
          </Routes>
        </MemoryRouter>
      )

      const cancelButton = await screen.findByRole('button', { name: /cancel/i })
      await user.click(cancelButton)

      await waitFor(() => {
        expect(executionAPI.cancel).toHaveBeenCalledWith(executionId)
        expect(refetch).toHaveBeenCalled()
        expect(screen.getByText(/execution cancelled successfully/i)).toBeInTheDocument()
      })
    })

    it('should show retry button for failed execution', async () => {
      const failedExecution: Execution = {
        ...mockExecution,
        status: 'failed',
        completedSteps: 1,
        failedSteps: 1,
      }

      ;(useExecution as any).mockReturnValue({
        execution: failedExecution,
        loading: false,
        error: null,
        refetch: vi.fn(),
      })

      ;(executionAPI.getSteps as any).mockResolvedValueOnce({ steps: mockSteps })

      render(
        <MemoryRouter initialEntries={[`/executions/${executionId}`]}>
          <Routes>
            <Route path="/executions/:id" element={<ExecutionDetail />} />
          </Routes>
        </MemoryRouter>
      )

      await waitFor(() => {
        expect(screen.getByRole('button', { name: /retry/i })).toBeInTheDocument()
      })
    })

    it('should retry execution when retry button clicked', async () => {
      const user = userEvent.setup()
      const failedExecution: Execution = {
        ...mockExecution,
        status: 'failed',
      }

      ;(useExecution as any).mockReturnValue({
        execution: failedExecution,
        loading: false,
        error: null,
        refetch: vi.fn(),
      })

      ;(executionAPI.getSteps as any).mockResolvedValueOnce({ steps: mockSteps })
      ;(executionAPI.retry as any).mockResolvedValueOnce({
        ...failedExecution,
        id: 'exec-456',
        status: 'queued',
      })

      render(
        <MemoryRouter initialEntries={[`/executions/${executionId}`]}>
          <Routes>
            <Route path="/executions/:id" element={<ExecutionDetail />} />
          </Routes>
        </MemoryRouter>
      )

      const retryButton = await screen.findByRole('button', { name: /retry/i })
      await user.click(retryButton)

      await waitFor(() => {
        expect(executionAPI.retry).toHaveBeenCalledWith(executionId)
        expect(screen.getByText(/execution started/i)).toBeInTheDocument()
      })
    })
  })

  describe('Navigation', () => {
    it('should have back button to execution list', async () => {
      (useExecution as any).mockReturnValue({
        execution: mockExecution,
        loading: false,
        error: null,
        refetch: vi.fn(),
      })

      ;(executionAPI.getSteps as any).mockResolvedValueOnce({ steps: mockSteps })

      render(
        <MemoryRouter initialEntries={[`/executions/${executionId}`]}>
          <Routes>
            <Route path="/executions/:id" element={<ExecutionDetail />} />
          </Routes>
        </MemoryRouter>
      )

      await waitFor(() => {
        const backButton = screen.getByRole('link', { name: /back to executions/i })
        expect(backButton).toHaveAttribute('href', '/executions')
      })
    })
  })
})
