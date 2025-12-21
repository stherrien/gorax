import { Handle, Position } from '@xyflow/react'

export interface LoopNodeData {
  label: string
  source?: string
  itemVariable?: string
  indexVariable?: string
  maxIterations?: number
  onError?: 'stop' | 'continue'
}

export interface LoopNodeProps {
  id: string
  data: LoopNodeData
  selected?: boolean
}

export default function LoopNode({ data, selected }: LoopNodeProps) {
  return (
    <div
      className={`
        px-4 py-3 rounded-lg shadow-lg min-w-[200px]
        bg-gradient-to-br from-purple-500 to-violet-600
        ${selected ? 'ring-2 ring-white ring-offset-2 ring-offset-gray-900' : ''}
      `}
    >
      {/* Input handle */}
      <Handle
        type="target"
        position={Position.Top}
        className="w-3 h-3 bg-white border-2 border-purple-500"
      />

      <div className="space-y-2">
        <div className="flex items-center space-x-2">
          <span className="text-lg">🔁</span>
          <div className="flex-1">
            <p className="text-white font-medium text-sm">{data.label}</p>
            <p className="text-white/70 text-xs">Loop</p>
          </div>
        </div>

        {/* Display configuration details */}
        {data.source && (
          <div className="mt-2 p-2 bg-white/10 rounded text-xs text-white/90 font-mono break-all">
            {data.source}
          </div>
        )}

        {(data.itemVariable || data.indexVariable) && (
          <div className="mt-1 space-y-1">
            {data.itemVariable && (
              <p className="text-white/80 text-xs">
                Item: <span className="font-mono text-white">{data.itemVariable}</span>
              </p>
            )}
            {data.indexVariable && (
              <p className="text-white/80 text-xs">
                Index: <span className="font-mono text-white">{data.indexVariable}</span>
              </p>
            )}
          </div>
        )}

        {data.maxIterations && data.maxIterations !== 1000 && (
          <p className="text-white/80 text-xs">
            Max: {data.maxIterations}
          </p>
        )}

        {data.onError && (
          <p className="text-white/80 text-xs">
            {data.onError === 'continue' ? 'Continue on error' : 'Stop on error'}
          </p>
        )}
      </div>

      {/* Output handle */}
      <Handle
        type="source"
        position={Position.Bottom}
        className="w-3 h-3 bg-white border-2 border-purple-500"
      />
    </div>
  )
}
