import { useState, useEffect } from 'react'

interface ConfirmBulkDialogProps {
  open: boolean
  action: string
  count: number
  destructive?: boolean
  message?: string
  onConfirm: () => void
  onCancel: () => void
}

export function ConfirmBulkDialog({
  open,
  action,
  count,
  destructive = false,
  message,
  onConfirm,
  onCancel,
}: ConfirmBulkDialogProps) {
  const [confirmText, setConfirmText] = useState('')

  useEffect(() => {
    if (!open) {
      setConfirmText('')
    }
  }, [open])

  if (!open) {
    return null
  }

  const canConfirm = !destructive || confirmText === 'DELETE'
  const actionCapitalized = action.charAt(0).toUpperCase() + action.slice(1)

  const defaultMessage = `Are you sure you want to ${action} ${count} ${count === 1 ? 'item' : 'items'}?`

  return (
    <div
      className="fixed inset-0 bg-black/50 flex items-center justify-center z-50"
      onClick={onCancel}
      role="dialog"
      aria-modal="true"
    >
      <div
        className="bg-gray-800 rounded-lg p-6 max-w-md w-full mx-4 shadow-xl"
        onClick={(e) => e.stopPropagation()}
      >
        <h2 className="text-xl font-bold text-white mb-4">
          {actionCapitalized} {count} {count === 1 ? 'item' : 'items'}
        </h2>

        <p className="text-gray-300 mb-6">
          {message || defaultMessage}
        </p>

        {destructive && !message && (
          <p className="text-yellow-500 text-sm mb-6 font-semibold">
            This action cannot be undone.
          </p>
        )}

        {destructive && (
          <div className="mb-6">
            <label htmlFor="confirm-input" className="block text-sm text-gray-400 mb-2">
              Type <span className="font-bold text-white">DELETE</span> to confirm
            </label>
            <input
              id="confirm-input"
              type="text"
              value={confirmText}
              onChange={(e) => setConfirmText(e.target.value)}
              placeholder="Type DELETE to confirm"
              className="w-full px-3 py-2 bg-gray-700 text-white rounded-lg border border-gray-600 focus:outline-none focus:ring-2 focus:ring-red-500"
            />
          </div>
        )}

        <div className="flex justify-end space-x-3">
          <button
            onClick={onCancel}
            className="px-4 py-2 bg-gray-700 text-white rounded-lg hover:bg-gray-600 transition-colors"
          >
            Cancel
          </button>
          <button
            onClick={onConfirm}
            disabled={!canConfirm}
            className={`px-4 py-2 rounded-lg transition-colors ${
              destructive
                ? 'bg-red-600 hover:bg-red-700 text-white disabled:bg-gray-700 disabled:text-gray-500'
                : 'bg-primary-600 hover:bg-primary-700 text-white disabled:bg-gray-700 disabled:text-gray-500'
            } disabled:cursor-not-allowed`}
          >
            {actionCapitalized}
          </button>
        </div>
      </div>
    </div>
  )
}
