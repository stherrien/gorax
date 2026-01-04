import { FC, useState, useCallback } from 'react'
import { useNavigate } from 'react-router-dom'
import { ChatPanel } from '../components/aibuilder/ChatPanel'
import { WorkflowPreview } from '../components/aibuilder/WorkflowPreview'
import type { GeneratedWorkflow, GeneratedNode } from '../types/aibuilder'

export const AIWorkflowBuilder: FC = () => {
  const navigate = useNavigate()
  const [previewWorkflow, setPreviewWorkflow] = useState<GeneratedWorkflow | null>(null)
  const [selectedNode, setSelectedNode] = useState<GeneratedNode | null>(null)

  const handleWorkflowGenerated = useCallback((workflow: GeneratedWorkflow) => {
    setPreviewWorkflow(workflow)
    setSelectedNode(null)
  }, [])

  const handleApply = useCallback((workflowId: string) => {
    // Navigate to the workflow editor with the new workflow
    navigate(`/workflows/${workflowId}`)
  }, [navigate])

  const handleNodeClick = useCallback((node: GeneratedNode) => {
    setSelectedNode(node)
  }, [])

  const handleOpenEditor = useCallback(() => {
    // For now, just show a message - in production this would open
    // a modal or navigate to an editor with the workflow pre-loaded
    if (previewWorkflow) {
      // TODO: Implement workflow editor navigation with preview workflow
      // This would typically create a new workflow from the preview and navigate to it
      // For now, we just show an alert as this is a placeholder
      alert('Opening in editor with preview workflow (not yet implemented)')
    }
  }, [previewWorkflow])

  return (
    <div className="flex h-screen bg-gray-50">
      {/* Left Panel - Chat */}
      <div className="flex w-1/2 flex-col border-r border-gray-200">
        <ChatPanel
          onWorkflowGenerated={handleWorkflowGenerated}
          onApply={handleApply}
        />
      </div>

      {/* Right Panel - Preview */}
      <div className="flex w-1/2 flex-col">
        {/* Preview Header */}
        <div className="border-b border-gray-200 bg-white px-4 py-3">
          <h2 className="text-lg font-semibold text-gray-900">
            Workflow Preview
          </h2>
          <p className="text-sm text-gray-500">
            Visual preview of generated workflow
          </p>
        </div>

        {/* Preview Content */}
        <div className="flex-1 overflow-auto p-4">
          {previewWorkflow ? (
            <div className="space-y-4">
              <WorkflowPreview
                workflow={previewWorkflow}
                onNodeClick={handleNodeClick}
                onOpenEditor={handleOpenEditor}
              />

              {/* Node Details Panel */}
              {selectedNode && (
                <NodeDetailsPanel
                  node={selectedNode}
                  onClose={() => setSelectedNode(null)}
                />
              )}
            </div>
          ) : (
            <EmptyPreview />
          )}
        </div>
      </div>
    </div>
  )
}

// Empty state component
const EmptyPreview: FC = () => (
  <div className="flex h-full flex-col items-center justify-center text-center">
    <div className="mb-4 text-6xl text-gray-300">📊</div>
    <h3 className="mb-2 text-lg font-medium text-gray-600">
      No Workflow Yet
    </h3>
    <p className="max-w-md text-sm text-gray-400">
      Start a conversation in the chat panel to generate a workflow.
      The preview will appear here as you build.
    </p>
  </div>
)

// Node details panel component
interface NodeDetailsPanelProps {
  node: GeneratedNode
  onClose: () => void
}

const NodeDetailsPanel: FC<NodeDetailsPanelProps> = ({ node, onClose }) => (
  <div className="rounded-lg border border-gray-200 bg-white p-4">
    <div className="mb-3 flex items-center justify-between">
      <h4 className="font-medium text-gray-900">Node Details</h4>
      <button
        onClick={onClose}
        className="text-gray-400 hover:text-gray-600"
      >
        &times;
      </button>
    </div>

    <div className="space-y-3">
      <DetailRow label="Name" value={node.name} />
      <DetailRow label="Type" value={node.type} />
      <DetailRow label="ID" value={node.id} mono />
      {node.description && (
        <DetailRow label="Description" value={node.description} />
      )}
      {node.config && Object.keys(node.config).length > 0 && (
        <div>
          <span className="text-sm font-medium text-gray-500">Configuration</span>
          <pre className="mt-1 overflow-auto rounded bg-gray-50 p-2 text-xs text-gray-700">
            {JSON.stringify(node.config, null, 2)}
          </pre>
        </div>
      )}
    </div>
  </div>
)

// Detail row helper component
interface DetailRowProps {
  label: string
  value: string
  mono?: boolean
}

const DetailRow: FC<DetailRowProps> = ({ label, value, mono }) => (
  <div>
    <span className="text-sm font-medium text-gray-500">{label}</span>
    <p className={`text-sm text-gray-900 ${mono ? 'font-mono' : ''}`}>
      {value}
    </p>
  </div>
)

export default AIWorkflowBuilder
