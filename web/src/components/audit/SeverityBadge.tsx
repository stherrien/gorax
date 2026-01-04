import { AuditSeverity, SEVERITY_LABELS, SEVERITY_COLORS } from '../../types/audit'

interface SeverityBadgeProps {
  severity: AuditSeverity
  size?: 'sm' | 'md' | 'lg'
}

const sizeClasses = {
  sm: 'px-2 py-0.5 text-xs',
  md: 'px-2.5 py-1 text-sm',
  lg: 'px-3 py-1.5 text-base',
}

const colorClasses: Record<string, string> = {
  blue: 'bg-blue-100 text-blue-800 border-blue-200',
  yellow: 'bg-yellow-100 text-yellow-800 border-yellow-200',
  orange: 'bg-orange-100 text-orange-800 border-orange-200',
  red: 'bg-red-100 text-red-800 border-red-200',
}

export function SeverityBadge({ severity, size = 'md' }: SeverityBadgeProps) {
  const color = SEVERITY_COLORS[severity]
  const label = SEVERITY_LABELS[severity]

  return (
    <span
      className={`inline-flex items-center rounded-md border font-medium ${sizeClasses[size]} ${colorClasses[color]}`}
      role="status"
      aria-label={`Severity: ${label}`}
    >
      {label}
    </span>
  )
}
