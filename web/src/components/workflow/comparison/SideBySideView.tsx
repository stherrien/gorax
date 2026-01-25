/**
 * SideBySideView - Visual side-by-side comparison of workflow versions
 * Shows two panels with synchronized scrolling and diff highlighting
 */

import { useRef, useCallback, useState } from 'react'
import type { WorkflowDiff, NodeDiff, EdgeDiff } from '../../../types/diff'
import type { WorkflowDefinition, WorkflowEdge } from '../../../api/workflows'
import { DiffBadge, DiffContainer, DiffSummaryDisplay, PropertyChangeDisplay, DIFF_COLORS } from './DiffHighlight'

interface SideBySideViewProps {
  baseDefinition: WorkflowDefinition
  compareDefinition: WorkflowDefinition
  baseVersion: number
  compareVersion: number
  diff: WorkflowDiff
  showUnchanged?: boolean
}

export default function SideBySideView({
  baseDefinition,
  compareDefinition,
  baseVersion,
  compareVersion,
  diff,
  showUnchanged = false,
}: SideBySideViewProps) {
  const [syncScroll, setSyncScroll] = useState(true)
  const [expandedNodes, setExpandedNodes] = useState<Set<string>>(new Set())
  const leftPanelRef = useRef<HTMLDivElement>(null)
  const rightPanelRef = useRef<HTMLDivElement>(null)
  const isScrolling = useRef(false)

  // Handle synchronized scrolling
  const handleScroll = useCallback((source: 'left' | 'right') => {
    if (!syncScroll || isScrolling.current) return

    isScrolling.current = true
    const sourcePanel = source === 'left' ? leftPanelRef.current : rightPanelRef.current
    const targetPanel = source === 'left' ? rightPanelRef.current : leftPanelRef.current

    if (sourcePanel && targetPanel) {
      targetPanel.scrollTop = sourcePanel.scrollTop
    }

    setTimeout(() => {
      isScrolling.current = false
    }, 50)
  }, [syncScroll])

  const toggleNodeExpand = (nodeId: string) => {
    setExpandedNodes((prev) => {
      const next = new Set(prev)
      if (next.has(nodeId)) {
        next.delete(nodeId)
      } else {
        next.add(nodeId)
      }
      return next
    })
  }

  const expandAll = () => {
    setExpandedNodes(new Set(diff.nodeDiffs.map((d) => d.nodeId)))
  }

  const collapseAll = () => {
    setExpandedNodes(new Set())
  }

  // Filter nodes based on showUnchanged setting
  const filteredNodeDiffs = showUnchanged
    ? diff.nodeDiffs
    : diff.nodeDiffs.filter((d) => d.status !== 'unchanged')

  const filteredEdgeDiffs = showUnchanged
    ? diff.edgeDiffs
    : diff.edgeDiffs.filter((d) => d.status !== 'unchanged')

  return (
    <div className="flex flex-col h-full bg-gray-900 rounded-lg overflow-hidden">
      {/* Summary Header */}
      <div className="p-4 bg-gray-800 border-b border-gray-700">
        <div className="flex items-center justify-between mb-3">
          <h3 className="text-lg font-semibold text-white">
            Visual Comparison
          </h3>
          <div className="flex items-center gap-4">
            <label className="flex items-center gap-2 text-sm text-gray-400">
              <input
                type="checkbox"
                checked={syncScroll}
                onChange={(e) => setSyncScroll(e.target.checked)}
                className="rounded bg-gray-700 border-gray-600 text-primary-600 focus:ring-primary-500"
              />
              Sync scroll
            </label>
            <div className="flex gap-2">
              <button
                onClick={expandAll}
                className="px-2 py-1 text-xs bg-gray-700 text-gray-300 hover:bg-gray-600 rounded transition-colors"
              >
                Expand all
              </button>
              <button
                onClick={collapseAll}
                className="px-2 py-1 text-xs bg-gray-700 text-gray-300 hover:bg-gray-600 rounded transition-colors"
              >
                Collapse all
              </button>
            </div>
          </div>
        </div>
        <DiffSummaryDisplay summary={diff.summary} />
      </div>

      {/* Side-by-side Panels */}
      <div className="flex-1 flex min-h-0">
        {/* Left Panel - Base Version */}
        <div className="flex-1 flex flex-col border-r border-gray-700">
          <div className="p-3 bg-gray-800/50 border-b border-gray-700">
            <div className="flex items-center gap-2">
              <span className="text-sm font-medium text-gray-300">Base</span>
              <span className="px-2 py-0.5 text-xs bg-gray-600 text-white rounded">
                v{baseVersion}
              </span>
              <span className="text-xs text-gray-500">
                {baseDefinition.nodes?.length || 0} nodes
              </span>
            </div>
          </div>
          <div
            ref={leftPanelRef}
            onScroll={() => handleScroll('left')}
            className="flex-1 overflow-y-auto p-4 space-y-3"
          >
            {/* Nodes Section */}
            <SectionHeader title="Nodes" count={filteredNodeDiffs.length} />
            {filteredNodeDiffs.map((nodeDiff) => (
              <NodeDiffCard
                key={nodeDiff.nodeId}
                nodeDiff={nodeDiff}
                side="base"
                expanded={expandedNodes.has(nodeDiff.nodeId)}
                onToggle={() => toggleNodeExpand(nodeDiff.nodeId)}
              />
            ))}

            {/* Edges Section */}
            {filteredEdgeDiffs.length > 0 && (
              <>
                <SectionHeader title="Connections" count={filteredEdgeDiffs.length} />
                {filteredEdgeDiffs.map((edgeDiff) => (
                  <EdgeDiffCard
                    key={edgeDiff.edgeId}
                    edgeDiff={edgeDiff}
                    side="base"
                  />
                ))}
              </>
            )}

            {/* Settings/Variables Changed Indicators */}
            {diff.settingsChanged && (
              <div className="p-3 bg-yellow-900/30 border border-yellow-500 rounded-lg">
                <span className="text-yellow-400 text-sm">Workflow settings changed</span>
              </div>
            )}
            {diff.variablesChanged && (
              <div className="p-3 bg-yellow-900/30 border border-yellow-500 rounded-lg">
                <span className="text-yellow-400 text-sm">Workflow variables changed</span>
              </div>
            )}
          </div>
        </div>

        {/* Right Panel - Compare Version */}
        <div className="flex-1 flex flex-col">
          <div className="p-3 bg-gray-800/50 border-b border-gray-700">
            <div className="flex items-center gap-2">
              <span className="text-sm font-medium text-gray-300">Compare</span>
              <span className="px-2 py-0.5 text-xs bg-primary-600 text-white rounded">
                v{compareVersion}
              </span>
              <span className="text-xs text-gray-500">
                {compareDefinition.nodes?.length || 0} nodes
              </span>
            </div>
          </div>
          <div
            ref={rightPanelRef}
            onScroll={() => handleScroll('right')}
            className="flex-1 overflow-y-auto p-4 space-y-3"
          >
            {/* Nodes Section */}
            <SectionHeader title="Nodes" count={filteredNodeDiffs.length} />
            {filteredNodeDiffs.map((nodeDiff) => (
              <NodeDiffCard
                key={nodeDiff.nodeId}
                nodeDiff={nodeDiff}
                side="compare"
                expanded={expandedNodes.has(nodeDiff.nodeId)}
                onToggle={() => toggleNodeExpand(nodeDiff.nodeId)}
              />
            ))}

            {/* Edges Section */}
            {filteredEdgeDiffs.length > 0 && (
              <>
                <SectionHeader title="Connections" count={filteredEdgeDiffs.length} />
                {filteredEdgeDiffs.map((edgeDiff) => (
                  <EdgeDiffCard
                    key={edgeDiff.edgeId}
                    edgeDiff={edgeDiff}
                    side="compare"
                  />
                ))}
              </>
            )}

            {/* Settings/Variables Changed Indicators */}
            {diff.settingsChanged && (
              <div className="p-3 bg-yellow-900/30 border border-yellow-500 rounded-lg">
                <span className="text-yellow-400 text-sm">Workflow settings changed</span>
              </div>
            )}
            {diff.variablesChanged && (
              <div className="p-3 bg-yellow-900/30 border border-yellow-500 rounded-lg">
                <span className="text-yellow-400 text-sm">Workflow variables changed</span>
              </div>
            )}
          </div>
        </div>
      </div>
    </div>
  )
}

