import React from 'react'
import { useCategories } from '../../hooks/useMarketplace'
import type { Category } from '../../types/marketplace'

export interface CategoryFilterProps {
  selectedCategories: string[]
  onChange: (categoryIds: string[]) => void
  className?: string
}

export const CategoryFilter: React.FC<CategoryFilterProps> = ({ selectedCategories, onChange, className = '' }) => {
  const { categories, loading } = useCategories()

  const handleToggle = (categoryId: string) => {
    if (selectedCategories.includes(categoryId)) {
      onChange(selectedCategories.filter((id) => id !== categoryId))
    } else {
      onChange([...selectedCategories, categoryId])
    }
  }

  const handleClear = () => {
    onChange([])
  }

  if (loading) {
    return (
      <div className={`p-4 ${className}`}>
        <div className="animate-pulse">
          <div className="h-4 bg-gray-200 rounded w-3/4 mb-3" />
          <div className="h-4 bg-gray-200 rounded w-1/2" />
        </div>
      </div>
    )
  }

  const rootCategories = categories.filter((cat) => !cat.parentId).sort((a, b) => a.displayOrder - b.displayOrder)

  return (
    <div className={className}>
      <div className="flex items-center justify-between mb-3">
        <h3 className="font-semibold text-gray-900">Categories</h3>
        {selectedCategories.length > 0 && (
          <button onClick={handleClear} className="text-sm text-blue-600 hover:text-blue-700">
            Clear all
          </button>
        )}
      </div>

      <div className="space-y-2">
        {rootCategories.map((category) => (
          <label key={category.id} className="flex items-center gap-2 cursor-pointer group">
            <input
              type="checkbox"
              checked={selectedCategories.includes(category.id)}
              onChange={() => handleToggle(category.id)}
              className="w-4 h-4 rounded border-gray-300 text-blue-600 focus:ring-blue-500"
            />
            <span className="text-sm text-gray-700 group-hover:text-gray-900 flex-1">{category.name}</span>
            <span className="text-xs text-gray-500">({category.templateCount})</span>
          </label>
        ))}
      </div>
    </div>
  )
}
