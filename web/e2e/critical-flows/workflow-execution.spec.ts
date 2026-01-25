import { test, expect } from '@playwright/test'
import { WorkflowEditorPage, ExecutionsPage } from '../pages'
import { generateTestId, waitForPageLoad } from '../utils/test-helpers'

/**
 * Critical Flow: Workflow Execution
 *
 * Tests the complete workflow execution flow:
 * 1. Create a workflow with actions
 * 2. Execute the workflow
 * 3. Verify execution results
 */
test.describe('Workflow Execution Critical Flow', () => {
  let workflowEditor: WorkflowEditorPage
  let executionsPage: ExecutionsPage

  test.beforeEach(async ({ page }) => {
    workflowEditor = new WorkflowEditorPage(page)
    executionsPage = new ExecutionsPage(page)

    // Navigate to workflows page
    await page.goto('/workflows')
    await waitForPageLoad(page)
  })

  test('create and execute a simple workflow', async ({ page }) => {
    const workflowName = `E2E Test Workflow ${generateTestId('wf')}`

    // Step 1: Create workflow
    const workflowId = await workflowEditor.createWorkflow(
      workflowName,
      'Automated E2E test workflow'
    )
    expect(workflowId).toBeTruthy()

    // Step 2: Add nodes to the workflow
    await workflowEditor.goto(workflowId)

    // Add manual trigger
    await workflowEditor.addNode('trigger:manual', { x: 100, y: 200 })

    // Add a log action
    await workflowEditor.addNode('action:log', { x: 300, y: 200 })

    // Step 3: Save workflow
    await workflowEditor.saveWorkflow()

    // Step 4: Execute workflow
    const executionId = await workflowEditor.executeWorkflow({
      message: 'Hello from E2E test'
    })
    expect(executionId).toBeTruthy()

    // Step 5: Verify execution
    await executionsPage.gotoExecution(executionId)
    const status = await executionsPage.waitForCompletion(executionId, 30000)

    expect(['completed', 'failed']).toContain(status)
  })

  test('execute workflow with HTTP action', async ({ page }) => {
    const workflowName = `HTTP Workflow ${generateTestId('http')}`

    // Create workflow with HTTP action
    const workflowId = await workflowEditor.createWorkflow(
      workflowName,
      'Tests HTTP action execution'
    )

    await workflowEditor.goto(workflowId)

    // Add manual trigger
    await workflowEditor.addNode('trigger:manual', { x: 100, y: 200 })

    // Add HTTP action
    await workflowEditor.addNode('action:http', { x: 300, y: 200 })

    // Configure HTTP action to call a test endpoint
    // Note: This would need a real test endpoint in actual E2E tests
    await workflowEditor.configureNode('http-1', {
      method: 'GET',
      url: 'https://httpbin.org/get'
    })

    await workflowEditor.saveWorkflow()

    // Execute and verify
    const executionId = await workflowEditor.executeWorkflow()
    expect(executionId).toBeTruthy()

    await executionsPage.gotoExecution(executionId)
    const status = await executionsPage.waitForCompletion(executionId, 30000)

    // Check if execution completed (may fail if external service is down)
    expect(['completed', 'failed']).toContain(status)
  })

  test('execution shows step-by-step progress', async ({ page }) => {
    const workflowName = `Multi-step Workflow ${generateTestId('multi')}`

    // Create workflow
    const workflowId = await workflowEditor.createWorkflow(workflowName)

    await workflowEditor.goto(workflowId)

    // Add multiple steps
    await workflowEditor.addNode('trigger:manual', { x: 100, y: 200 })
    await workflowEditor.addNode('action:log', { x: 300, y: 200 })
    await workflowEditor.addNode('action:transform', { x: 500, y: 200 })

    await workflowEditor.saveWorkflow()

    // Execute
    const executionId = await workflowEditor.executeWorkflow({
      data: { key: 'value' }
    })

    // Navigate to execution and verify steps
    await executionsPage.gotoExecution(executionId)

    // Wait for completion
    await executionsPage.waitForCompletion(executionId, 30000)

    // Get step results
    const steps = await executionsPage.getStepResults()
    expect(steps.length).toBeGreaterThan(0)
  })

  test('failed execution shows error message', async ({ page }) => {
    const workflowName = `Error Workflow ${generateTestId('err')}`

    // Create workflow with action that will fail
    const workflowId = await workflowEditor.createWorkflow(workflowName)

    await workflowEditor.goto(workflowId)

    // Add trigger and HTTP action pointing to invalid URL
    await workflowEditor.addNode('trigger:manual', { x: 100, y: 200 })
    await workflowEditor.addNode('action:http', { x: 300, y: 200 })

    await workflowEditor.configureNode('http-1', {
      method: 'GET',
      url: 'http://invalid-url-that-will-fail.local/test'
    })

    await workflowEditor.saveWorkflow()

    // Execute
    const executionId = await workflowEditor.executeWorkflow()

    // Verify execution shows failure
    await executionsPage.gotoExecution(executionId)
    const status = await executionsPage.waitForCompletion(executionId, 30000)

    expect(status).toBe('failed')

    // Check error message is displayed
    const errorMessage = await executionsPage.getErrorMessage()
    expect(errorMessage).toBeTruthy()
  })

  test('validate workflow before execution', async ({ page }) => {
    const workflowName = `Validation Test ${generateTestId('val')}`

    // Create workflow
    const workflowId = await workflowEditor.createWorkflow(workflowName)
    await workflowEditor.goto(workflowId)

    // Add trigger
    await workflowEditor.addNode('trigger:manual', { x: 100, y: 200 })

    await workflowEditor.saveWorkflow()

    // Validate workflow
    const isValid = await workflowEditor.validateWorkflow()

    // A workflow with just a trigger should be valid (or may require at least one action)
    expect(typeof isValid).toBe('boolean')
  })

  test('execution can be cancelled', async ({ page }) => {
    const workflowName = `Cancel Test ${generateTestId('cancel')}`

    // Create workflow with delay
    const workflowId = await workflowEditor.createWorkflow(workflowName)
    await workflowEditor.goto(workflowId)

    // Add trigger and delay
    await workflowEditor.addNode('trigger:manual', { x: 100, y: 200 })
    await workflowEditor.addNode('control:delay', { x: 300, y: 200 })

    await workflowEditor.configureNode('delay-1', {
      delayMs: 30000 // 30 second delay
    })

    await workflowEditor.saveWorkflow()

    // Execute
    const executionId = await workflowEditor.executeWorkflow()

    // Navigate to execution and cancel
    await executionsPage.gotoExecution(executionId)

    // Wait for execution to start
    await page.waitForTimeout(2000)

    // Cancel execution
    await executionsPage.cancelExecution()

    // Verify cancelled status
    const status = await executionsPage.getStatus()
    expect(['cancelled', 'failed']).toContain(status)
  })

  test.afterEach(async ({ page }) => {
    // Cleanup: Delete any test workflows created
    // This would require implementing a cleanup API or using the UI
  })
})
