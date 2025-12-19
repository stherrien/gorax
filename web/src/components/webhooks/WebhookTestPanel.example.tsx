/**
 * WebhookTestPanel Usage Example
 *
 * This component provides a complete webhook testing interface for the Gorax platform.
 * It allows users to test webhooks with custom payloads, headers, and methods.
 *
 * Features:
 * - HTTP method selection (GET, POST, PUT, DELETE, PATCH)
 * - JSON payload editor with validation
 * - Custom header key-value pairs
 * - Sample payload templates (GitHub, Stripe, etc.)
 * - Response viewer with status code color coding
 * - Test history (last 5 tests)
 * - Link to execution details
 *
 * Usage:
 * ```tsx
 * import WebhookTestPanel from './components/webhooks/WebhookTestPanel'
 *
 * function MyWebhookPage() {
 *   const webhookId = 'wh-123' // Get from route params or props
 *
 *   return (
 *     <div className="container mx-auto p-6">
 *       <h1 className="text-2xl font-bold text-white mb-6">Test Webhook</h1>
 *       <WebhookTestPanel webhookId={webhookId} />
 *     </div>
 *   )
 * }
 * ```
 *
 * API Integration:
 * The component uses the webhookAPI.test() method which expects:
 * - webhookId: string - The ID of the webhook to test
 * - input: TestWebhookInput - Contains method, headers, and body
 *
 * Response format:
 * - success: boolean
 * - statusCode: number
 * - responseTimeMs: number
 * - executionId?: string
 * - error?: string
 */

import { BrowserRouter } from 'react-router-dom'
import WebhookTestPanel from './WebhookTestPanel'

export default function WebhookTestPanelExample() {
  return (
    <BrowserRouter>
      <div className="min-h-screen bg-gray-900 p-8">
        <div className="max-w-4xl mx-auto">
          <h1 className="text-3xl font-bold text-white mb-2">Webhook Testing Interface</h1>
          <p className="text-gray-400 mb-8">
            Test your webhook with custom payloads and view execution results
          </p>

          <WebhookTestPanel webhookId="wh-example-123" />

          <div className="mt-8 p-6 bg-gray-800 border border-gray-700 rounded-lg">
            <h2 className="text-xl font-semibold text-white mb-4">How to Use</h2>
            <ol className="list-decimal list-inside space-y-2 text-gray-300">
              <li>Select the HTTP method for your webhook test</li>
              <li>Optionally choose a sample payload template or enter your own JSON</li>
              <li>Add custom headers if needed (e.g., authorization tokens)</li>
              <li>Click "Send Test Request" to execute the webhook</li>
              <li>View the response status, execution time, and any errors</li>
              <li>Access test history to quickly retry previous configurations</li>
            </ol>
          </div>
        </div>
      </div>
    </BrowserRouter>
  )
}
