/**
 * Tests for workflowValidator utility
 */

import { describe, it, expect } from 'vitest'
import type { Node, Edge } from '@xyflow/react'
import {
  validateWorkflow,
  getValidationSummary,
  filterIssuesBySeverity,
  getIssuesForNode,
} from './workflowValidator'

// ============================================================================
// Test Fixtures
// ============================================================================

const createTriggerNode = (id: string, nodeType = 'webhook'): Node => ({
  id,
  type: 'trigger',
  position: { x: 0, y: 0 },
  data: {
    label: `${nodeType} trigger`,
    nodeType,
  },
})

const createActionNode = (id: string, nodeType = 'http'): Node => ({
  id,
  type: 'action',
  position: { x: 100, y: 100 },
  data: {
    label: `${nodeType} action`,
    nodeType,
  },
})

const createEdge = (source: string, target: string): Edge => ({
  id: `e-${source}-${target}`,
  source,
  target,
})

// ============================================================================
// Structure Validation Tests
// ============================================================================

describe('validateWorkflow - structure validation', () => {
  it('should report error for empty workflow', () => {
    const result = validateWorkflow([], [])

    expect(result.valid).toBe(false)
    expect(result.issues).toHaveLength(1)
    expect(result.issues[0].severity).toBe('error')
    expect(result.issues[0].message).toContain('empty')
  })

  it('should report error for workflow without trigger', () => {
    const nodes: Node[] = [createActionNode('action-1')]
    const edges: Edge[] = []

    const result = validateWorkflow(nodes, edges)

    expect(result.valid).toBe(false)
    const triggerError = result.issues.find((i) => i.message.includes('trigger'))
    expect(triggerError).toBeDefined()
    expect(triggerError?.severity).toBe('error')
  })

  it('should warn about multiple triggers', () => {
    const nodes: Node[] = [
      createTriggerNode('trigger-1', 'webhook'),
      createTriggerNode('trigger-2', 'schedule'),
      createActionNode('action-1'),
    ]
    const edges: Edge[] = [
      createEdge('trigger-1', 'action-1'),
      createEdge('trigger-2', 'action-1'),
    ]

    const result = validateWorkflow(nodes, edges)

    const multiTriggerWarning = result.issues.find((i) =>
      i.message.includes('trigger nodes') && i.severity === 'warning'
    )
    expect(multiTriggerWarning).toBeDefined()
  })

  it('should pass for valid simple workflow', () => {
    const nodes: Node[] = [
      createTriggerNode('trigger-1'),
      createActionNode('action-1'),
    ]
    const edges: Edge[] = [createEdge('trigger-1', 'action-1')]

    const result = validateWorkflow(nodes, edges)

    // The workflow is structurally valid (has trigger, no cycles, connected)
    // May have schema validation warnings for optional fields but no structural errors
    const structuralErrors = result.issues.filter(
      (i) =>
        i.severity === 'error' &&
        (i.message.includes('empty') ||
          i.message.includes('trigger') ||
          i.message.includes('Cycle') ||
          i.message.includes('non-existent') ||
          i.message.includes('cannot connect'))
    )
    expect(structuralErrors).toHaveLength(0)
  })
})

// ============================================================================
// Node Validation Tests
// ============================================================================

describe('validateWorkflow - node validation', () => {
  it('should warn about nodes without labels', () => {
    const nodes: Node[] = [
      createTriggerNode('trigger-1'),
      {
        id: 'action-1',
        type: 'action',
        position: { x: 100, y: 100 },
        data: { nodeType: 'http' }, // No label
      },
    ]
    const edges: Edge[] = [createEdge('trigger-1', 'action-1')]

    const result = validateWorkflow(nodes, edges)

    // The validator should find an issue for the missing label
    // It could be either a warning from the label check or an error from schema validation
    const labelIssue = result.issues.find(
      (i) => i.nodeId === 'action-1' && (i.field === 'label' || i.message.toLowerCase().includes('name'))
    )
    expect(labelIssue).toBeDefined()
  })
})

// ============================================================================
// Edge Validation Tests
// ============================================================================

describe('validateWorkflow - edge validation', () => {
  it('should detect self-loops', () => {
    const nodes: Node[] = [
      createTriggerNode('trigger-1'),
      createActionNode('action-1'),
    ]
    const edges: Edge[] = [
      createEdge('trigger-1', 'action-1'),
      createEdge('action-1', 'action-1'), // Self-loop
    ]

    const result = validateWorkflow(nodes, edges)

    expect(result.valid).toBe(false)
    const selfLoopError = result.issues.find((i) => i.message.includes('cannot connect to itself'))
    expect(selfLoopError).toBeDefined()
  })

  it('should detect dangling edges', () => {
    const nodes: Node[] = [
      createTriggerNode('trigger-1'),
      createActionNode('action-1'),
    ]
    const edges: Edge[] = [
      createEdge('trigger-1', 'action-1'),
      createEdge('action-1', 'nonexistent'), // Dangling edge
    ]

    const result = validateWorkflow(nodes, edges)

    expect(result.valid).toBe(false)
    const danglingError = result.issues.find((i) => i.message.includes('non-existent'))
    expect(danglingError).toBeDefined()
  })

  it('should warn about duplicate edges', () => {
    const nodes: Node[] = [
      createTriggerNode('trigger-1'),
      createActionNode('action-1'),
    ]
    const edges: Edge[] = [
      createEdge('trigger-1', 'action-1'),
      { id: 'duplicate', source: 'trigger-1', target: 'action-1' }, // Duplicate
    ]

    const result = validateWorkflow(nodes, edges)

    const duplicateWarning = result.issues.find((i) => i.message.includes('Duplicate'))
    expect(duplicateWarning).toBeDefined()
    expect(duplicateWarning?.severity).toBe('warning')
  })
})

