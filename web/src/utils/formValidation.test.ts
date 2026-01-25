import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import {
  validateCronExpression,
  validateTimezone,
  validateRequired,
  validateScheduleForm,
  validateWorkflowForm,
  createAsyncValidator,
  validateCronExpressionAsync,
} from './formValidation'

describe('validateCronExpression', () => {
  describe('valid expressions', () => {
    const validCrons = [
      { cron: '* * * * *', desc: 'every minute' },
      { cron: '0 * * * *', desc: 'every hour' },
      { cron: '0 0 * * *', desc: 'every day at midnight' },
      { cron: '0 9 * * *', desc: 'every day at 9 AM' },
      { cron: '0 9 * * 1', desc: 'every Monday at 9 AM' },
      { cron: '0 0 1 * *', desc: 'first of every month' },
      { cron: '*/5 * * * *', desc: 'every 5 minutes' },
      { cron: '0 9-17 * * *', desc: 'every hour 9 AM to 5 PM' },
      { cron: '0 9 * * 1-5', desc: 'weekdays at 9 AM' },
      { cron: '0 0 * * 0,6', desc: 'weekends at midnight' },
      { cron: '0 9 * * MON-FRI', desc: 'weekdays using names' },
      { cron: '0 0 1 JAN *', desc: 'first of January using names' },
      { cron: '30 4 1,15 * *', desc: '4:30 AM on 1st and 15th' },
      { cron: '0 0 * * *', desc: 'midnight' },
      // 6-part cron with seconds
      { cron: '0 * * * * *', desc: 'every minute at second 0' },
      { cron: '0 0 9 * * *', desc: 'daily at 9 AM with seconds' },
      { cron: '*/30 * * * * *', desc: 'every 30 seconds' },
    ]

    validCrons.forEach(({ cron, desc }) => {
      it(`should accept "${cron}" (${desc})`, () => {
        const result = validateCronExpression(cron)
        expect(result.valid).toBe(true)
        expect(result.errors).toHaveLength(0)
      })
    })
  })

  describe('invalid expressions', () => {
    it('should reject empty string', () => {
      const result = validateCronExpression('')
      expect(result.valid).toBe(false)
      expect(result.errors[0].code).toBe('required')
    })

    it('should reject whitespace only', () => {
      const result = validateCronExpression('   ')
      expect(result.valid).toBe(false)
      expect(result.errors[0].code).toBe('required')
    })

    it('should reject too few parts', () => {
      const result = validateCronExpression('* * *')
      expect(result.valid).toBe(false)
      expect(result.errors[0].code).toBe('invalid_cron_format')
      expect(result.errors[0].message).toContain('5 or 6 parts')
    })

    it('should reject too many parts', () => {
      const result = validateCronExpression('* * * * * * *')
      expect(result.valid).toBe(false)
      expect(result.errors[0].code).toBe('invalid_cron_format')
    })

    it('should reject invalid minute value', () => {
      const result = validateCronExpression('60 * * * *')
      expect(result.valid).toBe(false)
      expect(result.errors[0].message).toContain('minute')
    })

    it('should reject invalid hour value', () => {
      const result = validateCronExpression('0 24 * * *')
      expect(result.valid).toBe(false)
      expect(result.errors[0].message).toContain('hour')
    })

    it('should reject invalid day of month', () => {
      const result = validateCronExpression('0 0 32 * *')
      expect(result.valid).toBe(false)
      expect(result.errors[0].message).toContain('day of month')
    })

    it('should reject invalid month', () => {
      const result = validateCronExpression('0 0 * 13 *')
      expect(result.valid).toBe(false)
      expect(result.errors[0].message).toContain('month')
    })

    it('should reject invalid day of week', () => {
      const result = validateCronExpression('0 0 * * 7')
      expect(result.valid).toBe(false)
      expect(result.errors[0].message).toContain('day of week')
    })

    it('should reject invalid step value', () => {
      const result = validateCronExpression('*/0 * * * *')
      expect(result.valid).toBe(false)
      expect(result.errors[0].code).toBe('invalid_cron_step')
    })

    it('should reject invalid range (start > end)', () => {
      const result = validateCronExpression('0 17-9 * * *')
      expect(result.valid).toBe(false)
      expect(result.errors[0].code).toBe('invalid_cron_range')
    })

    it('should reject invalid range values', () => {
      const result = validateCronExpression('0 25-30 * * *')
      expect(result.valid).toBe(false)
      expect(result.errors[0].code).toBe('invalid_cron_range')
    })

    it('should reject invalid list value', () => {
      const result = validateCronExpression('0 0 * * 1,8')
      expect(result.valid).toBe(false)
      expect(result.errors[0].message).toContain('day of week')
    })

    it('should reject gibberish', () => {
      const result = validateCronExpression('abc def ghi jkl mno')
      expect(result.valid).toBe(false)
    })

    it('should reject invalid month name', () => {
      const result = validateCronExpression('0 0 * FOO *')
      expect(result.valid).toBe(false)
    })

    it('should reject invalid day name', () => {
      const result = validateCronExpression('0 0 * * BAR')
      expect(result.valid).toBe(false)
    })
  })

  it('should use custom field name', () => {
    const result = validateCronExpression('', 'schedule')
    expect(result.errors[0].field).toBe('schedule')
  })
})

