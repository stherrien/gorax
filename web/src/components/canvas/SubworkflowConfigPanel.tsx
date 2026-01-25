import { useState } from 'react'
import WorkflowSelector from '../workflows/WorkflowSelector'

interface SubworkflowConfig {
  workflow_id: string
  workflow_name?: string
  input_mapping: Record<string, string>
  output_mapping: Record<string, string>
  mode: 'sync' | 'async'
  timeout?: string
  inherit_context: boolean
}

interface SubworkflowConfigPanelProps {
  config: Partial<SubworkflowConfig>
  onChange: (config: Partial<SubworkflowConfig>) => void
  currentWorkflowId?: string
}

export default function SubworkflowConfigPanel({
  config,
  onChange,
  currentWorkflowId,
}: SubworkflowConfigPanelProps) {
  const [newInputKey, setNewInputKey] = useState('')
  const [newInputValue, setNewInputValue] = useState('')
  const [newOutputKey, setNewOutputKey] = useState('')
  const [newOutputValue, setNewOutputValue] = useState('')

  const inputMapping = config.input_mapping || {}
  const outputMapping = config.output_mapping || {}
  const mode = config.mode || 'sync'
  const timeout = config.timeout || '30s'
  const inheritContext = config.inherit_context ?? false

  const handleWorkflowChange = (workflowId: string, workflowName: string) => {
    onChange({
      ...config,
      workflow_id: workflowId,
      workflow_name: workflowName,
    })
  }

  const handleAddInputMapping = () => {
    if (newInputKey && newInputValue) {
      onChange({
        ...config,
        input_mapping: {
          ...inputMapping,
          [newInputKey]: newInputValue,
        },
      })
      setNewInputKey('')
      setNewInputValue('')
    }
  }

  const handleRemoveInputMapping = (key: string) => {
    const { [key]: _, ...rest } = inputMapping
    onChange({
      ...config,
      input_mapping: rest,
    })
  }

  const handleAddOutputMapping = () => {
    if (newOutputKey && newOutputValue) {
      onChange({
        ...config,
        output_mapping: {
          ...outputMapping,
          [newOutputKey]: newOutputValue,
        },
      })
      setNewOutputKey('')
      setNewOutputValue('')
    }
  }

  const handleRemoveOutputMapping = (key: string) => {
    const { [key]: _, ...rest } = outputMapping
    onChange({
      ...config,
      output_mapping: rest,
    })
  }

  const handleModeChange = (newMode: 'sync' | 'async') => {
    onChange({
      ...config,
      mode: newMode,
    })
  }

  const handleTimeoutChange = (newTimeout: string) => {
    onChange({
      ...config,
      timeout: newTimeout,
    })
  }

  const handleInheritContextChange = (inherit: boolean) => {
    onChange({
      ...config,
      inherit_context: inherit,
    })
  }

  return (
    <div className="space-y-6">
      {/* Workflow Selection */}
      <WorkflowSelector
        value={config.workflow_id}
        onChange={handleWorkflowChange}
        excludeWorkflowId={currentWorkflowId}
        label="Target Workflow"
        placeholder="Select a workflow to execute..."
      />

      {/* Execution Mode */}
      <div className="space-y-2">
        <label className="block text-sm font-medium text-gray-300">Execution Mode</label>
        <div className="flex gap-3">
          <button
            type="button"
            onClick={() => handleModeChange('sync')}
            className={`flex-1 px-4 py-2 rounded-lg border text-sm font-medium transition-colors ${
              mode === 'sync'
                ? 'bg-blue-600 border-blue-500 text-white'
                : 'bg-gray-700 border-gray-600 text-gray-300 hover:border-gray-500'
            }`}
          >
            <div className="flex flex-col items-center">
              <span className="text-lg mb-1">⏱️</span>
              <span>Synchronous</span>
              <span className="text-xs opacity-75 mt-1">Wait for completion</span>
            </div>
          </button>
          <button
            type="button"
            onClick={() => handleModeChange('async')}
            className={`flex-1 px-4 py-2 rounded-lg border text-sm font-medium transition-colors ${
              mode === 'async'
                ? 'bg-blue-600 border-blue-500 text-white'
                : 'bg-gray-700 border-gray-600 text-gray-300 hover:border-gray-500'
            }`}
          >
            <div className="flex flex-col items-center">
              <span className="text-lg mb-1">🚀</span>
              <span>Asynchronous</span>
              <span className="text-xs opacity-75 mt-1">Fire and forget</span>
            </div>
          </button>
        </div>
      </div>

      {/* Timeout (only for sync mode) */}
      {mode === 'sync' && (
        <div className="space-y-2">
          <label className="block text-sm font-medium text-gray-300">
            Timeout
            <span className="ml-2 text-xs text-gray-500">(e.g., 30s, 5m, 1h)</span>
          </label>
          <input
            type="text"
            value={timeout}
            onChange={(e) => handleTimeoutChange(e.target.value)}
            placeholder="30s"
            className="w-full px-3 py-2 bg-gray-700 border border-gray-600 rounded-lg text-white placeholder-gray-400 focus:outline-none focus:ring-2 focus:ring-blue-500"
          />
          <p className="text-xs text-gray-400">
            Maximum time to wait for the subworkflow to complete
          </p>
        </div>
      )}

      {/* Inherit Context */}
      <div className="flex items-center justify-between p-3 bg-gray-700/50 rounded-lg border border-gray-600">
        <div>
          <p className="text-sm font-medium text-white">Inherit Parent Context</p>
          <p className="text-xs text-gray-400 mt-1">
            Pass all parent workflow data to the subworkflow
          </p>
        </div>
        <label className="relative inline-flex items-center cursor-pointer">
          <input
            type="checkbox"
            checked={inheritContext}
            onChange={(e) => handleInheritContextChange(e.target.checked)}
            className="sr-only peer"
          />
          <div className="w-11 h-6 bg-gray-600 peer-focus:outline-none peer-focus:ring-2 peer-focus:ring-blue-500 rounded-full peer peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-gray-300 after:border after:rounded-full after:h-5 after:w-5 after:transition-all peer-checked:bg-blue-600"></div>
        </label>
      </div>

      {/* Input Mapping */}
      <div className="space-y-3">
        <div className="flex items-center justify-between">
          <label className="block text-sm font-medium text-gray-300">Input Mapping</label>
          <span className="text-xs text-gray-500">
            {Object.keys(inputMapping).length} mapping{Object.keys(inputMapping).length !== 1 ? 's' : ''}
          </span>
        </div>

        {/* Existing mappings */}
        <div className="space-y-2">
          {Object.entries(inputMapping).map(([key, value]) => (
            <div key={key} className="flex items-center gap-2 p-2 bg-gray-700/50 rounded border border-gray-600">
              <div className="flex-1 grid grid-cols-2 gap-2">
                <span className="text-sm text-white font-mono truncate">{key}</span>
                <span className="text-sm text-gray-400 font-mono truncate">{value}</span>
              </div>
              <button
                type="button"
                onClick={() => handleRemoveInputMapping(key)}
                className="text-red-400 hover:text-red-300 p-1"
                title="Remove mapping"
              >
                ✕
              </button>
            </div>
          ))}
        </div>

        {/* Add new mapping */}
        <div className="flex gap-2">
          <input
            type="text"
            value={newInputKey}
            onChange={(e) => setNewInputKey(e.target.value)}
            placeholder="Field name"
            className="flex-1 px-3 py-2 bg-gray-700 border border-gray-600 rounded-lg text-white placeholder-gray-400 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
          />
          <input
            type="text"
            value={newInputValue}
            onChange={(e) => setNewInputValue(e.target.value)}
            placeholder="${trigger.data}"
            className="flex-1 px-3 py-2 bg-gray-700 border border-gray-600 rounded-lg text-white placeholder-gray-400 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
          />
          <button
            type="button"
            onClick={handleAddInputMapping}
            disabled={!newInputKey || !newInputValue}
            className="px-3 py-2 bg-blue-600 hover:bg-blue-700 disabled:bg-gray-600 disabled:cursor-not-allowed text-white rounded-lg text-sm font-medium transition-colors"
          >
            Add
          </button>
        </div>

        <p className="text-xs text-gray-400">
          Map parent workflow data to subworkflow input. Use expressions like ${'{'}trigger.data{'}'} or ${'{'}steps.stepId.output{'}'}
        </p>
      </div>

      {/* Output Mapping */}
      <div className="space-y-3">
        <div className="flex items-center justify-between">
          <label className="block text-sm font-medium text-gray-300">Output Mapping</label>
          <span className="text-xs text-gray-500">
            {Object.keys(outputMapping).length} mapping{Object.keys(outputMapping).length !== 1 ? 's' : ''}
          </span>
        </div>

        {/* Existing mappings */}
        <div className="space-y-2">
          {Object.entries(outputMapping).map(([key, value]) => (
            <div key={key} className="flex items-center gap-2 p-2 bg-gray-700/50 rounded border border-gray-600">
              <div className="flex-1 grid grid-cols-2 gap-2">
                <span className="text-sm text-white font-mono truncate">{key}</span>
                <span className="text-sm text-gray-400 font-mono truncate">{value}</span>
              </div>
              <button
                type="button"
                onClick={() => handleRemoveOutputMapping(key)}
                className="text-red-400 hover:text-red-300 p-1"
                title="Remove mapping"
              >
                ✕
              </button>
            </div>
          ))}
        </div>

        {/* Add new mapping */}
        <div className="flex gap-2">
          <input
            type="text"
            value={newOutputKey}
            onChange={(e) => setNewOutputKey(e.target.value)}
            placeholder="Variable name"
            className="flex-1 px-3 py-2 bg-gray-700 border border-gray-600 rounded-lg text-white placeholder-gray-400 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
          />
          <input
            type="text"
            value={newOutputValue}
            onChange={(e) => setNewOutputValue(e.target.value)}
            placeholder="${output.result}"
            className="flex-1 px-3 py-2 bg-gray-700 border border-gray-600 rounded-lg text-white placeholder-gray-400 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
          />
          <button
            type="button"
            onClick={handleAddOutputMapping}
            disabled={!newOutputKey || !newOutputValue}
            className="px-3 py-2 bg-blue-600 hover:bg-blue-700 disabled:bg-gray-600 disabled:cursor-not-allowed text-white rounded-lg text-sm font-medium transition-colors"
          >
            Add
          </button>
        </div>

        <p className="text-xs text-gray-400">
          Map subworkflow output to parent workflow variables. The mapped values can be used in subsequent steps.
        </p>
      </div>
    </div>
  )
}
