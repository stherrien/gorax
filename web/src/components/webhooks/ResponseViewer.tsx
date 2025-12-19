import { Link } from 'react-router-dom'
import type { TestWebhookResponse } from '../../api/webhooks'

interface ResponseViewerProps {
  response: TestWebhookResponse | null
  error: string | null
  getStatusCodeColor: (statusCode: number) => string
}

export default function ResponseViewer({ response, error, getStatusCodeColor }: ResponseViewerProps) {
  if (!response && !error) {
    return null
  }

  return (
    <div className="border-t border-gray-700 pt-6">
      <h3 className="text-white text-lg font-semibold mb-4">Response</h3>

      {error && (
        <div className="p-4 bg-red-900/20 border border-red-500/30 rounded-lg">
          <p className="text-red-400 text-sm">{error}</p>
        </div>
      )}

      {response && (
        <div className="space-y-3">
          <div className="flex items-center gap-4">
            <div>
              <span className="text-gray-400 text-sm">Status Code: </span>
              <span className={`text-lg font-semibold ${getStatusCodeColor(response.statusCode)}`}>
                {response.statusCode}
              </span>
            </div>
            <div>
              <span className="text-gray-400 text-sm">Response Time: </span>
              <span className="text-white text-sm font-medium">
                {response.responseTimeMs} ms
              </span>
            </div>
          </div>

          {response.executionId && (
            <div>
              <span className="text-gray-400 text-sm">Execution ID: </span>
              <Link
                to={`/executions/${response.executionId}`}
                className="text-primary-400 hover:text-primary-300 text-sm font-medium underline"
              >
                {response.executionId}
              </Link>
            </div>
          )}

          {response.error && (
            <div className="p-4 bg-red-900/20 border border-red-500/30 rounded-lg">
              <p className="text-red-400 text-sm">{response.error}</p>
            </div>
          )}
        </div>
      )}
    </div>
  )
}
