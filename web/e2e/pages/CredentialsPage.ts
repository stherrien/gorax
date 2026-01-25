import { Page, Locator, expect } from '@playwright/test'
import { waitForPageLoad, waitForApiResponse, generateTestId } from '../utils/test-helpers'

/**
 * Page Object Model for the Credentials Management page
 * Provides methods for creating, rotating, and managing credentials
 */
export class CredentialsPage {
  readonly page: Page
  readonly credentialsList: Locator
  readonly createButton: Locator
  readonly searchInput: Locator
  readonly typeFilter: Locator

  constructor(page: Page) {
    this.page = page
    this.credentialsList = page.locator('[data-testid="credentials-list"], .credentials-list')
    this.createButton = page.locator('button:has-text("Create"), [data-testid="create-credential"]')
    this.searchInput = page.locator('input[type="search"], [data-testid="search-credentials"]')
    this.typeFilter = page.locator('[data-testid="type-filter"], select[name="type"]')
  }

  /**
   * Navigate to the credentials page
   */
  async goto(): Promise<void> {
    await this.page.goto('/credentials')
    await waitForPageLoad(this.page)
  }

  /**
   * Create a new credential
   */
  async createCredential(
    name: string,
    type: 'api_key' | 'bearer_token' | 'basic_auth' | 'oauth2' | 'custom',
    values: Record<string, string>,
    description?: string
  ): Promise<string> {
    await this.createButton.click()

    // Wait for modal to open
    const modal = this.page.locator('[data-testid="credential-modal"], .credential-modal')
    await expect(modal).toBeVisible()

    // Fill name
    await modal.locator('input[name="name"], [data-testid="credential-name"]').fill(name)

    // Select type
    await modal.locator('select[name="type"], [data-testid="credential-type"]').selectOption(type)

    // Fill description if provided
    if (description) {
      await modal.locator('textarea[name="description"], [data-testid="credential-description"]').fill(description)
    }

    // Fill credential values based on type
    for (const [key, value] of Object.entries(values)) {
      const input = modal.locator(`input[name="${key}"], [data-field="${key}"]`)
      await input.fill(value)
    }

    // Submit
    await modal.locator('button:has-text("Create"), button[type="submit"]').click()

    // Wait for API response
    const response = await waitForApiResponse(this.page, /\/api\/v1\/credentials/)
    const data = await response.json()

    // Wait for modal to close
    await expect(modal).not.toBeVisible()

    return data.id || generateTestId('cred')
  }

  /**
   * Get a credential row by ID
   */
  getCredentialRow(credentialId: string): Locator {
    return this.credentialsList.locator(`tr:has([data-credential-id="${credentialId}"]), [data-credential-id="${credentialId}"]`)
  }

  /**
   * Rotate a credential with a new value
   */
  async rotateCredential(credentialId: string, newValue: string): Promise<void> {
    const row = this.getCredentialRow(credentialId)

    // Click rotate button
    const rotateButton = row.locator('button:has-text("Rotate"), [data-testid="rotate-credential"]')
    await rotateButton.click()

    // Wait for rotation modal
    const modal = this.page.locator('[data-testid="rotate-modal"], .rotation-modal')
    await expect(modal).toBeVisible()

    // Enter new value
    const newValueInput = modal.locator('input[name="newValue"], [data-testid="new-value"]')
    await newValueInput.fill(newValue)

    // Confirm rotation
    await modal.locator('button:has-text("Rotate"), button:has-text("Confirm")').click()

    // Wait for API response
    await waitForApiResponse(this.page, /\/api\/v1\/credentials\/.*\/rotate/)

    // Verify success message
    await expect(this.page.locator('text=/rotated|updated/i')).toBeVisible({ timeout: 5000 })
  }

  /**
   * Delete a credential
   */
  async deleteCredential(credentialId: string): Promise<void> {
    const row = this.getCredentialRow(credentialId)

    // Click delete button
    const deleteButton = row.locator('button:has-text("Delete"), [data-testid="delete-credential"]')
    await deleteButton.click()

    // Confirm deletion
    const confirmModal = this.page.locator('[data-testid="confirm-modal"], .confirm-dialog')
    await expect(confirmModal).toBeVisible()

    await confirmModal.locator('button:has-text("Delete"), button:has-text("Confirm")').click()

    // Wait for API response
    await waitForApiResponse(this.page, /\/api\/v1\/credentials/)

    // Verify row is removed
    await expect(row).not.toBeVisible()
  }