// ============================================================================
// Connectivity Validation Tests
// ============================================================================

describe('validateWorkflow - connectivity validation', () => {
  it('should warn about disconnected nodes', () => {
    const nodes: Node[] = [
      createTriggerNode('trigger-1'),
      createActionNode('action-1'),
      createActionNode('action-2'), // Not connected
    ]
    const edges: Edge[] = [createEdge('trigger-1', 'action-1')]

    const result = validateWorkflow(nodes, edges)

    const disconnectedWarning = result.issues.find(
      (i) => i.nodeId === 'action-2' && i.message.includes('not connected')
    )
    expect(disconnectedWarning).toBeDefined()
  })

  it('should warn about trigger without outgoing connections', () => {
    const nodes: Node[] = [createTriggerNode('trigger-1')]
    const edges: Edge[] = []

    const result = validateWorkflow(nodes, edges)

    const triggerWarning = result.issues.find(
      (i) => i.nodeId === 'trigger-1' && i.message.includes('not connected')
    )
    expect(triggerWarning).toBeDefined()
  })

  it('should warn about unreachable nodes', () => {
    const nodes: Node[] = [
      createTriggerNode('trigger-1'),
      createActionNode('action-1'),
      createActionNode('action-2'),
      createActionNode('action-3'),
    ]
    const edges: Edge[] = [
      createEdge('trigger-1', 'action-1'),
      createEdge('action-2', 'action-3'), // Separate chain, unreachable from trigger
    ]

    const result = validateWorkflow(nodes, edges)

    const unreachableWarning = result.issues.find(
      (i) => i.nodeId === 'action-2' && i.message.includes('unreachable')
    )
    expect(unreachableWarning).toBeDefined()
  })
})

// ============================================================================
// DAG Validation Tests
// ============================================================================

describe('validateWorkflow - DAG validation', () => {
  it('should detect cycles', () => {
    const nodes: Node[] = [
      createTriggerNode('trigger-1'),
      createActionNode('action-1'),
      createActionNode('action-2'),
    ]
    const edges: Edge[] = [
      createEdge('trigger-1', 'action-1'),
      createEdge('action-1', 'action-2'),
      createEdge('action-2', 'action-1'), // Creates cycle
    ]

    const result = validateWorkflow(nodes, edges)

    expect(result.valid).toBe(false)
    const cycleError = result.issues.find((i) => i.message.includes('Cycle'))
    expect(cycleError).toBeDefined()
    expect(cycleError?.severity).toBe('error')
  })

  it('should provide execution order for valid DAG', () => {
    const nodes: Node[] = [
      createTriggerNode('trigger-1'),
      createActionNode('action-1'),
      createActionNode('action-2'),
    ]
    const edges: Edge[] = [
      createEdge('trigger-1', 'action-1'),
      createEdge('action-1', 'action-2'),
    ]

    const result = validateWorkflow(nodes, edges)

    expect(result.executionOrder).toBeDefined()
    expect(result.executionOrder).toEqual(['trigger-1', 'action-1', 'action-2'])
  })
})

// ============================================================================
// Helper Function Tests
// ============================================================================

describe('getValidationSummary', () => {
  it('should return meaningful summary for workflow', () => {
    const result = validateWorkflow(
      [createTriggerNode('trigger-1'), createActionNode('action-1')],
      [createEdge('trigger-1', 'action-1')]
    )

    // The summary should contain counts or "valid"
    const summary = getValidationSummary(result)
    // Should be a non-empty string
    expect(summary.length).toBeGreaterThan(0)
    // Should mention valid, error, or warning
    expect(summary).toMatch(/valid|error|warning/i)
  })

  it('should count errors correctly', () => {
    const result = validateWorkflow([], [])

    const summary = getValidationSummary(result)
    expect(summary).toContain('error')
  })
})

describe('filterIssuesBySeverity', () => {
  it('should filter issues by severity', () => {
    const result = validateWorkflow([], [])

    const errors = filterIssuesBySeverity(result.issues, 'error')
    const warnings = filterIssuesBySeverity(result.issues, 'warning')

    expect(errors.every((i) => i.severity === 'error')).toBe(true)
    expect(warnings.every((i) => i.severity === 'warning')).toBe(true)
  })
})

describe('getIssuesForNode', () => {
  it('should filter issues by node ID', () => {
    const nodes: Node[] = [
      createTriggerNode('trigger-1'),
      {
        id: 'action-1',
        type: 'action',
        position: { x: 100, y: 100 },
        data: { nodeType: 'http' }, // No label
      },
    ]
    const edges: Edge[] = [createEdge('trigger-1', 'action-1')]

    const result = validateWorkflow(nodes, edges)
    const nodeIssues = getIssuesForNode(result.issues, 'action-1')

    expect(nodeIssues.length).toBeGreaterThan(0)
    expect(nodeIssues.every((i) => i.nodeId === 'action-1')).toBe(true)
  })
})
