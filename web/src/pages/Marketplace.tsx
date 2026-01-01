import React, { useState } from 'react'
import { useMarketplace, useMarketplaceCategories, useMarketplaceTemplate } from '../hooks/useMarketplace'
import type { SearchFilter, MarketplaceTemplate } from '../types/marketplace'

export const Marketplace: React.FC = () => {
  const [filter, setFilter] = useState<SearchFilter>({
    limit: 20,
    page: 0,
    sortBy: 'popular',
  })
  const [selectedTemplate, setSelectedTemplate] = useState<string | null>(null)
  const [showInstallModal, setShowInstallModal] = useState(false)
  const [showRatingModal, setShowRatingModal] = useState(false)

  const { templates, loading, error, refresh } = useMarketplace(filter)
  const { categories } = useMarketplaceCategories()

  const handleCategoryChange = (category: string) => {
    setFilter((prev) => ({ ...prev, category: category || undefined, page: 0 }))
  }

  const handleSearchChange = (search: string) => {
    setFilter((prev) => ({ ...prev, searchQuery: search || undefined, page: 0 }))
  }

  const handleSortChange = (sortBy: 'popular' | 'recent' | 'rating') => {
    setFilter((prev) => ({ ...prev, sortBy, page: 0 }))
  }

  const handleTemplateClick = (templateId: string) => {
    setSelectedTemplate(templateId)
  }

  const handleInstallClick = (templateId: string) => {
    setSelectedTemplate(templateId)
    setShowInstallModal(true)
  }

  const handleRateClick = (templateId: string) => {
    setSelectedTemplate(templateId)
    setShowRatingModal(true)
  }

  if (loading && templates.length === 0) {
    return (
      <div className="flex items-center justify-center h-screen">
        <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600"></div>
      </div>
    )
  }

  if (error) {
    return (
      <div className="flex items-center justify-center h-screen">
        <div className="text-red-600">Error loading templates: {error.message}</div>
      </div>
    )
  }

  return (
    <div className="container mx-auto px-4 py-8">
      <div className="mb-8">
        <h1 className="text-3xl font-bold mb-2">Template Marketplace</h1>
        <p className="text-gray-600">Discover and install workflow templates from the community</p>
      </div>

      <div className="flex gap-6">
        {/* Sidebar */}
        <aside className="w-64 flex-shrink-0">
          <div className="bg-white rounded-lg shadow p-4 sticky top-4">
            <h2 className="font-semibold mb-4">Filters</h2>

            {/* Search */}
            <div className="mb-4">
              <label className="block text-sm font-medium text-gray-700 mb-2">Search</label>
              <input
                type="text"
                placeholder="Search templates..."
                className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                onChange={(e) => handleSearchChange(e.target.value)}
              />
            </div>

            {/* Categories */}
            <div className="mb-4">
              <label className="block text-sm font-medium text-gray-700 mb-2">Category</label>
              <select
                className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                value={filter.category || ''}
                onChange={(e) => handleCategoryChange(e.target.value)}
              >
                <option value="">All Categories</option>
                {categories.map((cat) => (
                  <option key={cat} value={cat}>
                    {cat.charAt(0).toUpperCase() + cat.slice(1)}
                  </option>
                ))}
              </select>
            </div>

            {/* Sort */}
            <div className="mb-4">
              <label className="block text-sm font-medium text-gray-700 mb-2">Sort By</label>
              <select
                className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                value={filter.sortBy || 'popular'}
                onChange={(e) => handleSortChange(e.target.value as any)}
              >
                <option value="popular">Most Popular</option>
                <option value="recent">Recently Added</option>
                <option value="rating">Highest Rated</option>
              </select>
            </div>

            {/* Verified Filter */}
            <div className="mb-4">
              <label className="flex items-center">
                <input
                  type="checkbox"
                  className="rounded border-gray-300 text-blue-600 focus:ring-blue-500"
                  checked={filter.isVerified || false}
                  onChange={(e) => setFilter((prev) => ({ ...prev, isVerified: e.target.checked || undefined }))}
                />
                <span className="ml-2 text-sm text-gray-700">Verified Only</span>
              </label>
            </div>
          </div>
        </aside>

        {/* Main Content */}
        <main className="flex-1">
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
            {templates.map((template) => (
              <TemplateCard
                key={template.id}
                template={template}
                onViewDetails={() => handleTemplateClick(template.id)}
                onInstall={() => handleInstallClick(template.id)}
                onRate={() => handleRateClick(template.id)}
              />
            ))}
          </div>

          {templates.length === 0 && (
            <div className="text-center py-12">
              <p className="text-gray-500 text-lg">No templates found</p>
            </div>
          )}
        </main>
      </div>

      {/* Template Detail Modal */}
      {selectedTemplate && !showInstallModal && !showRatingModal && (
        <TemplateDetailModal templateId={selectedTemplate} onClose={() => setSelectedTemplate(null)} />
      )}

      {/* Install Modal */}
      {showInstallModal && selectedTemplate && (
        <InstallModal
          templateId={selectedTemplate}
          onClose={() => {
            setShowInstallModal(false)
            setSelectedTemplate(null)
          }}
          onSuccess={() => {
            setShowInstallModal(false)
            setSelectedTemplate(null)
            refresh()
          }}
        />
      )}

      {/* Rating Modal */}
      {showRatingModal && selectedTemplate && (
        <RatingModal
          templateId={selectedTemplate}
          onClose={() => {
            setShowRatingModal(false)
            setSelectedTemplate(null)
          }}
          onSuccess={() => {
            setShowRatingModal(false)
            setSelectedTemplate(null)
            refresh()
          }}
        />
      )}
    </div>
  )
}

