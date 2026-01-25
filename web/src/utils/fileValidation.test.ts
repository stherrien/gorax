import { describe, it, expect, vi } from 'vitest'
import {
  validateFile,
  validateFiles,
  getValidFiles,
  getFileExtension,
  isSupportedExtension,
  containsFiles,
  formatFileSize,
  readFileAsText,
  MAX_FILE_SIZE_BYTES,
  LARGE_FILE_THRESHOLD_BYTES,
  SUPPORTED_EXTENSIONS,
} from './fileValidation'

describe('fileValidation', () => {
  describe('getFileExtension', () => {
    it('should return lowercase extension for standard files', () => {
      expect(getFileExtension('workflow.json')).toBe('.json')
      expect(getFileExtension('workflow.YAML')).toBe('.yaml')
      expect(getFileExtension('workflow.YML')).toBe('.yml')
    })

    it('should return empty string for files without extension', () => {
      expect(getFileExtension('workflow')).toBe('')
      expect(getFileExtension('README')).toBe('')
    })

    it('should handle files with multiple dots', () => {
      expect(getFileExtension('my.workflow.json')).toBe('.json')
      expect(getFileExtension('test.spec.yaml')).toBe('.yaml')
    })

    it('should handle hidden files', () => {
      expect(getFileExtension('.gitignore')).toBe('.gitignore')
      expect(getFileExtension('.env')).toBe('.env')
    })
  })

  describe('isSupportedExtension', () => {
    it('should return true for supported extensions', () => {
      expect(isSupportedExtension('.json')).toBe(true)
      expect(isSupportedExtension('.yaml')).toBe(true)
      expect(isSupportedExtension('.yml')).toBe(true)
    })

    it('should return false for unsupported extensions', () => {
      expect(isSupportedExtension('.txt')).toBe(false)
      expect(isSupportedExtension('.xml')).toBe(false)
      expect(isSupportedExtension('.csv')).toBe(false)
      expect(isSupportedExtension('')).toBe(false)
    })
  })

  describe('validateFile', () => {
    function createMockFile(name: string, size: number, type: string = ''): File {
      const blob = new Blob([''], { type })
      return new File([blob], name, { type })
    }

    it('should accept valid JSON files', () => {
      const file = createMockFile('workflow.json', 1000, 'application/json')
      const result = validateFile(file)

      expect(result.valid).toBe(true)
      expect(result.file).toBe(file)
      expect(result.extension).toBe('.json')
    })

    it('should accept valid YAML files', () => {
      const yamlFile = createMockFile('workflow.yaml', 1000, 'text/yaml')
      expect(validateFile(yamlFile).valid).toBe(true)

      const ymlFile = createMockFile('workflow.yml', 1000, 'text/yaml')
      expect(validateFile(ymlFile).valid).toBe(true)
    })

    it('should reject files without extension', () => {
      const file = createMockFile('workflow', 1000)
      const result = validateFile(file)

      expect(result.valid).toBe(false)
      expect(result.error).toContain('no extension')
    })

    it('should reject unsupported file types', () => {
      const file = createMockFile('workflow.txt', 1000, 'text/plain')
      const result = validateFile(file)

      expect(result.valid).toBe(false)
      expect(result.error).toContain('Unsupported file type')
      expect(result.error).toContain('.txt')
    })

    it('should reject files that are too large', () => {
      const largeFile = createMockFile('workflow.json', MAX_FILE_SIZE_BYTES + 1, 'application/json')
      Object.defineProperty(largeFile, 'size', { value: MAX_FILE_SIZE_BYTES + 1 })

      const result = validateFile(largeFile)

      expect(result.valid).toBe(false)
      expect(result.error).toContain('too large')
    })

    it('should mark large files correctly', () => {
      const smallFile = createMockFile('small.json', 100)
      expect(validateFile(smallFile).isLargeFile).toBe(false)

      const largeFile = createMockFile('large.json', LARGE_FILE_THRESHOLD_BYTES + 1)
      Object.defineProperty(largeFile, 'size', { value: LARGE_FILE_THRESHOLD_BYTES + 1 })
      expect(validateFile(largeFile).isLargeFile).toBe(true)
    })
  })

  describe('validateFiles', () => {
    function createMockFile(name: string): File {
      return new File([''], name)
    }

    it('should validate multiple files', () => {
      const files = [
        createMockFile('valid.json'),
        createMockFile('invalid.txt'),
        createMockFile('another.yaml'),
      ]

      const results = validateFiles(files)

      expect(results).toHaveLength(3)
      expect(results[0].valid).toBe(true)
      expect(results[1].valid).toBe(false)
      expect(results[2].valid).toBe(true)
    })
  })

  describe('getValidFiles', () => {
    function createMockFile(name: string): File {
      return new File([''], name)
    }

    it('should return only valid files', () => {
      const files = [
        createMockFile('valid.json'),
        createMockFile('invalid.txt'),
        createMockFile('another.yaml'),
      ]

      const validFiles = getValidFiles(files)

      expect(validFiles).toHaveLength(2)
      expect(validFiles.map((f) => f.name)).toEqual(['valid.json', 'another.yaml'])
    })
  })

  describe('containsFiles', () => {
    it('should return true when dataTransfer has files', () => {
      const dataTransfer = {
        items: [
          { kind: 'file', type: 'application/json' },
        ],
        files: [],
      } as unknown as DataTransfer

      expect(containsFiles(dataTransfer)).toBe(true)
    })

    it('should return false when no files in items', () => {
      const dataTransfer = {
        items: [
          { kind: 'string', type: 'text/plain' },
        ],
        files: [],
      } as unknown as DataTransfer

      expect(containsFiles(dataTransfer)).toBe(false)
    })

    it('should check files length as fallback', () => {
      const dataTransfer = {
        files: { length: 1 },
      } as unknown as DataTransfer

      expect(containsFiles(dataTransfer)).toBe(true)
    })
  })

  describe('formatFileSize', () => {
    it('should format bytes correctly', () => {
      expect(formatFileSize(0)).toBe('0 Bytes')
      expect(formatFileSize(500)).toBe('500 Bytes')
    })

    it('should format kilobytes correctly', () => {
      expect(formatFileSize(1024)).toBe('1 KB')
      expect(formatFileSize(2048)).toBe('2 KB')
      expect(formatFileSize(1536)).toBe('1.5 KB')
    })

    it('should format megabytes correctly', () => {
      expect(formatFileSize(1024 * 1024)).toBe('1 MB')
      expect(formatFileSize(1024 * 1024 * 5.5)).toBe('5.5 MB')
    })

    it('should format gigabytes correctly', () => {
      expect(formatFileSize(1024 * 1024 * 1024)).toBe('1 GB')
    })
  })

  describe('readFileAsText', () => {
    it('should read file content as text', async () => {
      const content = '{"nodes": [], "edges": []}'
      const file = new File([content], 'workflow.json', { type: 'application/json' })

      const result = await readFileAsText(file)

      expect(result).toBe(content)
    })

    it('should handle empty files', async () => {
      const file = new File([''], 'empty.json')

      const result = await readFileAsText(file)

      expect(result).toBe('')
    })
  })

  describe('SUPPORTED_EXTENSIONS', () => {
    it('should include all expected extensions', () => {
      expect(SUPPORTED_EXTENSIONS).toContain('.json')
      expect(SUPPORTED_EXTENSIONS).toContain('.yaml')
      expect(SUPPORTED_EXTENSIONS).toContain('.yml')
      expect(SUPPORTED_EXTENSIONS).toHaveLength(3)
    })
  })
})
