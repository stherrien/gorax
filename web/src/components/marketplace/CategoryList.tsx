import React from 'react'
import { CategoryCard } from './CategoryCard'
import { useCategories } from '../../hooks/useMarketplace'
import type { Category } from '../../types/marketplace'

export interface CategoryListProps {
  onCategoryClick?: (category: Category) => void
  className?: string
}

export const CategoryList: React.FC<CategoryListProps> = ({ onCategoryClick, className = '' }) => {
  const { categories, loading, error } = useCategories()

  if (loading) {
    return (
      <div className={`flex items-center justify-center py-12 ${className}`}>
        <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600" />
      </div>
    )
  }

  if (error) {
    return (
      <div className={`text-center py-12 ${className}`}>
        <p className="text-red-600">Failed to load categories: {error.message}</p>
      </div>
    )
  }

  if (categories.length === 0) {
    return (
      <div className={`text-center py-12 ${className}`}>
        <p className="text-gray-500">No categories available</p>
      </div>
    )
  }

  // Filter to show only root categories (no parent)
  const rootCategories = categories.filter((cat) => !cat.parentId).sort((a, b) => a.displayOrder - b.displayOrder)

  return (
    <div className={className}>
      <h2 className="text-2xl font-bold mb-6">Browse by Category</h2>
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
        {rootCategories.map((category) => (
          <CategoryCard key={category.id} category={category} onClick={onCategoryClick} />
        ))}
      </div>
    </div>
  )
}
