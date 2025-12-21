/**
 * AI Workflow Builder Types
 * Types for the AI-powered workflow generation feature
 */

// Conversation status
export type ConversationStatus = 'active' | 'completed' | 'abandoned'

// Message role
export type MessageRole = 'user' | 'assistant' | 'system'

// Node category for organization
export type NodeCategory = 'trigger' | 'action' | 'control' | 'integration'

/**
 * Build request for workflow generation
 */
export interface BuildRequest {
  description: string
  context?: BuildContext
  constraints?: BuildConstraints
}

/**
 * Context for workflow generation
 */
export interface BuildContext {
  available_credentials?: string[]
  available_integrations?: string[]
  existing_workflows?: string[]
  custom_data?: Record<string, unknown>
}

/**
 * Constraints for workflow generation
 */
export interface BuildConstraints {
  max_nodes?: number
  allowed_types?: string[]
  forbidden_types?: string[]
  require_trigger?: boolean
}

/**
 * Result of workflow generation
 */
export interface BuildResult {
  conversation_id: string
  workflow?: GeneratedWorkflow
  explanation: string
  warnings?: string[]
  suggestions?: string[]
  prompt_tokens?: number
  completion_tokens?: number
}

/**
 * Generated workflow structure
 */
export interface GeneratedWorkflow {
  name: string
  description?: string
  definition: WorkflowDefinition
}

/**
 * Workflow definition with nodes and edges
 */
export interface WorkflowDefinition {
  nodes: GeneratedNode[]
  edges?: GeneratedEdge[]
}

/**
 * Generated node in the workflow
 */
export interface GeneratedNode {
  id: string
  type: string
  name: string
  description?: string
  config?: Record<string, unknown>
  position?: NodePosition
}

/**
 * Position on the canvas
 */
export interface NodePosition {
  x: number
  y: number
}

/**
 * Generated edge connecting nodes
 */
export interface GeneratedEdge {
  id: string
  source: string
  target: string
  sourceHandle?: string
  targetHandle?: string
  label?: string
}

/**
 * Conversation for multi-turn workflow building
 */
export interface Conversation {
  id: string
  tenant_id: string
  user_id: string
  status: ConversationStatus
  current_workflow?: GeneratedWorkflow
  messages: ConversationMessage[]
  created_at: string
  updated_at: string
}

/**
 * Message in a conversation
 */
export interface ConversationMessage {
  id: string
  conversation_id?: string
  role: MessageRole
  content: string
  workflow?: GeneratedWorkflow
  prompt_tokens?: number
  completion_tokens?: number
  created_at: string
}

/**
 * Refine request for modifying an existing workflow
 */
export interface RefineRequest {
  conversation_id: string
  message: string
}

/**
 * Apply request for creating a real workflow
 */
export interface ApplyRequest {
  conversation_id: string
  workflow_name?: string
}

/**
 * Node template for available node types
 */
export interface NodeTemplate {
  id?: string
  tenant_id?: string
  node_type: string
  name: string
  description: string
  category: NodeCategory
  config_schema?: Record<string, unknown>
  example_config?: Record<string, unknown>
  llm_description: string
  is_active: boolean
}

// API Response types
export interface ConversationsListResponse {
  data: Conversation[]
}

export interface ConversationResponse {
  data: Conversation
}

export interface ApplyResponse {
  workflow_id: string
}

// Helper functions

/**
 * Get status label for display
 */
export function getStatusLabel(status: ConversationStatus): string {
  const labels: Record<ConversationStatus, string> = {
    active: 'Active',
    completed: 'Completed',
    abandoned: 'Abandoned',
  }
  return labels[status]
}

/**
 * Get status color for badges
 */
export function getStatusColor(status: ConversationStatus): string {
  const colors: Record<ConversationStatus, string> = {
    active: 'blue',
    completed: 'green',
    abandoned: 'gray',
  }
  return colors[status]
}

/**
 * Get role label for display
 */
export function getRoleLabel(role: MessageRole): string {
  const labels: Record<MessageRole, string> = {
    user: 'You',
    assistant: 'AI Assistant',
    system: 'System',
  }
  return labels[role]
}

/**
 * Get category label for display
 */
export function getCategoryLabel(category: NodeCategory): string {
  const labels: Record<NodeCategory, string> = {
    trigger: 'Triggers',
    action: 'Actions',
    control: 'Control Flow',
    integration: 'Integrations',
  }
  return labels[category]
}

/**
 * Get category icon
 */
export function getCategoryIcon(category: NodeCategory): string {
  const icons: Record<NodeCategory, string> = {
    trigger: '⚡',
    action: '🔧',
    control: '🔀',
    integration: '🔗',
  }
  return icons[category]
}

/**
 * Check if conversation is active
 */
export function isConversationActive(conversation: Conversation): boolean {
  return conversation.status === 'active'
}

/**
 * Get the last message from a conversation
 */
export function getLastMessage(
  conversation: Conversation
): ConversationMessage | undefined {
  if (conversation.messages.length === 0) {
    return undefined
  }
  return conversation.messages[conversation.messages.length - 1]
}

/**
 * Get messages by role
 */
export function getMessagesByRole(
  messages: ConversationMessage[],
  role: MessageRole
): ConversationMessage[] {
  return messages.filter((m) => m.role === role)
}
