import React, { useState } from 'react'
import { CronBuilder } from './CronBuilder'
import { TimezoneSelector } from './TimezoneSelector'
import { SchedulePreview } from './SchedulePreview'

/**
 * Example integration of the Cron Builder components
 * This shows how to use CronBuilder, TimezoneSelector, and SchedulePreview together
 */
export const ScheduleBuilderExample: React.FC = () => {
  const [cronExpression, setCronExpression] = useState('0 9 * * *')
  const [timezone, setTimezone] = useState('UTC')

  const handleSubmit = () => {
    console.log('Schedule created:', {
      cronExpression,
      timezone,
    })
    // In a real application, you would call the API here:
    // await scheduleAPI.create(workflowId, { name, cron_expression: cronExpression, timezone, enabled: true })
  }

  return (
    <div className="max-w-3xl mx-auto p-6 space-y-6">
      <div>
        <h1 className="text-2xl font-bold mb-2">Schedule Builder</h1>
        <p className="text-gray-600">
          Configure when your workflow should run automatically
        </p>
      </div>

      <div className="bg-white p-6 rounded-lg border border-gray-200 space-y-6">
        <CronBuilder value={cronExpression} onChange={setCronExpression} />

        <TimezoneSelector
          value={timezone}
          onChange={setTimezone}
          searchable
          showCurrentTime
        />

        {cronExpression && (
          <div className="pt-4 border-t border-gray-200">
            <SchedulePreview
              cronExpression={cronExpression}
              timezone={timezone}
              count={10}
            />
          </div>
        )}
      </div>

      <div className="flex justify-end gap-3">
        <button
          type="button"
          className="px-4 py-2 text-sm border border-gray-300 rounded-md hover:bg-gray-50"
        >
          Cancel
        </button>
        <button
          type="button"
          onClick={handleSubmit}
          className="px-4 py-2 text-sm bg-blue-600 text-white rounded-md hover:bg-blue-700"
        >
          Create Schedule
        </button>
      </div>
    </div>
  )
}
