import { useState, useEffect } from 'react'
import { webhookAPI } from '../../api/webhooks'
import type {
  FilterOperator,
  WebhookFilterCreateInput,
  WebhookFilterUpdateInput,
  TestFilterResult,
} from '../../api/webhooks'

interface FilterBuilderProps {
  webhookId: string
}

interface FilterRule {
  id?: string
  fieldPath: string
  operator: FilterOperator
  value: string
  logicGroup: number
  enabled: boolean
  isNew?: boolean
  isDirty?: boolean
}

interface ValidationError {
  filterId: string
  field: string
  message: string
}

const OPERATORS: { value: FilterOperator; label: string; description: string }[] = [
  { value: 'equals', label: 'Equals', description: 'Exact match' },
  { value: 'not_equals', label: 'Not Equals', description: 'Does not match' },
  { value: 'contains', label: 'Contains', description: 'String contains substring' },
  { value: 'not_contains', label: 'Not Contains', description: 'String does not contain' },
  { value: 'starts_with', label: 'Starts With', description: 'String starts with prefix' },
  { value: 'ends_with', label: 'Ends With', description: 'String ends with suffix' },
  { value: 'regex', label: 'Regex', description: 'Matches regular expression' },
  { value: 'gt', label: 'Greater Than', description: 'Number > value' },
  { value: 'gte', label: 'Greater Than or Equal', description: 'Number >= value' },
  { value: 'lt', label: 'Less Than', description: 'Number < value' },
  { value: 'lte', label: 'Less Than or Equal', description: 'Number <= value' },
  { value: 'in', label: 'In', description: 'Value in array' },
  { value: 'not_in', label: 'Not In', description: 'Value not in array' },
  { value: 'exists', label: 'Exists', description: 'Field exists in payload' },
  { value: 'not_exists', label: 'Not Exists', description: 'Field does not exist' },
]

const parseValue = (value: unknown): string => {
  if (typeof value === 'string') return value
  if (typeof value === 'number') return String(value)
  if (typeof value === 'boolean') return String(value)
  if (Array.isArray(value)) return JSON.stringify(value)
  if (value === null || value === undefined) return ''
  return JSON.stringify(value)
}

const serializeValue = (value: string, operator: FilterOperator): unknown => {
  if (operator === 'exists' || operator === 'not_exists') {
    return null
  }

  if (operator === 'in' || operator === 'not_in') {
    try {
      return JSON.parse(value)
    } catch {
      return value.split(',').map(v => v.trim())
    }
  }

  if (
    operator === 'gt' ||
    operator === 'gte' ||
    operator === 'lt' ||
    operator === 'lte'
  ) {
    const num = parseFloat(value)
    return isNaN(num) ? value : num
  }

  return value
}

const validateFieldPath = (path: string): string | null => {
  if (!path.trim()) {
    return 'Field path is required'
  }

  if (!path.startsWith('$')) {
    return 'Field path must start with $ (e.g., $.data.status)'
  }

  return null
}

const validateRegex = (pattern: string): string | null => {
  try {
    new RegExp(pattern)
    return null
  } catch {
    return 'Invalid regex pattern'
  }
}

const validateFilter = (filter: FilterRule): ValidationError[] => {
  const errors: ValidationError[] = []
  const filterId = filter.id || 'new'

  const pathError = validateFieldPath(filter.fieldPath)
  if (pathError) {
    errors.push({ filterId, field: 'fieldPath', message: pathError })
  }

  if (filter.operator === 'regex') {
    const regexError = validateRegex(filter.value)
    if (regexError) {
      errors.push({ filterId, field: 'value', message: regexError })
    }
  }

  if (
    filter.operator !== 'exists' &&
    filter.operator !== 'not_exists' &&
    !filter.value.trim()
  ) {
    errors.push({ filterId, field: 'value', message: 'Value is required' })
  }

  return errors
}

