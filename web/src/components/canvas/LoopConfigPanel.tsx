interface LoopConfig {
  source: string
  itemVariable: string
  indexVariable: string
  maxIterations: number
  onError: 'stop' | 'continue'
}

interface LoopConfigPanelProps {
  config: LoopConfig
  onChange: (config: LoopConfig) => void
}

export default function LoopConfigPanel({ config, onChange }: LoopConfigPanelProps) {
  const handleChange = (field: keyof LoopConfig, value: string | number) => {
    onChange({
      ...config,
      [field]: value,
    })
  }

  return (
    <div className="space-y-4 p-4">
      <h3 className="text-lg font-semibold text-white">Loop Configuration</h3>

      {/* Source Array */}
      <div>
        <label htmlFor="loop-source" className="block text-sm font-medium text-gray-300 mb-2">
          Source Array
        </label>
        <input
          id="loop-source"
          type="text"
          value={config.source}
          onChange={(e) => handleChange('source', e.target.value)}
          placeholder="${steps.node_id.output.items}"
          className="w-full px-3 py-2 bg-gray-700 border border-gray-600 rounded-md text-white placeholder-gray-400 focus:outline-none focus:ring-2 focus:ring-purple-500"
        />
        <p className="mt-1 text-xs text-gray-400">
          Expression that resolves to an array (e.g., $&#123;steps.node_id.output.items&#125;)
        </p>
      </div>

      {/* Item Variable */}
      <div>
        <label htmlFor="loop-item-var" className="block text-sm font-medium text-gray-300 mb-2">
          Item Variable
        </label>
        <input
          id="loop-item-var"
          type="text"
          value={config.itemVariable}
          onChange={(e) => handleChange('itemVariable', e.target.value)}
          placeholder="item"
          className="w-full px-3 py-2 bg-gray-700 border border-gray-600 rounded-md text-white placeholder-gray-400 focus:outline-none focus:ring-2 focus:ring-purple-500"
        />
        <p className="mt-1 text-xs text-gray-400">
          Variable name to access current item in each iteration
        </p>
      </div>

      {/* Index Variable */}
      <div>
        <label htmlFor="loop-index-var" className="block text-sm font-medium text-gray-300 mb-2">
          Index Variable
        </label>
        <input
          id="loop-index-var"
          type="text"
          value={config.indexVariable}
          onChange={(e) => handleChange('indexVariable', e.target.value)}
          placeholder="index"
          className="w-full px-3 py-2 bg-gray-700 border border-gray-600 rounded-md text-white placeholder-gray-400 focus:outline-none focus:ring-2 focus:ring-purple-500"
        />
        <p className="mt-1 text-xs text-gray-400">
          Optional variable name to access current index (0-based)
        </p>
      </div>

      {/* Max Iterations */}
      <div>
        <label htmlFor="loop-max-iter" className="block text-sm font-medium text-gray-300 mb-2">
          Max Iterations
        </label>
        <input
          id="loop-max-iter"
          type="number"
          value={config.maxIterations}
          onChange={(e) => handleChange('maxIterations', parseInt(e.target.value))}
          min="1"
          max="10000"
          className="w-full px-3 py-2 bg-gray-700 border border-gray-600 rounded-md text-white focus:outline-none focus:ring-2 focus:ring-purple-500"
        />
        <p className="mt-1 text-xs text-gray-400">
          Safety limit for maximum number of iterations (default: 1000)
        </p>
      </div>

      {/* Error Strategy */}
      <div>
        <label htmlFor="loop-error-strategy" className="block text-sm font-medium text-gray-300 mb-2">
          Error Strategy
        </label>
        <select
          id="loop-error-strategy"
          value={config.onError}
          onChange={(e) => handleChange('onError', e.target.value as 'stop' | 'continue')}
          className="w-full px-3 py-2 bg-gray-700 border border-gray-600 rounded-md text-white focus:outline-none focus:ring-2 focus:ring-purple-500"
        >
          <option value="stop">Stop on error</option>
          <option value="continue">Continue on error</option>
        </select>
        <p className="mt-1 text-xs text-gray-400">
          How to handle errors in each iteration
        </p>
      </div>
    </div>
  )
}