  /**
   * Test a credential's validity
   */
  async testCredential(credentialId: string): Promise<boolean> {
    const row = this.getCredentialRow(credentialId)

    // Click test button
    const testButton = row.locator('button:has-text("Test"), [data-testid="test-credential"]')
    await testButton.click()

    // Wait for test response
    await waitForApiResponse(this.page, /\/api\/v1\/credentials\/.*\/test/)

    // Check result
    const successIndicator = row.locator('.test-success, [data-testid="test-success"]')
    const failureIndicator = row.locator('.test-failure, [data-testid="test-failure"]')

    if (await successIndicator.isVisible()) {
      return true
    }

    return !(await failureIndicator.isVisible())
  }

  /**
   * Search for credentials
   */
  async search(query: string): Promise<void> {
    await this.searchInput.fill(query)
    await this.page.waitForTimeout(500) // Debounce
    await waitForApiResponse(this.page, /\/api\/v1\/credentials/)
  }

  /**
   * Filter credentials by type
   */
  async filterByType(type: string): Promise<void> {
    await this.typeFilter.selectOption(type)
    await waitForApiResponse(this.page, /\/api\/v1\/credentials/)
  }

  /**
   * Get list of all visible credentials
   */
  async getCredentialsList(): Promise<{ id: string; name: string; type: string }[]> {
    const credentials: { id: string; name: string; type: string }[] = []
    const rows = this.credentialsList.locator('tr, .credential-row').filter({ has: this.page.locator('[data-credential-id]') })
    const count = await rows.count()

    for (let i = 0; i < count; i++) {
      const row = rows.nth(i)
      const id = await row.getAttribute('data-credential-id') ||
                 await row.locator('[data-testid="credential-id"]').textContent() || ''
      const name = await row.locator('.credential-name, [data-testid="credential-name"]').textContent() || ''
      const type = await row.locator('.credential-type, [data-testid="credential-type"]').textContent() || ''

      credentials.push({
        id: id.trim(),
        name: name.trim(),
        type: type.trim()
      })
    }

    return credentials
  }

  /**
   * View credential details
   */
  async viewCredentialDetails(credentialId: string): Promise<void> {
    const row = this.getCredentialRow(credentialId)
    await row.click()

    // Wait for detail panel or modal
    const detailPanel = this.page.locator('[data-testid="credential-details"], .credential-details')
    await expect(detailPanel).toBeVisible()
  }

  /**
   * Edit credential metadata (name, description)
   */
  async editCredential(credentialId: string, updates: { name?: string; description?: string }): Promise<void> {
    const row = this.getCredentialRow(credentialId)

    // Click edit button
    const editButton = row.locator('button:has-text("Edit"), [data-testid="edit-credential"]')
    await editButton.click()

    // Wait for edit modal
    const modal = this.page.locator('[data-testid="edit-modal"], .edit-modal')
    await expect(modal).toBeVisible()

    // Update fields
    if (updates.name) {
      await modal.locator('input[name="name"]').fill(updates.name)
    }
    if (updates.description) {
      await modal.locator('textarea[name="description"]').fill(updates.description)
    }

    // Save changes
    await modal.locator('button:has-text("Save"), button[type="submit"]').click()

    // Wait for API response
    await waitForApiResponse(this.page, /\/api\/v1\/credentials/)
  }

  /**
   * Check if a credential exists
   */
  async credentialExists(name: string): Promise<boolean> {
    await this.search(name)
    const row = this.credentialsList.locator(`tr:has-text("${name}")`)
    return await row.isVisible()
  }

  /**
   * Get credential usage information (which workflows use it)
   */
  async getCredentialUsage(credentialId: string): Promise<string[]> {
    await this.viewCredentialDetails(credentialId)

    const usageSection = this.page.locator('[data-testid="credential-usage"], .usage-section')
    if (!(await usageSection.isVisible())) {
      return []
    }

    const workflowLinks = usageSection.locator('a, .workflow-link')
    const count = await workflowLinks.count()
    const workflows: string[] = []

    for (let i = 0; i < count; i++) {
      const text = await workflowLinks.nth(i).textContent()
      if (text) workflows.push(text.trim())
    }

    return workflows
  }

  /**
   * View credential version history
   */
  async viewVersionHistory(credentialId: string): Promise<{ version: number; createdAt: string }[]> {
    await this.viewCredentialDetails(credentialId)

    const historyTab = this.page.locator('button:has-text("History"), [data-testid="version-history-tab"]')
    await historyTab.click()

    const versions: { version: number; createdAt: string }[] = []
    const versionRows = this.page.locator('[data-testid="version-row"], .version-row')
    const count = await versionRows.count()

    for (let i = 0; i < count; i++) {
      const row = versionRows.nth(i)
      const versionText = await row.locator('.version-number').textContent() || '0'
      const dateText = await row.locator('.version-date').textContent() || ''

      versions.push({
        version: parseInt(versionText),
        createdAt: dateText.trim()
      })
    }

    return versions
  }
}
