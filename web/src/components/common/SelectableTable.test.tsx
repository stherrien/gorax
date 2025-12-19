import { render, screen, fireEvent } from '@testing-library/react'
import { describe, it, expect, vi } from 'vitest'
import { SelectableTable } from './SelectableTable'

interface TestItem {
  id: string
  name: string
}

describe('SelectableTable', () => {
  const items: TestItem[] = [
    { id: '1', name: 'Item 1' },
    { id: '2', name: 'Item 2' },
    { id: '3', name: 'Item 3' },
  ]

  const columns = [
    { header: 'Name', accessor: (item: TestItem) => item.name },
  ]

  it('should render table with items', () => {
    const onSelectionChange = vi.fn()

    render(
      <SelectableTable
        items={items}
        columns={columns}
        selectedIds={new Set()}
        onSelectionChange={onSelectionChange}
      />
    )

    expect(screen.getByText('Item 1')).toBeInTheDocument()
    expect(screen.getByText('Item 2')).toBeInTheDocument()
    expect(screen.getByText('Item 3')).toBeInTheDocument()
  })

  it('should render checkboxes for each row', () => {
    const onSelectionChange = vi.fn()

    render(
      <SelectableTable
        items={items}
        columns={columns}
        selectedIds={new Set()}
        onSelectionChange={onSelectionChange}
      />
    )

    const checkboxes = screen.getAllByRole('checkbox')
    expect(checkboxes.length).toBe(4) // 1 header + 3 rows
  })

  it('should call onSelectionChange when row checkbox is clicked', () => {
    const onSelectionChange = vi.fn()

    render(
      <SelectableTable
        items={items}
        columns={columns}
        selectedIds={new Set()}
        onSelectionChange={onSelectionChange}
      />
    )

    const checkboxes = screen.getAllByRole('checkbox')
    fireEvent.click(checkboxes[1]) // Click first row checkbox

    expect(onSelectionChange).toHaveBeenCalledWith('1')
  })

  it('should render checked checkboxes for selected items', () => {
    const onSelectionChange = vi.fn()
    const selectedIds = new Set(['1', '3'])

    render(
      <SelectableTable
        items={items}
        columns={columns}
        selectedIds={selectedIds}
        onSelectionChange={onSelectionChange}
      />
    )

    const checkboxes = screen.getAllByRole('checkbox')
    expect(checkboxes[1]).toBeChecked() // Row 1
    expect(checkboxes[2]).not.toBeChecked() // Row 2
    expect(checkboxes[3]).toBeChecked() // Row 3
  })

  it('should call onSelectAll when header checkbox is clicked', () => {
    const onSelectAll = vi.fn()
    const onSelectionChange = vi.fn()

    render(
      <SelectableTable
        items={items}
        columns={columns}
        selectedIds={new Set()}
        onSelectionChange={onSelectionChange}
        onSelectAll={onSelectAll}
      />
    )

    const headerCheckbox = screen.getAllByRole('checkbox')[0]
    fireEvent.click(headerCheckbox)

    expect(onSelectAll).toHaveBeenCalled()
  })

  it('should show indeterminate state when some items selected', () => {
    const onSelectAll = vi.fn()
    const onSelectionChange = vi.fn()
    const selectedIds = new Set(['1'])

    render(
      <SelectableTable
        items={items}
        columns={columns}
        selectedIds={selectedIds}
        onSelectionChange={onSelectionChange}
        onSelectAll={onSelectAll}
      />
    )

    const headerCheckbox = screen.getAllByRole('checkbox')[0] as HTMLInputElement
    expect(headerCheckbox.indeterminate).toBe(true)
  })

  it('should show checked state when all items selected', () => {
    const onSelectAll = vi.fn()
    const onSelectionChange = vi.fn()
    const selectedIds = new Set(['1', '2', '3'])

    render(
      <SelectableTable
        items={items}
        columns={columns}
        selectedIds={selectedIds}
        onSelectionChange={onSelectionChange}
        onSelectAll={onSelectAll}
      />
    )

    const headerCheckbox = screen.getAllByRole('checkbox')[0]
    expect(headerCheckbox).toBeChecked()
  })

  it('should render empty state when no items', () => {
    const onSelectionChange = vi.fn()

    render(
      <SelectableTable
        items={[]}
        columns={columns}
        selectedIds={new Set()}
        onSelectionChange={onSelectionChange}
        emptyMessage="No items found"
      />
    )

    expect(screen.getByText('No items found')).toBeInTheDocument()
  })

  it('should render custom cell content', () => {
    const customColumns = [
      {
        header: 'Custom',
        accessor: (item: TestItem) => <span data-testid="custom-cell">{item.name} - Custom</span>,
      },
    ]

    render(
      <SelectableTable
        items={items}
        columns={customColumns}
        selectedIds={new Set()}
        onSelectionChange={vi.fn()}
      />
    )

    const customCells = screen.getAllByTestId('custom-cell')
    expect(customCells.length).toBe(3)
    expect(customCells[0]).toHaveTextContent('Item 1 - Custom')
  })

  it('should handle shift+click for range selection', () => {
    const onRangeSelect = vi.fn()
    const onSelectionChange = vi.fn()

    render(
      <SelectableTable
        items={items}
        columns={columns}
        selectedIds={new Set(['1'])}
        onSelectionChange={onSelectionChange}
        onRangeSelect={onRangeSelect}
      />
    )

    const checkboxes = screen.getAllByRole('checkbox')
    fireEvent.click(checkboxes[3], { shiftKey: true }) // Shift+click third row

    expect(onRangeSelect).toHaveBeenCalledWith('1', '3')
  })
})
