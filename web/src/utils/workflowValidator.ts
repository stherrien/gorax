/**
 * Comprehensive workflow validation utility
 * Provides real-time validation with actionable suggestions
 */

import type { Node, Edge } from '@xyflow/react'
import type { ValidationResult, ValidationIssue, ValidationSeverity } from '../types/workflow'
import { NODE_SCHEMAS, getNodeSchema } from '../data/nodeSchemas'
import { detectCycles, getTopologicalOrder } from './dagValidation'

// ============================================================================
// Main Validation Function
// ============================================================================

/**
 * Validate a workflow and return detailed issues
 */
export function validateWorkflow(nodes: Node[], edges: Edge[]): ValidationResult {
  const issues: ValidationIssue[] = []

  // Run all validation rules
  issues.push(...validateStructure(nodes, edges))
  issues.push(...validateNodes(nodes))
  issues.push(...validateEdges(nodes, edges))
  issues.push(...validateConnectivity(nodes, edges))
  issues.push(...validateDAG(nodes, edges))
  issues.push(...validateExpressions(nodes))

  // Get execution order if valid
  const topologicalResult = getTopologicalOrder(nodes, edges)
  const executionOrder = topologicalResult.success ? topologicalResult.order : undefined

  // Determine if workflow is valid (no errors)
  const valid = !issues.some((issue) => issue.severity === 'error')

  return {
    valid,
    issues,
    executionOrder,
  }
}

// ============================================================================
// Structure Validation
// ============================================================================

function validateStructure(nodes: Node[], _edges: Edge[]): ValidationIssue[] {
  const issues: ValidationIssue[] = []

  // Check for empty workflow
  if (nodes.length === 0) {
    issues.push(createIssue({
      severity: 'error',
      message: 'Workflow is empty',
      suggestion: 'Add at least one trigger node to start your workflow',
    }))
    return issues
  }

  // Check for trigger node
  const triggerNodes = nodes.filter((node) => node.type === 'trigger')
  if (triggerNodes.length === 0) {
    issues.push(createIssue({
      severity: 'error',
      message: 'Workflow must have a trigger',
      suggestion: 'Add a Webhook, Schedule, or Manual trigger node',
    }))
  }

  // Warn about multiple triggers
  if (triggerNodes.length > 1) {
    issues.push(createIssue({
      severity: 'warning',
      message: `Workflow has ${triggerNodes.length} trigger nodes`,
      suggestion: 'Consider using a single trigger for clarity, or use parallel execution',
    }))
  }

  return issues
}

// ============================================================================
// Node Validation
// ============================================================================

function validateNodes(nodes: Node[]): ValidationIssue[] {
  const issues: ValidationIssue[] = []

  for (const node of nodes) {
    const nodeType = (node.data as { nodeType?: string })?.nodeType || node.type || 'unknown'
    const schema = getNodeSchema(nodeType)

    // Validate against schema if available
    if (schema) {
      issues.push(...validateNodeAgainstSchema(node, schema))
    }

    // Validate node label
    const label = (node.data as { label?: string })?.label
    if (!label || label.trim() === '') {
      issues.push(createIssue({
        severity: 'warning',
        nodeId: node.id,
        field: 'label',
        message: 'Node has no name',
        suggestion: 'Add a descriptive name to help identify this node',
        autoFixable: true,
      }))
    }

    // Check for duplicate node IDs (shouldn't happen, but safety check)
    const duplicates = nodes.filter((n) => n.id === node.id)
    if (duplicates.length > 1) {
      issues.push(createIssue({
        severity: 'error',
        nodeId: node.id,
        message: 'Duplicate node ID detected',
        suggestion: 'This is a system error. Please reload the workflow.',
      }))
    }
  }

  return issues
}

