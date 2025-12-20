import { Handle, Position } from '@xyflow/react'
import { useState, useEffect } from 'react'
import { workflowsApi } from '../../api/workflows'

interface SubWorkflowNodeData {
  label: string
  workflowId?: string
  workflowName?: string
  inputMapping?: Record<string, string>
  outputMapping?: Record<string, string>
  waitForResult?: boolean
  timeoutMs?: number
}

interface SubWorkflowNodeProps {
  data: SubWorkflowNodeData
  selected?: boolean
}

export default function SubWorkflowNode({ data, selected }: SubWorkflowNodeProps) {
  const [workflowName, setWorkflowName] = useState<string>(data.workflowName || '')

  useEffect(() => {
    // Fetch workflow name if we have an ID but no name
    if (data.workflowId && !data.workflowName) {
      workflowsApi.get(data.workflowId)
        .then(workflow => setWorkflowName(workflow.name))
        .catch(() => setWorkflowName('Unknown Workflow'))
    }
  }, [data.workflowId, data.workflowName])

  const inputCount = Object.keys(data.inputMapping || {}).length
  const outputCount = Object.keys(data.outputMapping || {}).length
  const executionMode = data.waitForResult ? 'Sync' : 'Async'

  return (
    <div
      className={`
        px-4 py-3 rounded-lg shadow-lg min-w-[220px]
        bg-gradient-to-br from-indigo-500 to-blue-600
        ${selected ? 'ring-2 ring-white ring-offset-2 ring-offset-gray-900' : ''}
      `}
    >
      {/* Input handle */}
      <Handle
        type="target"
        position={Position.Top}
        className="w-3 h-3 bg-white border-2 border-indigo-500"
      />

      <div className="space-y-2">
        <div className="flex items-center space-x-2">
          <span className="text-lg">🔗</span>
          <div className="flex-1">
            <p className="text-white font-medium text-sm">{data.label}</p>
            <p className="text-white/70 text-xs">Sub-Workflow</p>
          </div>
        </div>

        {/* Workflow info */}
        {(data.workflowId || workflowName) && (
          <div className="mt-2 p-2 bg-white/10 rounded space-y-1">
            <div className="flex items-center space-x-1">
              <span className="text-white/90 text-xs">📋</span>
              <p className="text-white/90 text-xs font-mono truncate">
                {workflowName || data.workflowId}
              </p>
            </div>
          </div>
        )}

        {/* Configuration details */}
        <div className="space-y-1">
          {inputCount > 0 && (
            <div className="flex items-center justify-between text-xs text-white/80">
              <span>Inputs:</span>
              <span className="font-semibold">{inputCount}</span>
            </div>
          )}
          {outputCount > 0 && (
            <div className="flex items-center justify-between text-xs text-white/80">
              <span>Outputs:</span>
              <span className="font-semibold">{outputCount}</span>
            </div>
          )}
          <div className="flex items-center justify-between text-xs text-white/80">
            <span>Mode:</span>
            <span className={`font-semibold px-1.5 py-0.5 rounded ${
              executionMode === 'Sync'
                ? 'bg-green-500/20 text-green-200'
                : 'bg-yellow-500/20 text-yellow-200'
            }`}>
              {executionMode}
            </span>
          </div>
          {data.timeoutMs && data.timeoutMs > 0 && (
            <div className="flex items-center justify-between text-xs text-white/80">
              <span>Timeout:</span>
              <span className="font-semibold">{data.timeoutMs}ms</span>
            </div>
          )}
        </div>
      </div>

      {/* Output handle */}
      <Handle
        type="source"
        position={Position.Bottom}
        className="w-3 h-3 bg-white border-2 border-indigo-500"
      />
    </div>
  )
}
