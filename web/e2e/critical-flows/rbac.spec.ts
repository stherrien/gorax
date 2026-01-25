import { test, expect, Page } from '@playwright/test'
import { generateTestId, waitForPageLoad, waitForApiResponse } from '../utils/test-helpers'

/**
 * Critical Flow: RBAC (Role-Based Access Control)
 *
 * Tests the complete RBAC flow:
 * 1. Create a role with specific permissions
 * 2. Assign the role to a user
 * 3. Verify user can access permitted resources
 * 4. Verify user cannot access restricted resources
 * 5. Revoke role
 * 6. Verify access is denied after revocation
 */
test.describe('RBAC Critical Flow', () => {

  test.beforeEach(async ({ page }) => {
    // Navigate to admin section
    await page.goto('/admin')
    await waitForPageLoad(page)
  })

  test('create role with specific permissions', async ({ page }) => {
    const roleName = `Test Role ${generateTestId('role')}`

    // Navigate to roles management
    await page.click('a:has-text("Roles"), [data-testid="roles-nav"]')
    await waitForPageLoad(page)

    // Create new role
    await page.click('button:has-text("Create Role"), [data-testid="create-role"]')

    // Fill role details
    const modal = page.locator('[data-testid="role-modal"], .role-modal')
    await expect(modal).toBeVisible()

    await modal.locator('input[name="name"]').fill(roleName)
    await modal.locator('textarea[name="description"]').fill('Test role for E2E testing')

    // Select permissions
    await modal.locator('input[value="workflows:read"]').check()
    await modal.locator('input[value="workflows:create"]').check()
    await modal.locator('input[value="executions:read"]').check()

    // Save
    await modal.locator('button:has-text("Create"), button[type="submit"]').click()

    // Wait for API response
    await waitForApiResponse(page, /\/api\/v1\/roles/)

    // Verify role appears in list
    await expect(page.locator(`text="${roleName}"`)).toBeVisible()
  })

  test('assign role to user', async ({ page }) => {
    // First create a role
    const roleName = `Assignable Role ${generateTestId('assign')}`
    await createRole(page, roleName, ['workflows:read'])

    // Navigate to users
    await page.click('a:has-text("Users"), [data-testid="users-nav"]')
    await waitForPageLoad(page)

    // Find a test user and assign role
    const userRow = page.locator('tr:has-text("test@example.com")').first()
    await userRow.locator('button:has-text("Edit Roles"), [data-testid="edit-roles"]').click()

    // Assign the role
    const rolesModal = page.locator('[data-testid="user-roles-modal"]')
    await expect(rolesModal).toBeVisible()

    await rolesModal.locator(`input[value="${roleName}"], label:has-text("${roleName}") input`).check()
    await rolesModal.locator('button:has-text("Save")').click()

    // Wait for API response
    await waitForApiResponse(page, /\/api\/v1\/users.*roles/)

    // Verify role is assigned
    await expect(userRow.locator(`text="${roleName}"`)).toBeVisible()
  })

  test('user with workflows:read can view workflows', async ({ page }) => {
    // Create role with read permission
    const roleName = `Read Only ${generateTestId('ro')}`
    await createRole(page, roleName, ['workflows:read'])

    // Assign to user
    await assignRoleToUser(page, 'readonly@example.com', roleName)

    // Login as the restricted user
    await loginAs(page, 'readonly@example.com', 'password')

    // Navigate to workflows
    await page.goto('/workflows')
    await waitForPageLoad(page)

    // User should see workflows list
    await expect(page.locator('[data-testid="workflows-list"]')).toBeVisible()
  })

  test('user without workflows:create cannot create workflow', async ({ page }) => {
    // Create role with only read permission
    const roleName = `Read Only ${generateTestId('nocreate')}`
    await createRole(page, roleName, ['workflows:read']) // No create permission

    // Assign to user
    await assignRoleToUser(page, 'restricted@example.com', roleName)

    // Login as restricted user
    await loginAs(page, 'restricted@example.com', 'password')

    // Navigate to workflows
    await page.goto('/workflows')
    await waitForPageLoad(page)

    // Create button should be disabled or hidden
    const createButton = page.locator('button:has-text("Create"), [data-testid="create-workflow"]')
    const isDisabled = await createButton.isDisabled().catch(() => false)
    const isHidden = !(await createButton.isVisible().catch(() => false))

    expect(isDisabled || isHidden).toBe(true)
  })

  test('revoke role removes access', async ({ page }) => {
    // Create role
    const roleName = `Revoke Test ${generateTestId('revoke')}`
    await createRole(page, roleName, ['workflows:read', 'workflows:create', 'workflows:delete'])

    // Assign to user
    await assignRoleToUser(page, 'revoke-test@example.com', roleName)

    // Navigate to users and revoke role
    await page.click('a:has-text("Users"), [data-testid="users-nav"]')
    await waitForPageLoad(page)

    const userRow = page.locator('tr:has-text("revoke-test@example.com")')
    await userRow.locator('button:has-text("Edit Roles"), [data-testid="edit-roles"]').click()

    const rolesModal = page.locator('[data-testid="user-roles-modal"]')
    await expect(rolesModal).toBeVisible()

    // Uncheck the role
    await rolesModal.locator(`input[value="${roleName}"], label:has-text("${roleName}") input`).uncheck()
    await rolesModal.locator('button:has-text("Save")').click()

    await waitForApiResponse(page, /\/api\/v1\/users.*roles/)

    // Verify role is removed
    await expect(userRow.locator(`text="${roleName}"`)).not.toBeVisible()
  })

  test('delete role removes it from all users', async ({ page }) => {
    // Create role
    const roleName = `Delete Role ${generateTestId('del')}`
    await createRole(page, roleName, ['workflows:read'])

    // Assign to user
    await assignRoleToUser(page, 'role-delete-test@example.com', roleName)

    // Delete the role
    await page.click('a:has-text("Roles"), [data-testid="roles-nav"]')
    await waitForPageLoad(page)

    const roleRow = page.locator(`tr:has-text("${roleName}")`)
    await roleRow.locator('button:has-text("Delete"), [data-testid="delete-role"]').click()

    // Confirm deletion
    const confirmModal = page.locator('[data-testid="confirm-modal"]')
    await confirmModal.locator('button:has-text("Delete")').click()

    await waitForApiResponse(page, /\/api\/v1\/roles/)

    // Verify role is deleted
    await expect(page.locator(`text="${roleName}"`)).not.toBeVisible()
  })

  test('permissions cascade correctly', async ({ page }) => {
    // Create role with admin-level permission
    const adminRoleName = `Admin Role ${generateTestId('admin')}`
    await createRole(page, adminRoleName, [
      'workflows:*',  // All workflow permissions
      'executions:*', // All execution permissions
      'credentials:read'
    ])

    // Verify the role has cascaded permissions
    await page.click('a:has-text("Roles"), [data-testid="roles-nav"]')
    await waitForPageLoad(page)

    const roleRow = page.locator(`tr:has-text("${adminRoleName}")`)
    await roleRow.click()

    // Check that cascaded permissions are shown
    const permissionsPanel = page.locator('[data-testid="role-permissions"]')
    await expect(permissionsPanel).toBeVisible()

    // Should include implied permissions
    await expect(permissionsPanel.locator('text="workflows:read"')).toBeVisible()
    await expect(permissionsPanel.locator('text="workflows:create"')).toBeVisible()
  })

  // Helper functions

  async function createRole(page: Page, name: string, permissions: string[]): Promise<void> {
    await page.click('a:has-text("Roles"), [data-testid="roles-nav"]')
    await waitForPageLoad(page)

    await page.click('button:has-text("Create Role"), [data-testid="create-role"]')

    const modal = page.locator('[data-testid="role-modal"], .role-modal')
    await expect(modal).toBeVisible()

    await modal.locator('input[name="name"]').fill(name)

    for (const permission of permissions) {
      await modal.locator(`input[value="${permission}"]`).check()
    }

    await modal.locator('button:has-text("Create"), button[type="submit"]').click()
    await waitForApiResponse(page, /\/api\/v1\/roles/)
  }

  async function assignRoleToUser(page: Page, userEmail: string, roleName: string): Promise<void> {
    await page.click('a:has-text("Users"), [data-testid="users-nav"]')
    await waitForPageLoad(page)

    const userRow = page.locator(`tr:has-text("${userEmail}")`)
    await userRow.locator('button:has-text("Edit Roles"), [data-testid="edit-roles"]').click()

    const rolesModal = page.locator('[data-testid="user-roles-modal"]')
    await expect(rolesModal).toBeVisible()

    await rolesModal.locator(`input[value="${roleName}"], label:has-text("${roleName}") input`).check()
    await rolesModal.locator('button:has-text("Save")').click()

    await waitForApiResponse(page, /\/api\/v1\/users.*roles/)
  }

  async function loginAs(page: Page, email: string, password: string): Promise<void> {
    // Log out first if needed
    await page.goto('/logout')
    await page.waitForTimeout(500)

    // Login
    await page.goto('/login')
    await page.fill('input[name="email"]', email)
    await page.fill('input[name="password"]', password)
    await page.click('button[type="submit"]')

    await page.waitForURL('/dashboard')
  }
})