// ============================================================================
// Helper Components
// ============================================================================

interface SectionHeaderProps {
  title: string
  count: number
}

function SectionHeader({ title, count }: SectionHeaderProps) {
  return (
    <div className="flex items-center gap-2 pb-2 border-b border-gray-700">
      <h4 className="text-sm font-medium text-gray-400 uppercase">{title}</h4>
      <span className="px-1.5 py-0.5 text-xs bg-gray-700 text-gray-300 rounded">
        {count}
      </span>
    </div>
  )
}

interface NodeDiffCardProps {
  nodeDiff: NodeDiff
  side: 'base' | 'compare'
  expanded: boolean
  onToggle: () => void
}

function NodeDiffCard({ nodeDiff, side, expanded, onToggle }: NodeDiffCardProps) {
  const node = side === 'base' ? nodeDiff.baseNode : nodeDiff.compareNode
  const isVisible = side === 'base'
    ? nodeDiff.status !== 'added'
    : nodeDiff.status !== 'removed'

  // Placeholder for removed/added nodes on opposite side
  if (!isVisible) {
    return (
      <div className="p-3 bg-gray-800/30 border border-dashed border-gray-600 rounded-lg">
        <span className="text-gray-500 text-sm italic">
          {nodeDiff.status === 'added' ? 'Node added in compare version' : 'Node removed in compare version'}
        </span>
      </div>
    )
  }

  if (!node) return null

  const nodeType = node.type || 'unknown'
  const nodeLabel = (node.data as { label?: string })?.label || node.id

  return (
    <DiffContainer status={nodeDiff.status} className="overflow-hidden">
      <button
        onClick={onToggle}
        className="w-full p-3 flex items-center justify-between text-left hover:bg-gray-800/50 transition-colors"
      >
        <div className="flex items-center gap-3">
          <NodeTypeIcon type={nodeType} />
          <div>
            <div className="flex items-center gap-2">
              <span className="font-medium text-white">{nodeLabel}</span>
              <DiffBadge status={nodeDiff.status} />
            </div>
            <div className="text-xs text-gray-400 mt-0.5">
              {nodeType} • {node.id}
            </div>
          </div>
        </div>
        <svg
          className={`w-4 h-4 text-gray-400 transition-transform ${expanded ? 'rotate-180' : ''}`}
          fill="none"
          viewBox="0 0 24 24"
          stroke="currentColor"
        >
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 9l-7 7-7-7" />
        </svg>
      </button>

      {expanded && (
        <div className="px-3 pb-3 border-t border-gray-700/50">
          {/* Property Changes */}
          {nodeDiff.propertyChanges && nodeDiff.propertyChanges.length > 0 && (
            <PropertyChangeDisplay changes={nodeDiff.propertyChanges} className="mt-3" />
          )}

          {/* Node Data Preview */}
          <div className="mt-3">
            <div className="text-xs font-medium text-gray-400 uppercase mb-2">Node Data</div>
            <pre className="text-xs font-mono p-2 bg-gray-900 rounded overflow-x-auto text-gray-300">
              {JSON.stringify(node.data, null, 2)}
            </pre>
          </div>
        </div>
      )}
    </DiffContainer>
  )
}

