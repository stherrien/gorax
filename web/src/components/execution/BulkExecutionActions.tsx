import { useState } from 'react'
import { executionAPI } from '../../api/executions'
import { ConfirmBulkDialog } from '../common/ConfirmBulkDialog'

interface BulkExecutionActionsProps {
  selectedIds: string[]
  onSuccess: () => void
  onError: (message: string) => void
}

export function BulkExecutionActions({
  selectedIds,
  onSuccess,
  onError,
}: BulkExecutionActionsProps) {
  const [showDeleteDialog, setShowDeleteDialog] = useState(false)
  const [showRetryDialog, setShowRetryDialog] = useState(false)
  const [isProcessing, setIsProcessing] = useState(false)

  const handleDelete = async () => {
    setIsProcessing(true)
    try {
      const result = await executionAPI.bulkDelete(selectedIds)

      if (result.failed.length > 0) {
        const failedCount = result.failed.length
        const successCount = result.success.length
        onError(
          `Deleted ${successCount} executions. Failed to delete ${failedCount} executions.`
        )
      } else {
        onSuccess()
      }

      setShowDeleteDialog(false)
    } catch (error) {
      onError(error instanceof Error ? error.message : 'Failed to delete executions')
    } finally {
      setIsProcessing(false)
    }
  }

  const handleRetry = async () => {
    setIsProcessing(true)
    try {
      const result = await executionAPI.bulkRetry(selectedIds)

      if (result.failed.length > 0) {
        const failedCount = result.failed.length
        const successCount = result.success.length
        onError(
          `Retried ${successCount} executions. Failed to retry ${failedCount} executions.`
        )
      } else {
        onSuccess()
      }

      setShowRetryDialog(false)
    } catch (error) {
      onError(error instanceof Error ? error.message : 'Failed to retry executions')
    } finally {
      setIsProcessing(false)
    }
  }

  return (
    <>
      <button
        onClick={() => setShowRetryDialog(true)}
        disabled={isProcessing}
        className="px-3 py-1.5 bg-blue-600 text-white rounded text-sm hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
      >
        Retry Failed
      </button>

      <button
        onClick={() => setShowDeleteDialog(true)}
        disabled={isProcessing}
        className="px-3 py-1.5 bg-red-600 text-white rounded text-sm hover:bg-red-700 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
      >
        Delete
      </button>

      <ConfirmBulkDialog
        open={showDeleteDialog}
        action="delete"
        count={selectedIds.length}
        destructive={true}
        onConfirm={handleDelete}
        onCancel={() => setShowDeleteDialog(false)}
      />

      <ConfirmBulkDialog
        open={showRetryDialog}
        action="retry"
        count={selectedIds.length}
        message="This will create new executions for the selected failed executions."
        onConfirm={handleRetry}
        onCancel={() => setShowRetryDialog(false)}
      />
    </>
  )
}
