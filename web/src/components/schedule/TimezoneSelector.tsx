import React, { useState, useMemo } from 'react'
import { format } from 'date-fns'

export interface TimezoneSelectorProps {
  value: string
  onChange: (timezone: string) => void
  disabled?: boolean
  searchable?: boolean
  showCurrentTime?: boolean
}

interface TimezoneOption {
  value: string
  label: string
  region: string
  offset: string
}

const POPULAR_TIMEZONES = [
  'UTC',
  'America/New_York',
  'America/Los_Angeles',
  'Europe/London',
]

const ALL_TIMEZONES: TimezoneOption[] = [
  // America
  { value: 'America/New_York', label: 'New York (EST/EDT)', region: 'America', offset: 'UTC-5' },
  { value: 'America/Chicago', label: 'Chicago (CST/CDT)', region: 'America', offset: 'UTC-6' },
  { value: 'America/Denver', label: 'Denver (MST/MDT)', region: 'America', offset: 'UTC-7' },
  { value: 'America/Los_Angeles', label: 'Los Angeles (PST/PDT)', region: 'America', offset: 'UTC-8' },
  { value: 'America/Toronto', label: 'Toronto', region: 'America', offset: 'UTC-5' },
  { value: 'America/Vancouver', label: 'Vancouver', region: 'America', offset: 'UTC-8' },
  { value: 'America/Mexico_City', label: 'Mexico City', region: 'America', offset: 'UTC-6' },
  { value: 'America/Sao_Paulo', label: 'São Paulo', region: 'America', offset: 'UTC-3' },
  { value: 'America/Buenos_Aires', label: 'Buenos Aires', region: 'America', offset: 'UTC-3' },

  // Europe
  { value: 'Europe/London', label: 'London (GMT/BST)', region: 'Europe', offset: 'UTC+0' },
  { value: 'Europe/Paris', label: 'Paris (CET/CEST)', region: 'Europe', offset: 'UTC+1' },
  { value: 'Europe/Berlin', label: 'Berlin', region: 'Europe', offset: 'UTC+1' },
  { value: 'Europe/Rome', label: 'Rome', region: 'Europe', offset: 'UTC+1' },
  { value: 'Europe/Madrid', label: 'Madrid', region: 'Europe', offset: 'UTC+1' },
  { value: 'Europe/Amsterdam', label: 'Amsterdam', region: 'Europe', offset: 'UTC+1' },
  { value: 'Europe/Stockholm', label: 'Stockholm', region: 'Europe', offset: 'UTC+1' },
  { value: 'Europe/Moscow', label: 'Moscow', region: 'Europe', offset: 'UTC+3' },

  // Asia
  { value: 'Asia/Tokyo', label: 'Tokyo', region: 'Asia', offset: 'UTC+9' },
  { value: 'Asia/Shanghai', label: 'Shanghai', region: 'Asia', offset: 'UTC+8' },
  { value: 'Asia/Hong_Kong', label: 'Hong Kong', region: 'Asia', offset: 'UTC+8' },
  { value: 'Asia/Singapore', label: 'Singapore', region: 'Asia', offset: 'UTC+8' },
  { value: 'Asia/Seoul', label: 'Seoul', region: 'Asia', offset: 'UTC+9' },
  { value: 'Asia/Mumbai', label: 'Mumbai', region: 'Asia', offset: 'UTC+5:30' },
  { value: 'Asia/Dubai', label: 'Dubai', region: 'Asia', offset: 'UTC+4' },
  { value: 'Asia/Bangkok', label: 'Bangkok', region: 'Asia', offset: 'UTC+7' },

  // Pacific
  { value: 'Pacific/Auckland', label: 'Auckland', region: 'Pacific', offset: 'UTC+12' },
  { value: 'Pacific/Sydney', label: 'Sydney', region: 'Pacific', offset: 'UTC+10' },
  { value: 'Pacific/Honolulu', label: 'Honolulu', region: 'Pacific', offset: 'UTC-10' },

  // Africa
  { value: 'Africa/Cairo', label: 'Cairo', region: 'Africa', offset: 'UTC+2' },
  { value: 'Africa/Johannesburg', label: 'Johannesburg', region: 'Africa', offset: 'UTC+2' },
  { value: 'Africa/Lagos', label: 'Lagos', region: 'Africa', offset: 'UTC+1' },
]

const getCurrentTimeInTimezone = (_timezone: string): string => {
  try {
    const now = new Date()
    return format(now, 'h:mm a')
  } catch {
    return ''
  }
}

export const TimezoneSelector: React.FC<TimezoneSelectorProps> = ({
  value,
  onChange,
  disabled,
  searchable,
  showCurrentTime,
}) => {
  const [searchTerm, setSearchTerm] = useState('')

  const filteredTimezones = useMemo(() => {
    if (!searchTerm) return ALL_TIMEZONES

    const term = searchTerm.toLowerCase()
    return ALL_TIMEZONES.filter(
      (tz) =>
        tz.label.toLowerCase().includes(term) ||
        tz.value.toLowerCase().includes(term) ||
        tz.region.toLowerCase().includes(term)
    )
  }, [searchTerm])

  const groupedTimezones = useMemo(() => {
    const groups: Record<string, TimezoneOption[]> = {}

    filteredTimezones.forEach((tz) => {
      if (!groups[tz.region]) {
        groups[tz.region] = []
      }
      groups[tz.region].push(tz)
    })

    return groups
  }, [filteredTimezones])

  const currentTime = showCurrentTime ? getCurrentTimeInTimezone(value) : null

  const selectedTimezone = ALL_TIMEZONES.find((tz) => tz.value === value)

  return (
    <div className="space-y-2">
      <label className="block text-sm font-medium text-gray-700">
        Timezone
        {selectedTimezone && (
          <span className="ml-2 text-xs text-gray-500">{selectedTimezone.offset}</span>
        )}
      </label>

      {searchable && (
        <input
          type="text"
          placeholder="Search timezone..."
          value={searchTerm}
          onChange={(e) => setSearchTerm(e.target.value)}
          className="w-full px-3 py-2 border border-gray-300 rounded-md text-sm mb-2"
          disabled={disabled}
        />
      )}

      <select
        value={value}
        onChange={(e) => onChange(e.target.value)}
        disabled={disabled}
        className="w-full px-3 py-2 border border-gray-300 rounded-md text-sm focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
      >
        <optgroup label="Popular">
          {POPULAR_TIMEZONES.map((tz) => {
            const tzOption = ALL_TIMEZONES.find((t) => t.value === tz)
            return (
              <option key={tz} value={tz}>
                {tz === 'UTC' ? 'UTC' : tzOption?.label || tz}
              </option>
            )
          })}
        </optgroup>

        {Object.entries(groupedTimezones).map(([region, timezones]) => (
          <optgroup key={region} label={region}>
            {timezones.map((tz) => (
              <option key={tz.value} value={tz.value}>
                {tz.label} ({tz.offset})
              </option>
            ))}
          </optgroup>
        ))}
      </select>

      {showCurrentTime && currentTime && (
        <div className="text-xs text-gray-500">
          Current time: {currentTime}
        </div>
      )}
    </div>
  )
}
