import { useState } from 'react'
import { webhookAPI } from '../../api/webhooks'
import type { TestWebhookInput, TestWebhookResponse } from '../../api/webhooks'
import ResponseViewer from './ResponseViewer'
import HistoryList from './HistoryList'

interface WebhookTestPanelProps {
  webhookId: string
}

interface HeaderPair {
  id: string
  key: string
  value: string
}

interface TestHistory {
  id: string
  timestamp: Date
  method: string
  headers: Record<string, string>
  body: unknown
  response: TestWebhookResponse | null
  error: string | null
}

const createSampleTemplates = () => ({
  none: { label: 'None', payload: '' },
  simple: {
    label: 'Simple Object',
    payload: JSON.stringify({ message: 'Hello, World!', timestamp: Date.now() }, null, 2),
  },
  github: {
    label: 'GitHub Push',
    payload: JSON.stringify(
      {
        ref: 'refs/heads/main',
        repository: {
          name: 'my-repo',
          full_name: 'user/my-repo',
        },
        pusher: {
          name: 'user',
          email: 'user@example.com',
        },
        commits: [
          {
            id: 'abc123',
            message: 'Update README',
            author: { name: 'user', email: 'user@example.com' },
          },
        ],
      },
      null,
      2
    ),
  },
  stripe: {
    label: 'Stripe Payment',
    payload: JSON.stringify(
      {
        id: 'evt_123456',
        type: 'payment_intent.succeeded',
        data: {
          object: {
            id: 'pi_123456',
            amount: 2000,
            currency: 'usd',
            status: 'succeeded',
          },
        },
      },
      null,
      2
    ),
  },
})

const SAMPLE_TEMPLATES = createSampleTemplates()

const getStatusCodeColor = (statusCode: number): string => {
  if (statusCode >= 200 && statusCode < 300) {
    return 'text-green-400'
  }
  return 'text-red-400'
}

const formatTimestamp = (date: Date): string => {
  const now = new Date()
  const diffMs = now.getTime() - date.getTime()
  const diffSecs = Math.floor(diffMs / 1000)
  const diffMins = Math.floor(diffSecs / 60)
  const diffHours = Math.floor(diffMins / 60)

  if (diffSecs < 10) return 'just now'
  if (diffSecs < 60) return `${diffSecs} seconds ago`
  if (diffMins === 1) return '1 minute ago'
  if (diffMins < 60) return `${diffMins} minutes ago`
  if (diffHours === 1) return '1 hour ago'
  if (diffHours < 24) return `${diffHours} hours ago`
  return date.toLocaleString()
}

