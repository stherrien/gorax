import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import {
  escapeCsvValue,
  convertToCSV,
  downloadCSV,
  formatDateForCSV,
  truncateForCSV,
} from './csvExport'

describe('csvExport', () => {
  describe('escapeCsvValue', () => {
    it('should return empty string for null', () => {
      expect(escapeCsvValue(null)).toBe('')
    })

    it('should return empty string for undefined', () => {
      expect(escapeCsvValue(undefined)).toBe('')
    })

    it('should convert numbers to strings', () => {
      expect(escapeCsvValue(123)).toBe('123')
      expect(escapeCsvValue(45.67)).toBe('45.67')
    })

    it('should convert booleans to strings', () => {
      expect(escapeCsvValue(true)).toBe('true')
      expect(escapeCsvValue(false)).toBe('false')
    })

    it('should leave simple strings unchanged', () => {
      expect(escapeCsvValue('hello')).toBe('hello')
      expect(escapeCsvValue('test123')).toBe('test123')
    })

    it('should wrap strings with commas in quotes', () => {
      expect(escapeCsvValue('hello, world')).toBe('"hello, world"')
      expect(escapeCsvValue('a,b,c')).toBe('"a,b,c"')
    })

    it('should wrap strings with newlines in quotes', () => {
      expect(escapeCsvValue('line1\nline2')).toBe('"line1\nline2"')
    })

    it('should escape quotes by doubling them', () => {
      expect(escapeCsvValue('He said "hello"')).toBe('"He said ""hello"""')
    })

    it('should handle strings with both commas and quotes', () => {
      expect(escapeCsvValue('Hello, "world"')).toBe('"Hello, ""world"""')
    })
  })

  describe('convertToCSV', () => {
    it('should convert simple array to CSV', () => {
      const data = [
        { id: '1', name: 'Alice' },
        { id: '2', name: 'Bob' },
      ]
      const csv = convertToCSV(data, ['id', 'name'])

      expect(csv).toBe('id,name\n1,Alice\n2,Bob')
    })

    it('should use custom header labels', () => {
      const data = [
        { id: '1', name: 'Alice' },
        { id: '2', name: 'Bob' },
      ]
      const csv = convertToCSV(data, ['id', 'name'], ['ID', 'Name'])

      expect(csv).toBe('ID,Name\n1,Alice\n2,Bob')
    })

    it('should handle missing values', () => {
      const data = [
        { id: '1', name: 'Alice', age: 30 },
        { id: '2', name: 'Bob' },
      ]
      const csv = convertToCSV(data, ['id', 'name', 'age'])

      expect(csv).toBe('id,name,age\n1,Alice,30\n2,Bob,')
    })

    it('should escape values with commas', () => {
      const data = [
        { id: '1', description: 'Hello, world' },
      ]
      const csv = convertToCSV(data, ['id', 'description'])

      expect(csv).toBe('id,description\n1,"Hello, world"')
    })

    it('should escape values with quotes', () => {
      const data = [
        { id: '1', message: 'He said "hello"' },
      ]
      const csv = convertToCSV(data, ['id', 'message'])

      expect(csv).toBe('id,message\n1,"He said ""hello"""')
    })

    it('should handle empty array', () => {
      const data: { id: string; name: string }[] = []
      const csv = convertToCSV(data, ['id', 'name'])

      expect(csv).toBe('id,name')
    })

    it('should handle numeric values', () => {
      const data = [
        { id: 1, value: 123.45 },
        { id: 2, value: 67 },
      ]
      const csv = convertToCSV(data, ['id', 'value'])

      expect(csv).toBe('id,value\n1,123.45\n2,67')
    })
  })

  describe('downloadCSV', () => {
    let createElementSpy: any
    let createObjectURLSpy: any
    let revokeObjectURLSpy: any
    let appendChildSpy: any
    let removeChildSpy: any
    let mockAnchor: any

    beforeEach(() => {
      vi.useFakeTimers()

      mockAnchor = {
        href: '',
        download: '',
        click: vi.fn(),
        style: {},
        parentNode: null as unknown,
      }

      createElementSpy = vi.spyOn(document, 'createElement').mockReturnValue(mockAnchor as any)
      createObjectURLSpy = vi.spyOn(URL, 'createObjectURL').mockReturnValue('blob:mock-url')
      revokeObjectURLSpy = vi.spyOn(URL, 'revokeObjectURL').mockImplementation(() => {})
      appendChildSpy = vi.spyOn(document.body, 'appendChild').mockImplementation((node) => {
        // Simulate DOM attachment by setting parentNode
        mockAnchor.parentNode = document.body
        return node
      })
      removeChildSpy = vi.spyOn(document.body, 'removeChild').mockImplementation((node) => {
        mockAnchor.parentNode = null
        return node
      })
    })

    afterEach(() => {
      vi.useRealTimers()
      createElementSpy.mockRestore()
      createObjectURLSpy.mockRestore()
      revokeObjectURLSpy.mockRestore()
      appendChildSpy.mockRestore()
      removeChildSpy.mockRestore()
    })

    it('should create a blob with CSV content', () => {
      const csv = 'id,name\n1,Alice'
      downloadCSV(csv, 'test.csv')

      expect(createObjectURLSpy).toHaveBeenCalled()
      const blob = createObjectURLSpy.mock.calls[0][0]
      expect(blob).toBeInstanceOf(Blob)
      expect(blob.type).toBe('text/csv;charset=utf-8;')
    })

    it('should create an anchor element', () => {
      const csv = 'id,name\n1,Alice'
      downloadCSV(csv, 'test.csv')

      expect(createElementSpy).toHaveBeenCalledWith('a')
    })

    it('should trigger download with correct filename', () => {
      const csv = 'id,name\n1,Alice'

      downloadCSV(csv, 'export.csv')

      const mockAnchor = createElementSpy.mock.results[0].value
      expect(mockAnchor.download).toBe('export.csv')
      expect(mockAnchor.click).toHaveBeenCalled()
    })

    it('should append and remove anchor from DOM', () => {
      const csv = 'id,name\n1,Alice'
      downloadCSV(csv, 'test.csv')

      expect(appendChildSpy).toHaveBeenCalled()

      // Run the setTimeout callback to trigger cleanup
      vi.runAllTimers()

      expect(removeChildSpy).toHaveBeenCalled()
    })

    it('should revoke object URL after download', () => {
      const csv = 'id,name\n1,Alice'
      downloadCSV(csv, 'test.csv')

      // Run the setTimeout callback to trigger cleanup
      vi.runAllTimers()

      expect(revokeObjectURLSpy).toHaveBeenCalledWith('blob:mock-url')
    })
  })

  describe('formatDateForCSV', () => {
    it('should format valid date string to ISO', () => {
      const date = '2024-01-15T10:30:00Z'
      const result = formatDateForCSV(date)

      expect(result).toBe('2024-01-15T10:30:00.000Z')
    })

    it('should return N/A for null', () => {
      expect(formatDateForCSV(null)).toBe('N/A')
    })

    it('should return N/A for undefined', () => {
      expect(formatDateForCSV(undefined)).toBe('N/A')
    })

    it('should return N/A for empty string', () => {
      expect(formatDateForCSV('')).toBe('N/A')
    })
  })

  describe('truncateForCSV', () => {
    it('should leave short strings unchanged', () => {
      const str = 'Hello world'
      expect(truncateForCSV(str)).toBe('Hello world')
    })

    it('should truncate long strings to default 500 chars', () => {
      const str = 'a'.repeat(600)
      const result = truncateForCSV(str)

      expect(result.length).toBe(503) // 500 + '...'
      expect(result.endsWith('...')).toBe(true)
    })

    it('should truncate to custom max length', () => {
      const str = 'a'.repeat(100)
      const result = truncateForCSV(str, 50)

      expect(result.length).toBe(53) // 50 + '...'
      expect(result.endsWith('...')).toBe(true)
    })

    it('should not truncate string at exact max length', () => {
      const str = 'a'.repeat(500)
      const result = truncateForCSV(str, 500)

      expect(result).toBe(str)
      expect(result.endsWith('...')).toBe(false)
    })
  })
})
