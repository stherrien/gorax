/**
 * Workflow file parser utilities
 * Parses JSON/YAML workflow files and validates their structure
 */

import type { Node, Edge } from '@xyflow/react'
import type { WorkflowEdge } from '../types/workflow'
import { getFileExtension, type SupportedExtension } from './fileValidation'
import {
  deserializeNodeFromBackend,
  type BackendNode,
} from './nodeTypeMapper'

// We use a simple YAML parser that handles most common cases
// For production, consider using 'yaml' or 'js-yaml' packages

/**
 * Result of parsing a workflow file
 */
export interface WorkflowParseResult {
  success: boolean
  nodes?: Node[]
  edges?: Edge[]
  name?: string
  description?: string
  error?: string
  errorDetails?: string[]
  warnings?: string[]
}

/**
 * Raw workflow structure from file
 */
interface RawWorkflowFile {
  name?: string
  description?: string
  definition?: {
    nodes?: unknown[]
    edges?: unknown[]
  }
  nodes?: unknown[]
  edges?: unknown[]
  // Legacy format support
  workflow?: {
    nodes?: unknown[]
    edges?: unknown[]
  }
}

/**
 * Parse YAML content (simple parser for common cases)
 * Note: For complex YAML, consider using 'yaml' or 'js-yaml' packages
 */
function parseYAML(content: string): unknown {
  // Remove comments
  const lines = content.split('\n').filter(line => !line.trim().startsWith('#'))
  const cleanContent = lines.join('\n')

  // Try to parse as JSON first (YAML is a superset of JSON)
  try {
    return JSON.parse(cleanContent)
  } catch {
    // Continue with YAML parsing
  }

  // Simple YAML parser for common workflow structures
  // This handles basic YAML but not all edge cases
  // For production, use a proper YAML library
  const result: Record<string, unknown> = {}
  const stack: { indent: number; obj: Record<string, unknown>; key?: string }[] = [
    { indent: -1, obj: result },
  ]

  for (const line of lines) {
    if (!line.trim()) continue

    const indent = line.search(/\S/)
    const trimmed = line.trim()

    // Skip if it's an array item indicator for now
    if (trimmed.startsWith('- ')) {
      // Handle array items - find the current array context
      while (stack.length > 1 && stack[stack.length - 1].indent >= indent) {
        stack.pop()
      }
      const current = stack[stack.length - 1]
      if (current.key && !Array.isArray(current.obj[current.key])) {
        current.obj[current.key] = []
      }
      if (current.key && Array.isArray(current.obj[current.key])) {
        const itemContent = trimmed.slice(2).trim()
        if (itemContent.includes(':')) {
          // Object in array
          const item: Record<string, unknown> = {}
          const [key, value] = itemContent.split(':').map(s => s.trim())
          if (value) {
            item[key] = parseValue(value)
          }
          (current.obj[current.key] as unknown[]).push(item)
          stack.push({ indent, obj: item, key: undefined })
        } else {
          // Simple value in array
          (current.obj[current.key] as unknown[]).push(parseValue(itemContent))
        }
      }
      continue
    }

    // Key-value pair
    const colonIndex = trimmed.indexOf(':')
    if (colonIndex === -1) continue

    const key = trimmed.slice(0, colonIndex).trim()
    const value = trimmed.slice(colonIndex + 1).trim()

    // Pop stack to find parent
    while (stack.length > 1 && stack[stack.length - 1].indent >= indent) {
      stack.pop()
    }

    const parent = stack[stack.length - 1].obj

    if (value) {
      // Simple key: value
      parent[key] = parseValue(value)
    } else {
      // Object or array follows
      parent[key] = {}
      stack.push({ indent, obj: parent[key] as Record<string, unknown>, key })
    }
  }

  return result
}

/**
 * Parse a YAML/JSON value
 */
