import { useState } from 'react'
import { Trash2, Play, Pause, Download, Copy, Loader2 } from 'lucide-react'
import { useBulkWorkflows } from '../../hooks/useBulkWorkflows'
import { ConfirmBulkDialog } from '../common/ConfirmBulkDialog'
import type { BulkOperationResult } from '../../api/workflows'

interface BulkActionsToolbarProps {
  selectedCount: number
  selectedWorkflowIds: string[]
  onClearSelection: () => void
  onOperationComplete: () => void
}

export function BulkActionsToolbar({
  selectedCount,
  selectedWorkflowIds,
  onClearSelection,
  onOperationComplete,
}: BulkActionsToolbarProps) {
  const {
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
  } = useBulkWorkflows()

  const [confirmDialog, setConfirmDialog] = useState<{
    open: boolean
    action: 'delete' | 'enable' | 'disable' | 'clone' | null
  }>({
    open: false,
    action: null,
  })

  const [toast, setToast] = useState<{
    type: 'success' | 'error'
    message: string
  } | null>(null)

  const showToast = (type: 'success' | 'error', message: string) => {
    setToast({ type, message })
    setTimeout(() => setToast(null), 5000)
  }

  const handleOperationResult = (result: BulkOperationResult, action: string) => {
    if (result.failures.length === 0) {
      showToast('success', `Successfully ${action} ${result.success_count} workflow(s)`)
      onOperationComplete()
      onClearSelection()
    } else {
      const message = `${action} completed: ${result.success_count} succeeded, ${result.failures.length} failed`
      showToast('error', message)
      onOperationComplete()
    }
  }

  const handleDelete = async () => {
    setConfirmDialog({ open: false, action: null })
    try {
      const result = await bulkDelete(selectedWorkflowIds)
      handleOperationResult(result, 'deleted')
    } catch (error: any) {
      showToast('error', error.message || 'Delete operation failed')
    }
  }

  const handleEnable = async () => {
    setConfirmDialog({ open: false, action: null })
    try {
      const result = await bulkEnable(selectedWorkflowIds)
      handleOperationResult(result, 'enabled')
    } catch (error: any) {
      showToast('error', error.message || 'Enable operation failed')
    }
  }

  const handleDisable = async () => {
    setConfirmDialog({ open: false, action: null })
    try {
      const result = await bulkDisable(selectedWorkflowIds)
      handleOperationResult(result, 'disabled')
    } catch (error: any) {
      showToast('error', error.message || 'Disable operation failed')
    }
  }

  const handleExport = async () => {
    try {
      const result = await bulkExport(selectedWorkflowIds)
      handleOperationResult(result.result, 'exported')
    } catch (error: any) {
      showToast('error', error.message || 'Export operation failed')
    }
  }

  const handleClone = async () => {
    setConfirmDialog({ open: false, action: null })
    try {
      const result = await bulkClone(selectedWorkflowIds)
      handleOperationResult(result.result, 'cloned')
    } catch (error: any) {
      showToast('error', error.message || 'Clone operation failed')
    }
  }

  const handleConfirm = () => {
    switch (confirmDialog.action) {
      case 'delete':
        handleDelete()
        break
      case 'enable':
        handleEnable()
        break
      case 'disable':
        handleDisable()
        break
      case 'clone':
        handleClone()
        break
    }
  }

  if (selectedCount === 0) {
    return null
  }

  const isLoading = bulkDeleting || bulkEnabling || bulkDisabling || bulkExporting || bulkCloning

  return (
    <>
      <div className="sticky top-0 z-10 bg-primary-600 border-b border-primary-700 px-4 py-3 mb-4 rounded-lg shadow-lg">
        <div className="flex items-center justify-between">
          <div className="flex items-center space-x-4">
            <div className="flex items-center space-x-2">
              <span className="inline-flex items-center justify-center px-2 py-1 text-xs font-bold text-white bg-primary-700 rounded-full">
                {selectedCount}
              </span>
              <span className="text-white font-medium">
                {selectedCount === 1 ? 'workflow' : 'workflows'} selected
              </span>
            </div>
            <div className="flex items-center space-x-2">
              <button
                onClick={() => setConfirmDialog({ open: true, action: 'delete' })}
                disabled={isLoading}
                className="flex items-center space-x-1 px-3 py-1.5 bg-red-600 hover:bg-red-700 disabled:bg-gray-600 disabled:cursor-not-allowed text-white text-sm font-medium rounded-lg transition-colors"
                title="Delete selected workflows"
              >
                {bulkDeleting ? (
                  <Loader2 className="w-4 h-4 animate-spin" />
                ) : (
                  <Trash2 className="w-4 h-4" />
                )}
                <span>Delete</span>
              </button>

              <button
                onClick={() => setConfirmDialog({ open: true, action: 'enable' })}
                disabled={isLoading}
                className="flex items-center space-x-1 px-3 py-1.5 bg-green-600 hover:bg-green-700 disabled:bg-gray-600 disabled:cursor-not-allowed text-white text-sm font-medium rounded-lg transition-colors"
                title="Enable selected workflows"
              >
                {bulkEnabling ? (
                  <Loader2 className="w-4 h-4 animate-spin" />
                ) : (
                  <Play className="w-4 h-4" />
                )}
                <span>Enable</span>
              </button>

              <button
                onClick={() => setConfirmDialog({ open: true, action: 'disable' })}
                disabled={isLoading}
                className="flex items-center space-x-1 px-3 py-1.5 bg-yellow-600 hover:bg-yellow-700 disabled:bg-gray-600 disabled:cursor-not-allowed text-white text-sm font-medium rounded-lg transition-colors"
                title="Disable selected workflows"
              >
                {bulkDisabling ? (
                  <Loader2 className="w-4 h-4 animate-spin" />
                ) : (
                  <Pause className="w-4 h-4" />
                )}
                <span>Disable</span>
              </button>

              <button
                onClick={handleExport}
                disabled={isLoading}
                className="flex items-center space-x-1 px-3 py-1.5 bg-blue-600 hover:bg-blue-700 disabled:bg-gray-600 disabled:cursor-not-allowed text-white text-sm font-medium rounded-lg transition-colors"
                title="Export selected workflows"
              >
                {bulkExporting ? (
                  <Loader2 className="w-4 h-4 animate-spin" />
                ) : (
                  <Download className="w-4 h-4" />
                )}
                <span>Export</span>
              </button>

              <button
                onClick={() => setConfirmDialog({ open: true, action: 'clone' })}
                disabled={isLoading}
                className="flex items-center space-x-1 px-3 py-1.5 bg-purple-600 hover:bg-purple-700 disabled:bg-gray-600 disabled:cursor-not-allowed text-white text-sm font-medium rounded-lg transition-colors"
                title="Clone selected workflows"
              >
                {bulkCloning ? (
                  <Loader2 className="w-4 h-4 animate-spin" />
                ) : (
                  <Copy className="w-4 h-4" />
                )}
                <span>Clone</span>
              </button>
            </div>
          </div>
          <button
            onClick={onClearSelection}
            disabled={isLoading}
            className="text-white hover:text-gray-200 text-sm font-medium disabled:opacity-50 disabled:cursor-not-allowed"
          >
            Clear selection
          </button>
        </div>
      </div>

      {toast && (
        <div
          className={`fixed top-4 right-4 z-50 px-4 py-3 rounded-lg shadow-lg ${
            toast.type === 'success'
              ? 'bg-green-600 text-white'
              : 'bg-red-600 text-white'
          }`}
        >
          <div className="flex items-center space-x-2">
            <span>{toast.message}</span>
          </div>
        </div>
      )}

      <ConfirmBulkDialog
        open={confirmDialog.open}
        action={confirmDialog.action || ''}
        count={selectedCount}
        destructive={confirmDialog.action === 'delete'}
        message={
          confirmDialog.action === 'delete'
            ? `Are you sure you want to delete ${selectedCount} workflow(s)? This action cannot be undone.`
            : undefined
        }
        onConfirm={handleConfirm}
        onCancel={() => setConfirmDialog({ open: false, action: null })}
      />
    </>
  )
}
