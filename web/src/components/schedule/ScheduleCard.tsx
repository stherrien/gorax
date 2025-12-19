import { Link } from 'react-router-dom'
import { formatDistanceToNow } from 'date-fns'
import type { Schedule } from '../../api/schedules'

interface ScheduleCardProps {
  schedule: Schedule
  workflowName: string
  onToggle: (id: string, currentEnabled: boolean) => void
  onEdit: (id: string) => void
  onDelete: (id: string) => void
  disabled?: boolean
}

export default function ScheduleCard({
  schedule,
  workflowName,
  onToggle,
  onEdit,
  onDelete,
  disabled = false,
}: ScheduleCardProps) {
  return (
    <div className="bg-gray-800 rounded-lg p-4 hover:bg-gray-700/50 transition-colors">
      <div className="flex items-start justify-between mb-3">
        <div className="flex-1">
          <h3 className="text-white font-medium text-lg mb-1">{schedule.name}</h3>
          <Link
            to={`/workflows/${schedule.workflowId}`}
            className="text-primary-400 hover:text-primary-300 text-sm transition-colors"
          >
            {workflowName}
          </Link>
        </div>
        <div className="flex items-center space-x-2">
          <StatusBadge enabled={schedule.enabled} />
          <ToggleSwitch
            enabled={schedule.enabled}
            disabled={disabled}
            onChange={() => onToggle(schedule.id, schedule.enabled)}
          />
        </div>
      </div>

      <div className="space-y-2 mb-3">
        <div className="flex items-center text-sm">
          <span className="text-gray-400 w-24">Cron:</span>
          <code className="text-gray-300 font-mono bg-gray-900/50 px-2 py-1 rounded">
            {schedule.cronExpression}
          </code>
        </div>
        <div className="flex items-center text-sm">
          <span className="text-gray-400 w-24">Timezone:</span>
          <span className="text-gray-300">{schedule.timezone}</span>
        </div>
        <div className="flex items-center text-sm">
          <span className="text-gray-400 w-24">Next run:</span>
          <span className="text-gray-300">
            {formatNextRun(schedule.nextRunAt)}
          </span>
        </div>
        {schedule.lastRunAt && (
          <div className="flex items-center text-sm">
            <span className="text-gray-400 w-24">Last run:</span>
            <span className="text-gray-300">
              {formatDistanceToNow(new Date(schedule.lastRunAt), { addSuffix: true })}
            </span>
          </div>
        )}
      </div>

      <div className="flex space-x-2 pt-3 border-t border-gray-700">
        <button
          onClick={() => onEdit(schedule.id)}
          disabled={disabled}
          className="px-3 py-1 text-sm text-gray-300 hover:text-white transition-colors disabled:opacity-50"
        >
          Edit
        </button>
        <button
          onClick={() => onDelete(schedule.id)}
          disabled={disabled}
          className="px-3 py-1 text-sm text-red-400 hover:text-red-300 transition-colors disabled:opacity-50"
        >
          Delete
        </button>
      </div>
    </div>
  )
}

function StatusBadge({ enabled }: { enabled: boolean }) {
  return (
    <span
      className={`inline-flex px-2 py-1 text-xs font-medium rounded-full ${
        enabled
          ? 'bg-green-500/20 text-green-400'
          : 'bg-gray-500/20 text-gray-400'
      }`}
    >
      {enabled ? 'Enabled' : 'Disabled'}
    </span>
  )
}

interface ToggleSwitchProps {
  enabled: boolean
  disabled?: boolean
  onChange: () => void
}

function ToggleSwitch({ enabled, disabled, onChange }: ToggleSwitchProps) {
  return (
    <button
      type="button"
      role="switch"
      aria-checked={enabled}
      disabled={disabled}
      onClick={onChange}
      className={`relative inline-flex h-6 w-11 items-center rounded-full transition-colors focus:outline-none focus:ring-2 focus:ring-primary-500 focus:ring-offset-2 focus:ring-offset-gray-800 disabled:opacity-50 disabled:cursor-not-allowed ${
        enabled ? 'bg-primary-600' : 'bg-gray-600'
      }`}
    >
      <span
        className={`inline-block h-4 w-4 transform rounded-full bg-white transition-transform ${
          enabled ? 'translate-x-6' : 'translate-x-1'
        }`}
      />
    </button>
  )
}

function formatNextRun(nextRunAt?: string): string {
  if (!nextRunAt) return 'Not scheduled'

  const nextRun = new Date(nextRunAt)
  const now = new Date()

  if (nextRun < now) return 'Overdue'

  return formatDistanceToNow(nextRun, { addSuffix: true })
}
