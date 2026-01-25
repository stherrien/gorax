import { AuditCategory, CATEGORY_LABELS } from '../../types/audit'
import {
  LockClosedIcon,
  ShieldCheckIcon,
  DocumentTextIcon,
  CogIcon,
  PlayIcon,
  LinkIcon,
  KeyIcon,
  UserGroupIcon,
  ServerIcon,
} from '@heroicons/react/24/outline'

interface CategoryIconProps {
  category: AuditCategory
  size?: number
  showLabel?: boolean
  className?: string
}

const iconComponents: Record<AuditCategory, React.ComponentType<any>> = {
  [AuditCategory.Authentication]: LockClosedIcon,
  [AuditCategory.Authorization]: ShieldCheckIcon,
  [AuditCategory.DataAccess]: DocumentTextIcon,
  [AuditCategory.Configuration]: CogIcon,
  [AuditCategory.Workflow]: PlayIcon,
  [AuditCategory.Integration]: LinkIcon,
  [AuditCategory.Credential]: KeyIcon,
  [AuditCategory.UserManagement]: UserGroupIcon,
  [AuditCategory.System]: ServerIcon,
}

const colorClasses: Record<AuditCategory, string> = {
  [AuditCategory.Authentication]: 'text-blue-600',
  [AuditCategory.Authorization]: 'text-purple-600',
  [AuditCategory.DataAccess]: 'text-green-600',
  [AuditCategory.Configuration]: 'text-gray-600',
  [AuditCategory.Workflow]: 'text-indigo-600',
  [AuditCategory.Integration]: 'text-cyan-600',
  [AuditCategory.Credential]: 'text-amber-600',
  [AuditCategory.UserManagement]: 'text-pink-600',
  [AuditCategory.System]: 'text-slate-600',
}

export function CategoryIcon({
  category,
  size = 20,
  showLabel = false,
  className = '',
}: CategoryIconProps) {
  const Icon = iconComponents[category]
  const label = CATEGORY_LABELS[category]
  const colorClass = colorClasses[category]

  if (showLabel) {
    return (
      <span className={`inline-flex items-center gap-2 ${className}`}>
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
      className={`${colorClass} ${className}`}
      style={{ width: size, height: size }}
      aria-label={`Category: ${label}`}
      title={label}
    />
  )
}
