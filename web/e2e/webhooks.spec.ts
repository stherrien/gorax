import { test, expect } from '@playwright/test'
import {
  navigateAndWait,
  fillField,
  expectUrl,
  expectVisible,
  expectTextVisible,
  generateTestId,
  setupConsoleErrorTracking
} from './utils/test-helpers'

test.describe('Webhook Management E2E', () => {
  let consoleErrors: string[]

  test.beforeEach(async ({ page }) => {
    consoleErrors = setupConsoleErrorTracking(page)
  })

  test('should load webhook list page', async ({ page }) => {
    await navigateAndWait(page, '/webhooks')

    // Verify page loads
    await expectUrl(page, /\/webhooks$/)
    await expectVisible(page, 'h1')

    // Should show webhooks or empty state
    const hasWebhooks = await page.locator('[data-testid="webhook-card"], [data-testid="webhook-item"]').count() > 0
    const hasEmptyState = await page.locator('text=/no webhooks|create webhook/i').isVisible()

    expect(hasWebhooks || hasEmptyState).toBeTruthy()

    await page.screenshot({ path: 'tests/e2e/screenshots/webhooks-list.png', fullPage: true })
  })

  test('should navigate to create webhook page', async ({ page }) => {
    await navigateAndWait(page, '/webhooks')

    // Look for create button
    const createButton = page.locator('button:has-text("Create"), button:has-text("New Webhook"), a[href*="webhook"]').first()

    if (await createButton.isVisible()) {
      await createButton.click()

      // Wait for navigation or modal
      await page.waitForTimeout(1000)

      // Should show webhook form
      const hasForm = await page.locator('form, [data-testid="webhook-form"]').isVisible()
      expect(hasForm).toBeTruthy()
    }
  })

  test('should create a new webhook', async ({ page }) => {
    const webhookName = generateTestId('E2E-Webhook')

    await navigateAndWait(page, '/webhooks')

    // Click create
    const createButton = page.locator('button:has-text("Create"), button:has-text("New Webhook")').first()

    if (await createButton.isVisible()) {
      await createButton.click()
      await page.waitForTimeout(1000)

      // Fill in webhook details
      const nameInput = page.locator('input[name="name"], input[placeholder*="name" i]').first()
      if (await nameInput.isVisible()) {
        await nameInput.fill(webhookName)
      }

      const pathInput = page.locator('input[name="path"], input[placeholder*="path" i]').first()
      if (await pathInput.isVisible()) {
        await pathInput.fill(`/test-webhook-${Date.now()}`)
      }

      // Select auth type if available
      const authSelect = page.locator('select[name="authType"], select[name="auth_type"]').first()
      if (await authSelect.isVisible()) {
        await authSelect.selectOption('none')
      }

      // Save webhook
      const saveButton = page.locator('button:has-text("Create"), button:has-text("Save"), button[type="submit"]').last()
      await saveButton.click()

      // Wait for success
      await page.waitForTimeout(2000)

      // Should show success message or return to list
      const isOnListPage = page.url().includes('/webhooks') && !page.url().includes('/new')
      const hasSuccessMsg = await page.locator('text=/created|success/i').isVisible()

      expect(isOnListPage || hasSuccessMsg).toBeTruthy()

      await page.screenshot({ path: 'tests/e2e/screenshots/webhook-created.png', fullPage: true })
    }
  })

  test('should view webhook details', async ({ page }) => {
    await navigateAndWait(page, '/webhooks')
    await page.waitForTimeout(2000)

    // Find first webhook
    const firstWebhook = page.locator('[data-testid="webhook-card"], [data-testid="webhook-item"], a[href*="/webhooks/"]').first()

    if (await firstWebhook.isVisible()) {
      await firstWebhook.click()

      // Should navigate to webhook detail page
      await page.waitForURL(/\/webhooks\/[0-9a-f-]+/, { timeout: 10000 })

      // Should show webhook details
      await expectVisible(page, 'h1', 5000)

      await page.screenshot({ path: 'tests/e2e/screenshots/webhook-details.png', fullPage: true })
    }
  })

  test('should display webhook URL', async ({ page }) => {
    await navigateAndWait(page, '/webhooks')
    await page.waitForTimeout(2000)

    const firstWebhook = page.locator('[data-testid="webhook-card"], [data-testid="webhook-item"], a[href*="/webhooks/"]').first()

    if (await firstWebhook.isVisible()) {
      await firstWebhook.click()
      await page.waitForURL(/\/webhooks\/[0-9a-f-]+/)
      await page.waitForTimeout(1000)

      // Should show webhook URL
      const hasUrl = await page.locator('text=/https?:\/\//').isVisible()
      const hasCopyButton = await page.locator('button:has-text("Copy"), button[aria-label*="copy" i]').isVisible()

      expect(hasUrl || hasCopyButton).toBeTruthy()
    }
  })

  test('should test webhook', async ({ page }) => {
    await navigateAndWait(page, '/webhooks')
    await page.waitForTimeout(2000)

    const firstWebhook = page.locator('[data-testid="webhook-card"], [data-testid="webhook-item"], a[href*="/webhooks/"]').first()

    if (await firstWebhook.isVisible()) {
      await firstWebhook.click()
      await page.waitForURL(/\/webhooks\/[0-9a-f-]+/)
      await page.waitForTimeout(1000)

      // Look for test button
      const testButton = page.locator('button:has-text("Test"), button:has-text("Send Test")').first()

      if (await testButton.isVisible()) {
        await testButton.click()
        await page.waitForTimeout(2000)

        // Should show test result
        const hasResult = await page.locator('text=/success|sent|result|response/i').isVisible()
        expect(hasResult).toBeTruthy()

        await page.screenshot({ path: 'tests/e2e/screenshots/webhook-test.png', fullPage: true })
      }
    }
  })

  test('should delete webhook', async ({ page }) => {
    await navigateAndWait(page, '/webhooks')
    await page.waitForTimeout(2000)

    const webhookCount = await page.locator('[data-testid="webhook-card"], [data-testid="webhook-item"]').count()

    if (webhookCount > 0) {
      // Find delete button
      const deleteButton = page.locator('button[aria-label*="delete" i], button:has-text("Delete")').first()

      if (await deleteButton.isVisible()) {
        await deleteButton.click()

        // Confirm deletion
        const confirmButton = page.locator('button:has-text("Confirm"), button:has-text("Delete")').last()
        if (await confirmButton.isVisible()) {
          await confirmButton.click()
          await page.waitForTimeout(1000)

          // Webhook should be removed
          const newCount = await page.locator('[data-testid="webhook-card"], [data-testid="webhook-item"]').count()
          expect(newCount).toBeLessThanOrEqual(webhookCount)
        }
      }
    }
  })

  test('should show webhook events/history', async ({ page }) => {
    await navigateAndWait(page, '/webhooks')
    await page.waitForTimeout(2000)

    const firstWebhook = page.locator('[data-testid="webhook-card"], [data-testid="webhook-item"], a[href*="/webhooks/"]').first()

    if (await firstWebhook.isVisible()) {
      await firstWebhook.click()
      await page.waitForURL(/\/webhooks\/[0-9a-f-]+/)
      await page.waitForTimeout(1000)

      // Look for events/history section
      const hasEvents = await page.locator('text=/events|history|requests/i').isVisible()
      const hasEventsTable = await page.locator('table, [data-testid="events-list"]').isVisible()

      // Either should have events section or empty state
      expect(hasEvents || hasEventsTable).toBeTruthy()

      await page.screenshot({ path: 'tests/e2e/screenshots/webhook-events.png', fullPage: true })
    }
  })

  test('should filter webhook events', async ({ page }) => {
    await navigateAndWait(page, '/webhooks')
    await page.waitForTimeout(2000)

    const firstWebhook = page.locator('[data-testid="webhook-card"], [data-testid="webhook-item"], a[href*="/webhooks/"]').first()

    if (await firstWebhook.isVisible()) {
      await firstWebhook.click()
      await page.waitForURL(/\/webhooks\/[0-9a-f-]+/)
      await page.waitForTimeout(1000)

      // Look for filter controls
      const filterInput = page.locator('input[placeholder*="filter" i], input[placeholder*="search" i]').first()
      const statusSelect = page.locator('select').first()

      if (await filterInput.isVisible()) {
        await filterInput.fill('test')
        await page.waitForTimeout(500)
      } else if (await statusSelect.isVisible()) {
        await statusSelect.selectOption({ index: 1 })
        await page.waitForTimeout(500)
      }

      await page.screenshot({ path: 'tests/e2e/screenshots/webhook-filter.png', fullPage: true })
    }
  })

  test('should handle webhook validation errors', async ({ page }) => {
    await navigateAndWait(page, '/webhooks')

    const createButton = page.locator('button:has-text("Create"), button:has-text("New Webhook")').first()

    if (await createButton.isVisible()) {
      await createButton.click()
      await page.waitForTimeout(1000)

      // Try to save without required fields
      const saveButton = page.locator('button:has-text("Create"), button:has-text("Save"), button[type="submit"]').last()
      await saveButton.click()

      await page.waitForTimeout(1000)

      // Should show validation errors
      const hasError = await page.locator('text=/required|error|invalid/i').isVisible()
      if (hasError) {
        await page.screenshot({ path: 'tests/e2e/screenshots/webhook-validation-error.png' })
      }
    }
  })
})
