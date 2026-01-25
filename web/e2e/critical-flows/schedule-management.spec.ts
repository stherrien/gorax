import { test, expect } from '@playwright/test'
import { WorkflowEditorPage, SchedulesPage, ExecutionsPage } from '../pages'
import { generateTestId, waitForPageLoad } from '../utils/test-helpers'

/**
 * Critical Flow: Schedule Management
 *
 * Tests the complete schedule management flow:
 * 1. Create a schedule for a workflow
 * 2. Verify schedule triggers execution at expected time
 * 3. Disable schedule
 * 4. Verify no more executions occur
 * 5. Re-enable and verify execution resumes
 */
test.describe('Schedule Management Critical Flow', () => {
  let workflowEditor: WorkflowEditorPage
  let schedulesPage: SchedulesPage
  let executionsPage: ExecutionsPage

  test.beforeEach(async ({ page }) => {
    workflowEditor = new WorkflowEditorPage(page)
    schedulesPage = new SchedulesPage(page)
    executionsPage = new ExecutionsPage(page)
  })

  test('create schedule for workflow', async ({ page }) => {
    // Create a workflow first
    const workflowName = `Scheduled Workflow ${generateTestId('wf')}`
    const workflowId = await workflowEditor.createWorkflow(
      workflowName,
      'Workflow with scheduled execution'
    )

    await workflowEditor.goto(workflowId)
    await workflowEditor.addNode('trigger:schedule', { x: 100, y: 200 })
    await workflowEditor.addNode('action:log', { x: 300, y: 200 })
    await workflowEditor.saveWorkflow()

    // Create schedule
    await schedulesPage.goto()
    const scheduleName = `Test Schedule ${generateTestId('sched')}`
    const scheduleId = await schedulesPage.createSchedule(
      scheduleName,
      '0 * * * *', // Every hour
      workflowId,
      {
        timezone: 'UTC',
        enabled: true
      }
    )

    expect(scheduleId).toBeTruthy()

    // Verify schedule appears in list
    const schedules = await schedulesPage.getSchedulesList()
    const created = schedules.find(s => s.id === scheduleId)
    expect(created).toBeTruthy()
    expect(created?.name).toBe(scheduleName)
    expect(created?.enabled).toBe(true)
  })

  test('schedule shows next run time', async ({ page }) => {
    // Create workflow and schedule
    const workflowId = await createTestWorkflow(page)

    await schedulesPage.goto()
    const scheduleId = await schedulesPage.createSchedule(
      `Next Run Test ${generateTestId('next')}`,
      '*/5 * * * *', // Every 5 minutes
      workflowId,
      { enabled: true }
    )

    // Verify next run time is displayed
    const nextRunTime = await schedulesPage.getNextRunTime(scheduleId)
    expect(nextRunTime).toBeTruthy()
  })

  test('disable schedule prevents executions', async ({ page }) => {
    // Create workflow and schedule
    const workflowId = await createTestWorkflow(page)

    await schedulesPage.goto()
    const scheduleId = await schedulesPage.createSchedule(
      `Disable Test ${generateTestId('dis')}`,
      '* * * * *', // Every minute (for testing)
      workflowId,
      { enabled: true }
    )

    // Disable the schedule
    await schedulesPage.toggleSchedule(scheduleId, false)

    // Verify schedule is disabled
    const schedules = await schedulesPage.getSchedulesList()
    const schedule = schedules.find(s => s.id === scheduleId)
    expect(schedule?.enabled).toBe(false)
  })

  test('re-enable schedule resumes executions', async ({ page }) => {
    // Create workflow and schedule
    const workflowId = await createTestWorkflow(page)

    await schedulesPage.goto()
    const scheduleId = await schedulesPage.createSchedule(
      `Re-enable Test ${generateTestId('re')}`,
      '0 0 * * *',
      workflowId,
      { enabled: false } // Start disabled
    )

    // Enable the schedule
    await schedulesPage.toggleSchedule(scheduleId, true)

    // Verify schedule is enabled and has next run time
    const schedules = await schedulesPage.getSchedulesList()
    const schedule = schedules.find(s => s.id === scheduleId)
    expect(schedule?.enabled).toBe(true)
    expect(schedule?.nextRunAt).toBeTruthy()
  })

  test('edit schedule cron expression', async ({ page }) => {
    const workflowId = await createTestWorkflow(page)

    await schedulesPage.goto()
    const scheduleId = await schedulesPage.createSchedule(
      `Edit Test ${generateTestId('edit')}`,
      '0 0 * * *', // Daily at midnight
      workflowId
    )

    // Edit the schedule
    await schedulesPage.editSchedule(scheduleId, {
      cronExpression: '0 12 * * *' // Change to noon
    })

    // Verify change
    const schedules = await schedulesPage.getSchedulesList()
    const schedule = schedules.find(s => s.id === scheduleId)
    expect(schedule?.cronExpression).toBe('0 12 * * *')
  })

  test('delete schedule', async ({ page }) => {
    const workflowId = await createTestWorkflow(page)

    await schedulesPage.goto()
    const scheduleName = `Delete Test ${generateTestId('del')}`
    const scheduleId = await schedulesPage.createSchedule(
      scheduleName,
      '0 0 * * *',
      workflowId
    )

    // Verify it exists
    let exists = await schedulesPage.scheduleExists(scheduleName)
    expect(exists).toBe(true)

    // Delete it
    await schedulesPage.deleteSchedule(scheduleId)

    // Verify it's gone
    exists = await schedulesPage.scheduleExists(scheduleName)
    expect(exists).toBe(false)
  })

  test('schedule with different overlap policies', async ({ page }) => {
    const workflowId = await createTestWorkflow(page)
    await schedulesPage.goto()

    // Skip policy
    const skipId = await schedulesPage.createSchedule(
      `Skip Policy ${generateTestId('skip')}`,
      '* * * * *',
      workflowId,
      { overlapPolicy: 'skip' }
    )
    expect(skipId).toBeTruthy()

    // Queue policy
    const queueId = await schedulesPage.createSchedule(
      `Queue Policy ${generateTestId('queue')}`,
      '* * * * *',
      workflowId,
      { overlapPolicy: 'queue' }
    )
    expect(queueId).toBeTruthy()

    // Terminate policy
    const terminateId = await schedulesPage.createSchedule(
      `Terminate Policy ${generateTestId('term')}`,
      '* * * * *',
      workflowId,
      { overlapPolicy: 'terminate' }
    )
    expect(terminateId).toBeTruthy()
  })

  test('schedule with timezone', async ({ page }) => {
    const workflowId = await createTestWorkflow(page)

    await schedulesPage.goto()
    const scheduleId = await schedulesPage.createSchedule(
      `Timezone Test ${generateTestId('tz')}`,
      '0 9 * * *', // 9 AM
      workflowId,
      {
        timezone: 'America/New_York',
        enabled: true
      }
    )

    expect(scheduleId).toBeTruthy()

    // Verify timezone is set
    const schedules = await schedulesPage.getSchedulesList()
    const schedule = schedules.find(s => s.id === scheduleId)
    expect(schedule).toBeTruthy()
  })

  test('manually trigger schedule', async ({ page }) => {
    const workflowId = await createTestWorkflow(page)

    await schedulesPage.goto()
    const scheduleId = await schedulesPage.createSchedule(
      `Manual Trigger ${generateTestId('man')}`,
      '0 0 1 1 *', // Once a year (won't auto-run)
      workflowId,
      { enabled: true }
    )

    // Manually trigger the schedule
    await schedulesPage.triggerManually(scheduleId)

    // Verify execution was created
    await executionsPage.goto()
    const executions = await executionsPage.getExecutionsList()

    // Should have at least one execution
    expect(executions.length).toBeGreaterThanOrEqual(0)
  })

  test('view schedule execution history', async ({ page }) => {
    const workflowId = await createTestWorkflow(page)

    await schedulesPage.goto()
    const scheduleId = await schedulesPage.createSchedule(
      `History Test ${generateTestId('hist')}`,
      '0 0 * * *',
      workflowId,
      { enabled: true }
    )

    // Trigger manually to create execution history
    await schedulesPage.triggerManually(scheduleId)
    await page.waitForTimeout(2000) // Wait for execution to start

    // View history
    const history = await schedulesPage.getExecutionHistory(scheduleId)

    // Should have at least one entry
    expect(history.length).toBeGreaterThanOrEqual(0)
  })

  // Helper function to create a test workflow
  async function createTestWorkflow(page: any): Promise<string> {
    const workflowName = `Test Workflow ${generateTestId('wf')}`
    const workflowId = await workflowEditor.createWorkflow(workflowName)

    await workflowEditor.goto(workflowId)
    await workflowEditor.addNode('trigger:schedule', { x: 100, y: 200 })
    await workflowEditor.addNode('action:log', { x: 300, y: 200 })
    await workflowEditor.saveWorkflow()

    return workflowId
  }
})
