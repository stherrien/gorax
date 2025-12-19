import { useCallback, useState, useEffect, useRef } from 'react'
import {
  ReactFlow,
  Background,
  Controls,
  MiniMap,
  ReactFlowProvider,
  useNodesState,
  useEdgesState,
  addEdge,
  useReactFlow,
  type Node,
  type Edge,
  type Connection,
} from '@xyflow/react'
import '@xyflow/react/dist/style.css'
import { nodeTypes } from '../nodes/nodeTypes'
import { detectCycles, isValidDAG } from '../../utils/dagValidation'

interface WorkflowCanvasProps {
  initialNodes?: Node[]
  initialEdges?: Edge[]
  onSave?: (workflow: { nodes: Node[]; edges: Edge[] }) => void
  onChange?: (workflow: { nodes: Node[]; edges: Edge[] }) => void
  onNodeSelect?: (node: Node | null) => void
}

function WorkflowCanvasInner({
  initialNodes = [],
  initialEdges = [],
  onSave,
  onChange,
  onNodeSelect,
}: WorkflowCanvasProps) {
  const [nodes, setNodes, onNodesChange] = useNodesState(initialNodes)
  const [edges, setEdges, onEdgesChange] = useEdgesState(initialEdges)
  const [validationError, setValidationError] = useState<string | null>(null)
  const [cycleError, setCycleError] = useState<string | null>(null)
  const [nodesInCycle, setNodesInCycle] = useState<Set<string>>(new Set())
  const reactFlowWrapper = useRef<HTMLDivElement>(null)
  const { screenToFlowPosition } = useReactFlow()

  // Notify parent of changes
  useEffect(() => {
    if (onChange) {
      onChange({ nodes, edges })
    }
  }, [nodes, edges, onChange])

  // Apply visual styling to nodes in cycle
  useEffect(() => {
    if (nodesInCycle.size === 0) return

    setNodes((nds) =>
      nds.map((node) => ({
        ...node,
        className: nodesInCycle.has(node.id) ? 'cycle-error' : node.className || '',
        style: {
          ...node.style,
          ...(nodesInCycle.has(node.id) && {
            border: '2px solid #ef4444',
            boxShadow: '0 0 10px rgba(239, 68, 68, 0.5)',
          }),
        },
      }))
    )
  }, [nodesInCycle, setNodes])

  const onConnect = useCallback(
    (connection: Connection) => {
      // Create hypothetical edge to test for cycles
      const newEdge: Edge = {
        id: `e${connection.source}-${connection.target}`,
        source: connection.source!,
        target: connection.target!,
      }
      const hypotheticalEdges = [...edges, newEdge]

      // Check if this connection would create a cycle
      const cycles = detectCycles(nodes, hypotheticalEdges)

      if (cycles.length > 0) {
        // Show error and highlight nodes in cycle
        const cycleNodes = new Set(cycles[0].filter((id) => id !== cycles[0][cycles[0].length - 1]))
        setNodesInCycle(cycleNodes)
        setCycleError(`Cannot add connection: would create a cycle (${cycles[0].join(' → ')})`)

        // Clear error after 5 seconds
        setTimeout(() => {
          setCycleError(null)
          setNodesInCycle(new Set())
        }, 5000)

        return
      }

      // Connection is valid, add it
      setCycleError(null)
      setNodesInCycle(new Set())
      setEdges((eds) => addEdge(connection, eds))
    },
    [setEdges, nodes, edges]
  )

  const onDragOver = useCallback((event: React.DragEvent) => {
    event.preventDefault()
    event.dataTransfer.dropEffect = 'move'
  }, [])

  const onDrop = useCallback(
    (event: React.DragEvent) => {
      event.preventDefault()

      const data = event.dataTransfer.getData('application/reactflow')
      if (!data) return

      const nodeData = JSON.parse(data)
      const position = screenToFlowPosition({
        x: event.clientX,
        y: event.clientY,
      })

      const newNode: Node = {
        id: `${nodeData.nodeType}-${Date.now()}`,
        type: nodeData.type,
        position,
        data: {
          label: nodeData.label,
          nodeType: nodeData.nodeType,
        },
      }

      setNodes((nds) => [...nds, newNode])
    },
    [screenToFlowPosition, setNodes]
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

    // Check for cycles (DAG validation)
    if (!isValidDAG(nodes, edges)) {
      const cycles = detectCycles(nodes, edges)
      return `Workflow contains a cycle: ${cycles[0].join(' → ')}`
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

      {/* Cycle Error */}
      {cycleError && (
        <div className="bg-yellow-900/20 border border-yellow-500/30 text-yellow-400 px-4 py-3 text-sm">
          {cycleError}
        </div>
      )}

      {/* Canvas */}
      <div className="flex-1" ref={reactFlowWrapper}>
        <ReactFlow
          nodes={nodes}
          edges={edges}
          onNodesChange={onNodesChange}
          onEdgesChange={onEdgesChange}
          onConnect={onConnect}
          onNodeClick={(_, node) => onNodeSelect?.(node)}
          onPaneClick={() => onNodeSelect?.(null)}
          onDrop={onDrop}
          onDragOver={onDragOver}
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

export default function WorkflowCanvas(props: WorkflowCanvasProps) {
  return (
    <ReactFlowProvider>
      <WorkflowCanvasInner {...props} />
    </ReactFlowProvider>
  )
}
