import React, { useState } from 'react'
import { StarRating } from './StarRating'
import { useReviewHelpful, useReportReview } from '../../hooks/useMarketplace'
import type { TemplateReview, ReviewReportReason } from '../../types/marketplace'

export interface ReviewCardProps {
  review: TemplateReview
  onDelete?: (reviewId: string) => void
  canDelete?: boolean
  className?: string
}

export const ReviewCard: React.FC<ReviewCardProps> = ({ review, onDelete, canDelete = false, className = '' }) => {
  const { hasVoted, loading: voteLoading, toggleVote } = useReviewHelpful(review.id)
  const { report, loading: reportLoading } = useReportReview()
  const [showReportModal, setShowReportModal] = useState(false)
  const [reportReason, setReportReason] = useState<ReviewReportReason>('spam')
  const [reportDetails, setReportDetails] = useState('')

  const handleReportSubmit = async () => {
    try {
      await report(review.id, { reason: reportReason, details: reportDetails })
      setShowReportModal(false)
      setReportDetails('')
      alert('Review reported successfully')
    } catch (error) {
      alert('Failed to report review')
    }
  }

  if (review.isHidden) {
    return null
  }

  return (
    <>
      <div className={`bg-white rounded-lg border border-gray-200 p-4 ${className}`}>
        {/* Header */}
        <div className="flex items-start justify-between mb-3">
          <div className="flex items-center gap-3">
            <div className="w-10 h-10 rounded-full bg-blue-500 flex items-center justify-center text-white font-semibold">
              {review.userName.charAt(0).toUpperCase()}
            </div>
            <div>
              <div className="font-medium text-gray-900">{review.userName}</div>
              <div className="text-sm text-gray-500">
                {new Date(review.createdAt).toLocaleDateString('en-US', {
                  year: 'numeric',
                  month: 'long',
                  day: 'numeric',
                })}
              </div>
            </div>
          </div>
          <StarRating rating={review.rating} size="sm" />
        </div>

        {/* Comment */}
        {review.comment && <p className="text-gray-700 mb-4">{review.comment}</p>}

        {/* Actions */}
        <div className="flex items-center gap-4 text-sm">
          <button
            onClick={toggleVote}
            disabled={voteLoading}
            className={`flex items-center gap-1 ${
              hasVoted ? 'text-blue-600 font-medium' : 'text-gray-600 hover:text-blue-600'
            } transition-colors disabled:opacity-50`}
          >
            <svg className="w-4 h-4" fill="currentColor" viewBox="0 0 20 20">
              <path d="M2 10.5a1.5 1.5 0 113 0v6a1.5 1.5 0 01-3 0v-6zM6 10.333v5.43a2 2 0 001.106 1.79l.05.025A4 4 0 008.943 18h5.416a2 2 0 001.962-1.608l1.2-6A2 2 0 0015.56 8H12V4a2 2 0 00-2-2 1 1 0 00-1 1v.667a4 4 0 01-.8 2.4L6.8 7.933a4 4 0 00-.8 2.4z" />
            </svg>
            Helpful ({review.helpfulCount})
          </button>

          <button
            onClick={() => setShowReportModal(true)}
            className="text-gray-600 hover:text-red-600 transition-colors"
          >
            Report
          </button>

          {canDelete && onDelete && (
            <button onClick={() => onDelete(review.id)} className="text-gray-600 hover:text-red-600 transition-colors">
              Delete
            </button>
          )}
        </div>
      </div>

      {/* Report Modal */}
      {showReportModal && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50 p-4">
          <div className="bg-white rounded-lg max-w-md w-full p-6">
            <h3 className="text-lg font-semibold mb-4">Report Review</h3>

            <div className="mb-4">
              <label className="block text-sm font-medium text-gray-700 mb-2">Reason</label>
              <select
                className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                value={reportReason}
                onChange={(e) => setReportReason(e.target.value as ReviewReportReason)}
              >
                <option value="spam">Spam</option>
                <option value="inappropriate">Inappropriate Content</option>
                <option value="offensive">Offensive Language</option>
                <option value="misleading">Misleading Information</option>
                <option value="other">Other</option>
              </select>
            </div>

            <div className="mb-4">
              <label className="block text-sm font-medium text-gray-700 mb-2">Details (optional)</label>
              <textarea
                className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                rows={3}
                placeholder="Provide additional context..."
                value={reportDetails}
                onChange={(e) => setReportDetails(e.target.value)}
              />
            </div>

            <div className="flex gap-2">
              <button
                onClick={handleReportSubmit}
                disabled={reportLoading}
                className="flex-1 px-4 py-2 bg-red-600 text-white rounded-md hover:bg-red-700 transition-colors disabled:opacity-50"
              >
                {reportLoading ? 'Reporting...' : 'Submit Report'}
              </button>
              <button
                onClick={() => setShowReportModal(false)}
                className="px-4 py-2 border border-gray-300 rounded-md hover:bg-gray-50 transition-colors"
              >
                Cancel
              </button>
            </div>
          </div>
        </div>
      )}
    </>
  )
}