function validateNodeAgainstSchema(node: Node, schema: typeof NODE_SCHEMAS[string]): ValidationIssue[] {
  const issues: ValidationIssue[] = []
  const data = node.data as Record<string, unknown>

  for (const field of schema.fields) {
    const value = data[field.name]

    // Check required fields
    if (field.required && (value === undefined || value === null || value === '')) {
      issues.push(createIssue({
        severity: 'error',
        nodeId: node.id,
        field: field.name,
        message: `${field.label} is required`,
        suggestion: `Enter a value for ${field.label}`,
        autoFixable: field.defaultValue !== undefined,
      }))
      continue
    }

    // Skip validation if empty and not required
    if (value === undefined || value === null || value === '') {
      continue
    }

    // Validate field-specific rules
    if (field.validation) {
      if (field.validation.pattern && typeof value === 'string') {
        const regex = new RegExp(field.validation.pattern)
        if (!regex.test(value)) {
          issues.push(createIssue({
            severity: 'error',
            nodeId: node.id,
            field: field.name,
            message: `${field.label} has invalid format`,
            suggestion: `Check the format of ${field.label}`,
          }))
        }
      }

      if (typeof value === 'number') {
        if (field.validation.min !== undefined && value < field.validation.min) {
          issues.push(createIssue({
            severity: 'error',
            nodeId: node.id,
            field: field.name,
            message: `${field.label} must be at least ${field.validation.min}`,
            suggestion: `Increase ${field.label} to at least ${field.validation.min}`,
            autoFixable: true,
          }))
        }
        if (field.validation.max !== undefined && value > field.validation.max) {
          issues.push(createIssue({
            severity: 'error',
            nodeId: node.id,
            field: field.name,
            message: `${field.label} must be at most ${field.validation.max}`,
            suggestion: `Reduce ${field.label} to at most ${field.validation.max}`,
            autoFixable: true,
          }))
        }
      }

      if (typeof value === 'string') {
        if (field.validation.minLength !== undefined && value.length < field.validation.minLength) {
          issues.push(createIssue({
            severity: 'error',
            nodeId: node.id,
            field: field.name,
            message: `${field.label} must be at least ${field.validation.minLength} characters`,
          }))
        }
        if (field.validation.maxLength !== undefined && value.length > field.validation.maxLength) {
          issues.push(createIssue({
            severity: 'error',
            nodeId: node.id,
            field: field.name,
            message: `${field.label} must be at most ${field.validation.maxLength} characters`,
          }))
        }
      }
    }

    // Validate JSON fields
    if (field.type === 'json' && typeof value === 'string' && value.trim() !== '') {
      try {
        JSON.parse(value)
      } catch {
        issues.push(createIssue({
          severity: 'error',
          nodeId: node.id,
          field: field.name,
          message: `${field.label} contains invalid JSON`,
          suggestion: 'Check JSON syntax - ensure proper quoting and brackets',
        }))
      }
    }
  }

  return issues
}

// ============================================================================
// Edge Validation
// ============================================================================

function validateEdges(nodes: Node[], edges: Edge[]): ValidationIssue[] {
  const issues: ValidationIssue[] = []
  const nodeIds = new Set(nodes.map((n) => n.id))

  for (const edge of edges) {
    // Check for dangling edges
    if (!nodeIds.has(edge.source)) {
      issues.push(createIssue({
        severity: 'error',
        message: `Edge references non-existent source node: ${edge.source}`,
        suggestion: 'Remove or reconnect this edge',
      }))
    }

    if (!nodeIds.has(edge.target)) {
      issues.push(createIssue({
        severity: 'error',
        message: `Edge references non-existent target node: ${edge.target}`,
        suggestion: 'Remove or reconnect this edge',
      }))
    }

    // Check for self-loops
    if (edge.source === edge.target) {
      issues.push(createIssue({
        severity: 'error',
        nodeId: edge.source,
        message: 'Node cannot connect to itself',
        suggestion: 'Remove the self-referencing connection',
      }))
    }
  }

  // Check for duplicate edges
  const edgeSet = new Set<string>()
  for (const edge of edges) {
    const key = `${edge.source}-${edge.target}`
    if (edgeSet.has(key)) {
      issues.push(createIssue({
        severity: 'warning',
        message: `Duplicate connection from ${edge.source} to ${edge.target}`,
        suggestion: 'Remove the duplicate connection',
      }))
    }
    edgeSet.add(key)
  }

  return issues
}

// ============================================================================
// Connectivity Validation
// ============================================================================

function validateConnectivity(nodes: Node[], edges: Edge[]): ValidationIssue[] {
  const issues: ValidationIssue[] = []

  // Build adjacency and reverse adjacency lists
  const outgoing = new Map<string, string[]>()
  const incoming = new Map<string, string[]>()

  for (const node of nodes) {
    outgoing.set(node.id, [])
    incoming.set(node.id, [])
  }

  for (const edge of edges) {
    outgoing.get(edge.source)?.push(edge.target)
    incoming.get(edge.target)?.push(edge.source)
  }

  // Check for disconnected nodes
  for (const node of nodes) {
    const nodeType = node.type || 'unknown'
    const hasIncoming = (incoming.get(node.id)?.length || 0) > 0
    const hasOutgoing = (outgoing.get(node.id)?.length || 0) > 0

    // Triggers should have no incoming but should have outgoing
    if (nodeType === 'trigger') {
      if (hasIncoming) {
        issues.push(createIssue({
          severity: 'warning',
          nodeId: node.id,
          message: 'Trigger nodes should not have incoming connections',
          suggestion: 'Triggers start workflows - remove incoming connections',
        }))
      }
      if (!hasOutgoing) {
        issues.push(createIssue({
          severity: 'warning',
          nodeId: node.id,
          message: 'Trigger is not connected to any other node',
          suggestion: 'Connect this trigger to an action or control node',
        }))
      }
    }
    // Non-triggers should have at least one incoming connection
    else if (!hasIncoming) {
      issues.push(createIssue({
        severity: 'warning',
        nodeId: node.id,
        message: 'Node is not connected to the workflow',
        suggestion: 'Connect this node to a trigger or another node',
      }))
    }
  }

  // Check reachability from triggers
  const triggerNodes = nodes.filter((n) => n.type === 'trigger')
  const reachable = new Set<string>()

  function dfs(nodeId: string) {
    if (reachable.has(nodeId)) return
    reachable.add(nodeId)
    for (const target of outgoing.get(nodeId) || []) {
      dfs(target)
    }
  }

  for (const trigger of triggerNodes) {
    dfs(trigger.id)
  }

  for (const node of nodes) {
    if (node.type !== 'trigger' && !reachable.has(node.id)) {
      issues.push(createIssue({
        severity: 'warning',
        nodeId: node.id,
        message: 'Node is unreachable from any trigger',
        suggestion: 'Connect this node to the main workflow path',
      }))
    }
  }

  return issues
}

