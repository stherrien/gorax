import React, { useState, useRef, useEffect } from 'react'

export interface PromptEditorProps {
  value: string
  onChange: (value: string) => void
  label?: string
  placeholder?: string
  disabled?: boolean
  readOnly?: boolean
  rows?: number
  error?: string
  helperText?: string
  showVariables?: boolean
  variables?: string[]
  maxLength?: number
  showCharCount?: boolean
  className?: string
}

export const PromptEditor: React.FC<PromptEditorProps> = ({
  value,
  onChange,
  label,
  placeholder,
  disabled = false,
  readOnly = false,
  rows = 4,
  error,
  helperText,
  showVariables = false,
  variables = [],
  maxLength,
  showCharCount = false,
  className = '',
}) => {
  const [isVariablePickerOpen, setIsVariablePickerOpen] = useState(false)
  const textareaRef = useRef<HTMLTextAreaElement>(null)
  const cursorPositionRef = useRef<number>(0)
  const pickerRef = useRef<HTMLDivElement>(null)

  // Close variable picker when clicking outside
  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (pickerRef.current && !pickerRef.current.contains(event.target as Node)) {
        setIsVariablePickerOpen(false)
      }
    }

    if (isVariablePickerOpen) {
      document.addEventListener('mousedown', handleClickOutside)
    }

    return () => {
      document.removeEventListener('mousedown', handleClickOutside)
    }
  }, [isVariablePickerOpen])

  const handleChange = (e: React.ChangeEvent<HTMLTextAreaElement>) => {
    onChange(e.target.value)
    cursorPositionRef.current = e.target.selectionStart || 0
  }

  const handleInsertVariable = (variable: string) => {
    const textarea = textareaRef.current
    if (!textarea) return

    const start = cursorPositionRef.current
    const before = value.slice(0, start)
    const after = value.slice(start)
    const variableTemplate = `{{${variable}}}`

    const newValue = before + variableTemplate + after
    onChange(newValue)

    // Update cursor position for next insertion
    cursorPositionRef.current = start + variableTemplate.length

    setIsVariablePickerOpen(false)
  }

  const handleTextareaClick = () => {
    cursorPositionRef.current = textareaRef.current?.selectionStart || 0
  }

  const handleTextareaKeyUp = () => {
    cursorPositionRef.current = textareaRef.current?.selectionStart || 0
  }

  const getCharCountClass = () => {
    if (!maxLength) return 'text-gray-500'
    const percentage = value.length / maxLength
    if (percentage > 1) return 'text-red-600'
    if (percentage > 0.9) return 'text-yellow-600'
    return 'text-gray-500'
  }

  return (
    <div className={className}>
      {label && (
        <label className="block text-sm font-medium text-gray-700 mb-1">{label}</label>
      )}

      <div className="relative">
        <textarea
          ref={textareaRef}
          value={value}
          onChange={handleChange}
          onClick={handleTextareaClick}
          onKeyUp={handleTextareaKeyUp}
          placeholder={placeholder}
          disabled={disabled}
          readOnly={readOnly}
          rows={rows}
          className={`w-full px-3 py-2 border rounded-md focus:ring-2 focus:ring-blue-500 focus:outline-none resize-y font-mono text-sm ${
            error ? 'border-red-500' : 'border-gray-300'
          } ${disabled ? 'bg-gray-100 cursor-not-allowed' : 'bg-white'}`}
        />

        {showVariables && variables.length > 0 && (
          <div className="absolute right-2 top-2" ref={pickerRef}>
            <button
              type="button"
              onClick={() => setIsVariablePickerOpen(!isVariablePickerOpen)}
              disabled={disabled || readOnly}
              className="px-2 py-1 text-xs font-medium text-blue-600 bg-blue-50 rounded hover:bg-blue-100 disabled:opacity-50 disabled:cursor-not-allowed"
            >
              Insert Variable
            </button>

            {isVariablePickerOpen && (
              <div className="absolute right-0 mt-1 w-56 bg-white border border-gray-300 rounded-md shadow-lg z-10 max-h-48 overflow-y-auto">
                <div className="py-1">
                  {variables.map((variable) => (
                    <button
                      key={variable}
                      type="button"
                      onClick={() => handleInsertVariable(variable)}
                      className="w-full px-3 py-2 text-left text-sm hover:bg-gray-100 focus:bg-gray-100 focus:outline-none font-mono"
                    >
                      {variable}
                    </button>
                  ))}
                </div>
              </div>
            )}
          </div>
        )}
      </div>

      <div className="flex items-center justify-between mt-1">
        <div className="flex-1">
          {error && <p className="text-red-500 text-xs">{error}</p>}
          {!error && helperText && <p className="text-gray-500 text-xs">{helperText}</p>}
        </div>
        {showCharCount && maxLength && (
          <span className={`text-xs ${getCharCountClass()}`}>
            {value.length} / {maxLength}
          </span>
        )}
      </div>
    </div>
  )
}
