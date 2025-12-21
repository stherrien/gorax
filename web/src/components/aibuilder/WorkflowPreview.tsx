import { FC, useMemo } from 'react'
import type { GeneratedWorkflow, GeneratedNode, GeneratedEdge } from '../../types/aibuilder'

interface WorkflowPreviewProps {
  workflow: GeneratedWorkflow
  onNodeClick?: (node: GeneratedNode) => void
  onOpenEditor?: () => void
  className?: string
}

// Node type to icon mapping
const nodeIcons: Record<string, string> = {
  'trigger:webhook': '🔗',
  'trigger:schedule': '⏰',
  'action:http': '🌐',
  'action:transform': '🔄',
  'action:formula': '🧮',
  'action:code': '💻',
  'action:email': '📧',
  'control:if': '🔀',
  'control:loop': '🔁',
  'control:parallel': '⚡',
  'control:delay': '⏳',
  'integration:slack': '💬',
  'integration:jira': '📋',
  'integration:github': '🐙',
}

// Node type to color mapping
const nodeColors: Record<string, { bg: string; border: string; text: string }> = {
  trigger: { bg: 'bg-green-50', border: 'border-green-300', text: 'text-green-700' },
  action: { bg: 'bg-blue-50', border: 'border-blue-300', text: 'text-blue-700' },
  control: { bg: 'bg-purple-50', border: 'border-purple-300', text: 'text-purple-700' },
  integration: { bg: 'bg-orange-50', border: 'border-orange-300', text: 'text-orange-700' },
}

function getNodeCategory(nodeType: string): string {
  const category = nodeType.split(':')[0]
  return category in nodeColors ? category : 'action'
}

function getNodeIcon(nodeType: string): string {
  return nodeIcons[nodeType] || '📦'
}

export const WorkflowPreview: FC<WorkflowPreviewProps> = ({
  workflow,
  onNodeClick,
  onOpenEditor,
  className = '',
}) => {
  const { nodes, edges } = workflow.definition

  // Calculate SVG dimensions based on node positions
  const dimensions = useMemo(() => {
    if (nodes.length === 0) return { width: 400, height: 200 }

    const positions = nodes.map(n => n.position || { x: 0, y: 0 })
    const maxX = Math.max(...positions.map(p => p.x)) + 200
    const maxY = Math.max(...positions.map(p => p.y)) + 100

    return {
      width: Math.max(400, maxX),
      height: Math.max(200, maxY),
    }
  }, [nodes])

  // Build node position map for edge rendering
  const nodePositions = useMemo(() => {
    const map = new Map<string, { x: number; y: number }>()
    nodes.forEach(node => {
      map.set(node.id, node.position || { x: 0, y: 0 })
    })
    return map
  }, [nodes])

  return (
    <div className={`rounded-lg border border-gray-200 bg-white ${className}`}>
      {/* Header */}
      <div className="flex items-center justify-between border-b border-gray-200 px-4 py-3">
        <div>
          <h3 className="font-medium text-gray-900">{workflow.name}</h3>
          {workflow.description && (
            <p className="text-sm text-gray-500">{workflow.description}</p>
          )}
        </div>
        <div className="flex items-center space-x-2">
          <span className="rounded bg-gray-100 px-2 py-1 text-xs text-gray-600">
            {nodes.length} nodes
          </span>
          {onOpenEditor && (
            <button
              onClick={onOpenEditor}
              className="rounded bg-blue-100 px-3 py-1 text-xs font-medium text-blue-700 hover:bg-blue-200"
            >
              Open in Editor
            </button>
          )}
        </div>
      </div>

      {/* Canvas */}
      <div className="relative overflow-auto p-4" style={{ maxHeight: '400px' }}>
        <svg
          width={dimensions.width}
          height={dimensions.height}
          className="overflow-visible"
        >
          {/* Render edges */}
          {edges?.map((edge) => (
            <EdgeLine
              key={edge.id}
              edge={edge}
              nodePositions={nodePositions}
            />
          ))}
        </svg>

        {/* Render nodes as absolute positioned divs */}
        {nodes.map((node) => (
          <NodeCard
            key={node.id}
            node={node}
            onClick={() => onNodeClick?.(node)}
          />
        ))}
      </div>

      {/* Legend */}
      <div className="border-t border-gray-200 px-4 py-2">
        <div className="flex flex-wrap items-center gap-4 text-xs">
          <LegendItem color="green" label="Trigger" />
          <LegendItem color="blue" label="Action" />
          <LegendItem color="purple" label="Control" />
          <LegendItem color="orange" label="Integration" />
        </div>
      </div>
    </div>
  )
}

