import { useState } from 'react'
import { webhookAPI, type WebhookEvent } from '../../api/webhooks'

const MAX_REPLAY_COUNT = 5

interface ReplayModalProps {
  event: WebhookEvent
  onClose: () => void
  onSuccess: (executionId: string) => void
}

export function ReplayModal({ event, onClose, onSuccess }: ReplayModalProps) {
  const [payload, setPayload] = useState(JSON.stringify(event.requestBody, null, 2))
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const isAtMaxReplay = event.replayCount >= MAX_REPLAY_COUNT
  const replaysRemaining = MAX_REPLAY_COUNT - event.replayCount

  const handleBackdropClick = (e: React.MouseEvent) => {
    if (e.target === e.currentTarget) {
      onClose()
    }
  }

  const handleReplay = async () => {
    setError(null)

    // Validate JSON
    let parsedPayload: unknown
    try {
      parsedPayload = JSON.parse(payload)
    } catch {
      setError('Invalid JSON format')
      return
    }

    setLoading(true)

    try {
      const result = await webhookAPI.replayEvent(event.id, parsedPayload)

      if (result.success && result.executionId) {
        onSuccess(result.executionId)
        onClose()
      } else {
        setError(result.error || 'Replay failed')
      }
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Network error'
      setError(errorMessage)
    } finally {
      setLoading(false)
    }
  }

  return (
    <div
      data-testid="replay-modal-backdrop"
      onClick={handleBackdropClick}
      className="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4"
    >
      <div
        role="dialog"
        aria-modal="true"
        aria-labelledby="replay-modal-title"
        className="bg-gray-800 rounded-lg p-6 max-w-3xl w-full max-h-[90vh] overflow-y-auto"
      >
        {/* Header */}
        <div className="flex justify-between items-start mb-6">
          <div>
            <h2 id="replay-modal-title" className="text-xl font-bold text-white mb-1">
              Replay Webhook Event
            </h2>
            <p className="text-sm text-gray-400">Event ID: {event.id}</p>
            <p className="text-sm text-gray-400">
              Replayed {event.replayCount} time{event.replayCount !== 1 ? 's' : ''}
            </p>
          </div>
          <button
            onClick={onClose}
            disabled={loading}
            className="text-gray-400 hover:text-white transition-colors disabled:opacity-50"
            aria-label="Close"
          >
            <svg className="w-6 h-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={2}
                d="M6 18L18 6M6 6l12 12"
              />
            </svg>
          </button>
        </div>

        {/* Warning Messages */}
        {isAtMaxReplay && (
          <div className="mb-4 p-3 bg-red-900/30 border border-red-600/50 rounded-lg">
            <p className="text-red-300 text-sm font-medium">Cannot replay this event</p>
            <p className="text-red-400 text-sm">
              Maximum replay limit ({MAX_REPLAY_COUNT}) has been reached.
            </p>
          </div>
        )}

        {!isAtMaxReplay && replaysRemaining <= 1 && (
          <div className="mb-4 p-3 bg-yellow-900/30 border border-yellow-600/50 rounded-lg">
            <p className="text-yellow-300 text-sm">
              {replaysRemaining} replay remaining before limit is reached.
            </p>
          </div>
        )}

        {/* Error Message */}
        {error && (
          <div className="mb-4 p-3 bg-red-900/30 border border-red-600/50 rounded-lg">
            <p className="text-red-300 text-sm">{error}</p>
          </div>
        )}

        {/* Payload Editor */}
        <div className="mb-6">
          <label htmlFor="payload-editor" className="block text-sm font-medium text-gray-400 mb-2">
            Payload
          </label>
          <textarea
            id="payload-editor"
            value={payload}
            onChange={(e) => setPayload(e.target.value)}
            disabled={loading || isAtMaxReplay}
            className="w-full h-80 px-4 py-3 bg-gray-900 text-white rounded-lg border border-gray-700 focus:outline-none focus:border-primary-500 font-mono text-sm disabled:opacity-50 disabled:cursor-not-allowed"
            placeholder="Enter JSON payload..."
          />
          <p className="mt-2 text-xs text-gray-500">
            You can modify the payload before replaying the event.
          </p>
        </div>

        {/* Actions */}
        <div className="flex justify-end gap-3">
          <button
            onClick={onClose}
            disabled={loading}
            className="px-4 py-2 bg-gray-700 text-white rounded-lg text-sm font-medium hover:bg-gray-600 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
          >
            Cancel
          </button>
          <button
            onClick={handleReplay}
            disabled={loading || isAtMaxReplay}
            className="px-4 py-2 bg-primary-600 text-white rounded-lg text-sm font-medium hover:bg-primary-700 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
          >
            {loading ? 'Replaying...' : 'Replay Event'}
          </button>
        </div>
      </div>
    </div>
  )
}
