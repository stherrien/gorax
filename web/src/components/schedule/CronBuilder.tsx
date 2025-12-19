import React, { useState, useMemo, useRef, useEffect } from 'react'

export interface CronBuilderProps {
  value: string
  onChange: (value: string) => void
  disabled?: boolean
}

interface CronPreset {
  label: string
  value: string
  description: string
}

const PRESETS: CronPreset[] = [
  { label: 'Every minute', value: '* * * * *', description: 'Runs every minute' },
  { label: 'Every hour', value: '0 * * * *', description: 'Runs at minute 0 of every hour' },
  {
    label: 'Every day at midnight',
    value: '0 0 * * *',
    description: 'Runs at 00:00 every day',
  },
  {
    label: 'Every day at 9 AM',
    value: '0 9 * * *',
    description: 'Runs at 09:00 every day',
  },
  {
    label: 'Weekly on Monday at 9 AM',
    value: '0 9 * * 1',
    description: 'Runs at 09:00 every Monday',
  },
  {
    label: 'Monthly on the 1st',
    value: '0 0 1 * *',
    description: 'Runs at 00:00 on day 1 of every month',
  },
  {
    label: 'Weekdays at 9 AM',
    value: '0 9 * * 1-5',
    description: 'Runs at 09:00 Monday through Friday',
  },
]

const parseCronExpression = (cron: string): string => {
  if (!cron) return 'No schedule set'

  const presetMatch = PRESETS.find((p) => p.value === cron)
  if (presetMatch) return presetMatch.description

  try {
    const parts = cron.split(' ')
    if (parts.length !== 5) return 'Invalid cron expression'

    const [minute, hour, dayOfMonth, month, dayOfWeek] = parts

    const descriptions: string[] = []

    if (dayOfWeek !== '*') {
      const days = dayOfWeek.split(',').map((d) => {
        const dayMap: Record<string, string> = {
          '0': 'Sunday',
          '1': 'Monday',
          '2': 'Tuesday',
          '3': 'Wednesday',
          '4': 'Thursday',
          '5': 'Friday',
          '6': 'Saturday',
        }
        if (d.includes('-')) {
          const [start, end] = d.split('-')
          return `${dayMap[start]} through ${dayMap[end]}`
        }
        return dayMap[d] || d
      })

      if (dayOfWeek === '1-5') {
        descriptions.push('weekdays')
      } else {
        descriptions.push(`on ${days.join(', ')}`)
      }
    }

    if (hour !== '*' && minute !== '*') {
      const h = parseInt(hour)
      const m = parseInt(minute)
      const ampm = h >= 12 ? 'PM' : 'AM'
      const displayHour = h > 12 ? h - 12 : h === 0 ? 12 : h
      descriptions.push(`at ${displayHour}:${m.toString().padStart(2, '0')} ${ampm}`)
    } else if (hour === '*' && minute !== '*') {
      descriptions.push(`at minute ${minute}`)
    } else if (hour !== '*' && minute === '*') {
      descriptions.push(`every minute of hour ${hour}`)
    }

    if (dayOfMonth !== '*' && dayOfWeek === '*') {
      descriptions.push(`on day ${dayOfMonth} of the month`)
    }

    if (month !== '*') {
      const monthMap: Record<string, string> = {
        '1': 'January',
        '2': 'February',
        '3': 'March',
        '4': 'April',
        '5': 'May',
        '6': 'June',
        '7': 'July',
        '8': 'August',
        '9': 'September',
        '10': 'October',
        '11': 'November',
        '12': 'December',
      }
      descriptions.push(`in ${monthMap[month] || month}`)
    }

    if (descriptions.length === 0) {
      return 'Runs every minute'
    }

    let result = 'Every '
    if (dayOfWeek !== '*') {
      result += descriptions[0]
      if (descriptions.length > 1) {
        result += ' ' + descriptions.slice(1).join(' ')
      }
    } else if (dayOfMonth !== '*') {
      result = 'Every month ' + descriptions.join(' ')
    } else if (hour !== '*') {
      result = 'Every day ' + descriptions.join(' ')
    } else {
      result += descriptions.join(' ')
    }

    return result
  } catch {
    return 'Invalid cron expression'
  }
}

