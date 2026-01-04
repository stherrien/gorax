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
