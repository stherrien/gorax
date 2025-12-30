import { createContext, useContext, useState, ReactNode } from 'react'

interface WorkflowSelectionContextValue {
  selectedWorkflowIds: string[]
  toggleSelection: (id: string) => void
  selectAll: (ids: string[]) => void
  clearSelection: () => void
  isSelected: (id: string) => boolean
}

const WorkflowSelectionContext = createContext<WorkflowSelectionContextValue | null>(null)

export function useWorkflowSelection() {
  const context = useContext(WorkflowSelectionContext)
  if (!context) {
    throw new Error('useWorkflowSelection must be used within WorkflowSelectionProvider')
  }
  return context
}

interface WorkflowSelectionProviderProps {
  children: ReactNode
}

export function WorkflowSelectionProvider({ children }: WorkflowSelectionProviderProps) {
  const [selectedWorkflowIds, setSelectedWorkflowIds] = useState<string[]>([])

  const toggleSelection = (id: string) => {
    setSelectedWorkflowIds((prev) =>
      prev.includes(id) ? prev.filter((wid) => wid !== id) : [...prev, id]
    )
  }

  const selectAll = (ids: string[]) => {
    setSelectedWorkflowIds(ids)
  }

  const clearSelection = () => {
    setSelectedWorkflowIds([])
  }

  const isSelected = (id: string) => {
    return selectedWorkflowIds.includes(id)
  }

  return (
    <WorkflowSelectionContext.Provider
      value={{
        selectedWorkflowIds,
        toggleSelection,
        selectAll,
        clearSelection,
        isSelected,
      }}
    >
      {children}
    </WorkflowSelectionContext.Provider>
  )
}
