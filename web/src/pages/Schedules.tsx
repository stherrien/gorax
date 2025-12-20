import { useState, useMemo } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { useSchedules, useScheduleMutations } from '../hooks/useSchedules'
import { useWorkflows } from '../hooks/useWorkflows'
import ScheduleList from '../components/schedule/ScheduleList'
import ScheduleCalendar from '../components/schedule/ScheduleCalendar'
import ScheduleTimeline from '../components/schedule/ScheduleTimeline'

type ViewMode = 'list' | 'calendar' | 'timeline'
type StatusFilter = 'all' | 'enabled' | 'disabled'

export default function Schedules() {
  const navigate = useNavigate()
  const [viewMode, setViewMode] = useState<ViewMode>('list')
  const [statusFilter, setStatusFilter] = useState<StatusFilter>('all')
  const [searchQuery, setSearchQuery] = useState('')
  const [sortBy, setSortBy] = useState<'name' | 'nextRun' | 'lastRun'>('nextRun')

  const params = useMemo(() => {
    const p: any = {}
    if (statusFilter !== 'all') {
      p.enabled = statusFilter === 'enabled'
    }
    if (searchQuery) {
      p.search = searchQuery
    }
    return p
  }, [statusFilter, searchQuery])

  const { schedules, total, loading, error, refetch } = useSchedules(params)
  const { workflows } = useWorkflows()
  const { toggleSchedule, deleteSchedule, updating, deleting } = useScheduleMutations()

  const [toggleError, setToggleError] = useState<string | null>(null)
  const [deleteConfirm, setDeleteConfirm] = useState<string | null>(null)
  const [deleteError, setDeleteError] = useState<string | null>(null)

  const handleToggle = async (id: string, currentEnabled: boolean) => {
    try {
      setToggleError(null)
      await toggleSchedule(id, !currentEnabled)
      await refetch()
    } catch (error: any) {
      setToggleError(error.message || 'Toggle failed')
      setTimeout(() => setToggleError(null), 3000)
    }
  }

  const handleDeleteClick = (id: string) => {
    setDeleteConfirm(id)
    setDeleteError(null)
  }

  const handleDeleteConfirm = async () => {
    if (!deleteConfirm) return

    try {
      await deleteSchedule(deleteConfirm)
      setDeleteConfirm(null)
      setDeleteError(null)
      await refetch()
    } catch (error: any) {
      setDeleteError(error.message || 'Delete failed')
    }
  }

  const handleDeleteCancel = () => {
    setDeleteConfirm(null)
    setDeleteError(null)
  }

  const handleEdit = (id: string) => {
    navigate(`/schedules/${id}/edit`)
  }

  if (loading) {
    return (
      <div className="h-64 flex items-center justify-center">
        <div className="text-white text-lg">Loading schedules...</div>
      </div>
    )
  }

  if (error) {
    return (
      <div className="h-64 flex items-center justify-center">
        <div className="text-center">
          <div className="text-red-400 text-lg mb-4">Failed to fetch schedules</div>
          <div className="text-gray-400 text-sm">{error.message}</div>
        </div>
      </div>
    )
  }

  return (
    <div>
      <div className="flex justify-between items-center mb-6">
        <div>
          <h1 className="text-2xl font-bold text-white">Schedules</h1>
          <p className="text-gray-400 text-sm mt-1">{total} total</p>
        </div>
        <Link
          to="/schedules/new"
          className="px-4 py-2 bg-primary-600 text-white rounded-lg text-sm font-medium hover:bg-primary-700 transition-colors"
        >
          Create Schedule
        </Link>
      </div>

      {toggleError && (
        <div className="mb-4 p-3 bg-red-800/50 rounded-lg text-red-200 text-sm">
          {toggleError}
        </div>
      )}

      <div className="mb-6 space-y-4">
        <div className="flex items-center space-x-4">
          <TabButton
            active={viewMode === 'list'}
            onClick={() => setViewMode('list')}
          >
            List View
          </TabButton>
          <TabButton
            active={viewMode === 'calendar'}
            onClick={() => setViewMode('calendar')}
          >
            Calendar View
          </TabButton>
          <TabButton
            active={viewMode === 'timeline'}
            onClick={() => setViewMode('timeline')}
          >
            Timeline View
          </TabButton>
        </div>

        <div className="flex items-center space-x-4">
          <div>
            <label htmlFor="status-filter" className="sr-only">
              Status
            </label>
            <select
              id="status-filter"
              value={statusFilter}
              onChange={(e) => setStatusFilter(e.target.value as StatusFilter)}
              className="px-3 py-2 bg-gray-800 text-white rounded-lg text-sm border border-gray-700 focus:outline-none focus:ring-2 focus:ring-primary-500"
            >
              <option value="all">All Schedules</option>
              <option value="enabled">Enabled Only</option>
              <option value="disabled">Disabled Only</option>
            </select>
          </div>

          {viewMode === 'list' && (
            <div>
              <label htmlFor="sort-by" className="sr-only">
                Sort by
              </label>
              <select
                id="sort-by"
                value={sortBy}
                onChange={(e) => setSortBy(e.target.value as any)}
                className="px-3 py-2 bg-gray-800 text-white rounded-lg text-sm border border-gray-700 focus:outline-none focus:ring-2 focus:ring-primary-500"
              >
                <option value="nextRun">Sort by Next Run</option>
                <option value="name">Sort by Name</option>
                <option value="lastRun">Sort by Last Run</option>
              </select>
            </div>
          )}

          <div className="flex-1">
            <label htmlFor="search" className="sr-only">
              Search
            </label>
            <input
              id="search"
              type="text"
              placeholder="Search schedules..."
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              className="w-full px-3 py-2 bg-gray-800 text-white rounded-lg text-sm border border-gray-700 focus:outline-none focus:ring-2 focus:ring-primary-500"
            />
          </div>
        </div>
      </div>

      {viewMode === 'list' && (
        <ScheduleList
          schedules={schedules}
          workflows={workflows}
          onToggle={handleToggle}
          onEdit={handleEdit}
          onDelete={handleDeleteClick}
          sortBy={sortBy}
          disabled={updating || deleting}
        />
      )}

      {viewMode === 'calendar' && (
        <ScheduleCalendar schedules={schedules} />
      )}

      {viewMode === 'timeline' && (
        <ScheduleTimeline schedules={schedules} workflows={workflows} />
      )}

      {deleteConfirm && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
          <div className="bg-gray-800 rounded-lg p-6 w-96">
            <h3 className="text-white text-lg font-semibold mb-4">Delete Schedule</h3>
            <p className="text-gray-400 mb-4">
              Are you sure you want to delete this schedule? This action cannot be undone.
            </p>
            {deleteError && <div className="text-xs text-red-400 mb-4">{deleteError}</div>}
            <div className="flex space-x-2 justify-end">
              <button
                onClick={handleDeleteCancel}
                disabled={deleting}
                className="px-4 py-2 bg-gray-700 text-white rounded-lg text-sm font-medium hover:bg-gray-600 transition-colors disabled:opacity-50"
              >
                Cancel
              </button>
              <button
                onClick={handleDeleteConfirm}
                disabled={deleting}
                className="px-4 py-2 bg-red-600 text-white rounded-lg text-sm font-medium hover:bg-red-700 transition-colors disabled:opacity-50"
              >
                {deleting ? 'Deleting...' : 'Confirm'}
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}

interface TabButtonProps {
  active: boolean
  onClick: () => void
  children: React.ReactNode
}

function TabButton({ active, onClick, children }: TabButtonProps) {
  return (
    <button
      onClick={onClick}
      className={`px-4 py-2 text-sm font-medium rounded-lg transition-colors ${
        active
          ? 'bg-primary-600 text-white'
          : 'bg-gray-800 text-gray-400 hover:text-white hover:bg-gray-700'
      }`}
    >
      {children}
    </button>
  )
}