// ============================================================================
// DAG Validation
// ============================================================================

function validateDAG(nodes: Node[], edges: Edge[]): ValidationIssue[] {
  const issues: ValidationIssue[] = []

  const cycles = detectCycles(nodes, edges)

  for (const cycle of cycles) {
    issues.push(createIssue({
      severity: 'error',
      message: `Cycle detected: ${cycle.join(' → ')}`,
      suggestion: 'Remove one of the connections to break the cycle',
    }))
  }

  return issues
}

// ============================================================================
// Expression Validation
// ============================================================================

function validateExpressions(nodes: Node[]): ValidationIssue[] {
  const issues: ValidationIssue[] = []
  const expressionPattern = /\{\{([^}]+)\}\}/g

  // Collect available variables from previous nodes
  const availableVariables = new Set<string>([
    'trigger',
    'trigger.data',
    'env',
  ])

  for (const node of nodes) {
    const data = node.data as Record<string, unknown>
    const label = (data.label as string) || node.id

    // Check all string fields for expressions
    for (const [key, value] of Object.entries(data)) {
      if (typeof value === 'string') {
        const matches = [...value.matchAll(expressionPattern)]

        for (const match of matches) {
          const expression = match[1].trim()
          const parts = expression.split('.')

          // Check for steps.* references
          if (parts[0] === 'steps' && parts.length >= 2) {
            const referencedNode = parts[1]
            // Find the referenced node
            const targetNode = nodes.find((n) => {
              const nodeLabel = ((n.data as { label?: string })?.label || '').toLowerCase().replace(/\s+/g, '_')
              return nodeLabel === referencedNode || n.id === referencedNode
            })

            if (!targetNode) {
              issues.push(createIssue({
                severity: 'warning',
                nodeId: node.id,
                field: key,
                message: `Expression references unknown node: ${referencedNode}`,
                suggestion: 'Check that the referenced node exists and has the correct name',
              }))
            }
          }
        }
      }
    }

    // Add this node's output as available variable
    availableVariables.add(`steps.${label.toLowerCase().replace(/\s+/g, '_')}`)
  }

  return issues
}

// ============================================================================
// Helper Functions
// ============================================================================

let issueIdCounter = 0

function createIssue(params: {
  severity: ValidationSeverity
  message: string
  nodeId?: string
  field?: string
  suggestion?: string
  autoFixable?: boolean
}): ValidationIssue {
  return {
    id: `issue-${++issueIdCounter}`,
    ...params,
  }
}

/**
 * Get human-readable summary of validation result
 */
export function getValidationSummary(result: ValidationResult): string {
  const errorCount = result.issues.filter((i) => i.severity === 'error').length
  const warningCount = result.issues.filter((i) => i.severity === 'warning').length
  const infoCount = result.issues.filter((i) => i.severity === 'info').length

  const parts: string[] = []
  if (errorCount > 0) parts.push(`${errorCount} error${errorCount > 1 ? 's' : ''}`)
  if (warningCount > 0) parts.push(`${warningCount} warning${warningCount > 1 ? 's' : ''}`)
  if (infoCount > 0) parts.push(`${infoCount} info`)

  if (parts.length === 0) {
    return 'Workflow is valid'
  }

  return parts.join(', ')
}

/**
 * Filter issues by severity
 */
export function filterIssuesBySeverity(
  issues: ValidationIssue[],
  severity: ValidationSeverity
): ValidationIssue[] {
  return issues.filter((issue) => issue.severity === severity)
}

/**
 * Get issues for a specific node
 */
export function getIssuesForNode(issues: ValidationIssue[], nodeId: string): ValidationIssue[] {
  return issues.filter((issue) => issue.nodeId === nodeId)
}
