import React, { useState } from 'react'
import { StarRatingInput } from './StarRatingInput'
import { useCreateReview } from '../../hooks/useMarketplace'

export interface ReviewFormProps {
  templateId: string
  existingReview?: { rating: number; comment: string }
  onSuccess?: () => void
  onCancel?: () => void
  className?: string
}

export const ReviewForm: React.FC<ReviewFormProps> = ({
  templateId,
  existingReview,
  onSuccess,
  onCancel,
  className = '',
}) => {
  const [rating, setRating] = useState(existingReview?.rating || 5)
  const [comment, setComment] = useState(existingReview?.comment || '')
  const { createReview, loading, error } = useCreateReview(templateId)

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()

    try {
      await createReview({ rating, comment: comment.trim() || undefined })
      if (onSuccess) {
        onSuccess()
      }
    } catch (err) {
      // Error is already handled by the hook
    }
  }

  return (
    <form onSubmit={handleSubmit} className={`bg-white rounded-lg border border-gray-200 p-6 ${className}`}>
      <h3 className="text-lg font-semibold mb-4">{existingReview ? 'Update Your Review' : 'Write a Review'}</h3>

      {/* Rating Input */}
      <div className="mb-4">
        <label className="block text-sm font-medium text-gray-700 mb-2">Your Rating</label>
        <StarRatingInput value={rating} onChange={setRating} size="lg" />
        <p className="text-sm text-gray-600 mt-1">{rating} out of 5 stars</p>
      </div>

      {/* Comment Input */}
      <div className="mb-4">
        <label className="block text-sm font-medium text-gray-700 mb-2">Your Review (Optional)</label>
        <textarea
          className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
          rows={4}
          placeholder="Share your experience with this template..."
          value={comment}
          onChange={(e) => setComment(e.target.value)}
          maxLength={2000}
        />
        <p className="text-sm text-gray-500 mt-1">{comment.length} / 2000 characters</p>
      </div>

      {/* Error Message */}
      {error && (
        <div className="mb-4 p-3 bg-red-50 border border-red-200 rounded-md">
          <p className="text-sm text-red-600">{error.message}</p>
        </div>
      )}

      {/* Actions */}
      <div className="flex gap-2">
        <button
          type="submit"
          disabled={loading}
          className="flex-1 px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
        >
          {loading ? 'Submitting...' : existingReview ? 'Update Review' : 'Submit Review'}
        </button>
        {onCancel && (
          <button
            type="button"
            onClick={onCancel}
            className="px-4 py-2 border border-gray-300 rounded-md hover:bg-gray-50 transition-colors"
          >
            Cancel
          </button>
        )}
      </div>
    </form>
  )
}
