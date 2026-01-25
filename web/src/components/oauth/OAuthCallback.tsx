import { useEffect, useState } from 'react'
import { useNavigate, useParams, useSearchParams } from 'react-router-dom'
import { handleCallback } from '../../api/oauth'

export function OAuthCallback() {
  const { provider } = useParams<{ provider: string }>()
  const [searchParams] = useSearchParams()
  const navigate = useNavigate()
  const [status, setStatus] = useState<'processing' | 'success' | 'error'>('processing')
  const [error, setError] = useState<string>('')

  useEffect(() => {
    const processCallback = async () => {
      if (!provider) {
        setStatus('error')
        setError('Missing provider parameter')
        return
      }

      // Check for OAuth error
      const errorParam = searchParams.get('error')
      if (errorParam) {
        setStatus('error')
        setError(searchParams.get('error_description') || errorParam)
        return
      }

      // Get OAuth callback parameters
      const code = searchParams.get('code')
      const state = searchParams.get('state')

      if (!code || !state) {
        setStatus('error')
        setError('Missing required OAuth parameters')
        return
      }

      try {
        // Handle callback
        await handleCallback(provider, code, state)
        setStatus('success')

        // Close popup if in popup window
        if (window.opener) {
          // Notify parent window
          window.opener.postMessage({ type: 'oauth_success', provider }, window.location.origin)
          // Close popup after a short delay
          setTimeout(() => {
            window.close()
          }, 1500)
        } else {
          // Redirect to connections page after a short delay
          setTimeout(() => {
            navigate('/oauth/connections')
          }, 2000)
        }
      } catch (err) {
        setStatus('error')
        setError(err instanceof Error ? err.message : 'OAuth authentication failed')
      }
    }

    processCallback()
  }, [provider, searchParams, navigate])

  return (
    <div className="min-h-screen flex items-center justify-center bg-gray-50 px-4">
      <div className="max-w-md w-full bg-white rounded-lg shadow-lg p-8">
        {status === 'processing' && (
          <div className="text-center">
            <div className="inline-block animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600 mb-4"></div>
            <h2 className="text-xl font-semibold text-gray-900 mb-2">Connecting...</h2>
            <p className="text-gray-600">Please wait while we complete the OAuth flow.</p>
          </div>
        )}

        {status === 'success' && (
          <div className="text-center">
            <div className="inline-flex items-center justify-center w-12 h-12 bg-green-100 rounded-full mb-4">
              <svg
                className="w-6 h-6 text-green-600"
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
              >
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  strokeWidth={2}
                  d="M5 13l4 4L19 7"
                />
              </svg>
            </div>
            <h2 className="text-xl font-semibold text-gray-900 mb-2">Success!</h2>
            <p className="text-gray-600">
              Your {provider} account has been successfully connected.
            </p>
            {!window.opener && (
              <p className="text-sm text-gray-500 mt-4">Redirecting...</p>
            )}
          </div>
        )}

        {status === 'error' && (
          <div className="text-center">
            <div className="inline-flex items-center justify-center w-12 h-12 bg-red-100 rounded-full mb-4">
              <svg
                className="w-6 h-6 text-red-600"
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
              >
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  strokeWidth={2}
                  d="M6 18L18 6M6 6l12 12"
                />
              </svg>
            </div>
            <h2 className="text-xl font-semibold text-gray-900 mb-2">Connection Failed</h2>
            <p className="text-gray-600 mb-4">{error}</p>
            {!window.opener && (
              <button
                onClick={() => navigate('/oauth/connections')}
                className="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700"
              >
                Back to Connections
              </button>
            )}
            {window.opener && (
              <p className="text-sm text-gray-500">You can close this window.</p>
            )}
          </div>
        )}
      </div>
    </div>
  )
}
