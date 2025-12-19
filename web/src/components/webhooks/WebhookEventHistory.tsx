import { useState, useEffect, useMemo } from 'react'
import { Link } from 'react-router-dom'
import { webhookAPI, type WebhookEvent } from '../../api/webhooks'

interface WebhookEventHistoryProps {
  webhookId: string
}

type SortOrder = 'asc' | 'desc'
type EventStatus = 'all' | 'received' | 'processed' | 'filtered' | 'failed'

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
    // Mock implementation for now
    console.log('Export CSV clicked')
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

        <button
          onClick={handleExport}
          className="px-4 py-2 bg-gray-700 text-white rounded-lg text-sm font-medium hover:bg-gray-600 transition-colors"
        >
          Export CSV
        </button>
      </div>

      {/* Events Table */}
      <div className="bg-gray-800 rounded-lg overflow-hidden">
        <table className="w-full" role="table">
          <thead>
            <tr className="border-b border-gray-700">
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
            </tr>
          </thead>
          <tbody>
            {filteredAndSortedEvents.map((event) => (
              <tr
                key={event.id}
                role="row"
                onClick={() => handleEventClick(event)}
                className="border-b border-gray-700 hover:bg-gray-700/50 cursor-pointer"
              >
                <td className="px-6 py-4 text-gray-300 text-sm">{event.id}</td>
                <td className="px-6 py-4 text-gray-300 text-sm">
                  {formatTime(event.createdAt)}
                </td>
                <td className="px-6 py-4 text-gray-300 text-sm">{event.requestMethod}</td>
                <td className="px-6 py-4">
                  <StatusBadge status={event.status} />
                </td>
                <td className="px-6 py-4 text-gray-300 text-sm">
                  {event.responseStatus ?? 'N/A'}
                </td>
                <td className="px-6 py-4 text-gray-300 text-sm">
                  {event.processingTimeMs ? `${event.processingTimeMs} ms` : 'N/A'}
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
