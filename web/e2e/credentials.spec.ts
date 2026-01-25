import { test, expect, waitForLoading, navigateTo, fillFormField, clickButton, expectToast, expectError, searchFor, expectTableRow } from './setup'

test.describe('Credentials Management', () => {
  test.beforeEach(async ({ authenticatedPage }) => {
    await navigateTo(authenticatedPage, 'Credentials')
  })

  test('should display credentials list', async ({ authenticatedPage: page }) => {
    await waitForLoading(page)

    // Verify page heading
    await expect(page.locator('h1, h2', { hasText: /Credentials/i })).toBeVisible()

    // Should show credentials or empty state
    const hasCredentials = await page.locator('[data-testid="credential-card"], [data-testid="credential-item"]').count() > 0
    const hasEmptyState = await page.locator('[data-testid="empty-state"]').isVisible()

    expect(hasCredentials || hasEmptyState).toBeTruthy()
  })

  test('should create new credential', async ({ authenticatedPage: page }) => {
    await waitForLoading(page)

    // Click create button
    await clickButton(page, 'Create Credential')

    // Fill credential form
    await fillFormField(page, 'Name', 'Test API Key')
    await fillFormField(page, 'Description', 'API key for testing')

    // Select credential type
    await page.selectOption('[name="type"]', 'api_key')

    // Fill credential data
    await fillFormField(page, 'API Key', 'test-api-key-12345')

    // Submit form
    await clickButton(page, 'Create')

    // Verify success
    await expectToast(page, /created|success/i)

    // Verify credential appears in list
    await expectTableRow(page, 'Test API Key')
  })

  test('should search credentials', async ({ authenticatedPage: page }) => {
    await waitForLoading(page)

    const credentialCount = await page.locator('[data-testid="credential-card"], [data-testid="credential-item"]').count()

    if (credentialCount > 0) {
      // Get first credential name
      const firstCredName = await page.locator('[data-testid="credential-name"]').first().textContent()

      // Search for it
      await searchFor(page, firstCredName || 'test')

      // Verify results
      const visibleCount = await page.locator('[data-testid="credential-card"]:visible, [data-testid="credential-item"]:visible').count()
      expect(visibleCount).toBeGreaterThan(0)
    }
  })
})
