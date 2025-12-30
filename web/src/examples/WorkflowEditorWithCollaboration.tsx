/**
 * Example integration of collaboration features into WorkflowEditor
 *
 * This file demonstrates how to integrate the collaboration hooks and components
 * into the existing WorkflowEditor component.
 */

import { useState, useCallback } from 'react'
import { useParams } from 'react-router-dom'
import type { Node, Edge } from '@xyflow/react'
import WorkflowCanvas from '../components/canvas/WorkflowCanvas'
import NodePalette from '../components/canvas/NodePalette'
import PropertyPanel from '../components/canvas/PropertyPanel'
import { useCollaboration } from '../hooks/useCollaboration'
import {
  UserPresenceIndicator,
  CollaboratorList,
  CollaboratorCursors,
  NodeLockIndicator,
} from '../components/collaboration'
import type { EditOperation } from '../types/collaboration'

// Mock user - replace with actual auth context
const getCurrentUser = () => ({
  id: 'user-123',
  name: 'John Doe',
  email: 'john@example.com',
})

export default function WorkflowEditorWithCollaboration() {
  const { id: workflowId } = useParams()
  const currentUser = getCurrentUser()

  const [nodes, setNodes] = useState<Node[]>([])
  const [edges, setEdges] = useState<Edge[]>([])
  const [selectedNode, setSelectedNode] = useState<Node | null>(null)

  // Setup collaboration
  const collaboration = useCollaboration(
    workflowId || null,
    currentUser.id,
    currentUser.name,
    {
      enabled: !!workflowId,
      onUserJoined: (user) => {
        console.log('User joined:', user.user_name)
        // Could show toast notification
      },
      onUserLeft: (userId) => {
        console.log('User left:', userId)
      },
      onLockAcquired: (lock) => {
        console.log('Lock acquired:', lock.element_id, 'by', lock.user_name)
        // Visual feedback that element is locked
      },
      onLockReleased: (elementId) => {
        console.log('Lock released:', elementId)
      },
      onLockFailed: (elementId, reason, currentLock) => {
        console.warn('Lock failed:', elementId, reason)
        // Show error message
        if (currentLock) {
          alert(`Node is being edited by ${currentLock.user_name}`)
        }
      },
      onChangeApplied: (operation) => {
        console.log('Remote change applied:', operation.type, operation.element_id)
        // Apply remote changes to local state
        applyRemoteOperation(operation)
      },
      onError: (error) => {
        console.error('Collaboration error:', error)
      },
    }
  )

  // Apply remote operations to local state
  const applyRemoteOperation = useCallback((operation: EditOperation) => {
    switch (operation.type) {
      case 'node_add':
        setNodes((prev) => [...prev, operation.data])
        break

      case 'node_update':
        setNodes((prev) =>
          prev.map((node) =>
            node.id === operation.element_id
              ? { ...node, data: { ...node.data, ...operation.data } }
              : node
          )
        )
        break

      case 'node_delete':
        setNodes((prev) => prev.filter((node) => node.id !== operation.element_id))
        break

      case 'node_move':
        setNodes((prev) =>
          prev.map((node) =>
            node.id === operation.element_id
              ? { ...node, position: operation.data.position }
              : node
          )
        )
        break

      case 'edge_add':
        setEdges((prev) => [...prev, operation.data])
        break

      case 'edge_delete':
        setEdges((prev) => prev.filter((edge) => edge.id !== operation.element_id))
        break
    }
  }, [])

  // Track cursor movement on canvas
  const handleCanvasMouseMove = useCallback(
    (event: React.MouseEvent) => {
      const x = event.clientX
      const y = event.clientY

      // Throttle cursor updates (every 100ms)
      if (Date.now() % 100 < 16) {
        collaboration.updateCursor(x, y)
      }
    },
    [collaboration]
  )

  // Handle node selection with locking (used by onNodeSelect)
  const handleNodeSelect = useCallback(
    (node: Node | null) => {
      if (!node) {
        // Handle deselection
        if (selectedNode && collaboration.isLockedByMe(selectedNode.id)) {
          collaboration.releaseLock(selectedNode.id)
        }
        setSelectedNode(null)
        return
      }
      // Check if node is locked by another user
      if (collaboration.isLockedByOther(node.id)) {
        const lock = collaboration.getElementLock(node.id)
        alert(`Node is being edited by ${lock?.user_name}`)
        return
      }

      // Acquire lock if not already locked by me
      if (!collaboration.isLockedByMe(node.id)) {
        collaboration.acquireLock(node.id, 'node')
      }

      setSelectedNode(node)

      // Update selection in collaboration
      collaboration.updateSelection('node', [node.id])
    },
    [collaboration, selectedNode]
  )

  // Handle node updates with broadcasting
  const handleNodeUpdate = useCallback(
    (nodeId: string, updates: any) => {
      // Update local state
      setNodes((prev) =>
        prev.map((node) =>
          node.id === nodeId ? { ...node, data: { ...node.data, ...updates } } : node
        )
      )

      // Broadcast change to others
      const operation: EditOperation = {
        type: 'node_update',
        element_id: nodeId,
        data: updates,
        user_id: currentUser.id,
        timestamp: new Date().toISOString(),
      }

      collaboration.broadcastChange(operation)
    },
    [collaboration, currentUser.id]
  )

  // Handle node add from palette with broadcasting
  const handleNodeAddFromPalette = useCallback(
    (nodeInfo: { type: string; nodeType: string }) => {
      // Create a new node from palette info
      const newNode: Node = {
        id: `node-${Date.now()}`,
        type: nodeInfo.nodeType,
        position: { x: 250, y: 250 },
        data: { type: nodeInfo.type, label: nodeInfo.type },
      }

      // Update local state
      setNodes((prev) => [...prev, newNode])

      // Broadcast change
      const operation: EditOperation = {
        type: 'node_add',
        element_id: newNode.id,
        data: newNode,
        user_id: currentUser.id,
        timestamp: new Date().toISOString(),
      }

      collaboration.broadcastChange(operation)

      // Auto-lock new node
      collaboration.acquireLock(newNode.id, 'node')
    },
    [collaboration, currentUser.id]
  )

  // Release lock when deselecting node
  const handleNodeDeselect = useCallback(() => {
    if (selectedNode && collaboration.isLockedByMe(selectedNode.id)) {
      collaboration.releaseLock(selectedNode.id)
    }
    setSelectedNode(null)
  }, [selectedNode, collaboration])

  return (
    <div className="h-screen flex flex-col bg-gray-50 dark:bg-gray-900">
      {/* Header with presence indicator */}
      <div className="flex items-center justify-between px-4 py-3 bg-white dark:bg-gray-800 border-b border-gray-200 dark:border-gray-700">
        <h1 className="text-xl font-semibold text-gray-900 dark:text-white">
          Workflow Editor
        </h1>

        <UserPresenceIndicator
          connected={collaboration.connected}
          users={collaboration.users}
          currentUserId={currentUser.id}
        />
      </div>

      {/* Main content */}
      <div className="flex-1 flex overflow-hidden">
        {/* Left sidebar - Node palette */}
        <div className="w-64 bg-white dark:bg-gray-800 border-r border-gray-200 dark:border-gray-700 overflow-y-auto">
          <NodePalette onAddNode={handleNodeAddFromPalette} />

          {/* Collaborators list */}
          <div className="p-4 border-t border-gray-200 dark:border-gray-700">
            <CollaboratorList
              users={collaboration.users}
              currentUserId={currentUser.id}
            />
          </div>
        </div>

        {/* Center - Canvas */}
        <div
          className="flex-1 relative"
          onMouseMove={handleCanvasMouseMove}
        >
          {/* Show reconnecting overlay */}
          {collaboration.reconnecting && (
            <div className="absolute inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
              <div className="bg-white dark:bg-gray-800 rounded-lg p-6 shadow-lg">
                <div className="text-lg font-semibold text-gray-900 dark:text-white mb-2">
                  Reconnecting...
                </div>
                <div className="text-sm text-gray-600 dark:text-gray-400">
                  Trying to restore collaboration connection
                </div>
              </div>
            </div>
          )}

          <WorkflowCanvas
            initialNodes={nodes}
            initialEdges={edges}
            onNodeSelect={handleNodeSelect}
            onChange={({ nodes: newNodes, edges: newEdges }) => {
              setNodes(newNodes)
              setEdges(newEdges)
            }}
          />

          {/* Collaborator cursors overlay */}
          <CollaboratorCursors
            users={collaboration.users}
            currentUserId={currentUser.id}
          />

          {/* Lock indicators on nodes */}
          {nodes.map((node) => {
            const lock = collaboration.getElementLock(node.id)
            return lock ? (
              <div
                key={node.id}
                className="absolute"
                style={{
                  left: node.position.x,
                  top: node.position.y,
                }}
              >
                <NodeLockIndicator
                  lock={lock}
                  currentUserId={currentUser.id}
                />
              </div>
            ) : null
          })}
        </div>

        {/* Right sidebar - Properties panel */}
        {selectedNode && (
          <div className="w-80 bg-white dark:bg-gray-800 border-l border-gray-200 dark:border-gray-700 overflow-y-auto">
            <PropertyPanel
              node={selectedNode}
              onUpdate={handleNodeUpdate}
              onClose={handleNodeDeselect}
            />

            {/* Show lock warning if editing locked node */}
            {collaboration.isLockedByOther(selectedNode.id) && (
              <div className="p-4 bg-yellow-50 dark:bg-yellow-900 border-t border-yellow-200 dark:border-yellow-700">
                <div className="text-sm text-yellow-800 dark:text-yellow-200">
                  This node is being edited by{' '}
                  {collaboration.getElementLock(selectedNode.id)?.user_name}
                </div>
              </div>
            )}
          </div>
        )}
      </div>
    </div>
  )
}
