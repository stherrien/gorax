import { useState, useRef, useEffect } from 'react'
import { useExecutionLogs } from '../../hooks/useMonitoring'
import type { ExecutionLog, LogLevel } from '../../types/management'

interface LogViewerProps {
  executionId: string
  autoScroll?: boolean
  maxHeight?: string
}

const levelColors: Record<LogLevel, string> = {
  debug: 'text-gray-400',
  info: 'text-blue-400',
  warn: 'text-yellow-400',
  error: 'text-red-400',
}

const levelBgColors: Record<LogLevel, string> = {
  debug: 'bg-gray-500/10',
  info: 'bg-blue-500/10',
  warn: 'bg-yellow-500/10',
  error: 'bg-red-500/10',
}

export function LogViewer({ executionId, autoScroll = true, maxHeight = '500px' }: LogViewerProps) {
  const [levelFilter, setLevelFilter] = useState<LogLevel | 'all'>('all')
  const [searchQuery, setSearchQuery] = useState('')
  const [isAutoScrollEnabled, setIsAutoScrollEnabled] = useState(autoScroll)
  const logContainerRef = useRef<HTMLDivElement>(null)

  const { logs, loading, error, refetch } = useExecutionLogs(executionId, {
    level: levelFilter !== 'all' ? levelFilter : undefined,
    search: searchQuery || undefined,
    limit: 500,
  })

  // Auto-scroll to bottom when new logs arrive
  useEffect(() => {
    if (isAutoScrollEnabled && logContainerRef.current) {
      logContainerRef.current.scrollTop = logContainerRef.current.scrollHeight
    }
  }, [logs, isAutoScrollEnabled])

  const handleExport = (format: 'txt' | 'json') => {
    const content =
      format === 'json'
        ? JSON.stringify(logs, null, 2)
        : logs.map((log) => `[${log.timestamp}] [${log.level.toUpperCase()}] ${log.message}`).join('\n')

    const blob = new Blob([content], { type: format === 'json' ? 'application/json' : 'text/plain' })
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = `execution-${executionId}-logs.${format}`
    document.body.appendChild(a)
    a.click()
    document.body.removeChild(a)
    URL.revokeObjectURL(url)
  }

  return (
    <div className="bg-gray-800 rounded-lg">
      {/* Toolbar */}
      <div className="px-4 py-3 border-b border-gray-700 flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
        <div className="flex items-center gap-3">
          <h3 className="text-lg font-semibold text-white">Logs</h3>
          <span className="text-sm text-gray-400">({logs.length} entries)</span>
        </div>

        <div className="flex flex-wrap items-center gap-2">
          {/* Search */}
          <div className="relative">
            <input
              type="text"
              placeholder="Search logs..."
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              className="w-48 px-3 py-1.5 pl-8 bg-gray-700 border border-gray-600 rounded-lg text-white text-sm placeholder-gray-400 focus:outline-none focus:ring-2 focus:ring-primary-500"
            />
            <svg
              className="absolute left-2.5 top-1/2 -translate-y-1/2 w-3.5 h-3.5 text-gray-400"
              fill="none"
              viewBox="0 0 24 24"
              stroke="currentColor"
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={2}
                d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z"
              />
            </svg>
          </div>

          {/* Level filter */}
          <select
            value={levelFilter}
            onChange={(e) => setLevelFilter(e.target.value as LogLevel | 'all')}
            className="px-3 py-1.5 bg-gray-700 border border-gray-600 rounded-lg text-white text-sm focus:outline-none focus:ring-2 focus:ring-primary-500"
          >
            <option value="all">All Levels</option>
            <option value="debug">Debug</option>
            <option value="info">Info</option>
            <option value="warn">Warning</option>
            <option value="error">Error</option>
          </select>

          {/* Auto-scroll toggle */}
          <button
            onClick={() => setIsAutoScrollEnabled(!isAutoScrollEnabled)}
            className={`p-1.5 rounded-lg transition-colors ${
              isAutoScrollEnabled
                ? 'bg-primary-600 text-white'
                : 'bg-gray-700 text-gray-400 hover:text-white'
            }`}
            title={isAutoScrollEnabled ? 'Auto-scroll on' : 'Auto-scroll off'}
          >
            <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={2}
                d="M19 14l-7 7m0 0l-7-7m7 7V3"
              />
            </svg>
          </button>

          {/* Refresh */}
          <button
            onClick={() => refetch()}
            disabled={loading}
            className="p-1.5 bg-gray-700 text-gray-400 hover:text-white rounded-lg transition-colors disabled:opacity-50"
            title="Refresh logs"
          >
            <svg
              className={`w-4 h-4 ${loading ? 'animate-spin' : ''}`}
              fill="none"
              viewBox="0 0 24 24"
              stroke="currentColor"
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={2}
                d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"
              />
            </svg>
          </button>

          {/* Export dropdown */}
          <div className="relative group">
            <button className="p-1.5 bg-gray-700 text-gray-400 hover:text-white rounded-lg transition-colors">
              <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  strokeWidth={2}
                  d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-4l-4 4m0 0l-4-4m4 4V4"
                />
              </svg>
            </button>
            <div className="absolute right-0 mt-1 w-32 bg-gray-700 rounded-lg shadow-lg opacity-0 invisible group-hover:opacity-100 group-hover:visible transition-all z-10">
              <button
                onClick={() => handleExport('txt')}
                className="w-full px-3 py-2 text-left text-sm text-gray-300 hover:text-white hover:bg-gray-600 rounded-t-lg"
              >
                Export as TXT
              </button>
              <button
                onClick={() => handleExport('json')}
                className="w-full px-3 py-2 text-left text-sm text-gray-300 hover:text-white hover:bg-gray-600 rounded-b-lg"
              >
                Export as JSON
              </button>
            </div>
          </div>
        </div>
      </div>

      {/* Error state */}
      {error && (
        <div className="p-4 text-red-400 text-center">Failed to load logs: {error.message}</div>
      )}

      {/* Log container */}
      <div
        ref={logContainerRef}
        className="overflow-auto font-mono text-sm"
        style={{ maxHeight }}
      >
        {loading && logs.length === 0 ? (
          <div className="p-4 space-y-2">
            {[1, 2, 3, 4, 5].map((i) => (
              <div key={i} className="h-6 bg-gray-700 rounded animate-pulse" />
            ))}
          </div>
        ) : logs.length === 0 ? (
          <div className="p-8 text-center text-gray-400">No logs found</div>
        ) : (
          <table className="w-full">
            <tbody>
              {logs.map((log: ExecutionLog) => (
                <LogEntry key={log.id} log={log} />
              ))}
            </tbody>
          </table>
        )}
      </div>
    </div>
  )
}

