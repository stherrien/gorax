export interface PrioritySelectorProps {
  value?: number
  onChange: (priority: number) => void
  disabled?: boolean
  id?: string
}

const priorityOptions = [
  { value: 0, label: 'Low (0)' },
  { value: 1, label: 'Normal (1)' },
  { value: 2, label: 'High (2)' },
  { value: 3, label: 'Critical (3)' },
]

export default function PrioritySelector({
  value = 1,
  onChange,
  disabled = false,
  id = 'priority-selector',
}: PrioritySelectorProps) {
  const handleChange = (e: React.ChangeEvent<HTMLSelectElement>) => {
    const newValue = parseInt(e.target.value, 10)
    onChange(newValue)
  }

  return (
    <div>
      <label htmlFor={id} className="block text-sm font-medium text-gray-300 mb-2">
        Priority
      </label>
      <select
        id={id}
        value={value}
        onChange={handleChange}
        disabled={disabled}
        className="w-full px-3 py-2 bg-gray-700 text-white rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-primary-500 disabled:opacity-50 disabled:cursor-not-allowed"
      >
        {priorityOptions.map((option) => (
          <option key={option.value} value={option.value}>
            {option.label}
          </option>
        ))}
      </select>
    </div>
  )
}
