/**
 * Escapes a value for CSV format by:
 * 1. Converting to string
 * 2. Wrapping in quotes if it contains comma, newline, or quote
 * 3. Escaping internal quotes by doubling them
 */
export function escapeCsvValue(value: unknown): string {
  if (value === null || value === undefined) {
    return ''
  }

  const str = String(value)

  if (str.includes(',') || str.includes('\n') || str.includes('"')) {
    return `"${str.replace(/"/g, '""')}"`
  }

  return str
}

/**
 * Converts an array of objects to CSV format
 * @param data Array of objects to convert
 * @param headers Array of header names (keys to extract from objects)
 * @param headerLabels Optional custom labels for headers
 * @returns CSV string
 */
export function convertToCSV<T extends Record<string, unknown>>(
  data: T[],
  headers: (keyof T)[],
  headerLabels?: string[]
): string {
  const labels = headerLabels || headers.map(h => String(h))

  const csvRows: string[] = []

  // Add header row
  csvRows.push(labels.map(label => escapeCsvValue(label)).join(','))

  // Add data rows
  for (const row of data) {
    const values = headers.map(header => {
      const value = row[header]
      return escapeCsvValue(value)
    })
    csvRows.push(values.join(','))
  }

  return csvRows.join('\n')
}

/**
 * Downloads a CSV file to the user's device
 * @param csv CSV content string
 * @param filename Name of the file to download
 */
export function downloadCSV(csv: string, filename: string): void {
  const blob = new Blob([csv], { type: 'text/csv;charset=utf-8;' })
  const url = URL.createObjectURL(blob)

  const link = document.createElement('a')
  link.href = url
  link.download = filename
  link.style.display = 'none'

  document.body.appendChild(link)
  link.click()

  // Use timeout to ensure click event has processed before cleanup
  setTimeout(() => {
    // Check if link is still attached before removing (handles test environments)
    if (link.parentNode === document.body) {
      document.body.removeChild(link)
    }
    URL.revokeObjectURL(url)
  }, 0)
}

/**
 * Formats a date to ISO string or returns 'N/A' if null/undefined
 */
export function formatDateForCSV(date: string | null | undefined): string {
  if (!date) return 'N/A'
  return new Date(date).toISOString()
}

/**
 * Truncates a string to a maximum length for CSV export
 */
export function truncateForCSV(value: string, maxLength: number = 500): string {
  if (value.length <= maxLength) {
    return value
  }
  return value.substring(0, maxLength) + '...'
}