function LogEntry({ log }: { log: ExecutionLog }) {
  const [isExpanded, setIsExpanded] = useState(false)

  const hasData = log.data && Object.keys(log.data).length > 0

  return (
    <>
      <tr
        className={`border-b border-gray-700/50 hover:bg-gray-700/30 cursor-pointer ${levelBgColors[log.level]}`}
        onClick={() => hasData && setIsExpanded(!isExpanded)}
      >
        <td className="px-3 py-1.5 text-gray-500 whitespace-nowrap w-32">
          {new Date(log.timestamp).toLocaleTimeString('en-US', {
            hour: '2-digit',
            minute: '2-digit',
            second: '2-digit',
          })}
        </td>
        <td className={`px-3 py-1.5 w-16 ${levelColors[log.level]} font-medium uppercase text-xs`}>
          {log.level}
        </td>
        {log.nodeName && (
          <td className="px-3 py-1.5 text-gray-400 w-32 truncate">{log.nodeName}</td>
        )}
        <td className="px-3 py-1.5 text-gray-200 break-all">
          {log.message}
          {hasData && (
            <span className="ml-2 text-gray-500">
              {isExpanded ? '▼' : '▶'}
            </span>
          )}
        </td>
      </tr>
      {isExpanded && hasData && (
        <tr className="bg-gray-900/50">
          <td colSpan={4} className="px-3 py-2">
            <pre className="text-xs text-gray-400 overflow-x-auto">
              {JSON.stringify(log.data, null, 2)}
            </pre>
          </td>
        </tr>
      )}
    </>
  )
}
