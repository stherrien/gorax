import { useEffect, useRef } from 'react'
import { useExecutionTraceStore } from '../../stores/executionTraceStore'
import type { TimelineEventType } from '../../stores/executionTraceStore'
import '../../styles/executionTrace.css'

/**
 * Get icon component for event type
 */
function EventIcon({ type }: { type: TimelineEventType }) {
  const iconClass = 'w-4 h-4'

  switch (type) {
    case 'started':
      return (
        <div
          className="event-icon event-icon-started"
          data-testid="event-icon-started"
          aria-label="Started event"
        >
          <svg className={iconClass} fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              strokeWidth={2}
              d="M14.752 11.168l-3.197-2.132A1 1 0 0010 9.87v4.263a1 1 0 001.555.832l3.197-2.132a1 1 0 000-1.664z"
            />
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              strokeWidth={2}
              d="M21 12a9 9 0 11-18 0 9 9 0 0118 0z"
            />
          </svg>
        </div>
      )

    case 'completed':
      return (
        <div
          className="event-icon event-icon-completed"
          data-testid="event-icon-completed"
          aria-label="Completed event"
        >
          <svg className={iconClass} fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={3} d="M5 13l4 4L19 7" />
          </svg>
        </div>
      )

    case 'failed':
      return (
        <div
          className="event-icon event-icon-failed"
          data-testid="event-icon-failed"
          aria-label="Failed event"
        >
          <svg className={iconClass} fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
          </svg>
        </div>
      )

    case 'progress':
      return (
        <div
          className="event-icon event-icon-progress"
          data-testid="event-icon-progress"
          aria-label="Progress event"
        >
          <svg className={iconClass} fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              strokeWidth={2}
              d="M13 10V3L4 14h7v7l9-11h-7z"
            />
          </svg>
        </div>
      )
  }
}

/**
 * Format timestamp for display
 */
function formatTimestamp(timestamp: string): string {
  const date = new Date(timestamp)
  return date.toLocaleTimeString('en-US', {
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit',
    hour12: false,
  })
}

/**
 * Get CSS class for event type
 */
function getEventClass(type: TimelineEventType): string {
  return `event-${type}`
}

/**
 * Render metadata if present
 */
function EventMetadata({ metadata }: { metadata?: Record<string, unknown> }) {
  if (!metadata || Object.keys(metadata).length === 0) {
    return null
  }

  return (
    <div className="event-metadata">
      {Object.entries(metadata).map(([key, value]) => (
        <div key={key} className="metadata-item">
          <span className="metadata-key">{key}:</span>
          <span className="metadata-value">{JSON.stringify(value)}</span>
        </div>
      ))}
    </div>
  )
}

/**
 * ExecutionTimeline displays a chronological list of execution events
 *
 * Features:
 * - Shows event type, timestamp, node name, status icon
 * - Auto-scrolls to latest event
 * - Color-coded by status
 * - Accessible with ARIA labels
 * - Displays metadata when present
 */
export function ExecutionTimeline() {
  const { timelineEvents } = useExecutionTraceStore()
  const latestEventRef = useRef<HTMLLIElement>(null)

  // Auto-scroll to latest event when new events arrive
  useEffect(() => {
    if (latestEventRef.current && latestEventRef.current.scrollIntoView) {
      latestEventRef.current.scrollIntoView({
        behavior: 'smooth',
        block: 'nearest',
      })
    }
  }, [timelineEvents.length])

  // Empty state
  if (timelineEvents.length === 0) {
    return (
      <div className="timeline-empty" data-testid="timeline-empty">
        <div className="empty-icon">
          <svg className="w-12 h-12" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              strokeWidth={1.5}
              d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2"
            />
          </svg>
        </div>
        <p className="empty-title">No execution events</p>
        <p className="empty-subtitle">Start an execution to see timeline</p>
      </div>
    )
  }

  return (
    <div className="timeline-container overflow-y-auto" data-testid="timeline-container">
      <ul className="timeline-list" role="list" aria-label="Execution timeline">
        {timelineEvents.map((event, index) => {
          const isLatest = index === timelineEvents.length - 1
          const eventClass = getEventClass(event.type)

          return (
            <li
              key={`${event.nodeId}-${event.timestamp}-${index}`}
              ref={isLatest ? latestEventRef : null}
              className={`timeline-event ${eventClass}`}
              data-testid={`timeline-event-${index}`}
              data-latest={isLatest}
              role="listitem"
            >
              <div className="event-icon-wrapper">
                <EventIcon type={event.type} />
              </div>

              <div className="event-content">
                <div className="event-header">
                  <span className="event-message">{event.message}</span>
                  <span className="event-timestamp" data-testid="event-timestamp">
                    {formatTimestamp(event.timestamp)}
                  </span>
                </div>

                <div className="event-details">
                  <span className="event-node-id">Node: {event.nodeId}</span>
                </div>

                <EventMetadata metadata={event.metadata} />
              </div>
            </li>
          )
        })}
      </ul>
    </div>
  )
}
