import { useState } from 'react'
import { useNavigate, useParams, Navigate } from 'react-router-dom'
import { ScheduleForm, ScheduleFormData } from '../components/schedule/ScheduleForm'
import { useSchedule, useScheduleMutations } from '../hooks/useSchedules'
import { isValidResourceId } from '../utils/routing'

export default function EditSchedule() {
  const navigate = useNavigate()
  const { id } = useParams<{ id: string }>()

  // Guard against invalid IDs
  if (!isValidResourceId(id)) {
    return <Navigate to="/schedules" replace />
  }

  const { schedule, loading, error } = useSchedule(id)
  const { updateSchedule } = useScheduleMutations()

  const [updateError, setUpdateError] = useState<string | null>(null)

  const handleSubmit = async (data: ScheduleFormData) => {
    try {
      setUpdateError(null)
      await updateSchedule(id, {
        name: data.name,
        cronExpression: data.cronExpression,
        timezone: data.timezone,
        enabled: data.enabled,
      })
      navigate('/schedules')
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Failed to update schedule'
      setUpdateError(errorMessage)
    }
  }

  const handleCancel = () => {
    navigate('/schedules')
  }

  if (loading) {
    return (
      <div className="max-w-3xl mx-auto p-6">
        <div className="h-64 flex items-center justify-center">
          <div className="text-white text-lg">Loading schedule...</div>
        </div>
      </div>
    )
  }

  if (error || !schedule) {
    return (
      <div className="max-w-3xl mx-auto p-6">
        <div className="bg-red-800/50 rounded-lg p-4">
          <p className="text-red-200">
            {error?.message || 'Failed to load schedule'}
          </p>
        </div>
      </div>
    )
  }

  return (
    <div className="max-w-3xl mx-auto p-6">
      <div className="mb-6">
        <h1 className="text-2xl font-bold text-white">Edit Schedule</h1>
        <p className="text-gray-400 text-sm mt-1">
          Update your workflow schedule configuration
        </p>
      </div>

      {updateError && (
        <div className="mb-6 bg-red-800/50 rounded-lg p-4">
          <p className="text-red-200 text-sm">{updateError}</p>
        </div>
      )}

      <div className="bg-gray-800 rounded-lg border border-gray-700 p-6">
        <ScheduleForm
          initialData={{
            name: schedule.name,
            cronExpression: schedule.cronExpression,
            timezone: schedule.timezone,
            enabled: schedule.enabled,
          }}
          onSubmit={handleSubmit}
          onCancel={handleCancel}
          submitLabel="Update Schedule"
        />
      </div>
    </div>
  )
}