function parseValue(value: string): unknown {
  // Remove quotes
  if ((value.startsWith('"') && value.endsWith('"')) ||
      (value.startsWith("'") && value.endsWith("'"))) {
    return value.slice(1, -1)
  }

  // Boolean
  if (value === 'true') return true
  if (value === 'false') return false

  // Null
  if (value === 'null' || value === '~') return null

  // Number
  const num = Number(value)
  if (!isNaN(num)) return num

  // String
  return value
}

/**
 * Parse workflow file content
 */
export function parseWorkflowContent(content: string, extension: SupportedExtension): WorkflowParseResult {
  const warnings: string[] = []

  try {
    let parsed: unknown

    if (extension === '.json') {
      try {
        parsed = JSON.parse(content)
      } catch (jsonError) {
        const error = jsonError as Error
        return {
          success: false,
          error: 'Invalid JSON format',
          errorDetails: [
            'The file contains invalid JSON.',
            error.message,
            'Please check for missing commas, brackets, or quotes.',
          ],
        }
      }
    } else {
      // YAML parsing
      try {
        parsed = parseYAML(content)
      } catch (yamlError) {
        const error = yamlError as Error
        return {
          success: false,
          error: 'Invalid YAML format',
          errorDetails: [
            'The file contains invalid YAML.',
            error.message,
            'Please check for proper indentation and syntax.',
          ],
        }
      }
    }

    // Validate structure
    if (!parsed || typeof parsed !== 'object') {
      return {
        success: false,
        error: 'Invalid workflow structure',
        errorDetails: ['The file does not contain a valid workflow object.'],
      }
    }

    const rawWorkflow = parsed as RawWorkflowFile

    // Extract nodes and edges from various possible structures
    let rawNodes: unknown[] = []
    let rawEdges: unknown[] = []
    const name = rawWorkflow.name
    const description = rawWorkflow.description

    // Try different structures
    if (rawWorkflow.definition?.nodes) {
      rawNodes = rawWorkflow.definition.nodes
      rawEdges = rawWorkflow.definition.edges || []
    } else if (rawWorkflow.nodes) {
      rawNodes = rawWorkflow.nodes
      rawEdges = rawWorkflow.edges || []
    } else if (rawWorkflow.workflow?.nodes) {
      rawNodes = rawWorkflow.workflow.nodes
      rawEdges = rawWorkflow.workflow.edges || []
      warnings.push('Using legacy workflow format. Consider updating to the current format.')
    }

    if (rawNodes.length === 0) {
      return {
        success: false,
        error: 'Workflow has no nodes',
        errorDetails: [
          'The workflow file must contain at least one node.',
          'Expected structure: { "nodes": [...], "edges": [...] }',
          'Or: { "definition": { "nodes": [...], "edges": [...] } }',
        ],
      }
    }

    // Parse and validate nodes
    const nodeResult = parseNodes(rawNodes)
    if (!nodeResult.success) {
      return {
        success: false,
        error: 'Invalid node structure',
        errorDetails: nodeResult.errors,
      }
    }
    warnings.push(...nodeResult.warnings)

    // Parse and validate edges
    const edgeResult = parseEdges(rawEdges, nodeResult.nodes!)
    if (!edgeResult.success) {
      return {
        success: false,
        error: 'Invalid edge structure',
        errorDetails: edgeResult.errors,
      }
    }
    warnings.push(...edgeResult.warnings)

    return {
      success: true,
      nodes: nodeResult.nodes,
      edges: edgeResult.edges,
      name,
      description,
      warnings: warnings.length > 0 ? warnings : undefined,
    }
  } catch (error) {
    const errorMessage = error instanceof Error ? error.message : 'Unknown parsing error'
    return {
      success: false,
      error: 'Failed to parse workflow file',
      errorDetails: [errorMessage],
    }
  }
}

/**
 * Parse and validate nodes array
 */
