import { test, expect, waitForLoading, navigateTo, searchFor, fillFormField, expectTableRow } from './setup';

test.describe('Audit Logs', () => {
  test.use({ storageState: 'admin-auth.json' }); // Use admin authentication

  test.beforeEach(async ({ adminPage }) => {
    // Navigate to audit logs page
    await navigateTo(adminPage, 'Audit Logs');
  });

  test('should display audit logs', async ({ adminPage: page }) => {
    // Wait for logs to load
    await waitForLoading(page);

    // Verify audit logs header
    await expect(page.locator('h1, h2', { hasText: 'Audit Logs' })).toBeVisible();

    // Verify audit log table/list
    const logRows = page.locator('[data-testid="audit-log-row"]');
    await expect(logRows.first()).toBeVisible();

    // Verify columns
    await expect(page.locator('th:has-text("Event Type")')).toBeVisible();
    await expect(page.locator('th:has-text("User")')).toBeVisible();
    await expect(page.locator('th:has-text("Resource")')).toBeVisible();
    await expect(page.locator('th:has-text("Action")')).toBeVisible();
    await expect(page.locator('th:has-text("Timestamp")')).toBeVisible();
  });

  test('should filter by event type', async ({ adminPage: page }) => {
    // Wait for page load
    await waitForLoading(page);

    // Open event type filter
    await page.click('[data-testid="event-type-filter"]');

    // Select "workflow.created"
    await page.click('[data-testid="event-type-workflow.created"]');

    // Wait for filtered results
    await waitForLoading(page);

    // Verify all visible logs are workflow.created
    const eventTypes = await page
      .locator('[data-testid="audit-log-event-type"]')
      .allTextContents();

    eventTypes.forEach((eventType) => {
      expect(eventType).toContain('workflow.created');
    });
  });

  test('should filter by user', async ({ adminPage: page }) => {
    // Wait for page load
    await waitForLoading(page);

    // Open user filter
    await page.click('[data-testid="user-filter"]');

    // Select specific user (first one in dropdown)
    await page.click('[data-testid="user-filter-option"]').first();

    // Wait for filtered results
    await waitForLoading(page);

    // Verify logs are filtered
    const logRows = page.locator('[data-testid="audit-log-row"]');
    await expect(logRows).not.toHaveCount(0);

    // Get selected user
    const selectedUser = await page
      .locator('[data-testid="user-filter-selected"]')
      .textContent();

    // Verify all logs belong to selected user
    const users = await page
      .locator('[data-testid="audit-log-user"]')
      .allTextContents();

    users.forEach((user) => {
      expect(user).toContain(selectedUser || '');
    });
  });

  test('should filter by date range', async ({ adminPage: page }) => {
    // Wait for page load
    await waitForLoading(page);

    // Set date range
    const today = new Date();
    const lastWeek = new Date(today);
    lastWeek.setDate(today.getDate() - 7);

    // Open date filter
    await page.click('[data-testid="date-filter"]');

    // Set start date
    await page.fill('[data-testid="start-date"]', lastWeek.toISOString().split('T')[0]);

    // Set end date
    await page.fill('[data-testid="end-date"]', today.toISOString().split('T')[0]);

    // Apply filter
    await page.click('[data-testid="apply-date-filter"]');

    // Wait for filtered results
    await waitForLoading(page);

    // Verify logs are within date range
    const timestamps = await page
      .locator('[data-testid="audit-log-timestamp"]')
      .allTextContents();

    timestamps.forEach((timestamp) => {
      const logDate = new Date(timestamp);
      expect(logDate.getTime()).toBeGreaterThanOrEqual(lastWeek.getTime());
      expect(logDate.getTime()).toBeLessThanOrEqual(today.getTime());
    });
  });

  test('should search audit logs', async ({ adminPage: page }) => {
    // Wait for page load
    await waitForLoading(page);

    // Search for specific resource
    await searchFor(page, 'workflow-123');

    // Verify results contain search term
    const logRows = page.locator('[data-testid="audit-log-row"]');
    await expect(logRows.first()).toBeVisible();

    const firstRowText = await logRows.first().textContent();
    expect(firstRowText).toContain('workflow-123');
  });

  test('should view audit log details', async ({ adminPage: page }) => {
    // Wait for logs to load
    await waitForLoading(page);

    // Click on first log row
    const firstRow = page.locator('[data-testid="audit-log-row"]').first();
    await firstRow.click();

    // Verify details modal/panel opens
    const detailsPanel = page.locator('[data-testid="audit-log-details"]');
    await expect(detailsPanel).toBeVisible();

    // Verify details fields
    await expect(detailsPanel.locator('[data-testid="event-type"]')).toBeVisible();
    await expect(detailsPanel.locator('[data-testid="user-info"]')).toBeVisible();
    await expect(detailsPanel.locator('[data-testid="resource-info"]')).toBeVisible();
    await expect(detailsPanel.locator('[data-testid="action"]')).toBeVisible();
    await expect(detailsPanel.locator('[data-testid="timestamp"]')).toBeVisible();
    await expect(detailsPanel.locator('[data-testid="metadata"]')).toBeVisible();

    // Verify metadata contains IP address and user agent
    const metadata = detailsPanel.locator('[data-testid="metadata"]');
    await expect(metadata).toContainText('ip_address');
    await expect(metadata).toContainText('user_agent');
  });

  test('should export audit logs', async ({ adminPage: page }) => {
    // Wait for page load
    await waitForLoading(page);

    // Click export button
    const downloadPromise = page.waitForEvent('download');
    await page.click('[data-testid="export-button"]');

    // Select export format (CSV)
    await page.click('[data-testid="export-format-csv"]');

    // Wait for download
    const download = await downloadPromise;

    // Verify download
    expect(download.suggestedFilename()).toContain('audit-logs');
    expect(download.suggestedFilename()).toContain('.csv');

    // Optionally: verify download size is not zero
    const path = await download.path();
    expect(path).toBeTruthy();
  });

  test('should display audit statistics', async ({ adminPage: page }) => {
    // Navigate to audit dashboard/statistics
    await page.click('[data-testid="audit-statistics-tab"]');

    // Wait for statistics to load
    await waitForLoading(page);

    // Verify statistics cards
    await expect(page.locator('[data-testid="total-events"]')).toBeVisible();
    await expect(page.locator('[data-testid="unique-users"]')).toBeVisible();
    await expect(page.locator('[data-testid="event-types"]')).toBeVisible();

    // Verify charts
    await expect(page.locator('[data-testid="activity-chart"]')).toBeVisible();
    await expect(page.locator('[data-testid="action-distribution"]')).toBeVisible();

    // Verify table/list of top events
    await expect(page.locator('[data-testid="top-events-table"]')).toBeVisible();
  });

  test('should display daily activity trend', async ({ adminPage: page }) => {
    // Navigate to statistics
    await page.click('[data-testid="audit-statistics-tab"]');
    await waitForLoading(page);

    // Verify activity chart
    const activityChart = page.locator('[data-testid="activity-chart"]');
    await expect(activityChart).toBeVisible();

    // Verify chart has data points
    const dataPoints = activityChart.locator('.recharts-line-dots circle, .chart-point');
    await expect(dataPoints.first()).toBeVisible();

    // Verify chart legend
    await expect(activityChart.locator('.recharts-legend, .chart-legend')).toBeVisible();
  });

  test('should filter statistics by time period', async ({ adminPage: page }) => {
    // Navigate to statistics
    await page.click('[data-testid="audit-statistics-tab"]');
    await waitForLoading(page);

    // Select time period (last 7 days)
    await page.click('[data-testid="time-period-selector"]');
    await page.click('[data-testid="period-7days"]');

    // Wait for updated statistics
    await waitForLoading(page);

    // Verify statistics updated
    const totalEvents = page.locator('[data-testid="total-events"]');
    await expect(totalEvents).toBeVisible();

    // Verify chart updated
    const chartTitle = page.locator('[data-testid="chart-title"]');
    await expect(chartTitle).toContainText('Last 7 Days');
  });

  test('should paginate through audit logs', async ({ adminPage: page }) => {
    // Wait for page load
    await waitForLoading(page);

    // Get first page logs
    const firstPageFirstLog = await page
      .locator('[data-testid="audit-log-row"]')
      .first()
      .textContent();

    // Click next page
    await page.click('[data-testid="pagination-next"]');

    // Wait for page load
    await waitForLoading(page);

    // Get second page logs
    const secondPageFirstLog = await page
      .locator('[data-testid="audit-log-row"]')
      .first()
      .textContent();

    // Verify logs changed
    expect(firstPageFirstLog).not.toBe(secondPageFirstLog);

    // Verify pagination controls
    await expect(page.locator('[data-testid="pagination-prev"]')).toBeEnabled();
    await expect(page.locator('[data-testid="current-page"]')).toContainText('2');
  });

  test('should clear all filters', async ({ adminPage: page }) => {
    // Wait for page load
    await waitForLoading(page);

    // Apply multiple filters
    await page.click('[data-testid="event-type-filter"]');
    await page.click('[data-testid="event-type-workflow.created"]');
    await searchFor(page, 'test');

    // Wait for filtered results
    await waitForLoading(page);

    // Get filtered count
    const filteredCount = await page
      .locator('[data-testid="audit-log-row"]')
      .count();

    // Click clear filters
    await page.click('[data-testid="clear-filters"]');

    // Wait for unfiltered results
    await waitForLoading(page);

    // Get unfiltered count
    const unfilteredCount = await page
      .locator('[data-testid="audit-log-row"]')
      .count();

    // Verify more results after clearing
    expect(unfilteredCount).toBeGreaterThan(filteredCount);

    // Verify search is cleared
    const searchInput = page.locator('input[type="search"]');
    await expect(searchInput).toHaveValue('');
  });
});

test.describe('Audit Logs - Real-time Updates', () => {
  test('should show new logs in real-time', async ({ adminPage: page, authenticatedPage: userPage }) => {
    // Open audit logs as admin
    await navigateTo(page, 'Audit Logs');
    await waitForLoading(page);

    // Get current log count
    const initialCount = await page.locator('[data-testid="audit-log-row"]').count();

    // Perform action in another tab (create workflow)
    await userPage.goto('/workflows');
    await userPage.click('[data-testid="create-workflow"]');
    await fillFormField(userPage, 'Name', 'Test Workflow for Audit');
    await userPage.click('button[type="submit"]');

    // Wait a bit for audit log to be created
    await page.waitForTimeout(2000);

    // Reload or wait for real-time update
    await page.reload();
    await waitForLoading(page);

    // Verify new log appeared
    const newCount = await page.locator('[data-testid="audit-log-row"]').count();
    expect(newCount).toBeGreaterThan(initialCount);

    // Verify the new log is visible
    await expectTableRow(page, 'workflow.created');
  });
});
