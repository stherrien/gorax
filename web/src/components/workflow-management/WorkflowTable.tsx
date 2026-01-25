import { Link } from 'react-router-dom'
import type { Workflow } from '../../api/workflows'

interface WorkflowTableProps {
  workflows: Workflow[]
  selectedIds: string[]
  onSelectAll: (selected: boolean) => void
  onSelectOne: (id: string) => void
  onRun?: (id: string) => void
  onDelete?: (id: string) => void
  running?: boolean
  deleting?: boolean
}

const statusColors: Record<string, string> = {
  active: 'bg-green-500/20 text-green-400',
  draft: 'bg-yellow-500/20 text-yellow-400',
  inactive: 'bg-gray-500/20 text-gray-400',
}

export function WorkflowTable({
  workflows,
  selectedIds,
  onSelectAll,
  onSelectOne,
  onRun,
  onDelete,
  running,
  deleting,
}: WorkflowTableProps) {
  const allSelected = workflows.length > 0 && selectedIds.length === workflows.length
  const someSelected = selectedIds.length > 0 && selectedIds.length < workflows.length

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleDateString('en-US', {
      month: 'short',
      day: 'numeric',
      year: 'numeric',
    })
  }

  const formatTime = (dateString: string) => {
    return new Date(dateString).toLocaleTimeString('en-US', {
      hour: '2-digit',
      minute: '2-digit',
    })
  }

  return (
    <div className="bg-gray-800 rounded-lg overflow-hidden">
      <div className="overflow-x-auto">
        <table className="w-full">
          <thead>
            <tr className="border-b border-gray-700">
              <th className="px-4 py-3 w-12">
                <input
                  type="checkbox"
                  checked={allSelected}
                  ref={(input) => {
                    if (input) {
                      input.indeterminate = someSelected
                    }
                  }}
                  onChange={(e) => onSelectAll(e.target.checked)}
                  className="w-4 h-4 bg-gray-700 border-gray-600 rounded text-primary-600 focus:ring-primary-500 focus:ring-2"
                />
              </th>
              <th className="text-left px-4 py-3 text-sm font-medium text-gray-400">
                Name
              </th>
              <th className="text-left px-4 py-3 text-sm font-medium text-gray-400">
                Status
              </th>
              <th className="text-left px-4 py-3 text-sm font-medium text-gray-400 hidden sm:table-cell">
                Version
              </th>
              <th className="text-left px-4 py-3 text-sm font-medium text-gray-400 hidden md:table-cell">
                Nodes
              </th>
              <th className="text-left px-4 py-3 text-sm font-medium text-gray-400 hidden lg:table-cell">
                Updated
              </th>
              <th className="text-right px-4 py-3 text-sm font-medium text-gray-400">
                Actions
              </th>
            </tr>
          </thead>
          <tbody>
            {workflows.map((workflow) => (
              <tr
                key={workflow.id}
                className="border-b border-gray-700 hover:bg-gray-700/50 transition-colors"
              >
                <td className="px-4 py-3">
                  <input
                    type="checkbox"
                    checked={selectedIds.includes(workflow.id)}
                    onChange={() => onSelectOne(workflow.id)}
                    className="w-4 h-4 bg-gray-700 border-gray-600 rounded text-primary-600 focus:ring-primary-500 focus:ring-2"
                    onClick={(e) => e.stopPropagation()}
                  />
                </td>
                <td className="px-4 py-3">
                  <Link
                    to={`/workflows/${workflow.id}`}
                    className="hover:text-primary-400 transition-colors"
                  >
                    <p className="text-white font-medium">{workflow.name}</p>
                    {workflow.description && (
                      <p className="text-gray-400 text-sm truncate max-w-xs">
                        {workflow.description}
                      </p>
                    )}
                  </Link>
                </td>
                <td className="px-4 py-3">
                  <span
                    className={`inline-flex px-2 py-1 text-xs font-medium rounded-full ${
                      statusColors[workflow.status] || statusColors.inactive
                    }`}
                  >
                    {workflow.status}
                  </span>
                </td>
                <td className="px-4 py-3 text-gray-300 hidden sm:table-cell">
                  v{workflow.version}
                </td>
                <td className="px-4 py-3 text-gray-300 hidden md:table-cell">
                  {workflow.definition?.nodes?.length ?? 0}
                </td>
                <td className="px-4 py-3 text-gray-300 hidden lg:table-cell">
                  <div>
                    <p className="text-sm">{formatDate(workflow.updatedAt)}</p>
                    <p className="text-xs text-gray-500">{formatTime(workflow.updatedAt)}</p>
                  </div>
                </td>
                <td className="px-4 py-3 text-right">
                  <div className="flex justify-end items-center gap-2">
                    <Link
                      to={`/workflows/${workflow.id}`}
                      className="px-3 py-1 text-sm text-gray-300 hover:text-white transition-colors"
                    >
                      Edit
                    </Link>
                    {workflow.status === 'active' && onRun && (
                      <button
                        onClick={() => onRun(workflow.id)}
                        disabled={running}
                        className="px-3 py-1 text-sm text-primary-400 hover:text-primary-300 transition-colors disabled:opacity-50"
                      >
                        Run
                      </button>
                    )}
                    {onDelete && (
                      <button
                        onClick={() => onDelete(workflow.id)}
                        disabled={deleting}
                        className="px-3 py-1 text-sm text-red-400 hover:text-red-300 transition-colors disabled:opacity-50"
                      >
                        Delete
                      </button>
                    )}
                  </div>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      {/* Empty state */}
      {workflows.length === 0 && (
        <div className="p-8 text-center text-gray-400">
          No workflows found
        </div>
      )}
    </div>
  )
}
