import { useParams } from 'react-router-dom'

/**
 * Reserved route keywords that should never be treated as resource IDs
 */
const RESERVED_ROUTES = ['new', 'create', 'edit', 'list', 'settings'] as const

/**
 * UUID validation regex (RFC 4122 compliant)
 */
const UUID_REGEX = /^[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/i

/**
 * Validates if a string is a valid resource ID
 * @param id - The ID string to validate
 * @returns true if the ID is valid and not a reserved keyword
 */
export function isValidResourceId(id: string | undefined | null): id is string {
  if (!id) return false

  // Check if it's a reserved keyword
  if (RESERVED_ROUTES.includes(id.toLowerCase() as any)) {
    return false
  }

  // Validate UUID format
  return UUID_REGEX.test(id)
}

/**
 * Custom hook to get and validate a resource ID from URL parameters
 * @param paramName - The parameter name (defaults to 'id')
 * @returns The validated resource ID or null if invalid
 */
export function useValidatedResourceId(paramName: string = 'id'): string | null {
  const params = useParams<Record<string, string>>()
  const id = params[paramName]

  return isValidResourceId(id) ? id : null
}

/**
 * Type guard to ensure an ID is validated before use
 * Use this in components to conditionally render based on ID validity
 *
 * @example
 * const id = useValidatedResourceId()
 * if (!id) return <Navigate to="/list" replace />
 * // Now TypeScript knows id is a valid string
 */

/**
 * Validates multiple IDs at once
 * @param ids - Object of ID values to validate
 * @returns Object with validation results
 */
export function validateResourceIds(
  ids: Record<string, string | undefined>
): Record<string, boolean> {
  const results: Record<string, boolean> = {}

  for (const [key, value] of Object.entries(ids)) {
    results[key] = isValidResourceId(value)
  }

  return results
}

/**
 * Extracts a valid ID from params or returns null
 * Useful for non-hook contexts
 * @param params - URLSearchParams or Record
 * @param paramName - The parameter name
 * @returns The validated ID or null
 */
export function getValidatedId(
  params: URLSearchParams | Record<string, string>,
  paramName: string = 'id'
): string | null {
  const id = params instanceof URLSearchParams
    ? params.get(paramName) ?? undefined
    : params[paramName]

  return isValidResourceId(id) ? id : null
}
