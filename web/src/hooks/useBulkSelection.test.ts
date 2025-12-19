import { renderHook, act } from '@testing-library/react'
import { describe, it, expect } from 'vitest'
import { useBulkSelection } from './useBulkSelection'

interface TestItem {
  id: string
  name: string
}

describe('useBulkSelection', () => {
  const items: TestItem[] = [
    { id: '1', name: 'Item 1' },
    { id: '2', name: 'Item 2' },
    { id: '3', name: 'Item 3' },
    { id: '4', name: 'Item 4' },
    { id: '5', name: 'Item 5' },
  ]

  it('should initialize with empty selection', () => {
    const { result } = renderHook(() => useBulkSelection<TestItem>())

    expect(result.current.selected.size).toBe(0)
    expect(result.current.count).toBe(0)
    expect(result.current.isSelected('1')).toBe(false)
  })

  it('should toggle single selection', () => {
    const { result } = renderHook(() => useBulkSelection<TestItem>())

    act(() => {
      result.current.toggle('1')
    })

    expect(result.current.isSelected('1')).toBe(true)
    expect(result.current.count).toBe(1)

    act(() => {
      result.current.toggle('1')
    })

    expect(result.current.isSelected('1')).toBe(false)
    expect(result.current.count).toBe(0)
  })

  it('should select multiple items', () => {
    const { result } = renderHook(() => useBulkSelection<TestItem>())

    act(() => {
      result.current.toggle('1')
      result.current.toggle('2')
      result.current.toggle('3')
    })

    expect(result.current.count).toBe(3)
    expect(result.current.isSelected('1')).toBe(true)
    expect(result.current.isSelected('2')).toBe(true)
    expect(result.current.isSelected('3')).toBe(true)
  })

  it('should toggle all items', () => {
    const { result } = renderHook(() => useBulkSelection<TestItem>())

    act(() => {
      result.current.toggleAll(items)
    })

    expect(result.current.count).toBe(5)
    items.forEach((item) => {
      expect(result.current.isSelected(item.id)).toBe(true)
    })

    act(() => {
      result.current.toggleAll(items)
    })

    expect(result.current.count).toBe(0)
  })

  it('should select range of items', () => {
    const { result } = renderHook(() => useBulkSelection<TestItem>())

    act(() => {
      result.current.selectRange('2', '4', items)
    })

    expect(result.current.count).toBe(3)
    expect(result.current.isSelected('1')).toBe(false)
    expect(result.current.isSelected('2')).toBe(true)
    expect(result.current.isSelected('3')).toBe(true)
    expect(result.current.isSelected('4')).toBe(true)
    expect(result.current.isSelected('5')).toBe(false)
  })

  it('should select range in reverse order', () => {
    const { result } = renderHook(() => useBulkSelection<TestItem>())

    act(() => {
      result.current.selectRange('4', '2', items)
    })

    expect(result.current.count).toBe(3)
    expect(result.current.isSelected('2')).toBe(true)
    expect(result.current.isSelected('3')).toBe(true)
    expect(result.current.isSelected('4')).toBe(true)
  })

  it('should handle range with only one item', () => {
    const { result } = renderHook(() => useBulkSelection<TestItem>())

    act(() => {
      result.current.selectRange('2', '2', items)
    })

    expect(result.current.count).toBe(1)
    expect(result.current.isSelected('2')).toBe(true)
  })

  it('should clear all selections', () => {
    const { result } = renderHook(() => useBulkSelection<TestItem>())

    act(() => {
      result.current.toggle('1')
      result.current.toggle('2')
      result.current.toggle('3')
    })

    expect(result.current.count).toBe(3)

    act(() => {
      result.current.clear()
    })

    expect(result.current.count).toBe(0)
    expect(result.current.selected.size).toBe(0)
  })

  it('should handle selecting items that do not exist in range', () => {
    const { result } = renderHook(() => useBulkSelection<TestItem>())

    act(() => {
      result.current.selectRange('99', '100', items)
    })

    expect(result.current.count).toBe(0)
  })

  it('should get array of selected IDs', () => {
    const { result } = renderHook(() => useBulkSelection<TestItem>())

    act(() => {
      result.current.toggle('1')
      result.current.toggle('3')
      result.current.toggle('5')
    })

    const selectedIds = Array.from(result.current.selected)
    expect(selectedIds).toContain('1')
    expect(selectedIds).toContain('3')
    expect(selectedIds).toContain('5')
    expect(selectedIds.length).toBe(3)
  })

  it('should support partial selection with toggleAll', () => {
    const { result } = renderHook(() => useBulkSelection<TestItem>())

    // Select 2 items manually
    act(() => {
      result.current.toggle('1')
      result.current.toggle('2')
    })

    expect(result.current.count).toBe(2)

    // Toggle all should select remaining items
    act(() => {
      result.current.toggleAll(items)
    })

    // Should deselect all since some were already selected
    expect(result.current.count).toBe(0)
  })
})
