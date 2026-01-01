import { useState } from 'react'
import { Link } from 'react-router-dom'
import { useWorkflows, useWorkflowMutations } from '../hooks/useWorkflows'
import { WorkflowSelectionProvider, useWorkflowSelection } from '../components/workflows/WorkflowSelectionContext'
import { BulkActionsToolbar } from '../components/workflows/BulkActionsToolbar'

function WorkflowListContent() {
  const { workflows, loading, error, refetch } = useWorkflows()
  const { deleteWorkflow, executeWorkflow, deleting, executing } = useWorkflowMutations()
  const {
    selectedWorkflowIds,
    toggleSelection,
    selectAll,
    clearSelection,
    isSelected,
  } = useWorkflowSelection()

  const [deleteConfirm, setDeleteConfirm] = useState<string | null>(null)
  const [deleteError, setDeleteError] = useState<string | null>(null)
  const [executionMessage, setExecutionMessage] = useState<string | null>(null)

  const handleDeleteClick = (workflowId: string) => {
    setDeleteConfirm(workflowId)
    setDeleteError(null)
  }

  const handleDeleteConfirm = async () => {
    if (!deleteConfirm) return

    try {
      await deleteWorkflow(deleteConfirm)
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

  const handleRun = async (workflowId: string) => {
    setExecutionMessage(null)

    try {
      const result = await executeWorkflow(workflowId)
      setExecutionMessage(`Execution started: ${result.executionId}`)
    } catch (error: any) {
      setExecutionMessage(`Execution failed: ${error.message}`)
    }
  }

  const handleSelectAll = (e: React.ChangeEvent<HTMLInputElement>) => {
    if (e.target.checked && workflows) {
      selectAll(workflows.map((w) => w.id))
    } else {
      clearSelection()
    }
  }

  const workflowCount = workflows?.length ?? 0
  const allSelected = workflowCount > 0 && selectedWorkflowIds.length === workflowCount
  const someSelected = selectedWorkflowIds.length > 0 && selectedWorkflowIds.length < workflowCount

  // Loading state
  if (loading) {
    return (
      <div className="h-64 flex items-center justify-center">
        <div className="text-white text-lg">Loading workflows...</div>
      </div>
    )
  }

  // Error state
  if (error) {
    return (
      <div className="h-64 flex items-center justify-center">
        <div className="text-center">
          <div className="text-red-400 text-lg mb-4">Failed to fetch workflows</div>
          <div className="text-gray-400 text-sm">{error.message}</div>
        </div>
      </div>
    )
  }

  // Empty state
  if (workflowCount === 0) {
    return (
      <div>
        <div className="flex justify-between items-center mb-6">
          <h1 className="text-2xl font-bold text-white">Workflows</h1>
          <Link
            to="/workflows/new"
            className="px-4 py-2 bg-primary-600 text-white rounded-lg text-sm font-medium hover:bg-primary-700 transition-colors"
          >
            New Workflow
          </Link>
        </div>

        <div className="h-64 flex items-center justify-center bg-gray-800 rounded-lg">
          <div className="text-center">
            <div className="text-gray-400 text-lg mb-4">No workflows found</div>
            <Link
              to="/workflows/new"
              className="px-4 py-2 bg-primary-600 text-white rounded-lg text-sm font-medium hover:bg-primary-700 transition-colors inline-block"
            >
              Create your first workflow
            </Link>
          </div>
        </div>
      </div>
    )
  }

  return (
    <div>
      <div className="flex justify-between items-center mb-6">
        <h1 className="text-2xl font-bold text-white">Workflows</h1>
        <Link
          to="/workflows/new"
          className="px-4 py-2 bg-primary-600 text-white rounded-lg text-sm font-medium hover:bg-primary-700 transition-colors"
        >
          New Workflow
        </Link>
      </div>

      {executionMessage && (
        <div className="mb-4 p-3 bg-gray-800 rounded-lg text-white text-sm">
          {executionMessage}
        </div>
      )}

      <BulkActionsToolbar
        selectedCount={selectedWorkflowIds.length}
        selectedWorkflowIds={selectedWorkflowIds}
        onClearSelection={clearSelection}
        onOperationComplete={refetch}
      />

      <div className="bg-gray-800 rounded-lg overflow-hidden">
        <table className="w-full">
          <thead>
            <tr className="border-b border-gray-700">
              <th className="px-6 py-4 w-12">
                <input
                  type="checkbox"
                  checked={allSelected}
                  ref={(input) => {
                    if (input) {
                      input.indeterminate = someSelected
                    }
                  }}
                  onChange={handleSelectAll}
                  className="w-4 h-4 bg-gray-700 border-gray-600 rounded text-primary-600 focus:ring-primary-500 focus:ring-2"
                />
              </th>
              <th className="text-left px-6 py-4 text-sm font-medium text-gray-400">Name</th>
              <th className="text-left px-6 py-4 text-sm font-medium text-gray-400">Status</th>
              <th className="text-left px-6 py-4 text-sm font-medium text-gray-400">Version</th>
              <th className="text-left px-6 py-4 text-sm font-medium text-gray-400">Updated</th>
              <th className="text-right px-6 py-4 text-sm font-medium text-gray-400">Actions</th>
            </tr>
          </thead>
          <tbody>
            {workflows.map((workflow) => (
              <tr key={workflow.id} className="border-b border-gray-700 hover:bg-gray-700/50">
                <td className="px-6 py-4">
                  <input
                    type="checkbox"
                    checked={isSelected(workflow.id)}
                    onChange={() => toggleSelection(workflow.id)}
                    className="w-4 h-4 bg-gray-700 border-gray-600 rounded text-primary-600 focus:ring-primary-500 focus:ring-2"
                    onClick={(e) => e.stopPropagation()}
                  />
                </td>
                <td className="px-6 py-4">
                  <Link to={`/workflows/${workflow.id}`} className="hover:text-primary-400">
                    <p className="text-white font-medium">{workflow.name}</p>
                    {workflow.description && (
                      <p className="text-gray-400 text-sm">{workflow.description}</p>
                    )}
                  </Link>
                </td>
                <td className="px-6 py-4">
                  <StatusBadge status={workflow.status} />
                </td>
                <td className="px-6 py-4 text-gray-300">v{workflow.version}</td>
                <td className="px-6 py-4 text-gray-300">
                  {new Date(workflow.updatedAt).toLocaleDateString()}
                </td>
                <td className="px-6 py-4 text-right">
                  <div className="flex justify-end space-x-2">
                    <Link
                      to={`/workflows/${workflow.id}`}
                      className="px-3 py-1 text-sm text-gray-300 hover:text-white transition-colors"
                    >
                      Edit
                    </Link>
                    {workflow.status === 'active' && (
                      <button
                        onClick={() => handleRun(workflow.id)}
                        disabled={executing}
                        className="px-3 py-1 text-sm text-primary-400 hover:text-primary-300 transition-colors disabled:opacity-50"
                      >
                        Run
                      </button>
                    )}
                    <button
                      onClick={() => handleDeleteClick(workflow.id)}
                      disabled={deleting}
                      className="px-3 py-1 text-sm text-red-400 hover:text-red-300 transition-colors disabled:opacity-50"
                    >
                      Delete
                    </button>
                  </div>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      {/* Delete Confirmation Dialog */}
      {deleteConfirm && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
          <div className="bg-gray-800 rounded-lg p-6 w-96">
            <h3 className="text-white text-lg font-semibold mb-4">Delete Workflow</h3>
            <p className="text-gray-400 mb-4">
              Are you sure you want to delete this workflow? This action cannot be undone.
            </p>
            {deleteError && (
              <div className="text-xs text-red-400 mb-4">{deleteError}</div>
            )}
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

function StatusBadge({ status }: { status: string }) {
  const colors = {
    active: 'bg-green-500/20 text-green-400',
    draft: 'bg-yellow-500/20 text-yellow-400',
    inactive: 'bg-gray-500/20 text-gray-400',
  }

  return (
    <span
      className={`inline-flex px-2 py-1 text-xs font-medium rounded-full ${
        colors[status as keyof typeof colors]
      }`}
    >
      {status}
    </span>
  )
}

export default function WorkflowList() {
  return (
    <WorkflowSelectionProvider>
      <WorkflowListContent />
    </WorkflowSelectionProvider>
  )
}
