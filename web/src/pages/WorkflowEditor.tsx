import { useState, useEffect } from 'react'
import { useParams, useNavigate, Link } from 'react-router-dom'
import type { Node, Edge } from '@xyflow/react'
import WorkflowCanvas from '../components/canvas/WorkflowCanvas'
import NodePalette from '../components/canvas/NodePalette'
import PropertyPanel from '../components/canvas/PropertyPanel'
import { useWorkflow, useWorkflowMutations } from '../hooks/useWorkflows'

export default function WorkflowEditor() {
  const { id } = useParams()
  const navigate = useNavigate()
  const isNewWorkflow = id === 'new'

  // Load existing workflow if editing
  const { workflow, loading, error } = useWorkflow(isNewWorkflow ? null : id || null)
  const { createWorkflow, updateWorkflow, creating, updating } = useWorkflowMutations()

  // Form state
  const [name, setName] = useState('')
  const [description, setDescription] = useState('')
  const [nodes, setNodes] = useState<Node[]>([])
  const [edges, setEdges] = useState<Edge[]>([])
  const [selectedNode, setSelectedNode] = useState<Node | null>(null)
  const [validationError, setValidationError] = useState<string | null>(null)

  // Load workflow data when editing
  useEffect(() => {
    if (workflow && !isNewWorkflow) {
      setName(workflow.name)
      setDescription(workflow.description || '')
      setNodes(workflow.definition?.nodes || [])
      setEdges(workflow.definition?.edges || [])
    }
  }, [workflow, isNewWorkflow])

  const handleAddNode = (nodeData: { type: string; nodeType: string }) => {
    const newNode: Node = {
      id: `node-${Date.now()}`,
      type: nodeData.type,
      position: { x: 250, y: 100 },
      data: {
        nodeType: nodeData.nodeType,
        label: `New ${nodeData.nodeType}`,
      },
    }
    setNodes((prev) => [...prev, newNode])
  }

  const handleCanvasChange = (workflow: { nodes: Node[]; edges: Edge[] }) => {
    setNodes(workflow.nodes)
    setEdges(workflow.edges)
  }

  const handleNodeSelect = (node: Node | null) => {
    setSelectedNode(node)
  }

  const handleNodeUpdate = (nodeId: string, data: any) => {
    setNodes((prev) =>
      prev.map((node) => (node.id === nodeId ? { ...node, data: { ...node.data, ...data } } : node))
    )
  }

  const handleSave = async () => {
    // Validate
    if (!name || name.trim() === '') {
      setValidationError('Workflow name is required')
      return
    }

    setValidationError(null)

    const workflowData = {
      name,
      description,
      definition: {
        nodes: nodes.map((node) => ({
          id: node.id,
          type: node.type || 'default',
          position: node.position,
          data: node.data || {},
        })),
        edges: edges.map((edge) => ({
          id: edge.id,
          source: edge.source,
          target: edge.target,
          sourceHandle: edge.sourceHandle || undefined,
          targetHandle: edge.targetHandle || undefined,
        })),
      },
    }

    try {
      if (isNewWorkflow) {
        const newWorkflow = await createWorkflow(workflowData)
        navigate(`/workflows/${newWorkflow.id}`)
      } else {
        await updateWorkflow(id!, workflowData)
        // Stay on the same page after update
      }
    } catch (err) {
      console.error('Failed to save workflow:', err)
    }
  }

  // Loading state
  if (loading && !isNewWorkflow) {
    return (
      <div className="flex items-center justify-center h-screen bg-gray-900">
        <div className="text-white text-xl">Loading workflow...</div>
      </div>
    )
  }

  // Error state
  if (error && !isNewWorkflow) {
    return (
      <div className="flex items-center justify-center h-screen bg-gray-900">
        <div className="text-center">
          <div className="text-red-400 text-xl mb-4">Failed to load workflow</div>
          <div className="text-gray-400">{error.message}</div>
          <Link
            to="/workflows"
            className="mt-4 inline-block px-4 py-2 bg-primary-600 text-white rounded-lg hover:bg-primary-700"
          >
            Back to Workflows
          </Link>
        </div>
      </div>
    )
  }

  return (
    <div className="h-screen bg-gray-900 flex flex-col">
      {/* Header */}
      <div className="bg-gray-800 border-b border-gray-700 px-6 py-4">
        <div className="flex items-center justify-between mb-4">
          <div className="flex items-center space-x-4">
            <Link
              to="/workflows"
              className="text-gray-400 hover:text-white transition-colors"
              aria-label="Back to workflows"
            >
              ← Back to Workflows
            </Link>
            <h1 className="text-2xl font-bold text-white">
              {isNewWorkflow ? 'New Workflow' : 'Edit Workflow'}
            </h1>
          </div>
          <button
            onClick={handleSave}
            disabled={creating || updating}
            className="px-6 py-2 bg-primary-600 text-white rounded-lg font-medium hover:bg-primary-700 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
          >
            {creating || updating ? 'Saving...' : 'Save Workflow'}
          </button>
        </div>

        {/* Workflow Metadata */}
        <div className="grid grid-cols-2 gap-4">
          <div>
            <label htmlFor="workflow-name" className="block text-sm font-medium text-gray-300 mb-2">
              Workflow Name *
            </label>
            <input
              id="workflow-name"
              type="text"
              value={name}
              onChange={(e) => {
                setName(e.target.value)
                if (validationError) setValidationError(null)
              }}
              placeholder="Enter workflow name"
              className="w-full px-4 py-2 bg-gray-700 text-white rounded-lg focus:outline-none focus:ring-2 focus:ring-primary-500"
            />
            {validationError && <div className="mt-1 text-sm text-red-400">{validationError}</div>}
          </div>
          <div>
            <label htmlFor="workflow-description" className="block text-sm font-medium text-gray-300 mb-2">
              Description
            </label>
            <textarea
              id="workflow-description"
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              placeholder="Enter workflow description"
              rows={1}
              className="w-full px-4 py-2 bg-gray-700 text-white rounded-lg focus:outline-none focus:ring-2 focus:ring-primary-500 resize-none"
            />
          </div>
        </div>
      </div>

      {/* Editor Layout */}
      <div className="flex-1 flex overflow-hidden">
        {/* Node Palette */}
        <NodePalette onAddNode={handleAddNode} />

        {/* Canvas */}
        <div className="flex-1 relative">
          <WorkflowCanvas
            initialNodes={nodes}
            initialEdges={edges}
            onChange={handleCanvasChange}
            onNodeSelect={handleNodeSelect}
          />
        </div>

        {/* Property Panel */}
        <PropertyPanel
          node={selectedNode}
          onUpdate={handleNodeUpdate}
          onClose={() => setSelectedNode(null)}
        />
      </div>
    </div>
  )
}
