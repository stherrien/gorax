import React from 'react'
import type { ReviewSortOption } from '../../types/marketplace'

export interface ReviewSortDropdownProps {
  value: ReviewSortOption
  onChange: (sortBy: ReviewSortOption) => void
  className?: string
}

export const ReviewSortDropdown: React.FC<ReviewSortDropdownProps> = ({ value, onChange, className = '' }) => {
  const sortOptions: { value: ReviewSortOption; label: string }[] = [
    { value: 'recent', label: 'Most Recent' },
    { value: 'helpful', label: 'Most Helpful' },
    { value: 'rating_high', label: 'Highest Rating' },
    { value: 'rating_low', label: 'Lowest Rating' },
  ]

  return (
    <select
      value={value}
      onChange={(e) => onChange(e.target.value as ReviewSortOption)}
      className={`px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 ${className}`}
      aria-label="Sort reviews"
    >
      {sortOptions.map((option) => (
        <option key={option.value} value={option.value}>
          {option.label}
        </option>
      ))}
    </select>
  )
}