// Node card component
interface NodeCardProps {
  node: GeneratedNode
  onClick?: () => void
}

const NodeCard: FC<NodeCardProps> = ({ node, onClick }) => {
  const category = getNodeCategory(node.type)
  const colors = nodeColors[category]
  const icon = getNodeIcon(node.type)
  const position = node.position || { x: 0, y: 0 }

  return (
    <div
      className={`absolute cursor-pointer rounded-lg border-2 p-3 shadow-sm transition-shadow hover:shadow-md ${colors.bg} ${colors.border}`}
      style={{
        left: position.x,
        top: position.y,
        width: '160px',
      }}
      onClick={onClick}
    >
      <div className="flex items-center space-x-2">
        <span className="text-lg">{icon}</span>
        <div className="min-w-0 flex-1">
          <div className={`truncate text-sm font-medium ${colors.text}`}>
            {node.name}
          </div>
          <div className="truncate text-xs text-gray-500">
            {node.type.split(':')[1] || node.type}
          </div>
        </div>
      </div>
      {node.description && (
        <p className="mt-1 truncate text-xs text-gray-500" title={node.description}>
          {node.description}
        </p>
      )}
    </div>
  )
}

// Edge line component
interface EdgeLineProps {
  edge: GeneratedEdge
  nodePositions: Map<string, { x: number; y: number }>
}

const EdgeLine: FC<EdgeLineProps> = ({ edge, nodePositions }) => {
  const sourcePos = nodePositions.get(edge.source)
  const targetPos = nodePositions.get(edge.target)

  if (!sourcePos || !targetPos) return null

  // Calculate edge path (simple straight line with offset for node size)
  const nodeWidth = 160
  const nodeHeight = 60

  const startX = sourcePos.x + nodeWidth / 2
  const startY = sourcePos.y + nodeHeight
  const endX = targetPos.x + nodeWidth / 2
  const endY = targetPos.y

  // Create a curved path
  const midY = (startY + endY) / 2
  const path = `M ${startX} ${startY} C ${startX} ${midY}, ${endX} ${midY}, ${endX} ${endY}`

  return (
    <g>
      <path
        d={path}
        fill="none"
        stroke="#9CA3AF"
        strokeWidth={2}
        markerEnd="url(#arrowhead)"
      />
      {edge.label && (
        <text
          x={(startX + endX) / 2}
          y={midY - 5}
          textAnchor="middle"
          className="fill-gray-500 text-xs"
        >
          {edge.label}
        </text>
      )}
      {/* Arrow marker definition */}
      <defs>
        <marker
          id="arrowhead"
          markerWidth="10"
          markerHeight="7"
          refX="9"
          refY="3.5"
          orient="auto"
        >
          <polygon points="0 0, 10 3.5, 0 7" fill="#9CA3AF" />
        </marker>
      </defs>
    </g>
  )
}

// Legend item component
interface LegendItemProps {
  color: 'green' | 'blue' | 'purple' | 'orange'
  label: string
}

const LegendItem: FC<LegendItemProps> = ({ color, label }) => {
  const colorClasses: Record<string, string> = {
    green: 'bg-green-200',
    blue: 'bg-blue-200',
    purple: 'bg-purple-200',
    orange: 'bg-orange-200',
  }

  return (
    <div className="flex items-center space-x-1">
      <div className={`h-3 w-3 rounded ${colorClasses[color]}`} />
      <span className="text-gray-600">{label}</span>
    </div>
  )
}

export default WorkflowPreview
