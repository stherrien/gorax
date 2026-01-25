import { useState } from 'react'
import { useAuthorize, openOAuthPopup } from '../../hooks/useOAuth'
import { useQueryClient } from '@tanstack/react-query'

interface OAuthConnectButtonProps {
  providerKey: string
  scopes?: string[]
  onSuccess?: () => void
  onError?: (error: Error) => void
  children?: React.ReactNode
  className?: string
}

export function OAuthConnectButton({
  providerKey,
  scopes,
  onSuccess,
  onError,
  children,
  className = 'px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 disabled:opacity-50',
}: OAuthConnectButtonProps) {
  const [isConnecting, setIsConnecting] = useState(false)
  const authorize = useAuthorize()
  const queryClient = useQueryClient()

  const handleConnect = async () => {
    setIsConnecting(true)
    try {
      // Get authorization URL
      const authUrl = await authorize.mutateAsync({
        provider_key: providerKey,
        scopes,
      })

      // Open OAuth popup
      await openOAuthPopup(authUrl, providerKey)

      // Refresh connections after popup closes
      await queryClient.invalidateQueries({ queryKey: ['oauth', 'connections'] })

      if (onSuccess) {
        onSuccess()
      }
    } catch (error) {
      console.error('OAuth connection failed:', error)
      if (onError && error instanceof Error) {
        onError(error)
      }
    } finally {
      setIsConnecting(false)
    }
  }

  return (
    <button onClick={handleConnect} disabled={isConnecting || authorize.isPending} className={className}>
      {isConnecting || authorize.isPending
        ? 'Connecting...'
        : children || `Connect ${providerKey}`}
    </button>
  )
}
