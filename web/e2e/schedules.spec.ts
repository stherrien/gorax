import { test, expect, waitForLoading, navigateTo, fillFormField, clickButton, expectToast, searchFor } from './setup'

test.describe('Schedules Management', () => {
  test.beforeEach(async ({ authenticatedPage }) => {
    await navigateTo(authenticatedPage, 'Schedules')
  })

  test('should display schedules list', async ({ authenticatedPage: page }) => {
    await waitForLoading(page)

    await expect(page.locator('h1, h2', { hasText: /Schedules/i })).toBeVisible()

    const hasSchedules = await page.locator('[data-testid="schedule-card"], [data-testid="schedule-item"]').count() > 0
    const hasEmptyState = await page.locator('[data-testid="empty-state"]').isVisible()

    expect(hasSchedules || hasEmptyState).toBeTruthy()
  })

  test('should create new schedule', async ({ authenticatedPage: page }) => {
    await waitForLoading(page)

    await clickButton(page, 'Create Schedule')

    await fillFormField(page, 'Name', 'Daily Test Schedule')
    await fillFormField(page, 'Cron Expression', '0 0 * * *')

    // Select workflow
    const workflowSelect = page.locator('select[name="workflow"], select[name="workflowId"]')
    if (await workflowSelect.isVisible()) {
      await workflowSelect.selectOption({ index: 0 })
    }

    await clickButton(page, 'Create')

    await expectToast(page, /created|success/i)
  })

  test('should edit schedule', async ({ authenticatedPage: page }) => {
    await waitForLoading(page)

    const scheduleCount = await page.locator('[data-testid="schedule-card"], [data-testid="schedule-item"]').count()

    if (scheduleCount > 0) {
      const editButton = page.locator('button[aria-label*="edit" i], button:has-text("Edit")').first()
      await editButton.click()

      await fillFormField(page, 'Name', 'Updated Schedule Name')

      await clickButton(page, 'Save')

      await expectToast(page, /updated|success/i)
    }
  })

  test('should delete schedule', async ({ authenticatedPage: page }) => {
    await waitForLoading(page)

    const initialCount = await page.locator('[data-testid="schedule-card"], [data-testid="schedule-item"]').count()

    if (initialCount > 0) {
      const deleteButton = page.locator('button[aria-label*="delete" i], button:has-text("Delete")').first()
      await deleteButton.click()

      await clickButton(page, 'Confirm')

      await expectToast(page, /deleted|success/i)
    }
  })

  test('should enable/disable schedule', async ({ authenticatedPage: page }) => {
    await waitForLoading(page)

    const scheduleCount = await page.locator('[data-testid="schedule-card"], [data-testid="schedule-item"]').count()

    if (scheduleCount > 0) {
      const toggleButton = page.locator('button[role="switch"], input[type="checkbox"]').first()

      if (await toggleButton.isVisible()) {
        await toggleButton.click()
        await page.waitForTimeout(1000)
      }
    }
  })

  test('should view schedule history', async ({ authenticatedPage: page }) => {
    await waitForLoading(page)

    const scheduleCount = await page.locator('[data-testid="schedule-card"], [data-testid="schedule-item"]').count()

    if (scheduleCount > 0) {
      await page.locator('[data-testid="schedule-card"], [data-testid="schedule-item"]').first().click()

      const hasHistory = await page.locator('text=/history|executions|runs/i').isVisible()
      expect(hasHistory).toBeTruthy()
    }
  })
})

test.describe('Schedule Form Validation', () => {
  test.beforeEach(async ({ authenticatedPage }) => {
    await navigateTo(authenticatedPage, 'Schedules')
    await waitForLoading(authenticatedPage)
    await clickButton(authenticatedPage, 'Create Schedule')
  })

  test('should show error when name is empty', async ({ authenticatedPage: page }) => {
    // Leave name empty, fill cron expression
    await fillFormField(page, 'Cron Expression', '0 0 * * *')

    await clickButton(page, 'Create')

    // Should show validation error for name
    const errorMessage = page.locator('text=/name.*required|required.*name/i')
    await expect(errorMessage).toBeVisible({ timeout: 5000 })
  })

  test('should show error for invalid cron expression', async ({ authenticatedPage: page }) => {
    await fillFormField(page, 'Name', 'Test Schedule')

    // Enter invalid cron expression
    await fillFormField(page, 'Cron Expression', 'invalid cron')

    await clickButton(page, 'Create')

    // Should show validation error for cron
    const errorMessage = page.locator('text=/invalid.*cron|cron.*invalid/i')
    await expect(errorMessage).toBeVisible({ timeout: 5000 })
  })

  test('should display cron preview for valid expression', async ({ authenticatedPage: page }) => {
    await fillFormField(page, 'Name', 'Test Schedule')
    await fillFormField(page, 'Cron Expression', '0 9 * * 1-5')

    // Should show schedule preview
    const preview = page.locator('text=/next.*run|weekday|monday.*friday/i')
    await expect(preview).toBeVisible({ timeout: 5000 })
  })

  test('should allow timezone selection', async ({ authenticatedPage: page }) => {
    await fillFormField(page, 'Name', 'Test Schedule')
    await fillFormField(page, 'Cron Expression', '0 0 * * *')

    // Find and change timezone
    const timezoneSelect = page.locator('select[name*="timezone" i], [data-testid="timezone-select"]')
    if (await timezoneSelect.isVisible()) {
      await timezoneSelect.selectOption('America/New_York')
      await expect(timezoneSelect).toHaveValue('America/New_York')
    }
  })

  test('should toggle enabled state', async ({ authenticatedPage: page }) => {
    await fillFormField(page, 'Name', 'Test Schedule')
    await fillFormField(page, 'Cron Expression', '0 0 * * *')

    // Find and toggle enabled checkbox
    const enabledCheckbox = page.locator('input[name*="enabled" i], [data-testid="enabled-checkbox"]')
    if (await enabledCheckbox.isVisible()) {
      const initialChecked = await enabledCheckbox.isChecked()
      await enabledCheckbox.click()
      const newChecked = await enabledCheckbox.isChecked()
      expect(newChecked).not.toBe(initialChecked)
    }
  })

  test('should display form error summary for multiple errors', async ({ authenticatedPage: page }) => {
    // Submit empty form
    await clickButton(page, 'Create')

    // Should show error summary or multiple field errors
    const errorSummary = page.locator('[data-testid="error-summary"], .error-summary')
    const fieldErrors = page.locator('.field-error, [role="alert"]')

    const hasErrorSummary = await errorSummary.isVisible()
    const hasFieldErrors = await fieldErrors.count() > 0

    expect(hasErrorSummary || hasFieldErrors).toBeTruthy()
  })
})
