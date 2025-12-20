import { useState } from 'react'
import { useNavigate, useSearchParams } from 'react-router-dom'
import { ScheduleForm, ScheduleFormData } from '../components/schedule/ScheduleForm'
import { scheduleAPI } from '../api/schedules'

export default function CreateSchedule() {
  const navigate = useNavigate()
  const [searchParams] = useSearchParams()
  const workflowId = searchParams.get('workflowId')

  const [error, setError] = useState<string | null>(null)

  const handleSubmit = async (data: ScheduleFormData) => {
    if (!workflowId) {
      setError('Workflow ID is required')
      return
    }

    try {
      setError(null)
      await scheduleAPI.create(workflowId, {
        name: data.name,
        cronExpression: data.cronExpression,
        timezone: data.timezone,
        enabled: data.enabled,
      })
      navigate('/schedules')
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Failed to create schedule'
      setError(errorMessage)
    }
  }

  const handleCancel = () => {
    navigate('/schedules')
  }

  if (!workflowId) {
    return (
      <div className="max-w-3xl mx-auto p-6">
        <div className="bg-red-50 border border-red-200 rounded-lg p-4">
          <p className="text-red-600">Workflow ID is required to create a schedule</p>
        </div>
      </div>
    )
  }

  return (
    <div className="max-w-3xl mx-auto p-6">
      <div className="mb-6">
        <h1 className="text-2xl font-bold text-white">Create Schedule</h1>
        <p className="text-gray-400 text-sm mt-1">
          Configure when your workflow should run automatically
        </p>
      </div>

      {error && (
        <div className="mb-6 bg-red-50 border border-red-200 rounded-lg p-4">
          <p className="text-red-600 text-sm">{error}</p>
        </div>
      )}

      <div className="bg-white rounded-lg border border-gray-200 p-6">
        <ScheduleForm onSubmit={handleSubmit} onCancel={handleCancel} />
      </div>
    </div>
  )
}
