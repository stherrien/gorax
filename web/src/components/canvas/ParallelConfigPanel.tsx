interface ParallelConfig {
  errorStrategy: 'fail_fast' | 'wait_all'
  maxConcurrency: number
}

interface ParallelConfigPanelProps {
  config: ParallelConfig
  onChange: (config: ParallelConfig) => void
}

export default function ParallelConfigPanel({ config, onChange }: ParallelConfigPanelProps) {
  const handleChange = (field: keyof ParallelConfig, value: string | number) => {
    onChange({
      ...config,
      [field]: value,
    })
  }

  return (
    <div className="space-y-4 p-4">
      <h3 className="text-lg font-semibold text-white">Parallel Configuration</h3>

      {/* Error Strategy */}
      <div>
        <label htmlFor="parallel-error-strategy" className="block text-sm font-medium text-gray-300 mb-2">
          Error Strategy
        </label>
        <select
          id="parallel-error-strategy"
          value={config.errorStrategy}
          onChange={(e) => handleChange('errorStrategy', e.target.value as 'fail_fast' | 'wait_all')}
          className="w-full px-3 py-2 bg-gray-700 border border-gray-600 rounded-md text-white focus:outline-none focus:ring-2 focus:ring-purple-500"
        >
          <option value="fail_fast">Fail fast (stop on first error)</option>
          <option value="wait_all">Wait all (complete all branches)</option>
        </select>
        <p className="mt-1 text-xs text-gray-400">
          How to handle errors in parallel branches
        </p>
      </div>

      {/* Max Concurrency */}
      <div>
        <label htmlFor="parallel-max-concurrency" className="block text-sm font-medium text-gray-300 mb-2">
          Max Concurrency
        </label>
        <input
          id="parallel-max-concurrency"
          type="number"
          value={config.maxConcurrency}
          onChange={(e) => handleChange('maxConcurrency', parseInt(e.target.value) || 0)}
          min="0"
          max="100"
          className="w-full px-3 py-2 bg-gray-700 border border-gray-600 rounded-md text-white focus:outline-none focus:ring-2 focus:ring-purple-500"
        />
        <p className="mt-1 text-xs text-gray-400">
          Maximum branches running simultaneously (0 = unlimited)
        </p>
      </div>

      {/* Info Box */}
      <div className="p-3 bg-blue-900/20 border border-blue-500/30 rounded-md">
        <p className="text-sm text-blue-400">
          Connect multiple nodes to this parallel node to execute them concurrently.
          All branches will merge back after completion.
        </p>
      </div>
    </div>
  )
}