function parseNodes(rawNodes: unknown[]): {
  success: boolean
  nodes?: Node[]
  errors: string[]
  warnings: string[]
} {
  const errors: string[] = []
  const warnings: string[] = []
  const nodes: Node[] = []
  const seenIds = new Set<string>()

  for (let i = 0; i < rawNodes.length; i++) {
    const rawNode = rawNodes[i]
    const nodeIndex = i + 1

    if (!rawNode || typeof rawNode !== 'object') {
      errors.push(`Node ${nodeIndex}: Invalid node structure (not an object)`)
      continue
    }

    const nodeObj = rawNode as Record<string, unknown>

    // Check for required fields
    if (!nodeObj.id) {
      errors.push(`Node ${nodeIndex}: Missing required field 'id'`)
      continue
    }

    const nodeId = String(nodeObj.id)

    // Check for duplicate IDs
    if (seenIds.has(nodeId)) {
      errors.push(`Node ${nodeIndex}: Duplicate node ID '${nodeId}'`)
      continue
    }
    seenIds.add(nodeId)

    // Parse position
    let position = { x: 0, y: 0 }
    if (nodeObj.position && typeof nodeObj.position === 'object') {
      const pos = nodeObj.position as Record<string, unknown>
      position = {
        x: typeof pos.x === 'number' ? pos.x : 0,
        y: typeof pos.y === 'number' ? pos.y : 0,
      }
    } else {
      // Auto-position nodes that don't have positions
      position = { x: 100 + (i % 4) * 200, y: 100 + Math.floor(i / 4) * 150 }
      warnings.push(`Node '${nodeId}': No position specified, auto-positioned`)
    }

    // Detect node format and convert
    let node: Node

    // Check if it's backend format (type: "trigger:webhook") or frontend format
    const nodeType = nodeObj.type as string | undefined
    if (nodeType && nodeType.includes(':')) {
      // Backend format - convert to frontend
      const backendNode: BackendNode = {
        id: nodeId,
        type: nodeType,
        position,
        data: {
          name: (nodeObj.data as Record<string, unknown>)?.name as string || nodeType,
          config: ((nodeObj.data as Record<string, unknown>)?.config as Record<string, unknown>) || {},
        },
      }
      const frontendNode = deserializeNodeFromBackend(backendNode)
      node = {
        id: frontendNode.id,
        type: frontendNode.type,
        position: frontendNode.position,
        data: frontendNode.data,
      }
    } else {
      // Frontend format or generic format
      node = {
        id: nodeId,
        type: (nodeType || 'action') as string,
        position,
        data: nodeObj.data as Record<string, unknown> || { label: nodeId },
      }

      // Ensure label exists
      if (!node.data.label) {
        node.data = { ...node.data, label: nodeId }
      }
    }

    nodes.push(node)
  }

  return {
    success: errors.length === 0,
    nodes: errors.length === 0 ? nodes : undefined,
    errors,
    warnings,
  }
}

/**
 * Parse and validate edges array
 */
