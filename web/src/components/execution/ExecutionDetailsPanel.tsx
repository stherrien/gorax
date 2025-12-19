import { useState, KeyboardEvent } from 'react'
import { useExecutionTraceStore } from '../../stores/executionTraceStore'
import { ExecutionTimeline } from './ExecutionTimeline'
import { StepLogViewer } from './StepLogViewer'
import '../../styles/executionTrace.css'

interface ExecutionDetailsPanelProps {
  selectedNodeId: string | null
}

type TabType = 'timeline' | 'logs'

/**
 * Get overall execution status from node statuses
 */
function getExecutionStatus(nodeStatuses: Record<string, string>): string {
  const statuses = Object.values(nodeStatuses)

  if (statuses.length === 0) return 'idle'
  if (statuses.some((s) => s === 'failed')) return 'failed'
  if (statuses.some((s) => s === 'running')) return 'running'
  if (statuses.every((s) => s === 'completed')) return 'completed'

  return 'pending'
}

/**
 * ExecutionDetailsPanel combines timeline and log viewer in a tabbed interface
 *
 * Features:
 * - Tabs for switching between timeline and logs
 * - Sticky header with execution ID and status
 * - Keyboard navigation support
 * - Fully accessible with ARIA labels
 */
export function ExecutionDetailsPanel({ selectedNodeId }: ExecutionDetailsPanelProps) {
  const { currentExecutionId, nodeStatuses } = useExecutionTraceStore()
  const [activeTab, setActiveTab] = useState<TabType>('timeline')

  // Don't render if no execution is active
  if (!currentExecutionId) {
    return null
  }

  const executionStatus = getExecutionStatus(nodeStatuses)

  const handleTabClick = (tab: TabType) => {
    setActiveTab(tab)
  }

  const handleTabKeyDown = (e: KeyboardEvent<HTMLButtonElement>, tab: TabType) => {
    if (e.key === 'Enter' || e.key === ' ') {
      e.preventDefault()
      setActiveTab(tab)
    } else if (e.key === 'ArrowRight') {
      e.preventDefault()
      const nextTab = tab === 'timeline' ? 'logs' : 'timeline'
      setActiveTab(nextTab)
      // Focus next tab
      const nextButton = e.currentTarget.parentElement?.querySelector(
        `[data-tab="${nextTab}"]`
      ) as HTMLButtonElement
      nextButton?.focus()
    } else if (e.key === 'ArrowLeft') {
      e.preventDefault()
      const prevTab = tab === 'timeline' ? 'logs' : 'timeline'
      setActiveTab(prevTab)
      // Focus previous tab
      const prevButton = e.currentTarget.parentElement?.querySelector(
        `[data-tab="${prevTab}"]`
      ) as HTMLButtonElement
      prevButton?.focus()
    }
  }

  return (
    <div className="execution-details-panel" data-testid="execution-details-panel">
      {/* Sticky Header */}
      <div className="panel-header sticky" data-testid="panel-header">
        <div className="header-content">
          <div className="execution-info">
            <h2 className="execution-id">{currentExecutionId}</h2>
            <span className={`execution-status status-${executionStatus}`}>
              {executionStatus}
            </span>
          </div>
        </div>

        {/* Tab Navigation */}
        <div className="tab-list" role="tablist" aria-label="Execution details">
          <button
            role="tab"
            data-tab="timeline"
            aria-selected={activeTab === 'timeline'}
            aria-controls="timeline-panel"
            className={`tab-button ${activeTab === 'timeline' ? 'active' : ''}`}
            onClick={() => handleTabClick('timeline')}
            onKeyDown={(e) => handleTabKeyDown(e, 'timeline')}
            type="button"
          >
            <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={2}
                d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z"
              />
            </svg>
            Timeline
          </button>

          <button
            role="tab"
            data-tab="logs"
            aria-selected={activeTab === 'logs'}
            aria-controls="logs-panel"
            className={`tab-button ${activeTab === 'logs' ? 'active' : ''}`}
            onClick={() => handleTabClick('logs')}
            onKeyDown={(e) => handleTabKeyDown(e, 'logs')}
            type="button"
          >
            <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={2}
                d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"
              />
            </svg>
            Logs
          </button>
        </div>
      </div>

      {/* Tab Panels */}
      <div className="panel-content">
        {activeTab === 'timeline' && (
          <div
            role="tabpanel"
            id="timeline-panel"
            aria-labelledby="timeline-tab"
            className="tab-panel"
          >
            <ExecutionTimeline />
          </div>
        )}

        {activeTab === 'logs' && (
          <div
            role="tabpanel"
            id="logs-panel"
            aria-labelledby="logs-tab"
            className="tab-panel"
          >
            <StepLogViewer selectedNodeId={selectedNodeId} />
          </div>
        )}
      </div>
    </div>
  )
}
