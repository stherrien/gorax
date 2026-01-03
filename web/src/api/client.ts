// Custom error classes
export class APIError extends Error {
  constructor(
    message: string,
    public status?: number,
    public response?: any
  ) {
    super(message)
    this.name = 'APIError'
  }
}

export class AuthError extends APIError {
  constructor(message: string, status: number, response?: any) {
    super(message, status, response)
    this.name = 'AuthError'
  }
}

export class NotFoundError extends APIError {
  constructor(message: string, status: number, response?: any) {
    super(message, status, response)
    this.name = 'NotFoundError'
  }
}

export class ValidationError extends APIError {
  constructor(message: string, status: number, response?: any) {
    super(message, status, response)
    this.name = 'ValidationError'
  }
}

export class ServerError extends APIError {
  constructor(message: string, status: number, response?: any) {
    super(message, status, response)
    this.name = 'ServerError'
  }
}

export class NetworkError extends APIError {
  constructor(message: string, originalError?: Error) {
    super(message)
    this.name = 'NetworkError'
    if (originalError) {
      this.stack = originalError.stack
    }
  }
}

interface RequestOptions {
  params?: Record<string, any>
  headers?: Record<string, string>
  retries?: number
  timeout?: number
}

export class APIClient {
  constructor(private baseURL: string) {}

  private getAuthToken(): string | null {
    return localStorage.getItem('auth_token')
  }

  private buildURL(path: string, params?: Record<string, any>): string {
    const url = `${this.baseURL}${path}`
    if (!params || Object.keys(params).length === 0) {
      return url
    }

    const searchParams = new URLSearchParams()
    Object.entries(params).forEach(([key, value]) => {
      searchParams.append(key, String(value))
    })

    return `${url}?${searchParams.toString()}`
  }

  private getHeaders(customHeaders?: Record<string, string>): HeadersInit {
    const headers: Record<string, string> = {
      'Content-Type': 'application/json',
      ...customHeaders,
    }

    const token = this.getAuthToken()
    if (token) {
      headers['Authorization'] = `Bearer ${token}`
    }

    // Add tenant ID header for development mode
    // In production, this would come from the auth token
    if (!token) {
      headers['X-Tenant-ID'] = '00000000-0000-0000-0000-000000000001'
    }

    return headers
  }

  private async fetchWithTimeout(
    url: string,
    options: RequestInit,
    timeout?: number
  ): Promise<Response> {
    if (!timeout) {
      return fetch(url, options)
    }

    const controller = new AbortController()
    const timeoutId = setTimeout(() => controller.abort(), timeout)

    try {
      const response = await fetch(url, {
        ...options,
        signal: controller.signal,
      })
      clearTimeout(timeoutId)
      return response
    } catch (error: any) {
      clearTimeout(timeoutId)
      if (error.name === 'AbortError') {
        throw new NetworkError('Request timeout')
      }
      throw error
    }
  }

  private async handleResponse(response: Response): Promise<any> {
    // Handle successful responses
    if (response.ok) {
      // 204 No Content
      if (response.status === 204) {
        return {}
      }
      return await response.json()
    }

    // Handle error responses
    // Read body as text first, then try to parse as JSON
    // (avoids "body stream already read" error)
    let errorData: any
    const text = await response.text()
    try {
      errorData = JSON.parse(text)
    } catch {
      errorData = { error: text || response.statusText }
    }

    const errorMessage = errorData.error || errorData.message || response.statusText

    // Categorize errors by status code
    if (response.status === 401 || response.status === 403) {
      // Handle authentication/authorization errors
      const authError = new AuthError(errorMessage, response.status, errorData)

      // If 401, redirect to login (unless already on login page)
      if (response.status === 401 && !window.location.pathname.includes('/login')) {
        // Store the current path for redirect after login
        localStorage.setItem('redirect_after_login', window.location.pathname)

        // In development, just show error
        // In production, redirect to Kratos login
        if (import.meta.env.MODE === 'production') {
          // TODO: Redirect to Kratos login URL
          console.error('Unauthorized - redirect to login')
        } else {
          console.error('Unauthorized (dev mode) - ', errorMessage)
        }
      }

      throw authError
    }

    if (response.status === 404) {
      throw new NotFoundError(errorMessage, response.status, errorData)
    }

    if (response.status === 400 || response.status === 422) {
      throw new ValidationError(errorMessage, response.status, errorData)
    }

    if (response.status >= 500) {
      throw new ServerError(errorMessage, response.status, errorData)
    }

    throw new APIError(errorMessage, response.status, errorData)
  }

  private isRetryableError(error: any): boolean {
    // Only retry server errors (5xx)
    return error instanceof ServerError
  }

  private async sleep(ms: number): Promise<void> {
    return new Promise((resolve) => setTimeout(resolve, ms))
  }

  private async request(
    method: string,
    path: string,
    options: RequestOptions & { body?: any } = {}
  ): Promise<any> {
    const { params, headers, retries = 0, timeout, body } = options

    const url = this.buildURL(path, params)
    const requestHeaders = this.getHeaders(headers)

    const fetchOptions: RequestInit = {
      method,
      headers: requestHeaders,
    }

    if (body !== undefined) {
      fetchOptions.body = JSON.stringify(body)
    }

    let lastError: any
    const maxAttempts = retries + 1

    for (let attempt = 0; attempt < maxAttempts; attempt++) {
      try {
        const response = await this.fetchWithTimeout(url, fetchOptions, timeout)
        return await this.handleResponse(response)
      } catch (error: any) {
        lastError = error

        // If this is a fetch/network error, wrap it
        if (!(error instanceof APIError)) {
          lastError = new NetworkError(error.message, error)
        }

        // Don't retry if it's not a retryable error or if we're out of attempts
        const isLastAttempt = attempt === maxAttempts - 1
        if (!this.isRetryableError(lastError) || isLastAttempt) {
          throw lastError
        }

        // Wait before retrying (exponential backoff)
        const delay = Math.min(1000 * Math.pow(2, attempt), 5000)
        await this.sleep(delay)
      }
    }

    throw lastError
  }

  async get(path: string, options?: RequestOptions): Promise<any> {
    return this.request('GET', path, options)
  }

  async post(path: string, body: any, options?: RequestOptions): Promise<any> {
    return this.request('POST', path, { ...options, body })
  }

  async put(path: string, body: any, options?: RequestOptions): Promise<any> {
    return this.request('PUT', path, { ...options, body })
  }

  async delete(path: string, options?: RequestOptions): Promise<any> {
    return this.request('DELETE', path, options)
  }

  async patch(path: string, body: any, options?: RequestOptions): Promise<any> {
    return this.request('PATCH', path, { ...options, body })
  }
}

// Default API client instance
// In development, use empty base URL so requests go through Vite proxy
// In production, use VITE_API_URL or fallback to same-origin
const apiBaseURL = import.meta.env.VITE_API_URL || ''
export const apiClient = new APIClient(apiBaseURL)
