import { describe, it, expect } from 'vitest'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { OAuthConnectionCard } from './OAuthConnectionCard'
import type { OAuthConnection } from '../../types/oauth'

const createTestQueryClient = () =>
  new QueryClient({
    defaultOptions: {
      queries: { retry: false },
      mutations: { retry: false },
    },
  })

const wrapper = ({ children }: { children: React.ReactNode }) => (
  <QueryClientProvider client={createTestQueryClient()}>{children}</QueryClientProvider>
)

describe('OAuthConnectionCard', () => {
  const mockConnection: OAuthConnection = {
    id: 'conn-1',
    user_id: 'user-1',
    tenant_id: 'tenant-1',
    provider_key: 'github',
    provider_user_id: 'gh-123',
    provider_username: 'testuser',
    provider_email: 'test@example.com',
    scopes: ['read:user', 'repo'],
    status: 'active',
    created_at: '2024-01-01T00:00:00Z',
    updated_at: '2024-01-01T00:00:00Z',
    last_used_at: '2024-01-02T00:00:00Z',
  }

  it('should render connection information', () => {
    render(<OAuthConnectionCard connection={mockConnection} />, { wrapper })

    expect(screen.getByText('GitHub')).toBeInTheDocument()
    expect(screen.getByText('@testuser')).toBeInTheDocument()
    expect(screen.getByText('test@example.com')).toBeInTheDocument()
  })

  it('should display connection status', () => {
    render(<OAuthConnectionCard connection={mockConnection} />, { wrapper })

    expect(screen.getByText('active')).toBeInTheDocument()
  })

  it('should display scopes', () => {
    render(<OAuthConnectionCard connection={mockConnection} />, { wrapper })

    expect(screen.getByText('read:user')).toBeInTheDocument()
    expect(screen.getByText('repo')).toBeInTheDocument()
  })

  it('should show Test and Disconnect buttons for active connections', () => {
    render(<OAuthConnectionCard connection={mockConnection} />, { wrapper })

    expect(screen.getByRole('button', { name: /test/i })).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /disconnect/i })).toBeInTheDocument()
  })

  it('should show revoke confirmation when Disconnect is clicked', async () => {
    const user = userEvent.setup()
    render(<OAuthConnectionCard connection={mockConnection} />, { wrapper })

    await user.click(screen.getByRole('button', { name: /disconnect/i }))

    expect(screen.getByText(/are you sure/i)).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /yes/i })).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /no/i })).toBeInTheDocument()
  })

  it('should cancel revoke confirmation when No is clicked', async () => {
    const user = userEvent.setup()
    render(<OAuthConnectionCard connection={mockConnection} />, { wrapper })

    await user.click(screen.getByRole('button', { name: /disconnect/i }))
    await user.click(screen.getByRole('button', { name: /no/i }))

    expect(screen.queryByText(/are you sure/i)).not.toBeInTheDocument()
    expect(screen.getByRole('button', { name: /disconnect/i })).toBeInTheDocument()
  })

  it('should display expired status correctly', () => {
    const expiredConnection: OAuthConnection = {
      ...mockConnection,
      status: 'expired',
    }

    render(<OAuthConnectionCard connection={expiredConnection} />, { wrapper })

    expect(screen.getByText('expired')).toBeInTheDocument()
  })

  it('should display revoked status correctly', () => {
    const revokedConnection: OAuthConnection = {
      ...mockConnection,
      status: 'revoked',
    }

    render(<OAuthConnectionCard connection={revokedConnection} />, { wrapper })

    expect(screen.getByText('revoked')).toBeInTheDocument()
  })

  it('should display token expiry warning', () => {
    const connectionWithExpiry: OAuthConnection = {
      ...mockConnection,
      token_expiry: '2024-12-31T23:59:59Z',
    }

    render(<OAuthConnectionCard connection={connectionWithExpiry} />, { wrapper })

    expect(screen.getByText(/expires:/i)).toBeInTheDocument()
  })
})
