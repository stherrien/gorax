import { describe, it, expect } from 'vitest'
import { render, screen } from '@testing-library/react'
import PriorityBadge from './PriorityBadge'

describe('PriorityBadge', () => {
  it('renders Low priority badge with correct styling', () => {
    render(<PriorityBadge priority={0} />)
    const badge = screen.getByText('Low')
    expect(badge).toBeInTheDocument()
    expect(badge).toHaveClass('bg-gray-500/20')
    expect(badge).toHaveClass('text-gray-400')
  })

  it('renders Normal priority badge with correct styling', () => {
    render(<PriorityBadge priority={1} />)
    const badge = screen.getByText('Normal')
    expect(badge).toBeInTheDocument()
    expect(badge).toHaveClass('bg-blue-500/20')
    expect(badge).toHaveClass('text-blue-400')
  })

  it('renders High priority badge with correct styling', () => {
    render(<PriorityBadge priority={2} />)
    const badge = screen.getByText('High')
    expect(badge).toBeInTheDocument()
    expect(badge).toHaveClass('bg-yellow-500/20')
    expect(badge).toHaveClass('text-yellow-400')
  })

  it('renders Critical priority badge with correct styling', () => {
    render(<PriorityBadge priority={3} />)
    const badge = screen.getByText('Critical')
    expect(badge).toBeInTheDocument()
    expect(badge).toHaveClass('bg-red-500/20')
    expect(badge).toHaveClass('text-red-400')
  })

  it('clamps priority values above 3 to Critical', () => {
    render(<PriorityBadge priority={10} />)
    const badge = screen.getByText('Critical')
    expect(badge).toBeInTheDocument()
    expect(badge).toHaveClass('bg-red-500/20')
  })

  it('clamps negative priority values to Low', () => {
    render(<PriorityBadge priority={-5} />)
    const badge = screen.getByText('Low')
    expect(badge).toBeInTheDocument()
    expect(badge).toHaveClass('bg-gray-500/20')
  })

  it('applies correct base classes for badge styling', () => {
    render(<PriorityBadge priority={1} />)
    const badge = screen.getByText('Normal')
    expect(badge).toHaveClass('inline-flex')
    expect(badge).toHaveClass('px-2')
    expect(badge).toHaveClass('py-1')
    expect(badge).toHaveClass('text-xs')
    expect(badge).toHaveClass('font-medium')
    expect(badge).toHaveClass('rounded-full')
  })

  it('renders with optional size prop for larger badge', () => {
    render(<PriorityBadge priority={2} size="lg" />)
    const badge = screen.getByText('High')
    expect(badge).toHaveClass('px-3')
    expect(badge).toHaveClass('py-1')
    expect(badge).toHaveClass('text-sm')
  })

  it('renders with default size when size prop not provided', () => {
    render(<PriorityBadge priority={1} />)
    const badge = screen.getByText('Normal')
    expect(badge).toHaveClass('px-2')
    expect(badge).toHaveClass('text-xs')
  })
})
