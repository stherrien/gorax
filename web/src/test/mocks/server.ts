/**
 * MSW server setup for tests.
 * Intercepts API requests in Node.js test environment.
 */

import { setupServer } from 'msw/node'
import { handlers } from './handlers'

// Create server with default handlers
export const server = setupServer(...handlers)

// Re-export everything for convenience
export * from './handlers'
export * from './data'
