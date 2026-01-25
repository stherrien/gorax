import { AuditStatus, STATUS_LABELS } from '../../types/audit'
import {
  CheckCircleIcon,
  XCircleIcon,
  ExclamationCircleIcon,
} from '@heroicons/react/24/solid'

interface StatusIconProps {
  status: AuditStatus
  size?: number
  showLabel?: boolean
}

const iconComponents = {
  success: CheckCircleIcon,
  failure: XCircleIcon,
  partial: ExclamationCircleIcon,
}

const colorClasses = {
  success: 'text-green-500',
  failure: 'text-red-500',
  partial: 'text-yellow-500',
}

export function StatusIcon({
  status,
  size = 20,
  showLabel = false,
}: StatusIconProps) {
  const Icon = iconComponents[status]
  const label = STATUS_LABELS[status]
  const colorClass = colorClasses[status]

  if (showLabel) {
    return (
      <span className="inline-flex items-center gap-1.5">
        <Icon
          className={colorClass}
          style={{ width: size, height: size }}
          aria-hidden="true"
        />
        <span className="text-sm font-medium text-gray-700">{label}</span>
      </span>
    )
  }

  return (
    <Icon
      className={colorClass}
      style={{ width: size, height: size }}
      aria-label={`Status: ${label}`}
      title={label}
    />
  )
}
