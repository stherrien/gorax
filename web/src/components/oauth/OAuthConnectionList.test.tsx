import { describe, it, expect } from 'vitest'
import { render, screen } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { OAuthConnectionList } from './OAuthConnectionList'
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

describe('OAuthConnectionList', () => {
  const mockConnections: OAuthConnection[] = [
    {
      id: 'conn-1',
      user_id: 'user-1',
      tenant_id: 'tenant-1',
      provider_key: 'github',
      provider_username: 'testuser',
      scopes: ['read:user'],
      status: 'active',
      created_at: '2024-01-01T00:00:00Z',
      updated_at: '2024-01-01T00:00:00Z',
    },
    {
      id: 'conn-2',
      user_id: 'user-1',
      tenant_id: 'tenant-1',
      provider_key: 'google',
      provider_email: 'test@example.com',
      scopes: ['email', 'profile'],
      status: 'active',
      created_at: '2024-01-02T00:00:00Z',
      updated_at: '2024-01-02T00:00:00Z',
    },
  ]

  it('should render loading state', () => {
    render(<OAuthConnectionList connections={[]} isLoading={true} />, { wrapper })

    // Should show skeleton loaders
    const skeletons = screen.getAllByRole('generic')
    expect(skeletons.length).toBeGreaterThan(0)
  })

  it('should render empty state when no connections', () => {
    render(<OAuthConnectionList connections={[]} isLoading={false} />, { wrapper })

    expect(screen.getByText(/no oauth connections/i)).toBeInTheDocument()
  })

  it('should render list of connections', () => {
    render(<OAuthConnectionList connections={mockConnections} isLoading={false} />, { wrapper })

    expect(screen.getByText('GitHub')).toBeInTheDocument()
    expect(screen.getByText('Google')).toBeInTheDocument()
  })

  it('should group connections by provider', () => {
    const connections: OAuthConnection[] = [
      mockConnections[0],
      {
        ...mockConnections[0],
        id: 'conn-3',
        provider_username: 'testuser2',
      },
    ]

    render(<OAuthConnectionList connections={connections} isLoading={false} />, { wrapper })

    // Should show count in header
    expect(screen.getByText(/github \(2\)/i)).toBeInTheDocument()
  })

  it('should handle empty connections array', () => {
    render(<OAuthConnectionList connections={[]} />, { wrapper })

    expect(screen.getByText(/no oauth connections/i)).toBeInTheDocument()
  })
})
