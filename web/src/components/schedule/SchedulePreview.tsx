import React, { useState, useEffect } from 'react'
import { scheduleAPI, PreviewScheduleResponse } from '../../api/schedules'
import { formatDistanceToNow, format } from 'date-fns'

export interface SchedulePreviewProps {
  cronExpression: string
  timezone: string
  count?: number
}

const formatRelativeTime = (dateStr: string): string => {
  const date = new Date(dateStr)
  const now = new Date()

  const tomorrow = new Date(now)
  tomorrow.setDate(tomorrow.getDate() + 1)
  tomorrow.setHours(0, 0, 0, 0)

  const dateNorm = new Date(date)
  dateNorm.setHours(0, 0, 0, 0)

  if (dateNorm.getTime() === tomorrow.getTime()) {
    return 'tomorrow'
  }

  return formatDistanceToNow(date, { addSuffix: true })
}

export const SchedulePreview: React.FC<SchedulePreviewProps> = ({
  cronExpression,
  timezone,
  count = 10,
}) => {
  const [preview, setPreview] = useState<PreviewScheduleResponse | null>(null)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    if (!cronExpression) {
      setPreview(null)
      setError(null)
      return
    }

    const fetchPreview = async () => {
      setLoading(true)
      setError(null)

      try {
        const result = await scheduleAPI.preview(cronExpression, timezone, count)
        setPreview(result)
      } catch (err) {
        setError(err instanceof Error ? err.message : 'Failed to fetch preview')
        setPreview(null)
      } finally {
        setLoading(false)
      }
    }

    fetchPreview()
  }, [cronExpression, timezone, count])

  if (!cronExpression) {
    return (
      <div className="text-sm text-gray-600 p-4 bg-gray-50 rounded-md">
        No schedule set
      </div>
    )
  }

  if (loading) {
    return (
      <div className="text-sm text-gray-600 p-4 bg-gray-50 rounded-md">
        Loading preview...
      </div>
    )
  }

  if (error) {
    return (
      <div className="text-sm text-red-600 p-4 bg-red-50 rounded-md border border-red-200">
        Error: {error}
      </div>
    )
  }

  if (!preview || !preview.next_runs || preview.next_runs.length === 0) {
    return null
  }

  return (
    <div className="space-y-2">
      <div className="flex items-center justify-between">
        <h3 className="text-sm font-medium text-gray-700">Next Run Times</h3>
        <span className="text-xs text-gray-500">{preview.timezone}</span>
      </div>

      <div className="bg-white border border-gray-200 rounded-md overflow-hidden">
        <div className="max-h-64 overflow-y-auto">
          {preview.next_runs.map((runTime, index) => {
            const date = new Date(runTime)
            const formattedDate = format(date, 'MMM dd, yyyy')
            const formattedTime = format(date, 'h:mm a')
            const relative = formatRelativeTime(runTime)

            return (
              <div
                key={index}
                className="px-4 py-3 border-b border-gray-100 last:border-0 hover:bg-gray-50"
              >
                <div className="flex items-center justify-between">
                  <div>
                    <div className="text-sm font-medium text-gray-900">
                      {formattedDate} at {formattedTime}
                    </div>
                    <div className="text-xs text-gray-500 mt-1">{relative}</div>
                  </div>
                  <div className="text-xs text-gray-400">#{index + 1}</div>
                </div>
              </div>
            )
          })}
        </div>
      </div>

      <div className="text-xs text-gray-500 text-right">
        Showing {preview.count} upcoming execution{preview.count !== 1 ? 's' : ''}
      </div>
    </div>
  )
}
