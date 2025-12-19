interface TestHistory {
  id: string
  timestamp: Date
  method: string
  headers: Record<string, string>
  body: unknown
  response: { statusCode: number } | null
  error: string | null
}

interface HistoryListProps {
  history: TestHistory[]
  onSelectItem: (item: TestHistory) => void
  getStatusCodeColor: (statusCode: number) => string
  formatTimestamp: (date: Date) => string
}

export default function HistoryList({
  history,
  onSelectItem,
  getStatusCodeColor,
  formatTimestamp,
}: HistoryListProps) {
  if (history.length === 0) {
    return null
  }

  return (
    <div className="border-t border-gray-700 pt-6">
      <h3 className="text-white text-lg font-semibold mb-4">History</h3>
      <ul className="space-y-2">
        {history.map(item => (
          <li
            key={item.id}
            onClick={() => onSelectItem(item)}
            className="p-3 bg-gray-700 rounded-lg cursor-pointer hover:bg-gray-600 transition-colors"
          >
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-3">
                <span className="text-white text-sm font-medium">{item.method}</span>
                {item.response && (
                  <span className={`text-sm ${getStatusCodeColor(item.response.statusCode)}`}>
                    {item.response.statusCode}
                  </span>
                )}
                {item.error && <span className="text-red-400 text-sm">Error</span>}
              </div>
              <span className="text-gray-400 text-xs">{formatTimestamp(item.timestamp)}</span>
            </div>
          </li>
        ))}
      </ul>
    </div>
  )
}