describe('validateTimezone', () => {
  describe('valid timezones', () => {
    const validTimezones = [
      'UTC',
      'America/New_York',
      'America/Los_Angeles',
      'Europe/London',
      'Asia/Tokyo',
      'Australia/Sydney',
    ]

    validTimezones.forEach((tz) => {
      it(`should accept "${tz}"`, () => {
        const result = validateTimezone(tz)
        expect(result.valid).toBe(true)
      })
    })
  })

  it('should accept empty string (optional field)', () => {
    const result = validateTimezone('')
    expect(result.valid).toBe(true)
  })

  it('should reject invalid timezone', () => {
    const result = validateTimezone('Invalid/Timezone')
    expect(result.valid).toBe(false)
    expect(result.errors[0].code).toBe('invalid_timezone')
  })

  it('should reject random string', () => {
    const result = validateTimezone('foobar')
    expect(result.valid).toBe(false)
  })

  it('should use custom field name', () => {
    const result = validateTimezone('invalid', 'tz')
    expect(result.errors[0].field).toBe('tz')
  })
})

describe('validateRequired', () => {
  it('should reject null', () => {
    const result = validateRequired(null, 'field')
    expect(result.valid).toBe(false)
    expect(result.errors[0].code).toBe('required')
  })

  it('should reject undefined', () => {
    const result = validateRequired(undefined, 'field')
    expect(result.valid).toBe(false)
  })

  it('should reject empty string', () => {
    const result = validateRequired('', 'field')
    expect(result.valid).toBe(false)
  })

  it('should reject whitespace only string', () => {
    const result = validateRequired('   ', 'field')
    expect(result.valid).toBe(false)
  })

  it('should accept non-empty string', () => {
    const result = validateRequired('value', 'field')
    expect(result.valid).toBe(true)
  })

  it('should accept number', () => {
    const result = validateRequired(0, 'field')
    expect(result.valid).toBe(true)
  })

  it('should accept boolean false', () => {
    const result = validateRequired(false, 'field')
    expect(result.valid).toBe(true)
  })

  it('should accept empty array', () => {
    const result = validateRequired([], 'field')
    expect(result.valid).toBe(true)
  })

  it('should include field name in error message', () => {
    const result = validateRequired(null, 'myField')
    expect(result.errors[0].message).toContain('myField')
  })
})

describe('validateScheduleForm', () => {
  it('should accept valid form data', () => {
    const result = validateScheduleForm({
      name: 'My Schedule',
      cron_expression: '0 9 * * *',
      timezone: 'UTC',
      enabled: true,
    })
    expect(result.valid).toBe(true)
  })

  it('should accept form data without optional fields', () => {
    const result = validateScheduleForm({
      name: 'My Schedule',
      cron_expression: '0 9 * * *',
    })
    expect(result.valid).toBe(true)
  })

  it('should reject empty name', () => {
    const result = validateScheduleForm({
      name: '',
      cron_expression: '0 9 * * *',
    })
    expect(result.valid).toBe(false)
    expect(result.errors.some((e) => e.field === 'name')).toBe(true)
  })

  it('should reject invalid cron expression', () => {
    const result = validateScheduleForm({
      name: 'My Schedule',
      cron_expression: 'invalid',
    })
    expect(result.valid).toBe(false)
    expect(result.errors.some((e) => e.field === 'cron_expression')).toBe(true)
  })

  it('should reject invalid timezone', () => {
    const result = validateScheduleForm({
      name: 'My Schedule',
      cron_expression: '0 9 * * *',
      timezone: 'Invalid/TZ',
    })
    expect(result.valid).toBe(false)
    expect(result.errors.some((e) => e.field === 'timezone')).toBe(true)
  })

  it('should collect multiple errors', () => {
    const result = validateScheduleForm({
      name: '',
      cron_expression: 'invalid',
      timezone: 'bad',
    })
    expect(result.valid).toBe(false)
    expect(result.errors.length).toBeGreaterThanOrEqual(2)
  })
})

