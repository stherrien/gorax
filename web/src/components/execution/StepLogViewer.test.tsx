import { describe, it, expect, beforeEach, vi } from 'vitest'
import { render, screen, fireEvent } from '@testing-library/react'
import { StepLogViewer } from './StepLogViewer'
import { useExecutionTraceStore } from '../../stores/executionTraceStore'
import type { StepInfo } from '../../lib/websocket'

// Mock the execution trace store
vi.mock('../../stores/executionTraceStore', () => ({
  useExecutionTraceStore: vi.fn(),
}))

// Mock clipboard API
Object.defineProperty(navigator, 'clipboard', {
  value: {
    writeText: vi.fn(() => Promise.resolve()),
  },
  writable: true,
})

describe('StepLogViewer', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  describe('Empty state', () => {
    it('should render empty state when no node is selected', () => {
      vi.mocked(useExecutionTraceStore).mockReturnValue({
        stepLogs: {},
      } as any)

      render(<StepLogViewer selectedNodeId={null} />)

      expect(screen.getByText(/no node selected/i)).toBeInTheDocument()
    })

    it('should show message when selected node has no logs', () => {
      vi.mocked(useExecutionTraceStore).mockReturnValue({
        stepLogs: {},
      } as any)

      render(<StepLogViewer selectedNodeId="node-1" />)

      expect(screen.getByText(/no logs available/i)).toBeInTheDocument()
    })
  })

  describe('Log display', () => {
    const mockStepLog: StepInfo = {
      step_id: 'step-1',
      node_id: 'node-1',
      node_type: 'action:http',
      status: 'completed',
      output_data: { status: 200, body: 'Success' },
      duration_ms: 1500,
      started_at: '2025-01-01T12:00:00Z',
      completed_at: '2025-01-01T12:00:01.5Z',
    }

    it('should display step status', () => {
      vi.mocked(useExecutionTraceStore).mockReturnValue({
        stepLogs: { 'node-1': [mockStepLog] },
      } as any)

      render(<StepLogViewer selectedNodeId="node-1" />)

      const statusBadge = screen.getByText('completed')
      expect(statusBadge).toHaveClass('step-status')
    })

    it('should display node type', () => {
      vi.mocked(useExecutionTraceStore).mockReturnValue({
        stepLogs: { 'node-1': [mockStepLog] },
      } as any)

      render(<StepLogViewer selectedNodeId="node-1" />)

      expect(screen.getByText(/action:http/i)).toBeInTheDocument()
    })

    it('should display duration when available', () => {
      vi.mocked(useExecutionTraceStore).mockReturnValue({
        stepLogs: { 'node-1': [mockStepLog] },
      } as any)

      render(<StepLogViewer selectedNodeId="node-1" />)

      expect(screen.getByText(/1500ms/i)).toBeInTheDocument()
    })

    it('should display formatted output data', () => {
      vi.mocked(useExecutionTraceStore).mockReturnValue({
        stepLogs: { 'node-1': [mockStepLog] },
      } as any)

      render(<StepLogViewer selectedNodeId="node-1" />)

      expect(screen.getByText(/"status": 200/)).toBeInTheDocument()
      expect(screen.getByText(/"body": "Success"/)).toBeInTheDocument()
    })

    it('should display timestamps', () => {
      vi.mocked(useExecutionTraceStore).mockReturnValue({
        stepLogs: { 'node-1': [mockStepLog] },
      } as any)

      render(<StepLogViewer selectedNodeId="node-1" />)

      expect(screen.getByText(/started at/i)).toBeInTheDocument()
      expect(screen.getByText(/completed at/i)).toBeInTheDocument()
    })
  })

  describe('Error handling', () => {
    it('should display error message when step failed', () => {
      const failedStep: StepInfo = {
        step_id: 'step-1',
        node_id: 'node-1',
        node_type: 'action:http',
        status: 'failed',
        error: 'Connection timeout',
      }

      vi.mocked(useExecutionTraceStore).mockReturnValue({
        stepLogs: { 'node-1': [failedStep] },
      } as any)

      render(<StepLogViewer selectedNodeId="node-1" />)

      expect(screen.getByText(/connection timeout/i)).toBeInTheDocument()
    })

    it('should highlight error message in red', () => {
      const failedStep: StepInfo = {
        step_id: 'step-1',
        node_id: 'node-1',
        node_type: 'action:http',
        status: 'failed',
        error: 'Error occurred',
      }

      vi.mocked(useExecutionTraceStore).mockReturnValue({
        stepLogs: { 'node-1': [failedStep] },
      } as any)

      render(<StepLogViewer selectedNodeId="node-1" />)

      const errorSection = screen.getByTestId('error-section')
      expect(errorSection).toHaveClass('error-section')
    })
  })

  describe('Copy to clipboard', () => {
    const mockStepLog: StepInfo = {
      step_id: 'step-1',
      node_id: 'node-1',
      node_type: 'action:http',
      status: 'completed',
      output_data: { result: 'success' },
    }

    it('should show copy button for output data', () => {
      vi.mocked(useExecutionTraceStore).mockReturnValue({
        stepLogs: { 'node-1': [mockStepLog] },
      } as any)

      render(<StepLogViewer selectedNodeId="node-1" />)

      expect(screen.getByRole('button', { name: /copy/i })).toBeInTheDocument()
    })

    it('should copy output data to clipboard when button clicked', async () => {
      vi.mocked(useExecutionTraceStore).mockReturnValue({
        stepLogs: { 'node-1': [mockStepLog] },
      } as any)

      render(<StepLogViewer selectedNodeId="node-1" />)

      const copyButton = screen.getByRole('button', { name: /copy/i })
      fireEvent.click(copyButton)

      expect(navigator.clipboard.writeText).toHaveBeenCalledWith(
        JSON.stringify(mockStepLog.output_data, null, 2)
      )
    })

    it('should show success feedback after copying', async () => {
      vi.mocked(useExecutionTraceStore).mockReturnValue({
        stepLogs: { 'node-1': [mockStepLog] },
      } as any)

      render(<StepLogViewer selectedNodeId="node-1" />)

      const copyButton = screen.getByRole('button', { name: /copy/i })
      fireEvent.click(copyButton)

      expect(await screen.findByText(/copied/i)).toBeInTheDocument()
    })
  })

  describe('Multiple steps', () => {
    const mockStepLogs: StepInfo[] = [
      {
        step_id: 'step-1',
        node_id: 'node-1',
        node_type: 'action:http',
        status: 'completed',
        duration_ms: 100,
      },
      {
        step_id: 'step-2',
        node_id: 'node-1',
        node_type: 'action:http',
        status: 'completed',
        duration_ms: 200,
      },
    ]

    it('should display all steps for a node', () => {
      vi.mocked(useExecutionTraceStore).mockReturnValue({
        stepLogs: { 'node-1': mockStepLogs },
      } as any)

      render(<StepLogViewer selectedNodeId="node-1" />)

      expect(screen.getByText(/step-1/i)).toBeInTheDocument()
      expect(screen.getByText(/step-2/i)).toBeInTheDocument()
    })

    it('should show steps in chronological order', () => {
      vi.mocked(useExecutionTraceStore).mockReturnValue({
        stepLogs: { 'node-1': mockStepLogs },
      } as any)

      render(<StepLogViewer selectedNodeId="node-1" />)

      const steps = screen.getAllByTestId(/step-log-/)
      expect(steps).toHaveLength(2)
    })
  })

  describe('Accessibility', () => {
    const mockStepLog: StepInfo = {
      step_id: 'step-1',
      node_id: 'node-1',
      node_type: 'action:http',
      status: 'completed',
    }

    it('should have proper heading hierarchy', () => {
      vi.mocked(useExecutionTraceStore).mockReturnValue({
        stepLogs: { 'node-1': [mockStepLog] },
      } as any)

      render(<StepLogViewer selectedNodeId="node-1" />)

      expect(screen.getByRole('heading', { name: /step logs/i })).toBeInTheDocument()
    })

    it('should have accessible copy button', () => {
      vi.mocked(useExecutionTraceStore).mockReturnValue({
        stepLogs: {
          'node-1': [
            {
              ...mockStepLog,
              output_data: { test: 'data' },
            },
          ],
        },
      } as any)

      render(<StepLogViewer selectedNodeId="node-1" />)

      const copyButton = screen.getByRole('button', { name: /copy/i })
      expect(copyButton).toHaveAttribute('aria-label')
    })

    it('should mark code blocks with proper semantic HTML', () => {
      vi.mocked(useExecutionTraceStore).mockReturnValue({
        stepLogs: {
          'node-1': [
            {
              ...mockStepLog,
              output_data: { test: 'data' },
            },
          ],
        },
      } as any)

      render(<StepLogViewer selectedNodeId="node-1" />)

      expect(screen.getByRole('code')).toBeInTheDocument()
    })
  })
})
