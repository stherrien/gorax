import { useState, useMemo } from 'react'

interface NodeDefinition {
  type: 'trigger' | 'action' | 'ai' | 'control'
  nodeType: string
  label: string
  description: string
  icon: string
}

interface NodePaletteProps {
  onAddNode: (node: { type: string; nodeType: string }) => void
}

const NODE_DEFINITIONS: NodeDefinition[] = [
  // Triggers
  {
    type: 'trigger',
    nodeType: 'webhook',
    label: 'Webhook',
    description: 'Trigger workflow via HTTP request',
    icon: '🔗',
  },
  {
    type: 'trigger',
    nodeType: 'schedule',
    label: 'Schedule',
    description: 'Trigger workflow on a schedule',
    icon: '⏰',
  },
  {
    type: 'trigger',
    nodeType: 'manual',
    label: 'Manual',
    description: 'Trigger workflow manually',
    icon: '👆',
  },

  // Actions
  {
    type: 'action',
    nodeType: 'http',
    label: 'HTTP Request',
    description: 'Make HTTP requests to external APIs',
    icon: '🌐',
  },
  {
    type: 'action',
    nodeType: 'transform',
    label: 'Transform',
    description: 'Transform and manipulate data',
    icon: '🔄',
  },
  {
    type: 'action',
    nodeType: 'email',
    label: 'Email',
    description: 'Send email notifications',
    icon: '📧',
  },
  {
    type: 'action',
    nodeType: 'script',
    label: 'Run Script',
    description: 'Execute custom JavaScript code',
    icon: '📜',
  },
  {
    type: 'action',
    nodeType: 'slack_send_message',
    label: 'Slack: Send Message',
    description: 'Send a message to a Slack channel',
    icon: '💬',
  },
  {
    type: 'action',
    nodeType: 'slack_send_dm',
    label: 'Slack: Send DM',
    description: 'Send a direct message to a Slack user',
    icon: '✉️',
  },
  {
    type: 'action',
    nodeType: 'slack_update_message',
    label: 'Slack: Update Message',
    description: 'Update an existing Slack message',
    icon: '✏️',
  },
  {
    type: 'action',
    nodeType: 'slack_add_reaction',
    label: 'Slack: Add Reaction',
    description: 'Add an emoji reaction to a message',
    icon: '👍',
  },

  // AI Actions
  {
    type: 'ai',
    nodeType: 'ai_chat',
    label: 'AI: Chat Completion',
    description: 'Generate AI responses with LLM',
    icon: '🤖',
  },
  {
    type: 'ai',
    nodeType: 'ai_summarize',
    label: 'AI: Summarize',
    description: 'Summarize text using AI',
    icon: '📝',
  },
  {
    type: 'ai',
    nodeType: 'ai_classify',
    label: 'AI: Classify',
    description: 'Classify text into categories',
    icon: '🏷️',
  },
  {
    type: 'ai',
    nodeType: 'ai_extract',
    label: 'AI: Extract Entities',
    description: 'Extract named entities from text',
    icon: '🔍',
  },
  {
    type: 'ai',
    nodeType: 'ai_embed',
    label: 'AI: Generate Embeddings',
    description: 'Create vector embeddings for text',
    icon: '📊',
  },

  // Controls
  {
    type: 'control',
    nodeType: 'conditional',
    label: 'Conditional',
    description: 'Branch based on conditions',
    icon: '🔀',
  },
  {
    type: 'control',
    nodeType: 'loop',
    label: 'Loop',
    description: 'Iterate over arrays',
    icon: '🔁',
  },
  {
    type: 'control',
    nodeType: 'parallel',
    label: 'Parallel',
    description: 'Execute branches concurrently',
    icon: '⚡',
  },
  {
    type: 'control',
    nodeType: 'delay',
    label: 'Delay',
    description: 'Wait for a specified time',
    icon: '⏸️',
  },
]

