import React, { useState, useEffect } from 'react'
import { FeaturedBadge } from './FeaturedBadge'
import { StarRating } from './StarRating'
import { useFeaturedTemplates } from '../../hooks/useMarketplace'
import type { MarketplaceTemplate } from '../../types/marketplace'

export interface FeaturedTemplatesProps {
  onTemplateClick?: (template: MarketplaceTemplate) => void
  className?: string
}

export const FeaturedTemplates: React.FC<FeaturedTemplatesProps> = ({ onTemplateClick, className = '' }) => {
  const { templates, loading, error } = useFeaturedTemplates(5)
  const [currentIndex, setCurrentIndex] = useState(0)
  const [isAutoPlaying, setIsAutoPlaying] = useState(true)

  // Auto-play carousel
  useEffect(() => {
    if (!isAutoPlaying || templates.length <= 1) return

    const interval = setInterval(() => {
      setCurrentIndex((current) => (current + 1) % templates.length)
    }, 5000)

    return () => clearInterval(interval)
  }, [isAutoPlaying, templates.length])

  const goToSlide = (index: number) => {
    setCurrentIndex(index)
    setIsAutoPlaying(false)
  }

  const goToPrevious = () => {
    setCurrentIndex((current) => (current - 1 + templates.length) % templates.length)
    setIsAutoPlaying(false)
  }

  const goToNext = () => {
    setCurrentIndex((current) => (current + 1) % templates.length)
    setIsAutoPlaying(false)
  }

  if (loading) {
    return (
      <div className={`bg-gradient-to-r from-blue-600 to-purple-600 rounded-lg p-12 ${className}`}>
        <div className="animate-pulse flex flex-col items-center">
          <div className="h-8 bg-white/20 rounded w-64 mb-4" />
          <div className="h-4 bg-white/20 rounded w-96" />
        </div>
      </div>
    )
  }

  if (error || templates.length === 0) {
    return null
  }

  const currentTemplate = templates[currentIndex]

  return (
    <div className={`relative bg-gradient-to-r from-blue-600 to-purple-600 rounded-lg overflow-hidden ${className}`}>
      {/* Main Content */}
      <div className="relative z-10 p-8 md:p-12">
        <div className="max-w-3xl">
          <FeaturedBadge size="lg" className="mb-4" />
          <h2 className="text-3xl md:text-4xl font-bold text-white mb-4">{currentTemplate.name}</h2>
          <p className="text-lg text-blue-100 mb-6 line-clamp-2">{currentTemplate.description}</p>

          <div className="flex flex-wrap items-center gap-4 mb-6">
            <StarRating rating={currentTemplate.averageRating} showCount count={currentTemplate.totalRatings} />
            <span className="text-white">•</span>
            <span className="text-blue-100">{currentTemplate.downloadCount} downloads</span>
            <span className="text-white">•</span>
            <span className="text-blue-100">By {currentTemplate.authorName}</span>
          </div>

          <button
            onClick={() => onTemplateClick && onTemplateClick(currentTemplate)}
            className="px-6 py-3 bg-white text-blue-600 font-semibold rounded-lg hover:bg-blue-50 transition-colors"
          >
            View Template
          </button>
        </div>
      </div>

      {/* Navigation */}
      {templates.length > 1 && (
        <>
          <button
            onClick={goToPrevious}
            className="absolute left-4 top-1/2 -translate-y-1/2 z-20 w-10 h-10 flex items-center justify-center bg-white/20 hover:bg-white/30 backdrop-blur-sm rounded-full transition-colors"
            aria-label="Previous template"
          >
            <svg className="w-6 h-6 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 19l-7-7 7-7" />
            </svg>
          </button>

          <button
            onClick={goToNext}
            className="absolute right-4 top-1/2 -translate-y-1/2 z-20 w-10 h-10 flex items-center justify-center bg-white/20 hover:bg-white/30 backdrop-blur-sm rounded-full transition-colors"
            aria-label="Next template"
          >
            <svg className="w-6 h-6 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 5l7 7-7 7" />
            </svg>
          </button>

          {/* Dots */}
          <div className="absolute bottom-4 left-1/2 -translate-x-1/2 z-20 flex gap-2">
            {templates.map((_, index) => (
              <button
                key={index}
                onClick={() => goToSlide(index)}
                className={`w-2 h-2 rounded-full transition-all ${
                  index === currentIndex ? 'bg-white w-8' : 'bg-white/50 hover:bg-white/75'
                }`}
                aria-label={`Go to slide ${index + 1}`}
              />
            ))}
          </div>
        </>
      )}

      {/* Background Pattern */}
      <div className="absolute inset-0 opacity-10">
        <svg className="w-full h-full" viewBox="0 0 100 100" preserveAspectRatio="none">
          <defs>
            <pattern id="grid" width="10" height="10" patternUnits="userSpaceOnUse">
              <path d="M 10 0 L 0 0 0 10" fill="none" stroke="white" strokeWidth="0.5" />
            </pattern>
          </defs>
          <rect width="100" height="100" fill="url(#grid)" />
        </svg>
      </div>
    </div>
  )
}
