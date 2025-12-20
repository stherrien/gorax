import { useState, useCallback } from 'react'

export interface BulkSelectionReturn<T> {
  selected: Set<string>
  isSelected: (id: string) => boolean
  toggle: (id: string) => void
  toggleAll: (items: T[]) => void
  selectRange: (startId: string, endId: string, items: T[]) => void
  clear: () => void
  count: number
}

interface Identifiable {
  id: string
}

export function useBulkSelection<T extends Identifiable>(): BulkSelectionReturn<T> {
  const [selected, setSelected] = useState<Set<string>>(new Set())

  const isSelected = useCallback(
    (id: string): boolean => {
      return selected.has(id)
    },
    [selected]
  )

  const toggle = useCallback((id: string): void => {
    setSelected((prev) => {
      const next = new Set(prev)
      if (next.has(id)) {
        next.delete(id)
      } else {
        next.add(id)
      }
      return next
    })
  }, [])

  const toggleAll = useCallback(
    (items: T[]): void => {
      setSelected((prev) => {
        // If all items are already selected OR some items are selected, deselect all
        if (prev.size > 0) {
          return new Set()
        }
        // If none are selected, select all
        return new Set(items.map((item) => item.id))
      })
    },
    []
  )

  const selectRange = useCallback(
    (startId: string, endId: string, items: T[]): void => {
      const startIndex = items.findIndex((item) => item.id === startId)
      const endIndex = items.findIndex((item) => item.id === endId)

      if (startIndex === -1 || endIndex === -1) {
        return
      }

      const start = Math.min(startIndex, endIndex)
      const end = Math.max(startIndex, endIndex)

      const rangeIds = items.slice(start, end + 1).map((item) => item.id)

      setSelected((prev) => {
        const next = new Set(prev)
        rangeIds.forEach((id) => next.add(id))
        return next
      })
    },
    []
  )

  const clear = useCallback((): void => {
    setSelected(new Set())
  }, [])

  return {
    selected,
    isSelected,
    toggle,
    toggleAll,
    selectRange,
    clear,
    count: selected.size,
  }
}
