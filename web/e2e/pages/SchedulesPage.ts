import { Page, Locator, expect } from '@playwright/test'
import { waitForPageLoad, waitForApiResponse, generateTestId } from '../utils/test-helpers'

/**
 * Page Object Model for the Schedules Management page
 * Provides methods for creating and managing workflow schedules
 */
export class SchedulesPage {
  readonly page: Page
  readonly schedulesList: Locator
  readonly createButton: Locator
  readonly searchInput: Locator
  readonly statusFilter: Locator
  readonly workflowFilter: Locator

  constructor(page: Page) {
    this.page = page
    this.schedulesList = page.locator('[data-testid="schedules-list"], .schedules-list')
    this.createButton = page.locator('button:has-text("Create"), [data-testid="create-schedule"]')
    this.searchInput = page.locator('input[type="search"], [data-testid="search-schedules"]')
    this.statusFilter = page.locator('[data-testid="status-filter"], select[name="status"]')
    this.workflowFilter = page.locator('[data-testid="workflow-filter"], select[name="workflow"]')
  }

  /**
   * Navigate to the schedules page
   */
  async goto(): Promise<void> {
    await this.page.goto('/schedules')
    await waitForPageLoad(this.page)
  }

  /**
   * Create a new schedule
   */
  async createSchedule(
    name: string,
    cronExpression: string,
    workflowId: string,
    options?: {
      timezone?: string
      overlapPolicy?: 'skip' | 'queue' | 'terminate'
      enabled?: boolean
      description?: string
    }
  ): Promise<string> {
    await this.createButton.click()

    // Wait for modal to open
    const modal = this.page.locator('[data-testid="schedule-modal"], .schedule-modal')
    await expect(modal).toBeVisible()

    // Fill name
    await modal.locator('input[name="name"], [data-testid="schedule-name"]').fill(name)

    // Select workflow
    await modal.locator('select[name="workflowId"], [data-testid="schedule-workflow"]').selectOption(workflowId)

    // Fill cron expression
    await modal.locator('input[name="cronExpression"], [data-testid="cron-expression"]').fill(cronExpression)

    // Set timezone if provided
    if (options?.timezone) {
      await modal.locator('select[name="timezone"], [data-testid="timezone"]').selectOption(options.timezone)
    }

    // Set overlap policy if provided
    if (options?.overlapPolicy) {
      await modal.locator('select[name="overlapPolicy"], [data-testid="overlap-policy"]').selectOption(options.overlapPolicy)
    }

    // Set enabled state
    if (options?.enabled !== undefined) {
      const enabledCheckbox = modal.locator('input[name="enabled"], [data-testid="schedule-enabled"]')
      if (options.enabled) {
        await enabledCheckbox.check()
      } else {
        await enabledCheckbox.uncheck()
      }
    }

    // Fill description if provided
    if (options?.description) {
      await modal.locator('textarea[name="description"], [data-testid="schedule-description"]').fill(options.description)
    }

    // Submit
    await modal.locator('button:has-text("Create"), button[type="submit"]').click()

    // Wait for API response
    const response = await waitForApiResponse(this.page, /\/api\/v1\/schedules/)
    const data = await response.json()

    // Wait for modal to close
    await expect(modal).not.toBeVisible()

    return data.id || generateTestId('schedule')
  }

  /**
   * Get a schedule row by ID
   */
  getScheduleRow(scheduleId: string): Locator {
    return this.schedulesList.locator(`tr:has([data-schedule-id="${scheduleId}"]), [data-schedule-id="${scheduleId}"]`)
  }

  /**
   * Toggle schedule enabled/disabled state
   */
  async toggleSchedule(scheduleId: string, enabled: boolean): Promise<void> {
    const row = this.getScheduleRow(scheduleId)

    // Find the toggle/switch
    const toggle = row.locator('input[type="checkbox"], [data-testid="schedule-toggle"]')

    if (enabled) {
      await toggle.check()
    } else {
      await toggle.uncheck()
    }

    // Wait for API response
    await waitForApiResponse(this.page, /\/api\/v1\/schedules/)
  }

