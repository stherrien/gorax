/**
 * Example component demonstrating WebSocket usage for real-time execution monitoring
 *
 * This is a reference implementation showing how to use the useExecutionUpdates hook
 * to display live execution progress.
 */

import { useState } from 'react'
import { useExecutionUpdates } from '../hooks/useExecutionUpdates'

interface ExecutionMonitorProps {
  executionId: string
  onComplete?: () => void
  onError?: (error: string) => void
}

export function ExecutionMonitor({
  executionId,
  onComplete,
  onError,
}: ExecutionMonitorProps) {
  const {
    connected,
    reconnecting,
    reconnectAttempt,
    currentStatus,
    currentProgress,
    completedSteps,
    latestUpdate,
  } = useExecutionUpdates(executionId, {
    enabled: true,
    onStatusChange: (status) => {
      console.log('[ExecutionMonitor] Status changed:', status)
    },
    onProgress: (progress) => {
      console.log(
        `[ExecutionMonitor] Progress: ${progress.completed_steps}/${progress.total_steps} (${progress.percentage.toFixed(1)}%)`
      )
    },
    onStepComplete: (step) => {
      console.log('[ExecutionMonitor] Step completed:', step.node_id)
    },
    onComplete: (output) => {
      console.log('[ExecutionMonitor] Execution completed:', output)
      onComplete?.()
    },
    onError: (error) => {
      console.error('[ExecutionMonitor] Execution failed:', error)
      onError?.(error)
    },
  })

  return (
    <div className="execution-monitor">
      {/* Connection Status */}
      <div className="connection-status">
        {connected && !reconnecting && (
          <span className="status-indicator connected">🟢 Connected</span>
        )}
        {reconnecting && (
          <span className="status-indicator reconnecting">
            🟡 Reconnecting (attempt {reconnectAttempt})...
          </span>
        )}
        {!connected && !reconnecting && (
          <span className="status-indicator disconnected">🔴 Disconnected</span>
        )}
      </div>

      {/* Execution Status */}
      <div className="execution-status">
        <h3>Execution Status</h3>
        <div className="status-badge" data-status={currentStatus}>
          {currentStatus || 'pending'}
        </div>
      </div>

      {/* Progress Bar */}
      {currentProgress && (
        <div className="progress-section">
          <h3>Progress</h3>
          <div className="progress-bar">
            <div
              className="progress-fill"
              style={{ width: `${currentProgress.percentage}%` }}
            />
          </div>
          <div className="progress-text">
            {currentProgress.completed_steps} / {currentProgress.total_steps} steps
            ({currentProgress.percentage.toFixed(1)}%)
          </div>
        </div>
      )}

      {/* Completed Steps */}
      {completedSteps.length > 0 && (
        <div className="completed-steps">
          <h3>Completed Steps</h3>
          <ul className="steps-list">
            {completedSteps.map((step, index) => (
              <li key={step.node_id || index} className="step-item">
                <div className="step-header">
                  <span className="step-icon">✓</span>
                  <span className="step-name">{step.node_id}</span>
                  <span className="step-type">{step.node_type}</span>
                </div>
                {step.duration_ms && (
                  <div className="step-duration">{step.duration_ms}ms</div>
                )}
              </li>
            ))}
          </ul>
        </div>
      )}

      {/* Latest Update (for debugging) */}
      {latestUpdate && (
        <details className="latest-update">
          <summary>Latest Update ({latestUpdate.type})</summary>
          <pre>{JSON.stringify(latestUpdate, null, 2)}</pre>
        </details>
      )}
    </div>
  )
}

/**
 * Example: Minimal usage
 */
export function MinimalExecutionMonitor({ executionId }: { executionId: string }) {
  const { connected, currentStatus, currentProgress } = useExecutionUpdates(executionId)

  return (
    <div>
      <div>Status: {connected ? '🟢' : '🔴'}</div>
      <div>Execution: {currentStatus || 'pending'}</div>
      {currentProgress && (
        <div>
          Progress: {currentProgress.completed_steps}/{currentProgress.total_steps}
        </div>
      )}
    </div>
  )
}

/**
 * Example: With callbacks
 */
