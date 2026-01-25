import React, { useState } from 'react'
import { ReviewCard } from './ReviewCard'
import { ReviewSortDropdown } from './ReviewSortDropdown'
import { useMarketplaceReviews } from '../../hooks/useMarketplace'
import type { ReviewSortOption } from '../../types/marketplace'

export interface ReviewListProps {
  templateId: string
  currentUserId?: string
  onReviewDeleted?: () => void
  className?: string
}

export const ReviewList: React.FC<ReviewListProps> = ({
  templateId,
  currentUserId,
  onReviewDeleted,
  className = '',
}) => {
  const [sortBy, setSortBy] = useState<ReviewSortOption>('recent')
  const [page, setPage] = useState(0)
  const limit = 10

  const { reviews, loading, error, deleteReview, refresh } = useMarketplaceReviews(
    templateId,
    sortBy,
    limit,
    page * limit
  )

  const handleDelete = async (reviewId: string) => {
    if (window.confirm('Are you sure you want to delete this review?')) {
      try {
        await deleteReview(reviewId)
        if (onReviewDeleted) {
          onReviewDeleted()
        }
      } catch (error) {
        alert('Failed to delete review')
      }
    }
  }

  const handleSortChange = (newSortBy: ReviewSortOption) => {
    setSortBy(newSortBy)
    setPage(0) // Reset to first page when changing sort
  }

  if (loading && reviews.length === 0) {
    return (
      <div className={`flex items-center justify-center py-12 ${className}`}>
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600" />
      </div>
    )
  }

  if (error) {
    return (
      <div className={`text-center py-12 ${className}`}>
        <p className="text-red-600">Failed to load reviews: {error.message}</p>
        <button
          onClick={refresh}
          className="mt-4 px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 transition-colors"
        >
          Retry
        </button>
      </div>
    )
  }

  if (reviews.length === 0) {
    return (
      <div className={`text-center py-12 ${className}`}>
        <p className="text-gray-500">No reviews yet. Be the first to review this template!</p>
      </div>
    )
  }

  return (
    <div className={className}>
      {/* Sort Controls */}
      <div className="flex items-center justify-between mb-4">
        <h3 className="text-lg font-semibold">Reviews ({reviews.length})</h3>
        <ReviewSortDropdown value={sortBy} onChange={handleSortChange} />
      </div>

      {/* Reviews */}
      <div className="space-y-4">
        {reviews.map((review) => (
          <ReviewCard
            key={review.id}
            review={review}
            onDelete={handleDelete}
            canDelete={currentUserId === review.userId}
          />
        ))}
      </div>

      {/* Pagination */}
      {reviews.length === limit && (
        <div className="flex justify-center gap-2 mt-6">
          <button
            onClick={() => setPage((p) => Math.max(0, p - 1))}
            disabled={page === 0}
            className="px-4 py-2 border border-gray-300 rounded-md hover:bg-gray-50 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
          >
            Previous
          </button>
          <span className="px-4 py-2 text-gray-700">Page {page + 1}</span>
          <button
            onClick={() => setPage((p) => p + 1)}
            className="px-4 py-2 border border-gray-300 rounded-md hover:bg-gray-50 transition-colors"
          >
            Next
          </button>
        </div>
      )}
    </div>
  )
}
