/**
 * Node Type Mapper - Converts between frontend ReactFlow node types and backend API types
 *
 * Frontend uses: { type: 'trigger', data: { nodeType: 'webhook' } }
 * Backend expects: { type: 'trigger:webhook' }
 *
 * This utility ensures type safety between frontend canvas and backend API.
 */

// Backend node types (from internal/workflow/model.go)
export const BACKEND_NODE_TYPES = {
  // Triggers
  TRIGGER_WEBHOOK: 'trigger:webhook',
  TRIGGER_SCHEDULE: 'trigger:schedule',

  // Actions
  ACTION_HTTP: 'action:http',
  ACTION_TRANSFORM: 'action:transform',
  ACTION_FORMULA: 'action:formula',
  ACTION_CODE: 'action:code',
  ACTION_EMAIL: 'action:email',
  ACTION_SUBWORKFLOW: 'action:subworkflow',

  // Slack actions (use slack: prefix, not action:)
  SLACK_SEND_MESSAGE: 'slack:send_message',
  SLACK_SEND_DM: 'slack:send_dm',
  SLACK_UPDATE_MESSAGE: 'slack:update_message',
  SLACK_ADD_REACTION: 'slack:add_reaction',

  // Control flow
  CONTROL_IF: 'control:if',
  CONTROL_LOOP: 'control:loop',
  CONTROL_PARALLEL: 'control:parallel',
  CONTROL_FORK: 'control:fork',
  CONTROL_JOIN: 'control:join',
  CONTROL_DELAY: 'control:delay',
  CONTROL_TRY: 'control:try',
  CONTROL_CATCH: 'control:catch',
  CONTROL_FINALLY: 'control:finally',
  CONTROL_RETRY: 'control:retry',
  CONTROL_CIRCUIT_BREAKER: 'control:circuit_breaker',
} as const

export type BackendNodeType = typeof BACKEND_NODE_TYPES[keyof typeof BACKEND_NODE_TYPES]

// Frontend node type categories (used in ReactFlow)
export type FrontendNodeCategory = 'trigger' | 'action' | 'ai' | 'control'

// Mapping from frontend (category + nodeType) to backend type
const FRONTEND_TO_BACKEND_MAP: Record<string, BackendNodeType> = {
  // Triggers
  'trigger:webhook': 'trigger:webhook',
  'trigger:schedule': 'trigger:schedule',
  'trigger:manual': 'trigger:webhook', // Manual triggers use webhook infrastructure

  // Standard actions
  'action:http': 'action:http',
  'action:transform': 'action:transform',
  'action:formula': 'action:formula',
  'action:email': 'action:email',
  'action:script': 'action:code', // Frontend "script" maps to backend "code"
  'action:code': 'action:code',
  'action:subworkflow': 'action:subworkflow',

  // Slack actions (special prefix)
  'action:slack_send_message': 'slack:send_message',
  'action:slack_send_dm': 'slack:send_dm',
  'action:slack_update_message': 'slack:update_message',
  'action:slack_add_reaction': 'slack:add_reaction',

  // AI actions (map to action:code with AI config for now)
  'ai:ai_chat': 'action:code',
  'ai:ai_summarize': 'action:code',
  'ai:ai_classify': 'action:code',
  'ai:ai_extract': 'action:code',
  'ai:ai_embed': 'action:code',

  // Control flow
  'control:conditional': 'control:if', // Frontend "conditional" maps to backend "if"
  'control:if': 'control:if',
  'control:loop': 'control:loop',
  'control:parallel': 'control:parallel',
  'control:fork': 'control:fork',
  'control:join': 'control:join',
  'control:delay': 'control:delay',
  'control:try': 'control:try',
  'control:retry': 'control:retry',
}

// Mapping from backend type to frontend (category, nodeType)
const BACKEND_TO_FRONTEND_MAP: Record<string, { category: FrontendNodeCategory; nodeType: string }> = {
  'trigger:webhook': { category: 'trigger', nodeType: 'webhook' },
  'trigger:schedule': { category: 'trigger', nodeType: 'schedule' },

  'action:http': { category: 'action', nodeType: 'http' },
  'action:transform': { category: 'action', nodeType: 'transform' },
  'action:formula': { category: 'action', nodeType: 'formula' },
  'action:code': { category: 'action', nodeType: 'script' },
  'action:email': { category: 'action', nodeType: 'email' },
  'action:subworkflow': { category: 'action', nodeType: 'subworkflow' },

  'slack:send_message': { category: 'action', nodeType: 'slack_send_message' },
  'slack:send_dm': { category: 'action', nodeType: 'slack_send_dm' },
  'slack:update_message': { category: 'action', nodeType: 'slack_update_message' },
  'slack:add_reaction': { category: 'action', nodeType: 'slack_add_reaction' },

  'control:if': { category: 'control', nodeType: 'conditional' },
  'control:loop': { category: 'control', nodeType: 'loop' },
  'control:parallel': { category: 'control', nodeType: 'parallel' },
  'control:fork': { category: 'control', nodeType: 'fork' },
  'control:join': { category: 'control', nodeType: 'join' },
  'control:delay': { category: 'control', nodeType: 'delay' },
  'control:try': { category: 'control', nodeType: 'try' },
  'control:retry': { category: 'control', nodeType: 'retry' },
}

