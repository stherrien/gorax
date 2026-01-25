/**
 * FileUploadZone - Drag-and-drop overlay for workflow file uploads
 *
 * Provides a visual drop zone that appears when users drag files over the canvas.
 * Supports JSON and YAML workflow files with animated feedback.
 */

import { useCallback } from 'react'
import { SUPPORTED_EXTENSIONS } from '../../utils/fileValidation'

interface FileUploadZoneProps {
  /** Whether files are being dragged over the zone */
  isDragging: boolean
  /** Whether to show the zone (controlled externally) */
  isVisible?: boolean
  /** Callback when files are dropped */
  onDrop: (event: React.DragEvent) => void
  /** Callback for drag enter */
  onDragEnter: (event: React.DragEvent) => void
  /** Callback for drag leave */
  onDragLeave: (event: React.DragEvent) => void
  /** Callback for drag over */
  onDragOver: (event: React.DragEvent) => void
  /** Reference to hidden file input for click-to-upload */
  fileInputRef?: React.RefObject<HTMLInputElement | null>
  /** Callback when files are selected via file picker */
  onFileInputChange?: (event: React.ChangeEvent<HTMLInputElement>) => void
  /** Whether to show the click-to-upload button */
  showUploadButton?: boolean
  /** Callback to open file picker */
  onOpenFilePicker?: () => void
}

/**
 * FileUploadZone component
 * Renders a drop zone overlay with visual feedback for file uploads
 */
export function FileUploadZone({
  isDragging,
  isVisible,
  onDrop,
  onDragEnter,
  onDragLeave,
  onDragOver,
  fileInputRef,
  onFileInputChange,
  showUploadButton = false,
  onOpenFilePicker,
}: FileUploadZoneProps) {
  // Show if visible prop is true or if dragging
  const shouldShow = isVisible || isDragging

  const handleDrop = useCallback(
    (event: React.DragEvent) => {
      event.preventDefault()
      event.stopPropagation()
      onDrop(event)
    },
    [onDrop]
  )

  // Don't render if not visible
  if (!shouldShow) {
    return (
      <>
        {/* Hidden file input for programmatic file selection */}
        {fileInputRef && (
          <input
            ref={(el) => {
              (fileInputRef as React.MutableRefObject<HTMLInputElement | null>).current = el
            }}
            type="file"
            accept={SUPPORTED_EXTENSIONS.join(',')}
            onChange={onFileInputChange}
            className="hidden"
            aria-hidden="true"
          />
        )}
      </>
    )
  }

  return (
    <>
      {/* Hidden file input */}
      {fileInputRef && (
        <input
          ref={(el) => {
            (fileInputRef as React.MutableRefObject<HTMLInputElement | null>).current = el
          }}
          type="file"
          accept={SUPPORTED_EXTENSIONS.join(',')}
          onChange={onFileInputChange}
          className="hidden"
          aria-hidden="true"
        />
      )}

      {/* Drop zone overlay */}
      <div
        className={`
          absolute inset-0 z-50
          flex flex-col items-center justify-center
          transition-all duration-200 ease-in-out
          ${isDragging
            ? 'bg-primary-900/90 border-4 border-dashed border-primary-400'
            : 'bg-gray-900/80 border-2 border-dashed border-gray-600'
          }
        `}
        onDrop={handleDrop}
        onDragEnter={onDragEnter}
        onDragLeave={onDragLeave}
        onDragOver={onDragOver}
        role="region"
        aria-label="Drop workflow file here to import"
      >
        {/* Icon */}
        <div
          className={`
            mb-6 transition-transform duration-200
            ${isDragging ? 'scale-110' : 'scale-100'}
          `}
        >
          <svg
            className={`w-20 h-20 ${isDragging ? 'text-primary-300' : 'text-gray-400'}`}
            fill="none"
            stroke="currentColor"
            viewBox="0 0 24 24"
            aria-hidden="true"
          >
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              strokeWidth={1.5}
              d="M7 16a4 4 0 01-.88-7.903A5 5 0 1115.9 6L16 6a5 5 0 011 9.9M15 13l-3-3m0 0l-3 3m3-3v12"
            />
          </svg>
        </div>

        {/* Text content */}
        <div className="text-center px-8">
          <h3
            className={`
              text-xl font-semibold mb-2
              ${isDragging ? 'text-primary-200' : 'text-gray-200'}
            `}
          >
            {isDragging ? 'Drop to import workflow' : 'Import Workflow'}
          </h3>

          <p className={`text-sm mb-4 ${isDragging ? 'text-primary-300' : 'text-gray-400'}`}>
            {isDragging
              ? 'Release to import your workflow file'
              : 'Drag and drop a JSON or YAML workflow file here'
            }
          </p>

          {/* Supported formats */}
          <div className="flex items-center justify-center gap-2 mb-4">
            {SUPPORTED_EXTENSIONS.map((ext) => (
              <span
                key={ext}
                className={`
                  px-2 py-1 rounded text-xs font-mono
                  ${isDragging
                    ? 'bg-primary-700/50 text-primary-200'
                    : 'bg-gray-700/50 text-gray-300'
                  }
                `}
              >
                {ext}
              </span>
            ))}
          </div>

          {/* Upload button */}
          {showUploadButton && !isDragging && onOpenFilePicker && (
            <button
              type="button"
              onClick={onOpenFilePicker}
              className="
                px-4 py-2 mt-2
                bg-primary-600 hover:bg-primary-700
                text-white text-sm font-medium
                rounded-lg transition-colors
                focus:outline-none focus:ring-2 focus:ring-primary-500 focus:ring-offset-2 focus:ring-offset-gray-900
              "
            >
              Or click to browse
            </button>
          )}
        </div>

        {/* Animated border glow when dragging */}
        {isDragging && (
          <div
            className="absolute inset-0 pointer-events-none animate-pulse"
            style={{
              boxShadow: 'inset 0 0 60px rgba(59, 130, 246, 0.3)',
            }}
          />
        )}
      </div>
    </>
  )
}

/**
 * Compact file upload trigger button
 * Can be placed in toolbars or menus
 */
interface FileUploadButtonProps {
  onClick: () => void
  disabled?: boolean
  className?: string
}

export function FileUploadButton({
  onClick,
  disabled = false,
  className = '',
}: FileUploadButtonProps) {
  return (
    <button
      type="button"
      onClick={onClick}
      disabled={disabled}
      className={`
        inline-flex items-center gap-2
        px-3 py-2
        bg-gray-700 hover:bg-gray-600
        text-gray-200 text-sm font-medium
        rounded-lg transition-colors
        disabled:opacity-50 disabled:cursor-not-allowed
        focus:outline-none focus:ring-2 focus:ring-primary-500
        ${className}
      `}
      title="Import workflow from file"
    >
      <svg
        className="w-4 h-4"
        fill="none"
        stroke="currentColor"
        viewBox="0 0 24 24"
        aria-hidden="true"
      >
        <path
          strokeLinecap="round"
          strokeLinejoin="round"
          strokeWidth={2}
          d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-8l-4-4m0 0L8 8m4-4v12"
        />
      </svg>
      <span>Import</span>
    </button>
  )
}

export default FileUploadZone