// Template Card Component
const TemplateCard: React.FC<{
  template: MarketplaceTemplate
  onViewDetails: () => void
  onInstall: () => void
  onRate: () => void
}> = ({ template, onViewDetails, onInstall, onRate }) => {
  return (
    <div className="bg-white rounded-lg shadow hover:shadow-lg transition-shadow cursor-pointer" onClick={onViewDetails}>
      <div className="p-6">
        <div className="flex items-start justify-between mb-2">
          <h3 className="text-lg font-semibold text-gray-900 flex-1">{template.name}</h3>
          {template.isVerified && (
            <span className="inline-flex items-center px-2 py-1 rounded-full text-xs font-medium bg-blue-100 text-blue-800">
              Verified
            </span>
          )}
        </div>

        <p className="text-gray-600 text-sm mb-4 line-clamp-2">{template.description}</p>

        <div className="flex items-center gap-2 mb-4">
          <span className="inline-flex items-center px-2 py-1 rounded-full text-xs font-medium bg-gray-100 text-gray-800">
            {template.category}
          </span>
          {template.tags.slice(0, 2).map((tag) => (
            <span key={tag} className="inline-flex items-center px-2 py-1 rounded-full text-xs font-medium bg-gray-100 text-gray-600">
              {tag}
            </span>
          ))}
        </div>

        <div className="flex items-center justify-between text-sm text-gray-500 mb-4">
          <div className="flex items-center gap-1">
            <span className="text-yellow-500">★</span>
            <span>{template.averageRating.toFixed(1)}</span>
            <span>({template.totalRatings})</span>
          </div>
          <div>{template.downloadCount} downloads</div>
        </div>

        <div className="text-xs text-gray-500 mb-4">
          By {template.authorName} • v{template.version}
        </div>

        <div className="flex gap-2">
          <button
            className="flex-1 px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 transition-colors"
            onClick={(e) => {
              e.stopPropagation()
              onInstall()
            }}
          >
            Install
          </button>
          <button
            className="px-4 py-2 border border-gray-300 rounded-md hover:bg-gray-50 transition-colors"
            onClick={(e) => {
              e.stopPropagation()
              onRate()
            }}
          >
            Rate
          </button>
        </div>
      </div>
    </div>
  )
}

// Template Detail Modal Component
const TemplateDetailModal: React.FC<{
  templateId: string
  onClose: () => void
}> = ({ templateId, onClose }) => {
  const { template, loading } = useMarketplaceTemplate(templateId)

  if (loading || !template) {
    return (
      <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
        <div className="bg-white rounded-lg p-8">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600"></div>
        </div>
      </div>
    )
  }

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50 p-4" onClick={onClose}>
      <div className="bg-white rounded-lg max-w-4xl w-full max-h-[90vh] overflow-y-auto" onClick={(e) => e.stopPropagation()}>
        <div className="p-6">
          <div className="flex items-start justify-between mb-4">
            <div>
              <h2 className="text-2xl font-bold">{template.name}</h2>
              <p className="text-gray-600">By {template.authorName} • v{template.version}</p>
            </div>
            <button onClick={onClose} className="text-gray-500 hover:text-gray-700">
              <span className="text-2xl">&times;</span>
            </button>
          </div>

          <div className="mb-6">
            <p className="text-gray-700">{template.description}</p>
          </div>

          <div className="grid grid-cols-2 gap-4 mb-6">
            <div>
              <div className="text-sm text-gray-600">Category</div>
              <div className="font-medium">{template.category}</div>
            </div>
            <div>
              <div className="text-sm text-gray-600">Downloads</div>
              <div className="font-medium">{template.downloadCount}</div>
            </div>
            <div>
              <div className="text-sm text-gray-600">Rating</div>
              <div className="font-medium">
                {template.averageRating.toFixed(1)} / 5.0 ({template.totalRatings} ratings)
              </div>
            </div>
            <div>
              <div className="text-sm text-gray-600">Published</div>
              <div className="font-medium">{new Date(template.publishedAt).toLocaleDateString()}</div>
            </div>
          </div>

          {template.tags.length > 0 && (
            <div className="mb-6">
              <div className="text-sm text-gray-600 mb-2">Tags</div>
              <div className="flex flex-wrap gap-2">
                {template.tags.map((tag) => (
                  <span key={tag} className="inline-flex items-center px-3 py-1 rounded-full text-sm font-medium bg-gray-100 text-gray-800">
                    {tag}
                  </span>
                ))}
              </div>
            </div>
          )}
        </div>
      </div>
    </div>
  )
}

