/**
 * CSRF Protection Hook
 *
 * Provides CSRF token management for frontend applications.
 * Works with the backend CSRF middleware.
 */

import { useState, useCallback, useEffect } from 'react';

const CSRF_COOKIE_NAME = 'csrf_token';
const CSRF_HEADER_NAME = 'X-CSRF-Token';

/**
 * Get CSRF token from cookie
 */
function getCSRFTokenFromCookie(): string | null {
  const cookies = document.cookie.split(';');
  for (const cookie of cookies) {
    const [name, value] = cookie.trim().split('=');
    if (name === CSRF_COOKIE_NAME) {
      return value;
    }
  }
  return null;
}

/**
 * Hook to manage CSRF token
 */
export function useCSRF() {
  const [token, setToken] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  // Initialize token from cookie on mount
  useEffect(() => {
    const cookieToken = getCSRFTokenFromCookie();
    if (cookieToken) {
      setToken(cookieToken);
    }
  }, []);

  // Fetch fresh CSRF token from server
  const refreshToken = useCallback(async () => {
    setLoading(true);
    setError(null);

    try {
      const response = await fetch('/api/v1/csrf/token', {
        method: 'GET',
        credentials: 'include',
      });

      if (!response.ok) {
        throw new Error('Failed to fetch CSRF token');
      }

      const data = await response.json();
      setToken(data.token);
      return data.token;
    } catch (err) {
      const message = err instanceof Error ? err.message : 'Failed to fetch CSRF token';
      setError(message);
      return null;
    } finally {
      setLoading(false);
    }
  }, []);

  // Get current token (from state or cookie)
  const getToken = useCallback((): string | null => {
    if (token) return token;
    return getCSRFTokenFromCookie();
  }, [token]);

  // Get headers object with CSRF token
  const getHeaders = useCallback((): Record<string, string> => {
    const currentToken = getToken();
    if (currentToken) {
      return { [CSRF_HEADER_NAME]: currentToken };
    }
    return {};
  }, [getToken]);

  return {
    token,
    loading,
    error,
    getToken,
    getHeaders,
    refreshToken,
  };
}

/**
 * Create headers with CSRF token for fetch requests
 */
export function createCSRFHeaders(additionalHeaders?: Record<string, string>): Record<string, string> {
  const token = getCSRFTokenFromCookie();
  const headers: Record<string, string> = {
    ...additionalHeaders,
  };

  if (token) {
    headers[CSRF_HEADER_NAME] = token;
  }

  return headers;
}

/**
 * Wrapper for fetch that automatically includes CSRF token
 */
export async function csrfFetch(
  url: string,
  options: RequestInit = {}
): Promise<Response> {
  const method = (options.method ?? 'GET').toUpperCase();

  // Only add CSRF token for state-changing methods
  if (['POST', 'PUT', 'PATCH', 'DELETE'].includes(method)) {
    const token = getCSRFTokenFromCookie();
    if (token) {
      const headers = new Headers(options.headers);
      headers.set(CSRF_HEADER_NAME, token);
      options.headers = headers;
    }
  }

  // Always include credentials to send/receive cookies
  options.credentials = 'include';

  return fetch(url, options);
}

export default useCSRF;
