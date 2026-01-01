import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import { CredentialForm } from './CredentialForm'
import { ThemeProvider } from '../../contexts/ThemeContext'
import type { Credential, CredentialType } from '../../api/credentials'

const renderWithTheme = (ui: React.ReactElement) => {
  return render(<ThemeProvider>{ui}</ThemeProvider>)
}

describe('CredentialForm', () => {
  const mockOnSubmit = vi.fn()
  const mockOnCancel = vi.fn()

  const mockExistingCredential: Credential = {
    id: 'cred-123',
    tenantId: 'tenant-1',
    name: 'Test Credential',
    type: 'api_key',
    description: 'Test description',
    createdAt: '2024-01-15T10:00:00Z',
    updatedAt: '2024-01-15T10:00:00Z',
  }

  beforeEach(() => {
    vi.clearAllMocks()
  })

  describe('Create mode', () => {
    it('renders create form', () => {
      renderWithTheme(<CredentialForm onSubmit={mockOnSubmit} onCancel={mockOnCancel} />)

      expect(screen.getByText('Create Credential')).toBeInTheDocument()
      expect(screen.getByLabelText(/Name/i)).toBeInTheDocument()
      expect(screen.getByLabelText(/Type/i)).toBeInTheDocument()
      expect(screen.getByRole('button', { name: /Create/i })).toBeInTheDocument()
    })

    it('shows all credential type options', () => {
      renderWithTheme(<CredentialForm onSubmit={mockOnSubmit} onCancel={mockOnCancel} />)

      const typeSelect = screen.getByLabelText(/Type/i)
      fireEvent.click(typeSelect)

      expect(screen.getByText('API Key')).toBeInTheDocument()
      expect(screen.getByText('OAuth2')).toBeInTheDocument()
      expect(screen.getByText('Basic Auth')).toBeInTheDocument()
      expect(screen.getByText('Bearer Token')).toBeInTheDocument()
    })

    it('shows API key fields when type is api_key', () => {
      renderWithTheme(<CredentialForm onSubmit={mockOnSubmit} onCancel={mockOnCancel} />)

      const typeSelect = screen.getByLabelText(/Type/i)
      fireEvent.change(typeSelect, { target: { value: 'api_key' } })

      expect(screen.getByLabelText(/API Key/i)).toBeInTheDocument()
    })

    it('shows OAuth2 fields when type is oauth2', () => {
      renderWithTheme(<CredentialForm onSubmit={mockOnSubmit} onCancel={mockOnCancel} />)

      const typeSelect = screen.getByLabelText(/Type/i)
      fireEvent.change(typeSelect, { target: { value: 'oauth2' } })

      expect(screen.getByLabelText(/Client ID/i)).toBeInTheDocument()
      expect(screen.getByLabelText(/Client Secret/i)).toBeInTheDocument()
      expect(screen.getByLabelText(/Auth URL/i)).toBeInTheDocument()
      expect(screen.getByLabelText(/Token URL/i)).toBeInTheDocument()
    })

    it('shows basic auth fields when type is basic_auth', () => {
      renderWithTheme(<CredentialForm onSubmit={mockOnSubmit} onCancel={mockOnCancel} />)

      const typeSelect = screen.getByLabelText(/Type/i)
      fireEvent.change(typeSelect, { target: { value: 'basic_auth' } })

      expect(screen.getByLabelText(/Username/i)).toBeInTheDocument()
      expect(screen.getByLabelText(/Password/i)).toBeInTheDocument()
    })

    it('shows bearer token field when type is bearer_token', () => {
      renderWithTheme(<CredentialForm onSubmit={mockOnSubmit} onCancel={mockOnCancel} />)

      const typeSelect = screen.getByLabelText(/Type/i)
      fireEvent.change(typeSelect, { target: { value: 'bearer_token' } })

      expect(screen.getByLabelText(/Token/i)).toBeInTheDocument()
    })

    it('validates required name field', async () => {
      renderWithTheme(<CredentialForm onSubmit={mockOnSubmit} onCancel={mockOnCancel} />)

      const submitButton = screen.getByRole('button', { name: /Create/i })
      fireEvent.click(submitButton)

      await waitFor(() => {
        expect(screen.getByText(/Name is required/i)).toBeInTheDocument()
      })
      expect(mockOnSubmit).not.toHaveBeenCalled()
    })

    it('validates required credential value fields', async () => {
      renderWithTheme(<CredentialForm onSubmit={mockOnSubmit} onCancel={mockOnCancel} />)

      const nameInput = screen.getByLabelText(/Name/i)
      fireEvent.change(nameInput, { target: { value: 'Test' } })

      const typeSelect = screen.getByLabelText(/Type/i)
      fireEvent.change(typeSelect, { target: { value: 'api_key' } })

      const submitButton = screen.getByRole('button', { name: /Create/i })
      fireEvent.click(submitButton)

      await waitFor(() => {
        expect(screen.getByText(/API Key is required/i)).toBeInTheDocument()
      })
      expect(mockOnSubmit).not.toHaveBeenCalled()
    })

    it('submits valid API key credential', async () => {
      renderWithTheme(<CredentialForm onSubmit={mockOnSubmit} onCancel={mockOnCancel} />)

      fireEvent.change(screen.getByLabelText(/Name/i), { target: { value: 'My API Key' } })
      fireEvent.change(screen.getByLabelText(/Description/i), {
        target: { value: 'Test description' },
      })
      fireEvent.change(screen.getByLabelText(/Type/i), { target: { value: 'api_key' } })
      fireEvent.change(screen.getByLabelText(/API Key/i), { target: { value: 'sk-test-123' } })

      const submitButton = screen.getByRole('button', { name: /Create/i })
      fireEvent.click(submitButton)

      await waitFor(() => {
        expect(mockOnSubmit).toHaveBeenCalledWith({
          name: 'My API Key',
          description: 'Test description',
          type: 'api_key',
          value: { apiKey: 'sk-test-123' },
        })
      })
    })

    it('submits valid OAuth2 credential', async () => {
      renderWithTheme(<CredentialForm onSubmit={mockOnSubmit} onCancel={mockOnCancel} />)

      fireEvent.change(screen.getByLabelText(/Name/i), { target: { value: 'OAuth App' } })
      fireEvent.change(screen.getByLabelText(/Type/i), { target: { value: 'oauth2' } })
      fireEvent.change(screen.getByLabelText(/Client ID/i), { target: { value: 'client-123' } })
      fireEvent.change(screen.getByLabelText(/Client Secret/i), {
        target: { value: 'secret-456' },
      })
      fireEvent.change(screen.getByLabelText(/Auth URL/i), {
        target: { value: 'https://auth.example.com' },
      })
      fireEvent.change(screen.getByLabelText(/Token URL/i), {
        target: { value: 'https://token.example.com' },
      })

      const submitButton = screen.getByRole('button', { name: /Create/i })
      fireEvent.click(submitButton)

      await waitFor(() => {
        expect(mockOnSubmit).toHaveBeenCalledWith({
          name: 'OAuth App',
          type: 'oauth2',
          value: {
            clientId: 'client-123',
            clientSecret: 'secret-456',
            authUrl: 'https://auth.example.com',
            tokenUrl: 'https://token.example.com',
          },
        })
      })
    })

    it('includes optional expiration date when provided', async () => {
      renderWithTheme(<CredentialForm onSubmit={mockOnSubmit} onCancel={mockOnCancel} />)

      fireEvent.change(screen.getByLabelText(/Name/i), { target: { value: 'Test' } })
      fireEvent.change(screen.getByLabelText(/Type/i), { target: { value: 'api_key' } })
      fireEvent.change(screen.getByLabelText(/API Key/i), { target: { value: 'key-123' } })
      fireEvent.change(screen.getByLabelText(/Expiration Date/i), {
        target: { value: '2025-12-31' },
      })

      const submitButton = screen.getByRole('button', { name: /Create/i })
      fireEvent.click(submitButton)

      await waitFor(() => {
        expect(mockOnSubmit).toHaveBeenCalledWith(
          expect.objectContaining({
            expiresAt: expect.stringContaining('2025-12-31'),
          })
        )
      })
    })

    it('calls onCancel when cancel button is clicked', () => {
      renderWithTheme(<CredentialForm onSubmit={mockOnSubmit} onCancel={mockOnCancel} />)

      const cancelButton = screen.getByRole('button', { name: /Cancel/i })
      fireEvent.click(cancelButton)

      expect(mockOnCancel).toHaveBeenCalled()
      expect(mockOnSubmit).not.toHaveBeenCalled()
    })

    it('shows password strength indicator', () => {
      renderWithTheme(<CredentialForm onSubmit={mockOnSubmit} onCancel={mockOnCancel} />)

      const typeSelect = screen.getByLabelText(/Type/i)
      fireEvent.change(typeSelect, { target: { value: 'basic_auth' } })

      const passwordInput = screen.getByLabelText(/Password/i)
      fireEvent.change(passwordInput, { target: { value: 'weak' } })

      expect(screen.getByText(/Weak/i)).toBeInTheDocument()
    })

    it('updates password strength indicator', () => {
      renderWithTheme(<CredentialForm onSubmit={mockOnSubmit} onCancel={mockOnCancel} />)

      const typeSelect = screen.getByLabelText(/Type/i)
      fireEvent.change(typeSelect, { target: { value: 'basic_auth' } })

      const passwordInput = screen.getByLabelText(/Password/i)
      fireEvent.change(passwordInput, { target: { value: 'Str0ng!Pass123' } })

      expect(screen.getByText(/Strong/i)).toBeInTheDocument()
    })
  })

  describe('Edit mode', () => {
    it('renders edit form with existing credential data', () => {
      renderWithTheme(
        <CredentialForm
          credential={mockExistingCredential}
          onSubmit={mockOnSubmit}
          onCancel={mockOnCancel}
        />
      )

      expect(screen.getByText('Edit Credential')).toBeInTheDocument()
      expect(screen.getByDisplayValue('Test Credential')).toBeInTheDocument()
      expect(screen.getByDisplayValue('Test description')).toBeInTheDocument()
      expect(screen.getByRole('button', { name: /Save/i })).toBeInTheDocument()
    })

    it('disables type selection in edit mode', () => {
      renderWithTheme(
        <CredentialForm
          credential={mockExistingCredential}
          onSubmit={mockOnSubmit}
          onCancel={mockOnCancel}
        />
      )

      const typeSelect = screen.getByLabelText(/Type/i)
      expect(typeSelect).toBeDisabled()
    })

    it('does not show credential value fields in edit mode', () => {
      renderWithTheme(
        <CredentialForm
          credential={mockExistingCredential}
          onSubmit={mockOnSubmit}
          onCancel={mockOnCancel}
        />
      )

      // Should not show API key input in edit mode
      expect(screen.queryByLabelText(/API Key/i)).not.toBeInTheDocument()
    })

    it('shows rotate credential link in edit mode', () => {
      renderWithTheme(
        <CredentialForm
          credential={mockExistingCredential}
          onSubmit={mockOnSubmit}
          onCancel={mockOnCancel}
        />
      )

      expect(screen.getByText(/Rotate credential/i)).toBeInTheDocument()
    })

    it('submits updated metadata only', async () => {
      renderWithTheme(
        <CredentialForm
          credential={mockExistingCredential}
          onSubmit={mockOnSubmit}
          onCancel={mockOnCancel}
        />
      )

      fireEvent.change(screen.getByLabelText(/Name/i), { target: { value: 'Updated Name' } })
      fireEvent.change(screen.getByLabelText(/Description/i), {
        target: { value: 'Updated description' },
      })

      const submitButton = screen.getByRole('button', { name: /Save/i })
      fireEvent.click(submitButton)

      await waitFor(() => {
        expect(mockOnSubmit).toHaveBeenCalledWith({
          name: 'Updated Name',
          description: 'Updated description',
        })
      })
    })

    it('allows updating expiration date', async () => {
      renderWithTheme(
        <CredentialForm
          credential={mockExistingCredential}
          onSubmit={mockOnSubmit}
          onCancel={mockOnCancel}
        />
      )

      fireEvent.change(screen.getByLabelText(/Expiration Date/i), {
        target: { value: '2026-12-31' },
      })

      const submitButton = screen.getByRole('button', { name: /Save/i })
      fireEvent.click(submitButton)

      await waitFor(() => {
        expect(mockOnSubmit).toHaveBeenCalledWith(
          expect.objectContaining({
            expiresAt: expect.stringContaining('2026-12-31'),
          })
        )
      })
    })
  })

  describe('Loading state', () => {
    it('disables form during submission', () => {
      renderWithTheme(<CredentialForm onSubmit={mockOnSubmit} onCancel={mockOnCancel} loading={true} />)

      expect(screen.getByLabelText(/Name/i)).toBeDisabled()
      expect(screen.getByRole('button', { name: /Creating/i })).toBeDisabled()
    })

    it('shows loading text on submit button', () => {
      renderWithTheme(<CredentialForm onSubmit={mockOnSubmit} onCancel={mockOnCancel} loading={true} />)

      expect(screen.getByRole('button', { name: /Creating/i })).toBeInTheDocument()
    })
  })

  describe('Error handling', () => {
    it('displays error message when provided', () => {
      renderWithTheme(
        <CredentialForm
          onSubmit={mockOnSubmit}
          onCancel={mockOnCancel}
          error="Failed to create credential"
        />
      )

      expect(screen.getByText('Failed to create credential')).toBeInTheDocument()
    })

    it('dismisses error message', () => {
      const { rerender } = renderWithTheme(
        <CredentialForm
          onSubmit={mockOnSubmit}
          onCancel={mockOnCancel}
          error="Failed to create credential"
        />
      )

      expect(screen.getByText('Failed to create credential')).toBeInTheDocument()

      rerender(
        <ThemeProvider>
          <CredentialForm onSubmit={mockOnSubmit} onCancel={mockOnCancel} error={null} />
        </ThemeProvider>
      )

      expect(screen.queryByText('Failed to create credential')).not.toBeInTheDocument()
    })
  })
})
