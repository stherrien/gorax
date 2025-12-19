import { useState } from 'react'
import {
  startOfMonth,
  endOfMonth,
  startOfWeek,
  endOfWeek,
  addDays,
  addMonths,
  subMonths,
  format,
  isSameMonth,
  isSameDay,
  isToday,
  startOfDay,
} from 'date-fns'
import type { Schedule } from '../../api/schedules'

interface ScheduleCalendarProps {
  schedules: Schedule[]
  onDayClick?: (date: Date, schedules: Schedule[]) => void
}

export default function ScheduleCalendar({
  schedules,
  onDayClick,
}: ScheduleCalendarProps) {
  const [currentMonth, setCurrentMonth] = useState(new Date())

  const renderHeader = () => {
    return (
      <div className="flex items-center justify-between mb-4">
        <h2 className="text-xl font-semibold text-white">
          {format(currentMonth, 'MMMM yyyy')}
        </h2>
        <div className="flex space-x-2">
          <button
            onClick={() => setCurrentMonth(subMonths(currentMonth, 1))}
            aria-label="Previous month"
            className="px-3 py-1 bg-gray-700 text-white rounded hover:bg-gray-600 transition-colors"
          >
            &lt;
          </button>
          <button
            onClick={() => setCurrentMonth(new Date())}
            className="px-3 py-1 bg-gray-700 text-white rounded hover:bg-gray-600 transition-colors"
          >
            Today
          </button>
          <button
            onClick={() => setCurrentMonth(addMonths(currentMonth, 1))}
            aria-label="Next month"
            className="px-3 py-1 bg-gray-700 text-white rounded hover:bg-gray-600 transition-colors"
          >
            &gt;
          </button>
        </div>
      </div>
    )
  }

  const renderDays = () => {
    const days = ['Sun', 'Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat']

    return (
      <div className="grid grid-cols-7 gap-2 mb-2">
        {days.map((day) => (
          <div
            key={day}
            className="text-center text-sm font-medium text-gray-400 py-2"
          >
            {day}
          </div>
        ))}
      </div>
    )
  }

  const renderCells = () => {
    const monthStart = startOfMonth(currentMonth)
    const monthEnd = endOfMonth(monthStart)
    const startDate = startOfWeek(monthStart)
    const endDate = endOfWeek(monthEnd)

    const rows = []
    let days = []
    let day = startDate

    while (day <= endDate) {
      for (let i = 0; i < 7; i++) {
        const cloneDay = day
        const daySchedules = getSchedulesForDay(day, schedules)
        const scheduleCount = daySchedules.length

        days.push(
          <button
            key={day.toString()}
            onClick={() => onDayClick?.(cloneDay, daySchedules)}
            className={getCellClassName(day, currentMonth, scheduleCount)}
          >
            <div className="relative">
              <span className="text-sm">{format(day, 'd')}</span>
              {scheduleCount > 0 && (
                <div className="mt-1 flex justify-center">
                  <span className="inline-flex items-center justify-center w-5 h-5 text-xs bg-primary-600 text-white rounded-full">
                    {scheduleCount}
                  </span>
                </div>
              )}
            </div>
          </button>
        )
        day = addDays(day, 1)
      }
      rows.push(
        <div key={day.toString()} className="grid grid-cols-7 gap-2">
          {days}
        </div>
      )
      days = []
    }

    return <div className="space-y-2">{rows}</div>
  }

  return (
    <div className="bg-gray-800 rounded-lg p-4">
      {renderHeader()}
      {renderDays()}
      {renderCells()}
    </div>
  )
}

function getCellClassName(
  day: Date,
  currentMonth: Date,
  scheduleCount: number
): string {
  const baseClasses =
    'w-full aspect-square flex flex-col items-center justify-center rounded-lg transition-colors'

  const isCurrentMonth = isSameMonth(day, currentMonth)
  const isTodayDate = isToday(day)
  const hasSchedules = scheduleCount > 0

  let classes = baseClasses

  if (!isCurrentMonth) {
    classes += ' text-gray-600 bg-gray-900/20'
  } else if (isTodayDate) {
    classes += ' bg-primary-600/20 text-primary-400 font-bold border-2 border-primary-600'
  } else if (hasSchedules) {
    classes += ' text-white bg-gray-700 hover:bg-gray-600'
  } else {
    classes += ' text-gray-300 hover:bg-gray-700/50'
  }

  return classes
}

function getSchedulesForDay(day: Date, schedules: Schedule[]): Schedule[] {
  const dayStart = startOfDay(day)

  return schedules.filter((schedule) => {
    if (!schedule.nextRunAt) return false

    const nextRun = startOfDay(new Date(schedule.nextRunAt))
    return isSameDay(nextRun, dayStart)
  })
}