  /**
   * Delete a schedule
   */
  async deleteSchedule(scheduleId: string): Promise<void> {
    const row = this.getScheduleRow(scheduleId)

    // Click delete button
    const deleteButton = row.locator('button:has-text("Delete"), [data-testid="delete-schedule"]')
    await deleteButton.click()

    // Confirm deletion
    const confirmModal = this.page.locator('[data-testid="confirm-modal"], .confirm-dialog')
    await expect(confirmModal).toBeVisible()

    await confirmModal.locator('button:has-text("Delete"), button:has-text("Confirm")').click()

    // Wait for API response
    await waitForApiResponse(this.page, /\/api\/v1\/schedules/)

    // Verify row is removed
    await expect(row).not.toBeVisible()
  }

  /**
   * Edit a schedule
   */
  async editSchedule(
    scheduleId: string,
    updates: {
      name?: string
      cronExpression?: string
      timezone?: string
      overlapPolicy?: string
    }
  ): Promise<void> {
    const row = this.getScheduleRow(scheduleId)

    // Click edit button
    const editButton = row.locator('button:has-text("Edit"), [data-testid="edit-schedule"]')
    await editButton.click()

    // Wait for modal
    const modal = this.page.locator('[data-testid="schedule-modal"], .schedule-modal')
    await expect(modal).toBeVisible()

    // Update fields
    if (updates.name) {
      await modal.locator('input[name="name"]').fill(updates.name)
    }
    if (updates.cronExpression) {
      await modal.locator('input[name="cronExpression"]').fill(updates.cronExpression)
    }
    if (updates.timezone) {
      await modal.locator('select[name="timezone"]').selectOption(updates.timezone)
    }
    if (updates.overlapPolicy) {
      await modal.locator('select[name="overlapPolicy"]').selectOption(updates.overlapPolicy)
    }

    // Save
    await modal.locator('button:has-text("Save"), button[type="submit"]').click()

    // Wait for API response
    await waitForApiResponse(this.page, /\/api\/v1\/schedules/)
  }

  /**
   * View execution history for a schedule
   */
  async viewHistory(scheduleId: string): Promise<void> {
    const row = this.getScheduleRow(scheduleId)

    // Click history button or row
    const historyButton = row.locator('button:has-text("History"), [data-testid="view-history"]')
    if (await historyButton.isVisible()) {
      await historyButton.click()
    } else {
      await row.click()
    }

    // Wait for history panel/page
    const historySection = this.page.locator('[data-testid="schedule-history"], .schedule-history')
    await expect(historySection).toBeVisible()
  }

  /**
   * Get list of all visible schedules
   */
  async getSchedulesList(): Promise<{ id: string; name: string; cronExpression: string; enabled: boolean; nextRunAt?: string }[]> {
    const schedules: { id: string; name: string; cronExpression: string; enabled: boolean; nextRunAt?: string }[] = []
    const rows = this.schedulesList.locator('tr, .schedule-row').filter({ has: this.page.locator('[data-schedule-id]') })
    const count = await rows.count()

    for (let i = 0; i < count; i++) {
      const row = rows.nth(i)
      const id = await row.getAttribute('data-schedule-id') ||
                 await row.locator('[data-testid="schedule-id"]').textContent() || ''
      const name = await row.locator('.schedule-name, [data-testid="schedule-name"]').textContent() || ''
      const cronExpression = await row.locator('.cron-expression, [data-testid="cron-expression"]').textContent() || ''
      const enabledElement = row.locator('input[type="checkbox"], [data-testid="schedule-toggle"]')
      const enabled = await enabledElement.isChecked()
      const nextRunAt = await row.locator('.next-run, [data-testid="next-run"]').textContent() || undefined

      schedules.push({
        id: id.trim(),
        name: name.trim(),
        cronExpression: cronExpression.trim(),
        enabled,
        nextRunAt: nextRunAt?.trim()
      })
    }

    return schedules
  }

  /**
   * Search for schedules
   */
  async search(query: string): Promise<void> {
    await this.searchInput.fill(query)
    await this.page.waitForTimeout(500) // Debounce
    await waitForApiResponse(this.page, /\/api\/v1\/schedules/)
  }

  /**
   * Filter schedules by status
   */
  async filterByStatus(status: 'all' | 'enabled' | 'disabled'): Promise<void> {
    await this.statusFilter.selectOption(status)
    await waitForApiResponse(this.page, /\/api\/v1\/schedules/)
  }

  /**
   * Filter schedules by workflow
   */
  async filterByWorkflow(workflowId: string): Promise<void> {
    await this.workflowFilter.selectOption(workflowId)
    await waitForApiResponse(this.page, /\/api\/v1\/schedules/)
  }

