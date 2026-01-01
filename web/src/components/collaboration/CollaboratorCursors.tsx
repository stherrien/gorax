import { MousePointer2 } from 'lucide-react'
import type { UserPresence } from '../../types/collaboration'

interface CollaboratorCursorsProps {
  users: UserPresence[]
  currentUserId: string
  zoom?: number
}

export default function CollaboratorCursors({ users, currentUserId, zoom = 1 }: CollaboratorCursorsProps) {
  const otherUsers = users.filter((u) => u.user_id !== currentUserId && u.cursor)

  return (
    <>
      {otherUsers.map((user) => {
        if (!user.cursor) return null

        return (
          <div
            key={user.user_id}
            className="absolute pointer-events-none z-50 transition-transform duration-150"
            style={{
              left: `${user.cursor.x}px`,
              top: `${user.cursor.y}px`,
              transform: `scale(${1 / zoom})`,
            }}
          >
            <MousePointer2
              className="w-5 h-5"
              style={{ color: user.color }}
              fill={user.color}
            />
            <div
              className="mt-1 px-2 py-1 rounded text-xs text-white whitespace-nowrap shadow-lg"
              style={{ backgroundColor: user.color }}
            >
              {user.user_name}
            </div>
          </div>
        )
      })}
    </>
  )
}
