/**
 * DiffHighlight - Utility component and functions for diff visualization
 * Provides color-coded highlighting for workflow comparison
 */

import type { ReactNode } from 'react'
import type {
  DiffStatus,
  NodeDiff,
  EdgeDiff,
  WorkflowDiff,
  DiffSummary,
  PropertyChange,
} from '../../../types/diff'
import { DIFF_COLORS } from '../../../types/diff'
import type { WorkflowNode, WorkflowEdge, WorkflowDefinition } from '../../../api/workflows'

// ============================================================================
// Diff Computation Functions
// ============================================================================

/**
 * Computes the diff between two workflow definitions
 */
export function computeWorkflowDiff(
  baseDefinition: WorkflowDefinition,
  compareDefinition: WorkflowDefinition,
  baseVersion: number,
  compareVersion: number
): WorkflowDiff {
  const nodeDiffs = computeNodeDiffs(
    baseDefinition.nodes || [],
    compareDefinition.nodes || []
  )

  const edgeDiffs = computeEdgeDiffs(
    baseDefinition.edges || [],
    compareDefinition.edges || []
  )

  // Cast to allow optional settings and variables properties
  const baseWithExtras = baseDefinition as WorkflowDefinition & { settings?: unknown; variables?: unknown }
  const compareWithExtras = compareDefinition as WorkflowDefinition & { settings?: unknown; variables?: unknown }

  const settingsChanged = !deepEqual(
    baseWithExtras.settings,
    compareWithExtras.settings
  )

  const variablesChanged = !deepEqual(
    baseWithExtras.variables,
    compareWithExtras.variables
  )

  const summary = computeSummary(nodeDiffs, edgeDiffs)

  return {
    baseVersion,
    compareVersion,
    summary,
    nodeDiffs,
    edgeDiffs,
    settingsChanged,
    variablesChanged,
  }
}

/**
 * Computes diffs for workflow nodes
 */
function computeNodeDiffs(
  baseNodes: WorkflowNode[],
  compareNodes: WorkflowNode[]
): NodeDiff[] {
  const diffs: NodeDiff[] = []
  const baseNodeMap = new Map(baseNodes.map((n) => [n.id, n]))
  const compareNodeMap = new Map(compareNodes.map((n) => [n.id, n]))

  // Check for removed and modified nodes
  for (const baseNode of baseNodes) {
    const compareNode = compareNodeMap.get(baseNode.id)

    if (!compareNode) {
      diffs.push({
        nodeId: baseNode.id,
        status: 'removed',
        baseNode,
      })
    } else if (!deepEqual(baseNode, compareNode)) {
      diffs.push({
        nodeId: baseNode.id,
        status: 'modified',
        baseNode,
        compareNode,
        propertyChanges: computePropertyChanges(baseNode, compareNode),
      })
    } else {
      diffs.push({
        nodeId: baseNode.id,
        status: 'unchanged',
        baseNode,
        compareNode,
      })
    }
  }

  // Check for added nodes
  for (const compareNode of compareNodes) {
    if (!baseNodeMap.has(compareNode.id)) {
      diffs.push({
        nodeId: compareNode.id,
        status: 'added',
        compareNode,
      })
    }
  }

  return diffs
}

/**
 * Computes diffs for workflow edges
 */
function computeEdgeDiffs(
  baseEdges: WorkflowEdge[],
  compareEdges: WorkflowEdge[]
): EdgeDiff[] {
  const diffs: EdgeDiff[] = []
  const baseEdgeMap = new Map(baseEdges.map((e) => [e.id, e]))
  const compareEdgeMap = new Map(compareEdges.map((e) => [e.id, e]))

  // Check for removed and modified edges
  for (const baseEdge of baseEdges) {
    const compareEdge = compareEdgeMap.get(baseEdge.id)

    if (!compareEdge) {
      diffs.push({
        edgeId: baseEdge.id,
        status: 'removed',
        baseEdge,
      })
    } else if (!deepEqual(baseEdge, compareEdge)) {
      diffs.push({
        edgeId: baseEdge.id,
        status: 'modified',
        baseEdge,
        compareEdge,
      })
    } else {
      diffs.push({
        edgeId: baseEdge.id,
        status: 'unchanged',
        baseEdge,
        compareEdge,
      })
    }
  }

  // Check for added edges
  for (const compareEdge of compareEdges) {
    if (!baseEdgeMap.has(compareEdge.id)) {
      diffs.push({
        edgeId: compareEdge.id,
        status: 'added',
        compareEdge,
      })
    }
  }

  return diffs
}