export default function NodePalette({ onAddNode }: NodePaletteProps) {
  const [searchQuery, setSearchQuery] = useState('')
  const [expandedSections, setExpandedSections] = useState({
    triggers: true,
    actions: true,
    ai: true,
    controls: true,
  })
  const [hoveredNode, setHoveredNode] = useState<string | null>(null)

  const filteredNodes = useMemo(() => {
    if (!searchQuery) return NODE_DEFINITIONS

    const query = searchQuery.toLowerCase()
    return NODE_DEFINITIONS.filter(
      (node) =>
        node.label.toLowerCase().includes(query) ||
        node.description.toLowerCase().includes(query) ||
        node.nodeType.toLowerCase().includes(query)
    )
  }, [searchQuery])

  const triggerNodes = filteredNodes.filter((n) => n.type === 'trigger')
  const actionNodes = filteredNodes.filter((n) => n.type === 'action')
  const aiNodes = filteredNodes.filter((n) => n.type === 'ai')
  const controlNodes = filteredNodes.filter((n) => n.type === 'control')

  const toggleSection = (section: 'triggers' | 'actions' | 'ai' | 'controls') => {
    setExpandedSections((prev) => ({
      ...prev,
      [section]: !prev[section],
    }))
  }

  const handleNodeClick = (node: NodeDefinition) => {
    onAddNode({
      type: node.type,
      nodeType: node.nodeType,
    })
  }

  const handleDragStart = (event: React.DragEvent, node: NodeDefinition) => {
    event.dataTransfer.setData('application/reactflow', JSON.stringify(node))
    event.dataTransfer.effectAllowed = 'move'
  }

  return (
    <div className="w-64 bg-gray-800 border-r border-gray-700 flex flex-col h-full">
      {/* Header */}
      <div className="p-4 border-b border-gray-700">
        <h2 className="text-white font-semibold text-lg mb-3">Node Palette</h2>

        {/* Search */}
        <input
          type="text"
          placeholder="Search nodes..."
          value={searchQuery}
          onChange={(e) => setSearchQuery(e.target.value)}
          className="w-full px-3 py-2 bg-gray-700 text-white rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-primary-500 placeholder-gray-500"
        />
      </div>

      {/* Node List */}
      <div className="flex-1 overflow-y-auto">
        {filteredNodes.length === 0 ? (
          <div className="p-4 text-center text-gray-400 text-sm">No nodes found</div>
        ) : (
          <>
            {/* Triggers Section */}
            {triggerNodes.length > 0 && (
              <NodeSection
                title="Triggers"
                nodes={triggerNodes}
                expanded={expandedSections.triggers}
                onToggle={() => toggleSection('triggers')}
                onNodeClick={handleNodeClick}
                onDragStart={handleDragStart}
                hoveredNode={hoveredNode}
                onNodeHover={setHoveredNode}
              />
            )}

            {/* Actions Section */}
            {actionNodes.length > 0 && (
              <NodeSection
                title="Actions"
                nodes={actionNodes}
                expanded={expandedSections.actions}
                onToggle={() => toggleSection('actions')}
                onNodeClick={handleNodeClick}
                onDragStart={handleDragStart}
                hoveredNode={hoveredNode}
                onNodeHover={setHoveredNode}
              />
            )}

            {/* AI Section */}
            {aiNodes.length > 0 && (
              <NodeSection
                title="AI"
                nodes={aiNodes}
                expanded={expandedSections.ai}
                onToggle={() => toggleSection('ai')}
                onNodeClick={handleNodeClick}
                onDragStart={handleDragStart}
                hoveredNode={hoveredNode}
                onNodeHover={setHoveredNode}
              />
            )}

            {/* Controls Section */}
            {controlNodes.length > 0 && (
              <NodeSection
                title="Controls"
                nodes={controlNodes}
                expanded={expandedSections.controls}
                onToggle={() => toggleSection('controls')}
                onNodeClick={handleNodeClick}
                onDragStart={handleDragStart}
                hoveredNode={hoveredNode}
                onNodeHover={setHoveredNode}
              />
            )}
          </>
        )}
      </div>
    </div>
  )
}

interface NodeSectionProps {
  title: string
  nodes: NodeDefinition[]
  expanded: boolean
  onToggle: () => void
  onNodeClick: (node: NodeDefinition) => void
  onDragStart: (event: React.DragEvent, node: NodeDefinition) => void
  hoveredNode: string | null
  onNodeHover: (nodeType: string | null) => void
}

function NodeSection({
  title,
  nodes,
  expanded,
  onToggle,
  onNodeClick,
  onDragStart,
  hoveredNode,
  onNodeHover,
}: NodeSectionProps) {
  return (
    <div className="border-b border-gray-700">
      {/* Section Header */}
      <button
        onClick={onToggle}
        className="w-full px-4 py-3 flex items-center justify-between text-gray-300 hover:bg-gray-700/50 transition-colors"
      >
        <span className="font-medium text-sm">{title}</span>
        <span className="text-xs">{expanded ? '▼' : '▶'}</span>
      </button>

      {/* Section Content */}
      {expanded && (
        <div className="pb-2">
          {nodes.map((node) => (
            <div
              key={node.nodeType}
              draggable
              onDragStart={(e) => onDragStart(e, node)}
              onClick={() => onNodeClick(node)}
              onMouseEnter={() => onNodeHover(node.nodeType)}
              onMouseLeave={() => onNodeHover(null)}
              className="mx-2 mb-2 p-3 bg-gray-700 rounded-lg cursor-move hover:bg-gray-600 transition-colors relative"
            >
              <div className="flex items-center space-x-3">
                <span className="text-2xl">{node.icon}</span>
                <div className="flex-1">
                  <div className="text-white text-sm font-medium">{node.label}</div>
                  {hoveredNode === node.nodeType && (
                    <div className="text-gray-400 text-xs mt-1">{node.description}</div>
                  )}
                </div>
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  )
}
