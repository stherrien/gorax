import type { UserPresence } from '../../types/collaboration'

interface CollaboratorListProps {
  users: UserPresence[]
  currentUserId: string
}

export default function CollaboratorList({ users, currentUserId }: CollaboratorListProps) {
  return (
    <div className="flex flex-col space-y-2">
      <h3 className="text-sm font-semibold text-gray-700 dark:text-gray-300">Active Collaborators</h3>
      <div className="flex flex-col space-y-1">
        {users.map((user) => (
          <div key={user.user_id} className="flex items-center space-x-2 p-2 rounded hover:bg-gray-50 dark:hover:bg-gray-800">
            <div
              className="w-3 h-3 rounded-full"
              style={{ backgroundColor: user.color }}
              title={user.user_name}
            />
            <span className="text-sm text-gray-700 dark:text-gray-300">
              {user.user_name} {user.user_id === currentUserId && '(You)'}
            </span>
          </div>
        ))}
      </div>
    </div>
  )
}
