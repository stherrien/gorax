import { describe, it, expect, beforeEach, vi } from 'vitest'
import { render, screen, within } from '@testing-library/react'
import { ExecutionTimeline } from './ExecutionTimeline'
import { useExecutionTraceStore } from '../../stores/executionTraceStore'
import type { TimelineEvent } from '../../stores/executionTraceStore'

// Mock the execution trace store
vi.mock('../../stores/executionTraceStore', () => ({
  useExecutionTraceStore: vi.fn(),
}))

describe('ExecutionTimeline', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  describe('Empty state', () => {
    it('should render empty state when no events exist', () => {
      vi.mocked(useExecutionTraceStore).mockReturnValue({
        timelineEvents: [],
      } as any)

      render(<ExecutionTimeline />)

      expect(screen.getByText(/no execution events/i)).toBeInTheDocument()
    })

    it('should display helpful message in empty state', () => {
      vi.mocked(useExecutionTraceStore).mockReturnValue({
        timelineEvents: [],
      } as any)

      render(<ExecutionTimeline />)

      expect(screen.getByText(/start an execution to see timeline/i)).toBeInTheDocument()
    })
  })

  describe('Timeline rendering', () => {
    const mockEvents: TimelineEvent[] = [
      {
        timestamp: '2025-01-01T12:00:00Z',
        nodeId: 'node-1',
        type: 'started',
        message: 'Execution started',
      },
      {
        timestamp: '2025-01-01T12:00:01Z',
        nodeId: 'node-2',
        type: 'progress',
        message: 'Processing step 1',
      },
      {
        timestamp: '2025-01-01T12:00:02Z',
        nodeId: 'node-3',
        type: 'completed',
        message: 'Execution completed successfully',
      },
    ]

    it('should render all timeline events', () => {
      vi.mocked(useExecutionTraceStore).mockReturnValue({
        timelineEvents: mockEvents,
      } as any)

      render(<ExecutionTimeline />)

      expect(screen.getByText('Execution started')).toBeInTheDocument()
      expect(screen.getByText('Processing step 1')).toBeInTheDocument()
      expect(screen.getByText('Execution completed successfully')).toBeInTheDocument()
    })

    it('should display events in chronological order', () => {
      vi.mocked(useExecutionTraceStore).mockReturnValue({
        timelineEvents: mockEvents,
      } as any)

      render(<ExecutionTimeline />)

      const events = screen.getAllByTestId(/timeline-event/)
      expect(events).toHaveLength(3)
      expect(within(events[0]).getByText('Execution started')).toBeInTheDocument()
      expect(within(events[2]).getByText('Execution completed successfully')).toBeInTheDocument()
    })

    it('should show formatted timestamps for each event', () => {
      vi.mocked(useExecutionTraceStore).mockReturnValue({
        timelineEvents: [mockEvents[0]],
      } as any)

      render(<ExecutionTimeline />)

      expect(screen.getByTestId('event-timestamp')).toBeInTheDocument()
    })

    it('should display node IDs for each event', () => {
      vi.mocked(useExecutionTraceStore).mockReturnValue({
        timelineEvents: [mockEvents[0]],
      } as any)

      render(<ExecutionTimeline />)

      expect(screen.getByText(/node-1/i)).toBeInTheDocument()
    })
  })

  describe('Event type indicators', () => {
    it('should show started icon for started events', () => {
      const startedEvent: TimelineEvent = {
        timestamp: '2025-01-01T12:00:00Z',
        nodeId: 'node-1',
        type: 'started',
        message: 'Started',
      }

      vi.mocked(useExecutionTraceStore).mockReturnValue({
        timelineEvents: [startedEvent],
      } as any)

      render(<ExecutionTimeline />)

      expect(screen.getByTestId('event-icon-started')).toBeInTheDocument()
    })

    it('should show completed icon for completed events', () => {
      const completedEvent: TimelineEvent = {
        timestamp: '2025-01-01T12:00:00Z',
        nodeId: 'node-1',
        type: 'completed',
        message: 'Completed',
      }

      vi.mocked(useExecutionTraceStore).mockReturnValue({
        timelineEvents: [completedEvent],
      } as any)

      render(<ExecutionTimeline />)

      expect(screen.getByTestId('event-icon-completed')).toBeInTheDocument()
    })

    it('should show failed icon for failed events', () => {
      const failedEvent: TimelineEvent = {
        timestamp: '2025-01-01T12:00:00Z',
        nodeId: 'node-1',
        type: 'failed',
        message: 'Failed',
      }

      vi.mocked(useExecutionTraceStore).mockReturnValue({
        timelineEvents: [failedEvent],
      } as any)

      render(<ExecutionTimeline />)

      expect(screen.getByTestId('event-icon-failed')).toBeInTheDocument()
    })

    it('should show progress icon for progress events', () => {
      const progressEvent: TimelineEvent = {
        timestamp: '2025-01-01T12:00:00Z',
        nodeId: 'node-1',
        type: 'progress',
        message: 'In progress',
      }

      vi.mocked(useExecutionTraceStore).mockReturnValue({
        timelineEvents: [progressEvent],
      } as any)

      render(<ExecutionTimeline />)

      expect(screen.getByTestId('event-icon-progress')).toBeInTheDocument()
    })
  })

  describe('Color coding', () => {
    it('should apply green color for completed events', () => {
      const completedEvent: TimelineEvent = {
        timestamp: '2025-01-01T12:00:00Z',
        nodeId: 'node-1',
        type: 'completed',
        message: 'Completed',
      }

      vi.mocked(useExecutionTraceStore).mockReturnValue({
        timelineEvents: [completedEvent],
      } as any)

      render(<ExecutionTimeline />)

      const event = screen.getByTestId('timeline-event-0')
      expect(event).toHaveClass('event-completed')
    })

    it('should apply red color for failed events', () => {
      const failedEvent: TimelineEvent = {
        timestamp: '2025-01-01T12:00:00Z',
        nodeId: 'node-1',
        type: 'failed',
        message: 'Failed',
      }

      vi.mocked(useExecutionTraceStore).mockReturnValue({
        timelineEvents: [failedEvent],
      } as any)

      render(<ExecutionTimeline />)

      const event = screen.getByTestId('timeline-event-0')
      expect(event).toHaveClass('event-failed')
    })

    it('should apply blue color for started events', () => {
      const startedEvent: TimelineEvent = {
        timestamp: '2025-01-01T12:00:00Z',
        nodeId: 'node-1',
        type: 'started',
        message: 'Started',
      }

      vi.mocked(useExecutionTraceStore).mockReturnValue({
        timelineEvents: [startedEvent],
      } as any)

      render(<ExecutionTimeline />)

      const event = screen.getByTestId('timeline-event-0')
      expect(event).toHaveClass('event-started')
    })
  })

  describe('Auto-scroll behavior', () => {
    it('should have a scrollable container', () => {
      vi.mocked(useExecutionTraceStore).mockReturnValue({
        timelineEvents: [
          {
            timestamp: '2025-01-01T12:00:00Z',
            nodeId: 'node-1',
            type: 'started',
            message: 'Started',
          },
        ],
      } as any)

      render(<ExecutionTimeline />)

      const container = screen.getByTestId('timeline-container')
      expect(container).toHaveClass('overflow-y-auto')
    })

    it('should mark latest event for auto-scroll', () => {
      const mockEvents: TimelineEvent[] = [
        {
          timestamp: '2025-01-01T12:00:00Z',
          nodeId: 'node-1',
          type: 'started',
          message: 'Event 1',
        },
        {
          timestamp: '2025-01-01T12:00:01Z',
          nodeId: 'node-2',
          type: 'progress',
          message: 'Event 2',
        },
      ]

      vi.mocked(useExecutionTraceStore).mockReturnValue({
        timelineEvents: mockEvents,
      } as any)

      render(<ExecutionTimeline />)

      const lastEvent = screen.getByTestId('timeline-event-1')
      expect(lastEvent).toHaveAttribute('data-latest', 'true')
    })
  })

  describe('Accessibility', () => {
    it('should have proper ARIA labels for timeline', () => {
      vi.mocked(useExecutionTraceStore).mockReturnValue({
        timelineEvents: [
          {
            timestamp: '2025-01-01T12:00:00Z',
            nodeId: 'node-1',
            type: 'started',
            message: 'Started',
          },
        ],
      } as any)

      render(<ExecutionTimeline />)

      const timeline = screen.getByRole('list', { name: /execution timeline/i })
      expect(timeline).toBeInTheDocument()
    })

    it('should mark each event as a list item', () => {
      vi.mocked(useExecutionTraceStore).mockReturnValue({
        timelineEvents: [
          {
            timestamp: '2025-01-01T12:00:00Z',
            nodeId: 'node-1',
            type: 'started',
            message: 'Started',
          },
          {
            timestamp: '2025-01-01T12:00:01Z',
            nodeId: 'node-2',
            type: 'completed',
            message: 'Completed',
          },
        ],
      } as any)

      render(<ExecutionTimeline />)

      const items = screen.getAllByRole('listitem')
      expect(items).toHaveLength(2)
    })

    it('should have descriptive alt text for event icons', () => {
      const failedEvent: TimelineEvent = {
        timestamp: '2025-01-01T12:00:00Z',
        nodeId: 'node-1',
        type: 'failed',
        message: 'Failed',
      }

      vi.mocked(useExecutionTraceStore).mockReturnValue({
        timelineEvents: [failedEvent],
      } as any)

      render(<ExecutionTimeline />)

      const icon = screen.getByTestId('event-icon-failed')
      expect(icon).toHaveAttribute('aria-label', 'Failed event')
    })
  })

  describe('Metadata display', () => {
    it('should display metadata when present', () => {
      const eventWithMetadata: TimelineEvent = {
        timestamp: '2025-01-01T12:00:00Z',
        nodeId: 'node-1',
        type: 'completed',
        message: 'Completed',
        metadata: {
          duration: 1500,
          status: 'success',
        },
      }

      vi.mocked(useExecutionTraceStore).mockReturnValue({
        timelineEvents: [eventWithMetadata],
      } as any)

      render(<ExecutionTimeline />)

      expect(screen.getByText(/duration/i)).toBeInTheDocument()
      expect(screen.getByText(/1500/i)).toBeInTheDocument()
    })

    it('should not show metadata section when metadata is absent', () => {
      const eventWithoutMetadata: TimelineEvent = {
        timestamp: '2025-01-01T12:00:00Z',
        nodeId: 'node-1',
        type: 'started',
        message: 'Started',
      }

      vi.mocked(useExecutionTraceStore).mockReturnValue({
        timelineEvents: [eventWithoutMetadata],
      } as any)

      render(<ExecutionTimeline />)

      expect(screen.queryByText(/metadata/i)).not.toBeInTheDocument()
    })
  })
})