export default function FilterBuilder({ webhookId }: FilterBuilderProps) {
  const [filters, setFilters] = useState<FilterRule[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [validationErrors, setValidationErrors] = useState<ValidationError[]>([])
  const [showTestPanel, setShowTestPanel] = useState(false)
  const [testPayload, setTestPayload] = useState('')
  const [testResult, setTestResult] = useState<TestFilterResult | null>(null)
  const [testError, setTestError] = useState<string | null>(null)
  const [testLoading, setTestLoading] = useState(false)

  useEffect(() => {
    loadFilters()
  }, [webhookId])

  const loadFilters = async () => {
    try {
      setLoading(true)
      setError(null)
      const response = await webhookAPI.getFilters(webhookId)
      const loadedFilters: FilterRule[] = response.filters.map(f => ({
        id: f.id,
        fieldPath: f.fieldPath,
        operator: f.operator,
        value: parseValue(f.value),
        logicGroup: f.logicGroup,
        enabled: f.enabled,
      }))
      setFilters(loadedFilters)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to fetch filters')
    } finally {
      setLoading(false)
    }
  }

  const addFilter = () => {
    const newFilter: FilterRule = {
      fieldPath: '',
      operator: 'equals',
      value: '',
      logicGroup: 0,
      enabled: true,
      isNew: true,
    }
    setFilters([...filters, newFilter])
  }

  const updateFilter = (index: number, updates: Partial<FilterRule>) => {
    const updatedFilters = [...filters]
    updatedFilters[index] = {
      ...updatedFilters[index],
      ...updates,
      isDirty: true,
    }
    setFilters(updatedFilters)

    setValidationErrors(prevErrors =>
      prevErrors.filter(e => e.filterId !== updatedFilters[index].id)
    )
  }

  const saveFilter = async (index: number) => {
    const filter = filters[index]
    const errors = validateFilter(filter)

    if (errors.length > 0) {
      setValidationErrors([...validationErrors, ...errors])
      return
    }

    try {
      const serializedValue = serializeValue(filter.value, filter.operator)

      if (filter.isNew) {
        const input: WebhookFilterCreateInput = {
          fieldPath: filter.fieldPath,
          operator: filter.operator,
          value: serializedValue,
          logicGroup: filter.logicGroup,
          enabled: filter.enabled,
        }
        const created = await webhookAPI.createFilter(webhookId, input)

        const updatedFilters = [...filters]
        updatedFilters[index] = {
          id: created.id,
          fieldPath: created.fieldPath,
          operator: created.operator,
          value: parseValue(created.value),
          logicGroup: created.logicGroup,
          enabled: created.enabled,
        }
        setFilters(updatedFilters)
      } else if (filter.id) {
        const updates: WebhookFilterUpdateInput = {
          fieldPath: filter.fieldPath,
          operator: filter.operator,
          value: serializedValue,
          logicGroup: filter.logicGroup,
          enabled: filter.enabled,
        }
        await webhookAPI.updateFilter(webhookId, filter.id, updates)

        const updatedFilters = [...filters]
        updatedFilters[index] = { ...filter, isDirty: false }
        setFilters(updatedFilters)
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to save filter')
    }
  }

  const deleteFilter = async (index: number) => {
    const filter = filters[index]

    if (filter.id) {
      try {
        await webhookAPI.deleteFilter(webhookId, filter.id)
      } catch (err) {
        setError(err instanceof Error ? err.message : 'Failed to delete filter')
        return
      }
    }

    setFilters(filters.filter((_, i) => i !== index))
    setValidationErrors(prevErrors =>
      prevErrors.filter(e => e.filterId !== filter.id)
    )
  }

  const toggleEnabled = async (index: number) => {
    const filter = filters[index]

    if (!filter.id) {
      updateFilter(index, { enabled: !filter.enabled })
      return
    }

    try {
      const updated = await webhookAPI.updateFilter(webhookId, filter.id, {
        enabled: !filter.enabled,
      })

      const updatedFilters = [...filters]
      updatedFilters[index] = {
        ...filter,
        enabled: updated.enabled,
      }
      setFilters(updatedFilters)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to update filter')
    }
  }

  const runTest = async () => {
    setTestError(null)
    setTestResult(null)

    if (!testPayload.trim()) {
      setTestError('Test payload is required')
      return
    }

    let payload: unknown
    try {
      payload = JSON.parse(testPayload)
    } catch {
      setTestError('Invalid JSON format')
      return
    }

    try {
      setTestLoading(true)
      const result = await webhookAPI.testFilters(webhookId, {
        filters: filters.map(f => ({
          fieldPath: f.fieldPath,
          operator: f.operator,
          value: serializeValue(f.value, f.operator),
          logicGroup: f.logicGroup,
          enabled: f.enabled,
        })),
        payload,
      })
      setTestResult(result)
    } catch (err) {
      setTestError(err instanceof Error ? err.message : 'Test failed')
    } finally {
      setTestLoading(false)
    }
  }

  const getFilterError = (filterId: string | undefined, field: string): string | null => {
    if (!filterId) {
      const error = validationErrors.find(e => e.filterId === 'new' && e.field === field)
      return error?.message || null
    }
    const error = validationErrors.find(e => e.filterId === filterId && e.field === field)
    return error?.message || null
  }

  const groupFilters = () => {
    const grouped = new Map<number, FilterRule[]>()
    filters.forEach(filter => {
      const group = grouped.get(filter.logicGroup) || []
      group.push(filter)
      grouped.set(filter.logicGroup, group)
    })
    return Array.from(grouped.entries()).sort((a, b) => a[0] - b[0])
  }

  if (loading) {
    return (
      <div className="bg-gray-800 border border-gray-700 rounded-lg p-6">
        <p className="text-gray-400">Loading filters...</p>
      </div>
    )
  }

  if (error) {
    return (
      <div className="bg-gray-800 border border-gray-700 rounded-lg p-6">
        <p className="text-red-400">{error}</p>
        <button
          onClick={loadFilters}
          className="mt-4 px-4 py-2 bg-primary-600 text-white rounded-lg text-sm hover:bg-primary-700"
        >
          Retry
        </button>
      </div>
    )
  }

  const groupedFilters = groupFilters()

  return (
    <div className="bg-gray-800 border border-gray-700 rounded-lg p-6 space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-white text-xl font-semibold">Webhook Filters</h2>
          <p className="text-gray-400 text-sm mt-1">
            Filter incoming webhook payloads based on field values
          </p>
        </div>
        <button
          onClick={addFilter}
          className="px-4 py-2 bg-primary-600 text-white rounded-lg text-sm hover:bg-primary-700"
        >
          Add Filter
        </button>
      </div>

      {filters.length === 0 ? (
        <div className="text-center py-8">
          <p className="text-gray-400">No filters configured</p>
          <p className="text-gray-500 text-sm mt-1">
            Click "Add Filter" to create your first filter rule
          </p>
        </div>
      ) : (
        <div className="space-y-4">
          {groupedFilters.map(([groupId, groupFilters], groupIndex) => (
            <div key={groupId} className="space-y-2">
              {groupIndex > 0 && (
                <div className="flex items-center gap-2 py-2">
                  <div className="flex-1 border-t border-gray-600"></div>
                  <span className="text-primary-400 text-sm font-medium px-3 py-1 bg-gray-700 rounded">
                    OR
                  </span>
                  <div className="flex-1 border-t border-gray-600"></div>
                </div>
              )}

              {groupFilters.map((filter, filterIndex) => {
                const globalIndex = filters.indexOf(filter)
                const isFirstInGroup = filterIndex === 0
                const fieldPathError = getFilterError(filter.id, 'fieldPath')
                const valueError = getFilterError(filter.id, 'value')

                return (
                  <div key={filter.id || `new-${globalIndex}`}>
                    {!isFirstInGroup && (
                      <div className="flex items-center gap-2 py-1 pl-4">
                        <span className="text-gray-400 text-sm font-medium">AND</span>
                      </div>
                    )}

                    <div className="bg-gray-700 rounded-lg p-4 space-y-3">
                      <div className="flex items-start gap-3">
                        <div className="flex-1 grid grid-cols-1 md:grid-cols-3 gap-3">
                          <div>
                            <label className="block text-sm font-medium text-gray-300 mb-1">
                              Field Path
                            </label>
                            <input
                              type="text"
                              value={filter.fieldPath}
                              onChange={e =>
                                updateFilter(globalIndex, { fieldPath: e.target.value })
                              }
                              placeholder="$.data.status"
                              className={`w-full px-3 py-2 bg-gray-600 text-white rounded text-sm focus:outline-none focus:ring-2 ${
                                fieldPathError
                                  ? 'ring-2 ring-red-500'
                                  : 'focus:ring-primary-500'
                              }`}
                            />
                            {fieldPathError && (
                              <p className="text-xs text-red-400 mt-1">{fieldPathError}</p>
                            )}
                          </div>

                          <div>
                            <label
                              htmlFor={`operator-${globalIndex}`}
                              className="block text-sm font-medium text-gray-300 mb-1"
                            >
                              Operator
                            </label>
                            <select
                              id={`operator-${globalIndex}`}
                              value={filter.operator}
                              onChange={e =>
                                updateFilter(globalIndex, {
                                  operator: e.target.value as FilterOperator,
                                })
                              }
                              className="w-full px-3 py-2 bg-gray-600 text-white rounded text-sm focus:outline-none focus:ring-2 focus:ring-primary-500"
                              aria-label="Operator"
                            >
                              {OPERATORS.map(op => (
                                <option key={op.value} value={op.value}>
                                  {op.label}
                                </option>
                              ))}
                            </select>
                          </div>

                          <div>
                            <label className="block text-sm font-medium text-gray-300 mb-1">
                              Value
                            </label>
                            <input
                              type="text"
                              value={filter.value}
                              onChange={e =>
                                updateFilter(globalIndex, { value: e.target.value })
                              }
                              placeholder="Value"
                              disabled={
                                filter.operator === 'exists' ||
                                filter.operator === 'not_exists'
                              }
                              className={`w-full px-3 py-2 bg-gray-600 text-white rounded text-sm focus:outline-none focus:ring-2 ${
                                valueError
                                  ? 'ring-2 ring-red-500'
                                  : 'focus:ring-primary-500'
                              } disabled:opacity-50`}
                            />
                            {valueError && (
                              <p className="text-xs text-red-400 mt-1">{valueError}</p>
                            )}
                          </div>
                        </div>

                        <div className="flex flex-col gap-2 pt-6">
                          <button
                            onClick={() => saveFilter(globalIndex)}
                            className="px-3 py-2 bg-green-600 text-white rounded text-sm hover:bg-green-700 whitespace-nowrap"
                            aria-label="Save"
                          >
                            Save
                          </button>
                          <button
                            onClick={() => deleteFilter(globalIndex)}
                            className="px-3 py-2 bg-red-600 text-white rounded text-sm hover:bg-red-700 whitespace-nowrap"
                            aria-label="Delete"
                          >
                            Delete
                          </button>
                        </div>
                      </div>

                      <div className="flex items-center gap-4 pt-2 border-t border-gray-600">
                        <label className="flex items-center gap-2">
                          <input
                            type="checkbox"
                            checked={filter.enabled}
                            onChange={() => toggleEnabled(globalIndex)}
                            className="w-4 h-4 text-primary-600 bg-gray-600 border-gray-500 rounded focus:ring-primary-500"
                            aria-label="Enabled"
                          />
                          <span className="text-sm text-gray-300">Enabled</span>
                        </label>

                        <div className="flex items-center gap-2">
                          <label
                            htmlFor={`logic-group-${globalIndex}`}
                            className="text-sm text-gray-300"
                          >
                            Logic Group:
                          </label>
                          <input
                            id={`logic-group-${globalIndex}`}
                            type="number"
                            min="0"
                            value={filter.logicGroup}
                            onChange={e =>
                              updateFilter(globalIndex, {
                                logicGroup: parseInt(e.target.value) || 0,
                              })
                            }
                            className="w-16 px-2 py-1 bg-gray-600 text-white rounded text-sm focus:outline-none focus:ring-2 focus:ring-primary-500"
                            aria-label="Logic Group"
                          />
                        </div>
                      </div>
                    </div>
                  </div>
                )
              })}
            </div>
          ))}
        </div>
      )}

      <div className="pt-4 border-t border-gray-700">
        <button
          onClick={() => setShowTestPanel(!showTestPanel)}
          className="w-full px-4 py-2 bg-gray-700 text-white rounded-lg text-sm hover:bg-gray-600"
        >
          {showTestPanel ? 'Hide' : 'Test Filters'}
        </button>

        {showTestPanel && (
          <div className="mt-4 space-y-4">
            <div>
              <label htmlFor="test-payload" className="block text-sm font-medium text-gray-300 mb-2">
                Test Payload (JSON)
              </label>
              <textarea
                id="test-payload"
                value={testPayload}
                onChange={e => setTestPayload(e.target.value)}
                placeholder='{"data": {"status": "active", "amount": 150}}'
                rows={6}
                className="w-full px-3 py-2 bg-gray-700 text-white rounded text-sm focus:outline-none focus:ring-2 focus:ring-primary-500 font-mono"
              />
            </div>

            <button
              onClick={runTest}
              disabled={testLoading}
              className="w-full px-4 py-2 bg-primary-600 text-white rounded-lg text-sm hover:bg-primary-700 disabled:opacity-50"
            >
              {testLoading ? 'Testing...' : 'Run Test'}
            </button>

            {testError && (
              <div className="p-3 bg-red-900/20 border border-red-500 rounded text-sm">
                <p className="text-red-400">{testError}</p>
              </div>
            )}

            {testResult && (
              <div
                className={`p-4 rounded border ${
                  testResult.passed
                    ? 'bg-green-900/20 border-green-500'
                    : 'bg-red-900/20 border-red-500'
                }`}
              >
                <div className="flex items-center gap-2 mb-2">
                  <span
                    className={`font-semibold ${
                      testResult.passed ? 'text-green-400' : 'text-red-400'
                    }`}
                  >
                    {testResult.passed ? 'Passed' : 'Failed'}
                  </span>
                </div>
                <p className="text-gray-300 text-sm">{testResult.reason}</p>
                {Object.keys(testResult.details).length > 0 && (
                  <pre className="mt-2 p-2 bg-gray-900 rounded text-xs text-gray-400 overflow-auto">
                    {JSON.stringify(testResult.details, null, 2)}
                  </pre>
                )}
              </div>
            )}
          </div>
        )}
      </div>

      <div className="text-xs text-gray-500 space-y-1">
        <p>
          <strong>Logic Groups:</strong> Filters with the same group number are combined with AND
          logic. Different groups are combined with OR logic.
        </p>
        <p>
          <strong>Example:</strong> Group 0: status=active AND amount&gt;100, Group 1:
          type=premium (Either condition passes)
        </p>
      </div>
    </div>
  )
}
