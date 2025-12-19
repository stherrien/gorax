import React, { useState } from 'react'
import { useTemplateMutations } from '../../hooks/useTemplates'
import type { TemplateCategory, WorkflowDefinition } from '../../api/templates'

interface SaveAsTemplateProps {
  workflowId: string
  workflowName: string
  definition: WorkflowDefinition
  onSuccess?: () => void
  onCancel?: () => void
}

const categories: { value: TemplateCategory; label: string }[] = [
  { value: 'security', label: 'Security' },
  { value: 'monitoring', label: 'Monitoring' },
  { value: 'integration', label: 'Integration' },
  { value: 'dataops', label: 'Data Ops' },
  { value: 'devops', label: 'Dev Ops' },
  { value: 'other', label: 'Other' },
]

export function SaveAsTemplate({
  workflowId,
  workflowName,
  definition,
  onSuccess,
  onCancel,
}: SaveAsTemplateProps) {
  const [name, setName] = useState(workflowName)
  const [description, setDescription] = useState('')
  const [category, setCategory] = useState<TemplateCategory>('other')
  const [tags, setTags] = useState<string[]>([])
  const [tagInput, setTagInput] = useState('')
  const [isPublic, setIsPublic] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const { createFromWorkflow, creating } = useTemplateMutations()

  const handleAddTag = () => {
    const trimmed = tagInput.trim()
    if (trimmed && !tags.includes(trimmed)) {
      setTags([...tags, trimmed])
      setTagInput('')
    }
  }

  const handleRemoveTag = (tagToRemove: string) => {
    setTags(tags.filter((tag) => tag !== tagToRemove))
  }

  const handleKeyPress = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter') {
      e.preventDefault()
      handleAddTag()
    }
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setError(null)

    if (!name.trim()) {
      setError('Template name is required')
      return
    }

    if (!category) {
      setError('Please select a category')
      return
    }

    try {
      await createFromWorkflow(workflowId, {
        name: name.trim(),
        description: description.trim(),
        category,
        definition,
        tags,
        isPublic,
      })

      if (onSuccess) {
        onSuccess()
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to save template')
    }
  }

  return (
    <div className="save-template-dialog">
      <div className="dialog-header">
        <h2>Save as Template</h2>
        {onCancel && (
          <button onClick={onCancel} className="close-button">
            ✕
          </button>
        )}
      </div>

      <form onSubmit={handleSubmit} className="save-template-form">
        {error && (
          <div className="error-message" role="alert">
            {error}
          </div>
        )}

        <div className="form-group">
          <label htmlFor="template-name">
            Template Name <span className="required">*</span>
          </label>
          <input
            id="template-name"
            type="text"
            value={name}
            onChange={(e) => setName(e.target.value)}
            placeholder="Enter template name"
            required
            className="form-input"
          />
        </div>

        <div className="form-group">
          <label htmlFor="template-description">Description</label>
          <textarea
            id="template-description"
            value={description}
            onChange={(e) => setDescription(e.target.value)}
            placeholder="Describe what this template does"
            rows={3}
            className="form-textarea"
          />
        </div>

        <div className="form-group">
          <label htmlFor="template-category">
            Category <span className="required">*</span>
          </label>
          <select
            id="template-category"
            value={category}
            onChange={(e) => setCategory(e.target.value as TemplateCategory)}
            required
            className="form-select"
          >
            {categories.map((cat) => (
              <option key={cat.value} value={cat.value}>
                {cat.label}
              </option>
            ))}
          </select>
        </div>

        <div className="form-group">
          <label htmlFor="template-tags">Tags</label>
          <div className="tags-input-wrapper">
            <div className="tags-list">
              {tags.map((tag) => (
                <span key={tag} className="tag">
                  {tag}
                  <button
                    type="button"
                    onClick={() => handleRemoveTag(tag)}
                    className="tag-remove"
                    aria-label={`Remove ${tag} tag`}
                  >
                    ✕
                  </button>
                </span>
              ))}
            </div>
            <div className="tag-input-group">
              <input
                id="template-tags"
                type="text"
                value={tagInput}
                onChange={(e) => setTagInput(e.target.value)}
                onKeyPress={handleKeyPress}
                placeholder="Add a tag"
                className="form-input"
              />
              <button
                type="button"
                onClick={handleAddTag}
                className="btn-secondary"
                disabled={!tagInput.trim()}
              >
                Add
              </button>
            </div>
          </div>
        </div>

        <div className="form-group">
          <label className="checkbox-label">
            <input
              type="checkbox"
              checked={isPublic}
              onChange={(e) => setIsPublic(e.target.checked)}
              className="form-checkbox"
            />
            <span>Make this template public</span>
          </label>
          <p className="help-text">
            Public templates can be used by all tenants in the system
          </p>
        </div>

        <div className="template-preview">
          <h3>Template Preview</h3>
          <div className="preview-stats">
            <div className="stat">
              <span className="stat-label">Nodes:</span>
              <span className="stat-value">{definition.nodes.length}</span>
            </div>
            <div className="stat">
              <span className="stat-label">Connections:</span>
              <span className="stat-value">{definition.edges.length}</span>
            </div>
          </div>
        </div>

        <div className="dialog-actions">
          {onCancel && (
            <button
              type="button"
              onClick={onCancel}
              className="btn-secondary"
              disabled={creating}
            >
              Cancel
            </button>
          )}
          <button type="submit" className="btn-primary" disabled={creating}>
            {creating ? 'Saving...' : 'Save Template'}
          </button>
        </div>
      </form>
    </div>
  )
}
