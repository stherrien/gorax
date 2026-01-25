import React, { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { FeaturedTemplates, CategoryList, CategoryFilter, TemplateCard } from '../components/marketplace'
import { useMarketplace, useMarketplaceTemplate } from '../hooks/useMarketplace'
import type { SearchFilter, MarketplaceTemplate, Category } from '../types/marketplace'

export const Marketplace: React.FC = () => {
  const navigate = useNavigate()
  const [filter, setFilter] = useState<SearchFilter>({
    limit: 20,
    page: 0,
    sortBy: 'popular',
  })
  const [searchQuery, setSearchQuery] = useState('')
  const [selectedCategories, setSelectedCategories] = useState<string[]>([])
  const [showCategoryBrowse, setShowCategoryBrowse] = useState(true)
  const [showInstallModal, setShowInstallModal] = useState(false)
  const [selectedTemplate, setSelectedTemplate] = useState<MarketplaceTemplate | null>(null)
  const [workflowName, setWorkflowName] = useState('')

  const { templates, loading, error, refresh } = useMarketplace({
    ...filter,
    searchQuery: searchQuery || undefined,
    category: selectedCategories[0] || undefined,
  })

  const { install } = useMarketplaceTemplate(selectedTemplate?.id || '')

  const handleSearchChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setSearchQuery(e.target.value)
    setFilter((prev) => ({ ...prev, page: 0 }))
    setShowCategoryBrowse(false)
  }

  const handleSearchSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    setShowCategoryBrowse(false)
  }

  const handleCategoryClick = (category: Category) => {
    setSelectedCategories([category.id])
    setFilter((prev) => ({ ...prev, page: 0 }))
    setShowCategoryBrowse(false)
  }

  const handleSortChange = (sortBy: 'popular' | 'recent' | 'rating') => {
    setFilter((prev) => ({ ...prev, sortBy, page: 0 }))
  }

  const handleTemplateClick = (template: MarketplaceTemplate) => {
    navigate(`/marketplace/templates/${template.id}`)
  }

  const handleInstallClick = (template: MarketplaceTemplate) => {
    setSelectedTemplate(template)
    setShowInstallModal(true)
  }

  const handleInstall = async () => {
    if (!workflowName.trim()) {
      alert('Please enter a workflow name')
      return
    }

    try {
      await install({ workflowName: workflowName.trim() })
      setShowInstallModal(false)
      setSelectedTemplate(null)
      setWorkflowName('')
      alert('Template installed successfully!')
      refresh()
    } catch (err) {
      alert('Failed to install template')
    }
  }

  const handleCategoryFilterChange = (categoryIds: string[]) => {
    setSelectedCategories(categoryIds)
    setFilter((prev) => ({ ...prev, page: 0 }))
  }

  const handleClearFilters = () => {
    setSearchQuery('')
    setSelectedCategories([])
    setFilter((prev) => ({ ...prev, category: undefined, page: 0 }))
    setShowCategoryBrowse(true)
  }

  return (
    <div className="min-h-screen bg-gray-50">
      {/* Hero Section with Search */}
      <div className="bg-white border-b border-gray-200">
        <div className="container mx-auto px-4 py-8">
          <h1 className="text-4xl font-bold text-gray-900 mb-2">Template Marketplace</h1>
          <p className="text-lg text-gray-600 mb-6">
            Discover and install workflow templates created by the community
          </p>

          {/* Search Bar */}
          <form onSubmit={handleSearchSubmit} className="max-w-2xl">
            <div className="relative">
              <svg
                className="absolute left-3 top-1/2 -translate-y-1/2 w-5 h-5 text-gray-400"
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
              >
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  strokeWidth={2}
                  d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z"
                />
              </svg>
              <input
                type="text"
                placeholder="Search templates..."
                className="w-full pl-10 pr-4 py-3 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                value={searchQuery}
                onChange={handleSearchChange}
              />
            </div>
          </form>
        </div>
      </div>

      <div className="container mx-auto px-4 py-8">
        {/* Featured Templates */}
        {showCategoryBrowse && (
          <div className="mb-12">
            <FeaturedTemplates onTemplateClick={handleTemplateClick} />
          </div>
        )}

        {/* Category Browse */}
        {showCategoryBrowse && (
          <div className="mb-12">
            <CategoryList onCategoryClick={handleCategoryClick} />
          </div>
        )}

        {/* Templates Grid with Filters */}
        {!showCategoryBrowse && (
          <div className="flex gap-6">
            {/* Sidebar Filters */}
            <aside className="w-64 flex-shrink-0">
              <div className="bg-white rounded-lg shadow p-4 sticky top-4">
                <div className="flex items-center justify-between mb-4">
                  <h2 className="font-semibold">Filters</h2>
                  {(searchQuery || selectedCategories.length > 0) && (
                    <button onClick={handleClearFilters} className="text-sm text-blue-600 hover:text-blue-700">
                      Clear all
                    </button>
                  )}
                </div>

                {/* Category Filter */}
                <CategoryFilter selectedCategories={selectedCategories} onChange={handleCategoryFilterChange} />

                {/* Sort */}
                <div className="mt-6">
                  <h3 className="font-semibold text-gray-900 mb-3">Sort By</h3>
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

                {/* Featured Filter */}
                <div className="mt-6">
                  <label className="flex items-center cursor-pointer">
                    <input
                      type="checkbox"
                      className="w-4 h-4 rounded border-gray-300 text-blue-600 focus:ring-blue-500"
                      checked={filter.isFeatured || false}
                      onChange={(e) =>
                        setFilter((prev) => ({ ...prev, isFeatured: e.target.checked || undefined, page: 0 }))
                      }
                    />
                    <span className="ml-2 text-sm text-gray-700">Featured Only</span>
                  </label>
                </div>

                {/* Verified Filter */}
                <div className="mt-3">
                  <label className="flex items-center cursor-pointer">
                    <input
                      type="checkbox"
                      className="w-4 h-4 rounded border-gray-300 text-blue-600 focus:ring-blue-500"
                      checked={filter.isVerified || false}
                      onChange={(e) =>
                        setFilter((prev) => ({ ...prev, isVerified: e.target.checked || undefined, page: 0 }))
                      }
                    />
                    <span className="ml-2 text-sm text-gray-700">Verified Only</span>
                  </label>
                </div>
              </div>
            </aside>

            {/* Main Content */}
            <main className="flex-1">
              {/* Results Header */}
              <div className="mb-6">
                <h2 className="text-2xl font-bold text-gray-900">
                  {searchQuery ? `Results for "${searchQuery}"` : 'All Templates'}
                </h2>
                <p className="text-gray-600 mt-1">{templates.length} templates found</p>
              </div>

              {/* Loading State */}
              {loading && templates.length === 0 && (
                <div className="flex items-center justify-center py-12">
                  <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600" />
                </div>
              )}

              {/* Error State */}
              {error && (
                <div className="text-center py-12">
                  <p className="text-red-600">Error loading templates: {error.message}</p>
                  <button
                    onClick={refresh}
                    className="mt-4 px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 transition-colors"
                  >
                    Retry
                  </button>
                </div>
              )}

              {/* Empty State */}
              {!loading && !error && templates.length === 0 && (
                <div className="text-center py-12">
                  <svg className="w-16 h-16 text-gray-400 mx-auto mb-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path
                      strokeLinecap="round"
                      strokeLinejoin="round"
                      strokeWidth={2}
                      d="M20 13V6a2 2 0 00-2-2H6a2 2 0 00-2 2v7m16 0v5a2 2 0 01-2 2H6a2 2 0 01-2-2v-5m16 0h-2.586a1 1 0 00-.707.293l-2.414 2.414a1 1 0 01-.707.293h-3.172a1 1 0 01-.707-.293l-2.414-2.414A1 1 0 006.586 13H4"
                    />
                  </svg>
                  <p className="text-gray-500 text-lg mb-2">No templates found</p>
                  <p className="text-gray-400 mb-4">Try adjusting your filters or search query</p>
                  <button
                    onClick={handleClearFilters}
                    className="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 transition-colors"
                  >
                    Clear Filters
                  </button>
                </div>
              )}

              {/* Templates Grid */}
              {!loading && !error && templates.length > 0 && (
                <>
                  <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
                    {templates.map((template) => (
                      <TemplateCard
                        key={template.id}
                        template={template}
                        onViewDetails={handleTemplateClick}
                        onInstall={handleInstallClick}
                      />
                    ))}
                  </div>

                  {/* Pagination */}
                  {templates.length === filter.limit && (
                    <div className="flex justify-center gap-2 mt-8">
                      <button
                        onClick={() => setFilter((prev) => ({ ...prev, page: Math.max(0, (prev.page || 0) - 1) }))}
                        disabled={(filter.page || 0) === 0}
                        className="px-4 py-2 border border-gray-300 rounded-md hover:bg-gray-50 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
                      >
                        Previous
                      </button>
                      <span className="px-4 py-2 text-gray-700">Page {(filter.page || 0) + 1}</span>
                      <button
                        onClick={() => setFilter((prev) => ({ ...prev, page: (prev.page || 0) + 1 }))}
                        className="px-4 py-2 border border-gray-300 rounded-md hover:bg-gray-50 transition-colors"
                      >
                        Next
                      </button>
                    </div>
                  )}
                </>
              )}
            </main>
          </div>
        )}
      </div>

      {/* Install Modal */}
      {showInstallModal && selectedTemplate && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50 p-4">
          <div className="bg-white rounded-lg max-w-md w-full p-6">
            <h2 className="text-xl font-bold mb-2">Install Template</h2>
            <p className="text-gray-600 mb-4">Installing: {selectedTemplate.name}</p>
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
            <div className="flex gap-2">
              <button
                onClick={handleInstall}
                className="flex-1 px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 transition-colors"
              >
                Install
              </button>
              <button
                onClick={() => {
                  setShowInstallModal(false)
                  setSelectedTemplate(null)
                  setWorkflowName('')
                }}
                className="px-4 py-2 border border-gray-300 rounded-md hover:bg-gray-50 transition-colors"
              >
                Cancel
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}

export default Marketplace
