import { Users, Wifi, WifiOff } from 'lucide-react'
import type { UserPresence } from '../../types/collaboration'

interface UserPresenceIndicatorProps {
  connected: boolean
  users: UserPresence[]
  currentUserId: string
}

export default function UserPresenceIndicator({ connected, users, currentUserId }: UserPresenceIndicatorProps) {
  const otherUsers = users.filter((u) => u.user_id !== currentUserId)

  return (
    <div className="flex items-center space-x-2 px-3 py-2 bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg shadow-sm">
      {connected ? (
        <Wifi className="w-4 h-4 text-green-500" />
      ) : (
        <WifiOff className="w-4 h-4 text-gray-400" />
      )}

      <div className="flex items-center space-x-1">
        <Users className="w-4 h-4 text-gray-500" />
        <span className="text-sm text-gray-700 dark:text-gray-300">
          {otherUsers.length === 0 ? 'Only you' : `${otherUsers.length + 1} editing`}
        </span>
      </div>

      {otherUsers.length > 0 && (
        <div className="flex -space-x-2">
          {otherUsers.slice(0, 3).map((user) => (
            <div
              key={user.user_id}
              className="w-6 h-6 rounded-full border-2 border-white dark:border-gray-800 flex items-center justify-center text-xs font-medium text-white"
              style={{ backgroundColor: user.color }}
              title={user.user_name}
            >
              {user.user_name.charAt(0).toUpperCase()}
            </div>
          ))}
          {otherUsers.length > 3 && (
            <div className="w-6 h-6 rounded-full border-2 border-white dark:border-gray-800 bg-gray-400 flex items-center justify-center text-xs font-medium text-white">
              +{otherUsers.length - 3}
            </div>
          )}
        </div>
      )}
    </div>
  )
}
