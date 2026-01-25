/**
 * Mock data for testing.
 * Use these factories to create consistent test data.
 */

export interface MockSchedule {
  id: string
  workflow_id: string
  tenant_id: string
  name: string
  cron_expression: string
  timezone: string
  enabled: boolean
  created_at: string
  updated_at: string
  next_run_at: string | null
  last_run_at: string | null
}

export interface MockWorkflow {
  id: string
  tenant_id: string
  name: string
  description: string
  version: number
  status: 'draft' | 'published' | 'archived'
  created_at: string
  updated_at: string
}

export interface MockExecution {
  id: string
  workflow_id: string
  tenant_id: string
  status: 'pending' | 'running' | 'completed' | 'failed' | 'cancelled'
  started_at: string
  completed_at: string | null
  error: string | null
}

export interface MockUser {
  id: string
  email: string
  name: string
  role: 'admin' | 'user' | 'viewer'
  tenant_id: string
}

// Factory functions
let idCounter = 0

function generateId(prefix = 'mock'): string {
  return `${prefix}-${++idCounter}-${Date.now()}`
}

export function createMockSchedule(overrides: Partial<MockSchedule> = {}): MockSchedule {
  const id = generateId('schedule')
  const now = new Date().toISOString()

  return {
    id,
    workflow_id: generateId('workflow'),
    tenant_id: generateId('tenant'),
    name: 'Test Schedule',
    cron_expression: '0 9 * * *',
    timezone: 'UTC',
    enabled: true,
    created_at: now,
    updated_at: now,
    next_run_at: new Date(Date.now() + 86400000).toISOString(),
    last_run_at: null,
    ...overrides,
  }
}

export function createMockWorkflow(overrides: Partial<MockWorkflow> = {}): MockWorkflow {
  const id = generateId('workflow')
  const now = new Date().toISOString()

  return {
    id,
    tenant_id: generateId('tenant'),
    name: 'Test Workflow',
    description: 'A test workflow for testing',
    version: 1,
    status: 'draft',
    created_at: now,
    updated_at: now,
    ...overrides,
  }
}

export function createMockExecution(overrides: Partial<MockExecution> = {}): MockExecution {
  const id = generateId('execution')
  const now = new Date().toISOString()

  return {
    id,
    workflow_id: generateId('workflow'),
    tenant_id: generateId('tenant'),
    status: 'completed',
    started_at: now,
    completed_at: now,
    error: null,
    ...overrides,
  }
}

export function createMockUser(overrides: Partial<MockUser> = {}): MockUser {
  const id = generateId('user')

  return {
    id,
    email: 'test@example.com',
    name: 'Test User',
    role: 'user',
    tenant_id: generateId('tenant'),
    ...overrides,
  }
}

// Predefined test data sets
export const mockSchedules = {
  single: createMockSchedule({ name: 'Daily Report' }),
  list: [
    createMockSchedule({ name: 'Daily Report', cron_expression: '0 9 * * *' }),
    createMockSchedule({ name: 'Weekly Cleanup', cron_expression: '0 0 * * 0', enabled: false }),
    createMockSchedule({ name: 'Hourly Check', cron_expression: '0 * * * *' }),
  ],
}

export const mockWorkflows = {
  single: createMockWorkflow({ name: 'My Workflow' }),
  list: [
    createMockWorkflow({ name: 'Email Notification', status: 'published' }),
    createMockWorkflow({ name: 'Data Sync', status: 'draft' }),
    createMockWorkflow({ name: 'Report Generator', status: 'archived' }),
  ],
}

export const mockExecutions = {
  running: createMockExecution({ status: 'running', completed_at: null }),
  completed: createMockExecution({ status: 'completed' }),
  failed: createMockExecution({ status: 'failed', error: 'Connection timeout' }),
}

// Response wrappers (matching backend API format)
export function wrapData<T>(data: T): { data: T } {
  return { data }
}

export function wrapPaginated<T>(
  data: T[],
  limit = 10,
  offset = 0,
  total?: number
): { data: T[]; limit: number; offset: number; total: number } {
  return {
    data,
    limit,
    offset,
    total: total ?? data.length,
  }
}

export function wrapError(
  message: string,
  code = 'error'
): { error: string; code: string } {
  return { error: message, code }
}

// Reset counter between tests
export function resetMockIds(): void {
  idCounter = 0
}
