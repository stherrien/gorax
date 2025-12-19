import { useState, useRef, useEffect } from 'react'
import { format, startOfDay, endOfDay, subDays, startOfMonth, endOfMonth } from 'date-fns'

export interface DateRange {
  startDate: Date
  endDate: Date
}

export interface DateRangePickerProps {
  value: DateRange | null
  onChange: (value: DateRange | null) => void
  placeholder?: string
  disabled?: boolean
  className?: string
}

interface DatePreset {
  label: string
  getValue: () => DateRange
}

const presets: DatePreset[] = [
  {
    label: 'Today',
    getValue: () => ({
      startDate: startOfDay(new Date()),
      endDate: endOfDay(new Date()),
    }),
  },
  {
    label: 'Yesterday',
    getValue: () => {
      const yesterday = subDays(new Date(), 1)
      return {
        startDate: startOfDay(yesterday),
        endDate: endOfDay(yesterday),
      }
    },
  },
  {
    label: 'Last 7 days',
    getValue: () => ({
      startDate: startOfDay(subDays(new Date(), 6)),
      endDate: endOfDay(new Date()),
    }),
  },
  {
    label: 'Last 30 days',
    getValue: () => ({
      startDate: startOfDay(subDays(new Date(), 29)),
      endDate: endOfDay(new Date()),
    }),
  },
  {
    label: 'This month',
    getValue: () => ({
      startDate: startOfMonth(new Date()),
      endDate: endOfMonth(new Date()),
    }),
  },
]

function formatDateRange(range: DateRange): string {
  return `${format(range.startDate, 'MMM d, yyyy')} - ${format(range.endDate, 'MMM d, yyyy')}`
}

export default function DateRangePicker({
  value,
  onChange,
  placeholder = 'Select date range',
  disabled = false,
  className = '',
}: DateRangePickerProps) {
  const [isOpen, setIsOpen] = useState(false)
  const containerRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (containerRef.current && !containerRef.current.contains(event.target as Node)) {
        setIsOpen(false)
      }
    }

    if (isOpen) {
      document.addEventListener('mousedown', handleClickOutside)
      return () => document.removeEventListener('mousedown', handleClickOutside)
    }
  }, [isOpen])

  const handlePresetClick = (preset: DatePreset) => {
    const range = preset.getValue()
    onChange(range)
  }

  const handleClear = () => {
    onChange(null)
  }

  const handleApply = () => {
    setIsOpen(false)
  }

  return (
    <div ref={containerRef} className={`relative ${className}`}>
      <button
        type="button"
        onClick={() => !disabled && setIsOpen(!isOpen)}
        disabled={disabled}
        className={`
          w-full px-4 py-2 text-left bg-gray-700 text-white rounded-lg
          border border-gray-600 hover:border-gray-500
          focus:outline-none focus:ring-2 focus:ring-primary-500
          disabled:opacity-50 disabled:cursor-not-allowed
          flex items-center justify-between
        `}
      >
        <span className={value ? 'text-white' : 'text-gray-400'}>
          {value ? formatDateRange(value) : placeholder}
        </span>
        <svg
          className="w-5 h-5 text-gray-400"
          fill="none"
          stroke="currentColor"
          viewBox="0 0 24 24"
        >
          <path
            strokeLinecap="round"
            strokeLinejoin="round"
            strokeWidth={2}
            d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z"
          />
        </svg>
      </button>

      {isOpen && (
        <div
          role="dialog"
          className="absolute z-50 mt-2 w-80 bg-gray-800 border border-gray-700 rounded-lg shadow-xl p-4"
        >
          <div className="space-y-3">
            <div className="text-sm font-medium text-gray-300 mb-2">Quick Select</div>

            <div className="grid grid-cols-2 gap-2">
              {presets.map((preset) => (
                <button
                  key={preset.label}
                  type="button"
                  onClick={() => handlePresetClick(preset)}
                  className="px-3 py-2 text-sm bg-gray-700 hover:bg-gray-600 text-white rounded-md transition-colors"
                >
                  {preset.label}
                </button>
              ))}
            </div>

            {value && (
              <div className="pt-3 border-t border-gray-700">
                <div className="text-xs text-gray-400 mb-2">Selected Range</div>
                <div className="text-sm text-white bg-gray-700/50 px-3 py-2 rounded-md">
                  {formatDateRange(value)}
                </div>
              </div>
            )}

            <div className="flex gap-2 pt-3 border-t border-gray-700">
              <button
                type="button"
                onClick={handleClear}
                className="flex-1 px-3 py-2 text-sm bg-gray-700 hover:bg-gray-600 text-white rounded-md transition-colors"
              >
                Clear
              </button>
              <button
                type="button"
                onClick={handleApply}
                className="flex-1 px-3 py-2 text-sm bg-primary-600 hover:bg-primary-700 text-white rounded-md transition-colors"
              >
                Apply
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
