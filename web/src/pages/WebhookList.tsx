import { useState, useMemo } from 'react'
import { Link } from 'react-router-dom'
import { useWebhooks, useWebhookMutations } from '../hooks/useWebhooks'
import { useWorkflows } from '../hooks/useWorkflows'
import type { WebhookAuthType } from '../api/webhooks'
import PriorityBadge from '../components/webhooks/PriorityBadge'

const PAGE_SIZE = 20

export default function WebhookList() {
  const [page, setPage] = useState(1)

  const params = useMemo(() => {
    return { page, limit: PAGE_SIZE }
  }, [page])

  const { webhooks, total, loading, error, refetch } = useWebhooks(params)
  const { deleteWebhook, updateWebhook, deleting, updating } = useWebhookMutations()
  const { workflows } = useWorkflows()

  const [deleteConfirm, setDeleteConfirm] = useState<string | null>(null)
  const [deleteError, setDeleteError] = useState<string | null>(null)
  const [copyMessage, setCopyMessage] = useState<string | null>(null)
  const [toggleError, setToggleError] = useState<string | null>(null)

  const totalPages = Math.ceil(total / PAGE_SIZE)

  const handleDeleteClick = (webhookId: string) => {
    setDeleteConfirm(webhookId)
    setDeleteError(null)
  }

  const handleDeleteConfirm = async () => {
    if (!deleteConfirm) return

    try {
      await deleteWebhook(deleteConfirm)
      setDeleteConfirm(null)
      setDeleteError(null)
      await refetch()
    } catch (error: any) {
      setDeleteError(error.message || 'Delete failed')
    }
  }

  const handleDeleteCancel = () => {
    setDeleteConfirm(null)
    setDeleteError(null)
  }

  const handleCopyURL = async (url: string) => {
    try {
      await navigator.clipboard.writeText(url)
      setCopyMessage('URL copied to clipboard!')
      setTimeout(() => setCopyMessage(null), 3000)
    } catch (error) {
      setCopyMessage('Failed to copy URL')
      setTimeout(() => setCopyMessage(null), 3000)
    }
  }

  const handleToggleEnabled = async (webhookId: string, currentEnabled: boolean) => {
    try {
      setToggleError(null)
      await updateWebhook(webhookId, { enabled: !currentEnabled })
      await refetch()
    } catch (error: any) {
      setToggleError(error.message || 'Toggle failed')
      setTimeout(() => setToggleError(null), 3000)
    }
  }

  const getWorkflowName = (workflowId: string) => {
    const workflow = workflows.find((w) => w.id === workflowId)
    return workflow?.name || 'Unknown Workflow'
  }

  const formatRelativeTime = (dateString?: string) => {
    if (!dateString) return 'Never'

    const date = new Date(dateString)
    const now = new Date()
    const diffMs = now.getTime() - date.getTime()
    const diffSecs = Math.floor(diffMs / 1000)
    const diffMins = Math.floor(diffSecs / 60)
    const diffHours = Math.floor(diffMins / 60)
    const diffDays = Math.floor(diffHours / 24)

    if (diffSecs < 60) return 'just now'
    if (diffMins < 60) return `${diffMins} min${diffMins > 1 ? 's' : ''} ago`
    if (diffHours < 24) return `${diffHours} hour${diffHours > 1 ? 's' : ''} ago`
    if (diffDays < 30) return `${diffDays} day${diffDays > 1 ? 's' : ''} ago`

    return date.toLocaleDateString()
  }

  // Loading state
  if (loading) {
    return (
      <div className="h-64 flex items-center justify-center">
        <div className="text-white text-lg">Loading webhooks...</div>
      </div>
    )
  }

  // Error state
  if (error) {
    return (
      <div className="h-64 flex items-center justify-center">
        <div className="text-center">
          <div className="text-red-400 text-lg mb-4">Failed to fetch webhooks</div>
          <div className="text-gray-400 text-sm">{error.message}</div>
        </div>
      </div>
    )
  }

  // Empty state
  if (webhooks.length === 0) {
    return (
      <div>
        <div className="flex justify-between items-center mb-6">
          <h1 className="text-2xl font-bold text-white">Webhooks</h1>
          <Link
            to="/webhooks/new"
            className="px-4 py-2 bg-primary-600 text-white rounded-lg text-sm font-medium hover:bg-primary-700 transition-colors"
          >
            Create Webhook
          </Link>
        </div>

        <div className="h-64 flex items-center justify-center bg-gray-800 rounded-lg">
          <div className="text-center">
            <div className="text-gray-400 text-lg mb-4">No webhooks found</div>
            <Link
              to="/webhooks/new"
              className="px-4 py-2 bg-primary-600 text-white rounded-lg text-sm font-medium hover:bg-primary-700 transition-colors inline-block"
            >
              Create your first webhook
            </Link>
          </div>
        </div>
      </div>
    )
  }

  return (
    <div>
      <div className="flex justify-between items-center mb-6">
        <h1 className="text-2xl font-bold text-white">Webhooks</h1>
        <Link
          to="/webhooks/new"
          className="px-4 py-2 bg-primary-600 text-white rounded-lg text-sm font-medium hover:bg-primary-700 transition-colors"
        >
          Create Webhook
        </Link>
      </div>

      {copyMessage && (
        <div className="mb-4 p-3 bg-gray-800 rounded-lg text-white text-sm">
          {copyMessage}
        </div>
      )}

      {toggleError && (
        <div className="mb-4 p-3 bg-red-800/50 rounded-lg text-red-200 text-sm">
          {toggleError}
        </div>
      )}

      <div className="bg-gray-800 rounded-lg overflow-hidden">
        <table className="w-full">
          <thead>
            <tr className="border-b border-gray-700">
              <th className="text-left px-6 py-4 text-sm font-medium text-gray-400">Name</th>
              <th className="text-left px-6 py-4 text-sm font-medium text-gray-400">Path</th>
              <th className="text-left px-6 py-4 text-sm font-medium text-gray-400">Workflow</th>
              <th className="text-left px-6 py-4 text-sm font-medium text-gray-400">Auth Type</th>
              <th className="text-left px-6 py-4 text-sm font-medium text-gray-400">Priority</th>
              <th className="text-left px-6 py-4 text-sm font-medium text-gray-400">Enabled</th>
              <th className="text-left px-6 py-4 text-sm font-medium text-gray-400">
                Trigger Count
              </th>
              <th className="text-left px-6 py-4 text-sm font-medium text-gray-400">
                Last Triggered
              </th>
              <th className="text-right px-6 py-4 text-sm font-medium text-gray-400">Actions</th>
            </tr>
          </thead>
          <tbody>
            {webhooks.map((webhook) => (
              <tr key={webhook.id} className="border-b border-gray-700 hover:bg-gray-700/50">
                <td className="px-6 py-4">
                  <p className="text-white font-medium">{webhook.name}</p>
                </td>
                <td className="px-6 py-4 text-gray-300 font-mono text-sm">{webhook.path}</td>
                <td className="px-6 py-4">
                  <Link
                    to={`/workflows/${webhook.workflowId}`}
                    className="text-primary-400 hover:text-primary-300 transition-colors"
                  >
                    {getWorkflowName(webhook.workflowId)}
                  </Link>
                </td>
                <td className="px-6 py-4">
                  <AuthTypeBadge authType={webhook.authType} />
                </td>
                <td className="px-6 py-4">
                  <PriorityBadge priority={webhook.priority} />
                </td>
                <td className="px-6 py-4">
                  <ToggleSwitch
                    enabled={webhook.enabled}
                    disabled={updating}
                    onChange={() => handleToggleEnabled(webhook.id, webhook.enabled)}
                  />
                </td>
                <td className="px-6 py-4 text-gray-300">{webhook.triggerCount}</td>
                <td className="px-6 py-4 text-gray-300 text-sm">
                  {formatRelativeTime(webhook.lastTriggeredAt)}
                </td>
                <td className="px-6 py-4 text-right">
                  <div className="flex justify-end space-x-2">
                    <Link
                      to={`/webhooks/${webhook.id}`}
                      className="px-3 py-1 text-sm text-gray-300 hover:text-white transition-colors"
                    >
                      Edit
                    </Link>
                    <button
                      onClick={() => handleCopyURL(webhook.url)}
                      className="px-3 py-1 text-sm text-primary-400 hover:text-primary-300 transition-colors"
                    >
                      Copy URL
                    </button>
                    <button
                      onClick={() => handleDeleteClick(webhook.id)}
                      disabled={deleting}
                      className="px-3 py-1 text-sm text-red-400 hover:text-red-300 transition-colors disabled:opacity-50"
                    >
                      Delete
                    </button>
                  </div>
                </td>
              </tr>
            ))}
          </tbody>
        </table>

        {/* Pagination */}
        {totalPages > 1 && (
          <div className="px-4 py-3 bg-gray-700 flex items-center justify-between">
            <div className="text-sm text-gray-400">
              Page {page} of {totalPages}
            </div>
            <div className="flex space-x-2">
              <button
                onClick={() => setPage((p) => Math.max(1, p - 1))}
                disabled={page === 1}
                className="px-3 py-1 bg-gray-600 text-white rounded text-sm hover:bg-gray-500 disabled:opacity-50 disabled:cursor-not-allowed"
              >
                Previous
              </button>
              <button
                onClick={() => setPage((p) => Math.min(totalPages, p + 1))}
                disabled={page === totalPages}
                className="px-3 py-1 bg-gray-600 text-white rounded text-sm hover:bg-gray-500 disabled:opacity-50 disabled:cursor-not-allowed"
              >
                Next
              </button>
            </div>
          </div>
        )}
      </div>

      {/* Delete Confirmation Dialog */}
      {deleteConfirm && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
          <div className="bg-gray-800 rounded-lg p-6 w-96">
            <h3 className="text-white text-lg font-semibold mb-4">Delete Webhook</h3>
            <p className="text-gray-400 mb-4">
              Are you sure you want to delete this webhook? This action cannot be undone.
            </p>
            {deleteError && <div className="text-xs text-red-400 mb-4">{deleteError}</div>}
            <div className="flex space-x-2 justify-end">
              <button
                onClick={handleDeleteCancel}
                disabled={deleting}
                className="px-4 py-2 bg-gray-700 text-white rounded-lg text-sm font-medium hover:bg-gray-600 transition-colors disabled:opacity-50"
              >
                Cancel
              </button>
              <button
                onClick={handleDeleteConfirm}
                disabled={deleting}
                className="px-4 py-2 bg-red-600 text-white rounded-lg text-sm font-medium hover:bg-red-700 transition-colors disabled:opacity-50"
              >
                {deleting ? 'Deleting...' : 'Confirm'}
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
    <span
      className={`inline-flex px-2 py-1 text-xs font-medium rounded-full ${
        colors[authType]
      }`}
    >
      {authType}
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
