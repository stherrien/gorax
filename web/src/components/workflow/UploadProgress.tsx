/**
 * UploadProgress - Progress indicator for workflow file uploads
 *
 * Shows upload status, progress bar, file information, and allows cancellation.
 */

import type { UploadStatus, UploadResult } from '../../hooks/useFileUpload'
import { formatFileSize } from '../../utils/fileValidation'

interface UploadProgressProps {
  /** Current upload state */
  uploadState: UploadResult
  /** Callback to cancel/reset the upload */
  onCancel?: () => void
  /** Callback to accept the uploaded workflow */
  onAccept?: () => void
  /** Callback to dismiss the success state */
  onDismiss?: () => void
  /** Whether to show as a floating notification or inline */
  variant?: 'floating' | 'inline'
}

/**
 * Get status text for display
 */
function getStatusText(status: UploadStatus): string {
  switch (status) {
    case 'validating':
      return 'Validating file...'
    case 'reading':
      return 'Reading file...'
    case 'parsing':
      return 'Parsing workflow...'
    case 'success':
      return 'Upload complete!'
    case 'error':
      return 'Upload failed'
    default:
      return ''
  }
}

/**
 * Get status icon
 */
function StatusIcon({ status }: { status: UploadStatus }) {
  if (status === 'success') {
    return (
      <svg
        className="w-5 h-5 text-green-400"
        fill="none"
        stroke="currentColor"
        viewBox="0 0 24 24"
        aria-hidden="true"
      >
        <path
          strokeLinecap="round"
          strokeLinejoin="round"
          strokeWidth={2}
          d="M5 13l4 4L19 7"
        />
      </svg>
    )
  }

  if (status === 'error') {
    return (
      <svg
        className="w-5 h-5 text-red-400"
        fill="none"
        stroke="currentColor"
        viewBox="0 0 24 24"
        aria-hidden="true"
      >
        <path
          strokeLinecap="round"
          strokeLinejoin="round"
          strokeWidth={2}
          d="M6 18L18 6M6 6l12 12"
        />
      </svg>
    )
  }

  // Loading spinner
  return (
    <svg
      className="w-5 h-5 text-primary-400 animate-spin"
      fill="none"
      viewBox="0 0 24 24"
      aria-hidden="true"
    >
      <circle
        className="opacity-25"
        cx="12"
        cy="12"
        r="10"
        stroke="currentColor"
        strokeWidth="4"
      />
      <path
        className="opacity-75"
        fill="currentColor"
        d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
      />
    </svg>
  )
}

/**
 * UploadProgress component
 */
