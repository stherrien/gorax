interface SuccessRateGaugeProps {
  successRate: number
  label?: string
}

export default function SuccessRateGauge({
  successRate,
  label = 'Success Rate',
}: SuccessRateGaugeProps) {
  const percentage = Math.round(successRate * 100)
  const color =
    percentage >= 90 ? 'green' : percentage >= 70 ? 'yellow' : 'red'

  const colorClasses = {
    green: 'text-green-600 bg-green-100',
    yellow: 'text-yellow-600 bg-yellow-100',
    red: 'text-red-600 bg-red-100',
  }

  return (
    <div className="flex flex-col items-center">
      <div
        className={`w-32 h-32 rounded-full flex items-center justify-center ${colorClasses[color]}`}
      >
        <div className="text-center">
          <div className="text-4xl font-bold">{percentage}%</div>
          <div className="text-xs uppercase font-medium">{label}</div>
        </div>
      </div>
    </div>
  )
}
