import { Link } from 'react-router-dom'

const executions = [
  {
    id: '1',
    workflowName: 'Hello World Workflow',
    status: 'completed',
    triggerType: 'webhook',
    startedAt: '2024-01-15T10:30:00Z',
    duration: '1.2s',
  },
  {
    id: '2',
    workflowName: 'Data Sync Pipeline',
    status: 'completed',
    triggerType: 'schedule',
    startedAt: '2024-01-15T10:00:00Z',
    duration: '45.3s',
  },
  {
    id: '3',
    workflowName: 'Alert Notification',
    status: 'failed',
    triggerType: 'webhook',
    startedAt: '2024-01-15T09:45:00Z',
    duration: '0.8s',
  },
  {
    id: '4',
    workflowName: 'Hello World Workflow',
    status: 'running',
    triggerType: 'manual',
    startedAt: '2024-01-15T10:35:00Z',
    duration: '-',
  },
]

export default function ExecutionList() {
  return (
    <div>
      <div className="flex justify-between items-center mb-6">
        <h1 className="text-2xl font-bold text-white">Executions</h1>
        <div className="flex space-x-2">
          <select className="px-4 py-2 bg-gray-700 text-white rounded-lg text-sm border border-gray-600">
            <option>All Workflows</option>
            <option>Hello World Workflow</option>
            <option>Data Sync Pipeline</option>
          </select>
          <select className="px-4 py-2 bg-gray-700 text-white rounded-lg text-sm border border-gray-600">
            <option>All Statuses</option>
            <option>Completed</option>
            <option>Running</option>
            <option>Failed</option>
          </select>
        </div>
      </div>

      <div className="bg-gray-800 rounded-lg overflow-hidden">
        <table className="w-full">
          <thead>
            <tr className="border-b border-gray-700">
              <th className="text-left px-6 py-4 text-sm font-medium text-gray-400">Execution ID</th>
              <th className="text-left px-6 py-4 text-sm font-medium text-gray-400">Workflow</th>
              <th className="text-left px-6 py-4 text-sm font-medium text-gray-400">Status</th>
              <th className="text-left px-6 py-4 text-sm font-medium text-gray-400">Trigger</th>
              <th className="text-left px-6 py-4 text-sm font-medium text-gray-400">Started</th>
              <th className="text-left px-6 py-4 text-sm font-medium text-gray-400">Duration</th>
              <th className="text-right px-6 py-4 text-sm font-medium text-gray-400">Actions</th>
            </tr>
          </thead>
          <tbody>
            {executions.map((execution) => (
              <tr key={execution.id} className="border-b border-gray-700 hover:bg-gray-700/50">
                <td className="px-6 py-4">
                  <Link
                    to={`/executions/${execution.id}`}
                    className="text-primary-400 hover:text-primary-300 font-mono text-sm"
                  >
                    {execution.id.substring(0, 8)}...
                  </Link>
                </td>
                <td className="px-6 py-4 text-white">{execution.workflowName}</td>
                <td className="px-6 py-4">
                  <StatusBadge status={execution.status} />
                </td>
                <td className="px-6 py-4 text-gray-300 capitalize">{execution.triggerType}</td>
                <td className="px-6 py-4 text-gray-300">
                  {new Date(execution.startedAt).toLocaleString()}
                </td>
                <td className="px-6 py-4 text-gray-300">{execution.duration}</td>
                <td className="px-6 py-4 text-right">
                  <Link
                    to={`/executions/${execution.id}`}
                    className="px-3 py-1 text-sm text-gray-300 hover:text-white transition-colors"
                  >
                    View
                  </Link>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  )
}

function StatusBadge({ status }: { status: string }) {
  const colors = {
    completed: 'bg-green-500/20 text-green-400',
    running: 'bg-blue-500/20 text-blue-400',
    failed: 'bg-red-500/20 text-red-400',
    pending: 'bg-yellow-500/20 text-yellow-400',
  }

  return (
    <span className={`inline-flex px-2 py-1 text-xs font-medium rounded-full ${colors[status as keyof typeof colors]}`}>
      {status}
    </span>
  )
}