describe('validateWorkflowForm', () => {
  it('should accept valid form data', () => {
    const result = validateWorkflowForm({
      name: 'My Workflow',
      description: 'A test workflow',
    })
    expect(result.valid).toBe(true)
  })

  it('should accept form data without description', () => {
    const result = validateWorkflowForm({
      name: 'My Workflow',
    })
    expect(result.valid).toBe(true)
  })

  it('should reject empty name', () => {
    const result = validateWorkflowForm({
      name: '',
    })
    expect(result.valid).toBe(false)
    expect(result.errors[0].field).toBe('name')
  })

  it('should reject name that is too long', () => {
    const result = validateWorkflowForm({
      name: 'a'.repeat(101),
    })
    expect(result.valid).toBe(false)
  })
})

describe('createAsyncValidator', () => {
  beforeEach(() => {
    vi.useFakeTimers()
  })

  afterEach(() => {
    vi.useRealTimers()
  })

  it('should debounce validation calls', async () => {
    const validateFn = vi.fn().mockResolvedValue({ valid: true, errors: [] })
    const validator = createAsyncValidator(validateFn, 100)

    // Call multiple times rapidly
    validator('a')
    validator('ab')
    validator('abc')

    // Fast-forward past debounce
    await vi.runAllTimersAsync()

    // Should only call validateFn once with the last value
    expect(validateFn).toHaveBeenCalledTimes(1)
    expect(validateFn).toHaveBeenCalledWith('abc')
  })

  it('should return validation result', async () => {
    const validateFn = vi.fn().mockResolvedValue({
      valid: false,
      errors: [{ field: 'test', message: 'error', code: 'error' }],
    })
    const validator = createAsyncValidator(validateFn, 100)

    const resultPromise = validator('test')
    await vi.runAllTimersAsync()
    const result = await resultPromise

    expect(result.valid).toBe(false)
    expect(result.errors).toHaveLength(1)
  })

  it('should handle validation errors', async () => {
    const validateFn = vi.fn().mockRejectedValue(new Error('Network error'))
    const validator = createAsyncValidator(validateFn, 100)

    const resultPromise = validator('test')
    await vi.runAllTimersAsync()
    const result = await resultPromise

    expect(result.valid).toBe(false)
    expect(result.errors[0].code).toBe('async_error')
    expect(result.errors[0].message).toBe('Network error')
  })

  it('should use default debounce time', async () => {
    const validateFn = vi.fn().mockResolvedValue({ valid: true, errors: [] })
    const validator = createAsyncValidator(validateFn) // default 300ms

    validator('test')

    // At 200ms, should not have called yet
    vi.advanceTimersByTime(200)
    expect(validateFn).not.toHaveBeenCalled()

    // At 300ms, should call
    vi.advanceTimersByTime(100)
    await vi.runAllTimersAsync()
    expect(validateFn).toHaveBeenCalledTimes(1)
  })
})

describe('validateCronExpressionAsync', () => {
  beforeEach(() => {
    vi.stubGlobal('fetch', vi.fn())
  })

  afterEach(() => {
    vi.unstubAllGlobals()
  })

  it('should return client-side errors without calling API', async () => {
    const result = await validateCronExpressionAsync('invalid')

    expect(fetch).not.toHaveBeenCalled()
    expect(result.valid).toBe(false)
  })

  it('should call API for valid client-side cron', async () => {
    vi.mocked(fetch).mockResolvedValue({
      ok: true,
      json: () => Promise.resolve({ valid: true }),
    } as Response)

    const result = await validateCronExpressionAsync('0 9 * * *')

    expect(fetch).toHaveBeenCalledWith(
      '/api/v1/schedules/parse-cron',
      expect.objectContaining({
        method: 'POST',
        body: JSON.stringify({ cron_expression: '0 9 * * *', timezone: 'UTC' }),
      })
    )
    expect(result.valid).toBe(true)
  })

  it('should use custom API endpoint', async () => {
    vi.mocked(fetch).mockResolvedValue({
      ok: true,
      json: () => Promise.resolve({ valid: true }),
    } as Response)

    await validateCronExpressionAsync('0 9 * * *', 'UTC', '/custom/endpoint')

    expect(fetch).toHaveBeenCalledWith(
      '/custom/endpoint',
      expect.anything()
    )
  })

  it('should return server error messages', async () => {
    vi.mocked(fetch).mockResolvedValue({
      ok: false,
      json: () => Promise.resolve({ error: 'Server says invalid' }),
    } as Response)

    const result = await validateCronExpressionAsync('0 9 * * *')

    expect(result.valid).toBe(false)
    expect(result.errors[0].message).toBe('Server says invalid')
    expect(result.errors[0].code).toBe('server_validation_error')
  })

  it('should fallback to client validation on network error', async () => {
    vi.mocked(fetch).mockRejectedValue(new Error('Network error'))

    const result = await validateCronExpressionAsync('0 9 * * *')

    // Should still pass because client-side validation passed
    expect(result.valid).toBe(true)
  })
})
