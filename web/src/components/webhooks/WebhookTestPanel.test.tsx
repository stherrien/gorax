import { describe, it, expect, beforeEach, vi, afterEach } from 'vitest'
import { render, screen, waitFor, within } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import userEvent from '@testing-library/user-event'
import WebhookTestPanel from './WebhookTestPanel'
import { webhookAPI } from '../../api/webhooks'
import type { TestWebhookResponse } from '../../api/webhooks'

// Mock the webhook API
vi.mock('../../api/webhooks', () => ({
  webhookAPI: {
    test: vi.fn(),
  },
}))

// Helper to render with router
const renderWithRouter = (component: React.ReactElement) => {
  return render(<MemoryRouter>{component}</MemoryRouter>)
}

describe('WebhookTestPanel', () => {
  const mockWebhookId = 'wh-123'
  const defaultProps = {
    webhookId: mockWebhookId,
  }

  beforeEach(() => {
    vi.clearAllMocks()
  })

  afterEach(() => {
    vi.clearAllMocks()
  })

  describe('Initial Render', () => {
    it('should render webhook test panel', () => {
      renderWithRouter(<WebhookTestPanel {...defaultProps} />)

      expect(screen.getByText(/webhook test/i)).toBeInTheDocument()
    })

    it('should show HTTP method selector with POST as default', () => {
      renderWithRouter(<WebhookTestPanel {...defaultProps} />)

      const methodSelect = screen.getByLabelText(/http method/i)
      expect(methodSelect).toBeInTheDocument()
      expect(methodSelect).toHaveValue('POST')
    })

    it('should show all HTTP methods in selector', () => {
      renderWithRouter(<WebhookTestPanel {...defaultProps} />)

      const methodSelect = screen.getByLabelText(/http method/i) as HTMLSelectElement
      const options = Array.from(methodSelect.options).map(opt => opt.value)

      expect(options).toContain('GET')
      expect(options).toContain('POST')
      expect(options).toContain('PUT')
      expect(options).toContain('DELETE')
      expect(options).toContain('PATCH')
    })

    it('should show JSON payload editor', () => {
      renderWithRouter(<WebhookTestPanel {...defaultProps} />)

      expect(screen.getByLabelText(/payload \(json\)/i)).toBeInTheDocument()
    })

    it('should show custom headers input', () => {
      renderWithRouter(<WebhookTestPanel {...defaultProps} />)

      expect(screen.getByText(/^headers$/i)).toBeInTheDocument()
    })

    it('should have send test request button', () => {
      renderWithRouter(<WebhookTestPanel {...defaultProps} />)

      expect(screen.getByRole('button', { name: /send test/i })).toBeInTheDocument()
    })

    it('should show sample payload templates dropdown', () => {
      renderWithRouter(<WebhookTestPanel {...defaultProps} />)

      expect(screen.getByLabelText(/sample payload/i)).toBeInTheDocument()
    })

    it('should not show response viewer initially', () => {
      renderWithRouter(<WebhookTestPanel {...defaultProps} />)

      expect(screen.queryByText(/response/i)).not.toBeInTheDocument()
      expect(screen.queryByText(/status code/i)).not.toBeInTheDocument()
    })
  })

  describe('HTTP Method Selection', () => {
    it('should allow changing HTTP method', async () => {
      const user = userEvent.setup()
      renderWithRouter(<WebhookTestPanel {...defaultProps} />)

      const methodSelect = screen.getByLabelText(/http method/i)
      await user.selectOptions(methodSelect, 'PUT')

      expect(methodSelect).toHaveValue('PUT')
    })

    it('should persist method selection', async () => {
      const user = userEvent.setup()
      renderWithRouter(<WebhookTestPanel {...defaultProps} />)

      const methodSelect = screen.getByLabelText(/http method/i) as HTMLSelectElement
      await user.selectOptions(methodSelect, 'DELETE')

      expect(methodSelect.value).toBe('DELETE')
    })
  })

  describe('JSON Payload Editor', () => {
    it('should allow entering JSON payload', async () => {
      const user = userEvent.setup()
      renderWithRouter(<WebhookTestPanel {...defaultProps} />)

      const payloadEditor = screen.getByLabelText(/payload \(json\)/i) as HTMLTextAreaElement
      const testPayload = '{"user": "john", "action": "login"}'

      await user.clear(payloadEditor)
      await user.click(payloadEditor)
      await user.paste(testPayload)

      expect(payloadEditor.value).toBe(testPayload)
    })

    it('should show JSON validation error for invalid JSON', async () => {
      const user = userEvent.setup()
      renderWithRouter(<WebhookTestPanel {...defaultProps} />)

      const payloadEditor = screen.getByLabelText(/payload \(json\)/i)
      await user.clear(payloadEditor)
      await user.click(payloadEditor)
      await user.paste('{invalid json}')

      const sendButton = screen.getByRole('button', { name: /send test/i })
      await user.click(sendButton)

      await waitFor(() => {
        expect(screen.getByText('Invalid JSON format')).toBeInTheDocument()
      })
    })

    it('should accept empty payload', async () => {
      const user = userEvent.setup()
      const mockResponse: TestWebhookResponse = {
        success: true,
        statusCode: 200,
        responseTimeMs: 100,
        executionId: 'exec-123',
      }
      vi.mocked(webhookAPI.test).mockResolvedValue(mockResponse)

      renderWithRouter(<WebhookTestPanel {...defaultProps} />)

      const sendButton = screen.getByRole('button', { name: /send test/i })
      await user.click(sendButton)

      await waitFor(() => {
        expect(webhookAPI.test).toHaveBeenCalled()
      })
    })

    it('should have monospace font for payload editor', () => {
      renderWithRouter(<WebhookTestPanel {...defaultProps} />)

      const payloadEditor = screen.getByLabelText(/payload \(json\)/i)
      expect(payloadEditor).toHaveClass('font-mono')
    })
  })

  describe('Custom Headers', () => {
    it('should allow adding header key-value pairs', async () => {
      const user = userEvent.setup()
      renderWithRouter(<WebhookTestPanel {...defaultProps} />)

      const addHeaderButton = screen.getByRole('button', { name: /add header/i })
      await user.click(addHeaderButton)

      expect(screen.getByPlaceholderText(/header name/i)).toBeInTheDocument()
      expect(screen.getByPlaceholderText(/header value/i)).toBeInTheDocument()
    })

    it('should allow entering header values', async () => {
      const user = userEvent.setup()
      renderWithRouter(<WebhookTestPanel {...defaultProps} />)

      const addHeaderButton = screen.getByRole('button', { name: /add header/i })
      await user.click(addHeaderButton)

      const headerName = screen.getByPlaceholderText(/header name/i)
      const headerValue = screen.getByPlaceholderText(/header value/i)

      await user.type(headerName, 'X-Custom-Header')
      await user.type(headerValue, 'custom-value')

      expect(headerName).toHaveValue('X-Custom-Header')
      expect(headerValue).toHaveValue('custom-value')
    })

    it('should allow removing header pairs', async () => {
      const user = userEvent.setup()
      renderWithRouter(<WebhookTestPanel {...defaultProps} />)

      const addHeaderButton = screen.getByRole('button', { name: /add header/i })
      await user.click(addHeaderButton)

      expect(screen.getByPlaceholderText(/header name/i)).toBeInTheDocument()

      const removeButton = screen.getByRole('button', { name: /remove header/i })
      await user.click(removeButton)

      expect(screen.queryByPlaceholderText(/header name/i)).not.toBeInTheDocument()
    })

    it('should allow multiple header pairs', async () => {
      const user = userEvent.setup()
      renderWithRouter(<WebhookTestPanel {...defaultProps} />)

      const addHeaderButton = screen.getByRole('button', { name: /add header/i })
      await user.click(addHeaderButton)
      await user.click(addHeaderButton)

      const headerNameInputs = screen.getAllByPlaceholderText(/header name/i)
      expect(headerNameInputs).toHaveLength(2)
    })
  })

  describe('Sample Payload Templates', () => {
    it('should show template options', () => {
      renderWithRouter(<WebhookTestPanel {...defaultProps} />)

      const templateSelect = screen.getByLabelText(/sample payload/i) as HTMLSelectElement
      const options = Array.from(templateSelect.options).map(opt => opt.text)

      expect(options).toContain('None')
      expect(options.length).toBeGreaterThan(1)
    })

    it('should have common webhook templates', () => {
      renderWithRouter(<WebhookTestPanel {...defaultProps} />)

      const templateSelect = screen.getByLabelText(/sample payload/i) as HTMLSelectElement
      const options = Array.from(templateSelect.options).map(opt => opt.text)

      expect(options).toContain('GitHub Push')
      expect(options).toContain('Stripe Payment')
      expect(options).toContain('Simple Object')
    })

    it('should populate payload editor when template selected', async () => {
      const user = userEvent.setup()
      renderWithRouter(<WebhookTestPanel {...defaultProps} />)

      const templateSelect = screen.getByLabelText(/sample payload/i)
      await user.selectOptions(templateSelect, 'simple')

      const payloadEditor = screen.getByLabelText(/payload \(json\)/i) as HTMLTextAreaElement
      expect(payloadEditor.value).toBeTruthy()
      expect(JSON.parse(payloadEditor.value)).toBeDefined()
    })

    it('should clear payload when None template selected', async () => {
      const user = userEvent.setup()
      renderWithRouter(<WebhookTestPanel {...defaultProps} />)

      const templateSelect = screen.getByLabelText(/sample payload/i)
      const payloadEditor = screen.getByLabelText(/payload \(json\)/i) as HTMLTextAreaElement

      await user.selectOptions(templateSelect, 'simple')
      expect(payloadEditor.value).toBeTruthy()

      await user.selectOptions(templateSelect, 'none')
      expect(payloadEditor.value).toBe('')
    })
  })

  describe('Send Test Request', () => {
    it('should send test request when button clicked', async () => {
      const user = userEvent.setup()
      const mockResponse: TestWebhookResponse = {
        success: true,
        statusCode: 200,
        responseTimeMs: 120,
        executionId: 'exec-123',
      }
      vi.mocked(webhookAPI.test).mockResolvedValue(mockResponse)

      renderWithRouter(<WebhookTestPanel {...defaultProps} />)

      const sendButton = screen.getByRole('button', { name: /send test/i })
      await user.click(sendButton)

      await waitFor(() => {
        expect(webhookAPI.test).toHaveBeenCalledWith(mockWebhookId, expect.any(Object))
      })
    })

    it('should send correct method in request', async () => {
      const user = userEvent.setup()
      const mockResponse: TestWebhookResponse = {
        success: true,
        statusCode: 200,
        responseTimeMs: 100,
      }
      vi.mocked(webhookAPI.test).mockResolvedValue(mockResponse)

      renderWithRouter(<WebhookTestPanel {...defaultProps} />)

      const methodSelect = screen.getByLabelText(/http method/i)
      await user.selectOptions(methodSelect, 'PUT')

      const sendButton = screen.getByRole('button', { name: /send test/i })
      await user.click(sendButton)

      await waitFor(() => {
        expect(webhookAPI.test).toHaveBeenCalledWith(
          mockWebhookId,
          expect.objectContaining({ method: 'PUT' })
        )
      })
    })

    it('should send payload in request', async () => {
      const user = userEvent.setup()
      const mockResponse: TestWebhookResponse = {
        success: true,
        statusCode: 200,
        responseTimeMs: 100,
      }
      vi.mocked(webhookAPI.test).mockResolvedValue(mockResponse)

      renderWithRouter(<WebhookTestPanel {...defaultProps} />)

      const payloadEditor = screen.getByLabelText(/payload \(json\)/i)
      const testPayload = '{"test": "data"}'
      await user.clear(payloadEditor)
      await user.click(payloadEditor)
      await user.paste(testPayload)

      const sendButton = screen.getByRole('button', { name: /send test/i })
      await user.click(sendButton)

      await waitFor(() => {
        expect(webhookAPI.test).toHaveBeenCalledWith(
          mockWebhookId,
          expect.objectContaining({ body: { test: 'data' } })
        )
      })
    })

    it('should send headers in request', async () => {
      const user = userEvent.setup()
      const mockResponse: TestWebhookResponse = {
        success: true,
        statusCode: 200,
        responseTimeMs: 100,
      }
      vi.mocked(webhookAPI.test).mockResolvedValue(mockResponse)

      renderWithRouter(<WebhookTestPanel {...defaultProps} />)

      const addHeaderButton = screen.getByRole('button', { name: /add header/i })
      await user.click(addHeaderButton)

      const headerName = screen.getByPlaceholderText(/header name/i)
      const headerValue = screen.getByPlaceholderText(/header value/i)
      await user.type(headerName, 'X-Test')
      await user.type(headerValue, 'test-value')

      const sendButton = screen.getByRole('button', { name: /send test/i })
      await user.click(sendButton)

      await waitFor(() => {
        expect(webhookAPI.test).toHaveBeenCalledWith(
          mockWebhookId,
          expect.objectContaining({
            headers: { 'X-Test': 'test-value' },
          })
        )
      })
    })

    it('should show loading state while sending', async () => {
      const user = userEvent.setup()
      vi.mocked(webhookAPI.test).mockImplementation(
        () => new Promise(resolve => setTimeout(() => resolve({
          success: true,
          statusCode: 200,
          responseTimeMs: 100,
        }), 100))
      )

      renderWithRouter(<WebhookTestPanel {...defaultProps} />)

      const sendButton = screen.getByRole('button', { name: /send test/i })
      await user.click(sendButton)

      expect(screen.getByText(/sending/i)).toBeInTheDocument()
      expect(sendButton).toBeDisabled()

      await waitFor(() => {
        expect(screen.queryByText(/sending/i)).not.toBeInTheDocument()
      })
    })

    it('should disable form fields while sending', async () => {
      const user = userEvent.setup()
      vi.mocked(webhookAPI.test).mockImplementation(
        () => new Promise(resolve => setTimeout(() => resolve({
          success: true,
          statusCode: 200,
          responseTimeMs: 100,
        }), 100))
      )

      renderWithRouter(<WebhookTestPanel {...defaultProps} />)

      const sendButton = screen.getByRole('button', { name: /send test/i })
      await user.click(sendButton)

      const methodSelect = screen.getByLabelText(/http method/i)
      const payloadEditor = screen.getByLabelText(/payload \(json\)/i)

      expect(methodSelect).toBeDisabled()
      expect(payloadEditor).toBeDisabled()
    })
  })

  describe('Response Viewer - Success', () => {
    it('should show response viewer after successful test', async () => {
      const user = userEvent.setup()
      const mockResponse: TestWebhookResponse = {
        success: true,
        statusCode: 200,
        responseTimeMs: 120,
        executionId: 'exec-123',
      }
      vi.mocked(webhookAPI.test).mockResolvedValue(mockResponse)

      renderWithRouter(<WebhookTestPanel {...defaultProps} />)

      const sendButton = screen.getByRole('button', { name: /send test/i })
      await user.click(sendButton)

      await waitFor(() => {
        expect(screen.getByText('Response')).toBeInTheDocument()
      })
    })

    it('should display status code with green color for 2xx', async () => {
      const user = userEvent.setup()
      const mockResponse: TestWebhookResponse = {
        success: true,
        statusCode: 201,
        responseTimeMs: 100,
      }
      vi.mocked(webhookAPI.test).mockResolvedValue(mockResponse)

      renderWithRouter(<WebhookTestPanel {...defaultProps} />)

      const sendButton = screen.getByRole('button', { name: /send test/i })
      await user.click(sendButton)

      await waitFor(() => {
        const statusCodes = screen.getAllByText(/201/i)
        expect(statusCodes[0]).toHaveClass('text-green-400')
      })
    })

    it('should display response time in milliseconds', async () => {
      const user = userEvent.setup()
      const mockResponse: TestWebhookResponse = {
        success: true,
        statusCode: 200,
        responseTimeMs: 150,
      }
      vi.mocked(webhookAPI.test).mockResolvedValue(mockResponse)

      renderWithRouter(<WebhookTestPanel {...defaultProps} />)

      const sendButton = screen.getByRole('button', { name: /send test/i })
      await user.click(sendButton)

      await waitFor(() => {
        expect(screen.getByText(/150\s*ms/i)).toBeInTheDocument()
      })
    })

    it('should display execution ID with link', async () => {
      const user = userEvent.setup()
      const mockResponse: TestWebhookResponse = {
        success: true,
        statusCode: 200,
        responseTimeMs: 100,
        executionId: 'exec-456',
      }
      vi.mocked(webhookAPI.test).mockResolvedValue(mockResponse)

      renderWithRouter(<WebhookTestPanel {...defaultProps} />)

      const sendButton = screen.getByRole('button', { name: /send test/i })
      await user.click(sendButton)

      await waitFor(() => {
        const executionLink = screen.getByRole('link', { name: /exec-456/i })
        expect(executionLink).toBeInTheDocument()
        expect(executionLink).toHaveAttribute('href', '/executions/exec-456')
      })
    })

    it('should not show execution ID if not provided', async () => {
      const user = userEvent.setup()
      const mockResponse: TestWebhookResponse = {
        success: true,
        statusCode: 200,
        responseTimeMs: 100,
      }
      vi.mocked(webhookAPI.test).mockResolvedValue(mockResponse)

      renderWithRouter(<WebhookTestPanel {...defaultProps} />)

      const sendButton = screen.getByRole('button', { name: /send test/i })
      await user.click(sendButton)

      await waitFor(() => {
        expect(screen.getByText('Response')).toBeInTheDocument()
      })

      expect(screen.queryByText(/execution id/i)).not.toBeInTheDocument()
    })
  })

  describe('Response Viewer - Errors', () => {
    it('should display status code with red color for 4xx', async () => {
      const user = userEvent.setup()
      const mockResponse: TestWebhookResponse = {
        success: false,
        statusCode: 400,
        responseTimeMs: 50,
        error: 'Bad request',
      }
      vi.mocked(webhookAPI.test).mockResolvedValue(mockResponse)

      renderWithRouter(<WebhookTestPanel {...defaultProps} />)

      const sendButton = screen.getByRole('button', { name: /send test/i })
      await user.click(sendButton)

      await waitFor(() => {
        const statusCodes = screen.getAllByText(/400/i)
        expect(statusCodes[0]).toHaveClass('text-red-400')
      })
    })

    it('should display status code with red color for 5xx', async () => {
      const user = userEvent.setup()
      const mockResponse: TestWebhookResponse = {
        success: false,
        statusCode: 500,
        responseTimeMs: 50,
        error: 'Server error',
      }
      vi.mocked(webhookAPI.test).mockResolvedValue(mockResponse)

      renderWithRouter(<WebhookTestPanel {...defaultProps} />)

      const sendButton = screen.getByRole('button', { name: /send test/i })
      await user.click(sendButton)

      await waitFor(() => {
        const statusCodes = screen.getAllByText(/500/i)
        expect(statusCodes[0]).toHaveClass('text-red-400')
      })
    })

    it('should display error message when test fails', async () => {
      const user = userEvent.setup()
      const mockResponse: TestWebhookResponse = {
        success: false,
        statusCode: 500,
        responseTimeMs: 50,
        error: 'Workflow execution failed: timeout',
      }
      vi.mocked(webhookAPI.test).mockResolvedValue(mockResponse)

      renderWithRouter(<WebhookTestPanel {...defaultProps} />)

      const sendButton = screen.getByRole('button', { name: /send test/i })
      await user.click(sendButton)

      await waitFor(() => {
        expect(screen.getByText(/workflow execution failed: timeout/i)).toBeInTheDocument()
      })
    })

    it('should handle network errors', async () => {
      const user = userEvent.setup()
      vi.mocked(webhookAPI.test).mockRejectedValue(new Error('Network error'))

      renderWithRouter(<WebhookTestPanel {...defaultProps} />)

      const sendButton = screen.getByRole('button', { name: /send test/i })
      await user.click(sendButton)

      await waitFor(() => {
        expect(screen.getByText(/network error/i)).toBeInTheDocument()
      })
    })
  })

  describe('Request/Response History', () => {
    it('should show history section', async () => {
      const user = userEvent.setup()
      const mockResponse: TestWebhookResponse = {
        success: true,
        statusCode: 200,
        responseTimeMs: 100,
      }
      vi.mocked(webhookAPI.test).mockResolvedValue(mockResponse)

      renderWithRouter(<WebhookTestPanel {...defaultProps} />)

      const sendButton = screen.getByRole('button', { name: /send test/i })
      await user.click(sendButton)

      await waitFor(() => {
        expect(screen.getByText(/history/i)).toBeInTheDocument()
      })
    })

    it('should add test to history after completion', async () => {
      const user = userEvent.setup()
      const mockResponse: TestWebhookResponse = {
        success: true,
        statusCode: 200,
        responseTimeMs: 100,
      }
      vi.mocked(webhookAPI.test).mockResolvedValue(mockResponse)

      renderWithRouter(<WebhookTestPanel {...defaultProps} />)

      const sendButton = screen.getByRole('button', { name: /send test/i })
      await user.click(sendButton)

      await waitFor(() => {
        const historySection = screen.getByText(/history/i).parentElement
        expect(historySection).toBeInTheDocument()
      })
    })

    it('should show last 5 tests in history', async () => {
      const user = userEvent.setup()
      const mockResponse: TestWebhookResponse = {
        success: true,
        statusCode: 200,
        responseTimeMs: 100,
      }
      vi.mocked(webhookAPI.test).mockResolvedValue(mockResponse)

      renderWithRouter(<WebhookTestPanel {...defaultProps} />)

      const sendButton = screen.getByRole('button', { name: /send test/i })

      // Send 6 test requests
      for (let i = 0; i < 6; i++) {
        await user.click(sendButton)
        await waitFor(() => {
          expect(screen.getByText('Response')).toBeInTheDocument()
        })
      }

      // Should only show 5 items in history
      const historySection = screen.getByText('History').parentElement
      const historyItems = within(historySection!).getAllByRole('listitem')
      expect(historyItems.length).toBeLessThanOrEqual(5)
    })

    it('should show timestamp for each history item', async () => {
      const user = userEvent.setup()
      const mockResponse: TestWebhookResponse = {
        success: true,
        statusCode: 200,
        responseTimeMs: 100,
      }
      vi.mocked(webhookAPI.test).mockResolvedValue(mockResponse)

      renderWithRouter(<WebhookTestPanel {...defaultProps} />)

      const sendButton = screen.getByRole('button', { name: /send test/i })
      await user.click(sendButton)

      await waitFor(() => {
        const historySection = screen.getByText(/history/i).parentElement
        const historyItem = within(historySection!).getByRole('listitem')
        // Should show relative time like "just now" or "1 minute ago"
        expect(historyItem.textContent).toMatch(/ago|now/i)
      })
    })

    it('should allow clicking history item to load that test', async () => {
      const user = userEvent.setup()
      const mockResponse: TestWebhookResponse = {
        success: true,
        statusCode: 200,
        responseTimeMs: 100,
      }
      vi.mocked(webhookAPI.test).mockResolvedValue(mockResponse)

      renderWithRouter(<WebhookTestPanel {...defaultProps} />)

      // Send test with specific payload
      const payloadEditor = screen.getByLabelText(/payload \(json\)/i)
      await user.clear(payloadEditor)
      await user.click(payloadEditor)
      await user.paste('{"first": "test"}')

      const sendButton = screen.getByRole('button', { name: /send test/i })
      await user.click(sendButton)

      await waitFor(() => {
        expect(screen.getByText('History')).toBeInTheDocument()
      })

      // Clear payload
      await user.clear(payloadEditor)
      expect(payloadEditor).toHaveValue('')

      // Click history item
      const historySection = screen.getByText('History').parentElement
      const historyItem = within(historySection!).getByRole('listitem')
      await user.click(historyItem)

      // Payload should be restored
      await waitFor(() => {
        expect(payloadEditor).toHaveValue(JSON.stringify({ first: 'test' }, null, 2))
      })
    })
  })

  describe('Accessibility', () => {
    it('should have proper labels for all inputs', () => {
      renderWithRouter(<WebhookTestPanel {...defaultProps} />)

      expect(screen.getByLabelText(/http method/i)).toBeInTheDocument()
      expect(screen.getByLabelText(/payload \(json\)/i)).toBeInTheDocument()
      expect(screen.getByLabelText(/sample payload/i)).toBeInTheDocument()
    })

    it('should have proper ARIA labels for buttons', () => {
      renderWithRouter(<WebhookTestPanel {...defaultProps} />)

      expect(screen.getByRole('button', { name: /send test/i })).toBeInTheDocument()
      expect(screen.getByRole('button', { name: /add header/i })).toBeInTheDocument()
    })

    it('should show validation errors with proper semantics', async () => {
      const user = userEvent.setup()
      renderWithRouter(<WebhookTestPanel {...defaultProps} />)

      const payloadEditor = screen.getByLabelText(/payload \(json\)/i)
      await user.clear(payloadEditor)
      await user.click(payloadEditor)
      await user.paste('invalid json')

      const sendButton = screen.getByRole('button', { name: /send test/i })
      await user.click(sendButton)

      await waitFor(() => {
        const errorMessage = screen.getByText('Invalid JSON format')
        expect(errorMessage).toBeInTheDocument()
        expect(errorMessage).toHaveClass('text-red-400')
      })
    })
  })

  describe('Dark Theme Styling', () => {
    it('should use dark background colors', () => {
      renderWithRouter(<WebhookTestPanel {...defaultProps} />)

      const panel = screen.getByText('Webhook Test').closest('.bg-gray-800')
      expect(panel).toBeInTheDocument()
    })

    it('should use white text', () => {
      renderWithRouter(<WebhookTestPanel {...defaultProps} />)

      const title = screen.getByText(/webhook test/i)
      expect(title).toHaveClass('text-white')
    })

    it('should use proper dark theme for inputs', () => {
      renderWithRouter(<WebhookTestPanel {...defaultProps} />)

      const payloadEditor = screen.getByLabelText(/payload \(json\)/i)
      expect(payloadEditor).toHaveClass('bg-gray-700')
      expect(payloadEditor).toHaveClass('text-white')
    })
  })
})
