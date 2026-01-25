import { afterEach, beforeAll, afterAll, vi } from 'vitest'
import { cleanup } from '@testing-library/react'
import '@testing-library/jest-dom/vitest'
import { server, resetMockIds } from './mocks/server'

// Start MSW server before all tests
beforeAll(() => {
  server.listen({ onUnhandledRequest: 'warn' })
})

// Reset handlers and mock IDs after each test
afterEach(() => {
  server.resetHandlers()
  resetMockIds()
})

// Close MSW server after all tests
afterAll(() => {
  server.close()
})

// Suppress React 18 act() warnings and React Router v7 future flag warnings in tests
// React Testing Library handles act() automatically, but some async updates
// may still trigger warnings. This configures the test environment to handle them.
const originalError = console.error
const originalWarn = console.warn
beforeAll(() => {
  console.error = (...args: any[]) => {
    if (
      typeof args[0] === 'string' &&
      args[0].includes('Warning: An update to') &&
      args[0].includes('was not wrapped in act')
    ) {
      return
    }
    originalError.call(console, ...args)
  }

  console.warn = (...args: any[]) => {
    if (
      typeof args[0] === 'string' &&
      (args[0].includes('React Router Future Flag Warning') ||
       args[0].includes('v7_startTransition') ||
       args[0].includes('v7_relativeSplatPath') ||
       args[0].includes('linearGradient') ||
       args[0].includes('incorrect casing'))
    ) {
      return
    }
    originalWarn.call(console, ...args)
  }
})

afterAll(() => {
  console.error = originalError
  console.warn = originalWarn
})

// Cleanup after each test
afterEach(() => {
  cleanup()
  vi.clearAllMocks()
})

// Mock localStorage
const localStorageMock = {
  getItem: vi.fn(),
  setItem: vi.fn(),
  removeItem: vi.fn(),
  clear: vi.fn(),
  length: 0,
  key: vi.fn(),
}
global.localStorage = localStorageMock as any

// Note: fetch is intercepted by MSW, not mocked directly
// For tests that need to control fetch without MSW, use server.use(...)
// to add one-off handlers