/**
 * Computes property-level changes between two nodes
 */
function computePropertyChanges(
  baseNode: WorkflowNode,
  compareNode: WorkflowNode
): PropertyChange[] {
  const changes: PropertyChange[] = []

  // Compare data properties
  const baseData = baseNode.data || {}
  const compareData = compareNode.data || {}

  const allKeys = new Set([...Object.keys(baseData), ...Object.keys(compareData)])

  for (const key of allKeys) {
    const baseValue = baseData[key]
    const compareValue = compareData[key]

    if (!(key in baseData)) {
      changes.push({ path: `data.${key}`, baseValue: undefined, compareValue, type: 'added' })
    } else if (!(key in compareData)) {
      changes.push({ path: `data.${key}`, baseValue, compareValue: undefined, type: 'removed' })
    } else if (!deepEqual(baseValue, compareValue)) {
      changes.push({ path: `data.${key}`, baseValue, compareValue, type: 'modified' })
    }
  }

  // Check position changes
  if (baseNode.position?.x !== compareNode.position?.x ||
      baseNode.position?.y !== compareNode.position?.y) {
    changes.push({
      path: 'position',
      baseValue: baseNode.position,
      compareValue: compareNode.position,
      type: 'modified',
    })
  }

  return changes
}

/**
 * Computes summary statistics for the diff
 */
function computeSummary(nodeDiffs: NodeDiff[], edgeDiffs: EdgeDiff[]): DiffSummary {
  const nodesAdded = nodeDiffs.filter((d) => d.status === 'added').length
  const nodesRemoved = nodeDiffs.filter((d) => d.status === 'removed').length
  const nodesModified = nodeDiffs.filter((d) => d.status === 'modified').length
  const nodesUnchanged = nodeDiffs.filter((d) => d.status === 'unchanged').length

  const edgesAdded = edgeDiffs.filter((d) => d.status === 'added').length
  const edgesRemoved = edgeDiffs.filter((d) => d.status === 'removed').length
  const edgesModified = edgeDiffs.filter((d) => d.status === 'modified').length
  const edgesUnchanged = edgeDiffs.filter((d) => d.status === 'unchanged').length

  return {
    nodesAdded,
    nodesRemoved,
    nodesModified,
    nodesUnchanged,
    edgesAdded,
    edgesRemoved,
    edgesModified,
    edgesUnchanged,
    totalChanges: nodesAdded + nodesRemoved + nodesModified + edgesAdded + edgesRemoved + edgesModified,
  }
}

// ============================================================================
// Deep Equality Helper
// ============================================================================

function deepEqual(a: unknown, b: unknown): boolean {
  if (a === b) return true
  if (a === null || b === null) return false
  if (typeof a !== typeof b) return false

  if (typeof a === 'object') {
    const aObj = a as Record<string, unknown>
    const bObj = b as Record<string, unknown>

    if (Array.isArray(aObj) !== Array.isArray(bObj)) return false

    const aKeys = Object.keys(aObj)
    const bKeys = Object.keys(bObj)

    if (aKeys.length !== bKeys.length) return false

    for (const key of aKeys) {
      if (!deepEqual(aObj[key], bObj[key])) return false
    }

    return true
  }

  return false
}

// ============================================================================
// UI Components
// ============================================================================

interface DiffBadgeProps {
  status: DiffStatus
  className?: string
}

/**
 * Badge component showing diff status
 */
