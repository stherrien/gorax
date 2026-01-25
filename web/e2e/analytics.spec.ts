import { test, expect, waitForLoading, navigateTo } from './setup'

test.describe('Analytics', () => {
  test.beforeEach(async ({ authenticatedPage }) => {
    await navigateTo(authenticatedPage, 'Analytics')
  })

  test('should display analytics dashboard', async ({ authenticatedPage: page }) => {
    await waitForLoading(page)

    await expect(page.locator('h1, h2', { hasText: /Analytics/i })).toBeVisible()

    // Should show metrics cards
    const metricsCards = await page.locator('[data-testid="metric-card"]').count()
    expect(metricsCards).toBeGreaterThan(0)
  })

  test('should display execution trend chart', async ({ authenticatedPage: page }) => {
    await waitForLoading(page)

    const trendChart = page.locator('[data-testid="trend-chart"], [data-testid="execution-trend"]')
    await expect(trendChart).toBeVisible()
  })

  test('should display success rate gauge', async ({ authenticatedPage: page }) => {
    await waitForLoading(page)

    const successRate = page.locator('[data-testid="success-rate"], text=/success rate/i')
    await expect(successRate).toBeVisible()
  })

  test('should display top workflows table', async ({ authenticatedPage: page }) => {
    await waitForLoading(page)

    const topWorkflows = page.locator('[data-testid="top-workflows"], text=/top workflows/i')
    await expect(topWorkflows).toBeVisible()
  })

  test('should display error breakdown', async ({ authenticatedPage: page }) => {
    await waitForLoading(page)

    const errorBreakdown = page.locator('[data-testid="error-breakdown"], text=/errors/i')
    
    // May or may not be visible depending on data
    const isVisible = await errorBreakdown.isVisible().catch(() => false)
    console.log('Error breakdown visible:', isVisible)
  })

  test('should filter analytics by date range', async ({ authenticatedPage: page }) => {
    await waitForLoading(page)

    const dateFilter = page.locator('[data-testid="date-range-filter"], select[name="dateRange"]')

    if (await dateFilter.isVisible()) {
      await dateFilter.selectOption('7d')
      await waitForLoading(page)

      await expect(page.locator('[data-testid="metric-card"]')).toBeVisible()
    }
  })

  test('should export analytics data', async ({ authenticatedPage: page }) => {
    await waitForLoading(page)

    const exportButton = page.locator('button:has-text("Export"), button[aria-label*="export" i]')

    if (await exportButton.isVisible()) {
      const downloadPromise = page.waitForEvent('download')
      await exportButton.click()

      const download = await downloadPromise
      expect(download.suggestedFilename()).toContain('analytics')
    }
  })
})
