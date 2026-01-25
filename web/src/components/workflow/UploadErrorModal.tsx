/**
 * UploadErrorModal - Modal for displaying detailed upload error messages
 *
 * Shows validation errors, parsing errors, and suggestions for fixing issues.
 */

import { useCallback, useEffect, useRef } from 'react'

interface UploadErrorModalProps {
  /** Whether the modal is open */
  isOpen: boolean
  /** Error message title */
  error: string
  /** Detailed error information */
  errorDetails?: string[]
  /** Name of the file that failed */
  fileName?: string
  /** Callback to close the modal */
  onClose: () => void
  /** Callback to retry the upload */
  onRetry?: () => void
}

/**
 * UploadErrorModal component
 */
export function UploadErrorModal({
  isOpen,
  error,
  errorDetails,
  fileName,
  onClose,
  onRetry,
}: UploadErrorModalProps) {
  const modalRef = useRef<HTMLDivElement>(null)
  const closeButtonRef = useRef<HTMLButtonElement>(null)

  // Focus management
  useEffect(() => {
    if (isOpen) {
      closeButtonRef.current?.focus()
    }
  }, [isOpen])

  // Close on escape key
  useEffect(() => {
    if (!isOpen) return

    const handleKeyDown = (event: KeyboardEvent) => {
      if (event.key === 'Escape') {
        onClose()
      }
    }

    document.addEventListener('keydown', handleKeyDown)
    return () => document.removeEventListener('keydown', handleKeyDown)
  }, [isOpen, onClose])

  // Close when clicking outside
  const handleBackdropClick = useCallback(
    (event: React.MouseEvent) => {
      if (event.target === event.currentTarget) {
        onClose()
      }
    },
    [onClose]
  )

  if (!isOpen) {
    return null
  }

  return (
    <div
      className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/60 backdrop-blur-sm"
      onClick={handleBackdropClick}
      role="dialog"
      aria-modal="true"
      aria-labelledby="error-modal-title"
      aria-describedby="error-modal-description"
    >
      <div
        ref={modalRef}
        className="
          w-full max-w-md
          bg-gray-800 border border-gray-700
          rounded-lg shadow-2xl
          overflow-hidden
        "
      >
        {/* Header */}
        <div className="flex items-center justify-between p-4 border-b border-gray-700 bg-red-900/20">
          <div className="flex items-center gap-3">
            {/* Error icon */}
            <div className="flex-shrink-0 w-10 h-10 flex items-center justify-center rounded-full bg-red-500/20">
              <svg
                className="w-6 h-6 text-red-400"
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
                aria-hidden="true"
              >
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  strokeWidth={2}
                  d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z"
                />
              </svg>
            </div>
            <div>
              <h2
                id="error-modal-title"
                className="text-lg font-semibold text-red-300"
              >
                Import Failed
              </h2>
              {fileName && (
                <p className="text-sm text-gray-400 truncate max-w-[200px]" title={fileName}>
                  {fileName}
                </p>
              )}
            </div>
          </div>

          {/* Close button */}
          <button
            ref={closeButtonRef}
            type="button"
            onClick={onClose}
            className="
              text-gray-400 hover:text-gray-200
              transition-colors
              focus:outline-none focus:ring-2 focus:ring-red-500 rounded
            "
            aria-label="Close error dialog"
          >
            <svg
              className="w-5 h-5"
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
        </div>

        {/* Content */}
        <div className="p-4" id="error-modal-description">
          {/* Main error message */}
          <p className="text-gray-200 font-medium mb-3">{error}</p>

          {/* Error details */}
          {errorDetails && errorDetails.length > 0 && (
            <div className="mb-4">
              <p className="text-sm text-gray-400 mb-2">Details:</p>
              <ul className="space-y-2">
                {errorDetails.map((detail, index) => (
                  <li
                    key={index}
                    className="
                      flex items-start gap-2
                      text-sm text-gray-300
                      bg-gray-700/50 rounded px-3 py-2
                    "
                  >
                    <svg
                      className="w-4 h-4 text-gray-400 flex-shrink-0 mt-0.5"
                      fill="none"
                      stroke="currentColor"
                      viewBox="0 0 24 24"
                    >
                      <path
                        strokeLinecap="round"
                        strokeLinejoin="round"
                        strokeWidth={2}
                        d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"
                      />
                    </svg>
                    <span>{detail}</span>
                  </li>
                ))}
              </ul>
            </div>
          )}

          {/* Help section */}
          <div className="bg-gray-700/30 rounded-lg p-3 mb-4">
            <p className="text-sm text-gray-400 mb-2">Need help?</p>
            <ul className="text-sm text-gray-300 space-y-1">
              <li className="flex items-center gap-2">
                <span className="text-primary-400">&#x2022;</span>
                Ensure the file is a valid JSON or YAML workflow file
              </li>
              <li className="flex items-center gap-2">
                <span className="text-primary-400">&#x2022;</span>
                Check that the file contains <code className="text-xs bg-gray-700 px-1 rounded">nodes</code> and <code className="text-xs bg-gray-700 px-1 rounded">edges</code> arrays
              </li>
              <li className="flex items-center gap-2">
                <span className="text-primary-400">&#x2022;</span>
                Verify each node has a unique <code className="text-xs bg-gray-700 px-1 rounded">id</code> and <code className="text-xs bg-gray-700 px-1 rounded">type</code>
              </li>
            </ul>
          </div>
        </div>

        {/* Footer */}
        <div className="flex items-center justify-end gap-2 p-4 border-t border-gray-700 bg-gray-800/50">
          {onRetry && (
            <button
              type="button"
              onClick={onRetry}
              className="
                px-4 py-2
                bg-primary-600 hover:bg-primary-700
                text-white text-sm font-medium
                rounded-lg transition-colors
                focus:outline-none focus:ring-2 focus:ring-primary-500 focus:ring-offset-2 focus:ring-offset-gray-800
              "
            >
              Try Another File
            </button>
          )}
          <button
            type="button"
            onClick={onClose}
            className="
              px-4 py-2
              bg-gray-700 hover:bg-gray-600
              text-gray-200 text-sm font-medium
              rounded-lg transition-colors
              focus:outline-none focus:ring-2 focus:ring-gray-500 focus:ring-offset-2 focus:ring-offset-gray-800
            "
          >
            Close
          </button>
        </div>
      </div>
    </div>
  )
}

