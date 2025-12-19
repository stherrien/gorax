import { useWorkflowStore, WorkflowNode } from '../../stores/workflowStore'

interface PropertyPanelProps {
  nodeId: string
}

export default function PropertyPanel({ nodeId }: PropertyPanelProps) {
  const { nodes, updateNode, deleteNode, selectNode } = useWorkflowStore()

  const node = nodes.find((n) => n.id === nodeId) as WorkflowNode | undefined

  if (!node) return null

  const handleLabelChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    updateNode(nodeId, { label: e.target.value })
  }

  const handleConfigChange = (key: string, value: string) => {
    const currentConfig = (node.data.config as Record<string, unknown>) || {}
    updateNode(nodeId, {
      config: { ...currentConfig, [key]: value },
    })
  }

  const handleDelete = () => {
    deleteNode(nodeId)
  }

  const handleClose = () => {
    selectNode(null)
  }

  return (
    <div className="w-80 bg-gray-800 border-l border-gray-700 overflow-y-auto">
      <div className="p-4 border-b border-gray-700 flex items-center justify-between">
        <h3 className="text-white font-semibold">Properties</h3>
        <button
          onClick={handleClose}
          className="text-gray-400 hover:text-white transition-colors"
        >
          ✕
        </button>
      </div>

      <div className="p-4 space-y-6">
        {/* Basic Info */}
        <div>
          <label className="block text-gray-400 text-sm mb-2">Label</label>
          <input
            type="text"
            value={(node.data.label as string) || ''}
            onChange={handleLabelChange}
            className="w-full px-3 py-2 bg-gray-700 border border-gray-600 rounded-lg text-white text-sm focus:outline-none focus:border-primary-500"
          />
        </div>

        <div>
          <label className="block text-gray-400 text-sm mb-2">Type</label>
          <p className="text-white text-sm">{node.type}</p>
        </div>

        {/* Type-specific config */}
        {node.type === 'trigger' && (
          <TriggerConfig node={node} onConfigChange={handleConfigChange} />
        )}

        {node.type === 'action' && (
          <ActionConfig node={node} onConfigChange={handleConfigChange} />
        )}

        {/* Delete button */}
        <div className="pt-4 border-t border-gray-700">
          <button
            onClick={handleDelete}
            className="w-full px-4 py-2 bg-red-600/20 text-red-400 rounded-lg text-sm font-medium hover:bg-red-600/30 transition-colors"
          >
            Delete Node
          </button>
        </div>
      </div>
    </div>
  )
}

interface ConfigProps {
  node: WorkflowNode
  onConfigChange: (key: string, value: string) => void
}

function TriggerConfig({ node, onConfigChange }: ConfigProps) {
  const config = (node.data.config as Record<string, string>) || {}
  const triggerType = (node.data as { triggerType?: string }).triggerType

  if (triggerType === 'webhook') {
    return (
      <div className="space-y-4">
        <h4 className="text-gray-400 text-xs uppercase tracking-wider">
          Webhook Configuration
        </h4>
        <div>
          <label className="block text-gray-400 text-sm mb-2">Auth Type</label>
          <select
            value={config.auth_type || 'none'}
            onChange={(e) => onConfigChange('auth_type', e.target.value)}
            className="w-full px-3 py-2 bg-gray-700 border border-gray-600 rounded-lg text-white text-sm focus:outline-none focus:border-primary-500"
          >
            <option value="none">None</option>
            <option value="basic">Basic Auth</option>
            <option value="signature">Signature</option>
            <option value="api_key">API Key</option>
          </select>
        </div>
      </div>
    )
  }

  if (triggerType === 'schedule') {
    return (
      <div className="space-y-4">
        <h4 className="text-gray-400 text-xs uppercase tracking-wider">
          Schedule Configuration
        </h4>
        <div>
          <label className="block text-gray-400 text-sm mb-2">Cron Expression</label>
          <input
            type="text"
            value={config.cron || ''}
            onChange={(e) => onConfigChange('cron', e.target.value)}
            placeholder="0 * * * *"
            className="w-full px-3 py-2 bg-gray-700 border border-gray-600 rounded-lg text-white text-sm focus:outline-none focus:border-primary-500"
          />
        </div>
        <div>
          <label className="block text-gray-400 text-sm mb-2">Timezone</label>
          <input
            type="text"
            value={config.timezone || 'UTC'}
            onChange={(e) => onConfigChange('timezone', e.target.value)}
            className="w-full px-3 py-2 bg-gray-700 border border-gray-600 rounded-lg text-white text-sm focus:outline-none focus:border-primary-500"
          />
        </div>
      </div>
    )
  }

  return null
}

function ActionConfig({ node, onConfigChange }: ConfigProps) {
  const config = (node.data.config as Record<string, string>) || {}
  const actionType = (node.data as { actionType?: string }).actionType

  if (actionType === 'http') {
    return (
      <div className="space-y-4">
        <h4 className="text-gray-400 text-xs uppercase tracking-wider">
          HTTP Request Configuration
        </h4>
        <div>
          <label className="block text-gray-400 text-sm mb-2">Method</label>
          <select
            value={config.method || 'GET'}
            onChange={(e) => onConfigChange('method', e.target.value)}
            className="w-full px-3 py-2 bg-gray-700 border border-gray-600 rounded-lg text-white text-sm focus:outline-none focus:border-primary-500"
          >
            <option value="GET">GET</option>
            <option value="POST">POST</option>
            <option value="PUT">PUT</option>
            <option value="PATCH">PATCH</option>
            <option value="DELETE">DELETE</option>
          </select>
        </div>
        <div>
          <label className="block text-gray-400 text-sm mb-2">URL</label>
          <input
            type="text"
            value={config.url || ''}
            onChange={(e) => onConfigChange('url', e.target.value)}
            placeholder="https://api.example.com/endpoint"
            className="w-full px-3 py-2 bg-gray-700 border border-gray-600 rounded-lg text-white text-sm focus:outline-none focus:border-primary-500"
          />
        </div>
        <div>
          <label className="block text-gray-400 text-sm mb-2">Timeout (seconds)</label>
          <input
            type="number"
            value={config.timeout || '30'}
            onChange={(e) => onConfigChange('timeout', e.target.value)}
            className="w-full px-3 py-2 bg-gray-700 border border-gray-600 rounded-lg text-white text-sm focus:outline-none focus:border-primary-500"
          />
        </div>
      </div>
    )
  }

  if (actionType === 'transform') {
    return (
      <div className="space-y-4">
        <h4 className="text-gray-400 text-xs uppercase tracking-wider">
          Transform Configuration
        </h4>
        <div>
          <label className="block text-gray-400 text-sm mb-2">Expression</label>
          <textarea
            value={config.expression || ''}
            onChange={(e) => onConfigChange('expression', e.target.value)}
            placeholder="steps.previous_step.body.data"
            rows={3}
            className="w-full px-3 py-2 bg-gray-700 border border-gray-600 rounded-lg text-white text-sm focus:outline-none focus:border-primary-500 font-mono"
          />
        </div>
      </div>
    )
  }

  return (
    <div className="text-gray-400 text-sm">
      Configuration for {actionType} not yet implemented.
    </div>
  )
}
