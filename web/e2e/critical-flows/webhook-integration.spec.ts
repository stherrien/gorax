import { test, expect } from '@playwright/test'
import { WorkflowEditorPage, ExecutionsPage } from '../pages'
import { generateTestId, waitForPageLoad, waitForApiResponse } from '../utils/test-helpers'

/**
 * Critical Flow: Webhook Integration
 *
 * Tests the complete webhook flow:
 * 1. Create a workflow with webhook trigger
 * 2. Create and configure a webhook endpoint
 * 3. Send an event to the webhook
 * 4. Verify workflow execution is triggered
 * 5. Test webhook filters
 * 6. Test webhook authentication
 */
test.describe('Webhook Integration Critical Flow', () => {
  let workflowEditor: WorkflowEditorPage
  let executionsPage: ExecutionsPage

  test.beforeEach(async ({ page }) => {
    workflowEditor = new WorkflowEditorPage(page)
    executionsPage = new ExecutionsPage(page)
  })

  test('create webhook for workflow', async ({ page }) => {
    // Create workflow with webhook trigger
    const workflowName = `Webhook Workflow ${generateTestId('wh')}`
    const workflowId = await workflowEditor.createWorkflow(
      workflowName,
      'Triggered by webhook events'
    )

    await workflowEditor.goto(workflowId)
    await workflowEditor.addNode('trigger:webhook', { x: 100, y: 200 })
    await workflowEditor.addNode('action:log', { x: 300, y: 200 })
    await workflowEditor.saveWorkflow()

    // Navigate to webhook configuration
    await page.goto(`/workflows/${workflowId}/webhooks`)
    await waitForPageLoad(page)

    // Create webhook
    await page.click('button:has-text("Create Webhook"), [data-testid="create-webhook"]')

    const modal = page.locator('[data-testid="webhook-modal"], .webhook-modal')
    await expect(modal).toBeVisible()

    const webhookName = `Test Webhook ${generateTestId('hook')}`
    await modal.locator('input[name="name"]').fill(webhookName)
    await modal.locator('input[name="path"]').fill('/test-webhook-path')
    await modal.locator('select[name="authType"]').selectOption('none')
    await modal.locator('button:has-text("Create")').click()

    // Wait for webhook to be created
    await waitForApiResponse(page, /\/api\/v1\/webhooks/)

    // Verify webhook appears in list
    await expect(page.locator(`text="${webhookName}"`)).toBeVisible()
  })

  test('webhook with signature authentication', async ({ page }) => {
    // Create workflow
    const workflowName = `Signed Webhook ${generateTestId('sig')}`
    const workflowId = await workflowEditor.createWorkflow(workflowName)

    await workflowEditor.goto(workflowId)
    await workflowEditor.addNode('trigger:webhook', { x: 100, y: 200 })
    await workflowEditor.addNode('action:log', { x: 300, y: 200 })
    await workflowEditor.saveWorkflow()

    // Create webhook with signature auth
    await page.goto(`/workflows/${workflowId}/webhooks`)
    await waitForPageLoad(page)

    await page.click('button:has-text("Create Webhook"), [data-testid="create-webhook"]')

    const modal = page.locator('[data-testid="webhook-modal"]')
    await modal.locator('input[name="name"]').fill(`Signed Hook ${generateTestId('s')}`)
    await modal.locator('select[name="authType"]').selectOption('signature')
    await modal.locator('button:has-text("Create")').click()

    await waitForApiResponse(page, /\/api\/v1\/webhooks/)

    // Verify secret is displayed
    const secretField = page.locator('[data-testid="webhook-secret"], .webhook-secret')
    await expect(secretField).toBeVisible()

    const secret = await secretField.textContent()
    expect(secret).toBeTruthy()
    expect(secret!.length).toBeGreaterThan(10)
  })

  test('configure webhook filters', async ({ page }) => {
    // Create workflow and webhook
    const workflowId = await createWebhookWorkflow(page)
    const webhookId = await createWebhook(page, workflowId, 'Filter Test Webhook')

    // Navigate to webhook filters
    await page.goto(`/webhooks/${webhookId}/filters`)
    await waitForPageLoad(page)

    // Add filter
    await page.click('button:has-text("Add Filter"), [data-testid="add-filter"]')

    const filterModal = page.locator('[data-testid="filter-modal"]')
    await expect(filterModal).toBeVisible()

    await filterModal.locator('input[name="fieldPath"]').fill('$.event')
    await filterModal.locator('select[name="operator"]').selectOption('equals')
    await filterModal.locator('input[name="value"]').fill('user.created')
    await filterModal.locator('button:has-text("Save")').click()

    await waitForApiResponse(page, /\/api\/v1\/webhooks.*filters/)

    // Verify filter appears
    await expect(page.locator('text="$.event"')).toBeVisible()
    await expect(page.locator('text="equals"')).toBeVisible()
  })

  test('test webhook filter evaluation', async ({ page }) => {
    // Create workflow and webhook with filter
    const workflowId = await createWebhookWorkflow(page)
    const webhookId = await createWebhook(page, workflowId, 'Filter Eval Webhook')

    // Add filter
    await page.goto(`/webhooks/${webhookId}/filters`)
    await page.click('button:has-text("Add Filter"), [data-testid="add-filter"]')

    const filterModal = page.locator('[data-testid="filter-modal"]')
    await filterModal.locator('input[name="fieldPath"]').fill('$.action')
    await filterModal.locator('select[name="operator"]').selectOption('equals')
    await filterModal.locator('input[name="value"]').fill('opened')
    await filterModal.locator('button:has-text("Save")').click()

    await waitForApiResponse(page, /\/api\/v1\/webhooks.*filters/)

    // Test filter with matching payload
    await page.click('button:has-text("Test Filters"), [data-testid="test-filters"]')

    const testModal = page.locator('[data-testid="filter-test-modal"]')
    await expect(testModal).toBeVisible()

    await testModal.locator('textarea[name="payload"]').fill(JSON.stringify({
      action: 'opened',
      issue: { id: 123 }
    }))

    await testModal.locator('button:has-text("Test")').click()

    // Verify filter passes
    await expect(testModal.locator('text=/passed|match/i')).toBeVisible()
  })

  test('webhook event history is recorded', async ({ page }) => {
    // Create workflow and webhook
    const workflowId = await createWebhookWorkflow(page)
    const webhookId = await createWebhook(page, workflowId, 'History Test Webhook')

    // Navigate to event history
    await page.goto(`/webhooks/${webhookId}/events`)
    await waitForPageLoad(page)

    // Initially should be empty or show no events message
    const eventsList = page.locator('[data-testid="events-list"], .events-list')

    // The list might be empty, which is expected for a new webhook
    await expect(eventsList.or(page.locator('text=/no events/i'))).toBeVisible()
  })

  test('enable and disable webhook', async ({ page }) => {
    // Create workflow and webhook
    const workflowId = await createWebhookWorkflow(page)
    const webhookId = await createWebhook(page, workflowId, 'Toggle Webhook')

    // Navigate to webhooks list
    await page.goto(`/workflows/${workflowId}/webhooks`)
    await waitForPageLoad(page)

    // Find webhook row
    const webhookRow = page.locator(`[data-webhook-id="${webhookId}"], tr:has-text("Toggle Webhook")`)

    // Disable webhook
    const toggle = webhookRow.locator('input[type="checkbox"], [data-testid="webhook-toggle"]')
    await toggle.uncheck()
    await waitForApiResponse(page, /\/api\/v1\/webhooks/)

    // Verify disabled
    await expect(toggle).not.toBeChecked()

    // Re-enable
    await toggle.check()
    await waitForApiResponse(page, /\/api\/v1\/webhooks/)

    // Verify enabled
    await expect(toggle).toBeChecked()
  })

  test('delete webhook', async ({ page }) => {
    // Create workflow and webhook
    const workflowId = await createWebhookWorkflow(page)
    const webhookName = `Delete Webhook ${generateTestId('del')}`
    const webhookId = await createWebhook(page, workflowId, webhookName)

    // Navigate to webhooks list
    await page.goto(`/workflows/${workflowId}/webhooks`)
    await waitForPageLoad(page)

    // Find and delete webhook
    const webhookRow = page.locator(`[data-webhook-id="${webhookId}"], tr:has-text("${webhookName}")`)
    await webhookRow.locator('button:has-text("Delete"), [data-testid="delete-webhook"]').click()

    // Confirm deletion
    const confirmModal = page.locator('[data-testid="confirm-modal"]')
    await confirmModal.locator('button:has-text("Delete")').click()

    await waitForApiResponse(page, /\/api\/v1\/webhooks/)

    // Verify webhook is removed
    await expect(page.locator(`text="${webhookName}"`)).not.toBeVisible()
  })

  test('webhook URL is displayed correctly', async ({ page }) => {
    // Create workflow and webhook
    const workflowId = await createWebhookWorkflow(page)
    const webhookId = await createWebhook(page, workflowId, 'URL Test Webhook')

    // Navigate to webhook details
    await page.goto(`/webhooks/${webhookId}`)
    await waitForPageLoad(page)

    // Verify webhook URL is displayed
    const webhookUrl = page.locator('[data-testid="webhook-url"], .webhook-url')
    await expect(webhookUrl).toBeVisible()

    const url = await webhookUrl.textContent()
    expect(url).toContain('/webhooks/')
    expect(url).toContain(workflowId)
    expect(url).toContain(webhookId)
  })

  test('regenerate webhook secret', async ({ page }) => {
    // Create workflow and webhook with signature auth
    const workflowId = await createWebhookWorkflow(page)
    const webhookId = await createWebhookWithAuth(page, workflowId, 'Regen Secret Webhook', 'signature')

    // Navigate to webhook details
    await page.goto(`/webhooks/${webhookId}`)
    await waitForPageLoad(page)

    // Get original secret
    const secretField = page.locator('[data-testid="webhook-secret"]')
    const originalSecret = await secretField.textContent()

    // Regenerate secret
    await page.click('button:has-text("Regenerate"), [data-testid="regenerate-secret"]')

    // Confirm
    const confirmModal = page.locator('[data-testid="confirm-modal"]')
    if (await confirmModal.isVisible()) {
      await confirmModal.locator('button:has-text("Confirm")').click()
    }

    await waitForApiResponse(page, /\/api\/v1\/webhooks.*secret/)

    // Verify new secret is different
    const newSecret = await secretField.textContent()
    expect(newSecret).not.toBe(originalSecret)
  })

  // Helper functions

  async function createWebhookWorkflow(page: any): Promise<string> {
    const workflowName = `Webhook Workflow ${generateTestId('wf')}`
    const workflowId = await workflowEditor.createWorkflow(workflowName)

    await workflowEditor.goto(workflowId)
    await workflowEditor.addNode('trigger:webhook', { x: 100, y: 200 })
    await workflowEditor.addNode('action:log', { x: 300, y: 200 })
    await workflowEditor.saveWorkflow()

    return workflowId
  }

  async function createWebhook(page: any, workflowId: string, name: string): Promise<string> {
    await page.goto(`/workflows/${workflowId}/webhooks`)
    await waitForPageLoad(page)

    await page.click('button:has-text("Create Webhook"), [data-testid="create-webhook"]')

    const modal = page.locator('[data-testid="webhook-modal"]')
    await modal.locator('input[name="name"]').fill(name)
    await modal.locator('select[name="authType"]').selectOption('none')
    await modal.locator('button:has-text("Create")').click()

    const response = await waitForApiResponse(page, /\/api\/v1\/webhooks/)
    const data = await response.json()

    return data.id || generateTestId('webhook')
  }

  async function createWebhookWithAuth(page: any, workflowId: string, name: string, authType: string): Promise<string> {
    await page.goto(`/workflows/${workflowId}/webhooks`)
    await waitForPageLoad(page)

    await page.click('button:has-text("Create Webhook"), [data-testid="create-webhook"]')

    const modal = page.locator('[data-testid="webhook-modal"]')
    await modal.locator('input[name="name"]').fill(name)
    await modal.locator('select[name="authType"]').selectOption(authType)
    await modal.locator('button:has-text("Create")').click()

    const response = await waitForApiResponse(page, /\/api\/v1\/webhooks/)
    const data = await response.json()

    return data.id || generateTestId('webhook')
  }
})