/**
 * Simple inline error display (for non-modal errors)
 */
interface UploadErrorInlineProps {
  error: string
  errorDetails?: string[]
  onDismiss?: () => void
  onRetry?: () => void
}

export function UploadErrorInline({
  error,
  errorDetails,
  onDismiss,
  onRetry,
}: UploadErrorInlineProps) {
  return (
    <div className="bg-red-900/20 border border-red-500/30 rounded-lg p-4">
      <div className="flex items-start gap-3">
        <svg
          className="w-5 h-5 text-red-400 flex-shrink-0 mt-0.5"
          fill="none"
          stroke="currentColor"
          viewBox="0 0 24 24"
        >
          <path
            strokeLinecap="round"
            strokeLinejoin="round"
            strokeWidth={2}
            d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"
          />
        </svg>

        <div className="flex-1 min-w-0">
          <p className="text-sm font-medium text-red-300">{error}</p>

          {errorDetails && errorDetails.length > 0 && (
            <ul className="mt-2 text-sm text-red-200 space-y-1">
              {errorDetails.slice(0, 3).map((detail, index) => (
                <li key={index} className="truncate" title={detail}>
                  {detail}
                </li>
              ))}
              {errorDetails.length > 3 && (
                <li className="text-red-400">
                  +{errorDetails.length - 3} more issues
                </li>
              )}
            </ul>
          )}

          {(onDismiss || onRetry) && (
            <div className="flex items-center gap-2 mt-3">
              {onRetry && (
                <button
                  type="button"
                  onClick={onRetry}
                  className="
                    text-sm text-red-300 hover:text-red-200
                    underline underline-offset-2
                    transition-colors
                  "
                >
                  Try again
                </button>
              )}
              {onDismiss && (
                <button
                  type="button"
                  onClick={onDismiss}
                  className="
                    text-sm text-gray-400 hover:text-gray-300
                    transition-colors
                  "
                >
                  Dismiss
                </button>
              )}
            </div>
          )}
        </div>
      </div>
    </div>
  )
}

export default UploadErrorModal