/**
 * Convert frontend node type to backend format
 *
 * @param category - Frontend node category ('trigger', 'action', 'ai', 'control')
 * @param nodeType - Specific node type ('webhook', 'http', 'conditional', etc.)
 * @returns Backend node type string (e.g., 'trigger:webhook', 'control:if')
 */
export function toBackendNodeType(category: string, nodeType: string): string {
  const key = `${category}:${nodeType}`
  const mapped = FRONTEND_TO_BACKEND_MAP[key]

  if (mapped) {
    return mapped
  }

  // Fallback: construct type directly (may not be valid on backend)
  console.warn(`Unknown node type mapping: ${key}. Using fallback.`)
  return key
}

/**
 * Convert backend node type to frontend format
 *
 * @param backendType - Backend node type string (e.g., 'trigger:webhook')
 * @returns Object with category and nodeType for ReactFlow
 */
export function toFrontendNodeType(backendType: string): { category: FrontendNodeCategory; nodeType: string } {
  const mapped = BACKEND_TO_FRONTEND_MAP[backendType]

  if (mapped) {
    return mapped
  }

  // Fallback: parse the type string
  const [category, nodeType] = backendType.split(':')
  console.warn(`Unknown backend type: ${backendType}. Using parsed fallback.`)
  return {
    category: (category as FrontendNodeCategory) || 'action',
    nodeType: nodeType || 'unknown',
  }
}

/**
 * Serialize a ReactFlow node for backend API
 * Converts frontend node structure to backend-compatible format
 */
export interface FrontendNode {
  id: string
  type?: string
  position: { x: number; y: number }
  data: {
    nodeType?: string
    label?: string
    [key: string]: unknown
  }
}

export interface BackendNode {
  id: string
  type: string
  position: { x: number; y: number }
  data: {
    name: string
    config: Record<string, unknown>
  }
}

export function serializeNodeForBackend(node: FrontendNode): BackendNode {
  const category = node.type || 'action'
  const nodeType = node.data?.nodeType || 'unknown'
  const backendType = toBackendNodeType(category, nodeType)

  // Extract config from node data (remove UI-specific fields)
  // eslint-disable-next-line @typescript-eslint/no-unused-vars
  const { nodeType: _nodeType, label, ...config } = node.data || {}

  return {
    id: node.id,
    type: backendType,
    position: node.position,
    data: {
      name: label || `${nodeType} node`,
      config: config as Record<string, unknown>,
    },
  }
}

/**
 * Deserialize a backend node for ReactFlow
 * Converts backend node structure to frontend-compatible format
 */
export function deserializeNodeFromBackend(node: BackendNode): FrontendNode {
  const { category, nodeType } = toFrontendNodeType(node.type)

  return {
    id: node.id,
    type: category,
    position: node.position,
    data: {
      nodeType,
      label: node.data?.name || `${nodeType} node`,
      ...node.data?.config,
    },
  }
}

/**
 * Serialize entire workflow definition for backend
 */
export interface FrontendWorkflowDefinition {
  nodes: FrontendNode[]
  edges: Array<{
    id: string
    source: string
    target: string
    sourceHandle?: string
    targetHandle?: string
    label?: string
  }>
}

export interface BackendWorkflowDefinition {
  nodes: BackendNode[]
  edges: Array<{
    id: string
    source: string
    target: string
    sourceHandle?: string
    targetHandle?: string
    label?: string
  }>
}

export function serializeWorkflowForBackend(definition: FrontendWorkflowDefinition): BackendWorkflowDefinition {
  return {
    nodes: definition.nodes.map(serializeNodeForBackend),
    edges: definition.edges.map(edge => ({
      id: edge.id,
      source: edge.source,
      target: edge.target,
      sourceHandle: edge.sourceHandle,
      targetHandle: edge.targetHandle,
      label: edge.label,
    })),
  }
}

export function deserializeWorkflowFromBackend(definition: BackendWorkflowDefinition): FrontendWorkflowDefinition {
  return {
    nodes: definition.nodes.map(deserializeNodeFromBackend),
    edges: definition.edges,
  }
}

/**
 * Validate that a node type is valid for the backend
 */
export function isValidBackendNodeType(type: string): boolean {
  return Object.values(BACKEND_NODE_TYPES).includes(type as BackendNodeType)
}

/**
 * Get all valid trigger types
 */
export function getTriggerTypes(): BackendNodeType[] {
  return [
    BACKEND_NODE_TYPES.TRIGGER_WEBHOOK,
    BACKEND_NODE_TYPES.TRIGGER_SCHEDULE,
  ]
}

/**
 * Check if a type is a trigger type
 */
export function isTriggerType(type: string): boolean {
  return type.startsWith('trigger:')
}
