import { useState } from 'react'
import { Link } from 'react-router-dom'
import { format, addHours, isWithinInterval } from 'date-fns'
import type { Schedule } from '../../api/schedules'

interface Workflow {
  id: string
  name: string
}

interface ScheduleTimelineProps {
  schedules: Schedule[]
  workflows?: Workflow[]
  hoursToShow?: number
}

interface TimelineSchedule {
  schedule: Schedule
  workflowName: string
  hourOffset: number
}

export default function ScheduleTimeline({
  schedules,
  workflows = [],
  hoursToShow = 24,
}: ScheduleTimelineProps) {
  const [hoveredSchedule, setHoveredSchedule] = useState<string | null>(null)

  const now = new Date()
  const endTime = addHours(now, hoursToShow)

  const timelineSchedules = getTimelineSchedules(
    schedules,
    workflows,
    now,
    endTime
  )

  if (timelineSchedules.length === 0) {
    return (
      <div className="bg-gray-800 rounded-lg p-8 text-center">
        <p className="text-gray-400">
          No scheduled runs in the next {hoursToShow} hours
        </p>
      </div>
    )
  }

  return (
    <div className="bg-gray-800 rounded-lg p-4">
      <div className="flex items-center justify-between mb-4">
        <h2 className="text-xl font-semibold text-white">
          Next {hoursToShow} Hours
        </h2>
        <div className="text-sm text-gray-400">
          {format(now, 'MMM d, yyyy HH:mm')}
        </div>
      </div>

      <div className="relative">
        <div className="flex items-start space-x-4 overflow-x-auto pb-4">
          {/* Now marker */}
          <div className="flex-shrink-0 w-1 bg-primary-600 h-full absolute left-0 z-10">
            <div className="absolute -top-1 -left-3 bg-primary-600 text-white text-xs px-2 py-1 rounded">
              Now
            </div>
          </div>

          {/* Timeline grid */}
          <div className="flex space-x-4 pl-8 min-w-full">
            {Array.from({ length: hoursToShow }, (_, i) => {
              const hourTime = addHours(now, i)
              const hourSchedules = timelineSchedules.filter(
                (ts) => ts.hourOffset === i
              )

              return (
                <div
                  key={i}
                  className="flex-shrink-0 w-32 border-l border-gray-700 pl-2"
                >
                  <div className="text-sm font-medium text-gray-400 mb-2">
                    {format(hourTime, 'HH:mm')}
                  </div>

                  <div className="space-y-2">
                    {hourSchedules.map((ts) => (
                      <div
                        key={ts.schedule.id}
                        className="relative"
                        onMouseEnter={() => setHoveredSchedule(ts.schedule.id)}
                        onMouseLeave={() => setHoveredSchedule(null)}
                      >
                        <Link
                          to={`/workflows/${ts.schedule.workflowId}`}
                          className="block bg-primary-600/20 border border-primary-600 rounded px-2 py-1 hover:bg-primary-600/30 transition-colors"
                        >
                          <div className="text-sm text-white truncate">
                            {ts.schedule.name}
                          </div>
                          <div className="text-xs text-primary-300">
                            {format(new Date(ts.schedule.nextRunAt!), 'HH:mm')}
                          </div>
                        </Link>

                        {hoveredSchedule === ts.schedule.id && (
                          <div className="absolute z-20 left-0 top-full mt-1 bg-gray-900 border border-gray-700 rounded-lg p-3 shadow-lg min-w-64">
                            <div className="text-white font-medium mb-1">
                              {ts.schedule.name}
                            </div>
                            <div className="text-sm text-gray-300 mb-2">
                              {ts.workflowName}
                            </div>
                            <div className="text-xs text-gray-400 space-y-1">
                              <div>
                                <span className="text-gray-500">Cron:</span>{' '}
                                <code className="text-gray-300 bg-gray-800 px-1 rounded">
                                  {ts.schedule.cronExpression}
                                </code>
                              </div>
                              <div>
                                <span className="text-gray-500">Timezone:</span>{' '}
                                {ts.schedule.timezone}
                              </div>
                              <div>
                                <span className="text-gray-500">Next run:</span>{' '}
                                {format(
                                  new Date(ts.schedule.nextRunAt!),
                                  'MMM d, yyyy HH:mm:ss'
                                )}
                              </div>
                            </div>
                          </div>
                        )}
                      </div>
                    ))}
                  </div>
                </div>
              )
            })}
          </div>
        </div>
      </div>
    </div>
  )
}

function getTimelineSchedules(
  schedules: Schedule[],
  workflows: Workflow[],
  startTime: Date,
  endTime: Date
): TimelineSchedule[] {
  return schedules
    .filter((schedule) => {
      if (!schedule.nextRunAt || !schedule.enabled) return false

      const nextRun = new Date(schedule.nextRunAt)
      return isWithinInterval(nextRun, { start: startTime, end: endTime })
    })
    .map((schedule) => {
      const nextRun = new Date(schedule.nextRunAt!)
      const hourOffset = Math.floor(
        (nextRun.getTime() - startTime.getTime()) / (1000 * 60 * 60)
      )

      const workflow = workflows.find((w) => w.id === schedule.workflowId)
      const workflowName = workflow?.name || 'Unknown Workflow'

      return {
        schedule,
        workflowName,
        hourOffset,
      }
    })
    .sort((a, b) => {
      const aTime = new Date(a.schedule.nextRunAt!).getTime()
      const bTime = new Date(b.schedule.nextRunAt!).getTime()
      return aTime - bTime
    })
}