interface EdgeDiffCardProps {
  edgeDiff: EdgeDiff
  side: 'base' | 'compare'
}

function EdgeDiffCard({ edgeDiff, side }: EdgeDiffCardProps) {
  const edge = side === 'base' ? edgeDiff.baseEdge : edgeDiff.compareEdge
  const isVisible = side === 'base'
    ? edgeDiff.status !== 'added'
    : edgeDiff.status !== 'removed'

  if (!isVisible) {
    return (
      <div className="p-2 bg-gray-800/30 border border-dashed border-gray-600 rounded">
        <span className="text-gray-500 text-xs italic">
          {edgeDiff.status === 'added' ? 'Connection added' : 'Connection removed'}
        </span>
      </div>
    )
  }

  if (!edge) return null

  const colors = DIFF_COLORS[edgeDiff.status]

  return (
    <div className={`p-2 rounded border ${colors.bg} ${colors.border}`}>
      <div className="flex items-center gap-2 text-sm">
        <span className="text-gray-300 font-mono text-xs">{edge.source}</span>
        <svg className="w-4 h-4 text-gray-500" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M14 5l7 7m0 0l-7 7m7-7H3" />
        </svg>
        <span className="text-gray-300 font-mono text-xs">{edge.target}</span>
        <DiffBadge status={edgeDiff.status} />
      </div>
      {(edge as WorkflowEdge & { label?: string }).label && (
        <div className="text-xs text-gray-500 mt-1">Label: {(edge as WorkflowEdge & { label?: string }).label}</div>
      )}
    </div>
  )
}

interface NodeTypeIconProps {
  type: string
}

function NodeTypeIcon({ type }: NodeTypeIconProps) {
  const iconByType: Record<string, string> = {
    trigger: '⚡',
    action: '▶️',
    ai: '🤖',
    conditional: '🔀',
    loop: '🔄',
    parallel: '⚡',
    fork: '🔱',
    join: '🔗',
    subworkflow: '📦',
    retry: '🔁',
    try: '🛡️',
  }

  return (
    <span className="text-lg" role="img" aria-label={`${type} node`}>
      {iconByType[type] || '📋'}
    </span>
  )
}
