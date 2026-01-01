import { describe, it, expect } from 'vitest'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { WorkflowSelectionProvider, useWorkflowSelection } from './WorkflowSelectionContext'

// Test component that uses the hook
function TestConsumer() {
  const {
    selectedWorkflowIds,
    toggleSelection,
    selectAll,
    clearSelection,
    isSelected,
  } = useWorkflowSelection()

  return (
    <div>
      <div data-testid="selected-count">{selectedWorkflowIds.length}</div>
      <div data-testid="selected-ids">{selectedWorkflowIds.join(',')}</div>
      <div data-testid="is-wf1-selected">{isSelected('wf-1').toString()}</div>
      <div data-testid="is-wf2-selected">{isSelected('wf-2').toString()}</div>
      <div data-testid="is-wf3-selected">{isSelected('wf-3').toString()}</div>
      <button onClick={() => toggleSelection('wf-1')}>Toggle WF1</button>
      <button onClick={() => toggleSelection('wf-2')}>Toggle WF2</button>
      <button onClick={() => toggleSelection('wf-3')}>Toggle WF3</button>
      <button onClick={() => selectAll(['wf-1', 'wf-2', 'wf-3'])}>Select All</button>
      <button onClick={() => clearSelection()}>Clear</button>
    </div>
  )
}

describe('WorkflowSelectionContext', () => {
  describe('useWorkflowSelection hook', () => {
    it('should throw error when used outside provider', () => {
      // Suppress console.error for this test
      const consoleSpy = vi.spyOn(console, 'error').mockImplementation(() => {})

      expect(() => {
        render(<TestConsumer />)
      }).toThrow('useWorkflowSelection must be used within WorkflowSelectionProvider')

      consoleSpy.mockRestore()
    })
  })

  describe('WorkflowSelectionProvider', () => {
    it('should start with empty selection', () => {
      render(
        <WorkflowSelectionProvider>
          <TestConsumer />
        </WorkflowSelectionProvider>
      )

      expect(screen.getByTestId('selected-count')).toHaveTextContent('0')
      expect(screen.getByTestId('selected-ids')).toHaveTextContent('')
    })

    it('should toggle selection on', async () => {
      const user = userEvent.setup()
      render(
        <WorkflowSelectionProvider>
          <TestConsumer />
        </WorkflowSelectionProvider>
      )

      await user.click(screen.getByText('Toggle WF1'))

      expect(screen.getByTestId('selected-count')).toHaveTextContent('1')
      expect(screen.getByTestId('selected-ids')).toHaveTextContent('wf-1')
      expect(screen.getByTestId('is-wf1-selected')).toHaveTextContent('true')
    })

    it('should toggle selection off', async () => {
      const user = userEvent.setup()
      render(
        <WorkflowSelectionProvider>
          <TestConsumer />
        </WorkflowSelectionProvider>
      )

      // Toggle on
      await user.click(screen.getByText('Toggle WF1'))
      expect(screen.getByTestId('is-wf1-selected')).toHaveTextContent('true')

      // Toggle off
      await user.click(screen.getByText('Toggle WF1'))
      expect(screen.getByTestId('is-wf1-selected')).toHaveTextContent('false')
      expect(screen.getByTestId('selected-count')).toHaveTextContent('0')
    })

    it('should handle multiple selections', async () => {
      const user = userEvent.setup()
      render(
        <WorkflowSelectionProvider>
          <TestConsumer />
        </WorkflowSelectionProvider>
      )

      await user.click(screen.getByText('Toggle WF1'))
      await user.click(screen.getByText('Toggle WF2'))
      await user.click(screen.getByText('Toggle WF3'))

      expect(screen.getByTestId('selected-count')).toHaveTextContent('3')
      expect(screen.getByTestId('is-wf1-selected')).toHaveTextContent('true')
      expect(screen.getByTestId('is-wf2-selected')).toHaveTextContent('true')
      expect(screen.getByTestId('is-wf3-selected')).toHaveTextContent('true')
    })

    it('should select all workflows', async () => {
      const user = userEvent.setup()
      render(
        <WorkflowSelectionProvider>
          <TestConsumer />
        </WorkflowSelectionProvider>
      )

      await user.click(screen.getByText('Select All'))

      expect(screen.getByTestId('selected-count')).toHaveTextContent('3')
      expect(screen.getByTestId('selected-ids')).toHaveTextContent('wf-1,wf-2,wf-3')
    })

    it('should clear all selections', async () => {
      const user = userEvent.setup()
      render(
        <WorkflowSelectionProvider>
          <TestConsumer />
        </WorkflowSelectionProvider>
      )

      // Select some workflows
      await user.click(screen.getByText('Toggle WF1'))
      await user.click(screen.getByText('Toggle WF2'))
      expect(screen.getByTestId('selected-count')).toHaveTextContent('2')

      // Clear all
      await user.click(screen.getByText('Clear'))
      expect(screen.getByTestId('selected-count')).toHaveTextContent('0')
      expect(screen.getByTestId('is-wf1-selected')).toHaveTextContent('false')
      expect(screen.getByTestId('is-wf2-selected')).toHaveTextContent('false')
    })

    it('should correctly report isSelected', async () => {
      const user = userEvent.setup()
      render(
        <WorkflowSelectionProvider>
          <TestConsumer />
        </WorkflowSelectionProvider>
      )

      // Initially all false
      expect(screen.getByTestId('is-wf1-selected')).toHaveTextContent('false')
      expect(screen.getByTestId('is-wf2-selected')).toHaveTextContent('false')
      expect(screen.getByTestId('is-wf3-selected')).toHaveTextContent('false')

      // Select wf-2 only
      await user.click(screen.getByText('Toggle WF2'))

      expect(screen.getByTestId('is-wf1-selected')).toHaveTextContent('false')
      expect(screen.getByTestId('is-wf2-selected')).toHaveTextContent('true')
      expect(screen.getByTestId('is-wf3-selected')).toHaveTextContent('false')
    })

    it('should replace selection when selectAll is called', async () => {
      const user = userEvent.setup()
      render(
        <WorkflowSelectionProvider>
          <TestConsumer />
        </WorkflowSelectionProvider>
      )

      // First toggle some
      await user.click(screen.getByText('Toggle WF1'))
      expect(screen.getByTestId('selected-ids')).toHaveTextContent('wf-1')

      // Select all should replace, not append
      await user.click(screen.getByText('Select All'))
      expect(screen.getByTestId('selected-ids')).toHaveTextContent('wf-1,wf-2,wf-3')
      expect(screen.getByTestId('selected-count')).toHaveTextContent('3')
    })
  })
})
