import { useState } from 'react'
import { workflowAPI } from '../api/workflows'
import type {
  BulkOperationResult,
  BulkExportResponse,
  BulkCloneResponse,
} from '../api/workflows'

/**
 * Hook for bulk workflow operations
 */
export function useBulkWorkflows() {
  const [bulkDeleting, setBulkDeleting] = useState(false)
  const [bulkEnabling, setBulkEnabling] = useState(false)
  const [bulkDisabling, setBulkDisabling] = useState(false)
  const [bulkExporting, setBulkExporting] = useState(false)
  const [bulkCloning, setBulkCloning] = useState(false)

  const bulkDelete = async (workflowIds: string[]): Promise<BulkOperationResult> => {
    try {
      setBulkDeleting(true)
      const result = await workflowAPI.bulkDelete(workflowIds)
      return result
    } finally {
      setBulkDeleting(false)
    }
  }

  const bulkEnable = async (workflowIds: string[]): Promise<BulkOperationResult> => {
    try {
      setBulkEnabling(true)
      const result = await workflowAPI.bulkEnable(workflowIds)
      return result
    } finally {
      setBulkEnabling(false)
    }
  }

  const bulkDisable = async (workflowIds: string[]): Promise<BulkOperationResult> => {
    try {
      setBulkDisabling(true)
      const result = await workflowAPI.bulkDisable(workflowIds)
      return result
    } finally {
      setBulkDisabling(false)
    }
  }

  const bulkExport = async (workflowIds: string[]): Promise<BulkExportResponse> => {
    try {
      setBulkExporting(true)
      const result = await workflowAPI.bulkExport(workflowIds)

      // Trigger download
      const dataStr = JSON.stringify(result.export, null, 2)
      const dataBlob = new Blob([dataStr], { type: 'application/json' })
      const url = URL.createObjectURL(dataBlob)
      const link = document.createElement('a')
      link.href = url
      link.download = `workflows-export-${new Date().toISOString().split('T')[0]}.json`
      document.body.appendChild(link)
      link.click()
      document.body.removeChild(link)
      URL.revokeObjectURL(url)

      return result
    } finally {
      setBulkExporting(false)
    }
  }

  const bulkClone = async (workflowIds: string[]): Promise<BulkCloneResponse> => {
    try {
      setBulkCloning(true)
      const result = await workflowAPI.bulkClone(workflowIds)
      return result
    } finally {
      setBulkCloning(false)
    }
  }

  return {
    bulkDelete,
    bulkEnable,
    bulkDisable,
    bulkExport,
    bulkClone,
    bulkDeleting,
    bulkEnabling,
    bulkDisabling,
    bulkExporting,
    bulkCloning,
    isLoading: bulkDeleting || bulkEnabling || bulkDisabling || bulkExporting || bulkCloning,
  }
}
