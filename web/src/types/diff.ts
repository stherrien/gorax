/**
 * Type definitions for workflow version comparison and diff visualization
 */

import type { WorkflowDefinition, WorkflowNode, WorkflowEdge } from '../api/workflows'

// ============================================================================
// Version Types
// ============================================================================

export interface WorkflowVersionInfo {
  id: string
  workflowId: string
  version: number
  definition: WorkflowDefinition
  createdAt: string
  createdBy: string
  comment?: string
}

// ============================================================================
// Diff Status Types
// ============================================================================

export type DiffStatus = 'added' | 'removed' | 'modified' | 'unchanged'

export interface DiffMeta {
  status: DiffStatus
  changes?: string[]
}

// ============================================================================
// Node Diff Types
// ============================================================================

export interface NodeDiff {
  nodeId: string
  status: DiffStatus
  baseNode?: WorkflowNode
  compareNode?: WorkflowNode
  propertyChanges?: PropertyChange[]
}

export interface PropertyChange {
  path: string
  baseValue: unknown
  compareValue: unknown
  type: 'added' | 'removed' | 'modified'
}

// ============================================================================
// Edge Diff Types
// ============================================================================

export interface EdgeDiff {
  edgeId: string
  status: DiffStatus
  baseEdge?: WorkflowEdge
  compareEdge?: WorkflowEdge
}

// ============================================================================
// Workflow Diff Result
// ============================================================================

export interface WorkflowDiff {
  baseVersion: number
  compareVersion: number
  summary: DiffSummary
  nodeDiffs: NodeDiff[]
  edgeDiffs: EdgeDiff[]
  settingsChanged: boolean
  variablesChanged: boolean
}

export interface DiffSummary {
  nodesAdded: number
  nodesRemoved: number
  nodesModified: number
  nodesUnchanged: number
  edgesAdded: number
  edgesRemoved: number
  edgesModified: number
  edgesUnchanged: number
  totalChanges: number
}

// ============================================================================
// Comparison State Types
// ============================================================================

export interface ComparisonState {
  baseVersionId: string | null
  compareVersionId: string | null
  viewMode: 'visual' | 'json'
  diff: WorkflowDiff | null
  loading: boolean
  error: string | null
}

// ============================================================================
// Display Options
// ============================================================================

export interface DiffDisplayOptions {
  showUnchanged: boolean
  highlightPropertyChanges: boolean
  expandedNodes: Set<string>
  syncScroll: boolean
}

// ============================================================================
// Color Scheme
// ============================================================================

export const DIFF_COLORS = {
  added: {
    bg: 'bg-green-900/30',
    border: 'border-green-500',
    text: 'text-green-400',
    badge: 'bg-green-600',
  },
  removed: {
    bg: 'bg-red-900/30',
    border: 'border-red-500',
    text: 'text-red-400',
    badge: 'bg-red-600',
  },
  modified: {
    bg: 'bg-yellow-900/30',
    border: 'border-yellow-500',
    text: 'text-yellow-400',
    badge: 'bg-yellow-600',
  },
  unchanged: {
    bg: 'bg-gray-800',
    border: 'border-gray-600',
    text: 'text-gray-400',
    badge: 'bg-gray-600',
  },
} as const

// ============================================================================
// Utility Type Guards
// ============================================================================

export function isNodeAdded(diff: NodeDiff): boolean {
  return diff.status === 'added'
}

export function isNodeRemoved(diff: NodeDiff): boolean {
  return diff.status === 'removed'
}

export function isNodeModified(diff: NodeDiff): boolean {
  return diff.status === 'modified'
}

export function isNodeUnchanged(diff: NodeDiff): boolean {
  return diff.status === 'unchanged'
}
