/**
 * Node Transformation Utility
 *
 * Provides bidirectional transformation between frontend (ReactFlow) and backend (API) node formats.
 *
 * Frontend Node:
 *   - type: ReactFlow category ("trigger", "action", "ai", "control")
 *   - data.type: Full qualified type ("trigger:webhook")
 *   - data.nodeType: Short type ("webhook")
 *   - data.label: Display name
 *   - data.[...]: Config fields spread at top level
 *
 * Backend Node:
 *   - type: Full qualified type ("trigger:webhook")
 *   - data.name: Display name
 *   - data.config: Config fields nested in config object
 */

import type { Node, Edge } from '@xyflow/react'

// ==========================================
// Frontend (ReactFlow) Types
// ==========================================

export interface FrontendNodePosition {
  x: number
  y: number
}

export interface FrontendNodeData {
  label: string // Display name
  type: string // Full qualified type: "trigger:webhook"
  nodeType: string // Short type: "webhook"
  [key: string]: unknown // Config fields at top level
}

export interface FrontendNode extends Node {
  type: string // ReactFlow category: "trigger", "action", "ai", "control"
  position: FrontendNodePosition
  data: FrontendNodeData
}

export interface FrontendEdge extends Edge {
  id: string
  source: string
  target: string
  sourceHandle?: string
  targetHandle?: string
}

// ==========================================
// Backend (API) Types
// ==========================================

export interface BackendNodePosition {
  x: number
  y: number
}

export interface BackendNodeData {
  name: string // From frontend label
  config: Record<string, unknown> // All config fields nested here
}

export interface BackendNode {
  id: string
  type: string // Full qualified type: "trigger:webhook"
  position: BackendNodePosition
  data: BackendNodeData
}

export interface BackendEdge {
  id: string
  source: string
  target: string
  sourceHandle?: string
  targetHandle?: string
}

export interface BackendWorkflowDefinition {
  nodes: BackendNode[]
  edges: BackendEdge[]
}

// ==========================================
// Helper Functions
// ==========================================

/**
 * Parse a full qualified type into category and short type.
 * "trigger:webhook" -> ["trigger", "webhook"]
 * "action" -> ["action", "action"]
 */
export function parseFullType(fullType: string): [string, string] {
  if (!fullType) {
    return ['action', 'unknown']
  }
  const parts = fullType.split(':')
  if (parts.length >= 2) {
    return [parts[0], parts.slice(1).join(':')]
  }
  // Single part type - use as both category and nodeType
  return [fullType, fullType]
}

/**
 * Build a full qualified type from category and short type.
 * ("trigger", "webhook") -> "trigger:webhook"
 */
export function buildFullType(category: string, nodeType: string): string {
  if (!category || !nodeType) {
    return category || nodeType || 'action:unknown'
  }
  // Don't double-prefix if already full type
  if (nodeType.includes(':')) {
    return nodeType
  }
  return `${category}:${nodeType}`
}

/**
 * Get the ReactFlow rendering category from a full type.
 * Maps to custom node component types registered in ReactFlow.
 */
export function getReactFlowCategory(fullType: string): string {
  const [category] = parseFullType(fullType)
  // Map to valid ReactFlow node types
  const categoryMap: Record<string, string> = {
    trigger: 'trigger',
    action: 'action',
    ai: 'ai',
    control: 'control',
    // Legacy mappings
    conditional: 'control',
    loop: 'control',
    parallel: 'control',
  }
  return categoryMap[category] || category
}

// ==========================================
// Node Transformation Functions
// ==========================================

/**
 * Transform a frontend (ReactFlow) node to backend (API) format.
 *
 * Extracts:
 * - data.type -> node.type (full qualified)
 * - data.label -> data.name
 * - Remaining data fields -> data.config
 */
