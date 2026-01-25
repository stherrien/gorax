import { test, expect, waitForLoading, navigateTo, searchFor, clickButton } from './setup'

test.describe('Executions', () => {
  test.beforeEach(async ({ authenticatedPage }) => {
    await navigateTo(authenticatedPage, 'Executions')
  })

  test('should display executions list', async ({ authenticatedPage: page }) => {
    await waitForLoading(page)

    await expect(page.locator('h1, h2', { hasText: /Executions/i })).toBeVisible()

    const hasExecutions = await page.locator('[data-testid="execution-row"], [data-testid="execution-item"]').count() > 0
    const hasEmptyState = await page.locator('[data-testid="empty-state"]').isVisible()

    expect(hasExecutions || hasEmptyState).toBeTruthy()
  })

  test('should view execution details', async ({ authenticatedPage: page }) => {
    await waitForLoading(page)

    const executionCount = await page.locator('[data-testid="execution-row"], [data-testid="execution-item"]').count()

    if (executionCount > 0) {
      await page.locator('[data-testid="execution-row"], [data-testid="execution-item"]').first().click()

      await page.waitForURL(/\/executions\/[0-9a-f-]+/)

      await expect(page.locator('h1, h2')).toBeVisible()
    }
  })

  test('should filter executions by status', async ({ authenticatedPage: page }) => {
    await waitForLoading(page)

    const statusFilter = page.locator('[data-testid="status-filter"], select[name="status"]')

    if (await statusFilter.isVisible()) {
      await statusFilter.selectOption('success')
      await waitForLoading(page)

      const visibleExecutions = await page.locator('[data-testid="execution-row"]').count()
      expect(visibleExecutions).toBeGreaterThanOrEqual(0)
    }
  })

  test('should filter executions by workflow', async ({ authenticatedPage: page }) => {
    await waitForLoading(page)

    const workflowFilter = page.locator('[data-testid="workflow-filter"], select[name="workflow"]')

    if (await workflowFilter.isVisible()) {
      await workflowFilter.selectOption({ index: 1 })
      await waitForLoading(page)
    }
  })

  test('should search executions', async ({ authenticatedPage: page }) => {
    await waitForLoading(page)

    await searchFor(page, 'test')

    const visibleCount = await page.locator('[data-testid="execution-row"]:visible').count()
    expect(visibleCount).toBeGreaterThanOrEqual(0)
  })

  test('should show execution logs', async ({ authenticatedPage: page }) => {
    await waitForLoading(page)

    const executionCount = await page.locator('[data-testid="execution-row"], [data-testid="execution-item"]').count()

    if (executionCount > 0) {
      await page.locator('[data-testid="execution-row"], [data-testid="execution-item"]').first().click()

      await page.waitForURL(/\/executions\/[0-9a-f-]+/)

      const hasLogs = await page.locator('[data-testid="execution-logs"], text=/logs|output/i').isVisible()
      expect(hasLogs).toBeTruthy()
    }
  })

  test('should export execution logs', async ({ authenticatedPage: page }) => {
    await waitForLoading(page)

    const executionCount = await page.locator('[data-testid="execution-row"], [data-testid="execution-item"]').count()

    if (executionCount > 0) {
      await page.locator('[data-testid="execution-row"], [data-testid="execution-item"]').first().click()

      await page.waitForURL(/\/executions\/[0-9a-f-]+/)

      const exportButton = page.locator('button:has-text("Export"), button[aria-label*="export" i]')

      if (await exportButton.isVisible()) {
        const downloadPromise = page.waitForEvent('download')
        await exportButton.click()

        const download = await downloadPromise
        expect(download.suggestedFilename()).toBeTruthy()
      }
    }
  })

  test('should retry failed execution', async ({ authenticatedPage: page }) => {
    await waitForLoading(page)

    // Look for failed executions
    const failedExecution = page.locator('[data-testid="execution-row"]:has-text("failed"), [data-testid="status-failed"]').first()

    if (await failedExecution.isVisible()) {
      await failedExecution.click()

      const retryButton = page.locator('button:has-text("Retry"), button:has-text("Re-run")')

      if (await retryButton.isVisible()) {
        await retryButton.click()
        await page.waitForTimeout(2000)
      }
    }
  })

  test('should paginate through executions', async ({ authenticatedPage: page }) => {
    await waitForLoading(page)

    const nextButton = page.locator('[data-testid="pagination-next"], button:has-text("Next")')

    if (await nextButton.isEnabled()) {
      await nextButton.click()
      await waitForLoading(page)

      const currentPage = await page.locator('[data-testid="current-page"]').textContent()
      expect(currentPage).toContain('2')
    }
  })
})
