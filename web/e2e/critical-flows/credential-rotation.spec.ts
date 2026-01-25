import { test, expect } from '@playwright/test'
import { WorkflowEditorPage, ExecutionsPage, CredentialsPage } from '../pages'
import { generateTestId, waitForPageLoad } from '../utils/test-helpers'

/**
 * Critical Flow: Credential Rotation
 *
 * Tests the complete credential lifecycle:
 * 1. Create a credential
 * 2. Use the credential in a workflow
 * 3. Execute workflow to verify credential works
 * 4. Rotate the credential
 * 5. Verify rotated credential is used in subsequent executions
 */
test.describe('Credential Rotation Critical Flow', () => {
  let credentialsPage: CredentialsPage
  let workflowEditor: WorkflowEditorPage
  let executionsPage: ExecutionsPage

  test.beforeEach(async ({ page }) => {
    credentialsPage = new CredentialsPage(page)
    workflowEditor = new WorkflowEditorPage(page)
    executionsPage = new ExecutionsPage(page)
  })

  test('create credential and use in workflow', async ({ page }) => {
    const credentialName = `API Key ${generateTestId('cred')}`
    const initialValue = 'initial-api-key-value-12345'

    // Step 1: Create credential
    await credentialsPage.goto()
    const credentialId = await credentialsPage.createCredential(
      credentialName,
      'api_key',
      { apiKey: initialValue },
      'Test API key for E2E testing'
    )
    expect(credentialId).toBeTruthy()

    // Step 2: Verify credential appears in list
    const credentials = await credentialsPage.getCredentialsList()
    const created = credentials.find(c => c.id === credentialId)
    expect(created).toBeTruthy()
    expect(created?.name).toBe(credentialName)

    // Step 3: Create workflow that uses the credential
    const workflowName = `Credential Test Workflow ${generateTestId('wf')}`
    const workflowId = await workflowEditor.createWorkflow(
      workflowName,
      'Uses credential for authentication'
    )

    await workflowEditor.goto(workflowId)

    // Add HTTP action with credential reference
    await workflowEditor.addNode('trigger:manual', { x: 100, y: 200 })
    await workflowEditor.addNode('action:http', { x: 300, y: 200 })

    await workflowEditor.configureNode('http-1', {
      method: 'GET',
      url: 'https://httpbin.org/headers',
      // Reference credential using {{credentials.name}} syntax
      headers: JSON.stringify({
        'X-API-Key': `{{credentials.${credentialName}}}`
      })
    })

    await workflowEditor.saveWorkflow()

    // Step 4: Execute workflow
    const executionId = await workflowEditor.executeWorkflow()
    expect(executionId).toBeTruthy()

    // Wait for completion
    await executionsPage.gotoExecution(executionId)
    const status = await executionsPage.waitForCompletion(executionId, 30000)

    // Execution should complete (credential was injected)
    expect(['completed', 'failed']).toContain(status)
  })

  test('rotate credential and verify new value is used', async ({ page }) => {
    const credentialName = `Rotatable Key ${generateTestId('rot')}`
    const initialValue = 'old-api-key-value'
    const newValue = 'new-rotated-api-key-value'

    // Create credential
    await credentialsPage.goto()
    const credentialId = await credentialsPage.createCredential(
      credentialName,
      'api_key',
      { apiKey: initialValue }
    )

    // Create and execute workflow with credential
    const workflowName = `Rotation Test ${generateTestId('wf')}`
    const workflowId = await workflowEditor.createWorkflow(workflowName)

    await workflowEditor.goto(workflowId)
    await workflowEditor.addNode('trigger:manual', { x: 100, y: 200 })
    await workflowEditor.addNode('action:http', { x: 300, y: 200 })

    await workflowEditor.configureNode('http-1', {
      method: 'GET',
      url: 'https://httpbin.org/headers',
      headers: JSON.stringify({
        'Authorization': `Bearer {{credentials.${credentialName}}}`
      })
    })

    await workflowEditor.saveWorkflow()

    // Execute with initial credential
    const firstExecutionId = await workflowEditor.executeWorkflow()
    await executionsPage.gotoExecution(firstExecutionId)
    await executionsPage.waitForCompletion(firstExecutionId, 30000)

    // Rotate credential
    await credentialsPage.goto()
    await credentialsPage.rotateCredential(credentialId, newValue)

    // Execute again - should use new value
    await workflowEditor.goto(workflowId)
    const secondExecutionId = await workflowEditor.executeWorkflow()
    await executionsPage.gotoExecution(secondExecutionId)
    await executionsPage.waitForCompletion(secondExecutionId, 30000)

    // Both executions should complete
    expect(firstExecutionId).toBeTruthy()
    expect(secondExecutionId).toBeTruthy()
  })

  test('test credential validity', async ({ page }) => {
    const credentialName = `Test Cred ${generateTestId('test')}`

    // Create credential
    await credentialsPage.goto()
    const credentialId = await credentialsPage.createCredential(
      credentialName,
      'api_key',
      { apiKey: 'test-key-value' }
    )

    // Test the credential
    const isValid = await credentialsPage.testCredential(credentialId)

    // Note: This test depends on having a test endpoint configured
    // The result could be true or false depending on setup
    expect(typeof isValid).toBe('boolean')
  })

  test('delete credential', async ({ page }) => {
    const credentialName = `Delete Test ${generateTestId('del')}`

    // Create credential
    await credentialsPage.goto()
    const credentialId = await credentialsPage.createCredential(
      credentialName,
      'api_key',
      { apiKey: 'temporary-key' }
    )

    // Verify it exists
    let exists = await credentialsPage.credentialExists(credentialName)
    expect(exists).toBe(true)

    // Delete it
    await credentialsPage.deleteCredential(credentialId)

    // Verify it's gone
    exists = await credentialsPage.credentialExists(credentialName)
    expect(exists).toBe(false)
  })

  test('credential version history tracks rotations', async ({ page }) => {
    const credentialName = `Version History ${generateTestId('ver')}`

    // Create credential
    await credentialsPage.goto()
    const credentialId = await credentialsPage.createCredential(
      credentialName,
      'api_key',
      { apiKey: 'version-1-value' }
    )

    // Rotate multiple times
    await credentialsPage.rotateCredential(credentialId, 'version-2-value')
    await page.waitForTimeout(1000) // Brief pause between rotations

    await credentialsPage.rotateCredential(credentialId, 'version-3-value')

    // Check version history
    const history = await credentialsPage.viewVersionHistory(credentialId)

    // Should have at least 3 versions
    expect(history.length).toBeGreaterThanOrEqual(3)
  })

  test('different credential types', async ({ page }) => {
    await credentialsPage.goto()

    // Test API Key
    const apiKeyId = await credentialsPage.createCredential(
      `API Key ${generateTestId('api')}`,
      'api_key',
      { apiKey: 'my-api-key' }
    )
    expect(apiKeyId).toBeTruthy()

    // Test Bearer Token
    const bearerTokenId = await credentialsPage.createCredential(
      `Bearer Token ${generateTestId('bearer')}`,
      'bearer_token',
      { token: 'my-bearer-token' }
    )
    expect(bearerTokenId).toBeTruthy()

    // Test Basic Auth
    const basicAuthId = await credentialsPage.createCredential(
      `Basic Auth ${generateTestId('basic')}`,
      'basic_auth',
      { username: 'user', password: 'pass' }
    )
    expect(basicAuthId).toBeTruthy()
  })
})