export function transformNodeToBackend(frontendNode: FrontendNode | Node): BackendNode {
  const data = (frontendNode.data || {}) as Record<string, unknown>
  const { label, type, nodeType, ...configFields } = data

  // Get the full qualified type from data.type, or construct from node.type + nodeType
  const fullType =
    (type as string) ||
    (frontendNode.type && nodeType
      ? buildFullType(frontendNode.type, nodeType as string)
      : frontendNode.type || 'action:unknown')

  return {
    id: frontendNode.id,
    type: fullType,
    position: {
      x: frontendNode.position?.x ?? 0,
      y: frontendNode.position?.y ?? 0,
    },
    data: {
      name: (label as string) || 'Unnamed Node',
      config: configFields,
    },
  }
}

/**
 * Transform a backend (API) node to frontend (ReactFlow) format.
 *
 * Maps:
 * - node.type -> data.type (full qualified)
 * - node.type split -> node.type (category) + data.nodeType (short)
 * - data.name -> data.label
 * - data.config -> spread to data top level
 */
export function transformNodeToFrontend(backendNode: BackendNode): FrontendNode {
  const fullType = backendNode.type || 'action:unknown'
  const [, nodeType] = parseFullType(fullType)
  const reactFlowCategory = getReactFlowCategory(fullType)

  const backendData = (backendNode.data || {}) as BackendNodeData
  const config = (backendData.config || {}) as Record<string, unknown>

  return {
    id: backendNode.id,
    type: reactFlowCategory,
    position: {
      x: backendNode.position?.x ?? 0,
      y: backendNode.position?.y ?? 0,
    },
    data: {
      label: backendData.name || 'Unnamed Node',
      type: fullType,
      nodeType: nodeType,
      ...config,
    },
  }
}

// ==========================================
// Edge Transformation Functions
// ==========================================

/**
 * Transform a frontend edge to backend format.
 * Edges are mostly the same, just clean up undefined fields.
 */
export function transformEdgeToBackend(frontendEdge: FrontendEdge | Edge): BackendEdge {
  const result: BackendEdge = {
    id: frontendEdge.id,
    source: frontendEdge.source,
    target: frontendEdge.target,
  }

  // Only include handles if they exist
  if (frontendEdge.sourceHandle) {
    result.sourceHandle = frontendEdge.sourceHandle
  }
  if (frontendEdge.targetHandle) {
    result.targetHandle = frontendEdge.targetHandle
  }

  return result
}

/**
 * Transform a backend edge to frontend format.
 */
export function transformEdgeToFrontend(backendEdge: BackendEdge): FrontendEdge {
  return {
    id: backendEdge.id,
    source: backendEdge.source,
    target: backendEdge.target,
    sourceHandle: backendEdge.sourceHandle,
    targetHandle: backendEdge.targetHandle,
  }
}

// ==========================================
// Workflow Definition Transformation
// ==========================================

/**
 * Transform frontend workflow data to backend format.
 */
export function transformWorkflowToBackend(
  nodes: (FrontendNode | Node)[],
  edges: (FrontendEdge | Edge)[]
): BackendWorkflowDefinition {
  return {
    nodes: nodes.map(transformNodeToBackend),
    edges: edges.map(transformEdgeToBackend),
  }
}

// Generic workflow definition type for API compatibility
// The API may return nodes with `data: Record<string, unknown>` format
interface GenericWorkflowDefinition {
  nodes?: Array<{
    id: string
    type: string
    position?: { x: number; y: number }
    data?: Record<string, unknown>
  }>
  edges?: Array<{
    id: string
    source: string
    target: string
    sourceHandle?: string
    targetHandle?: string
  }>
}

/**
 * Transform backend workflow definition to frontend format.
 * Accepts both strict BackendWorkflowDefinition and looser API types.
 */
export function transformWorkflowToFrontend(
  definition: BackendWorkflowDefinition | GenericWorkflowDefinition | null | undefined
): {
  nodes: FrontendNode[]
  edges: FrontendEdge[]
} {
  if (!definition) {
    return { nodes: [], edges: [] }
  }

  return {
    nodes: (definition.nodes || []).map((node) => transformNodeToFrontend(node as BackendNode)),
    edges: (definition.edges || []).map((edge) => transformEdgeToFrontend(edge as BackendEdge)),
  }
}
