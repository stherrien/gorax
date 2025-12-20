import { useState, useEffect, useMemo } from 'react'
import { Link } from 'react-router-dom'
import { webhookAPI, type WebhookEvent } from '../../api/webhooks'
import { ReplayModal } from './ReplayModal'
import { convertToCSV, downloadCSV, formatDateForCSV, truncateForCSV } from '../../utils/csvExport'

interface WebhookEventHistoryProps {
  webhookId: string
}

type SortOrder = 'asc' | 'desc'
type EventStatus = 'all' | 'received' | 'processed' | 'filtered' | 'failed'

const MAX_REPLAY_COUNT = 5
const MAX_BATCH_SIZE = 10

export function WebhookEventHistory({ webhookId }: WebhookEventHistoryProps) {
  const [events, setEvents] = useState<WebhookEvent[]>([])
  const [total, setTotal] = useState(0)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<Error | null>(null)
  const [page, setPage] = useState(1)
  const [sortOrder, setSortOrder] = useState<SortOrder>('desc')
  const [statusFilter, setStatusFilter] = useState<EventStatus>('all')
  const [searchTerm, setSearchTerm] = useState('')
  const [selectedEvent, setSelectedEvent] = useState<WebhookEvent | null>(null)
  const [replayEvent, setReplayEvent] = useState<WebhookEvent | null>(null)
  const [selectedEvents, setSelectedEvents] = useState<Set<string>>(new Set())
  const [showBatchConfirm, setShowBatchConfirm] = useState(false)
  const [notification, setNotification] = useState<string | null>(null)
  const [notificationType, setNotificationType] = useState<'success' | 'error' | 'warning'>('success')

  const limit = 20

  useEffect(() => {
    fetchEvents()
  }, [webhookId, page])

  const fetchEvents = async () => {
    try {
      setLoading(true)
      setError(null)
      const response = await webhookAPI.getEvents(webhookId, { page, limit })
      setEvents(response.events)
      setTotal(response.total)
    } catch (err) {
      setError(err as Error)
    } finally {
      setLoading(false)
    }
  }

  const filteredAndSortedEvents = useMemo(() => {
    let result = [...events]

    // Filter by status
    if (statusFilter !== 'all') {
      result = result.filter((event) => event.status === statusFilter)
    }

    // Filter by search term in payload
    if (searchTerm) {
      result = result.filter((event) => {
        const payloadStr = JSON.stringify(event.requestBody).toLowerCase()
        return payloadStr.includes(searchTerm.toLowerCase())
      })
    }

    // Sort by time
    result.sort((a, b) => {
      const timeA = new Date(a.createdAt).getTime()
      const timeB = new Date(b.createdAt).getTime()
      return sortOrder === 'desc' ? timeB - timeA : timeA - timeB
    })

    return result
  }, [events, sortOrder, statusFilter, searchTerm])

  const handleSort = () => {
    setSortOrder(sortOrder === 'desc' ? 'asc' : 'desc')
  }

  const handleEventClick = (event: WebhookEvent) => {
    setSelectedEvent(event)
  }

  const closeModal = () => {
    setSelectedEvent(null)
  }

  const handleExport = () => {
    const eventsToExport = filteredAndSortedEvents.map((event) => ({
      eventId: event.id,
      timestamp: formatDateForCSV(event.createdAt),
      method: event.requestMethod,
      status: event.status,
      responseCode: event.responseStatus ?? 'N/A',
      processingTime: event.processingTimeMs ?? 'N/A',
      executionId: event.executionId ?? 'N/A',
      errorMessage: event.errorMessage ?? '',
      replayCount: event.replayCount,
      sourceIp: event.metadata?.sourceIp ?? 'N/A',
      userAgent: truncateForCSV(event.metadata?.userAgent ?? 'N/A', 100),
      contentType: event.metadata?.contentType ?? 'N/A',
      contentLength: event.metadata?.contentLength ?? 'N/A',
      receivedAt: event.metadata?.receivedAt ? formatDateForCSV(event.metadata.receivedAt) : 'N/A',
      headers: truncateForCSV(JSON.stringify(event.requestHeaders), 200),
      payload: truncateForCSV(JSON.stringify(event.requestBody), 500),
    }))

    const csv = convertToCSV(
      eventsToExport,
      [
        'eventId',
        'timestamp',
        'method',
        'status',
        'responseCode',
        'processingTime',
        'executionId',
        'errorMessage',
        'replayCount',
        'sourceIp',
        'userAgent',
        'contentType',
        'contentLength',
        'receivedAt',
        'headers',
        'payload',
      ],
      [
        'Event ID',
        'Timestamp',
        'Method',
        'Status',
        'Response Code',
        'Processing Time (ms)',
        'Execution ID',
        'Error Message',
        'Replay Count',
        'Source IP',
        'User Agent',
        'Content Type',
        'Content Length',
        'Received At',
        'Headers',
        'Payload',
      ]
    )

    const filename = `webhook-events-${webhookId}-${new Date().toISOString().split('T')[0]}.csv`
    downloadCSV(csv, filename)
  }

  const handleCheckboxChange = (eventId: string, checked: boolean) => {
    const newSelection = new Set(selectedEvents)

    if (checked) {
      if (newSelection.size >= MAX_BATCH_SIZE) {
        showNotification(`Maximum ${MAX_BATCH_SIZE} events can be selected for batch replay`, 'warning')
        return
      }
      newSelection.add(eventId)
    } else {
      newSelection.delete(eventId)
    }

    setSelectedEvents(newSelection)
  }

  const handleBatchReplay = () => {
    setShowBatchConfirm(true)
  }

  const confirmBatchReplay = async () => {
    try {
      const eventIds = Array.from(selectedEvents)
      const result = await webhookAPI.batchReplayEvents(webhookId, eventIds)

      const successCount = Object.values(result.results).filter(r => r.success).length
      const failureCount = Object.values(result.results).filter(r => !r.success).length

      if (failureCount === 0) {
        showNotification(`Successfully replayed ${successCount} event${successCount !== 1 ? 's' : ''}`, 'success')
      } else {
        showNotification(`${successCount} succeeded, ${failureCount} failed`, 'warning')
      }

      setSelectedEvents(new Set())
      setShowBatchConfirm(false)
      fetchEvents()
    } catch (err) {
      showNotification('Failed to replay events', 'error')
    }
  }

  const handleReplaySuccess = (executionId: string) => {
    showNotification(`Event replayed successfully. Execution ID: ${executionId}`, 'success')
    setReplayEvent(null)
    fetchEvents()
  }

  const showNotification = (message: string, type: 'success' | 'error' | 'warning') => {
    setNotification(message)
    setNotificationType(type)
    setTimeout(() => setNotification(null), 5000)
  }

  // Loading state
  if (loading && events.length === 0) {
    return (
      <div className="h-64 flex items-center justify-center">
        <div className="text-white text-lg">Loading events...</div>
      </div>
    )
  }

  // Error state
  if (error) {
    return (
      <div className="h-64 flex items-center justify-center">
        <div className="text-center">
          <div className="text-red-400 text-lg mb-4">Failed to fetch events</div>
          <div className="text-gray-400 text-sm">{error.message}</div>
        </div>
      </div>
    )
  }

  // Empty state
  if (events.length === 0) {
    return (
      <div className="h-64 flex items-center justify-center bg-gray-800 rounded-lg">
        <div className="text-center">
          <div className="text-gray-400 text-lg mb-4">No events found</div>
          <div className="text-gray-500 text-sm">Webhook events will appear here once received</div>
        </div>
      </div>
    )
  }

  const totalPages = Math.ceil(total / limit)
  const canGoPrevious = page > 1
  const canGoNext = page < totalPages

  return (
    <div>
      {/* Notification Banner */}
      {notification && (
        <div
          className={`mb-4 p-3 rounded-lg ${
            notificationType === 'success'
              ? 'bg-green-900/30 border border-green-600/50 text-green-300'
              : notificationType === 'error'
              ? 'bg-red-900/30 border border-red-600/50 text-red-300'
              : 'bg-yellow-900/30 border border-yellow-600/50 text-yellow-300'
          }`}
        >
          {notification}
        </div>
      )}

      {/* Filters and Actions */}
      <div className="mb-4 flex items-center justify-between gap-4">
        <div className="flex items-center gap-4">
          <div>
            <label htmlFor="status-filter" className="sr-only">
              Filter by status
            </label>
            <select
              id="status-filter"
              aria-label="Filter by status"
              value={statusFilter}
              onChange={(e) => setStatusFilter(e.target.value as EventStatus)}
              className="px-3 py-2 bg-gray-800 text-white rounded-lg text-sm border border-gray-700 focus:outline-none focus:border-primary-500"
            >
              <option value="all">All Status</option>
              <option value="received">Received</option>
              <option value="processed">Processed</option>
              <option value="filtered">Filtered</option>
              <option value="failed">Failed</option>
            </select>
          </div>

          <div>
            <label htmlFor="search-payload" className="sr-only">
              Search payload
            </label>
            <input
              id="search-payload"
              type="text"
              placeholder="Search payload..."
              value={searchTerm}
              onChange={(e) => setSearchTerm(e.target.value)}
              className="px-3 py-2 bg-gray-800 text-white rounded-lg text-sm border border-gray-700 focus:outline-none focus:border-primary-500"
            />
          </div>
        </div>

        <div className="flex items-center gap-2">
          {selectedEvents.size > 0 && (
            <button
              onClick={handleBatchReplay}
              className="px-4 py-2 bg-primary-600 text-white rounded-lg text-sm font-medium hover:bg-primary-700 transition-colors"
            >
              Replay Selected ({selectedEvents.size})
            </button>
          )}
          <button
            onClick={handleExport}
            className="px-4 py-2 bg-gray-700 text-white rounded-lg text-sm font-medium hover:bg-gray-600 transition-colors"
          >
            Export CSV
          </button>
        </div>
      </div>

      {/* Events Table */}
      <div className="bg-gray-800 rounded-lg overflow-hidden">
        <table className="w-full" role="table">
          <thead>
            <tr className="border-b border-gray-700">
              <th className="text-left px-6 py-4 text-sm font-medium text-gray-400" role="columnheader">
                <input
                  type="checkbox"
                  checked={
                    filteredAndSortedEvents.length > 0 &&
                    filteredAndSortedEvents.every(
                      (e) => selectedEvents.has(e.id) || e.replayCount >= MAX_REPLAY_COUNT
                    )
                  }
                  onChange={(e) => {
                    if (e.target.checked) {
                      const newSelection = new Set<string>()
                      let count = 0
                      for (const event of filteredAndSortedEvents) {
                        if (event.replayCount < MAX_REPLAY_COUNT && count < MAX_BATCH_SIZE) {
                          newSelection.add(event.id)
                          count++
                        }
                      }
                      setSelectedEvents(newSelection)
                    } else {
                      setSelectedEvents(new Set())
                    }
                  }}
                  className="w-4 h-4 text-primary-600 bg-gray-700 border-gray-600 rounded focus:ring-primary-500 focus:ring-2"
                />
              </th>
              <th className="text-left px-6 py-4 text-sm font-medium text-gray-400" role="columnheader">
                Event ID
              </th>
              <th className="text-left px-6 py-4 text-sm font-medium text-gray-400" role="columnheader">
                <button onClick={handleSort} className="hover:text-white transition-colors">
                  Time {sortOrder === 'desc' ? '↓' : '↑'}
                </button>
              </th>
              <th className="text-left px-6 py-4 text-sm font-medium text-gray-400" role="columnheader">
                Method
              </th>
              <th className="text-left px-6 py-4 text-sm font-medium text-gray-400" role="columnheader">
                Status
              </th>
              <th className="text-left px-6 py-4 text-sm font-medium text-gray-400" role="columnheader">
                Response Code
              </th>
              <th className="text-left px-6 py-4 text-sm font-medium text-gray-400" role="columnheader">
                Processing Time
              </th>
              <th className="text-left px-6 py-4 text-sm font-medium text-gray-400" role="columnheader">
                Actions
              </th>
            </tr>
          </thead>
          <tbody>
            {filteredAndSortedEvents.map((event) => (
              <tr
                key={event.id}
                role="row"
                className="border-b border-gray-700 hover:bg-gray-700/50"
              >
                <td className="px-6 py-4" onClick={(e) => e.stopPropagation()}>
                  <input
                    type="checkbox"
                    checked={selectedEvents.has(event.id)}
                    disabled={event.replayCount >= MAX_REPLAY_COUNT}
                    onChange={(e) => handleCheckboxChange(event.id, e.target.checked)}
                    className="w-4 h-4 text-primary-600 bg-gray-700 border-gray-600 rounded focus:ring-primary-500 focus:ring-2 disabled:opacity-50"
                  />
                </td>
                <td
                  className="px-6 py-4 text-gray-300 text-sm cursor-pointer"
                  onClick={() => handleEventClick(event)}
                >
                  {event.id}
                  {event.replayCount > 0 && (
                    <span className="ml-2 text-xs px-2 py-0.5 bg-blue-500/20 text-blue-400 rounded">
                      {event.replayCount}×
                    </span>
                  )}
                </td>
                <td
                  className="px-6 py-4 text-gray-300 text-sm cursor-pointer"
                  onClick={() => handleEventClick(event)}
                >
                  {formatTime(event.createdAt)}
                </td>
                <td
                  className="px-6 py-4 text-gray-300 text-sm cursor-pointer"
                  onClick={() => handleEventClick(event)}
                >
                  {event.requestMethod}
                </td>
                <td className="px-6 py-4 cursor-pointer" onClick={() => handleEventClick(event)}>
                  <StatusBadge status={event.status} />
                </td>
                <td
                  className="px-6 py-4 text-gray-300 text-sm cursor-pointer"
                  onClick={() => handleEventClick(event)}
                >
                  {event.responseStatus ?? 'N/A'}
                </td>
                <td
                  className="px-6 py-4 text-gray-300 text-sm cursor-pointer"
                  onClick={() => handleEventClick(event)}
                >
                  {event.processingTimeMs ? `${event.processingTimeMs} ms` : 'N/A'}
                </td>
                <td className="px-6 py-4" onClick={(e) => e.stopPropagation()}>
                  <button
                    onClick={() => setReplayEvent(event)}
                    disabled={event.replayCount >= MAX_REPLAY_COUNT}
                    title={
                      event.replayCount >= MAX_REPLAY_COUNT
                        ? 'Maximum replay limit reached'
                        : 'Replay this event'
                    }
                    className="px-3 py-1 text-xs bg-primary-600 text-white rounded hover:bg-primary-700 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
                  >
                    Replay
                  </button>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      {/* Pagination */}
      {totalPages > 1 && (
        <nav
          className="mt-4 flex items-center justify-between"
          role="navigation"
          aria-label="Pagination"
        >
          <div className="text-sm text-gray-400">
            Page {page} of {totalPages}
          </div>
          <div className="flex gap-2">
            <button
              onClick={() => setPage(page - 1)}
              disabled={!canGoPrevious}
              className="px-4 py-2 bg-gray-700 text-white rounded-lg text-sm font-medium hover:bg-gray-600 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
            >
              Previous
            </button>
            <button
              onClick={() => setPage(page + 1)}
              disabled={!canGoNext}
              className="px-4 py-2 bg-gray-700 text-white rounded-lg text-sm font-medium hover:bg-gray-600 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
            >
              Next
            </button>
          </div>
        </nav>
      )}

      {/* Event Detail Modal */}
      {selectedEvent && (
        <EventDetailModal event={selectedEvent} onClose={closeModal} />
      )}

      {/* Replay Modal */}
      {replayEvent && (
        <ReplayModal
          event={replayEvent}
          onClose={() => setReplayEvent(null)}
          onSuccess={handleReplaySuccess}
        />
      )}

      {/* Batch Replay Confirmation Modal */}
      {showBatchConfirm && (
        <div
          className="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4"
          onClick={(e) => {
            if (e.target === e.currentTarget) {
              setShowBatchConfirm(false)
            }
          }}
        >
          <div
            role="dialog"
            aria-modal="true"
            aria-labelledby="batch-confirm-title"
            className="bg-gray-800 rounded-lg p-6 max-w-md w-full"
          >
            <h2 id="batch-confirm-title" className="text-xl font-bold text-white mb-4">
              Confirm Batch Replay
            </h2>
            <p className="text-gray-300 mb-6">
              Are you sure you want to replay {selectedEvents.size} event
              {selectedEvents.size !== 1 ? 's' : ''}?
            </p>
            <div className="flex justify-end gap-3">
              <button
                onClick={() => setShowBatchConfirm(false)}
                className="px-4 py-2 bg-gray-700 text-white rounded-lg text-sm font-medium hover:bg-gray-600 transition-colors"
              >
                Cancel
              </button>
              <button
                onClick={confirmBatchReplay}
                className="px-4 py-2 bg-primary-600 text-white rounded-lg text-sm font-medium hover:bg-primary-700 transition-colors"
              >
                Confirm
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}

function StatusBadge({ status }: { status: WebhookEvent['status'] }) {
  const colors = {
    received: 'bg-blue-500/20 text-blue-400',
    processed: 'bg-green-500/20 text-green-400',
    filtered: 'bg-yellow-500/20 text-yellow-400',
    failed: 'bg-red-500/20 text-red-400',
  }

  return (
    <span
      data-testid={`status-badge-${status}`}
      className={`inline-flex px-2 py-1 text-xs font-medium rounded-full ${colors[status]}`}
    >
      {status}
    </span>
  )
}

function formatTime(timestamp: string): string {
  const date = new Date(timestamp)
  return date.toLocaleTimeString('en-US', {
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit',
    hour12: false,
  })
}

interface EventDetailModalProps {
  event: WebhookEvent
  onClose: () => void
}

function EventDetailModal({ event, onClose }: EventDetailModalProps) {
  const handleBackdropClick = (e: React.MouseEvent) => {
    if (e.target === e.currentTarget) {
      onClose()
    }
  }

  return (
    <div
      data-testid="modal-backdrop"
      onClick={handleBackdropClick}
      className="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4"
    >
      <div
        role="dialog"
        aria-modal="true"
        aria-labelledby="modal-title"
        className="bg-gray-800 rounded-lg p-6 max-w-3xl w-full max-h-[90vh] overflow-y-auto"
      >
        <div className="flex justify-between items-start mb-6">
          <h2 id="modal-title" className="text-xl font-bold text-white">
            Event Details
          </h2>
          <button
            onClick={onClose}
            className="text-gray-400 hover:text-white transition-colors"
            aria-label="Close"
          >
            <svg className="w-6 h-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        </div>

        <div className="space-y-6">
          {/* Basic Info */}
          <div className="grid grid-cols-2 gap-4">
            <div>
              <h3 className="text-sm font-medium text-gray-400 mb-1">Event ID</h3>
              <p className="text-white">{event.id}</p>
            </div>
            <div>
              <h3 className="text-sm font-medium text-gray-400 mb-1">Status</h3>
              <StatusBadge status={event.status} />
            </div>
            <div>
              <h3 className="text-sm font-medium text-gray-400 mb-1">Response Status</h3>
              <p className="text-white">{event.responseStatus ?? 'N/A'}</p>
            </div>
            <div>
              <h3 className="text-sm font-medium text-gray-400 mb-1">Processing Time</h3>
              <p className="text-white">
                {event.processingTimeMs ? `${event.processingTimeMs} ms` : 'N/A'}
              </p>
            </div>
          </div>

          {/* Error Message */}
          {event.errorMessage && (
            <div>
              <h3 className="text-sm font-medium text-gray-400 mb-2">Error</h3>
              <div className="p-3 bg-red-500/10 border border-red-500/20 rounded-lg">
                <p className="text-red-400 text-sm">{event.errorMessage}</p>
              </div>
            </div>
          )}

          {/* Execution Link */}
          {event.executionId && (
            <div>
              <h3 className="text-sm font-medium text-gray-400 mb-2">Triggered Execution</h3>
              <Link
                to={`/executions/${event.executionId}`}
                className="text-primary-400 hover:text-primary-300 text-sm inline-flex items-center gap-1"
              >
                View Execution
                <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    strokeWidth={2}
                    d="M10 6H6a2 2 0 00-2 2v10a2 2 0 002 2h10a2 2 0 002-2v-4M14 4h6m0 0v6m0-6L10 14"
                  />
                </svg>
              </Link>
            </div>
          )}

          {/* Request Metadata */}
          {event.metadata && (
            <div>
              <h3 className="text-sm font-medium text-gray-400 mb-2">Request Details</h3>
              <div className="p-4 bg-gray-900 rounded-lg">
                <div className="grid grid-cols-2 gap-4">
                  <div>
                    <p className="text-xs text-gray-500 mb-1">Source IP</p>
                    <p className="text-sm text-gray-300">{event.metadata.sourceIp}</p>
                  </div>
                  <div>
                    <p className="text-xs text-gray-500 mb-1">Received At</p>
                    <p className="text-sm text-gray-300">
                      {new Date(event.metadata.receivedAt).toLocaleString()}
                    </p>
                  </div>
                  <div>
                    <p className="text-xs text-gray-500 mb-1">User Agent</p>
                    <p className="text-sm text-gray-300 break-all">
                      {event.metadata.userAgent || 'N/A'}
                    </p>
                  </div>
                  <div>
                    <p className="text-xs text-gray-500 mb-1">Content Type</p>
                    <p className="text-sm text-gray-300">{event.metadata.contentType || 'N/A'}</p>
                  </div>
                  <div>
                    <p className="text-xs text-gray-500 mb-1">Content Length</p>
                    <p className="text-sm text-gray-300">{event.metadata.contentLength} bytes</p>
                  </div>
                </div>
              </div>
            </div>
          )}

          {/* Request Headers */}
          <div>
            <h3 className="text-sm font-medium text-gray-400 mb-2">Request Headers</h3>
            <div className="p-4 bg-gray-900 rounded-lg overflow-x-auto">
              <pre className="text-sm text-gray-300">
                {JSON.stringify(event.requestHeaders, null, 2)}
              </pre>
            </div>
          </div>

          {/* Request Body */}
          <div>
            <h3 className="text-sm font-medium text-gray-400 mb-2">Request Body</h3>
            <div className="p-4 bg-gray-900 rounded-lg overflow-x-auto">
              <pre className="text-sm text-gray-300">
                {JSON.stringify(event.requestBody, null, 2)}
              </pre>
            </div>
          </div>
        </div>

        <div className="mt-6 flex justify-end">
          <button
            onClick={onClose}
            aria-label="Close modal"
            className="px-4 py-2 bg-gray-700 text-white rounded-lg text-sm font-medium hover:bg-gray-600 transition-colors"
          >
            Close
          </button>
        </div>
      </div>
    </div>
  )
}
