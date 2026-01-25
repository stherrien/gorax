import { Fragment } from 'react'
import { Dialog, Transition } from '@headlessui/react'
import { XMarkIcon, ClipboardDocumentIcon } from '@heroicons/react/24/outline'
import { AuditEvent, EVENT_TYPE_LABELS } from '../../types/audit'
import { SeverityBadge } from './SeverityBadge'
import { StatusIcon } from './StatusIcon'
import { CategoryIcon } from './CategoryIcon'

interface AuditEventDetailModalProps {
  event: AuditEvent | null
  isOpen: boolean
  onClose: () => void
}

export function AuditEventDetailModal({
  event,
  isOpen,
  onClose,
}: AuditEventDetailModalProps) {
  if (!event) return null

  const copyEventId = () => {
    navigator.clipboard.writeText(event.id)
  }

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleString()
  }

  return (
    <Transition appear show={isOpen} as={Fragment}>
      <Dialog as="div" className="relative z-50" onClose={onClose}>
        <Transition.Child
          as={Fragment}
          enter="ease-out duration-300"
          enterFrom="opacity-0"
          enterTo="opacity-100"
          leave="ease-in duration-200"
          leaveFrom="opacity-100"
          leaveTo="opacity-0"
        >
          <div className="fixed inset-0 bg-black bg-opacity-25" />
        </Transition.Child>

        <div className="fixed inset-0 overflow-y-auto">
          <div className="flex min-h-full items-center justify-center p-4">
            <Transition.Child
              as={Fragment}
              enter="ease-out duration-300"
              enterFrom="opacity-0 scale-95"
              enterTo="opacity-100 scale-100"
              leave="ease-in duration-200"
              leaveFrom="opacity-100 scale-100"
              leaveTo="opacity-0 scale-95"
            >
              <Dialog.Panel className="w-full max-w-3xl transform overflow-hidden rounded-lg bg-white shadow-xl transition-all">
                <div className="border-b border-gray-200 bg-white px-6 py-4">
                  <div className="flex items-center justify-between">
                    <Dialog.Title className="text-lg font-semibold text-gray-900">
                      Audit Event Details
                    </Dialog.Title>
                    <button
                      onClick={onClose}
                      className="rounded-md text-gray-400 hover:text-gray-500"
                    >
                      <XMarkIcon className="h-6 w-6" />
                    </button>
                  </div>
                </div>

                <div className="max-h-[70vh] overflow-y-auto px-6 py-4">
                  <div className="space-y-6">
                    {/* Event ID */}
                    <div>
                      <label className="block text-sm font-medium text-gray-700">
                        Event ID
                      </label>
                      <div className="mt-1 flex items-center gap-2">
                        <code className="flex-1 rounded bg-gray-100 px-3 py-2 text-sm">
                          {event.id}
                        </code>
                        <button
                          onClick={copyEventId}
                          className="rounded-md p-2 text-gray-400 hover:bg-gray-100 hover:text-gray-600"
                          title="Copy Event ID"
                        >
                          <ClipboardDocumentIcon className="h-5 w-5" />
                        </button>
                      </div>
                    </div>

                    {/* Status and Severity */}
                    <div className="grid grid-cols-2 gap-4">
                      <div>
                        <label className="block text-sm font-medium text-gray-700">
                          Status
                        </label>
                        <div className="mt-1">
                          <StatusIcon status={event.status} showLabel size={18} />
                        </div>
                      </div>
                      <div>
                        <label className="block text-sm font-medium text-gray-700">
                          Severity
                        </label>
                        <div className="mt-1">
                          <SeverityBadge severity={event.severity} />
                        </div>
                      </div>
                    </div>

                    {/* Category and Event Type */}
                    <div className="grid grid-cols-2 gap-4">
                      <div>
                        <label className="block text-sm font-medium text-gray-700">
                          Category
                        </label>
                        <div className="mt-1">
                          <CategoryIcon category={event.category} showLabel />
                        </div>
                      </div>
                      <div>
                        <label className="block text-sm font-medium text-gray-700">
                          Event Type
                        </label>
                        <div className="mt-1 text-sm text-gray-900">
                          {EVENT_TYPE_LABELS[event.eventType]}
                        </div>
                      </div>
                    </div>

                    {/* Action */}
                    <div>
                      <label className="block text-sm font-medium text-gray-700">
                        Action
                      </label>
                      <div className="mt-1 text-sm text-gray-900">{event.action}</div>
                    </div>

                    {/* User Information */}
                    <div className="grid grid-cols-2 gap-4">
                      <div>
                        <label className="block text-sm font-medium text-gray-700">
                          User Email
                        </label>
                        <div className="mt-1 text-sm text-gray-900">
                          {event.userEmail || 'N/A'}
                        </div>
                      </div>
                      <div>
                        <label className="block text-sm font-medium text-gray-700">
                          User ID
                        </label>
                        <div className="mt-1 text-sm text-gray-900">
                          {event.userId || 'N/A'}
                        </div>
                      </div>
                    </div>

                    {/* Resource Information */}
                    <div>
                      <label className="block text-sm font-medium text-gray-700">
                        Resource
                      </label>
                      <div className="mt-1 space-y-1 text-sm text-gray-900">
                        <div>
                          <span className="font-medium">Type:</span> {event.resourceType}
                        </div>
                        <div>
                          <span className="font-medium">Name:</span>{' '}
                          {event.resourceName || 'N/A'}
                        </div>
                        <div>
                          <span className="font-medium">ID:</span>{' '}
                          {event.resourceId || 'N/A'}
                        </div>
                      </div>
                    </div>

                    {/* Network Information */}
                    <div className="grid grid-cols-2 gap-4">
                      <div>
                        <label className="block text-sm font-medium text-gray-700">
                          IP Address
                        </label>
                        <div className="mt-1 text-sm text-gray-900">
                          {event.ipAddress || 'N/A'}
                        </div>
                      </div>
                      <div>
                        <label className="block text-sm font-medium text-gray-700">
                          Timestamp
                        </label>
                        <div className="mt-1 text-sm text-gray-900">
                          {formatDate(event.createdAt)}
                        </div>
                      </div>
                    </div>

                    {/* User Agent */}
                    {event.userAgent && (
                      <div>
                        <label className="block text-sm font-medium text-gray-700">
                          User Agent
                        </label>
                        <div className="mt-1 text-sm text-gray-600">
                          {event.userAgent}
                        </div>
                      </div>
                    )}

                    {/* Error Message */}
                    {event.errorMessage && (
                      <div>
                        <label className="block text-sm font-medium text-gray-700">
                          Error Message
                        </label>
                        <div className="mt-1 rounded-md bg-red-50 p-3 text-sm text-red-800">
                          {event.errorMessage}
                        </div>
                      </div>
                    )}

                    {/* Metadata */}
                    {event.metadata && Object.keys(event.metadata).length > 0 && (
                      <div>
                        <label className="block text-sm font-medium text-gray-700">
                          Metadata
                        </label>
                        <pre className="mt-1 overflow-x-auto rounded-md bg-gray-100 p-3 text-xs">
                          {JSON.stringify(event.metadata, null, 2)}
                        </pre>
                      </div>
                    )}
                  </div>
                </div>

                <div className="border-t border-gray-200 bg-gray-50 px-6 py-4">
                  <button
                    onClick={onClose}
                    className="w-full rounded-md bg-white px-4 py-2 text-sm font-medium text-gray-700 shadow-sm ring-1 ring-inset ring-gray-300 hover:bg-gray-50"
                  >
                    Close
                  </button>
                </div>
              </Dialog.Panel>
            </Transition.Child>
          </div>
        </div>
      </Dialog>
    </Transition>
  )
}
