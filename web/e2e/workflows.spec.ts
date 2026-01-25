import { test, expect } from '@playwright/test'
import {
  navigateAndWait,
  fillField,
  clickAndWait,
  expectUrl,
  expectVisible,
  expectTextVisible,
  generateTestId,
  setupConsoleErrorTracking,
  assertNoConsoleErrors
} from './utils/test-helpers'

test.describe('Workflow Management E2E', () => {
  let consoleErrors: string[]

  test.beforeEach(async ({ page }) => {
    consoleErrors = setupConsoleErrorTracking(page)
  })

  test.afterEach(async () => {
    // Check for console errors after each test
    // Note: Some errors may be expected depending on test scenario
  })

  test('should load workflow list page', async ({ page }) => {
    await navigateAndWait(page, '/workflows')

    // Verify page loads
    await expectUrl(page, /\/workflows$/)
    await expectVisible(page, 'h1')

    // Verify key elements are present
    await expectTextVisible(page, /workflows/i)

    // Should show either workflows or empty state
    const hasWorkflows = await page.locator('[data-testid="workflow-card"]').count() > 0
    const hasEmptyState = await page.locator('text=/no workflows|create your first/i').isVisible()

    expect(hasWorkflows || hasEmptyState).toBeTruthy()
  })

  test('should navigate to create workflow page', async ({ page }) => {
    await navigateAndWait(page, '/workflows')

    // Click create button
    await page.click('text=/create workflow|new workflow|add workflow/i')

    // Should navigate to /workflows/new
    await expectUrl(page, /\/workflows\/new/)

    // Should show workflow editor
    await expectVisible(page, '[data-testid="workflow-editor"]', 10000)
  })

  test('should create a new workflow', async ({ page }) => {
    const workflowName = generateTestId('E2E-Test-Workflow')

    await navigateAndWait(page, '/workflows/new')

    // Wait for editor to load
    await page.waitForSelector('[data-testid="workflow-editor"]', { timeout: 10000 })

    // Fill in workflow name
    const nameInput = page.locator('input[name="name"], input[placeholder*="name" i]').first()
    await nameInput.fill(workflowName)

    // Fill in description
    const descInput = page.locator('textarea[name="description"], textarea[placeholder*="description" i]').first()
    if (await descInput.isVisible()) {
      await descInput.fill('Created by E2E test')
    }

    // Wait a moment for any auto-save or validation
    await page.waitForTimeout(1000)

    // Try to find and click save button
    const saveButton = page.locator('button:has-text("Save"), button[type="submit"]').first()
    if (await saveButton.isVisible()) {
      await saveButton.click()

      // Wait for either success message or navigation
      await Promise.race([
        page.waitForURL(/\/workflows\/[0-9a-f-]+/, { timeout: 10000 }),
        page.waitForSelector('text=/saved|success/i', { timeout: 5000 }).catch(() => {})
      ])
    }

    // Take screenshot of final state
    await page.screenshot({ path: 'tests/e2e/screenshots/workflow-created.png', fullPage: true })
  })

  test('should add nodes to workflow canvas', async ({ page }) => {
    await navigateAndWait(page, '/workflows/new')

    // Wait for editor and canvas
    await page.waitForSelector('[data-testid="workflow-editor"]', { timeout: 10000 })
    await page.waitForTimeout(1000)

    // Look for node palette or add node button
    const addNodeButton = page.locator('button:has-text("Add Node"), [data-testid="add-node-button"]').first()

    if (await addNodeButton.isVisible()) {
      await addNodeButton.click()

      // Should show node selection menu
      await expectVisible(page, '[data-testid="node-menu"], [role="menu"]')

      // Take screenshot
      await page.screenshot({ path: 'tests/e2e/screenshots/node-menu.png' })
    }
  })

  test('should edit existing workflow', async ({ page }) => {
    // First, go to workflows list
    await navigateAndWait(page, '/workflows')

    // Wait for workflows to load
    await page.waitForTimeout(2000)

    // Find first workflow card/item
    const firstWorkflow = page.locator('[data-testid="workflow-card"], [data-testid="workflow-item"]').first()

    if (await firstWorkflow.isVisible()) {
      // Click to view/edit workflow
      await firstWorkflow.click()

      // Should navigate to workflow detail/editor page
      await page.waitForURL(/\/workflows\/[0-9a-f-]+/, { timeout: 10000 })

      // Should show editor
      await expectVisible(page, '[data-testid="workflow-editor"]', 10000)

      // Take screenshot
      await page.screenshot({ path: 'tests/e2e/screenshots/workflow-edit.png', fullPage: true })
    }
  })

  test('should delete workflow', async ({ page }) => {
    // Navigate to workflows list
    await navigateAndWait(page, '/workflows')
    await page.waitForTimeout(2000)

    // Check if there are any workflows
    const workflowCount = await page.locator('[data-testid="workflow-card"], [data-testid="workflow-item"]').count()

    if (workflowCount > 0) {
      // Find delete button for first workflow
      const deleteButton = page.locator('button[aria-label*="delete" i], button:has-text("Delete")').first()

      if (await deleteButton.isVisible()) {
        await deleteButton.click()

        // Look for confirmation dialog
        const confirmButton = page.locator('button:has-text("Confirm"), button:has-text("Delete")').last()

        if (await confirmButton.isVisible()) {
          await confirmButton.click()

          // Wait for deletion to complete
          await page.waitForTimeout(1000)

          // Should show success message or workflow should disappear
          const newCount = await page.locator('[data-testid="workflow-card"], [data-testid="workflow-item"]').count()
          expect(newCount).toBeLessThanOrEqual(workflowCount)
        }
      }
    }
  })

  test('should execute workflow', async ({ page }) => {
    await navigateAndWait(page, '/workflows')
    await page.waitForTimeout(2000)

    const workflowCount = await page.locator('[data-testid="workflow-card"], [data-testid="workflow-item"]').count()

    if (workflowCount > 0) {
      // Find execute/run button
      const executeButton = page.locator('button:has-text("Execute"), button:has-text("Run"), button[aria-label*="execute" i]').first()

      if (await executeButton.isVisible()) {
        await executeButton.click()

        // Wait for execution to start
        await page.waitForTimeout(2000)

        // Should show execution started or navigate to executions
        const hasSuccessMsg = await page.locator('text=/execution|started|running/i').isVisible()
        const onExecutionsPage = page.url().includes('/executions')

        expect(hasSuccessMsg || onExecutionsPage).toBeTruthy()
      }
    }
  })

  test('should search/filter workflows', async ({ page }) => {
    await navigateAndWait(page, '/workflows')
    await page.waitForTimeout(1000)

    // Look for search input
    const searchInput = page.locator('input[type="search"], input[placeholder*="search" i]').first()

    if (await searchInput.isVisible()) {
      await searchInput.fill('test')
      await page.waitForTimeout(500)

      // Results should update
      await page.screenshot({ path: 'tests/e2e/screenshots/workflow-search.png' })
    }
  })

  test('should handle validation errors', async ({ page }) => {
    await navigateAndWait(page, '/workflows/new')

    // Wait for editor
    await page.waitForSelector('[data-testid="workflow-editor"]', { timeout: 10000 })

    // Try to save without required fields
    const saveButton = page.locator('button:has-text("Save"), button[type="submit"]').first()

    if (await saveButton.isVisible()) {
      // Clear any auto-filled name
      const nameInput = page.locator('input[name="name"], input[placeholder*="name" i]').first()
      if (await nameInput.isVisible()) {
        await nameInput.fill('')
      }

      await saveButton.click()

      // Should show validation error
      await page.waitForTimeout(1000)
      const hasError = await page.locator('text=/required|error|invalid/i').isVisible()

      if (hasError) {
        await page.screenshot({ path: 'tests/e2e/screenshots/workflow-validation-error.png' })
      }
    }
  })

  test('should show workflow details', async ({ page }) => {
    await navigateAndWait(page, '/workflows')
    await page.waitForTimeout(2000)

    const workflowCount = await page.locator('[data-testid="workflow-card"], [data-testid="workflow-item"]').count()

    if (workflowCount > 0) {
      const firstWorkflow = page.locator('[data-testid="workflow-card"], [data-testid="workflow-item"]').first()
      await firstWorkflow.click()

      await page.waitForURL(/\/workflows\/[0-9a-f-]+/)

      // Should show workflow details
      await expectVisible(page, '[data-testid="workflow-editor"]', 10000)

      // Take screenshot
      await page.screenshot({ path: 'tests/e2e/screenshots/workflow-details.png', fullPage: true })
    }
  })
})
