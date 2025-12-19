/**
 * ExecutionCanvas - Visualizes workflow execution on canvas with real-time updates
 *
 * Integrates:
 * - WorkflowCanvas (read-only view)
 * - ExecutionDetailsPanel (Timeline + LogViewer)
 * - WebSocket real-time updates via useExecutionTrace
 * - Node selection for detailed log viewing
 */

import { useState, useEffect } from 'react'
import {
  ReactFlow,
  Background,
  Controls,
  MiniMap,
  useNodesState,
  useEdgesState,
  type Node,
  type Edge,
} from '@xyflow/react'
import '@xyflow/react/dist/style.css'
import { useWorkflow } from '../../hooks/useWorkflows'
import { useExecutionTrace } from '../../hooks/useExecutionTrace'
import { useExecutionTraceStore } from '../../stores/executionTraceStore'
import { ExecutionDetailsPanel } from './ExecutionDetailsPanel'
import { nodeTypes } from '../nodes/nodeTypes'

export interface ExecutionCanvasProps {
  workflowId: string
  executionId: string
  className?: string
}

/**
 * ExecutionCanvas Component
 * Displays workflow execution with real-time status updates
 */
export function ExecutionCanvas({
  workflowId,
  executionId,
  className = '',
}: ExecutionCanvasProps) {
  // Load workflow definition
  const { workflow, loading, error } = useWorkflow(workflowId)

  // Connect to WebSocket for real-time updates
  const { connected, reconnecting, reconnectAttempt } = useExecutionTrace(executionId, {
    enabled: true,
  })

  // Selected node for log viewing
  const [selectedNodeId, setSelectedNodeId] = useState<string | null>(null)

  // ReactFlow state
  const [nodes, , onNodesChange] = useNodesState<Node>(workflow?.definition?.nodes || [])
  const [edges, , onEdgesChange] = useEdgesState<Edge>(workflow?.definition?.edges || [])

  // Set execution ID in store on mount
  useEffect(() => {
    if (executionId) {
      useExecutionTraceStore.getState().setCurrentExecutionId(executionId)
    }

    // Reset store on unmount
    return () => {
      useExecutionTraceStore.getState().reset()
    }
  }, [executionId])

  // Update nodes/edges when workflow loads
  useEffect(() => {
    if (workflow?.definition) {
      // We don't need to manually update - useNodesState/useEdgesState handle this
    }
  }, [workflow])

  // Handle node click
  const handleNodeClick = (_: any, node: Node) => {
    setSelectedNodeId(node.id)
  }

  // Handle pane click (deselect)
  const handlePaneClick = () => {
    setSelectedNodeId(null)
  }

  // Loading state
  if (loading) {
    return (
      <div
        className="flex items-center justify-center h-full bg-gray-900"
        data-testid="execution-canvas-loading"
      >
        <div className="text-center">
          <div className="text-white text-xl mb-2">Loading workflow...</div>
          <div className="text-gray-400 text-sm">Preparing execution canvas</div>
        </div>
      </div>
    )
  }

  // Error state
  if (error) {
    return (
      <div
        className="flex items-center justify-center h-full bg-gray-900"
        data-testid="execution-canvas-error"
      >
        <div className="text-center">
          <div className="text-red-400 text-xl mb-2">Failed to load workflow</div>
          <div className="text-gray-400 text-sm">{error.message}</div>
        </div>
      </div>
    )
  }

  if (!workflow) {
    return (
      <div
        className="flex items-center justify-center h-full bg-gray-900"
        data-testid="execution-canvas-error"
      >
        <div className="text-gray-400 text-lg">Workflow not found</div>
      </div>
    )
  }

  return (
    <div
      className={`execution-canvas-container h-full flex ${className}`}
      data-testid="execution-canvas"
      role="region"
      aria-label="Execution workflow canvas"
    >
      {/* Canvas Section (Left) */}
      <div className="flex-1 relative flex flex-col">
        {/* Connection Status Header */}
        <div className="bg-gray-800 border-b border-gray-700 px-4 py-2 flex items-center justify-between">
          <div>
            <h2 className="text-white text-lg font-semibold">{workflow.name}</h2>
            <p className="text-gray-400 text-sm">
              Execution: {executionId ? `${executionId.substring(0, 8)}...` : 'N/A'}
            </p>
          </div>
          <ConnectionStatus
            connected={connected}
            reconnecting={reconnecting}
            reconnectAttempt={reconnectAttempt}
          />
        </div>

        {/* Canvas */}
        <div className="flex-1">
          <ReactFlow
            nodes={nodes}
            edges={edges}
            onNodesChange={onNodesChange}
            onEdgesChange={onEdgesChange}
            onNodeClick={handleNodeClick}
            onPaneClick={handlePaneClick}
            nodeTypes={nodeTypes}
            nodesDraggable={false}
            nodesConnectable={false}
            elementsSelectable={true}
            fitView
          >
            <Background />
            <Controls />
            <MiniMap />
          </ReactFlow>
        </div>
      </div>

      {/* Details Panel (Right) */}
      <div className="w-96 border-l border-gray-700 bg-gray-900 overflow-hidden">
        <ExecutionDetailsPanel selectedNodeId={selectedNodeId} />
      </div>
    </div>
  )
}

/**
 * ConnectionStatus Component
 * Shows WebSocket connection status
 */
interface ConnectionStatusProps {
  connected: boolean
  reconnecting: boolean
  reconnectAttempt: number
}

function ConnectionStatus({ connected, reconnecting, reconnectAttempt }: ConnectionStatusProps) {
  if (connected) {
    return (
      <div
        className="flex items-center space-x-2 text-green-400"
        data-testid="connection-status-connected"
        role="status"
        aria-live="polite"
      >
        <div className="w-2 h-2 bg-green-400 rounded-full animate-pulse" />
        <span className="text-sm">Connected</span>
      </div>
    )
  }

  if (reconnecting) {
    return (
      <div
        className="flex items-center space-x-2 text-yellow-400"
        data-testid="connection-status-reconnecting"
        role="status"
        aria-live="polite"
      >
        <div className="w-2 h-2 bg-yellow-400 rounded-full animate-pulse" />
        <span className="text-sm">Reconnecting... (Attempt {reconnectAttempt})</span>
      </div>
    )
  }

  return (
    <div
      className="flex items-center space-x-2 text-red-400"
      data-testid="connection-status-disconnected"
      role="status"
      aria-live="polite"
    >
      <div className="w-2 h-2 bg-red-400 rounded-full" />
      <span className="text-sm">Disconnected</span>
    </div>
  )
}
