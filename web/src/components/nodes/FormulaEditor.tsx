import { useState, useEffect } from 'react';

export interface FormulaEditorProps {
  value: string;
  onChange: (value: string) => void;
  context?: Record<string, unknown>;
  placeholder?: string;
  error?: string;
}

const BUILT_IN_FUNCTIONS = [
  { name: 'upper', description: 'Converts string to uppercase', example: 'upper("hello")' },
  { name: 'lower', description: 'Converts string to lowercase', example: 'lower("HELLO")' },
  { name: 'trim', description: 'Removes whitespace', example: 'trim("  hello  ")' },
  { name: 'concat', description: 'Concatenates strings', example: 'concat("hello", " ", "world")' },
  { name: 'substr', description: 'Extracts substring', example: 'substr("hello", 0, 5)' },
  { name: 'now', description: 'Returns current time', example: 'now()' },
  { name: 'dateFormat', description: 'Formats a date', example: 'dateFormat(now(), "2006-01-02")' },
  { name: 'dateParse', description: 'Parses a date string', example: 'dateParse("2025-12-17", "2006-01-02")' },
  { name: 'addDays', description: 'Adds days to date', example: 'addDays(now(), 5)' },
  { name: 'round', description: 'Rounds to nearest integer', example: 'round(4.6)' },
  { name: 'ceil', description: 'Rounds up', example: 'ceil(4.1)' },
  { name: 'floor', description: 'Rounds down', example: 'floor(4.9)' },
  { name: 'abs', description: 'Absolute value', example: 'abs(-5)' },
  { name: 'min', description: 'Returns minimum value', example: 'min(5, 3, 7)' },
  { name: 'max', description: 'Returns maximum value', example: 'max(5, 3, 7)' },
  { name: 'len', description: 'Returns length of array or string', example: 'len([1, 2, 3])' },
];

const EXAMPLE_FORMULAS = [
  { label: 'String manipulation', formula: 'upper(trim(name))' },
  { label: 'Math calculation', formula: 'round((price * quantity) * (1 + taxRate))' },
  { label: 'Conditional', formula: 'age > 18 ? "adult" : "minor"' },
  { label: 'Date formatting', formula: 'dateFormat(now(), "2006-01-02")' },
  { label: 'Array length', formula: 'len(items)' },
];

export const FormulaEditor: React.FC<FormulaEditorProps> = ({
  value,
  onChange,
  context = {},
  placeholder = 'Enter formula expression...',
  error,
}) => {
  const [showHelp, setShowHelp] = useState(false);
  const [showExamples, setShowExamples] = useState(false);

  return (
    <div className="formula-editor">
      <div className="mb-2 flex justify-between items-center">
        <label className="block text-sm font-medium text-gray-700">
          Formula Expression
        </label>
        <div className="flex gap-2">
          <button
            type="button"
            onClick={() => setShowExamples(!showExamples)}
            className="text-xs text-blue-600 hover:text-blue-800"
          >
            {showExamples ? 'Hide' : 'Show'} Examples
          </button>
          <button
            type="button"
            onClick={() => setShowHelp(!showHelp)}
            className="text-xs text-blue-600 hover:text-blue-800"
          >
            {showHelp ? 'Hide' : 'Show'} Functions
          </button>
        </div>
      </div>

      <textarea
        value={value}
        onChange={(e) => onChange(e.target.value)}
        placeholder={placeholder}
        className={`w-full h-24 px-3 py-2 text-sm font-mono border rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 ${
          error ? 'border-red-500' : 'border-gray-300'
        }`}
        data-testid="formula-input"
      />

      {error && (
        <p className="mt-1 text-sm text-red-600" data-testid="formula-error">
          {error}
        </p>
      )}

      {showExamples && (
        <div className="mt-2 p-3 bg-gray-50 border border-gray-200 rounded-md">
          <h4 className="text-sm font-semibold mb-2">Example Formulas</h4>
          <div className="space-y-1">
            {EXAMPLE_FORMULAS.map((example, idx) => (
              <div key={idx} className="flex items-center justify-between text-xs">
                <div>
                  <span className="text-gray-600">{example.label}:</span>{' '}
                  <code className="text-blue-600">{example.formula}</code>
                </div>
                <button
                  type="button"
                  onClick={() => onChange(example.formula)}
                  className="text-blue-600 hover:text-blue-800"
                >
                  Use
                </button>
              </div>
            ))}
          </div>
        </div>
      )}

      {showHelp && (
        <div className="mt-2 p-3 bg-gray-50 border border-gray-200 rounded-md max-h-64 overflow-y-auto">
          <h4 className="text-sm font-semibold mb-2">Available Functions</h4>
          <div className="space-y-2">
            {BUILT_IN_FUNCTIONS.map((func, idx) => (
              <div key={idx} className="text-xs">
                <div className="font-medium text-gray-900">{func.name}()</div>
                <div className="text-gray-600">{func.description}</div>
                <code className="text-blue-600 text-xs">{func.example}</code>
              </div>
            ))}
          </div>
        </div>
      )}

      {Object.keys(context).length > 0 && (
        <div className="mt-2">
          <details className="text-xs">
            <summary className="cursor-pointer text-gray-600 hover:text-gray-800">
              Available Variables
            </summary>
            <div className="mt-1 p-2 bg-gray-50 border border-gray-200 rounded-md">
              <pre className="text-xs overflow-auto">
                {JSON.stringify(context, null, 2)}
              </pre>
            </div>
          </details>
        </div>
      )}
    </div>
  );
};