export default function WebhookTestPanel({ webhookId }: WebhookTestPanelProps) {
  const [method, setMethod] = useState('POST')
  const [payload, setPayload] = useState('')
  const [headers, setHeaders] = useState<HeaderPair[]>([])
  const [isLoading, setIsLoading] = useState(false)
  const [response, setResponse] = useState<TestWebhookResponse | null>(null)
  const [error, setError] = useState<string | null>(null)
  const [validationError, setValidationError] = useState<string | null>(null)
  const [history, setHistory] = useState<TestHistory[]>([])

  const addHeader = () => {
    setHeaders([...headers, { id: Date.now().toString(), key: '', value: '' }])
  }

  const removeHeader = (id: string) => {
    setHeaders(headers.filter(h => h.id !== id))
  }

  const updateHeader = (id: string, field: 'key' | 'value', value: string) => {
    setHeaders(headers.map(h => (h.id === id ? { ...h, [field]: value } : h)))
  }

  const handleTemplateSelect = (templateKey: string) => {
    const template = SAMPLE_TEMPLATES[templateKey as keyof typeof SAMPLE_TEMPLATES]
    if (template) {
      setPayload(template.payload)
    }
  }

  const validatePayload = (): boolean => {
    if (!payload.trim()) {
      setValidationError(null)
      return true
    }

    try {
      JSON.parse(payload)
      setValidationError(null)
      return true
    } catch {
      setValidationError('Invalid JSON format')
      return false
    }
  }

  const buildHeaders = (): Record<string, string> => {
    const headerObj: Record<string, string> = {}
    headers.forEach(h => {
      if (h.key.trim()) {
        headerObj[h.key] = h.value
      }
    })
    return headerObj
  }

  const createTestInput = (): TestWebhookInput => {
    const input: TestWebhookInput = { method }

    const headerObj = buildHeaders()
    if (Object.keys(headerObj).length > 0) {
      input.headers = headerObj
    }

    if (payload.trim()) {
      input.body = JSON.parse(payload)
    }

    return input
  }

  const createHistoryItem = (
    result: TestWebhookResponse | null,
    err: string | null,
    input: TestWebhookInput
  ): TestHistory => ({
    id: Date.now().toString(),
    timestamp: new Date(),
    method,
    headers: input.headers || {},
    body: input.body || null,
    response: result,
    error: err,
  })

  const addToHistory = (item: TestHistory) => {
    setHistory([item, ...history].slice(0, 5))
  }

  const handleSendTest = async () => {
    setValidationError(null)
    setError(null)

    if (!validatePayload()) {
      return
    }

    setIsLoading(true)

    try {
      const input = createTestInput()
      const result = await webhookAPI.test(webhookId, input)

      setResponse(result)
      setError(null)
      addToHistory(createHistoryItem(result, null, input))
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Unknown error occurred'
      setError(errorMessage)
      setResponse(null)

      const input = createTestInput()
      addToHistory(createHistoryItem(null, errorMessage, input))
    } finally {
      setIsLoading(false)
    }
  }

  const loadHistoryItem = (item: TestHistory) => {
    setMethod(item.method)
    setPayload(item.body ? JSON.stringify(item.body, null, 2) : '')

    const headerPairs: HeaderPair[] = Object.entries(item.headers).map(([key, value], index) => ({
      id: `${Date.now()}-${index}`,
      key,
      value,
    }))
    setHeaders(headerPairs)

    setResponse(item.response)
    setError(item.error)
  }

  return (
    <div className="bg-gray-800 border border-gray-700 rounded-lg p-6 space-y-6">
      {/* Header */}
      <div>
        <h2 className="text-white text-xl font-semibold">Webhook Test</h2>
        <p className="text-gray-400 text-sm mt-1">
          Test your webhook with custom payloads and headers
        </p>
      </div>

      {/* Configuration Form */}
      <div className="space-y-4">
        {/* HTTP Method */}
        <div>
          <label htmlFor="http-method" className="block text-sm font-medium text-gray-300 mb-2">
            HTTP Method
          </label>
          <select
            id="http-method"
            value={method}
            onChange={e => setMethod(e.target.value)}
            disabled={isLoading}
            className="w-full px-3 py-2 bg-gray-700 text-white rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-primary-500 disabled:opacity-50"
          >
            <option value="GET">GET</option>
            <option value="POST">POST</option>
            <option value="PUT">PUT</option>
            <option value="DELETE">DELETE</option>
            <option value="PATCH">PATCH</option>
          </select>
        </div>

        {/* Sample Payload Template */}
        <div>
          <label
            htmlFor="sample-payload"
            className="block text-sm font-medium text-gray-300 mb-2"
          >
            Sample Payload
          </label>
          <select
            id="sample-payload"
            onChange={e => handleTemplateSelect(e.target.value)}
            disabled={isLoading}
            className="w-full px-3 py-2 bg-gray-700 text-white rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-primary-500 disabled:opacity-50"
          >
            <option value="none">None</option>
            <option value="simple">Simple Object</option>
            <option value="github">GitHub Push</option>
            <option value="stripe">Stripe Payment</option>
          </select>
        </div>

        {/* JSON Payload Editor */}
        <div>
          <label htmlFor="payload" className="block text-sm font-medium text-gray-300 mb-2">
            Payload (JSON)
          </label>
          <textarea
            id="payload"
            value={payload}
            onChange={e => {
              setPayload(e.target.value)
              setValidationError(null)
            }}
            disabled={isLoading}
            placeholder='{"key": "value"}'
            rows={8}
            className="w-full px-3 py-2 bg-gray-700 text-white rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-primary-500 font-mono disabled:opacity-50"
          />
          {validationError && (
            <div className="mt-2 text-xs text-red-400">{validationError}</div>
          )}
        </div>

        {/* Custom Headers */}
        <div>
          <div className="flex items-center justify-between mb-2">
            <label className="block text-sm font-medium text-gray-300">Headers</label>
            <button
              onClick={addHeader}
              disabled={isLoading}
              className="text-sm text-primary-400 hover:text-primary-300 disabled:opacity-50"
              aria-label="Add header"
            >
              + Add Header
            </button>
          </div>
          {headers.length > 0 && (
            <div className="space-y-2">
              {headers.map(header => (
                <div key={header.id} className="flex gap-2">
                  <input
                    type="text"
                    value={header.key}
                    onChange={e => updateHeader(header.id, 'key', e.target.value)}
                    disabled={isLoading}
                    placeholder="Header name"
                    className="flex-1 px-3 py-2 bg-gray-700 text-white rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-primary-500 disabled:opacity-50"
                  />
                  <input
                    type="text"
                    value={header.value}
                    onChange={e => updateHeader(header.id, 'value', e.target.value)}
                    disabled={isLoading}
                    placeholder="Header value"
                    className="flex-1 px-3 py-2 bg-gray-700 text-white rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-primary-500 disabled:opacity-50"
                  />
                  <button
                    onClick={() => removeHeader(header.id)}
                    disabled={isLoading}
                    className="px-3 py-2 bg-gray-600 text-gray-300 rounded-lg text-sm hover:bg-gray-500 disabled:opacity-50"
                    aria-label="Remove header"
                  >
                    ✕
                  </button>
                </div>
              ))}
            </div>
          )}
        </div>

        {/* Send Button */}
        <button
          onClick={handleSendTest}
          disabled={isLoading}
          className="w-full px-4 py-2 bg-primary-600 text-white rounded-lg text-sm font-medium hover:bg-primary-700 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
        >
          {isLoading ? 'Sending...' : 'Send Test Request'}
        </button>
      </div>

      {/* Response Viewer */}
      <ResponseViewer
        response={response}
        error={error}
        getStatusCodeColor={getStatusCodeColor}
      />

      {/* History */}
      <HistoryList
        history={history}
        onSelectItem={loadHistoryItem}
        getStatusCodeColor={getStatusCodeColor}
        formatTimestamp={formatTimestamp}
      />
    </div>
  )
}
