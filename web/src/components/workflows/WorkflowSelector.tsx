import { useState, useEffect, useMemo } from 'react'
import { workflowAPI, type Workflow } from '../../api/workflows'

interface WorkflowSelectorProps {
  value?: string
  onChange: (workflowId: string, workflowName: string) => void
  excludeWorkflowId?: string // Exclude current workflow to prevent self-reference
  label?: string
  placeholder?: string
  disabled?: boolean
}

export default function WorkflowSelector({
  value,
  onChange,
  excludeWorkflowId,
  label = 'Select Workflow',
  placeholder = 'Choose a workflow...',
  disabled = false,
}: WorkflowSelectorProps) {
  const [workflows, setWorkflows] = useState<Workflow[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [searchTerm, setSearchTerm] = useState('')

  useEffect(() => {
    const fetchWorkflows = async () => {
      try {
        setLoading(true)
        setError(null)
        const data = await workflowAPI.list()
        setWorkflows(data.workflows)
      } catch (err) {
        setError('Failed to load workflows')
        console.error('Error fetching workflows:', err)
      } finally {
        setLoading(false)
      }
    }

    fetchWorkflows()
  }, [])

  // Filter workflows based on search term and exclude current workflow
  const filteredWorkflows = useMemo(() => {
    return workflows
      .filter((wf) => {
        // Exclude current workflow if specified
        if (excludeWorkflowId && wf.id === excludeWorkflowId) {
          return false
        }

        // Only show active workflows
        if (wf.status !== 'active') {
          return false
        }

        // Filter by search term
        if (searchTerm) {
          const term = searchTerm.toLowerCase()
          return (
            wf.name.toLowerCase().includes(term) ||
            wf.description?.toLowerCase().includes(term)
          )
        }

        return true
      })
      .sort((a, b) => {
        // Sort by created date (newest first)
        return new Date(b.createdAt).getTime() - new Date(a.createdAt).getTime()
      })
  }, [workflows, searchTerm, excludeWorkflowId])

  // Find selected workflow
  const selectedWorkflow = workflows.find((wf) => wf.id === value)

  const handleChange = (e: React.ChangeEvent<HTMLSelectElement>) => {
    const workflowId = e.target.value
    const workflow = workflows.find((wf) => wf.id === workflowId)
    if (workflow) {
      onChange(workflowId, workflow.name)
    }
  }

  if (error) {
    return (
      <div className="space-y-2">
        <label className="block text-sm font-medium text-gray-300">{label}</label>
        <div className="text-red-400 text-sm p-2 bg-red-900/20 rounded border border-red-700">
          {error}
        </div>
      </div>
    )
  }

  return (
    <div className="space-y-2">
      <label className="block text-sm font-medium text-gray-300">{label}</label>

      {/* Search input */}
      {workflows.length > 5 && (
        <input
          type="text"
          value={searchTerm}
          onChange={(e) => setSearchTerm(e.target.value)}
          placeholder="Search workflows..."
          className="w-full px-3 py-2 bg-gray-700 border border-gray-600 rounded-lg text-white placeholder-gray-400 focus:outline-none focus:ring-2 focus:ring-blue-500 text-sm"
          disabled={disabled}
        />
      )}

      {/* Workflow selector */}
      <select
        value={value || ''}
        onChange={handleChange}
        disabled={disabled || loading}
        className="w-full px-3 py-2 bg-gray-700 border border-gray-600 rounded-lg text-white focus:outline-none focus:ring-2 focus:ring-blue-500 disabled:opacity-50 disabled:cursor-not-allowed"
      >
        <option value="">{loading ? 'Loading...' : placeholder}</option>
        {filteredWorkflows.map((workflow) => (
          <option key={workflow.id} value={workflow.id}>
            {workflow.name}
          </option>
        ))}
      </select>

      {/* Selected workflow details */}
      {selectedWorkflow && (
        <div className="p-3 bg-gray-700/50 rounded-lg border border-gray-600">
          <div className="flex items-start justify-between">
            <div>
              <p className="text-sm font-medium text-white">{selectedWorkflow.name}</p>
              {selectedWorkflow.description && (
                <p className="text-xs text-gray-400 mt-1">{selectedWorkflow.description}</p>
              )}
            </div>
            <span className="px-2 py-0.5 text-xs font-medium bg-green-500/20 text-green-400 rounded">
              {selectedWorkflow.status}
            </span>
          </div>
        </div>
      )}

      {/* Empty state */}
      {!loading && filteredWorkflows.length === 0 && (
        <div className="text-gray-400 text-sm p-3 bg-gray-700/30 rounded border border-gray-600">
          {searchTerm ? 'No workflows found matching your search' : 'No active workflows available'}
        </div>
      )}

      {/* Count indicator */}
      {!loading && filteredWorkflows.length > 0 && (
        <p className="text-xs text-gray-400">
          {filteredWorkflows.length} workflow{filteredWorkflows.length !== 1 ? 's' : ''} available
        </p>
      )}
    </div>
  )
}
