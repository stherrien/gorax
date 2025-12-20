import { useState } from 'react'
import { useTemplates, useTemplateMutations } from '../../hooks/useTemplates'
import type { Template, TemplateCategory } from '../../api/templates'

interface TemplateBrowserProps {
  onSelectTemplate?: (template: Template) => void
  onClose?: () => void
}

const categories: { value: TemplateCategory; label: string }[] = [
  { value: 'security', label: 'Security' },
  { value: 'monitoring', label: 'Monitoring' },
  { value: 'integration', label: 'Integration' },
  { value: 'dataops', label: 'Data Ops' },
  { value: 'devops', label: 'Dev Ops' },
  { value: 'other', label: 'Other' },
]

export function TemplateBrowser({ onSelectTemplate, onClose }: TemplateBrowserProps) {
  const [selectedCategory, setSelectedCategory] = useState<string>('')
  const [searchQuery, setSearchQuery] = useState('')

  const { templates, loading, error } = useTemplates({
    category: selectedCategory || undefined,
    search: searchQuery || undefined,
  })

  const { instantiating } = useTemplateMutations()

  const handleUseTemplate = async (template: Template) => {
    if (onSelectTemplate) {
      onSelectTemplate(template)
    }
  }

  const filteredTemplates = templates

  return (
    <div className="template-browser">
      <div className="template-browser-header">
        <h2>Template Library</h2>
        {onClose && (
          <button onClick={onClose} className="close-button">
            ✕
          </button>
        )}
      </div>

      <div className="template-browser-filters">
        <div className="search-box">
          <input
            type="text"
            placeholder="Search templates..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            className="search-input"
          />
        </div>

        <div className="category-filters">
          <button
            className={`category-button ${!selectedCategory ? 'active' : ''}`}
            onClick={() => setSelectedCategory('')}
          >
            All
          </button>
          {categories.map((cat) => (
            <button
              key={cat.value}
              className={`category-button ${selectedCategory === cat.value ? 'active' : ''}`}
              onClick={() => setSelectedCategory(cat.value)}
            >
              {cat.label}
            </button>
          ))}
        </div>
      </div>

      <div className="template-browser-content">
        {loading && <div className="loading">Loading templates...</div>}

        {error && (
          <div className="error-message">
            Failed to load templates: {error.message}
          </div>
        )}

        {!loading && !error && filteredTemplates.length === 0 && (
          <div className="empty-state">
            <p>No templates found</p>
            <p className="empty-state-hint">
              Try adjusting your filters or search query
            </p>
          </div>
        )}

        {!loading && !error && filteredTemplates.length > 0 && (
          <div className="template-grid">
            {filteredTemplates.map((template) => (
              <TemplateCard
                key={template.id}
                template={template}
                onUse={handleUseTemplate}
                isUsing={instantiating}
              />
            ))}
          </div>
        )}
      </div>
    </div>
  )
}

interface TemplateCardProps {
  template: Template
  onUse: (template: Template) => void
  isUsing: boolean
}

function TemplateCard({ template, onUse, isUsing }: TemplateCardProps) {
  const [showDetails, setShowDetails] = useState(false)

  const nodeCount = template.definition.nodes.length
  const edgeCount = template.definition.edges.length

  return (
    <div className="template-card">
      <div className="template-card-header">
        <h3>{template.name}</h3>
        <span className="template-category">{template.category}</span>
      </div>

      <div className="template-card-body">
        <p className="template-description">{template.description}</p>

        <div className="template-stats">
          <span>{nodeCount} nodes</span>
          <span>{edgeCount} connections</span>
        </div>

        {template.tags && template.tags.length > 0 && (
          <div className="template-tags">
            {template.tags.map((tag) => (
              <span key={tag} className="tag">
                {tag}
              </span>
            ))}
          </div>
        )}

        {template.isPublic && (
          <div className="template-badge">
            <span className="public-badge">Public Template</span>
          </div>
        )}
      </div>

      <div className="template-card-actions">
        <button
          onClick={() => setShowDetails(!showDetails)}
          className="btn-secondary"
        >
          {showDetails ? 'Hide' : 'View'} Details
        </button>
        <button
          onClick={() => onUse(template)}
          disabled={isUsing}
          className="btn-primary"
        >
          {isUsing ? 'Using...' : 'Use Template'}
        </button>
      </div>

      {showDetails && (
        <div className="template-details">
          <h4>Template Structure</h4>
          <div className="template-nodes-list">
            {template.definition.nodes.map((node) => (
              <div key={node.id} className="node-item">
                <span className="node-type">{node.type}</span>
                <span className="node-name">{node.data.name}</span>
              </div>
            ))}
          </div>
        </div>
      )}
    </div>
  )
}
