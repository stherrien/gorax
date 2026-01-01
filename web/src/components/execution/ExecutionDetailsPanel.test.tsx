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

  describe('Execution status display', () => {
    it('should show idle status when no node statuses', () => {
      vi.mocked(useExecutionTraceStore).mockReturnValue({
        currentExecutionId: 'exec-123',
        nodeStatuses: {},
      } as any)

      render(<ExecutionDetailsPanel selectedNodeId={null} />)

      expect(screen.getByText('idle')).toBeInTheDocument()
    })

    it('should show failed status when any node has failed', () => {
      vi.mocked(useExecutionTraceStore).mockReturnValue({
        currentExecutionId: 'exec-123',
        nodeStatuses: { 'node-1': 'completed', 'node-2': 'failed' },
      } as any)

      render(<ExecutionDetailsPanel selectedNodeId={null} />)

      expect(screen.getByText('failed')).toBeInTheDocument()
    })

    it('should show running status when any node is running', () => {
      vi.mocked(useExecutionTraceStore).mockReturnValue({
        currentExecutionId: 'exec-123',
        nodeStatuses: { 'node-1': 'completed', 'node-2': 'running' },
      } as any)

      render(<ExecutionDetailsPanel selectedNodeId={null} />)

      expect(screen.getByText('running')).toBeInTheDocument()
    })

    it('should show completed status when all nodes completed', () => {
      vi.mocked(useExecutionTraceStore).mockReturnValue({
        currentExecutionId: 'exec-123',
        nodeStatuses: { 'node-1': 'completed', 'node-2': 'completed' },
      } as any)

      render(<ExecutionDetailsPanel selectedNodeId={null} />)

      expect(screen.getByText('completed')).toBeInTheDocument()
    })

    it('should show pending status as default', () => {
      vi.mocked(useExecutionTraceStore).mockReturnValue({
        currentExecutionId: 'exec-123',
        nodeStatuses: { 'node-1': 'pending', 'node-2': 'queued' },
      } as any)

      render(<ExecutionDetailsPanel selectedNodeId={null} />)

      expect(screen.getByText('pending')).toBeInTheDocument()
    })

    it('should have status-specific CSS class', () => {
      vi.mocked(useExecutionTraceStore).mockReturnValue({
        currentExecutionId: 'exec-123',
        nodeStatuses: { 'node-1': 'failed' },
      } as any)

      render(<ExecutionDetailsPanel selectedNodeId={null} />)

      const statusElement = screen.getByText('failed')
      expect(statusElement).toHaveClass('status-failed')
    })
  })

  describe('Extended keyboard navigation', () => {
    beforeEach(() => {
      vi.mocked(useExecutionTraceStore).mockReturnValue({
        currentExecutionId: 'exec-123',
        nodeStatuses: {},
      } as any)
    })

    it('should support ArrowLeft key navigation', () => {
      render(<ExecutionDetailsPanel selectedNodeId={null} />)

      const logsTab = screen.getByRole('tab', { name: /logs/i })
      const timelineTab = screen.getByRole('tab', { name: /timeline/i })

      // First switch to logs tab
      fireEvent.click(logsTab)
      expect(logsTab).toHaveAttribute('aria-selected', 'true')

      // Focus logs tab and press ArrowLeft
      logsTab.focus()
      fireEvent.keyDown(logsTab, { key: 'ArrowLeft' })

      expect(timelineTab).toHaveFocus()
    })

    it('should support Space key to activate tab', () => {
      render(<ExecutionDetailsPanel selectedNodeId={null} />)

      const logsTab = screen.getByRole('tab', { name: /logs/i })
      fireEvent.keyDown(logsTab, { key: ' ' })

      expect(logsTab).toHaveAttribute('aria-selected', 'true')
    })

    it('should wrap ArrowRight from logs to timeline', () => {
      render(<ExecutionDetailsPanel selectedNodeId={null} />)

      const logsTab = screen.getByRole('tab', { name: /logs/i })
      const timelineTab = screen.getByRole('tab', { name: /timeline/i })

      // Focus logs tab and press ArrowRight (should wrap to timeline)
      logsTab.focus()
      fireEvent.keyDown(logsTab, { key: 'ArrowRight' })

      expect(timelineTab).toHaveFocus()
    })

    it('should wrap ArrowLeft from timeline to logs', () => {
      render(<ExecutionDetailsPanel selectedNodeId={null} />)

      const timelineTab = screen.getByRole('tab', { name: /timeline/i })
      const logsTab = screen.getByRole('tab', { name: /logs/i })

      // Focus timeline tab and press ArrowLeft (should wrap to logs)
      timelineTab.focus()
      fireEvent.keyDown(timelineTab, { key: 'ArrowLeft' })

      expect(logsTab).toHaveFocus()
    })
  })

  describe('Tab switching', () => {
    beforeEach(() => {
      vi.mocked(useExecutionTraceStore).mockReturnValue({
        currentExecutionId: 'exec-123',
        nodeStatuses: {},
      } as any)
    })

    it('should switch from logs back to timeline', () => {
      render(<ExecutionDetailsPanel selectedNodeId={null} />)

      const logsTab = screen.getByRole('tab', { name: /logs/i })
      const timelineTab = screen.getByRole('tab', { name: /timeline/i })

      // Switch to logs
      fireEvent.click(logsTab)
      expect(screen.getByTestId('step-log-viewer')).toBeInTheDocument()
      expect(screen.queryByTestId('execution-timeline')).not.toBeInTheDocument()

      // Switch back to timeline
      fireEvent.click(timelineTab)
      expect(screen.getByTestId('execution-timeline')).toBeInTheDocument()
      expect(screen.queryByTestId('step-log-viewer')).not.toBeInTheDocument()
    })

    it('should deselect logs tab when switching to timeline', () => {
      render(<ExecutionDetailsPanel selectedNodeId={null} />)

      const logsTab = screen.getByRole('tab', { name: /logs/i })
      const timelineTab = screen.getByRole('tab', { name: /timeline/i })

      // Switch to logs
      fireEvent.click(logsTab)
      expect(logsTab).toHaveAttribute('aria-selected', 'true')

      // Switch back to timeline
      fireEvent.click(timelineTab)
      expect(logsTab).toHaveAttribute('aria-selected', 'false')
      expect(timelineTab).toHaveAttribute('aria-selected', 'true')
    })

    it('should update tab panel id based on active tab', () => {
      render(<ExecutionDetailsPanel selectedNodeId={null} />)

      // Initially timeline panel
      const timelinePanel = screen.getByRole('tabpanel')
      expect(timelinePanel).toHaveAttribute('id', 'timeline-panel')

      // Switch to logs
      const logsTab = screen.getByRole('tab', { name: /logs/i })
      fireEvent.click(logsTab)

      const logsPanel = screen.getByRole('tabpanel')
      expect(logsPanel).toHaveAttribute('id', 'logs-panel')
    })
  })

  describe('Log viewer node selection', () => {
    beforeEach(() => {
      vi.mocked(useExecutionTraceStore).mockReturnValue({
        currentExecutionId: 'exec-123',
        nodeStatuses: {},
      } as any)
    })

    it('should pass null selected node ID to log viewer', () => {
      render(<ExecutionDetailsPanel selectedNodeId={null} />)

      const logsTab = screen.getByRole('tab', { name: /logs/i })
      fireEvent.click(logsTab)

      expect(screen.getByText(/Log Viewer - none/)).toBeInTheDocument()
    })

    it('should pass specific node ID to log viewer', () => {
      render(<ExecutionDetailsPanel selectedNodeId="test-node-123" />)

      const logsTab = screen.getByRole('tab', { name: /logs/i })
      fireEvent.click(logsTab)

      expect(screen.getByText(/Log Viewer - test-node-123/)).toBeInTheDocument()
    })
  })

  describe('Tab button styling', () => {
    beforeEach(() => {
      vi.mocked(useExecutionTraceStore).mockReturnValue({
        currentExecutionId: 'exec-123',
        nodeStatuses: {},
      } as any)
    })

    it('should have active class on selected tab', () => {
      render(<ExecutionDetailsPanel selectedNodeId={null} />)

      const timelineTab = screen.getByRole('tab', { name: /timeline/i })
      expect(timelineTab).toHaveClass('active')
    })

    it('should not have active class on unselected tab', () => {
      render(<ExecutionDetailsPanel selectedNodeId={null} />)

      const logsTab = screen.getByRole('tab', { name: /logs/i })
      expect(logsTab).not.toHaveClass('active')
    })

    it('should update active class when switching tabs', () => {
      render(<ExecutionDetailsPanel selectedNodeId={null} />)

      const timelineTab = screen.getByRole('tab', { name: /timeline/i })
      const logsTab = screen.getByRole('tab', { name: /logs/i })

      fireEvent.click(logsTab)

      expect(logsTab).toHaveClass('active')
      expect(timelineTab).not.toHaveClass('active')
    })
  })
})
