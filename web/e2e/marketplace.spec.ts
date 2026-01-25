import { test, expect, waitForAPICall, fillFormField, clickButton, expectToast, waitForLoading, navigateTo, searchFor, expectTableRow } from './setup';

test.describe('Marketplace', () => {
  test.beforeEach(async ({ authenticatedPage }) => {
    // Navigate to marketplace before each test
    await navigateTo(authenticatedPage, 'Marketplace');
  });

  test('should display marketplace templates', async ({ authenticatedPage: page }) => {
    // Wait for templates to load
    await waitForLoading(page);

    // Verify marketplace header
    await expect(page.locator('h1, h2', { hasText: 'Marketplace' })).toBeVisible();

    // Verify template cards are displayed
    const templateCards = page.locator('[data-testid="template-card"]');
    await expect(templateCards).not.toHaveCount(0);

    // Verify each card has required elements
    const firstCard = templateCards.first();
    await expect(firstCard.locator('[data-testid="template-name"]')).toBeVisible();
    await expect(firstCard.locator('[data-testid="template-category"]')).toBeVisible();
    await expect(firstCard.locator('[data-testid="template-rating"]')).toBeVisible();
  });

  test('should filter templates by category', async ({ authenticatedPage: page }) => {
    // Wait for page load
    await waitForLoading(page);

    // Click on category filter
    await page.click('[data-testid="category-filter-automation"]');

    // Wait for filtered results
    await waitForLoading(page);

    // Verify URL includes category parameter
    expect(page.url()).toContain('category=automation');

    // Verify only automation templates are shown
    const templateCards = page.locator('[data-testid="template-card"]');
    const count = await templateCards.count();

    for (let i = 0; i < count; i++) {
      const category = await templateCards.nth(i)
        .locator('[data-testid="template-category"]')
        .textContent();
      expect(category?.toLowerCase()).toContain('automation');
    }
  });

  test('should search for templates', async ({ authenticatedPage: page }) => {
    // Wait for page load
    await waitForLoading(page);

    // Search for "webhook"
    await searchFor(page, 'webhook');

    // Verify results contain search term
    const templateCards = page.locator('[data-testid="template-card"]');
    await expect(templateCards.first()).toBeVisible();

    const firstTemplate = await templateCards.first()
      .locator('[data-testid="template-name"]')
      .textContent();
    expect(firstTemplate?.toLowerCase()).toContain('webhook');
  });

  test('should view template details', async ({ authenticatedPage: page }) => {
    // Wait for templates to load
    await waitForLoading(page);

    // Click on first template
    const firstTemplate = page.locator('[data-testid="template-card"]').first();
    const templateName = await firstTemplate
      .locator('[data-testid="template-name"]')
      .textContent();

    await firstTemplate.click();

    // Wait for details page
    await page.waitForURL(/\/marketplace\/templates\/[a-f0-9-]+/);

    // Verify template details are displayed
    await expect(page.locator('h1, h2', { hasText: templateName || '' })).toBeVisible();
    await expect(page.locator('[data-testid="template-description"]')).toBeVisible();
    await expect(page.locator('[data-testid="template-definition"]')).toBeVisible();

    // Verify action buttons
    await expect(page.locator('button:has-text("Install")')).toBeVisible();
    await expect(page.locator('button:has-text("Rate")')).toBeVisible();
  });

  test('should install template', async ({ authenticatedPage: page }) => {
    // Navigate to template details
    await waitForLoading(page);
    await page.locator('[data-testid="template-card"]').first().click();
    await page.waitForURL(/\/marketplace\/templates\/[a-f0-9-]+/);

    // Click install button
    await clickButton(page, 'Install');

    // Fill workflow name in modal/form
    await fillFormField(page, 'Workflow Name', 'My Installed Workflow');

    // Submit installation
    await clickButton(page, 'Install Template');

    // Wait for success message
    await expectToast(page, 'Template installed successfully');

    // Verify redirect to workflows page
    await page.waitForURL('/workflows');

    // Verify new workflow appears in list
    await expectTableRow(page, 'My Installed Workflow');
  });

  test('should rate template', async ({ authenticatedPage: page }) => {
    // Navigate to template details
    await waitForLoading(page);
    await page.locator('[data-testid="template-card"]').first().click();
    await page.waitForURL(/\/marketplace\/templates\/[a-f0-9-]+/);

    // Click rate button
    await clickButton(page, 'Rate');

    // Select 5 stars
    await page.click('[data-testid="star-5"]');

    // Enter review comment
    await fillFormField(page, 'Comment', 'Excellent template! Works perfectly.');

    // Submit review
    await clickButton(page, 'Submit Review');

    // Wait for success
    await expectToast(page, 'Review submitted successfully');

    // Verify review appears in reviews section
    await expect(page.locator('[data-testid="review-list"]')).toContainText('Excellent template!');

    // Verify rating updated
    const ratingElement = page.locator('[data-testid="template-rating"]');
    await expect(ratingElement).toContainText('5.0');
  });

  test('should view template reviews', async ({ authenticatedPage: page }) => {
    // Navigate to template with reviews
    await waitForLoading(page);
    await page.locator('[data-testid="template-card"]').first().click();
    await page.waitForURL(/\/marketplace\/templates\/[a-f0-9-]+/);

    // Scroll to reviews section
    await page.locator('[data-testid="reviews-section"]').scrollIntoViewIfNeeded();

    // Verify reviews are displayed
    const reviews = page.locator('[data-testid="review-item"]');
    const reviewCount = await reviews.count();

    if (reviewCount > 0) {
      // Verify review components
      const firstReview = reviews.first();
      await expect(firstReview.locator('[data-testid="reviewer-name"]')).toBeVisible();
      await expect(firstReview.locator('[data-testid="review-rating"]')).toBeVisible();
      await expect(firstReview.locator('[data-testid="review-comment"]')).toBeVisible();
    }
  });

  test('should display trending templates', async ({ authenticatedPage: page }) => {
    // Click on trending tab/section
    await page.click('[data-testid="trending-tab"]');

    // Wait for trending templates to load
    await waitForLoading(page);

    // Verify trending templates are displayed
    const trendingCards = page.locator('[data-testid="template-card"]');
    await expect(trendingCards.first()).toBeVisible();

    // Verify sorting (most recent downloads should be first)
    const firstCard = trendingCards.first();
    const downloadCount = await firstCard
      .locator('[data-testid="download-count"]')
      .textContent();

    expect(downloadCount).toBeTruthy();
  });

  test('should display popular templates', async ({ authenticatedPage: page }) => {
    // Click on popular tab/section
    await page.click('[data-testid="popular-tab"]');

    // Wait for popular templates to load
    await waitForLoading(page);

    // Verify popular templates are displayed
    const popularCards = page.locator('[data-testid="template-card"]');
    await expect(popularCards.first()).toBeVisible();

    // Verify templates have download counts
    const downloadCounts = await popularCards
      .locator('[data-testid="download-count"]')
      .allTextContents();

    expect(downloadCounts.length).toBeGreaterThan(0);
  });

  test('should handle empty search results', async ({ authenticatedPage: page }) => {
    // Wait for page load
    await waitForLoading(page);

    // Search for non-existent template
    await searchFor(page, 'xyznonexistenttemplate123');

    // Verify empty state
    const emptyState = page.locator('[data-testid="empty-state"]');
    await expect(emptyState).toBeVisible();
    await expect(emptyState).toContainText('No templates found');
  });

  test('should prevent duplicate installation', async ({ authenticatedPage: page }) => {
    // Navigate to template details
    await waitForLoading(page);
    await page.locator('[data-testid="template-card"]').first().click();
    await page.waitForURL(/\/marketplace\/templates\/[a-f0-9-]+/);

    // Get template ID from URL
    const templateId = page.url().split('/').pop();

    // Check if already installed (button should be disabled or show "Installed")
    const installButton = page.locator('button:has-text("Install"), button:has-text("Installed")');

    const buttonText = await installButton.textContent();
    if (buttonText?.includes('Installed')) {
      // Verify button is disabled
      await expect(installButton).toBeDisabled();

      // Verify "Already installed" message
      await expect(page.locator('text=Already installed')).toBeVisible();
    }
  });
});

test.describe('Marketplace - Publish Template', () => {
  test('should publish new template', async ({ authenticatedPage: page }) => {
    // Navigate to marketplace
    await navigateTo(page, 'Marketplace');

    // Click "Publish Template" button
    await clickButton(page, 'Publish Template');

    // Fill in template details
    await fillFormField(page, 'Name', 'My Custom Template');
    await fillFormField(page, 'Description', 'This is a custom template for testing');

    // Select category
    await page.selectOption('[name="category"]', 'automation');

    // Add tags
    await fillFormField(page, 'Tags', 'test, automation, custom');

    // Select workflow to use as template
    await page.selectOption('[name="workflow"]', { index: 0 });

    // Set version
    await fillFormField(page, 'Version', '1.0.0');

    // Submit
    await clickButton(page, 'Publish');

    // Wait for success
    await expectToast(page, 'Template published successfully');

    // Verify redirect to template details
    await page.waitForURL(/\/marketplace\/templates\/[a-f0-9-]+/);

    // Verify template details
    await expect(page.locator('h1, h2', { hasText: 'My Custom Template' })).toBeVisible();
  });
});
