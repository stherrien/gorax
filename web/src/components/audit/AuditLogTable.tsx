import { useState } from 'react'
import { AuditEvent } from '../../types/audit'
import { SeverityBadge } from './SeverityBadge'
import { StatusIcon } from './StatusIcon'
import { CategoryIcon } from './CategoryIcon'
import { AuditEventDetailModal } from './AuditEventDetailModal'
import {
  ChevronUpIcon,
  ChevronDownIcon,
  ChevronLeftIcon,
  ChevronRightIcon,
} from '@heroicons/react/24/outline'

interface AuditLogTableProps {
  events: AuditEvent[]
  total: number
  currentPage: number
  pageSize: number
  onPageChange: (page: number) => void
  onSort?: (field: string, direction: 'ASC' | 'DESC') => void
  sortBy?: string
  sortDirection?: 'ASC' | 'DESC'
  isLoading?: boolean
}

export function AuditLogTable({
  events,
  total,
  currentPage,
  pageSize,
  onPageChange,
  onSort,
  sortBy = 'created_at',
  sortDirection = 'DESC',
  isLoading = false,
}: AuditLogTableProps) {
  const [selectedEvent, setSelectedEvent] = useState<AuditEvent | null>(null)
  const [isModalOpen, setIsModalOpen] = useState(false)

  const totalPages = Math.ceil(total / pageSize)
  const startIndex = (currentPage - 1) * pageSize + 1
  const endIndex = Math.min(currentPage * pageSize, total)

  const handleRowClick = (event: AuditEvent) => {
    setSelectedEvent(event)
    setIsModalOpen(true)
  }

  const handleSort = (field: string) => {
    if (!onSort) return

    const newDirection =
      sortBy === field && sortDirection === 'DESC' ? 'ASC' : 'DESC'
    onSort(field, newDirection)
  }

  const SortIcon = ({ field }: { field: string }) => {
    if (sortBy !== field) return null
    return sortDirection === 'DESC' ? (
      <ChevronDownIcon className="h-4 w-4" />
    ) : (
      <ChevronUpIcon className="h-4 w-4" />
    )
  }

  const formatDate = (dateString: string) => {
    const date = new Date(dateString)
    return date.toLocaleString(undefined, {
      month: 'short',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
    })
  }

  if (isLoading) {
    return (
      <div className="flex h-64 items-center justify-center">
        <div className="text-sm text-gray-500">Loading audit logs...</div>
      </div>
    )
  }

  if (events.length === 0) {
    return (
      <div className="flex h-64 items-center justify-center rounded-lg border border-gray-200 bg-gray-50">
        <div className="text-center">
          <p className="text-sm font-medium text-gray-900">No audit events found</p>
          <p className="mt-1 text-sm text-gray-500">
            Try adjusting your filters or time range
          </p>
        </div>
      </div>
    )
  }

  return (
    <div className="flex flex-col">
      <div className="-mx-4 -my-2 overflow-x-auto sm:-mx-6 lg:-mx-8">
        <div className="inline-block min-w-full py-2 align-middle sm:px-6 lg:px-8">
          <div className="overflow-hidden shadow ring-1 ring-black ring-opacity-5 sm:rounded-lg">
            <table className="min-w-full divide-y divide-gray-300">
              <thead className="bg-gray-50">
                <tr>
                  <th
                    scope="col"
                    className="cursor-pointer px-3 py-3.5 text-left text-sm font-semibold text-gray-900"
                    onClick={() => handleSort('created_at')}
                  >
                    <div className="flex items-center gap-1">
                      Timestamp
                      <SortIcon field="created_at" />
                    </div>
                  </th>
                  <th
                    scope="col"
                    className="px-3 py-3.5 text-left text-sm font-semibold text-gray-900"
                  >
                    User
                  </th>
                  <th
                    scope="col"
                    className="cursor-pointer px-3 py-3.5 text-left text-sm font-semibold text-gray-900"
                    onClick={() => handleSort('category')}
                  >
                    <div className="flex items-center gap-1">
                      Category
                      <SortIcon field="category" />
                    </div>
                  </th>
                  <th
                    scope="col"
                    className="px-3 py-3.5 text-left text-sm font-semibold text-gray-900"
                  >
                    Action
                  </th>
                  <th
                    scope="col"
                    className="px-3 py-3.5 text-left text-sm font-semibold text-gray-900"
                  >
                    Resource
                  </th>
                  <th
                    scope="col"
                    className="cursor-pointer px-3 py-3.5 text-left text-sm font-semibold text-gray-900"
                    onClick={() => handleSort('status')}
                  >
                    <div className="flex items-center gap-1">
                      Status
                      <SortIcon field="status" />
                    </div>
                  </th>
                  <th
                    scope="col"
                    className="cursor-pointer px-3 py-3.5 text-left text-sm font-semibold text-gray-900"
                    onClick={() => handleSort('severity')}
                  >
                    <div className="flex items-center gap-1">
                      Severity
                      <SortIcon field="severity" />
                    </div>
                  </th>
                </tr>
              </thead>
              <tbody className="divide-y divide-gray-200 bg-white">
                {events.map((event) => (
                  <tr
                    key={event.id}
                    onClick={() => handleRowClick(event)}
                    className="cursor-pointer hover:bg-gray-50"
                  >
                    <td className="whitespace-nowrap px-3 py-4 text-sm text-gray-500">
                      {formatDate(event.createdAt)}
                    </td>
                    <td className="px-3 py-4 text-sm">
                      <div className="text-gray-900">{event.userEmail}</div>
                      <div className="text-gray-500">{event.ipAddress}</div>
                    </td>
                    <td className="px-3 py-4 text-sm">
                      <CategoryIcon category={event.category} size={18} />
                    </td>
                    <td className="px-3 py-4 text-sm text-gray-900">
                      {event.action}
                    </td>
                    <td className="px-3 py-4 text-sm">
                      <div className="text-gray-900">{event.resourceName}</div>
                      <div className="text-gray-500">{event.resourceType}</div>
                    </td>
                    <td className="px-3 py-4 text-sm">
                      <StatusIcon status={event.status} size={18} />
                    </td>
                    <td className="px-3 py-4 text-sm">
                      <SeverityBadge severity={event.severity} size="sm" />
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>

          {/* Pagination */}
          <div className="mt-4 flex items-center justify-between">
            <div className="text-sm text-gray-700">
              Showing <span className="font-medium">{startIndex}</span> to{' '}
              <span className="font-medium">{endIndex}</span> of{' '}
              <span className="font-medium">{total}</span> results
            </div>
            <div className="flex items-center gap-2">
              <button
                onClick={() => onPageChange(currentPage - 1)}
                disabled={currentPage === 1}
                className="rounded-md border border-gray-300 bg-white px-3 py-2 text-sm font-medium text-gray-700 hover:bg-gray-50 disabled:cursor-not-allowed disabled:opacity-50"
              >
                <ChevronLeftIcon className="h-5 w-5" />
              </button>
              <span className="text-sm text-gray-700">
                Page {currentPage} of {totalPages}
              </span>
              <button
                onClick={() => onPageChange(currentPage + 1)}
                disabled={currentPage === totalPages}
                className="rounded-md border border-gray-300 bg-white px-3 py-2 text-sm font-medium text-gray-700 hover:bg-gray-50 disabled:cursor-not-allowed disabled:opacity-50"
              >
                <ChevronRightIcon className="h-5 w-5" />
              </button>
            </div>
          </div>
        </div>
      </div>

      <AuditEventDetailModal
        event={selectedEvent}
        isOpen={isModalOpen}
        onClose={() => setIsModalOpen(false)}
      />
    </div>
  )
}
