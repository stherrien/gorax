import { describe, it, expect, vi } from 'vitest'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { OAuthProviderCard } from './OAuthProviderCard'
import type { OAuthProvider } from '../../types/oauth'

describe('OAuthProviderCard', () => {
  const mockProvider: OAuthProvider = {
    id: 'prov-1',
    provider_key: 'github',
    name: 'GitHub',
    description: 'Connect your GitHub account',
    auth_url: 'https://github.com/login/oauth/authorize',
    token_url: 'https://github.com/login/oauth/access_token',
    user_info_url: 'https://api.github.com/user',
    default_scopes: ['read:user', 'repo'],
    status: 'active',
    created_at: '2024-01-01T00:00:00Z',
    updated_at: '2024-01-01T00:00:00Z',
  }

  it('should render provider information', () => {
    render(
      <OAuthProviderCard provider={mockProvider} isConnected={false} onConnect={vi.fn()} />
    )

    expect(screen.getByText('GitHub')).toBeInTheDocument()
    expect(screen.getByText('Connect your GitHub account')).toBeInTheDocument()
  })

  it('should display default scopes', () => {
    render(
      <OAuthProviderCard provider={mockProvider} isConnected={false} onConnect={vi.fn()} />
    )

    expect(screen.getByText('read:user')).toBeInTheDocument()
    expect(screen.getByText('repo')).toBeInTheDocument()
  })

  it('should show Connect button when not connected', () => {
    render(
      <OAuthProviderCard provider={mockProvider} isConnected={false} onConnect={vi.fn()} />
    )

    expect(screen.getByRole('button', { name: /connect/i })).toBeInTheDocument()
  })

  it('should show Connected status when connected', () => {
    render(
      <OAuthProviderCard provider={mockProvider} isConnected={true} onConnect={vi.fn()} />
    )

    expect(screen.getByText(/connected/i)).toBeInTheDocument()
    expect(screen.queryByRole('button', { name: /connect/i })).not.toBeInTheDocument()
  })

  it('should call onConnect when Connect button is clicked', async () => {
    const onConnect = vi.fn()
    const user = userEvent.setup()

    render(<OAuthProviderCard provider={mockProvider} isConnected={false} onConnect={onConnect} />)

    await user.click(screen.getByRole('button', { name: /connect/i }))

    expect(onConnect).toHaveBeenCalledTimes(1)
  })

  it('should use provider branding for known providers', () => {
    render(
      <OAuthProviderCard provider={mockProvider} isConnected={false} onConnect={vi.fn()} />
    )

    // GitHub icon should be rendered
    expect(screen.getByText('🐙')).toBeInTheDocument()
  })
})
