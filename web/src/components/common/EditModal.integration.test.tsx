import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { useState } from 'react'
import { EditModal, EditField } from './EditModal'

/**
 * Integration tests for EditModal component.
 * These tests verify the complete edit workflow including:
 * - Opening/closing the modal
 * - Form state management
 * - Save/cancel flow
 * - Error handling
 * - Multi-field interactions
 */

interface Task {
  id: string
  title: string
  description: string
  status: 'pending' | 'in_progress' | 'completed'
  priority: number
  enabled: boolean
}

// Simulated dashboard component that uses EditModal
function TaskDashboard({
  initialTasks,
  onSave,
}: {
  initialTasks: Task[]
  onSave: (task: Task) => Promise<void>
}) {
  const [tasks, setTasks] = useState<Task[]>(initialTasks)
  const [editingTask, setEditingTask] = useState<Task | null>(null)

  const taskFields: EditField[] = [
    { name: 'title', label: 'Title', type: 'text', required: true, maxLength: 100 },
    { name: 'description', label: 'Description', type: 'textarea', maxLength: 500 },
    {
      name: 'status',
      label: 'Status',
      type: 'select',
      required: true,
      options: [
        { value: 'pending', label: 'Pending' },
        { value: 'in_progress', label: 'In Progress' },
        { value: 'completed', label: 'Completed' },
      ],
    },
    { name: 'priority', label: 'Priority', type: 'number', min: 1, max: 5 },
    { name: 'enabled', label: 'Enabled', type: 'checkbox' },
  ]

  const handleSave = async (values: Record<string, unknown>) => {
    if (!editingTask) return

    const updatedTask = { ...editingTask, ...values } as Task
    await onSave(updatedTask)

    setTasks((prev) => prev.map((t) => (t.id === updatedTask.id ? updatedTask : t)))
    setEditingTask(null)
  }

  return (
    <div data-testid="task-dashboard">
      <h1>Task Dashboard</h1>
      <ul data-testid="task-list">
        {tasks.map((task) => (
          <li key={task.id} data-testid={`task-item-${task.id}`}>
            <span data-testid={`task-title-${task.id}`}>{task.title}</span>
            <span data-testid={`task-status-${task.id}`}>{task.status}</span>
            <button
              data-testid={`edit-task-${task.id}`}
              onClick={() => setEditingTask(task)}
            >
              Edit
            </button>
          </li>
        ))}
      </ul>

      {editingTask && (
        <EditModal
          isOpen={true}
          title={`Edit Task: ${editingTask.title}`}
          fields={taskFields}
          initialValues={editingTask}
          onSave={handleSave}
          onCancel={() => setEditingTask(null)}
          itemId={editingTask.id}
        />
      )}
    </div>
  )
}

