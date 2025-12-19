import ScheduleCard from './ScheduleCard'
import type { Schedule } from '../../api/schedules'

interface Workflow {
  id: string
  name: string
}

interface ScheduleListProps {
  schedules: Schedule[]
  workflows: Workflow[]
  onToggle: (id: string, currentEnabled: boolean) => void
  onEdit: (id: string) => void
  onDelete: (id: string) => void
  sortBy?: 'name' | 'nextRun' | 'lastRun'
  disabled?: boolean
}

export default function ScheduleList({
  schedules,
  workflows,
  onToggle,
  onEdit,
  onDelete,
  sortBy = 'nextRun',
  disabled = false,
}: ScheduleListProps) {
  if (schedules.length === 0) {
    return (
      <div className="h-64 flex items-center justify-center bg-gray-800 rounded-lg">
        <div className="text-center">
          <div className="text-gray-400 text-lg">No schedules found</div>
        </div>
      </div>
    )
  }

  const sortedSchedules = sortSchedules(schedules, sortBy)

  return (
    <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
      {sortedSchedules.map((schedule) => (
        <ScheduleCard
          key={schedule.id}
          schedule={schedule}
          workflowName={getWorkflowName(schedule.workflowId, workflows)}
          onToggle={onToggle}
          onEdit={onEdit}
          onDelete={onDelete}
          disabled={disabled}
        />
      ))}
    </div>
  )
}

function getWorkflowName(workflowId: string, workflows: Workflow[]): string {
  const workflow = workflows.find((w) => w.id === workflowId)
  return workflow?.name || 'Unknown Workflow'
}

function sortSchedules(
  schedules: Schedule[],
  sortBy: 'name' | 'nextRun' | 'lastRun'
): Schedule[] {
  const sorted = [...schedules]

  switch (sortBy) {
    case 'name':
      return sorted.sort((a, b) => a.name.localeCompare(b.name))

    case 'nextRun':
      return sorted.sort((a, b) => {
        if (!a.nextRunAt) return 1
        if (!b.nextRunAt) return -1
        return new Date(a.nextRunAt).getTime() - new Date(b.nextRunAt).getTime()
      })

    case 'lastRun':
      return sorted.sort((a, b) => {
        if (!a.lastRunAt) return 1
        if (!b.lastRunAt) return -1
        return new Date(b.lastRunAt).getTime() - new Date(a.lastRunAt).getTime()
      })

    default:
      return sorted
  }
}
