import { useRef, useEffect } from 'react'

interface Identifiable {
  id: string
}

interface Column<T> {
  header: string
  accessor: (item: T) => React.ReactNode
}

interface SelectableTableProps<T extends Identifiable> {
  items: T[]
  columns: Column<T>[]
  selectedIds: Set<string>
  onSelectionChange: (id: string) => void
  onSelectAll?: () => void
  onRangeSelect?: (startId: string, endId: string) => void
  emptyMessage?: string
}

export function SelectableTable<T extends Identifiable>({
  items,
  columns,
  selectedIds,
  onSelectionChange,
  onSelectAll,
  onRangeSelect,
  emptyMessage = 'No items to display',
}: SelectableTableProps<T>) {
  const headerCheckboxRef = useRef<HTMLInputElement>(null)
  const lastSelectedIdRef = useRef<string | null>(null)

  const allSelected = items.length > 0 && selectedIds.size === items.length
  const someSelected = selectedIds.size > 0 && selectedIds.size < items.length

  useEffect(() => {
    if (headerCheckboxRef.current) {
      headerCheckboxRef.current.indeterminate = someSelected
    }
  }, [someSelected])

  const handleRowClick = (id: string, event: React.MouseEvent) => {
    if (event.shiftKey && onRangeSelect && lastSelectedIdRef.current) {
      onRangeSelect(lastSelectedIdRef.current, id)
    } else if (event.shiftKey && onRangeSelect && selectedIds.size > 0) {
      // If shift is pressed but no lastSelectedId, use the first selected item
      const firstSelectedId = Array.from(selectedIds)[0]
      onRangeSelect(firstSelectedId, id)
    } else {
      onSelectionChange(id)
      lastSelectedIdRef.current = id
    }
  }

  if (items.length === 0) {
    return (
      <div className="text-center py-12 text-gray-400">
        {emptyMessage}
      </div>
    )
  }

  return (
    <div className="overflow-x-auto">
      <table className="w-full">
        <thead className="bg-gray-700">
          <tr>
            <th className="px-4 py-3 text-left w-12">
              {onSelectAll ? (
                <input
                  ref={headerCheckboxRef}
                  type="checkbox"
                  checked={allSelected}
                  onChange={onSelectAll}
                  className="w-4 h-4 rounded border-gray-600 bg-gray-700 text-primary-500 focus:ring-primary-500 focus:ring-offset-gray-800 cursor-pointer"
                  aria-label="Select all"
                />
              ) : (
                <input
                  ref={headerCheckboxRef}
                  type="checkbox"
                  checked={allSelected}
                  readOnly
                  className="w-4 h-4 rounded border-gray-600 bg-gray-700 text-primary-500 focus:ring-primary-500 focus:ring-offset-gray-800 opacity-50 cursor-not-allowed"
                  aria-label="Select all"
                />
              )}
            </th>
            {columns.map((column, index) => (
              <th
                key={index}
                className="px-4 py-3 text-left text-xs font-medium text-gray-400 uppercase tracking-wider"
              >
                {column.header}
              </th>
            ))}
          </tr>
        </thead>
        <tbody className="divide-y divide-gray-700">
          {items.map((item) => (
            <tr
              key={item.id}
              className={`hover:bg-gray-700/50 ${
                selectedIds.has(item.id) ? 'bg-gray-700/30' : ''
              }`}
            >
              <td className="px-4 py-3">
                <input
                  type="checkbox"
                  checked={selectedIds.has(item.id)}
                  onChange={() => {}}
                  onClick={(e) => {
                    e.stopPropagation()
                    handleRowClick(item.id, e)
                  }}
                  className="w-4 h-4 rounded border-gray-600 bg-gray-700 text-primary-500 focus:ring-primary-500 focus:ring-offset-gray-800 cursor-pointer"
                  aria-label={`Select ${item.id}`}
                />
              </td>
              {columns.map((column, index) => (
                <td key={index} className="px-4 py-3 text-sm">
                  {column.accessor(item)}
                </td>
              ))}
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  )
}
