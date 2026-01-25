import { Page, expect } from '@playwright/test'

/**
 * Test helper utilities for E2E tests
 */

export const TEST_USER = {
  tenantId: 'test-tenant',
  userId: 'test-user',
  email: 'test@example.com'
}

export const API_BASE_URL = process.env.API_BASE_URL || 'http://localhost:8080'

/**
 * Wait for page to be fully loaded with no network activity
 */
export async function waitForPageLoad(page: Page) {
  await page.waitForLoadState('networkidle')
}

/**
 * Navigate to a page and wait for it to load
 */
export async function navigateAndWait(page: Page, path: string) {
  await page.goto(path)
  await waitForPageLoad(page)
}

/**
 * Fill a form field and wait for any validation
 */
export async function fillField(page: Page, selector: string, value: string) {
  await page.fill(selector, value)
  await page.waitForTimeout(100) // Brief wait for validation
}

/**
 * Click a button and wait for any resulting navigation/state change
 */
export async function clickAndWait(page: Page, selector: string) {
  await page.click(selector)
  await page.waitForTimeout(500) // Wait for UI updates
}

/**
 * Wait for a success toast/notification to appear
 */
export async function waitForSuccessMessage(page: Page) {
  await expect(page.locator('[data-testid="toast-success"]').or(
    page.locator('text=/saved|created|updated|deleted|success/i').first()
  )).toBeVisible({ timeout: 5000 })
}

/**
 * Wait for an error toast/notification to appear
 */
export async function waitForErrorMessage(page: Page) {
  await expect(page.locator('[data-testid="toast-error"]').or(
    page.locator('text=/error|failed/i').first()
  )).toBeVisible({ timeout: 5000 })
}

/**
 * Check if an element is visible on the page
 */
export async function expectVisible(page: Page, selector: string, timeout = 5000) {
  await expect(page.locator(selector)).toBeVisible({ timeout })
}

/**
 * Check if text is visible on the page
 */
export async function expectTextVisible(page: Page, text: string | RegExp, timeout = 5000) {
  await expect(page.locator(`text=${text instanceof RegExp ? text.source : text}`)).toBeVisible({ timeout })
}

/**
 * Check if the current URL matches a pattern
 */
export async function expectUrl(page: Page, pattern: string | RegExp) {
  await expect(page).toHaveURL(pattern)
}

/**
 * Take a screenshot with a descriptive name
 */
export async function takeScreenshot(page: Page, name: string) {
  await page.screenshot({
    path: `tests/e2e/screenshots/${name}.png`,
    fullPage: true
  })
}

/**
 * Generate a unique test ID for entities
 */
export function generateTestId(prefix: string): string {
  return `${prefix}-${Date.now()}-${Math.random().toString(36).substring(7)}`
}

/**
 * Wait for API request to complete
 */
export async function waitForApiResponse(page: Page, urlPattern: string | RegExp) {
  const response = await page.waitForResponse(
    response => {
      const url = response.url()
      if (typeof urlPattern === 'string') {
        return url.includes(urlPattern)
      }
      return urlPattern.test(url)
    },
    { timeout: 10000 }
  )
  return response
}

/**
 * Mock API response for testing
 */
export async function mockApiResponse(
  page: Page,
  urlPattern: string | RegExp,
  responseData: any,
  status = 200
) {
  await page.route(urlPattern, route => {
    route.fulfill({
      status,
      contentType: 'application/json',
      body: JSON.stringify(responseData)
    })
  })
}

/**
 * Clear all mocked routes
 */
export async function clearMocks(page: Page) {
  await page.unroute('**/*')
}

/**
 * Check console for errors
 */
export function setupConsoleErrorTracking(page: Page): string[] {
  const consoleErrors: string[] = []

  page.on('console', msg => {
    if (msg.type() === 'error') {
      consoleErrors.push(msg.text())
    }
  })

  page.on('pageerror', error => {
    consoleErrors.push(error.message)
  })

  return consoleErrors
}

/**
 * Assert no console errors occurred
 */
export function assertNoConsoleErrors(consoleErrors: string[]) {
  if (consoleErrors.length > 0) {
    throw new Error(`Console errors detected:\n${consoleErrors.join('\n')}`)
  }
}
