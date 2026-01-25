import { Page, Locator, expect } from '@playwright/test'
import { waitForPageLoad, waitForApiResponse } from '../utils/test-helpers'

/**
 * Step result from an execution
 */
export interface StepResult {
  nodeId: string
  nodeName: string
  status: 'pending' | 'running' | 'completed' | 'failed' | 'skipped'
  duration?: number
  error?: string
  output?: unknown
}

/**
 * Page Object Model for the Executions page
 * Provides methods for viewing and managing workflow executions
 */
export class ExecutionsPage {
  readonly page: Page
  readonly executionsList: Locator
  readonly searchInput: Locator
  readonly statusFilter: Locator
  readonly dateFilter: Locator
  readonly refreshButton: Locator

  constructor(page: Page) {
    this.page = page
    this.executionsList = page.locator('[data-testid="executions-list"], .executions-list')
    this.searchInput = page.locator('input[type="search"], [data-testid="search-executions"]')
    this.statusFilter = page.locator('[data-testid="status-filter"], select[name="status"]')
    this.dateFilter = page.locator('[data-testid="date-filter"], input[type="date"]')
    this.refreshButton = page.locator('button:has-text("Refresh"), [data-testid="refresh-executions"]')
  }

  /**
   * Navigate to the executions list page
   */
  async goto(): Promise<void> {
    await this.page.goto('/executions')
    await waitForPageLoad(this.page)
    await expect(this.executionsList).toBeVisible({ timeout: 10000 })
  }

  /**
   * Navigate to a specific execution detail page
   */
  async gotoExecution(executionId: string): Promise<void> {
    await this.page.goto(`/executions/${executionId}`)
    await waitForPageLoad(this.page)
  }

  /**
   * Search for executions
   */
  async search(query: string): Promise<void> {
    await this.searchInput.fill(query)
    await this.page.waitForTimeout(500) // Debounce
    await waitForApiResponse(this.page, /\/api\/v1\/executions/)
  }

  /**
   * Filter executions by status
   */
  async filterByStatus(status: 'all' | 'running' | 'completed' | 'failed'): Promise<void> {
    await this.statusFilter.selectOption(status)
    await waitForApiResponse(this.page, /\/api\/v1\/executions/)
  }

  /**
   * Filter executions by date range
   */
  async filterByDateRange(startDate: string, endDate: string): Promise<void> {
    const startInput = this.page.locator('[data-testid="date-start"], input[name="startDate"]')
    const endInput = this.page.locator('[data-testid="date-end"], input[name="endDate"]')

    await startInput.fill(startDate)
    await endInput.fill(endDate)
    await waitForApiResponse(this.page, /\/api\/v1\/executions/)
  }

  /**
   * Click on an execution to view details
   */
  async viewExecution(executionId: string): Promise<void> {
    const row = this.executionsList.locator(`tr:has-text("${executionId}"), [data-execution-id="${executionId}"]`)
    await row.click()
    await this.page.waitForURL(/\/executions\/[a-f0-9-]+/)
    await waitForPageLoad(this.page)
  }

  /**
   * Wait for an execution to complete (with polling)
   */
  async waitForCompletion(executionId: string, timeout: number = 60000): Promise<string> {
    const startTime = Date.now()

    while (Date.now() - startTime < timeout) {
      await this.gotoExecution(executionId)
      const status = await this.getStatus()

      if (status === 'completed' || status === 'failed') {
        return status
      }

      // Refresh and wait
      await this.page.waitForTimeout(2000)
      await this.page.reload()
    }

    throw new Error(`Execution ${executionId} did not complete within ${timeout}ms`)
  }

  /**
   * Get the current execution status
   */
  async getStatus(): Promise<string> {
    const statusBadge = this.page.locator('[data-testid="execution-status"], .status-badge')
    const text = await statusBadge.textContent()
    return (text || 'unknown').toLowerCase()
  }

  /**
   * Get all step results from the execution
   */
  async getStepResults(): Promise<StepResult[]> {
    const steps: StepResult[] = []
    const stepElements = this.page.locator('[data-testid="execution-step"], .execution-step')
    const count = await stepElements.count()

    for (let i = 0; i < count; i++) {
      const step = stepElements.nth(i)
      const nodeId = await step.getAttribute('data-node-id') || ''
      const nodeName = await step.locator('.step-name, [data-testid="step-name"]').textContent() || ''
      const statusText = await step.locator('.step-status, [data-testid="step-status"]').textContent() || ''

      steps.push({
        nodeId,
        nodeName,
        status: statusText.toLowerCase() as StepResult['status']
      })
    }

    return steps
  }

