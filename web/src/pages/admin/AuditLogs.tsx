import { useState } from 'react'
import { useAuditEvents, useAuditStats } from '../../hooks/useAudit'
import {
  QueryFilter,
  TimeRangePreset,
  getTimeRangeFromPreset,
} from '../../types/audit'
import { AuditLogTable } from '../../components/audit/AuditLogTable'
import { AuditFilterPanel } from '../../components/audit/AuditFilterPanel'
import { AuditStatsCards } from '../../components/audit/AuditStatsCards'
import { AuditExportButton } from '../../components/audit/AuditExportButton'
import { ArrowPathIcon } from '@heroicons/react/24/outline'

export function AuditLogs() {
  const [currentPage, setCurrentPage] = useState(1)
  const [pageSize] = useState(50)
  const [appliedFilter, setAppliedFilter] = useState<QueryFilter>(() => {
    const timeRange = getTimeRangeFromPreset(TimeRangePreset.Last24Hours)
    return {
      ...timeRange,
      limit: pageSize,
      offset: 0,
      sortBy: 'created_at',
      sortDirection: 'DESC',
    }
  })
  const [workingFilter, setWorkingFilter] = useState<QueryFilter>(appliedFilter)
  const [autoRefresh, setAutoRefresh] = useState(false)

  // Query audit events
  const {
    data: eventsData,
    isLoading: isLoadingEvents,
    refetch: refetchEvents,
  } = useAuditEvents(appliedFilter, true)

  // Query audit stats
  const {
    data: stats,
    isLoading: isLoadingStats,
    refetch: refetchStats,
  } = useAuditStats(
    {
      startDate: appliedFilter.startDate || '',
      endDate: appliedFilter.endDate || '',
    },
    !!appliedFilter.startDate && !!appliedFilter.endDate
  )

  const handlePageChange = (page: number) => {
    setCurrentPage(page)
    setAppliedFilter({
      ...appliedFilter,
      offset: (page - 1) * pageSize,
    })
  }

  const handleSort = (field: string, direction: 'ASC' | 'DESC') => {
    setAppliedFilter({
      ...appliedFilter,
      sortBy: field,
      sortDirection: direction,
      offset: 0,
    })
    setCurrentPage(1)
  }

  const handleApplyFilters = () => {
    setAppliedFilter({
      ...workingFilter,
      limit: pageSize,
      offset: 0,
    })
    setCurrentPage(1)
  }

  const handleResetFilters = () => {
    const timeRange = getTimeRangeFromPreset(TimeRangePreset.Last24Hours)
    const defaultFilter: QueryFilter = {
      ...timeRange,
      limit: pageSize,
      offset: 0,
      sortBy: 'created_at',
      sortDirection: 'DESC',
    }
    setWorkingFilter(defaultFilter)
    setAppliedFilter(defaultFilter)
    setCurrentPage(1)
  }

  const handleRefresh = () => {
    refetchEvents()
    refetchStats()
  }

  return (
    <div className="flex h-screen flex-col">
      {/* Header */}
      <div className="border-b border-gray-200 bg-white px-6 py-4">
        <div className="flex items-center justify-between">
          <div>
            <h1 className="text-2xl font-bold text-gray-900">Audit Logs</h1>
            <p className="mt-1 text-sm text-gray-500">
              View and analyze security and compliance audit logs
            </p>
          </div>
          <div className="flex items-center gap-3">
            <label className="flex items-center gap-2">
              <input
                type="checkbox"
                checked={autoRefresh}
                onChange={(e) => setAutoRefresh(e.target.checked)}
                className="h-4 w-4 rounded border-gray-300 text-indigo-600 focus:ring-indigo-500"
              />
              <span className="text-sm text-gray-700">Auto-refresh</span>
            </label>
            <button
              onClick={handleRefresh}
              className="inline-flex items-center gap-2 rounded-md bg-white px-3 py-2 text-sm font-semibold text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 hover:bg-gray-50"
            >
              <ArrowPathIcon className="h-5 w-5" />
              Refresh
            </button>
            <AuditExportButton filter={appliedFilter} />
          </div>
        </div>
      </div>

      {/* Stats Cards */}
      {stats && (
        <div className="border-b border-gray-200 bg-gray-50 px-6 py-4">
          <AuditStatsCards stats={stats} isLoading={isLoadingStats} />
        </div>
      )}

      {/* Main Content */}
      <div className="flex flex-1 overflow-hidden">
        {/* Filter Panel */}
        <AuditFilterPanel
          filter={workingFilter}
          onFilterChange={setWorkingFilter}
          onApply={handleApplyFilters}
          onReset={handleResetFilters}
        />

        {/* Events Table */}
        <div className="flex-1 overflow-y-auto p-6">
          {eventsData && (
            <AuditLogTable
              events={eventsData.events}
              total={eventsData.total}
              currentPage={currentPage}
              pageSize={pageSize}
              onPageChange={handlePageChange}
              onSort={handleSort}
              sortBy={appliedFilter.sortBy}
              sortDirection={appliedFilter.sortDirection}
              isLoading={isLoadingEvents}
            />
          )}
        </div>
      </div>
    </div>
  )
}
