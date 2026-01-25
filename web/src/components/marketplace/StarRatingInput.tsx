import React, { useState } from 'react'

export interface StarRatingInputProps {
  value: number
  onChange: (rating: number) => void
  maxRating?: number
  size?: 'sm' | 'md' | 'lg'
  disabled?: boolean
  className?: string
}

export const StarRatingInput: React.FC<StarRatingInputProps> = ({
  value,
  onChange,
  maxRating = 5,
  size = 'md',
  disabled = false,
  className = '',
}) => {
  const [hoverRating, setHoverRating] = useState<number | null>(null)

  const sizeClasses = {
    sm: 'text-lg',
    md: 'text-2xl',
    lg: 'text-3xl',
  }

  const displayRating = hoverRating !== null ? hoverRating : value

  const handleClick = (rating: number) => {
    if (!disabled) {
      onChange(rating)
    }
  }

  const handleMouseEnter = (rating: number) => {
    if (!disabled) {
      setHoverRating(rating)
    }
  }

  const handleMouseLeave = () => {
    setHoverRating(null)
  }

  return (
    <div className={`flex gap-1 ${className}`} onMouseLeave={handleMouseLeave}>
      {Array.from({ length: maxRating }, (_, i) => i + 1).map((rating) => (
        <button
          key={rating}
          type="button"
          className={`${sizeClasses[size]} transition-colors ${
            disabled ? 'cursor-not-allowed opacity-50' : 'cursor-pointer hover:scale-110'
          } ${rating <= displayRating ? 'text-yellow-400' : 'text-gray-300'}`}
          onClick={() => handleClick(rating)}
          onMouseEnter={() => handleMouseEnter(rating)}
          disabled={disabled}
          aria-label={`Rate ${rating} out of ${maxRating} stars`}
        >
          ★
        </button>
      ))}
    </div>
  )
}
