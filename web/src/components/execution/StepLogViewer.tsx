import { useState } from 'react'
import { useExecutionTraceStore } from '../../stores/executionTraceStore'
import type { StepInfo } from '../../lib/websocket'
import '../../styles/executionTrace.css'

interface StepLogViewerProps {
  selectedNodeId: string | null
}

/**
 * Format timestamp for display
 */
function formatTimestamp(timestamp: string): string {
  const date = new Date(timestamp)
  return date.toLocaleString('en-US', {
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit',
    hour12: false,
  })
}

/**
 * Copy button with feedback
 */
function CopyButton({ data }: { data: any }) {
  const [copied, setCopied] = useState(false)

  const handleCopy = async () => {
    try {
      await navigator.clipboard.writeText(JSON.stringify(data, null, 2))
      setCopied(true)
      setTimeout(() => setCopied(false), 2000)
    } catch (error) {
      console.error('Failed to copy:', error)
    }
  }

  return (
    <button
      onClick={handleCopy}
      className="copy-button"
      aria-label="Copy output data to clipboard"
      type="button"
    >
      {copied ? (
        <>
          <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
          </svg>
          <span>Copied!</span>
        </>
      ) : (
        <>
          <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              strokeWidth={2}
              d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z"
            />
          </svg>
          <span>Copy</span>
        </>
      )}
    </button>
  )
}

/**
 * Render a single step log entry
 */
function StepLogEntry({ step, index }: { step: StepInfo; index: number }) {
  return (
    <div className="step-log-entry" data-testid={`step-log-${index}`}>
      <div className="step-header">
        <div className="step-title">
          <span className="step-id">{step.step_id}</span>
          <span className={`step-status status-${step.status}`}>{step.status}</span>
        </div>
        <span className="step-type">{step.node_type}</span>
      </div>

      {/* Timestamps */}
      {(step.started_at || step.completed_at) && (
        <div className="step-timestamps">
          {step.started_at && (
            <div className="timestamp-item">
              <span className="timestamp-label">Started at:</span>
              <span className="timestamp-value">{formatTimestamp(step.started_at)}</span>
            </div>
          )}
          {step.completed_at && (
            <div className="timestamp-item">
              <span className="timestamp-label">Completed at:</span>
              <span className="timestamp-value">{formatTimestamp(step.completed_at)}</span>
            </div>
          )}
        </div>
      )}

      {/* Duration */}
      {step.duration_ms !== undefined && (
        <div className="step-duration">
          <span className="duration-label">Duration:</span>
          <span className="duration-value">{step.duration_ms}ms</span>
        </div>
      )}

      {/* Error section */}
      {step.error && (
        <div className="error-section" data-testid="error-section">
          <div className="error-header">
            <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={2}
                d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"
              />
            </svg>
            <span>Error</span>
          </div>
          <pre className="error-message">{step.error}</pre>
        </div>
      )}

      {/* Output data */}
      {step.output_data && (
        <div className="output-section">
          <div className="output-header">
            <span className="output-label">Output Data</span>
            <CopyButton data={step.output_data} />
          </div>
          <pre className="output-data" role="code">
            <code>{JSON.stringify(step.output_data, null, 2)}</code>
          </pre>
        </div>
      )}
    </div>
  )
}

/**
 * StepLogViewer displays detailed logs for a selected node
 *
 * Features:
 * - Shows input/output data with JSON syntax highlighting
 * - Displays error messages
 * - Shows timestamps and duration
 * - Copy to clipboard functionality
 * - Supports multiple steps per node
 */
export function StepLogViewer({ selectedNodeId }: StepLogViewerProps) {
  const { stepLogs } = useExecutionTraceStore()

  // Empty state: no node selected
  if (!selectedNodeId) {
    return (
      <div className="log-viewer-empty">
        <div className="empty-icon">
          <svg className="w-12 h-12" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              strokeWidth={1.5}
              d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"
            />
          </svg>
        </div>
        <p className="empty-title">No node selected</p>
        <p className="empty-subtitle">Select a node to view its execution logs</p>
      </div>
    )
  }

  const logs = stepLogs[selectedNodeId] || []

  // Empty state: no logs for selected node
  if (logs.length === 0) {
    return (
      <div className="log-viewer-empty">
        <div className="empty-icon">
          <svg className="w-12 h-12" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              strokeWidth={1.5}
              d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"
            />
          </svg>
        </div>
        <p className="empty-title">No logs available</p>
        <p className="empty-subtitle">This node has not been executed yet</p>
      </div>
    )
  }

  return (
    <div className="log-viewer-container">
      <h3 className="log-viewer-title">Step Logs - {selectedNodeId}</h3>
      <div className="log-viewer-content">
        {logs.map((step, index) => (
          <StepLogEntry key={`${step.step_id}-${index}`} step={step} index={index} />
        ))}
      </div>
    </div>
  )
}
