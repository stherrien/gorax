import {
  LineChart,
  Line,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  Legend,
  ResponsiveContainer,
} from 'recharts'
import type { TimeSeriesPoint } from '../../types/analytics'

interface ExecutionTrendChartProps {
  data: TimeSeriesPoint[]
}

export default function ExecutionTrendChart({ data }: ExecutionTrendChartProps) {
  const chartData = data.map((point) => ({
    timestamp: new Date(point.timestamp).toLocaleDateString(),
    executions: point.executionCount,
    successful: point.successCount,
    failed: point.failureCount,
  }))

  return (
    <ResponsiveContainer width="100%" height={300}>
      <LineChart data={chartData}>
        <CartesianGrid strokeDasharray="3 3" />
        <XAxis dataKey="timestamp" />
        <YAxis />
        <Tooltip />
        <Legend />
        <Line
          type="monotone"
          dataKey="executions"
          stroke="#6366f1"
          strokeWidth={2}
          name="Total Executions"
        />
        <Line
          type="monotone"
          dataKey="successful"
          stroke="#10b981"
          strokeWidth={2}
          name="Successful"
        />
        <Line
          type="monotone"
          dataKey="failed"
          stroke="#ef4444"
          strokeWidth={2}
          name="Failed"
        />
      </LineChart>
    </ResponsiveContainer>
  )
}
