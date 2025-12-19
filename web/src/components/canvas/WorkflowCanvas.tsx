import { useCallback, useState, useEffect } from 'react'
import {
  ReactFlow,
  Background,
  Controls,
  MiniMap,
  useNodesState,
  useEdgesState,
  addEdge,
  type Node,
  type Edge,
  type Connection,
} from '@xyflow/react'
import '@xyflow/react/dist/style.css'
import { nodeTypes } from '../nodes/nodeTypes'

interface WorkflowCanvasProps {
  initialNodes?: Node[]
  initialEdges?: Edge[]
  onSave?: (workflow: { nodes: Node[]; edges: Edge[] }) => void
  onChange?: (workflow: { nodes: Node[]; edges: Edge[] }) => void
  onNodeSelect?: (node: Node | null) => void
}

export default function WorkflowCanvas({
  initialNodes = [],
  initialEdges = [],
  onSave,
  onChange,
  onNodeSelect,
}: WorkflowCanvasProps) {
  const [nodes, setNodes, onNodesChange] = useNodesState(initialNodes)
  const [edges, setEdges, onEdgesChange] = useEdgesState(initialEdges)
  const [validationError, setValidationError] = useState<string | null>(null)

  // Notify parent of changes
  useEffect(() => {
    if (onChange) {
      onChange({ nodes, edges })
    }
  }, [nodes, edges, onChange])

  const onConnect = useCallback(
    (connection: Connection) => {
      setEdges((eds) => addEdge(connection, eds))
    },
    [setEdges]
  )

  const handleAddNode = useCallback(() => {
    const newNode: Node = {
      id: `node-${Date.now()}`,
      type: 'action',
      position: { x: Math.random() * 400, y: Math.random() * 400 },
      data: { label: 'New Node' },
    }
    setNodes((nds) => [...nds, newNode])
  }, [setNodes])

  const validateWorkflow = (): string | null => {
    // Check if workflow has at least one node
    if (nodes.length === 0) {
      return 'Workflow must have at least one node'
    }

    // Check if workflow has a trigger node
    const hasTrigger = nodes.some((node) => node.type === 'trigger')
    if (!hasTrigger) {
      return 'Workflow must have a trigger node'
    }

    return null
  }

  const handleSave = useCallback(() => {
    setValidationError(null)

    const error = validateWorkflow()
    if (error) {
      setValidationError(error)
      return
    }

    if (onSave) {
      onSave({ nodes, edges })
    }
  }, [nodes, edges, onSave])

  return (
    <div className="w-full h-full flex flex-col">
      {/* Toolbar */}
      <div className="bg-gray-800 border-b border-gray-700 p-4 flex items-center justify-between">
        <div className="flex items-center space-x-2">
          <button
            onClick={handleAddNode}
            className="px-3 py-2 bg-primary-600 text-white rounded-lg text-sm font-medium hover:bg-primary-700 transition-colors"
          >
            Add Node
          </button>
        </div>

        <div className="flex items-center space-x-2">
          <button
            onClick={handleSave}
            className="px-4 py-2 bg-green-600 text-white rounded-lg text-sm font-medium hover:bg-green-700 transition-colors"
          >
            Save
          </button>
        </div>
      </div>

      {/* Validation Error */}
      {validationError && (
        <div className="bg-red-900/20 border border-red-500/30 text-red-400 px-4 py-3 text-sm">
          {validationError}
        </div>
      )}

      {/* Canvas */}
      <div className="flex-1">
        <ReactFlow
          nodes={nodes}
          edges={edges}
          onNodesChange={onNodesChange}
          onEdgesChange={onEdgesChange}
          onConnect={onConnect}
          onNodeClick={(_, node) => onNodeSelect?.(node)}
          onPaneClick={() => onNodeSelect?.(null)}
          nodeTypes={nodeTypes}
          fitView
        >
          <Background />
          <Controls />
          <MiniMap />
        </ReactFlow>
      </div>
    </div>
  )
}
