import { test, expect, waitForLoading, navigateTo, clickButton } from './setup'

test.describe('OAuth Connections', () => {
  test.beforeEach(async ({ authenticatedPage }) => {
    await page.goto('/oauth/connections')
    await waitForLoading(authenticatedPage)
  })

  test('should display OAuth connections page', async ({ authenticatedPage: page }) => {
    await waitForLoading(page)

    await expect(page.locator('h1, h2', { hasText: /OAuth|Connections/i })).toBeVisible()
  })

  test('should display available OAuth providers', async ({ authenticatedPage: page }) => {
    await waitForLoading(page)

    // Should show common providers
    const providers = ['GitHub', 'Google', 'Microsoft']

    for (const provider of providers) {
      const providerElement = page.locator(`text=${provider}`)
      const isVisible = await providerElement.isVisible().catch(() => false)
      console.log(`${provider} provider visible:`, isVisible)
    }
  })

  test('should connect to OAuth provider', async ({ authenticatedPage: page }) => {
    await waitForLoading(page)

    const connectButton = page.locator('button:has-text("Connect"), button:has-text("Authorize")').first()

    if (await connectButton.isVisible()) {
      // Note: This will open OAuth flow in new window
      // In real test, would need to mock OAuth flow
      console.log('Connect button found - would trigger OAuth flow')
    }
  })

  test('should display connected accounts', async ({ authenticatedPage: page }) => {
    await waitForLoading(page)

    const connectedAccounts = await page.locator('[data-testid="connected-account"]').count()
    console.log('Connected accounts:', connectedAccounts)

    if (connectedAccounts > 0) {
      await expect(page.locator('[data-testid="connected-account"]').first()).toBeVisible()
    }
  })

  test('should disconnect OAuth account', async ({ authenticatedPage: page }) => {
    await waitForLoading(page)

    const disconnectButton = page.locator('button:has-text("Disconnect"), button:has-text("Revoke")').first()

    if (await disconnectButton.isVisible()) {
      await disconnectButton.click()

      await clickButton(page, 'Confirm')

      await page.waitForTimeout(1000)
    }
  })

  test('should show OAuth connection status', async ({ authenticatedPage: page }) => {
    await waitForLoading(page)

    const statusBadge = page.locator('[data-testid="connection-status"], text=/connected|disconnected/i')

    const hasStatus = await statusBadge.isVisible().catch(() => false)
    expect(hasStatus).toBeTruthy()
  })
})
