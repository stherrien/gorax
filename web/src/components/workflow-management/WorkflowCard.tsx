import { Link } from 'react-router-dom'
import type { Workflow } from '../../api/workflows'

interface WorkflowCardProps {
  workflow: Workflow
  isSelected?: boolean
  onSelect?: (id: string) => void
  onRun?: (id: string) => void
  onDelete?: (id: string) => void
  running?: boolean
  deleting?: boolean
}

const statusColors: Record<string, string> = {
  active: 'bg-green-500/20 text-green-400 border-green-500/30',
  draft: 'bg-yellow-500/20 text-yellow-400 border-yellow-500/30',
  inactive: 'bg-gray-500/20 text-gray-400 border-gray-500/30',
}

const statusIcons: Record<string, React.ReactNode> = {
  active: (
    <svg className="w-3 h-3" fill="currentColor" viewBox="0 0 24 24">
      <circle cx="12" cy="12" r="4" />
    </svg>
  ),
  draft: (
    <svg className="w-3 h-3" fill="none" viewBox="0 0 24 24" stroke="currentColor">
      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z" />
    </svg>
  ),
  inactive: (
    <svg className="w-3 h-3" fill="none" viewBox="0 0 24 24" stroke="currentColor">
      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M10 9v6m4-6v6m7-3a9 9 0 11-18 0 9 9 0 0118 0z" />
    </svg>
  ),
}

export function WorkflowCard({
  workflow,
  isSelected,
  onSelect,
  onRun,
  onDelete,
  running,
  deleting,
}: WorkflowCardProps) {
  const formatDate = (dateString: string) => {
    const date = new Date(dateString)
    const now = new Date()
    const diffMs = now.getTime() - date.getTime()
    const diffDays = Math.floor(diffMs / (1000 * 60 * 60 * 24))

    if (diffDays === 0) {
      return 'Today'
    } else if (diffDays === 1) {
      return 'Yesterday'
    } else if (diffDays < 7) {
      return `${diffDays} days ago`
    } else {
      return date.toLocaleDateString()
    }
  }

  const nodeCount = workflow.definition?.nodes?.length ?? 0

  return (
    <div
      className={`bg-gray-800 rounded-lg border ${
        isSelected ? 'border-primary-500 ring-2 ring-primary-500/20' : 'border-gray-700'
      } hover:border-gray-600 transition-all`}
    >
      <div className="p-4">
        {/* Header */}
        <div className="flex items-start justify-between gap-3">
          <div className="flex items-start gap-3 flex-1 min-w-0">
            {onSelect && (
              <input
                type="checkbox"
                checked={isSelected}
                onChange={() => onSelect(workflow.id)}
                className="mt-1 w-4 h-4 bg-gray-700 border-gray-600 rounded text-primary-600 focus:ring-primary-500 focus:ring-2"
                onClick={(e) => e.stopPropagation()}
              />
            )}
            <div className="flex-1 min-w-0">
              <Link
                to={`/workflows/${workflow.id}`}
                className="text-white font-medium hover:text-primary-400 transition-colors block truncate"
              >
                {workflow.name}
              </Link>
              {workflow.description && (
                <p className="text-gray-400 text-sm mt-1 line-clamp-2">
                  {workflow.description}
                </p>
              )}
            </div>
          </div>

          {/* Status badge */}
          <span
            className={`inline-flex items-center gap-1 px-2 py-1 text-xs font-medium rounded-full border ${
              statusColors[workflow.status] || statusColors.inactive
            }`}
          >
            {statusIcons[workflow.status]}
            {workflow.status}
          </span>
        </div>

        {/* Stats */}
        <div className="mt-4 flex items-center gap-4 text-sm text-gray-400">
          <span className="flex items-center gap-1">
            <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 3v2m6-2v2M9 19v2m6-2v2M5 9H3m2 6H3m18-6h-2m2 6h-2M7 19h10a2 2 0 002-2V7a2 2 0 00-2-2H7a2 2 0 00-2 2v10a2 2 0 002 2zM9 9h6v6H9V9z" />
            </svg>
            {nodeCount} nodes
          </span>
          <span className="flex items-center gap-1">
            <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M7 7h.01M7 3h5c.512 0 1.024.195 1.414.586l7 7a2 2 0 010 2.828l-7 7a2 2 0 01-2.828 0l-7-7A1.994 1.994 0 013 12V7a4 4 0 014-4z" />
            </svg>
            v{workflow.version}
          </span>
          <span className="flex items-center gap-1">
            <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
            </svg>
            {formatDate(workflow.updatedAt)}
          </span>
        </div>

        {/* Actions */}
        <div className="mt-4 pt-4 border-t border-gray-700 flex items-center gap-2">
          <Link
            to={`/workflows/${workflow.id}`}
            className="flex-1 px-3 py-1.5 bg-gray-700 text-white text-sm font-medium rounded-lg hover:bg-gray-600 transition-colors text-center"
          >
            Edit
          </Link>
          {workflow.status === 'active' && onRun && (
            <button
              onClick={() => onRun(workflow.id)}
              disabled={running}
              className="flex-1 px-3 py-1.5 bg-primary-600 text-white text-sm font-medium rounded-lg hover:bg-primary-700 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
            >
              {running ? 'Running...' : 'Run'}
            </button>
          )}
          {onDelete && (
            <button
              onClick={() => onDelete(workflow.id)}
              disabled={deleting}
              className="px-3 py-1.5 text-red-400 text-sm font-medium rounded-lg hover:bg-red-500/10 transition-colors disabled:opacity-50"
              title="Delete workflow"
            >
              <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
              </svg>
            </button>
          )}
        </div>
      </div>
    </div>
  )
}
