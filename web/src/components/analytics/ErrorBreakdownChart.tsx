import { PieChart, Pie, Cell, ResponsiveContainer, Tooltip, Legend } from 'recharts'
import type { ErrorInfo } from '../../types/analytics'

interface ErrorBreakdownChartProps {
  errors: ErrorInfo[]
}

interface ChartDataItem {
  name: string
  value: number
  percentage: number
  [key: string]: string | number  // Allow index signature for Recharts compatibility
}

const COLORS = ['#ef4444', '#f97316', '#f59e0b', '#eab308', '#84cc16']

export default function ErrorBreakdownChart({ errors }: ErrorBreakdownChartProps) {
  const chartData: ChartDataItem[] = errors.slice(0, 5).map((error) => ({
    name: error.errorMessage.substring(0, 30) + '...',
    value: error.errorCount,
    percentage: error.percentage,
  }))

  const renderLabel = (props: { payload?: ChartDataItem }) => {
    if (props.payload) {
      return `${props.payload.percentage.toFixed(1)}%`
    }
    return ''
  }

  return (
    <ResponsiveContainer width="100%" height={300}>
      <PieChart>
        <Pie
          data={chartData}
          cx="50%"
          cy="50%"
          labelLine={false}
          label={renderLabel}
          outerRadius={80}
          fill="#8884d8"
          dataKey="value"
        >
          {chartData.map((_, index) => (
            <Cell key={`cell-${index}`} fill={COLORS[index % COLORS.length]} />
          ))}
        </Pie>
        <Tooltip />
        <Legend />
      </PieChart>
    </ResponsiveContainer>
  )
}
