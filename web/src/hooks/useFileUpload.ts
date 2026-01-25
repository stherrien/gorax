/**
 * Custom hook for handling workflow file uploads
 * Provides drag-and-drop and file input handling with validation and parsing
 */

import { useState, useCallback, useRef } from 'react'
import type { Node, Edge } from '@xyflow/react'
import {
  validateFile,
  type FileValidationResult,
  LARGE_FILE_THRESHOLD_BYTES,
} from '../utils/fileValidation'
import {
  parseWorkflowFile,
  validateWorkflowStructure,
  type WorkflowParseResult,
} from '../utils/workflowParser'

/**
 * Upload state
 */
export type UploadStatus = 'idle' | 'validating' | 'reading' | 'parsing' | 'success' | 'error'

/**
 * Upload result
 */
export interface UploadResult {
  status: UploadStatus
  progress: number
  fileName?: string
  fileSize?: number
  nodes?: Node[]
  edges?: Edge[]
  workflowName?: string
  workflowDescription?: string
  error?: string
  errorDetails?: string[]
  warnings?: string[]
}

/**
 * Hook return type
 */
export interface UseFileUploadReturn {
  /** Current upload state */
  uploadState: UploadResult
  /** Whether the user is currently dragging files over the drop zone */
  isDragging: boolean
  /** Process a file for upload */
  uploadFile: (file: File) => Promise<UploadResult>
  /** Process files from a drop event */
  handleDrop: (event: React.DragEvent) => Promise<void>
  /** Handle drag enter event */
  handleDragEnter: (event: React.DragEvent) => void
  /** Handle drag leave event */
  handleDragLeave: (event: React.DragEvent) => void
  /** Handle drag over event */
  handleDragOver: (event: React.DragEvent) => void
  /** Reset upload state */
  resetUpload: () => void
  /** Clear error state */
  clearError: () => void
  /** Accept the uploaded workflow (confirms user wants to use it) */
  acceptUpload: () => { nodes: Node[]; edges: Edge[]; name?: string; description?: string } | null
  /** Reference for file input element */
  fileInputRef: React.RefObject<HTMLInputElement | null>
  /** Handle file input change */
  handleFileInputChange: (event: React.ChangeEvent<HTMLInputElement>) => Promise<void>
  /** Open file picker dialog */
  openFilePicker: () => void
}

/**
 * Initial upload state
 */
const initialUploadState: UploadResult = {
  status: 'idle',
  progress: 0,
}

/**
 * Hook for handling workflow file uploads
 */
