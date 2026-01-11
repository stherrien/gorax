import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { OAuthConnectButton } from './OAuthConnectButton'
import * as useOAuthHook from '../../hooks/useOAuth'

vi.mock('../../hooks/useOAuth', async () => {
  const actual = await vi.importActual('../../hooks/useOAuth')
  return {
    ...actual,
    useAuthorize: vi.fn(),
    openOAuthPopup: vi.fn(),
  }
})

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

describe('OAuthConnectButton', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('should render button with default text', () => {
    vi.mocked(useOAuthHook.useAuthorize).mockReturnValue({
      mutateAsync: vi.fn(),
      isPending: false,
    } as any)

    render(<OAuthConnectButton providerKey="github" />, { wrapper })

    expect(screen.getByRole('button', { name: /connect github/i })).toBeInTheDocument()
  })

  it('should render button with custom children', () => {
    vi.mocked(useOAuthHook.useAuthorize).mockReturnValue({
      mutateAsync: vi.fn(),
      isPending: false,
    } as any)

    render(<OAuthConnectButton providerKey="github">Custom Text</OAuthConnectButton>, { wrapper })

    expect(screen.getByRole('button', { name: /custom text/i })).toBeInTheDocument()
  })

  it('should show connecting state when pending', () => {
    vi.mocked(useOAuthHook.useAuthorize).mockReturnValue({
      mutateAsync: vi.fn(),
      isPending: true,
    } as any)

    render(<OAuthConnectButton providerKey="github" />, { wrapper })

    expect(screen.getByText(/connecting/i)).toBeInTheDocument()
  })

  it('should disable button when connecting', () => {
    vi.mocked(useOAuthHook.useAuthorize).mockReturnValue({
      mutateAsync: vi.fn(),
      isPending: true,
    } as any)

    render(<OAuthConnectButton providerKey="github" />, { wrapper })

    expect(screen.getByRole('button')).toBeDisabled()
  })

  it('should apply custom className', () => {
    vi.mocked(useOAuthHook.useAuthorize).mockReturnValue({
      mutateAsync: vi.fn(),
      isPending: false,
    } as any)

    render(<OAuthConnectButton providerKey="github" className="custom-class" />, { wrapper })

    const button = screen.getByRole('button')
    expect(button).toHaveClass('custom-class')
  })

  it('should call onSuccess callback when connection succeeds', async () => {
    const onSuccess = vi.fn()
    const mockMutateAsync = vi.fn().mockResolvedValue('https://auth.url')
    const mockOpenPopup = vi.fn().mockResolvedValue(undefined)

    vi.mocked(useOAuthHook.useAuthorize).mockReturnValue({
      mutateAsync: mockMutateAsync,
      isPending: false,
    } as any)
    vi.mocked(useOAuthHook.openOAuthPopup).mockImplementation(mockOpenPopup)

    const user = userEvent.setup()
    render(<OAuthConnectButton providerKey="github" onSuccess={onSuccess} />, { wrapper })

    await user.click(screen.getByRole('button'))

    // Wait for async operations
    await vi.waitFor(() => {
      expect(onSuccess).toHaveBeenCalled()
    })
  })

  it('should call onError callback when connection fails', async () => {
    const onError = vi.fn()
    const mockError = new Error('Connection failed')
    const mockMutateAsync = vi.fn().mockRejectedValue(mockError)

    vi.mocked(useOAuthHook.useAuthorize).mockReturnValue({
      mutateAsync: mockMutateAsync,
      isPending: false,
    } as any)

    const user = userEvent.setup()
    render(<OAuthConnectButton providerKey="github" onError={onError} />, { wrapper })

    await user.click(screen.getByRole('button'))

    // Wait for async operations
    await vi.waitFor(() => {
      expect(onError).toHaveBeenCalledWith(mockError)
    })
  })
})
