import { create } from 'zustand'
import {
  Connection,
  Edge,
  EdgeChange,
  Node,
  NodeChange,
  addEdge,
  OnNodesChange,
  OnEdgesChange,
  OnConnect,
  applyNodeChanges,
  applyEdgeChanges,
} from '@xyflow/react'

// Node data types - using index signature for ReactFlow compatibility
export type TriggerNodeData = {
  label: string
  triggerType: 'webhook' | 'schedule'
  config: Record<string, unknown>
  [key: string]: unknown
}

export type ActionNodeData = {
  label: string
  actionType: 'http' | 'transform' | 'formula' | 'code' | 'script' | 'email' | 'slack_send_message' | 'slack_send_dm' | 'slack_update_message' | 'slack_add_reaction'
  config: Record<string, unknown>
  [key: string]: unknown
}

export type ControlNodeData = {
  label: string
  controlType: 'if' | 'loop' | 'parallel'
  config: Record<string, unknown>
  [key: string]: unknown
}

export type NodeData = TriggerNodeData | ActionNodeData | ControlNodeData

export type WorkflowNode = Node<NodeData>

// Workflow state
export interface Workflow {
  id: string
  name: string
  description: string
  status: 'draft' | 'active' | 'inactive'
  version: number
  createdAt: string
  updatedAt: string
}

interface WorkflowState {
  // Current workflow
  workflow: Workflow | null
  nodes: WorkflowNode[]
  edges: Edge[]

  // UI state
  selectedNode: string | null
  isDirty: boolean

  // Actions
  setWorkflow: (workflow: Workflow | null) => void
  setNodes: (nodes: WorkflowNode[]) => void
  setEdges: (edges: Edge[]) => void
  onNodesChange: OnNodesChange<WorkflowNode>
  onEdgesChange: OnEdgesChange
  onConnect: OnConnect

  addNode: (node: WorkflowNode) => void
  updateNode: (id: string, data: Partial<WorkflowNode['data']>) => void
  deleteNode: (id: string) => void

  selectNode: (id: string | null) => void

  resetWorkflow: () => void
  setDirty: (dirty: boolean) => void
}

const initialNodes: WorkflowNode[] = []
const initialEdges: Edge[] = []

export const useWorkflowStore = create<WorkflowState>((set, get) => ({
  // Initial state
  workflow: null,
  nodes: initialNodes,
  edges: initialEdges,
  selectedNode: null,
  isDirty: false,

  // Workflow actions
  setWorkflow: (workflow) => set({ workflow }),

  setNodes: (nodes) => set({ nodes, isDirty: true }),

  setEdges: (edges) => set({ edges, isDirty: true }),

  onNodesChange: (changes: NodeChange<WorkflowNode>[]) => {
    set({
      nodes: applyNodeChanges(changes, get().nodes),
      isDirty: true,
    })
  },

  onEdgesChange: (changes: EdgeChange[]) => {
    set({
      edges: applyEdgeChanges(changes, get().edges),
      isDirty: true,
    })
  },

  onConnect: (connection: Connection) => {
    set({
      edges: addEdge(connection, get().edges),
      isDirty: true,
    })
  },

  addNode: (node) => {
    set({
      nodes: [...get().nodes, node],
      isDirty: true,
    })
  },

  updateNode: (id, data) => {
    set({
      nodes: get().nodes.map((node) =>
        node.id === id ? { ...node, data: { ...node.data, ...data } } : node
      ),
      isDirty: true,
    })
  },

  deleteNode: (id) => {
    set({
      nodes: get().nodes.filter((node) => node.id !== id),
      edges: get().edges.filter((edge) => edge.source !== id && edge.target !== id),
      selectedNode: get().selectedNode === id ? null : get().selectedNode,
      isDirty: true,
    })
  },

  selectNode: (id) => set({ selectedNode: id }),

  resetWorkflow: () => set({
    workflow: null,
    nodes: initialNodes,
    edges: initialEdges,
    selectedNode: null,
    isDirty: false,
  }),

  setDirty: (dirty) => set({ isDirty: dirty }),
}))
