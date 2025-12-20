export interface PriorityBadgeProps {
  priority: number
  size?: 'sm' | 'lg'
}

export default function PriorityBadge({ priority, size = 'sm' }: PriorityBadgeProps) {
  const levels = ['Low', 'Normal', 'High', 'Critical']
  const colors = [
    'bg-gray-500/20 text-gray-400',
    'bg-blue-500/20 text-blue-400',
    'bg-yellow-500/20 text-yellow-400',
    'bg-red-500/20 text-red-400',
  ]

  const clampedPriority = Math.max(0, Math.min(priority, 3))

  const sizeClasses = size === 'lg'
    ? 'px-3 py-1 text-sm'
    : 'px-2 py-1 text-xs'

  return (
    <span
      className={`inline-flex font-medium rounded-full ${sizeClasses} ${colors[clampedPriority]}`}
    >
      {levels[clampedPriority]}
    </span>
  )
}
