import { useState } from 'react'
import { useParams, Link, Navigate } from 'react-router-dom'
import { useWebhook, useWebhookEvents, useWebhookMutations } from '../hooks/useWebhooks'
import { useWorkflows } from '../hooks/useWorkflows'
import type { WebhookAuthType } from '../api/webhooks'
import FilterBuilder from '../components/webhooks/FilterBuilder'
import { isValidResourceId } from '../utils/routing'

export default function WebhookDetail() {
  const { id } = useParams<{ id: string }>()

  // Hooks must be called unconditionally before any early returns
  const { webhook, loading, error, refetch } = useWebhook(id || '')
  const { events, loading: eventsLoading } = useWebhookEvents(id || '', { limit: 10 })
  const { updateWebhook, regenerateSecret, updating, regenerating } = useWebhookMutations()
  const { workflows } = useWorkflows()
  const [showRegenerateConfirm, setShowRegenerateConfirm] = useState(false)
  const [newSecret, setNewSecret] = useState<string | null>(null)
  const [copyMessage, setCopyMessage] = useState<string | null>(null)
  const [actionError, setActionError] = useState<string | null>(null)

  // Guard against invalid IDs (after all hooks are called)
  if (!isValidResourceId(id)) {
    return <Navigate to="/webhooks" replace />
  }

  const getWorkflowName = (workflowId: string) => {
    const workflow = workflows.find((w) => w.id === workflowId)
    return workflow?.name || 'Unknown Workflow'
  }

  const handleCopyURL = async () => {
    if (!webhook) return
    try {
      await navigator.clipboard.writeText(webhook.url)
      setCopyMessage('URL copied!')
      setTimeout(() => setCopyMessage(null), 3000)
    } catch {
      setCopyMessage('Failed to copy')
      setTimeout(() => setCopyMessage(null), 3000)
    }
  }

  const handleCopySecret = async () => {
    if (!newSecret) return
    try {
      await navigator.clipboard.writeText(newSecret)
      setCopyMessage('Secret copied!')
      setTimeout(() => setCopyMessage(null), 3000)
    } catch {
      setCopyMessage('Failed to copy')
      setTimeout(() => setCopyMessage(null), 3000)
    }
  }

  const handleToggleEnabled = async () => {
    if (!webhook) return
    try {
      setActionError(null)
      await updateWebhook(webhook.id, { enabled: !webhook.enabled })
      await refetch()
    } catch (err: any) {
      setActionError(err.message || 'Failed to update')
    }
  }

  const handleRegenerateSecret = async () => {
    if (!webhook) return
    try {
      setActionError(null)
      const result = await regenerateSecret(webhook.id)
      setNewSecret(result.secret)
      setShowRegenerateConfirm(false)
    } catch (err: any) {
      setActionError(err.message || 'Failed to regenerate secret')
    }
  }

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleString()
  }

  const formatRelativeTime = (dateString?: string) => {
    if (!dateString) return 'Never'
    const date = new Date(dateString)
    const now = new Date()
    const diffMs = now.getTime() - date.getTime()
    const diffMins = Math.floor(diffMs / 60000)
    const diffHours = Math.floor(diffMins / 60)
    const diffDays = Math.floor(diffHours / 24)

    if (diffMins < 60) return `${diffMins} min${diffMins !== 1 ? 's' : ''} ago`
    if (diffHours < 24) return `${diffHours} hour${diffHours !== 1 ? 's' : ''} ago`
    return `${diffDays} day${diffDays !== 1 ? 's' : ''} ago`
  }

  if (loading) {
    return (
      <div className="h-64 flex items-center justify-center">
        <div className="text-white text-lg">Loading webhook...</div>
      </div>
    )
  }

  if (error) {
    return (
      <div className="h-64 flex items-center justify-center">
        <div className="text-center">
          <div className="text-red-400 text-lg mb-4">Failed to fetch webhook</div>
          <div className="text-gray-400 text-sm">{error.message}</div>
        </div>
      </div>
    )
  }

  if (!webhook) {
    return (
      <div className="h-64 flex items-center justify-center">
        <div className="text-gray-400 text-lg">Webhook not found</div>
      </div>
    )
  }

  return (
    <div>
      {/* Header */}
      <div className="mb-6">
        <Link
          to="/webhooks"
          className="text-gray-400 hover:text-white text-sm mb-2 inline-block"
        >
          &larr; Back to Webhooks
        </Link>
        <div className="flex items-center justify-between">
          <h1 className="text-2xl font-bold text-white">{webhook.name}</h1>
          <Link
            to={`/webhooks/${webhook.id}/edit`}
            className="px-4 py-2 bg-primary-600 text-white rounded-lg text-sm font-medium hover:bg-primary-700 transition-colors"
          >
            Edit Webhook
          </Link>
        </div>
      </div>

      {/* Messages */}
      {copyMessage && (
        <div className="mb-4 p-3 bg-green-800/50 rounded-lg text-green-200 text-sm">
          {copyMessage}
        </div>
      )}
      {actionError && (
        <div className="mb-4 p-3 bg-red-800/50 rounded-lg text-red-200 text-sm">
          {actionError}
        </div>
      )}

      {/* Main Content */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* Details Panel */}
        <div className="lg:col-span-2 space-y-6">
          {/* Basic Info */}
          <div className="bg-gray-800 rounded-lg p-6">
            <h2 className="text-lg font-semibold text-white mb-4">Details</h2>
            <dl className="space-y-4">
              <div>
                <dt className="text-sm text-gray-400">Path</dt>
                <dd className="text-white font-mono">{webhook.path}</dd>
              </div>
              <div>
                <dt className="text-sm text-gray-400">Full URL</dt>
                <dd className="flex items-center space-x-2">
                  <span className="text-white font-mono text-sm break-all">{webhook.url}</span>
                  <button
                    onClick={handleCopyURL}
                    className="px-2 py-1 text-xs bg-gray-700 text-gray-300 rounded hover:bg-gray-600 transition-colors"
                  >
                    Copy
                  </button>
                </dd>
              </div>
              <div>
                <dt className="text-sm text-gray-400">Authentication</dt>
                <dd>
                  <AuthTypeBadge authType={webhook.authType} />
                </dd>
              </div>
              <div>
                <dt className="text-sm text-gray-400">Linked Workflow</dt>
                <dd>
                  <Link
                    to={`/workflows/${webhook.workflowId}`}
                    className="text-primary-400 hover:text-primary-300"
                  >
                    {getWorkflowName(webhook.workflowId)}
                  </Link>
                </dd>
              </div>
              <div>
                <dt className="text-sm text-gray-400">Enabled</dt>
                <dd>
                  <ToggleSwitch
                    enabled={webhook.enabled}
                    disabled={updating}
                    onChange={handleToggleEnabled}
                  />
                </dd>
              </div>
            </dl>
          </div>

          {/* Secret Management */}
          {webhook.authType === 'signature' && (
            <div className="bg-gray-800 rounded-lg p-6">
              <h2 className="text-lg font-semibold text-white mb-4">Webhook Secret</h2>
              {newSecret ? (
                <div className="space-y-3">
                  <div className="p-3 bg-yellow-900/30 border border-yellow-600/50 rounded-lg">
                    <p className="text-yellow-300 text-sm mb-2">
                      Save this secret - it won't be shown again!
                    </p>
                    <div className="flex items-center space-x-2">
                      <code className="flex-1 text-white font-mono text-sm bg-gray-900 p-2 rounded break-all">
                        {newSecret}
                      </code>
                      <button
                        onClick={handleCopySecret}
                        className="px-3 py-2 bg-primary-600 text-white rounded text-sm hover:bg-primary-700"
                      >
                        Copy
                      </button>
                    </div>
                  </div>
                  <button
                    onClick={() => setNewSecret(null)}
                    className="text-gray-400 text-sm hover:text-white"
                  >
                    Dismiss
                  </button>
                </div>
              ) : (
                <div>
                  <p className="text-gray-400 text-sm mb-3">
                    Use this secret to verify webhook signatures. Regenerating will invalidate the
                    current secret.
                  </p>
                  <button
                    onClick={() => setShowRegenerateConfirm(true)}
                    disabled={regenerating}
                    className="px-4 py-2 bg-gray-700 text-white rounded-lg text-sm font-medium hover:bg-gray-600 disabled:opacity-50 transition-colors"
                  >
                    {regenerating ? 'Regenerating...' : 'Regenerate Secret'}
                  </button>
                </div>
              )}
            </div>
          )}

          {/* Event Filters */}
          <div className="bg-gray-800 rounded-lg p-6">
            <div className="mb-4">
              <h2 className="text-lg font-semibold text-white mb-2">Event Filters</h2>
              <p className="text-gray-400 text-sm">
                Configure filters to control which webhook events trigger workflow execution. Events
                that don't match all filters will be ignored.
              </p>
            </div>
            <FilterBuilder webhookId={webhook.id} />
          </div>

          {/* Event History */}
          <div className="bg-gray-800 rounded-lg p-6">
            <div className="flex items-center justify-between mb-4">
              <h2 className="text-lg font-semibold text-white">Event History</h2>
              <span className="text-sm text-gray-400">Last 10 events</span>
            </div>
            {eventsLoading ? (
              <div className="text-gray-400">Loading events...</div>
            ) : events.length === 0 ? (
              <div className="text-gray-400">No events yet</div>
            ) : (
              <div className="space-y-2">
                {events.map((event) => (
                  <div
                    key={event.id}
                    className="flex items-center justify-between p-3 bg-gray-700/50 rounded-lg"
                  >
                    <div className="flex items-center space-x-3">
                      <StatusBadge status={event.status} />
                      <span className="text-white text-sm">{event.requestMethod}</span>
                      <span className="text-gray-400 text-sm">
                        {formatRelativeTime(event.createdAt)}
                      </span>
                    </div>
                    <div className="text-gray-400 text-sm">
                      {event.processingTimeMs ? `${event.processingTimeMs}ms` : '-'}
                    </div>
                  </div>
                ))}
              </div>
            )}
          </div>
        </div>

        {/* Stats Sidebar */}
        <div className="space-y-6">
          <div className="bg-gray-800 rounded-lg p-6">
            <h2 className="text-lg font-semibold text-white mb-4">Statistics</h2>
            <dl className="space-y-4">
              <div>
                <dt className="text-sm text-gray-400">Total Triggers</dt>
                <dd className="text-2xl font-bold text-white">{webhook.triggerCount}</dd>
              </div>
              <div>
                <dt className="text-sm text-gray-400">Last Triggered</dt>
                <dd className="text-white">{formatRelativeTime(webhook.lastTriggeredAt)}</dd>
              </div>
              <div>
                <dt className="text-sm text-gray-400">Created</dt>
                <dd className="text-white text-sm">{formatDate(webhook.createdAt)}</dd>
              </div>
              <div>
                <dt className="text-sm text-gray-400">Last Updated</dt>
                <dd className="text-white text-sm">{formatDate(webhook.updatedAt)}</dd>
              </div>
            </dl>
          </div>

          <div className="bg-gray-800 rounded-lg p-6">
            <h2 className="text-lg font-semibold text-white mb-4">Priority</h2>
            <PriorityBadge priority={webhook.priority} />
          </div>
        </div>
      </div>

      {/* Regenerate Confirmation Modal */}
      {showRegenerateConfirm && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
          <div className="bg-gray-800 rounded-lg p-6 w-96">
            <h3 className="text-white text-lg font-semibold mb-4">Regenerate Secret</h3>
            <p className="text-gray-400 mb-4">
              Are you sure you want to regenerate the webhook secret? The current secret will be
              invalidated immediately.
            </p>
            <div className="flex space-x-2 justify-end">
              <button
                onClick={() => setShowRegenerateConfirm(false)}
                disabled={regenerating}
                className="px-4 py-2 bg-gray-700 text-white rounded-lg text-sm font-medium hover:bg-gray-600 transition-colors disabled:opacity-50"
              >
                Cancel
              </button>
              <button
                onClick={handleRegenerateSecret}
                disabled={regenerating}
                className="px-4 py-2 bg-primary-600 text-white rounded-lg text-sm font-medium hover:bg-primary-700 transition-colors disabled:opacity-50"
              >
                {regenerating ? 'Regenerating...' : 'Confirm'}
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}

function AuthTypeBadge({ authType }: { authType: WebhookAuthType }) {
  const colors = {
    none: 'bg-gray-500/20 text-gray-400',
    signature: 'bg-blue-500/20 text-blue-400',
    basic: 'bg-purple-500/20 text-purple-400',
    api_key: 'bg-green-500/20 text-green-400',
  }

  return (
    <span className={`inline-flex px-2 py-1 text-xs font-medium rounded-full ${colors[authType]}`}>
      {authType}
    </span>
  )
}

function StatusBadge({ status }: { status: string }) {
  const colors: Record<string, string> = {
    received: 'bg-gray-500/20 text-gray-400',
    processed: 'bg-green-500/20 text-green-400',
    filtered: 'bg-yellow-500/20 text-yellow-400',
    failed: 'bg-red-500/20 text-red-400',
  }

  return (
    <span
      className={`inline-flex px-2 py-1 text-xs font-medium rounded-full ${
        colors[status] || colors.received
      }`}
    >
      {status}
    </span>
  )
}

function PriorityBadge({ priority }: { priority: number }) {
  const levels = ['Low', 'Normal', 'High', 'Critical']
  const colors = [
    'bg-gray-500/20 text-gray-400',
    'bg-blue-500/20 text-blue-400',
    'bg-yellow-500/20 text-yellow-400',
    'bg-red-500/20 text-red-400',
  ]

  const level = Math.min(priority, 3)

  return (
    <span className={`inline-flex px-3 py-1 text-sm font-medium rounded-full ${colors[level]}`}>
      {levels[level]}
    </span>
  )
}

interface ToggleSwitchProps {
  enabled: boolean
  disabled?: boolean
  onChange: () => void
}

function ToggleSwitch({ enabled, disabled, onChange }: ToggleSwitchProps) {
  return (
    <button
      type="button"
      role="switch"
      aria-checked={enabled}
      disabled={disabled}
      onClick={onChange}
      className={`relative inline-flex h-6 w-11 items-center rounded-full transition-colors focus:outline-none focus:ring-2 focus:ring-primary-500 focus:ring-offset-2 focus:ring-offset-gray-800 disabled:opacity-50 disabled:cursor-not-allowed ${
        enabled ? 'bg-primary-600' : 'bg-gray-600'
      }`}
    >
      <span
        className={`inline-block h-4 w-4 transform rounded-full bg-white transition-transform ${
          enabled ? 'translate-x-6' : 'translate-x-1'
        }`}
      />
    </button>
  )
}
