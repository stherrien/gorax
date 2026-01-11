import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import { MemoryRouter, Route, Routes } from 'react-router-dom'
import { OAuthCallback } from './OAuthCallback'
import * as oauthAPI from '../../api/oauth'

vi.mock('../../api/oauth', () => ({
  handleCallback: vi.fn(),
}))

const renderWithRouter = (provider: string, searchParams: string) => {
  const initialEntries = [`/oauth/callback/${provider}?${searchParams}`]
  return render(
    <MemoryRouter initialEntries={initialEntries}>
      <Routes>
        <Route path="/oauth/callback/:provider" element={<OAuthCallback />} />
      </Routes>
    </MemoryRouter>
  )
}

describe('OAuthCallback', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('should show processing state initially', () => {
    vi.mocked(oauthAPI.handleCallback).mockImplementation(
      () => new Promise(() => {}) // Never resolves
    )

    renderWithRouter('github', 'code=abc123&state=xyz789')

    expect(screen.getByText(/connecting/i)).toBeInTheDocument()
  })

  it('should show success state when callback succeeds', async () => {
    vi.mocked(oauthAPI.handleCallback).mockResolvedValue({
      success: true,
      provider: 'github',
      connection: {
        id: 'conn-1',
        user_id: 'user-1',
        tenant_id: 'tenant-1',
        provider_key: 'github',
        scopes: [],
        status: 'active',
        created_at: '2024-01-01T00:00:00Z',
        updated_at: '2024-01-01T00:00:00Z',
      },
    })

    renderWithRouter('github', 'code=abc123&state=xyz789')

    await waitFor(() => {
      expect(screen.getByText('Success!')).toBeInTheDocument()
      expect(screen.getByText(/successfully connected/i)).toBeInTheDocument()
    })
  })

  it('should show error state when OAuth error is present', async () => {
    renderWithRouter('github', 'error=access_denied&error_description=User denied access')

    await waitFor(() => {
      expect(screen.getByText(/connection failed/i)).toBeInTheDocument()
      expect(screen.getByText(/user denied access/i)).toBeInTheDocument()
    })
  })

  it('should show error state when code is missing', async () => {
    renderWithRouter('github', 'state=xyz789')

    await waitFor(() => {
      expect(screen.getByText(/connection failed/i)).toBeInTheDocument()
      expect(screen.getByText(/missing required oauth parameters/i)).toBeInTheDocument()
    })
  })

  it('should show error state when state is missing', async () => {
    renderWithRouter('github', 'code=abc123')

    await waitFor(() => {
      expect(screen.getByText(/connection failed/i)).toBeInTheDocument()
      expect(screen.getByText(/missing required oauth parameters/i)).toBeInTheDocument()
    })
  })

  it('should show error state when callback API fails', async () => {
    vi.mocked(oauthAPI.handleCallback).mockRejectedValue(new Error('API error'))

    renderWithRouter('github', 'code=abc123&state=xyz789')

    await waitFor(() => {
      expect(screen.getByText(/connection failed/i)).toBeInTheDocument()
      expect(screen.getByText(/api error/i)).toBeInTheDocument()
    })
  })

  it('should handle OAuth error description', async () => {
    renderWithRouter('github', 'error=server_error&error_description=Internal server error')

    await waitFor(() => {
      expect(screen.getByText(/internal server error/i)).toBeInTheDocument()
    })
  })

  it('should handle OAuth error without description', async () => {
    renderWithRouter('github', 'error=access_denied')

    await waitFor(() => {
      expect(screen.getByText(/access_denied/i)).toBeInTheDocument()
    })
  })
})