describe('EditModal Integration Tests', () => {
  const mockOnSave = vi.fn()

  const testTasks: Task[] = [
    {
      id: 'task-1',
      title: 'Complete documentation',
      description: 'Write comprehensive docs',
      status: 'pending',
      priority: 1,
      enabled: true,
    },
    {
      id: 'task-2',
      title: 'Review code',
      description: 'Code review for PR #123',
      status: 'in_progress',
      priority: 2,
      enabled: false,
    },
  ]

  beforeEach(() => {
    vi.clearAllMocks()
    mockOnSave.mockResolvedValue(undefined)
  })

  describe('Complete Edit Workflow', () => {
    it('should open modal when clicking edit button on a task', async () => {
      const user = userEvent.setup()

      render(<TaskDashboard initialTasks={testTasks} onSave={mockOnSave} />)

      // Verify modal is not open initially
      expect(screen.queryByRole('dialog')).not.toBeInTheDocument()

      // Click edit button
      const editButton = screen.getByTestId('edit-task-task-1')
      await user.click(editButton)

      // Verify modal opens with correct title
      const modal = screen.getByRole('dialog')
      expect(modal).toBeInTheDocument()
      expect(screen.getByText('Edit Task: Complete documentation')).toBeInTheDocument()
    })

    it('should display task data in form fields when modal opens', async () => {
      const user = userEvent.setup()

      render(<TaskDashboard initialTasks={testTasks} onSave={mockOnSave} />)

      await user.click(screen.getByTestId('edit-task-task-1'))

      expect(screen.getByLabelText(/Title/)).toHaveValue('Complete documentation')
      expect(screen.getByLabelText(/Description/)).toHaveValue('Write comprehensive docs')
      expect(screen.getByLabelText(/Status/)).toHaveValue('pending')
      expect(screen.getByLabelText(/Priority/)).toHaveValue(1)
      expect(screen.getByRole('switch', { name: /Enabled/i })).toBeChecked()
    })

    it('should update task and close modal on successful save', async () => {
      const user = userEvent.setup()

      render(<TaskDashboard initialTasks={testTasks} onSave={mockOnSave} />)

      await user.click(screen.getByTestId('edit-task-task-1'))

      // Change title
      const titleInput = screen.getByLabelText(/Title/)
      await user.clear(titleInput)
      await user.type(titleInput, 'Updated documentation task')

      // Change status
      await user.selectOptions(screen.getByLabelText(/Status/), 'completed')

      // Save
      await user.click(screen.getByRole('button', { name: /save/i }))

      await waitFor(() => {
        expect(mockOnSave).toHaveBeenCalledWith(
          expect.objectContaining({
            id: 'task-1',
            title: 'Updated documentation task',
            status: 'completed',
          })
        )
      })

      // Modal should close
      expect(screen.queryByRole('dialog')).not.toBeInTheDocument()

      // Dashboard should show updated task
      expect(screen.getByTestId('task-title-task-1')).toHaveTextContent('Updated documentation task')
      expect(screen.getByTestId('task-status-task-1')).toHaveTextContent('completed')
    })

    it('should close modal without saving when cancel is clicked', async () => {
      const user = userEvent.setup()

      render(<TaskDashboard initialTasks={testTasks} onSave={mockOnSave} />)

      await user.click(screen.getByTestId('edit-task-task-1'))

      // Make changes
      const titleInput = screen.getByLabelText(/Title/)
      await user.clear(titleInput)
      await user.type(titleInput, 'This should not be saved')

      // Cancel
      await user.click(screen.getByRole('button', { name: /cancel/i }))

      // Modal should close
      expect(screen.queryByRole('dialog')).not.toBeInTheDocument()

      // onSave should not have been called
      expect(mockOnSave).not.toHaveBeenCalled()

      // Original task title should be preserved
      expect(screen.getByTestId('task-title-task-1')).toHaveTextContent('Complete documentation')
    })
  })

  describe('Error Handling Workflow', () => {
    it('should display validation errors and prevent save', async () => {
      const user = userEvent.setup()

      render(<TaskDashboard initialTasks={testTasks} onSave={mockOnSave} />)

      await user.click(screen.getByTestId('edit-task-task-1'))

      // Clear required field
      const titleInput = screen.getByLabelText(/Title/)
      await user.clear(titleInput)

      // Try to save
      await user.click(screen.getByRole('button', { name: /save/i }))

      // Validation error should appear
      expect(screen.getByText('Title is required')).toBeInTheDocument()
      expect(mockOnSave).not.toHaveBeenCalled()

      // Modal should still be open
      expect(screen.getByRole('dialog')).toBeInTheDocument()
    })

    it('should display API error and keep modal open on save failure', async () => {
      const user = userEvent.setup()
      mockOnSave.mockRejectedValueOnce(new Error('Server error: Unable to update task'))

      render(<TaskDashboard initialTasks={testTasks} onSave={mockOnSave} />)

      await user.click(screen.getByTestId('edit-task-task-1'))

      await user.click(screen.getByRole('button', { name: /save/i }))

      await waitFor(() => {
        expect(screen.getByText('Server error: Unable to update task')).toBeInTheDocument()
      })

      // Modal should still be open
      expect(screen.getByRole('dialog')).toBeInTheDocument()

      // User can retry
      expect(screen.getByRole('button', { name: /save/i })).not.toBeDisabled()
    })

    it('should clear errors after user fixes validation issue', async () => {
      const user = userEvent.setup()

      render(<TaskDashboard initialTasks={testTasks} onSave={mockOnSave} />)

      await user.click(screen.getByTestId('edit-task-task-1'))

      // Clear required field
      const titleInput = screen.getByLabelText(/Title/)
      await user.clear(titleInput)

      // Try to save - should show error
      await user.click(screen.getByRole('button', { name: /save/i }))
      expect(screen.getByText('Title is required')).toBeInTheDocument()

      // Fix the issue
      await user.type(titleInput, 'Fixed title')

      // Error should be cleared
      expect(screen.queryByText('Title is required')).not.toBeInTheDocument()

      // Now save should work
      await user.click(screen.getByRole('button', { name: /save/i }))

      await waitFor(() => {
        expect(mockOnSave).toHaveBeenCalled()
      })
    })
  })

  describe('Multi-Task Editing', () => {
    it('should switch between editing different tasks correctly', async () => {
      const user = userEvent.setup()

      render(<TaskDashboard initialTasks={testTasks} onSave={mockOnSave} />)

      // Edit first task
      await user.click(screen.getByTestId('edit-task-task-1'))
      expect(screen.getByLabelText(/Title/)).toHaveValue('Complete documentation')

      // Cancel and edit second task
      await user.click(screen.getByRole('button', { name: /cancel/i }))
      await user.click(screen.getByTestId('edit-task-task-2'))

      // Should show second task data
      expect(screen.getByLabelText(/Title/)).toHaveValue('Review code')
      expect(screen.getByLabelText(/Status/)).toHaveValue('in_progress')
      expect(screen.getByRole('switch', { name: /Enabled/i })).not.toBeChecked()
    })

    it('should maintain separate state for different tasks', async () => {
      const user = userEvent.setup()

      render(<TaskDashboard initialTasks={testTasks} onSave={mockOnSave} />)

      // Edit first task and save
      await user.click(screen.getByTestId('edit-task-task-1'))
      const titleInput = screen.getByLabelText(/Title/)
      await user.clear(titleInput)
      await user.type(titleInput, 'Updated task 1')
      await user.click(screen.getByRole('button', { name: /save/i }))

      await waitFor(() => {
        expect(screen.queryByRole('dialog')).not.toBeInTheDocument()
      })

      // Edit second task - should show its original data
      await user.click(screen.getByTestId('edit-task-task-2'))
      expect(screen.getByLabelText(/Title/)).toHaveValue('Review code')

      // Verify first task was updated
      expect(screen.getByTestId('task-title-task-1')).toHaveTextContent('Updated task 1')
      // Second task unchanged
      expect(screen.getByTestId('task-title-task-2')).toHaveTextContent('Review code')
    })
  })

  describe('Field Interactions', () => {
    it('should handle all field types correctly in a single save', async () => {
      const user = userEvent.setup()

      render(<TaskDashboard initialTasks={testTasks} onSave={mockOnSave} />)

      await user.click(screen.getByTestId('edit-task-task-1'))

      // Update text field
      const titleInput = screen.getByLabelText(/Title/)
      await user.clear(titleInput)
      await user.type(titleInput, 'Multi-field test')

      // Update textarea
      const descInput = screen.getByLabelText(/Description/)
      await user.clear(descInput)
      await user.type(descInput, 'Testing all fields')

      // Update select
      await user.selectOptions(screen.getByLabelText(/Status/), 'completed')

      // Update number
      const priorityInput = screen.getByLabelText(/Priority/)
      await user.clear(priorityInput)
      await user.type(priorityInput, '5')

      // Toggle checkbox
      await user.click(screen.getByRole('switch', { name: /Enabled/i }))

      // Save
      await user.click(screen.getByRole('button', { name: /save/i }))

      await waitFor(() => {
        expect(mockOnSave).toHaveBeenCalledWith(
          expect.objectContaining({
            id: 'task-1',
            title: 'Multi-field test',
            description: 'Testing all fields',
            status: 'completed',
            priority: 5,
            enabled: false,
          })
        )
      })
    })
  })

  describe('Concurrent Operations', () => {
    it('should prevent closing during save operation', async () => {
      const user = userEvent.setup()

      // Slow save operation
      mockOnSave.mockImplementation(
        () => new Promise((resolve) => setTimeout(resolve, 200))
      )

      render(<TaskDashboard initialTasks={testTasks} onSave={mockOnSave} />)

      await user.click(screen.getByTestId('edit-task-task-1'))
      await user.click(screen.getByRole('button', { name: /save/i }))

      // Try to click backdrop during save
      const backdrop = screen.getByTestId('edit-modal-backdrop')
      await user.click(backdrop)

      // Modal should still be open (during save)
      expect(screen.getByRole('dialog')).toBeInTheDocument()

      // Wait for save to complete
      await waitFor(() => {
        expect(screen.queryByRole('dialog')).not.toBeInTheDocument()
      })
    })
  })
})
