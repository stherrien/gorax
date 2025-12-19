import { describe, it, expect, beforeEach, vi } from 'vitest'
import { render, screen, fireEvent } from '@testing-library/react'
import { ExecutionDetailsPanel } from './ExecutionDetailsPanel'
import { useExecutionTraceStore } from '../../stores/executionTraceStore'

// Mock the execution trace store
vi.mock('../../stores/executionTraceStore', () => ({
  useExecutionTraceStore: vi.fn(),
}))

// Mock child components
vi.mock('./ExecutionTimeline', () => ({
  ExecutionTimeline: () => <div data-testid="execution-timeline">Timeline</div>,
}))

vi.mock('./StepLogViewer', () => ({
  StepLogViewer: ({ selectedNodeId }: { selectedNodeId: string | null }) => (
    <div data-testid="step-log-viewer">
      Log Viewer - {selectedNodeId || 'none'}
    </div>
  ),
}))

describe('ExecutionDetailsPanel', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  describe('Panel visibility', () => {
    it('should render panel when execution ID is provided', () => {
      vi.mocked(useExecutionTraceStore).mockReturnValue({
        currentExecutionId: 'exec-123',
        nodeStatuses: {},
      } as any)

      render(<ExecutionDetailsPanel selectedNodeId={null} />)

      expect(screen.getByTestId('execution-details-panel')).toBeInTheDocument()
    })

    it('should not render panel when no execution is active', () => {
      vi.mocked(useExecutionTraceStore).mockReturnValue({
        currentExecutionId: null,
        nodeStatuses: {},
      } as any)

      render(<ExecutionDetailsPanel selectedNodeId={null} />)

      expect(screen.queryByTestId('execution-details-panel')).not.toBeInTheDocument()
    })
  })

  describe('Header section', () => {
    it('should display execution ID in header', () => {
      vi.mocked(useExecutionTraceStore).mockReturnValue({
        currentExecutionId: 'exec-123',
        nodeStatuses: {},
      } as any)

      render(<ExecutionDetailsPanel selectedNodeId={null} />)

      expect(screen.getByText(/exec-123/i)).toBeInTheDocument()
    })

    it('should show execution status in header', () => {
      vi.mocked(useExecutionTraceStore).mockReturnValue({
        currentExecutionId: 'exec-123',
        nodeStatuses: { 'node-1': 'running', 'node-2': 'completed' },
      } as any)

      render(<ExecutionDetailsPanel selectedNodeId={null} />)

      expect(screen.getByText(/running/i)).toBeInTheDocument()
    })

    it('should have sticky header', () => {
      vi.mocked(useExecutionTraceStore).mockReturnValue({
        currentExecutionId: 'exec-123',
        nodeStatuses: {},
      } as any)

      render(<ExecutionDetailsPanel selectedNodeId={null} />)

      const header = screen.getByTestId('panel-header')
      expect(header).toHaveClass('sticky')
    })
  })

  describe('Tab navigation', () => {
    beforeEach(() => {
      vi.mocked(useExecutionTraceStore).mockReturnValue({
        currentExecutionId: 'exec-123',
        nodeStatuses: {},
      } as any)
    })

    it('should show timeline tab by default', () => {
      render(<ExecutionDetailsPanel selectedNodeId={null} />)

      const timelineTab = screen.getByRole('tab', { name: /timeline/i })
      expect(timelineTab).toHaveAttribute('aria-selected', 'true')
    })

    it('should show logs tab', () => {
      render(<ExecutionDetailsPanel selectedNodeId={null} />)

      expect(screen.getByRole('tab', { name: /logs/i })).toBeInTheDocument()
    })

    it('should switch to logs tab when clicked', () => {
      render(<ExecutionDetailsPanel selectedNodeId={null} />)

      const logsTab = screen.getByRole('tab', { name: /logs/i })
      fireEvent.click(logsTab)

      expect(logsTab).toHaveAttribute('aria-selected', 'true')
    })

    it('should display timeline content when timeline tab is active', () => {
      render(<ExecutionDetailsPanel selectedNodeId={null} />)

      expect(screen.getByTestId('execution-timeline')).toBeInTheDocument()
    })

    it('should display logs content when logs tab is active', () => {
      render(<ExecutionDetailsPanel selectedNodeId={null} />)

      const logsTab = screen.getByRole('tab', { name: /logs/i })
      fireEvent.click(logsTab)

      expect(screen.getByTestId('step-log-viewer')).toBeInTheDocument()
    })

    it('should pass selected node ID to log viewer', () => {
      render(<ExecutionDetailsPanel selectedNodeId="node-1" />)

      const logsTab = screen.getByRole('tab', { name: /logs/i })
      fireEvent.click(logsTab)

      expect(screen.getByText(/node-1/)).toBeInTheDocument()
    })
  })

  describe('Keyboard navigation', () => {
    beforeEach(() => {
      vi.mocked(useExecutionTraceStore).mockReturnValue({
        currentExecutionId: 'exec-123',
        nodeStatuses: {},
      } as any)
    })

    it('should support arrow key navigation between tabs', () => {
      render(<ExecutionDetailsPanel selectedNodeId={null} />)

      const timelineTab = screen.getByRole('tab', { name: /timeline/i })
      const logsTab = screen.getByRole('tab', { name: /logs/i })

      timelineTab.focus()
      fireEvent.keyDown(timelineTab, { key: 'ArrowRight' })

      expect(logsTab).toHaveFocus()
    })

    it('should support Enter key to activate tab', () => {
      render(<ExecutionDetailsPanel selectedNodeId={null} />)

      const logsTab = screen.getByRole('tab', { name: /logs/i })
      fireEvent.keyDown(logsTab, { key: 'Enter' })

      expect(logsTab).toHaveAttribute('aria-selected', 'true')
    })
  })

  describe('Accessibility', () => {
    beforeEach(() => {
      vi.mocked(useExecutionTraceStore).mockReturnValue({
        currentExecutionId: 'exec-123',
        nodeStatuses: {},
      } as any)
    })

    it('should have proper ARIA role for tab list', () => {
      render(<ExecutionDetailsPanel selectedNodeId={null} />)

      expect(screen.getByRole('tablist')).toBeInTheDocument()
    })

    it('should have proper ARIA labels for tabs', () => {
      render(<ExecutionDetailsPanel selectedNodeId={null} />)

      const timelineTab = screen.getByRole('tab', { name: /timeline/i })
      const logsTab = screen.getByRole('tab', { name: /logs/i })

      expect(timelineTab).toHaveAttribute('aria-controls')
      expect(logsTab).toHaveAttribute('aria-controls')
    })

    it('should have proper ARIA role for tab panels', () => {
      render(<ExecutionDetailsPanel selectedNodeId={null} />)

      expect(screen.getByRole('tabpanel')).toBeInTheDocument()
    })

    it('should announce active tab to screen readers', () => {
      render(<ExecutionDetailsPanel selectedNodeId={null} />)

      const logsTab = screen.getByRole('tab', { name: /logs/i })
      fireEvent.click(logsTab)

      expect(logsTab).toHaveAttribute('aria-selected', 'true')
    })
  })
})