export function ExecutionMonitorWithCallbacks({ executionId }: { executionId: string }) {
  const [notifications, setNotifications] = useState<string[]>([])

  const addNotification = (message: string) => {
    setNotifications((prev) => [...prev, message])
  }

  const { currentProgress } = useExecutionUpdates(executionId, {
    onStatusChange: (status) => addNotification(`Status changed to: ${status}`),
    onStepComplete: (step) => addNotification(`Step completed: ${step.node_id}`),
    onComplete: () => addNotification('Execution completed successfully!'),
    onError: (error) => addNotification(`Error: ${error}`),
  })

  return (
    <div>
      {currentProgress && (
        <progress
          value={currentProgress.completed_steps}
          max={currentProgress.total_steps}
        />
      )}
      <ul>
        {notifications.map((msg, i) => (
          <li key={i}>{msg}</li>
        ))}
      </ul>
    </div>
  )
}

/**
 * Example: Dashboard view (monitoring multiple executions)
 */
export function ExecutionDashboard() {
  const [activeExecutions] = useState<string[]>([
    'exec-1',
    'exec-2',
    'exec-3',
  ])

  return (
    <div className="execution-dashboard">
      <h2>Active Executions</h2>
      <div className="execution-grid">
        {activeExecutions.map((execId) => (
          <ExecutionCard key={execId} executionId={execId} />
        ))}
      </div>
    </div>
  )
}

function ExecutionCard({ executionId }: { executionId: string }) {
  const { connected, currentStatus, currentProgress } = useExecutionUpdates(executionId)

  return (
    <div className={`execution-card ${currentStatus}`}>
      <div className="card-header">
        <span className="execution-id">{executionId}</span>
        <span className={`connection-dot ${connected ? 'connected' : 'disconnected'}`} />
      </div>
      <div className="card-body">
        <div className="status">{currentStatus}</div>
        {currentProgress && (
          <div className="progress-mini">
            <div
              className="progress-mini-bar"
              style={{ width: `${currentProgress.percentage}%` }}
            />
            <span className="progress-mini-text">
              {currentProgress.percentage.toFixed(0)}%
            </span>
          </div>
        )}
      </div>
    </div>
  )
}

/**
 * Example CSS (for reference)
 * Exported for documentation purposes
 */
export const exampleCSS = `
.execution-monitor {
  padding: 1rem;
  border: 1px solid #ddd;
  border-radius: 8px;
}

.connection-status {
  margin-bottom: 1rem;
  padding: 0.5rem;
  background: #f5f5f5;
  border-radius: 4px;
}

.status-indicator {
  font-size: 14px;
  font-weight: 500;
}

.status-indicator.connected {
  color: #22c55e;
}

.status-indicator.reconnecting {
  color: #eab308;
}

.status-indicator.disconnected {
  color: #ef4444;
}

.status-badge {
  display: inline-block;
  padding: 0.25rem 0.75rem;
  border-radius: 4px;
  font-weight: 500;
  text-transform: uppercase;
  font-size: 12px;
}

.status-badge[data-status="pending"] {
  background: #e2e8f0;
  color: #64748b;
}

.status-badge[data-status="running"] {
  background: #dbeafe;
  color: #3b82f6;
}

.status-badge[data-status="completed"] {
  background: #dcfce7;
  color: #22c55e;
}

.status-badge[data-status="failed"] {
  background: #fee2e2;
  color: #ef4444;
}

.progress-bar {
  width: 100%;
  height: 8px;
  background: #e2e8f0;
  border-radius: 4px;
  overflow: hidden;
}

.progress-fill {
  height: 100%;
  background: #3b82f6;
  transition: width 0.3s ease;
}

.progress-text {
  margin-top: 0.5rem;
  font-size: 14px;
  color: #64748b;
}

.steps-list {
  list-style: none;
  padding: 0;
  margin: 0;
}

.step-item {
  padding: 0.75rem;
  border: 1px solid #e2e8f0;
  border-radius: 4px;
  margin-bottom: 0.5rem;
}

.step-header {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.step-icon {
  color: #22c55e;
}

.step-name {
  font-weight: 500;
}

.step-type {
  color: #64748b;
  font-size: 12px;
}

.step-duration {
  margin-top: 0.25rem;
  font-size: 12px;
  color: #64748b;
}
`
