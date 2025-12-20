import { describe, it, expect, vi } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import WebhookEditForm from './WebhookEditForm'
import type { Webhook } from '../../api/webhooks'

describe('WebhookEditForm', () => {
  const mockWebhook: Webhook = {
    id: 'wh-1',
    tenantId: 'tenant-1',
    workflowId: 'wf-1',
    name: 'Test Webhook',
    path: '/test',
    authType: 'signature',
    enabled: true,
    priority: 1,
    triggerCount: 10,
    createdAt: '2024-01-01T00:00:00Z',
    updatedAt: '2024-01-01T00:00:00Z',
    url: 'https://api.example.com/test',
  }

  it('renders form with webhook data', () => {
    const onSave = vi.fn()
    const onCancel = vi.fn()

    render(<WebhookEditForm webhook={mockWebhook} onSave={onSave} onCancel={onCancel} />)

    expect(screen.getByLabelText(/name/i)).toHaveValue('Test Webhook')
    expect(screen.getByLabelText(/path/i)).toHaveValue('/test')
    expect(screen.getByLabelText(/priority/i)).toHaveValue('1')
  })

  it('renders priority selector with correct value', () => {
    const onSave = vi.fn()
    const onCancel = vi.fn()

    render(
      <WebhookEditForm
        webhook={{ ...mockWebhook, priority: 3 }}
        onSave={onSave}
        onCancel={onCancel}
      />
    )

    const prioritySelect = screen.getByLabelText(/priority/i) as HTMLSelectElement
    expect(prioritySelect.value).toBe('3')
  })

  it('allows changing priority value', async () => {
    const user = userEvent.setup()
    const onSave = vi.fn()
    const onCancel = vi.fn()

    render(<WebhookEditForm webhook={mockWebhook} onSave={onSave} onCancel={onCancel} />)

    const prioritySelect = screen.getByLabelText(/priority/i)
    await user.selectOptions(prioritySelect, '2')

    expect((prioritySelect as HTMLSelectElement).value).toBe('2')
  })

  it('calls onSave with updated priority when form is submitted', async () => {
    const user = userEvent.setup()
    const onSave = vi.fn()
    const onCancel = vi.fn()

    render(<WebhookEditForm webhook={mockWebhook} onSave={onSave} onCancel={onCancel} />)

    const prioritySelect = screen.getByLabelText(/priority/i)
    await user.selectOptions(prioritySelect, '3')

    const saveButton = screen.getByRole('button', { name: /save/i })
    await user.click(saveButton)

    await waitFor(() => {
      expect(onSave).toHaveBeenCalledWith(
        expect.objectContaining({
          priority: 3,
        })
      )
    })
  })

  it('calls onCancel when cancel button is clicked', async () => {
    const user = userEvent.setup()
    const onSave = vi.fn()
    const onCancel = vi.fn()

    render(<WebhookEditForm webhook={mockWebhook} onSave={onSave} onCancel={onCancel} />)

    const cancelButton = screen.getByRole('button', { name: /cancel/i })
    await user.click(cancelButton)

    expect(onCancel).toHaveBeenCalled()
  })

  it('disables form during save operation', async () => {
    const user = userEvent.setup()
    const onSave = vi.fn().mockImplementation(() => new Promise((resolve) => setTimeout(resolve, 100)))
    const onCancel = vi.fn()

    render(<WebhookEditForm webhook={mockWebhook} onSave={onSave} onCancel={onCancel} />)

    const saveButton = screen.getByRole('button', { name: /save/i })
    await user.click(saveButton)

    expect(screen.getByLabelText(/priority/i)).toBeDisabled()
    expect(screen.getByRole('button', { name: /saving/i })).toBeDisabled()
  })

  it('validates required fields before save', async () => {
    const user = userEvent.setup()
    const onSave = vi.fn()
    const onCancel = vi.fn()

    render(<WebhookEditForm webhook={mockWebhook} onSave={onSave} onCancel={onCancel} />)

    const nameInput = screen.getByLabelText(/name/i)
    await user.clear(nameInput)

    const saveButton = screen.getByRole('button', { name: /save/i })
    await user.click(saveButton)

    expect(screen.getByText(/name is required/i)).toBeInTheDocument()
    expect(onSave).not.toHaveBeenCalled()
  })

  it('shows error message when save fails', async () => {
    const user = userEvent.setup()
    const onSave = vi.fn().mockRejectedValue(new Error('Save failed'))
    const onCancel = vi.fn()

    render(<WebhookEditForm webhook={mockWebhook} onSave={onSave} onCancel={onCancel} />)

    const saveButton = screen.getByRole('button', { name: /save/i })
    await user.click(saveButton)

    await waitFor(() => {
      expect(screen.getByText(/save failed/i)).toBeInTheDocument()
    })
  })

  it('renders enabled toggle switch', () => {
    const onSave = vi.fn()
    const onCancel = vi.fn()

    render(<WebhookEditForm webhook={mockWebhook} onSave={onSave} onCancel={onCancel} />)

    const toggle = screen.getByRole('switch')
    expect(toggle).toBeChecked()
  })

  it('allows toggling enabled state', async () => {
    const user = userEvent.setup()
    const onSave = vi.fn()
    const onCancel = vi.fn()

    render(<WebhookEditForm webhook={mockWebhook} onSave={onSave} onCancel={onCancel} />)

    const toggle = screen.getByRole('switch')
    await user.click(toggle)

    expect(toggle).not.toBeChecked()
  })

  it('includes priority in saved data when toggling enabled', async () => {
    const user = userEvent.setup()
    const onSave = vi.fn()
    const onCancel = vi.fn()

    render(
      <WebhookEditForm
        webhook={{ ...mockWebhook, priority: 2 }}
        onSave={onSave}
        onCancel={onCancel}
      />
    )

    const toggle = screen.getByRole('switch')
    await user.click(toggle)

    const saveButton = screen.getByRole('button', { name: /save/i })
    await user.click(saveButton)

    await waitFor(() => {
      expect(onSave).toHaveBeenCalledWith(
        expect.objectContaining({
          priority: 2,
          enabled: false,
        })
      )
    })
  })
})
