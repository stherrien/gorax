/**
 * Workflow type definitions for the visual workflow builder
 * These types align with the backend Go types in internal/workflow/
 */

import type { Node, Edge } from '@xyflow/react'

// ============================================================================
// Core Workflow Types
// ============================================================================

export type WorkflowStatus = 'draft' | 'active' | 'inactive' | 'archived'

export interface Workflow {
  id: string
  tenantId: string
  name: string
  description?: string
  status: WorkflowStatus
  version: number
  definition: WorkflowDefinition
  createdAt: string
  updatedAt: string
  createdBy?: string
  updatedBy?: string
}

export interface WorkflowDefinition {
  nodes: WorkflowNode[]
  edges: WorkflowEdge[]
  variables?: WorkflowVariable[]
  settings?: WorkflowSettings
}

export interface WorkflowSettings {
  timeout?: number
  retryPolicy?: RetryPolicy
  logging?: LoggingConfig
}

export interface RetryPolicy {
  maxRetries: number
  initialDelay: number
  maxDelay: number
  backoffMultiplier: number
}

export interface LoggingConfig {
  level: 'debug' | 'info' | 'warn' | 'error'
  includeInput: boolean
  includeOutput: boolean
}

// ============================================================================
// Node Types
// ============================================================================

export type NodeType =
  | 'trigger'
  | 'action'
  | 'ai'
  | 'conditional'
  | 'loop'
  | 'parallel'
  | 'fork'
  | 'join'
  | 'subworkflow'
  | 'retry'
  | 'try'

export type TriggerType = 'webhook' | 'schedule' | 'manual' | 'event'

export type ActionType =
  | 'http'
  | 'transform'
  | 'script'
  | 'email'
  | 'slack_send_message'
  | 'slack_send_dm'
  | 'slack_update_message'
  | 'slack_add_reaction'
  | 'github_create_issue'
  | 'github_create_pr'
  | 'jira_create_issue'
  | 'pagerduty_trigger'
  | 'human_task'

export type AIActionType =
  | 'ai_chat'
  | 'ai_summarize'
  | 'ai_classify'
  | 'ai_extract'
  | 'ai_embed'

export type ControlType = 'conditional' | 'loop' | 'parallel' | 'fork' | 'join'

// Base node data interface
export interface BaseNodeData {
  label: string
  description?: string
  [key: string]: unknown // Index signature for ReactFlow compatibility
}

// Trigger node data
export interface TriggerNodeData extends BaseNodeData {
  triggerType: TriggerType
  config: TriggerConfig
}

export interface TriggerConfig {
  // Webhook config
  path?: string
  method?: 'GET' | 'POST' | 'PUT' | 'DELETE' | 'PATCH'
  authType?: 'none' | 'basic' | 'signature' | 'api_key'
  secret?: string
  priority?: number

  // Schedule config
  cron?: string
  timezone?: string

  // Event config
  eventType?: string
  filter?: string
}

// Action node data
export interface ActionNodeData extends BaseNodeData {
  actionType: ActionType
  config: ActionConfig
  credentialId?: string
}

export interface ActionConfig {
  // HTTP config
  url?: string
  method?: 'GET' | 'POST' | 'PUT' | 'DELETE' | 'PATCH'
  headers?: Record<string, string>
  body?: string
  timeout?: number

  // Transform config
  expression?: string
  mapping?: Record<string, string>

  // Script config
  language?: 'javascript' | 'python'
  code?: string

  // Email config
  to?: string
  subject?: string
  bodyTemplate?: string

  // Slack config
  channel?: string
  user?: string
  text?: string
  blocks?: string
  ts?: string
  emoji?: string

  // GitHub config
  repository?: string
  title?: string
  bodyContent?: string

  // Jira config
  project?: string
  issueType?: string
  summary?: string

  // PagerDuty config
  serviceId?: string
  severity?: 'critical' | 'error' | 'warning' | 'info'

  // Human task config
  assignee?: string
  taskTitle?: string
  taskDescription?: string
  dueDate?: string
}

// AI node data
export interface AINodeData extends BaseNodeData {
  aiActionType: AIActionType
  config: AIConfig
}

export interface AIConfig {
  model?: string
  provider?: 'openai' | 'anthropic' | 'bedrock'
  credentialId?: string
  prompt?: string
  systemPrompt?: string
  temperature?: number
  maxTokens?: number
  categories?: string[]
  entityTypes?: string[]
}

// Control node data
export interface ConditionalNodeData extends BaseNodeData {
  controlType: 'conditional'
  condition: string
  trueLabel?: string
  falseLabel?: string
}

export interface LoopNodeData extends BaseNodeData {
  controlType: 'loop'
  source: string
  itemVariable: string
  indexVariable?: string
  maxIterations?: number
  onError: 'stop' | 'continue' | 'skip'
}

export interface ParallelNodeData extends BaseNodeData {
  controlType: 'parallel'
  errorStrategy: 'fail_fast' | 'continue_on_error' | 'wait_all'
  maxConcurrency?: number
}

