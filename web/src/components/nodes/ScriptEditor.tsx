import { useState } from 'react';

export interface ScriptEditorProps {
  value: string;
  onChange: (value: string) => void;
  timeout?: number;
  onTimeoutChange?: (timeout: number) => void;
  context?: Record<string, unknown>;
  placeholder?: string;
  error?: string;
}

const EXAMPLE_SCRIPTS = [
  {
    label: 'Simple calculation',
    script: `// Access trigger data
var x = context.trigger.x;
var y = context.trigger.y;

// Perform calculation
return x + y;`,
  },
  {
    label: 'Data transformation',
    script: `// Transform array of objects
var users = context.trigger.users;
var result = [];

for (var i = 0; i < users.length; i++) {
  result.push({
    name: users[i].name.toUpperCase(),
    email: users[i].email.toLowerCase()
  });
}

return result;`,
  },
  {
    label: 'Conditional logic',
    script: `// Check condition and return different values
var age = context.trigger.age;

if (age >= 18) {
  return {
    status: "adult",
    canVote: true
  };
} else {
  return {
    status: "minor",
    canVote: false
  };
}`,
  },
  {
    label: 'String manipulation',
    script: `// Access previous step output
var text = context.steps.step1.result;

// Clean and format text
var cleaned = text.trim().toLowerCase();
var words = cleaned.split(' ');

return {
  original: text,
  wordCount: words.length,
  firstWord: words[0]
};`,
  },
  {
    label: 'Array operations',
    script: `// Filter and map array
var numbers = context.trigger.numbers;

// Filter even numbers
var evens = [];
for (var i = 0; i < numbers.length; i++) {
  if (numbers[i] % 2 === 0) {
    evens.push(numbers[i]);
  }
}

// Calculate sum
var sum = 0;
for (var i = 0; i < evens.length; i++) {
  sum += evens[i];
}

return {
  evens: evens,
  sum: sum,
  average: sum / evens.length
};`,
  },
];

const API_DOCUMENTATION = [
  {
    name: 'context.trigger',
    description: 'Access data from the workflow trigger',
    example: 'var name = context.trigger.name;',
  },
  {
    name: 'context.steps',
    description: 'Access outputs from previous workflow steps',
    example: 'var result = context.steps.step1.result;',
  },
  {
    name: 'context.env',
    description: 'Access environment variables (tenant_id, execution_id, workflow_id)',
    example: 'var tenant = context.env.tenant_id;',
  },
  {
    name: 'JSON.parse()',
    description: 'Parse JSON string to object',
    example: 'var obj = JSON.parse(\'{"key": "value"}\');',
  },
  {
    name: 'JSON.stringify()',
    description: 'Convert object to JSON string',
    example: 'var json = JSON.stringify({key: "value"});',
  },
  {
    name: 'Math.*',
    description: 'Mathematical functions (sqrt, pow, round, floor, ceil, abs, min, max, random)',
    example: 'var rounded = Math.round(3.7);',
  },
];

export const ScriptEditor: React.FC<ScriptEditorProps> = ({
  value,
  onChange,
  timeout = 30,
  onTimeoutChange,
  context = {},
  placeholder = 'Enter JavaScript code...',
  error,
}) => {
  const [showExamples, setShowExamples] = useState(false);
  const [showAPI, setShowAPI] = useState(false);

  const handleTimeoutChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const newTimeout = parseInt(e.target.value, 10) || 0;
    onTimeoutChange?.(newTimeout);
  };

  return (
    <div className="script-editor">
      {/* Header with controls */}
      <div className="mb-2 flex justify-between items-center">
        <label className="block text-sm font-medium text-gray-700">
          JavaScript Code
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
            onClick={() => setShowAPI(!showAPI)}
            className="text-xs text-blue-600 hover:text-blue-800"
          >
            {showAPI ? 'Hide' : 'Show'} API
          </button>
        </div>
      </div>

      {/* Script textarea */}
      <textarea
        value={value}
        onChange={(e) => onChange(e.target.value)}
        placeholder={placeholder}
        className={`w-full h-64 px-3 py-2 text-sm font-mono border rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 ${
          error ? 'border-red-500' : 'border-gray-300'
        }`}
        data-testid="script-input"
        spellCheck={false}
      />

      {/* Error message */}
      {error && (
        <p className="mt-1 text-sm text-red-600" data-testid="script-error">
          {error}
        </p>
      )}

      {/* Timeout configuration */}
      {onTimeoutChange && (
        <div className="mt-3">
          <label className="block text-sm font-medium text-gray-700 mb-1">
            Timeout (seconds)
          </label>
          <input
            type="number"
            value={timeout}
            onChange={handleTimeoutChange}
            min={1}
            max={300}
            className="w-32 px-3 py-2 text-sm border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
            data-testid="timeout-input"
          />
          <p className="mt-1 text-xs text-gray-500">
            Maximum execution time (1-300 seconds)
          </p>
        </div>
      )}

      {/* Security notice */}
      <div className="mt-3 p-2 bg-yellow-50 border border-yellow-200 rounded-md">
        <p className="text-xs text-yellow-800">
          <span className="font-semibold">Sandboxed Environment:</span> No file system or network access.
          Scripts run in an isolated JavaScript VM with resource limits.
        </p>
      </div>

      {/* Examples section */}
      {showExamples && (
        <div className="mt-3 p-3 bg-gray-50 border border-gray-200 rounded-md">
          <h4 className="text-sm font-semibold mb-2">Example Scripts</h4>
          <div className="space-y-3">
            {EXAMPLE_SCRIPTS.map((example, idx) => (
              <div key={idx} className="border-b border-gray-200 pb-2 last:border-0">
                <div className="flex items-center justify-between mb-1">
                  <span className="text-sm font-medium text-gray-700">{example.label}</span>
                  <button
                    type="button"
                    onClick={() => onChange(example.script)}
                    className="text-xs text-blue-600 hover:text-blue-800 font-medium"
                  >
                    Use
                  </button>
                </div>
                <pre className="text-xs text-gray-600 bg-white p-2 rounded border border-gray-200 overflow-x-auto">
                  <code>{example.script}</code>
                </pre>
              </div>
            ))}
          </div>
        </div>
      )}

      {/* API documentation */}
      {showAPI && (
        <div className="mt-3 p-3 bg-gray-50 border border-gray-200 rounded-md max-h-96 overflow-y-auto">
          <h4 className="text-sm font-semibold mb-2">Available JavaScript API</h4>
          <div className="space-y-2">
            {API_DOCUMENTATION.map((api, idx) => (
              <div key={idx} className="text-xs border-b border-gray-200 pb-2 last:border-0">
                <div className="font-medium text-gray-900 font-mono">{api.name}</div>
                <div className="text-gray-600 mt-1">{api.description}</div>
                <code className="block mt-1 text-blue-600 bg-white p-1 rounded border border-gray-200">
                  {api.example}
                </code>
              </div>
            ))}
          </div>
        </div>
      )}

      {/* Context viewer */}
      {Object.keys(context).length > 0 && (
        <div className="mt-3">
          <details className="text-xs">
            <summary className="cursor-pointer text-gray-600 hover:text-gray-800 font-medium">
              Available Context
            </summary>
            <div className="mt-2 p-2 bg-gray-50 border border-gray-200 rounded-md max-h-48 overflow-auto">
              <pre className="text-xs">
                {JSON.stringify(context, null, 2)}
              </pre>
            </div>
          </details>
        </div>
      )}
    </div>
  );
};
