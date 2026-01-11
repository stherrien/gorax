import React from 'react'
import type { Category } from '../../types/marketplace'

export interface CategoryCardProps {
  category: Category
  onClick?: (category: Category) => void
  className?: string
}

export const CategoryCard: React.FC<CategoryCardProps> = ({ category, onClick, className = '' }) => {
  const handleClick = () => {
    if (onClick) {
      onClick(category)
    }
  }

  return (
    <button
      onClick={handleClick}
      className={`group bg-white rounded-lg border-2 border-gray-200 p-6 hover:border-blue-500 hover:shadow-lg transition-all duration-200 text-left w-full ${className}`}
    >
      {/* Icon */}
      <div className="flex items-center gap-4 mb-3">
        <div className="w-12 h-12 rounded-lg bg-blue-100 flex items-center justify-center text-2xl group-hover:bg-blue-500 group-hover:text-white transition-colors">
          {category.icon ? (
            <span>{getIconEmoji(category.icon)}</span>
          ) : (
            <span className="text-blue-600 group-hover:text-white">📦</span>
          )}
        </div>
        <div className="flex-1">
          <h3 className="font-semibold text-gray-900 group-hover:text-blue-600 transition-colors">{category.name}</h3>
          <p className="text-sm text-gray-500">{category.templateCount} templates</p>
        </div>
      </div>

      {/* Description */}
      {category.description && (
        <p className="text-sm text-gray-600 line-clamp-2 mb-2">{category.description}</p>
      )}

      {/* Arrow */}
      <div className="flex justify-end">
        <svg
          className="w-5 h-5 text-gray-400 group-hover:text-blue-600 group-hover:translate-x-1 transition-all"
          fill="none"
          stroke="currentColor"
          viewBox="0 0 24 24"
        >
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 5l7 7-7 7" />
        </svg>
      </div>
    </button>
  )
}

// Helper function to map icon names to emojis
function getIconEmoji(iconName: string): string {
  const iconMap: Record<string, string> = {
    link: '🔗',
    zap: '⚡',
    database: '💾',
    bell: '🔔',
    server: '🖥️',
    shield: '🛡️',
    'bar-chart': '📊',
    'message-circle': '💬',
    activity: '📈',
    clock: '⏰',
    package: '📦',
    code: '💻',
    cloud: '☁️',
    settings: '⚙️',
    users: '👥',
  }
  return iconMap[iconName] || '📦'
}