  /**
   * Get the output of a specific step
   */
  async getStepOutput(nodeId: string): Promise<unknown> {
    const step = this.page.locator(`[data-node-id="${nodeId}"], .execution-step:has([data-node-id="${nodeId}"])`)
    await step.click()

    const outputPanel = this.page.locator('[data-testid="step-output"], .step-output-panel')
    await expect(outputPanel).toBeVisible()

    const outputText = await outputPanel.textContent()
    try {
      return JSON.parse(outputText || '{}')
    } catch {
      return outputText
    }
  }

  /**
   * Get the error message if execution failed
   */
  async getErrorMessage(): Promise<string | null> {
    const errorElement = this.page.locator('[data-testid="execution-error"], .error-message')

    if (await errorElement.isVisible()) {
      return await errorElement.textContent()
    }

    return null
  }

  /**
   * Cancel a running execution
   */
  async cancelExecution(): Promise<void> {
    const cancelButton = this.page.locator('button:has-text("Cancel"), [data-testid="cancel-execution"]')
    await cancelButton.click()

    // Confirm cancellation
    const confirmButton = this.page.locator('button:has-text("Confirm"), [data-testid="confirm-cancel"]')
    if (await confirmButton.isVisible()) {
      await confirmButton.click()
    }

    // Wait for status to update
    await waitForApiResponse(this.page, /\/api\/v1\/executions\/.*\/cancel/)
  }

  /**
   * Retry a failed execution
   */
  async retryExecution(): Promise<string> {
    const retryButton = this.page.locator('button:has-text("Retry"), [data-testid="retry-execution"]')
    await retryButton.click()

    // Wait for new execution to be created
    const response = await waitForApiResponse(this.page, /\/api\/v1\/executions/)
    const data = await response.json()

    return data.id || data.execution_id
  }

  /**
   * Get execution duration
   */
  async getDuration(): Promise<number | null> {
    const durationElement = this.page.locator('[data-testid="execution-duration"], .duration')

    if (await durationElement.isVisible()) {
      const text = await durationElement.textContent()
      // Parse duration string (e.g., "2.5s", "1m 30s")
      const match = text?.match(/(\d+(?:\.\d+)?)\s*(ms|s|m|h)?/)
      if (match) {
        const value = parseFloat(match[1])
        const unit = match[2] || 's'
        switch (unit) {
          case 'ms': return value
          case 's': return value * 1000
          case 'm': return value * 60 * 1000
          case 'h': return value * 60 * 60 * 1000
        }
      }
    }

    return null
  }

  /**
   * Refresh the executions list or detail view
   */
  async refresh(): Promise<void> {
    await this.refreshButton.click()
    await waitForApiResponse(this.page, /\/api\/v1\/executions/)
  }

  /**
   * Get the list of all visible executions
   */
  async getExecutionsList(): Promise<{ id: string; status: string; workflowName: string }[]> {
    const executions: { id: string; status: string; workflowName: string }[] = []
    const rows = this.executionsList.locator('tr, .execution-row').filter({ has: this.page.locator('[data-execution-id]') })
    const count = await rows.count()

    for (let i = 0; i < count; i++) {
      const row = rows.nth(i)
      const id = await row.getAttribute('data-execution-id') ||
                 await row.locator('[data-testid="execution-id"]').textContent() || ''
      const status = await row.locator('.status-badge, [data-testid="status"]').textContent() || ''
      const workflowName = await row.locator('.workflow-name, [data-testid="workflow-name"]').textContent() || ''

      executions.push({
        id: id.trim(),
        status: status.toLowerCase().trim(),
        workflowName: workflowName.trim()
      })
    }

    return executions
  }

  /**
   * Check if there are any running executions
   */
  async hasRunningExecutions(): Promise<boolean> {
    const runningBadge = this.executionsList.locator('text=/running/i')
    return await runningBadge.isVisible()
  }

  /**
   * Download execution logs
   */
  async downloadLogs(): Promise<void> {
    const downloadButton = this.page.locator('button:has-text("Download Logs"), [data-testid="download-logs"]')
    await downloadButton.click()

    // Wait for download to start
    const [download] = await Promise.all([
      this.page.waitForEvent('download'),
      downloadButton.click()
    ])

    await download.saveAs(`./downloads/${download.suggestedFilename()}`)
  }

  /**
   * View the execution timeline/graph
   */
  async viewTimeline(): Promise<void> {
    const timelineTab = this.page.locator('button:has-text("Timeline"), [data-testid="timeline-tab"]')
    await timelineTab.click()
    await expect(this.page.locator('[data-testid="execution-timeline"], .execution-timeline')).toBeVisible()
  }

  /**
   * View the execution logs
   */
  async viewLogs(): Promise<void> {
    const logsTab = this.page.locator('button:has-text("Logs"), [data-testid="logs-tab"]')
    await logsTab.click()
    await expect(this.page.locator('[data-testid="execution-logs"], .execution-logs')).toBeVisible()
  }
}