export function UploadProgress({
  uploadState,
  onCancel,
  onAccept,
  onDismiss,
  variant = 'floating',
}: UploadProgressProps) {
  const { status, progress, fileName, fileSize, nodes, warnings } = uploadState

  // Don't render if idle
  if (status === 'idle') {
    return null
  }

  const isProcessing = ['validating', 'reading', 'parsing'].includes(status)
  const isComplete = status === 'success'
  const hasError = status === 'error'

  const containerClasses = variant === 'floating'
    ? 'fixed bottom-4 right-4 z-50 w-80 bg-gray-800 border border-gray-700 rounded-lg shadow-xl'
    : 'w-full bg-gray-800 border border-gray-700 rounded-lg'

  return (
    <div className={containerClasses} role="status" aria-live="polite">
      {/* Header */}
      <div className="flex items-center justify-between p-3 border-b border-gray-700">
        <div className="flex items-center gap-2">
          <StatusIcon status={status} />
          <span className="text-sm font-medium text-gray-200">
            {getStatusText(status)}
          </span>
        </div>

        {/* Close button for completed/error states */}
        {(isComplete || hasError) && (onCancel || onDismiss) && (
          <button
            type="button"
            onClick={hasError ? onCancel : onDismiss}
            className="text-gray-400 hover:text-gray-200 transition-colors"
            aria-label="Close"
          >
            <svg
              className="w-4 h-4"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={2}
                d="M6 18L18 6M6 6l12 12"
              />
            </svg>
          </button>
        )}
      </div>

      {/* Content */}
      <div className="p-3">
        {/* File info */}
        {fileName && (
          <div className="flex items-center gap-2 mb-3">
            <svg
              className="w-4 h-4 text-gray-400 flex-shrink-0"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={2}
                d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"
              />
            </svg>
            <div className="min-w-0 flex-1">
              <p className="text-sm text-gray-200 truncate" title={fileName}>
                {fileName}
              </p>
              {fileSize && (
                <p className="text-xs text-gray-400">
                  {formatFileSize(fileSize)}
                </p>
              )}
            </div>
          </div>
        )}

        {/* Progress bar */}
        {isProcessing && (
          <div className="mb-3">
            <div className="h-1.5 bg-gray-700 rounded-full overflow-hidden">
              <div
                className="h-full bg-primary-500 rounded-full transition-all duration-300 ease-out"
                style={{ width: `${progress}%` }}
              />
            </div>
            <p className="text-xs text-gray-400 mt-1 text-right">
              {progress}%
            </p>
          </div>
        )}

        {/* Success info */}
        {isComplete && nodes && (
          <div className="mb-3">
            <p className="text-sm text-green-400">
              Successfully parsed {nodes.length} node{nodes.length !== 1 ? 's' : ''}
            </p>

            {/* Warnings */}
            {warnings && warnings.length > 0 && (
              <div className="mt-2 p-2 bg-yellow-900/20 border border-yellow-500/30 rounded">
                <p className="text-xs text-yellow-400 font-medium mb-1">
                  Warnings:
                </p>
                <ul className="text-xs text-yellow-300 space-y-0.5">
                  {warnings.slice(0, 3).map((warning, index) => (
                    <li key={index} className="truncate" title={warning}>
                      {warning}
                    </li>
                  ))}
                  {warnings.length > 3 && (
                    <li className="text-yellow-400">
                      +{warnings.length - 3} more...
                    </li>
                  )}
                </ul>
              </div>
            )}
          </div>
        )}

        {/* Actions */}
        <div className="flex items-center gap-2">
          {isProcessing && onCancel && (
            <button
              type="button"
              onClick={onCancel}
              className="
                flex-1 px-3 py-1.5
                bg-gray-700 hover:bg-gray-600
                text-gray-200 text-sm font-medium
                rounded transition-colors
              "
            >
              Cancel
            </button>
          )}

          {isComplete && (
            <>
              {onAccept && (
                <button
                  type="button"
                  onClick={onAccept}
                  className="
                    flex-1 px-3 py-1.5
                    bg-green-600 hover:bg-green-700
                    text-white text-sm font-medium
                    rounded transition-colors
                  "
                >
                  Import to Canvas
                </button>
              )}
              {onDismiss && (
                <button
                  type="button"
                  onClick={onDismiss}
                  className="
                    px-3 py-1.5
                    bg-gray-700 hover:bg-gray-600
                    text-gray-200 text-sm font-medium
                    rounded transition-colors
                  "
                >
                  Cancel
                </button>
              )}
            </>
          )}
        </div>
      </div>
    </div>
  )
}

/**
 * Inline progress bar (simpler version)
 */
interface InlineProgressBarProps {
  progress: number
  status: UploadStatus
}

export function InlineProgressBar({ progress, status }: InlineProgressBarProps) {
  if (status === 'idle' || status === 'success' || status === 'error') {
    return null
  }

  return (
    <div className="w-full">
      <div className="flex items-center justify-between mb-1">
        <span className="text-xs text-gray-400">{getStatusText(status)}</span>
        <span className="text-xs text-gray-400">{progress}%</span>
      </div>
      <div className="h-1 bg-gray-700 rounded-full overflow-hidden">
        <div
          className="h-full bg-primary-500 rounded-full transition-all duration-300"
          style={{ width: `${progress}%` }}
        />
      </div>
    </div>
  )
}

export default UploadProgress