export function DiffBadge({ status, className = '' }: DiffBadgeProps) {
  const colors = DIFF_COLORS[status]
  const labels: Record<DiffStatus, string> = {
    added: 'Added',
    removed: 'Removed',
    modified: 'Modified',
    unchanged: 'Unchanged',
  }

  return (
    <span
      className={`px-2 py-0.5 text-xs font-medium text-white rounded ${colors.badge} ${className}`}
    >
      {labels[status]}
    </span>
  )
}

interface DiffContainerProps {
  status: DiffStatus
  children: ReactNode
  className?: string
}

/**
 * Container with diff-status-based styling
 */
export function DiffContainer({ status, children, className = '' }: DiffContainerProps) {
  const colors = DIFF_COLORS[status]

  return (
    <div className={`rounded-lg border ${colors.bg} ${colors.border} ${className}`}>
      {children}
    </div>
  )
}

interface DiffSummaryDisplayProps {
  summary: DiffSummary
  className?: string
}

/**
 * Component to display diff summary statistics
 */
export function DiffSummaryDisplay({ summary, className = '' }: DiffSummaryDisplayProps) {
  return (
    <div className={`flex flex-wrap gap-4 text-sm ${className}`}>
      {summary.nodesAdded > 0 && (
        <div className="flex items-center gap-1">
          <span className="w-3 h-3 rounded bg-green-500" />
          <span className="text-green-400">{summary.nodesAdded} node(s) added</span>
        </div>
      )}
      {summary.nodesRemoved > 0 && (
        <div className="flex items-center gap-1">
          <span className="w-3 h-3 rounded bg-red-500" />
          <span className="text-red-400">{summary.nodesRemoved} node(s) removed</span>
        </div>
      )}
      {summary.nodesModified > 0 && (
        <div className="flex items-center gap-1">
          <span className="w-3 h-3 rounded bg-yellow-500" />
          <span className="text-yellow-400">{summary.nodesModified} node(s) modified</span>
        </div>
      )}
      {summary.edgesAdded > 0 && (
        <div className="flex items-center gap-1">
          <span className="w-3 h-3 rounded bg-green-500" />
          <span className="text-green-400">{summary.edgesAdded} connection(s) added</span>
        </div>
      )}
      {summary.edgesRemoved > 0 && (
        <div className="flex items-center gap-1">
          <span className="w-3 h-3 rounded bg-red-500" />
          <span className="text-red-400">{summary.edgesRemoved} connection(s) removed</span>
        </div>
      )}
      {summary.totalChanges === 0 && (
        <div className="text-gray-400">No changes detected</div>
      )}
    </div>
  )
}

interface PropertyChangeDisplayProps {
  changes: PropertyChange[]
  className?: string
}

/**
 * Component to display property-level changes
 */
export function PropertyChangeDisplay({ changes, className = '' }: PropertyChangeDisplayProps) {
  if (changes.length === 0) return null

  return (
    <div className={`space-y-2 ${className}`}>
      <div className="text-xs font-medium text-gray-400 uppercase">Property Changes</div>
      <div className="space-y-1">
        {changes.map((change, index) => (
          <div
            key={`${change.path}-${index}`}
            className="text-xs font-mono p-2 bg-gray-900 rounded"
          >
            <span className="text-gray-400">{change.path}: </span>
            {change.type === 'added' && (
              <span className="text-green-400">+ {formatValue(change.compareValue)}</span>
            )}
            {change.type === 'removed' && (
              <span className="text-red-400">- {formatValue(change.baseValue)}</span>
            )}
            {change.type === 'modified' && (
              <>
                <span className="text-red-400">{formatValue(change.baseValue)}</span>
                <span className="text-gray-500"> → </span>
                <span className="text-green-400">{formatValue(change.compareValue)}</span>
              </>
            )}
          </div>
        ))}
      </div>
    </div>
  )
}

function formatValue(value: unknown): string {
  if (value === undefined) return 'undefined'
  if (value === null) return 'null'
  if (typeof value === 'string') return `"${value}"`
  if (typeof value === 'object') return JSON.stringify(value)
  return String(value)
}

// ============================================================================
// Export index component for convenience
// ============================================================================

export { DIFF_COLORS }
