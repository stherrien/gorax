import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, fireEvent } from '@testing-library/react'
import { CredentialList } from './CredentialList'
import { ThemeProvider } from '../../contexts/ThemeContext'
import type { Credential } from '../../api/credentials'

const renderWithTheme = (ui: React.ReactElement) => {
  return render(<ThemeProvider>{ui}</ThemeProvider>)
}

describe('CredentialList', () => {
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
      description: 'OAuth integration',
      expiresAt: '2026-12-31T23:59:59Z', // Future date
      createdAt: '2024-01-15T11:00:00Z',
      updatedAt: '2024-01-15T11:00:00Z',
    },
    {
      id: 'cred-3',
      tenantId: 'tenant-1',
      name: 'Expired Token',
      type: 'bearer_token',
      expiresAt: '2024-01-01T00:00:00Z', // Past date
      createdAt: '2024-01-15T12:00:00Z',
      updatedAt: '2024-01-15T12:00:00Z',
    },
  ]

  const defaultProps = {
    credentials: mockCredentials,
    loading: false,
    onSelect: vi.fn(),
    onEdit: vi.fn(),
    onDelete: vi.fn(),
    onTest: vi.fn(),
  }

  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('renders list of credentials', () => {
    renderWithTheme(<CredentialList {...defaultProps} />)

    expect(screen.getByText('Production API Key')).toBeInTheDocument()
    expect(screen.getByText('OAuth App')).toBeInTheDocument()
    expect(screen.getByText('Expired Token')).toBeInTheDocument()
  })

  it('displays credential types', () => {
    renderWithTheme(<CredentialList {...defaultProps} />)

    expect(screen.getByText('API Key')).toBeInTheDocument()
    expect(screen.getByText('OAuth2')).toBeInTheDocument()
    expect(screen.getByText('Bearer Token')).toBeInTheDocument()
  })

  it('displays credential descriptions', () => {
    renderWithTheme(<CredentialList {...defaultProps} />)

    expect(screen.getByText('Main API key')).toBeInTheDocument()
    expect(screen.getByText('OAuth integration')).toBeInTheDocument()
  })

  it('shows expiration warning for expiring credentials', () => {
    renderWithTheme(<CredentialList {...defaultProps} />)

    // Should show expiration date for cred-2 (expires in future)
    expect(screen.getByText(/Expires:/)).toBeInTheDocument()
  })

  it('shows expired badge for expired credentials', () => {
    renderWithTheme(<CredentialList {...defaultProps} />)

    // Should show expired badge for cred-3
    expect(screen.getByText('Expired')).toBeInTheDocument()
  })

  it('calls onSelect when credential is clicked', () => {
    renderWithTheme(<CredentialList {...defaultProps} />)

    const firstCredential = screen.getByText('Production API Key')
    fireEvent.click(firstCredential)

    expect(defaultProps.onSelect).toHaveBeenCalledWith('cred-1')
  })

  it('calls onEdit when edit button is clicked', () => {
    renderWithTheme(<CredentialList {...defaultProps} />)

    const editButtons = screen.getAllByText('Edit')
    fireEvent.click(editButtons[0])

    expect(defaultProps.onEdit).toHaveBeenCalledWith('cred-1')
  })

  it('calls onDelete when delete button is clicked', () => {
    renderWithTheme(<CredentialList {...defaultProps} />)

    const deleteButtons = screen.getAllByText('Delete')
    fireEvent.click(deleteButtons[0])

    expect(defaultProps.onDelete).toHaveBeenCalledWith('cred-1')
  })

  it('calls onTest when test button is clicked', () => {
    renderWithTheme(<CredentialList {...defaultProps} />)

    const testButtons = screen.getAllByText('Test')
    fireEvent.click(testButtons[0])

    expect(defaultProps.onTest).toHaveBeenCalledWith('cred-1')
  })

  it('displays loading state', () => {
    // Loading state only shows when there are no credentials
    renderWithTheme(<CredentialList {...defaultProps} credentials={[]} loading={true} />)

    expect(screen.getByText(/Loading/i)).toBeInTheDocument()
  })

  it('displays empty state when no credentials', () => {
    renderWithTheme(<CredentialList {...defaultProps} credentials={[]} />)

    expect(screen.getByText(/No credentials found/i)).toBeInTheDocument()
  })

  it('displays empty state message', () => {
    renderWithTheme(<CredentialList {...defaultProps} credentials={[]} />)

    expect(screen.getByText(/Create your first credential/i)).toBeInTheDocument()
  })

  it('filters credentials by search term', () => {
    renderWithTheme(<CredentialList {...defaultProps} searchTerm="OAuth" />)

    expect(screen.getByText('OAuth App')).toBeInTheDocument()
    expect(screen.queryByText('Production API Key')).not.toBeInTheDocument()
  })

  it('filters credentials by type', () => {
    renderWithTheme(<CredentialList {...defaultProps} filterType="oauth2" />)

    expect(screen.getByText('OAuth App')).toBeInTheDocument()
    expect(screen.queryByText('Production API Key')).not.toBeInTheDocument()
  })

  it('applies both search and type filters', () => {
    renderWithTheme(<CredentialList {...defaultProps} searchTerm="API" filterType="api_key" />)

    expect(screen.getByText('Production API Key')).toBeInTheDocument()
    expect(screen.queryByText('OAuth App')).not.toBeInTheDocument()
  })

  it('shows credential count', () => {
    renderWithTheme(<CredentialList {...defaultProps} />)

    expect(screen.getByText(/3 credentials/i)).toBeInTheDocument()
  })

  it('shows filtered count when filters applied', () => {
    renderWithTheme(<CredentialList {...defaultProps} filterType="api_key" />)

    expect(screen.getByText(/1 credential/i)).toBeInTheDocument()
  })

  it('highlights selected credential', () => {
    renderWithTheme(<CredentialList {...defaultProps} selectedId="cred-2" />)

    // Find the container div (not just the closest div which might be a child)
    const oauthText = screen.getByText('OAuth App')
    const selectedItem = oauthText.closest('.border')
    expect(selectedItem?.className).toContain('bg-primary-600/20')
  })

  it('sorts credentials by name', () => {
    renderWithTheme(<CredentialList {...defaultProps} sortBy="name" />)

    const names = screen.getAllByTestId('credential-name')
    // When sorted by name: Expired Token, OAuth App, Production API Key
    expect(names[0].textContent).toBe('Expired Token')
    expect(names[1].textContent).toBe('OAuth App')
    expect(names[2].textContent).toBe('Production API Key')
  })

  it('sorts credentials by creation date', () => {
    renderWithTheme(<CredentialList {...defaultProps} sortBy="created" />)

    const names = screen.getAllByTestId('credential-name')
    // When sorted by creation date (ascending): Production, OAuth, Expired
    expect(names[0].textContent).toBe('Production API Key')
    expect(names[1].textContent).toBe('OAuth App')
    expect(names[2].textContent).toBe('Expired Token')
  })

  it('sorts credentials by type', () => {
    renderWithTheme(<CredentialList {...defaultProps} sortBy="type" />)

    const types = screen.getAllByTestId('credential-type')
    // When sorted by type: api_key, basic_auth, bearer_token, oauth2
    expect(types[0].textContent).toBe('API Key')
    expect(types[1].textContent).toBe('Bearer Token')
    expect(types[2].textContent).toBe('OAuth2')
  })

  it('shows action menu for each credential', () => {
    renderWithTheme(<CredentialList {...defaultProps} />)

    const editButtons = screen.getAllByText('Edit')
    expect(editButtons).toHaveLength(3)

    const deleteButtons = screen.getAllByText('Delete')
    expect(deleteButtons).toHaveLength(3)

    const testButtons = screen.getAllByText('Test')
    expect(testButtons).toHaveLength(3)
  })

  it('disables test button during loading', () => {
    renderWithTheme(<CredentialList {...defaultProps} loading={true} />)

    // Get test buttons by role since there are credentials present
    const buttons = screen.getAllByRole('button', { name: 'Test' })
    buttons.forEach((button) => {
      expect(button).toBeDisabled()
    })
  })

  it('formats creation date', () => {
    renderWithTheme(<CredentialList {...defaultProps} />)

    // Should format dates in readable format - there will be multiple "Created:" texts
    const createdTexts = screen.getAllByText(/Created:/)
    expect(createdTexts.length).toBeGreaterThan(0)
  })
})
