import React from 'react'
import { StarRating } from './StarRating'
import { FeaturedBadge } from './FeaturedBadge'
import type { MarketplaceTemplate } from '../../types/marketplace'

export interface TemplateCardProps {
  template: MarketplaceTemplate
  onViewDetails?: (template: MarketplaceTemplate) => void
  onInstall?: (template: MarketplaceTemplate) => void
  showActions?: boolean
  className?: string
}

export const TemplateCard: React.FC<TemplateCardProps> = ({
  template,
  onViewDetails,
  onInstall,
  showActions = true,
  className = '',
}) => {
  const handleCardClick = () => {
    if (onViewDetails) {
      onViewDetails(template)
    }
  }

  const handleInstallClick = (e: React.MouseEvent) => {
    e.stopPropagation()
    if (onInstall) {
      onInstall(template)
    }
  }

  return (
    <div
      className={`group bg-white rounded-lg border border-gray-200 hover:border-blue-500 hover:shadow-xl transition-all duration-200 cursor-pointer overflow-hidden ${className}`}
      onClick={handleCardClick}
    >
      {/* Card Header */}
      <div className="p-6">
        <div className="flex items-start justify-between mb-3">
          <h3 className="text-lg font-semibold text-gray-900 group-hover:text-blue-600 transition-colors flex-1 line-clamp-2">
            {template.name}
          </h3>
          <div className="flex gap-2 ml-2">
            {template.isFeatured && <FeaturedBadge size="sm" />}
            {template.isVerified && (
              <span className="inline-flex items-center px-2 py-1 rounded-full text-xs font-medium bg-blue-100 text-blue-800">
                <svg className="w-3 h-3 mr-1" fill="currentColor" viewBox="0 0 20 20">
                  <path
                    fillRule="evenodd"
                    d="M6.267 3.455a3.066 3.066 0 001.745-.723 3.066 3.066 0 013.976 0 3.066 3.066 0 001.745.723 3.066 3.066 0 012.812 2.812c.051.643.304 1.254.723 1.745a3.066 3.066 0 010 3.976 3.066 3.066 0 00-.723 1.745 3.066 3.066 0 01-2.812 2.812 3.066 3.066 0 00-1.745.723 3.066 3.066 0 01-3.976 0 3.066 3.066 0 00-1.745-.723 3.066 3.066 0 01-2.812-2.812 3.066 3.066 0 00-.723-1.745 3.066 3.066 0 010-3.976 3.066 3.066 0 00.723-1.745 3.066 3.066 0 012.812-2.812zm7.44 5.252a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z"
                    clipRule="evenodd"
                  />
                </svg>
                Verified
              </span>
            )}
          </div>
        </div>

        {/* Description */}
        <p className="text-gray-600 text-sm mb-4 line-clamp-2">{template.description}</p>

        {/* Category & Tags */}
        <div className="flex items-center gap-2 mb-4 flex-wrap">
          <span className="inline-flex items-center px-2.5 py-1 rounded-full text-xs font-medium bg-blue-100 text-blue-800">
            {template.category}
          </span>
          {template.tags.slice(0, 2).map((tag) => (
            <span
              key={tag}
              className="inline-flex items-center px-2.5 py-1 rounded-full text-xs font-medium bg-gray-100 text-gray-700"
            >
              {tag}
            </span>
          ))}
          {template.tags.length > 2 && (
            <span className="inline-flex items-center px-2.5 py-1 rounded-full text-xs font-medium bg-gray-100 text-gray-600">
              +{template.tags.length - 2}
            </span>
          )}
        </div>

        {/* Stats */}
        <div className="flex items-center justify-between mb-4 text-sm">
          <StarRating rating={template.averageRating} showCount count={template.totalRatings} size="sm" />
          <div className="flex items-center gap-1 text-gray-600">
            <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={2}
                d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-4l-4 4m0 0l-4-4m4 4V4"
              />
            </svg>
            {template.downloadCount.toLocaleString()}
          </div>
        </div>

        {/* Author & Version */}
        <div className="flex items-center justify-between text-xs text-gray-500 mb-4 pb-4 border-b border-gray-200">
          <div className="flex items-center gap-1">
            <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={2}
                d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z"
              />
            </svg>
            {template.authorName}
          </div>
          <div>v{template.version}</div>
        </div>

        {/* Actions */}
        {showActions && (
          <div className="flex gap-2">
            <button
              onClick={handleInstallClick}
              className="flex-1 px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 transition-colors font-medium text-sm"
            >
              Install
            </button>
            <button
              onClick={(e) => {
                e.stopPropagation()
                handleCardClick()
              }}
              className="px-4 py-2 border border-gray-300 rounded-md hover:bg-gray-50 transition-colors text-sm font-medium text-gray-700"
            >
              Details
            </button>
          </div>
        )}
      </div>
    </div>
  )
}
