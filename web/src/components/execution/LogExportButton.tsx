import { useState } from 'react'
import { workflowAPI } from '../../api/workflows'

interface LogExportButtonProps {
  executionId: string
  className?: string
}

type ExportFormat = 'txt' | 'json' | 'csv'

/**
 * LogExportButton - Dropdown button for exporting execution logs
 *
 * Features:
 * - Export in TXT, JSON, or CSV format
 * - Loading state during export
 * - Error handling with notification
 * - Automatic file download
 */
export function LogExportButton({ executionId, className = '' }: LogExportButtonProps) {
  const [isOpen, setIsOpen] = useState(false)
  const [isLoading, setIsLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const handleExport = async (format: ExportFormat) => {
    try {
      setIsLoading(true)
      setError(null)
      setIsOpen(false)

      await workflowAPI.exportLogs(executionId, format)
    } catch (err) {
      const message = err instanceof Error ? err.message : 'Failed to export logs'
      setError(message)
      setTimeout(() => setError(null), 5000)
    } finally {
      setIsLoading(false)
    }
  }

  return (
    <div className={`log-export-button-container ${className}`}>
      <div className="relative inline-block">
        <button
          onClick={() => setIsOpen(!isOpen)}
          disabled={isLoading}
          className="export-button"
          aria-label="Export logs"
          aria-expanded={isOpen}
          aria-haspopup="true"
        >
          {isLoading ? (
            <>
              <svg className="animate-spin h-4 w-4" viewBox="0 0 24 24">
                <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" fill="none" />
                <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" />
              </svg>
              <span>Exporting...</span>
            </>
          ) : (
            <>
              <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-4l-4 4m0 0l-4-4m4 4V4" />
              </svg>
              <span>Export</span>
              <svg className="w-4 h-4 ml-1" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 9l-7 7-7-7" />
              </svg>
            </>
          )}
        </button>

        {isOpen && !isLoading && (
          <div className="export-dropdown" role="menu">
            <button
              onClick={() => handleExport('txt')}
              className="export-option"
              role="menuitem"
            >
              <span className="option-label">Text (.txt)</span>
              <span className="option-description">Human-readable format</span>
            </button>
            <button
              onClick={() => handleExport('json')}
              className="export-option"
              role="menuitem"
            >
              <span className="option-label">JSON (.json)</span>
              <span className="option-description">Structured data format</span>
            </button>
            <button
              onClick={() => handleExport('csv')}
              className="export-option"
              role="menuitem"
            >
              <span className="option-label">CSV (.csv)</span>
              <span className="option-description">Spreadsheet format</span>
            </button>
          </div>
        )}
      </div>

      {error && (
        <div className="export-error" role="alert">
          <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
          </svg>
          <span>{error}</span>
        </div>
      )}

      <style>{`
        .log-export-button-container {
          position: relative;
        }

        .export-button {
          display: flex;
          align-items: center;
          gap: 0.5rem;
          padding: 0.5rem 1rem;
          background-color: #3b82f6;
          color: white;
          border: none;
          border-radius: 0.375rem;
          cursor: pointer;
          font-size: 0.875rem;
          font-weight: 500;
          transition: background-color 0.2s;
        }

        .export-button:hover:not(:disabled) {
          background-color: #2563eb;
        }

        .export-button:disabled {
          opacity: 0.6;
          cursor: not-allowed;
        }

        .export-dropdown {
          position: absolute;
          top: 100%;
          right: 0;
          margin-top: 0.5rem;
          background-color: white;
          border: 1px solid #e5e7eb;
          border-radius: 0.375rem;
          box-shadow: 0 10px 15px -3px rgba(0, 0, 0, 0.1);
          min-width: 200px;
          z-index: 50;
        }

        .export-option {
          display: flex;
          flex-direction: column;
          align-items: flex-start;
          width: 100%;
          padding: 0.75rem 1rem;
          border: none;
          background: none;
          cursor: pointer;
          text-align: left;
          transition: background-color 0.2s;
        }

        .export-option:hover {
          background-color: #f3f4f6;
        }

        .export-option:not(:last-child) {
          border-bottom: 1px solid #e5e7eb;
        }

        .option-label {
          font-weight: 500;
          color: #111827;
          font-size: 0.875rem;
        }

        .option-description {
          font-size: 0.75rem;
          color: #6b7280;
          margin-top: 0.25rem;
        }

        .export-error {
          display: flex;
          align-items: center;
          gap: 0.5rem;
          margin-top: 0.5rem;
          padding: 0.75rem 1rem;
          background-color: #fef2f2;
          border: 1px solid #fecaca;
          border-radius: 0.375rem;
          color: #991b1b;
          font-size: 0.875rem;
        }

        .animate-spin {
          animation: spin 1s linear infinite;
        }

        @keyframes spin {
          from {
            transform: rotate(0deg);
          }
          to {
            transform: rotate(360deg);
          }
        }
      `}</style>
    </div>
  )
}
