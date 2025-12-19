import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, fireEvent } from '@testing-library/react'
import { CredentialPicker } from './CredentialPicker'
import type { Credential } from '../../api/credentials'

describe('CredentialPicker', () => {
  const mockCredentials: Credential[] = [
    {
      id: 'cred-1',
      tenantId: 'tenant-1',
      name: 'Production API Key',
      type: 'api_key',
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
    {
      id: 'cred-3',
      tenantId: 'tenant-1',
      name: 'Basic Auth Credential',
      type: 'basic_auth',
      createdAt: '2024-01-15T12:00:00Z',
      updatedAt: '2024-01-15T12:00:00Z',
    },
  ]

  const defaultProps = {
    credentials: mockCredentials,
    onSelect: vi.fn(),
  }

  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('renders credential picker button', () => {
    render(<CredentialPicker {...defaultProps} />)

    expect(screen.getByRole('button', { name: /Select Credential/i })).toBeInTheDocument()
  })

  it('shows dropdown when button is clicked', () => {
    render(<CredentialPicker {...defaultProps} />)

    const button = screen.getByRole('button', { name: /Select Credential/i })
    fireEvent.click(button)

    expect(screen.getByText('Production API Key')).toBeInTheDocument()
    expect(screen.getByText('OAuth App')).toBeInTheDocument()
    expect(screen.getByText('Basic Auth Credential')).toBeInTheDocument()
  })

  it('closes dropdown when credential is selected', () => {
    render(<CredentialPicker {...defaultProps} />)

    const button = screen.getByRole('button', { name: /Select Credential/i })
    fireEvent.click(button)

    const credential = screen.getByText('Production API Key')
    fireEvent.click(credential)

    // Dropdown should close
    expect(screen.queryByText('OAuth App')).not.toBeInTheDocument()
  })

  it('calls onSelect with credential name template', () => {
    render(<CredentialPicker {...defaultProps} />)

    const button = screen.getByRole('button')
    fireEvent.click(button)

    const credential = screen.getByText('Production API Key')
    fireEvent.click(credential)

    expect(defaultProps.onSelect).toHaveBeenCalledWith('{{credentials.Production API Key}}')
  })

  it('calls onSelect with escaped credential name', () => {
    const credentialsWithSpaces: Credential[] = [
      {
        id: 'cred-1',
        tenantId: 'tenant-1',
        name: 'My Special Key',
        type: 'api_key',
        createdAt: '2024-01-15T10:00:00Z',
        updatedAt: '2024-01-15T10:00:00Z',
      },
    ]

    render(<CredentialPicker {...defaultProps} credentials={credentialsWithSpaces} />)

    const button = screen.getByRole('button')
    fireEvent.click(button)

    const credential = screen.getByText('My Special Key')
    fireEvent.click(credential)

    expect(defaultProps.onSelect).toHaveBeenCalledWith('{{credentials.My Special Key}}')
  })

  it('displays credential type badges', () => {
    render(<CredentialPicker {...defaultProps} />)

    const button = screen.getByRole('button')
    fireEvent.click(button)

    expect(screen.getByText('API Key')).toBeInTheDocument()
    expect(screen.getByText('OAuth2')).toBeInTheDocument()
    expect(screen.getByText('Basic Auth')).toBeInTheDocument()
  })

  it('filters credentials by type when filterType is provided', () => {
    render(<CredentialPicker {...defaultProps} filterType="api_key" />)

    const button = screen.getByRole('button')
    fireEvent.click(button)

    expect(screen.getByText('Production API Key')).toBeInTheDocument()
    expect(screen.queryByText('OAuth App')).not.toBeInTheDocument()
    expect(screen.queryByText('Basic Auth Credential')).not.toBeInTheDocument()
  })

  it('shows empty state when no credentials available', () => {
    render(<CredentialPicker {...defaultProps} credentials={[]} />)

    const button = screen.getByRole('button')
    fireEvent.click(button)

    expect(screen.getByText(/No credentials available/i)).toBeInTheDocument()
  })

  it('shows empty state when filtered credentials is empty', () => {
    render(<CredentialPicker {...defaultProps} filterType="bearer_token" />)

    const button = screen.getByRole('button')
    fireEvent.click(button)

    expect(screen.getByText(/No credentials available/i)).toBeInTheDocument()
  })

  it('shows create credential link in empty state', () => {
    const onCreate = vi.fn()
    render(<CredentialPicker {...defaultProps} credentials={[]} onCreate={onCreate} />)

    const button = screen.getByRole('button')
    fireEvent.click(button)

    expect(screen.getByText(/Create a credential/i)).toBeInTheDocument()
  })

  it('calls onCreate when create link is clicked', () => {
    const onCreate = vi.fn()
    render(<CredentialPicker {...defaultProps} credentials={[]} onCreate={onCreate} />)

    const button = screen.getByRole('button')
    fireEvent.click(button)

    const createLink = screen.getByText(/Create a credential/i)
    fireEvent.click(createLink)

    expect(onCreate).toHaveBeenCalled()
  })

  it('displays custom placeholder text', () => {
    render(<CredentialPicker {...defaultProps} placeholder="Choose a credential" />)

    expect(screen.getByRole('button', { name: /Choose a credential/i })).toBeInTheDocument()
  })

  it('displays selected credential name when value is provided', () => {
    render(<CredentialPicker {...defaultProps} value="{{credentials.Production API Key}}" />)

    expect(screen.getByText('Production API Key')).toBeInTheDocument()
  })

  it('extracts credential name from template syntax', () => {
    render(<CredentialPicker {...defaultProps} value="{{credentials.OAuth App}}" />)

    expect(screen.getByText('OAuth App')).toBeInTheDocument()
  })

  it('shows placeholder when value does not match any credential', () => {
    render(<CredentialPicker {...defaultProps} value="{{credentials.NonExistent}}" />)

    expect(screen.getByRole('button', { name: /Select Credential/i })).toBeInTheDocument()
  })

  it('supports keyboard navigation', () => {
    render(<CredentialPicker {...defaultProps} />)

    const button = screen.getByRole('button')
    fireEvent.click(button)

    // Should be able to tab through options
    const firstOption = screen.getByText('Production API Key')
    expect(firstOption).toBeInTheDocument()
  })

  it('closes dropdown when clicking outside', () => {
    render(
      <div>
        <div data-testid="outside">Outside</div>
        <CredentialPicker {...defaultProps} />
      </div>
    )

    const button = screen.getByRole('button')
    fireEvent.click(button)

    expect(screen.getByText('Production API Key')).toBeInTheDocument()

    const outside = screen.getByTestId('outside')
    fireEvent.mouseDown(outside)

    expect(screen.queryByText('OAuth App')).not.toBeInTheDocument()
  })

  it('shows loading state', () => {
    render(<CredentialPicker {...defaultProps} loading={true} />)

    expect(screen.getByRole('button')).toBeDisabled()
    expect(screen.getByText(/Loading/i)).toBeInTheDocument()
  })

  it('disables picker when disabled prop is true', () => {
    render(<CredentialPicker {...defaultProps} disabled={true} />)

    const button = screen.getByRole('button')
    expect(button).toBeDisabled()
  })

  it('shows credential count in dropdown', () => {
    render(<CredentialPicker {...defaultProps} />)

    const button = screen.getByRole('button')
    fireEvent.click(button)

    expect(screen.getByText(/3 credentials/i)).toBeInTheDocument()
  })

  it('shows search input when many credentials', () => {
    const manyCredentials = Array.from({ length: 10 }, (_, i) => ({
      id: `cred-${i}`,
      tenantId: 'tenant-1',
      name: `Credential ${i}`,
      type: 'api_key' as const,
      createdAt: '2024-01-15T10:00:00Z',
      updatedAt: '2024-01-15T10:00:00Z',
    }))

    render(<CredentialPicker {...defaultProps} credentials={manyCredentials} />)

    const button = screen.getByRole('button')
    fireEvent.click(button)

    expect(screen.getByPlaceholderText(/Search/i)).toBeInTheDocument()
  })

  it('filters credentials by search term', () => {
    const manyCredentials = Array.from({ length: 10 }, (_, i) => ({
      id: `cred-${i}`,
      tenantId: 'tenant-1',
      name: i === 0 ? 'OAuth App' : `Credential ${i}`,
      type: 'api_key' as const,
      createdAt: '2024-01-15T10:00:00Z',
      updatedAt: '2024-01-15T10:00:00Z',
    }))

    render(<CredentialPicker {...defaultProps} credentials={manyCredentials} />)

    const button = screen.getByRole('button')
    fireEvent.click(button)

    const searchInput = screen.getByPlaceholderText(/Search/i)
    fireEvent.change(searchInput, { target: { value: 'OAuth' } })

    expect(screen.getByText('OAuth App')).toBeInTheDocument()
    expect(screen.queryByText('Credential 1')).not.toBeInTheDocument()
  })
})
