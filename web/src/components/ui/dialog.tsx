import { ReactNode, HTMLAttributes, createContext, useContext, useState } from 'react'

interface DialogContextValue {
  open: boolean
  onOpenChange: (open: boolean) => void
}

const DialogContext = createContext<DialogContextValue | null>(null)

interface DialogProps {
  children: ReactNode
  open?: boolean
  onOpenChange?: (open: boolean) => void
}

export function Dialog({ children, open: controlledOpen, onOpenChange }: DialogProps) {
  const [internalOpen, setInternalOpen] = useState(false)
  const open = controlledOpen ?? internalOpen
  const handleOpenChange = onOpenChange ?? setInternalOpen

  return (
    <DialogContext.Provider value={{ open, onOpenChange: handleOpenChange }}>
      {children}
    </DialogContext.Provider>
  )
}

export function DialogTrigger({ children, asChild }: { children: ReactNode; asChild?: boolean }) {
  const context = useContext(DialogContext)
  if (!context) throw new Error('DialogTrigger must be used within Dialog')

  const handleClick = () => context.onOpenChange(true)

  if (asChild) {
    return <span onClick={handleClick}>{children}</span>
  }

  return <button onClick={handleClick}>{children}</button>
}

export function DialogContent({ children, className = '' }: { children: ReactNode; className?: string }) {
  const context = useContext(DialogContext)
  if (!context) throw new Error('DialogContent must be used within Dialog')

  if (!context.open) return null

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center">
      <div className="fixed inset-0 bg-black/50" onClick={() => context.onOpenChange(false)} />
      <div className={`relative z-50 bg-white rounded-lg shadow-lg p-6 max-w-lg w-full mx-4 ${className}`}>
        {children}
      </div>
    </div>
  )
}

interface DialogHeaderProps extends HTMLAttributes<HTMLDivElement> {
  children: ReactNode
}

export function DialogHeader({ children, className = '', ...props }: DialogHeaderProps) {
  return (
    <div className={`flex flex-col space-y-1.5 text-center sm:text-left ${className}`} {...props}>
      {children}
    </div>
  )
}

export function DialogTitle({ children, className = '' }: { children: ReactNode; className?: string }) {
  return <h2 className={`text-lg font-semibold leading-none tracking-tight ${className}`}>{children}</h2>
}

export function DialogDescription({ children, className = '' }: { children: ReactNode; className?: string }) {
  return <p className={`text-sm text-gray-500 ${className}`}>{children}</p>
}
