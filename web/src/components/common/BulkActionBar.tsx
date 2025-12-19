interface BulkActionBarProps {
  count: number
  onClear: () => void
  children: React.ReactNode
}

export function BulkActionBar({ count, onClear, children }: BulkActionBarProps) {
  if (count === 0) {
    return null
  }

  return (
    <div className="sticky top-0 z-10 bg-primary-600 border-b border-primary-700 px-4 py-3 mb-4 rounded-lg shadow-lg">
      <div className="flex items-center justify-between">
        <div className="flex items-center space-x-4">
          <span className="text-white font-medium">
            {count} {count === 1 ? 'item' : 'items'} selected
          </span>
          <div className="flex items-center space-x-2">
            {children}
          </div>
        </div>
        <button
          onClick={onClear}
          className="text-white hover:text-gray-200 text-sm font-medium"
        >
          Clear selection
        </button>
      </div>
    </div>
  )
}
