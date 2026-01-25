/**
 * File validation utilities for workflow file uploads
 * Validates file types, sizes, and content structure
 */

// Supported file extensions
export const SUPPORTED_EXTENSIONS = ['.json', '.yaml', '.yml'] as const
export type SupportedExtension = typeof SUPPORTED_EXTENSIONS[number]

// MIME types for supported files
export const SUPPORTED_MIME_TYPES = [
  'application/json',
  'text/json',
  'application/x-yaml',
  'text/yaml',
  'text/x-yaml',
  'application/yaml',
  'text/plain', // Some browsers report YAML as text/plain
] as const

// File size limits
export const MAX_FILE_SIZE_BYTES = 10 * 1024 * 1024 // 10MB
export const LARGE_FILE_THRESHOLD_BYTES = 1 * 1024 * 1024 // 1MB - show progress for larger files

/**
 * Result of file validation
 */
export interface FileValidationResult {
  valid: boolean
  error?: string
  file?: File
  extension?: SupportedExtension
  isLargeFile?: boolean
}

/**
 * Get file extension from filename
 */
export function getFileExtension(filename: string): string {
  const lastDot = filename.lastIndexOf('.')
  if (lastDot === -1) return ''
  return filename.slice(lastDot).toLowerCase()
}

/**
 * Check if a file extension is supported
 */
export function isSupportedExtension(extension: string): extension is SupportedExtension {
  return SUPPORTED_EXTENSIONS.includes(extension as SupportedExtension)
}

/**
 * Validate a single file for workflow upload
 */
export function validateFile(file: File): FileValidationResult {
  // Check file extension
  const extension = getFileExtension(file.name)
  if (!extension) {
    return {
      valid: false,
      error: 'File has no extension. Please upload a .json, .yaml, or .yml file.',
    }
  }

  if (!isSupportedExtension(extension)) {
    return {
      valid: false,
      error: `Unsupported file type: ${extension}. Please upload a .json, .yaml, or .yml file.`,
    }
  }

  // Check file size
  if (file.size > MAX_FILE_SIZE_BYTES) {
    const maxSizeMB = MAX_FILE_SIZE_BYTES / (1024 * 1024)
    const fileSizeMB = (file.size / (1024 * 1024)).toFixed(2)
    return {
      valid: false,
      error: `File is too large (${fileSizeMB}MB). Maximum file size is ${maxSizeMB}MB.`,
    }
  }

  // Check MIME type (less strict due to browser inconsistencies)
  // We allow text/plain because some browsers report YAML as plain text
  const mimeType = file.type
  if (mimeType && !SUPPORTED_MIME_TYPES.includes(mimeType as typeof SUPPORTED_MIME_TYPES[number])) {
    // Log warning but don't fail - rely on extension
    console.warn(`[fileValidation] Unexpected MIME type: ${mimeType} for ${file.name}`)
  }

  return {
    valid: true,
    file,
    extension,
    isLargeFile: file.size > LARGE_FILE_THRESHOLD_BYTES,
  }
}

/**
 * Validate multiple files
 * Returns validation results for all files
 */
export function validateFiles(files: FileList | File[]): FileValidationResult[] {
  const fileArray = Array.from(files)
  return fileArray.map(validateFile)
}

/**
 * Filter valid files from a list
 */
export function getValidFiles(files: FileList | File[]): File[] {
  const results = validateFiles(files)
  return results
    .filter((result) => result.valid && result.file)
    .map((result) => result.file!)
}

/**
 * Check if DataTransfer contains files (for drag events)
 */
export function containsFiles(dataTransfer: DataTransfer): boolean {
  if (dataTransfer.items) {
    return Array.from(dataTransfer.items).some(
      (item) => item.kind === 'file'
    )
  }
  return dataTransfer.files.length > 0
}

/**
 * Check if DataTransfer contains supported file types
 */
export function containsSupportedFiles(dataTransfer: DataTransfer): boolean {
  if (dataTransfer.items) {
    return Array.from(dataTransfer.items).some((item) => {
      if (item.kind !== 'file') return false
      // Can't check extension during dragover, only type
      // Allow if we can't determine type (will validate on drop)
      if (!item.type) return true
      return SUPPORTED_MIME_TYPES.includes(item.type as typeof SUPPORTED_MIME_TYPES[number])
    })
  }
  return true // Allow by default, validate on drop
}

/**
 * Read file as text
 */
export function readFileAsText(file: File): Promise<string> {
  return new Promise((resolve, reject) => {
    const reader = new FileReader()

    reader.onload = () => {
      if (typeof reader.result === 'string') {
        resolve(reader.result)
      } else {
        reject(new Error('Failed to read file as text'))
      }
    }

    reader.onerror = () => {
      reject(new Error(`Failed to read file: ${reader.error?.message || 'Unknown error'}`))
    }

    reader.readAsText(file)
  })
}

/**
 * Read file with progress callback for large files
 */
export function readFileWithProgress(
  file: File,
  onProgress?: (progress: number) => void
): Promise<string> {
  return new Promise((resolve, reject) => {
    const reader = new FileReader()

    reader.onprogress = (event) => {
      if (event.lengthComputable && onProgress) {
        const progress = Math.round((event.loaded / event.total) * 100)
        onProgress(progress)
      }
    }

    reader.onload = () => {
      if (typeof reader.result === 'string') {
        onProgress?.(100)
        resolve(reader.result)
      } else {
        reject(new Error('Failed to read file as text'))
      }
    }

    reader.onerror = () => {
      reject(new Error(`Failed to read file: ${reader.error?.message || 'Unknown error'}`))
    }

    reader.readAsText(file)
  })
}

/**
 * Format file size for display
 */
export function formatFileSize(bytes: number): string {
  if (bytes === 0) return '0 Bytes'

  const k = 1024
  const sizes = ['Bytes', 'KB', 'MB', 'GB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))

  return `${parseFloat((bytes / Math.pow(k, i)).toFixed(2))} ${sizes[i]}`
}
