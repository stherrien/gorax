import { test, expect, waitForLoading, navigateTo, fillFormField, clickButton, expectToast, createTestWorkflow } from './setup'

test.describe('Critical User Flows', () => {
  test('complete workflow lifecycle', async ({ authenticatedPage: page }) => {
    await page.goto('/workflows')
    await waitForLoading(page)

    await clickButton(page, 'Create Workflow')
    await fillFormField(page, 'Name', 'E2E Lifecycle Test Workflow')
    await fillFormField(page, 'Description', 'Testing complete lifecycle')
    await clickButton(page, 'Create')
    await expectToast(page, /created|success/i)

    await page.waitForURL(/\/workflows\/[0-9a-f-]+/)
    await page.waitForTimeout(2000)

    const executeButton = page.locator('button:has-text("Execute"), button:has-text("Run")')
    if (await executeButton.isVisible()) {
      await executeButton.click()
      await page.waitForTimeout(2000)
    }

    await navigateTo(page, 'Executions')
    await waitForLoading(page)

    const executions = await page.locator('[data-testid="execution-row"]').count()
    expect(executions).toBeGreaterThan(0)
  })

  test('dashboard overview check', async ({ authenticatedPage: page }) => {
    await page.goto('/')
    await waitForLoading(page)

    const hasCharts = await page.locator('canvas, svg, [data-testid$="-chart"]').count() > 0
    expect(hasCharts).toBeTruthy()
  })
})
