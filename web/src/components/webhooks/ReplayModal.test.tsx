import { describe, it, expect, beforeEach, vi } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import userEvent from '@testing-library/user-event'
import { ReplayModal } from './ReplayModal'
import type { WebhookEvent } from '../../api/webhooks'
import { webhookAPI } from '../../api/webhooks'

// Mock the webhooks API
vi.mock('../../api/webhooks', () => ({
  webhookAPI: {
    replayEvent: vi.fn(),
  },
}))

// Helper to wrap component with router
const renderWithRouter = (component: React.ReactElement) => {
  return render(<MemoryRouter>{component}</MemoryRouter>)
}

describe('ReplayModal', () => {
  const mockEvent: WebhookEvent = {
    id: 'evt-1',
    webhookId: 'wh-1',
    executionId: 'exec-1',
    requestMethod: 'POST',
    requestHeaders: {
      'Content-Type': 'application/json',
    },
    requestBody: { data: 'test payload', userId: 123 },
    responseStatus: 200,
    processingTimeMs: 120,
    status: 'processed',
    replayCount: 2,
    createdAt: '2024-01-15T10:00:00Z',
  }

  const mockOnClose = vi.fn()
  const mockOnSuccess = vi.fn()

  beforeEach(() => {
    vi.clearAllMocks()
  })

  describe('Initial Render', () => {
    it('should render modal with title', () => {
      renderWithRouter(
        <ReplayModal event={mockEvent} onClose={mockOnClose} onSuccess={mockOnSuccess} />
      )

      expect(screen.getByRole('dialog')).toBeInTheDocument()
      expect(screen.getByText(/replay webhook event/i)).toBeInTheDocument()
    })

    it('should display event ID', () => {
      renderWithRouter(
        <ReplayModal event={mockEvent} onClose={mockOnClose} onSuccess={mockOnSuccess} />
      )

      expect(screen.getByText(/evt-1/i)).toBeInTheDocument()
    })

    it('should display replay count', () => {
      renderWithRouter(
        <ReplayModal event={mockEvent} onClose={mockOnClose} onSuccess={mockOnSuccess} />
      )

      expect(screen.getByText(/replayed 2 times/i)).toBeInTheDocument()
    })

    it('should show payload editor with original payload', () => {
      renderWithRouter(
        <ReplayModal event={mockEvent} onClose={mockOnClose} onSuccess={mockOnSuccess} />
      )

      const editor = screen.getByLabelText(/payload/i) as HTMLTextAreaElement
      expect(editor).toBeInTheDocument()
      expect(JSON.parse(editor.value)).toEqual(mockEvent.requestBody)
    })

    it('should have cancel and replay buttons', () => {
      renderWithRouter(
        <ReplayModal event={mockEvent} onClose={mockOnClose} onSuccess={mockOnSuccess} />
      )

      expect(screen.getByRole('button', { name: /cancel/i })).toBeInTheDocument()
      expect(screen.getByRole('button', { name: /replay event/i })).toBeInTheDocument()
    })

    it('should display close button in header', () => {
      renderWithRouter(
        <ReplayModal event={mockEvent} onClose={mockOnClose} onSuccess={mockOnSuccess} />
      )

      const closeButton = screen.getByLabelText(/close/i)
      expect(closeButton).toBeInTheDocument()
    })
  })

  describe('Payload Editing', () => {
    it('should allow editing payload', async () => {
      const user = userEvent.setup()
      renderWithRouter(
        <ReplayModal event={mockEvent} onClose={mockOnClose} onSuccess={mockOnSuccess} />
      )

      const editor = screen.getByLabelText(/payload/i) as HTMLTextAreaElement
      await user.clear(editor)
      await user.click(editor)
      await user.paste('{"modified": "payload"}')

      expect(editor.value).toBe('{"modified": "payload"}')
    })

    it('should show validation error for invalid JSON', async () => {
      const user = userEvent.setup()
      renderWithRouter(
        <ReplayModal event={mockEvent} onClose={mockOnClose} onSuccess={mockOnSuccess} />
      )

      const editor = screen.getByLabelText(/payload/i)
      await user.clear(editor)
      await user.click(editor)
      await user.paste('{invalid json}')

      const replayButton = screen.getByRole('button', { name: /replay event/i })
      await user.click(replayButton)

      await waitFor(() => {
        expect(screen.getByText('Invalid JSON format')).toBeInTheDocument()
      })
    })

    it('should format JSON with proper indentation', () => {
      renderWithRouter(
        <ReplayModal event={mockEvent} onClose={mockOnClose} onSuccess={mockOnSuccess} />
      )

      const editor = screen.getByLabelText(/payload/i) as HTMLTextAreaElement
      expect(editor.value).toContain('  ') // Should have 2-space indentation
    })

    it('should have monospace font for payload editor', () => {
      renderWithRouter(
        <ReplayModal event={mockEvent} onClose={mockOnClose} onSuccess={mockOnSuccess} />
      )

      const editor = screen.getByLabelText(/payload/i)
      expect(editor).toHaveClass('font-mono')
    })
  })

  describe('Replay Functionality', () => {
    it('should call replayEvent API with original payload when replaying without modification', async () => {
      const user = userEvent.setup()
      vi.mocked(webhookAPI.replayEvent).mockResolvedValue({
        success: true,
        executionId: 'exec-new',
      })

      renderWithRouter(
        <ReplayModal event={mockEvent} onClose={mockOnClose} onSuccess={mockOnSuccess} />
      )

      const replayButton = screen.getByRole('button', { name: /replay event/i })
      await user.click(replayButton)

      await waitFor(() => {
        expect(webhookAPI.replayEvent).toHaveBeenCalledWith('evt-1', mockEvent.requestBody)
      })
    })

    it('should call replayEvent API with modified payload when edited', async () => {
      const user = userEvent.setup()
      vi.mocked(webhookAPI.replayEvent).mockResolvedValue({
        success: true,
        executionId: 'exec-new',
      })

      renderWithRouter(
        <ReplayModal event={mockEvent} onClose={mockOnClose} onSuccess={mockOnSuccess} />
      )

      const editor = screen.getByLabelText(/payload/i)
      const modifiedPayload = { modified: 'data' }
      await user.clear(editor)
      await user.click(editor)
      await user.paste(JSON.stringify(modifiedPayload))

      const replayButton = screen.getByRole('button', { name: /replay event/i })
      await user.click(replayButton)

      await waitFor(() => {
        expect(webhookAPI.replayEvent).toHaveBeenCalledWith('evt-1', modifiedPayload)
      })
    })

    it('should show loading state during replay', async () => {
      const user = userEvent.setup()
      vi.mocked(webhookAPI.replayEvent).mockImplementation(
        () => new Promise(resolve => setTimeout(() => resolve({ success: true }), 100))
      )

      renderWithRouter(
        <ReplayModal event={mockEvent} onClose={mockOnClose} onSuccess={mockOnSuccess} />
      )

      const replayButton = screen.getByRole('button', { name: /replay event/i })
      await user.click(replayButton)

      expect(screen.getByRole('button', { name: /replaying/i })).toBeInTheDocument()
      expect(replayButton).toBeDisabled()

      await waitFor(() => {
        expect(screen.queryByRole('button', { name: /replaying/i })).not.toBeInTheDocument()
      })
    })

    it('should disable all buttons during replay', async () => {
      const user = userEvent.setup()
      vi.mocked(webhookAPI.replayEvent).mockImplementation(
        () => new Promise(resolve => setTimeout(() => resolve({ success: true }), 100))
      )

      renderWithRouter(
        <ReplayModal event={mockEvent} onClose={mockOnClose} onSuccess={mockOnSuccess} />
      )

      const replayButton = screen.getByRole('button', { name: /replay event/i })
      await user.click(replayButton)

      const cancelButton = screen.getByRole('button', { name: /cancel/i })
      expect(cancelButton).toBeDisabled()
      expect(replayButton).toBeDisabled()
    })

    it('should call onSuccess callback after successful replay', async () => {
      const user = userEvent.setup()
      vi.mocked(webhookAPI.replayEvent).mockResolvedValue({
        success: true,
        executionId: 'exec-new',
      })

      renderWithRouter(
        <ReplayModal event={mockEvent} onClose={mockOnClose} onSuccess={mockOnSuccess} />
      )

      const replayButton = screen.getByRole('button', { name: /replay event/i })
      await user.click(replayButton)

      await waitFor(() => {
        expect(mockOnSuccess).toHaveBeenCalledWith('exec-new')
      })
    })

    it('should close modal after successful replay', async () => {
      const user = userEvent.setup()
      vi.mocked(webhookAPI.replayEvent).mockResolvedValue({
        success: true,
        executionId: 'exec-new',
      })

      renderWithRouter(
        <ReplayModal event={mockEvent} onClose={mockOnClose} onSuccess={mockOnSuccess} />
      )

      const replayButton = screen.getByRole('button', { name: /replay event/i })
      await user.click(replayButton)

      await waitFor(() => {
        expect(mockOnClose).toHaveBeenCalled()
      })
    })

    it('should display error message if replay fails', async () => {
      const user = userEvent.setup()
      vi.mocked(webhookAPI.replayEvent).mockResolvedValue({
        success: false,
        error: 'Max replay count exceeded',
      })

      renderWithRouter(
        <ReplayModal event={mockEvent} onClose={mockOnClose} onSuccess={mockOnSuccess} />
      )

      const replayButton = screen.getByRole('button', { name: /replay event/i })
      await user.click(replayButton)

      await waitFor(() => {
        expect(screen.getByText(/max replay count exceeded/i)).toBeInTheDocument()
      })
    })

    it('should handle network errors', async () => {
      const user = userEvent.setup()
      vi.mocked(webhookAPI.replayEvent).mockRejectedValue(new Error('Network error'))

      renderWithRouter(
        <ReplayModal event={mockEvent} onClose={mockOnClose} onSuccess={mockOnSuccess} />
      )

      const replayButton = screen.getByRole('button', { name: /replay event/i })
      await user.click(replayButton)

      await waitFor(() => {
        expect(screen.getByText(/network error/i)).toBeInTheDocument()
      })
    })

    it('should not close modal after failed replay', async () => {
      const user = userEvent.setup()
      vi.mocked(webhookAPI.replayEvent).mockResolvedValue({
        success: false,
        error: 'Replay failed',
      })

      renderWithRouter(
        <ReplayModal event={mockEvent} onClose={mockOnClose} onSuccess={mockOnSuccess} />
      )

      const replayButton = screen.getByRole('button', { name: /replay event/i })
      await user.click(replayButton)

      await waitFor(() => {
        expect(screen.getByText(/replay failed/i)).toBeInTheDocument()
      })

      expect(mockOnClose).not.toHaveBeenCalled()
    })
  })

  describe('Modal Close Behavior', () => {
    it('should close modal when clicking cancel button', async () => {
      const user = userEvent.setup()
      renderWithRouter(
        <ReplayModal event={mockEvent} onClose={mockOnClose} onSuccess={mockOnSuccess} />
      )

      const cancelButton = screen.getByRole('button', { name: /cancel/i })
      await user.click(cancelButton)

      expect(mockOnClose).toHaveBeenCalled()
    })

    it('should close modal when clicking close button', async () => {
      const user = userEvent.setup()
      renderWithRouter(
        <ReplayModal event={mockEvent} onClose={mockOnClose} onSuccess={mockOnSuccess} />
      )

      const closeButton = screen.getByLabelText(/close/i)
      await user.click(closeButton)

      expect(mockOnClose).toHaveBeenCalled()
    })

    it('should close modal when clicking backdrop', async () => {
      const user = userEvent.setup()
      renderWithRouter(
        <ReplayModal event={mockEvent} onClose={mockOnClose} onSuccess={mockOnSuccess} />
      )

      const backdrop = screen.getByTestId('replay-modal-backdrop')
      await user.click(backdrop)

      expect(mockOnClose).toHaveBeenCalled()
    })

    it('should not close modal when clicking inside modal content', async () => {
      const user = userEvent.setup()
      renderWithRouter(
        <ReplayModal event={mockEvent} onClose={mockOnClose} onSuccess={mockOnSuccess} />
      )

      const modal = screen.getByRole('dialog')
      await user.click(modal)

      expect(mockOnClose).not.toHaveBeenCalled()
    })
  })

  describe('Max Replay Warning', () => {
    it('should show warning when replay count is at limit', () => {
      const eventAtLimit: WebhookEvent = {
        ...mockEvent,
        replayCount: 5,
      }

      renderWithRouter(
        <ReplayModal event={eventAtLimit} onClose={mockOnClose} onSuccess={mockOnSuccess} />
      )

      expect(screen.getByText(/cannot replay/i)).toBeInTheDocument()
      expect(screen.getByText(/maximum replay limit/i)).toBeInTheDocument()
    })

    it('should disable replay button when at max replay count', () => {
      const eventAtLimit: WebhookEvent = {
        ...mockEvent,
        replayCount: 5,
      }

      renderWithRouter(
        <ReplayModal event={eventAtLimit} onClose={mockOnClose} onSuccess={mockOnSuccess} />
      )

      const replayButton = screen.getByRole('button', { name: /replay event/i })
      expect(replayButton).toBeDisabled()
    })

    it('should show warning approaching limit when replay count is 4', () => {
      const eventNearLimit: WebhookEvent = {
        ...mockEvent,
        replayCount: 4,
      }

      renderWithRouter(
        <ReplayModal event={eventNearLimit} onClose={mockOnClose} onSuccess={mockOnSuccess} />
      )

      expect(screen.getByText(/1 replay remaining/i)).toBeInTheDocument()
    })
  })

  describe('Accessibility', () => {
    it('should have proper ARIA attributes', () => {
      renderWithRouter(
        <ReplayModal event={mockEvent} onClose={mockOnClose} onSuccess={mockOnSuccess} />
      )

      const dialog = screen.getByRole('dialog')
      expect(dialog).toHaveAttribute('aria-modal', 'true')
      expect(dialog).toHaveAttribute('aria-labelledby')
    })

    it('should have proper label for payload editor', () => {
      renderWithRouter(
        <ReplayModal event={mockEvent} onClose={mockOnClose} onSuccess={mockOnSuccess} />
      )

      expect(screen.getByLabelText(/payload/i)).toBeInTheDocument()
    })

    it('should have close button with aria-label', () => {
      renderWithRouter(
        <ReplayModal event={mockEvent} onClose={mockOnClose} onSuccess={mockOnSuccess} />
      )

      const closeButton = screen.getByLabelText(/close/i)
      expect(closeButton).toHaveAttribute('aria-label')
    })
  })

  describe('Dark Theme Styling', () => {
    it('should use dark background colors', () => {
      renderWithRouter(
        <ReplayModal event={mockEvent} onClose={mockOnClose} onSuccess={mockOnSuccess} />
      )

      const modal = screen.getByRole('dialog')
      expect(modal).toHaveClass('bg-gray-800')
    })

    it('should use white text for title', () => {
      renderWithRouter(
        <ReplayModal event={mockEvent} onClose={mockOnClose} onSuccess={mockOnSuccess} />
      )

      const title = screen.getByText(/replay webhook event/i)
      expect(title).toHaveClass('text-white')
    })

    it('should use dark theme for payload editor', () => {
      renderWithRouter(
        <ReplayModal event={mockEvent} onClose={mockOnClose} onSuccess={mockOnSuccess} />
      )

      const editor = screen.getByLabelText(/payload/i)
      expect(editor).toHaveClass('bg-gray-900')
      expect(editor).toHaveClass('text-white')
    })
  })
})