// Install Modal Component
const InstallModal: React.FC<{
  templateId: string
  onClose: () => void
  onSuccess: () => void
}> = ({ templateId, onClose, onSuccess }) => {
  const [workflowName, setWorkflowName] = useState('')
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const { install } = useMarketplaceTemplate(templateId)

  const handleInstall = async () => {
    if (!workflowName.trim()) {
      setError('Workflow name is required')
      return
    }

    try {
      setLoading(true)
      setError(null)
      await install({ workflowName: workflowName.trim() })
      onSuccess()
    } catch (err) {
      setError((err as Error).message)
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50 p-4" onClick={onClose}>
      <div className="bg-white rounded-lg max-w-md w-full" onClick={(e) => e.stopPropagation()}>
        <div className="p-6">
          <h2 className="text-xl font-bold mb-4">Install Template</h2>

          <div className="mb-4">
            <label className="block text-sm font-medium text-gray-700 mb-2">Workflow Name</label>
            <input
              type="text"
              className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
              placeholder="Enter workflow name"
              value={workflowName}
              onChange={(e) => setWorkflowName(e.target.value)}
            />
          </div>

          {error && <div className="mb-4 text-red-600 text-sm">{error}</div>}

          <div className="flex gap-2">
            <button
              className="flex-1 px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 transition-colors disabled:opacity-50"
              onClick={handleInstall}
              disabled={loading}
            >
              {loading ? 'Installing...' : 'Install'}
            </button>
            <button className="px-4 py-2 border border-gray-300 rounded-md hover:bg-gray-50 transition-colors" onClick={onClose}>
              Cancel
            </button>
          </div>
        </div>
      </div>
    </div>
  )
}

// Rating Modal Component
const RatingModal: React.FC<{
  templateId: string
  onClose: () => void
  onSuccess: () => void
}> = ({ templateId, onClose, onSuccess }) => {
  const [rating, setRating] = useState(5)
  const [comment, setComment] = useState('')
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const { rate } = useMarketplaceTemplate(templateId)

  const handleRate = async () => {
    try {
      setLoading(true)
      setError(null)
      await rate({ rating, comment: comment.trim() || undefined })
      onSuccess()
    } catch (err) {
      setError((err as Error).message)
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50 p-4" onClick={onClose}>
      <div className="bg-white rounded-lg max-w-md w-full" onClick={(e) => e.stopPropagation()}>
        <div className="p-6">
          <h2 className="text-xl font-bold mb-4">Rate Template</h2>

          <div className="mb-4">
            <label className="block text-sm font-medium text-gray-700 mb-2">Rating</label>
            <div className="flex gap-2">
              {[1, 2, 3, 4, 5].map((star) => (
                <button
                  key={star}
                  className={`text-3xl ${star <= rating ? 'text-yellow-500' : 'text-gray-300'}`}
                  onClick={() => setRating(star)}
                >
                  ★
                </button>
              ))}
            </div>
          </div>

          <div className="mb-4">
            <label className="block text-sm font-medium text-gray-700 mb-2">Comment (optional)</label>
            <textarea
              className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
              rows={4}
              placeholder="Share your experience with this template"
              value={comment}
              onChange={(e) => setComment(e.target.value)}
            />
          </div>

          {error && <div className="mb-4 text-red-600 text-sm">{error}</div>}

          <div className="flex gap-2">
            <button
              className="flex-1 px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 transition-colors disabled:opacity-50"
              onClick={handleRate}
              disabled={loading}
            >
              {loading ? 'Submitting...' : 'Submit Rating'}
            </button>
            <button className="px-4 py-2 border border-gray-300 rounded-md hover:bg-gray-50 transition-colors" onClick={onClose}>
              Cancel
            </button>
          </div>
        </div>
      </div>
    </div>
  )
}

export default Marketplace
