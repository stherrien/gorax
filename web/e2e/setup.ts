import { test as base, expect, Page } from '@playwright/test';

// Extend base test with custom fixtures
export const test = base.extend<{
  authenticatedPage: Page;
  adminPage: Page;
}>({
  // Authenticated user page fixture
  authenticatedPage: async ({ page }, use) => {
    // Navigate to login
    await page.goto('/login');

    // Fill in test credentials
    await page.fill('input[name="email"]', 'test@example.com');
    await page.fill('input[name="password"]', 'testpassword');

    // Submit login form
    await page.click('button[type="submit"]');

    // Wait for redirect to dashboard
    await page.waitForURL('/dashboard');

    // Verify logged in
    await expect(page.locator('[data-testid="user-menu"]')).toBeVisible();

    // Use the authenticated page
    await use(page);
  },

  // Admin user page fixture
  adminPage: async ({ page }, use) => {
    // Navigate to login
    await page.goto('/login');

    // Fill in admin credentials
    await page.fill('input[name="email"]', 'admin@example.com');
    await page.fill('input[name="password"]', 'adminpassword');

    // Submit login form
    await page.click('button[type="submit"]');

    // Wait for redirect
    await page.waitForURL('/dashboard');

    // Verify admin access
    await expect(page.locator('[data-testid="admin-menu"]')).toBeVisible();

    // Use the admin page
    await use(page);
  },
});

// Custom assertions and helpers
export { expect };

// Helper function to wait for API call
export async function waitForAPICall(page: Page, url: string): Promise<void> {
  await page.waitForResponse((response) =>
    response.url().includes(url) && response.status() === 200
  );
}

// Helper function to fill form field
export async function fillFormField(
  page: Page,
  label: string,
  value: string
): Promise<void> {
  const input = page.locator(`label:has-text("${label}") + input, input[placeholder*="${label}" i]`);
  await input.fill(value);
}

// Helper function to select from dropdown
export async function selectDropdownOption(
  page: Page,
  label: string,
  option: string
): Promise<void> {
  const select = page.locator(`label:has-text("${label}") + select, select[aria-label="${label}"]`);
  await select.selectOption(option);
}

// Helper function to click button by text
export async function clickButton(page: Page, text: string): Promise<void> {
  const button = page.locator(`button:has-text("${text}")`);
  await button.click();
}

// Helper function to verify toast message
export async function expectToast(page: Page, message: string): Promise<void> {
  const toast = page.locator('[role="alert"], [data-testid="toast"]', { hasText: message });
  await expect(toast).toBeVisible({ timeout: 5000 });
}

// Helper function to verify error message
export async function expectError(page: Page, message: string): Promise<void> {
  const error = page.locator('[role="alert"], .error-message', { hasText: message });
  await expect(error).toBeVisible();
}

// Helper function to wait for loading to complete
export async function waitForLoading(page: Page): Promise<void> {
  await page.waitForLoadState('networkidle');
  const loader = page.locator('[data-testid="loading"], .loading, .spinner');
  await expect(loader).toBeHidden({ timeout: 10000 });
}

// Helper to create a test workflow
export async function createTestWorkflow(
  page: Page,
  name: string,
  description?: string
): Promise<string> {
  // Navigate to workflows
  await page.goto('/workflows');

  // Click create button
  await clickButton(page, 'Create Workflow');

  // Fill form
  await fillFormField(page, 'Name', name);
  if (description) {
    await fillFormField(page, 'Description', description);
  }

  // Submit
  await clickButton(page, 'Create');

  // Wait for redirect and get workflow ID from URL
  await page.waitForURL(/\/workflows\/[a-f0-9-]+/);
  const url = page.url();
  const workflowId = url.split('/').pop() || '';

  return workflowId;
}

// Helper to navigate to specific section
export async function navigateTo(page: Page, section: string): Promise<void> {
  const navLink = page.locator(`nav a:has-text("${section}")`);
  await navLink.click();
  await waitForLoading(page);
}

// Helper to verify table row
export async function expectTableRow(
  page: Page,
  rowText: string
): Promise<void> {
  const row = page.locator(`tr:has-text("${rowText}")`);
  await expect(row).toBeVisible();
}

// Helper to search in table/list
export async function searchFor(page: Page, query: string): Promise<void> {
  const searchInput = page.locator('input[type="search"], input[placeholder*="Search" i]');
  await searchInput.fill(query);
  await page.waitForTimeout(500); // Debounce delay
  await waitForLoading(page);
}

// Helper to verify empty state
export async function expectEmptyState(page: Page, message?: string): Promise<void> {
  const emptyState = page.locator('[data-testid="empty-state"], .empty-state');
  await expect(emptyState).toBeVisible();

  if (message) {
    await expect(emptyState).toContainText(message);
  }
}

// Mock API response helper
export async function mockAPIResponse(
  page: Page,
  url: string,
  response: any,
  status: number = 200
): Promise<void> {
  await page.route(`**/${url}`, (route) => {
    route.fulfill({
      status,
      contentType: 'application/json',
      body: JSON.stringify(response),
    });
  });
}

// Helper to take screenshot on failure
export async function screenshotOnFailure(
  page: Page,
  testInfo: any
): Promise<void> {
  if (testInfo.status !== 'passed') {
    const screenshot = await page.screenshot();
    await testInfo.attach('screenshot', {
      body: screenshot,
      contentType: 'image/png',
    });
  }
}