export interface ForkNodeData extends BaseNodeData {
  controlType: 'fork'
  branches: string[]
}

export interface JoinNodeData extends BaseNodeData {
  controlType: 'join'
  waitFor: 'all' | 'any' | 'first'
}

// Union type for all node data
export type WorkflowNodeData =
  | TriggerNodeData
  | ActionNodeData
  | AINodeData
  | ConditionalNodeData
  | LoopNodeData
  | ParallelNodeData
  | ForkNodeData
  | JoinNodeData

// ReactFlow-compatible workflow node
export type WorkflowNode = Node<WorkflowNodeData>

// ============================================================================
// Edge Types
// ============================================================================

export interface WorkflowEdge extends Edge {
  label?: string
  condition?: string // For conditional edges
  sourceHandle?: string
  targetHandle?: string
}

// ============================================================================
// Variable Types
// ============================================================================

export interface WorkflowVariable {
  name: string
  type: 'string' | 'number' | 'boolean' | 'object' | 'array'
  defaultValue?: unknown
  description?: string
  required?: boolean
}

// ============================================================================
// Validation Types
// ============================================================================

export type ValidationSeverity = 'error' | 'warning' | 'info'

export interface ValidationIssue {
  id: string
  nodeId?: string
  field?: string
  severity: ValidationSeverity
  message: string
  suggestion?: string
  autoFixable?: boolean
}

export interface ValidationResult {
  valid: boolean
  issues: ValidationIssue[]
  executionOrder?: string[]
}

// ============================================================================
// Execution Types
// ============================================================================

export type ExecutionStatus =
  | 'pending'
  | 'running'
  | 'completed'
  | 'failed'
  | 'cancelled'
  | 'timeout'
  | 'skipped'

export interface NodeExecutionState {
  nodeId: string
  status: ExecutionStatus
  startedAt?: string
  completedAt?: string
  input?: unknown
  output?: unknown
  error?: string
  duration?: number
}

export interface WorkflowExecution {
  id: string
  workflowId: string
  workflowVersion: number
  status: ExecutionStatus
  trigger: {
    type: TriggerType
    data: unknown
  }
  nodeStates: Record<string, NodeExecutionState>
  startedAt: string
  completedAt?: string
  error?: string
}

// ============================================================================
// Node Schema Types (for dynamic configuration)
// ============================================================================

export type FieldType =
  | 'text'
  | 'textarea'
  | 'number'
  | 'select'
  | 'multiselect'
  | 'boolean'
  | 'json'
  | 'expression'
  | 'credential'
  | 'variable'

export interface FieldOption {
  value: string
  label: string
  description?: string
}

export interface FieldSchema {
  name: string
  label: string
  type: FieldType
  required?: boolean
  defaultValue?: unknown
  placeholder?: string
  description?: string
  options?: FieldOption[]
  validation?: {
    pattern?: string
    min?: number
    max?: number
    minLength?: number
    maxLength?: number
  }
  dependsOn?: {
    field: string
    value: unknown
  }
}

export interface NodeSchema {
  type: string
  label: string
  description: string
  icon: string
  category: 'trigger' | 'action' | 'ai' | 'control'
  fields: FieldSchema[]
  inputs?: number
  outputs?: number
  outputLabels?: string[]
}

// ============================================================================
// Builder State Types
// ============================================================================

export interface WorkflowBuilderState {
  workflow: Workflow | null
  nodes: WorkflowNode[]
  edges: WorkflowEdge[]
  selectedNode: string | null
  selectedEdge: string | null
  isDirty: boolean
  validationResult: ValidationResult | null
  undoStack: WorkflowDefinition[]
  redoStack: WorkflowDefinition[]
}

export interface WorkflowBuilderActions {
  setWorkflow: (workflow: Workflow | null) => void
  setNodes: (nodes: WorkflowNode[]) => void
  setEdges: (edges: WorkflowEdge[]) => void
  addNode: (node: WorkflowNode) => void
  updateNode: (nodeId: string, data: Partial<WorkflowNodeData>) => void
  deleteNode: (nodeId: string) => void
  addEdge: (edge: WorkflowEdge) => void
  updateEdge: (edgeId: string, data: Partial<WorkflowEdge>) => void
  deleteEdge: (edgeId: string) => void
  selectNode: (nodeId: string | null) => void
  selectEdge: (edgeId: string | null) => void
  undo: () => void
  redo: () => void
  validate: () => ValidationResult
  reset: () => void
}

// ============================================================================
// API Response Types
// ============================================================================

export interface WorkflowListResponse {
  workflows: Workflow[]
  total: number
  page: number
  pageSize: number
}

export interface WorkflowVersionResponse {
  id: string
  workflowId: string
  version: number
  definition: WorkflowDefinition
  createdAt: string
  createdBy?: string
  comment?: string
}

export interface DryRunResponse {
  valid: boolean
  executionOrder: string[]
  errors: ValidationIssue[]
  warnings: ValidationIssue[]
  variableMapping: Record<string, string>
}