  /**
   * Get the next run time for a schedule
   */
  async getNextRunTime(scheduleId: string): Promise<string | null> {
    const row = this.getScheduleRow(scheduleId)
    const nextRunElement = row.locator('.next-run, [data-testid="next-run"]')

    if (await nextRunElement.isVisible()) {
      return await nextRunElement.textContent()
    }

    return null
  }

  /**
   * Get the last run time for a schedule
   */
  async getLastRunTime(scheduleId: string): Promise<string | null> {
    const row = this.getScheduleRow(scheduleId)
    const lastRunElement = row.locator('.last-run, [data-testid="last-run"]')

    if (await lastRunElement.isVisible()) {
      return await lastRunElement.textContent()
    }

    return null
  }

  /**
   * Check if a schedule exists
   */
  async scheduleExists(name: string): Promise<boolean> {
    await this.search(name)
    const row = this.schedulesList.locator(`tr:has-text("${name}")`)
    return await row.isVisible()
  }

  /**
   * Manually trigger a schedule
   */
  async triggerManually(scheduleId: string): Promise<void> {
    const row = this.getScheduleRow(scheduleId)

    // Click trigger button
    const triggerButton = row.locator('button:has-text("Run Now"), [data-testid="trigger-schedule"]')
    await triggerButton.click()

    // Confirm if needed
    const confirmModal = this.page.locator('[data-testid="confirm-modal"]')
    if (await confirmModal.isVisible()) {
      await confirmModal.locator('button:has-text("Confirm")').click()
    }

    // Wait for execution to be created
    await waitForApiResponse(this.page, /\/api\/v1\/executions|\/api\/v1\/schedules.*trigger/)
  }

  /**
   * Get schedule execution history
   */
  async getExecutionHistory(scheduleId: string): Promise<{ executionId: string; status: string; startedAt: string }[]> {
    await this.viewHistory(scheduleId)

    const history: { executionId: string; status: string; startedAt: string }[] = []
    const rows = this.page.locator('[data-testid="history-row"], .history-row')
    const count = await rows.count()

    for (let i = 0; i < count; i++) {
      const row = rows.nth(i)
      const executionId = await row.locator('[data-testid="execution-id"]').textContent() || ''
      const status = await row.locator('[data-testid="execution-status"]').textContent() || ''
      const startedAt = await row.locator('[data-testid="started-at"]').textContent() || ''

      history.push({
        executionId: executionId.trim(),
        status: status.trim(),
        startedAt: startedAt.trim()
      })
    }

    return history
  }

  /**
   * Use the cron builder to set cron expression
   */
  async useCronBuilder(options: {
    frequency: 'minute' | 'hour' | 'day' | 'week' | 'month'
    at?: string
    dayOfWeek?: string[]
    dayOfMonth?: number
  }): Promise<void> {
    // Click on cron builder toggle
    const builderToggle = this.page.locator('button:has-text("Builder"), [data-testid="cron-builder-toggle"]')
    await builderToggle.click()

    const builder = this.page.locator('[data-testid="cron-builder"], .cron-builder')
    await expect(builder).toBeVisible()

    // Select frequency
    await builder.locator('select[name="frequency"], [data-testid="frequency"]').selectOption(options.frequency)

    // Set time if provided
    if (options.at) {
      await builder.locator('input[name="time"], [data-testid="time-input"]').fill(options.at)
    }

    // Set day of week if provided
    if (options.dayOfWeek) {
      for (const day of options.dayOfWeek) {
        await builder.locator(`input[value="${day}"], [data-day="${day}"]`).check()
      }
    }

    // Set day of month if provided
    if (options.dayOfMonth) {
      await builder.locator('input[name="dayOfMonth"], [data-testid="day-of-month"]').fill(options.dayOfMonth.toString())
    }
  }

  /**
   * Preview next run times
   */
  async previewNextRuns(count: number = 5): Promise<string[]> {
    const previewButton = this.page.locator('button:has-text("Preview"), [data-testid="preview-runs"]')
    await previewButton.click()

    const nextRuns: string[] = []
    const previewItems = this.page.locator('[data-testid="preview-item"], .preview-item')
    const itemCount = Math.min(await previewItems.count(), count)

    for (let i = 0; i < itemCount; i++) {
      const text = await previewItems.nth(i).textContent()
      if (text) nextRuns.push(text.trim())
    }

    return nextRuns
  }
}