function parseEdges(rawEdges: unknown[], nodes: Node[]): {
  success: boolean
  edges?: Edge[]
  errors: string[]
  warnings: string[]
} {
  const errors: string[] = []
  const warnings: string[] = []
  const edges: Edge[] = []
  const nodeIds = new Set(nodes.map((n) => n.id))
  const seenEdges = new Set<string>()

  for (let i = 0; i < rawEdges.length; i++) {
    const rawEdge = rawEdges[i]
    const edgeIndex = i + 1

    if (!rawEdge || typeof rawEdge !== 'object') {
      errors.push(`Edge ${edgeIndex}: Invalid edge structure (not an object)`)
      continue
    }

    const edgeObj = rawEdge as Record<string, unknown>

    // Check for required fields
    if (!edgeObj.source) {
      errors.push(`Edge ${edgeIndex}: Missing required field 'source'`)
      continue
    }
    if (!edgeObj.target) {
      errors.push(`Edge ${edgeIndex}: Missing required field 'target'`)
      continue
    }

    const source = String(edgeObj.source)
    const target = String(edgeObj.target)

    // Validate source and target exist
    if (!nodeIds.has(source)) {
      errors.push(`Edge ${edgeIndex}: Source node '${source}' does not exist`)
      continue
    }
    if (!nodeIds.has(target)) {
      errors.push(`Edge ${edgeIndex}: Target node '${target}' does not exist`)
      continue
    }

    // Check for self-loops
    if (source === target) {
      errors.push(`Edge ${edgeIndex}: Self-loop detected (source equals target)`)
      continue
    }

    // Check for duplicates
    const edgeKey = `${source}-${target}`
    if (seenEdges.has(edgeKey)) {
      warnings.push(`Edge ${edgeIndex}: Duplicate edge from '${source}' to '${target}'`)
      continue
    }
    seenEdges.add(edgeKey)

    const edge: Edge = {
      id: (edgeObj.id as string) || `e-${source}-${target}`,
      source,
      target,
      sourceHandle: edgeObj.sourceHandle as string | undefined,
      targetHandle: edgeObj.targetHandle as string | undefined,
    }

    // Add label if present
    if (edgeObj.label) {
      (edge as WorkflowEdge).label = String(edgeObj.label)
    }

    edges.push(edge)
  }

  return {
    success: errors.length === 0,
    edges: errors.length === 0 ? edges : undefined,
    errors,
    warnings,
  }
}

/**
 * Parse a workflow file (combines file reading and parsing)
 */
export async function parseWorkflowFile(
  file: File,
  onProgress?: (progress: number) => void
): Promise<WorkflowParseResult> {
  const extension = getFileExtension(file.name) as SupportedExtension

  try {
    // Read file content
    const content = await new Promise<string>((resolve, reject) => {
      const reader = new FileReader()

      reader.onprogress = (event) => {
        if (event.lengthComputable && onProgress) {
          // Reading is 50% of the process
          const progress = Math.round((event.loaded / event.total) * 50)
          onProgress(progress)
        }
      }

      reader.onload = () => {
        if (typeof reader.result === 'string') {
          onProgress?.(50)
          resolve(reader.result)
        } else {
          reject(new Error('Failed to read file as text'))
        }
      }

      reader.onerror = () => {
        reject(new Error(`Failed to read file: ${reader.error?.message || 'Unknown error'}`))
      }

      reader.readAsText(file)
    })

    // Parse content (remaining 50%)
    onProgress?.(75)
    const result = parseWorkflowContent(content, extension)
    onProgress?.(100)

    return result
  } catch (error) {
    const errorMessage = error instanceof Error ? error.message : 'Unknown error'
    return {
      success: false,
      error: 'Failed to read workflow file',
      errorDetails: [errorMessage],
    }
  }
}

/**
 * Validate that a workflow has the required structure
 */
export function validateWorkflowStructure(nodes: Node[], edges: Edge[]): {
  valid: boolean
  errors: string[]
  warnings: string[]
} {
  const errors: string[] = []
  const warnings: string[] = []

  // Check for trigger node
  const triggerNodes = nodes.filter((n) => n.type === 'trigger')
  if (triggerNodes.length === 0) {
    warnings.push('Workflow has no trigger node. Add a trigger to start the workflow.')
  } else if (triggerNodes.length > 1) {
    warnings.push(`Workflow has ${triggerNodes.length} trigger nodes. Consider using a single trigger.`)
  }

  // Check for disconnected nodes
  const connectedNodes = new Set<string>()
  for (const edge of edges) {
    connectedNodes.add(edge.source)
    connectedNodes.add(edge.target)
  }

  for (const node of nodes) {
    if (node.type !== 'trigger' && !connectedNodes.has(node.id)) {
      warnings.push(`Node '${node.data?.label || node.id}' is not connected to any other node.`)
    }
  }

  return {
    valid: errors.length === 0,
    errors,
    warnings,
  }
}
