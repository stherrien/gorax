import React from 'react'
import { StarRating } from './StarRating'
import type { RatingDistribution } from '../../types/marketplace'

export interface RatingSummaryProps {
  distribution: RatingDistribution
  className?: string
}

export const RatingSummary: React.FC<RatingSummaryProps> = ({ distribution, className = '' }) => {
  const { averageRating, totalRatings, rating1Percent, rating2Percent, rating3Percent, rating4Percent, rating5Percent } =
    distribution

  const ratings = [
    { stars: 5, percent: rating5Percent },
    { stars: 4, percent: rating4Percent },
    { stars: 3, percent: rating3Percent },
    { stars: 2, percent: rating2Percent },
    { stars: 1, percent: rating1Percent },
  ]

  return (
    <div className={`bg-white rounded-lg p-6 ${className}`}>
      <h3 className="text-lg font-semibold mb-4">Rating Summary</h3>

      <div className="flex items-start gap-8 mb-6">
        {/* Average Rating */}
        <div className="text-center">
          <div className="text-4xl font-bold text-gray-900 mb-2">{averageRating.toFixed(1)}</div>
          <StarRating rating={averageRating} size="lg" className="mb-2" />
          <div className="text-sm text-gray-600">{totalRatings} ratings</div>
        </div>

        {/* Rating Distribution */}
        <div className="flex-1">
          {ratings.map(({ stars, percent }) => (
            <div key={stars} className="flex items-center gap-3 mb-2">
              <div className="text-sm font-medium text-gray-700 w-8">{stars}★</div>
              <div className="flex-1 h-3 bg-gray-200 rounded-full overflow-hidden">
                <div
                  className="h-full bg-yellow-400 transition-all duration-300"
                  style={{ width: `${percent}%` }}
                />
              </div>
              <div className="text-sm text-gray-600 w-12 text-right">{percent.toFixed(0)}%</div>
            </div>
          ))}
        </div>
      </div>
    </div>
  )
}
