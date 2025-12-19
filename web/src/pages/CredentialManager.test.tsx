import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import { CredentialManager } from './CredentialManager'
import { useCredentialStore } from '../stores/credentialStore'
import type { Credential } from '../api/credentials'

vi.mock('../stores/credentialStore')

describe('CredentialManager', () => {
  const mockCredentials: Credential[] = [
    {
      id: 'cred-1',
      tenantId: 'tenant-1',
      name: 'Production API Key',
      type: 'api_key',
      description: 'Main API key',
      createdAt: '2024-01-15T10:00:00Z',
      updatedAt: '2024-01-15T10:00:00Z',
    },
    {
      id: 'cred-2',
      tenantId: 'tenant-1',
      name: 'OAuth App',
      type: 'oauth2',
      createdAt: '2024-01-15T11:00:00Z',
      updatedAt: '2024-01-15T11:00:00Z',
    },
  ]

  const mockStore = {
    credentials: mockCredentials,
    loading: false,
    error: null,
    selectedCredential: null,
    fetchCredentials: vi.fn(),
    createCredential: vi.fn(),
    updateCredential: vi.fn(),
    deleteCredential: vi.fn(),
    testCredential: vi.fn(),
    selectCredential: vi.fn(),
    clearError: vi.fn(),
  }

  beforeEach(() => {
    vi.clearAllMocks()
    ;(useCredentialStore as any).mockReturnValue(mockStore)
  })

  it('renders credential manager page', () => {
    render(<CredentialManager />)

    expect(screen.getByText('Credentials')).toBeInTheDocument()
    expect(screen.getByText('Production API Key')).toBeInTheDocument()
    expect(screen.getByText('OAuth App')).toBeInTheDocument()
  })

  it('fetches credentials on mount', () => {
    render(<CredentialManager />)

    expect(mockStore.fetchCredentials).toHaveBeenCalled()
  })

  it('shows create credential button', () => {
    render(<CredentialManager />)

    expect(screen.getByRole('button', { name: /Create Credential/i })).toBeInTheDocument()
  })

  it('opens create form when create button is clicked', () => {
    render(<CredentialManager />)

    const createButton = screen.getByRole('button', { name: /Create Credential/i })
    fireEvent.click(createButton)

    expect(screen.getByText('Create Credential')).toBeInTheDocument()
  })

  it('closes create form when cancel is clicked', () => {
    render(<CredentialManager />)

    const createButton = screen.getByRole('button', { name: /Create Credential/i })
    fireEvent.click(createButton)

    // Form heading should be visible
    expect(screen.getByRole('heading', { name: 'Create Credential' })).toBeInTheDocument()

    const cancelButton = screen.getByRole('button', { name: /Cancel/i })
    fireEvent.click(cancelButton)

    // Form heading should no longer be visible after cancel
    expect(screen.queryByRole('heading', { name: 'Create Credential' })).not.toBeInTheDocument()
    // But the create button should still be there
    expect(screen.getByRole('button', { name: /Create Credential/i })).toBeInTheDocument()
  })

  it('creates credential when form is submitted', async () => {
    mockStore.createCredential.mockResolvedValueOnce(undefined)

    render(<CredentialManager />)

    const createButton = screen.getByRole('button', { name: /Create Credential/i })
    fireEvent.click(createButton)

    // Fill form
    fireEvent.change(screen.getByLabelText(/Name/i), { target: { value: 'New Credential' } })
    fireEvent.change(screen.getByLabelText(/Type/i), { target: { value: 'api_key' } })
    fireEvent.change(screen.getByLabelText(/API Key/i), { target: { value: 'key-123' } })

    const submitButton = screen.getByRole('button', { name: /^Create$/i })
    fireEvent.click(submitButton)

    await waitFor(() => {
      expect(mockStore.createCredential).toHaveBeenCalledWith({
        name: 'New Credential',
        type: 'api_key',
        value: { apiKey: 'key-123' },
      })
    })
  })

  it('shows edit form when edit button is clicked', () => {
    render(<CredentialManager />)

    const editButtons = screen.getAllByText('Edit')
    fireEvent.click(editButtons[0])

    expect(screen.getByText('Edit Credential')).toBeInTheDocument()
    expect(screen.getByDisplayValue('Production API Key')).toBeInTheDocument()
  })

  it('updates credential when edit form is submitted', async () => {
    mockStore.updateCredential.mockResolvedValueOnce(undefined)

    render(<CredentialManager />)

    const editButtons = screen.getAllByText('Edit')
    fireEvent.click(editButtons[0])

    fireEvent.change(screen.getByLabelText(/Name/i), { target: { value: 'Updated Name' } })

    const submitButton = screen.getByRole('button', { name: /Save/i })
    fireEvent.click(submitButton)

    await waitFor(() => {
      expect(mockStore.updateCredential).toHaveBeenCalledWith('cred-1', {
        name: 'Updated Name',
        description: 'Main API key',
      })
    })
  })

  it('shows delete confirmation when delete button is clicked', () => {
    render(<CredentialManager />)

    const deleteButtons = screen.getAllByText('Delete')
    fireEvent.click(deleteButtons[0])

    expect(screen.getByText(/Are you sure you want to delete/i)).toBeInTheDocument()
    expect(screen.getByText('Production API Key')).toBeInTheDocument()
  })

  it('deletes credential when confirmed', async () => {
    mockStore.deleteCredential.mockResolvedValueOnce(undefined)

    render(<CredentialManager />)

    const deleteButtons = screen.getAllByText('Delete')
    fireEvent.click(deleteButtons[0])

    const confirmButton = screen.getByRole('button', { name: /Confirm/i })
    fireEvent.click(confirmButton)

    await waitFor(() => {
      expect(mockStore.deleteCredential).toHaveBeenCalledWith('cred-1')
    })
  })

  it('cancels delete when cancel is clicked', () => {
    render(<CredentialManager />)

    const deleteButtons = screen.getAllByText('Delete')
    fireEvent.click(deleteButtons[0])

    const cancelButton = screen.getAllByRole('button', { name: /Cancel/i })[0]
    fireEvent.click(cancelButton)

    expect(screen.queryByText(/Are you sure you want to delete/i)).not.toBeInTheDocument()
  })

  it('tests credential when test button is clicked', async () => {
    mockStore.testCredential.mockResolvedValueOnce({
      success: true,
      message: 'Connection successful',
      testedAt: '2024-01-15T10:00:00Z',
    })

    render(<CredentialManager />)

    const testButtons = screen.getAllByText('Test')
    fireEvent.click(testButtons[0])

    await waitFor(() => {
      expect(mockStore.testCredential).toHaveBeenCalledWith('cred-1')
    })

    expect(screen.getByText(/Connection successful/i)).toBeInTheDocument()
  })

  it('shows error message when test fails', async () => {
    mockStore.testCredential.mockResolvedValueOnce({
      success: false,
      message: 'Authentication failed',
      testedAt: '2024-01-15T10:00:00Z',
    })

    render(<CredentialManager />)

    const testButtons = screen.getAllByText('Test')
    fireEvent.click(testButtons[0])

    await waitFor(() => {
      expect(screen.getByText(/Authentication failed/i)).toBeInTheDocument()
    })
  })

  it('shows search input', () => {
    render(<CredentialManager />)

    expect(screen.getByPlaceholderText(/Search credentials/i)).toBeInTheDocument()
  })

  it('filters credentials by search term', () => {
    render(<CredentialManager />)

    const searchInput = screen.getByPlaceholderText(/Search credentials/i)
    fireEvent.change(searchInput, { target: { value: 'OAuth' } })

    expect(screen.getByText('OAuth App')).toBeInTheDocument()
    expect(screen.queryByText('Production API Key')).not.toBeInTheDocument()
  })

  it('shows type filter', () => {
    render(<CredentialManager />)

    expect(screen.getByLabelText(/Filter by type/i)).toBeInTheDocument()
  })

  it('filters credentials by type', () => {
    render(<CredentialManager />)

    const typeFilter = screen.getByLabelText(/Filter by type/i)
    fireEvent.change(typeFilter, { target: { value: 'oauth2' } })

    expect(screen.getByText('OAuth App')).toBeInTheDocument()
    expect(screen.queryByText('Production API Key')).not.toBeInTheDocument()
  })

  it('shows sort options', () => {
    render(<CredentialManager />)

    expect(screen.getByLabelText(/Sort by/i)).toBeInTheDocument()
  })

  it('sorts credentials', () => {
    render(<CredentialManager />)

    const sortSelect = screen.getByLabelText(/Sort by/i)
    fireEvent.change(sortSelect, { target: { value: 'name' } })

    const names = screen.getAllByTestId('credential-name')
    expect(names[0]).toHaveTextContent('OAuth App')
    expect(names[1]).toHaveTextContent('Production API Key')
  })

  it('displays loading state', () => {
    ;(useCredentialStore as any).mockReturnValue({
      ...mockStore,
      loading: true,
      credentials: [],
    })

    render(<CredentialManager />)

    expect(screen.getByText(/Loading/i)).toBeInTheDocument()
  })

  it('displays error state', () => {
    ;(useCredentialStore as any).mockReturnValue({
      ...mockStore,
      error: 'Failed to fetch credentials',
    })

    render(<CredentialManager />)

    expect(screen.getByText('Failed to fetch credentials')).toBeInTheDocument()
  })

  it('clears error when dismissed', () => {
    ;(useCredentialStore as any).mockReturnValue({
      ...mockStore,
      error: 'Failed to fetch credentials',
    })

    render(<CredentialManager />)

    const dismissButton = screen.getByRole('button', { name: /dismiss/i })
    fireEvent.click(dismissButton)

    expect(mockStore.clearError).toHaveBeenCalled()
  })

  it('displays empty state when no credentials', () => {
    ;(useCredentialStore as any).mockReturnValue({
      ...mockStore,
      credentials: [],
    })

    render(<CredentialManager />)

    expect(screen.getByText(/No credentials found/i)).toBeInTheDocument()
  })

  it('shows credential count', () => {
    render(<CredentialManager />)

    expect(screen.getByText(/2 credentials/i)).toBeInTheDocument()
  })

  it('refreshes credentials when refresh button is clicked', () => {
    render(<CredentialManager />)

    const refreshButton = screen.getByRole('button', { name: /Refresh/i })
    fireEvent.click(refreshButton)

    expect(mockStore.fetchCredentials).toHaveBeenCalledTimes(2) // Once on mount, once on click
  })
})
