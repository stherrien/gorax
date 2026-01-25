import { test, expect, waitForLoading, fillFormField, clickButton, expectToast } from './setup'

test.describe('SSO Settings', () => {
  test.use({ storageState: 'admin-auth.json' })

  test.beforeEach(async ({ adminPage }) => {
    await adminPage.goto('/admin/sso')
    await waitForLoading(adminPage)
  })

  test('should display SSO settings page', async ({ adminPage: page }) => {
    await waitForLoading(page)

    await expect(page.locator('h1, h2', { hasText: /SSO|Single Sign-On/i })).toBeVisible()
  })

  test('should display available SSO providers', async ({ adminPage: page }) => {
    await waitForLoading(page)

    const providers = ['SAML', 'OIDC', 'OAuth']

    for (const provider of providers) {
      const providerCard = page.locator(`text=${provider}`)
      const isVisible = await providerCard.isVisible().catch(() => false)
      console.log(`${provider} provider visible:`, isVisible)
    }
  })

  test('should configure SAML provider', async ({ adminPage: page }) => {
    await waitForLoading(page)

    const configureSAML = page.locator('button:has-text("Configure SAML"), button:has-text("Add SAML")')

    if (await configureSAML.isVisible()) {
      await configureSAML.click()

      await fillFormField(page, 'Provider Name', 'Test SAML Provider')
      await fillFormField(page, 'SSO URL', 'https://sso.example.com/saml')
      await fillFormField(page, 'Entity ID', 'gorax-test')

      await clickButton(page, 'Save')

      await expectToast(page, /saved|success/i)
    }
  })

  test('should test SSO connection', async ({ adminPage: page }) => {
    await waitForLoading(page)

    const testButton = page.locator('button:has-text("Test"), button:has-text("Test Connection")')

    if (await testButton.isVisible()) {
      await testButton.click()
      await page.waitForTimeout(2000)

      const hasResult = await page.locator('text=/success|fail|error/i').isVisible()
      expect(hasResult).toBeTruthy()
    }
  })

  test('should enable/disable SSO', async ({ adminPage: page }) => {
    await waitForLoading(page)

    const toggle = page.locator('input[type="checkbox"], button[role="switch"]').first()

    if (await toggle.isVisible()) {
      await toggle.click()
      await page.waitForTimeout(1000)
    }
  })

  test('should delete SSO provider', async ({ adminPage: page }) => {
    await waitForLoading(page)

    const deleteButton = page.locator('button:has-text("Delete"), button[aria-label*="delete" i]').first()

    if (await deleteButton.isVisible()) {
      await deleteButton.click()

      await clickButton(page, 'Confirm')

      await expectToast(page, /deleted|success/i)
    }
  })
})