export function useFileUpload(
  options?: {
    onUploadStart?: () => void
    onUploadComplete?: (result: UploadResult) => void
    onUploadError?: (error: string) => void
  }
): UseFileUploadReturn {
  const { onUploadStart, onUploadComplete, onUploadError } = options || {}

  const [uploadState, setUploadState] = useState<UploadResult>(initialUploadState)
  const [isDragging, setIsDragging] = useState(false)
  const dragCounterRef = useRef(0)
  const fileInputRef = useRef<HTMLInputElement | null>(null)

  /**
   * Reset upload state
   */
  const resetUpload = useCallback(() => {
    setUploadState(initialUploadState)
    setIsDragging(false)
    dragCounterRef.current = 0
  }, [])

  /**
   * Clear error state
   */
  const clearError = useCallback(() => {
    setUploadState((prev) => ({
      ...prev,
      status: 'idle',
      error: undefined,
      errorDetails: undefined,
    }))
  }, [])

  /**
   * Accept the uploaded workflow
   */
  const acceptUpload = useCallback(() => {
    if (uploadState.status !== 'success' || !uploadState.nodes) {
      return null
    }

    const result = {
      nodes: uploadState.nodes,
      edges: uploadState.edges || [],
      name: uploadState.workflowName,
      description: uploadState.workflowDescription,
    }

    // Reset state after accepting
    resetUpload()

    return result
  }, [uploadState, resetUpload])

  /**
   * Process a file for upload
   */
  const uploadFile = useCallback(async (file: File): Promise<UploadResult> => {
    onUploadStart?.()

    // Start validation
    setUploadState({
      status: 'validating',
      progress: 0,
      fileName: file.name,
      fileSize: file.size,
    })

    // Validate file
    const validationResult: FileValidationResult = validateFile(file)
    if (!validationResult.valid) {
      const errorResult: UploadResult = {
        status: 'error',
        progress: 0,
        fileName: file.name,
        fileSize: file.size,
        error: validationResult.error || 'File validation failed',
      }
      setUploadState(errorResult)
      onUploadError?.(errorResult.error!)
      return errorResult
    }

    // Update state for reading/parsing
    const isLargeFile = file.size > LARGE_FILE_THRESHOLD_BYTES
    setUploadState({
      status: 'reading',
      progress: 10,
      fileName: file.name,
      fileSize: file.size,
    })

    // Parse file with progress tracking for large files
    const onProgress = isLargeFile
      ? (progress: number) => {
          setUploadState((prev) => ({
            ...prev,
            status: progress < 50 ? 'reading' : 'parsing',
            progress: 10 + Math.round(progress * 0.8), // Scale to 10-90%
          }))
        }
      : undefined

    const parseResult: WorkflowParseResult = await parseWorkflowFile(file, onProgress)

    if (!parseResult.success) {
      const errorResult: UploadResult = {
        status: 'error',
        progress: 0,
        fileName: file.name,
        fileSize: file.size,
        error: parseResult.error || 'Failed to parse workflow',
        errorDetails: parseResult.errorDetails,
      }
      setUploadState(errorResult)
      onUploadError?.(errorResult.error!)
      return errorResult
    }

    // Validate workflow structure
    const structureValidation = validateWorkflowStructure(
      parseResult.nodes!,
      parseResult.edges!
    )

    // Combine warnings
    const allWarnings = [
      ...(parseResult.warnings || []),
      ...structureValidation.warnings,
    ]

    // Success!
    const successResult: UploadResult = {
      status: 'success',
      progress: 100,
      fileName: file.name,
      fileSize: file.size,
      nodes: parseResult.nodes,
      edges: parseResult.edges,
      workflowName: parseResult.name,
      workflowDescription: parseResult.description,
      warnings: allWarnings.length > 0 ? allWarnings : undefined,
    }
    setUploadState(successResult)
    onUploadComplete?.(successResult)
    return successResult
  }, [onUploadStart, onUploadComplete, onUploadError])

  /**
   * Handle drag enter event
   */
  const handleDragEnter = useCallback((event: React.DragEvent) => {
    event.preventDefault()
    event.stopPropagation()
    dragCounterRef.current++

    // Only show drag state if dragging files
    if (event.dataTransfer.types.includes('Files')) {
      setIsDragging(true)
    }
  }, [])

  /**
   * Handle drag leave event
   */
  const handleDragLeave = useCallback((event: React.DragEvent) => {
    event.preventDefault()
    event.stopPropagation()
    dragCounterRef.current--

    if (dragCounterRef.current === 0) {
      setIsDragging(false)
    }
  }, [])

  /**
   * Handle drag over event
   */
  const handleDragOver = useCallback((event: React.DragEvent) => {
    event.preventDefault()
    event.stopPropagation()
    event.dataTransfer.dropEffect = 'copy'
  }, [])

  /**
   * Handle drop event
   */
  const handleDrop = useCallback(async (event: React.DragEvent) => {
    event.preventDefault()
    event.stopPropagation()
    setIsDragging(false)
    dragCounterRef.current = 0

    const files = event.dataTransfer.files
    if (files.length === 0) return

    // Process first file only (could extend to handle multiple)
    const file = files[0]
    await uploadFile(file)
  }, [uploadFile])

  /**
   * Handle file input change
   */
  const handleFileInputChange = useCallback(async (event: React.ChangeEvent<HTMLInputElement>) => {
    const files = event.target.files
    if (!files || files.length === 0) return

    const file = files[0]
    await uploadFile(file)

    // Reset input so same file can be selected again
    if (fileInputRef.current) {
      fileInputRef.current.value = ''
    }
  }, [uploadFile])

  /**
   * Open file picker dialog
   */
  const openFilePicker = useCallback(() => {
    fileInputRef.current?.click()
  }, [])

  return {
    uploadState,
    isDragging,
    uploadFile,
    handleDrop,
    handleDragEnter,
    handleDragLeave,
    handleDragOver,
    resetUpload,
    clearError,
    acceptUpload,
    fileInputRef,
    handleFileInputChange,
    openFilePicker,
  }
}

export default useFileUpload