export const CronBuilder: React.FC<CronBuilderProps> = ({ value, onChange, disabled }) => {
  const [isPresetsOpen, setIsPresetsOpen] = useState(false)
  const [mode, setMode] = useState<'simple' | 'advanced'>('simple')
  const presetsRef = useRef<HTMLDivElement>(null)

  const description = useMemo(() => parseCronExpression(value), [value])
  const isValid = useMemo(() => {
    if (!value) return true
    const parts = value.split(' ')
    return parts.length === 5
  }, [value])

  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (presetsRef.current && !presetsRef.current.contains(event.target as Node)) {
        setIsPresetsOpen(false)
      }
    }

    document.addEventListener('mousedown', handleClickOutside)
    return () => document.removeEventListener('mousedown', handleClickOutside)
  }, [])

  const handlePresetSelect = (preset: CronPreset) => {
    onChange(preset.value)
    setIsPresetsOpen(false)
  }

  const handleInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    onChange(e.target.value)
  }

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <label className="block text-sm font-medium text-gray-700">
          Cron Expression
        </label>
        <div className="flex gap-2">
          <button
            type="button"
            onClick={() => setMode(mode === 'simple' ? 'advanced' : 'simple')}
            className="text-sm text-blue-600 hover:text-blue-800"
            disabled={disabled}
          >
            {mode === 'simple' ? 'Advanced' : 'Simple'}
          </button>
        </div>
      </div>

      {mode === 'advanced' ? (
        <div>
          <input
            type="text"
            value={value}
            onChange={handleInputChange}
            placeholder="* * * * *"
            className={`w-full px-3 py-2 border rounded-md font-mono text-sm ${
              !isValid ? 'border-red-500' : 'border-gray-300'
            }`}
            disabled={disabled}
            aria-label="Cron expression input"
          />
        </div>
      ) : (
        <div className="space-y-2">
          <div className="grid grid-cols-5 gap-2">
            <div>
              <label htmlFor="minute" className="block text-xs text-gray-600 mb-1">
                Minute
              </label>
              <input
                id="minute"
                type="text"
                placeholder="*"
                className="w-full px-2 py-1 border border-gray-300 rounded text-sm"
                disabled={disabled}
                aria-label="Minute"
              />
            </div>
            <div>
              <label htmlFor="hour" className="block text-xs text-gray-600 mb-1">
                Hour
              </label>
              <input
                id="hour"
                type="text"
                placeholder="*"
                className="w-full px-2 py-1 border border-gray-300 rounded text-sm"
                disabled={disabled}
                aria-label="Hour"
              />
            </div>
            <div>
              <label htmlFor="day" className="block text-xs text-gray-600 mb-1">
                Day
              </label>
              <input
                id="day"
                type="text"
                placeholder="*"
                className="w-full px-2 py-1 border border-gray-300 rounded text-sm"
                disabled={disabled}
              />
            </div>
            <div>
              <label htmlFor="month" className="block text-xs text-gray-600 mb-1">
                Month
              </label>
              <input
                id="month"
                type="text"
                placeholder="*"
                className="w-full px-2 py-1 border border-gray-300 rounded text-sm"
                disabled={disabled}
              />
            </div>
            <div>
              <label htmlFor="weekday" className="block text-xs text-gray-600 mb-1">
                Weekday
              </label>
              <input
                id="weekday"
                type="text"
                placeholder="*"
                className="w-full px-2 py-1 border border-gray-300 rounded text-sm"
                disabled={disabled}
              />
            </div>
          </div>
        </div>
      )}

      <div className="relative" ref={presetsRef}>
        <button
          type="button"
          onClick={() => setIsPresetsOpen(!isPresetsOpen)}
          className="px-3 py-2 text-sm border border-gray-300 rounded-md hover:bg-gray-50"
          disabled={disabled}
        >
          Presets
        </button>

        {isPresetsOpen && (
          <div className="absolute z-10 mt-1 w-64 bg-white border border-gray-300 rounded-md shadow-lg">
            <div className="max-h-64 overflow-y-auto">
              {PRESETS.map((preset) => (
                <button
                  key={preset.value}
                  type="button"
                  onClick={() => handlePresetSelect(preset)}
                  className="w-full px-4 py-2 text-left text-sm hover:bg-gray-100 border-b border-gray-100 last:border-0"
                >
                  <div className="font-medium">{preset.label}</div>
                  <div className="text-xs text-gray-500 font-mono">{preset.value}</div>
                </button>
              ))}
            </div>
          </div>
        )}
      </div>

      <div
        className={`text-sm ${!isValid ? 'text-red-600' : 'text-gray-600'} bg-gray-50 p-3 rounded-md`}
      >
        {description}
      </div>
    </div>
  )
}
