import { Lock } from 'lucide-react'
import type { EditLock } from '../../types/collaboration'

interface NodeLockIndicatorProps {
  lock: EditLock
  currentUserId: string
}

export default function NodeLockIndicator({ lock, currentUserId }: NodeLockIndicatorProps) {
  const isLockedByMe = lock.user_id === currentUserId

  return (
    <div
      className={`absolute top-0 right-0 m-2 px-2 py-1 rounded-md flex items-center space-x-1 text-xs ${
        isLockedByMe
          ? 'bg-blue-100 text-blue-700 border border-blue-300'
          : 'bg-yellow-100 text-yellow-700 border border-yellow-300'
      }`}
    >
      <Lock className="w-3 h-3" />
      <span>{isLockedByMe ? 'Editing' : `${lock.user_name} editing`}</span>
    </div>
  )
}
